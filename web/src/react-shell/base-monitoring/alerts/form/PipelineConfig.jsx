import React from 'react'

/**
 * Pipeline 配置组件
 * 对齐夜莺 PipelineConfigsNG：Relabel 规则、Annotations 模板、Enrich Queries
 */

function RelabelItem({ item, index, onChange, onRemove }) {
  const update = (field, val) => onChange(index, { ...item, [field]: val })
  return (
    <div className='fx-alert-pipeline-row'>
      <input value={item.source_labels || ''} onChange={(e) => update('source_labels', e.target.value)} placeholder='source_labels（逗号分隔）' />
      <input value={item.regex || ''} onChange={(e) => update('regex', e.target.value)} placeholder='regex' />
      <input value={item.target_label || ''} onChange={(e) => update('target_label', e.target.value)} placeholder='target_label' />
      <input value={item.replacement || ''} onChange={(e) => update('replacement', e.target.value)} placeholder='replacement' />
      <select value={item.action || 'replace'} onChange={(e) => update('action', e.target.value)}>
        <option value='replace'>replace</option>
        <option value='keep'>keep</option>
        <option value='drop'>drop</option>
        <option value='labelmap'>labelmap</option>
      </select>
      <button type='button' onClick={() => onRemove(index)}>删除</button>
    </div>
  )
}

function AnnotationItem({ item, index, onChange, onRemove }) {
  return (
    <div className='fx-alert-pipeline-row'>
      <input
        value={item.key || ''}
        onChange={(e) => onChange(index, { ...item, key: e.target.value })}
        placeholder='key（如 summary, runbook_url）'
        style={{ maxWidth: 200 }}
      />
      <textarea
        rows={2}
        value={item.value || ''}
        onChange={(e) => onChange(index, { ...item, value: e.target.value })}
        placeholder='value（支持 Go template 语法）'
        style={{ flex: 1 }}
      />
      <button type='button' onClick={() => onRemove(index)}>删除</button>
    </div>
  )
}

function EnrichQueryItem({ item, index, onChange, onRemove }) {
  return (
    <div className='fx-alert-pipeline-row'>
      <input
        value={item.query || ''}
        onChange={(e) => onChange(index, { ...item, query: e.target.value })}
        placeholder='附加查询语句'
        style={{ flex: 1 }}
      />
      <button type='button' onClick={() => onRemove(index)}>删除</button>
    </div>
  )
}

export function PipelineConfig({ value, onChange }) {
  const config = value || { relabel_configs: [], annotations: [], enrich_queries: [] }
  const update = (patch) => onChange?.({ ...config, ...patch })

  const updateListItem = (listKey, index, item) => {
    const list = [...(config[listKey] || [])]
    list[index] = item
    update({ [listKey]: list })
  }

  const removeListItem = (listKey, index) => {
    update({ [listKey]: (config[listKey] || []).filter((_, i) => i !== index) })
  }

  const addRelabel = () => {
    update({ relabel_configs: [...(config.relabel_configs || []), { source_labels: '', regex: '(.*)', target_label: '', replacement: '$1', action: 'replace' }] })
  }

  const addAnnotation = () => {
    update({ annotations: [...(config.annotations || []), { key: '', value: '' }] })
  }

  const addEnrichQuery = () => {
    update({ enrich_queries: [...(config.enrich_queries || []), { query: '' }] })
  }

  return (
    <div className='fx-alert-pipeline'>
      <div className='fx-alert-pipeline-section'>
        <div className='fx-alert-effective-ranges-head'>
          <span className='fx-alert-effective-label'>Relabel 规则</span>
          <button type='button' onClick={addRelabel}>添加</button>
        </div>
        {(config.relabel_configs || []).length === 0 && (
          <div className='fx-alert-effective-hint'>暂无 Relabel 规则</div>
        )}
        {(config.relabel_configs || []).map((item, index) => (
          <RelabelItem
            key={index}
            item={item}
            index={index}
            onChange={(i, v) => updateListItem('relabel_configs', i, v)}
            onRemove={(i) => removeListItem('relabel_configs', i)}
          />
        ))}
      </div>
      <div className='fx-alert-pipeline-section'>
        <div className='fx-alert-effective-ranges-head'>
          <span className='fx-alert-effective-label'>Annotations 模板</span>
          <button type='button' onClick={addAnnotation}>添加</button>
        </div>
        {(config.annotations || []).length === 0 && (
          <div className='fx-alert-effective-hint'>暂无 Annotations</div>
        )}
        {(config.annotations || []).map((item, index) => (
          <AnnotationItem
            key={index}
            item={item}
            index={index}
            onChange={(i, v) => updateListItem('annotations', i, v)}
            onRemove={(i) => removeListItem('annotations', i)}
          />
        ))}
      </div>
      <div className='fx-alert-pipeline-section'>
        <div className='fx-alert-effective-ranges-head'>
          <span className='fx-alert-effective-label'>Enrich Queries</span>
          <button type='button' onClick={addEnrichQuery}>添加</button>
        </div>
        {(config.enrich_queries || []).length === 0 && (
          <div className='fx-alert-effective-hint'>暂无附加查询</div>
        )}
        {(config.enrich_queries || []).map((item, index) => (
          <EnrichQueryItem
            key={index}
            item={item}
            index={index}
            onChange={(i, v) => updateListItem('enrich_queries', i, v)}
            onRemove={(i) => removeListItem('enrich_queries', i)}
          />
        ))}
      </div>
    </div>
  )
}

