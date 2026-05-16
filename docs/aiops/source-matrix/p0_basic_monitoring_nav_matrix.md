# FindX P0 导航同源矩阵

生成时间：2026-05-07 13:42（UTC+8）
状态：P0-N9E-NAV-REAL 编码前门禁，不代表前端代码已经完成

## 1. 结论

FindX 导航必须使用自有壳层、自有登录、自有权限和自有视觉风格，但菜单层级、路由语义和页面归属必须按成熟源码同源实现。不得 iframe 参考站，不得嵌入参考站 SSO，不得把成熟页面拆成低价值孤岛入口。

当前 FindX 导航与成熟源码仍有差异，进入前端实现前必须按本矩阵修正：

| 项目 | 结论 |
| --- | --- |
| 主导航事实源 | `D:\项目迁移文件\平台源码\fe-main\src\components\SideMenu\menu.tsx` |
| 路由事实源 | `D:\项目迁移文件\平台源码\fe-main\src\routers\index.tsx` |
| 当前 FindX 导航文件 | `D:\ai-workbench\web\src\router\nav.js` |
| 当前 FindX 路由文件 | `D:\ai-workbench\web\src\router\index.js` |
| 准入状态 | `SOURCE_READY_DOM_PARTIAL` |
| 编码前阻断 | 必须保留成熟源码的菜单分组、tabs 子级、权限态、404/403/登录态和旧路由兼容；用户侧不得出现外部品牌 |

## 2. 成熟源码导航结构

事实源：`D:\项目迁移文件\平台源码\fe-main\src\components\SideMenu\menu.tsx`

| 一级菜单 key | 用户侧 FindX 命名 | 成熟源码子结构 | FindX 实现要求 |
| --- | --- | --- | --- |
| `infrastructure` | 基础设施 | `business_group` tabs，子项 `/targets` | 保留业务组/资源目标结构；CMDB、主机、Agent 在线作为 FindX 增强能力挂入基础设施或 Agent 管理中心 |
| `integrations` | 集成中心 | `/datasources`、`/components`、`/embedded-products`、动态 embedded product menu | 必须恢复为一级菜单；数据源、模板中心、采集插件、接入指引归这里，不散落到仪表盘或平台治理 |
| `explorer` | 数据查询 | `metrics` tabs：`/metric/explorer`、`/metrics-built-in`、`/object/explorer`、`/recording-rules`；`/log/explorer`；`dashboards` tabs：`/dashboards` | 指标、日志、对象、记录规则、仪表盘同属数据查询；不得把仪表盘拆成独立弱化一级入口 |
| `monitors` | 告警 | `rules` tabs：`/alert-rules`、`/alert-mutes`、`/alert-subscribes`；`job`；`events`；`event-pipelines` | 告警规则、屏蔽、订阅、事件、告警自愈、工作流结构按源码保留；event-pipeline 入口可并入 AI SRE 工作流，但页面结构不改 |
| `notification` | 通知 | `/notification-rules`、`/notification-channels`、`/notification-templates`、deprecated help routes | 通知规则、通知媒介、消息模板保持独立成熟页面；deprecated 入口只做兼容跳转，不做新主入口 |
| `organization` | 人员组织 | `/users`、`/user-groups`、`/roles` | 用户、团队、角色按源码结构接入；接收人、授权团队等对象统一回到团队组织 |
| `setting` | 系统配置 | `ai-config` tabs、`/system/site-settings`、`/system/variable-settings`、`/system/sso-settings`、`/system/alerting-engines`、`/system/version` | 用户侧一级菜单必须叫“系统配置”；AI 模型配置、审计日志作为 FindX 增强项加入，不替换原系统配置结构 |

## 3. 成熟源码路由语义

事实源：`D:\项目迁移文件\平台源码\fe-main\src\routers\index.tsx`

| 路由族 | 核心组件 | 路由动作 | FindX 等价要求 |
| --- | --- | --- | --- |
| 登录与回调 | `Login`、`LoginCallback*` | `/login`、CAS/OAuth/自定义/DingTalk/Feishu 回调 | FindX 使用自有登录与回调；不得嵌入参考站登录或 SSO |
| 权限与异常 | `getMenuPerm`、`Page403`、`NotFound`、`OutOfService` | 菜单权限白名单、403、404、out-of-service | FindX 路由必须保留 401/403/404/BLOCKED 区分，不能全部重定向首页 |
| 动态扩展 | `dynamicPackages`、`dynamicPages`、`plusLoader.routes` | 动态路由注册 | FindX 后续插件/组件扩展必须有注册表，不允许静态假入口 |
| 数据查询 | `MetricExplore`、`LogExplore`、`ObjectExplore`、`RecordingRule` | `/metric/explorer`、`/log/explorer`、`/object/explorer`、`/recording-rules` | 查询页状态流按源码保留；历史、联想、添加面板、Table/Graph 不得静态化 |
| 仪表盘 | `Dashboard`、`DashboardDetail`、`DashboardShare` | `/dashboards`、`/dashboards/:id`、`/dashboards/share/:id` | 仪表盘列表、详情、分享、变量、添加图表同源实现 |
| 告警 | `AlertRules`、`Shield`、`Subscribe`、`Event`、`historyEvents` | 规则、屏蔽、订阅、当前/历史事件 | 不把订阅、屏蔽、事件混成自研页 |
| 工作流 | `TaskTpl`、`Task`、event-pipeline routes | 任务模板/任务执行/事件流水线 | 入口可归 AI SRE，但列表、编辑、调试、执行记录结构不改 |
| 通知 | `notificationChannels`、`notificationRules`、`notificationTemplates` | 规则、媒介、模板 | 按源码页面拆分，禁止用一个复杂抽屉替代 |
| 系统配置 | `VariableConfigs`、`SiteSettings`、`SSOConfigs`、`Servers` | 变量、站点、SSO、告警引擎、版本 | 加入 AI 模型配置和审计日志时不得删原配置族 |

## 4. 当前 FindX 导航差异

事实源：`D:\ai-workbench\web\src\router\nav.js`

| 当前项 | 当前状态 | 与同源结构的差异 | 必须处理 |
| --- | --- | --- | --- |
| 基础设施 | 已存在 | 包含资产中心、业务空间、主机资产、探针与采集；与成熟源码 `infrastructure` 不完全一致，但可作为 CMDB/Agent 增强域 | 保留自有视觉，补业务组、资源目标、CMDB、主机、Agent 在线同源结构 |
| 集成中心 | 缺失一级菜单 | 当前 `/integrations` 只是重定向，不是主导航 | 恢复一级菜单，承载数据源、模板中心、采集插件、接入指引和 embedded product registry |
| 数据查询 | 已存在 | 当前包含数据源/指标/日志/Trace；成熟源码中数据源属于集成中心，仪表盘属于数据查询 | 数据源移回集成中心；数据查询承载指标、日志、对象、记录规则、仪表盘 |
| 监控仪表盘 | 当前一级菜单 | 成熟源码中仪表盘是数据查询下的 tabs，不是独立一级 | 下线独立一级；保留兼容跳转到数据查询 / 仪表盘 |
| 告警 | 已存在但过窄 | 只有规则和事件，缺屏蔽、订阅、告警自愈、event-pipeline 执行结构 | 补全规则、屏蔽、订阅、事件、自愈、工作流兼容入口 |
| 通知 | 已存在但过窄 | 缺消息模板，通知规则/媒介结构需同源 | 补通知规则、通知媒介、消息模板 |
| AI SRE | FindX 增强项 | 非成熟源码一级项，但属于 FindX 长期增强域 | 保留一级；工作流入口指向 AI SRE 工作流，内部保留成熟 event-pipeline/job workflow 结构 |
| 人员组织 | 已存在 | 子项命名与成熟源码基本一致 | 补用户、团队、角色源码结构和权限态 |
| 平台治理 | 当前命名 | 与既定用户侧命名不一致，应为“系统配置” | 改名为系统配置；平台运行自检作为系统配置或 AI SRE 增强入口，不抢占一级命名 |
| 链路监控 | 缺失一级菜单 | SkyWalking UI/OAP 前端链路未在主导航显式承载 | 新增一级菜单“链路监控”，承载服务目录、拓扑、Trace、Profiling、告警、接入；无运行态时显示 BLOCKED |
| Agent 管理中心 | 仅作为基础设施子项表达 | SkyWalking Agent、采集插件、巡检工具、远程安装、心跳、包仓库需要长期控制面 | 可作为基础设施子项或独立增强入口，但用户侧统一叫 FindX Agent / Agent 管理中心 |

## 5. 目标导航结构

FindX 后续实现的主导航固定为：

| 顺序 | 一级菜单 | 来源 | 子项要求 |
| --- | --- | --- | --- |
| 1 | 基础设施 | 成熟源码 + AutoOps/AIOps 增强 | 业务组、资源目标、CMDB、主机资产、Agent 在线、Agent 安装 |
| 2 | 集成中心 | 成熟源码 | 数据源、模板中心、采集插件、接入指引、embedded products 兼容注册表 |
| 3 | 数据查询 | 成熟源码 + SigNoZ/SkyWalking 查询关联 | 指标查询、内置指标、对象查询、记录规则、日志查询、仪表盘 |
| 4 | 告警 | 成熟源码 | 告警规则、屏蔽、订阅、告警事件、告警自愈、事件流水线兼容入口 |
| 5 | 通知 | 成熟源码 | 通知规则、通知媒介、消息模板、旧通知配置兼容跳转 |
| 6 | 人员组织 | 成熟源码 | 用户、团队、角色 |
| 7 | 系统配置 | 成熟源码 + FindX 增强 | AI 模型配置、站点设置、变量设置、SSO、告警引擎、审计日志、版本 |
| 8 | 链路监控 | SkyWalking 源码 | 服务目录、拓扑、Trace、Profiling、告警、接入；未配置时 BLOCKED |
| 9 | AI SRE | FindX 增强 | 诊断会话、知识库、工作流、自动修复、Evidence Chain |

## 6. 兼容跳转规则

| 旧入口 | 新入口 | 要求 |
| --- | --- | --- |
| `/dashboards?section=list` | `/query?section=dashboards` 或等价新路由 | 旧链接必须可用，不得 404 |
| `/query?section=datasources` | `/integrations?section=datasources` | 数据源归集成中心；旧查询入口兼容跳转 |
| `/platform?section=models` | `/system?section=ai-models` 或等价系统配置入口 | 用户侧叫系统配置，不叫平台治理 |
| `/assets?section=agents` | `/assets?section=agents` 或 `/agent-center` | Agent 管理中心命名统一 FindX Agent |
| `/alerts?section=pipeline` | `/aiops?section=workflow` | 入口并入 AI SRE，内部页面结构沿用成熟工作流/事件流水线 |
| `/logs`、`/tracing` | 日志中心、链路监控 | 未配置组件时必须 BLOCKED，不得假成功 |

## 7. 编码准入清单

实现 `P0-N9E-NAV-REAL` 前必须满足：

| 门禁 | 状态 | 说明 |
| --- | --- | --- |
| 源码菜单证据 | `READY` | `menu.tsx` 已读取 |
| 源码路由证据 | `READY` | `routers/index.tsx` 已读取 |
| FindX 当前导航证据 | `READY` | `web/src/router/nav.js` 已读取 |
| FindX 当前路由证据 | `READY` | `web/src/router/index.js` 已读取 |
| 运行态 DOM 证据 | `PARTIAL` | 已有基础监控页面快照；导航展开/收起、快速跳转、窄屏仍需 MCP 补测 |
| 用户侧品牌扫描 | `REQUIRED` | 实现后必须扫描菜单、标题、权限对象、审计对象，外部品牌不得出现在用户侧 |
| 构建验证 | `REQUIRED` | 前端变更后必须同步 WSL 并执行 `npm run build` |
| 浏览器验证 | `REQUIRED` | MCP 登录后点击每个一级和二级入口，覆盖正常、旧链接跳转、未配置 BLOCKED |

未完成以上门禁时，不允许把导航改造标记为 PASS。
