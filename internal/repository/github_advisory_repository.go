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

// component 是本包在日志中使用的组件名称标识。
const component = "GithubAdvisoryRepository"

// GithubAdvisoryRepository 封装了对 github_advisories 表的所有持久化操作。
// 通过 *gorm.DB 与数据库交互，所有方法均支持 context 传递以实现超时与取消控制。
type GithubAdvisoryRepository struct {
	db *gorm.DB
}

// NewGithubAdvisoryRepository 构造一个绑定到指定 *gorm.DB 实例的 Repository。
//
// 参数：
//   - db : 已初始化的 GORM 数据库连接
//
// 返回：
//   - *GithubAdvisoryRepository
func NewGithubAdvisoryRepository(db *gorm.DB) *GithubAdvisoryRepository {
	return &GithubAdvisoryRepository{db: db}
}

// Migrate 执行 AutoMigrate，确保 github_advisories 表存在且结构与模型一致。
// 该方法幂等，可安全地多次调用。
//
// 参数：
//   - ctx : 请求上下文，用于超时与取消控制
//
// 返回：
//   - error : 迁移失败时返回包装后的错误，成功时返回 nil
func (r *GithubAdvisoryRepository) Migrate(ctx context.Context) error {
	start := time.Now()
	utilities.LogStart(component, "Migrate")

	if err := r.db.WithContext(ctx).AutoMigrate(&model.GithubAdvisory{}); err != nil {
		utilities.LogError(component, "Migrate", err, time.Since(start))
		return fmt.Errorf("auto-migrate github_advisories: %w", err)
	}

	utilities.LogSuccess(component, "Migrate", time.Since(start))
	return nil
}

// Upsert 插入或更新单条安全公告记录。
// 以 ghsa_id（主键）作为冲突判断依据；冲突时更新除主键外的所有可变字段，
// 确保数据库中始终保存最新版本的公告内容。
//
// 业务规则：
//   - 新记录：直接插入
//   - 已存在（ghsa_id 冲突）：更新 cve_id、url、html_url、summary、description、
//     type、severity、published_at、updated_at、withdrawn_at、vulnerabilities、references
//
// 参数：
//   - ctx      : 请求上下文
//   - advisory : 待写入的公告指针，GHSAID 字段不可为空
//
// 返回：
//   - error : 写入失败时返回包装后的错误，成功时返回 nil
func (r *GithubAdvisoryRepository) Upsert(ctx context.Context, advisory *model.GithubAdvisory) error {
	start := time.Now()
	utilities.LogStart(component, "Upsert")

	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			// 明确指定冲突检测列为主键 ghsa_id，避免依赖数据库隐式推断。
			Columns: []clause.Column{{Name: "ghsa_id"}},
			// 冲突时更新所有非主键字段，保持数据与上游 API 同步。
			DoUpdates: clause.AssignmentColumns([]string{
				"cve_id", "url", "html_url", "summary", "description",
				"type", "severity", "published_at", "updated_at",
				"withdrawn_at", "vulnerabilities", "references",
			}),
		}).
		Create(advisory)

	if result.Error != nil {
		utilities.LogError(component, "Upsert", result.Error, time.Since(start),
			"ghsa_id="+advisory.GHSAID)
		return fmt.Errorf("upsert advisory %s: %w", advisory.GHSAID, result.Error)
	}

	action := "inserted"
	if result.RowsAffected == 0 {
		action = "skipped(no-change)"
	} else if result.RowsAffected > 1 {
		action = "updated"
	}

	utilities.LogSuccess(component, "Upsert", time.Since(start),
		"ghsa_id="+advisory.GHSAID,
		"action="+action,
		fmt.Sprintf("rows_affected=%d", result.RowsAffected))
	return nil
}

// BulkUpsert 在单个数据库事务中批量插入或更新公告记录。
// 以 ghsa_id（主键）作为冲突检测列；冲突时更新所有可变字段。
// 每批最多处理 100 条，任意一条失败将回滚整个事务，保证原子性。
//
// 业务规则：
//   - 新记录：直接插入
//   - 已存在（ghsa_id 冲突）：更新除主键外的所有字段
//
// 参数：
//   - ctx        : 请求上下文
//   - advisories : 待写入的公告切片，传入空切片或 nil 时直接返回 nil（无操作）
//
// 返回：
//   - error : 事务失败时返回包装后的错误，成功时返回 nil
func (r *GithubAdvisoryRepository) BulkUpsert(ctx context.Context, advisories []model.GithubAdvisory) error {
	if len(advisories) == 0 {
		return nil
	}

	start := time.Now()
	utilities.LogStart(component, "BulkUpsert")

	total := len(advisories)
	var rowsAffected int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Clauses(clause.OnConflict{
				// 明确指定冲突检测列为主键 ghsa_id。
				Columns: []clause.Column{{Name: "ghsa_id"}},
				// 冲突时更新所有非主键字段，保持数据与上游 API 同步。
				DoUpdates: clause.AssignmentColumns([]string{
					"cve_id", "url", "html_url", "summary", "description",
					"type", "severity", "published_at", "updated_at",
					"withdrawn_at", "vulnerabilities", "references",
				}),
			}).
			CreateInBatches(advisories, 100)

		if result.Error != nil {
			return result.Error
		}

		rowsAffected = result.RowsAffected
		utilities.LogProgress(component, "BulkUpsert",
			fmt.Sprintf("事务提交：input=%d rows_affected=%d", total, rowsAffected))
		return nil
	})

	if err != nil {
		utilities.LogError(component, "BulkUpsert", err, time.Since(start),
			fmt.Sprintf("input=%d", total))
		return fmt.Errorf("bulk upsert %d advisories: %w", total, err)
	}

	utilities.LogSuccess(component, "BulkUpsert", time.Since(start),
		fmt.Sprintf("input=%d", total),
		fmt.Sprintf("rows_affected=%d", rowsAffected))
	return nil
}

// GetByGHSAID 根据 GHSA 标识符查询单条公告记录。
// 记录不存在时返回 (nil, nil)，而非错误，调用方需自行判断。
//
// 参数：
//   - ctx    : 请求上下文
//   - ghsaID : 目标公告的 GHSA 标识符（如 GHSA-xxxx-xxxx-xxxx）
//
// 返回：
//   - *model.GithubAdvisory : 查询到的记录指针；记录不存在时为 nil
//   - error                 : 数据库查询异常时返回错误，记录不存在不视为错误
func (r *GithubAdvisoryRepository) GetByGHSAID(ctx context.Context, ghsaID string) (*model.GithubAdvisory, error) {
	start := time.Now()

	var advisory model.GithubAdvisory
	result := r.db.WithContext(ctx).
		Where("ghsa_id = ?", ghsaID).
		First(&advisory)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 记录不存在不视为错误，返回 nil
			return nil, nil
		}
		utilities.LogError(component, "GetByGHSAID", result.Error, time.Since(start),
			"ghsa_id="+ghsaID)
		return nil, fmt.Errorf("get advisory %s: %w", ghsaID, result.Error)
	}

	utilities.LogSuccess(component, "GetByGHSAID", time.Since(start), "ghsa_id="+ghsaID)
	return &advisory, nil
}

// Count 返回数据库中 github_advisories 表的总行数。
//
// 参数：
//   - ctx : 请求上下文
//
// 返回：
//   - int64 : 总行数
//   - error : 查询失败时返回错误
func (r *GithubAdvisoryRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.GithubAdvisory{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count advisories: %w", err)
	}
	return count, nil
}
