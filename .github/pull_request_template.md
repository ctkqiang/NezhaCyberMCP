## Description

<!-- Provide a clear and concise summary of your changes -->
<!-- What problem does this PR solve? What feature does it add? -->

**Problem:**
Describe the issue or requirement this PR addresses.

**Solution:**
Explain how this PR solves the problem.

---

## Type of Change

<!-- Mark the relevant option with an "x" -->

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] Feature (non-breaking change that adds functionality)
- [ ] Refactor (code improvement without changing behavior)
- [ ] Documentation (documentation updates only)
- [ ] Performance (performance improvement)
- [ ] Security (security improvement or vulnerability fix)
- [ ] Chore (dependency update, linting, etc.)
- [ ] Breaking change (breaking API change)

---

## Related Issues

<!-- Link to GitHub issues -->

Closes: #<!-- issue number -->
Related to: #<!-- issue number -->

---

## Testing

### Test Plan

<!-- Describe how to test these changes -->

1. Set environment: `cp .env.example .env && vim .env`
2. Build the binary: `make build`
3. Run with MCP Inspector: `make run`
4. Verify expected tool output in the inspector

### Test Coverage

- [ ] Unit tests added/updated (`go test ./test/...`)
- [ ] Integration tests added/updated
- [ ] Manual testing completed with MCP Inspector

### Test Results

<!-- Paste test output -->

```
$ go test -v ./test/...
ok    nezha_cyber_mcp/test    1.234s
PASS
```

---

## Changes Overview

### Files Changed

<!-- List key files modified — use actual paths from this project -->

- `internal/functions/actions.go` — MCP tool implementations & DB query logic
- `internal/services/mcp.go` — MCPServer tool/resource/prompt registration
- `internal/model/*.go` — GORM data models
- `internal/repository/*.go` — Repository layer (upsert, batch write, query)
- `internal/services/circl.go` — CIRCL API client & scraper
- `internal/services/mycert.go` — MyCERT HTML scraper
- `internal/services/github.go` — GitHub Advisory API client
- `internal/job/advisory_job.go` — cron scheduler & sync orchestration
- `internal/utilities/logger.go` — Structured logger
- `internal/utilities/aws.go` — AWS environment detection & credential validation

### Lines Added/Removed

- **+N lines** added
- **-N lines** removed

---

## Security & Compliance

### Security Impact

- [ ] No security impact
- [ ] Handles sensitive data (API keys, credentials)
- [ ] Affects authentication/authorization (AWS IAM, GitHub token)
- [ ] Modifies database access patterns (GORM repository layer)
- [ ] Introduces external API calls (CIRCL / GitHub / MyCERT)
- [ ] Affects logging/audit trails (`utilities/logger.go`)

**Security Considerations:**

- All API credentials loaded from environment variables only (never hardcoded) [OK]
- AWS credentials validated via `utilities.IsRunInAWS()` before use [OK]
- Database queries use GORM parameterized builder — no raw SQL concatenation [OK]
- Sensitive values masked via `utilities.Mask()` before logging [OK]
- MCP tool inputs validated against JSON Schema by `go-sdk/mcp` [OK]

---

## Database Changes

- [ ] No database changes
- [ ] Schema migration required (GORM AutoMigrate via `advisory_job.go`)

### Migration Details

```go
// All migrations run via MigrateAll(ctx) in advisory_job.go
// GORM AutoMigrate handles schema changes automatically.
// Upsert pattern used for all bulk writes:
tx.Clauses(clause.OnConflict{UpdateAll: true}).
    CreateInBatches(items, 100).Error
```

**Supported databases:** PostgreSQL · MySQL · SQLite · Amazon Aurora DSQL

---

## MCP Tool Changes

<!-- If this PR adds, modifies, or removes MCP tools -->

- [ ] No MCP tool changes
- [ ] New tool registered in `internal/services/mcp.go`
- [ ] Existing tool logic modified in `internal/functions/actions.go`
- [ ] Tool removed (confirm no AI client depends on it)

**Tool Details:**

| Tool Name | Change | Description |
|-----------|--------|-------------|
| `tool_name` | added / modified / removed | Brief description |

---

## Performance Impact

### Benchmarks

**Data ingestion:** CIRCL pagination fetches 100 records per request with 500ms rate limiting.
**Batch writes:** `CreateInBatches(100)` wrapped in a transaction — atomic, idempotent.
**Memory Impact:** No significant change
**Database Load:** Describe any new queries or index requirements

---

## Deployment Notes

### Runtime Mode Affected

- [ ] Local mode (`IS_LOCAL=true`, stdio MCP)
- [ ] AWS Lambda sync mode (EventBridge cron, data ingestion only)
- [ ] AWS Lambda MCP HTTP mode (`MCP_HTTP_MODE=true`, SSE endpoint)

### Breaking Changes

- [ ] No breaking changes
- [ ] MCP tool interface changed (requires AI client reconfiguration)
- [ ] Environment variable added or renamed

**New Environment Variables:**

```dotenv
# Add to .env.example if introducing new variables
NEW_VAR=xxxxxxxxxx
```

### Dependencies

- [ ] No dependency changes
- [ ] Updated: `go get -u ./...`
- [ ] New dependency added to `go.mod`

**New Dependencies:**

```
module nezha_cyber_mcp — go 1.26.1
```

---

## Deployment Checklist

- [ ] Code follows project style guidelines (`code-format.md`)
- [ ] Self-review completed
- [ ] Comments added in Simplified Chinese for all exported functions
- [ ] Documentation updated (`README.md` / `README_zh.md`)
- [ ] No new warnings: `go vet ./...`
- [ ] Tests passing: `go test ./test/...`
- [ ] No hardcoded credentials or secrets
- [ ] All environment-specific values use environment variables
- [ ] Logging uses `utilities.LogStart/LogProgress/LogSuccess/LogError/LogWarn`
- [ ] Error handling uses `fmt.Errorf("operation: %w", err)` wrapping
- [ ] Database operations use `WithContext(ctx)`
- [ ] Bulk writes use `OnConflict{UpdateAll: true}` + `CreateInBatches(100)`

---

## CI/CD Pipeline

**Build Status:**

- [ ] GitHub Actions passing
- [ ] `go build ./...` succeeds
- [ ] `go test ./test/...` passing
- [ ] `go vet ./...` passing
- [ ] Landing page deployment passing (`deploy_landing_page.yml`)

---

## Review Checklist

### Code Review

- [ ] Logic is correct and handles edge cases
- [ ] Code follows naming conventions (`camelCase` private, `PascalCase` exported)
- [ ] Method receivers use single/double letter abbreviations (`r`, `svc`, etc.)
- [ ] No unnecessary complexity or premature abstraction
- [ ] Error handling is comprehensive — no `_ = err` without justification

### Testing

- [ ] Test cases cover happy path and failure cases
- [ ] Tests are isolated and stateless
- [ ] Test file placed in `/test/` directory

### Documentation

- [ ] Exported functions have Simplified Chinese doc comments
- [ ] `README.md` and `README_zh.md` updated if behaviour changes
- [ ] Architecture diagrams updated if component relationships change (`docs/`)

### Security

- [ ] No credentials hardcoded
- [ ] Sensitive data masked in logs via `utilities.Mask()`
- [ ] Input validation implemented at system boundaries
- [ ] SQL queries use GORM parameterized builder only
- [ ] External API calls use HTTPS

### Performance

- [ ] No goroutine leaks — all goroutines respect `ctx` cancellation
- [ ] Database queries use `WithContext(ctx)`
- [ ] Batch writes use `CreateInBatches(100)` inside a transaction
- [ ] No unnecessary allocations in hot paths

---

## Additional Context

### Related PRs

- Depends on: #[related-pr]
- Blocked by: #[blocking-pr]

### Known Limitations

- [ ] None
- Known issue: [Description]
- Future improvement: [Description]

---

## Author Self-Review

Before requesting review, I have:

- [ ] Run `go fmt ./...`
- [ ] Run `go vet ./...`
- [ ] Run `go test ./test/...` locally
- [ ] Verified no sensitive data in commits (check `.env` is in `.gitignore`)
- [ ] Reviewed my own diff for logic errors
- [ ] Updated `README.md` and/or `README_zh.md` if needed
- [ ] Confirmed all new constants have Simplified Chinese comments
- [ ] Confirmed commit message follows `git-commit-message.md` format

---

**Reviewer:** @ctkqiang
**License:** MIT — Copyright (c) 2026 ctkqiang
