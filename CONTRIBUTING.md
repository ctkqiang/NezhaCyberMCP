# Contributing to 哪吒网络安全 MCP / 贡献指南

Thank you for your interest in contributing to 哪吒网络安全 MCP (NezhaCyberMCP).
感谢你对哪吒网络安全 MCP (NezhaCyberMCP) 的贡献兴趣。

---

## Table of Contents / 目录

- [Code of Conduct / 行为准则](#code-of-conduct--行为准则)
- [Getting Started / 快速开始](#getting-started--快速开始)
- [Development Workflow / 开发流程](#development-workflow--开发流程)
- [Code Standards / 代码规范](#code-standards--代码规范)
- [Commit Messages / 提交信息](#commit-messages--提交信息)
- [Pull Request Process / PR 流程](#pull-request-process--pr-流程)
- [Reporting Issues / 报告问题](#reporting-issues--报告问题)

---

## Code of Conduct / 行为准则

Be respectful, constructive, and professional in all interactions.
在所有交流中保持尊重、建设性和专业态度。

---

## Getting Started / 快速开始

### Prerequisites / 前置条件

- Go 1.26.1+
- PostgreSQL 14+ (or SQLite for local testing / 本地测试可使用 SQLite)
- `make`
- `golangci-lint` (for linting / 用于代码规范检查)

### Setup / 环境配置

```bash
git clone https://github.com/ctkqiang/NezhaCyberMCP.git
cd NezhaCyberMCP
cp .env.example .env
# Edit .env with your local database credentials
# 编辑 .env，填入本地数据库凭证
make build
make run
```

---

## Development Workflow / 开发流程

1. **Fork** the repository and create a feature branch from `main`.
   **Fork** 仓库并从 `main` 创建功能分支。

   ```bash
   git checkout -b feat/your-feature-name
   ```

2. Make your changes following the [Code Standards](#code-standards--代码规范).
   按照[代码规范](#code-standards--代码规范)进行修改。

3. Run all checks locally before pushing:
   推送前在本地运行所有检查：

   ```bash
   go fmt ./...
   go vet ./...
   make lint
   make test
   ```

4. Open a Pull Request against `main`.
   向 `main` 分支提交 Pull Request。

---

## Code Standards / 代码规范

All code must follow the conventions defined in [`.trae/rules/code-format.md`](.trae/rules/code-format.md).
所有代码必须遵循 [`.trae/rules/code-format.md`](.trae/rules/code-format.md) 中定义的规范。

Key rules / 核心规则：

| Rule / 规则 | Requirement / 要求 |
|-------------|-------------------|
| Comments / 注释 | All exported functions must have Simplified Chinese doc comments / 所有导出函数必须有简体中文文档注释 |
| Error handling / 错误处理 | Use `fmt.Errorf("operation: %w", err)` — never ignore errors / 使用 `fmt.Errorf` 包装，从不忽略错误 |
| Logging / 日志 | Use `utilities.LogStart/LogProgress/LogSuccess/LogError/LogWarn` — no `fmt.Println` / 使用规范日志函数 |
| Database / 数据库 | Always use `WithContext(ctx)` and `CreateInBatches(100)` in transactions / 始终使用 `WithContext(ctx)` 和事务内批量写入 |
| Credentials / 凭证 | Never hardcode — always use environment variables / 从不硬编码，始终使用环境变量 |
| Struct tags / 结构体标签 | All fields must have both `json` and `gorm` tags / 所有字段必须同时有 `json` 和 `gorm` tag |

---

## Commit Messages / 提交信息

Follow the format defined in [`.trae/rules/git-commit-message.md`](.trae/rules/git-commit-message.md).
遵循 [`.trae/rules/git-commit-message.md`](.trae/rules/git-commit-message.md) 中定义的格式。

```
<type>: <title in Simplified Chinese, max 50 chars>

改了什么：
- Specific description of what changed

为什么改：
- Reason for the change
```

**Valid types / 有效类型:** `feat` `fix` `refactor` `test` `chore` `docs` `perf`

---

## Pull Request Process / PR 流程

1. Fill in the PR template completely / 完整填写 PR 模板。
2. Ensure all CI checks pass / 确保所有 CI 检查通过。
3. Request review from `@ctkqiang` / 请求 `@ctkqiang` 审查。
4. Address all review comments before merge / 在合并前处理所有审查意见。

### PR Checklist / PR 检查清单

- [ ] `go fmt ./...` run / 已运行
- [ ] `go vet ./...` passing / 通过
- [ ] `make lint` passing / 通过
- [ ] `make test` passing / 通过
- [ ] No hardcoded credentials / 无硬编码凭证
- [ ] Exported functions have Simplified Chinese comments / 导出函数有简体中文注释
- [ ] `README.md` and `README_zh.md` updated if needed / 按需更新 README
- [ ] Commit message follows the format / 提交信息符合规范

---

## Reporting Issues / 报告问题

- **Bug reports / 缺陷报告**: Use the [Bug Report template](/.github/ISSUE_TEMPLATE/bug_report.md).
- **Feature requests / 功能请求**: Use the [Feature Request template](/.github/ISSUE_TEMPLATE/feature_request.md).
- **Security vulnerabilities / 安全漏洞**: See [SECURITY.md](./SECURITY.md) — do NOT open a public issue.
  安全漏洞请参阅 [SECURITY.md](./SECURITY.md)，**请勿**公开提 Issue。

---

## Project Structure / 项目结构

```
NezhaCyberMCP/
├── main.go                        # Entry point / 入口
├── internal/
│   ├── functions/actions.go       # MCP tool implementations / MCP 工具实现
│   ├── services/mcp.go            # MCPServer & tool registration / 工具注册
│   ├── services/circl.go          # CIRCL CVE scraper / CIRCL 采集
│   ├── services/mycert.go         # MyCERT scraper / MyCERT 爬虫
│   ├── services/github.go         # GitHub Advisory client / GitHub 客户端
│   ├── services/database.go       # DB connection / 数据库连接
│   ├── model/                     # GORM models / 数据模型
│   ├── repository/                # Repository layer / 仓库层
│   ├── job/advisory_job.go        # cron scheduler / 定时调度
│   └── utilities/                 # Logger, AWS utils, config loader
├── test/                          # All tests / 所有测试
├── docs/                          # PlantUML source diagrams / 架构图源文件
├── out/                           # Rendered diagram images / 渲染图表
└── landing/                       # Nuxt.js landing page / 官网
```

---

**License / 许可证:** MIT — Copyright (c) 2026 ctkqiang
