# FindX React-only 前端闭环计划

更新时间：2026-05-08 01:17（UTC+8）

## 1. 结论

FindX 前端最终架构固定为 **React-only**。

React-only 的含义：

- FindX 使用自有登录、导航、权限、审计、主题、错误脱敏和品牌风格。
- 成熟源码提供页面结构、路由语义、组件拆分、状态流、API 行为、按钮动作、空态、错误态、权限态和功能点事实源。
- Vue 只允许作为迁移期临时桥，状态必须标记为 `TEMP_BRIDGE`、`REPLACED` 或 `REMOVE_AFTER_REACT`。
- Vue workbench 不得作为最终页面结构、功能点或验收基线。
- 禁止 iframe、WebView、参考站嵌入、参考站 SSO、参考站侧边栏和运行态会话。
- Browser Use 不可用时使用 Playwright MCP，并记录原因；真实浏览器不可用时只能标记 `BLOCKED` 或 `NOT_RUN`。

## 2. 技术栈基线

| 域 | 成熟源码技术栈 | FindX 策略 |
| --- | --- | --- |
| 基础监控 | React 17、React Router 5、Ant Design 4、ahooks、CodeMirror、AntV、react-grid-layout、reactflow、uplot | 作为 P0/P1 React 兼容基线，按源码迁移页面结构和状态流 |
| 日志中心 | React 18、Webpack、Ant Design 5、React Query/Redux | 迁移日志查询、字段筛选、Saved Views、Pipeline、Trace 关联语义；React 18 专属依赖单独做兼容评估 |
| 链路监控 | Vue 3、Pinia、Vue Router、GraphQL | 不 iframe；按页面结构、store、GraphQL query、Trace/Topology/Profiling 状态流做 React 等价迁移 |
| CMDB / Agent 在线 | AutoOps/AIOps Web 源码 | 迁移 CMDB 树、主机表、分组、Agent 在线、部署、心跳、终端/监控弹窗 |
| Agent 管理中心 | FindX Agent 控制面 + SkyWalking Agent + Categraf + Catpaw | 用户侧统一 FindX Agent，迁移能力包、配置模板、接入向导、状态流和 Evidence Chain |

React 18 升级不能随手改依赖。若需要全局升级，必须单独建立依赖变更任务，覆盖基础监控、日志中心、路由、UI 组件、测试工具、构建产物和浏览器回归兼容矩阵。

## 3. FindX React Shell 边界

React Shell 只负责平台级能力：

- 自有登录页、登录过期跳转和会话恢复。
- 自有主导航、二级导航、面包屑和菜单激活。
- 主题变量、字体、间距、颜色、图标风格和响应式布局。
- 权限、团队、业务组、租户、审计上下文注入。
- Adapter 入口、401/403/503 处理、错误脱敏和 `BLOCKED` 状态。
- 用户侧品牌脱敏：FindX、FindX Agent、链路监控、日志中心、Agent 管理中心、AI SRE。

React Shell 不得：

- 用通用卡片页面替代成熟页面。
- 改写成熟页面内部表格列、筛选、按钮动作、抽屉、弹窗、图表、变量、历史记录、模板导入。
- 用静态列表替代真实接口状态。
- 绕过 FindX 登录、权限、审计和错误脱敏。

## 4. 导航和路由

用户侧主导航固定为：

- 基础设施
- 集成中心
- 数据查询
- 仪表盘
- 告警
- 通知
- 链路监控
- 日志中心
- Agent 管理中心
- AI SRE
- 组织权限
- 平台治理

归属规则：

- 数据源、系统集成、模板中心、采集模板、接入指引归集成中心。
- 指标查询、内置指标、对象快捷视图、记录规则归数据查询。
- 仪表盘列表、详情、变量、Panel、模板导入归仪表盘。
- 告警规则、事件、屏蔽、订阅、自愈、事件流水线归告警。
- 通知规则、通知媒介、消息模板、测试发送归通知。
- SkyWalking 服务目录、拓扑、Trace、Profiling、链路告警、接入归链路监控。
- SigNoZ 日志检索、字段筛选、上下文、Pipeline、Saved Views、Trace 关联归日志中心。
- SkyWalking Agent、多语言探针、网关探针、Browser Agent、Categraf 插件、Catpaw 巡检工具归 Agent 管理中心。

路由 alias 可以保留，例如 `/integrations?section=datasources`、`/query?section=metrics`、`/tracing?section=overview`，但 alias 必须落回同源迁移的页面状态流，不得落到静态 holder。

## 5. 迁移方法

每个页面切片编码前必须输出源码证据清单：

- 来源源码路径。
- 上游路由。
- 核心页面组件。
- 子组件拆分。
- API 调用和请求参数。
- 状态流和缓存策略。
- 按钮真实动作。
- 表格列、筛选区、工具栏、分页、批量操作。
- 抽屉、弹窗、详情页。
- 空态、错误态、权限态、加载态。
- FindX 命名替换点。
- Adapter 契约差距。
- MCP/Playwright 验收入口。

实现优先级：

1. 能同源迁移的页面结构和组件语义，直接迁移到 React-only 工作台。
2. 后端契约完整支持的动作，接真实接口。
3. 后端契约缺失但成熟源码有真实动作的控件，显示 `BLOCKED_BY_CONTRACT` 并记录缺失 API、字段、错误码和数据模型。
4. 组件不可用、数据源未配置、下游不可达时显示 `BLOCKED` 或脱敏 503，不显示假数据。
5. 不能一次完成全域时按任务切片拆分，但每个切片自身不能是假按钮或半语义控件。

## 6. P0 基础监控页面闭环

基础监控按 Nightingale React 源码迁移，优先顺序：

1. 数据源：类型选择、配置示例、URL、认证引用、Header、TLS、超时、测试连接、状态、启停、编辑、删除、权限、审计、错误脱敏。
2. 系统集成和模板中心：图标、分类、说明、接入流程、dashboard、alert、record-rule、预览、导入、冲突处理、失败回滚。
3. 指标查询：数据源选择、PromQL 输入、autocomplete、内置指标、历史记录、新增 query panel、即时/区间查询、Table/Graph、Time、Unit、CSV 导出。
4. 仪表盘：列表、分组、详情、变量、数据源、时间范围、刷新、Panel、添加图表、三点菜单、分享、放大、查询参数、编辑、导入导出。
5. 告警和通知：规则、记录规则、事件、屏蔽、订阅、自愈、事件流水线、通知规则、媒介、模板、发送测试、启停、克隆、删除、授权团队。
6. 组织和系统配置：用户、团队、角色、权限、SSO、站点设置、告警引擎、AI 模型配置、审计日志。

## 7. 链路监控和日志中心

链路监控按 SkyWalking UI/OAP 源码迁移：

- 路由：服务目录、层级、拓扑、Trace 详情、Profiling、告警、接入。
- 页面：dashboard list/new/edit/widget/trace、layer、alarm、profiling。
- 状态：trace store、topology store、dashboard store、alarm store、profile store。
- 查询：GraphQL query、fragments、selector、duration、loading、cancel、error。
- 图形：拓扑交互、Trace waterfall、span 详情、Profiling 任务、告警视图。
- 缺依赖：OAP、BanyanDB、APM Adapter、GraphQL proxy 不可用时显示 `BLOCKED`。

日志中心按 SigNoZ 源码迁移：

- 路由：logs explorer、live logs、pipelines、saved views、trace detail、traces explorer。
- 页面：LogsModulePage、LiveLogs、TracesModulePage、TraceDetail、ExplorerCard、QuickFilters、LogDetail。
- 查询：query builder、字段筛选、聚合、上下文、导出、保存视图、Trace 关联。
- API：logs、pipeline、saveView、trace 调用语义。
- 缺依赖：ClickHouse、Logs Adapter、Query Service、Pipeline API 不可用时显示 `BLOCKED`。

## 8. SkyWalking Agent 一等主线

SkyWalking Agent 不是链路监控附属说明，也不是 P5 后面一句话。它必须贯穿：

- P0：源码矩阵、版本矩阵、许可证/NOTICE、包形态、安装方式、配置项、数据到达验收。
- P2：Agent 管理中心前端入口、能力包目录、接入向导、状态展示、`BLOCKED_BY_CONTRACT`。
- P3：链路监控联动，服务目录/Trace/拓扑/告警可反查探针状态和配置证据。
- P5：完整生命周期，包仓库、离线包、签名、本机/远程/K8s 安装、配置下发、心跳、数据到达、升级、回滚、卸载。

必须纳入 FindX Agent 能力包：

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

未补齐源码、包契约、安装脚本、配置模板、心跳和数据到达验证前，只能显示 `BLOCKED` 或 `BLOCKED_BY_CONTRACT`。

## 9. 多 Agent 任务领取制

默认并行池：

- Worker A：当前最高优先级实现切片。
- Worker B：下一个写集不冲突的前端/Adapter 切片。
- Explorer/QA C：只读源码证据、diff 审查、品牌/敏感/静态假按钮审查。
- Worker D：仅在句柄稳定且写集完全互斥时临时启用。

所有子代理必须：

- 显式 `model: "gpt-5.5"`。
- 绑定 `FX-NIGHT-*` 或 `FX-TASK-*`。
- 记录 agent id、写集、禁止写集、证据源和验收门禁。
- 完成、失败、超时、`BLOCKED` 或不再需要时立即关闭。
- 连续两轮 wait 超时立即关闭，不堆句柄。

## 10. 测试门禁

每个切片必须执行或明确记录 `BLOCKED` / `NOT_RUN`：

- Windows build：`cd D:\ai-workbench\web && npm run build`。
- WSL build：同步后执行 `cd /opt/ai-workbench/web && npm run build`。
- Lint：存在 `npm run lint` 或等价命令时必须执行。
- 后端变更：`cd /opt/ai-workbench/api && go test -count=1 ./...` 和 `go build -o api-linux .`。
- MCP/Playwright：真实登录、导航点击、主流程、异常路径、权限路径、组件不可用 `BLOCKED`、窄屏回归。
- 敏感扫描、品牌扫描、静态假按钮扫描。
- 任务板、执行日志和必要文档更新。
