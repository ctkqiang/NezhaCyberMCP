package model

import "time"

// MycertAdvisory 表示从 MyCERT（马来西亚计算机应急响应小组）网站抓取的单条安全公告记录。
// 该结构体同时作为 GORM 数据模型，映射到数据库表 mycert_advisories。
//
// 字段说明：
//   - AdvisoryID  : MyCERT 公告唯一标识符，作为主键（来自 URL 参数 id）
//   - Title       : 公告标题（长文本）
//   - PublishedAt : 公告发布时间，可为空（部分公告未标注日期）
//   - Category    : 公告分类（如 Advisory、Alert 等）
//   - Summary     : 公告摘要（长文本）
//   - DetailURL   : 公告详情页完整 URL
//   - FullContent : 公告详情页正文全文（长文本，FetchDetail=true 时填充）
//   - ScrapedAt   : 本条记录最后一次被抓取并写入数据库的时间，由 GORM autoUpdateTime 自动维护
type MycertAdvisory struct {
	AdvisoryID  string     `json:"advisory_id"  gorm:"primaryKey;column:advisory_id"`
	Title       string     `json:"title"        gorm:"column:title;type:text"`
	PublishedAt *time.Time `json:"published_at" gorm:"column:published_at"`
	Category    string     `json:"category"     gorm:"column:category"`
	Summary     string     `json:"summary"      gorm:"column:summary;type:text"`
	DetailURL   string     `json:"detail_url"   gorm:"column:detail_url"`
	FullContent string     `json:"full_content" gorm:"column:full_content;type:text"`
	ScrapedAt   time.Time  `json:"scraped_at"   gorm:"column:scraped_at;autoUpdateTime"`
}
