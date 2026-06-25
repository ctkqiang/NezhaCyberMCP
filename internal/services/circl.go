package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"nezha_cyber_mcp/internal/model"
	"nezha_cyber_mcp/internal/repository"
	"nezha_cyber_mcp/internal/utilities"
)

const (
	// circlComponent 是本服务在日志中使用的组件名称标识。
	circlComponent = "CirclCVEService"

	// circlPrimaryBase 是 CIRCL Vulnerability-Lookup API 的主端点基础地址。
	circlPrimaryBase = "https://cve.circl.lu/api"

	// circlFallbackBase 是 CIRCL API 的备用端点基础地址，主端点不可用时自动切换。
	circlFallbackBase = "https://vulnerability.circl.lu/api"

	// circlUserAgent 是向 CIRCL API 发起请求时使用的 User-Agent 标识。
	circlUserAgent = "NezhaCyberMCP-CVE-Sync/1.0 (research; contact: research@example.com)"

	// circlDefaultTimeout 是单次 HTTP 请求的默认超时时间。
	circlDefaultTimeout = 30 * time.Second

	// circlDefaultPerPage 是每页返回的最大条目数（CIRCL API 上限为 1000）。
	circlDefaultPerPage = 100

	// circlDefaultRetryMax 是遇到 429 / 5xx 时的最大重试次数。
	circlDefaultRetryMax = 5

	// circlDefaultRetryBackoff 是首次重试前的等待基准时间，每次重试翻倍（指数退避）。
	circlDefaultRetryBackoff = 2 * time.Second

	// circlDefaultRateLimit 是相邻两次 API 请求之间的最小间隔，避免触发速率限制。
	// CIRCL 公共 API 无明确文档限制，保守设置为 500ms。
	circlDefaultRateLimit = 500 * time.Millisecond
)

// ---- 解析用内部结构体（仅在本包内使用，不导出）----

// circlAPIResponse 是 CIRCL API /vulnerability/{id} 端点返回的原始 JSON 结构。
type circlAPIResponse struct {
	DataType    string           `json:"dataType"`
	DataVersion string           `json:"dataVersion"`
	CVEMetadata circlCVEMetadata `json:"cveMetadata"`
	Containers  circlContainers  `json:"containers"`
}

// circlCVEMetadata 对应 API 响应中的 cveMetadata 字段。
type circlCVEMetadata struct {
	State         string `json:"state"`
	CVEID         string `json:"cveId"`
	AssignerOrgID string `json:"assignerOrgId"`
	AssignerShort string `json:"assignerShortName"`
	DateUpdated   string `json:"dateUpdated"`
	DateReserved  string `json:"dateReserved"`
	DatePublished string `json:"datePublished"`
}

// circlContainers 对应 API 响应中的 containers 字段。
type circlContainers struct {
	CNA circlCNAContainer `json:"cna"`
}

// circlCNAContainer 对应 containers.cna 字段，包含漏洞的核心描述信息。
type circlCNAContainer struct {
	Title        string             `json:"title"`
	Descriptions []circlDescription `json:"descriptions"`
	Affected     json.RawMessage    `json:"affected"`
	References   json.RawMessage    `json:"references"`
	Metrics      []circlMetric      `json:"metrics"`
	ProblemTypes []circlProblemType `json:"problemTypes"`
}

// circlDescription 对应单条描述条目。
type circlDescription struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

// circlMetric 对应 metrics 数组中的单个元素，用于提取严重程度。
type circlMetric struct {
	Other *circlMetricOther `json:"other"`
}

// circlMetricOther 对应 metrics[].other 字段。
type circlMetricOther struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

// circlProblemType 对应 problemTypes 数组中的单个元素。
type circlProblemType struct {
	Descriptions []circlCWEDescription `json:"descriptions"`
}

// circlCWEDescription 对应 problemTypes[].descriptions 中的单个 CWE 描述。
type circlCWEDescription struct {
	Type        string `json:"type"`
	Lang        string `json:"lang"`
	Description string `json:"description"`
	CWEID       string `json:"cweId"`
}

// ---- 配置与服务结构体 ----

// CirclScraperConfig 保存 CIRCL CVE API 客户端的可调参数。
// 所有字段均有合理默认值，可通过 NewCirclCVEService 的 cfg 参数覆盖。
type CirclScraperConfig struct {
	// MaxPages 限制最多请求的页数，0 表示不限制（拉取全部数据）。
	MaxPages int

	// RequestTimeout 是单次 HTTP 请求的超时时间。
	RequestTimeout time.Duration

	// PerPage 是每页返回的条目数，最大 1000。
	PerPage int

	// RetryMax 是遇到 429 / 5xx 时的最大重试次数。
	RetryMax int

	// RetryBackoff 是首次重试前的等待基准时间，每次重试翻倍。
	RetryBackoff time.Duration

	// RateLimit 是相邻两次请求之间的最小等待时间，用于主动限速。
	RateLimit time.Duration

	// APIToken 是 CIRCL API 的认证令牌（Bearer Token）。
	// CIRCL 公共 API 目前不强制要求认证，但提供 Token 可提升速率限制。
	// 从环境变量 CIRCL_API_TOKEN 读取，不得硬编码。
	APIToken string
}

// defaultCirclConfig 返回一组合理的默认 CIRCL API 客户端配置。
func defaultCirclConfig() CirclScraperConfig {
	return CirclScraperConfig{
		MaxPages:       0,
		RequestTimeout: circlDefaultTimeout,
		PerPage:        circlDefaultPerPage,
		RetryMax:       circlDefaultRetryMax,
		RetryBackoff:   circlDefaultRetryBackoff,
		RateLimit:      circlDefaultRateLimit,
		APIToken:       os.Getenv("CIRCL_API_TOKEN"),
	}
}

// CirclCVEService 负责通过 CIRCL Vulnerability-Lookup API 拉取 CVE 数据并持久化到数据库。
// 支持主端点与备用端点自动切换，内置速率限制与指数退避重试。
type CirclCVEService struct {
	repo    *repository.CirclCVERepository
	config  CirclScraperConfig
	client  *http.Client
	baseURL string
}

// NewCirclCVEService 构造 CirclCVEService 实例。
// cfg 为 nil 时使用 defaultCirclConfig() 中的默认值；非 nil 时仅覆盖非零字段。
//
// 参数：
//   - repo : 已初始化的 CIRCL CVE Repository
//   - cfg  : 可选的配置覆盖，传 nil 使用默认值
//
// 返回：
//   - *CirclCVEService
func NewCirclCVEService(
	repo *repository.CirclCVERepository,
	cfg *CirclScraperConfig,
) *CirclCVEService {
	c := defaultCirclConfig()
	if cfg != nil {
		if cfg.MaxPages > 0 {
			c.MaxPages = cfg.MaxPages
		}
		if cfg.RequestTimeout > 0 {
			c.RequestTimeout = cfg.RequestTimeout
		}
		if cfg.PerPage > 0 && cfg.PerPage <= 1000 {
			c.PerPage = cfg.PerPage
		}
		if cfg.RetryMax > 0 {
			c.RetryMax = cfg.RetryMax
		}
		if cfg.RetryBackoff > 0 {
			c.RetryBackoff = cfg.RetryBackoff
		}
		if cfg.RateLimit > 0 {
			c.RateLimit = cfg.RateLimit
		}
		if cfg.APIToken != "" {
			c.APIToken = cfg.APIToken
		}
	}

	return &CirclCVEService{
		repo:    repo,
		config:  c,
		client:  &http.Client{Timeout: c.RequestTimeout},
		baseURL: circlPrimaryBase,
	}
}

// NewCirclCVEServiceWithBaseURL 构造 CirclCVEService 实例，并允许覆盖 API 基础地址。
// 该函数主要用于单元测试，通过传入 httptest.Server 的 URL 来模拟 API 响应。
//
// 参数：
//   - repo    : 已初始化的 CIRCL CVE Repository
//   - cfg     : 可选的配置覆盖，传 nil 使用默认值
//   - baseURL : 覆盖默认的 API 基础地址（如 httptest.Server.URL）
//
// 返回：
//   - *CirclCVEService
func NewCirclCVEServiceWithBaseURL(
	repo *repository.CirclCVERepository,
	cfg *CirclScraperConfig,
	baseURL string,
) *CirclCVEService {
	svc := NewCirclCVEService(repo, cfg)
	svc.baseURL = baseURL
	return svc
}

// ParseCirclResponseForTest 将 CIRCL API 响应 JSON 解析为 CirclCVE 模型。
// 该函数仅供测试包调用，用于直接验证解析与归一化逻辑，不执行数据库操作。
//
// 参数：
//   - data : CIRCL API 响应的原始 JSON 字节
//
// 返回：
//   - *model.CirclCVE : 归一化后的 CVE 记录
//   - error           : JSON 解析失败或必要字段缺失时返回错误
func ParseCirclResponseForTest(data []byte) (*model.CirclCVE, error) {
	return parseCirclResponse(data)
}

// ScrapeAndPersist 通过 CIRCL API 分页拉取全部 CVE 漏洞记录，
// 每页拉取完成后立即批量 Upsert 到数据库，控制内存占用。
// 若主端点连续失败，自动切换到备用端点重试。
//
// 参数：
//   - ctx : 请求上下文，支持超时与取消
//
// 返回：
//   - int   : 本次运行成功持久化的 CVE 总数
//   - error : 请求或写库失败时返回错误
func (s *CirclCVEService) ScrapeAndPersist(ctx context.Context) (int, error) {
	start := time.Now()
	utilities.LogStart(circlComponent, "ScrapeAndPersist")

	page := 1
	total := 0

	for {
		if s.config.MaxPages > 0 && page > s.config.MaxPages {
			utilities.LogProgress(circlComponent, "ScrapeAndPersist",
				fmt.Sprintf("已达到页数上限 (%d)，停止拉取", s.config.MaxPages))
			break
		}

		listURL := fmt.Sprintf("%s/vulnerability/?page=%d&per_page=%d",
			s.baseURL, page, s.config.PerPage)

		utilities.LogProgress(circlComponent, "ScrapeAndPersist",
			fmt.Sprintf("正在请求第 %d 页: %s", page, listURL))

		// 主动限速：相邻两次请求之间等待，避免触发服务端速率限制。
		if page > 1 {
			select {
			case <-ctx.Done():
				return total, ctx.Err()
			case <-time.After(s.config.RateLimit):
			}
		}

		body, err := s.fetchWithRetry(ctx, listURL)
		if err != nil {
			utilities.LogError(circlComponent, "ScrapeAndPersist", err, time.Since(start),
				fmt.Sprintf("page=%d", page))
			return total, err
		}

		var listResp model.CirclVulnerabilityListResponse
		if err := json.Unmarshal(body, &listResp); err != nil {
			// 部分端点直接返回 ID 字符串数组，尝试兼容解析。
			var idList []string
			if jsonErr := json.Unmarshal(body, &idList); jsonErr != nil {
				return total, fmt.Errorf("第 %d 页列表 JSON 解析失败: %w", page, err)
			}
			listResp.Data = make([]model.CirclVulnerabilityListItem, len(idList))
			for i, id := range idList {
				listResp.Data[i] = model.CirclVulnerabilityListItem{VulnID: id}
			}
		}

		if len(listResp.Data) == 0 {
			utilities.LogProgress(circlComponent, "ScrapeAndPersist",
				fmt.Sprintf("第 %d 页返回空数据，已到达末页", page))
			break
		}

		utilities.LogProgress(circlComponent, "ScrapeAndPersist",
			fmt.Sprintf("第 %d 页获取到 %d 条 CVE ID，开始逐条拉取详情", page, len(listResp.Data)))

		cves, err := s.fetchDetailBatch(ctx, listResp.Data)
		if err != nil {
			utilities.LogError(circlComponent, "ScrapeAndPersist", err, time.Since(start),
				fmt.Sprintf("page=%d", page))
			return total, err
		}

		if len(cves) > 0 {
			if err := s.repo.BulkUpsert(ctx, cves); err != nil {
				utilities.LogError(circlComponent, "ScrapeAndPersist", err, time.Since(start))
				return total, err
			}
			total += len(cves)
			utilities.LogProgress(circlComponent, "ScrapeAndPersist",
				fmt.Sprintf("第 %d 页已持久化 %d 条 CVE（累计: %d）", page, len(cves), total))
		}

		// 判断是否已到最后一页：返回条数小于每页上限，或 total 字段已全部拉取完毕。
		if len(listResp.Data) < s.config.PerPage {
			break
		}
		if listResp.Total > 0 && total >= listResp.Total {
			break
		}

		page++
	}

	utilities.LogSuccess(circlComponent, "ScrapeAndPersist", time.Since(start),
		fmt.Sprintf("pages=%d", page),
		fmt.Sprintf("total_cves=%d", total))

	return total, nil
}

// FetchByCVEID 按 CVE ID 拉取单条漏洞详情并持久化到数据库。
// 适用于按需查询特定 CVE 的场景。
//
// 参数：
//   - ctx   : 请求上下文
//   - cveID : CVE 编号，如 CVE-2021-44228
//
// 返回：
//   - *model.CirclCVE : 解析并归一化后的 CVE 记录
//   - error           : 请求或解析失败时返回错误
func (s *CirclCVEService) FetchByCVEID(ctx context.Context, cveID string) (*model.CirclCVE, error) {
	start := time.Now()
	utilities.LogStart(circlComponent, "FetchByCVEID")

	detailURL := fmt.Sprintf("%s/vulnerability/%s", s.baseURL, cveID)

	body, err := s.fetchWithRetry(ctx, detailURL)
	if err != nil {
		utilities.LogError(circlComponent, "FetchByCVEID", err, time.Since(start),
			"cve_id="+cveID)
		return nil, err
	}

	cve, err := parseCirclResponse(body)
	if err != nil {
		utilities.LogError(circlComponent, "FetchByCVEID", err, time.Since(start),
			"cve_id="+cveID)
		return nil, fmt.Errorf("解析 CVE %s 详情失败: %w", cveID, err)
	}

	if err := s.repo.BulkUpsert(ctx, []model.CirclCVE{*cve}); err != nil {
		utilities.LogError(circlComponent, "FetchByCVEID", err, time.Since(start),
			"cve_id="+cveID)
		return nil, err
	}

	utilities.LogSuccess(circlComponent, "FetchByCVEID", time.Since(start),
		"cve_id="+cveID,
		"severity="+cve.Severity)
	return cve, nil
}

// fetchDetailBatch 对列表中的每条 CVE ID 逐一拉取详情，
// 解析失败的条目记录警告后跳过，不中断整批处理。
//
// 参数：
//   - ctx   : 请求上下文
//   - items : 列表页返回的 CVE ID 摘要切片
//
// 返回：
//   - []model.CirclCVE : 成功解析的 CVE 记录切片
//   - error            : 上下文取消时返回错误，单条解析失败不返回错误
func (s *CirclCVEService) fetchDetailBatch(
	ctx context.Context,
	items []model.CirclVulnerabilityListItem,
) ([]model.CirclCVE, error) {
	cves := make([]model.CirclCVE, 0, len(items))

	for i, item := range items {
		select {
		case <-ctx.Done():
			return cves, ctx.Err()
		default:
		}

		// 批内相邻请求之间主动限速。
		if i > 0 {
			time.Sleep(s.config.RateLimit)
		}

		detailURL := fmt.Sprintf("%s/vulnerability/%s", s.baseURL, item.VulnID)
		body, err := s.fetchWithRetry(ctx, detailURL)
		if err != nil {
			utilities.LogWarn(circlComponent, "fetchDetailBatch",
				fmt.Sprintf("拉取 %s 详情失败，跳过: %v", item.VulnID, err),
				0)
			continue
		}

		cve, err := parseCirclResponse(body)
		if err != nil {
			utilities.LogWarn(circlComponent, "fetchDetailBatch",
				fmt.Sprintf("解析 %s 详情失败，跳过: %v", item.VulnID, err),
				0)
			continue
		}

		cves = append(cves, *cve)
	}

	return cves, nil
}

// fetchWithRetry 发起 GET 请求，遇到 429 或 5xx 时按指数退避策略自动重试。
// 若主端点连续失败超过重试上限，自动切换到备用端点。
//
// 参数：
//   - ctx : 请求上下文
//   - url : 目标 URL
//
// 返回：
//   - []byte : 响应体字节切片
//   - error  : 重试耗尽或非预期错误时返回
func (s *CirclCVEService) fetchWithRetry(ctx context.Context, url string) ([]byte, error) {
	backoff := s.config.RetryBackoff

	for attempt := 0; attempt <= s.config.RetryMax; attempt++ {
		body, statusCode, err := s.doRequest(ctx, url)

		if err == nil {
			return body, nil
		}

		// 非 429 / 5xx 的错误不重试，直接返回。
		if statusCode != 429 && (statusCode < 500 || statusCode >= 600) {
			return nil, err
		}

		if attempt == s.config.RetryMax {
			// 主端点耗尽重试次数，尝试切换到备用端点。
			if s.baseURL == circlPrimaryBase {
				utilities.LogWarn(circlComponent, "fetchWithRetry",
					fmt.Sprintf("主端点不可用，切换到备用端点 %s", circlFallbackBase),
					0)
				s.baseURL = circlFallbackBase
				fallbackURL := strings.Replace(url, circlPrimaryBase, circlFallbackBase, 1)
				body, _, fallbackErr := s.doRequest(ctx, fallbackURL)
				if fallbackErr != nil {
					return nil, fmt.Errorf("主备端点均不可用: %w", fallbackErr)
				}
				return body, nil
			}
			return nil, fmt.Errorf("请求 %s 失败，已重试 %d 次: %w", url, s.config.RetryMax, err)
		}

		utilities.LogWarn(circlComponent, "fetchWithRetry",
			fmt.Sprintf("请求失败 (status=%d)，第 %d/%d 次重试，等待 %s",
				statusCode, attempt+1, s.config.RetryMax, backoff),
			0)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}

		backoff *= 2
	}

	return nil, fmt.Errorf("请求 %s 失败，已耗尽所有重试次数", url)
}

// doRequest 执行单次 HTTP GET 请求，返回响应体、HTTP 状态码和错误。
// 非 2xx 响应视为错误，同时返回状态码供调用方判断是否重试。
//
// 参数：
//   - ctx : 请求上下文
//   - url : 目标 URL
//
// 返回：
//   - []byte : 响应体字节切片（仅 2xx 时有效）
//   - int    : HTTP 状态码
//   - error  : 请求失败或非 2xx 时返回错误
func (s *CirclCVEService) doRequest(ctx context.Context, url string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("构建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", circlUserAgent)
	req.Header.Set("Accept", "application/json")

	// 若配置了 API Token，以 Bearer 方式附加到请求头。
	if s.config.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.APIToken)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("读取响应体失败: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, resp.StatusCode, nil
}

// parseCirclResponse 将 CIRCL API 返回的原始 JSON 字节解析并归一化为 CirclCVE 模型。
// 该函数提取 CVE 元数据、标题、英文描述、严重程度、CWE 列表、受影响软件包和参考链接。
//
// 参数：
//   - data : CIRCL API 响应的原始 JSON 字节
//
// 返回：
//   - *model.CirclCVE : 归一化后的 CVE 记录
//   - error           : JSON 解析失败或必要字段缺失时返回错误
func parseCirclResponse(data []byte) (*model.CirclCVE, error) {
	var raw circlAPIResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("JSON 反序列化失败: %w", err)
	}

	if raw.CVEMetadata.CVEID == "" {
		return nil, fmt.Errorf("响应中缺少 cveId 字段")
	}

	cve := &model.CirclCVE{
		CVEID:          raw.CVEMetadata.CVEID,
		State:          raw.CVEMetadata.State,
		AssignerOrgID:  raw.CVEMetadata.AssignerOrgID,
		AssignerShort:  raw.CVEMetadata.AssignerShort,
		Title:          raw.Containers.CNA.Title,
		Severity:       extractCirclSeverity(raw.Containers.CNA.Metrics),
		AffectedJSON:   raw.Containers.CNA.Affected,
		ReferencesJSON: raw.Containers.CNA.References,
	}

	// 提取英文描述，优先取 lang=en 的条目。
	cve.Description = extractCirclEnglishDescription(raw.Containers.CNA.Descriptions)

	// 提取 CWE ID 列表并序列化为 JSON 数组。
	cweIDs := extractCirclCWEIDs(raw.Containers.CNA.ProblemTypes)
	if len(cweIDs) > 0 {
		cweJSON, err := json.Marshal(cweIDs)
		if err == nil {
			cve.CWEIDs = cweJSON
		}
	}

	// 解析时间字段，格式为 RFC3339（如 2021-12-10T00:00:00.000Z）。
	cve.DatePublished = parseCirclTime(raw.CVEMetadata.DatePublished)
	cve.DateUpdated = parseCirclTime(raw.CVEMetadata.DateUpdated)
	cve.DateReserved = parseCirclTime(raw.CVEMetadata.DateReserved)

	return cve, nil
}

// extractCirclEnglishDescription 从描述列表中提取英文描述文本。
// 优先返回 lang=en 的条目；若无英文条目，返回第一条；若列表为空，返回空字符串。
//
// 参数：
//   - descs : 描述条目切片
//
// 返回：
//   - string : 描述文本
func extractCirclEnglishDescription(descs []circlDescription) string {
	for _, d := range descs {
		if strings.HasPrefix(d.Lang, "en") {
			return d.Value
		}
	}
	if len(descs) > 0 {
		return descs[0].Value
	}
	return ""
}

// extractCirclSeverity 从 metrics 字段中提取归一化的严重程度字符串。
// 支持 CVSS v3.x 的 baseSeverity 字段和 other.content 中的自由文本。
// 返回值统一为小写：critical | high | medium | low | unknown。
//
// 参数：
//   - metrics : metrics 数组
//
// 返回：
//   - string : 归一化后的严重程度
func extractCirclSeverity(metrics []circlMetric) string {
	for _, m := range metrics {
		if m.Other == nil {
			continue
		}

		// 尝试解析 other.content 为包含 baseSeverity 的 CVSS 对象。
		var cvssContent struct {
			BaseSeverity string `json:"baseSeverity"`
			Other        string `json:"other"`
		}
		if err := json.Unmarshal(m.Other.Content, &cvssContent); err == nil {
			if cvssContent.BaseSeverity != "" {
				return normalizeCirclSeverity(cvssContent.BaseSeverity)
			}
			if cvssContent.Other != "" {
				return normalizeCirclSeverity(cvssContent.Other)
			}
		}

		// 尝试将 content 直接解析为字符串（部分条目直接存储严重程度字符串）。
		var rawStr string
		if err := json.Unmarshal(m.Other.Content, &rawStr); err == nil && rawStr != "" {
			return normalizeCirclSeverity(rawStr)
		}
	}
	return "unknown"
}

// normalizeCirclSeverity 将任意大小写的严重程度字符串归一化为小写标准值。
// 未识别的值统一返回 "unknown"。
//
// 参数：
//   - s : 原始严重程度字符串
//
// 返回：
//   - string : 归一化后的严重程度（critical | high | medium | low | unknown）
func normalizeCirclSeverity(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium", "moderate":
		return "medium"
	case "low":
		return "low"
	default:
		return "unknown"
	}
}

// extractCirclCWEIDs 从 problemTypes 字段中提取所有 CWE ID 字符串列表。
// 去重后返回，保持原始顺序。
//
// 参数：
//   - problemTypes : problemTypes 数组
//
// 返回：
//   - []string : CWE ID 列表（如 ["CWE-502", "CWE-400"]）
func extractCirclCWEIDs(problemTypes []circlProblemType) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)

	for _, pt := range problemTypes {
		for _, desc := range pt.Descriptions {
			if desc.CWEID != "" {
				if _, exists := seen[desc.CWEID]; !exists {
					seen[desc.CWEID] = struct{}{}
					result = append(result, desc.CWEID)
				}
			}
		}
	}
	return result
}

// parseCirclTime 将 RFC3339 格式的时间字符串解析为 *time.Time。
// 解析失败或输入为空时返回 nil，不返回错误（时间字段为可选）。
//
// 参数：
//   - s : RFC3339 格式的时间字符串
//
// 返回：
//   - *time.Time : 解析成功时返回指针，失败时返回 nil
func parseCirclTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	// 尝试标准 RFC3339 格式。
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// 部分时间戳格式为 "2021-12-10T00:00:00.000Z"，尝试带毫秒的格式。
		t, err = time.Parse("2006-01-02T15:04:05.000Z", s)
		if err != nil {
			return nil
		}
	}
	return &t
}
