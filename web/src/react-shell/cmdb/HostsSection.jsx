import React, { useEffect, useState } from 'react'
import { ASSET_BLOCKERS, assetsApi, formatAssetError, splitText } from '../api/assets.js'
import { displayText, fmtTime, hostIp, hostKey, hostName, isHostOnline, normalizeTags } from './assetsModel.js'
import { Blocked, ErrorBox, Feedback, Field, Modal, Status, Tags } from './Shared.jsx'
import { Pagination } from '../shared/ConfirmModal.jsx'

const PAGE_SIZE = 20

export function HostsSection({ groups, workspaces, initialQuery, onRefreshAll, embedded = false }) {
  const [filters, setFilters] = useState({ keyword: initialQuery.q || '', resource_group_id: initialQuery.group || '', workspace_id: initialQuery.workspace || '', online: initialQuery.online || '', tag: initialQuery.tag || '' })
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [feedback, setFeedback] = useState('')
  const [modal, setModal] = useState(null)
  const [blocked, setBlocked] = useState('')
  const [page, setPage] = useState(1)

  const load = async () => {
    setLoading(true); setError(''); setPage(1)
    try { setRows(await assetsApi.hosts.list(filters)) } catch (err) { setRows([]); setError(formatAssetError(err)) } finally { setLoading(false) }
  }

  useEffect(() => { load() }, [])
  useEffect(() => {
    if (initialQuery.group !== undefined && initialQuery.group !== filters.resource_group_id) setFilters(prev => ({ ...prev, resource_group_id: initialQuery.group || '' }))
  }, [initialQuery.group])

  const patch = (key, value) => setFilters(prev => ({ ...prev, [key]: value }))
  const realAction = async (kind, row, value) => {
    setFeedback(''); setBlocked('')
    try {
      if (kind === 'tags') await assetsApi.hosts.updateTags(hostKey(row), splitText(value))
      if (kind === 'group') await assetsApi.hosts.bindResourceGroup(hostKey(row), value.trim())
      if (kind === 'workspace') await assetsApi.hosts.bindWorkspace(hostKey(row), value.trim())
      setFeedback('主机绑定信息已保存。'); setModal(null); await load(); onRefreshAll?.()
    } catch (err) { setError(formatAssetError(err)); throw err }
  }

  return (
    <section className={embedded ? '' : 'fx-assets-work'}>
      <HostFilter filters={filters} groups={groups} workspaces={workspaces} onPatch={patch} onSearch={load} loading={loading} />
      <div className='fx-assets-toolbar'>
        <button type='button' onClick={() => setBlocked(ASSET_BLOCKERS.hostCreate)}>新建主机</button>
        <button type='button' onClick={() => setBlocked(ASSET_BLOCKERS.hostCreate)}>Excel 导入</button>
        <button type='button' onClick={() => setBlocked(ASSET_BLOCKERS.terminal)}>终端</button>
        <button type='button' onClick={load}>刷新</button>
      </div>
      <ErrorBox>{error}</ErrorBox><Feedback>{feedback}</Feedback>{blocked && <Blocked>{blocked}</Blocked>}
      <HostTable rows={rows.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)} loading={loading} onDetail={row => setModal({ type: 'detail', row })} onTags={row => setModal({ type: 'tags', row, value: normalizeTags(row.tags).join(', ') })} onGroup={row => setModal({ type: 'group', row, value: row.resource_group_id || '' })} onWorkspace={row => setModal({ type: 'workspace', row, value: row.workspace_id || '' })} onBlocked={setBlocked} />
      <Pagination total={rows.length} page={page} pageSize={PAGE_SIZE} onPageChange={setPage} />
      {modal && <HostModal modal={modal} groups={groups} workspaces={workspaces} onClose={() => setModal(null)} onSave={realAction} />}
    </section>
  )
}

function HostFilter({ filters, groups, workspaces, onPatch, onSearch, loading }) {
  return (
    <div className='fx-assets-filter'>
      <input value={filters.keyword} onChange={e => onPatch('keyword', e.target.value)} placeholder='主机名 / IP / Agent' />
      <select value={filters.workspace_id} onChange={e => onPatch('workspace_id', e.target.value)}><option value=''>全部业务组</option>{workspaces.map(row => <option key={row.id} value={row.id}>{displayText(row.name)}</option>)}</select>
      <select value={filters.resource_group_id} onChange={e => onPatch('resource_group_id', e.target.value)}><option value=''>全部资源组</option>{groups.map(row => <option key={row.id} value={row.id}>{displayText(row.name)}</option>)}</select>
      <select value={filters.online} onChange={e => onPatch('online', e.target.value)}><option value=''>全部状态</option><option value='true'>在线</option><option value='false'>离线</option></select>
      <input value={filters.tag} onChange={e => onPatch('tag', e.target.value)} placeholder='标签' />
      <button className='fx-assets-button is-primary' type='button' onClick={onSearch}>{loading ? '查询中...' : '查询'}</button>
    </div>
  )
}

function HostTable({ rows, loading, onDetail, onTags, onGroup, onWorkspace, onBlocked }) {
  return (
    <div className='fx-assets-table'>
      <table><thead><tr><th>主机</th><th>IP</th><th>状态</th><th>业务组</th><th>资源组</th><th>FindX Agent</th><th>标签</th><th>最近心跳</th><th>操作</th></tr></thead>
        <tbody>{rows.map(row => <tr key={hostKey(row)}><td><strong>{hostName(row)}</strong><div className='fx-assets-muted'>{displayText(row.os)} {displayText(row.arch, '')}</div></td><td>{hostIp(row)}</td><td><Status ok={isHostOnline(row)}>{isHostOnline(row) ? '在线' : '离线'}</Status></td><td>{displayText(row.workspace_name || row.workspace_id)}</td><td>{displayText(row.resource_group_name || row.resource_group_id)}</td><td>{displayText(row.agent_id || row.agent_status)}</td><td><Tags items={row.tags} /></td><td>{fmtTime(row.last_seen_at || row.last_seen || row.updated_at)}</td><td className='fx-assets-actions'><button type='button' onClick={() => onDetail(row)}>详情</button><button type='button' onClick={() => onTags(row)}>标签</button><button type='button' onClick={() => onGroup(row)}>资源组</button><button type='button' onClick={() => onWorkspace(row)}>业务组</button><button type='button' onClick={() => onBlocked(ASSET_BLOCKERS.terminal)}>终端</button><button type='button' onClick={() => onBlocked(ASSET_BLOCKERS.monitor)}>监控</button></td></tr>)}</tbody>
      </table>
      {!rows.length && <div className='fx-assets-empty'>{loading ? '正在加载主机资产...' : '暂无主机资产。'}</div>}
    </div>
  )
}

function HostModal({ modal, groups, workspaces, onClose, onSave }) {
  const [value, setValue] = useState(modal.value || '')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const submit = async () => {
    setSaving(true); setError('')
    try { await onSave(modal.type, modal.row, value) } catch (err) { setError(formatAssetError(err)) } finally { setSaving(false) }
  }
  if (modal.type === 'detail') {
    return <Modal title='主机详情' onClose={onClose}><div className='fx-assets-form'><HostDetail row={modal.row} /><Blocked>{ASSET_BLOCKERS.hostCreate}</Blocked><Blocked>{ASSET_BLOCKERS.terminal}</Blocked></div></Modal>
  }
  const title = modal.type === 'tags' ? '编辑标签' : modal.type === 'group' ? '绑定资源组' : '绑定业务组'
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

function HostDetail({ row }) {
  const entries = { 主机: hostName(row), IP: hostIp(row), 状态: displayText(row.status), 系统: `${displayText(row.os)} ${displayText(row.arch, '')}`, 来源: displayText(row.source), Agent: displayText(row.agent_id || row.agent_status), 最近心跳: fmtTime(row.last_seen_at || row.last_seen), 更新时间: fmtTime(row.updated_at) }
  return <div className='fx-assets-table'><table><tbody>{Object.entries(entries).map(([key, value]) => <tr key={key}><th>{key}</th><td>{value}</td></tr>)}</tbody></table></div>
}
