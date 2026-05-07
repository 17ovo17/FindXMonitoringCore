# Superseded: Replaced on 2026-05-07 by `docs/aiops/findx_full_stack_observability_long_term_plan.md`. Historical evidence only; not an implementation entrypoint.

# FindX Monitoring Core P0-P3 实施总计划

生成时间：2026-05-04 04:30（UTC+8）

## 1. 文档定位

本文档是 FindX Monitoring Core 从 P0 到 P3 的讨论总计划，用于指导主代理、子代理、QA 与后续 Git 稳定切片持续推进。当前主代理在 Claude 缺席时承担编排者、审计者、评分者和 Git 门禁职责；子代理承担执行层 work unit。

本文档不记录真实密钥、认证票据、Cookie、完整连接串、SSH 私钥或生产数据。所有敏感示例统一使用 `<TOKEN>`、`<API_KEY>`、`<DB_DSN>`、`<LOGIN_USER>`、`<SSH_KEY>`、`<COOKIE>`。

## 2. 总体目标

FindX Monitoring Core 必须形成自有监控闭环：

```text
target/datasource/agent
  -> query gateway
  -> alert rule/evaluator
  -> current/history event
  -> notification/silence/subscription/oncall
  -> dashboard/template
  -> findx-agents inspect/diagnose/session
  -> AI diagnose
  -> remediation precheck/dry-run/approve/execute/verify/rollback
  -> audit/knowledge archive
```

核心交付原则：

- 全功能推进 P0/P1/P2/P3，不以占位页、静态假数据、半截接口或一次性 PoC 作为完成结论。
- API、数据模型、权限、审计、UI、Agent、AI 与修复执行链路最终归 FindX 主线承载。
- Categraf 作为 `findx-agents` 的采集插件生态底座，先明确协议、安全边界和配置分发，再迁移具体采集能力。
- Catpaw 授权衍生能力进入 `findx-agents` 前，必须补齐来源说明、修改说明、授权边界、NOTICE 和合规材料。
- 所有写操作必须有权限、审计、幂等、回滚或失败清理策略。
- 所有 AI 输出必须绑定 evidence refs，不允许凭空生成诊断结论或执行建议。

## 3. 阶段地图与 P0 状态

| 阶段 | 当前状态 | 主要输出 | 进入下一阶段条件 |
| --- | --- | --- | --- |
| P0 | 已完成并推送：`81b4531`，QA `98/100` 通过 | target、datasource、agent register、agent heartbeat、health、alert rule、current/history event、tryrun、rollback、query gateway | Windows + WSL Go 测试/构建通过，QA 已确认核心路径、权限、断连、脱敏和降级闭环。 |
| P1 | 进行中，P1-BE-1/2/3 已完成并推送：`82ea50d`、`1b37905`、`1c4045e` | evaluator、scheduler、current/history event 自动闭环已完成；notification、silence、subscription、oncall、permission、audit、Dashboard、template、`/monitor` 工作台待推进 | P1-BE-3 QA `92/100 PASS`，Windows/WSL `go test -count=1 ./...` 与 build 通过；下一阶段从 P1-FE-1 或 P1-BE-4 基座开始。 |
| P2 | 待协议先行 | findx-agents 协议、能力目录、安全模型、证据模型、control plane、session/evidence、Categraf 融合、Catpaw 衍生 inspector/tool registry | 协议、安全审批、离线保护、配置分发、合规 NOTICE 与 Agent 回归完成。 |
| P3 | 待 evidence chain 先行 | AI evidence chain、remediation plan、precheck、dry-run、approve、execute、verify、rollback | 低风险动作闭环通过，失败保护和审计闭环完整，敏感信息不进入 AI prompt 和报告。 |

P0 验证命令摘要：

```bash
# Windows 项目目录
cd D:\ai-workbench\api
go test ./...
go build -o api.exe .

# WSL 镜像目录
cd /opt/ai-workbench/api
go test ./...
go build -o api-linux .
```

P0 验收摘要：

- QA 评分：`98/100`。
- 推送提交：`81b4531`。
- 已覆盖：target CRUD、datasource/query gateway、agent register/heartbeat、alert rule 生命周期、tryrun、current/history event 状态机、权限、断连、降级、脱敏、审计。
- 遗留处理：P1-BE-1/2/3 已把 evaluator、scheduler、fingerprint 幂等和 eval log 接入真实事件闭环；P1 继续推进 `/monitor` 前端工作台、Dashboard、模板、通知、权限、团队、值班、订阅、静默。

## 4. P1 后端与前端工作单

P1 告警后端闭环的前三个稳定切片已完成：P1-BE-1 域常量/状态机/fingerprint 幂等、P1-BE-2 evaluator service + 真实 tryrun、P1-BE-3 scheduler + current/history event 自动闭环。后续从 `/monitor` FindX Monitoring Core 前端入口或 P1-BE-4 通知/权限/值班/订阅/静默基座继续推进，不继续堆旧 `Alerts.vue` 或 `OnCallConfig.vue`。

| Work Unit | 执行角色 | 写集边界 | 依赖 | 验收命令/用例 | 是否允许 Git 稳定切片 |
| --- | --- | --- | --- | --- | --- |
| P1-BE-1 域常量/状态机/fingerprint 幂等 | `go-backend` | 仅 `D:\ai-workbench\api\` 内 monitor 相关 model/service/store/test；禁止改 web、docs、配置、密钥 | 已完成并推送，提交 `82ea50d` | 已按切片验收完成；后续如改 fingerprint 或事件状态机，必须重新标记 DATA_CHANGE 并补 Windows/WSL `go test -count=1 ./...` 与 build | 已完成。切片名：`findx-p1-be-domain-fingerprint` |
| P1-BE-2 evaluator service + 真实 tryrun | `go-backend` | 仅 `D:\ai-workbench\api\` 内 evaluator/query gateway/rule tryrun 相关文件和测试；禁止引入新外部依赖 | 已完成并推送，提交 `1b37905` | 已按切片验收完成；后续如改 tryrun 响应、PromQL 校验或 eval log 写入语义，必须重新标记 API_CONTRACT_CHANGE/DATA_CHANGE | 已完成。切片名：`findx-p1-be-evaluator-tryrun` |
| P1-BE-3 scheduler + current/history event 自动闭环 | `go-backend` | 仅 `D:\ai-workbench\api\` 内 scheduler/evaluator/event store/test；禁止改前端和文档 | 已完成并推送，提交 `1c4045e feat: add FindX alert scheduler event closure` | QA `92/100 PASS`；Windows/WSL `go test -count=1 ./...` 与 build 通过；非阻断风险进入后续治理 | 已完成。切片名：`findx-p1-be-scheduler-event-closure` |
| P1-BE-4 silence/subscription/notification/oncall/pipeline/permission/audit 分组实现 | `go-backend` | 仅 `D:\ai-workbench\api\` 内 monitor 扩展模块；每次派发必须再拆写集，禁止一次性改完全部域 | 依赖 P1-BE-3 的事件闭环；需要先定义 API_CONTRACT_CHANGE、权限矩阵、审计字段和脱敏规则 | `cd /opt/ai-workbench/api && go test ./... && go build -o api-linux .`；用例覆盖 silence 匹配预览、subscription 去重、notification tryrun、oncall escalation、pipeline 失败清理、permission 403、audit 不含敏感值 | 允许，但必须拆成多个子切片，不允许一个 commit 混合全部域 |
| P1-FE-1 `/monitor` 工作台与 API 封装 | `vue-frontend` | `D:\ai-workbench\web\src\api\monitor.js`、`D:\ai-workbench\web\src\views\monitoring\*`、`D:\ai-workbench\web\src\components\monitoring\*`、必要 router 注册；禁止修改旧 `Alerts.vue`/`OnCallConfig.vue` 作为主入口 | 依赖 P0/P1 后端契约；如后端未全量完成，前端必须使用真实 API 错误态和空态，不写静态假数据 | `cd /opt/ai-workbench/web && npm run build`；MCP/Playwright 覆盖 `/monitor` 入口、接口失败态、空态、权限态、列表筛选、刷新 | 允许。建议切片名：`findx-p1-fe-monitor-workbench-api` |
| P1-FE-2 规则、事件、Dashboard、模板、通知、权限页面 | `vue-frontend` | 仅 `web/src/views/monitoring/*`、`web/src/components/monitoring/*`、`web/src/api/monitor.js` 必要扩展；不同页面分批派发 | 依赖 P1-FE-1；规则/事件依赖 P1-BE-2/P1-BE-3；通知/权限依赖 P1-BE-4 | `cd /opt/ai-workbench/web && npm run build`；MCP/Playwright 覆盖规则 tryrun、事件处置、Dashboard 真实取数失败态、模板 diff、通知 tryrun、权限 403 | 允许，但必须按页面或功能域拆分稳定切片 |

### P1-BE-1 DATA_CHANGE 与兼容说明

- DATA_CHANGE：P1-BE-1 改变 `monitor_alert_events_current` 写入语义和 fingerprint 生成语义，但不改 DDL。
- fingerprint 新语义：由 `rule_id + datasource_id + event_key + target_id + target_ident + 非敏感 labels` 生成 SHA-256；外部传入 fingerprint 不可信，不参与幂等。
- current event upsert：由 `REPLACE INTO` 风险语义改为保留首 ID 的 `INSERT ... ON DUPLICATE KEY UPDATE` 语义，count 累加，last_seen/updated_at 不回退。
- 兼容边界：FindX 是新平台，不迁移旧 Nightingale 或旧生产历史告警事件；本切片不做历史数据迁移。若未来导入旧事件，需要单独迁移/回填 fingerprint 并列回滚方案。
- 回滚方式：Git 回滚相关 Go 文件即可恢复旧写入逻辑；如未来已写入新事件，回滚前需评估 current event fingerprint 语义差异。
- 验证证据：P1-BE-1 已完成并推送，提交 `82ea50d`；后续只在修改 fingerprint 或事件写入语义时重新补 DATA_CHANGE 与回归证据。

### P1-BE-2 API_CONTRACT_CHANGE、DATA_CHANGE 与安全边界

- 目标边界：P1-BE-2 将告警规则 tryrun 从 dry validation 升级为真实 Prometheus instant query；新增无 Gin 依赖的 Prometheus query gateway service，并新增 evaluator core，供 tryrun 和后续 scheduler 复用。
- 写入边界：tryrun 只写 eval log，不写正式 `monitor_alert_events_current` 或 history event；正式事件闭环仍由 P1-BE-3 scheduler/evaluator 切片承接。
- 兼容响应：旧响应字段 `ok`、`status`、`checks`、`rule`、`eval_log` 必须保留；新增 `eval` 对象承载本次真实评估摘要、Prometheus instant query 归一化结果、`for_duration_ms`、`no_data_policy` 判定和最终评估状态。
- API_CONTRACT_CHANGE：新增 `eval` 对象；`rule.query` 在 tryrun 安全响应中允许按实现脱敏或置空，但不得删除 `rule` 字段本身。前端和 QA 应按“旧字段仍存在、新字段可消费”的兼容契约回归。
- API_CONTRACT_CHANGE：PromQL 校验统一复用 query gateway 安全口径，限制长度不超过 4096 字符，拒绝控制字符和高风险 admin/delete 语义；规则保存、tryrun、instant query、range query 必须使用同一校验策略。
- validation fail 兼容旧行为：输入校验失败时仍返回 HTTP 200，`ok=false`、`status=invalid`，写入失败原因摘要到 eval log，但不得请求 Prometheus。
- datasource not found：数据源不存在返回 HTTP 404，不降级为 Prometheus 网络错误或通用 503。
- Prometheus 安全 503：Prometheus 网络错误、非 2xx 响应、返回 `status:error`、invalid JSON 均返回安全 HTTP 503；响应体和 eval log 只保留脱敏后的错误分类、datasource id、rule id、trace id 和可操作摘要。
- 脱敏要求：raw PromQL 不进入 eval log `details`；响应、日志、eval log 均不得泄露 URL、token、password、secret、cookie、auth、DSN 或 upstream body。需要定位问题时只能记录 query hash、datasource id、duration、HTTP 状态类别和错误分类。
- 脱敏要求：Prometheus `warnings` 属于 upstream body 的一部分，必须在 gateway 层集中脱敏；非敏感 warning 可保留，含 token/password/secret/cookie/auth/DSN/API key/private 等片段的内容必须替换为 `<REDACTED>` 或安全摘要。
- `no_data_policy`：必须支持 `keep_state`、`alerting`、`ok` 三种策略；tryrun 仅返回本次评估结论，不修改正式事件状态。
- `for_duration`：解析规则配置并返回毫秒值 `for_duration_ms`；tryrun 不做跨周期 pending 存储，也不把单次 tryrun 结果提升为 pending/alerting 的正式状态机事实。
- DATA_CHANGE：无 DDL 变更；eval log `details` 写入内容从 dry validation 细节升级为安全评估摘要，属于写入语义变化。历史 eval log 不迁移，后续查询展示必须兼容旧 dry validation 记录和新真实评估摘要。
- 验证证据：P1-BE-2 已完成并推送，提交 `1b37905`；后续如修改 tryrun 响应、PromQL 校验、Prometheus gateway 或 eval log 写入语义，需要重新覆盖 valid 命中、valid no data + 三种 `no_data_policy`、validation fail 不请求 Prometheus、datasource not found 404、Prometheus 网络错误、非 2xx、`status:error`、invalid JSON、raw PromQL/URL/token/cookie/DSN 脱敏、tryrun 不写 current/history event、eval log 可追踪且不含敏感值。

### P1-BE-3 DATA_CHANGE 与自动事件闭环

- 目标边界：P1-BE-3 在 P1-BE-2 evaluator/gateway 基础上新增告警规则 scheduler，周期性评估启用规则，并把真实评估结果写入 `monitor_alert_events_current` 或恢复到 `monitor_alert_events_history`。
- 启动边界：`main.go` 只挂载 scheduler 启动函数；scheduler 必须提供可测试的单轮入口，ticker 只负责周期触发，不承载业务逻辑。
- 配置边界：新增 `monitoring.alert_scheduler.enabled`、`interval_seconds`、`timeout_seconds`、`max_concurrency`；默认不应让失败规则制造事件，超时和并发必须有上限。
- DATA_CHANGE：无 DDL 变更；`monitor_alert_events_current` 将由 scheduler 自动写入和合并，`monitor_alert_events_history` 将接收 scheduler recovery 动作产生的 resolved 事件；eval log 会新增 scheduler 评估状态和安全摘要。
- 事件语义：`Triggered=true` 且存在 candidates 时 upsert current event；`no_data_policy=alerting` 的 no data candidate 也 upsert current；`no_data_policy=ok`、scalar/string ok、empty vector ok 需要恢复同 rule 的旧 current event；`keep_state` no data 只写 eval log，不恢复也不新增。
- 失败保护：Prometheus 网络错误、非 2xx、`status:error`、invalid JSON、datasource not found、PromQL 校验失败、evaluator invalid 均只写 eval log，不得制造 current/history event。
- 幂等要求：同一 candidate 重复评估必须复用 P1-BE-1 fingerprint 合并语义，不新增重复 current event；同一轮 scheduler 不允许重入，单条规则失败不影响其他规则。
- 脱敏要求：scheduler 日志、eval log details、事件 labels/annotations 不得包含 raw PromQL、URL、token/password/secret/cookie/auth/DSN 或 upstream body；事件只允许保存 query hash、rule id、datasource id、candidate labels 的脱敏结果和值摘要。
- 验证矩阵：至少覆盖 firing 写 current、重复 firing 幂等、no_data alerting 写 current、no_data ok 恢复 history、keep_state 保持 current、旧 series 被恢复、upstream/status:error/invalid JSON 不造事件、datasource not found 不 panic、disabled rule 不评估、并发 RunOnce 锁保护、Windows/WSL Go test + build。
- QA 闭环补强：P1-BE-3 复审发现非敏感 label key 中的 URL/DSN/token/password/cookie/secret/auth value 可能进入 event labels、annotations 或 target_ident；实现必须在 scheduler 写入 current/history event 和 eval log details 前做 value 级脱敏，并补充 `instance`、`endpoint`、`token` 等样本测试。
- 全量验证补强：P1-BE-3 不只跑定向测试，还必须跑 Windows 与 WSL 的 `go test -count=1 ./...` 和构建；如果 WSL `/opt/ai-workbench` 与 Windows 工作区源码漂移，必须先同步完整 `api/**/*.go` 后再判定结果。
- 安全基线补强：本轮全量测试基线修复不得通过放宽测试掩盖真实问题；CORS 空配置默认必须保持 localhost-only，Workflow DSL 创建/更新必须在 parse 后执行 graph validation，坏 DSL 返回 400 且不得持久化。
- 完成证据：P1-BE-3 已完成并推送，提交 `1c4045e feat: add FindX alert scheduler event closure`；QA `92/100 PASS`，Windows/WSL `go test -count=1 ./...` 与 build 通过。

### P1 QA 非阻断风险关闭计划

| 风险 | 当前状态 | 关闭动作 |
| --- | --- | --- |
| 大基数候选截断语义 | 非阻断风险 | 在 evaluator/scheduler 后续切片中定义候选上限、截断标记、总量估算、用户提示和 eval log 字段，补大基数回归。 |
| Workflow API 严格 DSL 400 回归 | 非阻断风险 | 固定回归坏 DSL 创建/更新返回 400 且不持久化，覆盖缺节点、悬空边、循环、非法节点类型和错误分支。 |
| 持久化失败一致性 | 非阻断风险 | 补事务边界、失败响应、审计降级、eval log/current/history 一致性和存储失败注入测试。 |
| 统一 `.gitattributes` | 非阻断风险 | 作为独立治理切片补齐 Markdown、Go、Vue、Shell、PowerShell、YAML、JSON 的换行和文本归一化策略。 |

## 5. P2 findx-agents 工作单

P2 先落协议、能力目录、安全模型、证据模型，再迁移采集/巡检工具。Categraf `exec` 默认禁用；Catpaw 授权衍生能力进入实现前必须补合规材料。

| Work Unit | 执行角色 | 写集边界 | 依赖 | 验收命令/用例 | 是否允许 Git 稳定切片 |
| --- | --- | --- | --- | --- | --- |
| P2-Agent-0 协议先行文档门禁 | `ops-diagnostician` + `doc-writer` + `qa-tester` | 仅协议设计文档、门禁说明和文档-only QA 报告；禁止修改 `api/`、`web/`、配置文件和运行产物 | `go-backend` 实现前置；必须先输出 FACP v1、Capability Catalog、ToolDefinition、Evidence Artifact、Config Distribution、Security Model，并通过文档-only QA；未通过前不得实现 config/reload/upgrade/inspect/diagnose/session/remediation | 文档-only QA 覆盖协议字段、错误码、权限、审计、脱敏、离线保护、配置分发、证据引用、安全边界和敏感信息扫描；Go/Vue 构建标记 `NOT_RUN` 并说明文档-only 原因 | 允许文档-only 稳定切片；所有 P2 代码切片必须后置 |
| P2-Agent-1 协议与结构化能力模型 | `ops-diagnostician` + `go-backend` | 设计文档、协议常量、API model/store/test；禁止写真实密钥和生产路径 | 依赖 P0 agent register/heartbeat 和 P2-Agent-0；需明确 register、heartbeat、config-pull、capabilities、reload、upgrade 的请求/响应/错误码 | `cd /opt/ai-workbench/api && go test ./... && go build -o api-linux .`；用例覆盖能力目录上报、版本兼容、未知能力拒绝、脱敏响应 | 允许。建议切片名：`findx-p2-agent-protocol-capability` |
| P2-Agent-2 control plane/session/evidence | `go-backend` | 仅 `api/` 内 agent control、session、evidence model/store/handler/test；禁止前端并行改同一契约 | 依赖 P2-Agent-1；需要先定义 session 状态机、证据引用、超时、审计和权限 | `cd /opt/ai-workbench/api && go test ./... && go build -o api-linux .`；用例覆盖 session start/output/error/timeout、Agent 离线、evidence refs 不丢失、越权 403 | 允许。建议切片名：`findx-p2-agent-session-evidence` |
| P2-Agent-3 Categraf 插件融合与配置分发 | `go-backend` + `ops-diagnostician` | agent 配置模板、插件目录、配置分发 API/test；禁止启用高风险 exec 默认能力 | 依赖 P2-Agent-1、P2-Agent-2；需要配置 hash/version、reload 幂等、失败回滚 | `cd /opt/ai-workbench/api && go test ./... && go build -o api-linux .`；用例覆盖 config-pull、hash 未变不 reload、配置非法拒绝、离线保护、exec 默认禁用 | 允许，但必须独立于 Catpaw 衍生切片 |
| P2-Agent-4 Catpaw 衍生 inspector/tool registry | `ops-diagnostician` + `go-backend` | inspector/tool registry 设计、授权衍生说明、API model/store/test；禁止复制未确认授权边界的实现 | 依赖 P2-Agent-2；合规 NOTICE、来源说明、修改说明必须先完成 | `cd /opt/ai-workbench/api && go test ./... && go build -o api-linux .`；用例覆盖 tool registry 权限、inspect run 成功/失败/超时、artifact 脱敏保存、Agent 离线 | 允许，但必须与合规 NOTICE 同切片或后置于合规切片 |
| P2-Agent-5 安全审批/离线保护/合规 NOTICE | `ops-diagnostician` + `doc-writer` + `qa-tester` | 仅授权文档、NOTICE、测试报告和必要安全策略说明；禁止修改业务代码，除非另派 `go-backend` | 依赖 P2-Agent-1 到 P2-Agent-4 的协议与授权边界 | 文档审查用例：敏感信息扫描、授权来源完整、审批链路完整、离线保护策略完整；代码相关时补 WSL Go 构建 | 允许。建议作为 P2 合规门禁切片 |

## 6. P3 AI 与修复执行链路工作单

P3 必须先建立 evidence chain，再进入 remediation plan/precheck/dry-run/approve/execute/verify/rollback。所有执行动作必须经过审批和审计，不允许任意命令透传。

| Work Unit | 执行角色 | 写集边界 | 依赖 | 验收命令/用例 | 是否允许 Git 稳定切片 |
| --- | --- | --- | --- | --- | --- |
| P3-AI-0 低风险修复模板与禁止动作清单 | `ops-diagnostician` + `doc-writer` + `qa-tester` | 仅修复模板、禁止动作清单、门禁说明和文档-only QA 报告；禁止修改 `api/`、`web/`、配置文件和运行产物 | remediation execute 实现前置；必须先定义允许动作、禁止动作、风险分级、审批策略、回滚边界、执行白名单，并明确 AI 禁止 raw command；未通过前不得实现 remediation execute | 文档-only QA 覆盖模板字段、风险分级、审批过期、执行白名单、禁止动作拒绝、回滚边界、审计字段、AI 输出约束和敏感信息扫描；Go/Vue 构建标记 `NOT_RUN` 并说明文档-only 原因 | 允许文档-only 稳定切片；execute 代码切片必须后置 |
| P3-AI-1 evidence chain | `go-backend` + `ops-diagnostician` | 仅 `api/` 内 evidence、diagnose、knowledge/runbook 引用模型和测试；必要设计文档另行授权 | 依赖 P1 事件闭环和 P2 session/evidence；需要证据引用模型、脱敏摘要、追溯关系 | `cd /opt/ai-workbench/api && go test ./... && go build -o api-linux .`；用例覆盖事件、规则、目标、查询结果、巡检结果、通知记录进入 evidence chain，AI 输出必须引用 evidence refs | 允许。建议切片名：`findx-p3-ai-evidence-chain` |
| P3-AI-2 remediation plan/precheck/dry-run/approve/execute/verify/rollback | `go-backend` + `qa-tester` + `ops-diagnostician` | 仅 `api/` 内 remediation plan/run/step/approval/audit/test；禁止实现任意命令透传，禁止写真实凭据 | 依赖 P3-AI-1、P3-AI-0 和 P2 Agent control/session；需复用已通过的低风险动作白名单、审批模型、回滚模型，且不得允许 AI 生成 raw command 直接执行 | `cd /opt/ai-workbench/api && go test ./... && go build -o api-linux .`；用例覆盖 plan 草稿、precheck 失败、dry-run 不改状态、approve 过期、execute 固定 step、verify 失败、rollback 绑定原 run、重复提交幂等、禁止动作拒绝 | 允许，但 execute 相关必须单独 QA 通过后再提交 |

## 7. 主代理评分审计闭环

主代理评分用于进入 Git 门禁前的审计，不替代 QA。总分低于 `98/100` 的维度必须给出优化动作；存在 P0/P1 阻断时，无论分数多少均退回归属子代理修正。

| 维度 | 权重 | 98 分基线 | 低于 98 的典型扣分项 | 优化动作 |
| --- | --- | --- | --- | --- |
| 功能正确性 | 30% | 正常、异常、边界、权限、断连和幂等路径均有证据 | tryrun 与正式写入边界不清；事件状态机漏非法迁移；scheduler 并发未证实；UI 只测正常路径 | 补充复现用例和失败路径；为状态机、fingerprint、scheduler、权限写单元或集成测试；QA 重新回归 |
| 代码质量 | 20% | 分层清楚，复用已有 helper，单函数和文件复杂度受控 | 跨 handler/service/store 调用；重复状态常量；大函数混合查询、写入、审计；公共抽象无两个调用点 | 回派拆分函数和模块；集中域常量；删除重复实现；公共抽象不达标时降为模块私有 |
| 安全治理 | 20% | 认证、授权、脱敏、审计、审批、回滚和失败清理完整 | 响应或日志可能泄露 token/DSN/Cookie；Agent 高风险能力默认开启；远程执行缺审批；审计缺 actor/trace_id | 敏感字段统一 mask；高风险能力默认关闭；补 approval/precheck；审计记录 actor、scope、trace_id、diff 摘要 |
| 验证闭环 | 20% | WSL 构建、Go 测试、前端构建、API/UI 回归按变更类型完成 | 只在 Windows 通过未同步 WSL；UI 变更未跑浏览器；API_CONTRACT_CHANGE 未 curl；DATA_CHANGE 未验证旧数据 | 同步到 `/opt/ai-workbench` 后执行命令；补 curl/MCP/Playwright 证据；契约和数据变更单独列验证矩阵 |
| 文档同步 | 10% | 计划、API 契约、测试报告、main-log 或讨论文档与代码一致 | 代码已变文档仍旧；测试基准无新增用例编号；Git 切片说明缺回滚和未覆盖项 | 回派 `doc-writer` 更新文档；QA 报告引用统一测试基准；提交说明补验证、回滚、未覆盖项 |

评分规则：

- `>= 98` 且无阻断项：可进入 Git 稳定切片门禁。
- `90-97`：必须列出扣分项、优化动作、责任角色和是否接受残余风险；主代理明确同意后才可进入 Git 门禁。
- `80-89`：需讨论或回派修正，不能直接提交功能切片。
- `< 80`：退回归属子代理修正。
- 存在 P0/P1 缺陷、P0/P1 RISK、敏感信息泄露、权限绕过、API_CONTRACT_CHANGE/DATA_CHANGE 未验证：直接阻断。

## 8. 实时 Git 稳定切片策略

稳定切片定义：

- 功能域清晰：一个 commit 只承载一个 work unit 或一个 work unit 内的明确子域。
- 写集清晰：后端、前端、文档、QA 报告尽量分开；同一契约需要同步提交时必须说明原因。
- 验证清晰：每个切片写明已执行命令、正常路径、异常路径、边界或权限路径、未覆盖项。
- 风险清晰：API_CONTRACT_CHANGE、DATA_CHANGE、依赖变更、安全审批和合规材料必须打标。
- 回滚清晰：说明 Git 文件回滚、数据库回滚、配置回滚、运行态回滚或人工处理边界。

提交策略：

- 文档-only 切片可不跑 Go/Vue 构建，但必须执行敏感信息扫描，标记 `NOT_RUN` 并说明原因。
- 后端切片必须同步到 `/opt/ai-workbench` 后执行 `go test ./...` 和 `go build -o api-linux .`。
- 前端切片必须同步到 `/opt/ai-workbench` 后执行 `npm run build`，UI 变更还需 MCP/Playwright 回归。
- QA FAIL、P0/P1 RISK、敏感信息风险、权限绕过、契约或数据变更未验证时不提交。
- P1-BE-4、P1-FE-2、P2-Agent-3、P3-AI-2 属于多域工作单，必须继续拆成更小稳定切片。

推荐切片顺序：

1. `findx-doc-master-plan-closure`
2. `findx-p1-be-domain-fingerprint`
3. `findx-p1-be-evaluator-tryrun`
4. `findx-p1-be-scheduler-event-closure`
5. `findx-p1-fe-monitor-workbench-api`
6. `findx-p1-fe-rules-events-dashboard`
7. `findx-p1-be-notification-permission-audit`
8. `findx-p2-agent-protocol-doc-gate`
9. `findx-p2-agent-protocol-capability`
10. `findx-p2-agent-session-evidence`
11. `findx-p2-agent-collector-config`
12. `findx-p2-agent-inspector-registry`
13. `findx-p2-agent-security-notice`
14. `findx-p3-ai-evidence-chain`
15. `findx-p3-remediation-policy-doc-gate`
16. `findx-p3-remediation-safe-loop`

## 9. 下一步立即执行顺序

1. 本轮先补文档闭环：同步 README、立项文档和本文档，明确 P1-BE-1/2/3 已完成并推送，以及 QA 非阻断风险关闭计划。
2. 由 `qa-tester` 按文档-only 口径检查路径边界、敏感信息、禁止短语、work unit 可执行性和测试基准引用需求。
3. P2 开发前置项：先执行 P2-Agent-0 协议先行文档门禁，输出 FACP v1、Capability Catalog、ToolDefinition、Evidence Artifact、Config Distribution、Security Model，并通过文档-only QA；未通过前不得派发任何 config/reload/upgrade/inspect/diagnose/session/remediation 实现。
4. P3 开发前置项：先执行 P3-AI-0 低风险修复模板与禁止动作清单，明确允许动作、禁止动作、风险分级、审批策略、回滚边界、执行白名单和 AI 禁止 raw command，并通过文档-only QA；未通过前不得实现 remediation execute。
5. 下一轮执行可从 P1-FE-1 `/monitor` 工作台与 API 封装开始，使用已稳定的 P1-BE-1/2/3 后端契约做真实 API 空态、错误态、权限态和列表筛选。
6. 或从 P1-BE-4 基座开始，优先拆 notification、silence、subscription、oncall、pipeline、permission、audit 的 API_CONTRACT_CHANGE、权限矩阵、审计字段和脱敏规则。
7. P1-FE-1 与 P1-BE-4 可以并行准备只读契约和 UI 骨架，但同一 API 契约未稳定前不得写静态假数据冒充联调。
8. QA 非阻断风险按本文件“P1 QA 非阻断风险关闭计划”逐项关闭；大基数候选截断、Workflow DSL 400、持久化失败一致性、统一 `.gitattributes` 必须进入后续 checklist。

## 10. 验证与验收基线

后端变更：

```bash
cd /opt/ai-workbench/api
go test ./...
go build -o api-linux .
```

前端变更：

```bash
cd /opt/ai-workbench/web
npm run build
```

UI 变更：

- 使用 MCP/Playwright 做真实浏览器验证。
- 覆盖正常、异常、边界、权限、断连路径。
- 不允许静态假数据冒充真实 API 联调结果。

API 变更：

- 使用 curl 或测试脚本验证状态码、响应结构、认证、输入校验、错误提示和脱敏。
- API_CONTRACT_CHANGE 必须列出旧契约、新契约、兼容性、前端调用点、curl 验证和文档同步项。

数据变更：

- DATA_CHANGE 必须列出影响表/字段、旧数据兼容、迁移需求、回滚方式、幂等性和受影响接口/页面。

文档变更：

- 确认命令、路径、端口、接口清单与代码一致。
- 文档-only 不跑构建时标记 `NOT_RUN`，说明原因，并完成敏感信息扫描。

## 11. 设计债检查

| 检查项 | 结论 |
| --- | --- |
| 重复逻辑 | 本次仅更新讨论总计划，未新增业务实现或重复执行入口。后续 P1/P2/P3 要求先搜再写，复用现有 handler/service/store/test 模式。 |
| 复杂度 | 已将后续实施拆成 P1-BE、P1-FE、P2-Agent、P3-AI 工作单，并对多域工作单要求继续拆分稳定切片。 |
| 边界 | 本次只补齐执行主计划中的 P2/P3 实施前门禁，实际写入仅 `D:\ai-workbench\discuss\findx_monitoring_execution_master_plan.md`；不修改 README、docs、api、web、.claude、.codex、配置文件或运行产物。 |
| 依赖 | 未新增 Go module、npm 包、外部服务或运行时依赖。 |
| 兼容 | P0 状态改为已完成并推送；P1-BE-1/2/3 状态改为已完成并推送；P1 剩余域与 P2/P3 均保持向后推进，不改变已完成接口事实。 |
| 测试 | 文档-only 变更不执行 Go/Vue 构建；需要 QA 文档审查和敏感信息扫描。代码切片必须按第 10 节执行。 |
| 回滚 | 文档变更可按 Git 文件版本回滚；后续代码切片必须在各自提交说明中补充运行态回滚策略。 |
| 遗留风险 | P1 Dashboard、模板、通知、权限、团队、值班、订阅、静默仍未实现；P2/P3 尚未实现；P1-BE-4、P1-FE-2、P2-Agent-3、P3-AI-2 范围较大，必须二次拆分。 |

## 12. 脱敏检查

| 检查项 | 结果 |
| --- | --- |
| 真实 token/cookie/DSN/SSH 私钥 | 未写入。 |
| 示例敏感值 | 统一使用 `<TOKEN>`、`<API_KEY>`、`<DB_DSN>`、`<LOGIN_USER>`、`<SSH_KEY>`、`<COOKIE>`。 |
| 响应、日志、审计要求 | 已明确不得回显敏感信息，Agent、AI 和 remediation 链路均要求脱敏。 |
| 高风险能力 | 已明确 Categraf `exec` 默认禁用，修复执行链路禁止任意命令透传。 |

## 13. 仍需补充

- 由 `qa-tester` 对本文档执行文档-only 审查，并输出 PASS/FAIL/BLOCKED/NOT_RUN/RISK。
- P1-FE-1 开始前，需要确认 `/monitor` 工作台首屏信息架构、真实 API 空态/错误态/权限态和浏览器回归路径。
- P1-BE-4 开始前，需要拆分 notification、silence、subscription、oncall、pipeline、permission、audit 的写集、API_CONTRACT_CHANGE、权限矩阵、审计字段和脱敏规则。
- QA 非阻断风险关闭计划：大基数候选截断语义、Workflow API 严格 DSL 400 回归、持久化失败一致性、统一 `.gitattributes`。
- P2 开发前置项：P2-Agent-0 必须先输出 FACP v1、Capability Catalog、ToolDefinition、Evidence Artifact、Config Distribution、Security Model，并通过文档-only QA；否则不得实现 config/reload/upgrade/inspect/diagnose/session/remediation。
- P3 开发前置项：P3-AI-0 必须先输出低风险修复模板与禁止动作清单，覆盖允许动作、禁止动作、风险分级、审批策略、回滚边界、执行白名单和 AI 禁止 raw command，并通过文档-only QA；否则不得实现 remediation execute。
- P2-Agent-4 开始前，需要补齐 Catpaw 授权衍生来源说明、修改说明、授权边界和 NOTICE。
- P3-AI-2 开始前，需要明确首批低风险动作白名单、审批过期策略和失败回滚边界。

## 14. 自评分建议

本文档自评分建议：**98/100**。

| 维度 | 分数 | 说明 | 低于 98 的优化动作 |
| --- | --- | --- | --- |
| 功能正确性 | 98 | P0 状态、P1-BE-1/2/3 完成状态、P1 剩余 work units、P2/P3 依赖和验收用例已明确。 | 暂无。 |
| 代码质量 | 98 | 文档-only 变更，不涉及业务代码；后续已要求分层、复用和复杂度控制。 | 暂无。 |
| 安全治理 | 98 | 已覆盖脱敏、权限、审批、审计、离线保护、exec 默认禁用和禁止任意命令透传。 | 暂无。 |
| 验证闭环 | 98 | 已列出文档-only、后端、前端、UI、API、数据变更的验证基线，并记录 P1-BE-3 QA 92/100 PASS、Windows/WSL `go test -count=1 ./...` 与 build 通过。 | 暂无。 |
| 文档同步 | 98 | 已补齐 P0 推送状态、P1-BE-1/2/3 提交号、P1 下一步顺序、主代理评分审计和 QA 非阻断风险关闭计划。 | 暂无。 |

建议结论：**通过文档切片，进入 QA 文档审查；QA 通过后可提交 `findx-doc-master-plan-closure` 稳定切片。**
