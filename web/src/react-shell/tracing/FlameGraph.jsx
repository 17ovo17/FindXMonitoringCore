import React, { useCallback, useMemo, useRef, useState } from 'react'

const ROW_HEIGHT = 20
const MIN_WIDTH_PX = 2
const COLORS = [
  '#fde68a', '#fcd34d', '#fbbf24', '#f59e0b',
  '#f97316', '#ea580c', '#dc2626', '#b91c1c',
]

function colorByDepth(depth) {
  return COLORS[Math.min(depth, COLORS.length - 1)]
}

/**
 * 将树形火焰图数据展平为可渲染的矩形列表
 * @param {object} root - {name, value, children: [...]}
 * @param {number} totalValue - 根节点的 value（用于计算宽度比例）
 * @param {number} svgWidth - SVG 总宽度
 * @returns {Array} [{name, value, x, y, width, depth, percentage}]
 */
function flattenTree(root, totalValue, svgWidth) {
  const rects = []
  const queue = [{ node: root, x: 0, depth: 0 }]
  while (queue.length > 0) {
    const { node, x, depth } = queue.shift()
    const width = (node.value / totalValue) * svgWidth
    if (width >= MIN_WIDTH_PX) {
      rects.push({
        name: node.name || '(unknown)',
        value: node.value,
        x,
        y: depth * ROW_HEIGHT,
        width,
        depth,
        percentage: ((node.value / totalValue) * 100).toFixed(1),
      })
    }
    if (Array.isArray(node.children)) {
      let childX = x
      for (const child of node.children) {
        queue.push({ node: child, x: childX, depth: depth + 1 })
        childX += (child.value / totalValue) * svgWidth
      }
    }
  }
  return rects
}

function maxDepth(node, d = 0) {
  if (!node?.children?.length) return d
  return Math.max(...node.children.map(c => maxDepth(c, d + 1)))
}

/**
 * SVG 火焰图组件
 * @param {{data: {name, value, children}, width?: number}} props
 */
export function FlameGraph({ data, width = 800 }) {
  const containerRef = useRef(null)
  const [tooltip, setTooltip] = useState(null)
  const [zoomRoot, setZoomRoot] = useState(null)

  const activeData = zoomRoot || data

  const rects = useMemo(() => {
    if (!activeData || !activeData.value) return []
    return flattenTree(activeData, activeData.value, width)
  }, [activeData, width])

  const depth = useMemo(() => activeData ? maxDepth(activeData) : 0, [activeData])
  const svgHeight = (depth + 1) * ROW_HEIGHT + 4

  const handleMouseEnter = useCallback((rect, e) => {
    const containerRect = containerRef.current?.getBoundingClientRect()
    const x = e.clientX - (containerRect?.left || 0)
    const y = e.clientY - (containerRect?.top || 0)
    setTooltip({ ...rect, px: x, py: y })
  }, [])

  const handleMouseLeave = useCallback(() => setTooltip(null), [])

  const handleClick = useCallback((rect) => {
    if (!data) return
    const findNode = (node, name, value) => {
      if (node.name === name && node.value === value) return node
      if (node.children) {
        for (const child of node.children) {
          const found = findNode(child, name, value)
          if (found) return found
        }
      }
      return null
    }
    const target = findNode(data, rect.name, rect.value)
    if (target && target !== data) setZoomRoot(target)
  }, [data])

  const handleReset = useCallback(() => setZoomRoot(null), [])

  if (!data || !data.value) {
    return <div className='fx-flame-empty'>暂无火焰图数据</div>
  }

  return (
    <div ref={containerRef} className='fx-flame-container' style={{ position: 'relative', overflow: 'auto' }}>
      {zoomRoot && (
        <button type='button' className='fx-flame-reset' onClick={handleReset}>
          返回全局视图
        </button>
      )}
      <svg width={width} height={svgHeight} className='fx-flame-svg'>
        {rects.map((rect, i) => (
          <g key={i}
            onMouseEnter={e => handleMouseEnter(rect, e)}
            onMouseLeave={handleMouseLeave}
            onClick={() => handleClick(rect)}
            style={{ cursor: 'pointer' }}
          >
            <rect
              x={rect.x}
              y={rect.y}
              width={rect.width}
              height={ROW_HEIGHT - 2}
              fill={colorByDepth(rect.depth)}
              stroke='#fff'
              strokeWidth={0.5}
              rx={2}
            />
            {rect.width > 40 && (
              <text
                x={rect.x + 4}
                y={rect.y + ROW_HEIGHT - 6}
                fontSize={11}
                fill='#1a1a1a'
                style={{ pointerEvents: 'none', userSelect: 'none' }}
              >
                {rect.name.length > Math.floor(rect.width / 7) ? rect.name.slice(0, Math.floor(rect.width / 7)) + '...' : rect.name}
              </text>
            )}
          </g>
        ))}
      </svg>
      {tooltip && (
        <div className='fx-flame-tooltip' style={{ left: tooltip.px + 12, top: tooltip.py - 40 }}>
          <strong>{tooltip.name}</strong>
          <span>{tooltip.value.toLocaleString()} ({tooltip.percentage}%)</span>
        </div>
      )}
    </div>
  )
}
