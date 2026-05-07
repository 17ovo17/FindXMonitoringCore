# 告警与通知运行态 DOM 证据

采集时间：2026-05-07 15:40（UTC+8）

参考页面：

```text
http://198.18.20.146:17000/alert-rules
http://198.18.20.146:17000/notification-channels
http://198.18.20.146:17000/event-pipelines
```

## 告警规则 DOM 摘要

```yaml
- side menu:
  - 告警:
    - 规则管理
    - 告警自愈
    - 告警事件
    - 工作流
- page tabs:
  - 告警规则
  - 屏蔽规则
  - 订阅规则
- left panel:
  - 预置筛选: 全部规则
  - 业务组
  - search: 请输入搜索关键字
- toolbar:
  - reload
  - datasource combobox
  - severity combobox
  - search: 搜索名称或标签
  - 新增
  - 导入
  - 更多操作
  - column visibility
- table:
  - columns: 状态, 类型, 名称, 更新时间, 用户名, 启用, 操作
  - row actions: 克隆, 删除
  - pagination: 共 12 条
```

## 通知媒介 DOM 摘要

```yaml
- side menu:
  - 通知:
    - 通知规则
    - 通知媒介
    - 消息模板
- page title: 通知媒介
- toolbar:
  - search: 请输入搜索关键字
  - 新增
  - 导入
  - 导出
- table:
  - columns: 名称, 发送类型, 更新人, 更新时间, 启用, 操作
  - rows:
    - Email / smtp / SMTP
    - PagerDuty / pagerduty
    - FlashDuty / flashduty
    - Callback / http
    - Dingtalk / http
    - FeishuApp / script
  - row actions: 克隆, 删除
  - pagination: 共 21 条
```

## 工作流 DOM 摘要

```yaml
- page title: 工作流
- toolbar:
  - search: 请输入搜索关键字
  - 新增
- table:
  - columns: 名称, 备注, 授权团队, 更新人, 更新时间, 操作
  - rows:
    - deepseek / clims业务组 / root / 编辑 / 删除
    - kimi / clims业务组 / root / 编辑 / 删除
    - 通义千问 / 海关技术中心 / clims业务组 / root / 编辑 / 删除
  - pagination: 共 3 条
```

## 验收含义

- 告警规则、屏蔽规则、订阅规则是同一规则管理组内的 Tabs，不得拆散或混入通知。
- 告警自愈、告警事件、工作流是告警域内独立入口；FindX 可把工作流入口并入 AI SRE，但页面结构必须沿用工作流列表、执行记录、编辑、调试。
- 通知规则、通知媒介、消息模板是通知域的三个独立页面，不得合并成一个复杂自研页面。
- 所有按钮必须接入真实动作；未完成时显示 `BLOCKED`，不能静态展示。
