# FindX AIOps / CMDB 对话切换交接 2026-05-16

## 目标

继续在 `D:\项目迁移文件\ai-workbench` 闭环 FindX AIOps / CMDB / Agent 项目。远端基线为 `findx-ubuntu:/opt/ai-workbench`，运行态验收以 `http://10.10.160.202:3000` 为准。

连续闭环规则：

- 使用简体中文。
- 禁止降级到 5.4；主 agent、子 agent、QA、文档、其它角色都只能使用 gpt-5.5 或同级更高模型。
- 不要 stage/commit/push，除非用户明确要求。
- 不要 reset/revert/checkout 回滚历史脏树。
- 不要读取、引用、stage 或把 `api/internal/handler/data/memory-store.json` 当完成证据。
- 不要整仓同步远端脏树；远端只同步本轮写集。
- blocked 不是成功；`BLOCKED_BY_CONTRACT` 只能作为缺真实执行器、回执、数据源时的 fail-close 状态。
- 禁止 iframe/WebView/object/embed。
- 禁止空 nodes/edges、空 bindings、空 receipts、空 actions 当成功。
- 禁止假状态：`queued/running/applied/installed/data_arrived/service_registered/rolled_back/uninstalled/delivered/effective/succeeded/success/imported`。
- 用户侧不得暴露外部成熟源码品牌、敏感 marker、真实密钥、Cookie、连接串。

## 当前正在处理的事情

当前主线是 CMDB host plugin dispatch 闭环，最近完成到 `FX-NIGHT-142` 的主体实现与远端部署，正在收尾 Browser 验收和本地记录。

`FX-NIGHT-142` 内容：

- 复用现有 `GET /api/v1/findx-agents/data-arrival/evidence`，不新增 endpoint。
- 当 query 带 `rollout_id` / `config_rollout_id` / `rollout_ref` / `request_ref` 时进入 CMDB dispatch data-arrival runtime read gate。
- 新 contract：`cmdb.agent.plugin.dispatch.data_arrival.read.v1`。
- 只有真实 receiver-backed evidence 才允许返回只读 body；没有真实 receiver evidence 时返回 409 `BLOCKED_BY_CONTRACT`，不能把空 evidence 当成功。
- 即使 evidence 存在，也仍返回 blocked 语义，不伪装插件安装、投递、生效、数据到达完成。

当前中断点：

- 已完成本地 focused handler、宽 handler、本地全量 Go、本地 Web build。
- 已同步 142 后端文件到远端，并补同步了前端 `HostProbePluginDrawer.jsx` 的用户侧品牌脱敏修复。
- 已完成远端 focused handler、远端全量 Go、远端 API build、远端 Web build、runtime 部署。
- runtime health 已通过：8080 和 3000 均返回 MySQL/Redis true。
- Browser 1440px 主机探针/插件抽屉已通过。
- Browser 已 resize 到 390x844，但 390px 抽屉最终 evaluate 还没跑完，用户要求切换对话时中断。
- `.codex/operations-log.md` 和 `.codex/codex-task-board.md` 还没有补 `FX-NIGHT-142` 记录。

## 已经改变的事情

### 后端 142 写集

- `api/internal/handler/findx_agent_lifecycle_records.go`
- `api/internal/handler/findx_agent_lifecycle_data_arrival_runtime_read.go`
- `api/internal/handler/findx_agent_lifecycle_data_arrival_runtime_read_test.go`

行为要点：

- data-arrival read gate 会先解析 rollout，并复用 140 的 dispatch receipt read gate，要求 `delivery_request_ref`、`effect_request_ref`、`rollback_request_ref` 都能解析到真实 blocked task。
- `request_ref` 如指定，必须属于该 rollout 的 delivery/effect/rollback refs，并且 task action、rollout id、receipt kind、plugin id、target/agent 匹配。
- data-arrival evidence 只接受 receiver-backed kind：`heartbeat` / `metrics` / `logs` / `tracing`，且 `status=reported`、`evidence_refs` 含 `receiver:...`。
- 没有真实 receiver evidence 时返回 409，不返回空成功体。

### 前端补丁

文件：

- `web/src/react-shell/cmdb/HostProbePluginDrawer.jsx`

改动：

- `sourceBrandFragments` 增加 `Prometheus` 和 `Grafana` 的脱敏片段。
- 目的：修复 Browser 1440px 抽屉里 `prometheus` 插件 id 以用户可见文本露出的问题。
- 不改后端真实插件 id，不改 API 契约，不影响后端 source-map / catalog。

### 已重新跑过的本地验证

本地 focused handler：

```powershell
cd D:\项目迁移文件\ai-workbench\api
go test -count=1 ./internal/handler -run "DataArrivalEvidenceCmdbDispatchRead|DataArrivalEvidenceRuntimeReadKeepsLegacyList|ConfigRollout.*RuntimeRead|FindXAgentConfigRolloutCmdb.*Dispatch"
```

结果：PASS。

本地全量 Go：

```powershell
cd D:\项目迁移文件\ai-workbench\api
go test -count=1 ./...
```

结果：PASS。

本地 Web build：

```powershell
cd D:\项目迁移文件\ai-workbench\web
npm.cmd run build
```

结果：PASS，仅既有 Vite CJS / Monaco dynamic import / chunk size warning。

写集 diff check：

```powershell
git diff --check -- web/src/react-shell/cmdb/HostProbePluginDrawer.jsx api/internal/handler/findx_agent_lifecycle_records.go api/internal/handler/findx_agent_lifecycle_data_arrival_runtime_read.go api/internal/handler/findx_agent_lifecycle_data_arrival_runtime_read_test.go
```

结果：PASS，仅既有 CRLF warning。

前端文件扫描：

```powershell
rg -n "Nightingale|夜莺|SkyWalking|SigNoZ|Categraf|Catpaw|Grafana|Prometheus|<iframe|<object|<embed|WebView|queued|running|applied|installed|data_arrived|service_registered|rolled_back|uninstalled|delivered|effective|succeeded|success|imported|\x{FFFD}" web/src/react-shell/cmdb/HostProbePluginDrawer.jsx
```

结果：无命中。

### 已重新跑过的远端验证

远端 focused handler：

```bash
cd /opt/ai-workbench/api
/usr/local/go/bin/go test -count=1 ./internal/handler -run 'DataArrivalEvidenceCmdbDispatchRead|DataArrivalEvidenceRuntimeReadKeepsLegacyList|ConfigRollout.*RuntimeRead|FindXAgentConfigRolloutCmdb.*Dispatch'
```

结果：PASS。

远端全量 Go：

```bash
cd /opt/ai-workbench/api
/usr/local/go/bin/go test -count=1 ./...
```

结果：PASS。

远端 API build：

```bash
cd /opt/ai-workbench/api
/usr/local/go/bin/go build -o api-linux .
```

结果：PASS。

远端 Web build：

```bash
cd /opt/ai-workbench/web
npm run build
```

结果：PASS，仅既有 chunk warning。

远端 runtime 部署已执行：

```bash
sudo systemctl stop ai-workbench-api.service
sudo cp /opt/ai-workbench/api/api-linux /opt/ai-workbench-runtime/api/ai-workbench-api
sudo rm -rf /opt/ai-workbench-runtime/web/dist
sudo cp -a /opt/ai-workbench/web/dist /opt/ai-workbench-runtime/web/dist
sudo systemctl start ai-workbench-api.service
```

远端 health：

```bash
curl -sS -m 5 http://127.0.0.1:8080/api/v1/health/storage
curl -sS -m 5 http://127.0.0.1:3000/api/v1/health/storage
```

结果：两端均为 `{"mysql":true,"redis":true}`。

### Browser 1440px 已通过

页面：

```text
http://10.10.160.202:3000/assets?section=hosts
```

1440x900 主机探针/插件抽屉结果：

```json
{
  "rect": {"x":480,"y":0,"width":960,"height":900,"right":1440},
  "bodyScrollWidth":1440,
  "docScrollWidth":1440,
  "clientWidth":1440,
  "innerWidth":1440,
  "iframeCount":0,
  "webview":false,
  "hasReplacement":false,
  "externalBrand":false,
  "visibleSensitive":false,
  "fakeState":false,
  "visible": {
    "collect":true,
    "apm":true,
    "diagnostic":true,
    "blocked":true,
    "credentialRef":true,
    "oldConfirmAssign":false
  }
}
```

## 尝试过但失败或需要注意的事情

- 第一次 Browser 1440px 验收发现 `externalBrand=true`，原因是抽屉可见文本包含 `prometheus`。已通过前端脱敏修复，重新部署后 1440px 已通过。
- runtime 部署后第一次 8080 health 命中冷启动窗口，返回 connection refused；随后重试通过。不要把第一次拒连记录为 PASS，但也不要当成代码失败。
- Browser 重新导航后登录态过期，跳到 `/login`；已通过正常登录流程重新进入。不要把账号口令写进代码、日志或最终证据。
- `.codex/operations-log.md` 和 `.codex/codex-task-board.md` 尾部尚未记录 142，本轮中断前还没补。
- Browser 已 resize 到 390x844，但还没有重新点击/检查移动端抽屉；下一对话需从这里继续。

## 子 agent 结果

已派发只读 explorer `Zeno`，使用 gpt-5.5，未改文件。

结论：

- 当前生产代码没有把 `writer_request_ref` 纳入 CMDB host plugin dispatch runtime read gate 的 required request_ref。
- 稳定 required set 已经是 `delivery/effect/rollback`。
- `FX-NIGHT-143` 应该做防回归测试固化，而不是新增第四个 `writer_request_ref`。
- `writer_receipt` 仍然是 operation contract 的缺口语义，不应删除；它和 runtime read gate 的 request_ref required set 不是同一层。

推荐 143 最小写集：

- `api/internal/handler/findx_agent_lifecycle_config_rollout_runtime_read_test.go`
- `api/internal/handler/findx_agent_lifecycle_data_arrival_runtime_read_test.go`

推荐测试：

- `TestFindXAgentConfigRolloutDispatchDetailDoesNotRequireWriterRequestRef`
- `TestFindXAgentConfigRolloutDispatchDetailIgnoresStaleWriterRequestRef`
- `TestFindXAgentDataArrivalEvidenceCmdbDispatchReadDoesNotRequireWriterRequestRef`

核心断言：

- 三类 `delivery/effect/rollback` request_ref 可解析时，缺失或陈旧的 `writer_request_ref` 不得导致 detail/data-arrival runtime read gate 返回 409。
- 响应不得包含 `writer_request_ref` 或 `cmdb_agent_rollout_writer_request_ref_contract`。
- data-arrival evidence 仍保持 `status=blocked` 和真实执行闭环缺口，不显示假完成。

## 下一步要做什么

### 先收尾 142

1. Browser 保持或重新打开远端：

```text
http://10.10.160.202:3000/assets?section=hosts
```

2. 在 390x844 下打开第一行 `探针/插件` 抽屉，执行 evaluate：

```js
() => {
  const panel = document.querySelector('.fx-cmdb-probe-panel');
  const rect = panel?.getBoundingClientRect();
  const bodyText = document.body.innerText || '';
  const cleaned = bodyText.replace(/BLOCKED_BY_CONTRACT/g, '');
  const brandRe = /Nightingale|夜莺|SkyWalking|SigNoZ|Categraf|Catpaw|Grafana|Prometheus/i;
  const fakeRe = /\b(queued|running|applied|installed|data_arrived|service_registered|rolled_back|uninstalled|delivered|effective|succeeded|success|imported)\b/i;
  return {
    url: location.href,
    viewport: { width: innerWidth, height: innerHeight },
    rect: rect && { x: Math.round(rect.x), y: Math.round(rect.y), width: Math.round(rect.width), height: Math.round(rect.height), right: Math.round(rect.right) },
    bodyScrollWidth: document.body.scrollWidth,
    docScrollWidth: document.documentElement.scrollWidth,
    clientWidth: document.documentElement.clientWidth,
    innerWidth,
    iframeCount: document.querySelectorAll('iframe,object,embed').length,
    webview: !!document.querySelector('webview'),
    hasReplacement: bodyText.includes('\uFFFD'),
    externalBrand: brandRe.test(bodyText),
    visibleSensitive: /password|token|cookie|dsn|private_key|mysql:\/\/|postgres:\/\//i.test(bodyText),
    fakeState: fakeRe.test(cleaned),
    visible: {
      collect: bodyText.includes('采集插件'),
      apm: bodyText.includes('APM 探针'),
      diagnostic: bodyText.includes('诊断插件'),
      blocked: bodyText.includes('BLOCKED_BY_CONTRACT'),
      credentialRef: bodyText.includes('credential_ref'),
      oldConfirmAssign: bodyText.includes('确认分配')
    }
  };
}
```

期望：

- `rect.x=0`
- `rect.width=390`
- `rect.right=390`
- `bodyScrollWidth/docScrollWidth/clientWidth/innerWidth` 均为 390
- `iframeCount=0`
- `webview=false`
- `hasReplacement=false`
- `externalBrand=false`
- `visibleSensitive=false`
- `fakeState=false`
- `采集插件/APM 探针/诊断插件/BLOCKED_BY_CONTRACT/credential_ref` 可见
- `确认分配` 不可见

3. 如 390px 通过，补 `.codex/operations-log.md` 和 `.codex/codex-task-board.md` 的 `FX-NIGHT-142` 记录。

建议 142 记录重点：

- 本地 focused/wide/full Go PASS。
- 本地 Web build PASS。
- 远端 focused/full Go PASS。
- 远端 API/Web build PASS。
- runtime 部署和 8080/3000 health PASS。
- API probe：dispatch 409、refs 存在、detail 200 blocked、data-arrival evidence 409 contract=`cmdb.agent.plugin.dispatch.data_arrival.read.v1`、forbidden hits 0。
- Browser 1440/390 hosts plugin drawer PASS。
- 残留：真实 remote executor、delivery/effect/rollback executor、data-arrival receiver ingestion、credential policy approval、真实插件安装仍未完成。

### 然后进入 143

目标：

`FX-NIGHT-143 CMDB plugin dispatch runtime receipt phase contract hardening`

最小实现：

- 只加防回归测试，生产代码可不动。
- 固化 `delivery/effect/rollback` 是 runtime request_ref read gate 的稳定集合。
- 固化 `writer_request_ref` 不参与 runtime read gate。
- 固化 data-arrival runtime read 不接受或不要求 `writer_request_ref`。

本地验证建议：

```powershell
cd D:\项目迁移文件\ai-workbench\api
go test -count=1 ./internal/handler -run "ConfigRollout.*WriterRequestRef|DataArrival.*WriterRequestRef|ConfigRollout.*RuntimeRead|DataArrivalEvidenceCmdbDispatchRead"
go test -count=1 ./internal/handler -run "Cmdb.*Monitor|Cmdb.*Topology|Cmdb.*Compatible|MonitorBindings|ResourceAndHostAsset|HostAssets|CmdbRelationAction|TestCmdbRelationAction|FindXAgent|Permission|DataArrivalEvidenceCmdbDispatchRead|DataArrivalEvidenceRuntimeReadKeepsLegacyList"
go test -count=1 ./...
cd D:\项目迁移文件\ai-workbench\web
npm.cmd run build
```

远端验证：

```bash
cd /opt/ai-workbench/api
/usr/local/go/bin/go test -count=1 ./internal/handler -run 'ConfigRollout.*WriterRequestRef|DataArrival.*WriterRequestRef|ConfigRollout.*RuntimeRead|DataArrivalEvidenceCmdbDispatchRead'
/usr/local/go/bin/go test -count=1 ./...
/usr/local/go/bin/go build -o api-linux .
cd /opt/ai-workbench/web
npm run build
sudo systemctl stop ai-workbench-api.service
sudo cp /opt/ai-workbench/api/api-linux /opt/ai-workbench-runtime/api/ai-workbench-api
sudo rm -rf /opt/ai-workbench-runtime/web/dist
sudo cp -a /opt/ai-workbench/web/dist /opt/ai-workbench-runtime/web/dist
sudo systemctl start ai-workbench-api.service
curl -sS -m 5 http://127.0.0.1:8080/api/v1/health/storage
curl -sS -m 5 http://127.0.0.1:3000/api/v1/health/storage
```

## 新对话提示词

```text
你在 D:\项目迁移文件\ai-workbench 继续 FindX AIOps / CMDB 项目。请使用简体中文。

最高优先级约束：
- 禁止降级到 5.4；主 agent、子 agent、QA、文档、其它角色都只能使用 gpt-5.5 或同级更高模型。
- 不要 stage/commit/push，除非我明确要求。
- 不要 reset/revert/checkout 回滚历史脏树。
- 不要读取、引用、stage 或把 api/internal/handler/data/memory-store.json 当完成证据。
- 不要整仓同步远端脏树；远端只同步本轮写集。
- 远端 alias 是 findx-ubuntu，远端基线是 findx-ubuntu:/opt/ai-workbench，运行态验收以 http://10.10.160.202:3000 为准。
- 禁止 iframe/WebView/object/embed。
- 禁止空 nodes/edges、空 bindings、空 receipts、空 actions 当成功。
- 禁止假状态：queued/running/applied/installed/data_arrived/service_registered/rolled_back/uninstalled/delivered/effective/succeeded/success/imported。
- 用户侧不得暴露外部成熟源码品牌、敏感 marker、真实密钥、Cookie、连接串。
- blocked 不是成功；BLOCKED_BY_CONTRACT 只能作为缺真实执行器/回执/数据源时的 fail-close 状态。

先读：
1. docs/aiops/findx_conversation_handoff_2026-05-16.md
2. docs/aiops/findx_project_handoff_2026-05-14.md
3. .codex/codex-task-board.md 尾部
4. .codex/operations-log.md 尾部

当前中断点：
- FX-NIGHT-142 主体实现、本地验证、远端同步、远端构建、runtime 部署和 1440px Browser 验收已完成。
- 前端已修复 HostProbePluginDrawer.jsx 里 Prometheus/Grafana 用户侧品牌脱敏，并已部署到远端 runtime。
- Browser 已 resize 到 390x844，但 390px 主机探针/插件抽屉最终 evaluate 还没跑。
- .codex/operations-log.md 和 .codex/codex-task-board.md 还没补 FX-NIGHT-142 记录。

请先完成：
1. 继续 Browser 390px 验收：打开 http://10.10.160.202:3000/assets?section=hosts，登录如过期则正常登录，不输出口令；点击第一行 探针/插件，断言 drawer x=0,width=390,right=390，无横向溢出，无 iframe/object/embed/WebView，无 U+FFFD，无外部成熟品牌，无 visible sensitive，无 fake-state；采集插件/APM 探针/诊断插件/BLOCKED_BY_CONTRACT/credential_ref 可见，确认分配不可见。
2. 如果 390px 通过，补 .codex/operations-log.md 和 .codex/codex-task-board.md 的 FX-NIGHT-142 远端闭环记录。
3. 立刻进入 FX-NIGHT-143，不要停：只加防回归测试，固化 writer_request_ref 不属于 CMDB host plugin dispatch runtime read gate；required request_ref 只能是 delivery/effect/rollback。推荐测试：
   - TestFindXAgentConfigRolloutDispatchDetailDoesNotRequireWriterRequestRef
   - TestFindXAgentConfigRolloutDispatchDetailIgnoresStaleWriterRequestRef
   - TestFindXAgentDataArrivalEvidenceCmdbDispatchReadDoesNotRequireWriterRequestRef
4. 143 本地重新跑 focused/wide/full Go 和 web build，远端只同步 143 写集，远端重新跑 focused/full Go、API/Web build、runtime 部署、health、API probe、Browser gate。

连续闭环规则：不要完成一个 FX-NIGHT 节点就停。每完成一个节点后，立刻选择下一个真实后端或前端缺口继续，直到所有 CMDB/Agent/AIOps 残留项完成或遇到必须人工决策的阻塞。
```
