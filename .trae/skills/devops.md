---
skill: DevOps
category: DevOps / 基础设施
proficiency: 75%
---

## DevOps

**熟练度：75%**

熟悉容器化技术与 CI/CD 流程，能够使用 Docker 构建、打包与部署应用服务，掌握 Docker Compose 编排多容器环境。具备 Git 版本控制与分支管理的丰富实践经验，熟悉 GitHub Actions 自动化工作流配置，有服务器部署与运维的实际经验。

### 核心能力

- **容器化**：Docker 镜像构建、多阶段构建优化、Docker Compose 多服务编排
- **数据库容器化**：Docker 运行 PostgreSQL / MySQL，持久化卷挂载与环境变量注入
- **版本控制**：Git 分支策略（feature / main / release），PR 审查流程，语义化提交规范
- **CI/CD**：GitHub Actions 自动化测试、构建与部署流水线配置
- **环境管理**：`.env` 文件管理多环境配置，敏感信息通过环境变量注入，不硬编码凭据
- **进程管理**：信号处理（SIGTERM / SIGINT）实现优雅关闭，`defer` 确保资源释放
- **Kubernetes 基础**：了解 Pod、Service、Deployment 基本概念与 kubectl 常用操作

### 专项领域

- **Go 服务容器化**：NezhaCyberMCP 服务的 Docker 打包与 PostgreSQL 容器编排
- **定时任务运维**：cron 调度器的优雅启停，任务执行日志结构化输出与监控
- **多环境配置管理**：通过 `.env` 文件区分开发 / 测试 / 生产环境，支持 Docker 环境变量覆盖

### 代表项目

| 项目 | 技术栈 | 说明 |
|------|--------|------|
| NezhaCyberMCP | Docker / PostgreSQL / Go | 容器化部署的安全公告同步服务 |
| 哪吒 AI | Vercel / NextJS | 前端自动化部署至 Vercel |
