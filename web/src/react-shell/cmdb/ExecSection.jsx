import React, { useState } from 'react'
import { CMDB_EXECUTION_BLOCKERS, cmdbApi, cmdbContractMessage, isCmdbContractBlocked } from '../api/cmdb.js'
import { Blocked, ErrorBox, Field } from './Shared.jsx'

/**
 * C05: 命令执行组件
 * 命令输入 + 超时设置 + sudo 开关 + 输出显示。
 */
export function ExecSection({ hostId, hostName: name, hostIp: ip, onClose }) {
  const [command, setCommand] = useState('')
  const [timeout, setTimeout_] = useState(30)
  const [sudo, setSudo] = useState(false)
  const [executing, setExecuting] = useState(false)
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [result, setResult] = useState(null)
  const [history, setHistory] = useState([])

  const handleExec = async () => {
    if (!command.trim()) {
      setError('请输入要执行的命令')
      return
    }

    setExecuting(true)
    setError('')
    setBlocked('')
    setResult(null)

    try {
      const res = await cmdbApi.exec(hostId, {
        command: command.trim(),
        timeout: Number(timeout) || 30,
        sudo,
      })
      setResult(res)
      setHistory(prev => [{ command: command.trim(), time: new Date().toLocaleTimeString(), exitCode: res.exit_code ?? '-' }, ...prev.slice(0, 19)])
      if (res?.exit_code !== 0) {
        setError(`命令退出码: ${res?.exit_code ?? '未知'}`)
      }
    } catch (err) {
      if (isCmdbContractBlocked(err)) {
        setBlocked(cmdbContractMessage(err, CMDB_EXECUTION_BLOCKERS.exec))
      } else {
        setError(err?.message || '执行失败')
      }
    } finally {
      setExecuting(false)
    }
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleExec()
    }
  }

  return (
    <div className='fx-assets-form'>
      <h3 style={{ margin: '0 0 12px', fontSize: 16, color: '#193a63' }}>
        命令执行 - {name} ({ip})
      </h3>

      <Blocked>{CMDB_EXECUTION_BLOCKERS.exec}</Blocked>

      <Field label='命令'>
        <textarea rows='3' value={command} onChange={e => setCommand(e.target.value)} onKeyDown={handleKeyDown} placeholder='输入要执行的命令...' disabled={executing} style={{ fontFamily: 'monospace' }} />
      </Field>

      <div className='fx-assets-form-grid'>
        <Field label='超时（秒）'>
          <input type='number' min='1' max='300' value={timeout} onChange={e => setTimeout_(e.target.value)} disabled={executing} />
        </Field>
        <Field label='Sudo'>
          <label style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer' }}>
            <input type='checkbox' checked={sudo} onChange={e => setSudo(e.target.checked)} disabled={executing} />
            <span>以 root 权限执行</span>
          </label>
        </Field>
      </div>

      <ErrorBox>{error}</ErrorBox>
      {blocked && <Blocked>{blocked}</Blocked>}

      {result && (
        <div className='fx-exec-output'>
          <h4 style={{ margin: '0 0 8px', fontSize: 13, color: '#526984' }}>执行回执</h4>
          {result.stdout && <pre className='fx-exec-stdout'>{result.stdout}</pre>}
          {result.stderr && <pre className='fx-exec-stderr'>{result.stderr}</pre>}
          <div className='fx-assets-muted' style={{ marginTop: 4 }}>
            退出码: {result.exit_code ?? '-'} | 耗时: {result.duration_ms ?? '-'}ms | {result.executed_at || '-'}
          </div>
        </div>
      )}

      {history.length > 0 && (
        <div style={{ marginTop: 12 }}>
          <h4 style={{ margin: '0 0 8px', fontSize: 13, color: '#526984' }}>历史命令</h4>
          <div className='fx-assets-table'><table><thead><tr><th>命令</th><th>退出码</th><th>时间</th></tr></thead>
            <tbody>{history.map((h, i) => <tr key={i}><td style={{ fontFamily: 'monospace', fontSize: 12 }}>{h.command}</td><td>{h.exitCode}</td><td>{h.time}</td></tr>)}</tbody>
          </table></div>
        </div>
      )}

      <footer>
        <button type='button' disabled={executing || !command.trim()} onClick={handleExec}>
          {executing ? '执行中...' : '提交执行'}
        </button>
        {onClose && <button type='button' onClick={onClose} style={{ marginLeft: 8 }}>关闭</button>}
      </footer>
    </div>
  )
}
