# Catpaw 授权衍生边界

生成时间：2026-05-04 07:40（UTC+8）

本文定义 Catpaw 能力衍生到 `findx-agents` 与 FindX Monitoring Core 时的公开边界。真实授权文件、授权联系人、密钥、私有仓库地址、Cookie、完整连接串和平台内网信息不得写入本文；需要引用时使用 `<AUTH_RECORD>`、`<TOKEN>`、`<BASE_URL>` 等占位符。

## 定位

Catpaw 是历史探针和远程诊断能力来源之一。FindX Monitoring Core 的新主线是 `/api/v1/monitor/*` 与 `/api/v1/findx-agents/*`，Catpaw 旧入口 `/api/v1/catpaw/*` 只做兼容回归。Catpaw 授权能力可以在授权范围内衍生到 `findx-agents`，但不得把旧 Catpaw 直接包装成 FindX 主线。

## 可衍生能力

| 能力 | 可衍生方向 | FindX 归属 |
|------|-----------|-----------|
| 巡检采集 | 主机、进程、端口、systemd、磁盘、网络、安全基线等巡检证据 | `findx-agents` evidence collector |
| 诊断报告 | 把巡检结果转为结构化 evidence refs | AI evidence chain |
| 远程会话 | 在权限、审计、命令分级和确认边界内提供受控会话 | 后续 remediation/tool runtime |
| 结构化工具 | 白名单工具、参数 schema、只读查询、有限写操作 | FindX Agent tool whitelist |
| 自动修复执行 | precheck、dry-run、approve、execute、verify、rollback | `/api/v1/remediation/*` 规划待实现 |

## 禁止扩大边界

- 未经授权，不得把 Catpaw 能力、二进制或源码作为独立产品重新分发。
- 不得在文档中公开真实授权记录、私有仓库地址、密钥、Cookie、完整 `<DB_DSN>` 或 SSH 私钥。
- 不得绕过 FindX 的权限、审计、命令分级和审批流程执行 raw command。
- 不得把 `/api/v1/catpaw/*` 描述为 FindX 新主线；新功能必须走 `/api/v1/findx-agents/*` 或未来 `/api/v1/remediation/*`。
- 不得把规划中的自动修复写成已实现；接口未注册、QA 未 PASS 前只能标记为规划待实现、NOT_RUN 或 BLOCKED。

## 授权记录要求

正式分发或交付前，主代理或合规负责人必须在受控位置维护：

| 项 | 要求 |
|----|------|
| 授权证明 | `<AUTH_RECORD>`，不进入公开仓库 |
| 授权范围 | 功能范围、分发范围、期限、限制 |
| 衍生清单 | 进入 `findx-agents` 的文件、模块、二进制、配置 |
| 修改说明 | 与原 Catpaw 能力相比的改动摘要 |
| 风险复核 | 权限、审计、命令执行、敏感信息、回滚能力 |

## QA 验收口径

- 文档-only 用例：`QA-DOC-FINDX-008`。
- Agent 生命周期：`FINDX-P2-AGENT-*`。
- AI 证据链：`FINDX-P3-AI-*`。
- 自动修复规划：`FINDX-P3-REMED-*`，未实现前不得写 PASS。
