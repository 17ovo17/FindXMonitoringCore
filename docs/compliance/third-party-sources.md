# 第三方来源与融合登记

生成时间：2026-05-07 09:09（UTC+8）  
状态：内部合规与开发证据文档，不作为用户侧展示文案

本文记录 FindX 可能参考、迁移、适配、编排或授权衍生的第三方来源。正式分发前必须以实际导入文件清单、版本、许可证、NOTICE、修改说明和授权材料复核。用户侧菜单、页面标题、权限对象、审计对象、通知内容和 Agent 包展示名称必须使用 FindX / FindX Agent 品牌。

## 来源总表

| 来源 | 典型许可证/授权 | FindX 用途 | 用户侧品牌策略 | 分发前必查 |
| --- | --- | --- | --- | --- |
| Nightingale | Apache-2.0 | 基础监控页面结构、路由、告警、通知、仪表盘、模板、数据查询、系统配置 | 用户侧显示 FindX | 版本、文件清单、NOTICE、修改说明 |
| SkyWalking UI/OAP | Apache-2.0 | 链路监控 UI、OAP query protocol、拓扑、Trace、Profiling、告警 | 用户侧显示链路监控 | 版本、OAP API 边界、NOTICE、存储依赖说明 |
| SkyWalking Java Agent | Apache-2.0 | FindX Agent Java 探针能力包 | 用户侧显示 FindX Agent Java 探针 | 版本、插件清单、分发方式 |
| SkyWalking Python Agent | Apache-2.0 | FindX Agent Python 探针能力包 | 用户侧显示 FindX Agent Python 探针 | 版本、运行时依赖 |
| SkyWalking Node.js Agent | Apache-2.0 | FindX Agent Node.js 探针能力包 | 用户侧显示 FindX Agent Node.js 探针 | 版本、运行时依赖 |
| SkyWalking PHP Agent | Apache-2.0 | FindX Agent PHP 探针能力包 | 用户侧显示 FindX Agent PHP 探针 | 版本、扩展安装说明 |
| SkyWalking Go Agent | Apache-2.0 | FindX Agent Go 探针能力包 | 用户侧显示 FindX Agent Go 探针 | 版本、构建/注入方式 |
| SkyWalking Rust Agent | Apache-2.0 | FindX Agent Rust 探针能力包 | 用户侧显示 FindX Agent Rust 探针 | 版本、运行时依赖 |
| SkyWalking Ruby Agent | Apache-2.0 | FindX Agent Ruby 探针能力包 | 用户侧显示 FindX Agent Ruby 探针 | 版本、运行时依赖 |
| SkyWalking Nginx Lua Agent | Apache-2.0 | FindX Agent 网关/反向代理探针能力包 | 用户侧显示 FindX Agent 网关探针 | OpenResty/Nginx 兼容矩阵 |
| SkyWalking Kong Agent | Apache-2.0 | FindX Agent Kong 插件能力包 | 用户侧显示 FindX Agent Kong 探针 | Kong 版本矩阵 |
| SkyWalking Client JS | Apache-2.0 | FindX 前端监控探针能力包 | 用户侧显示 FindX Browser Agent | 浏览器兼容、隐私脱敏 |
| BanyanDB | Apache-2.0 | SkyWalking 链路监控推荐存储候选 | 用户侧不展示 | 部署、备份、容量、许可证 |
| SigNoZ | Apache-2.0 | 日志中心、日志检索、Pipeline、Saved Views、Trace 关联 | 用户侧显示日志中心 | 版本、文件清单、NOTICE |
| ClickHouse | Apache-2.0 | 日志和分析存储 | 用户侧不展示 | 数据保留、容量、备份 |
| OpenTelemetry Collector | Apache-2.0 | 遥测接入、协议转换、pipeline 管理 | 用户侧显示接入网关或采集网关 | processor/exporter 清单 |
| AutoOps/AIOps | 以源码授权为准 | CMDB、主机、Agent 在线、部署和心跳结构 | 用户侧显示 CMDB / Agent 管理 | 授权边界、导入文件清单 |
| Categraf | MIT | FindX Agent 采集插件生态、配置模板、指标采集能力包 | 用户侧显示 FindX Agent 插件 | MIT 文本、插件清单、修改说明 |
| Catpaw | 授权衍生 | FindX Agent 巡检、诊断、结构化执行、远程会话能力 | 用户侧显示 FindX Agent 诊断能力 | `<AUTH_RECORD>`、衍生范围 |
| Qdrant | Apache-2.0 | 知识库向量索引加速层 | 用户侧显示向量索引或知识库索引 | 数据可重建、备份策略 |
| Prometheus | Apache-2.0 | 指标采集/查询兼容 | 用户侧显示指标数据源 | 数据源配置和许可证 |
| VictoriaMetrics | Apache-2.0 | 指标存储和查询候选 | 用户侧显示指标数据源 | 版本、容量、许可证 |
| Redis | BSD-3-Clause 或发行版本对应许可 | 缓存、会话、队列 | 用户侧不展示 | 版本、部署和安全配置 |
| MariaDB/MySQL | GPL/商业许可等，按发行版本确认 | 权威业务数据库 | 用户侧不展示 | 版本、驱动、迁移和备份 |
| MinIO | AGPLv3/商业许可，按使用方式确认 | 对象存储、证据产物、离线包 | 用户侧不展示 | 分发模式和许可证复核 |

## SkyWalking Agent 上游仓库登记

以下仓库只允许作为内部来源、开发证据、授权边界和源码矩阵引用。用户侧包名、菜单、页面标题、权限对象和审计对象必须使用 FindX Agent 命名。

| 能力包 | source_url | FindX 用户侧名称 | P0 源码矩阵要求 |
| --- | --- | --- | --- |
| Java Agent | `https://github.com/apache/skywalking-java` | FindX Agent Java 探针 | 记录源码路径、版本/commit、插件清单、JVM 参数、配置模板、安装/升级/回滚 |
| Python Agent | `https://github.com/apache/skywalking-python` | FindX Agent Python 探针 | 记录源码路径、版本/commit、运行时依赖、启动注入、配置模板、安装/升级/回滚 |
| Node.js Agent | `https://github.com/apache/skywalking-nodejs` | FindX Agent Node.js 探针 | 记录源码路径、版本/commit、npm 包、框架插件、启动参数、配置模板 |
| PHP Agent | `https://github.com/apache/skywalking-php` | FindX Agent PHP 探针 | 记录源码路径、版本/commit、扩展安装、FPM/Apache 兼容、配置模板 |
| Go Agent | `https://github.com/apache/skywalking-go` | FindX Agent Go 探针 | 记录源码路径、版本/commit、构建/注入方式、模块依赖、回滚策略 |
| Rust Agent | `https://github.com/apache/skywalking-rust` | FindX Agent Rust 探针 | 记录源码路径、版本/commit、crate、运行时、配置模板、数据到达验证 |
| Ruby Agent | `https://github.com/apache/skywalking-ruby` | FindX Agent Ruby 探针 | 记录源码路径、版本/commit、gem、框架兼容、启动注入、配置模板 |
| Nginx Lua Agent | `https://github.com/apache/skywalking-nginx-lua` | FindX Agent 网关探针 | 记录源码路径、版本/commit、OpenResty/Nginx 兼容、Lua 模块、重载/回滚 |
| Kong Agent | `https://github.com/apache/skywalking-kong` | FindX Agent Kong 探针 | 记录源码路径、版本/commit、Kong 版本矩阵、插件安装、接入验证 |
| Browser Client JS | `https://github.com/apache/skywalking-client-js` | FindX Browser Agent | 记录源码路径、版本/commit、SDK 引入方式、隐私脱敏、RUM 事件、SourceMap 策略 |

当前本地成熟源码根目录为 `D:\项目迁移文件\平台源码`，已确认存在 SkyWalking UI/OAP 相关源码目录；旧路径 `D:\平台源码` 仅为迁移前历史绝对路径。独立 Agent 仓库如未落地，P0 阶段必须补齐源码目录、版本登记或明确 `BLOCKED`，不能直接进入实现。

## 品牌脱敏规则

- 用户侧统一使用 FindX、FindX Agent、链路监控、日志中心、Agent 管理中心、AI SRE、模板中心等命名。
- 外部来源名称只允许出现在本文件、源码矩阵、内部审计、授权边界、NOTICE、归档文档和开发证据中。
- 从 Categraf/Catpaw 迁移或适配的能力，用户侧统一称 FindX Agent 插件、FindX Agent 诊断能力或 FindX Agent 巡检能力。
- 基础监控源码中的外部品牌标识在用户侧替换为 FindX，不改变原功能语义。

## SkyWalking 与 SigNoZ 存储边界

- SkyWalking OAP 默认优先使用 BanyanDB 作为链路监控存储候选，也保留 ES/OpenSearch 兼容选项。
- SigNoZ 默认围绕 ClickHouse 承载日志、事件和查询分析能力。
- ClickHouse 不是 SkyWalking OAP 的直接替代存储；如果要共用数据，必须通过 OTel Collector、Adapter 或数据同步设计完成，不得把两个产品的存储契约混用。
- ES/OpenSearch 只在链路兼容或日志搜索明确需要时作为可选组件引入。

## FindX Agent 来源边界

FindX Agent 是控制面、安装器和能力包编排层，不把所有探针强行揉成一个运行时二进制。多语言 Agent、浏览器 Agent、网关 Agent、采集插件、巡检工具作为被 FindX Agent 管理的能力包存在。

分发前必须记录：

- 能力包名称和用户侧展示名称。
- 上游来源、版本、commit、许可证。
- 导入文件或二进制清单。
- 修改说明和 NOTICE。
- Windows/Linux/Kubernetes 支持范围。
- 安装、升级、回滚、卸载和审计行为。

## 版本登记模板

| 字段 | 示例占位 |
| --- | --- |
| source_name | `<SOURCE_NAME>` |
| source_url | `<SOURCE_URL>` |
| version_or_commit | `<VERSION_OR_COMMIT>` |
| license | `<LICENSE>` |
| imported_files | `<FILE_LIST>` |
| modified_files | `<MODIFIED_FILE_LIST>` |
| findx_module | `<FINDX_MODULE>` |
| user_facing_brand | `<FINDX_BRAND_NAME>` |
| distribution_mode | `<SOURCE_OR_BINARY_OR_DOC_ONLY>` |
| notice_required | `<YES_OR_NO>` |
| security_review | `<PENDING_OR_DONE>` |

## 当前状态

本文件是来源登记和边界说明，不代表所有来源已经进入业务代码。每个开发切片仍必须在源码矩阵中记录实际引用路径、实现范围、修改清单和验收证据。
