# 第三方来源与融合登记

生成时间：2026-05-04 07:40（UTC+8）

本文用于记录 FindX Monitoring Core 可能参考、复用或授权衍生的第三方来源。当前为文档闭环登记，不代表所有来源均已融合到业务代码；正式分发前必须以实际代码、文件清单、许可证文本和修改说明复核。

## 来源总表

| 来源 | 许可证/授权 | FindX 用途 | 当前状态 | 正式分发前待补 |
|------|-------------|-----------|---------|---------------|
| Nightingale | Apache-2.0 | 告警、Dashboard、通知、模板、权限、事件流水线、任务等设计参考和可融合来源 | 参考/可融合 | 具体版本、文件清单、修改说明、NOTICE、许可证文本 |
| Categraf | MIT | `findx-agents` 插件生态、采集插件和指标上报模型复用 | 可复用 | 插件来源版本、保留声明、修改说明、二进制分发说明 |
| Catpaw | 授权衍生 | 巡检、诊断、远程会话、结构化工具、自动修复执行能力的授权衍生 | 授权边界待随实现登记 | 授权记录、衍生文件清单、限制说明、内部保管位置 |

## Nightingale Apache-2.0 参考/可融合边界

可参考或可融合方向：

- 告警规则、事件、通知、静默、订阅、值班、团队权限等领域模型。
- Dashboard 和模板中心的交互与导入导出思路。
- 任务、流水线和审计相关成熟设计。

必须遵守：

- 保留 Apache-2.0 要求的许可证和 NOTICE。
- 记录实际融合文件、修改说明和来源版本。
- FindX 最终运行时主入口仍为 `/api/v1/monitor/*`、`/api/v1/findx-agents/*`，不得把 Nightingale 描述为运行时主依赖。

## Categraf MIT 插件复用边界

可复用方向：

- 插件采集生态和指标采集模式。
- Linux/Windows 主机指标、进程、端口、中间件等插件能力。
- 进入 `findx-agents` 发行版的插件适配层。

必须遵守：

- 保留 MIT 许可证文本和上游版权声明。
- 记录插件来源、版本、修改说明。
- 分发二进制或镜像时同步携带许可证材料。

## Catpaw 授权衍生边界

可衍生方向以授权记录为准，文档层只写公开边界：

- 主机巡检与诊断报告。
- 远程会话能力的受控衍生。
- 结构化工具调用能力。
- 自动修复执行能力的受控执行器。

必须遵守：

- 不公开真实授权材料，只使用 `<AUTH_RECORD>` 占位。
- 不把旧 Catpaw API 当作 FindX 新主入口。
- 不扩大到授权未覆盖的分发、商用、闭源、再授权或第三方交付场景。

## 版本登记模板

| 字段 | 示例占位 |
|------|---------|
| source_name | `Nightingale` / `Categraf` / `Catpaw` |
| source_url | `<BASE_URL>` |
| version_or_commit | `<VERSION_OR_COMMIT>` |
| license | `Apache-2.0` / `MIT` / `<AUTH_RECORD>` |
| imported_files | `<FILE_LIST>` |
| modified_files | `<FILE_LIST>` |
| modification_summary | `<SUMMARY>` |
| notice_required | `yes/no` |
| reviewer | `<LOGIN_USER>` |
| review_time | `YYYY-MM-DD HH:mm（UTC+8）` |
