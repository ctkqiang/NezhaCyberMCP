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
	// 来源：https://vulnerability.circl.lu/documentation/api.html
	// 公开端点无需认证，认证仅用于评论、bundle 和用户管理等写操作。
	circlPrimaryBase = "https://vulnerability.circl.lu/api"

	// circlFallbackBase 是 CIRCL API 的备用端点基础地址，主端点不可用时自动切换。
	// cve.circl.lu 是旧版 cve-search 服务，与 vulnerability.circl.lu 共享相同的 CVE 数据格式。
	circlFallbackBase = "https://cve.circl.lu/api"

	// circlUserAgent 是向 CIRCL API 发起请求时使用的 User-Agent 标识。
	circlUserAgent = "NezhaCyberMCP-CVE-Sync/1.0 (research; contact: research@example.com)"

	// circlDefaultTimeout 是单次 HTTP 请求的默认超时时间。
	circlDefaultTimeout = 30 * time.Second

	// circlDefaultLastCount 是 /api/last 端点每次返回的最近 CVE 条数。
	// CIRCL 文档说明该端点固定返回最近 30 条，包含 CAPEC、CWE 和 CPE 扩展信息。
	circlDefaultLastCount = 30

	// circlDefaultRetryMax 是遇到 429 / 5xx 时的最大重试次数。
	circlDefaultRetryMax = 5

	// circlDefaultRetryBackoff 是首次重试前的等待基准时间，每次重试翻倍（指数退避）。
	circlDefaultRetryBackoff = 2 * time.Second

	// circlDefaultRateLimit 是相邻两次 API 请求之间的最小间隔，避免触发速率限制。
	// CIRCL 公共 API 无明确文档限制，保守设置为 500ms。
	circlDefaultRateLimit = 500 * time.Millisecond
)

// ---- 解析用内部结构体（仅在本包内使用，不导出）----

// circlAPIResponse 是 CIRCL API /api/cve/{id} 端点返回的原始 JSON 结构。
// 格式遵循 CVE JSON 5.1 规范（dataType: "CVE_RECORD"）。
// 来源：https://vulnerability.circl.lu/api/cve/CVE-2021-44228
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
// 包含 cna（CVE Numbering Authority）和可选的 adp（Authorized Data Publisher）容器。
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

// circlMetricContent 对应 metrics[].other.content 字段的真实结构。
// 根据 CIRCL API 实际响应（如 CVE-2021-44228），content 是一个包含 "other" 键的对象：
//
//	{"other": "critical"}
//
// 部分条目也可能包含标准 CVSS baseSeverity 字段。
type circlMetricContent struct {
	Other        string `json:"other"`
	BaseSeverity string `json:"baseSeverity"`
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

// circlDbInfo 对应 /api/dbInfo 端点的响应结构，用于健康检查。
type circlDbInfo struct {
	LastUpdates map[string]interface{} `json:"last_updates"`
	DbSizes     map[string]int         `json:"db_sizes"`
}

// ---- 配置与服务结构体 ----

// CirclScraperConfig 保存 CIRCL CVE API 客户端的可调参数。
// 所有字段均有合理默认值，可通过 NewCirclCVEService 的 cfg 参数覆盖。
type CirclScraperConfig struct {
	// MaxRounds 限制最多调用 /api/last 的轮次，0 表示只拉取一次（/api/last 固定返回最近 30 条）。
	// 若需要按 vendor/product 搜索，使用 SearchTargets 字段。
	MaxRounds int

	// SearchTargets 是按 vendor/product 搜索的目标列表，格式为 "vendor/product"。
	// 例如：["microsoft/office", "apache/log4j"]
	// 若为空，则使用 /api/last 拉取最近更新的 CVE。
	SearchTargets []string

	// RequestTimeout 是单次 HTTP 请求的超时时间。
	RequestTimeout time.Duration

	// RetryMax 是遇到 429 / 5xx 时的最大重试次数。
	RetryMax int

	// RetryBackoff 是首次重试前的等待基准时间，每次重试翻倍。
	RetryBackoff time.Duration

	// RateLimit 是相邻两次请求之间的最小等待时间，用于主动限速。
	RateLimit time.Duration

	// APIToken 是 CIRCL API 的认证令牌。
	// 公开端点（/api/cve、/api/last、/api/search、/api/browse）无需认证。
	// 仅评论、bundle 和用户管理等写操作需要认证。
	// 从环境变量 CIRCL_API_TOKEN 读取，不得硬编码。
	APIToken string
}

// defaultCirclConfig 返回一组合理的默认 CIRCL API 客户端配置。
func defaultCirclConfig() CirclScraperConfig {
	return CirclScraperConfig{
		MaxRounds:      1,
		SearchTargets:  nil,
		RequestTimeout: circlDefaultTimeout,
		RetryMax:       circlDefaultRetryMax,
		RetryBackoff:   circlDefaultRetryBackoff,
		RateLimit:      circlDefaultRateLimit,
		APIToken:       os.Getenv("CIRCL_API_TOKEN"),
	}
}

// CirclCVEService 负责通过 CIRCL Vulnerability-Lookup API 拉取 CVE 数据并持久化到数据库。
// 支持主端点与备用端点自动切换，内置速率限制与指数退避重试。
//
// 支持的 API 端点（来源：https://vulnerability.circl.lu/documentation/api.html）：
//   - GET /api/cve/{CVE-ID}           : 按 CVE ID 查询单条漏洞详情
//   - GET /api/last                   : 获取最近 30 条更新的 CVE（含 CAPEC/CWE/CPE 扩展）
//   - GET /api/search/{vendor}/{product} : 按厂商和产品搜索漏洞
//   - GET /api/browse                 : 获取所有厂商列表
//   - GET /api/browse/{vendor}        : 获取指定厂商的所有产品
//   - GET /api/dbInfo                 : 获取数据库状态和更新时间
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
		if cfg.MaxRounds > 0 {
			c.MaxRounds = cfg.MaxRounds
		}
		if len(cfg.SearchTargets) > 0 {
			c.SearchTargets = cfg.SearchTargets
		}
		if cfg.RequestTimeout > 0 {
			c.RequestTimeout = cfg.RequestTimeout
		}
		if cfg.RetryMax > 0 {
			c.RetryMax = cfg.RetryMax
		}
		if cfg.RetryBackoff > 0 {
			c.RetryBackoff = cfg.RetryBackoff
		}
		if cfg.RateLimit >= 0 {
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

// CheckDBInfo 调用 /api/dbInfo 端点，验证 CIRCL API 可达性并记录数据库状态。
// 适用于服务启动时的健康检查。
//
// 参数：
//   - ctx : 请求上下文
//
// 返回：
//   - error : 请求失败或解析失败时返回错误
func (s *CirclCVEService) CheckDBInfo(ctx context.Context) error {
	start := time.Now()
	utilities.LogStart(circlComponent, "CheckDBInfo")

	url := s.baseURL + "/dbInfo"
	body, err := s.fetchWithRetry(ctx, url)
	if err != nil {
		utilities.LogError(circlComponent, "CheckDBInfo", err, time.Since(start))
		return fmt.Errorf("CheckDBInfo: %w", err)
	}

	var info circlDbInfo
	if err := json.Unmarshal(body, &info); err != nil {
		utilities.LogError(circlComponent, "CheckDBInfo", err, time.Since(start))
		return fmt.Errorf("CheckDBInfo 解析响应失败: %w", err)
	}

	total := info.DbSizes["total"]
	utilities.LogSuccess(circlComponent, "CheckDBInfo", time.Since(start),
		fmt.Sprintf("total_vulns=%d", total),
		fmt.Sprintf("sources=%d", len(info.DbSizes)))
	return nil
}

// ScrapeAndPersist 通过 CIRCL API 拉取 CVE 漏洞记录并持久化到数据库。
//
// 拉取策略（按优先级）：
//  1. 若 config.SearchTargets 非空，对每个 "vendor/product" 调用 /api/search/{vendor}/{product}
//  2. 否则调用 /api/last 获取最近 30 条更新的 CVE（可通过 MaxRounds 控制轮次）
//
// 每批拉取完成后立即批量 Upsert 到数据库，控制内存占用。
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

	var total int
	var err error

	if len(s.config.SearchTargets) > 0 {
		total, err = s.scrapeBySearchTargets(ctx, start)
	} else {
		total, err = s.scrapeByLast(ctx, start)
	}

	if err != nil {
		return total, err
	}

	utilities.LogSuccess(circlComponent, "ScrapeAndPersist", time.Since(start),
		fmt.Sprintf("total_cves=%d", total))
	return total, nil
}

// scrapeByLast 通过 /api/last 端点拉取最近更新的 CVE。
// 每轮固定返回 30 条，通过 MaxRounds 控制总轮次。
//
// 参数：
//   - ctx   : 请求上下文
//   - start : 操作起始时间，用于计算 elapsed
//
// 返回：
//   - int   : 本轮持久化的 CVE 总数
//   - error : 请求或写库失败时返回错误
func (s *CirclCVEService) scrapeByLast(ctx context.Context, start time.Time) (int, error) {
	total := 0
	maxRounds := s.config.MaxRounds
	if maxRounds <= 0 {
		maxRounds = 1
	}

	for round := 1; round <= maxRounds; round++ {
		if round > 1 {
			select {
			case <-ctx.Done():
				return total, ctx.Err()
			case <-time.After(s.config.RateLimit):
			}
		}

		lastURL := s.baseURL + "/last"
		utilities.LogProgress(circlComponent, "scrapeByLast",
			fmt.Sprintf("第 %d/%d 轮，请求 %s", round, maxRounds, lastURL))

		body, err := s.fetchWithRetry(ctx, lastURL)
		if err != nil {
			utilities.LogError(circlComponent, "scrapeByLast", err, time.Since(start),
				fmt.Sprintf("round=%d", round))
			return total, err
		}

		cves, err := parseCirclLastResponse(body)
		if err != nil {
			utilities.LogError(circlComponent, "scrapeByLast", err, time.Since(start),
				fmt.Sprintf("round=%d", round))
			return total, err
		}

		if len(cves) == 0 {
			utilities.LogProgress(circlComponent, "scrapeByLast",
				fmt.Sprintf("第 %d 轮返回空数据，停止", round))
			break
		}

		if err := s.repo.BulkUpsert(ctx, cves); err != nil {
			utilities.LogError(circlComponent, "scrapeByLast", err, time.Since(start))
			return total, err
		}
		total += len(cves)
		utilities.LogProgress(circlComponent, "scrapeByLast",
			fmt.Sprintf("第 %d 轮持久化 %d 条（累计: %d）", round, len(cves), total))
	}

	return total, nil
}

// scrapeBySearchTargets 对每个 "vendor/product" 目标调用 /api/search/{vendor}/{product}。
//
// 参数：
//   - ctx   : 请求上下文
//   - start : 操作起始时间，用于计算 elapsed
//
// 返回：
//   - int   : 本次持久化的 CVE 总数
//   - error : 请求或写库失败时返回错误
func (s *CirclCVEService) scrapeBySearchTargets(ctx context.Context, start time.Time) (int, error) {
	total := 0

	for i, target := range s.config.SearchTargets {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}

		if i > 0 {
			time.Sleep(s.config.RateLimit)
		}

		parts := strings.SplitN(target, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			utilities.LogWarn(circlComponent, "scrapeBySearchTargets",
				fmt.Sprintf("无效的搜索目标格式 %q，期望 vendor/product，跳过", target),
				time.Since(start))
			continue
		}

		searchURL := fmt.Sprintf("%s/search/%s/%s", s.baseURL, parts[0], parts[1])
		utilities.LogProgress(circlComponent, "scrapeBySearchTargets",
			fmt.Sprintf("搜索目标 %q: %s", target, searchURL))

		body, err := s.fetchWithRetry(ctx, searchURL)
		if err != nil {
			utilities.LogWarn(circlComponent, "scrapeBySearchTargets",
				fmt.Sprintf("搜索 %q 失败，跳过: %v", target, err),
				time.Since(start))
			continue
		}

		cves, err := parseCirclLastResponse(body)
		if err != nil {
			utilities.LogWarn(circlComponent, "scrapeBySearchTargets",
				fmt.Sprintf("解析 %q 响应失败，跳过: %v", target, err),
				time.Since(start))
			continue
		}

		if len(cves) == 0 {
			utilities.LogProgress(circlComponent, "scrapeBySearchTargets",
				fmt.Sprintf("目标 %q 返回空数据", target))
			continue
		}

		if err := s.repo.BulkUpsert(ctx, cves); err != nil {
			utilities.LogError(circlComponent, "scrapeBySearchTargets", err, time.Since(start),
				"target="+target)
			return total, err
		}
		total += len(cves)
		utilities.LogProgress(circlComponent, "scrapeBySearchTargets",
			fmt.Sprintf("目标 %q 持久化 %d 条（累计: %d）", target, len(cves), total))
	}

	return total, nil
}

// FetchByCVEID 按 CVE ID 调用 /api/cve/{CVE-ID} 拉取单条漏洞详情并持久化到数据库。
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

	// 正确端点：/api/cve/{CVE-ID}（来源：CIRCL API 文档）
	detailURL := fmt.Sprintf("%s/cve/%s", s.baseURL, cveID)

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

	if err := s.repo.Upsert(ctx, cve); err != nil {
		utilities.LogError(circlComponent, "FetchByCVEID", err, time.Since(start),
			"cve_id="+cveID)
		return nil, err
	}

	utilities.LogSuccess(circlComponent, "FetchByCVEID", time.Since(start),
		"cve_id="+cveID,
		"severity="+cve.Severity)
	return cve, nil
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

	// 公开端点无需认证；若配置了 Token，以 Bearer 方式附加，可用于写操作端点。
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

// parseCirclLastResponse 解析 /api/last 或 /api/search/{vendor}/{product} 端点返回的 JSON 数组。
// 这两个端点均返回 CVE 记录数组，每条记录与 /api/cve/{id} 格式相同。
// 解析失败的单条记录记录警告后跳过，不中断整批处理。
//
// 参数：
//   - data : API 响应的原始 JSON 字节（数组格式）
//
// 返回：
//   - []model.CirclCVE : 成功解析的 CVE 记录切片
//   - error            : 顶层 JSON 解析失败时返回错误
func parseCirclLastResponse(data []byte) ([]model.CirclCVE, error) {
	var rawItems []json.RawMessage
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return nil, fmt.Errorf("解析 CVE 数组失败: %w", err)
	}

	cves := make([]model.CirclCVE, 0, len(rawItems))
	for _, item := range rawItems {
		cve, err := parseCirclResponse(item)
		if err != nil {
			utilities.LogWarn(circlComponent, "parseCirclLastResponse",
				fmt.Sprintf("跳过无效条目: %v", err),
				0)
			continue
		}
		cves = append(cves, *cve)
	}
	return cves, nil
}

// parseCirclResponse 将 CIRCL API /api/cve/{id} 返回的原始 JSON 字节解析并归一化为 CirclCVE 模型。
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
//
// 根据 CIRCL API 真实响应（验证于 CVE-2021-44228），metrics[].other.content 的结构为：
//
//	{"other": "critical"}
//
// 同时兼容标准 CVSS 格式中的 baseSeverity 字段。
// 返回值统一为小写：critical | high | medium | low | unknown。
//
// 参数：
//   - metrics : metrics 数组
//
// 返回：
//   - string : 归一化后的严重程度
func extractCirclSeverity(metrics []circlMetric) string {
	for _, m := range metrics {
		if m.Other == nil || len(m.Other.Content) == 0 {
			continue
		}

		// 优先尝试解析为 circlMetricContent 对象（真实 API 响应格式）。
		// 真实格式：{"other": "critical"} 或 {"baseSeverity": "HIGH"}
		var content circlMetricContent
		if err := json.Unmarshal(m.Other.Content, &content); err == nil {
			if content.Other != "" {
				return normalizeCirclSeverity(content.Other)
			}
			if content.BaseSeverity != "" {
				return normalizeCirclSeverity(content.BaseSeverity)
			}
		}

		// 兼容：部分条目将 content 直接存储为字符串（如 "critical"）。
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
//   - []string : CWE ID 列表（如 ["CWE-502", "CWE-400", "CWE-20"]）
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
//   - s : RFC3339 格式的时间字符串（如 "2021-12-10T00:00:00.000Z"）
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
