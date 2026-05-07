# FindX P5 AI SRE / Evidence Chain 同源矩阵

生成时间：2026-05-07 16:18（UTC+8）
状态：`P5-AISRE-EVIDENCE-CHAIN` 编码前门禁，不代表 FindX 已完成实现

## 1. 结论

AI SRE 不能只做聊天窗口、静态诊断报告或复盘页面复制。FindX 必须把指标、日志、Trace、拓扑、告警、CMDB、Agent、巡检、工作流、Runbook、知识库、审计、人工确认全部结构化进入 Evidence Chain；AI SRE 只能基于真实证据诊断、解释、编排和复盘，证据缺失时必须返回数据缺失或 `BLOCKED`，不得编造根因。

源码证据：

- [FindX AI SRE / Evidence Chain 源码检查证据](evidence/findx_aisre_evidence_chain_source_snapshot.md)
- [P3 链路监控同源矩阵](p3_skywalking_apm_real_matrix.md)
- [P3 日志中心同源矩阵](p3_signoz_logs_real_matrix.md)
- [P4 Agent Suite 同源矩阵](p4_categraf_catpaw_agent_suite_matrix.md)

当前阻断：

- FindX 当前已有 AI SRE 会话、推理链、工作流、知识库、Runbook、AI 配置和部分 Agent 状态源码。
- 当前还没有统一 Evidence Chain 主体、Evidence Item、证据引用、证据权限、证据保留、Qdrant sidecar 和完整 UI。
- 当前问诊、巡检、复盘、工作流、知识检索仍存在自研弱化结构和旧命名风险，进入实现时必须改为 FindX 品牌和成熟状态流。

执行硬约束：

- 不 iframe。
- 不嵌入参考站或既有部署系统。
- 不接入参考站 SSO。
- 不做 MVP、最小实现、最小确认或最小验证。
- 不用静态假按钮、静态假数据、静态成功态冒充成熟功能。
- 不把 API 测试、静态检查或截图相似度当作页面验收。
- 后续实现必须自己跑 lint、build、后端测试、MCP 浏览器真实登录和点击回归，并按报错修复后重跑。

## 2. FindX 页面规划

| FindX 页面 | 当前源码入口 | 成熟/同源要求 |
| --- | --- | --- |
| AI SRE 诊断会话 | `web\src\views\Workbench.vue`、`api\internal\handler\aiops*.go` | 会话列表、模式、消息、推理链、证据、数据源、建议动作、拓扑、反馈、归档、WS 断线恢复 |
| 巡检会话 | `AiopsWorkbench.vue?section=diagnosis`、`aiops_inspection_*` | 巡检对象、层级、指标、进程、Agent 状态、执行进度、报告、失败态 |
| 工作流 | `WorkflowHub.vue`、`components\workflow`、`api\internal\workflow` | 工作流列表、编辑、版本、授权团队、执行记录、调试、步骤、失败重试、审批、回滚；合并成熟 event-pipeline 结构 |
| 自动修复 | `AiopsWorkbench.vue?section=remediation` | 不能空壳；必须有审批、计划、执行、回滚、影响范围、证据和审计 |
| 复盘报告 | `incident_review.yaml`、`incident_postmortem.yaml`、诊断记录 | 与诊断会话不同结构：时间线、影响面、证据列表、处置动作、责任归因、后续项、导出 |
| Evidence Chain | 待新增 | 链列表、详情、source filter、query/action、证据项、引用对象、权限、保留、导出 |
| 知识库 | `KnowledgeCenter.vue`、`KnowledgeBase.vue`、`embedding`、`knowledge` | MySQL 权威存储、Qdrant sidecar、BM25 fallback、RRF、索引状态、chunk 引用入链 |
| AI 模型配置 | `AIConfig.vue`、`api\internal\aiconfig` | LLM、Embedding、Reranker、凭据引用、模型健康、超时、审计；所有 AI 能力统一读取 |

用户侧统一显示 FindX、AI SRE、FindX Agent、链路监控、日志中心、证据链。外部来源名称只允许出现在本源码矩阵、合规登记、归档和开发证据中。

## 3. Evidence Chain 数据来源门禁

| 来源 | 必须证据项 | 上游矩阵 |
| --- | --- | --- |
| 指标 | datasource_id、PromQL、时间范围、query type、结果摘要、图表引用、错误码 | P0 指标查询、P0 数据源 |
| 日志 | query、字段筛选、时间范围、log row id、上下文、Pipeline 版本、Saved View | P3 日志中心 |
| Trace / Span | trace_id、span_id、service、instance、endpoint、duration、error、attached logs | P3 链路监控 |
| 拓扑 | service/instance/node、edge、critical path、依赖方向、生成来源 | P3 链路监控、P2 CMDB |
| Profiling | task id、language、agent capability、status、flame graph/result ref | P3 链路监控、P4 Agent |
| 告警 | rule id、event id、severity、labels、annotations、状态、处理动作 | P1 告警通知 |
| CMDB | business、host、service、owner、team、tags、change ref、relation | P2 CMDB |
| FindX Agent | heartbeat、package、config version、install task、data arrival、upgrade/rollback | P0 SkyWalking Agent、P4 Agent Suite |
| 巡检诊断 | plugin、target、params、session、tool output、report、selftest | P4 Agent Suite |
| 工作流 | workflow id、version、run id、step id、inputs/outputs、approval、rollback | 本矩阵 + P1 工作流并入 |
| Runbook | runbook id、version、execution id、variables、output、rollback steps | 本矩阵 |
| 知识库 | document id、chunk id、source id、score、recall method、model ref | 本矩阵 |
| 审计 | actor、permission、operation、object、request id、sanitized error | 平台治理 |
| 人工确认 | reviewer、decision、comment、time、scope、evidence refs | 本矩阵 |

任何来源未接通时，AI SRE 必须在 UI 和 API 中暴露明确 `DATA_MISSING` / `BLOCKED_*` / `COMPONENT_UNAVAILABLE`，不能把缺失证据当作正常结果。

## 4. 状态流门禁

| 状态流 | 当前源码 | P5 要求 |
| --- | --- | --- |
| 会话生命周期 | `AIOpsSession` + `ChatSession` | active、archived、closed、blocked；支持会话 scope、证据链、模型配置引用 |
| 推理链 | `ReasoningStep` | 每一步必须有输入引用、工具、输出引用、状态、耗时、错误、证据 refs，不只展示文本 |
| 数据源使用 | `DataSourceUsage` | source 必须映射统一数据源中心和权限上下文；错误脱敏 |
| 建议动作 | `SuggestedAction` | read-only、approval-required、execute、rollback、external-link 分级；远程写动作必须审批和审计 |
| WebSocket | `useAiopsWebSocket.js`、`AIOpsSessionWS` | 协议版本、断线恢复、心跳、取消、错误码、重放、订阅清理 |
| 巡检 | `AIOpsInspection` | 目标、层级、任务、进度、失败、报告、证据 refs、Agent 数据到达 |
| 工作流执行 | `WorkflowRun`、`WorkflowRunStep`、Engine events | 列表、详情、步骤、节点日志、审批、重试、回滚、证据 refs |
| 复盘报告 | builtin `incident_*` workflow | 时间线、影响面、根因候选、处置动作、验证结果、后续项，不得和诊断会话同页同结构 |
| 知识检索 | `HybridSearcher` | BM25、Qdrant、hybrid、RRF、reranker、badcase、索引状态、引用入链 |
| AI 模型配置 | `AIConfig.vue`、`AIProvider` | 所有 AI 调用只引用 model_ref/provider_ref，不在功能页保存或回显密钥 |

## 5. API / 数据契约门禁

进入 P5 编码前必须先打标并设计以下契约：

| 契约 | 标记 | 必须字段 |
| --- | --- | --- |
| Evidence Chain | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | chain_id、tenant_id、team_id、scope、trigger_type、trigger_ref、status、severity、owner、created_by、retention_until |
| Evidence Item | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | item_id、chain_id、source_type、source_id、source_ref、time_range、query_or_action、input_digest、output_ref、summary、confidence、status、sanitized_error |
| Reasoning Step | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | step_id、chain_id、session_id、sequence、tool、input_refs、output_refs、evidence_refs、status、latency_ms、model_ref、permission_context、redaction_policy |
| AI SRE Session | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | session_id、mode、scope、chain_id、audience、status、model_ref、created_by、last_message_at |
| Workflow Evidence | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | workflow_id、version、run_id、step_id、approval_id、rollback_id、evidence_refs、audit_id |
| Runbook Evidence | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | runbook_id、version、execution_id、target_scope、variables_ref、output_ref、rollback_ref、evidence_refs |
| Knowledge Vector Index | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | collection、document_id、chunk_id、embedding_model_ref、vector_status、indexed_at、last_error、qdrant_point_id |
| Knowledge Reference | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | document_id、chunk_id、source_id、doc_version、recall_method、score、reranker_score、evidence_item_id、permission_context |
| Model Call Audit | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | model_ref、provider_ref、purpose、prompt_digest、tokens、latency_ms、status、sanitized_error、chain_id、redaction_policy |

凭据只能通过 `credential_ref`、`model_ref`、`provider_ref`、`<TOKEN>`、`<DB_DSN>`、`<API_KEY>` 等引用或占位符表达。

## 6. 知识库向量架构

FindX 知识库长期固定为：

| 层级 | 角色 | 完成定义 |
| --- | --- | --- |
| MySQL/MariaDB | 权威基石 | 文档、chunk、case、runbook、badcase、版本、权限、索引状态权威存储 |
| Qdrant sidecar | 内置向量数据库加速层 | collection 管理、upsert、delete、search、rebuild、health、snapshot、数据可重建 |
| BM25 | 关键词检索和降级 | Qdrant 不可用时仍可检索；结果标记 fallback |
| RRF | 混合召回 | BM25 和 Qdrant 结果融合；记录 recall_method 和 score |
| Reranker | 可选重排 | 引用模型配置，不散落密钥 |

当前源码内存向量只能作为开发期能力证据，不能作为最终内置向量数据库交付。

阻断状态：

- `BLOCKED_QDRANT_NOT_IMPLEMENTED`
- `BLOCKED_INDEX_STATUS_NOT_IMPLEMENTED`
- `BLOCKED_KNOWLEDGE_REF_CHAIN_NOT_IMPLEMENTED`

## 7. 页面结构门禁

| 页面 | 必须结构 | 真实动作 |
| --- | --- | --- |
| AI SRE 诊断会话 | 会话 rail、模式、scope、消息、推理链、证据、数据源、建议动作、拓扑、反馈 | 新建、重命名、删除、发送、取消、重试、执行建议动作、归档案例 |
| 巡检会话 | 巡检目标、层级、Agent 状态、执行进度、指标/进程/日志/Trace 证据、报告 | 开始、暂停/取消、重试、查看证据、生成报告 |
| 工作流 | 列表、授权团队、版本、编辑、调试、执行记录、步骤详情、审批、回滚 | 新建、编辑、调试、运行、重试、回滚、查看证据 |
| 复盘报告 | 时间线、影响范围、证据项、根因候选、处置动作、验证结果、后续项、导出 | 生成、编辑、引用证据、导出、归档 |
| Evidence Chain | 链列表、source filter、状态、详情、证据项、对象跳转、权限、保留 | 查询、过滤、导出、查看原始摘要、补充人工确认 |
| 知识库索引 | 文档、chunk、索引状态、BM25/Qdrant 状态、重建进度、badcase | 上传、重建、搜索、标注 badcase、查看引用 |
| AI 配置 | LLM、Embedding、Reranker、健康、超时、审计 | 新增、编辑、测试、设默认、禁用、删除 |

按钮、下拉、时间、搜索、联想、历史、导出、重试、回滚、审批、归档、复制等控件有真实语义时必须接入真实状态流；未实现只能显示 `BLOCKED`，不得假成功。

## 8. 工作流与成熟事件流水线合并

工作流入口归入 FindX AI SRE，但内部结构不能被弱化：

- 列表必须保留名称、备注、授权团队、更新人、更新时间、操作列。
- 编辑必须保留节点、参数、调试、执行记录、授权、发布/版本。
- 执行记录必须保留状态、耗时、输入、输出、错误、步骤、重试、回滚和证据引用。
- 事件流水线、诊断工作流、自动修复工作流不能混成一个静态卡片页。
- 所有 AI 节点必须只引用 `系统配置 / AI 模型配置` 的模型配置。
- 工作流运行、节点输出、人工审批、回滚动作必须进入 Evidence Chain。

## 9. 品牌与脱敏规则

- 用户侧统一使用 FindX、AI SRE、FindX Agent、链路监控、日志中心、证据链。
- 旧来源产品名只允许保留在源码矩阵、合规、归档和开发证据。
- 当前源码中 `AI WorkBench`、旧探针名、旧来源名需要在实现阶段统一收敛。
- `AIOps` 作为内部包名或开发术语可留在源码迁移过程中；不得出现在用户侧菜单、页面标题、权限对象、审计对象或最终产品文案中。
- 审计对象、权限对象、页面标题、菜单名不得出现外部来源品牌。
- 错误提示不得泄露 SQL、内部路径、token、Cookie、Bearer、完整 DSN、SSH 私钥或会话 ID。

## 10. MCP、lint、构建验收

P5 代码实现后必须由执行者自己完成：

- `cd /opt/ai-workbench/web && npm run lint` 或项目等价 lint 命令。
- `cd /opt/ai-workbench/web && npm run build`。
- `cd /opt/ai-workbench/api && go test -count=1 ./...`。
- `cd /opt/ai-workbench/api && go build -o api-linux .`。
- MCP 浏览器打开 FindX 自有登录页并真实登录。
- MCP 浏览器逐项点击 AI SRE 诊断会话、巡检会话、工作流、复盘报告、Evidence Chain、知识库、AI 模型配置。
- 正常、异常、空态、权限、组件不可用、窄屏回归。
- 敏感信息扫描、乱码扫描、外部品牌用户侧扫描。

API 测试不能替代浏览器真实点击；静态检查通过不能标 PASS。发现前端报错、lint 报错、构建失败、浏览器报错时，必须修复后重跑，不能把错误留给用户。

本次 P5 文档闭环未执行 MCP 浏览器回归，因为它是编码前源码矩阵，不是功能实现。后续任何 P5 实现切片必须补运行态 DOM / MCP 浏览器证据后才能声明完成。

## 11. 阻断项

- AI SRE 只做聊天窗口，没有 Evidence Chain：FAIL。
- 诊断会话、巡检会话、复盘报告页面结构一样：FAIL。
- 工作流丢失成熟事件流水线结构、执行记录、授权团队、调试、步骤详情：FAIL。
- AI 使用不走系统配置的 AI 模型配置：FAIL。
- 证据不足仍输出确定根因：FAIL。
- Qdrant 未实现却声称内置向量数据库完成：FAIL。
- 内存向量被当作最终向量数据库：FAIL。
- 建议动作、回滚、审批、导出、归档、搜索、重建等按钮静态展示：FAIL。
- 用户侧出现外部来源品牌：FAIL。
- 未自己跑 lint/build/test/MCP 浏览器真实点击却标 PASS：FAIL。
