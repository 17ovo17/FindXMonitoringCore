# SkyWalking Agent 本地源码检查证据

生成时间：2026-05-07 15:55（UTC+8）

## 1. 检查范围

本证据只记录 FindX P0 SkyWalking Agent 接入前置源码检查结果。用户侧产品命名必须使用 FindX Agent / FindX Browser Agent / FindX 网关探针；外部来源名称只允许出现在本类内部证据、合规来源登记、归档和开发矩阵中。

检查目录：

- `D:\平台源码`

已确认存在：

| 本地目录 | 结论 |
| --- | --- |
| `D:\平台源码\skywalking-booster-ui-main` | 链路监控前端 UI 源码存在 |
| `D:\平台源码\skywalking-master` | OAP/API/WebApp proxy 源码存在 |

## 2. 独立 Agent 仓库检查结果

以下独立 Agent 仓库在本地 `D:\平台源码` 下未发现。进入 FindX Agent 包仓库、远程安装、配置下发、心跳、升级回滚、数据到达验证和 MCP 浏览器回归前，必须补齐本地源码、版本/commit、许可证、NOTICE、制品形态、配置项和安装方式。

| 能力包 | 期望本地路径 | 当前状态 |
| --- | --- | --- |
| Java 探针 | `D:\平台源码\skywalking-java` | `BLOCKED_LOCAL_SOURCE_MISSING` |
| Python 探针 | `D:\平台源码\skywalking-python` | `BLOCKED_LOCAL_SOURCE_MISSING` |
| Node.js 探针 | `D:\平台源码\skywalking-nodejs` | `BLOCKED_LOCAL_SOURCE_MISSING` |
| PHP 探针 | `D:\平台源码\skywalking-php` | `BLOCKED_LOCAL_SOURCE_MISSING` |
| Go 探针 | `D:\平台源码\skywalking-go` | `BLOCKED_LOCAL_SOURCE_MISSING` |
| Rust 探针 | `D:\平台源码\skywalking-rust` | `BLOCKED_LOCAL_SOURCE_MISSING` |
| Ruby 探针 | `D:\平台源码\skywalking-ruby` | `BLOCKED_LOCAL_SOURCE_MISSING` |
| Nginx Lua 网关探针 | `D:\平台源码\skywalking-nginx-lua` | `BLOCKED_LOCAL_SOURCE_MISSING` |
| Kong 网关探针 | `D:\平台源码\skywalking-kong` | `BLOCKED_LOCAL_SOURCE_MISSING` |
| Browser Client JS | `D:\平台源码\skywalking-client-js` | `BLOCKED_LOCAL_SOURCE_MISSING` |

## 3. FindX 实施门禁

- 不允许把上述能力只做成静态清单。
- 不允许在缺少本地源码和版本证据时声明 FindX Agent 探针接入完成。
- 不允许用链路监控 UI/OAP 源码替代独立 Agent 仓库证据。
- 不允许把 Agent 在线等同于链路数据到达成功。
- 必须实现包仓库、安装向导、远程安装、配置模板、配置下发、心跳状态、数据到达验证、漂移检测、升级回滚、卸载、审计和 Evidence Chain。

## 4. 后续补齐要求

每个独立 Agent 仓库落地后，必须补充：

- 本地源码路径、版本、commit、许可证、NOTICE。
- 包形态、签名、校验和、离线包策略。
- Linux、Windows、Kubernetes 适配方式。
- 配置模板字段、密文字段引用、默认值和校验规则。
- 远程安装、升级、回滚、卸载状态流。
- OAP 连通、Trace/Metric/Log correlation/RUM 数据到达验证。
- MCP 浏览器真实点击证据和敏感信息扫描结果。
