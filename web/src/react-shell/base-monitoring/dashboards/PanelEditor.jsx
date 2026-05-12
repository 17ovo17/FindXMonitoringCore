import React, { useState, useMemo, useEffect } from 'react'
import PanelChart from './PanelChart.jsx'
import { ValueMappingsEditor, OverridesEditor, DataLinksEditor } from './FieldEditors.jsx'
import { TimeseriesOptions, StatOptions, TableOptions, PieOptions, HexbinOptions, BarGaugeOptions, TextOptions, GaugeOptions, HeatmapOptions, BarChartOptions, IframeOptions, TableNGOptions } from './TypeOptions.jsx'
import { UNIT_CATEGORIES } from './unitFormat.js'

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

/* DEGRADE-010: 数据源选择器 */
function DatasourceSelector({ value, onChange }) {
  const [datasources, setDatasources] = useState([])
  useEffect(() => {
    fetch('/api/v1/monitor/datasources').then((r) => r.json()).then((data) => {
      const list = data?.dat || data?.data || data || []
      setDatasources(Array.isArray(list) ? list : [])
    }).catch(() => setDatasources([]))
  }, [])
  return (
    <label className="fx-pe-field">
      <span>数据源</span>
      <select value={value || ''} onChange={(e) => onChange(e.target.value)}>
        <option value="">默认</option>
        {datasources.map((ds) => (
          <option key={ds.id || ds.name} value={ds.id || ds.name}>{ds.name || ds.id}</option>
        ))}
      </select>
    </label>
  )
}

/* DEGRADE-011: Monaco 查询编辑器 */
function MonacoQueryEditor({ value, onChange }) {
  const [MonacoEditor, setMonacoEditor] = useState(null)
  useEffect(() => {
    import('@monaco-editor/react').then((mod) => setMonacoEditor(() => mod.default))
  }, [])
  if (!MonacoEditor) {
    return <textarea className="fx-pe-query-input" value={value} onChange={(e) => onChange(e.target.value)} rows={4} placeholder="输入 PromQL 表达式" />
  }
  return (
    <MonacoEditor height="80px" language="promql" theme="vs-light" value={value}
      onChange={(v) => onChange(v || '')}
      options={{ minimap: { enabled: false }, lineNumbers: 'off', scrollBeyondLastLine: false, fontSize: 13, wordWrap: 'on', automaticLayout: true }}
    />
  )
}

function QueryEditor({ targets, onChange }) {
  const addTarget = () => onChange([...targets, { expr: '', legendFormat: '' }])
  const updateTarget = (index, field, value) => onChange(targets.map((t, i) => i === index ? { ...t, [field]: value } : t))
  const removeTarget = (index) => onChange(targets.filter((_, i) => i !== index))

  return (
    <div className="fx-pe-queries">
      <strong>查询</strong>
      {targets.map((target, i) => (
        <div key={i} className="fx-pe-query-row">
          <div className="fx-pe-query-row__head">
            <span className="fx-pe-query-letter">{String.fromCharCode(65 + i)}</span>
            <button type="button" className="fx-pe-query-remove" onClick={() => removeTarget(i)}>x</button>
          </div>
          <MonacoQueryEditor value={target.expr} onChange={(v) => updateTarget(i, 'expr', v)} />
          <input className="fx-pe-legend-input" placeholder={'Legend 格式 (如 {{instance}})'} value={target.legendFormat || ''} onChange={(e) => updateTarget(i, 'legendFormat', e.target.value)} />
        </div>
      ))}
      <button type="button" className="fx-pe-add-query" onClick={addTarget}>+ 添加查询</button>
    </div>
  )
}

/* DEGRADE-014: 完整单位选择器 */
function UnitPicker({ value, onChange }) {
  return (
    <label className="fx-pe-field">
      <span>数值单位</span>
      <select value={value || 'none'} onChange={(e) => onChange(e.target.value)}>
        {UNIT_CATEGORIES.map((cat) => (
          <optgroup key={cat.label} label={cat.label}>
            {cat.units.map((u) => <option key={u.key} value={u.key}>{u.label}</option>)}
          </optgroup>
        ))}
      </select>
    </label>
  )
}

function ThresholdsEditor({ thresholds, onChange }) {
  const addThreshold = () => onChange([...thresholds, { value: 0, color: '#e6550d' }])
  const updateThreshold = (index, field, value) => onChange(thresholds.map((t, i) => i === index ? { ...t, [field]: value } : t))
  const removeThreshold = (index) => onChange(thresholds.filter((_, i) => i !== index))
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

/* 根据图表类型渲染对应选项面板 (DEGRADE-009) */
function TypeSpecificOptions({ panelType, options, onChange }) {
  switch (panelType) {
    case 'timeseries': return <TimeseriesOptions options={options} onChange={onChange} />
    case 'stat': return <StatOptions options={options} onChange={onChange} />
    case 'table': return <TableOptions options={options} onChange={onChange} />
    case 'pie': return <PieOptions options={options} onChange={onChange} />
    case 'hexbin': return <HexbinOptions options={options} onChange={onChange} />
    case 'barGauge': return <BarGaugeOptions options={options} onChange={onChange} />
    case 'text': return <TextOptions options={options} onChange={onChange} />
    case 'gauge': return <GaugeOptions options={options} onChange={onChange} />
    case 'heatmap': return <HeatmapOptions options={options} onChange={onChange} />
    case 'barchart': return <BarChartOptions options={options} onChange={onChange} />
    case 'iframe': return <IframeOptions options={options} onChange={onChange} />
    case 'tableNG': return <TableNGOptions options={options} onChange={onChange} />
    default: return null
  }
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
  const [displayOpts, setDisplayOpts] = useState(initial.displayOptions || {})
  const [thresholds, setThresholds] = useState(initial.thresholds || [])
  const [unit, setUnit] = useState(initial.unit || 'none')
  const [repeatVariable, setRepeatVariable] = useState(initial.repeat || '')
  const [selectedDatasource, setSelectedDatasource] = useState(datasourceId || '')
  const [valueMappings, setValueMappings] = useState(initial.valueMappings || [])
  const [overrides, setOverrides] = useState(initial.overrides || [])
  const [dataLinks, setDataLinks] = useState(initial.dataLinks || [])

  const variableNames = useMemo(() => {
    if (!dashboardVariables || !Array.isArray(dashboardVariables)) return []
    return dashboardVariables.map((v) => v.name || v.key).filter(Boolean)
  }, [dashboardVariables])

  const previewPanel = useMemo(() => ({
    type: panelType, title, targets, displayOptions: displayOpts, unit,
    raw: { type: panelType, title, targets, displayOptions: displayOpts, unit, content: displayOpts.content, url: displayOpts.url },
  }), [panelType, title, targets, displayOpts, unit])

  const handleSave = () => {
    onSave({
      ...initial, type: panelType, title, description, targets,
      displayOptions: displayOpts, thresholds, unit,
      repeat: repeatVariable || undefined,
      valueMappings, overrides, dataLinks,
    })
  }

  return (
    <div className="fx-pe-overlay">
      <header className="fx-pe-header">
        <div className="fx-pe-header__left">
          <DatasourceSelector value={selectedDatasource} onChange={setSelectedDatasource} />
          <div className="fx-pe-type-selector">
            {PANEL_TYPES.map((t) => (
              <button key={t.key} type="button" className={panelType === t.key ? 'is-active' : ''} onClick={() => setPanelType(t.key)}>{t.label}</button>
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
            <PanelChart panel={previewPanel} timeRange={timeRange} datasourceId={selectedDatasource || datasourceId} />
          </div>
          <QueryEditor targets={targets} onChange={setTargets} />
        </div>
        <aside className="fx-pe-sidebar">
          <div className="fx-pe-section">
            <strong>基本设置</strong>
            <label className="fx-pe-field"><span>标题</span><input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="面板标题" /></label>
            <label className="fx-pe-field"><span>描述</span><textarea value={description} onChange={(e) => setDescription(e.target.value)} placeholder="面板描述" rows={2} /></label>
          </div>
          <TypeSpecificOptions panelType={panelType} options={displayOpts} onChange={setDisplayOpts} />
          <ThresholdsEditor thresholds={thresholds} onChange={setThresholds} />
          <div className="fx-pe-section"><strong>单位</strong><UnitPicker value={unit} onChange={setUnit} /></div>
          <ValueMappingsEditor mappings={valueMappings} onChange={setValueMappings} />
          <OverridesEditor overrides={overrides} onChange={setOverrides} />
          <DataLinksEditor links={dataLinks} onChange={setDataLinks} />
          <div className="fx-pe-section">
            <strong>重复</strong>
            <label className="fx-pe-field"><span>按变量重复</span>
              <select value={repeatVariable} onChange={(e) => setRepeatVariable(e.target.value)}>
                <option value="">不重复</option>
                {variableNames.map((name) => <option key={name} value={name}>{name}</option>)}
              </select>
            </label>
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
