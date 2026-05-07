# 仪表盘详情运行态 DOM 证据

采集时间：2026-05-07 15:30（UTC+8）

参考页面：

```text
http://198.18.20.146:17000/dashboards/4?datasource=1
```

MCP 浏览器跳转后实际页面：

```text
http://198.18.20.146:17000/dashboards/4?datasource=1&server=198.18.20.20
```

页面标题：

```text
clims Nginx Stub - 明析监控平台
```

## 关键 DOM 摘要

```yaml
- link "仪表盘列表":
  - /url: /dashboards
- title:
  - text: clims Nginx Stub
  - dropdown: true
- button "添加图表":
  - action: open add panel menu
- refresh controls:
  - button "sync"
  - button "Off down"
- time range:
  - button "最近 1 小时 down"
- dashboard actions:
  - button "setting"
  - button "link"
  - button "fullscreen"
- variables:
  - label: datasource
    combobox: true
    value: 普罗米修斯
  - label: server
    combobox: true
    value: 198.18.20.20
- panels:
  - name: Requests
    value: 3.266295M
  - name: Active connections
    value: "257"
  - name: Waiting connections
    value: "254"
  - name: Reading connections
    value: "1"
  - name: Writing connections
    value: "2"
  - name: handled
    value: 381.409k
```

## 验收含义

- 仪表盘详情页顶部不是静态标题，必须有返回列表、仪表盘下拉切换、添加图表、刷新、自动刷新、时间范围、设置、分享链接、全屏。
- `datasource`、`server` 等模板变量不是静态标签，必须是可选择、可搜索、可跟 URL 参数联动的变量控件。
- Panel 必须来自真实 dashboard configs 和查询渲染，不能用静态卡片冒充。
- 用户侧最终标题和导航必须替换为 FindX 品牌；本文件只作为内部运行态证据保存参考来源。
