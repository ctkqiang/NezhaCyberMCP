---
name: Bug Report / 缺陷报告
about: Report a bug in 哪吒网络安全 MCP (NezhaCyberMCP)
title: '[BUG] '
labels: bug
assignees: ctkqiang
---

## Bug Description / 缺陷描述

<!-- A clear and concise description of what the bug is -->
<!-- 请清晰简洁地描述缺陷内容 -->

## Runtime Environment / 运行时环境

<!-- Mark the mode where the bug occurs / 标记缺陷发生的运行模式 -->

- [ ] Local mode / 本地模式 (`IS_LOCAL=true`, stdio MCP)
- [ ] AWS Lambda sync mode / Lambda 同步模式 (EventBridge cron)
- [ ] AWS Lambda MCP HTTP mode / Lambda MCP HTTP 模式 (`MCP_HTTP_MODE=true`)

**Go version / Go 版本:** <!-- e.g. go1.26.1 -->
**OS / 操作系统:** <!-- e.g. macOS 14, Ubuntu 22.04, Amazon Linux 2023 -->
**Database / 数据库:** <!-- e.g. PostgreSQL 16, SQLite, Amazon Aurora DSQL -->

## Affected Component / 受影响组件

<!-- Mark all that apply / 勾选所有适用项 -->

- [ ] MCP server startup / tool registration / MCP 服务器启动与工具注册 (`internal/services/mcp.go`)
- [ ] MCP tool execution / MCP 工具执行 (`internal/functions/actions.go`)
- [ ] CIRCL CVE data ingestion / CIRCL CVE 数据采集 (`internal/services/circl.go`)
- [ ] MyCERT advisory scraping / MyCERT 公告爬取 (`internal/services/mycert.go`)
- [ ] GitHub Advisory sync / GitHub Advisory 同步 (`internal/services/github.go`)
- [ ] Database connection / migration / 数据库连接与迁移 (`internal/services/database.go`)
- [ ] Repository layer / 仓库层 (`internal/repository/`)
- [ ] cron scheduler / cron 调度器 (`internal/job/advisory_job.go`)
- [ ] AWS credential validation / AWS 凭证校验 (`internal/utilities/aws.go`)
- [ ] Config loader / 配置加载器 (`internal/utilities/config_loader.go`)
- [ ] Logging / 日志 (`internal/utilities/logger.go`)
- [ ] Landing page / 官网 (`landing/`)
- [ ] Other / 其他: <!-- describe / 描述 -->

## Affected MCP Tool / 受影响的 MCP 工具

<!-- If the bug is in a specific MCP tool, name it / 若缺陷在特定 MCP 工具中，请填写工具名 -->

Tool name / 工具名: <!-- e.g. get_cve, search_cves, filter_by_severity, vuln_trends, match_inventory -->

## Steps to Reproduce / 复现步骤

1. Configure `.env` with / 配置 `.env`: `...`
2. Run / 运行: `...`
3. Call MCP tool / 调用 MCP 工具: `...`
4. Observe / 观察到: `...`

## Expected Behavior / 预期行为

<!-- What should have happened / 描述应该发生什么 -->

## Actual Behavior / 实际行为

<!-- What actually happened / 描述实际发生了什么 -->

## Error Output / Logs / 错误输出与日志

<!-- Paste relevant log output. Sensitive values (credentials, tokens) must be redacted. -->
<!-- 粘贴相关日志输出。敏感值（凭证、令牌）必须脱敏处理。 -->

```
paste log output here / 在此粘贴日志输出
```

## Environment Variables (redacted) / 环境变量（已脱敏）

<!-- List relevant env vars with values replaced by xxxxxxxxxx -->
<!-- 列出相关环境变量，值替换为 xxxxxxxxxx -->

```dotenv
IS_LOCAL=xxxxxxxxxx
DB_HOST=xxxxxxxxxx
DB_PORT=xxxxxxxxxx
DB_USER=xxxxxxxxxx
DB_NAME=xxxxxxxxxx
GITHUB_TOKEN=xxxxxxxxxx
AWS_ACCESS_KEY_ID=xxxxxxxxxx
AWS_SECRET_ACCESS_KEY=xxxxxxxxxx
SECRET_ARN=xxxxxxxxxx
S3_ENV_BUCKET=xxxxxxxxxx
```

## Additional Context / 补充说明

<!-- Any other context, screenshots, or related issues -->
<!-- 其他上下文、截图或关联 Issue -->
