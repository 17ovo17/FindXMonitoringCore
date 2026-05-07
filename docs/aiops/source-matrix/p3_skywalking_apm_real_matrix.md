# FindX P3 链路监控同源矩阵

生成时间：2026-05-07 14:10（UTC+8）
状态：`P3-SKYWALKING-APM-REAL` 编码前门禁，不代表 FindX 已完成实现

## 1. 结论

链路监控不能只做后端 OAP Adapter，也不能只保留一个 Trace 搜索框。FindX 必须按成熟 UI 源码迁移服务目录、层级菜单、Dashboard、拓扑、Trace、日志关联、Profiling、告警、设置和 OAP GraphQL 状态流；用户侧统一显示“链路监控”，外部来源名称只保留在内部源码矩阵和合规登记。

源码证据：

- [SkyWalking 链路监控源码检查证据](evidence/skywalking_apm_source_snapshot.md)

当前阻断：

- UI/OAP/WebApp 源码存在。
- 未配置可访问 OAP / UI 运行态，运行态 DOM 证据为 `BLOCKED_RUNTIME_UNAVAILABLE`。

## 2. FindX 路由规划

| FindX 页面 | 成熟路由/源码 | 页面结构要求 |
| --- | --- | --- |
| 链路监控总览 | `src\router\layer.ts` + OAP `getMenuItems` | 动态 layer 菜单、分组、时间范围、自动刷新 |
| 服务目录 | `Layer.vue`、`selectors.ts`、`listServices` | 服务列表、layer、状态、正常/异常、搜索 |
| 服务详情 | `dashboard\Edit.vue` | 指标面板、时间范围、实例、端点、拓扑、Trace、日志、Profiling tabs |
| 拓扑 | `dashboard\related\topology` | service/pod/process topology、图交互、关联 Trace |
| Trace 查询 | `dashboard\related\trace`、`dashboard\Trace.vue` | Trace 列表、筛选、时间线、Span 树、Span 详情、tag autocomplete |
| 慢调用/错误链路 | Trace 查询条件 + Dashboard widgets | 按错误、耗时、端点、服务、实例过滤 |
| Profiling | `profile`、`async-profiling`、`continuous-profiling`、`pprof`、`ebpf`、`network-profiling` | 任务创建、任务列表、进度、结果、栈图、网络拓扑 |
| 告警 | `Alarm.vue`、`views\alarm` | 告警列表、筛选、快照、Trace ID 关联 |
| 设置 | `Settings.vue`、`views\settings` | TTL、集群节点、debug config dump、通用设置 |

FindX 不能把这些入口拆成低价值孤岛，也不能省略成熟 UI 中已有的真实控件。

## 3. UI 状态流门禁

| 状态流 | 成熟源码 | FindX 实现要求 |
| --- | --- | --- |
| 全局时间 | `store\modules\app.ts` | 时间范围、step、UTC、冷存储、自动刷新必须保留 |
| 菜单 | `getMenuItems` + `router\layer.ts` | OAP 动态菜单映射为 FindX 链路监控内部导航 |
| 选择器 | `store\modules\selectors.ts` | 服务、实例、端点、进程、目标服务选择器必须保留远程搜索和错误态 |
| GraphQL | `graphql\base.ts`、`graphql\index.ts` | 超时、取消请求、errors、非 2xx 响应不能假成功 |
| Dashboard | `store\modules\dashboard.ts` | 布局、widget、图表、编辑、视图模式按源码保留 |
| Trace | `store\modules\trace.ts`、`related\trace` | Trace V1/V2、Span、logs、attached events、tag autocomplete |
| Topology | `store\modules\topology.ts` | 节点、边、图布局、Trace 关联、服务关系 |
| Profiling | `profile`、`async`、`pprof`、`ebpf` stores | 任务状态、进度、结果、失败态 |

## 4. GraphQL / Adapter 契约

FindX 需要新增 APM Adapter / OAP proxy，但不把成熟 OAP GraphQL 原样暴露成业务前端公共 API。

内部 GraphQL 能力必须覆盖：

| 能力 | 成熟查询语义 | FindX Adapter 要求 |
| --- | --- | --- |
| OAP 信息 | version、time info、records TTL、metrics TTL | 初始化健康检查和能力探测 |
| 动态菜单 | `getMenuItems` | 映射成 FindX 链路监控内部导航 |
| 服务目录 | `listServices(layer)` | 权限过滤、业务组映射、错误脱敏 |
| 实例/端点/进程 | `listInstances`、`findEndpoint`、`listProcesses` | 保留远程搜索、duration、limit |
| Dashboard 指标 | dashboard fragments | Widget 查询、图表、单位、时间范围 |
| 拓扑 | topology fragments | service/pod/process topology |
| Trace | `queryBasicTraces`、`queryTrace`、`queryTraces` | 支持 Trace V1/V2、冷数据、超时、空结果 |
| Trace tag | `queryTraceTagAutocompleteKeys/Values` | tag 联想不能做静态控件 |
| 日志关联 | `queryLogs`、browser error log | 与 FindX 日志中心和 Evidence Chain 打通 |
| 告警 | `getAlarm`、alarm tags | 与 FindX 告警事件和 AI SRE 复盘打通 |
| Profiling | profile / pprof / ebpf / async fragments | 任务、进度、结果、失败态 |

错误模型：

- OAP 未配置：503 + `BLOCKED_OAP_NOT_CONFIGURED`
- OAP 超时：503 + `OAP_QUERY_TIMEOUT`
- GraphQL errors：502/503 + sanitized message
- 权限不足：403
- 登录过期：跳转 FindX 自有登录页
- 空结果：正常空态，不伪造数据

## 5. 存储与兼容性

链路监控和日志中心不能强行共用一个 UI 或一个存储模型。

| 能力 | 推荐路径 | 说明 |
| --- | --- | --- |
| 链路监控 | SkyWalking OAP + BanyanDB 优先，ES/OpenSearch 可作为兼容选项 | OAP query protocol 是链路事实源 |
| 日志中心 | SigNoZ Query Service + ClickHouse 优先 | 日志检索、Pipeline、Saved Views 按日志中心事实源 |
| Trace/Log 关联 | OTel trace/log 字段、trace_id、span_id、service、resource labels | 通过 FindX Adapter 做统一关联 |
| 搜索引擎 | OAP 内部 query protocol、日志中心 ClickHouse 查询、必要时 ES/OpenSearch 兼容 | 不能用一个搜索框替代成熟查询结构 |

是否部署 ES：

- SkyWalking 可使用 ES/OpenSearch 作为兼容存储，但不是必须唯一方案。
- SigNoZ 当前事实源通常围绕 ClickHouse 查询模型。
- FindX 应以统一数据源中心管理 OAP、BanyanDB、ClickHouse、ES/OpenSearch 可选项，不把存储配置散落到页面里。

## 6. 与 FindX Agent 的关系

链路监控页面必须与 FindX Agent 控制面联动：

| 链路页面 | FindX Agent 联动 |
| --- | --- |
| 服务目录 | 展示服务/实例探针覆盖率、语言、版本、最后 Trace 时间 |
| 实例详情 | 跳转主机、进程、Agent 心跳、配置版本 |
| Trace 缺失 | 检查采样、探针加载、OAP 连通、配置漂移 |
| 拓扑缺口 | 关联网关探针、Nginx Lua/Kong、服务探针状态 |
| Profiling | 校验对应语言探针或 eBPF 能力包是否可用 |
| 告警复盘 | 引用 Agent 安装/升级/配置下发/回滚证据 |

## 7. 前端迁移规则

当前 FindX 长期方向为 React-first 壳层。SkyWalking Booster UI 是 Vue 3 + Pinia。

迁移规则：

- 不 iframe。
- 不嵌入参考站 SSO。
- 不省组件。
- 可采用同源迁移或等价实现，但状态流、路由语义、GraphQL 查询、错误态、空态、按钮动作必须一致。
- 用户侧样式使用 FindX 主题，结构和功能点按成熟源码。
- 若局部保留 Vue 子模块作为构建期源代码迁移对象，必须仍由 FindX 自有登录、导航、权限、审计和主题接管，不出现外部品牌。

## 8. MCP 与构建验收

编码前必须补运行态证据；编码后必须覆盖：

- 链路监控总览。
- 服务目录。
- 服务详情 Dashboard。
- 服务拓扑。
- Trace 查询。
- Trace 详情 Span 树和 Span 详情。
- 慢调用、错误链路、tag autocomplete。
- Profiling 任务创建、列表、进度、结果。
- 告警列表、告警快照、Trace 跳转。
- 设置页 TTL、集群节点、debug config dump。
- OAP 未配置显示 `BLOCKED`，不得伪装成功。
- WSL 前端 build、后端测试、敏感信息扫描。

## 9. 阻断项

- 只实现后端 OAP proxy，前端仍是空壳：FAIL。
- 只做 Trace 搜索，不做服务目录、拓扑、Profiling、告警、设置：FAIL。
- GraphQL errors 被吞掉或显示成功：FAIL。
- 用户侧出现外部品牌标题、菜单、权限对象、审计对象：FAIL。
- 未配置 OAP 时页面静态展示假数据：FAIL。
- 未接入 FindX Agent 状态与 Evidence Chain：P1 RISK，阻断进入完成态。
