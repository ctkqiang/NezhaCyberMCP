package job

import (
	"context"
	"fmt"
	"nezha_cyber_mcp/internal/repository"
	"nezha_cyber_mcp/internal/services"
	"nezha_cyber_mcp/internal/utilities"
	"time"

	"github.com/robfig/cron/v3"
)

const component = "AdvisoryJob"

// AdvisoryJob 封装了 GitHub 和 MyCERT 安全公告定时同步任务。
// 内部持有一个 cron 调度器，按照指定时区的 cron 表达式定期触发抓取流程。
type AdvisoryJob struct {
	scheduler      *cron.Cron
	dbCfg          services.DatabaseConfiguration
	scraperCfg     *services.AdvisoryScraperConfig
	mycertCfg      *services.MycertScraperConfig
}

// NewAdvisoryJob 构造 AdvisoryJob 实例。
// 时区通过 DB_TIMEZONE 环境变量指定（如 "Asia/Shanghai"），
// 解析失败时回退到 UTC，确保程序不会因时区配置错误而崩溃。
//
// 参数：
//   - dbCfg      : 数据库连接配置，每次任务触发时重新建立连接
//   - scraperCfg : GitHub API 客户端配置
//   - mycertCfg  : MyCERT 爬虫配置，传 nil 使用默认值
//   - timezone   : IANA 时区名称（如 "Asia/Shanghai"），空字符串时使用 UTC
//
// 返回：
//   - *AdvisoryJob
//   - error : 时区名称无效时返回错误
func NewAdvisoryJob(
	dbCfg services.DatabaseConfiguration,
	scraperCfg *services.AdvisoryScraperConfig,
	mycertCfg *services.MycertScraperConfig,
	timezone string,
) (*AdvisoryJob, error) {
	if timezone == "" {
		timezone = "UTC"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("无效的时区 %q: %w", timezone, err)
	}

	scheduler := cron.New(cron.WithLocation(loc))

	return &AdvisoryJob{
		scheduler:  scheduler,
		dbCfg:      dbCfg,
		scraperCfg: scraperCfg,
		mycertCfg:  mycertCfg,
	}, nil
}

// MigrateAll 在程序启动时立即建立一次数据库连接，
// 对 github_advisories 和 mycert_advisories 两张表执行幂等迁移，
// 确保表结构在第一次 cron 触发前就已存在。
//
// 参数：
//   - ctx : 请求上下文
//
// 返回：
//   - error : 连接或迁移失败时返回错误
func (j *AdvisoryJob) MigrateAll(ctx context.Context) error {
	start := time.Now()
	utilities.LogStart(component, "MigrateAll")

	db, err := services.InitDatabase(ctx, j.dbCfg)
	if err != nil {
		utilities.LogError(component, "MigrateAll", err, time.Since(start))
		return fmt.Errorf("MigrateAll: 连接数据库失败: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			utilities.Warn("MigrateAll: 关闭数据库连接失败: %v", closeErr)
		}
	}()

	githubRepo := repository.NewGithubAdvisoryRepository(db.DB())
	if err := githubRepo.Migrate(ctx); err != nil {
		utilities.LogError(component, "MigrateAll", err, time.Since(start),
			"table=github_advisories")
		return fmt.Errorf("MigrateAll: github_advisories: %w", err)
	}

	mycertRepo := repository.NewMycertAdvisoryRepository(db.DB())
	if err := mycertRepo.Migrate(ctx); err != nil {
		utilities.LogError(component, "MigrateAll", err, time.Since(start),
			"table=mycert_advisories")
		return fmt.Errorf("MigrateAll: mycert_advisories: %w", err)
	}

	utilities.LogSuccess(component, "MigrateAll", time.Since(start),
		"tables=github_advisories,mycert_advisories")
	return nil
}

// Start 向调度器注册每日 00:00 触发的任务，并启动调度器。
// 调度器在后台 goroutine 中运行，不阻塞调用方。
// 程序退出时应调用 Stop 优雅关闭。
//
// 参数：
//   - ctx : 根上下文，传递给每次任务执行，支持整体取消
//
// 返回：
//   - error : 注册 cron 表达式失败时返回错误
func (j *AdvisoryJob) Start(ctx context.Context) error {
	// "0 0 * * *" 表示每天 00:00:00 触发（标准 5 字段格式：分 时 日 月 周）。
	_, err := j.scheduler.AddFunc("0 0 * * *", func() {
		j.run(ctx)
	})
	if err != nil {
		return fmt.Errorf("注册 cron 任务失败: %w", err)
	}

	j.scheduler.Start()

	utilities.LogProgress(component, "Start",
		"定时任务已启动，将在每日 00:00 执行 GitHub Advisory 同步")

	return nil
}

// Stop 优雅停止调度器，等待当前正在执行的任务完成后再退出。
// 应在程序收到退出信号时调用（通常通过 defer 注册）。
func (j *AdvisoryJob) Stop() {
	utilities.LogProgress(component, "Stop", "正在停止定时任务调度器...")
	stopCtx := j.scheduler.Stop()
	<-stopCtx.Done()
	utilities.LogProgress(component, "Stop", "调度器已完全停止")
}

// run 是每次 cron 触发时执行的核心逻辑。
// 为每次执行独立建立数据库连接，任务结束后关闭，避免长连接空闲超时。
//
// 参数：
//   - ctx : 根上下文，若程序已收到退出信号则任务会提前终止
func (j *AdvisoryJob) run(ctx context.Context) {
	start := time.Now()
	utilities.LogStart(component, "Run")

	db, err := services.InitDatabase(ctx, j.dbCfg)
	if err != nil {
		utilities.LogError(component, "Run", err, time.Since(start),
			"step=InitDatabase")
		return
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			utilities.Warn("任务结束后关闭数据库连接失败: %v", closeErr)
		}
	}()

	repo := repository.NewGithubAdvisoryRepository(db.DB())

	// 执行表迁移（幂等），确保表结构始终与模型一致。
	if err := repo.Migrate(ctx); err != nil {
		utilities.LogError(component, "Run", err, time.Since(start),
			"step=Migrate")
		return
	}

	githubSvc := services.NewGithubAdvisoryService(repo, j.scraperCfg)

	githubTotal, err := githubSvc.ScrapeAndPersist(ctx)
	if err != nil {
		utilities.LogError(component, "Run", err, time.Since(start),
			"step=GitHub.ScrapeAndPersist",
			fmt.Sprintf("partial_total=%d", githubTotal))
		return
	}

	githubCount, _ := repo.Count(ctx)
	utilities.LogSuccess(component, "Run.GitHub", time.Since(start),
		fmt.Sprintf("scraped_this_run=%d", githubTotal),
		fmt.Sprintf("total_in_db=%d", githubCount))

	mycertRepo := repository.NewMycertAdvisoryRepository(db.DB())

	if err := mycertRepo.Migrate(ctx); err != nil {
		utilities.LogError(component, "Run", err, time.Since(start),
			"step=MyCERT.Migrate")
		return
	}

	mycertSvc := services.NewMycertAdvisoryService(mycertRepo, j.mycertCfg)

	mycertTotal, err := mycertSvc.ScrapeAndPersist(ctx)
	if err != nil {
		utilities.LogError(component, "Run", err, time.Since(start),
			"step=MyCERT.ScrapeAndPersist",
			fmt.Sprintf("partial_total=%d", mycertTotal))
		return
	}

	mycertCount, _ := mycertRepo.Count(ctx)
	utilities.LogSuccess(component, "Run.MyCERT", time.Since(start),
		fmt.Sprintf("scraped_this_run=%d", mycertTotal),
		fmt.Sprintf("total_in_db=%d", mycertCount))
}
