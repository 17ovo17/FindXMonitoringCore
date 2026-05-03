# 夜莺深度集成与 findx-agents 立项开发计划

生成时间：2026-05-04 01:30（UTC+8）

## 1. 立项结论

AI WorkBench 后续应把成熟监控能力深度交给 Nightingale（夜莺）和 Categraf，平台自身只保留差异化入口、AI 问诊、诊断报告、知识沉淀、Hermes/消息入口和 findx-agents 存活视图。

总体定位：

- Nightingale：监控事实源，负责目标资产、业务组、数据源、告警规则、告警事件、订阅、静默、通知、仪表盘、记录规则、指标字典、集成模板、事件流水线和监控运维能力。
- Categraf：采集底座，负责主机、数据库、中间件、Kubernetes、日志、黑盒拨测、硬件与云服务等成熟采集能力。
- findx-agents：平台自有探针品牌和安装/存活入口，基于 Categraf 做产品化包装，不直接从 Catpaw 派生，不 fork Categraf 采集插件。
- AI WorkBench：统一品牌入口、AI 问诊、证据编排、诊断报告、模板推荐、知识沉淀、Hermes/消息入口、findx-agents 存活视图和 Nightingale/Categraf 的受控聚合入口。

核心原则：除 AI 问诊和 findx-agents 存活视图外，监控、告警、仪表盘、通知、规则、事件流水线、指标字典和采集模板一律优先复用 Nightingale/Categraf 的成熟能力。FindX 只做编排、同步、品牌化入口和 AI 增强，不建设第二套监控事实源。

## 2. 源码审计范围与关键证据

本次审计以用户提供的 `D:\平台源码`、当前 git 工作区 `D:\ai-workbench` 和 WSL 运行项目 `/opt/ai-workbench` 为准。

| 组件 | 本地路径 | 审计结论 |
| --- | --- | --- |
| Nightingale | `D:\平台源码\nightingale-main (1)\nightingale-main` | 本地 LICENSE 为 Apache-2.0。源码具备 targets、busi groups、datasources、alert rules/events、dashboards、notify、message templates、event pipelines、embedded products、builtin components/payloads、AI assistant、integrations 模板。适合作为监控事实源和模板事实源。 |
| Categraf | `D:\平台源码\categraf-main (1)\categraf-main` | 本地 LICENSE 为 MIT。源码具备 remote write、N9E heartbeat、local/http providers、92 个 `conf/input.*` 模板目录、105 个 toml 配置文件和大量 inputs 插件。适合作为 findx-agents 采集底座。 |
| Catpaw | `D:\平台源码\catpaw-master\catpaw-master` | 本地 LICENSE 为 AGPL-3.0。具备插件巡检、AI 诊断、AI chat、server WebSocket 和告警上报。可作为思路参考，不建议直接派生为 findx-agents。 |
| AutoOps | `D:\平台源码\AutoOps-main\AutoOps-main` | 有 Agent 部署、CMDB、任务、K8s、工具市场思路；旧 Agent 部署存在硬编码 token、临时编译和重复自研问题，不建议作为新探针底座。 |
| AI WorkBench | `/opt/ai-workbench` | 当前已有 AI 问诊、Prometheus、Catpaw、N9E Redis/MySQL 读取、拓扑、工作流、知识库。需要把监控/告警/模板事实源迁移到 Nightingale，并把 Catpaw 命名迁移到 findx-agents。 |

Nightingale 模板证据：

- `integrations/` 下有 65 个集成目录。
- 文件类型统计：178 个 `.json`、76 个 `.png`、70 个 `.toml`、64 个 `.md`、1 个 `.svg`、1 个 `.example`。
- 模板类型统计：42 个告警规则 JSON、118 个 dashboard JSON、16 个 metrics JSON、2 个 record-rules JSON、64 个 markdown 文档、70 个 Categraf toml。
- 典型模板包括 Linux、Kubernetes、MySQL、Redis、ClickHouse、Kafka、Elasticsearch、Nginx、HTTP_Response、Net_Response、Ping、Procstat、Oracle、PostgreSQL 等。

Nightingale API/路由证据：

- `center/router/router.go` 暴露 `/metrics/desc`、`/metric-views`、`/builtin-metric-filters`、`/builtin-metrics`、`/builtin-metric-promql`。
- `center/router/router.go` 暴露 `/integrations/icon/:cate/:name`、`/boards`、`/board/:bid`、`/board/:bid/pure`、`/embedded-dashboards`。
- `center/router/router.go` 暴露 `/notify-tpls`、`/message-templates`、`/notify-channel-configs`、`/event-pipelines`、`/event-pipeline/:id/trigger`、`/event-pipeline/:id/stream`。
- `center/router/router.go` 暴露 service API：`/v1/n9e/builtin-components`、`/v1/n9e/builtin-payloads`、`/v1/n9e/message-templates`、`/v1/n9e/event-pipelines`、`/v1/n9e/notify-channels`。
- `center/router/router_builtin.go` 从 `integrations` 读取 dashboards、alerts、icons、markdown。
- `center/integration/init.go` 从 `integrations` 初始化 builtin components、alerts、dashboards、metrics，并维护 `BuiltinPayloadInFile`。
- `center/router/router_builtin_payload.go` 支持 payload 类型 `alert`、`dashboard`、`collect`，其中 `collect` 会按 TOML 校验。
- `doc/api/event-pipeline.md` 明确 event pipeline 的列表、详情、创建、更新、删除、tryrun、trigger、stream、executions 和 service API。
- `doc/api/embedded-product.md` 明确第三方产品嵌入入口，可用于 Nightingale 与 FindX 的双向入口整合。

Categraf 证据：

- `conf/config.toml` 默认 `providers = ["local"]`，提供 `[global.labels]`、`[[writers]]`、`[heartbeat]`、`[ibex]`、`[prometheus]`。
- `conf/config.toml` 默认 writers 指向 `/prometheus/v1/write`，heartbeat 指向 `/v1/n9e/heartbeat`。
- `heartbeat/heartbeat.go` 上报 `agent_version`、`os`、`arch`、`hostname`、`cpu_num`、`cpu_util`、`mem_util`、`global_labels`、`host_ip`、`unixtime`。
- `inputs/provider_manager.go` 支持 `local` 与 `http` provider。
- `inputs/http_provider.go` 支持远端配置拉取、version/hash、按 global labels 与 agent hostname 查询、动态 reload。
- `inputs/local_provider.go` 按 `input.*` 目录加载 `.yaml/.yml/.json/.toml`。

AI WorkBench 现状证据：

- `api/main.go` 当前已有 `/api/v1/catpaw/*`、`/api/v1/remote/install-catpaw`、`/api/v1/remote/uninstall-catpaw`、`/api/v1/prometheus/*`、`/api/v1/n9e/agents`、`/api/v1/n9e/alerts`。
- `api/internal/handler/n9e_agents.go` 当前直接读取 Redis `n9e_meta_*`。
- `api/internal/handler/n9e_alerts.go` 当前直接查询 MySQL `alert_cur_event`。
- `web/src/views/CatpawInstall.vue`、`web/src/views/CatpawChatPanel.vue`、`web/src/views/Diagnose.vue`、`web/src/views/TopologyHub.vue` 仍存在 Catpaw 命名和入口。

目录中的 Dify、FastGPT、RAGFlow 属于 AI/RAG 参考项目，不是本次“夜莺 + findx-agents”监控集成主线。其中 Dify/FastGPT 本地 LICENSE 带有额外商业条件，RAGFlow 为 Apache-2.0；后续若要引入其代码或 UI，需要单独立项做许可证、产品边界和安全评估。本轮不再把它们作为 Nightingale 集成扣分项。

## 3. 许可证与命名策略

### 3.1 许可证事实

- Nightingale 本地源码 LICENSE 为 Apache License 2.0。
- Categraf 本地源码 LICENSE 为 MIT。
- Catpaw 本地源码 LICENSE 为 GNU AGPL-3.0。

### 3.2 许可证影响

Nightingale 和 Categraf 适合做深度集成、二次包装和产品化组合，但仍需保留开源许可证、版权声明、NOTICE/THIRD_PARTY 文件和修改说明。

Catpaw 是 AGPL-3.0。如果直接复制、派生或重命名为 findx-agents，后续网络服务分发、商业闭源和私有化交付会带来更强的开源义务与合规风险。因此 findx-agents 不应以 Catpaw 代码派生作为主路线。

### 3.3 命名策略

产品与 UI 命名：

- 探针统一命名为 `findx-agents`。
- 平台内监控入口可命名为 `FindX Monitor`、`FindX Observability`、`FindX 告警中心`、`FindX 仪表盘`、`FindX 模板中心`。
- 用户可见的探针服务名建议为 `findx-agents.service`。
- 安装路径建议为 `/opt/findx-agents`。
- 配置路径建议为 `/etc/findx-agents`。
- 日志路径建议为 `/var/log/findx-agents`。

合规与工程命名：

- 源码和文档保留 Nightingale/Categraf 的开源归属说明。
- 内部 connector 可使用 `nightingale`、`n9e`、`categraf` 作为技术标识。
- 不在产品主导航里暴露 Catpaw；旧 Catpaw 入口后续迁移为 findx-agents 兼容页。

## 4. 目标与非目标

### 4.1 项目目标

1. Nightingale 成为监控事实源。
2. Nightingale integrations/builtin components/builtin payloads 成为模板事实源。
3. Categraf 成为 findx-agents 的采集底座。
4. AI WorkBench 保留 AI 问诊、诊断报告、知识沉淀和探针存活视图。
5. 平台前端继续使用 FindX 自有命名和体验，不要求用户直接理解 Nightingale 内部对象。
6. 所有监控对象、告警事件、规则、仪表盘、通知、事件流水线、指标字典尽量走 Nightingale API 或嵌入能力。
7. AI 问诊从 Nightingale 获取目标、告警、指标、规则、仪表盘链接、记录规则、模板说明、markdown runbook、事件流水线执行上下文。
8. 旧 Catpaw 相关命名、API 和页面逐步迁移到 findx-agents。

### 4.2 非目标

本项目不做：

- 重写 Nightingale 告警引擎。
- 重写 Nightingale 仪表盘。
- 重写 Nightingale 通知渠道。
- 重写 Nightingale 事件流水线引擎。
- 重写 Categraf 采集插件。
- 直接从 Catpaw 派生新探针。
- 第一阶段支持自动远程修复。
- 第一阶段支持任意远程命令、任意 SQL、任意 PromQL、任意 HTTP 请求。
- 直接修改 Nightingale 前端源码作为主路线。

## 5. 目标架构

```text
用户 / Hermes / Web
        |
        v
AI WorkBench / FindX 入口层
  - AI 问诊
  - findx-agents 存活视图
  - FindX 模板中心
  - Nightingale API 聚合与品牌化包装
  - Dashboard / embedded product 受控嵌入
  - 诊断报告与知识库
        |
        +----------------------------+--------------------------+
        |                            |                          |
        v                            v                          v
Nightingale                      AI 诊断引擎              Template Catalog
  - Targets                        - LLM                    - sync metadata
  - Busi Groups                    - evidence chain         - hash/version
  - Datasources                    - metric explain         - license/source
  - Alert Rules                    - runbook recommend      - install state
  - Events                         - template recommend
  - Dashboards
  - Metrics Desc
  - Builtin Payloads
  - Notify / Event Pipelines
        ^
        |
findx-agents
  - Categraf binary/config
  - native N9E heartbeat
  - remote write
  - local/http input provider
  - optional FindX alive heartbeat
```

## 6. 能力归属矩阵

| 能力 | 归属 | AI WorkBench 做什么 |
| --- | --- | --- |
| AI 问诊 | AI WorkBench | 保留主入口、诊断编排、报告和知识沉淀。 |
| 探针存活 | AI WorkBench + Nightingale | 自有 findx-agents 存活视图；同时读取 Nightingale target heartbeat。 |
| 主机/目标资产 | Nightingale | 只做品牌化列表、搜索、跳转、诊断入口。 |
| 指标采集 | Categraf | 生成 findx-agents 安装包和配置模板，不 fork inputs。 |
| 告警规则 | Nightingale | 提供模板安装、封装入口或跳转，不重写规则引擎。 |
| 活跃/历史告警 | Nightingale | 拉取事件作为 AI 问诊证据和前端摘要。 |
| 仪表盘 | Nightingale | 嵌入 `/board/:bid/pure`、使用 embedded dashboards 或创建品牌化跳转。 |
| 指标字典 | Nightingale | 同步 `/metrics/desc` 和 builtin metrics，为 AI 解释指标含义。 |
| 记录规则 | Nightingale | 只做模板展示、安装入口和 AI 推荐，不自建执行器。 |
| 通知渠道/模板 | Nightingale | 复用 notify tpl、message templates、notify rules、notify channels。 |
| 事件流水线 | Nightingale | 复用 event pipelines；AI 问诊可作为只读触发源或 webhook 输出。 |
| 模板中心 | FindX 编排 + Nightingale/Categraf 来源 | 做索引、筛选、预览、同步、安装编排、版本/hash/归属管理。 |
| 远程任务/Ibex | Nightingale/Categraf | 后续评估；第一阶段不开放写操作。 |
| Catpaw 插件巡检 | Deferred | 后续按 findx-agents 插件或专项迁移，不进第一阶段。 |

## 7. FindX 模板中心：深度模板集成方案

用户要求“模板那些全都要”，因此 FindX 模板中心需要覆盖 Nightingale 与 Categraf 的全量成熟模板能力，但不复制建设执行引擎。

### 7.1 模板来源

1. Nightingale integrations 文件源：
   - `alerts/*.json`
   - `dashboards/*.json`
   - `metrics/*.json`
   - `record-rules/*.json`
   - `markdown/*.md`
   - `icon/*`
   - `*.toml`
2. Nightingale builtin API 源：
   - `/api/n9e/builtin-components`
   - `/api/n9e/builtin-payloads`
   - `/api/n9e/builtin-payloads/cates`
   - `/api/n9e/builtin-metrics`
   - `/api/n9e/builtin-metric-promql`
   - `/api/n9e/metrics/desc`
3. Nightingale notification/event 源：
   - `/api/n9e/notify-tpls`
   - `/api/n9e/message-templates`
   - `/api/n9e/notify-channel-configs`
   - `/api/n9e/notify-rules`
   - `/api/n9e/event-pipelines`
4. Nightingale dashboard/embed 源：
   - `/api/n9e/boards`
   - `/api/n9e/board/:bid`
   - `/api/n9e/board/:bid/pure`
   - `/api/n9e/embedded-dashboards`
   - `/api/n9e/embedded-product`
5. Categraf 采集模板源：
   - `conf/config.toml`
   - `conf/input.*/*.toml`
   - `inputs/local_provider.go`
   - `inputs/http_provider.go`

### 7.2 模板类型

| 类型 | 上游来源 | FindX 展示/使用方式 | 是否自建执行 |
| --- | --- | --- | --- |
| integration component | Nightingale integrations / builtin-components | 模板目录、图标、说明、组件筛选 | 否 |
| dashboard payload | Nightingale dashboards / builtin-payloads | 预览、安装、嵌入、跳转 | 否 |
| alert payload | Nightingale alerts / builtin-payloads | 预览、安装、AI 推荐 | 否 |
| collect payload | Nightingale collect payload / Categraf toml | 生成 findx-agents input 配置 | 否 |
| builtin metrics | Nightingale metrics JSON / builtin-metrics | 指标字典、PromQL 推荐、诊断解释 | 否 |
| metrics desc | Nightingale `etc/metrics.yaml` / `/metrics/desc` | AI 报告中的指标含义、单位、阈值解释 | 否 |
| record rules | Nightingale record-rules | 预览、安装、依赖提示 | 否 |
| markdown/runbook | Nightingale markdown | 模板说明、安装说明、AI 引用材料 | 否 |
| icon/image | Nightingale icon/markdown assets | 模板市场展示 | 否 |
| notify template | Nightingale notify-tpls/message-templates | 通知内容预览和复用 | 否 |
| notify channel/rule | Nightingale notify channel/rule | 只做入口、同步摘要和跳转 | 否 |
| event pipeline | Nightingale event-pipelines | 触发、试跑、执行记录读取、AI 回调 | 否 |
| embedded product | Nightingale embedded-product | 双向入口嵌入 | 否 |

### 7.3 模板目录模型

FindX 本地只保存索引和同步状态，不保存第二套事实源。建议新增只读/缓存型元数据结构：

```text
findx_template_catalog
  id
  source                 # nightingale_file / nightingale_api / categraf_file / findx_overlay
  source_path_or_url
  upstream_type          # dashboard / alert / collect / metric / record_rule / markdown / notify / event_pipeline
  component_ident
  component_name
  cate
  payload_type
  name
  tags
  description
  content_hash
  upstream_uuid
  upstream_version
  upstream_license
  installed_status       # not_installed / installed / drifted / conflict / unsupported
  installed_ref          # Nightingale object id or ident
  dependencies           # datasource, labels, collector, Categraf input
  compatibility          # Nightingale version, Categraf version, OS/arch
  last_sync_at
  last_install_at
```

约束：

- `content_hash` 是幂等和漂移判断的核心，不靠名称猜测。
- `upstream_license` 必须保留 Nightingale/Categraf 归属。
- `installed_status` 只描述安装状态，不成为规则、dashboard、通知的事实源。
- 写入 Nightingale 必须走 Nightingale API，不直接写 MySQL/Redis。
- 模板内容中的凭据、URL、token、DSN 必须用占位符或变量，不进入日志和前端 raw error。

### 7.4 同步流程

1. Scan：扫描 Nightingale integrations、Categraf conf/input、Nightingale API。
2. Normalize：统一为 component、payload、asset、dependency 四类元数据。
3. Validate：校验 JSON/TOML、dashboard schema、alert rule schema、collect toml、event pipeline schema。
4. Hash：按规范化内容生成 hash，忽略创建时间、更新人等非语义字段。
5. Store：写入 FindX 模板索引缓存。
6. Diff：对比 Nightingale 已安装对象，标记 installed/drifted/conflict。
7. Preview：展示 dashboard、alert、metrics、record rule、collect toml、markdown。
8. Install：通过 Nightingale API 安装或更新。
9. Rollback：保留安装前对象引用和 hash，支持回滚到上一版本或删除本次安装对象。

### 7.5 安装策略

- dashboard：优先走 Nightingale board/builtin payload 能力安装，再用 `/board/:bid/pure` 或 embedded dashboards 嵌入。
- alert：走 Nightingale alert rule/builtin payload 能力安装，保留业务组、数据源、标签变量确认步骤。
- collect：生成 findx-agents 的 Categraf `input.*` 配置，由本地 provider 或 HTTP provider 下发。
- metric desc：同步为只读指标字典，用于 AI 解释，不反向覆盖夜莺默认字典。
- record rule：走 Nightingale 规则能力安装，必须校验 datasource 与 PromQL。
- notify/message template：走 Nightingale message templates / notify templates API，默认只读预览，P1 后允许管理员安装。
- event pipeline：走 Nightingale event pipeline API，P1 只读与试跑，P2 才允许管理员创建/更新。
- embedded product：用于把 FindX AI 问诊入口注册到 Nightingale，也可把 Nightingale dashboard 作为 FindX 嵌入入口。

## 8. Categraf 与 findx-agents 深度包装方案

### 8.1 包装原则

findx-agents 不是新采集器，而是 Categraf 的产品化包装：

- 不 fork Categraf inputs。
- 不改 Categraf remote write 协议。
- 不改 Categraf N9E heartbeat 语义。
- 不改 Categraf input TOML 主结构。
- 只包装安装目录、服务名、默认配置、模板生成、升级卸载、存活状态和 FindX 品牌入口。

### 8.2 配置模型

建议路径：

- 安装路径：`/opt/findx-agents`
- 配置路径：`/etc/findx-agents`
- 采集模板：`/etc/findx-agents/input.*/*.toml`
- 日志路径：`/var/log/findx-agents`
- systemd：`findx-agents.service`

默认配置：

- `[global] providers = ["local"]`，第一阶段使用本地配置。
- `[global.labels]` 注入 `env`、`region`、`tenant`、`business`、`agent_id` 等非敏感标签。
- `[[writers]]` 指向 Nightingale remote write 入口。
- `[heartbeat]` 指向 Nightingale `/v1/n9e/heartbeat`。
- 可选 FindX alive heartbeat 只上报平台存活需要的最小字段。

P1 可开启 Categraf `http_provider`：

- FindX 提供 `/api/v1/findx/agents/configs` 远程配置接口。
- Categraf 按 version/hash 拉取配置。
- 接口按 agent hostname、global labels、agent_id 返回差异化 input 配置。
- 远程配置接口必须认证、限流、审计、脱敏，并禁止返回任意命令。

### 8.3 模板分层

| 层级 | 示例 | 用途 |
| --- | --- | --- |
| base profile | linux_base | CPU、内存、磁盘、网络、进程、systemd、self_metrics |
| middleware profile | mysql、redis、nginx、kafka、clickhouse | 中间件采集模板 |
| kubernetes profile | kubernetes、cadvisor、prometheus | K8s 与容器采集模板 |
| blackbox profile | ping、http_response、net_response、dns_query、x509_cert | 黑盒探测与证书探测 |
| hardware profile | ipmi、smart、snmp、nvidia_smi、dcgm | 硬件/GPU/网络设备 |
| custom profile | prometheus、exec、procstat、filecount | 受控扩展，默认需要管理员确认 |

### 8.4 与 Nightingale integrations 对齐

FindX 模板中心需要把 Categraf input 模板与 Nightingale integration 关联：

- `Linux` integration 关联 `input.cpu`、`input.mem`、`input.disk`、`input.diskio`、`input.net`、`input.netstat`、`input.processes`、`input.systemd`。
- `MySQL` integration 关联 `input.mysql`，并关联 MySQL dashboards/alerts/metrics。
- `Redis` integration 关联 `input.redis` 与 Redis dashboards/alerts。
- `Nginx` integration 关联 `input.nginx` 与 Nginx dashboards/metrics。
- `Kubernetes` integration 关联 `input.kubernetes`、`input.cadvisor`、`input.prometheus` 与 record rules。
- `HTTP_Response`、`Net_Response`、`Ping` 分别关联黑盒拨测 input 与对应告警/dashboard。

这样用户在 FindX 里选择“安装 MySQL 监控模板”时，实际动作是：

1. 生成 findx-agents 的 `input.mysql/mysql.toml`。
2. 安装 Nightingale MySQL dashboard。
3. 安装 Nightingale MySQL alert rules。
4. 同步 MySQL metrics desc/builtin metrics。
5. 在诊断报告里把 MySQL runbook/markdown 作为引用材料。

## 9. Nightingale Dashboard 与嵌入 PoC 方案

### 9.1 嵌入路径

优先级从高到低：

1. `/board/:bid/pure`：用于 FindX 前端 iframe 或后端代理嵌入，适合第一轮 PoC。
2. `/embedded-dashboards`：用于受控配置已嵌入 dashboard。
3. `/embedded-product`：用于把 FindX AI 问诊入口注册到 Nightingale，或把 Nightingale 作为 FindX 的外部入口。
4. 直接 board API 渲染：仅作为后续深度品牌化方案，第一阶段不建议重写渲染器。

### 9.2 P0/P1 PoC 验收

PoC 选择 Linux integration：

- 导入或定位 `Linux/dashboards/categraf-overview.json`。
- 通过 Nightingale board 能力生成 dashboard。
- FindX 页面展示 dashboard iframe 或安全代理链接。
- Dashboard 中 target、datasource、业务组权限正确。
- Nightingale 断连时 FindX 显示 503/不可用状态，不白屏，不泄露内部 URL、token、DSN。
- 不修改 Nightingale 前端源码。

## 10. AI 问诊如何使用模板

AI 问诊不只读告警事件，还要把 Nightingale 模板体系纳入证据链。

### 10.1 证据链输入

从告警触发 AI 问诊时，NightingaleEvidenceProvider 应收集：

- alert event：告警名、级别、触发值、触发时间、tags、target ident。
- alert rule：规则名称、PromQL、阈值、持续时间、通知规则、规则备注。
- target：主机、业务组、labels、agent heartbeat、Categraf 版本。
- datasource：数据源状态、查询入口、类型。
- dashboard：关联 component 的 dashboard 链接。
- metric desc：指标中文/英文描述、单位、类型、推荐 PromQL。
- record rules：相关预聚合规则。
- integration markdown：上游模板说明、安装说明、排查建议。
- notify context：通知模板、通知规则、通知渠道摘要。
- event pipeline：已触发或可触发的流水线、最近执行状态。

### 10.2 报告输出

AI 诊断报告建议固定结构：

1. 结论摘要：根因、置信度、影响范围。
2. 告警证据：事件、规则、PromQL、触发值。
3. 指标解释：引用 Nightingale metrics desc/builtin metrics。
4. 图表入口：关联 dashboard/pure board 链接。
5. 运行手册：引用 integration markdown/runbook。
6. 建议动作：只读建议；涉及写操作时要求二次确认。
7. 验证步骤：推荐 PromQL、dashboard 面板、恢复判定。
8. 模板建议：建议安装的 dashboard、alert、collect、record-rule 模板。

### 10.3 安全边界

- AI 只能推荐模板安装，不能默认直接写 Nightingale。
- AI 不能执行任意 PromQL，只能使用白名单模板或受控查询。
- AI 不能展示内部连接串、完整 token、cookie、SSH key、数据库 DSN。
- AI 回写 Nightingale 事件备注或 event pipeline 输出时，必须脱敏并记录审计。

## 11. 分阶段开发计划

### Phase 0：立项、合规和模板基线

目标：冻结技术路线、许可证边界、命名边界、模板边界和迁移边界。

工作项：

- 明确 Nightingale 为监控事实源和模板事实源。
- 明确 Categraf 为 findx-agents 采集底座。
- 明确 Catpaw 不作为 findx-agents 派生源码。
- 新增开源组件 NOTICE/THIRD_PARTY 说明。
- 定义产品命名、内部命名和兼容命名。
- 定义禁用重复建设清单：告警、仪表盘、通知、事件流水线、采集插件、目标资产。
- 建立 FindX 模板中心元数据模型和同步任务设计。

验收：

- 文档确认能力归属矩阵。
- 产品命名中出现 `findx-agents`。
- 合规说明覆盖 Nightingale/Categraf/Catpaw。
- 模板中心覆盖 integrations、dashboards、alerts、metrics、record-rules、collect toml、notify templates、message templates、event pipelines、embedded dashboards。

### Phase 1：Nightingale Connector 与模板只读目录

目标：AI WorkBench 通过受控 connector 访问 Nightingale，不再直接依赖 Nightingale MySQL/Redis 表结构。

工作项：

- 新增 `api/internal/nightingale/` connector 包。
- 支持 Nightingale service API base URL、BasicAuth/token、timeout、retry、TLS 配置。
- 封装 targets、target tags、busi groups、datasources、alert-cur-events、alert-his-events、dashboards、recording-rules、builtin-components、builtin-payloads、builtin-metrics、message-templates、event-pipelines。
- 替换当前直接读 `n9e.redis_*` 和 `n9e.mysql_dsn` 的列表接口，逐步切到 Nightingale API。
- 新增 FindX 模板中心只读 API：列表、详情、筛选、预览、hash、安装状态。
- 所有 Nightingale 错误统一脱敏，不返回内部连接串。
- 加 connector 单元测试和 fake server 测试。

验收：

- `GET /api/v1/findx/targets` 能通过 Nightingale API 返回目标列表。
- `GET /api/v1/findx/alerts/active` 能返回 Nightingale 活跃告警摘要。
- `GET /api/v1/findx/templates` 能返回 integrations/builtin payload 目录。
- 断开 Nightingale 时返回 503 和可读错误，不影响 AI WorkBench 基础启动。

### Phase 2：findx-agents 包装层

目标：用 findx-agents 品牌封装 Categraf。

工作项：

- 设计 findx-agents 安装目录、服务名、配置目录和日志目录。
- 生成 Categraf 配置模板：writers、heartbeat、global labels、inputs。
- 默认 heartbeat 指向 Nightingale `/v1/n9e/heartbeat`。
- 增加可选 FindX alive heartbeat：只上报 agent_id、hostname、ip、version、status、last_seen，不承载监控数据。
- 保留 Categraf 原始能力和配置结构，不 fork 采集插件。
- 安装脚本从 `catpaw` 改为 `findx-agents`。
- 前端探针管理页从 Catpaw 重命名为 findx-agents。

验收：

- Linux 目标机安装后出现 `findx-agents.service`。
- Nightingale targets 中能看到对应主机 heartbeat。
- AI WorkBench findx-agents 页面能看到在线/离线状态。
- 不再安装 `/usr/local/bin/catpaw` 或 `/etc/catpaw`。

### Phase 3：模板安装与 Dashboard 嵌入

目标：用户可在 FindX 模板中心安装 Nightingale/Categraf 成熟模板。

工作项：

- 支持 dashboard payload 安装和 `/board/:bid/pure` 嵌入。
- 支持 alert payload 安装。
- 支持 collect payload 生成 Categraf `input.*` 配置。
- 支持 record-rules 安装。
- 支持 metric desc/builtin metrics 同步为 AI 只读指标字典。
- 支持 markdown/runbook 预览和 AI 引用。
- 支持模板安装前差异预览、变量填写、权限确认和回滚。

验收：

- Linux、MySQL、Redis 至少三个 integration 可完整展示 dashboard/alert/metrics/collect/markdown。
- Linux dashboard 可在 FindX 内嵌展示。
- MySQL 模板安装后，findx-agents 生成 input 配置，Nightingale 可见对应 dashboard/alert。
- 安装失败时不留下半写状态，错误脱敏。

### Phase 4：监控能力下沉 Nightingale

目标：把监控管理能力从 AI WorkBench 自研逻辑迁移到 Nightingale。

工作项：

- 告警中心改读 Nightingale 活跃/历史事件。
- 告警规则创建、更新、删除走 Nightingale 接口或嵌入页。
- 仪表盘入口优先嵌入 Nightingale dashboard。
- 数据源管理优先使用 Nightingale datasource。
- 通知渠道和订阅管理优先跳转或代理 Nightingale 页面。
- AI WorkBench 本地只保留 AI 诊断记录、知识库、用户会话和 findx-agents 存活缓存。

验收：

- 用户在 FindX 告警中心看到的数据与 Nightingale 活跃告警一致。
- 用户在 FindX 仪表盘看到的是 Nightingale dashboard 或 Nightingale dashboard 的受控嵌入。
- AI WorkBench 不再维护第二套告警规则事实源。

### Phase 5：AI 问诊与 Nightingale 模板证据链融合

目标：AI 问诊从 Nightingale 获取更完整证据，并能推荐模板。

工作项：

- AIOps 诊断上下文新增 NightingaleEvidenceProvider。
- 诊断输入包括 target、busi group、active alerts、alert rule、PromQL、dashboard link、datasource status、metric desc、record rules、integration markdown、event pipeline status。
- AI 结论可回写到 AI WorkBench 诊断记录。
- 可选将诊断摘要回写到 Nightingale 事件备注、通知或 webhook。
- PromQL 查询优先走 Nightingale datasource/query 或受控 Prometheus 模板。
- suggestedActions 保持只读，写操作必须二次确认。

验收：

- 对某个 Nightingale 告警一键发起 AI 问诊。
- 诊断报告能引用 Nightingale 事件、规则、目标、PromQL、指标解释、dashboard、markdown/runbook。
- AI 能推荐缺失模板，但不会自动写入。
- 诊断结果不出现裸 JSON、内部连接串、真实 token 或上游 raw error。

### Phase 6：通知、事件流水线与 Hermes

目标：复用 Nightingale 通知与事件流水线，把 AI 问诊作为增强节点接入。

工作项：

- 读取 message templates、notify templates、notify channels、notify rules。
- 支持 event pipeline 列表、详情、tryrun、trigger、stream、executions。
- AI 问诊完成后可生成脱敏摘要，作为 event pipeline 输出或 Hermes 推送内容。
- Hermes 只作为消息触达和异步入口，不成为监控事实源。

验收：

- FindX 能展示 Nightingale 通知模板和事件流水线摘要。
- AI 诊断摘要可通过受控 webhook/event pipeline 流转。
- 所有写操作要求管理员权限、二次确认和审计。

### Phase 7：Catpaw 兼容迁移与删除计划

目标：平滑迁移旧 Catpaw 页面、API 和数据。

工作项：

- `/api/v1/catpaw/*` 标记为 deprecated。
- 新增 `/api/v1/findx/agents/*`。
- 旧 `catpaw_agents` 表迁移或映射到 `findx_agents`。
- 旧页面 `CatpawInstall.vue` 重命名或替换为 `FindxAgents.vue`。
- 知识库同义词从 Catpaw 扩展到 findx-agents。
- 文档中说明 Catpaw 仅为历史名称。

验收：

- 新用户路径不再出现 Catpaw。
- 老接口可在兼容期返回 deprecation header。
- 旧数据可被 findx-agents 页面读取。

### Phase 8：联调、压测和交付

目标：形成可交付的一体化方案。

工作项：

- WSL/Linux 单包构建。
- Nightingale + Categraf/findx-agents + AI WorkBench docker compose 或 systemd 联调。
- findx-agents 安装、卸载、重装、升级、离线、恢复测试。
- Nightingale API 断连、认证失败、超时、返回结构变化测试。
- 告警事件到 AI 问诊闭环测试。
- 许可证和 NOTICE 文件检查。

验收：

- `go build` 和 `npm run build` 通过。
- Nightingale 目标、告警、仪表盘、模板在 FindX 入口可见。
- findx-agents 在线状态准确。
- AI 问诊能基于 Nightingale 告警完成一次闭环诊断。

## 12. P0 / P1 / P2 排期

### P0：必须先做

1. 能力边界冻结：Nightingale 为监控事实源和模板事实源，AI WorkBench 不重写 Nightingale 成熟能力。
2. 许可证策略冻结：findx-agents 不直接从 Catpaw 派生。
3. Nightingale connector 基础配置与认证。
4. 只读拉取 Nightingale targets、active alerts、datasources、builtin-components、builtin-payloads、builtin-metrics。
5. FindX 模板中心只读目录：integrations、dashboards、alerts、metrics、record-rules、markdown、collect toml。
6. findx-agents 命名、服务名、路径和安装模板设计。
7. Categraf heartbeat 到 Nightingale 打通。
8. AI WorkBench findx-agents 存活缓存和页面入口。
9. Dashboard 嵌入 PoC：`/board/:bid/pure`。
10. 敏感配置治理：Nightingale、Categraf、AI key、数据库连接串不得明文进入文档、日志或前端。

### P1：第一轮可用闭环

1. FindX 告警中心读取 Nightingale 活跃/历史告警。
2. FindX 目标资产读取 Nightingale targets/busi groups。
3. FindX 仪表盘嵌入 Nightingale dashboard。
4. FindX 模板中心支持安装 dashboard、alert、collect、record-rule。
5. AI 问诊读取 Nightingale 告警、目标、规则、数据源、PromQL、metric desc、dashboard、markdown/runbook 证据。
6. 旧 Catpaw 页面改名为 findx-agents 并保留兼容层。
7. findx-agents 安装/卸载/升级脚本。
8. Nightingale API fake server 测试和 WSL 联调。

### P2：增强与商业化体验

1. 告警规则的品牌化编辑入口。
2. Nightingale 通知渠道、消息模板和订阅策略的品牌化入口。
3. Nightingale event pipelines 的品牌化入口、试跑和执行记录展示。
4. AI 推荐告警规则，人工确认后写入 Nightingale。
5. AI 推荐 dashboard，人工确认后写入 Nightingale。
6. Hermes 接收 Nightingale 告警摘要和 AI 诊断结果。
7. Categraf http_provider 远程配置下发。
8. findx-agents 多平台安装包、升级通道、版本矩阵。
9. 执行动作二次确认与审计。

### Deferred

- Catpaw 插件巡检迁移。
- Catpaw AI chat 能力迁移。
- 远程 shell/自动修复。
- Ibex 远程任务深度接入。
- 直接改造 Nightingale 前端源码。

## 13. 主要风险

| 风险 | 等级 | 说明 | 缓解 |
| --- | --- | --- | --- |
| Catpaw AGPL 派生风险 | P0 | 直接重命名派生会带来更强开源义务。 | findx-agents 基于 Categraf 包装，不直接复制 Catpaw。 |
| 双事实源 | P0 | AI WorkBench 与 Nightingale 同时维护告警、资产、模板会冲突。 | Nightingale 为事实源，AI WorkBench 做视图、索引缓存和 AI。 |
| 模板安装漂移 | P1 | 上游模板升级后，本地安装对象可能漂移。 | hash、版本、source、installed_ref、diff 和 rollback。 |
| Nightingale API 变动 | P1 | 直接依赖内部表或不稳定 API 会脆弱。 | connector 层隔离，fake server 测试。 |
| Categraf 包装过深 | P1 | fork 太深会难以跟随上游。 | 只包装安装、配置和命名，不改采集插件核心。 |
| 敏感配置泄露 | P0 | token、DSN、密码可能进入配置和日志。 | 环境变量/secret、脱敏、审计、文档占位符。 |
| UI 命名和开源归属冲突 | P1 | 用户侧要 FindX 命名，合规侧要保留归属。 | 产品命名和合规说明分层。 |
| AI 问诊证据污染 | P1 | 旧 Catpaw 数据和 Nightingale 数据同时存在。 | 诊断 evidence 标准化，标记 source、freshness、upstream_ref。 |
| Dashboard 嵌入权限 | P1 | iframe/session/proxy 处理不当会越权或白屏。 | 优先 pure board + 后端代理/同源鉴权，覆盖权限测试。 |
| Event pipeline 写操作风险 | P1 | 流水线可触发外部动作。 | P1 只读/试跑，P2 写操作二次确认、审计、权限控制。 |

## 14. 需要修改的主要模块

AI WorkBench 后端：

- `api/main.go`
- `api/internal/handler/n9e_agents.go`
- `api/internal/handler/n9e_alerts.go`
- `api/internal/handler/catpaw.go`
- `api/internal/handler/remote.go`
- `api/internal/store/agents.go`
- `api/internal/model/diagnose.go`
- `api/internal/handler/aiops_*`
- 新增 `api/internal/nightingale/`
- 新增 `api/internal/findxagent/`
- 新增 `api/internal/templatecatalog/`

AI WorkBench 前端：

- `web/src/views/CatpawInstall.vue` -> `FindxAgents.vue`
- `web/src/views/CatpawChatPanel.vue` 后续移除或兼容隐藏
- `web/src/views/Diagnose.vue`
- `web/src/views/Workbench.vue`
- `web/src/views/DataSource.vue`
- `web/src/views/Alerts.vue`
- `web/src/router/index.js`
- 新增 FindX 目标、告警、仪表盘、模板中心入口

文档与脚本：

- `README.md`
- `docs/aiops/`
- `docs/运维手册.md`
- `scripts/build-one-package.sh`
- `scripts/install-one-service.sh`
- 新增 findx-agents 安装模板、Nightingale/Categraf 第三方许可证说明、模板中心使用手册。

## 15. 验证计划

代码验证：

- 后端：`cd /opt/ai-workbench/api && go build -o /tmp/aiw-nightingale-plan-api .`
- 前端：`cd /opt/ai-workbench/web && npm run build`
- Connector：fake Nightingale server 单测覆盖 targets、alerts、datasources、builtin-components、builtin-payloads、message-templates、event-pipelines、timeout、401/403、schema drift。
- Template catalog：JSON/TOML 校验、hash 幂等、drift 标记、安装失败回滚。

联调验证：

- Nightingale 启动后，AI WorkBench 能读取 targets。
- Categraf/findx-agents heartbeat 后，Nightingale target 在线。
- AI WorkBench findx-agents 页面显示在线。
- Nightingale 产生告警后，FindX 告警中心显示一致。
- 从告警触发 AI 问诊，诊断报告引用 Nightingale 证据和模板说明。
- Linux dashboard 可通过 `/board/:bid/pure` 在 FindX 内嵌展示。
- MySQL/Redis/Linux 模板可完整预览并安装。

安全验证：

- 配置、日志、错误、前端响应不泄露真实 `<API_KEY>`、`<TOKEN>`、`<DB_DSN>`、`<PASSWORD>`、`<SSH_KEY>`。
- Nightingale connector 认证失败返回可读错误，不返回内部栈。
- 任意写操作默认拒绝或要求二次确认。
- Event pipeline trigger/stream 必须有权限、审计和速率限制。

## 16. 上一版扣分点补齐情况

| 上一版扣分点 | 本轮补齐方式 | 剩余风险 |
| --- | --- | --- |
| Nightingale 集成不够深 | 已纳入 integrations、builtin components、builtin payloads、metrics desc、record rules、message templates、notify templates、notify channels、event pipelines、embedded dashboards/products。 | 仍需运行态 PoC 验证 API 字段。 |
| 模板体系未完整覆盖 | 新增 FindX 模板中心方案，覆盖 dashboard/alert/collect/metric/record-rule/markdown/notify/event pipeline 全类型。 | 模板安装的字段映射需要开发阶段逐项验证。 |
| Categraf 模板不够细 | 已审 92 个 input 目录、105 个 toml，明确 findx-agents 只生成/管理 Categraf 配置，不 fork 插件。 | http_provider 远程下发需要安全 PoC。 |
| 未做 dashboard 嵌入 PoC | 已指定 `/board/:bid/pure` 为 P0 PoC 路径，并补权限/断连/白屏验收。 | 尚未真实启动 Nightingale 浏览器验证。 |
| 未启动 Nightingale/Categraf 实例联调 | 补充 fake server、docker compose/systemd 联调和 P0/P1 验收路径。 | 本轮仍为静态审计和文档立项，没有宣称已完成真实联调。 |
| Dify/FastGPT/RAGFlow 未逐项审 | 明确它们不是本次监控集成主线，仅作为 AI/RAG 参考项目；后续引入需单独立项。 | 若要复用其代码/UI，仍需许可证专项。 |
| 许可证判断不能替代法律意见 | 保留为残余风险，新增 NOTICE/THIRD_PARTY 和正式法务确认要求。 | 需正式法务确认。 |

## 17. 本计划自评分

综合评分：93 / 100。

| 维度 | 权重 | 得分 | 说明 |
| --- | ---: | ---: | --- |
| 源码证据充分性 | 30 | 28 | 已补充 Nightingale integrations/API/router/builtin payload/component/event pipeline/embedded product 与 Categraf providers/input 模板证据，也审计了 AI WorkBench 当前 N9E/Catpaw 接入边界。未做全量逐行审计。 |
| 架构一致性 | 25 | 24 | 能力归属更清晰，符合“除 AI 问诊和探针存活外，成熟监控能力用夜莺/Categraf”的目标。 |
| 许可证与风险识别 | 20 | 18 | 识别 Nightingale Apache-2.0、Categraf MIT、Catpaw AGPL-3.0 关键差异，并补 NOTICE/THIRD_PARTY 要求。仍需正式法务确认。 |
| 可执行性 | 15 | 14 | 已拆 Phase/P0/P1/P2、模板中心元数据、同步流程、安装策略、PoC 验收和模块清单。 |
| 验证完整性 | 10 | 9 | 已补 fake server、Dashboard 嵌入 PoC、模板校验、联调、安全验证；本轮仍未真实启动 Nightingale/Categraf 实例。 |

扣分项保留：

- 尚未启动 Nightingale/Categraf 实例做真实 API 联调。
- 尚未完成 FindX 前端嵌入 Nightingale dashboard 的代码级 PoC。
- 许可证判断基于本地源码 LICENSE，不替代正式法律意见。
- 模板安装字段映射需在开发阶段逐项用真实对象验证。

建议结论：本计划已达到“可立项，建议进入 P0 设计与 PoC”的程度。是否进入开发由主 AI/项目负责人决策。

## 18. 参考源

- 本地源码：`D:\平台源码\nightingale-main (1)\nightingale-main`
- 本地源码：`D:\平台源码\categraf-main (1)\categraf-main`
- 本地源码：`D:\平台源码\catpaw-master\catpaw-master`
- 本地源码：`D:\平台源码\AutoOps-main\AutoOps-main`
- 当前 git 工作区：`D:\ai-workbench`
- 当前 WSL 运行项目：`/opt/ai-workbench`
- 官方仓库：`https://github.com/ccfos/nightingale`
- 官方仓库：`https://github.com/flashcatcloud/categraf`
