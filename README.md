# FindX Monitoring Core

FindX 是面向企业运维和可观测场景的一体化监控平台，覆盖指标监控、日志中心、链路追踪、CMDB 资产管理、Agent 生命周期、AI SRE 诊断和 Evidence Chain。

## 核心特性

- **统一可观测**：指标、日志、Trace 三大支柱统一采集、存储、查询和告警
- **智能诊断**：AI SRE 工作流自动关联证据链，辅助根因定位
- **Agent 管理中心**：采集引擎全生命周期管理（安装、配置下发、升级、回滚、卸载）
- **巡检探针**：结构化巡检任务编排与执行，支持定时和事件触发
- **CMDB 资产**：模型驱动的资产管理，支持拓扑关系和监控绑定
- **仪表盘**：多类型面板（折线图、柱状图、热力图、饼图、表格、文本）
- **告警中心**：规则引擎、事件聚合、屏蔽、订阅、自愈和通知媒介
- **链路后端**：服务拓扑、Trace 检索、Span 详情、Profiling
- **指标引擎**：PromQL 查询、内置指标库、模板中心
- **智能网关**：多协议数据接入与路由

## 架构概览

```
┌─────────────────────────────────────────────────────────┐
│                    React 前端壳层                         │
│  (仪表盘 / 告警 / 链路 / 日志 / CMDB / Agent / AI SRE)   │
└────────────────────────┬────────────────────────────────┘
                         │ HTTP/WebSocket
┌────────────────────────▼────────────────────────────────┐
│                   Go API 服务层                           │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐  │
│  │ Handler  │ │Middleware│ │ Service  │ │  Store    │  │
│  └──────────┘ └──────────┘ └──────────┘ └───────────┘  │
└────┬──────────────┬──────────────┬──────────────┬───────┘
     │              │              │              │
┌────▼───┐   ┌─────▼────┐  ┌─────▼────┐  ┌─────▼────┐
│Prometheus│  │ClickHouse│  │  MySQL   │  │  Redis   │
│(指标引擎)│  │(日志/链路)│  │(元数据)  │  │(缓存/队列)│
└─────────┘  └──────────┘  └──────────┘  └──────────┘
```

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+ / npm 9+
- MySQL 8.0+
- Redis 6+
- Prometheus (指标存储)

### 后端启动

```bash
cd api
go build -o findx-api ./...
./findx-api -c config.yaml
```

### 前端启动

```bash
cd web
npm install
npm run dev
```

访问 `http://localhost:5173` 进入平台。

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端框架 | React 19 + Vite |
| UI 组件 | Ant Design 5 |
| 图表 | uPlot + SVG 自绘 |
| 代码编辑 | Monaco Editor |
| 后端框架 | Go + Gin |
| ORM | GORM |
| 时序存储 | Prometheus / VictoriaMetrics |
| 日志存储 | ClickHouse |
| 关系存储 | MySQL |
| 缓存 | Redis |

## 目录结构

```
findx-monitoring-core/
├── api/                          # Go 后端
│   ├── internal/
│   │   ├── handler/              # HTTP 路由处理
│   │   ├── model/                # 数据模型
│   │   ├── store/                # 数据持久化
│   │   ├── middleware/           # 中间件（认证、审计）
│   │   ├── workflow/             # AI SRE 工作流
│   │   ├── tracing/             # 链路后端
│   │   ├── monitoring/          # 指标引擎
│   │   ├── knowledge/           # 知识库
│   │   └── scheduler/           # 巡检调度
│   └── go.mod
├── web/                          # React 前端
│   ├── src/
│   │   └── react-shell/
│   │       ├── base-monitoring/ # 指标监控、仪表盘、告警
│   │       ├── tracing/         # 链路追踪
│   │       ├── logs/            # 日志中心
│   │       ├── cmdb/            # CMDB 资产
│   │       ├── agents/          # Agent 管理中心
│   │       ├── ai-sre/          # AI SRE 诊断
│   │       ├── platform/        # 平台设置
│   │       └── shared/          # 公共组件
│   └── package.json
└── scripts/                      # 构建与验证脚本
```

## 模块说明

| 模块 | 说明 |
|------|------|
| 基础监控 | 数据源管理、指标查询、仪表盘、内置模板 |
| 告警中心 | 告警规则、记录规则、事件、屏蔽、订阅、通知 |
| 链路追踪 | 服务目录、拓扑、Trace 检索、Profiling |
| 日志中心 | Logs Explorer、字段筛选、聚合、Saved Views |
| CMDB | 模型管理、实例、关系拓扑、监控绑定 |
| Agent 管理 | 采集引擎安装、配置下发、插件管理、生命周期 |
| 巡检探针 | 定时巡检、事件触发、结构化报告 |
| AI SRE | 诊断会话、证据链、工作流、健康检查 |
| 智能网关 | 多协议接入、数据路由、心跳管理 |

## 开发指南

```bash
# 后端构建与检查
cd api && go build ./... && go vet ./...

# 前端构建
cd web && npm run build

# 运行测试
cd api && go test ./...
cd web && npm test
```

## 许可证

MIT License
