import React, { useState } from 'react'

/**
 * ConfigParamsForm — 根据 params.json 的参数定义动态生成表单
 * 支持类型：string, password, bool, int, duration, enum, string_array
 */

const DURATION_UNITS = [
  { value: 's', label: '秒' },
  { value: 'm', label: '分' },
  { value: 'h', label: '时' },
]

function parseDuration(val) {
  if (!val || typeof val !== 'string') return { num: '', unit: 's' }
  const match = val.match(/^(\d+)(s|m|h)$/)
  if (match) return { num: match[1], unit: match[2] }
  return { num: val.replace(/[^0-9]/g, '') || '', unit: 's' }
}

function formatDuration(num, unit) {
  if (!num && num !== 0) return ''
  return `${num}${unit}`
}

export function ConfigParamsForm({ params, values, onChange }) {
  if (!params || params.length === 0) {
    return (
      <div className='fx-plugin-wizard-body'>
        <p className='fx-agent-muted'>该模板无需额外参数配置。</p>
      </div>
    )
  }

  return (
    <div className='fx-param-form'>
      {params.map(param => (
        <ParamField
          key={param.key}
          param={param}
          value={values[param.key]}
          onChange={v => onChange(param.key, v)}
        />
      ))}
    </div>
  )
}

function ParamField({ param, value, onChange }) {
  const { label, type, placeholder, options, required } = param
  const fieldType = (type || 'string').toLowerCase()

  // string_array: 多行输入或逗号分隔
  if (fieldType === 'string_array') {
    return <StringArrayField label={label} value={value} onChange={onChange} placeholder={placeholder} required={required} />
  }

  // enum: 下拉选择
  if (fieldType === 'enum' || (fieldType === 'string' && Array.isArray(options) && options.length > 0)) {
    return <EnumField label={label} value={value} onChange={onChange} options={options} required={required} />
  }

  // bool: 复选框
  if (fieldType === 'bool' || fieldType === 'boolean') {
    return <BoolField label={label} value={value} onChange={onChange} />
  }

  // int / number: 数字输入
  if (fieldType === 'int' || fieldType === 'number' || fieldType === 'float') {
    return <IntField label={label} value={value} onChange={onChange} placeholder={placeholder} required={required} />
  }

  // duration: 数字 + 单位选择
  if (fieldType === 'duration') {
    return <DurationField label={label} value={value} onChange={onChange} required={required} />
  }

  // password: 密码输入 + 显示/隐藏切换
  if (fieldType === 'password' || fieldType === 'secret') {
    return <PasswordField label={label} value={value} onChange={onChange} placeholder={placeholder} required={required} />
  }

  // string (默认): 文本输入
  return <StringField label={label} value={value} onChange={onChange} placeholder={placeholder} required={required} />
}

/* --- 各类型字段组件 --- */

function StringField({ label, value, onChange, placeholder, required }) {
  return (
    <label className='fx-param-field'>
      <span className='fx-param-field__label'>
        {label}{required && <em className='fx-param-required'>*</em>}
      </span>
      <input
        type='text'
        value={value ?? ''}
        onChange={e => onChange(e.target.value)}
        placeholder={placeholder || ''}
      />
    </label>
  )
}

function PasswordField({ label, value, onChange, placeholder, required }) {
  const [visible, setVisible] = useState(false)
  return (
    <label className='fx-param-field'>
      <span className='fx-param-field__label'>
        {label}{required && <em className='fx-param-required'>*</em>}
      </span>
      <span className='fx-param-password-wrap'>
        <input
          type={visible ? 'text' : 'password'}
          value={value ?? ''}
          onChange={e => onChange(e.target.value)}
          placeholder={placeholder || ''}
          autoComplete='off'
        />
        <button
          type='button'
          className='fx-param-password-toggle'
          onClick={() => setVisible(v => !v)}
          title={visible ? '隐藏' : '显示'}
        >
          {visible ? '隐藏' : '显示'}
        </button>
      </span>
    </label>
  )
}

function BoolField({ label, value, onChange }) {
  const checked = value === true || value === 'true'
  return (
    <label className='fx-param-field fx-param-field--bool'>
      <span className='fx-param-field__label'>{label}</span>
      <input
        type='checkbox'
        checked={checked}
        onChange={e => onChange(e.target.checked ? 'true' : 'false')}
      />
      <span className='fx-param-bool-text'>{checked ? '是' : '否'}</span>
    </label>
  )
}

function IntField({ label, value, onChange, placeholder, required }) {
  return (
    <label className='fx-param-field'>
      <span className='fx-param-field__label'>
        {label}{required && <em className='fx-param-required'>*</em>}
      </span>
      <input
        type='number'
        value={value ?? ''}
        onChange={e => onChange(e.target.value)}
        placeholder={placeholder || ''}
      />
    </label>
  )
}

function DurationField({ label, value, onChange, required }) {
  const { num, unit } = parseDuration(value)
  const handleNumChange = (e) => {
    onChange(formatDuration(e.target.value, unit))
  }
  const handleUnitChange = (e) => {
    onChange(formatDuration(num, e.target.value))
  }
  return (
    <label className='fx-param-field'>
      <span className='fx-param-field__label'>
        {label}{required && <em className='fx-param-required'>*</em>}
      </span>
      <span className='fx-param-duration-wrap'>
        <input
          type='number'
          min='0'
          value={num}
          onChange={handleNumChange}
          className='fx-param-duration-num'
        />
        <select value={unit} onChange={handleUnitChange} className='fx-param-duration-unit'>
          {DURATION_UNITS.map(u => (
            <option key={u.value} value={u.value}>{u.label}</option>
          ))}
        </select>
      </span>
    </label>
  )
}

function EnumField({ label, value, onChange, options, required }) {
  const optionList = (options || []).map(opt =>
    typeof opt === 'string' ? { value: opt, label: opt } : opt
  )
  return (
    <label className='fx-param-field'>
      <span className='fx-param-field__label'>
        {label}{required && <em className='fx-param-required'>*</em>}
      </span>
      <select value={value ?? ''} onChange={e => onChange(e.target.value)}>
        <option value=''>请选择</option>
        {optionList.map(opt => (
          <option key={opt.value} value={opt.value}>{opt.label}</option>
        ))}
      </select>
    </label>
  )
}

function StringArrayField({ label, value, onChange, placeholder, required }) {
  const displayValue = Array.isArray(value) ? value.join('\n') : (value ?? '')
  return (
    <label className='fx-param-field'>
      <span className='fx-param-field__label'>
        {label}{required && <em className='fx-param-required'>*</em>}
      </span>
      <textarea
        rows={3}
        value={displayValue}
        onChange={e => onChange(e.target.value)}
        placeholder={placeholder || '每行一个值，或用逗号分隔'}
      />
      <small className='fx-param-field__hint'>支持逗号分隔或每行一个值</small>
    </label>
  )
}
