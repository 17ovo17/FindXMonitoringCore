# FindX AIOps 文档索引

生成时间：2026-05-07 09:09（UTC+8）  
状态：当前有效索引

## 当前唯一实施入口

- [FindX 全栈可观测长期开发计划](findx_full_stack_observability_long_term_plan.md)

开发、评审、测试、归档和 Git 闭环均以该主计划为准。旧方向文档只作为历史证据，不再作为实施依据。

## 开发前必须同步核对

- [FindX 全栈可观测长期开发计划](findx_full_stack_observability_long_term_plan.md)
- [统一测试基准](../testing/AI_WorkBench_统一测试基准.md)
- [第三方来源与融合登记](../compliance/third-party-sources.md)
- [AGENTS.md](../../AGENTS.md)
- [源码矩阵总索引](source-matrix/README.md)

## 源码矩阵入口

源码矩阵按实际闭环顺序维护，作为所有实现切片的编码前门禁。当前已覆盖基础监控、链路监控、日志中心、Agent Suite、AI SRE / Evidence Chain、知识库 / 向量索引等矩阵。

- [源码矩阵总索引](source-matrix/README.md)
- [P0 源码矩阵锁定表](source-matrix/p0_source_matrix_lock.md)
- [P5 AI SRE / Evidence Chain 同源矩阵](source-matrix/p5_aisre_evidence_chain_matrix.md)
- [P6 知识库 / Qdrant 向量索引同源矩阵](source-matrix/p6_knowledge_qdrant_vector_matrix.md)

说明：长期主计划的实施阶段仍按平台建设顺序推进；源码矩阵文件名按当前文档闭环顺序落地，执行时以矩阵总索引和具体矩阵内容为准，不能只按编号猜测范围。

## 仍有效的专题文档

| 文档 | 用途 | 状态 |
| --- | --- | --- |
| [findx_monitoring_core_api_contract.md](findx_monitoring_core_api_contract.md) | 历史 API 契约和已有接口证据 | 参考，实施时需与新主计划重新校准 |
| [源码矩阵总索引](source-matrix/README.md) | 当前页面、API、Agent、AI SRE、知识库编码前门禁 | 有效 |
| [../compliance/third-party-sources.md](../compliance/third-party-sources.md) | 第三方来源、许可、品牌脱敏边界 | 有效 |
| [../testing/AI_WorkBench_统一测试基准.md](../testing/AI_WorkBench_统一测试基准.md) | 测试准出门禁 | 有效 |

## SkyWalking 前端链路补充

主计划已经明确 SkyWalking 不只是 OAP 后端 Adapter。链路监控前端必须按 `D:\平台源码\skywalking-booster-ui-main` 的 `src\router`、`src\views`、`src\store\modules`、`src\graphql\query` 和 `src\graphql\fragments` 做同源状态流迁移或等价实现。FindX 用户侧使用“链路监控”命名，内部保留服务目录、拓扑、Trace、Profiling、告警、接入、GraphQL 错误态、超时、取消请求和组件不可用 `BLOCKED`。

## SigNoZ 前端链路补充

日志中心必须按 `D:\平台源码\signoz-develop\frontend` 的 `src\AppRoutes\routes.ts`、`src\constants\routes.ts`、`src\AppRoutes\pageComponents.ts`、`src\components\Logs`、`src\components\LogDetail`、`src\components\QuickFilters`、`src\components\ExplorerCard`、`src\api\logs`、`src\api\pipeline`、`src\api\saveView`、`src\api\trace` 做同源状态流迁移或等价实现。FindX 用户侧使用“日志中心”命名，内部保留日志检索、字段筛选、上下文、聚合、live tail、Pipeline、Saved Views、Trace 关联和组件不可用 `BLOCKED`。

## SkyWalking Agent 补充

SkyWalking Agent 生态必须进入 FindX Agent 管理中心，而不是停留在清单或后端包管理。Java、Python、Node.js、PHP、Go、Rust、Ruby、Nginx Lua、Kong、Browser Client JS 等能力包必须覆盖包仓库、安装向导、配置模板、远程下发、心跳、数据到达、版本治理、漂移检测、升级回滚、卸载和 Evidence Chain。

P0 源码矩阵必须逐项登记 `skywalking-java`、`skywalking-python`、`skywalking-nodejs`、`skywalking-php`、`skywalking-go`、`skywalking-rust`、`skywalking-ruby`、`skywalking-nginx-lua`、`skywalking-kong`、`skywalking-client-js` 的上游仓库、本地源码路径、版本/commit、许可证、配置项、安装方式、数据到达验证和 FindX Agent 用户侧命名。独立 Agent 源码未落地时只能标记 `BLOCKED`，不能用静态清单代替实现。

## Superseded 文档

以下文档不再作为实施入口，仅保留为历史上下文：

- `docs/aiops/findx_flashcat_full_stack_observability_long_term_plan.md`
- `docs/aiops/nightingale_findx_agents_project_plan.md`
- `discuss/findx_full_implementation_plan.md`
- `discuss/findx_execution_task_breakdown.md`
- `discuss/findx_monitoring_execution_master_plan.md`
- `discuss/findx_agent_management_workbench_plan.md`
- `discuss/findx_global_page_structure_matrix.md`
- `discuss/findx_nightingale_one_to_one_next_slices.md`
- `discuss/findx_ui_navigation_one_to_one_audit.md`

归档索引见：

- [2026-05 FindX plan reset](../archive/2026-05-findx-plan-reset/README.md)

## 禁止事项摘要

- 禁止 iframe 嵌入参考站。
- 禁止嵌入参考站 SSO。
- 禁止 MVP、最小实现、最小验证。
- 禁止静态假按钮和错误语义映射。
- 禁止只看截图不看源码。
- 禁止用户侧出现外部品牌。
- 禁止未跑浏览器真实交互却写 PASS。
