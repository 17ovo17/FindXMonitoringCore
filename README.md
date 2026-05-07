# FindX

FindX 是一个面向可观测、Agent 管理、AI SRE 和 Evidence Chain 的长期平台项目。当前正式方向不是嵌入参考站，也不是继续维护弱化自研页面，而是基于成熟平台源码复刻页面结构、组件状态流和真实功能点，再统一替换为 FindX 自有品牌、风格、权限、审计、配置和数据源。

## 当前唯一长期开发计划

当前唯一实施入口：

- [FindX 全栈可观测长期开发计划](docs/aiops/findx_full_stack_observability_long_term_plan.md)

旧方向计划、讨论稿和阶段性审计材料只保留为历史证据或归档参考，不再作为开发实施依据。

## 开发前必须读

- [FindX 全栈可观测长期开发计划](docs/aiops/findx_full_stack_observability_long_term_plan.md)
- [AIOps 文档索引](docs/aiops/README.md)
- [统一测试基准](docs/testing/AI_WorkBench_统一测试基准.md)
- [第三方来源与融合登记](docs/compliance/third-party-sources.md)
- [AGENTS.md](AGENTS.md)

## 硬性执行规则

- 禁止 iframe 嵌入参考站，禁止嵌入参考站 SSO，FindX 必须使用自有登录、导航、权限和主题壳层。
- 禁止 MVP、最小实现、最小验证、占位页面、静态假按钮和未验证 PASS。
- 每个页面或功能切片编码前必须读取成熟源码和运行态 DOM，输出源码路径、路由、核心组件、API 调用、状态流、按钮真实动作、空态、错误态、权限态和 FindX 替换点。
- 用户侧只能出现 FindX / FindX Agent 等 FindX 品牌；外部来源名称只允许出现在内部开发证据、合规登记和归档文档中。
- API 测试不能替代 MCP 浏览器真实登录和点击回归。
- 组件不可用必须显示明确 `BLOCKED`，不能伪装成功。
- 不删除历史证据；旧文档先标记 superseded 或归档索引，截图、JSON、DOM snapshot 和测试产物先列候选清单。

## 后续闭环入口

文档收敛完成后，从 `P0-SOURCE-MATRIX-LOCK` 开始：

当前源码矩阵总入口：

- [源码矩阵总索引](docs/aiops/source-matrix/README.md)

已闭环的编码前门禁包括：

1. 基础监控页面/路由/API/状态流矩阵。
2. SkyWalking OAP/UI/Agent 仓库矩阵。
3. SigNoZ 日志中心矩阵。
4. AutoOps/AIOps CMDB 矩阵。
5. Categraf/Catpaw 到 FindX Agent 的能力矩阵。
6. AI SRE / Evidence Chain 矩阵。
7. 知识库 / Qdrant 向量索引矩阵。
8. 统一配置与数据源契约。
9. React-first FindX 自有壳层与导航。

任何实现切片都必须先通过源码证据门禁，再进入编码、WSL 构建、MCP 浏览器回归和敏感信息扫描。
