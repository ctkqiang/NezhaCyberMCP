package test

import (
	"context"
	"encoding/json"
	"nezha_cyber_mcp/internal/model"
	"nezha_cyber_mcp/internal/repository"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestDB 创建一个内存 SQLite 数据库并自动迁移 github_advisories 表结构。
//
// 参数：
//   - t : 测试上下文，初始化失败时终止测试
//
// 返回：
//   - *gorm.DB : 已完成迁移的测试数据库连接
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存 SQLite 失败: %v", err)
	}
	if err := db.AutoMigrate(&model.GithubAdvisory{}); err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}
	return db
}

// sampleAdvisory 构造一条用于测试的 GithubAdvisory 样本数据。
//
// 参数：
//   - ghsaID : 指定的 GHSA 标识符，用于区分不同测试用例
//
// 返回：
//   - model.GithubAdvisory : 填充了所有字段的样本公告
func sampleAdvisory(ghsaID string) model.GithubAdvisory {
	now := time.Now().UTC().Truncate(time.Second)
	cveID := "CVE-2024-12345"
	return model.GithubAdvisory{
		GHSAID:          ghsaID,
		CVEID:           &cveID,
		URL:             "https://api.github.com/advisories/" + ghsaID,
		HTMLURL:         "https://github.com/advisories/" + ghsaID,
		Summary:         "测试公告摘要",
		Description:     "漏洞的详细描述信息。",
		Type:            "reviewed",
		Severity:        "high",
		PublishedAt:     &now,
		UpdatedAt:       &now,
		WithdrawnAt:     nil,
		Vulnerabilities: json.RawMessage(`[{"package":{"name":"test-pkg","ecosystem":"npm"}}]`),
		References:      json.RawMessage(`[{"url":"https://example.com/advisory"}]`),
	}
}

// TestUpsert_Insert 验证 Upsert 在记录不存在时能正确执行插入操作，
// 并通过 GetByGHSAID 确认数据已写入且字段值正确。
func TestUpsert_Insert(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewGithubAdvisoryRepository(db)
	ctx := context.Background()

	adv := sampleAdvisory("GHSA-test-0001-aaaa")
	if err := repo.Upsert(ctx, &adv); err != nil {
		t.Fatalf("Upsert 插入失败: %v", err)
	}

	got, err := repo.GetByGHSAID(ctx, "GHSA-test-0001-aaaa")
	if err != nil {
		t.Fatalf("GetByGHSAID 查询失败: %v", err)
	}
	if got == nil {
		t.Fatal("期望查询到记录，实际返回 nil")
	}
	if got.Summary != adv.Summary {
		t.Errorf("Summary = %q; 期望 %q", got.Summary, adv.Summary)
	}
	if got.Severity != "high" {
		t.Errorf("Severity = %q; 期望 high", got.Severity)
	}
}

// TestUpsert_Update 验证 Upsert 在记录已存在时能正确执行更新操作，
// 确保冲突时所有可变字段均被刷新为最新值。
func TestUpsert_Update(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewGithubAdvisoryRepository(db)
	ctx := context.Background()

	adv := sampleAdvisory("GHSA-test-0002-bbbb")
	if err := repo.Upsert(ctx, &adv); err != nil {
		t.Fatalf("初始 Upsert 失败: %v", err)
	}

	adv.Summary = "补丁发布后的更新摘要"
	adv.Severity = "critical"
	if err := repo.Upsert(ctx, &adv); err != nil {
		t.Fatalf("更新 Upsert 失败: %v", err)
	}

	got, err := repo.GetByGHSAID(ctx, "GHSA-test-0002-bbbb")
	if err != nil {
		t.Fatalf("GetByGHSAID 查询失败: %v", err)
	}
	if got.Summary != "补丁发布后的更新摘要" {
		t.Errorf("Summary = %q; 期望更新后的值", got.Summary)
	}
	if got.Severity != "critical" {
		t.Errorf("Severity = %q; 期望 critical", got.Severity)
	}
}

// TestBulkUpsert 验证 BulkUpsert 能在单个事务中批量写入多条公告，
// 并通过 Count 确认数据库行数与写入数量一致。
func TestBulkUpsert(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewGithubAdvisoryRepository(db)
	ctx := context.Background()

	advisories := []model.GithubAdvisory{
		sampleAdvisory("GHSA-bulk-0001-aaaa"),
		sampleAdvisory("GHSA-bulk-0002-bbbb"),
		sampleAdvisory("GHSA-bulk-0003-cccc"),
	}

	if err := repo.BulkUpsert(ctx, advisories); err != nil {
		t.Fatalf("BulkUpsert 失败: %v", err)
	}

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count 查询失败: %v", err)
	}
	if count != 3 {
		t.Errorf("Count = %d; 期望 3", count)
	}
}

// TestBulkUpsert_Empty 验证 BulkUpsert 在传入 nil 或空切片时为无操作，
// 不应返回任何错误。
func TestBulkUpsert_Empty(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewGithubAdvisoryRepository(db)
	ctx := context.Background()

	if err := repo.BulkUpsert(ctx, nil); err != nil {
		t.Errorf("BulkUpsert(nil) 应为无操作，实际返回错误: %v", err)
	}
	if err := repo.BulkUpsert(ctx, []model.GithubAdvisory{}); err != nil {
		t.Errorf("BulkUpsert([]) 应为无操作，实际返回错误: %v", err)
	}
}

// TestGetByGHSAID_NotFound 验证 GetByGHSAID 在记录不存在时返回 (nil, nil)，
// 而非将"记录不存在"视为错误。
func TestGetByGHSAID_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewGithubAdvisoryRepository(db)
	ctx := context.Background()

	got, err := repo.GetByGHSAID(ctx, "GHSA-does-not-exist-0000")
	if err != nil {
		t.Fatalf("记录不存在时不应返回错误，实际得到: %v", err)
	}
	if got != nil {
		t.Errorf("记录不存在时期望返回 nil，实际得到 %+v", got)
	}
}

// TestCount_Empty 验证空表时 Count 返回 0。
func TestCount_Empty(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewGithubAdvisoryRepository(db)
	ctx := context.Background()

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count 查询失败: %v", err)
	}
	if count != 0 {
		t.Errorf("空表 Count = %d; 期望 0", count)
	}
}

// TestMigrate 验证 Migrate 能成功创建表结构，且多次调用保持幂等性（不报错）。
func TestMigrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开 SQLite 失败: %v", err)
	}
	repo := repository.NewGithubAdvisoryRepository(db)
	ctx := context.Background()

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("第一次 Migrate 失败: %v", err)
	}
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("第二次 Migrate 失败（应为幂等操作）: %v", err)
	}
}
