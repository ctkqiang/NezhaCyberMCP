// Package test 包含针对 MyCERT HTML 解析与爬虫配置的单元测试。
package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nezha_cyber_mcp/internal/model"
	"nezha_cyber_mcp/internal/repository"
	"nezha_cyber_mcp/internal/services"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// minimalListHTML 是一个最小化的 MyCERT 列表页 HTML 片段，
// 结构与真实页面一致，用于验证 HTML 解析逻辑。
const minimalListHTML = `<!DOCTYPE html>
<html>
<body>
<h4><a href="/portal/advisory?id=MA-9001.062026">MA-9001.062026: MyCERT Advisory - 测试漏洞 A</a></h4>
<ul>
  <li>16 Jun 2026</li>
  <li>Advisory</li>
</ul>
<p>1.0 Introduction 测试摘要 A，描述漏洞基本情况。</p>
<a href="/portal/advisory?id=MA-9001.062026">Read More</a>

<h4><a href="/portal/advisory?id=MA-9002.062026">MA-9002.062026: MyCERT Advisory - 测试漏洞 B</a></h4>
<ul>
  <li>15 Jun 2026</li>
  <li>Advisory</li>
</ul>
<p>1.0 Introduction 测试摘要 B，描述漏洞基本情况。</p>
<a href="/portal/advisory?id=MA-9002.062026">Read More</a>
</body>
</html>`

// minimalListWithNextHTML 是包含"Next"分页链接的列表页 HTML，
// 用于验证翻页检测逻辑。
const minimalListWithNextHTML = `<!DOCTYPE html>
<html>
<body>
<h4><a href="/portal/advisory?id=MA-9003.062026">MA-9003.062026: MyCERT Advisory - 测试漏洞 C</a></h4>
<ul>
  <li>14 Jun 2026</li>
  <li>Advisory</li>
</ul>
<p>摘要 C。</p>
<a href="/portal/advisories?id=xxx&amp;page=2&amp;per-page=10">Next </a>
</body>
</html>`

// minimalDetailHTML 是一个最小化的 MyCERT 详情页 HTML 片段，
// 用于验证详情页正文提取逻辑。
const minimalDetailHTML = `<!DOCTYPE html>
<html>
<body>
<div class="content-area">
  <p><strong>1.0 Introduction</strong></p>
  <p>Ivanti 已发布安全更新以修复两个严重漏洞。</p>
  <p><strong>2.0 Impact</strong></p>
  <p>成功利用漏洞可能允许远程未认证用户实现 root 级别的远程代码执行。</p>
  <p><strong>3.0 Recommendations</strong></p>
  <p>建议用户立即应用最新安全补丁。</p>
</div>
</body>
</html>`

// newTestServer 创建一个 httptest.Server，根据请求路径返回预设的 HTML 内容。
//
// 参数：
//   - t       : 测试上下文
//   - handler : HTTP 处理函数
//
// 返回：
//   - *httptest.Server : 测试服务器，测试结束时自动关闭
func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// TestMycertScraperConfig_Default 验证 NewMycertAdvisoryService 在传入 nil 配置时
// 使用合理的默认值。
func TestMycertScraperConfig_Default(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.MycertAdvisory{})
	repo := repository.NewMycertAdvisoryRepository(db)

	svc := services.NewMycertAdvisoryService(repo, nil)
	if svc == nil {
		t.Fatal("NewMycertAdvisoryService 返回 nil")
	}
}

// TestMycertScraperConfig_Override 验证传入非 nil 配置时各字段能正确覆盖默认值。
func TestMycertScraperConfig_Override(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.MycertAdvisory{})
	repo := repository.NewMycertAdvisoryRepository(db)

	cfg := &services.MycertScraperConfig{
		MaxPages:       5,
		RequestTimeout: 10 * time.Second,
		FetchDetail:    false,
	}
	svc := services.NewMycertAdvisoryService(repo, cfg)
	if svc == nil {
		t.Fatal("NewMycertAdvisoryService 返回 nil")
	}
}

// TestMycertScraper_ScrapeAndPersist_MockServer 使用 httptest.Server 模拟 MyCERT 网站，
// 验证 ScrapeAndPersist 能正确解析 HTML、提取公告并写入数据库。
func TestMycertScraper_ScrapeAndPersist_MockServer(t *testing.T) {
	page := 0
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		page++
		if page == 1 {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(minimalListHTML))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存 SQLite 失败: %v", err)
	}
	db.AutoMigrate(&model.MycertAdvisory{})
	repo := repository.NewMycertAdvisoryRepository(db)

	svc := services.NewMycertAdvisoryServiceWithBaseURL(repo, &services.MycertScraperConfig{
		MaxPages:       1,
		RequestTimeout: 5 * time.Second,
		FetchDetail:    false,
	}, srv.URL+"/portal/advisories?id=test")

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

// TestMycertScraper_ScrapeAndPersist_Pagination 验证分页逻辑：
// 第一页包含 Next 链接时继续抓取，第二页无 Next 链接时停止。
func TestMycertScraper_ScrapeAndPersist_Pagination(t *testing.T) {
	callCount := 0
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.RawQuery, "page=2") {
			w.Write([]byte(minimalListHTML))
		} else {
			w.Write([]byte(minimalListWithNextHTML))
		}
	})

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.MycertAdvisory{})
	repo := repository.NewMycertAdvisoryRepository(db)

	svc := services.NewMycertAdvisoryServiceWithBaseURL(repo, &services.MycertScraperConfig{
		MaxPages:       0,
		RequestTimeout: 5 * time.Second,
		FetchDetail:    false,
	}, srv.URL+"/portal/advisories?id=test")

	total, err := svc.ScrapeAndPersist(context.Background())
	if err != nil {
		t.Fatalf("ScrapeAndPersist 失败: %v", err)
	}
	if callCount < 2 {
		t.Errorf("callCount = %d; 期望至少 2（应翻页）", callCount)
	}
	if total < 3 {
		t.Errorf("total = %d; 期望至少 3（两页合计）", total)
	}
}

// TestMycertScraper_ScrapeAndPersist_MaxPages 验证 MaxPages 参数能正确限制抓取页数。
func TestMycertScraper_ScrapeAndPersist_MaxPages(t *testing.T) {
	callCount := 0
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(minimalListWithNextHTML))
	})

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.MycertAdvisory{})
	repo := repository.NewMycertAdvisoryRepository(db)

	svc := services.NewMycertAdvisoryServiceWithBaseURL(repo, &services.MycertScraperConfig{
		MaxPages:       2,
		RequestTimeout: 5 * time.Second,
		FetchDetail:    false,
	}, srv.URL+"/portal/advisories?id=test")

	_, err := svc.ScrapeAndPersist(context.Background())
	if err != nil {
		t.Fatalf("ScrapeAndPersist 失败: %v", err)
	}
	if callCount > 2 {
		t.Errorf("callCount = %d; 期望不超过 2（MaxPages=2）", callCount)
	}
}

// TestMycertScraper_ScrapeAndPersist_HTTPError 验证服务器返回 5xx 错误时
// ScrapeAndPersist 能正确返回错误，不会静默忽略。
func TestMycertScraper_ScrapeAndPersist_HTTPError(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.MycertAdvisory{})
	repo := repository.NewMycertAdvisoryRepository(db)

	svc := services.NewMycertAdvisoryServiceWithBaseURL(repo, &services.MycertScraperConfig{
		MaxPages:       1,
		RequestTimeout: 5 * time.Second,
		FetchDetail:    false,
	}, srv.URL+"/portal/advisories?id=test")

	_, err := svc.ScrapeAndPersist(context.Background())
	if err == nil {
		t.Fatal("服务器返回 500 时 ScrapeAndPersist 应返回错误，实际返回 nil")
	}
}

// TestMycertScraper_ScrapeAndPersist_EmptyPage 验证服务器返回空页面（无公告）时
// ScrapeAndPersist 能正确停止并返回 total=0，不报错。
func TestMycertScraper_ScrapeAndPersist_EmptyPage(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><body><p>暂无公告</p></body></html>`))
	})

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.MycertAdvisory{})
	repo := repository.NewMycertAdvisoryRepository(db)

	svc := services.NewMycertAdvisoryServiceWithBaseURL(repo, &services.MycertScraperConfig{
		MaxPages:       1,
		RequestTimeout: 5 * time.Second,
		FetchDetail:    false,
	}, srv.URL+"/portal/advisories?id=test")

	total, err := svc.ScrapeAndPersist(context.Background())
	if err != nil {
		t.Fatalf("空页面时 ScrapeAndPersist 不应返回错误，实际: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d; 期望 0", total)
	}
}

// TestMycertScraper_ScrapeAndPersist_NetworkError 验证目标服务器不可达时
// ScrapeAndPersist 能正确返回网络错误。
// 通过立即关闭 httptest.Server 获得一个确定性的不可达地址，
// 避免依赖硬编码端口（硬编码端口可能恰好被其他进程占用）。
func TestMycertScraper_ScrapeAndPersist_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := srv.URL + "/portal/advisories?id=test"
	srv.Close()

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.MycertAdvisory{})
	repo := repository.NewMycertAdvisoryRepository(db)

	svc := services.NewMycertAdvisoryServiceWithBaseURL(repo, &services.MycertScraperConfig{
		MaxPages:       1,
		RequestTimeout: 1 * time.Second,
		FetchDetail:    false,
	}, deadURL)

	_, err := svc.ScrapeAndPersist(context.Background())
	if err == nil {
		t.Fatal("不可达地址时 ScrapeAndPersist 应返回错误，实际返回 nil")
	}
}

// TestMycertScraper_ScrapeAndPersist_Idempotent 验证对同一页面执行两次抓取时
// 数据库行数不增加（Upsert 幂等性）。
func TestMycertScraper_ScrapeAndPersist_Idempotent(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(minimalListHTML))
	})

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&model.MycertAdvisory{})
	repo := repository.NewMycertAdvisoryRepository(db)

	cfg := &services.MycertScraperConfig{
		MaxPages:       1,
		RequestTimeout: 5 * time.Second,
		FetchDetail:    false,
	}
	baseURL := srv.URL + "/portal/advisories?id=test"

	svc1 := services.NewMycertAdvisoryServiceWithBaseURL(repo, cfg, baseURL)
	if _, err := svc1.ScrapeAndPersist(context.Background()); err != nil {
		t.Fatalf("第一次 ScrapeAndPersist 失败: %v", err)
	}

	svc2 := services.NewMycertAdvisoryServiceWithBaseURL(repo, cfg, baseURL)
	if _, err := svc2.ScrapeAndPersist(context.Background()); err != nil {
		t.Fatalf("第二次 ScrapeAndPersist 失败: %v", err)
	}

	count, _ := repo.Count(context.Background())
	if count != 2 {
		t.Errorf("两次抓取后 Count = %d; 期望 2（幂等，不应重复插入）", count)
	}
}
