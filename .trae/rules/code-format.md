---
alwaysApply: true
scene: code_generation
---

## 代码格式规范

本规范从项目现有代码中提炼，所有新增或修改的代码必须严格遵循。

---

### 1. 包声明与导入

- 包名使用**小写单词**，与目录名一致
- 导入分三组，组间空一行，顺序固定：
  1. 标准库
  2. 项目内部包（`nezha_cyber_mcp/internal/...`）
  3. 第三方库

```go
import (
    "context"
    "fmt"
    "time"

    "nezha_cyber_mcp/internal/model"
    "nezha_cyber_mcp/internal/utilities"

    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)
```

---

### 2. 常量

- 相关常量统一用 `const ( ... )` 块声明，不单独散落
- 每个常量必须有**简体中文单行注释**，说明其含义和取值依据
- 常量名使用 `camelCase`（包级私有）或 `PascalCase`（导出）

```go
const (
    // component 是本包在日志中使用的组件名称标识。
    component = "GithubAdvisoryService"

    // defaultRequestTimeout 是单次 HTTP 请求的默认超时时间。
    defaultRequestTimeout = 30 * time.Second

    // defaultPerPage 是每页返回的条目数，最大 100。
    defaultPerPage = 100
)
```

---

### 3. 结构体与模型

- 结构体名使用 `PascalCase`
- 每个字段必须同时标注 `json` 和 `gorm` tag，tag 值对齐（用空格补齐）
- 可空字段使用指针类型（`*time.Time`、`*string`）
- `gorm` tag 中必须显式指定 `column` 名称和 `type`（文本类字段用 `type:text`）

```go
type MycertAdvisory struct {
    AdvisoryID  string     `json:"advisory_id"  gorm:"primaryKey;column:advisory_id"`
    Title       string     `json:"title"        gorm:"column:title;type:text"`
    PublishedAt *time.Time `json:"published_at" gorm:"column:published_at"`
    FullContent string     `json:"full_content" gorm:"column:full_content;type:text"`
    ScrapedAt   time.Time  `json:"scraped_at"   gorm:"column:scraped_at;autoUpdateTime"`
}
```

---

### 4. 函数与方法注释

每个导出函数和方法必须有**简体中文文档注释**，格式固定：

```go
// FunctionName 一句话说明函数的功能。
// 可以有第二行补充说明。
//
// 参数：
//   - paramA : 参数含义说明
//   - paramB : 参数含义说明
//
// 返回：
//   - ReturnType : 返回值含义说明
//   - error      : 失败时返回包装后的错误，成功时返回 nil
func FunctionName(paramA string, paramB int) (ReturnType, error) {
```

- 未导出函数若逻辑非显而易见，也必须加注释
- 禁止空注释（`// TODO` 除外，但必须附带具体说明）

---

### 5. 错误处理

- 错误必须用 `fmt.Errorf("操作描述: %w", err)` 包装后向上传递，保留调用链
- 禁止 `_ = err` 忽略错误，除非有明确注释说明原因
- 每个错误分支必须调用 `utilities.LogError` 记录日志后再返回

```go
if err := repo.Migrate(ctx); err != nil {
    utilities.LogError(component, "Run", err, time.Since(start), "step=Migrate")
    return fmt.Errorf("migrate github_advisories: %w", err)
}
```

---

### 6. 日志规范

使用 `utilities` 包提供的结构化日志函数，禁止使用 `fmt.Println` 或 `log.Printf` 输出业务日志：

| 场景 | 函数 |
|------|------|
| 操作开始 | `utilities.LogStart(component, operation)` |
| 中间进度 | `utilities.LogProgress(component, operation, msg, details...)` |
| 操作成功 | `utilities.LogSuccess(component, operation, elapsed, details...)` |
| 操作失败 | `utilities.LogError(component, operation, err, elapsed, details...)` |
| 警告 | `utilities.LogWarn(component, operation, msg, elapsed, details...)` |

- `details` 统一用 `"key=value"` 格式传入
- `elapsed` 统一用 `time.Since(start)` 计算，`start` 在函数入口处声明

```go
func (r *Repo) BulkUpsert(ctx context.Context, items []model.T) error {
    start := time.Now()
    utilities.LogStart(component, "BulkUpsert")

    // ... 业务逻辑 ...

    utilities.LogSuccess(component, "BulkUpsert", time.Since(start),
        fmt.Sprintf("rows=%d", len(items)))
    return nil
}
```

---

### 7. 数据库操作

- 所有数据库操作必须传入 `ctx`：`r.db.WithContext(ctx).XXX`
- Upsert 统一使用 `clause.OnConflict{UpdateAll: true}`
- 批量写入使用 `CreateInBatches(items, 100)`，每批不超过 100 条
- 批量操作必须包在 `Transaction` 中，保证原子性

```go
err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    return tx.Clauses(clause.OnConflict{UpdateAll: true}).
        CreateInBatches(items, 100).Error
})
```

---

### 8. 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 包级常量（私有） | `camelCase` | `defaultPerPage` |
| 包级常量（导出） | `PascalCase` | `ApiBase` |
| 结构体 | `PascalCase` | `GithubAdvisoryRepository` |
| 方法接收者 | 单字母或双字母缩写 | `r`、`j`、`svc` |
| 局部变量 | `camelCase` | `start`、`elapsed` |
| 错误变量 | `err` | `err` |
| 上下文变量 | `ctx` | `ctx` |

---

### 9. 构造函数

- 构造函数统一命名为 `NewXxx`，返回指针类型
- nil 配置参数必须有默认值回退逻辑，不允许直接 panic

```go
func NewMycertAdvisoryService(
    repo *repository.MycertAdvisoryRepository,
    cfg  *MycertScraperConfig,
) *MycertAdvisoryService {
    c := defaultMycertConfig()
    if cfg != nil {
        if cfg.MaxPages > 0 {
            c.MaxPages = cfg.MaxPages
        }
        // ... 其他字段覆盖
    }
    return &MycertAdvisoryService{repo: repo, config: c}
}
```

---

### 10. 禁止事项

- 禁止硬编码凭据（密码、Token、密钥），必须从环境变量读取
- 禁止在业务代码中使用 `os.Exit`，只允许在 `main()` 中使用
- 禁止裸 `panic`，除非是程序初始化阶段的不可恢复错误
- 禁止在注释中使用 emoji
- 禁止无意义注释（如 `// 这里做了一些事情`）
