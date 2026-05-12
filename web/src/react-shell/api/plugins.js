import { get, post, put, normalizeList } from './http.js'

export const pluginApi = {
  // 插件目录
  listPlugins: (category) => {
    const params = category ? { category } : {}
    return get('/findx-agents/plugins', { params }).then(normalizeList)
  },

  // 配置管理
  getConfig: (agentId) => get(`/findx-agents/${agentId}/config`),
  updateConfig: (agentId, plugins) => put(`/findx-agents/${agentId}/config`, { plugins }),
  patchPlugin: (agentId, pluginId, enabled) =>
    fetch(`/api/v1/findx-agents/${agentId}/plugins/${pluginId}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled }),
    }).then(r => r.json()),
  updatePluginConfig: (agentId, pluginId, config) =>
    post(`/findx-agents/${agentId}/plugins/${pluginId}/config`, { config }),

  // 配置下发
  configPush: (body) => post('/findx-agents/config-push', body),

  // 环境探测
  getEnvironment: (agentId) => get(`/findx-agents/${agentId}/environment`),
  autoAdapt: (agentId) => post(`/findx-agents/${agentId}/auto-adapt`),

  // 插件启停
  startPlugin: (agentId, pluginId) =>
    post(`/findx-agents/${agentId}/plugins/${pluginId}/start`),
  stopPlugin: (agentId, pluginId) =>
    post(`/findx-agents/${agentId}/plugins/${pluginId}/stop`),
}
