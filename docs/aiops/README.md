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

## 仍有效的专题文档

| 文档 | 用途 | 状态 |
| --- | --- | --- |
| [findx_monitoring_core_api_contract.md](findx_monitoring_core_api_contract.md) | 历史 API 契约和已有接口证据 | 参考，实施时需与新主计划重新校准 |
| [findx_nightingale_one_to_one_baseline_matrix.md](findx_nightingale_one_to_one_baseline_matrix.md) | 页面结构门禁历史矩阵 | 参考，后续需改写为 P0 源码矩阵 |
| [findx_page_structure_one_to_one_baseline.md](findx_page_structure_one_to_one_baseline.md) | 页面结构基线历史材料 | 参考，不能替代源码证据 |
| [../compliance/third-party-sources.md](../compliance/third-party-sources.md) | 第三方来源、许可、品牌脱敏边界 | 有效 |
| [../testing/AI_WorkBench_统一测试基准.md](../testing/AI_WorkBench_统一测试基准.md) | 测试准出门禁 | 有效 |

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
