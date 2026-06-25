package repository

import (
	"context"
	"fmt"
	"time"

	"nezha_cyber_mcp/internal/model"
	"nezha_cyber_mcp/internal/utilities"

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

// BulkUpsert 在单个数据库事务中批量插入或更新 MyCERT 公告记录。
// 以 AdvisoryID（主键）作为冲突判断依据，冲突时更新所有可变字段。
// 每批最多处理 100 条，任意一条失败将回滚整个事务，保证原子性。
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

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Clauses(clause.OnConflict{
				UpdateAll: true,
			}).
			CreateInBatches(advisories, 100)

		if result.Error != nil {
			return result.Error
		}

		utilities.LogProgress(mycertComponent, "BulkUpsert",
			fmt.Sprintf("已提交 %d 行", result.RowsAffected))
		return nil
	})

	if err != nil {
		utilities.LogError(mycertComponent, "BulkUpsert", err, time.Since(start),
			fmt.Sprintf("batch_size=%d", len(advisories)))
		return fmt.Errorf("bulk upsert %d mycert advisories: %w", len(advisories), err)
	}

	utilities.LogSuccess(mycertComponent, "BulkUpsert", time.Since(start),
		fmt.Sprintf("total=%d", len(advisories)))
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
	result := r.db.WithContext(ctx).Model(&model.MycertAdvisory{}).Count(&count)
	return count, result.Error
}
