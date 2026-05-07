# FindX P0 源码矩阵

本目录是 `P0-SOURCE-MATRIX-LOCK` 的证据入口。后续任何页面、API、Agent、日志、链路、CMDB、模板或工作流切片编码前，必须先在这里补齐成熟源码证据和运行态 DOM 证据。

当前矩阵：

- [P0 源码矩阵锁定表](p0_source_matrix_lock.md)
- [P0 导航同源矩阵](p0_basic_monitoring_nav_matrix.md)
- [P0 统一配置与数据源契约矩阵](p0_config_datasource_contract_matrix.md)
- [P0 React-first 壳层与导航迁移矩阵](p0_web_react_shell_nav_matrix.md)
- [P0 数据源页面同源矩阵](p0_datasource_real_matrix.md)
- [P0 指标查询页面同源矩阵](p0_metric_explorer_real_matrix.md)
- [P0 仪表盘页面同源矩阵](p0_dashboard_real_matrix.md)
- [P0 模板中心页面同源矩阵](p0_template_center_real_matrix.md)
- [P1 告警与通知页面同源矩阵](p1_alert_notification_real_matrix.md)
- [P1 人员组织与系统配置同源矩阵](p1_org_system_real_matrix.md)
- [P2 AutoOps CMDB 与 Agent 在线同源矩阵](p2_autoops_cmdb_agent_real_matrix.md)
- [P3 链路监控同源矩阵](p3_skywalking_apm_real_matrix.md)
- [P3 日志中心同源矩阵](p3_signoz_logs_real_matrix.md)
- [P4 Agent Suite 同源矩阵](p4_categraf_catpaw_agent_suite_matrix.md)
- [P5 AI SRE / Evidence Chain 同源矩阵](p5_aisre_evidence_chain_matrix.md)
- [P6 知识库 / Qdrant 向量索引同源矩阵](p6_knowledge_qdrant_vector_matrix.md)
- [P0 SkyWalking Agent 到 FindX Agent 控制面矩阵](p0_skywalking_agent_real_matrix.md)
- [SkyWalking Agent 本地源码检查证据](evidence/skywalking_agent_local_source_snapshot.md)
- [SigNoZ 日志中心源码检查证据](evidence/signoz_logs_source_snapshot.md)
- [Categraf / Catpaw 到 FindX Agent Suite 源码检查证据](evidence/categraf_catpaw_agent_suite_source_snapshot.md)
- [FindX AI SRE / Evidence Chain 源码检查证据](evidence/findx_aisre_evidence_chain_source_snapshot.md)
- [FindX 知识库 / Qdrant 向量索引源码检查证据](evidence/findx_knowledge_qdrant_source_snapshot.md)
- [运行态 DOM 证据](evidence/)

硬规则：

- 没有源码路径、路由、核心组件、API、状态流、按钮动作、空态、错误态、权限态，不允许编码。
- 没有运行态 DOM / MCP 浏览器证据，不允许把页面声明为一比一完成。
- 独立源码未落地时只能标记 `BLOCKED`，不能用静态清单代替实现。
- 用户侧命名必须使用 FindX / FindX Agent；外部来源名称只允许留在本目录、合规登记、归档和开发证据中。
