import React from 'react'

/**
 * 骨架屏组件
 * 支持：line / circle / rect 三种形状
 * 动画：shimmer 效果
 */

export function Skeleton({ type = 'line', width, height, count = 1, className = '' }) {
  const baseClass = `fx-skeleton fx-skeleton--${type} ${className}`

  const style = {}
  if (width) style.width = typeof width === 'number' ? `${width}px` : width
  if (height) style.height = typeof height === 'number' ? `${height}px` : height

  if (type === 'circle') {
    const size = height || width || 40
    style.width = typeof size === 'number' ? `${size}px` : size
    style.height = style.width
  }

  if (count > 1) {
    return (
      <div className="fx-skeleton-group">
        {Array.from({ length: count }, (_, i) => (
          <div key={i} className={baseClass} style={style} aria-hidden="true" />
        ))}
      </div>
    )
  }

  return <div className={baseClass} style={style} aria-hidden="true" />
}

export function SkeletonCard({ lines = 3 }) {
  return (
    <div className="fx-skeleton-card">
      <Skeleton type="rect" width="100%" height={20} />
      <Skeleton type="line" count={lines} />
    </div>
  )
}
