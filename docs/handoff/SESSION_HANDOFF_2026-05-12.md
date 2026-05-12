# FindX 平台开发会话交接 — 2026-05-12

## 项目背景

FindX 智能运维平台，参考多个成熟开源项目构建：
- **夜莺 Nightingale (fe-main)** — 仪表盘、指标查询、告警规则
- **SkyWalking booster UI** — 链路监控（服务/Trace/拓扑）
- **SigNoZ** — 日志中心
- **LWOPS Lerwee** — AI 助手对话
- **内部 CMDB/Agent 源码** — 配置管理、Agent 生命周期

## 环境

| 环境 | 地址 |
|------|------|
| 远端 Ubuntu | SSH: `findx-ubuntu` (10.10.160.202) |
| API 服务 | http://127.0.0.1:8080 (ai-workbench-api.service) |
| Web 服务 | http://10.10.160.202:3000 (nginx) |
| 项目路径（本机） | `D:\ai-workbench` |
| 项目路径（远端） | `/opt/ai-workbench/api`（代码）、`/opt/ai-workbench-runtime`（运行时） |
| 登录凭据 | admin / admin123 |
| GOPROXY | `https://goproxy.cn,direct` |

## 数据源（远端已部署）

| 服务 | 端口 | 状态 |
|------|------|------|
| MySQL | 3306 | ✅ systemd |
| Redis | 6379 | ✅ systemd |
| Prometheus | 9090 | ✅ apt install，systemd |
| Node Exporter | 9100 | ✅ systemd |
| SkyWalking OAP | 11800(gRPC) / 12800(HTTP) / 9091(PromQL) | ✅ systemd skywalking-oap.service，H2 内存存储 |
| FindX API | 8080 | ✅ systemd ai-workbench-api.service |
| FindX Web (nginx) | 3000 | ✅ systemd |

## 已完成工作（本次会话）

### P0-P4 横向铺设
- CMDB 对象建模（GORM 迁移 + 分类树 + 模型 + 属性 + 实例，端到端验证）
- 仪表盘 Panel 渲染（d3 图表 + 时序/Stat/Table）
- 链路监控前端（服务/Trace/拓扑）
- 日志中心前端（检索 + 实时）
- Agent 心跳真实验证（三重检查）

### P5-P8 扩展模块
- Agent 完整生命周期（安装/升级/回滚/卸载 API + 前端向导）
- MCP Server 注册管理（4 个预置 server）
- AI 助手 SSE 流式聊天
- 组织权限 + SSO 配置
- AI SRE 增强（健康检查、工作流、Evidence Chain、知识库）

### 端到端打通
- **监控**: Prometheus → FindX API → 仪表盘真实渲染（CPU/内存/负载 node_exporter 指标）
- **告警**: PromQL 评估 → 事件 → 通知渠道（log/webhook）真实分发
- **CMDB 自动发现**: Prometheus node_uname_info → CMDB 实例 upsert（幂等）
- **模板导入**: POST /dashboard-templates/:id/import 真实创建仪表盘
- **AI 聊天**: Accept:text/event-stream → LLM /chat/completions stream:true → SSE 转发
- **Tracing 代理**: FindX /tracing/* → SkyWalking OAP GraphQL (12800)

### 布局对齐成熟源码
- 仪表盘：react-grid-layout 拖拽 + Panel 编辑器全屏 Modal + 时间范围选择器（夜莺风格）
- 指标查询：PromQL 代码风格输入 + 相对时间 + 多查询面板（夜莺风格）
- 链路监控：Condition bar + 瀑布图 + d3 力导向拓扑（SkyWalking 风格）
- 日志中心：查询构建器 + List/Chart/Table + 抽屉式详情（SigNoZ 风格）

### 关键提交（本次会话）
```
a9b71b7 feat(tracing): proxy FindX tracing API to SkyWalking OAP GraphQL
0985ae5 feat(alert): implement alert event → notification channel closure
dbe041e feat(ai-sre): implement real SSE streaming for AI chat via LLM
833e42f refactor(tracing): align UI with SkyWalking booster patterns
bbe13d5 refactor(logs): align UI with SigNoZ LogsExplorer patterns
153507e feat(templates): wire dashboard template import end-to-end
f2de99e feat(cmdb): implement auto-discovery from Prometheus node_exporter
ed44384 refactor(dashboard): rewrite detail view to match Nightingale patterns
69b17c8 refactor(explorer): rewrite metric query page to match Nightingale patterns
37f25fb feat(ai-sre): implement P8 commercial features
8baed64 feat(org): implement organization governance + SSO configuration
205d5ff feat(agent): implement complete lifecycle API and enhanced UI
53fe2df feat(mcp): implement MCP Server registry with GORM + management UI
85fc6b1 feat(ai-sre): implement AI assistant chat interface with SSE streaming
87fc13f feat(logs): implement log search and live tail UI
96b5f3e feat(agent): implement real heartbeat verification
a170c62 feat(tracing): implement service catalog and trace search UI
65ed6da feat(dashboard): implement panel chart rendering with d3
d560a27 feat(cmdb): GORM migration + object modeling UI
```

## 正在进行的工作（中断点）

**当前任务**: 修正前端 `web/src/react-shell/api/tracing.js` 的 API 路径

**问题**: 前端调用 `/apm/*` 路径，后端新实现的代理在 `/tracing/*`。需要批量替换：
- `/apm/services` → `/tracing/selectors/services`
- `/apm/endpoints` → `/tracing/selectors/endpoints`
- `/apm/instances` → `/tracing/selectors/instances`
- `/apm/traces/query` → `/tracing/traces/query`
- `/apm/traces/:id/spans` → `/tracing/traces/:id/spans`
- `/apm/topology` → `/tracing/topology`

其他路径（overview、profiling、alarms、settings）后端未实现，保持 404/BLOCKED。

## 未完成工作（路线图）

### 短期（下一阶段）
1. **前端 tracing API 路径修正**（上面中断点）
2. **日志后端接入 Loki**: 远端安装 Loki + Promtail，FindX /logs/query 代理到 Loki LogQL
3. **MCP Server 工具调用实现**: 目前只是注册，真实 CMDB MCP 需要响应 `list_tools` / `call_tool` 协议
4. **LLM API Key 配置**: 设置 `AI_WORKBENCH_API_KEY` 和 `AI_WORKBENCH_BASE_URL` 环境变量验证 AI 聊天真实对话
5. **FindX Agent APM 上报**: 让 categraf/catpaw agent 向 SkyWalking OAP 11800 gRPC 上报 Trace

### 中期
6. **通知渠道实现**: dingtalk/wecom/feishu/email 真实发送（目前是占位符）
7. **Agent 安装脚本**: POST /findx-agents/:id/install 目前存 event，需要实际执行 bash 脚本到目标主机
8. **仪表盘变量**: 支持 `$ident` `$instance` 等变量的下拉选择和动态替换
9. **告警事件持久化**: 事件进 MySQL + 分页查询 + 确认/抑制
10. **告警静默规则**: 支持标签匹配静默
11. **仪表盘分享 Token**: 生成带 token 的只读链接

### 长期
12. **生产加固**: API 限流、审计日志完整化、Redis 缓存、备份/灾恢
13. **真实测试覆盖**: 单元测试 + 集成测试 + E2E 测试套件
14. **Agent 真实部署**: 在多台主机部署 FindX Agent，验证分布式监控
15. **多租户 + 业务组隔离**
16. **SkyWalking BanyanDB 持久化**: 替换 H2 内存存储（目前重启 OAP 丢数据）
17. **多数据源支持**: Prometheus 外支持 VictoriaMetrics / Elasticsearch / TDengine
18. **AIOps 能力深化**: 异常检测、根因分析、容量预测真正接入 LLM/ML

## 重要技术决策

- **后端**: Go 1.21 / Gin / GORM + raw sql 双写（新模块用 GORM，旧模块保留 raw sql）
- **前端**: React 17 / Vite / JSX（no TypeScript）/ d3 / marked / sanitize-html / react-grid-layout
- **CSS**: `--fx-*` CSS 变量，不引入 Ant Design 等重型框架
- **API 前缀**: `/api/v1/*`，JWT 认证
- **GORM 模式**: `GetDB()` / `GormOK()` 双写（MySQL 可用时写库，否则走内存 fallback）
- **SkyWalking 存储**: H2 内存（重启丢数据，生产需要换 BanyanDB）

## 部署脚本（快速复用）

```bash
# 前端部署
cd D:\ai-workbench\web && npm run build
ssh findx-ubuntu "rm -f /opt/ai-workbench-runtime/web/dist/static/index-*.js /opt/ai-workbench-runtime/web/dist/static/index-*.css"
scp -r D:\ai-workbench\web\dist\. findx-ubuntu:/opt/ai-workbench-runtime/web/dist/

# 后端部署
ssh findx-ubuntu "cd /opt/ai-workbench/api && go build -o api-linux . && sudo install -m 0755 api-linux /opt/ai-workbench-runtime/api/ai-workbench-api && sudo systemctl restart ai-workbench-api.service"

# 同步 Go 源文件到远端（agent 通常需要这一步）
scp D:\ai-workbench\api\... findx-ubuntu:/opt/ai-workbench/api/...

# 健康检查
ssh findx-ubuntu "curl -fsS http://127.0.0.1:8080/api/v1/health/storage"
```

## 重启后恢复说明

1. 检查 systemd 服务状态：
   ```bash
   ssh findx-ubuntu "systemctl is-active prometheus skywalking-oap ai-workbench-api"
   ```
2. 如果 SkyWalking OAP 未自动启动（H2 内存存储，需手动启动）：
   ```bash
   ssh findx-ubuntu "sudo systemctl start skywalking-oap"
   ```
3. 浏览器验证：http://10.10.160.202:3000 登录 admin/admin123
4. 进入仪表盘：http://10.10.160.202:3000/dashboards?section=detail&id=dash-node-01 应能看到真实 CPU/内存/负载图表
