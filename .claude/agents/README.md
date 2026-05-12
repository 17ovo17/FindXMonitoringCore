# .claude/agents/ — FindX 子 agent 角色定义

本目录是**主 agent 通过 Task 工具派发子 agent 时注入的角色上下文**，对应用户级 `~/.claude/CLAUDE.md` §子agent 角色体系。

## 角色清单

| 角色 | 文件 | 职责面 |
|---|---|---|
| Go 后端工程师 | `go-backend.md` | Gin handler、GORM、路由、Service 边界 |
| React 前端工程师 | `react-frontend.md` | react-shell JSX、api/*.js、d3 可视化 |
| QA 测试员 | `qa-tester.md` | 六维评分、BLOCKED_BY_CONTRACT、回归测试 |
| 文档维护员 | `doc-writer.md` | docs/**、operations-log.md、handoff 续写 |
| 运维诊断师 | `ops-diagnostician.md` | Prometheus/SkyWalking/systemd/catpaw/Loki |

## 派发约定

- 调用 Task 工具时**必须**将目标角色 md 文件的完整内容注入 prompt
- 同批并行子 agent **不得**修改同一文件
- 敏感信息一律使用占位符（`<API_KEY>` / `<TOKEN>` / `<DB_DSN>` / `<BASE_URL>` / `<SSH_KEY>`）
- 模型统一用 `opus`（本项目本轮约定，覆盖项目级 CLAUDE.md 中的 Codex 路线）

## 与 docs/ai/ 下业务专家的关系

`docs/ai/SysDiag-Inspector.agent.md` 与 `Topo-Architect.agent.md` 是**业务领域专家**角色（诊断报告结构、拓扑 JSON 规范），与本目录下的 **SDLC 执行层角色**正交，互不替代：

- SDLC 角色（本目录）：做编码、测试、文档、诊断工程实现
- 业务专家（docs/ai/）：输出格式规范、评分规则、领域约束

`ops-diagnostician.md` 在具体诊断实现时会引用这两份业务专家定义。

## Codex 路线

项目级 CLAUDE.md 提到 `.codex/agents/` 的 Codex 子代理体系，**本轮暂不启用**。所有子 agent 通过 Claude 的 Task 工具派发，不走 Codex CLI/MCP。
