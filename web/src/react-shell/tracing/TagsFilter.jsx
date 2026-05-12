import React, { useEffect, useState } from 'react'
import { formatTracingError, tracingApi } from '../api/tracing.js'

/**
 * 结构化 Tags 过滤器
 * 每行一个 tag: key (下拉/自由输入) + 操作符 (=) + value 输入
 * 支持添加/删除 tag 行
 */
export function TagsFilter({ value, onChange }) {
  const [tagKeys, setTagKeys] = useState([])
  const [tagKeysBlocked, setTagKeysBlocked] = useState(false)
  const [rows, setRows] = useState(() => parseTagsToRows(value))

  // 尝试从后端获取 tag keys
  useEffect(() => {
    tracingApi.selectors.tagKeys({})
      .then(keys => {
        if (Array.isArray(keys) && keys.length > 0) {
          setTagKeys(keys.map(k => k.label || k.value || k.key || k))
        } else {
          setTagKeysBlocked(true)
        }
      })
      .catch(() => { setTagKeysBlocked(true) })
  }, [])

  // 同步外部 value 变化
  useEffect(() => {
    const parsed = parseTagsToRows(value)
    if (JSON.stringify(parsed) !== JSON.stringify(rows)) {
      setRows(parsed)
    }
  }, [value])

  const emitChange = (nextRows) => {
    setRows(nextRows)
    const serialized = nextRows
      .filter(r => r.key.trim())
      .map(r => `${r.key}=${r.value}`)
      .join('\n')
    onChange(serialized)
  }

  const updateRow = (idx, field, val) => {
    const next = rows.map((r, i) => i === idx ? { ...r, [field]: val } : r)
    emitChange(next)
  }

  const addRow = () => {
    emitChange([...rows, { key: '', op: '=', value: '' }])
  }

  const removeRow = (idx) => {
    const next = rows.filter((_, i) => i !== idx)
    emitChange(next.length ? next : [{ key: '', op: '=', value: '' }])
  }

  return (
    <div className='fx-tracing-tags-filter'>
      <label style={{ fontSize: 12, fontWeight: 700, color: 'var(--fx-muted, #66758d)' }}>
        Tags 过滤 {tagKeysBlocked && <span style={{ fontWeight: 400 }}>(key 自由输入)</span>}
      </label>
      {rows.map((row, idx) => (
        <div key={idx} className='fx-tracing-tag-row'>
          {tagKeysBlocked ? (
            <input
              className='fx-tag-key'
              value={row.key}
              onChange={e => updateRow(idx, 'key', e.target.value)}
              placeholder='tag key'
            />
          ) : (
            <select
              className='fx-tag-key'
              value={row.key}
              onChange={e => updateRow(idx, 'key', e.target.value)}
            >
              <option value=''>选择 key</option>
              {tagKeys.map(k => <option key={k} value={k}>{k}</option>)}
            </select>
          )}
          <span className='fx-tag-op'>=</span>
          <input
            className='fx-tag-value'
            value={row.value}
            onChange={e => updateRow(idx, 'value', e.target.value)}
            placeholder='tag value'
          />
          <button type='button' onClick={() => removeRow(idx)} title='删除此行'>✕</button>
        </div>
      ))}
      <button type='button' onClick={addRow} style={{ alignSelf: 'flex-start', fontSize: 12 }}>+ 添加 Tag</button>
    </div>
  )
}

function parseTagsToRows(tagsStr) {
  if (!tagsStr || !tagsStr.trim()) return [{ key: '', op: '=', value: '' }]
  const lines = tagsStr.split('\n').filter(l => l.trim())
  if (!lines.length) return [{ key: '', op: '=', value: '' }]
  return lines.map(line => {
    const eqIdx = line.indexOf('=')
    if (eqIdx === -1) return { key: line.trim(), op: '=', value: '' }
    return { key: line.slice(0, eqIdx).trim(), op: '=', value: line.slice(eqIdx + 1).trim() }
  })
}
