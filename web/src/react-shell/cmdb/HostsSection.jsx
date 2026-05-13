import React, { useEffect, useMemo, useState } from 'react'
import { ASSET_BLOCKERS, assetsApi, formatAssetError, splitText } from '../api/assets.js'
import { displayText, fmtTime, hostIp, hostKey, hostName, isHostOnline, normalizeTags } from './assetsModel.js'
import { Blocked, ErrorBox, Feedback, Field, Modal, Status, Tags } from './Shared.jsx'
import { Pagination, useConfirm } from '../shared/ConfirmModal.jsx'
import { ExecSection } from './ExecSection.jsx'
import { TerminalSection } from './TerminalSection.jsx'
import { UploadSection } from './UploadSection.jsx'

const PAGE_SIZE = 20

export function HostsSection({ groups, workspaces, initialQuery, onRefreshAll, embedded = false }) {
  const [filters, setFilters] = useState({ keyword: initialQuery.q || '', resource_group_id: initialQuery.group || '', workspace_id: initialQuery.workspace || '', online: initialQuery.online || '', tag: initialQuery.tag || '' })
  const [rows, setRows] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [feedback, setFeedback] = useState('')
  const [modal, setModal] = useState(null)
  const [page, setPage] = useState(1)
  const [selectedGroup, setSelectedGroup] = useState(initialQuery.group || '')
  const [treeSearch, setTreeSearch] = useState('')
  const [treeExpanded, setTreeExpanded] = useState(true)
  const { confirm, modal: confirmModal } = useConfirm()

  const load = async () => {
    setLoading(true); setError(''); setPage(1)
    try { setRows(await assetsApi.hosts.list(filters)) } catch (err) { setRows([]); setError(formatAssetError(err)) } finally { setLoading(false) }
  }

  useEffect(() => { load() }, [])
  useEffect(() => {
    if (initialQuery.group !== undefined && initialQuery.group !== filters.resource_group_id) {
      setFilters(prev => ({ ...prev, resource_group_id: initialQuery.group || '' }))
    }
  }, [initialQuery.group])

  const patch = (key, value) => setFilters(prev => ({ ...prev, [key]: value }))

  const handleGroupSelect = (groupId) => {
    setSelectedGroup(groupId)
    setFilters(prev => ({ ...prev, resource_group_id: groupId }))
    setTimeout(load, 0)
  }

  const realAction = async (kind, row, value) => {
    setFeedback('')
    try {
      if (kind === 'tags') await assetsApi.hosts.updateTags(hostKey(row), splitText(value))
      if (kind === 'group') await assetsApi.hosts.bindResourceGroup(hostKey(row), value.trim())
      if (kind === 'workspace') await assetsApi.hosts.bindWorkspace(hostKey(row), value.trim())
      setFeedback('主机绑定信息已保存。'); setModal(null); await load(); onRefreshAll?.()
    } catch (err) { setError(formatAssetError(err)); throw err }
  }

  const handleDelete = async (row) => {
    const ok = await confirm({ title: '删除主机', message: `确认删除主机「${hostName(row)}」？此操作不可恢复。`, confirmText: '删除', danger: true })
    if (!ok) return
    setFeedback('BLOCKED_BY_CONTRACT: 主机删除缺少审计、回滚和关联资源影响分析契约，当前不会删除主机。')
  }

  // C01: 左侧分组树过滤
  const filteredGroups = useMemo(() => {
    if (!treeSearch) return groups
    const kw = treeSearch.toLowerCase()
    return groups.filter(g => displayText(g.name, '').toLowerCase().includes(kw))
  }, [groups, treeSearch])

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
          <button type='button' onClick={() => setModal({ type: 'import' })}>导入主机</button>
          <button type='button' onClick={() => setModal({ type: 'terminal-select' })}>终端</button>
          <button type='button' onClick={load}>刷新</button>
        </div>
        <ErrorBox>{error}</ErrorBox><Feedback>{feedback}</Feedback>
        <HostTable rows={rows.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)} loading={loading} onDetail={row => setModal({ type: 'detail', row })} onTags={row => setModal({ type: 'tags', row, value: normalizeTags(row.tags).join(', ') })} onGroup={row => setModal({ type: 'group', row, value: row.resource_group_id || '' })} onWorkspace={row => setModal({ type: 'workspace', row, value: row.workspace_id || '' })} onTerminal={row => setModal({ type: 'terminal', row })} onUpload={row => setModal({ type: 'upload', row })} onExec={row => setModal({ type: 'exec', row })} onDelete={handleDelete} />
        <Pagination total={rows.length} page={page} pageSize={PAGE_SIZE} onPageChange={setPage} />
        {modal && <HostModal modal={modal} groups={groups} workspaces={workspaces} onClose={() => setModal(null)} onSave={realAction} />}
        {confirmModal}
      </div>
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

function HostTable({ rows, loading, onDetail, onTags, onGroup, onWorkspace, onTerminal, onUpload, onExec, onDelete }) {
  return (
    <div className='fx-assets-table'>
      <table><thead><tr><th>主机名</th><th>IP</th><th>CPU</th><th>内存</th><th>磁盘</th><th>状态</th><th>操作</th></tr></thead>
        <tbody>{rows.map(row => <tr key={hostKey(row)}>
          <td><strong>{hostName(row)}</strong><div className='fx-assets-muted'>{displayText(row.os)} {displayText(row.arch, '')}</div></td>
          <td>{hostIp(row)}</td>
          <td>{displayText(row.cpu_cores || row.cpu, '-')}</td>
          <td>{formatMemory(row.memory_total || row.memory)}</td>
          <td>{formatDisk(row.disk_total || row.disk)}</td>
          <td><Status ok={isHostOnline(row)}>{isHostOnline(row) ? '在线' : '离线'}</Status></td>
          <td className='fx-assets-actions'>
            <button type='button' onClick={() => onDetail(row)}>详情</button>
            <button type='button' onClick={() => onTerminal(row)}>终端</button>
            <button type='button' onClick={() => onUpload(row)}>上传</button>
            <button type='button' onClick={() => onExec(row)}>执行</button>
            <button type='button' onClick={() => onTags(row)}>标签</button>
            <button type='button' className='is-danger' onClick={() => onDelete(row)}>删除</button>
          </td>
        </tr>)}</tbody>
      </table>
      {!rows.length && <div className='fx-assets-empty'>{loading ? '正在加载主机资产...' : '暂无主机资产。'}</div>}
    </div>
  )
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
function HostModal({ modal, groups, workspaces, onClose, onSave }) {
  const [value, setValue] = useState(modal.value || '')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const submit = async () => {
    setSaving(true); setError('')
    try { await onSave(modal.type, modal.row, value) } catch (err) { setError(formatAssetError(err)) } finally { setSaving(false) }
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

  if (modal.type === 'import') {
    return (
      <Modal title='导入主机' onClose={onClose}>
        <div className='fx-assets-form'>
          <Blocked>{ASSET_BLOCKERS.hostCreate}</Blocked>
          <p className='fx-assets-muted'>云导入和 Excel 导入需要凭据引用、字段映射、审计和回滚契约；当前不会生成临时主机或导入完成提示。</p>
          <footer><button type='button' onClick={onClose}>关闭</button></footer>
        </div>
      </Modal>
    )
  }

  if (modal.type === 'terminal-select') {
    return (
      <Modal title='选择终端主机' onClose={onClose}>
        <div className='fx-assets-form'>
          <Blocked>{ASSET_BLOCKERS.terminal}</Blocked>
          <p className='fx-assets-muted'>请在主机列表中选择具体主机后查看阻断详情；未接入 WebSSH 通道前不会建立连接。</p>
          <footer><button type='button' onClick={onClose}>关闭</button></footer>
        </div>
      </Modal>
    )
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

// C02: 主机详情内容
function HostDetailDrawer({ row }) {
  const basicInfo = [
    ['主机名', hostName(row)],
    ['IP 地址', hostIp(row)],
    ['操作系统', `${displayText(row.os)} ${displayText(row.arch, '')}`],
    ['CPU', displayText(row.cpu_cores || row.cpu, '-')],
    ['内存', formatMemory(row.memory_total || row.memory)],
    ['磁盘', formatDisk(row.disk_total || row.disk)],
    ['状态', isHostOnline(row) ? '在线' : '离线'],
    ['来源', displayText(row.source)],
    ['最近心跳', fmtTime(row.last_seen_at || row.last_seen || row.updated_at)],
  ]
  const connInfo = [
    ['SSH 端口', displayText(row.ssh_port || '22')],
    ['SSH 用户', displayText(row.ssh_user, '***')],
    ['密码/密钥', '******'],
  ]
  return (
    <div>
      <h3 style={{ margin: '0 0 8px', fontSize: 14, color: '#193a63' }}>基本信息</h3>
      <div className='fx-assets-table'><table><tbody>{basicInfo.map(([k, v]) => <tr key={k}><th style={{ width: 100 }}>{k}</th><td>{v}</td></tr>)}</tbody></table></div>
      <h3 style={{ margin: '12px 0 8px', fontSize: 14, color: '#193a63' }}>连接信息（脱敏）</h3>
      <div className='fx-assets-table'><table><tbody>{connInfo.map(([k, v]) => <tr key={k}><th style={{ width: 100 }}>{k}</th><td>{v}</td></tr>)}</tbody></table></div>
      <h3 style={{ margin: '12px 0 8px', fontSize: 14, color: '#193a63' }}>关联仪表盘</h3>
      <p className='fx-assets-muted'>暂无关联仪表盘。可在监控模块中配置主机仪表盘后自动关联。</p>
      <h3 style={{ margin: '12px 0 8px', fontSize: 14, color: '#193a63' }}>标签</h3>
      <Tags items={row.tags} />
    </div>
  )
}
