package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	giteaURL       = os.Getenv("GITEA_URL")
	giteaPublicURL = os.Getenv("GITEA_PUBLIC_URL")
	giteaToken     = os.Getenv("GITEA_TOKEN")
	giteaOwner     = os.Getenv("GITEA_OWNER")
	apiKey         = os.Getenv("ZIP_AGENT_API_KEY")
	workDir        = "/tmp/zip-agent"
)

func main() {
	if giteaURL == "" || giteaToken == "" || giteaOwner == "" {
		log.Fatal("Missing required env: GITEA_URL, GITEA_TOKEN, GITEA_OWNER")
	}
	
	// 如果没有设置公开 URL，使用内部 URL
	if giteaPublicURL == "" {
		giteaPublicURL = giteaURL
	}

	os.MkdirAll(workDir, 0755)

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/upload", authMiddleware(uploadHandler))
	http.HandleFunc("/delete", authMiddleware(deleteHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("ZIP Agent starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if apiKey != "" {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer "+apiKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type UploadResponse struct {
	Success bool   `json:"success"`
	GitURL  string `json:"git_url,omitempty"`
	Error   string `json:"error,omitempty"`
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 限制上传大小 100MB
	r.ParseMultipartForm(100 << 20)

	projectID := r.FormValue("project_id")
	if projectID == "" {
		respondJSON(w, UploadResponse{Error: "project_id required"}, 400)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respondJSON(w, UploadResponse{Error: "file required"}, 400)
		return
	}
	defer file.Close()

	// 读取 ZIP 内容
	zipData, err := io.ReadAll(file)
	if err != nil {
		respondJSON(w, UploadResponse{Error: "failed to read file"}, 500)
		return
	}

	// 处理上传
	gitURL, err := processUpload(projectID, zipData)
	if err != nil {
		log.Printf("Upload failed for %s: %v", projectID, err)
		respondJSON(w, UploadResponse{Error: err.Error()}, 500)
		return
	}

	respondJSON(w, UploadResponse{Success: true, GitURL: gitURL}, 200)
}

func processUpload(projectID string, zipData []byte) (string, error) {
	repoName := fmt.Sprintf("project-%s", projectID)
	extractDir := filepath.Join(workDir, repoName)

	// 清理旧目录
	os.RemoveAll(extractDir)
	os.MkdirAll(extractDir, 0755)
	defer os.RemoveAll(extractDir)

	// 解压 ZIP
	if err := unzip(zipData, extractDir); err != nil {
		return "", fmt.Errorf("unzip failed: %w", err)
	}

	// 检查是否需要创建仓库
	repoExists, err := checkRepoExists(repoName)
	if err != nil {
		return "", fmt.Errorf("check repo failed: %w", err)
	}

	if !repoExists {
		if err := createRepo(repoName); err != nil {
			return "", fmt.Errorf("create repo failed: %w", err)
		}
	}

	// Git 操作
	if err := gitPush(extractDir, repoName); err != nil {
		return "", fmt.Errorf("git push failed: %w", err)
	}

	// 返回公开 URL 供 Coolify 使用
	gitURL := fmt.Sprintf("%s/%s/%s.git", giteaPublicURL, giteaOwner, repoName)
	return gitURL, nil
}

func unzip(data []byte, dest string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	// 找到公共前缀（处理 ZIP 内有单个根目录的情况）
	var prefix string
	if len(reader.File) > 0 {
		first := reader.File[0].Name
		if strings.Contains(first, "/") {
			prefix = strings.Split(first, "/")[0] + "/"
		}
	}

	for _, f := range reader.File {
		name := f.Name
		// 去掉公共前缀
		if prefix != "" && strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
		}
		if name == "" {
			continue
		}

		// 过滤 macOS 垃圾文件
		if shouldSkipFile(name) {
			continue
		}

		path := filepath.Join(dest, name)

		// 安全检查：防止路径遍历
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
			continue
		}

		os.MkdirAll(filepath.Dir(path), 0755)

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// shouldSkipFile 检查是否应该跳过该文件（macOS/Windows 垃圾文件）
func shouldSkipFile(name string) bool {
	// 获取文件名（不含路径）
	baseName := filepath.Base(name)
	
	// macOS 资源分支文件（以 ._ 开头）
	if strings.HasPrefix(baseName, "._") {
		return true
	}
	
	// macOS __MACOSX 目录
	if strings.HasPrefix(name, "__MACOSX/") || name == "__MACOSX" {
		return true
	}
	
	// macOS .DS_Store
	if baseName == ".DS_Store" {
		return true
	}
	
	// Windows Thumbs.db
	if baseName == "Thumbs.db" || baseName == "desktop.ini" {
		return true
	}
	
	return false
}

func checkRepoExists(name string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s", giteaURL, giteaOwner, name)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "token "+giteaToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}

func createRepo(name string) error {
	url := fmt.Sprintf("%s/api/v1/user/repos", giteaURL)
	body := map[string]interface{}{
		"name":     name,
		"private":  false,
		"auto_init": false,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	req.Header.Set("Authorization", "token "+giteaToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create repo failed: %s", string(respBody))
	}

	return nil
}

func gitPush(dir, repoName string) error {
	remoteURL := fmt.Sprintf("%s/%s/%s.git", giteaURL, giteaOwner, repoName)
	// 使用 token 认证
	remoteURL = strings.Replace(remoteURL, "://", fmt.Sprintf("://oauth2:%s@", giteaToken), 1)

	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "nomo@nomoo.top"},
		{"git", "config", "user.name", "Nomo Bot"},
		{"git", "add", "."},
		{"git", "commit", "-m", fmt.Sprintf("Upload at %s", time.Now().Format(time.RFC3339))},
		{"git", "branch", "-M", "main"},
		{"git", "remote", "add", "origin", remoteURL},
		{"git", "push", "-f", "origin", "main"},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		output, err := cmd.CombinedOutput()
		if err != nil {
			// 忽略 remote already exists 错误
			if strings.Contains(string(output), "already exists") {
				continue
			}
			return fmt.Errorf("%s failed: %s", args[0], string(output))
		}
	}

	return nil
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		respondJSON(w, map[string]string{"error": "project_id required"}, 400)
		return
	}

	repoName := fmt.Sprintf("project-%s", projectID)
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s", giteaURL, giteaOwner, repoName)

	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "token "+giteaToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		respondJSON(w, map[string]string{"error": err.Error()}, 500)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 404 {
		respondJSON(w, map[string]string{"error": "delete failed"}, 500)
		return
	}

	respondJSON(w, map[string]string{"success": "true"}, 200)
}

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
