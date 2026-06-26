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

	"github.com/aws/aws-lambda-go/lambda"
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

	// 根据运行时环境选择执行模式：
	//   - 本地模式：启动 cron 调度器 + MCP stdio 服务器（长驻进程）
	//   - AWS Lambda 模式：注册 Lambda 处理函数，由 Lambda 运行时驱动单次执行
	if utilities.IsLocalMode() {
		utilities.LogProgress("Main", "Startup", "运行模式=local，启动 cron+MCP stdio 服务器")

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		if err := run(ctx); err != nil {
			utilities.Error("fatal: %v", err)
			os.Exit(1)
		}
		return
	}

	utilities.LogProgress("Main", "Startup", "运行模式=lambda，注册 Lambda 处理函数")
	lambda.StartWithOptions(lambdaHandler, lambda.WithContext(context.Background()))
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

	timezone := getEnv("DB_TIMEZONE", "Asia/Shanghai")

	// MCP 服务器立即以 nil DB 启动，确保 initialize 握手不被任何 I/O 延迟阻断。
	// DB 环境解析、连接、迁移和 cron 调度全部在后台 goroutine 中完成。
	// DB 就绪后通过 SetDB 热注入，无需重启服务器。
	// DB 不可用时工具调用返回明确错误，握手和工具列表查询始终正常。
	mcpServer := services.NewMCPServer(nil)

	go func() {
		dbEnv, err := utilities.ResolveDBEnvironment()
		if err != nil {
			utilities.LogError("Main", "BackgroundInit", err, 0, "step=ResolveDBEnvironment")
			return
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

		mycertCfg := &services.MycertScraperConfig{
			MaxPages:       0,
			RequestTimeout: 30 * time.Second,
			FetchDetail:    true,
		}

		// CIRCL_API_TOKEN 从环境变量读取，不在此处硬编码。
		circlCfg := &services.CirclScraperConfig{
			RequestTimeout: 30 * time.Second,
			RetryMax:       5,
			RetryBackoff:   2 * time.Second,
			RateLimit:      500 * time.Millisecond,
		}

		// 为 MCP 服务器建立独立的长连接，就绪后热注入。
		mcpDB, dbErr := services.InitDatabase(ctx, dbCfg)
		if dbErr != nil {
			utilities.LogWarn("Main", "BackgroundInit",
				fmt.Sprintf("数据库连接失败，MCP 工具将返回错误直到 DB 就绪: %v", dbErr),
				0, "step=InitMCPDatabase",
			)
		} else {
			defer func() {
				if closeErr := mcpDB.Close(); closeErr != nil {
					utilities.Warn("关闭 MCP 数据库连接失败: %v", closeErr)
				}
			}()
			mcpServer.SetDB(mcpDB.DB())
		}

		advisoryJob, err := job.NewAdvisoryJob(
			dbCfg, scraperCfg, mycertCfg, circlCfg, timezone,
			IsGithubAdvisoryTurnOn,
			IsMycertAdvisoryTurnOn,
			IsCirclCVETurnOn,
		)
		if err != nil {
			utilities.LogError("Main", "BackgroundInit", err, 0, "step=NewAdvisoryJob")
			return
		}

		if err := advisoryJob.MigrateAll(ctx); err != nil {
			utilities.LogError("Main", "BackgroundInit", err, 0, "step=MigrateAll")
			return
		}

		advisoryJob.RunNow(ctx)

		if err := advisoryJob.Start(ctx); err != nil {
			utilities.LogError("Main", "BackgroundInit", err, 0, "step=Start")
			return
		}

		<-ctx.Done()
		advisoryJob.Stop()
	}()

	utilities.LogProgress("Main", "Startup",
		fmt.Sprintf(
			"MCP 服务器已就绪（stdio 模式），定时任务将按计划（时区 %s）同步数据",
			timezone,
		),
	)

	// Run 以 stdio 传输模式阻塞运行，直到客户端断开连接或 ctx 被取消。
	if err := mcpServer.Run(ctx); err != nil {
		utilities.LogError("Main", "MCPServer", err, 0)
		return fmt.Errorf("MCP 服务器异常退出: %w", err)
	}

	utilities.LogProgress("Main", "Shutdown", "MCP 服务器已正常关闭")
	return nil
}

// LambdaEvent 是 Lambda 函数接收的触发事件结构。
// 支持 EventBridge 定时触发（source="aws.events"）和手动调用（source=""）。
// 其余字段由 Lambda 运行时填充，业务逻辑无需关心。
type LambdaEvent struct {
	// Source 标识触发来源，EventBridge 定时规则为 "aws.events"。
	Source string `json:"source"`

	// DetailType 描述事件类型，如 "Scheduled Event"。
	DetailType string `json:"detail-type"`
}

// LambdaResponse 是 Lambda 函数的返回结构，供调用方或 EventBridge 记录。
type LambdaResponse struct {
	// Status 表示本次执行结果，"ok" 或 "error"。
	Status string `json:"status"`

	// Message 包含执行摘要或错误描述。
	Message string `json:"message"`
}

// lambdaHandler 是注册到 AWS Lambda 运行时的处理函数。
// 每次 Lambda 调用触发一次完整的数据同步流程（迁移 + 同步），不启动 cron 和 MCP 服务器。
// ctx 由 Lambda 运行时注入，携带调用超时信息。
//
// 参数：
//   - ctx   : Lambda 运行时上下文，含调用截止时间
//   - event : 触发事件（EventBridge 定时规则或手动调用）
//
// 返回：
//   - LambdaResponse : 执行结果摘要
//   - error          : 执行失败时返回错误，Lambda 运行时将标记本次调用为失败
func lambdaHandler(ctx context.Context, event LambdaEvent) (LambdaResponse, error) {
	start := time.Now()
	utilities.LogStart("Main", "LambdaHandler")
	utilities.LogProgress("Main", "LambdaHandler",
		fmt.Sprintf("触发来源: source=%q detail-type=%q", event.Source, event.DetailType),
	)

	if err := runLambda(ctx); err != nil {
		utilities.LogError("Main", "LambdaHandler", err, time.Since(start))
		return LambdaResponse{Status: "error", Message: err.Error()}, err
	}

	utilities.LogSuccess("Main", "LambdaHandler", time.Since(start))
	return LambdaResponse{
		Status:  "ok",
		Message: fmt.Sprintf("数据同步完成，耗时 %s", time.Since(start).Round(time.Millisecond)),
	}, nil
}

// runLambda 在 AWS Lambda 环境中执行一次完整的数据同步流程。
// 与 run() 的区别：
//   - 不启动 cron 调度器（Lambda 由 EventBridge 定时规则驱动）
//   - 不启动 MCP stdio 服务器（Lambda 无持久连接）
//   - 执行完成后立即返回，由 Lambda 运行时负责进程生命周期
//
// 参数：
//   - ctx : Lambda 运行时上下文，携带调用超时信息
//
// 返回：
//   - error : 初始化、迁移或同步失败时返回错误
func runLambda(ctx context.Context) error {
	utilities.LogProgress("Main", "runLambda", "正在初始化 Lambda 执行环境...")

	dbEnv, err := utilities.ResolveDBEnvironment()
	if err != nil {
		utilities.LogError("Main", "runLambda", err, 0, "step=ResolveDBEnvironment")
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
		utilities.LogError("Main", "runLambda", err, 0, "step=NewAdvisoryJob")
		return fmt.Errorf("初始化定时任务失败: %w", err)
	}

	if err := advisoryJob.MigrateAll(ctx); err != nil {
		utilities.LogError("Main", "runLambda", err, 0, "step=MigrateAll")
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	// Lambda 模式下直接执行一次全量同步，不启动 cron 调度器。
	advisoryJob.RunNow(ctx)

	utilities.LogProgress("Main", "runLambda", "Lambda 数据同步执行完毕",
		fmt.Sprintf("timezone=%s", timezone),
	)
	return nil
}

// buildDBConfig 根据已解析的运行时环境类型构造 DatabaseConfiguration。
//
// 优先级（从高到低）：
//  1. DB_SQLITE=true：无论环境如何，使用 SQLite，固定路径 /tmp/nezha_cyber.db。
//     适用于 Lambda 无状态场景或本地快速测试，无需任何外部数据库。
//  2. AWS 环境（DBEnvAWS）：使用 Amazon Aurora DSQL，从 DSQL_ENDPOINT 读取端点。
//  3. 本地环境（DBEnvLocal）：使用 PostgreSQL，从 DB_* 环境变量读取连接参数。
//
// 参数：
//   - env : 由 utilities.ResolveDBEnvironment() 返回的环境类型
//
// 返回：
//   - services.DatabaseConfiguration : 填充完毕的数据库连接配置
func buildDBConfig(env utilities.DBEnvironment) services.DatabaseConfiguration {
	if getEnv("DB_SQLITE", "false") == "true" {
		utilities.LogProgress("Main", "buildDBConfig", "数据库驱动=SQLite",
			fmt.Sprintf("path=%s", "/tmp/nezha_cyber.db"),
		)
		return services.DatabaseConfiguration{
			Type: services.SQLite,
		}
	}

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
		dsqlEndpoint := getEnv("DSQL_ENDPOINT", "")
		if dsqlEndpoint == "" {
			// DSQL_ENDPOINT 未配置，自动回退到 SQLite。
			// 这允许 Lambda 在没有外部数据库的情况下正常运行（数据存储在 /tmp，重启后清空）。
			utilities.LogProgress("Main", "buildDBConfig",
				"DSQL_ENDPOINT 未设置，自动回退到 SQLite",
				fmt.Sprintf("path=%s", "/tmp/nezha_cyber.db"),
			)
			return services.DatabaseConfiguration{
				Type: services.SQLite,
			}
		}
		base.Type = services.AmazonAuroraDSQL
		base.Host = dsqlEndpoint
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
