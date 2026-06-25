package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"nezha_cyber_mcp/internal/model"
	"nezha_cyber_mcp/internal/utilities"
	"strings"
	"time"

	"gorm.io/gorm"
)

// actionsComponent 是本包在日志中使用的组件名称标识。
const actionsComponent = "MCPActions"

// GetCVEParams 是 get_cve 工具的输入参数。
type GetCVEParams struct {
	// ID 是 CVE 唯一标识符，格式为 CVE-YYYY-NNNNN（如 CVE-2021-44228）。
	ID string `json:"id"`
}

// SearchCVEsParams 是 search_cves 工具的输入参数。
// 所有字段均为可选，至少提供一个过滤条件。
type SearchCVEsParams struct {
	// Keyword 在 title 和 description 字段中进行全文模糊匹配。
	Keyword string `json:"keyword,omitempty"`

	// Vendor 按分配机构简称过滤（如 "apache"、"microsoft"）。
	Vendor string `json:"vendor,omitempty"`

	// Product 在 affected JSON 字段中按产品名称过滤。
	Product string `json:"product,omitempty"`

	// CWE 按 CWE 编号过滤（如 "CWE-79"）。
	CWE string `json:"cwe,omitempty"`

	// DateFrom 是发布时间范围的起始日期，格式 YYYY-MM-DD。
	DateFrom string `json:"date_from,omitempty"`

	// DateTo 是发布时间范围的结束日期，格式 YYYY-MM-DD。
	DateTo string `json:"date_to,omitempty"`

	// Status 按 CVE 状态过滤：PUBLISHED | REJECTED | RESERVED。
	Status string `json:"status,omitempty"`

	// Limit 限制返回条数，默认 20，最大 100。
	Limit int `json:"limit,omitempty"`
}

// SearchByCPEParams 是 search_by_cpe 工具的输入参数。
type SearchByCPEParams struct {
	// CPEString 是 CPE 2.3 格式字符串（如 "cpe:2.3:a:apache:log4j:*:*:*:*:*:*:*:*"）
	// 或简短的 vendor/product 片段（如 "apache:log4j"）。
	CPEString string `json:"cpe_string"`
}

// BulkGetParams 是 bulk_get 工具的输入参数。
type BulkGetParams struct {
	// IDs 是 CVE ID 列表，最多 50 条。
	IDs []string `json:"ids"`
}

// FilterBySeverityParams 是 filter_by_severity 工具的输入参数。
type FilterBySeverityParams struct {
	// Severities 是严重程度过滤列表，可选值：critical | high | medium | low | unknown。
	// 传入多个值时取并集（OR 逻辑）。
	Severities []string `json:"severities"`

	// DateFrom 可选，限制发布时间范围起始日期，格式 YYYY-MM-DD。
	DateFrom string `json:"date_from,omitempty"`

	// DateTo 可选，限制发布时间范围结束日期，格式 YYYY-MM-DD。
	DateTo string `json:"date_to,omitempty"`

	// Limit 限制返回条数，默认 50，最大 200。
	Limit int `json:"limit,omitempty"`
}

// GetCWEParams 是 get_cwe 工具的输入参数。
type GetCWEParams struct {
	// CVEID 是目标 CVE 的唯一标识符。
	CVEID string `json:"cve_id"`
}

// GetReferencesParams 是 get_references 工具的输入参数。
type GetReferencesParams struct {
	// CVEID 是目标 CVE 的唯一标识符。
	CVEID string `json:"cve_id"`
}

// RelatedCVEsParams 是 related_cves 工具的输入参数。
type RelatedCVEsParams struct {
	// CVEID 是目标 CVE 的唯一标识符。
	CVEID string `json:"cve_id"`

	// Limit 限制返回条数，默认 10，最大 50。
	Limit int `json:"limit,omitempty"`
}

// WhatsNewParams 是 whats_new 工具的输入参数。
type WhatsNewParams struct {
	// SinceDate 是查询起始日期，格式 YYYY-MM-DD（如 "2024-01-01"）。
	SinceDate string `json:"since_date"`

	// MinSeverity 是最低严重程度过滤：critical | high | medium | low。
	// 传入 "high" 时返回 critical 和 high；传入 "critical" 时只返回 critical。
	MinSeverity string `json:"min_severity,omitempty"`

	// Limit 限制返回条数，默认 50，最大 200。
	Limit int `json:"limit,omitempty"`
}

// VulnTrendsParams 是 vuln_trends 工具的输入参数。
type VulnTrendsParams struct {
	// DateFrom 是统计时间范围起始日期，格式 YYYY-MM-DD。
	DateFrom string `json:"date_from"`

	// DateTo 是统计时间范围结束日期，格式 YYYY-MM-DD。
	DateTo string `json:"date_to"`

	// GroupBy 是分组维度：day | week | month | severity。
	GroupBy string `json:"group_by"`
}

// TopVendorsParams 是 top_vendors 工具的输入参数。
type TopVendorsParams struct {
	// DateFrom 可选，限制统计时间范围起始日期，格式 YYYY-MM-DD。
	DateFrom string `json:"date_from,omitempty"`

	// DateTo 可选，限制统计时间范围结束日期，格式 YYYY-MM-DD。
	DateTo string `json:"date_to,omitempty"`

	// Limit 限制返回条数，默认 10，最大 50。
	Limit int `json:"limit,omitempty"`
}

// TopProductsParams 是 top_products 工具的输入参数。
type TopProductsParams struct {
	// Vendor 可选，限制统计范围到指定厂商。
	Vendor string `json:"vendor,omitempty"`

	// DateFrom 可选，限制统计时间范围起始日期，格式 YYYY-MM-DD。
	DateFrom string `json:"date_from,omitempty"`

	// DateTo 可选，限制统计时间范围结束日期，格式 YYYY-MM-DD。
	DateTo string `json:"date_to,omitempty"`

	// Limit 限制返回条数，默认 10，最大 50。
	Limit int `json:"limit,omitempty"`
}

// SeverityDistributionParams 是 severity_distribution 工具的输入参数。
type SeverityDistributionParams struct {
	// DateFrom 可选，限制统计时间范围起始日期，格式 YYYY-MM-DD。
	DateFrom string `json:"date_from,omitempty"`

	// DateTo 可选，限制统计时间范围结束日期，格式 YYYY-MM-DD。
	DateTo string `json:"date_to,omitempty"`

	// Vendor 可选，限制统计范围到指定厂商。
	Vendor string `json:"vendor,omitempty"`
}

// MatchInventoryParams 是 match_inventory 工具的输入参数。
type MatchInventoryParams struct {
	// Packages 是软件包列表，每条格式为 "vendor:product" 或 "vendor:product:version"。
	Packages []string `json:"packages,omitempty"`

	// CPEList 是 CPE 2.3 格式字符串列表。
	CPEList []string `json:"cpe_list,omitempty"`

	// Limit 限制每个包返回的匹配 CVE 条数，默认 5，最大 20。
	Limit int `json:"limit,omitempty"`
}

// ---- 输出结构体 ----

// CVERecord 是工具返回的标准化 CVE 记录，字段类型明确，供 LLM 直接使用。
type CVERecord struct {
	CVEID         string          `json:"cve_id"`
	State         string          `json:"state"`
	AssignerShort string          `json:"assigner_short"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	Severity      string          `json:"severity"`
	CWEIDs        []string        `json:"cwe_ids"`
	Affected      json.RawMessage `json:"affected"`
	References    json.RawMessage `json:"references"`
	DatePublished *time.Time      `json:"date_published"`
	DateUpdated   *time.Time      `json:"date_updated"`
}

// TrendPoint 是 vuln_trends 工具返回的单个时间点数据。
type TrendPoint struct {
	Period string `json:"period"`
	Count  int64  `json:"count"`
}

// VendorCount 是 top_vendors / top_products 工具返回的单条统计记录。
type VendorCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// SeverityCount 是 severity_distribution 工具返回的单条统计记录。
type SeverityCount struct {
	Severity string `json:"severity"`
	Count    int64  `json:"count"`
}

// InventoryMatch 是 match_inventory 工具返回的单个包匹配结果。
type InventoryMatch struct {
	Package string      `json:"package"`
	CVEs    []CVERecord `json:"cves"`
	Total   int         `json:"total"`
}

// ---- Actions 结构体 ----

// Actions 封装了所有 MCP 工具的数据库查询逻辑。
// 通过 *gorm.DB 与数据库交互，所有方法均支持 context 传递。
type Actions struct {
	db *gorm.DB
}

// NewActions 构造一个绑定到指定 *gorm.DB 实例的 Actions。
//
// 参数：
//   - db : 已初始化的 GORM 数据库连接
//
// 返回：
//   - *Actions
func NewActions(db *gorm.DB) *Actions {
	return &Actions{db: db}
}

// ---- 辅助函数 ----

// toRecord 将 model.CirclCVE 转换为标准化的 CVERecord 输出结构。
func toRecord(c model.CirclCVE) CVERecord {
	var cweIDs []string
	if len(c.CWEIDs) > 0 {
		// 忽略解析错误，返回空切片而非 nil。
		_ = json.Unmarshal(c.CWEIDs, &cweIDs)
	}
	if cweIDs == nil {
		cweIDs = []string{}
	}
	return CVERecord{
		CVEID:         c.CVEID,
		State:         c.State,
		AssignerShort: c.AssignerShort,
		Title:         c.Title,
		Description:   c.Description,
		Severity:      c.Severity,
		CWEIDs:        cweIDs,
		Affected:      c.AffectedJSON,
		References:    c.ReferencesJSON,
		DatePublished: c.DatePublished,
		DateUpdated:   c.DateUpdated,
	}
}

// clampLimit 将 limit 限制在 [1, max] 范围内，若 limit <= 0 则使用 defaultVal。
func clampLimit(limit, defaultVal, max int) int {
	if limit <= 0 {
		return defaultVal
	}
	if limit > max {
		return max
	}
	return limit
}

// parseDateParam 将 YYYY-MM-DD 格式字符串解析为 time.Time。
// 解析失败时返回零值和错误。
func parseDateParam(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("日期格式无效 %q，期望 YYYY-MM-DD: %w", s, err)
	}
	return t, nil
}

// severityOrder 返回 minSeverity 对应的严重程度集合（含更高级别）。
// 例如传入 "high" 返回 ["critical", "high"]。
func severityOrder(minSeverity string) []string {
	order := []string{"critical", "high", "medium", "low", "unknown"}
	target := strings.ToLower(strings.TrimSpace(minSeverity))
	for i, s := range order {
		if s == target {
			return order[:i+1]
		}
	}
	return order
}

// ---- [ready] 工具实现 ----

// GetCVE 按 CVE ID 查询单条完整漏洞记录。
// 对应 MCP 工具：get_cve
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含目标 CVE ID
//
// 返回：
//   - *CVERecord : 完整的 CVE 记录
//   - error      : CVE 不存在或查询失败时返回错误
func (a *Actions) GetCVE(ctx context.Context, params GetCVEParams) (*CVERecord, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "GetCVE")

	if strings.TrimSpace(params.ID) == "" {
		return nil, fmt.Errorf("参数 id 不能为空")
	}

	var cve model.CirclCVE
	result := a.db.WithContext(ctx).
		Where("cve_id = ?", strings.ToUpper(strings.TrimSpace(params.ID))).
		First(&cve)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("CVE %q 不存在", params.ID)
		}
		utilities.LogError(actionsComponent, "GetCVE", result.Error, time.Since(start))
		return nil, fmt.Errorf("查询 CVE 失败: %w", result.Error)
	}

	rec := toRecord(cve)
	utilities.LogSuccess(actionsComponent, "GetCVE", time.Since(start), "cve_id="+params.ID)
	return &rec, nil
}

// SearchCVEs 按多维度条件搜索 CVE 记录。
// 对应 MCP 工具：search_cves
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 搜索条件（keyword/vendor/product/cwe/date_range/status/limit）
//
// 返回：
//   - []CVERecord : 匹配的 CVE 记录列表
//   - int         : 符合条件的总记录数（不受 limit 限制）
//   - error       : 查询失败时返回错误
func (a *Actions) SearchCVEs(ctx context.Context, params SearchCVEsParams) ([]CVERecord, int64, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "SearchCVEs")

	limit := clampLimit(params.Limit, 20, 100)
	q := a.db.WithContext(ctx).Model(&model.CirclCVE{})

	if params.Keyword != "" {
		kw := "%" + params.Keyword + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ?", kw, kw)
	}
	if params.Vendor != "" {
		q = q.Where("assigner_short ILIKE ?", "%"+params.Vendor+"%")
	}
	if params.Product != "" {
		q = q.Where("affected::text ILIKE ?", "%"+params.Product+"%")
	}
	if params.CWE != "" {
		q = q.Where("cwe_ids::text ILIKE ?", "%"+params.CWE+"%")
	}
	if params.Status != "" {
		q = q.Where("state = ?", strings.ToUpper(params.Status))
	}
	if params.DateFrom != "" {
		t, err := parseDateParam(params.DateFrom)
		if err != nil {
			return nil, 0, err
		}
		q = q.Where("date_published >= ?", t)
	}
	if params.DateTo != "" {
		t, err := parseDateParam(params.DateTo)
		if err != nil {
			return nil, 0, err
		}
		q = q.Where("date_published <= ?", t.Add(24*time.Hour-time.Second))
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		utilities.LogError(actionsComponent, "SearchCVEs", err, time.Since(start))
		return nil, 0, fmt.Errorf("统计搜索结果失败: %w", err)
	}

	var cves []model.CirclCVE
	if err := q.Order("date_published DESC NULLS LAST").Limit(limit).Find(&cves).Error; err != nil {
		utilities.LogError(actionsComponent, "SearchCVEs", err, time.Since(start))
		return nil, 0, fmt.Errorf("搜索 CVE 失败: %w", err)
	}

	records := make([]CVERecord, len(cves))
	for i, c := range cves {
		records[i] = toRecord(c)
	}

	utilities.LogSuccess(actionsComponent, "SearchCVEs", time.Since(start),
		fmt.Sprintf("matched=%d returned=%d", total, len(records)))
	return records, total, nil
}

// SearchByCPE 在 affected JSON 字段中按 CPE 字符串搜索匹配的 CVE。
// 对应 MCP 工具：search_by_cpe
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含 CPE 字符串（支持完整 CPE 2.3 或 vendor:product 片段）
//
// 返回：
//   - []CVERecord : 匹配的 CVE 记录列表
//   - error       : 查询失败时返回错误
func (a *Actions) SearchByCPE(ctx context.Context, params SearchByCPEParams) ([]CVERecord, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "SearchByCPE")

	if strings.TrimSpace(params.CPEString) == "" {
		return nil, fmt.Errorf("参数 cpe_string 不能为空")
	}

	var cves []model.CirclCVE
	err := a.db.WithContext(ctx).
		Where("affected::text ILIKE ?", "%"+params.CPEString+"%").
		Order("date_published DESC NULLS LAST").
		Limit(50).
		Find(&cves).Error
	if err != nil {
		utilities.LogError(actionsComponent, "SearchByCPE", err, time.Since(start))
		return nil, fmt.Errorf("按 CPE 搜索失败: %w", err)
	}

	records := make([]CVERecord, len(cves))
	for i, c := range cves {
		records[i] = toRecord(c)
	}

	utilities.LogSuccess(actionsComponent, "SearchByCPE", time.Since(start),
		fmt.Sprintf("cpe=%q returned=%d", params.CPEString, len(records)))
	return records, nil
}

// BulkGet 按 CVE ID 列表批量查询多条 CVE 记录。
// 对应 MCP 工具：bulk_get
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含 CVE ID 列表（最多 50 条）
//
// 返回：
//   - []CVERecord : 找到的 CVE 记录列表（未找到的 ID 不会报错，仅不出现在结果中）
//   - []string    : 未找到的 CVE ID 列表
//   - error       : 查询失败时返回错误
func (a *Actions) BulkGet(ctx context.Context, params BulkGetParams) ([]CVERecord, []string, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "BulkGet")

	if len(params.IDs) == 0 {
		return nil, nil, fmt.Errorf("参数 ids 不能为空")
	}
	if len(params.IDs) > 50 {
		return nil, nil, fmt.Errorf("ids 最多支持 50 条，当前传入 %d 条", len(params.IDs))
	}

	// 统一转大写，与数据库存储格式一致。
	upperIDs := make([]string, len(params.IDs))
	for i, id := range params.IDs {
		upperIDs[i] = strings.ToUpper(strings.TrimSpace(id))
	}

	var cves []model.CirclCVE
	if err := a.db.WithContext(ctx).
		Where("cve_id IN ?", upperIDs).
		Find(&cves).Error; err != nil {
		utilities.LogError(actionsComponent, "BulkGet", err, time.Since(start))
		return nil, nil, fmt.Errorf("批量查询 CVE 失败: %w", err)
	}

	// 计算未找到的 ID。
	found := make(map[string]struct{}, len(cves))
	for _, c := range cves {
		found[c.CVEID] = struct{}{}
	}
	var notFound []string
	for _, id := range upperIDs {
		if _, ok := found[id]; !ok {
			notFound = append(notFound, id)
		}
	}

	records := make([]CVERecord, len(cves))
	for i, c := range cves {
		records[i] = toRecord(c)
	}

	utilities.LogSuccess(actionsComponent, "BulkGet", time.Since(start),
		fmt.Sprintf("requested=%d found=%d not_found=%d", len(params.IDs), len(records), len(notFound)))
	return records, notFound, nil
}

// FilterBySeverity 按严重程度过滤 CVE 记录。
// 对应 MCP 工具：filter_by_severity
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含严重程度列表和可选的时间范围
//
// 返回：
//   - []CVERecord : 匹配的 CVE 记录列表
//   - int64       : 符合条件的总记录数
//   - error       : 查询失败时返回错误
func (a *Actions) FilterBySeverity(ctx context.Context, params FilterBySeverityParams) ([]CVERecord, int64, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "FilterBySeverity")

	if len(params.Severities) == 0 {
		return nil, 0, fmt.Errorf("参数 severities 不能为空，可选值：critical | high | medium | low | unknown")
	}

	// 归一化输入的严重程度值。
	normalized := make([]string, 0, len(params.Severities))
	valid := map[string]struct{}{
		"critical": {}, "high": {}, "medium": {}, "low": {}, "unknown": {},
	}
	for _, s := range params.Severities {
		sv := strings.ToLower(strings.TrimSpace(s))
		if _, ok := valid[sv]; !ok {
			return nil, 0, fmt.Errorf("无效的严重程度值 %q，可选值：critical | high | medium | low | unknown", s)
		}
		normalized = append(normalized, sv)
	}

	limit := clampLimit(params.Limit, 50, 200)
	q := a.db.WithContext(ctx).Model(&model.CirclCVE{}).Where("severity IN ?", normalized)

	if params.DateFrom != "" {
		t, err := parseDateParam(params.DateFrom)
		if err != nil {
			return nil, 0, err
		}
		q = q.Where("date_published >= ?", t)
	}
	if params.DateTo != "" {
		t, err := parseDateParam(params.DateTo)
		if err != nil {
			return nil, 0, err
		}
		q = q.Where("date_published <= ?", t.Add(24*time.Hour-time.Second))
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		utilities.LogError(actionsComponent, "FilterBySeverity", err, time.Since(start))
		return nil, 0, fmt.Errorf("统计严重程度过滤结果失败: %w", err)
	}

	var cves []model.CirclCVE
	if err := q.Order("date_published DESC NULLS LAST").Limit(limit).Find(&cves).Error; err != nil {
		utilities.LogError(actionsComponent, "FilterBySeverity", err, time.Since(start))
		return nil, 0, fmt.Errorf("按严重程度过滤失败: %w", err)
	}

	records := make([]CVERecord, len(cves))
	for i, c := range cves {
		records[i] = toRecord(c)
	}

	utilities.LogSuccess(actionsComponent, "FilterBySeverity", time.Since(start),
		fmt.Sprintf("severities=%v matched=%d returned=%d", normalized, total, len(records)))
	return records, total, nil
}

// GetCWE 提取指定 CVE 关联的 CWE 编号列表。
// 对应 MCP 工具：get_cwe
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含目标 CVE ID
//
// 返回：
//   - []string : CWE 编号列表（如 ["CWE-502", "CWE-400"]）
//   - error    : CVE 不存在或查询失败时返回错误
func (a *Actions) GetCWE(ctx context.Context, params GetCWEParams) ([]string, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "GetCWE")

	if strings.TrimSpace(params.CVEID) == "" {
		return nil, fmt.Errorf("参数 cve_id 不能为空")
	}

	var cve model.CirclCVE
	result := a.db.WithContext(ctx).
		Select("cve_id", "cwe_ids").
		Where("cve_id = ?", strings.ToUpper(strings.TrimSpace(params.CVEID))).
		First(&cve)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("CVE %q 不存在", params.CVEID)
		}
		utilities.LogError(actionsComponent, "GetCWE", result.Error, time.Since(start))
		return nil, fmt.Errorf("查询 CWE 失败: %w", result.Error)
	}

	var cweIDs []string
	if len(cve.CWEIDs) > 0 {
		if err := json.Unmarshal(cve.CWEIDs, &cweIDs); err != nil {
			utilities.LogError(actionsComponent, "GetCWE", err, time.Since(start))
			return nil, fmt.Errorf("解析 CWE 数据失败: %w", err)
		}
	}
	if cweIDs == nil {
		cweIDs = []string{}
	}

	utilities.LogSuccess(actionsComponent, "GetCWE", time.Since(start),
		"cve_id="+params.CVEID,
		fmt.Sprintf("cwe_count=%d", len(cweIDs)))
	return cweIDs, nil
}

// GetReferences 提取指定 CVE 的参考链接列表。
// 对应 MCP 工具：get_references
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含目标 CVE ID
//
// 返回：
//   - json.RawMessage : 参考链接 JSON 数组（原始格式，保留所有字段）
//   - error           : CVE 不存在或查询失败时返回错误
func (a *Actions) GetReferences(ctx context.Context, params GetReferencesParams) (json.RawMessage, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "GetReferences")

	if strings.TrimSpace(params.CVEID) == "" {
		return nil, fmt.Errorf("参数 cve_id 不能为空")
	}

	var cve model.CirclCVE
	result := a.db.WithContext(ctx).
		Select("cve_id", "references").
		Where("cve_id = ?", strings.ToUpper(strings.TrimSpace(params.CVEID))).
		First(&cve)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("CVE %q 不存在", params.CVEID)
		}
		utilities.LogError(actionsComponent, "GetReferences", result.Error, time.Since(start))
		return nil, fmt.Errorf("查询参考链接失败: %w", result.Error)
	}

	refs := cve.ReferencesJSON
	if len(refs) == 0 {
		refs = json.RawMessage("[]")
	}

	utilities.LogSuccess(actionsComponent, "GetReferences", time.Since(start), "cve_id="+params.CVEID)
	return refs, nil
}

// RelatedCVEs 查找与指定 CVE 相关的其他 CVE。
// 相关性判断依据：相同分配机构（assigner_short）或共享 CWE 编号。
// 对应 MCP 工具：related_cves
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含目标 CVE ID 和返回条数限制
//
// 返回：
//   - []CVERecord : 相关 CVE 记录列表，按发布时间降序排列
//   - error       : 目标 CVE 不存在或查询失败时返回错误
func (a *Actions) RelatedCVEs(ctx context.Context, params RelatedCVEsParams) ([]CVERecord, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "RelatedCVEs")

	if strings.TrimSpace(params.CVEID) == "" {
		return nil, fmt.Errorf("参数 cve_id 不能为空")
	}

	limit := clampLimit(params.Limit, 10, 50)
	targetID := strings.ToUpper(strings.TrimSpace(params.CVEID))

	// 先查目标 CVE 的 assigner_short 和 cwe_ids。
	var target model.CirclCVE
	if err := a.db.WithContext(ctx).
		Select("cve_id", "assigner_short", "cwe_ids").
		Where("cve_id = ?", targetID).
		First(&target).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("CVE %q 不存在", params.CVEID)
		}
		utilities.LogError(actionsComponent, "RelatedCVEs", err, time.Since(start))
		return nil, fmt.Errorf("查询目标 CVE 失败: %w", err)
	}

	q := a.db.WithContext(ctx).
		Where("cve_id != ?", targetID).
		Where("state = ?", "PUBLISHED")

	// 按 assigner_short 或 cwe_ids 匹配相关 CVE。
	if target.AssignerShort != "" && len(target.CWEIDs) > 0 {
		q = q.Where("assigner_short = ? OR cwe_ids::text ILIKE ?",
			target.AssignerShort, "%"+string(target.CWEIDs)+"%")
	} else if target.AssignerShort != "" {
		q = q.Where("assigner_short = ?", target.AssignerShort)
	} else if len(target.CWEIDs) > 0 {
		q = q.Where("cwe_ids::text ILIKE ?", "%"+string(target.CWEIDs)+"%")
	} else {
		return []CVERecord{}, nil
	}

	var cves []model.CirclCVE
	if err := q.Order("date_published DESC NULLS LAST").Limit(limit).Find(&cves).Error; err != nil {
		utilities.LogError(actionsComponent, "RelatedCVEs", err, time.Since(start))
		return nil, fmt.Errorf("查询相关 CVE 失败: %w", err)
	}

	records := make([]CVERecord, len(cves))
	for i, c := range cves {
		records[i] = toRecord(c)
	}

	utilities.LogSuccess(actionsComponent, "RelatedCVEs", time.Since(start),
		"cve_id="+params.CVEID,
		fmt.Sprintf("related=%d", len(records)))
	return records, nil
}

// WhatsNew 查询指定日期之后发布的新 CVE，可按最低严重程度过滤。
// 对应 MCP 工具：whats_new
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含起始日期和最低严重程度
//
// 返回：
//   - []CVERecord : 新发布的 CVE 记录列表，按发布时间降序排列
//   - int64       : 符合条件的总记录数
//   - error       : 查询失败时返回错误
func (a *Actions) WhatsNew(ctx context.Context, params WhatsNewParams) ([]CVERecord, int64, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "WhatsNew")

	if strings.TrimSpace(params.SinceDate) == "" {
		return nil, 0, fmt.Errorf("参数 since_date 不能为空，格式 YYYY-MM-DD")
	}

	since, err := parseDateParam(params.SinceDate)
	if err != nil {
		return nil, 0, err
	}

	limit := clampLimit(params.Limit, 50, 200)
	q := a.db.WithContext(ctx).Model(&model.CirclCVE{}).
		Where("date_published >= ?", since).
		Where("state = ?", "PUBLISHED")

	if params.MinSeverity != "" {
		severities := severityOrder(params.MinSeverity)
		q = q.Where("severity IN ?", severities)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		utilities.LogError(actionsComponent, "WhatsNew", err, time.Since(start))
		return nil, 0, fmt.Errorf("统计新 CVE 数量失败: %w", err)
	}

	var cves []model.CirclCVE
	if err := q.Order("date_published DESC").Limit(limit).Find(&cves).Error; err != nil {
		utilities.LogError(actionsComponent, "WhatsNew", err, time.Since(start))
		return nil, 0, fmt.Errorf("查询新 CVE 失败: %w", err)
	}

	records := make([]CVERecord, len(cves))
	for i, c := range cves {
		records[i] = toRecord(c)
	}

	utilities.LogSuccess(actionsComponent, "WhatsNew", time.Since(start),
		fmt.Sprintf("since=%s matched=%d returned=%d", params.SinceDate, total, len(records)))
	return records, total, nil
}

// VulnTrends 统计指定时间范围内的漏洞数量趋势。
// 对应 MCP 工具：vuln_trends
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含时间范围和分组维度（day | week | month | severity）
//
// 返回：
//   - []TrendPoint : 趋势数据点列表
//   - error        : 查询失败时返回错误
func (a *Actions) VulnTrends(ctx context.Context, params VulnTrendsParams) ([]TrendPoint, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "VulnTrends")

	if params.DateFrom == "" || params.DateTo == "" {
		return nil, fmt.Errorf("参数 date_from 和 date_to 均不能为空")
	}

	from, err := parseDateParam(params.DateFrom)
	if err != nil {
		return nil, err
	}
	to, err := parseDateParam(params.DateTo)
	if err != nil {
		return nil, err
	}

	type row struct {
		Period string
		Count  int64
	}

	var rows []row
	groupBy := strings.ToLower(strings.TrimSpace(params.GroupBy))

	switch groupBy {
	case "day":
		err = a.db.WithContext(ctx).
			Model(&model.CirclCVE{}).
			Select("TO_CHAR(date_published, 'YYYY-MM-DD') AS period, COUNT(*) AS count").
			Where("date_published BETWEEN ? AND ?", from, to.Add(24*time.Hour-time.Second)).
			Where("state = ?", "PUBLISHED").
			Group("period").
			Order("period ASC").
			Scan(&rows).Error
	case "week":
		err = a.db.WithContext(ctx).
			Model(&model.CirclCVE{}).
			Select("TO_CHAR(DATE_TRUNC('week', date_published), 'YYYY-MM-DD') AS period, COUNT(*) AS count").
			Where("date_published BETWEEN ? AND ?", from, to.Add(24*time.Hour-time.Second)).
			Where("state = ?", "PUBLISHED").
			Group("period").
			Order("period ASC").
			Scan(&rows).Error
	case "month":
		err = a.db.WithContext(ctx).
			Model(&model.CirclCVE{}).
			Select("TO_CHAR(DATE_TRUNC('month', date_published), 'YYYY-MM') AS period, COUNT(*) AS count").
			Where("date_published BETWEEN ? AND ?", from, to.Add(24*time.Hour-time.Second)).
			Where("state = ?", "PUBLISHED").
			Group("period").
			Order("period ASC").
			Scan(&rows).Error
	case "severity":
		err = a.db.WithContext(ctx).
			Model(&model.CirclCVE{}).
			Select("severity AS period, COUNT(*) AS count").
			Where("date_published BETWEEN ? AND ?", from, to.Add(24*time.Hour-time.Second)).
			Where("state = ?", "PUBLISHED").
			Group("period").
			Order("count DESC").
			Scan(&rows).Error
	default:
		return nil, fmt.Errorf("无效的 group_by 值 %q，可选值：day | week | month | severity", params.GroupBy)
	}

	if err != nil {
		utilities.LogError(actionsComponent, "VulnTrends", err, time.Since(start))
		return nil, fmt.Errorf("统计漏洞趋势失败: %w", err)
	}

	points := make([]TrendPoint, len(rows))
	for i, r := range rows {
		points[i] = TrendPoint{Period: r.Period, Count: r.Count}
	}

	utilities.LogSuccess(actionsComponent, "VulnTrends", time.Since(start),
		fmt.Sprintf("group_by=%s points=%d", groupBy, len(points)))
	return points, nil
}

// TopVendors 统计指定时间范围内漏洞数量最多的厂商排行。
// 对应 MCP 工具：top_vendors
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含可选时间范围和返回条数限制
//
// 返回：
//   - []VendorCount : 厂商漏洞数量排行列表
//   - error         : 查询失败时返回错误
func (a *Actions) TopVendors(ctx context.Context, params TopVendorsParams) ([]VendorCount, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "TopVendors")

	limit := clampLimit(params.Limit, 10, 50)
	q := a.db.WithContext(ctx).Model(&model.CirclCVE{}).
		Where("state = ?", "PUBLISHED").
		Where("assigner_short != ?", "")

	if params.DateFrom != "" {
		t, err := parseDateParam(params.DateFrom)
		if err != nil {
			return nil, err
		}
		q = q.Where("date_published >= ?", t)
	}
	if params.DateTo != "" {
		t, err := parseDateParam(params.DateTo)
		if err != nil {
			return nil, err
		}
		q = q.Where("date_published <= ?", t.Add(24*time.Hour-time.Second))
	}

	type row struct {
		Name  string
		Count int64
	}
	var rows []row
	if err := q.
		Select("assigner_short AS name, COUNT(*) AS count").
		Group("assigner_short").
		Order("count DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		utilities.LogError(actionsComponent, "TopVendors", err, time.Since(start))
		return nil, fmt.Errorf("统计厂商排行失败: %w", err)
	}

	result := make([]VendorCount, len(rows))
	for i, r := range rows {
		result[i] = VendorCount{Name: r.Name, Count: r.Count}
	}

	utilities.LogSuccess(actionsComponent, "TopVendors", time.Since(start),
		fmt.Sprintf("returned=%d", len(result)))
	return result, nil
}

// TopProducts 统计指定时间范围内漏洞数量最多的产品排行。
// 通过解析 affected JSON 字段中的 product 名称进行统计。
// 对应 MCP 工具：top_products
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含可选厂商过滤、时间范围和返回条数限制
//
// 返回：
//   - []VendorCount : 产品漏洞数量排行列表
//   - error         : 查询失败时返回错误
func (a *Actions) TopProducts(ctx context.Context, params TopProductsParams) ([]VendorCount, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "TopProducts")

	limit := clampLimit(params.Limit, 10, 50)

	// 通过 PostgreSQL JSON 函数从 affected 数组中提取 product 字段进行统计。
	q := a.db.WithContext(ctx).Model(&model.CirclCVE{}).
		Where("state = ?", "PUBLISHED").
		Where("affected IS NOT NULL AND affected != 'null'")

	if params.Vendor != "" {
		q = q.Where("assigner_short ILIKE ?", "%"+params.Vendor+"%")
	}
	if params.DateFrom != "" {
		t, err := parseDateParam(params.DateFrom)
		if err != nil {
			return nil, err
		}
		q = q.Where("date_published >= ?", t)
	}
	if params.DateTo != "" {
		t, err := parseDateParam(params.DateTo)
		if err != nil {
			return nil, err
		}
		q = q.Where("date_published <= ?", t.Add(24*time.Hour-time.Second))
	}

	type row struct {
		Name  string
		Count int64
	}
	var rows []row

	// 使用 PostgreSQL jsonb_array_elements 展开 affected 数组并提取 product 字段。
	rawSQL := `
		SELECT elem->>'product' AS name, COUNT(*) AS count
		FROM circl_cves,
		     jsonb_array_elements(affected::jsonb) AS elem
		WHERE state = 'PUBLISHED'
		  AND affected IS NOT NULL
		  AND affected != 'null'
		  AND elem->>'product' IS NOT NULL
		  AND elem->>'product' != ''
	`
	args := []interface{}{}

	if params.Vendor != "" {
		rawSQL += " AND assigner_short ILIKE ?"
		args = append(args, "%"+params.Vendor+"%")
	}
	if params.DateFrom != "" {
		t, _ := parseDateParam(params.DateFrom)
		rawSQL += " AND date_published >= ?"
		args = append(args, t)
	}
	if params.DateTo != "" {
		t, _ := parseDateParam(params.DateTo)
		rawSQL += " AND date_published <= ?"
		args = append(args, t.Add(24*time.Hour-time.Second))
	}

	rawSQL += fmt.Sprintf(" GROUP BY name ORDER BY count DESC LIMIT %d", limit)

	if err := a.db.WithContext(ctx).Raw(rawSQL, args...).Scan(&rows).Error; err != nil {
		utilities.LogError(actionsComponent, "TopProducts", err, time.Since(start))
		return nil, fmt.Errorf("统计产品排行失败: %w", err)
	}

	result := make([]VendorCount, len(rows))
	for i, r := range rows {
		result[i] = VendorCount{Name: r.Name, Count: r.Count}
	}

	utilities.LogSuccess(actionsComponent, "TopProducts", time.Since(start),
		fmt.Sprintf("returned=%d", len(result)))
	return result, nil
}

// SeverityDistribution 统计各严重程度的 CVE 数量分布。
// 对应 MCP 工具：severity_distribution
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含可选时间范围和厂商过滤
//
// 返回：
//   - []SeverityCount : 各严重程度的数量统计列表
//   - error           : 查询失败时返回错误
func (a *Actions) SeverityDistribution(ctx context.Context, params SeverityDistributionParams) ([]SeverityCount, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "SeverityDistribution")

	q := a.db.WithContext(ctx).Model(&model.CirclCVE{}).
		Where("state = ?", "PUBLISHED")

	if params.Vendor != "" {
		q = q.Where("assigner_short ILIKE ?", "%"+params.Vendor+"%")
	}
	if params.DateFrom != "" {
		t, err := parseDateParam(params.DateFrom)
		if err != nil {
			return nil, err
		}
		q = q.Where("date_published >= ?", t)
	}
	if params.DateTo != "" {
		t, err := parseDateParam(params.DateTo)
		if err != nil {
			return nil, err
		}
		q = q.Where("date_published <= ?", t.Add(24*time.Hour-time.Second))
	}

	type row struct {
		Severity string
		Count    int64
	}
	var rows []row
	if err := q.
		Select("severity, COUNT(*) AS count").
		Group("severity").
		Order("count DESC").
		Scan(&rows).Error; err != nil {
		utilities.LogError(actionsComponent, "SeverityDistribution", err, time.Since(start))
		return nil, fmt.Errorf("统计严重程度分布失败: %w", err)
	}

	result := make([]SeverityCount, len(rows))
	for i, r := range rows {
		result[i] = SeverityCount{Severity: r.Severity, Count: r.Count}
	}

	utilities.LogSuccess(actionsComponent, "SeverityDistribution", time.Since(start),
		fmt.Sprintf("buckets=%d", len(result)))
	return result, nil
}

// MatchInventory 对软件包清单或 CPE 列表进行漏洞匹配，返回每个包对应的 CVE 列表。
// 对应 MCP 工具：match_inventory
//
// 参数：
//   - ctx    : 请求上下文
//   - params : 包含软件包列表（vendor:product 格式）或 CPE 列表
//
// 返回：
//   - []InventoryMatch : 每个包的匹配结果列表
//   - error            : 查询失败时返回错误
func (a *Actions) MatchInventory(ctx context.Context, params MatchInventoryParams) ([]InventoryMatch, error) {
	start := time.Now()
	utilities.LogStart(actionsComponent, "MatchInventory")

	if len(params.Packages) == 0 && len(params.CPEList) == 0 {
		return nil, fmt.Errorf("参数 packages 或 cpe_list 至少提供一个")
	}

	limit := clampLimit(params.Limit, 5, 20)

	// 合并 packages 和 cpe_list 为统一的搜索词列表。
	targets := make([]string, 0, len(params.Packages)+len(params.CPEList))
	targets = append(targets, params.Packages...)
	targets = append(targets, params.CPEList...)

	results := make([]InventoryMatch, 0, len(targets))

	for _, target := range targets {
		var cves []model.CirclCVE
		err := a.db.WithContext(ctx).
			Where("affected::text ILIKE ? OR assigner_short ILIKE ?",
				"%"+target+"%", "%"+strings.SplitN(target, ":", 2)[0]+"%").
			Where("state = ?", "PUBLISHED").
			Order("date_published DESC NULLS LAST").
			Limit(limit).
			Find(&cves).Error
		if err != nil {
			utilities.LogError(actionsComponent, "MatchInventory", err, time.Since(start),
				"target="+target)
			continue
		}

		records := make([]CVERecord, len(cves))
		for i, c := range cves {
			records[i] = toRecord(c)
		}

		results = append(results, InventoryMatch{
			Package: target,
			CVEs:    records,
			Total:   len(records),
		})
	}

	utilities.LogSuccess(actionsComponent, "MatchInventory", time.Since(start),
		fmt.Sprintf("targets=%d matched_packages=%d", len(targets), len(results)))
	return results, nil
}

// ---- [needs data] 工具存根 ----

// ErrNotImplemented 是 [needs data] 工具返回的标准错误，说明该功能需要额外数据源。
type ErrNotImplemented struct {
	Tool       string
	DataSource string
}

func (e *ErrNotImplemented) Error() string {
	return fmt.Sprintf("工具 %q 需要额外数据源 %q，当前尚未集成", e.Tool, e.DataSource)
}

// GetKEVStatus 查询 CVE 是否在 CISA KEV（已知被利用漏洞）目录中。
// [needs data: CISA KEV] - 需要集成 https://www.cisa.gov/known-exploited-vulnerabilities-catalog
//
// 参数：
//   - ctx   : 请求上下文
//   - cveID : 目标 CVE ID
//
// 返回：
//   - error : 始终返回 ErrNotImplemented
func (a *Actions) GetKEVStatus(_ context.Context, cveID string) error {
	return &ErrNotImplemented{Tool: "get_kev_status", DataSource: "CISA KEV"}
}

// GetEPSS 查询 CVE 的 EPSS（漏洞利用预测评分系统）分数。
// [needs data: EPSS feed] - 需要集成 https://www.first.org/epss/
//
// 参数：
//   - ctx   : 请求上下文
//   - cveID : 目标 CVE ID
//
// 返回：
//   - error : 始终返回 ErrNotImplemented
func (a *Actions) GetEPSS(_ context.Context, cveID string) error {
	return &ErrNotImplemented{Tool: "get_epss", DataSource: "EPSS feed (first.org)"}
}

// Prioritize 对 CVE 列表按 CVSS + EPSS + KEV 综合评分进行优先级排序。
// [needs data: EPSS + CISA KEV] - 需要 EPSS 和 KEV 数据源才能完整实现。
//
// 参数：
//   - ctx     : 请求上下文
//   - cveList : CVE ID 列表
//
// 返回：
//   - error : 始终返回 ErrNotImplemented
func (a *Actions) Prioritize(_ context.Context, cveList []string) error {
	return &ErrNotImplemented{Tool: "prioritize", DataSource: "EPSS feed + CISA KEV"}
}

// MatchSBOM 解析 SBOM 文件并匹配相关 CVE。
// [needs data: SBOM parsing] - 需要集成 CycloneDX 或 SPDX 解析库。
//
// 参数：
//   - ctx      : 请求上下文
//   - sbomData : SBOM 文件内容（JSON 格式）
//
// 返回：
//   - error : 始终返回 ErrNotImplemented
func (a *Actions) MatchSBOM(_ context.Context, sbomData []byte) error {
	return &ErrNotImplemented{Tool: "match_sbom", DataSource: "SBOM parsing (CycloneDX/SPDX)"}
}
