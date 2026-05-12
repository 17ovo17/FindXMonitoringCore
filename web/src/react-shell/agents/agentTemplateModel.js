import { pluginConfigSpec } from './agentPluginConfigModel.js'

const KUBERNETES_BLOCKED_NOTICE = '# BLOCKED_BY_CONTRACT: Kubernetes 执行器、集群凭据、包签名、回执和数据到达验证未开放；以下内容仅用于复制预览，不会触发真实集群变更。'

const kubernetesPreview = lines => [KUBERNETES_BLOCKED_NOTICE, ...lines].join('\n')

export const installCommands = [
  {
    id: 'linux-curl',
    label: 'Linux 本机安装',
    desc: '使用 curl 下载安装脚本并传入一次性 token。真实 URL、包签名和审计回执必须由 Agent Adapter 生成。',
    command: "curl -kfsSL '<FINDX_URL>/api/v1/findx-agents/installers/linux.sh?package=<PACKAGE_ID>' -o /tmp/findx-agent-install.sh && FINDX_BASE_URL='<FINDX_URL>' FINDX_TOKEN='<TOKEN>' sh /tmp/findx-agent-install.sh",
  },
  {
    id: 'windows-cmd',
    label: 'Windows CMD 本机安装',
    desc: '使用 certutil 下载批处理脚本并传入一次性 token。脚本必须只安装 FindX Agent 内置包。',
    command: 'certutil -urlcache -f "<FINDX_URL>/api/v1/findx-agents/installers/windows.bat?package=<PACKAGE_ID>" "%TEMP%\\findx-agent-install.bat" && set "FINDX_BASE_URL=<FINDX_URL>" && set "FINDX_TOKEN=<TOKEN>" && "%TEMP%\\findx-agent-install.bat"',
  },
  {
    id: 'windows-powershell',
    label: 'Windows PowerShell 本机安装',
    desc: '使用 PowerShell 下载脚本并传入一次性 token。真实执行需要记录审计和回滚计划。',
    command: "Invoke-WebRequest -Uri '<FINDX_URL>/api/v1/findx-agents/installers/windows.ps1?package=<PACKAGE_ID>' -OutFile \"$env:TEMP\\findx-agent-install.ps1\"; $env:FINDX_BASE_URL='<FINDX_URL>'; $env:FINDX_TOKEN='<TOKEN>'; powershell -ExecutionPolicy Bypass -File \"$env:TEMP\\findx-agent-install.ps1\"",
  },
  {
    id: 'helm',
    label: 'Kubernetes Helm 预览',
    desc: '预览 FindX Agent Helm values 和渲染命令；真实执行必须等待 Kubernetes 生命周期契约开放。',
    command: kubernetesPreview([
      "helm template <RELEASE_NAME> <HELM_CHART_REF> --namespace <NAMESPACE> --set findx.baseUrl='<FINDX_URL>' --set findx.token='<TOKEN>' --set agent.packageId='<PACKAGE_ID>' --set packageRepositoryRef='<PACKAGE_REPOSITORY_REF>' --set signatureRef='<SIGNATURE_REF>'",
      '# required_refs: <CLUSTER_REF> <VALUES_REF> <RBAC_REF> <SERVICE_ACCOUNT_REF> <ROLLOUT_STRATEGY_REF> <DATA_ARRIVAL_VALIDATOR_REF> <EVIDENCE_CHAIN_REF>',
    ]),
  },
  {
    id: 'kubernetes-daemonset',
    label: 'Kubernetes DaemonSet 预览',
    desc: '预览节点级 FindX Agent DaemonSet manifest；RBAC、镜像签名和回执验证未开放前保持阻断。',
    command: kubernetesPreview([
      'apiVersion: apps/v1',
      'kind: DaemonSet',
      'metadata:',
      '  name: findx-agent',
      '  namespace: <NAMESPACE>',
      'spec:',
      '  selector: { matchLabels: { app: findx-agent } }',
      '  template:',
      '    metadata: { labels: { app: findx-agent } }',
      '    spec:',
      '      serviceAccountName: <SERVICE_ACCOUNT_REF>',
      '      containers:',
      '        - name: findx-agent',
      "          image: '<IMAGE_REF>'",
      "          env: [{ name: FINDX_BASE_URL, value: '<FINDX_URL>' }, { name: FINDX_TOKEN, value: '<TOKEN>' }, { name: FINDX_PACKAGE_ID, value: '<PACKAGE_ID>' }]",
    ]),
  },
  {
    id: 'kubernetes-operator',
    label: 'Kubernetes Operator 预览',
    desc: '预览 FindX Agent Operator 自定义资源；控制器、CRD、权限和回执契约未开放前保持阻断。',
    command: kubernetesPreview([
      'apiVersion: agent.findx.io/v1alpha1',
      'kind: FindXAgentInstallation',
      'metadata:',
      '  name: <RELEASE_NAME>',
      '  namespace: <NAMESPACE>',
      'spec:',
      '  clusterRef: <CLUSTER_REF>',
      '  packageId: <PACKAGE_ID>',
      '  packageRepositoryRef: <PACKAGE_REPOSITORY_REF>',
      '  imageRef: <IMAGE_REF>',
      '  credentialRef: <CREDENTIAL_REF>',
      '  workloadSelectorRef: <WORKLOAD_SELECTOR_REF>',
      '  rolloutStrategyRef: <ROLLOUT_STRATEGY_REF>',
      '  dataArrivalValidatorRef: <DATA_ARRIVAL_VALIDATOR_REF>',
      '  evidenceChainRef: <EVIDENCE_CHAIN_REF>',
    ]),
  },
  {
    id: 'kubernetes-sidecar',
    label: 'Kubernetes Sidecar 预览',
    desc: '预览工作负载注入 FindX Agent sidecar 的 patch；只展示 manifest，不写入集群。',
    command: kubernetesPreview([
      'apiVersion: apps/v1',
      'kind: Deployment',
      'metadata:',
      '  name: <WORKLOAD_NAME>',
      '  namespace: <NAMESPACE>',
      'spec:',
      '  template:',
      '    spec:',
      '      containers:',
      '        - name: findx-agent-sidecar',
      "          image: '<IMAGE_REF>'",
      "          envFrom: [{ secretRef: { name: '<SECRET_REF>' } }, { configMapRef: { name: '<CONFIG_MAP_REF>' } }]",
      "          env: [{ name: FINDX_BASE_URL, value: '<FINDX_URL>' }, { name: FINDX_PACKAGE_ID, value: '<PACKAGE_ID>' }]",
    ]),
  },
  {
    id: 'kubernetes-initcontainer',
    label: 'Kubernetes InitContainer 预览',
    desc: '预览 FindX Agent 初始化容器 patch；真实注入和回滚必须等待执行器与审计契约开放。',
    command: kubernetesPreview([
      'apiVersion: apps/v1',
      'kind: Deployment',
      'metadata:',
      '  name: <WORKLOAD_NAME>',
      '  namespace: <NAMESPACE>',
      'spec:',
      '  template:',
      '    spec:',
      '      initContainers:',
      '        - name: findx-agent-init',
      "          image: '<IMAGE_REF>'",
      "          env: [{ name: FINDX_BASE_URL, value: '<FINDX_URL>' }, { name: FINDX_TOKEN, value: '<TOKEN>' }, { name: FINDX_PACKAGE_ID, value: '<PACKAGE_ID>' }]",
      "          volumeMounts: [{ name: '<CONFIG_VOLUME_REF>', mountPath: '/findx-agent' }]",
    ]),
  },
]

const trimOrigin = origin => String(origin || '').replace(/\/+$/, '')

export const resolveFindxOrigin = () => {
  if (typeof window !== 'undefined' && window.location?.origin) return window.location.origin
  return '<FINDX_URL>'
}

export const packageDownloadUrl = (packageId, origin = resolveFindxOrigin()) => {
  const base = trimOrigin(origin)
  return `${base}/api/v1/findx-agents/package-downloads/${encodeURIComponent(packageId || 'agent-core')}`
}

export const renderInstallCommand = (item, { packageId, origin = resolveFindxOrigin() } = {}) => {
  const base = trimOrigin(origin)
  const packageKey = packageId || 'agent-core'
  return String(item?.command || '')
    .replaceAll('<FINDX_URL>', base)
    .replaceAll('<PACKAGE_ID>', encodeURIComponent(packageKey))
    .replaceAll('<PACKAGE_URL>', encodeURIComponent(packageKey))
}

export const configTemplates = [
  {
    id: 'agent-core',
    name: 'Agent 基础配置',
    configKind: '基础配置',
    scope: '注册 / 心跳 / 任务 / 标签',
    fields: ['agent_id', 'heartbeat_interval', 'global_labels', 'task_channel', 'credential_ref'],
    remoteDelivery: true,
    targetModes: ['全部 Agent', '业务组', '主机', '能力包'],
    rolloutScopes: ['注册参数', '心跳周期', '任务通道', '全局标签'],
    rolloutStrategies: ['保存模板', '灰度下发', '全量下发', '回滚'],
    rollbackPolicy: '按配置版本回滚到上一套稳定基础配置',
    capabilityPackages: ['FindX Agent 核心'],
  },
  {
    id: 'metrics',
    name: '指标采集配置',
    configKind: '采集配置',
    scope: '主机 / 容器 / 进程',
    fields: ['scrape_interval', 'labels', 'resource_group', 'credential_ref'],
    remoteDelivery: true,
    targetModes: ['全部 Agent', '业务组', '主机', '能力包'],
    rolloutScopes: ['主机指标', '容器指标', '进程指标'],
    rolloutStrategies: ['保存模板', '灰度下发', '全量下发', '回滚'],
    rollbackPolicy: '按采集配置版本回滚到上一套稳定版本',
    capabilityPackages: ['主机采集能力包', '容器采集能力包'],
  },
  {
    id: 'logs',
    name: '日志采集配置',
    configKind: '日志配置',
    scope: '文件 / 标准输出 / 系统日志',
    fields: ['paths', 'parser', 'pipeline_ref', 'labels'],
    remoteDelivery: true,
    targetModes: ['全部 Agent', '业务组', '主机', '能力包'],
    rolloutScopes: ['文件日志', '容器标准输出', '系统日志'],
    rolloutStrategies: ['保存模板', '灰度下发', '全量下发', '回滚'],
    rollbackPolicy: '按管道引用和采集路径回滚',
    capabilityPackages: ['日志采集能力包'],
  },
  {
    id: 'tracing',
    name: '应用链路配置',
    configKind: '链路配置',
    scope: '应用 / 服务 / 网关',
    fields: ['service_name', 'collector_endpoint', 'sampling', 'propagation'],
    remoteDelivery: true,
    targetModes: ['业务组', '主机', '能力包', '服务'],
    rolloutScopes: ['应用链路', '服务调用', '网关入口'],
    rolloutStrategies: ['保存模板', '灰度下发', '全量下发', '回滚'],
    rollbackPolicy: '保留采样率、传播协议和接入端点版本快照',
    capabilityPackages: ['应用链路能力包', '网关链路能力包'],
  },
  {
    id: 'profiling',
    name: '性能分析配置',
    configKind: '性能分析配置',
    scope: '应用运行时',
    fields: ['target_runtime', 'duration', 'cpu', 'memory'],
    remoteDelivery: true,
    targetModes: ['业务组', '主机', '能力包'],
    rolloutScopes: ['CPU 分析', '内存分析', '运行时分析'],
    rolloutStrategies: ['保存模板', '灰度下发', '全量下发', '回滚'],
    rollbackPolicy: '到期自动关闭并恢复默认采样策略',
    capabilityPackages: ['应用链路能力包'],
  },
  {
    id: 'inspection',
    name: '巡检诊断配置',
    configKind: '巡检配置',
    scope: '主机 / 服务 / 端口',
    fields: ['check_set', 'schedule', 'risk_level', 'evidence_chain'],
    remoteDelivery: true,
    targetModes: ['全部 Agent', '业务组', '主机', '能力包'],
    rolloutScopes: ['主机巡检', '服务巡检', '端口巡检'],
    rolloutStrategies: ['保存模板', '灰度下发', '全量下发', '回滚'],
    rollbackPolicy: '回滚到上一套巡检集和调度周期',
    capabilityPackages: ['巡检诊断能力包'],
  },
  {
    id: 'host-plugin',
    name: '主机插件配置',
    configKind: '插件配置',
    scope: 'Linux / Windows 主机插件',
    fields: ['plugin_id', 'plugin_version', 'config_format', 'config_snippet_ref', 'provider_mode', 'reload_strategy', 'remote_mutation', 'canary_percent', 'rollback_ref', 'audit_reason', 'change_ticket', 'credential_ref'],
    remoteDelivery: true,
    targetModes: ['全部 Agent', '业务组', '主机', '能力包'],
    rolloutScopes: ['系统插件', '进程插件', '磁盘插件', '网络插件', 'HTTP Provider 配置'],
    rolloutStrategies: ['保存模板', '灰度下发', '全量下发', '回滚'],
    rollbackPolicy: '按插件配置版本和 TOML 片段引用回滚，并保留远程修改审计',
    capabilityPackages: ['主机采集能力包', '巡检诊断能力包'],
    pluginConfig: pluginConfigSpec('input.cpu / input.mem / input.disk / input.net', 'local-reload'),
  },
  {
    id: 'container-plugin',
    name: '容器插件配置',
    configKind: '插件配置',
    scope: 'Kubernetes / Docker 插件',
    fields: ['plugin_id', 'plugin_version', 'namespace', 'workload_selector', 'config_format', 'config_snippet_ref', 'provider_mode', 'restart_strategy', 'remote_mutation', 'canary_percent', 'rollback_ref', 'audit_reason', 'change_ticket', 'credential_ref'],
    remoteDelivery: true,
    targetModes: ['业务组', '集群', '命名空间', '能力包'],
    rolloutScopes: ['集群插件', '命名空间插件', '工作负载插件', '容器采集插件'],
    rolloutStrategies: ['保存模板', '灰度下发', '全量下发', '回滚'],
    rollbackPolicy: '回滚到上一套集群选择器、工作负载选择器和 TOML 片段引用',
    capabilityPackages: ['容器采集能力包'],
    pluginConfig: pluginConfigSpec('input.docker / input.cadvisor / input.kubernetes', 'rolling-restart'),
  },
  {
    id: 'gateway-plugin',
    name: '网关插件配置',
    configKind: '插件配置',
    scope: '网关 / 反向代理 / 边缘节点',
    fields: ['plugin_id', 'plugin_version', 'gateway_id', 'route_selector', 'collector_endpoint', 'config_format', 'config_snippet_ref', 'provider_mode', 'reload_strategy', 'remote_mutation', 'canary_percent', 'rollback_ref', 'audit_reason', 'change_ticket', 'credential_ref'],
    remoteDelivery: true,
    targetModes: ['业务组', '网关', '主机', '能力包'],
    rolloutScopes: ['网关插件', '路由规则', '入口链路'],
    rolloutStrategies: ['保存模板', '灰度下发', '全量下发', '回滚'],
    rollbackPolicy: '回滚到上一套路由选择器、链路上下文和插件配置片段',
    capabilityPackages: ['网关链路能力包'],
    pluginConfig: pluginConfigSpec('gateway module / reverse proxy plugin', 'reload'),
  },
  {
    id: 'browser-probe',
    name: '前端体验配置',
    configKind: '前端体验配置',
    scope: 'Web 应用 / 页面 / 域名',
    fields: ['app_id', 'domain', 'collector_endpoint', 'trace_context', 'sampling'],
    remoteDelivery: true,
    targetModes: ['业务组', 'Web 应用', '域名', '能力包'],
    rolloutScopes: ['前端链路', '页面性能', '错误采集'],
    rolloutStrategies: ['保存模板', '灰度下发', '全量下发', '回滚'],
    rollbackPolicy: '回滚到上一套应用标识、采样率和上下文传播配置',
    capabilityPackages: ['前端体验能力包'],
  },
]
