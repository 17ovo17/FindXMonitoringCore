<h1 align="center">FindX Monitoring Core</h1>
<p align="center">新一代监控核心平台：监控事实、告警事件、Agent 证据链、AI 问诊与自动修复闭环</p>
<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21-00ADD8?logo=go" />
  <img src="https://img.shields.io/badge/Vue-3.5-4FC08D?logo=vue.js" />
  <img src="https://img.shields.io/badge/MySQL-8-4479A1?logo=mysql" />
  <img src="https://img.shields.io/badge/License-MIT-green" />
</p>

---

## 什么是 FindX Monitoring Core

FindX Monitoring Core 是 FindX 项目的新主线，不是 Nightingale 外壳，也不是把旧 Catpaw/N9E 页面换名复用；它是一套面向现代运维场景的新一代监控核心平台。FindX 会参考 Nightingale 在告警、Dashboard、通知、模板、权限、事件流水线等领域的成熟设计，并在合规边界内融合可复用源码，但最终运行时 API、数据模型、权限、UI、AI 证据链、Agent 与自动修复闭环都归 FindX 自己。

| 主线域 | 主入口 | 当前状态 |
|------|------|---------|
| Monitoring Core 健康与 Query Gateway | `/api/v1/monitor/health`、`/api/v1/monitor/query`、`/api/v1/monitor/query-range` | 已实现并纳入 QA 回归 |
| Target 管理 | `/api/v1/monitor/targets` | 已实现并纳入权限/持久化回归 |
| Alert Rule 与 Scheduler | `/api/v1/monitor/alert-rules` | 已实现规则 CRUD、启停、克隆、试运行、回滚；P1-BE-3 自动事件闭环已 QA PASS |
| Event 处置 | `/api/v1/monitor/events/current`、`/api/v1/monitor/events/history`、`/api/v1/monitor/events/:id/*` | 已实现 current/history 与 ack/assign/resolve/archive 闭环 |
| FindX Agent 注册与心跳 | `/api/v1/findx-agents`、`/api/v1/findx-agents/register`、`/api/v1/findx-agents/heartbeat` | 契约已定义，基础注册/心跳/list 已落地；深度巡检与安装发行仍属 P2 |
| Dashboard、模板、通知、静默、订阅、团队、权限审计、值班 | 以 FindX Monitoring Core 域模型建设 | 规划/待实现或部分历史能力待迁移，禁止标记为完成 |
| AI evidence chain | 监控事件 + Query Gateway + Target/Agent + Runbook/知识库引用 | 契约已定义，P3 深度闭环待实现 |
| Remediation 自动修复 | 规划路径 `/api/v1/remediation/*` | 规划待实现；未注册前只能写 NOT_RUN/BLOCKED，不得写 PASS |
| Catpaw/N9E 兼容入口 | `/api/v1/catpaw/*`、`/api/v1/n9e/*` | 仅做历史兼容回归，不作为新功能主入口 |

历史 AI WorkBench 能力仍作为 FindX 的诊断、工作流、知识库和 AIOps 基座延续。公开项目式招募、交流群、抓虫宣传不再作为首屏定位。下文保留的 AI WorkBench 描述仅用于说明历史能力来源和已沉淀模块。

## 历史能力说明：AI WorkBench

AI WorkBench 是一个全栈 AIOps 平台，将 LLM 推理能力与结构化工作流引擎结合，实现从告警接入到智能诊断、知识沉淀的全链路闭环。

**核心理念：LLM 负责理解和推理，代码负责精确计算，工作流负责编排协调。**

## 核心能力

| 能力 | 说明 |
|------|------|
| **智能诊断** | 结构化推理引擎（假设→证据→验证循环），不是让 LLM 猜 |
| **工作流编排** | Go 原生 DAG 引擎，20 种节点类型，**30 个内置工作流**（5 类） |
| **Prompt 外化** | 全部 LLM 节点 system_prompt 外化到 `api/assets/prompts/*.txt`，可独立迭代 |
| **混合搜索** | BM25 + 向量搜索 + RRF 融合 + Reranker 四层检索 |
| **时序分析** | CUSUM 突变检测 + 趋势回归 + 异常评分 + 周期性检测 |
| **告警闭环** | 多源接入 → 风暴抑制 → 智能路由 → 自动诊断 → 知识沉淀 |
| **知识沉淀** | 案例库 + Runbook + 文档管理，诊断完成自动归档 |
| **监控集成** | Prometheus 指标扫描 → AI 自动适配 → 标准名映射 |

## 技术栈

```
后端    Go 1.21 + Gin          前端    Vue 3.5 + Vite 5 + Element Plus
数据库  MySQL 8                 缓存    Redis（可选，内存 fallback）
监控    Prometheus              搜索    BM25 + 向量(DashScope/OpenAI兼容)
沙箱    goja (JS)               分词    go-ego/gse
```

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- MySQL 8（必须，自动建表）
- Redis（可选）
- Prometheus（可选）

### 1. 克隆项目

```bash
git clone https://github.com/17ovo17/AI-WorkBench.git
cd AI-WorkBench
```

### 2. 配置后端

```bash
cd api
cp config.yaml.example config.yaml
# 编辑 config.yaml，填入 MySQL 连接信息
# AI 模型、Embedding、数据源等在前端界面配置
```

### 3. 启动后端

```bash
go mod tidy
go run main.go
# 服务监听 http://localhost:8080
```

### 4. 启动前端

```bash
cd ../web
npm install
npm run dev
# 前端监听 http://localhost:3000
```

### 5. 访问平台

打开浏览器访问 `http://localhost:3000`，进入「AI 模型」页面配置你的 LLM API Key。

## 项目结构

```
AI-WorkBench/
├── api/                          # Go 后端
│   ├── main.go                   # 入口（156 个 API 端点）
│   ├── config.yaml.example       # 配置模板
│   └── internal/
│       ├── handler/              # HTTP 处理层
│       ├── store/                # 数据持久层（MySQL + 内存 fallback）
│       ├── model/                # 数据模型
│       ├── workflow/             # 工作流引擎
│       │   ├── engine/           # DAG 核心 + 20 种节点
│       │   │   └── builtin/      # 内置工作流 YAML
│       │   └── node/             # 节点实现
│       ├── reasoning/            # 结构化推理引擎
│       ├── timeseries/           # 时序分析引擎
│       ├── embedding/            # 混合搜索引擎
│       ├── knowledge/            # 知识库管理
│       ├── correlator/           # 跨数据源关联
│       ├── eventbus/             # 事件总线
│       ├── middleware/           # 熔断器 + 限流
│       ├── scheduler/            # 定时调度
│       └── security/             # 安全机制
├── web/                          # Vue 3 前端
│   └── src/
│       ├── views/                # 26 个页面
│       └── components/           # 可复用组件
├── docker/                       # Docker + Prometheus 配置
├── scripts/                      # 启停脚本
└── docs/                         # 项目文档
```

## 工作流引擎

Go 原生实现的 DAG 工作流引擎，零外部依赖。

### 20 种节点类型

| 类型 | 作用 |
|------|------|
| `start` / `end` | 工作流入口和出口 |
| `llm` | 调用大模型推理 |
| `knowledge_retrieval` | 知识库混合搜索 |
| `http_request` | HTTP 请求 |
| `code` | JavaScript 沙箱执行 |
| `condition` | 条件分支（支持 in/starts_with/regex 等） |
| `loop` / `iteration` | 循环和迭代 |
| `sub_workflow` | 子工作流编排 |
| `human_input` | 暂停等待用户输入 |
| `agent` | LLM + 工具自主循环 |
| `tool` | 工具调用 |
| `template_transform` | 模板渲染 |
| `variable_aggregator` / `variable_assigner` | 变量操作 |
| `parameter_extractor` / `question_classifier` | LLM 提取/分类 |
| `list_filter` / `document_extractor` | 数据处理 |

### 30 个内置工作流（5 类）

| 分类 | 数量 | 代表工作流 |
|------|------|-----------|
| **诊断** | 10 | diagnosis / smart_diagnosis / alert_diagnosis / domain_diagnosis / container_diagnosis / jvm_diagnosis / db_lock_analysis / slow_query_diagnosis / log_analysis / incident_postmortem |
| **巡检** | 7 | health_inspection / business_inspection / dependency_health / storage_health_check / middleware_inspection / network_check / incident_review |
| **分析** | 6 | metrics_insight / metrics_analysis / capacity_forecast / slo_compliance / traffic_anomaly_detect / incident_timeline |
| **安全** | 4 | security_compliance / security_audit / ssl_audit / config_drift_detect |
| **操作** | 3 | runbook_execute / change_rollback / knowledge_enrich |

> 📘 **详细流程图**：每个工作流的节点流向、数据流表、LLM 角色定位详见 [`docs/architecture/工作流流程图.md`](docs/architecture/工作流流程图.md)（含 mermaid 图）

30 个 YAML 通过别名映射合并为 9 个对外核心工作流：`smart_diagnosis / domain_diagnosis / health_inspection / metrics_insight / security_compliance / incident_review / network_check / runbook_execute / knowledge_enrich`（详见 [系统架构全景 §4.5](docs/architecture/系统架构全景.md)）。

### 智能路由

用户输入自然语言，系统自动选择最匹配的工作流：

```
"JVM Full GC 频繁"  → domain_diagnosis (domain=jvm)
"做一次全面巡检"     → health_inspection (scope=full)
"SSL 证书快过期了"   → security_compliance (audit_type=ssl)
```

## 安全机制

| 机制 | 说明 |
|------|------|
| JS 沙箱 | goja 运行时冻结全局对象 + 删除 eval/Function + Unicode 归一化 |
| 命令安全 | L0-L4 四级分级，工作流模式 L2 降级 L1，L3/L4 拒绝 |
| 审计日志 | HMAC-SHA256 签名防篡改 |
| 告警风暴 | 60 秒内同 IP >20 条自动抑制 |
| LLM 并发 | 信号量限制 5 并发 + 熔断器 |
| API 限流 | 令牌桶算法 100 req/s |

## API 概览

156 个 REST API 端点，按功能分组：

- **对话与 AI** — 通用对话、Agent 对话
- **AIOps 智能运维** — 会话、巡检、WebSocket 实时推理
- **诊断** — 启动诊断、记录管理、反馈、归档
- **告警** — Catpaw/Alertmanager/夜莺 webhook 接入
- **知识库** — 案例 CRUD、文档上传、Runbook 管理、语义搜索
- **工作流** — CRUD、执行、SSE 流式、定时调度
- **拓扑** — 业务拓扑管理、AI 生成、巡检
- **监控** — Prometheus 集成、指标扫描、AI 适配
- **设置** — AI 模型、Embedding、Reranker、数据源、通知渠道

完整 API 文档见 [`docs/API文档.md`](docs/API文档.md)。

## FindX Monitoring Core 项目规划

近期主线是把 AI WorkBench / FindX 演进为 **FindX Monitoring Core**：FindX 是主平台、主视觉、主入口和最终运行主体；Nightingale 是被深度嵌入、完整对标和可融合改造的成熟监控能力来源。方向不是把 FindX 做成 Nightingale 外壳，也不是只接几个 Nightingale API；而是把 Nightingale 已经具备的监控控制面能力完整纳入 FindX 的产品框架、数据模型、权限体系、页面导航和运行闭环。

硬边界：**Nightingale 有的基础监控能力，一个都不能少**。FindX 的 AI 问诊、知识库记忆、Agent 巡检、单机对话和自动修复是锦上添花的增强层，不能替代 target、业务组、Dashboard、告警、通知、事件流水线、模板、权限、审计等基础能力。Nightingale 的功能行为可以参考、融合和改造；最终 API、数据库、UI、Agent、AI 证据链和自动修复闭环归 FindX 自己承载。

Nightingale 全功能覆盖基线如下，详细源码证据和实现顺序见 [`docs/aiops/nightingale_findx_agents_project_plan.md`](docs/aiops/nightingale_findx_agents_project_plan.md)。

| 基础域 | FindX 必须覆盖的 Nightingale 能力 |
| --- | --- |
| 身份与资源边界 | 用户、个人资料、密码、自助 token、SourceToken/API Token、团队/用户组、角色、操作点、业务组、多租户隔离 |
| 监控对象与采集 | target、host meta、机器存活、PushGW/Agent heartbeat、target update、Categraf 插件生态、采集模板、remote write |
| 数据与查询 | 数据源、PromQL 即时/区间查询、批量查询、日志查询、Trace logs、ES/OpenSearch/TDengine/DB 查询、指标说明、指标视图、SavedView |
| Dashboard 与资产 | Dashboard/Board、Chart、Chart 分享、注解、变量、纯净视图、嵌入产品、内置组件、builtin payload、builtin metrics、integrations 模板 |
| 告警与事件 | 告警规则、PromRule 导入、记录规则、evaluator、current/history event、event detail、eval detail、聚合视图、事件处置时间线 |
| 通知协同 | 通知渠道、通知配置、通知规则、通知模板、消息模板、通知记录、联系人 key、webhook/script/provider、静默、订阅、值班、升级 |
| 流水线与任务 | 事件流水线、processor、try-run、trigger、SSE stream、执行记录、任务模板、任务记录、后台任务、导入导出、版本回滚 |
| AI 与工具基础 | AI assistant、AI agent、LLM config、MCP server、AI skill、内置工具、内置技能包；这些 Nightingale 已有能力在 FindX 中必须受控融合 |
| FindX 增强层 | 知识库记忆标签、Agent/CMDB 工作台、单机 Agent 对话、Catpaw 授权衍生巡检、evidence chain、自动修复审批执行闭环 |

Categraf 插件生态会直接保留并改造成 `findx-agents` 发行版；已获授权的 Catpaw 能力会衍生进 `findx-agents`，提供巡检、诊断、远程会话、结构化工具和自动修复执行能力。AI 问诊直接基于 FindX 自有监控事实、Agent 巡检证据、通知记录、自动修复记录和知识库/Runbook 案例做推理，并把告警触发、Agent 补证据、审批、执行、验证、回滚、审计串成正式闭环。

所有能力以 **WSL/Linux 兼容为准**：后端构建、前端构建、Agent 安装路径、服务名、日志路径、Run & Debug 脚本和浏览器回归都必须在 `/opt/ai-workbench` 或目标 Linux 运行态验证通过；Windows 结果只作为开发侧辅助证据。

### 当前进展 / 已落库切片

当前已实现切片记录：

- **P0 文档规划闭环**：已明确 FindX Monitoring Core 新平台定位、Nightingale 参考/改进边界、Categraf 复用、Catpaw 授权衍生、AI 问诊与自动修复核心增强、WSL/Linux 验收基准。
- **P0 后端 API 基座**：已完成并通过 QA 门禁，提交记录为 `81b4531 feat: add FindX monitoring core P0 APIs`；QA 最终评分 **98/100**，结论 **通过**。
- **P1-BE-3 alert scheduler current/history 自动闭环**：已完成规则调度、current event upsert、history recovery、失败只写 eval log、不制造误告警、并发 RunOnce 锁保护和安全摘要要求；提交记录为 `1c4045e feat: add FindX alert scheduler event closure`，QA **92/100 PASS**，Windows/WSL `go test -count=1 ./...` 与 build 通过。
- **安全基线修复**：CORS 空配置默认保持 localhost-only；Workflow DSL 创建/更新在 parse 后执行严格 graph validation，坏 DSL 返回 400 且不得持久化。

已落库能力包括：monitor health、targets、findx-agents register/heartbeat/list、alert-rules、current/history events、query gateway datasources/query/query-range/metrics/labels/label-values、alert scheduler 自动评估闭环。当前实现确认读接口走平台认证，写接口要求 admin，Agent token 默认拒绝匿名；Prometheus 上游失败返回 503；查询审计仅记录 hash/stats；事件终态保护已纳入后端状态约束。

已完成 Windows 与 WSL/Linux 双环境验证，其中 WSL 侧已在 `/opt/ai-workbench/api` 执行后端单元测试和 `go build -o api-linux .`；P1-BE-3 切片额外通过 Windows/WSL `go test -count=1 ./...` 与 build。P1 剩余 Dashboard、模板、通知、权限审计、团队/值班/订阅/静默，P2 `findx-agents` 深度融合，P3 AI 问诊与自动修复仍按后续阶段推进，不按已完成能力描述。

| 阶段 | 目标 |
|------|------|
| **P0：FindX Core 基座** | 已完成后端 API 基座与 QA 门禁：target、datasource、query、alert rule、event、audit、heartbeat、`/api/v1/monitor/*` 和 `/api/v1/findx-agents/*` 主接口 |
| **P1：监控控制面全功能补齐** | 已完成 P1-BE-3 alert scheduler 自动事件闭环；继续补 Dashboard/Chart/分享/注解、模板中心、通知全链路、权限审计、team/oncall/subscription/silence、记录规则、聚合视图、Trace logs、SavedView、SourceToken、嵌入产品 |
| **P2：事件流水线、模板生态与 Agent/CMDB** | 补事件流水线 processor/stream/execution、任务/任务模板、integrations 资产安装 diff/回滚；建设 Agent/CMDB 工作台、机器存活、业务组分配、凭据引用、远程/本机安装、配置分发、单机对话 |
| **P3：AI 问诊与自动修复** | 告警触发 AI 问诊、Agent 巡检补证据、知识库记忆标签注入、evidence chain、修复 precheck/dry-run/approve/execute/verify/rollback 全链路 |
| **P4：生产化治理与生态闭环** | 补 LICENSE/NOTICE/来源版本/修改说明、模板市场治理、Run & Debug 脚本、审计留存、容量/性能压测、Linux/WSL 安装升级回滚手册 |

### 后续 QA 风险与改进动作

| 风险 | 后续动作 |
|------|----------|
| 大基数候选截断语义 | 明确 evaluator 对超大候选集的截断策略、返回摘要、审计字段和用户可见提示，避免“部分评估”被误认为全量成功。 |
| Workflow API 严格 DSL 400 回归 | 将坏 DSL 创建/更新返回 400 且不持久化纳入固定回归，覆盖缺节点、悬空边、循环、非法节点类型和错误分支。 |
| 持久化失败一致性 | 补齐写入失败后的响应、审计、eval log、current/history event 一致性要求，避免成功响应和实际落库不一致。 |
| 统一 `.gitattributes` | 后续补齐仓库换行符、Markdown、Go、Vue、脚本文件策略，降低 Windows/WSL/Linux 交叉开发漂移。 |

### 运维落地闭环

FindX Monitoring Core 的交付不只看 API 能否调用，还必须覆盖团队、权限、告警升级、模板导入导出、多租户、审计、通知通道、值班日历、Run & Debug 脚本、日志输出和回滚说明。`findx-agents` 发行版需要形成 `findx-agents.service`、`/opt/findx-agents`、`/etc/findx-agents`、`/var/log/findx-agents` 的安装、升级、回滚、卸载和故障定位闭环。

### 文档自评分

本轮文档自评分 **97/100**。加分点是已经把 Nightingale 全功能域、Agent/CMDB 工作台、知识库记忆标签、FindX AI 增强层和 Linux/WSL 验收基线写成统一路线；扣分点是 P1/P2/P3 大量能力仍待代码实现验证，Catpaw 衍生 LICENSE/NOTICE/来源版本/修改说明仍需随 `findx-agents` 实现真实落仓。后续每个稳定切片必须补 QA 证据、同步 WSL/Linux 构建结果、补许可证材料，并把大基数截断、Workflow DSL 400、持久化失败一致性、`.gitattributes` 纳入固定后续计划。

边界原则：

- 不允许 MVP、占位页、半截接口、静态假数据、只读目录或一次性 PoC 作为交付结论；阶段拆分只是交付顺序，不代表功能缩水。
- FindX 是新平台，不需要照顾现有生产 Nightingale 的无缝迁移，也不迁移历史监控数据或历史告警事件。
- Nightingale 是完整监控能力基线、参考实现、源码参考和可融合对象；FindX 最终独立运行，不把 Nightingale 作为运行期数据来源。
- 不能把“AI 问诊、Agent 巡检、自动修复”当成删减 Nightingale 基础能力的理由；基础监控域缺一项即视为规划缺口。
- 可参考和改造 Nightingale 的告警、Dashboard、通知、模板、权限、事件流水线、任务等成熟设计，但最终 API、数据模型、UI、权限、AI、Agent、自动修复都归 FindX 自己。
- findx-agents 直接保留 Categraf 插件生态；Catpaw 授权能力衍生进 findx-agents，形成采集 + 巡检 + 诊断 + 会话 + 自动修复执行的一体化 Agent。
- 自动修复作为正式核心项目纳入路线，必须完整实现权限、审批、试跑、验证、回滚和审计。
- 产品/UI/API 使用 FindX 命名，合规文档保留 Nightingale、Categraf、Catpaw 的来源说明、授权记录和修改说明。

详细立项文档见 [`docs/aiops/nightingale_findx_agents_project_plan.md`](docs/aiops/nightingale_findx_agents_project_plan.md)。

## 前端页面

| 页面 | 功能 |
|------|------|
| 运维总览 | 监控健康、告警态势、任务进度、AI 诊断与知识沉淀总览 |
| 监控运维 | Nightingale 全功能域在 FindX 内的主工作台：业务组、监控对象、机器存活、数据源、查询、Dashboard、模板、告警、事件、通知、静默、订阅、值班、审计 |
| Agent/CMDB | 主机资产、业务组分配、Agent 生命周期、远程安装、本机安装、配置分发、单机对话、巡检证据、凭据引用、安装任务审计 |
| 智能诊断 | AIOps 多轮对话、告警问诊、事件证据链、Agent 巡检补证据 |
| 知识中心 | 案例库、Runbook、文档管理、语义搜索、LTM/STM/历史记录、标签 key 分类分级和 Prompt 注入策略 |
| 工作流/自愈 | 诊断工作流、受控修复计划、审批、试跑、执行、验证、回滚 |
| 告警中心 | current/history event、聚合视图、处置时间线、通知记录、升级状态 |
| 业务拓扑 | 拓扑可视化、AI 生成、业务巡检、业务组资源关系 |
| AI 模型 | LLM Provider 配置 |
| 系统配置 | 数据源 + 指标映射 + Embedding + 通知 |

## 许可证

MIT License
