# FindX P0 React-first 壳层与导航迁移矩阵

生成时间：2026-05-07 14:22（UTC+8）
状态：P0-WEB-REACT-SHELL-NAV 编码前门禁，不代表前端迁移已经完成

## 1. 结论

FindX 前端长期目标锁定为 React-first 单一主应用壳层。原因是基础监控事实源和日志中心事实源都是 React 体系，成熟页面结构、状态流、表格、抽屉、查询历史、模板导入、仪表盘编辑等大量能力可以最大程度按同源方式迁移。当前 Vue 自研页面只能作为待替换对象，不再作为成熟能力的设计基线。

链路监控事实源是 Vue 3 + Pinia + Vue Router。FindX 不使用 iframe，也不嵌入参考站或参考站 SSO；SkyWalking 前端能力必须按源码结构和状态流等价迁移到 FindX React 壳层，或在受控迁移期以同仓构建的内部适配层承载，但最终用户侧必须是 FindX 自有登录、导航、主题、权限和品牌。

## 2. 技术栈事实

| 来源 | 技术栈 | 关键依赖 | 对 FindX 的影响 |
| --- | --- | --- | --- |
| 当前 FindX `web/package.json` | Vue 3.5、Vue Router 4、Element Plus、Vite 5 | `vue`、`vue-router`、`element-plus` | 当前壳层与两个 React 成熟源码不兼容，继续在 Vue 上重写会持续丢功能 |
| 基础监控源码 `D:\项目迁移文件\平台源码\fe-main` | React 17、React Router 5、Ant Design 4、Vite、TypeScript | `react`、`react-router-dom`、`antd`、`ahooks`、`uplot`、`react-grid-layout`、`@fc-components/codemirror-promql` | 数据源、指标、仪表盘、模板、告警、通知、系统配置应按 React 同源迁移 |
| 日志中心源码 `D:\项目迁移文件\平台源码\signoz-develop\frontend` | React 18、React Router 5/6 compat、Redux/React Query、Ant Design 5、Webpack | `react`、`react-query`、`antd`、`uplot`、`react-virtuoso`、`xstate` | 日志检索、字段筛选、Pipeline、Saved Views、Trace 关联应按 React 同源迁移 |
| 链路监控源码 `D:\项目迁移文件\平台源码\skywalking-booster-ui-main` | Vue 3、Pinia、Vue Router 4、Element Plus、Vite 6 | `pinia`、`vue-router`、`element-plus`、`echarts`、`vue-grid-layout`、`d3-flame-graph` | 不能 iframe；需把路由、store、GraphQL、topology/trace/profile 状态流等价迁移或受控适配 |

## 3. React-first 壳层边界

| 壳层能力 | FindX 自有实现 | 不允许 |
| --- | --- | --- |
| 登录态 | 自有登录页、token、过期跳转、强制改密、权限上下文 | 嵌入参考站登录、嵌入参考站 SSO、复用参考站 Cookie |
| 导航 | 使用 FindX 主导航规划：基础设施、集成中心、数据查询、告警、通知、人员组织、系统配置、链路监控、AI SRE | 直接搬参考站侧边栏、iframe 内嵌整站导航 |
| 主题 | FindX 主题变量、品牌文案、图标、颜色、间距 | 用户侧出现外部品牌、原始产品标题、原始权限对象 |
| 权限 | FindX 权限域、按钮级权限、403/401、审计上下文 | 直接暴露成熟产品原权限模型为最终公共契约 |
| Adapter | 基础监控、链路、日志、CMDB、Agent、AI SRE 统一 Adapter | 前端直接调用外部 GraphQL/ClickHouse/OAP 原始公共契约 |
| BLOCKED | 组件未配置、源码未落地、运行态不可用时显示明确 BLOCKED | 静态假页面、假按钮、空壳成功态 |

## 4. 迁移策略

| 阶段 | 目标 | 规则 |
| --- | --- | --- |
| P0 壳层冻结 | 建立 React-first 壳层目录、路由、导航、权限上下文、主题变量和 API client 规范 | 不迁移业务页面前先锁接口和导航；旧 Vue 入口只做兼容跳转或明确 BLOCKED |
| P0 基础监控迁移 | 数据源、指标查询、仪表盘、模板中心、告警、通知、系统配置按 React 源码同源迁移 | 以 `fe-main` 路由、组件、services、状态流为事实源；不得低配重写 |
| P1 日志中心迁移 | 日志检索、字段筛选、上下文、聚合、Pipeline、Saved Views、Trace 关联按 React 源码迁移 | 以 SigNoZ `routes.ts`、components、api 目录为事实源；接 FindX Logs Adapter |
| P1 链路监控迁移 | 服务目录、拓扑、Trace、Profiling、告警、接入按 SkyWalking 源码迁移 | Vue/Pinia 状态流必须逐项映射到 React 状态层；未映射完不得声称完成 |
| P2 Agent 管理中心 | FindX Agent 包仓库、远程安装、配置下发、心跳、升级、回滚 | SkyWalking Agent、采集插件、巡检工具作为能力包进入 Agent 管理中心 |

## 5. 当前 Vue 前端处置

| 当前文件/能力 | 处置策略 |
| --- | --- |
| `web/src/App.vue` | 只作为当前壳层事实；React-first 壳层实现后下线或保留兼容期 |
| `web/src/router/nav.js` | 按 P0 导航同源矩阵重建为 React 导航配置；旧路径保留 redirect |
| `web/src/router/index.js` | 迁移为 React Router；旧 Vue 路由保留兼容映射清单 |
| `web/src/views/*Workbench.vue` | 当前多数是自研聚合页；成熟能力迁移完成后不能作为验收基线 |
| `web/src/views/DataSource.vue` | 当前静态类型表单不能作为数据源主路径；由成熟数据源页替换 |
| `web/src/components/monitoring/*` | 当前自研查询/告警面板只做参考，不作为同源实现 |

## 6. SkyWalking 迁移要求

链路监控不得因为事实源是 Vue 而被省略或弱化。迁移必须逐项覆盖：

| 能力 | SkyWalking 源码事实 | React 迁移要求 |
| --- | --- | --- |
| 路由 | `src/router/index.ts`、`src/router/constants.ts` | FindX `/tracing/*` 路由映射服务目录、拓扑、Trace、Profiling、告警、设置 |
| 状态流 | `src/store/modules/dashboard.ts`、`trace.ts`、`topology.ts`、`alarm.ts`、`profile.ts` | 等价迁移 loading、selected、duration、entity、filters、cancel、error |
| GraphQL | `src/graphql/base.ts`、`src/graphql/query/*`、OAP query plugin | 前端只调用 FindX APM Adapter；保留 timeout、AbortController、错误脱敏 |
| 图表/拓扑 | ECharts、D3、grid layout、flame graph | 迁移真实交互，不做静态拓扑图或说明页 |
| Agent 接入 | 多语言 Agent 作为 FindX Agent 能力包 | Java/Python/Node.js/PHP/Go/Rust/Ruby/Nginx Lua/Kong/Browser Client JS 进入包仓库和接入向导 |

## 7. 依赖治理

React-first 迁移会改变前端依赖，必须独立标记依赖变更：

| 依赖类别 | 建议 |
| --- | --- |
| React 核心 | 统一 React 18，避免同时维护 React 17/18 两套运行时 |
| 路由 | 使用 React Router 6，给 React Router 5 来源代码建立兼容适配层 |
| UI 组件 | 统一 FindX 主题 Ant Design 5；对基础监控 AntD 4 组件做兼容封装或逐步升级 |
| 状态 | React Query + Zustand/Redux Toolkit 二选一作为统一状态层；SkyWalking Pinia 语义按模块映射 |
| 图表 | 保留 uPlot、ECharts、D3、react-grid-layout 等成熟源码真实依赖，不为了省组件砍功能 |
| 构建 | Vite 优先；若迁移 SigNoZ Webpack 代码，先抽取源码逻辑再统一构建 |

## 8. 编码准入

| 门禁 | 状态 | 说明 |
| --- | --- | --- |
| 当前 FindX 技术栈证据 | `READY` | `web/package.json` 已读取 |
| 基础监控 React 证据 | `READY` | `fe-main/package.json` 已读取 |
| 日志中心 React 证据 | `READY` | `signoz-develop/frontend/package.json` 已读取 |
| 链路监控 Vue 证据 | `READY` | `skywalking-booster-ui-main/package.json` 已读取 |
| 依赖变更方案 | `REQUIRED` | 迁移前必须列出新增/移除依赖、兼容策略和构建命令 |
| 路由兼容方案 | `REQUIRED` | 旧 Vue 路由、历史链接、用户收藏链接必须有 redirect 或 BLOCKED |
| WSL 构建 | `REQUIRED` | 前端变更后执行 `cd /opt/ai-workbench/web && npm run build` |
| MCP 回归 | `REQUIRED` | 登录后点击每个一级/二级菜单；链路/日志未配置时必须 BLOCKED |

未满足以上门禁时，不允许开始 React 壳层代码迁移，也不允许把 Vue 自研页面标记为成熟源码一比一完成。
