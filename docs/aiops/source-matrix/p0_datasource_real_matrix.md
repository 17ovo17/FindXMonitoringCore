# FindX P0 数据源页面同源矩阵

生成时间：2026-05-07 14:41（UTC+8）
状态：P0-N9E-DATASOURCE-REAL 编码前门禁，不代表 FindX 数据源页面已经完成

## 1. 结论

数据源页面必须按成熟源码的列表、搜索、类型选择、配置表单、详情、启停、删除、认证、Header、TLS、集群、保存并测试流程同源实现。当前 FindX 自研 `DataSource.vue` 和 `MonitorDatasourceQueryPanel.vue` 不能作为主路径验收基线。

| 项目 | 事实 |
| --- | --- |
| 页面源码 | `D:\项目迁移文件\平台源码\fe-main\src\pages\datasource` |
| 路由源码 | `D:\项目迁移文件\平台源码\fe-main\src\routers\index.tsx`：`/datasources`、`/datasources/:action/:type`、`/datasources/:action/:type/:id` |
| API 源码 | `D:\项目迁移文件\平台源码\fe-main\src\pages\datasource\services.ts` |
| 当前 FindX 弱化页 | `D:\ai-workbench\web\src\views\DataSource.vue`、`D:\ai-workbench\web\src\components\monitoring\MonitorDatasourceQueryPanel.vue` |
| 参考 DOM | `docs\aiops\source-matrix\evidence\n9e_datasources_snapshot.md` |
| 编码准入 | `SOURCE_READY_DOM_PARTIAL`；新增/编辑/详情/错误/权限 DOM 仍需 MCP 补齐 |

## 2. 页面结构清单

| 区域 | 成熟源码证据 | 结构与动作 | FindX 要求 |
| --- | --- | --- | --- |
| 页面容器 | `index.tsx` + `PageLayout title={t('title')}` | 页面标题、外层 `fc-border` 内容区 | FindX 视觉换皮，但保留标题区和内容密度 |
| 搜索 | `index.tsx` `Input prefix={<SearchOutlined />}` + `useDebounce` | 300px 搜索框，500ms debounce，过滤表格数据源名称 | 不能做大块静态搜索；搜索结果必须驱动表格 |
| 新增入口 | `index.tsx` `Button type='primary'` | 打开数据源类型选择 Modal | 按钮真实打开 Modal，不能直接跳静态表单 |
| 类型选择 | `SourceCards/index.tsx` | Modal 宽 960，按 plugin list 渲染图标卡片，点击进入 `/datasources/add/:type` | 图标、分类、说明需保留；用户侧命名 FindX |
| 列表表格 | `TableSource/index.tsx` | AntD small table、分页、排序、筛选、loading | FindX 表格必须保留过滤、排序、分页和 loading |
| 详情 | `Detail.tsx` | 点击名称打开详情抽屉/详情视图 | 详情不能省略；字段脱敏 |
| 新增/编辑表单 | `Form.tsx` + `Datasources/Form.tsx` | breadcrumb、loading、按类型分发到具体表单 | 必须保留类型分发表单，不用一个通用弱表单代替 |

## 3. 表格列和按钮动作

| 列/动作 | 源码 | 行为 |
| --- | --- | --- |
| 类型 | `TableSource defaultColumns[plugin_type]` | 显示数据源图标和类型名，支持类型 filters、排序 |
| 名称 | `Rename` 包裹 `<a onClick nameClick>` | 点击名称打开详情；支持重命名；默认数据源显示 check 图标 |
| 集群 | `cluster_name` + `getServerClusters()` | 显示集群名；集群不存在时 Warning 图标提示 |
| 状态 | `status` | enabled/disabled 图标和文字，支持状态 filters、排序 |
| 启停 | `Popconfirm` + `updateDataSourceStatus` | enabled 切 disabled，disabled 切 enabled，成功后刷新列表 |
| 删除 | `Modal.confirm` + `deleteDataSourceById` | 仅 disabled 状态显示删除；删除后刷新 |
| 授权 | plus `AuthList` | 支持授权类型的扩展编辑 | FindX 需要映射为权限/凭据引用，不回显真实凭据 |
| 更多 | CloudWatch label mapping | 三点菜单扩展动作 | FindX 后续扩展必须走同类菜单，不做侧边栏孤岛 |

## 4. API 契约

事实源：`D:\项目迁移文件\平台源码\fe-main\src\pages\datasource\services.ts`

| 成熟 API | 方法 | 用途 | FindX Adapter 要求 |
| --- | --- | --- | --- |
| `/api/n9e/datasource/plugin/list` | POST | 获取数据源插件类型清单 | FindX 用户侧命名脱敏；返回类型、分类、图标、说明 |
| `/api/n9e/datasource/list` | POST | 获取数据源列表 | 必须支持类型、名称、状态、集群、权限过滤 |
| `/api/n9e/datasource/desc` | POST | 获取详情 | 凭据脱敏；敏感字段只返回 `secret_set`/引用 |
| `/api/n9e/datasource/upsert` | POST | 新增/编辑，含 `is_test`、`force_save` | API_CONTRACT_CHANGE；必须支持保存并测试/强制保存语义 |
| `/api/n9e/datasource/status/update` | POST | 启停 | 权限、审计、幂等 |
| `/api/n9e/datasource` | DELETE | 删除 | 仅 disabled 可删；删除需审计 |
| `/api/n9e/server-clusters` | GET | 获取集群 | 用于集群选择和缺失提示 |

FindX 不要求用户侧 API 名称保留外部前缀，但必须保持语义等价。若新增 `/api/v1/findx/datasources/*`，必须提供旧路径兼容或明确迁移策略。

## 5. Prometheus 表单结构

事实源：`D:\项目迁移文件\平台源码\fe-main\src\pages\datasource\Datasources\Prometheus\Form.tsx`

| 表单能力 | 源码组件 | 要求 |
| --- | --- | --- |
| 名称 | `Name` | 必填、编辑回填、错误定位 |
| HTTP URL | `HTTP` | URL 提示包括 Prometheus/Thanos/VictoriaMetrics/M3/SLS/Mimir |
| Basic Auth | `BasicAuth` | 凭据引用，不回显真实密码 |
| Skip TLS Verify | `SkipTLSVerify` | TLS 校验可配置，默认安全策略需明确 |
| mTLS | `MTLS` | 高级设置折叠，证书字段脱敏 |
| Headers | `Headers` | key/value 转换为对象；敏感 header 脱敏 |
| Remote Write URL | `settings.write_addr` | 支持写入地址配置 |
| Read Addr | `settings.internal_addr` | 支持内网读地址 |
| Cluster | `Cluster` | 未选集群时弹确认，支持 focus 回到字段 |
| Default | `is_default` | 企业扩展能力，FindX 需定义默认数据源语义 |
| TSDB type | `settings.prometheus.tsdb_type` | 支持 Prometheus/Thanos/VictoriaMetrics/M3/SLS 搜索选择 |
| Description | `Description` | 备注 |
| Help Drawer | `prom_installation.md` + `Drawer` | 帮助说明抽屉保留，但用户侧品牌和链接需按 FindX 合规处理 |
| Footer | `Footer` | 保存并测试、强制保存等动作必须按源码语义 |

## 6. FindX 当前差异

| 当前实现 | 问题 | 处理 |
| --- | --- | --- |
| `web/src/views/DataSource.vue` 静态类型和默认 URL | 缺 plugin list、图标卡片、表格过滤、启停、详情、保存并测试、TLS/Header/Cluster 完整表单 | 下线为兼容页或替换为同源数据源页 |
| `MonitorDatasourceQueryPanel.vue` 同时承担数据源列表和查询 | 混合数据源管理和指标查询，不符合成熟页面结构 | 数据源管理回集成中心；指标查询回数据查询 |
| `api/internal/monitoring/datasources.go` 只覆盖 Prometheus | 不能支撑多类型数据源中心 | 扩展统一数据源 Adapter 和凭据引用 |
| 配置中 `prometheus.url` 和 `data_sources[]` 并存 | bootstrap 与权威配置边界不清 | `prometheus.url` 仅作为首启默认，MySQL data_sources 为权威 |

## 7. 编码准入

| 门禁 | 状态 | 说明 |
| --- | --- | --- |
| 源码路径 | `READY` | datasource 目录、TableSource、SourceCards、Form、Prometheus Form 已读取 |
| API 证据 | `READY` | services.ts 已读取 |
| DOM 证据 | `PARTIAL` | 已有列表页快照；新增、编辑、详情、删除、测试连接仍需 MCP |
| API_CONTRACT_CHANGE | `EXPECTED` | 新 FindX 数据源 API 必须写契约 |
| DATA_CHANGE | `EXPECTED` | data_sources、credential_refs、audit、test_status 需要表设计 |
| 安全 | `REQUIRED` | 凭据引用不回显；错误脱敏；删除/启停/保存审计 |
| 验证 | `REQUIRED` | WSL 前端 build、后端 test/build、MCP 登录后覆盖新增/编辑/启停/删除/测试连接/权限态 |

未补齐以上门禁时，不允许开始 P0 数据源代码迁移，也不允许把当前自研数据源页判为 PASS。
