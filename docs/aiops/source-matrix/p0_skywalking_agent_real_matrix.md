# FindX P0 SkyWalking Agent 到 FindX Agent 控制面矩阵

生成时间：2026-05-07 15:25（UTC+8）
状态：P0-SKYWALKING-AGENT-REAL 编码前门禁，不代表 FindX Agent 控制面已经完成

## 1. 结论

SkyWalking Agent 生态必须作为 FindX Agent 管理中心的一等能力接入，不是链路监控页面的附属说明，也不是单纯包名清单。FindX 用户侧统一显示 FindX Agent / FindX Browser Agent / FindX 网关探针等命名；外部来源名称只允许留在本源码矩阵、合规来源登记、归档和开发证据中。

当前本地只确认存在链路监控 UI/OAP 源码。本次本地目录检查证据见：

- [SkyWalking Agent 本地源码检查证据](evidence/skywalking_agent_local_source_snapshot.md)

| 本地路径 | 状态 | 作用 |
| --- | --- | --- |
| `D:\平台源码\skywalking-booster-ui-main` | `SOURCE_PRESENT` | 链路监控前端结构、路由、状态流、GraphQL 调用事实源 |
| `D:\平台源码\skywalking-master` | `SOURCE_PRESENT` | OAP/query protocol、WebApp proxy、后端查询语义事实源 |

以下独立 Agent 仓库当前未落地到 `D:\平台源码`。进入 FindX Agent 包仓库、远程安装、配置模板、心跳或数据到达实现前，必须补齐本地源码目录、版本/commit、许可证、NOTICE、制品形态、配置项、安装方式、回滚方式和数据到达验证证据；未补齐时状态保持 `BLOCKED_LOCAL_SOURCE_MISSING`。

## 2. Agent 能力包矩阵

| 能力包 | 上游源码事实源 | 当前本地源码状态 | FindX 用户侧命名 | 制品/安装形态必须记录 | 数据到达验收 |
| --- | --- | --- | --- | --- | --- |
| Java Agent | `https://github.com/apache/skywalking-java` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Java 探针 | jar 包、JVM `-javaagent` 参数、配置文件、插件目录、服务名、实例名、采样、日志关联、TLS/proxy | OAP 连通、Trace 到达、Metric 到达、Log correlation 到达、最后上报时间 |
| Python Agent | `https://github.com/apache/skywalking-python` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Python 探针 | pip/离线包、启动注入、环境变量、框架兼容、进程重启、虚拟环境 | Trace 到达、错误上报、异步框架 span、配置漂移检测 |
| Node.js Agent | `https://github.com/apache/skywalking-nodejs` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Node.js 探针 | npm/离线包、启动参数、框架插件、环境变量、PM2/systemd/Docker 重启 | Trace 到达、错误事件、服务实例识别、版本上报 |
| PHP Agent | `https://github.com/apache/skywalking-php` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent PHP 探针 | 扩展安装、`php.ini`、FPM/Apache 适配、版本矩阵、重载策略 | Trace 到达、FPM 进程状态、扩展加载状态 |
| Go Agent | `https://github.com/apache/skywalking-go` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Go 探针 | 编译/注入方式、模块依赖、构建流水线、版本兼容、回滚策略 | Trace 到达、构建产物版本、运行态接入校验 |
| Rust Agent | `https://github.com/apache/skywalking-rust` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Rust 探针 | crate、运行时、配置项、Trace propagation、服务名、标签 | Trace 到达、服务/实例识别、错误脱敏 |
| Ruby Agent | `https://github.com/apache/skywalking-ruby` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Ruby 探针 | gem、框架兼容、启动注入、配置文件、进程重启 | Trace 到达、错误上报、版本上报 |
| Nginx Lua Agent | `https://github.com/apache/skywalking-nginx-lua` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent 网关探针 | OpenResty/Nginx 版本、Lua 模块、网关配置、灰度重载、回滚脚本 | 网关 Trace 到达、路由标签、配置校验 |
| Kong Agent | `https://github.com/apache/skywalking-kong` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Agent Kong 探针 | Kong 版本、插件安装、服务/路由绑定、重载策略、回滚脚本 | Kong 插件状态、Trace 到达、插件配置版本 |
| Browser Client JS | `https://github.com/apache/skywalking-client-js` | `BLOCKED_LOCAL_SOURCE_MISSING` | FindX Browser Agent | SDK 引入、采样、隐私脱敏、SourceMap、会话/错误/性能事件、CSP 兼容 | RUM 事件到达、错误事件、页面性能、用户隐私脱敏 |

P0 阶段还必须继续审计是否纳入 .NET、C++、Rover/eBPF、SWCK、Satellite、Service Mesh/Sidecar、移动端和小程序接入。若纳入，必须按同一矩阵补齐本地源码、包形态、安装方式、权限、回滚和数据到达验收。

## 3. FindX Agent 控制面要求

SkyWalking Agent 能力进入 FindX Agent 后，控制面必须覆盖完整生命周期：

| 控制面模块 | 必须能力 | 禁止事项 |
| --- | --- | --- |
| 包仓库 | 包名、语言、运行时、平台、架构、版本、commit、许可证、NOTICE、签名、校验和、依赖、兼容范围、离线包 | 只展示静态包名；不记录来源和签名 |
| 安装向导 | 选择主机/服务/Kubernetes workload、选择语言探针、生成安装计划、凭据引用、安装前检测、维护窗口、风险提示 | 直接拼接真实凭据；跳过安装前检测 |
| 远程安装 | Linux SSH/shell/systemd/Docker，Windows WinRM/PowerShell/Service/IIS，Kubernetes Helm/Operator/Webhook/DaemonSet/Sidecar/InitContainer | 只支持单机 shell；不支持 Windows/Linux/K8s 差异 |
| 配置模板 | OAP endpoint、service name、instance name、namespace、environment、采样、插件 include/exclude、日志关联、Trace propagation、TLS/proxy、标签、资源限制 | 把配置做成不可编辑静态说明 |
| 配置下发 | 单机、批量、业务组、namespace/workload、灰度批次、维护窗口、失败阈值、自动回滚、漂移检测 | 无审计地下发；失败后无回滚 |
| 心跳状态 | 控制面心跳、进程/服务状态、探针加载状态、OAP 连通性、最后上报时间、配置版本 | 把 Agent 在线等同于链路数据可用 |
| 数据到达验证 | Trace、Metric、Log correlation、RUM、网关 Trace、Profiling 相关信号按能力包逐项验证 | 未验证数据到达就显示成功 |
| 版本治理 | 语言版本、运行时版本、Agent 版本、插件版本、配置版本、兼容风险、升级建议、已知问题、回滚目标 | 只给“升级”按钮，不展示兼容风险 |
| 审计与证据链 | 包下载、签名校验、安装脚本、远程命令、配置变更、重启、回滚、卸载、数据到达证据进入 Evidence Chain | 日志输出真实 token、完整 DSN、Cookie 或会话 ID |

## 4. API / 数据契约门禁

进入代码实现前，必须先完成以下契约设计并打标：

| 契约 | 标记 | 必须字段 |
| --- | --- | --- |
| Agent 包仓库 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | package_id、capability_type、language、runtime、os、arch、version、commit、source_url、license_ref、artifact_ref、signature_ref、checksum、compatibility、status |
| 安装计划 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | target_scope、hosts/workloads、package_id、config_template_id、credential_ref、schedule_window、precheck_result、rollback_plan、approval_state |
| 配置模板 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | template_id、package_id、fields_schema、secret_fields、defaults、validation_rules、version、rollback_target |
| 下发任务 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | task_id、target_ids、expected_config_version、batch_policy、failure_threshold、status、audit_id、evidence_chain_id |
| 心跳与状态 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | agent_id、host_id、service_id、process_status、control_heartbeat_at、probe_loaded、oap_reachable、last_trace_at、last_metric_at、last_log_link_at |
| 数据到达验证 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | verification_id、capability_type、sample_trace_id、sample_metric_ref、sample_log_ref、rum_event_ref、status、error_code、sanitized_error |
| 卸载与回滚 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | uninstall_plan、rollback_version、cleanup_scope、residual_files、service_restart, audit_id、evidence_chain_id |

真实凭据必须只通过 `credential_ref`、`<TOKEN>`、`<DB_DSN>`、`<API_KEY>` 等占位符或引用表达。API、日志、文档、浏览器错误态不得输出真实 token、Cookie、完整连接串、SSH 私钥、Bearer 值或会话 ID。

## 5. 前端页面结构门禁

FindX Agent 管理中心必须有真实页面和状态流，不允许只写后台任务：

| 页面 | 结构要求 | 真实动作 |
| --- | --- | --- |
| 包仓库 | 分类、语言、运行时、平台、架构、版本、来源、签名状态、兼容风险、操作列 | 查看详情、导入包、校验签名、启用/禁用、设为推荐、删除/归档 |
| 安装向导 | 选择目标、选择能力包、配置模板、安装前检测、执行计划、确认、进度 | 生成计划、precheck、提交任务、轮询状态、失败重试 |
| 配置模板 | 字段 schema、默认值、密文字段引用、校验结果、版本历史 | 新建、编辑、验证、灰度下发、回滚 |
| 下发任务 | 批次、目标、状态、失败原因、日志摘要、审计引用 | 暂停、继续、取消、重试、回滚、查看证据 |
| 心跳与覆盖率 | 主机、服务、语言、版本、控制面心跳、探针加载、OAP 连通、数据到达 | 筛选、刷新、批量修复、查看详情 |
| 漂移检测 | 期望配置、实际配置、差异、风险、修复建议 | 重新下发、忽略、创建变更单 |
| 升级回滚 | 当前版本、目标版本、兼容风险、批次策略、回滚点 | 升级、灰度、暂停、回滚、验证 |
| 卸载记录 | 卸载计划、清理范围、残留文件、服务恢复、审计 | 确认卸载、检查残留、回滚 |

所有按钮在成熟产品或商业 Agent 控制面中有真实语义的，FindX 中必须接入真实状态流；未完成时标记 `BLOCKED`，不得显示为可用成功态。

## 6. 与链路监控前端的关系

链路监控页面和 Agent 管理中心是两个域，但必须打通：

| 来源页面 | 目标页面 | 关联要求 |
| --- | --- | --- |
| 链路监控服务目录 | Agent 覆盖率详情 | 根据 service/instance 看到对应语言探针、版本、心跳和最后 Trace |
| Trace 详情 | Agent 数据到达验证 | Trace 缺失或断点异常时能定位探针加载、采样、OAP 连通和配置版本 |
| 拓扑 | 网关/服务探针状态 | Nginx Lua/Kong/应用探针的节点状态可解释拓扑缺口 |
| 链路告警 | 安装/配置任务证据 | 告警复盘能引用安装、升级、配置下发、回滚、数据到达证据 |
| AI SRE | Evidence Chain | AI 诊断只能引用真实 Agent 状态和数据到达证据，缺失时返回数据缺失 |

## 7. BLOCKED 与验收规则

当前状态：

| 项目 | 状态 | 阻断原因 |
| --- | --- | --- |
| SkyWalking UI/OAP 源码 | `SOURCE_PRESENT` | 仍需运行态 DOM 和 APM Adapter 契约 |
| 多语言 Agent 独立源码 | `BLOCKED_LOCAL_SOURCE_MISSING` | `D:\平台源码` 下未发现 `skywalking-java`、`skywalking-python`、`skywalking-nodejs`、`skywalking-php`、`skywalking-go`、`skywalking-rust`、`skywalking-ruby`、`skywalking-nginx-lua`、`skywalking-kong`、`skywalking-client-js` |
| FindX Agent 包实现 | `BLOCKED_BY_SOURCE_MATRIX` | 包仓库/安装计划/配置模板/心跳/数据到达/升级回滚契约未完成 |

P5 完成前必须用 MCP 浏览器覆盖：

- 包仓库列表、详情、导入、签名校验、兼容风险。
- Linux 安装计划、配置下发、心跳异常、数据未到达 `BLOCKED`、升级回滚、卸载确认。
- Windows 安装计划、PowerShell/Service/IIS 相关配置、心跳异常、回滚。
- Kubernetes 安装计划、namespace/workload 选择、Webhook/DaemonSet/Sidecar 策略、RBAC/TLS 错误态。
- Browser Agent 接入向导、隐私脱敏、RUM 数据到达。
- 链路监控服务目录、Trace 详情、拓扑、告警与 Agent 状态跳转。

未补齐源码、契约、页面结构、MCP 浏览器真实回归和敏感信息扫描前，不允许把 SkyWalking Agent 到 FindX Agent 控制面声明为完成。
