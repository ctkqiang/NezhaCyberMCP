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

type DatabaseType int

const (
	PostgreSQL DatabaseType = iota
	MySQL
	SQLServer
	Oracle
	QuestDB
)

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

type Repository[T any] interface {
	GetAll(ctx context.Context) ([]T, error)
	Get(ctx context.Context, id int) (*T, error)
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id int) error
}

type Database struct{ db *gorm.DB }

func (d *Database) DB() *gorm.DB { return d.db }

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("获取 sql.DB 句柄失败: %w", err)
	}
	return sqlDB.Close()
}

func buildDSN(cfg DatabaseConfiguration) (gorm.Dialector, error) {
	switch cfg.Type {
	case PostgreSQL:
		var b strings.Builder
		fmt.Fprintf(&b, "host=%s user=%s password=%s dbname=%s port=%s",
			cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port)
		if cfg.SSLMode != nil {
			fmt.Fprintf(&b, " sslmode=%s", *cfg.SSLMode)
		}
		return postgres.Open(b.String()), nil

	case MySQL:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		return mysql.Open(dsn), nil

	case SQLServer:
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		return sqlserver.Open(dsn), nil

	default:
		return nil, fmt.Errorf("数据库类型 %d 不受支持或尚未实现", cfg.Type)
	}
}

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
		return nil, fmt.Errorf("ping 数据库失败: %w", err)
	}

	return &Database{db: db}, nil
}
