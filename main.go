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

	// 从环境变量读取数据库连接参数，未设置时使用括号内的默认值。
	// 生产环境请通过 .env 文件或容器环境变量注入，切勿硬编码凭据。
	dbCfg := services.DatabaseConfiguration{
		Type:            services.PostgreSQL,
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "nezha_cyber"),
		MaxOpenConns:    25,              // 连接池最大打开连接数
		MaxIdleConns:    5,               // 连接池最大空闲连接数
		ConnMaxLifetime: 5 * time.Minute, // 单条连接最大存活时间
	}

	// 配置 GitHub API 客户端参数。
	// GITHUB_TOKEN 从环境变量读取，认证后速率限制从 60 次/小时提升至 5000 次/小时。
	// MaxPages=0 表示拉取全部页面（GitHub Advisory 目前约 400 页）。
	scraperCfg := &services.AdvisoryScraperConfig{
		MaxPages:       0,
		RequestTimeout: 30 * time.Second,
		PerPage:        100,
		RetryMax:       5,
		RetryBackoff:   2 * time.Second,
		Token:          getEnv("GITHUB_TOKEN", ""),
	}

	// 从 DB_TIMEZONE 读取时区，用于 cron 表达式的时间解析。
	// .env 中配置为 Asia/Shanghai，即每日 00:00 CST 触发。
	timezone := getEnv("DB_TIMEZONE", "Asia/Shanghai")

	advisoryJob, err := job.NewAdvisoryJob(dbCfg, scraperCfg, timezone)
	if err != nil {
		utilities.LogError("Main", "Startup", err, 0)
		return fmt.Errorf("初始化定时任务失败: %w", err)
	}

	// 启动 cron 调度器（后台 goroutine），注册每日 00:00 触发的任务。
	if err := advisoryJob.Start(ctx); err != nil {
		utilities.LogError("Main", "Startup", err, 0)
		return fmt.Errorf("启动定时任务失败: %w", err)
	}
	// 程序退出时优雅停止调度器，等待当前任务执行完毕。
	defer advisoryJob.Stop()

	utilities.LogProgress("Main", "Startup",
		fmt.Sprintf("服务已就绪，定时任务将在每日 00:00 (%s) 同步 GitHub Advisory", timezone))

	// 阻塞主 goroutine，直到收到 SIGINT / SIGTERM 信号。
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
