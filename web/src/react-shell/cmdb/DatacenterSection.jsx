import React, { useEffect, useState, useCallback, useMemo } from 'react'
import { get } from '../api/http.js'
import { displayText } from './assetsModel.js'
import { ErrorBox } from './Shared.jsx'

/**
 * 机房视图页面
 * 布局：左侧机房列表树 + 右侧机柜可视化
 */

const UNIT_HEIGHT = 16
const TOTAL_UNITS = 42
const RACK_HEIGHT = TOTAL_UNITS * UNIT_HEIGHT

const STATUS_COLORS = {
  occupied: '#3b82f6',
  empty: '#e5e7eb',
  alert: '#dc2626',
}

export function DatacenterSection() {
  const [datacenters, setDatacenters] = useState([])
  const [selectedDc, setSelectedDc] = useState(null)
  const [racks, setRacks] = useState([])
  const [selectedRack, setSelectedRack] = useState(null)
  const [devices, setDevices] = useState([])
  const [loading, setLoading] = useState(true)
  const [rackLoading, setRackLoading] = useState(false)
  const [deviceLoading, setDeviceLoading] = useState(false)
  const [error, setError] = useState('')
  const [searchText, setSearchText] = useState('')

  // 加载机房列表
  useEffect(() => {
    let alive = true
    setLoading(true)
    get('/cmdb/datacenters')
      .then(resp => {
        if (!alive) return
        const list = Array.isArray(resp?.items || resp?.data || resp) ? (resp?.items || resp?.data || resp) : []
        setDatacenters(list)
        if (list.length) setSelectedDc(list[0])
        setLoading(false)
      })
      .catch(err => {
        if (alive) { setError(err.message); setLoading(false) }
      })
    return () => { alive = false }
  }, [])
  // 加载机柜列表
  useEffect(() => {
    if (!selectedDc) return
    let alive = true
    setRackLoading(true)
    setSelectedRack(null)
    setDevices([])
    get(`/cmdb/datacenters/${encodeURIComponent(selectedDc.id)}/racks`)
      .then(resp => {
        if (alive) {
          const list = Array.isArray(resp?.items || resp?.data || resp) ? (resp?.items || resp?.data || resp) : []
          setRacks(list)
          setRackLoading(false)
        }
      })
      .catch(err => {
        if (alive) { setError(err.message); setRackLoading(false) }
      })
    return () => { alive = false }
  }, [selectedDc])

  // 点击机柜加载设备
  const handleRackClick = useCallback((rack) => {
    setSelectedRack(rack)
    setDeviceLoading(true)
    get(`/cmdb/racks/${encodeURIComponent(rack.id)}/devices`)
      .then(resp => {
        const list = Array.isArray(resp?.items || resp?.data || resp) ? (resp?.items || resp?.data || resp) : []
        setDevices(list)
        setDeviceLoading(false)
      })
      .catch(err => {
        setError(err.message)
        setDeviceLoading(false)
      })
  }, [])

  // 搜索过滤
  const filteredDcs = useMemo(() => {
    if (!searchText.trim()) return datacenters
    const q = searchText.toLowerCase()
    return datacenters.filter(dc =>
      (dc.name || '').toLowerCase().includes(q) ||
      (dc.location || '').toLowerCase().includes(q)
    )
  }, [datacenters, searchText])

  if (loading) return <div className='fx-assets-empty'>正在加载机房数据...</div>
  if (error && !datacenters.length) return <ErrorBox>{error}</ErrorBox>

  return (
    <section className='fx-datacenter-section'>
      {error && <ErrorBox>{error}</ErrorBox>}
      <div className='fx-datacenter-split'>
        {/* 左侧：机房列表树 */}
        <aside className='fx-datacenter-tree'>
          <div className='fx-datacenter-tree-head'>
            <h3>机房列表</h3>
            <input
              type='text'
              placeholder='搜索机房...'
              value={searchText}
              onChange={e => setSearchText(e.target.value)}
              className='fx-datacenter-search'
            />
          </div>
          <div className='fx-datacenter-tree-list'>
            {filteredDcs.map(dc => (
              <button
                key={dc.id}
                type='button'
                className={`fx-datacenter-tree-item ${selectedDc?.id === dc.id ? 'is-active' : ''}`}
                onClick={() => setSelectedDc(dc)}
              >
                <strong>{displayText(dc.name)}</strong>
                <span className='fx-datacenter-tree-meta'>
                  {displayText(dc.location || dc.address, '')}
                  {dc.rack_count ? ` | ${dc.rack_count} 机柜` : ''}
                </span>
              </button>
            ))}
            {!filteredDcs.length && <div className='fx-assets-empty'>暂无匹配机房</div>}
          </div>
        </aside>

        {/* 右侧：机柜网格 + U位可视化 */}
        <div className='fx-datacenter-content'>
          {selectedDc && (
            <>
              <div className='fx-datacenter-content-head'>
                <h3>{displayText(selectedDc.name)} — 机柜布局</h3>
                {rackLoading && <span className='fx-datacenter-loading'>加载中...</span>}
              </div>
              <RackGrid
                racks={racks}
                selectedRack={selectedRack}
                onRackClick={handleRackClick}
                rackLoading={rackLoading}
              />
              {selectedRack && (
                <RackView
                  rack={selectedRack}
                  devices={devices}
                  loading={deviceLoading}
                />
              )}
            </>
          )}
          {!selectedDc && <div className='fx-assets-empty'>请选择一个机房</div>}
        </div>
      </div>
    </section>
  )
}
/**
 * 机柜网格布局（每行 4-6 个机柜）
 */
function RackGrid({ racks, selectedRack, onRackClick, rackLoading }) {
  if (!racks.length && !rackLoading) {
    return <div className='fx-assets-empty'>暂无机柜数据</div>
  }
  return (
    <div className='fx-datacenter-rack-grid'>
      {racks.map(rack => {
        const used = rack.used_units || 0
        const total = rack.total_units || TOTAL_UNITS
        const percent = total > 0 ? Math.round((used / total) * 100) : 0
        const isActive = selectedRack?.id === rack.id
        return (
          <button
            key={rack.id}
            type='button'
            className={`fx-datacenter-rack-card ${isActive ? 'is-active' : ''}`}
            onClick={() => onRackClick(rack)}
            title={`${rack.name}: ${used}/${total} U 已用 (${percent}%)`}
          >
            <strong>{displayText(rack.name)}</strong>
            <span className='fx-datacenter-rack-usage'>{used}/{total} U</span>
            <div className='fx-datacenter-rack-bar'>
              <i style={{
                width: `${percent}%`,
                background: percent > 80 ? '#dc2626' : percent > 60 ? '#f59e0b' : '#10b981',
              }} />
            </div>
            <span className='fx-datacenter-rack-percent'>{percent}%</span>
          </button>
        )
      })}
    </div>
  )
}

/**
 * 机柜 U 位可视化组件
 * 42 个 U 位格子（从上到下 42→1）
 * 设备占据多个 U 位（如服务器占 2U）
 * 颜色编码：已占用(蓝色)、空闲(灰色)、告警(红色)
 */
function RackView({ rack, devices, loading }) {
  const totalUnits = rack.total_units || TOTAL_UNITS
  const [tooltip, setTooltip] = useState(null)

  // 构建 U 位占用映射
  const occupiedMap = useMemo(() => {
    const map = {}
    for (const device of devices) {
      const startU = device.start_u || device.position || 1
      const size = device.size || device.height || 1
      for (let u = startU; u < startU + size; u++) {
        map[u] = device
      }
    }
    return map
  }, [devices])

  // 构建渲染行（合并连续设备占用）
  const rows = useMemo(() => {
    const result = []
    const rendered = new Set()
    for (let u = totalUnits; u >= 1; u--) {
      if (rendered.has(u)) continue
      const device = occupiedMap[u]
      if (device) {
        const startU = device.start_u || device.position || u
        const size = device.size || device.height || 1
        // 标记所有占用的 U 位
        for (let i = startU; i < startU + size; i++) rendered.add(i)
        const status = device.alert ? 'alert' : 'occupied'
        result.push({ u, device, size, status, isStart: u === startU + size - 1 })
      } else {
        result.push({ u, device: null, size: 1, status: 'empty', isStart: true })
      }
    }
    return result
  }, [occupiedMap, totalUnits])

  if (loading) return <div className='fx-assets-empty'>加载设备数据...</div>

  return (
    <div className='fx-datacenter-rackview'>
      <div className='fx-datacenter-rackview-head'>
        <h4>{displayText(rack.name)} — U位布局 ({totalUnits}U)</h4>
        <div className='fx-datacenter-legend'>
          <span><i style={{ background: STATUS_COLORS.occupied }} />已占用</span>
          <span><i style={{ background: STATUS_COLORS.empty }} />空闲</span>
          <span><i style={{ background: STATUS_COLORS.alert }} />告警</span>
        </div>
      </div>
      <div className='fx-datacenter-rack-body' style={{ height: totalUnits * UNIT_HEIGHT + 2 }}>
        {Array.from({ length: totalUnits }, (_, i) => {
          const u = totalUnits - i
          const device = occupiedMap[u]
          const status = device ? (device.alert ? 'alert' : 'occupied') : 'empty'
          const color = STATUS_COLORS[status]
          return (
            <div
              key={u}
              className={`fx-rack-u ${status}`}
              style={{ height: UNIT_HEIGHT, background: color }}
              onMouseEnter={() => device && setTooltip({ u, device })}
              onMouseLeave={() => setTooltip(null)}
              onClick={() => device?.instance_id && (window.location.hash = `#cmdb/instance-detail?id=${device.instance_id}`)}
            >
              <span className='fx-rack-u-label'>{u}</span>
              {device && <span className='fx-rack-u-name'>{displayText(device.device_name || device.name, '')}</span>}
              {device?.ip && <span className='fx-rack-u-ip'>{device.ip}</span>}
            </div>
          )
        })}
      </div>
      {tooltip && (
        <div className='fx-datacenter-tooltip'>
          <strong>U{tooltip.u}: {displayText(tooltip.device.device_name || tooltip.device.name, '设备')}</strong>
          {tooltip.device.ip && <span>IP: {tooltip.device.ip}</span>}
          {tooltip.device.size > 1 && <span>占用: {tooltip.device.size}U</span>}
          {tooltip.device.model && <span>型号: {tooltip.device.model}</span>}
          {tooltip.device.status && <span>状态: {tooltip.device.status}</span>}
        </div>
      )}
    </div>
  )
}
