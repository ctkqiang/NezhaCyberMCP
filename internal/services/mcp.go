package services

import (
	"context"
	"encoding/json"
	"fmt"
	"nezha_cyber_mcp/internal/functions"
	"nezha_cyber_mcp/internal/utilities"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

const mcpComponent = "MCPServer"

type MCPServer struct {
	server  *mcp.Server
	actions *functions.Actions
}

func NewMCPServer(db *gorm.DB) *MCPServer {
	actions := functions.NewActions(db)

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "nezha-cyber-mcp",
			Version: "v1.0.0",
		},
		&mcp.ServerOptions{
			Instructions: "NezhaCyber MCP 服务器，提供 CVE 漏洞数据库的查询、分析与优先级排序能力。" +
				"所有工具返回结构化 JSON，供 LLM 进行推理与报告生成。",
		},
	)

	svc := &MCPServer{server: server, actions: actions}
	svc.registerTools()
	svc.registerResources()
	svc.registerPrompts()

	return svc
}

func (s *MCPServer) Run(ctx context.Context) error {
	utilities.LogStart(mcpComponent, "Run")
	return s.server.Run(ctx, &mcp.StdioTransport{})
}

func (s *MCPServer) Server() *mcp.Server {
	return s.server
}

// SetDB 在运行时将数据库连接热注入到 MCPServer 及其 Actions。
// 供后台初始化 goroutine 在 DB 就绪后调用，无需重启 MCP 服务器。
//
// 参数：
//   - db : 已初始化的 GORM 数据库连接
func (s *MCPServer) SetDB(db *gorm.DB) {
	s.actions.SetDB(db)
}

func (s *MCPServer) registerTools() {
	mcp.AddTool(s.server, &mcp.Tool{
		Name: "get_cve",
		Description: "按 CVE ID 查询单条完整漏洞记录。" +
			"返回 CVE 的标题、描述、严重程度、CWE 编号、受影响软件包、参考链接及时间戳。" +
			"当用户提供具体 CVE ID（如 CVE-2021-44228）时调用此工具。",
	}, s.toolGetCVE)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "search_cves",
		Description: "按多维度条件搜索 CVE 记录。支持关键词全文搜索、厂商过滤、产品过滤、" +
			"CWE 编号过滤、发布时间范围过滤、状态过滤。" +
			"当用户需要查找某类漏洞、某厂商漏洞或某时间段内的漏洞时调用此工具。",
	}, s.toolSearchCVEs)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "search_by_cpe",
		Description: "在受影响软件包字段中按 CPE 字符串搜索匹配的 CVE。" +
			"支持完整 CPE 2.3 格式（如 cpe:2.3:a:apache:log4j:*）或简短的 vendor:product 片段。" +
			"当用户提供 CPE 字符串或需要查找特定软件组件的漏洞时调用此工具。",
	}, s.toolSearchByCPE)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "bulk_get",
		Description: "按 CVE ID 列表批量查询多条 CVE 记录，最多 50 条。" +
			"返回找到的记录列表及未找到的 ID 列表。" +
			"当用户提供多个 CVE ID 需要批量查询时调用此工具。",
	}, s.toolBulkGet)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "filter_by_severity",
		Description: "按严重程度过滤 CVE 记录。可选值：critical | high | medium | low | unknown。" +
			"支持多值并集过滤，可附加时间范围限制。" +
			"当用户需要查看特定严重程度的漏洞时调用此工具。",
	}, s.toolFilterBySeverity)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "get_kev_status",
		Description: "[needs data: CISA KEV] 查询 CVE 是否在 CISA 已知被利用漏洞目录中。" +
			"当前尚未集成 KEV 数据源，调用将返回明确的未实现错误。",
	}, s.toolGetKEVStatus)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "get_epss",
		Description: "[needs data: EPSS feed] 查询 CVE 的 EPSS 漏洞利用预测评分（0-1）。" +
			"当前尚未集成 EPSS 数据源，调用将返回明确的未实现错误。",
	}, s.toolGetEPSS)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "prioritize",
		Description: "[needs data: EPSS + CISA KEV] 对 CVE 列表按 CVSS + EPSS + KEV 综合评分排序。" +
			"当前尚未集成 EPSS 和 KEV 数据源，调用将返回明确的未实现错误。",
	}, s.toolPrioritize)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "match_inventory",
		Description: "对软件包清单或 CPE 列表进行漏洞匹配，返回每个包对应的 CVE 列表。" +
			"packages 格式为 vendor:product 或 vendor:product:version；" +
			"cpe_list 为 CPE 2.3 格式字符串列表。" +
			"当用户提供软件资产清单需要检查漏洞暴露面时调用此工具。",
	}, s.toolMatchInventory)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "match_sbom",
		Description: "[needs data: SBOM parsing] 解析 CycloneDX 或 SPDX 格式的 SBOM 文件并匹配相关 CVE。" +
			"当前尚未集成 SBOM 解析库，调用将返回明确的未实现错误。",
	}, s.toolMatchSBOM)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "get_cwe",
		Description: "提取指定 CVE 关联的 CWE 弱点编号列表（如 CWE-79、CWE-502）。" +
			"当用户需要了解漏洞的根本弱点类型时调用此工具。",
	}, s.toolGetCWE)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "get_references",
		Description: "提取指定 CVE 的参考链接列表，包含 NVD、厂商公告、PoC 等外部资源链接。" +
			"当用户需要查找漏洞的官方修复建议或技术细节时调用此工具。",
	}, s.toolGetReferences)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "related_cves",
		Description: "查找与指定 CVE 相关的其他 CVE，相关性依据：相同分配机构或共享 CWE 编号。" +
			"当用户需要了解某漏洞的关联漏洞族群时调用此工具。",
	}, s.toolRelatedCVEs)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "whats_new",
		Description: "查询指定日期之后发布的新 CVE，可按最低严重程度过滤。" +
			"当用户需要了解最新漏洞动态或进行每日/每周漏洞情报更新时调用此工具。",
	}, s.toolWhatsNew)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "vuln_trends",
		Description: "统计指定时间范围内的漏洞数量趋势，支持按 day | week | month | severity 分组。" +
			"当用户需要分析漏洞数量随时间的变化趋势时调用此工具。",
	}, s.toolVulnTrends)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "top_vendors",
		Description: "统计指定时间范围内漏洞数量最多的厂商排行榜。" +
			"当用户需要了解哪些厂商的产品漏洞最多时调用此工具。",
	}, s.toolTopVendors)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "top_products",
		Description: "统计指定时间范围内漏洞数量最多的产品排行榜，可按厂商过滤。" +
			"当用户需要了解哪些产品漏洞最多时调用此工具。",
	}, s.toolTopProducts)

	mcp.AddTool(s.server, &mcp.Tool{
		Name: "severity_distribution",
		Description: "统计各严重程度（critical/high/medium/low/unknown）的 CVE 数量分布。" +
			"可按时间范围和厂商过滤。当用户需要了解漏洞严重程度整体分布时调用此工具。",
	}, s.toolSeverityDistribution)
}

func (s *MCPServer) toolGetCVE(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.GetCVEParams,
) (*mcp.CallToolResult, *functions.CVERecord, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:get_cve")

	rec, err := s.actions.GetCVE(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:get_cve", err, time.Since(start))
		return toolError(err), nil, nil
	}

	utilities.LogSuccess(mcpComponent, "tool:get_cve", time.Since(start), "cve_id="+params.ID)
	return nil, rec, nil
}

func (s *MCPServer) toolSearchCVEs(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.SearchCVEsParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:search_cves")

	records, total, err := s.actions.SearchCVEs(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:search_cves", err, time.Since(start))
		return toolError(err), nil, nil
	}

	if len(records) == 0 {
		return toolEmpty("search_cves", "未找到匹配的 CVE 记录"), nil, nil
	}

	result := map[string]any{
		"total":   total,
		"count":   len(records),
		"results": records,
	}
	utilities.LogSuccess(mcpComponent, "tool:search_cves", time.Since(start),
		fmt.Sprintf("total=%d returned=%d", total, len(records)))
	return nil, result, nil
}

func (s *MCPServer) toolSearchByCPE(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.SearchByCPEParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:search_by_cpe")

	records, err := s.actions.SearchByCPE(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:search_by_cpe", err, time.Since(start))
		return toolError(err), nil, nil
	}

	if len(records) == 0 {
		return toolEmpty("search_by_cpe", fmt.Sprintf("未找到 CPE %q 相关的 CVE 记录", params.CPEString)), nil, nil
	}

	result := map[string]any{
		"cpe_string": params.CPEString,
		"count":      len(records),
		"results":    records,
	}
	utilities.LogSuccess(mcpComponent, "tool:search_by_cpe", time.Since(start),
		fmt.Sprintf("cpe=%q returned=%d", params.CPEString, len(records)))
	return nil, result, nil
}

func (s *MCPServer) toolBulkGet(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.BulkGetParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:bulk_get")

	records, notFound, err := s.actions.BulkGet(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:bulk_get", err, time.Since(start))
		return toolError(err), nil, nil
	}

	result := map[string]any{
		"found":     records,
		"not_found": notFound,
		"count":     len(records),
	}
	utilities.LogSuccess(mcpComponent, "tool:bulk_get", time.Since(start),
		fmt.Sprintf("found=%d not_found=%d", len(records), len(notFound)))
	return nil, result, nil
}

func (s *MCPServer) toolFilterBySeverity(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.FilterBySeverityParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:filter_by_severity")

	records, total, err := s.actions.FilterBySeverity(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:filter_by_severity", err, time.Since(start))
		return toolError(err), nil, nil
	}

	if len(records) == 0 {
		return toolEmpty("filter_by_severity", "未找到符合条件的 CVE 记录"), nil, nil
	}

	result := map[string]any{
		"total":      total,
		"count":      len(records),
		"severities": params.Severities,
		"results":    records,
	}
	utilities.LogSuccess(mcpComponent, "tool:filter_by_severity", time.Since(start),
		fmt.Sprintf("total=%d returned=%d", total, len(records)))
	return nil, result, nil
}

func (s *MCPServer) toolGetKEVStatus(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params struct {
		CVEID string `json:"cve_id"`
	},
) (*mcp.CallToolResult, any, error) {
	err := s.actions.GetKEVStatus(ctx, params.CVEID)
	return toolNotImplemented(err), nil, nil
}

func (s *MCPServer) toolGetEPSS(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params struct {
		CVEID string `json:"cve_id"`
	},
) (*mcp.CallToolResult, any, error) {
	err := s.actions.GetEPSS(ctx, params.CVEID)
	return toolNotImplemented(err), nil, nil
}

func (s *MCPServer) toolPrioritize(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params struct {
		CVEList []string `json:"cve_list"`
	},
) (*mcp.CallToolResult, any, error) {
	err := s.actions.Prioritize(ctx, params.CVEList)
	return toolNotImplemented(err), nil, nil
}

func (s *MCPServer) toolMatchInventory(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.MatchInventoryParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:match_inventory")

	matches, err := s.actions.MatchInventory(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:match_inventory", err, time.Since(start))
		return toolError(err), nil, nil
	}

	result := map[string]any{
		"matched_packages": len(matches),
		"results":          matches,
	}
	utilities.LogSuccess(mcpComponent, "tool:match_inventory", time.Since(start),
		fmt.Sprintf("packages=%d", len(matches)))
	return nil, result, nil
}

func (s *MCPServer) toolMatchSBOM(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params struct {
		SBOMData string `json:"sbom_data"`
	},
) (*mcp.CallToolResult, any, error) {
	err := s.actions.MatchSBOM(ctx, []byte(params.SBOMData))
	return toolNotImplemented(err), nil, nil
}

func (s *MCPServer) toolGetCWE(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.GetCWEParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:get_cwe")

	cweIDs, err := s.actions.GetCWE(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:get_cwe", err, time.Since(start))
		return toolError(err), nil, nil
	}

	result := map[string]any{
		"cve_id":  params.CVEID,
		"cwe_ids": cweIDs,
		"count":   len(cweIDs),
	}
	utilities.LogSuccess(mcpComponent, "tool:get_cwe", time.Since(start), "cve_id="+params.CVEID)
	return nil, result, nil
}

func (s *MCPServer) toolGetReferences(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.GetReferencesParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:get_references")

	refs, err := s.actions.GetReferences(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:get_references", err, time.Since(start))
		return toolError(err), nil, nil
	}

	var parsed any
	if err := json.Unmarshal(refs, &parsed); err != nil {
		parsed = string(refs)
	}

	result := map[string]any{
		"cve_id":     params.CVEID,
		"references": parsed,
	}
	utilities.LogSuccess(mcpComponent, "tool:get_references", time.Since(start), "cve_id="+params.CVEID)
	return nil, result, nil
}

func (s *MCPServer) toolRelatedCVEs(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.RelatedCVEsParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:related_cves")

	records, err := s.actions.RelatedCVEs(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:related_cves", err, time.Since(start))
		return toolError(err), nil, nil
	}

	if len(records) == 0 {
		return toolEmpty("related_cves", fmt.Sprintf("未找到与 %q 相关的 CVE", params.CVEID)), nil, nil
	}

	result := map[string]any{
		"cve_id":  params.CVEID,
		"count":   len(records),
		"results": records,
	}
	utilities.LogSuccess(mcpComponent, "tool:related_cves", time.Since(start),
		fmt.Sprintf("cve_id=%s related=%d", params.CVEID, len(records)))
	return nil, result, nil
}

func (s *MCPServer) toolWhatsNew(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.WhatsNewParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:whats_new")

	records, total, err := s.actions.WhatsNew(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:whats_new", err, time.Since(start))
		return toolError(err), nil, nil
	}

	if len(records) == 0 {
		return toolEmpty("whats_new", fmt.Sprintf("自 %s 起无新发布的 CVE", params.SinceDate)), nil, nil
	}

	result := map[string]any{
		"since_date":   params.SinceDate,
		"min_severity": params.MinSeverity,
		"total":        total,
		"count":        len(records),
		"results":      records,
	}
	utilities.LogSuccess(mcpComponent, "tool:whats_new", time.Since(start),
		fmt.Sprintf("since=%s total=%d returned=%d", params.SinceDate, total, len(records)))
	return nil, result, nil
}

func (s *MCPServer) toolVulnTrends(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.VulnTrendsParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:vuln_trends")

	points, err := s.actions.VulnTrends(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:vuln_trends", err, time.Since(start))
		return toolError(err), nil, nil
	}

	result := map[string]any{
		"date_from": params.DateFrom,
		"date_to":   params.DateTo,
		"group_by":  params.GroupBy,
		"count":     len(points),
		"data":      points,
	}
	utilities.LogSuccess(mcpComponent, "tool:vuln_trends", time.Since(start),
		fmt.Sprintf("group_by=%s points=%d", params.GroupBy, len(points)))
	return nil, result, nil
}

func (s *MCPServer) toolTopVendors(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.TopVendorsParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:top_vendors")

	vendors, err := s.actions.TopVendors(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:top_vendors", err, time.Since(start))
		return toolError(err), nil, nil
	}

	result := map[string]any{
		"count": len(vendors),
		"data":  vendors,
	}
	utilities.LogSuccess(mcpComponent, "tool:top_vendors", time.Since(start),
		fmt.Sprintf("returned=%d", len(vendors)))
	return nil, result, nil
}

func (s *MCPServer) toolTopProducts(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.TopProductsParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:top_products")

	products, err := s.actions.TopProducts(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:top_products", err, time.Since(start))
		return toolError(err), nil, nil
	}

	result := map[string]any{
		"vendor": params.Vendor,
		"count":  len(products),
		"data":   products,
	}
	utilities.LogSuccess(mcpComponent, "tool:top_products", time.Since(start),
		fmt.Sprintf("returned=%d", len(products)))
	return nil, result, nil
}

func (s *MCPServer) toolSeverityDistribution(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	params functions.SeverityDistributionParams,
) (*mcp.CallToolResult, any, error) {
	start := time.Now()
	utilities.LogStart(mcpComponent, "tool:severity_distribution")

	dist, err := s.actions.SeverityDistribution(ctx, params)
	if err != nil {
		utilities.LogError(mcpComponent, "tool:severity_distribution", err, time.Since(start))
		return toolError(err), nil, nil
	}

	result := map[string]any{
		"count": len(dist),
		"data":  dist,
	}
	utilities.LogSuccess(mcpComponent, "tool:severity_distribution", time.Since(start),
		fmt.Sprintf("buckets=%d", len(dist)))
	return nil, result, nil
}

func (s *MCPServer) registerResources() {
	s.server.AddResourceTemplate(
		&mcp.ResourceTemplate{
			Name:        "cve_record",
			URITemplate: "cve://{cve_id}",
			MIMEType:    "application/json",
			Description: "按 CVE ID 读取单条完整漏洞记录，URI 格式：cve://CVE-2021-44228",
		},
		s.resourceCVERecord,
	)

	s.server.AddResource(
		&mcp.Resource{
			Name:        "today_digest",
			URI:         "cve://digest/today",
			MIMEType:    "application/json",
			Description: "今日新增 critical 和 high 级别 CVE 的摘要列表，供快速情报感知使用。",
		},
		s.resourceTodayDigest,
	)
}

func (s *MCPServer) resourceCVERecord(
	ctx context.Context,
	req *mcp.ReadResourceRequest,
) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI
	cveID := strings.TrimPrefix(uri, "cve://")
	if cveID == "" {
		return nil, fmt.Errorf("无效的资源 URI %q，期望格式：cve://CVE-YYYY-NNNNN", uri)
	}

	rec, err := s.actions.GetCVE(ctx, functions.GetCVEParams{ID: cveID})
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(rec)
	if err != nil {
		return nil, fmt.Errorf("序列化 CVE 记录失败: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: uri, MIMEType: "application/json", Text: string(data)},
		},
	}, nil
}

func (s *MCPServer) resourceTodayDigest(
	ctx context.Context,
	_ *mcp.ReadResourceRequest,
) (*mcp.ReadResourceResult, error) {
	today := time.Now().Format("2006-01-02")

	records, total, err := s.actions.WhatsNew(ctx, functions.WhatsNewParams{
		SinceDate:   today,
		MinSeverity: "high",
		Limit:       50,
	})
	if err != nil {
		return nil, fmt.Errorf("获取今日摘要失败: %w", err)
	}

	digest := map[string]any{
		"date":    today,
		"total":   total,
		"count":   len(records),
		"results": records,
	}

	data, err := json.Marshal(digest)
	if err != nil {
		return nil, fmt.Errorf("序列化今日摘要失败: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "cve://digest/today", MIMEType: "application/json", Text: string(data)},
		},
	}, nil
}

func (s *MCPServer) registerPrompts() {
	s.server.AddPrompt(
		&mcp.Prompt{
			Name:        "triage_for_stakeholder",
			Description: "将 CVE 列表转化为面向非技术利益相关者的风险摘要，使用业务语言描述影响与建议行动。",
		},
		promptTriageForStakeholder,
	)

	s.server.AddPrompt(
		&mcp.Prompt{
			Name:        "patching_priorities",
			Description: "为指定产品生成补丁优先级建议，按严重程度和可利用性排序，输出结构化的修复计划。",
		},
		promptPatchingPriorities,
	)

	s.server.AddPrompt(
		&mcp.Prompt{
			Name:        "draft_security_advisory",
			Description: "根据 CVE 记录起草安全公告，包含漏洞描述、受影响版本、修复建议和参考链接。",
		},
		promptDraftSecurityAdvisory,
	)

	s.server.AddPrompt(
		&mcp.Prompt{
			Name:        "weekly_vuln_report",
			Description: "根据软件资产清单生成每周漏洞报告，汇总新增漏洞、严重程度分布和修复优先级。",
		},
		promptWeeklyVulnReport,
	)
}

func promptTriageForStakeholder(
	_ context.Context,
	req *mcp.GetPromptRequest,
) (*mcp.GetPromptResult, error) {
	cveList := req.Params.Arguments["cve_list"]
	if cveList == "" {
		return nil, fmt.Errorf("参数 cve_list 不能为空，请提供逗号分隔的 CVE ID 列表")
	}

	text := fmt.Sprintf(
		"你是一名安全顾问，需要向非技术背景的业务负责人解释以下漏洞的业务风险。\n\n"+
			"CVE 列表：%s\n\n"+
			"请按以下结构输出：\n"+
			"1. 执行摘要（2-3 句话，说明整体风险级别）\n"+
			"2. 每个 CVE 的业务影响（避免技术术语，聚焦业务中断、数据泄露、合规风险）\n"+
			"3. 建议的行动项（按优先级排序，包含时间建议）\n"+
			"4. 不采取行动的潜在后果\n\n"+
			"语言要求：简洁、清晰、无技术术语。",
		cveList,
	)

	return &mcp.GetPromptResult{
		Description: "面向非技术利益相关者的 CVE 风险分类提示",
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: text}},
		},
	}, nil
}

func promptPatchingPriorities(
	_ context.Context,
	req *mcp.GetPromptRequest,
) (*mcp.GetPromptResult, error) {
	product := req.Params.Arguments["product"]
	if product == "" {
		return nil, fmt.Errorf("参数 product 不能为空，请提供产品名称")
	}
	cveData := req.Params.Arguments["cve_data"]

	text := fmt.Sprintf(
		"你是一名漏洞管理工程师，需要为产品 %q 制定补丁优先级计划。\n\n"+
			"CVE 数据：\n%s\n\n"+
			"请按以下结构输出：\n"+
			"1. 立即修复（Critical/High，有已知利用）\n"+
			"2. 计划修复（High/Medium，无已知利用）\n"+
			"3. 监控观察（Low/Medium，暂无修复必要）\n"+
			"4. 每个优先级的具体 CVE 列表及修复建议\n"+
			"5. 预计修复工作量评估\n\n"+
			"输出格式：结构化 JSON，字段包含 priority、cve_ids、rationale、estimated_effort。",
		product, cveData,
	)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("产品 %q 的补丁优先级建议", product),
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: text}},
		},
	}, nil
}

func promptDraftSecurityAdvisory(
	_ context.Context,
	req *mcp.GetPromptRequest,
) (*mcp.GetPromptResult, error) {
	cveID := req.Params.Arguments["cve_id"]
	if cveID == "" {
		return nil, fmt.Errorf("参数 cve_id 不能为空")
	}
	cveData := req.Params.Arguments["cve_data"]

	text := fmt.Sprintf(
		"你是一名安全工程师，需要根据以下 CVE 信息起草一份正式的安全公告。\n\n"+
			"CVE ID：%s\n"+
			"CVE 数据：\n%s\n\n"+
			"安全公告必须包含以下章节：\n"+
			"1. 标题（格式：[严重程度] 产品名称 - 漏洞类型）\n"+
			"2. 摘要（1-2 句话描述漏洞）\n"+
			"3. 受影响版本（列表格式）\n"+
			"4. 漏洞详情（技术描述，包含攻击向量和影响范围）\n"+
			"5. 修复建议（升级版本或临时缓解措施）\n"+
			"6. 参考链接\n"+
			"7. 发布日期和严重程度评级\n\n"+
			"输出格式：Markdown。",
		cveID, cveData,
	)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("CVE %s 的安全公告草稿", cveID),
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: text}},
		},
	}, nil
}

func promptWeeklyVulnReport(
	_ context.Context,
	req *mcp.GetPromptRequest,
) (*mcp.GetPromptResult, error) {
	inventory := req.Params.Arguments["inventory"]
	if inventory == "" {
		return nil, fmt.Errorf("参数 inventory 不能为空，请提供软件资产清单（逗号分隔的 vendor:product 列表）")
	}
	weekData := req.Params.Arguments["week_data"]

	text := fmt.Sprintf(
		"你是一名安全运营工程师，需要根据以下软件资产清单和本周漏洞数据生成每周漏洞报告。\n\n"+
			"资产清单：%s\n"+
			"本周漏洞数据：\n%s\n\n"+
			"报告必须包含以下章节：\n"+
			"1. 本周概况（新增漏洞总数、严重程度分布）\n"+
			"2. 高危漏洞详情（Critical/High 级别，逐条说明）\n"+
			"3. 受影响资产清单（哪些产品受到影响）\n"+
			"4. 本周修复进度（已修复 vs 待修复）\n"+
			"5. 下周行动计划（优先级排序的修复任务）\n"+
			"6. 趋势分析（与上周对比）\n\n"+
			"输出格式：Markdown，适合直接发送给安全团队。",
		inventory, weekData,
	)

	return &mcp.GetPromptResult{
		Description: "每周漏洞报告生成提示",
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: text}},
		},
	}, nil
}

func toolError(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: err.Error()},
		},
	}
}

func toolEmpty(tool, message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf(`{"tool":"%s","empty":true,"message":"%s"}`, tool, message),
			},
		},
	}
}

func toolNotImplemented(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf(`{"error":"not_implemented","detail":"%s"}`, err.Error()),
			},
		},
	}
}
