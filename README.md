# FindX

FindX 是面向可观测、Agent 管理、AI SRE 和 Evidence Chain 的长期平台项目。当前正式路线不是 iframe 嵌入参考站，也不是继续维护弱化自研页面，而是：

- 基于成熟平台源码迁移页面结构、路由语义、组件拆分、状态流、API 行为和真实功能点。
- 使用 FindX 自有登录、导航、权限、审计、主题、错误脱敏、数据源配置和品牌风格。
- 最终前端架构为 **React-only**。Vue 仅允许作为临时桥接或待替换对象，不作为最终验收基线。

## 功能清单与达成里程碑

更新日期：2026-05-11 10:20（UTC+8）。状态说明：✅ 已初步闭环；🚧 开发中；📋 规划中。凡涉及真实执行器、生产数据源或外部依赖契约缺失的功能，必须继续显示阻断状态，不能用预览命令或任务创建冒充完成。

| 功能模块 | 功能项 | 状态 |
| --- | --- | --- |
| 平台底座 | FindX 自有登录、运行时入口、导航、404 与基础权限校验 | ✅ |
| 平台底座 | React-only Shell 与 Vue 临时桥退场边界 | ✅ |
| 平台底座 | Windows / WSL / 远端 Ubuntu 构建、部署、健康检查基线 | ✅ |
| 平台底座 | React 19+ 依赖升级与兼容性验证 | 📋 规划中 |
| 基础监控 | 数据源管理、系统集成、指标查询、历史记录、表格/图形视图 | 🚧 开发中 |
| 基础监控 | 内置指标、仪表盘、变量、模板中心、导入导出、分享视图 | 🚧 开发中 |
| 基础监控 | 告警规则、事件、屏蔽、订阅、自愈、通知规则、通知媒介、消息模板 | 🚧 开发中 |
| 基础监控 | 组织权限、业务组、系统配置、AI 模型配置 | 🚧 开发中 |
| 链路监控 | 服务目录、实例、端点、拓扑、Trace 检索与 Trace 详情入口 | 🚧 开发中 |
| 链路监控 | Profiling、链路告警、接入向导、Trace 与 Agent 状态反查 | 🚧 开发中 |
| 链路监控 | 生产 OAP / GraphQL / Query Adapter 与 Trace span 详情 | 📋 规划中 |
| 日志中心 | Logs Explorer、字段筛选、上下文、聚合、Saved Views、Trace 关联入口 | 🚧 开发中 |
| 日志中心 | Live Tail、生产 Pipeline 部署/回滚、通用 OTel 数据源 | 📋 规划中 |
| CMDB / 资产 | CMDB 树、模型、实例、主机表、业务组、资源组、Agent 在线入口 | 🚧 开发中 |
| CMDB / 资产 | 主机终端、监控弹窗、部署进度、Agent 心跳和数据到达联动 | 🚧 开发中 |
| Agent 管理中心 | 能力包目录、语言/平台筛选、Linux curl、Windows certutil、PowerShell 命令预览与复制 | 🚧 开发中 |
| Agent 管理中心 | 包仓库、签名校验、安装计划、执行回执、任务状态、日志、审计 | 🚧 开发中 |
| Agent 生命周期 | Linux 本机安装、systemd 注册、卸载、幂等、失败恢复 | 📋 规划中 |
| Agent 生命周期 | Windows 本机安装、Windows Service、卸载、失败恢复 | 📋 规划中 |
| Agent 生命周期 | SSH、WinRM、IIS、Docker、Helm、Operator、DaemonSet、Sidecar、InitContainer | 📋 规划中 |
| 采集插件配置 | Categraf 插件配置远程修改、灰度下发、全量下发、回滚、漂移检测 | 🚧 开发中 |
| 巡检诊断 | Catpaw 巡检诊断、结构化执行、Evidence Chain 证据串联 | 📋 规划中 |
| 数据到达 | heartbeat、metrics、logs、tracing 接收与证据展示 | ✅ |
| 数据到达 | profiling、inspection、RUM、gateway trace 真实接收与验证 | 📋 规划中 |
| AI SRE | 诊断会话、工作流、健康检查、复盘报告、Evidence Chain、知识库入口 | 🚧 开发中 |
| 治理发布 | GitHub SSH 上传、README 里程碑、工作树分桶、禁止敏感文件入库 | ✅ |

下一步：继续推进真实 Agent 执行器、包校验、服务注册、心跳、数据到达、卸载、回滚、审计和 Evidence Chain；`BLOCKED_BY_CONTRACT` 仍是后续闭环输入，不能当作最终完成。

## 当前唯一长期开发计划

当前实施入口：

- [FindX 全栈可观测长期开发计划](docs/aiops/findx_full_stack_observability_long_term_plan.md)
- [FindX React-only 前端闭环计划](docs/aiops/findx_react_only_frontend_long_term_plan.md)
- [AIOps 文档索引](docs/aiops/README.md)
- [源码矩阵索引](docs/aiops/source-matrix/README.md)

旧方向计划、讨论稿和阶段性审计材料只保留为历史证据或归档参考，不再作为实施依据。

## 开发前必须读

- [AGENTS.md](AGENTS.md)
- [FindX 全栈可观测长期开发计划](docs/aiops/findx_full_stack_observability_long_term_plan.md)
- [FindX React-only 前端闭环计划](docs/aiops/findx_react_only_frontend_long_term_plan.md)
- [AIOps 文档索引](docs/aiops/README.md)
- [统一测试基准](docs/testing/AI_WorkBench_统一测试基准.md)
- [第三方来源与融合登记](docs/compliance/third-party-sources.md)
- [.codex/codex-task-board.md](.codex/codex-task-board.md)
- [.codex/operations-log.md](.codex/operations-log.md)

## 硬性执行规则

- 禁止 iframe / WebView / 参考站嵌入。
- 禁止嵌入参考站登录、SSO、侧边栏或运行态会话。
- 禁止把 Vue workbench 补成完成态。
- 禁止自研弱化页面。
- 禁止 MVP、最小实现、最小验证、占位页、静态假按钮、静态假数据和未验证 PASS。
- 用户侧只显示 FindX / FindX Agent / 链路监控 / 日志中心 / Agent 管理中心 / AI SRE 等 FindX 命名。
- 外部产品名称只允许出现在内部源码证据、合规登记、任务板、执行日志和归档文档。
- 成熟源码有真实动作的控件必须接真实动作；后端契约缺失时显示 `BLOCKED_BY_CONTRACT`。
- API 测试不能替代 MCP/Playwright 浏览器真实登录和点击回归。
- Browser Use 插件不可用时，必须使用 Playwright MCP 做真实浏览器回归并记录原因；真实浏览器不可用时只能标记 `BLOCKED` 或 `NOT_RUN`。
- 未跑 WSL build、lint/测试、浏览器回归和扫描，不得写 PASS。
- 所有子代理必须显式使用 `model: "gpt-5.5"`，禁止 fallback 到 5.4。
- FindX Agent 生命周期不能以预览命令或阻断响应冒充完成：本机安装、远程下发、远程安装、卸载、配置下发、插件下发都必须有真实执行成功、服务状态、心跳、数据到达、审计和失败恢复证据。
- Linux `curl -kfsSL`、Windows CMD `certutil -urlcache -f`、PowerShell `Invoke-WebRequest` 安装入口，以及 SSH、WinRM、systemd、Windows Service、IIS、Docker、Helm、Operator、DaemonSet、Sidecar、InitContainer 场景，都必须按平台完成安装、升级、回滚、卸载和数据到达验证。
- SkyWalking Agent、Categraf、Catpaw 与所有 FindX Agent 监控插件必须 Windows/Linux 双实现、双验证；缺少环境时先安装环境再测试监控。

## 成熟源码事实源

| 域 | 事实源 | FindX 实施方式 |
| --- | --- | --- |
| 基础监控 | `D:\项目迁移文件\平台源码\fe-main` | 按源码迁移数据源、系统集成、指标查询、仪表盘、模板中心、告警、通知、组织权限、系统配置 |
| 链路监控 | `D:\项目迁移文件\平台源码\skywalking-booster-ui-main`、`D:\项目迁移文件\平台源码\skywalking-master` | 按源码迁移服务目录、拓扑、Trace、Profiling、告警、OAP query、GraphQL/query 状态流 |
| SkyWalking Agent | 上游 Java/Python/Node.js/PHP/Go/Rust/Ruby/Nginx Lua/Kong/Browser Client JS 仓库 | 作为 FindX Agent 能力包纳入包仓库、安装向导、配置模板、心跳、数据到达验证和 Evidence Chain |
| 日志中心 | `D:\项目迁移文件\平台源码\signoz-develop\frontend` | 按源码迁移日志检索、字段筛选、上下文、聚合、live tail、Pipeline、Saved Views、Trace 关联 |
| CMDB / Agent 在线 | `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main` | 按源码迁移 CMDB 树、主机表、分组、Agent 在线、部署、心跳、统计、终端/监控弹窗 |
| 采集插件 / 巡检诊断 | `D:\项目迁移文件\平台源码\categraf-main (1)`、`D:\项目迁移文件\平台源码\catpaw-master` | 作为 FindX Agent 插件、采集配置、远程下发/修改、巡检诊断、结构化执行和 Evidence Chain 能力来源 |

## 夜间闭环入口

从 `FX-NIGHT-000-DOC-GOAL-LOCK` 开始：

1. 文档和 Goal 指令锁定。
2. 工作树分桶和 Git 边界锁定。
3. React-only Shell。
4. 基础监控同源迁移。
5. CMDB / Agent 在线前端迁移。
6. SkyWalking 链路监控迁移。
7. SkyWalking Agent 能力包入口和矩阵补齐。
8. SigNoZ 日志中心迁移。
9. AI SRE / Evidence Chain 前端入口。
10. P0-P4 前端闭环后，再进入 FindX Agent 生命周期深度闭环。

FindX Agent 生命周期深度闭环的准入条件：先完成包仓库、签名校验、安装计划、执行任务、配置/插件下发、回滚、卸载、心跳、数据到达、审计和 Evidence Chain 契约；再按 Windows/Linux 双系统和必要 Kubernetes 场景做真实安装、远程下发、远程安装、配置下发、插件下发、卸载与失败恢复回归。`BLOCKED_BY_CONTRACT` 只能说明诚实阻断，不能作为安装或下发成功。

## 验证门禁

每个 UI/API 切片必须记录：

- 源码证据：路径、路由、组件、API、状态流、按钮动作、空态、错误态、权限态、FindX 替换点。
- Windows build：`cd D:\ai-workbench\web && npm run build`。
- WSL build：同步后执行 `cd /opt/ai-workbench/web && npm run build`。
- Lint：存在 `npm run lint` 或等价命令时必须执行；不存在则记录 `NOT_RUN` 和原因。
- 后端变更：`cd /opt/ai-workbench/api && go test -count=1 ./...` 和 `go build -o api-linux .`。
- MCP/Playwright：真实登录、导航点击、主流程、异常路径、权限路径、组件不可用 `BLOCKED`、窄屏回归。
- 敏感扫描、品牌扫描、静态假按钮扫描。
- `.codex/codex-task-board.md` 和 `.codex/operations-log.md` 更新。
