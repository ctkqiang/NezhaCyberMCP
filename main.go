package main

import (
	"context"
	"fmt"
	"nezha_cyber_mcp/internal/repository"
	"nezha_cyber_mcp/internal/services"
	"nezha_cyber_mcp/internal/utilities"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		utilities.Warn("未找到 .env 文件，使用系统环境变量: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := GetGithubAdvisoryService(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

// GetGithubAdvisoryService 初始化数据库连接、执行表迁移，
// 并启动 GitHub 安全公告的抓取与持久化流程。
// 所有子步骤均通过 ctx 支持超时与取消控制。
//
// 参数：
//   - ctx : 请求上下文，由 main 注入，携带信号取消能力
//
// 返回：
//   - error : 任意步骤失败时返回包装后的错误，成功时返回 nil
func GetGithubAdvisoryService(ctx context.Context) error {
	utilities.LogStart("Main", "Startup")

	dbCfg := services.DatabaseConfiguration{
		Type:            services.PostgreSQL,
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "nezha_cyber"),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	utilities.LogProgress("Main", "Startup", "正在连接数据库",
		"host="+dbCfg.Host,
		"port="+dbCfg.Port,
		"db="+dbCfg.DBName,
		"user="+utilities.Mask(dbCfg.User))

	db, err := services.InitDatabase(ctx, dbCfg)
	if err != nil {
		utilities.LogError("Main", "Startup", err, 0)
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			utilities.Warn("关闭数据库连接失败: %v", closeErr)
		}
	}()

	repo := repository.NewGithubAdvisoryRepository(db.DB())

	if err := repo.Migrate(ctx); err != nil {
		utilities.LogError("Main", "Migrate", err, 0)
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	scraperCfg := &services.AdvisoryScraperConfig{
		MaxPages:       10,
		RequestTimeout: 30 * time.Second,
		Parallelism:    2,
		RateLimit:      time.Second,
	}

	svc := services.NewGithubAdvisoryService(repo, scraperCfg)

	start := time.Now()
	total, err := svc.ScrapeAndPersist(ctx)
	if err != nil {
		utilities.LogError("Main", "ScrapeAndPersist", err, time.Since(start))
		return fmt.Errorf("抓取公告失败: %w", err)
	}

	count, _ := repo.Count(ctx)
	utilities.LogSuccess("Main", "ScrapeAndPersist", time.
		Since(start),
		fmt.Sprintf("scraped_this_run=%d", total),
		fmt.Sprintf("total_in_db=%d", count),
	)

	return nil
}

// getEnv 读取指定环境变量的值。
// 若环境变量未设置或为空字符串，则返回 fallback 默认值。
//
// 参数：
//   - key      : 环境变量名称
//   - fallback : 环境变量未设置时的默认值
//
// 返回：
//   - string : 环境变量的值或默认值
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
