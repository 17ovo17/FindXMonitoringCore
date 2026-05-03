# FindX Monitoring Core API 契约、P0 验证证据与 P1 后端告警闭环

生成时间：2026-05-04 06:59（UTC+8）

## 1. 文档定位

本文档为 FindX Monitoring Core P0-1/P0-2/P0-3 的 API 契约、测试清单、P0 验证证据，以及 P1-BE-1/2/3 已落库的后端告警闭环契约，用于指导后端、前端、QA 和文档同步。FindX Monitoring Core 是 FindX 新平台主线，负责承载自有监控事实源、Target、Agent、告警规则、告警事件、诊断与后续修复闭环。

当前完成边界：

- P0 已完成并通过 QA 门禁，覆盖 Target、Agent、Alert Rule、Event 和 Query Gateway 后端主线。
- P1 后端告警闭环已部分完成，覆盖事件 fingerprint 状态机、真实 Prometheus evaluator/tryrun、scheduler 自动评估与恢复闭环。
- P1 的 Dashboard、通知、细粒度权限、前端页面和完整 AI/修复链路仍待后续实现与 QA 回填，本文不得将其描述为已完成。

兼容策略：

- 新平台主线 API 使用 `/api/v1/monitor/*` 与 `/api/v1/findx-agents/*`。
- 旧 `/api/v1/n9e/*` 仅作为 Nightingale 兼容、导入或过渡入口。
- 旧 `/api/v1/catpaw/*` 仅作为 Catpaw 兼容或过渡入口，新 Agent 主线迁移到 `/api/v1/findx-agents/*`。
- 新功能不得继续以 `/n9e/*` 或 `/catpaw/*` 作为用户主入口。

## 2. 通用约定

### 2.1 认证与权限

| 接口类型 | 认证要求 | 权限要求 | 说明 |
| --- | --- | --- | --- |
| 读接口 | 平台登录态或平台 API Token | 普通读权限 | 包括健康、Target 列表、Agent 列表、规则列表、事件列表和详情。 |
| 写接口 | 平台登录态或平台 API Token | `adminRequired` | 包括创建、更新、删除、启停、克隆、回滚、事件处置等会改变状态的操作。 |
| Agent 注册/心跳 | `X-Agent-Token: <TOKEN>` 必填 | Agent 通道权限 | 默认拒绝匿名 Agent 写入。仅当服务端配置 `FINDX_AGENT_TOKEN` 或 `findx_agents.shared_token`，且请求头 `X-Agent-Token` 与配置值匹配时，才允许 register/heartbeat。若确需本地或测试环境匿名接入，必须显式设置 `findx_agents.allow_anonymous: true`；该配置禁止在生产环境使用。 |
| AI/诊断/巡检触发 | 平台登录态或平台 API Token | 读权限或 `adminRequired`，按动作风险分级 | 只读诊断可读权限；可能触发远程执行、修复计划或写状态的接口必须 `adminRequired`。 |

敏感信息规则：

- 请求示例统一使用 `<TOKEN>`，不得写真实 token、Cookie、连接串、SSH 私钥、数据库 DSN。
- 响应、错误、日志、测试报告和 AI prompt 不得回显 Authorization、Cookie、`X-Agent-Token`、完整 DSN、完整内部 URL、SSH key 或原始堆栈。
- Agent token 比较必须使用常量时间比较；文档中只描述策略，不记录实际值。
- Agent 匿名接入默认关闭；`findx_agents.allow_anonymous: true` 只能用于本地开发、自动化测试或隔离测试环境，生产环境不得开启。

### 2.2 通用响应

成功响应按资源类型直接返回对象、数组或操作结果：

```json
{
  "ok": true
}
```

错误响应统一保持可读且脱敏：

```json
{
  "error": "invalid heartbeat payload"
}
```

后续可扩展为：

```json
{
  "error": "invalid heartbeat payload",
  "code": "INVALID_PAYLOAD",
  "trace_id": "req-xxxx"
}
```

扩展时必须保证前端兼容旧 `error` 字段。

### 2.3 错误模型

| HTTP 状态码 | 触发场景 | 响应要求 |
| --- | --- | --- |
| 400 | JSON 解析失败、字段缺失、IP 格式非法、非法状态值、非法分页参数、非法动作参数 | 返回用户可理解错误，不暴露内部结构。 |
| 401 | 未登录、平台 token 无效、Agent token 缺失、Agent token 无效、未配置共享 token 且未显式允许匿名 Agent | 返回认证失败，不说明真实 token 配置状态，不回显请求 token 或配置值。 |
| 403 | 登录有效但缺少 `adminRequired` 或对应资源权限 | 返回权限不足，不泄露资源详情。 |
| 404 | Target、Agent、Rule、Event 不存在或当前用户不可见 | 返回资源不存在。 |
| 503 | MySQL、Prometheus、Agent 通道、诊断依赖等外部依赖不可用 | 返回降级状态和可读提示；不得误报成功。 |
| 500 | 未预期内部错误 | 返回通用错误和 trace_id；不得返回 SQL、堆栈、本地路径、密钥或完整连接串。 |

### 2.4 安全基线补充

CORS 与 Workflow DSL 的后端基线已进入契约约束：

- CORS 默认只允许 localhost 开发源访问，不默认放开通配来源。
- 只有显式配置 `server.allowed_origins: ["*"]` 时，服务端才允许 wildcard CORS；其他情况下应按配置白名单收敛来源。
- Workflow DSL 创建和更新入口在 `ParseDSL` 后必须执行 `graph.Validate()`。
- 坏 DSL 必须返回 400，且不得持久化 workflow、版本或衍生执行图。
- 错误响应和日志不得包含完整 DSL 内部堆栈、真实密钥、Cookie、完整 DSN 或完整内部 URL。

## 3. P0-1 已实现 API

当前 P0-1 后端已接入以下路由：

```text
GET  /api/v1/monitor/health
GET  /api/v1/monitor/targets
GET  /api/v1/findx-agents
POST /api/v1/findx-agents/register
POST /api/v1/findx-agents/heartbeat
```

同批实现中也已有 Target 写接口：

```text
POST   /api/v1/monitor/targets
GET    /api/v1/monitor/targets/:id
PUT    /api/v1/monitor/targets/:id
DELETE /api/v1/monitor/targets/:id
```

### 3.1 GET /api/v1/monitor/health

用途：查看 FindX Monitoring Core 运行健康、存储状态、Target 数量和 Agent 在线情况。

权限：读接口。

请求参数：无。

响应字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `status` | string | 是 | `healthy`、`degraded`、`empty`。 |
| `mode` | string | 是 | 固定主线模式，当前为 `findx-core`。 |
| `storage` | object | 是 | 存储健康状态，例如 MySQL 与内存 fallback 状态。 |
| `targets` | number | 是 | Target 总数。 |
| `agents` | number | 是 | Agent 总数。 |
| `agent_online` | number | 是 | 在线 Agent 数量。 |
| `generated_at` | string | 是 | 服务端生成时间，ISO/RFC3339。 |

示例响应：

```json
{
  "status": "healthy",
  "mode": "findx-core",
  "storage": {
    "mysql": true,
    "memory_fallback": false
  },
  "targets": 12,
  "agents": 10,
  "agent_online": 9,
  "generated_at": "2026-05-04T03:24:00+08:00"
}
```

状态语义：

| `status` | 语义 |
| --- | --- |
| `healthy` | 存储可用且已有 Agent 数据。 |
| `degraded` | 存储或依赖降级，例如 MySQL 不可用。 |
| `empty` | 核心可运行但还没有 Agent 数据。 |

### 3.2 GET /api/v1/monitor/targets

用途：查询监控目标列表。

权限：读接口。

查询参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `status` | string | 否 | 按 Target 状态过滤，例如 `online`、`offline`、`unknown`。 |

响应：`MonitorTarget[]`。

`MonitorTarget` 字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | Target ID。 |
| `ident` | string | 否 | 目标唯一标识，优先来自 Agent ident。 |
| `name` | string | 否 | 展示名称。 |
| `ip` | string | 否 | 目标 IP。 |
| `hostname` | string | 否 | 主机名。 |
| `os` | string | 否 | 操作系统。 |
| `arch` | string | 否 | CPU 架构。 |
| `environment` | string | 否 | 环境，例如 `prod`、`staging`、`dev`。 |
| `business_group` | string | 否 | 业务组。 |
| `owner` | string | 否 | 负责人。 |
| `status` | string | 是 | 目标状态。 |
| `source` | string | 是 | 来源，例如 `findx-agent`、`manual`、`import`。 |
| `labels` | object | 否 | 标签键值对。 |
| `metadata` | object | 否 | 元数据键值对。 |
| `last_seen` | string | 否 | 最近上报时间。 |
| `created_at` | string | 是 | 创建时间。 |
| `updated_at` | string | 是 | 更新时间。 |

示例响应：

```json
[
  {
    "id": "target-001",
    "ident": "host-a",
    "name": "host-a",
    "ip": "10.0.0.10",
    "hostname": "host-a",
    "os": "linux",
    "arch": "amd64",
    "environment": "prod",
    "business_group": "core",
    "owner": "ops",
    "status": "online",
    "source": "findx-agent",
    "labels": {
      "region": "cn-east"
    },
    "metadata": {
      "kernel": "linux"
    },
    "last_seen": "2026-05-04T03:24:00+08:00",
    "created_at": "2026-05-04T03:00:00+08:00",
    "updated_at": "2026-05-04T03:24:00+08:00"
  }
]
```

### 3.3 POST /api/v1/monitor/targets

用途：手工创建或导入 Target。

权限：`adminRequired`。

请求字段：同 `MonitorTarget` 可写字段。`id` 可省略，由服务端生成或归一化；`created_at`、`updated_at`、`last_seen` 由服务端维护。

最小合法请求必须满足 `ident`、`ip`、`hostname` 至少一个非空；如果传入 `ip`，必须是合法 IP。

示例请求：

```json
{
  "ident": "host-a",
  "name": "host-a",
  "ip": "10.0.0.10",
  "hostname": "host-a",
  "environment": "prod",
  "business_group": "core",
  "owner": "ops",
  "status": "online",
  "source": "manual",
  "labels": {
    "region": "cn-east"
  },
  "metadata": {
    "rack": "rack-01"
  }
}
```

响应：创建或更新后的 `MonitorTarget`。

### 3.4 GET /api/v1/monitor/targets/:id

用途：查询单个 Target 详情。

权限：读接口。

路径参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | Target ID。 |

响应：`MonitorTarget`。

错误：

- 404：Target 不存在或不可见。

### 3.5 PUT /api/v1/monitor/targets/:id

用途：更新 Target。

权限：`adminRequired`。

路径参数：`id` 为 Target ID。

请求字段：同 `POST /api/v1/monitor/targets`。

约束：

- 路径 `id` 优先于请求体 `id`。
- `ident`、`ip`、`hostname` 至少一个非空。
- `ip` 非空时必须合法。

响应：更新后的 `MonitorTarget`。

### 3.6 DELETE /api/v1/monitor/targets/:id

用途：删除 Target。

权限：`adminRequired`。

响应：

```json
{
  "ok": true
}
```

错误：

- 404：Target 不存在或不可见。

### 3.7 GET /api/v1/findx-agents

用途：查询 FindX Agent 列表。

权限：读接口。

响应：`FindXAgent[]`。

`FindXAgent` 字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | Agent ID。 |
| `ident` | string | 否 | Agent 唯一标识。 |
| `target_id` | string | 否 | 绑定 Target ID。 |
| `ip` | string | 否 | Agent IP。 |
| `hostname` | string | 是 | 主机名。 |
| `os` | string | 否 | 操作系统。 |
| `arch` | string | 否 | CPU 架构。 |
| `version` | string | 否 | Agent 版本。 |
| `collector` | string | 否 | 采集器来源，例如 `categraf`、`findx-agent`。 |
| `status` | string | 是 | `online`、`offline`、`unknown` 等。 |
| `capabilities` | string[] | 否 | Agent 能力，例如 `metrics`、`inspect`、`diagnose`、`session`、`remediation`。 |
| `global_labels` | object | 否 | 全局标签。 |
| `config_version` | string | 否 | Agent 配置版本。 |
| `last_seen` | string | 是 | 最近心跳时间。 |
| `created_at` | string | 是 | 创建时间。 |
| `updated_at` | string | 是 | 更新时间。 |

### 3.8 POST /api/v1/findx-agents/register

用途：Agent 初次注册。当前契约与 heartbeat 相同，服务端可复用心跳处理，返回 Agent 与 Target 绑定结果。

权限：Agent 通道。默认必须配置 `FINDX_AGENT_TOKEN` 或 `findx_agents.shared_token`，且请求头必须包含：

```text
X-Agent-Token: <TOKEN>
```

认证策略：
- 请求头 `X-Agent-Token` 缺失或不匹配时返回 401，且不得写入 Agent 或 Target。
- 服务端未配置 `FINDX_AGENT_TOKEN` 和 `findx_agents.shared_token` 时，默认仍返回 401。
- 仅当 `findx_agents.allow_anonymous: true` 显式开启时，register 可在无 token 情况下接入；该模式仅限本地开发、自动化测试或隔离测试环境，生产环境禁止开启。

请求字段：`FindXAgentHeartbeat`。

响应字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `ok` | boolean | 是 | 操作是否成功。 |
| `agent` | object | 是 | 注册后的 `FindXAgent`。 |
| `target` | object | 是 | upsert 后的 `MonitorTarget`。 |

### 3.9 POST /api/v1/findx-agents/heartbeat

用途：Agent 周期心跳，上报身份、主机属性、能力、标签和配置版本。

权限：Agent 通道；认证策略同注册接口。默认必须提供有效 `X-Agent-Token: <TOKEN>`，未配置共享 token 时也不得自动放行匿名写入。

请求头示例：

```text
X-Agent-Token: <TOKEN>
Content-Type: application/json
```

请求字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `ident` | string | 条件必填 | 与 `ip`、`hostname` 至少一个非空。 |
| `ip` | string | 条件必填 | 与 `ident`、`hostname` 至少一个非空；非空时必须是合法 IP。 |
| `hostname` | string | 条件必填 | 与 `ident`、`ip` 至少一个非空。 |
| `os` | string | 否 | 操作系统。 |
| `arch` | string | 否 | CPU 架构。 |
| `version` | string | 否 | Agent 版本。 |
| `collector` | string | 否 | 采集核心，例如 `categraf`。 |
| `capabilities` | string[] | 否 | Agent 能力列表。 |
| `global_labels` | object | 否 | 全局标签。 |
| `config_version` | string | 否 | 当前配置版本。 |
| `unixtime` | number | 否 | Agent 本地 Unix 时间。 |

示例请求：

```json
{
  "ident": "host-a",
  "ip": "10.0.0.10",
  "hostname": "host-a",
  "os": "linux",
  "arch": "amd64",
  "version": "0.1.0",
  "collector": "categraf",
  "capabilities": ["metrics", "inspect", "diagnose"],
  "global_labels": {
    "region": "cn-east",
    "env": "prod"
  },
  "config_version": "cfg-20260504-001",
  "unixtime": 1777826640
}
```

示例响应：

```json
{
  "ok": true,
  "agent": {
    "id": "agent-001",
    "ident": "host-a",
    "target_id": "target-001",
    "ip": "10.0.0.10",
    "hostname": "host-a",
    "os": "linux",
    "arch": "amd64",
    "version": "0.1.0",
    "collector": "categraf",
    "status": "online",
    "capabilities": ["metrics", "inspect", "diagnose"],
    "global_labels": {
      "region": "cn-east",
      "env": "prod"
    },
    "config_version": "cfg-20260504-001",
    "last_seen": "2026-05-04T03:24:00+08:00",
    "created_at": "2026-05-04T03:00:00+08:00",
    "updated_at": "2026-05-04T03:24:00+08:00"
  },
  "target": {
    "id": "target-001",
    "ident": "host-a",
    "ip": "10.0.0.10",
    "hostname": "host-a",
    "status": "online",
    "source": "findx-agent",
    "created_at": "2026-05-04T03:00:00+08:00",
    "updated_at": "2026-05-04T03:24:00+08:00"
  }
}
```

错误：

- 400：请求体不是合法 JSON、`ident/ip/hostname` 全为空、IP 非法。
- 401：`X-Agent-Token` 缺失、`X-Agent-Token` 不匹配、服务端未配置共享 token 且 `findx_agents.allow_anonymous` 未显式开启；失败时不得写入 Agent 或 Target。

## 4. P0-2 已通过 QA 门禁 API：Alert Rules 全生命周期

P0-2 后端 API 已通过 QA 门禁，覆盖 FindX 自有告警规则生命周期、版本回滚、tryrun 与事件入口在正常、异常、权限、断连和脱敏路径下的稳定性要求。

### 4.1 路由清单

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
GET    /api/v1/monitor/alert-rules/:id/versions
POST   /api/v1/monitor/alert-rules/:id/rollback
POST   /api/v1/monitor/alert-rules/import
GET    /api/v1/monitor/alert-rules/export
```

权限：

- `GET`：读接口。
- `POST/PUT/DELETE/enable/disable/clone/rollback/import`：`adminRequired`。
- `tryrun`：默认 `adminRequired`；如后续只做只读查询且无审计风险，可降为具备规则编辑权限的用户。

### 4.2 MonitorAlertRule 字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | Rule ID。 |
| `name` | string | 是 | 规则名称。 |
| `query` | string | 是 | PromQL 或 FindX 查询表达式。 |
| `severity` | string | 是 | 严重级别，例如 `critical`、`warning`、`info`。 |
| `datasource_id` | string | 是 | 数据源 ID。 |
| `target_selector` | object | 否 | Target 标签选择器。 |
| `labels` | object | 否 | 事件标签。 |
| `annotations` | object | 否 | 事件说明、summary、description、runbook 等。 |
| `enabled` | boolean | 是 | 是否启用。 |
| `version` | number | 是 | 规则版本号。 |
| `for_duration` | string | 是 | 持续满足时间，例如 `5m`。 |
| `no_data_policy` | string | 是 | 无数据策略，例如 `ignore`、`alert`、`resolve`。 |
| `status` | string | 是 | 规则状态，例如 `active`、`disabled`、`draft`。 |
| `created_by` | string | 否 | 创建人。 |
| `updated_by` | string | 否 | 更新人。 |
| `created_at` | string | 是 | 创建时间。 |
| `updated_at` | string | 是 | 更新时间。 |

创建请求示例：

```json
{
  "name": "CPU 使用率过高",
  "query": "avg(rate(node_cpu_seconds_total{mode!='idle'}[5m])) by (instance) > 0.9",
  "severity": "critical",
  "datasource_id": "platform-prom",
  "target_selector": {
    "env": "prod"
  },
  "labels": {
    "service": "core"
  },
  "annotations": {
    "summary": "CPU 使用率超过阈值",
    "runbook": "docs/runbooks/cpu-high.md"
  },
  "enabled": true,
  "for_duration": "5m",
  "no_data_policy": "ignore"
}
```

### 4.3 规则列表

`GET /api/v1/monitor/alert-rules`

查询参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `enabled` | boolean | 否 | 按启用状态过滤。 |
| `severity` | string | 否 | 按级别过滤。 |
| `datasource_id` | string | 否 | 按数据源过滤。 |
| `q` | string | 否 | 按名称或标签搜索。 |
| `page` | number | 否 | 页码，默认 1。 |
| `page_size` | number | 否 | 每页数量，默认 20，最大 100。 |

建议响应：

```json
{
  "items": [],
  "page": 1,
  "page_size": 20,
  "total": 0
}
```

### 4.4 规则详情、更新、删除

`GET /api/v1/monitor/alert-rules/:id`

- 返回 `MonitorAlertRule`。
- 404：规则不存在或不可见。

`PUT /api/v1/monitor/alert-rules/:id`

- 请求字段同创建。
- 更新必须生成新版本，`version` 自增。
- 必须写审计记录，包含操作者、旧值摘要、新值摘要、来源、trace_id。

`DELETE /api/v1/monitor/alert-rules/:id`

- 建议软删除或归档，避免历史事件无法追溯。
- 响应 `{ "ok": true }`。

### 4.5 启用、禁用、克隆

`POST /api/v1/monitor/alert-rules/:id/enable`

`POST /api/v1/monitor/alert-rules/:id/disable`

请求：

```json
{
  "reason": "变更窗口结束"
}
```

响应：更新后的 `MonitorAlertRule`。

`POST /api/v1/monitor/alert-rules/:id/clone`

请求：

```json
{
  "name": "CPU 使用率过高 - 副本",
  "enabled": false
}
```

响应：新建的 `MonitorAlertRule`。

### 4.6 试运行

`POST /api/v1/monitor/alert-rules/:id/tryrun`

用途：验证查询、数据源、目标选择器、阈值判断和无数据策略，不产生正式 current event。

请求：

```json
{
  "range": "30m",
  "sample_limit": 20,
  "override_query": ""
}
```

响应字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `ok` | boolean | 是 | 试运行是否通过。 |
| `status` | string | 是 | `pass`、`fail`、`warning`。 |
| `checks` | array | 是 | 分项检查。 |
| `rule` | object | 否 | 试运行使用的规则快照。 |
| `eval_log` | object | 是 | 本次评估日志。 |

`MonitorTryCheck`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `name` | string | 检查项名称。 |
| `status` | string | `pass`、`fail`、`warning`。 |
| `message` | string | 可读说明，必须脱敏。 |

P1-BE-2 evaluator/tryrun 已落库契约：

- 提交记录：`1b37905 feat: add FindX alert rule evaluator tryrun`。
- tryrun 使用真实 Prometheus-compatible 查询路径执行规则表达式，不使用纯 mock 结果伪造通过。
- tryrun 只写 `monitor_alert_eval_logs` 评估日志，不写 `monitor_alert_events_current`，不写 `monitor_alert_events_history`，不制造正式告警事件。
- tryrun 响应保留 P0 兼容字段，例如 `ok`、`status`、`checks`、`rule`、`eval_log`，新增 evaluator 字段不得破坏旧前端或测试调用方。
- Prometheus 上游断连、超时、认证失败或 5xx 失败，按外部依赖失败返回 503 或等价安全失败状态；不得包装为成功，也不得生成错误告警风暴。
- eval log 可以记录规则 ID、数据源 ID、状态、耗时、结果规模、trace_id 和脱敏错误摘要；不得记录 raw PromQL、完整 URL、token、Cookie、DSN、Authorization 头或完整内部错误。
- 响应和日志中的错误信息必须面向用户可读且脱敏，不能回显 Prometheus 原始认证错误、完整请求 URL 或堆栈。

### 4.7 版本与回滚

`GET /api/v1/monitor/alert-rules/:id/versions`

返回 `MonitorAlertRuleVersion[]`。

`POST /api/v1/monitor/alert-rules/:id/rollback`

请求：

```json
{
  "version": 3,
  "reason": "回滚误改阈值"
}
```

约束：

- 回滚应创建新版本，而不是覆盖旧版本。
- 回滚必须记录操作者、目标版本、原当前版本、原因和 trace_id。
- 回滚失败不得改变当前规则。

### 4.8 导入与导出

`POST /api/v1/monitor/alert-rules/import`

用途：导入 FindX 规则包或从兼容入口转换后的规则。

要求：

- 必须校验 schema、名称冲突、数据源存在性、标签合法性。
- 默认 dry-run，明确传入 `apply: true` 才落库。
- 导入报告不得包含真实 token、DSN 或内部完整 URL。

`GET /api/v1/monitor/alert-rules/export`

用途：导出规则包，用于备份、迁移、评审和回滚。

要求：

- 导出内容必须脱敏。
- 导出包应包含 schema_version、rules、generated_at、generated_by。

## 5. P0-2 已通过 QA 门禁 API：Events 全生命周期

P0-2 后端 API 已覆盖 current event、history event 与事件处置。AI 诊断、Agent 巡检和修复计划入口属于后续阶段边界，本文仅保留契约约束，不按 P0 已上线能力描述。

### 5.1 路由清单

```text
GET  /api/v1/monitor/events/current
GET  /api/v1/monitor/events/history
GET  /api/v1/monitor/events/:id
POST /api/v1/monitor/events/:id/ack
POST /api/v1/monitor/events/:id/assign
POST /api/v1/monitor/events/:id/mute
POST /api/v1/monitor/events/:id/resolve
POST /api/v1/monitor/events/:id/archive
POST /api/v1/monitor/events/:id/diagnose
POST /api/v1/monitor/events/:id/inspect
POST /api/v1/monitor/events/:id/remediation-plan
GET  /api/v1/monitor/events/:id/actions
```

权限：

- `GET current/history/:id/actions`：读接口。
- `ack/assign/mute/resolve/archive`：`adminRequired` 或事件处置权限。
- `diagnose/inspect`：只读诊断可读权限；触发远程 Agent 操作时必须按 Agent 能力和风险提升权限。
- `remediation-plan`：`adminRequired`，只创建计划，不直接执行修复；执行链路另属 remediation 契约。

### 5.2 MonitorAlertEvent 字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | Event ID。 |
| `rule_id` | string | 否 | 关联规则 ID。 |
| `rule_version` | number | 否 | 触发时规则版本。 |
| `event_key` | string | 是 | 事件业务键。 |
| `name` | string | 是 | 事件名称。 |
| `severity` | string | 是 | 严重级别。 |
| `status` | string | 是 | `firing`、`acknowledged`、`muted`、`resolved`、`archived`。 |
| `datasource_id` | string | 否 | 数据源 ID。 |
| `target_id` | string | 否 | Target ID。 |
| `target_ident` | string | 否 | Target 标识。 |
| `labels` | object | 否 | 事件标签。 |
| `annotations` | object | 否 | 摘要、描述、Runbook、AI hints。 |
| `value` | string | 否 | 触发值，建议字符串化避免精度和单位歧义。 |
| `fingerprint` | string | 是 | 去重指纹。 |
| `count` | number | 是 | 聚合次数。 |
| `first_seen` | string | 是 | 首次触发时间。 |
| `last_seen` | string | 是 | 最近触发时间。 |
| `ack_by` | string | 否 | 确认人。 |
| `assignee` | string | 否 | 负责人。 |
| `resolution` | string | 否 | 恢复或关闭说明。 |
| `archived_at` | string | 否 | 归档时间。 |
| `resolved_at` | string | 否 | 恢复时间。 |
| `action_log` | array | 否 | 处置动作记录。 |
| `created_at` | string | 是 | 创建时间。 |
| `updated_at` | string | 是 | 更新时间。 |

### 5.3 当前事件与历史事件

`GET /api/v1/monitor/events/current`

用途：查询当前未归档事件，覆盖 firing、acknowledged、muted 等状态。

`GET /api/v1/monitor/events/history`

用途：查询历史事件，覆盖 resolved、archived 等状态。

查询参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `status` | string | 否 | 事件状态。 |
| `severity` | string | 否 | 严重级别。 |
| `rule_id` | string | 否 | 规则 ID。 |
| `target_id` | string | 否 | Target ID。 |
| `target_ident` | string | 否 | Target 标识。 |
| `datasource_id` | string | 否 | 数据源 ID。 |
| `from` | string | 否 | 起始时间。 |
| `to` | string | 否 | 结束时间。 |
| `q` | string | 否 | 名称、标签、摘要搜索。 |
| `page` | number | 否 | 页码，默认 1。 |
| `page_size` | number | 否 | 每页数量，默认 20，最大 100。 |

建议响应：

```json
{
  "items": [],
  "page": 1,
  "page_size": 20,
  "total": 0
}
```

### 5.4 事件详情

`GET /api/v1/monitor/events/:id`

响应：`MonitorAlertEvent`。

要求：

- 详情必须包含 action_log 或提供 actions 链接。
- 不得暴露查询原始错误中的 SQL、DSN、token、内部路径。
- 对不可见资源返回 404 或 403，按权限模型统一。

### 5.5 事件处置动作

通用动作请求字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `reason` | string | 否 | 操作原因。 |
| `trace_id` | string | 否 | 前端或调用方传入的追踪 ID。 |

`POST /api/v1/monitor/events/:id/ack`

用途：确认事件。

请求：

```json
{
  "reason": "值班人员已接手"
}
```

`POST /api/v1/monitor/events/:id/assign`

用途：分派事件。

请求：

```json
{
  "assignee": "ops-user",
  "reason": "交由业务值班处理"
}
```

`POST /api/v1/monitor/events/:id/mute`

用途：临时静默事件。

请求：

```json
{
  "duration": "30m",
  "reason": "发布窗口内已知波动"
}
```

`POST /api/v1/monitor/events/:id/resolve`

用途：人工标记恢复。

请求：

```json
{
  "resolution": "扩容后指标恢复",
  "reason": "人工确认"
}
```

`POST /api/v1/monitor/events/:id/archive`

用途：归档事件。

请求：

```json
{
  "reason": "事件已复盘并关闭"
}
```

响应：更新后的 `MonitorAlertEvent`。

状态机要求：

- `firing` 可进入 `acknowledged`、`muted`、`resolved`。
- `acknowledged` 可进入 `muted`、`resolved`、`assign` 后保持处置中。
- `muted` 到期后可回到 `firing` 或因恢复进入 `resolved`。
- `resolved` 可进入 `archived`。
- `archived` 不允许再次处置，只允许查询。

P1-BE-1 fingerprint 与事件状态机已落库契约：

- 提交记录：`82ea50d feat: harden FindX alert event fingerprint state machine`。
- `fingerprint` 由服务端按规则、目标、事件业务键和非敏感标签等稳定字段生成，不信任外部请求或上游传入的 fingerprint。
- 服务端生成 fingerprint 时必须忽略敏感 labels/metadata，且 fingerprint 字符串本身不得泄露 token、Cookie、Authorization、DSN、密钥、密码或完整内部 URL。
- current event upsert 以服务端规范化 fingerprint 为幂等键；同一事件重复 firing 应更新 `count`、`last_seen`、`value`、labels/annotations 的脱敏结果，而不是制造重复 current event。
- 不同规则、不同目标或不同非敏感业务标签必须生成不同 fingerprint，避免错误合并无关事件。
- current event 进入 `resolved` 或 `archived` 等终态后受终态保护，后续 upsert 不得把终态事件悄悄改回 firing；如需新一轮事件，必须由服务端生成新事件生命周期。
- 事件响应、action log、eval log 和持久化前的 labels/metadata 需要执行敏感信息脱敏；不得把 raw label secret、完整 URL、token、Cookie、DSN 或内部路径暴露给前端、日志、审计或 AI prompt。

### 5.6 诊断、巡检、修复计划入口

`POST /api/v1/monitor/events/:id/diagnose`

用途：基于事件、Target、指标和上下文触发 AI 诊断。

请求：

```json
{
  "question": "分析本次 CPU 告警原因",
  "include_logs": false,
  "include_agent_evidence": true
}
```

响应建议：

```json
{
  "ok": true,
  "diagnosis_id": "diag-001",
  "status": "queued",
  "evidence_refs": ["event:event-001", "target:target-001"]
}
```

`POST /api/v1/monitor/events/:id/inspect`

用途：触发关联 Agent 巡检或采证。

请求：

```json
{
  "plugins": ["cpu", "process"],
  "timeout_seconds": 60
}
```

响应建议：

```json
{
  "ok": true,
  "inspection_id": "inspect-001",
  "status": "queued"
}
```

`POST /api/v1/monitor/events/:id/remediation-plan`

用途：基于事件创建自动修复计划草案，不直接执行。

请求：

```json
{
  "strategy": "safe",
  "dry_run": true,
  "reason": "生成修复建议供审批"
}
```

响应建议：

```json
{
  "ok": true,
  "plan_id": "remediation-plan-001",
  "status": "draft",
  "requires_approval": true
}
```

要求：

- 诊断结论必须绑定 evidence_refs。
- 巡检必须校验 Agent 在线状态、能力列表和超时。
- 修复计划必须默认 dry-run 或 draft，不允许绕过审批直接执行。

### 5.7 事件动作日志

`GET /api/v1/monitor/events/:id/actions`

响应：`MonitorAlertAction[]`。

`MonitorAlertAction` 字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | Action ID。 |
| `event_id` | string | 是 | Event ID。 |
| `action` | string | 是 | `ack`、`assign`、`mute`、`resolve`、`archive`、`diagnose`、`inspect`、`remediation_plan`。 |
| `actor` | string | 否 | 操作者。 |
| `from` | string | 否 | 原状态或原值。 |
| `to` | string | 否 | 新状态或新值。 |
| `reason` | string | 否 | 操作原因。 |
| `assignee` | string | 否 | 分派对象。 |
| `trace_id` | string | 否 | 追踪 ID。 |
| `created_at` | string | 是 | 动作时间。 |

## 6. P1 后端告警闭环契约：Evaluator 与 Scheduler

本章节只记录 P1-BE-1/2/3 已落库的后端告警闭环能力，不代表 P1 Dashboard、通知、权限、前端页面或完整 AI/修复链路已完成。

### 6.1 P1-BE 提交与状态

| 子任务 | 提交 | 当前状态 | QA/验证 |
| --- | --- | --- | --- |
| P1-BE-1 | `82ea50d feat: harden FindX alert event fingerprint state machine` | 已落库；服务端 fingerprint、current upsert 幂等、终态保护和脱敏语义生效 | 纳入后端测试与回归范围。 |
| P1-BE-2 | `1b37905 feat: add FindX alert rule evaluator tryrun` | 已落库；tryrun 真实查询 Prometheus，只写 eval log，不写 current/history | 纳入后端测试与回归范围。 |
| P1-BE-3 | `1c4045e feat: add FindX alert scheduler event closure` | 已落库；scheduler 默认禁用，支持 RunOnce、并发锁、firing/no_data/ok/keep_state 闭环 | QA **92/100 PASS**。 |

### 6.2 Scheduler 配置与运行边界

P1-BE-3 scheduler 后端能力由配置显式开启，默认禁用：

```yaml
monitoring:
  alert_scheduler:
    enabled: false
    interval_seconds: 60
    timeout_seconds: 5
    max_concurrency: 4
```

配置语义：

| 配置项 | 类型 | 默认/边界 | 说明 |
| --- | --- | --- | --- |
| `monitoring.alert_scheduler.enabled` | boolean | 默认 `false` | 只有显式开启才启动后台循环，避免升级后自动制造事件。 |
| `monitoring.alert_scheduler.interval_seconds` | number | 默认 60，最小 10 | 后台循环间隔。 |
| `monitoring.alert_scheduler.timeout_seconds` | number | 默认 5，范围 1-30 | 单次规则评估超时。 |
| `monitoring.alert_scheduler.max_concurrency` | number | 默认 4，范围 1-16 | 单次调度内最大并发规则数。 |

运行约束：

- `RunMonitorAlertEvaluationOnce` 可用于单次测试、集成验证和手动回归，不要求启动后台循环。
- scheduler 必须持有并发锁；同一进程内并发 RunOnce 或后台 tick 不得重叠执行。
- scheduler 只评估启用且可评估的数据源/规则；禁用规则不得制造或恢复事件。
- 每条规则评估必须写 eval log，记录状态、耗时、规则 ID、数据源 ID、脱敏错误摘要和 trace_id/actor 等可追踪字段。
- eval log 写入失败时，本轮不得继续写 current/history event，避免出现无评估证据的事件。
- 查询或评估失败只写 eval log，不制造 current event，也不恢复 existing current event；不得把 Prometheus 上游失败解释成业务恢复。
- scheduler 日志、eval log 和响应不得包含 raw PromQL、完整 URL、token、Cookie、DSN、Authorization 头或完整内部错误。

### 6.3 Scheduler 事件闭环语义

| evaluator 状态 | current 写入 | history 写入 | 说明 |
| --- | --- | --- | --- |
| `firing` | 写入或幂等更新 current event | 不写 history | 使用服务端 fingerprint upsert；重复 firing 只更新同一 current 生命周期。 |
| `no_data_alerting` | 写入或幂等更新 current event | 不写 history | 无数据策略要求告警时，以 `value=no_data` 或等价字段写入 current。 |
| `ok` | 不新增 current | 将同规则已不活跃 current 恢复到 history | 恢复时写 `resolved_at`、`resolution` 和 scheduler action log。 |
| `no_data_ok` | 不新增 current | 将同规则已不活跃 current 恢复到 history | 无数据策略视为恢复时执行事件闭环。 |
| `keep_state` / `no_data_keep_state` | 不新增 current | 不恢复 history | 保持现有状态，不制造事件，也不关闭 current。 |
| `error` / query failed | 不新增 current | 不恢复 history | 只写失败 eval log；Prometheus 上游失败按 503/安全失败处理。 |

恢复规则：

- 恢复到 history 时必须按 fingerprint 判断当前活跃集合，避免误恢复仍在 firing 的事件。
- 恢复动作必须保留原事件主体信息、count、first_seen、last_seen、labels/annotations 脱敏结果和 action log。
- scheduler 自动恢复的 action actor 使用系统调度身份，不伪装成真实用户。
- `keep_state` 不恢复旧事件、不新增新事件，也不覆盖人工处置状态。

### 6.4 P1-BE-3 验证证据

P1-BE-3 提交记录为 `1c4045e feat: add FindX alert scheduler event closure`，QA 评分 **92/100**，结论 **PASS**。

已通过验证：

```powershell
cd D:\ai-workbench\api && go test -count=1 ./...
cd D:\ai-workbench\api && go build ./...
```

```bash
cd /opt/ai-workbench/api && go test -count=1 ./...
cd /opt/ai-workbench/api && go build -o api-linux .
```

验证覆盖：

- Windows 后端全量测试与构建通过。
- WSL `/opt/ai-workbench/api` 后端全量测试与 Linux 构建通过。
- scheduler 默认禁用、配置边界、RunOnce、并发锁、firing/no_data_alerting 写 current、ok/no_data_ok 恢复到 history、keep_state 保持状态、失败只写 eval log 等路径纳入验证证据。

## 7. 测试清单

测试结果状态只能使用 PASS、FAIL、BLOCKED、NOT_RUN、RISK。未真实执行不得写 PASS。

### 7.1 P0-1 API 测试

| case_id | 层 | 场景 | 优先级 | 期望 |
| --- | --- | --- | --- | --- |
| QA-API-FINDX-P0-1-001 | API | `GET /api/v1/monitor/health` 正常返回 | P0 | 200；包含 `status/mode/storage/targets/agents/agent_online/generated_at`；`mode=findx-core`。 |
| QA-API-FINDX-P0-1-002 | API | MySQL 不可用或存储降级 | P0 | 200 或 503 按实现约定返回可读降级状态；不得白屏或误报完全健康。 |
| QA-API-FINDX-P0-1-003 | API | `GET /api/v1/monitor/targets` 无过滤 | P0 | 200；返回数组；字段符合 `MonitorTarget`。 |
| QA-API-FINDX-P0-1-004 | API | `GET /api/v1/monitor/targets?status=online` | P1 | 200；只返回匹配状态数据。 |
| QA-API-FINDX-P0-1-005 | API | `POST /api/v1/monitor/targets` 正常创建 | P0 | admin 用户 200；返回 Target；至少可用 `ident/ip/hostname` 之一创建。 |
| QA-API-FINDX-P0-1-006 | API | Target 创建字段缺失 | P0 | 400；`ident/ip/hostname` 全空时失败。 |
| QA-API-FINDX-P0-1-007 | API | Target 创建非法 IP | P0 | 400；错误信息可读且不暴露内部路径。 |
| QA-API-FINDX-P0-1-008 | API | 非 admin 写 Target | P0 | 403；不得创建或修改数据。 |
| QA-API-FINDX-P0-1-009 | API | `GET /api/v1/findx-agents` | P0 | 200；返回数组；字段符合 `FindXAgent`。 |
| QA-API-FINDX-P0-1-010 | API | Agent register 携带正确 `X-Agent-Token` 正常注册 | P0 | 200；返回 `ok=true`、`agent`、`target`。 |
| QA-API-FINDX-P0-1-011 | API | Agent heartbeat 携带正确 `X-Agent-Token` 正常上报 | P0 | 200；更新 `last_seen`；关联 Target 状态为在线或符合实现状态。 |
| QA-API-FINDX-P0-1-012 | API | Agent heartbeat 非法 JSON | P0 | 400；返回脱敏错误。 |
| QA-API-FINDX-P0-1-013 | API | Agent heartbeat 非法 IP | P0 | 400；不得写入 Agent 或 Target。 |
| QA-API-FINDX-P0-1-014 | API | Agent register/heartbeat 缺失 `X-Agent-Token` | P0 | 401；不得写入 Agent、Target 或心跳。 |
| QA-API-FINDX-P0-1-015 | API | Agent register/heartbeat 使用错误 token | P0 | 401；响应不得提示真实 token 或配置项值，不得写入数据。 |
| QA-API-FINDX-P0-1-016 | API | 未配置 Agent token 且 `findx_agents.allow_anonymous=false` 或缺省 | P0 | register/heartbeat 默认拒绝匿名写入，返回 401；不得因 token 未配置而自动放行。 |
| QA-SEC-FINDX-P0-1-017 | 安全 | 响应与日志敏感信息扫描 | P0 | 不出现真实 token、Cookie、完整 DSN、SSH key、堆栈、本地敏感路径。 |
| QA-REG-FINDX-P0-1-018 | 回归 | 旧 `/api/v1/n9e/*` 与 `/api/v1/catpaw/*` 兼容入口 | P1 | 旧入口仍按兼容目标可访问；新主线入口不依赖旧入口成功。 |
| QA-API-FINDX-P0-1-019 | API | 显式设置 `findx_agents.allow_anonymous=true` 的本地/测试环境 | P1 | register/heartbeat 可无 token 接入；测试报告必须注明仅限本地/测试环境，生产配置不得开启。 |

### 7.2 P0-2 Alert Rules 测试

| case_id | 层 | 场景 | 优先级 | 期望 |
| --- | --- | --- | --- | --- |
| QA-API-FINDX-P0-2-RULE-001 | API | 创建合法告警规则 | P0 | admin 用户 200/201；返回 Rule；`version=1` 或符合实现约定。 |
| QA-API-FINDX-P0-2-RULE-002 | API | 创建缺少 `name/query/severity/datasource_id` | P0 | 400；不得写入。 |
| QA-API-FINDX-P0-2-RULE-003 | API | 非法 `severity/no_data_policy/for_duration` | P0 | 400；错误可读。 |
| QA-API-FINDX-P0-2-RULE-004 | API | 非 admin 创建、更新、删除规则 | P0 | 403；无数据变更。 |
| QA-API-FINDX-P0-2-RULE-005 | API | 规则列表分页与过滤 | P1 | `page/page_size/total/items` 正确；最大 page_size 生效。 |
| QA-API-FINDX-P0-2-RULE-006 | API | 更新规则生成新版本 | P0 | 规则版本自增；旧版本可查；审计记录存在。 |
| QA-API-FINDX-P0-2-RULE-007 | API | 启用/禁用规则 | P0 | 状态变更正确；重复操作幂等或返回明确错误。 |
| QA-API-FINDX-P0-2-RULE-008 | API | 克隆规则 | P1 | 创建新 ID；默认禁用或按请求设置；不覆盖原规则。 |
| QA-API-FINDX-P0-2-RULE-009 | API | tryrun 成功 | P0 | 返回 `ok/status/checks/eval_log`；不生成正式 current event。 |
| QA-API-FINDX-P0-2-RULE-010 | API | tryrun 数据源断连 | P0 | 503 或 `status=fail`；返回脱敏错误；不写正式事件。 |
| QA-API-FINDX-P0-2-RULE-011 | API | 回滚到历史版本 | P0 | 生成新版本；当前规则恢复目标内容；审计记录完整。 |
| QA-API-FINDX-P0-2-RULE-012 | API | 回滚不存在版本 | P0 | 404 或 400；当前规则不变。 |
| QA-API-FINDX-P0-2-RULE-013 | API | 导入规则 dry-run | P1 | 只返回校验报告；不落库。 |
| QA-API-FINDX-P0-2-RULE-014 | API | 导出规则脱敏 | P1 | 导出包不包含 token、DSN、Cookie、内部完整 URL。 |

### 7.3 P0-2 Events 测试

| case_id | 层 | 场景 | 优先级 | 期望 |
| --- | --- | --- | --- | --- |
| QA-API-FINDX-P0-2-EVENT-001 | API | 当前事件列表 | P0 | 200；返回分页结构；包含 firing/acknowledged/muted 等当前状态。 |
| QA-API-FINDX-P0-2-EVENT-002 | API | 历史事件列表 | P0 | 200；返回 resolved/archived 历史数据。 |
| QA-API-FINDX-P0-2-EVENT-003 | API | 事件详情 | P0 | 200；字段符合 `MonitorAlertEvent`；action_log 可查。 |
| QA-API-FINDX-P0-2-EVENT-004 | API | 不存在事件详情 | P0 | 404；错误脱敏。 |
| QA-API-FINDX-P0-2-EVENT-005 | API | ack 事件 | P0 | 状态和 `ack_by` 正确；动作日志记录。 |
| QA-API-FINDX-P0-2-EVENT-006 | API | assign 事件 | P0 | `assignee` 正确；动作日志记录。 |
| QA-API-FINDX-P0-2-EVENT-007 | API | mute 事件 | P0 | 静默时长合法；状态正确；到期策略明确。 |
| QA-API-FINDX-P0-2-EVENT-008 | API | resolve 事件 | P0 | `resolved_at/resolution` 正确；动作日志记录。 |
| QA-API-FINDX-P0-2-EVENT-009 | API | archive 事件 | P0 | `archived_at` 正确；归档后不允许再次处置。 |
| QA-API-FINDX-P0-2-EVENT-010 | API | 非 admin 事件处置 | P0 | 403；事件状态不变。 |
| QA-API-FINDX-P0-2-EVENT-011 | API | 非法状态流转 | P0 | 400；例如 archived 再 ack 应失败。 |
| QA-API-FINDX-P0-2-EVENT-012 | API | diagnose 触发 | P1 | 返回 diagnosis_id/status/evidence_refs；诊断结论必须绑定证据。 |
| QA-API-FINDX-P0-2-EVENT-013 | API | inspect 触发但 Agent 离线 | P0 | 503 或明确失败状态；不得误报 queued 成功。 |
| QA-API-FINDX-P0-2-EVENT-014 | API | remediation-plan 创建 | P0 | 返回 draft 计划；`requires_approval=true`；不直接执行修复。 |
| QA-SEC-FINDX-P0-2-EVENT-015 | 安全 | 事件详情和动作日志脱敏 | P0 | 不包含 token、Cookie、完整 DSN、SSH key、堆栈、本地敏感路径。 |

### 7.4 UI 与兼容回归测试

| case_id | 层 | 场景 | 优先级 | 期望 |
| --- | --- | --- | --- | --- |
| QA-UI-FINDX-P0-1-001 | UI | Monitoring Core 健康页或入口 | P1 | 能展示 `healthy/degraded/empty`；降级态有可读提示。 |
| QA-UI-FINDX-P0-1-002 | UI | Target 列表 | P1 | 正常、空态、错误态、过滤状态完整。 |
| QA-UI-FINDX-P0-1-003 | UI | Agent 列表 | P1 | 显示在线状态、版本、能力、最近心跳。 |
| QA-UI-FINDX-P0-2-001 | UI | Alert Rule 列表与编辑 | P1 | 列表、详情、创建、编辑、启停、tryrun、回滚入口完整。 |
| QA-UI-FINDX-P0-2-002 | UI | Events 当前与历史视图 | P1 | 筛选、详情、处置动作、动作日志、诊断入口完整。 |
| QA-COMPAT-FINDX-001 | 兼容 | `/n9e/*` 旧入口 | P1 | 旧入口作为兼容入口保留；新 UI 不以其作为主入口。 |
| QA-COMPAT-FINDX-002 | 兼容 | `/catpaw/*` 旧入口 | P1 | 旧入口作为兼容入口保留；Agent 新入口使用 `/findx-agents/*`。 |

### 7.5 构建与断连验证

| case_id | 层 | 场景 | 优先级 | 期望 |
| --- | --- | --- | --- | --- |
| QA-BUILD-FINDX-001 | 构建 | WSL 后端编译 | P0 | `/opt/ai-workbench/api` 执行 `go build -o api-linux .` 通过。 |
| QA-BUILD-FINDX-002 | 构建 | WSL 前端构建 | P0 | `/opt/ai-workbench/web` 执行 `npm run build` 通过。 |
| QA-FAULT-FINDX-001 | 断连 | MySQL 断连 | P0 | health 显示 degraded 或 503；规则和事件接口不误报成功。 |
| QA-FAULT-FINDX-002 | 断连 | Prometheus/数据源断连 | P0 | query/tryrun/evaluator 失败可读；不生成错误告警风暴。 |
| QA-FAULT-FINDX-003 | 断连 | Agent 离线 | P0 | Agent 状态更新；inspect/diagnose/remediation 不误报远程执行成功。 |

### 7.6 P0-3 Query Gateway 测试

| case_id | 层 | 场景 | 优先级 | 期望 |
| --- | --- | --- | --- | --- |
| QA-API-FINDX-P0-3-001 | API | 数据源列表查询 | P0 | `GET /api/v1/monitor/datasources` 返回分页结构；字段符合 `MonitorDatasource`；敏感配置只返回脱敏摘要。 |
| QA-API-FINDX-P0-3-002 | API | 创建合法数据源 | P0 | admin 用户 200/201；返回数据源对象；配置密文字段不在响应中回显。 |
| QA-API-FINDX-P0-3-003 | API | 创建缺少 `name/type/url` 或类型非法 | P0 | 400；不得写入数据源。 |
| QA-API-FINDX-P0-3-004 | API | 非 admin 写数据源 | P0 | 403；数据源不变。 |
| QA-API-FINDX-P0-3-005 | API | 即时查询成功 | P0 | `POST /api/v1/monitor/query` 返回 `status=success`、`data`、`stats`；写入查询审计。 |
| QA-API-FINDX-P0-3-006 | API | 区间查询成功 | P0 | `POST /api/v1/monitor/query-range` 支持 `start/end/step`；限制最大区间和最大点数。 |
| QA-API-FINDX-P0-3-007 | API | 非法 PromQL、非法时间范围或超大 step | P0 | 400；错误可读且脱敏；不写成功审计。 |
| QA-API-FINDX-P0-3-008 | API | 数据源断连、超时或上游 5xx | P0 | 503；返回降级状态；不误报成功。 |
| QA-API-FINDX-P0-3-009 | API | metrics/labels/label-values 查询 | P1 | 支持按数据源、label matcher、时间范围过滤；返回数组或分页结构。 |
| QA-SEC-FINDX-P0-3-010 | 安全 | 查询网关脱敏扫描 | P0 | 响应、审计、错误和日志不包含 token、Cookie、完整 DSN、SSH key、内部完整 URL 或原始堆栈。 |

### 7.7 P1-BE 后端告警闭环测试

| case_id | 层 | 场景 | 优先级 | 期望 |
| --- | --- | --- | --- | --- |
| QA-API-FINDX-P1-BE-1-001 | API/Store | 外部传入不同 fingerprint 但规范化事件相同 | P0 | 服务端忽略外部 fingerprint，按规范化 fingerprint 幂等 upsert current，`count/last_seen` 更新，不生成重复事件。 |
| QA-API-FINDX-P1-BE-1-002 | API/Store | 不同规则、目标或非敏感业务标签 | P0 | 生成不同 fingerprint，不错误合并无关事件。 |
| QA-API-FINDX-P1-BE-1-003 | API/Store | resolved/archived 终态事件再次 upsert | P0 | 终态受保护，不被悄悄改回 firing。 |
| QA-SEC-FINDX-P1-BE-1-004 | 安全 | labels/metadata 含 token、Cookie、DSN、Authorization、完整 URL | P0 | fingerprint、响应、日志、action log 和 eval log 均不泄露敏感原文。 |
| QA-API-FINDX-P1-BE-2-001 | API | `POST /api/v1/monitor/alert-rules/:id/tryrun` 成功查询 Prometheus | P0 | 返回兼容字段；写 eval log；不写 current/history。 |
| QA-API-FINDX-P1-BE-2-002 | API | tryrun Prometheus 上游断连、超时、认证失败或 5xx | P0 | 返回 503 或安全失败状态；只写失败 eval log；不生成事件；错误脱敏。 |
| QA-SEC-FINDX-P1-BE-2-003 | 安全 | tryrun 响应、日志、eval log 脱敏扫描 | P0 | 不包含 raw PromQL、完整 URL、token、Cookie、DSN、Authorization 头或完整内部错误。 |
| QA-SCHED-FINDX-P1-BE-3-001 | Scheduler | `monitoring.alert_scheduler.enabled` 默认值 | P0 | 默认禁用，不启动后台循环，不自动制造事件。 |
| QA-SCHED-FINDX-P1-BE-3-002 | Scheduler | 配置边界 | P0 | `interval_seconds`、`timeout_seconds`、`max_concurrency` 按默认值和上下界收敛。 |
| QA-SCHED-FINDX-P1-BE-3-003 | Scheduler | RunOnce 单次评估 | P0 | 可在测试中直接执行；返回规则数、成功/失败、事件创建/恢复等统计。 |
| QA-SCHED-FINDX-P1-BE-3-004 | Scheduler | 并发 RunOnce 或 tick 重叠 | P0 | 并发锁生效，重叠执行被拒绝或跳过，不重复写事件。 |
| QA-SCHED-FINDX-P1-BE-3-005 | Scheduler | evaluator `firing` | P0 | 写入或幂等更新 current event，按服务端 fingerprint 去重。 |
| QA-SCHED-FINDX-P1-BE-3-006 | Scheduler | evaluator `no_data_alerting` | P0 | 写入 current event，明确无数据告警语义。 |
| QA-SCHED-FINDX-P1-BE-3-007 | Scheduler | evaluator `ok` 或 `no_data_ok` | P0 | 将已不活跃 current 恢复到 history，写恢复时间、原因和 action log。 |
| QA-SCHED-FINDX-P1-BE-3-008 | Scheduler | evaluator `keep_state` 或 `no_data_keep_state` | P0 | 不恢复、不新增、不覆盖人工处置状态。 |
| QA-SCHED-FINDX-P1-BE-3-009 | Scheduler | 查询失败或 eval log 写入失败 | P0 | 查询失败只写失败 eval log；eval log 写入失败则不写 event。 |
| QA-BUILD-FINDX-P1-BE-3-010 | 构建 | Windows 与 WSL 后端验证 | P0 | Windows `go test -count=1 ./...`、Windows `go build ./...`、WSL `go test -count=1 ./...`、WSL `go build -o api-linux .` 均通过。 |

### 7.8 P1 后续风险回归项（含非阻断/阻断）

| case_id | 层 | 场景 | 优先级 | 当前状态 | 期望 |
| --- | --- | --- | --- | --- | --- |
| QA-RISK-FINDX-P1-001 | API/性能 | 大基数候选截断语义 | P2 | RISK | 明确截断、分页、limit、提示字段和审计记录，避免静默丢数据。 |
| QA-RISK-FINDX-P1-002 | API/Workflow | Workflow API 严格 DSL 400 回归 | P2 | RISK | 创建/更新坏 DSL 返回 400，且不持久化；覆盖 `ParseDSL` 后 `graph.Validate()`。 |
| QA-RISK-FINDX-P1-003 | Store | 持久化失败一致性 | P2 | RISK | eval log、current、history 任一失败时状态一致，避免无证据事件或半恢复。 |
| QA-RISK-FINDX-P1-004 | API/安全 | P1 扩展 API 权限与脱敏 | P1 | RISK | dashboard、template、notification、silence、subscription、oncall、pipeline、task、permission、audit 的读写权限、租户/团队边界、审计、错误脱敏和敏感字段脱敏均需逐接口验证；未验证前阻断 P1 扩展能力标记为完成。 |
| QA-RISK-FINDX-P1-005 | 工程治理 | 统一 `.gitattributes` | P3 | RISK | 统一跨 Windows/WSL 的换行与文本属性，降低文档和代码 diff 噪声。 |

## 8. 状态闭环与 P0-3 查询网关契约

### 8.1 P0 与 P1 后端状态闭环

| 子阶段 | 范围 | 当前状态 | 已验证门禁 |
| --- | --- | --- | --- |
| P0-1 | target、datasource 基础语义、agent register、agent heartbeat、health | P0 后端 API 已通过 QA 门禁；target 与 agent heartbeat 主线已落库；datasource 基础语义由 P0-3 查询网关统一承载 | health 降级、target CRUD、agent token、heartbeat upsert、权限、断连、脱敏、WSL 后端编译。 |
| P0-2 | alert rule、current/history event、tryrun、rollback、event action | P0 后端 API 已通过 QA 门禁；告警规则、当前/历史事件、tryrun、回滚和事件动作已落库 | 规则 CRUD、版本自增、回滚生成新版本、tryrun 不落正式事件、事件状态机、动作审计、权限、断连、脱敏。 |
| P0-3 | datasource、query、query-range、metrics、labels、label-values | P0 后端 API 已通过 QA 门禁；查询网关已作为 evaluator、dashboard、AI 问诊和修复验证的统一后端入口落库 | 数据源配置脱敏、PromQL 校验、时间范围限制、Prometheus 上游失败 503、查询审计只记 hash/stats、权限、错误模型、与 P0-2 tryrun/evaluator 对接。 |
| P1-BE-1 | fingerprint、current upsert、事件状态机、labels/metadata 脱敏 | P1 后端告警闭环已部分完成；服务端 fingerprint 和终态保护已落库 | 服务端生成 fingerprint、不信任外部传入、current upsert 幂等、终态保护、敏感 labels/metadata 脱敏。 |
| P1-BE-2 | evaluator、真实 Prometheus tryrun、eval log | P1 后端告警闭环已部分完成；tryrun 真实查询 Prometheus，仅写 eval log | Prometheus 上游失败安全 503、tryrun 不写 current/history、兼容字段保留、raw PromQL/URL/token/Cookie/DSN 不进入日志或响应。 |
| P1-BE-3 | alert scheduler、RunOnce、自动写入与恢复闭环 | P1 后端告警闭环已部分完成；scheduler 默认禁用，QA 92/100 PASS | 默认禁用、配置边界、RunOnce、并发锁、firing/no_data_alerting 写 current、ok/no_data_ok 恢复到 history、keep_state 不恢复不新增、失败只写 eval log。 |
| P1 其他 | Dashboard、通知、细粒度权限、前端页面、完整 AI/修复链路 | 待实现，不得视为已完成 | 后续必须补齐 API/UI/权限/断连/脱敏/回归证据后才能改为 PASS。 |

P0 门禁结论已确认为 `PASS`。后续 P1/P2/P3 门禁结论仍只允许：

- `PASS`：构建、API、权限、断连、脱敏和回归均有证据。
- `FAIL`：存在 P0/P1 缺陷、敏感信息风险、误报成功、状态机错误或未授权写入。
- `BLOCKED`：缺少运行环境、依赖不可用或需要真实敏感配置。
- `NOT_RUN`：未执行，不能作为通过依据。
- `RISK`：存在非阻断风险，必须记录 owner、影响面和关闭计划。

### 8.2 P0 验证证据

P0 已完成并推送，提交记录为 `81b4531 feat: add FindX monitoring core P0 APIs`。QA 最终评分为 **98/100**，结论为 **通过**。

Windows 验证：

```powershell
cd D:\ai-workbench\api && go test ./internal/store ./internal/model
cd D:\ai-workbench\api && go test ./internal/handler -run 'TestMonitor|TestSanitizeDatasourceURL'
cd D:\ai-workbench\api && go test main.go main_test.go -run 'TestRequireAdminToken|TestMonitorRead'
cd D:\ai-workbench\api && go build ./...
```

WSL/Linux 验证：

```bash
cd /opt/ai-workbench/api && go test ./internal/store ./internal/model
cd /opt/ai-workbench/api && go test ./internal/handler -run 'TestMonitor|TestSanitizeDatasourceURL'
cd /opt/ai-workbench/api && go test main.go main_test.go -run 'TestRequireAdminToken|TestMonitorRead'
cd /opt/ai-workbench/api && go build -o api-linux .
```

已验证范围：

- monitor health、targets、findx-agents register/heartbeat/list。
- alert-rules、current/history events。
- query gateway datasources/query/query-range/metrics/labels/label-values。
- 读接口平台认证，写接口 admin。
- Agent token 默认拒绝匿名。
- Prometheus 上游失败返回 503。
- 查询审计只记录 hash/stats，不记录原始敏感配置。
- 事件终态保护。

### 8.3 P0-3 路由清单

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

权限：

- `GET datasources/:id/metrics/labels/label-values`：读接口。
- `POST/PUT/DELETE datasources`：`adminRequired`。
- `datasources/:id/test`：`adminRequired`，测试结果必须脱敏。
- `query/query-range`：读权限即可查询授权数据源；高成本查询可按团队、数据源或查询长度提升权限。

### 8.4 MonitorDatasource 字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | 数据源 ID。 |
| `name` | string | 是 | 展示名称，工作区内唯一。 |
| `type` | string | 是 | `prometheus`、`victoriametrics`、`mimir`、`loki`、`elasticsearch`；P0-3 优先实现 Prometheus-compatible 指标查询。 |
| `url` | string | 是 | 数据源基础 URL；响应中只返回脱敏摘要或主机摘要。 |
| `auth_type` | string | 否 | `none`、`bearer`、`basic`、`header`；密文字段只允许写入，不允许回显。 |
| `headers` | object | 否 | 自定义头的脱敏摘要；真实值不得出现在响应、日志和审计详情中。 |
| `timeout_seconds` | number | 是 | 查询超时，建议默认 10，最大 60。 |
| `max_range` | string | 是 | 单次区间查询最大跨度，例如 `24h`。 |
| `max_series` | number | 是 | 单次返回序列上限。 |
| `status` | string | 是 | `enabled`、`disabled`、`degraded`。 |
| `labels` | object | 否 | 数据源标签，例如环境、区域、业务组。 |
| `created_by` | string | 否 | 创建人。 |
| `updated_by` | string | 否 | 更新人。 |
| `created_at` | string | 是 | 创建时间。 |
| `updated_at` | string | 是 | 更新时间。 |

创建请求示例：

```json
{
  "name": "platform-prom",
  "type": "prometheus",
  "url": "https://prometheus.example.internal",
  "auth_type": "bearer",
  "credential": "<TOKEN>",
  "timeout_seconds": 10,
  "max_range": "24h",
  "max_series": 2000,
  "labels": {
    "env": "prod"
  }
}
```

响应不得包含 `credential` 原文：

```json
{
  "id": "ds-001",
  "name": "platform-prom",
  "type": "prometheus",
  "url": "https://prometheus.example.internal",
  "auth_type": "bearer",
  "credential_set": true,
  "timeout_seconds": 10,
  "max_range": "24h",
  "max_series": 2000,
  "status": "enabled",
  "labels": {
    "env": "prod"
  },
  "created_at": "2026-05-04T03:30:00+08:00",
  "updated_at": "2026-05-04T03:30:00+08:00"
}
```

### 8.5 POST /api/v1/monitor/query

用途：执行 Prometheus-compatible 即时查询，为规则 tryrun、Dashboard panel、AI 问诊和自动修复 precheck 提供统一入口。

请求：

```json
{
  "datasource_id": "ds-001",
  "query": "up{job=\"node\"}",
  "time": "2026-05-04T03:30:00+08:00",
  "timeout_seconds": 10,
  "dedup": true,
  "trace_id": "req-xxxx"
}
```

响应：

```json
{
  "status": "success",
  "datasource_id": "ds-001",
  "data": {
    "resultType": "vector",
    "result": []
  },
  "stats": {
    "series": 0,
    "elapsed_ms": 120
  },
  "trace_id": "req-xxxx"
}
```

约束：

- `query` 必须非空并限制长度。
- 只允许访问当前用户可读的数据源。
- 查询失败不得包装成 `status=success`。
- 查询审计记录必须包含操作者、数据源、查询摘要、耗时、结果规模、状态和 trace_id；不得记录密钥。

### 8.6 POST /api/v1/monitor/query-range

用途：执行区间查询，服务 Dashboard、规则评估、趋势诊断和修复验证。

请求：

```json
{
  "datasource_id": "ds-001",
  "query": "rate(node_cpu_seconds_total{mode!=\"idle\"}[5m])",
  "start": "2026-05-04T02:30:00+08:00",
  "end": "2026-05-04T03:30:00+08:00",
  "step": "30s",
  "timeout_seconds": 10
}
```

必须校验：

- `start < end`。
- 区间不超过数据源 `max_range`。
- `step` 合法且不会造成超大点数。
- 返回序列不超过 `max_series`，超过时返回 400 或裁剪状态，不能静默丢数据。

### 8.7 GET /metrics、/labels、/label-values

`GET /api/v1/monitor/metrics`

查询参数：`datasource_id`、`q`、`limit`、`from`、`to`。

`GET /api/v1/monitor/labels`

查询参数：`datasource_id`、`match[]`、`from`、`to`、`limit`。

`GET /api/v1/monitor/label-values`

查询参数：`datasource_id`、`label`、`match[]`、`from`、`to`、`limit`。

要求：

- `datasource_id` 必须存在且当前用户可读。
- `label-values` 的 `label` 必填，非法 label 返回 400。
- 上游断连、超时或认证失败返回 503，错误必须脱敏。
- 返回结果必须限制数量，避免前端渲染和后端内存压力。

### 8.8 P0-3 与 P0-2/P1-BE 的对接边界

| 调用方 | 对接方式 | 门禁 |
| --- | --- | --- |
| alert rule tryrun | 通过 `query/query-range` 执行规则表达式 | tryrun 失败不生成正式 current event；审计记录标记 `source=tryrun`。 |
| evaluator | 通过查询网关读取指标 | 数据源断连时写失败 eval log 并返回安全失败；不得生成告警风暴。 |
| scheduler | 通过 evaluator 读取规则状态并写事件闭环 | 默认禁用；RunOnce 可测试；firing/no_data_alerting 写 current；ok/no_data_ok 恢复 history；keep_state 不变；失败只写 eval log。 |
| dashboard panel | 通过查询网关统一取数 | panel 错误态可读；不得泄露上游原始错误。 |
| AI 问诊 | 通过查询网关拉取证据 | AI prompt 只接收脱敏摘要和 evidence ref，不接收密钥或完整内部错误。 |
| remediation verify | 通过查询网关验证修复效果 | verify 失败时进入 rollback 或人工处理，不能自动标记成功。 |

### 8.9 P1 监控扩展 API 契约草案（待实现）

本节回应 P1 扩展 API 审计缺口，仅定义后续实现必须满足的契约草案。以下 dashboard、template、notification、silence、subscription、oncall、pipeline、task、permission、audit 能力均为 **待实现/不得视为已完成**；只有完成后端实现、前端调用、权限验证、断连验证、脱敏验证和 QA 回归后，才能从待实现改为已完成。

通用约束：

- 主线前缀统一使用 `/api/v1/monitor/*`，不得以旧 `/api/v1/n9e/*` 或 `/api/v1/catpaw/*` 作为新功能主入口。
- 列表接口必须支持 `page`、`page_size`、`limit` 或等价分页参数；大基数候选被截断时必须返回 `truncated=true`、`next_cursor` 或明确提示，并写入脱敏审计。
- 写接口必须记录 actor、资源 ID、动作、结果、trace_id 和脱敏差异摘要；不得记录 token、Cookie、完整 DSN、SSH key、完整内部 URL 或原始堆栈。
- 错误模型遵循 2.3：400 输入错误、401 未认证、403 无权限、404 不存在或不可见、409 状态冲突、429/503 限流或外部依赖不可用、500 内部错误；错误正文必须可读且脱敏。
- 涉及通知渠道、执行任务、值班人和权限的响应必须按当前用户可见范围裁剪；不可见资源统一返回 404 或空集合，不泄露存在性。

| 能力 | 路由草案 | 核心字段 | 权限 | 错误模型 | 验收要点 |
| --- | --- | --- | --- | --- | --- |
| dashboard | `GET /api/v1/monitor/dashboards`、`POST /api/v1/monitor/dashboards`、`GET/PUT/DELETE /api/v1/monitor/dashboards/:id`、`POST /api/v1/monitor/dashboards/:id/clone` | `id`、`name`、`folder_id`、`owner_team`、`panels`、`variables`、`time_range`、`datasource_refs`、`version`、`created_by`、`updated_at` | 读接口需 dashboard read；创建/更新/删除/克隆需 dashboard write 或 `adminRequired`；panel 数据查询仍受 datasource 权限约束 | 400 非法 panel/query/time_range；403 无 dashboard 或 datasource 权限；404 dashboard 不存在或不可见；409 版本冲突；503 查询网关不可用 | 待实现/不得视为已完成；panel query 不泄露上游原始错误；导出/克隆不含密钥；并发更新使用 `version` 或 ETag 防覆盖。 |
| template | `GET /api/v1/monitor/templates`、`POST /api/v1/monitor/templates`、`GET/PUT/DELETE /api/v1/monitor/templates/:id`、`POST /api/v1/monitor/templates/:id/render` | `id`、`kind`、`name`、`description`、`variables`、`body`、`scope`、`version`、`checksum` | 读接口需 template read；写接口需 template write 或 `adminRequired`；render 可读但不得产生状态变更 | 400 模板变量缺失或非法；403 无 scope 权限；404 不存在；409 版本冲突；500 渲染内部错误需脱敏 | 待实现/不得视为已完成；render 只返回脱敏结果；模板不得包含真实密钥；删除前检查 dashboard/rule/notification 引用。 |
| notification | `GET /api/v1/monitor/notification-channels`、`POST /api/v1/monitor/notification-channels`、`GET/PUT/DELETE /api/v1/monitor/notification-channels/:id`、`POST /api/v1/monitor/notification-channels/:id/test` | `id`、`name`、`type`、`enabled`、`target`、`secret_set`、`rate_limit`、`routing_labels`、`created_by` | 读接口返回脱敏配置；写与 test 需 notification write 或 `adminRequired`；test 必须审计 | 400 非法渠道配置；401/403 无权限；404 不可见；429 触发限流；503 下游通知服务不可用 | 待实现/不得视为已完成；响应只返回 `secret_set` 和目标摘要；test 不发送真实敏感内容；失败不回显 webhook 完整 URL 或 token。 |
| silence | `GET /api/v1/monitor/silences`、`POST /api/v1/monitor/silences`、`GET/PUT/DELETE /api/v1/monitor/silences/:id`、`POST /api/v1/monitor/silences/:id/expire` | `id`、`matchers`、`starts_at`、`ends_at`、`reason`、`created_by`、`status`、`affected_count` | 读需 silence read；创建/更新/过期需 silence write 或 `adminRequired`；跨团队 matcher 需更高权限 | 400 非法 matcher 或时间窗口；403 越权静默；404 不存在；409 已过期或状态冲突 | 待实现/不得视为已完成；大基数 `affected_count` 必须可分页或标记截断；静默原因不得包含密钥；过期操作幂等。 |
| subscription | `GET /api/v1/monitor/subscriptions`、`POST /api/v1/monitor/subscriptions`、`GET/PUT/DELETE /api/v1/monitor/subscriptions/:id` | `id`、`user_id`、`team_id`、`matchers`、`channels`、`severity_filter`、`schedule_ref`、`enabled` | 用户可管理本人订阅；团队订阅需 team admin；全局订阅需 `adminRequired` | 400 非法 matcher/channel；403 越权订阅；404 channel 或 schedule 不可见；409 重复订阅 | 待实现/不得视为已完成；不得泄露其他用户联系方式；列表按用户/团队裁剪；通知渠道引用必须脱敏。 |
| oncall | `GET /api/v1/monitor/oncall/schedules`、`POST /api/v1/monitor/oncall/schedules`、`GET/PUT/DELETE /api/v1/monitor/oncall/schedules/:id`、`GET /api/v1/monitor/oncall/schedules/:id/current` | `id`、`name`、`team_id`、`timezone`、`rotations`、`escalation_policy`、`current_user_ref` | 读需 oncall read；写需 oncall admin 或 `adminRequired`；`current` 只返回脱敏联系人摘要 | 400 非法时区/轮值规则；403 越权团队；404 schedule 不存在；409 轮值时间冲突 | 待实现/不得视为已完成；不得回显手机号、邮箱等完整个人信息；升级策略变更需审计；跨时区边界需测试。 |
| pipeline | `GET /api/v1/monitor/pipelines`、`POST /api/v1/monitor/pipelines`、`GET/PUT/DELETE /api/v1/monitor/pipelines/:id`、`POST /api/v1/monitor/pipelines/:id/validate` | `id`、`name`、`source`、`stages`、`filters`、`actions`、`enabled`、`version` | 读需 pipeline read；写需 pipeline write 或 `adminRequired`；validate 不得落库 | 400 非法 stage/action；403 越权 action；404 不存在；409 版本冲突；503 依赖校验不可用 | 待实现/不得视为已完成；validate 坏配置返回 400 且不持久化；action 不得绕过 remediation 审批；日志脱敏。 |
| task | `GET /api/v1/monitor/tasks`、`POST /api/v1/monitor/tasks`、`GET /api/v1/monitor/tasks/:id`、`POST /api/v1/monitor/tasks/:id/cancel`、`GET /api/v1/monitor/tasks/:id/logs` | `id`、`type`、`status`、`target_refs`、`params_summary`、`progress`、`result_summary`、`created_by`、`trace_id` | 读需 task read 且目标可见；创建/取消需 task write 或 `adminRequired`；高风险任务必须走 remediation plan | 400 非法参数；403 目标或动作无权限；404 task 不可见；409 已完成不可取消；503 Agent 离线或队列不可用 | 待实现/不得视为已完成；任务日志只返回脱敏摘要；取消幂等；Agent 离线不得误报成功；部分成功需结构化返回。 |
| permission | `GET /api/v1/monitor/permissions/me`、`GET /api/v1/monitor/permissions/resources/:type/:id`、`POST /api/v1/monitor/permissions/check` | `subject`、`resource_type`、`resource_id`、`actions`、`decision`、`reason_code`、`scope` | `me` 当前用户可读；资源权限详情需 security/admin；check 可用于前端能力裁剪但不得泄露不可见资源详情 | 400 非法资源类型；401 未认证；403 无安全管理权限；404 资源不可见 | 待实现/不得视为已完成；前端按钮隐藏不能替代后端鉴权；权限失败不泄露资源存在性；审计 check 聚合统计而非记录敏感上下文。 |
| audit | `GET /api/v1/monitor/audit-logs`、`GET /api/v1/monitor/audit-logs/:id`、`POST /api/v1/monitor/audit-logs/export` | `id`、`actor`、`action`、`resource`、`result`、`diff_summary`、`ip_hash`、`trace_id`、`created_at` | 读需 audit read 或 `adminRequired`；导出需 audit export；普通用户只能看本人低敏动作摘要 | 400 非法过滤或导出范围；403 无审计权限；404 不可见；429 导出限流；503 存储不可用 | 待实现/不得视为已完成；审计正文不得含密钥；导出需脱敏和水印/trace_id；过滤大结果集必须分页与截断提示。 |

### 8.10 Evidence Chain 契约草案（待实现）

本节定义诊断、AI 问诊、修复计划和执行验证必须共享的证据链契约。Evidence Chain 当前仍为 **待实现/不得视为已完成**；P0/P1-BE 已有 `evidence_refs` 字段和查询网关对接边界，但完整证据归档、AI schema、反馈归档和保留策略仍需后续实现与 QA 回归。

#### 8.10.1 EvidenceRef

`EvidenceRef` 是所有诊断、AI 输入、AI 输出、修复计划和执行记录引用证据的最小单元。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `ref_id` | string | 是 | 全局唯一引用，例如 `event:event-001`、`metric:bundle-001:cpu`。 |
| `source_type` | string | 是 | 证据来源枚举，见 8.10.4。 |
| `source_id` | string | 否 | 原始资源 ID；不可见资源需返回脱敏摘要。 |
| `captured_at` | string | 是 | 采集时间，RFC3339。 |
| `time_range` | object | 否 | `{ "start": "...", "end": "..." }`，用于指标、日志、事件窗口。 |
| `digest` | string | 是 | 证据内容摘要 hash，不得由敏感原文直接展示。 |
| `summary` | string | 是 | 面向 AI 与用户的脱敏摘要。 |
| `sensitivity` | string | 是 | `public`、`internal`、`confidential`、`restricted`。 |
| `redaction_status` | string | 是 | `redacted`、`not_needed`、`blocked`。 |
| `retention_until` | string | 否 | 证据保留截止时间。 |

#### 8.10.2 EvidenceBundle

`EvidenceBundle` 将一次诊断或修复验证所需的证据打包，供 AI 输入、人工审查和审计回放使用。

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `bundle_id` | string | 是 | 证据包 ID。 |
| `case_id` | string | 否 | 关联事件、诊断或修复计划 ID。 |
| `refs` | `EvidenceRef[]` | 是 | 证据引用列表。 |
| `summary` | object | 是 | 聚合摘要，包括时间窗、目标、规则、异常指标、缺失证据。 |
| `unavailable_refs` | array | 否 | 不可用证据清单，记录 `source_type`、`reason_code`、`fallback_used`。 |
| `created_by` | string | 是 | `system`、用户 ID 或任务 ID。 |
| `created_at` | string | 是 | 创建时间。 |
| `trace_id` | string | 否 | 请求追踪。 |

证据不可用降级：

- Prometheus、日志源、Agent 或存储不可用时，不得伪造证据；必须在 `unavailable_refs` 中记录脱敏 `reason_code`，例如 `SOURCE_UNAVAILABLE`、`TIMEOUT`、`PERMISSION_DENIED`、`REDACTED_BLOCKED`。
- 证据不足时 AI 输出 `confidence` 不得高于中等阈值，且 `requires_more_evidence=true`。
- 只要关键证据缺失，修复计划不得自动进入可执行状态；必须进入 `manual_required` 或等待补证。

#### 8.10.3 AI 输入摘要与 AI 输出 schema

AI 输入只允许包含脱敏摘要和 evidence ref，不允许包含原始 token、Cookie、完整 DSN、SSH key、完整内部 URL、完整 PromQL 认证上下文、原始堆栈或用户隐私明文。

AI 输入摘要草案：

```json
{
  "case_id": "event-001",
  "event_summary": {
    "severity": "critical",
    "status": "firing",
    "target_ref": "target:target-001",
    "rule_ref": "rule:rule-001"
  },
  "evidence_bundle_id": "bundle-001",
  "evidence_refs": ["event:event-001", "metric:bundle-001:cpu"],
  "unavailable_refs": [
    {
      "source_type": "agent_inspect",
      "reason_code": "SOURCE_UNAVAILABLE",
      "fallback_used": false
    }
  ],
  "constraints": {
    "no_secret": true,
    "no_direct_execute": true,
    "requires_human_approval_for_remediation": true
  }
}
```

AI 输出 schema 草案：

```json
{
  "diagnosis_id": "diag-001",
  "status": "completed",
  "summary": "CPU saturation is likely caused by a known batch job.",
  "root_causes": [
    {
      "title": "batch job CPU spike",
      "confidence": 0.72,
      "evidence_refs": ["metric:bundle-001:cpu"]
    }
  ],
  "recommended_actions": [
    {
      "action_type": "manual_check",
      "risk_level": "low",
      "requires_approval": false,
      "evidence_refs": ["event:event-001"]
    }
  ],
  "requires_more_evidence": false,
  "safety_notes": ["no direct remediation executed"]
}
```

#### 8.10.4 证据来源枚举、保留与反馈归档

证据来源枚举：

| 枚举 | 说明 |
| --- | --- |
| `alert_event` | 当前/历史告警事件。 |
| `alert_rule` | 告警规则、版本和 tryrun/evaluator 结果。 |
| `query_gateway` | `query`、`query-range`、metrics、labels、label-values 的脱敏结果摘要。 |
| `datasource_metadata` | 数据源类型、状态、能力和脱敏配置摘要。 |
| `target_inventory` | Target、Agent、主机基础信息和状态。 |
| `agent_heartbeat` | Agent 心跳与能力摘要。 |
| `agent_inspect` | Agent 巡检或诊断输出的脱敏摘要。 |
| `workflow_run` | Workflow 执行状态、步骤摘要和脱敏错误。 |
| `remediation_run` | 修复执行、验证和回滚结果摘要。 |
| `operator_feedback` | 人工确认、驳回、补充事实和复盘结论。 |

保留策略：

- 默认证据摘要至少保留 180 天；高敏原始证据如必须留存，应独立加密或托管，本文档不定义真实密钥或存储位置。
- `restricted` 级证据默认只保留摘要和 digest；原始内容不得进入 AI prompt、前端普通详情、测试报告或导出文件。
- 证据删除或过期后，引用仍需保留 `ref_id`、`digest`、`retention_until` 和不可用原因，保证审计链可解释。

敏感字段脱敏：

- 字段名命中 `token`、`secret`、`password`、`authorization`、`cookie`、`dsn`、`credential`、`private_key`、`webhook` 等敏感语义时，默认替换为 `<REDACTED>` 或布尔摘要。
- URL 只允许保留 scheme、host 摘要和路径模板；不得保留完整 query、认证信息或内部路径细节。
- 错误摘要只允许保留错误类别、状态码、依赖名和 trace_id，不得保留原始堆栈或完整请求体。

诊断反馈归档：

- 人工反馈必须写入 `operator_feedback` 证据来源，字段至少包含 `feedback_id`、`diagnosis_id`、`decision`、`comment_summary`、`evidence_refs`、`created_by`、`created_at`。
- 反馈正文必须脱敏；如包含生产数据或隐私信息，应进入 `redaction_status=blocked` 并只保留摘要。
- AI 诊断被人工驳回时，后续相同事件不得直接复用旧结论作为自动修复依据。

### 8.11 Remediation 安全执行契约草案（待实现）

本节定义自动修复的安全执行边界。Remediation 全链路当前仍为 **待实现/不得视为已完成**；现有事件入口 `POST /api/v1/monitor/events/:id/remediation-plan` 只允许创建草案计划，不代表执行链路已经完成。

#### 8.11.1 资源模型

| 资源 | 路由草案 | 核心字段 | 说明 |
| --- | --- | --- | --- |
| `templates` | `GET/POST /api/v1/monitor/remediation/templates`、`GET/PUT/DELETE /api/v1/monitor/remediation/templates/:id` | `id`、`name`、`risk_level`、`action_type`、`inputs_schema`、`precheck`、`rollback_template`、`enabled` | 修复动作模板；高风险模板默认禁用或需审批。 |
| `plans` | `GET/POST /api/v1/monitor/remediation/plans`、`GET/PUT /api/v1/monitor/remediation/plans/:id`、`POST /api/v1/monitor/remediation/plans/:id/submit` | `id`、`event_id`、`status`、`risk_level`、`evidence_bundle_id`、`requires_approval`、`created_by` | 修复计划；提交后进入审批或人工处理。 |
| `plan_steps` | `GET/POST /api/v1/monitor/remediation/plans/:id/steps`、`PUT/DELETE /api/v1/monitor/remediation/plan-steps/:step_id` | `id`、`plan_id`、`order`、`action_type`、`target_refs`、`params_summary`、`risk_level`、`rollback_ref` | 计划步骤；参数必须只保存脱敏摘要和受控引用。 |
| `approvals` | `GET /api/v1/monitor/remediation/approvals`、`POST /api/v1/monitor/remediation/plans/:id/approve`、`POST /api/v1/monitor/remediation/plans/:id/reject` | `id`、`plan_id`、`approver`、`decision`、`reason_summary`、`created_at` | 审批记录；审批人与计划创建人默认不得相同。 |
| `runs` | `POST /api/v1/monitor/remediation/plans/:id/runs`、`GET /api/v1/monitor/remediation/runs`、`GET /api/v1/monitor/remediation/runs/:id`、`POST /api/v1/monitor/remediation/runs/:id/cancel` | `id`、`plan_id`、`status`、`idempotency_key`、`started_by`、`started_at`、`finished_at` | 执行实例；必须具备幂等键、并发锁和可取消语义。 |
| `run_steps` | `GET /api/v1/monitor/remediation/runs/:id/steps`、`POST /api/v1/monitor/remediation/run-steps/:id/retry` | `id`、`run_id`、`plan_step_id`、`status`、`attempt`、`result_summary`、`started_at`、`finished_at` | 执行步骤；重试需遵循幂等和最大次数。 |
| `artifacts` | `GET /api/v1/monitor/remediation/runs/:id/artifacts`、`GET /api/v1/monitor/remediation/artifacts/:id` | `id`、`run_id`、`step_id`、`type`、`digest`、`summary`、`sensitivity`、`retention_until` | 执行产物；默认只返回摘要和 digest。 |
| `audit_logs` | `GET /api/v1/monitor/remediation/audit-logs`、`GET /api/v1/monitor/remediation/audit-logs/:id` | `id`、`actor`、`action`、`resource`、`result`、`diff_summary`、`trace_id`、`created_at` | 修复审计；不得包含真实密钥、完整命令输出敏感内容或完整连接串。 |

权限要求：

- 查看计划和执行记录需 remediation read 且事件/目标可见。
- 创建计划需 remediation plan write；提交、审批、执行、取消和回滚需 remediation operate 或 `adminRequired`。
- 高风险动作、跨团队目标、生产环境写操作必须审批；审批权限不得仅由前端控制。
- 所有执行入口必须校验 `idempotency_key` 或等价幂等标识，防止重复点击或重放请求造成重复执行。

#### 8.11.2 状态机

计划状态：

```text
draft -> precheck_pending -> precheck_passed -> approval_pending -> approved -> executable
draft -> precheck_failed -> manual_required
approval_pending -> rejected -> manual_required
approved -> expired
```

执行状态：

```text
created -> dry_run_pending -> dry_run_passed -> running -> verifying -> succeeded
created -> dry_run_failed -> manual_required
running -> partial_succeeded -> verifying
running -> failed -> rollback_pending -> rollback_running -> rolled_back
running -> failed -> rollback_pending -> rollback_failed -> manual_required
verifying -> verify_failed -> rollback_pending
any_active -> cancelled
```

步骤状态：

```text
pending -> running -> succeeded
pending -> skipped
running -> failed
failed -> retrying -> running
failed -> rollback_pending
rollback_pending -> rollback_succeeded
rollback_pending -> rollback_failed -> manual_required
```

状态机约束：

- 未通过 precheck 的计划不得进入 `approved` 或 `executable`。
- dry-run 失败不得进入真实 execute；必须进入 `manual_required` 或等待修正。
- verify 失败不得标记执行成功；必须进入 rollback 或人工处理。
- rollback 失败必须进入 `manual_required`，不得自动吞掉失败。
- 终态 `succeeded`、`rolled_back`、`cancelled`、`manual_required` 不得被普通重试改写；如需重开必须创建新 run。

#### 8.11.3 低风险动作白名单与执行流程

低风险动作白名单初稿：

| action_type | 允许条件 | 说明 |
| --- | --- | --- |
| `read_metric_snapshot` | 只读；通过查询网关执行；目标和数据源均可见 | 用于修复前后验证。 |
| `collect_agent_snapshot` | Agent 在线；只读能力；输出脱敏 | 用于进程、磁盘、网络摘要采集。 |
| `restart_service_dry_run` | dry-run only；不改变状态 | 仅验证目标服务存在和权限，不执行重启。 |
| `clear_temp_files_dry_run` | dry-run only；路径必须在模板白名单内 | 只估算可清理文件，不删除。 |
| `notify_owner` | 通知渠道已脱敏且可见；内容不含密钥 | 仅发送人工处理提醒。 |

默认不在白名单内的动作均视为中高风险，例如真实重启服务、删除文件、修改配置、扩容缩容、执行 shell、变更防火墙、修改数据库、修改权限；这些动作必须审批，且需具备 rollback 或 manual_required 策略。

标准执行流程：

| 阶段 | 契约 |
| --- | --- |
| `precheck` | 校验目标可见、Agent 在线、模板启用、参数合法、证据充分、风险等级和审批需求；失败进入 `manual_required`。 |
| `dry-run` | 验证动作将影响的对象、预计变更和回滚可行性；不得改变生产状态。 |
| `approve` | 记录审批人、审批时间、审批摘要和 evidence refs；高风险动作必须审批。 |
| `execute` | 使用幂等键和并发锁执行；步骤结果写 `run_steps` 和 artifacts 摘要。 |
| `verify` | 通过 query gateway、agent snapshot 或事件状态验证效果；失败不得标记成功。 |
| `rollback` | 使用预定义 rollback 模板或人工 runbook；失败进入 `manual_required`。 |
| `manual_required` | 暂停自动化，保留证据、产物、错误摘要和下一步人工建议。 |

#### 8.11.4 异常与安全边界

幂等与并发：

- 同一 `plan_id`、`idempotency_key` 和目标集合的 execute 请求必须返回同一 run 或 409 冲突，不得重复执行。
- 同一目标同一时间只允许一个高风险 remediation run；低风险只读动作可并发但必须限流。
- 重试必须记录 `attempt`，超过最大次数进入 `manual_required`。

超时与离线：

- precheck、dry-run、execute、verify、rollback 均必须配置超时；超时写入步骤失败并触发 rollback 或人工处理。
- Agent 离线、能力缺失或版本不满足时返回 503 或 `manual_required`，不得误报执行成功。
- 队列、Agent 通道或查询网关不可用时不得生成成功 run；必须保留失败摘要和 trace_id。

部分成功与回滚失败：

- 多目标或多步骤执行出现部分成功时，run 状态必须为 `partial_succeeded` 或 `failed`，并逐步骤记录成功、失败、跳过和需回滚目标。
- rollback 成功只能说明已尝试恢复，不得删除原始失败证据。
- rollback 失败必须升级为 `manual_required`，保留失败 artifacts 摘要、受影响目标和人工处置建议。

脱敏与审计：

- 命令、参数、stdout/stderr、环境变量、文件路径、连接信息和 webhook 地址必须经过敏感字段扫描；高敏内容只保留 `<REDACTED>`、digest 或摘要。
- 审计日志至少记录 actor、action、resource、decision、result、trace_id、created_at、evidence_refs；不得记录真实密钥。
- 所有 AI 推荐的 remediation action 在执行前必须转化为受控 template/plan_step；AI 输出不得直接成为可执行命令。

## 9. 契约变更标记

API_CONTRACT_CHANGE：

- 新主线确认为 `/api/v1/monitor/*` 与 `/api/v1/findx-agents/*`。
- P0-1 已实现路由包括 health、targets、findx-agents、register、heartbeat。
- Agent register/heartbeat 认证策略变更为默认拒绝匿名写入：必须配置 `FINDX_AGENT_TOKEN` 或 `findx_agents.shared_token` 并提供正确 `X-Agent-Token: <TOKEN>`；仅本地/测试环境可显式设置 `findx_agents.allow_anonymous: true` 放行匿名接入，生产环境禁止开启。
- P0-2 已实现 alert-rules 与 events 全生命周期路由，并已通过 P0 后端 API QA 门禁。
- P0-3 已新增查询网关主线：`/api/v1/monitor/datasources`、`/api/v1/monitor/query`、`/api/v1/monitor/query-range`、`/api/v1/monitor/metrics`、`/api/v1/monitor/labels`、`/api/v1/monitor/label-values`。
- P1-BE-1 已将事件 fingerprint 与状态机语义固化为服务端契约：服务端生成 fingerprint、不信任外部传入、current upsert 幂等、终态保护、敏感 labels/metadata 脱敏。
- P1-BE-2 已将 evaluator/tryrun 语义固化为服务端契约：tryrun 真实查询 Prometheus，只写 eval log，不写 current/history；Prometheus 上游失败使用 503 或安全失败状态；raw PromQL、URL、token、Cookie、DSN 不进入日志或响应。
- P1-BE-3 已将 scheduler 后端闭环固化为服务端契约：默认禁用，支持 `monitoring.alert_scheduler.enabled`、`interval_seconds`、`timeout_seconds`、`max_concurrency`，RunOnce 可测试，并发锁保护，按 evaluator 状态写 current 或恢复 history。
- CORS 安全基线：默认 localhost-only，只有显式 `server.allowed_origins: ["*"]` 才允许 wildcard。
- Workflow DSL 安全基线：创建/更新在 `ParseDSL` 后必须 `graph.Validate()`，坏 DSL 返回 400 且不得持久化。
- 旧 `/api/v1/n9e/*` 与 `/api/v1/catpaw/*` 保留为兼容入口，不作为新功能主入口。

DATA_CHANGE：

- P0-1 已涉及 `monitor_targets` 与 `findx_agents` 语义。
- P0-2 已涉及 `monitor_alert_rules`、`monitor_alert_rule_versions`、`monitor_alert_eval_logs`、`monitor_alert_events_current`、`monitor_alert_events_history`、`monitor_alert_actions` 等语义。
- P0-3 已涉及 `monitor_datasources`、`monitor_query_audit_logs`、`monitor_metric_metadata` 等语义；真实凭据必须加密或托管，不进入响应、日志、审计正文和 AI prompt。
- P1-BE-1/2/3 继续复用并强化 `monitor_alert_eval_logs`、`monitor_alert_events_current`、`monitor_alert_events_history` 的写入语义；scheduler 失败路径不得产生无 eval log 事件。
- 具体表结构、索引、迁移、回滚和幂等策略以已落库实现和后续迁移文档为准。

后续风险（含非阻断/阻断）：

- 大基数候选截断语义仍需补充明确的 limit、分页、截断提示和审计规则。
- Workflow API 严格 DSL 400 回归仍需持续覆盖，防止坏 DSL 在创建/更新路径被持久化。
- 持久化失败一致性仍需继续压测，特别是 eval log、current、history 分步写入的失败补偿。
- P1 扩展 API 权限与脱敏属于 P1 阻断风险；dashboard、template、notification、silence、subscription、oncall、pipeline、task、permission、audit 未逐接口验证前，不得标记 P1 扩展能力完成。
- Evidence Chain 与 Remediation 全链路仍为契约草案；证据归档、AI schema、反馈归档、审批、执行、验证和回滚未实现与 QA 前，不得视为完成。
- 统一 `.gitattributes` 仍待工程治理处理，降低 Windows/WSL 跨环境换行差异。

## 10. 技术债检查

| 检查项 | 结论 |
| --- | --- |
| 重复逻辑 | 本文档复用现有 model 字段与规划文档中的路由方向，未新增重复代码。 |
| 复杂度 | 文档按 P0-1、P0-2、P1-BE 后端告警闭环、P0-3 查询网关、测试清单和验证证据拆分，避免把契约、测试和治理混在单段描述中。 |
| 边界 | 本次仅更新 `docs/aiops/findx_monitoring_core_api_contract.md`，不修改 `README.md`、`api`、`web`、`.claude`、`.codex` 或密钥配置。 |
| 依赖 | 未新增 Go module、npm 包或外部服务依赖。 |
| 兼容 | 明确 `/n9e/*` 与 `/catpaw/*` 为兼容入口，新平台主线为 `/monitor/*` 与 `/findx-agents/*`。 |
| 测试 | 提供 API、UI、兼容、断连、权限、边界、敏感信息脱敏测试清单；已补充 P1-BE fingerprint、tryrun、scheduler、CORS、Workflow DSL、大基数候选截断、持久化失败一致性、P1 扩展 API 权限与脱敏风险回归项；P0 与 P1-BE-3 后端验证证据已记录。 |
| 回滚 | 文档变更可直接按文件版本回滚；P0-2 规则回滚语义已在契约中约束为生成新版本。 |
| 遗留风险 | P1 后端告警闭环仅 P1-BE-1/2/3 已部分完成；P1 的前端入口、Dashboard、通知、细粒度权限、Agent 深度融合、Evidence Chain、AI 问诊和 Remediation 修复全链路仍需按后续实现与 QA 结果回填；大基数候选截断、Workflow API 严格 DSL 400 回归、持久化失败一致性、P1 扩展 API 权限与脱敏和统一 `.gitattributes` 仍为后续项。 |
