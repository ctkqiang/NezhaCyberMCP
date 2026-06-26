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

// run 初始化所有依赖，并行启动定时任务调度器与 MCP stdio 服务器，
// 阻塞直到收到退出信号或 MCP 服务器退出。
//
// 启动顺序：
//  1. 建立数据库连接（通过 AdvisoryJob 内部完成）
//  2. 执行数据库迁移（幂等）
//  3. 立即执行一次数据同步（RunNow）
//  4. 启动定时任务调度器（goroutine）
//  5. 构建 MCPServer 并以 stdio 模式运行（阻塞主 goroutine）
//
// 参数：
//   - ctx : 根上下文，携带信号取消能力
//
// 返回：
//   - error : 初始化失败时返回错误
func run(ctx context.Context) error {
	utilities.LogProgress("Main", "Startup", "正在初始化...")

	dbEnv, err := utilities.ResolveDBEnvironment()
	if err != nil {
		utilities.LogError("Main", "Startup", err, 0, "step=ResolveDBEnvironment")
		return fmt.Errorf("数据库环境解析失败: %w", err)
	}

	dbCfg := buildDBConfig(dbEnv)

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
		RequestTimeout: 30 * time.Second,
		RetryMax:       5,
		RetryBackoff:   2 * time.Second,
		RateLimit:      500 * time.Millisecond,
	}

	advisoryJob, err := job.NewAdvisoryJob(
		dbCfg, scraperCfg, mycertCfg, circlCfg, timezone,
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

	// 启动时立即执行一次全量同步，确保数据库有最新数据供 MCP 工具查询。
	advisoryJob.RunNow(ctx)

	if err := advisoryJob.Start(ctx); err != nil {
		utilities.LogError("Main", "Startup", err, 0)
		return fmt.Errorf("启动定时任务失败: %w", err)
	}
	defer advisoryJob.Stop()

	// 为 MCP 服务器建立独立的长连接，与定时任务的短连接隔离。
	// 定时任务每次触发时自行建立/关闭连接；MCP 服务器需要持久连接响应实时查询。
	mcpDB, err := services.InitDatabase(ctx, dbCfg)
	if err != nil {
		utilities.LogError("Main", "Startup", err, 0, "step=InitMCPDatabase")
		return fmt.Errorf("初始化 MCP 数据库连接失败: %w", err)
	}
	defer func() {
		if closeErr := mcpDB.Close(); closeErr != nil {
			utilities.Warn("关闭 MCP 数据库连接失败: %v", closeErr)
		}
	}()

	mcpServer := services.NewMCPServer(mcpDB.DB())

	utilities.LogProgress("Main", "Startup",
		fmt.Sprintf(
			"MCP 服务器已就绪（stdio 模式），定时任务将按计划（时区 %s）同步数据",
			timezone,
		),
	)

	// Run 以 stdio 传输模式阻塞运行，直到客户端断开连接或 ctx 被取消。
	// 当进程收到 SIGINT/SIGTERM 时，ctx 取消会同时终止 MCP 服务器和定时任务。
	if err := mcpServer.Run(ctx); err != nil {
		utilities.LogError("Main", "MCPServer", err, 0)
		return fmt.Errorf("MCP 服务器异常退出: %w", err)
	}

	utilities.LogProgress(
		"Main",
		"Shutdown",
		"MCP 服务器已正常关闭",
	)

	return nil
}

// buildDBConfig 根据已解析的运行时环境类型构造 DatabaseConfiguration。
//
// 本地环境（DBEnvLocal）：
//   - 驱动固定为 PostgreSQL，从 DB_* 环境变量读取连接参数。
//
// AWS 环境（DBEnvAWS）：
//   - 驱动切换为 AmazonAuroraDSQL，从 DSQL_ENDPOINT 读取集群端点，
//     密码字段留空（由 InitDatabase 在运行时动态生成 IAM token 填充）。
//
// 参数：
//   - env : 由 utilities.ResolveDBEnvironment() 返回的环境类型
//
// 返回：
//   - services.DatabaseConfiguration : 填充完毕的数据库连接配置
func buildDBConfig(env utilities.DBEnvironment) services.DatabaseConfiguration {
	base := services.DatabaseConfiguration{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "nezha_cyber"),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	switch env {
	case utilities.DBEnvAWS:
		base.Type = services.AmazonAuroraDSQL
		base.Host = getEnv("DSQL_ENDPOINT", base.Host)
		base.AWSRegion = utilities.AWSRegion("us-east-1")
		base.Password = ""
		utilities.LogProgress("Main", "buildDBConfig", "数据库驱动=AmazonAuroraDSQL",
			fmt.Sprintf("endpoint=%s", base.Host),
			fmt.Sprintf("region=%s", base.AWSRegion),
		)

	default:
		base.Type = services.PostgreSQL
		utilities.LogProgress("Main", "buildDBConfig", "数据库驱动=PostgreSQL",
			fmt.Sprintf("host=%s", base.Host),
			fmt.Sprintf("port=%s", base.Port),
		)
	}

	return base
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
