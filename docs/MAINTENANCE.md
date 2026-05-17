# FindX 智能运维平台 — 维护文档

> 最后更新：2026-05-17
> 维护人：AI Workbench Team

---

## 一、项目概览

| 项目 | 说明 |
|------|------|
| 名称 | FindX Monitoring Core |
| 仓库 | git@github.com:17ovo17/FindXMonitoringCore.git |
| 技术栈 | Go 1.21 + React 17 + Vite 5 + MySQL + Redis + Prometheus + ES 8.12 |
| 远端 | 10.10.160.202 (findx-ubuntu) |
| API 端口 | 8080 |
| 前端端口 | 3000 (Nginx) |

---

## 二、系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                      FindX 智能运维平台                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  React SPA  │  │  Nginx :3000│  │  API :8080  │        │
│  │  (Vite 构建) │──│  反向代理    │──│  (Gin)      │        │
│  └─────────────┘  └─────────────┘  └──────┬──────┘        │
│                                            │               │
│  ┌─────────────────────────────────────────┼───────────┐   │
│  │              AI Engine (aiengine/)       │           │   │
│  │  ┌───────────┐ ┌──────────┐ ┌─────────┐│           │   │
│  │  │ Context   │ │ Prompt   │ │ Memory  ││           │   │
│  │  │ Engine    │ │ Builder  │ │ Manager ││           │   │
│  │  └───────────┘ └──────────┘ └─────────┘│           │   │
│  │  ┌───────────┐ ┌──────────┐ ┌─────────┐│           │   │
│  │  │ Autonomous│ │ Anomaly  │ │ NL2Query││           │   │
│  │  │ Investigat│ │ Detector │ │         ││           │   │
│  │  └───────────┘ └──────────┘ └─────────┘│           │   │
│  │  ┌───────────┐ ┌──────────┐ ┌─────────┐│           │   │
│  │  │ Dimension │ │ Incident │ │ Error   ││           │   │
│  │  │ Drilldown │ │ Lifecycle│ │Classifie││           │   │
│  │  └───────────┘ └──────────┘ └─────────┘│           │   │
│  └─────────────────────────────────────────┘           │   │
│                                                        │   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │   │
│  │Scheduler │ │ Notifier │ │ Workflow │ │ Sandbox  │ │   │
│  │告警评估   │ │10种通道   │ │DAG引擎   │ │三级权限   │ │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ │   │
│                                                        │   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │   │
│  │  MySQL   │ │  Redis   │ │Prometheus│ │   ES     │ │   │
│  │  85 表   │ │  缓存    │ │  指标    │ │  日志    │ │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ │   │
└─────────────────────────────────────────────────────────────┘
```

---

## 三、目录结构

```
ai-workbench/
├── api/                          # 后端 Go 代码
│   ├── main.go                   # 入口
│   ├── routes*.go                # 路由注册（按域拆分）
│   ├── config.yaml               # 配置文件
│   ├── assets/integrations/      # 40个集成模板 params.json
│   ├── internal/
│   │   ├── aiengine/             # AI 引擎核心（Context Engineering）
│   │   ├── handler/              # HTTP Handler（按域拆分）
│   │   ├── model/                # 数据模型
│   │   ├── store/                # 存储层（GORM + 内存双模）
│   │   ├── scheduler/            # 告警评估 + 抑制
│   │   ├── notifier/             # 通知发送（10种通道）
│   │   ├── evaluator/            # PromQL 评估器
│   │   ├── workflow/engine/      # 工作流 DAG 引擎 + 30 YAML
│   │   ├── sandbox/              # AI 执行沙箱
│   │   ├── eventbus/             # 事件总线
│   │   └── security/             # 命令白名单/黑名单
│   └── embed_integrations.go     # embed.FS 嵌入
├── web/                          # 前端代码
│   ├── src/react-shell/          # React SPA 主体
│   │   ├── ai-sre/              # AI 诊断/Evidence/异常/事件
│   │   ├── base-monitoring/     # 告警/仪表盘/数据源/通知/模板
│   │   ├── cmdb/               # CMDB 资产管理
│   │   ├── agents/             # Agent 管理
│   │   ├── tracing/            # 链路追踪
│   │   ├── logs/               # 日志中心
│   │   ├── overview/           # 全局概览
│   │   ├── org/                # 人员组织
│   │   ├── platform/           # 系统配置 + 沙箱
│   │   ├── probes/             # 业务拨测
│   │   ├── shared/             # 通用组件
│   │   ├── stores/             # zustand 全局状态
│   │   ├── api/                # API 封装层（18个文件）
│   │   └── navigation.js       # 导航系统
│   └── vite.config.js
├── tmp/catpaw-conf/             # 34个 Catpaw 配置模板
├── docs/                        # 文档（80+个）
└── scripts/                     # 构建脚本
```

---

## 四、本地开发

### 环境要求
- Go 1.21+
- Node.js 20+
- MySQL 8.0+
- Redis 6+
- Prometheus（可选，用于指标查询）

### 启动后端
```bash
cd api
# 设置环境变量
export MYSQL_DSN="user:pass@tcp(127.0.0.1:3306)/ai_workbench?charset=utf8mb4&parseTime=True"
export AI_WORKBENCH_API_KEY="your-llm-api-key"
export AI_WORKBENCH_BASE_URL="https://api.openai.com/v1"

go run .
# 监听 :8080
```

### 启动前端
```bash
cd web
npm install
npm run dev
# 监听 :5173，代理 /api → :8080
```

### 构建
```bash
# 后端
cd api && go build -o api-linux .

# 前端
cd web && npm run build
# 产出 web/dist/
```

---

## 五、远端部署

### 服务器信息
```
Host: 10.10.160.202
User: root
SSH: ssh findx-ubuntu
```

### 部署步骤
```bash
# 1. 交叉编译
cd api && GOOS=linux GOARCH=amd64 go build -o api-linux .

# 2. 停止旧服务
ssh findx-ubuntu "fuser -k 8080/tcp"

# 3. 上传二进制
scp api-linux findx-ubuntu:/opt/ai-workbench/api/api-linux

# 4. 上传前端
scp -r web/dist/. findx-ubuntu:/opt/ai-workbench-runtime/web/dist/

# 5. 启动服务
ssh findx-ubuntu "cd /opt/ai-workbench/api && nohup ./api-linux > /tmp/findx-api.log 2>&1 &"

# 6. 验证
ssh findx-ubuntu "curl -s http://localhost:8080/api/v1/alert-mutes"
ssh findx-ubuntu "curl -sI http://localhost:3000"
```

### 服务管理
```bash
# 查看日志
ssh findx-ubuntu "tail -100 /tmp/findx-api.log"

# 重启 API
ssh findx-ubuntu "fuser -k 8080/tcp; sleep 2; cd /opt/ai-workbench/api && nohup ./api-linux > /tmp/findx-api.log 2>&1 &"

# 重启 Nginx
ssh findx-ubuntu "systemctl restart nginx"

# 查看端口占用
ssh findx-ubuntu "ss -tlnp | grep -E '8080|3000|9090|3306|6379|9200'"
```

---

## 六、AI 引擎架构

### Context Engineering 核心理念
- **不堆积无效历史** — 动态选择最相关的上下文注入
- **结构化状态迁移** — 每步只传必要 JSON，不传原始数据
- **意图驱动** — 根据用户意图（问答/脚本/拓扑/自愈/诊断）动态构建 prompt

### 模块说明

| 模块 | 文件 | 对标 | 作用 |
|------|------|------|------|
| ContextEngine | aiengine/context_engine.go | Hermes | 动态上下文构建 |
| PromptBuilder | aiengine/prompt_builder.go | Hermes | 按意图构建 system prompt |
| MemoryManager | aiengine/memory_manager.go | Hermes | 跨会话记忆 |
| Curator | aiengine/curator.go | Hermes | 选择最相关历史 |
| ErrorClassifier | aiengine/error_classifier.go | n9e | 告警分类→路由工作流 |
| Trajectory | aiengine/trajectory.go | Hermes | 执行轨迹记录 |
| AutonomousInvestigator | aiengine/autonomous_investigator.go | Neubird | 自主调查（5轮） |
| AnomalyDetector | aiengine/anomaly_detector.go | Datadog | 无规则异常检测 |
| DimensionDrilldown | aiengine/dimension_drilldown.go | Honeycomb | 维度下钻 |
| IncidentLifecycle | aiengine/incident_lifecycle.go | Rootly | 8阶段事件自动化 |
| NL2Query | aiengine/nl2query.go | Splunk | 中文→PromQL/LogQL |

### AI 沙箱权限

| 模式 | Level 0 (只读) | Level 1 (写) | Level 2 (危险) |
|------|---------------|-------------|---------------|
| 默认权限 | ✅ 直接执行 | ❌ 拒绝 | ❌ 拒绝 |
| 自动审查 | ✅ 直接执行 | ✅ 执行+审计 | ❌ 拒绝 |
| 完全访问 | ✅ 直接执行 | ✅ 执行+审计 | ⏳ 举手确认(30s) |

---

## 七、告警引擎

### 架构（对标 n9e AlertRuleWorker）
- Per-rule worker：每条规则独立 goroutine
- 状态机：pending → firing → resolved
- ForDuration：满足持续时间后才升级为 firing
- 通知分发：10 种通道（webhook/email/dingtalk/wecom/feishu/telegram/lark/feishucard/callback/mattermost）
- 告警抑制：父告警触发时自动抑制子告警

### 配置
```yaml
# config.yaml
scheduler:
  enabled: true
  eval_interval: 15s  # 评估间隔
```

---

## 八、集成模板

### 当前覆盖（40个）
P0（15个）：MySQL/Redis/Nginx/Linux/Docker/PostgreSQL/ES/Kafka/MongoDB/K8s/HTTP/Ping/Net/Prometheus/Java
P1（25个）：ClickHouse/Consul/HAProxy/Jenkins/RabbitMQ/SNMP/SQLServer/SpringBoot/Tomcat/VictoriaMetrics/ZooKeeper/MinIO/Gitlab/Canal/Ceph/Doris/IPMI/Logstash/Oracle/PHP/Procstat/Systemd/TDEngine/Whois/Windows

### 添加新模板
1. 创建 `api/assets/integrations/<name>/params.json`
2. 定义 params 数组和 toml_template
3. 重新编译后端（embed.FS 自动加载）

---

## 九、常见问题

### API 启动后返回 404
端口被旧进程占用。执行 `fuser -k 8080/tcp` 后重启。

### 前端构建报错 Unexpected
子 agent 清理 PENDING 时可能破坏 JSX 语法。检查报错行的 if/else 结构。

### AI 对话无响应
检查环境变量 `AI_WORKBENCH_API_KEY` 和 `AI_WORKBENCH_BASE_URL` 是否配置。

### Prometheus 查询超时
检查 config.yaml 中 `prometheus.url` 是否可达。

---

## 十、Git 提交规范

```
feat(模块): 一句话描述    # 新功能
fix(模块): 一句话描述     # 修复
refactor(模块): 描述      # 重构
docs(模块): 描述          # 文档
```

每完成一个 Part push 一次。当前 main 分支直接开发。
