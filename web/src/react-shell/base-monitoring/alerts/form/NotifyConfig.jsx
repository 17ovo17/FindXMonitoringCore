import React, { useEffect, useState } from 'react'
import { alertsApi } from '../../../api/alerts.js'
import { makeError } from '../alertModel.js'

/**
 * 通知配置组件
 * 对齐夜莺 Notify 组件：通知渠道、接收组、回调地址、通知模板、重复通知间隔
 */
export function NotifyConfig({ value, onChange }) {
  const config = value || {
    notify_channels: [],
    notify_groups: [],
    callbacks: [],
    notify_template_id: '',
    notify_repeat_step: 60,
    notify_max_number: 0,
    notify_recovered: true,
  }

  const [channels, setChannels] = useState([])
  const [templates, setTemplates] = useState([])
  const [loadError, setLoadError] = useState('')

  useEffect(() => {
    alertsApi.listNotificationChannels()
      .then((list) => setChannels(Array.isArray(list) ? list : []))
      .catch((err) => setLoadError(makeError(err, '加载通知渠道失败')))
    alertsApi.listNotificationTemplates()
      .then((list) => setTemplates(Array.isArray(list) ? list : []))
      .catch(() => {})
  }, [])

  const update = (patch) => onChange?.({ ...config, ...patch })

  const toggleChannel = (channelId) => {
    const current = config.notify_channels || []
    const next = current.includes(channelId)
      ? current.filter((id) => id !== channelId)
      : [...current, channelId]
    update({ notify_channels: next })
  }

  const updateCallback = (index, val) => {
    const cbs = [...(config.callbacks || [])]
    cbs[index] = val
    update({ callbacks: cbs })
  }

  const removeCallback = (index) => {
    update({ callbacks: (config.callbacks || []).filter((_, i) => i !== index) })
  }

  const addCallback = () => {
    update({ callbacks: [...(config.callbacks || []), ''] })
  }

  return (
    <div className='fx-alert-notify'>
      {loadError && <div className='fx-alert-message is-error'>{loadError}</div>}
      <div className='fx-alert-notify-section'>
        <span className='fx-alert-effective-label'>通知渠道</span>
        <div className='fx-alert-effective-daylist'>
          {channels.map((ch) => (
            <label key={ch.id || ch.key || ch.name} className='fx-alert-check-inline'>
              <input
                type='checkbox'
                checked={(config.notify_channels || []).includes(String(ch.id || ch.key))}
                onChange={() => toggleChannel(String(ch.id || ch.key))}
              />
              {ch.name || ch.label || ch.key}
            </label>
          ))}
          {channels.length === 0 && !loadError && <span className='fx-alert-hint-text'>暂无可用渠道</span>}
        </div>
      </div>
      <div className='fx-alert-notify-section'>
        <span className='fx-alert-effective-label'>接收组 / 接收人</span>
        <input
          value={(config.notify_groups || []).join(',')}
          onChange={(e) => update({ notify_groups: e.target.value ? e.target.value.split(',') : [] })}
          placeholder='多个接收组用逗号分隔'
        />
      </div>
      <div className='fx-alert-notify-section'>
        <span className='fx-alert-effective-label'>通知模板</span>
        <select
          value={config.notify_template_id || ''}
          onChange={(e) => update({ notify_template_id: e.target.value })}
        >
          <option value=''>默认模板</option>
          {templates.map((tpl) => (
            <option key={tpl.id || tpl.name} value={String(tpl.id || tpl.name)}>
              {tpl.name || tpl.title}
            </option>
          ))}
        </select>
      </div>
      <div className='fx-alert-form'>
        <label>
          <span>重复通知间隔（分钟）</span>
          <input
            type='number'
            min={0}
            value={config.notify_repeat_step ?? 60}
            onChange={(e) => update({ notify_repeat_step: Number(e.target.value) || 0 })}
          />
        </label>
        <label>
          <span>最大通知次数（0=不限）</span>
          <input
            type='number'
            min={0}
            value={config.notify_max_number ?? 0}
            onChange={(e) => update({ notify_max_number: Number(e.target.value) || 0 })}
          />
        </label>
      </div>
      <div className='fx-alert-notify-section'>
        <label className='fx-alert-check-inline'>
          <input
            type='checkbox'
            checked={config.notify_recovered !== false}
            onChange={(e) => update({ notify_recovered: e.target.checked })}
          />
          恢复时通知
        </label>
      </div>
      <div className='fx-alert-notify-section'>
        <div className='fx-alert-effective-ranges-head'>
          <span className='fx-alert-effective-label'>回调 URL</span>
          <button type='button' onClick={addCallback}>添加</button>
        </div>
        {(config.callbacks || []).map((cb, index) => (
          <div key={index} className='fx-alert-effective-range'>
            <input
              value={cb}
              onChange={(e) => updateCallback(index, e.target.value)}
              placeholder='https://example.com/webhook'
              style={{ flex: 1 }}
            />
            <button type='button' onClick={() => removeCallback(index)}>删除</button>
          </div>
        ))}
      </div>
    </div>
  )
}
