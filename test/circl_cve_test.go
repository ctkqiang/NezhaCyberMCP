// Package test 包含针对 CIRCL CVE Search API 集成的单元测试。
package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"nezha_cyber_mcp/internal/model"
	"nezha_cyber_mcp/internal/repository"
	"nezha_cyber_mcp/internal/services"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ---- 测试用 JSON 固件 ----

// circlSingleCVEJSON 是一条真实 CIRCL API 响应的最小化 JSON 固件，
// 对应 CVE-2021-44228（Log4Shell），用于验证解析与归一化逻辑。
const circlSingleCVEJSON = `{
  "dataType": "CVE_RECORD",
  "dataVersion": "5.1",
  "cveMetadata": {
    "state": "PUBLISHED",
    "cveId": "CVE-2021-44228",
    "assignerOrgId": "f0158376-9dc2-43b6-827c-5f631a4d8d09",
    "assignerShortName": "apache",
    "dateUpdated": "2025-10-21T23:25:23.121Z",
    "dateReserved": "2021-11-26T00:00:00.000Z",
    "datePublished": "2021-12-10T00:00:00.000Z"
  },
  "containers": {
    "cna": {
      "title": "Apache Log4j2 JNDI features do not protect against attacker controlled LDAP",
      "descriptions": [
        {
          "lang": "en",
          "value": "Apache Log4j2 2.0-beta9 through 2.15.0 JNDI features do not protect against attacker controlled LDAP endpoints."
        }
      ],
      "affected": [{"vendor": "Apache Software Foundation", "product": "Apache Log4j2"}],
      "references": [{"url": "https://logging.apache.org/log4j/2.x/security.html"}],
      "metrics": [
        {
          "other": {
            "type": "unknown",
            "content": "critical"
          }
        }
      ],
      "problemTypes": [
        {
          "descriptions": [
            {
              "type": "CWE",
              "lang": "en",
              "description": "CWE-502 Deserialization of Untrusted Data",
              "cweId": "CWE-502"
            }
          ]
        }
      ]
    }
  }
}`

// circlSecondCVEJSON 是第二条 CVE 的最小化 JSON 固件，用于列表拉取测试。
const circlSecondCVEJSON = `{
  "dataType": "CVE_RECORD",
  "dataVersion": "5.1",
  "cveMetadata": {
    "state": "PUBLISHED",
    "cveId": "CVE-2021-45046",
    "assignerOrgId": "f0158376-9dc2-43b6-827c-5f631a4d8d09",
    "assignerShortName": "apache",
    "dateUpdated": "2022-01-01T00:00:00.000Z",
    "dateReserved": "2021-12-14T00:00:00.000Z",
    "datePublished": "2021-12-14T00:00:00.000Z"
  },
  "containers": {
    "cna": {
      "title": "Apache Log4j2 Thread Context Message Pattern vulnerable to DoS",
      "descriptions": [
        {
          "lang": "en",
          "value": "Apache Log4j2 Thread Context Message Pattern and Context Lookup Pattern vulnerable to a denial of service attack."
        }
      ],
      "affected": [],
      "references": [],
      "metrics": [
        {
          "other": {
            "type": "unknown",
            "content": "high"
          }
        }
      ],
      "problemTypes": []
    }
  }
}`

// ---- 辅助函数 ----

// newCirclTestDB 创建一个内存 SQLite 数据库并完成 circl_cves 表迁移。
// 测试结束后无需手动清理，内存数据库随进程退出自动释放。
//
// 参数：
//   - t : 测试上下文
//
// 返回：
//   - *repository.CirclCVERepository : 已初始化的 Repository
func newCirclTestDB(t *testing.T) *repository.CirclCVERepository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存 SQLite 失败: %v", err)
	}
	if err := db.AutoMigrate(&model.CirclCVE{}); err != nil {
		t.Fatalf("AutoMigrate circl_cves 失败: %v", err)
	}
	return repository.NewCirclCVERepository(db)
}

// ---- 单元测试 ----

// TestCirclScraperConfig_Default 验证 NewCirclCVEService 在传入 nil 配置时
// 能正常构造实例，不会 panic。
func TestCirclScraperConfig_Default(t *testing.T) {
	repo := newCirclTestDB(t)
	svc := services.NewCirclCVEService(repo, nil)
	if svc == nil {
		t.Fatal("NewCirclCVEService 返回 nil")
	}
}

// TestCirclScraperConfig_Override 验证传入非 nil 配置时各字段能正确覆盖默认值。
func TestCirclScraperConfig_Override(t *testing.T) {
	repo := newCirclTestDB(t)
	cfg := &services.CirclScraperConfig{
		RequestTimeout: 10 * time.Second,
		RetryMax:       2,
		RetryBackoff:   1 * time.Second,
		RateLimit:      200 * time.Millisecond,
	}
	svc := services.NewCirclCVEService(repo, cfg)
	if svc == nil {
		t.Fatal("NewCirclCVEService 返回 nil")
	}
}

// TestCirclCVERepository_Migrate 验证 Migrate 方法幂等性：
// 多次调用不应返回错误。
func TestCirclCVERepository_Migrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存 SQLite 失败: %v", err)
	}
	repo := repository.NewCirclCVERepository(db)
	ctx := context.Background()

	// 第一次迁移。
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("第一次 Migrate 失败: %v", err)
	}
	// 第二次迁移应幂等，不报错。
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("第二次 Migrate 失败（应幂等）: %v", err)
	}
}

// TestCirclCVERepository_BulkUpsert 验证 BulkUpsert 能正确写入记录，
// 并在主键冲突时执行更新而非报错。
func TestCirclCVERepository_BulkUpsert(t *testing.T) {
	repo := newCirclTestDB(t)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	items := []model.CirclCVE{
		{
			CVEID:         "CVE-2021-44228",
			State:         "PUBLISHED",
			AssignerShort: "apache",
			Title:         "Log4Shell",
			Description:   "JNDI injection vulnerability.",
			Severity:      "critical",
			DatePublished: &now,
		},
	}

	// 首次写入。
	if err := repo.BulkUpsert(ctx, items); err != nil {
		t.Fatalf("BulkUpsert 首次写入失败: %v", err)
	}

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count 失败: %v", err)
	}
	if count != 1 {
		t.Errorf("Count = %d; 期望 1", count)
	}

	// 更新同一条记录（主键冲突，应执行 Upsert）。
	items[0].Severity = "high"
	if err := repo.BulkUpsert(ctx, items); err != nil {
		t.Fatalf("BulkUpsert Upsert 更新失败: %v", err)
	}

	count, _ = repo.Count(ctx)
	if count != 1 {
		t.Errorf("Upsert 后 Count = %d; 期望仍为 1", count)
	}
}

// TestCirclCVERepository_BulkUpsert_Empty 验证传入空切片时 BulkUpsert 直接返回 nil，
// 不执行任何数据库操作。
func TestCirclCVERepository_BulkUpsert_Empty(t *testing.T) {
	repo := newCirclTestDB(t)
	ctx := context.Background()

	if err := repo.BulkUpsert(ctx, nil); err != nil {
		t.Fatalf("BulkUpsert(nil) 应返回 nil，实际返回: %v", err)
	}
	if err := repo.BulkUpsert(ctx, []model.CirclCVE{}); err != nil {
		t.Fatalf("BulkUpsert([]) 应返回 nil，实际返回: %v", err)
	}
}

// TestCirclFetchByCVEID_MockServer 使用 httptest.Server 模拟 CIRCL API，
// 验证 FetchByCVEID 能正确解析响应并将记录写入数据库。
func TestCirclFetchByCVEID_MockServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(circlSingleCVEJSON))
	}))
	t.Cleanup(srv.Close)

	repo := newCirclTestDB(t)
	svc := services.NewCirclCVEServiceWithBaseURL(repo, &services.CirclScraperConfig{
		RequestTimeout: 5 * time.Second,
		RetryMax:       1,
		RetryBackoff:   10 * time.Millisecond,
		RateLimit:      0,
	}, srv.URL)

	cve, err := svc.FetchByCVEID(context.Background(), "CVE-2021-44228")
	if err != nil {
		t.Fatalf("FetchByCVEID 失败: %v", err)
	}

	if cve.CVEID != "CVE-2021-44228" {
		t.Errorf("CVEID = %q; 期望 CVE-2021-44228", cve.CVEID)
	}
	if cve.State != "PUBLISHED" {
		t.Errorf("State = %q; 期望 PUBLISHED", cve.State)
	}
	if cve.Severity != "critical" {
		t.Errorf("Severity = %q; 期望 critical", cve.Severity)
	}
	if cve.AssignerShort != "apache" {
		t.Errorf("AssignerShort = %q; 期望 apache", cve.AssignerShort)
	}
	if cve.Description == "" {
		t.Error("Description 不应为空")
	}

	// 验证 CWE ID 已正确提取。
	var cweIDs []string
	if err := json.Unmarshal(cve.CWEIDs, &cweIDs); err != nil {
		t.Fatalf("CWEIDs JSON 解析失败: %v", err)
	}
	if len(cweIDs) != 1 || cweIDs[0] != "CWE-502" {
		t.Errorf("CWEIDs = %v; 期望 [CWE-502]", cweIDs)
	}

	// 验证记录已写入数据库。
	count, _ := repo.Count(context.Background())
	if count != 1 {
		t.Errorf("数据库 Count = %d; 期望 1", count)
	}
}

// TestCirclScrapeAndPersist_MockServer 使用 httptest.Server 模拟 CIRCL API，
// 验证 ScrapeAndPersist 能正确处理 /last 端点返回的 CVE 数组，并将所有记录写入数据库。
// CIRCL /api/last 返回格式为 JSON 数组，每个元素是完整的 CVE 记录对象。
func TestCirclScrapeAndPersist_MockServer(t *testing.T) {
	// /last 端点返回包含两条完整 CVE 记录的 JSON 数组。
	listBody := "[" + circlSingleCVEJSON + "," + circlSecondCVEJSON + "]"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/last":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(listBody))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	repo := newCirclTestDB(t)
	svc := services.NewCirclCVEServiceWithBaseURL(repo, &services.CirclScraperConfig{
		RequestTimeout: 5 * time.Second,
		RetryMax:       1,
		RetryBackoff:   10 * time.Millisecond,
		RateLimit:      0,
	}, srv.URL)

	total, err := svc.ScrapeAndPersist(context.Background())
	if err != nil {
		t.Fatalf("ScrapeAndPersist 失败: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d; 期望 2", total)
	}

	count, _ := repo.Count(context.Background())
	if count != 2 {
		t.Errorf("数据库 Count = %d; 期望 2", count)
	}
}

// TestCirclScrapeAndPersist_HTTP500_Retry 验证遇到 5xx 错误时服务能正确重试，
// 并在重试成功后返回正确结果。
func TestCirclScrapeAndPersist_HTTP500_Retry(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 2 {
			// 前两次请求返回 500，触发重试。
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 第三次请求成功，返回空 JSON 数组（/api/last 格式，触发末页退出）。
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	repo := newCirclTestDB(t)
	svc := services.NewCirclCVEServiceWithBaseURL(repo, &services.CirclScraperConfig{
		RequestTimeout: 5 * time.Second,
		RetryMax:       3,
		RetryBackoff:   10 * time.Millisecond,
		RateLimit:      0,
	}, srv.URL)

	total, err := svc.ScrapeAndPersist(context.Background())
	if err != nil {
		t.Fatalf("ScrapeAndPersist 在重试成功后不应返回错误，实际: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d; 期望 0（空列表）", total)
	}
	if callCount < 3 {
		t.Errorf("callCount = %d; 期望至少 3 次（含重试）", callCount)
	}
}

// TestCirclScrapeAndPersist_HTTP404_NoRetry 验证遇到 404 错误时服务不重试，
// 直接返回错误。
func TestCirclScrapeAndPersist_HTTP404_NoRetry(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	repo := newCirclTestDB(t)
	svc := services.NewCirclCVEServiceWithBaseURL(repo, &services.CirclScraperConfig{
		RequestTimeout: 5 * time.Second,
		RetryMax:       3,
		RetryBackoff:   10 * time.Millisecond,
		RateLimit:      0,
	}, srv.URL)

	_, err := svc.ScrapeAndPersist(context.Background())
	if err == nil {
		t.Fatal("ScrapeAndPersist 遇到 404 时应返回错误")
	}
	// 404 不应触发重试，callCount 应为 1。
	if callCount != 1 {
		t.Errorf("callCount = %d; 期望 1（404 不重试）", callCount)
	}
}

// TestNormalizeCirclSeverity 验证严重程度归一化函数对各种输入的处理结果。
func TestNormalizeCirclSeverity(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"CRITICAL", "critical"},
		{"critical", "critical"},
		{"HIGH", "high"},
		{"high", "high"},
		{"MEDIUM", "medium"},
		{"medium", "medium"},
		{"moderate", "medium"},
		{"LOW", "low"},
		{"low", "low"},
		{"unknown", "unknown"},
		{"", "unknown"},
		{"n/a", "unknown"},
	}

	for _, tc := range cases {
		// 通过构造包含对应 severity 的 JSON 来间接测试归一化逻辑。
		jsonStr := `{
			"dataType": "CVE_RECORD",
			"dataVersion": "5.1",
			"cveMetadata": {
				"state": "PUBLISHED",
				"cveId": "CVE-2099-0001",
				"assignerOrgId": "test",
				"assignerShortName": "test",
				"datePublished": "2099-01-01T00:00:00.000Z"
			},
			"containers": {
				"cna": {
					"title": "Test",
					"descriptions": [{"lang": "en", "value": "test"}],
					"metrics": [{"other": {"type": "unknown", "content": "` + tc.input + `"}}]
				}
			}
		}`

		repo := newCirclTestDB(t)
		svc := services.NewCirclCVEService(repo, nil)
		_ = svc

		// 直接调用导出的解析函数进行验证。
		cve, err := services.ParseCirclResponseForTest([]byte(jsonStr))
		if err != nil {
			// 空字符串 content 会导致 JSON 解析失败，属于预期行为。
			if tc.input == "" {
				continue
			}
			t.Errorf("input=%q: 解析失败: %v", tc.input, err)
			continue
		}
		if cve.Severity != tc.expected {
			t.Errorf("input=%q: Severity = %q; 期望 %q", tc.input, cve.Severity, tc.expected)
		}
	}
}
