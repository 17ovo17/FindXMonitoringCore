import React, { useRef, useState } from 'react'

function Modal({ title, children, onClose }) {
  return <div className='fx-dash-modal'><div className='fx-dash-modal__body'><header><h2>{title}</h2><button type='button' onClick={onClose}>x</button></header>{children}</div></div>
}

/** D13: 导入 JSON 弹窗 */
export function ImportJsonModal({ onClose, onImport }) {
  const [jsonText, setJsonText] = useState('')
  const [preview, setPreview] = useState(null)
  const [parseError, setParseError] = useState('')
  const fileRef = useRef(null)

  const parseJson = (text) => {
    setJsonText(text)
    setParseError('')
    setPreview(null)
    if (!text.trim()) return
    try {
      const obj = JSON.parse(text)
      const title = obj.title || obj.name || '未命名'
      const panels = Array.isArray(obj.panels) ? obj.panels : []
      setPreview({ title, panelCount: panels.length, data: obj })
    } catch { setParseError('JSON 格式无效') }
  }

  const handleFile = (e) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = (ev) => parseJson(ev.target.result)
    reader.readAsText(file)
  }

  return (
    <Modal title='导入仪表盘 JSON' onClose={onClose}>
      <div className='fx-dash-form'>
        <label><span>粘贴 JSON</span>
          <textarea rows={6} value={jsonText} onChange={(e) => parseJson(e.target.value)} placeholder='粘贴仪表盘 JSON 配置...' />
        </label>
        <label><span>或上传文件</span>
          <input type='file' accept='.json' ref={fileRef} onChange={handleFile} />
        </label>
        {parseError && <div className='fx-dash-alert is-error'>{parseError}</div>}
        {preview && (
          <div style={{ padding: '8px', background: '#f5f7fa', borderRadius: 4, fontSize: 13 }}>
            <p>名称：<b>{preview.title}</b></p>
            <p>Panel 数量：<b>{preview.panelCount}</b></p>
          </div>
        )}
        <div className='fx-dash-actions'>
          <button type='button' onClick={onClose}>取消</button>
          <button type='button' className='is-primary' disabled={!preview} onClick={() => onImport(preview.data)}>确认导入</button>
        </div>
      </div>
    </Modal>
  )
}

/** D15: 公开确认弹窗 */
export function ShareConfirmModal({ row, onClose, onConfirm }) {
  return (
    <Modal title='公开仪表盘' onClose={onClose}>
      <div className='fx-dash-form'>
        <p>确定将仪表盘 <b>{row.title}</b> 设为公开？公开后其他用户可查看。</p>
        <div className='fx-dash-actions'>
          <button type='button' onClick={onClose}>取消</button>
          <button type='button' className='is-primary' onClick={onConfirm}>确认公开</button>
        </div>
      </div>
    </Modal>
  )
}
