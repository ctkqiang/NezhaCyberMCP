package repository

import (
	"context"
	"fmt"
	"nezha_cyber_mcp/internal/model"
	"nezha_cyber_mcp/internal/utilities"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// mycertComponent 是本 Repository 在日志中使用的组件名称标识。
const mycertComponent = "MycertAdvisoryRepository"

// MycertAdvisoryRepository 封装了对 mycert_advisories 表的所有持久化操作。
// 通过 *gorm.DB 与数据库交互，所有方法均支持 context 传递以实现超时与取消控制。
type MycertAdvisoryRepository struct {
	db *gorm.DB
}

// NewMycertAdvisoryRepository 构造一个绑定到指定 *gorm.DB 实例的 Repository。
//
// 参数：
//   - db : 已初始化的 GORM 数据库连接
//
// 返回：
//   - *MycertAdvisoryRepository
func NewMycertAdvisoryRepository(db *gorm.DB) *MycertAdvisoryRepository {
	return &MycertAdvisoryRepository{db: db}
}

// Migrate 执行 AutoMigrate，确保 mycert_advisories 表存在且结构与模型一致。
// 该方法幂等，可安全地多次调用。
//
// 参数：
//   - ctx : 请求上下文
//
// 返回：
//   - error : 迁移失败时返回包装后的错误，成功时返回 nil
func (r *MycertAdvisoryRepository) Migrate(ctx context.Context) error {
	start := time.Now()
	utilities.LogStart(mycertComponent, "Migrate")

	if err := r.db.WithContext(ctx).AutoMigrate(&model.MycertAdvisory{}); err != nil {
		utilities.LogError(mycertComponent, "Migrate", err, time.Since(start))
		return fmt.Errorf("auto-migrate mycert_advisories: %w", err)
	}

	utilities.LogSuccess(mycertComponent, "Migrate", time.Since(start))
	return nil
}

// Upsert 插入或更新单条 MyCERT 公告记录。
// 以 advisory_id（主键）作为冲突检测列；冲突时更新所有可变字段，
// 确保数据库中始终保存最新版本的公告内容。
//
// 业务规则：
//   - 新记录：直接插入
//   - 已存在（advisory_id 冲突）：更新 title、published_at、category、
//     summary、detail_url、full_content、scraped_at
//
// 参数：
//   - ctx      : 请求上下文
//   - advisory : 待写入的公告指针，AdvisoryID 字段不可为空
//
// 返回：
//   - error : 写入失败时返回包装后的错误，成功时返回 nil
func (r *MycertAdvisoryRepository) Upsert(ctx context.Context, advisory *model.MycertAdvisory) error {
	start := time.Now()
	utilities.LogStart(mycertComponent, "Upsert")

	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "advisory_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"title", "published_at", "category",
				"summary", "detail_url", "full_content", "scraped_at",
			}),
		}).
		Create(advisory)

	if result.Error != nil {
		utilities.LogError(mycertComponent, "Upsert", result.Error, time.Since(start),
			"advisory_id="+advisory.AdvisoryID)
		return fmt.Errorf("upsert mycert advisory %s: %w", advisory.AdvisoryID, result.Error)
	}

	action := "inserted"
	if result.RowsAffected == 0 {
		action = "skipped(no-change)"
	} else if result.RowsAffected > 1 {
		action = "updated"
	}

	utilities.LogSuccess(mycertComponent, "Upsert", time.Since(start),
		"advisory_id="+advisory.AdvisoryID,
		"action="+action,
		fmt.Sprintf("rows_affected=%d", result.RowsAffected))
	return nil
}

// BulkUpsert 在单个数据库事务中批量插入或更新 MyCERT 公告记录。
// 以 advisory_id（主键）作为冲突检测列；冲突时更新所有可变字段。
// 每批最多处理 100 条，任意一条失败将回滚整个事务，保证原子性。
//
// 业务规则：
//   - 新记录：直接插入
//   - 已存在（advisory_id 冲突）：更新除主键外的所有字段
//
// 参数：
//   - ctx        : 请求上下文
//   - advisories : 待写入的公告切片，传入空切片或 nil 时直接返回 nil（无操作）
//
// 返回：
//   - error : 事务失败时返回包装后的错误，成功时返回 nil
func (r *MycertAdvisoryRepository) BulkUpsert(ctx context.Context, advisories []model.MycertAdvisory) error {
	if len(advisories) == 0 {
		return nil
	}

	start := time.Now()
	utilities.LogStart(mycertComponent, "BulkUpsert")

	total := len(advisories)
	var rowsAffected int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "advisory_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"title", "published_at", "category",
					"summary", "detail_url", "full_content", "scraped_at",
				}),
			}).
			CreateInBatches(advisories, 100)

		if result.Error != nil {
			return result.Error
		}

		rowsAffected = result.RowsAffected
		utilities.LogProgress(mycertComponent, "BulkUpsert",
			fmt.Sprintf("事务提交：input=%d rows_affected=%d", total, rowsAffected))
		return nil
	})
	if err != nil {
		utilities.LogError(mycertComponent, "BulkUpsert", err, time.Since(start),
			fmt.Sprintf("input=%d", total))
		return fmt.Errorf("bulk upsert %d mycert advisories: %w", total, err)
	}

	utilities.LogSuccess(mycertComponent, "BulkUpsert", time.Since(start),
		fmt.Sprintf("input=%d", total),
		fmt.Sprintf("rows_affected=%d", rowsAffected))
	return nil
}

// Count 返回 mycert_advisories 表中的记录总数。
//
// 参数：
//   - ctx : 请求上下文
//
// 返回：
//   - int64 : 记录总数
//   - error : 查询失败时返回错误
func (r *MycertAdvisoryRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.MycertAdvisory{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count mycert_advisories: %w", err)
	}
	return count, nil
}
