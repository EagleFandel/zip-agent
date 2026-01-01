# ZIP Agent

ZIP 上传中转服务，将用户上传的 ZIP 文件转换为 Git 仓库。

## 架构

```
用户 → Vercel (nomo) → ZIP Agent → Gitea → Coolify
```

## 部署

### 1. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 填入配置
```

### 2. 启动服务

```bash
docker-compose up -d
```

### 3. 初始化 Gitea

1. 访问 `http://your-server:3001`
2. 完成初始安装（SQLite3）
3. 创建管理员账号
4. 生成 API Token（设置 → 应用）
5. 更新 .env 中的 GITEA_TOKEN

### 4. 重启服务

```bash
docker-compose restart zip-agent
```

## API

### 上传 ZIP

```bash
POST /upload
Content-Type: multipart/form-data
Authorization: Bearer YOUR_API_KEY

file: <zip-file>
project_id: <project-uuid>
```

响应：
```json
{
  "success": true,
  "git_url": "http://git.example.com/nomo-admin/project-xxx.git"
}
```

### 删除仓库

```bash
DELETE /delete?project_id=xxx
Authorization: Bearer YOUR_API_KEY
```

### 健康检查

```bash
GET /health
```

## 环境变量

| 变量 | 说明 | 必填 |
|------|------|------|
| GITEA_URL | Gitea 内部地址 | 是 |
| GITEA_TOKEN | Gitea API Token | 是 |
| GITEA_OWNER | 仓库所有者用户名 | 是 |
| GITEA_PUBLIC_URL | Gitea 公开地址（给 Coolify 用） | 是 |
| ZIP_AGENT_API_KEY | API 认证密钥 | 推荐 |
| PORT | 服务端口 | 否（默认 8080） |

## 注意事项

- Gitea 仓库必须是公开的，Coolify 才能 clone
- ZIP 文件大小限制 100MB
- 每次上传会强制覆盖（force push）
