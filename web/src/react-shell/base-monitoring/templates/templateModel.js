/**
 * 模板中心数据模型
 * 对齐夜莺 builtInComponents 类型定义和工具函数
 */
import { normalizeList, redactText } from '../../api/http.js'

// 类型枚举
export const TypeEnum = {
  alert: 'alert',
  dashboard: 'dashboard',
  collect: 'collect',
  metric: 'metric',
}

// 标准化组件列表
export const normalizeComponents = (value) => {
  const rows = normalizeList(value)
  return rows.map((row) => ({
    id: row.id,
    ident: row.ident || row.name || '',
    logo: row.logo || '',
    readme: row.readme || '',
    disabled: Number(row.disabled) || 0,
  }))
}

// 标准化 Payload 列表
export const normalizePayloads = (value) => {
  const rows = normalizeList(value)
  return rows.map((row) => ({
    id: row.id,
    uuid: row.uuid || '',
    type: row.type || '',
    component_id: row.component_id,
    cate: row.cate || '',
    name: row.name || '',
    tags: row.tags || '',
    note: row.note || '',
    content: row.content || '',
    updated_by: row.updated_by || '',
  }))
}

// 格式化 JSON 内容（美化显示）
export const formatBeautifyJson = (content, mode) => {
  try {
    const parsed = typeof content === 'string' ? JSON.parse(content) : content
    const result = mode === 'array' ? [parsed] : parsed
    return JSON.stringify(result, null, 2)
  } catch {
    return content || ''
  }
}

// 批量格式化
export const formatBeautifyJsons = (contents) => {
  const results = []
  for (const content of contents) {
    try {
      const parsed = typeof content === 'string' ? JSON.parse(content) : content
      results.push(parsed)
    } catch {
      // 跳过无法解析的内容
    }
  }
  return JSON.stringify(results, null, 2)
}
