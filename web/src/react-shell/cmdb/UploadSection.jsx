import React, { useState } from 'react'
import { CMDB_EXECUTION_BLOCKERS, cmdbApi, cmdbContractMessage, isCmdbContractBlocked } from '../api/cmdb.js'
import { Blocked, ErrorBox, Field } from './Shared.jsx'

/**
 * C04: 文件上传组件
 * 文件选择 + 目标路径输入 + 后端执行回执展示。
 */
export function UploadSection({ hostId, hostName: name, hostIp: ip, onClose }) {
  const [file, setFile] = useState(null)
  const [targetPath, setTargetPath] = useState('/tmp')
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [result, setResult] = useState(null)

  const handleFileChange = (e) => {
    const selected = e.target.files?.[0]
    if (selected) {
      setFile(selected)
      setError('')
      setBlocked('')
      setResult(null)
    }
  }

  const handleUpload = async () => {
    if (!file) {
      setError('请选择要上传的文件')
      return
    }
    if (!targetPath.trim()) {
      setError('请输入目标路径')
      return
    }

    setUploading(true)
    setError('')
    setBlocked('')
    setResult(null)

    try {
      const formData = new FormData()
      formData.append('file', file)
      formData.append('target_path', targetPath.trim())

      const res = await cmdbApi.upload(hostId, formData)
      setResult(res)
    } catch (err) {
      if (isCmdbContractBlocked(err)) {
        setBlocked(cmdbContractMessage(err, CMDB_EXECUTION_BLOCKERS.upload))
      } else {
        setError(err?.message || '上传失败')
      }
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className='fx-assets-form'>
      <h3 style={{ margin: '0 0 12px', fontSize: 16, color: '#193a63' }}>
        文件上传 - {name} ({ip})
      </h3>

      <Blocked>{CMDB_EXECUTION_BLOCKERS.upload}</Blocked>

      <Field label='选择文件'>
        <input type='file' onChange={handleFileChange} disabled={uploading} />
        {file && <span className='fx-assets-muted' style={{ marginTop: 4, display: 'block' }}>{file.name} ({(file.size / 1024).toFixed(1)} KB)</span>}
      </Field>

      <Field label='目标路径'>
        <input value={targetPath} onChange={e => setTargetPath(e.target.value)} placeholder='/tmp' disabled={uploading} />
      </Field>

      {uploading && <div className='fx-assets-muted'>正在等待后端执行回执，不生成本地百分比。</div>}

      <ErrorBox>{error}</ErrorBox>
      {blocked && <Blocked>{blocked}</Blocked>}

      {result && (
        <div className='fx-assets-table' style={{ marginTop: 8 }}>
          <table><tbody>
            <tr><th>文件名</th><td>{result.filename || file?.name || '-'}</td></tr>
            <tr><th>大小</th><td>{result.size ?? file?.size ?? '-'} bytes</td></tr>
            <tr><th>目标路径</th><td>{result.target_path || targetPath}</td></tr>
            <tr><th>执行回执时间</th><td>{result.uploaded_at || result.executed_at || '-'}</td></tr>
          </tbody></table>
        </div>
      )}

      <footer>
        <button type='button' disabled={uploading || !file} onClick={handleUpload}>
          {uploading ? '提交中...' : '提交上传'}
        </button>
        {onClose && <button type='button' onClick={onClose} style={{ marginLeft: 8 }}>关闭</button>}
      </footer>
    </div>
  )
}
