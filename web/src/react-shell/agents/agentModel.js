export const sections = [
  { value: 'overview', label: '概览', desc: '统一查看 FindX Agent 能力覆盖、在线状态和阻断契约。' },
  { value: 'hosts', label: '主机 Agent', desc: '联动 CMDB 主机、安装状态、心跳、数据到达、安装计划和配置下发。' },
  { value: 'packages', label: '能力包', desc: '按统一能力域管理采集、日志、应用链路、网关、前端体验和巡检能力包。' },
  { value: 'plugin-catalog', label: '插件目录', desc: '查看和管理所有可用的采集、诊断和 APM 插件。' },
  { value: 'config-push', label: '配置下发', desc: '批量下发插件配置到多个 Agent。' },
  { value: 'environment', label: '环境适配', desc: '探测 Agent 环境并自动推荐插件配置。' },
]

export const sectionSet = new Set(sections.map(item => item.value))
export const hostMergedSections = new Set(['install', 'templates', 'heartbeat', 'data-arrival', 'config', 'plugins'])

export { capabilityPackages } from './agentCapabilityModel.js'
export { installCommands, configTemplates } from './agentTemplateModel.js'
export { displayText, fmtTime, agentKey, agentOnline, sourceStateLabel, rowText } from './agentTextModel.js'
