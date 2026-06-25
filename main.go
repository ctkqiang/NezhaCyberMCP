package main

import (
	"context"
	"fmt"
	"nezha_cyber_mcp/internal/job"
	"nezha_cyber_mcp/internal/services"
	"nezha_cyber_mcp/internal/utilities"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

const (
	// IsGithubAdvisoryTurnOn 控制是否启用 GitHub Advisory 数据源同步。
	IsGithubAdvisoryTurnOn = false

	// IsMycertAdvisoryTurnOn 控制是否启用 MyCERT 公告数据源同步。
	IsMycertAdvisoryTurnOn = true

	// IsCirclCVETurnOn 控制是否启用 CIRCL CVE Search API 数据源同步。
	IsCirclCVETurnOn = true
)

func main() {
	if err := godotenv.Load(); err != nil {
		utilities.Warn("未找到 .env 文件，使用系统环境变量: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

// run 初始化所有依赖，启动定时任务调度器，并阻塞直到收到退出信号。
//
// 参数：
//   - ctx : 根上下文，携带信号取消能力
//
// 返回：
//   - error : 初始化失败时返回错误
func run(ctx context.Context) error {
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

	scraperCfg := &services.AdvisoryScraperConfig{
		MaxPages:       0,
		RequestTimeout: 30 * time.Second,
		PerPage:        100,
		RetryMax:       5,
		RetryBackoff:   2 * time.Second,
		Token:          getEnv("GITHUB_TOKEN", ""),
	}

	timezone := getEnv("DB_TIMEZONE", "Asia/Shanghai")

	mycertCfg := &services.MycertScraperConfig{
		MaxPages:       0,
		RequestTimeout: 30 * time.Second,
		FetchDetail:    true,
	}

	// circlCfg 的 APIToken 从环境变量 CIRCL_API_TOKEN 读取，不在此处硬编码。
	// CIRCL 公共 API 无需强制认证，Token 为空时以匿名方式访问。
	circlCfg := &services.CirclScraperConfig{
		MaxPages:       0,
		RequestTimeout: 30 * time.Second,
		PerPage:        100,
		RetryMax:       5,
		RetryBackoff:   2 * time.Second,
		RateLimit:      500 * time.Millisecond,
	}

	advisoryJob, err := job.NewAdvisoryJob(dbCfg, scraperCfg, mycertCfg, circlCfg, timezone,
		IsGithubAdvisoryTurnOn,
		IsMycertAdvisoryTurnOn,
		IsCirclCVETurnOn,
	)
	if err != nil {

		utilities.LogError("Main", "Startup", err, 0)
		return fmt.Errorf("初始化定时任务失败: %w", err)
	}

	if err := advisoryJob.MigrateAll(ctx); err != nil {
		utilities.LogError("Main", "Startup", err, 0)
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	advisoryJob.RunNow(ctx)

	if err := advisoryJob.Start(ctx); err != nil {
		utilities.LogError("Main", "Startup", err, 0)
		return fmt.Errorf("启动定时任务失败: %w", err)
	}
	defer advisoryJob.Stop()

	utilities.LogProgress("Main", "Startup",
		fmt.Sprintf("服务已就绪，定时任务将在每日 00:00 (%s) 同步 GitHub Advisory", timezone),
	)

	<-ctx.Done()

	utilities.LogProgress("Main", "Shutdown", "收到退出信号，正在优雅关闭...")
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
