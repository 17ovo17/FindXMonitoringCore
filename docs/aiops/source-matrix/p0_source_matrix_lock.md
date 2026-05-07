# FindX P0 源码矩阵锁定表

生成时间：2026-05-07 13:08（UTC+8）
状态：P0 编码前门禁证据，不代表业务代码已经实现
适用范围：基础监控、链路监控、日志中心、CMDB、Agent 管理、采集插件、巡检诊断、AI SRE、Evidence Chain

## 1. 门禁结论

本轮已完成成熟源码入口级只读审计，并补充关键参考站运行态 DOM 证据。结论如下：

| 域 | 源码状态 | DOM 状态 | 编码准入结论 |
| --- | --- | --- | --- |
| 基础监控页面 | `SOURCE_PRESENT`，`D:\平台源码\fe-main` 已确认 | `DOM_PARTIAL`，已抓数据源、指标查询、仪表盘、模板中心、工作流、通知媒介 | 允许进入逐页面细化矩阵；未细化页面仍不得编码 |
| 链路监控 UI/OAP | `SOURCE_PRESENT`，`skywalking-booster-ui-main` 与 `skywalking-master` 已确认 | `DOM_PENDING`，当前未接入可运行 SkyWalking UI 参考页 | 只能做源码矩阵和 Adapter 设计，不能声明 UI 完成 |
| SkyWalking Agent 生态 | `SOURCE_BLOCKED_PARTIAL`，本地只有 UI/OAP，独立 Agent 仓库未落地 | `DOM_NOT_APPLICABLE` | P0 必须补齐独立仓库源码或版本登记后再做 Agent 包实现 |
| 日志中心 | `SOURCE_PRESENT`，`D:\平台源码\signoz-develop\frontend` 已确认 | `DOM_PENDING`，当前未接入可运行日志中心参考页 | 只能做源码矩阵和 Logs Adapter 设计，不能声明 UI 完成 |
| CMDB / Agent 在线 | `SOURCE_PRESENT`，`D:\平台源码\AutoOps-main\AutoOps-main` 已确认 | `DOM_PENDING` | 允许进入 CMDB/Agent 页面级矩阵；实现前必须补 MCP DOM |
| 采集插件 | `SOURCE_PRESENT`，`D:\平台源码\categraf-main (1)\categraf-main` 已确认 | `DOM_NOT_APPLICABLE` | 允许进入 FindX Agent 插件目录和配置模板矩阵 |
| 巡检诊断 | `SOURCE_PRESENT`，`D:\平台源码\catpaw-master\catpaw-master` 已确认 | `DOM_NOT_APPLICABLE` | 允许进入诊断会话、结构化执行和 Evidence Chain 矩阵 |

阻断项：

- SkyWalking Java/Python/Node.js/PHP/Go/Rust/Ruby/Nginx Lua/Kong/Browser Client JS 独立仓库当前未在 `D:\平台源码` 中发现，后续不能把 Agent 包目录当作已实现。
- SkyWalking UI、SigNoZ UI、AutoOps CMDB 的运行态 DOM 仍需按实际可访问环境补齐。
- 以下外部品牌只允许出现在本证据文档、合规登记和开发矩阵中，用户侧必须替换为 FindX / FindX Agent / 链路监控 / 日志中心。

## 2. 运行态 DOM 证据

本轮通过 MCP 浏览器访问参考站 `http://198.18.20.146:17000`，生成以下快照：

| 页面 | URL | 证据文件 | 已观察结构 |
| --- | --- | --- | --- |
| 数据源 | `/datasources` | [n9e_datasources_snapshot.md](evidence/n9e_datasources_snapshot.md) | 页面标题、顶部工具区、列表容器 |
| 指标查询 | `/metric/explorer` | [n9e_metric_explorer_snapshot.md](evidence/n9e_metric_explorer_snapshot.md) | 查询卡片、`添加面板` 按钮、结果区域 |
| 仪表盘列表 | `/dashboards` | [n9e_dashboards_snapshot.md](evidence/n9e_dashboards_snapshot.md) | 仪表盘列表入口和工具区 |
| 模板中心 | `/components` | [n9e_components_snapshot.md](evidence/n9e_components_snapshot.md) | 模板中心标题、创建按钮、组件列表/抽屉结构 |
| 工作流 | `/event-pipelines` | [n9e_event_pipelines_snapshot.md](evidence/n9e_event_pipelines_snapshot.md) | 工作流标题、新增按钮、表格、分页 |
| 通知媒介 | `/notification-channels` | [n9e_notification_channels_snapshot.md](evidence/n9e_notification_channels_snapshot.md) | 通知媒介列表、操作入口 |

注意：这些 DOM 证据只证明参考结构存在，不等于 FindX 已实现。后续每个页面切片必须补充更深层交互：新增、编辑、删除、导入、抽屉、弹窗、权限、空态、错误态。

## 3. 基础监控源码矩阵

事实源：`D:\平台源码\fe-main`

| 能力 | 路由/菜单事实源 | 核心组件事实源 | API 事实源 | 状态流和按钮动作 | FindX 替换要求 |
| --- | --- | --- | --- | --- | --- |
| 主导航 | `src\components\SideMenu\menu.tsx` | `getMenuList` 定义基础设施、集成中心、数据查询、告警、通知、人员组织、系统配置 | `src\routers\index.tsx` 权限和路由守卫 | 菜单分组、tabs、deprecated 菜单隐藏、动态 embedded product 菜单 | FindX 自有导航承载；结构同源，用户侧不出现外部品牌 |
| 路由总线 | `src\routers\index.tsx` | `Content`、`RouteWithSubRoutes`、`dynamicPackages`、`dynamicPages` | `getMenuPerm` | 403、404、out-of-service、登录回调、动态插件路由 | React-first 壳层可替换，路由语义不可弱化 |
| 数据源 | `/datasources`、`/datasources/:action/:type/:id` | `src\pages\datasource\index.tsx`、`Form.tsx`、`Detail.tsx`、`components\items*`、`Datasources\*\Form.tsx` | `src\pages\datasource\services.ts`：`/api/n9e/datasource/plugin/list`、`/list`、`/desc`、`/status/update`、Delete、`/api/n9e/server-clusters` | 类型选择、描述、URL/Header/TLS/认证、状态启停、删除、集群、测试连接路径需逐页补齐 | 页面结构按源码；凭据只引用不回显；错误脱敏 |
| 指标查询 | `/metric/explorer` | `src\pages\explorer\Metric.tsx`、`Explorer.tsx`、`Prometheus\index.tsx`、`Prometheus\HistoricalRecords\index.tsx`、`PromGraphCpt` | `src\services\metric.ts`：`/api/n9e/query`、`query-bench`、`prometheus/api/v1/*`、`tag-pairs`、`tag-metrics` | `添加面板` 是 `setPanels([...panels, { uuid }])` 新增 query panel；历史记录走 localStorage；Time/Unit/Table/Graph/Autocomplete 走 PromGraph 语义 | 禁止把添加面板做成收藏；内置指标不是常驻大搜索框 |
| 日志查询 | `/log/explorer`、`/log/index-patterns` | `src\pages\explorer\Log.tsx`、`LogsViewer`、`Elasticsearch`、`Loki`、`IndexPatterns` | `Elasticsearch\services.ts`、`Loki\services.ts` | 字段列表、上下文、表格/原文、索引模式、日志行操作 | 与 SigNoZ 日志中心统一入口时保留成熟交互 |
| 仪表盘 | `/dashboards`、`/dashboards/:id`、`/dashboards/share/:id` | `src\pages\dashboard\List`、`Detail\Detail.tsx`、`Detail\Title.tsx`、`Variables`、`Panels`、`Editor` | `src\services\dashboardV2.ts`：boards、clone、export、configs、query-range-batch、annotations、proxy labels/series/query | 变量初始化、变量搜索/编辑、时间范围、刷新、添加图表、导入 panel、分享、全屏、annotations、重复 panel | 变量必须可选/可搜/可编辑；添加图表、三点菜单、分享、放大不能静态化 |
| 模板中心 | `/components`、`/components/alert/detail`、`/components/dashboard/detail` | `src\pages\builtInComponents\List.tsx`、`Dashboards`、`AlertRules`、`CollectTpls`、`Metrics`、`Instructions`、`LogoPicker`、`PayloadFormModal` | `src\pages\builtInComponents\services.ts`：`/api/n9e/builtin-components`、`/builtin-payloads`、`/builtin-payloads/cates` | 左侧组件列表、readme 抽屉、tabs、图标、分类、说明、导入 dashboard/alert/collect/metrics | 用户侧统一 FindX / FindX Agent；`categraf-agent` 文案改 `findx-agent`，不删除模板含义 |
| 告警规则 | `/alert-rules`、add/edit、mutes、subscribes、events | `src\pages\alertRules`、`warning\shield`、`warning\subscribe`、`event`、`historyEvents` | `alertRules\services.ts`、`shield.ts`、`subscribe.ts`、`warning.ts` | 规则表、克隆/编辑/删除、触发器、通知、事件设置、屏蔽、订阅、事件详情 | 订阅、屏蔽、模板、事件流水线不得混成一个自研页 |
| 工作流 / 事件流水线 | `/event-pipelines`、`/event-pipelines-executions` | `src\pages\eventPipeline\pages\List`、`Form`、`Processor`、`TestModal`、`Executions` | `src\pages\eventPipeline\services.ts`：event-pipeline CRUD、tryrun、tagkeys/tagvalues、executions | 列表、新增、编辑、删除、processor tryrun、pipeline tryrun、执行详情抽屉 | 入口并入 AI SRE 工作流，内部结构按源码保留 |
| 通知 | `/notification-rules`、`/notification-channels`、`/notification-templates` | `notificationChannels`、`notificationRules`、`notificationTemplates` | 对应 `services.ts` | 通知媒介列表、表单类型、测试、模板、规则 | 页面按源码结构；禁止复杂自研抽屉替代成熟列表/表单 |
| 组织与系统配置 | `/users`、`/user-groups`、`/roles`、`/system/*`、`/ai-config/*` | `user`、`permissions`、`siteSettings`、`variableConfigs`、`help\SSOConfigs`、`aiConfig` | `manage.ts`、`help.ts`、`common.ts` | 用户、团队、角色、变量、站点、SSO、告警引擎、AI 配置 | AI 模型配置归系统配置统一入口；审计日志为 FindX 增强项 |

## 4. 链路监控源码矩阵

事实源：

- UI：`D:\平台源码\skywalking-booster-ui-main`
- OAP/API：`D:\平台源码\skywalking-master\oap-server\server-query-plugin`

| 能力 | 路由/组件事实源 | 状态流事实源 | API/Query 事实源 | FindX 实现要求 | 当前状态 |
| --- | --- | --- | --- | --- | --- |
| 路由总线 | `src\router\index.ts` 合并 marketplace、layers、alarm、dashboard、settings、trace | Vue Router guard | `src\router\constants.ts` 定义 `/dashboard/list`、`/dashboard/new`、`/dashboard/:layerId/:entity/:name`、`/traces/:traceId`、`/alerting` | FindX `/tracing/*` alias 必须落回同源状态流 | `SOURCE_PRESENT` |
| 服务/层级/总览 | `src\views\Layer.vue`、`src\views\dashboard\List.vue`、`New.vue`、`Edit.vue`、`Widget.vue` | `src\store\modules\dashboard.ts`、`selectors.ts` | `src\graphql\query\dashboard.ts`、`selector.ts` | React 迁移时必须等价保留 Pinia selector/loading/duration/session 语义 | `SOURCE_PRESENT` |
| 拓扑 | dashboard widget + topology view | `src\store\modules\topology.ts` | `src\graphql\query\topology.ts`：instance、endpoint、services、process、hierarchy topology | 拓扑不能做静态图，必须保留节点/边、层级、时间范围、loading、错误态 | `SOURCE_PRESENT` |
| Trace | `src\views\dashboard\Trace.vue`、`src\router\trace.ts` | `src\store\modules\trace.ts` | `src\graphql\query\trace.ts`：traces、spans、tag keys/values、cold stage、v2 support | Trace 列表、详情、span、标签筛选、冷数据和取消请求必须保留 | `SOURCE_PRESENT` |
| Profiling/eBPF | dashboard/profile views | `profile.ts`、`async-profiling.ts`、`continous-profiling.ts`、`network-profiling.ts`、`ebpf.ts` | `profile.ts`、`async-profile.ts`、`ebpf.ts` | Profiling 任务、进度、火焰图、网络 profiling 不能弱化成说明页 | `SOURCE_PRESENT` |
| 告警 | `src\views\Alarm.vue`、`src\views\alarm\*` | `src\store\modules\alarm.ts` | `src\graphql\query\alarm.ts`：alarms、tag keys/values | 告警查询、过滤、详情和错误态走 APM Adapter | `SOURCE_PRESENT` |
| GraphQL 基础 | `src\graphql\base.ts` | AbortController、timeout、全局取消 | BasePath 为 `/graphql` | FindX APM Adapter 代理字段和错误，不直接暴露外部 GraphQL 为公共契约 | `SOURCE_PRESENT` |
| OAP Query | `server-query-plugin\query-graphql-plugin`、`promql-plugin`、`logql-plugin` | Java query handler/service | GraphQL schema/resolver、PromQL/LogQL plugin | 需要 API_CONTRACT_CHANGE 时先写 Adapter 契约 | `SOURCE_PRESENT` |

运行态阻断：当前未配置可访问 SkyWalking UI/OAP 运行态，因此链路监控 UI 完成态必须保持 `BLOCKED`，直到 MCP 浏览器覆盖服务目录、拓扑、Trace、Profiling、告警和配置缺失错误态。

## 5. SkyWalking Agent 仓库矩阵

当前本地 `D:\平台源码` 未发现以下独立仓库目录。P0 后续必须补齐源码目录、版本/commit、许可证、NOTICE、制品形态和配置模板后，才能进入 FindX Agent 包实现。

| 能力包 | 上游仓库 | 本地源码状态 | FindX 包名 | 必须验证 |
| --- | --- | --- | --- | --- |
| Java | `https://github.com/apache/skywalking-java` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Java 探针 | JVM 参数、插件清单、服务名、OAP 连通、Trace/Metric/Log 到达 |
| Python | `https://github.com/apache/skywalking-python` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Python 探针 | pip/离线包、环境变量、进程重启、数据到达 |
| Node.js | `https://github.com/apache/skywalking-nodejs` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Node.js 探针 | npm 包、启动参数、框架插件、错误上报 |
| PHP | `https://github.com/apache/skywalking-php` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent PHP 探针 | 扩展安装、php.ini、FPM/Apache、健康检查 |
| Go | `https://github.com/apache/skywalking-go` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Go 探针 | 编译/注入方式、模块依赖、回滚 |
| Rust | `https://github.com/apache/skywalking-rust` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Rust 探针 | crate、运行时、Trace 传播、数据到达 |
| Ruby | `https://github.com/apache/skywalking-ruby` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Ruby 探针 | gem、框架兼容、启动注入 |
| Nginx Lua | `https://github.com/apache/skywalking-nginx-lua` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent 网关探针 | OpenResty/Nginx 兼容、Lua 模块、重载/回滚 |
| Kong | `https://github.com/apache/skywalking-kong` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Kong 探针 | Kong 插件安装、路由/服务绑定、接入验证 |
| Browser Client JS | `https://github.com/apache/skywalking-client-js` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Browser Agent | SDK 引入、隐私脱敏、RUM 事件、SourceMap |

## 6. 日志中心源码矩阵

事实源：`D:\平台源码\signoz-develop\frontend`

| 能力 | 路由事实源 | 页面/组件事实源 | API 事实源 | FindX 实现要求 | 当前状态 |
| --- | --- | --- | --- | --- | --- |
| 路由 | `src\AppRoutes\routes.ts`、`src\constants\routes.ts` | `LogsModulePage`、`LogsExplorer`、`LiveLogs`、`Pipelines`、`TracesModulePage`、`TraceDetail` | routes 中 `LOGS_EXPLORER`、`LIVE_LOGS`、`LOGS_PIPELINES`、`LOGS_SAVE_VIEWS`、`TRACE_DETAIL` | FindX `/logs/*` alias 必须落回同源状态流 | `SOURCE_PRESENT` |
| 日志检索 | `/logs/logs-explorer` | `components\Logs`、`LogDetail`、`QuickFilters`、`TabsAndFilters`、`container\LogsExplorer*` | `src\api\logs\GetLogs.ts` `/logs`、`GetLogsAggregate.ts` `/logs/aggregate`、`GetSearchFields.ts` `/logs/fields` | 字段筛选、上下文、聚合、视图切换、loading/error/empty 必须保留 | `SOURCE_PRESENT` |
| Live tail | `/logs/logs-explorer/live` | `container\LiveLogs`、`LogLiveTail`、`LiveLogsTopNav` | `src\api\logs\livetail.ts` `/logs/tail` | WebSocket/stream、停止/重试、权限和错误态必须保留 | `SOURCE_PRESENT` |
| Pipeline | `/logs/pipelines` | `pages\Pipelines`、`container\PipelinePage` | `src\api\pipeline\get.ts`、`post.ts`、`preview.ts` | Pipeline 列表、编辑、preview、失败提示不能静态化 | `SOURCE_PRESENT` |
| Saved Views | `/logs/saved-views` | `container\LogsExplorerViews` | `src\api\saveView\getAllViews.ts`、`saveView.ts`、`updateView.ts`、`deleteView.ts` | 保存、更新、删除、加载视图必须接入真实 API | `SOURCE_PRESENT` |
| Trace 关联 | `/trace/:id`、`/traces-explorer` | `container\TraceDetail`、`TracesExplorer`、`TraceFlameGraph` | `src\api\trace\getTraceItem.ts` `/traces/:id`、`getSpans.ts` `/getFilteredSpans/aggregates` | 日志到 Trace、Trace 到日志的跳转和字段上下文必须保留 | `SOURCE_PRESENT` |

运行态阻断：当前没有可访问 SigNoZ 运行态参考页，P4 完成前必须补 MCP DOM：日志查询、字段筛选、上下文、聚合、live tail、Pipeline preview、Saved Views、Trace 关联、组件缺失 `BLOCKED`。

## 7. CMDB 与 Agent 在线源码矩阵

事实源：`D:\平台源码\AutoOps-main\AutoOps-main`

| 能力 | 路由事实源 | 页面事实源 | API 事实源 | 状态流和动作 | FindX 实现要求 |
| --- | --- | --- | --- | --- | --- |
| CMDB 主机 | `web\src\router\cmdb.js`：`/cmdb/ecs` | `web\src\views\cmdb\cmdbHost.vue`、`Host\CmdbHostTable.vue`、`CreateHost.vue`、`EditHost.vue` | `web\src\api\cmdb.js`：host list/create/update/delete/info/search | 搜索、分页、分组、创建、编辑、删除、云主机导入 | FindX CMDB 主机表按源码结构，权限审计补齐 |
| CMDB 分组 | `/cmdb/group` | `cmdbGroup.vue`、`Host\CmdbGroup.vue` | `grouplist`、`groupadd`、`groupupdate`、`groupdelete`、`grouplistwithhosts` | 分组树/列表、增删改、主机关联 | 不能弱化成静态树 |
| 终端/监控弹窗 | `/cmdb/ssh` | `Host\SSH.vue`、`Terminal.vue`、`MonitorDialog.vue`、`ProcessMonitorDialog.vue`、`TcpPortMonitorDialog.vue` | `hostssh/connect`、upload、command、monitor hosts metrics、ports、top-processes | WebSocket、命令执行、文件上传、监控详情 | 远程命令必须权限、审计、脱敏 |
| Agent 列表 | `web\src\router\Tools.js`：`/ops/agent` | `web\src\views\Tools\Agent.vue` | `cmdbAPI.getAgentList` 等 | 状态筛选、版本搜索、部署、批量卸载、重启、详情、健康状态、最后心跳 | 合并进 FindX Agent 管理中心，用户侧命名 FindX Agent |
| Agent 部署 | Tools routes + deploy dialogs | `DeployDialog.vue`、`SelectDeployHost.vue`、`DeployManage.vue`、`DeployProgress.vue` | `web\src\api\tool.js`、部署状态轮询 | 选择主机、部署计划、进度、日志、卸载、轮询 | 与 SkyWalking Agent 包仓库、Categraf/Catpaw 能力包统一 |

运行态阻断：AutoOps/AIOps 运行态 DOM 尚未抓取，P2 编码前必须补 CMDB 树、主机表、Agent 列表、部署流程、心跳异常和权限态。

## 8. FindX Agent 插件与巡检矩阵

| 来源 | 本地源码 | 关键事实 | FindX 映射 | 编码前还缺 |
| --- | --- | --- | --- | --- |
| Categraf 插件生态 | `D:\平台源码\categraf-main (1)\categraf-main` | `inputs\*` 插件目录；`conf\input.*\*.toml` 配置模板；`agent\install\service_linux.go`、`service_windows.go`；`heartbeat`、`writer`、`api` | FindX Agent 插件目录、配置模板、下发、回滚、心跳、指标到达 | 插件分类、模板元数据、默认启停、配置字段 schema |
| Catpaw 巡检诊断 | `D:\平台源码\catpaw-master\catpaw-master` | `agent`、`chat`、`conf.d\p.*`、`digcore\diagnose`、`digcore\config`、`design.d` | FindX Agent 巡检诊断、结构化执行、诊断会话、Evidence Chain | 诊断工具 schema、执行权限、结果模型、AI 模型配置入口 |
| SkyWalking language agents | 上游 URL 已登记，独立本地源码未落地 | 多语言应用探针、网关探针、Browser Agent | FindX Agent 能力包仓库、远程安装、配置下发、数据到达验证 | 本地源码/版本登记、包签名、安装脚本、兼容矩阵 |

## 9. 后续每个切片必须补的页面级字段

每个页面/功能切片进入编码前，必须在本目录追加页面级矩阵，字段固定为：

| 字段 | 要求 |
| --- | --- |
| 参考源码路径 | 精确到路由、页面、组件、API 文件 |
| 参考运行态 URL | 精确到页面和关键操作路径 |
| 页面标题 | 用户侧 FindX 命名替换表 |
| 顶部操作区 | 主按钮、更多菜单、导入导出、刷新、时间范围 |
| 筛选区 | 搜索、下拉、标签、历史、收藏、字段筛选 |
| 主体结构 | 表格、图表、树、拓扑、Trace、日志列表、Panel |
| 表格列/图表区 | 列名、排序、分页、空态、错误态 |
| 抽屉/弹窗 | 新增、编辑、详情、导入、测试、确认、预览 |
| API 调用 | 请求、响应、错误码、权限、超时、取消请求 |
| 状态流 | loading、empty、error、success、blocked、dirty、selected |
| 按钮真实动作 | 每个按钮必须有源码动作和 FindX 映射 |
| 权限态 | 401、403、无权限按钮隐藏/禁用 |
| 错误态 | 外部错误脱敏，组件不可用显示 `BLOCKED` |
| 验证 | WSL 构建、MCP 登录点击、正常/异常/边界或权限 |

未补齐以上字段时，不允许进入代码实现，也不允许写 PASS。
