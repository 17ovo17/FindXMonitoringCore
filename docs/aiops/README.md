# FindX AIOps 文档索引

更新时间：2026-05-08 01:17（UTC+8）

本目录是 FindX 全量闭环的当前文档入口。开发、审查、测试、归档和 Git 门禁均以这里列出的文档为准。

## 当前唯一实施入口

- [FindX 全栈可观测长期开发计划](findx_full_stack_observability_long_term_plan.md)
- [FindX React-only 前端闭环计划](findx_react_only_frontend_long_term_plan.md)
- [源码矩阵索引](source-matrix/README.md)

`findx_react_first_frontend_long_term_plan.md` 已被 React-only 计划取代，仅保留为 superseded 历史文件。

## 开发前必须核对

- [AGENTS.md](../../AGENTS.md)
- [项目 README](../../README.md)
- [统一测试基准](../testing/AI_WorkBench_统一测试基准.md)
- [第三方来源与融合登记](../compliance/third-party-sources.md)
- [.claude/codex-task-board.md](../../.claude/codex-task-board.md)
- [.claude/operations-log.md](../../.claude/operations-log.md)

## 有效源码矩阵

- [P0 源码矩阵锁定表](source-matrix/p0_source_matrix_lock.md)
- [P0 导航同源矩阵](source-matrix/p0_basic_monitoring_nav_matrix.md)
- [P0 React-only Shell 与导航矩阵](source-matrix/p0_web_react_shell_nav_matrix.md)
- [P0 数据源同源矩阵](source-matrix/p0_datasource_real_matrix.md)
- [P0 指标查询同源矩阵](source-matrix/p0_metric_explorer_real_matrix.md)
- [P0 仪表盘同源矩阵](source-matrix/p0_dashboard_real_matrix.md)
- [P0 模板中心同源矩阵](source-matrix/p0_template_center_real_matrix.md)
- [P0 SkyWalking Agent 到 FindX Agent 控制面矩阵](source-matrix/p0_skywalking_agent_real_matrix.md)
- [P1 告警与通知同源矩阵](source-matrix/p1_alert_notification_real_matrix.md)
- [P1 组织与系统配置同源矩阵](source-matrix/p1_org_system_real_matrix.md)
- [P2 AutoOps CMDB 与 Agent 在线同源矩阵](source-matrix/p2_autoops_cmdb_agent_real_matrix.md)
- [P3 链路监控同源矩阵](source-matrix/p3_skywalking_apm_real_matrix.md)
- [P3 日志中心同源矩阵](source-matrix/p3_signoz_logs_real_matrix.md)
- [P4 Agent Suite 同源矩阵](source-matrix/p4_categraf_catpaw_agent_suite_matrix.md)
- [P5 AI SRE / Evidence Chain 同源矩阵](source-matrix/p5_aisre_evidence_chain_matrix.md)
- [P6 知识库 / Qdrant 向量索引同源矩阵](source-matrix/p6_knowledge_qdrant_vector_matrix.md)

## V2 硬规则摘要

- 前端最终架构为 React-only。
- Vue 只允许作为 `TEMP_BRIDGE`、`REPLACED` 或 `REMOVE_AFTER_REACT`，不能作为最终验收基线。
- 基础监控按 Nightingale React 源码迁移。
- 链路监控按 SkyWalking UI/OAP 源码迁移。
- SkyWalking Agent 是 FindX Agent 管理中心的一等能力，贯穿 P0 矩阵、P2 入口、P3 链路联动、P5 生命周期闭环。
- 日志中心按 SigNoZ 源码迁移。
- CMDB / Agent 在线按 AutoOps/AIOps 源码迁移。
- Categraf / Catpaw 作为 FindX Agent 能力来源，用户侧统一 FindX Agent 命名。
- 禁止 iframe、参考站 SSO、弱化自研页面、静态假按钮、静态假数据、最小实现、最小验证。
- API 测试不能替代 MCP/Playwright 真实浏览器回归。
- 未执行 WSL build、lint/测试、浏览器回归、敏感扫描、品牌扫描和静态假按钮扫描，不得写 PASS。

## 多 Agent 执行记录要求

每个任务必须在 `.claude/codex-task-board.md` 记录：

- `FX-NIGHT-*` 或 `FX-TASK-*` ID
- 状态：`READY`、`CLAIMED`、`IN_PROGRESS`、`DONE`、`FAIL`、`BLOCKED`、`NOT_RUN`、`RISK`
- agent nickname/id
- 写集和禁止写集
- 成熟源码证据
- 验证门禁
- 关闭状态和残留风险

每次派发、关闭、超时、回派、接管、验证、扫描和 P0/P1 风险处理，必须追加到 `.claude/operations-log.md`。

## Superseded 文档

以下文档不再作为实施入口，只保留历史上下文：

- `docs/aiops/findx_flashcat_full_stack_observability_long_term_plan.md`
- `docs/aiops/findx_react_first_frontend_long_term_plan.md`
- `docs/aiops/nightingale_findx_agents_project_plan.md`
- `discuss/findx_full_implementation_plan.md`
- `discuss/findx_execution_task_breakdown.md`
- `discuss/findx_monitoring_execution_master_plan.md`
- `discuss/findx_agent_management_workbench_plan.md`
- `discuss/findx_global_page_structure_matrix.md`
- `discuss/findx_nightingale_one_to_one_next_slices.md`
- `discuss/findx_ui_navigation_one_to_one_audit.md`

归档索引：[`docs/archive/2026-05-findx-plan-reset/README.md`](../archive/2026-05-findx-plan-reset/README.md)
