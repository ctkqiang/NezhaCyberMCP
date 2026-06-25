package model

import (
	"encoding/json"
	"time"
)

// CirclCVE 表示从 CIRCL Vulnerability-Lookup API 抓取的单条 CVE 漏洞记录。
// 该结构体同时作为 GORM 数据模型，映射到数据库表 circl_cves。
//
// 字段说明：
//   - CVEID           : CVE 唯一标识符，作为主键（如 CVE-2021-44228）
//   - State           : CVE 状态，枚举值：PUBLISHED | REJECTED | RESERVED
//   - AssignerOrgID   : 分配机构的组织 UUID
//   - AssignerShort   : 分配机构的简短名称（如 apache）
//   - Title           : 漏洞标题（来自 CNA 容器）
//   - Description     : 漏洞详细描述（英文，长文本）
//   - Severity        : 严重程度，从 metrics 字段归一化提取（critical/high/medium/low/unknown）
//   - CWEIDs          : 关联的 CWE 编号列表，以 JSON 数组形式存储
//   - AffectedJSON    : 受影响软件包列表，以 JSON 数组形式存储
//   - ReferencesJSON  : 参考链接列表，以 JSON 数组形式存储
//   - DatePublished   : CVE 首次发布时间，可为空
//   - DateUpdated     : CVE 最后更新时间，可为空
//   - DateReserved    : CVE ID 预留时间，可为空
//   - ScrapedAt       : 本条记录最后一次被抓取并写入数据库的时间
type CirclCVE struct {
	CVEID          string          `json:"cve_id"          gorm:"primaryKey;column:cve_id"`
	State          string          `json:"state"           gorm:"column:state"`
	AssignerOrgID  string          `json:"assigner_org_id" gorm:"column:assigner_org_id"`
	AssignerShort  string          `json:"assigner_short"  gorm:"column:assigner_short"`
	Title          string          `json:"title"           gorm:"column:title;type:text"`
	Description    string          `json:"description"     gorm:"column:description;type:text"`
	Severity       string          `json:"severity"        gorm:"column:severity"`
	CWEIDs         json.RawMessage `json:"cwe_ids"         gorm:"column:cwe_ids;type:text"`
	AffectedJSON   json.RawMessage `json:"affected"        gorm:"column:affected;type:text"`
	ReferencesJSON json.RawMessage `json:"references"      gorm:"column:references;type:text"`
	DatePublished  *time.Time      `json:"date_published"  gorm:"column:date_published"`
	DateUpdated    *time.Time      `json:"date_updated"    gorm:"column:date_updated"`
	DateReserved   *time.Time      `json:"date_reserved"   gorm:"column:date_reserved"`
	ScrapedAt      time.Time       `json:"scraped_at"      gorm:"column:scraped_at;autoUpdateTime"`
}

// CirclVulnerabilityListItem 是 CIRCL API /vulnerability/ 列表端点返回的单条摘要记录。
type CirclVulnerabilityListItem struct {
	VulnID string `json:"vuln_id"`
}

// CirclVulnerabilityListResponse 是 CIRCL API /vulnerability/ 列表端点的完整响应结构。
type CirclVulnerabilityListResponse struct {
	Page    int                          `json:"page"`
	PerPage int                          `json:"per_page"`
	Total   int                          `json:"total"`
	Data    []CirclVulnerabilityListItem `json:"data"`
}
