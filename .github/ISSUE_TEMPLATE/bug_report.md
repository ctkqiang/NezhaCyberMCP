---
name: Bug Report
about: Report a bug in NezhaCyberMCP (哪吒网络安全 MCP)
title: '[BUG] '
labels: bug
assignees: ctkqiang
---

## Bug Description

<!-- A clear and concise description of what the bug is -->

## Runtime Environment

<!-- Mark the mode where the bug occurs -->

- [ ] Local mode (`IS_LOCAL=true`, stdio MCP)
- [ ] AWS Lambda sync mode (EventBridge cron)
- [ ] AWS Lambda MCP HTTP mode (`MCP_HTTP_MODE=true`)

**Go version:** <!-- e.g. go1.26.1 -->
**OS:** <!-- e.g. macOS 14, Ubuntu 22.04, Amazon Linux 2023 -->
**Database:** <!-- e.g. PostgreSQL 16, SQLite, Amazon Aurora DSQL -->

## Affected Component

<!-- Mark all that apply -->

- [ ] MCP server startup / tool registration (`internal/services/mcp.go`)
- [ ] MCP tool execution (`internal/functions/actions.go`)
- [ ] CIRCL CVE data ingestion (`internal/services/circl.go`)
- [ ] MyCERT advisory scraping (`internal/services/mycert.go`)
- [ ] GitHub Advisory sync (`internal/services/github.go`)
- [ ] Database connection / migration (`internal/services/database.go`)
- [ ] Repository layer (`internal/repository/`)
- [ ] cron scheduler (`internal/job/advisory_job.go`)
- [ ] AWS credential validation (`internal/utilities/aws.go`)
- [ ] Logging (`internal/utilities/logger.go`)
- [ ] Landing page (`landing/`)
- [ ] Other: <!-- describe -->

## Affected MCP Tool

<!-- If the bug is in a specific MCP tool, name it -->

Tool name: <!-- e.g. get_cve, search_cves, filter_by_severity, vuln_trends, match_inventory -->

## Steps to Reproduce

1. Configure `.env` with: `...`
2. Run: `...`
3. Call MCP tool: `...`
4. Observe: `...`

## Expected Behavior

<!-- What should have happened -->

## Actual Behavior

<!-- What actually happened -->

## Error Output / Logs

<!-- Paste relevant log output. Sensitive values (credentials, tokens) must be redacted. -->

```
paste log output here
```

## Environment Variables (redacted)

<!-- List relevant env vars with values replaced by xxxxxxxxxx -->

```dotenv
IS_LOCAL=xxxxxxxxxx
DB_HOST=xxxxxxxxxx
DB_PORT=xxxxxxxxxx
DB_USER=xxxxxxxxxx
DB_NAME=xxxxxxxxxx
GITHUB_TOKEN=xxxxxxxxxx
AWS_ACCESS_KEY_ID=xxxxxxxxxx
AWS_SECRET_ACCESS_KEY=xxxxxxxxxx
```

## Additional Context

<!-- Any other context, screenshots, or related issues -->
