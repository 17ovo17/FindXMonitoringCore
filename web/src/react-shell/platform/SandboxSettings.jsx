import React, { useCallback, useEffect, useState } from 'react'
import { get, put } from '../api/http.js'

const MODES = [
  { value: 'readonly', label: '只读模式', icon: '\u{1F6E1}', desc: 'AI 只能查询，不能执行任何写操作。' },
  { value: 'auto_review', label: '自动审查', icon: '\u{1F50D}', desc: 'AI 可执行低风险操作，自动记录审计，高风险被拒绝。' },
  { value: 'full_access', label: '完全访问', icon: '⚠️', desc: 'AI 可执行所有操作，高风险需要你确认。' },
]

function ModeSelector({ current, onChange, saving }) {
  return (
    <div className='fx-sandbox-modes'>
      {MODES.map(mode => (
        <button
          key={mode.value}
          type='button'
          className={`fx-sandbox-mode-card${current === mode.value ? ' is-active' : ''}`}
          onClick={() => onChange(mode.value)}
          disabled={saving}
        >
          <span className='fx-sandbox-mode-icon'>{mode.icon}</span>
          <strong>{mode.label}</strong>
          <span className='fx-sandbox-mode-desc'>{mode.desc}</span>
        </button>
      ))}
    </div>
  )
}

function AuditTable({ entries }) {
  if (!entries.length) {
    return <div className='fx-sandbox-empty'>暂无审计记录。</div>
  }
  return (
    <div className='fx-sandbox-audit-table-wrap'>
      <table className='fx-sandbox-audit-table'>
        <thead>
          <tr>
            <th>时间</th>
            <th>工具</th>
            <th>风险</th>
            <th>状态</th>
            <th>耗时</th>
            <th>错误</th>
          </tr>
        </thead>
        <tbody>
          {entries.map((entry, idx) => (
            <tr key={entry.id || idx}>
              <td>{entry.created_at ? new Date(entry.created_at).toLocaleString() : '-'}</td>
              <td>{entry.tool_name}</td>
              <td><span className={`fx-sandbox-risk fx-sandbox-risk-${entry.risk_level}`}>L{entry.risk_level}</span></td>
              <td><span className={`fx-sandbox-status fx-sandbox-status-${entry.status}`}>{entry.status}</span></td>
              <td>{entry.duration_ms ? `${entry.duration_ms}ms` : '-'}</td>
              <td className='fx-sandbox-audit-error'>{entry.error || '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export function SandboxSettings() {
  const [policy, setPolicy] = useState(null)
  const [audit, setAudit] = useState([])
  const [error, setError] = useState('')
  const [feedback, setFeedback] = useState('')
  const [saving, setSaving] = useState(false)
  const [deniedInput, setDeniedInput] = useState('')

  const loadPolicy = useCallback(async () => {
    try {
      const resp = await get('/sandbox/policy')
      const data = resp?.data || resp
      setPolicy(data)
      setDeniedInput((data?.denied_commands || []).join('\n'))
      setError('')
    } catch (err) {
      setError(err?.message || '加载策略失败')
    }
  }, [])

  const loadAudit = useCallback(async () => {
    try {
      const resp = await get('/sandbox/audit', { params: { limit: 50 } })
      const data = resp?.data || resp
      setAudit(data?.entries || [])
    } catch {
      // 静默
    }
  }, [])

  useEffect(() => {
    loadPolicy()
    loadAudit()
  }, [loadPolicy, loadAudit])

  const handleModeChange = async (mode) => {
    setSaving(true)
    setError('')
    setFeedback('')
    try {
      const payload = {
        ...policy,
        mode,
        denied_commands: deniedInput.split('\n').map(s => s.trim()).filter(Boolean),
      }
      await put('/sandbox/policy', payload)
      setFeedback('策略已更新。')
      await loadPolicy()
    } catch (err) {
      setError(err?.message || '更新策略失败')
    } finally {
      setSaving(false)
    }
  }

  const handleSaveBlacklist = async () => {
    if (!policy) return
    setSaving(true)
    setError('')
    setFeedback('')
    try {
      const payload = {
        ...policy,
        denied_commands: deniedInput.split('\n').map(s => s.trim()).filter(Boolean),
      }
      await put('/sandbox/policy', payload)
      setFeedback('黑名单已更新。')
      await loadPolicy()
    } catch (err) {
      setError(err?.message || '更新黑名单失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <section className='fx-sandbox-settings'>
      <h2>AI 执行沙箱</h2>
      <p className='fx-sandbox-desc'>
        控制 AI 在执行工具时的权限策略。根据风险等级采用不同的权限模型，保障操作安全。
      </p>

      {error && <div className='fx-platform-error'>{error}</div>}
      {feedback && <div className='fx-platform-feedback'>{feedback}</div>}

      {policy && (
        <>
          <h3>权限模式</h3>
          <ModeSelector current={policy.mode} onChange={handleModeChange} saving={saving} />

          <h3>命令黑名单</h3>
          <p className='fx-sandbox-note'>以下命令将被无条件拒绝执行（每行一条）：</p>
          <textarea
            className='fx-sandbox-blacklist'
            value={deniedInput}
            onChange={e => setDeniedInput(e.target.value)}
            rows={6}
            placeholder='rm -rf /&#10;dd if=&#10;mkfs'
          />
          <button type='button' className='fx-sandbox-save-btn' onClick={handleSaveBlacklist} disabled={saving}>
            {saving ? '保存中...' : '保存黑名单'}
          </button>

          <div className='fx-sandbox-meta'>
            <span>最大超时: {policy.max_timeout}s</span>
            <span>审计全部操作: {policy.audit_all ? '是' : '否'}</span>
          </div>
        </>
      )}

      <h3>审计日志</h3>
      <div className='fx-sandbox-audit-toolbar'>
        <button type='button' onClick={loadAudit}>刷新</button>
      </div>
      <AuditTable entries={audit} />
    </section>
  )
}
