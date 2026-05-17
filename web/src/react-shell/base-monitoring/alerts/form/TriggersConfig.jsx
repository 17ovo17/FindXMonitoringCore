import React from 'react'
import { severities, noDataPolicies } from '../alertModel.js'
import { AlertPromQLEditor } from './PromQLEditor.jsx'

/**
 * Triggers 配置组件
 * 对齐夜莺 Triggers 组件：多条件触发、severity、比较操作符、阈值、for duration、nodata 策略、恢复条件
 * 增强：多查询条件（A、B、C...）+ JOIN 操作
 */

const operators = [
  { value: '>', label: '>' },
  { value: '>=', label: '>=' },
  { value: '<', label: '<' },
  { value: '<=', label: '<=' },
  { value: '==', label: '==' },
  { value: '!=', label: '!=' },
]

const JOIN_TYPES = [
  { value: 'none', label: '无 JOIN' },
  { value: 'left', label: 'LEFT JOIN' },
  { value: 'right', label: 'RIGHT JOIN' },
  { value: 'inner', label: 'INNER JOIN' },
]

const emptyQuery = (letter) => ({
  ref: letter,
  expr: '',
  legendFormat: '',
  join: 'none',
  join_on: '',
})

function QueryItem({ query, index, onChange, onRemove, canRemove, datasourceId }) {
  const letter = String.fromCharCode(65 + index)
  const update = (field, val) => onChange(index, { ...query, [field]: val, ref: letter })

  return (
    <div className="fx-alert-query-item">
      <div className="fx-alert-query-item__head">
        <span className="fx-alert-query-letter">{letter}</span>
        {index > 0 && (
          <select
            className="fx-alert-query-join"
            value={query.join || 'none'}
            onChange={(e) => update('join', e.target.value)}
          >
            {JOIN_TYPES.map((j) => <option key={j.value} value={j.value}>{j.label}</option>)}
          </select>
        )}
        {index > 0 && query.join && query.join !== 'none' && (
          <input
            className="fx-alert-query-join-on"
            value={query.join_on || ''}
            onChange={(e) => update('join_on', e.target.value)}
            placeholder="ON 标签（逗号分隔）"
            style={{ width: 180 }}
          />
        )}
        <span style={{ flex: 1 }} />
        {canRemove && (
          <button type="button" className="fx-alert-link" style={{ color: '#dc2626' }} onClick={() => onRemove(index)}>
            删除
          </button>
        )}
      </div>
      <AlertPromQLEditor
        value={query.expr || ''}
        onChange={(val) => update('expr', val)}
        datasourceId={datasourceId}
        height={80}
        showPreview={false}
        placeholder={`查询 ${letter} 的 PromQL 表达式`}
      />
      <input
        className="fx-alert-query-legend"
        value={query.legendFormat || ''}
        onChange={(e) => update('legendFormat', e.target.value)}
        placeholder="Legend 格式（如 {{instance}}）"
        style={{ marginTop: 4, width: '100%', padding: '4px 8px', fontSize: 12, border: '1px solid var(--fx-border, #e5e7eb)', borderRadius: 4 }}
      />
    </div>
  )
}

function QueriesConfig({ queries, onChange, datasourceId }) {
  const addQuery = () => {
    const letter = String.fromCharCode(65 + queries.length)
    onChange([...queries, emptyQuery(letter)])
  }
  const updateQuery = (index, query) => {
    const next = [...queries]
    next[index] = query
    onChange(next)
  }
  const removeQuery = (index) => onChange(queries.filter((_, i) => i !== index))

  return (
    <div className="fx-alert-queries-section">
      <div className="fx-alert-effective-ranges-head">
        <span className="fx-alert-effective-label">查询条件</span>
        <button type="button" onClick={addQuery}>添加查询</button>
      </div>
      {queries.map((query, index) => (
        <QueryItem
          key={index}
          query={query}
          index={index}
          onChange={updateQuery}
          onRemove={removeQuery}
          canRemove={queries.length > 1}
          datasourceId={datasourceId}
        />
      ))}
    </div>
  )
}

function TriggerItem({ trigger, index, onChange, onRemove, canRemove }) {
  const update = (field, val) => onChange(index, { ...trigger, [field]: val })
  return (
    <div className='fx-alert-trigger-item'>
      <div className='fx-alert-trigger-row'>
        <label>
          <span>级别</span>
          <select value={trigger.severity || 'warning'} onChange={(e) => update('severity', e.target.value)}>
            {severities.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
          </select>
        </label>
        <label>
          <span>操作符</span>
          <select value={trigger.operator || '>'} onChange={(e) => update('operator', e.target.value)}>
            {operators.map((op) => <option key={op.value} value={op.value}>{op.label}</option>)}
          </select>
        </label>
        <label>
          <span>阈值</span>
          <input
            type='number'
            value={trigger.value ?? 0}
            onChange={(e) => update('value', Number(e.target.value))}
          />
        </label>
        <label>
          <span>持续时间</span>
          <input
            value={trigger.for_duration || ''}
            onChange={(e) => update('for_duration', e.target.value)}
            placeholder='如 5m, 10m'
          />
        </label>
        {canRemove && <button type='button' onClick={() => onRemove(index)}>删除</button>}
      </div>
    </div>
  )
}

export function TriggersConfig({ value, onChange, datasourceId }) {
  const config = value || {
    queries: [emptyQuery('A')],
    triggers: [{ severity: 'warning', operator: '>', value: 0, for_duration: '5m' }],
    nodata_trigger: { enable: false, severity: 'warning', action: 'keep_state' },
    recover_config: { enable: true, recover_duration: 0 },
  }

  const update = (patch) => onChange?.({ ...config, ...patch })

  const queries = config.queries || [emptyQuery('A')]

  const updateTrigger = (index, trigger) => {
    const triggers = [...(config.triggers || [])]
    triggers[index] = trigger
    update({ triggers })
  }

  const removeTrigger = (index) => {
    update({ triggers: (config.triggers || []).filter((_, i) => i !== index) })
  }

  const addTrigger = () => {
    update({
      triggers: [...(config.triggers || []), { severity: 'warning', operator: '>', value: 0, for_duration: '5m' }],
    })
  }

  const updateNodata = (patch) => {
    update({ nodata_trigger: { ...(config.nodata_trigger || {}), ...patch } })
  }

  const updateRecover = (patch) => {
    update({ recover_config: { ...(config.recover_config || {}), ...patch } })
  }

  return (
    <div className='fx-alert-triggers'>
      <QueriesConfig
        queries={queries}
        onChange={(q) => update({ queries: q })}
        datasourceId={datasourceId}
      />
      <div className='fx-alert-triggers-section'>
        <div className='fx-alert-effective-ranges-head'>
          <span className='fx-alert-effective-label'>触发条件</span>
          <button type='button' onClick={addTrigger}>添加条件</button>
        </div>
        {(config.triggers || []).map((trigger, index) => (
          <TriggerItem
            key={index}
            trigger={trigger}
            index={index}
            onChange={updateTrigger}
            onRemove={removeTrigger}
            canRemove={(config.triggers || []).length > 1}
          />
        ))}
      </div>
      <div className='fx-alert-triggers-section'>
        <span className='fx-alert-effective-label'>Nodata 策略</span>
        <div className='fx-alert-trigger-row'>
          <label className='fx-alert-check-inline'>
            <input
              type='checkbox'
              checked={config.nodata_trigger?.enable || false}
              onChange={(e) => updateNodata({ enable: e.target.checked })}
            />
            启用 Nodata 触发
          </label>
          {config.nodata_trigger?.enable && (
            <>
              <label>
                <span>级别</span>
                <select
                  value={config.nodata_trigger?.severity || 'warning'}
                  onChange={(e) => updateNodata({ severity: e.target.value })}
                >
                  {severities.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
                </select>
              </label>
              <label>
                <span>动作</span>
                <select
                  value={config.nodata_trigger?.action || 'keep_state'}
                  onChange={(e) => updateNodata({ action: e.target.value })}
                >
                  {noDataPolicies.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
                </select>
              </label>
            </>
          )}
        </div>
      </div>
      <div className='fx-alert-triggers-section'>
        <span className='fx-alert-effective-label'>恢复条件</span>
        <div className='fx-alert-trigger-row'>
          <label className='fx-alert-check-inline'>
            <input
              type='checkbox'
              checked={config.recover_config?.enable !== false}
              onChange={(e) => updateRecover({ enable: e.target.checked })}
            />
            启用自动恢复
          </label>
          {config.recover_config?.enable !== false && (
            <label>
              <span>恢复等待时间（秒）</span>
              <input
                type='number'
                min={0}
                value={config.recover_config?.recover_duration ?? 0}
                onChange={(e) => updateRecover({ recover_duration: Number(e.target.value) || 0 })}
              />
            </label>
          )}
        </div>
      </div>
    </div>
  )
}

