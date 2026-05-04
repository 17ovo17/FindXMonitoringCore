# FindX 全栈可观测与 AI SRE — P0~P6 全量实施计划

> 文档性质：全阶段实施规划
> 当前状态：规划中，不代表当前已经实现
> 本轮范围：只改写实施计划文档，不实施业务代码、不修改 README、不提交、不推送
> 日期：2026-05-04 23:55（UTC+8）
> 参考来源：成熟开源监控平台设计、主机巡检与 AI SRE 实践、现有 FindX 项目架构

## 1. 总览

本计划围绕 **FindX Core**、**findx-agents**、**AI SRE** 三条主线展开。FindX Core 负责可观测基础能力、资源治理、告警通知、团队权限和平台运行自检；findx-agents 负责主机接入、采集插件、配置下发、远程安装、调试、自升级、命令执行、自愈和单机巡检入口；AI SRE 负责主机健康检查、单机巡检、AI 问诊、RCA、证据链、自动修复和复盘闭环。

| 阶段 | 主题 | 核心交付 |
|---|---|---|
| P0 | FindX Core 基础监控可信基座 | 业务空间、资源组、主机资产、监控仪表盘、资产总览 |
| P1 | 告警响应与通知闭环 | 通知渠道、消息模板、通知策略、通知记录、静默、订阅、事件流水线基础 |
| P2 | 团队权限与资源治理 | 用户、团队、角色、权限矩阵、API Token、审计日志、业务空间授权 |
| P3 | findx-agents 控制面 | Agent 注册心跳、能力目录、Categraf 插件目录、采集配置中心化下发、配置版本 diff/回滚、远程安装、本机安装指令、凭证引用、远程调试、自升级、结构化命令执行、告警自愈、findx-agents inspector 单机巡检、诊断会话、证据产物、审计脱敏 |
| P4 | 多信号数据与可观测底座 | 日志、Trace/APM、RUM、指标目录、智能告警、保存视图、多数据源 |
| P5 | AI SRE 与证据链 | AI 主机健康检查、单机巡检、AI 问诊、RCA、Evidence Chain、自动修复、复盘报告 |
| P6 | 商业化与治理增强 | Status Page、on-call 深化、SSO、合规、容量成本、模板市场 |

## 2. 全局命名与边界

### 2.1 产品命名

- 产品、页面、API、模型、表名、任务名统一使用 FindX 风格。
- 平台能力分为三类：FindX Core、findx-agents、AI SRE。
- 业务归属使用 **业务空间**，资产归类使用 **资源组**，组织协作使用 **团队、角色、权限矩阵**。
- 仪表盘能力统一命名为 **监控仪表盘**。
- Agent 能力统一归入 **findx-agents 控制面**。
- 健康检查归入 **AI SRE**，不作为平台管理模块的主机检查能力。
- 可以在设计说明中抽象引用“成熟开源监控平台设计”，不得把外部产品名作为 FindX 模块名、API 名、表名或任务名。
- 单机巡检可以吸收 Catpaw 授权能力/思想，但 Catpaw 仅作为内部能力来源说明，不作为用户侧产品命名，不进入导航、页面标题、API 分组、模型名、表名、权限对象和审计对象。

### 2.2 API 命名风格

后续实施建议使用如下 FindX 风格 API 前缀：

```text
/api/v1/workspaces/*
/api/v1/resource-groups/*
/api/v1/host-assets/*
/api/v1/monitor/dashboards/*
/api/v1/notifications/*
/api/v1/teams/*
/api/v1/roles/*
/api/v1/permissions/*
/api/v1/audit-logs/*
/api/v1/findx-agents/*
/api/v1/aiops/host-health/*
/api/v1/evidence/*
/api/v1/remediation/*
/api/v1/platform/self-check/*
```

### 2.3 健康检查归属

**AI 主机健康检查 / 单机巡检** 归入 AI SRE。

输入范围：

- 主机
- Agent
- 指标
- 日志
- Trace
- 巡检工具
- 采集配置快照
- 历史告警与变更记录

输出范围：

- `evidence_refs`
- 风险等级
- 异常摘要
- 建议动作
- 修复候选
- 需人工确认项

平台管理只保留 **平台运行自检**，用于检查 FindX 自身组件、任务队列、数据库连通性、缓存、模型配置、数据源连通性和后台任务状态。平台管理不承载主机健康检查、单机巡检或 AI 健康诊断。

## 3. 工程边界

### 3.1 后端分层

```text
api/
  main.go
  internal/
    model/
    store/
    handler/
    workflow/
    reasoning/
    timeseries/
    embedding/
```

实施业务代码时遵守 `model -> store -> handler -> route` 的最小闭环。新增公共抽象必须有真实调用点，不为了规划提前堆通用工具包。

### 3.2 前端分层

```text
web/src/
  router/
  views/
  components/
  api/
  composables/
```

页面入口围绕工作台组织，组件只承载局部交互，API 封装不保存业务状态，路由按页面懒加载接入。

### 3.3 findx-agents 控制面分层

```text
findx-agents 控制面
  Agent 注册与心跳
  Agent 能力目录
  Categraf 插件目录
  采集配置中心
  配置版本 diff 与回滚
  远程安装任务
  本机安装指令
  凭证引用
  远程调试
  自升级
  结构化命令执行
  告警自愈入口
  findx-agents inspector 单机巡检
  诊断会话
  证据产物
  审计脱敏
```

### 3.4 敏感信息边界

- 示例不得写真实密钥、Cookie、Token、DSN、SSH 私钥或完整连接串。
- 示例统一使用 `<TOKEN>`、`<API_KEY>`、`<DB_DSN>`、`<LOGIN_USER>`、`<SSH_KEY_REF>` 等占位符。
- 凭证只允许以引用方式出现在任务、审计和配置中，不展示明文。
- 远程执行、调试、自升级、自动修复、通知测试必须记录审计日志并脱敏输出。

## 4. P0：FindX Core 基础监控可信基座

> 目标：建立可被后续告警、权限、Agent、AI SRE 复用的基础监控数据和页面骨架。

### P0-T1：业务空间

**场景**：按业务域组织资源、指标、告警、团队和权限边界。

**建议 API**：

```text
GET    /api/v1/workspaces
POST   /api/v1/workspaces
GET    /api/v1/workspaces/:id
PUT    /api/v1/workspaces/:id
DELETE /api/v1/workspaces/:id
POST   /api/v1/workspaces/:id/members
DELETE /api/v1/workspaces/:id/members/:member_id
```

**建议模型**：

| 模型 | 字段方向 |
|---|---|
| Workspace | 名称、描述、负责人、标签、状态、创建时间、更新时间 |
| WorkspaceMember | 业务空间、用户、团队、角色、加入时间 |
| WorkspaceBinding | 业务空间与资源组、主机资产、监控仪表盘、通知策略的绑定关系 |

**验收**：

- 可创建、编辑、删除、查看业务空间。
- 成员引用不回显敏感信息。
- 后续资源、告警、仪表盘、AI SRE 任务可绑定业务空间。
- 删除前校验资源引用关系。

### P0-T2：资源组与主机资产

**场景**：管理主机资产、资源组归属、标签、Agent 状态、业务空间归属。

**建议 API**：

```text
GET    /api/v1/resource-groups
POST   /api/v1/resource-groups
PUT    /api/v1/resource-groups/:id
DELETE /api/v1/resource-groups/:id
GET    /api/v1/host-assets
GET    /api/v1/host-assets/:id
PUT    /api/v1/host-assets/:id/tags
PUT    /api/v1/host-assets/:id/resource-group
PUT    /api/v1/host-assets/:id/workspace
```

**主机资产字段方向**：

| 字段 | 说明 |
|---|---|
| host_id | FindX 内部主机 ID |
| hostname | 主机名 |
| ip_list | IP 列表 |
| os | 操作系统 |
| arch | 架构 |
| workspace_id | 业务空间 |
| resource_group_id | 资源组 |
| tags | 标签 |
| agent_id | 关联 Agent |
| agent_status | Agent 在线状态 |
| last_seen_at | 最近心跳时间 |

**验收**：

- 支持按业务空间、资源组、标签、在线状态过滤。
- 主机资产可与 findx-agents 心跳关联。
- 离线主机不触发误删，只更新状态与时间。

### P0-T3：监控仪表盘

**场景**：创建、查看、编辑、克隆监控仪表盘，支撑日常巡检和资产总览。

**建议 API**：

```text
GET    /api/v1/monitor/dashboards
GET    /api/v1/monitor/dashboards/:id
POST   /api/v1/monitor/dashboards
PUT    /api/v1/monitor/dashboards/:id
DELETE /api/v1/monitor/dashboards/:id
POST   /api/v1/monitor/dashboards/:id/clone
POST   /api/v1/monitor/dashboards/:id/share
```

**配置示例**：

```json
{
  "title": "Linux 主机基础监控",
  "workspace_id": "ws-001",
  "resource_group_id": "rg-001",
  "panels": [
    {
      "id": "panel-1",
      "type": "timeseries",
      "title": "CPU 使用率",
      "datasource_id": "ds-1",
      "query": "100 - (avg by(instance)(rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
      "position": { "x": 0, "y": 0, "w": 12, "h": 8 }
    }
  ],
  "variables": [],
  "time_range": { "from": "now-1h", "to": "now" }
}
```

**验收**：

- 创建、编辑、删除、克隆、查看均可用。
- 查询失败时展示用户可理解错误，不暴露内部路径、SQL、完整连接串。
- 仪表盘可绑定业务空间和资源组。

### P0-T4：内置监控仪表盘模板

**范围**：

- Linux 主机基础模板
- Windows 主机基础模板
- MySQL 基础模板
- Redis 基础模板
- Kubernetes 节点基础模板

**验收**：

- 模板可预览、导入、克隆。
- 导入时可选择业务空间和资源组。
- 模板变量能与主机资产标签、资源组绑定。

### P0-T5：资产总览

**建议 API**：

```text
GET /api/v1/assets/overview
```

**响应示例**：

```json
{
  "hosts": { "total": 120, "online": 115, "offline": 5 },
  "agents": { "total": 118, "alive": 112, "dead": 6 },
  "dashboards": { "total": 24 },
  "alerts": { "firing": 3, "acknowledged": 1 },
  "workspaces": { "total": 8 },
  "resource_groups": { "total": 18 }
}
```

**验收**：

- 总览数据可按业务空间过滤。
- Agent 状态来自 findx-agents 心跳或未来兼容数据。
- 资产总览不执行主机健康检查，只展示资产和状态汇总。

## 5. P1：告警响应与通知闭环

> 目标：告警触发后可以被正确路由、通知、记录、静默、订阅、确认和升级，为后续 AI SRE 自愈留出入口。

### P1-T1：通知渠道

**建议 API**：

```text
GET    /api/v1/notifications/channels
POST   /api/v1/notifications/channels
PUT    /api/v1/notifications/channels/:id
DELETE /api/v1/notifications/channels/:id
POST   /api/v1/notifications/channels/:id/test
```

**能力范围**：

- 邮件、Webhook、即时消息、短信等渠道抽象。
- 凭证字段只保存引用或加密后的密文，不回显明文。
- 测试发送写入审计日志。

**验收**：

- 渠道创建、编辑、禁用、删除可用。
- `token`、`secret`、`password`、`webhook` 等字段脱敏返回。
- 示例只使用 `<TOKEN>`、`<API_KEY>`。

### P1-T2：消息模板

**建议 API**：

```text
GET    /api/v1/notifications/templates
POST   /api/v1/notifications/templates
PUT    /api/v1/notifications/templates/:id
DELETE /api/v1/notifications/templates/:id
POST   /api/v1/notifications/templates/:id/preview
```

**能力范围**：

- 模板变量。
- 渠道类型适配。
- 预览渲染。
- 版本记录。
- 默认模板。

**验收**：

- 预览结果不泄露内部路径、SQL、完整连接串或认证票据。
- 模板变量缺失时返回 4xx 用户输入错误。

### P1-T3：通知策略

**建议 API**：

```text
GET    /api/v1/notifications/policies
POST   /api/v1/notifications/policies
PUT    /api/v1/notifications/policies/:id
DELETE /api/v1/notifications/policies/:id
POST   /api/v1/notifications/policies/:id/try-run
```

**能力范围**：

- 按业务空间、资源组、标签、告警等级、时间窗口匹配。
- 按团队、角色、用户、排班、升级策略投递。
- 支持策略优先级、启停状态、dry-run。

**验收**：

- 策略 dry-run 可返回匹配原因和预计接收方。
- 接收方展示脱敏信息。
- 删除前校验是否被告警规则或订阅引用。

### P1-T4：通知记录

**建议 API**：

```text
GET /api/v1/notifications/records
GET /api/v1/notifications/records/:id
POST /api/v1/notifications/records/:id/retry
```

**能力范围**：

- 投递状态。
- 失败原因。
- 重试次数。
- 渠道。
- 接收方脱敏展示。
- 关联告警事件。

**验收**：

- 通知失败可追踪原因。
- 重试动作写入审计日志。
- 响应中不出现真实凭证。

### P1-T5：静默与订阅

**建议 API**：

```text
GET    /api/v1/alerts/silences
POST   /api/v1/alerts/silences
PUT    /api/v1/alerts/silences/:id
DELETE /api/v1/alerts/silences/:id
GET    /api/v1/alerts/subscriptions
POST   /api/v1/alerts/subscriptions
DELETE /api/v1/alerts/subscriptions/:id
```

**能力范围**：

- 按业务空间、资源组、标签、告警等级、时间窗口静默。
- 按团队、角色、用户订阅。
- 静默到期自动恢复。

**验收**：

- 静默只影响匹配范围，不吞掉事件流水。
- 订阅可与通知策略联动。
- 过期静默不继续生效。

### P1-T6：事件流水线基础

**建议 API**：

```text
GET  /api/v1/alerts/events
GET  /api/v1/alerts/events/:id
POST /api/v1/alerts/events/:id/ack
POST /api/v1/alerts/events/:id/close
POST /api/v1/alerts/events/:id/aiops-diagnosis
```

**能力范围**：

- 告警事件入库。
- 状态流转。
- 确认、关闭、处理备注。
- 通知联动。
- AI SRE 诊断入口占位。

**验收**：

- 告警事件状态流转有审计记录。
- AI SRE 入口只创建诊断请求，不在 P1 实现完整 AI 诊断。

## 6. P2：团队权限与资源治理

> 目标：补齐多团队协作、权限矩阵、API Token、审计日志和业务空间授权，为 P3 findx-agents 控制面建立安全边界。

### P2-T1：用户、团队、角色

**建议 API**：

```text
GET    /api/v1/users
POST   /api/v1/users
PUT    /api/v1/users/:id
DELETE /api/v1/users/:id
GET    /api/v1/teams
POST   /api/v1/teams
PUT    /api/v1/teams/:id
DELETE /api/v1/teams/:id
GET    /api/v1/roles
POST   /api/v1/roles
PUT    /api/v1/roles/:id
DELETE /api/v1/roles/:id
```

**能力范围**：

- 用户基础资料。
- 团队成员管理。
- 角色定义。
- 团队与业务空间绑定。
- 角色与权限矩阵绑定。

**验收**：

- 删除前校验引用关系。
- 用户信息不回显敏感凭证。
- 团队可以被通知策略、业务空间授权和审计过滤复用。

### P2-T2：权限矩阵

**建议 API**：

```text
GET  /api/v1/permissions/matrix
PUT  /api/v1/permissions/matrix
POST /api/v1/permissions/check
```

**资源维度**：

- 业务空间
- 资源组
- 主机资产
- 监控仪表盘
- 告警事件
- 通知渠道
- 通知策略
- 团队
- 角色
- findx-agents
- AI SRE
- 证据链
- 自动修复
- 平台运行自检

**动作维度**：

- read
- write
- execute
- approve
- admin

**验收**：

- 权限检查能覆盖读、写、执行、审批、管理。
- findx-agents 的远程安装、调试、命令执行、自升级必须要求 execute 或 approve。
- AI SRE 自动修复必须要求 approve。

### P2-T3：业务空间授权

**建议 API**：

```text
GET    /api/v1/workspaces/:id/authorizations
POST   /api/v1/workspaces/:id/authorizations
DELETE /api/v1/workspaces/:id/authorizations/:auth_id
```

**能力范围**：

- 授权团队访问业务空间。
- 授权角色操作资源组、主机资产、监控仪表盘、通知策略。
- 授权结果可被后端权限中间件复用。

**验收**：

- 非授权用户不能查看业务空间下的主机、仪表盘、告警和证据。
- 授权变更写入审计日志。

### P2-T4：API Token

**建议 API**：

```text
GET    /api/v1/api-tokens
POST   /api/v1/api-tokens
DELETE /api/v1/api-tokens/:id
POST   /api/v1/api-tokens/:id/rotate
```

**要求**：

- Token 只在创建时展示一次。
- 存储哈希。
- 支持过期时间、作用域、业务空间限制。
- 轮换动作写审计日志。

**验收**：

- 列表和详情不回显真实 Token。
- 示例统一使用 `<TOKEN>`。
- 过期 Token 无法访问 API。

### P2-T5：审计日志

**建议 API**：

```text
GET /api/v1/audit-logs
GET /api/v1/audit-logs/:id
```

**审计范围**：

- 用户与团队变更。
- 权限矩阵变更。
- 业务空间授权变更。
- 通知渠道与策略变更。
- findx-agents 远程安装、调试、命令执行、自升级。
- AI SRE 诊断、证据读取、自动修复审批与执行。
- 平台运行自检。

**验收**：

- 审计日志可按用户、团队、业务空间、资源、动作、结果过滤。
- 审计详情脱敏展示。
- 敏感字段只展示引用或掩码。

## 7. P3：findx-agents 控制面

> 目标：把 Agent 从边缘能力提升为 FindX 的核心控制面，完成主机接入、能力治理、采集配置、安装调试、命令执行、自升级、自愈和单机巡检基础。

### P3-T1：Agent 注册心跳

**建议 API**：

```text
POST /api/v1/findx-agents/register
POST /api/v1/findx-agents/heartbeat
GET  /api/v1/findx-agents
GET  /api/v1/findx-agents/:id
PUT  /api/v1/findx-agents/:id/labels
```

**能力范围**：

- 首次注册。
- 周期心跳。
- 版本上报。
- 主机资产绑定。
- 业务空间与资源组归属。
- 在线、离线、待升级、异常状态。

**验收**：

- 心跳超时可标记离线。
- Agent 与主机资产可双向查看。
- 注册和归属变更写入审计日志。

### P3-T2：能力目录

**建议 API**：

```text
GET /api/v1/findx-agents/:id/capabilities
GET /api/v1/findx-agents/capabilities/summary
```

**能力目录字段**：

- OS 与架构。
- Agent 版本。
- 采集插件能力。
- 巡检工具能力。
- 命令执行能力。
- 远程调试能力。
- 自升级能力。
- 自愈动作能力。

**验收**：

- 能力变更有版本记录。
- 不具备能力的 Agent 不展示对应动作入口。
- 权限不足时不返回可执行入口。

### P3-T3：Categraf 插件目录

**建议 API**：

```text
GET  /api/v1/findx-agents/plugins
GET  /api/v1/findx-agents/plugins/:name
POST /api/v1/findx-agents/plugins/:name/validate-config
```

**能力范围**：

- 插件名称、版本、平台支持。
- 默认配置模板。
- 配置校验规则。
- 指标样例。
- 与指标目录的映射关系。

**验收**：

- 插件目录不依赖前端硬编码。
- 配置校验失败返回用户可理解错误。
- 插件模板不含真实凭证。

### P3-T4：采集配置中心化下发

**建议 API**：

```text
GET    /api/v1/findx-agents/config-templates
POST   /api/v1/findx-agents/configs
GET    /api/v1/findx-agents/configs/:id
GET    /api/v1/findx-agents/configs/:id/diff
POST   /api/v1/findx-agents/configs/:id/publish
POST   /api/v1/findx-agents/configs/:id/rollback
GET    /api/v1/findx-agents/config-runs/:run_id
```

**能力范围**：

- 配置模板。
- Agent 分组选择。
- 中心化下发。
- 配置版本 diff。
- 配置发布。
- 配置回滚。
- 灰度发布。
- 下发结果追踪。

**验收**：

- 配置变更可比较、可发布、可回滚。
- 发布失败不影响上一版本采集。
- 凭证只以 `<SSH_KEY_REF>` 或内部凭证引用出现。
- 配置 diff 脱敏展示。

### P3-T5：远程安装与本机安装指令

**建议 API**：

```text
POST /api/v1/findx-agents/install-tasks
GET  /api/v1/findx-agents/install-tasks/:id
GET  /api/v1/findx-agents/install-command
POST /api/v1/findx-agents/install-command/rotate
```

**能力范围**：

- 远程安装任务。
- 本机安装指令生成。
- 安装结果追踪。
- 凭证引用。
- 安装包版本选择。
- 业务空间与资源组归属。

**验收**：

- 远程安装只引用凭证，不回显明文。
- 本机安装指令不包含真实密钥。
- 安装失败给出可理解原因和下一步动作。
- 安装、重试、取消均写入审计日志。

### P3-T6：凭证引用

**建议 API**：

```text
GET    /api/v1/findx-agents/credential-refs
POST   /api/v1/findx-agents/credential-refs
DELETE /api/v1/findx-agents/credential-refs/:id
```

**能力范围**：

- SSH 凭证引用。
- API 凭证引用。
- 凭证用途范围。
- 凭证有效期。
- 凭证轮换状态。

**验收**：

- 响应中不出现真实密钥、私钥、密码、Cookie、完整连接串。
- 仅展示凭证引用 ID、别名、用途、创建者和状态。
- 使用凭证的任务必须记录审计日志。

### P3-T7：远程调试

**建议 API**：

```text
POST /api/v1/findx-agents/:id/debug
GET  /api/v1/findx-agents/debug-runs/:run_id
```

**能力范围**：

- 查看 Agent 状态。
- 拉取采集配置摘要。
- 执行安全诊断命令。
- 收集调试日志片段。
- 输出证据产物。

**验收**：

- 调试入口受权限控制。
- 调试输出脱敏。
- 调试结果关联 `evidence_refs`。
- 不允许执行未登记的任意命令。

### P3-T8：自升级

**建议 API**：

```text
POST /api/v1/findx-agents/:id/upgrade
POST /api/v1/findx-agents/upgrade-tasks
GET  /api/v1/findx-agents/upgrade-tasks/:id
POST /api/v1/findx-agents/upgrade-tasks/:id/rollback
```

**能力范围**：

- 单 Agent 升级。
- 批量升级。
- 灰度发布。
- 失败回滚。
- 版本锁定。
- 状态追踪。

**验收**：

- 升级任务可观察、可取消、可回滚。
- 离线 Agent 不执行升级，只标记待处理。
- 升级失败不删除旧版本。

### P3-T9：结构化命令执行

**建议 API**：

```text
POST /api/v1/findx-agents/:id/commands
GET  /api/v1/findx-agents/command-runs/:run_id
POST /api/v1/findx-agents/command-runs/:run_id/cancel
```

**输出结构示例**：

```json
{
  "run_id": "cmd-001",
  "agent_id": "agent-001",
  "status": "success",
  "exit_code": 0,
  "stdout_ref": "artifact-stdout-001",
  "stderr_ref": "artifact-stderr-001",
  "evidence_refs": ["ev-001"],
  "redaction": "applied"
}
```

**验收**：

- 命令必须来自结构化命令目录。
- 命令执行必须校验权限、业务空间和资源范围。
- 标准输出和错误输出保存为证据产物引用，不直接无限制回显。

### P3-T10：告警自愈入口

**建议 API**：

```text
POST /api/v1/findx-agents/self-healing/candidates
POST /api/v1/findx-agents/self-healing/:id/execute
GET  /api/v1/findx-agents/self-healing/:id
```

**能力范围**：

- 基于告警事件生成自愈候选。
- 校验 Agent 能力。
- 校验权限和审批状态。
- 执行受控动作。
- 生成证据产物。

**验收**：

- 自愈候选不自动执行高风险动作。
- 执行前必须有审计和审批策略。
- 失败时生成证据并给出回滚建议。

### P3-T11：findx-agents inspector 单机巡检

**建议 API**：

```text
POST /api/v1/findx-agents/:id/inspections
GET  /api/v1/findx-agents/inspections/:run_id
GET  /api/v1/findx-agents/inspections/:run_id/evidence
```

**能力范围**：

- 单机巡检工具编排。
- 巡检项目录。
- 巡检结果结构化。
- 巡检证据产物。
- 与 AI SRE 主机健康检查联动。
- 可吸收 Catpaw 授权能力/思想作为内部实现来源，用户侧统一呈现为 FindX 单机巡检能力。

**验收**：

- 用户侧入口展示为 FindX 单机巡检，不把底层工具或内部来源作为主产品入口。
- 巡检结果输出 `evidence_refs`。
- 巡检不能绕过权限和审计。

### P3-T12：诊断会话、证据产物、审计脱敏

**建议 API**：

```text
POST /api/v1/findx-agents/diagnosis-sessions
GET  /api/v1/findx-agents/diagnosis-sessions/:id
POST /api/v1/findx-agents/diagnosis-sessions/:id/attach-evidence
GET  /api/v1/evidence/:id
```

**能力范围**：

- Agent 诊断会话。
- 远程调试输出入证据链。
- 命令执行输出入证据链。
- 巡检输出入证据链。
- 敏感字段统一脱敏。

**验收**：

- 诊断会话可关联 Agent、主机资产、告警事件、业务空间。
- 证据引用可追溯来源。
- 审计日志记录操作者、时间、资源、动作、结果、脱敏状态。

## 8. P4：多信号数据与可观测底座

> 目标：从指标扩展到日志、Trace/APM、RUM 和多数据源，为 AI SRE 提供更完整的证据输入。

### P4-T1：日志查询

**建议 API**：

```text
POST /api/v1/logs/query
GET  /api/v1/logs/fields
GET  /api/v1/logs/indexes
```

**能力范围**：

- 关键字查询。
- 字段过滤。
- 上下文展开。
- 保存查询。
- 与主机资产、业务空间关联。

**验收**：

- 查询错误不暴露内部连接信息。
- 日志结果可作为 AI SRE 证据输入。

### P4-T2：Trace/APM

**建议 API**：

```text
POST /api/v1/traces/query
GET  /api/v1/traces/services
GET  /api/v1/traces/operations
GET  /api/v1/apm/services
GET  /api/v1/apm/services/:id/overview
```

**能力范围**：

- 服务列表。
- Trace 查询。
- Span 明细。
- 延迟、错误率、吞吐概览。
- 与告警事件和 AI SRE 诊断联动。

**验收**：

- Trace 数据可按业务空间过滤。
- Trace 片段可生成证据引用。

### P4-T3：RUM 基础接入

**建议 API**：

```text
GET  /api/v1/rum/apps
POST /api/v1/rum/apps
GET  /api/v1/rum/apps/:id/overview
POST /api/v1/rum/events/query
```

**能力范围**：

- Web/App 应用接入。
- 页面性能。
- JS 错误。
- 会话维度。
- 与前端告警联动。

**验收**：

- RUM 接入示例不包含真实 Token。
- 用户隐私字段需要脱敏或聚合。

### P4-T4：指标目录

**建议 API**：

```text
GET  /api/v1/metrics/catalog
GET  /api/v1/metrics/catalog/:name
POST /api/v1/metrics/catalog/sync
```

**能力范围**：

- 指标名称。
- 标签集合。
- 单位。
- 来源插件。
- 业务含义。
- 推荐告警规则。

**验收**：

- 指标目录可关联 Categraf 插件目录。
- 指标说明可被仪表盘、告警、AI SRE 复用。

### P4-T5：智能告警

**建议 API**：

```text
GET    /api/v1/alerts/rules
POST   /api/v1/alerts/rules
PUT    /api/v1/alerts/rules/:id
DELETE /api/v1/alerts/rules/:id
POST   /api/v1/alerts/rules/:id/simulate
```

**能力范围**：

- 静态阈值。
- 动态基线。
- 同比环比。
- 异常检测。
- 规则模拟。
- 与通知策略联动。

**验收**：

- 规则模拟可解释触发原因。
- 告警规则绑定业务空间和资源组。
- AI 推荐规则不得直接写成已自动落地，需人工确认。

### P4-T6：保存视图

**建议 API**：

```text
GET    /api/v1/saved-views
POST   /api/v1/saved-views
PUT    /api/v1/saved-views/:id
DELETE /api/v1/saved-views/:id
```

**能力范围**：

- 保存日志查询视图。
- 保存 Trace 查询视图。
- 保存资产过滤视图。
- 保存告警事件过滤视图。
- 团队共享。

**验收**：

- 保存视图受业务空间权限控制。
- 删除前提示影响范围。

### P4-T7：多数据源

**建议 API**：

```text
GET    /api/v1/datasources
POST   /api/v1/datasources
PUT    /api/v1/datasources/:id
DELETE /api/v1/datasources/:id
POST   /api/v1/datasources/:id/test
```

**能力范围**：

- 指标数据源。
- 日志数据源。
- Trace 数据源。
- 数据源连通性测试。
- 凭证引用。

**验收**：

- 连通性测试属于数据源能力，不等同主机健康检查。
- 数据源错误信息脱敏。
- 凭证只以引用方式保存和展示。

## 9. P5：AI SRE 与证据链

> 目标：把主机健康检查、单机巡检、AI 问诊、RCA、证据链、自动修复和复盘报告组织成可审计、可回滚、可验证的 AI SRE 闭环。

### P5-T1：AI 主机健康检查

**建议 API**：

```text
POST /api/v1/aiops/host-health/check
GET  /api/v1/aiops/host-health/runs/:run_id
GET  /api/v1/aiops/host-health/runs/:run_id/evidence
```

**输入**：

- 主机。
- Agent。
- 指标。
- 日志。
- Trace。
- 巡检工具。
- 采集配置快照。

**输出示例**：

```json
{
  "run_id": "health-001",
  "host_id": "host-001",
  "risk_level": "medium",
  "summary": "磁盘增长速度异常，建议检查日志轮转策略。",
  "evidence_refs": ["ev-101", "ev-102"],
  "recommended_actions": ["inspect_log_rotation", "expand_disk_threshold"],
  "requires_approval": true
}
```

**验收**：

- 主机健康检查入口在 AI SRE。
- 输出必须包含 `evidence_refs`、风险等级、建议动作。
- 平台管理不出现主机健康检查入口。

### P5-T2：单机巡检

**建议 API**：

```text
POST /api/v1/aiops/host-health/inspections
GET  /api/v1/aiops/host-health/inspections/:run_id
GET  /api/v1/aiops/host-health/inspections/:run_id/evidence
```

**能力范围**：

- 选择主机。
- 选择巡检项。
- 调用 findx-agents 单机巡检能力。
- 汇总指标、日志、Trace、命令输出。
- 生成风险等级和建议动作。

**验收**：

- 巡检结果结构化。
- 巡检证据可追溯到 Agent、工具、时间、操作者。
- 巡检不能绕过 findx-agents 权限和审计。

### P5-T3：AI 问诊

**建议 API**：

```text
POST /api/v1/aiops/diagnosis-sessions
GET  /api/v1/aiops/diagnosis-sessions/:id
POST /api/v1/aiops/diagnosis-sessions/:id/messages
POST /api/v1/aiops/diagnosis-sessions/:id/attach-evidence
```

**能力范围**：

- 基于告警、主机、仪表盘、日志、Trace 发起问诊。
- 支持多轮上下文。
- 引用证据产物。
- 输出假设、验证步骤、风险判断和建议动作。

**验收**：

- AI 结论必须关联证据。
- 无证据的建议标记为低置信度。
- 输出不包含敏感字段明文。

### P5-T4：RCA

**建议 API**：

```text
POST /api/v1/aiops/rca
GET  /api/v1/aiops/rca/:id
POST /api/v1/aiops/rca/:id/verify
```

**能力范围**：

- 事件时间线。
- 指标异常点。
- 日志异常片段。
- Trace 异常链路。
- 变更记录。
- Agent 巡检结果。
- 根因候选排序。

**验收**：

- RCA 输出根因候选，不把猜测写成事实。
- 每个根因候选关联证据引用和置信度。
- 验证步骤可执行或可人工确认。

### P5-T5：Evidence Chain

**建议 API**：

```text
GET  /api/v1/evidence
GET  /api/v1/evidence/:id
POST /api/v1/evidence
POST /api/v1/evidence/:id/redact
GET  /api/v1/evidence/chains/:chain_id
```

**证据类型**：

- metric
- log
- trace
- command
- config
- inspection
- alert
- remediation
- screenshot
- note

**验收**：

- 证据链记录来源、时间、操作者、关联资源、脱敏状态。
- 证据内容大字段可用引用保存。
- 删除或脱敏证据需要审计。

### P5-T6：自动修复

**建议 API**：

```text
POST /api/v1/remediation/candidates
GET  /api/v1/remediation/:id
POST /api/v1/remediation/:id/approve
POST /api/v1/remediation/:id/execute
POST /api/v1/remediation/:id/verify
POST /api/v1/remediation/:id/rollback
```

**能力范围**：

- 基于 AI SRE 诊断生成修复候选。
- 执行前 precheck。
- dry-run。
- 审批。
- 执行。
- 验证。
- 回滚。
- 复盘归档。

**验收**：

- 高风险修复必须人工审批。
- 修复执行通过 findx-agents 结构化命令或受控动作完成。
- 每一步都有审计和证据引用。
- 失败后给出回滚或人工处理建议。

### P5-T7：复盘报告

**建议 API**：

```text
POST /api/v1/aiops/postmortems
GET  /api/v1/aiops/postmortems/:id
PUT  /api/v1/aiops/postmortems/:id
POST /api/v1/aiops/postmortems/:id/publish
```

**能力范围**：

- 基于告警事件、诊断会话、证据链、修复记录生成草稿。
- 人工确认事实、时间线和行动项。
- 归档与搜索。

**验收**：

- AI 只生成草稿，不自动发布。
- 报告引用证据链。
- 行动项可追踪负责人和状态。

## 10. P6：商业化与治理增强

> 目标：补齐商业化、合规、运营治理和生态扩展能力。这些是长期能力，不得写成当前已经完成。

| 任务 | 范围 | 验收方向 |
|---|---|---|
| Status Page | 服务状态公开页、事件公告、订阅通知 | 可选择公开范围，事件状态与告警事件联动 |
| on-call 深化 | 排班、升级策略、临时替换、通知策略联动 | 值班变更可审计，升级策略可模拟 |
| SSO | OAuth、CAS、企业身份提供方接入 | 登录失败不泄露内部配置，用户映射可审计 |
| 合规 | 审计留存、脱敏策略、权限证明、证据保全 | 审计和证据链可导出，敏感字段可配置脱敏 |
| 容量成本 | 资源趋势、容量建议、成本归因、预算提醒 | 建议只作为决策输入，不自动执行资源变更 |
| 模板市场 | 仪表盘模板、告警模板、巡检模板、修复模板 | 模板可预览、导入、版本管理和安全校验 |

## 11. 平台管理边界

平台管理只负责 FindX 自身运行治理：

- 系统配置。
- 模型配置。
- 数据源配置入口。
- 平台运行自检。
- 后台任务状态。
- 审计日志入口。
- 版本信息。

平台运行自检只检查 FindX 自身组件和依赖，例如 API 服务、数据库连通性、缓存、任务队列、模型配置、数据源连通性。平台管理不负责主机健康检查、单机巡检、Agent 巡检或 AI 健康诊断。

## 12. 全阶段验证要求

后续真正实施业务代码时，每个实现类任务至少提供：

- Go 编译：`cd /opt/ai-workbench/api && go build -o api-linux .`
- 前端构建：`cd /opt/ai-workbench/web && npm run build`
- WSL 同步验证。
- 正常路径验证。
- 异常路径验证。
- 权限或边界路径验证。
- 敏感信息脱敏验证。
- 审计日志验证。
- 未覆盖项说明。

文档-only 任务不执行业务构建，但必须执行：

- 关键命名自检。
- API 命名自检。
- 健康检查归属自检。
- 敏感信息自检。
- 长期能力表述自检。

## 13. 子代理派发模板

```text
[TASK_MARKER: YYYYMMDD-HHMMSS-XXXX]

## 角色
你是 D:\ai-workbench 项目的 {role} 子代理，model 必须使用 gpt-5.5。

## 任务目标
{具体任务描述}

## 允许路径
{允许修改的路径}

## 禁止路径
{禁止修改的路径}

## 写集边界
{本任务允许写入的文件或目录}

## 禁止扩大范围
- 不修改未授权业务代码。
- 不修改 README，除非本任务明确授权。
- 不写运行产物、真实密钥、认证票据、Cookie、完整连接串、会话 ID。
- 不把长期能力写成当前已经完成。

## 产品命名
- 使用 FindX、FindX Core、findx-agents、AI SRE。
- 使用业务空间、资源组、主机资产、监控仪表盘、通知渠道、通知策略、消息模板、团队、角色、权限矩阵、审计日志。
- 不使用外部产品命名作为模块名、API 名、模型名、页面名或任务名。

## 敏感信息规则
- 不写真实密钥、Cookie、Token、DSN、SSH 私钥。
- 示例统一使用 <TOKEN>、<API_KEY>、<DB_DSN>、<LOGIN_USER>、<SSH_KEY_REF>。

## 验收标准
1. 构建或文档验证通过。
2. 正常、异常、权限或边界路径有验证证据。
3. 无敏感信息泄露。
4. API、模型、页面、任务命名均为 FindX 风格。
5. findx-agents 能力归入控制面，不被放到边缘阶段。
6. 主机健康检查归入 AI SRE，平台管理只保留平台运行自检。
7. 输出技术债检查和未覆盖项。
```

## 14. 阶段依赖

```text
P0 FindX Core 基础监控可信基座
  -> P1 告警响应与通知闭环
  -> P2 团队权限与资源治理
  -> P3 findx-agents 控制面
  -> P4 多信号数据与可观测底座
  -> P5 AI SRE 与证据链
  -> P6 商业化与治理增强
```

依赖说明：

- P0 提供业务空间、资源组、主机资产和监控仪表盘。
- P1 基于 P0 的资源归属完成告警响应和通知闭环。
- P2 提供团队、角色、权限矩阵、API Token、审计日志，为 P3/P5 的高风险动作提供安全边界。
- P3 基于 P0/P2 完成 findx-agents 控制面，并向 P5 提供单机巡检、结构化命令和证据产物。
- P4 扩展日志、Trace/APM、RUM、指标目录和多数据源，为 P5 提供多信号证据。
- P5 基于 P1/P3/P4 完成 AI SRE 闭环。
- P6 在前序能力稳定后扩展商业化与治理增强。

## 15. 验收标准

### 15.1 命名验收

- 文档标题、阶段、API、模型、页面、任务名均为 FindX 风格。
- 基础监控归入 FindX Core。
- Agent 相关能力统一归入 findx-agents 控制面。
- 主机健康检查和单机巡检归入 AI SRE。
- 平台管理只保留平台运行自检。

### 15.2 findx-agents 验收

- P3 明确覆盖 Agent 注册心跳。
- P3 明确覆盖能力目录。
- P3 明确覆盖 Categraf 插件目录。
- P3 明确覆盖采集配置中心化下发。
- P3 明确覆盖配置版本 diff 与回滚。
- P3 明确覆盖远程安装和本机安装指令。
- P3 明确覆盖凭证引用。
- P3 明确覆盖远程调试。
- P3 明确覆盖自升级。
- P3 明确覆盖结构化命令执行。
- P3 明确覆盖告警自愈。
- P3 明确覆盖 findx-agents inspector 单机巡检。
- P3 明确覆盖诊断会话、证据产物、审计脱敏。

### 15.3 AI SRE 验收

- P5 明确 AI 主机健康检查输入：主机、Agent、指标、日志、Trace、巡检工具。
- P5 明确 AI 主机健康检查输出：`evidence_refs`、风险等级、建议动作。
- AI 问诊、RCA、Evidence Chain、自动修复、复盘报告均不写成已完成能力。
- 自动修复包含审批、执行、验证、回滚和审计。

### 15.4 安全验收

- 示例不包含真实密钥、Cookie、Token、完整 DSN、SSH 私钥。
- 凭证只写引用，不写明文。
- 远程安装、远程调试、结构化命令执行、自升级、自动修复必须有权限控制、审计日志和脱敏输出。

### 15.5 规划验收

- 本文只作为实施计划，不宣称长期能力已经完成。
- 每个阶段都有目标、任务、API 方向和验收方向。
- 全阶段验证要求、子代理派发模板和验收标准均使用 FindX 命名。
