---
skill: 数据库
category: 数据库
proficiency: 80%
---

## 数据库

**熟练度：80%**

精通关系型与非关系型数据库的设计、查询优化与运维管理。熟练运用 MySQL、TiDB、PostgreSQL、SQLite 等关系型数据库，以及 Redis、MongoDB、Elasticsearch 等 NoSQL 数据库，能够根据业务场景合理选型并设计高效的数据库架构。具备 GORM、MyBatis、Hibernate 等 ORM 框架的深度使用经验。

### 核心能力

- **关系型数据库**：MySQL、TiDB、PostgreSQL、SQLite，掌握索引设计、事务管理与查询优化
- **NoSQL 数据库**：Redis（缓存、分布式锁）、MongoDB（文档存储）、Elasticsearch（全文检索）
- **ORM 框架**：GORM（Go）、MyBatis / Hibernate（Java），熟练处理 Upsert、批量写入、关联查询
- **数据库迁移**：AutoMigrate（GORM）、Flyway / Liquibase，幂等迁移保障版本一致性
- **连接池管理**：HikariCP（Java）、GORM 连接池配置，MaxOpenConns / MaxIdleConns 调优
- **数据建模**：ER 图设计，范式化与反范式化权衡，JSONB 字段存储半结构化数据
- **容器化部署**：Docker 运行 PostgreSQL / MySQL，挂载持久化卷，环境变量注入配置

### 专项领域

- **幂等写入设计**：`ON CONFLICT DO UPDATE`（PostgreSQL）实现 Upsert，保障多次同步数据一致性
- **批量事务处理**：`CreateInBatches` 分批写入，单事务原子性保障，防止大批量 SQL 超限
- **多数据库兼容**：同一 ORM 层兼容 PostgreSQL / MySQL / SQLite，通过驱动切换适配不同环境

### 代表项目

| 项目             | 数据库              | 说明                            |
| ---------------- | ------------------- | ------------------------------- |
| NezhaCyberMCP    | PostgreSQL / SQLite | 安全公告持久化，Upsert 幂等写入 |
| 链家房源数据爬虫 | MySQL               | 房源数据存储与统计查询          |
| 宝玲宠物平台     | PostgreSQL          | 宠物诊所业务数据管理            |
