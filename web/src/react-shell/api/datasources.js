import { get, del, normalizeList, post } from './http.js'

export const DATASOURCE_CONTRACT_BLOCKERS = {
  create: '缺少数据源单条新增/upsert 契约、plugin/list、凭据引用、Header/TLS/mTLS 持久化契约，当前不执行新增保存。',
  edit: '缺少数据源详情、单条 upsert、凭据不回显、Header/TLS/mTLS 持久化契约，当前不执行编辑保存。',
  toggle: '缺少 status/update、权限点、审计和状态回滚契约，当前不执行启停。',
  delete: '缺少 delete、幂等、审计和权限契约，当前不执行删除。',
  pluginList: '缺少 plugin/list 契约，当前仅展示 FindX 已知类型入口。',
  clusters: '缺少 server-clusters 契约，当前不校验采集集群有效性。',
  testNonMetric: '该类型缺少对应测试查询契约，当前 /monitor/query query=up 仅覆盖明确的指标型数据源。',
}

export const datasourceApi = {
  async list() {
    try {
      const data = await get('/data-sources')
      return { rows: normalizeList(data), source: 'GET /api/v1/data-sources' }
    } catch (error) {
      if (error?.status && error.status !== 404 && error.status !== 405) throw error
      const data = await get('/monitor/datasources')
      return {
        rows: normalizeList(data),
        source: 'fallback GET /api/v1/monitor/datasources',
      }
    }
  },

  save(payload) {
    return post('/monitor/datasources', payload)
  },

  remove(id) {
    return del(`/monitor/datasources/${id}`)
  },

  testConnection() {
    return get('/monitor/labels')
  },

  test(datasourceId) {
    return post('/monitor/query', {
      datasource_id: datasourceId,
      query: 'up',
      timeout_seconds: 3,
    })
  },
}
