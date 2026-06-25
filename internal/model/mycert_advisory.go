package model

import "time"

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
