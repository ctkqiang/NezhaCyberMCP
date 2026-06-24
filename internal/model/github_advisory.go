package model

import (
	"encoding/json"
	"time"
)

// GithubAdvisory 表示从 GitHub Advisory Database 抓取的单条安全公告记录。
// 该结构体同时作为 GORM 数据模型，映射到数据库表 github_advisories。
//
// 字段说明：
//   - GHSAID          : GitHub Security Advisory 唯一标识符，作为主键（如 GHSA-xxxx-xxxx-xxxx）
//   - CVEID           : 对应的 CVE 编号，可为空（部分公告尚未分配 CVE）
//   - URL             : GitHub Advisory API 端点地址
//   - HTMLURL         : 公告在 GitHub 网站上的可读页面地址
//   - Summary         : 漏洞摘要（单行简短描述）
//   - Description     : 漏洞详细描述（长文本，存储为 TEXT 类型）
//   - Type            : 公告类型，枚举值：reviewed | unreviewed | malware
//   - Severity        : 严重程度，枚举值：low | medium | high | critical
//   - PublishedAt     : 公告首次发布时间，可为空
//   - UpdatedAt       : 公告最后更新时间，可为空
//   - WithdrawnAt     : 公告撤回时间，未撤回时为 nil
//   - Vulnerabilities : 受影响软件包列表，以 JSON 数组形式存储（PostgreSQL 使用 JSONB，其他数据库使用 TEXT）
//   - References      : 参考链接列表，以 JSON 数组形式存储
type GithubAdvisory struct {
	GHSAID          string          `json:"ghsa_id"         gorm:"primaryKey;column:ghsa_id"`
	CVEID           *string         `json:"cve_id"          gorm:"column:cve_id"`
	URL             string          `json:"url"             gorm:"column:url"`
	HTMLURL         string          `json:"html_url"        gorm:"column:html_url"`
	Summary         string          `json:"summary"         gorm:"column:summary"`
	Description     string          `json:"description"     gorm:"column:description;type:text"`
	Type            string          `json:"type"            gorm:"column:type"`     // reviewed | unreviewed | malware
	Severity        string          `json:"severity"        gorm:"column:severity"` // low | medium | high | critical
	PublishedAt     *time.Time      `json:"published_at"    gorm:"column:published_at"`
	UpdatedAt       *time.Time      `json:"updated_at"      gorm:"column:updated_at"`
	WithdrawnAt     *time.Time      `json:"withdrawn_at"    gorm:"column:withdrawn_at"`
	Vulnerabilities json.RawMessage `json:"vulnerabilities" gorm:"column:vulnerabilities;type:text"` // JSON 数组，存储受影响软件包
	References      json.RawMessage `json:"references"      gorm:"column:references;type:text"`      // JSON 数组，存储参考链接
}
