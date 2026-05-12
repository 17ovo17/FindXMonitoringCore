import React from 'react'

const daysOfWeek = [
  { value: 1, label: '周一' },
  { value: 2, label: '周二' },
  { value: 3, label: '周三' },
  { value: 4, label: '周四' },
  { value: 5, label: '周五' },
  { value: 6, label: '周六' },
  { value: 0, label: '周日' },
]

function TimeRangeRow({ range, index, onChange, onRemove }) {
  return (
    <div className='fx-alert-effective-range'>
      <input
        type='time'
        value={range.start}
        onChange={(e) => onChange(index, { ...range, start: e.target.value })}
      />
      <span>~</span>
      <input
        type='time'
        value={range.end}
        onChange={(e) => onChange(index, { ...range, end: e.target.value })}
      />
      <button type='button' onClick={() => onRemove(index)}>删除</button>
    </div>
  )
}

/**
 * 生效时间配置组件
 * 对齐夜莺 Effective 组件：支持按星期/时间段配置生效时间
 * 数据结构：{ enable_status, days_of_week: number[], time_ranges: [{start, end}] }
 */
export function EffectiveTimeConfig({ value, onChange }) {
  const config = value || { enable_status: true, days_of_week: [1, 2, 3, 4, 5, 6, 0], time_ranges: [] }

  const update = (patch) => onChange?.({ ...config, ...patch })

  const toggleDay = (day) => {
    const days = config.days_of_week || []
    const next = days.includes(day) ? days.filter((d) => d !== day) : [...days, day]
    update({ days_of_week: next })
  }

  const updateRange = (index, range) => {
    const ranges = [...(config.time_ranges || [])]
    ranges[index] = range
    update({ time_ranges: ranges })
  }

  const removeRange = (index) => {
    const ranges = (config.time_ranges || []).filter((_, i) => i !== index)
    update({ time_ranges: ranges })
  }

  const addRange = () => {
    const ranges = [...(config.time_ranges || []), { start: '09:00', end: '18:00' }]
    update({ time_ranges: ranges })
  }

  return (
    <div className='fx-alert-effective'>
      <div className='fx-alert-effective-enable'>
        <label className='fx-alert-check'>
          <input
            type='checkbox'
            checked={config.enable_status !== false}
            onChange={(e) => update({ enable_status: e.target.checked })}
          />
          启用生效时间
        </label>
      </div>
      <div className='fx-alert-effective-days'>
        <span className='fx-alert-effective-label'>生效星期</span>
        <div className='fx-alert-effective-daylist'>
          {daysOfWeek.map((day) => (
            <label key={day.value} className='fx-alert-check-inline'>
              <input
                type='checkbox'
                checked={(config.days_of_week || []).includes(day.value)}
                onChange={() => toggleDay(day.value)}
              />
              {day.label}
            </label>
          ))}
        </div>
      </div>
      <div className='fx-alert-effective-ranges'>
        <div className='fx-alert-effective-ranges-head'>
          <span className='fx-alert-effective-label'>时间段</span>
          <button type='button' onClick={addRange}>添加时间段</button>
        </div>
        {(config.time_ranges || []).length === 0 && (
          <div className='fx-alert-effective-hint'>未配置时间段表示全天生效</div>
        )}
        {(config.time_ranges || []).map((range, index) => (
          <TimeRangeRow
            key={index}
            range={range}
            index={index}
            onChange={updateRange}
            onRemove={removeRange}
          />
        ))}
      </div>
    </div>
  )
}
