# FindX Monitoring Core API 文档

更新时间：2026-05-04 07:30（UTC+8）

本文档的主线是 FindX Monitoring Core。新功能、测试矩阵、运维手册和前端入口必须优先引用 `/api/v1/monitor/*` 与 `/api/v1/findx-agents/*`；旧 `/api/v1/catpaw/*`、`/api/v1/n9e/*` 只作为兼容入口保留，不再描述为新主线。自动修复 `/api/v1/remediation/*` 目前属于规划待实现接口，未在代码路由中注册前不得写成已实现或 QA PASS。

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
| FindX Agents | `/api/v1/findx-agents`、`/api/v1/findx-agents/register`、`/api/v1/findx-agents/heartbeat` | 基础契约已落地 | Agent 注册、心跳、列表；深度巡检/安装发行仍属 P2 |
| Remediation | `/api/v1/remediation/*` | 规划待实现 | plan/approve/execute/verify/rollback/audit 未注册前不得调用 |

## 认证、脱敏与错误模型

| 类型 | 约定 |
|------|------|
| 读接口 | 使用登录态 `Authorization: Bearer <TOKEN>`，未认证返回 401/403 |
| 写接口 | 需要管理员权限，使用管理员登录态或 `X-Admin-Token: <TOKEN>` |
| Agent 上报 | 使用 Agent token 或服务端约定认证，占位统一写 `<TOKEN>` |
| 敏感值 | 文档和示例只能使用 `<TOKEN>`、`<API_KEY>`、`<DB_DSN>`、`<BASE_URL>`、`<LOGIN_USER>`、`<SSH_KEY>` |
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

Query Gateway 不得返回真实认证信息、完整 `<DB_DSN>`、Cookie 或平台内部连接串。非法 PromQL、非法时间范围和上游不可达必须向调用方返回可理解错误，Workflow/AI 调用方不得吞掉 400/503 后生成伪证据。

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
| `/api/v1/n9e/agents` | N9E Agent 兼容查询 | `/api/v1/findx-agents` |
| `/api/v1/n9e/alerts` | N9E 告警兼容查询 | `/api/v1/monitor/events/*` |

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
获取值班组列表。

### POST /api/v1/oncall/groups
保存值班组。

### DELETE /api/v1/oncall/groups/:id
删除指定值班组。

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
获取仪表盘汇总数据。

### GET /api/v1/correlate
获取关联分析结果。

### POST /api/v1/workflows/route-preview
预览工作流路由结果。

## N9e 集成

### GET /api/v1/n9e/agents
获取 N9e Agent 列表。

### GET /api/v1/n9e/alerts
获取 N9e 告警列表。

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

## 指标映射

### POST /api/v1/metrics/scan
扫描指标。

### POST /api/v1/metrics/auto-adapt
自动适配指标。

### GET /api/v1/metrics/mappings
获取指标映射列表。

### PUT /api/v1/metrics/mappings/:id
更新指定指标映射。

### POST /api/v1/metrics/mappings/confirm
确认指标映射。

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

本摘要根据 `api/main.go` 当前注册路由补强。FindX Monitoring Core 的新主入口为 `/api/v1/monitor/*` 与 `/api/v1/findx-agents/*`；旧 `/api/v1/catpaw/*`、`/api/v1/n9e/*` 继续作为兼容入口保留，不作为新功能主入口。未来 `/api/v1/remediation/*` 尚未注册，本文仅列为规划/待实现接口，后续必须以代码路由、权限、审计和测试矩阵验收为准。

## 通用认证与安全约定

| 类型 | 约定 |
|------|------|
| 读接口 | 通常使用 `Authorization: Bearer <TOKEN>` 登录态访问 |
| 写接口 | 通常需要管理员权限，可使用 `X-Admin-Token: <TOKEN>` 或登录态管理员权限 |
| Agent 上报 | Agent 注册和心跳使用服务端约定的 Agent token，占位写作 `<TOKEN>` |
| 敏感字段 | 文档、日志、测试报告中统一使用 `<TOKEN>`、`<DB_DSN>`、`<SECRET>`、`<AGENT_ID>`，不得写真实值 |
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
- Query Gateway 不应返回真实认证信息；Prometheus URL、token、完整 `<DB_DSN>` 必须脱敏。
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
- 所有响应和日志必须脱敏，不得包含真实 `<TOKEN>`、Cookie、完整 `<DB_DSN>` 或 SSH 私钥。

## 兼容入口说明

| 兼容入口 | 当前用途 | 新主入口建议 |
|---------|---------|-------------|
| `/api/v1/catpaw/heartbeat` | 旧 Catpaw 探针心跳 | 新功能优先使用 `/api/v1/findx-agents/heartbeat` |
| `/api/v1/catpaw/report` | 旧 Catpaw 巡检报告上报 | FindX Agent 后续应使用独立上报契约，当前不把 Catpaw 当新主入口 |
| `/api/v1/catpaw/agents` | 旧 Catpaw Agent 列表 | 新列表使用 `/api/v1/findx-agents` |
| `/api/v1/catpaw/chat-ws` | 旧 Catpaw 交互式会话 | 自动修复待 `/api/v1/remediation/*` 落地后统一治理 |
| `/api/v1/n9e/agents` | N9e Agent 兼容查询 | FindX Agent 使用 `/api/v1/findx-agents` |
| `/api/v1/n9e/alerts` | N9e 告警兼容查询 | FindX Event 使用 `/api/v1/monitor/events/*` |

兼容入口只用于历史数据、迁移期适配或旧页面保留。新增页面、测试矩阵和运维手册应优先引用 FindX Monitoring Core 主入口。

---
