---
name: Feature Request
about: Suggest a new feature or improvement for NezhaCyberMCP (哪吒网络安全 MCP)
title: '[FEAT] '
labels: enhancement
assignees: ctkqiang
---

## Feature Description

<!-- A clear and concise description of the feature you are requesting -->

## Problem Statement

<!-- Is your feature request related to a problem? Describe it. -->
<!-- e.g. "I need to query CVEs by EPSS score but there is no tool for it." -->

## Proposed Solution

<!-- Describe the solution you would like -->

## Feature Category

<!-- Mark the most relevant category -->

- [ ] New MCP tool (registered in `internal/services/mcp.go` + implemented in `internal/functions/actions.go`)
- [ ] New data source integration (new service in `internal/services/`, model in `internal/model/`, repository in `internal/repository/`)
- [ ] Enhancement to existing MCP tool
- [ ] Database / query performance improvement
- [ ] New runtime mode or deployment option
- [ ] Landing page (`landing/`) improvement
- [ ] Documentation improvement (`README.md` / `README_zh.md` / `docs/`)
- [ ] CI/CD or build tooling (`Makefile`, `.github/workflows/`)
- [ ] Other: <!-- describe -->

## Proposed MCP Tool (if applicable)

<!-- If this feature involves a new MCP tool, describe its interface -->

**Tool name:** <!-- e.g. get_epss -->
**Description:** <!-- one-line description -->
**Input parameters:**

```json
{
  "cve_id": "string",
  "min_score": "number"
}
```

**Expected output:**

```json
{
  "cve_id": "CVE-2021-44228",
  "epss_score": 0.975,
  "percentile": 0.999
}
```

## Proposed Data Source (if applicable)

<!-- If this feature requires a new external data source -->

**Source name:** <!-- e.g. CISA KEV, EPSS, NVD -->
**Endpoint:** <!-- e.g. https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json -->
**Authentication:** <!-- None / API key / OAuth -->
**Data format:** <!-- JSON / CSV / HTML -->
**Sync frequency:** <!-- e.g. every 6 hours -->

## Alternatives Considered

<!-- Describe any alternative solutions or features you have considered -->

## Implementation Notes

<!-- Optional: any technical hints, constraints, or design considerations -->

<!-- Example:
- Should follow the existing repository pattern (BulkUpsert + CreateInBatches(100))
- New model must include json and gorm tags aligned per code-format.md
- New service must use utilities.LogStart/LogSuccess/LogError
- New tool must be registered with JSON Schema validation in mcp.go
-->

## Additional Context

<!-- Any other context, screenshots, references, or related issues -->

<!-- Related placeholder tools already registered (not yet implemented):
- get_kev_status — CISA Known Exploited Vulnerabilities
- get_epss       — EPSS Exploit Prediction Scoring System
- prioritize     — CVSS + EPSS + KEV composite scoring
- match_sbom     — CycloneDX / SPDX SBOM parsing
-->
