# FindX P2 AutoOps CMDB 与 Agent 在线同源矩阵

生成时间：2026-05-07 14:04（UTC+8）
状态：`P2-AUTOOPS-CMDB-AGENT-REAL` 编码前门禁，不代表 FindX 已完成实现

## 1. 结论

CMDB、主机、Agent 在线、部署、心跳、终端、上传、命令执行、主机监控必须以 AutoOps/AIOps 源码里的成熟结构为事实源，再统一迁移到 FindX 自有壳层、FindX Agent 命名、权限、审计、Evidence Chain 和统一配置中心。不能把 CMDB 简化成静态资产表，也不能把 Agent 在线简化成一个状态字段。

源码证据：

- [AutoOps CMDB 与 Agent 源码检查证据](evidence/autoops_cmdb_agent_source_snapshot.md)

当前阻断：

- 本地源码存在。
- 未提供可访问 AutoOps/AIOps 运行态 URL，运行态 DOM 证据为 `BLOCKED_RUNTIME_UNAVAILABLE`。
- 编码前必须补 MCP 浏览器真实页面证据。

## 2. 路由与页面矩阵

| FindX 域 | AutoOps 路由 | 源码组件 | FindX 用户侧命名 |
| --- | --- | --- | --- |
| CMDB / Agent 管理 | `/cmdb/ecs` | `web\src\views\cmdb\cmdbHost.vue` | 主机管理 |
| CMDB / Agent 管理 | `/cmdb/group` | `web\src\views\cmdb\cmdbGroup.vue` | 业务分组 |
| CMDB / Agent 管理 | `/cmdb/db` | `web\src\views\cmdb\cmdbDB.vue` | 数据库资产 |
| CMDB / Agent 管理 | `/cmdb/ssh` | `web\src\views\cmdb\Host\SSH.vue` | 终端登录 |
| CMDB / Agent 管理 | `/ops/agent` | `web\src\views\Tools\Agent.vue` | FindX Agent 在线 |
| Agent 管理中心 | `/ops/tools` | `web\src\views\Tools\Tools.vue` | 工具与安装 |
| Agent 管理中心 | 部署管理内部视图 | `web\src\views\Tools\DeployManage.vue` | 部署任务 |

FindX 导航建议：

- 一级：基础设施
- 二级：CMDB / Agent 管理
- 子页：主机管理、业务分组、数据库资产、终端登录、FindX Agent 在线、部署任务

## 3. 主机管理结构要求

源码：

- `web\src\views\cmdb\cmdbHost.vue`
- `web\src\views\cmdb\Host\CmdbGroup.vue`
- `web\src\views\cmdb\Host\CmdbHostTable.vue`

必须保留：

| 区域 | 成熟结构 | FindX 实现要求 |
| --- | --- | --- |
| 左侧分组树 | 搜索、展开、折叠、全部展开、全部折叠、新建、编辑、删除 | 不能改成顶部下拉；树状态必须可持久化或可恢复 |
| 搜索区 | 主机名称、IP 地址、主机状态、搜索、重置 | 与 API 查询参数一致 |
| 新建入口 | 导入主机、Excel 导入、云主机 | 保留下拉语义，不合并成一个静态按钮 |
| 主机表格 | 主机名称、公网/内网 IP、CPU、内存、磁盘、进程、端口、配置、存活状态、认证状态、主机类型、操作 | 列结构和操作语义按源码保留 |
| 操作列 | 详情、编辑、上传、执行命令、删除、监控 | 必须接真实状态流、权限和审计 |
| 分页 | 总数、pageSize、跳页 | 不允许无限滚动替代表格分页 |

## 4. 主机详情与远程能力

源码：

- `web\src\views\cmdb\Host\HostSsh.vue`
- `web\src\views\cmdb\Host\Terminal.vue`
- `web\src\views\cmdb\Host\MonitorDialog.vue`
- `web\src\views\cmdb\Host\ProcessMonitorDialog.vue`
- `web\src\views\cmdb\Host\TcpPortMonitorDialog.vue`

必须保留：

| 能力 | 源码语义 | FindX 增强 |
| --- | --- | --- |
| 主机详情 | 抽屉展示基本信息、连接地址、认证类型、描述、仪表盘 | 接入 Evidence Chain，敏感连接信息脱敏 |
| SSH 终端 | WebSocket 连接主机 | 必须走凭据引用、权限、审计、命令记录脱敏 |
| 文件上传 | 目标路径、文件选择、上传进度 | 文件大小、扩展名、路径安全、审计 |
| 执行命令 | 指定主机执行命令 | 必须有超时、审批/权限、输出脱敏 |
| 主机监控 | CPU、内存、磁盘、进程、端口 | 与指标/日志/Trace 证据关联 |

## 5. FindX Agent 在线结构要求

源码：

- `web\src\views\Tools\Agent.vue`
- `web\src\views\Tools\SelectDeployHost.vue`
- `web\src\api\cmdb.js`

必须保留：

| 区域 | 成熟结构 | FindX 实现要求 |
| --- | --- | --- |
| 搜索区 | 主机名称、Agent 状态、版本、搜索、重置 | 用户侧显示 FindX Agent 状态 |
| 主操作 | 部署 Agent、批量卸载 | 支持多主机选择、确认、任务状态 |
| 表格 | 主机名称、IP、版本、状态、监听端口、安装进度、健康状态、最后心跳、更新时间、操作 | 不允许只显示在线/离线 |
| 状态 | 部署中、部署失败、运行中、启动异常 | 状态值集中定义，并与后端任务状态兼容 |
| 操作 | 重新部署、重启、卸载、删除、查看详情 | 每个动作必须有确认、权限、审计、错误态 |
| 详情弹窗 | 安装路径、端口、进程 ID、健康状态、最后心跳、错误信息 | 错误信息脱敏，不能暴露 token、DSN、路径敏感值 |
| 轮询 | 部署/卸载/重启后刷新和轮询 | 商业实现不能只轮询 15 秒；需要后台任务、SSE/WebSocket 或可恢复轮询 |

## 6. 部署任务结构要求

源码：

- `web\src\views\Tools\DeployManage.vue`
- `web\src\views\Tools\DeployDialog.vue`
- `web\src\views\Tools\DeployProgress.vue`
- `web\src\views\Tools\ServiceMarket.vue`
- `web\src\api\tool.js`

必须保留：

| 区域 | 成熟结构 | FindX 实现要求 |
| --- | --- | --- |
| 服务市场 | 可部署服务、服务详情、版本 | 作为 FindX Agent 能力包、插件包和服务包入口 |
| 部署对话框 | 主机、版本、安装目录、环境变量、自动启动 | 与 FindX Agent 包仓库和配置模板合并 |
| 部署管理 | 服务名称、状态、列表、分页 | 保留部署历史和状态过滤 |
| 日志查看 | 部署进度对话框 | 接入 Evidence Chain 和错误脱敏 |
| 卸载 | 确认后删除部署 | 支持回滚、残留检查、审计 |

## 7. API / 数据契约门禁

进入实现前必须打标：

- `API_CONTRACT_CHANGE`
- `DATA_CHANGE`

需要设计的 FindX 契约：

| 契约 | 必须字段 |
| --- | --- |
| 主机资产 | host_id、name、public_ip、private_ip、vendor、group_id、ssh_credential_ref、status、is_alive、cpu、memory、disk、tags、audit fields |
| 分组树 | group_id、parent_id、name、path、sort、children_count、host_count、permission scope |
| 远程通道 | host_id、credential_ref、protocol、session_id、command_id、timeout、audit_id、sanitized_output |
| 文件上传 | host_id、target_path、file_ref、size、checksum、status、audit_id |
| Agent 记录 | agent_id、host_id、package_id、version、status、install_progress、health、pid、port、last_heartbeat_at、last_data_at、error_code |
| Agent 任务 | task_id、target_hosts、action、batch_policy、status、progress、failure_reason、rollback_plan、evidence_chain_id |
| 部署记录 | deploy_id、service_id、version、host_id、install_dir、env_refs、status、logs_ref、rollback_ref |

## 8. 与 SkyWalking Agent 的合并规则

AutoOps Agent 在线能力与 SkyWalking 多语言探针不是两个互相替代的系统。

FindX 规则：

- AutoOps 源码提供主机、远程安装、部署任务、心跳、状态、终端、上传、命令执行的控制面事实源。
- SkyWalking Agent 矩阵提供 Java、Python、Node.js、PHP、Go、Rust、Ruby、Nginx Lua、Kong、Browser Client JS 等能力包事实源。
- FindX Agent 管理中心必须把两者合并：AutoOps 控制面 + SkyWalking 探针包 + Categraf 插件 + Catpaw 巡检工具。
- 用户侧统一显示 FindX Agent / FindX Browser Agent / FindX 网关探针。
- 不允许用户安装多个互相无关系的 Agent 控制面；安装器可以统一，运行时能力包按语言和场景分开托管。

## 9. 安全与审计门禁

| 场景 | 要求 |
| --- | --- |
| SSH/WinRM/K8s 凭据 | 只能使用凭据引用，不回显真实私钥、密码、token |
| 终端命令 | 记录 actor、host、command_id、开始/结束时间、退出码、脱敏输出 |
| 文件上传 | 校验路径、大小、类型、权限，记录审计 |
| Agent 安装 | 记录包来源、版本、签名、校验和、安装脚本、回滚点 |
| 心跳 | 区分控制面心跳、进程状态、探针加载、数据到达 |
| 错误态 | 下游错误、命令失败、连接失败必须脱敏 |
| 权限态 | 非授权用户不能执行终端、上传、命令、安装、卸载、删除 |

## 10. 运行态与验收阻断

当前状态：

| 项目 | 状态 | 说明 |
| --- | --- | --- |
| AutoOps 源码 | `SOURCE_PRESENT` | 本地源码已读取 |
| CMDB 运行态 DOM | `BLOCKED_RUNTIME_UNAVAILABLE` | 未提供可访问 AutoOps/AIOps 页面 |
| Agent 运行态 DOM | `BLOCKED_RUNTIME_UNAVAILABLE` | 未提供可访问 AutoOps/AIOps 页面 |
| FindX 实现 | `BLOCKED_BY_SOURCE_MATRIX` | API、数据、权限、审计、运行态证据未完成 |

验收必须覆盖：

- 主机分组树：搜索、展开、折叠、新建、编辑、删除。
- 主机列表：搜索、分页、详情、编辑、上传、命令、终端、删除、监控。
- Agent 在线：搜索、部署、批量卸载、重新部署、重启、删除、详情、轮询、心跳异常。
- 部署任务：创建、状态、日志、卸载、失败态、回滚态。
- Linux、Windows、Kubernetes 目标差异。
- MCP 浏览器真实登录和点击回归。
- WSL 前端 build、后端测试、敏感信息扫描。
