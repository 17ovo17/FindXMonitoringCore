import React, { useCallback, useEffect, useRef, useState } from 'react'
import { get } from '../api/http.js'
import './overview.css'

const REFRESH_INTERVAL = 30000

function StatCard({ label, value, color, icon }) {
  return (
    <div className="fx-overview-stat" style={{ borderTopColor: color }}>
      <div className="fx-overview-stat__icon" style={{ color }}>{icon}</div>
      <div className="fx-overview-stat__body">
        <span className="fx-overview-stat__value" style={{ color }}>{value}</span>
        <span className="fx-overview-stat__label">{label}</span>
      </div>
    </div>
  )
}

function HealthDonut({ healthy, warning, critical }) {
  const total = healthy + warning + critical || 1
  const hPct = (healthy / total) * 100
  const wPct = (warning / total) * 100
  const cPct = (critical / total) * 100

  const hAngle = (healthy / total) * 360
  const wAngle = (warning / total) * 360

  const toArc = (startAngle, endAngle, radius, cx, cy) => {
    const startRad = ((startAngle - 90) * Math.PI) / 180
    const endRad = ((endAngle - 90) * Math.PI) / 180
    const x1 = cx + radius * Math.cos(startRad)
    const y1 = cy + radius * Math.sin(startRad)
    const x2 = cx + radius * Math.cos(endRad)
    const y2 = cy + radius * Math.sin(endRad)
    const largeArc = endAngle - startAngle > 180 ? 1 : 0
    return `M ${x1} ${y1} A ${radius} ${radius} 0 ${largeArc} 1 ${x2} ${y2}`
  }

  const r = 40
  const cx = 50
  const cy = 50

  return (
    <div className="fx-overview-donut">
      <svg viewBox="0 0 100 100" className="fx-overview-donut__svg">
        {healthy > 0 && (
          <path d={toArc(0, hAngle, r, cx, cy)} fill="none" stroke="#2ea043" strokeWidth="12" strokeLinecap="round" />
        )}
        {warning > 0 && (
          <path d={toArc(hAngle, hAngle + wAngle, r, cx, cy)} fill="none" stroke="#d4a017" strokeWidth="12" strokeLinecap="round" />
        )}
        {critical > 0 && (
          <path d={toArc(hAngle + wAngle, 360, r, cx, cy)} fill="none" stroke="#c5352b" strokeWidth="12" strokeLinecap="round" />
        )}
      </svg>
      <div className="fx-overview-donut__legend">
        <span className="fx-overview-donut__item"><i style={{ background: '#2ea043' }} />健康 {hPct.toFixed(0)}%</span>
        <span className="fx-overview-donut__item"><i style={{ background: '#d4a017' }} />警告 {wPct.toFixed(0)}%</span>
        <span className="fx-overview-donut__item"><i style={{ background: '#c5352b' }} />严重 {cPct.toFixed(0)}%</span>
      </div>
    </div>
  )
}

function TopAlertHosts({ hosts }) {
  if (!hosts.length) return <div className="fx-overview-empty">暂无告警主机</div>
  return (
    <table className="fx-overview-table">
      <thead>
        <tr>
          <th>主机名</th>
          <th>指标</th>
          <th>当前值</th>
          <th>级别</th>
        </tr>
      </thead>
      <tbody>
        {hosts.map((h, i) => (
          <tr key={i}>
            <td>{h.hostname || h.host || '-'}</td>
            <td>{h.metric || '-'}</td>
            <td>{h.value != null ? h.value : '-'}</td>
            <td>
              <span className={`fx-overview-severity fx-overview-severity--${h.severity || 'warning'}`}>
                {h.severity === 'critical' ? '严重' : h.severity === 'warning' ? '警告' : '信息'}
              </span>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

function RecentEvents({ events }) {
  if (!events.length) return <div className="fx-overview-empty">暂无告警事件</div>
  return (
    <ul className="fx-overview-events">
      {events.map((ev, i) => (
        <li key={i} className="fx-overview-event">
          <span className="fx-overview-event__time">{ev.time || ev.trigger_time || '-'}</span>
          <span className="fx-overview-event__rule">{ev.rule_name || ev.rule || '-'}</span>
          <span className="fx-overview-event__host">{ev.hostname || ev.host || '-'}</span>
        </li>
      ))}
    </ul>
  )
}

export function OverviewPage({ query, onNavigate }) {
  const [stats, setStats] = useState({ critical: 0, warning: 0, total: 0, healthy_pct: 0 })
  const [topHosts, setTopHosts] = useState([])
  const [recentEvents, setRecentEvents] = useState([])
  const [health, setHealth] = useState({ healthy: 0, warning: 0, critical: 0 })
  const timerRef = useRef(null)

  const fetchData = useCallback(async () => {
    try {
      const [alertRes, hostRes] = await Promise.allSettled([
        get('/alert/events', { params: { limit: 10, status: 'firing' } }),
        get('/assets/hosts', { params: { limit: 100 } }),
      ])

      const events = alertRes.status === 'fulfilled' ? (Array.isArray(alertRes.value) ? alertRes.value : alertRes.value?.items || alertRes.value?.data || []) : []
      const hosts = hostRes.status === 'fulfilled' ? (Array.isArray(hostRes.value) ? hostRes.value : hostRes.value?.items || hostRes.value?.data || []) : []

      const criticalEvents = events.filter((e) => e.severity === 'critical' || e.level === 'critical')
      const warningEvents = events.filter((e) => e.severity === 'warning' || e.level === 'warning')

      const totalResources = hosts.length
      const healthyHosts = hosts.filter((h) => h.status === 'healthy' || h.status === 'online' || h.state === 'up')
      const warningHosts = hosts.filter((h) => h.status === 'warning')
      const criticalHosts = hosts.filter((h) => h.status === 'critical' || h.status === 'offline' || h.state === 'down')
      const healthyPct = totalResources > 0 ? Math.round((healthyHosts.length / totalResources) * 100) : 0

      setStats({
        critical: criticalEvents.length,
        warning: warningEvents.length,
        total: totalResources,
        healthy_pct: healthyPct,
      })

      setHealth({
        healthy: healthyHosts.length,
        warning: warningHosts.length,
        critical: criticalHosts.length,
      })

      const sorted = [...events]
        .filter((e) => e.severity === 'critical' || e.level === 'critical' || e.severity === 'warning' || e.level === 'warning')
        .slice(0, 5)
        .map((e) => ({
          hostname: e.hostname || e.host || e.target || '-',
          metric: e.metric || e.rule_name || '-',
          value: e.value || e.current_value || '-',
          severity: e.severity || e.level || 'warning',
        }))
      setTopHosts(sorted)

      setRecentEvents(events.slice(0, 10).map((e) => ({
        time: e.trigger_time || e.time || e.created_at || '-',
        rule_name: e.rule_name || e.rule || '-',
        hostname: e.hostname || e.host || e.target || '-',
      })))
    } catch {
      // 静默处理，保持现有数据
    }
  }, [])

  useEffect(() => {
    fetchData()
    timerRef.current = setInterval(fetchData, REFRESH_INTERVAL)
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [fetchData])

  const quickActions = [
    { label: '查告警', action: () => onNavigate({ path: '/alerts', query: { section: 'events' } }) },
    { label: '查主机', action: () => onNavigate({ path: '/assets', query: { section: 'hosts' } }) },
    { label: '查日志', action: () => onNavigate({ path: '/logs', query: { section: 'query' } }) },
    { label: '触发巡检', action: () => onNavigate({ path: '/aiops', query: { section: 'health' } }) },
  ]

  return (
    <div className="fx-overview-page">
      <div className="fx-overview-stats">
        <StatCard label="严重告警" value={stats.critical} color="#c5352b" icon="!" />
        <StatCard label="警告告警" value={stats.warning} color="#d4a017" icon="!" />
        <StatCard label="资源总数" value={stats.total} color="var(--fx-blue)" icon="#" />
        <StatCard label="健康率" value={`${stats.healthy_pct}%`} color="#2ea043" icon="%" />
      </div>

      <div className="fx-overview-middle">
        <div className="fx-overview-panel">
          <h3 className="fx-overview-panel__title">TOP5 告警主机</h3>
          <TopAlertHosts hosts={topHosts} />
        </div>
        <div className="fx-overview-panel">
          <h3 className="fx-overview-panel__title">资源健康分布</h3>
          <HealthDonut healthy={health.healthy} warning={health.warning} critical={health.critical} />
        </div>
      </div>

      <div className="fx-overview-bottom">
        <div className="fx-overview-panel">
          <h3 className="fx-overview-panel__title">最近告警事件</h3>
          <RecentEvents events={recentEvents} />
        </div>
        <div className="fx-overview-panel">
          <h3 className="fx-overview-panel__title">AI 快捷操作</h3>
          <div className="fx-overview-actions">
            {quickActions.map((a) => (
              <button key={a.label} type="button" className="fx-overview-action-btn" onClick={a.action}>
                {a.label}
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

