# AutoOps CMDB 与 Agent 源码检查证据

生成时间：2026-05-07 14:04（UTC+8）

## 1. 检查范围

本证据记录 `P2-AUTOOPS-CMDB-AGENT-REAL` 编码前源码检查结果。当前只确认本地源码结构，未获得可访问运行态站点，因此运行态 DOM 状态为 `BLOCKED_RUNTIME_UNAVAILABLE`。

源码根目录：

- `D:\平台源码\AutoOps-main\AutoOps-main`

前端关键目录：

- `web\src\views\cmdb`
- `web\src\views\cmdb\Host`
- `web\src\views\Tools`
- `web\src\api\cmdb.js`
- `web\src\api\tool.js`
- `web\src\router\cmdb.js`
- `web\src\router\Tools.js`

## 2. 路由证据

| 路由 | 源码组件 | 源码标题语义 |
| --- | --- | --- |
| `/cmdb/ecs` | `views/cmdb/cmdbHost.vue` | 资产管理 / 主机管理 |
| `/cmdb/group` | `views/cmdb/cmdbGroup.vue` | 资产管理 / 业务分组 |
| `/cmdb/db` | `views/cmdb/cmdbDB.vue` | 资产管理 / 数据管理 |
| `/cmdb/ssh` | `views/cmdb/Host/SSH.vue` | 资产管理 / 终端登录 |
| `/cmdb/dbdetails` | `views/cmdb/DBdetails.vue` | 数据详情 |
| `/ops/tools` | `views/Tools/Tools.vue` | 运维工具 |
| `/ops/agent` | `views/Tools/Agent.vue` | agent 列表 |

FindX 用户侧必须统一显示 FindX 品牌和 FindX Agent 命名，不显示外部产品名。

## 3. CMDB 主机页面源码结构

源码：

- `web\src\views\cmdb\cmdbHost.vue`
- `web\src\views\cmdb\Host\CmdbGroup.vue`
- `web\src\views\cmdb\Host\CmdbHostTable.vue`
- `web\src\views\cmdb\Host\CreateHost.vue`
- `web\src\views\cmdb\Host\EditHost.vue`
- `web\src\views\cmdb\Host\CreateCloud.vue`
- `web\src\views\cmdb\Host\CreateExcel.vue`
- `web\src\views\cmdb\Host\HostSsh.vue`
- `web\src\views\cmdb\Host\Terminal.vue`
- `web\src\views\cmdb\Host\MonitorDialog.vue`
- `web\src\views\cmdb\Host\ProcessMonitorDialog.vue`
- `web\src\views\cmdb\Host\TcpPortMonitorDialog.vue`

源码结构摘要：

- 左侧分组树：搜索、展开、折叠、全部展开、全部折叠、新建分组、编辑分组、删除分组。
- 右侧主机筛选：主机名称、IP 地址、主机状态、搜索、重置。
- 新建下拉：导入主机、Excel 导入、云主机。
- 主机操作：终端、详情、编辑、上传、执行命令、删除、监控。
- 主机表格：主机名称、公网/内网 IP、CPU 使用率、内存使用率、磁盘使用率、进程、端口、配置、存活状态、认证状态、主机类型、操作。
- 主机详情抽屉：仪表盘、基本信息、连接地址、认证类型、描述、监控数据。
- 监控弹窗：主机总览、进程监控、TCP 端口监听状态。
- 文件上传弹窗：目标主机、目标路径、文件选择、进度。

## 4. Agent 页面源码结构

源码：

- `web\src\views\Tools\Agent.vue`
- `web\src\views\Tools\SelectDeployHost.vue`
- `web\src\api\cmdb.js`

源码结构摘要：

- 搜索表单：主机名称、Agent 状态、版本、搜索、重置。
- 操作按钮：部署 Agent、批量卸载。
- 列表：主机名称、IP 地址、版本、状态、监听端口、安装进度、健康状态、最后心跳、更新时间、操作。
- 状态枚举：部署中、部署失败、运行中、启动异常。
- 操作列：重新部署、重启、卸载、删除、查看详情。
- 详情弹窗：主机名称、IP 地址、版本、状态、安装路径、监听端口、进程 ID、健康状态、最后心跳、创建时间、更新时间、错误信息。
- 部署选择：通过 `SelectDeployHost.vue` 选择主机后批量提交部署。
- 状态流：部署、卸载、重启后调用列表刷新并启动轮询；源码中 `MAX_POLLING_COUNT = 5`，FindX 后续必须重新评估商业化长任务轮询和后台任务模型。

## 5. API 证据

CMDB API 源码：`web\src\api\cmdb.js`

| 能力 | 成熟接口语义 |
| --- | --- |
| 主机列表 | `GET cmdb/hostlist`、`GET cmdb/host/list` |
| 主机详情 | `GET cmdb/hostinfo` |
| 新建主机 | `POST cmdb/hostcreate` |
| 编辑主机 | `PUT cmdb/hostupdate` |
| 删除主机 | `DELETE cmdb/hostdelete` |
| 按名称/IP/状态/分组搜索 | `cmdb/hostbyname`、`cmdb/hostbyip`、`cmdb/hostbystatus`、`cmdb/hostgroup` |
| 分组列表 | `GET cmdb/grouplist`、`GET cmdb/grouplistwithhosts` |
| 分组新增/编辑/删除 | `cmdb/groupadd`、`cmdb/groupupdate`、`cmdb/groupdelete` |
| 云主机导入 | `cmdb/hostcloudcreatealiyun`、`cmdb/hostcloudcreatetencent` |
| Excel 导入/模板 | `POST cmdb/hostimport`、`GET cmdb/hosttemplate` |
| SSH WebSocket | `/api/v1/cmdb/hostssh/connect/:hostId` |
| 上传文件 | `POST cmdb/hostssh/upload/:hostId` |
| 执行命令 | `GET cmdb/hostssh/command/:hostId` |
| 主机监控 | `GET monitor/hosts/:id/all-metrics`、`GET monitor/hosts`、`GET monitor/hosts/:id/top-processes`、`GET monitor/hosts/:id/ports` |
| Agent 部署 | `POST monitor/agent/deploy` |
| Agent 卸载 | `DELETE monitor/agent/uninstall` |
| Agent 状态 | `GET monitor/agent/status/:id` |
| Agent 重启 | `POST monitor/agent/restart/:id` |
| Agent 列表 | `GET monitor/agent/list` |
| Agent 删除 | `DELETE monitor/agent/delete/:id` |

部署管理 API 源码：`web\src\api\tool.js`

| 能力 | 成熟接口语义 |
| --- | --- |
| 可部署服务列表 | `GET /api/v1/tool/services` |
| 服务详情 | `GET /api/v1/tool/services/:serviceId` |
| 创建部署任务 | `POST /api/v1/tool/deploy` |
| 部署历史 | `GET /api/v1/tool/deploy/list` |
| 部署状态 | `GET /api/v1/tool/deploy/:id/status` |
| 卸载服务 | `DELETE /api/v1/tool/deploy/:id` |

## 6. 运行态缺口

当前未提供可访问 AutoOps/AIOps 运行态 URL，无法完成 MCP 浏览器 DOM 证据。

状态：

- `SOURCE_PRESENT`
- `RUNTIME_DOM_BLOCKED`

编码前必须补齐：

- CMDB 主机管理真实页面截图/DOM。
- Agent 管理真实页面截图/DOM。
- 部署/卸载/心跳/错误态真实交互证据。
- 权限态、登录态、空态、接口错误态证据。

未补齐运行态证据前，不允许声明 CMDB / Agent 在线能力完成。
