import React, { useState } from 'react'
import { displayJson, noDataPolicies, severities } from '../alertModel.js'
import { EffectiveTimeConfig } from './EffectiveTimeConfig.jsx'
import { NotifyConfig } from './NotifyConfig.jsx'
import { PipelineConfig } from './PipelineConfig.jsx'
import { TriggersConfig } from './TriggersConfig.jsx'

/**
 * 规则编辑表单
 * 集成生效时间、通知配置、Pipeline、Triggers 四个完整区域
 */
export function RuleFormModal({ draft, setDraft, saving, error, onSubmit, onClose, onTryRun }) {
  const updateField = (field, value) => setDraft({ ...draft, [field]: value })
  const [fieldErrors, setFieldErrors] = useState({})

  const validate = () => {
    const errors = {}
    if (!draft.name?.trim()) errors.name = '规则名称不能为空'
    if (!draft.query?.trim()) errors.query = '查询语句不能为空'
    setFieldErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleSubmit = () => {
    if (validate()) onSubmit()
  }

  return (
    <div className='fx-alert-modal'>
      <div className='fx-alert-modal__body'>
        <header>
          <h2>{draft.id ? '编辑规则' : '新建规则'}</h2>
          <button type='button' onClick={onClose}>关闭</button>
        </header>
        <div className='fx-alert-form-sections'>
          <section>
            <h3>基础信息</h3>
            <div className='fx-alert-form'>
              <label><span className='fx-field-required'>名称</span><input value={draft.name} onChange={(e) => updateField('name', e.target.value)} />{fieldErrors.name && <span className='fx-field-error'>{fieldErrors.name}</span>}</label>
              <label><span>数据源 ID</span><input value={draft.datasourceId} onChange={(e) => updateField('datasourceId', e.target.value)} /></label>
              <label><span>级别</span><select value={draft.severity} onChange={(e) => updateField('severity', e.target.value)}>{severities.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select></label>
              <label className='fx-alert-check'><input type='checkbox' checked={draft.enabled} onChange={(e) => updateField('enabled', e.target.checked)} />启用</label>
            </div>
          </section>
          <section>
            <h3>规则条件</h3>
            <div className='fx-alert-form'>
              <label className='is-wide'><span className='fx-field-required'>查询语句</span><textarea rows={5} value={draft.query} onChange={(e) => updateField('query', e.target.value)} />{fieldErrors.query && <span className='fx-field-error'>{fieldErrors.query}</span>}</label>
              <label className='is-wide'><span>目标选择器 JSON</span><textarea rows={3} value={draft.targetSelector} onChange={(e) => updateField('targetSelector', e.target.value)} /></label>
            </div>
          </section>
          <section>
            <h3>触发条件</h3>
            <TriggersConfig
              value={draft.triggers_config}
              onChange={(val) => updateField('triggers_config', val)}
              datasourceId={draft.datasourceId}
            />
          </section>
          <section>
            <h3>生效时间</h3>
            <EffectiveTimeConfig
              value={draft.effective_time}
              onChange={(val) => updateField('effective_time', val)}
            />
          </section>
          <section>
            <h3>通知配置</h3>
            <NotifyConfig
              value={draft.notify_config}
              onChange={(val) => updateField('notify_config', val)}
            />
          </section>
          <section>
            <h3>事件处理 Pipeline</h3>
            <PipelineConfig
              value={draft.pipeline_config}
              onChange={(val) => updateField('pipeline_config', val)}
            />
          </section>
          <section>
            <h3>标签 / 注解</h3>
            <div className='fx-alert-form'>
              <label className='is-wide'><span>标签 JSON</span><textarea rows={3} value={draft.labels} onChange={(e) => updateField('labels', e.target.value)} /></label>
              <label className='is-wide'><span>注解 JSON</span><textarea rows={3} value={draft.annotations} onChange={(e) => updateField('annotations', e.target.value)} /></label>
            </div>
          </section>
        </div>
        {error && <div className='fx-alert-message is-error'>{error}</div>}
        <div className='fx-alert-actions'>
          <button type='button' onClick={onClose}>取消</button>
          <button type='button' onClick={onTryRun}>试运行</button>
          <button type='button' className='is-primary' disabled={saving} onClick={handleSubmit}>{saving ? '保存中...' : '保存'}</button>
        </div>
      </div>
    </div>
  )
}
