# FindX P1 人员组织与系统配置同源矩阵

生成时间：2026-05-07 13:54（UTC+8）
状态：`P1-N9E-ORG-SYSTEM-REAL` 编码前门禁，不代表 FindX 已完成实现

## 1. 结论

人员组织和系统配置必须按成熟源码的真实页面结构、状态流和 API 语义迁移，不允许重新设计成弱化表单或静态配置页。FindX 只替换用户侧品牌、导航、权限审计、AI 模型统一配置、审计日志和系统治理增强项。

运行态 DOM 证据：

- [人员组织与系统配置运行态结构证据](evidence/n9e_org_system_live_snapshot.md)

## 2. 导航与路由事实源

| 域 | 路由 | 源码事实源 | FindX 用户侧命名 |
| --- | --- | --- | --- |
| 人员组织 | `/users` | `D:\项目迁移文件\平台源码\fe-main\src\pages\user\users.tsx` | 用户管理 |
| 人员组织 | `/user-groups` | `D:\项目迁移文件\平台源码\fe-main\src\pages\user\groups.tsx` | 团队管理 |
| 人员组织 | `/roles` | `D:\项目迁移文件\平台源码\fe-main\src\pages\permissions\index.tsx` | 角色管理 |
| 系统配置 | `/system/site-settings` | `D:\项目迁移文件\平台源码\fe-main\src\pages\siteSettings\index.tsx` | 站点设置 |
| 系统配置 | `/system/variable-settings` | `D:\项目迁移文件\平台源码\fe-main\src\pages\variableConfigs\index.tsx` | 变量设置 |
| 系统配置 | `/system/sso-settings` | `D:\项目迁移文件\平台源码\fe-main\src\pages\help\SSOConfigs` | 单点登录 |
| 系统配置 | `/system/alerting-engines` | `D:\项目迁移文件\平台源码\fe-main\src\pages\help\servers` | 告警引擎 |
| 系统配置增强 | `/system/ai-models` | `D:\项目迁移文件\平台源码\fe-main\src\pages\aiConfig\llmConfigs` + FindX AI 配置域 | AI 模型配置 |
| 系统配置增强 | `/system/audit-logs` | FindX 审计域 | 审计日志 |

菜单事实源：

- `D:\项目迁移文件\平台源码\fe-main\src\components\SideMenu\menu.tsx`
- 人员组织：`/users`、`/user-groups`、`/roles`
- 系统配置：AI 配置 tab、站点设置、变量设置、单点登录、告警引擎、关于

FindX 实现要求：

- FindX 主导航继续使用自有壳层和既定品牌风格。
- 页面内部结构按源码迁移，不能 iframe 参考站。
- 用户侧不出现外部品牌；内部开发证据可保留来源路径。

## 3. 用户管理页面结构

源码：

- `D:\项目迁移文件\平台源码\fe-main\src\pages\user\users.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\user\component\createModal\index.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\user\component\userForm\index.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\services\manage.ts`

结构必须保留：

| 区域 | 成熟结构 | FindX 实现要求 |
| --- | --- | --- |
| 标题区 | 用户管理、使用说明、当前用户角色提示 | 保留标题与帮助入口，帮助内容替换为 FindX 文档 |
| 搜索区 | 关键词输入，触发表格查询 | 不能做静态搜索框 |
| 主操作 | 新增用户 | 打开真实创建弹窗，提交后刷新表格 |
| 表格 | 用户名、显示名、角色、业务组、团队、创建时间、最后活跃时间、操作 | 列结构和分页按源码实现 |
| 操作列 | 编辑、修改密码、启停/禁用、删除、业务组跳转 | 每个按钮必须接真实 API 和权限态 |
| 弹窗 | 用户表单、密码表单、确认删除 | 表单校验、错误提示、敏感字段脱敏必须保留 |
| 空态 | 暂无数据 | 保留空态，不用假数据填充 |

API 事实源：

| 动作 | 源码函数 | 成熟接口语义 |
| --- | --- | --- |
| 列表 | `getUserInfoList` | `GET /api/n9e/users` |
| 新建 | `createUser` | `POST /api/n9e/users` |
| 详情 | `getUserInfo` | `GET /api/n9e/user/:id/profile` |
| 编辑 | `changeUserInfo` | `PUT /api/n9e/user/:id/profile` |
| 状态 | `changeStatus` | `PUT /api/n9e/user/:id/status` |
| 密码 | `changeUserPassword` | `PUT /api/n9e/user/:id/password` |
| 删除 | `deleteUser` | `DELETE /api/n9e/user/:id` |

## 4. 团队与业务组页面结构

源码：

- `D:\项目迁移文件\平台源码\fe-main\src\pages\user\groups.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\user\business.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\user\component\teamForm\index.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\user\component\businessForm\index.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\user\component\addUser\index.tsx`

结构必须保留：

| 页面 | 成熟结构 | FindX 实现要求 |
| --- | --- | --- |
| 团队管理 | 左侧团队列表/树，右侧详情与成员表 | 站点设置决定列表或树形展示；保留搜索、选择、添加成员、删除成员、分页 |
| 业务组管理 | 左侧业务组列表/树，右侧详情与授权团队表 | 业务组展示模式、分隔符、团队授权结构按源码实现 |
| 成员选择 | 弹窗内搜索、表格、多选、保留已选 | 不能简化为普通输入框 |
| 编辑弹窗 | 表单校验、成员权限、角色选择 | 与源码同源状态流，失败脱敏提示 |

API 事实源：

| 动作 | 成熟接口语义 |
| --- | --- |
| 团队列表 | `GET /api/n9e/user-groups` |
| 团队详情 | `GET /api/n9e/user-group/:id` |
| 创建团队 | `POST /api/n9e/user-groups` |
| 编辑团队 | `PUT /api/n9e/user-group/:id` |
| 删除团队 | `DELETE /api/n9e/user-group/:id` |
| 增加团队成员 | `POST /api/n9e/user-group/:id/members` |
| 删除团队成员 | `DELETE /api/n9e/user-group/:id/members` |
| 业务组列表 | `GET /api/n9e/busi-groups` |
| 业务组详情 | `GET /api/n9e/busi-group/:id` |
| 创建业务组 | `POST /api/n9e/busi-groups` |
| 编辑业务组 | `PUT /api/n9e/busi-group/:id` |
| 删除业务组 | `DELETE /api/n9e/busi-group/:id` |
| 增加业务组成员 | `POST /api/n9e/busi-group/:id/members` |
| 删除业务组成员 | `DELETE /api/n9e/busi-group/:id/members` |

## 5. 角色权限页面结构

源码：

- `D:\项目迁移文件\平台源码\fe-main\src\pages\permissions\index.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\permissions\RoleFormModal.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\permissions\Operations.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\permissions\services.ts`

结构必须保留：

| 区域 | 成熟结构 | FindX 实现要求 |
| --- | --- | --- |
| 左侧 | 角色列表，包含内置角色和自定义角色 | 保留选中态、创建、编辑、删除 |
| 右侧 | 角色详情、备注、权限分组 | 权限域按 FindX 导航和功能域映射 |
| 权限树 | 操作项分组与勾选 | 不允许只给一个角色名称输入框 |
| 弹窗 | 创建/编辑角色 | 权限树、备注、校验、错误态必须完整 |

API 事实源：

| 动作 | 成熟接口语义 |
| --- | --- |
| 角色列表 | `GET /api/n9e/roles` |
| 创建角色 | `POST /api/n9e/roles` |
| 编辑角色 | `PUT /api/n9e/roles` |
| 删除角色 | `DELETE /api/n9e/role/:id` |
| 操作点列表 | `GET /api/n9e/operation` |
| 读取角色操作点 | `GET /api/n9e/role/:roleId/operations` |
| 保存角色操作点 | `POST /api/n9e/role/:roleId/operations` |

## 6. 系统配置页面结构

源码：

- `D:\项目迁移文件\平台源码\fe-main\src\pages\siteSettings\index.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\siteSettings\services.ts`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\variableConfigs\index.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\variableConfigs\FormModal.tsx`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\help\SSOConfigs`
- `D:\项目迁移文件\平台源码\fe-main\src\pages\help\servers`

结构必须保留：

| 页面 | 成熟结构 | FindX 增强 |
| --- | --- | --- |
| 站点设置 | 单页表单、展示模式、日志开关、字体、保存 | FindX 品牌设置、登录页标题、错误脱敏设置可作为扩展项 |
| 变量设置 | 列表、创建、编辑、删除、变量值密文处理 | 凭据引用必须不回显真实值 |
| 单点登录 | SSO 配置表单和保存动作 | 不能嵌入参考站 SSO；必须对接 FindX 自有登录态 |
| 告警引擎 | 引擎集群、实例、数据源、心跳表格 | 与 FindX 告警引擎状态和审计统一 |
| AI 模型配置 | AI LLM 配置源码 + FindX 全局 AI 配置域 | 所有 AI 能力统一读取这里，不允许分散到日志、链路、告警或工作流页面 |
| 审计日志 | FindX 增强项 | 记录登录、权限、配置、远程执行、AI 调用、Agent 生命周期、模板导入等动作 |

API 事实源：

| 页面 | 成熟接口语义 |
| --- | --- |
| 站点设置 | `GET /api/n9e/site-info`、`PUT /api/n9e/config` |
| 变量设置 | `GET/POST/PUT/DELETE /api/n9e/user-variable-configs`、`GET /api/n9e/auth/rsa-configs` |
| SSO | `GET/POST /api/n9e/sso-configs`、`PUT/DELETE /api/n9e/sso-config/:id` |
| 告警引擎 | `GET /api/n9e/alerting-engines`、版本和心跳相关接口按 `help/servers` 源码确认 |
| AI LLM | `GET/POST/PUT/DELETE /api/n9e/ai-llm-configs`、模型测试接口按 `aiConfig/llmConfigs/services.ts` 确认 |
| AI Agent | `GET/POST/PUT/DELETE /api/n9e/ai-agents`，FindX 中归入 AI SRE / 系统配置统一模型引用 |

## 7. AI 配置统一规则

用户已经锁定：所有 AI 配置统一进入系统配置 / AI 模型配置。

FindX 实现要求：

- Nightingale 源码里的 AI LLM、AI Agent、Skills、MCP Servers 结构作为实现参考。
- FindX 用户侧入口只显示“AI 模型配置”“AI 能力配置”“AI SRE 工作流”等 FindX 命名。
- 任何日志、链路、告警、模板、工作流页面需要 AI 能力时，只读取系统配置里的模型、凭据引用、超时、审计策略。
- 禁止页面内散落独立 API Key 输入框。
- API 响应、浏览器错误、日志、测试报告不得输出真实密钥、Bearer、Cookie、完整连接串或会话 ID。

## 8. 权限、错误态和审计门禁

| 门禁 | 要求 |
| --- | --- |
| 权限态 | 非管理员访问角色、用户、系统配置、AI 配置、审计日志必须返回 403 或页面权限态 |
| 登录态 | 登录过期必须跳转 FindX 自有登录页，不跳参考站 SSO |
| 错误态 | 下游接口错误显示用户可理解信息，不展示 SQL、内部路径、堆栈、完整连接串 |
| 审计 | 新增、编辑、删除、启停、密码、SSO、AI 模型、变量、告警引擎配置必须进入审计日志 |
| 空态 | 保留成熟空态，不用假数据伪装 |
| BLOCKED | AI 配置源码存在但参考站运行态不可用，编码前必须补 FindX 自有运行态或标记 BLOCKED |

## 9. FindX 替换表

| 内部来源语义 | FindX 用户侧显示 |
| --- | --- |
| n9e / Nightingale | FindX |
| business group | 业务组 |
| user group / team | 团队 |
| alerting engine | 告警引擎 |
| AI LLM configs | AI 模型配置 |
| AI agents | AI 能力配置 / AI SRE Agent 配置 |
| contact / notification helper docs | FindX 通知与集成文档 |

## 10. 编码前阻断项

- 未补运行态 DOM 证据，不允许声明页面一比一完成。
- 未按源码保留用户、团队、角色、变量、SSO、告警引擎真实按钮动作，不允许开发。
- 未接入 FindX 自有登录、权限、审计、错误脱敏，不允许开发。
- AI 配置不能继续分散在各业务页面。
- 用户侧不得出现外部品牌、源码仓库名或参考站标题。
- API-only 验证不能替代 MCP 浏览器真实登录、点击、新建、编辑、删除、权限和异常回归。
