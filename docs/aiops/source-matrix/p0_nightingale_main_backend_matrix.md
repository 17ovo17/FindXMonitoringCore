# P0 Nightingale 主仓库后端事实源矩阵

更新时间：2026-05-10 03:15（UTC+8）

## 结论

`D:\项目迁移文件\平台源码\nightingale-main (1)\nightingale-main` 是独立的 Nightingale 主仓库事实源，不能被 `fe-main` 前端源码替代。

本矩阵用于补齐基础监控、告警通知、内置资产和事件流水线的后端契约、状态流和服务端行为证据。React-only 页面迁移仍以 `D:\项目迁移文件\平台源码\fe-main` 作为 UI 结构和交互事实源；后端 API、权限、内置资产初始化、告警生命周期和事件流水线必须同时回查本矩阵。

## 源码入口

| 能力面 | 当前源码路径 | 复核结论 |
| --- | --- | --- |
| 主仓库定位 | `README.md` | Nightingale 是以告警、告警处理和分发为核心的监控项目，连接已有数据源，推荐 Categraf 通过 Remote Write 接入 |
| Go 主工程 | `go.mod` | 模块为 `github.com/ccfos/nightingale/v6`，属于完整后端主仓库 |
| API 总路由 | `center/router/router.go` | 覆盖数据源、查询、用户/团队/业务组、目标、仪表盘、告警、通知、事件流水线、内置组件、Saved View、AI Agent、LLM、Skill、MCP 等 |
| 内置资产路由 | `center/router/router_builtin.go` | 提供内置 dashboard、alert、icon、favorite 等读取和管理入口 |
| 集成初始化 | `center/integration/init.go` | 从 `integrations` 目录初始化 builtin component、icon、markdown、alerts、dashboards、metrics 等 payload |
| 事件流水线契约 | `doc/api/event-pipeline.md` | 覆盖 Pipeline CRUD、tryrun、API trigger、SSE stream、executions、execution stats 和 service API |
| 内嵌前端资源 | `front/statik/statik.go` | 仅证明主仓库具备静态资源嵌入机制，不作为 React UI 同源迁移主证据 |
| 数据源适配 | `datasource/**` | 覆盖 Prometheus、MySQL、PostgreSQL、ES/OpenSearch、TDengine、ClickHouse/Doris 等适配方向 |
| Prometheus 兼容 | `prom/**` | Prom client、reader、option 等查询与兼容层事实源 |
| 告警引擎 | `alert/**` | eval、record、dispatch、mute、queue、sender、pipeline 等告警生命周期事实源 |
| 存储和总线 | `storage/**` | Redis、Pub/Sub、存储抽象事实源 |
| 内置集成资产 | `integrations/**` | Linux、Windows、Kubernetes、MySQL、Redis、Nginx、Prometheus、VictoriaMetrics、ClickHouse、Elasticsearch、Kafka、Docker、SNMP、NVIDIA 等组件目录 |

## 与 `fe-main` 的边界

| 维度 | `fe-main` | `nightingale-main (1)` |
| --- | --- | --- |
| 主要用途 | React UI 结构、菜单、页面组件、表单、抽屉、状态展示、交互语义 | 后端契约、权限点、服务 API、状态机、内置资产初始化、告警生命周期 |
| 迁移用途 | FindX React-only 页面同源迁移事实源 | FindX 后端契约、Adapter、BLOCKED_BY_CONTRACT 条件和状态流事实源 |
| 不能替代的内容 | 不能证明完整后端状态流 | 不能替代 React 页面源码和交互布局 |
| 共同要求 | 页面按钮有真实动作时，必须同时对照 UI 源码和主仓库 API/状态流；契约缺失时显示 `BLOCKED_BY_CONTRACT` |

## 必须纳入的状态流

| 状态流 | 主仓库证据 | FindX 映射要求 |
| --- | --- | --- |
| 数据源与查询 | `datasource/**`、`prom/**`、`center/router/router.go` | 数据源插件、启停、测试、查询、日志/SQL/PromQL/ES/OpenSearch 元数据查询不能只照 UI 表单实现 |
| 告警生命周期 | `alert/**`、告警相关 router | alert rule / recording rule 到 current event、history event、event detail、notification record 的闭环必须保留 |
| 通知链路 | notification 相关 router、sender/queue 模块 | 通知媒介、通知规则、模板、试发、记录和失败队列需要真实契约或 `BLOCKED_BY_CONTRACT` |
| 事件流水线 | `doc/api/event-pipeline.md`、`alert/pipeline/**` | Pipeline CRUD、processor tryrun、API trigger、SSE stream、execution record、stats 不能弱化成静态页 |
| 内置资产 | `center/integration/init.go`、`center/router/router_builtin.go`、`integrations/**` | 模板中心、集成中心、内置 dashboard/alert/collect 导入必须按 payload 类型和初始化语义实现 |
| AI/MCP/Skill 补充 | `router_ai_agent.go`、`router_ai_llm_config.go`、`router_ai_skill.go`、`router_mcp_server.go` | 仅作为 FindX AI SRE 配置面补充证据，不替代真实指标、日志、Trace、CMDB 和 Agent 数据 |

## 验收影响

- 后续基础监控 P0/P1 切片不能只引用 `fe-main` 作为 Nightingale 全量事实源。
- 涉及数据源、查询、告警、通知、内置资产、事件流水线、AI/MCP 配置面的后端契约时，必须同时引用 `nightingale-main (1)`。
- `dify-main`、`FastGPT-main`、`ragflow-main` 仍是 AI/RAG/知识库候选源，未完成独立矩阵前不得标记 `SOURCE_PRESENT`。
- `prometheus`、`prometheus-3.11.2.windows-amd64` 是运行工具/发行包；`mysql` 是运行态数据目录；空目录是杂项候选，均不得作为源码矩阵 PASS。

