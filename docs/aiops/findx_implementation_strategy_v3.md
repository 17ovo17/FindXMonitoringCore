# FindX 全栈可观测平台 — 完整实施计划 v4

更新时间：2026-05-11 21:30（UTC+8）

状态：当前唯一实施主计划。替代 v3 和原有长期计划。

---

## 1. 总则

FindX 基于成熟平台源码复用，统一替换为 FindX 自有品牌、风格、权限、审计、配置、数据源和 AI SRE 能力。

**不可违背：**
- 成熟源码是事实源，不自研弱化页面
- 前端 React-only，保持现有 --fx-* CSS 变量 UI 风格
- 禁止 iframe/WebView/参考站嵌入
- 禁止 MVP/假按钮/假数据/假成功
- 契约缺失显示 BLOCKED_BY_CONTRACT
- 真实凭据不输出
- Agent 生命周期必须真实闭环
- AI 基于证据，不编造结论

## 2. 技术架构决策

### 2.1 持久化层：GORM

- 新模块（CMDB 对象建模、MCP server、Agent 生命周期）使用 GORM GetDB() 模式
- 现有模块（监控、告警、通知）保持原生 SQL，不强制迁移
- GORM 支持 AutoMigrate、JSON 字段、Preload/Association，适合动态 schema

### 2.2 AI 架构：MCP 插线板

```
FindX AI 助手 (LLM + Tool Calling)
    ↓ MCP Protocol
┌──────────────────────────────────────────────────┐
│  夜莺 MCP Server (官方开源)                       │
│  ├── 读写：指标查询、仪表盘、告警规则、数据源     │
├──────────────────────────────────────────────────┤
│  prometheus-mcp-server (开源备用)                 │
│  alertmanager-mcp-server (开源备用)               │
├──────────────────────────────────────────────────┤
│  findx-cmdb-mcp-server (自建)                     │
│  ├── 读写：资产 CRUD、关联、负责人、拓扑          │
├──────────────────────────────────────────────────┤
│  findx-agent-mcp-server (自建)                    │
│  ├── 读写：探针部署、配置下发、状态、心跳验证     │
├──────────────────────────────────────────────────┤
│  findx-knowledge-mcp-server (自建)                │
│  ├── 读写：知识库、Runbook                        │
└──────────────────────────────────────────────────┘
```

### 2.3 前端：React-only + 成熟源码迁移

参考源码的组件结构、状态流、API 调用模式，用 FindX JSX + --fx-* CSS 变量实现。UI 风格不变。

## 3. 成熟源码事实源

| 域 | 事实源 | 参考内容 |
|------|------|------|
| 基础监控前端 | `D:\项目迁移文件\平台源码\fe-main` | 数据源表单、仪表盘编辑器、告警规则、通知、指标查询 |
| 链路监控前端 | `D:\项目迁移文件\平台源码\skywalking-booster-ui-main` | 服务目录、拓扑、Trace、Profiling |
| 链路监控后端 | `D:\项目迁移文件\平台源码\skywalking-master` | OAP query、GraphQL、存储模型 |
| 日志中心前端 | `D:\项目迁移文件\平台源码\signoz-develop\frontend` | 日志检索、字段筛选、Pipeline、Saved Views |
| CMDB/Agent | `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main` | 主机表、分组、Agent 在线、部署、心跳 |
| 采集插件 | `D:\项目迁移文件\平台源码\categraf-main (1)` | 插件目录、配置模板 |
| 巡检诊断 | `D:\项目迁移文件\平台源码\catpaw-master` | 巡检执行、结构化报告 |
| CMDB 对象建模 | `D:\测试\LWOPS_*\reverse-poc\public\captures\cmdb-*` | 模型分类树、30+属性、关联、自动发现、监控绑定、拓扑、审批、数据质量 |
| AI 智能问答 | `D:\测试\LWOPS_*\reverse-poc\public\captures\lerwee-*` | 31类预定义问题、Tool Calling、SSE 流式 |
| AI 告警预测 | `D:\测试\LWOPS_*\reverse-poc\public\captures\calculate-*` | 趋势预测、异常检测 |
| 告警中心 | `D:\测试\LWOPS_*\modules\aialert\` | 统一告警接入、SNMP Trap、webhook |
| 导航/菜单 | `D:\测试\LWOPS_*\reverse-poc\public\captures\menu-tree.json` | 194项菜单、21应用模块 |

## 4. CMDB 对象建模（参考 LWOPS，GORM 实现）

### 4.1 模型分类树

```
计算资源
├── 服务器（X86ServerBasic）
├── 虚拟化（平台/宿主机/虚拟机/存储）
├── 容器（Docker/K8s）
└── 云平台（ECS/RDS/EIP/ALB/Redis + 阿里云/腾讯云/火山引擎）

系统软件
├── 操作系统
├── 数据库（MySQL/PostgreSQL/MongoDB/Redis/达梦/GaussDB/Kingbase）
└── 中间件（Tomcat/Nginx/Apache/WebLogic/JBoss/Kafka/RabbitMQ）

应用软件（业务系统/子系统/云应用）
网络资源（网络设备/交换机/路由器/防火墙/负载均衡）
存储资源
机房资源（机房/机柜）
组织人员（部门/用户）
其他（打印机/物联网/域名/License/合同）
```

### 4.2 用户自定义能力

- 用户可创建自定义模型（自定义分类、名称、图标）
- 用户可为任何模型添加自定义属性（字段名、类型、是否必填、是否唯一、是否自动发现）
- 属性类型：char/int/float/ip/boolean/enum/array/struct/file/image
- 属性分组标签（基本信息/系统资源/系统信息/自定义）
- 关联关系定义（belong/default/自定义，n:1/1:n/n:n）

### 4.3 实例字段体系（操作系统为例，30+ 字段）

| 分组 | 字段 | 类型 | 自动发现 |
|------|------|------|---------|
| 基本信息 | IP地址 | ip | ✅ |
| 基本信息 | 名称 | char | ✅ |
| 基本信息 | 是否监控 | boolean | ❌ |
| 基本信息 | 分组 | array(enum) | ✅ |
| 基本信息 | 系统版本 | char | ✅ |
| 基本信息 | 系统信息(uname) | char | ✅ |
| 基本信息 | 资产负责人 | char | ❌ |
| 基本信息 | 联系电话 | char | ❌ |
| 基本信息 | 系统运行天数 | int | ✅ |
| 基本信息 | NTP/DNS 状态 | enum | ❌ |
| 基本信息 | 厂商 | char | ❌ |
| 基本信息 | 附件/图片 | file/image | ❌ |
| 系统资源 | 内存大小 | float(B) | ✅ |
| 系统资源 | 虚拟内存 | int(B) | ✅ |
| 系统资源 | 网卡数量 | int | ✅ |
| 系统资源 | 磁盘空间 | float(B) | ✅ |
| 系统资源 | 文件系统 | struct[] | ✅ |
| 系统资源 | 最大进程数 | int | ✅ |
| 系统信息 | 所属业务 | array(ref) | ❌ |
| 系统信息 | 管理IP | ip | ❌ |
| 系统信息 | 系统属性 | enum(物理/虚拟) | ❌ |
| Agent | 探针版本 | char | ✅ |
| Agent | 心跳时间 | datetime | ✅ |
| Agent | 配置版本 | char | ✅ |
| Agent | 数据到达时间 | datetime | ✅ |

### 4.4 关键能力

- 自动发现：catpaw/categraf 上报 → CMDB 字段自动填充
- 变更审计：所有字段变更记录
- 监控绑定：实例 ↔ Prometheus 目标
- 拓扑视图：实例间关联可视化
- 数据质量：巡检规则检查字段完整性（后续）
- 审批流：实例创建/变更走审批（后续）

## 5. Agent 完整生命周期

### 5.1 能力包

SkyWalking Agent（Java/Python/Node.js/PHP/Go/Rust/Ruby/Nginx Lua/Kong/Browser JS）+ Categraf + Catpaw

### 5.2 安装方式

- 本机：Linux curl -kfsSL、Windows CMD certutil、PowerShell Invoke-WebRequest
- 远程：SSH/shell/systemd、WinRM/PowerShell/Windows Service
- K8s：Helm/Operator/DaemonSet/Sidecar/InitContainer

### 5.3 生命周期闭环

安装 → 配置下发 → 插件下发 → 心跳验证 → 数据到达验证 → 升级 → 回滚 → 卸载 → 审计 → Evidence Chain

每一步必须有真实执行证据，不能用命令预览或 BLOCKED_BY_CONTRACT 冒充完成。

### 5.4 心跳验证要求

- 控制面心跳（Agent → 平台 API）
- 进程状态（systemd/Windows Service 存活）
- 服务状态（端口监听、健康检查）
- 探针加载状态
- OAP/Prometheus 连通性
- 最后 Trace/Metric/Log/RUM 到达时间

## 6. AI SRE + Evidence Chain

### 6.1 AI 助手（参考 LWOPS Lerwee）

- 预定义问题模板（31+ 分类：监控/告警/资产/网络/系统/通用）
- Tool Calling 通过 MCP 查询真实数据
- SSE 流式输出
- 会话管理

### 6.2 AI 告警预测（参考 LWOPS calculate）

- 趋势预测任务（CPU/磁盘/内存使用率）
- 单指标异常检测
- 预测告警列表和统计

### 6.3 Evidence Chain

接入：指标、告警、日志、Trace、CMDB、Agent 心跳、安装任务、配置下发、巡检、工作流、知识库

## 7. 导航结构

```
工作台（概览、AI 助手）
资产中心（资产概览、对象建模、实例管理、Agent 状态、拓扑视图）
集成中心（数据源、模板中心、系统集成）
数据查询（指标查询、仪表盘）
告警（规则管理、告警屏蔽、告警订阅、告警事件、历史事件、链路告警、事件流水线）
通知（通知规则、通知媒介、消息模板）
链路监控（链路总览、服务目录、服务拓扑、Trace 检索、Profiling、链路设置）
日志中心（日志检索、实时日志、字段筛选、上下文、聚合分析、接入管道、保存视图、Trace 关联）
人员组织（用户管理、团队组织、业务组、角色管理）
系统配置（AI 模型配置、站点设置、变量设置、单点登录、告警引擎、运行自检、审计日志）
AI SRE（诊断会话、工作流、健康检查、复盘报告、Evidence Chain、知识库、自动修复）
```

## 8. 实施阶段

### P0：工作树治理 ✅ 已完成

### P1：基础监控完整（参考夜莺 fe-main）

| 模块 | 参考源码 | 内容 |
|------|---------|------|
| 数据源表单 | fe-main/src/pages/datasource | Prometheus/ES/Loki 创建编辑表单、Auth、TLS、Headers |
| 仪表盘编辑器 | fe-main/src/pages/dashboard | Panel 配置、变量、模板导入导出、图表渲染 |
| 指标查询增强 | fe-main/src/pages/explorer | 自动补全、内置指标库、记录规则 |
| 告警规则 | fe-main/src/pages/alertRules | 规则创建/编辑、PromQL 条件、持续时间、标签 |
| 告警事件 | fe-main/src/pages/alertCurEvent | 实时事件列表、确认、屏蔽、历史 |
| 通知配置 | fe-main/src/pages/notificationChannels | 通知媒介、模板、测试发送 |
| 模板中心 | fe-main/src/pages/builtInComponents | 内置仪表盘模板、告警规则模板 |

### P2：CMDB + Agent（参考 AutoOps + LWOPS，GORM）

| 模块 | 参考源码 | 内容 |
|------|---------|------|
| GORM 基础设施 | — | 引入 GORM、GetDB()、AutoMigrate |
| 对象建模后端 | LWOPS captures | 分类树、模型 CRUD、属性定义、关联关系（GORM） |
| 对象建模前端 | LWOPS captures | 模型分类树 UI、属性编辑器、用户自定义模型/属性 |
| 实例管理 | LWOPS captures + AutoOps | 实例列表、详情、创建/编辑、搜索、分组 |
| 自动发现 | AutoOps + catpaw | Agent 上报 → CMDB 字段自动填充 |
| 监控绑定 | LWOPS captures | 实例 ↔ Prometheus 目标关联 |
| Agent 心跳验证 | AutoOps | 真实心跳验证（不信任当前数据，需测试） |
| Agent 部署 | AutoOps + catpaw | 本机/远程/K8s 安装、配置下发 |
| 拓扑视图 | LWOPS captures | 实例间关联可视化 |
| 变更审计 | LWOPS captures | 字段变更记录 |

### P3：链路监控（参考 SkyWalking）

| 模块 | 参考源码 | 内容 |
|------|---------|------|
| 服务目录 | skywalking-booster-ui-main | 服务列表、健康状态、覆盖率 |
| 服务拓扑 | skywalking-booster-ui-main | 拓扑图、节点详情、调用关系 |
| Trace 检索 | skywalking-booster-ui-main | Trace 列表、Span 详情、时间线 |
| Profiling | skywalking-booster-ui-main | CPU/内存 profiling |
| 链路告警 | skywalking-booster-ui-main | 链路相关告警规则 |
| Agent 联动 | — | 服务目录 ↔ Agent 状态、Trace ↔ 探针状态 |

### P4：日志中心（参考 SigNoZ）

| 模块 | 参考源码 | 内容 |
|------|---------|------|
| 日志检索 | signoz-develop/frontend | 全文搜索、字段筛选、时间范围 |
| 实时日志 | signoz-develop/frontend | Live tail |
| 聚合分析 | signoz-develop/frontend | 字段聚合、统计图表 |
| Pipeline | signoz-develop/frontend | 日志接入管道配置 |
| Saved Views | signoz-develop/frontend | 保存查询视图 |
| Trace 关联 | signoz-develop/frontend | 日志 ↔ Trace 关联 |

### P5：FindX Agent 完整生命周期

| 模块 | 内容 |
|------|------|
| 包仓库 | 包名、版本、平台、架构、签名、校验和、离线包 |
| 安装向导 | 选择主机、能力包、配置模板、凭据引用、安装前检查 |
| 本机安装 | Linux curl / Windows certutil / PowerShell |
| 远程安装 | SSH/WinRM/K8s Helm/Operator/DaemonSet |
| 配置下发 | 单机/批量/灰度/回滚/漂移检测 |
| 插件远程修改 | Categraf 插件配置远程修改、下发、回滚、审计 |
| 心跳+数据到达 | 真实验证，不信任静态状态 |
| 升级/回滚/卸载 | 完整生命周期 + Evidence Chain |

### P6：AI SRE + MCP（参考 LWOPS Lerwee + calculate）

| 模块 | 内容 |
|------|------|
| MCP 基础设施 | server 注册、配置、健康检查、权限 |
| 夜莺 MCP Server | 读写：指标、仪表盘、告警、数据源 |
| findx-cmdb-mcp-server | 资产 CRUD、关联、负责人 |
| findx-agent-mcp-server | 探针部署、配置下发、状态 |
| findx-knowledge-mcp-server | 知识库、Runbook |
| AI 助手 UI | 对话界面、31类预定义问题、SSE 流式、Tool Calling |
| AI 告警预测 | 趋势预测、异常检测 |
| Evidence Chain | 全链路证据接入 |
| 诊断会话 | 基于证据的诊断、解释、编排 |
| 知识库 | MySQL + Qdrant + BM25 + RRF |

### P7：组织权限治理

组织、团队、角色、权限、SSO、审计日志、站点设置

### P8：商业化补齐

SLO/SLA、事件管理、值班升级、变更关联、Synthetic Monitoring、云资源观测、K8s 观测、网络观测、RUM、数据治理

## 9. 验证标准

| 场景 | 要求 |
|------|------|
| 平台后端日常 | 远端 Ubuntu: go test + go build |
| 平台后端里程碑 | 远端 Ubuntu: test + build + deploy + health + Playwright |
| 探针代码 | Windows + Linux 双 build + 双 deploy + 心跳/数据到达验证 |
| 前端日常 | npm run build |
| 前端里程碑 | build + Playwright 真实登录 http://10.10.160.202:3000 |
| Agent 心跳 | 不信任静态数据，必须真实验证进程存活+API 心跳+数据到达 |

## 10. 环境信息

| 环境 | 用途 | 地址 |
|------|------|------|
| 远端 Ubuntu | 唯一 Linux 验证环境 | SSH: findx-ubuntu, Web: http://10.10.160.202:3000 |
| 远端源码 | /opt/ai-workbench | runtime: /opt/ai-workbench-runtime |
| 远端服务 | ai-workbench-api.service | API: :8080, Web: :3000 |
| Windows 本机 | 开发 + 探针双平台验证 | D:\ai-workbench |
| WSL | 已废弃 | — |
