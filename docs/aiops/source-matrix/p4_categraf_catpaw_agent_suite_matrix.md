# FindX P4 Agent Suite 同源矩阵

生成时间：2026-05-07 14:09（UTC+8）
状态：`P4-CATEGRAF-CATPAW-AGENT-SUITE` 编码前门禁，不代表 FindX 已完成实现

## 1. 结论

FindX Agent Suite 必须把成熟采集插件和巡检诊断源码作为事实源接入，不能只做一个静态 Agent 列表。采集插件、配置模板、远程安装、配置下发、心跳、版本、数据到达、主动巡检、诊断会话、诊断记录、自检、Evidence Chain 都必须进入同一套 FindX Agent 控制面。

源码证据：

- [Categraf / Catpaw 到 FindX Agent Suite 源码检查证据](evidence/categraf_catpaw_agent_suite_source_snapshot.md)
- [P0 SkyWalking Agent 到 FindX Agent 控制面矩阵](p0_skywalking_agent_real_matrix.md)

当前阻断：

- 本地采集插件和巡检诊断源码存在。
- 未配置可访问 Agent Suite 运行态页面，运行态 DOM 证据为 `BLOCKED_RUNTIME_UNAVAILABLE`。
- FindX Agent 包仓库、配置模板、远程安装、任务下发、心跳、数据到达、诊断会话、Evidence Chain 契约未完成。

## 2. FindX 路由规划

| FindX 页面 | 成熟源码 | 页面结构要求 |
| --- | --- | --- |
| Agent 管理中心 | AutoOps Agent + FindX Agent 控制面 | 概览、包仓库、安装计划、心跳、版本、覆盖率、任务、审计 |
| 采集插件目录 | `inputs\*`、`conf\input.*\*.toml` | 分类、图标、说明、配置模板、适用场景、依赖、版本、启用状态 |
| 采集配置 | `conf\config.toml`、`conf\input.*` | global、writer、heartbeat、prometheus、logs、input 配置、密文引用、校验 |
| 配置下发 | service/update/config 源码 + FindX 任务系统 | 目标、灰度、批次、失败阈值、进度、日志、回滚、审计 |
| K8s 接入 | `k8s\daemonset.yaml`、`deployment.yaml`、`sidecar.yaml` | namespace、workload、RBAC/TLS、DaemonSet/Sidecar/Deployment 策略 |
| 巡检诊断 | `catpaw main.go`、`agent\inspect.go`、`digcore\diagnose` | 插件、目标、参数、执行计划、流式输出、报告、记录、证据 |
| AI 排障会话 | `chat`、`digcore\server\session*` | 多轮会话、工具调用、权限确认、流式输出、审计 |
| 诊断工具自检 | `selftest`、`diagnose.RunSelfTest` | 工具列表、执行状态、失败原因、重试、结果 |
| Evidence Chain | event model + FindX Evidence Chain | 采集、告警、日志、Trace、巡检、诊断、任务、回滚全部入链 |

用户侧不能显示来源产品菜单名。入口必须使用 FindX Agent、Agent 管理中心、采集插件、巡检诊断、诊断会话、证据链等 FindX 命名。

## 3. 采集插件状态流门禁

| 状态流 | 成熟源码 | FindX 实现要求 |
| --- | --- | --- |
| Agent 生命周期 | `agent\agent.go` | Metrics、Logs、Prometheus、Ibex 等模块 start/stop/restart 状态必须保留 |
| 指标采集 | `agent\metrics_agent.go`、`inputs\*` | 插件实例、采集周期、标签、错误、采样结果、写入队列 |
| 日志采集 | `agent\logs_agent.go` | file/container/kubernetes/listener、pipeline、auditor、diagnostic、endpoint 状态 |
| 写入端 | `[[writers]]` | remote write、TLS、auth、headers、timeout、批量、队列、失败重试 |
| 心跳 | `[heartbeat]` | 心跳 URL、interval、auth、headers、timeout、最后心跳、脱敏错误 |
| Prometheus scrape | `[prometheus]` | scrape config、WAL、冷启动、配置错误、目标状态 |
| K8s | `k8s\*.yaml` | DaemonSet/Deployment/Sidecar、RBAC、CA/token/kubeconfig、滚动更新 |
| 安装/更新 | `agent\install`、`agent\update` | Linux/Windows/macOS/FreeBSD 差异、service、权限、回滚；非 Linux 平台必须做源码复核和实机/等价环境验证 |
| 远程任务 | `agent\ibex_agent.go`、`[ibex]`、`agent\install`、`agent\update` | 远程任务、安装、更新、回滚必须进入 FindX 任务系统，记录目标、脚本、权限、输出、审计和 Evidence Chain |

成熟源码中已有真实语义的插件启停、测试、状态、配置、更新、日志采集和心跳动作，在 FindX 中不得变成静态展示。

本机脚本安装属于 P5 Agent 阶段规划项，当前不得提前落成静态按钮或半成品接口。后续实现必须复用 FindX Agent 安装计划和内置包仓库，并覆盖以下状态流：

| 入口 | 示例命令 | FindX 状态流要求 |
| --- | --- | --- |
| Linux 本机脚本 | `curl -kfsSL "<FINDX_URL>/api/v1/agent-install/scripts/linux" -o /tmp/findx-agent-install.sh` 后执行 `sudo sh /tmp/findx-agent-install.sh --server "<FINDX_URL>" --token "<TOKEN>" --mode run` | 生成安装计划、校验短期 token、下载内置包、校验 checksum/签名、注册 systemd/进程、写入配置、上报心跳、验证数据到达、失败回滚 |
| Windows CMD 本机脚本 | `certutil -urlcache -f "<FINDX_URL>/api/v1/agent-install/scripts/windows-cmd" "%TEMP%\findx-agent-install.bat"` 后执行脚本 | 校验 token、下载内置包、处理代理/杀软拦截提示、注册 Windows Service、写入配置、上报心跳、验证数据到达、失败回滚 |
| Windows PowerShell 本机脚本 | `Invoke-WebRequest -UseBasicParsing -Uri "<FINDX_URL>/api/v1/agent-install/scripts/windows-powershell" -OutFile <script>` 后以 `ExecutionPolicy Bypass` 执行 | 支持无 GUI 环境、企业代理、权限检查、服务注册、配置下发、脚本输出脱敏、审计入链 |
| 远程安装 | Linux SSH/shell/systemd、Windows WinRM/PowerShell/Service、Kubernetes Helm/Operator/Manifest | 与本机安装共用包仓库、配置模板、安装计划、批次策略、失败阈值、回滚计划、审计和 Evidence Chain |

脚本、包仓库、安装计划和心跳接口进入实现前必须标记 `API_CONTRACT_CHANGE` + `DATA_CHANGE`。文档、测试、日志和页面只能使用 `<TOKEN>`、`<FINDX_URL>`、`<PACKAGE_SHA256>` 等占位符，不得写入真实 token、Cookie、Bearer、SSH key、完整连接串或会话 ID。

## 4. 巡检诊断状态流门禁

| 状态流 | 成熟源码 | FindX 实现要求 |
| --- | --- | --- |
| 插件加载 | `agent\LoadPlugin`、`HandleChangedPlugin`、`HandleNewPlugin` | 插件目录、配置变更、热加载、新增/删除、失败态 |
| 事件模型 | `docs\event-model.md` | event_time、event_status、alert_key、labels、attrs、description 必须结构化入库 |
| 主动巡检 | `inspect <plugin> [target]` | 插件、目标、参数、运行态 OS、报告、记录 |
| 诊断记录 | `diagnose list/show` | 记录列表、详情、保留策略、权限、删除/归档策略 |
| AI 排障 | `chat`、`ChatStream` | 多轮会话、工具调用、shell 权限确认、流式输出、模型状态 |
| 远程 session | `digcore\server\session*` | inspect/diagnose/chat session、stream callback、session input、错误态 |
| WebSocket | `digcore\server\conn.go` | 注册、heartbeat、alert buffer、重连、TLS、鉴权失败退避 |
| 诊断工具 | `plugins\*\diagnose.go`、`sysdiag` | 工具注册、selftest、超时、失败、输出脱敏 |

AI 模型配置必须统一从 FindX `系统配置 / AI 模型配置` 读取。Agent 页面不得保存或回显真实模型密钥。

## 5. API / 数据契约门禁

进入代码实现前，必须先完成以下契约设计并打标：

| 契约 | 标记 | 必须字段 |
| --- | --- | --- |
| 采集插件包 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | package_id、source_type、capability、plugin_name、version、os、arch、config_schema、artifact_ref、checksum、signature_ref、status |
| 插件配置模板 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | template_id、plugin_name、fields_schema、secret_fields、defaults、validation_rules、examples、version、rollback_target |
| 采集配置实例 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | config_id、target_scope、plugin_name、template_id、credential_refs、labels、interval、expected_version、status |
| 下发任务 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | task_id、target_ids、batch_policy、failure_threshold、precheck_result、progress、logs_ref、rollback_plan、audit_id、evidence_chain_id |
| 心跳与数据到达 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | agent_id、host_id、plugin_name、control_heartbeat_at、config_version、last_metric_at、last_log_at、last_error_code、sanitized_error |
| 巡检会话 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | session_id、mode、plugin、target、params、status、stream_ref、report_ref、model_ref、audit_id、evidence_chain_id |
| 诊断记录 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | record_id、source_event_id、plugin、target、status、summary、tools_used、evidence_refs、retention_until |
| 事件入链 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | event_id、event_status、alert_key、labels、attrs、description、source_agent_id、source_plugin、evidence_chain_id |

真实凭据必须只通过 `credential_ref`、`model_ref`、`<TOKEN>`、`<DB_DSN>`、`<API_KEY>` 等引用或占位符表达。API、日志、文档、浏览器错误态不得输出真实 token、Cookie、完整连接串、SSH 私钥、Bearer 值或会话 ID。

## 6. 页面结构门禁

| 页面 | 结构要求 | 真实动作 |
| --- | --- | --- |
| 插件目录 | 分类、图标、说明、适用对象、配置模板、版本、依赖、数据类型、状态 | 查看详情、导入模板、启用/禁用、校验、设为推荐 |
| 配置模板 | 字段 schema、默认值、密文字段引用、示例、版本历史、校验结果 | 新建、编辑、验证、灰度下发、回滚 |
| 下发任务 | 目标、批次、进度、失败原因、日志摘要、审计、证据链 | 暂停、继续、取消、重试、回滚、查看证据 |
| 数据到达 | 指标、日志、心跳、writer、pipeline、OAP/Logs Adapter 状态 | 刷新、筛选、批量修复、查看详情 |
| 巡检插件 | 插件、目标、参数、风险等级、最近执行、状态 | 立即巡检、定时巡检、查看报告 |
| 诊断会话 | 事件/目标、工具调用、流式输出、AI 轮次、报告、证据 | 开始、暂停/取消、重试、保存、入链 |
| 诊断记录 | 列表、筛选、保留策略、摘要、工具、证据 | 查看、归档、关联复盘、导出 |
| 自检 | 工具列表、依赖、权限、状态、错误 | 执行、重试、查看输出 |

所有按钮必须有真实状态流；未实现时标记 `BLOCKED`，不得显示成功态。

## 7. 与链路、日志、CMDB 和 AI SRE 的关系

| 来源 | 联动要求 |
| --- | --- |
| CMDB 主机 | Agent 安装目标、业务组、系统、标签、OS、架构、凭据引用 |
| 链路监控 | 服务/实例探针状态、Trace 数据到达、Profiling 能力包、网关探针 |
| 日志中心 | 日志采集状态、日志关联到达、OTel Collector 连通性、Pipeline 变更 |
| 指标查询 | 采集插件数据到达、remote write 状态、插件错误 |
| 告警事件 | Event 入链，巡检诊断和 AI SRE 可引用事件和采集状态 |
| AI SRE | 只能基于真实插件、采集、心跳、巡检、诊断、日志、Trace 证据输出诊断 |

任一联动链路未补齐时，只能显示 `BLOCKED_AGENT_SUITE_INCOMPLETE` 或对应脱敏阻断状态，不能标完成。

## 8. 品牌与 AI 配置规则

- 用户侧统一使用 FindX Agent / findx-agent 命名。
- 模板含义不能删除；原采集器或巡检工具模板要按 FindX 品牌替换文案、图标、说明和导入路径。
- 外部来源名称只允许保留在源码矩阵、合规登记、归档和开发证据中。
- AI 配置统一进入 `系统配置 / AI 模型配置`。
- 巡检诊断页面只引用 `model_ref`，不保存真实密钥。
- Chat / inspect / diagnose 的工具执行必须有权限确认、超时、审计、输出脱敏和 Evidence Chain。

## 9. MCP、lint、构建验收

编码后必须覆盖：

- FindX 自有登录。
- Agent 管理中心、插件目录、配置模板、下发任务、心跳、数据到达。
- Linux 安装、Windows 安装、Kubernetes DaemonSet/Sidecar/Deployment 接入。
- 插件配置校验、密文字段引用、错误脱敏。
- 指标到达、日志到达、heartbeat 异常、writer 失败、pipeline 失败。
- 主动巡检、诊断会话、AI 排障会话、诊断记录、自检。
- AI 模型配置统一读取系统配置。
- 组件缺失显示 `BLOCKED`，不得伪装成功。
- 菜单、页面标题、权限对象、审计对象均不得显示外部来源品牌。
- 窄屏回归。
- 前端 lint/build、后端测试/build、敏感信息扫描。

## 10. Agent 生命周期完成矩阵

后续实现验收必须逐项标记 Linux、Windows、Kubernetes 的支持状态，不能用单一“Agent 已完成”覆盖所有平台。

| 生命周期阶段 | Linux | Windows | Kubernetes | 完成定义 |
| --- | --- | --- | --- | --- |
| 安装计划 | `REQUIRED` | `REQUIRED` | `REQUIRED` | 目标选择、凭据引用、包版本、配置模板、风险提示、回滚计划 |
| 远程执行 | `REQUIRED` | `REQUIRED` | `REQUIRED` | SSH/shell/systemd、WinRM/PowerShell/Service、RBAC/Helm/Manifest/Webhook 策略 |
| 包校验 | `REQUIRED` | `REQUIRED` | `REQUIRED` | checksum、签名、来源、许可证、版本、架构 |
| 服务注册 | `REQUIRED` | `REQUIRED` | `REQUIRED` | systemd/进程、Windows Service、workload/DaemonSet/Sidecar 状态 |
| 配置下发 | `REQUIRED` | `REQUIRED` | `REQUIRED` | 版本化配置、密文字段引用、灰度批次、失败阈值 |
| 心跳上报 | `REQUIRED` | `REQUIRED` | `REQUIRED` | 控制面心跳、进程状态、配置版本、最后错误 |
| 数据到达 | `REQUIRED` | `REQUIRED` | `REQUIRED` | 指标、日志、Trace/RUM/网关信号按能力包验证 |
| 异常恢复 | `REQUIRED` | `REQUIRED` | `REQUIRED` | 失败重试、断点续传、断连重连、错误脱敏 |
| 升级回滚 | `REQUIRED` | `REQUIRED` | `REQUIRED` | 版本兼容、批次、暂停、回滚、残留检查 |
| 卸载 | `REQUIRED` | `REQUIRED` | `REQUIRED` | 清理范围、服务恢复、残留文件、审计 |
| 审计入链 | `REQUIRED` | `REQUIRED` | `REQUIRED` | 任务、输出、错误、数据到达、回滚、诊断全部进入 Evidence Chain |

## 11. 阻断项

- 只做静态 Agent 包列表：FAIL。
- 只做后台任务，没有前端配置模板、下发、心跳、数据到达、诊断会话：FAIL。
- 插件配置、巡检、诊断、selftest、Pipeline、回滚按钮变成静态展示：FAIL。
- AI 模型配置散落在 Agent 页面或回显密钥：FAIL。
- 用户侧出现外部来源品牌标题、菜单、权限对象、审计对象：FAIL。
- 未用 MCP 浏览器做真实登录点击回归却标 PASS：FAIL。
