# ZIP Agent

ZIP upload bridge service that converts user-provided ZIP archives into Git repositories for automated deployment workflows.
It is designed to sit between NOMO and Gitea, so Coolify can deploy from a stable Git URL.
This project currently focuses on fast upload-to-repo conversion with forced sync to `main`.

ZIP 上传中转服务，将用户上传的 ZIP 文件转换为 Git 仓库，供 Coolify 部署使用。

## 架构

```text
用户上传 ZIP -> NOMO -> ZIP Agent -> Gitea -> Coolify 部署
```

## 线上地址

`https://zip.nomoo.top`

## 功能

- 接收 ZIP 文件上传
- 自动解压并处理目录结构（去除单个根目录前缀）
- 自动创建 Gitea 仓库（不存在时）
- 初始化 Git 并强制推送到 `main`
- 支持按 `project_id` 删除仓库
- 过滤 macOS/Windows 常见垃圾文件（如 `__MACOSX`、`.DS_Store`、`Thumbs.db`）

## API

### 健康检查

```bash
GET /health
```

响应示例：

```json
{"status":"ok"}
```

### 上传 ZIP

```bash
POST /upload
Content-Type: multipart/form-data
Authorization: Bearer YOUR_API_KEY

file: <zip-file>
project_id: <project-uuid>
```

响应示例：

```json
{
  "success": true,
  "git_url": "https://git.nomoo.top/nomo-admin/project-xxx.git"
}
```

### 删除仓库

```bash
DELETE /delete?project_id=xxx
Authorization: Bearer YOUR_API_KEY
```

说明：

- 当前实现同时兼容 `POST /delete?project_id=xxx` 与 `DELETE`。

## 部署

### 环境变量

| 变量 | 说明 | 必填 |
|------|------|------|
| `GITEA_URL` | Gitea 内部地址（服务端 API 与 Git 推送使用） | 是 |
| `GITEA_TOKEN` | Gitea API Token | 是 |
| `GITEA_OWNER` | 仓库所有者用户名 | 是 |
| `GITEA_PUBLIC_URL` | 对外可访问的 Gitea 地址（返回给 Coolify） | 否，未设置时回退为 `GITEA_URL` |
| `ZIP_AGENT_API_KEY` | API 认证密钥，留空则不校验 | 否（推荐设置） |
| `PORT` | 服务端口 | 否（默认 `8080`） |

### Docker Compose

```bash
docker-compose up -d
```

### 在 Coolify 中部署

1. 创建新的 Docker Compose 服务。
2. 粘贴 `docker-compose.yml` 内容。
3. 设置环境变量。
4. 配置域名 `zip.nomoo.top`。
5. 部署并验证 `/health`。

## 本地开发

```bash
# 安装依赖
go mod download

# 运行
GITEA_URL=http://localhost:3001 \
GITEA_TOKEN=your-token \
GITEA_OWNER=nomo-admin \
go run main.go

# 测试健康检查
curl http://localhost:8080/health
```

## 注意事项

- Gitea 仓库需要可被 Coolify 访问（通常为公开仓库）。
- 单次上传大小限制为 `100MB`。
- 每次上传会覆盖远端 `main`（`git push -f`）。
- 当 ZIP 内只有一层根目录时，会自动去掉该目录前缀。

## 开源协作

- 贡献指南：[`CONTRIBUTING.md`](CONTRIBUTING.md)
- 行为准则：[`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md)
- 安全策略：[`SECURITY.md`](SECURITY.md)
- 版本变更：[`CHANGELOG.md`](CHANGELOG.md)

## 发布流程

1. 在 `CHANGELOG.md` 维护即将发布版本条目。
2. 合并到 `main` 后创建语义化版本 tag（例如 `v0.1.0`）。
3. 推送 `main` 与 tag 到远程仓库。
4. 在 GitHub Release 页面选择对应 tag，粘贴 release notes 发布。

## License

MIT License，见 [`LICENSE`](LICENSE)。
