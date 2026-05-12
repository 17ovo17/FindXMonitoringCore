import React, { useRef, useState } from 'react'
import { alertsApi } from '../../../api/alerts.js'
import { displayJson, makeError } from '../alertModel.js'

/**
 * 批量操作/导入/导出组件
 * 对齐夜莺 MoreOperations + Import + Export
 */

export function BatchRuleActions({ selectedIds, selectedRules, onRefresh, onMessage }) {
  const [importing, setImporting] = useState(false)
  const fileRef = useRef(null)

  const batchEnable = async () => {
    if (!selectedIds.length) { onMessage?.('请先选择规则'); return }
    try {
      await alertsApi.batchEnableRules(selectedIds)
      onRefresh?.()
    } catch (err) {
      onMessage?.(makeError(err, '批量启用失败'))
    }
  }

  const batchDisable = async () => {
    if (!selectedIds.length) { onMessage?.('请先选择规则'); return }
    try {
      await alertsApi.batchDisableRules(selectedIds)
      onRefresh?.()
    } catch (err) {
      onMessage?.(makeError(err, '批量停用失败'))
    }
  }

  const batchDelete = async () => {
    if (!selectedIds.length) { onMessage?.('请先选择规则'); return }
    if (!window.confirm(`确认删除 ${selectedIds.length} 条规则？此操作不可恢复。`)) return
    try {
      await alertsApi.batchDeleteRules(selectedIds)
      onRefresh?.()
    } catch (err) {
      onMessage?.(makeError(err, '批量删除失败'))
    }
  }

  const exportRules = () => {
    if (!selectedRules.length) { onMessage?.('请先选择要导出的规则'); return }
    const data = selectedRules.map((rule) => rule.raw || rule)
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `alert-rules-export-${Date.now()}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  const handleImportFile = async (event) => {
    const file = event.target.files?.[0]
    if (!file) return
    setImporting(true)
    try {
      const text = await file.text()
      const rules = JSON.parse(text)
      if (!Array.isArray(rules)) throw new Error('导入文件必须是 JSON 数组')
      await alertsApi.importRules({ rules })
      onRefresh?.()
      onMessage?.(`成功导入 ${rules.length} 条规则`)
    } catch (err) {
      onMessage?.(makeError(err, '导入失败'))
    } finally {
      setImporting(false)
      if (fileRef.current) fileRef.current.value = ''
    }
  }

  return (
    <div className='fx-alert-batch-actions'>
      <button type='button' onClick={batchEnable} disabled={!selectedIds.length}>批量启用</button>
      <button type='button' onClick={batchDisable} disabled={!selectedIds.length}>批量停用</button>
      <button type='button' className='is-danger' onClick={batchDelete} disabled={!selectedIds.length}>批量删除</button>
      <button type='button' onClick={exportRules} disabled={!selectedRules.length}>导出</button>
      <label className='fx-alert-import-btn'>
        <input ref={fileRef} type='file' accept='.json' onChange={handleImportFile} hidden />
        <button type='button' disabled={importing} onClick={() => fileRef.current?.click()}>
          {importing ? '导入中...' : '导入'}
        </button>
      </label>
    </div>
  )
}
