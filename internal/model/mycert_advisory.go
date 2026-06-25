package model

import "time"

// MycertAdvisory 表示从 MyCERT 门户抓取的单条安全公告记录。
// 该结构体同时作为 GORM 数据模型，映射到数据库表 mycert_advisories。
//
// 字段说明：
//   - AdvisoryID  : MyCERT 公告唯一编号，作为主键（如 MA-1458.062026）
//   - Title       : 公告完整标题
//   - PublishedAt : 公告发布日期，可为空
//   - Category    : 公告分类（如 Advisory、Alert 等）
//   - Summary     : 列表页摘要文本（截断版本）
//   - DetailURL   : 公告详情页完整 URL
//   - FullContent : 详情页正文全文（Markdown 格式）
//   - ScrapedAt   : 本条记录最后一次被抓取的时间戳
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
