# FindX 平台实施策略 v3

更新时间：2026-05-11 20:00（UTC+8）

状态：当前唯一实施主计划。替代原有逐文件闭包策略。

## 1. 策略变更摘要

| 维度 | 旧策略 | 新策略 |
|------|--------|--------|
| 验证环境 | Windows + WSL + 远端 Ubuntu 三环境 | 平台后端只需远端 Ubuntu；探针代码才需 Windows+Linux 双通过 |
| P0 收敛粒度 | 逐文件闭包，每个 2-3 文件拆函数到 ≤50 行 | 批量提交（编译通过+无敏感泄露），后续功能开发中顺带重构 |
| 验证深度 | 每个闭包都跑 focused+full test + build + deploy + health | 日常只需 go test + go build；里程碑节点才做完整 deploy+health+browser |
| 推进节奏 | 单文件闭包 → 提交 → 下一个文件 | 按功能模块批量推进，前后端联动闭环 |
| CMDB | 基础设施 + Agent 管理中心分离 | 合并为"资产中心"，完整对象建模 |
| AI 架构 | catpaw 巡检对话 | MCP 插线板 + 全平台 AI 助手 |

## 2. 不变的原则

- 成熟源码是事实源，不自研弱化页面
- 前端最终架构为 React-only
- 禁止 iframe/WebView/参考站嵌入
- 禁止 MVP/假按钮/假数据/假成功
- 契约缺失时显示 BLOCKED_BY_CONTRACT，不伪装完成
- 真实凭据不输出，使用占位符
- FindX Agent 生命周期必须最终做到真实闭环
- AI SRE 基于已有证据，不编造结论

## 3. 验证标准

### 平台后端（Go API）

| 场景 | 验证要求 |
|------|---------|
| 日常开发 | 远端 Ubuntu: `go test ./...` + `go build` |
| 里程碑提交 | 远端 Ubuntu: test + build + deploy + health + Playwright browser |
| API 契约变更 | 远端 Ubuntu: test + build + deploy + health + 相关 API curl |

### 探针代码（catpaw/agent 相关）

| 场景 | 验证要求 |
|------|---------|
| 日常开发 | Windows: `go test` + 远端 Ubuntu: `go test` |
| 里程碑提交 | Windows + Ubuntu 双 build + 双 deploy + 心跳/数据到达 |

### 前端（React）

| 场景 | 验证要求 |
|------|---------|
| 日常开发 | `npm run build` 通过 |
| 里程碑提交 | build + Playwright 真实登录远端 http://10.10.160.202:3000 |

### 里程碑节点定义

- 完成一个功能模块
- 准备 push 到远端仓库
- 用户明确要求完整验证

## 4. 架构设计

### 4.1 资产中心（CMDB 统一化）

原"基础设施"+"Agent 管理中心"合并为**资产中心**。

基于 LWOPS API 抓包反向设计，采用完整对象建模：

```
资产中心
├── 对象建模（模型分类树、模型 CRUD、属性定义、关联关系）
├── 实例管理（列表、详情、创建/编辑、搜索、分组）
├── Agent 集成（探针状态、部署、心跳、配置下发）
├── 自动发现（catpaw/categraf 上报 → CMDB 字段自动填充）
├── 监控绑定（实例 ↔ Prometheus 目标关联）
└── 拓扑视图（实例间关联可视化）
```

对象建模体系（参考 LWOPS）：

```
计算资源
├── 服务器
├── 虚拟化（平台/宿主机/虚拟机/存储）
├── 容器（Docker/K8s）
└── 云平台（ECS/RDS/EIP/ALB/Redis + 多云）

系统软件
├── 操作系统
├── 数据库
└── 中间件

应用软件（业务系统/子系统）
网络资源（网络设备）
存储资源
机房资源（机房/机柜）
组织人员（部门/用户）
```

实例字段体系（操作系统为例，30+ 字段）：
- 基本信息：IP、名称、是否监控、分组、系统版本、uname、负责人、联系电话、运行天数、NTP/DNS 状态、厂商
- 系统资源：内存、虚拟内存、网卡数、磁盘空间、文件系统(结构化)、最大进程数
- 系统信息：所属业务、管理IP、系统属性(物理/虚拟)
- Agent 状态：探针版本、心跳时间、配置版本、数据到达时间

### 4.2 AI 架构（MCP 插线板）

AI SRE 通过 MCP 协议标准化接入所有数据源，不硬编码查询逻辑：

```
FindX AI SRE (LLM)
    ↓ MCP Protocol (标准化工具调用)
┌─────────────────────────────────────────┐
│  MCP Server 插线板                       │
├─────────────────────────────────────────┤
│  prometheus-mcp-server (开源)            │ ← PromQL 查询、指标历史
│  alertmanager-mcp-server (开源)          │ ← 告警查询、确认、静默
│  findx-cmdb-mcp-server (自建)            │ ← 资产查询、关联、负责人
│  findx-agent-mcp-server (自建)           │ ← 探针状态、触发部署
│  findx-knowledge-mcp-server (自建)       │ ← 知识库、Runbook
└─────────────────────────────────────────┘
```

AI 助手能力（参考 LWOPS Lerwee，但用 MCP 实现）：
- 监控助手：查指标、查历史、查趋势
- 告警助手：查未恢复告警、告警统计、告警 TOP N
- 资产助手：查资产详情、查负责人、查业务归属
- 预测能力：趋势预测、单指标异常检测

### 4.3 告警中心增强

- 统一告警接入 API（支持第三方 webhook 推送，参考 LWOPS appid+token+签名模式）
- 告警等级：5 级（信息/警告/次要/严重/紧急）
- 告警生命周期：触发 → 确认 → 恢复 → 关闭
- 告警与 CMDB 关联：告警对象 → 资产实例 → 负责人 → 自动通知
- AI 告警预测：基于 Prometheus 历史数据做趋势预测

### 4.4 导航结构

```
工作台（概览、AI 助手）
资产中心（对象建模、实例管理、Agent 状态、拓扑）
监控中心（数据源、仪表盘、指标查询、模板中心）
告警中心（实时告警、告警规则、告警预测、通知）
链路监控（服务目录、拓扑、Trace、Profiling）
日志中心（日志检索、Pipeline、Saved Views）
AI SRE（诊断会话、Evidence Chain、知识库）
系统管理（组织权限、系统配置、审计日志）
```

## 5. 实施阶段

### 阶段 A：P0 快速收尾（当前）

- 批量提交剩余后端脏文件（编译通过+无敏感即可）
- 提交本策略文档
- P0 工作树治理完成

### 阶段 B：React Shell + 资产中心

| 序号 | 模块 | 说明 | 依赖 |
|------|------|------|------|
| B1 | React Shell + 登录 | 壳层、路由、登录、主题、全局错误 | 无 |
| B2 | 导航 + 权限 | 侧边栏、面包屑、角色权限、路由守卫 | B1 |
| B3 | 资产中心 - 对象建模 | 模型分类树、模型 CRUD、属性定义、关联关系 | B2 |
| B4 | 资产中心 - 实例管理 | 实例列表、详情、创建/编辑、搜索、分组 | B3 |
| B5 | 资产中心 - Agent 集成 | 探针状态、部署、心跳、配置下发 | B4 |
| B6 | 资产中心 - 自动发现 | catpaw/categraf 上报 → CMDB 字段自动填充 | B5 |
| B7 | 资产中心 - 监控绑定 | 实例 ↔ Prometheus 目标关联 | B4 |

### 阶段 C：监控 + 告警

| 序号 | 模块 | 说明 | 依赖 |
|------|------|------|------|
| C1 | 数据源中心 | 数据源 CRUD、测试连接、凭据引用 | B2 |
| C2 | 仪表盘 | 列表、详情、Panel、变量、模板导入 | C1 |
| C3 | 指标查询 | 指标浏览器、PromQL、内置指标 | C1 |
| C4 | 告警中心 | 告警规则、事件、屏蔽、第三方接入 | C1 |
| C5 | 通知 | 通知媒介、消息模板、测试发送 | C4 |

### 阶段 D：AI + MCP

| 序号 | 模块 | 说明 | 依赖 |
|------|------|------|------|
| D1 | MCP 基础设施 | MCP server 注册、配置、健康检查 | B2 |
| D2 | prometheus-mcp-server 接入 | 开源 server 部署 + FindX 集成 | D1, C1 |
| D3 | alertmanager-mcp-server 接入 | 开源 server 部署 + FindX 集成 | D1, C4 |
| D4 | findx-cmdb-mcp-server | 自建：资产查询、关联、负责人 | D1, B4 |
| D5 | AI 助手 UI | 对话界面、预定义问题、SSE 流式 | D2-D4 |
| D6 | AI 告警预测 | 趋势预测任务、异常检测 | D2, C4 |

### 阶段 E：垂直能力

| 序号 | 模块 | 说明 |
|------|------|------|
| E1 | 链路监控 | SkyWalking 服务目录、拓扑、Trace、Profiling |
| E2 | 日志中心 | SigNoZ 日志检索、字段筛选、Pipeline |
| E3 | FindX Agent 完整生命周期 | 包仓库、安装、配置下发、升级、回滚、卸载 |
| E4 | AI SRE 完整 | 诊断会话、工作流、Evidence Chain |

### 阶段 F：治理与发布

- 变更审计、数据质量巡检、审批流
- 完整测试覆盖
- 文档完善、发布流程

## 6. 提交规范

### 日常提交

```bash
# 远端 Ubuntu 验证
ssh findx-ubuntu "cd /opt/ai-workbench/api && go test -count=1 ./... && go build -o api-linux ."
# 按 pathspec 提交
git add -- <具体文件列表>
git commit -m "<type>: <description>"
```

### 里程碑提交

```bash
# 远端完整验证
ssh findx-ubuntu "cd /opt/ai-workbench/api && go test -count=1 ./... && go build -o api-linux . && sudo install -m 0755 api-linux /opt/ai-workbench-runtime/api/ai-workbench-api && sudo systemctl restart ai-workbench-api.service"
ssh findx-ubuntu "sleep 4 && curl -fsS http://127.0.0.1:8080/api/v1/health/storage"
# Playwright 浏览器验证（如有 UI 变更）
```

### 禁止提交清单

- `.codex/**`、`.claude/**`、`.playwright-mcp/**`、`.test-evidence/**`
- `api/data/**`、`logs/**`、`web/dist/**`、`web/node_modules/**`
- `*.pem`、`*.key`、`*.log`、`*.exe`
- `api/internal/handler/data/memory-store.json`
- runtime data、真实凭据

## 7. 技术债管理

函数行数超标、历史乱码等不再阻塞提交：
- 功能开发触及相关文件时顺带修复
- 不专门开闭包处理纯重构
- 敏感信息泄露必须立即修复

## 8. 成熟源码事实源

| 域 | 事实源路径 |
|------|------|
| 基础监控 | `D:\项目迁移文件\平台源码\fe-main` |
| 链路监控 | `D:\项目迁移文件\平台源码\skywalking-booster-ui-main` |
| 日志中心 | `D:\项目迁移文件\平台源码\signoz-develop\frontend` |
| CMDB/Agent | `D:\项目迁移文件\平台源码\AutoOps-main\AutoOps-main` |
| 采集插件 | `D:\项目迁移文件\平台源码\categraf-main (1)` |
| 巡检诊断 | `D:\项目迁移文件\平台源码\catpaw-master` |
| CMDB 对象建模参考 | `D:\测试\LWOPS_安全测试资料_2026-05-10\reverse-poc\public\captures\` |

## 9. 环境信息

| 环境 | 用途 | 地址 |
|------|------|------|
| 远端 Ubuntu | 唯一 Linux 验证环境 | SSH: findx-ubuntu, Web: http://10.10.160.202:3000 |
| 远端源码 | /opt/ai-workbench | runtime: /opt/ai-workbench-runtime |
| 远端服务 | ai-workbench-api.service | API: :8080, Web: :3000 |
| Windows 本机 | 开发 + 探针双平台验证 | D:\ai-workbench |
| WSL | 已废弃 | — |

## 10. 安全边界（已知风险，不阻塞开发）

- `ssh.InsecureIgnoreHostKey()` — 生产需替换
- WinRM Basic/AllowUnencrypted — 明文传输
- WMI `Win32_Process.Create` — 只证明进程启动
- password 注入 PowerShell 脚本 — 进程列表可见
- 以上在阶段 E3 Agent 完整生命周期时解决
