import React, { useEffect, useMemo, useState, useCallback } from 'react'
import { FxDrawer } from '../shared/FxDrawer.jsx'
import { get, post } from '../api/http.js'
import { displayText, fmtTime, hostIp, hostKey, hostName, isHostOnline } from './assetsModel.js'
import { ErrorBox, Feedback, Status } from './Shared.jsx'

const TABS = [
  { key: 'basic', label: '基础信息' },
  { key: 'monitor', label: '监控数据' },
  { key: 'relations', label: '关联关系' },
  { key: 'alerts', label: '告警历史' },
  { key: 'changes', label: '变更记录' },
  { key: 'agent', label: 'Agent状态' },
]

export function InstanceDetailDrawer({ row, open, onClose }) {
  const [activeTab, setActiveTab] = useState('basic')

  if (!row) return null

  return (
    <FxDrawer open={open} onClose={onClose} title={`实例详情 - ${hostName(row)}`} width='xl'>
      <div className='fx-instance-drawer'>
        <nav className='fx-instance-drawer-tabs'>
          {TABS.map(tab => (
            <button
              key={tab.key}
              type='button'
              className={activeTab === tab.key ? 'is-active' : ''}
              onClick={() => setActiveTab(tab.key)}
            >
              {tab.label}
            </button>
          ))}
        </nav>
        <div className='fx-instance-drawer-body'>
          {activeTab === 'basic' && <BasicInfoTab row={row} />}
          {activeTab === 'monitor' && <MonitorTab row={row} />}
          {activeTab === 'relations' && <RelationsTab row={row} />}
          {activeTab === 'alerts' && <AlertsTab row={row} />}
          {activeTab === 'changes' && <ChangesTab row={row} />}
          {activeTab === 'agent' && <AgentTab row={row} />}
        </div>
      </div>
    </FxDrawer>
  )
}

function BasicInfoTab({ row }) {
  const [editing, setEditing] = useState(false)
  const [form, setForm] = useState({})
  const [saving, setSaving] = useState(false)
  const [feedback, setFeedback] = useState('')

  const fields = useMemo(() => [
    ['主机名', hostName(row)],
    ['IP 地址', hostIp(row)],
    ['操作系统', displayText(row.os, '-')],
    ['架构', displayText(row.arch, '-')],
    ['CPU', displayText(row.cpu_cores || row.cpu, '-')],
    ['内存', formatBytes(row.memory_total || row.memory)],
    ['磁盘', formatBytes(row.disk_total || row.disk)],
    ['所属业务', displayText(row.business_name, '-')],
    ['类型', displayText(row.classification, '-')],
    ['子类型', displayText(row.subtype, '-')],
    ['负责人', displayText(row.owner, '-')],
    ['维护状态', displayText(row.maintenance_status, '正常')],
    ['采集情况', displayText(row.collection_status, '未监控')],
    ['告警级别', displayText(row.alert_level, 'none')],
    ['Agent ID', displayText(row.agent_id || row.agent_status, '-')],
    ['来源', displayText(row.source, '-')],
    ['最近心跳', fmtTime(row.last_seen_at || row.last_seen || row.updated_at)],
    ['创建时间', fmtTime(row.created_at)],
  ], [row])

  const handleSave = async () => {
    setSaving(true)
    try {
      await post(`/cmdb/instances/${encodeURIComponent(hostKey(row))}`, form)
      setFeedback('保存成功')
      setEditing(false)
    } catch (err) {
      setFeedback(`保存失败: ${err.message}`)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
        <h4 style={{ margin: 0, color: '#193a63' }}>属性信息</h4>
        {!editing && <button type='button' className='fx-assets-button' onClick={() => setEditing(true)}>编辑</button>}
        {editing && (
          <div style={{ display: 'flex', gap: 8 }}>
            <button type='button' className='fx-assets-button is-primary' disabled={saving} onClick={handleSave}>{saving ? '保存中...' : '保存'}</button>
            <button type='button' className='fx-assets-button' onClick={() => setEditing(false)}>取消</button>
          </div>
        )}
      </div>
      <Feedback>{feedback}</Feedback>
      <div className='fx-assets-table'>
        <table>
          <tbody>
            {fields.map(([label, value]) => (
              <tr key={label}>
                <th style={{ width: 120, background: '#f8fbff' }}>{label}</th>
                <td>{value}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function Sparkline({ data, color = '#1769ff', height = 40, width = 200 }) {
  if (!data || !data.length) return <span style={{ color: '#60728e', fontSize: 12 }}>暂无数据</span>
  const max = Math.max(...data, 1)
  const min = Math.min(...data, 0)
  const range = max - min || 1
  const points = data.map((v, i) => {
    const x = (i / (data.length - 1)) * width
    const y = height - ((v - min) / range) * (height - 4) - 2
    return `${x},${y}`
  }).join(' ')
  return (
    <svg width={width} height={height} style={{ display: 'block' }}>
      <polyline points={points} fill='none' stroke={color} strokeWidth='1.5' strokeLinejoin='round' />
    </svg>
  )
}

function MonitorTab({ row }) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let alive = true
    setLoading(true)
    get(`/cmdb/instances/${encodeURIComponent(hostKey(row))}/metrics`)
      .then(resp => {
        if (alive) { setData(resp?.data || resp); setLoading(false) }
      })
      .catch(err => {
        if (alive) { setError(err.message); setLoading(false) }
      })
    return () => { alive = false }
  }, [row])

  if (loading) return <div className='fx-assets-empty'>正在加载监控数据...</div>
  if (error) return <ErrorBox>{error}</ErrorBox>

  const cpu = data?.cpu || []
  const memory = data?.memory || []
  const disk = data?.disk || []
  const network = data?.network || []

  return (
    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
      <MetricCard title='CPU 使用率' data={cpu} color='#3b82f6' unit='%' />
      <MetricCard title='内存使用率' data={memory} color='#10b981' unit='%' />
      <MetricCard title='磁盘使用率' data={disk} color='#f59e0b' unit='%' />
      <MetricCard title='网络流量' data={network} color='#8b5cf6' unit='Mbps' />
    </div>
  )
}

function MetricCard({ title, data, color, unit }) {
  const latest = data.length ? data[data.length - 1] : null
  return (
    <div style={{ border: '1px solid #e6edf6', borderRadius: 8, padding: 12 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
        <span style={{ fontSize: 13, color: '#193a63', fontWeight: 600 }}>{title}</span>
        {latest !== null && <span style={{ fontSize: 12, color }}>{latest.toFixed(1)}{unit}</span>}
      </div>
      <Sparkline data={data} color={color} width={180} height={36} />
    </div>
  )
}

function RelationsTab({ row }) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let alive = true
    setLoading(true)
    get(`/cmdb/topology`, { params: { instance_id: hostKey(row) } })
      .then(resp => {
        if (alive) { setData(resp?.data || resp); setLoading(false) }
      })
      .catch(err => {
        if (alive) { setError(err.message); setLoading(false) }
      })
    return () => { alive = false }
  }, [row])

  if (loading) return <div className='fx-assets-empty'>正在加载关联关系...</div>
  if (error) return <ErrorBox>{error}</ErrorBox>

  const upstream = Array.isArray(data?.upstream) ? data.upstream : []
  const downstream = Array.isArray(data?.downstream) ? data.downstream : []

  return (
    <div>
      <h4 style={{ margin: '0 0 8px', color: '#193a63', fontSize: 14 }}>上游依赖 ({upstream.length})</h4>
      {upstream.length ? (
        <div className='fx-assets-table'>
          <table><thead><tr><th>名称</th><th>类型</th><th>IP</th><th>关系</th></tr></thead>
            <tbody>{upstream.map((item, i) => (
              <tr key={item.id || i}>
                <td>{displayText(item.name)}</td>
                <td>{displayText(item.type || item.object_name, '-')}</td>
                <td>{displayText(item.ip, '-')}</td>
                <td>{displayText(item.relation_name, '-')}</td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      ) : <div className='fx-assets-empty'>暂无上游依赖</div>}

      <h4 style={{ margin: '16px 0 8px', color: '#193a63', fontSize: 14 }}>下游服务 ({downstream.length})</h4>
      {downstream.length ? (
        <div className='fx-assets-table'>
          <table><thead><tr><th>名称</th><th>类型</th><th>IP</th><th>关系</th></tr></thead>
            <tbody>{downstream.map((item, i) => (
              <tr key={item.id || i}>
                <td>{displayText(item.name)}</td>
                <td>{displayText(item.type || item.object_name, '-')}</td>
                <td>{displayText(item.ip, '-')}</td>
                <td>{displayText(item.relation_name, '-')}</td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      ) : <div className='fx-assets-empty'>暂无下游服务</div>}
    </div>
  )
}

function AlertsTab({ row }) {
  const [alerts, setAlerts] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let alive = true
    setLoading(true)
    get(`/cmdb/instances/${encodeURIComponent(hostKey(row))}/alerts`)
      .then(resp => {
        if (alive) { setAlerts(Array.isArray(resp?.data || resp) ? (resp?.data || resp) : []); setLoading(false) }
      })
      .catch(err => {
        if (alive) { setError(err.message); setLoading(false) }
      })
    return () => { alive = false }
  }, [row])

  if (loading) return <div className='fx-assets-empty'>正在加载告警历史...</div>
  if (error) return <ErrorBox>{error}</ErrorBox>
  if (!alerts.length) return <div className='fx-assets-empty'>暂无告警记录</div>

  const levelColors = { critical: '#dc2626', warning: '#f59e0b', info: '#3b82f6' }

  return (
    <div className='fx-assets-table'>
      <table>
        <thead><tr><th>级别</th><th>告警内容</th><th>触发时间</th><th>恢复时间</th><th>状态</th></tr></thead>
        <tbody>{alerts.slice(0, 50).map((alert, i) => (
          <tr key={alert.id || i}>
            <td><span style={{ color: levelColors[alert.level] || '#6b7280', fontWeight: 600, fontSize: 12 }}>{displayText(alert.level, '-')}</span></td>
            <td>{displayText(alert.message || alert.content, '-')}</td>
            <td>{fmtTime(alert.triggered_at || alert.created_at)}</td>
            <td>{fmtTime(alert.resolved_at)}</td>
            <td><Status ok={alert.status === 'resolved'}>{alert.status === 'resolved' ? '已恢复' : '告警中'}</Status></td>
          </tr>
        ))}</tbody>
      </table>
    </div>
  )
}

function ChangesTab({ row }) {
  const [changes, setChanges] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let alive = true
    setLoading(true)
    get(`/cmdb/instances/${encodeURIComponent(hostKey(row))}/changes`)
      .then(resp => {
        if (alive) { setChanges(Array.isArray(resp?.data || resp) ? (resp?.data || resp) : []); setLoading(false) }
      })
      .catch(err => {
        if (alive) { setError(err.message); setLoading(false) }
      })
    return () => { alive = false }
  }, [row])

  if (loading) return <div className='fx-assets-empty'>正在加载变更记录...</div>
  if (error) return <ErrorBox>{error}</ErrorBox>
  if (!changes.length) return <div className='fx-assets-empty'>暂无变更记录</div>

  return (
    <div className='fx-assets-table'>
      <table>
        <thead><tr><th>变更字段</th><th>变更前</th><th>变更后</th><th>操作人</th><th>时间</th></tr></thead>
        <tbody>{changes.slice(0, 50).map((change, i) => (
          <tr key={change.id || i}>
            <td>{displayText(change.field || change.attribute, '-')}</td>
            <td style={{ color: '#dc2626' }}>{displayText(change.old_value || change.before, '-')}</td>
            <td style={{ color: '#10b981' }}>{displayText(change.new_value || change.after, '-')}</td>
            <td>{displayText(change.operator || change.user, '-')}</td>
            <td>{fmtTime(change.changed_at || change.created_at)}</td>
          </tr>
        ))}</tbody>
      </table>
    </div>
  )
}

function AgentTab({ row }) {
  const [state, setState] = useState({ loading: true, data: null, error: '' })
  const [deploying, setDeploying] = useState(false)
  const [feedback, setFeedback] = useState('')

  useEffect(() => {
    let alive = true
    setState({ loading: true, data: null, error: '' })
    get(`/cmdb/instances/${encodeURIComponent(hostKey(row))}/agent-status`)
      .then(resp => {
        if (alive) setState({ loading: false, data: resp?.data || resp, error: '' })
      })
      .catch(err => {
        if (alive) setState({ loading: false, data: null, error: err.message })
      })
    return () => { alive = false }
  }, [row])

  const handleDeploy = async () => {
    setDeploying(true)
    setFeedback('')
    try {
      await post(`/cmdb/instances/${encodeURIComponent(hostKey(row))}/deploy-config`)
      setFeedback('配置下发成功')
    } catch (err) {
      setFeedback(`下发失败: ${err.message}`)
    } finally {
      setDeploying(false)
    }
  }

  const handleAiDiagnose = async () => {
    setFeedback('AI排障功能已触发，请等待诊断结果...')
    try {
      await post(`/cmdb/instances/${encodeURIComponent(hostKey(row))}/ai-diagnose`)
      setFeedback('AI排障完成，请查看诊断报告。')
    } catch (err) {
      setFeedback(`AI排障失败: ${err.message}`)
    }
  }

  if (state.loading) return <div className='fx-assets-empty'>正在加载Agent状态...</div>
  if (state.error) return <ErrorBox>{state.error}</ErrorBox>

  const data = state.data || {}
  const plugins = Array.isArray(data.plugins) ? data.plugins : []

  return (
    <div>
      <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
        <button type='button' className='fx-assets-button is-primary' disabled={deploying} onClick={handleDeploy}>
          {deploying ? '下发中...' : '下发配置'}
        </button>
        <button type='button' className='fx-assets-button' onClick={handleAiDiagnose}>AI排障</button>
      </div>
      <Feedback>{feedback}</Feedback>

      <h4 style={{ margin: '12px 0 8px', color: '#193a63', fontSize: 14 }}>心跳状态</h4>
      <div className='fx-assets-table'>
        <table>
          <tbody>
            <tr><th style={{ width: 120 }}>Agent ID</th><td>{displayText(data.agent_id || row.agent_id, '-')}</td></tr>
            <tr><th>状态</th><td><Status ok={data.heartbeat_status === 'online' || isHostOnline(row)}>{data.heartbeat_status === 'online' || isHostOnline(row) ? '在线' : '离线'}</Status></td></tr>
            <tr><th>版本</th><td>{displayText(data.version, '-')}</td></tr>
            <tr><th>最近心跳</th><td>{fmtTime(data.last_heartbeat || row.last_seen_at)}</td></tr>
            <tr><th>启动时间</th><td>{fmtTime(data.started_at)}</td></tr>
          </tbody>
        </table>
      </div>

      <h4 style={{ margin: '16px 0 8px', color: '#193a63', fontSize: 14 }}>已启用插件 ({plugins.length})</h4>
      {plugins.length ? (
        <div className='fx-assets-table'>
          <table>
            <thead><tr><th>插件名称</th><th>版本</th><th>状态</th></tr></thead>
            <tbody>{plugins.map((p, i) => (
              <tr key={p.name || i}>
                <td>{displayText(p.name || p.plugin_name, '-')}</td>
                <td>{displayText(p.version, '-')}</td>
                <td><Status ok={p.status === 'running' || p.enabled}>{p.status === 'running' || p.enabled ? '运行中' : '已停止'}</Status></td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      ) : <div className='fx-assets-empty'>暂无已启用插件</div>}
    </div>
  )
}

function formatBytes(val) {
  if (!val) return '-'
  const num = Number(val)
  if (Number.isNaN(num)) return String(val)
  if (num > 1e12) return `${(num / 1099511627776).toFixed(1)} TB`
  if (num > 1e9) return `${(num / 1073741824).toFixed(1)} GB`
  if (num > 1e6) return `${(num / 1048576).toFixed(0)} MB`
  return String(val)
}
