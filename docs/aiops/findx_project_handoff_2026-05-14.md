# FindX 项目交接总览

更新时间：2026-05-14 09:20（UTC+8）

本文用于新任务窗口快速接手 `D:\项目迁移文件\ai-workbench`。它只记录当前闭环事实、未提交写集、风险和下一步，不替代 `.codex/codex-task-board.md` 与 `.codex/operations-log.md`。

## 项目定位

FindX 是一个 React-only 前端加 Go API 后端的智能运维 / AIOps 平台。当前开发策略是按成熟源码、抓包和截图做同源迁移与契约闭环，缺少真实后端、真实执行器或真实数据源时必须返回或展示 `BLOCKED_BY_CONTRACT`，禁止把空数组、静态按钮、加载态或演示数据当成功。

主要能力域：

- 基础监控、数据源、指标查询、仪表盘、模板中心。
- 告警、通知、组织与系统配置。
- CMDB / 资产 / Agent 在线。
- 链路追踪、日志中心、业务拨测状态页。
- 后续 Evidence Chain、Agent 生命周期、真实执行器和数据到达回执。

## 当前硬约束

- 对话和文档使用简体中文，时间格式 `YYYY-MM-DD HH:mm（UTC+8）`。
- 用户明确要求：禁止降级到 5.4，无论主 agent、子 agent 还是其它执行角色。
- 不要 stage、commit、push，除非用户明确要求。
- 不要重置或回滚历史脏树；工作树中有大量无关旧改动，必须按显式 pathspec 操作。
- `api/internal/handler/data/memory-store.json` 是运行态 / 敏感候选文件，不能作为完成证据，不能 stage。
- 禁止 iframe、WebView、object、embed。
- 禁止假成功态：空 `nodes/edges`、空 monitor bindings、`queued/running/applied/installed/data_arrived/service_registered/rolled_back/uninstalled`、假 `READY`、假“绑定成功”等都不能作为成功。
- Service 是边界：业务逻辑放在对应 Service/handler/store/model 内，不能散落到无业务归属的 utils。
- 引入能力优先复用成熟方案和现有项目模式；新增接口要避免重复，优先复用已有 endpoint。

## 之前已经完成的主线

### 已提交闭环

- `3503828 backend/frontend: add cmdb asset parity contracts`
  - 完成 CMDB / 资产同源字段、GORM 持久化、host ops 映射阻断和相邻契约。
  - 真实远程终端、上传 writer、命令执行器、部署执行器仍是后续切片。

- `6fbb46a backend: harden cmdb topology binding contracts`
  - `GET /api/v1/cmdb/instances/:id/topology`、`GET/POST /api/v1/cmdb/monitor-bindings/*path` 先返回结构化 `BLOCKED_BY_CONTRACT`，包含 expected schema、field matrix、source evidence 和 missing contracts。
  - 不返回假 `nodes/edges/code:0`，不回显敏感 request marker，不伪装 monitor binding 写入成功。

- 其它已完成分支可在 `.codex/codex-task-board.md` 查：
  - 119B 系列：监控 contract matrix 与 Nightingale 相关 gap 拆分。
  - 124：SkyWalking OAP adapter blocked contracts。
  - 125：SigNoZ logs adapter blocked contracts。
  - 129：业务拨测状态页。

### 本窗口完成但尚未提交的 CMDB 闭环

这些是当前本地未提交切片，不能当作 Git 已落库事实：

- `FX-NIGHT-126D-CMDB-RELATION-GRAPH-BACKEND-CONTRACT`
  - 新增 GORM-aware 关系读取存储 `api/internal/store/cmdb_relations.go`。
  - `GET /api/v1/cmdb/instances/:id/topology` 只有在真实 `CmdbInstanceRelation`、相关实例和关系类型都存在时才返回 `code:0/status=ready/nodes/edges`。
  - 空关系、缺实例、悬挂关系、缺关系类型继续 404/409 fail-close。
  - 节点不输出 raw `CmdbInstance.Data`，测试覆盖 password/token/dsn 不回显。

- `FX-NIGHT-126E-CMDB-MONITOR-BINDING-AUDIT-LOG-CONTRACT`
  - 新增 `CmdbMonitorBinding` model、GORM AutoMigrate、store 保存 / 查询能力。
  - 复用现有 `/api/v1/cmdb/monitor-bindings/*path`，不新增重复接口。
  - POST 只有真实 CMDB instance 且含 `hostid/templateid/cmdb_attr_id/server_attr_id` 才持久化。
  - 成功写入同步写 `monitor_audit_logs`，`scope=cmdb`、`resource_type=cmdb_monitor_binding`、`action=cmdb.monitor_binding.save`。
  - 响应带 `log_query`，可通过 FindX `findx_audit` 查询。
  - 敏感字段递归脱敏，fake state 不回显。

- `FX-NIGHT-126F-CMDB-MONITOR-BINDING-FRONTEND-CONSUMER`
  - 前端 `cmdbApi.relations.monitorBindings` 消费同一 monitor binding contract。
  - `监控绑定` 动作只在后端真实 ready 且有真实 bindings 时显示绑定列表。
  - 409 或空绑定显示契约审计、missing contracts、expected schema、field matrix 和 `findx_audit` 查询。

- `FX-NIGHT-126G-CMDB-MONITOR-BINDING-DRAWER-PIXEL-PARITY`
  - 根据 `D:\项目迁移文件\测试cmdb图片\图片\*.png` 中成熟 CMDB 抽屉样式，把 monitor binding 面板改为右侧抽屉：
    - 灰色遮罩。
    - 右侧白色 panel。
    - 顶部标题 / 关闭按钮。
    - 蓝色竖线分组标题。
    - `模型对接 / 关联属性 / 日志审计` 分段。
  - 桌面 1440px 抽屉右对齐宽 1056；移动 390px 全宽。

- `FX-NIGHT-126H-CMDB-RELATION-ACTION-DRAWER-AUDIT`
  - 关系拓扑二级动作 `展开 / 详情 / 关系 / 拓扑` 不再只是禁用按钮。
  - 点击动作打开同款右抽屉审计，显示：
    - 动作目标：instance_id、node_id、object_id、relation_id。
    - 业务上下文阻断原因。
    - 缺失契约：`cmdb_relation_action_store`、`cmdb_relation_action_delivery_receipt_contract` 等。
    - FindX audit log query：`resource_type=cmdb_relation_action`。
  - 仍然不伪装真实详情 / 关系 / 拓扑执行。

## 当前正在做什么

当前任务是“换任务窗口前交接”。本窗口刚刚：

- 补齐 126F、126G、126H 的本地闭环记录。
- 将当前项目状态写入本文档。
- 未执行 stage / commit / push。
- 未变更 `api/internal/handler/data/memory-store.json`。

当前主要未提交写集：

- `api/internal/model/cmdb.go`
- `api/internal/store/gorm.go`
- `api/internal/store/cmdb_relations.go`
- `api/internal/store/cmdb_monitor_bindings.go`
- `api/internal/handler/cmdb_topology_graph.go`
- `api/internal/handler/cmdb_topology_graph_test.go`
- `api/internal/handler/cmdb_monitor_bindings.go`
- `api/internal/handler/cmdb_monitor_bindings_test.go`
- `api/internal/handler/cmdb_topology_bindings_contract.go`
- `api/internal/handler/cmdb_compatible_test.go`
- `web/src/react-shell/api/cmdb.js`
- `web/src/react-shell/cmdb/RelationTopologySection.jsx`
- `web/src/react-shell/cmdb/RelationTopologySection.css`
- `.codex/codex-task-board.md`
- `.codex/operations-log.md`
- `docs/aiops/findx_project_handoff_2026-05-14.md`

注意：`web/src/react-shell/cmdb/RelationTopologySection.jsx` 与 `RelationTopologySection.css` 当前是未跟踪新文件；不要误以为已经提交。

## 最近验证证据

最近本地验证已经跑过：

- 后端 126D/126E：
  - `go test -count=1 ./internal/handler -run "Cmdb.*Monitor|Cmdb.*Topology|Cmdb.*Compatible|MonitorBindings|ResourceAndHostAsset|HostAssets"` PASS。
  - `go test -count=1 ./...` PASS。
- 前端 126F/126G/126H：
  - `cd web && npm.cmd run build` PASS。
  - Vite 仍有既有 CJS Node API、Monaco dynamic/static import、chunk size warning，不是本轮新错误。
- 静态扫描：
  - 前端写集 U+FFFD 0。
  - fake/embed/brand scan 0。
  - 敏感扫描只命中既有 API client 本地 auth header：`Authorization: Bearer ${token}`。
- Browser / Playwright 本地：
  - 390px monitor binding drawer：`bodyScrollWidth=390`，drawer/backdrop 可见，无 fake ready、无 iframe、无 visible sensitive/external brand/mojibake。
  - 1440px monitor binding drawer：panel `x=384,width=1056,height=900`，右对齐，无横向溢出。
  - 390px 关系动作 `详情` drawer：动作目标、业务上下文、缺失契约、findx_audit query 可见。
  - 1440px 关系动作 `关系` drawer：右对齐，`bodyScrollWidth=1440`。
  - console error 只有预期的 contract probe 409。

如果新窗口要继续声明完成，必须重新运行需要的验证，不要只复用本文档结果。

## 未来要做什么

建议下一窗口优先按这个顺序推进：

1. 补齐 126H operations-log 后续验证记录，并核对当前 `.codex/codex-task-board.md` / `.codex/operations-log.md` 是否含 NUL 历史控制字符。它们已有历史 binary 判断风险，不要误判为源码乱码。
2. 对当前未提交 CMDB 写集做最终 pathspec gate：
   - 禁止 stage `.codex/**`、`.claude/**`、`.playwright-mcp/**`、`.test-evidence/**`、runtime data、`api/internal/handler/data/memory-store.json`。
   - 如果需要提交，只 stage 明确候选源码和正式 docs。
3. 如要把 126D-126H 合并成可提交切片，先重新跑：
   - `cd api && go test -count=1 ./internal/handler -run "Cmdb.*Monitor|Cmdb.*Topology|Cmdb.*Compatible|MonitorBindings|ResourceAndHostAsset|HostAssets"`
   - `cd api && go test -count=1 ./...`
   - `cd web && npm.cmd run build`
   - 前端 U+FFFD / iframe / brand / fake-state / sensitive 扫描。
   - Browser 390px 与 1440px topology、monitor binding、relation action drawer。
4. 后端后续切片：
   - 真实 monitor template runtime 校验。
   - monitor binding rollback / delivery receipt / effect receipt。
   - CMDB relation action store。
   - 递归多层 relation path 和拓扑 layout 坐标。
   - topology 二级 / 三级动作真实落库与审计回执。
5. 前端后续切片：
   - 用截图建立自动 pixel-diff 工具，而不是只靠人工视觉检查。
   - 进一步对齐成熟 CMDB 资源列表、机房视图、审计记录、发现管理的抽屉和表格布局。
   - 移动端每次都测 `bodyScrollWidth === innerWidth`。
6. 远端 / 部署后续切片：
   - WSL runtime sync。
   - remote Ubuntu runtime sync。
   - pathspec commit gate。
   - 只在用户明确要求时 commit / push。

## 新窗口提示词

把下面整段贴给新窗口：

```text
你在 D:\项目迁移文件\ai-workbench 继续 FindX AIOps 项目。请使用简体中文。用户明确要求：禁止降级到 5.4，无论主 agent、子 agent 或其它角色；不要 stage/commit/push，除非用户明确要求；不要重置或回滚历史脏树；api/internal/handler/data/memory-store.json 是 runtime/sensitive candidate，不能 stage、不能作为证据。

先读 docs/aiops/findx_project_handoff_2026-05-14.md，再读 .codex/codex-task-board.md 和 .codex/operations-log.md 的尾部。当前重点是 CMDB 126D-126H 本地未提交闭环：
- 126D：GORM-backed CMDB relation graph 后端契约，真实 CmdbInstanceRelation + related instances + relation types 才返回 ready nodes/edges，坏关系 fail-close。
- 126E：CmdbMonitorBinding model/store/handler，复用 /api/v1/cmdb/monitor-bindings/*path，真实绑定写入 FindX audit log，敏感字段脱敏，空绑定继续 BLOCKED_BY_CONTRACT。
- 126F：前端消费同一 monitor-binding contract，不新增重复接口。
- 126G：monitor binding 右侧抽屉像素风格对齐用户截图，桌面右对齐、移动全宽。
- 126H：关系二级动作 展开/详情/关系/拓扑 打开同款右抽屉审计，展示动作目标、业务上下文、missing contracts 和 findx_audit query；仍然不伪造真实执行。

重要证据源：
- D:\测试\LWOPS_安全测试资料_2026-05-10\reverse-poc\public\captures\cmdb-instance-default-topology.json
- D:\测试\LWOPS_安全测试资料_2026-05-10\reverse-poc\public\captures\cmdb-instance-default-topology-request.txt
- D:\测试\LWOPS_安全测试资料_2026-05-10\reverse-poc\public\captures\cmdb-model-relations-operatingsystem1.json
- D:\测试\LWOPS_安全测试资料_2026-05-10\reverse-poc\public\captures\cmdb-task-log-status0-55274.json
- D:\项目迁移文件\测试cmdb图片\图片\*.png

当前主要写集包括：
api/internal/model/cmdb.go
api/internal/store/gorm.go
api/internal/store/cmdb_relations.go
api/internal/store/cmdb_monitor_bindings.go
api/internal/handler/cmdb_topology_graph.go
api/internal/handler/cmdb_topology_graph_test.go
api/internal/handler/cmdb_monitor_bindings.go
api/internal/handler/cmdb_monitor_bindings_test.go
api/internal/handler/cmdb_topology_bindings_contract.go
api/internal/handler/cmdb_compatible_test.go
web/src/react-shell/api/cmdb.js
web/src/react-shell/cmdb/RelationTopologySection.jsx
web/src/react-shell/cmdb/RelationTopologySection.css
docs/aiops/findx_project_handoff_2026-05-14.md

接手后请先完成：
1. 检查被中断后的 .codex/operations-log.md 是否已补齐 126H 记录；如未补齐，补齐。
2. 重新运行本地验证，不要复用旧结果直接声称 PASS：
   cd api && go test -count=1 ./internal/handler -run "Cmdb.*Monitor|Cmdb.*Topology|Cmdb.*Compatible|MonitorBindings|ResourceAndHostAsset|HostAssets"
   cd api && go test -count=1 ./...
   cd web && npm.cmd run build
   前端 U+FFFD / iframe / external brand / fake-state / sensitive 扫描
   Browser/Playwright 390px 和 1440px topology、monitor binding drawer、relation action drawer
3. 下一目标优先选真实后端缺口：monitor template runtime 校验、binding rollback/delivery/effect receipt、relation action store、递归 relation path/layout 坐标。blocked 不是成功。

禁止：
- iframe/WebView/object/embed。
- 空 nodes/edges 或空 bindings 当成功。
- 假 queued/running/applied/installed/data_arrived/service_registered/rolled_back/uninstalled。
- 用户侧暴露外部成熟源码品牌、敏感 marker、真实密钥、Cookie、连接串。
- 新增重复 monitor binding / relation action endpoint，优先复用现有接口和 Service 边界。
```
