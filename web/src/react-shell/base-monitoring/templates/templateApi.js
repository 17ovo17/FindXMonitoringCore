/**
 * 模板中心 API 层
 * 对齐夜莺 builtInComponents 接口
 */
import { get, post, put, del } from '../../api/http.js'

const compact = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, v]) => v !== '' && v !== undefined && v !== null),
)

export const templateApi = {
  // 组件列表
  listComponents(params) {
    return get('/monitor/builtin-components', { params: compact(params) })
  },
  // 创建组件
  createComponents(data) {
    return post('/monitor/builtin-components', data)
  },
  // 更新组件
  updateComponent(data) {
    return put('/monitor/builtin-components', data)
  },
  // 删除组件
  deleteComponents(ids) {
    return del('/monitor/builtin-components', { data: { ids } })
  },
  // Payload 分类
  listPayloadCates(params) {
    return get('/monitor/builtin-payloads/cates', { params: compact(params) })
  },
  // Payload 列表
  listPayloads(params) {
    return get('/monitor/builtin-payloads', { params: compact(params) })
  },
  // 获取单个 Payload
  getPayload(id) {
    return get(`/monitor/builtin-payload/${encodeURIComponent(id)}`)
  },
  // 创建 Payload
  createPayloads(data) {
    return post('/monitor/builtin-payloads', data)
  },
  // 更新 Payload
  updatePayload(data) {
    return put('/monitor/builtin-payloads', data)
  },
  // 删除 Payload
  deletePayloads(ids) {
    return del('/monitor/builtin-payloads', { data: { ids } })
  },
  // 仪表盘模板导入
  importDashboardTemplate(id, body) {
    return post(`/monitor/dashboard-templates/${encodeURIComponent(id)}/import`, body)
  },
}
