# AI Workbench API 文档

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
