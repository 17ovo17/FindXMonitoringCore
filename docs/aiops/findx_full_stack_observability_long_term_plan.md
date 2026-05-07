# FindX 全栈可观测长期开发计划

生成时间：2026-05-07 09:09（UTC+8）  
状态：当前唯一长期实施主计划  
适用范围：FindX 平台前端、后端、Agent、数据源、观测链路、AI SRE、知识库、测试与合规

## 1. 总则

FindX 的长期路线是：基于用户提供的成熟平台源码，复用成熟产品的信息架构、页面结构、组件状态流、API 语义和功能闭环，再统一替换为 FindX 自有品牌、风格、权限、审计、配置、数据源和 AI SRE 能力。

本计划明确废止“直接 iframe 嵌入参考站点”“嵌入已有系统 SSO”“只复刻截图”“自研弱化页面”“最小实现”“最小验证”“静态假按钮”等方向。参考站点只作为运行态对照和验收证据；真正实现必须从源码、路由、组件、状态流、API 调用和运行态 DOM 中提取事实。

用户侧只能出现 FindX、FindX Agent、链路监控、日志中心、Agent 管理中心、AI SRE 等 FindX 命名。外部来源名称只允许出现在内部开发证据、合规来源登记、授权边界和归档文档中。

## 2. 不可违背的实施原则

- 不做 iframe 嵌入参考站，不嵌入参考站登录、导航或 SSO。
- 不做 MVP、不做占位页、不做静态按钮、不做最小实现、不做最小验证。
- 不因为“先跑通”而阉割成熟源码已有的功能点、状态流、抽屉、弹窗、筛选、批量操作、历史记录、联想、模板导入、变量选择、分享、放大、三点菜单、启停、克隆、删除、导入导出。
- 不以截图为事实源；每个页面切片必须读取成熟源码和运行态 DOM，记录源码路径、路由、核心组件、API 调用、状态流、按钮真实动作、空态、错误态、权限态和 FindX 替换点。
- 不把外部产品原 API 直接作为 FindX 对外公共契约暴露给业务前端；必须经 FindX Adapter、权限、审计、错误脱敏和统一配置层收敛。
- 不在用户侧菜单、页面标题、权限对象、审计对象、通知内容中暴露外部品牌。
- 不在文档、日志、测试报告、prompt 中写真实 token、cookie、Bearer、DSN、SSH 私钥、完整连接串或会话 ID。
- 不删除历史证据；旧方向先归档或标记 superseded，截图、JSON、DOM snapshot 和测试产物先列候选清单，再人工确认。

## 3. 总体架构路线

FindX 前端采用 React-first 方向重构壳层和主要观测工作台。若现有 Vue 页面仍承担历史功能，按切片逐步下线或替换，不继续在弱化 Vue 页面上堆补丁。成熟源码若为 React 体系，优先使用同技术体系迁移组件结构、状态流和交互语义，以降低翻译损耗。

FindX 平台由以下层组成：

- FindX Web Shell：自有登录、导航、主题、权限、审计上下文、全局错误、登录过期跳转。
- FindX API Gateway：统一认证、权限、租户、团队、审计、错误脱敏、限流和 Adapter 路由。
- Source-Compatible UI Layer：按成熟源码复刻页面结构、组件拆分、状态流和功能点，只替换视觉和品牌。
- Adapter Layer：基础监控、链路监控、日志中心、CMDB、Agent、工作流、模板、知识库的统一后端适配层。
- Data Source Center：统一配置 Prometheus/VictoriaMetrics、MySQL/MariaDB、Redis、MinIO、Qdrant、SkyWalking OAP、BanyanDB、ClickHouse、OTel Collector、可选 ES/OpenSearch 等依赖。
- FindX Agent Control Plane：统一安装、升级、回滚、配置下发、心跳、健康、审计、包仓库和生命周期管理。
- AI SRE 与 Evidence Chain：只基于真实证据诊断，不编造缺失数据。

## 4. 导航规划

FindX 主导航按成熟观测平台的信息架构收敛，同时加入 FindX 增强域：

| 一级导航 | 二级能力 | 事实源/实现依据 | 用户侧命名 |
| --- | --- | --- | --- |
| 基础设施 | 业务组、资源、主机、CMDB、Agent 在线、Agent 安装、Agent 包管理 | 基础监控源码 + AutoOps/AIOps CMDB + FindX Agent | 基础设施 / CMDB / Agent 管理 |
| 集成中心 | 数据源、采集模板、插件、模板中心、接入指引 | 基础监控源码 + Categraf 能力包 | 集成中心 / 模板中心 / FindX Agent 插件 |
| 数据查询 | 指标查询、日志查询、Trace 查询、Saved Views、查询历史 | 基础监控源码 + SigNoZ + SkyWalking | 数据查询 |
| 仪表盘 | 列表、详情、变量、时间、刷新、Panel、分享、导入导出、模板导入 | 基础监控源码 | 仪表盘 |
| 告警 | 规则、事件、屏蔽、订阅、自愈、记录规则、事件流水线 | 基础监控源码 + AI SRE 工作流 | 告警 |
| 通知 | 通知规则、通知媒介、消息模板、升级策略、测试发送 | 基础监控源码 | 通知 |
| 链路监控 | 服务目录、拓扑、Trace、慢调用、错误链路、Profiling、告警、接入 | SkyWalking UI/OAP/Agent 源码 | 链路监控 |
| 日志中心 | 日志检索、字段筛选、上下文、聚合、Pipeline、Saved Views、Trace 关联 | SigNoZ 前端和 API 源码 | 日志中心 |
| Agent 管理中心 | 包仓库、远程安装、配置下发、灰度、心跳、版本、漂移、回滚、卸载 | FindX Agent 控制面 + SkyWalking Agent 生态 + Categraf/Catpaw | Agent 管理中心 |
| AI SRE | 诊断会话、工作流、健康检查、复盘报告、证据链、自动修复编排 | FindX 增强层 + 成熟平台事件/指标/日志/Trace/CMDB | AI SRE |
| 组织权限 | 用户、团队、角色、权限、SSO、审计对象 | 基础监控源码 + FindX IAM | 组织权限 |
| 平台治理 | 全局设置、AI 模型配置、审计日志、数据保留、组件状态、授权边界 | FindX 平台治理 | 平台治理 |

## 5. 数据源与配置一致性

所有平台组件使用统一配置中心和数据源中心。配置必须分层管理：全局配置、环境配置、租户/团队配置、数据源配置、凭据引用、运行时覆盖。

| 组件 | 默认定位 | 统一配置要求 |
| --- | --- | --- |
| MariaDB/MySQL | 权威业务数据、用户、权限、配置、知识库元数据 | FindX authoritative store；迁移脚本、备份、回滚必须纳入 |
| Redis | 缓存、会话、队列、限流、短期状态 | 必须有过期策略、命名空间、降级策略 |
| MinIO/S3 | 截图、报告、附件、证据产物、离线包 | 对象路径不含敏感信息；支持生命周期策略 |
| Qdrant | 知识库向量索引加速层 | MySQL 是权威基石，Qdrant 可重建；支持索引状态和重建任务 |
| BM25 | 关键词检索与向量不可用降级 | 与向量检索通过 RRF 融合 |
| Prometheus/VictoriaMetrics | 指标查询与告警 | 统一数据源配置、测试连接、超时、错误脱敏 |
| SkyWalking OAP | 链路、服务、拓扑、Profiling、APM 告警 | 通过 FindX APM Adapter 访问，不直暴 GraphQL 给业务前端 |
| BanyanDB | SkyWalking 推荐链路存储 | 默认链路存储候选；保留 ES/OpenSearch 兼容选项 |
| SigNoZ Query Service | 日志、Trace/日志关联、查询视图 | 通过 FindX Logs Adapter 访问 |
| ClickHouse | SigNoZ 日志和事件分析存储 | 不作为 SkyWalking OAP 的直接替代存储；通过日志域使用 |
| OTel Collector | Telemetry 接入、协议转换、路由 | FindX 接入层统一管理 pipeline、采样、脱敏 |
| ES/OpenSearch | 可选搜索/链路兼容存储 | 仅在明确需要 SkyWalking 兼容或日志搜索场景时部署 |

配置文件必须支持本地开发、测试、生产、离线交付四类场景。真实凭据只以引用方式保存和传递，前端不回显密文，错误不显示完整连接串。

## 6. 成熟源码事实源矩阵

| 域 | 源码事实源 | FindX 实施方式 |
| --- | --- | --- |
| 基础监控页面 | `D:\平台源码\fe-main` | 同源复刻页面结构、路由、状态流、API 语义；替换为 FindX 壳层、品牌、主题、权限、审计 |
| 链路监控 | `D:\平台源码\skywalking-booster-ui-main`、`D:\平台源码\skywalking-master` | 复刻服务目录、拓扑、Trace、Profiling、告警和接入流程；通过 APM Adapter 统一 OAP 查询 |
| 日志中心 | `D:\平台源码\signoz-develop\frontend` | 复刻日志检索、字段筛选、上下文、聚合、Pipeline、Saved Views、Trace 关联 |
| CMDB 与 Agent 在线 | `D:\平台源码\AutoOps-main\AutoOps-main` | 复刻 CMDB 树、主机表、分组、Agent 在线、部署、心跳、统计、终端/监控弹窗 |
| FindX Agent 采集插件 | `D:\平台源码\categraf-main (1)` | 作为 FindX Agent 插件能力包来源；用户侧统一称 FindX Agent |
| 巡检诊断 | `D:\平台源码\catpaw-master` | 作为 FindX Agent 诊断、结构化执行、巡检、Evidence Chain 能力来源；用户侧统一称 FindX Agent |

编码前必须输出源码证据清单。没有源码证据的切片不得进入实现。

## 7. 页面与功能复刻要求

### 7.1 数据源

必须包含类型选择、配置示例、URL、认证引用、Header、TLS、超时、测试连接、状态、启停、编辑、删除、权限、审计、错误脱敏和导入导出。不得把成熟配置流程压成一个自研复杂抽屉。

### 7.2 指标查询

必须按成熟源码语义实现数据源选择、PromQL 输入、联想、内置指标、历史记录、添加面板、即时查询、区间查询、Table/Graph、Time、Unit、CSV 导出、空结果、非法查询、超时、取消请求。添加面板是新增 query panel，不是收藏。

### 7.3 仪表盘

必须包含列表、分组、详情、变量、变量搜索、变量编辑、数据源选择、时间范围、刷新、Panel、三点菜单、分享、放大、查询参数、添加图表、编辑、复制、删除、导入、导出、模板导入。变量不能做静态标签。

### 7.4 模板中心

必须包含图标、分类、说明、适用组件、dashboard、alert、record-rule、预览、导入、冲突处理、导入结果、失败回滚。同类监控模板按成熟产品方式聚合；`categraf-agent` 等用户侧文案统一替换为 `findx-agent` 或 FindX Agent，但不能删除模板含义。

### 7.5 告警与通知

告警规则、屏蔽、订阅、事件、事件流水线、告警自愈、通知规则、通知媒介、消息模板必须按成熟源码结构实现。不能把订阅、屏蔽、模板、事件流水线混成一个自研页面。

### 7.6 链路监控

必须覆盖服务目录、服务实例、端点、拓扑、Trace 检索、Trace 详情、慢调用、错误链路、Profiling、告警、接入指引、Agent 状态、采样与标签。FindX 后端通过 APM Adapter 调用 OAP/query protocol。

#### 7.6.1 SkyWalking 前端链路不可省略

链路监控前端不能只写后端 OAP Adapter，也不能退化成 FindX 自研静态拓扑页。必须按 SkyWalking Booster UI 的真实前端链路迁移或同源状态流替换：

- 路由事实源：`D:\平台源码\skywalking-booster-ui-main\src\router\index.ts` 汇总 `routesDashboard`、`routesLayers`、`routesAlarm`、`routesTrace`；`src\router\constants.ts` 定义 `/dashboard/list`、`/dashboard/new`、`/dashboard/:layerId/:entity/:serviceId/:name`、`/traces/:traceId`、`/alerting` 等路径；`src\router\trace.ts` 将 Trace 详情挂到 `views\dashboard\Trace.vue`。
- 页面事实源：`src\views\dashboard\List.vue`、`New.vue`、`Edit.vue`、`Widget.vue`、`Trace.vue`、`src\views\Layer.vue`、`src\views\Alarm.vue`、`src\views\alarm\*` 是链路工作台、看板、Trace、告警视图的页面结构事实源。
- 状态流事实源：`src\store\modules\trace.ts`、`topology.ts`、`dashboard.ts`、`alarm.ts`、`profile.ts`、`async-profiling.ts`、`continous-profiling.ts`、`network-profiling.ts` 是前端状态事实源。FindX 如果迁移到 React，必须把 Pinia 状态、loading、selector、duration、trace condition、current trace/span、topology graph、dashboard session 等语义完整迁移到 React store，不得只保留页面外观。
- 查询事实源：`src\graphql\base.ts` 的基础路径为 `/graphql`，`src\graphql\query\trace.ts`、`topology.ts`、`dashboard.ts`、`alarm.ts`、`profile.ts`、`async-profile.ts`、`ebpf.ts`、`selector.ts` 和对应 `fragments\*` 是 GraphQL 请求和字段事实源。FindX 前端不得直接暴露外部 GraphQL 给业务页面，必须经 FindX APM Adapter 代理，但请求语义、字段映射、错误态、超时和取消请求要与源码一致。
- FindX 路由映射：用户侧显示 `/tracing` 域和“链路监控”命名，内部必须保留服务/实例/端点/层级拓扑/Trace 详情/Profiling/告警/设置/接入的同源页面状态。允许增加 FindX alias，例如 `/tracing?section=overview`、`/tracing/services`、`/tracing/topology`、`/tracing/traces/:traceId`、`/tracing/profiling`、`/tracing/alarms`，但 alias 必须落回同源状态流，不得另造静态页面。
- FindX Shell 边界：React-first 壳层只负责登录、导航、主题、权限、审计、错误脱敏和品牌替换；链路页面内部结构、图表工作台、Trace 详情、拓扑交互、Profiling 任务、告警视图、时间选择器、GraphQL 状态和错误处理必须按 SkyWalking Booster UI 源码实现或等价迁移。
- 组件不可用：未配置 OAP、BanyanDB、GraphQL proxy、APM Adapter 或数据源不可达时，前端必须显示明确 `BLOCKED` / 503 脱敏错误和重试入口，不得显示空壳拓扑、假 Trace 或静态列表。
- 验收入口：P3 之前必须先在 P0 源码矩阵中记录 SkyWalking 前端路由、页面、store、GraphQL query、状态流、按钮动作、空态、错误态、权限态和 FindX 替换点；P3 开发完成后必须用 MCP 浏览器覆盖服务目录、拓扑、Trace 详情、Profiling、告警、配置缺失 BLOCKED 和窄屏回归。

### 7.7 日志中心

必须覆盖日志检索、字段筛选、上下文、聚合、Pipeline、Saved Views、Trace 关联、导出、错误态、空态、权限态。日志中心通过 Logs Adapter 统一 ClickHouse/SigNoZ 查询。

#### 7.7.1 SigNoZ 前端链路不可省略

日志中心不能只接一个后端日志查询 API，也不能把 SigNoZ 的 Logs Explorer 弱化成静态表格。必须按 SigNoZ 前端源码的真实路由、页面、查询构造和 API 状态流迁移或同源实现：

- 路由事实源：`D:\平台源码\signoz-develop\frontend\src\AppRoutes\routes.ts` 绑定 `Logs`、`LogsExplorer`、`LiveLogs`、`PipelinePage`、`LogsSaveViews`、`TracesExplorer`、`TracesSaveViews`、`TraceDetail`；`src\constants\routes.ts` 定义 `/logs`、`/logs/logs-explorer`、`/logs/logs-explorer/live`、`/logs/pipelines`、`/logs/saved-views`、`/trace/:id`、`/traces-explorer`、`/traces/saved-views`。
- 页面事实源：`src\AppRoutes\pageComponents.ts` 将日志和 Trace 能力映射到 `pages\LogsModulePage`、`pages\LiveLogs`、`pages\TracesModulePage`、`pages\TraceDetail`；`src\components\Logs\*`、`LogDetail\*`、`QuickFilters\*`、`TabsAndFilters\*`、`ExplorerCard\*`、`ExplorerControlPanel\*`、`ExplorerOptions\*` 是日志结果、详情、字段筛选、快捷过滤、保存视图和查询控制面的结构事实源。
- API 事实源：`src\api\logs\GetLogs.ts` 调用 `/logs`，`GetLogsAggregate.ts` 负责聚合，`GetSearchFields.ts` 负责字段，`livetail.ts` 负责实时日志；`src\api\pipeline\get.ts`、`post.ts`、`preview.ts` 对应 `/logs/pipelines/:version` 等 Pipeline 流程；`src\api\saveView\getAllViews.ts`、`saveView.ts`、`updateView.ts`、`deleteView.ts` 对应 `/explorer/views`；`src\api\trace\getTraceItem.ts` 调用 `/traces/:id`，`getSpans.ts` 调用 `/getFilteredSpans/aggregates`。
- 查询状态流：必须保留 query builder、时间范围、字段选择、selected field、quick filter、exclude filter、聚合、分页、live tail、raw/table/list 视图、上下文跳转、保存视图、Pipeline preview、Trace 关联、错误态、空态、加载态和取消请求。
- FindX 路由映射：用户侧显示 `/logs` 域和“日志中心”命名，允许提供 `/logs?section=explorer`、`/logs/explorer`、`/logs/live`、`/logs/pipelines`、`/logs/saved-views`、`/logs/trace/:id` 等 FindX alias，但 alias 必须落回 SigNoZ 同源状态流，不得另造静态日志表。
- FindX Adapter 边界：前端不得直接把 SigNoZ 原 API 暴露为 FindX 公共契约；FindX Logs Adapter 负责权限、租户、审计、ClickHouse/SigNoZ 查询、错误脱敏和组件健康。Adapter 字段可以重命名为 FindX 契约，但必须完整覆盖 SigNoZ 页面所需字段。
- 组件不可用：未配置 Query Service、ClickHouse、OTel Collector、Logs Adapter 或日志数据源不可达时，日志页面必须显示明确 `BLOCKED` / 503 脱敏错误和重试入口，不得显示假日志、假 Pipeline 或静态 Saved Views。
- 验收入口：P4 之前必须在 P0 源码矩阵中记录 SigNoZ 路由、页面、API、查询状态、按钮动作、空态、错误态、权限态和 FindX 替换点；P4 完成后必须用 MCP 浏览器覆盖日志检索、字段筛选、上下文、聚合、live tail、Pipeline preview、Saved Views、Trace 关联、组件缺失 BLOCKED 和窄屏回归。

### 7.8 CMDB 与 Agent 在线

必须覆盖 CMDB 树、模型、实例、主机表、分组、标签、批量操作、Agent 在线状态、部署状态、心跳、版本、统计、终端/监控相关弹窗。不得重做弱化 CMDB。

## 8. FindX Agent 长期方案

FindX Agent 是控制面和安装器，不强行把所有探针揉成一个运行时二进制。FindX Agent 管理多语言探针、采集插件、巡检工具、浏览器探针、网关插件和配置包的生命周期。

### 8.1 SkyWalking Agent 生态纳入

以下能力作为 FindX Agent 管理的能力包纳入：

- Java Agent
- Python Agent
- Node.js Agent
- PHP Agent
- Go Agent
- Rust Agent
- Ruby Agent
- Nginx Lua Agent
- Kong Agent
- Browser Client JS

以上不是简单列名，P0 源码矩阵必须把每个 Agent 的上游仓库、源码路径、制品形态、配置项、安装方式、回滚方式和数据到达验证全部登记清楚。当前本地只确认已有 `D:\平台源码\skywalking-booster-ui-main` 和 `D:\平台源码\skywalking-master`；如果独立 Agent 源码尚未落到 `D:\平台源码`，必须在 P0 阶段补齐源码目录或明确标记 `BLOCKED`，不得用包名或静态说明冒充完成。

| 能力包 | 上游源码事实源 | P0 必须记录 | FindX 实现要求 |
| --- | --- | --- | --- |
| Java 探针 | `https://github.com/apache/skywalking-java` | 插件目录、JVM 参数、配置文件、服务名、采样、日志关联、升级回滚 | FindX Agent Java 探针，支持远程安装、配置下发、心跳、数据到达 |
| Python 探针 | `https://github.com/apache/skywalking-python` | 运行时版本、启动注入、环境变量、依赖安装、异步框架兼容 | FindX Agent Python 探针，支持 pip/离线包、进程检测、回滚 |
| Node.js 探针 | `https://github.com/apache/skywalking-nodejs` | npm 包、启动参数、框架插件、环境变量、错误上报 | FindX Agent Node.js 探针，支持服务重启、配置漂移检测 |
| PHP 探针 | `https://github.com/apache/skywalking-php` | 扩展安装、php.ini、FPM/Apache 适配、版本矩阵 | FindX Agent PHP 探针，支持扩展启停、健康检查 |
| Go 探针 | `https://github.com/apache/skywalking-go` | 编译/注入方式、模块依赖、构建流水线、回滚策略 | FindX Agent Go 探针，支持构建期提示和运行态接入校验 |
| Rust 探针 | `https://github.com/apache/skywalking-rust` | crate、运行时、配置项、服务名、Trace 传播 | FindX Agent Rust 探针，支持配置模板和数据到达校验 |
| Ruby 探针 | `https://github.com/apache/skywalking-ruby` | gem、框架兼容、启动注入、错误上报 | FindX Agent Ruby 探针，支持远程配置和版本治理 |
| Nginx Lua 探针 | `https://github.com/apache/skywalking-nginx-lua` | OpenResty/Nginx 版本、Lua 模块、网关配置、重载策略 | FindX Agent 网关探针，支持配置校验、灰度重载、回滚 |
| Kong 插件 | `https://github.com/apache/skywalking-kong` | Kong 版本、插件安装、路由/服务绑定、重载策略 | FindX Agent Kong 探针，支持插件分发和接入验证 |
| Browser Client JS | `https://github.com/apache/skywalking-client-js` | SDK 引入、隐私脱敏、采样、SourceMap、会话/错误/性能事件 | FindX Browser Agent，支持前端接入向导、数据脱敏和 RUM 验收 |

P0 源码矩阵阶段继续审计并补齐：.NET、C++、Rover/eBPF、SWCK、Satellite、移动端/小程序接入、网格/Sidecar 场景。

#### 8.1.1 SkyWalking Agent 控制面不可省略

SkyWalking Agent 不只是能力包清单，必须进入 FindX Agent 管理中心的安装、配置、心跳、升级、回滚和证据链：

- 包目录：Java、Python、Node.js、PHP、Go、Rust、Ruby、Nginx Lua、Kong、Browser Client JS 必须作为 FindX Agent 能力包展示，用户侧命名为 FindX Agent Java 探针、FindX Agent Python 探针、FindX Browser Agent 等，不显示外部品牌。
- 远程安装：Linux 必须支持 SSH/shell/systemd/Docker/Kubernetes 注入；Windows 必须支持 WinRM/PowerShell/Windows Service/IIS/.NET 相关探针场景；Kubernetes 必须支持 Helm/Operator/Webhook/DaemonSet/Sidecar/InitContainer/RBAC/TLS。
- 配置模板：必须覆盖 OAP endpoint、service name、instance name、namespace、environment、sampling、ignore suffix、plugin include/exclude、log correlation、trace propagation、meter/log/reporting、TLS/proxy、资源限制和标签。
- 配置下发：支持单机、批量、业务组、Kubernetes namespace/workload、灰度批次、维护窗口、失败阈值、自动回滚和配置漂移检测。
- 心跳与数据到达：Agent 在线不等于链路数据可用。必须分别展示控制面心跳、进程/服务状态、探针加载状态、OAP 连通性、Trace 到达、指标到达、日志关联到达和最后上报时间。
- 版本治理：必须展示语言、运行时、Agent 版本、插件版本、配置版本、升级建议、兼容风险、已知问题和回滚目标。
- 安全审计：远程命令、包下载、签名校验、安装脚本、配置变更、重启、回滚、卸载必须进入审计和 Evidence Chain，敏感参数使用 `<TOKEN>`、`<DB_DSN>`、`<API_KEY>` 等占位符。
- 前端入口：Agent 管理中心必须提供包仓库、安装向导、主机/服务选择、Kubernetes 注入、配置模板、下发任务、心跳状态、版本分布、漂移检测、升级回滚、卸载记录和接入验证页面，不能只写后台任务。
- 验收入口：P5 完成前必须用 MCP 浏览器覆盖 Linux、Windows、Kubernetes 至少各一个安装计划流程、配置下发、心跳异常、数据未到达 BLOCKED、升级回滚和卸载确认。

### 8.2 远程安装与系统适配

Linux：

- SSH、sudo、shell、systemd、Docker、离线包、代理网络、内网仓库。
- 安装、升级、回滚、卸载、状态检查、日志采集、端口检测、连通性验证。

Windows：

- WinRM、PowerShell、Windows Service、任务计划、IIS/.NET 注入、日志采集、端口检测、权限检测。
- 支持 GUI-less 远程安装、离线包、企业代理、杀软拦截提示。

Kubernetes：

- Helm、Operator、Admission Webhook、DaemonSet、Sidecar、InitContainer、RBAC、ServiceAccount、TLS、Secret、ConfigMap。
- 支持命名空间级灰度、工作负载选择器、自动注入、卸载清理和回滚。

### 8.3 商业化 Agent 控制面能力

- 覆盖率分析：主机、服务、语言、集群、命名空间、业务组。
- 接入成功率：安装任务、探针启动、连通性、数据到达。
- 版本分布：Agent 包、插件、探针、配置模板。
- 配置漂移：期望配置、实际配置、差异、修复建议。
- 灰度发布：批次、比例、业务组、维护窗口、失败阈值、自动回滚。
- 离线仓库：包签名、校验、制品镜像、内网同步。
- 心跳与健康：在线、离线、异常、心跳延迟、采集延迟、上报状态。
- 回滚与卸载：版本回退、配置回退、残留清理、审计记录。
- 安全：最小权限、命令白名单、脚本签名、执行审计、敏感输出脱敏。

Agent 生命周期完成定义：安装计划生成、凭据引用、远程执行、包校验、服务注册、配置下发、心跳上报、数据到达、异常恢复、升级、回滚、卸载、审计和 Evidence Chain 全部闭环。

## 9. AI SRE 与 Evidence Chain

AI SRE 不直接替代监控、日志、Trace、CMDB、Agent 或工作流能力；它只在已有证据上做诊断、解释、编排、复盘和修复建议。

Evidence Chain 必须接入：

- 指标查询结果、告警触发值、规则版本。
- 日志查询、字段筛选、上下文片段。
- Trace、Span、服务拓扑、慢调用、错误链路。
- CMDB 实例、主机、业务组、变更、Agent 心跳。
- 巡检任务、远程执行、脚本输出、文件证据。
- 工作流执行记录、人工审批、修复动作、回滚动作。
- 知识库引用、Runbook 版本、模型配置引用。

证据缺失时必须返回“数据缺失/组件不可用/BLOCKED”，不得编造结论。所有 AI 调用必须走系统配置中的 AI 模型配置，包括模型、凭据引用、超时、重试、审计策略、脱敏策略和成本控制。

## 10. 知识库与检索架构

知识库采用 MySQL/MariaDB 做权威基石，Qdrant 做内置向量数据库加速层，BM25 做关键词检索和向量不可用降级，RRF 做混合召回融合。

- MySQL/MariaDB 保存文档、段落、权限、版本、来源、审计、引用关系和索引状态。
- Qdrant 保存可重建的向量索引，不作为权威数据源。
- BM25 保存关键词倒排索引，支持无向量环境和精确关键词召回。
- 索引任务必须有状态、进度、失败原因、重试、重建、增量同步。
- 查询必须同时验证权限、租户、团队、文档状态和敏感信息脱敏。
- Evidence Chain 引用知识库时必须记录文档版本、段落 ID、召回方式、相关度和模型配置引用。

## 11. 兼容性与部署策略

- 单机开发：MariaDB、Redis、MinIO、Qdrant、Prometheus/VictoriaMetrics 可本地启动，链路/日志组件可显示明确 BLOCKED。
- 标准生产：MariaDB、Redis、MinIO、Qdrant、VictoriaMetrics、SkyWalking OAP+BanyanDB、SigNoZ+ClickHouse、OTel Collector。
- 大规模生产：支持多副本、分片、冷热分层、队列削峰、限流、审计归档、对象存储生命周期。
- 离线交付：支持离线包、离线 Agent 仓库、镜像仓库、证书、许可证和升级包。
- 可选兼容：ES/OpenSearch 只在链路兼容或日志搜索明确需要时部署，不作为默认必选项。

## 12. 还必须补齐的长期能力

- SLO/SLA：目标、错误预算、燃尽率、多窗口告警。
- 事件管理：值班、升级、分派、确认、恢复、复盘、行动项。
- 变更关联：发布、配置、CMDB、告警、Trace、日志关联。
- Synthetic Monitoring：HTTP、TCP、DNS、浏览器脚本、地域探测。
- 云资源观测：主流云账号、资源同步、成本、标签、告警。
- Kubernetes 观测：集群、节点、Pod、Workload、Service、Ingress、事件、资源请求与限制。
- 网络观测：拓扑、链路质量、DNS、端口、连接、流量。
- RUM 与移动端：Browser Client JS、移动端、小程序、前端错误、性能、会话。
- 数据治理：采样、脱敏、保留、冷热分层、导出、删除、合规审计。
- 权限治理：租户、团队、业务组、资源级权限、审批、审计。
- 可靠性治理：组件健康、队列积压、任务失败、告警风暴、降级模式。

## 13. 实施阶段

### P0-SOURCE-MATRIX-LOCK

只读审计所有成熟源码入口和参考站 DOM，输出页面/路由/API/状态流矩阵。没有矩阵不得进入编码。

### P0-CONFIG-DATASOURCE-CONTRACT

统一配置文件、统一数据源中心、凭据引用、测试连接、错误脱敏、组件 BLOCKED 状态和审计模型。

### P0-WEB-REACT-SHELL-NAV

建设 React-first FindX 自有壳层、登录、导航、主题、权限、审计和全局错误。禁止嵌入参考站。

### P0-BASE-MONITORING-REAL

按成熟源码重做数据源、指标查询、仪表盘、模板中心、告警、通知、系统配置等基础监控页面。

### P1-ORG-PERMISSION-GOVERNANCE

组织、团队、角色、权限、SSO、审计日志、站点设置、AI 模型配置、平台治理。

### P2-CMDB-AGENT-REAL

按 AutoOps/AIOps 源码接入 CMDB、主机、Agent 在线、部署、心跳、统计和终端/监控弹窗。

### P3-APM-SKYWALKING-REAL

按 SkyWalking UI/OAP/Agent 源码接入链路监控、服务、拓扑、Trace、Profiling、告警和接入。前端必须包含 SkyWalking Booster UI 的路由、页面、GraphQL query、store 状态流和图表/Trace/拓扑真实交互；FindX React-first 壳层只做品牌、导航、权限、审计和主题，不得把链路监控改成 iframe、静态拓扑或后端-only Adapter。

### P4-LOGS-SIGNOZ-REAL

按 SigNoZ 源码接入日志检索、字段筛选、上下文、聚合、Pipeline、Saved Views 和 Trace 关联。前端必须包含 SigNoZ 的 logs/traces routes、LogsModulePage、LiveLogs、TracesModulePage、TraceDetail、query builder、字段筛选、Saved Views、Pipeline preview 和 API 状态流；FindX Logs Adapter 只做权限、审计、错误脱敏和数据源收敛。

### P5-FINDX-AGENT-SUITE

接入多语言探针、采集插件、巡检工具、远程安装、配置下发、包仓库、灰度、回滚和卸载。SkyWalking Agent 生态必须进入 FindX Agent 管理中心的包目录、安装向导、配置模板、心跳、数据到达验证、版本治理、漂移检测、升级回滚、卸载和 Evidence Chain。

编码前门禁参考：[P4 Agent Suite 同源矩阵](source-matrix/p4_categraf_catpaw_agent_suite_matrix.md) 和 [P0 SkyWalking Agent 到 FindX Agent 控制面矩阵](source-matrix/p0_skywalking_agent_real_matrix.md)。

### P6-AISRE-EVIDENCE-CHAIN

完成 Evidence Chain、诊断会话、工作流、复盘报告、自动修复编排、AI 模型统一配置。

编码前门禁参考：[P5 AI SRE / Evidence Chain 同源矩阵](source-matrix/p5_aisre_evidence_chain_matrix.md)。

### P7-KNOWLEDGE-VECTOR

完成 MySQL 权威知识库、Qdrant 向量索引、BM25、RRF、索引任务、权限和引用链。

编码前门禁参考：[P6 知识库 / Qdrant 向量索引同源矩阵](source-matrix/p6_knowledge_qdrant_vector_matrix.md)。源码矩阵文件名按文档闭环顺序落地，实施阶段按本主计划推进。

### P8-SLO-INCIDENT-CHANGE

补齐 SLO/SLA、事件管理、变更关联、升级策略和复盘行动项。

### P9-CLOUD-K8S-NETWORK-RUM

补齐云资源、Kubernetes、网络观测、RUM、移动端和 Synthetic Monitoring。

### P10-HARDENING-COMMERCIALIZATION

完成性能、容量、数据保留、冷热分层、离线交付、授权边界、升级回滚、合规材料和商业化验收。

## 14. 每个切片的准入清单

编码前必须提交：

- 参考源码路径。
- 路由和页面入口。
- 核心组件。
- API 调用和请求/响应字段。
- 状态流和关键状态枚举。
- 按钮真实动作。
- 抽屉/弹窗/详情页。
- 空态、错误态、权限态、加载态。
- FindX 命名替换表。
- 是否 API_CONTRACT_CHANGE。
- 是否 DATA_CHANGE。
- 回滚策略。
- MCP 浏览器验证路径。

## 15. 测试与验收

每个 UI/API 切片必须覆盖：

- WSL/Linux 构建或对应后端测试。
- MCP 浏览器真实登录、点击、表单、抽屉/弹窗、错误态、窄屏回归。
- 正常路径、异常路径、边界或权限路径。
- 组件不可用时明确显示 BLOCKED，不得伪装成功。
- 成熟源码中有真实动作的控件，在 FindX 中必须接入真实动作；未完成直接 FAIL。
- API 测试不能替代浏览器真实交互。
- 未真实执行不得写 PASS。
- P0/P1 RISK、敏感信息泄露、权限绕过、错误未脱敏、静态假按钮、乱码、外部品牌用户侧暴露全部阻断。

## 16. Git 与文档治理

- 本计划是当前唯一长期实施主计划。
- 旧方向文档全部标记 superseded 或列入归档索引，不再作为实施依据。
- 历史证据、截图、DOM snapshot、JSON、测试报告不直接删除；先建立候选清单并检查引用。
- 每次计划变更必须同步 README 和 `docs/aiops/README.md`。
- 业务代码、运行态配置、密钥、Cookie、DSN、会话 ID 不得进入文档计划提交。

## 17. 下一轮开发入口

文档提交后，从 `P0-SOURCE-MATRIX-LOCK` 开始闭环：

1. 基础监控页面/路由/API/状态流矩阵。
2. SkyWalking OAP/UI/Agent 仓库矩阵。
3. SigNoZ 日志中心矩阵。
4. AutoOps/AIOps CMDB 矩阵。
5. Categraf/Catpaw 到 FindX Agent 的能力矩阵。
6. 统一配置与数据源契约。
7. React-first FindX 自有壳层和导航。

任何实现切片都必须先通过源码证据门禁。
