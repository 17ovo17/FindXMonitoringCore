# FindX P0 模板中心页面同源矩阵

生成时间：2026-05-07 15:42（UTC+8）
状态：P0-N9E-TEMPLATE-CENTER-REAL 编码前门禁，不代表 FindX 模板中心已经完成

## 1. 结论

模板中心必须按成熟源码同源实现：图标网格、搜索、组件创建/编辑/删除、抽屉详情、采集说明、指标说明、仪表盘模板、告警规则模板、采集模板、Firemap 扩展、预览、导入、导出、批量操作和权限态都不能省。FindX 只替换视觉、品牌、权限审计和文案脱敏；不能把模板中心做成简略卡片或静态说明。

用户侧命名规则：

| 来源文案 | 用户侧替换 | 处理要求 |
| --- | --- | --- |
| `N9E` / `n9e` | FindX | 平台模板、权限对象、页面标题、审计对象均替换 |
| `categraf-agent` / `categraf` 作为采集代理产品名 | findx-agent / FindX Agent | 保留插件名、配置字段、采集含义，不删除模板 |
| 外部说明链接 | FindX 内部文档链接 | 不直接指向外部品牌文档 |
| token、Bearer、DSN 示例 | `<TOKEN>`、`<DB_DSN>` | 代码块和错误态必须脱敏 |

## 2. 源码事实

| 模块 | 源码路径 | 作用 |
| --- | --- | --- |
| 模板中心入口 | `D:\平台源码\fe-main\src\pages\builtInComponents\List.tsx` | 图标网格、搜索、创建、抽屉、Tabs、Readme 编辑 |
| API 服务 | `D:\平台源码\fe-main\src\pages\builtInComponents\services.ts` | builtin components / payloads CRUD |
| 类型 | `D:\平台源码\fe-main\src\pages\builtInComponents\types.ts` | `Component`、`Payload`、`TypeEnum` |
| 组件创建编辑 | `components\ComponentFormModal`、`LogoPicker` | 组件 ident、logo、readme、disabled |
| Payload 创建编辑 | `components\PayloadFormModal`、`ResultModal` | JSON/YAML CodeMirror、cate、name、uuid、批量结果 |
| 采集说明 | `Instructions\index.tsx` | Markdown/编辑 readme |
| 采集模板 | `CollectTpls\index.tsx`、`GroupSelectModal.tsx`、`services.ts` | 采集模板列表、创建采集配置、业务组选择 |
| 指标说明 | `Metrics\index.tsx`、`metricsBuiltin\services.ts` | 内置指标表格、单位、表达式、导入导出、列设置 |
| 仪表盘模板 | `Dashboards\index.tsx`、`Dashboards\Import.tsx`、`Dashboards\Detail.tsx` | dashboard payload 列表、预览、导入业务组、导出 |
| 告警规则模板 | `AlertRules\index.tsx`、`AlertRules\Import.tsx`、`AlertRules\Detail.tsx`、`services.ts` | alert payload 列表、分类、数据源替换、导入规则 |
| 菜单路由 | `SideMenu\menu.tsx`、`routers\index.tsx` | `/components` 归属集成中心 |
| 运行态 DOM | `docs\aiops\source-matrix\evidence\n9e_components_snapshot.md`、`n9e_template_center_live_snapshot.md` | 参考页面结构 |

## 3. 入口与网格结构

事实源：`List.tsx`。

| 区域/动作 | 源码语义 | FindX 要求 |
| --- | --- | --- |
| 页面标题 | `PageLayout title={t('title')}`、`SafetyCertificateOutlined` | 用户侧标题“模板中心”，不出现外部品牌 |
| 搜索框 | `LIST_SEARCH_VALUE` 本地缓存，按 `ident` 模糊过滤 | 真实过滤图标网格 |
| 创建组件 | `ComponentFormModal action='create'`，权限 `/components/add` | 支持 logo、ident、readme、disabled；权限和审计接入 |
| 图标网格 | `builtin-cates-grid`，每项 `img src={item.logo}` + `ident` | 图标必须保留；相同监控类型按组件聚合，不拆散 |
| URL 联动 | `?component=<ident>` | 点击组件打开抽屉并写入 URL；关闭时移除参数 |
| 组件编辑 | `ComponentFormModal action='edit'`，权限 `/components/put` | 修改 logo/readme/disabled 后刷新 |
| 组件删除 | `deleteComponents([id])`，权限 `/components/del` | 确认弹窗、失败错误态、审计 |
| 禁用状态 | `disabled === 1` 显示 `StopOutlined` | 用户侧显示禁用状态，不删除条目 |

运行态 DOM 证据显示至少有 cAdvisor、CloudWatch、Consul、Docker、Elasticsearch、Java、Kafka、Kubernetes、Linux、MySQL、Nginx、Prometheus、Redis、Windows、ClickHouse 等图标项。FindX 实现必须保留图标和组件说明，不允许只列几个模板名称。

## 4. 抽屉详情与 Tabs

事实源：`List.tsx` Drawer + Tabs。

| Tab | 源码组件 | 结构与动作 | FindX 要求 |
| --- | --- | --- | --- |
| 采集说明 | `Instructions` | Markdown/readme 展示；有权限可编辑并 `putComponent` 保存 | 说明、配置示例、部署步骤、注意事项不能简化；敏感示例脱敏 |
| 采集模板 | `CollectTpls`，Plus 条件 | YAML/TOML 模板、业务组选择、跳转创建采集配置 | 用户侧使用 FindX Agent / findx-agent 命名 |
| 指标说明 | `Metrics` | collector/name/unit/expression/note/update_by/操作列 | 内置指标说明必须完整，单位和表达式不丢 |
| 仪表盘 | `Dashboards` | dashboard payload 表、标签、备注、预览、导入业务组、导出 | 导入后生成真实 dashboard 对象 |
| 告警规则 | `AlertRules` | cate 选择、搜索、payload 表、数据源类型替换、导入业务组 | 导入后生成真实 alert rule；通知规则清空策略保留 |
| Firemap | Plus/Ent 扩展 | 扩展模板 | 未启用时明确 BLOCKED/不可用，不做假入口 |

抽屉 footer 在采集说明 tab 下提供“编辑”；切换 tab 使用 `BUILT_IN_ACTIVE_TAB_KEY` 本地缓存。FindX 必须保留 Tab 状态和可编辑状态，不允许把说明、指标、仪表盘、告警规则拆成互不相关入口。

## 5. Payload 类型与 record-rule 缺口

事实源：`types.ts`。

| TypeEnum | 含义 | 当前源码状态 | FindX 要求 |
| --- | --- | --- | --- |
| `collect` | 采集模板 | `CollectTpls` | 映射到 FindX Agent 采集配置模板 |
| `metric` | 内置指标 | `Metrics` | 指标说明、表达式、单位、导入导出 |
| `dashboard` | 仪表盘模板 | `Dashboards` | 导入生成真实仪表盘 |
| `alert` | 告警规则模板 | `AlertRules` | 导入生成真实告警规则 |
| `firemap` | 扩展模板 | Ent/Plus 扩展 | 未启用时 BLOCKED |
| `record-rule` | 记录规则/预计算模板 | 本目录未发现独立 TypeEnum | 不能删除需求；必须在 P1 告警/记录规则矩阵中补成熟源码事实和导入路径，未补前标记 `BLOCKED_SOURCE_NOT_MAPPED` |

## 6. 模板 API 契约

事实源：`services.ts`。

| API | 方法 | 用途 | FindX 要求 |
| --- | --- | --- | --- |
| `/api/n9e/builtin-components` | GET | 获取组件列表 | Adapter 脱敏后提供；用户侧路径不暴露外部品牌 |
| `/api/n9e/builtin-components` | POST | 创建组件 | 权限、审计、logo 校验、ident 唯一 |
| `/api/n9e/builtin-components` | PUT | 更新组件/readme/disabled | 保留说明编辑和禁用状态 |
| `/api/n9e/builtin-components` | DELETE | 删除组件 | 确认、权限、引用检查 |
| `/api/n9e/builtin-payloads/cates` | GET | 获取 payload 分类 | 告警、采集模板分类筛选 |
| `/api/n9e/builtin-payloads` | GET | 查询 payload | 支持 component_id、type、cate、query |
| `/api/n9e/builtin-payload/:id` | GET | 获取单条 payload 内容 | 预览/编辑 |
| `/api/n9e/builtin-payload?uuid=` | GET | 通过 uuid 获取 payload | 详情预览 |
| `/api/n9e/builtin-payloads` | POST | 创建 payload | JSON/YAML 校验，批量结果 ResultModal |
| `/api/n9e/builtin-payloads` | PUT | 更新 payload | system 创建的条目禁编辑策略保留 |
| `/api/n9e/builtin-payloads` | DELETE | 删除 payload | system 条目禁删策略保留 |

模板导入后的目标 API：

| 模板类型 | 源码导入路径 | 目标对象 |
| --- | --- | --- |
| dashboard | `Dashboards\Import.tsx` | 真实 dashboard，进入仪表盘列表和详情 |
| alert | `AlertRules\Import.tsx` + `AlertRules\services.ts` `/api/n9e/busi-group/:id/alert-rules/import` | 真实 alert rule，通知规则置空策略保留 |
| collect | `CollectTpls\index.tsx` 打开 `/collects/add/:group_id?...` | 真实采集配置 |
| metric | `metricsBuiltin` import/export | 内置指标记录 |

## 7. 仪表盘模板语义

事实源：`Dashboards\index.tsx`、`Import.tsx`、`Detail.tsx`。

| 控件/动作 | 源码语义 | FindX 要求 |
| --- | --- | --- |
| 搜索 | 300px Input，debounce 查询 payload | 真实过滤 |
| 表格 | rowSelection、name link、tags、note、updated_by、operations | 不改成卡片式静态列表 |
| 标签过滤 | 点击 Tag 追加 query | 保留标签过滤 |
| 预览 | name Link 到 dashboard/detail `__uuid__` | 预览必须渲染 dashboard JSON |
| 导入业务组 | 单个/批量 selectedRows，`Import` 选择业务组 | 导入生成真实 dashboard |
| 批量导出 | `Export` + `formatBeautifyJsons` | 导出 JSON，敏感字段脱敏 |
| 创建/编辑 | `PayloadFormModal contentMode='json'` | CodeMirror JSON、格式化、校验 |
| 删除 | system 条目禁删，非 system 确认删除 | 权限/审计/错误态 |

## 8. 告警规则模板语义

事实源：`AlertRules\index.tsx`、`Import.tsx`、`Detail.tsx`。

| 控件/动作 | 源码语义 | FindX 要求 |
| --- | --- | --- |
| 分类 | `getCates` + `Select showSearch` | 告警模板按分类聚合 |
| 搜索 | query debounce | 真实过滤 |
| 预览 | Link 到 alert/detail `uuid`，复用告警规则 Form 只读/预览 | 必须展示成熟告警规则结构 |
| 导入业务组 | 选择业务组、数据源 cate/value | 导入时转换 datasource，清空内置通知规则设置 |
| 单个/批量导出 | `formatBeautifyJson` / `formatBeautifyJsons` | JSON 导出 |
| 创建/编辑 | `PayloadFormModal showCate contentMode='json'` | JSON 校验、分类 autocomplete |
| 删除 | system 条目禁删 | 权限和审计 |

## 9. 指标说明语义

事实源：`Metrics\index.tsx`、`metricsBuiltin\services.ts`。

| 控件/动作 | 源码语义 | FindX 要求 |
| --- | --- | --- |
| 表格列 | collector、name、unit、expression、note、updated_by、operations | 列不可省 |
| 单位 | `getUnitLabel`、`UnitPicker` 说明 | 单位说明必须可读 |
| 表达式 | 支持多行 PromQL 展示 | 不截断为一句 |
| 列设置 | `OrganizeColumns` + localStorage | 保留列配置 |
| 导入/导出 | metricsBuiltin Import/Export | 真实导入导出 |
| 权限 | `getMenuPerm` 判断 add/edit/delete | 按权限隐藏操作 |

## 10. 采集模板语义

事实源：`CollectTpls\index.tsx`、`PayloadFormModal`。

| 控件/动作 | 源码语义 | FindX 要求 |
| --- | --- | --- |
| 搜索 | query debounce | 真实过滤 |
| 分类 | cate 字段、`getCates` | 保留分类 |
| 创建采集 | `GroupSelectModal` 后打开 `/collects/add/:group_id?...` | 映射到 FindX Agent 采集配置创建 |
| YAML/TOML 编辑 | `PayloadFormModal contentMode='yaml'`，CodeMirror toml | 保留配置模板编辑 |
| system 条目保护 | system 条目禁删/禁编辑 | 保留内置模板保护 |

## 11. 品牌脱敏与模板说明保留

模板中心是内部开发证据允许记录外部来源，但用户侧必须替换：

| 场景 | 处理 |
| --- | --- |
| 页面标题、菜单、按钮、权限、审计对象 | 只显示 FindX / FindX Agent / findx-agent |
| 模板卡片 `N9E` | 用户侧改为 FindX 平台模板 |
| 采集说明中的 `categraf` 产品名 | 用户侧改为 findx-agent；插件名和配置字段保留 |
| 代码块中的 token/Bearer/DSN/证书路径 | 用 `<TOKEN>`、`<DB_DSN>`、`<CERT_PATH>` 等占位符或示例路径 |
| 外部 HelpLink | 改为 FindX 内部文档入口 |

不能因为脱敏而删除配置说明、模板说明、图标、分类或导入流程。

## 12. 编码准入

| 门禁 | 状态 | 要求 |
| --- | --- | --- |
| 源码路径 | `READY` | 已定位 List、services、types、Instructions、CollectTpls、Metrics、Dashboards、AlertRules |
| DOM 证据 | `READY_PARTIAL` | 已补列表和 cAdvisor 抽屉 DOM；后续还需仪表盘 tab、告警规则 tab、导入弹窗 MCP 点击证据 |
| record-rule 模板 | `BLOCKED_SOURCE_NOT_MAPPED` | 不得删除；P1 记录规则矩阵补源后接入 |
| 品牌脱敏 | `REQUIRED` | 用户侧不出现外部产品名；`categraf-agent` -> `findx-agent` |
| API 契约 | `REQUIRED` | Adapter 字段、导入冲突、回滚、审计、权限、错误码 |
| 验证 | `REQUIRED` | MCP 登录后覆盖搜索、组件抽屉、说明编辑、指标、仪表盘导入、告警导入、采集配置、失败态 |

未满足本矩阵前，不允许把模板中心做成静态简略列表，也不允许删除成熟模板说明来规避品牌或实现复杂度。
