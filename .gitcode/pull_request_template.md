## Description / 描述

<!-- Provide a clear and concise summary of your changes -->
<!-- 请简洁清晰地描述本次变更内容 -->

**Problem / 问题：**
Describe the issue or requirement this PR addresses.
描述本 PR 解决的问题或需求。

**Solution / 解决方案：**
Explain how this PR solves the problem.
说明本 PR 如何解决该问题。

---

## Type of Change / 变更类型

<!-- Mark the relevant option with an "x" / 在相关选项前填写 "x" -->

- [ ] Bug fix / 缺陷修复 (non-breaking change that fixes an issue)
- [ ] Feature / 新功能 (non-breaking change that adds functionality)
- [ ] Refactor / 重构 (code improvement without changing behavior)
- [ ] Documentation / 文档 (documentation updates only)
- [ ] Performance / 性能优化 (performance improvement)
- [ ] Security / 安全 (security improvement or vulnerability fix)
- [ ] Chore / 杂项 (dependency update, linting, etc.)
- [ ] Breaking change / 破坏性变更 (breaking API change)

---

## Related Issues / 关联 Issue

<!-- Link to issues / 关联 Issue 编号 -->

Closes / 关闭: #<!-- issue number -->
Related to / 关联: #<!-- issue number -->

---

## Testing / 测试

### Test Plan / 测试计划

<!-- Describe how to test these changes / 描述如何测试本次变更 -->

1. Set environment / 配置环境: `cp .env.example .env && vim .env`
2. Build the binary / 构建二进制: `make build`
3. Run with MCP Inspector / 使用 MCP Inspector 运行: `make run`
4. Verify expected tool output in the inspector / 在 Inspector 中验证工具输出

### Test Coverage / 测试覆盖

- [ ] Unit tests added/updated / 单元测试已添加或更新 (`go test ./test/...`)
- [ ] Integration tests added/updated / 集成测试已添加或更新
- [ ] Manual testing completed with MCP Inspector / 已通过 MCP Inspector 完成手动测试

### Test Results / 测试结果

<!-- Paste test output / 粘贴测试输出 -->

```
$ go test -v ./test/...
ok    nezha_cyber_mcp/test    1.234s
PASS
```

---

## Changes Overview / 变更概览

### Files Changed / 变更文件

<!-- List key files modified / 列出主要变更文件 -->

- `internal/functions/actions.go` — MCP tool implementations & DB query logic / MCP 工具实现与数据库查询逻辑
- `internal/services/mcp.go` — MCPServer tool/resource/prompt registration / MCPServer 工具、资源、提示注册
- `internal/model/*.go` — GORM data models / GORM 数据模型
- `internal/repository/*.go` — Repository layer (upsert, batch write, query) / 仓库层（Upsert、批量写入、查询）
- `internal/services/circl.go` — CIRCL API client & scraper / CIRCL API 客户端与爬虫
- `internal/services/mycert.go` — MyCERT HTML scraper / MyCERT HTML 爬虫
- `internal/services/github.go` — GitHub Advisory API client / GitHub Advisory API 客户端
- `internal/job/advisory_job.go` — cron scheduler & sync orchestration / cron 调度器与同步编排
- `internal/utilities/logger.go` — Structured logger / 结构化日志
- `internal/utilities/aws.go` — AWS environment detection & credential validation / AWS 环境检测与凭证校验

### Lines Added/Removed / 新增/删除行数

- **+N lines** added / 新增
- **-N lines** removed / 删除

---

## Security & Compliance / 安全与合规

### Security Impact / 安全影响

- [ ] No security impact / 无安全影响
- [ ] Handles sensitive data (API keys, credentials) / 涉及敏感数据（API 密钥、凭证）
- [ ] Affects authentication/authorization (AWS IAM, GitHub token) / 影响认证/授权
- [ ] Modifies database access patterns (GORM repository layer) / 修改数据库访问模式
- [ ] Introduces external API calls (CIRCL / GitHub / MyCERT) / 引入外部 API 调用
- [ ] Affects logging/audit trails (`utilities/logger.go`) / 影响日志/审计链路

**Security Considerations / 安全说明：**

- All API credentials loaded from environment variables only (never hardcoded) / 所有凭证仅从环境变量读取，从不硬编码 [OK]
- AWS credentials validated via `utilities.IsRunInAWS()` before use / AWS 凭证在使用前通过 `utilities.IsRunInAWS()` 校验 [OK]
- Database queries use GORM parameterized builder — no raw SQL concatenation / 数据库查询使用 GORM 参数化构建器，不拼接原始 SQL [OK]
- Sensitive values masked via `utilities.Mask()` before logging / 敏感值在写入日志前通过 `utilities.Mask()` 脱敏 [OK]
- MCP tool inputs validated against JSON Schema by `go-sdk/mcp` / MCP 工具输入由 `go-sdk/mcp` 根据 JSON Schema 验证 [OK]

---

## Database Changes / 数据库变更

- [ ] No database changes / 无数据库变更
- [ ] Schema migration required (GORM AutoMigrate via `advisory_job.go`) / 需要 Schema 迁移

### Migration Details / 迁移详情

```go
// All migrations run via MigrateAll(ctx) in advisory_job.go
// 所有迁移通过 advisory_job.go 中的 MigrateAll(ctx) 执行
// GORM AutoMigrate handles schema changes automatically.
// GORM AutoMigrate 自动处理 Schema 变更。
// Upsert pattern used for all bulk writes:
// 所有批量写入使用 Upsert 模式：
tx.Clauses(clause.OnConflict{UpdateAll: true}).
    CreateInBatches(items, 100).Error
```

**Supported databases / 支持的数据库:** PostgreSQL · MySQL · SQLite · Amazon Aurora DSQL

---

## MCP Tool Changes / MCP 工具变更

<!-- If this PR adds, modifies, or removes MCP tools / 若本 PR 新增、修改或删除了 MCP 工具 -->

- [ ] No MCP tool changes / 无 MCP 工具变更
- [ ] New tool registered in `internal/services/mcp.go` / 在 `mcp.go` 中注册了新工具
- [ ] Existing tool logic modified in `internal/functions/actions.go` / 修改了 `actions.go` 中的现有工具逻辑
- [ ] Tool removed (confirm no AI client depends on it) / 删除了工具（确认无 AI 客户端依赖）

**Tool Details / 工具详情：**

| Tool Name / 工具名 | Change / 变更              | Description / 描述           |
| ------------------ | -------------------------- | ---------------------------- |
| `tool_name`        | added / modified / removed | Brief description / 简要描述 |

---

## Performance Impact / 性能影响

### Benchmarks / 基准测试

**Data ingestion / 数据采集:** CIRCL pagination fetches 100 records per request with 500ms rate limiting. / CIRCL 分页每次请求获取 100 条记录，限速 500ms。
**Batch writes / 批量写入:** `CreateInBatches(100)` wrapped in a transaction — atomic, idempotent. / `CreateInBatches(100)` 包裹在事务中，原子且幂等。
**Memory Impact / 内存影响:** No significant change / 无显著变化
**Database Load / 数据库负载:** Describe any new queries or index requirements / 描述新增查询或索引需求

---

## Deployment Notes / 部署说明

### Runtime Mode Affected / 受影响的运行时模式

- [ ] Local mode / 本地模式 (`IS_LOCAL=true`, stdio MCP)
- [ ] AWS Lambda sync mode / Lambda 同步模式 (EventBridge cron, data ingestion only)
- [ ] AWS Lambda MCP HTTP mode / Lambda MCP HTTP 模式 (`MCP_HTTP_MODE=true`, SSE endpoint)

### Breaking Changes / 破坏性变更

- [ ] No breaking changes / 无破坏性变更
- [ ] MCP tool interface changed (requires AI client reconfiguration) / MCP 工具接口变更（需重新配置 AI 客户端）
- [ ] Environment variable added or renamed / 新增或重命名了环境变量

**New Environment Variables / 新增环境变量：**

```dotenv
# Add to .env.example if introducing new variables
# 若引入新变量，请同步更新 .env.example
NEW_VAR=xxxxxxxxxx
```

### Dependencies / 依赖变更

- [ ] No dependency changes / 无依赖变更
- [ ] Updated / 已更新: `go get -u ./...`
- [ ] New dependency added to `go.mod` / 在 `go.mod` 中新增了依赖

---

## Deployment Checklist / 部署检查清单

- [ ] Code follows project style guidelines (`code-format.md`) / 代码符合项目规范
- [ ] Self-review completed / 已完成自我审查
- [ ] Comments added in Simplified Chinese for all exported functions / 所有导出函数已添加简体中文注释
- [ ] Documentation updated (`README.md` / `README_zh.md`) / 已更新文档
- [ ] No new warnings: `go vet ./...` / 无新警告
- [ ] Tests passing: `go test ./test/...` / 测试通过
- [ ] No hardcoded credentials or secrets / 无硬编码凭证或密钥
- [ ] All environment-specific values use environment variables / 所有环境相关值均使用环境变量
- [ ] Logging uses `utilities.LogStart/LogProgress/LogSuccess/LogError/LogWarn` / 日志使用规范函数
- [ ] Error handling uses `fmt.Errorf("operation: %w", err)` wrapping / 错误处理使用 `fmt.Errorf` 包装
- [ ] Database operations use `WithContext(ctx)` / 数据库操作使用 `WithContext(ctx)`
- [ ] Bulk writes use `OnConflict{UpdateAll: true}` + `CreateInBatches(100)` / 批量写入使用规范模式

---

## CI/CD Pipeline / CI/CD 流水线

**Build Status / 构建状态：**

- [ ] Actions passing / Actions 通过
- [ ] `go build ./...` succeeds / 编译成功
- [ ] `go test ./test/...` passing / 测试通过
- [ ] `go vet ./...` passing / 静态检查通过
- [ ] Landing page deployment passing / 官网部署通过 (`deploy_landing_page.yml`)

---

## Review Checklist / 审查清单

### Code Review / 代码审查

- [ ] Logic is correct and handles edge cases / 逻辑正确，覆盖边界情况
- [ ] Code follows naming conventions (`camelCase` private, `PascalCase` exported) / 命名规范正确
- [ ] Method receivers use single/double letter abbreviations (`r`, `svc`, etc.) / 方法接收者使用单/双字母缩写
- [ ] No unnecessary complexity or premature abstraction / 无不必要的复杂度或过早抽象
- [ ] Error handling is comprehensive — no `_ = err` without justification / 错误处理完整

### Testing / 测试

- [ ] Test cases cover happy path and failure cases / 测试覆盖正常路径和失败路径
- [ ] Tests are isolated and stateless / 测试相互隔离且无状态
- [ ] Test file placed in `/test/` directory / 测试文件位于 `/test/` 目录

### Documentation / 文档

- [ ] Exported functions have Simplified Chinese doc comments / 导出函数有简体中文文档注释
- [ ] `README.md` and `README_zh.md` updated if behaviour changes / 行为变更时已更新 README
- [ ] Architecture diagrams updated if component relationships change (`docs/`) / 组件关系变更时已更新架构图

### Security / 安全

- [ ] No credentials hardcoded / 无硬编码凭证
- [ ] Sensitive data masked in logs via `utilities.Mask()` / 敏感数据已脱敏
- [ ] Input validation implemented at system boundaries / 系统边界已实现输入验证
- [ ] SQL queries use GORM parameterized builder only / SQL 查询仅使用 GORM 参数化构建器
- [ ] External API calls use HTTPS / 外部 API 调用使用 HTTPS

### Performance / 性能

- [ ] No goroutine leaks — all goroutines respect `ctx` cancellation / 无 goroutine 泄漏
- [ ] Database queries use `WithContext(ctx)` / 数据库查询使用 `WithContext(ctx)`
- [ ] Batch writes use `CreateInBatches(100)` inside a transaction / 批量写入在事务内使用 `CreateInBatches(100)`
- [ ] No unnecessary allocations in hot paths / 热路径无不必要内存分配

---

## Additional Context / 补充说明

### Related PRs / 关联 PR

- Depends on / 依赖: #[related-pr]
- Blocked by / 阻塞于: #[blocking-pr]

### Known Limitations / 已知限制

- [ ] None / 无
- Known issue / 已知问题: [Description / 描述]
- Future improvement / 未来改进: [Description / 描述]

---

## Author Self-Review / 作者自查

Before requesting review, I have / 提交审查前，我已完成：

- [ ] Run `go fmt ./...`
- [ ] Run `go vet ./...`
- [ ] Run `go test ./test/...` locally / 本地运行测试
- [ ] Verified no sensitive data in commits (check `.env` is in `.gitignore`) / 确认提交中无敏感数据
- [ ] Reviewed my own diff for logic errors / 已审查自己的 diff
- [ ] Updated `README.md` and/or `README_zh.md` if needed / 按需更新了 README
- [ ] Confirmed all new constants have Simplified Chinese comments / 确认所有新常量有简体中文注释
- [ ] Confirmed commit message follows `git-commit-message.md` format / 确认提交信息符合规范

---

**Reviewer / 审查人:** @ctkqiang
**License / 许可证:** MIT — Copyright (c) 2026 ctkqiang
