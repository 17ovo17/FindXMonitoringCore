import React from 'react'
import { severities, noDataPolicies } from '../alertModel.js'

/**
 * Triggers 配置组件
 * 对齐夜莺 Triggers 组件：多条件触发、severity、比较操作符、阈值、for duration、nodata 策略、恢复条件
 */

const operators = [
  { value: '>', label: '>' },
  { value: '>=', label: '>=' },
  { value: '<', label: '<' },
  { value: '<=', label: '<=' },
  { value: '==', label: '==' },
  { value: '!=', label: '!=' },
]

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

export function TriggersConfig({ value, onChange }) {
  const config = value || {
    triggers: [{ severity: 'warning', operator: '>', value: 0, for_duration: '5m' }],
    nodata_trigger: { enable: false, severity: 'warning', action: 'keep_state' },
    recover_config: { enable: true, recover_duration: 0 },
  }

  const update = (patch) => onChange?.({ ...config, ...patch })

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

