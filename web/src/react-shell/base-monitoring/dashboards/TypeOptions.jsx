import React from 'react'

const CALC_OPTIONS = [
  { key: 'last', label: '最新值' },
  { key: 'first', label: '第一个值' },
  { key: 'max', label: '最大值' },
  { key: 'min', label: '最小值' },
  { key: 'avg', label: '平均值' },
  { key: 'sum', label: '总和' },
  { key: 'count', label: '计数' },
]

const TEXT_MODES = [
  { key: 'valueAndName', label: '值和名称' },
  { key: 'value', label: '仅值' },
  { key: 'name', label: '仅名称' },
]

const COLOR_MODES = [
  { key: 'value', label: '值颜色' },
  { key: 'background', label: '背景颜色' },
  { key: 'none', label: '无' },
]

const INTERPOLATION = [
  { key: 'linear', label: '线性' },
  { key: 'smooth', label: '平滑' },
  { key: 'stepBefore', label: '阶梯(前)' },
  { key: 'stepAfter', label: '阶梯(后)' },
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

function Field({ label, children }) {
  return <label className="fx-pe-field"><span>{label}</span>{children}</label>
}

function SelectField({ label, value, options, onChange }) {
  return (
    <Field label={label}>
      <select value={value || options[0]?.key} onChange={(e) => onChange(e.target.value)}>
        {options.map((o) => <option key={o.key} value={o.key}>{o.label}</option>)}
      </select>
    </Field>
  )
}

function RangeField({ label, value, min, max, onChange }) {
  return (
    <Field label={label}>
      <input type="range" min={min} max={max} value={value || min} onChange={(e) => onChange(Number(e.target.value))} />
      <span className="fx-pe-field__value">{value || min}</span>
    </Field>
  )
}

/**
 * DEGRADE-009: Timeseries 选项
 */
export function TimeseriesOptions({ options, onChange }) {
  const u = (k, v) => onChange({ ...options, [k]: v })
  return (
    <div className="fx-pe-section">
      <strong>时序图选项</strong>
      <RangeField label="线宽" value={options.lineWidth || 1} min={1} max={5} onChange={(v) => u('lineWidth', v)} />
      <RangeField label="填充透明度" value={options.fillOpacity || 10} min={0} max={100} onChange={(v) => u('fillOpacity', v)} />
      <RangeField label="点大小" value={options.pointSize || 0} min={0} max={10} onChange={(v) => u('pointSize', v)} />
      <SelectField label="堆叠模式" value={options.stackMode} options={STACK_MODES} onChange={(v) => u('stackMode', v)} />
      <SelectField label="曲线插值" value={options.interpolation} options={INTERPOLATION} onChange={(v) => u('interpolation', v)} />
      <SelectField label="图例位置" value={options.legendPosition} options={LEGEND_POSITIONS} onChange={(v) => u('legendPosition', v)} />
      <Field label="渐变填充">
        <input type="checkbox" checked={options.gradient || false} onChange={(e) => u('gradient', e.target.checked)} />
      </Field>
    </div>
  )
}

/**
 * DEGRADE-009: Stat 选项
 */
export function StatOptions({ options, onChange }) {
  const u = (k, v) => onChange({ ...options, [k]: v })
  return (
    <div className="fx-pe-section">
      <strong>指标值选项</strong>
      <SelectField label="文本模式" value={options.textMode} options={TEXT_MODES} onChange={(v) => u('textMode', v)} />
      <SelectField label="颜色模式" value={options.colorMode} options={COLOR_MODES} onChange={(v) => u('colorMode', v)} />
      <SelectField label="计算方式" value={options.calc} options={CALC_OPTIONS} onChange={(v) => u('calc', v)} />
      <SelectField label="方向" value={options.direction} options={[{ key: 'vertical', label: '垂直' }, { key: 'horizontal', label: '水平' }]} onChange={(v) => u('direction', v)} />
    </div>
  )
}

/**
 * DEGRADE-009: Table 选项
 */
export function TableOptions({ options, onChange }) {
  const u = (k, v) => onChange({ ...options, [k]: v })
  return (
    <div className="fx-pe-section">
      <strong>表格选项</strong>
      <Field label="显示表头"><input type="checkbox" checked={options.showHeader !== false} onChange={(e) => u('showHeader', e.target.checked)} /></Field>
      <SelectField label="颜色模式" value={options.colorMode} options={COLOR_MODES} onChange={(v) => u('colorMode', v)} />
      <SelectField label="计算方式" value={options.calc} options={CALC_OPTIONS} onChange={(v) => u('calc', v)} />
      <SelectField label="显示模式" value={options.displayMode} options={[{ key: 'table', label: '表格' }, { key: 'list', label: '列表' }]} onChange={(v) => u('displayMode', v)} />
    </div>
  )
}

export function PieOptions({ options, onChange }) {
  const u = (k, v) => onChange({ ...options, [k]: v })
  return (
    <div className="fx-pe-section">
      <strong>饼图选项</strong>
      <SelectField label="文本模式" value={options.textMode} options={TEXT_MODES} onChange={(v) => u('textMode', v)} />
      <SelectField label="颜色模式" value={options.colorMode} options={COLOR_MODES} onChange={(v) => u('colorMode', v)} />
      <SelectField label="计算方式" value={options.calc} options={CALC_OPTIONS} onChange={(v) => u('calc', v)} />
      <SelectField label="图例位置" value={options.legendPosition} options={LEGEND_POSITIONS} onChange={(v) => u('legendPosition', v)} />
    </div>
  )
}

export function HexbinOptions({ options, onChange }) {
  const u = (k, v) => onChange({ ...options, [k]: v })
  return (
    <div className="fx-pe-section">
      <strong>蜂窝图选项</strong>
      <SelectField label="文本模式" value={options.textMode} options={TEXT_MODES} onChange={(v) => u('textMode', v)} />
      <SelectField label="计算方式" value={options.calc} options={CALC_OPTIONS} onChange={(v) => u('calc', v)} />
      <Field label="颜色范围起始"><input type="color" value={options.colorRangeStart || '#eef5ff'} onChange={(e) => u('colorRangeStart', e.target.value)} /></Field>
      <Field label="颜色范围结束"><input type="color" value={options.colorRangeEnd || '#1769ff'} onChange={(e) => u('colorRangeEnd', e.target.value)} /></Field>
    </div>
  )
}

export function BarGaugeOptions({ options, onChange }) {
  const u = (k, v) => onChange({ ...options, [k]: v })
  return (
    <div className="fx-pe-section">
      <strong>排行榜选项</strong>
      <SelectField label="计算方式" value={options.calc} options={CALC_OPTIONS} onChange={(v) => u('calc', v)} />
      <Field label="基础颜色"><input type="color" value={options.baseColor || '#1769ff'} onChange={(e) => u('baseColor', e.target.value)} /></Field>
      <SelectField label="显示模式" value={options.displayMode} options={[{ key: 'gradient', label: '渐变' }, { key: 'solid', label: '纯色' }]} onChange={(v) => u('displayMode', v)} />
      <SelectField label="排序" value={options.sort} options={[{ key: 'desc', label: '降序' }, { key: 'asc', label: '升序' }, { key: 'none', label: '不排序' }]} onChange={(v) => u('sort', v)} />
    </div>
  )
}

export function TextOptions({ options, onChange }) {
  const u = (k, v) => onChange({ ...options, [k]: v })
  return (
    <div className="fx-pe-section">
      <strong>文本卡片选项</strong>
      <RangeField label="字号" value={options.fontSize || 14} min={10} max={48} onChange={(v) => u('fontSize', v)} />
      <Field label="文字颜色"><input type="color" value={options.textColor || '#17233c'} onChange={(e) => u('textColor', e.target.value)} /></Field>
      <Field label="背景色"><input type="color" value={options.backgroundColor || '#ffffff'} onChange={(e) => u('backgroundColor', e.target.value)} /></Field>
      <SelectField label="对齐方式" value={options.textAlign} options={[{ key: 'left', label: '左对齐' }, { key: 'center', label: '居中' }, { key: 'right', label: '右对齐' }]} onChange={(v) => u('textAlign', v)} />
      <Field label="内容 (Markdown)"><textarea value={options.content || ''} onChange={(e) => u('content', e.target.value)} rows={4} style={{ width: '100%', fontFamily: 'monospace', fontSize: 12 }} /></Field>
    </div>
  )
}

export function GaugeOptions({ options, onChange }) {
  const u = (k, v) => onChange({ ...options, [k]: v })
  return (
    <div className="fx-pe-section">
      <strong>仪表图选项</strong>
      <SelectField label="文本模式" value={options.textMode} options={TEXT_MODES} onChange={(v) => u('textMode', v)} />
      <SelectField label="计算方式" value={options.calc} options={CALC_OPTIONS} onChange={(v) => u('calc', v)} />
      <Field label="最小值"><input type="number" value={options.min || 0} onChange={(e) => u('min', Number(e.target.value))} /></Field>
      <Field label="最大值"><input type="number" value={options.max || 100} onChange={(e) => u('max', Number(e.target.value))} /></Field>
    </div>
  )
}

export function HeatmapOptions({ options, onChange }) {
  const u = (k, v) => onChange({ ...options, [k]: v })
  return (
    <div className="fx-pe-section">
      <strong>色块图选项</strong>
      <SelectField label="计算方式" value={options.calc} options={CALC_OPTIONS} onChange={(v) => u('calc', v)} />
      <SelectField label="配色方案" value={options.colorScheme} options={[{ key: 'blue', label: '蓝色' }, { key: 'green', label: '绿色' }, { key: 'red', label: '红色' }, { key: 'spectral', label: '光谱' }]} onChange={(v) => u('colorScheme', v)} />
    </div>
  )
}

export function BarChartOptions({ options, onChange }) {
  const u = (k, v) => onChange({ ...options, [k]: v })
  return (
    <div className="fx-pe-section">
      <strong>柱状图选项</strong>
      <SelectField label="计算方式" value={options.calc} options={CALC_OPTIONS} onChange={(v) => u('calc', v)} />
    </div>
  )
}

export function IframeOptions({ options, onChange }) {
  const u = (k, v) => onChange({ ...options, [k]: v })
  return (
    <div className="fx-pe-section">
      <strong>内嵌文档选项</strong>
      <Field label="URL"><input value={options.url || ''} onChange={(e) => u('url', e.target.value)} placeholder="https://..." /></Field>
    </div>
  )
}

export function TableNGOptions({ options, onChange }) {
  const u = (k, v) => onChange({ ...options, [k]: v })
  return (
    <div className="fx-pe-section">
      <strong>表格 NG 选项</strong>
      <Field label="显示表头"><input type="checkbox" checked={options.showHeader !== false} onChange={(e) => u('showHeader', e.target.checked)} /></Field>
      <SelectField label="颜色模式" value={options.colorMode} options={COLOR_MODES} onChange={(v) => u('colorMode', v)} />
      <SelectField label="计算方式" value={options.calc} options={CALC_OPTIONS} onChange={(v) => u('calc', v)} />
    </div>
  )
}
