---
skill: Go Lang
category: 后端开发
proficiency: 90%
---

## Go Lang

**熟练度：90%**

精通 Go 语言核心特性，包括 goroutine 并发模型、channel 通信机制、interface 多态设计与 context 传播控制。熟练运用 Gin、Fiber 等 Web 框架构建高性能 RESTful API 服务，深入掌握 GORM ORM 框架进行多数据库（PostgreSQL、MySQL、SQLite）的持久化操作与事务管理。具备从零设计并落地微服务架构的实战经验，擅长利用 Go 的低内存占用与高吞吐特性构建网络安全工具、数据采集系统与定时任务调度平台。

### 核心能力

- **并发编程**：熟练运用 goroutine、channel、sync 包实现高并发任务调度与数据管道
- **Web 服务**：基于 Gin / net/http 构建 RESTful API，支持中间件链、路由分组与优雅关闭
- **数据库集成**：使用 GORM 对接 PostgreSQL / MySQL / SQLite，掌握 AutoMigrate、事务、Upsert（OnConflict）等高级用法
- **定时任务**：使用 robfig/cron 实现多时区 cron 调度，支持优雅停止与任务隔离
- **HTTP 客户端**：封装带指数退避重试、速率限制、Token 认证的 HTTP 客户端，对接 GitHub REST API 等第三方服务
- **HTML 解析**：使用 golang.org/x/net/html 进行 DOM 树遍历，实现结构化网页数据提取
- **日志与可观测性**：设计结构化日志系统，支持多级别（DEBUG / INFO / WARN / ERROR）、ANSI 彩色输出与运行时内存追踪
- **测试**：使用标准库 testing 包配合 SQLite 内存数据库编写单元测试，覆盖 Repository 层、Service 层与 HTTP Mock 场景

### 专项领域

- **网络安全工具开发**：构建安全公告自动采集与持久化平台（NezhaCyberMCP），对接 GitHub Advisory API 与 MyCERT 门户
- **数据采集系统**：实现带分页、限速、幂等写入的多源数据同步管道
- **CLI 工具**：使用 Go 标准库构建跨平台命令行工具，无外部依赖，单二进制分发

### 代表项目

| 项目             | 技术栈                        | 说明                                                 |
| ---------------- | ----------------------------- | ---------------------------------------------------- |
| NezhaCyberMCP    | Go / GORM / PostgreSQL / cron | 多源安全公告自动采集与定时同步平台                   |
| 哪吒 AI 后端服务 | Go / Gin / REST API           | 网络安全平台后端，提供恶意软件扫描与数据泄露检测接口 |
