import React, { useEffect, useMemo, useState, useCallback } from 'react'
import { assetsApi, formatAssetError, splitText } from '../api/assets.js'
import { cmdbApi } from '../api/cmdb.js'
import { post } from '../api/http.js'
import { displayText, fmtTime, hostIp, hostKey, hostName, isHostOnline, normalizeTags } from './assetsModel.js'
import { Blocked, ErrorBox, Feedback, Field, Modal, Status, Tags } from './Shared.jsx'
import { Pagination, useConfirm } from '../shared/ConfirmModal.jsx'
import { ExecSection } from './ExecSection.jsx'
import { HostProbePluginDrawer } from './HostProbePluginDrawer.jsx'
import { TerminalSection } from './TerminalSection.jsx'
import { UploadSection } from './UploadSection.jsx'
import { InstanceDetailDrawer } from './InstanceDetailDrawer.jsx'

const CLASSIFICATION_TYPES = [
  '物理机', '虚拟机', '容器', '云主机', '网络设备', '存储设备',
  '安全设备', '数据库', '中间件', '应用服务', '负载均衡', 'GPU服务器',
]

const ALERT_LEVELS = [
  { value: 'critical', label: '紧急', color: '#dc2626' },
  { value: 'warning', label: '警告', color: '#f59e0b' },
  { value: 'info', label: '信息', color: '#3b82f6' },
  { value: 'none', label: '正常', color: '#10b981' },
]

const COLLECTION_STATUS_MAP = {
  '正常': { color: '#10b981', bg: '#ecfdf5' },
  '异常': { color: '#dc2626', bg: '#fef2f2' },
  '未监控': { color: '#6b7280', bg: '#f3f4f6' },
}

const MAINTENANCE_STATUS_MAP = {
  '维护中': { color: '#f59e0b', bg: '#fffbeb' },
  '正常': { color: '#10b981', bg: '#ecfdf5' },
}

const PAGE_SIZE = 20

export function HostsSection({ groups, workspaces, initialQuery, onRefreshAll, embedded = false }) {
  const [filters, setFilters] = useState({ keyword: initialQuery.q || '', resource_group_id: initialQuery.group || '', workspace_id: initialQuery.workspace || '', online: initialQuery.online || '', tag: initialQuery.tag || '', alert_level: '', ip_exact: '', classification: '', maintenance_status: '', collection_status: '' })
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [feedback, setFeedback] = useState('')
  const [modal, setModal] = useState(null)
  const [page, setPage] = useState(1)
  const [selectedGroup, setSelectedGroup] = useState(initialQuery.group || '')
  const [treeSearch, setTreeSearch] = useState('')
  const [treeExpanded, setTreeExpanded] = useState(true)
  const [selectedIds, setSelectedIds] = useState([])
  const [batchModal, setBatchModal] = useState(null)
  const [drawerRow, setDrawerRow] = useState(null)
  const { confirm, modal: confirmModal } = useConfirm()

  const load = async (nextFilters = filters) => {
    setLoading(true); setError(''); setPage(1)
    try { setRows(await assetsApi.hosts.list(nextFilters)) } catch (err) { setRows([]); setError(formatAssetError(err)) } finally { setLoading(false) }
  }

  useEffect(() => { load() }, [])
  useEffect(() => {
    if (initialQuery.group !== undefined && initialQuery.group !== filters.resource_group_id) {
      const nextFilters = { ...filters, resource_group_id: initialQuery.group || '' }
      setFilters(nextFilters)
      load(nextFilters)
    }
  }, [initialQuery.group])

  const patch = (key, value) => setFilters(prev => ({ ...prev, [key]: value }))

  const handleGroupSelect = (groupId) => {
    setSelectedGroup(groupId)
    const nextFilters = { ...filters, resource_group_id: groupId }
    setFilters(nextFilters)
    load(nextFilters)
  }

  const realAction = async (kind, row, value) => {
    setFeedback('')
    try {
      if (kind === 'tags') await assetsApi.hosts.updateTags(hostKey(row), splitText(value))
      if (kind === 'group') await assetsApi.hosts.bindResourceGroup(hostKey(row), value.trim())
      if (kind === 'workspace') await assetsApi.hosts.bindWorkspace(hostKey(row), value.trim())
      if (kind === 'assign') {
        const groupId = (value.resource_group_id || '').trim()
        const workspaceId = (value.workspace_id || '').trim()
        if (groupId !== (row.resource_group_id || '')) await assetsApi.hosts.bindResourceGroup(hostKey(row), groupId)
        if (workspaceId !== (row.workspace_id || '')) await assetsApi.hosts.bindWorkspace(hostKey(row), workspaceId)
      }
      setFeedback('主机绑定信息已保存。'); setModal(null); await load(); onRefreshAll?.()
    } catch (err) { setError(formatAssetError(err)); throw err }
  }

  const handleDelete = async (row) => {
    const ok = await confirm({ title: '删除主机', message: `确认删除主机「${hostName(row)}」？此操作不可恢复。`, confirmText: '删除', danger: true })
    if (!ok) return
    setFeedback('')
  }

  // C01: 左侧分组树过滤
  const filteredGroups = useMemo(() => {
    if (!treeSearch) return groups
    const kw = treeSearch.toLowerCase()
    return groups.filter(g => displayText(g.name, '').toLowerCase().includes(kw))
  }, [groups, treeSearch])

  // 前端过滤增强
  const filteredRows = useMemo(() => {
    let result = rows
    if (filters.alert_level) {
      result = result.filter(r => (r.alert_level || 'none') === filters.alert_level)
    }
    if (filters.ip_exact) {
      result = result.filter(r => hostIp(r) === filters.ip_exact)
    }
    if (filters.classification) {
      result = result.filter(r => (r.classification || '') === filters.classification)
    }
    if (filters.maintenance_status) {
      result = result.filter(r => (r.maintenance_status || '正常') === filters.maintenance_status)
    }
    if (filters.collection_status) {
      result = result.filter(r => (r.collection_status || '未监控') === filters.collection_status)
    }
    return result
  }, [rows, filters.alert_level, filters.ip_exact, filters.classification, filters.maintenance_status, filters.collection_status])

  // 批量操作
  const handleBatchAction = useCallback(async (action, payload) => {
    if (!selectedIds.length) return
    setFeedback('')
    try {
      if (action === 'update-group') {
        await post('/cmdb/batch/update-group', { resource_ids: selectedIds, group: payload })
      } else if (action === 'update-owner') {
        await post('/cmdb/batch/update-owner', { resource_ids: selectedIds, owner: payload })
      } else if (action === 'set-maintenance') {
        await post('/cmdb/batch/update-maintenance', { resource_ids: selectedIds, maintenance: true })
      } else if (action === 'cancel-maintenance') {
        await post('/cmdb/batch/update-maintenance', { resource_ids: selectedIds, maintenance: false })
      } else if (action === 'delete') {
        await post('/cmdb/batch/delete', { resource_ids: selectedIds })
      } else if (action === 'deploy-agent') {
        await post('/cmdb/batch/deploy-agent', { resource_ids: selectedIds })
      }
      setFeedback(`批量操作成功，影响 ${selectedIds.length} 条资源。`)
      setSelectedIds([])
      setBatchModal(null)
      await load()
      onRefreshAll?.()
    } catch (err) {
      setError(formatAssetError(err))
    }
  }, [selectedIds])

  const handleRowClick = useCallback((row) => {
    setDrawerRow(row)
  }, [])

  return (
    <section className={embedded ? '' : 'fx-assets-split'}>
      {/* C01: 左侧分组树 */}
      {!embedded && (
        <aside className='fx-assets-tree'>
          <div className='fx-assets-toolbar' style={{ marginBottom: 8 }}>
            <input value={treeSearch} onChange={e => setTreeSearch(e.target.value)} placeholder='搜索分组...' style={{ flex: 1 }} />
            <button type='button' onClick={() => setTreeExpanded(!treeExpanded)}>{treeExpanded ? '折叠' : '展开'}</button>
          </div>
          <button type='button' className={`fx-assets-tree-row ${!selectedGroup ? 'is-active' : ''}`} onClick={() => handleGroupSelect('')}>
            <strong>全部主机</strong><span>所有分组</span>
          </button>
          {treeExpanded && filteredGroups.map(g => (
            <button type='button' key={g.id} className={`fx-assets-tree-row ${selectedGroup === g.id ? 'is-active' : ''}`} onClick={() => handleGroupSelect(g.id)}>
              <strong>{displayText(g.name)}</strong><span>{g.parent_id ? '子分组' : '根分组'}</span>
            </button>
          ))}
          {treeExpanded && !filteredGroups.length && <div className='fx-assets-empty'>暂无分组</div>}
        </aside>
      )}
      {/* 右侧主内容 */}
      <div className={embedded ? '' : 'fx-assets-detail'}>
        <HostFilter filters={filters} groups={groups} workspaces={workspaces} onPatch={patch} onSearch={load} loading={loading} />
        <div className='fx-assets-toolbar'>
          <button type='button' onClick={() => setModal({ type: 'host-create-contract' })}>新增主机</button>
          <button type='button' onClick={() => setModal({ type: 'terminal-select' })}>终端</button>
          <button type='button' onClick={load}>刷新</button>
        </div>
        {selectedIds.length > 0 && (
          <BatchToolbar
            count={selectedIds.length}
            onAction={(action) => {
              if (action === 'update-group') setBatchModal({ type: 'group' })
              else if (action === 'update-owner') setBatchModal({ type: 'owner' })
              else if (action === 'set-maintenance') handleBatchAction('set-maintenance')
              else if (action === 'cancel-maintenance') handleBatchAction('cancel-maintenance')
              else if (action === 'delete') {
                confirm({ title: '批量删除', message: `确认删除选中的 ${selectedIds.length} 条资源？此操作不可恢复。`, confirmText: '删除', danger: true })
                  .then(ok => { if (ok) handleBatchAction('delete') })
              }
              else if (action === 'deploy-agent') handleBatchAction('deploy-agent')
            }}
            onClear={() => setSelectedIds([])}
          />
        )}
        <ErrorBox>{error}</ErrorBox><Feedback>{feedback}</Feedback>
        <HostTable rows={filteredRows.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)} allRows={filteredRows} loading={loading} selectedIds={selectedIds} onSelectChange={setSelectedIds} onRowClick={handleRowClick} onDetail={row => setModal({ type: 'detail', row })} onChat={row => setModal({ type: 'host-chat-blocked', row })} onAssign={row => setModal({ type: 'assign', row, resource_group_id: row.resource_group_id || '', workspace_id: row.workspace_id || '' })} onProbe={row => setModal({ type: 'probe-plugin', row })} onTags={row => setModal({ type: 'tags', row, value: normalizeTags(row.tags).join(', ') })} onTerminal={row => setModal({ type: 'terminal', row })} onUpload={row => setModal({ type: 'upload', row })} onExec={row => setModal({ type: 'exec', row })} onDelete={handleDelete} />
        <Pagination total={filteredRows.length} page={page} pageSize={PAGE_SIZE} onPageChange={setPage} />
        {modal && <HostModal modal={modal} groups={groups} workspaces={workspaces} onClose={() => setModal(null)} onSave={realAction} confirm={confirm} />}
        {batchModal && <BatchModal type={batchModal.type} groups={groups} onClose={() => setBatchModal(null)} onConfirm={(value) => { handleBatchAction(batchModal.type === 'group' ? 'update-group' : 'update-owner', value); setBatchModal(null) }} />}
        {drawerRow && <InstanceDetailDrawer row={drawerRow} open={!!drawerRow} onClose={() => setDrawerRow(null)} />}
        {confirmModal}
      </div>
    </section>
  )
}

function HostFilter({ filters, groups, workspaces, onPatch, onSearch, loading }) {
  return (
    <div className='fx-assets-filter' style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))' }}>
      <input value={filters.keyword} onChange={e => onPatch('keyword', e.target.value)} placeholder='主机名 / IP / Agent' />
      <select value={filters.workspace_id} onChange={e => onPatch('workspace_id', e.target.value)}><option value=''>全部业务组</option>{workspaces.map(row => <option key={row.id} value={row.id}>{displayText(row.name)}</option>)}</select>
      <select value={filters.resource_group_id} onChange={e => onPatch('resource_group_id', e.target.value)}><option value=''>全部资源组</option>{groups.map(row => <option key={row.id} value={row.id}>{displayText(row.name)}</option>)}</select>
      <select value={filters.online} onChange={e => onPatch('online', e.target.value)}><option value=''>全部状态</option><option value='true'>在线</option><option value='false'>离线</option></select>
      <input value={filters.tag} onChange={e => onPatch('tag', e.target.value)} placeholder='标签' />
      <select value={filters.alert_level} onChange={e => onPatch('alert_level', e.target.value)}><option value=''>全部告警级别</option><option value='critical'>紧急</option><option value='warning'>警告</option><option value='info'>信息</option></select>
      <input value={filters.ip_exact} onChange={e => onPatch('ip_exact', e.target.value)} placeholder='IP精准匹配' />
      <select value={filters.classification} onChange={e => onPatch('classification', e.target.value)}><option value=''>全部类型</option>{CLASSIFICATION_TYPES.map(t => <option key={t} value={t}>{t}</option>)}</select>
      <select value={filters.maintenance_status} onChange={e => onPatch('maintenance_status', e.target.value)}><option value=''>全部维护情况</option><option value='维护中'>维护中</option><option value='正常'>正常</option></select>
      <select value={filters.collection_status} onChange={e => onPatch('collection_status', e.target.value)}><option value=''>全部采集情况</option><option value='正常'>正常</option><option value='异常'>异常</option><option value='未监控'>未监控</option></select>
      <button className='fx-assets-button is-primary' type='button' onClick={onSearch}>{loading ? '查询中...' : '查询'}</button>
    </div>
  )
}

function BatchToolbar({ count, onAction, onClear }) {
  return (
    <div className='fx-assets-toolbar fx-batch-toolbar'>
      <span style={{ fontSize: 13, color: '#193a63' }}>已选 <strong>{count}</strong> 项</span>
      <button type='button' onClick={() => onAction('update-group')}>修改分组</button>
      <button type='button' onClick={() => onAction('update-owner')}>修改负责人</button>
      <button type='button' onClick={() => onAction('set-maintenance')}>设为维护</button>
      <button type='button' onClick={() => onAction('cancel-maintenance')}>取消维护</button>
      <button type='button' onClick={() => onAction('deploy-agent')}>下发Agent</button>
      <button type='button' className='is-danger' onClick={() => onAction('delete')}>批量删除</button>
      <button type='button' onClick={onClear}>取消选择</button>
    </div>
  )
}

function BatchModal({ type, groups, onClose, onConfirm }) {
  const [value, setValue] = useState('')
  const title = type === 'group' ? '修改分组' : '修改负责人'
  return (
    <Modal title={title} onClose={onClose}>
      <div className='fx-assets-form'>
        {type === 'group' && (
          <Field label='目标分组'>
            <select value={value} onChange={e => setValue(e.target.value)}>
              <option value=''>请选择分组</option>
              {groups.map(g => <option key={g.id} value={g.id}>{displayText(g.name)}</option>)}
            </select>
          </Field>
        )}
        {type === 'owner' && (
          <Field label='负责人'>
            <input value={value} onChange={e => setValue(e.target.value)} placeholder='输入负责人名称' />
          </Field>
        )}
        <footer>
          <button type='button' disabled={!value} onClick={() => onConfirm(value)}>确认</button>
          <button type='button' onClick={onClose} style={{ marginLeft: 8, background: '#fff', color: '#21344d', borderColor: '#cdd8e8' }}>取消</button>
        </footer>
      </div>
    </Modal>
  )
}

function CollectionBadge({ status }) {
  const label = status || '未监控'
  const style = COLLECTION_STATUS_MAP[label] || COLLECTION_STATUS_MAP['未监控']
  return <span style={{ display: 'inline-block', padding: '2px 8px', borderRadius: 999, fontSize: 12, color: style.color, background: style.bg }}>{label}</span>
}

function MaintenanceBadge({ status }) {
  const label = status || '正常'
  const style = MAINTENANCE_STATUS_MAP[label] || MAINTENANCE_STATUS_MAP['正常']
  return <span style={{ display: 'inline-block', padding: '2px 8px', borderRadius: 999, fontSize: 12, color: style.color, background: style.bg }}>{label}</span>
}

function AlertDot({ level }) {
  const info = ALERT_LEVELS.find(a => a.value === level) || ALERT_LEVELS[3]
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 12 }}>
      <i style={{ width: 8, height: 8, borderRadius: '50%', background: info.color, display: 'inline-block' }} />
      {info.label}
    </span>
  )
}

function HostTable({ rows, allRows = rows, loading, selectedIds, onSelectChange, onRowClick, onDetail, onChat, onAssign, onProbe, onTags, onTerminal, onUpload, onExec, onDelete }) {
  const cmdbColumns = useMemo(() => buildCmdbHostColumns(allRows), [allRows])
  const useCmdbProjection = cmdbColumns.length > 0
  const tableMinWidth = useCmdbProjection ? Math.max(880, 520 + cmdbColumns.length * 148) : 1280

  const allSelected = rows.length > 0 && rows.every(r => selectedIds.includes(hostKey(r)))
  const toggleAll = () => {
    if (allSelected) {
      onSelectChange(selectedIds.filter(id => !rows.some(r => hostKey(r) === id)))
    } else {
      const newIds = [...new Set([...selectedIds, ...rows.map(r => hostKey(r))])]
      onSelectChange(newIds)
    }
  }
  const toggleRow = (row) => {
    const id = hostKey(row)
    onSelectChange(selectedIds.includes(id) ? selectedIds.filter(i => i !== id) : [...selectedIds, id])
  }

  return (
    <div className={`fx-assets-table fx-assets-host-table ${useCmdbProjection ? 'is-cmdb-projection' : ''}`}>
      <table style={{ minWidth: tableMinWidth }}>
        <thead>
          <tr>
            <th style={{ width: 36 }}><input type='checkbox' checked={allSelected} onChange={toggleAll} aria-label='全选' /></th>
            {useCmdbProjection ? (
              <>
                <th>资源</th>
                {cmdbColumns.map(column => (
                  <th key={column.attr} className='fx-assets-cmdb-column-head'>
                    <strong>{column.label}</strong>
                    <span>{[column.tag, column.valueType, column.unit].filter(Boolean).join(' · ') || 'CMDB'}</span>
                  </th>
                ))}
                <th>所属业务</th>
                <th>采集情况</th>
                <th>类型</th>
                <th>子类型</th>
                <th>维护情况</th>
                <th>负责人</th>
                <th>告警级别</th>
                <th>状态</th>
                <th>操作</th>
              </>
            ) : (
              <><th>主机名</th><th>IP</th><th>所属业务</th><th>采集情况</th><th>类型</th><th>子类型</th><th>维护情况</th><th>负责人</th><th>告警级别</th><th>CPU</th><th>内存</th><th>磁盘</th><th>状态</th><th>操作</th></>
            )}
          </tr>
        </thead>
        <tbody>{rows.map(row => (
          <tr key={hostKey(row)} onClick={() => onRowClick(row)} style={{ cursor: 'pointer' }}>
            <td onClick={e => e.stopPropagation()}><input type='checkbox' checked={selectedIds.includes(hostKey(row))} onChange={() => toggleRow(row)} aria-label={`选择 ${hostName(row)}`} /></td>
            {useCmdbProjection ? (
              <>
                <td className='fx-assets-host-resource-cell'>
                  <strong>{hostName(row)}</strong>
                  <div className='fx-assets-muted'>{hostIp(row)}</div>
                  <div className='fx-assets-muted'>{displayText(row?.cmdb_instance?.object_name || row?.cmdb_instance?.object_id, '未关联 CMDB')}</div>
                </td>
                {cmdbColumns.map(column => (
                  <td key={column.attr} className={column.sensitive || column.masked ? 'fx-assets-cmdb-sensitive' : ''}>
                    {renderCmdbHostCell(row, column)}
                  </td>
                ))}
                <td>{displayText(row.business_name, '-')}</td>
                <td><CollectionBadge status={row.collection_status} /></td>
                <td>{displayText(row.classification, '-')}</td>
                <td>{displayText(row.subtype, '-')}</td>
                <td><MaintenanceBadge status={row.maintenance_status} /></td>
                <td>{displayText(row.owner, '-')}</td>
                <td><AlertDot level={row.alert_level} /></td>
                <td><Status ok={isHostOnline(row)}>{isHostOnline(row) ? '在线' : '离线'}</Status></td>
              </>
            ) : (
              <>
                <td><strong>{hostName(row)}</strong><div className='fx-assets-muted'>{displayText(row.os)} {displayText(row.arch, '')}</div></td>
                <td>{hostIp(row)}</td>
                <td>{displayText(row.business_name, '-')}</td>
                <td><CollectionBadge status={row.collection_status} /></td>
                <td>{displayText(row.classification, '-')}</td>
                <td>{displayText(row.subtype, '-')}</td>
                <td><MaintenanceBadge status={row.maintenance_status} /></td>
                <td>{displayText(row.owner, '-')}</td>
                <td><AlertDot level={row.alert_level} /></td>
                <td>{displayText(row.cpu_cores || row.cpu, '-')}</td>
                <td>{formatMemory(row.memory_total || row.memory)}</td>
                <td>{formatDisk(row.disk_total || row.disk)}</td>
                <td><Status ok={isHostOnline(row)}>{isHostOnline(row) ? '在线' : '离线'}</Status></td>
              </>
            )}
            <td className='fx-assets-actions' onClick={e => e.stopPropagation()}>
              <button type='button' onClick={() => onDetail(row)}>详情</button>
              <button type='button' onClick={() => onChat(row)}>AI 对话</button>
              <button type='button' onClick={() => onProbe(row)}>探针/插件</button>
              <details className='fx-assets-more-actions'>
                <summary>更多</summary>
                <div onClick={event => { event.currentTarget.closest('details').open = false }}>
                  <button type='button' onClick={() => onAssign(row)}>分配</button>
                  <button type='button' onClick={() => onTerminal(row)}>终端</button>
                  <button type='button' onClick={() => onUpload(row)}>上传</button>
                  <button type='button' onClick={() => onExec(row)}>执行</button>
                  <button type='button' onClick={() => onTags(row)}>标签</button>
                  <button type='button' className='is-danger' onClick={() => onDelete(row)}>删除</button>
                </div>
              </details>
            </td>
          </tr>
        ))}</tbody>
      </table>
      {!rows.length && <div className='fx-assets-empty'>{loading ? '正在加载主机资产...' : '暂无主机资产。'}</div>}
    </div>
  )
}

const isCmdbColumnVisible = column => column?.visible !== false && column?.visible !== 'false'

const cmdbSensitiveColumnPattern = /(pass|secret|token|auth|cookie|private|dsn|credential|cert|owner|phone|mobile|account|user)/i

const buildCmdbHostColumns = rows => {
  const byAttr = new Map()
  for (const row of rows) {
    for (const source of asArray(row?.cmdb_columns)) {
      const attr = cmdbText(source?.attr, '')
      if (!attr || !isCmdbColumnVisible(source)) continue
      const column = {
        attr,
        label: cmdbText(source?.label || source?.name || attr, attr),
        valueType: cmdbText(source?.value_type || source?.valueType, ''),
        tag: cmdbText(source?.tag, ''),
        unit: cmdbText(source?.unit, ''),
        sort: cmdbSortValue(source),
        sensitive: source?.sensitive === true || source?.sensitive === 'true' || cmdbSensitiveColumnPattern.test(`${source?.attr || ''} ${source?.label || ''}`),
        masked: source?.masked === true || source?.masked === 'true',
      }
      if (column.sensitive) column.masked = true
      const existing = byAttr.get(attr)
      if (!existing || column.sort < existing.sort || (existing.label === attr && column.label !== attr)) {
        byAttr.set(attr, column)
      }
    }
  }
  return Array.from(byAttr.values())
    .filter(column => rows.some(row => hasCmdbProjectionValue(row, column)))
    .sort((a, b) => a.sort - b.sort || a.label.localeCompare(b.label, 'zh-CN'))
}

const cmdbRawValue = (row, attr) => row?.cmdb_values && Object.prototype.hasOwnProperty.call(row.cmdb_values, attr) ? row.cmdb_values[attr] : undefined

const hasCmdbProjectionValue = (row, column) => {
  const raw = cmdbRawValue(row, column.attr)
  if (raw !== undefined && raw !== null && raw !== '') return true
  const fallback = cmdbAssetFallbackValue(row, column.attr)
  return fallback !== undefined && fallback !== null && fallback !== '' && fallback !== '-'
}

const cmdbAssetFallbackValue = (row, attr) => {
  const key = String(attr || '').toLowerCase()
  if (['name', 'hostname', 'host_name', 'instance_name'].includes(key)) return hostName(row)
  if (['ip_address', 'mgmt_ip', 'ip', 'host_ip', 'os001'].includes(key)) return hostIp(row)
  if (['cpu', 'cpu_cores', 'cpu_core'].includes(key)) return displayText(row?.cpu_cores || row?.cpu, '-')
  if (['memory', 'memory_total', 'mem_total'].includes(key)) return formatMemory(row?.memory_total || row?.memory)
  if (['disk', 'disk_total', 'disk_size'].includes(key)) return formatDisk(row?.disk_total || row?.disk)
  if (key === 'agent_status') return isHostOnline(row) ? '在线' : '离线'
  return undefined
}

const formatCmdbTableValue = value => {
  if (value === null || value === undefined || value === '') return '-'
  if (Array.isArray(value)) return value.map(item => displayText(item, '')).filter(Boolean).join(', ') || '-'
  if (typeof value === 'object') return '结构化字段'
  return displayText(value)
}

const renderCmdbHostCell = (row, column) => {
  const raw = cmdbRawValue(row, column.attr)
  const fallback = raw === undefined ? cmdbAssetFallbackValue(row, column.attr) : undefined
  const hasValue = raw !== undefined && raw !== null && raw !== ''
  if ((column.sensitive || column.masked) && hasValue) return '******'
  const value = formatCmdbTableValue(raw === undefined ? fallback : raw)
  if (value === '-' || !column.unit || column.sensitive || column.masked) return value
  return `${value} ${column.unit}`
}

function formatMemory(val) {
  if (!val) return '-'
  const num = Number(val)
  if (Number.isNaN(num)) return String(val)
  if (num > 1e9) return `${(num / 1073741824).toFixed(1)} GB`
  if (num > 1e6) return `${(num / 1048576).toFixed(0)} MB`
  return String(val)
}

function formatDisk(val) {
  if (!val) return '-'
  const num = Number(val)
  if (Number.isNaN(num)) return String(val)
  if (num > 1e12) return `${(num / 1099511627776).toFixed(1)} TB`
  if (num > 1e9) return `${(num / 1073741824).toFixed(0)} GB`
  return String(val)
}

// C02: 主机详情抽屉 + 编辑/终端/上传/执行弹窗
const asArray = value => Array.isArray(value) ? value : []

const pickCmdbInstanceId = row => row?.cmdb_instance?.instance_id || ''

const cmdbText = (value, fallback = '-') => value === null || value === undefined || value === '' ? fallback : String(value)

const cmdbSortValue = item => {
  const raw = item?.sort ?? item?.Sort ?? item?.order ?? item?.Order ?? 0
  const num = Number(raw)
  return Number.isFinite(num) ? num : 0
}

const pickCmdbDetailPayload = raw => raw?.data?.data || raw?.data || raw

const normalizeCmdbDetailGroups = raw => {
  const payload = pickCmdbDetailPayload(raw)
  return asArray(payload?.base)
    .map((group, groupIndex) => ({
      tag: cmdbText(group?.tag || group?.label || group?.name, `分组 ${groupIndex + 1}`),
      sort: cmdbSortValue(group),
      infos: asArray(group?.infos)
        .map((info, infoIndex) => ({
          key: cmdbText(info?.key || info?.field || info?.id || `${groupIndex}-${infoIndex}`),
          label: cmdbText(info?.label || info?.name || info?.field || info?.key, '未命名字段'),
          value: cmdbText(info?.value ?? info?.display_value ?? info?.text),
          unit: cmdbText(info?.unit, ''),
          sort: cmdbSortValue(info),
        }))
        .sort((a, b) => a.sort - b.sort),
    }))
    .filter(group => group.infos.length)
    .sort((a, b) => a.sort - b.sort)
}

function HostModal({ modal, groups, workspaces, onClose, onSave, confirm }) {
  const [value, setValue] = useState(modal.value || '')
  const [assignment, setAssignment] = useState({ resource_group_id: modal.resource_group_id || '', workspace_id: modal.workspace_id || '' })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const submit = async () => {
    setSaving(true); setError('')
    try { await onSave(modal.type, modal.row, modal.type === 'assign' ? assignment : value) } catch (err) { setError(formatAssetError(err)) } finally { setSaving(false) }
  }

  // C02: 详情抽屉
  if (modal.type === 'detail') {
    return (
      <Modal title='主机详情' onClose={onClose}>
        <div className='fx-assets-form'>
          <HostDetailDrawer row={modal.row} />
          <footer>
            <button type='button' onClick={onClose}>关闭</button>
          </footer>
        </div>
      </Modal>
    )
  }

  if (modal.type === 'probe-plugin') {
    return <HostProbePluginDrawer row={modal.row} onClose={onClose} confirm={confirm} />
  }

  if (modal.type === 'terminal') {
    return (
      <Modal title={`终端 - ${hostName(modal.row)}`} onClose={onClose}>
        <TerminalSection hostId={hostKey(modal.row)} hostName={hostName(modal.row)} hostIp={hostIp(modal.row)} onClose={onClose} />
      </Modal>
    )
  }
  if (modal.type === 'upload') {
    return (
      <Modal title={`文件上传 - ${hostName(modal.row)}`} onClose={onClose}>
        <UploadSection hostId={hostKey(modal.row)} hostName={hostName(modal.row)} hostIp={hostIp(modal.row)} onClose={onClose} />
      </Modal>
    )
  }
  if (modal.type === 'exec') {
    return (
      <Modal title={`命令执行 - ${hostName(modal.row)}`} onClose={onClose}>
        <ExecSection hostId={hostKey(modal.row)} hostName={hostName(modal.row)} hostIp={hostIp(modal.row)} onClose={onClose} />
      </Modal>
    )
  }

  if (modal.type === 'host-create-contract') {
    return (
      <HostContractPanel title='新增主机' onClose={onClose}>
        <HostCreateContract groups={groups} workspaces={workspaces} />
      </HostContractPanel>
    )
  }

  if (modal.type === 'host-chat-blocked') {
    return (
      <HostContractPanel title={`单机诊断对话 - ${hostName(modal.row)}`} onClose={onClose}>
        <HostChatBlocked row={modal.row} />
      </HostContractPanel>
    )
  }

  if (modal.type === 'terminal-select') {
    return (
      <Modal title='选择终端主机' onClose={onClose}>
        <div className='fx-assets-form'>
          <p className='fx-assets-muted'>请在主机列表中选择具体主机后查看详情。</p>
          <footer><button type='button' onClick={onClose}>关闭</button></footer>
        </div>
      </Modal>
    )
  }

  const title = modal.type === 'tags' ? '编辑标签' : modal.type === 'group' ? '绑定资源组' : '绑定业务组'
  if (modal.type === 'assign') {
    return (
      <Modal title={`分配主机 - ${hostName(modal.row)}`} onClose={onClose}>
        <div className='fx-assets-form'>
          <div className='fx-assets-host-assignment'>
            <strong>{hostName(modal.row)}</strong>
            <span>{hostIp(modal.row)} · Agent {displayText(modal.row.agent_id || modal.row.agent_status)}</span>
          </div>
          <div className='fx-assets-form-grid'>
            <Field label='业务组'><select value={assignment.workspace_id} onChange={e => setAssignment(prev => ({ ...prev, workspace_id: e.target.value }))}><option value=''>解除绑定</option>{workspaces.map(row => <option key={row.id} value={row.id}>{displayText(row.name)}</option>)}</select></Field>
            <Field label='资源组'><select value={assignment.resource_group_id} onChange={e => setAssignment(prev => ({ ...prev, resource_group_id: e.target.value }))}><option value=''>解除绑定</option>{groups.map(row => <option key={row.id} value={row.id}>{displayText(row.name)}</option>)}</select></Field>
          </div>
          <p className='fx-assets-muted'>保存后会写入 FindX 审计日志，日志中心可按 scope=cmdb、resource_type=host_asset 查询分配记录。</p>
          <ErrorBox>{error}</ErrorBox><footer><button type='button' disabled={saving} onClick={submit}>{saving ? '保存中...' : '保存分配'}</button></footer>
        </div>
      </Modal>
    )
  }
  return (
    <Modal title={title} onClose={onClose}>
      <div className='fx-assets-form'>
        {modal.type === 'tags' && <Field label='标签'><textarea rows='4' value={value} onChange={e => setValue(e.target.value)} /></Field>}
        {modal.type === 'group' && <Field label='资源组'><select value={value} onChange={e => setValue(e.target.value)}><option value=''>解除绑定</option>{groups.map(row => <option key={row.id} value={row.id}>{displayText(row.name)}</option>)}</select></Field>}
        {modal.type === 'workspace' && <Field label='业务组'><select value={value} onChange={e => setValue(e.target.value)}><option value=''>解除绑定</option>{workspaces.map(row => <option key={row.id} value={row.id}>{displayText(row.name)}</option>)}</select></Field>}
        <ErrorBox>{error}</ErrorBox><footer><button type='button' disabled={saving} onClick={submit}>{saving ? '保存中...' : '保存'}</button></footer>
      </div>
    </Modal>
  )
}

function HostContractPanel({ title, onClose, children }) {
  return (
    <div className='fx-assets-contract-drawer' role='dialog' aria-modal='true' aria-labelledby='fx-assets-contract-title'>
      <button type='button' className='fx-assets-contract-backdrop' aria-label='关闭面板' onClick={onClose} />
      <section className='fx-assets-contract-panel'>
        <header>
          <div>
            <strong id='fx-assets-contract-title'>{title}</strong>
            <span>FindX CMDB 契约审计</span>
          </div>
          <button type='button' onClick={onClose} aria-label='关闭面板'>×</button>
        </header>
        <div className='fx-assets-contract-body'>{children}</div>
      </section>
    </div>
  )
}

function HostCreateContract({ groups, workspaces }) {
  const requiredFields = [
    ['hostname', '主机名，CMDB host.name / hostname，不允许为空'],
    ['primary_ip', '主 IP，需通过 IP 格式校验并作为资产唯一性候选'],
    ['os_type', '操作系统类型，例如 linux / windows / unix'],
    ['resource_group_id', `资源组引用，当前可选 ${groups.length} 个资源组`],
    ['workspace_id', `业务组引用，当前可选 ${workspaces.length} 个业务组`],
    ['credential_ref', '连接凭据引用，仅保存引用，不接收明文密钥'],
    ['owner_ref', '资产负责人或服务归属引用'],
    ['audit_reason', '新增原因，用于 FindX 审计链路'],
  ]
  return (
    <div className='fx-assets-contract-content'>
      <Blocked>缺少后端 POST /host-assets 或 CMDB instance create 契约，当前只展示新增主机字段契约，不会创建主机。</Blocked>
      <p className='fx-assets-muted'>新增主机必须由 CMDB 模型字段驱动，并在后端提供唯一性校验、凭据引用校验、审计写入和回滚契约后才能开放提交。</p>
      <section className='fx-cmdb-blocked-grid'>
        {requiredFields.map(([field, desc]) => (
          <article key={field}>
            <strong>{field}</strong>
            <code>{desc}</code>
          </article>
        ))}
      </section>
      <div className='fx-assets-contract-audit'>
        <strong>missing contracts</strong>
        <span>POST /host-assets 或 CMDB instance create</span>
        <span>CMDB host model required fields schema</span>
        <span>credential_ref validator</span>
        <span>asset uniqueness and rollback receipt</span>
      </div>
    </div>
  )
}

function HostChatBlocked({ row }) {
  const [state, setState] = useState({ loading: true, audit: null, error: '', message: '' })
  const [draft, setDraft] = useState('请基于这台主机的 CMDB、Agent 和监控上下文做只读诊断预检。')
  const hostId = hostKey(row)

  useEffect(() => {
    let alive = true
    setState({ loading: true, audit: null, error: '', message: '' })
    cmdbApi.hostAiSession.preflight(hostId)
      .then(audit => {
        if (alive) setState({ loading: false, audit, error: '', message: '' })
      })
      .catch(err => {
        if (alive) setState({ loading: false, audit: null, error: formatAssetError(err), message: '' })
      })
    return () => { alive = false }
  }, [hostId])

  const requestPreflight = async () => {
    setState(prev => ({ ...prev, loading: true, error: '', message: '' }))
    try {
      const audit = await cmdbApi.hostAiSession.request(hostId, {
        message: draft,
        tool: 'readonly_diagnosis',
        metadata: {
          host_id: hostId,
          host_name: hostName(row),
          host_ip: hostIp(row),
          cmdb_instance_id: pickCmdbInstanceId(row),
        },
      })
      setState({ loading: false, audit, error: '', message: '请求已进入单机诊断会话预检，未建立真实会话，未调用命令或工具。' })
    } catch (err) {
      setState(prev => ({ ...prev, loading: false, error: formatAssetError(err), message: '' }))
    }
  }

  const audit = state.audit
  const hostContext = audit?.host_context || {}
  const missingContracts = normalizeContractRows(audit?.missing_contracts)
  const auditQuery = audit?.findx_audit_query || {}

  return (
    <div className='fx-assets-contract-content'>
      <div className='fx-assets-host-assignment'>
        <strong>{hostName(row)}</strong>
        <span>{hostIp(row)} · Agent {displayText(row.agent_id || row.agent_status, '未上报')}</span>
      </div>
      {state.loading && <p className='fx-assets-muted'>正在读取单机诊断会话 runtime 契约...</p>}
      <ErrorBox>{state.error}</ErrorBox>
      <Feedback>{state.message}</Feedback>
      <Blocked>{audit?.message || '单机诊断对话缺少会话传输、主机上下文、工具审计、输出回执和风险策略，当前不会建立真实连接。'}</Blocked>
      <section className='fx-host-ai-layout'>
        <div className='fx-host-ai-context'>
          <h4>主机上下文</h4>
          <dl>
            <dt>资源</dt><dd>{hostName(row)} / {hostIp(row)}</dd>
            <dt>CMDB 实例</dt><dd>{displayText(hostContext?.cmdb_instance?.instance_id || pickCmdbInstanceId(row), '未建立映射')}</dd>
            <dt>CMDB 模型</dt><dd>{displayText(hostContext?.cmdb_instance?.object_name || hostContext?.cmdb_instance?.object_id, '未建立映射')}</dd>
            <dt>监控对象</dt><dd>{displayText(hostContext?.monitor_target?.target_id || row.host_id || row.ident, '未上报')}</dd>
            <dt>Agent</dt><dd>{displayText(hostContext?.agent?.agent_id || row.agent_id || row.agent_status, '未上报')}</dd>
            <dt>执行边界</dt><dd>{displayText(audit?.preflight?.readonly_context_only ? '只读上下文' : '待确认')}</dd>
          </dl>
        </div>
        <div className='fx-host-ai-request'>
          <h4>请求预检</h4>
          <textarea rows='5' value={draft} onChange={event => setDraft(event.target.value)} />
          <button type='button' onClick={requestPreflight} disabled={state.loading}>请求诊断预检</button>
          <p>预检只返回契约审计，不创建会话、不调用模型、不执行命令。</p>
        </div>
      </section>
      <section className='fx-cmdb-blocked-grid'>
        {missingContracts.map(item => (
          <article key={item.id}>
            <strong>{item.id}</strong>
            <code>{item.status || ''}</code>
          </article>
        ))}
      </section>
      <div className='fx-assets-contract-audit'>
        <strong>findx_audit query</strong>
        <span>source={displayText(auditQuery.source, 'findx_audit')}</span>
        <span>scope={displayText(auditQuery.scope, 'cmdb')}</span>
        <span>resource_type={displayText(auditQuery.resource_type, 'cmdb_host_ai_session')}</span>
        <span>action={displayText(auditQuery.action, 'cmdb.host_ai_session.preflight')}</span>
        <span>host_id={displayText(auditQuery.host_id || hostId)}</span>
      </div>
    </div>
  )
}

const normalizeContractRows = items => {
  const rows = asArray(items)
    .map(item => typeof item === 'string'
      ? { id: item, status: '' }
      : { id: cmdbText(item?.id || item?.contract_id || item?.key || item?.name, ''), status: cmdbText(item?.status || item?.state || item?.reason, '') })
    .filter(item => item.id)
  return rows.length ? rows : [
    { id: 'cmdb_ai_host_session_transport_contract', status: '' },
    { id: 'cmdb_ai_host_context_contract', status: '' },
    { id: 'cmdb_ai_tool_audit_contract', status: '' },
    { id: 'cmdb_ai_output_receipt_contract', status: '' },
  ]
}

// C02: 主机详情内容
function HostDetailDrawer({ row }) {
  const [detailState, setDetailState] = useState({ loading: false, groups: [], blocked: '' })
  const cmdbInstanceId = pickCmdbInstanceId(row)
  const basicInfo = [
    ['主机名', hostName(row)],
    ['IP 地址', hostIp(row)],
    ['操作系统', `${displayText(row.os)} ${displayText(row.arch, '')}`],
    ['CPU', displayText(row.cpu_cores || row.cpu, '-')],
    ['内存', formatMemory(row.memory_total || row.memory)],
    ['磁盘', formatDisk(row.disk_total || row.disk)],
    ['状态', isHostOnline(row) ? '在线' : '离线'],
    ['来源', displayText(row.source)],
    ['CMDB 实例', displayText(cmdbInstanceId, '未关联')],
    ['最近心跳', fmtTime(row.last_seen_at || row.last_seen || row.updated_at)],
  ]

  useEffect(() => {
    let cancelled = false
    if (!cmdbInstanceId) {
      setDetailState({ loading: false, groups: [], blocked: '未关联 CMDB 实例，使用主机资产基础信息。' })
      return () => { cancelled = true }
    }
    const loadDetail = async () => {
      setDetailState({ loading: true, groups: [], blocked: '' })
      try {
        const raw = await cmdbApi.instances.compatibleDetail(cmdbInstanceId)
        const groups = normalizeCmdbDetailGroups(raw)
        if (cancelled) return
        if (!groups.length) {
          setDetailState({ loading: false, groups: [], blocked: 'CMDB detail-compatible 未返回 base[].infos 字段契约，已降级展示主机资产基础信息。' })
          return
        }
        setDetailState({ loading: false, groups, blocked: '' })
      } catch (err) {
        if (cancelled) return
        const status = err?.status ? `HTTP ${err.status}` : '请求失败'
        setDetailState({ loading: false, groups: [], blocked: `CMDB detail-compatible ${status}，已降级展示主机资产基础信息。` })
      }
    }
    loadDetail()
    return () => { cancelled = true }
  }, [cmdbInstanceId])

  return (
    <div>
      <h3 style={{ margin: '0 0 8px', fontSize: 14, color: '#193a63' }}>主机资产基础信息</h3>
      <div className='fx-assets-table'><table><tbody>{basicInfo.map(([k, v]) => <tr key={k}><th style={{ width: 100 }}>{k}</th><td>{v}</td></tr>)}</tbody></table></div>
      <h3 style={{ margin: '12px 0 8px', fontSize: 14, color: '#193a63' }}>CMDB 动态详情</h3>
      {detailState.loading && <p className='fx-assets-muted'>正在读取 CMDB detail-compatible 详情...</p>}
      {detailState.blocked && <Blocked>{detailState.blocked}</Blocked>}
      {detailState.groups.map(group => (
        <section key={group.tag} style={{ marginTop: 10 }}>
          <h4 style={{ margin: '0 0 6px', fontSize: 13, color: '#193a63' }}>{group.tag}</h4>
          <div className='fx-assets-table'>
            <table><tbody>{group.infos.map(info => (
              <tr key={info.key}>
                <th style={{ width: 140 }}>{info.label}</th>
                <td>{info.value}{info.unit ? ` ${info.unit}` : ''}</td>
              </tr>
            ))}</tbody></table>
          </div>
        </section>
      ))}
      <h3 style={{ margin: '12px 0 8px', fontSize: 14, color: '#193a63' }}>标签</h3>
      <Tags items={row.tags} />
    </div>
  )
}
