package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// DatabaseType 是数据库驱动类型的枚举。
type DatabaseType int

const (
	PostgreSQL DatabaseType = iota // PostgreSQL（默认，推荐生产环境使用）
	MySQL                          // MySQL / MariaDB
	SQLServer                      // Microsoft SQL Server
	Oracle                         // Oracle Database（暂未实现）
	QuestDB                        // QuestDB 时序数据库（暂未实现）
)

// DatabaseConfiguration 保存建立数据库连接所需的全部参数。
//
// 字段说明：
//   - Type            : 数据库驱动类型，见 DatabaseType 枚举
//   - Host            : 数据库服务器主机名或 IP 地址
//   - Port            : 数据库监听端口（PostgreSQL 默认 5432，MySQL 默认 3306）
//   - User            : 数据库登录用户名
//   - Password        : 数据库登录密码
//   - DBName          : 目标数据库名称
//   - SSLMode         : TLS/SSL 模式；nil 时使用驱动默认值（PostgreSQL 为 "prefer"）
//   - MaxOpenConns    : 连接池最大打开连接数；0 表示使用驱动默认值
//   - MaxIdleConns    : 连接池最大空闲连接数；0 表示使用驱动默认值
//   - ConnMaxLifetime : 连接最大存活时间；0 表示不限制
type DatabaseConfiguration struct {
	Type     DatabaseType
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  *string // nil => 使用驱动默认值 (libpq "prefer")

	// 可选的连接池调优参数；零值将回退到合理的默认值。
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Repository 是通用 CRUD 操作的泛型接口，供各业务 Repository 实现。
// T 为具体的数据模型类型。
type Repository[T any] interface {
	// GetAll 返回表中所有记录。
	GetAll(ctx context.Context) ([]T, error)

	// Get 根据整型主键查询单条记录。
	Get(ctx context.Context, id int) (*T, error)

	// Create 插入一条新记录。
	Create(ctx context.Context, entity *T) error

	// Update 更新一条已有记录。
	Update(ctx context.Context, entity *T) error

	// Delete 根据整型主键删除一条记录。
	Delete(ctx context.Context, id int) error
}

// Database 封装了 GORM 数据库连接，提供统一的访问与关闭接口。
type Database struct{ db *gorm.DB }

// DB 返回底层的 *gorm.DB 实例，供 Repository 层直接使用。
func (d *Database) DB() *gorm.DB { return d.db }

// Close 关闭底层的 sql.DB 连接池，释放所有数据库连接。
// 应在程序退出前通过 defer 调用。
//
// 返回：
//   - error : 获取 sql.DB 句柄失败或关闭失败时返回错误
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("获取 sql.DB 句柄失败: %w", err)
	}
	return sqlDB.Close()
}

// buildDSN 根据 DatabaseConfiguration 构造对应数据库驱动的 DSN 并返回 gorm.Dialector。
// 目前支持 PostgreSQL、MySQL、SQL Server；其他类型返回错误。
//
// 参数：
//   - cfg : 数据库连接配置
//
// 返回：
//   - gorm.Dialector : 对应驱动的方言实例
//   - error          : 不支持的数据库类型时返回错误
func buildDSN(cfg DatabaseConfiguration) (gorm.Dialector, error) {
	switch cfg.Type {
	case PostgreSQL:
		// PostgreSQL DSN 使用键值对格式，SSLMode 可选。
		var b strings.Builder
		fmt.Fprintf(&b, "host=%s user=%s password=%s dbname=%s port=%s",
			cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port)
		if cfg.SSLMode != nil {
			fmt.Fprintf(&b, " sslmode=%s", *cfg.SSLMode)
		}
		return postgres.Open(b.String()), nil

	case MySQL:
		// MySQL DSN 使用 charset=utf8mb4 确保完整 Unicode 支持，parseTime=True 自动解析时间类型。
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		return mysql.Open(dsn), nil

	case SQLServer:
		// SQL Server DSN 使用标准 URL 格式。
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		return sqlserver.Open(dsn), nil

	default:
		return nil, fmt.Errorf("数据库类型 %d 不受支持或尚未实现", cfg.Type)
	}
}

// InitDatabase 根据配置初始化数据库连接，配置连接池参数，并通过 Ping 验证连通性。
// 若 Ping 失败，会自动关闭已建立的连接并返回错误。
//
// 参数：
//   - ctx : 请求上下文，用于 Ping 超时控制
//   - cfg : 数据库连接配置
//
// 返回：
//   - *Database : 初始化成功的数据库封装实例
//   - error     : 连接、配置或 Ping 失败时返回包装后的错误
func InitDatabase(ctx context.Context, cfg DatabaseConfiguration) (*Database, error) {
	dialector, err := buildDSN(cfg)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取 sql.DB 句柄失败: %w", err)
	}

	// 仅在配置值大于零时覆盖连接池参数，否则保留驱动默认值。
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	// 通过 Ping 验证数据库实际可达，失败时关闭连接并返回错误。
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping 数据库失败: %w", err)
	}

	return &Database{db: db}, nil
}
