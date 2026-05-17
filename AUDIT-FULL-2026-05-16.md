# FindX 平台完整审计报告（详细版）

> 审计日期：2026-05-16
> 对比源码：Nightingale(n9e) + SkyWalking + SignOz + Categraf + Catpaw
> 对比目标：https://8.lwops.cn (LWOPS 运维智能体)
> 逆向数据：D:\测试\ 目录
> 审计范围：后端引擎 + 前端功能点 + UI 层级 + Agent 下发 + CMDB 对齐 + 探针对齐

---

## 一、Codex Agent 下发设计审计

### 1.1 当前实现分析

Codex 在 `web/src/react-shell/agents/` 下实现了一套 Agent 管理中心，包含：
- AgentPage.jsx — 主页面（6 个 Tab：概览/主机Agent/能力包/插件目录/配置下发/环境适配）
- HostsSection.jsx — 主机列表 + 安装抽屉 + 配置下发抽屉
- agentCapabilityModel.js — 7 个能力包定义
- agentTemplateModel.js — 8 种安装命令预览

### 1.2 设计问题

| 问题 | 严重度 | 说明 |
|------|--------|------|
| 🔴 全部是占位符 | 致命 | 所有关键字段都是 `<TOKEN>`、`<CREDENTIAL_REF>`、`<PACKAGE_ID>` 等占位符，无真实执行能力 |
| 🔴 BLOCKED_BY_CONTRACT 到处都是 | 致命 | 页面上大量显示"契约阻断"提示，用户看到的全是不可用状态 |
| 🔴 无真实安装执行器 | 致命 | 只有命令预览（curl/certutil/PowerShell），无 SSH 推送和执行逻辑 |
| 🔴 无配置模板渲染 | 致命 | 配置下发只是 JSON 预览，无 .toml 模板参数填写和渲染 |
| 🟡 页面过长 | 体验差 | HostsSection 单文件 280 行，InstallDrawer 和 ConfigDrawer 内容堆叠，无步骤分离 |
| 🟡 字段过于抽象 | 不实用 | "能力域"、"契约"、"证据链"等概念对运维人员不友好 |
| 🟡 缺少进度反馈 | 体验差 | 安装/下发后只显示 BLOCKED 提示，无步骤条/进度条 |

### 1.3 字段合理性审计

**安装抽屉（InstallDrawer）字段**：
- ✅ 下发目标（TargetPicker）— 合理
- ✅ 能力包选择 — 合理但名称太抽象（"FindX Agent 核心"不如"Categraf 采集器"直观）
- ✅ 操作系统选择 — 合理
- ✅ 安装方式选择 — 合理（Linux curl / Windows CMD / PowerShell / K8s）
- ❌ 缺少：凭证选择（SSH 密码/密钥）
- ❌ 缺少：安装路径配置
- ❌ 缺少：服务名称配置
- ❌ 缺少：端口配置
- ❌ 缺少：安装前检查（端口冲突、磁盘空间、依赖检查）

**配置下发抽屉（ConfigDrawer）字段**：
- ✅ 模板选择 — 合理
- ✅ 下发模式（全部Agent/业务组/主机/能力包）— 合理
- ✅ 下发策略（保存模板/灰度/全量/回滚）— 合理
- ❌ 缺少：具体配置参数填写（如 MySQL 的 address/username/password）
- ❌ 缺少：配置预览（渲染后的 .toml 内容）
- ❌ 缺少：配置对比（当前 vs 新配置 diff）
- ❌ 缺少：影响范围确认
- ❌ 缺少：回滚版本选择

### 1.4 对比运维规范的建议

**正确的 Agent 下发流程应该是**：
```
1. 选择目标主机（从 CMDB 勾选，显示 IP/主机名/OS/当前 Agent 状态）
2. 选择凭证（SSH 密钥/密码，从凭证库选择）
3. 安装前检查（连通性测试 + 端口检查 + 磁盘空间）
4. 选择安装包（Categraf/Catpaw/SkyWalking Agent + 版本）
5. 配置参数（根据包类型动态表单：采集间隔、上报地址、标签等）
6. 确认执行（显示将要执行的命令摘要）
7. 执行安装（实时日志输出 + 步骤进度条）
8. 验证结果（心跳检查 + 数据到达验证）
```

**Codex 的实现跳过了 2/3/5/6/7/8 步**，只做了 1 和 4 的壳。

---

## 二、源码功能点一比一对比

### 2.1 Nightingale (n9e) 后端 — 65 个路由文件

| n9e 路由文件 | 功能 | FindX 对应 | 实现深度 |
|-------------|------|-----------|----------|
| router_alert_rule.go | 告警规则 CRUD + 克隆 + 启停 + 试运行 | ✅ 有路由 | 🟡 有 CRUD 无试运行 |
| router_mute.go | 告警屏蔽（时间窗口+标签匹配） | ❌ 无实现 | 🔴 |
| router_alert_subscribe.go | 告警订阅（按标签/业务组） | ❌ 无实现 | 🔴 |
| router_event_pipeline.go | 事件流水线（8种处理器） | ❌ 无实现 | 🔴 |
| router_notify_channel.go | 通知媒介（20+类型） | 🟡 有 CRUD | 🔴 无类型逻辑 |
| router_notify_rule.go | 通知规则（条件+升级+静默） | 🟡 有 CRUD | 🔴 无条件引擎 |
| router_notify_tpl.go | 通知模板（Go template渲染） | 🟡 有 CRUD | 🔴 无渲染 |
| router_message_template.go | 消息模板管理 | 🟡 有 | 🟡 |
| router_recording_rule.go | 记录规则 | ❌ 无实现 | 🔴 |
| router_dashboard.go | 仪表盘 CRUD + 克隆 + 分享 | ✅ 有 | 🟡 |
| router_datasource.go | 数据源管理（9种类型） | 🟡 只有 Prometheus | 🔴 |
| router_query.go | 数据查询代理 | 🟡 只代理 Prom | 🔴 |
| router_target.go | 目标管理（心跳+标签+业务组） | ✅ 有 | 🟡 |
| router_heartbeat.go | Agent 心跳接收 | ✅ 有 | 🟢 |
| router_builtin.go | 内置组件管理 | ✅ 有 | 🟡 |
| router_builtin_component.go | 内置组件 CRUD | ✅ 有 | 🟡 |
| router_builtin_payload.go | 内置 Payload（仪表盘/告警模板） | ✅ 有 | 🟡 |
| router_builtin_metrics.go | 内置指标说明 | ❌ 无 | 🔴 |
| router_builtin_metric_filter.go | 指标过滤器 | ❌ 无 | 🔴 |
| router_user.go | 用户管理 | ✅ 有 | 🟢 |
| router_user_group.go | 用户组/团队 | ✅ 有 | 🟡 |
| router_role.go | 角色管理 | ✅ 有 | 🟡 |
| router_role_operation.go | 角色权限操作 | ✅ 有 | 🟡 |
| router_busi_group.go | 业务组管理 | ✅ 有 | 🟡 |
| router_login.go | 登录/SSO | ✅ 有登录 | 🔴 无 SSO |
| router_self.go | 个人设置 | ✅ 有 | 🟢 |
| router_config.go | 系统配置 | ✅ 有 | 🟡 |
| router_configs.go | 全局配置 | ✅ 有 | 🟡 |
| router_mcp_server.go | MCP 服务管理 | ✅ 有 | 🟡 |
| router_ai_llm_config.go | AI LLM 配置 | ✅ 有 | 🟢 |
| router_ai_assistant.go | AI 助手 | ✅ 有（AIOps） | 🟡 |
| router_ai_skill.go | AI 技能 | ❌ 无 | 🟡 |
| router_ai_agent.go | AI Agent 配置 | ✅ 有 | 🟡 |
| router_saved_view.go | 保存视图 | ✅ 有 | 🟢 |
| router_metric_view.go | 指标视图 | ❌ 无 | 🟡 |
| router_metric_desc.go | 指标描述 | ❌ 无 | 🔴 |
| router_chart_share.go | 图表分享 | ❌ 无 | 🟡 |
| router_dash_annotation.go | 仪表盘注解 | ❌ 无 | 🟡 |
| router_embedded.go | 嵌入式仪表盘 | ❌ 无 | 🟡 |
| router_es.go | ES 数据源代理 | ❌ 无 | 🔴 |
| router_es_index_pattern.go | ES 索引模式 | ❌ 无 | 🔴 |
| router_opensearch.go | OpenSearch 代理 | ❌ 无 | 🔴 |
| router_tdengine.go | TDengine 代理 | ❌ 无 | 🔴 |
| router_proxy.go | 通用代理 | ❌ 无 | 🔴 |
| router_task.go | 任务执行 | ❌ 无 | 🔴 |
| router_task_tpl.go | 任务模板 | ❌ 无 | 🔴 |
| router_source_token.go | 数据源 Token | ❌ 无 | 🟡 |
| router_trace_logs.go | Trace-Log 关联 | 🟡 有基础 | 🟡 |
| router_notification_record.go | 通知记录查询 | ❌ 无 | 🔴 |
| router_user_variable_config.go | 用户变量配置 | ❌ 无 | 🟡 |
| router_captcha.go | 验证码 | ❌ 无 | 🟡 |
| router_crypto.go | 加密工具 | ❌ 无 | 🟡 |
| router_event_detail.go | 事件详情 | 🟡 有基础 | 🟡 |

**统计**：n9e 65 个路由文件，FindX 有对应实现的约 35 个（54%），其中真正有业务深度的约 15 个（23%）。

### 2.2 Nightingale 告警引擎（alert/）— 完全缺失

| n9e 模块 | 功能 | FindX | 状态 |
|----------|------|-------|------|
| alert/eval/eval.go | PromQL 定时评估循环（每 15s 执行一次规则） | scheduler/ 有框架但无真实 PromQL 评估 | 🔴 |
| alert/eval/alert_rule.go | 规则状态机（pending→firing→resolved） | 无 | 🔴 |
| alert/mute/ | 屏蔽匹配引擎（标签+时间窗口） | 无 | 🔴 |
| alert/dispatch/ | 事件分发（按规则路由到通知渠道） | 无 | 🔴 |
| alert/sender/ | 通知发送（20+ provider） | notifier/ 空壳 | 🔴 |
| alert/sender/provider/ | 29 个通知 provider 实现 | 无 | 🔴 |
| alert/pipeline/ | 事件流水线引擎 | 无 | 🔴 |
| alert/pipeline/processor/ | 8 种处理器（relabel/drop/callback/AI摘要等） | 无 | 🔴 |
| alert/record/ | 记录规则执行 | 无 | 🔴 |
| alert/naming/ | 告警引擎命名和发现 | 无 | 🔴 |
| alert/queue/ | 事件队列 | 无 | 🔴 |

### 2.3 Nightingale 数据源（datasource/）— 严重不足

| n9e 数据源 | FindX 支持 | 状态 |
|-----------|-----------|------|
| Prometheus / VictoriaMetrics | ✅ 有代理 | 🟢 |
| Elasticsearch | ❌ | 🔴 |
| ClickHouse | ❌ | 🔴 |
| TDengine | ❌ | 🔴 |
| OpenSearch | ❌ | 🔴 |
| Doris | ❌ | 🔴 |
| MySQL | ❌ | 🔴 |
| PostgreSQL | ❌ | 🔴 |
| VictoriaLogs | ❌ | 🔴 |

### 2.4 Nightingale 集成模板（integrations/）— 数据未导入

n9e 有 65 个集成组件，每个包含：
- `collect/` — Categraf 采集配置 .toml 模板
- `dashboards/` — 预置仪表盘 JSON
- `alerts/` — 预置告警规则 JSON
- `metrics/` — 指标说明（名称/单位/描述/多语言）
- `markdown/` — 接入文档
- `icon/` — 图标

**FindX 现状**：有 TemplatesPage 框架和 API（builtin-components/builtin-payloads），但 65 个组件的实际数据（.toml + 仪表盘 + 告警 + 指标说明）完全未导入。

**65 个集成列表**：
AMD_ROCm_SMI, AliYun, AppDynamics, AutoMQ, Bind, Canal, Ceph, ClickHouse, CloudWatch, Consul, Dns_Query, Docker, Doris, Elasticsearch, Exec, Filecount, Gitlab, GoogleCloud, HAProxy, HTTP_Response, IPMI, IPVS, Java, Jenkins, Jolokia_Agent, Kafka, Kubernetes, Ldap, Linux, Logstash, MinIO, MongoDB, Mtail, MySQL, N9E, NFSClient, NSQ, NVIDIA, Net_Response, Netstat_Filter, Nginx, Oracle, PHP, Ping, PostgreSQL, Procstat, Prometheus, RabbitMQ, Redis, SMART, SNMP, SQLServer, SpringBoot, Switch_Legacy, Systemd, TDEngine, TiDB, Tomcat, VictoriaMetrics, Whois, Windows, XSKYApi, ZooKeeper, cAdvisor, vSphere

### 2.5 Categraf 探针对比（112 个采集插件）

Categraf 源码有 112 个 input 插件。FindX 的定位是**管理和下发配置**，不含采集器本体，这是正确的设计。但配置下发需要：

| 需要实现 | FindX 现状 | 状态 |
|----------|-----------|------|
| 每个插件的 .toml 配置模板 | 无（只有抽象的 configTemplates） | 🔴 |
| 配置参数动态表单（根据插件类型生成） | 无 | 🔴 |
| 配置渲染引擎（占位符→真实值） | 无 | 🔴 |
| SSH 推送到 /opt/categraf/conf/input.xxx/ | 有基础 remote/exec | 🟡 |
| reload 命令执行 | 无 | 🔴 |
| 配置漂移检测 | 无 | 🔴 |
| 配置回滚 | 无 | 🔴 |
| 下发回执验证 | 无 | 🔴 |

### 2.6 Catpaw 巡检探针对比（35+ 插件）

Catpaw 源码有 35+ 巡检插件（cert/cpu/disk/diskio/dns/docker/etcd/exec/filecheck/filefd/http/mem/mount/net/ntp/ping/procfd/procnum/redis/sockstat/sysctl/systemd/tcpstate/uptime/zombie 等）。

| 需要实现 | FindX 现状 | 状态 |
|----------|-----------|------|
| 巡检插件配置模板 | tmp/catpaw-conf/ 有 .toml 模板 | 🟢 有模板文件 |
| 远程安装 catpaw 二进制 | 无 | 🔴 |
| 远程推送插件配置 | 无 | 🔴 |
| 巡检结果接收和展示 | 有 CatpawReport handler | 🟡 |
| 巡检调度（定时触发） | scheduler 有框架 | 🟡 |

### 2.7 SkyWalking Agent 探针对比

| 需要实现 | FindX 现状 | 状态 |
|----------|-----------|------|
| Java Agent 配置生成（-javaagent 参数） | 无 | 🔴 |
| Go Agent 配置生成 | 无 | 🔴 |
| Python Agent 配置生成 | 无 | 🔴 |
| Node.js Agent 配置生成 | 无 | 🔴 |
| .NET Agent 配置生成 | 无 | 🔴 |
| Agent 包远程推送 | 无 | 🔴 |
| 应用启动脚本注入 | 无 | 🔴 |
| Trace 数据到达验证 | 有 APMTraceReceive handler | 🟡 |
| 接入向导页面 | 无 | 🔴 |

### 2.8 SkyWalking Booster UI 功能对比

| SkyWalking 功能 | FindX 对应 | 实现深度 |
|----------------|-----------|----------|
| 服务列表（SLA/CPM/延迟/错误率） | TracingListServicesSW | 🟡 代理到 OAP |
| 实例列表（JVM/CLR 指标） | TracingListInstancesSW | 🟡 |
| 端点列表（TOP N 慢端点） | TracingListEndpointsSW | 🟡 |
| 服务拓扑（力导向图+指标弹窗） | TracingGetTopologySW | 🟡 缺指标弹窗 |
| Trace 检索（多条件过滤） | TracingQueryTracesSW | 🟡 |
| Trace 瀑布图（Span树+时间轴+Tag+Log） | TraceWaterfall.jsx | 🟡 缺 Span Log |
| Profiling 任务（创建/取消/结果） | APMListProfilingTasksSW | 🟡 代理 |
| 告警列表 + 确认 | APMListAlarmsSW | 🟡 |
| 设置（采样率/慢端点阈值） | APMGetSettingsSW | 🟡 |
| Dashboard 编辑器（Widget/Metric/Topology） | 无 | 🔴 |
| 日志关联（Trace→Log） | TraceLinkSection | 🟡 |
| 浏览器端 RUM | 无 | 🔴 |
| 事件时间线 | 无 | 🔴 |
| Marketplace（插件市场） | 无 | 🔴 |

### 2.9 SignOz 功能对比

| SignOz 功能 | FindX 对应 | 实现深度 |
|------------|-----------|----------|
| Logs Explorer（查询+字段+聚合） | LogsPage | 🟡 |
| Traces Explorer（列表+瀑布图+过滤） | TracingPage | 🟡 |
| Infrastructure Monitoring（主机/容器/K8s） | 无独立页面 | 🔴 |
| Alerts（规则+历史+通道） | AlertsPage | 🟡 |
| Dashboards（面板+变量+模板） | DashboardsPage | 🟡 |
| Services（APM 服务列表） | ServicesSection | 🟡 |
| Messaging Queues 监控 | 无 | 🔴 |
| Exceptions/Errors 追踪 | 无 | 🔴 |
| Pipelines（日志处理管道） | PipelinesSection | 🟡 框架 |
| Integrations Marketplace | IntegrationsPage | 🟡 框架 |
| Onboarding（接入向导） | 无 | 🔴 |
| Billing/License | 无 | N/A |

---

## 三、LWOPS CMDB 功能点一比一对齐

### 3.1 LWOPS CMDB 完整菜单（已通过 Playwright 验证）

```
CMDB
├── 概况（资源分类统计卡片 + 告警状态分布）
├── 全文检索（跨模型全文搜索）
├── 资源管理
│   ├── 资源列表（表格+筛选+批量操作+下发Agent）
│   ├── 业务视图（树形业务组+资源关联）
│   ├── 机房视图（机柜/U位可视化）
│   └── 回收站（软删除恢复）
├── 模型配置
│   ├── 模型管理（模型树+属性定义+图标）
│   ├── 关联类型（模型间关系定义）
│   └── 属性单位（自定义单位）
├── 发现管理
│   ├── 自动发现（凭证+发现规则+范围）
│   ├── 执行记录（发现任务日志）
│   └── 自动化映射（发现结果→模型实例映射）
├── 资源审批
│   ├── 我的申请
│   ├── 我的待办
│   └── 已归档
├── 资源报表
│   ├── 资源统计（按类型/业务/机房统计）
│   ├── 变更统计
│   ├── 模型变更
│   ├── 实例变更TOP
│   ├── 自定义报表
│   └── 云平台账单
├── 资产消费
│   ├── 合规性检查
│   ├── 合规性统计
│   ├── 变更通知
│   ├── 备件管理
│   ├── 巡检配置
│   ├── 关系查询
│   └── 事件订阅
└── 审计记录
    ├── 通知记录
    ├── 变更记录
    └── 订阅记录
```

### 3.2 LWOPS 监控中心资源列表字段（已验证）

**表头**：对象名称 | 业务名称 | IP | 采集情况 | 类型 | 子类型 | 分组 | 维护情况 | 负责人

**筛选器**：
- 告警级别快筛（紧急/严重/次要/警告/信息/正常/未监控）
- 关键字（名称/IP）
- IP（支持精准匹配）
- 类型（操作系统/数据库/中间件/网络设备/服务器/存储/链路/虚拟化/探测/云平台/容器/物联网）
- 分组
- 维护情况
- 采集情况（正常/异常/—）
- 模板
- 标签
- 所属业务

**操作列**：编辑 | 删除 | 更多（含下发Agent、监控绑定等）

**视图切换**：列表视图 / 其他视图

### 3.3 FindX CMDB 对比 LWOPS — 逐项对齐

| LWOPS 功能 | FindX 实现 | 差距 |
|-----------|-----------|------|
| 概况页（12 类资源统计卡片+告警分布） | OverviewSection 有但简单 | 🟡 缺告警分布 |
| 全文检索 | 有入口 | 🟡 |
| 资源列表表格 | HostsSection 有 | 🔴 字段不全（见下） |
| 资源列表筛选器（9 个筛选条件） | 只有搜索框 | 🔴 |
| 资源列表批量操作 | 无 | 🔴 |
| 资源列表"更多"操作（下发Agent） | 无 | 🔴 |
| 业务视图（树形） | BusinessSection 有 | 🟡 |
| 机房视图（机柜U位可视化） | 无 | 🔴 |
| 回收站 | 无 | 🔴 |
| 模型管理（模型树+属性） | ModelTreeSection + ModelDetailSection | 🟢 |
| 关联类型 | 有 | 🟡 |
| 属性单位 | 无 | 🔴 |
| 自动发现（凭证+规则+执行） | 无 | 🔴 |
| 执行记录 | 无 | 🔴 |
| 自动化映射 | 无 | 🔴 |
| 资源审批流程 | 有 API 但前端简单 | 🟡 |
| 资源报表（6种） | 无 | 🔴 |
| 资产消费（7种） | 无 | 🔴 |
| 审计记录（3种） | 有基础审计 | 🟡 |

### 3.4 CMDB 资源列表字段对比

| LWOPS 字段 | FindX HostsSection 字段 | 差距 |
|-----------|------------------------|------|
| 对象名称 | ✅ 主机名 | 🟢 |
| 业务名称 | ❌ 无 | 🔴 |
| IP | ✅ 有 | 🟢 |
| 采集情况（正常/异常/—） | ❌ 无（只有 Agent 安装状态） | 🔴 |
| 类型（操作系统/数据库/中间件等12类） | ❌ 无分类 | 🔴 |
| 子类型（Linux/Windows/MySQL/Nginx等） | ❌ 无 | 🔴 |
| 分组 | ❌ 无 | 🔴 |
| 维护情况（维护中/未维护） | ❌ 无 | 🔴 |
| 负责人 | ❌ 无 | 🔴 |

**FindX HostsSection 有但 LWOPS 没有的字段**：
- CMDB 登记状态
- FindX Agent 安装状态
- 心跳状态
- 数据到达证据
- 配置版本

**结论**：FindX 的 Agent 管理页面字段偏向"Agent 生命周期管理"视角，而 LWOPS 偏向"运维资产管理"视角。两者需要融合——资源列表应该同时展示资产属性和 Agent/采集状态。

---

## 四、前端 UI 层级（2级/3级抽屉）详细审计

### 4.1 需要实现的抽屉体系

| 场景 | 1级 | 2级 | 3级 | FindX 现状 |
|------|-----|-----|-----|-----------|
| 告警规则编辑 | 列表 | 全屏编辑抽屉（Tab：基础/触发/通知/生效时间） | PromQL编辑器/通知选择 | ❌ 只有 Modal |
| 告警事件详情 | 列表 | 事件详情抽屉 | 关联告警/通知轨迹/操作历史 | 🟡 基础 Drawer |
| 仪表盘 Panel 编辑 | 面板列表 | 左图右配置分栏 | 数据源/阈值/Overrides/变量 | ❌ 简单编辑器 |
| 集成组件详情 | 组件列表 | 5-Tab 抽屉（说明/仪表盘/告警/采集/指标） | 导入确认/配置预览 | ❌ 基础 Drawer |
| 通知规则编辑 | 列表 | 编辑抽屉 | 条件编辑器/模板预览 | ❌ 只有 CRUD |
| CMDB 实例详情 | 列表 | 多Tab抽屉（属性/监控/关联/告警/变更/Agent） | 属性编辑/监控绑定 | ❌ 无 |
| 事件流水线 | 列表 | 全屏编辑器 | 处理器配置弹窗 | ❌ 无 |
| Agent 安装向导 | 主机列表 | 步骤条抽屉（检查→选包→配置→执行→验证） | 参数表单/日志输出 | 🟡 简单 Drawer |
| Agent 配置下发 | 主机列表 | 步骤条抽屉（选模板→填参数→预览→确认→执行） | .toml 编辑器/diff 对比 | 🟡 简单 Drawer |

### 4.2 通用 Drawer 组件需求

FindX 需要一个通用 Drawer 基础组件，支持：
- 多宽度（sm: 400px / md: 600px / lg: 800px / xl: 1000px / full: 100%）
- 多级嵌套（2级推出时1级半透明）
- 底部固定操作栏（保存/取消/删除）
- URL 联动（打开抽屉同步 query 参数，刷新不丢失）
- 关闭前未保存提示
- Tab 切换
- 步骤条模式

---

## 五、n9e 前端页面一比一对比

### 5.1 n9e fe-main 页面列表 vs FindX

| n9e 页面目录 | 功能 | FindX 对应 | 实现 |
|-------------|------|-----------|------|
| pages/alertRules/ | 告警规则（列表+表单+多Tab） | AlertRulesSection | 🟡 缺表单深度 |
| pages/alertCurEvent/ | 当前告警事件 | AlertEventsSection | 🟡 |
| pages/historyEvents/ | 历史事件 | AlertEventsSection | 🟡 |
| pages/warning/shield/ | 告警屏蔽 | ❌ 无页面 | 🔴 |
| pages/warning/subscribe/ | 告警订阅 | ❌ 无页面 | 🔴 |
| pages/eventPipeline/ | 事件流水线 | ❌ 无页面 | 🔴 |
| pages/notificationRules/ | 通知规则 | NotificationRulesSection | 🟡 缺条件编辑 |
| pages/notificationChannels/ | 通知媒介 | NotificationChannelsSection | 🟡 缺类型表单 |
| pages/notificationTemplates/ | 消息模板 | NotificationTemplatesSection | 🟡 缺预览 |
| pages/dashboard/ | 仪表盘（13种面板+编辑器+变量） | DashboardsPage | 🟡 面板类型少 |
| pages/explorer/ | 数据查询（Prom/ES/Loki） | MetricExplorerPage | 🟡 只有 Prom |
| pages/datasource/ | 数据源管理 | DatasourcePage | 🟡 |
| pages/targets/ | 目标管理 | 在 Agent 页面 | 🟡 |
| pages/hosts/ | 主机管理 | HostsSection | 🟡 |
| pages/builtInComponents/ | 集成模板中心 | TemplatesPage | 🟡 缺数据 |
| pages/metricsBuiltin/ | 内置指标 | 无 | 🔴 |
| pages/recordingRules/ | 记录规则 | 无 | 🔴 |
| pages/user/ | 用户管理 | UsersSection | 🟢 |
| pages/contacts/ | 联系人 | 无独立页面 | 🟡 |
| pages/permissions/ | 权限管理 | RolesSection | 🟡 |
| pages/siteSettings/ | 站点设置 | SiteSettingsSection | 🟡 |
| pages/variableConfigs/ | 变量配置 | VariablesSection | 🟡 |
| pages/aiConfig/ | AI 配置 | PlatformPage models section | 🟢 |
| pages/task/ | 任务执行 | 无 | 🔴 |
| pages/taskTpl/ | 任务模板 | 无 | 🔴 |
| pages/taskOutput/ | 任务输出 | 无 | 🔴 |
| pages/log/ | 日志查看 | LogsPage | 🟡 |
| pages/logExplorer/ | 日志探索器 | LogsPage | 🟡 |
| pages/traceCpt/ | Trace 组件 | TracingPage | 🟡 |
| pages/chart/ | 图表组件 | 无独立页面 | 🟡 |
| pages/embeddedDashboards/ | 嵌入式仪表盘 | 无 | 🔴 |
| pages/embeddedProduct/ | 嵌入式产品 | 无 | 🔴 |
| pages/monitor/ | 监控总览 | 无 | 🔴 |

---

## 六、优化建议

### 6.1 Agent 下发页面优化

1. **去掉所有 BLOCKED_BY_CONTRACT 提示** — 用户不关心内部契约，要么能用要么不显示
2. **拆分步骤** — 安装向导改为步骤条（连通性检查→选包→配置→执行→验证），不要堆在一个 Drawer 里
3. **字段命名运维化** — "能力包"改为"采集器类型"，"配置模板"改为"采集配置"，"证据链"改为"数据验证"
4. **加入真实参数表单** — 根据选择的采集类型（MySQL/Redis/Nginx 等）动态生成配置表单
5. **加入实时日志** — 安装执行时显示 SSH 输出日志
6. **页面拆短** — HostsSection 280 行太长，拆分为独立组件

### 6.2 CMDB 页面优化

1. **资源列表加入 LWOPS 字段** — 业务名称、类型、子类型、分组、维护情况、负责人、采集情况
2. **加入筛选器** — 至少支持类型、采集情况、维护情况、IP 精准匹配
3. **加入批量操作** — 批量下发 Agent、批量修改分组、批量删除
4. **实例详情改为多 Tab 抽屉** — 属性/监控/关联/告警/变更/Agent 六个 Tab
5. **实现机房视图** — 机柜 U 位可视化
6. **实现自动发现** — 凭证+IP 范围+发现规则+执行记录

---

## 七、总结

### 实现率统计

| 维度 | 源码功能点 | FindX 已实现 | 实现率 |
|------|-----------|-------------|--------|
| n9e 后端路由 | 65 | 35（有壳） / 15（有深度） | 54% / 23% |
| n9e 告警引擎 | 11 模块 | 0 | 0% |
| n9e 通知 provider | 29 | 0 | 0% |
| n9e 数据源类型 | 9 | 1 | 11% |
| n9e 集成模板 | 65 | 0（数据未导入） | 0% |
| n9e 前端页面 | 38 | 20（有壳） / 8（有深度） | 53% / 21% |
| Categraf 插件配置下发 | 112 插件 | 0 | 0% |
| Catpaw 巡检下发 | 35 插件 | 0 | 0% |
| SkyWalking Agent 下发 | 5 语言 | 0 | 0% |
| SkyWalking UI 功能 | 14 | 8（代理） | 57% |
| SignOz 功能 | 12 | 5（框架） | 42% |
| LWOPS CMDB 功能 | 30+ 子页面 | 8（有内容） | 27% |
| 2级/3级抽屉场景 | 9 | 0 | 0% |

### 最终评价

**Codex 的工作模式是"广度优先"** — 快速铺设了大量路由和页面框架，但几乎没有深入任何一个模块的真实业务逻辑。这导致：

- 后端 279 个 handler 文件中，真正有业务深度的不到 50 个
- 前端看起来页面很多，但点进去大部分是空壳或只有列表
- Agent 下发页面设计过于抽象（"契约"、"证据链"），不符合运维人员的使用习惯
- CMDB 字段严重不足，缺少 LWOPS 的核心字段（类型/子类型/采集情况/维护情况/负责人）

**需要从"铺骨架"转向"填肌肉"** — 选择 3-5 个核心模块做到真正可用，而不是继续铺新的空壳。

---

## 八、推进改法（具体实施方案）

### 原则：代码逻辑一比一对齐源码，UI 风格用自己的

- 后端：参考 n9e 源码的数据结构、接口签名、业务逻辑，用 Go + Gin 重写（不是 copy，是理解后重实现）
- 前端：参考 n9e fe-main 的页面结构、表单字段、交互流程，用自己的 React + 自研 CSS 风格实现
- 探针下发：参考 n9e integrations/ 的模板结构，实现配置渲染 + SSH 推送 + 验证闭环

---

### 8.1 告警评估引擎（P0 第一优先）

**参考源码**：`nightingale-main/alert/eval/eval.go`

**n9e 核心逻辑**：
- `AlertRuleWorker` 结构体：每条规则一个 worker goroutine
- 使用 `robfig/cron` 定时调度（默认 15s 一次）
- 每次执行：取规则 → 构造 PromQL → 查询 Prometheus → 对比阈值 → 生成/更新事件
- 事件状态机：pending → firing → resolved
- 支持多查询条件 JOIN（left/right/inner）

**FindX 推进改法**：

| 步骤 | 改什么文件 | 参考源码 | 具体做什么 |
|------|-----------|---------|-----------|
| 1 | `api/internal/scheduler/alert_eval.go`（新建） | `alert/eval/eval.go` | 实现 AlertRuleWorker：每条规则一个 goroutine，cron 调度 |
| 2 | `api/internal/scheduler/alert_eval.go` | `alert/eval/eval.go` 的 `Work()` 方法 | 实现评估循环：取规则→查 Prometheus→对比阈值→生成事件 |
| 3 | `api/internal/model/monitoring_alert.go` | `models/alert_cur_event.go` | 补充事件字段：fingerprint、count、first_seen、last_seen、状态机 |
| 4 | `api/internal/store/monitoring_alert_events.go` | `models/alert_cur_event.go` | 实现事件持久化：upsert by fingerprint |
| 5 | `api/internal/scheduler/monitor_alert_scheduler.go` | `alert/eval/eval.go` 的 `Start()` | 启动时加载所有 enabled 规则，规则变更时热更新 worker |

**验证方式**：
```bash
# 1. 创建一条规则：cpu_usage_active > 80
# 2. 等待 15s
# 3. 查询 GET /api/v1/monitor/events/current 应该有事件产生
# 4. 指标恢复后事件应自动 resolved
```

---

### 8.2 通知发送引擎（P0 第二优先）

**参考源码**：`nightingale-main/alert/sender/sender.go` + `alert/sender/provider/`

**n9e 核心逻辑**：
- `Sender` 接口：`Send(ctx MessageContext)`
- `MessageContext`：包含 Users + Rule + Events
- 每种通知类型一个 Sender 实现（DingtalkSender、EmailSender、WebhookSender 等）
- 模板渲染：Go `html/template`，变量包含 `{{.RuleName}}`、`{{.Events}}`、`{{.Severity}}` 等
- Provider 注册表：`provider/registry.go` 统一管理

**FindX 推进改法**：

| 步骤 | 改什么文件 | 参考源码 | 具体做什么 |
|------|-----------|---------|-----------|
| 1 | `api/internal/notifier/sender.go`（新建） | `alert/sender/sender.go` | 定义 Sender 接口 + MessageContext |
| 2 | `api/internal/notifier/webhook.go`（新建） | `alert/sender/webhook.go` | 实现 Webhook 发送（POST JSON） |
| 3 | `api/internal/notifier/email.go`（新建） | `alert/sender/provider/email_provider.go` | 实现邮件发送（net/smtp） |
| 4 | `api/internal/notifier/dingtalk.go`（新建） | `alert/sender/dingtalk.go` | 实现钉钉机器人发送 |
| 5 | `api/internal/notifier/feishu.go`（新建） | `alert/sender/provider/feishuapp_provider.go` | 实现飞书发送 |
| 6 | `api/internal/notifier/template.go`（新建） | `alert/sender/sender.go` 的模板部分 | Go template 渲染通知内容 |
| 7 | `api/internal/notifier/dispatch.go`（新建） | `alert/dispatch/` | 事件→通知规则匹配→发送 |

**前端对应**：
| 步骤 | 改什么文件 | 参考源码 | 具体做什么 |
|------|-----------|---------|-----------|
| 1 | `web/src/react-shell/base-monitoring/notifications/ChannelFormDingtalk.jsx`（新建） | `fe-main/src/pages/notificationChannels/pages/Form/` | 钉钉配置表单（Webhook URL + 签名密钥） |
| 2 | `web/src/react-shell/base-monitoring/notifications/ChannelFormEmail.jsx`（新建） | 同上 | 邮件配置表单（SMTP 地址/端口/账号/密码） |
| 3 | `web/src/react-shell/base-monitoring/notifications/ChannelFormFeishu.jsx`（新建） | 同上 | 飞书配置表单（App ID + Secret + Webhook） |
| 4 | `web/src/react-shell/base-monitoring/notifications/ChannelFormWebhook.jsx`（新建） | 同上 | 通用 Webhook 表单（URL + Headers + Body 模板） |

---

### 8.3 告警屏蔽 + 订阅（P0 第三优先）

**参考源码**：`nightingale-main/models/alert_mute.go` + `fe-main/src/pages/warning/shield/`

**n9e 屏蔽逻辑**：
- 屏蔽条件：标签匹配（key=value / key=~regex）+ 时间窗口（开始/结束）+ 业务组
- 匹配引擎：事件产生时遍历所有 mute 规则，命中则不发送通知
- 前端：添加屏蔽表单（标签条件编辑器 + 时间选择器 + 备注）

**FindX 推进改法**：

| 步骤 | 改什么文件 | 具体做什么 |
|------|-----------|-----------|
| 1 | `api/internal/model/alert_mute.go`（新建） | 定义 AlertMute 结构体（ID/标签条件/时间窗口/创建人/备注） |
| 2 | `api/internal/store/alert_mutes.go`（新建） | CRUD + 按时间过期自动清理 |
| 3 | `api/internal/handler/alert_mute.go`（新建） | HTTP handler：列表/创建/删除 |
| 4 | `api/internal/notifier/dispatch.go` | 发送前检查 mute 规则，命中则跳过 |
| 5 | `web/src/react-shell/base-monitoring/alerts/MuteSection.jsx`（新建） | 屏蔽列表 + 添加表单（标签条件编辑器 + 时间选择器） |

**告警订阅同理**：
- 模型：AlertSubscribe（标签条件 + 通知渠道 + 升级策略）
- 逻辑：事件产生时匹配订阅规则，命中则额外发送给订阅人

---

### 8.4 Categraf 集成模板导入 + 配置下发（P0 第四优先）

**参考源码**：`nightingale-main/integrations/` + `categraf-main/inputs/`

**推进改法**：

| 步骤 | 改什么文件 | 具体做什么 |
|------|-----------|-----------|
| 1 | `api/assets/integrations/`（新建目录） | 从 n9e 源码复制 65 个集成目录（collect/ + dashboards/ + alerts/ + metrics/ + markdown/） |
| 2 | `api/internal/handler/integration_import.go`（新建） | 启动时扫描 integrations/ 目录，导入到 builtin_components + builtin_payloads 表 |
| 3 | `api/internal/handler/categraf_config_render.go`（新建） | .toml 模板渲染引擎：读取模板 → 替换占位符（address/username/password 等）→ 输出最终配置 |
| 4 | `api/internal/handler/categraf_deploy.go`（新建） | SSH 推送流程：连接目标 → 创建目录 → 写入 .toml → 执行 reload → 验证心跳 |
| 5 | `web/src/react-shell/base-monitoring/templates/DeployWizard.jsx`（新建） | 下发向导：选组件→填参数（动态表单）→选目标→确认→执行→验证 |

**配置渲染示例**（以 MySQL 为例）：

模板（`integrations/MySQL/collect/mysql/mysql.toml`）：
```toml
[[instances]]
address = "{{.address}}"
username = "{{.username}}"
password = "{{.password}}"
```

用户填写参数后渲染为：
```toml
[[instances]]
address = "192.168.1.84:3306"
username = "monitor"
password = "xxx"
```

SSH 推送到目标机：`/opt/categraf/conf/input.mysql/mysql.toml`

执行 reload：`systemctl reload categraf` 或 `kill -HUP $(pidof categraf)`

验证：查询 Prometheus `mysql_global_status_uptime{instance="192.168.1.84:9100"}` 有数据

---

### 8.5 CMDB 资源列表字段补齐（P1）

**参考**：LWOPS https://8.lwops.cn/app-monitor/resources 的表头和筛选器

**推进改法**：

| 步骤 | 改什么文件 | 具体做什么 |
|------|-----------|-----------|
| 1 | `api/internal/model/monitoring.go` | MonitorTarget 补充字段：business_name、classification（类型）、subtype（子类型）、group_name、maintenance_status、owner、collection_status |
| 2 | `api/internal/store/init_schema.go` | ALTER TABLE monitor_targets ADD COLUMN ... |
| 3 | `api/internal/handler/monitoring.go` | ListMonitorTargets 支持筛选参数：classification、subtype、maintenance_status、collection_status、keyword、ip（精准）、template_id、tag、business |
| 4 | `web/src/react-shell/cmdb/HostsSection.jsx` | 表头改为：对象名称/业务名称/IP/采集情况/类型/子类型/分组/维护情况/负责人 |
| 5 | `web/src/react-shell/cmdb/HostsSection.jsx` | 加入筛选器栏（9 个筛选条件，参考 LWOPS） |
| 6 | `web/src/react-shell/cmdb/HostsSection.jsx` | 操作列加入"更多"下拉：下发Agent/监控绑定/维护设置 |

---

### 8.6 通用 Drawer 组件（P1 基础设施）

**推进改法**：

| 步骤 | 改什么文件 | 具体做什么 |
|------|-----------|-----------|
| 1 | `web/src/react-shell/shared/Drawer.jsx`（新建） | 通用 Drawer：props 包含 width(sm/md/lg/xl/full)、title、footer、onClose、maskClosable |
| 2 | `web/src/react-shell/shared/Drawer.jsx` | 支持嵌套：2 级 Drawer 推出时 1 级半透明 |
| 3 | `web/src/react-shell/shared/Drawer.jsx` | 底部固定操作栏（保存/取消按钮固定在底部） |
| 4 | `web/src/react-shell/shared/Drawer.jsx` | URL 联动：打开时 pushState query 参数，关闭时 popState |
| 5 | `web/src/react-shell/shared/Drawer.jsx` | 未保存提示：关闭前检查 dirty 状态 |
| 6 | `web/src/react-shell/shared/StepDrawer.jsx`（新建） | 步骤条 Drawer：用于安装向导、配置下发等多步骤流程 |

---

### 8.7 告警规则全屏编辑抽屉（P1）

**参考源码**：`fe-main/src/pages/alertRules/Form/` 目录结构

n9e 告警规则表单有 5 个 Tab：
- `Base.tsx` — 基础信息（名称/级别/备注）
- `Rule/` — 触发条件（支持 9 种数据源的查询配置）
- `Effective/` — 生效时间（周几/时间段）
- `Notify/` — 通知配置（通知规则选择 + 任务模板）
- `EventSettings/` — 事件设置（附加标签/注解）

**FindX 推进改法**：

| 步骤 | 改什么文件 | 具体做什么 |
|------|-----------|-----------|
| 1 | `web/src/react-shell/base-monitoring/alerts/form/RuleFormDrawer.jsx`（新建） | 全屏 Drawer，内含 Tab 切换 |
| 2 | `web/src/react-shell/base-monitoring/alerts/form/BaseTab.jsx`（新建） | 基础信息 Tab：名称/级别/数据源/备注 |
| 3 | `web/src/react-shell/base-monitoring/alerts/form/TriggerTab.jsx`（新建） | 触发条件 Tab：PromQL 输入 + 阈值 + 持续时间 + 多条件 |
| 4 | `web/src/react-shell/base-monitoring/alerts/form/EffectiveTab.jsx`（新建） | 生效时间 Tab：周几选择 + 时间段 |
| 5 | `web/src/react-shell/base-monitoring/alerts/form/NotifyTab.jsx`（新建） | 通知配置 Tab：选择通知规则/渠道 |
| 6 | `web/src/react-shell/base-monitoring/alerts/form/LabelsTab.jsx`（新建） | 附加标签/注解 Tab |

---

### 8.8 仪表盘面板类型补齐（P1）

**参考源码**：`fe-main/src/pages/dashboard/Editor/Options/` + `Renderer/Renderer/`

n9e 有 13 种面板，FindX 至少需要补齐 6 种：

| 面板类型 | 参考源码 | FindX 改什么 |
|---------|---------|-------------|
| Gauge（仪表盘） | `Renderer/Renderer/Gauge/` | `web/src/react-shell/base-monitoring/dashboards/chartRenderers.jsx` 新增 |
| Stat（单值） | `Renderer/Renderer/Stat/` | 同上 |
| Table（表格） | `Renderer/Renderer/Table/` | 同上 |
| Pie（饼图） | `Renderer/Renderer/Pie/` | 同上 |
| BarChart（柱状图） | `Renderer/Renderer/BarChart/` | 同上 |
| Text/Markdown | `Renderer/Renderer/Text/` | 同上 |

---

### 8.9 事件流水线（P1）

**参考源码**：`nightingale-main/alert/pipeline/` + `fe-main/src/pages/eventPipeline/`

n9e 事件流水线有 6 种处理器：
- `relabel` — 标签重写（source_labels → target_label）
- `eventdrop` — 事件丢弃（按条件过滤）
- `eventupdate` — 事件更新（修改级别/标签/注解）
- `callback` — 回调（HTTP POST 到外部系统）
- `logic` — 逻辑判断（条件分支）
- `aisummary` — AI 摘要（调用 LLM 生成事件摘要）

**FindX 推进改法**：

| 步骤 | 改什么文件 | 具体做什么 |
|------|-----------|-----------|
| 1 | `api/internal/model/event_pipeline.go`（新建） | Pipeline + Processor 模型定义 |
| 2 | `api/internal/handler/event_pipeline.go`（新建） | Pipeline CRUD handler |
| 3 | `api/internal/notifier/pipeline_engine.go`（新建） | Pipeline 执行引擎：按顺序执行处理器 |
| 4 | `api/internal/notifier/processor_relabel.go`（新建） | relabel 处理器实现 |
| 5 | `api/internal/notifier/processor_drop.go`（新建） | eventdrop 处理器实现 |
| 6 | `api/internal/notifier/processor_callback.go`（新建） | callback 处理器实现 |
| 7 | `web/src/react-shell/base-monitoring/alerts/PipelineEditor.jsx`（新建） | 可视化 Pipeline 编辑器（处理器列表 + 拖拽排序 + 配置弹窗） |

---

### 8.10 Catpaw 巡检下发（P1）

**参考源码**：`catpaw-master/plugins/` + `catpaw-master/conf.d/`

**推进改法**：

| 步骤 | 改什么文件 | 具体做什么 |
|------|-----------|-----------|
| 1 | `api/assets/catpaw-plugins/`（新建） | 从 catpaw 源码整理 35 个插件的配置模板（.toml） |
| 2 | `api/internal/handler/catpaw_deploy.go`（新建） | SSH 安装 catpaw 二进制 + 推送插件配置 |
| 3 | `api/internal/handler/catpaw_deploy.go` | 安装流程：检查是否已安装 → 下载二进制 → 创建 systemd service → 启动 → 推送配置 → 验证心跳 |
| 4 | `web/src/react-shell/agents/CatpawDeployWizard.jsx`（新建） | 巡检下发向导：选插件→填参数→选目标→执行→验证 |

---

### 8.11 SkyWalking Agent 下发（P1）

**推进改法**：

| 步骤 | 改什么文件 | 具体做什么 |
|------|-----------|-----------|
| 1 | `api/assets/skywalking-agents/`（新建） | 各语言 Agent 配置模板（Java agent.config / Go env / Python ini） |
| 2 | `api/internal/handler/skywalking_deploy.go`（新建） | 配置生成：根据语言+服务名+OAP地址生成配置文件 |
| 3 | `api/internal/handler/skywalking_deploy.go` | SSH 推送：上传 Agent 包 + 配置文件 + 修改启动脚本注入 -javaagent |
| 4 | `web/src/react-shell/agents/SkywalkingDeployWizard.jsx`（新建） | 接入向导：选语言→填服务名→选目标→生成配置→执行→验证 Trace 到达 |

---

### 8.12 CMDB 实例详情多 Tab 抽屉（P1）

**参考**：LWOPS 实例详情页

**推进改法**：

| 步骤 | 改什么文件 | 具体做什么 |
|------|-----------|-----------|
| 1 | `web/src/react-shell/cmdb/InstanceDetailDrawer.jsx`（新建） | 全宽 Drawer，6 个 Tab |
| 2 | Tab 1：基础信息 | 所有属性字段表格（key-value 形式） |
| 3 | Tab 2：监控数据 | 嵌入 CPU/内存/磁盘/网络图表（调用 monitor/query） |
| 4 | Tab 3：关联关系 | 上下游拓扑图（D3 力导向） |
| 5 | Tab 4：告警历史 | 该实例的告警事件列表 |
| 6 | Tab 5：变更记录 | 属性变更审计日志 |
| 7 | Tab 6：Agent 状态 | Agent 心跳 + 配置下发入口 + 数据到达验证 |

---

### 8.13 PromQL 自动补全（P2）

**参考源码**：n9e fe-main 使用 Monaco Editor + Prometheus HTTP API

**推进改法**：

| 步骤 | 改什么文件 | 具体做什么 |
|------|-----------|-----------|
| 1 | 后端已有 `/api/v1/monitor/metrics` 和 `/api/v1/monitor/labels` | 确保返回完整指标名和标签列表 |
| 2 | `web/src/react-shell/base-monitoring/query/PromQLEditor.jsx`（新建） | 基于 Monaco Editor 的 PromQL 编辑器 |
| 3 | 同上 | 注册自动补全 provider：输入时请求 metrics/labels API |
| 4 | 同上 | 语法高亮：PromQL 关键字（rate/sum/avg/by/without 等） |

---

## 九、实施顺序（Sprint 规划）

### Sprint 1（2 周）— 告警引擎 + 通知发送

目标：让告警从"只能看规则"变成"真正能触发通知"

1. 告警评估引擎（8.1）
2. 通知发送 — Webhook + 邮件 + 钉钉（8.2）
3. 告警屏蔽基础实现（8.3）

验证：创建规则 → 触发告警 → 收到钉钉/邮件通知 → 屏蔽后不再通知

### Sprint 2（2 周）— Categraf 下发 + 集成模板

目标：让"资源列表"能一键下发采集配置

4. 65 个集成模板数据导入（8.4 步骤 1-2）
5. 配置渲染引擎（8.4 步骤 3）
6. SSH 推送 + reload（8.4 步骤 4）
7. 前端下发向导（8.4 步骤 5）

验证：选 MySQL 模板 → 填参数 → 选目标主机 → 执行 → Prometheus 出现 mysql_* 指标

### Sprint 3（2 周）— CMDB 对齐 + UI 层级

目标：资源列表对齐 LWOPS 字段，加入抽屉体系

8. 通用 Drawer 组件（8.6）
9. CMDB 字段补齐 + 筛选器（8.5）
10. 实例详情多 Tab 抽屉（8.12）
11. 告警规则全屏编辑抽屉（8.7）

### Sprint 4（2 周）— 面板 + 探针 + 流水线

12. 仪表盘面板类型补齐（8.8）
13. Catpaw 巡检下发（8.10）
14. SkyWalking Agent 下发（8.11）
15. 事件流水线基础实现（8.9）

### Sprint 5（1 周）— 收尾

16. PromQL 自动补全（8.13）
17. 告警订阅
18. 通知记录查询
19. 记录规则管理
