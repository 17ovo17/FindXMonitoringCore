# FindX Monitoring Core API 文档

更新时间：2026-05-05 16:31（UTC+8）

本文档的主线是 FindX Monitoring Core。新功能、测试矩阵、运维手册和前端入口必须优先引用 `/api/v1/monitor/*` 与 `/api/v1/findx-agents/*`；旧 Catpaw 与旧告警平台兼容入口只作为兼容入口保留，不再描述为新主线。自动修复 `/api/v1/remediation/*` 目前属于规划待实现接口，未在代码路由中注册前不得写成已实现或 QA PASS。

## FindX 主入口速览

| 域 | 路径 | 状态 | 说明 |
|----|------|------|------|
| Monitoring Core 健康 | `GET /api/v1/monitor/health` | 已实现并 QA 回归 | 查询监控核心、默认 datasource 与 Query Gateway 健康 |
| Datasource | `GET /api/v1/monitor/datasources` | 已实现并 QA 回归 | 当前以 Prometheus 兼容数据源为主 |
| Query Gateway | `POST /api/v1/monitor/query`、`POST /api/v1/monitor/query-range` | 已实现并 QA 回归 | 即时/区间 PromQL 查询；上游失败应返回 503 |
| 指标元数据 | `GET /api/v1/monitor/metrics`、`GET /api/v1/monitor/labels`、`GET /api/v1/monitor/label-values` | 已实现并 QA 回归 | 大基数候选必须有截断语义或等价风险记录 |
| Target | `/api/v1/monitor/targets` | 已实现并 QA 回归 | 被监控对象登记、读取、更新、删除 |
| Alert Rule | `/api/v1/monitor/alert-rules` | 已实现并 QA 回归 | 规则 CRUD、启停、克隆、试运行、回滚 |
| Event | `/api/v1/monitor/events/current`、`/api/v1/monitor/events/history`、`/api/v1/monitor/events/:id/*` | 已实现并 QA 回归 | 当前/历史事件与 ack/assign/resolve/archive |
| 业务空间 | `/api/v1/workspaces` | 已实现并 QA 回归 | 复用现有拓扑业务存储作为 P0-T1 兼容层 |
| 资源组与主机资产 | `/api/v1/resource-groups`、`/api/v1/host-assets` | 已实现并 QA 回归 | P0-T2 资源分组、主机归属、主机标签和空间绑定 |
| 监控仪表盘 | `/api/v1/monitor/dashboards` | P0-T3 后端代码已落地，待主代理验收 | 主入口是监控仪表盘；不等于运维总览或资产大盘，当前不得写成 QA PASS |
| 内置监控仪表盘模板 | `/api/v1/monitor/dashboard-templates` | P0-T4 代码计划/待主代理验收 | 模板只作为监控仪表盘页面内能力，不新增侧边栏入口 |
| FindX Agents | `/api/v1/findx-agents`、`/api/v1/findx-agents/register`、`/api/v1/findx-agents/heartbeat` | 基础契约已落地 | Agent 注册、心跳、列表；深度巡检/安装发行仍属 P2 |
| Remediation | `/api/v1/remediation/*` | 规划待实现 | plan/approve/execute/verify/rollback/audit 未注册前不得调用 |

## 认证、脱敏与错误模型

| 类型 | 约定 |
|------|------|
| 读接口 | 使用登录态 `Authorization: Bearer <TOKEN>`，未认证返回 401/403 |
| 写接口 | 需要管理员权限，使用管理员登录态或 `X-Admin-Token: <TOKEN>` |
| Agent 上报 | 使用 Agent token 或服务端约定认证，占位统一写 `<TOKEN>` |
| 敏感值 | 文档和示例中的认证或密钥类占位只能使用 `<TOKEN>`；不得写真实密钥、Cookie、DSN、SSH 私钥或完整连接串 |
| 错误码 | 400 参数/PromQL 错误；401/403 认证权限错误；503 外部依赖不可用；500 内部错误 |

## `/api/v1/monitor/*` 主入口

| 方法 | 路径 | 摘要 | 权限 |
|------|------|------|------|
| GET | `/api/v1/monitor/health` | 查询监控核心健康状态 | 登录态 |
| GET | `/api/v1/monitor/datasources` | 列出监控数据源 | 登录态 |
| POST | `/api/v1/monitor/query` | 执行即时 PromQL 查询 | 登录态 |
| POST | `/api/v1/monitor/query-range` | 执行区间 PromQL 查询 | 登录态 |
| GET | `/api/v1/monitor/metrics` | 查询指标名列表 | 登录态 |
| GET | `/api/v1/monitor/labels` | 查询 label 名称列表 | 登录态 |
| GET | `/api/v1/monitor/label-values` | 查询指定 label 的候选值 | 登录态 |
| GET | `/api/v1/monitor/targets` | 获取 Target 列表 | 登录态 |
| POST | `/api/v1/monitor/targets` | 新增或保存 Target | 管理员 |
| GET | `/api/v1/monitor/targets/:id` | 获取 Target 详情 | 登录态 |
| PUT | `/api/v1/monitor/targets/:id` | 更新 Target | 管理员 |
| DELETE | `/api/v1/monitor/targets/:id` | 删除 Target | 管理员 |
| GET | `/api/v1/monitor/alert-rules` | 获取告警规则列表 | 登录态 |
| POST | `/api/v1/monitor/alert-rules` | 创建告警规则 | 管理员 |
| GET | `/api/v1/monitor/alert-rules/:id` | 获取规则详情 | 登录态 |
| PUT | `/api/v1/monitor/alert-rules/:id` | 更新规则 | 管理员 |
| DELETE | `/api/v1/monitor/alert-rules/:id` | 删除规则 | 管理员 |
| POST | `/api/v1/monitor/alert-rules/:id/enable` | 启用规则 | 管理员 |
| POST | `/api/v1/monitor/alert-rules/:id/disable` | 禁用规则 | 管理员 |
| POST | `/api/v1/monitor/alert-rules/:id/clone` | 克隆规则 | 管理员 |
| POST | `/api/v1/monitor/alert-rules/:id/tryrun` | 试运行规则，不创建正式事件 | 管理员 |
| POST | `/api/v1/monitor/alert-rules/:id/rollback` | 回滚规则版本 | 管理员 |
| GET | `/api/v1/monitor/events/current` | 查询当前活跃事件 | 登录态 |
| GET | `/api/v1/monitor/events/history` | 查询历史事件 | 登录态 |
| GET | `/api/v1/monitor/events/:id` | 查询事件详情 | 登录态 |
| POST | `/api/v1/monitor/events/:id/ack` | 确认事件 | 管理员 |
| POST | `/api/v1/monitor/events/:id/assign` | 分派事件 | 管理员 |
| POST | `/api/v1/monitor/events/:id/resolve` | 解决事件 | 管理员 |
| POST | `/api/v1/monitor/events/:id/archive` | 归档事件 | 管理员 |
| GET | `/api/v1/monitor/dashboard-templates` | 获取内置监控仪表盘模板列表 | `monitor.dashboard:read` |
| GET | `/api/v1/monitor/dashboard-templates/:id` | 获取内置监控仪表盘模板详情 | `monitor.dashboard:read` |
| POST | `/api/v1/monitor/dashboard-templates/:id/import` | 导入模板为普通监控仪表盘 | `monitor.dashboard:create` |

Query Gateway 不得返回真实认证信息、Cookie 或平台内部连接串。非法 PromQL、非法时间范围和上游不可达必须向调用方返回可理解错误，Workflow/AI 调用方不得吞掉 400/503 后生成伪证据。

## P0-T3 监控仪表盘（后端代码已落地，待主代理验收）

截至 2026-05-05 16:20（UTC+8）只读核对，后端真实路由已注册 `/api/v1/monitor/dashboards` 路由组，并已具备列表、创建、详情、更新、删除、克隆和分享动作。本文档只按真实代码同步契约，状态写为“待主代理验收”，不得写成最终 PASS 或 QA PASS。

该切片为 `API_CONTRACT_CHANGE`：新增监控仪表盘 API 组。该切片同时为 `DATA_CHANGE`：后端 `api/internal/store/init.go` 已新增 `monitor_dashboards` 表，用于存储仪表盘主记录、变量 JSON、面板 JSON、版本、状态、分享开关和分享 token 哈希。

命名边界：用户侧主入口统一为 **监控仪表盘**。监控仪表盘不等于运维总览，也不等于资产大盘；资产汇总应归入资产总览，旧 `/api/v1/dashboard/summary` 只表示旧汇总接口，不代表 P0-T3 监控仪表盘 CRUD 能力。

### 路由与权限

| 方法 | 路径 | 摘要 | 权限 | 状态 |
|------|------|------|------|------|
| GET | `/api/v1/monitor/dashboards` | 获取监控仪表盘列表，按更新时间倒序，MySQL 模式最多返回 1000 条 | `monitor.dashboard:read`；普通 user 和 admin 均可读 | 代码已落地，待主代理验收 |
| GET | `/api/v1/monitor/dashboards/:id` | 获取监控仪表盘详情 | `monitor.dashboard:read`；普通 user 和 admin 均可读 | 代码已落地，待主代理验收 |
| POST | `/api/v1/monitor/dashboards` | 创建监控仪表盘 | `monitor.dashboard:create`；admin 可执行 | 代码已落地，待主代理验收 |
| PUT | `/api/v1/monitor/dashboards/:id` | 更新监控仪表盘；目标不存在返回 404，不执行 upsert | `monitor.dashboard:update`；admin 可执行 | 代码已落地，待主代理验收 |
| DELETE | `/api/v1/monitor/dashboards/:id` | 删除监控仪表盘 | `monitor.dashboard:delete`；admin 可执行 | 代码已落地，待主代理验收 |
| POST | `/api/v1/monitor/dashboards/:id/clone` | 克隆监控仪表盘，克隆后标题追加 `副本`，不复制分享状态 | `monitor.dashboard:clone`；admin 可执行 | 代码已落地，待主代理验收 |
| POST | `/api/v1/monitor/dashboards/:id/share` | 开启分享状态并写入分享 token 哈希；响应不返回明文 token、hash 或 URL | `monitor.dashboard:share`；admin 可执行 | 代码已落地，待主代理验收 |

### 请求字段

POST 与 PUT 请求体使用 JSON。客户端可写字段仅包括 `title`、`description`、`workspace_id`、`resource_group_id`、`tags`、`variables`、`panels`、`status`；客户端传入 `shared`、`share_token_set`、`share_summary`、`created_by`、`updated_by` 会被后端忽略。PUT 时路径 `:id` 会覆盖请求体内的 `id`，目标不存在返回 404，不执行 upsert。

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `title` | string | 是 | 标题，去除首尾空白后不能为空，最长 120 个字符 |
| `description` | string | 否 | 描述，后端会去除首尾空白 |
| `workspace_id` | string | 否 | 业务空间 ID，后端会去除首尾空白 |
| `resource_group_id` | string | 否 | 资源组 ID，后端会去除首尾空白 |
| `tags` | string[] | 否 | 标签数组，后端会去空、去重并标准化 |
| `variables` | object | 否 | 仪表盘变量 JSON；为空时允许，非空必须是 JSON 对象 |
| `panels` | object[] | 否 | 面板 JSON 数组；为空时允许，非空必须是 JSON 数组 |
| `status` | string | 否 | 允许 `active`、`draft`、`archived`；为空时默认为 `active` |

### 响应字段

列表接口返回监控仪表盘数组；创建、详情、更新、克隆返回单个监控仪表盘对象；删除返回 `{"ok": true}`；分享返回分享动作结果。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 监控仪表盘 ID |
| `title` | string | 标题 |
| `description` | string | 描述 |
| `workspace_id` | string | 业务空间 ID |
| `resource_group_id` | string | 资源组 ID |
| `tags` | string[] | 标签数组 |
| `variables` | object | 仪表盘变量 JSON；响应会递归脱敏敏感键和值 |
| `panels` | object[] | 面板 JSON 数组；响应会递归脱敏敏感键和值 |
| `version` | number | 版本号；创建为 1，更新递增 |
| `status` | string | `active`、`draft` 或 `archived` |
| `shared` | boolean | 是否已开启分享 |
| `share_token_set` | boolean | 是否已设置分享 token 哈希；不返回明文 token |
| `share_summary` | string | 分享摘要，仅分享开启后返回 |
| `created_by` | string | 创建人标识 |
| `updated_by` | string | 更新人标识 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

分享动作响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 监控仪表盘 ID |
| `share_enabled` | boolean | 是否已开启分享 |
| `share_summary` | string | 分享摘要，不包含明文 token、hash 或 URL |

### 错误模型

| 状态码 | 场景 | 响应示例 |
|------|------|------|
| 400 | JSON 无法解析、标题为空、标题超长、状态非法、`variables` 不是对象、`panels` 不是数组 | `{"error":"invalid dashboard","checks":["title is required"]}` |
| 401 | 未登录或登录态无效 | `{"error":"unauthorized"}` |
| 403 | 普通用户执行创建、更新、删除、克隆或分享 | `{"error":"forbidden"}` |
| 404 | 指定监控仪表盘不存在 | `{"error":"dashboard not found"}` |
| 500 | 列表、详情、保存、删除、克隆或分享过程中的内部错误 | `{"error":"dashboard save failed"}` |

### 数据变更

`monitor_dashboards` 表字段：`id`、`title`、`description`、`workspace_id`、`resource_group_id`、`tags`、`variables`、`panels`、`version`、`status`、`shared`、`share_token_hash`、`created_by`、`updated_by`、`created_at`、`updated_at`。索引包括 `idx_md_workspace`、`idx_md_resource_group`、`idx_md_status`、`idx_md_updated`。当前代码通过 `CREATE TABLE IF NOT EXISTS` 创建表；回滚需由主代理结合数据库状态确认是否只移除路由/代码，或额外处理表数据备份与表删除。

### 安全与兼容边界

- 响应会递归脱敏 `variables` 和 `panels` 中包含 token、cookie、dsn、password、secret、api_key、apikey、authorization 等语义的键；包含敏感查询参数或疑似携带凭据的 URL 字符串会返回 `REDACTED_URL`。
- 分享能力只能作为监控仪表盘详情或动作能力承载，不作为独立导航主入口；响应不返回明文 token。
- 监控仪表盘不承载运维总览或资产大盘职责；资产数量、在线率、告警汇总等应由资产总览或对应监控事件接口承载。
- 本节只说明代码契约已落地并待主代理验收，未声明构建通过、UI 回归通过或 QA PASS。

## P0-T4 内置监控仪表盘模板（代码计划/待主代理验收）

P0-T4 是 `API_CONTRACT_CHANGE`：新增内置监控仪表盘模板 API。P0-T4 不是 `DATA_CHANGE`：模板属于内置静态能力，不新增数据库表，不修改既有 schema。本文档只同步契约计划和验收边界，不声明构建通过、UI 回归通过或 QA PASS。

模板只作为监控仪表盘页面内能力出现，用于在创建监控仪表盘时选择、预览并导入常用模板；不新增独立侧边栏入口。导入结果是一条普通监控仪表盘，后续查看、更新、删除、克隆和分享继续复用 P0-T3 `/api/v1/monitor/dashboards` 能力。

### 路由与权限

| 方法 | 路径 | 摘要 | 权限 | 状态 |
|------|------|------|------|------|
| GET | `/api/v1/monitor/dashboard-templates` | 获取内置监控仪表盘模板列表 | `monitor.dashboard:read` | 代码计划/待主代理验收 |
| GET | `/api/v1/monitor/dashboard-templates/:id` | 获取内置监控仪表盘模板详情 | `monitor.dashboard:read` | 代码计划/待主代理验收 |
| POST | `/api/v1/monitor/dashboard-templates/:id/import` | 导入指定模板，生成一条普通监控仪表盘 | `monitor.dashboard:create` | 代码计划/待主代理验收 |

### 内置模板

| 模板 ID 建议 | 名称 | 用途 |
|------|------|------|
| `linux-host-basic` | Linux 主机基础 | Linux 主机 CPU、内存、磁盘、网络等基础观测 |
| `windows-host-basic` | Windows 主机基础 | Windows 主机 CPU、内存、磁盘、网络等基础观测 |
| `mysql-basic` | MySQL 基础 | MySQL 连接、吞吐、延迟和错误等基础观测 |
| `redis-basic` | Redis 基础 | Redis 内存、命令、连接、持久化等基础观测 |
| `kubernetes-node-basic` | Kubernetes 节点基础 | Kubernetes 节点资源、Pod 承载和运行状态基础观测 |

### 模板详情响应

模板列表返回模板摘要数组；模板详情返回单个模板对象。模板对象用于页面预览和导入前确认，不代表已创建的监控仪表盘记录。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 内置模板 ID |
| `name` | string | 模板名称 |
| `description` | string | 模板说明 |
| `category` | string | 模板分类，如 `host`、`database`、`kubernetes` |
| `tags` | string[] | 模板标签 |
| `variables` | object | 模板变量定义；不得包含真实认证、连接或凭据信息 |
| `panels` | object[] | 面板定义；不得包含真实认证、连接或凭据信息 |
| `preview` | object | 页面预览所需的非敏感摘要 |

### 导入请求与响应

导入请求体使用 JSON。客户端可选择业务空间、资源组、标签和变量覆盖值；导入后由后端创建一条普通监控仪表盘。

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `workspace_id` | string | 否 | 导入后监控仪表盘归属的业务空间 ID |
| `resource_group_id` | string | 否 | 导入后监控仪表盘归属的资源组 ID |
| `tags` | string[] | 否 | 导入后附加到监控仪表盘的标签 |
| `variables` | object | 否 | 导入变量覆盖值；必须是 JSON 对象 |

导入响应返回普通监控仪表盘对象，字段与 P0-T3 监控仪表盘详情一致，至少包含 `id`、`title`、`workspace_id`、`resource_group_id`、`tags`、`variables`、`panels`、`version`、`status`、`created_at`、`updated_at`。响应不得返回模板内部敏感配置或导入过程中的凭据信息。

### 错误模型

| 状态码 | 场景 | 响应示例 |
|------|------|------|
| 400 | 导入请求 JSON 无法解析、`variables` 不是对象、标签格式非法 | `{"error":"invalid template import request"}` |
| 401 | 未登录或登录态无效 | `{"error":"unauthorized"}` |
| 403 | 无 `monitor.dashboard` 对应动作权限 | `{"error":"forbidden"}` |
| 404 | 指定模板不存在 | `{"error":"dashboard template not found"}` |
| 500 | 模板加载或导入创建失败 | `{"error":"dashboard template import failed"}` |

### 安全与兼容边界

- 模板和导入响应不得包含真实 token、Cookie、DSN、URL、密码或 SSH 私钥；示例中的认证或凭证类字段只能使用 `<TOKEN>` 等占位符。
- 导入生成的是普通监控仪表盘，不新增模板持久化表，不改变 P0-T3 `monitor_dashboards` 数据语义。
- 权限复用 `monitor.dashboard`：模板列表和详情使用 read，导入使用 create。
- 本切片状态为代码计划/待主代理验收；若后续代码已落地，只能改为“代码已落地待主代理验收”，不得直接写 PASS 或 QA PASS。

## `/api/v1/workspaces` 主入口

P0-T1 新增业务空间 API，用于按业务边界组织主机、端点、负责人、标签和运行状态。该切片为 `API_CONTRACT_CHANGE`：新增 API；`DATA_CHANGE`：无，不新增表、不修改 schema，数据复用现有拓扑业务存储，业务空间的 `description`、`owner`、`status`、`tags` 字段映射到 `topology_businesses.attributes`。

| 方法 | 路径 | 摘要 | 权限 |
|------|------|------|------|
| GET | `/api/v1/workspaces` | 获取业务空间列表 | 登录态 |
| POST | `/api/v1/workspaces` | 创建业务空间 | 管理员 |
| GET | `/api/v1/workspaces/:id` | 获取业务空间详情 | 登录态 |
| PUT | `/api/v1/workspaces/:id` | 更新业务空间 | 管理员 |
| DELETE | `/api/v1/workspaces/:id` | 删除业务空间 | 管理员 |

### 请求字段

POST 与 PUT 请求体使用 JSON：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 业务空间名称，去除首尾空白后不能为空，最长 120 个字符 |
| `description` | string | 否 | 业务空间描述，最长 500 个字符，存入 `topology_businesses.attributes.description` |
| `owner` | string | 否 | 负责人或团队，最长 120 个字符，存入 `topology_businesses.attributes.owner` |
| `status` | string | 否 | 状态，允许 `active`、`disabled`、`archived`；为空时默认为 `active`，存入 `topology_businesses.attributes.status` |
| `tags` | string[] | 否 | 标签列表；服务端会去空、去重、排序，并以逗号分隔存入 `topology_businesses.attributes.tags` |
| `hosts` | string[] | 否 | 主机名或 IP 列表；服务端会去空、去重、排序，映射到拓扑业务存储的主机字段 |
| `endpoints` | object[] | 否 | 端点列表，映射到拓扑业务存储的端点字段 |

`endpoints` 元素字段：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `ip` | string | 是 | 端点 IP，必须能被服务端解析为合法 IP；为空元素会被忽略 |
| `port` | number | 否 | 端口号，允许 0-65535 |
| `service_name` | string | 否 | 服务名称，服务端会去除首尾空白 |
| `protocol` | string | 否 | 协议，服务端会去除首尾空白 |

示例：

```bash
curl -X POST '<BASE_URL>/api/v1/workspaces' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "支付核心链路",
    "description": "支付核心链路的业务空间",
    "owner": "SRE",
    "status": "active",
    "tags": ["payment", "core"],
    "hosts": ["pay-app-01"],
    "endpoints": [
      {"ip": "192.0.2.10", "port": 8080, "service_name": "payment-api", "protocol": "HTTP"}
    ]
  }'
```

### 响应字段

GET 列表返回业务空间数组；POST、GET by id、PUT 返回单个业务空间对象；DELETE 返回 `{"ok": true}`。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 业务空间 ID，复用拓扑业务存储 ID |
| `name` | string | 业务空间名称 |
| `description` | string | 业务空间描述，来自 `topology_businesses.attributes.description` |
| `owner` | string | 负责人或团队，来自 `topology_businesses.attributes.owner` |
| `status` | string | 状态，来自 `topology_businesses.attributes.status`；缺失或非法时响应为 `active` |
| `tags` | string[] | 标签列表，由 `topology_businesses.attributes.tags` 拆分并标准化 |
| `hosts` | string[] | 主机名或 IP 列表 |
| `endpoints` | object[] | 端点列表，字段为 `ip`、`port`、`service_name`、`protocol` |
| `resource_count` | number | 资源计数，按主机集合与端点数量计算 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

### 错误码

| 状态码 | 场景 | 响应示例 |
|------|------|------|
| 400 | JSON 无法解析、`name` 为空、字段超长、`status` 非法、端点 IP 或端口非法 | `{"error":"name is required"}` |
| 401 | 未提供登录态或登录态无效 | `{"error":"unauthorized"}` |
| 403 | 非管理员执行 POST、PUT、DELETE | `{"error":"forbidden"}` |
| 404 | 指定业务空间不存在 | `{"error":"workspace not found"}` |
| 500 | 内部错误 | `{"error":"内部错误描述"}` |

## 资源组与主机资产

P0-T2 新增资源组与主机资产 API，用于在 FindX Monitoring Core 中维护资源分组、主机归属、主机标签和空间绑定关系。该切片为 `API_CONTRACT_CHANGE`：新增 API；`DATA_CHANGE`：新增 `resource_groups` 表；`monitor_targets` 表结构不变，主机归属写入 `labels` 中的 `workspace_id`、`resource_group_id`、`tags` 字段。

安全边界：本章节 API 只负责资源元数据和归属关系维护，不包含远程执行、远程安装、采集配置下发、凭据分发或 Agent 安装发行。主机资产响应会过滤敏感 `labels`，不得返回真实认证信息、Cookie、完整连接串、密钥、内部路径或运行时会话标识。

认证权限：

| 操作 | 权限 |
|------|------|
| GET `/api/v1/resource-groups`、GET `/api/v1/resource-groups/:id`、GET `/api/v1/host-assets`、GET `/api/v1/host-assets/:id` | 需要登录态 |
| POST `/api/v1/resource-groups`、PUT `/api/v1/resource-groups/:id`、DELETE `/api/v1/resource-groups/:id` | 需要管理员角色 |
| PUT `/api/v1/host-assets/:id/tags`、PUT `/api/v1/host-assets/:id/resource-group`、PUT `/api/v1/host-assets/:id/workspace` | 需要管理员角色 |
| 未登录或登录态无效 | 返回 401 |
| 普通用户执行写入、删除、绑定或标签更新 | 返回 403 |

### 资源组 API

| 方法 | 路径 | 摘要 | 权限 |
|------|------|------|------|
| GET | `/api/v1/resource-groups` | 获取资源组列表 | 登录态 |
| POST | `/api/v1/resource-groups` | 创建资源组 | 管理员 |
| GET | `/api/v1/resource-groups/:id` | 获取资源组详情 | 登录态 |
| PUT | `/api/v1/resource-groups/:id` | 更新资源组 | 管理员 |
| DELETE | `/api/v1/resource-groups/:id` | 删除资源组 | 管理员 |

资源组创建与更新请求体使用 JSON：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 资源组名称，去除首尾空白后不能为空 |
| `description` | string | 否 | 资源组描述 |
| `workspace_id` | string | 否 | 归属空间 ID；为空表示暂不绑定空间 |
| `parent_id` | string | 否 | 父资源组 ID；为空表示顶层资源组 |
| `tags` | string[] | 否 | 资源组标签；服务端会去空、去重和标准化 |

示例：

```bash
curl -X POST '<BASE_URL>/api/v1/resource-groups' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "支付核心主机",
    "description": "支付核心链路的主机资源集合",
    "workspace_id": "workspace-pay",
    "tags": ["payment", "core"]
  }'
```

资源组响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 资源组 ID |
| `name` | string | 资源组名称 |
| `description` | string | 资源组描述 |
| `workspace_id` | string | 归属空间 ID |
| `parent_id` | string | 父资源组 ID |
| `status` | string | 资源组状态 |
| `tags` | string[] | 资源组标签 |
| `created_at` | string | 创建时间 |
| `updated_at` | string | 更新时间 |

GET 列表返回资源组数组；POST、GET by id、PUT 返回单个资源组对象；DELETE 返回 `{"ok": true}`。

### 主机资产 API

| 方法 | 路径 | 摘要 | 权限 |
|------|------|------|------|
| GET | `/api/v1/host-assets` | 获取主机资产列表 | 登录态 |
| GET | `/api/v1/host-assets/:id` | 获取主机资产详情 | 登录态 |
| PUT | `/api/v1/host-assets/:id/tags` | 更新主机标签 | 管理员 |
| PUT | `/api/v1/host-assets/:id/resource-group` | 绑定或调整主机所属资源组 | 管理员 |
| PUT | `/api/v1/host-assets/:id/workspace` | 绑定或调整主机所属空间 | 管理员 |

主机资产来自现有监控对象存储，`monitor_targets` 表结构不变。服务端通过 `labels.workspace_id`、`labels.resource_group_id`、`labels.tags` 维护空间、资源组和标签归属。面向前端和外部调用方的响应字段会脱敏，敏感 `labels` 不得透出。

GET `/api/v1/host-assets` 支持的查询参数：

| 参数 | 类型 | 说明 |
|------|------|------|
| `workspace_id` | string | 按空间过滤 |
| `resource_group_id` | string | 按资源组过滤 |
| `tag` | string | 按单个标签过滤 |
| `tags` | string | 按标签过滤，多个标签可用逗号分隔 |
| `status` | string | 按主机资产状态过滤 |
| `online` | boolean | 按在线状态过滤 |
| `keyword` | string | 按主机标识、主机名、IP 或标签关键字过滤 |

主机绑定与标签更新请求体：

| 接口 | 请求字段 | 说明 |
|------|------|------|
| PUT `/api/v1/host-assets/:id/tags` | `tags` string[] | 更新主机标签；写入 `labels.tags` |
| PUT `/api/v1/host-assets/:id/resource-group` | `resource_group_id` string | 更新主机资源组归属；写入 `labels.resource_group_id` |
| PUT `/api/v1/host-assets/:id/workspace` | `workspace_id` string | 更新主机空间归属；写入 `labels.workspace_id` |

示例：

```bash
curl -X PUT '<BASE_URL>/api/v1/host-assets/host-001/resource-group' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"resource_group_id":"rg-pay-core"}'
```

主机资产响应字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `host_id` | string | 主机资产 ID，来自监控对象 ID |
| `ident` | string | 主机唯一标识 |
| `hostname` | string | 主机名 |
| `ip_list` | string[] | 主机 IP 列表 |
| `os` | string | 操作系统 |
| `arch` | string | CPU 架构 |
| `workspace_id` | string | 当前空间归属，来自 `labels.workspace_id` |
| `resource_group_id` | string | 当前资源组归属，来自 `labels.resource_group_id` |
| `tags` | string[] | 主机标签，来自 `labels.tags` 的标准化结果 |
| `agent_id` | string | 关联 Agent ID |
| `agent_status` | string | Agent 当前状态 |
| `agent_version` | string | Agent 版本 |
| `last_seen_at` | string | 最近一次观测时间 |
| `status` | string | 主机资产状态 |
| `source` | string | 主机资产来源 |
| `labels` | object | 脱敏后的非敏感标签 |
| `updated_at` | string | 更新时间 |

GET 列表当前返回主机资产数组；如后续实现分页，必须以代码路由和响应结构为准，不在当前契约中承诺分页对象。GET by id 和 PUT 绑定类接口返回单个主机资产对象。

### 错误码

| 状态码 | 场景 | 响应示例 |
|------|------|------|
| 400 | JSON 无法解析、必填字段为空、字段类型不正确、标签格式非法、过滤参数非法 | `{"error":"invalid request"}` |
| 401 | 未提供登录态或登录态无效 | `{"error":"unauthorized"}` |
| 403 | 普通用户执行资源组写入、删除、主机绑定或标签更新 | `{"error":"forbidden"}` |
| 404 | 资源组或主机资产不存在 | `{"error":"resource not found"}` |
| 409 | 删除仍有关联主机资产的资源组时可能返回 409；最终以当前后端实现为准 | `{"error":"resource group has hosts"}` |
| 500 | 内部错误 | `{"error":"内部错误描述"}` |

### 验证摘要

本切片已覆盖以下验证口径，详细证据以主代理门禁记录和 QA 报告为准：

| 验证项 | 摘要 |
|------|------|
| Windows go test | 在 Windows 工作目录执行 Go 单元测试，覆盖资源组、主机资产、权限和标签归属逻辑 |
| WSL go test/build | 同步到 WSL `/opt/ai-workbench` 后执行 Go 测试与后端构建 |
| WSL build-web | 在 WSL `/opt/ai-workbench/web` 执行前端构建，验证页面入口和 API 封装可编译 |
| 运行态 401 | 未登录访问新增 GET 与写接口返回 401，未泄露内部错误、敏感 labels 或连接信息 |
| Playwright UI 回归 | 覆盖资源组列表、主机资产列表、绑定资源组、绑定空间和标签更新的真实浏览器回归 |

## `/api/v1/findx-agents/*` 主入口

| 方法 | 路径 | 摘要 | 权限 |
|------|------|------|------|
| GET | `/api/v1/findx-agents` | 获取 FindX Agent 列表和心跳状态 | 登录态 |
| POST | `/api/v1/findx-agents/register` | Agent 注册 | Agent token 或服务端约定 |
| POST | `/api/v1/findx-agents/heartbeat` | Agent 心跳 | Agent token 或服务端约定 |

## 兼容入口

| 兼容入口 | 当前用途 | 新主入口 |
|---------|---------|---------|
| `/api/v1/catpaw/heartbeat` | 旧 Catpaw 探针心跳 | `/api/v1/findx-agents/heartbeat` |
| `/api/v1/catpaw/report` | 旧 Catpaw 巡检报告 | FindX Agent 后续独立证据上报契约 |
| `/api/v1/catpaw/agents` | 旧 Catpaw Agent 列表 | `/api/v1/findx-agents` |
| `/api/v1/catpaw/chat-ws` | 旧 Catpaw 会话 | 未来 Remediation/结构化工具链 |
| 旧告警平台 Agent 兼容查询入口 | Agent 兼容查询 | `/api/v1/findx-agents` |
| 旧告警平台事件兼容查询入口 | 告警兼容查询 | `/api/v1/monitor/events/*` |

以下历史 API 清单保留用于回归和迁移核对。

# AI Workbench 历史 API 文档

更新时间：2026-05-03 14:36（UTC+8）

本文档根据 `api/main.go` 中注册的路由重写，覆盖 156 个 GET/POST/PUT/DELETE API 端点。`/api/v1/download` 为静态文件分发路由，未计入 API 端点总数。

## 通用约定

- Base URL：`http://localhost:8080`
- 默认 API 前缀：`/api/v1`
- 指标暴露端点：`/metrics` 不使用 `/api/v1` 前缀。
- 写操作（POST/PUT/DELETE）通常需要 `X-Admin-Token` 请求头或 `Authorization: Bearer <token>`。
- 登录态接口使用 `Authorization: Bearer <token>`。
- 请求与响应默认使用 JSON；失败响应通常为 `{"error": "中文错误描述"}`。
- 分页接口通常使用 `page`、`limit` 查询参数，列表响应通常包含 `total` 字段。
- 本文档仅列出路由路径与用途说明；不展开请求体、查询参数或响应结构。路径参数以 `:id`、`:name`、`:ip`、`:key`、`:action` 等形式直接体现在路径中。

## 状态码

| 码值 | 含义 |
|------|------|
| 200 | 成功 |
| 400 | 参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 内部错误 |
| 503 | 依赖服务不可用 |

---

## AI 模型与对话

### GET /api/v1/models
获取可用 AI 模型列表。

### POST /api/v1/chat
发起普通对话请求。

### GET /api/v1/chat/sessions
获取对话会话列表。

### POST /api/v1/chat/sessions
创建新的对话会话。

### GET /api/v1/chat/sessions/:id
获取指定对话会话详情。

### PUT /api/v1/chat/sessions/:id
重命名指定对话会话。

### DELETE /api/v1/chat/sessions/:id
删除指定对话会话。

### POST /api/v1/agent/llm/chat
发起 Agent LLM 对话请求。

## AIOps 智能运维

### POST /api/v1/aiops/sessions
创建 AIOps 智能运维会话。

### POST /api/v1/aiops/sessions/:id/messages
向指定 AIOps 会话发送消息。

### GET /api/v1/aiops/sessions/:id/messages
获取指定 AIOps 会话的消息列表。

### GET /api/v1/aiops/ws/sessions/:id
建立指定 AIOps 会话的 WebSocket 连接。

### POST /api/v1/aiops/sessions/:id/actions/execute
执行指定 AIOps 会话中的动作。

### POST /api/v1/aiops/inspections
创建 AIOps 巡检任务。

### GET /api/v1/aiops/inspections/:id/progress
获取指定 AIOps 巡检任务进度。

### GET /api/v1/aiops/inspections/:id/report
获取指定 AIOps 巡检报告。

### POST /api/v1/aiops/data/prometheus/query
执行 AIOps Prometheus 即时查询。

### POST /api/v1/aiops/data/prometheus/query_range
执行 AIOps Prometheus 区间查询。

### POST /api/v1/aiops/data/catpaw/query
执行 AIOps CatPaw 数据查询。

### POST /api/v1/aiops/topology/generate
生成 AIOps 拓扑数据。

## 诊断

### POST /api/v1/diagnose
启动诊断任务。

### GET /api/v1/diagnose
获取诊断任务列表。

### GET /api/v1/diagnose/compare
对比诊断任务结果。

### DELETE /api/v1/diagnose
清理诊断任务数据。

### DELETE /api/v1/diagnose/:id
删除指定诊断任务。

## CatPaw 探针

### POST /api/v1/catpaw/heartbeat
接收 CatPaw 探针心跳。

### POST /api/v1/catpaw/report
接收 CatPaw 探针上报数据。

### GET /api/v1/catpaw/agents
获取 CatPaw 探针列表。

### DELETE /api/v1/catpaw/agents/:ip
删除指定 IP 的 CatPaw 探针记录。

### GET /api/v1/catpaw/chat-ws
建立 CatPaw 对话 WebSocket 连接。

## 告警管理

### POST /api/v1/alert/webhook
接收外部告警 Webhook。

### POST /api/v1/alert/catpaw
接收 CatPaw 告警。

### GET /api/v1/alerts
获取告警列表。

### PUT /api/v1/alerts/:id/resolve
将指定告警标记为已解决。

### PUT /api/v1/alerts/:id/:action
对指定告警执行动作。

### DELETE /api/v1/alerts/:id
删除指定告警。

## 远程执行

### POST /api/v1/remote/exec
执行远程命令。

### POST /api/v1/remote/check-port
检查远程端口连通性。

### POST /api/v1/remote/install-catpaw
远程安装 CatPaw 探针。

### POST /api/v1/remote/uninstall-catpaw
远程卸载 CatPaw 探针。

### POST /api/v1/remote/install-cmd
生成远程安装命令。

## 平台与监控

### GET /metrics
暴露平台运行指标。

### GET /api/v1/platform/ip
获取平台 IP 信息。

### GET /api/v1/prometheus/instances
获取 Prometheus 实例列表。

### GET /api/v1/prometheus/hosts
获取 Prometheus 主机列表。

### GET /api/v1/prometheus/metrics
获取 Prometheus 指标列表。

### GET /api/v1/health/datasources
检查数据源健康状态。

### GET /api/v1/health/ai-providers
检查 AI 提供商健康状态。

### GET /api/v1/health/storage
检查存储健康状态。

### GET /api/v1/audit/events
获取审计事件列表。

## 值班管理

### GET /api/v1/oncall/config
获取值班配置。

### POST /api/v1/oncall/config
保存值班配置。

### GET /api/v1/oncall/groups
获取响应团队列表。

### POST /api/v1/oncall/groups
保存响应团队。

### DELETE /api/v1/oncall/groups/:id
删除指定响应团队。

### GET /api/v1/oncall/channels
获取值班通知渠道列表。

### POST /api/v1/oncall/channels
保存值班通知渠道。

### DELETE /api/v1/oncall/channels/:id
删除指定值班通知渠道。

### GET /api/v1/oncall/schedules
获取值班排班列表。

### POST /api/v1/oncall/schedules
保存值班排班。

### DELETE /api/v1/oncall/schedules/:id
删除指定值班排班。

### GET /api/v1/oncall/records
获取值班记录列表。

### POST /api/v1/oncall/test-send
发送值班通知测试消息。

## 拓扑管理

### GET /api/v1/topology
获取拓扑数据。

### POST /api/v1/topology
保存拓扑数据。

### GET /api/v1/topology/resources
获取拓扑资源列表。

### POST /api/v1/topology/discover
执行拓扑发现。

### POST /api/v1/topology/ai/generate
使用 AI 生成拓扑。

### GET /api/v1/topology/businesses
获取拓扑业务列表。

### POST /api/v1/topology/businesses
保存拓扑业务。

### GET /api/v1/topology/businesses/:id
获取指定拓扑业务详情。

### GET /api/v1/topology/businesses/:id/inspect
巡检指定拓扑业务。

### DELETE /api/v1/topology/businesses/:id
删除指定拓扑业务。

## 配置管理

### GET /api/v1/ai-providers
获取 AI 提供商配置。

### POST /api/v1/ai-providers
保存 AI 提供商配置。

### GET /api/v1/data-sources
获取数据源配置。

### POST /api/v1/data-sources
保存数据源配置。

### GET /api/v1/credentials
获取凭据列表。

### POST /api/v1/credentials
保存凭据。

### DELETE /api/v1/credentials/:id
删除指定凭据。

## 认证与用户

### POST /api/v1/auth/login
用户登录。

### POST /api/v1/auth/logout
用户登出。

### GET /api/v1/auth/me
获取当前登录用户信息。

### POST /api/v1/auth/change-password
修改当前登录用户密码。

### GET /api/v1/user-profiles
获取用户画像列表。

### POST /api/v1/user-profiles
保存用户画像。

### DELETE /api/v1/user-profiles/:id
删除指定用户画像。

## 仪表盘与关联

### GET /api/v1/dashboard/summary
获取旧汇总页数据。该接口不是 P0-T3 监控仪表盘主入口，也不等同于资产总览或监控仪表盘 CRUD；新监控仪表盘能力应以后续 `/api/v1/monitor/dashboards` 真实实现为准。

### GET /api/v1/correlate
获取关联分析结果。

### POST /api/v1/workflows/route-preview
预览工作流路由结果。

## 旧告警平台兼容集成

### 旧告警平台 Agent 查询
获取旧告警平台 Agent 列表。

### 旧告警平台告警查询
获取旧告警平台告警列表。

## 知识库-案例

### GET /api/v1/knowledge/cases
获取知识库案例列表。

### GET /api/v1/knowledge/cases/export
导出知识库案例。

### POST /api/v1/knowledge/cases
创建知识库案例。

### POST /api/v1/knowledge/cases/import
导入知识库案例。

### GET /api/v1/knowledge/cases/:id
获取指定知识库案例详情。

### PUT /api/v1/knowledge/cases/:id
更新指定知识库案例。

### DELETE /api/v1/knowledge/cases/:id
删除指定知识库案例。

## 知识库-文档

### POST /api/v1/knowledge/documents/upload
上传知识库文档。

### GET /api/v1/knowledge/documents
获取知识库文档列表。

### GET /api/v1/knowledge/documents/:id
获取指定知识库文档详情。

### DELETE /api/v1/knowledge/documents/:id
删除指定知识库文档。

### POST /api/v1/knowledge/documents/:id/reindex
重建指定知识库文档索引。

### POST /api/v1/knowledge/search
执行知识库检索。

### GET /api/v1/knowledge/search/stats
获取知识库检索统计。

### POST /api/v1/knowledge/search/badcase
提交知识库检索 badcase。

### POST /api/v1/knowledge/reindex-all
重建全部知识库索引。

## 知识库-Runbook

### GET /api/v1/knowledge/runbooks
获取 Runbook 列表。

### GET /api/v1/knowledge/runbooks/:id
获取指定 Runbook 详情。

### POST /api/v1/knowledge/runbooks
创建 Runbook。

### PUT /api/v1/knowledge/runbooks/:id
更新指定 Runbook。

### DELETE /api/v1/knowledge/runbooks/:id
删除指定 Runbook。

### POST /api/v1/knowledge/runbooks/:id/execute
执行指定 Runbook。

### GET /api/v1/knowledge/runbooks/:id/history
获取指定 Runbook 执行历史。

## 诊断工作流与反馈

### POST /api/v1/diagnosis/start
启动诊断工作流。

### POST /api/v1/diagnosis/feedback
提交诊断反馈。

### GET /api/v1/diagnosis/feedback
获取指定诊断反馈列表。

### GET /api/v1/diagnosis/feedback/all
获取全部诊断反馈列表。

### GET /api/v1/diagnosis/feedback/stats
获取诊断反馈统计。

### GET /api/v1/diagnosis/verifications
获取诊断验证列表。

### POST /api/v1/diagnosis/archive
归档诊断结果。

## 工作流管理

### GET /api/v1/workflows
获取工作流列表。

### GET /api/v1/workflows/:id
获取指定工作流详情。

### POST /api/v1/workflows
创建工作流。

### PUT /api/v1/workflows/:id
更新指定工作流。

### DELETE /api/v1/workflows/:id
删除指定工作流。

### POST /api/v1/workflows/:id/run
运行指定工作流。

### POST /api/v1/workflows/:id/stream
以流式方式运行指定工作流。

### GET /api/v1/workflows/:id/runs
获取指定工作流运行记录。

## 定时调度

### GET /api/v1/schedules
获取定时调度列表。

### POST /api/v1/schedules
创建定时调度。

### DELETE /api/v1/schedules/:id
删除指定定时调度。

## 通知渠道

### GET /api/v1/notifications/channels
获取通知渠道列表。

### POST /api/v1/notifications/channels
保存通知渠道。

### DELETE /api/v1/notifications/channels/:id
删除指定通知渠道。

## 数据源指标语义

### POST /api/v1/metrics/scan
扫描指标目录。

### POST /api/v1/metrics/auto-adapt
自动适配数据源指标语义。

### GET /api/v1/metrics/mappings
获取数据源指标语义列表。

### PUT /api/v1/metrics/mappings/:id
更新指定数据源指标语义。

### POST /api/v1/metrics/mappings/confirm
确认数据源指标语义。

## Prompt 模板

### GET /api/v1/prompts
获取 Prompt 模板列表。

### GET /api/v1/prompts/:name
获取指定 Prompt 模板详情。

### POST /api/v1/prompts
创建 Prompt 模板。

### PUT /api/v1/prompts/:name
更新指定 Prompt 模板。

### DELETE /api/v1/prompts/:name
删除指定 Prompt 模板。

## AI 设置

### GET /api/v1/settings/ai
获取 AI 设置列表。

### PUT /api/v1/settings/ai
更新 AI 设置列表。

### GET /api/v1/settings/ai/:key
获取指定 AI 设置项。

### PUT /api/v1/settings/ai/:key
更新指定 AI 设置项。

### DELETE /api/v1/settings/ai/:key
删除指定 AI 设置项。

## Embedding 与 Reranker 设置

### GET /api/v1/settings/embedding
获取 Embedding 设置。

### PUT /api/v1/settings/embedding
更新 Embedding 设置。

### POST /api/v1/settings/embedding/test
测试 Embedding 连接。

### GET /api/v1/settings/reranker
获取 Reranker 设置。

### PUT /api/v1/settings/reranker
更新 Reranker 设置。
# FindX Monitoring Core API 摘要

更新时间：2026-05-04 07:06（UTC+8）

本摘要根据 `api/main.go` 当前注册路由补强。FindX Monitoring Core 的新主入口为 `/api/v1/monitor/*` 与 `/api/v1/findx-agents/*`；旧 Catpaw 与旧告警平台兼容入口继续保留，不作为新功能主入口。未来 `/api/v1/remediation/*` 尚未注册，本文仅列为规划/待实现接口，后续必须以代码路由、权限、审计和测试矩阵验收为准。

## 通用认证与安全约定

| 类型 | 约定 |
|------|------|
| 读接口 | 通常使用 `Authorization: Bearer <TOKEN>` 登录态访问 |
| 写接口 | 通常需要管理员权限，可使用 `X-Admin-Token: <TOKEN>` 或登录态管理员权限 |
| Agent 上报 | Agent 注册和心跳使用服务端约定的 Agent token，占位写作 `<TOKEN>` |
| 敏感字段 | 文档、日志、测试报告中的认证、密钥、凭证类示例统一只使用 `<TOKEN>`，不得写真实值 |
| 错误模型 | 400 表示请求参数或 PromQL 错误；401/403 表示认证或权限失败；503 表示外部依赖不可用；500 表示内部错误 |

## `/api/v1/monitor/*` 主入口

### 健康、数据源与查询网关

| 方法 | 路径 | 摘要 | 权限 |
|------|------|------|------|
| GET | `/api/v1/monitor/health` | 查询 FindX Monitoring Core 健康状态，覆盖默认 datasource 和查询链路 | 登录态 |
| GET | `/api/v1/monitor/datasources` | 列出监控数据源，当前以 Prometheus 兼容数据源为主 | 登录态 |
| POST | `/api/v1/monitor/query` | 执行即时 PromQL 查询，作为 Query Gateway 主入口 | 登录态 |
| POST | `/api/v1/monitor/query-range` | 执行区间 PromQL 查询，支持时间窗口和 step | 登录态 |
| GET | `/api/v1/monitor/metrics` | 查询指标名列表，用于指标选择和规则配置 | 登录态 |
| GET | `/api/v1/monitor/labels` | 查询 label 名称列表 | 登录态 |
| GET | `/api/v1/monitor/label-values` | 查询指定 label 的候选值 | 登录态 |

注意事项：
- Query Gateway 不应返回真实认证信息；Prometheus URL、token 和完整连接串必须脱敏。
- 大基数 label-values 或 metrics 查询可能出现候选截断，调用方必须把截断状态作为“不完整证据”处理。
- Workflow 或 AI 链路调用 Query Gateway 时，若返回 400，应展示可理解错误，不得吞掉错误后生成伪证据。

### Target 管理

| 方法 | 路径 | 摘要 | 权限 |
|------|------|------|------|
| GET | `/api/v1/monitor/targets` | 获取 Target 列表 | 登录态 |
| POST | `/api/v1/monitor/targets` | 新增或保存 Target | 管理员 |
| GET | `/api/v1/monitor/targets/:id` | 获取指定 Target 详情 | 登录态 |
| PUT | `/api/v1/monitor/targets/:id` | 更新指定 Target | 管理员 |
| DELETE | `/api/v1/monitor/targets/:id` | 删除指定 Target | 管理员 |

Target 用于统一被监控对象标识、数据源归属和标签映射。删除或更新 Target 前，应确认没有活跃事件、自动修复计划或未归档处置记录引用该对象。

### Alert Rule 管理

| 方法 | 路径 | 摘要 | 权限 |
|------|------|------|------|
| GET | `/api/v1/monitor/alert-rules` | 获取告警规则列表 | 登录态 |
| POST | `/api/v1/monitor/alert-rules` | 创建告警规则 | 管理员 |
| GET | `/api/v1/monitor/alert-rules/:id` | 获取告警规则详情 | 登录态 |
| PUT | `/api/v1/monitor/alert-rules/:id` | 更新告警规则 | 管理员 |
| DELETE | `/api/v1/monitor/alert-rules/:id` | 删除告警规则 | 管理员 |
| POST | `/api/v1/monitor/alert-rules/:id/enable` | 启用规则 | 管理员 |
| POST | `/api/v1/monitor/alert-rules/:id/disable` | 禁用规则 | 管理员 |
| POST | `/api/v1/monitor/alert-rules/:id/clone` | 克隆规则 | 管理员 |
| POST | `/api/v1/monitor/alert-rules/:id/tryrun` | 试运行规则，不应创建正式事件 | 管理员 |
| POST | `/api/v1/monitor/alert-rules/:id/rollback` | 回滚到历史版本 | 管理员 |

规则变更属于高风险写操作，必须覆盖 PromQL 400、空结果、大基数候选、持久化失败一致性和审计脱敏验证。

### Event 处置

| 方法 | 路径 | 摘要 | 权限 |
|------|------|------|------|
| GET | `/api/v1/monitor/events/current` | 查询当前活跃事件 | 登录态 |
| GET | `/api/v1/monitor/events/history` | 查询历史事件 | 登录态 |
| GET | `/api/v1/monitor/events/:id` | 查询事件详情 | 登录态 |
| POST | `/api/v1/monitor/events/:id/ack` | 确认事件 | 管理员 |
| POST | `/api/v1/monitor/events/:id/assign` | 分派事件 | 管理员 |
| POST | `/api/v1/monitor/events/:id/resolve` | 解决事件 | 管理员 |
| POST | `/api/v1/monitor/events/:id/archive` | 归档事件 | 管理员 |

事件详情应能串联 rule、target、datasource、query hash、labels/annotations、状态动作记录。任何事件动作都应进入审计或动作日志，敏感 label 值必须脱敏。

## `/api/v1/findx-agents/*` 主入口

| 方法 | 路径 | 摘要 | 权限 |
|------|------|------|------|
| GET | `/api/v1/findx-agents` | 获取 FindX Agent 列表和心跳状态 | 登录态 |
| POST | `/api/v1/findx-agents/register` | Agent 注册入口 | Agent token 或服务端约定 |
| POST | `/api/v1/findx-agents/heartbeat` | Agent 心跳入口 | Agent token 或服务端约定 |

Agent 注册与心跳接口面向机器端上报，不应返回真实 token、完整配置文件或主机敏感信息。升级、回滚、安装脚本仍需由运维流程和测试矩阵补齐，不应复用旧 Catpaw 路径作为新 FindX Agent 主入口。

## `/api/v1/remediation/*` 规划接口

以下接口为未来自动修复能力的规划/待实现摘要；当前 API 路由未注册，调用方不得按已实现能力依赖。

| 方法 | 规划路径 | 规划摘要 | 状态 |
|------|----------|---------|------|
| POST | `/api/v1/remediation/plans` | 根据事件、规则和证据链生成修复计划 | 待实现 |
| GET | `/api/v1/remediation/plans/:id` | 查看修复计划、审批状态、执行证据和回滚状态 | 待实现 |
| POST | `/api/v1/remediation/plans/:id/approve` | 审批修复计划 | 待实现 |
| POST | `/api/v1/remediation/plans/:id/reject` | 驳回修复计划 | 待实现 |
| POST | `/api/v1/remediation/plans/:id/execute` | 执行已审批计划 | 待实现 |
| POST | `/api/v1/remediation/plans/:id/rollback` | 执行回滚步骤 | 待实现 |
| GET | `/api/v1/remediation/actions` | 查询可用动作、风险分级和命令模板 | 待实现 |

上线前契约要求：
- 必须区分 plan、approval、execution、rollback 四类状态。
- 必须支持幂等、重复提交保护、超时、失败中止和审计。
- L4 禁止命令必须无条件拦截；L2/L3 必须有确认或审批。
- 所有响应和日志必须脱敏，不得包含真实 token、Cookie、完整连接串或 SSH 私钥。

## 兼容入口说明

| 兼容入口 | 当前用途 | 新主入口建议 |
|---------|---------|-------------|
| `/api/v1/catpaw/heartbeat` | 旧 Catpaw 探针心跳 | 新功能优先使用 `/api/v1/findx-agents/heartbeat` |
| `/api/v1/catpaw/report` | 旧 Catpaw 巡检报告上报 | FindX Agent 后续应使用独立上报契约，当前不把 Catpaw 当新主入口 |
| `/api/v1/catpaw/agents` | 旧 Catpaw Agent 列表 | 新列表使用 `/api/v1/findx-agents` |
| `/api/v1/catpaw/chat-ws` | 旧 Catpaw 交互式会话 | 自动修复待 `/api/v1/remediation/*` 落地后统一治理 |
| 旧告警平台 Agent 兼容查询入口 | Agent 兼容查询 | FindX Agent 使用 `/api/v1/findx-agents` |
| 旧告警平台事件兼容查询入口 | 告警兼容查询 | FindX Event 使用 `/api/v1/monitor/events/*` |

兼容入口只用于历史数据、迁移期适配或旧页面保留。新增页面、测试矩阵和运维手册应优先引用 FindX Monitoring Core 主入口。

---
