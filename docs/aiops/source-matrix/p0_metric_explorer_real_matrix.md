# FindX P0 指标查询页面同源矩阵

生成时间：2026-05-07 15:05（UTC+8）
状态：P0-N9E-METRIC-EXPLORER-REAL 编码前门禁，不代表 FindX 指标查询页面已经完成

## 1. 结论

指标查询必须按成熟源码同源实现。`添加面板` 是新增 query panel，不是收藏；`内置指标` 是 PromQL 输入辅助，不是常驻大搜索框；`历史记录` 是 localStorage 查询历史 Popover；`Enable autocomplete`、Table/Graph、Time、Unit、CSV 导出、分享、图表设置、Max data points、Min step 都有真实状态流和 API 语义，不能静态展示。

| 项目 | 事实 |
| --- | --- |
| 页面源码 | `D:\项目迁移文件\平台源码\fe-main\src\pages\explorer\Metric.tsx` |
| 面板源码 | `D:\项目迁移文件\平台源码\fe-main\src\pages\explorer\Explorer.tsx` |
| Prometheus 查询源码 | `D:\项目迁移文件\平台源码\fe-main\src\pages\explorer\Prometheus\index.tsx` |
| 图表组件源码 | `D:\项目迁移文件\平台源码\fe-main\src\components\PromGraphCpt` |
| PromQL 输入源码 | `D:\项目迁移文件\平台源码\fe-main\src\components\PromQLInputNG` |
| 内置指标源码 | `D:\项目迁移文件\平台源码\fe-main\src\components\PromQLInput\BuiltinMetrics` |
| 历史记录源码 | `D:\项目迁移文件\平台源码\fe-main\src\pages\explorer\Prometheus\HistoricalRecords\index.tsx` |
| API 源码 | `D:\项目迁移文件\平台源码\fe-main\src\services\metric.ts`、`PromGraphCpt\services.ts`、`metricsBuiltin\services.ts` |
| 参考 DOM | `docs\aiops\source-matrix\evidence\n9e_metric_explorer_snapshot.md` |

## 2. 页面结构

| 区域 | 源码证据 | 结构与动作 | FindX 要求 |
| --- | --- | --- | --- |
| 页面标题 | `Metric.tsx` `PageLayout title={t('title')}` | 标题、图标、文档入口 | FindX 标题为“指标查询”，不出现外部品牌 |
| 面板容器 | `Metric.tsx` `panels` state | 每个 panel 是 `bg-fc-100 fc-border rounded-lg p-4 max-h-[650px]`，内部渲染 `Explorer` | 可换皮肤，但布局和信息密度不弱化 |
| 添加面板 | `setPanels([...panels, { uuid }])` | 新增一个查询 panel；多 panel 时可关闭 | 禁止映射为收藏、保存视图或静态按钮 |
| 关闭面板 | `CloseCircleOutlined` | 多 panel 才显示；关闭当前 panel；若 AI chat 来源为该 panel 则关闭 AI chat | 关闭动作必须真实维护 panel state |
| 数据源选择 | `Explorer.tsx` + `DatasourceSelectV3` | 隐藏 `datasourceCate`，显示 datasource id 选择；按类型过滤；更新默认数据源和 querystring | 数据源切换必须影响查询 API 和历史记录 key |
| 视图保存 | `ViewSelect` | 保存/选择当前查询视图，支持 filters、modalState、oldFilterValues | 后续实现不能丢；未完成必须 BLOCKED |
| 查询主体 | `Prometheus/index.tsx` + `PromGraphCpt` | PromQL 输入、查询按钮、Table/Graph 结果 | 同源实现，不做自研面板 |

## 3. PromQL 输入与辅助

| 控件 | 源码证据 | 行为 | 用户指出的阻断 |
| --- | --- | --- | --- |
| Enable autocomplete | `PromGraphCpt/index.tsx` `completeEnabled` | Checkbox 控制 `PromQLInputNG enableAutocomplete` | 不能静态展示；必须真正影响补全请求 |
| PromQL 编辑器 | `PromQLInputNG/index.tsx` | Monaco 编辑器，API prefix 为 `/api/<path>/proxy/:datasource/api/v1`，支持 onBlur/onEnter 触发 | 输入框尺寸按源码，不做过大自研搜索框 |
| 内置指标 | `BuiltinMetrics/index.tsx`、`Content.tsx`、`MetricsList.tsx` | dropdown，含搜索、类型、collector、默认类型 tag、无限滚动、详情、选择后插入 expression 并更新 unit | 右侧作用必须是输入辅助；不能做常驻大搜索框 |
| 全局指标浏览 | `MetricsExplorer.tsx` | 点击 Global 图标打开 Metrics Explorer Modal，拉取 `label/__name__/values`，搜索后插入指标名 | 必须保留 modal/search/insert 语义 |
| 历史记录 | `HistoricalRecords/index.tsx` | localStorage key 为 `n9e-query-promql-history-<datasource>`，最多 100 条，Popover 搜索，点击回填 | 不是静态文本；要按数据源隔离 |
| AI 生成查询 | `AiButton` | 打开 AI query generator，执行后回填 PromQL | 全局 AI 配置必须走系统配置 / AI 模型配置 |

## 4. Table 语义

事实源：`D:\项目迁移文件\平台源码\fe-main\src\components\PromGraphCpt\Table.tsx`

| 控件/状态 | 行为 |
| --- | --- |
| Time | DatePicker，默认当前时间，可选择 evaluation time，禁用未来时间 |
| Unit | `UnitPicker`，影响 valueFormatter |
| CSV 导出 | 有数据时 `json2csv` + `downloadFile`，文件名含时间 |
| 查询 API | `GET <proxy>/<datasource>/api/v1/query` |
| 变量插值 | `instantInterpolateString` |
| 结果类型 | 支持 matrix/vector/scalar/string/streams |
| LIMIT | 超过 1000 只显示前 1000，并显示 Warning |
| 统计 | `loadTime`、`resultSeries` |
| 错误 | `Error executing query: <msg>`，FindX 需脱敏 |

## 5. Graph 语义

事实源：`D:\项目迁移文件\平台源码\fe-main\src\components\PromGraphCpt\Graph.tsx`

| 控件/状态 | 行为 |
| --- | --- |
| 时间范围 | `TimeRangePicker`，默认 now-1h 到 now |
| Max data points | 影响真实 step 计算，默认 panel width |
| Min step | 输入/blur/step 更新，影响 query_range |
| 图形类型 | Line / StackArea 切换，影响 drawStyle、fillOpacity、stack |
| 分享 | `setTmpChartData` 后打开 `/chart/:ids` |
| 设置 | `LineGraphStandardOptions` Popover/水平配置 |
| 查询 API | `GET <proxy>/<datasource>/api/v1/query_range` |
| 断点补全 | `completeBreakpoints(realStep, values)` |
| 图例/tooltip/unit | 读取 `siteInfo.explorer_timeseries_legend_columns` 与 highLevelConfig |

## 6. API 契约

| API | 来源 | 用途 | FindX 要求 |
| --- | --- | --- | --- |
| `/api/n9e/proxy/:datasource/api/v1/query` | `Table.tsx` + `getPromData` | 即时查询 | 支持 timeout、取消、401/403、503、空结果、错误脱敏 |
| `/api/n9e/proxy/:datasource/api/v1/query_range` | `Graph.tsx` + `getPromData` | 区间查询 | 支持 step/maxDataPoints/minStep 语义 |
| `/api/n9e/proxy/:datasource/api/v1/label/__name__/values` | `MetricsExplorer.tsx` | 全局指标浏览 | 不能静态列表；需数据源隔离 |
| `/api/n9e/builtin-metrics` | `metricsBuiltin/services.ts` | 内置指标列表 | 支持 query/type/collector/page/limit/unit |
| `/api/n9e/builtin-metrics/types` | `metricsBuiltin/services.ts` | 类型筛选 | 用于内置指标 dropdown |
| `/api/n9e/builtin-metrics/collectors` | `metricsBuiltin/services.ts` | collector 筛选 | 选择后写 localStorage 默认 collector |
| `/api/n9e/share-charts` | `PromGraphCpt/services.ts` | 临时分享图表 | FindX 需审计和过期策略 |

## 7. 当前 FindX 差异

| 当前实现 | 问题 | 处理 |
| --- | --- | --- |
| `MonitorDatasourceQueryPanel.vue` | 数据源管理、查询、labels/metrics 混在一个自研面板 | 下线为兼容/诊断工具，不作为指标查询主路径 |
| 查询输入过大/布局拥挤 | 与成熟源码输入区尺寸和右侧辅助不同 | 按 `PromGraphCpt` 布局重做 |
| 添加面板语义错误风险 | 用户已指出不能做收藏 | 固定为新增 panel state |
| Time/Unit 静态化风险 | 用户已指出后续所有控件不能静态 | 必须接入 DatePicker/UnitPicker 和查询状态 |
| 内置指标位置错误风险 | 用户已指出内置指标搜索框不能过大 | 按 dropdown + detail + infinite scroll 实现 |

## 8. 编码准入

| 门禁 | 状态 | 说明 |
| --- | --- | --- |
| 源码路径 | `READY` | Metric、Explorer、Prometheus、PromGraph、PromQLInput、BuiltinMetrics、HistoricalRecords 已读取 |
| DOM 证据 | `PARTIAL` | 参考页已有快照；还需 MCP 追加 autocomplete、历史记录、内置指标、Table/Graph、添加面板交互 |
| API_CONTRACT_CHANGE | `EXPECTED` | Query gateway、builtin metrics、share chart、history/view 保存需契约 |
| DATA_CHANGE | `POSSIBLE` | 保存视图、分享图表、内置指标、审计可能需表 |
| 安全 | `REQUIRED` | Prometheus/VictoriaMetrics 错误脱敏；不泄露 query 下游 URL、token、header |
| 验证 | `REQUIRED` | WSL build；MCP 登录点击数据源切换、查询、Table/Graph、Time、Unit、CSV、添加/关闭面板、内置指标、历史记录、非法 PromQL、空结果 |

未补齐以上门禁时，不允许开始 P0 指标查询代码迁移，也不允许把当前自研查询面板判为 PASS。
