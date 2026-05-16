# FindX P0 仪表盘页面同源矩阵

生成时间：2026-05-07 15:35（UTC+8）
状态：P0-N9E-DASHBOARD-REAL 编码前门禁，不代表 FindX 仪表盘已经完成

## 1. 结论

仪表盘必须按成熟源码同源实现列表、详情、变量、Panel、编辑器、渲染器、时间范围、刷新、分享、全屏、导入导出和模板导入。用户已经指出的 `datasource`、`server`、IP 变量必须是可选择、可搜索、可编辑、可跟 URL 联动的模板变量；`添加图表` 必须打开真实图表类型菜单和编辑器，不能是静态按钮或错误映射。

| 项目 | 事实 |
| --- | --- |
| 路由源码 | `D:\项目迁移文件\平台源码\fe-main\src\routers\index.tsx` |
| 列表源码 | `D:\项目迁移文件\平台源码\fe-main\src\pages\dashboard\List` |
| 详情源码 | `D:\项目迁移文件\平台源码\fe-main\src\pages\dashboard\Detail` |
| 变量源码 | `D:\项目迁移文件\平台源码\fe-main\src\pages\dashboard\Variables`、`VariableConfig` |
| Panel 源码 | `D:\项目迁移文件\平台源码\fe-main\src\pages\dashboard\Panels` |
| 编辑器源码 | `D:\项目迁移文件\平台源码\fe-main\src\pages\dashboard\Editor` |
| 渲染器源码 | `D:\项目迁移文件\平台源码\fe-main\src\pages\dashboard\Renderer` |
| 时间控件源码 | `D:\项目迁移文件\平台源码\fe-main\src\components\TimeRangePicker` |
| API 源码 | `D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts` |
| 参考 DOM | `docs\aiops\source-matrix\evidence\n9e_dashboards_snapshot.md`、`docs\aiops\source-matrix\evidence\n9e_dashboard_detail_live_snapshot.md` |

## 2. 路由与页面入口

| 路由 | 源码证据 | 页面语义 | FindX 要求 |
| --- | --- | --- | --- |
| `/dashboards` | `routers/index.tsx` `component={Dashboard}` | 仪表盘列表 | 用户侧进入“数据查询 / 仪表盘”或等价 FindX 导航；保留兼容跳转 |
| `/dashboards/:id` | `DashboardDetail` | 仪表盘详情 | 不 iframe，不嵌参考站；按源码结构迁移 |
| `/dashboard/:id` | `DashboardDetail` | 旧路由兼容 | 保留 redirect/alias，不产生双实现 |
| `/dashboards/share/:id` | `DashboardShare` | 公开分享页 | 分享动作、主题参数、权限和脱敏必须保留 |
| `/chart/:ids` | `Chart` | 临时图表分享 | 从 Panel/指标查询分享入口落到真实图表渲染 |

## 3. 列表页结构

事实源：`D:\项目迁移文件\平台源码\fe-main\src\pages\dashboard\List\index.tsx`、`Header.tsx`、`FormModal.tsx`、`Import\index.tsx`、`BatchClone.tsx`、`PublicForm.tsx`。

| 区域/动作 | 源码语义 | FindX 要求 |
| --- | --- | --- |
| 页面标题 | `PageLayout title={t('title')}`、图标、文档入口 | 用户侧标题“仪表盘”，不出现外部品牌文档链接 |
| 业务组侧栏 | `BusinessGroupSideBarWithAll`、`getDefaultGidsInDashboard`、本地选择缓存 | 保留业务组/公开仪表盘切换和筛选，不改成孤立卡片 |
| 搜索 | `searchVal` 本地缓存，过滤列表 | 搜索框要真实过滤列表，不能静态展示 |
| 刷新 | `RefreshIcon` 更新 `refreshKey` | 必须重新拉取列表 |
| 新增 | `FormModal`，调用 `createDashboard` | 真实创建仪表盘，成功后刷新 |
| 导入 | `Import`，支持内置/Grafana/JSON 等导入入口 | 必须保留导入流程、冲突、失败错误态 |
| 批量操作 | `selectRowKeys`、批量克隆、批量删除 | 选择、确认、权限和刷新必须真实 |
| 列设置 | `OrganizeColumns` + `LOCAL_STORAGE_KEY` | 列显示顺序和本地持久化不能丢 |
| 表格列 | 名称、标签、公开状态、更新时间、更新人、操作列等 | 列结构按源码，不压缩成弱化列表 |
| 操作列 | 编辑、分享、公开配置、克隆、导出、删除 | 三点菜单和确认弹窗必须接入真实 API |

## 4. 列表 API 契约

事实源：`D:\项目迁移文件\平台源码\fe-main\src\services\dashboardV2.ts`。

| API | 方法 | 源码用途 | FindX 要求 |
| --- | --- | --- | --- |
| `/api/n9e/busi-group/:id/boards` | GET | 单业务组仪表盘列表 | 通过 FindX Basic Monitoring Adapter 收敛，用户侧不暴露外部品牌路径 |
| `/api/n9e/busi-groups/boards` | GET | 多业务组仪表盘列表 | 支持 gids、权限、空态、错误脱敏 |
| `/api/n9e/busi-groups/public-boards` | GET | 公开仪表盘 | 保留公开分类和分享状态 |
| `/api/n9e/busi-group/:id/boards` | POST | 创建仪表盘 | 审计、权限、重名/非法输入错误态 |
| `/api/n9e/busi-group/:busiId/board/:id/clone` | POST | 克隆仪表盘 | 单个/批量克隆都要真实生成对象 |
| `/api/n9e/boards` | DELETE | 批量删除 | 确认弹窗、权限、审计、失败回滚 |
| `/api/n9e/busi-group/:id/dashboards/export` | POST | 导出仪表盘 | 导出文件脱敏，不带真实密钥 |
| `/api/n9e/board/:id/public` | PUT | 更新公开状态 | 权限和分享状态必须一致 |

以上路径在 FindX 中可以通过 Adapter 映射为新公共契约，但字段、状态流、错误态和权限语义必须同源，不允许只返回静态列表。

## 5. 详情页顶部结构

事实源：`D:\项目迁移文件\平台源码\fe-main\src\pages\dashboard\Detail\Title.tsx`；运行态 DOM 证据：`n9e_dashboard_detail_live_snapshot.md`。

| 控件 | 源码/DOM 证据 | 真实动作 | FindX 要求 |
| --- | --- | --- | --- |
| 返回列表 | DOM `link "仪表盘列表"`，源码 `goBack`/`Link` | 返回 `/dashboards` | 保留列表/详情层级 |
| 仪表盘标题下拉 | `getBusiGroupsDashboards`、`dashboardListDropdownSearch` | 搜索并切换同业务组仪表盘 | 标题不是静态文本，必须可切换 |
| 添加图表 | `visualizations` + `AddPanelIcon` + `onAddPanel` | 打开导入 Panel/行/图表类型菜单 | 必须支持 timeseries、stat、table、pie、gauge、bar、heatmap、text、iframe 等源码类型 |
| 刷新 | `TimeRangePickerWithRefresh` sync 按钮 | 触发刷新和查询 | 不能静态显示 |
| 自动刷新 | `intervalSeconds`、URL `__refresh` | Off/周期选择并刷新 | 必须和查询状态联动 |
| 时间范围 | `__from`、`__to`、localKey `${dashboardTimeCacheKey}_${id}` | 最近 1 小时、自定义、绝对时间 | 必须影响 Panel 查询 |
| 时区 | `showTimezone`、`timezone` | 时区选择并缓存 | 不可省略 |
| 设置 | `SettingOutlined`、`FormModal`、`DashboardLinks`、变量编辑入口 | 编辑基本信息、链接、变量、导入 Grafana URL | 设置必须是真实弹窗/抽屉 |
| 分享链接 | DOM `button "link"` | 分享/公开链接 | 权限、公开状态和脱敏必须接入 |
| 全屏 | URL `viewMode=fullscreen`、Esc 退出、resize | 隐藏壳层并重排 Panel | 必须真实改变模式 |

## 6. 变量系统

事实源：`Variables\index.tsx`、`Main.tsx`、`VariableManagerContext.tsx`、`Variable\*.tsx`、`EditModal\*.tsx`。

| 变量类型 | 源码证据 | 行为 | 用户指出的阻断 |
| --- | --- | --- | --- |
| datasource | `Variable\Datasource.tsx` | 从 `groupedDatasourceList` 按 cate/filter/regex 生成 options，`Select showSearch`，宽 180，URL/状态联动 | 不能显示成静态标签；必须可选择可搜索 |
| datasourceIdentifier | `Variable\DatasourceIdentifier.tsx` | 按 identifier 过滤数据源，支持 regex 和搜索 | 同上 |
| hostIdent | `Variable\HostIdent.tsx` | 主机标识选择，关联资源 | IP/host 变量必须来自模板配置和数据源 |
| query | `Variable\Query.tsx` + `EditModal\Querybuilder.tsx` | 通过数据源 query builder 拉取 options | 查询变量必须能预览、失败展示错误 |
| custom | `Variable\Custom.tsx` | 自定义值列表 | 支持多选/默认值/隐藏 |
| constant | `Variable\Constant.tsx` | 常量变量 | 可作为模板替换 |
| textbox | `Variable\Textbox.tsx` | 可输入值，默认值同步 | 不是固定标签 |
| 变量编辑 | `EditModal\index.tsx` | list/add/edit、上下移动、复制、删除、hide、defaultValue、URL 参数替换 | 后续实现必须完整保留 |

变量值会进入 `replaceTemplateVariables`、`replaceDatasourceVariables`、Panel 查询、Repeat 计算和 URL 参数。FindX 不得把变量区域做成只读 Tag。

## 7. Panel 布局与操作

事实源：`Panels\index.tsx`、`Panel.tsx`、`Row.tsx`、`EditorModal.tsx`、`Panels\utils.ts`。

| 功能 | 源码语义 | FindX 要求 |
| --- | --- | --- |
| 网格布局 | `ReactGridLayout`、`buildLayout`、`onDragStop`、`onResizeStop` | 拖拽/缩放后更新 configs |
| 添加 Panel | `EditorModal` `mode='add'`、`updatePanelsInsertNewPanelToRow` | 真实打开编辑器并保存 |
| 编辑 Panel | `mode='edit'`、`updatePanelsWithNewPanel` | 编辑 query、options、transformations 后保存 |
| 克隆 Panel | 生成新 uuid，插入新 Panel | 复制不是静态按钮 |
| 删除 Panel | `Modal.confirm` 后过滤 panels 并保存 | 删除必须更新 configs 和 UI |
| 复制 Panel 配置 | `navigator.clipboard.writeText` / `copy2ClipBoard` | 复制 JSON 配置 |
| 分享 Panel | `onShareClick` | 临时图表分享 |
| 行 Row | row/collapse/withPanels 删除、行内添加 | Row 不是普通卡片标题 |
| Repeat | `processRepeats`、变量驱动 | 变量改变后重复 Panel 重新计算 |
| 权限/过期 | `editable`、`isAuthorized`、`dashboardSaveMode` | 无权限或配置过期必须提示，不可悄悄失败 |

## 8. 图表编辑器与渲染器

事实源：`Editor`、`Editor\QueryEditor`、`Editor\Options`、`Renderer`、`Renderer\datasource`、`Renderer\Inspect`、`transformations`。

| 区域 | 必须能力 |
| --- | --- |
| 图表类型 | timeseries、stat、table/tableNG、pie、gauge、barGauge、barChart、heatmap、hexbin、text、iframe、row |
| Query Editor | 数据源选择、Prometheus/Elasticsearch/ClickHouse query builder、QueryOptions、表达式、legend、extra actions |
| Options | 标准选项、阈值、颜色、tooltip、legend、value mappings、overrides、data links、图表类型专属样式 |
| Transformations | organize、merge、join、group aggregate、series/table 转换、filter、sort、limit、计算字段等 |
| Renderer | 按 Panel type 渲染真实图表；支持 loading、empty、error、annotation、legend、tooltip、单位格式化 |
| Inspect | query/json tabs、多 query select、CodeMirror 展示真实查询和配置 |
| 数据查询 | `query-range-batch`、`query-instant-batch`、Prometheus proxy labels/series | 必须支持取消请求、超时、错误脱敏 |

## 9. 详情 API 契约

| API | 方法 | 源码用途 | FindX 要求 |
| --- | --- | --- | --- |
| `/api/n9e/board/:id` | GET | 获取详情 | configs 必须完整解析和迁移 |
| `/api/n9e/board/:id` | PUT | 更新名称/ident/tags/note | 权限、审计、输入校验 |
| `/api/n9e/board/:id/configs` | PUT | 更新 panels、变量、links、options | DATA_CHANGE；必须有版本/冲突/回滚策略 |
| `/api/n9e/board/:id/pure` | GET | 手动保存模式离开前比较 | 不可省略，否则会丢修改提醒 |
| `/api/n9e/builtin-boards-detail` | POST | 内置模板详情 | 模板中心导入后必须生成真实 dashboard |
| `/api/:path/query-range-batch` | POST | Panel 区间查询 | 支持 abort、超时、空态、错误脱敏 |
| `/api/:path/query-instant-batch` | POST | Panel 即时查询 | 同上 |
| `/api/:path/proxy/:datasource/api/v1/*` | GET | labels、values、series | 变量和编辑器依赖 |

## 10. 运行态 DOM 证据

MCP 浏览器打开 `http://198.18.20.146:17000/dashboards/4?datasource=1` 后确认：

| DOM 节点 | 证据含义 |
| --- | --- |
| `button "添加图表"` | 顶部必须有真实添加图表入口 |
| `button "sync"` + `button "Off down"` | 刷新和自动刷新存在 |
| `button "最近 1 小时 down"` | 时间范围不是静态文字 |
| `button "setting"`、`button "link"`、`button "fullscreen"` | 设置、分享、全屏存在 |
| `datasource` combobox `普罗米修斯` | 数据源变量可选 |
| `server` combobox `198.18.20.20` | 主机/IP 变量可选、可搜索 |
| Requests / Active connections 等 Panel | Panel 是 dashboard configs + 查询渲染结果 |

## 11. 编码准入

| 门禁 | 状态 | 要求 |
| --- | --- | --- |
| 源码路径 | `READY` | 已定位 List、Detail、Variables、Panels、Editor、Renderer、services |
| DOM 证据 | `READY_PARTIAL` | 已补详情页关键 DOM；后续还需列表页、添加图表菜单、变量编辑、Panel 菜单的 MCP 点击证据 |
| API 契约 | `REQUIRED` | 进入代码前必须决定 Adapter 公开契约、configs 版本、冲突与回滚 |
| 权限态 | `REQUIRED` | `/dashboards/put`、公开分享、无权限编辑、过期配置提示必须覆盖 |
| 错误态 | `REQUIRED` | 查询失败、模板变量失败、导入 JSON 错误、数据源不可用、保存冲突 |
| 验证 | `REQUIRED` | 前端变更后 WSL build；MCP 登录后覆盖列表、详情、变量、添加图表、编辑、分享、全屏、窄屏 |

未满足本矩阵前，不允许继续把当前自研仪表盘页面声明为成熟源码一比一完成，也不允许只补几个按钮冒充仪表盘工作台。
