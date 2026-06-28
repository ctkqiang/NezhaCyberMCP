# Changelog / 变更日志

All notable changes to 哪吒网络安全 MCP (NezhaCyberMCP) are documented here.
哪吒网络安全 MCP (NezhaCyberMCP) 的所有重要变更均记录于此。

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
格式遵循 [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) 规范。

---

## [Unreleased] / 待发布

### Added / 新增

- Bilingual (EN/ZH) GitHub, Gitee, and Gitcode community files / 双语社区文件（GitHub、Gitee、Gitcode）
- `SECURITY.md` — responsible disclosure policy / 安全漏洞负责任披露政策
- `CONTRIBUTING.md` — contributor guide / 贡献者指南
- `.golangci.yml` — golangci-lint v2 configuration with 17 linters / golangci-lint v2 配置，启用 17 个 linter
- `.github/workflows/ci.yml` — automated CI pipeline (build / vet / lint / test / security scan / Lambda package) / 自动化 CI 流水线
- `internal/utilities/config_loader.go` — three-tier AWS config loading: S3 `.env` > Secrets Manager > system env / 三级 AWS 配置加载
- `env.sh` wrapper in Lambda zip for `SECRET_ARN` injection at cold-start / Lambda zip 中的 `env.sh` 包装脚本，冷启动时注入 `SECRET_ARN`
- S3 `.env` source support (`S3_ENV_BUCKET` / `S3_ENV_KEY`) / S3 `.env` 来源支持
- Placeholder value filtering in `Makefile` (`xxxxxxxxxx` treated as unset) / Makefile 中占位符值过滤

### Changed / 变更

- `Makefile` `lambda` / `lambda-arm64` targets now generate `env.sh` and include it in the zip / `lambda` 目标现在生成 `env.sh` 并打包进 zip
- `getSecretName()` no longer hardcodes any ARN — fully driven by environment variables / `getSecretName()` 不再硬编码任何 ARN
- All `.github` community files updated to bilingual format / 所有 `.github` 社区文件更新为双语格式

### Security / 安全

- Removed hardcoded Secrets Manager ARN from `config_loader.go` / 从 `config_loader.go` 移除硬编码的 Secrets Manager ARN
- Added `utilities.Mask()` to Secrets Manager log output / 在 Secrets Manager 日志输出中添加 `utilities.Mask()` 脱敏
- Lambda WARN log suppressed when running inside Lambda runtime (no `.env` file expected) / 在 Lambda 运行时内抑制 `.env` 文件缺失的 WARN 日志

---

## [0.1.0] — 2026-06-01

### Added / 新增

- Initial release / 初始版本
- MCP server with 18 registered tools / 注册 18 个工具的 MCP 服务器
- CIRCL CVE Search API integration / CIRCL CVE Search API 集成
- MyCERT advisory HTML scraper / MyCERT 公告 HTML 爬虫
- GitHub Advisory Database sync (disabled by default) / GitHub Advisory Database 同步（默认禁用）
- Three runtime modes: local stdio / Lambda sync / Lambda MCP HTTP / 三种运行时模式
- GORM repository layer with PostgreSQL, MySQL, SQLite, Amazon Aurora DSQL support / 支持四种数据库的 GORM 仓库层
- `cron/v3` scheduler for periodic data sync / 基于 `cron/v3` 的定时数据同步
- Non-blocking MCP server startup with hot DB injection / 非阻塞 MCP 服务器启动与热注入数据库连接
- Nuxt.js landing page with S3 + CloudFront deployment / Nuxt.js 官网，S3 + CloudFront 部署
- PlantUML architecture and sequence diagrams (EN + ZH) / PlantUML 架构图与时序图（中英双语）
