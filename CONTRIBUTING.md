# Contributing

Thanks for considering contributing to ZIP Agent.
This document is primarily written in Chinese, with concise English sections where useful.

感谢你为 ZIP Agent 做贡献。本文档以中文为主，包含必要英文说明。

## 贡献范围

欢迎以下类型贡献：

- Bug 修复
- 功能增强
- 文档改进
- 测试与稳定性提升
- 部署与运维实践优化

## 开发前准备

1. Fork 并克隆仓库。
2. 安装 Go（建议 `1.21+`）。
3. 按 `.env.example` 准备环境变量。
4. 本地运行服务并验证健康检查：

```bash
go run main.go
curl http://localhost:8080/health
```

## 分支命名

建议使用以下前缀：

- `feat/<short-description>`
- `fix/<short-description>`
- `docs/<short-description>`
- `chore/<short-description>`

示例：`fix/public-url-fallback`

## 提交规范

建议使用 Conventional Commits 风格：

- `feat: add xxx`
- `fix: resolve xxx`
- `docs: update xxx`
- `chore: adjust xxx`

## Pull Request 要求

提交 PR 前，请确保：

- 变更目标清晰，描述完整。
- 如涉及行为变化，更新 `README.md` 和 `CHANGELOG.md`。
- 本地至少通过：

```bash
go test ./...
go build ./...
```

## PR 检查清单

- [ ] 代码仅包含与目标相关的修改
- [ ] 无敏感信息（token、密钥、内部地址等）
- [ ] 文档已同步更新
- [ ] 已补充或说明测试结果
- [ ] 向后兼容性已评估（如适用）

## 反馈与沟通

- 普通问题：GitHub Issues
- 安全问题：请遵循 `SECURITY.md` 私下报告

## 行为规范

参与协作即视为同意遵守 `CODE_OF_CONDUCT.md`。
