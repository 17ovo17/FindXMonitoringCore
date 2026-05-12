import React, { useCallback, useMemo, useState } from 'react'
import { TIME_RANGES } from './LogsViewKit.jsx'
import { fieldGroups } from './logsModel.js'

const OPERATORS = [
  { value: '=', label: '=' },
  { value: '!=', label: '!=' },
  { value: 'contains', label: 'contains' },
  { value: 'regex', label: 'regex' },
]

const LOGIC_OPTIONS = [
  { value: 'AND', label: 'AND' },
  { value: 'OR', label: 'OR' },
]

const ALL_FIELDS = fieldGroups.flatMap(g => g.fields)

/**
 * DEGRADE-048: 结构化查询构建器
 * 支持构建器模式（多条件行）和原始查询模式（Monaco Editor）
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
        <span className='fx-qb__divider' />
        {extraRight}
      </div>
      {mode === 'builder' ? (
        <BuilderMode
          conditions={conditions}
          onChange={onConditionsChange}
          freeText={query}
          onFreeTextChange={onQueryChange}
          onSubmit={onSubmit}
        />
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
    onChange([...conditions, { field: 'body', operator: '=', value: '', logic: 'AND' }])
  }

  const updateCondition = (idx, patch) => {
    const next = conditions.map((c, i) => i === idx ? { ...c, ...patch } : c)
    onChange(next)
  }

  const removeCondition = (idx) => {
    onChange(conditions.filter((_, i) => i !== idx))
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
      {conditions.map((cond, idx) => (
        <ConditionRow
          key={idx}
          condition={cond}
          index={idx}
          showLogic={idx > 0}
          onChange={patch => updateCondition(idx, patch)}
          onRemove={() => removeCondition(idx)}
        />
      ))}
      <button type='button' className='fx-qb__btn fx-qb__btn--ghost' onClick={addCondition} style={{ alignSelf: 'flex-start' }}>
        + 添加条件
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
