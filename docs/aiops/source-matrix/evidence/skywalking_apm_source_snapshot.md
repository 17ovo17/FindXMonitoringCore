# SkyWalking 链路监控源码检查证据

生成时间：2026-05-07 14:10（UTC+8）

## 1. 检查范围

本证据记录 `P3-SKYWALKING-APM-REAL` 编码前源码检查结果。当前只确认本地源码结构，未配置可访问 OAP / UI 运行态，因此运行态 DOM 状态为 `BLOCKED_RUNTIME_UNAVAILABLE`。

本地源码：

- UI：`D:\平台源码\skywalking-booster-ui-main`
- OAP query protocol：`D:\平台源码\skywalking-master\oap-server\server-query-plugin`
- WebApp proxy：`D:\平台源码\skywalking-master\apm-webapp`

## 2. UI 路由与页面证据

| 模块 | 源码路径 | 结构作用 |
| --- | --- | --- |
| 路由汇总 | `src\router\index.ts` | 聚合 marketplace、layer、alarm、dashboard、settings、trace 等路由 |
| 路由常量 | `src\router\constants.ts` | 定义 `/marketplace`、`/dashboard/list`、`/dashboard/...`、`/alerting`、`/settings`、`/traces/:traceId` |
| 动态层级路由 | `src\router\layer.ts` | 基于 OAP `getMenuItems` 生成 layer/dashboard 路由 |
| Dashboard 路由 | `src\router\dashboard.ts` | Dashboard 列表、新建、编辑、查看、服务关系、Pod、Process、Widget |
| Trace 路由 | `src\router\trace.ts` | Trace ID 直达页 |
| 告警路由 | `src\router\alarm.ts` | Alerting 页面 |
| 设置路由 | `src\router\settings.ts` | OAP 设置与 TTL 等 |

核心页面：

- `src\views\Layer.vue`
- `src\views\dashboard\List.vue`
- `src\views\dashboard\Edit.vue`
- `src\views\dashboard\Trace.vue`
- `src\views\dashboard\related\topology`
- `src\views\dashboard\related\trace`
- `src\views\dashboard\related\profile`
- `src\views\dashboard\related\async-profiling`
- `src\views\dashboard\related\continuous-profiling`
- `src\views\dashboard\related\ebpf`
- `src\views\dashboard\related\network-profiling`
- `src\views\dashboard\related\log`
- `src\views\Alarm.vue`
- `src\views\Settings.vue`

## 3. 状态流证据

| Store | 源码路径 | 状态语义 |
| --- | --- | --- |
| app | `src\store\modules\app.ts` | 时间范围、UTC、自动刷新、菜单、主题、TTL、版本 |
| selectors | `src\store\modules\selectors.ts` | 服务、实例、端点、进程、关系目标选择器 |
| dashboard | `src\store\modules\dashboard.ts` | dashboard 配置、布局、widget |
| topology | `src\store\modules\topology.ts` | 拓扑图数据和状态 |
| trace | `src\store\modules\trace.ts` | Trace 查询、Span、筛选 |
| log | `src\store\modules\log.ts` | 日志关联与查询 |
| alarm | `src\store\modules\alarm.ts` | 告警列表、筛选、快照 |
| profile / async / pprof / ebpf | `src\store\modules\profile.ts` 等 | Profiling 任务、进度、结果 |

重要状态事实：

- `app.ts` 初始化 OAP 时间、版本、菜单、TTL，并管理全局时间范围与自动刷新。
- `selectors.ts` 通过 GraphQL 查询服务、实例、端点、进程，维护当前选择对象。
- 路由 `layer.ts` 依赖 OAP 返回的菜单动态生成层级导航，FindX 不能把链路监控写成固定静态页面。

## 4. GraphQL 证据

前端 GraphQL 基础：

- `src\graphql\base.ts`
- `src\graphql\index.ts`
- `src\graphql\http\index.ts`
- `src\graphql\fragments`
- `src\graphql\query`

关键语义：

- GraphQL BasePath：`/graphql`
- 请求超时：`2 * 60 * 1000`
- 请求支持 AbortController 取消与超时。
- 非 2xx 响应返回 `errors`，前端状态流必须显示错误态，不能假成功。

关键查询：

| 能力 | 源码路径 | GraphQL 语义 |
| --- | --- | --- |
| 菜单与层级 | `fragments\app.ts`、`query\app.ts` | `getMenuItems`、版本、时间信息、TTL |
| 服务目录 | `fragments\selector.ts` | `listServices(layer)` |
| 实例 | `fragments\selector.ts` | `listInstances(duration, serviceId)` |
| 端点 | `fragments\selector.ts` | `findEndpoint(serviceId, keyword, limit, duration)` |
| 进程 | `fragments\selector.ts` | `listProcesses(instanceId, duration)` |
| 拓扑 | `fragments\topology.ts` | service/pod/process topology |
| Trace 查询 | `fragments\trace.ts` | `queryBasicTraces`、`queryTrace`、`queryTraces`、tag autocomplete |
| 日志关联 | `fragments\log.ts` | `queryLogs`、browser error log、log tags、Trace 关联字段 |
| 告警 | `fragments\alarm.ts` | `getAlarm`、tag keys、tag values、Trace ID |
| Profiling | `fragments\profile.ts`、`async-profile.ts`、`pprof.ts`、`ebpf.ts` | 创建任务、查询任务、进度、结果、栈图 |

## 5. OAP / WebApp 证据

OAP query plugin：

- `D:\平台源码\skywalking-master\oap-server\server-query-plugin\query-graphql-plugin`
- `logql-plugin`
- `promql-plugin`
- `traceql-plugin`
- `status-query-plugin`
- `zipkin-query-plugin`

WebApp proxy：

- `D:\平台源码\skywalking-master\apm-webapp\src\main\java\org\apache\skywalking\oap\server\webapp\OapProxyService.java`
- `D:\平台源码\skywalking-master\apm-webapp\src\main\java\org\apache\skywalking\oap\server\webapp\Configuration.java`
- `D:\平台源码\skywalking-master\apm-webapp\src\main\resources\application.yml`

关键事实：

- `OapProxyService` 对多个 OAP endpoint 做 round-robin 代理。
- proxy 使用 `/healthcheck` 健康检查。
- 支持 GET / POST 转发。
- FindX 后端必须做 APM Adapter / OAP proxy，不把 OAP GraphQL 作为最终业务前端公共契约裸暴露。

## 6. 运行态缺口

当前未配置可访问 SkyWalking UI / OAP 运行态，不能完成 MCP 浏览器 DOM 证据。

状态：

- `SOURCE_PRESENT`
- `RUNTIME_DOM_BLOCKED`

编码前必须补齐：

- 链路监控总览、服务目录、服务实例、端点、拓扑、Trace、Trace 详情、Profiling、告警、设置运行态 DOM。
- OAP GraphQL 正常、超时、错误、未配置、鉴权失败证据。
- 组件不可用 `BLOCKED` 页面。
- FindX 用户侧标题和导航去外部品牌证据。
