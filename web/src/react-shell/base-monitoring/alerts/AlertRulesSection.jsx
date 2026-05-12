import React, { useEffect, useMemo, useState } from 'react'
import { alertsApi } from '../../api/alerts.js'
import {
  displayDate,
  displayJson,
  filterText,
  getUniqueOptions,
  makeError,
  mapToPairs,
  normalizeRule,
  normalizeRuleDetail,
  rulePayload,
  severities,
  severityLabel,
} from './alertModel.js'
import { RuleFormModal } from './form/RuleFormModal.jsx'
import { BatchRuleActions } from './form/BatchRuleActions.jsx'
import { BusinessGroupTree } from './form/BusinessGroupTree.jsx'
import { ColumnConfigModal, allColumns, getVisibleColumns } from './form/ColumnConfig.jsx'
import { Pagination } from '../../shared/ConfirmModal.jsx'

const PAGE_SIZE = 20

const emptyDraft = {
  id: '',
  name: '',
  datasourceId: '',
  query: '',
  severity: 'warning',
  forDuration: '5m',
  noDataPolicy: 'keep_state',
  enabled: true,
  targetSelector: '{}',
  labels: '{}',
  annotations: '{}',
  effective_time: { enable_status: true, days_of_week: [1, 2, 3, 4, 5, 6, 0], time_ranges: [] },
  notify_config: { notify_channels: [], notify_groups: [], callbacks: [], notify_template_id: '', notify_repeat_step: 60, notify_max_number: 0, notify_recovered: true },
  pipeline_config: { relabel_configs: [], annotations: [], enrich_queries: [] },
  triggers_config: { triggers: [{ severity: 'warning', operator: '>', value: 0, for_duration: '5m' }], nodata_trigger: { enable: false, severity: 'warning', action: 'keep_state' }, recover_config: { enable: true, recover_duration: 0 } },
}

function ruleToDraft(rule) {
  return {
    id: rule.id,
    name: rule.name,
    datasourceId: rule.datasourceId === '-' ? '' : rule.datasourceId,
    query: rule.query,
    severity: rule.severity,
    forDuration: rule.forDuration,
    noDataPolicy: rule.noDataPolicy,
    enabled: rule.enabled,
    targetSelector: displayJson(rule.targetSelector),
    labels: displayJson(rule.labels),
    annotations: displayJson(rule.annotations),
    effective_time: rule.raw?.effective_time || emptyDraft.effective_time,
    notify_config: rule.raw?.notify_config || emptyDraft.notify_config,
    pipeline_config: rule.raw?.pipeline_config || emptyDraft.pipeline_config,
    triggers_config: rule.raw?.triggers_config || emptyDraft.triggers_config,
  }
}

function ConfirmDeleteModal({ rule, onCancel, onConfirm, deleting }) {
  return (
    <div className='fx-alert-modal'>
      <div className='fx-alert-modal__body is-confirm'>
        <header><h2>确认删除规则</h2><button type='button' onClick={onCancel}>关闭</button></header>
        <p>删除后将无法从当前页面恢复，请确认要删除规则：<strong>{rule.name}</strong></p>
        <div className='fx-alert-actions'>
          <button type='button' onClick={onCancel}>取消</button>
          <button type='button' className='is-danger' disabled={deleting} onClick={onConfirm}>{deleting ? '删除中...' : '确认删除'}</button>
        </div>
      </div>
    </div>
  )
}

export function AlertRulesSection() {
  const [rules, setRules] = useState([])
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState('')
  const [severity, setSeverity] = useState('')
  const [datasourceId, setDatasourceId] = useState('')
  const [businessGroup, setBusinessGroup] = useState('')
  const [selectedIds, setSelectedIds] = useState([])
  const [visibleCols, setVisibleCols] = useState(() => getVisibleColumns())
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState('')
  const [modal, setModal] = useState(null)
  const [deleteTarget, setDeleteTarget] = useState(null)
  const [draft, setDraft] = useState(null)
  const [formError, setFormError] = useState('')
  const [showColumnConfig, setShowColumnConfig] = useState(false)
  const [page, setPage] = useState(1)

  const loadRules = async () => {
    setLoading(true); setError('')
    try {
      const rows = await alertsApi.listRules(status ? { status } : {})
      setRules(rows.map(normalizeRule).filter((row) => row.id))
      setSelectedIds([])
    } catch (err) {
      setError(makeError(err, '规则加载失败'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { loadRules() }, [status])

  const datasourceOptions = useMemo(() => getUniqueOptions(rules, 'datasourceId'), [rules])
  const filtered = useMemo(() => {
    setPage(1)
    return rules.filter((rule) => {
    if (severity && rule.severity !== severity) return false
    if (datasourceId && rule.datasourceId !== datasourceId) return false
    if (businessGroup && rule.businessGroup !== businessGroup) return false
    return filterText([rule.name, rule.datasourceId, rule.businessGroup, rule.category, rule.query, ...mapToPairs(rule.labels)], keyword)
  })}, [rules, keyword, severity, datasourceId, businessGroup])

  const paged = useMemo(() => filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE), [filtered, page])

  const selectedRules = useMemo(() => filtered.filter((r) => selectedIds.includes(r.id)), [filtered, selectedIds])

  const openEdit = async (rule) => {
    setFormError('')
    try {
      const detail = normalizeRule(normalizeRuleDetail(await alertsApi.getRule(rule.id)))
      setDraft(ruleToDraft(detail))
    } catch (err) {
      setModal({ title: '详情加载失败', body: makeError(err) })
    }
  }

  const submitDraft = async () => {
    setSaving(true); setFormError('')
    try {
      const body = {
        ...rulePayload(draft),
        effective_time: draft.effective_time,
        notify_config: draft.notify_config,
        pipeline_config: draft.pipeline_config,
        triggers_config: draft.triggers_config,
      }
      const saved = normalizeRule(draft.id ? await alertsApi.updateRule(draft.id, body) : await alertsApi.createRule(body))
      setRules((rows) => [saved, ...rows.filter((row) => row.id !== saved.id)])
      setDraft(null)
    } catch (err) {
      setFormError(makeError(err, '保存失败'))
    } finally {
      setSaving(false)
    }
  }

  const tryRunDraft = async () => {
    try {
      const body = rulePayload(draft)
      const result = await alertsApi.tryRunRule(draft.id || 'draft', body)
      setModal({ title: '试运行结果', body: displayJson(result) })
    } catch (err) {
      setModal({ title: '试运行失败', body: makeError(err) })
    }
  }

  const confirmDelete = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await alertsApi.removeRule(deleteTarget.id)
      setDeleteTarget(null)
      await loadRules()
    } catch (err) {
      setModal({ title: '删除失败', body: makeError(err) })
    } finally {
      setDeleting(false)
    }
  }

  const rowAction = async (action, rule) => {
    try {
      if (action === 'enable') await alertsApi.enableRule(rule.id)
      if (action === 'disable') await alertsApi.disableRule(rule.id)
      if (action === 'clone') {
        const cloned = normalizeRule(await alertsApi.cloneRule(rule.id))
        setRules((rows) => [cloned, ...rows])
        return
      }
      if (action === 'delete') { setDeleteTarget(rule); return }
      if (action === 'tryrun') {
        const result = await alertsApi.tryRunRule(rule.id, {})
        setModal({ title: '试运行结果', body: displayJson(result) })
        return
      }
      await loadRules()
    } catch (err) {
      setModal({ title: '操作失败', body: makeError(err) })
    }
  }

  const toggleSelect = (id, checked) => {
    setSelectedIds((ids) => checked ? [...ids, id] : ids.filter((i) => i !== id))
  }

  const toggleSelectAll = (checked) => {
    setSelectedIds(checked ? filtered.map((r) => r.id) : [])
  }

  const isColVisible = (key) => visibleCols.includes(key)

  return (
    <section className='fx-alert-section fx-alert-section--with-tree'>
      <BusinessGroupTree selectedGroup={businessGroup} onSelect={setBusinessGroup} />
      <div className='fx-alert-section-main'>
        <div className='fx-alert-filterbar'>
          <button type='button' disabled={loading} onClick={loadRules}>{loading ? '刷新中...' : '刷新'}</button>
          <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder='搜索规则、业务组、标签、查询语句' />
          <select value={datasourceId} onChange={(e) => setDatasourceId(e.target.value)}>
            <option value=''>全部数据源</option>
            {datasourceOptions.map((item) => <option key={item} value={item}>{item}</option>)}
          </select>
          <select value={severity} onChange={(e) => setSeverity(e.target.value)}><option value=''>全部级别</option>{severities.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select>
          <select value={status} onChange={(e) => setStatus(e.target.value)}><option value=''>全部状态</option><option value='active'>启用</option><option value='disabled'>停用</option></select>
          <button type='button' className='is-primary' onClick={() => { setFormError(''); setDraft(emptyDraft) }}>新建</button>
          <button type='button' onClick={() => setShowColumnConfig(true)}>列配置</button>
        </div>
        <BatchRuleActions
          selectedIds={selectedIds}
          selectedRules={selectedRules}
          onRefresh={loadRules}
          onMessage={(msg) => setModal({ title: '提示', body: msg })}
        />
        {error && <div className='fx-alert-message is-error'>{error}</div>}
        <div className='fx-alert-table'>
          <table>
            <thead>
              <tr>
                <th><input type='checkbox' checked={selectedIds.length === filtered.length && filtered.length > 0} onChange={(e) => toggleSelectAll(e.target.checked)} /></th>
                {isColVisible('status') && <th>状态</th>}
                {isColVisible('name') && <th>名称</th>}
                {isColVisible('businessGroup') && <th>业务组</th>}
                {isColVisible('category') && <th>分类</th>}
                {isColVisible('datasourceId') && <th>数据源</th>}
                {isColVisible('severity') && <th>级别</th>}
                {isColVisible('labels') && <th>标签</th>}
                {isColVisible('updatedAt') && <th>更新时间</th>}
                {isColVisible('updatedBy') && <th>更新人</th>}
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {paged.map((rule) => (
                <tr key={rule.id}>
                  <td><input type='checkbox' checked={selectedIds.includes(rule.id)} onChange={(e) => toggleSelect(rule.id, e.target.checked)} /></td>
                  {isColVisible('status') && <td><button type='button' className={rule.enabled ? 'fx-alert-state is-on' : 'fx-alert-state'} onClick={() => rowAction(rule.enabled ? 'disable' : 'enable', rule)}>{rule.enabled ? '启用' : '停用'}</button></td>}
                  {isColVisible('name') && <td><button type='button' className='fx-alert-link' onClick={() => openEdit(rule)}>{rule.name}</button><small>{rule.id}</small></td>}
                  {isColVisible('businessGroup') && <td>{rule.businessGroup}</td>}
                  {isColVisible('category') && <td>{rule.category}</td>}
                  {isColVisible('datasourceId') && <td>{rule.datasourceId}</td>}
                  {isColVisible('severity') && <td><span className={`fx-alert-severity is-${rule.severity}`}>{severityLabel(rule.severity)}</span></td>}
                  {isColVisible('labels') && <td>{mapToPairs(rule.labels).slice(0, 3).map((item) => <span className='fx-alert-tag' key={item}>{item}</span>)}</td>}
                  {isColVisible('updatedAt') && <td>{displayDate(rule.updatedAt)}</td>}
                  {isColVisible('updatedBy') && <td>{rule.updatedBy}</td>}
                  <td><select onChange={(e) => { if (e.target.value) rowAction(e.target.value, rule); e.target.value = '' }}><option value=''>更多</option><option value='tryrun'>试运行</option><option value='clone'>克隆</option><option value='delete'>删除</option></select></td>
                </tr>
              ))}
            </tbody>
          </table>
          {!loading && filtered.length === 0 && <div className='fx-alert-empty'>暂无规则</div>}
          <Pagination total={filtered.length} page={page} pageSize={PAGE_SIZE} onPageChange={setPage} />
        </div>
      </div>
      {draft && <RuleFormModal draft={draft} setDraft={setDraft} saving={saving} error={formError} onSubmit={submitDraft} onTryRun={tryRunDraft} onClose={() => setDraft(null)} />}
      {deleteTarget && <ConfirmDeleteModal rule={deleteTarget} deleting={deleting} onCancel={() => setDeleteTarget(null)} onConfirm={confirmDelete} />}
      {showColumnConfig && <ColumnConfigModal onClose={() => setShowColumnConfig(false)} onSave={setVisibleCols} />}
      {modal && <div className='fx-alert-modal'><div className='fx-alert-modal__body'><header><h2>{modal.title}</h2><button type='button' onClick={() => setModal(null)}>关闭</button></header><pre>{modal.body}</pre></div></div>}
    </section>
  )
}


