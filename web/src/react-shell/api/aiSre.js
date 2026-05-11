import { get, del, isPermissionError, normalizeList, post, put, redactText } from './http.js'

export const AISRE_BLOCKERS = {
  evidence: 'BLOCKED_BY_CONTRACT: 指标、日志、链路、告警、CMDB、Agent、巡检和工作流的统一证据聚合契约未开放。',
  workflowRun: 'BLOCKED_BY_CONTRACT: 工作流执行路由缺少 AI SRE 权限、审批、审计、回滚和错误脱敏契约。',
  remediation: 'BLOCKED_BY_CONTRACT: 自动修复缺少审批、执行计划、远程执行、回滚、审计和 Evidence Chain 关联契约。',
  reportExport: 'BLOCKED_BY_CONTRACT: 复盘报告导出、分享、脱敏预览和留档契约未开放。',
  knowledgeWrite: 'BLOCKED_BY_CONTRACT: 知识库写入、导入、删除和权限失败态需要完整审计契约。',
  dataArrival: 'BLOCKED_BY_CONTRACT: Agent 数据到达、链路覆盖率和巡检数据质量缺少统一健康契约。',
}

const cleanParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== null && value !== undefined),
)

const unwrapData = value => value?.data ?? value

export const formatAiSreError = error => {
  if (isPermissionError(error)) return error.status === 401 ? '登录状态已过期，请重新登录。' : '当前账号没有 AI SRE 操作权限。'
  if ([404, 405, 501].includes(error?.status)) return `BLOCKED_BY_CONTRACT: ${redactText(error.message || '接口未开放')}`
  return redactText(error?.message || 'AI SRE 请求失败')
}

const normalizeMessages = value => {
  const data = unwrapData(value)
  if (Array.isArray(data?.messages)) return data.messages
  return normalizeList(data)
}

export const aiSreApi = {
  createSession: body => post('/aiops/sessions', body).then(unwrapData),
  sendMessage: (sessionId, body) => post(`/aiops/sessions/${encodeURIComponent(sessionId)}/messages`, body).then(unwrapData),
  getMessages: sessionId => get(`/aiops/sessions/${encodeURIComponent(sessionId)}/messages`).then(normalizeMessages),
  executeAction: (sessionId, body) => post(`/aiops/sessions/${encodeURIComponent(sessionId)}/actions/execute`, body).then(unwrapData),
  evidenceChain: params => get('/aiops/evidence-chain', { params: cleanParams(params) }).then(unwrapData),
  health: {
    datasources: () => get('/health/datasources'),
    aiProviders: () => get('/health/ai-providers'),
    storage: () => get('/health/storage'),
    agents: () => get('/health/agents'),
    prometheus: () => get('/health/prometheus'),
    check: target => post('/health/check', { target }).then(unwrapData),
    checkAll: () => post('/health/check-all').then(unwrapData),
    history: params => get('/health/history', { params: cleanParams(params) }).then(normalizeList),
  },
  workflows: {
    list: params => get('/workflows', { params: cleanParams(params) }).then(normalizeList),
    detail: id => get(`/workflows/${encodeURIComponent(id)}`),
    runs: id => get(`/workflows/${encodeURIComponent(id)}/runs`).then(normalizeList),
    create: body => post('/workflows', body).then(unwrapData),
    update: (id, body) => put(`/workflows/${encodeURIComponent(id)}`, body).then(unwrapData),
    toggle: (id, enabled) => put(`/workflows/${encodeURIComponent(id)}/toggle`, { enabled }).then(unwrapData),
    execute: id => post(`/workflows/${encodeURIComponent(id)}/execute`).then(unwrapData),
    runDetail: (wfId, runId) => get(`/workflows/${encodeURIComponent(wfId)}/runs/${encodeURIComponent(runId)}`).then(unwrapData),
  },
  inspections: {
    create: body => post('/aiops/inspections', body).then(unwrapData),
    progress: id => get(`/aiops/inspections/${encodeURIComponent(id)}/progress`).then(unwrapData),
    report: id => get(`/aiops/inspections/${encodeURIComponent(id)}/report`).then(unwrapData),
  },
  knowledge: {
    search: body => post('/knowledge/search', body),
    cases: params => get('/knowledge/cases', { params: cleanParams(params) }).then(normalizeList),
    list: params => get('/knowledge/documents', { params: cleanParams(params) }).then(normalizeList),
    detail: id => get(`/knowledge/documents/${encodeURIComponent(id)}`).then(unwrapData),
    create: body => post('/knowledge/documents', body).then(unwrapData),
    update: (id, body) => put(`/knowledge/documents/${encodeURIComponent(id)}`, body).then(unwrapData),
    remove: id => del(`/knowledge/documents/${encodeURIComponent(id)}`).then(unwrapData),
    runbooks: params => get('/knowledge/runbooks', { params: cleanParams(params) }).then(normalizeList),
    createRunbook: body => post('/knowledge/runbooks', body).then(unwrapData),
    removeRunbook: id => del(`/knowledge/runbooks/${encodeURIComponent(id)}`).then(unwrapData),
  },
  chat: {
    listSessions: (params = {}) => get('/aiops/sessions', { params: cleanParams(params) }).then(value => {
      const data = unwrapData(value)
      return { list: normalizeList(data), total: data?.total ?? 0 }
    }),
    createSession: body => post('/aiops/sessions', body).then(unwrapData),
    deleteSession: id => del(`/aiops/sessions/${encodeURIComponent(id)}`).then(unwrapData),
    renameSession: (id, title) => put(`/aiops/sessions/${encodeURIComponent(id)}`, { title }).then(unwrapData),
    getMessages: id => get(`/aiops/sessions/${encodeURIComponent(id)}/messages`).then(value => {
      const data = unwrapData(value)
      if (Array.isArray(data?.messages)) return data.messages
      return normalizeList(data)
    }),
    sendMessageSSE: (sessionId, body, onEvent) => {
      const controller = new AbortController()
      const url = `/api/v1/aiops/sessions/${encodeURIComponent(sessionId)}/messages`
      const authHeader = {}
      try {
        const stored = localStorage.getItem('findx_token') || sessionStorage.getItem('findx_token')
        if (stored) authHeader['Authorization'] = `Bearer ${stored}`
      } catch { /* ignore */ }
      const promise = fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Accept': 'text/event-stream', ...authHeader },
        body: JSON.stringify(body),
        signal: controller.signal,
      }).then(async resp => {
        if (!resp.ok) throw new Error(`HTTP ${resp.status}: 请求失败`)
        const reader = resp.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ''
        while (true) {
          const { done, value: chunk } = await reader.read()
          if (done) break
          buffer += decoder.decode(chunk, { stream: true })
          const lines = buffer.split('\n')
          buffer = lines.pop() || ''
          for (const line of lines) {
            if (line.startsWith('data:')) {
              const raw = line.slice(5).trim()
              if (raw === '[DONE]') { onEvent({ type: 'done' }); return }
              try {
                const parsed = JSON.parse(raw)
                onEvent(parsed)
              } catch {
                onEvent({ type: 'content', content: raw })
              }
            } else if (line.startsWith('event:')) {
              /* event type line — handled via data payload */
            }
          }
        }
        if (buffer.trim()) {
          onEvent({ type: 'content', content: buffer.trim() })
        }
        onEvent({ type: 'done' })
      })
      return { promise, abort: () => controller.abort() }
    },
  },
}
