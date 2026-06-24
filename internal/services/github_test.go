package services

import (
	"encoding/json"
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

// TestDefaultConfig 验证 defaultConfig 返回的默认配置值与预定义常量一致。
func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.RequestTimeout != defaultRequestTimeout {
		t.Errorf("RequestTimeout = %v; 期望 %v", cfg.RequestTimeout, defaultRequestTimeout)
	}
	if cfg.PerPage != defaultPerPage {
		t.Errorf("PerPage = %d; 期望 %d", cfg.PerPage, defaultPerPage)
	}
	if cfg.RetryMax != defaultRetryMax {
		t.Errorf("RetryMax = %d; 期望 %d", cfg.RetryMax, defaultRetryMax)
	}
	if cfg.RetryBackoff != defaultRetryBackoff {
		t.Errorf("RetryBackoff = %v; 期望 %v", cfg.RetryBackoff, defaultRetryBackoff)
	}
	if cfg.MaxPages != 0 {
		t.Errorf("MaxPages = %d; 期望 0（不限制页数）", cfg.MaxPages)
	}
	if cfg.Token != "" {
		t.Errorf("Token 默认值应为空字符串，实际得到 %q", cfg.Token)
	}
}

// TestNewGithubAdvisoryService_ConfigOverride 验证传入非 nil 配置时，
// 各字段能正确覆盖默认值。
func TestNewGithubAdvisoryService_ConfigOverride(t *testing.T) {
	override := &AdvisoryScraperConfig{
		MaxPages:       5,
		RequestTimeout: 10 * time.Second,
		PerPage:        50,
		RetryMax:       3,
		RetryBackoff:   4 * time.Second,
		Token:          "ghp_test_token",
	}

	svc := NewGithubAdvisoryService(nil, override)

	if svc.config.MaxPages != 5 {
		t.Errorf("MaxPages = %d; 期望 5", svc.config.MaxPages)
	}
	if svc.config.RequestTimeout != 10*time.Second {
		t.Errorf("RequestTimeout = %v; 期望 10s", svc.config.RequestTimeout)
	}
	if svc.config.PerPage != 50 {
		t.Errorf("PerPage = %d; 期望 50", svc.config.PerPage)
	}
	if svc.config.RetryMax != 3 {
		t.Errorf("RetryMax = %d; 期望 3", svc.config.RetryMax)
	}
	if svc.config.Token != "ghp_test_token" {
		t.Errorf("Token = %q; 期望 ghp_test_token", svc.config.Token)
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
	if svc.config.PerPage != def.PerPage {
		t.Errorf("PerPage = %d; 期望 %d", svc.config.PerPage, def.PerPage)
	}
	if svc.config.RetryMax != def.RetryMax {
		t.Errorf("RetryMax = %d; 期望 %d", svc.config.RetryMax, def.RetryMax)
	}
}

// TestNewGithubAdvisoryService_PerPageCap 验证 PerPage 超过 100 时不会被接受，
// 保持默认值（GitHub API 上限为 100）。
func TestNewGithubAdvisoryService_PerPageCap(t *testing.T) {
	svc := NewGithubAdvisoryService(nil, &AdvisoryScraperConfig{PerPage: 200})
	if svc.config.PerPage != defaultPerPage {
		t.Errorf("PerPage 超过 100 时应回退到默认值 %d，实际得到 %d", defaultPerPage, svc.config.PerPage)
	}
}

// TestParseAPIAdvisory_Valid 验证 parseAPIAdvisory 能正确解析合法的 API JSON 响应。
func TestParseAPIAdvisory_Valid(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	cveID := "CVE-2024-12345"
	raw := map[string]interface{}{
		"ghsa_id":         "GHSA-test-0001-aaaa",
		"cve_id":          cveID,
		"url":             "https://api.github.com/advisories/GHSA-test-0001-aaaa",
		"html_url":        "https://github.com/advisories/GHSA-test-0001-aaaa",
		"summary":         "测试漏洞摘要",
		"description":     "详细描述",
		"type":            "reviewed",
		"severity":        "high",
		"published_at":    now.Format(time.RFC3339),
		"updated_at":      now.Format(time.RFC3339),
		"withdrawn_at":    nil,
		"vulnerabilities": []interface{}{},
		"references":      []interface{}{},
	}

	b, _ := json.Marshal(raw)
	adv, err := parseAPIAdvisory(b)
	if err != nil {
		t.Fatalf("parseAPIAdvisory 失败: %v", err)
	}
	if adv.GHSAID != "GHSA-test-0001-aaaa" {
		t.Errorf("GHSAID = %q; 期望 GHSA-test-0001-aaaa", adv.GHSAID)
	}
	if adv.Severity != "high" {
		t.Errorf("Severity = %q; 期望 high", adv.Severity)
	}
	if adv.Summary != "测试漏洞摘要" {
		t.Errorf("Summary = %q; 期望 测试漏洞摘要", adv.Summary)
	}
}

// TestParseAPIAdvisory_EmptyGHSAID 验证 GHSA ID 为空时返回错误。
func TestParseAPIAdvisory_EmptyGHSAID(t *testing.T) {
	raw := json.RawMessage(`{"ghsa_id":"","summary":"test"}`)
	_, err := parseAPIAdvisory(raw)
	if err == nil {
		t.Error("期望返回错误，实际返回 nil")
	}
}

// TestParseAPIAdvisory_NullVulnerabilities 验证 vulnerabilities 为 null 时
// 被替换为空 JSON 数组，而非写入 NULL。
func TestParseAPIAdvisory_NullVulnerabilities(t *testing.T) {
	raw := json.RawMessage(`{
		"ghsa_id": "GHSA-null-vuln-test",
		"summary": "test",
		"vulnerabilities": null,
		"references": null
	}`)
	adv, err := parseAPIAdvisory(raw)
	if err != nil {
		t.Fatalf("parseAPIAdvisory 失败: %v", err)
	}
	if string(adv.Vulnerabilities) != "[]" {
		t.Errorf("Vulnerabilities = %s; 期望 []", adv.Vulnerabilities)
	}
	if string(adv.References) != "[]" {
		t.Errorf("References = %s; 期望 []", adv.References)
	}
}

// TestParseAPIAdvisory_InvalidJSON 验证非法 JSON 输入时返回错误。
func TestParseAPIAdvisory_InvalidJSON(t *testing.T) {
	_, err := parseAPIAdvisory(json.RawMessage(`{not valid json`))
	if err == nil {
		t.Error("期望返回 JSON 解析错误，实际返回 nil")
	}
}

// TestNextLink_MultipleRels 验证 Link 头包含多个 rel 时能精确提取 next。
func TestNextLink_MultipleRels(t *testing.T) {
	header := `<https://api.github.com/advisories?page=1>; rel="prev", ` +
		`<https://api.github.com/advisories?page=3>; rel="next", ` +
		`<https://api.github.com/advisories?page=100>; rel="last"`
	got := nextLink(header)
	want := "https://api.github.com/advisories?page=3"
	if got != want {
		t.Errorf("nextLink = %q; 期望 %q", got, want)
	}
}

// TestNextLink_ExtraSpaces 验证 Link 头中 rel 前后有多余空格时仍能正确解析。
func TestNextLink_ExtraSpaces(t *testing.T) {
	header := `<https://api.github.com/advisories?page=2>;  rel="next"`
	got := nextLink(header)
	if !strings.Contains(got, "page=2") {
		t.Errorf("nextLink = %q; 期望包含 page=2", got)
	}
}
