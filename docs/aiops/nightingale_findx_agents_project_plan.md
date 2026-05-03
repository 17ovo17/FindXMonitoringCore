# FindX × Nightingale 全功能深度集成与 findx-agents 立项开发计划

生成时间：2026-05-04 01:30（UTC+8）

## 1. 立项结论

AI WorkBench / FindX 的后续定位不应是“替代 Nightingale”，也不应只接入 Nightingale 的少量 API。正确定位是：**FindX 做 Nightingale 的全功能运维门户、产品化外壳、统一 UI 和 AI 增强层；Nightingale 继续作为核心监控平台、事实源和执行引擎。**

这意味着用户后续不需要从 Nightingale 换平台。既有 Nightingale 的能力、数据、团队、规则、模板和运维流程应全部保留，并通过 FindX 当前 UI 风格重新组织为更顺手的运维工作台。AI 问诊是锦上添花：它负责解释、归纳、推荐、复盘和知识沉淀，不负责替代 Nightingale 的规则引擎、告警引擎、通知引擎和 dashboard 引擎。

总体定位：

- Nightingale：核心监控平台和事实源，负责目标资产、业务组、数据源、告警规则、告警事件、订阅、静默、通知、仪表盘、记录规则、指标字典、集成模板、事件流水线、日志查询、任务模板、团队/RBAC 和监控运维能力。
- Categraf：采集底座，负责主机、数据库、中间件、Kubernetes、日志、黑盒拨测、硬件与云服务等成熟采集能力。
- findx-agents：平台自有探针品牌和安装/存活入口，基于 Categraf 做产品化包装，不直接从 Catpaw 派生，不 fork Categraf 采集插件。
- AI WorkBench / FindX：统一品牌入口、Nightingale 全功能运维门户、FindX 当前 UI 风格、团队/业务视图、模板中心、告警处置台、值班驾驶舱、AI 问诊、证据编排、诊断报告、模板推荐、知识沉淀、Hermes/消息入口和 findx-agents 存活视图。

核心原则：

- **不换平台**：Nightingale 是主平台和事实源，FindX 是统一入口和增强层。
- **全功能集成**：Nightingale 已有功能都纳入 FindX 路线图，不只接 targets/alerts/dashboard。
- **不重造引擎**：FindX 不重写 Nightingale 的告警、规则、通知、静默、订阅、记录规则、日志、任务、事件流水线、dashboard 和权限引擎。
- **体验统一**：用户侧使用 FindX 命名和 AI WorkBench 当前 UI 风格；能力侧保留 Nightingale/Categraf。
- **AI 锦上添花**：AI 问诊只做解释、归纳、推荐、复盘和知识沉淀，所有写入动作必须人工确认、权限校验和审计。

## 2. 源码审计范围与关键证据

本次审计以用户提供的 `D:\平台源码`、当前 git 工作区 `D:\ai-workbench` 和 WSL 运行项目 `/opt/ai-workbench` 为准。

| 组件 | 本地路径 | 审计结论 |
| --- | --- | --- |
| Nightingale | `D:\平台源码\nightingale-main (1)\nightingale-main` | 本地 LICENSE 为 Apache-2.0。源码具备 targets、busi groups、datasources、alert rules/events、dashboards、notify、message templates、event pipelines、embedded products、builtin components/payloads、AI assistant、integrations 模板。适合作为监控事实源和模板事实源。 |
| Categraf | `D:\平台源码\categraf-main (1)\categraf-main` | 本地 LICENSE 为 MIT。源码具备 remote write、N9E heartbeat、local/http providers、92 个 `conf/input.*` 模板目录、105 个 toml 配置文件和大量 inputs 插件。适合作为 findx-agents 采集底座。 |
| Catpaw | `D:\平台源码\catpaw-master\catpaw-master` | 本地 LICENSE 为 AGPL-3.0。具备 35 个插件目录、事件模型、插件巡检、`inspect/diagnose/chat` session、AI 工具注册、WebSocket 控制面、告警上报和“远程禁用 shell”的安全设计。适合作为 findx-agents Inspector 能力参考，不建议直接把源码合入 Categraf 主干。 |
| AutoOps | `D:\平台源码\AutoOps-main\AutoOps-main` | 有 Agent 部署、CMDB、任务、K8s、工具市场思路；旧 Agent 部署存在硬编码 token、临时编译和重复自研问题，不建议作为新探针底座。 |
| AI WorkBench | `/opt/ai-workbench` | 当前已有 AI 问诊、Prometheus、Catpaw、N9E Redis/MySQL 读取、拓扑、工作流、知识库。需要把监控/告警/模板事实源迁移到 Nightingale，并把 Catpaw 命名迁移到 findx-agents。 |

Nightingale 模板证据：

- `integrations/` 下有 65 个集成目录。
- 文件类型统计：178 个 `.json`、76 个 `.png`、70 个 `.toml`、64 个 `.md`、1 个 `.svg`、1 个 `.example`。
- 模板类型统计：42 个告警规则 JSON、118 个 dashboard JSON、16 个 metrics JSON、2 个 record-rules JSON、64 个 markdown 文档、70 个 Categraf toml。
- 典型模板包括 Linux、Kubernetes、MySQL、Redis、ClickHouse、Kafka、Elasticsearch、Nginx、HTTP_Response、Net_Response、Ping、Procstat、Oracle、PostgreSQL 等。

Nightingale API/路由证据：

- `center/router/router.go` 暴露 `/users`、`/user-groups`、`/roles`、`/auth/perms`、`/self/perms`、`/busi-groups`、`/busi-group/:id/perm/:perm`。
- `center/router/router.go` 暴露 `/targets`、`/targets/stats`、`/targets/tags`、`/targets/bgids`、`/target/extra-meta`。
- `center/router/router.go` 暴露 `/datasource/list`、`/datasource/upsert`、`/datasource/status/update`、`/datasource/plugin/list`。
- `center/router/router.go` 暴露 `/busi-group/:id/alert-rules`、`/alert-rule/:arid`、`/alert-rule/:arid/pure`、`/alert-rules/import`、`/alert-rules/clone`、`/alert-rules/notify-tryrun`、`/alert-rules/enable-tryrun`。
- `center/router/router.go` 暴露 `/alert-cur-events/list`、`/alert-cur-events/card`、`/alert-his-events/list`、`/alert-cur-events/stats`、`/alert-aggr-views`、`/event-notify-records/:eid`、`/trace-logs/:traceid`、`/alert-eval-detail/:id`。
- `center/router/router.go` 暴露 `/alert-mutes`、`/active-alert-mutes`、`/alert-mute-tryrun`、`/alert-subscribes`、`/alert-subscribe/alert-subscribes-tryrun`。
- `center/router/router.go` 暴露 `/recording-rules`、`/recording-rule/:rrid`。
- `center/router/router.go` 暴露 `/metrics/desc`、`/metric-views`、`/builtin-metric-filters`、`/builtin-metrics`、`/builtin-metric-promql`。
- `center/router/router.go` 暴露 `/integrations/icon/:cate/:name`、`/boards`、`/board/:bid`、`/board/:bid/pure`、`/embedded-dashboards`。
- `center/router/router.go` 暴露 `/notify-tpls`、`/message-templates`、`/notify-channel-configs`、`/event-pipelines`、`/event-pipeline/:id/trigger`、`/event-pipeline/:id/stream`。
- `center/router/router.go` 暴露 `/task-tpls`、`/job-tasks`、`/servers`、`/server-clusters`。
- `center/router/router.go` 暴露 `/es-index-pattern`、`/logs-query`、`/log-query-batch`、`/log-query`。
- `center/router/router.go` 暴露 `/embedded-product`、`/dashboard-annotations`、`/saved-views`、`/user-variable-configs`。
- `center/router/router.go` 暴露 AI 相关 `/ai-agents`、`/ai-llm-configs`、`/ai-skills`、`/mcp-servers`、`/assistant/*`。
- `center/router/router.go` 暴露 service API：`/v1/n9e/builtin-components`、`/v1/n9e/builtin-payloads`、`/v1/n9e/message-templates`、`/v1/n9e/event-pipelines`、`/v1/n9e/notify-channels`。
- `center/router/router.go` 暴露 service API：`/v1/n9e/users`、`/v1/n9e/user-groups`、`/v1/n9e/user-group-members`、`/v1/n9e/targets`、`/v1/n9e/alert-rules`、`/v1/n9e/alert-cur-events`、`/v1/n9e/alert-his-events`、`/v1/n9e/datasources`、`/v1/n9e/recording-rules`、`/v1/n9e/alert-mutes`、`/v1/n9e/task-tpls`。
- `center/router/router_builtin.go` 从 `integrations` 读取 dashboards、alerts、icons、markdown。
- `center/integration/init.go` 从 `integrations` 初始化 builtin components、alerts、dashboards、metrics，并维护 `BuiltinPayloadInFile`。
- `center/router/router_builtin_payload.go` 支持 payload 类型 `alert`、`dashboard`、`collect`，其中 `collect` 会按 TOML 校验。
- `doc/api/event-pipeline.md` 明确 event pipeline 的列表、详情、创建、更新、删除、tryrun、trigger、stream、executions 和 service API。
- `doc/api/embedded-product.md` 明确第三方产品嵌入入口，可用于 Nightingale 与 FindX 的双向入口整合。

Categraf 证据：

- `conf/config.toml` 默认 `providers = ["local"]`，提供 `[global.labels]`、`[[writers]]`、`[heartbeat]`、`[ibex]`、`[prometheus]`。
- `conf/config.toml` 默认 writers 指向 `/prometheus/v1/write`，heartbeat 指向 `/v1/n9e/heartbeat`。
- `heartbeat/heartbeat.go` 上报 `agent_version`、`os`、`arch`、`hostname`、`cpu_num`、`cpu_util`、`mem_util`、`global_labels`、`host_ip`、`unixtime`。
- `inputs/provider_manager.go` 支持 `local` 与 `http` provider。
- `inputs/http_provider.go` 支持远端配置拉取、version/hash、按 global labels 与 agent hostname 查询、动态 reload。
- `inputs/local_provider.go` 按 `input.*` 目录加载 `.yaml/.yml/.json/.toml`。
- `agent/agent.go` 将 metrics、logs、prometheus、ibex 拆成独立 `AgentModule`，适合在 findx-agents 外层再加 Inspector 模块。
- `inputs/huatuo` 已提供 sidecar 模式：Categraf 管理本地 `huatuo-bamai` 进程、覆盖配置、守护进程并抓 metrics。这个模式可作为 findx-inspector 侧车集成参考。
- `ibex/task.go` 支持远程任务执行，但具备任意脚本/命令属性；FindX 不开放裸脚本执行，必须先把任务封装为模板化、审批化、可审计、可回滚的结构化运维动作。

Catpaw 证据：

- `agent/inspect.go` 支持单机巡检模式：`RunInspect(pluginName, target)` 构造 `DiagnoseRequest{Mode: inspect}`，本机插件可直接巡检，远端插件可从配置解析 target。
- `agent/runner.go` 插件按 instance 定时 `Gather()`，产出事件并推送到 engine；panic 会被捕获并转为 critical event。
- `digcore/plugins/interface.go` 定义 `Plugin`、`Instance`、`Gatherer`、`Diagnosable`、`DiagnoseRegistrar`、`PluginCreators`，插件既能产出告警事件，也能注册诊断工具。
- `digcore/diagnose/types.go` 定义 `DiagnoseTool`、`ToolScopeLocal/Remote`、`CheckSnapshot`、`DiagnoseRequest`、`DiagnoseRecord`，区分 alert 与 inspect 两种诊断模式。
- `digcore/server/proto.go` 定义 Agent 与 Server 的 `register`、`heartbeat`、`alert_events`、`session_start`、`session_output`、`session_error` 协议。
- `digcore/server/session_handler.go` 支持远程 `inspect/diagnose/chat`，并明确远程 chat 不允许 shell，只使用结构化诊断工具。
- `docs/event-model.md` 定义 event_time、event_status、alert_key、labels、attrs、description，用 labels 生成稳定 AlertKey。
- `design.d/ai-diagnose.md` 明确 Catpaw 定位是按需诊断，不替代 Exporter；AI 调用结构化工具，告警失败不影响诊断，诊断失败不影响告警。
- `design.d/remote-security.md` 明确远程禁用 shell、结构化只读工具、Per-Agent Key、mTLS、session 签名、审计日志和 ACL 分层。

AI WorkBench 现状证据：

- `api/main.go` 当前已有 `/api/v1/catpaw/*`、`/api/v1/remote/install-catpaw`、`/api/v1/remote/uninstall-catpaw`、`/api/v1/prometheus/*`、`/api/v1/n9e/agents`、`/api/v1/n9e/alerts`。
- `api/internal/handler/n9e_agents.go` 当前直接读取 Redis `n9e_meta_*`。
- `api/internal/handler/n9e_alerts.go` 当前直接查询 MySQL `alert_cur_event`。
- `web/src/views/CatpawInstall.vue`、`web/src/views/CatpawChatPanel.vue`、`web/src/views/Diagnose.vue`、`web/src/views/TopologyHub.vue` 仍存在 Catpaw 命名和入口。

目录中的 Dify、FastGPT、RAGFlow 属于 AI/RAG 参考项目，不是本次“夜莺 + findx-agents”监控集成主线。其中 Dify/FastGPT 本地 LICENSE 带有额外商业条件，RAGFlow 为 Apache-2.0；后续若要引入其代码或 UI，需要单独立项做许可证、产品边界和安全评估。本轮不再把它们作为 Nightingale 集成扣分项。

## 3. 许可证与命名策略

### 3.1 许可证事实

- Nightingale 本地源码 LICENSE 为 Apache License 2.0。
- Categraf 本地源码 LICENSE 为 MIT。
- Catpaw 本地源码 LICENSE 为 GNU AGPL-3.0。

### 3.2 许可证影响

Nightingale 和 Categraf 适合做深度集成、二次包装和产品化组合，但仍需保留开源许可证、版权声明、NOTICE/THIRD_PARTY 文件和修改说明。

Catpaw 是 AGPL-3.0。如果直接复制、派生或重命名为 findx-agents，后续网络服务分发、商业闭源和私有化交付会带来更强的开源义务与合规风险。因此 findx-agents 不应以 Catpaw 代码派生作为主路线。

### 3.3 命名策略

产品与 UI 命名：

- 探针统一命名为 `findx-agents`。
- 平台内监控入口可命名为 `FindX Monitor`、`FindX Observability`、`FindX 告警中心`、`FindX 仪表盘`、`FindX 模板中心`。
- 用户可见的探针服务名建议为 `findx-agents.service`。
- 安装路径建议为 `/opt/findx-agents`。
- 配置路径建议为 `/etc/findx-agents`。
- 日志路径建议为 `/var/log/findx-agents`。

合规与工程命名：

- 源码和文档保留 Nightingale/Categraf 的开源归属说明。
- 内部 connector 可使用 `nightingale`、`n9e`、`categraf` 作为技术标识。
- 不在产品主导航里暴露 Catpaw；旧 Catpaw 入口后续迁移为 findx-agents 兼容页。

## 4. 全功能集成范围与目标

这次重新审计后，立项范围需要从“接入 Nightingale 的重点功能”升级为“FindX 承接 Nightingale 全功能入口”。这里的“全功能”不是复制一套 Nightingale，而是把 Nightingale 的成熟能力作为事实源和执行源，在 FindX 中用现有产品风格重新组织入口、权限、场景和 AI 增强。

### 4.0 全功能交付硬约束

本项目不允许以 MVP、占位页、半截接口、静态假数据、只读目录或一次性 PoC 作为交付结论。阶段拆分只是交付顺序，不代表功能缩水。每个被纳入当前阶段的功能域必须达到以下标准：

| 交付项 | 完整标准 |
| --- | --- |
| 功能生命周期 | 覆盖列表、详情、创建、编辑、删除、启停、导入、导出、克隆、校验、试跑、状态变更、批量操作中该功能域实际支持的全部动作。 |
| 权限与团队 | 完整接入 Nightingale role/operation、user group、busi group、ro/rw、team_ids，并映射到 FindX 团队/业务/值班权限。 |
| 审计与回滚 | 所有写操作必须有操作者、对象、前后差异、来源、审批状态、回滚引用和失败清理策略。 |
| 安全与脱敏 | token、password、secret、DSN、cookie、authorization、SSH key、内部 URL、上游 raw error 不进入前端、日志、AI prompt 和报告。 |
| 断连降级 | Nightingale、Categraf、findx-agents、AI、日志数据源任一不可用时，页面和工作流必须给出可读状态，不白屏、不误报成功。 |
| UI 完整性 | 使用 FindX 当前 UI 风格实现真实页面、筛选、表格、详情、弹窗、空状态、错误态、加载态和权限态，不以裸 iframe 代替主体验。 |
| 测试验收 | 覆盖正常路径、异常路径、权限路径、边界路径、断连路径、敏感信息扫描和 WSL/Linux 构建验证。 |
| 文档同步 | README、运维手册、模板说明、API 契约、回滚说明、许可证/NOTICE 同步更新。 |

“只读”只能作为安全策略的一种模式，例如诊断工具默认只读、AI 默认只给建议；不能把原本应有的写入生命周期省掉。危险写操作不是不做，而是必须通过权限、审批、审计、试跑、回滚和二次确认完整实现。

### 4.1 Nightingale 全功能域映射

| Nightingale 功能域 | 源码/API 证据 | FindX 集成形态 | 写入原则 |
| --- | --- | --- | --- |
| 身份认证、SSO、验证码、会话刷新 | `/auth/login`、`/auth/logout`、`/auth/refresh`、`/auth/sso-config`、`/auth/captcha`、`sso_config.go` | FindX 保持自有登录体验，P2 支持与 Nightingale 共用 SSO/OIDC/LDAP 身份源 | 不保存 Nightingale 明文密码 |
| 自身资料、API token、source token | `/self/profile`、`/self/password`、`/self/token`、`/source-token`、`user_token.go`、`source_token.go` | 个人设置页展示权限与 token 状态，敏感值只展示创建/失效状态 | token 只创建一次可见，日志脱敏 |
| 用户、角色、权限点 | `/users`、`/roles`、`/role/:id/ops`、`/operation`、`user.go`、`role.go`、`role_operation.go` | FindX 权限映射页，统一到“观察者/运维/管理员/模板管理员/告警管理员” | 不绕过 Nightingale `perm()` 校验 |
| 用户组/团队 | `/user-groups`、`/user-group/:id/members`、`/v1/n9e/user-group-members`、`user_group.go` | 映射为 FindX 团队、值班组、责任人矩阵 | 覆盖成员增删、同步、审计和权限验证 |
| 业务组/RBAC | `/busi-groups`、`/busi-group/:id/members`、`/busi-group/:id/perm/:perm`、`busi_group.go` | 映射为业务系统、拓扑业务、告警归属、模板安装范围 | 写操作必须校验 `ro/rw` |
| 目标资产、标签、元数据、心跳 | `/targets`、`/targets/stats`、`/targets/tags`、`/targets/bgids`、`/target/extra-meta`、`target.go`、`host_meta.go` | FindX 资产页、findx-agents 存活页、业务拓扑节点 | Nightingale target 为事实源 |
| 数据源与代理查询 | `/datasource/list`、`/datasource/plugin/list`、`/datasource/upsert`、`/datasource/query`、`/ds-query`、`proxy-query.md` | 数据源目录、查询能力封装、AI 证据查询 | 禁止任意 PromQL/SQL 裸透传，配置写入需审批和脱敏 |
| 时序指标、指标视图、内置 PromQL | `/metrics/desc`、`/metric-views`、`/builtin-metrics`、`/builtin-metric-promql`、`metric_view.go` | 指标字典、模板推荐、AI 解释、容量/SLO 证据 | 查询规则化，指标字典变更走审计 |
| Dashboard、看板、分享、标注 | `/boards`、`/board/:bid`、`/board/:bid/pure`、`/share-charts`、`/dashboard-annotations`、`board.go` | FindX 仪表盘页、业务健康页、变更标注、完整嵌入验证 | 看板写入走 Nightingale API |
| Embedded dashboards/products | `/embedded-dashboards`、`/embedded-product`、`embedded-product.md`、`embedded_product.go` | 双向入口：FindX 嵌入 Nightingale，看 Nightingale 时也可跳回 FindX AI 问诊 | 按 team_ids 控制可见性 |
| 告警规则生命周期 | `/alert-rules`、`/alert-rule/:arid`、`/import-prom-rule`、`/clone`、`/validate`、`/notify-tryrun`、`alert_rule.go` | FindX 告警规划、模板安装、品牌化编辑/跳转、覆盖率治理 | 完整写入生命周期，二次确认、审计、回滚 |
| 记录规则 | `/recording-rules`、`/recording-rule/:rrid`、`recording_rule.go` | SLO、容量、业务健康、模板安装依赖 | 不自建 recording 执行器 |
| 当前/历史告警事件 | `/alert-cur-events/list`、`/alert-his-events/list`、`/event-detail/:hash`、`alert_cur_event.go`、`alert_his_event.go` | 告警处置台、值班驾驶舱、AI 问诊入口、复盘时间线 | Nightingale event 为事实源 |
| 告警聚合、卡片、统计 | `/alert-cur-events/card`、`/alert-cur-events/stats`、`/alert-aggr-views`、`alert_aggr_view.go` | P0/P1 聚合、风暴视图、团队驾驶舱 | FindX 只做展示与排序 |
| 通知记录、评估详情、trace logs | `/event-notify-records/:eid`、`/alert-eval-detail/:id`、`/trace-logs/:traceid`、`notification_record.go` | 证据链、交接班、复盘、通知命中解释 | 脱敏展示，原始证据按权限查看 |
| 静默/变更窗口 | `/alert-mutes`、`/active-alert-mutes`、`/alert-mute-tryrun`、`alert_mute.go` | 变更窗口、静默建议、影响面检查 | 写入需审批、到期和审计 |
| 订阅策略 | `/alert-subscribes`、`/alert-subscribe/alert-subscribes-tryrun`、`alert_subscribe.go` | 团队订阅、个人订阅、值班触达偏好 | 不静默替用户订阅 |
| 通知模板、消息模板、渠道、规则 | `/notify-tpls`、`/message-templates`、`/notify-channel-configs`、`/notify-rules`、`notify_tpl.go`、`message_tpl.go` | FindX 通知中心、模板预览、升级策略、Hermes 文案生成 | 完整试跑、审批、写入、回滚、审计 |
| 事件流水线 | `/event-pipelines`、`/event-pipeline-tryrun`、`/trigger`、`/stream`、`/executions`、`event_pipeline.go` | 告警加工、降噪、AI 摘要输出、外部系统联动 | trigger/stream 必须限权限和速率 |
| 集成组件、payload、模板资产 | `/builtin-components`、`/builtin-payloads`、`/integrations/*`、`builtin_component.go`、`builtin_payload.go` | FindX 模板中心全量目录、安装向导、漂移检测 | 不复制为第二套模板事实源 |
| 日志查询、ES/OpenSearch、索引模式 | `/logs-query`、`/log-query-batch`、`/es-index-pattern`、`/os-*`、`es_index_pattern.go` | 日志检索入口、告警证据链、AI 日志摘要 | 字段级脱敏和查询配额 |
| TDengine/DB 查询 | `/tdengine-*`、`/db-*`、`/sql-template`、`router_datasource_db.go` | 数据源扩展入口和诊断证据查询 | 禁止任意 SQL 写操作 |
| 任务模板、任务记录、服务器集群 | `/task-tpls`、`/job-tasks`、`/servers`、`/server-clusters`、`task_tpl.go`、`task_record.go` | 运维任务目录、执行记录、任务权限、审批、审计 | 任意远程执行需结构化任务、审批和回滚，不开放裸 shell |
| 保存视图、收藏、用户变量、站点配置 | `/saved-views`、`/user-variable-configs`、`/config`、`/site-info`、`saved_view.go` | 用户工作台偏好、常用查询、团队视图模板 | 不写入敏感变量明文 |
| Webhook 与第三方入口 | `/webhooks`、`notify_rule.go`、`event_pipeline.go` | Hermes、ITSM、ChatOps、FindX AI 结果回流 | 签名、审计、重放保护 |
| Nightingale AI assistant/agents/skills/MCP | `/ai-agents`、`/ai-llm-configs`、`/ai-skills`、`/mcp-servers`、`/assistant/*`、`ai_agent.go` | 互操作与边界治理：FindX AI 做运维问诊，Nightingale AI 能力作为可选上游能力 | 避免双 AI 写入冲突 |
| Categraf/findx-agents 采集 | `/v1/n9e/heartbeat`、Categraf `writers`、`heartbeat`、`local/http provider` | findx-agents 安装、存活、远程配置、模板下发 | 不 fork Categraf inputs |

### 4.2 项目目标

1. FindX 成为 Nightingale 的完整运维门户和产品化外壳，用户后续不需要从 Nightingale 换平台。
2. Nightingale 成为监控事实源、规则事实源、模板事实源、通知事实源、事件事实源和权限事实源。
3. Categraf 成为 findx-agents 的采集底座，findx-agents 只做品牌、安装、配置、存活和远程配置体验。
4. FindX 前端继续使用现有 UI 风格，把 Nightingale 对象重组为值班驾驶舱、业务健康、告警处置、模板中心、团队权限、变更窗口、复盘和 AI 问诊。
5. Nightingale 的所有主要功能域都有 FindX 入口、权限映射、断连降级、审计要求和验收标准。
6. AI 问诊从 Nightingale 获取目标、告警、规则、指标、日志、Dashboard、通知记录、静默、订阅、事件流水线、Runbook、模板和团队上下文。
7. AI 只负责解释、归纳、推荐、复盘和知识沉淀；任何规则、通知、静默、订阅、事件流水线、任务写入都必须人工确认、权限校验和审计。
8. 旧 Catpaw 相关命名、API 和页面逐步迁移到 findx-agents，Catpaw 代码只作为历史参考。

### 4.3 非目标

本项目不做：

- 不替代 Nightingale，不把 FindX 做成第二套监控核心平台。
- 不把 Nightingale 数据搬到 FindX 后再由 FindX 作为新事实源。
- 不重写 Nightingale 的告警引擎、规则引擎、通知引擎、事件流水线引擎、Dashboard 引擎、日志查询引擎、任务执行引擎和权限引擎。
- 不重写 Categraf 采集插件，不 fork Categraf inputs。
- 不直接从 Catpaw 派生新探针。
- 不支持裸远程命令、裸 SQL、裸 PromQL、裸 HTTP 请求；所有动作必须被封装为结构化工具、模板或审批任务。
- 自动修复不作为默认能力；若后续立项，必须完整实现权限、审批、试跑、回滚、审计和失败保护。
- 不直接修改 Nightingale 前端源码作为主路线；优先用 Nightingale API、service API、pure board、embedded product 和受控反向代理。

## 5. 目标架构

```text
用户 / Hermes / Web
        |
        v
AI WorkBench / FindX 入口层
  - AI 问诊
  - findx-agents 存活视图
  - FindX 模板中心
  - Nightingale API 聚合与品牌化包装
  - Dashboard / embedded product 受控嵌入
  - 诊断报告与知识库
        |
        +----------------------------+--------------------------+
        |                            |                          |
        v                            v                          v
Nightingale                      AI 诊断引擎              Template Catalog
  - Targets                        - LLM                    - sync metadata
  - Busi Groups                    - evidence chain         - hash/version
  - Datasources                    - metric explain         - license/source
  - Alert Rules                    - runbook recommend      - install state
  - Events                         - template recommend
  - Dashboards
  - Metrics Desc
  - Builtin Payloads
  - Notify / Event Pipelines
        ^
        |
findx-agents
  - Categraf binary/config
  - Inspector sidecar/runtime
  - local inspect tools
  - remote structured diagnose sessions
  - native N9E heartbeat
  - remote write
  - local/http input provider
  - optional FindX alive heartbeat
```

## 6. 能力归属矩阵

| 能力 | 归属 | AI WorkBench 做什么 |
| --- | --- | --- |
| AI 问诊 | AI WorkBench | 保留主入口、诊断编排、报告和知识沉淀。 |
| 探针存活 | AI WorkBench + Nightingale | 自有 findx-agents 存活视图；同时读取 Nightingale target heartbeat。 |
| 主机/目标资产 | Nightingale | 只做品牌化列表、搜索、跳转、诊断入口。 |
| 指标采集 | Categraf | 生成 findx-agents 安装包和配置模板，不 fork inputs。 |
| 单机巡检 | findx-agents Inspector | 借鉴 Catpaw inspect 模式，提供本机 CPU/内存/磁盘/网络/进程/日志/证书/容器等结构化巡检。 |
| Agent 诊断工具 | findx-agents Inspector + FindX | Agent 执行结构化工具和采证，AI 推理在 FindX 平台侧完成；写类工具必须走审批任务。 |
| 远程诊断 session | FindX + findx-agents | 支持 inspect/diagnose/chat 三类受控 session；远程禁用裸 shell，开放结构化工具。 |
| 告警规则 | Nightingale | 提供模板安装、封装入口或跳转，不重写规则引擎。 |
| 活跃/历史告警 | Nightingale | 拉取事件作为 AI 问诊证据和前端摘要。 |
| 仪表盘 | Nightingale | 嵌入 `/board/:bid/pure`、使用 embedded dashboards 或创建品牌化跳转。 |
| 指标字典 | Nightingale | 同步 `/metrics/desc` 和 builtin metrics，为 AI 解释指标含义。 |
| 记录规则 | Nightingale | 只做模板展示、安装入口和 AI 推荐，不自建执行器。 |
| 通知渠道/模板 | Nightingale | 复用 notify tpl、message templates、notify rules、notify channels。 |
| 事件流水线 | Nightingale | 复用 event pipelines；AI 问诊可作为受控触发源或 webhook 输出。 |
| 模板中心 | FindX 编排 + Nightingale/Categraf 来源 | 做索引、筛选、预览、同步、安装编排、版本/hash/归属管理。 |
| 团队/用户组 | Nightingale + FindX 映射 | 可同步 Nightingale user-groups/team members，FindX 保留自己的登录和产品体验。 |
| 业务组/RBAC | Nightingale + FindX 映射 | 复用 Nightingale busi-groups、members、perm flag，FindX 做权限映射和运维视图。 |
| 值班/升级 | FindX + Nightingale 通知能力 | FindX 保留值班体验，复用 Nightingale notify rules/message templates/event pipelines。 |
| 变更窗口/静默 | Nightingale | 复用 alert-mutes、subscribes 和业务组权限，FindX 做变更视角入口和 AI 风险提示。 |
| 故障复盘/SLO | AI WorkBench | 以 Nightingale 事件、dashboard、通知记录、event pipeline 执行为证据生成复盘和服务健康报告。 |
| 日志查询/索引模式 | Nightingale | 复用 logs-query、log-query-batch、ES/OpenSearch index pattern，FindX 做证据链和脱敏摘要。 |
| 任务模板/任务记录 | Nightingale | 复用 task-tpls、job-tasks、servers、server-clusters，FindX 做目录、执行、审批、审计和结果归档。 |
| 保存视图/用户变量 | Nightingale | 复用 saved views、favorites、user variable configs，FindX 做用户工作台和团队视图。 |
| SSO/LDAP/OIDC | Nightingale + FindX | P2 统一身份源，不互相保存明文密码，不绕过 Nightingale 权限点。 |
| Nightingale AI assistant | Nightingale + FindX | 作为可选上游能力互操作；FindX AI 问诊保留主体验，避免双 AI 同时写规则。 |
| 远程任务/Ibex | Nightingale/Categraf | 纳入结构化任务治理；不开放裸脚本，必须有模板、审批、审计、回滚和失败保护。 |
| Catpaw 插件巡检 | findx-agents Inspector | 参考 Catpaw 能力做 clean-room 迁移或 AGPL 独立侧车；按功能域完整迁移本机、网络、容器、中间件巡检能力。 |

## 6.1 运维视角深度集成：团队、业务组、值班、变更和复盘

上一版重点解决了监控事实源、模板和采集底座；从真实运维团队日常使用看，还需要把“谁负责、谁能看、谁处理、怎么交接、怎么复盘”纳入 Nightingale 深度集成。

### 6.1.1 团队和权限能集成，但要做映射而不是混用账号

Nightingale 源码已经具备团队/RBAC 能力：

- `/users`、`/roles`、`/auth/perms`、`/self/perms`：用户、角色、权限。
- `/user-groups`、`/user-group/:id/members`、service API `/v1/n9e/user-group-members`：用户组和成员。
- `/busi-groups`、`/busi-group/:id/members`、`/busi-group/:id/perm/:perm`：业务组、成员和 `ro/rw` 权限。
- `models.User` 中有 `TeamsLst` 和 SSO/LDAP 同步相关字段，`models.UserGroupMemberSyncByUser` 支持按用户同步团队。
- `models.EventPipeline`、`models.EmbeddedProduct` 都支持 `team_ids`，说明 Nightingale 的事件流水线和嵌入产品也能按团队控制可见性。

FindX 侧建议做三层映射：

| Nightingale | FindX | 用途 |
| --- | --- | --- |
| user group / team | FindX team | 值班组、通知对象、权限范围、模板可见范围 |
| busi group | FindX business / topology business | 业务拓扑、告警归属、dashboard 范围、AI 问诊上下文 |
| role / operation perms | FindX permission profile | 只读、运维、管理员、模板安装、告警处理、事件流水线触发 |

不建议直接混用账号密码。推荐路线：

1. P0：完整同步 Nightingale user-groups、busi-groups、members、roles/perms 摘要，并建立可验证的 FindX 权限映射。
2. P1：建立 `findx_team_mapping` 和 `findx_business_mapping`，把 Nightingale team/busi group 与 FindX 值班组/业务拓扑绑定。
3. P1：FindX 页面按映射后的业务组权限过滤目标、告警、dashboard、模板和 AI 问诊入口。
4. P2：支持统一 SSO/LDAP/OIDC，FindX 和 Nightingale 使用同一个身份源，而不是互相保存明文密码。
5. P2：支持团队成员变更同步到值班组和通知规则，写操作必须具备审计、审批和回滚。

### 6.1.2 运维工作台应该围绕场景，而不是围绕上游对象

FindX UI 不应该照搬 Nightingale 菜单，也不应该让用户在 FindX 与 Nightingale 两套 UI 之间反复跳。建议按运维场景重组：

| FindX 场景 | 复用 Nightingale 能力 | FindX 增强 |
| --- | --- | --- |
| 值班驾驶舱 | active events、busi groups、notify rules | 当前责任人、P0/P1 聚合、AI 摘要、交接班 |
| 业务健康页 | targets、dashboards、record rules、busi group | 拓扑、SLO/错误预算、核心链路健康 |
| 告警处置台 | cur/his events、alert rules、mutes、subscribes | 一键 AI 问诊、证据链、处置建议、复盘草稿 |
| 模板中心 | integrations、builtin payloads、Categraf inputs | 品牌化预览、安装向导、依赖检查、漂移检测 |
| 团队与权限 | users、roles、user-groups、busi-groups | 映射视图、团队责任矩阵、值班组绑定 |
| 变更窗口 | alert mutes、dashboard annotations | 变更风险评估、静默建议、影响面说明 |
| 故障复盘 | events、notify records、event pipeline executions | 时间线、根因、改进项、知识库沉淀 |

### 6.1.3 UI 和外观坚持 FindX 当前风格

UI 原则：能力用 Nightingale，体验用 FindX。

当前 AI WorkBench 前端已有统一风格：

- `web/src/App.vue`：左侧导航、玻璃拟态面板、浅/深色主题、Element Plus 图标和按钮。
- `web/src/views/Dashboard.vue`：运维总览、统计卡片、快捷入口、告警分布。
- `web/src/views/OnCallConfig.vue`：值班组、通知渠道、升级策略、发送记录。
- `web/src/views/Workbench.vue`、`Diagnose.vue`、`Alerts.vue`、`Topology.vue`：智能诊断、告警处置、业务拓扑已经是 FindX 自有体验。

因此后续设计要求：

- 主导航、页面布局、颜色、卡片、按钮、表格、弹窗、空状态都沿用 FindX 当前风格。
- Nightingale dashboard 只在仪表盘细节页 iframe/代理嵌入；外层仍是 FindX 标题、筛选器、业务组选择和操作栏。
- Nightingale 表单不直接暴露给普通用户；复杂写操作优先跳转或管理员模式嵌入。
- 模板中心、团队映射、值班驾驶舱、故障复盘都用 FindX 组件重做展示，只调用 Nightingale API。
- 夜莺术语在 UI 中做产品化翻译：`busi group` 显示为“业务组/业务系统”，`user group` 显示为“团队”，`payload` 显示为“模板”。

### 6.1.4 运维闭环能力补充

除监控和模板外，建议补齐以下运维闭环：

1. 值班驾驶舱：按团队/业务组聚合 P0/P1、未认领、未恢复、升级中告警。
2. 责任矩阵：业务组 -> 团队 -> 值班组 -> 通知渠道 -> 升级策略。
3. 交接班摘要：过去 8/12/24 小时新增/恢复/静默/升级/AI 问诊/待跟进事项。
4. 变更窗口：从 Nightingale alert mute 读取静默窗口，AI 提示是否覆盖相关业务组和告警规则。
5. SLO/错误预算：基于 record rules、dashboard 指标和告警事件生成业务健康分。
6. 故障时间线：事件触发、通知发送、认领、静默、AI 问诊、pipeline 执行、恢复。
7. 复盘报告：AI 根据 Nightingale 事件和 FindX 诊断记录生成 Markdown 复盘，沉淀到知识库。
8. 模板运营：统计哪些团队安装了哪些模板、哪些模板漂移、哪些业务缺少 dashboard/alert/collect。

## 7. Nightingale 告警规划全量集成

用户明确要求“他们的告警规划，还有很多，所有功能都集成进来”。因此告警不应只做“告警列表 + AI 问诊”，而要把 Nightingale 从规则规划、规则生命周期、评估证据、通知升级、静默订阅、聚合视图到复盘治理的完整闭环纳入 FindX。

### 7.1 告警规则生命周期

| 生命周期 | Nightingale 能力 | FindX 入口 | AI 增强 |
| --- | --- | --- | --- |
| 模板发现 | builtin alert payloads、integrations alerts | 模板中心、业务缺口扫描 | 推荐缺失规则和推荐原因 |
| 新建规则 | `/busi-group/:id/alert-rules` | 品牌化向导或管理员嵌入页 | 生成待审批规则方案，审批后写入 |
| 导入规则 | `/alert-rules/import`、`/import-prom-rule` | Prometheus rule 导入、模板批量导入 | 解释迁移差异和风险 |
| 克隆规则 | `/alert-rules/clone`、`/alert-rules/clones` | 跨业务组复制、环境复制 | 检查标签/数据源/阈值适配 |
| 编辑规则 | `/alert-rule/:arid`、`/fields` | 规则详情、批量字段变更 | 给出阈值建议和影响面 |
| 校验规则 | `/busi-group/alert-rule/validate` | 提交前校验、CI 校验 | 解释校验失败原因 |
| 通知试跑 | `/alert-rules/notify-tryrun` | 通知预览、值班验证 | 检查通知噪声和遗漏 |
| 启用试跑 | `/alert-rules/enable-tryrun` | 灰度启用、启用前风险检查 | 给出启用建议 |
| 删除/禁用 | DELETE/PUT fields | 规则下线流程、回滚点 | 提醒依赖和覆盖缺口 |
| 版本/漂移 | FindX template hash + Nightingale object id | 规则版本、模板漂移、安装记录 | 解释上游模板变更 |

原则：规则的事实源始终是 Nightingale。FindX 可以做品牌化表单、导入向导、差异预览、审批和审计，但不能维护第二套规则状态。

### 7.2 告警分类与覆盖规划

FindX 告警规划默认按运维视角组织，不按源码对象堆菜单：

| 分类 | 覆盖对象 | 来源 |
| --- | --- | --- |
| 基础设施 | CPU、内存、磁盘、网络、文件系统、进程、systemd | Categraf input + Nightingale Linux 模板 |
| Kubernetes | Node、Pod、Deployment、容器、调度、资源、APIServer | Nightingale integrations + Categraf/K8s 采集 |
| 数据库 | MySQL、PostgreSQL、Redis、MongoDB、Oracle、ClickHouse | Nightingale integrations alerts/dashboards |
| 中间件 | Kafka、Elasticsearch、Nginx、RabbitMQ、Consul、ZooKeeper | integrations + builtin metrics |
| 黑盒拨测 | HTTP_Response、Net_Response、Ping、端口连通性 | Categraf blackbox/inputs |
| 业务/SLO | 请求成功率、延迟、吞吐、错误预算、核心链路 | recording rules + dashboards + FindX 拓扑 |
| 容量与趋势 | 磁盘耗尽、连接数、队列堆积、存储增长 | recording rules + AI 趋势分析 |
| 安全与合规 | 证书、端口、登录异常、配置漂移 | FindX 工作流 + Nightingale 事件证据 |
| 异常检测 | 突变、周期性偏离、同比环比异常 | FindX AI/时序分析推荐，Nightingale 规则承接 |

每个业务组需要一张“告警覆盖矩阵”：目标数、已安装模板、核心规则数、静默策略、订阅策略、通知规则、Dashboard、SLO 记录规则、最近 30 天告警噪声和漏报风险。

### 7.3 告警严重级别与值班策略

FindX 显示 P0/P1/P2/P3，但底层映射 Nightingale severity、labels、busi group 和 notify rule：

| FindX 等级 | 运维含义 | Nightingale 映射 | 默认动作 |
| --- | --- | --- | --- |
| P0 | 核心业务不可用、数据风险、全局故障 | 高 severity + 核心业务标签 + 关键规则 | 立即通知值班、升级负责人、生成 AI 摘要 |
| P1 | 单业务核心能力下降、关键组件故障 | 高/中 severity + 业务组规则 | 通知团队、15 分钟未处理升级 |
| P2 | 局部异常、容量风险、非核心链路 | 中/低 severity + 非核心规则 | 工作时间通知或订阅推送 |
| P3 | 提醒、趋势、低风险噪声 | 低 severity 或建议类规则 | 汇总日报/周报，不打扰值班 |

告警等级不是由 AI 临时猜测。AI 可以建议调整等级，但最终以 Nightingale 规则字段、标签治理和团队审批为准。

### 7.4 通知、静默、订阅和升级

Nightingale 已经具备完整触达能力，FindX 要全部纳入：

- 通知模板：`notify-tpls`、`message-templates`，用于告警正文、恢复正文、AI 摘要模板。
- 通知渠道：`notify-channel`、`notify-channel-configs`、FlashDuty、飞书、钉钉、PagerDuty 等上游渠道配置。
- 通知规则：`notify-rules`、`notify-rule/test`、`event-pipelines-tryrun`，用于路由、升级和试跑。
- 静默策略：`alert-mutes`、`active-alert-mutes`、`alert-mute-tryrun`，用于变更窗口、维护窗口、临时降噪。
- 订阅策略：`alert-subscribes`、`alert-subscribes-tryrun`，用于团队/个人订阅和业务关注。
- 通知记录：`event-notify-records/:eid`，用于证明“有没有通知、通知给谁、什么时候通知、是否被加工”。

FindX 的增强不是再造通知系统，而是做“运维能读懂的规划视图”：谁负责、谁订阅、谁升级、哪些规则会被静默、哪些事件会被 pipeline 加工、哪些 P0/P1 没有有效通知。

### 7.5 告警事件处置台

FindX 告警处置台必须覆盖以下 Nightingale 事件能力：

| 能力 | Nightingale 来源 | FindX 展示 |
| --- | --- | --- |
| 活跃事件 | `/alert-cur-events/list` | 告警列表、P0/P1 队列、未认领、处理中 |
| 事件卡片 | `/alert-cur-events/card`、`/card/details` | 按规则、业务组、团队、目标聚合 |
| 历史事件 | `/alert-his-events/list` | 历史查询、复盘证据、趋势统计 |
| 事件详情 | `/event-detail/:hash` | 指标、标签、规则、目标、恢复信息 |
| 事件统计 | `/alert-cur-events/stats` | 值班驾驶舱、业务健康分 |
| 聚合视图 | `/alert-aggr-views` | 风暴、重复告警、根因候选 |
| 通知记录 | `/event-notify-records/:eid` | 触达证据和升级链路 |
| 评估详情 | `/alert-eval-detail/:id` | 规则为什么触发、查询结果是什么 |
| Trace logs | `/trace-logs/:traceid` | 事件流水线、通知、处理链路排障 |

AI 问诊入口放在事件详情和聚合视图中，输入必须包含 event、rule、target、datasource、query、dashboard、notify record、mute/subscribe 命中情况和模板说明。

### 7.6 告警治理看板

FindX 需要新增“告警规划/治理”视图，用来长期替代人工 Excel：

1. 覆盖率：每个业务组是否有基础设施、K8s、数据库、中间件、黑盒、SLO、容量类规则。
2. 噪声率：近 7/30 天重复告警、短恢复告警、被静默告警、无人处理告警。
3. 漏报风险：有 targets/datasources/dashboard 但缺少 alert payload 或核心规则的对象。
4. 通知有效性：P0/P1 是否有 notify rule、message template、订阅人、升级链路和试跑记录。
5. 规则健康：禁用规则、查询失败规则、长期无数据规则、模板漂移规则。
6. 变更影响：当前 active mutes 覆盖哪些业务、哪些 P0/P1 被静默。
7. AI 建议：输出补规则、调阈值、降噪、补订阅、补模板方案；涉及写入的动作进入审批队列，审批后执行。

### 7.7 告警写操作门禁

所有告警相关写操作都要比普通配置更严格：

- P0：完成规则审计、详情、校验、试跑、导入预览、差异预览和审批流设计，写入链路必须可真实执行但默认需要管理员确认。
- P1：允许管理员通过 FindX 写入模板安装、规则创建、规则编辑、静默、订阅、通知模板，但必须二次确认、权限校验、审计和回滚引用。
- P2：允许 AI 生成规则方案、通知模板方案、静默建议和 pipeline 方案，全部进入待审批队列，审批后执行。
- 禁止 AI 静默创建 P0/P1 规则、删除规则、创建永久静默、改通知接收人、触发外部 pipeline。

## 8. FindX 模板中心：深度模板集成方案

用户要求“模板那些全都要”，因此 FindX 模板中心需要覆盖 Nightingale 与 Categraf 的全量成熟模板能力，但不复制建设执行引擎。

### 8.1 模板来源

1. Nightingale integrations 文件源：
   - `alerts/*.json`
   - `dashboards/*.json`
   - `metrics/*.json`
   - `record-rules/*.json`
   - `markdown/*.md`
   - `icon/*`
   - `*.toml`
2. Nightingale builtin API 源：
   - `/api/n9e/builtin-components`
   - `/api/n9e/builtin-payloads`
   - `/api/n9e/builtin-payloads/cates`
   - `/api/n9e/builtin-metrics`
   - `/api/n9e/builtin-metric-promql`
   - `/api/n9e/metrics/desc`
3. Nightingale notification/event 源：
   - `/api/n9e/notify-tpls`
   - `/api/n9e/message-templates`
   - `/api/n9e/notify-channel-configs`
   - `/api/n9e/notify-rules`
   - `/api/n9e/event-pipelines`
4. Nightingale dashboard/embed 源：
   - `/api/n9e/boards`
   - `/api/n9e/board/:bid`
   - `/api/n9e/board/:bid/pure`
   - `/api/n9e/embedded-dashboards`
   - `/api/n9e/embedded-product`
5. Categraf 采集模板源：
   - `conf/config.toml`
   - `conf/input.*/*.toml`
   - `inputs/local_provider.go`
   - `inputs/http_provider.go`

### 8.2 模板类型

| 类型 | 上游来源 | FindX 展示/使用方式 | 是否自建执行 |
| --- | --- | --- | --- |
| integration component | Nightingale integrations / builtin-components | 模板目录、图标、说明、组件筛选 | 否 |
| dashboard payload | Nightingale dashboards / builtin-payloads | 预览、安装、嵌入、跳转 | 否 |
| alert payload | Nightingale alerts / builtin-payloads | 预览、安装、AI 推荐 | 否 |
| collect payload | Nightingale collect payload / Categraf toml | 生成 findx-agents input 配置 | 否 |
| builtin metrics | Nightingale metrics JSON / builtin-metrics | 指标字典、PromQL 推荐、诊断解释 | 否 |
| metrics desc | Nightingale `etc/metrics.yaml` / `/metrics/desc` | AI 报告中的指标含义、单位、阈值解释 | 否 |
| record rules | Nightingale record-rules | 预览、安装、依赖提示 | 否 |
| markdown/runbook | Nightingale markdown | 模板说明、安装说明、AI 引用材料 | 否 |
| icon/image | Nightingale icon/markdown assets | 模板市场展示 | 否 |
| notify template | Nightingale notify-tpls/message-templates | 通知内容预览和复用 | 否 |
| notify channel/rule | Nightingale notify channel/rule | 入口、同步摘要、试跑、审批写入和回滚 | 否 |
| event pipeline | Nightingale event-pipelines | 触发、试跑、执行记录读取、AI 回调 | 否 |
| embedded product | Nightingale embedded-product | 双向入口嵌入 | 否 |

### 8.3 模板目录模型

FindX 本地保存索引、同步状态、安装审计和回滚引用，不保存第二套事实源。建议新增缓存型元数据结构：

```text
findx_template_catalog
  id
  source                 # nightingale_file / nightingale_api / categraf_file / findx_overlay
  source_path_or_url
  upstream_type          # dashboard / alert / collect / metric / record_rule / markdown / notify / event_pipeline
  component_ident
  component_name
  cate
  payload_type
  name
  tags
  description
  content_hash
  upstream_uuid
  upstream_version
  upstream_license
  installed_status       # not_installed / installed / drifted / conflict / unsupported
  installed_ref          # Nightingale object id or ident
  dependencies           # datasource, labels, collector, Categraf input
  compatibility          # Nightingale version, Categraf version, OS/arch
  last_sync_at
  last_install_at
```

约束：

- `content_hash` 是幂等和漂移判断的核心，不靠名称猜测。
- `upstream_license` 必须保留 Nightingale/Categraf 归属。
- `installed_status` 只描述安装状态，不成为规则、dashboard、通知的事实源。
- 写入 Nightingale 必须走 Nightingale API，不直接写 MySQL/Redis。
- 模板内容中的凭据、URL、token、DSN 必须用占位符或变量，不进入日志和前端 raw error。

### 8.4 同步流程

1. Scan：扫描 Nightingale integrations、Categraf conf/input、Nightingale API。
2. Normalize：统一为 component、payload、asset、dependency 四类元数据。
3. Validate：校验 JSON/TOML、dashboard schema、alert rule schema、collect toml、event pipeline schema。
4. Hash：按规范化内容生成 hash，忽略创建时间、更新人等非语义字段。
5. Store：写入 FindX 模板索引缓存。
6. Diff：对比 Nightingale 已安装对象，标记 installed/drifted/conflict。
7. Preview：展示 dashboard、alert、metrics、record rule、collect toml、markdown。
8. Install：通过 Nightingale API 安装或更新。
9. Rollback：保留安装前对象引用和 hash，支持回滚到上一版本或删除本次安装对象。

### 8.5 安装策略

- dashboard：优先走 Nightingale board/builtin payload 能力安装，再用 `/board/:bid/pure` 或 embedded dashboards 嵌入。
- alert：走 Nightingale alert rule/builtin payload 能力安装，保留业务组、数据源、标签变量确认步骤。
- collect：生成 findx-agents 的 Categraf `input.*` 配置，由本地 provider 或 HTTP provider 下发。
- metric desc：同步为指标字典，用于 AI 解释；如需覆盖夜莺默认字典，必须走差异预览、审批和回滚。
- record rule：走 Nightingale 规则能力安装，必须校验 datasource 与 PromQL。
- notify/message template：走 Nightingale message templates / notify templates API，支持预览、试跑、审批安装和回滚。
- event pipeline：走 Nightingale event pipeline API，支持详情、试跑、执行记录、审批创建、审批更新和回滚。
- embedded product：用于把 FindX AI 问诊入口注册到 Nightingale，也可把 Nightingale dashboard 作为 FindX 嵌入入口。

## 9. Categraf、Catpaw 与 findx-agents 深度融合方案

### 9.1 融合结论

findx-agents 不应只是“Categraf 改名”，也不应把 Catpaw 源码直接塞进 Categraf 主进程。推荐定位是：

```text
findx-agents
  = Categraf metrics/logs/heartbeat 采集核心
  + findx-inspector 单机巡检/结构化诊断侧车
  + FindX Agent Control Protocol
  + FindX 工作流与 AI 问诊联动
```

这样 Categraf 继续做成熟采集，Catpaw 的“单台机巡检、结构化工具、告警触发诊断、远程 session、安全模型”被吸收为 findx-inspector 能力层，FindX 平台负责任务编排、AI 推理、报告归档和权限审计。

### 9.2 合规路线：默认 clean-room，保留 AGPL 可选侧车

Catpaw 本地 LICENSE 为 AGPL-3.0，因此不建议把 Catpaw 源码直接合入 Categraf 或 FindX 闭源交付主线。建议两条路线并列：

| 路线 | 做法 | 优点 | 风险 |
| --- | --- | --- | --- |
| 默认路线：clean-room 能力复刻 | 只参考 Catpaw 的公开设计、事件模型和能力边界，重新实现 findx-inspector 的接口、协议和插件 | 许可证风险低，能深度融入 FindX/Categraf | 初期开发量更大 |
| 可选路线：AGPL 独立侧车 | Catpaw 作为独立进程/容器运行，保留许可证和源码公开要求，findx-agents 只做进程管理和协议桥接 | 快速复用已有功能 | AGPL 合规、交付说明、源码公开和维护复杂 |

工程默认采用第一条。第二条只在明确接受 AGPL 义务的部署形态中作为合规隔离模式，并必须提供源码归属、许可证展示、部署边界和升级策略。

### 9.3 Categraf 保持稳定采集核心

findx-agents 对 Categraf 的包装原则：

- 不 fork Categraf inputs。
- 不改 Categraf remote write 协议。
- 不改 Categraf N9E heartbeat 语义。
- 不改 Categraf input TOML 主结构。
- 只包装安装目录、服务名、默认配置、模板生成、升级卸载、存活状态和 FindX 品牌入口。

### 9.4 配置模型

建议路径：

- 安装路径：`/opt/findx-agents`
- 配置路径：`/etc/findx-agents`
- 采集模板：`/etc/findx-agents/input.*/*.toml`
- 日志路径：`/var/log/findx-agents`
- systemd：`findx-agents.service`

默认配置：

- `[global] providers = ["local", "http"]`，本地配置和远程配置都纳入正式能力；部署时可按安全策略启停。
- `[global.labels]` 注入 `env`、`region`、`tenant`、`business`、`agent_id` 等非敏感标签。
- `[[writers]]` 指向 Nightingale remote write 入口。
- `[heartbeat]` 指向 Nightingale `/v1/n9e/heartbeat`。
- FindX alive heartbeat 上报平台存活和 Agent 能力状态所需字段，字段清单固定、脱敏、可版本化。

P1 可开启 Categraf `http_provider`：

- FindX 提供 `/api/v1/findx/agents/configs` 远程配置接口。
- Categraf 按 version/hash 拉取配置。
- 接口按 agent hostname、global labels、agent_id 返回差异化 input 配置。
- 远程配置接口必须认证、限流、审计、脱敏，并禁止返回任意命令。

### 9.5 模板分层

| 层级 | 示例 | 用途 |
| --- | --- | --- |
| base profile | linux_base | CPU、内存、磁盘、网络、进程、systemd、self_metrics |
| middleware profile | mysql、redis、nginx、kafka、clickhouse | 中间件采集模板 |
| kubernetes profile | kubernetes、cadvisor、prometheus | K8s 与容器采集模板 |
| blackbox profile | ping、http_response、net_response、dns_query、x509_cert | 黑盒探测与证书探测 |
| hardware profile | ipmi、smart、snmp、nvidia_smi、dcgm | 硬件/GPU/网络设备 |
| custom profile | prometheus、exec、procstat、filecount | 受控扩展，默认需要管理员确认 |

### 9.6 与 Nightingale integrations 对齐

FindX 模板中心需要把 Categraf input 模板与 Nightingale integration 关联：

- `Linux` integration 关联 `input.cpu`、`input.mem`、`input.disk`、`input.diskio`、`input.net`、`input.netstat`、`input.processes`、`input.systemd`。
- `MySQL` integration 关联 `input.mysql`，并关联 MySQL dashboards/alerts/metrics。
- `Redis` integration 关联 `input.redis` 与 Redis dashboards/alerts。
- `Nginx` integration 关联 `input.nginx` 与 Nginx dashboards/metrics。
- `Kubernetes` integration 关联 `input.kubernetes`、`input.cadvisor`、`input.prometheus` 与 record rules。
- `HTTP_Response`、`Net_Response`、`Ping` 分别关联黑盒拨测 input 与对应告警/dashboard。

这样用户在 FindX 里选择“安装 MySQL 监控模板”时，实际动作是：

1. 生成 findx-agents 的 `input.mysql/mysql.toml`。
2. 安装 Nightingale MySQL dashboard。
3. 安装 Nightingale MySQL alert rules。
4. 同步 MySQL metrics desc/builtin metrics。
5. 在诊断报告里把 MySQL runbook/markdown 作为引用材料。

### 9.7 findx-inspector：Catpaw 风格单机巡检能力

findx-inspector 是 findx-agents 内的巡检与诊断执行层，能力参考 Catpaw，但接口重新设计：

| 能力 | Catpaw 证据 | findx-inspector 设计 |
| --- | --- | --- |
| 单机 inspect | `RunInspect(pluginName,target)`、`ModeInspect` | `InspectTask{scope,profile,target,timeout}`，本机巡检无需 AI key |
| 插件事件 | `types.Event`、`labels/attrs/description` | `InspectionFinding` 标准结果，可转为 Nightingale event 或 FindX report |
| 诊断工具注册 | `Diagnosable`、`DiagnoseTool`、`ToolScopeLocal/Remote` | `ToolSpec` + `ToolRunner`，只读、参数 schema、输出上限 |
| 本机工具 | CPU、mem、disk、netif、procnum、systemd、journal 等 | 当前交付切片完整覆盖 Linux 本机基础工具 |
| 远端工具 | Redis、etcd、HTTP、证书等 remote tools | P1 从 Nightingale target/datasource/template 获取凭据引用，不在 Agent 明文保存 |
| 远程 session | `inspect/diagnose/chat` | FindX 下发 session，Agent 回传 stream events |
| 安全模型 | 远程禁用 shell、结构化工具、审计 | 直接作为 findx-agents 安全基线 |

P0 巡检插件清单：

| profile | 检查项 | 数据来源 |
| --- | --- | --- |
| `linux_quick` | uptime、load、CPU top、内存、磁盘使用率、inode、磁盘 IO、网络错误、监听端口、僵尸进程 | `/proc`、`/sys`、只读系统命令 |
| `linux_deep` | dmesg、journal、conntrack、filefd、procfd、sysctl、systemd failed units、mount 异常 | 结构化只读工具 |
| `network_basic` | DNS、ping、TCP connect、HTTP probe、证书有效期 | Catpaw http/dns/ping/cert 思路 + Categraf blackbox 配置 |
| `container_basic` | docker ps、容器重启、镜像/日志摘要 | Docker 只读 API |
| `middleware_hint` | 根据 Nightingale/Categraf 模板识别是否应启用 MySQL/Redis/Nginx 巡检 | 模板中心 + targets labels |

### 9.8 不把巡检做成另一套监控

巡检与采集的边界：

- Categraf 持续采集指标，写入 Nightingale/Prometheus remote write。
- findx-inspector 按需巡检或定时低频巡检，产出结构化 finding/report。
- Nightingale 负责规则告警和事件事实源。
- FindX 负责巡检任务、AI 问诊、报告、知识库和工作流。

巡检结果默认不替代 Nightingale alert rule。需要转成告警时，优先通过 Nightingale event pipeline、webhook 或 FindX 告警接入，再由 Nightingale/FindX 统一治理。

### 9.9 Agent Control Protocol

建议新增 FindX Agent Control Protocol，借鉴 Catpaw WebSocket 协议但重新定义：

| 方向 | message type | 用途 |
| --- | --- | --- |
| Agent -> FindX | `register` | 上报 agent_id、公钥、hostname、ip、version、capabilities、plugins |
| Agent -> FindX | `heartbeat` | 上报存活、Categraf 状态、Inspector 状态、active sessions |
| Agent -> FindX | `inspection_result` | 上报巡检结果、finding、报告摘要 |
| Agent -> FindX | `session_output` | 流式回传工具执行、证据、AI 过程阶段 |
| Agent -> FindX | `audit_log` | 回传只读工具调用审计摘要 |
| FindX -> Agent | `config_update` | 下发 Categraf input profile 和 inspector profile |
| FindX -> Agent | `session_start` | 发起 inspect/diagnose/chat |
| FindX -> Agent | `session_cancel` | 取消长任务 |
| FindX -> Agent | `ping` | 控制面健康检查 |

认证要求：

- P0：注册 token + agent_id + HMAC 签名，不记录真实 token。
- P1：Per-Agent Key，Agent 生成 Ed25519 key，平台审核后接受。
- P2：mTLS、短期证书、token 轮转、session nonce 防重放。

### 9.10 与 Categraf 的集成形态

三种集成形态按优先级推进：

| 形态 | 描述 | 适用阶段 |
| --- | --- | --- |
| 同包双进程 | `findx-agents` 安装包内包含 `categraf` 和 `findx-inspector`，systemd 管理一个 supervisor 或两个 service | P0/P1 推荐 |
| Categraf sidecar input | 借鉴 `inputs/huatuo`，由 Categraf 管理 inspector 进程并抓 inspector metrics | P1 正式集成形态 |
| 单进程模块化 | 深度改造 Categraf AgentModule，新增 InspectorAgent | P2 架构增强形态 |

推荐先做同包双进程：维护边界清楚、Categraf 升级简单、Inspector 崩溃不影响指标采集。

### 9.11 AI 问诊与新 Agent 联动

FindX AI 问诊应从“读 Nightingale 证据”升级为“按需让 Agent 补证据”：

```text
Nightingale 告警 / 用户发起巡检
        |
        v
FindX Workflow
  1. 拉取 Nightingale event/rule/target/dashboard/log evidence
  2. 判断是否需要 Agent 补充证据
  3. 下发 findx-inspector inspect session
  4. 收集 InspectionFinding 和工具输出
  5. AI 生成诊断报告
  6. 写入 FindX 诊断记录/知识库
  7. 可选回写 Nightingale event note / pipeline / Hermes
```

AI 与 Agent 的职责边界：

| 职责 | FindX AI | findx-agents |
| --- | --- | --- |
| 选择工作流 | 是 | 否 |
| 读取 Nightingale 事件/规则/Dashboard | 是 | 否 |
| 执行本机只读工具 | 否 | 是 |
| 执行远端只读诊断工具 | 通过平台授权 | 是 |
| 执行 shell | 默认否 | 远程禁止，本地人工模式后续单独设计 |
| 生成根因报告 | 是 | 否 |
| 写规则/静默/通知/任务 | 生成建议 | 不执行 |
| 审计 | 平台审计 + Agent 本地审计 | 工具调用审计 |

### 9.12 工作流联动设计

建议新增或改造以下工作流：

| 工作流 | 触发 | Agent 动作 | Nightingale 动作 | 输出 |
| --- | --- | --- | --- | --- |
| `agent_quick_inspection` | 用户点“单机巡检” | `linux_quick` inspect | 读取 target/labels/dashboard | 巡检报告、风险项 |
| `alert_agent_diagnosis` | Nightingale P0/P1 告警 | 按规则类型补 CPU/磁盘/网络/日志证据 | 读取 event/rule/eval/trace/notify | 根因报告、处置建议 |
| `business_health_inspection` | 值班交接/定时 | 多 Agent 并行 quick inspect | 读取业务组 events/dashboards/SLO | 业务健康分、交接摘要 |
| `template_install_validate` | 安装模板后 | 检查 Categraf input 是否运行、端口/凭据是否通 | 读取 targets/datasource/rules | 安装验证报告 |
| `change_window_check` | 创建静默/变更前 | 检查目标当前健康和高风险项 | 读取 active mutes/events | 变更风险说明 |
| `incident_postmortem_agent` | 故障恢复后 | 拉取关键时间窗系统证据 | 读取 history events/notify/pipeline | 复盘草稿 |
| `log_drilldown_inspection` | 日志告警/用户询问 | journal/log tail/grep 结构化工具 | 读取 logs-query/index pattern | 日志摘要和证据 |
| `capacity_inspection` | 容量告警/定时 | 磁盘/连接/进程/IO 补证据 | 读取 record rules/dashboard | 容量趋势和扩容建议 |

工作流执行约束：

- 所有 Agent session 必须有 `workflow_run_id`、`operator`、`business_group`、`target_ident`、`purpose`。
- 并发限制按业务组、用户、Agent 三层控制。
- 工具输出必须截断、脱敏、结构化保存。
- Agent 断连时工作流降级为 Nightingale-only 诊断，不失败整个问诊。

### 9.13 巡检结果标准模型

FindX 本地建议新增标准结果，不直接保存 Catpaw 原始 event：

```text
InspectionRun
  id
  workflow_run_id
  source = findx_agents
  agent_id
  target_ident
  profile
  status              # running/success/failed/timeout/cancelled
  started_at
  finished_at
  summary
  report_ref

InspectionFinding
  run_id
  severity            # ok/info/warning/critical
  category            # cpu/mem/disk/net/process/log/cert/container/middleware
  check
  target
  title
  evidence
  attrs
  recommendation
  related_n9e_event_id
  related_rule_id
```

这套模型可同时承接 Catpaw 风格事件、Categraf 自身状态、FindX 工作流结果和 AI 报告。

### 9.14 安全底线

findx-agents 巡检能力必须默认安全：

- 远程 session 禁用 shell。
- 默认开放结构化只读工具；写类工具必须走审批任务，不开放裸 shell。
- 工具参数必须有 schema。
- 工具输出默认 32KB 上限，日志类工具可按配置提高但必须审计。
- 敏感字段过滤：password、token、secret、key、dsn、authorization、cookie。
- 所有 session 写本地 JSONL 审计，并上报平台审计摘要。
- Agent 以低权限用户运行，需要能力时用 Linux capabilities，不默认 root。
- 任何写操作、修复操作、重启服务、kill 进程、改配置都不进 P0/P1。

## 10. Nightingale Dashboard 与完整嵌入方案

### 10.1 嵌入路径

优先级从高到低：

1. `/board/:bid/pure`：用于 FindX 前端 iframe 或后端代理嵌入，是第一轮正式交付入口。
2. `/embedded-dashboards`：用于受控配置已嵌入 dashboard。
3. `/embedded-product`：用于把 FindX AI 问诊入口注册到 Nightingale，或把 Nightingale 作为 FindX 的外部入口。
4. 直接 board API 渲染：作为深度品牌化方案，需完整实现组件渲染、权限、变量、时间范围和数据源兼容后再启用。

### 10.2 P0/P1 完整验收

首个正式验收切片选择 Linux integration：

- 导入或定位 `Linux/dashboards/categraf-overview.json`。
- 通过 Nightingale board 能力生成 dashboard。
- FindX 页面展示 dashboard iframe 或安全代理链接。
- Dashboard 中 target、datasource、业务组权限正确。
- Nightingale 断连时 FindX 显示 503/不可用状态，不白屏，不泄露内部 URL、token、DSN。
- 不修改 Nightingale 前端源码。

## 11. AI 问诊、工作流与新 Agent 联动

AI 问诊不能只读告警事件，还要把 Nightingale 模板体系和 findx-agents 巡检体系纳入证据链。

### 11.1 证据链输入

从告警触发 AI 问诊时，NightingaleEvidenceProvider 应收集：

- alert event：告警名、级别、触发值、触发时间、tags、target ident。
- alert rule：规则名称、PromQL、阈值、持续时间、通知规则、规则备注。
- target：主机、业务组、labels、agent heartbeat、Categraf 版本。
- datasource：数据源状态、查询入口、类型。
- dashboard：关联 component 的 dashboard 链接。
- metric desc：指标中文/英文描述、单位、类型、推荐 PromQL。
- record rules：相关预聚合规则。
- integration markdown：上游模板说明、安装说明、排查建议。
- notify context：通知模板、通知规则、通知渠道摘要。
- event pipeline：已触发或可触发的流水线、最近执行状态。
- agent heartbeat：Categraf heartbeat、FindX alive heartbeat、Inspector 状态、active sessions。
- inspection evidence：单机巡检 finding、工具输出摘要、执行耗时、错误、审计引用。
- workflow context：工作流 run id、触发人、业务组、目标、前置步骤、降级路径。

### 11.2 报告输出

AI 诊断报告建议固定结构：

1. 结论摘要：根因、置信度、影响范围。
2. 告警证据：事件、规则、PromQL、触发值。
3. 指标解释：引用 Nightingale metrics desc/builtin metrics。
4. 图表入口：关联 dashboard/pure board 链接。
5. 运行手册：引用 integration markdown/runbook。
6. 建议动作：默认给出建议；涉及写操作时进入审批任务，审批后执行并记录审计。
7. 验证步骤：推荐 PromQL、dashboard 面板、恢复判定。
8. Agent 证据：列出 findx-inspector 执行过哪些只读工具、哪些失败、哪些结果被引用。
9. 模板建议：建议安装的 dashboard、alert、collect、record-rule、inspection profile 模板。

### 11.3 与工作流的联动方式

AI WorkBench 现有工作流体系可以把 findx-agents 当作一个受控工具节点，而不是把 Agent 变成另一个自主决策者。

建议新增节点/工具：

| 节点/工具 | 输入 | 输出 | 约束 |
| --- | --- | --- | --- |
| `nightingale_evidence` | event id、target ident、busi group | event/rule/dashboard/log/notify evidence | 读取证据并校验权限 |
| `agent_inspect` | agent id、profile、timeout、purpose | InspectionRun、findings | 远程禁用 shell |
| `agent_tool_call` | tool name、schema args | tool result summary | P0 仅平台内置工具 |
| `agent_session_stream` | session id | 流式阶段事件 | 支持取消和超时 |
| `inspection_report_save` | run + findings + AI report | 诊断记录/知识库引用 | 脱敏保存 |
| `n9e_feedback` | report summary、event id | event note/pipeline/webhook | 预览、审批、写入、回滚引用 |

典型链路：

```text
Nightingale P1 告警
  -> nightingale_evidence
  -> 判断告警类型和目标在线状态
  -> agent_inspect(linux_quick 或 middleware_hint)
  -> AI 根因分析
  -> inspection_report_save
  -> Hermes/通知摘要
  -> 可选 n9e_feedback
```

降级原则：

- Agent 离线：继续使用 Nightingale-only 证据链，报告标记“未取得主机补充证据”。
- Nightingale 断连：允许单机巡检，但不能生成“告警事实一致”的结论。
- AI 不可用：保留巡检原始报告和规则化摘要，不影响告警/巡检任务完成。

### 11.4 安全边界

- AI 只能推荐模板安装，不能默认直接写 Nightingale。
- AI 不能执行任意 PromQL，只能使用白名单模板或受控查询。
- AI 不能展示内部连接串、完整 token、cookie、SSH key、数据库 DSN。
- AI 回写 Nightingale 事件备注或 event pipeline 输出时，必须脱敏并记录审计。
- AI 不能直接要求 Agent 执行 shell、重启服务、kill 进程、修改文件或变更配置。
- Agent 工具输出进入 AI 前必须做字段脱敏、长度截断和敏感路径过滤。

## 12. 分阶段开发计划

### Phase 0：立项、合规和模板基线

目标：冻结技术路线、许可证边界、命名边界、模板边界、Agent 能力边界和迁移边界。

工作项：

- 明确 Nightingale 为监控事实源和模板事实源。
- 明确 Categraf 为 findx-agents 采集底座。
- 明确 Catpaw 不作为 findx-agents 派生源码。
- 新增开源组件 NOTICE/THIRD_PARTY 说明。
- 定义产品命名、内部命名和兼容命名。
- 定义禁用重复建设清单：告警、仪表盘、通知、事件流水线、采集插件、目标资产、任意远程执行。
- 定义 findx-agents 三层结构：Categraf 采集核心、findx-inspector 巡检侧车、FindX Agent Control Protocol。
- 定义 Catpaw clean-room 迁移原则：参考能力、事件模型和安全边界，不直接复制 AGPL 源码进入默认主线。
- 建立 FindX 模板中心元数据模型和同步任务设计。
- 定义 Nightingale team/user-group、busi-group、role/perms 与 FindX 团队/业务/权限的映射模型。
- 定义 InspectionRun、InspectionFinding、AgentSession、AgentAuditLog 标准模型。
- 明确 UI 统一沿用 FindX 当前风格，不照搬 Nightingale 菜单和页面视觉。

验收：

- 文档确认能力归属矩阵。
- 产品命名中出现 `findx-agents`。
- 合规说明覆盖 Nightingale/Categraf/Catpaw。
- 模板中心覆盖 integrations、dashboards、alerts、metrics、record-rules、collect toml、notify templates、message templates、event pipelines、embedded dashboards。
- 团队/RBAC 映射方案覆盖 user-groups、busi-groups、roles/perms、team_ids、ro/rw 权限。
- Agent 规划明确远程禁用 shell、结构化只读工具、审计、低权限运行和 clean-room/AGPL 侧车二选一路线。

### Phase 1：Nightingale Connector 与模板目录完整接入

目标：AI WorkBench 通过受控 connector 访问 Nightingale，不再直接依赖 Nightingale MySQL/Redis 表结构。

工作项：

- 新增 `api/internal/nightingale/` connector 包。
- 支持 Nightingale service API base URL、BasicAuth/token、timeout、retry、TLS 配置。
- 封装 targets、target tags、busi groups、datasources、alert-cur-events、alert-his-events、dashboards、recording-rules、builtin-components、builtin-payloads、builtin-metrics、message-templates、event-pipelines。
- 封装 users、roles、self perms、user-groups、user-group-members、busi-group members、busi-group perm check 的读取、校验、映射和审批写入能力。
- 替换当前直接读 `n9e.redis_*` 和 `n9e.mysql_dsn` 的列表接口，逐步切到 Nightingale API。
- 新增 FindX 模板中心 API：列表、详情、筛选、预览、hash、安装状态、差异预览、审批安装、回滚引用。
- 新增 FindX 团队映射 API：团队、业务组、成员、权限摘要、值班组绑定状态、审批绑定、审计记录。
- 所有 Nightingale 错误统一脱敏，不返回内部连接串。
- 加 connector 单元测试和 fake server 测试。

验收：

- `GET /api/v1/findx/targets` 能通过 Nightingale API 返回目标列表。
- `GET /api/v1/findx/alerts/active` 能返回 Nightingale 活跃告警摘要。
- `GET /api/v1/findx/templates` 能返回 integrations/builtin payload 目录。
- `GET /api/v1/findx/teams` 能返回 Nightingale user-groups/team 摘要。
- `GET /api/v1/findx/business-groups` 能返回 Nightingale busi-groups 和 ro/rw 权限摘要。
- 断开 Nightingale 时返回 503 和可读错误，不影响 AI WorkBench 基础启动。

### Phase 2：findx-agents 采集与巡检完整闭环

目标：用 findx-agents 品牌封装 Categraf，并落地可生产验收的 findx-inspector 单机巡检闭环。

工作项：

- 设计 findx-agents 安装目录、服务名、配置目录和日志目录。
- 生成 Categraf 配置模板：writers、heartbeat、global labels、inputs。
- 默认 heartbeat 指向 Nightingale `/v1/n9e/heartbeat`。
- 增加可选 FindX alive heartbeat：只上报 agent_id、hostname、ip、version、status、last_seen，不承载监控数据。
- 保留 Categraf 原始能力和配置结构，不 fork 采集插件。
- 新增 findx-inspector sidecar：完整交付 `linux_quick` profile，覆盖 uptime/load、CPU top、内存、磁盘、inode、网络错误、监听端口、僵尸进程、systemd failed units、超时、审计、错误态和脱敏。
- 新增 Agent Control Protocol P0：register、heartbeat、session_start、session_output、inspection_result。
- 新增 InspectionRun/InspectionFinding API 与存储，只保存结构化结果和脱敏证据。
- 远程 session 禁用 shell，工具输出截断，工具调用写本地 JSONL 审计。
- 安装脚本从 `catpaw` 改为 `findx-agents`。
- 前端探针管理页从 Catpaw 重命名为 findx-agents。

验收：

- Linux 目标机安装后出现 `findx-agents.service`。
- Nightingale targets 中能看到对应主机 heartbeat。
- AI WorkBench findx-agents 页面能看到在线/离线状态。
- 用户可对单台 Linux 主机发起 `linux_quick` 巡检，得到 InspectionRun 和 AI 可引用的 findings。
- Agent 离线、巡检超时、工具失败都能返回可读状态，不影响 Categraf 指标采集。
- 不再安装 `/usr/local/bin/catpaw` 或 `/etc/catpaw`。

### Phase 3：模板安装与 Dashboard 嵌入

目标：用户可在 FindX 模板中心安装 Nightingale/Categraf 成熟模板。

工作项：

- 支持 dashboard payload 安装和 `/board/:bid/pure` 嵌入。
- 支持 alert payload 安装。
- 支持 collect payload 生成 Categraf `input.*` 配置。
- 支持 record-rules 安装。
- 支持 metric desc/builtin metrics 同步为 AI 指标字典，并具备差异预览、版本标识和回滚说明。
- 支持 markdown/runbook 预览和 AI 引用。
- 支持模板安装前差异预览、变量填写、权限确认和回滚。

验收：

- Linux、MySQL、Redis 至少三个 integration 可完整展示 dashboard/alert/metrics/collect/markdown。
- Linux dashboard 可在 FindX 内嵌展示。
- MySQL 模板安装后，findx-agents 生成 input 配置，Nightingale 可见对应 dashboard/alert。
- 安装失败时不留下半写状态，错误脱敏。

### Phase 4：监控能力下沉 Nightingale

目标：把监控管理能力从 AI WorkBench 自研逻辑迁移到 Nightingale。

工作项：

- 告警中心改读 Nightingale 活跃/历史事件。
- 告警规则创建、更新、删除走 Nightingale 接口或嵌入页。
- 仪表盘入口优先嵌入 Nightingale dashboard。
- 数据源管理优先使用 Nightingale datasource。
- 通知渠道和订阅管理优先跳转或代理 Nightingale 页面。
- AI WorkBench 本地只保留 AI 诊断记录、知识库、用户会话和 findx-agents 存活缓存。

验收：

- 用户在 FindX 告警中心看到的数据与 Nightingale 活跃告警一致。
- 用户在 FindX 仪表盘看到的是 Nightingale dashboard 或 Nightingale dashboard 的受控嵌入。
- AI WorkBench 不再维护第二套告警规则事实源。

### Phase 5：AI 问诊与 Nightingale 模板证据链融合

目标：AI 问诊从 Nightingale 获取更完整证据，并能按需联动 findx-agents 补充单机证据。

工作项：

- AIOps 诊断上下文新增 NightingaleEvidenceProvider。
- 诊断输入包括 target、busi group、active alerts、alert rule、PromQL、dashboard link、datasource status、metric desc、record rules、integration markdown、event pipeline status。
- AIOps 工作流新增 AgentEvidenceProvider，按 target/agent_id 下发 `agent_inspect` 节点并收集 InspectionFinding。
- `alert_agent_diagnosis` 工作流：Nightingale 告警触发后，自动判断是否需要 `linux_quick`、`network_basic` 或 `middleware_hint`。
- `business_health_inspection` 工作流：按业务组并发触发多个 Agent quick inspect，生成值班交接摘要。
- AI 结论可回写到 AI WorkBench 诊断记录。
- 可选将诊断摘要回写到 Nightingale 事件备注、通知或 webhook。
- PromQL 查询优先走 Nightingale datasource/query 或受控 Prometheus 模板。
- suggestedActions 默认进入建议队列，写操作必须审批、二次确认、审计和回滚。

验收：

- 对某个 Nightingale 告警一键发起 AI 问诊。
- 诊断报告能引用 Nightingale 事件、规则、目标、PromQL、指标解释、dashboard、markdown/runbook 和 findx-agents 巡检证据。
- AI 能推荐缺失模板，但不会自动写入。
- Agent 离线时诊断降级为 Nightingale-only，并在报告中明确证据缺口。
- 诊断结果不出现裸 JSON、内部连接串、真实 token 或上游 raw error。

### Phase 6：通知、事件流水线与 Hermes

目标：复用 Nightingale 通知与事件流水线，把 AI 问诊作为增强节点接入。

工作项：

- 读取 message templates、notify templates、notify channels、notify rules。
- 支持 event pipeline 列表、详情、tryrun、trigger、stream、executions。
- AI 问诊完成后可生成脱敏摘要，作为 event pipeline 输出或 Hermes 推送内容。
- Hermes 只作为消息触达和异步入口，不成为监控事实源。

验收：

- FindX 能展示 Nightingale 通知模板和事件流水线摘要。
- AI 诊断摘要可通过受控 webhook/event pipeline 流转。
- 所有写操作要求管理员权限、二次确认和审计。

### Phase 6.5：团队、值班和运维闭环

目标：把 Nightingale 团队/业务组/权限接入 FindX 运维工作流，让告警处置能按责任团队闭环。

工作项：

- 同步 Nightingale user-groups、members、roles、self perms、busi-groups、busi-group members。
- 新增 FindX team/business mapping：把 Nightingale user group 映射为 FindX 团队，把 busi group 映射为 FindX 业务系统/拓扑业务。
- 值班组绑定 Nightingale team，通知渠道绑定 Nightingale notify channel/message template。
- 告警中心按团队、业务组、值班状态、P0/P1、未认领、升级中筛选。
- 运维总览新增“值班驾驶舱”：当前责任团队、未处理告警、升级计时、交接班摘要。
- 故障复盘从 Nightingale events、notify records、event pipeline executions、FindX AI 诊断记录生成时间线。
- 变更窗口从 Nightingale alert mutes 和 dashboard annotations 读取，AI 提示影响面和缺失静默。
- SLO/错误预算从 record rules、dashboard 指标和告警历史生成业务健康趋势。

验收：

- FindX 团队页能展示 Nightingale 团队/用户组、成员、关联业务组、ro/rw 权限。
- 告警中心能按团队/业务组过滤，过滤结果与 Nightingale 权限一致。
- 值班驾驶舱能展示当前值班组、未认领 P0/P1、升级策略和交接班摘要。
- 复盘报告能引用事件时间线、通知记录、AI 问诊和恢复证据。
- UI 保持 FindX 当前风格，不直接暴露 Nightingale 原始页面作为主体验。

### Phase 7：Catpaw 能力迁移与兼容删除计划

目标：平滑迁移旧 Catpaw 页面、API、数据和能力，把“Catpaw 式巡检”转为 findx-inspector 标准能力。

工作项：

- `/api/v1/catpaw/*` 标记为 deprecated。
- 新增 `/api/v1/findx/agents/*`。
- 旧 `catpaw_agents` 表迁移或映射到 `findx_agents`。
- 旧页面 `CatpawInstall.vue` 重命名或替换为 `FindxAgents.vue`。
- 知识库同义词从 Catpaw 扩展到 findx-agents。
- 按 clean-room 方式迁移 Catpaw 诊断工具能力：cpu、mem、disk、diskio、netif、procnum、systemd、journal/log、cert、http、redis。
- 可选 AGPL 侧车路线独立开关：明确许可证提示、源码归属、部署边界和数据协议。
- 文档中说明 Catpaw 仅为历史名称。

验收：

- 新用户路径不再出现 Catpaw。
- 老接口可在兼容期返回 deprecation header。
- 旧数据可被 findx-agents 页面读取。
- 新巡检结果统一进入 InspectionRun/InspectionFinding，不再暴露 Catpaw 原始 event 作为主模型。

### Phase 8：联调、压测和交付

目标：形成可交付的一体化方案。

工作项：

- WSL/Linux 单包构建。
- Nightingale + Categraf/findx-agents + AI WorkBench docker compose 或 systemd 联调。
- findx-agents 安装、卸载、重装、升级、离线、恢复测试。
- findx-inspector 单机巡检、并发 session、取消、超时、审计、脱敏测试。
- Nightingale API 断连、认证失败、超时、返回结构变化测试。
- 告警事件到 AI 问诊再到 Agent 补证据的闭环测试。
- 许可证和 NOTICE 文件检查。

验收：

- `go build` 和 `npm run build` 通过。
- Nightingale 目标、告警、仪表盘、模板在 FindX 入口可见。
- findx-agents 在线状态准确。
- AI 问诊能基于 Nightingale 告警和 findx-agents 巡检证据完成一次闭环诊断。

## 13. P0 / P1 / P2 排期

### P0：必须先做

1. 能力边界冻结：Nightingale 为监控事实源和执行引擎，FindX 不重写 Nightingale 成熟能力。
2. 许可证策略冻结：默认 clean-room 实现 findx-inspector；Catpaw AGPL 只作为可选独立侧车路线。
3. Nightingale 459 路由功能域清单冻结，Connector P0 对纳入范围完整覆盖 targets、busi-groups、teams/RBAC、datasources、alerts/events、mutes/subscribes、notify、dashboards、builtin payloads 的列表、详情、权限、错误和审计。
4. FindX 模板中心正式目录：integrations、dashboards、alerts、metrics、record-rules、markdown、collect toml、notify templates、message templates、event pipelines，包含预览、差异、安装、回滚和权限。
5. 告警规划基线：规则生命周期、P0/P1 等级映射、通知/静默/订阅、聚合视图、事件证据链、覆盖率治理。
6. findx-agents 命名、服务名、路径、安装模板、Categraf heartbeat 到 Nightingale。
7. findx-inspector P0 正式切片：完整交付 `linux_quick` 单机巡检、InspectionRun/InspectionFinding 模型、远程禁用裸 shell、结构化工具、工具审计、错误态和脱敏。
8. Agent Control Protocol P0：register、heartbeat、session_start、session_output、inspection_result。
9. AI WorkBench 工作流节点设计：`nightingale_evidence`、`agent_inspect`、`inspection_report_save`。
10. Dashboard 完整嵌入验收：`/board/:bid/pure`，覆盖鉴权、业务组权限、变量、时间范围、断连、白屏保护和敏感信息脱敏。
11. 敏感配置治理：Nightingale、Categraf、Agent token、AI key、数据库连接串不得明文进入文档、日志或前端。

### P1：第一轮可用闭环

1. FindX 告警中心读取 Nightingale 活跃/历史告警、聚合视图、通知记录、评估详情和 trace logs。
2. FindX 目标资产读取 Nightingale targets/busi groups/team 权限，并与 findx-agents 在线状态关联。
3. FindX 仪表盘嵌入 Nightingale dashboard，模板中心支持安装 dashboard、alert、collect、record-rule。
4. FindX 通知/静默/订阅中心支持 Nightingale notify/mute/subscribe 列表、详情、试跑、审批写入、审计和回滚。
5. findx-agents 安装/卸载/升级脚本，Categraf local/http provider 配置生成。
6. findx-inspector 支持 `linux_quick`、`linux_deep`、`network_basic` 三类巡检 profile。
7. `alert_agent_diagnosis` 工作流：Nightingale 告警触发 AI 问诊，并按需调用 Agent 补充本机证据。
8. `business_health_inspection` 工作流：按业务组批量巡检，生成值班交接摘要。
9. 旧 Catpaw 页面改名为 findx-agents 并保留兼容层。
10. Nightingale API fake server、Agent fake server、WSL 联调和敏感信息扫描。

### P2：增强与商业化体验

1. 告警规则、通知渠道、消息模板、订阅策略、静默策略、event pipelines 的品牌化编辑入口。
2. AI 推荐告警规则、dashboard、静默、通知模板、inspection profile，人工确认后写入 Nightingale/FindX。
3. Hermes 接收 Nightingale 告警摘要、Agent 巡检摘要和 AI 诊断结果。
4. Categraf http_provider 远程配置下发，Inspector profile 远程配置下发，支持 version/hash/drift。
5. Catpaw 能力 clean-room 迁移：redis、redis_sentinel、http、cert、docker、systemd、journal/log、filefd、procfd、sockstat、sysctl。
6. Per-Agent Key、mTLS、session nonce、防重放、Agent 审计日志归档。
7. findx-agents 多平台安装包、升级通道、版本矩阵。
8. Nightingale AI assistant/skills/MCP 与 FindX AI 的边界互操作。
9. 执行动作二次确认与审计；自动修复必须先封装为结构化任务并具备审批、试跑、回滚和失败保护。

### 独立专项

- Catpaw AGPL 独立侧车正式商业交付：需要许可证、源码发布、部署边界和升级策略专项。
- Catpaw AI chat 直接迁移：需要与 FindX AI 问诊边界、上下文和审计合并专项。
- 远程 shell/自动修复和本地人工 shell 模式：需要结构化任务、审批、试跑、回滚、最小权限和审计专项。
- Ibex 远程任务深度接入：需要模板任务、权限、审批、审计、回滚、并发和失败保护专项。
- 直接改造 Nightingale 前端源码：需要上游跟随、品牌隔离、合规和维护成本专项。

## 14. 主要风险

| 风险 | 等级 | 说明 | 缓解 |
| --- | --- | --- | --- |
| Catpaw AGPL 派生风险 | P0 | 直接重命名派生会带来更强开源义务。 | 默认 clean-room 实现 Inspector；AGPL 独立侧车只作为明确接受许可证义务的可选路线。 |
| 双事实源 | P0 | AI WorkBench 与 Nightingale 同时维护告警、资产、模板会冲突。 | Nightingale 为事实源，AI WorkBench 做视图、索引缓存和 AI。 |
| Agent 变成第二套监控 | P0 | 巡检 finding 如果替代 Nightingale rule/event，会造成双告警体系。 | Categraf 持续采集，Nightingale 规则告警，Inspector 只做按需/低频巡检和证据补充。 |
| 远程执行风险 | P0 | Agent 控制面一旦开放裸 shell 或脚本，会扩大攻击面。 | 远程禁用裸 shell，结构化工具、schema 参数、输出截断、本地审计；Ibex 只允许模板化任务、审批、审计和回滚。 |
| Agent 控制面认证风险 | P0 | 注册 token 泄露或重放可能影响大量主机。 | P0 HMAC，P1 Per-Agent Key，P2 mTLS/session nonce/token 轮转。 |
| 模板安装漂移 | P1 | 上游模板升级后，本地安装对象可能漂移。 | hash、版本、source、installed_ref、diff 和 rollback。 |
| Nightingale API 变动 | P1 | 直接依赖内部表或不稳定 API 会脆弱。 | connector 层隔离，fake server 测试。 |
| Categraf 包装过深 | P1 | fork 太深会难以跟随上游。 | 只包装安装、配置和命名，不改采集插件核心。 |
| 敏感配置泄露 | P0 | token、DSN、密码可能进入配置和日志。 | 环境变量/secret、脱敏、审计、文档占位符。 |
| UI 命名和开源归属冲突 | P1 | 用户侧要 FindX 命名，合规侧要保留归属。 | 产品命名和合规说明分层。 |
| AI 问诊证据污染 | P1 | 旧 Catpaw 数据、Nightingale 数据和 Agent 巡检数据同时存在。 | 诊断 evidence 标准化，标记 source、freshness、upstream_ref、agent_run_id。 |
| 工具输出泄密 | P1 | 日志、进程、配置、命令输出可能包含密钥或隐私。 | 敏感字段过滤、长度上限、路径过滤、报告脱敏、原始证据权限隔离。 |
| 巡检并发冲击主机 | P1 | 批量巡检可能造成 CPU/IO/网络抖动。 | Agent/业务组/用户三层并发限制，低优先级队列，超时和熔断。 |
| Dashboard 嵌入权限 | P1 | iframe/session/proxy 处理不当会越权或白屏。 | 优先 pure board + 后端代理/同源鉴权，覆盖权限测试。 |
| Event pipeline 写操作风险 | P1 | 流水线可触发外部动作。 | 支持试跑、审批写入、二次确认、审计、权限控制和回滚引用。 |

## 15. 需要修改的主要模块

AI WorkBench 后端：

- `api/main.go`
- `api/internal/handler/n9e_agents.go`
- `api/internal/handler/n9e_alerts.go`
- `api/internal/handler/catpaw.go`
- `api/internal/handler/remote.go`
- `api/internal/store/agents.go`
- `api/internal/model/diagnose.go`
- `api/internal/handler/aiops_*`
- 新增 `api/internal/nightingale/`
- 新增 `api/internal/findxagent/`
- 新增 `api/internal/templatecatalog/`
- 新增 `api/internal/inspection/`
- 新增 `api/internal/agentcontrol/`
- 新增工作流节点：`nightingale_evidence`、`agent_inspect`、`inspection_report_save`

AI WorkBench 前端：

- `web/src/views/CatpawInstall.vue` -> `FindxAgents.vue`
- `web/src/views/CatpawChatPanel.vue` 后续移除或兼容隐藏
- `web/src/views/Diagnose.vue`
- `web/src/views/Workbench.vue`
- `web/src/views/DataSource.vue`
- `web/src/views/Alerts.vue`
- `web/src/router/index.js`
- 新增 FindX 目标、告警、仪表盘、模板中心、Agent 巡检、巡检报告、Agent session 审计入口

文档与脚本：

- `README.md`
- `docs/aiops/`
- `docs/运维手册.md`
- `scripts/build-one-package.sh`
- `scripts/install-one-service.sh`
- 新增 findx-agents 安装模板、findx-inspector profile 模板、Nightingale/Categraf/Catpaw 第三方许可证说明、模板中心使用手册。

findx-agents / Inspector：

- 新增 `cmd/findx-agents-supervisor` 或安装脚本级 supervisor。
- 新增 `findx-inspector` 二进制或 sidecar。
- 新增 `profiles/linux_quick.yaml`、`profiles/linux_deep.yaml`、`profiles/network_basic.yaml`。
- 新增 Agent Control Protocol 客户端、HMAC/Per-Agent Key、session runner、tool registry、audit logger。
- Categraf 保持独立升级，不改 inputs 主逻辑。

## 16. 验证计划

代码验证：

- 后端：`cd /opt/ai-workbench/api && go build -o /tmp/aiw-nightingale-plan-api .`
- 前端：`cd /opt/ai-workbench/web && npm run build`
- Connector：fake Nightingale server 单测覆盖 targets、alerts、datasources、builtin-components、builtin-payloads、message-templates、event-pipelines、timeout、401/403、schema drift。
- Template catalog：JSON/TOML 校验、hash 幂等、drift 标记、安装失败回滚。
- Agent control：fake Agent server/client 覆盖 register、heartbeat、session_start、session_output、inspection_result、cancel、timeout、断线重连。
- Inspector：Linux `linux_quick` profile 单测/集成测试覆盖工具 schema、输出截断、敏感字段脱敏、JSONL 审计。
- Workflow：`alert_agent_diagnosis` 覆盖 Nightingale-only、Agent 在线补证据、Agent 离线降级三条路径。

联调验证：

- Nightingale 启动后，AI WorkBench 能读取 targets。
- Categraf/findx-agents heartbeat 后，Nightingale target 在线。
- AI WorkBench findx-agents 页面显示在线。
- Nightingale 产生告警后，FindX 告警中心显示一致。
- 从告警触发 AI 问诊，诊断报告引用 Nightingale 证据、模板说明和 findx-agents 巡检证据。
- 单台主机巡检可从 FindX 页面发起，Agent 返回 InspectionRun/InspectionFinding。
- 业务组巡检可并发触发多个 Agent，受并发限制和超时保护。
- Linux dashboard 可通过 `/board/:bid/pure` 在 FindX 内嵌展示。
- MySQL/Redis/Linux 模板可完整预览并安装。

安全验证：

- 配置、日志、错误、前端响应不泄露真实 `<API_KEY>`、`<TOKEN>`、`<DB_DSN>`、`<PASSWORD>`、`<SSH_KEY>`。
- Nightingale connector 认证失败返回可读错误，不返回内部栈。
- 任意写操作默认拒绝或要求二次确认。
- Event pipeline trigger/stream 必须有权限、审计和速率限制。
- Agent 远程 session 不注册 shell，不允许写文件、重启服务、kill 进程或修改配置。
- Agent 工具调用必须写审计日志，平台报告只展示脱敏摘要。

## 17. 上一版扣分点补齐情况

| 上一版扣分点 | 本轮补齐方式 | 剩余风险 |
| --- | --- | --- |
| Nightingale 集成不够深 | 已纳入 integrations、builtin components、builtin payloads、metrics desc、record rules、message templates、notify templates、notify channels、event pipelines、embedded dashboards/products。 | 仍需运行态完整联调验证 API 字段。 |
| 只覆盖部分 Nightingale 功能 | 重新抽取 `router.go` 459 条路由，按身份、团队、业务组、目标、数据源、指标、Dashboard、告警、通知、日志、任务、保存视图、Webhook、AI assistant 做全功能域映射。 | 仍需开发阶段逐字段适配。 |
| 告警规划不足 | 新增告警规则生命周期、告警分类、P0/P1 映射、通知/静默/订阅/升级、事件处置台、治理看板和写操作门禁。 | 规则写入仍需审批和回滚实现。 |
| 模板体系未完整覆盖 | 新增 FindX 模板中心方案，覆盖 dashboard/alert/collect/metric/record-rule/markdown/notify/event pipeline 全类型。 | 模板安装的字段映射需要开发阶段逐项验证。 |
| Categraf 模板不够细 | 已审 92 个 input 目录、105 个 toml，明确 findx-agents 生成/管理 Categraf 配置，不 fork 插件。 | http_provider 远程下发需要安全联调验证。 |
| Catpaw 能力被暂时搁置 | 新增 findx-inspector 方案：Categraf 采集核心 + Catpaw 风格单机巡检/结构化工具/远程 session/审计 + FindX 工作流联动。 | 默认 clean-room 需要开发；AGPL 侧车需专项合规。 |
| AI 问诊与 Agent 联动不足 | 新增 `agent_inspect` 工作流节点、`alert_agent_diagnosis`、`business_health_inspection`、InspectionRun/Finding 模型和 Agent 离线降级策略。 | 需要真实 Agent 联调验证延迟、并发和证据质量。 |
| 未做 dashboard 完整嵌入验证 | 已指定 `/board/:bid/pure` 为 P0 正式验收路径，并补权限、变量、时间范围、断连、白屏和脱敏验收。 | 尚未真实启动 Nightingale 浏览器验证。 |
| 未启动 Nightingale/Categraf/Agent 实例联调 | 补充 fake server、docker compose/systemd 联调、Agent control fake server、Inspector linux_quick 正式验收和 P0/P1 验收路径。 | 本轮仍为静态审计和文档立项，没有宣称已完成真实联调。 |
| Dify/FastGPT/RAGFlow 未逐项审 | 明确它们不是本次监控集成主线，仅作为 AI/RAG 参考项目；后续引入需单独立项。 | 若要复用其代码/UI，仍需许可证专项。 |
| 许可证判断不能替代法律意见 | 保留为残余风险，新增 NOTICE/THIRD_PARTY 和正式法务确认要求。 | 需正式法务确认。 |

## 18. 本计划自评分

综合评分：96 / 100。

| 维度 | 权重 | 得分 | 说明 |
| --- | ---: | ---: | --- |
| 源码证据充分性 | 30 | 29 | 已重新抽取 Nightingale 459 条路由，补充模型层、integrations、builtin payload、event pipeline、embedded product；审计 Categraf AgentModule/providers/heartbeat/huatuo sidecar/ibex；审计 Catpaw inspect、plugin、diagnose、server protocol、安全设计。未做全量逐行审计。 |
| 架构一致性 | 25 | 25 | 明确 FindX 是 Nightingale 全功能门户，Nightingale 是事实源和执行引擎，Categraf 是采集核心，findx-inspector 是巡检侧车，AI 问诊是增强层。 |
| 许可证与风险识别 | 20 | 18 | 识别 Nightingale Apache-2.0、Categraf MIT、Catpaw AGPL-3.0 关键差异，并补 NOTICE/THIRD_PARTY 要求。仍需正式法务确认。 |
| 可执行性 | 15 | 15 | 已拆 Phase/P0/P1/P2、全功能域映射、告警规划、模板中心、Agent control、InspectionRun/Finding、工作流节点、完整验收和模块清单。 |
| 验证完整性 | 10 | 9 | 已补 Nightingale fake server、Agent fake server、Inspector profile 测试、Dashboard 完整嵌入验收、模板校验、联调、安全验证；本轮仍未真实启动 Nightingale/Categraf/findx-agents 实例。 |

扣分项保留：

- 尚未启动 Nightingale/Categraf/findx-agents 实例做真实 API/Agent 联调。
- 尚未完成 FindX 前端嵌入 Nightingale dashboard 的代码级完整验收。
- 尚未完成 findx-inspector `linux_quick` 代码级完整验收。
- 许可证判断基于本地源码 LICENSE，不替代正式法律意见。
- 模板安装字段映射需在开发阶段逐项用真实对象验证。
- Catpaw 能力默认走 clean-room 迁移，真实功能覆盖率需要开发阶段逐插件验收。

建议结论：本计划已达到“可立项，建议进入 P0 正式设计与完整功能切片开发”的程度。相比上一版，核心补齐点是 Nightingale 全功能承接、完整告警规划、Categraf + Catpaw 风格巡检融合、新 Agent 与 AI 工作流联动，以及“不允许最小实现”的交付硬约束。是否进入开发由主 AI/项目负责人决策。

## 19. 参考源

- 本地源码：`D:\平台源码\nightingale-main (1)\nightingale-main`
- 本地源码：`D:\平台源码\categraf-main (1)\categraf-main`
- 本地源码：`D:\平台源码\catpaw-master\catpaw-master`
- 本地源码：`D:\平台源码\AutoOps-main\AutoOps-main`
- 当前 git 工作区：`D:\ai-workbench`
- 当前 WSL 运行项目：`/opt/ai-workbench`
- 官方仓库：`https://github.com/ccfos/nightingale`
- 官方仓库：`https://github.com/flashcatcloud/categraf`
