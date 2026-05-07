# FindX React-first 前端技术栈长期闭环计划

生成时间：2026-05-08 00:42（UTC+8）
状态：长期主计划的强制前端子计划
上级入口：[FindX 全栈可观测长期开发计划](findx_full_stack_observability_long_term_plan.md)
适用范围：FindX Web Shell、基础监控、链路监控、日志中心、CMDB、Agent 管理中心、AI SRE、Evidence Chain 相关前端实现

## 1. 结论

FindX 前端长期方向固定为 React-first。这里的 React-first 不是把参考站 iframe 进来，也不是把现有 Vue 页面继续补成完成态，而是用用户提供的成熟源码作为事实源，把成熟产品的页面结构、路由语义、组件拆分、状态流、API 调用、交互动作、错误态、空态和权限态迁移到 FindX 自有 React 壳层中。

FindX 自己负责登录、导航、主题、权限、审计、错误脱敏、品牌文案、统一配置、数据源中心、AI SRE、Evidence Chain 和 FindX Agent 控制面。成熟产品源码负责告诉我们页面应该怎么组织、控件应该做什么、状态应该怎么流转、接口应该如何被调用、功能点不能缺哪些。

本计划明确禁止以下方向：

- 禁止 iframe 或 WebView 嵌入参考站点。
- 禁止嵌入参考站登录、SSO、侧边栏或运行态会话。
- 禁止继续把 Vue workbench 补成最终完成态。
- 禁止只照截图做相似界面。
- 禁止静态假按钮、静态假数据、静态假列表和假成功提示。
- 禁止最小实现、最小验证、MVP、占位页和未验证 PASS。
- 禁止用户侧出现外部产品品牌、菜单名、页面标题、权限对象或审计对象。

## 2. 技术栈基线

### 2.1 当前 FindX Web 基线

当前 FindX Web 仍有 Vue 3.5、Element Plus、Vue Router、Vuex 等历史依赖，同时已引入 React 17、React DOM 17、React Router DOM 5 和 Vite React 插件。这个状态只能视为迁移期双栈，不是最终架构完成态。

迁移期规则：

- Vue 页面只作为兼容桥、过渡承载或待迁移对象。
- Vue 页面不得作为成熟页面结构和功能点的最终验收基线。
- 新的基础监控工作台必须优先落到 React 同系实现。
- Vue 与 React 的桥接只能负责挂载、路由转接和兼容迁移，不得在桥接层重写成熟业务逻辑。

### 2.2 基础监控 React 基线

基础监控事实源是 `D:\平台源码\fe-main`。该源码是 React 17、React Router 5、Ant Design 4、ahooks、CodeMirror、AntV、react-grid-layout、reactflow、uplot 等组合。

FindX 基础监控长期实现应优先与该技术栈兼容：

- React：以 React 17 兼容线作为 P0/P1 基线。
- 路由：保留 React Router 5 的路由语义和菜单激活方式，再由 FindX Shell 做统一入口映射。
- 组件：迁移成熟源码的页面组件结构、表格结构、抽屉弹窗、三点菜单、表单、查询输入、变量选择、模板导入等组件职责。
- 状态：迁移成熟源码的请求状态、表单状态、查询面板状态、历史记录、autocomplete、dashboard session、变量状态和批量操作状态。
- UI：使用 FindX 视觉皮肤，但不改变成熟页面的信息密度、操作路径、控件语义和布局重心。

### 2.3 多源前端兼容策略

不同成熟源码技术栈不同，FindX 不能因此退回 iframe 或自研弱化页面。

| 来源域 | 上游前端事实 | FindX 策略 |
| --- | --- | --- |
| 基础监控 | React 17 + React Router 5 + Ant Design 4 | 作为 P0/P1 React-first 主要兼容基线，页面结构和状态流同源迁移 |
| 日志中心 | React 18 + Webpack + Ant Design 5 + React Query/Redux | 迁移路由、查询构造、日志筛选、Saved Views、Pipeline、Trace 关联语义；React 18 专有依赖进入兼容评估，不直接污染基础监控基线 |
| 链路监控 | Vue 3 + Pinia + Vue Router + GraphQL | 不 iframe；按页面结构、store、GraphQL query、拓扑/Trace/Profiling 状态流做 React 等价迁移，或在受控切片中建设等价 React store |
| CMDB / Agent 在线 | AutoOps/AIOps Web 源码 | 迁移 CMDB 树、主机表、分组、Agent 在线、心跳、部署和终端弹窗语义 |
| 采集插件 / 巡检 | Categraf / Catpaw 源码 | 用户侧统一进入 FindX Agent；迁移插件目录、配置模板、下发、回滚、巡检诊断和证据产物语义 |

React 18 升级不是随手改依赖。若后续需要把 FindX 全局从 React 17 升到 React 18，必须单独建立 `API_CONTRACT_CHANGE` / 依赖变更任务，完成基础监控、日志中心、React Router、Ant Design、测试工具、构建产物和浏览器回归兼容矩阵后再执行。

## 3. FindX React Shell 边界

React Shell 是 FindX 的壳，不是成熟页面的替代品。它只负责平台级能力：

- 自有登录页、登录过期跳转和会话恢复。
- 自有主导航、二级导航、面包屑和菜单激活。
- 统一主题变量、字体、间距、颜色、图标风格和暗色/亮色策略。
- 权限、团队、业务组、租户、审计上下文注入。
- API Adapter 入口、错误脱敏、401/403/503 处理和组件不可用 `BLOCKED`。
- 用户侧品牌脱敏：FindX、FindX Agent、链路监控、日志中心、Agent 管理中心、AI SRE。
- 全局布局稳定性：响应式、窄屏、滚动容器、表格高度、抽屉层级、弹窗层级。

React Shell 不得做这些事：

- 不得把成熟页面内部表格列、筛选、按钮动作、抽屉、弹窗、图表、变量、查询历史、模板导入改成自研简化版。
- 不得用一个通用卡片页替代多个成熟页面。
- 不得用静态列表替代真实接口状态。
- 不得把成熟产品的导航整段嵌入 FindX。
- 不得绕过 FindX 登录、权限、审计和错误脱敏。

## 4. 导航和路由闭环

导航归 FindX Shell 管，页面结构和功能点归成熟源码事实源。

FindX 用户侧主导航长期稳定为：

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

基础监控的归属必须按成熟源码和运维场景放置：

- 数据源、系统集成、模板中心、采集模板、接入指引归集成中心。
- 指标查询、内置指标、对象快捷视图、记录规则归数据查询。
- 仪表盘列表、仪表盘详情、变量、Panel、模板导入归仪表盘。
- 告警规则、事件、屏蔽、订阅、自愈、事件流水线归告警。
- 通知规则、通知媒介、消息模板、通知测试归通知。
- 用户、团队、角色、SSO、站点设置、告警引擎、AI 模型配置、审计日志归组织权限或平台治理。

路由实现规则：

- FindX 可以提供更符合自身信息架构的 alias，例如 `/integrations?section=datasources`、`/query?section=metrics`、`/tracing?section=overview`。
- alias 必须落回同源迁移的页面状态流，不能落到静态 holder。
- 旧路由兼容必须明确 redirect 或 `BLOCKED_BY_CONTRACT`，不得随机拼路径。
- 菜单激活、面包屑、返回路径必须按成熟源码语义校对，不得出现模板导入弹到随机路径、仪表盘归属错位、记录规则跑到普通告警规则等问题。

## 5. 页面迁移方法

每个页面切片必须先完成源码证据清单，再编码。清单至少包含：

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
- FindX Adapter 契约差距。
- MCP 浏览器验收入口。

实现时按以下优先级处理：

1. 能同源迁移的页面结构和组件语义，直接迁移到 React-first 工作台。
2. 后端契约已完整支持的动作，接真实接口。
3. 后端契约未支持但成熟源码有真实动作的控件，显示 `BLOCKED_BY_CONTRACT`，并列出缺失 API、字段、错误码和数据模型。
4. 组件不可用、数据源未配置、下游不可达时，显示 `BLOCKED` 或脱敏 503，不显示假数据。
5. 不能一次完成全域时，按任务切片拆，但每个切片自身不能是假按钮或半语义控件。

## 6. 基础监控 P0 页面闭环

P0-BASE-MONITORING-REAL 是前端重构的第一批完成标准，不是一个泛泛页面集。

### 6.1 集成中心

必须按基础监控源码迁移：

- 数据源：类型选择、配置示例、URL、认证引用、Header、TLS、超时、测试连接、状态、启停、编辑、删除、权限、审计、错误脱敏。
- 系统集成：集成列表、图标、说明、接入流程、过滤、分类和详情。
- 模板中心：图标、分类、说明、适用组件、dashboard、alert、record-rule、预览、导入、冲突处理、失败回滚。
- 采集模板：以 FindX Agent 插件方式承载，不删模板含义；用户侧文案统一 FindX Agent。

数据源当前若只有列表读取，没有成熟数据源 upsert、delete、status、plugin、TLS/Header 完整契约，则新增、编辑、删除、测试连接必须按真实能力分级：能真实做的接真实接口，不能真实做的标 `BLOCKED_BY_CONTRACT`。

### 6.2 数据查询

必须按基础监控源码迁移：

- 指标查询：数据源选择、PromQL 输入、autocomplete、内置指标、历史记录、新增 query panel、即时查询、区间查询、Table/Graph、Time、Unit、CSV 导出。
- 内置指标：作为输入辅助和选择能力，不是常驻超大搜索框。
- 历史记录：轻量选择和回填能力，不做静态展示。
- 添加面板：新增查询 panel，不是收藏。
- 对象快捷视图、记录规则：按成熟源码的列表、编辑、启停、克隆、删除、权限态接入。

### 6.3 仪表盘

必须按基础监控源码迁移：

- 仪表盘列表、分组、搜索、导入导出、模板导入。
- 详情页 header、数据源、变量、时间范围、刷新、添加图表、Panel、三点菜单、分享、放大、查询参数、编辑、复制、删除。
- 模板变量必须可选择、可搜索、可编辑，不得做静态标签。
- 相同类型监控模板按成熟产品聚合，不得全部拆散。

### 6.4 告警和通知

必须按基础监控源码迁移：

- 告警规则、记录规则、事件、屏蔽、订阅、自愈、事件流水线。
- 通知规则、通知媒介、消息模板、发送测试、启停、克隆、删除、授权团队、更新人、更新时间。
- 工作流入口并入 AI SRE 工作流，但内部保留成熟 event-pipeline / workflow 的列表、执行记录、编辑、调试和权限结构。

## 7. 链路监控和日志中心

### 7.1 链路监控

链路监控不能只做 OAP Adapter，也不能把链路页面做成静态拓扑。必须按 `D:\平台源码\skywalking-booster-ui-main` 和 `D:\平台源码\skywalking-master` 做同源状态流迁移：

- 路由：服务目录、层级、拓扑、Trace 详情、Profiling、告警、接入。
- 页面：dashboard list/new/edit/widget/trace、layer、alarm、profiling 等页面结构。
- 状态：trace store、topology store、dashboard store、alarm store、profile store 等状态语义。
- 查询：GraphQL query、fragments、selector、duration、loading、cancel、error。
- 图形：拓扑交互、Trace waterfall、span 详情、profiling 任务、告警视图。
- 缺依赖：OAP、BanyanDB、APM Adapter、GraphQL proxy 不可用时显示 `BLOCKED`，不造假拓扑或假 Trace。

用户侧只显示“链路监控”。外部来源名只允许在内部证据、源码矩阵、合规登记和归档文档中出现。

### 7.2 日志中心

日志中心不能只接一个日志查询接口。必须按 `D:\平台源码\signoz-develop\frontend` 做同源状态流迁移：

- 路由：logs explorer、live logs、pipelines、saved views、trace detail、traces explorer。
- 页面：LogsModulePage、LiveLogs、TracesModulePage、TraceDetail、ExplorerCard、QuickFilters、LogDetail。
- 查询：query builder、字段筛选、聚合、上下文、导出、保存视图、Trace 关联。
- API：logs、pipeline、saveView、trace 等调用语义。
- 缺依赖：ClickHouse、Logs Adapter、Query Service、Pipeline API 不可用时显示 `BLOCKED`，不造假日志列表。

用户侧只显示“日志中心”。

## 8. Agent 管理中心前端闭环

Agent 管理中心必须是平台级产品，不是脚本清单。

必须覆盖：

- 包仓库：FindX Agent 能力包、版本、签名、校验、离线包、兼容矩阵。
- 安装向导：Linux、Windows、Kubernetes、本机脚本、远程安装、安装包安装。
- 本机安装：Linux `curl -kfsSL`、Windows CMD `certutil -urlcache -f`、Windows PowerShell `Invoke-WebRequest` 下载安装脚本并传入短期 `<TOKEN>`。
- 远程安装：SSH、WinRM、PowerShell、systemd、Windows Service、Helm、Operator、DaemonSet、Sidecar、InitContainer。
- 配置下发：服务名、实例名、OAP endpoint、采样、插件 include/exclude、日志关联、TLS/proxy、标签、资源限制。
- 心跳：控制面心跳、进程状态、探针加载状态、OAP 连通性、Trace 到达、指标到达、日志关联到达。
- 生命周期：升级、灰度、漂移检测、回滚、卸载、残留清理。
- Evidence Chain：远程命令、包校验、安装脚本、配置变更、重启、回滚、卸载全部留证。

SkyWalking 多语言 Agent、网关 Agent、Browser Client JS、采集插件和巡检工具在用户侧都作为 FindX Agent 能力包出现。独立源码或制品未落地时只能标 `BLOCKED`，不能用静态清单冒充。

## 9. Adapter 与数据源一致性

React 页面不能直接把成熟产品原 API 暴露成 FindX 公共契约。所有域必须通过 Adapter 收敛：

- Base Monitoring Adapter：数据源、指标查询、仪表盘、模板、告警、通知、工作流。
- APM Adapter：OAP/query protocol/GraphQL proxy、服务目录、拓扑、Trace、Profiling、告警。
- Logs Adapter：日志检索、字段筛选、上下文、聚合、Pipeline、Saved Views、Trace 关联。
- CMDB Adapter：CMDB 树、主机、业务组、Agent 在线、部署、心跳、统计。
- Agent Adapter：安装计划、包仓库、配置模板、任务、心跳、升级、回滚、卸载。

统一数据源中心必须覆盖 Prometheus/VictoriaMetrics、MySQL/MariaDB、Redis、MinIO、Qdrant、OAP、BanyanDB、ClickHouse、OTel Collector、可选 ES/OpenSearch。凭据只引用不回显，错误必须脱敏。

如果页面功能需要的字段、状态、错误码或数据模型缺失，前端必须标 `BLOCKED_BY_CONTRACT`，并记录 API_CONTRACT_CHANGE / DATA_CHANGE 候选，不得写假动作。

## 10. 测试和准出门禁

每个 React-first 切片至少必须完成以下验证，无法完成时只能标 `BLOCKED`、`NOT_RUN` 或 `RISK`：

- 源码证据门禁：源码路径、路由、组件、API、状态流、按钮动作、空态、错误态、权限态、FindX 替换点。
- Windows 构建：`cd D:\ai-workbench\web && npm run build`。
- WSL 构建：同步到 `/opt/ai-workbench` 后执行 `cd /opt/ai-workbench/web && npm run build`。
- MCP/Playwright 浏览器：真实登录、导航点击、主流程、异常路径、权限路径、组件不可用 `BLOCKED`、窄屏回归。
- API 验证：正常路径、非法输入、401、403、404、409、503、超时、取消请求、错误脱敏。
- 敏感扫描：token、cookie、Bearer、DSN、SSH key、private key、会话 ID。
- 品牌扫描：用户侧菜单、页面标题、权限对象、审计对象不得出现外部品牌。
- 静态假按钮扫描：成熟源码有真实动作的控件，在 FindX 中不得无动作或语义错误。

API 测试不能替代浏览器真实点击。Windows build 不能替代 WSL build。开发代理自测不能替代 QA 或主代理回归。

## 11. 多 Agent 执行制度

多 Agent 用于提速，但必须有边界。

规则：

- 所有子代理必须显式 `model: "gpt-5.5"`。
- 禁止 fallback 到 5.4。
- 每个子代理必须绑定任务 ID、agent id、写集、禁止写集、证据源、验收门禁。
- 同一批 agent 写集必须互斥。
- 默认最多保留 1 个执行型 worker + 1 个只读 QA/explorer；确需更多并行时必须先写任务池和关闭时机。
- 子代理完成、失败、超时、BLOCKED 或不再需要时必须立即关闭。
- 连续两轮等待超时、句柄异常、not_found、容量不可用时必须关闭该 agent，不得反复堆线程。
- 主代理等待期间必须做不冲突的只读审计、构建准备、敏感扫描、浏览器验证准备或文档门禁。

任务板必须记录每个 agent：

- `FX-TASK-*`。
- 状态：READY、CLAIMED、IN_PROGRESS、DONE、FAIL、BLOCKED、NOT_RUN、RISK。
- agent nickname/id。
- 写集和禁止写集。
- 来源证据。
- 验证结果。
- 关闭状态和残留风险。

## 12. 文档维护闭环

React-first 是长期方向，不能只存在于对话里。

每次发生以下变更，都必须同步文档：

- React/Vue 入口、依赖、Vite、路由、桥接方式变化。
- 页面迁移范围变化。
- Adapter 契约变化。
- 数据源配置变化。
- 测试门禁变化。
- 多 Agent 策略或反卡死策略变化。
- 用户侧品牌脱敏规则变化。
- `BLOCKED` / `BLOCKED_BY_CONTRACT` 定义变化。

必须同步的文档：

- `AGENTS.md`
- `README.md`
- `docs/aiops/README.md`
- `docs/aiops/findx_full_stack_observability_long_term_plan.md`
- `docs/aiops/findx_react_first_frontend_long_term_plan.md`
- `.claude/codex-task-board.md`
- `.claude/operations-log.md`

文档未覆盖本轮改动时，不得标 PASS，不得进入 Git 门禁，不得继续堆新功能。

## 13. 阶段路线

### P0-REACT-FOUNDATION

完成 React 依赖、Vite React 插件、React Shell 入口、Vue 兼容桥、路由 alias、主题变量、权限上下文、全局错误态和品牌替换规则。当前已进入迁移期，但不能因此标记完成；必须补齐 WSL build、MCP 回归和文档门禁。

### P0-BASE-MONITORING-REAL

按基础监控 React 源码真实逻辑迁移集成中心、数据查询、仪表盘、模板中心、告警、通知、系统配置。优先顺序为数据源、指标查询、仪表盘、模板中心、告警、通知、系统配置。

### P1-ORG-GOVERNANCE

迁移用户、团队、角色、权限、SSO、站点设置、告警引擎、AI 模型配置、审计日志。FindX 自有 IAM 和 AI 配置是准出边界。

### P2-CMDB-AGENT

按 AutoOps/AIOps 源码迁移 CMDB 树、主机表、分组、Agent 在线、部署、心跳、统计和终端/监控弹窗；与 FindX Agent 控制面合并命名和权限。

### P3-TRACING

按链路监控源码迁移服务目录、拓扑、Trace、Profiling、告警和接入。OAP / BanyanDB / Adapter 不可用时明确 `BLOCKED`。

### P4-LOGS

按日志中心源码迁移 logs explorer、field filters、context、aggregation、Pipeline、Saved Views、Trace 关联和 live tail。ClickHouse / Logs Adapter 不可用时明确 `BLOCKED`。

### P5-FINDX-AGENT-SUITE

完成 FindX Agent 包仓库、安装向导、本机脚本、远程安装、配置下发、心跳、数据到达、升级、回滚、卸载、审计和 Evidence Chain。

### P6-AISRE-EVIDENCE-CHAIN

完成 AI SRE 诊断会话、工作流、健康检查、复盘报告、Evidence Chain、模型配置统一读取和证据缺失返回。

### P7-KNOWLEDGE-VECTOR

完成 MySQL 权威知识库、Qdrant 内置向量索引、BM25 降级、RRF 融合、索引任务、索引状态 UI 和 Evidence Chain 引用。

## 14. 当前闭环状态

当前文档闭环状态：

- React-first 长期方向已落入主计划、README、AIOps 索引和 AGENTS 硬规则。
- 本文档作为 React-first 前端专题计划，成为开发前必须读文件。
- 当前代码工作树仍有多项未完成迁移和 RISK，不能把现有 Vue holder 或临时 React 页面视为完成态。
- 下一步必须继续处理数据源契约风险：若不能证明 per-item upsert 和 full round-trip 安全，则新增/编辑/删除必须标 `BLOCKED_BY_CONTRACT`，或先补后端契约。
