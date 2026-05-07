# FindX P0 统一配置与数据源契约矩阵

生成时间：2026-05-07 14:05（UTC+8）
状态：P0-CONFIG-DATASOURCE-CONTRACT 编码前门禁，不代表后端或前端契约已经完成

## 1. 结论

FindX 必须形成一套统一配置与数据源契约：MySQL 做权威配置和业务数据基石，Redis 做缓存/队列/运行态辅助，Prometheus/VictoriaMetrics、链路 OAP、日志存储、ClickHouse、Qdrant、CMDB、Agent 包仓库等都通过统一数据源中心、凭据引用、权限审计和错误脱敏接入。

当前代码存在配置分散、旧命名残留、部分数据源只在前端静态表单中表达、链路/日志/向量数据库缺少统一契约等问题。进入数据源、指标查询、链路监控、日志中心、知识库向量化或 Agent 包管理实现前，必须先收敛本矩阵。

| 项目 | 当前状态 | 结论 |
| --- | --- | --- |
| MySQL | `api/config.yaml` 与 `api/config.yaml.example` 中有 `mysql.dsn` | 作为权威基石保留，真实 DSN 只允许环境变量或本地配置，不进 Git |
| Redis | `redis.addr/db/password` 与旧 `n9e.redis_*` 并存 | 需要统一为 FindX runtime cache/queue 配置，旧来源只做兼容 |
| Prometheus | `prometheus.url`、`data_sources[]`、前端数据源页面并存 | 必须统一到数据源中心；`prometheus.url` 仅做默认 bootstrap |
| 基础监控数据源 | 成熟源码有完整 datasource plugin/list/upsert/status/test 结构 | FindX 需要同源实现数据源类型、表单、测试连接、启停、权限、审计 |
| 链路监控 | 本地有 UI/OAP 源码，但 FindX 配置缺 OAP endpoint/adapter 契约 | 必须新增 APM Adapter 配置，不直接暴露 GraphQL 为公共业务契约 |
| 日志中心 | 本地有日志前端源码，但 FindX 配置缺 Logs Adapter/ClickHouse/OTel Collector 契约 | 必须新增 Logs Adapter 配置和存储后端选择，不允许静态假日志页 |
| 知识库向量 | 当前 embedding/reranker 存在，Qdrant 缺配置契约 | MySQL 权威存储 + Qdrant 本地 sidecar 加速层 + BM25 降级路径 |
| Agent 包仓库 | FindX Agent API 有注册/心跳，SkyWalking Agent 多语言包缺本地源码与包契约 | 必须先定义包仓库、版本、签名、平台、远程安装、数据到达验证契约 |

## 2. 当前 FindX 配置事实

| 配置文件/代码 | 事实 | 问题 |
| --- | --- | --- |
| `api/config.yaml` | 存在 `ai`、`ai_providers`、`data_sources`、`embedding`、`mysql`、`n9e`、`prometheus`、`redis`、`reranker`、`scheduler`、`server`、`workflow.cache` | 旧 `n9e` 配置与 FindX 命名并存；链路/日志/Qdrant/Agent 包仓库缺显式配置 |
| `api/config.yaml.example` | 使用 `<DB_DSN>` 等占位符，包含 `findx_agents.shared_token`、`allow_anonymous` | 示例仍写“系统配置 → 数据源”等旧入口描述，后续需跟导航矩阵改为“集成中心 → 数据源” |
| `api/internal/monitoring/datasources.go` | 从 `data_sources[]` 和 `prometheus.url` 合并 Prometheus 数据源 | 仅覆盖 Prometheus；数据源类型、凭据引用、TLS/Header、测试连接状态仍不完整 |
| `api/internal/embedding/config.go` | embedding/reranker 从 `ai_settings` 读取，fallback 到 viper | 知识库向量数据库未纳入；API key 存储和脱敏策略需要统一凭据引用 |
| `api/main.go` | 已有 `/settings/ai`、`/settings/embedding`、`/settings/reranker`、`/monitor/datasources`、`/findx-agents/*` 等入口 | 数据源、AI、Embedding、Agent、旧兼容入口分散，需统一权限和审计契约 |
| `web/src/views/DataSource.vue` | 前端有 Prometheus、Loki、MySQL、ClickHouse 等静态类型和默认 URL | 不是成熟源码同源页面；不能作为 P0 数据源主路径 |
| `web/src/components/monitoring/MonitorDatasourceQueryPanel.vue` | 有数据源选择、查询、labels/metrics 等接口调用 | 当前是自研查询面板，不能替代成熟指标查询页 |

## 3. 目标配置分层

| 层级 | 作用 | 存储位置 | 规则 |
| --- | --- | --- | --- |
| Bootstrap 配置 | 启动服务、连接 MySQL、Redis、对象存储、默认管理员、安全开关 | `config.yaml` + 环境变量 | 只放启动必需项；真实敏感值不进 Git |
| 权威配置 | 数据源、凭据引用、AI 模型、向量索引、Agent 包、模板导入记录、权限审计策略 | MySQL | 所有业务配置以 MySQL 为准，支持审计、版本和回滚 |
| 加速/运行态 | Redis cache、任务锁、workflow cache、Qdrant 向量索引、短期状态 | Redis/Qdrant/内存 fallback | 不作为权威来源；可重建；不可覆盖 MySQL 真值 |
| 兼容来源 | 旧 `n9e`、旧 `catpaw`、旧 `/monitor/*`、旧配置字段 | Adapter/compat layer | 只做过渡读取和迁移，不作为新主契约 |

## 4. 统一数据源类型矩阵

| 类型 | 用户侧命名 | 权威存储 | 后端 Adapter | 前端结构来源 | 必须能力 |
| --- | --- | --- | --- | --- | --- |
| metrics-prometheus | 指标数据源 | MySQL data_sources + credential_refs | Basic Monitoring Adapter | 成熟基础监控数据源页 | URL、认证、Header、TLS、超时、测试连接、label/series/query、状态启停 |
| metrics-victoriametrics | 指标数据源 | MySQL data_sources + credential_refs | Basic Monitoring Adapter | 成熟基础监控数据源页 | PromQL 兼容、查询超时、错误脱敏、历史记录 |
| logs-loki | 日志数据源 | MySQL data_sources + credential_refs | Logs Adapter | 基础监控日志页/日志中心同源结构 | 字段、上下文、查询、权限 |
| logs-clickhouse | 日志存储 | MySQL data_sources + credential_refs | Logs Adapter | 日志中心源码 | 查询、聚合、字段筛选、Pipeline、Saved Views |
| traces-oap | 链路后端 | MySQL data_sources + credential_refs | APM Adapter | 链路监控源码 | GraphQL proxy、trace、topology、profiling、alarm、timeout/cancel |
| cmdb | CMDB 数据源 | MySQL + CMDB tables | CMDB Adapter | AutoOps/AIOps 源码 | 资产树、主机、分组、终端/监控弹窗 |
| agent-package-repo | Agent 包仓库 | MySQL package registry + object storage | FindX Agent Adapter | AutoOps/AIOps + Agent 长期计划 | 版本、平台、签名、安装脚本、回滚、数据到达验证 |
| vector-qdrant | 向量索引加速层 | MySQL 知识库权威表 + Qdrant collection | Knowledge Adapter | FindX 知识库增强 | 索引任务、重建、状态、BM25/RRF 降级、Evidence Chain 引用 |
| ai-model | AI 模型配置 | MySQL ai_settings/ai_providers + credential_refs | AI Config Adapter | 系统配置增强 | 模型、凭据引用、超时、审计、全局统一读取 |

## 5. 统一配置键建议

后续实现前必须冻结配置键，避免继续新增散落字段。

| 配置域 | 建议键 | 说明 |
| --- | --- | --- |
| MySQL | `storage.mysql.dsn_ref` 或保留 `mysql.dsn` bootstrap | 生产优先环境变量；文档只写 `<DB_DSN>` |
| Redis | `runtime.redis.addr/db/password_ref` | 替代旧 `n9e.redis_*` 作为新主配置 |
| 数据源中心 | `datasource.bootstrap[]` | 仅用于首启种子；启动后以 MySQL data_sources 为准 |
| 链路监控 | `adapters.apm.oap_base_url`、`adapters.apm.timeout_seconds` | 对接 OAP/Query Protocol；错误脱敏 |
| 日志中心 | `adapters.logs.base_url`、`adapters.logs.storage`、`adapters.logs.clickhouse_ref` | 对接日志 Adapter，不把外部 API 暴露给页面 |
| 向量数据库 | `knowledge.vector.provider=qdrant`、`knowledge.vector.qdrant_url`、`knowledge.vector.collection` | MySQL 为权威，Qdrant 可重建 |
| Agent 包仓库 | `findx_agents.package_repo`、`findx_agents.remote_install`、`findx_agents.signing` | 包仓库、签名、安装、回滚和审计 |
| AI 模型 | `ai_providers`、`ai_settings` | 全局 AI 能力只走系统配置 / AI 模型配置 |

## 6. API 契约门禁

任何代码实现触及以下内容必须标记 `API_CONTRACT_CHANGE` 或 `DATA_CHANGE`：

| 变更 | 标记 | 必须说明 |
| --- | --- | --- |
| 新增/修改数据源 API 字段 | `API_CONTRACT_CHANGE` | 类型、字段、错误码、权限、旧字段兼容、前端调用点 |
| 新增 data_sources/credentials 表或字段 | `DATA_CHANGE` | 迁移、回滚、唯一约束、敏感值加密/引用、审计 |
| 新增 APM/Logs Adapter API | `API_CONTRACT_CHANGE` | 不暴露外部 GraphQL/ClickHouse 原始契约；503 脱敏 |
| 新增 Qdrant 配置和索引任务 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | MySQL 权威、Qdrant 可重建、BM25 降级、索引状态 |
| 新增 Agent 包仓库/远程安装 API | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | 包签名、平台兼容、凭据引用、远程执行审计、回滚 |

## 7. 实现准入

| 门禁 | 状态 | 说明 |
| --- | --- | --- |
| 当前配置证据 | `READY` | `api/config.yaml`、`api/config.yaml.example`、`datasources.go`、`embedding/config.go` 已读取 |
| 成熟数据源源码证据 | `READY_PARTIAL` | 已在 P0 总矩阵记录数据源源码；页面级数据源矩阵仍需补 |
| 链路/日志配置证据 | `SOURCE_READY_DOM_PENDING` | 源码已存在，运行态 DOM 和 adapter 契约仍需补 |
| Qdrant/向量架构 | `PLANNED` | 长期计划已锁定 MySQL 权威 + Qdrant 加速 + BM25 降级 |
| 敏感信息策略 | `REQUIRED` | 凭据只引用不回显；日志、错误、文档不得出现真实 token/DSN/Cookie |
| 验证 | `REQUIRED` | 后端变更跑 Go test/build；前端变更跑 npm build；MCP 登录点击数据源和系统配置 |

未完成本矩阵前，不允许开始数据源、链路、日志、知识库向量或 Agent 包管理的代码实现。
