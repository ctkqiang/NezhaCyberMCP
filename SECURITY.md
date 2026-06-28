# Security Policy / 安全政策

## Supported Versions / 支持的版本

| Version / 版本 | Supported / 支持状态 |
|----------------|----------------------|
| `main` branch  | Active / 积极维护    |
| Older releases | Not supported / 不再支持 |

---

## Reporting a Vulnerability / 报告安全漏洞

**Do NOT open a public GitHub Issue for security vulnerabilities.**
**请勿通过公开的 GitHub Issue 报告安全漏洞。**

If you discover a security vulnerability in 哪吒网络安全 MCP (NezhaCyberMCP), please report it responsibly using one of the following methods:
如果你在哪吒网络安全 MCP (NezhaCyberMCP) 中发现安全漏洞，请通过以下方式负责任地披露：

### Preferred Method / 首选方式

Use GitHub's private vulnerability reporting:
使用 GitHub 私密漏洞报告功能：

**GitHub → Security → Report a vulnerability**

### Alternative / 备选方式

Email / 邮件: Contact the maintainer via GitHub profile.
通过 GitHub 个人主页联系维护者。

---

## What to Include / 报告内容

Please include the following in your report:
报告中请包含以下信息：

- **Description / 描述**: A clear description of the vulnerability and its potential impact.
  漏洞的清晰描述及其潜在影响。
- **Steps to reproduce / 复现步骤**: Detailed steps to reproduce the issue.
  详细的复现步骤。
- **Affected component / 受影响组件**: Which part of the codebase is affected.
  代码库中受影响的部分。
- **Suggested fix / 建议修复方案**: If you have a suggested fix, please include it.
  如有建议的修复方案，请一并提供。

---

## Response Timeline / 响应时间

| Action / 操作 | Timeline / 时间 |
|---------------|-----------------|
| Acknowledgement / 确认收到 | Within 48 hours / 48 小时内 |
| Initial assessment / 初步评估 | Within 7 days / 7 天内 |
| Fix & disclosure / 修复与披露 | Within 90 days / 90 天内 |

---

## Security Design / 安全设计

哪吒网络安全 MCP implements the following security measures:
哪吒网络安全 MCP 实现了以下安全措施：

### Credential Management / 凭证管理

- All credentials (database passwords, API tokens, AWS keys) are loaded exclusively from environment variables or AWS Secrets Manager — never hardcoded.
  所有凭证（数据库密码、API 令牌、AWS 密钥）仅从环境变量或 AWS Secrets Manager 加载，从不硬编码。
- AWS credentials are validated via `utilities.IsRunInAWS()` before use, rejecting placeholder values.
  AWS 凭证在使用前通过 `utilities.IsRunInAWS()` 校验，拒绝占位符值。
- Sensitive values are masked via `utilities.Mask()` before being written to logs.
  敏感值在写入日志前通过 `utilities.Mask()` 脱敏。

### Input Validation / 输入验证

- All MCP tool inputs are validated against JSON Schema by the `go-sdk/mcp` library before reaching application logic.
  所有 MCP 工具输入在到达应用逻辑前，由 `go-sdk/mcp` 库根据 JSON Schema 验证。
- SQL injection is prevented by GORM's parameterized query builder — raw SQL string concatenation is prohibited.
  SQL 注入由 GORM 的参数化查询构建器防止，禁止原始 SQL 字符串拼接。

### Transport Security / 传输安全

- **Local stdio mode**: Communication occurs over local process pipes — no network exposure.
  本地 stdio 模式：通信通过本地进程管道，无网络暴露。
- **Lambda SSE HTTP mode**: TLS termination is handled by AWS Lambda Function URL — the internal HTTP server never directly faces the internet.
  Lambda SSE HTTP 模式：TLS 终止由 AWS Lambda Function URL 处理，内部 HTTP 服务器从不直接暴露于互联网。

### Data Privacy / 数据隐私

- The database stores only publicly available CVE data from CIRCL, MyCERT, and GitHub Advisory Database.
  数据库仅存储来自 CIRCL、MyCERT 和 GitHub Advisory Database 的公开 CVE 数据。
- No personally identifiable information (PII), user sessions, or authentication tokens are persisted.
  不持久化任何个人身份信息（PII）、用户会话或认证令牌。

---

## Known Limitations / 已知限制

- MyCERT data is collected via HTML scraping. Changes to the MyCERT portal structure may affect data collection but do not introduce security vulnerabilities.
  MyCERT 数据通过 HTML 爬取收集。MyCERT 门户结构变化可能影响数据采集，但不会引入安全漏洞。
- The GitHub Advisory sync feature is disabled by default (`IsGithubAdvisoryTurnOn = false`) and requires a valid `GITHUB_TOKEN` to enable.
  GitHub Advisory 同步功能默认禁用（`IsGithubAdvisoryTurnOn = false`），启用需要有效的 `GITHUB_TOKEN`。

---

## License / 许可证

MIT License — Copyright (c) 2026 ctkqiang
