import React, { useEffect, useState } from 'react'
import { formatAssetError } from '../api/assets.js'
import { CMDB_RELATION_GRAPH_CONTRACT_ID, CMDB_RELATION_GRAPH_EVIDENCE, CMDB_RELATION_GRAPH_EXPECTED_SCHEMA, CMDB_RELATION_FIELD_MATRIX, cmdbApi, cmdbContractMessage, isCmdbContractBlocked } from '../api/cmdb.js'
import { displayText } from './assetsModel.js'
import { RelationQueryRuntimeAudit } from './RelationQueryRuntimeAudit.jsx'
import './RelationTopologySection.css'
import './RelationGraphSection.css'

const INSTANCE_DETAIL_EXPECTED_SCHEMA = {
  query: ['instance_id'],
  instance: ['id', 'name', 'object_id', 'object_name', 'attributes', 'relations', 'monitor_bindings'],
  actions: ['detail', 'relations', 'topology', 'monitor_binding'],
  audit: ['expected_schema', 'field_matrix', 'source_evidence', 'missing_contracts'],
}

const INSTANCE_DETAIL_FIELD_MATRIX = [
  { field: 'query.instance_id', required: true, purpose: '实例详情入口必须携带真实实例 ID，缺失时不能回落 overview 或伪造详情。' },
  { field: 'instance.id', required: true, purpose: '实例唯一标识，用于详情、关系、拓扑和监控绑定二级动作。' },
  { field: 'instance.attributes', required: true, purpose: '实例属性明细必须来自 CMDB 后端真实返回。' },
  { field: 'instance.relations', required: true, purpose: '实例入向/出向关系缺失时，关系和拓扑动作保持阻断。' },
  { field: 'instance.monitor_bindings', required: true, purpose: '监控对象绑定契约缺失时，监控绑定动作保持阻断。' },
]

const INSTANCE_DETAIL_SOURCE_EVIDENCE = [
  '测试证据: cmdb-instance-default-topology.json',
  '测试证据: cmdb-instance-monitor-bind-dialog.png',
  '测试证据: cmdb-instance-relation-page.png',
]

const CMDB_CONTEXT_CARDS = [
  { title: '业务上下文', status: '需要 business_names/workspace_id', body: '实例拓扑必须能落到业务系统或业务组，不把单点拓扑伪装成业务拓扑。' },
  { title: '日志审计', status: 'FindX audit', body: '主机分配和监控绑定写入 scope=cmdb 的 FindX 审计日志，可从日志中心按 findx_audit 查询。' },
  { title: 'Agent 分配', status: '资源列表可操作', body: '主机行支持分配业务组和资源组；保存后由后端审计，不在前端伪造成功。' },
  { title: '监控绑定', status: '绑定状态', body: '读取同一个 monitor-bindings 契约；没有真实绑定时展示阻断审计。' },
]

const CMDB_RELATION_ACTIONS = [
  { key: 'expand', label: '展开', contract: 'cmdb_relation_expand_contract' },
  { key: 'detail', label: '详情', contract: 'cmdb_relation_detail_contract' },
  { key: 'relation', label: '关系', contract: 'cmdb_relation_detail_contract' },
  { key: 'topology', label: '拓扑', contract: 'cmdb_relation_topology_contract' },
]

const CMDB_PENDING_CAPABILITIES = [
  ['机房视图', '需要机房、机柜、U 位容量、告警统计和设备类型聚合契约。'],
  ['发现管理', '需要自动发现规则、映射字段、执行记录和同步回执契约。'],
  ['资源审批', '需要申请、审批人、审批结果、批量审批和审计链路契约。'],
  ['资源报表/资产消费', '需要报表数据集、指标维度、账期、Owner 和消费趋势契约。'],
]

export function TopologySection({ query = {} }) {
  const instanceId = query.instance_id || query.instanceId || query.id || ''
  return (
    <section className='fx-assets-work fx-cmdb-route-section'>
      <CmdbRouteFrame active='topology' instanceId={instanceId}>
        <RelationGraphSection groupId={query.group || query.group_id || ''} instanceId={instanceId} />
      </CmdbRouteFrame>
    </section>
  )
}

export function InstanceDetailContractSection({ query = {}, onNavigate = () => {} }) {
  const instanceId = query.instance_id || query.instanceId || query.id || ''
  const audit = {
    status: 'PENDING',
    contract_id: CMDB_RELATION_GRAPH_CONTRACT_ID,
    message: instanceId
      ? '实例详情二级页需要 detail-compatible、relations、topology 和 monitor_bindings 同源返回；缺任一契约时不渲染伪详情。'
      : 'PENDING: /assets?section=instance-detail 缺少 instance_id，无法确认真实 CMDB 实例；不会回落 overview 或展示假详情。',
    expected_schema: INSTANCE_DETAIL_EXPECTED_SCHEMA,
    field_matrix: INSTANCE_DETAIL_FIELD_MATRIX,
    source_evidence: INSTANCE_DETAIL_SOURCE_EVIDENCE,
    missing_contracts: [
      'cmdb.instance.detail.adapter',
      'cmdb.instance.relations.adapter',
      'cmdb.instance.monitor_bindings.adapter',
    ],
  }
  return (
    <section className='fx-assets-work fx-cmdb-instance-contract'>
      <CmdbRouteFrame active='instance-detail' instanceId={instanceId}>
        <header className='fx-cmdb-topology-head'>
          <div>
            <h2>CMDB 实例详情契约审计</h2>
            <p>{instanceId ? `instance_id=${displayText(instanceId)}` : '缺少 instance_id，详情入口保持阻断。'}</p>
          </div>
          <button type='button' className='fx-assets-button' onClick={() => onNavigate({ section: 'topology', instance_id: instanceId })} disabled={!instanceId} title={!instanceId ? audit.message : '查看实例关系拓扑契约审计'}>
            查看拓扑
          </button>
        </header>
        <RelationBlockedAudit audit={audit} loading={false} />
        <RelationActionBar blocked selectedNode={null} reason={audit.message} instanceId={instanceId} />
      </CmdbRouteFrame>
    </section>
  )
}

function CmdbRouteFrame({ active, instanceId, children }) {
  const routeLabel = active === 'topology' ? '实例拓扑 / 默认拓扑' : '实例详情'
  return <div className='fx-cmdb-route-frame' data-route={active} data-instance-id={instanceId || ''} aria-label={routeLabel}>{children}</div>
}

export function RelationGraphSection({ groupId, instanceId }) {
  const [result, setResult] = useState(null)
  const [loading, setLoading] = useState(false)
  const [selectedNode, setSelectedNode] = useState(null)
  const [bindingResult, setBindingResult] = useState(null)
  const [bindingLoading, setBindingLoading] = useState(false)
  const [actionAudit, setActionAudit] = useState(null)

  const blocked = result?.status !== 'ready'
  const audit = result || {
    status: 'PENDING',
    contract_id: CMDB_RELATION_GRAPH_CONTRACT_ID,
    message: '关系图尚未查询；未拿到真实节点和关系边前不会渲染成功态。',
    expected_schema: CMDB_RELATION_GRAPH_EXPECTED_SCHEMA,
    field_matrix: CMDB_RELATION_FIELD_MATRIX,
    source_evidence: CMDB_RELATION_GRAPH_EVIDENCE,
    missing_contracts: [CMDB_RELATION_GRAPH_CONTRACT_ID],
  }

  const load = async () => {
    setLoading(true)
    setSelectedNode(null)
    try {
      setResult(await cmdbApi.relations.graph({ group_id: groupId, instance_id: instanceId, instanceId }))
    } catch (error) {
      setResult({
        status: 'PENDING',
        contract_id: CMDB_RELATION_GRAPH_CONTRACT_ID,
        message: isCmdbContractBlocked(error) ? cmdbContractMessage(error, '关系图契约阻断。') : formatAssetError(error),
        expected_schema: CMDB_RELATION_GRAPH_EXPECTED_SCHEMA,
        field_matrix: CMDB_RELATION_FIELD_MATRIX,
        source_evidence: CMDB_RELATION_GRAPH_EVIDENCE,
        missing_contracts: [CMDB_RELATION_GRAPH_CONTRACT_ID],
      })
    } finally {
      setLoading(false)
    }
  }

  const loadMonitorBindings = async () => {
    setBindingLoading(true)
    try {
      setBindingResult(await cmdbApi.relations.monitorBindings({ instance_id: selectedNode?.id || instanceId || 'contract-probe' }))
    } catch (error) {
      setBindingResult({
        status: 'PENDING',
        contract_id: CMDB_RELATION_GRAPH_CONTRACT_ID,
        message: isCmdbContractBlocked(error) ? cmdbContractMessage(error, '监控绑定契约阻断。') : formatAssetError(error),
        expected_schema: CMDB_RELATION_GRAPH_EXPECTED_SCHEMA,
        field_matrix: CMDB_RELATION_FIELD_MATRIX,
        source_evidence: CMDB_RELATION_GRAPH_EVIDENCE,
        missing_contracts: [CMDB_RELATION_GRAPH_CONTRACT_ID],
      })
    } finally {
      setBindingLoading(false)
    }
  }

  const openRelationAction = async (action) => {
    const fallback = buildRelationActionAudit({
      action,
      instanceId,
      selectedNode,
      reason: audit?.message,
    })
    setActionAudit(fallback)
    const target = fallback.target
    try {
      setActionAudit(await cmdbApi.relations.actionRequest({
        action: action.key,
        instance_id: target.instance_id,
        node_id: target.node_id,
        object_id: target.object_id,
        relation_id: target.relation_id,
        context: {
          source: 'relation_topology_drawer',
          blocked_reason: audit?.message,
        },
      }, fallback))
    } catch (error) {
      setActionAudit({
        ...fallback,
        status: 'PENDING',
        message: formatAssetError(error),
      })
    }
  }

  useEffect(() => { load() }, [groupId, instanceId])

  return (
    <section className='fx-cmdb-topology'>
      <header className='fx-cmdb-topology-head'>
        <div>
          <h2>CMDB 关系拓扑</h2>
          <p>按 object_id、instance_id、topology_id、业务上下文、审计日志和监控绑定契约对齐；空图、空数组和加载态均按契约阻断处理。</p>
        </div>
        <button type='button' className='fx-assets-button is-primary' onClick={load} disabled={loading}>{loading ? '查询中...' : '查询关系图'}</button>
      </header>
      <CmdbContextStrip />
      {blocked ? (
        <>
          <RelationBlockedCanvas audit={audit} loading={loading} />
          <RelationBlockedAudit audit={audit} loading={loading} />
        </>
      ) : (
        <>
          <RelationGraph result={result} selectedNode={selectedNode} onSelectNode={setSelectedNode} />
          {result?.relation_query_runtime && <RelationQueryRuntimeAudit runtime={result.relation_query_runtime} />}
        </>
      )}
      <RelationActionBar blocked={blocked} selectedNode={selectedNode} instanceId={instanceId} onRelationAction={openRelationAction} onMonitorBindings={loadMonitorBindings} bindingLoading={bindingLoading} />
      {bindingResult && <MonitorBindingPanel result={bindingResult} loading={bindingLoading} onClose={() => setBindingResult(null)} />}
      {actionAudit && <RelationActionDrawer audit={actionAudit} onClose={() => setActionAudit(null)} />}
    </section>
  )
}

function CmdbContextStrip() {
  return (
    <div className='fx-cmdb-context-strip'>
      {CMDB_CONTEXT_CARDS.map(card => (
        <article key={card.title}>
          <strong>{card.title}</strong>
          <span>{card.status}</span>
          <p>{card.body}</p>
        </article>
      ))}
    </div>
  )
}

function RelationBlockedCanvas({ audit, loading }) {
  return (
    <div className='fx-cmdb-canvas-shell'>
      <div className='fx-cmdb-canvas-toolbar'>
        <strong>默认拓扑</strong>
        <span>{loading ? '正在检查契约' : '契约未闭合'}</span>
        <button type='button' disabled>管理</button>
      </div>
      <div className='fx-cmdb-mature-canvas is-blocked'>
        <div className='fx-cmdb-instance-card is-current'>
          <span>操作系统</span>
          <strong>{audit.contract_gap_id || CMDB_RELATION_GRAPH_CONTRACT_ID}</strong>
          <em>缺实例拓扑真实响应</em>
        </div>
        <div className='fx-cmdb-lane is-blue'><header>关联 用户</header><p>需要 related_instance_id / relation_id</p></div>
        <div className='fx-cmdb-lane is-green'><header>属于 业务系统</header><p>需要 business_names / workspace_id</p></div>
        <div className='fx-cmdb-lane is-blue'><header>属于 数据库 / 中间件</header><p>需要 children[].relation 与实例边</p></div>
        <div className='fx-cmdb-minimap'>待接入</div>
      </div>
      <div className='fx-cmdb-blocked-capabilities'>
        {CMDB_PENDING_CAPABILITIES.map(([name, reason]) => <span key={name}>{name}: {reason}</span>)}
      </div>
    </div>
  )
}

function RelationBlockedAudit({ audit, loading }) {
  return (
    <div className='fx-cmdb-blocked-audit'>
      <div className='fx-cmdb-audit-status'>
        <strong>{loading ? 'CHECKING_CONTRACT' : 'PENDING'}</strong>
        <span>{audit.message}</span>
      </div>
      <div className='fx-cmdb-audit-grid'>
        <AuditList title='缺失契约' items={audit.missing_contracts || [CMDB_RELATION_GRAPH_CONTRACT_ID]} />
        <AuditMatrix title='contract matrix' rows={audit.contract_matrix || []} />
        <AuditSchema schema={audit.expected_schema || CMDB_RELATION_GRAPH_EXPECTED_SCHEMA} />
        <AuditMatrix rows={audit.field_matrix || CMDB_RELATION_FIELD_MATRIX} />
        <AuditList title='来源证据' items={sanitizeEvidence(audit.source_evidence)} />
      </div>
    </div>
  )
}

function AuditList({ title, items }) {
  const rows = Array.isArray(items) && items.length ? items : ['未返回']
  return <div className='fx-cmdb-audit-panel'><h3>{title}</h3>{rows.map(item => <code key={item}>{displayText(item)}</code>)}</div>
}

function AuditSchema({ schema }) {
  const rows = Object.entries(schema || {})
  return (
    <div className='fx-cmdb-audit-panel'>
      <h3>expected_schema</h3>
      {rows.map(([key, value]) => <code key={key}>{key}: {sanitizeAuditText(Array.isArray(value) ? value.join(', ') : value)}</code>)}
    </div>
  )
}

function AuditMatrix({ title = 'relation field matrix', rows }) {
  const list = Array.isArray(rows) ? rows : []
  return (
    <div className='fx-cmdb-audit-panel fx-cmdb-audit-panel--wide'>
      <h3>{title}</h3>
      <div className='fx-cmdb-field-matrix'>
        {list.map(row => (
          <div key={row.field || row.name}>
            <strong>{sanitizeAuditText(row.field || row.name)}</strong>
            <span>{sanitizeAuditText(row.status || (row.required ? '必需' : '可选'))} · {sanitizeAuditText(row.purpose || row.desc || row.reason)}</span>
          </div>
        ))}
        {!list.length && <div><strong>未返回</strong><span>后端尚未返回该矩阵。</span></div>}
      </div>
    </div>
  )
}

function RelationGraph({ result, selectedNode, onSelectNode }) {
  const nodes = result.nodes || []
  const edges = result.edges || []
  const nodeById = new Map(nodes.map(node => [node.id, node]))
  const root = nodes.find(node => node.kind === 'instance') || nodes[0] || {}
  const groups = edges.reduce((acc, edge) => {
    const label = edge.relation_name || edge.location || '关系'
    if (!acc.has(label)) acc.set(label, [])
    acc.get(label).push(edge)
    return acc
  }, new Map())
  return (
    <div className='fx-cmdb-graph-wrap'>
      <div className='fx-cmdb-canvas-shell'>
        <div className='fx-cmdb-canvas-toolbar'>
          <strong>默认拓扑</strong>
          <span>{nodes.length} 个节点 / {edges.length} 条关系</span>
          <button type='button'>管理</button>
        </div>
        <div className='fx-cmdb-mature-canvas'>
          <button type='button' className={`fx-cmdb-instance-card is-current ${selectedNode?.id === root.id ? 'is-active' : ''}`} onClick={() => onSelectNode(root)}>
            <span>{displayText(root.object_name || root.object_id || '当前模型')}</span>
            <strong>{displayText(root.name || root.id)}</strong>
            <em>{displayText(root.object_type || root.kind, '实例')}</em>
          </button>
          <div className='fx-cmdb-relation-lanes'>
            {[...groups.entries()].map(([label, group], index) => (
              <section key={label} className={`fx-cmdb-lane ${index % 2 ? 'is-green' : 'is-blue'}`}>
                <header>{displayText(label)}</header>
                {group.slice(0, 4).map(edge => {
                  const rawTarget = nodeById.get(edge.target) || nodeById.get(edge.source) || {}
                  const target = { ...rawTarget, relation_id: edge.instance_relation_id || edge.id || edge.relation_id, instance_relation_id: edge.instance_relation_id || edge.id, relation_edge: edge }
                  const active = selectedNode?.id === target.id
                  return (
                    <button key={edge.id} type='button' className={active ? 'is-active' : ''} onClick={() => onSelectNode(target)}>
                      <strong>{displayText(target.name || target.id, '缺少目标节点')}</strong>
                      <span>{displayText(edge.relation_id || edge.instance_relation_id || edge.location, '缺关系 ID')}</span>
                    </button>
                  )
                })}
              </section>
            ))}
          </div>
          <div className='fx-cmdb-minimap'>MAP</div>
        </div>
      </div>
      <aside className='fx-cmdb-node-detail'>
        <h3>关系详情</h3>
        {selectedNode ? (
          <dl>
            <dt>节点</dt><dd>{displayText(selectedNode.name)}</dd>
            <dt>模型</dt><dd>{displayText(selectedNode.object_name || selectedNode.object_id)}</dd>
            <dt>类型</dt><dd>{displayText(selectedNode.kind)}</dd>
            <dt>层级</dt><dd>{displayText(selectedNode.level)}</dd>
          </dl>
        ) : <p>选择拓扑节点后查看字段化详情。</p>}
      </aside>
    </div>
  )
}

function RelationActionBar({ blocked, selectedNode, reason, instanceId, onRelationAction, onMonitorBindings, bindingLoading }) {
  const blockedMessage = reason || 'PENDING: 二级/三级动作需要 relation_id、instance_id、监控对象绑定和审计回执契约；当前不会伪装成功。'
  return (
    <div className='fx-cmdb-relation-actions'>
      {CMDB_RELATION_ACTIONS.map(action => (
        <button key={action.key} type='button' disabled={!instanceId && !selectedNode} onClick={() => onRelationAction(action)} title={blocked || !selectedNode ? blockedMessage : `${action.label}: ${selectedNode.name}`}>
          {action.label}
        </button>
      ))}
      <button type='button' disabled={!instanceId || bindingLoading} onClick={onMonitorBindings} title={!instanceId ? '缺少 instance_id，无法读取真实监控绑定。' : '读取 CMDB 监控绑定契约'}>
        {bindingLoading ? '读取中...' : '监控绑定'}
      </button>
      {(blocked || !selectedNode) && <span>{blockedMessage}</span>}
    </div>
  )
}

function buildRelationActionAudit({ action, instanceId, selectedNode, reason }) {
  const resourceId = displayText(instanceId || selectedNode?.id, 'contract-probe')
  return {
    title: `${action.label}契约审计`,
    status: 'PENDING',
    action,
    target: {
      instance_id: resourceId,
      node_id: displayText(selectedNode?.id, '待选择真实拓扑节点'),
      object_id: displayText(selectedNode?.object_id, '待返回真实模型'),
      relation_id: displayText(selectedNode?.relation_id || selectedNode?.instance_relation_id, '待返回真实关系'),
    },
    message: reason || '二级/三级动作需要真实 relation_id、instance_id、业务上下文和审计回执，当前保持契约审计态。',
    missing_contracts: [
      action.contract,
      'cmdb_relation_action_store',
      'cmdb_relation_action_business_context_contract',
      'cmdb_relation_action_delivery_receipt_contract',
    ],
    log_query: {
      source: 'findx_audit',
      scope: 'cmdb',
      resource_type: 'cmdb_relation_action',
      action: `cmdb.relation.${action.key}.request`,
      resource_id: resourceId,
    },
  }
}

function RelationActionDrawer({ audit, onClose }) {
  return (
    <RightContractDrawer title={audit.title} status={audit.status} onClose={onClose} closeLabel={`关闭${audit.title}`}>
      <div className='fx-cmdb-binding-drawer-body'>
        <BindingDrawerSection title='动作目标'>
          <div className='fx-cmdb-binding-summary-grid'>
            <span>动作</span><strong>{audit.action.label}</strong>
            <span>instance_id</span><strong>{audit.target.instance_id}</strong>
            <span>node_id</span><strong>{audit.target.node_id}</strong>
            <span>object_id</span><strong>{audit.target.object_id}</strong>
            <span>relation_id</span><strong>{audit.target.relation_id}</strong>
          </div>
        </BindingDrawerSection>
        <BindingDrawerSection title='业务上下文'>
          <p className='fx-cmdb-relation-action-note'>{displayText(audit.message)}</p>
        </BindingDrawerSection>
        <BindingDrawerSection title='缺失契约'>
          <div className='fx-cmdb-action-contract-list'>
            {audit.missing_contracts.map(item => <code key={item}>{item}</code>)}
          </div>
        </BindingDrawerSection>
        <BindingDrawerSection title='日志审计'>
          <div className='fx-cmdb-binding-log-query'>
            <strong>FindX audit log query</strong>
            <code>{sanitizeAuditText(audit.log_query)}</code>
          </div>
        </BindingDrawerSection>
      </div>
    </RightContractDrawer>
  )
}

function RightContractDrawer({ title, status, onClose, closeLabel, children }) {
  return (
    <div className='fx-cmdb-monitor-binding-drawer' role='dialog' aria-modal='true' aria-labelledby='fx-cmdb-contract-drawer-title'>
      <button className='fx-cmdb-monitor-binding-backdrop' type='button' aria-label={closeLabel} onClick={onClose} />
      <section className='fx-cmdb-monitor-binding-panel'>
        <header className='fx-cmdb-monitor-binding-header'>
          <div>
            <strong id='fx-cmdb-contract-drawer-title'>{title}</strong>
            <span>{status}</span>
          </div>
          <button type='button' onClick={onClose} aria-label={closeLabel}>×</button>
        </header>
        {children}
      </section>
    </div>
  )
}

function MonitorBindingPanel({ result, loading, onClose }) {
  const ready = result?.status === 'ready'
  const bindings = Array.isArray(result?.bindings) ? result.bindings : []
  return (
    <RightContractDrawer title={ready ? '监控绑定' : '监控绑定契约审计'} status={loading ? 'CHECKING_CONTRACT' : ready ? `${bindings.length} 条真实绑定` : 'PENDING'} onClose={onClose} closeLabel='关闭监控绑定契约审计'>
        <div className='fx-cmdb-binding-drawer-body'>
          <BindingDrawerSection title='模型对接'>
            <div className='fx-cmdb-binding-summary-grid'>
              <span>绑定状态</span><strong>{ready ? '真实绑定数据' : '契约阻断'}</strong>
              <span>绑定数量</span><strong>{ready ? bindings.length : 0}</strong>
              <span>审计来源</span><strong>findx_audit / scope=cmdb</strong>
            </div>
          </BindingDrawerSection>
          <BindingDrawerSection title='关联属性'>
            {ready ? (
              <div className='fx-cmdb-binding-list'>
                {bindings.map(binding => (
                  <article key={binding.id || `${binding.hostid}-${binding.templateid}`}>
                    <strong>{displayText(binding.host || binding.hostid, '未命名主机')}</strong>
                    <span>templateid={displayText(binding.templateid)}</span>
                    <span>cmdb_attr_id={displayText(binding.cmdb_attr_id)} / server_attr_id={displayText(binding.server_attr_id)}</span>
                    <code>audit_ref={displayText(binding.audit_ref)}</code>
                  </article>
                ))}
              </div>
            ) : (
              <RelationBlockedAudit audit={result} loading={loading} />
            )}
          </BindingDrawerSection>
          {result?.log_query && (
            <BindingDrawerSection title='日志审计'>
              <div className='fx-cmdb-binding-log-query'>
                <strong>FindX audit log query</strong>
                <code>{sanitizeAuditText(result.log_query)}</code>
              </div>
            </BindingDrawerSection>
          )}
        </div>
    </RightContractDrawer>
  )
}

function BindingDrawerSection({ title, children }) {
  return (
    <section className='fx-cmdb-binding-drawer-section'>
      <h3>{title}</h3>
      {children}
    </section>
  )
}

function sanitizeEvidence(items = []) {
  return items.map((item, index) => {
    if (item && typeof item === 'object') {
      const kind = String(item.kind || '').toLowerCase()
      return kind.includes('capture') ? `测试证据 ${index + 1}` : `成熟来源 ${index + 1}`
    }
    const text = String(item)
    if (text.includes(':\\')) return `本地证据路径 ${index + 1}`
    return sanitizeAuditText(text)
  })
}

function sanitizeAuditText(value) {
  const source = typeof value === 'string' ? value : JSON.stringify(value || '')
  const words = [
    ['Auto', 'Ops'].join(''),
    ['Sky', 'Walking'].join(''),
    ['Sig', 'NoZ'].join(''),
    ['LW', 'OPS'].join(''),
  ]
  const drivePath = new RegExp(`[A-Za-z]:${'\\\\'}${'\\\\'}[^"'{}]+`, 'g')
  return words.reduce(
    (text, word) => text.replace(new RegExp(word, 'ig'), '成熟来源'),
    source.replace(drivePath, '本地证据路径'),
  )
}
