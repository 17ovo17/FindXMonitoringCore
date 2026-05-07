# FindX P3 日志中心同源矩阵

生成时间：2026-05-07 14:09（UTC+8）
状态：`P3-SIGNOZ-LOGS-REAL` 编码前门禁，不代表 FindX 已完成实现

## 1. 结论

日志中心不能只做后端日志查询接口，也不能把成熟日志平台的 Logs Explorer 弱化成一个静态表格。FindX 必须按成熟源码迁移或同源实现日志检索、字段筛选、上下文、聚合、Live tail、Pipeline、Saved Views、Trace 关联和 Trace 详情状态流；用户侧统一显示“日志中心”，外部来源名称只保留在内部源码矩阵和合规登记。

源码证据：

- [SigNoZ 日志中心源码检查证据](evidence/signoz_logs_source_snapshot.md)

当前阻断：

- 日志中心前端源码存在。
- 未配置可访问日志中心运行态参考页，运行态 DOM 证据为 `BLOCKED_RUNTIME_UNAVAILABLE`。
- FindX Logs Adapter、日志存储、查询服务、OTel Collector、权限审计和错误脱敏契约未完成。

## 2. FindX 路由规划

| FindX 页面 | 成熟路由/源码 | 页面结构要求 |
| --- | --- | --- |
| 日志中心 | `src\pages\LogsModulePage\LogsModulePage.tsx` | 二级 tab：日志检索、日志处理管道、保存视图 |
| 日志检索 | `/logs/logs-explorer`、`pages\LogsExplorer` | 左侧 Quick Filters、顶部 Toolbar、Search / Query Builder、查询区、频率图、结果区 |
| 实时日志 | `/logs/logs-explorer/live`、`pages\LiveLogs` | EventSource、列表面板、过滤输入、停止/重试、断流错误态 |
| 日志处理管道 | `/logs/pipelines`、`pages\Pipelines`、`container\PipelinePage` | Pipeline 列表、搜索、新增、编辑、拖拽、processor、preview、保存、部署历史 |
| 保存视图 | `/logs/saved-views`、`pages\SaveView`、`api\saveView` | 保存、加载、改名、更新、删除、权限、跳转回日志检索 |
| Trace 关联 | `/trace/:id`、`/traces-explorer`、`api\trace` | trace_id/span_id 关联、Trace 详情、Span 列表、聚合、过滤 |
| 字段管理 | `/logs-explorer/index-fields`、`pages\LogsSettings` | 字段选择、索引字段、搜索字段、错误态 |

FindX 可以提供 `/logs`、`/logs/explorer`、`/logs/live`、`/logs/pipelines`、`/logs/saved-views`、`/logs/trace/:id` 等 alias，但 alias 必须落回同源状态流，不得另造静态页面。

## 3. 页面结构门禁

| 页面区域 | 成熟源码 | FindX 实现要求 |
| --- | --- | --- |
| 日志模块二级导航 | `LogsModulePage.tsx`、`RouteTab` | 保留 Explorer / Pipelines / Views 三段式结构，文案替换为 FindX 用户侧命名 |
| 左侧筛选 | `QuickFilters\QuickFilters.tsx`、`FilterRenderers\Checkbox` | 支持字段组、checkbox、count、include/exclude、展开/收起、loading、empty |
| 顶部工具栏 | `Toolbar`、`LeftToolbarActions`、`RightToolbarActions` | 支持运行查询、视图切换、频率图显示、过滤区显示、loading 禁用 |
| 查询区 | `ExplorerCard`、`LogExplorerQuerySection` | 支持保存视图、query builder、search 模式、URL share、查询校验 |
| 结果区 | `LogsExplorerViews`、`components\Logs\*` | 支持频率图、列表视图、表格视图、原文视图、空态、错误态、加载态 |
| 日志详情 | `components\LogDetail\*` | 支持字段展开、复制、添加到查询、上下文、Trace 链接 |
| 实时日志 | `LiveLogsContainer`、`LiveLogsList`、`LiveLogsListChart` | 支持 event stream、断流、停止、重试、过滤 |
| Pipeline | `container\PipelinePage\*` | 支持 pipeline/processor 新增编辑、preview、保存配置、部署状态、Change History |
| 保存视图 | `pages\SaveView`、`api\saveView` | 支持角色权限、搜索、编辑、删除、跳回对应 explorer |

## 4. 状态流门禁

| 状态流 | 成熟源码 | FindX 实现要求 |
| --- | --- | --- |
| 查询构造 | `hooks\queryBuilder\useQueryBuilder.ts`、`useQueryBuilderOperations.ts` | 保留 DataSource.LOGS、builder/search 切换、formula、groupBy、URL 参数 |
| autocomplete | `useAutoComplete.ts`、`useFetchKeysAndValues.ts`、`useOptions.ts` | 字段 key/value 建议必须远程查询，不能静态写死 |
| 快速筛选显示 | `LOCALSTORAGE.SHOW_LOGS_QUICK_FILTERS` | 保留显示/隐藏偏好，不因刷新丢失 |
| Search / Query Builder | `LogsExplorer\index.tsx` | 多 query 或 groupBy 时自动切 Query Builder |
| 频率图 | `showFrequencyChart` + `LogsExplorerViews` | 支持显示/隐藏、加载、空数据、查询取消 |
| 实时日志 | `EventSourceProvider`、`LiveTail` | 支持心跳超时、停止、重连、登录过期 |
| Pipeline 轮询 | `Pipelines\index.tsx` | latest deployment 未完成时 3000ms refetch，完成后停止 |
| 保存视图 | `api\saveView` + `ExplorerCard` | 保存、更新、删除、加载必须带权限和错误态 |
| Trace 关联 | `api\trace`、`TraceDetail` | 支持 Trace 详情、spanId、上下游层级、聚合和过滤 |

## 5. API / Adapter 契约

FindX 需要新增 Logs Adapter，但不把成熟日志平台原 API 原样暴露成最终业务前端公共契约。

| 能力 | 成熟 API 语义 | FindX Adapter 要求 |
| --- | --- | --- |
| 日志查询 | `GET /logs` | 时间、条件、limit、offset、排序、空结果、取消请求、错误脱敏 |
| 聚合 | `GET /logs/aggregate` | 频率图、groupBy、统计函数、空聚合、超时 |
| 字段 | `GET /logs/fields` | 字段列表、字段类型、权限过滤、索引字段状态 |
| 实时日志 | `/logs/tail` event stream | 脱敏上下文、心跳超时、停止/重试、401 跳 FindX 登录 |
| Pipeline 获取 | `GET /logs/pipelines/:version` | latest、历史版本、部署状态、版本差异 |
| Pipeline 保存 | `POST /logs/pipelines` | 校验、审计、部署状态、失败回滚 |
| Pipeline 预览 | `POST /logs/pipelines/preview` | sample logs、处理前后对比、错误定位 |
| 保存视图 | `/explorer/views` CRUD | sourcePage、名称、query、extraData、权限、审计 |
| Trace 详情 | `GET /traces/:id` | trace/span 详情、上下游层级、错误脱敏 |
| Trace 过滤 | `POST /getFilteredSpans`、`/getFilteredSpans/aggregates` | tags、duration、exclude、groupBy、分页、排序；实现前必须按上游实际响应结构复核列表/聚合语义，避免按文件名直觉接反 |
| 租户与空间 | FindX 权限/租户上下文 | 业务组、团队、项目空间、数据源权限、审计主体必须进入 Adapter 查询上下文 |

错误模型：

- 日志查询服务未配置：503 + `BLOCKED_LOGS_QUERY_NOT_CONFIGURED`
- ClickHouse 未配置：503 + `BLOCKED_LOGS_STORAGE_NOT_CONFIGURED`
- OTel Collector 未配置：503 + `BLOCKED_OTEL_COLLECTOR_NOT_CONFIGURED`
- Pipeline 服务不可用：503 + `LOGS_PIPELINE_UNAVAILABLE`
- Agent 日志采集链路未形成证据：503 + `BLOCKED_AGENT_LOG_INDEX_UNAVAILABLE`
- 查询超时：503 + `LOGS_QUERY_TIMEOUT`
- 非法查询：400 + sanitized validation message
- 权限不足：403
- 登录过期：跳转 FindX 自有登录页
- 空结果：正常空态，不伪造数据

## 6. 存储与兼容性

日志中心、链路监控、指标查询不能强行共用一个简化查询模型。

| 能力 | 推荐路径 | 说明 |
| --- | --- | --- |
| 日志中心 | Query Service + ClickHouse 优先 | 保留日志检索、聚合、Pipeline、Saved Views 的查询模型 |
| Trace 关联 | OTel trace/log 字段、trace_id、span_id、service、resource labels | 通过 FindX Logs Adapter 和 APM Adapter 关联 |
| 实时日志 | OTel Collector / ingestion pipeline + event stream | 支持 tail、断流、重试、权限 |
| Pipeline | 日志处理管道服务 + 版本化配置 | 保存、preview、部署状态、历史版本 |
| 搜索兼容 | ES/OpenSearch 仅作为可选兼容 | 不作为默认必选；不能替代成熟日志中心结构 |

FindX 统一数据源中心必须管理 Query Service、ClickHouse、OTel Collector、可选 ES/OpenSearch、保留策略、索引字段和健康状态。页面内不得散落独立密钥或直连配置。

## 7. 与链路监控和 AI SRE 的关系

| 来源 | 联动要求 |
| --- | --- |
| Trace 详情 | 从 trace_id/span_id 跳到日志上下文，保留时间窗口和 service/resource 过滤 |
| 日志详情 | 能跳回链路详情，展示关联 span、service、instance、错误堆栈 |
| 告警事件 | 日志片段、聚合结果、Pipeline 变更进入告警复盘证据 |
| Agent 管理中心 | 日志采集状态、日志关联到达、OTel Collector 连通性进入 Agent 覆盖率 |
| AI SRE | 只能引用真实日志查询、Trace 关联、Pipeline 变更、Agent 状态；缺数据时返回数据缺失 |

日志中心与 FindX Agent / SkyWalking Agent 控制面的索引关系必须回链到 [P0 SkyWalking Agent 到 FindX Agent 控制面矩阵](p0_skywalking_agent_real_matrix.md)。日志采集状态、OTel Collector 连通性、trace/log 到达、Agent 覆盖率、安装任务、心跳、数据到达验证和 Evidence Chain 不能只停留在日志页面内；任一链路未补齐时，相关页面和 API 只能显示 `BLOCKED_AGENT_LOG_INDEX_UNAVAILABLE` 或对应脱敏阻断状态，不能标完成。

## 8. 前端迁移规则

当前 FindX 长期方向为 React-first 壳层。日志中心事实源本身是 React，因此应优先同源迁移组件、hook、query builder、API 状态流和测试结构。

迁移规则：

- 不 iframe。
- 不嵌入参考站 SSO。
- 不省组件。
- 不把 Logs Explorer 改成静态表格。
- 不把 Pipeline、Saved Views、Live tail、Trace 关联做成侧边栏孤岛。
- 用户侧样式使用 FindX 主题，结构和功能点按成熟源码。
- 用户侧菜单、页面标题、权限对象、审计对象统一使用 FindX 命名。

## 9. MCP、lint、构建验收

编码后必须覆盖：

- FindX 自有登录。
- 日志中心二级导航。
- 日志检索正常查询、非法查询、空结果、超时、权限不足。
- Quick Filters 展开/隐藏、字段选择、include/exclude。
- Search / Query Builder 切换、多 query 和 groupBy 自动切换。
- 字段 autocomplete 和 value suggestions。
- 频率图显示/隐藏、加载、空态。
- 日志列表、表格视图、原文视图、日志详情。
- 添加字段到查询、复制、上下文跳转。
- Live tail 启动、停止、断流、重试。
- Pipeline 新增、编辑、preview、保存、部署中轮询、Change History。
- Saved Views 保存、加载、更新、删除、权限不足。
- Trace 关联和 Trace 详情跳转。
- Query Service / ClickHouse / OTel Collector / Logs Adapter 未配置显示 `BLOCKED`，不得伪装成功。
- 窄屏回归。
- 前端 lint/build、后端测试/build、敏感信息扫描。

## 10. 阻断项

- 只实现日志 API，前端仍是空壳：FAIL。
- 只做静态日志表，不做 Quick Filters、Query Builder、频率图、详情、Pipeline、Saved Views：FAIL。
- 字段联想、历史视图、Pipeline preview、Live tail、Trace 关联变成静态按钮：FAIL。
- 用户侧出现外部品牌标题、菜单、权限对象、审计对象：FAIL。
- 未配置日志组件时页面显示假数据：FAIL。
- 未用 MCP 浏览器做真实登录点击回归却标 PASS：FAIL。
