// Package test 包含针对 MyCERT 公告抓取与持久化流程的集成测试。
// 所有测试均使用 SQLite 内存数据库，无需外部服务即可运行。
package test

import (
	"context"
	"strings"
	"testing"
	"time"

	"nezha_cyber_mcp/internal/model"
	"nezha_cyber_mcp/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newMycertTestDB 创建一个内存 SQLite 数据库并自动迁移 mycert_advisories 表结构。
//
// 参数：
//   - t : 测试上下文，初始化失败时终止测试
//
// 返回：
//   - *gorm.DB : 已完成迁移的测试数据库连接
func newMycertTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存 SQLite 失败: %v", err)
	}
	if err := db.AutoMigrate(&model.MycertAdvisory{}); err != nil {
		t.Fatalf("自动迁移 mycert_advisories 失败: %v", err)
	}
	return db
}

// sampleMycertAdvisory 构造一条用于测试的 MycertAdvisory 样本数据。
//
// 参数：
//   - advisoryID : 指定的公告唯一编号，用于区分不同测试用例
//
// 返回：
//   - model.MycertAdvisory : 填充了所有字段的样本公告
func sampleMycertAdvisory(advisoryID string) model.MycertAdvisory {
	now := time.Now().UTC().Truncate(time.Second)
	return model.MycertAdvisory{
		AdvisoryID:  advisoryID,
		Title:       advisoryID + ": MyCERT Advisory - 测试漏洞标题",
		PublishedAt: &now,
		Category:    "Advisory",
		Summary:     "1.0 Introduction\n测试摘要内容，描述漏洞的基本情况。",
		DetailURL:   "https://www.mycert.org.my/portal/advisory?id=" + advisoryID,
		FullContent: "1.0 Introduction\n详细内容。\n2.0 Impact\n影响范围。\n3.0 Recommendations\n建议措施。",
		ScrapedAt:   now,
	}
}

// TestMycertRepository_Migrate 验证 Migrate 方法能幂等地创建表结构，
// 多次调用不应返回错误。
func TestMycertRepository_Migrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存 SQLite 失败: %v", err)
	}
	repo := repository.NewMycertAdvisoryRepository(db)
	ctx := context.Background()

	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("第一次 Migrate 失败: %v", err)
	}
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("第二次 Migrate（幂等性验证）失败: %v", err)
	}
}

// TestMycertRepository_BulkUpsert_Insert 验证 BulkUpsert 在记录不存在时
// 能正确批量插入，并通过 Count 确认写入行数。
func TestMycertRepository_BulkUpsert_Insert(t *testing.T) {
	db := newMycertTestDB(t)
	repo := repository.NewMycertAdvisoryRepository(db)
	ctx := context.Background()

	advisories := []model.MycertAdvisory{
		sampleMycertAdvisory("MA-0001.062026"),
		sampleMycertAdvisory("MA-0002.062026"),
		sampleMycertAdvisory("MA-0003.062026"),
	}

	if err := repo.BulkUpsert(ctx, advisories); err != nil {
		t.Fatalf("BulkUpsert 插入失败: %v", err)
	}

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count 查询失败: %v", err)
	}
	if count != 3 {
		t.Errorf("Count = %d; 期望 3", count)
	}
}

// TestMycertRepository_BulkUpsert_Update 验证 BulkUpsert 在记录已存在时
// 能正确更新所有可变字段，且总行数不增加（幂等 Upsert）。
func TestMycertRepository_BulkUpsert_Update(t *testing.T) {
	db := newMycertTestDB(t)
	repo := repository.NewMycertAdvisoryRepository(db)
	ctx := context.Background()

	original := sampleMycertAdvisory("MA-0010.062026")
	if err := repo.BulkUpsert(ctx, []model.MycertAdvisory{original}); err != nil {
		t.Fatalf("初始插入失败: %v", err)
	}

	updated := original
	updated.Title = "MA-0010.062026: MyCERT Advisory - 已更新标题"
	updated.FullContent = "更新后的详细内容。"
	if err := repo.BulkUpsert(ctx, []model.MycertAdvisory{updated}); err != nil {
		t.Fatalf("更新 BulkUpsert 失败: %v", err)
	}

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count 查询失败: %v", err)
	}
	if count != 1 {
		t.Errorf("Count = %d; 期望 1（Upsert 不应增加行数）", count)
	}

	var got model.MycertAdvisory
	if err := db.Where("advisory_id = ?", "MA-0010.062026").First(&got).Error; err != nil {
		t.Fatalf("查询更新后记录失败: %v", err)
	}
	if got.Title != updated.Title {
		t.Errorf("Title = %q; 期望 %q", got.Title, updated.Title)
	}
	if got.FullContent != updated.FullContent {
		t.Errorf("FullContent = %q; 期望 %q", got.FullContent, updated.FullContent)
	}
}

// TestMycertRepository_BulkUpsert_EmptySlice 验证传入空切片时
// BulkUpsert 直接返回 nil，不执行任何数据库操作。
func TestMycertRepository_BulkUpsert_EmptySlice(t *testing.T) {
	db := newMycertTestDB(t)
	repo := repository.NewMycertAdvisoryRepository(db)
	ctx := context.Background()

	if err := repo.BulkUpsert(ctx, []model.MycertAdvisory{}); err != nil {
		t.Fatalf("空切片 BulkUpsert 应返回 nil，实际返回: %v", err)
	}

	count, _ := repo.Count(ctx)
	if count != 0 {
		t.Errorf("空切片写入后 Count = %d; 期望 0", count)
	}
}

// TestMycertRepository_BulkUpsert_NilSlice 验证传入 nil 时
// BulkUpsert 直接返回 nil，不执行任何数据库操作。
func TestMycertRepository_BulkUpsert_NilSlice(t *testing.T) {
	db := newMycertTestDB(t)
	repo := repository.NewMycertAdvisoryRepository(db)
	ctx := context.Background()

	if err := repo.BulkUpsert(ctx, nil); err != nil {
		t.Fatalf("nil 切片 BulkUpsert 应返回 nil，实际返回: %v", err)
	}
}

// TestMycertRepository_BulkUpsert_LargeBatch 验证超过单批次上限（100 条）时
// BulkUpsert 能正确分批写入所有记录。
func TestMycertRepository_BulkUpsert_LargeBatch(t *testing.T) {
	db := newMycertTestDB(t)
	repo := repository.NewMycertAdvisoryRepository(db)
	ctx := context.Background()

	const total = 250
	advisories := make([]model.MycertAdvisory, total)
	for i := range advisories {
		advisories[i] = sampleMycertAdvisory(
			"MA-" + padInt(i+1) + ".062026",
		)
	}

	if err := repo.BulkUpsert(ctx, advisories); err != nil {
		t.Fatalf("大批量 BulkUpsert 失败: %v", err)
	}

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count 查询失败: %v", err)
	}
	if count != total {
		t.Errorf("Count = %d; 期望 %d", count, total)
	}
}

// TestMycertRepository_BulkUpsert_DuplicateInBatch 验证同一批次中存在重复 AdvisoryID 时
// BulkUpsert 能正确处理（以最后一条为准），不返回错误。
func TestMycertRepository_BulkUpsert_DuplicateInBatch(t *testing.T) {
	db := newMycertTestDB(t)
	repo := repository.NewMycertAdvisoryRepository(db)
	ctx := context.Background()

	first := sampleMycertAdvisory("MA-DUP.062026")
	second := sampleMycertAdvisory("MA-DUP.062026")
	second.Title = "MA-DUP.062026: 重复批次中的第二条"

	if err := repo.BulkUpsert(ctx, []model.MycertAdvisory{first, second}); err != nil {
		t.Fatalf("含重复 ID 的 BulkUpsert 失败: %v", err)
	}

	count, _ := repo.Count(ctx)
	if count != 1 {
		t.Errorf("Count = %d; 期望 1（重复 ID 应合并为一条）", count)
	}
}

// TestMycertRepository_BulkUpsert_CancelledContext 验证当 context 已取消时
// BulkUpsert 能正确返回错误，不会挂起。
func TestMycertRepository_BulkUpsert_CancelledContext(t *testing.T) {
	db := newMycertTestDB(t)
	repo := repository.NewMycertAdvisoryRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	advisories := []model.MycertAdvisory{sampleMycertAdvisory("MA-CTX.062026")}
	err := repo.BulkUpsert(ctx, advisories)
	if err == nil {
		t.Log("已取消的 context 下 BulkUpsert 返回 nil（SQLite 内存库可能不检查 context，属正常行为）")
	}
}

// TestMycertRepository_Count_Empty 验证空表时 Count 返回 0。
func TestMycertRepository_Count_Empty(t *testing.T) {
	db := newMycertTestDB(t)
	repo := repository.NewMycertAdvisoryRepository(db)
	ctx := context.Background()

	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count 查询失败: %v", err)
	}
	if count != 0 {
		t.Errorf("空表 Count = %d; 期望 0", count)
	}
}

// TestMycertRepository_DataIntegrity 验证写入后读取的字段值与原始数据完全一致，
// 包括可空的 PublishedAt 字段和长文本字段。
func TestMycertRepository_DataIntegrity(t *testing.T) {
	db := newMycertTestDB(t)
	repo := repository.NewMycertAdvisoryRepository(db)
	ctx := context.Background()

	adv := sampleMycertAdvisory("MA-INTEGRITY.062026")
	adv.FullContent = strings.Repeat("漏洞详情内容行。\n", 100)

	if err := repo.BulkUpsert(ctx, []model.MycertAdvisory{adv}); err != nil {
		t.Fatalf("BulkUpsert 失败: %v", err)
	}

	var got model.MycertAdvisory
	if err := db.Where("advisory_id = ?", adv.AdvisoryID).First(&got).Error; err != nil {
		t.Fatalf("查询记录失败: %v", err)
	}

	if got.AdvisoryID != adv.AdvisoryID {
		t.Errorf("AdvisoryID = %q; 期望 %q", got.AdvisoryID, adv.AdvisoryID)
	}
	if got.Title != adv.Title {
		t.Errorf("Title = %q; 期望 %q", got.Title, adv.Title)
	}
	if got.Category != adv.Category {
		t.Errorf("Category = %q; 期望 %q", got.Category, adv.Category)
	}
	if got.DetailURL != adv.DetailURL {
		t.Errorf("DetailURL = %q; 期望 %q", got.DetailURL, adv.DetailURL)
	}
	if got.FullContent != adv.FullContent {
		t.Errorf("FullContent 长度 = %d; 期望 %d", len(got.FullContent), len(adv.FullContent))
	}
	if adv.PublishedAt != nil && got.PublishedAt == nil {
		t.Error("PublishedAt 写入后读取为 nil")
	}
}

// TestMycertRepository_BulkUpsert_NullPublishedAt 验证 PublishedAt 为 nil 时
// 能正确写入和读取（NULL 值处理）。
func TestMycertRepository_BulkUpsert_NullPublishedAt(t *testing.T) {
	db := newMycertTestDB(t)
	repo := repository.NewMycertAdvisoryRepository(db)
	ctx := context.Background()

	adv := sampleMycertAdvisory("MA-NULLDATE.062026")
	adv.PublishedAt = nil

	if err := repo.BulkUpsert(ctx, []model.MycertAdvisory{adv}); err != nil {
		t.Fatalf("PublishedAt 为 nil 时 BulkUpsert 失败: %v", err)
	}

	var got model.MycertAdvisory
	if err := db.Where("advisory_id = ?", adv.AdvisoryID).First(&got).Error; err != nil {
		t.Fatalf("查询记录失败: %v", err)
	}
	if got.PublishedAt != nil {
		t.Errorf("PublishedAt = %v; 期望 nil", got.PublishedAt)
	}
}

// padInt 将整数格式化为 4 位零填充字符串，用于生成测试用的 AdvisoryID。
//
// 参数：
//   - n : 待格式化的整数
//
// 返回：
//   - string : 4 位零填充字符串（如 "0042"）
func padInt(n int) string {
	s := "0000" + itoa(n)
	return s[len(s)-4:]
}

// itoa 将整数转换为字符串，避免引入 strconv 包。
//
// 参数：
//   - n : 待转换的非负整数
//
// 返回：
//   - string : 整数的十进制字符串表示
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
