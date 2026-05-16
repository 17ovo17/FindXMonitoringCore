# FindX P1 告警与通知页面同源矩阵

生成时间：2026-05-07 15:48（UTC+8）
状态：P1-N9E-ALERT-NOTIFICATION-REAL 编码前门禁，不代表 FindX 告警与通知已经完成

## 1. 结论

告警和通知必须按成熟源码结构分域实现：告警规则、屏蔽、订阅、事件、告警自愈、工作流、通知规则、通知媒介、消息模板各有独立页面、过滤器、表格、抽屉/弹窗、导入导出、权限和 API 状态流。禁止把订阅、屏蔽、通知媒介、消息模板、事件流水线混成一个自研页面。

FindX 可把“工作流”一级入口并入 AI SRE 工作流，但内部页面必须保留成熟源码的列表、授权团队、编辑、删除、执行记录、节点执行详情和调试/tryrun 结构。所有 AI 相关配置必须统一走“系统配置 / AI 模型配置”，工作流页面不得散落真实 API key、模型凭据或独立 AI 配置入口。

## 2. 导航与路由事实

| 域 | 路由 | 菜单归属 | 源码事实 |
| --- | --- | --- | --- |
| 告警规则 | `/alert-rules`、`/alert-rules/add/:bgid`、`/alert-rules/edit/:id` | 告警 / 规则管理 | `alertRules` |
| 屏蔽规则 | `/alert-mutes`、`/alert-mutes/add/:from?`、`/alert-mutes/edit/:id` | 告警 / 规则管理 Tabs | 成熟源码路由 `Shield`、`AddShield`、`ShieldEdit` |
| 订阅规则 | `/alert-subscribes`、`/alert-subscribes/add`、`/alert-subscribes/edit/:id` | 告警 / 规则管理 Tabs | 成熟源码路由 `Subscribe` |
| 告警事件 | `/alert-cur-events/:eventId?`、`/alert-his-events/:eventId?` | 告警 / 告警事件 | `alertCurEvent`、`historyEvents` |
| 告警自愈 | `/job-tpls`、`/job-tasks` 等 | 告警 / 告警自愈 | `taskTpl`、`task`、`taskOutput` |
| 工作流 | `/event-pipelines`、`/event-pipelines-executions`、`/event-pipelines-executions/detail/:id` | 告警 / 工作流，FindX 入口并入 AI SRE | `eventPipeline` |
| 通知规则 | `/notification-rules` | 通知 / 通知规则 | `notificationRules` |
| 通知媒介 | `/notification-channels`、`/notification-channels/add` | 通知 / 通知媒介 | `notificationChannels` |
| 消息模板 | `/notification-templates` | 通知 / 消息模板 | `notificationTemplates` |

## 3. 告警规则结构

事实源：`D:\项目迁移文件\平台源码\fe-main\src\pages\alertRules`。

| 区域/动作 | 源码组件 | FindX 要求 |
| --- | --- | --- |
| 列表页 | `index.tsx`、`List\index.tsx`、`ListNG.tsx` | 保留业务组侧栏、预置筛选、数据源筛选、级别筛选、搜索、列设置、分页 |
| 新增/编辑 | `Add.tsx`、`Edit.tsx`、`Form\index.tsx` | 真实表单，不做简化弹窗 |
| 基础配置 | `Form\Base.tsx`、`constants.ts` | 名称、备注、业务组、标签、启停、有效时间 |
| 规则配置 | `Form\Rule\Rule\*` | metric/log/host 等规则类型、Prometheus 表达式、Loki、host 值、预览 |
| 触发条件 | `Triggers\*` | code/builder、nodata、recover、join、anomaly trigger |
| 通知配置 | `Notify\index.tsx`、`NotificationRuleSelect.tsx`、`TaskTpls` | 通知规则选择、自愈模板绑定 |
| 事件处理 | `EventSettings`、`PipelineConfigs`、`PipelineConfigsNG` | relabel、annotations、enrich queries、processor、workflow item |
| 批量操作 | `MoreOperations.tsx`、`CloneToBgids`、`CloneToHosts`、`Export`、`Import` | 启停、克隆、导入、导出、删除不能静态化 |
| 事件抽屉 | `List\EventsDrawer` | 规则关联事件必须可查看 |

运行态 DOM 显示告警规则页有 Tabs：告警规则、屏蔽规则、订阅规则；左侧业务组列表；工具栏包含 reload、数据源、级别、搜索、新增、导入、更多操作、列设置；表格列包含状态、类型、名称、更新时间、用户名、启用、操作。FindX 必须保留这一布局。

## 4. 屏蔽与订阅

| 能力 | 源码/路由 | FindX 要求 |
| --- | --- | --- |
| 屏蔽规则列表 | `/alert-mutes` | 保留搜索、业务组/标签/时间范围/启停/编辑/删除；不得并入告警规则表格 |
| 屏蔽新增编辑 | `/alert-mutes/add/:from?`、`/alert-mutes/edit/:id` | 保留时间窗口、匹配条件、原因、权限 |
| 订阅规则列表 | `/alert-subscribes` | 保留订阅对象、匹配条件、接收配置、启停、编辑、删除 |
| 订阅新增编辑 | `/alert-subscribes/add`、`/alert-subscribes/edit/:id` | 保留团队/用户/标签匹配和通知关联 |

屏蔽、订阅是告警规则管理域的成熟能力，不得与通知规则、通知媒介、消息模板混为一个入口。

## 5. 告警事件

事实源：`alertCurEvent`、`historyEvents`。

| 能力 | 源码组件 | FindX 要求 |
| --- | --- | --- |
| 当前事件 | `alertCurEvent\pages\List`、`AlertCard`、`AlertTable`、`EventDetailDrawer` | 列表/卡片视图、聚合、数据源筛选、详情抽屉、删除/认领/恢复等真实状态 |
| 历史事件 | `historyEvents\index.tsx`、`ListNG`、`exportEvents.ts` | 历史检索、导出、删除、分页 |
| 事件筛选 | `getFilter`、`getRequestParamsByFilter`、`getProdOptions` | 标签、级别、数据源、时间、业务组筛选 |
| 事件详情 | 路由 `/alert-cur-events/:eventId`、`/alert-his-events/:eventId` | 详情页必须保留事件属性、标签、规则、时间线和关联跳转 |

事件数据必须进入 Evidence Chain；错误脱敏，不暴露内部连接串或规则执行堆栈。

## 6. 告警自愈

事实源：`taskTpl`、`task`、`taskOutput`。

| 能力 | 源码路径 | FindX 要求 |
| --- | --- | --- |
| 自愈模板列表 | `taskTpl\index.tsx` | 模板搜索、标签绑定、克隆、修改、详情、删除 |
| 自愈模板编辑 | `add.tsx`、`modify.tsx`、`clone.tsx`、`tplForm.tsx`、`editor.tsx` | 节点/脚本/参数/主机过滤/超时/输出处理不可省 |
| 执行任务 | `task`、`taskOutput` | 任务列表、添加、结果、详情、输出 |
| 主机筛选 | `hostsFilterModal` | 目标主机选择、预览、权限和审计 |

远程执行、脚本、凭据、输出日志必须接入权限审计和脱敏。

## 7. 工作流 / 事件流水线

事实源：`eventPipeline`。

| 能力 | 源码组件/API | FindX 要求 |
| --- | --- | --- |
| 列表 | `pages\List\index.tsx`、运行态 DOM | 搜索、新增、名称、备注、授权团队、更新人、更新时间、编辑、删除 |
| 新增编辑 | `pages\Add.tsx`、`Edit.tsx`、`Form\index.tsx` | 基本信息、授权团队、输入变量、processor 配置 |
| Processor | `Processor\AISummary.tsx`、`Callback.tsx`、`EventDrop.tsx`、`Relabel.tsx`、`EventUpdate` | 节点类型、参数、tryrun、错误态必须保留 |
| 测试 | `TestModal`、`eventProcessorTryrun`、`eventPipelineTryrun` | 调试不能做静态按钮 |
| 执行记录 | `pages\Executions\index.tsx`、`Detail.tsx`、`ItemDetailDrawer.tsx` | 状态、开始/结束、耗时、触发人、错误、节点结果 |
| API | `/api/n9e/event-pipelines`、`/api/n9e/event-pipeline`、`/api/n9e/event-pipeline-tryrun`、`/api/n9e/event-pipeline-executions` | 入口并入 AI SRE 但 API 语义保留，经 FindX Adapter、权限和审计 |

AI Summary 节点或任何 AI 处理器不得持有独立明文 API key。FindX 实现必须改为引用系统配置 / AI 模型配置，节点只保存模型引用、参数、超时和审计策略。

## 8. 通知规则

事实源：`notificationRules`。

| 区域/动作 | 源码组件/API | FindX 要求 |
| --- | --- | --- |
| 列表 | `pages\List.tsx` | 搜索、状态、渠道、更新时间、操作 |
| 新增编辑 | `pages\Add.tsx`、`Edit.tsx`、`Form\index.tsx` | 规则基本信息、匹配属性、时间窗口、通知媒介、消息模板、升级/抑制 |
| 渠道选择 | `ChannelSelect.tsx`、`ChannelParams\*` | 不同渠道参数表单保留 |
| 模板选择 | `TemplateSelect.tsx` | 消息模板关联 |
| 测试发送 | `TestButton.tsx`、`notifyRuleTest` | 必须真实测试，错误脱敏 |
| 详情 | `pages\Detail\*` | 统计、事件、关联规则、订阅规则 |
| API | `/api/n9e/notify-rules`、`/api/n9e/notify-rule/:id`、`/api/n9e/notify-rule/test` | Adapter 收敛、权限审计 |

## 9. 通知媒介

事实源：`notificationChannels`；运行态 DOM：`n9e_alert_notification_live_snapshot.md`。

| 区域/动作 | 源码组件/API | FindX 要求 |
| --- | --- | --- |
| 列表 | `pages\List`、`ListNG` | 搜索、新增、导入、导出、名称、发送类型、更新人、更新时间、启用、操作 |
| 新增编辑 | `pages\Add.tsx`、`Edit.tsx`、`Form\index.tsx` | 按类型渲染 SMTP、HTTP、Script、Dingtalk、Feishu、Wecom、PagerDuty、FlashDuty 等配置 |
| 联系方式字段 | `ContactKeysSelect` | 联系人字段选择真实加载 |
| 导入导出 | `List\Export.tsx`、services POST/PUT/DELETE | JSON 导入导出、错误态 |
| API | `/api/n9e/notify-channel-configs`、`/api/n9e/notify-channel-config/:id`、`/api/n9e/notify-channel-config/idents` | 密钥字段只引用/脱敏，不回显 |

运行态 DOM 显示通知媒介表格列为名称、发送类型、更新人、更新时间、启用、操作，行操作为克隆、删除。FindX 不得把这些变成抽象“渠道卡片”并丢失导入导出。

## 10. 消息模板

事实源：`notificationTemplates`。

| 区域/动作 | 源码组件/API | FindX 要求 |
| --- | --- | --- |
| 列表 | `pages\List\index.tsx` | 搜索、渠道筛选、模板项、详情 |
| 编辑器 | `FieldWithEditor`、`Editor\HTML.tsx`、`Markdown.tsx` | HTML/Markdown、多字段模板、变量提示 |
| 预览 | `preview` `/api/n9e/events-message` | 真实事件预览，错误脱敏 |
| CRUD | `/api/n9e/message-templates`、`/api/n9e/message-template/:id` | 创建、编辑、删除、权限审计 |

消息模板不得和通知媒介合并；通知规则通过模板选择引用消息模板。

## 11. API 契约门禁

| 契约 | 标记 | 必须说明 |
| --- | --- | --- |
| 告警规则 CRUD/导入/导出/启停/克隆 | `API_CONTRACT_CHANGE` 若改公共契约 | 字段、错误码、权限、审计、模板导入兼容 |
| 屏蔽/订阅 | `API_CONTRACT_CHANGE` | 匹配条件、时间、启停、权限 |
| 当前/历史事件 | `API_CONTRACT_CHANGE` | 查询过滤、分页、导出、删除、详情、脱敏 |
| 自愈任务 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | 远程执行、目标主机、脚本、输出、回滚、审计 |
| 工作流/事件流水线 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | 节点 schema、执行记录、AI 模型引用、tryrun |
| 通知规则/媒介/模板 | `API_CONTRACT_CHANGE` + `DATA_CHANGE` | 密文字段、导入导出、测试发送、预览、统计 |

所有写操作必须接入 FindX 权限、审计、幂等、错误脱敏和登录过期跳转。

## 12. 品牌与 AI 配置

| 场景 | 要求 |
| --- | --- |
| 用户侧菜单 | 告警、通知、AI SRE、系统配置等 FindX 命名；不出现外部品牌 |
| 工作流入口 | 可在 AI SRE 中承载，但内部沿用成熟工作流结构 |
| AI Summary / AI 处理器 | 只引用系统配置 / AI 模型配置；不保存明文 key |
| 通知媒介密钥 | webhook、SMTP、PagerDuty、FlashDuty、Feishu 等密钥只使用凭据引用或脱敏字段 |
| 外部 HelpLink | 替换为 FindX 内部文档 |

## 13. 编码准入

| 门禁 | 状态 | 要求 |
| --- | --- | --- |
| 源码路径 | `READY_PARTIAL` | 已定位 alertRules、eventPipeline、taskTpl、alertCurEvent、historyEvents、notificationRules、notificationChannels、notificationTemplates |
| DOM 证据 | `READY_PARTIAL` | 已补告警规则、通知媒介、工作流 DOM；后续需屏蔽、订阅、事件、自愈、通知规则、消息模板点击证据 |
| 工作流 AI 配置 | `REQUIRED` | AI 配置统一系统配置引用 |
| 权限态 | `REQUIRED` | 启停、克隆、删除、导入导出、测试发送、远程执行 |
| 错误态 | `REQUIRED` | 数据源不可用、规则校验失败、通知测试失败、工作流 tryrun 失败、远程执行失败 |
| 验证 | `REQUIRED` | MCP 登录后覆盖每个告警/通知二级页面，不能只测 API |

未满足本矩阵前，不允许继续堆自研弱化告警/通知页面，也不允许把成熟源码中有真实动作的按钮做成静态展示。
