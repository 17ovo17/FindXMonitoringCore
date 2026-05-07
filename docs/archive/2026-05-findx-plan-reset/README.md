# 2026-05 FindX Plan Reset 归档索引

生成时间：2026-05-07 09:09（UTC+8）

本目录记录 2026-05 FindX 长期计划重置时被标记为 superseded 的旧方向文档和待清理候选。归档策略是先标记、先索引、先保留证据，不直接删除历史材料。

## Superseded 文档

以下文档已不再作为实施入口：

- `docs/aiops/findx_flashcat_full_stack_observability_long_term_plan.md`
- `docs/aiops/nightingale_findx_agents_project_plan.md`
- `discuss/findx_full_implementation_plan.md`
- `discuss/findx_execution_task_breakdown.md`
- `discuss/findx_monitoring_execution_master_plan.md`
- `discuss/findx_agent_management_workbench_plan.md`
- `discuss/findx_global_page_structure_matrix.md`
- `discuss/findx_nightingale_one_to_one_next_slices.md`
- `discuss/findx_ui_navigation_one_to_one_audit.md`

当前唯一实施入口：

- `docs/aiops/findx_full_stack_observability_long_term_plan.md`

## 根目录临时产物候选清单

以下文件只列为候选，不在本轮自动删除：

- `400`
- `tmp_unicode_test.txt`
- `*-snapshot.md`
- `*-regression.json`
- `reference-*.json`
- `findx-*.json`
- `n9e-microfrontend-regression.json`
- `nav-regression-runtime.json`
- 根目录历史截图、浏览器回归截图、DOM snapshot 和测试报告

处理规则：

1. 先检查 README、docs、discuss、测试报告和代码引用关系。
2. 有引用则保留或更新引用。
3. 无法确认用途则保留。
4. 确认无用后单独提交清理，不与长期计划提交混在一起。
