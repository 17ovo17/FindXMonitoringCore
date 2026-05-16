import React, { useEffect, useMemo, useState } from 'react'
import { PROBE_BLOCKERS, formatProbeError, probesApi } from '../api/probes.js'
import './probes.css'

const sections = [
  { value: 'public', label: '状态页', desc: '展示业务拨测状态、90 天运行证据和人工事故。' },
  { value: 'config', label: '拨测配置', desc: '管理 HTTP、TCP、PING、DNS 检查项和通知/告警绑定。' },
  { value: 'incidents', label: '事故维护', desc: '维护状态页事故时间线和恢复记录。' },
]
const sectionSet = new Set(sections.map(item => item.value))
const checkTypes = ['http', 'tcp', 'ping', 'dns']

function SectionTabs({ section, onNavigate }) {
  return (
    <div className='fx-probe-tabs'>
      {sections.map(item => (
        <button key={item.value} type='button' className={section === item.value ? 'is-active' : ''} onClick={() => onNavigate({ section: item.value })}>
          {item.label}
        </button>
      ))}
    </div>
  )
}

function Blocked({ children }) {
  return <div className='fx-probe-blocked'><strong>PENDING</strong><span>{children}</span></div>
}

function ErrorBox({ children }) {
  return <div className='fx-probe-error'>{children}</div>
}

const statusLabel = (status) => ({
  up: '正常',
  operational: '正常',
  degraded: '受影响',
  down: '故障',
  disabled: '已停用',
  no_data: '暂无数据',
  unknown: '未知',
}[status] || '未知')

const incidentStatusLabel = (status) => ({
  investigating: '调查中',
  identified: '已定位',
  monitoring: '观察中',
  resolved: '已恢复',
}[status] || status || '未知')

const statusTone = (status) => {
  if (status === 'up' || status === 'operational') return 'is-ok'
  if (status === 'degraded') return 'is-warn'
  if (status === 'down') return 'is-bad'
  if (status === 'disabled') return 'is-muted'
  return 'is-unknown'
}

const formatPercent = (value) => typeof value === 'number' ? `${value.toFixed(3)}%` : '暂无真实数据'
const formatMs = (value) => typeof value === 'number' ? `${value}ms` : '暂无真实数据'
const currentTimeText = (value) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '暂无更新时间'

function UptimeBars({ bars = [] }) {
  const nextBars = bars.length ? bars : Array.from({ length: 90 }, (_, index) => ({ date: `day-${index}`, status: 'no_data' }))
  return (
    <div className='fx-uptime-bars' aria-label='90 天可用性条'>
      {nextBars.slice(-90).map((bar, index) => <span key={`${bar.date}-${index}`} className={`fx-uptime-bar ${statusTone(bar.status)}`} title={`${bar.date}: ${statusLabel(bar.status)}`} />)}
    </div>
  )
}

function SummaryCard({ label, value, hint }) {
  return <article className='fx-probe-summary-card'><span>{label}</span><strong>{value}</strong>{hint && <small>{hint}</small>}</article>
}

function StatusHero({ data }) {
  const status = data?.status || 'unknown'
  const isReady = (status === 'up' || status === 'operational') && data?.summary?.has_run_evidence
  const headline = isReady ? '所有启用拨测均有正常运行证据' : status === 'down' ? '业务拨测发现故障' : status === 'degraded' ? '部分业务拨测受影响' : '业务拨测状态未知'
  return (
    <section className={`fx-probe-hero ${statusTone(status)}`}>
      <div className='fx-probe-hero-icon'>{isReady ? '✓' : '!'}</div>
      <div>
        <h2>{headline}</h2>
        <p>{data?.status_reason || '状态页只基于真实拨测运行记录和人工事故维护聚合。'}</p>
        <small>最后更新：{currentTimeText(data?.updated_at)} · 共 {data?.summary?.total_checks || 0} 个监测项 · 运行中 {data?.summary?.running_checks || 0}</small>
      </div>
    </section>
  )
}

function SubscriptionStrip() {
  const blockedSubscriptions = [
    ['webhook', 'Webhook', 'Webhook 订阅需要通知回执、签名和退订审计契约'],
    ['email', '邮件订阅', '邮件订阅需要退订、防刷和投递回执契约'],
    ['rss', 'RSS', 'RSS 输出需要公开订阅和缓存刷新契约'],
  ]
  return (
    <section className='fx-probe-subscribe'>
      <span>订阅状态更新，第一时间收到事件通知。</span>
      <div>
        {blockedSubscriptions.map(([key, label, title]) => (
          <button key={key} type='button' disabled aria-disabled='true' title={title}>
            {label}
            <small>待配置</small>
          </button>
        ))}
      </div>
    </section>
  )
}

function StatusPageSection() {
  const [data, setData] = useState(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  const load = async () => {
    setLoading(true)
    setError('')
    try {
      setData(await probesApi.statusPage('main'))
    } catch (err) {
      setError(formatProbeError(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const summary = data?.summary || {}
  return (
    <div className='fx-probe-stack'>
      {loading && <div className='fx-probe-empty'>正在读取业务拨测状态页...</div>}
      {error && <ErrorBox>{error}</ErrorBox>}
      {data && <>
        <StatusHero data={data} />
        <SubscriptionStrip />
        <Blocked>{PROBE_BLOCKERS.subscription}</Blocked>
        <section className='fx-probe-summary-grid'>
          <SummaryCard label='90 天可用性' value={formatPercent(summary.uptime_90d)} hint={summary.has_run_evidence ? '基于真实运行记录' : '无运行记录不计算'} />
          <SummaryCard label='运行中监测' value={`${summary.running_checks || 0}/${summary.total_checks || 0}`} />
          <SummaryCard label='近 30 天平均响应' value={formatMs(summary.average_response_30d_ms)} />
          <SummaryCard label='近 90 天事件' value={summary.incident_count_90d || 0} />
        </section>
        {!summary.has_run_evidence && <div className='fx-probe-empty'>{summary.missing_run_evidence_note || '暂无真实拨测运行记录。'}</div>}
        <section className='fx-probe-status-groups'>
          {(data.groups || []).map(group => (
            <article key={group.id} className='fx-probe-group'>
              <header><h3>{group.name}</h3><span>近 90 天</span></header>
              {(group.checks || []).length === 0 && <div className='fx-probe-empty'>该分组暂无拨测检查项。</div>}
              {(group.checks || []).map(check => (
                <div key={check.id} className='fx-probe-service-row'>
                  <div className='fx-probe-service-main'>
                    <strong><span className={`fx-probe-dot ${statusTone(check.status)}`} />{check.name}</strong>
                    <span className={`fx-probe-pill ${statusTone(check.status)}`}>{statusLabel(check.status)}</span>
                  </div>
                  <UptimeBars bars={check.status_bar} />
                  <div className='fx-probe-service-meta'>
                    <span>{formatPercent(check.uptime_90d)} 可用性</span>
                    <span>{formatMs(check.response_time_ms)} 响应</span>
                    <span>24h/7d/30d/90d 仅在真实运行记录存在后计算</span>
                  </div>
                </div>
              ))}
            </article>
          ))}
        </section>
        <IncidentTimeline incidents={data.incidents || []} />
      </>}
    </div>
  )
}

function IncidentTimeline({ incidents }) {
  return (
    <section className='fx-probe-panel'>
      <header><h3>事件时间线</h3><span>人工维护 + 后续自动拨测事件</span></header>
      {incidents.length === 0 && <div className='fx-probe-empty'>近 90 天没有人工事故记录；自动告警和拨测事件仍等待契约接入。</div>}
      {incidents.map(item => (
        <article key={item.id} className='fx-probe-incident'>
          <strong>{item.title}</strong>
          <span className={`fx-probe-pill ${item.status === 'resolved' ? 'is-ok' : 'is-warn'}`}>{incidentStatusLabel(item.status)}</span>
          <p>{item.message || '未填写事件说明'}</p>
          <small>{currentTimeText(item.started_at)}</small>
        </article>
      ))}
    </section>
  )
}

const emptyCheck = {
  name: '',
  type: 'http',
  url: '',
  target: '',
  port: '',
  interval_seconds: 60,
  timeout_seconds: 10,
  retries: 0,
  enabled: true,
}

function ConfigSection() {
  const [checks, setChecks] = useState([])
  const [selected, setSelected] = useState(null)
  const [form, setForm] = useState(emptyCheck)
  const [filter, setFilter] = useState('')
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(true)

  const load = async () => {
    setLoading(true)
    setError('')
    try {
      setChecks(await probesApi.checks({ q: filter }))
    } catch (err) {
      setError(formatProbeError(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const openForm = (check = null) => {
    setSelected(check)
    setForm(check ? { ...emptyCheck, ...check, port: check.port || '' } : emptyCheck)
    setBlocked('')
  }

  const save = async (event) => {
    event.preventDefault()
    setError('')
    const payload = { ...form, port: Number(form.port) || 0 }
    try {
      if (selected?.id) await probesApi.updateCheck(selected.id, payload)
      else await probesApi.createCheck(payload)
      setSelected(null)
      setForm(emptyCheck)
      await load()
    } catch (err) {
      setError(formatProbeError(err))
    }
  }

  const toggle = async (check) => {
    try {
      if (check.enabled) await probesApi.disableCheck(check.id)
      else await probesApi.enableCheck(check.id)
      await load()
    } catch (err) {
      setError(formatProbeError(err))
    }
  }

  const runTest = async (check) => {
    setBlocked('')
    try {
      await probesApi.testCheck(check.id)
      setBlocked('测试执行返回了非阻断结果，请确认后端是否已经接入真实执行器和 evidence。')
    } catch (err) {
      setBlocked(formatProbeError(err))
    }
  }

  return (
    <div className='fx-probe-grid-page'>
      <section className='fx-probe-panel'>
        <header>
          <div><h3>拨测检查项</h3><span>配置可保存，运行成功必须等待真实执行器回执。</span></div>
          <button type='button' onClick={() => openForm()}>新增检查</button>
        </header>
        <div className='fx-probe-toolbar'>
          <input value={filter} placeholder='搜索名称、目标或业务组' onChange={event => setFilter(event.target.value)} />
          <button type='button' onClick={load}>查询</button>
        </div>
        {loading && <div className='fx-probe-empty'>正在读取检查项...</div>}
        {error && <ErrorBox>{error}</ErrorBox>}
        {blocked && <Blocked>{blocked}</Blocked>}
        <div className='fx-probe-table'>
          <div className='fx-probe-table-head'><span>名称</span><span>类型</span><span>目标</span><span>状态</span><span>操作</span></div>
          {checks.length === 0 && <div className='fx-probe-empty'>暂无检查项，新增配置后仍需真实执行器产生运行记录。</div>}
          {checks.map(check => (
            <div key={check.id} className='fx-probe-table-row'>
              <span>{check.name}</span>
              <span>{check.type?.toUpperCase()}</span>
              <span>{check.url || `${check.target}${check.port ? `:${check.port}` : ''}`}</span>
              <span className={`fx-probe-pill ${statusTone(check.status)}`}>{statusLabel(check.status)}</span>
              <span className='fx-probe-row-actions'>
                <button type='button' onClick={() => openForm(check)}>编辑</button>
                <button type='button' onClick={() => toggle(check)}>{check.enabled ? '停用' : '启用'}</button>
                <button type='button' onClick={() => runTest(check)}>测试</button>
              </span>
            </div>
          ))}
        </div>
      </section>
      <section className='fx-probe-panel'>
        <header><h3>{selected ? '编辑检查项' : '新增检查项'}</h3><span>保存配置不代表拨测成功</span></header>
        <ProbeForm form={form} setForm={setForm} onSubmit={save} />
        <Blocked>{PROBE_BLOCKERS.dryRun}</Blocked>
        <BindingPanel checks={checks} />
      </section>
    </div>
  )
}

function ProbeForm({ form, setForm, onSubmit }) {
  const update = (key, value) => setForm(current => ({ ...current, [key]: value }))
  return (
    <form className='fx-probe-form' onSubmit={onSubmit}>
      <label>名称<input value={form.name} onChange={event => update('name', event.target.value)} required /></label>
      <label>类型<select value={form.type} onChange={event => update('type', event.target.value)}>{checkTypes.map(type => <option key={type} value={type}>{type.toUpperCase()}</option>)}</select></label>
      {form.type === 'http' ? <label>URL<input value={form.url || ''} onChange={event => update('url', event.target.value)} placeholder='https://example.com/health' required /></label> : <label>目标<input value={form.target || ''} onChange={event => update('target', event.target.value)} placeholder='host 或域名' required /></label>}
      {['tcp', 'dns'].includes(form.type) && <label>端口<input type='number' min='0' value={form.port || ''} onChange={event => update('port', event.target.value)} /></label>}
      <label>间隔秒<input type='number' min='15' max='86400' value={form.interval_seconds} onChange={event => update('interval_seconds', Number(event.target.value))} /></label>
      <label>超时秒<input type='number' min='1' max='120' value={form.timeout_seconds} onChange={event => update('timeout_seconds', Number(event.target.value))} /></label>
      <label>重试次数<input type='number' min='0' max='10' value={form.retries} onChange={event => update('retries', Number(event.target.value))} /></label>
      <label className='fx-probe-check'><input type='checkbox' checked={!!form.enabled} onChange={event => update('enabled', event.target.checked)} />启用配置</label>
      <button type='submit'>保存配置</button>
    </form>
  )
}

function BindingPanel({ checks }) {
  const first = checks[0]
  const [message, setMessage] = useState('')
  const saveBlockedBinding = async (kind) => {
    setMessage('')
    if (!first) return setMessage('请先创建检查项。')
    try {
      if (kind === 'notification') {
        const res = await probesApi.saveNotificationBindings(first.id, { items: [{ channel_id: 'notify-console', enabled: true }] })
        setMessage(res?.capability?.message || PROBE_BLOCKERS.subscription)
      } else {
        const res = await probesApi.saveAlertBindings(first.id, { items: [{ alert_rule_id: 'probe-alert-rule', enabled: true }] })
        setMessage(res?.capability?.message || '告警生命周期仍为阻断契约。')
      }
    } catch (err) {
      setMessage(formatProbeError(err))
    }
  }
  return (
    <div className='fx-probe-binding'>
      <h4>通知与告警绑定</h4>
      <p>绑定配置可以保存，真实投递、恢复、压缩和升级必须等待后端回执契约。</p>
      <div>
        <button type='button' onClick={() => saveBlockedBinding('notification')}>保存通知绑定</button>
        <button type='button' onClick={() => saveBlockedBinding('alert')}>保存告警绑定</button>
      </div>
      {message && <Blocked>{message}</Blocked>}
    </div>
  )
}

function IncidentsSection() {
  const [items, setItems] = useState([])
  const [form, setForm] = useState({ title: '', status: 'investigating', severity: 'p2', message: '' })
  const [error, setError] = useState('')

  const load = async () => {
    try { setItems(await probesApi.incidents()) } catch (err) { setError(formatProbeError(err)) }
  }
  useEffect(() => { load() }, [])

  const save = async (event) => {
    event.preventDefault()
    setError('')
    try {
      await probesApi.createIncident(form)
      setForm({ title: '', status: 'investigating', severity: 'p2', message: '' })
      await load()
    } catch (err) {
      setError(formatProbeError(err))
    }
  }

  return (
    <div className='fx-probe-grid-page'>
      <section className='fx-probe-panel'>
        <header><h3>事故维护</h3><span>人工事件会进入状态页时间线。</span></header>
        {error && <ErrorBox>{error}</ErrorBox>}
        <form className='fx-probe-form' onSubmit={save}>
          <label>标题<input value={form.title} onChange={event => setForm({ ...form, title: event.target.value })} required /></label>
          <label>状态<select value={form.status} onChange={event => setForm({ ...form, status: event.target.value })}>{['investigating', 'identified', 'monitoring', 'resolved'].map(value => <option key={value} value={value}>{value}</option>)}</select></label>
          <label>级别<input value={form.severity} onChange={event => setForm({ ...form, severity: event.target.value })} /></label>
          <label>说明<textarea value={form.message} onChange={event => setForm({ ...form, message: event.target.value })} /></label>
          <button type='submit'>记录事件</button>
        </form>
      </section>
      <IncidentTimeline incidents={items} />
    </div>
  )
}

export function BusinessProbePage({ query, onNavigate }) {
  const section = sectionSet.has(query?.section) ? query.section : 'public'
  const meta = useMemo(() => sections.find(item => item.value === section) || sections[0], [section])
  const navigate = next => onNavigate({ ...query, ...next })
  return (
    <main className='fx-probe-page'>
      <header className='fx-probe-head'>
        <div><p>业务拨测</p><h1>{meta.label}</h1><span>{meta.desc}</span></div>
        <SectionTabs section={section} onNavigate={navigate} />
      </header>
      {section === 'public' && <StatusPageSection />}
      {section === 'config' && <ConfigSection />}
      {section === 'incidents' && <IncidentsSection />}
    </main>
  )
}
