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

// circlComponent 是本 Repository 在日志中使用的组件名称标识。
const circlComponent = "CirclCVERepository"

// CirclCVERepository 封装了对 circl_cves 表的所有持久化操作。
// 通过 *gorm.DB 与数据库交互，所有方法均支持 context 传递以实现超时与取消控制。
type CirclCVERepository struct {
	db *gorm.DB
}

// NewCirclCVERepository 构造一个绑定到指定 *gorm.DB 实例的 Repository。
//
// 参数：
//   - db : 已初始化的 GORM 数据库连接
//
// 返回：
//   - *CirclCVERepository
func NewCirclCVERepository(db *gorm.DB) *CirclCVERepository {
	return &CirclCVERepository{db: db}
}

// Migrate 执行 AutoMigrate，确保 circl_cves 表存在且结构与模型一致。
// 该方法幂等，可安全地多次调用。
//
// 参数：
//   - ctx : 请求上下文，用于超时与取消控制
//
// 返回：
//   - error : 迁移失败时返回包装后的错误，成功时返回 nil
func (r *CirclCVERepository) Migrate(ctx context.Context) error {
	start := time.Now()
	utilities.LogStart(circlComponent, "Migrate")

	if err := r.db.WithContext(ctx).AutoMigrate(&model.CirclCVE{}); err != nil {
		utilities.LogError(circlComponent, "Migrate", err, time.Since(start))
		return fmt.Errorf("auto-migrate circl_cves: %w", err)
	}

	utilities.LogSuccess(circlComponent, "Migrate", time.Since(start))
	return nil
}

// BulkUpsert 在单个数据库事务中批量插入或更新 CIRCL CVE 记录。
// 以 CVEID（主键）作为冲突判断依据，冲突时更新所有可变字段。
// 每批最多处理 100 条，任意一条失败将回滚整个事务，保证原子性。
//
// 参数：
//   - ctx   : 请求上下文
//   - items : 待写入的 CVE 切片，传入空切片或 nil 时直接返回 nil（无操作）
//
// 返回：
//   - error : 事务失败时返回包装后的错误，成功时返回 nil
func (r *CirclCVERepository) BulkUpsert(ctx context.Context, items []model.CirclCVE) error {
	if len(items) == 0 {
		return nil
	}

	start := time.Now()
	utilities.LogStart(circlComponent, "BulkUpsert")

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Clauses(clause.OnConflict{
				UpdateAll: true,
			}).
			CreateInBatches(items, 100)
		return result.Error
	})

	if err != nil {
		utilities.LogError(circlComponent, "BulkUpsert", err, time.Since(start),
			fmt.Sprintf("rows=%d", len(items)))
		return fmt.Errorf("bulk upsert circl_cves: %w", err)
	}

	utilities.LogSuccess(circlComponent, "BulkUpsert", time.Since(start),
		fmt.Sprintf("rows=%d", len(items)))
	return nil
}

// Count 返回 circl_cves 表中的记录总数。
//
// 参数：
//   - ctx : 请求上下文
//
// 返回：
//   - int64 : 记录总数
//   - error : 查询失败时返回包装后的错误，成功时返回 nil
func (r *CirclCVERepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.CirclCVE{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count circl_cves: %w", err)
	}
	return count, nil
}
