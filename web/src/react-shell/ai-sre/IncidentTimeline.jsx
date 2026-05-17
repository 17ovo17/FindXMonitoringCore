import React, { useCallback, useEffect, useState } from 'react'
import { ErrorBox } from './AiSreShared.jsx'

/* ─── 阶段配置 ─── */
const PHASES = [
  { key: 'detected', label: '检测', icon: '1' },
  { key: 'classified', label: '分类', icon: '2' },
  { key: 'notified', label: '通知', icon: '3' },
  { key: 'diagnosed', label: '诊断', icon: '4' },
  { key: 'remediated', label: '修复', icon: '5' },
  { key: 'verified', label: '验证', icon: '6' },
  { key: 'reviewed', label: '复盘', icon: '7' },
  { key: 'learned', label: '学习', icon: '8' },
]

const SEVERITY_COLORS = {
  critical: '#ef4444',
  high: '#f97316',
  medium: '#eab308',
  low: '#22c55e',
}

/* ─── 进度条组件 ─── */
function PhaseProgress({ currentPhase }) {
  const currentIdx = PHASES.findIndex(p => p.key === currentPhase)
  return (
    <div className='fx-incident-progress'>
      {PHASES.map((phase, idx) => {
        let status = 'pending'
        if (idx < currentIdx) status = 'completed'
        else if (idx === currentIdx) status = 'active'
        return (
          <div key={phase.key} className={`fx-incident-phase-step fx-incident-phase-${status}`}>
            <div className='fx-incident-phase-icon'>{phase.icon}</div>
            <div className='fx-incident-phase-label'>{phase.label}</div>
          </div>
        )
      })}
    </div>
  )
}

/* ─── 时间线条目 ─── */
function TimelineItem({ entry }) {
  return (
    <div className='fx-incident-timeline-item'>
      <div className='fx-incident-timeline-dot' />
      <div className='fx-incident-timeline-content'>
        <div className='fx-incident-timeline-action'>{entry.action}</div>
        <div className='fx-incident-timeline-detail'>{entry.detail}</div>
        <div className='fx-incident-timeline-meta'>
          <span className='fx-incident-timeline-actor'>{entry.actor === 'ai' ? 'AI' : entry.actor}</span>
          <span className='fx-incident-timeline-time'>{entry.timestamp || ''}</span>
        </div>
      </div>
    </div>
  )
}

/* ─── 复盘报告卡片 ─── */
function PostMortemCard({ postMortem }) {
  if (!postMortem) return null
  return (
    <div className='fx-incident-postmortem'>
      <h4>复盘报告</h4>
      <div className='fx-incident-postmortem-section'>
        <strong>摘要：</strong>
        <p>{postMortem.summary}</p>
      </div>
      <div className='fx-incident-postmortem-section'>
        <strong>影响：</strong>
        <p>{postMortem.impact}</p>
      </div>
      <div className='fx-incident-postmortem-section'>
        <strong>根因：</strong>
        <p>{postMortem.root_cause}</p>
      </div>
      {postMortem.action_items && postMortem.action_items.length > 0 && (
        <div className='fx-incident-postmortem-section'>
          <strong>改进项：</strong>
          <ul>
            {postMortem.action_items.map((item, i) => <li key={i}>{item}</li>)}
          </ul>
        </div>
      )}
      <div className='fx-incident-postmortem-section'>
        <strong>预防计划：</strong>
        <p>{postMortem.prevention_plan}</p>
      </div>
    </div>
  )
}

/* ─── 事件详情卡片 ─── */
function IncidentDetail({ incident, onAdvance, onBack }) {
  const [advancing, setAdvancing] = useState(false)

  const handleAdvance = async () => {
    setAdvancing(true)
    await onAdvance(incident.id)
    setAdvancing(false)
  }

  const canAdvance = incident.phase !== 'learned'

  return (
    <div className='fx-incident-detail'>
      <div className='fx-incident-detail-header'>
        <button type='button' onClick={onBack} className='fx-incident-back-btn'>← 返回列表</button>
        <h3>{incident.title}</h3>
        <span className='fx-incident-severity' style={{ color: SEVERITY_COLORS[incident.severity] || '#666' }}>
          {incident.severity}
        </span>
      </div>

      <PhaseProgress currentPhase={incident.phase} />

      {canAdvance && (
        <div className='fx-incident-actions'>
          <button type='button' onClick={handleAdvance} disabled={advancing} className='fx-incident-advance-btn'>
            {advancing ? '推进中...' : '推进到下一阶段'}
          </button>
        </div>
      )}

      <div className='fx-incident-timeline'>
        <h4>事件时间线</h4>
        {(incident.timeline || []).map((entry, idx) => (
          <TimelineItem key={idx} entry={entry} />
        ))}
      </div>

      {incident.post_mortem && <PostMortemCard postMortem={incident.post_mortem} />}

      {incident.mttr_seconds > 0 && (
        <div className='fx-incident-mttr'>
          <strong>MTTR：</strong>{Math.round(incident.mttr_seconds / 60)} 分钟
        </div>
      )}
    </div>
  )
}

/* ─── 主组件 ─── */
export function IncidentTimeline() {
  const [incidents, setIncidents] = useState([])
  const [selectedId, setSelectedId] = useState(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const loadIncidents = useCallback(async () => {
    setLoading(true)
    try {
      const resp = await fetch('/api/v1/ai/incidents')
      const data = await resp.json()
      if (data.code !== 0) throw new Error(data.error || '加载失败')
      setIncidents(data.data?.incidents || [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadIncidents() }, [loadIncidents])

  const handleAdvance = async (id) => {
    try {
      const resp = await fetch(`/api/v1/ai/incidents/${encodeURIComponent(id)}/advance`, { method: 'POST' })
      const data = await resp.json()
      if (data.code !== 0) throw new Error(data.error || '推进失败')
      await loadIncidents()
    } catch (err) {
      setError(err.message)
    }
  }

  const selected = incidents.find(i => i.id === selectedId)

  if (selected) {
    return (
      <section className='fx-incident-section'>
        <ErrorBox>{error}</ErrorBox>
        <IncidentDetail incident={selected} onAdvance={handleAdvance} onBack={() => setSelectedId(null)} />
      </section>
    )
  }

  return (
    <section className='fx-incident-section'>
      <h3>事件生命周期</h3>
      <ErrorBox>{error}</ErrorBox>
      {loading && <p>加载中...</p>}
      {!loading && incidents.length === 0 && <p className='fx-incident-empty'>暂无事件</p>}
      <div className='fx-incident-list'>
        {incidents.map(inc => (
          <div key={inc.id} className='fx-incident-card' onClick={() => setSelectedId(inc.id)}>
            <div className='fx-incident-card-header'>
              <span className='fx-incident-card-title'>{inc.title}</span>
              <span className='fx-incident-severity' style={{ color: SEVERITY_COLORS[inc.severity] || '#666' }}>
                {inc.severity}
              </span>
            </div>
            <div className='fx-incident-card-meta'>
              <span>阶段: {PHASES.find(p => p.key === inc.phase)?.label || inc.phase}</span>
              <span>告警数: {(inc.alert_ids || []).length}</span>
            </div>
            <PhaseProgress currentPhase={inc.phase} />
          </div>
        ))}
      </div>
    </section>
  )
}
