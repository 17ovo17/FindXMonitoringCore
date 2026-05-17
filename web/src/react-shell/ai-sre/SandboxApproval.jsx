import React, { useCallback, useEffect, useRef, useState } from 'react'
import { get, post } from '../api/http.js'
import '../platform/sandbox.css'

/**
 * AI 执行确认弹窗组件
 * 当 AI 要执行 Level 2 操作时，通过轮询获取待确认请求
 * 前端弹出确认卡片，30s 超时自动拒绝
 */

const POLL_INTERVAL = 2000 // 2s 轮询间隔

function ApprovalCard({ approval, onResolve }) {
  const [remaining, setRemaining] = useState(30)
  const timerRef = useRef(null)

  useEffect(() => {
    const expiresAt = new Date(approval.expires_at).getTime()
    const tick = () => {
      const now = Date.now()
      const left = Math.max(0, Math.ceil((expiresAt - now) / 1000))
      setRemaining(left)
      if (left <= 0) {
        onResolve(approval.id, 'deny')
      }
    }
    tick()
    timerRef.current = setInterval(tick, 1000)
    return () => clearInterval(timerRef.current)
  }, [approval.expires_at, approval.id, onResolve])

  const progressPct = Math.max(0, (remaining / 30) * 100)

  return (
    <div className='fx-sandbox-approval-card'>
      <div className='fx-sandbox-approval-header'>
        <span className='fx-sandbox-approval-icon'>&#9888;</span>
        <strong>AI 请求执行危险操作</strong>
      </div>
      <div className='fx-sandbox-approval-body'>
        <div className='fx-sandbox-approval-field'>
          <span className='fx-sandbox-approval-label'>操作</span>
          <span className='fx-sandbox-approval-value'>{approval.tool_name}</span>
        </div>
        <div className='fx-sandbox-approval-field'>
          <span className='fx-sandbox-approval-label'>描述</span>
          <span className='fx-sandbox-approval-value'>{approval.description}</span>
        </div>
        <div className='fx-sandbox-approval-field'>
          <span className='fx-sandbox-approval-label'>影响</span>
          <span className='fx-sandbox-approval-value'>{approval.impact}</span>
        </div>
        {approval.params && (
          <details className='fx-sandbox-approval-params'>
            <summary>参数详情</summary>
            <pre>{JSON.stringify(approval.params, null, 2)}</pre>
          </details>
        )}
      </div>
      <div className='fx-sandbox-approval-timer'>
        <div className='fx-sandbox-approval-progress' style={{ width: `${progressPct}%` }} />
        <span>{remaining}s 后自动拒绝</span>
      </div>
      <div className='fx-sandbox-approval-actions'>
        <button type='button' className='fx-sandbox-btn-deny' onClick={() => onResolve(approval.id, 'deny')}>
          拒绝
        </button>
        <button type='button' className='fx-sandbox-btn-approve' onClick={() => onResolve(approval.id, 'approve')}>
          确认执行
        </button>
      </div>
    </div>
  )
}

export function SandboxApprovalOverlay() {
  const [approvals, setApprovals] = useState([])
  const pollRef = useRef(null)

  const fetchApprovals = useCallback(async () => {
    try {
      const resp = await get('/sandbox/approvals')
      const data = resp?.data || resp
      setApprovals(data?.approvals || [])
    } catch {
      // 静默失败
    }
  }, [])

  useEffect(() => {
    fetchApprovals()
    pollRef.current = setInterval(fetchApprovals, POLL_INTERVAL)
    return () => clearInterval(pollRef.current)
  }, [fetchApprovals])

  const handleResolve = useCallback(async (id, action) => {
    try {
      await post(`/sandbox/approvals/${encodeURIComponent(id)}/resolve`, { action, user: 'web_user' })
      setApprovals(prev => prev.filter(a => a.id !== id))
    } catch {
      // 静默失败
    }
  }, [])

  if (!approvals.length) return null

  return (
    <div className='fx-sandbox-approval-overlay'>
      {approvals.map(approval => (
        <ApprovalCard key={approval.id} approval={approval} onResolve={handleResolve} />
      ))}
    </div>
  )
}
