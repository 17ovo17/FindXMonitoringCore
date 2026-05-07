# 模板中心运行态 DOM 证据

采集时间：2026-05-07 15:35（UTC+8）

参考页面：

```text
http://198.18.20.146:17000/components
```

详情页示例：

```text
http://198.18.20.146:17000/components?component=cAdvisor
```

## 列表关键 DOM 摘要

```yaml
- page title: 模板中心
- search:
  - textbox: 请输入搜索关键字
- action:
  - button: 创建
- grid:
  - cAdvisor
  - CloudWatch
  - Consul
  - Docker
  - Elasticsearch
  - Java
  - Kafka
  - Kubernetes
  - Linux
  - MySQL
  - Nginx
  - Prometheus
  - Redis
  - Windows
  - ClickHouse
```

## 抽屉关键 DOM 摘要

```yaml
- drawer title:
  - icon: component logo
  - name: cAdvisor
  - close: true
- tabs:
  - 采集说明
  - 指标说明
  - 仪表盘
  - 告警规则
- instructions:
  - heading: cadvisor
  - paragraph: cadvisor 采集插件说明
  - heading: Configuration
  - code block: TOML 配置示例
  - token example: sanitized as <TOKEN>
  - TLS example: ca/cert/key paths
  - url label explanation
- footer:
  - button: 编辑
```

## 验收含义

- 模板中心必须保留图标网格、搜索、创建、抽屉详情和 Tabs。
- 模板说明不是可省略的短文案，必须保留采集说明、配置示例、指标说明、仪表盘模板和告警规则模板。
- 说明中的敏感示例必须脱敏，例如 token 只能显示 `<TOKEN>`。
- 用户侧外部品牌和旧产品名必须替换为 FindX / FindX Agent / findx-agent，但模板含义、插件名称、配置字段和导入流程不能被删除。
