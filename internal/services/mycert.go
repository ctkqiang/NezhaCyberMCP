package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"nezha_cyber_mcp/internal/model"
	"nezha_cyber_mcp/internal/repository"
	"nezha_cyber_mcp/internal/utilities"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	// mycertComponent 是本服务在日志中使用的组件名称标识。
	mycertComponent = "MycertAdvisoryService"

	// mycertBaseURL 是 MyCERT 公告列表页的基础 URL，包含分类 ID 参数。
	mycertBaseURL = "https://www.mycert.org.my/portal/advisories?id=431fab9c-d24c-4a27-ba93-e92edafdefa5"

	// mycertDetailBase 是公告详情页的基础 URL 前缀。
	mycertDetailBase = "https://www.mycert.org.my/portal/advisory?id="

	// mycertOrigin 是 MyCERT 网站的根域名，用于拼接相对路径。
	mycertOrigin = "https://www.mycert.org.my"

	// mycertUserAgent 是向 MyCERT 发起请求时使用的 User-Agent 标识。
	mycertUserAgent = "MyCERT-Advisory-Sync/1.0 (research; contact: research@example.com)"

	// mycertPerPage 是每页显示的公告条数，与网站默认值一致。
	mycertPerPage = 10

	// mycertRequestDelay 是相邻两次 HTTP 请求之间的礼貌等待时间，避免对服务器造成压力。
	mycertRequestDelay = 1500 * time.Millisecond
)

// MycertScraperConfig 保存 MyCERT 爬虫的可调参数。
type MycertScraperConfig struct {
	// MaxPages 限制最多抓取的列表页数，0 表示不限制（抓取全部页面）。
	MaxPages int

	// RequestTimeout 是单次 HTTP 请求的超时时间。
	RequestTimeout time.Duration

	// FetchDetail 控制是否抓取每条公告的详情页全文。
	// 设为 false 时仅保存列表页摘要，速度更快但内容不完整。
	FetchDetail bool
}

// defaultMycertConfig 返回一组合理的默认爬虫配置。
func defaultMycertConfig() MycertScraperConfig {
	return MycertScraperConfig{
		MaxPages:       0,
		RequestTimeout: 30 * time.Second,
		FetchDetail:    true,
	}
}

// MycertAdvisoryService 负责从 MyCERT 门户抓取安全公告并持久化到数据库。
type MycertAdvisoryService struct {
	repo    *repository.MycertAdvisoryRepository
	config  MycertScraperConfig
	client  *http.Client
	baseURL string
}

// NewMycertAdvisoryService 构造 MycertAdvisoryService 实例。
// cfg 为 nil 时使用 defaultMycertConfig() 中的默认值；非 nil 时仅覆盖非零字段。
//
// 参数：
//   - repo : 已初始化的 MyCERT 公告 Repository
//   - cfg  : 可选的配置覆盖，传 nil 使用默认值
//
// 返回：
//   - *MycertAdvisoryService
func NewMycertAdvisoryService(
	repo *repository.MycertAdvisoryRepository,
	cfg *MycertScraperConfig,
) *MycertAdvisoryService {
	c := defaultMycertConfig()
	if cfg != nil {
		if cfg.MaxPages > 0 {
			c.MaxPages = cfg.MaxPages
		}
		if cfg.RequestTimeout > 0 {
			c.RequestTimeout = cfg.RequestTimeout
		}
		c.FetchDetail = cfg.FetchDetail
	}

	return &MycertAdvisoryService{
		repo:    repo,
		config:  c,
		client:  &http.Client{Timeout: c.RequestTimeout},
		baseURL: mycertBaseURL,
	}
}

// NewMycertAdvisoryServiceWithBaseURL 构造 MycertAdvisoryService 实例，
// 并允许调用方指定自定义的列表页基础 URL。
// 该构造器主要用于单元测试中注入 httptest.Server 地址，生产代码应使用 NewMycertAdvisoryService。
//
// 参数：
//   - repo    : 已初始化的 MyCERT 公告 Repository
//   - cfg     : 可选的配置覆盖，传 nil 使用默认值
//   - baseURL : 列表页基础 URL（不含 page/per-page 参数）
//
// 返回：
//   - *MycertAdvisoryService
func NewMycertAdvisoryServiceWithBaseURL(
	repo *repository.MycertAdvisoryRepository,
	cfg *MycertScraperConfig,
	baseURL string,
) *MycertAdvisoryService {
	svc := NewMycertAdvisoryService(repo, cfg)
	svc.baseURL = baseURL
	return svc
}

// ScrapeAndPersist 分页抓取 MyCERT 公告列表，可选地抓取每条详情页全文，
// 每页完成后立即批量 Upsert 到数据库。
//
// 参数：
//   - ctx : 请求上下文，支持超时与取消
//
// 返回：
//   - int   : 本次运行成功持久化的公告总数
//   - error : 请求或写库失败时返回错误
func (s *MycertAdvisoryService) ScrapeAndPersist(ctx context.Context) (int, error) {
	start := time.Now()
	utilities.LogStart(mycertComponent, "ScrapeAndPersist")

	page := 1
	total := 0

	for {
		if s.config.MaxPages > 0 && page > s.config.MaxPages {
			utilities.LogProgress(mycertComponent, "ScrapeAndPersist",
				fmt.Sprintf("已达到页数上限 (%d)，停止抓取", s.config.MaxPages))
			break
		}

		pageURL := fmt.Sprintf("%s&page=%d&per-page=%d", s.baseURL, page, mycertPerPage)
		utilities.LogProgress(mycertComponent, "ScrapeAndPersist",
			fmt.Sprintf("正在抓取第 %d 页: %s", page, pageURL))

		items, hasNext, err := s.scrapePage(ctx, pageURL)
		if err != nil {
			utilities.LogError(mycertComponent, "ScrapeAndPersist", err, time.Since(start),
				fmt.Sprintf("page=%d", page))
			return total, err
		}

		if len(items) == 0 {
			utilities.LogProgress(mycertComponent, "ScrapeAndPersist",
				fmt.Sprintf("第 %d 页无数据，停止抓取", page))
			break
		}

		if s.config.FetchDetail {
			for i := range items {
				if err := ctx.Err(); err != nil {
					return total, err
				}
				if err := s.enrichDetail(ctx, &items[i]); err != nil {
					utilities.Warn("[%s] 详情页抓取失败，跳过 %s: %v",
						mycertComponent, items[i].AdvisoryID, err)
				}
				time.Sleep(mycertRequestDelay)
			}
		}

		if err := s.repo.BulkUpsert(ctx, items); err != nil {
			utilities.LogError(mycertComponent, "ScrapeAndPersist", err, time.Since(start))
			return total, err
		}

		total += len(items)
		utilities.LogProgress(mycertComponent, "ScrapeAndPersist",
			fmt.Sprintf("已持久化 %d 条公告（累计: %d）", len(items), total))

		if !hasNext {
			break
		}

		page++
		time.Sleep(mycertRequestDelay)
	}

	utilities.LogSuccess(mycertComponent, "ScrapeAndPersist", time.Since(start),
		fmt.Sprintf("pages=%d", page),
		fmt.Sprintf("total_advisories=%d", total))

	return total, nil
}

// scrapePage 抓取单个列表页，解析出公告摘要列表和是否存在下一页。
//
// 参数：
//   - ctx     : 请求上下文
//   - pageURL : 列表页完整 URL
//
// 返回：
//   - []model.MycertAdvisory : 本页解析出的公告列表
//   - bool                   : 是否存在下一页
//   - error                  : 请求或解析失败时返回错误
func (s *MycertAdvisoryService) scrapePage(ctx context.Context, pageURL string) ([]model.MycertAdvisory, bool, error) {
	var (
		advisories []model.MycertAdvisory
		walk       func(*html.Node)
		body, err  = s.fetch(ctx, pageURL)
	)

	if err != nil {
		return nil, false, err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, false, fmt.Errorf("HTML 解析失败: %w", err)
	}

	hasNext := false

	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "h4":
				// 公告标题节点：<h4> 内含 <a> 链接，链接文本即标题，href 指向详情页。
				if adv := parseAdvisoryCard(n); adv != nil {
					advisories = append(advisories, *adv)
				}
			case "a":
				// 检测分页中的 "Next" 链接，判断是否存在下一页。
				if hasAttrContains(n, "href", "page=") && strings.Contains(textContent(n), "Next") {
					hasNext = true
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return advisories, hasNext, nil
}

// parseAdvisoryCard 从 <h4> 节点解析出单条公告的基础信息。
// MyCERT 列表页真实结构（通过 curl 验证）：
//
//	<h4>MA-XXXX.XXXXXX: MyCERT Advisory - ...</h4>
//	<ul class="meta">
//	  <li><small><i ...></i> 25 Jun 2026</small></li>
//	  <li><small><i ...></i> Advisory</small></li>
//	</ul>
//	<div class="desc-news"><p>摘要文本...</p></div>
//
// 注意：<h4> 内无 <a> 子节点，标题文本直接在 <h4> 里；
// 详情页 URL 通过 AdvisoryID 拼接 mycertDetailBase 得到。
//
// 参数：
//   - h4 : <h4> DOM 节点
//
// 返回：
//   - *model.MycertAdvisory : 解析成功时返回公告指针，无法提取 ID 时返回 nil
func parseAdvisoryCard(h4 *html.Node) *model.MycertAdvisory {
	title := strings.TrimSpace(textContent(h4))

	if title == "" {
		return nil
	}

	advisoryID := ""
	if idx := strings.Index(title, ":"); idx > 0 {
		advisoryID = strings.TrimSpace(title[:idx])
	}

	if advisoryID == "" {
		return nil
	}

	detailURL := mycertDetailBase + advisoryID

	var (
		publishedAt *time.Time
		category    string
		summary     string
	)

	sibling := h4.NextSibling

	for sibling != nil {
		if sibling.Type == html.ElementNode {
			switch sibling.Data {
			case "ul":
				liIdx := 0
				for li := sibling.FirstChild; li != nil; li = li.NextSibling {
					if li.Type != html.ElementNode || li.Data != "li" {
						continue
					}
					// 日期和分类文本嵌套在 <small> 里，用 textContent 提取纯文本后去除图标空白。
					text := strings.TrimSpace(textContent(li))
					switch liIdx {
					case 0:
						if t, err := time.Parse("02 Jan 2006", text); err == nil {
							publishedAt = &t
						}
					case 1:
						category = text
					}
					liIdx++
				}
			case "div":
				// 摘要在 <div class="desc-news"> 里，提取其纯文本作为摘要。
				if strings.Contains(attrVal(sibling, "class"), "desc-news") {
					summary = strings.TrimSpace(textContent(sibling))
				}
			case "h4":
				sibling = nil
				continue
			}
		}
		if sibling != nil {
			sibling = sibling.NextSibling
		}
	}

	return &model.MycertAdvisory{
		AdvisoryID:  advisoryID,
		Title:       title,
		PublishedAt: publishedAt,
		Category:    category,
		Summary:     summary,
		DetailURL:   detailURL,
		ScrapedAt:   time.Now(),
	}
}

// enrichDetail 抓取单条公告的详情页，将正文全文填充到 FullContent 字段。
//
// 参数：
//   - ctx : 请求上下文
//   - adv : 待填充的公告指针，DetailURL 字段必须已设置
//
// 返回：
//   - error : 请求或解析失败时返回错误
func (s *MycertAdvisoryService) enrichDetail(ctx context.Context, adv *model.MycertAdvisory) error {
	if adv.DetailURL == "" {
		return nil
	}

	body, err := s.fetch(ctx, adv.DetailURL)
	if err != nil {
		return err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("详情页 HTML 解析失败: %w", err)
	}

	// 提取详情页正文：查找 class 包含 "content" 或 "article" 的 <div>，
	// 回退到提取 <main> 标签，最终回退到 <body> 全文。
	content := extractMainContent(doc)
	adv.FullContent = strings.TrimSpace(content)
	adv.ScrapedAt = time.Now()

	return nil
}

// extractMainContent 从 HTML 文档中提取主要正文内容，返回纯文本。
// 优先级：class 含 "content" 的 <div> > <main> > <article> > <body>
//
// 参数：
//   - doc : 已解析的 HTML 文档根节点
//
// 返回：
//   - string : 提取到的纯文本内容
func extractMainContent(doc *html.Node) string {
	if n := findNodeByClassContains(doc, "div", "content"); n != nil {
		return textContent(n)
	}

	if n := findNodeByTag(doc, "main"); n != nil {
		return textContent(n)
	}

	if n := findNodeByTag(doc, "article"); n != nil {
		return textContent(n)
	}

	if n := findNodeByTag(doc, "body"); n != nil {
		return textContent(n)
	}

	return textContent(doc)
}

// fetch 发起单次 HTTP GET 请求，返回响应体字节切片。
// 自动设置合规的 User-Agent 和 Accept 头。
//
// 参数：
//   - ctx : 请求上下文
//   - url : 目标 URL
//
// 返回：
//   - []byte : 响应体
//   - error  : 非 2xx 状态码或网络错误时返回
func (s *MycertAdvisoryService) fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", mycertUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP 请求失败 [%s]: %w", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	return body, nil
}

// textContent 递归提取 HTML 节点及其所有子节点的纯文本内容，
// 在块级元素（div、p、li、h1-h6 等）之间插入换行符以保留可读性。
//
// 参数：
//   - n : HTML 节点
//
// 返回：
//   - string : 提取到的纯文本
func textContent(n *html.Node) string {
	var (
		sb      strings.Builder
		extract func(*html.Node)
	)

	extract = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
			return
		}

		if node.Type == html.ElementNode {
			switch node.Data {
			case "script", "style", "noscript":
				return
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}

		if node.Type == html.ElementNode {
			switch node.Data {
			case "p", "div", "li", "h1", "h2", "h3", "h4", "h5", "h6", "br", "tr":
				sb.WriteString("\n")
			}
		}
	}
	extract(n)
	return sb.String()
}

// attrVal 返回 HTML 节点指定属性的值，属性不存在时返回空字符串。
//
// 参数：
//   - n    : HTML 节点
//   - name : 属性名称
//
// 返回：
//   - string : 属性值或空字符串
func attrVal(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}

	return ""
}

// hasAttrContains 检查 HTML 节点的指定属性值是否包含给定子字符串。
//
// 参数：
//   - n      : HTML 节点
//   - attr   : 属性名称
//   - substr : 待检查的子字符串
//
// 返回：
//   - bool : 属性值包含子字符串时返回 true
func hasAttrContains(n *html.Node, attr, substr string) bool {
	return strings.Contains(attrVal(n, attr), substr)
}

// findNodeByTag 在 DOM 树中深度优先查找第一个匹配指定标签名的节点。
//
// 参数：
//   - n   : 搜索起始节点
//   - tag : HTML 标签名（小写）
//
// 返回：
//   - *html.Node : 找到的节点；未找到时返回 nil
func findNodeByTag(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findNodeByTag(c, tag); found != nil {
			return found
		}
	}

	return nil
}

// findNodeByClassContains 在 DOM 树中深度优先查找第一个 class 属性包含指定子字符串的节点。
//
// 参数：
//   - n       : 搜索起始节点
//   - tag     : HTML 标签名（小写），空字符串表示匹配任意标签
//   - partial : class 属性中需要包含的子字符串
//
// 返回：
//   - *html.Node : 找到的节点；未找到时返回 nil
func findNodeByClassContains(n *html.Node, tag, partial string) *html.Node {
	if n.Type == html.ElementNode {
		if (tag == "" || n.Data == tag) && strings.Contains(attrVal(n, "class"), partial) {
			return n
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findNodeByClassContains(c, tag, partial); found != nil {
			return found
		}
	}
	return nil
}
