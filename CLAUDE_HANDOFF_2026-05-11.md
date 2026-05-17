# FindX 闭环交接给 Claude（2026-05-11 18:30 UTC+8）

本文件用于把当前 `D:\ai-workbench` 的 FindX Goal 闭环开发临时交接给 Claude。接手方必须以当前代码、Git 工作树、构建、运行态和浏览器事实为准，不依赖任何上一轮对话记忆。

## 1. Claude 接手第一步

请先读取并遵守以下文件：

- `D:\ai-workbench\AGENTS.md`
- `D:\ai-workbench\README.md`
- `D:\ai-workbench\docs\aiops\README.md`
- `D:\ai-workbench\docs\aiops\findx_full_stack_observability_long_term_plan.md`
- `D:\ai-workbench\docs\aiops\findx_react_only_frontend_long_term_plan.md`
- `D:\ai-workbench\.codex\codex-task-board.md`
- `D:\ai-workbench\.codex\operations-log.md`

当前协作记录根是 `.codex/`。不要继续把 `.claude/` 作为当前任务记录根；旧 `.claude/` 只能作为历史参考。

成熟源码根固定为：

- `D:\项目迁移文件\平台源码`

远端 Ubuntu：

- SSH alias：`findx-ubuntu`
- 源码目录：`/opt/ai-workbench`
- runtime：`/opt/ai-workbench-runtime`
- 服务：`ai-workbench-api.service`
- Web 入口：`http://10.10.160.202:3000`

WSL：

- 源码目录：`/opt/ai-workbench`
- runtime：`/opt/ai-workbench-runtime`
- 服务：`ai-workbench.service`

## 2. 当前状态

当前主线仍是 `/goal`：持续推进 FindX 全平台初步闭环，不因单个切片完成而停止。

当前阶段仍在 P0 工作树治理和后端可提交闭包收敛。仓库历史脏区很多，不能整体提交，必须按显式 pathspec 分批收敛。

最近已经完成并提交的后端闭包：

- `7d6c5fe backend: split prometheus handlers`
- `b0d5cf4 backend: split workflow bridge registry`
- `bd32487 backend: split topology persistence handlers`
- `cb39c6f backend: add route dependency closure`
- `9437f70 backend: add agent lifecycle handler closure`
- `ad4d3d7 backend: add agent installer gate closure`
- `1923c1b backend: add monitoring handler closure`
- `419fec4 backend: add notification handler closure`

Prometheus 闭包刚完成：

- 提交：`7d6c5fe backend: split prometheus handlers`
- 文件：`api/internal/handler/prometheus.go`、`prometheus_context.go`、`prometheus_discovery.go`、`prometheus_display.go`、`prometheus_handlers.go`
- 验证：
  - cached diff check PASS
  - forbidden path scan PASS
  - Go 文件 <=400 行、函数 <=50 行 PASS
  - mojibake/private-key scan PASS
  - Windows focused/full Go PASS
  - clean staged-only focused/full Go PASS
  - WSL focused/full Go/build/runtime health PASS
  - remote Ubuntu focused/full Go/build/runtime health PASS
  - Browser NOT_RUN：后端同包机械拆分，无 UI/route 行为变化

## 3. 当前工作树高风险边界

接手后先运行：

```powershell
cd D:\ai-workbench
git status --short
git diff --stat
```

当前仍有大量未提交脏区。不要 reset、checkout、revert。

永久禁止 stage/commit：

- `api/internal/handler/data/memory-store.json`
- `.codex/**`
- `.claude/**`
- `.playwright-mcp/**`
- `.test-evidence/**`
- `api/data/**`
- `logs/**`
- `web/dist/**`
- `web/node_modules/**`
- `*.pem`
- `*.key`
- `*.log`
- `*.exe`
- runtime data、真实 token、cookie、DSN、session、private key、包制品

禁止命令：

- `git reset --hard`
- `git checkout --`
- `git revert`
- `git add .`
- `git add -A`

每个闭包只能显式 pathspec stage 自己的允许写集。

## 4. 下一步应该干什么

下一步优先处理 `remote` 后端闭包，但不能直接提交。

推荐任务 ID：

- `FX-NIGHT-119-REMOTE-EXEC-CLOSURE`

允许写集：

- `api/internal/handler/remote.go`
- `api/internal/handler/remote_exec.go`
- `api/internal/handler/remote_install.go`

禁止写集：

- `api/internal/handler/data/memory-store.json`
- 前端、docs、`.codex/**`、`.claude/**`
- runtime data、凭据、包制品
- `catpaw*.go`
- `whitebox*.go`
- `aiops_inspection_metrics*.go`
- `metrics_adapt.go`

为什么先做 remote：

- `remote.go` 当前已经把 `execWMI`、`execSSH`、`execWinRM`、`build*Install*` 等函数迁出到 `remote_exec.go` 和 `remote_install.go`。
- `remote_exec.go` 同时被 `remote.go`、`credential.go`、`diagnose.go` 依赖，属于更底层的远程执行闭包。
- QA 子代理 `Fermat 019e1691-f822-7ca2-b441-9e4b58475c8d` 结论：`remote` 可独立闭包，但当前是 `RISK`，需要先修乱码和假成功风险说明/门禁。

当前已知 remote 阻断：

- `api/internal/handler/remote.go` 当前函数扫描发现 `RemoteExec` 约 160 行，必须拆分到 <=50 行。
- `api/internal/handler/remote_install.go` 的 `buildInstallScript` 约 51 行，必须拆分到 <=50 行。
- `remote.go`、`remote_exec.go`、`remote_install.go` 存在历史中文乱码用户可见错误提示，提交前必须修成可读中文或英文。
- WMI 路径通过 `Win32_Process.Create` 返回 “process started”，不能证明远端命令完成、安装完成、心跳或数据到达。不能把 WMI 启动进程当作真实安装成功；如无法补真实 receipt，必须在输出/记录中标为 blocked/risk，不得 fake success。
- `ssh.InsecureIgnoreHostKey()`、WinRM Basic/AllowUnencrypted、password 注入本地 PowerShell 脚本文本均为安全风险，提交说明和任务板要记录边界；不得输出真实密码。

remote 闭包建议步骤：

1. 读取 `remote.go`、`remote_exec.go`、`remote_install.go`，先理解结构。
2. 搜索 `execWMI`、`execSSH`、`execWinRM`、`InstallCatpaw`、`GenerateInstallCmd` 的调用方。
3. 修 `RemoteExec` 超长函数：抽出参数绑定/校验、执行分发、响应处理 helper。
4. 修 `buildInstallScript` 超长函数：把 TOML 或脚本段拆成 helper，但不要新增过度抽象。
5. 修 remote 组中的乱码错误文案。
6. 不改变 API 路径和字段契约，除非明确记录 `API_CONTRACT_CHANGE`。
7. staged-only clean worktree 验证通过后才能提交。

## 5. remote 后应该干什么

remote 闭包完成后，处理 `catpaw` 后端闭包。

推荐任务 ID：

- `FX-NIGHT-120-CATPAW-REPORT-CLOSURE`

允许写集：

- `api/internal/handler/catpaw.go`
- `api/internal/handler/catpaw_report.go`

禁止写集：

- `api/internal/handler/catpaw_chat.go`（当前没有脏改，不要混入）
- `whitebox*.go`
- `aiops_inspection_metrics*.go`
- `metrics_adapt.go`
- 前端、docs、runtime data、凭据

当前已知 catpaw 风险：

- `catpaw_report.go` 约 385 行，接近 Go 文件 400 行上限。
- `summarizeWindowsCatpawReport` 明显超过 50 行，需要拆分。
- `pluginPayload` 通过字符串定位 markdown fenced JSON，解析脆弱，后续需要更稳健的解析或明确失败态。
- `isMojibake` 对单个问号较敏感，可能误伤正常 Windows 事件消息。
- raw report 可能含主机、进程、端口、路径信息，提交前要审计脱敏路径。

catpaw 后再处理：

- `FX-NIGHT-121-INSPECTION-WHITEBOX-METRICS-CLOSURE`
  - `api/internal/handler/aiops_inspection_metrics.go`
  - `api/internal/handler/aiops_inspection_metrics_fallback.go`
  - `api/internal/handler/metrics_adapt.go`
  - `api/internal/handler/whitebox_test.go`
  - `api/internal/handler/whitebox_aiops_audience_test.go`
  - `api/internal/handler/whitebox_aiops_v2_test.go`
  - 注意：该组此前被 QA 标记有大量 mojibake 风险，不要先提交。

后端最小闭包都干净后，再回到：

- React-only 前端架构门禁
- Agent 管理中心抽屉
- Categraf 插件远程配置下发
- data-arrival validator matrix
- Playwright 完整浏览器矩阵
- README 里程碑按功能清单持续维护
- GitHub push 仅在用户明确要求时执行

## 6. 每个切片必须怎么验证

Windows：

```powershell
cd D:\ai-workbench\api
go test -count=1 ./internal/handler -run '<当前模块>'
go test -count=1 ./...
```

clean staged-only：

- 从 clean HEAD 创建临时 worktree。
- 应用 `git diff --cached --binary -- <pathspec>`。
- 在临时 worktree 运行 focused 和 full Go test。
- 用完删除临时 worktree 和 patch 文件。

WSL：

- 只同步 staged pathspec 到 `/opt/ai-workbench`。
- 不同步 data、node_modules、dist、exe、api-linux、runtime data。

```bash
cd /opt/ai-workbench/api
go test -count=1 ./internal/handler -run '<当前模块>'
go test -count=1 ./...
go build -o api-linux .
install -m 0755 api-linux /opt/ai-workbench-runtime/api/ai-workbench-api
systemctl restart ai-workbench.service
curl -fsS http://127.0.0.1:8080/api/v1/health/storage
curl -fsS http://127.0.0.1:3000/api/v1/health/storage
```

远端 Ubuntu：

```powershell
ssh findx-ubuntu "cd /opt/ai-workbench/api && go test -count=1 ./... && go build -o api-linux ."
```

然后安装 runtime binary、重启 `ai-workbench-api.service`、检查 8080 和 3000 storage health。

UI 或 route 行为变化时必须 Playwright：

- 真实登录远端 `http://10.10.160.202:3000`
- 导航点击
- 主流程
- 异常路径
- 权限/过期路径
- BLOCKED_BY_CONTRACT 展示
- 390px 窄屏
- console errors 为 0，或仅保留明确记录的预期 401/409

后端同包机械拆分且无 route/UI 行为变化时，可以 Browser `NOT_RUN`，但必须明确记录原因，不能写 PASS。

## 7. 每次提交前固定扫描

```powershell
git diff --cached --name-status
git diff --cached --check
```

禁止路径扫描必须 PASS：

```powershell
$paths = git diff --cached --name-only
$bad = $paths | Where-Object {
  $_ -match '(^|/)(\.codex|\.claude|web|docs|api/data|logs)(/|$)|memory-store\.json|\.pem$|\.key$|\.log$|\.exe$|web/dist|node_modules'
}
if ($bad) { $bad; exit 1 } else { 'cached forbidden path scan PASS' }
```

Go 文件和函数长度必须 PASS：

- Go 文件 <=400 行
- Go 函数 <=50 行

敏感/乱码扫描必须 PASS：

- `�`
- `Ã`
- `Â`
- `â€`
- 私钥头
- Bearer 长 token
- AKIA 云密钥形态
- 用户可见历史 mojibake

## 8. Claude 可直接使用的接手提示词

请把下面整段作为 Claude 的接手提示词：

```text
你现在接手 D:\ai-workbench 的 FindX Goal 闭环开发。请先读取 D:\ai-workbench\CLAUDE_HANDOFF_2026-05-11.md、AGENTS.md、README.md、docs\aiops\README.md、docs\aiops\findx_full_stack_observability_long_term_plan.md、docs\aiops\findx_react_only_frontend_long_term_plan.md、.codex\codex-task-board.md、.codex\operations-log.md。

所有对话和文档使用简体中文。当前协作记录根是 .codex/，不要把 .claude/ 作为当前写入根。成熟源码根是 D:\项目迁移文件\平台源码。远端 Ubuntu 使用 ssh alias findx-ubuntu，源码在 /opt/ai-workbench，runtime 在 /opt/ai-workbench-runtime，服务 ai-workbench-api.service，Web 入口 http://10.10.160.202:3000。

当前主线是 /goal：持续推进 FindX 全平台初步闭环，不要完成一个小任务就停止。现在处于 P0 工作树治理和后端可提交闭包收敛。最近已提交 7d6c5fe backend: split prometheus handlers。下一步优先做 FX-NIGHT-119-REMOTE-EXEC-CLOSURE。

下一步允许写集仅限：
- api/internal/handler/remote.go
- api/internal/handler/remote_exec.go
- api/internal/handler/remote_install.go

禁止写集：
- api/internal/handler/data/memory-store.json
- .codex/**
- .claude/**
- .playwright-mcp/**
- .test-evidence/**
- api/data/**
- logs/**
- web/**
- docs/**
- runtime data
- 真实 token/cookie/DSN/session/private key
- catpaw*.go
- whitebox*.go
- aiops_inspection_metrics*.go
- metrics_adapt.go

禁止 git reset --hard、git checkout --、git revert、git add .、git add -A。只能按显式 pathspec stage。

remote 当前阻断：RemoteExec 超过 50 行；remote_install.go 的 buildInstallScript 约 51 行；remote 组存在用户可见历史乱码；WMI process started 不能冒充安装完成、心跳或数据到达；ssh.InsecureIgnoreHostKey、WinRM Basic/AllowUnencrypted、password 注入本地 PowerShell 脚本文本都必须记录安全边界，不得输出真实凭据。

请先只读复核 git status --short、git diff --stat 和 remote 组三个文件，再修 remote 闭包。修复后必须跑 cached diff check、禁止路径扫描、Go 文件/函数长度扫描、乱码/敏感扫描、Windows focused/full Go、clean staged-only Go、WSL focused/full Go/build/runtime health、remote Ubuntu focused/full Go/build/runtime health。若 route/UI 行为没变，Browser 记录 NOT_RUN 且说明是后端同包拆分；若行为有变，必须 Playwright 真实登录、导航、异常、权限、390px、console/network 检查。

完成 remote 并提交后不要停，继续 FX-NIGHT-120-CATPAW-REPORT-CLOSURE。catpaw 允许写集仅 api/internal/handler/catpaw.go 和 catpaw_report.go，先拆 summarizeWindowsCatpawReport 超长函数并处理解析/乱码风险。后续再处理 inspection/whitebox/metrics_adapt 组，不要混提交。

每次闭环都更新 .codex/codex-task-board.md 和 .codex/operations-log.md，但不要 stage .codex。GitHub push 只有用户明确要求时才执行。
```

## 9. 仍未完成的主线

平台底座和全平台闭环还远未完成，当前只是后端工作树治理中的一段。剩余长期主线包括：

- P0 工作树和运行态治理
- P1 React-only Shell 与导航归属
- P2 基础监控 Nightingale 同源迁移
- P3 SkyWalking 链路监控迁移
- P4 SigNoZ 日志中心迁移
- P5 CMDB / 资产 / Agent 管理中心
- P6 FindX Agent 生命周期：本机安装、远程下发、远程安装、配置下发、插件下发、升级、回滚、卸载、心跳、数据到达、审计、Evidence Chain
- P7 AI SRE / Evidence Chain / 知识库
- P8 治理、测试、文档、发布

`BLOCKED_BY_CONTRACT`、`DONE_WITH_RISK`、`NOT_RUN`、`RISK` 都不是最终完成，只是下一轮任务输入。

