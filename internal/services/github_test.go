package services

import (
	"strings"
	"testing"
	"time"
)

// TestNextLink 验证 nextLink 函数能正确从 HTTP Link 响应头中提取 rel="next" 的 URL。
func TestNextLink(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "包含 next 链接",
			header: `<https://api.github.com/advisories?page=2>; rel="next", <https://api.github.com/advisories?page=10>; rel="last"`,
			want:   "https://api.github.com/advisories?page=2",
		},
		{
			name:   "不含 next 链接",
			header: `<https://api.github.com/advisories?page=1>; rel="prev"`,
			want:   "",
		},
		{
			name:   "空响应头",
			header: "",
			want:   "",
		},
		{
			name:   "仅含 last 链接",
			header: `<https://api.github.com/advisories?page=10>; rel="last"`,
			want:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := nextLink(tc.header)
			if got != tc.want {
				t.Errorf("nextLink(%q) = %q; 期望 %q", tc.header, got, tc.want)
			}
		})
	}
}

// TestNormaliseSeverity 验证 normaliseSeverity 函数能将各种原始标签文本
// 正确映射为 GitHub Advisory API 规范的严重程度枚举值。
func TestNormaliseSeverity(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"critical", "critical"},
		{"CRITICAL", "critical"},
		{"high", "high"},
		{"HIGH", "high"},
		{"moderate", "medium"},  // GitHub 页面使用 "moderate"，API 规范为 "medium"
		{"medium", "medium"},
		{"MEDIUM", "medium"},
		{"low", "low"},
		{"LOW", "low"},
		{"", "unknown"},         // 空字符串应返回 unknown
		{"informational", "unknown"}, // 不在枚举范围内的值应返回 unknown
		{"  High  ", "high"},    // 前后空格应被正确去除
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := normaliseSeverity(tc.input)
			if got != tc.want {
				t.Errorf("normaliseSeverity(%q) = %q; 期望 %q", tc.input, got, tc.want)
			}
		})
	}
}

// TestParseRelativeTime 验证 parseRelativeTime 函数能正确解析 RFC 3339 格式的时间字符串，
// 并在输入为空或格式非法时返回 nil。
func TestParseRelativeTime(t *testing.T) {
	t.Run("合法的 RFC3339 字符串", func(t *testing.T) {
		raw := "2024-03-15T10:30:00Z"
		got := parseRelativeTime(raw)
		if got == nil {
			t.Fatal("期望返回非 nil 的 *time.Time")
		}
		want := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("解析结果 %v; 期望 %v", got, want)
		}
	})

	t.Run("空字符串应返回 nil", func(t *testing.T) {
		if got := parseRelativeTime(""); got != nil {
			t.Errorf("期望 nil，实际得到 %v", got)
		}
	})

	t.Run("非法格式应返回 nil", func(t *testing.T) {
		if got := parseRelativeTime("not-a-date"); got != nil {
			t.Errorf("期望 nil，实际得到 %v", got)
		}
	})

	t.Run("带时区偏移的 RFC3339 字符串", func(t *testing.T) {
		raw := "2024-06-01T08:00:00+08:00"
		got := parseRelativeTime(raw)
		if got == nil {
			t.Fatal("期望返回非 nil 的 *time.Time")
		}
		if got.IsZero() {
			t.Error("期望返回非零时间")
		}
	})
}

// TestDefaultConfig 验证 defaultConfig 返回的默认配置值与预定义常量一致。
func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.RequestTimeout != defaultRequestTimeout {
		t.Errorf("RequestTimeout = %v; 期望 %v", cfg.RequestTimeout, defaultRequestTimeout)
	}
	if cfg.Parallelism != defaultParallelism {
		t.Errorf("Parallelism = %d; 期望 %d", cfg.Parallelism, defaultParallelism)
	}
	if cfg.RateLimit != defaultRateLimit {
		t.Errorf("RateLimit = %v; 期望 %v", cfg.RateLimit, defaultRateLimit)
	}
	if cfg.MaxPages != 0 {
		t.Errorf("MaxPages = %d; 期望 0（不限制页数）", cfg.MaxPages)
	}
}

// TestNewGithubAdvisoryService_ConfigOverride 验证传入非 nil 配置时，
// 各字段能正确覆盖默认值。
func TestNewGithubAdvisoryService_ConfigOverride(t *testing.T) {
	override := &AdvisoryScraperConfig{
		MaxPages:       5,
		RequestTimeout: 10 * time.Second,
		Parallelism:    1,
		RateLimit:      2 * time.Second,
	}

	svc := NewGithubAdvisoryService(nil, override)

	if svc.config.MaxPages != 5 {
		t.Errorf("MaxPages = %d; 期望 5", svc.config.MaxPages)
	}
	if svc.config.RequestTimeout != 10*time.Second {
		t.Errorf("RequestTimeout = %v; 期望 10s", svc.config.RequestTimeout)
	}
	if svc.config.Parallelism != 1 {
		t.Errorf("Parallelism = %d; 期望 1", svc.config.Parallelism)
	}
}

// TestNewGithubAdvisoryService_NilConfigUsesDefaults 验证传入 nil 配置时，
// 服务使用 defaultConfig() 中的默认值。
func TestNewGithubAdvisoryService_NilConfigUsesDefaults(t *testing.T) {
	svc := NewGithubAdvisoryService(nil, nil)
	def := defaultConfig()

	if svc.config.RequestTimeout != def.RequestTimeout {
		t.Errorf("RequestTimeout = %v; 期望 %v", svc.config.RequestTimeout, def.RequestTimeout)
	}
	if svc.config.Parallelism != def.Parallelism {
		t.Errorf("Parallelism = %d; 期望 %d", svc.config.Parallelism, def.Parallelism)
	}
}

// TestExtractGHSALink_NoMatch 验证当 href 不包含 GHSA 路径时，
// 路径分割逻辑不会错误地提取出 GHSA ID。
func TestExtractGHSALink_NoMatch(t *testing.T) {
	href := "/some/other/path"
	var ghsaID string
	parts := strings.Split(href, "/")
	for _, p := range parts {
		if strings.HasPrefix(p, "GHSA-") {
			ghsaID = p
			break
		}
	}
	if ghsaID != "" {
		t.Errorf("非公告路径不应提取出 GHSA ID，实际得到 %q", ghsaID)
	}
}

// TestExtractGHSALink_WithAbsoluteURL 验证从绝对 URL 中能正确提取 GHSA ID。
func TestExtractGHSALink_WithAbsoluteURL(t *testing.T) {
	href := "https://github.com/advisories/GHSA-1234-5678-abcd"
	parts := strings.Split(href, "/")
	var found string
	for _, p := range parts {
		if strings.HasPrefix(p, "GHSA-") {
			found = p
			break
		}
	}
	if found != "GHSA-1234-5678-abcd" {
		t.Errorf("从绝对 URL 提取 GHSA ID 失败，实际得到 %q", found)
	}
}

// TestExtractGHSALink_WithRelativeURL 验证从相对路径中能正确提取 GHSA ID。
func TestExtractGHSALink_WithRelativeURL(t *testing.T) {
	href := "/advisories/GHSA-abcd-efgh-ijkl"
	parts := strings.Split(href, "/")
	var found string
	for _, p := range parts {
		if strings.HasPrefix(p, "GHSA-") {
			found = p
			break
		}
	}
	if found != "GHSA-abcd-efgh-ijkl" {
		t.Errorf("从相对路径提取 GHSA ID 失败，实际得到 %q", found)
	}
}
