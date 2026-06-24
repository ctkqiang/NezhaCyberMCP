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
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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
		"user="+utilities.Mask(dbCfg.User),
	)

	db, err := services.InitDatabase(ctx, dbCfg)
	if err != nil {
		utilities.LogError("Main", "Startup", err, 0)
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			utilities.Warn("关闭数据库连接失败: %v", closeErr)
		}
	}()

	repo := repository.NewGithubAdvisoryRepository(db.DB())

	if err := repo.Migrate(ctx); err != nil {
		utilities.LogError("Main", "Migrate", err, 0)
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "scrape error: %v\n", err)
		os.Exit(1)
	}

	// 查询数据库总行数，输出本次运行汇总信息。
	count, _ := repo.Count(ctx)
	utilities.LogSuccess("Main", "ScrapeAndPersist", time.Since(start),
		fmt.Sprintf("scraped_this_run=%d", total),
		fmt.Sprintf("total_in_db=%d", count))
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

func GetGithubAdvisoryService() error {
	return nil
}
