package services

import (
	"context"
	"fmt"
	"nezha_cyber_mcp/internal/utilities"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dsql/auth"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// DatabaseType 是数据库驱动类型的枚举。
type DatabaseType int

const (
	PostgreSQL       DatabaseType = iota // PostgreSQL（默认，推荐生产环境使用）
	MySQL                                // MySQL / MariaDB
	SQLServer                            // Microsoft SQL Server
	Oracle                               // Oracle Database（暂未实现）
	QuestDB                              // QuestDB 时序数据库（暂未实现）
	AmazonAuroraDSQL                     // Amazon Aurora DSQL 分布式 SQL 数据库（暂未实现）
)

// truncateBody 将原始 HTTP 响应体截断为最多 maxBytes 个字节，
// 并去除首尾空白，防止将大型 JSON 或 HTML 错误页面原样写入日志或错误消息。
// 若截断发生，在末尾追加 "…[truncated]" 标记。
func truncateBody(body []byte, maxBytes int) string {
	s := strings.TrimSpace(string(body))
	if len(s) <= maxBytes {
		return s
	}
	return s[:maxBytes] + "…[truncated]"
}

// String 返回 DatabaseType 的可读名称，用于日志输出。
func (t DatabaseType) String() string {
	switch t {
	case PostgreSQL:
		return "PostgreSQL"
	case MySQL:
		return "MySQL"
	case SQLServer:
		return "SQLServer"
	case Oracle:
		return "Oracle"
	case QuestDB:
		return "QuestDB"
	case AmazonAuroraDSQL:
		return "AmazonAuroraDSQL"
	default:
		return "Unknown"
	}
}

// DatabaseConfiguration 保存建立数据库连接所需的全部参数。
//
// 字段说明：
//   - Type            : 数据库驱动类型，见 DatabaseType 枚举
//   - Host            : 数据库服务器主机名或 IP 地址
//   - Port            : 数据库监听端口（PostgreSQL 默认 5432，MySQL 默认 3306）
//   - User            : 数据库登录用户名
//   - Password        : 数据库登录密码（AmazonAuroraDSQL 模式下由 InitDatabase 动态填充，无需手动设置）
//   - DBName          : 目标数据库名称
//   - SSLMode         : TLS/SSL 模式；nil 时使用驱动默认值（PostgreSQL 为 "prefer"）
//   - AWSRegion       : AWS 区域标识符，仅 AmazonAuroraDSQL 类型使用；空字符串时回退到 AWS_REGION 环境变量
//   - MaxOpenConns    : 连接池最大打开连接数；0 表示使用驱动默认值
//   - MaxIdleConns    : 连接池最大空闲连接数；0 表示使用驱动默认值
//   - ConnMaxLifetime : 连接最大存活时间；0 表示不限制
type DatabaseConfiguration struct {
	Type      DatabaseType
	Host      string
	Port      string
	User      string
	Password  string
	DBName    string
	SSLMode   *string // nil => 使用驱动默认值 (libpq "prefer")
	AWSRegion string  // 仅 AmazonAuroraDSQL 使用；空字符串时从 AWS_REGION 环境变量读取

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

// ValidateConfiguration 校验 DatabaseConfiguration 与当前运行时环境的兼容性。
//
// 防护规则：
//   - 本地开发环境（utilities.IsLocalMode() == true）中，禁止使用 AmazonAuroraDSQL 类型。
//     若检测到此冲突，立即返回错误，阻止连接建立，防止开发者意外连接生产数据库。
//   - 其他类型在本地模式下均允许使用。
//
// 参数：
//   - cfg : 待校验的数据库连接配置
//
// 返回：
//   - error : 配置与环境不兼容时返回描述性错误；兼容时返回 nil
func ValidateConfiguration(cfg DatabaseConfiguration) error {
	if utilities.IsLocalMode() && cfg.Type == AmazonAuroraDSQL {
		return fmt.Errorf(
			"配置错误：本地开发环境中禁止使用 AmazonAuroraDSQL 驱动 — " +
				"本地开发必须使用 PostgreSQL（Type=0）。" +
				"若需连接 AWS 生产数据库，请在云运行时环境中部署后再使用",
		)
	}
	return nil
}

// buildDSN 根据 DatabaseConfiguration 构造对应数据库驱动的 DSN 并返回 gorm.Dialector。
// 支持 PostgreSQL、MySQL、SQL Server 和 Amazon Aurora DSQL。
//
// 注意：AmazonAuroraDSQL 类型的 DSN 中密码字段由调用方（InitDatabase）在运行时动态注入，
// 此函数接收已填充 Password 字段的 cfg，直接构造完整 DSN。
//
// 参数：
//   - cfg : 数据库连接配置（AmazonAuroraDSQL 时 Password 字段应已包含 IAM token）
//
// 返回：
//   - gorm.Dialector : 对应驱动的方言实例
//   - error          : 不支持的数据库类型时返回错误
func buildDSN(cfg DatabaseConfiguration) (gorm.Dialector, error) {
	switch cfg.Type {
	case PostgreSQL:
		sslMode := "disable"
		if cfg.SSLMode != nil {
			sslMode = *cfg.SSLMode
		}
		var b strings.Builder
		fmt.Fprintf(&b, "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, sslMode)
		return postgres.Open(b.String()), nil

	case MySQL:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		return mysql.Open(dsn), nil

	case SQLServer:
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		return sqlserver.Open(dsn), nil

	case AmazonAuroraDSQL:
		// Aurora DSQL 底层走 PostgreSQL 协议，但强制要求 TLS（sslmode=require）。
		// 密码字段此时已由 InitDatabase 填充为 IAM SigV4 presigned token。
		// Aurora DSQL 不使用标准端口，默认端口为 5432，由调用方在 cfg.Port 中指定。
		port := cfg.Port
		if port == "" {
			port = "5432"
		}
		var b strings.Builder
		fmt.Fprintf(&b, "host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
			cfg.Host, cfg.User, cfg.Password, cfg.DBName, port)
		return postgres.Open(b.String()), nil

	default:
		return nil, fmt.Errorf("数据库类型 %d 不受支持或尚未实现", cfg.Type)
	}
}

// generateDSQLToken 使用 AWS SDK 为 Aurora DSQL 集群生成 IAM SigV4 presigned 认证 token。
// token 有效期为 15 分钟，每次建立连接前必须重新生成。
//
// 参数：
//   - ctx             : 请求上下文
//   - clusterEndpoint : Aurora DSQL 集群端点（即 cfg.Host）
//   - region          : AWS 区域标识符（如 "us-east-1"）
//
// 返回：
//   - string : 可直接用作 PostgreSQL 密码的 presigned token
//   - error  : AWS 配置加载失败或 token 生成失败时返回错误
func generateDSQLToken(ctx context.Context, clusterEndpoint, region string) (string, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("加载 AWS 默认配置失败: %w", err)
	}

	token, err := auth.GenerateDBConnectAdminAuthToken(
		ctx,
		clusterEndpoint,
		region,
		awsCfg.Credentials,
	)
	if err != nil {
		return "", fmt.Errorf("生成 Aurora DSQL IAM token 失败: %w", err)
	}

	return token, nil
}

// InitDatabase 根据配置初始化数据库连接，配置连接池参数，并通过 Ping 验证连通性。
// 若 Ping 失败，会自动关闭已建立的连接并返回错误。
//
// 对于 AmazonAuroraDSQL 类型，InitDatabase 会在建立连接前自动生成 IAM SigV4 presigned token
// 并将其注入 cfg.Password 字段，调用方无需手动提供密码。
// token 有效期为 15 分钟，因此每次调用 InitDatabase 都会重新生成。
//
// 参数：
//   - ctx : 请求上下文，用于 token 生成和 Ping 超时控制
//   - cfg : 数据库连接配置
//
// 返回：
//   - *Database : 初始化成功的数据库封装实例
//   - error     : token 生成、连接、配置或 Ping 失败时返回包装后的错误
func InitDatabase(ctx context.Context, cfg DatabaseConfiguration) (*Database, error) {
	start := time.Now()
	utilities.LogStart("Database", "InitDatabase")

	if err := ValidateConfiguration(cfg); err != nil {
		utilities.LogError("Database", "InitDatabase", err, time.Since(start), "step=ValidateConfiguration")
		return nil, err
	}

	if cfg.Type == AmazonAuroraDSQL {
		region := cfg.AWSRegion
		if region == "" {
			region = utilities.AWSRegion("us-east-1")
		}
		utilities.LogProgress("Database", "InitDatabase",
			"正在生成 Aurora DSQL IAM token",
			fmt.Sprintf("endpoint=%s", cfg.Host),
			fmt.Sprintf("region=%s", region),
		)
		token, err := generateDSQLToken(ctx, cfg.Host, region)
		if err != nil {
			utilities.LogError("Database", "InitDatabase", err, time.Since(start), "step=GenerateDSQLToken")
			return nil, fmt.Errorf("生成 Aurora DSQL IAM token 失败: %w", err)
		}
		cfg.Password = token
	}

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

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		utilities.LogError("Database", "InitDatabase", err, time.Since(start), "step=Ping")
		return nil, fmt.Errorf("ping 数据库失败: %w", err)
	}

	utilities.LogSuccess("Database", "InitDatabase", time.Since(start),
		fmt.Sprintf("driver=%s", cfg.Type.String()),
		fmt.Sprintf("host=%s", cfg.Host),
		fmt.Sprintf("db=%s", cfg.DBName),
	)
	return &Database{db: db}, nil
}
