import { pluginConfigSpec } from './agentPluginConfigModel.js'

const BLOCKED = 'BLOCKED_BY_CONTRACT'
const MISSING_SOURCE = 'LOCAL_SOURCE_MISSING'
const MISSING_PACKAGE = 'PACKAGE_MISSING'
const SOURCE_PRESENT = 'LOCAL_SOURCE_PRESENT'

const blocked = reason => `${BLOCKED}: ${reason}`

const methodRow = (platform, method, toolEvidence, sourceState, packageState, extra = {}) => ({
  platform,
  method,
  toolEvidence,
  sourceState,
  packageState,
  executor: extra.executor || BLOCKED,
  serviceRegistration: extra.serviceRegistration || BLOCKED,
  configDelivery: extra.configDelivery || BLOCKED,
  uninstall: extra.uninstall || BLOCKED,
  rollback: extra.rollback || BLOCKED,
  dataArrival: extra.dataArrival || BLOCKED,
  blocker: extra.blocker || blocked('缺少安装执行器、包仓库、服务回执和数据到达验证。'),
})

const bundledPreviewMatrix = (sourceState, packageState, blocker) => [
  methodRow('Linux', 'curl -kfsSL 命令预览', '系统自带 curl 可作为下载命令预览；未证明真实安装。', sourceState, packageState, { blocker }),
  methodRow('Windows CMD', 'certutil -urlcache -f 命令预览', '系统自带 certutil 可作为下载命令预览；未证明真实安装。', sourceState, packageState, { blocker }),
  methodRow('PowerShell', 'Invoke-WebRequest 命令预览', '系统自带 PowerShell 可作为下载命令预览；未证明真实安装。', sourceState, packageState, { blocker }),
  methodRow('Kubernetes', 'Helm / Operator / DaemonSet / Sidecar / InitContainer 预览', 'Kubernetes 原生命令形态可预览；未证明集群安装闭环。', sourceState, packageState, { blocker }),
]

const appProbeBlocker = blocked('应用链路探针本地源码或包仓库缺失，只能展示 blocked 控制面。')
const collectorBlocker = blocked('采集插件源码线索较完整，但 FindX 远程修改、reload、漂移检测、回滚和 receipt 契约缺失。')
const inspectionBlocker = blocked('巡检诊断源码有线索，但缺二进制包仓库、安装服务和执行回执。')
const coreBlocker = blocked('FindX Agent Core 缺少本地包仓库、签名、安装执行器、服务注册和心跳闭环。')

const pluginDeliveryScope = [
  'Agent',
  'CMDB 主机',
  '业务组',
  'namespace/workload',
]

const categrafPluginConfig = (pluginId, reloadStrategy) => ({
  ...pluginConfigSpec(pluginId, reloadStrategy),
  delivery_scope: pluginDeliveryScope,
  remote_mutation_status: BLOCKED,
  remote_reload_status: BLOCKED,
  drift_detection_status: BLOCKED,
  rollback_status: BLOCKED,
  receipt_status: BLOCKED,
})

const appPackage = (id, name, runtime) => ({
  id,
  name,
  capabilityDomain: '应用链路',
  runtime,
  os: ['Linux', 'Windows', 'Kubernetes'],
  sourceState: MISSING_SOURCE,
  packageState: MISSING_PACKAGE,
  packageShape: 'FindX Agent runtime package / SDK / sidecar / init container',
  telemetryKinds: ['tracing', 'profiling', 'topology'],
  configKeys: ['service_name', 'collector_endpoint', 'instance_name', 'environment', 'sampling'],
  configTemplateIds: ['tracing', 'profiling'],
  environmentMatrix: bundledPreviewMatrix(MISSING_SOURCE, MISSING_PACKAGE, appProbeBlocker),
  blockers: [appProbeBlocker],
})

export const capabilityPackages = [
  {
    id: 'agent-core',
    name: 'FindX Agent 核心',
    capabilityDomain: '基础 Agent',
    runtime: 'Agent Core',
    os: ['Linux', 'Windows', 'Kubernetes'],
    sourceState: MISSING_SOURCE,
    packageState: MISSING_PACKAGE,
    packageShape: 'service / daemon / sidecar',
    telemetryKinds: ['heartbeat', 'task', 'config'],
    configKeys: ['agent_id', 'heartbeat_interval', 'task_channel', 'global_labels', 'credential_ref'],
    configTemplateIds: ['agent-core'],
    environmentMatrix: bundledPreviewMatrix(MISSING_SOURCE, MISSING_PACKAGE, coreBlocker),
    blockers: [coreBlocker],
  },
  {
    id: 'host-collector',
    name: '主机采集能力包',
    capabilityDomain: '基础采集',
    runtime: 'Host',
    os: ['Linux', 'Windows'],
    sourceState: SOURCE_PRESENT,
    packageState: MISSING_PACKAGE,
    packageShape: 'FindX Agent collector plugins / service',
    telemetryKinds: ['metrics', 'process', 'host'],
    configKeys: ['scrape_interval', 'plugin_set', 'plugin_id', 'plugin_version', 'config_snippet_ref', 'provider_mode', 'reload_strategy', 'global_labels', 'credential_ref'],
    configTemplateIds: ['metrics', 'host-plugin'],
    pluginConfig: categrafPluginConfig('input.cpu / input.mem / input.disk / input.net', 'local-reload'),
    environmentMatrix: bundledPreviewMatrix(SOURCE_PRESENT, MISSING_PACKAGE, collectorBlocker),
    blockers: [collectorBlocker],
  },
  {
    id: 'container-collector',
    name: '容器采集能力包',
    capabilityDomain: '基础采集',
    runtime: 'Container',
    os: ['Linux', 'Kubernetes'],
    sourceState: SOURCE_PRESENT,
    packageState: MISSING_PACKAGE,
    packageShape: 'FindX Agent daemonset / sidecar / container plugin',
    telemetryKinds: ['metrics', 'container', 'workload'],
    configKeys: ['cluster_ref', 'namespace_selector', 'workload_selector', 'plugin_id', 'config_snippet_ref', 'provider_mode', 'restart_strategy', 'credential_ref'],
    configTemplateIds: ['metrics', 'container-plugin'],
    pluginConfig: categrafPluginConfig('input.docker / input.cadvisor / input.kubernetes', 'rolling-restart'),
    environmentMatrix: bundledPreviewMatrix(SOURCE_PRESENT, MISSING_PACKAGE, collectorBlocker),
    blockers: [collectorBlocker],
  },
  {
    id: 'log-collector',
    name: '日志采集能力包',
    capabilityDomain: '日志采集',
    runtime: 'Logs',
    os: ['Linux', 'Windows', 'Kubernetes'],
    sourceState: SOURCE_PRESENT,
    packageState: MISSING_PACKAGE,
    packageShape: 'FindX Agent file tailer / stdout collector / parser',
    telemetryKinds: ['logs', 'pipeline'],
    configKeys: ['paths', 'parser', 'pipeline_ref', 'labels', 'credential_ref'],
    configTemplateIds: ['logs'],
    environmentMatrix: bundledPreviewMatrix(SOURCE_PRESENT, MISSING_PACKAGE, collectorBlocker),
    blockers: [collectorBlocker],
  },
  {
    id: 'inspection-runner',
    name: '巡检诊断能力包',
    capabilityDomain: '巡检诊断',
    runtime: 'Inspection',
    os: ['Linux', 'Windows', 'Kubernetes'],
    sourceState: SOURCE_PRESENT,
    packageState: MISSING_PACKAGE,
    packageShape: 'FindX Agent check runner / diagnostic plugin',
    telemetryKinds: ['inspection', 'evidence'],
    configKeys: ['check_set', 'schedule', 'risk_level', 'evidence_chain', 'credential_ref'],
    configTemplateIds: ['inspection', 'host-plugin'],
    environmentMatrix: bundledPreviewMatrix(SOURCE_PRESENT, MISSING_PACKAGE, inspectionBlocker),
    blockers: [inspectionBlocker],
  },
  appPackage('java-app', 'Java 应用能力包', 'Java'),
  appPackage('python-app', 'Python 应用能力包', 'Python'),
  appPackage('nodejs-app', 'Node.js 应用能力包', 'Node.js'),
  appPackage('php-app', 'PHP 应用能力包', 'PHP'),
  appPackage('go-app', 'Go 应用能力包', 'Go'),
  appPackage('rust-app', 'Rust 应用能力包', 'Rust'),
  appPackage('ruby-app', 'Ruby 应用能力包', 'Ruby'),
  {
    id: 'gateway-probe',
    name: '网关链路能力包',
    capabilityDomain: '网关链路',
    runtime: 'Gateway',
    os: ['Linux', 'Kubernetes'],
    sourceState: MISSING_SOURCE,
    packageState: MISSING_PACKAGE,
    packageShape: 'FindX Agent gateway plugin / reverse proxy module',
    telemetryKinds: ['tracing', 'topology', 'logs'],
    configKeys: ['gateway_id', 'route_selector', 'collector_endpoint', 'sampling', 'reload_policy'],
    configTemplateIds: ['gateway-plugin', 'tracing', 'logs'],
    pluginConfig: categrafPluginConfig('gateway module / reverse proxy plugin', 'reload'),
    environmentMatrix: bundledPreviewMatrix(MISSING_SOURCE, MISSING_PACKAGE, appProbeBlocker),
    blockers: [appProbeBlocker, collectorBlocker],
  },
  {
    id: 'browser-client',
    name: '前端体验能力包',
    capabilityDomain: '前端体验',
    runtime: 'Web',
    os: ['Web'],
    sourceState: MISSING_SOURCE,
    packageState: MISSING_PACKAGE,
    packageShape: 'FindX Agent Browser Client JavaScript SDK',
    telemetryKinds: ['rum', 'tracing', 'errors'],
    configKeys: ['app_id', 'domain', 'collector_endpoint', 'trace_context', 'sampling'],
    configTemplateIds: ['browser-probe'],
    environmentMatrix: [
      methodRow('Browser', 'JavaScript SDK 接入片段预览', '浏览器脚本入口可预览；未证明 SDK 包仓库、签名和数据到达。', MISSING_SOURCE, MISSING_PACKAGE, { blocker: appProbeBlocker }),
    ],
    blockers: [appProbeBlocker],
  },
]
