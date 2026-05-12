import React, { useState, useMemo } from 'react'
import PanelChart from './PanelChart.jsx'

const PANEL_TYPES = [
  { key: 'timeseries', label: '时序图' },
  { key: 'barchart', label: '柱状图' },
  { key: 'stat', label: '指标值' },
  { key: 'tableNG', label: '表格 NG' },
  { key: 'table', label: '表格' },
  { key: 'pie', label: '饼图' },
  { key: 'hexbin', label: '蜂窝图' },
  { key: 'barGauge', label: '排行榜' },
  { key: 'text', label: '文本卡片' },
  { key: 'gauge', label: '仪表图' },
  { key: 'heatmap', label: '色块图' },
  { key: 'iframe', label: '内嵌文档' },
]

const STACK_MODES = [
  { key: 'none', label: '不堆叠' },
  { key: 'normal', label: '普通' },
  { key: 'percent', label: '百分比' },
]

const LEGEND_POSITIONS = [
  { key: 'bottom', label: '底部' },
  { key: 'right', label: '右侧' },
  { key: 'hidden', label: '隐藏' },
]

const UNIT_OPTIONS = [
  { key: 'none', label: '无' },
  { key: 'bytes', label: 'bytes' },
  { key: 'percent', label: '%' },
  { key: 'seconds', label: '秒 (s)' },
  { key: 'reqps', label: 'requests/s' },
  { key: 'ms', label: '毫秒 (ms)' },
]

function QueryEditor({ targets, onChange }) {
  const addTarget = () => {
    onChange([...targets, { expr: '', legendFormat: '' }])
  }
  const updateTarget = (index, field, value) => {
    const next = targets.map((t, i) => i === index ? { ...t, [field]: value } : t)
    onChange(next)
  }
  const removeTarget = (index) => {
    onChange(targets.filter((_, i) => i !== index))
  }

  return (
    <div className="fx-pe-queries">
      <strong>查询</strong>
      {targets.map((target, i) => (
        <div key={i} className="fx-pe-query-row">
          <div className="fx-pe-query-row__head">
            <span className="fx-pe-query-letter">{String.fromCharCode(65 + i)}</span>
            <button type="button" className="fx-pe-query-remove" onClick={() => removeTarget(i)}>x</button>
          </div>
          <textarea
            className="fx-pe-query-input"
            placeholder="输入 PromQL 表达式"
            value={target.expr}
            onChange={(e) => updateTarget(i, 'expr', e.target.value)}
            rows={3}
          />
          <input
            className="fx-pe-legend-input"
            placeholder={'Legend 格式 (如 {{instance}})'}
            value={target.legendFormat || ''}
            onChange={(e) => updateTarget(i, 'legendFormat', e.target.value)}
          />
        </div>
      ))}
      <button type="button" className="fx-pe-add-query" onClick={addTarget}>+ 添加查询</button>
    </div>
  )
}

function DisplayOptions({ options, onChange }) {
  const update = (key, value) => onChange({ ...options, [key]: value })

  return (
    <div className="fx-pe-section">
      <strong>显示选项</strong>
      <label className="fx-pe-field">
        <span>线宽</span>
        <input type="range" min="1" max="5" value={options.lineWidth || 1} onChange={(e) => update('lineWidth', Number(e.target.value))} />
        <span className="fx-pe-field__value">{options.lineWidth || 1}px</span>
      </label>
      <label className="fx-pe-field">
        <span>填充透明度</span>
        <input type="range" min="0" max="100" value={options.fillOpacity || 10} onChange={(e) => update('fillOpacity', Number(e.target.value))} />
        <span className="fx-pe-field__value">{options.fillOpacity || 10}%</span>
      </label>
      <label className="fx-pe-field">
        <span>点大小</span>
        <input type="range" min="0" max="10" value={options.pointSize || 0} onChange={(e) => update('pointSize', Number(e.target.value))} />
        <span className="fx-pe-field__value">{options.pointSize || 0}</span>
      </label>
      <label className="fx-pe-field">
        <span>堆叠模式</span>
        <select value={options.stackMode || 'none'} onChange={(e) => update('stackMode', e.target.value)}>
          {STACK_MODES.map((m) => <option key={m.key} value={m.key}>{m.label}</option>)}
        </select>
      </label>
      <label className="fx-pe-field">
        <span>图例位置</span>
        <select value={options.legendPosition || 'bottom'} onChange={(e) => update('legendPosition', e.target.value)}>
          {LEGEND_POSITIONS.map((p) => <option key={p.key} value={p.key}>{p.label}</option>)}
        </select>
      </label>
    </div>
  )
}

function ThresholdsEditor({ thresholds, onChange }) {
  const addThreshold = () => {
    onChange([...thresholds, { value: 0, color: '#e6550d' }])
  }
  const updateThreshold = (index, field, value) => {
    const next = thresholds.map((t, i) => i === index ? { ...t, [field]: value } : t)
    onChange(next)
  }
  const removeThreshold = (index) => {
    onChange(thresholds.filter((_, i) => i !== index))
  }

  return (
    <div className="fx-pe-section">
      <strong>阈值</strong>
      {thresholds.map((t, i) => (
        <div key={i} className="fx-pe-threshold-row">
          <input type="color" value={t.color} onChange={(e) => updateThreshold(i, 'color', e.target.value)} />
          <input type="number" value={t.value} onChange={(e) => updateThreshold(i, 'value', Number(e.target.value))} placeholder="值" />
          <button type="button" onClick={() => removeThreshold(i)}>x</button>
        </div>
      ))}
      <button type="button" className="fx-pe-add-query" onClick={addThreshold}>+ 添加阈值</button>
    </div>
  )
}

export default function PanelEditor({ panel, timeRange, datasourceId, dashboardVariables, onSave, onClose }) {
  const initial = panel || {}
  const initialTargets = Array.isArray(initial.targets)
    ? initial.targets
    : (initial.expr ? [{ expr: initial.expr, legendFormat: '' }] : [{ expr: '', legendFormat: '' }])

  const [panelType, setPanelType] = useState(initial.type || 'timeseries')
  const [title, setTitle] = useState(initial.title || '')
  const [description, setDescription] = useState(initial.description || '')
  const [targets, setTargets] = useState(initialTargets)
  const [displayOpts, setDisplayOpts] = useState(
    initial.displayOptions || { lineWidth: 1, fillOpacity: 10, pointSize: 0, stackMode: 'none', legendPosition: 'bottom' }
  )
  const [thresholds, setThresholds] = useState(initial.thresholds || [])
  const [unit, setUnit] = useState(initial.unit || 'none')
  const [repeatVariable, setRepeatVariable] = useState(initial.repeat || '')

  const variableNames = useMemo(() => {
    if (!dashboardVariables || !Array.isArray(dashboardVariables)) return []
    return dashboardVariables.map((v) => v.name || v.key).filter(Boolean)
  }, [dashboardVariables])

  const previewPanel = useMemo(() => ({
    type: panelType,
    title,
    targets,
    raw: { type: panelType, title, targets },
  }), [panelType, title, targets])

  const handleSave = () => {
    onSave({
      ...initial,
      type: panelType,
      title,
      description,
      targets,
      displayOptions: displayOpts,
      thresholds,
      unit,
      repeat: repeatVariable || undefined,
    })
  }

  return (
    <div className="fx-pe-overlay">
      <header className="fx-pe-header">
        <div className="fx-pe-header__left">
          <div className="fx-pe-type-selector">
            {PANEL_TYPES.map((t) => (
              <button
                key={t.key}
                type="button"
                className={panelType === t.key ? 'is-active' : ''}
                onClick={() => setPanelType(t.key)}
              >
                {t.label}
              </button>
            ))}
          </div>
        </div>
        <div className="fx-pe-header__right">
          <button type="button" onClick={onClose}>关闭</button>
        </div>
      </header>
      <div className="fx-pe-body">
        <div className="fx-pe-main">
          <div className="fx-pe-preview">
            <PanelChart panel={previewPanel} timeRange={timeRange} datasourceId={datasourceId} />
          </div>
          <QueryEditor targets={targets} onChange={setTargets} />
        </div>
        <aside className="fx-pe-sidebar">
          <div className="fx-pe-section">
            <strong>基本设置</strong>
            <label className="fx-pe-field">
              <span>标题</span>
              <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="面板标题" />
            </label>
            <label className="fx-pe-field">
              <span>描述</span>
              <textarea value={description} onChange={(e) => setDescription(e.target.value)} placeholder="面板描述" rows={2} />
            </label>
          </div>
          {panelType === 'timeseries' && (
            <DisplayOptions options={displayOpts} onChange={setDisplayOpts} />
          )}
          <ThresholdsEditor thresholds={thresholds} onChange={setThresholds} />
          <div className="fx-pe-section">
            <strong>单位</strong>
            <label className="fx-pe-field">
              <span>数值单位</span>
              <select value={unit} onChange={(e) => setUnit(e.target.value)}>
                {UNIT_OPTIONS.map((u) => <option key={u.key} value={u.key}>{u.label}</option>)}
              </select>
            </label>
          </div>
          <div className="fx-pe-section">
            <strong>重复</strong>
            <label className="fx-pe-field">
              <span>按变量重复</span>
              <select value={repeatVariable} onChange={(e) => setRepeatVariable(e.target.value)}>
                <option value="">不重复</option>
                {variableNames.map((name) => <option key={name} value={name}>{name}</option>)}
              </select>
            </label>
            {repeatVariable && <p style={{ fontSize: 11, color: 'var(--fx-muted)', margin: '4px 0 0' }}>Panel 将按变量「{repeatVariable}」的每个值重复生成</p>}
          </div>
        </aside>
      </div>
      <footer className="fx-pe-footer">
        <button type="button" onClick={onClose}>取消</button>
        <button type="button" className="is-primary" onClick={handleSave}>保存</button>
      </footer>
    </div>
  )
}
