# FindX Monitoring Core 与 findx-agents 正式立项计划

生成时间：2026-05-04 04:30（UTC+8）

最新更新：2026-05-04 12:40（UTC+8）

## 1. 立项结论

AI WorkBench / FindX 的项目定位明确为：**FindX Monitoring Core 是新一代监控核心平台**。FindX 是主平台、主视觉、主入口、主数据模型和最终运行主体；Nightingale 是被深度嵌入、完整对标和可融合改造的成熟监控能力来源；Categraf 是 `findx-agents` 采集生态的主要来源；Catpaw 是已授权巡检、诊断、会话、结构化工具、MCP bridge 与远程安全设计的衍生来源。

本轮用户确认的硬边界：

- **FindX 嵌入 Nightingale 做深度集成，不是把 FindX 搞成 Nightingale 外壳。**
- **Nightingale 有的基础监控能力，一个都不能少。**
- **FindX 的 AI 问诊、知识库、Agent 深度巡检、单机对话、自动修复是锦上添花的增强层，不能替代基础监控控制面。**
- **Agent 也必须保留基础采集、机器存活、业务组分配、配置分发和生命周期管理；FindX 增强的是 CMDB、远程/本机安装、巡检证据、单机对话、审批和修复闭环。**
- **Linux/WSL 是运行验收基准，Windows 只作为源码与开发侧辅助入口。**

当前阶段结论：

- **P0 已完成**：FindX 后端 API 基座已完成并推送，提交为 `81b4531`；QA 评分 **98/100**，结论为通过；Windows 与 WSL Go 测试/构建均已通过。
- **P0 文档规划已补齐**：FindX Monitoring Core 的新平台定位、Nightingale 参考/改进边界、Categraf 插件复用、Catpaw 授权衍生、AI 问诊与自动修复核心增强、WSL/Linux 兼容基准已进入文档闭环。
- **P1-BE-3 已完成**：alert scheduler 已形成 current/history event 自动闭环，覆盖定时评估、current upsert、history recovery、失败只写 eval log、并发 RunOnce 锁保护和安全摘要；提交为 `1c4045e feat: add FindX alert scheduler event closure`，QA **92/100 PASS**，Windows/WSL `go test -count=1 ./...` 与 build 均通过。
- **安全基线已补强**：CORS 空配置默认保持 localhost-only；Workflow DSL 创建/更新在 parse 后执行严格 graph validation，坏 DSL 返回 400 且不得持久化。
- **P1/P2/P3 继续按全功能路线推进**：P1 告警后端闭环部分已完成，Dashboard、模板、通知、权限、团队、值班、订阅、静默仍待实现；P2 `findx-agents` 深度融合、P3 AI 问诊与自动修复闭环仍待正式落地。
- **findx-agents 路线固定**：按“统一 Agent 分发版 + 统一控制协议 + 结构化诊断工具 + 受控修复执行底座”推进。
- **自动修复纳入核心闭环**：自动修复必须具备 `precheck -> dry-run -> approve -> execute -> verify -> rollback -> audit` 全链路，所有执行动作必须由审批后的计划驱动。

总体定位：

- FindX Monitoring Core：承载 target、datasource、query、alert rule、evaluator、event、notification、dashboard、template、pipeline、task、permission、audit、Agent 控制面、AI 问诊和自动修复闭环。
- Nightingale：成熟功能参考和可融合源码来源，重点参考告警、事件、Dashboard、通知、模板、权限、事件流水线、任务和前端编辑体验。
- Categraf：保留并改造 inputs、provider、remote_write、heartbeat 和配置模板生态，作为 `findx-agents` 采集核心。
- Catpaw：授权衍生 inspector、diagnose、session、tool registry、MCP bridge、event model 与 remote security 能力。
- findx-agents：FindX 自有 Agent，形态为 Categraf 采集核心 + Catpaw 衍生巡检诊断能力 + FindX Agent Control Protocol v1 + 受控 remediation executor。
- AI 问诊：基于 FindX 自有监控事件、指标、日志、Agent 巡检证据、通知记录、修复记录和知识库案例进行推理。

核心原则：

- **新平台优先**：不迁移历史监控数据，不迁移历史告警事件，不以既有生产平台无缝切换为项目约束。
- **参考成熟实现**：Nightingale 的成熟设计可参考、融合、改造，但最终 API、数据模型、UI、权限、AI、Agent、自动修复均归 FindX 主线。
- **一个都不能少**：用户/团队/角色/权限、业务组、target、机器存活、数据源、查询、Dashboard、Chart、分享、注解、告警规则、记录规则、current/history event、聚合视图、静默、订阅、通知全链路、值班/升级、事件流水线、任务/模板、integrations、SourceToken/API Token、SavedView、嵌入产品、日志/Trace、AI assistant/agent/MCP/Skill 都纳入覆盖范围。
- **全功能实现**：不允许 MVP、占位页、半截接口、静态假数据、只读目录或一次性 PoC 作为交付结论。
- **Agent 深度融合**：Categraf 插件直接复用，Catpaw 授权能力衍生进 `findx-agents`，形成可安装、可升级、可审计、可回滚的统一发行版。
- **安全前置**：Agent 注册、工具调用、AI prompt、审计日志、远程执行、自动修复均必须先定义权限、审批、脱敏、重放保护和审计策略。

## 2. 源码审计范围与关键证据

本次审计以用户提供的 `D:\平台源码`、当前 git 工作区 `D:\ai-workbench` 和 WSL 运行项目 `/opt/ai-workbench` 为准。

| 组件 | 本地路径 | 审计结论 |
| --- | --- | --- |
| Nightingale | `D:\平台源码\nightingale-main (1)\nightingale-main` | 本地 LICENSE 为 Apache-2.0。源码具备 targets、busi groups、datasources、alert rules/events、dashboards、notify、message templates、event pipelines、embedded products、builtin components/payloads、AI assistant、integrations 模板。适合作为 FindX Monitoring Core 的核心参考实现和可融合源码来源。 |
| Categraf | `D:\平台源码\categraf-main (1)\categraf-main` | 本地 LICENSE 为 MIT。源码具备 inputs、local/http provider、remote_write、heartbeat、配置模板和大量 input 插件。适合作为 `findx-agents` 采集核心，但 `inputs/exec` 必须默认禁用，且不得进入 AI 可调用工具目录。 |
| Catpaw | `D:\平台源码\catpaw-master\catpaw-master` | 本地 LICENSE 为 AGPL-3.0，用户已确认公司测试项目具备授权边界。源码具备 inspector、diagnose、session、tool registry、MCP bridge、远程安全设计和结构化输出基础。适合作为 `findx-agents` 巡检诊断与受控工具能力衍生来源。 |
| AutoOps | `D:\平台源码\AutoOps-main\AutoOps-main` | 有 Agent 部署、CMDB、任务、K8s、工具市场思路；旧 Agent 部署存在硬编码 token、临时编译和重复自研问题，不作为新探针底座。 |
| AI WorkBench | `D:\ai-workbench` / `/opt/ai-workbench` | 当前已有 AI 问诊、Prometheus、Catpaw、N9E Redis/MySQL 读取、拓扑、工作流、知识库。P0 已形成 FindX Monitoring Core 后端 API 基座；后续新主线继续建设 `/api/v1/monitor/*`、`/api/v1/findx-agents/*`、`/api/v1/remediation/*`。 |

Nightingale 关键证据：

- `center/router/router.go` 暴露大量路由，覆盖身份、用户、团队、业务组、目标、数据源、指标、Dashboard、告警、通知、事件流水线、日志、任务、保存视图、Webhook、AI assistant。
- `integrations/` 下包含集成目录，覆盖 dashboards、alerts、metrics、record-rules、markdown、icon、collect toml。
- `center/router/router_builtin.go` 和 `center/integration/init.go` 从 integrations 初始化 builtin components、alerts、dashboards、metrics。
- `doc/api/event-pipeline.md` 明确 event pipeline 的列表、详情、创建、更新、删除、tryrun、trigger、stream、executions 和 service API。
- Dashboard、告警规则、通知模板、事件流水线和模板中心适合做 FindX UI、模型和状态机对标。

Categraf 关键证据：

- `inputs/inputs.go`：集中注册、加载和调度 inputs，是 `findx-agents` 采集插件目录和能力目录的核心锚点。
- `inputs/provider_manager.go`：支持 `local` 与 `http` provider，是三层配置模型中本地层与控制面下发层的实现参考。
- `inputs/http_provider.go`：支持远端配置拉取、version/hash、labels、hostname 查询和动态 reload，可衍生为 FindX config-pull。
- `agent/heartbeat.go`：提供 Agent 心跳上报模式，是 FindX heartbeat、capabilities、offline policy 的改造基础。
- `agent/agent.go`：将 metrics、logs、prometheus、ibex 拆成独立 `AgentModule`，适合继续增加 inspector、diagnose、session、remediation。
- `writer/remote_write` 相关实现：可复用为指标上报通道，后续需接入 FindX target、tenant、business group、trace_id 与脱敏策略。
- `inputs/exec`：只能作为受控采集插件候选，生产默认禁用，不进入 AI 工具目录，不作为自动修复 raw command 通道。

Catpaw 关键证据：

- `agent/inspect.go`：支持 `RunInspect(pluginName, target)` 单机巡检模式，可衍生为 `inspection_template` 执行入口。
- `digcore/diagnose/types.go`：定义 `DiagnoseTool`、`ToolScopeLocal/Remote`、`CheckSnapshot`、`DiagnoseRequest`、`DiagnoseRecord`，是结构化诊断工具和 evidence schema 的重要依据。
- `digcore/server/proto.go`：定义 Agent 与 Server 的 `register`、`heartbeat`、`alert_events`、`session_start`、`session_output`、`session_error` 协议，可作为 FindX Agent Control Protocol v1 的参考。
- `registry.go`：工具注册与能力目录思想可衍生为 FindX `ToolDefinition` 与 capability catalog，但必须补齐权限、审批、脱敏和审计字段。
- `mcp/manager.go`：MCP bridge 管理能力可衍生为结构化工具桥接层，但只允许白名单工具被 AI 调用。
- `docs/event-model.md`：可作为 FindX Agent 事件模型、session event 与 diagnosis event 的参考。
- `design.d/remote-security.md` 与 remote security 相关实现：可作为远程会话、重放保护、审批、执行隔离、审计留痕的安全基线。
- 现有 remote/exec raw command 不能进入 AI 自动修复主链路；后续必须改为审批后的固定 plan step 与受控 executor。

AI WorkBench 现状证据：

- `api/main.go` 当前已有 `/api/v1/catpaw/*`、`/api/v1/remote/install-catpaw`、`/api/v1/remote/uninstall-catpaw`、`/api/v1/prometheus/*`、`/api/v1/n9e/agents`、`/api/v1/n9e/alerts`。
- `api/internal/handler/n9e_agents.go` 当前直接读取 Redis `n9e_meta_*`。
- `api/internal/handler/n9e_alerts.go` 当前直接查询 MySQL `alert_cur_event`。
- `web/src/views/CatpawInstall.vue`、`web/src/views/CatpawChatPanel.vue`、`web/src/views/Diagnose.vue`、`web/src/views/TopologyHub.vue` 仍存在 Catpaw 命名和入口，后续需要按 FindX 主线做兼容期治理。
- P0 后端 API 基座已提交 `81b4531`，QA 98/100 通过，Windows + WSL Go 测试/构建通过；P0 文档规划、P1-BE-1 `82ea50d`、P1-BE-2 `1b37905`、P1-BE-3 `1c4045e feat: add FindX alert scheduler event closure`、CORS localhost-only 默认基线、Workflow DSL 严格校验已纳入当前已实现切片记录。P1-BE-3 QA **92/100 PASS**，Windows/WSL `go test -count=1 ./...` 与 build 均通过；该状态不代表 P1 剩余 Dashboard/模板/通知/权限/团队/值班/订阅/静默、P2/P3 已实现。

## 3. 许可证、授权与命名策略

许可证事实：

- Nightingale 本地源码 LICENSE 为 Apache License 2.0。
- Categraf 本地源码 LICENSE 为 MIT。
- Catpaw 本地源码 LICENSE 为 GNU AGPL-3.0，用户已确认本公司测试项目具备授权边界，可规划为授权衍生。

合规要求：

- 引用、融合或改造 Nightingale/Categraf/Catpaw 代码时，必须保留对应 LICENSE、NOTICE、来源说明、修改说明和内部授权记录。
- Catpaw AGPL-3.0 授权衍生进入 `findx-agents` 时，必须补齐仓库级 LICENSE/NOTICE、来源版本、修改说明、授权记录和分发边界说明。
- 商业化、外部分发或客户现场交付前，必须由法务或授权负责人复核 Catpaw 衍生范围。
- 文档、日志、测试报告不得写真实密钥、认证票据、Cookie、完整 DSN、SSH 私钥；示例统一使用 `<TOKEN>`、`<API_KEY>`、`<DB_DSN>`、`<LOGIN_USER>`。

Catpaw 衍生合规章节模板：

| 模板项 | 必填内容 | 进入仓库前门禁 |
| --- | --- | --- |
| 来源版本 | 记录 Catpaw 上游仓库路径、commit/tag、LICENSE 文件摘要、取得授权的内部记录编号；不得写入真实账号、token 或合同敏感正文。 | 合规负责人确认来源版本可追溯，源码引用处保留 SPDX 或文件头说明。 |
| 修改说明 | 按文件或模块记录衍生范围，例如 inspector、diagnose、session、tool registry、MCP bridge、remote security；说明删除、重命名、重构和 FindX 安全增强点。 | 每次引入或改造必须有 diff 摘要、风险说明、回滚方式和测试证据。 |
| NOTICE | 仓库级 NOTICE 必须说明 Catpaw AGPL-3.0 来源、授权边界、FindX 修改范围、第三方依赖链；用户文档不得把 Catpaw 衍生误写成完全自研。 | NOTICE 与 LICENSE 同步进入评审清单，缺失时不得合并 `findx-agents` 发行物。 |
| 授权边界 | 明确本公司测试项目、内部部署、客户现场交付、SaaS/托管服务、二进制分发、源码分发的允许/禁止边界。 | 任何超出“公司测试项目授权边界”的用途必须先由法务或授权负责人书面复核。 |
| 商业分发复核 | 记录分发对象、交付形态、是否包含 AGPL 衍生代码、是否需要提供源码、是否需要附带许可证和 NOTICE。 | 商业化、外部分发、客户现场交付前必须完成复核；未复核状态只能保留在内部开发分支。 |
| 进入仓库前门禁 | 检查 LICENSE/NOTICE、来源版本、修改说明、授权边界、敏感信息扫描、raw command 阻断、AI 工具白名单、测试证据。 | 任一项缺失即标记 BLOCKED，不得发布 `findx-agents` 包或把能力暴露给 AI 自动修复链路。 |

命名策略：

- 平台主名：FindX / AI WorkBench。
- 监控核心：FindX Monitoring Core。
- 探针发行版：`findx-agents`。
- 探针服务名：`findx-agents.service`。
- 探针安装路径：`/opt/findx-agents`。
- 探针配置路径：`/etc/findx-agents`。
- 探针日志路径：`/var/log/findx-agents`。
- 内部参考命名可以保留 `nightingale`、`n9e`、`categraf`、`catpaw` 作为来源和兼容标识；用户主入口使用 FindX 命名。

## 4. FindX Monitoring Core 总体目标

FindX Monitoring Core 要建设为独立运行的新监控核心平台，覆盖成熟监控平台的主要能力域，并在 AI、Agent、自动修复和模板闭环上做增强。

核心目标：

1. FindX 自己承载 target、datasource、query、alert rule、evaluator、event、notification、dashboard、template、pipeline、task、permission、audit、team、oncall、subscription、silence。
2. FindX 运行态由自有 API、数据模型、权限、Agent 和 AI 闭环承载。
3. Categraf 插件生态直接用于 `findx-agents`，不重新发明成熟采集插件。
4. Catpaw 授权能力衍生进 `findx-agents`，提供巡检、诊断、会话、结构化工具和受控修复执行器。
5. AI 问诊直接读取 FindX 自有告警、指标、日志、Dashboard、通知、巡检、修复和知识库证据。
6. 自动修复执行链路必须具备 precheck、dry-run、approve、execute、verify、rollback、audit。
7. 前端保留 FindX 当前风格，必要时融合成熟页面源码，但最终路由、接口、文案和交互归 FindX。

全功能交付硬约束：

| 交付项 | 完整标准 |
| --- | --- |
| 功能生命周期 | 覆盖列表、详情、创建、编辑、删除、启停、导入、导出、克隆、校验、试跑、状态变更、批量操作中该功能域实际支持的动作。 |
| 权限与团队 | 接入 FindX user、team、role、business group、operation permission、API token。 |
| 审计与回滚 | 所有写操作必须记录操作者、对象、前后差异、来源、审批状态、回滚引用和失败清理策略。 |
| 安全与脱敏 | token、password、secret、DSN、cookie、authorization、SSH key、内部 URL、raw error 不进入前端、日志、AI prompt 和报告。 |
| 降级与错误态 | 数据源、Agent、AI、日志源、通知渠道不可用时，页面和工作流必须给出可读状态，不白屏、不误报成功。 |
| UI 完整性 | 使用 FindX 当前 UI 风格实现真实页面、筛选、表格、详情、弹窗、空状态、错误态、加载态和权限态。 |
| 测试验收 | 覆盖正常路径、异常路径、权限路径、边界路径、断连路径、敏感信息扫描和 WSL/Linux 构建验证。 |
| 文档同步 | 运维手册、模板说明、API 契约、回滚说明、许可证/NOTICE 同步更新。 |

## 5. Nightingale 参考与改进策略

Nightingale 是成熟核心参考。参考方式分三类：

| 类型 | 处理方式 | 说明 |
| --- | --- | --- |
| 成熟核心逻辑 | 参考或融合 | 告警规则、事件、通知、静默、订阅、模板、Dashboard、数据源、权限、事件流水线。 |
| UI/交互成熟页面 | 改造为 FindX 风格 | Dashboard 编辑器、规则编辑器、通知模板、模板中心、事件流水线等可以吸收。 |
| 不适合直接继承的部分 | FindX 重构 | 过强历史命名、旧式页面体验、与 FindX AI/Agent/自动修复冲突的结构。 |

参考流程：

1. 对对应模块做源码级拆解。
2. 提取数据模型、状态机、API 行为、边界条件和测试场景。
3. 映射成 FindX 的 model/service/store/handler/page/component。
4. 前端按 FindX UI 重做或融合。
5. 用成熟平台行为作为对标基准，FindX 运行态由自身服务承载。

FindX 改进点：

- 告警事件天然接入 AI 问诊。
- 告警事件天然可触发 Agent 巡检。
- 告警规则和 Dashboard 模板具备 diff、安装、回滚、漂移检测。
- 通知中心与 Hermes/ChatOps 打通。
- 自动修复作为一等能力，而不是外部脚本。
- Agent 是采集 + 巡检 + 诊断 + 会话 + 修复执行器。
- 权限、审批、审计、回滚贯穿所有写动作。
- UI 以运维工作台组织。
- 所有 AI 结论必须绑定 evidence ref。

### 5.1 Nightingale 能力对标 Gap Matrix

本矩阵用于把 Nightingale 成熟能力拆成 FindX 可执行建设清单。`当前状态` 只描述截至本文档更新时已有证据：P0 API 基座、P1-BE-1/2/3 与 alert scheduler 后端闭环有提交和 QA 事实；其余 P1/P2/P3 能力不得写成已完成。

| 能力域 | Nightingale 参考能力 | FindX 目标 API/模型/UI | 当前状态 | 缺口 | 验收用例 | 优先级 |
| --- | --- | --- | --- | --- | --- | --- |
| target | targets、busi groups、agent meta、标签和状态展示。 | `/api/v1/monitor/targets`、`monitor_targets`、目标列表/详情/标签/健康状态。 | P0 后端基座已通过，需继续与 team、tenant、agent health 贯通。 | UI 完整页、批量导入导出、目标与 Agent/事件/Dashboard 关联未完成。 | 创建目标、重复 ident 409、越权 403、Agent 断连显示 degraded、敏感标签脱敏。 | P0/P1 |
| datasource | datasources、Prometheus 兼容、鉴权配置、连通性检查。 | `/api/v1/monitor/datasources`、`/api/v1/monitor/query*`、`monitor_datasources`、数据源管理 UI。 | P0 查询与数据源契约已有基座。 | 多数据源路由、密钥轮换、UI 密钥不回显、断连 503 与重试策略需补全。 | 新建数据源不回显 secret、query-range 限制时间窗、上游断连返回 503、审计记录脱敏。 | P0/P1 |
| query | PromQL 查询、metrics、labels、label-values。 | `/api/v1/monitor/query`、`query-range`、`metrics`、`labels`、查询审计模型。 | P0 后端已纳入基座，P1-BE-3 scheduler 已调用查询链路。 | 查询限流、缓存、超大结果截断语义、UI 查询体验和错误态需补。 | 正常 PromQL、非法 PromQL 400、超大 series 标记 truncated、断连不生成告警事件。 | P0/P1 |
| alert rule | alert rule CRUD、clone、enable、disable、tryrun、版本。 | `/api/v1/monitor/alert-rules`、`monitor_alert_rules`、规则编辑器。 | P1-BE-1/2/3 已记录为后端切片，P1-BE-3 scheduler 闭环已 QA PASS。 | 前端编辑器、导入导出、通知/静默/订阅联动、规则模板安装仍未完成。 | 创建/更新/回滚版本、tryrun 不落正式事件、禁用规则不调度、非法 evaluator 400。 | P1 |
| event | current/history event、ack、assign、mute、resolve、archive、timeline。 | `/api/v1/monitor/events/current|history`、`monitor_alert_events_*`、事件列表/详情/时间线。 | P1-BE-3 current/history 自动闭环已完成并 QA 92/100 PASS。 | UI 处置、事件聚合、抑制、通知记录、AI 问诊入口和权限态需补。 | 触发 current、恢复进 history、ack/assign 越权 403、eval 失败只写 log 不造事件。 | P1 |
| dashboard | dashboards、panels、variables、annotations、share、builtin dashboards。 | `/api/v1/monitor/dashboards`、dashboard 模型、编辑器、分享和版本回滚。 | 待实现，当前仅规划。 | API、数据表、UI 编辑器、权限、模板安装、事件关联均待落地。 | 创建 dashboard、panel 校验、导入 JSON、版本回滚、无权限不可见、空数据不白屏。 | P1 |
| template | integrations 模板、builtin components、alerts、dashboards、metrics、markdown、icon。 | `/api/v1/templates`、`monitor_templates`、模板中心、安装 diff、漂移检测。 | 待实现，当前仅规划。 | 模板包格式、依赖检查、安装回滚、漂移检测、许可证说明和 UI 未完成。 | 模板 preview、install diff、失败回滚、drift detect、导入非法包 400。 | P1 |
| notification | notify channels、message templates、send records、retry。 | `monitor_notification_*`、通知渠道/模板/规则/记录 UI，Hermes/ChatOps 入口。 | 待实现，当前仅规划。 | 渠道适配、模板预览、重试、脱敏、失败降级、移动端入口未完成。 | webhook 发送成功、渠道断连失败记录、模板变量缺失 400、正文脱敏、重试次数可见。 | P1 |
| silence | silence 创建、匹配预览、过期、审计。 | `monitor_silences`、静默列表/匹配预览/过期任务。 | 待实现，当前仅规划。 | 匹配器模型、权限、过期调度、与 event/rule 联动未完成。 | 创建静默、匹配 preview、过期自动失效、越权 403、静默命中事件不重复通知。 | P1 |
| subscription | 用户、团队、业务组订阅和个人偏好。 | `monitor_subscriptions`、订阅偏好 UI、通知路由策略。 | 待实现，当前仅规划。 | 订阅模型、优先级、退订、团队继承、通知规则合并未完成。 | 用户订阅生效、团队订阅继承、退订后不通知、冲突策略可解释。 | P1 |
| oncall | 值班表、轮转、override、handover、escalation。 | `monitor_oncall_*`、值班日历、升级链、交接记录。 | 待实现，当前仅规划。 | 值班日历、节假日/override、升级链、通知联动和权限未完成。 | 当前值班人计算、临时 override、升级超时转派、交接审计。 | P1/P2 |
| event pipeline | event pipeline list/detail/create/update/delete、tryrun、trigger、stream、executions。 | `/api/v1/monitor/event-pipelines`、pipeline 执行记录、可视化编排 UI。 | 待实现，当前仅规划。 | DSL/规则模型、tryrun、执行流、失败补偿、权限和审计未完成。 | tryrun 不落正式动作、非法节点 400、执行失败可追踪、stream 断连可恢复。 | P2 |
| task | 任务、批量操作、后台调度、执行记录。 | `/api/v1/tasks`、`monitor_tasks`、任务中心、进度和失败重试。 | 待实现，当前仅规划。 | 任务模型、幂等、取消、重试、租户隔离、后台 worker 未完成。 | 创建任务、重复 request_id 幂等、取消任务、失败重试、进度可见。 | P2 |
| permission | users、teams、roles、busi groups、operation permissions、API token。 | FindX user/team/role/business group、资源级权限、API token、前端权限态。 | P0 已覆盖部分权限基线，完整权限域待实现。 | 团队/角色/资源矩阵、前端按钮态、API token 生命周期和审计未完成。 | 读权限、写权限、跨 team 403、资源不可见 404/403 策略、token 吊销。 | P1 |
| audit | 操作审计、差异摘要、来源、trace、回滚引用。 | `monitor_audit_logs`、统一 audit middleware、审计查询 UI。 | P0/P1-BE 已强调审计和脱敏，完整审计中心待实现。 | 差异摘要、审计查询、导出、保留策略、敏感字段扫描未完成。 | 写操作生成审计、token/DSN 不入审计、按 trace_id 查询、导出脱敏。 | P1 |
| team | 团队、业务组、成员、资源归属。 | `monitor_teams`、`monitor_business_groups`、团队管理和资源归属 UI。 | 待实现，当前仅规划。 | 团队模型、成员角色、资源迁移、团队删除保护未完成。 | 创建团队、成员授权、资源归属校验、删除有资源团队失败。 | P1 |
| alert escalation | 通知升级、值班升级、未确认升级、自动转派。 | notification rule escalation、oncall escalation、事件 action timeline。 | 待实现，当前仅规划。 | 升级状态机、超时调度、重复通知抑制、升级审计未完成。 | 未 ack 超时升级、升级到值班二线、恢复后停止升级、重复通知抑制。 | P1/P2 |
| import/export | dashboards、rules、templates、pipelines 的导入导出。 | 统一 import/export API、模板包、校验报告、回滚记录。 | 待实现，当前仅规划。 | 包格式、版本兼容、冲突策略、dry-run、失败清理未完成。 | 导入 dry-run、冲突提示、非法包 400、导出不含 secret、失败回滚。 | P1/P2 |
| multi-tenant | 多租户、业务组隔离、资源权限和审计隔离。 | `tenant_id` 贯穿 target/datasource/rule/event/agent/audit，租户切换 UI。 | P0/P1 文档和 FACP 信封已要求 `tenant_id`，运行态完整隔离待实现。 | 租户模型、查询隔离、Agent 归属、跨租户审计和限流未完成。 | 租户 A 不可见租户 B 资源、Agent 注册绑定租户、跨租户 token 403、审计隔离。 | P1/P2 |

### 5.2 Nightingale 全功能不漏项清单

本清单是范围门禁，不是建议项。只要 Nightingale 源码中存在的基础监控控制面能力没有在 FindX 中形成等价设计、等价接口或明确的融合路线，即视为规划缺口。AI、Agent、自动修复只能增强这些能力，不能成为删减这些能力的理由。

| 功能域 | Nightingale 源码证据 | FindX 覆盖要求 | 分层 |
| --- | --- | --- | --- |
| 用户 / 个人资料 / 密码 / 自助 token | `center/router/router.go`、`models/user.go`、`models/user_token.go`、`memsto/user_token_cache.go` | 用户 CRUD、个人资料、密码修改、自助 token、token 撤销、权限校验、脱敏显示、审计 | 基础能力 |
| 团队 / 用户组 | `router_user_group.go`、`models/user_group.go`、`models/user_group_member.go` | 团队列表、创建、编辑、删除、成员管理、团队通知对象、团队订阅、团队权限边界 | 基础能力 |
| 角色 / 权限 / 操作点 | `router_role.go`、`router_role_operation.go`、`models/role.go`、`models/role_operation.go` | 角色 CRUD、权限点列表、角色绑定操作、用户权限查询、菜单/按钮/API 权限一致性 | 基础能力 |
| 业务组 | `router_busi_group.go`、`models/busi_group.go`、`models/busi_group_member.go`、`memsto/busi_group_cache.go` | 业务组 CRUD、成员、权限检查、标签、Dashboard/规则/静默/订阅/任务归属、多租户隔离 | 基础能力 |
| Target / 机器 / 存活 | `router_target.go`、`pushgw/router/router.go`、`models/target.go`、`models/host_meta.go`、`memsto/target_cache.go` | target 列表、状态统计、心跳、metadata、筛选、删除、标签、备注、业务组绑定、Agent/PushGW 心跳与机器存活 | 基础能力 |
| 数据源 | `router_datasource.go`、`router_datasource_db.go`、`models/datasource.go`、`memsto/datasource_cache.go` | 数据源 CRUD、插件列表、详情、启停、删除、加密配置、多类型数据源连接状态 | 基础能力 |
| 查询 / Query Gateway | `router_query.go`、`router_es.go`、`router_opensearch.go`、`router_tdengine.go`、`doc/api/proxy-query.md` | PromQL、日志查询、SQL 模板、DB 元数据、ES/OpenSearch 字段和变量、TDengine 元数据、proxy 查询认证 | 基础能力 |
| 指标说明 / 指标视图 | `router_metric_desc.go`、`router_metric_view.go`、`models/metric_view.go`、`models/builtin_metrics.go` | metrics desc、metric views、内置指标、指标类型、采集器、指标过滤器、PromQL 推荐 | 基础能力 |
| Dashboard / Board | `router_board.go`、`router_dashboard.go`、`models/board.go`、`models/dashboard.go`、`models/board_payload.go` | Dashboard 列表、详情、创建、编辑、删除、公开、纯净视图、克隆、配置、业务组绑定 | 基础能力 |
| Chart / 分享 | `router_chart_share.go`、`models/chart.go`、`models/chart_group.go`、`models/chart_share.go` | 图表模型、图表分组、分享链接、分享权限、过期/撤销、外链安全 | 基础能力 |
| Dashboard 注解 | `router_dash_annotation.go`、`models/dash_annotation.go` | 注解新增、查询、编辑、删除、Dashboard 时间线展示、权限控制 | 基础能力 |
| 告警规则 | `router_alert_rule.go`、`models/alert_rule.go`、`models/prom_alert_rule.go`、`alert/eval/*` | 规则 CRUD、导入、PromRule 导入、克隆、批量修改、校验、relabel 测试、通知 try-run、规则详情 | 基础能力 |
| 记录规则 | `router_recording_rule.go`、`models/recording_rule.go`、`alert/record/*` | recording rule 列表、CRUD、字段批量修改、调度执行、写回、错误记录 | 基础能力 |
| 当前/历史事件与详情 | `router_alert_cur_event.go`、`router_alert_his_event.go`、`router_event_detail.go`、`alert/router/router.go` | current/history 列表、详情、card、统计、删除、event-detail、eval-detail、notify records、history 清理、Service API | 基础能力 |
| 聚合视图 | `router_alert_aggr_view.go`、`models/alert_aggr_view.go` | 告警聚合视图 CRUD、按维度分组展示、收藏/默认视图、权限边界 | 基础能力 |
| 静默 | `router_mute.go`、`models/alert_mute.go`、`alert/mute/*` | 静默列表、按业务组查询、创建、预览、编辑、删除、字段批量修改、try-run、匹配规则、过期处理 | 基础能力 |
| 订阅 | `router_alert_subscribe.go`、`models/alert_subscribe.go`、`memsto/alert_subscribe_cache.go` | 订阅列表、详情、创建、编辑、删除、try-run、用户/团队/业务组订阅、个人偏好 | 基础能力 |
| 通知模板 / 消息模板 | `router_notify_tpl.go`、`router_message_template.go`、`models/notify_tpl.go`、`models/message_tpl.go` | 模板 CRUD、内容编辑、变量预览、版本、回滚、Service API | 基础能力 |
| 通知规则 | `router_notify_rule.go`、`models/notify_rule.go`、`memsto/notify_rule_cache.go` | 通知规则 CRUD、测试、自定义参数、事件流水线 try-run、Service API、渠道/模板/订阅联动 | 基础能力 |
| 通知渠道 / 通知配置 | `router_notify_channel.go`、`router_notify_config.go`、`models/notify_channel.go`、`models/notify_config.go`、`alert/sender/provider/*` | 通知渠道配置、联系人 key、webhook、notify script、notify contact、SMTP 测试、FlashDuty/PagerDuty/飞书/钉钉/企微等扩展 | 基础能力 |
| 通知记录 | `router_notification_record.go`、`models/notification_record.go`、`alert/sender/notify_record_queue.go` | 事件通知记录、发送状态、失败原因、重试、通道、目标、脱敏正文、Service 添加记录 | 基础能力 |
| 值班 / 升级 | `alert/sender/provider/flashduty_provider.go`、`pagerduty_provider.go`、通知规则与外部值班集成 | 值班表、轮转、override、handover、未确认升级、通知升级、外部值班系统集成 | 基础能力 |
| 事件流水线 | `router_event_pipeline.go`、`models/event_pipeline.go`、`models/event_pipeline_execution.go`、`alert/pipeline/*`、`doc/api/event-pipeline.md` | Pipeline CRUD、try-run、processor try-run、API trigger、SSE stream、执行记录、统计、清理、Service API | 基础能力 |
| 事件处理器 | `memsto/event_processor_cache.go`、`alert/pipeline/processor/*` | relabel、event drop、event update、callback、if/switch、AI summary、启停、顺序、失败策略 | 基础能力 |
| 任务 / 任务模板 | `router_task.go`、`router_task_tpl.go`、`models/task_tpl.go`、`models/task_record.go` | 任务模板 CRUD、标签绑定、任务列表、任务创建、任务记录、统计、权限、审计 | 基础能力 |
| 内置资产 / integrations | `router_builtin*.go`、`models/builtin_*`、`integrations/*` | 内置分类、组件、payload、指标库、dashboard/alert/collect/markdown/icon 资产安装、diff、回滚、漂移检测 | 基础能力 |
| 嵌入产品 | `router_embedded.go`、`models/embedded_product.go`、`doc/api/embedded-product.md` | embedded dashboards、embedded product CRUD、排序、隐藏、删除、权限和入口治理 | 基础能力 |
| SourceToken / API Token | `router_source_token.go`、`models/source_token.go`、`models/user_token.go` | 用户 token、source token、创建、删除、鉴权、过期/撤销、最小权限、审计、脱敏展示 | 基础能力 |
| SavedView | `router_saved_view.go`、`models/saved_view.go` | 保存查询条件、列表、创建、编辑、删除、收藏/取消收藏、按页面/用户/团队隔离 | 基础能力 |
| 日志 / Trace | `router_trace_logs.go`、`router_es_index_pattern.go`、`models/es_index_pattern.go`、`alert/router/router.go` | logs-query、log-query-batch、trace-logs、ES index pattern、日志字段/索引/变量、事件关联日志 | 基础能力 |
| AI assistant | `router_ai_assistant.go`、`models/ai_assistant*.go`、`doc/api/ai-chat.md` | 会话创建/历史/删除、消息新建/详情/历史/取消、SSE stream、Service API、页面上下文、推荐动作 | FindX 增强，但 Nightingale 已有，不能遗漏 |
| AI agent / LLM config | `router_ai_agent.go`、`router_ai_llm_config.go`、`models/ai_agent.go`、`models/ai_llm_config.go` | Agent 配置 CRUD、LLM 配置 CRUD/test、默认模型、权限、脱敏、可用性探测 | FindX 增强，但 Nightingale 已有 |
| MCP Server | `router_mcp_server.go`、`models/ai_mcp_server.go`、`aiagent/mcp/*`、`doc/api/ai-mcp-server.md` | MCP server CRUD、test、tools list、stdio/SSE/jsonrpc client、工具权限、超时、审计 | FindX 增强，但 Nightingale 已有 |
| AI Skill | `router_ai_skill.go`、`models/ai_skill.go`、`models/ai_skill_file.go`、`aiagent/skill/*`、`doc/api/ai-skill.md` | skill CRUD、导入、更新导入、文件读取/删除、Service API、内置技能同步、技能文件安全扫描 | FindX 增强，但 Nightingale 已有 |
| 内置 AI 工具与技能包 | `aiagent/tools/*.go`、`aiagent/skill/embedded/builtin/*` | 告警、规则、Dashboard、数据源、指标、静默、通知规则、订阅、target、任务模板、团队、用户等工具；所有工具必须受控、可审计 | FindX 增强，但 Nightingale 已有 |
| Service API / Agent API | `center/router/router.go` 的 `/v1/n9e`、`alert/router/router.go`、`pushgw/router/router.go` | 页面 API、service API、agent/push API 分层；服务间鉴权、最小权限、审计、限流、脱敏 | 基础能力 |

范围门禁：

- 缺用户、团队、角色、权限、业务组任一域：文档或实现判定为 **FAIL**。
- 缺 target、机器存活、Agent 心跳、target metadata：判定为 **FAIL**。
- 缺告警规则、记录规则、current/history/detail、eval detail、Trace logs：判定为 **FAIL**。
- 缺 Dashboard、Chart、分享、注解、变量、模板：判定为 **FAIL**。
- 缺通知渠道、通知配置、通知规则、通知模板/消息模板、通知记录、静默、订阅、值班/升级：判定为 **FAIL**。
- 缺事件流水线、processor、try-run、执行记录、SSE stream：判定为 **FAIL**。
- 缺 integrations、builtin component、builtin payload、builtin metrics、metric view/desc：判定为 **FAIL**。
- 缺 SourceToken/API Token、SavedView、Embedded Product：判定为 **FAIL**。
- 缺 AI assistant、AI agent、LLM config、MCP、Skill：判定为 **FAIL**，因为 Nightingale 源码已有这些能力，FindX 不能把它们视为可选。

## 6. FindX 自有监控核心模块

后端建议新增或逐步形成以下模块：

```text
api/internal/monitoring/target
api/internal/monitoring/datasource
api/internal/monitoring/query
api/internal/monitoring/rule
api/internal/monitoring/evaluator
api/internal/monitoring/event
api/internal/monitoring/notification
api/internal/monitoring/silence
api/internal/monitoring/subscription
api/internal/monitoring/dashboard
api/internal/monitoring/template
api/internal/monitoring/pipeline
api/internal/monitoring/task
api/internal/monitoring/audit
api/internal/agent
api/internal/remediation
```

第一批主 API：

```text
GET    /api/v1/monitor/health
GET    /api/v1/monitor/targets
POST   /api/v1/monitor/targets
GET    /api/v1/monitor/targets/:id
PUT    /api/v1/monitor/targets/:id
DELETE /api/v1/monitor/targets/:id

GET    /api/v1/monitor/datasources
POST   /api/v1/monitor/datasources
POST   /api/v1/monitor/query
POST   /api/v1/monitor/query-range
GET    /api/v1/monitor/metrics
GET    /api/v1/monitor/labels
GET    /api/v1/monitor/label-values

POST   /api/v1/monitor/remote-write
POST   /api/v1/findx-agents/heartbeat
```

建议核心数据表：

```text
monitor_targets
monitor_target_labels
monitor_target_metadata
monitor_business_groups
monitor_datasources
monitor_metric_metadata
monitor_query_audit_logs
monitor_api_tokens
monitor_audit_logs
```

兼容入口治理：

- `/api/v1/n9e/*` 保留为兼容、导入或参考入口，新功能进入 FindX 主路径。
- `/api/v1/catpaw/*` 保留兼容期，新主路径迁移到 `/api/v1/findx-agents/*`。
- `/api/v1/prometheus/*` 逐步纳入 `/api/v1/monitor/query*` 查询网关。

## 7. 告警核心设计

告警核心参考成熟规则和事件模型，但实现为 FindX 自有 evaluator 和事件生命周期。

必须实现：

- 规则列表、详情、创建、编辑、删除。
- 启用、禁用、克隆、导入、导出。
- 规则校验、试跑、版本、diff、回滚。
- evaluator 定时调度。
- PromQL 查询。
- 阈值判断。
- 多条件判断。
- `for` 持续时间。
- 无数据策略。
- 恢复判断。
- current event。
- history event。
- 事件 fingerprint。
- 事件聚合、去重、抑制。
- 告警处置：认领、转派、静默、恢复、归档、备注。
- 告警动作时间线。
- 告警触发 AI 问诊。
- 告警触发 findx-agents 巡检。
- 告警触发自动修复计划。

建议 API：

```text
GET    /api/v1/monitor/alert-rules
POST   /api/v1/monitor/alert-rules
GET    /api/v1/monitor/alert-rules/:id
PUT    /api/v1/monitor/alert-rules/:id
DELETE /api/v1/monitor/alert-rules/:id
POST   /api/v1/monitor/alert-rules/:id/enable
POST   /api/v1/monitor/alert-rules/:id/disable
POST   /api/v1/monitor/alert-rules/:id/clone
POST   /api/v1/monitor/alert-rules/:id/tryrun
POST   /api/v1/monitor/alert-rules/:id/rollback

GET    /api/v1/monitor/events/current
GET    /api/v1/monitor/events/history
GET    /api/v1/monitor/events/:id
POST   /api/v1/monitor/events/:id/ack
POST   /api/v1/monitor/events/:id/assign
POST   /api/v1/monitor/events/:id/mute
POST   /api/v1/monitor/events/:id/resolve
POST   /api/v1/monitor/events/:id/archive
POST   /api/v1/monitor/events/:id/diagnose
POST   /api/v1/monitor/events/:id/inspect
POST   /api/v1/monitor/events/:id/remediation-plan
```

建议数据表：

```text
monitor_alert_rules
monitor_alert_rule_versions
monitor_alert_rule_eval_logs
monitor_alert_events_current
monitor_alert_events_history
monitor_alert_event_actions
monitor_alert_event_comments
monitor_alert_fingerprints
```

## 8. Dashboard 与模板中心设计

Dashboard 最终由 FindX 自己承载。

Dashboard 必须实现：

- dashboard list/detail/create/update/delete。
- panel list/detail/create/update/delete。
- timeseries panel。
- stat panel。
- table panel。
- log panel。
- variable。
- annotation。
- share。
- favorite。
- version。
- rollback。
- permission。
- import/export。
- AI 图表解释。
- 告警事件关联。

模板中心扩展 FindX 能力：

- dashboard template。
- alert rule template。
- collect template。
- recording rule template。
- remediation template。
- runbook template。
- markdown doc。
- icon / preview。
- install wizard。
- install diff。
- install status。
- uninstall。
- rollback。
- drift detect。
- version upgrade。
- Agent 配置下发。

建议 API：

```text
GET    /api/v1/monitor/dashboards
POST   /api/v1/monitor/dashboards
GET    /api/v1/monitor/dashboards/:id
PUT    /api/v1/monitor/dashboards/:id
DELETE /api/v1/monitor/dashboards/:id
POST   /api/v1/monitor/dashboards/:id/rollback

GET    /api/v1/templates
GET    /api/v1/templates/:id
POST   /api/v1/templates/:id/preview
POST   /api/v1/templates/:id/install
POST   /api/v1/templates/:id/rollback
GET    /api/v1/templates/installed
GET    /api/v1/templates/drift
```

建议数据表：

```text
monitor_dashboards
monitor_dashboard_panels
monitor_dashboard_variables
monitor_dashboard_annotations
monitor_dashboard_versions
monitor_dashboard_shares
monitor_templates
monitor_template_versions
monitor_template_installs
monitor_template_install_logs
monitor_template_drift_records
```

## 9. 通知、静默、订阅、值班设计

通知体系增强为 FindX 运维闭环。

必须实现：

- 通知渠道：webhook、email、企业微信、钉钉、飞书、Hermes、自定义 HTTP channel。
- 通知模板：create、update、preview、tryrun、version、rollback。
- 通知规则：event match、severity match、team match、business group match、escalation、retry。
- 通知记录：发送状态、失败原因、重试次数、关联事件、脱敏正文。
- 静默：create、update、delete、enable、expire、match preview、audit。
- 订阅：user subscription、team subscription、business subscription、personal preference。
- 值班：schedule、rotation、override、escalation、handover。

改进点：

- 通知记录直接进入告警证据链。
- AI 问诊能解释为什么通知了谁、为什么没有通知。
- 值班策略和团队权限统一。
- Hermes 作为移动端和消息端入口。
- 通知内容可由 AI 摘要增强，但发送前必须经过模板和脱敏处理。

建议数据表：

```text
monitor_notification_channels
monitor_notification_templates
monitor_notification_rules
monitor_notification_records
monitor_silences
monitor_subscriptions
monitor_oncall_schedules
monitor_oncall_rotations
monitor_oncall_overrides
```

## 10. findx-agents：统一发行版、控制协议、结构化诊断工具与受控修复底座

P2 的 `findx-agents` 不是旁路包装，而是 FindX 自有 Agent 发行版。目标是把 Categraf 成熟采集生态、Catpaw 授权衍生诊断能力、FindX 控制面和受控修复执行器合并成统一安装包、统一服务、统一配置、统一协议、统一审计。

### 10.0 Agent/CMDB 运维工作台

Agent 管理不能只是在线/离线表格。FindX 需要在自己的平台风格中融合 Nightingale 的机器存活、target、业务组和权限边界，同时吸收 AutoOps 的 CMDB 主机、分组、凭据引用、部署任务、执行日志和服务市场信息架构，形成面向运维人员的 Agent/CMDB 工作台。

定位边界：

- FindX 是主平台，Agent/CMDB 工作台属于 FindX 监控运维域，不做 AutoOps 页面搬运，也不做 Nightingale 黑壳挂载。
- Nightingale 的 target、host meta、heartbeat、业务组、权限、通知、事件关联是基础能力，必须完整接入。
- AutoOps 只作为信息架构参考：主机、分组、凭据引用、服务部署、任务进度、WebSocket 日志、Agent 状态等可以借鉴；裸命令、明文凭据、默认信任 SSH 私钥、不受控 WebSSH 不能直接采纳。
- Agent 工作台和 AI 问诊联动时，AI 只能调用结构化工具、巡检模板、证据查询和已审批动作；不得直接生成 raw shell、裸 SQL、裸 HTTP 或裸 PromQL 作为执行动作。

工作台能力矩阵：

| 能力 | 来源参考 | FindX 实现要求 | 验收门禁 |
| --- | --- | --- | --- |
| 机器存活 | Nightingale `target`、`host_meta`、PushGW/heartbeat | 同时展示 `target_alive`、`agent_alive`、`last_heartbeat_at`、`last_scrape_at`、`down_seconds`、`target_miss`、配置版本、能力摘要 | Agent 离线、采集断流、target miss、配置失败必须能区分展示，并写审计 |
| 业务组分配 | Nightingale `busi_group`、AutoOps 分组树 | CMDB 资产、Target、Dashboard、告警规则、通知、任务、Agent 全部挂业务组 | 跨业务组访问必须 403 或不可见；业务组变更写审计 |
| CMDB 资产 | AutoOps `cmdbHost`、云主机导入、主机表 | 主机名、IP、OS、环境、云厂商、负责人、标签、业务组、Agent 绑定、采集状态、风险状态 | 导入/同步必须 dry-run、冲突报告、失败回滚；不写真实凭据 |
| 凭据引用 | AutoOps 凭据/SSH 入口作为反面边界 | 只保存 credential_ref，不回显 secret；凭据加密、轮换、吊销、审批、使用审计 | 任何页面、日志、审计、AI prompt 不得出现真实密码、私钥、token |
| 远程安装 | AutoOps 部署向导、任务进度 | 选择业务组、主机、凭据引用、版本、配置模板；执行 precheck、dry-run、approve、install、verify、rollback | 安装失败保留旧状态；重复 request_id 幂等；执行日志脱敏 |
| 本机安装 | Categraf/Agent 发行经验 | 提供 Linux 安装包、systemd 服务、固定路径 `/opt/findx-agents`、`/etc/findx-agents`、`/var/log/findx-agents` | WSL/Linux 安装、启动、停止、卸载、回滚脚本可重复执行 |
| 配置分发 | Categraf provider/http provider、Nightingale integrations | 全局模板、业务组覆盖、主机覆盖；version/hash、diff、灰度、回滚、漂移检测 | 配置非法拒绝；reload 失败自动保留旧配置；配置展示脱敏 |
| 单机对话 | FindX AI 增强、Catpaw session | 绑定 target/agent/session/evidence；允许查询状态、拉取脱敏日志、执行受控巡检、生成修复建议 | 不开放任意 shell；执行类动作必须审批；会话事件逐条审计 |
| 巡检证据 | Catpaw inspector/diagnose、AutoOps 巡检思路 | 巡检结果结构化为 run_id、target_id、agent_id、findings、evidence_refs、safe_summary、missing_evidence | AI 诊断必须引用 evidence_refs；无证据时必须写“数据缺失” |
| 任务日志 | AutoOps task/websocket log、Nightingale task | 安装、配置、巡检、修复统一进入任务中心，支持进度、取消、重试、日志流、结果摘要 | 日志限长脱敏；任务失败可追踪；取消和重试幂等 |
| 服务市场/模板 | AutoOps 服务市场、Nightingale integrations | 转成采集模板、Dashboard 模板、告警模板、安装模板、Runbook 模板中心 | 模板安装必须 preview、diff、依赖检查、安装记录、失败回滚 |
| WebSSH/终端 | AutoOps WebSSH 作为高风险参考 | 若保留，只作为人工运维入口，默认关闭或强审批、强审计；不进入 AI 自动修复默认通道 | 未审批不可打开；会话录制/命令审计/敏感输出脱敏 |

Agent 页面建议拆分：

| 页面 | 内容 |
| --- | --- |
| Agent 总览 | 在线率、采集健康、业务组分布、版本分布、配置漂移、待处理安装/升级/失败任务 |
| 机器存活 | target 与 Agent 双状态表，支持业务组、标签、版本、采集状态、最后心跳、down_seconds 筛选 |
| CMDB 资产 | 主机资产、业务组、负责人、环境、云厂商、凭据引用、Agent 绑定、导入/同步 |
| 安装任务 | 远程安装、本机安装指引、升级、卸载、回滚、任务日志、失败重试 |
| 配置分发 | 模板、业务组覆盖、主机覆盖、diff、灰度、reload、漂移检测、回滚 |
| 单机对话 | 绑定单台机器的 AI 会话、巡检模板、证据引用、缺失证据、修复建议 |
| 巡检证据 | 单机巡检、业务组巡检、定时巡检、证据归档、Runbook 关联 |
| 审计与权限 | 凭据使用、安装任务、配置下发、会话、巡检、修复、审批、失败回滚审计 |

### 10.1 目标架构

```text
findx-agents
  ├── collector              # Categraf inputs / logs / prometheus scrape
  ├── provider               # local config / http config / emergency override
  ├── writer                 # remote_write / log/event writer
  ├── heartbeat              # heartbeat / liveness / load / version
  ├── capability             # capability catalog / tool catalog / profile catalog
  ├── inspector              # Catpaw 衍生巡检执行
  ├── diagnose               # DiagnoseTool / structured evidence
  ├── session                # inspect / diagnose / chat session
  ├── mcp_bridge             # 受控 MCP bridge，仅白名单工具可用
  ├── remediation_executor   # 审批后 plan step 执行
  ├── offline_queue          # 离线队列 / 失败重放 / 幂等保护
  ├── local_audit            # 本地审计 / 脱敏日志 / trace_id
  └── supervisor             # reload / upgrade / rollback / self-check
```

模块边界：

- `collector` 只负责采集，不提供 AI 任意命令能力。
- `provider` 合并三层配置，输出最终有效配置和配置版本。
- `capability` 上报 Agent 能力、工具、模板、系统环境和禁用原因。
- `inspector/diagnose` 输出结构化 evidence，不直接执行修复。
- `remediation_executor` 只执行已审批的固定 plan step，不接受 AI 生成的 raw command。
- `offline_queue` 只保存可重放的心跳、采集摘要、检查结果和执行回执；涉及敏感字段必须先脱敏。

### 10.2 FindX Agent Control Protocol v1

协议目标：统一 Agent 注册、身份认证、心跳、配置下发、能力目录、巡检诊断、受控执行、离线补偿和审计回执。

通用信封：

```json
{
  "protocol": "findx.agent.control.v1",
  "message_id": "msg-<uuid>",
  "agent_id": "agent-<id>",
  "tenant_id": "tenant-<id>",
  "timestamp": 1777840200,
  "nonce": "nonce-<random>",
  "trace_id": "trace-<id>",
  "type": "heartbeat",
  "body": {},
  "signature": "hmac-sha256:<signature>"
}
```

v1 消息类型：

| 类型 | 方向 | 用途 | 关键约束 |
| --- | --- | --- | --- |
| `register.request` | Agent -> Server | 首次注册申请 | 生产默认禁止匿名注册；必须带 bootstrap token 或注册凭证。 |
| `register.challenge` | Server -> Agent | 注册挑战 | 返回 nonce、审批状态、过期时间。 |
| `register.approved` | Server -> Agent | 注册审批通过 | 下发 per-agent key、agent_id、配置版本。 |
| `register.rejected` | Server -> Agent | 注册拒绝 | 返回脱敏原因，不暴露内部策略。 |
| `heartbeat` | Agent -> Server | 在线状态、版本、负载、能力摘要 | 必须带 timestamp、nonce、signature，服务端做重放保护。 |
| `capabilities.report` | Agent -> Server | 完整能力目录 | 包含采集、巡检、诊断、修复、MCP bridge、禁用原因。 |
| `config.pull` | Agent -> Server | 拉取配置 | 带当前 version/hash，服务端返回差异或 not_modified。 |
| `config.apply.result` | Agent -> Server | 配置应用结果 | 必须返回版本、成功/失败、错误码、回滚状态。 |
| `reload.request` | Server -> Agent | 触发 reload | 必须审计 actor、reason、trace_id。 |
| `upgrade.request` | Server -> Agent | 触发升级 | 必须有包签名、版本、回滚版本。 |
| `inspect.request` | Server -> Agent | 发起巡检模板 | 只能使用已注册 template_id；`inspection.request` 仅作为兼容别名。 |
| `inspect.result` | Agent -> Server | 回传巡检结果 | 输出 `evidence_refs` 和结构化 findings；`inspection.result` 仅作为兼容别名。 |
| `diagnose.request` | Server -> Agent | 调用结构化诊断工具 | 只能调用白名单 ToolDefinition。 |
| `diagnose.result` | Agent -> Server | 回传诊断结果 | stdout/stderr 必须脱敏和限长。 |
| `session.open` | Server -> Agent | 开启受控会话 | 仅用于 inspect/diagnose/chat，不提供任意 shell。 |
| `session.event` | 双向 | 会话事件流 | 每条事件必须绑定 session_id 与 trace_id。 |
| `remediation.precheck` | Server -> Agent | 修复前检查 | 不改变生产状态。 |
| `remediation.dry_run` | Server -> Agent | 修复试跑 | 输出影响面、风险、回滚条件。 |
| `remediation.execute` | Server -> Agent | 执行已审批步骤 | 只执行 plan step；拒绝 raw command。 |
| `remediation.verify` | Server -> Agent | 验证修复效果 | 结合查询网关、巡检和本地检查。 |
| `remediation.rollback` | Server -> Agent | 回滚已执行步骤 | 必须引用原 run_id/step_id。 |
| `audit.ack` | Agent -> Server | 执行回执 | 必须包含 actor、approver、target、status、evidence_refs。 |
| `queue.flush` | Agent -> Server | 离线队列补偿 | 按 message_id 幂等去重。 |
| `revoke` | Server -> Agent | 吊销 Agent 凭证 | Agent 必须停止接收控制命令并进入受限模式。 |

认证与重放保护：

- 生产默认禁止匿名注册。
- P2 必须实现 per-agent key、注册审批、吊销、nonce/timestamp 重放保护。
- 所有控制消息必须有 `message_id`、`timestamp`、`nonce`、`trace_id`、`signature`。
- 服务端保存短期 nonce 窗口，重复 message_id 或过期 timestamp 直接拒绝。
- Agent 被吊销后只能上报最小化离线状态，不再接受巡检、诊断或修复命令。

协议契约基线：

- wire type 使用点号命名，例如 `config.pull`；需求、UI 文案和文档可显示为 `config-pull`，但 API 层必须统一映射，避免同一动作出现双契约。
- 请求和响应均使用通用信封；`body` 承载业务字段，`signature` 覆盖 `protocol`、`message_id`、`agent_id`、`tenant_id`、`timestamp`、`nonce`、`trace_id`、`type` 和规范化后的 `body`。
- `timestamp` 使用 Unix 秒，服务端默认允许时间偏移不超过 300 秒；超窗返回 `FACP_REPLAY_WINDOW_EXPIRED`。
- `nonce` 至少 128 bit 随机值，同一 `agent_id + nonce` 在服务端保留窗口内只能使用一次；重复返回 `FACP_NONCE_REPLAY`。
- 所有响应必须包含 `status`、`code`、`message`、`server_time`、`trace_id`；失败响应不得暴露内部路径、完整 DSN、token、Cookie、SSH key、原始堆栈。
- 兼容版本由 `protocol` 与 `body.compat.min_version`、`body.compat.max_version` 协商；服务端不支持时返回 `FACP_UNSUPPORTED_VERSION`，Agent 必须进入受限模式而不是继续执行控制命令。
- 审计必须记录 actor、approver、agent_id、target、action、request_hash、response_code、trace_id、tenant_id、risk、evidence_refs；高风险动作还必须记录 approval_id。
- 脱敏策略在 Agent 本地和 Server 入库前各执行一次，字段名命中 `token`、`secret`、`password`、`authorization`、`cookie`、`dsn`、`private_key` 或值形态疑似凭证时统一替换为 `<REDACTED>`。

通用请求体字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `schema_version` | string | 是 | 业务 body schema 版本，v1 首版为 `1.0`。 |
| `compat.min_version` | string | 是 | Agent 可接受的最低协议版本。 |
| `compat.max_version` | string | 是 | Agent 可接受的最高协议版本。 |
| `idempotency_key` | string | 写动作必填 | 写动作、执行动作和回执去重键；只读动作可为空。 |
| `actor` | object | Server -> Agent 必填 | 发起人脱敏标识、角色和来源；不得包含真实登录 token。 |
| `approval_id` | string | 高风险动作必填 | 审批记录引用；`remediation.execute`、`upgrade.request` 必填。 |
| `redaction_profile` | string | 是 | 脱敏策略版本，例如 `findx-default-v1`。 |

通用成功响应体：

```json
{
  "status": "ok",
  "code": "OK",
  "message": "success",
  "server_time": 1777840201,
  "trace_id": "trace-<id>",
  "body": {
    "accepted": true,
    "version": "v1",
    "evidence_refs": []
  }
}
```

通用失败响应体：

```json
{
  "status": "error",
  "code": "FACP_SIGNATURE_INVALID",
  "message": "签名校验失败",
  "server_time": 1777840201,
  "trace_id": "trace-<id>",
  "body": {
    "retryable": false,
    "safe_message": "请求未通过身份校验",
    "evidence_refs": []
  }
}
```

错误码契约：

| 错误码 | HTTP/通道语义 | 可重试 | 说明 |
| --- | --- | --- | --- |
| `FACP_BAD_REQUEST` | 400 | 否 | body 缺字段、类型错误或 schema 不兼容。 |
| `FACP_UNAUTHORIZED` | 401 | 否 | 未注册、凭证缺失或 bootstrap token 无效。 |
| `FACP_FORBIDDEN` | 403 | 否 | Agent、actor、team、tenant 或 capability 无权限。 |
| `FACP_SIGNATURE_INVALID` | 401 | 否 | HMAC/签名校验失败。 |
| `FACP_NONCE_REPLAY` | 409 | 否 | nonce 或 message_id 在窗口内重复。 |
| `FACP_REPLAY_WINDOW_EXPIRED` | 408 | 是 | timestamp 超出允许窗口。 |
| `FACP_UNSUPPORTED_VERSION` | 426 | 否 | 协议版本或 body schema 不支持。 |
| `FACP_AGENT_REVOKED` | 403 | 否 | Agent 已吊销，只允许最小心跳。 |
| `FACP_CAPABILITY_DISABLED` | 409 | 否 | capability 被禁用、降级或不支持。 |
| `FACP_APPROVAL_REQUIRED` | 412 | 否 | 高风险动作缺少审批。 |
| `FACP_APPROVAL_EXPIRED` | 412 | 否 | approval_id 已过期或已被使用。 |
| `FACP_CONFLICT` | 409 | 视情况 | 配置版本、升级版本或 run 状态冲突。 |
| `FACP_RATE_LIMITED` | 429 | 是 | 控制面或 Agent 限流。 |
| `FACP_AGENT_OFFLINE` | 503 | 是 | Agent 不在线或会话不可达。 |
| `FACP_EXECUTION_TIMEOUT` | 504 | 是 | 巡检、诊断、会话或执行超时。 |
| `FACP_INTERNAL_ERROR` | 500 | 是 | 服务端或 Agent 内部错误；返回必须脱敏。 |

逐消息请求/响应契约：

| 动作 | 请求 body 必填 | 成功响应 body | 常见错误码 | 签名/nonce/timestamp | 脱敏与审计 | 兼容版本 |
| --- | --- | --- | --- | --- | --- | --- |
| register | `bootstrap_token` 或注册凭证摘要、hostname、ip 摘要、agent_version、labels、capability_summary、compat | `agent_id`、`challenge_nonce`、`approval_status`、`config_version`；审批通过后另发 per-agent key 的安全下发流程 | `FACP_UNAUTHORIZED`、`FACP_FORBIDDEN`、`FACP_CONFLICT` | bootstrap 阶段也必须签名；匿名注册生产默认关闭；timestamp 超窗拒绝 | IP、hostname、labels 入库前脱敏；审计注册来源、审批人和授权边界 | v1 必须支持 `register.request/challenge/approved/rejected`；v0 或未知版本拒绝 |
| heartbeat | `status`、`agent_version`、`config_version`、`load`、`capability_hash`、`queue_depth` | `server_time`、`next_heartbeat_seconds`、`config_hint`、`revoke_hint` | `FACP_SIGNATURE_INVALID`、`FACP_AGENT_REVOKED`、`FACP_RATE_LIMITED` | 每次心跳新 nonce；重复 message_id 幂等返回上一结果 | 负载和错误摘要脱敏；审计状态变化、离线恢复和吊销提示 | 允许 v1 小版本扩展字段，未知字段忽略但保留摘要 |
| config-pull | `current_version`、`current_hash`、`profile`、`capability_hash`、`emergency_override_hash` | `not_modified` 或 `config_version`、`config_hash`、`config_diff`、`rollback_version`、`apply_deadline` | `FACP_FORBIDDEN`、`FACP_CONFLICT`、`FACP_UNSUPPORTED_VERSION` | 配置响应必须由服务端签名；Agent 校验后才能 apply | config 不得含明文 secret；审计配置版本、diff 摘要和下发 actor | `config.pull` 为 wire type，`config-pull` 为文档别名 |
| reload | `reload_id`、`scope`、`reason`、`actor`、`idempotency_key` | `accepted`、`reload_id`、`old_config_version`、`new_config_version`、`rollback_available` | `FACP_APPROVAL_REQUIRED`、`FACP_CONFLICT`、`FACP_EXECUTION_TIMEOUT` | Server -> Agent 命令必须签名；Agent 回执也必须签名 | 审计 actor、原因、结果；错误输出限长脱敏 | v1 reload 只允许配置 reload，不允许执行任意脚本 |
| upgrade | `upgrade_id`、`package_url` 摘要、`package_signature`、`target_version`、`rollback_version`、`approval_id` | `accepted`、`upgrade_id`、`from_version`、`target_version`、`rollback_plan_ref` | `FACP_APPROVAL_REQUIRED`、`FACP_APPROVAL_EXPIRED`、`FACP_SIGNATURE_INVALID`、`FACP_CONFLICT` | 包签名和控制消息签名都必须校验；nonce 防重复升级 | package_url 入审计时只保留域名/包摘要；审计审批、版本和回滚结果 | 仅允许 v1 支持的升级器；不支持则返回 `FACP_CAPABILITY_DISABLED` |
| capabilities | `capability_hash`、`capabilities[]`、`disabled_reason`、`tool_catalog_hash` | `accepted`、`server_policy_hash`、`disabled_by_server[]`、`required_reports[]` | `FACP_BAD_REQUEST`、`FACP_FORBIDDEN`、`FACP_UNSUPPORTED_VERSION` | 大能力清单可分片，但每片都需签名和 nonce | disabled_reason 必须可展示且脱敏；审计 capability 变化 | 未识别 capability kind 标记 unsupported，不导致 Agent 整体失败 |
| inspect | `inspection_run_id`、`template_id`、`target`、`params`、`timeout_seconds`、`actor` | `accepted`、`inspection_run_id`；结果用 `inspect.result` 返回 `summary`、`findings`、`evidence_refs` | `FACP_CAPABILITY_DISABLED`、`FACP_AGENT_OFFLINE`、`FACP_EXECUTION_TIMEOUT` | 每个 run_id 幂等；重复请求返回既有 run 状态 | params、findings、stdout 摘要脱敏；审计模板、target、actor、证据引用 | `inspection.*` 只作兼容别名，新实现统一 `inspect.*` |
| diagnose | `diagnose_run_id`、`tool_id`、`input`、`timeout_seconds`、`actor` | `accepted`、`diagnose_run_id`；结果返回结构化 output、safe_summary、`evidence_refs` | `FACP_CAPABILITY_DISABLED`、`FACP_APPROVAL_REQUIRED`、`FACP_EXECUTION_TIMEOUT` | tool_id 必须来自 ToolDefinition；签名覆盖 input schema hash | stdout/stderr 限长脱敏；raw output 默认不入审计 | schema 小版本向后兼容，破坏性变更必须升 body schema |
| session | `session_id`、`session_type`、`target`、`allowed_tools[]`、`ttl_seconds`、`actor` | `accepted`、`session_id`、`expires_at`；事件用 `session.event` 双向流转 | `FACP_FORBIDDEN`、`FACP_CAPABILITY_DISABLED`、`FACP_AGENT_OFFLINE` | session 每条 event 都需 message_id、nonce、timestamp；过期拒绝 | 会话事件逐条审计；禁止任意 shell，输出脱敏和限长 | v1 session 仅允许 inspect/diagnose/chat，不兼容 raw shell |
| remediation | `run_id`、`plan_id`、`step_id`、`phase`、`approval_id`、`target`、`precheck_refs`、`rollback_ref` | `accepted`、`run_id`、`step_status`、`impact_summary`、`evidence_refs` | `FACP_APPROVAL_REQUIRED`、`FACP_APPROVAL_EXPIRED`、`FACP_CAPABILITY_DISABLED`、`FACP_EXECUTION_TIMEOUT` | execute/rollback 必须审批且幂等；raw command 字段出现即拒绝 | 审计 actor、approver、plan、step、前后状态；所有输出脱敏 | v1 phase 固定为 precheck/dry_run/execute/verify/rollback |
| revoke | `revoke_id`、`agent_id`、`reason`、`actor`、`effective_at` | `accepted`、`revoke_id`、`mode`=`restricted`、`allowed_actions` | `FACP_FORBIDDEN`、`FACP_AGENT_REVOKED`、`FACP_CONFLICT` | 吊销命令必须签名；Agent 确认回执必须签名 | 审计吊销原因、actor、agent、后续心跳；原因脱敏 | v1 吊销后只允许最小 heartbeat 和本地审计 flush |

### 10.3 三层配置模型

配置模型：

```text
base profile
  -> server policy
  -> emergency override
  -> effective config
```

| 层级 | 来源 | 用途 | 约束 |
| --- | --- | --- | --- |
| Base Profile | 安装包内置与 `/etc/findx-agents` | 基础采集、日志路径、默认禁用项、最小安全策略 | 可离线启动；不得包含真实密钥。 |
| Server Policy | FindX 控制面下发 | 采集模板、巡检模板、能力启停、限流、上报端点、租户标签 | 必须 version/hash；应用失败自动回滚上一版本。 |
| Emergency Override | 本地紧急配置或受控运维入口 | 断网降级、强制停用高风险工具、临时采集保护 | 必须本地审计；恢复后回传控制面。 |

配置合并规则：

- `effective_config` 必须可解释，保留每个字段来自哪一层。
- 高风险能力默认关闭，只有 Server Policy 明确启用且 Agent 能力满足时才生效。
- `inputs/exec` 默认禁用，且不进入 AI 可调用工具目录。
- 所有 token、password、secret、DSN、cookie、authorization、SSH key 在配置展示、日志、prompt、审计中必须脱敏。
- 配置应用必须支持 `precheck -> apply -> verify -> rollback`。

### 10.4 结构化 Capability Schema

能力目录用于让控制面、AI 问诊和自动修复知道 Agent 可以做什么、不能做什么、为什么不能做。

```json
{
  "agent_id": "agent-<id>",
  "version": "findx-agents/0.1.0",
  "platform": {
    "os": "linux",
    "arch": "amd64",
    "hostname": "host-01"
  },
  "capabilities": [
    {
      "id": "collector.prometheus.node",
      "kind": "collector",
      "status": "enabled",
      "risk": "low",
      "requires_approval": false,
      "evidence_types": ["metric"],
      "disabled_reason": ""
    },
    {
      "id": "diagnose.linux.process_basic",
      "kind": "diagnose_tool",
      "status": "enabled",
      "risk": "medium",
      "requires_approval": false,
      "evidence_types": ["process", "snapshot"],
      "disabled_reason": ""
    },
    {
      "id": "remediation.systemd.restart",
      "kind": "remediation_action",
      "status": "enabled",
      "risk": "high",
      "requires_approval": true,
      "evidence_types": ["precheck", "dry_run", "verify"],
      "disabled_reason": ""
    }
  ]
}
```

字段约束：

- `kind` 取值：`collector`、`inspector`、`diagnose_tool`、`mcp_tool`、`session`、`remediation_action`。
- `status` 取值：`enabled`、`disabled`、`degraded`、`unsupported`、`revoked`。
- `risk` 取值：`low`、`medium`、`high`、`critical`。
- `requires_approval=true` 的能力不得由 AI 直接执行。
- `disabled_reason` 必须是可展示的脱敏说明。

### 10.5 ToolDefinition

`ToolDefinition` 是 AI 可见工具目录的唯一入口。任何未注册、未授权、未审批的本地能力不得进入 AI prompt。

```json
{
  "tool_id": "diagnose.linux.process_basic",
  "name": "进程基础诊断",
  "kind": "diagnose_tool",
  "scope": "local",
  "risk": "medium",
  "description": "采集 CPU、内存、进程树和端口占用摘要。",
  "input_schema": {
    "type": "object",
    "required": ["target"],
    "properties": {
      "target": { "type": "string" },
      "limit": { "type": "integer", "minimum": 1, "maximum": 100 }
    }
  },
  "output_schema": {
    "type": "object",
    "required": ["summary", "findings", "evidence_refs"],
    "properties": {
      "summary": { "type": "string" },
      "findings": { "type": "array" },
      "evidence_refs": { "type": "array" }
    }
  },
  "approval_policy": {
    "required": false,
    "approver_role": "",
    "expire_seconds": 0
  },
  "execution_policy": {
    "timeout_seconds": 30,
    "max_output_bytes": 65536,
    "redact": true,
    "allow_raw_command": false
  },
  "audit_policy": {
    "record_input": true,
    "record_output_summary": true,
    "record_raw_output": false
  }
}
```

硬约束：

- AI prompt 只能看到 ToolDefinition 的脱敏摘要、schema 和风险说明。
- `allow_raw_command` 必须默认为 `false`。
- `inputs/exec`、remote/exec raw command、未审批 shell、任意脚本透传不得进入工具目录。
- 工具输出必须生成 `evidence_refs`，AI 结论必须引用这些证据。

### 10.6 巡检模板

首批巡检模板：

| 模板 | 用途 | 默认风险 | 关键输出 |
| --- | --- | --- | --- |
| `linux_quick` | Linux 快速健康检查 | low | CPU、内存、磁盘、负载、systemd 摘要。 |
| `linux_deep` | Linux 深度巡检 | medium | 内核、服务、日志、资源瓶颈、异常进程。 |
| `network_basic` | 网络基础诊断 | medium | 端口、路由、DNS、连接状态、丢包摘要。 |
| `process_basic` | 进程基础诊断 | medium | 进程树、资源占用、启动参数脱敏摘要。 |
| `container_basic` | 容器基础诊断 | medium | 容器状态、重启次数、资源限制、日志摘要。 |
| `mysql_basic` | MySQL 基础诊断 | medium | 连接、慢查询摘要、锁等待、容量风险。 |
| `redis_basic` | Redis 基础诊断 | medium | 内存、连接、慢日志摘要、持久化状态。 |
| `nginx_basic` | Nginx 基础诊断 | medium | 配置校验、连接、错误日志摘要、reload 可行性。 |
| `disk_cleanup_precheck` | 磁盘清理前检查 | high | 可清理目录、风险路径、回滚/不可回滚说明。 |
| `service_restart_precheck` | 服务重启前检查 | high | 服务状态、依赖、变更窗口、回滚条件。 |

巡检输出标准：

```json
{
  "inspection_run_id": "ir-<id>",
  "template_id": "linux_quick",
  "target": "host-01",
  "status": "success",
  "started_at": "2026-05-04T04:30:00+08:00",
  "finished_at": "2026-05-04T04:30:15+08:00",
  "summary": "主机负载正常，磁盘使用率偏高。",
  "findings": [
    {
      "severity": "warning",
      "title": "磁盘使用率偏高",
      "evidence_refs": ["evidence://inspection/ir-<id>/disk/0"]
    }
  ],
  "evidence_refs": ["evidence://inspection/ir-<id>/summary"]
}
```

### 10.7 离线保护

离线策略：

- Agent 心跳超时后标记 `offline`，控制面不得派发新巡检、诊断或修复执行。
- Agent 离线期间只允许本地采集和离线队列缓存，不执行来自控制面的高风险动作。
- 离线队列必须有最大容量、最大保留时间和敏感字段脱敏策略。
- 恢复在线后通过 `queue.flush` 上报补偿消息，服务端按 `message_id` 幂等去重。
- 配置下发失败时保留上一可用版本；连续失败进入 `degraded`，并上报可读错误码。
- Agent key 被吊销时进入 `revoked`，停止接收控制命令，仅保留本地审计和最小化状态上报。

### 10.8 安全审批

注册与身份：

- 生产默认禁止匿名注册。
- 注册必须使用 bootstrap token 或一次性注册凭证。
- 控制面必须记录注册申请、审批人、审批理由、过期时间、agent_id、per-agent key 指纹。
- 支持吊销、轮换、禁用、重新审批。

工具与执行审批：

- 诊断工具按风险级别分层，`high/critical` 必须审批。
- remediation action 必须经过 `precheck`、`dry-run` 和人工或策略审批。
- 审批记录必须绑定 `plan_id`、`run_id`、`target`、`tool_id/action_id`、`approver`、`expire_at`。
- 审批过期、目标变化、配置版本变化、Agent 能力变化时必须重新审批。
- AI 只能生成建议和 plan 草案，不能绕过审批直接执行。

脱敏与审计：

- 工具输出、AI prompt、审计日志必须统一脱敏。
- 审计日志不得记录真实 token、Cookie、完整 DSN、SSH 私钥、Authorization header、原始堆栈或敏感连接串。
- 审计记录必须保留 actor、approver、agent_id、target、trace_id、message_id、tool_id/action_id、状态、错误码、evidence_refs。

### 10.9 P2 Work Units

| Work Unit | 目标 | 允许写集建议 | 验收标准 |
| --- | --- | --- | --- |
| P2-A Agent 协议与身份 | 实现 FACP v1 信封、注册审批、per-agent key、nonce/timestamp 重放保护、吊销 | `api/internal/agent`、`api/internal/security`、Agent 协议包 | 匿名注册生产禁用；重复 nonce 拒绝；吊销后不可接收控制命令。 |
| P2-B Categraf 采集融合 | 复用 inputs/provider/remote_write/heartbeat，建立三层配置模型 | `findx-agents` 采集与配置模块 | 首批 inputs 可下发、reload、回滚；`inputs/exec` 默认禁用。 |
| P2-C Capability Catalog | 实现 capability schema、ToolDefinition、禁用原因和风险分级 | Agent capability 模块与服务端 capability API | heartbeat 与 capabilities 一致；AI 只能看到白名单工具。 |
| P2-D Catpaw 巡检诊断衍生 | 衍生 inspect/diagnose/session/tool registry/MCP bridge | Agent inspector/diagnose/session 模块 | 首批巡检模板输出结构化 findings 与 evidence_refs。 |
| P2-E 离线队列与降级 | 实现 offline queue、flush、幂等、degraded/revoked 状态 | Agent queue、服务端回执接口 | 断网后不误报成功；恢复后补偿消息幂等入库。 |
| P2-F 受控修复执行底座 | 实现 remediation executor 框架，只接受审批后 plan step | Agent remediation 模块与 `/api/v1/remediation/*` 对接 | raw command 不进入主链路；precheck/dry-run/execute/verify/rollback 有审计。 |
| P2-G 合规补齐 | 补 LICENSE/NOTICE/来源版本/修改说明/授权记录 | 文档与许可证目录 | Catpaw AGPL-3.0 衍生边界清晰，分发前可复核。 |

## 11. AI 问诊与自动修复

AI 问诊是 FindX 核心增强能力，不是外挂聊天框。完整链路必须从告警事件进入证据收集，再进入 AI 推理和受控修复闭环。

### 11.1 完整链路

```text
event
  -> inspection/diagnose
  -> evidence_refs
  -> AI hypothesis and plan draft
  -> remediation precheck
  -> remediation dry-run
  -> approve
  -> execute
  -> verify
  -> rollback or close
  -> audit
  -> knowledge
```

链路说明：

- `event`：来自 current event、history event、evaluator、通知回执或人工触发。
- `inspection/diagnose`：根据事件类型、target、business group 和 Agent capabilities 选择巡检模板与诊断工具。
- `evidence_refs`：所有巡检、指标、日志、通知、修复记录必须沉淀为可引用证据。
- `AI`：AI 只能基于 evidence_refs、知识库和 runbook 生成根因假设、影响面、风险等级和修复计划草案。
- `precheck`：验证目标、Agent 在线、能力、权限、配置版本、变更窗口、业务组、风险等级和回滚条件。
- `dry-run`：输出预计动作、影响面、失败条件、回滚策略，不改变生产状态。
- `approve`：记录审批人、审批理由、审批范围、过期时间、风险接受说明。
- `execute`：只执行已审批 plan 的固定 step，不接受 AI 输出的任意命令直传。
- `verify`：通过查询网关、Agent 巡检、事件状态和业务指标验证修复效果。
- `rollback`：失败或效果不达标时按 plan/run/step 回滚；无法自动回滚时进入人工处理。
- `audit/knowledge`：审计记录与知识库案例沉淀必须保留脱敏摘要、证据引用、决策理由和后续建议。

### 11.2 AI 证据源

- current event。
- history event。
- alert rule。
- target metadata。
- business group。
- owner/team。
- metrics query。
- logs query。
- dashboard panel。
- notification record。
- silence/subscription。
- inspection run。
- diagnose record。
- remediation precheck/dry-run/run/verify/rollback。
- knowledge case。
- runbook。

### 11.3 工作流节点

```text
monitor_event_get
monitor_rule_get
monitor_metric_query
monitor_log_query
monitor_dashboard_context
agent_capability_get
agent_inspect
agent_diagnose
inspection_report_save
evidence_ref_resolve
remediation_plan_generate
remediation_precheck
remediation_dry_run
remediation_approve
remediation_execute
remediation_verify
remediation_rollback
knowledge_case_search
knowledge_case_save
```

AI 输出必须包括：

- 根因假设。
- 证据链。
- 影响面。
- 风险等级。
- 建议动作。
- 自动修复计划草案。
- 是否建议新增或调整规则。
- 是否建议新增或调整 Dashboard。
- 是否建议补充模板。
- 是否建议沉淀知识库。

### 11.4 自动修复生命周期

```text
detect
  -> diagnose
  -> generate_plan
  -> precheck
  -> dry_run
  -> approve
  -> execute
  -> verify
  -> rollback_or_close
  -> audit
  -> knowledge_archive
```

首批自动修复动作：

- systemd service restart。
- service reload。
- disk cleanup。
- log cleanup。
- container restart。
- nginx reload。
- config rollback。
- stale process kill。
- cache cleanup。
- temporary silence create。
- collect config update proposal。
- alert threshold adjustment proposal。

API：

```text
GET    /api/v1/remediation/templates
POST   /api/v1/remediation/plans
GET    /api/v1/remediation/plans/:id
POST   /api/v1/remediation/plans/:id/precheck
POST   /api/v1/remediation/plans/:id/dry-run
POST   /api/v1/remediation/plans/:id/approve
POST   /api/v1/remediation/plans/:id/execute
POST   /api/v1/remediation/runs/:id/verify
POST   /api/v1/remediation/runs/:id/rollback
GET    /api/v1/remediation/runs
```

建议数据表：

```text
remediation_templates
remediation_plans
remediation_plan_steps
remediation_approvals
remediation_runs
remediation_run_steps
remediation_artifacts
remediation_audit_logs
```

## 12. 前端页面规划

一级导航：

- 运维总览。
- 资产与探针。
- 告警中心。
- 告警规则。
- Dashboard。
- 模板中心。
- 指标查询。
- 日志查询。
- 通知中心。
- 静默订阅。
- 事件流水线。
- 自动修复。
- 任务中心。
- AI 问诊。
- 知识库。
- 团队权限。
- 系统设置。

重点页面：

- `MonitorOverview.vue`
- `Targets.vue`
- `FindXAgents.vue`
- `AlertCenter.vue`
- `AlertRules.vue`
- `Dashboards.vue`
- `DashboardEditor.vue`
- `TemplateCenter.vue`
- `MetricExplorer.vue`
- `LogExplorer.vue`
- `NotificationCenter.vue`
- `SilenceSubscribe.vue`
- `EventPipelines.vue`
- `RemediationCenter.vue`
- `TaskCenter.vue`
- `AIDiagnose.vue`
- `TeamAccess.vue`

UI 约束：

- 使用现有 FindX 风格。
- 不做营销页。
- 不裸 iframe 作为主体验。
- 表格、筛选、详情、弹窗、加载态、错误态、权限态完整。
- 桌面和移动宽度不重叠。
- 危险动作必须二次确认。
- AI 建议与执行动作必须视觉区分。

## 13. 实施顺序

| 阶段 | 目标 | 当前状态 |
| --- | --- | --- |
| P0：FindX Core 基座 | target、datasource、query、alert rule、event、audit、heartbeat、`/api/v1/monitor/*` 和 `/api/v1/findx-agents/*` 主接口。 | **已通过**：提交 `81b4531`，QA 98/100，Windows + WSL Go 测试/构建通过。 |
| P1：监控控制面全功能补齐 | evaluator、current/history event、规则试跑、版本回滚、Dashboard/Chart/分享/注解、模板中心、通知全链路、权限、团队、值班/升级、订阅、静默、记录规则、聚合视图、Trace logs、SavedView、SourceToken、嵌入产品。 | **进行中**：P1 告警后端闭环部分已完成，P1-BE-1 `82ea50d`、P1-BE-2 `1b37905`、P1-BE-3 `1c4045e` 已推送；其余 Nightingale 基础域仍待实现。 |
| P2：事件流水线、模板生态与 Agent/CMDB | 事件流水线 processor/stream/execution、任务/任务模板、integrations 资产安装 diff/回滚、Categraf 插件复用、Catpaw 授权衍生 inspector/diagnose/session/tool registry/MCP bridge、Agent/CMDB 工作台、机器存活、业务组分配、凭据引用、远程/本机安装、配置分发、单机对话。 | 待全功能实施，不得标记已完成。 |
| P3：AI 问诊与自动修复 | event 到 evidence_refs 到 AI，再到知识库记忆标签注入、Agent 巡检补证据、remediation precheck/dry-run/approve/execute/verify/rollback/audit/knowledge。 | 待全功能实施，不得标记已完成。 |
| P4：生产化治理与生态闭环 | LICENSE/NOTICE/来源版本/修改说明、模板市场治理、Run & Debug 脚本、审计留存、容量/性能压测、Linux/WSL 安装升级回滚手册。 | 待 P1/P2/P3 运行证据形成后推进。 |

详细顺序：

1. P0 后端 API 基座已完成，后续只做缺陷修复、文档同步和兼容治理。
2. P1-BE-1 域常量/状态机/fingerprint 幂等已完成并推送，提交 `82ea50d`。
3. P1-BE-2 evaluator service + 真实 tryrun 已完成并推送，提交 `1b37905`。
4. P1-BE-3 alert scheduler current/history 自动闭环已完成并推送，提交 `1c4045e feat: add FindX alert scheduler event closure`；QA **92/100 PASS**，Windows/WSL `go test -count=1 ./...` 与 build 通过。
5. 下一步从 P1-FE-1 `/monitor` 工作台与 API 封装，或 P1-BE-4 audit/permission/silence/notification/subscription/oncall/pipeline 基座开始。
6. P1 剩余阶段完成 Dashboard、Chart、分享、注解、变量、模板中心、安装 diff、回滚和漂移检测。
7. P1 剩余阶段完成通知渠道、通知配置、通知规则、通知模板/消息模板、通知记录、静默、订阅、值班/升级、权限审计和团队/多租户闭环。
8. P1 必须显式补齐记录规则、告警聚合视图、Trace logs、SavedView、SourceToken/API Token、嵌入产品和 metric view/desc。
9. P2 先实现事件流水线 processor/try-run/trigger/stream/executions，再实现 integrations 资产安装 diff/回滚和任务/任务模板。
10. P2 基于 Categraf 建设 `findx-agents` 统一发行版，并完成机器存活、业务组分配、配置分发、采集模板和 remote_write 兼容。
11. P2 融合 Catpaw 授权衍生 inspector/diagnose/session/tool registry/MCP bridge/remote security，并补 Agent/CMDB 工作台、远程/本机安装、凭据引用、单机对话、巡检证据。
12. P2 完成 FACP v1、三层配置模型、capability catalog、ToolDefinition、巡检模板、离线队列、安全审批和受控修复执行底座。
13. P3 实现知识库记忆标签 Prompt 注入、AI 问诊 evidence chain 和自动修复闭环。
14. P4 完成许可证、NOTICE、来源版本、修改说明、模板市场治理、Run & Debug、容量压测和 Linux/WSL 安装升级回滚手册。
15. 每个稳定切片完成权限、审计、文档、测试基准、WSL 构建和 Git 落库。

## 13.1 P0 到 P3 实施闭环细化

### P0：监控核心基座

| 子阶段 | 功能域 | 状态 | QA 门禁 |
| --- | --- | --- | --- |
| P0-DOC | 定制化项目文档闭环 | 已通过 | 明确 FindX 是新监控核心平台；Nightingale 仅作参考/改进对象；Categraf 直接复用；Catpaw 授权衍生；AI 问诊和自动修复是核心增强；WSL/Linux 为兼容基准。 |
| P0-1 | target、datasource 基础语义、agent register、agent heartbeat、health | 已通过 | target CRUD、agent token、heartbeat upsert、health 降级、权限、断连、脱敏、WSL 编译。 |
| P0-2 | alert rule、current/history event、tryrun、rollback、action | 已通过 | 规则 CRUD、版本自增、回滚生成新版本、tryrun 不落正式事件、事件状态机、动作审计、权限和断连。 |
| P0-3 | datasource、query、query-range、metrics、labels、label-values | 已通过 | 数据源配置不回显密钥、PromQL 校验、时间范围限制、上游断连 503、查询审计、限流、与规则 tryrun/evaluator 对接。 |

P0 稳定切片结论：

- P0 已完成后端 API 基座并通过 QA，提交 `81b4531`。
- QA 评分 98/100，结论为通过。
- Windows + WSL Go 测试/构建通过。
- P0 文档规划切片已补齐项目定制化方向，删除旧方向表述风险。
- P1-BE-1 `82ea50d`、P1-BE-2 `1b37905`、P1-BE-3 `1c4045e` 已推送；P1-BE-3 QA **92/100 PASS**，Windows/WSL `go test -count=1 ./...` 与 build 通过。
- CORS 空配置默认保持 localhost-only，Workflow DSL 创建/更新严格校验并对坏 DSL 返回 400。
- P0 的通过结论不外推到 P1 剩余域、P2/P3。
- 后续如修改 P0 已有契约，必须重新标记 API_CONTRACT_CHANGE 或 DATA_CHANGE，并补回归证据。

### P1：Dashboard、模板、Evaluator、通知、权限审计

P1 已实现切片：

- P1-BE-3 alert scheduler current/history 自动闭环已完成，提交 `1c4045e feat: add FindX alert scheduler event closure`。
- QA 结果为 **92/100 PASS**，非阻断风险已纳入后续治理；Windows/WSL `go test -count=1 ./...` 与 build 均通过。
- scheduler 负责启用规则的周期性评估、current event upsert、恢复进入 history、eval log 安全摘要和并发 RunOnce 锁保护。
- Prometheus 网络错误、非 2xx、`status:error`、invalid JSON、datasource not found、PromQL 校验失败、evaluator invalid 只写 eval log，不得制造 current/history event。
- event labels、annotations、target_ident、eval log details 写入前必须做 value 级脱敏。

P1 剩余全量实施范围：

| 功能域 | 必须完成 | 验收重点 |
| --- | --- | --- |
| dashboard | 列表、详情、创建、编辑、删除、panel、变量、annotation、分享、收藏、版本、回滚、导入导出、AI 图表解释 | 图表真实取数、空态/错误态、权限态、版本回滚、模板安装联动。 |
| template | dashboard/rule/collect/remediation/runbook 模板、预览、安装、diff、漂移检测、升级、卸载、回滚 | 安装失败清理、漂移可读、版本可追踪、模板包脱敏。 |
| evaluator | 定时调度、规则分片、查询网关调用、no_data_policy、恢复判断、去重聚合、eval log | 数据源断连不制造告警风暴；评估日志可追踪但不泄露敏感信息。 |
| notification | channel、template、rule、record、retry、preview、tryrun、rollback | 发送内容脱敏、失败可重试、通知记录进入证据链。 |
| oncall | schedule、rotation、override、escalation、handover | 值班人与团队权限一致；升级链路可审计。 |
| silence | create、update、delete、enable、expire、match preview、audit | 静默匹配可解释；过期恢复正确。 |
| subscription | user/team/business subscription、personal preference | 订阅与通知规则合并时不重复轰炸。 |
| permission | user、team、role、business group、operation permission、API token | 读写边界清晰；越权返回 403；资源不可见按 404/403 统一。 |
| audit | 所有写操作审计、差异摘要、trace_id、来源、回滚引用 | 审计记录不含 token、Cookie、完整 DSN、SSH key 或堆栈。 |

### QA 风险纳入后续计划

| 风险 | 影响 | 后续动作 |
| --- | --- | --- |
| 大基数候选截断语义 | evaluator 面对超大 series/candidate 集时，如果只返回截断后的候选，可能让用户误以为全量评估已完成。 | 定义候选上限、截断标记、总量估算、用户提示、eval log 字段和告警降级策略；QA 增加超大候选集回归。 |
| Workflow API 严格 DSL 400 回归 | DSL graph validation 一旦回退，坏工作流可能被持久化并污染后续执行。 | 固定回归创建/更新坏 DSL 返回 400 且不落库，覆盖缺节点、悬空边、循环、非法节点类型、错误分支和 parse 后 graph validation。 |
| 持久化失败一致性 | 规则、事件、eval log、审计任一写入失败时，可能出现响应成功但实际落库不完整。 | 后续实现必须明确事务边界、补偿策略、失败响应、审计降级和重复提交幂等；QA 验证存储失败注入。 |
| 统一 `.gitattributes` | Windows 与 WSL/Linux 交叉开发会放大换行、脚本可执行位和文本文件归一化漂移。 | 新增统一 `.gitattributes` 作为独立文档/治理切片，覆盖 Markdown、Go、Vue、Shell、PowerShell、YAML、JSON 和生成物排除策略。 |

### P2：findx-agents 融合与运维闭环扩展

P2 的 `findx-agents` 是 Categraf 插件生态 + Catpaw 授权衍生 inspector/diagnose/session/tool registry/MCP bridge/remote security 的完整融合：

- 采集侧复用 Categraf inputs、local/http provider、remote_write、heartbeat 和配置模板生态。
- 控制侧新增 FindX Agent Control Protocol v1，支持 register、heartbeat、config-pull、reload、upgrade、capabilities、inspection、diagnose、session、remediation、revoke。
- 巡检侧衍生 Catpaw plugin/inspect/diagnose/session/event model/remote security，输出结构化 evidence refs。
- 执行侧新增 remediation executor，只接受经审批的 plan/run，不接受任意脚本透传。
- 安全侧强制 per-agent key、签名、timestamp、nonce、幂等 run id、重放保护、本地审计和脱敏日志。
- 发布侧形成 `findx-agents.service`、`/opt/findx-agents`、`/etc/findx-agents`、`/var/log/findx-agents` 的安装、升级、回滚和卸载闭环。

P2 验收必须覆盖：

- Categraf 首批 inputs 在 FindX 配置下发后正常采集。
- Agent heartbeat 与 capabilities 反映采集、巡检、诊断、会话、修复能力。
- `inputs/exec` 默认禁用且不进入 AI 可调用工具目录。
- Agent 匿名注册生产默认禁止。
- per-agent key、注册审批、吊销、nonce/timestamp 重放保护可验证。
- Catpaw 衍生 inspector 支持 `linux_quick`、`linux_deep`、`network_basic`、`process_basic`、`container_basic`、`mysql_basic`、`redis_basic`、`nginx_basic`、`disk_cleanup_precheck`、`service_restart_precheck`。
- Agent 离线、能力缺失、超时、执行失败均返回明确状态，不误报成功。
- 工具输出、AI prompt、审计日志统一脱敏。
- Catpaw AGPL-3.0 授权衍生的 LICENSE/NOTICE、来源版本、修改说明、授权记录补齐。

### P3：AI 问诊与自动修复核心项目

P3 将 AI 问诊与自动修复作为 FindX Monitoring Core 正式核心项目交付。链路必须完整覆盖：

```text
event/detect
  -> inspection/diagnose
  -> evidence_refs
  -> ai_reasoning
  -> remediation_precheck
  -> remediation_dry_run
  -> approve
  -> execute
  -> verify
  -> rollback_or_close
  -> audit
  -> knowledge_archive
```

P3 安全保护：

- `precheck` 必须验证目标、Agent 在线、能力、权限、变更窗口、业务组、风险等级和回滚条件。
- `dry-run` 必须输出预计动作、影响面、失败条件和回滚策略，不改变生产状态。
- `approve` 必须记录审批人、审批理由、过期时间和审批范围；审批过期后不能执行。
- `execute` 只执行已批准 plan 的固定 step，不接受 AI 输出的任意命令直传。
- `verify` 通过查询网关、Agent 巡检和事件状态验证修复效果。
- `rollback` 必须可追踪到原 plan/run/step，失败时进入人工处理。
- `audit` 必须记录 actor、approver、target、plan、run、step、trace_id、前后状态和证据引用。
- 失败保护必须覆盖超时、部分成功、Agent 断连、验证失败、回滚失败、重复提交和并发执行。

## 13.2 主代理评分审计与实时 Git 策略

Claude 缺席时，主代理接管编排、决策、验收、评分、审计、优化和 Git 门禁；子代理仍是唯一执行层。主代理不写业务代码，不绕过 QA，不在阻断项未关闭时推进新功能。

主代理评分维度：

| 维度 | 权重 | 低于通过线时的动作 |
| --- | --- | --- |
| 功能正确性 | 30% | 回派归属子代理补齐正常、异常、边界和权限路径。 |
| 代码质量 | 20% | 要求拆分 God Function、去重、补错误处理和复用项目既有模式。 |
| 安全治理 | 20% | 阻断合并，直到认证、授权、脱敏、远程执行和审计风险关闭。 |
| 验证闭环 | 20% | 缺少 WSL 编译、API/UI 回归或 QA 证据时不得 PASS。 |
| 文档同步 | 10% | API_CONTRACT_CHANGE、DATA_CHANGE、Runbook、测试基准未同步时回派 doc-writer。 |

实时 Git 稳定切片策略：

- 一个稳定切片必须完成本地验证、WSL 构建或等价验证、QA 门禁、敏感信息扫描和 diff 范围检查。
- 通过验证和 QA 的稳定切片立即 commit/push，提交信息说明阶段、范围、验证和回滚点。
- QA FAIL、P0/P1 RISK、敏感信息风险、契约/数据变更未验证时不提交。
- 并行子代理输出合并前必须确认写集互斥；冲突由主代理退回或拆分，不靠手工覆盖。
- 每个 commit 保持可回滚、可复现、可解释；大功能按 P1 dashboard、P1 evaluator、P2 agent、P3 remediation 等稳定切片落库。

## 13.3 运维落地与 WSL/Linux 兼容闭环

FindX Monitoring Core 的验收以 WSL/Linux 运行态为准，Windows 只作为开发辅助环境。所有核心能力在进入稳定切片前必须说明 Linux 路径、服务、日志、脚本、权限和回滚方式。

运维落地必须覆盖：

- 团队、权限、业务组、多租户和 API token 的资源边界。
- 告警升级、值班日历、oncall rotation、handover、override 和升级链路审计。
- 通知通道、通知模板、通知记录、失败重试和脱敏正文。
- 静默、订阅、团队订阅、业务订阅和个人偏好，避免重复轰炸。
- Dashboard、告警规则、采集模板、修复模板、Runbook 模板的导入、导出、diff、安装、升级、回滚和漂移检测。
- 审计日志必须覆盖 actor、team、tenant、resource、operation、trace_id、before/after 摘要、审批引用和回滚引用。
- Run & Debug 脚本必须覆盖 WSL `/opt/ai-workbench` 后端构建、前端构建、API 冒烟、UI 回归、Agent 安装/升级/回滚/卸载。
- 日志输出必须覆盖 API、scheduler、evaluator、notification、Agent control、remediation run、AI diagnose，不得输出真实 token、Cookie、完整 DSN、SSH key 或 upstream body。
- `findx-agents` Linux 发行形态固定为 `findx-agents.service`、`/opt/findx-agents`、`/etc/findx-agents`、`/var/log/findx-agents`，并提供安装、升级、回滚、卸载和离线诊断说明。

## 14. 测试与验收

后端验证：

```bash
cd /opt/ai-workbench/api
go build -o api-linux .
go test ./...
```

前端验证：

```bash
cd /opt/ai-workbench/web
npm run build
```

必须覆盖：

- 用户、团队、角色、权限点、业务组、SourceToken/API Token。
- target CRUD。
- 机器存活、Agent 心跳、target metadata、target miss、down_seconds。
- datasource CRUD。
- metric query。
- 日志查询、Trace logs、SavedView。
- metric view、metric desc、builtin metrics。
- dashboard create/edit/view。
- Chart、Chart 分享、Dashboard 注解、变量、嵌入产品。
- alert rule CRUD。
- recording rule CRUD。
- alert rule tryrun。
- evaluator trigger。
- event recover。
- event dedup。
- event detail、eval detail、聚合视图。
- silence match。
- subscription match。
- oncall current、override、handover、escalation。
- notification channel/rule/template/message template/record/config。
- notification send。
- template install/diff/rollback、integrations dashboard/alert/collect/markdown/icon 资产安装。
- event pipeline CRUD、processor try-run、trigger、stream、executions。
- task template、task record、后台任务重试和取消。
- findx-agents register。
- findx-agents heartbeat。
- findx-agents capabilities。
- Agent/CMDB 资产、业务组分配、凭据引用、远程安装、本机安装、配置分发、单机对话。
- findx-agents inspect。
- findx-agents diagnose。
- Agent offline queue。
- Agent revoke。
- AI assistant、AI agent、LLM config、MCP server、AI skill。
- 知识库记忆标签 LTM/STM/历史记录注入策略。
- AI diagnose from event。
- remediation precheck/dry-run/approve/execute/verify/rollback。
- permission denied。
- datasource unavailable。
- agent offline。
- sensitive scan。

质量门禁：

- Go 文件不超过 400 行。
- Vue 文件不超过 300 行。
- 单函数不超过 50 行。
- 不新增不可控依赖。
- 新 API 必须有契约说明。
- 数据库变更必须有幂等迁移和回滚说明。
- 所有写操作必须有审计。
- 所有外部调用必须有超时。
- Nightingale 基础域缺失、文档缺失或无法验证时，不得以 AI/Agent 增强能力抵扣；P0/P1 基础域漏项默认 FAIL。
- 所有敏感字段必须脱敏。

## 15. 风险与治理

主要风险：

- 自研监控核心、告警引擎、Dashboard、通知、自动修复工作量大，必须按功能域拆分，不能靠单次大改堆完。
- 成熟源码融合时容易留下旧命名、旧接口和双入口，必须建立来源说明和清理计划。
- Catpaw 衍生虽然已有授权前提，仍需在仓库保留授权记录、NOTICE、来源说明和修改说明。
- 自动修复存在生产风险，必须强制 precheck、dry-run、approval、verify、rollback、audit。
- Agent 远程执行、配置下发、日志查询、通知 webhook 都是高风险面，必须有权限、签名、限流、重放保护和脱敏。
- Categraf `inputs/exec` 如误入 AI 工具目录，会形成任意命令执行风险，必须在 P2 设计和测试中阻断。
- Agent 匿名注册如在生产开启，会形成资产污染和控制面劫持风险，必须生产默认禁止。

治理要求：

- 所有新功能先明确 API_CONTRACT_CHANGE 和 DATA_CHANGE。
- 所有新增状态值、动作名、错误码集中定义。
- 所有写操作必须具备审计记录。
- 所有 AI 输出必须引用 evidence ref。
- 所有模板安装必须有 diff、安装记录、失败清理和回滚。
- 所有 Agent 执行动作必须有 session/run id，可追踪到发起人和审批人。
- 工具输出、AI prompt、审计日志必须统一脱敏。
- Catpaw AGPL-3.0 授权衍生必须补 LICENSE/NOTICE、来源版本、修改说明、授权记录。

## 16. 自评分

目标分：**>= 96**。本版计划自评分：**97/100**。

评分口径：本次只评价正式立项文档的完整性和可执行性，不把尚未实现的 P1/P2/P3/P4 能力折算成运行成果。上一版扣分集中在 Nightingale 能力对标不够矩阵化、Agent/CMDB 工作台缺独立计划、知识库记忆策略未纳入主线、FACP v1 停留在消息名、Catpaw 衍生合规缺少模板、评分偏乐观；本版已经按源码审计结果补齐“一个都不能少”的全功能矩阵、Agent/CMDB 工作台、知识库记忆标签、验收 FAIL 门禁和生产化治理路线，因此文档规划维度提升到 97 分。扣分仍来自尚未形成对应运行证据、许可证材料未实际落仓、后续 API/DATA 变更仍需随实现回填。

| 维度 | 分值 | 评分理由 |
| --- | --- | --- |
| 战略完整性 | 98 | 已明确 FindX 是主平台，Nightingale 是完整基础监控能力基线和可融合来源；AI、知识库、Agent、自动修复是增强层而非基础能力替代品。 |
| Nightingale 功能覆盖 | 98 | 已补 `5.2 Nightingale 全功能不漏项清单`，覆盖用户/团队/权限、业务组、target、机器存活、数据源、查询、指标视图、Dashboard/Chart/分享/注解、告警/记录规则/事件/聚合视图、通知全链路、事件流水线、任务、integrations、SourceToken、SavedView、嵌入产品、日志/Trace、AI assistant/agent/MCP/Skill。 |
| FindX 增强闭环 | 97 | 已把知识库记忆标签、Agent/CMDB 工作台、Catpaw 衍生巡检、单机对话、evidence chain、自动修复审批执行闭环纳入 P2/P3。 |
| Agent/CMDB 深度 | 97 | 已补 10.0 独立章节，覆盖 Nightingale 机器存活/业务组、AutoOps CMDB/凭据/任务/服务市场参考、远程安装、本机安装、配置分发、单机对话、巡检证据和 WebSSH 风险边界。 |
| 知识库记忆策略 | 96 | 已确认 memory 属于知识库能力，LTM/STM/历史记录分层，LTM 的用户偏好、场景激活、自定义标签分策略注入；仍需后续 API、表结构和 Prompt Assembler 实现验证。 |
| 安全治理 | 96 | FACP v1 已补成协议契约，覆盖签名、nonce、timestamp、脱敏、审计、兼容版本、revoke、remediation 审批；Agent/CMDB 明确凭据引用和禁止裸命令边界。 |
| 生态复用 | 96 | Categraf inputs/provider/remote_write/heartbeat 作为采集底座；Nightingale integrations 资产治理纳入模板中心；Catpaw 授权能力衍生路线和合规模板已写入。 |
| 可实施性 | 97 | P1/P2/P3/P4 路线已按可验证稳定切片拆分，并给出 API、数据表、页面、验收、FAIL 门禁和 Git 策略。 |
| Linux/WSL 验收 | 96 | 明确 Windows 只作开发辅助，构建、安装、Agent 服务、日志路径、Run & Debug 和浏览器回归以 `/opt/ai-workbench` 或目标 Linux 为准。 |

扣分项：

- P1 后端告警闭环虽已有运行证据，但 Dashboard/Chart/分享/注解、模板、通知全链路、权限、团队、值班、订阅、静默、记录规则、聚合视图、Trace logs、SavedView、SourceToken、嵌入产品仍未实现；P2/P3/P4 仍只能给出全功能闭环设计和验收要求，不能提供运行证据。
- P1-BE-3 虽已完成，但大基数候选截断语义、Workflow API 严格 DSL 400 回归、持久化失败一致性、统一 `.gitattributes` 仍需进入后续稳定切片。
- Catpaw AGPL-3.0 衍生合规已补章节模板，但仓库级 LICENSE/NOTICE、来源版本、修改说明、授权记录和商业分发复核仍需在后续实施中真实落地。
- FACP v1 已从消息名补强为协议契约，但仍需后续后端与 Agent 代码落地后反向校准字段、错误码、状态机、签名实现和重放窗口。
- 知识库记忆标签策略已经形成讨论文档，但 Prompt Assembler、数据模型、注入预览、审计和 token 预算执行仍需正式实现。

补齐措施：

- P1/P2/P3/P4 每个稳定切片必须由 `go-backend`、`vue-frontend`、`qa-tester` 按写集互斥执行，并回填构建、API/UI、权限、断连和脱敏证据。
- 将大基数候选截断、Workflow DSL 400、持久化失败一致性、统一 `.gitattributes` 纳入下一轮 QA checklist 和稳定切片计划。
- P2 首轮必须优先关闭 `inputs/exec`、匿名注册、raw command、工具目录白名单、重放保护和脱敏审计风险。
- P3 首轮只落低风险动作模板和 proposal 类动作，所有 execute 能力必须绑定 precheck、dry-run、approve、verify、rollback 和 audit。
- P4 首轮补 LICENSE/NOTICE、来源版本、修改说明、授权记录、模板市场治理和 Linux/WSL 安装升级回滚手册。
- 文档后续需随 API_CONTRACT_CHANGE、DATA_CHANGE、许可证文件、NOTICE 文件和测试基准同步更新；本版文档闭环已补齐到“可派发实现与 QA 验收”的粒度，但尚未替代真实运行证据。
