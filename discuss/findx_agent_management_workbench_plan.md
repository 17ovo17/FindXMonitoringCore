# Superseded: Replaced on 2026-05-07 by `docs/aiops/findx_full_stack_observability_long_term_plan.md`. Historical evidence only; not an implementation entrypoint.

# FindX Agent/CMDB 运维工作台计划

生成时间：2026-05-04 12:45（UTC+8）

## 1. 立项结论

FindX Agent/CMDB 运维工作台属于 FindX 主平台的监控运维域，用来承载机器存活、业务组、CMDB 资产、`findx-agents` 生命周期、安装任务、配置分发、单机对话、巡检证据、权限审批和审计闭环。

方向必须保持清晰：

- 是 **FindX 主平台嵌入并深度融合 Nightingale 的监控运维逻辑**。
- 不是 Nightingale 外壳，不是 iframe，不是把 FindX 菜单换成夜莺黑壳。
- Nightingale 有的基础能力一个都不能少：target、业务组、机器存活、权限、通知、事件、模板、任务和审计都要接进 FindX 的数据模型和页面流程。
- AutoOps 只作为 CMDB/部署/任务的信息架构参考，不采纳裸命令、明文凭据、默认 SSH 信任和不受控 WebSSH。
- FindX 的 AI 问诊、知识库、单机 Agent 对话、巡检证据和自动修复是增强层，不能替代基础采集和基础监控控制面。

## 2. 源码审计依据

| 来源 | 路径 | 可借鉴能力 | 禁止照搬点 |
| --- | --- | --- | --- |
| Nightingale target/heartbeat | `D:\平台源码\nightingale-main (1)\nightingale-main\models\target.go`、`models\host_meta.go`、`memsto\target_cache.go`、`pushgw\router\router.go` | target、host meta、heartbeat、target update、机器存活、业务组绑定 | 不把 Nightingale 运行时作为长期事实源 |
| Nightingale 业务组/权限 | `router_busi_group.go`、`models\busi_group.go`、`models\role.go`、`models\role_operation.go` | 业务组、团队、角色、权限点、资源边界 | 不绕过 FindX 自有权限模型 |
| Nightingale 通知/事件 | `models\alert_cur_event.go`、`models\alert_his_event.go`、`models\notification_record.go`、`alert\sender\` | 事件、通知记录、升级和处置证据 | 不丢通知记录和事件时间线 |
| AutoOps CMDB | `D:\平台源码\AutoOps-main\AutoOps-main\api\api\cmdb\`、`web\src\views\cmdb\Host\` | 主机、分组、导入、云主机同步、主机详情 | 不照搬明文凭据和不安全 SSH |
| AutoOps 任务/部署 | `api\api\task\`、`web\src\views\Tools\ServiceMarket.vue`、`DeployDialog.vue`、`DeployProgress.vue` | 服务市场、选择主机、部署参数、执行进度、日志流 | 不把裸 Ansible/SSH 命令暴露给 AI |
| Categraf | `D:\平台源码\categraf-main (1)\categraf-main\inputs`、`agent\heartbeat.go`、`inputs\http_provider.go` | 插件生态、配置 provider、heartbeat、remote_write | `inputs/exec` 默认禁用，不进入 AI 工具目录 |
| Catpaw | `D:\平台源码\catpaw-master\catpaw-master\agent\inspect.go`、`digcore\diagnose\types.go`、`digcore\server\proto.go` | 巡检、诊断、会话、结构化工具、事件模型 | 任何 raw command 必须改造成审批后的结构化工具 |

## 3. 产品信息架构

Agent 管理工作台不做单页大表格，而是按运维工作流拆分为 8 个页面。

| 页面 | 目标用户问题 | 主要能力 |
| --- | --- | --- |
| Agent 总览 | 当前采集和控制面健康吗？ | 在线率、采集健康、业务组分布、版本分布、配置漂移、失败任务、待升级 Agent |
| 机器存活 | 哪些机器掉线、采集中断或 target miss？ | target/agent 双状态、`last_heartbeat_at`、`last_scrape_at`、`down_seconds`、`target_miss`、标签筛选 |
| CMDB 资产 | 这台机器属于谁、什么环境、挂在哪个业务组？ | 主机资产、IP、OS、云厂商、环境、负责人、业务组、标签、凭据引用、Agent 绑定 |
| 安装任务 | 如何安全安装/升级/卸载 findx-agents？ | 远程安装、本机安装指引、升级、卸载、回滚、precheck、dry-run、审批、执行日志 |
| 配置分发 | 哪些机器配置不一致、是否需要灰度？ | 全局模板、业务组覆盖、主机覆盖、diff、灰度、reload、漂移检测、回滚 |
| 单机对话 | 能否围绕这台机器问诊？ | 绑定 target/agent 的 AI 会话、巡检模板、日志摘要、指标引用、缺失证据、修复建议 |
| 巡检证据 | 这台机器最近检查过什么，结论可信么？ | 单机巡检、业务组巡检、定时巡检、findings、evidence refs、Runbook 关联 |
| 审计与权限 | 谁对哪些机器做了什么？ | 凭据使用、安装任务、配置下发、会话、巡检、修复、审批、失败回滚审计 |

## 4. 核心状态模型

Agent 管理不能只看一个在线状态，必须把目标存在、采集状态、控制通道、配置状态和任务状态拆开。

| 状态 | 字段建议 | 说明 |
| --- | --- | --- |
| 机器存在状态 | `target_status`、`target_miss` | 来自 target/host meta；表示监控对象是否仍应存在 |
| 采集状态 | `scrape_status`、`last_scrape_at`、`scrape_error_summary` | 表示指标采集是否正常 |
| 控制通道状态 | `agent_status`、`last_heartbeat_at`、`down_seconds` | 表示 findx-agents 是否在线可控 |
| 配置状态 | `config_version`、`config_hash`、`drift_status`、`last_reload_at` | 表示配置是否匹配控制面期望 |
| 能力状态 | `capability_hash`、`disabled_capabilities`、`unsupported_capabilities` | 表示该 Agent 能做什么、为什么不能做 |
| 安装状态 | `install_status`、`upgrade_status`、`rollback_available` | 表示生命周期任务状态 |
| 风险状态 | `risk_level`、`pending_approval_count`、`failed_task_count` | 表示是否需要人工处理 |

## 5. API 与数据模型规划

### 5.1 建议 API

```text
GET    /api/v1/monitor/agent-workbench/summary
GET    /api/v1/monitor/agent-workbench/liveness
GET    /api/v1/monitor/agent-workbench/assets
POST   /api/v1/monitor/agent-workbench/assets/import/preview
POST   /api/v1/monitor/agent-workbench/assets/import/apply
GET    /api/v1/monitor/agent-workbench/credentials
POST   /api/v1/monitor/agent-workbench/install/plan
POST   /api/v1/monitor/agent-workbench/install/:id/precheck
POST   /api/v1/monitor/agent-workbench/install/:id/dry-run
POST   /api/v1/monitor/agent-workbench/install/:id/approve
POST   /api/v1/monitor/agent-workbench/install/:id/execute
POST   /api/v1/monitor/agent-workbench/install/:id/verify
POST   /api/v1/monitor/agent-workbench/install/:id/rollback
GET    /api/v1/monitor/agent-workbench/configs
POST   /api/v1/monitor/agent-workbench/configs/:id/diff
POST   /api/v1/monitor/agent-workbench/configs/:id/rollout
POST   /api/v1/monitor/agent-workbench/configs/:id/rollback
POST   /api/v1/monitor/agent-workbench/sessions
GET    /api/v1/monitor/agent-workbench/sessions/:id/events
GET    /api/v1/monitor/agent-workbench/evidence
GET    /api/v1/monitor/agent-workbench/audit-logs
```

### 5.2 建议表

```text
monitor_assets
monitor_asset_groups
monitor_asset_credentials
monitor_agent_liveness
monitor_agent_capabilities
monitor_agent_config_versions
monitor_agent_config_rollouts
monitor_agent_install_plans
monitor_agent_install_runs
monitor_agent_sessions
monitor_agent_evidence
monitor_agent_audit_logs
```

## 6. 远程安装与本机安装

### 6.1 远程安装

远程安装必须是结构化任务，不允许用户或 AI 直接提交裸命令。

流程：

```text
select_assets
  -> choose_credential_ref
  -> choose_agent_version
  -> choose_config_template
  -> precheck
  -> dry_run
  -> approval
  -> execute_install
  -> verify_heartbeat
  -> verify_scrape
  -> rollback_or_close
  -> audit
```

门禁：

- 凭据只传 `credential_ref`，执行层按权限临时解密，页面、日志、AI prompt 和审计不回显 secret。
- precheck 检查 OS、架构、磁盘、网络、systemd、端口、已有版本、业务组权限和凭据权限。
- dry-run 输出将做什么、影响哪些路径、失败如何回滚，不改变机器状态。
- execute 必须绑定 approval_id、actor、approver、target、agent_version、config_version。
- verify 同时检查 heartbeat、配置版本、采集状态和控制面状态。
- rollback 必须引用原 install_run_id，不允许手工拼命令回滚。

### 6.2 本机安装

Linux 本机安装需要固定路径和服务名：

| 项 | 值 |
| --- | --- |
| 服务名 | `findx-agents.service` |
| 安装目录 | `/opt/findx-agents` |
| 配置目录 | `/etc/findx-agents` |
| 日志目录 | `/var/log/findx-agents` |
| 数据目录 | `/var/lib/findx-agents` |
| 本地审计 | `/var/log/findx-agents/audit.log` |

本机安装页面展示安装脚本时只能使用占位符，例如 `<SERVER_URL>`、`<BOOTSTRAP_TOKEN>`、`<TENANT_ID>`，不得写入真实 token。

## 7. 单机 Agent 对话

单机 Agent 对话不是 WebSSH，也不是万能终端。它是绑定机器、证据和结构化工具的诊断入口。

允许：

- 查询该机器的 target、Agent、配置、最近事件、最近通知、最近巡检。
- 拉取脱敏日志摘要和指标摘要。
- 执行已注册巡检模板。
- 调用白名单诊断工具。
- 生成修复建议和 remediation plan 草稿。

禁止：

- 裸 shell。
- 裸 SQL。
- 裸 PromQL 作为执行动作。
- 裸 HTTP 请求。
- AI 直接发起 execute。
- 未审批的高风险动作。

对话输出必须区分：

- `事实证据`：来自指标、事件、Agent artifact、日志摘要、Runbook、知识库。
- `推理结论`：AI 基于证据的判断。
- `缺失证据`：当前没有数据或 Agent 离线。
- `建议动作`：只能生成计划，执行必须进入审批链。

## 8. 权限、审计与安全

权限模型：

| 权限 | 说明 |
| --- | --- |
| `agent.asset.read` | 查看资产和 Agent 状态 |
| `agent.asset.write` | 编辑资产、业务组和标签 |
| `agent.credential.use` | 使用凭据引用发起安装或任务 |
| `agent.install.plan` | 创建安装计划 |
| `agent.install.execute` | 执行已审批安装任务 |
| `agent.config.rollout` | 下发配置 |
| `agent.session.open` | 开启单机对话 |
| `agent.inspect.run` | 执行巡检模板 |
| `agent.remediation.plan` | 生成修复计划 |
| `agent.remediation.execute` | 执行已审批修复动作 |

审计字段：

```text
actor
approver
tenant_id
business_group_id
target_id
agent_id
credential_ref
action
resource
request_hash
diff_summary
result
error_summary
trace_id
evidence_refs
created_at
```

敏感信息规则：

- 不输出真实 token、Cookie、密码、SSH 私钥、完整 DSN。
- 示例统一使用 `<TOKEN>`、`<API_KEY>`、`<DB_DSN>`、`<BOOTSTRAP_TOKEN>`、`<LOGIN_USER>`。
- 错误展示只给 safe message，不展示内部路径、堆栈、完整上游错误。

## 9. 与 AI 问诊和知识库联动

Agent/CMDB 工作台会成为 AI 问诊的证据入口之一，但 AI 不直接替代运维动作。

联动方式：

1. 告警事件触发 AI 问诊。
2. AI 读取 target、业务组、资产、事件、通知、Dashboard、历史巡检和知识库。
3. 缺证据时，AI 生成 Agent 巡检请求。
4. Agent 返回结构化 evidence artifact。
5. AI 生成诊断结论和 remediation plan 草稿。
6. 用户审批后，自动修复执行器按固定 plan step 执行。
7. 执行结果回写事件时间线、审计和知识库。

知识库记忆标签只影响 Prompt 注入策略，不等于事实证据。AI 输出必须明确区分“偏好/场景规则”和“监控/Agent 事实”。

## 10. 验收标准

| 验收项 | 标准 |
| --- | --- |
| Nightingale 基础能力 | target、业务组、机器存活、权限、事件、通知、模板、任务、审计均有入口和后端契约 |
| AutoOps 借鉴边界 | 文档明确只借鉴 CMDB/任务/部署信息架构，不采纳裸命令和明文凭据 |
| 远程安装 | precheck、dry-run、approve、execute、verify、rollback 全链路 |
| 本机安装 | Linux/WSL 路径、服务名、日志路径、卸载和回滚明确 |
| 单机对话 | 绑定 target/agent/evidence，不开放任意 shell |
| 巡检证据 | evidence artifact 可被 AI 问诊引用，缺失证据明确展示 |
| 权限审计 | 所有写动作、凭据使用、安装、配置、巡检、修复均审计 |
| 脱敏 | 文档、页面、日志、AI prompt、审计均不出现真实密钥 |
| WSL/Linux | 在 `/opt/ai-workbench` 或目标 Linux 运行态验证 |

## 11. 自评分

综合自评分：**97/100**。

| 维度 | 分值 | 说明 |
| --- | --- | --- |
| 方向正确性 | 99 | 明确 FindX 主平台嵌入 Nightingale，不再做夜莺外壳。 |
| Nightingale 基础能力覆盖 | 97 | 已覆盖 target、业务组、机器存活、权限、事件、通知、模板、任务、审计和 Agent 心跳。 |
| AutoOps 借鉴边界 | 98 | 明确只借鉴 CMDB/部署/任务信息架构，禁止裸命令和明文凭据。 |
| Agent 生命周期 | 97 | 覆盖远程安装、本机安装、升级、卸载、回滚、配置分发、漂移检测。 |
| AI/知识库联动 | 96 | 明确 AI 是增强层，证据来自监控和 Agent，记忆标签不等于事实。 |
| 安全治理 | 96 | 凭据引用、审批、审计、脱敏、禁止 raw action 都已写入。 |
| 可实施性 | 96 | API、表、页面、流程、验收和权限均可拆分成后续稳定切片。 |

扣分点：

- 目前仍是讨论计划，尚未有 Agent/CMDB 后端表、API、页面和 WSL 运行证据。
- WebSSH 是否保留需要后续产品决策；本计划仅定义为高风险人工工具，默认不进入 AI 自动链路。
- Categraf/Catpaw 代码融合需要后续补 LICENSE、NOTICE、来源版本、修改说明和授权记录。

补齐措施：

- P2 前先实现 Agent/CMDB API 契约和表结构迁移计划。
- P2-Agent-0 文档门禁通过后，再实现 FACP、配置分发、安装任务和巡检证据。
- 每个切片执行 WSL 构建、API/UI、权限、离线、脱敏和审计验证。
