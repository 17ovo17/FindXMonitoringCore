import React, { useEffect, useState, useCallback } from 'react'
import { get, post } from '../api/http.js'
import { ErrorBox } from './AgentShared.jsx'

const STEPS = ['选择模板', '填写参数', '预览配置', '选择目标', '执行下发']

const PARAM_DEFAULTS = { text: '', password: '', bool: false, number: 0, array: '' }

function paramDefault(param) {
  if (param.default !== undefined && param.default !== null) return param.default
  return PARAM_DEFAULTS[param.type] ?? ''
}

export function ConfigPushWizard({ onClose }) {
  const [step, setStep] = useState(0)
  const [templates, setTemplates] = useState([])
  const [selectedTemplate, setSelectedTemplate] = useState(null)
  const [templateDetail, setTemplateDetail] = useState(null)
  const [params, setParams] = useState({})
  const [tomlPreview, setTomlPreview] = useState('')
  const [targets, setTargets] = useState('')
  const [credential, setCredential] = useState({ user: 'root', port: '22', authType: 'password', secret: '' })
  const [results, setResults] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    get('/integration-templates').then(resp => {
      const list = Array.isArray(resp) ? resp : resp?.items || resp?.data || resp?.list || []
      setTemplates(list)
    }).catch(() => setTemplates([]))
  }, [])

  const selectTemplate = useCallback(async (tpl) => {
    setSelectedTemplate(tpl)
    setLoading(true)
    setError('')
    try {
      const detail = await get(`/integration-templates/${encodeURIComponent(tpl.id)}`)
      setTemplateDetail(detail)
      const initial = {}
      ;(detail.params || []).forEach(p => { initial[p.name || p.key] = paramDefault(p) })
      setParams(initial)
    } catch (err) {
      setError(err?.message || '获取模板详情失败')
    } finally { setLoading(false) }
  }, [])

  const renderPreview = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const resp = await post('/categraf/render', {
        template_id: selectedTemplate.id,
        params,
      })
      setTomlPreview(resp?.content || resp?.toml || resp?.rendered || '')
    } catch (err) {
      setError(err?.message || '渲染配置失败')
    } finally { setLoading(false) }
  }, [selectedTemplate, params])

  const handleDeploy = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const hostList = targets.split(/[\n,;]+/).map(s => s.trim()).filter(Boolean)
      const resp = await post('/categraf/deploy', {
        template_id: selectedTemplate.id,
        params,
        hosts: hostList,
        credential: { user: credential.user, port: Number(credential.port) || 22, auth_type: credential.authType, secret: credential.secret },
      })
      setResults(resp?.results || resp?.hosts || [])
    } catch (err) {
      setError(err?.message || '下发失败')
      setResults([{ host: targets, status: 'failed', message: err?.message || '下发失败' }])
    } finally { setLoading(false) }
  }, [selectedTemplate, params, targets, credential])

  const goNext = () => {
    if (step === 1) { renderPreview(); setStep(2) }
    else { setStep(s => s + 1) }
  }

  const canNext = () => {
    if (step === 0) return selectedTemplate !== null
    if (step === 1) return true
    if (step === 2) return tomlPreview.length > 0
    if (step === 3) return targets.trim().length > 0
    return true
  }

  const setParamValue = (key, value) => setParams(prev => ({ ...prev, [key]: value }))

  return (
    <section className='fx-agent-work'>
      <div className='fx-plugin-wizard-head'>
        <h3>配置下发向导</h3>
        <button type='button' onClick={onClose}>关闭</button>
      </div>
      <nav className='fx-plugin-wizard-steps'>
        {STEPS.map((s, i) => (
          <span key={i} className={i === step ? 'is-active' : i < step ? 'is-done' : ''}>
            {i + 1}. {s}
          </span>
        ))}
      </nav>
      <ErrorBox>{error}</ErrorBox>

      {step === 0 && <StepSelectTemplate templates={templates} selected={selectedTemplate} onSelect={selectTemplate} loading={loading} />}
      {step === 1 && <StepFillParams detail={templateDetail} params={params} onChange={setParamValue} />}
      {step === 2 && <StepPreviewToml toml={tomlPreview} loading={loading} />}
      {step === 3 && <StepSelectTargets targets={targets} onTargetsChange={setTargets} credential={credential} onCredentialChange={setCredential} />}
      {step === 4 && <StepExecute results={results} loading={loading} onDeploy={handleDeploy} />}

      {step < 4 && (
        <div className='fx-plugin-wizard-nav'>
          {step > 0 && <button type='button' onClick={() => setStep(s => s - 1)}>上一步</button>}
          <button type='button' disabled={!canNext() || loading} onClick={goNext}>
            {loading ? '处理中...' : '下一步'}
          </button>
        </div>
      )}
      {step === 4 && !results && (
        <div className='fx-plugin-wizard-nav'>
          <button type='button' onClick={() => setStep(3)}>上一步</button>
          <button type='button' disabled={loading} onClick={handleDeploy}>
            {loading ? '下发中...' : '确认下发'}
          </button>
        </div>
      )}
    </section>
  )
}

/* --- Step 0: 选择模板 --- */
function StepSelectTemplate({ templates, selected, onSelect, loading }) {
  return (
    <div className='fx-plugin-wizard-body'>
      <p>选择采集模板：</p>
      {loading && <p style={{ color: 'var(--fx-text-weak, #66758d)', fontSize: 13 }}>加载中...</p>}
      <div className='fx-template-grid'>
        {templates.map(tpl => (
          <div
            key={tpl.id}
            className={`fx-template-card ${selected?.id === tpl.id ? 'is-selected' : ''}`}
            onClick={() => onSelect(tpl)}
          >
            <strong>{tpl.name || tpl.title}</strong>
            <small>{tpl.description || tpl.category || ''}</small>
            {tpl.version && <span className='fx-template-card__ver'>v{tpl.version}</span>}
          </div>
        ))}
        {templates.length === 0 && !loading && <div className='fx-agent-muted'>暂无可用模板</div>}
      </div>
    </div>
  )
}

/* --- Step 1: 填写参数 --- */
function StepFillParams({ detail, params, onChange }) {
  const paramDefs = detail?.params || []
  if (!paramDefs.length) {
    return (
      <div className='fx-plugin-wizard-body'>
        <p style={{ color: 'var(--fx-text-weak, #66758d)' }}>该模板无需额外参数配置。</p>
      </div>
    )
  }
  return (
    <div className='fx-plugin-wizard-body'>
      <p>填写模板参数：</p>
      <div className='fx-param-form'>
        {paramDefs.map(p => {
          const key = p.name || p.key
          return <ParamField key={key} param={p} value={params[key]} onChange={v => onChange(key, v)} />
        })}
      </div>
    </div>
  )
}

function ParamField({ param, value, onChange }) {
  const key = param.name || param.key
  const label = param.label || param.name || param.key
  const hint = param.description || param.hint || ''
  const type = (param.type || 'text').toLowerCase()

  if (type === 'bool' || type === 'boolean') {
    return (
      <label className='fx-param-field'>
        <span className='fx-param-field__label'>{label}</span>
        <input type='checkbox' checked={!!value} onChange={e => onChange(e.target.checked)} />
        {hint && <small className='fx-param-field__hint'>{hint}</small>}
      </label>
    )
  }
  if (type === 'number' || type === 'int' || type === 'float') {
    return (
      <label className='fx-param-field'>
        <span className='fx-param-field__label'>{label}</span>
        <input type='number' value={value ?? ''} onChange={e => onChange(Number(e.target.value))} placeholder={hint} />
        {hint && <small className='fx-param-field__hint'>{hint}</small>}
      </label>
    )
  }
  if (type === 'password' || type === 'secret') {
    return (
      <label className='fx-param-field'>
        <span className='fx-param-field__label'>{label}</span>
        <input type='password' value={value ?? ''} onChange={e => onChange(e.target.value)} placeholder={hint} autoComplete='off' />
        {hint && <small className='fx-param-field__hint'>{hint}</small>}
      </label>
    )
  }
  if (type === 'array' || type === 'list') {
    return (
      <label className='fx-param-field'>
        <span className='fx-param-field__label'>{label}</span>
        <textarea rows={3} value={Array.isArray(value) ? value.join('\n') : (value ?? '')} onChange={e => onChange(e.target.value)} placeholder={hint || '每行一个值'} />
        {hint && <small className='fx-param-field__hint'>{hint}</small>}
      </label>
    )
  }
  if (type === 'select' && Array.isArray(param.options)) {
    return (
      <label className='fx-param-field'>
        <span className='fx-param-field__label'>{label}</span>
        <select value={value ?? ''} onChange={e => onChange(e.target.value)}>
          <option value=''>请选择</option>
          {param.options.map(opt => <option key={opt.value || opt} value={opt.value || opt}>{opt.label || opt}</option>)}
        </select>
        {hint && <small className='fx-param-field__hint'>{hint}</small>}
      </label>
    )
  }
  return (
    <label className='fx-param-field'>
      <span className='fx-param-field__label'>{label}</span>
      <input type='text' value={value ?? ''} onChange={e => onChange(e.target.value)} placeholder={hint} />
      {hint && <small className='fx-param-field__hint'>{hint}</small>}
    </label>
  )
}

/* --- Step 2: 预览 TOML --- */
function StepPreviewToml({ toml, loading }) {
  return (
    <div className='fx-plugin-wizard-body'>
      <p>渲染后的 .toml 配置预览：</p>
      {loading && <p style={{ color: 'var(--fx-text-weak, #66758d)', fontSize: 13 }}>渲染中...</p>}
      <pre className='fx-toml-preview'><code>{toml || '(空)'}</code></pre>
    </div>
  )
}

/* --- Step 3: 选择目标 --- */
function StepSelectTargets({ targets, onTargetsChange, credential, onCredentialChange }) {
  const setCred = (key, val) => onCredentialChange(prev => ({ ...prev, [key]: val }))
  return (
    <div className='fx-plugin-wizard-body'>
      <p>输入目标主机 IP（每行一个或逗号分隔）：</p>
      <textarea
        className='fx-target-input'
        rows={5}
        value={targets}
        onChange={e => onTargetsChange(e.target.value)}
        placeholder={'192.168.1.10\n192.168.1.11\n192.168.1.12'}
      />
      <div className='fx-credential-form'>
        <label className='fx-param-field'>
          <span className='fx-param-field__label'>SSH 用户</span>
          <input type='text' value={credential.user} onChange={e => setCred('user', e.target.value)} />
        </label>
        <label className='fx-param-field'>
          <span className='fx-param-field__label'>SSH 端口</span>
          <input type='number' value={credential.port} onChange={e => setCred('port', e.target.value)} />
        </label>
        <label className='fx-param-field'>
          <span className='fx-param-field__label'>认证方式</span>
          <select value={credential.authType} onChange={e => setCred('authType', e.target.value)}>
            <option value='password'>密码</option>
            <option value='key'>密钥</option>
          </select>
        </label>
        <label className='fx-param-field'>
          <span className='fx-param-field__label'>{credential.authType === 'password' ? '密码' : '私钥内容'}</span>
          <input type='password' value={credential.secret} onChange={e => setCred('secret', e.target.value)} autoComplete='off' />
        </label>
      </div>
    </div>
  )
}

/* --- Step 4: 执行下发 --- */
function StepExecute({ results, loading, onDeploy }) {
  if (loading) {
    return (
      <div className='fx-plugin-wizard-body'>
        <p style={{ color: 'var(--fx-text-weak, #66758d)' }}>正在下发配置到目标主机...</p>
      </div>
    )
  }
  if (!results) {
    return (
      <div className='fx-plugin-wizard-body'>
        <p>确认无误后点击「确认下发」开始部署。</p>
      </div>
    )
  }
  return (
    <div className='fx-plugin-wizard-body'>
      <h4>下发结果</h4>
      <table className='fx-plugin-result-table'>
        <thead><tr><th>主机</th><th>状态</th><th>信息</th></tr></thead>
        <tbody>
          {results.map((r, i) => (
            <tr key={i}>
              <td>{r.host || r.ip || r.target}</td>
              <td className={r.status === 'success' ? 'is-ok' : 'is-fail'}>{r.status === 'success' ? '成功' : '失败'}</td>
              <td>{r.message || r.error || '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
