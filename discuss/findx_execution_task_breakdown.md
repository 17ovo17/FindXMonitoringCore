# FindX 执行准备与门禁任务包

> 生成时间：2026-05-05 00:24（UTC+8）

## 1. 任务定位

本文只做 FindX 第一批执行切片的准备、门禁和子代理派发契约，不进入业务代码实现。

- 产品主语只保留 FindX、FindX Core、findx-agents、AI SRE。
- 文档内不写具体历史上游产品名、旧组织对象名、旧排班对象名；需要描述时统一使用“历史上游命名残留”“旧组织命名残留”“旧排班对象命名残留”。
- 健康检查归 AI SRE；平台治理只保留平台运行自检，不承接主机、服务、应用健康检查语义。
- 当前工作区已经存在脏改动，第一批子代理必须先通过 P0-Gate，再按切片串行或写集互斥执行。
- 当前已有后端权限矩阵和前端导航改动已完成 P0-Gate 闭环验收；后续 P0/P1/P3 功能切片必须以已提交基线为准，不得混入未纳入提交的脏文件。
- 子代理不得写入真实密钥、认证票据、Cookie、完整连接串、会话 ID；示例统一使用 `<TOKEN>`、`<API_KEY>`、`<DB_DSN>`、`<LOGIN_USER>`。

## 2. 当前工作区脏改动审计

当前工作区可分成 6 类脏改动。以下审计发现是后续执行的阻断输入；P0-Gate 已在 2026-05-05（UTC+8）完成闭环，未纳入提交的脏文件仍不得混入后续业务切片。

| 类别 | 文件 | 审计发现 | 阻断风险 | P0-Gate 处理要求 |
|---|---|---|---|---|
| 后端权限矩阵 | `api/main.go`、`api/internal/handler/monitoring_permissions.go`、`api/internal/handler/monitoring_permissions_test.go`、`api/internal/model/monitoring_permission.go` | 已经进入权限中间件、权限矩阵、权限查询接口和测试联动改动 | 这是 API_CONTRACT_CHANGE，必须单独验证状态码、响应结构、401/403、角色映射、旧入口兼容和前端消费口径 | 由 P0-Gate 串行收口，未形成验证证据前，后续业务切片不得复用为“已稳定能力” |
| 告警校验归属 | `api/internal/handler/monitoring_alert_validation.go` | 文件位于后端监控告警校验链路，但当前脏改动与权限矩阵一起出现 | 归属不清会造成权限矩阵、告警校验、通知策略三类切片互相踩写 | P0-Gate 必须明确它归属于“告警校验/规则校验”还是“权限门禁”，不能被业务空间、导航或通知切片顺手修改 |
| 前端导航 | `web/src/App.vue`、`web/src/router/index.js`、`web/src/router/nav.js`、各工作台壳页面 | 已经进入全局导航、section 路由、工作台拆分和入口页壳改造 | 当前存在命名风险和健康检查归属风险；P0-Gate 之前不允许判定通过 | 必须先修正历史上游命名残留、旧组织命名残留、旧排班对象命名残留；健康检查归 AI SRE，平台治理只保留平台运行自检 |
| 远程配置控制面 | `web/src/views/PluginConfig.vue`、相关接入页 | 涉及远程配置、命令、凭证、审计和生成配置内容 | 安全风险高，不应混入导航切片或普通资产页壳验收 | 作为 findx-agents 控制面或远程配置独立验收；必须验证凭证脱敏、命令边界、审计记录和权限拒绝 |
| 登录与改密 | `web/src/views/Login.vue`、`web/src/views/ChangePassword.vue` | 属于格式化、认证头和登录态处理风险 | 混入导航切片会扩大认证面风险，影响回归定位 | 作为认证入口风险单独评估；P0-Gate 只记录现状，不允许导航切片顺手改认证逻辑 |
| 协作边界文档 | `discuss/claude_codex_boundary.md` | 不属于本轮两份规划文档，也不是业务实现切片 | 混入第一批代码拆包会污染验收范围 | 单独由文档角色处理，不进入 P0/P1/P3 功能切片验收 |

P0-Gate 的特殊结论：已通过。后端权限矩阵、前端导航命名边界和健康检查归属已在独立提交中完成收口；健康检查归 AI SRE，平台治理只保留平台运行自检。

### 2.1 P0-Gate 闭环记录

记录日期：2026-05-05（UTC+8）。

| 项目 | 闭环证据 |
|---|---|
| 门禁状态 | P0-Gate 已闭环，允许进入后续 P0-T1 前置准入审计。 |
| 文档 commit | `f36927f docs: define FindX execution planning` |
| 代码 commit | `b3a8893 feat: add FindX monitor permission gate and navigation` |
| push 状态 | 已推送到 `origin/main`，`HEAD == origin/main == b3a8893`。 |
| QA 结论 | 92/100，PASS，无 P0/P1 阻断。 |
| Windows 验证 | Go 定向测试 PASS；`web npm run build` PASS；`git diff --check` PASS。 |
| WSL 验证 | `api go test ./...` PASS；`api go build -o api-linux .` PASS；`web npm run build` PASS。 |

未纳入提交的脏文件清单：

- `web/src/views/ChangePassword.vue`
- `web/src/views/Login.vue`
- `web/src/views/PluginConfig.vue`
- `web/src/views/monitoring/MonitorWorkbench.vue`

上述 4 个文件不属于 P0-Gate 已提交闭环范围，不得混入下一切片。进入 P0-T1 前，主代理必须先只读审计并隔离这些脏文件；不得直接 revert 用户改动；不得把这些文件混入 P0-T1 commit。

## 3. 第一批切片总览

第一批按“先门禁、再 FindX Core 基础、再通知闭环、再 findx-agents 控制面”的顺序执行。P0-Gate 是所有业务切片的进入条件。

| 顺序 | 切片编号 | 切片名称 | 主要依赖 | 首要产物 |
|---|---|---|---|---|
| 1 | P0-Gate | 现有权限矩阵与导航变更验收 | 当前脏改动审计 | 权限矩阵验收、导航命名收口、健康检查归属收口 |
| 2 | P0-T1 | 业务空间 | P0-Gate | 业务空间 API、数据模型、成员与授权边界、前端入口 |
| 3 | P0-T2 | 资源组与主机资产 | P0-Gate、P0-T1 | 资源组、主机资产、标签、归属绑定 |
| 4 | P0-T3 | 监控仪表盘基础能力 | P0-Gate、P0-T1、P0-T2 | 仪表盘列表、详情、克隆、分享、模板基础能力 |
| 5 | P1-T1 | 通知渠道 | P0-Gate、P0-T1 | 通知渠道 CRUD、测试发送、凭证脱敏 |
| 6 | P1-T2 | 消息模板 | P0-Gate、P1-T1 | 模板变量、预览、版本兼容 |
| 7 | P1-T3 | 通知策略 | P0-Gate、P0-T1、P1-T1、P1-T2 | 策略匹配、渠道绑定、静默与订阅挂接 |
| 8 | P3-T1 | findx-agents 注册心跳 | P0-Gate、P0-T1、P0-T2 | 注册、心跳、Agent 身份、审计 |
| 9 | P3-T2 | findx-agents 能力目录 | P3-T1 | 能力上报、能力查询、兼容性字段 |
| 10 | P3-T3 | findx-agents 插件目录 | P3-T1、P3-T2 | 插件元数据、版本、适配范围 |
| 11 | P3-T4 | findx-agents 采集配置下发 | P3-T1、P3-T2、P3-T3 | 配置生成、下发、回滚、审计 |

## 4. 第一批切片执行契约表

### 4.1 P0-Gate：现有权限矩阵与导航变更验收

| 字段 | 内容 |
|---|---|
| 任务目标 | 冻结现有后端权限矩阵、前端导航、命名边界和健康检查归属；本门禁已在 2026-05-05（UTC+8）闭环通过，后续切片只能基于已提交基线继续推进。 |
| 所属阶段 | P0 门禁阶段；所有 P0/P1/P3 功能切片之前执行。 |
| 允许写入路径 | `api/main.go`、`api/internal/handler/monitoring_permissions.go`、`api/internal/handler/monitoring_permissions_test.go`、`api/internal/model/monitoring_permission.go`、`web/src/App.vue`、`web/src/router/index.js`、`web/src/router/nav.js`、必要工作台壳页面。 |
| 禁止写入路径 | `web/src/views/PluginConfig.vue`、`web/src/views/Login.vue`、`web/src/views/ChangePassword.vue`、`discuss/claude_codex_boundary.md`、与权限矩阵和导航门禁无关的业务实现文件、运行产物、Git 元数据、配置密钥。 |
| API_CONTRACT_CHANGE | YES：后端权限矩阵、权限查询/校验接口和前端权限消费口径已经变化，必须单独验证响应结构、状态码、角色映射、401/403 和旧入口兼容。 |
| DATA_CHANGE | NO：本切片只验收现有门禁和导航改动，不新增表、字段、迁移或状态值；如发现权限状态值落库语义变化，立即升级为 DATA_CHANGE 并停止扩展。 |
| 前端入口 | `/assets`、`/query`、`/dashboards`、`/alerts`、`/notifications`、`/aiops`、`/org`、`/platform`；`/platform?section=health` 仅允许表达平台运行自检，主机/服务健康检查必须进入 AI SRE 归属。 |
| 验证命令 | 同步到 WSL 后执行：`cd /opt/ai-workbench/api && go build -o api-linux .`；`cd /opt/ai-workbench/api && go test ./...`；`cd /opt/ai-workbench/web && npm run build`；用 `<TOKEN>` 调用权限查询/校验接口并覆盖 200、401、403。 |
| QA 验收标准 | qa-tester 必须给出 PASS/FAIL/BLOCKED/RISK；命名风险、健康检查归属风险、权限矩阵 API_CONTRACT_CHANGE 未验证任一存在即 FAIL；浏览器必须覆盖导航跳转、旧入口重定向、权限拒绝态和平台运行自检入口。 |
| 是否允许 Git commit | NO：子代理不提交；只输出变更说明、验证证据、风险结论，主代理或 Claude 另行决定 Git 门禁。 |

### 4.2 P0-T1：业务空间

| 字段 | 内容 |
|---|---|
| 任务目标 | 建立 FindX Core 的第一层业务边界，提供业务空间 CRUD、成员边界、授权骨架和前端入口。 |
| 所属阶段 | P0 Core 基础阶段；依赖 P0-Gate。 |
| 允许写入路径 | `api/internal/handler/`、`api/internal/model/`、`api/internal/store/`、`api/main.go` 中与业务空间直接相关的文件；`web/src/views/AssetsWorkbench.vue`、`web/src/api/`、必要组件文件。 |
| 禁止写入路径 | findx-agents 控制面文件、通知切片文件、登录与改密文件、远程配置文件、长期规划文档、运行产物、Git 元数据、配置密钥。 |
| API_CONTRACT_CHANGE | YES：新增或调整业务空间 CRUD、成员、授权相关接口时，必须列出路径、方法、请求字段、响应字段、错误码、权限和兼容性。 |
| DATA_CHANGE | YES：涉及业务空间实体、成员关系、唯一性、默认值或旧数据兼容；必须说明迁移、幂等、回滚和内存 fallback 行为。 |
| 前端入口 | `/assets?section=business`；页面文案必须使用业务空间语义，不使用旧组织命名残留作为产品命名。 |
| 验证命令 | `cd /opt/ai-workbench/api && go build -o api-linux .`；`cd /opt/ai-workbench/api && go test ./...`；`cd /opt/ai-workbench/web && npm run build`；用 `<TOKEN>` 覆盖创建、查询、更新、删除、重复名称、无权限访问。 |
| QA 验收标准 | 正常路径：创建并查询业务空间；异常路径：非法名称或重复名称返回用户可理解错误；权限路径：无权限用户被拒绝且不泄露内部信息；浏览器覆盖列表、空态、表单校验和权限拒绝态。 |
| 是否允许 Git commit | NO：子代理不提交；完成后交主代理验收。 |

### 4.3 P0-T2：资源组与主机资产

| 字段 | 内容 |
|---|---|
| 任务目标 | 建立资源组、主机资产、标签和业务空间归属绑定，形成 FindX Core 的资源边界。 |
| 所属阶段 | P0 Core 基础阶段；依赖 P0-Gate 和 P0-T1。 |
| 允许写入路径 | `api/internal/handler/`、`api/internal/model/`、`api/internal/store/`、`api/main.go` 中与资源组、主机资产、标签归属直接相关的文件；`web/src/views/AssetsWorkbench.vue`、`web/src/api/`、必要组件文件。 |
| 禁止写入路径 | 通知切片文件、findx-agents 远程配置文件、登录与改密文件、平台自检文件、长期规划文档、运行产物、Git 元数据、配置密钥。 |
| API_CONTRACT_CHANGE | YES：新增或调整资源组、主机资产、标签、归属绑定接口时，必须列出契约、权限、分页过滤、错误码和旧入口兼容。 |
| DATA_CHANGE | YES：涉及资源组、主机资产、标签、归属关系和唯一约束；必须说明旧数据兼容、迁移需求、幂等导入和回滚方式。 |
| 前端入口 | `/assets?section=hosts`，必要时联动 `/assets?section=business`；不得把远程配置控制面混入本切片。 |
| 验证命令 | `cd /opt/ai-workbench/api && go build -o api-linux .`；`cd /opt/ai-workbench/api && go test ./...`；`cd /opt/ai-workbench/web && npm run build`；用 `<TOKEN>` 覆盖资源组创建、主机绑定、标签过滤、无业务空间归属、无权限访问。 |
| QA 验收标准 | 正常路径：资源组绑定主机并可按业务空间过滤；异常路径：不存在的主机或资源组返回 4xx；边界路径：空列表、分页末页、重复绑定幂等；权限路径：越权资源不可见。 |
| 是否允许 Git commit | NO：子代理不提交；完成后交主代理验收。 |

### 4.4 P0-T3：监控仪表盘基础能力

| 字段 | 内容 |
|---|---|
| 任务目标 | 提供监控仪表盘列表、详情、创建、更新、删除、克隆、分享和模板基础能力。 |
| 所属阶段 | P0 Core 展示阶段；依赖 P0-Gate、P0-T1、P0-T2。 |
| 允许写入路径 | `api/internal/handler/`、`api/internal/model/`、`api/internal/store/`、`api/main.go` 中与仪表盘直接相关的文件；`web/src/views/DashboardsWorkbench.vue`、`web/src/api/`、必要组件文件。 |
| 禁止写入路径 | 通知切片文件、findx-agents 控制面文件、登录与改密文件、业务空间和资源组无关重构、运行产物、Git 元数据、配置密钥。 |
| API_CONTRACT_CHANGE | YES：新增或调整仪表盘、图表、分享、模板接口时，必须列出字段、权限、错误码、分享可见性和兼容策略。 |
| DATA_CHANGE | YES：涉及仪表盘元数据、图表布局、模板、分享记录；必须说明 JSON 字段校验、旧数据兼容、复制幂等和回滚方式。 |
| 前端入口 | `/dashboards?section=list`、`/dashboards?section=templates`、`/dashboards?section=shares`。 |
| 验证命令 | `cd /opt/ai-workbench/api && go build -o api-linux .`；`cd /opt/ai-workbench/api && go test ./...`；`cd /opt/ai-workbench/web && npm run build`；用 `<TOKEN>` 覆盖新建、编辑、克隆、分享、模板查询、非法布局。 |
| QA 验收标准 | 正常路径：创建仪表盘并可打开详情；异常路径：非法图表配置返回 4xx；边界路径：空面板、最大面板数、分享关闭后不可访问；权限路径：非授权业务空间不可读写。 |
| 是否允许 Git commit | NO：子代理不提交；完成后交主代理验收。 |

### 4.5 P1-T1：通知渠道

| 字段 | 内容 |
|---|---|
| 任务目标 | 建立通知渠道 CRUD、测试发送、凭证脱敏和渠道可用性校验。 |
| 所属阶段 | P1 通知闭环阶段；依赖 P0-Gate 和 P0-T1。 |
| 允许写入路径 | `api/internal/handler/`、`api/internal/model/`、`api/internal/store/`、`api/main.go` 中与通知渠道直接相关的文件；`web/src/views/NotificationsWorkbench.vue`、`web/src/api/`、必要组件文件。 |
| 禁止写入路径 | 消息模板之外的通知策略文件、findx-agents 控制面文件、登录与改密文件、平台自检文件、运行产物、Git 元数据、配置密钥。 |
| API_CONTRACT_CHANGE | YES：新增或调整通知渠道接口、测试发送接口和凭证字段时，必须列出权限、错误码、超时、外部依赖失败和脱敏响应。 |
| DATA_CHANGE | YES：涉及通知渠道、凭证占位、启停状态、测试记录；必须说明敏感字段加密或脱敏、迁移、回滚和审计。 |
| 前端入口 | `/notifications?section=channels`。 |
| 验证命令 | `cd /opt/ai-workbench/api && go build -o api-linux .`；`cd /opt/ai-workbench/api && go test ./...`；`cd /opt/ai-workbench/web && npm run build`；用 `<TOKEN>` 和 `<API_KEY>` 占位覆盖创建渠道、测试发送失败、凭证回显脱敏、无权限访问。 |
| QA 验收标准 | 正常路径：创建渠道并保存成功；异常路径：无效地址或外部依赖失败返回 4xx/503；安全路径：接口和页面不回显真实凭证；权限路径：非授权用户不可测试发送。 |
| 是否允许 Git commit | NO：子代理不提交；完成后交主代理验收。 |

### 4.6 P1-T2：消息模板

| 字段 | 内容 |
|---|---|
| 任务目标 | 建立消息模板 CRUD、变量定义、预览渲染和版本兼容规则。 |
| 所属阶段 | P1 通知闭环阶段；依赖 P0-Gate 和 P1-T1。 |
| 允许写入路径 | `api/internal/handler/`、`api/internal/model/`、`api/internal/store/`、`api/main.go` 中与消息模板直接相关的文件；`web/src/views/NotificationsWorkbench.vue`、`web/src/api/`、必要组件文件。 |
| 禁止写入路径 | 通知渠道凭证实现、findx-agents 控制面文件、登录与改密文件、平台自检文件、运行产物、Git 元数据、配置密钥。 |
| API_CONTRACT_CHANGE | YES：新增或调整模板 CRUD、预览和变量接口时，必须列出模板语法、变量白名单、错误码和兼容策略。 |
| DATA_CHANGE | YES：涉及模板内容、变量声明、版本号、默认模板；必须说明旧模板兼容、回滚和内容长度边界。 |
| 前端入口 | `/notifications?section=templates`。 |
| 验证命令 | `cd /opt/ai-workbench/api && go build -o api-linux .`；`cd /opt/ai-workbench/api && go test ./...`；`cd /opt/ai-workbench/web && npm run build`；用 `<TOKEN>` 覆盖模板创建、变量预览、非法变量、超长内容、无权限访问。 |
| QA 验收标准 | 正常路径：模板可保存并预览；异常路径：未知变量或语法错误返回用户可理解错误；边界路径：空模板、超长模板、版本回退；安全路径：预览不执行脚本或泄露敏感上下文。 |
| 是否允许 Git commit | NO：子代理不提交；完成后交主代理验收。 |

### 4.7 P1-T3：通知策略

| 字段 | 内容 |
|---|---|
| 任务目标 | 建立通知策略匹配、渠道绑定、模板绑定、静默与订阅挂接，形成告警到通知的基础闭环；P1 只允许形成通知和事件闭环，最多创建 AI SRE 诊断入口占位，不得触发 Remediation 或 findx-agents 自愈执行。 |
| 所属阶段 | P1 通知闭环阶段；依赖 P0-Gate、P0-T1、P1-T1、P1-T2。 |
| 允许写入路径 | `api/internal/handler/`、`api/internal/model/`、`api/internal/store/`、`api/main.go` 中与通知策略直接相关的文件；`web/src/views/NotificationsWorkbench.vue`、`web/src/views/AlertingWorkbench.vue`、`web/src/api/`、必要组件文件。 |
| 禁止写入路径 | 通知渠道凭证底层实现、findx-agents 控制面文件、登录与改密文件、平台自检文件、运行产物、Git 元数据、配置密钥。 |
| API_CONTRACT_CHANGE | YES：新增或调整策略 CRUD、匹配测试、渠道绑定、模板绑定、静默订阅接口时，必须列出权限、错误码、匹配优先级和兼容策略。 |
| DATA_CHANGE | YES：涉及策略规则、绑定关系、优先级、静默和订阅引用；必须说明唯一约束、启停状态、幂等更新和回滚。 |
| 前端入口 | `/notifications?section=rules`，必要时联动 `/alerts?section=mutes`、`/alerts?section=subscribes`。 |
| 验证命令 | `cd /opt/ai-workbench/api && go build -o api-linux .`；`cd /opt/ai-workbench/api && go test ./...`；`cd /opt/ai-workbench/web && npm run build`；用 `<TOKEN>` 覆盖策略创建、匹配测试、禁用策略、缺失渠道、权限拒绝。 |
| QA 验收标准 | 正常路径：策略能匹配事件并选择渠道模板；异常路径：缺失渠道或模板返回 4xx；边界路径：优先级冲突、禁用策略、空匹配；权限路径：跨业务空间策略不可见不可用。 |
| 是否允许 Git commit | NO：子代理不提交；完成后交主代理验收。 |

### 4.8.0 P3-Min-Security Gate：findx-agents 最小安全门禁

P3-T3/P3-T4 进入前必须具备 agent 身份、权限动作、审计日志、凭证引用脱敏、任务幂等、失败回执和权限拒绝验证；否则 P3-T4 只能交付配置预览/diff，不允许真实下发。

### 4.8 P3-T1：findx-agents 注册心跳

| 字段 | 内容 |
|---|---|
| 任务目标 | 建立 findx-agents 注册、心跳、Agent 身份、状态更新和审计基础。 |
| 所属阶段 | P3 findx-agents 控制面阶段；依赖 P0-Gate、P0-T1、P0-T2。 |
| 允许写入路径 | `api/internal/handler/`、`api/internal/model/`、`api/internal/store/`、`api/main.go` 中与 findx-agents 注册心跳直接相关的文件；`web/src/views/AssetsWorkbench.vue`、`web/src/api/`、必要组件文件。 |
| 禁止写入路径 | 通知切片文件、消息模板文件、登录与改密文件、平台自检文件、远程配置生成逻辑、运行产物、Git 元数据、配置密钥。 |
| API_CONTRACT_CHANGE | YES：注册和心跳接口必须明确认证头、Agent 标识、幂等行为、错误码、审计字段和兼容策略。 |
| DATA_CHANGE | YES：涉及 Agent 资产、心跳时间、状态、业务空间归属、审计记录；必须说明离线判定、旧数据兼容、幂等注册和回滚方式。 |
| 前端入口 | `/assets?section=agents`；只展示注册与存活基础状态，不混入远程配置或命令能力。 |
| 验证命令 | `cd /opt/ai-workbench/api && go build -o api-linux .`；`cd /opt/ai-workbench/api && go test ./...`；`cd /opt/ai-workbench/web && npm run build`；用 `<TOKEN>` 覆盖注册、重复注册、心跳、无效 Agent 标识、未授权查询。 |
| QA 验收标准 | 正常路径：Agent 注册后心跳更新时间；异常路径：缺少认证或非法标识被拒绝；边界路径：重复注册幂等、离线状态计算；安全路径：令牌不写日志不回显。 |
| 是否允许 Git commit | NO：子代理不提交；完成后交主代理验收。 |

### 4.9 P3-T2：findx-agents 能力目录

| 字段 | 内容 |
|---|---|
| 任务目标 | 建立 findx-agents 能力上报、能力查询、版本兼容和能力启停语义。 |
| 所属阶段 | P3 findx-agents 控制面阶段；依赖 P3-T1。 |
| 允许写入路径 | `api/internal/handler/`、`api/internal/model/`、`api/internal/store/`、`api/main.go` 中与能力目录直接相关的文件；`web/src/views/AssetsWorkbench.vue`、`web/src/api/`、必要组件文件。 |
| 禁止写入路径 | 通知切片文件、登录与改密文件、远程配置下发文件、平台自检文件、运行产物、Git 元数据、配置密钥。 |
| API_CONTRACT_CHANGE | YES：能力上报和查询接口必须明确能力类型、版本、状态、兼容字段、错误码和权限。 |
| DATA_CHANGE | YES：涉及能力目录、Agent 能力绑定、版本状态；必须说明重复上报幂等、能力废弃、兼容查询和回滚方式。 |
| 前端入口 | `/assets?section=agents`；能力目录作为 Agent 详情或控制面子视图，不扩展全局导航。 |
| 验证命令 | `cd /opt/ai-workbench/api && go build -o api-linux .`；`cd /opt/ai-workbench/api && go test ./...`；`cd /opt/ai-workbench/web && npm run build`；用 `<TOKEN>` 覆盖能力上报、重复上报、未知能力、禁用能力、无权限查询。 |
| QA 验收标准 | 正常路径：能力上报后可查询；异常路径：未知能力类型返回 4xx；边界路径：版本为空、重复上报、禁用能力；权限路径：跨业务空间不可见。 |
| 是否允许 Git commit | NO：子代理不提交；完成后交主代理验收。 |

### 4.10 P3-T3：findx-agents 插件目录

| 字段 | 内容 |
|---|---|
| 任务目标 | 建立 findx-agents 插件目录、插件元数据、版本、适配范围和安全边界。 |
| 所属阶段 | P3 findx-agents 控制面阶段；依赖 P3-T1、P3-T2。 |
| 允许写入路径 | `api/internal/handler/`、`api/internal/model/`、`api/internal/store/`、`api/main.go` 中与插件目录直接相关的文件；`web/src/views/AssetsWorkbench.vue`、`web/src/views/PluginConfig.vue`、`web/src/api/`、必要组件文件。 |
| 禁止写入路径 | 通知切片文件、登录与改密文件、平台自检文件、与插件目录无关的远程命令执行链、运行产物、Git 元数据、配置密钥。 |
| API_CONTRACT_CHANGE | YES：插件目录接口必须明确插件标识、版本、能力依赖、适配范围、启停状态、权限和错误码。 |
| DATA_CHANGE | YES：涉及插件目录、版本元数据、能力绑定和适配范围；必须说明去重、版本兼容、废弃策略和回滚。 |
| 前端入口 | `/assets?section=agents&tab=plugins`；`PluginConfig.vue` 只允许在本切片或远程配置独立验收中处理，不得混入导航切片。 |
| 验证命令 | `cd /opt/ai-workbench/api && go build -o api-linux .`；`cd /opt/ai-workbench/api && go test ./...`；`cd /opt/ai-workbench/web && npm run build`；用 `<TOKEN>` 覆盖插件查询、版本过滤、禁用插件、未知能力依赖、无权限访问。 |
| QA 验收标准 | 正常路径：插件目录可查询并展示版本；异常路径：未知插件或不兼容能力返回用户可理解错误；边界路径：重复版本、禁用插件、空目录；安全路径：不展示真实凭证，不提供越权命令入口。 |
| 是否允许 Git commit | NO：子代理不提交；完成后交主代理验收。 |

### 4.11 P3-T4：findx-agents 采集配置下发

| 字段 | 内容 |
|---|---|
| 任务目标 | 建立采集配置生成、预览、下发、回滚、状态回执和审计闭环。 |
| 所属阶段 | P3 findx-agents 控制面阶段；依赖 P3-T1、P3-T2、P3-T3。 |
| 允许写入路径 | `api/internal/handler/`、`api/internal/model/`、`api/internal/store/`、`api/main.go` 中与采集配置下发直接相关的文件；`web/src/views/PluginConfig.vue`、`web/src/views/AssetsWorkbench.vue`、`web/src/api/`、必要组件文件。 |
| 禁止写入路径 | 通知切片文件、登录与改密文件、平台自检文件、无审计的远程命令入口、运行产物、Git 元数据、配置密钥。 |
| API_CONTRACT_CHANGE | YES：配置预览、下发、回滚和回执接口必须明确权限、审计字段、超时、错误码、幂等键和脱敏响应。 |
| DATA_CHANGE | YES：涉及配置版本、下发任务、回执状态、审计记录和回滚指针；必须说明旧配置保护、重复提交幂等、失败回滚和清理策略。 |
| 前端入口 | `/assets?section=agents&tab=plugins`；远程配置、命令、凭证、审计作为独立验收面，不得混入导航切片。 |
| 验证命令 | `cd /opt/ai-workbench/api && go build -o api-linux .`；`cd /opt/ai-workbench/api && go test ./...`；`cd /opt/ai-workbench/web && npm run build`；用 `<TOKEN>` 和 `<API_KEY>` 占位覆盖配置预览、下发、重复提交、回滚、失败回执、无权限访问。 |
| QA 验收标准 | 正常路径：配置预览后可下发并收到状态；异常路径：无效配置或 Agent 离线返回 4xx/503；边界路径：重复下发幂等、回滚到上一版本、超时回执；安全路径：凭证脱敏、审计完整、禁止无审计命令入口；权限路径：无 execute/approve 权限不得下发；失败路径：Agent 离线、版本不兼容、插件缺失时不得静默成功；审计路径：下发任务必须有 request_id/run_id、配置版本、回滚指针和审计记录。 |
| 是否允许 Git commit | NO：子代理不提交；完成后交主代理验收。 |

## 5. 子代理派发边界

### 5.1 go-backend

- 写集：仅限具体切片授权的 `api/` 文件。
- 负责内容：API 契约、handler/service/store/model、权限校验、错误模型、测试。
- 输出要求：标明 API_CONTRACT_CHANGE、DATA_CHANGE、正常路径、异常路径、边界或权限路径、敏感信息脱敏、回滚方式。
- 禁止事项：不得改前端、文档、运行配置、Git 元数据，不得写真实密钥。

### 5.2 vue-frontend

- 写集：仅限具体切片授权的 `web/src/` 文件。
- 负责内容：页面入口、导航消费、API 封装、空态、错误态、权限受限态、浏览器验证。
- 输出要求：路由与后端契约对齐，展示错误用户可理解，不展示堆栈、SQL、内部路径或完整连接信息。
- 禁止事项：不得改后端，不得扩展后端权限语义，不得把 `PluginConfig.vue`、登录与改密混入导航切片。

### 5.3 qa-tester

- 写集：只读。
- 负责内容：WSL 构建、API 验证、浏览器回归、敏感信息扫描、QA 评分门禁。
- 输出要求：结果只能使用 PASS、FAIL、BLOCKED、NOT_RUN、RISK；证据不足不得 PASS。
- 禁止事项：不得用开发自测代替 QA，不得在阻断项未关闭时给通过。

### 5.4 doc-writer

- 写集：仅限主线明确授权的文档。
- 负责内容：执行准备、测试计划、变更记录、术语收口。
- 输出要求：文档可执行、可验证、可回退；命名不引入历史上游命名残留。
- 禁止事项：不得改业务代码、运行配置、Git 元数据或敏感配置。

### 5.5 ops-diagnostician

- 写集：诊断设计、Runbook、指标映射、知识库、工作流说明类文档。
- 负责内容：AI SRE 健康检查归属、平台运行自检边界、证据链映射和诊断路径设计。
- 输出要求：明确平台自检与主机/服务健康检查边界。
- 禁止事项：不得把平台治理扩成主机健康检查，不得直接写业务代码。

## 6. 统一门禁

每个切片必须同时满足以下门禁，少一项都不能算收口。

| 门禁 | 要求 |
|---|---|
| WSL 验证 | 先按任务范围同步到 `/opt/ai-workbench`，再以 WSL 结果作为准结果。 |
| 后端 build/test | 至少执行 `cd /opt/ai-workbench/api && go build -o api-linux .`，并按切片范围补充 `go test ./...` 或定向测试。 |
| 前端 build | 涉及前端时至少执行 `cd /opt/ai-workbench/web && npm run build`。 |
| 浏览器验证 | 涉及 UI 时用真实浏览器检查导航、页面切换、空态、错误态、权限拒绝态和基础交互。 |
| API 验证 | 用 curl 或测试脚本验证状态码、响应结构、认证、权限、正常路径、异常路径和边界路径。 |
| 敏感信息扫描 | 检查代码、日志、页面和接口响应是否出现真实密钥、认证票据、Cookie、完整连接串、会话 ID；允许 `<TOKEN>`、`<API_KEY>`、`<DB_DSN>`、`<LOGIN_USER>` 等占位符，允许既有文件名如 `ChangePassword.vue`；禁止真实密钥、裸 Bearer 值、真实浏览器会话凭据、真实完整 DSN、SSH 私钥和 Go MySQL TCP DSN 片段。 |
| QA 评分门禁 | qa-tester 给出证据完整的结果；P0/P1 阻断、敏感信息风险、API_CONTRACT_CHANGE/DATA_CHANGE 未验证均不得通过。 |
| 命名自检 | 文档、导航、页面和接口响应不得把历史上游命名残留、旧组织命名残留、旧排班对象命名残留作为产品命名。 |

## 7. 近期不做

以下能力保留在长期规划语义中，不进入第一批执行：

第一批 P3-T1~P3-T4 不实现远程调试、自升级、结构化命令执行、告警自愈、自动修复执行链；若发现需要执行动作，必须拆到后续独立任务并走 P2/P3/P5 安全门禁。

- Electron
- RUM 完整平台
- APM 完整平台
- 日志平台完整实现
- 自动修复执行链
- Status Page
- 排班轮转深化
- 无审计远程命令

处理原则：

- 只保留长期语义，不拆入当前执行准备。
- 不把这些能力写成当前已经完成。
- 不因为长期能力扩大第一批切片范围。

## 8. 假设与收口建议

- 当前工作区里已经可见的脏改动，就是本轮执行准备的完整基线；如果后续新增脏改动，需要重新做一次分类审计。
- 第一批切片默认不新增第三方依赖；确需新增时必须先走依赖准入说明。
- 第一批切片默认不提交 Git；完成后只输出变更说明、验证证据和风险结论。
- `/opt/ai-workbench` 是 WSL 侧有效验证位点。
- P0-Gate 先于所有功能切片；如果门禁和功能目标冲突，先保门禁。

建议后续节奏：

1. P0-Gate 已闭环；进入 P0-T1 前，先只读审计并隔离 `web/src/views/ChangePassword.vue`、`web/src/views/Login.vue`、`web/src/views/PluginConfig.vue`、`web/src/views/monitoring/MonitorWorkbench.vue`，不得直接 revert 用户改动，不得混入 P0-T1 commit。
2. 再按 P0-T1、P0-T2、P0-T3、P1-T1、P1-T2、P1-T3、P3-T1、P3-T2、P3-T3、P3-T4 逐个收口。
3. 每个切片完成 WSL、build/test、浏览器、API、敏感信息、QA 门禁后，再进入下一个切片。
4. 任何需要扩大范围的请求，都先回到执行契约表，不边做边扩。
