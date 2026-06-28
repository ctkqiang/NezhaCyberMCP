---
name: Feature Request / 功能请求
about: Suggest a new feature or improvement for 哪吒网络安全 MCP (NezhaCyberMCP)
title: '[FEAT] '
labels: enhancement
assignees: ctkqiang
---

## Feature Description / 功能描述

<!-- A clear and concise description of the feature you are requesting -->
<!-- 请清晰简洁地描述你请求的功能 -->

## Problem Statement / 问题陈述

<!-- Is your feature request related to a problem? Describe it. -->
<!-- 你的功能请求是否与某个问题相关？请描述。 -->
<!-- e.g. "I need to query CVEs by EPSS score but there is no tool for it." -->
<!-- 例如："我需要按 EPSS 评分查询 CVE，但目前没有对应工具。" -->

## Proposed Solution / 建议方案

<!-- Describe the solution you would like -->
<!-- 描述你希望的解决方案 -->

## Feature Category / 功能分类

<!-- Mark the most relevant category / 勾选最相关的分类 -->

- [ ] New MCP tool / 新 MCP 工具 (registered in `internal/services/mcp.go` + implemented in `internal/functions/actions.go`)
- [ ] New data source integration / 新数据源集成 (new service in `internal/services/`, model in `internal/model/`, repository in `internal/repository/`)
- [ ] Enhancement to existing MCP tool / 增强现有 MCP 工具
- [ ] Database / query performance improvement / 数据库与查询性能优化
- [ ] New runtime mode or deployment option / 新运行时模式或部署选项
- [ ] Landing page improvement / 官网改进 (`landing/`)
- [ ] Documentation improvement / 文档改进 (`README.md` / `README_zh.md` / `docs/`)
- [ ] CI/CD or build tooling / CI/CD 与构建工具 (`Makefile`, `.github/workflows/`)
- [ ] Other / 其他: <!-- describe / 描述 -->

## Proposed MCP Tool (if applicable) / 建议的 MCP 工具（如适用）

<!-- If this feature involves a new MCP tool, describe its interface -->
<!-- 若此功能涉及新 MCP 工具，请描述其接口 -->

**Tool name / 工具名:** <!-- e.g. get_epss -->
**Description / 描述:** <!-- one-line description / 一句话描述 -->
**Input parameters / 输入参数:**

```json
{
  "cve_id": "string",
  "min_score": "number"
}
```

**Expected output / 预期输出:**

```json
{
  "cve_id": "CVE-2021-44228",
  "epss_score": 0.975,
  "percentile": 0.999
}
```

## Proposed Data Source (if applicable) / 建议的数据源（如适用）

<!-- If this feature requires a new external data source -->
<!-- 若此功能需要新的外部数据源 -->

**Source name / 数据源名称:** <!-- e.g. CISA KEV, EPSS, NVD -->
**Endpoint / 端点:** <!-- e.g. https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json -->
**Authentication / 认证方式:** <!-- None / API key / OAuth -->
**Data format / 数据格式:** <!-- JSON / CSV / HTML -->
**Sync frequency / 同步频率:** <!-- e.g. every 6 hours / 每 6 小时 -->

## Alternatives Considered / 已考虑的替代方案

<!-- Describe any alternative solutions or features you have considered -->
<!-- 描述你考虑过的其他解决方案或功能 -->

## Implementation Notes / 实现说明

<!-- Optional: any technical hints, constraints, or design considerations -->
<!-- 可选：技术提示、约束条件或设计考量 -->

<!--
Implementation must follow project conventions / 实现必须遵循项目规范：
- Repository pattern: BulkUpsert + CreateInBatches(100) / 仓库模式：BulkUpsert + CreateInBatches(100)
- New model must include json and gorm tags per code-format.md / 新模型必须包含 json 和 gorm tag
- New service must use utilities.LogStart/LogSuccess/LogError / 新服务必须使用规范日志函数
- New tool must be registered with JSON Schema validation in mcp.go / 新工具必须在 mcp.go 中注册并配置 JSON Schema 验证
- All exported function comments in Simplified Chinese / 所有导出函数注释使用简体中文
-->

## Additional Context / 补充说明

<!-- Any other context, screenshots, references, or related issues -->
<!-- 其他上下文、截图、参考资料或关联 Issue -->

<!--
Related placeholder tools already registered (not yet implemented) / 已注册但尚未实现的占位工具：
- get_kev_status — CISA Known Exploited Vulnerabilities / CISA 已知被利用漏洞目录
- get_epss       — EPSS Exploit Prediction Scoring System / EPSS 漏洞利用预测评分系统
- prioritize     — CVSS + EPSS + KEV composite scoring / CVSS + EPSS + KEV 综合评分排序
- match_sbom     — CycloneDX / SPDX SBOM parsing / CycloneDX / SPDX SBOM 解析
-->
