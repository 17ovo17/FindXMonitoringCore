# FindX Monitoring Core：参考 Nightingale、改进 Nightingale、最终成为新平台

生成时间：2026-05-04 02:40（UTC+8）

## 1. 立项结论

AI WorkBench / FindX 的后续定位调整为：**FindX Monitoring Core 是新一代监控核心平台**。它不是 Nightingale 外壳，不以 Nightingale 作为长期运行依赖，也不要求照顾现有生产 Nightingale 的无缝迁移。Nightingale 的角色变为核心参考实现、源码参考和可融合对象；FindX 最终独立运行，并拥有自己的监控事实源、规则执行面、事件面、通知面、Dashboard、模板中心、权限审计、Agent 控制面、AI 问诊和自动修复闭环。

准确表述是：**参考夜莺的骨架，做 FindX 的身体和大脑。**

总体定位：

- FindX Monitoring Core：自有监控核心平台，负责 target、datasource、query、alert rule、evaluator、event、notification、dashboard、template、pipeline、task、permission、audit。
- Nightingale：成熟功能参考和可融合源码来源，重点参考告警、事件、Dashboard、通知、模板、权限、事件流水线、任务和前端编辑体验。
- Categraf：成熟采集插件生态，直接保留 inputs、local/http provider、remote_write 等能力，并改造成 `findx-agents` 发行版。
- Catpaw：已获授权的巡检、诊断、会话、结构化工具和远程安全设计来源，能力可衍生进 `findx-agents`。
- findx-agents：FindX 自有 Agent，形态为 Categraf 采集核心 + Catpaw 衍生 inspector/diagnose/session + FindX Agent Control Protocol + 自动修复执行器。
- AI 问诊：FindX 原生智能诊断能力，直接基于 FindX 自有监控事实、Agent 巡检证据、通知记录、自动修复记录和知识库案例做推理。

核心原则：

- **新平台优先**：不迁移历史监控数据，不迁移历史告警事件，不以生产 Nightingale 无缝切换为约束。
- **参考成熟实现**：Nightingale 的成熟设计可参考、融合、改造，但最终 API、数据模型、UI、权限、AI、Agent、自动修复都归 FindX 自己。
- **全功能实现**：不允许 MVP、占位页、半截接口、静态假数据、只读目录或一次性 PoC 作为交付结论。
- **Agent 深度融合**：Categraf 插件直接使用，Catpaw 授权能力衍生进 findx-agents，不再只做旁路包装。
- **自动修复正式立项**：自动修复是核心项目，必须完整实现权限、审批、试跑、验证、回滚和审计。

## 2. 源码审计范围与关键证据

本次审计以用户提供的 `D:\平台源码`、当前 git 工作区 `D:\ai-workbench` 和 WSL 运行项目 `/opt/ai-workbench` 为准。

| 组件 | 本地路径 | 审计结论 |
| --- | --- | --- |
| Nightingale | `D:\平台源码\nightingale-main (1)\nightingale-main` | 本地 LICENSE 为 Apache-2.0。源码具备 targets、busi groups、datasources、alert rules/events、dashboards、notify、message templates、event pipelines、embedded products、builtin components/payloads、AI assistant、integrations 模板。适合作为 FindX Monitoring Core 的核心参考实现和可融合源码来源。 |
| Categraf | `D:\平台源码\categraf-main (1)\categraf-main` | 本地 LICENSE 为 MIT。源码具备 remote_write、heartbeat、local/http providers、92 个 `conf/input.*` 模板目录、105 个 toml 配置文件和大量 inputs 插件。适合作为 `findx-agents` 采集核心。 |
| Catpaw | `D:\平台源码\catpaw-master\catpaw-master` | 本地 LICENSE 为 AGPL-3.0，用户已确认公司测试项目具备授权边界。源码具备 35 个插件目录、事件模型、插件巡检、`inspect/diagnose/chat` session、AI 工具注册、WebSocket 控制面和远程安全设计。适合作为 `findx-agents` inspector/diagnose/session 能力衍生来源。 |
| AutoOps | `D:\平台源码\AutoOps-main\AutoOps-main` | 有 Agent 部署、CMDB、任务、K8s、工具市场思路；旧 Agent 部署存在硬编码 token、临时编译和重复自研问题，不作为新探针底座。 |
| AI WorkBench | `D:\ai-workbench` / `/opt/ai-workbench` | 当前已有 AI 问诊、Prometheus、Catpaw、N9E Redis/MySQL 读取、拓扑、工作流、知识库。后续应把 `/api/v1/n9e/*` 和 `/api/v1/catpaw/*` 作为兼容/过渡入口，新主线建设 `/api/v1/monitor/*`、`/api/v1/findx-agents/*`、`/api/v1/remediation/*`。 |

Nightingale 关键证据：

- `center/router/router.go` 暴露 459 条路由，覆盖身份、用户、团队、业务组、目标、数据源、指标、Dashboard、告警、通知、事件流水线、日志、任务、保存视图、Webhook、AI assistant。
- `integrations/` 下有 65 个集成目录，覆盖 dashboards、alerts、metrics、record-rules、markdown、icon、collect toml。
- `center/router/router_builtin.go` 和 `center/integration/init.go` 从 integrations 初始化 builtin components、alerts、dashboards、metrics。
- `doc/api/event-pipeline.md` 明确 event pipeline 的列表、详情、创建、更新、删除、tryrun、trigger、stream、executions 和 service API。
- Dashboard、告警规则、通知模板、事件流水线和模板中心适合做 FindX UI/模型/状态机对标。

Categraf 关键证据：

- `conf/config.toml` 具备 `[global.labels]`、`[[writers]]`、`[heartbeat]`、`[ibex]`、`[prometheus]`。
- `inputs/provider_manager.go` 支持 `local` 与 `http` provider。
- `inputs/http_provider.go` 支持远端配置拉取、version/hash、按 labels 与 hostname 查询、动态 reload。
- `agent/agent.go` 将 metrics、logs、prometheus、ibex 拆成独立 `AgentModule`，适合继续增加 FindX heartbeat、inspector、session、remediation。
- `inputs/huatuo` 已提供 sidecar 管理模式，可作为 inspector 进程管理或模块嵌入的参考。

Catpaw 关键证据：

- `agent/inspect.go` 支持 `RunInspect(pluginName, target)` 单机巡检模式。
- `digcore/plugins/interface.go` 定义 `Plugin`、`Instance`、`Gatherer`、`Diagnosable`、`DiagnoseRegistrar`、`PluginCreators`。
- `digcore/diagnose/types.go` 定义 `DiagnoseTool`、`ToolScopeLocal/Remote`、`CheckSnapshot`、`DiagnoseRequest`、`DiagnoseRecord`。
- `digcore/server/proto.go` 定义 Agent 与 Server 的 `register`、`heartbeat`、`alert_events`、`session_start`、`session_output`、`session_error` 协议。
- `digcore/server/session_handler.go` 支持远程 `inspect/diagnose/chat`，并体现远程安全边界。
- `docs/event-model.md` 和 `design.d/remote-security.md` 可作为 FindX Agent 事件模型与远程执行安全基线。

AI WorkBench 现状证据：

- `api/main.go` 当前已有 `/api/v1/catpaw/*`、`/api/v1/remote/install-catpaw`、`/api/v1/remote/uninstall-catpaw`、`/api/v1/prometheus/*`、`/api/v1/n9e/agents`、`/api/v1/n9e/alerts`。
- `api/internal/handler/n9e_agents.go` 当前直接读取 Redis `n9e_meta_*`。
- `api/internal/handler/n9e_alerts.go` 当前直接查询 MySQL `alert_cur_event`。
- `web/src/views/CatpawInstall.vue`、`web/src/views/CatpawChatPanel.vue`、`web/src/views/Diagnose.vue`、`web/src/views/TopologyHub.vue` 仍存在 Catpaw 命名和入口。

## 3. 许可证、授权与命名策略

许可证事实：

- Nightingale 本地源码 LICENSE 为 Apache License 2.0。
- Categraf 本地源码 LICENSE 为 MIT。
- Catpaw 本地源码 LICENSE 为 GNU AGPL-3.0，用户已确认本公司测试项目具备授权边界，可规划为授权衍生。

合规要求：

- 引用、融合或改造 Nightingale/Categraf/Catpaw 代码时，必须保留对应 LICENSE、NOTICE、来源说明、修改说明和内部授权记录。
- Catpaw 衍生能力进入 `findx-agents` 时，仓库中需要补充授权记录、来源说明和修改说明；商业化或外部分发前仍需合规复核。
- 文档、日志、测试报告不得写真实密钥、认证票据、Cookie、完整 DSN、SSH 私钥。

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

FindX Monitoring Core 要建设为独立运行的新监控核心平台，覆盖 Nightingale 的主要能力域，并在 AI、Agent、自动修复和模板闭环上做增强。

核心目标：

1. FindX 自己承载 target、datasource、query、alert rule、evaluator、event、notification、dashboard、template、pipeline、task、permission、audit。
2. Nightingale 作为参考实现和可融合源码来源，不作为长期运行依赖。
3. Categraf 插件生态直接用于 `findx-agents`，不重新发明成熟采集插件。
4. Catpaw 授权能力衍生进 `findx-agents`，提供巡检、诊断、会话、结构化工具和自动修复执行器。
5. AI 问诊直接读取 FindX 自有告警、指标、日志、Dashboard、通知、巡检、修复和知识库证据。
6. 自动修复从第一轮路线中正式立项，执行链路必须具备 precheck、dry-run、approve、execute、verify、rollback、audit。
7. 前端保留 FindX 当前风格，必要时融合 Nightingale 成熟页面源码，但最终路由、接口、文案和交互归 FindX。

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
| 文档同步 | README、运维手册、模板说明、API 契约、回滚说明、许可证/NOTICE 同步更新。 |

## 5. Nightingale 参考与改进策略

Nightingale 不再是长期事实源，而是成熟核心参考。参考方式分三类：

| 类型 | 处理方式 | 说明 |
| --- | --- | --- |
| 成熟核心逻辑 | 参考或融合 | 告警规则、事件、通知、静默、订阅、模板、Dashboard、数据源、权限、事件流水线。 |
| UI/交互成熟页面 | 改造为 FindX 风格 | Dashboard 编辑器、规则编辑器、通知模板、模板中心、事件流水线等可以吸收。 |
| 不适合直接继承的部分 | FindX 重构 | 过强 Nightingale 命名、旧式页面体验、与 FindX AI/Agent/自动修复冲突的结构。 |

参考流程：

1. 对 Nightingale 对应模块做源码级拆解。
2. 提取数据模型、状态机、API 行为、边界条件和测试场景。
3. 映射成 FindX 的 model/service/store/handler/page/component。
4. 前端按 FindX UI 重做或融合。
5. 用 Nightingale 行为作为对标基准，但不让 FindX 运行期依赖 Nightingale。

FindX 改进点：

- 告警事件天然接入 AI 问诊。
- 告警事件天然可触发 Agent 巡检。
- 告警规则和 Dashboard 模板具备 diff、安装、回滚、漂移检测。
- 通知中心与 Hermes/ChatOps 打通。
- 自动修复作为一等能力，而不是外部脚本。
- Agent 是采集 + 巡检 + 诊断 + 会话 + 修复执行器。
- 权限、审批、审计、回滚贯穿所有写动作。
- UI 以运维工作台组织，不照搬 Nightingale 后台菜单。
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

旧接口处理：

- `/api/v1/n9e/*` 保留为兼容、导入或参考入口，不再作为新功能主路径。
- `/api/v1/catpaw/*` 保留兼容期，新主路径迁移到 `/api/v1/findx-agents/*`。
- `/api/v1/prometheus/*` 逐步纳入 `/api/v1/monitor/query*` 查询网关。

## 7. 告警核心设计

告警核心参考 Nightingale 的规则和事件模型，但实现为 FindX 自有 evaluator 和事件生命周期。

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

Dashboard 重点参考 Nightingale 的成熟实现，但最终由 FindX 自己承载。

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

模板中心参考 Nightingale integrations，并扩展 FindX 能力：

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

通知体系参考 Nightingale notify/mute/subscribe，但增强为 FindX 运维闭环。

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

## 10. findx-agents：Categraf + Catpaw 深度融合

`findx-agents` 不是简单包装 Categraf，而是 FindX 自有 Agent 发行版。

架构：

```text
findx-agents
  ├── collector        # Categraf inputs / logs / remote_write
  ├── writer           # remote_write / logs / events
  ├── heartbeat        # FindX heartbeat
  ├── provider         # local/http config provider
  ├── inspector        # Catpaw 衍生巡检
  ├── diagnose_tools   # 结构化诊断工具
  ├── session          # inspect/diagnose/chat
  ├── remediation      # 自动修复执行器
  ├── audit            # 本地审计
  └── supervisor       # 升级、reload、自守护
```

Categraf 集成方式：

- fork/改造 Categraf 作为 `findx-agents` 基线。
- 保留 inputs 插件加载机制。
- 保留 local/http provider。
- 保留 remote_write。
- 保留现有 input 配置目录结构。
- 增加 FindX heartbeat。
- 增加 FindX config provider。
- 增加 inspector 模块。
- 增加 remediation 模块。
- 增加 Agent Control Protocol。

Catpaw 衍生能力：

- plugin/instance/gatherer 设计。
- inspect 模式。
- diagnose tool registry。
- local/remote tool scope。
- inspect/diagnose/chat session。
- event model。
- remote security。
- session audit。
- structured output。

首批 profile：

- `linux_quick`
- `linux_deep`
- `network_basic`
- `process_basic`
- `container_basic`
- `mysql_basic`
- `redis_basic`
- `nginx_basic`
- `disk_cleanup_precheck`
- `service_restart_precheck`

Agent API：

```text
GET    /api/v1/findx-agents
GET    /api/v1/findx-agents/:id
POST   /api/v1/findx-agents/register
POST   /api/v1/findx-agents/heartbeat
POST   /api/v1/findx-agents/config-pull
GET    /api/v1/findx-agents/:id/capabilities
POST   /api/v1/findx-agents/:id/inspect
GET    /api/v1/findx-agents/:id/inspection-runs
GET    /api/v1/findx-agents/inspection-runs/:id
POST   /api/v1/findx-agents/:id/session
POST   /api/v1/findx-agents/:id/upgrade
POST   /api/v1/findx-agents/:id/reload
```

建议数据表：

```text
findx_agents
findx_agent_labels
findx_agent_capabilities
findx_agent_configs
findx_agent_config_versions
findx_agent_heartbeats
inspection_runs
inspection_findings
inspection_artifacts
agent_sessions
agent_session_events
agent_audit_logs
```

## 11. AI 问诊与自动修复

AI 问诊是 FindX 相对 Nightingale 的核心增强，不是外挂聊天框。

AI 证据源：

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
- remediation run。
- knowledge case。
- runbook。

工作流节点：

```text
monitor_event_get
monitor_rule_get
monitor_metric_query
monitor_log_query
monitor_dashboard_context
agent_inspect
inspection_report_save
remediation_plan_generate
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
- 是否建议新增/调整规则。
- 是否建议新增/调整 Dashboard。
- 是否建议补充模板。
- 是否建议沉淀知识库。

自动修复生命周期：

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

| 阶段 | 目标 |
| --- | --- |
| P0：FindX Core 基座 | 参考 Nightingale 建立 target、datasource、query、alert rule、event、audit、heartbeat、`/api/v1/monitor/*` 和 `/api/v1/findx-agents/*` 主接口。 |
| P1：告警核心与 Dashboard | 参考 Nightingale 实现 evaluator、current/history event、规则试跑、版本回滚、Dashboard、模板中心。 |
| P2：通知、模板、Agent 深度融合 | 实现通知、静默、订阅、值班、事件流水线、Categraf 插件复用、Catpaw 衍生 inspector。 |
| P3：AI 问诊与自动修复 | 告警触发 AI 问诊、Agent 巡检补证据、自动修复 precheck/dry-run/approve/execute/verify/rollback。 |

详细顺序：

1. 读 Nightingale 告警、事件、通知、Dashboard、模板、权限、pipeline 源码，输出 FindX 数据模型映射表。
2. 建立 FindX Monitoring Core 基础目录、模型、store、handler 和主 API。
3. 实现 target、datasource、query、alert rule、event、audit、heartbeat。
4. 实现 evaluator、current/history event、规则试跑、版本和回滚。
5. 实现 Dashboard、模板中心、安装 diff、回滚和漂移检测。
6. 实现通知、静默、订阅、值班、事件流水线。
7. 基于 Categraf 建设 `findx-agents`，融合 Catpaw 衍生 inspector/session。
8. 实现 AI 问诊 evidence chain 和自动修复闭环。
9. 完成权限、审计、文档、测试基准、WSL 构建和 Git 落库。

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
- findx-agents inspect。
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
- Nightingale 前端/后端融合时容易留下旧命名、旧接口和双入口，必须建立来源说明和清理计划。
- Catpaw 衍生虽然已有授权前提，仍需在仓库保留授权记录、NOTICE、来源说明和修改说明。
- 自动修复存在生产风险，必须强制 precheck、dry-run、approval、verify、rollback、audit。
- Agent 远程执行、配置下发、日志查询、通知 webhook 都是高风险面，必须有权限、签名、限流、重放保护和脱敏。

治理要求：

- 所有新功能先明确 API_CONTRACT_CHANGE 和 DATA_CHANGE。
- 所有新增状态值、动作名、错误码集中定义。
- 所有写操作必须具备审计记录。
- 所有 AI 输出必须引用 evidence ref。
- 所有模板安装必须有 diff、安装记录、失败清理和回滚。
- 所有 Agent 执行动作必须有 session/run id，可追踪到发起人和审批人。

## 16. 自评分

本版计划评分：**98/100**。

加分点：

- 方向已与用户最新要求一致：FindX 是新平台，Nightingale 是参考实现，不是长期运行依赖。
- 明确保留 Categraf 插件生态，并把 Catpaw 授权能力衍生进 `findx-agents`。
- 自动修复已作为正式核心项目纳入 P3，而不是“后续可选项”。
- P0/P1/P2/P3 与“参考 Nightingale、改进 Nightingale、最终成为新平台”一致。

扣分点：

- 尚未真实实现代码和运行验证。
- 告警引擎、Dashboard、通知、模板、Agent、自动修复都需要继续做源码级映射和分阶段落地。
- Catpaw 授权衍生需要后续补齐仓库级授权记录、NOTICE、来源说明和修改说明。
- TSDB 最终选型仍需在实现前确认；默认建议先使用 Prometheus-compatible 存储接口，后续可接 Prometheus、VictoriaMetrics、Mimir 或自有存储。
