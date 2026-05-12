import React, { useMemo, useState } from 'react'
import { displayText, durationMs } from './tracingModel.js'
import { Empty, Status } from './TracingShared.jsx'

// Build span tree. Each span has spanId and parentSpanId; segments link via refs.
// We attempt a best-effort tree: parent lookup by (segmentId, parentSpanId) falling back to refs[0].parentSegmentId/parentSpanId.
function buildTree(spans) {
  const map = new Map()
  const normalized = spans.map((s, idx) => ({
    __idx: idx,
    spanId: s.spanId ?? s.id ?? idx,
    parentSpanId: s.parentSpanId ?? -1,
    segmentId: s.segmentId || '',
    parentSegmentId: s.refs?.[0]?.parentSegmentId || s.parentSegmentId || '',
    refParentSpanId: s.refs?.[0]?.parentSpanId,
    endpointName: s.endpointName || s.operationName || '',
    serviceCode: s.serviceCode || s.service || '',
    serviceInstanceName: s.serviceInstanceName || s.instance || '',
    type: s.type || (s.isEntry ? 'Entry' : s.isExit ? 'Exit' : 'Local'),
    isError: Boolean(s.isError),
    startTime: Number(s.startTime) || 0,
    endTime: Number(s.endTime) || 0,
    duration: durationMs(s),
    raw: s,
  }))
  normalized.forEach(n => { map.set(`${n.segmentId}#${n.spanId}`, { ...n, children: [] }) })
  const roots = []
  map.forEach(node => {
    let parentKey = `${node.segmentId}#${node.parentSpanId}`
    if (node.parentSpanId === -1 && node.parentSegmentId) {
      parentKey = `${node.parentSegmentId}#${node.refParentSpanId ?? 0}`
    }
    const parent = map.get(parentKey)
    if (parent && parent !== node) parent.children.push(node)
    else roots.push(node)
  })
  // Sort siblings by start time
  const sortRec = (list) => { list.sort((a, b) => a.startTime - b.startTime || a.__idx - b.__idx); list.forEach(n => sortRec(n.children)) }
  sortRec(roots)
  return roots
}

function flatten(roots, collapsed, level = 0, out = []) {
  roots.forEach(node => {
    const key = `${node.segmentId}#${node.spanId}`
    const hasChildren = node.children.length > 0
    const isCollapsed = collapsed.has(key)
    out.push({ ...node, level, hasChildren, isCollapsed, key })
    if (hasChildren && !isCollapsed) flatten(node.children, collapsed, level + 1, out)
  })
  return out
}

export function TraceWaterfall({ spans }) {
  const [collapsed, setCollapsed] = useState(() => new Set())
  const tree = useMemo(() => buildTree(spans || []), [spans])
  const flat = useMemo(() => flatten(tree, collapsed), [tree, collapsed])

  const timeRange = useMemo(() => {
    if (!spans.length) return { start: 0, end: 1 }
    let start = Infinity, end = 0
    spans.forEach(s => {
      const st = Number(s.startTime) || 0
      const et = Number(s.endTime) || st + (Number(s.duration) || 0)
      if (st && st < start) start = st
      if (et > end) end = et
    })
    if (!isFinite(start)) start = 0
    if (end <= start) end = start + 1
    return { start, end }
  }, [spans])

  const total = Math.max(1, timeRange.end - timeRange.start)
  const toggle = (key) => {
    setCollapsed(prev => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key); else next.add(key)
      return next
    })
  }

  if (!spans || !spans.length) return <Empty>暂无 Span 数据</Empty>

  return (
    <div className='fx-tracing-waterfall'>
      <div className='fx-tracing-waterfall__header'>
        <span>端点 / 方法</span>
        <span>服务 · 实例</span>
        <span style={{ textAlign: 'right' }}>耗时</span>
        <span style={{ textAlign: 'center' }}>状态</span>
        <span>时间轴</span>
      </div>
      {flat.map(node => {
        const st = node.startTime || timeRange.start
        const offset = Math.max(0, ((st - timeRange.start) / total) * 100)
        const width = Math.max(0.5, (node.duration / total) * 100)
        const indent = node.level * 14
        return (
          <div key={node.key} className={`fx-tracing-waterfall__row ${node.isError ? 'is-error' : ''}`}>
            <div className='fx-tracing-waterfall__name' style={{ paddingLeft: indent }}>
              <button type='button' onClick={() => node.hasChildren && toggle(node.key)} disabled={!node.hasChildren} aria-label={node.hasChildren ? (node.isCollapsed ? '展开' : '折叠') : ''}>
                {node.hasChildren ? (node.isCollapsed ? '+' : '−') : '·'}
              </button>
              <span className={`fx-span-type ${node.type === 'Entry' ? 'is-entry' : node.type === 'Exit' ? 'is-exit' : ''}`} title={node.type}>{node.type?.slice(0, 5) || 'Span'}</span>
              <span title={node.endpointName}>{displayText(node.endpointName)}</span>
            </div>
            <div className='fx-tracing-waterfall__service' title={`${node.serviceCode} / ${node.serviceInstanceName}`}>
              {displayText(node.serviceCode)}
              {node.serviceInstanceName ? ` · ${node.serviceInstanceName}` : ''}
            </div>
            <div className='fx-tracing-waterfall__duration'>{node.duration} ms</div>
            <div className='fx-tracing-waterfall__status'>
              <Status ok={!node.isError}>{node.isError ? '错误' : '正常'}</Status>
            </div>
            <div className={`fx-tracing-waterfall__bar ${node.isError ? 'is-error' : ''}`} title={`${node.duration} ms @ +${Math.round(st - timeRange.start)} ms`}>
              <div className='fx-tracing-waterfall__bar-fill' style={{ left: `${offset}%`, width: `${width}%` }} />
            </div>
          </div>
        )
      })}
    </div>
  )
}
