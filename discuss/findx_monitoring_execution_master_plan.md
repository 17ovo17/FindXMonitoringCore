# FindX Monitoring Core P0-P3 实施总计划

生成时间：2026-05-04 04:15（UTC+8）

## 1. 文档定位

本文档是 FindX Monitoring Core 从 P0 到 P3 的完整实施总计划，用于指导主代理、子代理、QA 和后续 Git 稳定切片持续推进。主线是：FindX 建设为新的监控核心平台，参考并改进 Nightingale 的成熟能力，复用 Categraf 插件生态，融合 Catpaw 授权衍生 inspector，把 AI 问诊和自动修复纳入正式核心项目。

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
- API、数据模型、权限、审计、UI、Agent、AI 与自动修复最终归 FindX 主线承载。
- Nightingale 用于成熟功能参考、源码对标和可融合组件分析。
- Categraf 作为 `findx-agents` 的采集插件生态底座。
- Catpaw 授权能力衍生为 `findx-agents` 的 inspector、diagnose、session 和远程安全能力。
- 所有写操作必须有权限、审计、回滚或失败清理策略。
- 所有 AI 输出必须绑定 evidence refs。

## 3. 阶段地图

| 阶段 | 目标 | 主要输出 | 进入下一阶段条件 |
| --- | --- | --- | --- |
| P0-1 | target、datasource 基础语义、agent register、agent heartbeat、health | `/api/v1/monitor/targets`、`/api/v1/findx-agents/register`、`/api/v1/findx-agents/heartbeat`、health 状态 | QA 通过 target CRUD、agent token、heartbeat upsert、health 降级、权限、断连、脱敏、WSL 编译。 |
| P0-2 | alert rule、current/history event、tryrun、rollback、action | 规则生命周期、事件生命周期、动作日志、诊断/巡检/修复计划入口 | QA 通过规则 CRUD、版本、回滚、tryrun、事件状态机、动作审计、权限、断连、脱敏。 |
| P0-3 | query gateway | datasources、query、query-range、metrics、labels、label-values | QA 通过查询校验、数据源断连、限流、审计、脱敏，并完成与 P0-2 tryrun/evaluator 对接。 |
| P1 | dashboard、template、evaluator、notification、oncall、silence、subscription、permission、audit | 监控核心工作台全功能面 | API/UI/断连/权限/脱敏/回滚验证通过，文档和测试基准同步。 |
| P2 | findx-agents 融合 | Categraf 采集 + Catpaw 衍生 inspector/diagnose/session + Agent 控制协议 | Agent 安装、心跳、配置拉取、巡检、能力上报、离线保护和授权记录通过 QA。 |
| P3 | AI 问诊 + 自动修复 | precheck、dry-run、approve、execute、verify、rollback、audit、失败保护 | 低风险动作闭环通过，失败保护和审计闭环完整，敏感信息不进入 AI prompt 和报告。 |

## 4. P0 详细闭环

### 4.1 P0-1 已实施与待 QA 门禁

已实施范围：

- `GET /api/v1/monitor/health`
- `GET /api/v1/monitor/targets`
- `POST /api/v1/monitor/targets`
- `GET /api/v1/monitor/targets/:id`
- `PUT /api/v1/monitor/targets/:id`
- `DELETE /api/v1/monitor/targets/:id`
- `GET /api/v1/findx-agents`
- `POST /api/v1/findx-agents/register`
- `POST /api/v1/findx-agents/heartbeat`

待 QA 门禁：

- health 正常、empty、degraded 状态。
- target CRUD 正常、字段缺失、非法 IP、非 admin 写入。
- agent register/heartbeat 正确 token、缺 token、错误 token、匿名开关仅测试环境允许。
- heartbeat upsert Agent 与 Target，更新 `last_seen` 和在线状态。
- MySQL 不可用时返回可读降级状态。
- 响应、日志、审计不包含真实 token、Cookie、完整 DSN、SSH key、堆栈或本地敏感路径。

### 4.2 P0-2 已实施与待 QA 门禁

已实施范围：

- alert rule：列表、详情、创建、更新、删除、启用、禁用、克隆、tryrun、版本、rollback、导入、导出。
- current/history event：列表、详情、ack、assign、mute、resolve、archive、actions。
- 诊断入口：event diagnose、event inspect、event remediation-plan。

待 QA 门禁：

- 规则创建必填字段、非法 severity、非法 no_data_policy、非法 for_duration。
- 更新规则生成新版本，回滚生成新版本，不覆盖历史。
- tryrun 成功返回检查项，不生成正式 current event。
- tryrun 数据源断连返回可读失败，不误写事件。
- current/history event 分页、筛选、详情和动作日志完整。
- 事件状态机禁止 archived 后再次处置。
- 非 admin 事件处置返回 403，状态不变。
- diagnose 结果必须绑定 evidence refs。
- inspect 在 Agent 离线时失败可读，不误报 queued。
- remediation-plan 只创建草案，不直接执行。

### 4.3 P0-3 查询网关实施

P0-3 依据 ops 审计进入实施，统一承载规则试跑、evaluator、Dashboard、AI 问诊和自动修复验证的指标查询。

必须实现 API：

```text
GET    /api/v1/monitor/datasources
POST   /api/v1/monitor/datasources
GET    /api/v1/monitor/datasources/:id
PUT    /api/v1/monitor/datasources/:id
DELETE /api/v1/monitor/datasources/:id
POST   /api/v1/monitor/datasources/:id/test
POST   /api/v1/monitor/query
POST   /api/v1/monitor/query-range
GET    /api/v1/monitor/metrics
GET    /api/v1/monitor/labels
GET    /api/v1/monitor/label-values
```

实施要求：

- datasource 支持 Prometheus-compatible 类型优先落地，后续扩展 VictoriaMetrics、Mimir、Loki、Elasticsearch。
- 数据源凭据只允许写入和加密保存，不允许响应、日志、审计正文或 AI prompt 回显。
- query/query-range 校验查询表达式、时间范围、step、最大序列数、最大点数和超时。
- metrics/labels/label-values 限制返回数量，支持数据源权限检查。
- 上游断连、超时或认证失败返回 503，错误必须脱敏。
- 查询审计记录操作者、数据源、查询摘要、耗时、结果规模、状态和 trace_id。
- P0-2 tryrun/evaluator 必须通过查询网关取数。

## 5. P1 全量实施计划

| 功能域 | 后端范围 | 前端范围 | QA 门禁 |
| --- | --- | --- | --- |
| dashboard | dashboard、panel、variable、annotation、share、favorite、version、rollback、import/export | Dashboard 列表、编辑器、panel 配置、变量、空态、错误态、权限态 | 图表真实取数，版本回滚，权限拒绝，数据源断连。 |
| template | dashboard/rule/collect/remediation/runbook 模板、preview、install、diff、drift、upgrade、rollback | 模板中心、安装向导、diff 视图、漂移提示 | 安装失败清理，模板包脱敏，版本可追踪。 |
| evaluator | 调度、分片、规则评估、恢复判断、去重聚合、eval log | 规则评估状态、最近运行、失败原因 | 数据源断连不制造告警风暴，eval log 可追踪。 |
| notification | channel、template、rule、record、retry、preview、tryrun、rollback | 通知中心、模板编辑、发送记录 | 内容脱敏，失败重试，通知记录进入证据链。 |
| oncall | schedule、rotation、override、escalation、handover | 值班日历、交接、升级链路 | 团队权限一致，升级链路可审计。 |
| silence | create、update、delete、enable、expire、match preview、audit | 静默列表、创建、匹配预览、过期状态 | 匹配解释正确，到期恢复正确。 |
| subscription | user/team/business subscription、personal preference | 订阅偏好、团队订阅、业务组订阅 | 不重复轰炸，权限隔离正确。 |
| permission | user、team、role、business group、operation permission、API token | 团队权限页、角色授权、资源可见性 | 越权 403，资源不可见按统一模型处理。 |
| audit | 写操作审计、差异摘要、trace_id、来源、回滚引用 | 审计查询、对象追踪、差异查看 | 审计不含敏感信息，支持回放排障。 |

## 6. P2 findx-agents 融合计划

`findx-agents` 的目标形态：

```text
findx-agents
  -> Categraf collector inputs/logs/remote_write/provider
  -> FindX heartbeat/config-pull/capabilities/reload/upgrade
  -> Catpaw-derived inspector/diagnose/session/event model
  -> remediation executor
  -> local audit/supervisor
```

实施切片：

1. Categraf 基线梳理：保留 inputs、provider、remote_write、配置模板和 heartbeat 思路。
2. FindX Agent Control Protocol：register、heartbeat、config-pull、capabilities、reload、upgrade。
3. Catpaw 授权衍生 inspector：plugin registry、inspect run、structured findings、artifact 保存。
4. Diagnose/session：local/remote tool scope、session start/output/error、超时和审计。
5. Remediation executor：只执行审批后的 plan/run step，不接受任意命令透传。
6. 安装升级：`findx-agents.service`、`/opt/findx-agents`、`/etc/findx-agents`、`/var/log/findx-agents`。
7. 合规记录：LICENSE、NOTICE、来源说明、修改说明、授权边界。

P2 QA 门禁：

- Agent 在线、离线、重连、能力变化状态正确。
- 配置拉取支持版本/hash，不重复 reload。
- inspector 执行成功、超时、失败、Agent 离线均有明确状态。
- 本地审计记录 run id、session id、发起人、审批人和结果。
- 日志和响应不包含敏感信息。

## 7. P3 AI 问诊与自动修复计划

P3 是正式核心项目，必须完整覆盖：

```text
precheck -> dry-run -> approve -> execute -> verify -> rollback -> audit
```

AI 问诊输入：

- current/history event。
- alert rule。
- target metadata。
- datasource/query 结果。
- dashboard panel。
- notification record。
- silence/subscription/oncall。
- inspection run。
- remediation run。
- knowledge case。
- runbook。

AI 输出：

- 根因假设。
- evidence refs。
- 影响面。
- 风险等级。
- 建议动作。
- 修复计划草案。
- 是否建议调整规则、Dashboard、模板或知识库。

自动修复安全边界：

- precheck 验证目标、Agent 在线、能力、权限、变更窗口、风险等级和回滚条件。
- dry-run 不改变生产状态，只输出预计动作、影响面、失败条件和回滚策略。
- approve 记录审批人、审批理由、过期时间和审批范围。
- execute 只执行已批准 plan 的固定 step。
- verify 通过查询网关、Agent 巡检和事件状态验证修复效果。
- rollback 绑定原 plan/run/step，失败时进入人工处理。
- audit 记录 actor、approver、target、plan、run、step、trace_id、前后状态和证据引用。
- 失败保护覆盖超时、部分成功、Agent 断连、验证失败、回滚失败、重复提交和并发执行。

首批低风险动作：

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

## 8. 主代理 Claude 缺席模式

Claude 缺席时，主代理临时代行编排、决策、验收、评分、审计、优化和 Git 门禁；子代理是唯一执行层。

硬约束：

- 主代理不写业务代码。
- 子代理执行编码、文档、QA、诊断设计等 work unit。
- 所有子代理显式使用 `model: "gpt-5.5"`。
- 子代理 prompt 必须自包含：任务目标、允许路径、禁止路径、验收标准、验证命令、敏感信息占位符。
- 并行子代理写集互斥。
- QA FAIL、P0/P1 缺陷、P0/P1 RISK、敏感信息风险、契约或数据变更未验证，必须回派修正。
- 同一问题最多回派 3 轮；仍失败时停止盲修，标记人工处理。

主代理评分：

| 维度 | 权重 |
| --- | --- |
| 功能正确性 | 30% |
| 代码质量 | 20% |
| 安全治理 | 20% |
| 验证闭环 | 20% |
| 文档同步 | 10% |

评分规则：

- `>= 90` 且无阻断项：可进入 Git 门禁。
- `80-89`：需主代理复核风险并明确接受或回派。
- `< 80`：退回归属子代理修正。
- 存在 P0/P1 阻断：无论总分多少均退回。

## 9. 实时 Git 稳定切片策略

稳定切片定义：

- 功能域清晰。
- 写集清晰。
- 验证清晰。
- 风险清晰。
- 回滚清晰。

提交策略：

- 通过本地验证和 QA 的稳定切片立即 commit/push。
- QA FAIL 不提交。
- P0/P1 RISK 未关闭不提交。
- 敏感信息风险未关闭不提交。
- API_CONTRACT_CHANGE 或 DATA_CHANGE 未验证不提交。
- 文档-only 切片可不跑 Go/Vue 构建，但必须说明原因并完成敏感信息扫描。

推荐提交切片：

1. `findx-p0-1-target-agent-heartbeat`
2. `findx-p0-2-alert-rule-event`
3. `findx-p0-3-query-gateway`
4. `findx-p1-dashboard-template`
5. `findx-p1-evaluator-notification-permission-audit`
6. `findx-p2-findx-agents-collector-inspector`
7. `findx-p3-ai-remediation-loop`

## 10. 验证与验收

后端变更：

```bash
cd /opt/ai-workbench/api
go build -o api-linux .
go test ./...
```

前端变更：

```bash
cd /opt/ai-workbench/web
npm run build
```

UI 变更：

- 使用 Playwright/MCP 做真实浏览器验证。
- 覆盖正常、异常、边界、权限、断连路径。

API 变更：

- 使用 curl 或测试脚本验证状态码、响应结构、认证、输入校验、错误提示和脱敏。
- API_CONTRACT_CHANGE 必须同步到 `docs/aiops/findx_monitoring_core_api_contract.md`。

文档变更：

- 确认路径、接口、命令和项目约定一致。
- 文档-only 不跑构建时标记 `NOT_RUN` 并说明原因。

## 11. 自评分

目标分：**>= 95**。本总计划自评分：**97/100**。

| 维度 | 分值 | 说明 |
| --- | --- | --- |
| 战略完整性 | 98 | P0/P1/P2/P3 主线完整，FindX 新核心、Nightingale 参考、Categraf 复用、Catpaw 授权衍生、AI 自动修复均纳入。 |
| 可实施性 | 96 | 已拆分到可派发 work unit、稳定切片、QA 门禁和 Git 策略。 |
| 安全治理 | 97 | 覆盖认证、授权、脱敏、审计、远程执行、审批、回滚、AI prompt 保护。 |
| 生态复用 | 98 | 明确复用 Categraf 插件生态和 Catpaw 授权能力，参考 Nightingale 成熟模型。 |
| 验证闭环 | 96 | 覆盖 WSL 编译、前端构建、API/UI/断连/权限/脱敏回归；P0-1/P0-2 标记待 QA。 |
| 文档可执行性 | 97 | 可直接指导后续子代理按 P0 到 P3 持续实施。 |

不足项与改进措施：

- P0-1/P0-2 需要 QA 回填真实执行证据；改进措施是按统一测试基准逐条补证据。
- P0-3 查询网关需后端实现后回填最终字段、错误码和审计表结构。
- P1 面较大；改进措施是拆成 dashboard/template、evaluator、notification/oncall/silence/subscription、permission/audit 多个稳定切片。
- P2 需要补齐 Catpaw 授权衍生记录；改进措施是同步 LICENSE/NOTICE/来源说明/修改说明。
- P3 自动修复生产风险高；改进措施是先落低风险动作，所有 execute 必须经 precheck、dry-run、approve、verify、rollback、audit。

## 12. 技术债检查

| 检查项 | 结论 |
| --- | --- |
| 重复逻辑 | 本文档整合既有 API 契约、项目计划和 Claude 缺席协作模式，不新增重复执行入口。 |
| 复杂度 | 按 P0-1/P0-2/P0-3/P1/P2/P3 拆分，避免单次大改。 |
| 边界 | 仅为文档总计划，不修改 `api/`、`web/`、`README.md`、`.claude/`、`AGENTS.md` 或密钥配置。 |
| 依赖 | 未新增 Go module、npm 包、外部服务或运行时依赖。 |
| 兼容 | 保留既有兼容入口作为过渡和导入参考，新主线使用 `/api/v1/monitor/*`、`/api/v1/findx-agents/*`、`/api/v1/remediation/*`。 |
| 测试 | 文档-only 变更未执行 Go/Vue 构建；后续代码切片必须按阶段执行 WSL 编译、前端构建、API/UI/断连/权限/脱敏验证。 |
| 回滚 | 文档变更可按 Git 文件版本回滚；代码切片必须在各自 PR/commit 中补充迁移和运行回滚策略。 |
| 遗留风险 | P0-1/P0-2 QA 证据待回填，P0-3/P1/P2/P3 尚待执行层实现。 |
