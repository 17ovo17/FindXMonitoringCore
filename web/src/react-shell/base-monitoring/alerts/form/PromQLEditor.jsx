import React, { useState, useEffect, useRef, useCallback } from 'react'

/**
 * 告警规则专用 PromQL 编辑器
 * - 基于 shared/PromQLEditor 的 Monaco 编辑器
 * - 增加即时预览图表（debounce 500ms 查询并显示折线图）
 * - 支持指标名、标签名、标签值自动补全
 */

const PREVIEW_DEBOUNCE = 500

function MiniChart({ data }) {
  const canvasRef = useRef(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !data || data.length === 0) return

    const ctx = canvas.getContext('2d')
    const dpr = window.devicePixelRatio || 1
    const rect = canvas.getBoundingClientRect()
    canvas.width = rect.width * dpr
    canvas.height = rect.height * dpr
    ctx.scale(dpr, dpr)
    ctx.clearRect(0, 0, rect.width, rect.height)

    const COLORS = ['#1769ff', '#e6550d', '#31a354', '#756bb1', '#636363']
    const padding = { top: 8, right: 8, bottom: 20, left: 40 }
    const w = rect.width - padding.left - padding.right
    const h = rect.height - padding.top - padding.bottom

    // 计算全局范围
    let xMin = Infinity, xMax = -Infinity, yMin = Infinity, yMax = -Infinity
    data.forEach((series) => {
      series.values.forEach(([t, v]) => {
        const ts = Number(t)
        const val = Number(v)
        if (ts < xMin) xMin = ts
        if (ts > xMax) xMax = ts
        if (val < yMin) yMin = val
        if (val > yMax) yMax = val
      })
    })
    if (yMin === yMax) { yMin -= 1; yMax += 1 }
    if (xMin === xMax) { xMin -= 60; xMax += 60 }

    // 绘制网格
    ctx.strokeStyle = 'var(--fx-border, #e5e7eb)'
    ctx.lineWidth = 0.5
    for (let i = 0; i <= 4; i++) {
      const y = padding.top + (h * i) / 4
      ctx.beginPath()
      ctx.moveTo(padding.left, y)
      ctx.lineTo(padding.left + w, y)
      ctx.stroke()
    }

    // 绘制数据线
    data.forEach((series, si) => {
      const color = COLORS[si % COLORS.length]
      ctx.strokeStyle = color
      ctx.lineWidth = 1.5
      ctx.beginPath()
      let started = false
      series.values.forEach(([t, v]) => {
        const x = padding.left + ((Number(t) - xMin) / (xMax - xMin)) * w
        const y = padding.top + h - ((Number(v) - yMin) / (yMax - yMin)) * h
        if (!started) { ctx.moveTo(x, y); started = true }
        else ctx.lineTo(x, y)
      })
      ctx.stroke()
    })
  }, [data])

  return (
    <canvas
      ref={canvasRef}
      style={{ width: '100%', height: '100%', display: 'block' }}
    />
  )
}

function FallbackEditor({ value, onChange, placeholder }) {
  return (
    <textarea
      className="fx-promql-fallback"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      rows={4}
      style={{
        width: '100%',
        fontFamily: 'monospace',
        fontSize: 13,
        padding: '8px 12px',
        border: '1px solid var(--fx-border, #e5e7eb)',
        borderRadius: 6,
        background: 'var(--fx-bg, #fff)',
        color: 'var(--fx-ink, #1d2b45)',
        resize: 'vertical',
      }}
    />
  )
}

export function AlertPromQLEditor({
  value = '',
  onChange,
  datasourceId,
  height = 120,
  showPreview = true,
  placeholder = '输入 PromQL 表达式...',
}) {
  const [MonacoEditor, setMonacoEditor] = useState(null)
  const [previewData, setPreviewData] = useState(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewError, setPreviewError] = useState(null)
  const debounceRef = useRef(null)

  // 动态加载 shared PromQLEditor
  useEffect(() => {
    import('../../../shared/PromQLEditor.jsx')
      .then((mod) => setMonacoEditor(() => mod.default))
      .catch(() => setMonacoEditor(null))
  }, [])

  // debounce 预览查询
  const queryPreview = useCallback((expr) => {
    if (debounceRef.current) clearTimeout(debounceRef.current)
    if (!expr || !expr.trim()) {
      setPreviewData(null)
      setPreviewError(null)
      return
    }
    debounceRef.current = setTimeout(async () => {
      setPreviewLoading(true)
      setPreviewError(null)
      try {
        const end = Math.floor(Date.now() / 1000)
        const start = end - 3600
        const step = 15
        const params = new URLSearchParams({
          query: expr.trim(),
          start: String(start),
          end: String(end),
          step: String(step),
        })
        if (datasourceId) params.set('datasource_id', datasourceId)
        const resp = await fetch(`/api/v1/prometheus/query_range?${params}`)
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
        const json = await resp.json()
        const result = json.data?.result || json.dat?.result || []
        const series = result.map((r) => ({
          metric: r.metric || {},
          values: (r.values || []).map(([t, v]) => [Number(t), Number(v)]),
        }))
        setPreviewData(series)
      } catch (err) {
        setPreviewError(err.message || '查询失败')
        setPreviewData(null)
      } finally {
        setPreviewLoading(false)
      }
    }, PREVIEW_DEBOUNCE)
  }, [datasourceId])

  const handleChange = useCallback((val) => {
    onChange?.(val)
    if (showPreview) queryPreview(val)
  }, [onChange, showPreview, queryPreview])

  useEffect(() => {
    if (showPreview && value) queryPreview(value)
    return () => { if (debounceRef.current) clearTimeout(debounceRef.current) }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="fx-alert-promql-editor">
      <div className="fx-alert-promql-editor__input">
        {MonacoEditor ? (
          <MonacoEditor value={value} onChange={handleChange} height={height} placeholder={placeholder} />
        ) : (
          <FallbackEditor value={value} onChange={handleChange} placeholder={placeholder} />
        )}
      </div>
      {showPreview && (
        <div className="fx-alert-promql-editor__preview" style={{ height: 100, marginTop: 8, border: '1px solid var(--fx-border, #e5e7eb)', borderRadius: 6, overflow: 'hidden', position: 'relative' }}>
          {previewLoading && (
            <div style={{ position: 'absolute', top: 4, right: 8, fontSize: 11, color: '#9ca3af' }}>查询中...</div>
          )}
          {previewError && (
            <div style={{ padding: 8, fontSize: 11, color: '#dc2626' }}>{previewError}</div>
          )}
          {previewData && previewData.length > 0 && (
            <MiniChart data={previewData} />
          )}
          {previewData && previewData.length === 0 && !previewLoading && !previewError && (
            <div style={{ padding: 8, fontSize: 11, color: '#9ca3af', textAlign: 'center' }}>无数据</div>
          )}
          {!previewData && !previewLoading && !previewError && (
            <div style={{ padding: 8, fontSize: 11, color: '#9ca3af', textAlign: 'center' }}>输入表达式后自动预览</div>
          )}
        </div>
      )}
    </div>
  )
}
