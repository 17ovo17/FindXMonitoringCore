# P0 快速收尾 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 批量提交所有剩余后端脏文件，结束 P0 工作树治理阶段，让仓库进入可正常功能开发状态。

**Architecture:** 当前所有脏文件已编译通过且测试 PASS，只需扫描敏感信息后按 pathspec 批量提交。不再逐文件拆函数。技术债（函数行数超标等）记录但不阻塞。

**Tech Stack:** Go 1.21+, Gin, MySQL, 远端 Ubuntu (findx-ubuntu)

---

### Task 1: 批量提交剩余后端脏文件

**Files:**
- Modified: `api/internal/handler/aiops_inspection_metrics.go`
- Modified: `api/internal/handler/dashboard_monitor.go`
- Modified: `api/internal/handler/dashboard_monitor_test.go`
- Modified: `api/internal/handler/metrics_adapt.go`
- Modified: `api/internal/handler/whitebox_test.go`
- Modified: `api/internal/store/init_test.go`
- New: `api/internal/handler/aiops_inspection_metrics_fallback.go`
- New: `api/internal/handler/whitebox_aiops_audience_test.go`
- New: `api/internal/handler/whitebox_aiops_v2_test.go`

**禁止提交（排除）：**
- `api/internal/handler/data/memory-store.json`

- [ ] **Step 1: 远端 Ubuntu 同步并验证**

```bash
scp api/internal/handler/aiops_inspection_metrics.go \
    api/internal/handler/aiops_inspection_metrics_fallback.go \
    api/internal/handler/dashboard_monitor.go \
    api/internal/handler/dashboard_monitor_test.go \
    api/internal/handler/metrics_adapt.go \
    api/internal/handler/whitebox_test.go \
    api/internal/handler/whitebox_aiops_audience_test.go \
    api/internal/handler/whitebox_aiops_v2_test.go \
    findx-ubuntu:/opt/ai-workbench/api/internal/handler/

scp api/internal/store/init_test.go \
    findx-ubuntu:/opt/ai-workbench/api/internal/store/

ssh findx-ubuntu "cd /opt/ai-workbench/api && go test -count=1 ./... && go build -o api-linux ."
```

Expected: 全部 PASS，build 成功

- [ ] **Step 2: Stage 允许文件**

```bash
git add -- \
  api/internal/handler/aiops_inspection_metrics.go \
  api/internal/handler/aiops_inspection_metrics_fallback.go \
  api/internal/handler/dashboard_monitor.go \
  api/internal/handler/dashboard_monitor_test.go \
  api/internal/handler/metrics_adapt.go \
  api/internal/handler/whitebox_test.go \
  api/internal/handler/whitebox_aiops_audience_test.go \
  api/internal/handler/whitebox_aiops_v2_test.go \
  api/internal/store/init_test.go
```

- [ ] **Step 3: 禁止路径扫描**

```bash
git diff --cached --name-only | grep -E '(\.codex|\.claude|web|docs|api/data|logs|memory-store\.json|\.pem$|\.key$|\.log$|\.exe$)' && echo "FAIL" || echo "PASS"
```

Expected: PASS

- [ ] **Step 4: 提交**

```bash
git commit -m "backend: batch commit remaining handler and store changes

- aiops_inspection_metrics: split with fallback handler
- dashboard_monitor: updates and test
- metrics_adapt: adaptation layer changes
- whitebox: additional test coverage (audience, aiops_v2)
- store/init_test: test updates

Technical debt noted (not blocking):
- Some functions exceed 50-line guideline
- Potential mojibake in inspection metrics (to fix during feature dev)

No API_CONTRACT_CHANGE. No DATA_CHANGE.
Verified: Windows go test PASS, remote Ubuntu go test + build PASS."
```

- [ ] **Step 5: 远端部署并验证 health（里程碑验证）**

```bash
ssh findx-ubuntu "sudo install -m 0755 /opt/ai-workbench/api/api-linux /opt/ai-workbench-runtime/api/ai-workbench-api && sudo systemctl restart ai-workbench-api.service && sleep 4 && curl -fsS http://127.0.0.1:8080/api/v1/health/storage"
```

Expected: `{"mysql":true,"redis":true}`

- [ ] **Step 6: 确认工作树状态**

```bash
git status --short -- api/
```

Expected: 只剩 `api/internal/handler/data/memory-store.json`（永久排除项）

### Task 2: 更新策略文档并提交

**Files:**
- New: `docs/aiops/findx_implementation_strategy_v3.md`（已创建）

- [ ] **Step 1: 提交策略文档**

```bash
git add -- docs/aiops/findx_implementation_strategy_v3.md
git commit -m "docs: add FindX implementation strategy v3

Key changes from previous approach:
- Platform backend: only remote Ubuntu verification needed
- Agent/probe code: Windows+Linux dual verification
- P0 worktree cleanup: batch commit, no per-file closure
- Shift to feature-module granularity (B1-B8)
- Technical debt tracked but not blocking"
```

### Task 3: P0 完成宣告

- [ ] **Step 1: 确认 git log**

```bash
git log --oneline -5
```

Expected: 看到 batch commit 和 strategy doc commit

- [ ] **Step 2: 确认 api/ 下无未提交修改**

```bash
git status --short -- api/ | grep -v "data/memory-store"
```

Expected: 无输出（干净）

**P0 工作树治理阶段完成。下一步进入阶段 B1：React Shell + 登录。**
