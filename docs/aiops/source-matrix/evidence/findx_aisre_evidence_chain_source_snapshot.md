# FindX AI SRE / Evidence Chain 源码检查证据

生成时间：2026-05-07 16:18（UTC+8）

状态：`P5-AISRE-EVIDENCE-CHAIN` 编码前源码证据。本文只记录当前 FindX 源码事实，不代表 AI SRE / Evidence Chain 已完成。

## 1. 本地源码入口

| 能力域 | 源码路径 | 当前证据 |
| --- | --- | --- |
| AI SRE 会话模型 | `api\internal\model\aiops.go` | `AIOpsSession`、`AIOpsMessage`、`ReasoningStep`、`DataSourceUsage`、`SuggestedAction`、`TopologyHighlight`、`AIOpsInspection` |
| AI SRE 会话 API | `api\internal\handler\aiops.go` | 会话创建、消息、WebSocket、建议动作、巡检、Prometheus 查询、Agent 查询、拓扑生成 |
| AI SRE 运行时 | `api\internal\handler\aiops_runtime.go`、`aiops_inspection_flow.go`、`aiops_inspection_metrics.go`、`aiops_match_render.go` | 诊断、巡检、拓扑、报告模式；生成推理步骤、数据源使用、建议动作和拓扑高亮 |
| AI SRE 输出护栏 | `api\internal\handler\aiops_guardrails.go` | 巡检报告补齐健康评分、异常层级、根因、处置建议和数据源段落 |
| 诊断引擎 | `api\internal\reasoning\engine.go` | 假设生成、工具验证、置信度阈值、Evidence、Treatment、Hypothesis |
| 诊断工具注册 | `api\internal\reasoning\tools.go` | `prometheus_query`、`check_alerts`、`check_diagnose_history`、`knowledge_search` |
| 工作流模型 | `api\internal\model\workflow.go`、`workflow_run_step.go` | `Workflow`、`WorkflowVersion`、`WorkflowRun`、`WorkflowRunStep` |
| 工作流引擎 | `api\internal\workflow\engine` | 超时、节点超时、重试、失败策略、事件流、StepRecorder、streaming |
| 内置工作流 | `api\internal\workflow\engine\builtin\*.yaml` | 诊断、巡检、复盘、时间线、知识增强、指标分析、日志分析、回滚等内置流程 |
| 知识库模型 | `api\internal\model\knowledge.go`、`document.go` | 诊断案例、反馈、Runbook、Runbook 执行、知识文档、分块字段 |
| 知识库索引 | `api\internal\knowledge\indexer.go` | 文档分块、案例/Runbook 同步、全量重建、加载到搜索引擎 |
| 搜索引擎 | `api\internal\embedding` | BM25、内存向量、API embedding、hybrid、RRF、reranker |
| AI 模型配置 | `api\internal\aiconfig\model.go`、`api\internal\model\ai_settings.go`、`settings.go` | 默认模型解析、AISetting、AIProvider |
| 前端 AI SRE 入口 | `web\src\views\AiopsWorkbench.vue` | 诊断、知识库、工作流、自动修复四个 section |
| 前端问诊工作台 | `web\src\views\Workbench.vue` | 会话列表、模式切换、WebSocket 状态、模型选择、消息区、推理链、数据源、建议动作、拓扑面板 |
| 前端 WebSocket | `web\src\composables\useAiopsWebSocket.js` | WS 连接、重连、推理步骤、诊断消息、拓扑更新、指标订阅、建议动作结果 |
| 前端工作流 | `web\src\views\WorkflowHub.vue`、`web\src\components\workflow` | 诊断工作流、工作流管理、画布、历史、内置参数 |
| 前端知识库 | `web\src\views\KnowledgeCenter.vue`、`KnowledgeBase.vue`、`web\src\components\knowledge` | 案例、文档、语义搜索、Runbook |
| 前端 AI 配置 | `web\src\views\AIConfig.vue`、`web\src\components\ai` | LLM、Embedding、Reranker 配置 |

## 2. 当前 API 路由证据

`api\main.go` 当前已有以下 AI SRE / Evidence Chain 相关入口：

| 路由 | 当前作用 | P5 结论 |
| --- | --- | --- |
| `POST /api/v1/aiops/sessions` | 创建 AI SRE 会话 | `SOURCE_PRESENT`，但需迁移用户侧命名为 AI SRE |
| `POST /api/v1/aiops/sessions/:id/messages` | 发送消息并返回诊断结果 | `SOURCE_PRESENT`，缺统一 Evidence Chain ID |
| `GET /api/v1/aiops/sessions/:id/messages` | 获取消息历史 | `SOURCE_PRESENT`，缺证据引用分页/过滤 |
| `GET /api/v1/aiops/ws/sessions/:id` | WebSocket 推理流 | `SOURCE_PRESENT`，需补协议版本、错误码、断线恢复验收 |
| `POST /api/v1/aiops/sessions/:id/actions/execute` | 执行建议动作 | `SOURCE_PRESENT`，当前限制 promql、copy command、link、topology；远程写操作必须另走审批和审计 |
| `POST /api/v1/aiops/inspections` | 创建巡检 | `SOURCE_PRESENT_PARTIAL`，当前是内存进度模型，未形成完整任务/证据链契约 |
| `GET /api/v1/aiops/inspections/:id/progress` | 获取巡检进度 | `SOURCE_PRESENT_PARTIAL` |
| `GET /api/v1/aiops/inspections/:id/report` | 获取巡检报告 | `SOURCE_PRESENT_PARTIAL` |
| `POST /api/v1/aiops/data/prometheus/query` | PromQL 查询 | `SOURCE_PRESENT`，后续需统一数据源权限、错误脱敏、查询记录 |
| `POST /api/v1/aiops/data/prometheus/query_range` | 区间查询 | `SOURCE_PRESENT`，当前复用即时查询处理，需要实现区间语义验收 |
| `POST /api/v1/aiops/data/catpaw/query` | 探针巡检查询 | `SOURCE_PRESENT_PARTIAL`，用户侧需改 FindX Agent 命名 |
| `POST /api/v1/aiops/topology/generate` | 生成拓扑 | `SOURCE_PRESENT_PARTIAL`，需对接 CMDB、APM、日志和 Agent 数据 |
| `POST /api/v1/diagnose`、`GET /api/v1/diagnose` | 诊断记录 | `SOURCE_PRESENT`，需成为 Evidence Chain 节点 |
| `GET /api/v1/audit/events` | 审计事件 | `SOURCE_PRESENT`，需统一证据引用 |
| `GET /api/v1/findx-agents`、`POST /register`、`POST /heartbeat` | FindX Agent 状态 | `SOURCE_PRESENT_PARTIAL`，需扩展安装、配置、数据到达、回滚 |
| `GET/POST/PUT/DELETE /api/v1/workflows` | 工作流定义 | `SOURCE_PRESENT`，需合并成熟事件流水线结构 |
| `POST /api/v1/workflows/:id/run`、`/stream` | 工作流执行与流式输出 | `SOURCE_PRESENT`，需把每步 Evidence Chain 化 |
| `GET /api/v1/workflows/:id/runs` | 工作流执行记录 | `SOURCE_PRESENT`，需补步骤详情、审批、回滚、证据引用 |
| `POST /api/v1/knowledge/search` | 知识检索 | `SOURCE_PRESENT`，当前搜索引擎是 BM25/内存向量/hybrid，不是 Qdrant |
| `POST /api/v1/knowledge/reindex-all` | 知识库重建 | `SOURCE_PRESENT`，后续需拆 MySQL 权威数据与 Qdrant 可重建索引状态 |
| `GET/PUT /api/v1/settings/ai`、`GET/POST /api/v1/ai-providers` | AI 配置 | `SOURCE_PRESENT`，后续所有 AI 能力必须统一引用 model_ref/provider_ref |

## 3. 当前状态流证据

| 状态流 | 当前源码证据 | 风险 |
| --- | --- | --- |
| 会话创建 | `AIOpsCreateSession` 保存 `ChatSession`，返回 `AIOpsSession` | 会话主体仍沿用 chat session 存储，缺 AI SRE 会话专表和证据链外键 |
| 消息发送 | `AIOpsPostMessage` 写用户消息，`runAIOpsDiagnosis` 生成助手消息 | 结果是消息级推理链，未拆成可审计 evidence item |
| WebSocket 推理流 | `AIOpsSessionWS` 支持 ping、interrupt、feedback、execute_action、subscribe_metrics、message | 协议无版本；浏览器端存在 `PLACEHOLDER_REST` 注释，后续实现必须消除占位和补测试 |
| 建议动作 | `executeAIOpsAction` 只允许 PromQL、复制命令、链接、拓扑 | 远程修复动作尚未进入审批、回滚和 Evidence Chain |
| 巡检报告 | `runAIOpsInspectionAnswer`、`ensureInspectionReportSections` | 当前可生成巡检摘要，但未绑定成熟 Agent 任务、数据到达和真实插件执行记录 |
| 推理引擎 | `DiagnosticEngine.Run` 假设验证循环 | `toolPrometheus` 当前说明为 metrics via workflow engine，仍是弱证据，需要真实 Adapter |
| 工作流执行 | Engine 超时、重试、事件、StepRecorder | 需要把节点输入/输出/错误/审批/回滚形成 Evidence Chain，不只保存 run step |
| 知识检索 | `HybridSearcher` provider 支持 builtin/api/hybrid，hybrid 用 RRF | 当前 `VectorEngine` 是进程内存 map，不满足内置 Qdrant 向量数据库要求 |
| AI 配置 | `AIConfig.vue`、`AIProvider`、`AISetting` | 模型密钥必须只保存在系统配置，其他页面只能引用，不得散落配置 |
| 品牌脱敏 | 当前前端/后端仍存在 `AIOps`、`AI WorkBench`、旧探针名等文本 | 用户侧必须统一为 FindX / AI SRE / FindX Agent；内部证据文档可保留来源名称 |

旧品牌扫描示例：

- `web\src\views\Workbench.vue` 当前存在 `AI WorkBench`、`AIOps` 等显示文本，后续实现必须迁移为 FindX / AI SRE 用户侧命名。
- `web\src\components\workbench\ReasoningBlock.vue` 当前存在旧探针查询标签，后续实现必须迁移为 FindX Agent 交叉验证等用户侧命名。
- `api\internal\handler\aiops_runtime.go`、`aiops_guardrails.go` 当前仍包含旧探针和旧平台提示文本，后续 API 响应、审计对象、页面标题不得暴露外部来源品牌。

## 4. 证据链缺口

当前源码已经有“推理链”和“数据源使用”的结构，但还不是完整 Evidence Chain。进入 P5 编码前必须补齐以下实体和状态：

| 缺口 | 必须补齐 |
| --- | --- |
| Evidence Chain 主体 | `evidence_chain_id`、scope、tenant/team、subject、incident/run/session 关联、created_by、status、retention |
| Evidence Item | source_type、source_id、source_ref、query/action、input_digest、output_ref、timestamp、confidence、sanitized_error、permission_context |
| 数据来源覆盖 | metrics、logs、trace/span、topology、profiling、alerts、CMDB、Agent、workflow、runbook、knowledge、audit、manual note |
| 不可用状态 | `DATA_MISSING`、`BLOCKED_*`、`QUERY_TIMEOUT`、`PERMISSION_DENIED`、`COMPONENT_UNAVAILABLE` |
| 防编造 | AI SRE 只能引用已登记证据；证据不足时明确返回数据缺失和下一步采集建议 |
| 引用可追溯 | 查询语句、时间范围、数据源 ID、版本、配置版本、模型配置引用、知识 chunk ID |
| 安全脱敏 | 不保存真实 token、cookie、Bearer、完整 DSN、SSH 私钥、会话 ID；输出只保留引用和摘要 |

## 5. 知识库与向量索引证据

当前源码支持：

- MySQL/MariaDB 或内存 fallback 的文档、案例、Runbook 权威存储。
- 文档按类型和格式分块。
- BM25。
- API embedding。
- 内存向量检索。
- BM25 + 向量 + RRF hybrid。
- 可选 reranker。

长期目标要求：

- MySQL/MariaDB 是知识库权威基石。
- Qdrant 是内置本地 sidecar 向量索引加速层，可删除重建，不作为权威数据。
- BM25 是关键词检索和 Qdrant 不可用时的降级路径。
- RRF 是混合召回融合路径。
- Evidence Chain 必须记录知识引用的 document_id、chunk_id、chunk_index、source_id、doc_version、recall_method、score、embedding_model_ref、reranker_ref。

当前阻断：

| 项目 | 状态 | 说明 |
| --- | --- | --- |
| MySQL/MariaDB 权威存储 | `SOURCE_PRESENT` | 已有 store/document/case/runbook 模型，需补明确迁移和索引状态 |
| BM25 | `SOURCE_PRESENT` | 可作为 fallback |
| RRF hybrid | `SOURCE_PRESENT` | 当前在进程内完成融合 |
| Qdrant sidecar | `BLOCKED_QDRANT_NOT_IMPLEMENTED` | 未见 Qdrant client、collection、upsert/search/delete/rebuild/status 实现 |
| 向量索引任务状态 | `BLOCKED_INDEX_STATUS_NOT_IMPLEMENTED` | 需要索引任务、进度、失败重试、可重建状态 UI |
| 知识引用入链 | `BLOCKED_KNOWLEDGE_REF_CHAIN_NOT_IMPLEMENTED` | 当前搜索结果未形成标准 evidence item |

## 6. 后续验收证据要求

P5 进入实现后必须自己执行并留证：

- 前端 lint/build。
- 后端测试/build。
- FindX 自有登录。
- AI SRE 诊断会话正常路径、数据缺失路径、权限路径、WebSocket 断线重连。
- 推理链展开、数据源引用、建议动作、拓扑联动、复制交接、反馈、归档案例。
- 工作流列表、执行、流式输出、步骤详情、失败节点、重试、审批、回滚。
- 巡检会话、诊断记录、复盘报告三者必须是不同页面结构和不同状态流。
- 知识检索 BM25、Qdrant unavailable fallback、hybrid、badcase、reindex。
- Evidence Chain 查询、详情、过滤、导出、权限、脱敏。
- 菜单、页面标题、权限对象、审计对象不出现外部来源品牌。
- 组件不可用必须显示 `BLOCKED` 或脱敏 503，不得假成功。
