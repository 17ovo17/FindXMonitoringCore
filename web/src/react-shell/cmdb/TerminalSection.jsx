import React from 'react'
import { CMDB_EXECUTION_BLOCKERS } from '../api/cmdb.js'
import { Blocked } from './Shared.jsx'

/**
 * C03: SSH 终端入口。
 * 后端 WebSSH 通道、Origin 策略和审计回执未闭合前，只展示阻断契约。
 */
export function TerminalSection({ hostId, hostName: name, hostIp: ip, onClose }) {
  return (
    <div className='fx-terminal-section'>
      <div className='fx-terminal-header'>
        <span className='fx-terminal-status'>未连接</span>
        <span className='fx-terminal-info'>{name || hostId} ({ip || '-'})</span>
        {onClose && <button type='button' onClick={onClose}>关闭</button>}
      </div>
      <Blocked>{CMDB_EXECUTION_BLOCKERS.terminal}</Blocked>
      <div className='fx-terminal-container' style={{ padding: 16 }}>
        <pre className='fx-exec-stdout' style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
{`WebSSH 会话未启动。

缺少契约：
- webssh_channel_contract
- credential_ref_resolver
- websocket_origin_policy
- terminal_audit_contract

当前页面不会建立 WebSocket，也不会生成本地连接回显。`}
        </pre>
      </div>
    </div>
  )
}
