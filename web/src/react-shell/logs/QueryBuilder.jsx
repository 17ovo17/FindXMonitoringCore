import React, { useCallback, useMemo, useState } from 'react'
import { TIME_RANGES } from './LogsViewKit.jsx'
import { fieldGroups } from './logsModel.js'

const OPERATORS = [
  { value: '=', label: '=' },
  { value: '!=', label: '!=' },
  { value: 'contains', label: 'contains' },
  { value: 'not_contains', label: 'not contains' },
  { value: '>', label: '>' },
  { value: '<', label: '<' },
  { value: 'regex', label: 'regex' },
]

const LOGIC_OPTIONS = [
  { value: 'AND', label: 'AND' },
  { value: 'OR', label: 'OR' },
]

const ALL_FIELDS = fieldGroups.flatMap(g => g.fields)

/**
 * 结构化查询构建器（增强版 - 对标 SignOz）
 * 支持 AND/OR 条件组（可嵌套）
 * 每个条件：字段 + 运算符(=, !=, contains, not_contains, >, <, regex) + 值
 * 可视化查询 → 自动生成 LogQL 语法
 * 切换按钮：构建器模式 / 原始查询模式
 */
export function QueryBuilder({
  query,
  onQueryChange,
  conditions,
  onConditionsChange,
  mode,
  onModeChange,
  onSubmit,
  loading,
  timeRange,
  onTimeRangeChange,
  extraRight,
}) {
  const [showLogQL, setShowLogQL] = useState(false)

  const logQL = useMemo(() => generateLogQL(conditions), [conditions])

  const handleKey = (event) => {
    if (event.key === 'Enter') onSubmit()
  }

  return (
    <div className='fx-query-builder'>
      <div className='fx-query-builder__head'>
        <ModeToggle mode={mode} onChange={onModeChange} />
        <select className='fx-qb__select' value={timeRange} onChange={e => onTimeRangeChange(e.target.value)}>
          {TIME_RANGES.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}
        </select>
        <button type='button' className='fx-qb__btn' onClick={onSubmit} disabled={loading}>
          {loading ? '检索中...' : '检索'}
        </button>
        {mode === 'builder' && (
          <button
            type='button'
            className='fx-qb__btn fx-qb__btn--ghost'
            onClick={() => setShowLogQL(!showLogQL)}
            style={{ fontSize: 11 }}
          >
            {showLogQL ? '隐藏 LogQL' : '查看 LogQL'}
          </button>
        )}
        <span className='fx-qb__divider' />
        {extraRight}
      </div>
      {mode === 'builder' ? (
        <>
          <BuilderMode
            conditions={conditions}
            onChange={onConditionsChange}
            freeText={query}
            onFreeTextChange={onQueryChange}
            onSubmit={onSubmit}
          />
          {showLogQL && (
            <div style={{ padding: '8px 12px', background: 'var(--fx-bg-subtle, #f8fbff)', borderRadius: 6, margin: '8px 0', border: '1px solid var(--fx-border, #e3e8f1)' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
                <span style={{ fontSize: 11, color: 'var(--fx-text-weak)', fontWeight: 600 }}>生成的 LogQL</span>
                <button
                  type='button'
                  style={{ fontSize: 10, padding: '2px 8px', border: '1px solid var(--fx-border)', borderRadius: 4, background: '#fff', cursor: 'pointer' }}
                  onClick={() => { onQueryChange(logQL); onModeChange('raw') }}
                >
                  复制到原始查询
                </button>
              </div>
              <pre style={{ margin: 0, fontSize: 12, fontFamily: 'ui-monospace, Consolas, monospace', whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                {logQL || '(空查询)'}
              </pre>
            </div>
          )}
        </>
      ) : (
        <RawMode query={query} onChange={onQueryChange} onKeyDown={handleKey} />
      )}
    </div>
  )
}

function ModeToggle({ mode, onChange }) {
  return (
    <div className='fx-segment'>
      <button
        type='button'
        className={'fx-segment__btn' + (mode === 'builder' ? ' is-active' : '')}
        onClick={() => onChange('builder')}
      >
        构建器
      </button>
      <button
        type='button'
        className={'fx-segment__btn' + (mode === 'raw' ? ' is-active' : '')}
        onClick={() => onChange('raw')}
      >
        原始查询
      </button>
    </div>
  )
}

function BuilderMode({ conditions, onChange, freeText, onFreeTextChange, onSubmit }) {
  const addCondition = () => {
    onChange([...conditions, { field: 'body', operator: '=', value: '', logic: 'AND', type: 'condition' }])
  }

  const addGroup = () => {
    onChange([...conditions, {
      type: 'group',
      logic: 'AND',
      children: [{ field: 'body', operator: '=', value: '', logic: 'AND', type: 'condition' }],
    }])
  }

  const updateCondition = (idx, patch) => {
    const next = conditions.map((c, i) => i === idx ? { ...c, ...patch } : c)
    onChange(next)
  }

  const removeCondition = (idx) => {
    onChange(conditions.filter((_, i) => i !== idx))
  }

  const updateGroupChild = (groupIdx, childIdx, patch) => {
    const next = conditions.map((c, i) => {
      if (i !== groupIdx || c.type !== 'group') return c
      const children = c.children.map((ch, ci) => ci === childIdx ? { ...ch, ...patch } : ch)
      return { ...c, children }
    })
    onChange(next)
  }

  const removeGroupChild = (groupIdx, childIdx) => {
    const next = conditions.map((c, i) => {
      if (i !== groupIdx || c.type !== 'group') return c
      const children = c.children.filter((_, ci) => ci !== childIdx)
      return { ...c, children }
    })
    onChange(next)
  }

  const addGroupChild = (groupIdx) => {
    const next = conditions.map((c, i) => {
      if (i !== groupIdx || c.type !== 'group') return c
      return { ...c, children: [...c.children, { field: 'body', operator: '=', value: '', logic: 'AND', type: 'condition' }] }
    })
    onChange(next)
  }

  const handleKey = (event) => {
    if (event.key === 'Enter') onSubmit()
  }

  return (
    <div className='fx-query-builder__body'>
      <input
        className='fx-qb__search'
        value={freeText}
        onChange={e => onFreeTextChange(e.target.value)}
        onKeyDown={handleKey}
        placeholder='输入关键词检索日志（可选，与条件组合使用）'
      />
      {conditions.map((cond, idx) => {
        if (cond.type === 'group') {
          return (
            <ConditionGroup
              key={idx}
              group={cond}
              index={idx}
              showLogic={idx > 0}
              onLogicChange={logic => updateCondition(idx, { logic })}
              onUpdateChild={(childIdx, patch) => updateGroupChild(idx, childIdx, patch)}
              onRemoveChild={(childIdx) => removeGroupChild(idx, childIdx)}
              onAddChild={() => addGroupChild(idx)}
              onRemove={() => removeCondition(idx)}
            />
          )
        }
        return (
          <ConditionRow
            key={idx}
            condition={cond}
            index={idx}
            showLogic={idx > 0}
            onChange={patch => updateCondition(idx, patch)}
            onRemove={() => removeCondition(idx)}
          />
        )
      })}
      <div style={{ display: 'flex', gap: 8, alignSelf: 'flex-start' }}>
        <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={addCondition}>
          + 添加条件
        </button>
        <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={addGroup}>
          + 添加条件组
        </button>
      </div>
    </div>
  )
}

function ConditionGroup({ group, index, showLogic, onLogicChange, onUpdateChild, onRemoveChild, onAddChild, onRemove }) {
  return (
    <div style={{ border: '1px solid var(--fx-border, #e3e8f1)', borderRadius: 8, padding: '8px 10px', margin: '4px 0', background: 'var(--fx-bg-subtle, #f8fbff)' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
        {showLogic && (
          <select
            className='fx-qb__select fx-condition-row__logic'
            value={group.logic}
            onChange={e => onLogicChange(e.target.value)}
          >
            {LOGIC_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
          </select>
        )}
        {!showLogic && <span className='fx-condition-row__logic-placeholder'>WHERE</span>}
        <span style={{ fontSize: 11, color: 'var(--fx-text-weak)', fontWeight: 600 }}>条件组</span>
        <span style={{ flex: 1 }} />
        <button type='button' className='fx-qb__btn fx-qb__btn--danger' onClick={onRemove} style={{ minWidth: 32, padding: 0, fontSize: 11 }}>
          删除组
        </button>
      </div>
      {(group.children || []).map((child, childIdx) => (
        <ConditionRow
          key={childIdx}
          condition={child}
          index={childIdx}
          showLogic={childIdx > 0}
          onChange={patch => onUpdateChild(childIdx, patch)}
          onRemove={() => onRemoveChild(childIdx)}
        />
      ))}
      <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={onAddChild} style={{ fontSize: 11, marginTop: 4 }}>
        + 组内添加条件
      </button>
    </div>
  )
}

function ConditionRow({ condition, index, showLogic, onChange, onRemove }) {
  return (
    <div className='fx-condition-row'>
      {showLogic && (
        <select
          className='fx-qb__select fx-condition-row__logic'
          value={condition.logic}
          onChange={e => onChange({ logic: e.target.value })}
        >
          {LOGIC_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>
      )}
      {!showLogic && <span className='fx-condition-row__logic-placeholder'>WHERE</span>}
      <select
        className='fx-qb__select'
        value={condition.field}
        onChange={e => onChange({ field: e.target.value })}
      >
        {ALL_FIELDS.map(f => <option key={f} value={f}>{f}</option>)}
      </select>
      <select
        className='fx-qb__select'
        value={condition.operator}
        onChange={e => onChange({ operator: e.target.value })}
      >
        {OPERATORS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
      <input
        className='fx-qb__search'
        style={{ flex: '1 1 200px', minWidth: 120 }}
        value={condition.value}
        onChange={e => onChange({ value: e.target.value })}
        placeholder='值'
      />
      <button type='button' className='fx-qb__btn fx-qb__btn--danger' onClick={onRemove} style={{ minWidth: 32, padding: 0 }}>
        ×
      </button>
    </div>
  )
}

function RawMode({ query, onChange, onKeyDown }) {
  const [MonacoEditor, setMonacoEditor] = useState(null)

  React.useEffect(() => {
    import('@monaco-editor/react').then(mod => setMonacoEditor(() => mod.default))
  }, [])

  if (!MonacoEditor) {
    return (
      <div className='fx-query-builder__body'>
        <textarea
          className='fx-qb__search'
          style={{ minHeight: 120, fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace', resize: 'vertical' }}
          value={query}
          onChange={e => onChange(e.target.value)}
          onKeyDown={onKeyDown}
          placeholder='输入原始查询语句...'
        />
      </div>
    )
  }

  return (
    <div className='fx-query-builder__body'>
      <div style={{ border: '1px solid var(--fx-log-border-strong, #d8e1ee)', borderRadius: 8, overflow: 'hidden' }}>
        <MonacoEditor
          height='140px'
          language='plaintext'
          theme='vs-dark'
          value={query}
          onChange={value => onChange(value || '')}
          options={{
            minimap: { enabled: false },
            lineNumbers: 'off',
            scrollBeyondLastLine: false,
            fontSize: 13,
            wordWrap: 'on',
            padding: { top: 8 },
          }}
        />
      </div>
    </div>
  )
}

function generateLogQL(conditions) {
  if (!conditions || !conditions.length) return ''

  const condToLogQL = (cond) => {
    if (!cond.field || !cond.value) return ''
    const field = cond.field
    const value = cond.value
    switch (cond.operator) {
      case '=': return `${field}="${value}"`
      case '!=': return `${field}!="${value}"`
      case 'contains': return `${field}=~".*${escapeRegex(value)}.*"`
      case 'not_contains': return `${field}!~".*${escapeRegex(value)}.*"`
      case '>': return `${field}>${value}`
      case '<': return `${field}<${value}`
      case 'regex': return `${field}=~"${value}"`
      default: return `${field}="${value}"`
    }
  }

  const groupToLogQL = (items, defaultLogic = 'AND') => {
    const parts = []
    for (const item of items) {
      if (item.type === 'group') {
        const childQL = groupToLogQL(item.children || [], 'AND')
        if (childQL) parts.push({ logic: item.logic || defaultLogic, expr: `(${childQL})` })
      } else {
        const expr = condToLogQL(item)
        if (expr) parts.push({ logic: item.logic || defaultLogic, expr })
      }
    }
    if (!parts.length) return ''
    return parts.map((p, i) => i === 0 ? p.expr : `${p.logic.toLowerCase() === 'or' ? ' or ' : ' and '}${p.expr}`).join('')
  }

  const result = groupToLogQL(conditions)
  return result ? `{${result}}` : ''
}

function escapeRegex(str) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}
