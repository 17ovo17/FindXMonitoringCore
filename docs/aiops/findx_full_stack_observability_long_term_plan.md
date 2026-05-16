# FindX 全栈可观测长期开发计划

更新时间：2026-05-08 01:17（UTC+8）

状态：当前唯一长期实施主计划。

## 1. 总则

FindX 的长期路线是：基于用户提供的成熟平台源码，复用成熟产品的信息架构、页面结构、组件状态流、API 语义和功能闭环，再统一替换为 FindX 自有品牌、风格、权限、审计、配置、数据源和 AI SRE 能力。

本计划明确废止以下方向：

- 直接 iframe / WebView 嵌入参考站。
- 嵌入参考站登录、SSO、侧边栏或运行态会话。
- 只复制截图。
- 自研弱化页面。
- MVP、最小实现、最小验证。
- 静态假按钮、静态假数据和假成功状态。
- 把 Vue workbench 补成最终完成态。

前端最终架构为 **React-only**。Vue 只允许作为迁移期临时桥，必须标记为 `TEMP_BRIDGE`、`REPLACED` 或 `REMOVE_AFTER_REACT`，不得作为最终验收基线。

## 2. 不可违背的实施原则

- 成熟源码是页面结构、路由语义、组件拆分、状态流、API 行为、按钮真实动作、错误态、空态、权限态和功能点事实源。
- FindX 自己负责登录、导航、权限、审计、主题、品牌、错误脱敏、统一配置、数据源中心、AI SRE、Evidence Chain 和 Agent 控制面。
- 用户侧只显示 FindX / FindX Agent / 链路监控 / 日志中心 / Agent 管理中心 / AI SRE 等 FindX 命名。
- 外部产品名称只允许出现在内部源码证据、合规登记、任务板、执行日志和归档文档。
- 成熟源码有真实动作的控件必须接真实动作；契约缺失时显示 `BLOCKED_BY_CONTRACT`。
- 组件不可用、数据源未配置、下游不可达时显示 `BLOCKED` 或脱敏错误，不造假数据。
- API 测试不能替代 MCP/Playwright 真实浏览器登录、点击、表单、抽屉/弹窗、错误态和窄屏回归。
- Browser Use 插件不可用时，必须切到 Playwright MCP 做真实浏览器回归并记录原因；真实浏览器不可用时只能标记 `BLOCKED` 或 `NOT_RUN`。
- 未跑 WSL build、lint/测试、浏览器回归和扫描，不得写 PASS。
- 所有子代理必须显式使用 `model: "gpt-5.5"`，禁止 fallback 到 5.4。

## 3. 成熟源码事实源

| 域 | 事实源 | FindX 实施方式 |
| --- | --- | --- |
| 基础监控 | `D:\项目迁移文件\平台源码\fe-main` | 数据源、系统集成、指标查询、仪表盘、模板中心、告警、通知、组织权限、系统配置按源码迁移 |
| 链路监控 | `D:\项目迁移文件\平台源码\skywalking-booster-ui-main`、`D:\项目迁移文件\平台源码\skywalking-master` | 服务目录、拓扑、Trace、Profiling、告警、OAP query、GraphQL/query 状态流按源码迁移 |
| SkyWalking Agent | Java、Python、Node.js、PHP、Go、Rust、Ruby、Nginx Lua、Kong、Browser Client JS 上游仓库 | 作为 FindX Agent 能力包进入包仓库、接入向导、配置模板、心跳、数据到达验证和 Evidence Chain |
| 日志中心 | `D:\项目迁移文件\平台源码\signoz-develop\frontend` | 日志检索、字段筛选、上下文、聚合、live tail、Pipeline、Saved Views、Trace 关联按源码迁移 |
| CMDB / Agent 在线 | `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main` | CMDB 树、主机表、分组、Agent 在线、部署、心跳、统计、终端/监控弹窗按源码迁移 |
| 采集插件 | `D:\项目迁移文件\平台源码\categraf-main (1)` | 作为 FindX Agent 插件目录、配置模板、采集插件能力来源 |
| 巡检诊断 | `D:\项目迁移文件\平台源码\catpaw-master` | 作为 FindX Agent 巡检诊断、结构化执行和 Evidence Chain 能力来源 |

## 4. 导航规划

用户侧主导航固定为：

- 基础设施
- 集成中心
- 数据查询
- 仪表盘
- 告警
- 通知
- 链路监控
- 日志中心
- Agent 管理中心
- AI SRE
- 组织权限
- 平台治理

归属规则：

- 数据源、系统集成、模板中心、采集模板、接入指引归集成中心。
- 指标查询、内置指标、对象快捷视图、记录规则归数据查询。
- 仪表盘列表、详情、变量、Panel、模板导入归仪表盘。
- 告警规则、事件、屏蔽、订阅、自愈、事件流水线归告警。
- 通知规则、通知媒介、消息模板、测试发送归通知。
- SkyWalking 服务目录、拓扑、Trace、Profiling、链路告警、接入归链路监控。
- SigNoZ 日志检索、字段筛选、上下文、Pipeline、Saved Views、Trace 关联归日志中心。
- SkyWalking Agent、多语言探针、网关探针、Browser Agent、Categraf 插件、Catpaw 巡检工具归 Agent 管理中心。
- Categraf 插件配置不仅是模板来源，也必须进入 FindX Agent 可配置下发范围；远程修改、灰度下发、回滚、漂移检测和审计链路未实现前只能显示 `BLOCKED_BY_CONTRACT`。
- FindX Agent 侧所有插件和监控能力都必须闭环；缺少 Windows、Linux、Kubernetes、SSH、WinRM、示例应用、OAP、Categraf、Catpaw 或日志/指标/Trace 环境时，先安装环境再验证监控与数据到达，不能用环境缺失跳过验收。

## 5. 数据源与配置一致性

所有平台组件使用统一配置中心和数据源中心：

- MySQL/MariaDB：权威业务数据、用户、权限、配置、知识库元数据。
- CMDB 自动发现和 Agent 生命周期相关资产状态的生产持久化必须沿用 Go GORM `store.GormOK()` + `store.GetDB()` 权威路径。旧实施文档 `docs/aiops/findx_implementation_strategy_v3.md` 已说明新模块使用 GORM `GetDB()` 模式；当前长期实施入口也同步继承该边界。
- Prometheus、FindX Agent、Categraf、Catpaw 等发现结果写入 CMDB 时，必须通过 `store.CreateCmdbInstance` / `store.UpdateCmdbInstance`，在生产 MySQL 可用时落到 GORM/MySQL；不得把 `api/internal/handler/data/memory-store.json` 当作 CMDB 自动发现或 Agent 生命周期完成态证据。
- GORM/MySQL 不可用时，只允许标记 `BLOCKED`、`RISK` 或 `memory fallback dev-only`；不得宣称生产持久化成功。`memory-store.json` 仅是运行态 fallback/sensitive candidate，不得 stage，不得作为验收证据。
- Redis：缓存、会话、队列、限流、短期状态。
- MinIO/S3：截图、报告、附件、证据产物、离线包。
- Qdrant：知识库向量索引加速层，可重建，不作为权威数据源。
- BM25：关键词检索和向量不可用降级路径。
- Prometheus/VictoriaMetrics：指标查询和告警。
- SkyWalking OAP：链路、服务、拓扑、Profiling、APM 告警。
- BanyanDB：SkyWalking 推荐链路存储候选。
- SigNoZ Query Service：日志、Trace/日志关联、查询视图。
- ClickHouse：SigNoZ 日志和事件分析存储，不作为 SkyWalking OAP 的直接替代存储。
- OTel Collector：Telemetry 接入、协议转换、路由、采样和脱敏。
- ES/OpenSearch：仅作为明确需要时的链路兼容或日志搜索可选项。

真实凭据只以引用方式保存和传递，前端不回显密文，错误不显示完整连接串。

## 6. 页面和功能迁移要求

每个切片编码前必须提交源码证据清单：

- 来源源码路径。
- 上游路由和 FindX 路由 alias。
- 核心组件和子组件拆分。
- API 调用、请求字段、响应字段。
- 状态流和关键状态枚举。
- 按钮真实动作。
- 表格列、筛选区、工具栏、分页、批量操作。
- 抽屉、弹窗、详情页。
- 空态、错误态、权限态、加载态。
- FindX 命名替换表。
- 是否 `API_CONTRACT_CHANGE`。
- 是否 `DATA_CHANGE`。
- 回滚策略。
- MCP/Playwright 验证路径。

## 7. SkyWalking Agent 一等主线

SkyWalking Agent 从现在起按一等能力处理，不是链路监控附属说明，也不是 P5 后面一句话。

必须纳入 FindX Agent 能力包：

- Java Agent
- Python Agent
- Node.js Agent
- PHP Agent
- Go Agent
- Rust Agent
- Ruby Agent
- Nginx Lua Agent
- Kong Agent
- Browser Client JS

贯穿阶段：

- P0：记录源码 URL、本地源码状态、版本/commit、许可证、NOTICE、包形态、安装方式、配置项、数据到达验收。
- P2：Agent 管理中心必须有能力包目录、接入向导、平台/语言筛选、配置模板、心跳/数据到达状态；契约缺失显示 `BLOCKED_BY_CONTRACT`。
- P3：链路监控必须和 Agent 状态联动，服务目录看覆盖率，Trace 详情反查探针状态，拓扑节点反查应用/网关/Kong 探针，链路告警复盘引用安装和配置证据。
- P5：完成完整生命周期，Linux `curl -kfsSL`、Windows CMD `certutil -urlcache -f`、PowerShell `Invoke-WebRequest`、SSH、WinRM、systemd、Windows Service、IIS、Docker、Helm、Operator、DaemonSet、Sidecar、InitContainer。

本机安装、远程下发、远程安装、卸载、配置下发、插件下发、升级、回滚、服务状态、心跳、数据到达、审计和失败恢复都必须证明真实成功。Linux `curl -kfsSL`、Windows CMD `certutil -urlcache -f`、PowerShell `Invoke-WebRequest` 不能只做命令预览或复制入口，必须能关联安装任务、执行日志、服务注册、心跳和数据到达证据。

SkyWalking Agent、Categraf、Catpaw 以及 FindX Agent 下所有监控插件必须覆盖 Windows/Linux 双实现、双验证；Kubernetes 场景按 Helm、Operator、DaemonSet、Sidecar、InitContainer、RBAC、TLS 和回滚单独闭环。

未补齐源码、包仓库、签名、安装计划、配置模板、真实安装/卸载/远程执行、心跳、数据到达验证前，不能用静态清单、命令预览或 `409 BLOCKED_BY_CONTRACT` 冒充完成。

## 8. Agent 控制面能力

FindX Agent 管理中心必须覆盖：

- 包仓库：包名、语言、运行时、平台、架构、版本、commit、许可证、NOTICE、签名、校验和、兼容矩阵、离线包。
- 安装向导：选择主机、服务、Kubernetes workload、能力包、配置模板、凭据引用、安装前检查、维护窗口、风险提示。
- 本机安装：Linux `curl -kfsSL`、Windows CMD `certutil -urlcache -f`、Windows PowerShell `Invoke-WebRequest`。
- 远程安装：Linux SSH/shell/systemd，Windows WinRM/PowerShell/Windows Service/IIS，Kubernetes Helm/Operator/Manifest/DaemonSet/Sidecar/InitContainer。
- 配置模板：OAP endpoint、service name、instance name、namespace、environment、采样、插件 include/exclude、日志关联、Trace propagation、TLS/proxy、标签、资源限制。
- 配置下发：单机、批量、业务组、namespace/workload、灰度批次、失败阈值、自动回滚、漂移检测。
- 插件远程修改：Categraf 输入插件、主机插件、容器插件、网关插件等配置必须能按 Agent/CMDB 目标远程修改、下发、回滚和审计；缺少远程执行契约时显示 `BLOCKED_BY_CONTRACT`，不得伪装修改成功。
- 心跳状态：控制面心跳、进程状态、服务状态、探针加载状态、OAP 连通性、最后 Trace/Metric/Log/RUM 到达时间。
- 数据到达验证：Trace、Metric、Log correlation、RUM、网关 Trace、Profiling 信号逐项验证。
- 升级、回滚、卸载和 Evidence Chain。

## 9. AI SRE 与 Evidence Chain

AI SRE 不替代监控、日志、Trace、CMDB、Agent 或工作流能力。它只能基于已有证据做诊断、解释、编排、复盘和修复建议。

Evidence Chain 必须接入：

- 指标查询结果、告警触发值、规则版本。
- 日志查询、字段筛选、上下文片段。
- Trace、Span、服务拓扑、慢调用、错误链路。
- CMDB 实例、主机、业务组、变更、Agent 心跳。
- 安装任务、配置下发、远程执行、数据到达验证。
- 巡检任务、脚本输出、文件证据。
- 工作流执行记录、人工审批、修复动作、回滚动作。
- 知识库引用、Runbook 版本、模型配置引用。

证据缺失时必须返回“数据缺失 / 组件不可用 / BLOCKED”，不得编造结论。所有 AI 调用必须走系统配置中的 AI 模型配置。

## 10. 知识库与检索架构

知识库使用 MySQL/MariaDB 做权威基石，Qdrant 做内置向量数据库加速层，BM25 做关键词检索和向量不可用降级路径，RRF 做混合召回融合。

- MySQL/MariaDB 保存文档、段落、权限、版本、来源、审计、引用关系和索引状态。
- Qdrant 保存可重建的向量索引，不作为权威数据源。
- BM25 支持无向量环境和精确关键词召回。
- 索引任务必须有状态、进度、失败原因、重试、重建和增量同步。
- Evidence Chain 引用知识库时必须记录文档版本、段落 ID、召回方式、相关度和模型配置引用。

## 11. 实施阶段

### FX-NIGHT-000-DOC-GOAL-LOCK

把 V2 计划、Goal 指令、React-only 边界、多 Agent 规则、SkyWalking Agent 主线写入项目文档。

### FX-NIGHT-001-WORKTREE-BASELINE

只读梳理脏工作树，分离文档、React 迁移、通知旧切片、运行态文件、未跟踪代码、删除候选、敏感候选。不 reset、不 revert、不删除历史证据。

### P0-SOURCE-MATRIX-LOCK

只读审计所有成熟源码入口和参考运行态 DOM，输出页面/路由/API/状态流矩阵。没有矩阵不得编码。

### P0-CONFIG-DATASOURCE-CONTRACT

统一配置文件、统一数据源中心、凭据引用、测试连接、错误脱敏、组件 `BLOCKED` 状态和审计模型。

### P0-WEB-REACT-ONLY-SHELL-NAV

建设 React-only FindX 自有壳层、登录、导航、主题、权限、审计和全局错误。禁止嵌入参考站。

### P0-BASE-MONITORING-REAL

按 Nightingale React 源码迁移数据源、系统集成、指标查询、仪表盘、模板中心、告警、通知、系统配置。

### P1-ORG-PERMISSION-GOVERNANCE

组织、团队、角色、权限、SSO、审计日志、站点设置、AI 模型配置、平台治理。

### P2-CMDB-AGENT-REAL

按 AutoOps/AIOps 源码迁移 CMDB、主机、Agent 在线、部署、心跳、统计和终端/监控弹窗；同时提供 SkyWalking Agent 能力包目录和接入入口。

持久化边界：CMDB 自动发现必须验证生产 MySQL/GORM 路径。Prometheus、FindX Agent、Categraf、Catpaw 发现结果必须经 `store.CreateCmdbInstance` / `store.UpdateCmdbInstance` 写入；`store.GormOK()` 为真时必须通过 `store.GetDB()` 落到 MySQL。后续 FX-NIGHT-119E / FX-NIGHT-126 验证必须覆盖 MySQL/GORM 持久化、服务重启后数据仍在、fallback 不污染 Git；GORM 不可用只能记录 `BLOCKED` / `RISK` / `memory fallback dev-only`，不能用 `memory-store.json` 证明完成。

### P3-APM-SKYWALKING-REAL

按 SkyWalking UI/OAP 源码迁移服务目录、拓扑、Trace、Profiling、告警和接入；与 Agent 状态联动。

### P4-LOGS-SIGNOZ-REAL

按 SigNoZ 源码迁移日志检索、字段筛选、上下文、聚合、Pipeline、Saved Views 和 Trace 关联。

### P5-FINDX-AGENT-SUITE

P0-P4 前端闭环后，完成 FindX Agent 包仓库、离线包、签名、本机/远程/K8s 安装、配置下发、心跳、数据到达、升级、回滚、卸载、审计和 Evidence Chain。

### P6-AISRE-EVIDENCE-CHAIN

完成诊断会话、工作流、健康检查、复盘报告、Evidence Chain、自动修复编排和 AI 模型统一配置。

### P7-KNOWLEDGE-VECTOR

完成 MySQL 权威知识库、Qdrant 向量索引、BM25、RRF、索引任务、权限和引用链。

### P8-P10 商业化补齐

补齐 SLO/SLA、事件管理、值班升级、变更关联、Synthetic Monitoring、云资源观测、Kubernetes 观测、网络观测、RUM、移动端、数据治理、权限治理、可靠性治理、离线交付、授权边界、升级回滚和合规材料。

## 12. 多 Agent 任务领取制

默认并行池：

- Worker A：当前最高优先级实现切片。
- Worker B：下一个写集不冲突的前端/Adapter 切片。
- Explorer/QA C：只读源码证据、diff 审查、品牌/敏感/静态假按钮审查。
- Worker D：仅在句柄稳定且写集完全互斥时临时启用。

所有子代理必须显式 `model: "gpt-5.5"`。每个 agent 必须有 `FX-NIGHT-*` 或 `FX-TASK-*`、agent id、写集、禁止写集、证据源和验收门禁。完成、失败、超时、`BLOCKED` 或不再需要时立即关闭。

## 13. 测试与验收

每个 UI/API 切片必须覆盖：

- 源码证据门禁。
- Windows build。
- WSL/Linux build 或对应后端测试。
- Lint 或 `NOT_RUN` 原因。
- MCP/Playwright 真实登录、点击、表单、抽屉/弹窗、错误态、窄屏回归。
- Browser Use 不可用时使用 Playwright MCP，并在任务板或执行日志记录降级原因。
- 正常路径、异常路径、边界或权限路径。
- 组件不可用时明确显示 `BLOCKED`。
- 成熟源码中有真实动作的控件，在 FindX 中必须接真实动作；未完成直接 `FAIL` 或 `BLOCKED_BY_CONTRACT`。
- 敏感扫描、品牌扫描、静态假按钮扫描。

阻断项：

- P0/P1 RISK。
- 敏感信息泄露。
- 权限绕过。
- 错误未脱敏。
- 静态假按钮。
- 乱码。
- 用户侧外部品牌暴露。
- 未真实执行却写 PASS。

## 14. Git 与文档治理

- 本计划是当前唯一长期实施主计划。
- 旧方向文档全部标记 superseded 或列入归档索引，不再作为实施依据。
- 历史证据、截图、DOM snapshot、JSON、测试报告不直接删除；先建立候选清单并检查引用。
- 每次计划变更必须同步 README、`docs/aiops/README.md`、AGENTS 和任务板。
- 业务代码、运行态配置、密钥、Cookie、DSN、会话 ID 不得进入文档计划提交。
- 每次实现、配置、依赖、路由、验证、测试、反卡死处理或 P0/P1 风险处理，都必须同步 `.codex/codex-task-board.md` 和 `.codex/operations-log.md`。
