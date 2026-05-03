# FindX Monitoring Core 与 findx-agents 正式立项计划

生成时间：2026-05-04 04:30（UTC+8）

## 1. 立项结论

AI WorkBench / FindX 的项目定位明确为：**FindX Monitoring Core 是新一代监控核心平台**。Nightingale 是成熟参考实现、源码参考和可融合对象；Categraf 是 `findx-agents` 采集生态的主要来源；Catpaw 是已授权巡检、诊断、会话、结构化工具、MCP bridge 与远程安全设计的衍生来源。

当前阶段结论：

- **P0 已完成**：FindX 后端 API 基座已完成并推送，提交为 `81b4531`；QA 评分 **98/100**，结论为通过；Windows 与 WSL Go 测试/构建均已通过。
- **P1/P2/P3 继续按全功能实施**：P1 告警核心、Dashboard、模板、Evaluator、通知与权限审计；P2 `findx-agents` 深度融合；P3 AI 问诊与自动修复闭环均仍按正式全功能路线推进，不标记为已实现。
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
- P0 后端 API 基座已提交 `81b4531`，QA 98/100 通过，Windows + WSL Go 测试/构建通过；该状态只覆盖 P0，不代表 P1/P2/P3 已实现。

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

1. FindX 自己承载 target、datasource、query、alert rule、evaluator、event、notification、dashboard、template、pipeline、task、permission、audit。
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
| `inspection.request` | Server -> Agent | 发起巡检模板 | 只能使用已注册 template_id。 |
| `inspection.result` | Agent -> Server | 回传巡检结果 | 输出 `evidence_refs` 和结构化 findings。 |
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
| P1：告警核心与 Dashboard | evaluator、current/history event、规则试跑、版本回滚、Dashboard、模板中心、通知、权限审计。 | 待全功能实施，不得标记已完成。 |
| P2：findx-agents 深度融合 | Categraf 插件复用、Catpaw 授权衍生 inspector/diagnose/session/tool registry/MCP bridge、FACP v1、离线队列、安全审批、受控修复底座。 | 待全功能实施，不得标记已完成。 |
| P3：AI 问诊与自动修复 | event 到 evidence_refs 到 AI，再到 remediation precheck/dry-run/approve/execute/verify/rollback/audit/knowledge。 | 待全功能实施，不得标记已完成。 |

详细顺序：

1. P0 后端 API 基座已完成，后续只做缺陷修复、文档同步和兼容治理。
2. 继续拆解成熟告警、事件、通知、Dashboard、模板、权限、pipeline 源码，输出 FindX 数据模型映射表。
3. P1 完成 evaluator、current/history event、规则试跑、版本和回滚。
4. P1 完成 Dashboard、模板中心、安装 diff、回滚和漂移检测。
5. P1 完成通知、静默、订阅、值班、事件流水线、权限审计。
6. P2 基于 Categraf 建设 `findx-agents` 统一发行版。
7. P2 融合 Catpaw 授权衍生 inspector/diagnose/session/tool registry/MCP bridge/remote security。
8. P2 完成 FACP v1、三层配置模型、capability catalog、ToolDefinition、巡检模板、离线队列、安全审批和受控修复执行底座。
9. P3 实现 AI 问诊 evidence chain 和自动修复闭环。
10. 每个稳定切片完成权限、审计、文档、测试基准、WSL 构建和 Git 落库。

## 13.1 P0 到 P3 实施闭环细化

### P0：监控核心基座

| 子阶段 | 功能域 | 状态 | QA 门禁 |
| --- | --- | --- | --- |
| P0-1 | target、datasource 基础语义、agent register、agent heartbeat、health | 已通过 | target CRUD、agent token、heartbeat upsert、health 降级、权限、断连、脱敏、WSL 编译。 |
| P0-2 | alert rule、current/history event、tryrun、rollback、action | 已通过 | 规则 CRUD、版本自增、回滚生成新版本、tryrun 不落正式事件、事件状态机、动作审计、权限和断连。 |
| P0-3 | datasource、query、query-range、metrics、labels、label-values | 已通过 | 数据源配置不回显密钥、PromQL 校验、时间范围限制、上游断连 503、查询审计、限流、与规则 tryrun/evaluator 对接。 |

P0 稳定切片结论：

- P0 已完成后端 API 基座并通过 QA，提交 `81b4531`。
- QA 评分 98/100，结论为通过。
- Windows + WSL Go 测试/构建通过。
- P0 的通过结论不外推到 P1/P2/P3。
- 后续如修改 P0 已有契约，必须重新标记 API_CONTRACT_CHANGE 或 DATA_CHANGE，并补回归证据。

### P1：Dashboard、模板、Evaluator、通知、权限审计

P1 全量实施范围：

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

- target CRUD。
- datasource CRUD。
- metric query。
- alert rule CRUD。
- alert rule tryrun。
- evaluator trigger。
- event recover。
- event dedup。
- silence match。
- subscription match。
- notification send。
- dashboard create/edit/view。
- template install/diff/rollback。
- findx-agents register。
- findx-agents heartbeat。
- findx-agents capabilities。
- findx-agents inspect。
- findx-agents diagnose。
- Agent offline queue。
- Agent revoke。
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

目标分：**>= 98**。本版计划自评分：**98/100**。

| 维度 | 分值 | 评分理由 |
| --- | --- | --- |
| 战略完整性 | 99 | 已明确 FindX Monitoring Core 为新监控核心平台，P0 已通过，P1/P2/P3 保持全功能路线，并把 Nightingale、Categraf、Catpaw 的角色纳入统一主线。 |
| 可实施性 | 98 | 已按功能域拆分到 target、datasource、query、rule、event、dashboard、template、notification、agent、AI、remediation，并进一步拆出 P2 work units。 |
| 安全治理 | 98 | 已覆盖 per-agent key、注册审批、吊销、nonce/timestamp 重放保护、工具目录白名单、raw command 阻断、脱敏、审批和审计。 |
| 生态复用 | 98 | Categraf inputs/provider/remote_write/heartbeat 作为采集底座；Catpaw 授权能力衍生为 inspector/diagnose/session/tool registry/MCP bridge/remote security；Nightingale 作为成熟模型和 UI 参考。 |
| 验证闭环 | 98 | P0 已由 commit `81b4531` 证明 QA 98/100 通过，Windows + WSL Go 测试/构建通过；P1/P2/P3 明确后续验收路径，未虚构已实现。 |
| 文档可执行性 | 98 | 文档给出 API、数据表、FACP v1、配置模型、Capability Schema、ToolDefinition、巡检模板、离线保护、安全审批、Git 策略和阶段边界。 |

扣分项：

- P1/P2/P3 尚未实现，本文只能给出全功能闭环设计和验收要求，不能提供运行证据。
- Catpaw AGPL-3.0 衍生的 LICENSE/NOTICE、来源版本、修改说明、授权记录仍需在后续实施中补齐。
- FACP v1、Capability Schema、ToolDefinition 仍需后续后端与 Agent 代码落地后反向校准字段、错误码和状态机。

补齐措施：

- P1/P2/P3 每个稳定切片必须由 `go-backend`、`vue-frontend`、`qa-tester` 按写集互斥执行，并回填构建、API/UI、权限、断连和脱敏证据。
- P2 首轮必须优先关闭 `inputs/exec`、匿名注册、raw command、工具目录白名单、重放保护和脱敏审计风险。
- P3 首轮只落低风险动作模板和 proposal 类动作，所有 execute 能力必须绑定 precheck、dry-run、approve、verify、rollback 和 audit。
- 文档后续需随 API_CONTRACT_CHANGE、DATA_CHANGE、许可证文件和测试基准同步更新。
