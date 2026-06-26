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

// AdvisoryJob 封装了 GitHub、MyCERT 和 CIRCL CVE 安全公告定时同步任务。
// 内部持有一个 cron 调度器，按照指定时区的 cron 表达式定期触发抓取流程。
// githubOn、mycertOn 和 circlOn 控制各数据源是否参与每次同步，false 时跳过对应抓取器。
type AdvisoryJob struct {
	scheduler  *cron.Cron
	dbCfg      services.DatabaseConfiguration
	scraperCfg *services.AdvisoryScraperConfig
	mycertCfg  *services.MycertScraperConfig
	circlCfg   *services.CirclScraperConfig
	githubOn   bool
	mycertOn   bool
	circlOn    bool
}

// NewAdvisoryJob 构造 AdvisoryJob 实例。
// 时区通过 DB_TIMEZONE 环境变量指定（如 "Asia/Shanghai"），
// 解析失败时回退到 UTC，确保程序不会因时区配置错误而崩溃。
//
// 参数：
//   - dbCfg      : 数据库连接配置，每次任务触发时重新建立连接
//   - scraperCfg : GitHub API 客户端配置
//   - mycertCfg  : MyCERT 爬虫配置，传 nil 使用默认值
//   - circlCfg   : CIRCL CVE API 客户端配置，传 nil 使用默认值
//   - timezone   : IANA 时区名称（如 "Asia/Shanghai"），空字符串时使用 UTC
//   - githubOn   : 是否启用 GitHub Advisory 同步，false 时跳过
//   - mycertOn   : 是否启用 MyCERT 公告同步，false 时跳过
//   - circlOn    : 是否启用 CIRCL CVE 同步，false 时跳过
//
// 返回：
//   - *AdvisoryJob
//   - error : 时区名称无效时返回错误
func NewAdvisoryJob(
	dbCfg services.DatabaseConfiguration,
	scraperCfg *services.AdvisoryScraperConfig,
	mycertCfg *services.MycertScraperConfig,
	circlCfg *services.CirclScraperConfig,
	timezone string,
	githubOn bool,
	mycertOn bool,
	circlOn bool,
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
		circlCfg:   circlCfg,
		githubOn:   githubOn,
		mycertOn:   mycertOn,
		circlOn:    circlOn,
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
			"table=mycert_advisories",
		)

		return fmt.Errorf("MigrateAll: mycert_advisories: %w", err)
	}

	circlRepo := repository.NewCirclCVERepository(db.DB())
	if err := circlRepo.Migrate(ctx); err != nil {
		utilities.LogError(component, "MigrateAll", err, time.Since(start),
			"table=circl_cves")
		return fmt.Errorf("MigrateAll: circl_cves: %w", err)
	}

	utilities.LogSuccess(component, "MigrateAll", time.Since(start),
		"tables=github_advisories,mycert_advisories,circl_cves")
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
	// "0 */3 * * *" 表示每 3 小时整点触发一次（00:00 / 03:00 / 06:00 … 21:00）。
	_, err := j.scheduler.AddFunc("0 */3 * * *", func() {
		j.run(ctx)
	})
	if err != nil {
		return fmt.Errorf("注册 cron 任务失败: %w", err)
	}

	j.scheduler.Start()

	utilities.LogProgress(component, "Start",
		"定时任务已启动，将每 3 小时执行一次 GitHub Advisory 与 MyCERT 同步")

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

// RunNow 立即同步执行一次完整的抓取流程（GitHub Advisory + MyCERT），
// 不等待 cron 调度触发。通常在程序启动后调用，确保数据库在第一次 cron 触发前已有数据。
//
// 参数：
//   - ctx : 请求上下文，支持超时与取消控制
func (j *AdvisoryJob) RunNow(ctx context.Context) {
	utilities.LogProgress(component, "RunNow", "启动时立即执行一次全量同步")
	j.run(ctx)
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

	if j.githubOn {
		j.runGithub(ctx, db, start)
	}
	if j.mycertOn {
		j.runMycert(ctx, db, start)
	}
	if j.circlOn {
		j.runCircl(ctx, db, start)
	}

	if !j.githubOn && !j.mycertOn && !j.circlOn {
		utilities.LogWarn(component, "Run",
			"所有数据源均已关闭，本次同步跳过", time.Since(start))
	}
}

// runGithub 执行 GitHub Advisory API 抓取并持久化。
//
// 参数：
//   - ctx   : 请求上下文
//   - db    : 已建立的数据库连接
//   - start : 本次 run 的起始时间，用于计算 elapsed
func (j *AdvisoryJob) runGithub(ctx context.Context, db *services.Database, start time.Time) {
	repo := repository.NewGithubAdvisoryRepository(db.DB())

	if err := repo.Migrate(ctx); err != nil {
		utilities.LogError(component, "Run.GitHub", err, time.Since(start),
			"step=Migrate")
		return
	}

	githubSvc := services.NewGithubAdvisoryService(repo, j.scraperCfg)

	githubTotal, err := githubSvc.ScrapeAndPersist(ctx)
	if err != nil {
		utilities.LogError(component, "Run.GitHub", err, time.Since(start),
			"step=ScrapeAndPersist",
			fmt.Sprintf("partial_total=%d", githubTotal))
		return
	}

	githubCount, _ := repo.Count(ctx)
	utilities.LogSuccess(component, "Run.GitHub", time.Since(start),
		fmt.Sprintf("scraped_this_run=%d", githubTotal),
		fmt.Sprintf("total_in_db=%d", githubCount))
}

// runMycert 执行 MyCERT 公告爬取并持久化。
//
// 参数：
//   - ctx   : 请求上下文
//   - db    : 已建立的数据库连接
//   - start : 本次 run 的起始时间，用于计算 elapsed
func (j *AdvisoryJob) runMycert(ctx context.Context, db *services.Database, start time.Time) {
	mycertRepo := repository.NewMycertAdvisoryRepository(db.DB())

	if err := mycertRepo.Migrate(ctx); err != nil {
		utilities.LogError(component, "Run.MyCERT", err, time.Since(start),
			"step=Migrate")
		return
	}

	mycertSvc := services.NewMycertAdvisoryService(mycertRepo, j.mycertCfg)

	mycertTotal, err := mycertSvc.ScrapeAndPersist(ctx)
	if err != nil {
		utilities.LogError(component, "Run.MyCERT", err, time.Since(start),
			"step=ScrapeAndPersist",
			fmt.Sprintf("partial_total=%d", mycertTotal))
		return
	}

	mycertCount, _ := mycertRepo.Count(ctx)
	utilities.LogSuccess(component, "Run.MyCERT", time.Since(start),
		fmt.Sprintf("scraped_this_run=%d", mycertTotal),
		fmt.Sprintf("total_in_db=%d", mycertCount))
}

// runCircl 执行 CIRCL CVE API 抓取并持久化。
//
// 参数：
//   - ctx   : 请求上下文
//   - db    : 已建立的数据库连接
//   - start : 本次 run 的起始时间，用于计算 elapsed
func (j *AdvisoryJob) runCircl(ctx context.Context, db *services.Database, start time.Time) {
	circlRepo := repository.NewCirclCVERepository(db.DB())

	if err := circlRepo.Migrate(ctx); err != nil {
		utilities.LogError(component, "Run.CIRCL", err, time.Since(start),
			"step=Migrate")
		return
	}

	circlSvc := services.NewCirclCVEService(circlRepo, j.circlCfg)

	circlTotal, err := circlSvc.ScrapeAndPersist(ctx)
	if err != nil {
		utilities.LogError(component, "Run.CIRCL", err, time.Since(start),
			"step=ScrapeAndPersist",
			fmt.Sprintf("partial_total=%d", circlTotal))
		return
	}

	circlCount, _ := circlRepo.Count(ctx)
	utilities.LogSuccess(component, "Run.CIRCL", time.Since(start),
		fmt.Sprintf("scraped_this_run=%d", circlTotal),
		fmt.Sprintf("total_in_db=%d", circlCount))
}
