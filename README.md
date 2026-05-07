# FindX

FindX 是面向可观测、Agent 管理、AI SRE 和 Evidence Chain 的长期平台项目。当前正式路线不是 iframe 嵌入参考站，也不是继续维护弱化自研页面，而是：

- 基于成熟平台源码迁移页面结构、路由语义、组件拆分、状态流、API 行为和真实功能点。
- 使用 FindX 自有登录、导航、权限、审计、主题、错误脱敏、数据源配置和品牌风格。
- 最终前端架构为 **React-only**。Vue 仅允许作为临时桥接或待替换对象，不作为最终验收基线。

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
- [.claude/codex-task-board.md](.claude/codex-task-board.md)
- [.claude/operations-log.md](.claude/operations-log.md)

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

## 成熟源码事实源

| 域 | 事实源 | FindX 实施方式 |
| --- | --- | --- |
| 基础监控 | `D:\平台源码\fe-main` | 按源码迁移数据源、系统集成、指标查询、仪表盘、模板中心、告警、通知、组织权限、系统配置 |
| 链路监控 | `D:\平台源码\skywalking-booster-ui-main`、`D:\平台源码\skywalking-master` | 按源码迁移服务目录、拓扑、Trace、Profiling、告警、OAP query、GraphQL/query 状态流 |
| SkyWalking Agent | 上游 Java/Python/Node.js/PHP/Go/Rust/Ruby/Nginx Lua/Kong/Browser Client JS 仓库 | 作为 FindX Agent 能力包纳入包仓库、安装向导、配置模板、心跳、数据到达验证和 Evidence Chain |
| 日志中心 | `D:\平台源码\signoz-develop\frontend` | 按源码迁移日志检索、字段筛选、上下文、聚合、live tail、Pipeline、Saved Views、Trace 关联 |
| CMDB / Agent 在线 | `D:\平台源码\AutoOps-main\AutoOps-main` | 按源码迁移 CMDB 树、主机表、分组、Agent 在线、部署、心跳、统计、终端/监控弹窗 |
| 采集插件 / 巡检诊断 | `D:\平台源码\categraf-main (1)`、`D:\平台源码\catpaw-master` | 作为 FindX Agent 插件、采集配置、巡检诊断、结构化执行和 Evidence Chain 能力来源 |

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

## 验证门禁

每个 UI/API 切片必须记录：

- 源码证据：路径、路由、组件、API、状态流、按钮动作、空态、错误态、权限态、FindX 替换点。
- Windows build：`cd D:\ai-workbench\web && npm run build`。
- WSL build：同步后执行 `cd /opt/ai-workbench/web && npm run build`。
- Lint：存在 `npm run lint` 或等价命令时必须执行；不存在则记录 `NOT_RUN` 和原因。
- 后端变更：`cd /opt/ai-workbench/api && go test -count=1 ./...` 和 `go build -o api-linux .`。
- MCP/Playwright：真实登录、导航点击、主流程、异常路径、权限路径、组件不可用 `BLOCKED`、窄屏回归。
- 敏感扫描、品牌扫描、静态假按钮扫描。
- `.claude/codex-task-board.md` 和 `.claude/operations-log.md` 更新。
