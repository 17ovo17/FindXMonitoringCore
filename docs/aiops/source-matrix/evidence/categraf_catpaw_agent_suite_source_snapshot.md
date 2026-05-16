# Categraf / Catpaw 到 FindX Agent Suite 源码检查证据

生成时间：2026-05-07 14:09（UTC+8）
状态：源码证据已检查；运行态 DOM 为 `BLOCKED_RUNTIME_UNAVAILABLE`

## 1. 本地源码路径

事实源目录：

- `D:\项目迁移文件\平台源码\categraf-main (1)\categraf-main`
- `D:\项目迁移文件\平台源码\catpaw-master\catpaw-master`

用户侧命名要求：

- `categraf` 能力进入 FindX Agent 插件、采集配置、采集模板、包仓库和状态检测。
- `catpaw` 能力进入 FindX Agent 巡检诊断、结构化执行、诊断会话和 Evidence Chain。
- 用户侧只能显示 FindX Agent、findx-agent、巡检诊断、诊断会话、证据链等 FindX 命名；外部来源名称只允许保留在本源码矩阵、合规登记、归档和开发证据中。

## 2. FindX Agent 采集插件源码证据

已检查入口：

| 类型 | 路径 | 结论 |
| --- | --- | --- |
| 主入口 | `main.go` | 支持运行、配置目录、POSIX/Windows 条件文件、打包发布 |
| Agent 组合 | `agent\agent.go` | 组合 `MetricsAgent`、`LogsAgent`、`PrometheusAgent`、`IbexAgent`，支持 start/stop/restart |
| 指标采集 | `agent\metrics_agent.go` | 加载 inputs、writer、reader，推送指标 |
| 日志采集 | `agent\logs_agent.go`、`agent\logs_endpoints.go` | 支持文件、listener、container、kubernetes、pipeline、auditor、diagnostic、http/tcp/kafka endpoint |
| Prometheus scrape | `agent\prometheus_agent.go`、`prometheus\*`、`conf\[prometheus]` | 支持 scrape config、WAL、兼容 Prometheus 抓取 |
| 心跳 | `conf\config.toml` `[heartbeat]` | 支持 heartbeat URL、interval、auth、headers、timeout |
| 写入端 | `conf\config.toml` `[[writers]]` | 支持 remote write URL、TLS、auth、headers、timeout、连接池 |
| 插件目录 | `inputs\*`、`conf\input.*\*.toml` | 存在近百类插件和配置模板，覆盖系统、数据库、中间件、容器、K8s、网络、证书等 |
| K8s 部署 | `k8s\daemonset.yaml`、`deployment.yaml`、`sidecar.yaml`、`in_cluster_scrape.yaml` | 支持 DaemonSet、Deployment、Sidecar、in-cluster scrape、CA/token/kubeconfig |
| 服务安装 | `agent\install\service_linux.go`、`service_windows.go`、`service_darwin.go`、`service_freebsd.go` | 存在多平台 service 安装入口；Linux 逻辑较完整，Windows/Darwin/FreeBSD 需实现前复核 |
| 更新 | `agent\update\update_linux.go`、`update_windows.go` | 支持 Linux/Windows 更新入口 |
| 默认配置 | `conf\config.toml` | 包含 global、log、writer、http、ibex、heartbeat、prometheus 等配置 |

插件目录样例：

- `inputs\cpu`
- `inputs\disk`
- `inputs\mem`
- `inputs\mysql`
- `inputs\redis`
- `inputs\nginx`
- `inputs\docker`
- `inputs\kubernetes`
- `inputs\prometheus`
- `inputs\snmp`
- `inputs\x509_cert`
- `inputs\oracle`
- `inputs\postgresql`
- `inputs\clickhouse`
- `inputs\elasticsearch`
- `inputs\mongodb`
- `inputs\kafka`
- `inputs\rabbitmq`

结论：FindX Agent 插件目录不能只做静态卡片，必须保留插件配置模板、输入类型、启停、下发、状态、版本、数据到达验证、回滚和模板说明。

## 3. FindX Agent 巡检诊断源码证据

已检查入口：

| 类型 | 路径 | 结论 |
| --- | --- | --- |
| 主入口 | `main.go` | 支持 `run`、`chat`、`inspect`、`diagnose`、`selftest` 子命令 |
| Agent 生命周期 | `agent\agent.go` | 加载配置、注册插件、启动 server connection、初始化诊断引擎、停止和 reload |
| 插件加载 | `agent\funcs.go`、`agent\agent.go` | 从 `conf.d\p.<plugin>` 加载插件配置，支持新增/变更/删除插件 reload |
| Inspect | `agent\inspect.go`、`main.go` | 支持 `inspect <plugin> [target]` 主动健康巡检 |
| Runner | `agent\runner.go` | 插件 runner 驱动事件采集 |
| 诊断引擎 | `digcore\diagnose\*` | AI 工具调用、诊断记录、chat stream、selftest |
| AI 配置 | `digcore\config\config.go`、`ai_config_test.go`、`conf.d\config.toml` | 支持模型优先级、超时、重试、并发、token 限制、诊断记录保留、OpenAI 兼容和 Bedrock |
| 远程连接 | `digcore\server\conn.go` | WebSocket 注册、heartbeat、alert ring buffer、重连、TLS client、headers |
| 远程会话 | `digcore\server\session.go`、`session_handler.go` | 支持 inspect、diagnose、chat session、stream callback、session input、error |
| 事件模型 | `docs\event-model.md` | Event 包含 event_time、event_status、alert_key、labels、attrs、description |
| 插件开发 | `docs\plugin-development.md` | 插件实现 Instance/Plugin、Init、Gather、diagnose tools、partials |
| 部署 | `docs\deployment.md` | 二进制、systemd、Docker、热加载 |
| 插件目录 | `plugins\*`、`conf.d\p.*\*.toml` | 覆盖系统、网络、存储、安全、日志、Redis、Redis Sentinel 等检查插件 |

巡检/诊断插件样例：

- `plugins\cpu`
- `plugins\mem`
- `plugins\disk`
- `plugins\diskio`
- `plugins\docker`
- `plugins\http`
- `plugins\logfile`
- `plugins\journaltail`
- `plugins\redis`
- `plugins\redis_sentinel`
- `plugins\sysdiag`
- `plugins\systemd`
- `plugins\tcpstate`
- `plugins\cert`
- `plugins\net`
- `plugins\ping`

结论：FindX 巡检诊断不能只写一个 AI 对话框，必须保留事件模型、插件执行、诊断工具、inspect/diagnose/chat/selftest、远程 session、流式输出、诊断记录、权限审计和 Evidence Chain。

## 4. 关键配置证据

FindX Agent 采集插件配置必须覆盖：

- 全局标签、hostname、interval、concurrency。
- writer remote write URL、TLS、auth、headers、timeout。
- heartbeat URL、interval、auth、headers、timeout。
- HTTP exporter。
- Prometheus scrape config 和 WAL。
- 日志采集 endpoints、file/container/kubernetes/listener、pipeline、auditor、processing rules。
- K8s DaemonSet、Deployment、Sidecar 和 scrape 配置。

FindX 巡检诊断配置必须覆盖：

- 全局 interval 和 labels。
- 通知渠道 console/webapi/事件推送。
- AI enabled、model priority、timeout、retry、并发、token 限制、tool timeout、诊断记录保留。
- 模型配置必须迁移到 FindX `系统配置 / AI 模型配置`，不能在 Agent 页面散落真实密钥。
- server WebSocket、TLS、agent id、alert buffer、heartbeat、remote session。
- 插件目录 `conf.d\p.<plugin>`、热加载和 partial 配置。

## 5. 运行态 DOM 状态

当前没有可访问的 FindX Agent Suite / 插件目录 / 巡检诊断运行态页面，因此运行态 DOM 证据为：

- `BLOCKED_RUNTIME_UNAVAILABLE`

后续编码完成前，必须用 MCP 浏览器补齐：

- 包仓库、插件目录、配置模板、导入、签名校验、启用/禁用。
- 采集插件配置下发、灰度、状态、回滚、数据到达验证。
- Linux / Windows / Kubernetes 安装计划和状态。
- 心跳、版本、配置漂移、Agent 日志采集状态。
- 巡检插件列表、主动巡检、诊断会话、流式输出、诊断记录、selftest。
- AI 模型配置从 FindX 系统配置读取，不在页面回显密钥。
- 组件不可用显示 `BLOCKED` 或脱敏 503，不显示假成功。
- 菜单、页面标题、权限对象、审计对象不显示外部来源品牌。

## 6. FindX 替换点

| 来源概念 | FindX 用户侧命名 |
| --- | --- |
| Categraf collector | FindX Agent 采集插件 |
| Categraf config | FindX Agent 采集配置 |
| Categraf heartbeat | FindX Agent 心跳 |
| Catpaw agent | FindX Agent 巡检诊断 |
| Catpaw inspect | 主动巡检 |
| Catpaw diagnose | 诊断会话 |
| Catpaw chat | AI 排障会话 |
| Catpaw selftest | 诊断工具自检 |
| Flashduty / PagerDuty 等通知示例 | FindX 通知通道 / Webhook 引用 |

外部来源名称只允许保留在源码矩阵、合规登记、归档和开发证据中；用户侧菜单、标题、权限对象、审计对象不得显示外部品牌。

## 7. 阻断结论

- `SOURCE_PRESENT`：本地采集插件和巡检诊断源码存在。
- `BLOCKED_RUNTIME_UNAVAILABLE`：没有可访问运行态页面，不能声明 UI 完成。
- `BLOCKED_BY_AGENT_CONTRACT`：包仓库、安装计划、配置模板、下发任务、心跳、数据到达、巡检会话、诊断记录、Evidence Chain 契约尚未实现。
- `BLOCKED_BY_AI_CONFIG_UNIFICATION`：诊断 AI 配置必须统一走 FindX 系统配置 AI 模型配置。
- `BLOCKED_BY_BROWSER_REGRESSION`：编码后必须补 MCP 浏览器真实登录、点击和窄屏回归。
