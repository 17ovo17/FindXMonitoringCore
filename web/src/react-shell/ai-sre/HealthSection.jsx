import React, { useCallback, useEffect, useRef, useState } from 'react'
import { AISRE_BLOCKERS, aiSreApi, formatAiSreError } from '../api/aiSre.js'
import { Blocked, Empty, ErrorBox, StatusPill } from './AiSreShared.jsx'

const healthCheckItems = [
  { key: 'datasources', label: '数据源连通性', loader: aiSreApi.health.datasources },
  { key: 'agents', label: 'Agent 存活', loader: aiSreApi.health.agents },
  { key: 'aiProviders', label: '服务健康', loader: aiSreApi.health.aiProviders },
  { key: 'prometheus', label: 'Prometheus 可达性', loader: aiSreApi.health.prometheus },
  { key: 'storage', label: '存储健康', loader: aiSreApi.health.storage },
]

const statusTone = status => {
  if (status === 'pass') return 'ok'
  if (status === 'warn') return 'warn'
  return 'blocked'
}

const statusLabel = status => {
  if (status === 'pass') return '正常'
  if (status === 'warn') return '警告'
  return '异常'
}

function HistoryTable({ history }) {
  if (!history.length) return <Empty>暂无检查历史</Empty>
  return (
    <div className='fx-aisre-health-history'>
      <table className='fx-aisre-table'>
        <thead>
          <tr><th>检查项</th><th>时间</th><th>状态</th><th>延迟</th></tr>
        </thead>
        <tbody>
          {history.map((row, i) => (
            <tr key={`${row.key}-${row.checkedAt}-${i}`}>
              <td>{row.label || row.key}</td>
              <td>{row.checkedAt}</td>
              <td><StatusPill tone={statusTone(row.status)}>{statusLabel(row.status)}</StatusPill></td>
              <td>{row.latency != null ? `${row.latency}ms` : '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export function HealthSection({ addEvidence }) {
  const [results, setResults] = useState({})
  const [history, setHistory] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [autoRefresh, setAutoRefresh] = useState(false)
  const timerRef = useRef(null)

  const runCheck = useCallback(async (item) => {
    const start = Date.now()
    try {
      await item.loader()
      const latency = Date.now() - start
      const entry = { key: item.key, label: item.label, status: 'pass', latency, checkedAt: new Date().toLocaleString('zh-CN', { hour12: false }) }
      setResults(prev => ({ ...prev, [item.key]: entry }))
      setHistory(prev => [entry, ...prev].slice(0, 10))
    } catch (err) {
      const latency = Date.now() - start
      const status = err?.status === 404 || err?.status === 501 ? 'fail' : 'warn'
      const entry = { key: item.key, label: item.label, status, latency, checkedAt: new Date().toLocaleString('zh-CN', { hour12: false }), error: formatAiSreError(err) }
      setResults(prev => ({ ...prev, [item.key]: entry }))
      setHistory(prev => [entry, ...prev].slice(0, 10))
    }
  }, [])

  const runAll = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      await Promise.allSettled(healthCheckItems.map(item => runCheck(item)))
      addEvidence({ category: 'metrics', title: '全部健康检查已执行', detail: `${healthCheckItems.length} 项` })
    } catch (err) {
      setError(formatAiSreError(err))
    } finally {
      setLoading(false)
    }
  }, [runCheck, addEvidence])

  useEffect(() => { runAll() }, [])

  useEffect(() => {
    if (autoRefresh) {
      timerRef.current = setInterval(runAll, 60000)
    } else {
      if (timerRef.current) clearInterval(timerRef.current)
    }
    return () => { if (timerRef.current) clearInterval(timerRef.current) }
  }, [autoRefresh, runAll])

  return (
    <section className='fx-aisre-panel'>
      <div className='fx-aisre-toolbar'>
        <h2>健康检查</h2>
        <button type='button' onClick={runAll} disabled={loading}>{loading ? '检查中...' : '全部检查'}</button>
        <label className='fx-aisre-toggle-label'>
          <input type='checkbox' checked={autoRefresh} onChange={e => setAutoRefresh(e.target.checked)} />
          自动刷新 (60s)
        </label>
      </div>
      <ErrorBox>{error}</ErrorBox>
      <div className='fx-aisre-health-grid'>
        {healthCheckItems.map(item => {
          const result = results[item.key]
          return (
            <article key={item.key} className='fx-aisre-health-card'>
              <div className='fx-aisre-health-card-head'>
                <h3>{item.label}</h3>
                <button type='button' onClick={() => runCheck(item)} className='fx-aisre-btn-sm'>检查</button>
              </div>
              <div className='fx-aisre-health-card-body'>
                <StatusPill tone={result ? statusTone(result.status) : 'blocked'}>
                  {result ? statusLabel(result.status) : '未检查'}
                </StatusPill>
                {result?.latency != null && <span className='fx-aisre-health-latency'>{result.latency}ms</span>}
                {result?.checkedAt && <span className='fx-aisre-health-time'>{result.checkedAt}</span>}
              </div>
              {result?.error && <ErrorBox>{result.error}</ErrorBox>}
            </article>
          )
        })}
      </div>
      <article className='fx-aisre-card' style={{ marginTop: 12 }}>
        <h3>数据到达</h3>
        <Blocked>{AISRE_BLOCKERS.dataArrival}</Blocked>
      </article>
      <div style={{ marginTop: 16 }}>
        <h3>检查历史 (最近 10 条)</h3>
        <HistoryTable history={history} />
      </div>
    </section>
  )
}
