package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nezha_cyber_mcp/internal/model"
	"nezha_cyber_mcp/internal/repository"
	"nezha_cyber_mcp/internal/utilities"
	"regexp"
	"strconv"
	"time"
)

const (
	// ApiBase 是 GitHub Advisory REST API 的基础地址。
	ApiBase = "https://api.github.com/advisories"

	// UserAgent 是向 GitHub API 发起请求时使用的 User-Agent 标识。
	UserAgent = "advisory-sync/1.0"

	// component 是本服务在日志中使用的组件名称标识。
	component = "GithubAdvisoryService"

	// defaultRequestTimeout 是单次 HTTP 请求的默认超时时间。
	defaultRequestTimeout = 30 * time.Second

	// defaultPerPage 是每次 API 请求返回的最大条目数（GitHub 上限为 100）。
	defaultPerPage = 100

	// defaultRetryMax 是遇到 429 / 5xx 时的最大重试次数。
	defaultRetryMax = 5

	// defaultRetryBackoff 是首次重试前的等待基准时间，每次重试翻倍（指数退避）。
	defaultRetryBackoff = 2 * time.Second
)

// linkRe 用于解析 HTTP Link 响应头中的分页链接。
// 格式示例：<https://api.github.com/advisories?page=2>; rel="next"
var linkRe = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)

// nextLink 解析 GitHub 风格的 Link 响应头，返回 rel="next" 对应的 URL。
// 若不存在下一页链接，返回空字符串。
//
// 参数：
//   - header : HTTP Link 响应头的原始字符串
//
// 返回：
//   - string : 下一页的完整 URL，不存在时为空字符串
func nextLink(header string) string {
	for _, m := range linkRe.FindAllStringSubmatch(header, -1) {
		if m[2] == "next" {
			return m[1]
		}
	}
	return ""
}

// AdvisoryScraperConfig 保存 API 客户端的可调参数。
// 所有字段均有合理默认值，可通过 NewGithubAdvisoryService 的 cfg 参数覆盖。
type AdvisoryScraperConfig struct {
	// MaxPages 限制最多请求的页数，0 表示不限制（拉取全部数据）。
	MaxPages int

	// RequestTimeout 是单次 HTTP 请求的超时时间。
	RequestTimeout time.Duration

	// PerPage 是每页返回的条目数，最大 100。
	PerPage int

	// RetryMax 是遇到 429 / 5xx 时的最大重试次数。
	RetryMax int

	// RetryBackoff 是首次重试前的等待基准时间，每次重试翻倍。
	RetryBackoff time.Duration

	// Token 是 GitHub Personal Access Token，用于提升 API 速率限制。
	// 未认证请求限额为 60 次/小时；认证后为 5000 次/小时。
	Token string
}

// defaultConfig 返回一组合理的默认 API 客户端配置。
func defaultConfig() AdvisoryScraperConfig {
	return AdvisoryScraperConfig{
		MaxPages:       0,
		RequestTimeout: defaultRequestTimeout,
		PerPage:        defaultPerPage,
		RetryMax:       defaultRetryMax,
		RetryBackoff:   defaultRetryBackoff,
	}
}

// GithubAdvisoryService 负责通过 GitHub REST API 拉取安全公告并持久化到数据库。
type GithubAdvisoryService struct {
	repo   *repository.GithubAdvisoryRepository
	config AdvisoryScraperConfig
	client *http.Client
}

// NewGithubAdvisoryService 构造 GithubAdvisoryService 实例。
// cfg 为 nil 时使用 defaultConfig() 中的默认值；非 nil 时仅覆盖非零字段。
//
// 参数：
//   - repo : 已初始化的公告 Repository
//   - cfg  : 可选的配置覆盖，传 nil 使用默认值
//
// 返回：
//   - *GithubAdvisoryService
func NewGithubAdvisoryService(
	repo *repository.GithubAdvisoryRepository,
	cfg *AdvisoryScraperConfig,
) *GithubAdvisoryService {
	c := defaultConfig()
	if cfg != nil {
		if cfg.MaxPages > 0 {
			c.MaxPages = cfg.MaxPages
		}
		if cfg.RequestTimeout > 0 {
			c.RequestTimeout = cfg.RequestTimeout
		}
		if cfg.PerPage > 0 && cfg.PerPage <= 100 {
			c.PerPage = cfg.PerPage
		}
		if cfg.RetryMax > 0 {
			c.RetryMax = cfg.RetryMax
		}
		if cfg.RetryBackoff > 0 {
			c.RetryBackoff = cfg.RetryBackoff
		}
		if cfg.Token != "" {
			c.Token = cfg.Token
		}
	}
	return &GithubAdvisoryService{
		repo:   repo,
		config: c,
		client: &http.Client{Timeout: c.RequestTimeout},
	}
}

// ScrapeAndPersist 通过 GitHub REST API 分页拉取全部安全公告，
// 每页拉取完成后立即批量 Upsert 到数据库，控制内存占用。
//
// 参数：
//   - ctx : 请求上下文，支持超时与取消
//
// 返回：
//   - int   : 本次运行成功持久化的公告总数
//   - error : 请求或写库失败时返回错误
func (s *GithubAdvisoryService) ScrapeAndPersist(ctx context.Context) (int, error) {
	start := time.Now()
	utilities.LogStart(component, "ScrapeAndPersist")

	url := fmt.Sprintf("%s?per_page=%d", ApiBase, s.config.PerPage)
	pageCount := 0
	total := 0

	for url != "" {
		pageCount++

		if s.config.MaxPages > 0 && pageCount > s.config.MaxPages {
			utilities.LogProgress(component, "ScrapeAndPersist",
				fmt.Sprintf("已达到页数上限 (%d)，停止拉取", s.config.MaxPages))
			break
		}

		utilities.LogProgress(component, "ScrapeAndPersist",
			fmt.Sprintf("正在请求第 %d 页: %s", pageCount, url))

		// 带指数退避的请求，自动处理 429 / 5xx。
		body, linkHeader, err := s.fetchWithRetry(ctx, url)
		if err != nil {
			utilities.LogError(component, "ScrapeAndPersist", err, time.Since(start),
				fmt.Sprintf("page=%d", pageCount))
			return total, err
		}

		// 将 JSON 数组解析为原始消息切片，再逐条转换为模型。
		var raw []json.RawMessage
		if err := json.Unmarshal(body, &raw); err != nil {
			return total, fmt.Errorf("第 %d 页 JSON 解析失败: %w", pageCount, err)
		}

		advisories := make([]model.GithubAdvisory, 0, len(raw))
		for _, item := range raw {
			adv, err := parseAPIAdvisory(item)
			if err != nil {
				utilities.Warn("[%s] 跳过无效条目: %v", component, err)
				continue
			}
			advisories = append(advisories, *adv)
		}

		utilities.LogProgress(component, "ScrapeAndPersist",
			fmt.Sprintf("第 %d 页解析出 %d 条公告", pageCount, len(advisories)))

		if len(advisories) > 0 {
			if err := s.repo.BulkUpsert(ctx, advisories); err != nil {
				utilities.LogError(component, "ScrapeAndPersist", err, time.Since(start))
				return total, err
			}
			total += len(advisories)
			utilities.LogProgress(component, "ScrapeAndPersist",
				fmt.Sprintf("已持久化 %d 条公告（累计: %d）", len(advisories), total))
		}

		// 从 Link 响应头获取下一页 URL；为空则说明已到最后一页。
		url = nextLink(linkHeader)
	}

	utilities.LogSuccess(component, "ScrapeAndPersist", time.Since(start),
		fmt.Sprintf("pages=%d", pageCount),
		fmt.Sprintf("total_advisories=%d", total))

	return total, nil
}

// fetchWithRetry 发起 GET 请求，遇到 429 或 5xx 时按指数退避策略自动重试。
// 429 响应会优先读取 Retry-After 响应头决定等待时间。
//
// 参数：
//   - ctx : 请求上下文
//   - url : 目标 URL
//
// 返回：
//   - []byte : 响应体字节切片
//   - string : Link 响应头原始字符串（用于分页）
//   - error  : 重试耗尽或非预期错误时返回
func (s *GithubAdvisoryService) fetchWithRetry(ctx context.Context, url string) ([]byte, string, error) {
	backoff := s.config.RetryBackoff

	for attempt := 0; attempt <= s.config.RetryMax; attempt++ {
		body, linkHeader, statusCode, err := s.doRequest(ctx, url)

		if err == nil {
			return body, linkHeader, nil
		}

		// 非 429 / 5xx 的错误不重试，直接返回。
		if statusCode != 429 && (statusCode < 500 || statusCode >= 600) {
			return nil, "", err
		}

		if attempt == s.config.RetryMax {
			return nil, "", fmt.Errorf("请求 %s 失败，已重试 %d 次: %w", url, s.config.RetryMax, err)
		}

		wait := backoff
		utilities.Warn("[%s] HTTP %d，第 %d/%d 次重试，等待 %v 后继续...",
			component, statusCode, attempt+1, s.config.RetryMax, wait)

		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		case <-time.After(wait):
		}

		backoff *= 2 // 指数退避：2s → 4s → 8s → 16s → 32s
	}

	return nil, "", fmt.Errorf("fetchWithRetry: 不应到达此处")
}

// doRequest 发起单次 HTTP GET 请求，返回响应体、Link 头、状态码和错误。
//
// 参数：
//   - ctx : 请求上下文
//   - url : 目标 URL
//
// 返回：
//   - []byte : 响应体
//   - string : Link 响应头
//   - int    : HTTP 状态码（请求失败时为 0）
//   - error  : 非 2xx 状态码或网络错误时返回
func (s *GithubAdvisoryService) doRequest(ctx context.Context, url string) ([]byte, string, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", 0, fmt.Errorf("构建请求失败: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	// 若配置了 Token，添加 Authorization 头以提升速率限制至 5000 次/小时。
	if s.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.Token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", 0, fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", resp.StatusCode, fmt.Errorf("读取响应体失败: %w", err)
	}

	if resp.StatusCode == 429 {
		// 优先读取 Retry-After 头决定等待时间。
		wait := s.config.RetryBackoff
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil {
				wait = time.Duration(secs) * time.Second
			}
		}
		return nil, "", 429, fmt.Errorf("HTTP 429 Too Many Requests，建议等待 %v", wait)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", resp.StatusCode,
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateBody(body, 200))
	}

	return body, resp.Header.Get("Link"), resp.StatusCode, nil
}

// apiAdvisoryRaw 是 GitHub Advisory API 响应的原始 JSON 结构，
// 仅映射需要的字段，其余字段通过 json.RawMessage 保留原始 JSON。
type apiAdvisoryRaw struct {
	GHSAID          string          `json:"ghsa_id"`
	CVEID           *string         `json:"cve_id"`
	URL             string          `json:"url"`
	HTMLURL         string          `json:"html_url"`
	Summary         string          `json:"summary"`
	Description     string          `json:"description"`
	Type            string          `json:"type"`
	Severity        string          `json:"severity"`
	PublishedAt     *time.Time      `json:"published_at"`
	UpdatedAt       *time.Time      `json:"updated_at"`
	WithdrawnAt     *time.Time      `json:"withdrawn_at"`
	Vulnerabilities json.RawMessage `json:"vulnerabilities"`
	References      json.RawMessage `json:"references"`
}

// parseAPIAdvisory 将单条 API 响应 JSON 转换为 model.GithubAdvisory。
//
// 参数：
//   - raw : 单条公告的原始 JSON 字节
//
// 返回：
//   - *model.GithubAdvisory : 转换成功时返回填充好的公告指针
//   - error                 : JSON 解析失败或 GHSA ID 为空时返回错误
func parseAPIAdvisory(raw json.RawMessage) (*model.GithubAdvisory, error) {
	var a apiAdvisoryRaw
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, fmt.Errorf("JSON 反序列化失败: %w", err)
	}
	if a.GHSAID == "" {
		return nil, fmt.Errorf("ghsa_id 为空，跳过该条目")
	}

	// Vulnerabilities / References 为 null 时替换为空 JSON 数组，避免数据库写入 NULL。
	if len(a.Vulnerabilities) == 0 || string(a.Vulnerabilities) == "null" {
		a.Vulnerabilities = json.RawMessage("[]")
	}
	if len(a.References) == 0 || string(a.References) == "null" {
		a.References = json.RawMessage("[]")
	}

	return &model.GithubAdvisory{
		GHSAID:          a.GHSAID,
		CVEID:           a.CVEID,
		URL:             a.URL,
		HTMLURL:         a.HTMLURL,
		Summary:         a.Summary,
		Description:     a.Description,
		Type:            a.Type,
		Severity:        a.Severity,
		PublishedAt:     a.PublishedAt,
		UpdatedAt:       a.UpdatedAt,
		WithdrawnAt:     a.WithdrawnAt,
		Vulnerabilities: a.Vulnerabilities,
		References:      a.References,
	}, nil
}
