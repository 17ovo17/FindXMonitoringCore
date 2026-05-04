# FindX 知识库记忆标签与 Prompt 注入策略计划

更新时间：2026-05-04 11:47（UTC+8）

## 1. 文档定位

本文档用于沉淀“知识库记忆标签 key + 分层分级 Prompt 注入策略”的讨论结论和实施入口，作为后续 `P1-KB` 切片的前置计划。

本能力属于知识库能力，不是 workflow 节点自身的普通记忆。workflow LLM node 后续只作为调用方之一接入统一 Prompt Assembler，不在节点内部各自维护记忆规则、标签策略或历史摘要逻辑。

本能力在 FindX Monitoring Core 中属于增强层。它服务于 AI 问诊、单机 Agent 对话、Runbook 推荐和自动修复计划生成，但不能替代 Nightingale 基础监控能力，也不能把“记忆”当成监控事实。用户偏好、场景激活和自定义标签可以影响表达方式、检索范围和工具选择；真正的诊断结论仍必须引用监控事件、Query Gateway 摘要、Target/Agent 证据、通知记录、巡检 artifact、Runbook 或知识库文档。

参考现状：

- `api/internal/model/knowledge.go`：已有案例库、Runbook 模型，尚无 memory tag 模型。
- `api/internal/store/knowledge_search.go`：已有知识库搜索事件、badcase 和统计，适合承接标签命中质量观测。
- `api/internal/handler/knowledge.go`、`api/internal/handler/documents.go`：已有案例、文档、搜索、重建索引接口。
- `api/internal/handler/chat.go`：当前 Chat 固定 system prompt，并把监控上下文追加到最后一条用户消息，尚无统一记忆注入层。
- `api/internal/workflow/node/llm.go`：当前 LLM node 只解析 `system_prompt`、`user_prompt`、`model`、`temperature`、`max_tokens`、`json_mode`，尚无知识库记忆注入能力。
- `web/src/views/KnowledgeCenter.vue`：知识中心已有案例库、Runbook、文档管理、诊断归档、语义搜索入口，适合作为 memory tag UI 的承载位置。

## 2. 核心结论

知识库记忆必须分层治理，不能把所有“记忆”默认塞进 Prompt。

三层 memory：

| 层级 | 名称 | 来源 | 注入方式 | 约束 |
| --- | --- | --- | --- | --- |
| LTM | Long-Term Memory，长期知识库记忆 | 用户偏好、场景激活、自定义标签及其绑定对象 | 由 Prompt Assembler 统一选择与压缩 | 需要权限、scope、token 预算和 citations |
| STM | Short-Term Memory，短期任务上下文 | 当前会话、当前告警、当前目标、当前 workflow run、当前 AIOps session 的临时状态 | 按调用场景传入，预算受控注入 | 不落入长期知识库，默认随会话或任务生命周期失效 |
| 历史聊天记录 | 历史 Chat/AIOps 对话 | `chat_sessions`、`chat_messages` 及后续语义化摘要 | 检索或摘要后注入，默认不全量 | 涉及隐私和敏感信息，必须显式授权或策略命中 |

LTM 必须再分三类：

| LTM 类型 | 定义 | 是否可全量注入 | 结论 |
| --- | --- | --- | --- |
| 用户偏好 | 用户长期稳定的输出风格、语言、格式、运维偏好、风险偏好 | 可以，但只能短摘要、结构化、预算受控 | 适合放入固定小预算，例如 200-400 tokens |
| 场景激活 | 特定场景自动激活的组织级或团队级规则，例如“告警诊断必须先列证据缺口” | 可以，但只能短摘要、结构化、预算受控 | 可按 scene key 全量注入该 scene 下的短摘要 |
| 自定义标签 | 用户或团队绑定到文档、Runbook、案例、目标、业务、指标或历史结论的标签 | 不可以默认全量注入 | 必须通过检索、显式选择或策略规则注入 |

可行性结论：

- 用户偏好和场景激活可以全量注入，但“全量”只允许短摘要、结构化、预算受控的全量，不允许把原始长文本、历史对话、文档正文无差别注入。
- 自定义标签必须检索注入、显式注入或策略注入，不能默认全量注入。原因是自定义标签可能数量大、跨权限域、语义松散、包含敏感片段，默认全量会放大 token 成本、隐私风险和 prompt 注入风险。
- Prompt 注入必须集中在知识库 Prompt Assembler，禁止 Chat、AIOps、workflow LLM node、后续 evidence chain 各自拼接一套不一致的记忆块。

## 3. 目标与非目标

目标：

- 为知识库增加可治理的 memory tag key 体系。
- 提供统一 Prompt Assembler，负责输入上下文、记忆标签、检索证据、token 预算、citations 和 debug/preview。
- 让 Chat、AIOps、workflow LLM node、FindX Monitoring Core 和后续 AI evidence chain 以同一策略接入。
- 为 Knowledge Center 提供标签管理、绑定关系、激活预览和注入预览 UI 的接口草案。

非目标：

- 不在本计划中实现代码。
- 不把 workflow LLM node 改造成独立记忆容器。
- 不默认把历史聊天记录全部注入 Prompt。
- 不把自定义标签当作全局 system prompt。
- 不新增真实密钥、Cookie、DSN、连接串、会话 ID 示例。

## 4. 数据模型草案

### 4.1 `knowledge_memory_tags`

用途：定义知识库记忆标签 key、类型、作用域、摘要、注入策略和安全属性。

建议字段：

| 字段 | 类型建议 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `VARCHAR(64)` | 是 | 标签记录 ID，使用项目统一 ID 生成策略 |
| `tag_key` | `VARCHAR(128)` | 是 | 稳定 key，例如 `user.pref.output_style`、`scene.aiops.alert_diagnosis`、`custom.jvm.gc` |
| `tag_type` | `VARCHAR(32)` | 是 | `user_preference`、`scene_activation`、`custom` |
| `scope_type` | `VARCHAR(32)` | 是 | `global`、`team`、`user`、`business`、`target`、`workflow`、`session` |
| `scope_id` | `VARCHAR(128)` | 否 | 作用域对象 ID；`global` 可为空 |
| `title` | `VARCHAR(255)` | 是 | 标签展示名 |
| `summary` | `TEXT` | 是 | 可注入短摘要，只允许结构化、短文本、脱敏内容 |
| `content_ref` | `JSON` | 否 | 原始来源引用，例如文档、Runbook、案例、历史摘要 ID；不直接存放敏感正文 |
| `activation_policy` | `JSON` | 否 | 激活规则，例如 scene、route、intent、target selector、query matcher、priority |
| `injection_policy` | `JSON` | 是 | 注入策略，例如 `always_summary`、`retrieval_only`、`explicit_only`、`policy_match`、预算上限 |
| `token_budget` | `INT` | 是 | 单标签最大注入预算，默认按 tag_type 给出上限 |
| `priority` | `INT` | 是 | 同预算竞争时的优先级，数值越大越优先 |
| `sensitivity` | `VARCHAR(32)` | 是 | `public`、`internal`、`confidential`、`restricted` |
| `enabled` | `TINYINT(1)` | 是 | 是否启用 |
| `created_by` | `VARCHAR(128)` | 是 | 创建人，使用登录态用户标识，不写真实账号示例 |
| `updated_by` | `VARCHAR(128)` | 否 | 最近更新人 |
| `created_at` | `DATETIME` | 是 | 创建时间 |
| `updated_at` | `DATETIME` | 是 | 更新时间 |

约束建议：

- `tag_key` 在 `scope_type + scope_id` 下唯一。
- `tag_type=user_preference` 默认只允许 `scope_type=user/team/global`。
- `tag_type=scene_activation` 必须有 `activation_policy.scene`。
- `tag_type=custom` 默认 `injection_policy.mode` 不得为 `always_summary`，除非后续加入显式白名单和管理员审批。

### 4.2 `knowledge_tag_bindings`

用途：把 memory tag 绑定到知识库对象、监控对象、会话摘要或后续 evidence 对象。

建议字段：

| 字段 | 类型建议 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `VARCHAR(64)` | 是 | 绑定记录 ID |
| `tag_id` | `VARCHAR(64)` | 是 | 关联 `knowledge_memory_tags.id` |
| `tag_key` | `VARCHAR(128)` | 是 | 冗余稳定 key，便于查询和审计 |
| `object_type` | `VARCHAR(32)` | 是 | `knowledge_document`、`diagnosis_case`、`runbook`、`monitor_target`、`business`、`metric`、`chat_summary`、`evidence_ref` |
| `object_id` | `VARCHAR(128)` | 是 | 绑定对象 ID |
| `object_title` | `VARCHAR(255)` | 否 | 绑定对象快照标题，便于预览 |
| `binding_role` | `VARCHAR(32)` | 是 | `source`、`citation`、`activation_hint`、`preference_source`、`evidence` |
| `weight` | `DOUBLE` | 是 | 绑定权重，默认 `1.0` |
| `snippet` | `TEXT` | 否 | 可注入短片段或摘要，必须脱敏，不能保存原始密钥和隐私全文 |
| `metadata` | `JSON` | 否 | 扩展信息，例如 chunk index、score、业务域、环境 |
| `created_by` | `VARCHAR(128)` | 是 | 创建人 |
| `created_at` | `DATETIME` | 是 | 创建时间 |

约束建议：

- `tag_id + object_type + object_id + binding_role` 唯一，避免重复绑定。
- 删除标签时应软删除或级联禁用绑定，避免 Prompt Assembler 读到悬挂引用。
- `object_type=chat_summary` 只能绑定语义化摘要，不绑定原始完整聊天记录。

## 5. Prompt Assembler 设计

Prompt Assembler 是知识库 memory 注入的唯一策略中心。调用方传入任务上下文，Assembler 输出可直接交给 LLM 的 messages 或 assembled prompt，同时返回 citations、预算消耗和调试信息。

### 5.1 输入

建议输入结构：

```json
{
  "scene": "chat|aiops|monitor_event|workflow_llm|evidence_chain|assemble_preview",
  "user_id": "<LOGIN_USER>",
  "team_id": "team-placeholder",
  "session_id": "session-placeholder",
  "target": {
    "target_id": "target-placeholder",
    "target_ident": "target-placeholder",
    "business_id": "business-placeholder"
  },
  "query": "用户问题或任务目标",
  "base_messages": [
    { "role": "system", "content": "基础系统提示词" },
    { "role": "user", "content": "用户输入" }
  ],
  "explicit_tag_keys": ["custom.jvm.gc"],
  "retrieval": {
    "enabled": true,
    "top_k": 5,
    "doc_type": "",
    "category": ""
  },
  "budget": {
    "total_tokens": 12000,
    "memory_tokens": 1800,
    "ltm_tokens": 900,
    "stm_tokens": 600,
    "history_tokens": 300
  },
  "debug": false
}
```

说明：

- `scene` 是场景激活的主 key。
- `explicit_tag_keys` 只表示用户或调用方显式选择，不等于跳过权限和敏感信息检查。
- `base_messages` 允许 Chat、AIOps、workflow LLM node 复用既有 prompt，但最终拼接顺序由 Assembler 决定。
- `session_id` 示例只使用占位符，不记录真实会话 ID。

### 5.2 输出

建议输出结构：

```json
{
  "messages": [
    { "role": "system", "content": "基础系统提示词 + 受控记忆摘要" },
    { "role": "user", "content": "用户输入 + 检索证据摘要" }
  ],
  "memory_blocks": [
    {
      "block_type": "ltm_user_preference",
      "tag_key": "user.pref.output_style",
      "content": "使用中文 Markdown，先结论后证据。",
      "tokens_estimated": 32,
      "citations": []
    }
  ],
  "citations": [
    {
      "ref_id": "kb-doc-placeholder",
      "source_type": "knowledge_document",
      "source_id": "doc-placeholder",
      "title": "文档标题",
      "snippet": "脱敏摘要片段",
      "score": 0.82
    }
  ],
  "budget_used": {
    "total_tokens_estimated": 4200,
    "memory_tokens_estimated": 860,
    "truncated": false
  },
  "debug": {
    "activated_tag_keys": ["scene.aiops.alert_diagnosis"],
    "skipped_tag_keys": [
      { "tag_key": "custom.secret.example", "reason": "sensitivity_restricted" }
    ],
    "policy_trace": ["scene_match", "permission_pass", "budget_pass"]
  }
}
```

### 5.3 Token 预算

建议默认预算：

| 块 | 默认上限 | 说明 |
| --- | --- | --- |
| LTM 用户偏好 | 200-400 tokens | 可全量短摘要注入 |
| LTM 场景激活 | 300-600 tokens | 可全量短摘要注入 |
| LTM 自定义标签 | 0 tokens 默认值 | 仅检索、显式或策略命中后占用预算 |
| STM 当前任务上下文 | 400-1200 tokens | 告警、目标、指标摘要、workflow run 状态 |
| 历史聊天记录 | 0-500 tokens | 默认不注入原文，只注入语义化摘要 |
| 检索证据 | 800-2000 tokens | 必须带 citations |

预算裁剪顺序：

1. 保留基础 system prompt 和当前用户问题。
2. 保留安全规则、权限规则和 evidence 约束。
3. 保留用户偏好短摘要。
4. 保留场景激活短摘要。
5. 按分数、显式选择、优先级保留自定义标签。
6. 历史聊天只保留摘要，不保留原始长对话。
7. 超预算时返回 `truncated=true` 和 debug/preview 原因。

### 5.4 注入策略

| 策略 | 适用对象 | 行为 |
| --- | --- | --- |
| `always_summary` | 用户偏好、场景激活 | 在权限通过和预算允许时注入短摘要 |
| `retrieval_only` | 自定义标签、知识文档、Runbook、案例 | 通过 query、scene、target、category 检索命中后注入 |
| `explicit_only` | 高风险自定义标签、history 摘要 | 只有用户或调用方显式选择才注入 |
| `policy_match` | 场景策略、自定义标签 | 满足 `activation_policy` 后注入 |
| `deny` | 禁用、越权、敏感、超预算对象 | 不注入，仅在 debug 中说明原因 |

注入顺序建议：

1. 安全和角色约束。
2. 用户偏好短摘要。
3. 场景激活短摘要。
4. STM 当前任务上下文。
5. 检索知识和自定义标签。
6. 历史聊天语义摘要。
7. citations 和证据缺口提示。

### 5.5 Citations

所有来自知识库、Runbook、案例、历史摘要、FindX evidence 的注入内容都必须产生 citations。

citations 至少包含：

- `ref_id`：本次 assembled prompt 内部引用 ID。
- `source_type`：来源类型。
- `source_id`：来源对象 ID。
- `title`：脱敏标题。
- `snippet`：短摘要，不包含真实密钥、Cookie、完整连接串或隐私全文。
- `score`：检索或策略命中分数。
- `tag_key`：如来自标签绑定则记录。

LLM 输出规范后续应要求引用 `ref_id`，避免脱离证据链生成诊断或修复建议。

### 5.6 Debug 与 Preview

`assemble-preview` 必须先于真实接入落地，用于验证策略而不调用上游 LLM。

Preview 内容：

- 最终 messages 预览。
- 每个 memory block 来源、tag key、预算、注入原因。
- 被跳过标签及原因。
- 检索结果和 citations。
- 敏感信息扫描结果。
- token 估算和裁剪说明。

Debug 限制：

- debug/preview 不得返回真实 `<TOKEN>`、Cookie、完整 `<DB_DSN>`、SSH 私钥、上游 AI provider key。
- restricted 内容只显示跳过原因，不显示原文。

## 6. API 草案

以下接口为规划草案，当前不得写成已实现或 QA PASS。统一前缀建议为 `/api/v1/knowledge`，沿用知识库能力域。

### 6.1 Memory Tags

| 方法 | 路径 | 用途 | 权限建议 |
| --- | --- | --- | --- |
| `GET` | `/api/v1/knowledge/memory-tags` | 标签列表，支持 `tag_type`、`scope_type`、`scope_id`、`enabled`、`keyword` | 登录态 |
| `POST` | `/api/v1/knowledge/memory-tags` | 创建标签 | 管理员或团队管理员 |
| `GET` | `/api/v1/knowledge/memory-tags/:id` | 标签详情 | 登录态 + scope 权限 |
| `PUT` | `/api/v1/knowledge/memory-tags/:id` | 更新标签 | 管理员或所有者 |
| `DELETE` | `/api/v1/knowledge/memory-tags/:id` | 禁用或删除标签 | 管理员或所有者 |

创建/更新请求重点字段：

```json
{
  "tag_key": "scene.aiops.alert_diagnosis",
  "tag_type": "scene_activation",
  "scope_type": "team",
  "scope_id": "team-placeholder",
  "title": "AIOps 告警诊断场景",
  "summary": "先列证据缺口，再给根因候选和下一步验证。",
  "activation_policy": { "scene": ["aiops", "monitor_event"], "priority": 80 },
  "injection_policy": { "mode": "always_summary", "max_tokens": 300 },
  "token_budget": 300,
  "priority": 80,
  "sensitivity": "internal",
  "enabled": true
}
```

### 6.2 Bindings

| 方法 | 路径 | 用途 | 权限建议 |
| --- | --- | --- | --- |
| `GET` | `/api/v1/knowledge/memory-tags/:id/bindings` | 查询某标签绑定对象 | 登录态 + scope 权限 |
| `POST` | `/api/v1/knowledge/memory-tags/:id/bindings` | 新增绑定 | 管理员或所有者 |
| `DELETE` | `/api/v1/knowledge/memory-tags/:id/bindings/:binding_id` | 删除绑定 | 管理员或所有者 |
| `GET` | `/api/v1/knowledge/bindings` | 按对象反查标签绑定 | 登录态 |

绑定请求示例：

```json
{
  "object_type": "runbook",
  "object_id": "runbook-placeholder",
  "object_title": "JVM GC 异常处置",
  "binding_role": "citation",
  "weight": 1.0,
  "snippet": "JVM GC 异常时先确认堆使用率、Full GC 次数和最近变更。",
  "metadata": { "category": "jvm", "chunk_index": 0 }
}
```

### 6.3 Active Tags

| 方法 | 路径 | 用途 | 权限建议 |
| --- | --- | --- | --- |
| `POST` | `/api/v1/knowledge/memory-tags/active` | 根据 scene、用户、目标、query 计算会激活的标签 | 登录态 |

请求示例：

```json
{
  "scene": "aiops",
  "user_id": "<LOGIN_USER>",
  "team_id": "team-placeholder",
  "target_id": "target-placeholder",
  "query": "分析 JVM Full GC 告警",
  "explicit_tag_keys": ["custom.jvm.gc"]
}
```

响应必须区分：

- `always_injected`：用户偏好、场景激活短摘要。
- `retrieval_candidates`：自定义标签候选。
- `explicit_selected`：显式选择标签。
- `skipped`：越权、禁用、敏感、超预算、策略不匹配。

### 6.4 Assemble Preview

| 方法 | 路径 | 用途 | 权限建议 |
| --- | --- | --- | --- |
| `POST` | `/api/v1/knowledge/assemble-preview` | 返回 Assembler 结果预览，不调用上游 LLM | 登录态 |

请求与输出参考第 5 节。

门禁要求：

- 先落地 `assemble-preview`，再接入 Chat、AIOps 和 workflow LLM node。
- preview 必须覆盖正常路径、越权路径、敏感标签跳过、超预算裁剪、自定义标签非默认全量注入。

## 7. 接入顺序

推荐按以下顺序进入 `P1-KB` 切片：

1. 策略骨架：定义 tag_type、scope、injection_policy、预算、citations、debug 数据结构和权限模型。
2. `assemble-preview`：先实现只预览、不调用 LLM 的 Assembler 验证入口。
3. Chat：把 `api/internal/handler/chat.go` 的固定 system prompt + 监控上下文拼接改为调用 Prompt Assembler。
4. AIOps：把 AIOps 诊断、巡检报告、实时问诊接入 scene 激活和 evidence citations。
5. workflow LLM node：`api/internal/workflow/node/llm.go` 只作为调用方接入 Assembler，不在 node 内实现私有记忆。
6. Knowledge Center UI：在 `KnowledgeCenter.vue` 或其子组件中增加 memory tag 管理、绑定管理、active tags 和 assemble preview。
7. history 语义化：对 `chat_messages` 做摘要、脱敏、标签化，再作为 `chat_summary` 绑定，不注入原始完整历史聊天。

## 8. 与 FindX Monitoring Core / P2 Agent / P3 AI Evidence Chain 的关系

### 8.1 FindX Monitoring Core

FindX Monitoring Core 的主线是监控事件、告警规则、Query Gateway、current/history event、审计和后续修复闭环。知识库 memory tag 不替代监控事件模型，也不替代告警证据。

关系：

- Monitoring Core 产生的事件、规则、目标、查询摘要可作为 STM 或 citations 来源。
- 场景激活可按 `monitor_event`、`alert_diagnosis`、`target_inspection` 等 scene 生效。
- AI 对监控事件的结论必须引用事件、Query Gateway 摘要、Target/Agent 证据、Runbook/Knowledge 引用，而不是只引用记忆标签。

### 8.2 P2 Agent

P2 Agent 侧重点是 findx-agents 协议、能力目录、session/evidence、配置分发、巡检与工具注册。memory tag 不直接下发为 Agent 执行策略。

关系：

- Agent 证据可以成为 Assembler 的 STM 或 citation。
- Agent capability、target labels、inspection artifact 可作为 activation_policy 的输入。
- 不允许把敏感 Agent 配置、token、完整主机私密信息写入 memory tag summary。

### 8.3 P3 AI Evidence Chain

P3 AI evidence chain 要求 AI 结论绑定 evidence refs。Prompt Assembler 是 evidence chain 的上游准备层之一。

关系：

- Assembler 输出 `citations`，后续可映射为 evidence refs。
- memory tag 只能提供偏好、场景规则、知识摘要和检索证据，不能替代真实证据。
- remediation plan/precheck/dry-run/approve/execute/verify/rollback 前，必须能追踪每条建议来自哪些 citations 或明确标记证据缺口。

## 9. 风险与门禁

| 风险 | 说明 | 门禁 |
| --- | --- | --- |
| 权限越界 | 标签可能跨团队、用户、目标、业务域 | Assembler 必须按 scope 权限过滤，preview 展示 skipped reason |
| 敏感信息 | 标签摘要、绑定 snippet、历史摘要可能含 token、Cookie、DSN、私钥、隐私 | 写入、预览、注入前做敏感信息扫描和脱敏；示例只使用 `<TOKEN>`、`<DB_DSN>` 等占位符 |
| history 隐私 | 历史聊天记录可能包含个人输入、故障细节、内部路径 | 默认不全量注入，必须先语义化摘要、脱敏、授权和可删除 |
| prompt 注入分散 | Chat、AIOps、workflow、evidence chain 各自拼 prompt 会失控 | 统一 Prompt Assembler，调用方不得私自拼 LTM |
| `file://prompts` 解析疑点 | workflow builtin YAML 大量使用 `file://prompts/*.txt`，但 LLM node 当前只做变量插值，需确认是否已有上游解析 | P1-KB 接入 workflow 前必须审计 file prompt 加载链路，未确认前不得假设 file prompt 已展开 |
| token 预算 | LTM、STM、history、检索证据叠加容易超预算 | 每类预算上限、裁剪顺序、`truncated=true` 和 preview 必须可见 |
| 自定义标签滥用 | 自定义标签数量大、语义杂、可能被用户当作隐形 system prompt | 默认不全量，必须检索、显式或策略命中；高敏标签只允许 explicit |
| citations 缺失 | LLM 输出无法追溯来源 | 所有知识库和历史摘要注入必须带 citations |
| 误把记忆当事实 | 用户偏好或场景规则不是事实证据 | 输出时区分“偏好/策略”和“证据/事实” |
| 数据兼容 | 新增表涉及 MySQL + 内存 fallback | 后续实现必须同步 MySQL DDL、内存 fallback、快照持久化和迁移策略 |

## 10. 验收建议

文档切片验收：

- 本文档只写入 `D:\ai-workbench\discuss\knowledge_memory_tag_injection_plan.md`。
- 不修改 `api/`、`web/`、`docs/`、`README.md`、`.claude/`、`.codex/`。
- 不执行 git stage/commit/push。
- 敏感信息检查：不包含真实密钥、Cookie、DSN、连接串、会话 ID。

后续代码切片验收：

- 后端：同步到 `/opt/ai-workbench/api` 后执行 `go test ./...` 和 `go build -o api-linux .`。
- 前端：同步到 `/opt/ai-workbench/web` 后执行 `npm run build`。
- API：`assemble-preview` 覆盖正常、越权、敏感跳过、预算裁剪、自定义标签非默认全量。
- UI：Knowledge Center 覆盖标签列表、创建/更新、绑定、active tags、assemble preview、错误态和权限态。
- QA：必须引用统一测试基准，新增 P1-KB 临时用例编号，例如 `QA-KB-MEM-001`。

## 11. 自评分与补齐

初始自评分：94/100。

扣分点：

- `file://prompts` 的实际解析链路仍需代码深挖确认，当前只能标记为疑点，不能写成事实。
- 历史聊天语义化的具体摘要算法、保留周期、删除策略仍待后续设计。
- 权限模型字段已给出 scope 草案，但尚未绑定现有用户、团队、管理员权限实现细节。

补齐动作：

- 已在风险门禁中单独列出 `file://prompts` 解析疑点，并要求 workflow 接入前必须审计。
- 已明确历史聊天记录默认不全量注入，必须先语义化摘要、脱敏、授权和可删除。
- 已在数据模型、API 草案和 Prompt Assembler 输入中加入 `scope_type`、`scope_id`、`user_id`、`team_id`、权限过滤和 skipped reason。
- 已明确自定义标签必须检索、显式或策略注入，不能默认全量。
- 已明确 citations、debug/preview、token 预算、敏感信息和 history 隐私门禁。

最终自评分：96/100。

### 11.1 与 FindX/Nightingale 深度集成后的复评分

复评分：**97/100**。

加分点：

- 已明确 memory 属于知识库能力，而不是 workflow 节点私有记忆。
- 已明确 LTM、STM、历史聊天记录三层，其中 LTM 拆为用户偏好、场景激活、自定义标签。
- 已明确用户偏好和场景激活只能做短摘要、结构化、预算受控的全量注入；自定义标签走检索、显式选择或策略注入。
- 已补充与 FindX Monitoring Core 的边界：记忆标签是增强层，不能替代 Nightingale 的监控事实、告警事实、事件流水线和通知记录。
- 已明确 AI 输出必须区分“偏好/场景规则”和“事实证据”，避免把记忆当作诊断证据。

扣分点：

- Prompt Assembler、memory schema、注入 preview、token budget、审计 API 尚未实现。
- 与 Agent/CMDB 单机对话、evidence chain、remediation plan 的运行态联动仍待后续切片验证。
- 需要后续补充标签 key 的 UI 管理页、批量导入导出、版本回滚和权限矩阵。

补齐措施：

- P3-AI-0 前先实现知识库记忆标签 API 契约、表结构、权限和注入预览。
- P3-AI-1 evidence chain 实现时，所有 AI 诊断必须同时输出 `evidence_refs`、`memory_refs` 和 `missing_evidence`，并在 UI 上分区展示。
- 单机 Agent 对话中，记忆标签只参与 Prompt Assembler，不得直接生成执行动作；执行动作必须进入 remediation 审批链。

建议结论：通过文档切片，进入 `P1-KB` 方案评审；代码实现前必须先落地 `assemble-preview` 和权限/脱敏门禁，再接入 Chat、AIOps、workflow LLM node 与 Knowledge Center UI。

## 12. 脱敏检查

| 检查项 | 结果 |
| --- | --- |
| 真实密钥、认证票据、Cookie | 未写入 |
| 完整 DSN、连接串、SSH 私钥 | 未写入 |
| 真实会话 ID | 未写入 |
| 示例敏感值 | 统一使用 `<TOKEN>`、`<DB_DSN>`、`<LOGIN_USER>` 等占位符 |
| debug/preview 约束 | 已要求不返回 restricted 原文和真实敏感值 |

## 13. 遗留风险

- 需要由后续 `go-backend` 切片确认 MySQL + 内存 fallback 的表结构、迁移、快照和权限实现。
- 需要由后续 `vue-frontend` 切片确认 Knowledge Center UI 的交互边界，避免把标签管理做成静态假数据。
- 需要由后续 `qa-tester` 对文档和实现分别补充 `QA-KB-MEM-*` 用例。
- 需要在 workflow 接入前确认 `file://prompts` 到真实 prompt 文本的解析位置和失败降级行为。
- 需要在 history 语义化前制定历史摘要保留、删除、脱敏、用户可见和撤销绑定策略。
