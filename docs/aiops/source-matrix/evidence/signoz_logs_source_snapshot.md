# SigNoZ 日志中心源码检查证据

生成时间：2026-05-07 14:09（UTC+8）
状态：源码证据已检查；运行态 DOM 为 `BLOCKED_RUNTIME_UNAVAILABLE`

## 1. 本地源码路径

事实源目录：

- `D:\平台源码\signoz-develop\frontend`

已检查入口：

| 类型 | 路径 | 结论 |
| --- | --- | --- |
| 路由 | `src\AppRoutes\routes.ts` | 包含 Logs、Logs Explorer、Live Logs、Pipelines、Saved Views、Trace Detail、Traces Explorer 路由 |
| 路由常量 | `src\constants\routes.ts` | 定义 `/logs/logs-explorer`、`/logs/logs-explorer/live`、`/logs/pipelines`、`/logs/saved-views`、`/trace/:id`、`/traces-explorer`、`/traces/saved-views` |
| 页面映射 | `src\AppRoutes\pageComponents.ts` | 将 Logs、LiveLogs、PipelinePage、TraceDetail、TracesExplorer、TracesSaveViews 映射为懒加载页面 |
| 日志模块页 | `src\pages\LogsModulePage\LogsModulePage.tsx` | 使用 `RouteTab` 组织 Explorer、Pipelines、Views 三个 tab |
| 日志查询页 | `src\pages\LogsExplorer\index.tsx` | 由 QuickFilters、Toolbar、ExplorerCard、LogExplorerQuerySection、LogsExplorerViews 组成 |
| 实时日志页 | `src\pages\LiveLogs\index.tsx` | 使用 `EventSourceProvider`、`LiveLogsContainer`、`useQueryBuilder` 和 `PANEL_TYPES.LIST` |
| Pipeline 页 | `src\pages\Pipelines\index.tsx` | 使用 latest pipeline 查询、Pipelines / Change History tabs、部署中 3000ms refetch |
| 日志组件 | `src\components\Logs\*`、`src\components\LogDetail\*`、`src\components\QuickFilters\*` | 覆盖列表视图、表格视图、原文视图、详情、字段筛选、添加到查询、复制等动作 |
| 查询构造 | `src\hooks\queryBuilder\*` | 覆盖 autocomplete、key/value suggestions、operator、tag validation、URL share |
| 日志 API | `src\api\logs\*` | 覆盖 `/logs`、`/logs/aggregate`、`/logs/fields`、`/logs/tail` |
| Pipeline API | `src\api\pipeline\*` | 覆盖 `/logs/pipelines/:version`、`/logs/pipelines`、`/logs/pipelines/preview` |
| Saved Views API | `src\api\saveView\*` | 覆盖 `/explorer/views` CRUD |
| Trace API | `src\api\trace\*` | 覆盖 `/traces/:id`、`/getFilteredSpans`、`/getFilteredSpans/aggregates`、tag filter/value |

## 2. 路由证据摘要

`src\AppRoutes\routes.ts` 中确认存在：

- `ROUTES.LOGS`
- `ROUTES.LOGS_EXPLORER`
- `ROUTES.OLD_LOGS_EXPLORER`
- `ROUTES.LIVE_LOGS`
- `ROUTES.LOGS_PIPELINES`
- `ROUTES.LOGS_SAVE_VIEWS`
- `ROUTES.TRACE_DETAIL`
- `ROUTES.TRACES_EXPLORER`
- `ROUTES.TRACES_SAVE_VIEWS`

`src\constants\routes.ts` 中确认存在：

- `/logs`
- `/logs/logs-explorer`
- `/logs/logs-explorer/live`
- `/logs/pipelines`
- `/logs/saved-views`
- `/trace/:id`
- `/traces-explorer`
- `/traces/saved-views`

结论：FindX 不能把日志中心压缩成一个静态日志表，必须保留日志检索、实时日志、Pipeline、Saved Views、Trace 关联和 Trace 详情入口。

## 3. 页面结构证据摘要

`src\pages\LogsModulePage\LogsModulePage.tsx`：

- 使用 `RouteTab` 作为日志模块的二级导航。
- tab 来源为 `logsExplorer`、`logsPipelines`、`logSaveView`。
- 当前路径通过 `useLocation().pathname` 计算 active tab。

`src\pages\LogsExplorer\index.tsx`：

- 左侧为 `QuickFilters`，支持显示/隐藏并写入 `LOCALSTORAGE.SHOW_LOGS_QUICK_FILTERS`。
- 顶部为 `Toolbar`，左侧 `LeftToolbarActions` 控制过滤器、Search / Query Builder 视图和 histogram 显示。
- 顶部右侧 `RightToolbarActions` 触发 `handleRunQuery` 并接收列表/图表 query key loading 状态。
- 查询区为 `ExplorerCard` + `LogExplorerQuerySection`。
- 结果区为 `LogsExplorerViews`，负责频率图、日志列表、loading、empty、error 等状态。
- 当 query builder 中存在多 query 或 groupBy 时，自动切到 Query Builder 视图。

`src\pages\LiveLogs\index.tsx`：

- 使用 `useShareBuilderUrl(liveLogsCompositeQuery)` 保持 URL 可分享。
- 使用 `useQueryBuilder().handleSetConfig(PANEL_TYPES.LIST, DataSource.LOGS)` 固定为日志列表形态。
- 使用 `EventSourceProvider` 包裹 `LiveLogsContainer`。

`src\pages\Pipelines\index.tsx`：

- 使用 `useQuery(['version', 'latest', 'pipeline'])` 获取 latest pipeline。
- 部署状态未结束时每 3000ms refetch。
- tab 包含 `Pipelines` 和 `Change History`。
- loading 状态使用 `Spinner`，错误通过 notification 展示。

## 4. API 证据摘要

| 文件 | 路径/动作 | FindX Adapter 要求 |
| --- | --- | --- |
| `src\api\logs\GetLogs.ts` | `GET /logs` | 支持条件、时间范围、分页、排序、空结果、错误脱敏 |
| `src\api\logs\GetLogsAggregate.ts` | `GET /logs/aggregate` | 支持聚合图、频率图、groupBy、空聚合 |
| `src\api\logs\GetSearchFields.ts` | `GET /logs/fields` | 支持字段筛选、字段建议、权限过滤 |
| `src\api\logs\livetail.ts` | `GET /logs/tail` event stream | 支持实时日志、心跳超时、停止、重试、登录过期 |
| `src\api\pipeline\get.ts` | `GET /logs/pipelines/:version` | 支持 latest、历史版本、部署状态 |
| `src\api\pipeline\post.ts` | `POST /logs/pipelines` | 支持保存 pipeline、校验、审计、错误脱敏 |
| `src\api\pipeline\preview.ts` | `POST /logs/pipelines/preview` | 支持 sample logs 预览、处理前后对比 |
| `src\api\saveView\getAllViews.ts` | `GET /explorer/views?sourcePage=...` | 支持日志和 Trace 视图区分、权限 |
| `src\api\saveView\saveView.ts` | `POST /explorer/views` | 支持保存查询、名称、extraData |
| `src\api\saveView\updateView.ts` | `PUT /explorer/views/:id` | 支持改名、更新查询和 extraData |
| `src\api\saveView\deleteView.ts` | `DELETE /explorer/views/:id` | 支持删除确认、权限和审计 |
| `src\api\trace\getTraceItem.ts` | `GET /traces/:id` | 支持 Trace 详情、spanId、上下游层级 |
| `src\api\trace\getSpans.ts` | `POST /getFilteredSpans/aggregates` | 支持 Trace 聚合、duration、exclude、tags |
| `src\api\trace\getSpansAggregate.ts` | `POST /getFilteredSpans` | 支持 Trace 列表、分页、排序、过滤 |

注意：Trace API 表按上游源码文件名记录；`getSpans.ts` 请求聚合端点，`getSpansAggregate.ts` 请求列表端点。后续 FindX Adapter 实现前必须以实际响应结构和页面调用点复核列表/聚合语义，不能按文件名直觉接线。

FindX Logs Adapter 不得把上游 API 原样暴露为最终业务公共契约；必须加权限、租户、审计、错误脱敏、组件健康和数据源配置一致性。

## 5. 运行态 DOM 状态

当前没有可访问的 SigNoZ 运行态参考页，因此运行态 DOM 证据为：

- `BLOCKED_RUNTIME_UNAVAILABLE`

后续 P4 编码完成前，必须用 MCP 浏览器补齐：

- 日志检索。
- Quick Filters 展开/隐藏。
- Search / Query Builder 切换。
- 字段 autocomplete。
- 频率图显示/隐藏。
- 日志列表、表格、原文、详情抽屉/详情区。
- 添加字段到查询、复制、上下文跳转。
- Live tail 启动、停止、断流重试。
- Pipeline 列表、新增、编辑、预览、保存、部署中轮询、Change History。
- Saved Views 保存、加载、更新、删除、权限不足。
- Trace 关联和 Trace 详情跳转。
- Agent 日志采集状态入口、覆盖率入口和数据到达证据跳转。
- Query Service / ClickHouse / OTel Collector / Logs Adapter 未配置时显示 `BLOCKED` 或脱敏 503。

## 6. FindX 替换点

| 来源概念 | FindX 用户侧命名 |
| --- | --- |
| Logs Explorer | 日志检索 |
| Live Logs | 实时日志 |
| Pipelines | 日志处理管道 |
| Views / Saved Views | 保存视图 |
| Trace Detail | Trace 详情 / 链路详情 |
| ClickHouse 查询 | 日志查询执行计划 / 高级查询 |

外部来源名称只允许保留在源码矩阵、合规登记、归档和开发证据中；用户侧菜单、标题、权限对象、审计对象不得显示外部品牌。

## 7. 阻断结论

- `SOURCE_PRESENT`：本地日志中心前端源码存在，路由、页面、组件、API、状态流可审计。
- `BLOCKED_RUNTIME_UNAVAILABLE`：没有可访问运行态参考页，不能声明 UI 一比一完成。
- `BLOCKED_BY_ADAPTER_CONTRACT`：FindX Logs Adapter、ClickHouse / Query Service / OTel Collector 健康检查、权限审计、错误脱敏契约尚未实现。
- `BLOCKED_BY_BROWSER_REGRESSION`：编码后必须补 MCP 浏览器真实登录、点击和窄屏回归。
