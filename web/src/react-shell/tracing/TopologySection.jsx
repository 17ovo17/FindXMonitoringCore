import React, { useCallback, useEffect, useRef, useState } from 'react'
import * as d3 from 'd3'
import { formatTracingError, tracingApi } from '../api/tracing.js'
import { displayText, entityOptions, layerOptions } from './tracingModel.js'
import { AgentEvidenceNotice, AgentLinkActions, Blocked, ErrorBox, Field } from './TracingShared.jsx'

const EMPTY_HINT = '拓扑数据源未接入'

function normalizeGraph(raw) {
  if (!raw) return { nodes: [], edges: [] }
  const nodes = Array.isArray(raw.nodes) ? raw.nodes : []
  const edges = Array.isArray(raw.calls) ? raw.calls : (Array.isArray(raw.edges) ? raw.edges : [])
  return {
    nodes: nodes.map(n => ({
      id: String(n.id || n.key || n.name || ''),
      name: n.name || n.serviceName || n.label || n.id,
      layer: n.layer || n.type || n.entity || '',
      isReal: n.isReal !== false,
      raw: n,
    })),
    edges: edges.map(e => ({
      id: String(e.id || e.key || ((e.source || e.sourceId) + '->' + (e.target || e.targetId))),
      source: String(e.source || e.sourceId || ''),
      target: String(e.target || e.targetId || ''),
      cpm: e.cpm || e.callsPerMin || e.calls,
      latency: e.latency || e.avgResponseTime,
      detectPoint: e.detectPoint || '',
      raw: e,
    })),
  }
}

function LegendSwatch({ color, label }) {
  return <span><span className='fx-dot' style={{ background: color }} />{label}</span>
}

function TopologyGraph({ nodes, edges, onSelectNode, onSelectEdge, onContextMenu }) {
  const svgRef = useRef(null)
  const containerRef = useRef(null)
  const zoomRef = useRef(null)

  const zoomIn = () => { if (zoomRef.current && svgRef.current) d3.select(svgRef.current).transition().duration(300).call(zoomRef.current.scaleBy, 1.3) }
  const zoomOut = () => { if (zoomRef.current && svgRef.current) d3.select(svgRef.current).transition().duration(300).call(zoomRef.current.scaleBy, 0.7) }
  const zoomReset = () => { if (zoomRef.current && svgRef.current) d3.select(svgRef.current).transition().duration(300).call(zoomRef.current.transform, d3.zoomIdentity) }

  useEffect(() => {
    const svg = svgRef.current
    const container = containerRef.current
    if (!svg || !container) return
    const width = container.clientWidth || 600
    const height = container.clientHeight || 520

    const sel = d3.select(svg)
    sel.selectAll('*').remove()
    sel.attr('viewBox', '0 0 ' + width + ' ' + height)

    // Zoom + pan behavior
    const g = sel.append('g')
    const zoom = d3.zoom().scaleExtent([0.2, 5]).on('zoom', (event) => { g.attr('transform', event.transform) })
    sel.call(zoom)
    zoomRef.current = zoom

    // Arrow marker
    const defs = sel.append('defs')
    defs.append('marker')
      .attr('id', 'fx-topo-arrow')
      .attr('viewBox', '0 -5 10 10')
      .attr('refX', 18).attr('refY', 0)
      .attr('markerWidth', 8).attr('markerHeight', 8)
      .attr('orient', 'auto')
      .append('path').attr('d', 'M0,-5L10,0L0,5').attr('fill', '#8ea3c7')

    if (!nodes.length) return

    const simNodes = nodes.map(n => Object.assign({}, n))
    const simEdges = edges.map(e => Object.assign({}, e))

    const sim = d3.forceSimulation(simNodes)
      .force('link', d3.forceLink(simEdges).id(d => d.id).distance(130).strength(0.5))
      .force('charge', d3.forceManyBody().strength(-340))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collide', d3.forceCollide(36))
      .alpha(0.9)

    const linkG = g.append('g').attr('stroke', '#8ea3c7').attr('stroke-opacity', 0.6)
    const link = linkG.selectAll('line')
      .data(simEdges).enter().append('line')
      .attr('stroke-width', d => Math.max(1, Math.min(4, Math.log10((Number(d.cpm) || 1) + 1) * 1.6)))
      .attr('marker-end', 'url(#fx-topo-arrow)')
      .style('cursor', 'pointer')
      .on('click', (_event, d) => onSelectEdge && onSelectEdge(d))

    const linkLabelG = g.append('g').attr('font-size', 10).attr('fill', '#66758d')
    const linkLabel = linkLabelG.selectAll('text')
      .data(simEdges).enter().append('text')
      .attr('text-anchor', 'middle').attr('pointer-events', 'none')
      .text(d => {
        const parts = []
        if (d.cpm !== undefined && d.cpm !== null) parts.push(d.cpm + ' cpm')
        if (d.latency !== undefined && d.latency !== null) parts.push(d.latency + ' ms')
        return parts.join(' · ')
      })

    const nodeG = g.append('g')
    const node = nodeG.selectAll('g')
      .data(simNodes).enter().append('g')
      .style('cursor', 'pointer')
      .on('click', (_event, d) => onSelectNode && onSelectNode(d))
      .on('contextmenu', (event, d) => { event.preventDefault(); onContextMenu && onContextMenu(event, d) })
      .call(d3.drag()
        .on('start', (event, d) => { if (!event.active) sim.alphaTarget(0.3).restart(); d.fx = d.x; d.fy = d.y })
        .on('drag', (event, d) => { d.fx = event.x; d.fy = event.y })
        .on('end', (event, d) => { if (!event.active) sim.alphaTarget(0); d.fx = null; d.fy = null }))

    node.append('circle').attr('r', 20).attr('fill', d => d.isReal ? '#1769ff' : '#bad3ff').attr('stroke', '#fff').attr('stroke-width', 2)
    node.append('text').attr('text-anchor', 'middle').attr('dy', 34).attr('font-size', 11).attr('fill', '#17233c')
      .text(d => (d.name || d.id).length > 18 ? (d.name || d.id).slice(0, 17) + '…' : (d.name || d.id))
    node.append('title').text(d => (d.name || d.id) + ' · ' + (d.layer || ''))

    sim.on('tick', () => {
      link.attr('x1', d => d.source.x).attr('y1', d => d.source.y).attr('x2', d => d.target.x).attr('y2', d => d.target.y)
      linkLabel.attr('x', d => ((d.source.x || 0) + (d.target.x || 0)) / 2).attr('y', d => ((d.source.y || 0) + (d.target.y || 0)) / 2 - 4)
      node.attr('transform', d => 'translate(' + d.x + ',' + d.y + ')')
    })

    return () => { sim.stop() }
  }, [nodes, edges, onSelectNode, onSelectEdge, onContextMenu])

  return (
    <div className='fx-tracing-topology-canvas' ref={containerRef}>
      <svg ref={svgRef} preserveAspectRatio='xMidYMid meet' />
      <div className='fx-tracing-topology-controls'>
        <button type='button' onClick={zoomIn} title='放大'>+</button>
        <button type='button' onClick={zoomOut} title='缩小'>−</button>
        <button type='button' onClick={zoomReset} title='重置'>⟲</button>
      </div>
      <div className='fx-tracing-topology-legend'>
        <LegendSwatch color='#1769ff' label='真实服务节点' />
        <LegendSwatch color='#bad3ff' label='未识别/Conjectural' />
        <div style={{ marginTop: 4 }}>边宽表示 CPM · 箭头指向被调用方 · 滚轮缩放 · 拖拽平移</div>
      </div>
    </div>
  )
}

export function TopologySection({ query, onNavigate }) {
  const q = query || {}
  const [filters, setFilters] = useState({
    layer: q.layer || 'GENERAL',
    entity: q.entity || 'service',
    depth: q.depth || '2',
    serviceId: q.serviceId || '',
  })
  const [graph, setGraph] = useState({ nodes: [], edges: [] })
  const [error, setError] = useState('')
  const [blocked, setBlocked] = useState('')
  const [loading, setLoading] = useState(false)
  const [selectedNode, setSelectedNode] = useState(null)
  const [selectedEdge, setSelectedEdge] = useState(null)
  const [contextMenu, setContextMenu] = useState(null)

  const patch = (key, value) => setFilters(prev => Object.assign({}, prev, { [key]: value }))

  const handleContextMenu = useCallback((event, node) => {
    setContextMenu({ x: event.clientX, y: event.clientY, node })
  }, [])

  const closeContextMenu = useCallback(() => setContextMenu(null), [])

  useEffect(() => {
    const handler = () => closeContextMenu()
    document.addEventListener('click', handler)
    return () => document.removeEventListener('click', handler)
  }, [closeContextMenu])

  const load = useCallback(async () => {
    setLoading(true); setError(''); setBlocked(''); setSelectedNode(null); setSelectedEdge(null)
    try {
      const raw = await tracingApi.topology({
        layer: filters.layer,
        entity: filters.entity,
        depth: filters.depth,
        serviceId: filters.serviceId,
      })
      const normalized = normalizeGraph(raw)
      setGraph(normalized)
      if (!normalized.nodes.length) setBlocked(EMPTY_HINT)
    } catch (err) {
      setGraph({ nodes: [], edges: [] })
      const msg = formatTracingError(err)
      if (msg.startsWith('PENDING') || (err && [404, 405, 501].includes(err.status))) {
        setBlocked(EMPTY_HINT + ' (' + msg + ')')
      } else {
        setError(msg)
        setBlocked(EMPTY_HINT)
      }
    } finally { setLoading(false) }
  }, [filters.layer, filters.entity, filters.depth, filters.serviceId])

  useEffect(() => { load() }, [load])

  const nodeCount = graph.nodes.length
  const edgeCount = graph.edges.length

  return (
    <section className='fx-tracing-work'>
      <div className='fx-tracing-condition-bar'>
        <Field label='层级'>
          <select value={filters.layer} onChange={e => patch('layer', e.target.value)}>
            {layerOptions.map(item => <option key={item}>{item}</option>)}
          </select>
        </Field>
        <Field label='对象'>
          <select value={filters.entity} onChange={e => patch('entity', e.target.value)}>
            {entityOptions.map(item => <option key={item} value={item}>{item}</option>)}
          </select>
        </Field>
        <Field label='深度'><input value={filters.depth} onChange={e => patch('depth', e.target.value)} /></Field>
        <Field label='服务 ID' className='is-flex'>
          <input value={filters.serviceId} onChange={e => patch('serviceId', e.target.value)} placeholder='可选: 以某服务为中心展开' />
        </Field>
        <div className='fx-tracing-condition-actions'>
          <button type='button' className='is-primary' onClick={load}>{loading ? '查询中...' : '查询拓扑'}</button>
          <button type='button' onClick={() => { setFilters({ layer: 'GENERAL', entity: 'service', depth: '2', serviceId: '' }) }}>重置</button>
        </div>
      </div>

      <AgentEvidenceNotice>拓扑节点探针证据以 serviceId 或 serviceInstanceId 为入口; 调用边需要 client/server 双端 Agent 映射, 契约缺失时不展示虚假覆盖率。</AgentEvidenceNotice>
      <AgentLinkActions onNavigate={onNavigate} q={filters.serviceId || q.q} />
      <ErrorBox>{error}</ErrorBox>
      {blocked && <Blocked>{blocked}</Blocked>}

      <div className='fx-tracing-trace-list__head' style={{ borderRadius: 8, marginTop: 12 }}>
        <div><strong>服务拓扑图</strong>节点 {nodeCount} · 调用边 {edgeCount}</div>
        <span>拖拽节点可调整布局 · 点击节点查看详情</span>
      </div>

      {nodeCount > 0 ? (
        <TopologyGraph nodes={graph.nodes} edges={graph.edges} onSelectNode={setSelectedNode} onSelectEdge={setSelectedEdge} onContextMenu={handleContextMenu} />
      ) : (
        <div className='fx-tracing-topology-canvas'>
          <div className='fx-tracing-topology-empty'>{blocked || (loading ? '加载拓扑...' : EMPTY_HINT)}</div>
        </div>
      )}

      {contextMenu && (
        <div className='fx-tracing-context-menu' style={{ left: contextMenu.x, top: contextMenu.y }}>
          <button type='button' onClick={() => { setSelectedNode(contextMenu.node); closeContextMenu() }}>查看详情</button>
          <button type='button' onClick={() => { onNavigate({ section: 'traces', serviceId: contextMenu.node.id }); closeContextMenu() }}>查看 Trace</button>
          <button type='button' onClick={() => { onNavigate({ section: 'services', q: contextMenu.node.name }); closeContextMenu() }}>查看指标</button>
        </div>
      )}

      {selectedEdge && (
        <div className='fx-tracing-edge-panel'>
          <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <h3>调用详情: {displayText(selectedEdge.source?.name || selectedEdge.source)} → {displayText(selectedEdge.target?.name || selectedEdge.target)}</h3>
            <button type='button' onClick={() => setSelectedEdge(null)} style={{ border: '1px solid var(--fx-border)', borderRadius: 6, background: '#fff', padding: '4px 10px', cursor: 'pointer' }}>关闭</button>
          </header>
          <div className='fx-edge-metrics'>
            <div className='fx-edge-metric'><span className='fx-label'>调用次数 (CPM)</span><span className='fx-value'>{selectedEdge.cpm ?? '-'}</span></div>
            <div className='fx-edge-metric'><span className='fx-label'>平均耗时</span><span className='fx-value'>{selectedEdge.latency ? selectedEdge.latency + ' ms' : '-'}</span></div>
            <div className='fx-edge-metric'><span className='fx-label'>检测点</span><span className='fx-value'>{displayText(selectedEdge.detectPoint || '-')}</span></div>
          </div>
        </div>
      )}

      {selectedNode && (
        <div className='fx-tracing-table' style={{ marginTop: 12 }}>
          <h3>节点详情</h3>
          <table>
            <tbody>
              <tr><th style={{ width: 120 }}>名称</th><td>{displayText(selectedNode.name)}</td></tr>
              <tr><th>ID</th><td style={{ fontFamily: 'monospace' }}>{displayText(selectedNode.id)}</td></tr>
              <tr><th>层级</th><td>{displayText(selectedNode.layer)}</td></tr>
              <tr>
                <th>操作</th>
                <td className='fx-tracing-actions'>
                  <button type='button' onClick={() => onNavigate({ section: 'services', q: selectedNode.name })}>在服务目录查看</button>
                  <button type='button' onClick={() => onNavigate({ section: 'traces', serviceId: selectedNode.id })}>Trace 检索</button>
                  <AgentLinkActions onNavigate={onNavigate} q={selectedNode.name || selectedNode.id} />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      )}
    </section>
  )
}
