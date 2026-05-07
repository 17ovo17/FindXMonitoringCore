# FindX 知识库 / Qdrant 向量索引源码检查证据

生成时间：2026-05-07 16:46（UTC+8）

状态：`P6-KNOWLEDGE-QDRANT-VECTOR` 编码前源码证据。本文只记录当前 FindX 源码事实，不代表 Qdrant 或知识库向量加速层已完成。

## 1. 本地源码入口

| 能力域 | 源码路径 | 当前证据 |
| --- | --- | --- |
| 知识文档模型 | `api\internal\model\document.go` | `KnowledgeDocument` 包含 doc_type、file_type、source_id、embedding_model、chunk_index、parent_id |
| 知识搜索质量模型 | `api\internal\model\knowledge_search.go` | `KnowledgeSearchEvent`、`KnowledgeSearchBadcase`、`KnowledgeSearchStats` |
| 案例与 Runbook 模型 | `api\internal\model\knowledge.go` | `DiagnosisCase`、`DiagnosisFeedback`、`Runbook`、`RunbookExecution` |
| 文档存储 | `api\internal\store\documents.go` | `knowledge_documents` 的 MySQL / 内存 fallback 保存、查询、删除、chunk window |
| 搜索统计存储 | `api\internal\store\knowledge_search.go` | 搜索事件、badcase、统计，支持 MySQL / 内存 fallback |
| MySQL 表结构 | `api\internal\store\init.go` | `knowledge_documents`、`knowledge_search_events`、`knowledge_search_badcases`、案例、Runbook 表 |
| 文档处理 | `api\internal\knowledge\parser.go`、`chunker*.go`、`indexer.go` | 文件解析、Markdown/结构化分块、索引、删除、重建、案例/Runbook 同步 |
| 搜索接口 | `api\internal\handler\documents.go` | upload/list/get/delete/reindex/search/stats/badcase/reindex-all |
| Embedding 配置接口 | `api\internal\handler\embedding_settings.go` | embedding/reranker 配置读取、保存、测试、掩码回显 |
| Embedding 配置 | `api\internal\embedding\config.go` | provider: builtin/api/hybrid；api url/key/model/dimensions/batch size；reranker |
| 搜索抽象 | `api\internal\embedding\provider.go` | `EmbeddingProvider`、`SearchProvider`、`Document`、`SearchResult` |
| BM25 | `api\internal\embedding\bm25.go` | 关键词检索引擎 |
| 内存向量 | `api\internal\embedding\vector.go` | `VectorEngine` 使用进程内 `map[string]*vectorDoc` 和余弦相似度 |
| Hybrid / RRF | `api\internal\embedding\search.go` | `HybridSearcher` 支持 builtin/api/hybrid；hybrid 用 RRF 融合 BM25 和向量结果 |
| API Embedding | `api\internal\embedding\api_embed.go` | 外部 embedding API 调用 |
| Reranker | `api\internal\embedding\reranker.go` | 可选重排 |
| 文档管理 UI | `web\src\components\knowledge\DocManager.vue` | 上传、类型筛选、关键词、列表、详情、删除 |
| 语义搜索 UI | `web\src\components\knowledge\SemanticSearch.vue` | 自然语言查询、doc type、topK、结果、chunk context、badcase |
| Embedding 配置 UI | `web\src\components\ai\EmbeddingConfig.vue` | provider、API URL、API Key、模型、维度、保存、测试连接 |
| AI 配置入口 | `web\src\views\AIConfig.vue` | LLM、Embedding、Reranker 三个配置 tab |

## 2. 当前 API 路由证据

`api\main.go` 当前知识库和向量配置相关入口：

| 路由 | 当前作用 | P6 结论 |
| --- | --- | --- |
| `POST /api/v1/knowledge/documents/upload` | 上传并解析文档，分块入库和索引 | `SOURCE_PRESENT` |
| `GET /api/v1/knowledge/documents` | 文档分页、类型、分类、关键词查询 | `SOURCE_PRESENT` |
| `GET /api/v1/knowledge/documents/:id` | 文档详情 | `SOURCE_PRESENT` |
| `DELETE /api/v1/knowledge/documents/:id` | 删除文档及分块 | `SOURCE_PRESENT` |
| `POST /api/v1/knowledge/documents/:id/reindex` | 单文档重建 | `SOURCE_PRESENT`，缺索引任务状态 |
| `POST /api/v1/knowledge/search` | 知识检索 | `SOURCE_PRESENT`，当前 engine 可返回 hybrid/fulltext，但 Qdrant 未实现 |
| `GET /api/v1/knowledge/search/stats` | 搜索质量统计 | `SOURCE_PRESENT` |
| `POST /api/v1/knowledge/search/badcase` | 标记不相关结果 | `SOURCE_PRESENT` |
| `POST /api/v1/knowledge/reindex-all` | 异步全量重建 | `SOURCE_PRESENT_PARTIAL`，当前只启动 goroutine，无任务 id、进度、失败重试 |
| `GET /api/v1/settings/embedding` | 获取 embedding 配置，掩码回显 key | `SOURCE_PRESENT` |
| `PUT /api/v1/settings/embedding` | 保存 embedding 配置并 reload searcher | `SOURCE_PRESENT`，provider 不含 qdrant |
| `POST /api/v1/settings/embedding/test` | 测试外部 embedding API | `SOURCE_PRESENT` |
| `GET /api/v1/settings/reranker` | 获取 reranker 配置 | `SOURCE_PRESENT` |
| `PUT /api/v1/settings/reranker` | 保存 reranker 配置 | `SOURCE_PRESENT` |

## 3. 当前状态流证据

| 状态流 | 当前源码证据 | 风险 |
| --- | --- | --- |
| 文档上传 | `UploadDocument` -> `ParseDocument` -> `IndexDocument` | 上传错误直接返回解析/索引错误文本，后续需统一错误脱敏 |
| 文档分块 | `ChunkDocumentByTypeAndFile`、`indexChunks` | 有 parent/chunk_index，但缺 chunk 版本、hash、权限和证据引用 |
| 单文档重建 | `ReindexDocumentHandler` -> `knowledge.ReindexDocument` | 无任务 id、进度、重试、索引状态 |
| 全量重建 | `ReindexAllHandler` 启动 goroutine -> `knowledge.RebuildAll` | 无任务状态、取消、失败重试、审计详情 |
| 搜索 | `SearchKnowledge` 优先 `embedding.GetSearcher().Search`，失败/空结果 fallback 到 `store.ListDocuments` | fallback 只记录 engine=fulltext/hybrid，未区分 BM25、向量、RRF、Qdrant unavailable |
| BM25 | `NewHybridSearcher` provider=builtin/hybrid 启用 BM25 | 可作为 Qdrant 不可用 fallback |
| 向量 | provider=api/hybrid 启用 `VectorEngine` | 当前向量索引只在进程内存，重启丢失，不是内置向量数据库 |
| RRF | `searchHybrid` + `mergeResults` | RRF 已存在，但未记录 recall_method 到搜索结果和 Evidence Chain |
| Reranker | `h.reranker.Rerank` | 可选，但未把 reranker score、model_ref 写入结果或证据链 |
| 配置 | `EmbeddingConfig.vue` + `/settings/embedding` | UI 只有 builtin/api/hybrid，未提供 Qdrant sidecar、collection、健康、索引状态 |
| 搜索质量 | search event + badcase | 已有质量数据，但未形成索引优化任务和证据链引用 |

## 4. Qdrant 当前阻断

本地代码搜索结果：未发现 Qdrant client、配置结构、collection 管理、upsert、delete、search、rebuild、health、snapshot 或 docker/sidecar 运行定义。

| 项目 | 状态 | 说明 |
| --- | --- | --- |
| Qdrant client | `BLOCKED_QDRANT_NOT_IMPLEMENTED` | 未见 Go client 或 HTTP client 实现 |
| Qdrant 配置 | `BLOCKED_QDRANT_CONFIG_NOT_IMPLEMENTED` | 未见 `knowledge.vector.qdrant_url`、collection、api key、timeout、TLS |
| Collection 管理 | `BLOCKED_QDRANT_COLLECTION_NOT_IMPLEMENTED` | 未见 create/ensure/index/schema/snapshot |
| Vector upsert/delete/search | `BLOCKED_QDRANT_VECTOR_OPS_NOT_IMPLEMENTED` | 当前 `VectorEngine` 是内存 map |
| 索引任务状态 | `BLOCKED_INDEX_STATUS_NOT_IMPLEMENTED` | 未见 indexing_jobs / progress / retry / cancel |
| 搜索结果引用 | `BLOCKED_KNOWLEDGE_REF_CHAIN_NOT_IMPLEMENTED` | 搜索结果未记录 recall_method、chunk id、model ref、Evidence Chain item |
| UI 健康与重建 | `BLOCKED_VECTOR_UI_NOT_IMPLEMENTED` | UI 没有 Qdrant 健康、collection、索引进度、降级状态 |

## 5. 实现前必须保留的当前能力

后续接入 Qdrant 时不能破坏已有能力：

- MySQL/MariaDB 仍是权威基石，`knowledge_documents`、案例、Runbook、badcase、search event 不可迁移成只存在 Qdrant。
- 内存 fallback 只能用于开发/组件不可用降级，不能作为生产权威存储。
- BM25 必须保留为关键词检索和 Qdrant 不可用降级。
- RRF hybrid 必须保留，后续从 BM25 + 内存向量升级为 BM25 + Qdrant 向量。
- Reranker 可选，必须继续从系统 AI 配置读取。
- 文档 chunk window、badcase、search stats、reindex 接口必须保留并增强任务状态。

## 6. 后续验收证据要求

P6 进入实现后必须自己执行并留证：

- 前端 lint/build。
- 后端 go test/build。
- FindX 自有登录。
- 文档上传、列表、详情、删除、单文档重建、全量重建。
- BM25 搜索、Qdrant 搜索、hybrid RRF、reranker、badcase、search stats。
- Qdrant 搜索、Qdrant 未配置、不可达、collection 缺失、upsert 失败、search 超时、权限不足、空结果；这些都是实现后必测项，当前源码阶段不可标 PASS。
- 索引任务进度、失败重试、取消、重新执行、审计。
- Evidence Chain 引用 document/chunk/search/model/index job。
- MCP 浏览器点击知识库、语义搜索、AI 模型配置、索引状态页面。
- 窄屏回归、敏感信息扫描、乱码扫描、用户侧外部品牌扫描。
