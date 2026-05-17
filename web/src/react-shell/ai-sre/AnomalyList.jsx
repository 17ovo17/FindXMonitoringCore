import React, { useCallback, useEffect, useState } from 'react'
import { ErrorBox } from './AiSreShared.jsx'

/* ─── 严重程度配色 ─── */
const SEVERITY_STYLES = {
  critical: { bg: '#fef2f2', border: '#ef4444', text: '#dc2626' },
  high: { bg: '#fff7ed', border: '#f97316', text: '#ea580c' },
  medium: { bg: '#fefce8', border: '#eab308', text: '#ca8a04' },
  low: { bg: '#f0fdf4', border: '#22c55e', text: '#16a34a' },
}

/* ─── 异常类型标签 ─── */
const ANOMALY_TYPE_LABELS = {
  spike: '突增',
  drop: '突降',
  trend: '趋势',
  cyclic: '周期异常',
  correlated: '关联异常',
}

/* ─── 异常卡片 ─── */
function AnomalyCard({ anomaly, onInvestigate }) {
  const style = SEVERITY_STYLES[anomaly.severity] || SEVERITY_STYLES.low
  return (
    <div
      className='fx-anomaly-card'
      style={{ backgroundColor: style.bg, borderLeft: `4px solid ${style.border}` }}
    >
      <div className='fx-anomaly-card-header'>
        <span className='fx-anomaly-metric'>{anomaly.metric_name}</span>
        <span className='fx-anomaly-type-badge'>{ANOMALY_TYPE_LABELS[anomaly.type] || anomaly.type}</span>
        <span className='fx-anomaly-severity' style={{ color: style.text }}>{anomaly.severity}</span>
      </div>
      <div className='fx-anomaly-card-body'>
        <p className='fx-anomaly-description'>{anomaly.description}</p>
        <div className='fx-anomaly-values'>
          <span>当前值: <strong>{Number(anomaly.current_value).toFixed(2)}</strong></span>
          <span>期望值: <strong>{Number(anomaly.expected_value).toFixed(2)}</strong></span>
          <span>偏差: <strong>{Number(anomaly.deviation).toFixed(1)}σ</strong></span>
        </div>
      </div>
      <div className='fx-anomaly-card-footer'>
        <span className='fx-anomaly-time'>{anomaly.started_at || ''}</span>
        <button type='button' className='fx-anomaly-investigate-btn' onClick={() => onInvestigate(anomaly)}>
          启动调查
        </button>
      </div>
    </div>
  )
}

/* ─── 主组件 ─── */
export function AnomalyList({ onNavigate, addEvidence }) {
  const [anomalies, setAnomalies] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [investigating, setInvestigating] = useState(null)

  const loadAnomalies = useCallback(async () => {
    setLoading(true)
    try {
      const resp = await fetch('/api/v1/ai/anomalies')
      const data = await resp.json()
      if (data.code !== 0) throw new Error(data.error || '加载失败')
      setAnomalies(data.data?.anomalies || [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadAnomalies() }, [loadAnomalies])

  const handleInvestigate = async (anomaly) => {
    setInvestigating(anomaly.id)
    setError('')
    try {
      const resp = await fetch('/api/v1/ai/investigate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: 'anomaly',
          data: { metric_name: anomaly.metric_name, anomaly_id: anomaly.id },
          context: anomaly.labels || {},
        }),
      })
      const data = await resp.json()
      if (data.code !== 0) throw new Error(data.error || '调查启动失败')
      if (addEvidence) {
        addEvidence({
          category: 'anomaly',
          title: `异常调查: ${anomaly.metric_name}`,
          detail: data.data?.conclusion?.root_cause || '调查进行中',
        })
      }
    } catch (err) {
      setError(err.message)
    } finally {
      setInvestigating(null)
    }
  }

  // 按严重程度排序
  const sortedAnomalies = [...anomalies].sort((a, b) => {
    const order = { critical: 0, high: 1, medium: 2, low: 3 }
    return (order[a.severity] ?? 4) - (order[b.severity] ?? 4)
  })

  return (
    <section className='fx-anomaly-section'>
      <div className='fx-anomaly-header'>
        <h3>异常检测</h3>
        <button type='button' onClick={loadAnomalies} disabled={loading} className='fx-anomaly-refresh-btn'>
          {loading ? '刷新中...' : '刷新'}
        </button>
      </div>
      <ErrorBox>{error}</ErrorBox>
      {!loading && anomalies.length === 0 && (
        <p className='fx-anomaly-empty'>当前无异常检测结果。系统每 30 秒自动检测一次。</p>
      )}
      <div className='fx-anomaly-list'>
        {sortedAnomalies.map(anomaly => (
          <AnomalyCard
            key={anomaly.id}
            anomaly={anomaly}
            onInvestigate={handleInvestigate}
          />
        ))}
      </div>
      {investigating && <p className='fx-anomaly-investigating'>正在对异常进行自主调查...</p>}
    </section>
  )
}
