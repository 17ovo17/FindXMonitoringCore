import React, { useEffect, useMemo, useState } from 'react'
import { notificationsApi } from '../../api/notifications.js'
import { Modal } from './NotificationsPage.jsx'
import { displayDate, displayJson, filterText, makeError, normalizeChannel, normalizeRule, normalizeTemplate, parseJson, parseLines } from './notificationModel.js'

const emptyDraft = { id: '', name: '', description: '', enabled: true, channelId: '', templateId: '', receivers: '', severities: 'critical,warning', alertRuleIds: '', conditions: '{}', timeWindow: '{}' }
const emptyRoute = { matchType: 'label', matchKey: '', matchOp: 'eq', matchValue: '', channelId: '', timeRange: '', bizGroup: '' }

function toDraft(rule) {
  const cfg = rule?.notifyConfigs?.[0] || {}
  return rule ? {
    id: rule.id,
    name: rule.name,
    description: rule.description,
    enabled: rule.enabled,
    channelId: cfg.channel_id || '',
    templateId: cfg.template_id || '',
    receivers: (cfg.receivers || []).join('\n'),
    severities: (cfg.severities || []).join(','),
    alertRuleIds: rule.alertRuleIds.join('\n'),
    conditions: displayJson(rule.conditions || {}),
    timeWindow: displayJson(rule.timeWindow || {}),
    routes: rule.raw?.routes || [],
  } : { ...emptyDraft, routes: [] }
}

function payloadFromDraft(draft) {
  return {
    id: draft.id || undefined,
    name: draft.name.trim(),
    description: draft.description.trim(),
    enabled: draft.enabled,
    alert_rule_ids: parseLines(draft.alertRuleIds),
    conditions: parseJson(draft.conditions, {}),
    time_window: parseJson(draft.timeWindow, {}),
    routes: (draft.routes || []).filter((r) => r.channelId),
    notify_configs: [{
      channel_id: draft.channelId,
      template_id: draft.templateId,
      receivers: parseLines(draft.receivers),
      severities: parseLines(draft.severities),
    }],
  }
}

function validateDryRunPayload(payload) {
  const errors = []
  const notifyConfig = payload?.notify_configs?.[0] || payload?.notify_config || {}
  if (!String(payload?.name || '').trim()) errors.push('名称不能为空')
  if (!String(notifyConfig.channel_id || '').trim()) errors.push('通知媒介不能为空')
  if (!Array.isArray(notifyConfig.receivers) || notifyConfig.receivers.length === 0) errors.push('接收方不能为空')
  if (!Array.isArray(notifyConfig.severities) || notifyConfig.severities.length === 0) errors.push('告警级别不能为空')
  return errors
}

function AdvancedRouting({ routes, onChange, channels }) {
  const addRoute = () => onChange([...routes, { ...emptyRoute }])
  const removeRoute = (idx) => onChange(routes.filter((_, i) => i !== idx))
  const updateRoute = (idx, key, val) => onChange(routes.map((r, i) => i === idx ? { ...r, [key]: val } : r))

  return (
    <div className='fx-notify-routing'>
      <div className='fx-notify-routing__head'>
        <strong>高级路由规则</strong>
        <button type='button' onClick={addRoute}>+ 添加路由</button>
      </div>
      {routes.length === 0 && <p className='fx-notify-routing__hint'>未配置高级路由，所有告警将发送到默认媒介。</p>}
      {routes.map((route, idx) => (
        <div key={idx} className='fx-notify-routing__item'>
          <div className='fx-notify-form'>
            <label><span>匹配类型</span>
              <select value={route.matchType} onChange={(e) => updateRoute(idx, 'matchType', e.target.value)}>
                <option value='label'>按标签</option>
                <option value='time'>按时间段</option>
                <option value='bizGroup'>按业务组</option>
              </select>
            </label>
            {route.matchType === 'label' && <>
              <label><span>标签 Key</span><input value={route.matchKey} placeholder='如 severity, service' onChange={(e) => updateRoute(idx, 'matchKey', e.target.value)} /></label>
              <label><span>匹配方式</span>
                <select value={route.matchOp} onChange={(e) => updateRoute(idx, 'matchOp', e.target.value)}>
                  <option value='eq'>等于</option><option value='neq'>不等于</option><option value='regex'>正则</option><option value='in'>包含</option>
                </select>
              </label>
              <label><span>匹配值</span><input value={route.matchValue} placeholder='critical' onChange={(e) => updateRoute(idx, 'matchValue', e.target.value)} /></label>
            </>}
            {route.matchType === 'time' && <>
              <label><span>时间段</span>
                <select value={route.timeRange} onChange={(e) => updateRoute(idx, 'timeRange', e.target.value)}>
                  <option value=''>请选择</option>
                  <option value='workday_09_18'>工作日 09:00-18:00</option>
                  <option value='workday_18_09'>工作日 18:00-次日09:00</option>
                  <option value='weekend'>周末全天</option>
                  <option value='holiday'>节假日</option>
                  <option value='always'>全时段</option>
                </select>
              </label>
            </>}
            {route.matchType === 'bizGroup' && <>
              <label><span>业务组</span><input value={route.bizGroup} placeholder='输入业务组名称' onChange={(e) => updateRoute(idx, 'bizGroup', e.target.value)} /></label>
            </>}
            <label><span>路由到媒介</span>
              <select value={route.channelId} onChange={(e) => updateRoute(idx, 'channelId', e.target.value)}>
                <option value=''>请选择</option>
                {channels.map((ch) => <option key={ch.id} value={ch.id}>{ch.name}</option>)}
              </select>
            </label>
          </div>
          <button type='button' className='is-danger fx-notify-routing__remove' onClick={() => removeRoute(idx)}>移除</button>
        </div>
      ))}
    </div>
  )
}

function RuleForm({ draft, setDraft, channels, templates, saving, onSubmit, onTest, onClose }) {
  return (
    <Modal title={draft.id ? '编辑通知规则' : '新建通知规则'} onClose={onClose}>
      <div className='fx-notify-form-sections'>
        <section><h3>基础信息</h3><div className='fx-notify-form'>
          <label><span>名称</span><input value={draft.name} onChange={(e) => setDraft({ ...draft, name: e.target.value })} /></label>
          <label className='fx-notify-check'><input type='checkbox' checked={draft.enabled} onChange={(e) => setDraft({ ...draft, enabled: e.target.checked })} />启用</label>
          <label className='is-wide'><span>描述</span><textarea rows={2} value={draft.description} onChange={(e) => setDraft({ ...draft, description: e.target.value })} /></label>
        </div></section>
        <section><h3>通知配置（默认）</h3><div className='fx-notify-form'>
          <label><span>媒介</span><select value={draft.channelId} onChange={(e) => setDraft({ ...draft, channelId: e.target.value })}><option value=''>请选择</option>{channels.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
          <label><span>模板</span><select value={draft.templateId} onChange={(e) => setDraft({ ...draft, templateId: e.target.value })}><option value=''>可选</option>{templates.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
          <label><span>级别</span><input value={draft.severities} onChange={(e) => setDraft({ ...draft, severities: e.target.value })} /></label>
          <label className='is-wide'><span>接收方</span><textarea rows={2} value={draft.receivers} onChange={(e) => setDraft({ ...draft, receivers: e.target.value })} /></label>
        </div></section>
        <section><h3>高级路由</h3>
          <AdvancedRouting routes={draft.routes || []} onChange={(routes) => setDraft({ ...draft, routes })} channels={channels} />
        </section>
        <section><h3>关联与条件</h3><div className='fx-notify-form'>
          <label className='is-wide'><span>关联告警规则 ID</span><textarea rows={2} value={draft.alertRuleIds} onChange={(e) => setDraft({ ...draft, alertRuleIds: e.target.value })} /></label>
          <label className='is-wide'><span>条件 JSON</span><textarea rows={4} value={draft.conditions} onChange={(e) => setDraft({ ...draft, conditions: e.target.value })} /></label>
          <label className='is-wide'><span>时间窗口 JSON</span><textarea rows={3} value={draft.timeWindow} onChange={(e) => setDraft({ ...draft, timeWindow: e.target.value })} /></label>
        </div></section>
      </div>
      <div className='fx-notify-actions'><button type='button' onClick={onClose}>取消</button><button type='button' onClick={onTest}>dry-run 测试</button><button type='button' className='is-primary' disabled={saving} onClick={onSubmit}>{saving ? '保存中...' : '保存'}</button></div>
    </Modal>
  )
}

function ImportRulesModal({ onClose, onImport }) {
  const [text, setText] = useState('')
  const [preview, setPreview] = useState(null)
  const [error, setError] = useState('')
  const fileRef = React.useRef(null)

  const parse = (raw) => {
    setError('')
    try {
      const data = JSON.parse(raw)
      const items = Array.isArray(data) ? data : [data]
      if (items.length === 0) { setError('JSON 为空数组'); return }
      setPreview(items)
    } catch { setError('JSON 格式不合法') }
  }

  const handleFile = (e) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = (ev) => { const content = ev.target.result; setText(content); parse(content) }
    reader.readAsText(file)
  }

  return (
    <Modal title='导入通知规则' onClose={onClose}>
      <div className='fx-notify-form'>
        <label className='is-wide'><span>粘贴 JSON 配置</span>
          <textarea rows={8} value={text} placeholder='[{"name":"...","notify_configs":[...]}]' onChange={(e) => setText(e.target.value)} />
        </label>
        <label className='is-wide'><span>或上传 JSON 文件</span><input ref={fileRef} type='file' accept='.json' onChange={handleFile} /></label>
      </div>
      {error && <div className='fx-notify-message is-error'>{error}</div>}
      {preview && <div style={{ margin: '8px 0' }}><strong>预览 ({preview.length} 条)</strong><pre>{displayJson(preview)}</pre></div>}
      <div className='fx-notify-actions'>
        <button type='button' onClick={onClose}>取消</button>
        {!preview && <button type='button' className='is-primary' onClick={() => parse(text)}>解析预览</button>}
        {preview && <button type='button' className='is-primary' onClick={() => onImport(preview)}>确认导入</button>}
      </div>
    </Modal>
  )
}

export function NotificationRulesSection() {
  const [rules, setRules] = useState([])
  const [channels, setChannels] = useState([])
  const [templates, setTemplates] = useState([])
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState('')
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [draft, setDraft] = useState(null)
  const [deleteTarget, setDeleteTarget] = useState(null)
  const [modal, setModal] = useState(null)
  const [showImport, setShowImport] = useState(false)

  const load = async () => {
    setLoading(true); setError('')
    try {
      const [ruleRows, channelRows, templateRows] = await Promise.all([notificationsApi.listRules(), notificationsApi.listChannels(), notificationsApi.listTemplates()])
      setRules(ruleRows.map(normalizeRule).filter((item) => item.id))
      setChannels(channelRows.map(normalizeChannel).filter((item) => item.id))
      setTemplates(templateRows.map(normalizeTemplate).filter((item) => item.id))
    } catch (err) { setError(makeError(err, '通知规则加载失败')) }
    finally { setLoading(false) }
  }

  useEffect(() => { load() }, [])

  const filtered = useMemo(() => rules.filter((item) => {
    if (status && String(item.enabled) !== status) return false
    return filterText([item.name, item.description, item.alertRuleIds.join(' ')], keyword)
  }), [rules, status, keyword])

  const save = async () => {
    setSaving(true); setError('')
    try {
      const body = payloadFromDraft(draft)
      if (draft.id) await notificationsApi.updateRule(draft.id, body)
      else await notificationsApi.saveRule(body)
      setDraft(null); await load()
    } catch (err) { setError(makeError(err, '规则保存失败')) }
    finally { setSaving(false) }
  }

  const dryRun = async (source) => {
    try {
      const rulePayload = source?.id ? source.raw : payloadFromDraft(draft)
      const errors = validateDryRunPayload(rulePayload)
      if (errors.length > 0) { setModal({ title: 'dry-run 前端校验失败', body: `请先补齐有效配置后再执行契约验证：\n${errors.map((item) => `- ${item}`).join('\n')}` }); return }
      const body = source?.id ? { rule_id: source.id, notify_config: source.raw.notify_configs?.[0] || {} } : { rule_id: draft?.id || 'draft', notify_config: rulePayload.notify_configs[0] }
      const result = await notificationsApi.testRule(body)
      const summary = result?.dry_run ? '契约验证通过，当前为 dry-run，外部投递已禁用。' : '后端未返回 dry_run=true，请人工确认契约。'
      setModal({ title: 'dry-run 测试结果', body: `${summary}\n\n${displayJson(result)}` })
    } catch (err) {
      if (err?.message === 'JSON 格式不合法') { setModal({ title: 'dry-run 前端校验失败', body: '条件 JSON 或时间窗口 JSON 格式不合法，请修正后再执行契约验证。' }); return }
      setModal({ title: 'dry-run 测试失败', body: makeError(err) })
    }
  }

  const rowAction = async (action, rule) => {
    try {
      if (action === 'edit') { setDraft(toDraft(normalizeRule(await notificationsApi.getRule(rule.id)))); return }
      if (action === 'enable') await notificationsApi.enableRule(rule.id)
      if (action === 'disable') await notificationsApi.disableRule(rule.id)
      if (action === 'clone') await notificationsApi.cloneRule(rule.id)
      if (action === 'test') { await dryRun(rule); return }
      if (action === 'detail') {
        const [statistics, events, alertRules, subRules] = await Promise.all([notificationsApi.getRuleStatistics(rule.id), notificationsApi.getRuleEvents(rule.id), notificationsApi.getRuleAlertRules(rule.id), notificationsApi.getRuleSubAlertRules(rule.id)])
        setModal({ title: '统计与关联详情', body: displayJson({ statistics, events, alert_rules: alertRules, sub_alert_rules: subRules }) }); return
      }
      if (action === 'delete') { setDeleteTarget(rule); return }
      await load()
    } catch (err) { setModal({ title: '操作失败', body: makeError(err) }) }
  }

  const remove = async () => {
    try { await notificationsApi.deleteRules([deleteTarget.id]); setDeleteTarget(null); await load() }
    catch (err) { setModal({ title: '删除失败', body: makeError(err) }) }
  }

  const handleImport = async (items) => {
    try {
      for (const item of items) await notificationsApi.saveRule(item)
      setShowImport(false); await load()
    } catch (err) { setError(makeError(err, '导入失败')) }
  }

  return (
    <section className='fx-notify-section'>
      <div className='fx-notify-filterbar'>
        <button type='button' disabled={loading} onClick={load}>{loading ? '刷新中...' : '刷新'}</button>
        <input value={keyword} onChange={(e) => setKeyword(e.target.value)} placeholder='搜索规则、描述、关联告警' />
        <select value={status} onChange={(e) => setStatus(e.target.value)}><option value=''>全部状态</option><option value='true'>启用</option><option value='false'>停用</option></select>
        <button type='button' className='is-primary' onClick={() => setDraft(toDraft(null))}>新建</button>
        <button type='button' onClick={() => setShowImport(true)}>导入</button>
      </div>
      {error && <div className='fx-notify-message is-error'>{error}</div>}
      <div className='fx-notify-table'>
        <table><thead><tr><th>状态</th><th>名称</th><th>通知媒介</th><th>关联告警</th><th>更新人</th><th>更新时间</th><th>操作</th></tr></thead>
          <tbody>{filtered.map((item) => <tr key={item.id}>
            <td><button type='button' className={item.enabled ? 'fx-notify-state is-on' : 'fx-notify-state'} onClick={() => rowAction(item.enabled ? 'disable' : 'enable', item)}>{item.enabled ? '启用' : '停用'}</button></td>
            <td><button type='button' className='fx-notify-link' onClick={() => rowAction('edit', item)}>{item.name}</button><small>{item.id}</small></td>
            <td>{item.notifyConfigs.map((cfg) => cfg.channel || cfg.channel_id).filter(Boolean).join(', ') || '-'}</td><td>{item.alertRuleIds.join(', ') || '-'}</td><td>{item.updatedBy}</td><td>{displayDate(item.updatedAt)}</td>
            <td><select onChange={(e) => { const value = e.target.value; e.target.value = ''; if (value) rowAction(value, item) }}><option value=''>更多</option><option value='edit'>编辑</option><option value='test'>dry-run 测试</option><option value='detail'>统计 / 关联</option><option value='clone'>克隆</option><option value='delete'>删除</option></select></td>
          </tr>)}</tbody></table>
        {!loading && filtered.length === 0 && <div className='fx-notify-empty'>暂无规则</div>}
      </div>
      {draft && <RuleForm draft={draft} setDraft={setDraft} channels={channels} templates={templates} saving={saving} onSubmit={save} onTest={() => dryRun()} onClose={() => setDraft(null)} />}
      {showImport && <ImportRulesModal onClose={() => setShowImport(false)} onImport={handleImport} />}
      {deleteTarget && <Modal title='确认删除通知规则' narrow onClose={() => setDeleteTarget(null)}><p>确认删除 {deleteTarget.name}？</p><div className='fx-notify-actions'><button type='button' onClick={() => setDeleteTarget(null)}>取消</button><button type='button' className='is-danger' onClick={remove}>确认删除</button></div></Modal>}
      {modal && <Modal title={modal.title} onClose={() => setModal(null)}><pre>{modal.body}</pre></Modal>}
    </section>
  )
}
