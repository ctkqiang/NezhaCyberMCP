package services

import (
	"context"
	"encoding/json"
	"fmt"
	"nezha_cyber_mcp/internal/model"
	"nezha_cyber_mcp/internal/repository"
	"nezha_cyber_mcp/internal/utilities"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

const (
	// ApiBase 是 GitHub Advisory REST API 的基础地址。
	ApiBase = "https://api.github.com/advisories"

	// UserAgent 是向 GitHub API 发起请求时使用的 User-Agent 标识。
	UserAgent = "advisory-sync/1.0"

	// advisoryBaseURL 是 GitHub 安全公告列表页面的入口地址。
	advisoryBaseURL = "https://github.com/advisories"

	// component 是本服务在日志中使用的组件名称标识。
	component = "GithubAdvisoryService"

	// defaultRequestTimeout 是单次 HTTP 请求的默认超时时间。
	defaultRequestTimeout = 30 * time.Second

	// defaultMaxDepth 是 colly 爬虫允许的最大页面跳转深度。
	defaultMaxDepth = 2

	// defaultParallelism 是爬虫对同一域名的最大并发请求数。
	defaultParallelism = 4

	// defaultRateLimit 是对同一域名相邻两次请求之间的最小间隔，防止触发限流。
	defaultRateLimit = 500 * time.Millisecond
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

// AdvisoryScraperConfig 保存爬虫的可调参数。
// 所有字段均有合理默认值，可通过 NewGithubAdvisoryService 的 cfg 参数覆盖。
type AdvisoryScraperConfig struct {
	// MaxPages 限制最多访问的列表页数，0 表示不限制（抓取全部页面）。
	MaxPages int

	// RequestTimeout 是单次 HTTP 请求的超时时间。
	RequestTimeout time.Duration

	// Parallelism 控制对同一域名的最大并发请求数。
	Parallelism int

	// RateLimit 是对同一域名相邻两次请求之间的最小间隔。
	RateLimit time.Duration
}

// defaultConfig 返回一组合理的默认爬虫配置。
func defaultConfig() AdvisoryScraperConfig {
	return AdvisoryScraperConfig{
		MaxPages:       0,
		RequestTimeout: defaultRequestTimeout,
		Parallelism:    defaultParallelism,
		RateLimit:      defaultRateLimit,
	}
}

// GithubAdvisoryService 负责协调 GitHub 安全公告的抓取与持久化流程。
// 内部使用 colly 异步爬虫抓取列表页，解析公告卡片后批量写入数据库。
type GithubAdvisoryService struct {
	repo   *repository.GithubAdvisoryRepository
	config AdvisoryScraperConfig
}

// NewGithubAdvisoryService 构造 GithubAdvisoryService 实例。
// cfg 为 nil 时使用 defaultConfig() 中的默认值；非 nil 时仅覆盖非零字段。
//
// 参数：
//   - repo : 已初始化的公告 Repository
//   - cfg  : 可选的爬虫配置覆盖，传 nil 使用默认值
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
		if cfg.Parallelism > 0 {
			c.Parallelism = cfg.Parallelism
		}
		if cfg.RateLimit > 0 {
			c.RateLimit = cfg.RateLimit
		}
	}
	return &GithubAdvisoryService{repo: repo, config: c}
}

// ScrapeAndPersist 从 github.com/advisories 抓取安全公告列表页，
// 解析每张公告卡片，并将结果批量 Upsert 到数据库。
// 每页抓取完成后立即写库，控制内存占用。
//
// 参数：
//   - ctx : 请求上下文，支持超时与取消
//
// 返回：
//   - int   : 本次运行成功持久化的公告总数
//   - error : 抓取或写库失败时返回错误；部分成功时同时返回已持久化数量与错误
func (s *GithubAdvisoryService) ScrapeAndPersist(ctx context.Context) (int, error) {
	start := time.Now()
	utilities.LogStart(component, "ScrapeAndPersist")

	collector := s.buildCollector()

	var (
		batch     []model.GithubAdvisory // 当前页解析出的公告缓冲
		pageCount int                    // 已访问的页面计数
		total     int                    // 累计持久化的公告数
		scrapeErr error                  // 记录最后一次错误，供 Wait() 后检查
	)

	// 匹配列表页上的每张公告卡片，解析后追加到当前页缓冲。
	collector.OnHTML("div[data-testid='advisory-list-item'], .js-advisory-list-item, article.Box-row", func(e *colly.HTMLElement) {
		advisory, err := parseAdvisoryCard(e)
		if err != nil {
			utilities.Warn("[%s] 跳过无效卡片: %v", component, err)
			return
		}
		batch = append(batch, *advisory)
	})

	// 跟随分页链接，访问下一页；若已达到 MaxPages 上限则停止。
	collector.OnHTML("a[rel='next'], .next_page", func(e *colly.HTMLElement) {
		if s.config.MaxPages > 0 && pageCount >= s.config.MaxPages {
			utilities.LogProgress(component, "ScrapeAndPersist",
				fmt.Sprintf("已达到页数上限 (%d)，停止翻页", s.config.MaxPages))
			return
		}
		next := e.Attr("href")
		if next == "" {
			return
		}
		if !strings.HasPrefix(next, "http") {
			next = "https://github.com" + next
		}
		_ = e.Request.Visit(next)
	})

	// 每次发起请求前记录日志并保存请求开始时间。
	collector.OnRequest(func(r *colly.Request) {
		pageCount++
		utilities.LogProgress(component, "ScrapeAndPersist",
			fmt.Sprintf("正在访问第 %d 页: %s", pageCount, r.URL.String()))
		r.Ctx.Put("start", time.Now())
	})

	// 收到响应后记录状态码与响应体大小。
	collector.OnResponse(func(r *colly.Response) {
		utilities.LogProgress(component, "ScrapeAndPersist",
			fmt.Sprintf("第 %d 页响应: HTTP %d (%d 字节)",
				pageCount, r.StatusCode, len(r.Body)))
	})

	// HTTP 请求失败时记录错误，后续在 Wait() 返回后统一处理。
	collector.OnError(func(r *colly.Response, err error) {
		scrapeErr = fmt.Errorf("HTTP %d on %s: %w", r.StatusCode, r.Request.URL, err)
		utilities.LogError(component, "ScrapeAndPersist", scrapeErr, time.Since(start))
	})

	// 每页抓取完成后将当前缓冲批量写入数据库，然后清空缓冲。
	collector.OnScraped(func(r *colly.Response) {
		if len(batch) == 0 {
			return
		}
		if err := s.repo.BulkUpsert(ctx, batch); err != nil {
			scrapeErr = err
			utilities.LogError(component, "ScrapeAndPersist", err, time.Since(start))
			return
		}
		total += len(batch)
		utilities.LogProgress(component, "ScrapeAndPersist",
			fmt.Sprintf("已持久化 %d 条公告（累计: %d）", len(batch), total))
		batch = batch[:0] // 清空缓冲，复用底层数组
	})

	// 访问入口页面，触发整个抓取流程。
	if err := collector.Visit(advisoryBaseURL); err != nil {
		utilities.LogError(component, "ScrapeAndPersist", err, time.Since(start))
		return 0, fmt.Errorf("initial visit failed: %w", err)
	}

	// 等待所有异步请求完成。
	collector.Wait()

	if scrapeErr != nil {
		utilities.LogError(component, "ScrapeAndPersist", scrapeErr, time.Since(start))
		return total, scrapeErr
	}

	utilities.LogSuccess(component, "ScrapeAndPersist", time.Since(start),
		fmt.Sprintf("pages=%d", pageCount),
		fmt.Sprintf("total_advisories=%d", total))

	return total, nil
}

// buildCollector 创建并配置 colly.Collector 实例。
// 限制只允许访问 github.com 域名，启用异步模式，并应用速率限制规则。
func (s *GithubAdvisoryService) buildCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.AllowedDomains("github.com"),
		colly.MaxDepth(defaultMaxDepth),
		colly.Async(true),
	)

	// 随机轮换 User-Agent，降低被识别为爬虫的概率。
	extensions.RandomUserAgent(c)

	// 对 github.com 域名应用并发与速率限制规则。
	_ = c.Limit(&colly.LimitRule{
		DomainGlob:  "*github.com*",
		Parallelism: s.config.Parallelism,
		Delay:       s.config.RateLimit,
		RandomDelay: s.config.RateLimit / 2, // 在固定间隔基础上叠加随机抖动
	})

	c.SetRequestTimeout(s.config.RequestTimeout)

	return c
}

// parseAdvisoryCard 从列表页的单张公告卡片 HTML 元素中提取 GithubAdvisory 数据。
//
// GitHub 公告列表页的卡片结构（简化示意）：
//
//	<article class="Box-row">
//	  <a href="/advisories/GHSA-xxxx-xxxx-xxxx">GHSA-xxxx-xxxx-xxxx</a>
//	  <span class="Label">critical</span>
//	  <p>漏洞摘要文本</p>
//	  <relative-time datetime="2024-01-01T00:00:00Z">...</relative-time>
//	</article>
//
// 参数：
//   - e : colly 解析出的 HTML 元素，对应一张公告卡片
//
// 返回：
//   - *model.GithubAdvisory : 解析成功时返回填充好的公告指针
//   - error                 : 无法提取 GHSA ID 时返回错误
func parseAdvisoryCard(e *colly.HTMLElement) (*model.GithubAdvisory, error) {
	// 从卡片链接中提取 GHSA ID 和详情页 URL。
	ghsaID, detailURL := extractGHSALink(e)
	if ghsaID == "" {
		return nil, fmt.Errorf("无法从卡片 HTML 中提取 GHSA ID")
	}

	// 提取严重程度标签并规范化为标准枚举值。
	severity := normaliseSeverity(
		e.ChildText(".Label--attention, .Label--danger, .Label--warning, .Label--success, [data-severity]"),
	)

	// 提取漏洞摘要，优先取 <p> 标签，回退到标题标签。
	summary := strings.TrimSpace(
		e.ChildText("p, .advisory-summary, [data-testid='advisory-summary']"),
	)
	if summary == "" {
		summary = strings.TrimSpace(e.ChildText("h3, h4"))
	}

	// 解析 <relative-time> 或 <time-ago> 元素的 datetime 属性为 *time.Time。
	publishedAt := parseRelativeTime(e.ChildAttr("relative-time, time-ago", "datetime"))

	advisory := &model.GithubAdvisory{
		GHSAID:      ghsaID,
		URL:         "https://api.github.com/advisories/" + ghsaID,
		HTMLURL:     detailURL,
		Summary:     summary,
		Severity:    severity,
		Type:        "reviewed",
		PublishedAt: publishedAt,
		// Vulnerabilities 和 References 由详情页富化步骤填充，此处初始化为空数组。
		Vulnerabilities: json.RawMessage("[]"),
		References:      json.RawMessage("[]"),
	}

	return advisory, nil
}

// extractGHSALink 遍历元素内所有 <a href> 链接，
// 找到第一个包含 /advisories/GHSA- 路径的链接，返回 GHSA ID 和绝对 HTML URL。
// 若未找到匹配链接，两个返回值均为空字符串。
//
// 参数：
//   - e : colly HTML 元素
//
// 返回：
//   - ghsaID  : GHSA 标识符（如 GHSA-xxxx-xxxx-xxxx），未找到时为空字符串
//   - htmlURL : 公告详情页的绝对 URL，未找到时为空字符串
func extractGHSALink(e *colly.HTMLElement) (ghsaID, htmlURL string) {
	e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
		if ghsaID != "" {
			return // 已找到，跳过后续链接
		}
		href := el.Attr("href")
		if strings.Contains(href, "/advisories/GHSA-") {
			parts := strings.Split(href, "/")
			for _, p := range parts {
				if strings.HasPrefix(p, "GHSA-") {
					ghsaID = p
					if strings.HasPrefix(href, "http") {
						htmlURL = href
					} else {
						// 相对路径补全为绝对 URL
						htmlURL = "https://github.com" + href
					}
					return
				}
			}
		}
	})
	return
}

// normaliseSeverity 将页面上的原始严重程度标签文本映射为
// GitHub Advisory API 规范的枚举值：low | medium | high | critical。
// 无法识别的输入统一返回 "unknown"。
//
// 参数：
//   - raw : 从 HTML 标签中提取的原始文本（大小写不敏感）
//
// 返回：
//   - string : 规范化后的严重程度字符串
func normaliseSeverity(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "moderate", "medium":
		return "medium"
	case "low":
		return "low"
	default:
		return "unknown"
	}
}

// parseRelativeTime 将 RFC 3339 格式的日期时间字符串解析为 *time.Time。
// 解析失败或输入为空时返回 nil，调用方将其视为可选字段处理。
//
// 参数：
//   - raw : RFC 3339 格式的日期时间字符串（如 "2024-01-01T00:00:00Z"）
//
// 返回：
//   - *time.Time : 解析成功时返回时间指针，失败或空输入时返回 nil
func parseRelativeTime(raw string) *time.Time {
	if raw == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return &t
}
