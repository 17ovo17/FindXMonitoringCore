import { AUTH_EXPIRED_EVENT, get, post, put, del, redactText } from './http.js'

export const CMDB_RELATION_GRAPH_CONTRACT_ID = 'FX-NIGHT-126C-CMDB-RELATION-GRAPH-TOPOLOGY-UI-CONTRACT'
export const CMDB_RESOURCE_STATS_CONTRACT_ID = 'FX-NIGHT-128M-CMDB-RESOURCE-STATISTICS-UI-CONTRACT'
export const CMDB_RESOURCE_APPROVAL_CONTRACT_ID = 'cmdb.resource.approval.runtime.v1'
export const CMDB_HOST_AI_SESSION_CONTRACT_ID = 'cmdb.ai.host_session.runtime.v1'

const CMDB_RESOURCE_STATS_EXPECTED_SCHEMA = {
  request: ['workspace_id', 'resource_group_id', 'object_id'],
  totals: ['cmdb_instances', 'cmdb_models', 'host_assets', 'findx_agents', 'monitor_binding_rows', 'monitor_bound_instances', 'relation_edges'],
  dimensions: ['model_distribution[]', 'resource_group_distribution[]', 'business_group_distribution[]'],
  audit: ['blocked_contracts[]', 'missing_contracts[]', 'findx_audit_query', 'source_evidence[]'],
}

const CMDB_RESOURCE_STATS_EVIDENCE = [
  'FindX CMDB resource-statistics API',
  'FindX CMDB object, instance, group and audit aggregation',
  'FindX audit query: scope=cmdb resource_type=resource_statistics',
]

const CMDB_RESOURCE_APPROVAL_VIEWS = new Set(['mine', 'todo', 'archive'])

const CMDB_RESOURCE_APPROVAL_EXPECTED_SCHEMA = {
  request: ['view=mine|todo|archive'],
  runtime: ['approval_requests[]', 'workflow_state', 'risk_policy', 'audit_receipt'],
  audit: ['status', 'contract_id', 'missing_contracts[]', 'blocked_contracts[]', 'findx_audit_query', 'view'],
}

const CMDB_RESOURCE_APPROVAL_MISSING_CONTRACTS = [
  'cmdb_resource_approval_store_contract',
  'cmdb_resource_approval_workflow_contract',
  'cmdb_operation_risk_policy_contract',
  'cmdb_approval_audit_receipt_contract',
]

const CMDB_HOST_AI_SESSION_MISSING_CONTRACTS = [
  'cmdb_resource_approval_runtime_contract',
  'cmdb_operation_risk_policy_contract',
  'cmdb_action_preflight_contract',
  'cmdb_action_audit_receipt_contract',
  'cmdb_ai_host_session_transport_contract',
  'cmdb_ai_host_context_contract',
  'cmdb_ai_tool_audit_contract',
  'cmdb_ai_output_receipt_contract',
  'cmdb_ai_session_scope_contract',
  'cmdb_ai_command_risk_policy_contract',
]

export const CMDB_RELATION_GRAPH_EVIDENCE = [
  '来源证据: CMDB API 契约，按 object_id + instance_id + topology_id 三元定位拓扑',
  '来源证据: CMDB 主机表、实例详情、业务视图、监控绑定和审计记录',
  '测试证据: cmdb-instance-default-topology.json',
  '测试证据: cmdb-instance-default-topology-request.txt',
  '测试证据: cmdb-instance-default-topology-page.png',
  '测试证据: cmdb-instance-relation-page.png',
  '测试证据: cmdb-instance-monitor-bind-dialog.png',
  '测试证据: cmdb-model-detail-relations-page.png',
  '测试证据: cmdb-model-detail-relations-topology-page.png',
  '测试证据: cmdb-model-manage.png',
  '测试证据: cmdb-business-view.png',
  '测试证据: 542b9694-17e6-44d7-824f-7dc37198a67d.png - 资源列表，模型树、是否监控、分组、负责人、编辑/克隆/更多',
  '测试证据: 616e6f53-c80d-42bf-92fc-839a839c5e24.png - 机房视图，机柜/设备/告警/容量/负责人/位置',
  '测试证据: d1e9ec97-b36d-4afe-9051-473516bf995a.png - 自动化映射抽屉，模型对接、关联属性、约束条件、同步频率、开启同步',
  '测试证据: ed156c73-143f-440c-a919-39631c252a60.png - 资源概况，资源总数、模型总数、新增/删除、变更近况、TOP5',
  '测试证据: fcb03e2c-4860-4ea4-94b9-7110054d1ff2.png - 资产消费/账单统计，账期、Owner、消费趋势、产品分布',
]

export const CMDB_RELATION_GRAPH_EXPECTED_SCHEMA = {
  request: ['object_id', 'instance_id', 'topology_id'],
  envelope: ['_id', 'object_id', 'name', 'data', 'default', 'filter_empty', 'status', 'sort', 'instance_id', 'business_status', 'business_names'],
  graph: ['nodes[]', 'edges[]', 'business_context', 'audit_context', 'expected_schema', 'field_matrix', 'source_evidence', 'missing_contracts'],
  tree: ['data.object', 'data.instances[]', 'data.children[]', 'data.relation.in[]', 'data.relation.out[]'],
  node: ['id', 'name', 'object_id', 'object_name', 'object_type', 'level', 'raw'],
  edge: ['id', 'source', 'target', 'relation_id', 'relation_name', 'instance_relation_id', 'direction', 'location', 'raw'],
  secondary_actions: ['instance_detail', 'relation_detail', 'topology', 'monitor_binding', 'audit_logs'],
}

export const CMDB_RELATION_FIELD_MATRIX = [
  { field: 'request.object_id + request.instance_id + request.topology_id', required: true, purpose: '成熟拓扑入口三元定位；缺任一维度不能声明像素级拓扑闭环', source: '成熟拓扑请求抓包' },
  { field: 'envelope.default/filter_empty/status/sort/in_inst_detail', required: true, purpose: '拓扑配置、默认态和实例详情内展示状态', source: '成熟拓扑响应 envelope' },
  { field: 'business_names', required: true, purpose: '把实例关系拓扑合并进业务上下文，而不是孤立展示节点', source: '成熟实例拓扑响应' },
  { field: 'data.object.id/name/type', required: true, purpose: '模型节点唯一标识和显示名称', source: '成熟拓扑响应 data.object' },
  { field: 'data.instances[].id/name/object_id/object_name/object_type', required: true, purpose: '实例节点身份，用于详情、关系、拓扑和监控绑定动作', source: '成熟拓扑响应 data.instances' },
  { field: 'data.instances[].in/out[].related_instance_id', required: true, purpose: '实例关系边端点；空边必须保持阻断', source: '成熟拓扑响应实例关系' },
  { field: 'related_object_id + instance_relation_id + relation_id', required: true, purpose: '二级/三级动作的最小可审计身份字段', source: '成熟拓扑响应实例关系' },
  { field: 'children[].relation.in/out[].asst_id/asst_name/side/position/direction/location', required: true, purpose: '关系类型、方向和画布布局语义', source: '成熟模型关系拓扑' },
  { field: 'monitor_bindings', required: true, purpose: '资源列表“是否监控”和监控绑定动作必须来自真实绑定契约', source: '成熟监控绑定弹窗' },
  { field: 'audit_context', required: true, purpose: '资源列表分配主机后进入 FindX 审计日志，可按 scope=cmdb/resource_type=host_asset 查询', source: 'FindX audit log adapter' },
  { field: 'expected_schema', required: true, purpose: '126B/126C 后端声明的前端消费契约', source: 'FindX relation graph adapter' },
  { field: 'field_matrix', required: true, purpose: '字段映射、来源和缺口审计', source: 'FindX relation graph adapter' },
  { field: 'source_evidence', required: true, purpose: 'FindX 截图、抓包和用户新增 CMDB 图片证据', source: 'FindX relation graph adapter' },
  { field: 'missing_contracts', required: true, purpose: '缺失契约 ID，阻断空图伪成功', source: 'FindX relation graph adapter' },
]

const relationBlockedAudit = (reason, extra = {}) => ({
  status: 'PENDING',
  contract_id: CMDB_RELATION_GRAPH_CONTRACT_ID,
  message: reason,
  contract_gap_id: extra.contract_gap_id || CMDB_RELATION_GRAPH_CONTRACT_ID,
  contract_matrix: extra.contract_matrix || [],
  expected_schema: extra.expected_schema || CMDB_RELATION_GRAPH_EXPECTED_SCHEMA,
  field_matrix: extra.field_matrix || CMDB_RELATION_FIELD_MATRIX,
  source_evidence: extra.source_evidence || CMDB_RELATION_GRAPH_EVIDENCE,
  missing_contracts: extra.missing_contracts || [CMDB_RELATION_GRAPH_CONTRACT_ID],
})

const resourceStatsBlockedAudit = (reason, extra = {}) => ({
  status: 'PENDING',
  contract_id: CMDB_RESOURCE_STATS_CONTRACT_ID,
  message: reason,
  expected_schema: extra.expected_schema || CMDB_RESOURCE_STATS_EXPECTED_SCHEMA,
  source_evidence: extra.source_evidence || CMDB_RESOURCE_STATS_EVIDENCE,
  blocked_contracts: asArray(extra.blocked_contracts),
  missing_contracts: asArray(extra.missing_contracts).length ? asArray(extra.missing_contracts) : [CMDB_RESOURCE_STATS_CONTRACT_ID],
  audit_context: extra.audit_context,
})

const normalizeApprovalView = (view) => CMDB_RESOURCE_APPROVAL_VIEWS.has(view) ? view : 'mine'

const resourceApprovalBlockedAudit = (reason, extra = {}) => {
  const view = normalizeApprovalView(extra.view)
  const missingContracts = normalizeContractItems(extra.missing_contracts)
  const blockedContracts = normalizeContractItems(extra.blocked_contracts)
  return {
    status: 'PENDING',
    contract_id: CMDB_RESOURCE_APPROVAL_CONTRACT_ID,
    message: reason,
    view,
    expected_schema: extra.expected_schema || CMDB_RESOURCE_APPROVAL_EXPECTED_SCHEMA,
    blocked_contracts: blockedContracts,
    missing_contracts: missingContracts.length ? missingContracts : CMDB_RESOURCE_APPROVAL_MISSING_CONTRACTS,
    findx_audit_query: extra.findx_audit_query || {
      source: 'findx_audit',
      scope: 'cmdb',
      resource_type: 'cmdb_resource_approval',
      action: 'cmdb.resource_approval.read',
      view,
    },
  }
}

const compactParams = (params = {}) => Object.fromEntries(
  Object.entries(params).filter(([, value]) => value !== '' && value !== null && value !== undefined),
)

const asArray = value => Array.isArray(value) ? value : []
const text = (value, fallback = '') => value === null || value === undefined || value === '' ? fallback : String(value)
const numberOrNull = value => {
  const num = Number(value)
  return Number.isFinite(num) ? num : null
}
const pickFirst = (source, keys, fallback = undefined) => {
  for (const key of keys) {
    if (source?.[key] !== undefined && source?.[key] !== null) return source[key]
  }
  return fallback
}

const pickGraphPayload = raw => raw?.data?.data || raw?.data || raw?.graph || raw

const edgeId = (source, target, relationId, index) => `${source || 'source'}__${target || 'target'}__${relationId || index}`

const topologyContractProbeId = 'contract-probe'

const authHeaders = () => {
  const token = typeof window !== 'undefined' ? window.localStorage.getItem('aiw-token') : ''
  return token ? { Authorization: `Bearer ${token}` } : {}
}

const emitAuthExpired = () => {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT))
}

const getCmdbTopologyContract = async (instanceId) => {
  const id = encodeURIComponent(text(instanceId, topologyContractProbeId))
  const resp = await fetch(`/api/v1/cmdb/instances/${id}/topology`, {
    headers: authHeaders(),
  })
  let body = null
  try {
    body = await resp.json()
  } catch {
    body = {}
  }
  if (resp.status === 409 && body) return body
  if (resp.status === 401) emitAuthExpired()
  if (!resp.ok) {
    const error = new Error(redactText(body?.message || body?.error || `HTTP ${resp.status}: CMDB topology contract request failed`))
    error.status = resp.status
    throw error
  }
  return body
}

const getCmdbMonitorBindingsContract = async (instanceId) => {
  const id = encodeURIComponent(text(instanceId, topologyContractProbeId))
  const resp = await fetch(`/api/v1/cmdb/monitor-bindings/${id}`, {
    headers: authHeaders(),
  })
  let body = null
  try {
    body = await resp.json()
  } catch {
    body = {}
  }
  if (resp.status === 409 && body) return body
  if (resp.status === 401) emitAuthExpired()
  if (!resp.ok) {
    const error = new Error(redactText(body?.message || body?.error || `HTTP ${resp.status}: CMDB monitor binding request failed`))
    error.status = resp.status
    throw error
  }
  return body
}

const getCmdbResourceStatisticsContract = async (params = {}) => {
  const query = new URLSearchParams(compactParams(params)).toString()
  const resp = await fetch(`/api/v1/cmdb/resource-statistics${query ? `?${query}` : ''}`, {
    headers: authHeaders(),
  })
  let body = null
  try {
    body = await resp.json()
  } catch {
    body = {}
  }
  if (resp.status === 409 && body) return body
  if (resp.status === 401) emitAuthExpired()
  if (!resp.ok) {
    const error = new Error(redactText(body?.message || body?.error || `HTTP ${resp.status}: CMDB resource statistics contract request failed`))
    error.status = resp.status
    error.body = body
    throw error
  }
  return body
}

const getCmdbResourceApprovalsContract = async (view) => {
  const cleanView = normalizeApprovalView(view)
  const resp = await fetch(`/api/v1/cmdb/approvals?view=${encodeURIComponent(cleanView)}`, {
    headers: authHeaders(),
  })
  let body = null
  try {
    body = await resp.json()
  } catch {
    body = {}
  }
  if (resp.status === 409 && body) return body
  if (resp.status === 401) emitAuthExpired()
  if (!resp.ok) {
    const error = new Error(redactText(body?.message || body?.error || `HTTP ${resp.status}: CMDB resource approval contract request failed`))
    error.status = resp.status
    error.body = body
    throw error
  }
  return body
}

const getCmdbHostAISessionContract = async (hostId) => {
  const id = encodeURIComponent(text(hostId, ''))
  const resp = await fetch(`/api/v1/cmdb/hosts/${id}/ai-session`, {
    headers: authHeaders(),
  })
  let body = null
  try {
    body = await resp.json()
  } catch {
    body = {}
  }
  if (resp.status === 409 && body) return body
  if (resp.status === 401) emitAuthExpired()
  if (!resp.ok) {
    const error = new Error(redactText(body?.message || body?.error || `HTTP ${resp.status}: CMDB host AI session contract request failed`))
    error.status = resp.status
    error.body = body
    throw error
  }
  return body
}

const createCmdbHostAISessionRequest = async (hostId, body) => {
  const id = encodeURIComponent(text(hostId, ''))
  const resp = await fetch(`/api/v1/cmdb/hosts/${id}/ai-session`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(),
    },
    body: JSON.stringify(body || {}),
  })
  let payload = null
  try {
    payload = await resp.json()
  } catch {
    payload = {}
  }
  if (resp.status === 409 && payload) return payload
  if (resp.status === 401) emitAuthExpired()
  if (!resp.ok) {
    const error = new Error(redactText(payload?.message || payload?.error || `HTTP ${resp.status}: CMDB host AI session request failed`))
    error.status = resp.status
    error.body = payload
    throw error
  }
  return payload
}

const createCmdbRelationActionRequest = async (body) => {
  const resp = await fetch('/api/v1/cmdb/relation-actions', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(),
    },
    body: JSON.stringify(body || {}),
  })
  let payload = null
  try {
    payload = await resp.json()
  } catch {
    payload = {}
  }
  if (resp.status === 401) emitAuthExpired()
  if ((resp.status === 409 || resp.status === 404) && payload) return payload
  if (!resp.ok) {
    const error = new Error(redactText(payload?.message || payload?.error || `HTTP ${resp.status}: CMDB relation action request failed`))
    error.status = resp.status
    throw error
  }
  return payload
}

const normalizeMonitorBindings = (raw) => {
  const auditSource = raw && typeof raw === 'object' ? raw : {}
  const code = String(auditSource.status || auditSource.code || auditSource.error_code || '').toLowerCase()
  const expected_schema = auditSource.expected_schema
  const field_matrix = auditSource.field_matrix
  const source_evidence = auditSource.source_evidence
  const missing_contracts = auditSource.missing_contracts
  if (code === 'blocked_by_contract' || code === 'blocked') {
    return relationBlockedAudit(auditSource.message || '监控绑定契约阻断；未拿到真实绑定数据前只显示审计态。', { expected_schema, field_matrix, source_evidence, missing_contracts })
  }
  const bindings = asArray(auditSource.bindings)
  if (!bindings.length) {
    return relationBlockedAudit('监控绑定接口未返回真实 bindings；空列表只能保持契约审计态。', { expected_schema, field_matrix, source_evidence, missing_contracts })
  }
  return {
    status: 'ready',
    bindings,
    expected_schema: expected_schema || CMDB_RELATION_GRAPH_EXPECTED_SCHEMA,
    field_matrix: field_matrix || CMDB_RELATION_FIELD_MATRIX,
    source_evidence: source_evidence || CMDB_RELATION_GRAPH_EVIDENCE,
    missing_contracts: asArray(missing_contracts),
    audit_context: auditSource.audit_context,
    log_query: auditSource.log_query,
  }
}

const normalizeRelationAction = (raw, fallback = {}) => {
  const payload = raw && typeof raw === 'object' ? raw : {}
  const request = payload.action_request || null
  const code = String(payload.status || payload.code || payload.error || '').toLowerCase()
  const missingContracts = asArray(payload.missing_contracts || fallback.missing_contracts)
  if (request && (code === 'recorded' || payload.code === 0)) {
    return {
      ...fallback,
      status: 'recorded',
      request,
      audit_ref: payload.audit_ref || request.audit_ref,
      missing_contracts: missingContracts,
      expected_schema: payload.expected_schema || fallback.expected_schema,
      field_matrix: payload.field_matrix || fallback.field_matrix,
      log_query: payload.log_query || fallback.log_query,
      message: '关系动作请求已写入 FindX 审计；投递和生效回执仍按契约阻断。',
    }
  }
  return {
    ...fallback,
    status: 'PENDING',
    message: payload.message || fallback.message || '关系动作请求被 CMDB 契约阻断。',
    missing_contracts: missingContracts.length ? missingContracts : fallback.missing_contracts,
    expected_schema: payload.expected_schema || fallback.expected_schema,
    field_matrix: payload.field_matrix || fallback.field_matrix,
    log_query: payload.log_query || fallback.log_query,
  }
}

const walkTopologyTree = (node, parentId = '', depth = 1, nodes = [], edges = []) => {
  if (!node || typeof node !== 'object') return { nodes, edges }
  const object = node.object || {}
  const nodeId = text(object.id || node.id || node.object_id || node.name || `level-${depth}-${nodes.length}`)
  nodes.push({
    id: nodeId,
    name: text(node.name || object.name || nodeId),
    object_id: text(object.id || node.object_id || nodeId),
    object_name: text(object.name || node.object_name || node.name || nodeId),
    object_type: object.type || node.object_type || '',
    level: node.level || depth,
    kind: 'model',
    raw: node,
  })
  if (parentId) {
    const relation = asArray(node.relation?.in)[0] || asArray(node.relation?.out)[0] || {}
    edges.push({
      id: edgeId(parentId, nodeId, relation.relation_id, edges.length),
      source: parentId,
      target: nodeId,
      relation_id: text(relation.relation_id),
      relation_name: text(relation.asst_name || relation.name || '模型关系'),
      direction: relation.direction || '',
      location: text(relation.location),
      raw: relation,
    })
  }
  asArray(node.instances).forEach((instance, index) => {
    const instanceId = text(instance.id || `${nodeId}-instance-${index}`)
    nodes.push({
      id: instanceId,
      name: text(instance.name || instanceId),
      object_id: text(instance.object_id || nodeId),
      object_name: text(instance.object_name || node.name || object.name || nodeId),
      object_type: instance.object_type || object.type || '',
      level: (node.level || depth) + 0.35,
      kind: 'instance',
      raw: instance,
    })
    edges.push({
      id: edgeId(nodeId, instanceId, 'contains', edges.length),
      source: nodeId,
      target: instanceId,
      relation_id: 'contains',
      relation_name: '包含实例',
      direction: '',
      location: 'local',
      raw: instance,
    })
    asArray(instance.out).forEach((relation, relationIndex) => {
      if (!relation.related_instance_id) return
      edges.push({
        id: edgeId(instanceId, relation.related_instance_id, relation.relation_id, relationIndex),
        source: instanceId,
        target: relation.related_instance_id,
        relation_id: text(relation.relation_id),
        relation_name: text(relation.asst_name || '实例关系'),
        direction: relation.direction || '',
        location: 'out',
        raw: relation,
      })
    })
  })
  asArray(node.children).forEach(child => walkTopologyTree(child, nodeId, depth + 1, nodes, edges))
  return { nodes, edges }
}

const normalizeRelationGraph = (raw) => {
  const payload = pickGraphPayload(raw)
  const auditSource = payload && typeof payload === 'object' ? payload : {}
  const expected_schema = auditSource.expected_schema || raw?.expected_schema
  const field_matrix = auditSource.field_matrix || raw?.field_matrix
  const contract_matrix = auditSource.contract_matrix || raw?.contract_matrix
  const contract_gap_id = auditSource.contract_gap_id || raw?.contract_gap_id
  const source_evidence = auditSource.source_evidence || raw?.source_evidence
  const missing_contracts = auditSource.missing_contracts || raw?.missing_contracts
  const code = String(auditSource.status || auditSource.code || auditSource.error_code || '').toLowerCase()
  if (code === 'blocked_by_contract' || code === 'blocked') {
    return relationBlockedAudit(auditSource.message || '关系图后端返回阻断状态。', { contract_gap_id, contract_matrix, expected_schema, field_matrix, source_evidence, missing_contracts })
  }

  const graphNodes = asArray(auditSource.nodes)
  const graphEdges = asArray(auditSource.edges)
  const treeRoot = auditSource.object || auditSource.children || auditSource.instances ? auditSource : null
  const normalized = graphNodes.length || graphEdges.length
    ? { nodes: graphNodes, edges: graphEdges }
    : treeRoot ? walkTopologyTree(treeRoot) : { nodes: [], edges: [] }

  if (!normalized.nodes.length || !normalized.edges.length) {
    return relationBlockedAudit('关系图 API 未返回可验证的真实节点和关系边；空图、空 nodes/edges 或 loading 不能视为成功。', { contract_gap_id, contract_matrix, expected_schema, field_matrix, source_evidence, missing_contracts })
  }

  return {
    status: 'ready',
    nodes: normalized.nodes,
    edges: normalized.edges,
    relation_queries: asArray(auditSource.relation_queries || raw?.relation_queries),
    relation_query_rules: asArray(auditSource.relation_query_rules || raw?.relation_query_rules),
    relation_query_runtime: auditSource.relation_query_runtime || raw?.relation_query_runtime || null,
    expected_schema: expected_schema || CMDB_RELATION_GRAPH_EXPECTED_SCHEMA,
    field_matrix: field_matrix || CMDB_RELATION_FIELD_MATRIX,
    source_evidence: source_evidence || CMDB_RELATION_GRAPH_EVIDENCE,
    missing_contracts: asArray(missing_contracts),
  }
}

const normalizeStatRows = (rows, labelKeys = ['name', 'label', 'model_name', 'group_name', 'workspace_name', 'object_name'], valueKeys = ['count', 'total', 'value', 'resource_count']) => (
  asArray(rows)
    .map((row, index) => {
      const source = row && typeof row === 'object' ? row : { name: row }
      return {
        id: text(pickFirst(source, ['id', 'key', 'object_id', 'group_id', 'workspace_id', 'name', 'label'], `row-${index}`)),
        label: text(pickFirst(source, labelKeys, `维度 ${index + 1}`)),
        value: numberOrNull(pickFirst(source, valueKeys, 0)) ?? 0,
        raw: source,
      }
    })
    .filter(row => row.label || row.value)
)

const normalizeContractItems = (items) => (
  asArray(items)
    .map((item) => {
      if (typeof item === 'string') return { id: item, status: '' }
      if (!item || typeof item !== 'object') return null
      const id = text(item.id || item.contract_id || item.key || item.name)
      if (!id) return null
      return { id, status: text(item.status || item.state || item.reason) }
    })
    .filter(Boolean)
)

const safeApprovalText = (value, fallback = '') => {
  if (value === null || value === undefined || value === '') return fallback
  if (typeof value === 'object') return fallback
  return redactText(String(value)).slice(0, 240)
}

const compactApprovalParts = (parts) => parts
  .map(part => safeApprovalText(part))
  .filter(Boolean)
  .join(' / ')

const summarizeApprovalEntity = (value, keys = ['name', 'id']) => {
  if (value === null || value === undefined || value === '') return ''
  if (typeof value !== 'object') return safeApprovalText(value)
  return compactApprovalParts(keys.map(key => value[key]))
}

const normalizeApprovalRisk = (value) => {
  if (value === null || value === undefined || value === '') return ''
  if (typeof value !== 'object') return safeApprovalText(value)
  return compactApprovalParts([
    value.level,
    value.risk_level,
    value.name,
    value.label,
    value.summary,
    value.reason,
  ])
}

const normalizeApprovalRequest = (item, index) => {
  if (!item || typeof item !== 'object') return null
  const id = safeApprovalText(pickFirst(item, ['id', 'request_id', 'approval_id', 'approval_request_id', 'key'], `approval-${index + 1}`))
  const resource = summarizeApprovalEntity(pickFirst(item, ['resource', 'resource_ref', 'target', 'asset']), ['name', 'resource_name', 'instance_name', 'object_name', 'id', 'resource_id', 'instance_id', 'object_id'])
    || compactApprovalParts([
      pickFirst(item, ['resource_name', 'asset_name', 'instance_name', 'object_name']),
      pickFirst(item, ['resource_id', 'asset_id', 'instance_id', 'object_id']),
    ])
  const requester = summarizeApprovalEntity(pickFirst(item, ['requester', 'requested_by', 'applicant', 'creator']), ['name', 'username', 'display_name', 'id'])
    || safeApprovalText(pickFirst(item, ['requester_name', 'requested_by_name', 'applicant_name', 'created_by']))
  const approver = summarizeApprovalEntity(pickFirst(item, ['approver', 'current_approver', 'reviewer', 'assignee']), ['name', 'username', 'display_name', 'id'])
    || safeApprovalText(pickFirst(item, ['approver_name', 'reviewer_name', 'assignee_name']))
  const missingContracts = normalizeContractItems([
    ...asArray(item.missing_contracts),
    ...asArray(item.workflow_state?.missing_contracts),
    ...asArray(item.risk_policy?.missing_contracts),
    ...asArray(item.audit_receipt?.missing_contracts),
  ])
  return {
    id,
    title: safeApprovalText(pickFirst(item, ['title', 'name', 'subject'], `审批请求 ${index + 1}`)),
    summary: safeApprovalText(pickFirst(item, ['summary', 'description', 'reason', 'comment'])),
    resource: resource || '-',
    action: safeApprovalText(pickFirst(item, ['action', 'action_name', 'operation', 'operation_type', 'request_type']), '-'),
    risk: normalizeApprovalRisk(pickFirst(item, ['risk', 'risk_level', 'risk_summary', 'risk_policy'])) || '-',
    status: safeApprovalText(pickFirst(item, ['status', 'state', 'workflow_status']), 'pending_review'),
    requester: requester || '-',
    approver: approver || '-',
    audit_ref: safeApprovalText(pickFirst(item, ['audit_ref', 'audit_id', 'receipt_id', 'audit_receipt_id'], item.audit_receipt?.audit_ref || item.audit_receipt?.id), '-'),
    risk_record_id: safeApprovalText(pickFirst(item, ['risk_record_id', 'risk_id', 'risk_ref']), '-'),
    created_at: safeApprovalText(pickFirst(item, ['created_at', 'createdAt', 'request_time', 'submitted_at'])),
    updated_at: safeApprovalText(pickFirst(item, ['updated_at', 'updatedAt', 'reviewed_at', 'decided_at'])),
    missing_contracts: missingContracts,
  }
}

const normalizeResourceStats = (raw) => {
  const payload = raw?.data?.data || raw?.data || raw?.statistics || raw
  const source = payload && typeof payload === 'object' ? payload : {}
  const code = String(source.status || source.code || source.error_code || raw?.status || '').toLowerCase()
  const blockedContracts = normalizeContractItems(source.blocked_contracts || raw?.blocked_contracts)
  const missingContracts = normalizeContractItems(source.missing_contracts || raw?.missing_contracts)
  const expectedSchema = source.expected_schema || raw?.expected_schema
  const sourceEvidence = source.source_evidence || raw?.source_evidence
  const auditContext = source.audit_context || raw?.audit_context

  if (code === 'blocked_by_contract' || code === 'blocked') {
    return resourceStatsBlockedAudit(source.message || raw?.message || '资源统计契约阻断，当前未拿到可核验的聚合数据。', {
      blocked_contracts: blockedContracts,
      missing_contracts: missingContracts,
      expected_schema: expectedSchema,
      source_evidence: sourceEvidence,
      audit_context: auditContext,
    })
  }

  const totalsSource = source.totals || source.summary || source.overview || source.aggregate || source
  const summary = {
    cmdb_instances: numberOrNull(pickFirst(totalsSource, ['cmdb_instances', 'total_resources', 'resource_total', 'total', 'resource_count'])),
    cmdb_models: numberOrNull(pickFirst(totalsSource, ['cmdb_models', 'model_count', 'models', 'object_count'])),
    host_assets: numberOrNull(pickFirst(totalsSource, ['host_assets', 'hosts', 'monitor_targets'])),
    findx_agents: numberOrNull(pickFirst(totalsSource, ['findx_agents', 'agents'])),
    monitor_binding_rows: numberOrNull(pickFirst(totalsSource, ['monitor_binding_rows', 'monitor_bindings'])),
    monitor_bound_instances: numberOrNull(pickFirst(totalsSource, ['monitor_bound_instances', 'bound_instances'])),
    relation_edges: numberOrNull(pickFirst(totalsSource, ['relation_edges', 'relations'])),
    resource_groups: numberOrNull(pickFirst(totalsSource, ['resource_groups'])),
    topology_businesses: numberOrNull(pickFirst(totalsSource, ['topology_businesses', 'workspaces', 'business_groups'])),
  }
  const dimensions = {
    by_model: normalizeStatRows(source.model_distribution || source.by_model || source.models, ['name', 'model_name', 'object_name', 'label', 'key']),
    by_group: normalizeStatRows(source.resource_group_distribution || source.by_group || source.groups || source.group_distribution, ['name', 'group_name', 'label', 'key']),
    by_workspace: normalizeStatRows(source.business_group_distribution || source.by_workspace || source.workspaces || source.workspace_distribution, ['name', 'workspace_name', 'label', 'key']),
    change_trend: normalizeStatRows(source.change_trend || source.trend || source.changes, ['date', 'day', 'time', 'label'], ['count', 'changed_count', 'value']),
    top_changed: normalizeStatRows(source.top_changed || source.change_top || source.top5, ['name', 'instance_name', 'object_name', 'label'], ['count', 'changed_count', 'value']),
  }
  const hasSummary = Object.values(summary).some(value => value !== null)
  const hasDimensions = Object.values(dimensions).some(rows => rows.length > 0)
  if (!hasSummary && !hasDimensions) {
    return resourceStatsBlockedAudit('资源统计接口未返回 summary 或维度明细；当前只展示契约缺口，不渲染空统计。', {
      blocked_contracts: blockedContracts,
      missing_contracts: missingContracts,
      expected_schema: expectedSchema,
      source_evidence: sourceEvidence,
      audit_context: auditContext,
    })
  }

  return {
    status: 'available',
    contract: source.contract || raw?.contract || 'cmdb.resource.statistics.read.v1',
    raw_status: source.status || raw?.status || '',
    message: blockedContracts.length ? '资源统计已读取真实聚合；以下运行时能力仍按契约缺口展示。' : '',
    summary,
    dimensions,
    blocked_contracts: blockedContracts,
    missing_contracts: missingContracts,
    expected_schema: expectedSchema || CMDB_RESOURCE_STATS_EXPECTED_SCHEMA,
    source_evidence: sourceEvidence || CMDB_RESOURCE_STATS_EVIDENCE,
    audit_context: auditContext,
    findx_audit_query: source.findx_audit_query || source.log_query || raw?.findx_audit_query,
    updated_at: source.updated_at || source.generated_at || source.timestamp || '',
  }
}

const normalizeResourceApprovals = (raw, view) => {
  const payload = raw?.data?.data || raw?.data || raw
  const source = payload && typeof payload === 'object' ? payload : {}
  const code = String(source.status || source.code || source.error_code || raw?.status || '').toLowerCase()
  const cleanView = normalizeApprovalView(source.view || view)
  const missingContracts = source.missing_contracts || raw?.missing_contracts
  const blockedContracts = source.blocked_contracts || raw?.blocked_contracts
  const expectedSchema = source.expected_schema || raw?.expected_schema
  const findxAuditQuery = source.findx_audit_query || source.log_query || raw?.findx_audit_query
  const approvalRequests = source.approval_requests
  const contractGaps = {
    workflow: normalizeContractItems(source.workflow_state?.missing_contracts || source.workflow_contract_gaps || source.workflow_gaps),
    risk: normalizeContractItems(source.risk_policy?.missing_contracts || source.risk_contract_gaps || source.risk_gaps),
    audit: normalizeContractItems(source.audit_receipt?.missing_contracts || source.audit_contract_gaps || source.audit_gaps),
  }

  if (code === 'blocked_by_contract' || code === 'blocked') {
    return resourceApprovalBlockedAudit(source.message || raw?.message || '资源审批 runtime 契约阻断，当前只展示后端审计缺口。', {
      view: cleanView,
      missing_contracts: missingContracts,
      blocked_contracts: blockedContracts,
      expected_schema: expectedSchema,
      findx_audit_query: findxAuditQuery,
    })
  }

  if (!Array.isArray(approvalRequests) || approvalRequests.length === 0) {
    return resourceApprovalBlockedAudit('资源审批接口未返回非空 approval_requests；当前按 fail-close 只展示契约审计，不把空列表视为成功。', {
      view: cleanView,
      missing_contracts: missingContracts,
      blocked_contracts: blockedContracts,
      expected_schema: expectedSchema,
      findx_audit_query: findxAuditQuery,
    })
  }

  const list = approvalRequests
    .map(normalizeApprovalRequest)
    .filter(Boolean)

  if (!list.length) {
    return resourceApprovalBlockedAudit('资源审批接口返回的 approval_requests 不符合前端白名单字段；当前按 fail-close 展示契约审计。', {
      view: cleanView,
      missing_contracts: missingContracts,
      blocked_contracts: blockedContracts,
      expected_schema: expectedSchema,
      findx_audit_query: findxAuditQuery,
    })
  }

  return {
    status: 'available',
    mode: 'list',
    available: true,
    view: cleanView,
    contract_id: CMDB_RESOURCE_APPROVAL_CONTRACT_ID,
    raw_status: safeApprovalText(source.status || raw?.status || ''),
    message: safeApprovalText(source.message || raw?.message || ''),
    list,
    approval_requests: list,
    blocked_contracts: normalizeContractItems(blockedContracts),
    missing_contracts: normalizeContractItems(missingContracts),
    contract_gaps: contractGaps,
    expected_schema: expectedSchema || CMDB_RESOURCE_APPROVAL_EXPECTED_SCHEMA,
    findx_audit_query: findxAuditQuery || resourceApprovalBlockedAudit('', { view: cleanView }).findx_audit_query,
  }
}

const hostAiSessionBlockedAudit = (reason, extra = {}) => ({
  status: 'PENDING',
  contract_id: CMDB_HOST_AI_SESSION_CONTRACT_ID,
  message: reason,
  host_context: extra.host_context || {},
  preflight: extra.preflight || {
    mode: 'host_ai_diagnosis',
    remote_command: 'blocked',
    tool_invocation: 'blocked',
    message_transport: 'blocked',
    output_receipt: 'blocked',
    readonly_context_only: true,
  },
  request_preview: extra.request_preview || null,
  missing_contracts: normalizeContractItems(extra.missing_contracts).length
    ? normalizeContractItems(extra.missing_contracts)
    : CMDB_HOST_AI_SESSION_MISSING_CONTRACTS.map(id => ({ id, status: '' })),
  blocked_contracts: normalizeContractItems(extra.blocked_contracts),
  findx_audit_query: extra.findx_audit_query || {
    source: 'findx_audit',
    scope: 'cmdb',
    resource_type: 'cmdb_host_ai_session',
    action: 'cmdb.host_ai_session.preflight',
  },
})

const normalizeHostAiSession = (raw) => {
  const payload = raw?.data?.data || raw?.data || raw
  const source = payload && typeof payload === 'object' ? payload : {}
  const code = String(source.status || source.code || source.error_code || raw?.status || '').toLowerCase()
  if (code === 'blocked_by_contract' || code === 'blocked') {
    return hostAiSessionBlockedAudit(source.message || raw?.message || '单机诊断对话 runtime 契约阻断，当前只展示主机上下文和审计缺口。', {
      host_context: source.host_context || raw?.host_context,
      preflight: source.preflight || raw?.preflight,
      request_preview: source.request_preview || raw?.request_preview,
      missing_contracts: source.missing_contracts || raw?.missing_contracts,
      blocked_contracts: source.blocked_contracts || raw?.blocked_contracts,
      findx_audit_query: source.findx_audit_query || raw?.findx_audit_query,
    })
  }
  return hostAiSessionBlockedAudit('单机诊断对话接口未返回可审计阻断契约；当前不建立真实会话。', {
    host_context: source.host_context,
    missing_contracts: source.missing_contracts,
    blocked_contracts: source.blocked_contracts,
    findx_audit_query: source.findx_audit_query,
  })
}

export const CMDB_EXECUTION_BLOCKERS = {
  terminal: 'PENDING: 远程终端缺少会话授权、审计记录、执行回执和安全隔离契约，当前只能显示阻断状态。',
  upload: 'PENDING: 文件上传缺少 SFTP/执行器、目标路径校验、审计记录、回滚和执行回执契约，当前只能显示阻断状态。',
  exec: 'PENDING: 命令执行缺少远程执行器、权限审批、审计记录、超时回执和输出脱敏契约，当前只能显示阻断状态。',
  deploy: 'PENDING: 部署任务缺少执行器、任务回执、日志轮询、回滚和 Evidence Chain 契约，当前只能显示阻断状态。',
  databaseTest: 'PENDING: 数据库连接测试缺少凭据引用、执行器、审计记录和脱敏错误模型，当前只能显示阻断状态。',
}

export const isCmdbContractBlocked = (error) => {
  const message = String(error?.message || error || '')
  return error?.status === 409 || /PENDING|HTTP 409/i.test(message)
}

export const cmdbContractMessage = (error, fallback) => {
  const message = String(error?.message || '')
  if (/PENDING/i.test(message)) return message.replace(/^HTTP 409:\s*/i, '')
  return fallback
}

export const isCmdbBlockedRecord = (record) => {
  const code = String(record?.code || record?.status || record?.error_code || '').toLowerCase()
  return code === 'blocked_by_contract' || code === 'blocked'
}

export const cmdbBlockedRecordMessage = (record, fallback) => {
  const missing = Array.isArray(record?.missing_contracts) ? record.missing_contracts.join('、') : ''
  if (record?.contract_id) {
    return `${fallback} contract_id=${record.contract_id}${missing ? `；缺口：${missing}` : ''}`
  }
  return fallback
}

export const isCmdbContractAudit = (record) => {
  const code = String(record?.status || record?.code || record?.error_code || '').toLowerCase()
  return code === 'blocked_by_contract' || code === 'blocked'
}

export const cmdbApi = {
  tree: () => get('/cmdb/tree'),
  objects: {
    list: (categoryId) => get('/cmdb/objects', { params: categoryId ? { category_id: categoryId } : {} }),
    get: (id) => get(`/cmdb/objects/${id}`),
    create: (body) => post('/cmdb/objects', body),
    update: (id, body) => put(`/cmdb/objects/${id}`, body),
    remove: (id) => del(`/cmdb/objects/${id}`),
  },
  attributes: {
    list: (objectId) => get(`/cmdb/objects/${objectId}/attributes`),
    create: (objectId, body) => post(`/cmdb/objects/${objectId}/attributes`, body),
    update: (id, body) => put(`/cmdb/attributes/${id}`, body),
    remove: (id) => del(`/cmdb/attributes/${id}`),
  },
  instances: {
    list: (objectId, params) => get(`/cmdb/objects/${objectId}/instances`, { params }),
    compatibleList: (objectId, params) => get(`/cmdb/objects/${objectId}/instances-compatible`, { params }),
    get: (id) => get(`/cmdb/instances/${id}`),
    compatibleDetail: (id) => get(`/cmdb/instances/${id}/detail-compatible`),
    create: (objectId, body) => post(`/cmdb/objects/${objectId}/instances`, body),
    update: (id, body) => put(`/cmdb/instances/${id}`, body),
    remove: (id) => del(`/cmdb/instances/${id}`),
  },
  // C03: SSH 终端
  terminal: {
    wsUrl: (hostId) => {
      const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      return `${proto}//${window.location.host}/api/v1/cmdb/hosts/${encodeURIComponent(hostId)}/terminal`
    },
  },
  // C04: 文件上传
  upload: (hostId, formData) => post(`/cmdb/hosts/${encodeURIComponent(hostId)}/upload`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000,
  }),
  // C05: 命令执行
  exec: (hostId, body) => post(`/cmdb/hosts/${encodeURIComponent(hostId)}/exec`, body),
  hostAiSession: {
    preflight: async (hostId) => {
      try {
        return normalizeHostAiSession(await getCmdbHostAISessionContract(hostId))
      } catch (error) {
        return hostAiSessionBlockedAudit(cmdbContractMessage(error, '单机诊断对话 runtime 后端契约未开放，当前不建立真实会话。'), {
          missing_contracts: error?.body?.missing_contracts,
          blocked_contracts: error?.body?.blocked_contracts,
          findx_audit_query: error?.body?.findx_audit_query,
        })
      }
    },
    request: async (hostId, body) => {
      try {
        return normalizeHostAiSession(await createCmdbHostAISessionRequest(hostId, body))
      } catch (error) {
        return hostAiSessionBlockedAudit(cmdbContractMessage(error, '单机诊断对话请求被 runtime 契约阻断，当前不建立真实会话。'), {
          missing_contracts: error?.body?.missing_contracts,
          blocked_contracts: error?.body?.blocked_contracts,
          findx_audit_query: error?.body?.findx_audit_query,
        })
      }
    },
  },
  // C07: 数据库资产
  databases: {
    list: (params) => get('/cmdb/databases', { params }),
    create: (body) => post('/cmdb/databases', body),
    get: (id) => get(`/cmdb/databases/${encodeURIComponent(id)}`),
    remove: (id) => del(`/cmdb/databases/${encodeURIComponent(id)}`),
    test: (id) => post(`/cmdb/databases/${encodeURIComponent(id)}/test`),
  },
  // C09: 部署任务
  deployTasks: {
    list: () => get('/cmdb/deploy-tasks'),
    create: (body) => post('/cmdb/deploy-tasks', body),
    get: (id) => get(`/cmdb/deploy-tasks/${encodeURIComponent(id)}`),
  },
  // C10: 导入
  import: {
    excel: (formData) => post('/cmdb/import/excel', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),
    confirm: (body) => post('/cmdb/import/confirm', body),
    cloud: (body) => post('/cmdb/import/cloud', body),
  },
  resourceStats: async (params) => {
    try {
      return normalizeResourceStats(await getCmdbResourceStatisticsContract(params))
    } catch (error) {
      if (isCmdbContractBlocked(error) || [404, 405, 501].includes(error?.status)) {
        return resourceStatsBlockedAudit(cmdbContractMessage(error, '资源统计后端契约未开放，无法确认真实聚合数据。'))
      }
      throw error
    }
  },
  resourceApprovals: async (view) => {
    const cleanView = normalizeApprovalView(view)
    try {
      return normalizeResourceApprovals(await getCmdbResourceApprovalsContract(cleanView), cleanView)
    } catch (error) {
      return resourceApprovalBlockedAudit(cmdbContractMessage(error, '资源审批 runtime 后端契约未开放，无法确认真实审批数据。'), {
        view: cleanView,
        missing_contracts: error?.body?.missing_contracts,
        blocked_contracts: error?.body?.blocked_contracts,
        expected_schema: error?.body?.expected_schema,
        findx_audit_query: error?.body?.findx_audit_query,
      })
    }
  },
  relations: {
    graph: async (params) => {
      try {
        const clean = compactParams(params)
        const instanceId = clean.instance_id || clean.instanceId || clean.id || topologyContractProbeId
        return normalizeRelationGraph(await getCmdbTopologyContract(instanceId))
      } catch (error) {
        if (isCmdbContractBlocked(error) || [404, 405, 501].includes(error?.status)) {
          return relationBlockedAudit(cmdbContractMessage(error, '关系图后端契约未开放，无法确认来源证据中的拓扑数据。'))
        }
        throw error
      }
    },
    monitorBindings: async (params) => {
      try {
        const clean = compactParams(params)
        const instanceId = clean.instance_id || clean.instanceId || clean.id || topologyContractProbeId
        return normalizeMonitorBindings(await getCmdbMonitorBindingsContract(instanceId))
      } catch (error) {
        if (isCmdbContractBlocked(error) || [404, 405, 501].includes(error?.status)) {
          return relationBlockedAudit(cmdbContractMessage(error, '监控绑定后端契约未开放，无法确认真实绑定数据。'))
        }
        throw error
      }
    },
    actionRequest: async (body, fallback) => {
      try {
        return normalizeRelationAction(await createCmdbRelationActionRequest(body), fallback)
      } catch (error) {
        if (isCmdbContractBlocked(error) || [404, 405, 501].includes(error?.status)) {
          return normalizeRelationAction({ status: 'PENDING', message: cmdbContractMessage(error, '关系动作后端契约未开放，无法记录真实动作请求。') }, fallback)
        }
        throw error
      }
    },
  },
}
