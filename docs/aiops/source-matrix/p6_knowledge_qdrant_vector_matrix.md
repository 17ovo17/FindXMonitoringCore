# FindX P6 知识库 / Qdrant 向量索引同源矩阵

生成时间：2026-05-07 16:46（UTC+8）
状态：`P6-KNOWLEDGE-QDRANT-VECTOR` 编码前门禁，不代表 FindX 已完成 Qdrant 接入

## 1. 结论

FindX 知识库长期架构固定为：MySQL/MariaDB 做权威基石，Qdrant 本地 sidecar 做内置向量索引加速层，BM25 做关键词检索和 Qdrant 不可用降级，RRF 做混合召回融合，Reranker 做可选重排。当前源码已经具备文档分块、MySQL/内存存储、BM25、API embedding、内存向量、RRF、reranker、badcase 和搜索统计，但 Qdrant 还没有实现。

源码证据：

- [FindX 知识库 / Qdrant 向量索引源码检查证据](evidence/findx_knowledge_qdrant_source_snapshot.md)
- [P0 统一配置与数据源契约矩阵](p0_config_datasource_contract_matrix.md)
- [P5 AI SRE / Evidence Chain 同源矩阵](p5_aisre_evidence_chain_matrix.md)

执行硬约束：

- 不把进程内存向量 `VectorEngine` 当作最终内置向量数据库。
- 不把 Qdrant 做成权威数据源；Qdrant 只保存可重建向量索引。
- 用户侧页面、菜单、权限对象和审计对象统一显示“内置向量索引”“语义检索加速”“索引健康”；Qdrant 只保留在内部源码矩阵、配置键、合规登记和开发证据中。
- 不删 BM25，不删 RRF，不删 badcase，不删 chunk context。
- 不做 MVP、最小实现、最小确认、最小验证。
- 不做静态索引状态、静态成功态或假向量结果。
- 后续实现必须自己跑 lint、build、后端测试、MCP 浏览器真实登录和点击回归，并按报错修复后重跑。

## 2. 目标架构

| 层级 | 组件 | 角色 | 完成定义 |
| --- | --- | --- | --- |
| 权威数据 | MySQL/MariaDB | 文档、chunk、案例、Runbook、badcase、search event、index job、权限、版本 | 数据可备份、可迁移、可审计；Qdrant 可由它重建 |
| 向量加速 | Qdrant sidecar | collection、point、payload、vector search、snapshot、health | 可配置、可重建、可降级、可观测 |
| 关键词检索 | BM25 | 中文/英文关键词召回、Qdrant 不可用 fallback | Qdrant 失败不影响基础检索 |
| 混合召回 | RRF | BM25 + Qdrant 融合排序 | 每条结果记录 recall_method 和分数 |
| 重排 | Reranker | 可选语义重排 | 引用系统 AI 配置，不散落密钥 |
| 证据链 | Evidence Chain | 搜索和引用可追溯 | document、chunk、query、score、model_ref、index_job_ref 入链 |

## 3. API / 数据契约门禁

进入代码实现前必须先设计并打标：

| 契约 | 标记 | 必须字段 |
| --- | --- | --- |
| VectorStore 抽象 | `API_CONTRACT_CHANGE` | `EnsureCollection`、`Upsert`、`Delete`、`Search`、`Health`、`Rebuild`、`Stats`、timeout、error model |
| Qdrant 配置 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | base_url、collection、api_key_ref、dimension、distance、timeout、tls、health_status、last_checked_at |
| 向量索引记录 | `DATA_CHANGE` | document_id、chunk_id、embedding_model_ref、vector_provider、collection、point_id、payload_hash、indexed_at、status、last_error |
| 索引任务 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | job_id、scope、status、total、done、failed、started_at、finished_at、retry_count、cancelled_by、audit_id |
| 搜索响应 | `API_CONTRACT_CHANGE` | query、items、engine、recall_method、bm25_score、vector_score、rrf_score、reranker_score、fallback_reason、evidence_refs、permission_context、redaction_policy |
| Badcase 反馈 | `DATA_CHANGE` | query、doc_id、chunk_id、reason、created_by、engine、recall_method、score_snapshot |
| Evidence Chain 引用 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | evidence_item_id、document_id、chunk_id、query_digest、model_ref、index_job_ref、permission_context、redaction_policy |

错误模型必须统一：

- Qdrant 未配置：503 + `BLOCKED_QDRANT_NOT_CONFIGURED`
- Qdrant 不可达：503 + `QDRANT_UNAVAILABLE`
- Collection 缺失：503 + `QDRANT_COLLECTION_MISSING`
- Upsert 失败：503 + `QDRANT_UPSERT_FAILED`
- Search 超时：503 + `QDRANT_SEARCH_TIMEOUT`
- Embedding API 不可用：503 + `EMBEDDING_PROVIDER_UNAVAILABLE`
- 权限不足：403
- 空结果：正常空态，不伪造结果
- Qdrant 失败但 BM25 可用：200 + `fallback_reason=qdrant_unavailable`

## 4. 当前源码到目标能力映射

| 当前能力 | 当前路径 | 目标处理 |
| --- | --- | --- |
| 文档权威存储 | `store\documents.go`、`model\document.go` | 保留并扩展 chunk version/hash/permission/index status |
| 文档分块 | `knowledge\chunker*.go`、`indexer.go` | 继续作为 Qdrant point payload 来源 |
| 搜索入口 | `handler\documents.go` `SearchKnowledge` | 扩展 engine、recall_method、fallback_reason、evidence_refs |
| BM25 | `embedding\bm25.go` | 保留为 fallback 和 RRF 一路 |
| 内存向量 | `embedding\vector.go` | 仅作为开发 fallback 或测试 provider，不作为生产默认 |
| RRF | `embedding\search.go` | 继续保留，结果需暴露 bm25/vector/rrf score |
| Reranker | `embedding\reranker.go` | 继续保留，补 reranker_score 和 model_ref |
| 配置 | `embedding\config.go`、`EmbeddingConfig.vue` | 增加 Qdrant sidecar 配置、health、collection、索引策略 |
| Reindex | `knowledge.RebuildAll`、`ReindexAllHandler` | 改为有 job id、进度、失败重试、取消、审计；任务状态写入 MySQL/MariaDB，不写成只存在于 Qdrant payload |
| UI 搜索 | `SemanticSearch.vue` | 补引擎状态、fallback 提示、引用链、索引健康 |
| UI 文档管理 | `DocManager.vue` | 补索引状态、重建进度、失败原因、重试 |

## 5. 页面结构门禁

| 页面 | 必须结构 | 真实动作 |
| --- | --- | --- |
| 文档管理 | 上传、类型/分类/关键词筛选、表格、详情、删除、索引状态、错误态、分页 | 上传、删除、单文档重建、查看 chunk、查看索引状态 |
| 语义搜索 | 查询框、类型、topK、引擎状态、结果、score、chunk context、fallback 提示 | 搜索、badcase、查看引用、跳文档、重新搜索 |
| 索引任务 | 任务列表、scope、进度、失败数、耗时、错误、审计、操作列 | 新建重建、取消、重试、查看失败 chunk |
| 向量索引健康 | Qdrant 状态、collection、dimension、point count、last sync、BM25 状态 | 测试连接、ensure collection、重建、降级切换 |
| AI 配置 | LLM、Embedding、Reranker、Qdrant、健康 | 保存、测试、设默认、禁用、删除、掩码回显 |
| Evidence Chain 引用 | search query、document/chunk、score、model_ref、index_job | 查看证据、导出、权限过滤 |

所有按钮必须接入真实状态流；未实现时显示 `BLOCKED`，不得展示为可用成功态。

## 6. 与 AI SRE / Evidence Chain 的关系

| AI SRE 场景 | 知识库要求 |
| --- | --- |
| 诊断会话引用知识 | 记录 document_id、chunk_id、recall_method、score、model_ref、query_digest |
| 巡检报告引用 Runbook | 记录 runbook_id、version、chunk_id、execution/ref |
| 复盘报告引用案例 | 记录 case_id、source_diagnosis_id、评价、权限 |
| 证据不足 | 返回 `DATA_MISSING` 或 fallback 说明，不编造知识依据 |
| Qdrant 不可用 | 显示 BM25 fallback，Evidence Chain 记录 fallback_reason |
| Badcase | 进入搜索质量统计和后续重建/调优任务 |

## 7. MCP、lint、构建验收

P6 代码实现后必须由执行者自己完成：

- `cd /opt/ai-workbench/web && npm run lint` 或项目等价 lint 命令。
- `cd /opt/ai-workbench/web && npm run build`。
- `cd /opt/ai-workbench/api && go test -count=1 ./...`。
- `cd /opt/ai-workbench/api && go build -o api-linux .`。
- MCP 浏览器打开 FindX 自有登录页并真实登录。
- MCP 浏览器覆盖文档上传、列表、详情、删除、单文档重建、全量重建、索引任务、语义搜索、badcase、AI 配置、Qdrant 不可用降级。
- 正常、异常、空态、权限、组件不可用、窄屏回归。
- 敏感信息扫描、乱码扫描、外部品牌用户侧扫描。

API 测试不能替代浏览器真实点击；静态检查通过不能标 PASS。发现 lint/build/test/MCP 报错时必须修复后重跑。

本次 P6 文档闭环未执行 MCP 浏览器回归，因为它是编码前源码矩阵，不是功能实现。后续任何 P6 实现切片必须补运行态 DOM / MCP 浏览器证据后才能声明完成。

本文件即使通过文档 QA，也只代表编码前门禁 PASS，不代表 UI、API、Qdrant、索引任务或浏览器功能验收 PASS。

## 8. 阻断项

- 把内存向量当作最终向量数据库：FAIL。
- Qdrant 未实现却声称内置向量数据库完成：FAIL。
- Qdrant 成为权威数据源，MySQL 不再能重建索引：FAIL。
- 删除 BM25、RRF、badcase、chunk context：FAIL。
- 重建索引没有任务状态、失败重试、取消和审计：FAIL。
- 搜索结果没有 recall_method、score、chunk 引用和 Evidence Chain 引用：FAIL。
- 配置页保存或回显真实 API Key：FAIL。
- 未自己跑 lint/build/test/MCP 浏览器真实点击却标 PASS：FAIL。
