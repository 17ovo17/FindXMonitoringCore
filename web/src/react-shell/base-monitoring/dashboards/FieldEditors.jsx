import React, { useState } from 'react'

/**
 * DEGRADE-012: ValueMappings 编辑器
 * 值到文本/颜色的映射规则列表
 */
export function ValueMappingsEditor({ mappings = [], onChange }) {
  const addMapping = () => {
    onChange([...mappings, { type: 'value', value: '', text: '', color: '#1769ff' }])
  }
  const update = (index, field, val) => {
    onChange(mappings.map((m, i) => i === index ? { ...m, [field]: val } : m))
  }
  const remove = (index) => onChange(mappings.filter((_, i) => i !== index))

  return (
    <div className="fx-pe-section">
      <strong>值映射 (Value Mappings)</strong>
      {mappings.map((m, i) => (
        <div key={i} className="fx-mapping-row">
          <select value={m.type || 'value'} onChange={(e) => update(i, 'type', e.target.value)}>
            <option value="value">精确值</option>
            <option value="range">范围</option>
            <option value="regex">正则</option>
            <option value="special">特殊值</option>
          </select>
          {m.type === 'range' ? (
            <>
              <input type="number" placeholder="从" value={m.from || ''} onChange={(e) => update(i, 'from', e.target.value)} />
              <input type="number" placeholder="到" value={m.to || ''} onChange={(e) => update(i, 'to', e.target.value)} />
            </>
          ) : (
            <input placeholder="匹配值" value={m.value || ''} onChange={(e) => update(i, 'value', e.target.value)} />
          )}
          <input placeholder="显示文本" value={m.text || ''} onChange={(e) => update(i, 'text', e.target.value)} />
          <input type="color" value={m.color || '#1769ff'} onChange={(e) => update(i, 'color', e.target.value)} />
          <button type="button" onClick={() => remove(i)}>x</button>
        </div>
      ))}
      <button type="button" className="fx-pe-add-query" onClick={addMapping}>+ 添加映射</button>
    </div>
  )
}

/**
 * DEGRADE-012: Overrides 编辑器
 * 按系列名/正则匹配覆盖样式
 */
export function OverridesEditor({ overrides = [], onChange }) {
  const addOverride = () => {
    onChange([...overrides, { matcher: { type: 'byName', value: '' }, properties: [] }])
  }
  const updateMatcher = (index, field, val) => {
    onChange(overrides.map((o, i) => i === index ? { ...o, matcher: { ...o.matcher, [field]: val } } : o))
  }
  const addProperty = (index) => {
    const next = [...overrides]
    next[index] = { ...next[index], properties: [...next[index].properties, { key: 'color', value: '#1769ff' }] }
    onChange(next)
  }
  const updateProperty = (oIdx, pIdx, field, val) => {
    const next = [...overrides]
    next[oIdx] = {
      ...next[oIdx],
      properties: next[oIdx].properties.map((p, i) => i === pIdx ? { ...p, [field]: val } : p),
    }
    onChange(next)
  }
  const removeOverride = (index) => onChange(overrides.filter((_, i) => i !== index))
  const removeProperty = (oIdx, pIdx) => {
    const next = [...overrides]
    next[oIdx] = { ...next[oIdx], properties: next[oIdx].properties.filter((_, i) => i !== pIdx) }
    onChange(next)
  }

  return (
    <div className="fx-pe-section">
      <strong>覆盖 (Overrides)</strong>
      {overrides.map((o, i) => (
        <div key={i} className="fx-override-row">
          <div className="fx-override-row__matcher">
            <select value={o.matcher?.type || 'byName'} onChange={(e) => updateMatcher(i, 'type', e.target.value)}>
              <option value="byName">按名称</option>
              <option value="byRegex">按正则</option>
              <option value="byType">按类型</option>
            </select>
            <input placeholder="匹配值" value={o.matcher?.value || ''} onChange={(e) => updateMatcher(i, 'value', e.target.value)} />
            <button type="button" onClick={() => removeOverride(i)}>x</button>
          </div>
          {o.properties.map((p, pi) => (
            <div key={pi} className="fx-override-row__prop">
              <select value={p.key} onChange={(e) => updateProperty(i, pi, 'key', e.target.value)}>
                <option value="color">颜色</option>
                <option value="lineWidth">线宽</option>
                <option value="fillOpacity">填充</option>
                <option value="pointSize">点大小</option>
                <option value="hidden">隐藏</option>
              </select>
              {p.key === 'color' ? (
                <input type="color" value={p.value || '#1769ff'} onChange={(e) => updateProperty(i, pi, 'value', e.target.value)} />
              ) : p.key === 'hidden' ? (
                <input type="checkbox" checked={p.value === true || p.value === 'true'} onChange={(e) => updateProperty(i, pi, 'value', e.target.checked)} />
              ) : (
                <input type="number" value={p.value || ''} onChange={(e) => updateProperty(i, pi, 'value', e.target.value)} />
              )}
              <button type="button" onClick={() => removeProperty(i, pi)}>x</button>
            </div>
          ))}
          <button type="button" className="fx-pe-add-query" style={{ marginTop: 4 }} onClick={() => addProperty(i)}>+ 属性</button>
        </div>
      ))}
      <button type="button" className="fx-pe-add-query" onClick={addOverride}>+ 添加覆盖</button>
    </div>
  )
}

/**
 * DEGRADE-012: DataLinks 编辑器
 * 点击数据点时跳转的链接模板
 */
export function DataLinksEditor({ links = [], onChange }) {
  const addLink = () => {
    onChange([...links, { title: '', url: '', targetBlank: true }])
  }
  const update = (index, field, val) => {
    onChange(links.map((l, i) => i === index ? { ...l, [field]: val } : l))
  }
  const remove = (index) => onChange(links.filter((_, i) => i !== index))

  return (
    <div className="fx-pe-section">
      <strong>数据链接 (Data Links)</strong>
      {links.map((l, i) => (
        <div key={i} className="fx-datalink-row">
          <input placeholder="链接标题" value={l.title || ''} onChange={(e) => update(i, 'title', e.target.value)} />
          <input placeholder="URL 模板 (支持 ${__value} ${__name})" value={l.url || ''} onChange={(e) => update(i, 'url', e.target.value)} />
          <label className="fx-settings-var-multi">
            <input type="checkbox" checked={l.targetBlank !== false} onChange={(e) => update(i, 'targetBlank', e.target.checked)} />
            <span>新窗口</span>
          </label>
          <button type="button" onClick={() => remove(i)}>x</button>
        </div>
      ))}
      <button type="button" className="fx-pe-add-query" onClick={addLink}>+ 添加链接</button>
    </div>
  )
}
