# FindX Fullstack Source Parity Modular Closure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不继承旧 Claude 或旧 `.codex` DONE 结论的前提下，按模块化测试重新推进 FindX 全栈一比一同源闭环：前端像素与交互同源、后端契约同源、Agent 执行链路真实闭环或结构化阻断。

**Architecture:** 计划采用“重新基线 + 模块包 + 分层验证 + 可提交闭包”的方式推进。每个模块先建立 contract matrix 和 focused tests，再补 API/前端/执行器，达到可提交闭包后才进入 Windows、WSL、远端 Ubuntu、Playwright 和 Git pathspec 全门禁；这不是最小验证，而是把全量验证从每个微改动下沉到模块闭包边界。

**Tech Stack:** Go 1.21 + Gin、MySQL/Redis、React-only Shell、Vite、Playwright MCP、WSL Ubuntu、Remote Ubuntu `findx-ubuntu`、成熟源码根 `D:\项目迁移文件\平台源码`。

---

## 0. 本计划的事实边界

本计划是 2026-05-12 重新分版计划，不能继承以下内容作为完成事实：

- 不能继承 Claude 接手期间的任何 DONE、PASS 或实现判断。
- 不能继承旧 `.codex/codex-task-board.md` 中 FX-NIGHT-104 到 FX-NIGHT-118 的完成状态。
- 不能继承旧 `.claude/**` 作为当前执行记录根。
- 不能把当前浏览器能打开、接口能返回、旧任务板写过 PASS 视为闭环完成。

每个任务开始必须重新读取当前代码、当前 Git 工作树、成熟源码、运行态和浏览器事实。旧记录只作为“待复核历史线索”，不作为验收依据。

## 1. 效率策略：模块化测试，不降低验收

以前慢的主要原因不是验证标准太高，而是验证粒度不合理：大量小改动反复触发全量 Windows/WSL/Remote/Playwright，同时脏工作树没有先形成可提交闭包。

## 1.1 前端二级/三级交互像素级同源硬门禁

前端同源迁移不能只复刻一级页面或首屏布局。每个页面切片必须覆盖成熟源码中的二级和三级交互，并按像素级位置、交互顺序、状态流和功能语义复原：

- 二级交互：二级导航、tab、筛选区展开、搜索联想、时间范围、数据源选择、变量选择、表格分页、批量操作、行内菜单、工具栏按钮、图表切换、视图保存、分享、导入导出。
- 三级交互：详情抽屉、编辑抽屉、创建/导入弹窗、确认弹窗、冲突处理、dry-run 预览、错误详情、权限拒绝、空态操作、嵌套 tab、表单动态字段、下拉级联、复制入口、命令预览、执行记录、审计/Evidence Chain 面板。
- 像素级验收：布局位置、间距、列宽、按钮位置、抽屉宽度、弹窗层级、loading/empty/error 状态、窄屏布局必须对照成熟源码 DOM 和截图证据。FindX 只允许替换品牌、主题 token、权限审计、错误脱敏、AI SRE 和 Evidence Chain；不得按自研架构重画页面。
- 功能语义验收：成熟源码中有真实动作的二级/三级控件，FindX 必须接真实 FindX 契约；契约缺失时显示精确 `BLOCKED_BY_CONTRACT` 和 `contract_gap_id`，不能隐藏控件、改成静态展示或假成功。
- 证据要求：每个前端任务必须记录 mature source 路径、组件名、交互入口、API 调用、状态枚举、DOM/截图证据、FindX 替换点和 Playwright 步骤。只看截图或只看一级页面不允许标 DONE。

新的测试层级如下：

- **L0 静态边界扫描**：工作树分桶、禁止路径、敏感信息、fake success、iframe/WebView、外部品牌、乱码、mojibake、U+FFFD。
- **L1 Focused Unit Tests**：只跑当前包或当前 handler/store/model 的单元测试，覆盖正常、异常、权限、脱敏、blocked gate。
- **L2 Module Integration Tests**：同一业务模块的 API 或前端状态流集成测试，验证 contract matrix 不返回假成功。
- **L3 Windows Full Gate**：可提交闭包形成后执行 `go test -count=1 ./...` 或 `npm.cmd run build`。
- **L4 WSL Runtime Gate**：同步当前闭包写集到 `/opt/ai-workbench`，执行 Go/web build；涉及 runtime 行为时部署 binary/dist 并检查 8080/3000 health。
- **L5 Remote Ubuntu Runtime Gate**：同步当前闭包写集到 `findx-ubuntu:/opt/ai-workbench`，执行 Go/web build；涉及 runtime 行为时部署并检查 8080/3000 health。
- **L6 Playwright Route Matrix**：只对当前模块相关路由跑真实登录、主流程、异常、权限/过期、blocked、抽屉/弹窗、390px；阶段合并时跑全平台 smoke。
- **L7 Git Pathspec Gate**：只 stage 当前任务允许写集；禁止 `git add .` 和 `git add -A`；cached diff 必须通过禁止路径和敏感扫描。

执行原则：

- 后端纯同包机械拆分：L0 + L1 + L3 + clean staged test；无 UI route 时 L6 记录 `NOT_RUN: backend-only no route change`。
- 后端 API 行为变更：L0 + L1 + L2 + L3 + L4 + L5 + curl 正常/异常/401/403/blocked/脱敏 + 相关 L6。
- 前端页面变更：L0 + focused build/type checks + L3 web build + 相关 L6；接入新 API 时补 L2/L4/L5。
- Agent 执行器/包仓库/数据到达：L0 + L1 + L2 + L3 + L4 + L5 + Linux/Windows 双环境或结构化 blocked；命令预览不算完成证据。

## 2. 当前脏区分桶门禁

### Task 0: 重新基线和工作树分桶

**Files:**
- Read: `D:\ai-workbench\AGENTS.md`
- Read: `D:\ai-workbench\README.md`
- Read: `D:\ai-workbench\docs\aiops\README.md`
- Read: `D:\ai-workbench\docs\aiops\findx_full_stack_observability_long_term_plan.md`
- Read: `D:\ai-workbench\docs\aiops\findx_react_only_frontend_long_term_plan.md`
- Read: `D:\ai-workbench\.codex\codex-task-board.md`
- Read: `D:\ai-workbench\.codex\operations-log.md`
- Read-only inspect: `D:\项目迁移文件\平台源码\**`
- Read-only inspect: `D:\测试\**`
- No write to business code.

**禁止写集:**
- `api/internal/handler/data/memory-store.json`
- `.codex/**` 以外的当前计划记录根，除非本任务只追加执行日志
- `.claude/**`
- `.playwright-mcp/**`
- `.test-evidence/**`
- `api/data/**`
- `web/dist/**`
- `web/node_modules/**`
- runtime data、真实凭据、包制品、私钥、日志文件

- [ ] **Step 1: 读取当前事实源**

Run:

```powershell
Get-Content D:\ai-workbench\AGENTS.md -TotalCount 260
Get-Content D:\ai-workbench\README.md -TotalCount 220
Get-Content D:\ai-workbench\docs\aiops\README.md -TotalCount 220
Get-Content D:\ai-workbench\docs\aiops\findx_full_stack_observability_long_term_plan.md -TotalCount 260
Get-Content D:\ai-workbench\docs\aiops\findx_react_only_frontend_long_term_plan.md -TotalCount 260
Get-Content D:\ai-workbench\.codex\codex-task-board.md -Tail 220
Get-Content D:\ai-workbench\.codex\operations-log.md -Tail 220
```

Expected: 只形成当前事实摘要，不把旧 DONE 当验收。

- [ ] **Step 2: 复核 Git 工作树**

Run:

```powershell
cd D:\ai-workbench
git status --short
git diff --stat
git diff --cached --name-status
git log --oneline -16
```

Expected:

- cached 为空，或只包含明确允许写集。
- `api/internal/handler/data/memory-store.json` 标记为运行态/敏感候选。
- 乱码路径文件如 `Dai-workbenchapigo.mod`、`Dai-workbenchapigo.sum` 先列入清理候选，不直接删除。

- [ ] **Step 3: 输出工作树分桶表**

在 `.codex/operations-log.md` 追加一条分桶摘要，必须包含：

- 文档/治理
- React-only 前端
- Go 后端
- Agent 生命周期
- 运行态/敏感候选
- 未跟踪源码
- 旧切片残留
- 测试产物/临时产物
- 删除候选

Expected: 没有 stage、commit、push。

- [ ] **Step 4: 确认成熟源码根存在**

Run:

```powershell
Get-ChildItem -Path D:\项目迁移文件\平台源码 -Directory | Select-Object -ExpandProperty Name
Get-ChildItem -Path D:\测试 -File -ErrorAction SilentlyContinue | Select-Object -ExpandProperty FullName
```

Expected: 能看到 `fe-main`、`skywalking-booster-ui-main`、`skywalking-master`、`signoz-develop`、`AutoOps-main`、`categraf-main (1)`、`catpaw-master`；`D:\测试` 中视觉素材按文件名登记。

**不能标 DONE 的条件:**
- 使用旧 `.claude` 或旧 `.codex` 的 DONE 直接跳过复核。
- 未识别 `memory-store.json` 或乱码路径文件。
- 未输出分桶表。

## 3. 后端契约总线

### Task 1: Contract Matrix Schema 和缺口状态模型

**FX-NIGHT:** `FX-NIGHT-119A-CONTRACT-MATRIX-SCHEMA`

**Files:**
- Create or modify after search: `D:\ai-workbench\api\internal\model\contract_matrix.go`
- Create or modify after search: `D:\ai-workbench\api\internal\handler\contract_matrix*.go`
- Create or modify after search: `D:\ai-workbench\api\internal\store\contract_matrix*.go`
- Modify only if route registration needed: `D:\ai-workbench\api\routes*.go`
- Test: `D:\ai-workbench\api\internal\handler\contract_matrix*_test.go`

**成熟源码证据:**
- `D:\项目迁移文件\平台源码\fe-main\src\services\**`
- `D:\项目迁移文件\平台源码\skywalking-booster-ui-main\src\graphql\**`
- `D:\项目迁移文件\平台源码\signoz-develop\frontend\src\api\**`
- `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main\api\api\**`
- `D:\项目迁移文件\平台源码\categraf-main (1)\**`
- `D:\项目迁移文件\平台源码\catpaw-master\**`

**状态枚举:**

```go
const (
    ContractReady             = "ready"
    ContractBlocked           = "blocked"
    ContractMissingBackend    = "missing_backend"
    ContractMissingDatasource = "missing_datasource"
    ContractMissingExecutor   = "missing_executor"
    ContractUnsafe            = "unsafe"
)
```

**错误响应规则:**

```json
{
  "code": "BLOCKED_BY_CONTRACT",
  "message": "能力缺少执行器或数据源契约，已阻断",
  "contract_gap_id": "FX-CONTRACT-AGENT-EXECUTOR-SSH",
  "status": "missing_executor",
  "safe_to_retry": false
}
```

- [ ] **Step 1: 搜索现有 contract/blocked 实现**

Run:

```powershell
cd D:\ai-workbench
rg -n "BLOCKED_BY_CONTRACT|missing_executor|contract_gap|blocked_by_contract|data_arrival|fake success|queued|succeeded" api
```

Expected: 得到复用点清单；若已有状态枚举，优先集中扩展，不新建平行枚举。

- [ ] **Step 2: 写 focused tests**

Tests must cover:

- `ready` 只能用于真实契约存在且可执行的路径。
- 缺 handler 返回 `missing_backend`。
- 缺数据源返回 `missing_datasource`。
- 缺执行器返回 `missing_executor`。
- 不安全输入返回 `unsafe`。
- 响应不得包含 token、password、cookie、DSN、private key。
- 不得把 `queued`、`running`、`succeeded`、`installed`、`data_arrived` 用于 blocked path。

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "Contract|Blocked|Matrix"
```

Expected before implementation: FAIL only because missing schema/handler.

- [ ] **Step 3: 实现最小 contract matrix 服务边界**

Implementation must stay inside model/store/handler/service-style boundary. Handler 只做参数、权限、响应；业务判断在对应 service/store 层或已有项目模式中。

- [ ] **Step 4: 验证 focused 和全量 Go**

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "Contract|Blocked|Matrix"
go test -count=1 ./...
```

Expected: PASS。

**不能标 DONE 的条件:**
- 前端 blocked 没有 `contract_gap_id`。
- blocked 路径出现假 `queued/running/succeeded/installed/data_arrived`。
- 使用字符串散落状态而非集中枚举。

### Task 2: 基础监控后端缺口矩阵

**FX-NIGHT:** `FX-NIGHT-119B-MONITORING-CONTRACT-MATRIX`

**Files:**
- Read source: `D:\项目迁移文件\平台源码\fe-main\src\services\**`
- Candidate modify: `D:\ai-workbench\api\routes_monitor.go`
- Candidate modify: `D:\ai-workbench\api\internal\handler\monitoring_*.go`
- Candidate modify: `D:\ai-workbench\api\internal\store\monitoring_*.go`
- Candidate modify: `D:\ai-workbench\api\internal\model\monitoring_*.go`
- Test: monitoring handler/store tests in the same package.

**Capabilities:**

- 数据源
- 系统集成
- 指标查询
- 内置指标
- 仪表盘
- 模板中心
- 告警规则
- 记录规则
- 事件
- 屏蔽
- 订阅
- 自愈
- 通知规则
- 通知媒介
- 消息模板
- 组织权限
- 系统配置

- [ ] **Step 1: 从 Nightingale services 抽取 API 行为**

Run:

```powershell
rg -n "request|fetch|axios|post|get|put|delete|import|export|clone|enable|disable|test|preview" D:\项目迁移文件\平台源码\fe-main\src\services
```

Expected: 生成能力到 API 行为映射表。

- [ ] **Step 2: 对 FindX 当前 API 做差异扫描**

Run:

```powershell
rg -n "datasource|integration|dashboard|template|alert|notify|notification|builtin|record|shield|subscribe|self|org|system" D:\ai-workbench\api
```

Expected: 找出 `ready / missing_backend / missing_datasource / unsafe`。

- [ ] **Step 3: 补 contract matrix 返回，不补假实现**

对每个缺口返回明确 gap id，例如：

- `FX-CONTRACT-MONITORING-DATASOURCE-TEST`
- `FX-CONTRACT-MONITORING-TEMPLATE-IMPORT`
- `FX-CONTRACT-MONITORING-ALERT-CLONE`
- `FX-CONTRACT-MONITORING-NOTIFICATION-DRYRUN`

- [ ] **Step 4: Focused tests**

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "Monitoring|Datasource|Template|Alert|Notification|Builtin"
go test -count=1 ./internal/store -run "Monitoring|Datasource|Template|Alert|Notification|Builtin"
```

Expected: PASS；blocked path 有 gap id，无假成功。

### Task 3: SkyWalking 后端缺口矩阵

**FX-NIGHT:** `FX-NIGHT-119C-APM-CONTRACT-MATRIX`

**Files:**
- Read: `D:\项目迁移文件\平台源码\skywalking-booster-ui-main\src\graphql\base.ts`
- Read: `D:\项目迁移文件\平台源码\skywalking-booster-ui-main\src\graphql\fragments\trace.ts`
- Read: `D:\项目迁移文件\平台源码\skywalking-booster-ui-main\src\graphql\fragments\topology.ts`
- Read: `D:\项目迁移文件\平台源码\skywalking-booster-ui-main\src\graphql\fragments\selector.ts`
- Read: `D:\项目迁移文件\平台源码\skywalking-booster-ui-main\src\graphql\fragments\profile.ts`
- Read: `D:\项目迁移文件\平台源码\skywalking-master\oap-server\**`
- Candidate modify: `D:\ai-workbench\api\routes_tracing.go`
- Candidate modify: `D:\ai-workbench\api\internal\handler\apm_*.go`
- Candidate modify: `D:\ai-workbench\api\internal\store\apm_*.go`
- Candidate modify: `D:\ai-workbench\api\internal\model\apm_*.go`

**Required contract endpoints:**

- `/api/v1/apm/selectors/services`
- `/api/v1/apm/selectors/instances`
- `/api/v1/apm/selectors/endpoints`
- `/api/v1/apm/topology`
- `/api/v1/apm/traces`
- `/api/v1/apm/traces/:traceId`
- `/api/v1/apm/traces/:traceId/spans/:spanId`
- `/api/v1/apm/profiling/tasks`
- `/api/v1/apm/alarms`
- `/api/v1/apm/settings`
- `/api/v1/apm/agent-linkage`

- [ ] **Step 1: 抽取 GraphQL/query 语义**

Run:

```powershell
rg -n "query|mutation|service|instance|endpoint|trace|span|topology|profile|alarm|settings" D:\项目迁移文件\平台源码\skywalking-booster-ui-main\src\graphql D:\项目迁移文件\平台源码\skywalking-master\oap-server
```

Expected: 每个 endpoint 有源 API、字段和错误态证据。

- [ ] **Step 2: FindX API 差异扫描**

Run:

```powershell
rg -n "apm|trace|span|topology|profil|alarm|selector|agent-linkage" D:\ai-workbench\api
```

Expected: 记录 ready 和 blocked 缺口，不创建静态 trace。

- [ ] **Step 3: 补缺口矩阵和 tests**

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "APM|Trace|Topology|Profil|AgentLinkage"
```

Expected: OAP 不可用时返回脱敏 `BLOCKED_BY_CONTRACT` 或 `missing_datasource`；不返回假 Trace。

### Task 4: SigNoZ 日志后端缺口矩阵

**FX-NIGHT:** `FX-NIGHT-119D-LOGS-CONTRACT-MATRIX`

**Files:**
- Read: `D:\项目迁移文件\平台源码\signoz-develop\frontend\src\AppRoutes\routes.ts`
- Read: `D:\项目迁移文件\平台源码\signoz-develop\frontend\src\components\ExplorerCard\**`
- Read: `D:\项目迁移文件\平台源码\signoz-develop\frontend\src\components\LogDetail\**`
- Read: `D:\项目迁移文件\平台源码\signoz-develop\frontend\src\container\LogsExplorer\**`
- Read: `D:\项目迁移文件\平台源码\signoz-develop\frontend\src\pages\LiveLogs\**`
- Read: `D:\项目迁移文件\平台源码\signoz-develop\frontend\src\container\PipelinePage\**`
- Candidate modify: `D:\ai-workbench\api\routes_logs.go`
- Candidate modify: `D:\ai-workbench\api\internal\handler\logs*.go`
- Candidate modify: `D:\ai-workbench\api\internal\store\logs*.go`
- Candidate modify: `D:\ai-workbench\api\internal\model\logs*.go`

**Required contract endpoints:**

- `/api/v1/logs/query`
- `/api/v1/logs/fields`
- `/api/v1/logs/context`
- `/api/v1/logs/aggregate`
- `/api/v1/logs/live`
- `/api/v1/logs/pipelines`
- `/api/v1/logs/pipelines/:id/deploy`
- `/api/v1/logs/pipelines/:id/rollback`
- `/api/v1/logs/saved-views`
- `/api/v1/logs/trace-link`
- `/api/v1/logs/agent-linkage`

- [ ] **Step 1: 抽取 SigNoZ Logs 状态流**

Run:

```powershell
rg -n "query|fields|context|aggregate|live|pipeline|saved|view|trace|log detail|drawer|SSE|websocket" D:\项目迁移文件\平台源码\signoz-develop\frontend\src
```

Expected: 每个日志能力都有源组件/API/状态/错误态证据。

- [ ] **Step 2: 差异扫描**

Run:

```powershell
rg -n "logs|fields|context|aggregate|live|pipeline|saved|trace-link|agent-linkage|loki|clickhouse" D:\ai-workbench\api
```

Expected: Loki proxy 只能标兼容来源，不能替代 SigNoZ Query Service/ClickHouse 语义。

- [ ] **Step 3: Focused tests**

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "Logs|Pipeline|Saved|TraceLink|Live"
```

Expected: audit source、Loki proxy、SigNoZ semantic gaps 均有明确 gap id；pipeline deploy/rollback 缺执行器时不返回成功。

### Task 5: CMDB / Agent 后端缺口矩阵

**FX-NIGHT:** `FX-NIGHT-119E-CMDB-AGENT-CONTRACT-MATRIX`

**Files:**
- Read: `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main\api\api\cmdb\**`
- Read: `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main\api\api\tool\**`
- Read: `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main\api\api\monitor\**`
- Read: `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main\api\common\util\websocket\**`
- Candidate modify: `D:\ai-workbench\api\routes_cmdb.go`
- Candidate modify: `D:\ai-workbench\api\internal\handler\cmdb_*.go`
- Candidate modify: `D:\ai-workbench\api\internal\store\cmdb_*.go`
- Candidate modify: `D:\ai-workbench\api\internal\model\cmdb_*.go`

**Capabilities:**

- CMDB 树、模型、实例、关系、回收站
- 主机、云主机、数据库资产 MySQL/PostgreSQL/Redis/Elasticsearch/MongoDB
- 主机凭据、跳板机、终端会话、文件上传
- 命令执行、部署任务、Agent 心跳
- 终端录像审计，缺回放契约时 blocked

**Persistence boundary:**

- 旧实施文档 `docs/aiops/findx_implementation_strategy_v3.md` 已说明 CMDB / Agent 等新模块使用 GORM `GetDB()` 模式；本任务作为当前实施入口必须同步执行同一边界。
- CMDB 自动发现结果，包括 Prometheus、FindX Agent、Categraf、Catpaw 来源，必须通过 `store.CreateCmdbInstance` / `store.UpdateCmdbInstance` 写入。生产 MySQL 可用时，`store.GormOK()` 必须为真并经 `store.GetDB()` 落到 GORM/MySQL。
- `api/internal/handler/data/memory-store.json` 只是运行态 fallback/sensitive candidate，不得作为 CMDB 自动发现、Agent 生命周期完成态或生产持久化成功证据。
- GORM 不可用时只能标记 `BLOCKED`、`RISK` 或 `memory fallback dev-only`，不得宣称生产持久化成功。

- [ ] **Step 1: AutoOps API 行为抽取**

Run:

```powershell
rg -n "cmdb|host|asset|database|mysql|redis|mongo|elastic|postgres|terminal|upload|command|deploy|agent|heartbeat|websocket|audit" D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main\api
```

Expected: 得到 CMDB/Agent 能力映射。

- [ ] **Step 2: FindX API 差异扫描**

Run:

```powershell
rg -n "cmdb|asset|host|database|terminal|upload|command|deploy|agent|heartbeat|credential|jump|audit" D:\ai-workbench\api
```

Expected: 标记每项 ready/missing/backend/executor/unsafe。

- [ ] **Step 3: Focused tests**

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "CMDB|Host|Database|Terminal|Deploy|Agent"
go test -count=1 ./internal/store -run "CMDB|Host|Database|Terminal|Deploy|Agent"
```

Expected: 终端、部署、命令执行、录像回放缺执行器时 blocked，不显示成功；CMDB 自动发现必须覆盖 MySQL/GORM 持久化、`store.GormOK()` + `store.GetDB()` 路径、重启后发现实例仍在，且 fallback 不写入或污染 Git 中的 `memory-store.json`。

## 4. Agent 生命周期模块包

### Task 6: Agent Package Repository

**FX-NIGHT:** `FX-NIGHT-120-AGENT-PACKAGE-REPOSITORY`

**Files:**
- Candidate create/modify: `D:\ai-workbench\api\internal\model\findx_agent_package*.go`
- Candidate create/modify: `D:\ai-workbench\api\internal\store\findx_agent_package*.go`
- Candidate create/modify: `D:\ai-workbench\api\internal\handler\findx_agent_package*.go`
- Candidate modify route: `D:\ai-workbench\api\routes_findx*.go` or existing routes file after search
- Candidate frontend: `D:\ai-workbench\web\src\react-shell\agents\**`
- Test: same package Go tests and Playwright `/agents?section=packages` or equivalent route.

**Package model fields:**

- `package_id`
- `name`
- `capability`
- `language`
- `platform`
- `arch`
- `version`
- `commit`
- `license`
- `notice`
- `checksum`
- `signature`
- `offline_path`
- `install_modes`

**Package families:**

- FindX Agent core
- SkyWalking Java/Python/Node.js/PHP/Go/Rust/Ruby/Nginx Lua/Kong/Browser Client JS
- Categraf
- Categraf plugins
- Catpaw
- 日志采集插件
- gateway trace/RUM/profiling 能力包

- [ ] **Step 1: 源码和许可证证据**

Run:

```powershell
rg -n "LICENSE|NOTICE|version|release|checksum|signature|agent|plugin|package" D:\项目迁移文件\平台源码\skywalking-master D:\项目迁移文件\平台源码\categraf-main` (1`) D:\项目迁移文件\平台源码\catpaw-master
```

Expected: 每个包类型有来源、许可证/NOTICE、可安装形态；缺本地包时 `missing_package_artifact`，不造包。

- [ ] **Step 2: 后端 focused tests**

APIs:

- list
- detail
- download
- verify
- enable
- disable
- import-offline-package

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "FindXAgentPackage|PackageRepository|OfflinePackage"
go test -count=1 ./internal/store -run "FindXAgentPackage|PackageRepository"
```

Expected: checksum mismatch、missing signature、unsafe path、missing artifact 均 blocked；download 不泄露本地绝对敏感路径。

- [ ] **Step 3: 前端 focused route**

Playwright route coverage:

- 能力包目录
- 平台/语言/能力筛选
- 版本详情
- 许可证/NOTICE
- 校验状态
- 离线包导入入口
- 缺签名 `BLOCKED_BY_CONTRACT`
- 390px

Expected: 用户侧只显示 FindX Agent 命名；外部产品名仅出现在内部证据或许可证/NOTICE 合规上下文。

### Task 7: Agent Install Executor State Machine

**FX-NIGHT:** `FX-NIGHT-121-AGENT-INSTALL-EXECUTOR-MATRIX`

**Files:**
- Candidate modify: `D:\ai-workbench\api\internal\handler\findx_agent_lifecycle*.go`
- Candidate modify: `D:\ai-workbench\api\internal\security\*_installer.go`
- Candidate modify: `D:\ai-workbench\api\internal\store\findx_agent_*`
- Candidate modify: `D:\ai-workbench\api\internal\model\findx_agent_*`
- Candidate frontend: `D:\ai-workbench\web\src\react-shell\agents\**`

**Required state machine:**

```text
planned
preflight_failed
blocked_by_contract
dispatching
running
receipt_pending
service_registered
heartbeat_seen
data_arrival_seen
failed
rolled_back
uninstalled
```

**Execution scenes:**

- Linux local: `curl -kfsSL`、systemd、status、uninstall、upgrade、rollback
- Windows local: CMD `certutil -urlcache -f`、PowerShell `Invoke-WebRequest`、Windows Service、status、uninstall、upgrade、rollback
- Remote: SSH、WinRM、文件下发、命令执行、服务注册、日志拉取、超时、重试、幂等、失败恢复
- Gateway: IIS、Nginx Lua、Kong、gateway trace 数据到达
- Kubernetes: Helm、Operator、DaemonSet、Sidecar、InitContainer、RBAC、namespace/workload、rollback

- [ ] **Step 1: 状态枚举集中化扫描**

Run:

```powershell
rg -n "planned|preflight_failed|blocked_by_contract|dispatching|running|receipt_pending|service_registered|heartbeat_seen|data_arrival_seen|failed|rolled_back|uninstalled|queued|succeeded|success|installed" D:\ai-workbench\api
```

Expected: 旧 `queued/succeeded/installed` 在 blocked path 被替换或防护；真实执行路径才允许进入运行态。

- [ ] **Step 2: 后端 tests**

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/security -run "Installer|Executor|Service|Systemd|Windows|Kubernetes"
go test -count=1 ./internal/handler -run "FindXAgent|Install|Executor|Receipt|Rollback|Uninstall"
```

Expected:

- 缺 SSH/WinRM/systemd/Windows Service/K8s executor 返回 `missing_executor`。
- 任务创建不等于执行成功。
- heartbeat 不等于 logs/traces/profiling/inspection/RUM/gateway trace 到达。
- 命令预览不等于安装完成。

- [ ] **Step 3: UI 抽屉和真实执行按钮门禁**

Agent 抽屉必须显示：

- Linux/Windows/K8s 安装入口
- `curl` / `certutil` / `Invoke-WebRequest` 复制入口
- 命令预览
- 真实执行按钮或精确 blocked gate
- 执行任务、日志、审计、Evidence Chain

Expected: 复制按钮成功只表示复制成功，不表示安装成功。

### Task 8: Categraf Plugin Remote Mutation

**FX-NIGHT:** `FX-NIGHT-122-CATEGRAF-PLUGIN-REMOTE-MUTATION`

**Files:**
- Read: `D:\项目迁移文件\平台源码\categraf-main (1)\**`
- Candidate modify: `D:\ai-workbench\api\internal\handler\findx_agent_lifecycle_config*.go`
- Candidate modify: `D:\ai-workbench\api\internal\store\findx_agent_*`
- Candidate frontend: `D:\ai-workbench\web\src\react-shell\agents\**`

**Plugin coverage:**

- CPU、mem、disk、net、process、docker、mysql、postgresql、redis、mongodb、elasticsearch、nginx、prometheus scrape

**Operations:**

- preview
- canary
- apply
- reload
- drift detect
- rollback
- audit
- evidence chain

- [ ] **Step 1: 源码抽取**

Run:

```powershell
rg -n "inputs|toml|reload|provider|drift|rollback|mysql|postgres|redis|mongodb|elasticsearch|nginx|prometheus|docker|process" D:\项目迁移文件\平台源码\categraf-main` (1`)
```

Expected: 每个插件模板有参数、敏感字段、reload 语义来源。

- [ ] **Step 2: 后端 tests**

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "Categraf|Plugin|Config|Rollout|Reload|Drift|Rollback"
```

Expected: 无远程 writer/reload receipt/drift checker 时返回 `BLOCKED_BY_CONTRACT`，不显示“修改成功”。

- [ ] **Step 3: Browser route**

Playwright coverage:

- 单 Agent、CMDB 主机、业务组、资源组、namespace/workload 范围选择
- TOML/YAML/JSON 预览
- 参数校验
- 敏感字段脱敏
- canary/apply/reload/rollback blocked 可见
- 390px

Expected: Linux/Windows 双范围可见；缺执行器时精确 blocked。

### Task 9: Data Arrival Validator Matrix

**FX-NIGHT:** `FX-NIGHT-123-DATA-ARRIVAL-VALIDATOR`

**Files:**
- Candidate modify: `D:\ai-workbench\api\internal\handler\findx_agent_lifecycle_data*.go`
- Candidate modify: `D:\ai-workbench\api\internal\store\findx_agent_*`
- Candidate modify: `D:\ai-workbench\api\internal\model\findx_agent_*`
- Candidate frontend: `D:\ai-workbench\web\src\react-shell\agents\**`

**Signals:**

- metrics
- logs
- traces
- profiling
- inspection
- RUM
- gateway trace

Each signal must include:

- signal type
- source agent
- package version
- config version
- first_seen_at
- last_seen_at
- sample evidence
- backend receiver
- related trace/log/metric ids
- blocked reason

- [ ] **Step 1: Receiver mapping scan**

Run:

```powershell
rg -n "heartbeat|remote_write|prometheus|logs|trace|span|profile|inspection|rum|gateway|receiver|arrival|evidence" D:\ai-workbench\api
```

Expected: 每个 signal 映射到真实 receiver 或 blocked reason。

- [ ] **Step 2: Tests**

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "DataArrival|Receiver|Heartbeat|Metrics|Logs|Traces|Profiling|Inspection|RUM|Gateway"
```

Expected: metrics/heartbeat 不能替代其他 signal；缺 receiver 返回 blocked reason。

- [ ] **Step 3: Browser evidence**

Playwright `/agents?section=hosts` drawer coverage:

- 每个 signal 的 reported/missing/blocked 状态
- sample evidence 脱敏
- related ids 深链
- 390px 不溢出

Expected: 不能把 heartbeat 显示为全量数据到达。

## 5. SkyWalking / SigNoZ / CMDB / 模板 / Evidence Chain

### Task 10: SkyWalking OAP Adapter

**FX-NIGHT:** `FX-NIGHT-124-SKYWALKING-OAP-ADAPTER`

**Files:**
- Candidate backend: `D:\ai-workbench\api\routes_tracing.go`
- Candidate backend: `D:\ai-workbench\api\internal\handler\apm_*.go`
- Candidate backend: `D:\ai-workbench\api\internal\store\apm_*.go`
- Candidate backend: `D:\ai-workbench\api\internal\model\apm_*.go`
- Candidate frontend: `D:\ai-workbench\web\src\react-shell\tracing\**`

- [ ] **Step 1: 实现服务/实例/端点 selector adapter**

Focused tests:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "APMSelector|Service|Instance|Endpoint"
```

- [ ] **Step 2: 实现 topology adapter**

Tests cover node/edge/metrics/hover/expand/focus/deep-link。

- [ ] **Step 3: 实现 trace list/detail/span/log linkage**

Tests cover trace not found、span not found、OAP unavailable、agent linkage missing。

- [ ] **Step 4: Playwright tracing matrix**

Routes:

- `/tracing?section=services`
- `/tracing?section=topology`
- `/tracing?section=traces`
- `/tracing/<traceId>?section=trace-detail`

Expected: OAP 不可用时脱敏 blocked，不造静态 trace。

### Task 11: SigNoZ Logs Adapter

**FX-NIGHT:** `FX-NIGHT-125-SIGNOZ-LOGS-ADAPTER`

**Files:**
- Candidate backend: `D:\ai-workbench\api\routes_logs.go`
- Candidate backend: `D:\ai-workbench\api\internal\handler\logs*.go`
- Candidate backend: `D:\ai-workbench\api\internal\store\logs*.go`
- Candidate backend: `D:\ai-workbench\api\internal\model\logs*.go`
- Candidate frontend: `D:\ai-workbench\web\src\react-shell\logs\**`

- [ ] **Step 1: Query/fields/context/aggregate**

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "LogsQuery|LogsFields|LogsContext|LogsAggregate"
```

Expected: Query Service/ClickHouse 缺失时 precise blocked；FindX audit source 不冒充生产日志。

- [ ] **Step 2: live tail SSE/WebSocket**

Tests cover pause/resume、disconnect/reconnect、unauthorized、blocked datasource。

- [ ] **Step 3: pipelines deploy/rollback 和 saved views**

Tests cover CRUD、deploy missing executor、rollback missing receipt、audit、脱敏。

- [ ] **Step 4: Playwright logs matrix**

Routes:

- `/logs?section=query`
- `/logs?section=context`
- `/logs?section=aggregate`
- `/logs?section=live`
- `/logs?section=pipelines`
- `/logs?section=saved-views`

Expected: drawer、fields sidebar、raw/json、trace link、390px 全覆盖。

### Task 12: AutoOps CMDB + Visual Parity

**FX-NIGHT:** `FX-NIGHT-126-AUTOOPS-CMDB-TEMPLATE-PARITY`

**Files:**
- Read: `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main\web\**`
- Read: `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main\api\api\cmdb\**`
- Read: `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main\api\api\tool\**`
- Read: `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main\api\api\monitor\**`
- Read images: `D:\测试\cmdb-instance-list.png`
- Read images: `D:\测试\cmdb-instance-detail.png`
- Read images: `D:\测试\cmdb-model-manage.png`
- Read images: `D:\测试\cmdb-business-view.png`
- Read images: `D:\测试\business-manage-overview.png`
- Read images: `D:\测试\devops-host.png`
- Read images: `D:\测试\monitor-resources.png`
- Read images: `D:\测试\profile-globalconfig.png`
- Candidate backend: `D:\ai-workbench\api\routes_cmdb.go`
- Candidate backend: `D:\ai-workbench\api\internal\handler\cmdb_*.go`
- Candidate frontend: `D:\ai-workbench\web\src\react-shell\cmdb\**`

- [ ] **Step 1: 建立视觉和交互基准**

Use Playwright screenshots and local images to compare:

- 树 + 表格 + 详情抽屉
- 主机表和主机 Agent 合并语义
- 数据库资产五类
- 终端/上传/命令/部署任务
- Agent 抽屉 tabs

Expected: 不按 FindX 自己重新设计布局；只做 FindX 品牌替换。

- [ ] **Step 2: 后端 focused tests**

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "CMDB|Host|Database|Terminal|Upload|Deploy|Agent"
```

- [ ] **Step 3: Frontend route matrix**

Routes:

- `/assets?section=overview`
- `/assets?section=hosts`
- `/assets?section=databases`
- `/assets?section=terminal`
- `/assets?section=deploy`
- `/agents?section=hosts`

Expected: CMDB 新增主机自动出现在主机 Agent；未安装显示未安装；已安装显示已安装；Agent 抽屉包含安装记录、配置版本、插件配置、数据到达、执行任务、审计、Evidence Chain。

Persistence expected: FX-NIGHT-126 不得使用 `memory-store.json` 作为 CMDB 新增主机或 Agent 生命周期完成态依据。必须验证 MySQL/GORM 持久化、服务重启后数据仍可查询、fallback 仅为 dev-only 风险路径且不进入 Git。

### Task 13: Nightingale Template Center

**FX-NIGHT:** `FX-NIGHT-127-N9E-TEMPLATE-CENTER-BACKEND-FRONTEND`

**Files:**
- Read: `D:\项目迁移文件\平台源码\fe-main\src\pages\builtInComponents\**`
- Read: `D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`
- Read: `D:\项目迁移文件\平台源码\fe-main\src\services\warning.ts`
- Read: `D:\项目迁移文件\平台源码\fe-main\src\components\DocumentDrawer\**`
- Candidate backend: monitoring template handlers/stores/models.
- Candidate frontend: `D:\ai-workbench\web\src\react-shell\integrations\**` or template center route after search.

- [ ] **Step 1: 后端 template import/dry-run/rollback**

Tests cover:

- 分类
- 组件详情
- payload 版本
- dashboard import
- alert import
- collect import
- record rule import
- 冲突检查
- dry-run preview
- rollback
- 审计
- 业务组权限
- 导入失败错误脱敏

- [ ] **Step 2: 前端 template center parity**

Playwright coverage:

- 分类图标
- 模板列表
- 模板详情
- 文档抽屉
- payload 预览
- 导入弹窗
- 冲突列表
- 失败回滚状态
- 导入结果详情
- 跳转到仪表盘/告警/采集配置

Expected: 缺契约时精确 blocked 到动作，不整页 blocked。

### Task 14: Evidence Chain Contract

**FX-NIGHT:** `FX-NIGHT-128-EVIDENCE-CHAIN-CONTRACT`

**Files:**
- Candidate backend: `D:\ai-workbench\api\internal\handler\aiops_evidence_chain*.go`
- Candidate backend: `D:\ai-workbench\api\internal\store\aiops_evidence*.go`
- Candidate backend: `D:\ai-workbench\api\internal\model\aiops_evidence*.go`
- Candidate frontend: `D:\ai-workbench\web\src\react-shell\aiops\**`

**Endpoints:**

- `/api/v1/aiops/evidence/search`
- `/api/v1/aiops/evidence/chains`
- `/api/v1/aiops/evidence/chains/:id`
- `/api/v1/aiops/diagnosis/sessions/:id/evidence`
- `/api/v1/aiops/reports/:id/export`
- `/api/v1/aiops/remediation/plans`
- `/api/v1/aiops/remediation/plans/:id/execute`

- [ ] **Step 1: evidence source model**

Sources:

- metrics
- logs
- trace
- CMDB
- Agent install
- Agent config
- Categraf plugin
- Catpaw inspection
- alert
- deployment/change
- workflow execution

- [ ] **Step 2: AI anti-fabrication tests**

Run:

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run "Evidence|Diagnosis|Remediation|Report"
```

Expected: 证据缺失显示缺失；AI 不编造；自动修复必须审批、执行计划、回滚、审计。

## 6. 前端像素一比一批次

### Task 15: React-only Pixel Parity Batches

**FX-NIGHT:** `FX-NIGHT-129-REACT-PIXEL-PARITY-BATCHES`

**Files:**
- Candidate frontend: `D:\ai-workbench\web\src\react-shell\**`
- Candidate frontend: `D:\ai-workbench\web\src\components\react\**`
- Candidate API wrappers: `D:\ai-workbench\web\src\api\*.js`
- Do not expand Vue workbench as final baseline.

**Batch order:**

1. Shell/Nav
2. `/query?section=metrics`
3. `/integrations?section=datasources`
4. `/dashboards?section=list`
5. `/alerts?section=rules`
6. `/logs?section=query`
7. `/tracing?section=services`
8. `/assets?section=overview`
9. `/agents?section=hosts`

- [ ] **Step 1: Per-route mature source DOM map**

For each route record:

- source path
- route
- components
- API calls
- state flow
- buttons/actions
- empty state
- error state
- permission state
- FindX replacement points

- [ ] **Step 2: Per-route Playwright matrix**

Each route:

- real login
- navigation click
- main flow
- abnormal input
- permission/expired auth
- drawer/modal/dropdown/search/pagination
- 390px
- console error 0 or explained pre-login 401
- iframe/WebView 0
- external user-side brand 0
- mojibake 0
- sensitive text 0
- fake/mock/placeholder/static success 0

Expected: 前端无法接入的动作显示 precise `BLOCKED_BY_CONTRACT`，并能映射到 Task 1 contract gap id。

## 7. 多 Agent 编排

默认并行：

- `go-backend` worker：后端契约、store、handler、状态机；写集限定 `api/**`。
- `react-frontend` worker：React-only 页面、像素一比一、API 封装；写集限定 `web/src/react-shell/**`、`web/src/components/react/**`、必要 `web/src/api/**`。
- `qa-tester` explorer：只读；成熟源码证据、浏览器矩阵、敏感/fake-state/品牌/iframe/乱码扫描。
- 主 agent：编排、同步、构建、Playwright、Git pathspec、任务板和 operations log 门禁。

每个子 agent prompt 必须包含：

- FX-NIGHT id
- role/id
- 目标
- 允许写集
- 禁止写集
- 成熟源码证据路径
- 验收标准
- Windows/WSL/Remote Ubuntu 验证命令
- 关闭要求

同一文件、同一页面、同一契约面不得并行写。子 agent 完成、失败、BLOCKED、超时或不再需要时立即关闭。

## 8. Git 和发布门禁

### Task 16: 可提交闭包和 Git Gate

**Files:**
- No code write.
- Stage only explicit pathspec of the current module closure.

- [ ] **Step 1: staged-only 检查**

Run:

```powershell
cd D:\ai-workbench
git diff --cached --name-status
git diff --cached --check
git diff --cached --name-only
```

Expected: 只包含当前任务允许写集。

- [ ] **Step 2: 禁止路径扫描**

Run:

```powershell
git diff --cached --name-only | Select-String -Pattern '(^|/)(\.codex|\.claude|\.playwright-mcp|\.test-evidence)(/|$)|memory-store\.json|api/data/|web/dist/|web/node_modules/|\.pem$|\.key$|\.log$|\.exe$|\.env$'
```

Expected: 无输出。

- [ ] **Step 3: 敏感和 fake-state 扫描**

Run:

```powershell
git diff --cached | Select-String -Pattern 'Bearer\s+[A-Za-z0-9._-]+|password\s*[:=]|cookie\s*[:=]|private key|BEGIN .*PRIVATE KEY|DSN|queued|succeeded|installed|data_arrived|mock|placeholder|fake'
```

Expected: 无真实敏感信息；测试向量必须逐条说明；blocked path 不允许假成功状态。

- [ ] **Step 4: 提交**

Commit message format:

```text
<scope>: <module closure summary>

- Source evidence: <mature source paths>
- Contract: <ready/blocked/missing ids>
- Tests: <focused/full/runtime/browser>
- Risk: <explicit residual risk or none>
```

Expected: 小提交，不混 docs/backend/frontend/runtime。

## 9. 当前优先级

立即执行顺序：

1. Task 0：重新基线和工作树分桶。
2. Task 1：Contract Matrix Schema。
3. Task 2-5：按后端域拆 contract matrix，不再把所有后端缺口塞进一个巨大 119。
4. Task 6-9：Agent 包仓库、执行器状态机、Categraf 远程配置、数据到达。
5. Task 10-14：SkyWalking、SigNoZ、CMDB、模板中心、Evidence Chain。
6. Task 15：前端像素一比一批次。
7. Task 16：每个可提交闭包的 Git gate。

## 10. 自检结果

Spec coverage:

- 后端契约总线：Task 1-5 覆盖。
- Agent 生命周期：Task 6-9 覆盖。
- SkyWalking：Task 10 覆盖。
- SigNoZ：Task 11 覆盖。
- CMDB/模板/测试素材：Task 12-13 覆盖。
- AI SRE/Evidence Chain：Task 14 覆盖。
- 前端像素一比一：Task 15 覆盖。
- 工作树治理和高效率模块化测试：Task 0、L0-L7、Task 16 覆盖。

Placeholder scan:

- 本计划未使用 TBD/TODO/implement later。
- 每个任务均包含明确文件边界、证据路径、测试命令和不能标 DONE 的条件或验收规则。

Type/status consistency:

- Contract status 使用 `ready / blocked / missing_backend / missing_datasource / missing_executor / unsafe`。
- Agent execution status 使用统一状态机，不混用 `queued/succeeded/installed/data_arrived` 作为 blocked path 状态。
