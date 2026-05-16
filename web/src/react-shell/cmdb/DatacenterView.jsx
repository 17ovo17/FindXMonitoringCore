import React, { useEffect, useState, useCallback } from 'react'
import { get } from '../api/http.js'
import { displayText } from './assetsModel.js'
import { ErrorBox } from './Shared.jsx'

export function DatacenterView() {
  const [datacenters, setDatacenters] = useState([])
  const [selectedDc, setSelectedDc] = useState(null)
  const [racks, setRacks] = useState([])
  const [selectedRack, setSelectedRack] = useState(null)
  const [units, setUnits] = useState([])
  const [loading, setLoading] = useState(true)
  const [rackLoading, setRackLoading] = useState(false)
  const [unitLoading, setUnitLoading] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    let alive = true
    setLoading(true)
    get('/cmdb/datacenters')
      .then(resp => {
        if (!alive) return
        const list = Array.isArray(resp?.data || resp) ? (resp?.data || resp) : []
        setDatacenters(list)
        if (list.length) setSelectedDc(list[0])
        setLoading(false)
      })
      .catch(err => {
        if (alive) { setError(err.message); setLoading(false) }
      })
    return () => { alive = false }
  }, [])

  useEffect(() => {
    if (!selectedDc) return
    let alive = true
    setRackLoading(true)
    setSelectedRack(null)
    setUnits([])
    get(`/cmdb/datacenters/${encodeURIComponent(selectedDc.id)}/racks`)
      .then(resp => {
        if (alive) {
          setRacks(Array.isArray(resp?.data || resp) ? (resp?.data || resp) : [])
          setRackLoading(false)
        }
      })
      .catch(err => {
        if (alive) { setError(err.message); setRackLoading(false) }
      })
    return () => { alive = false }
  }, [selectedDc])

  const handleRackClick = useCallback((rack) => {
    setSelectedRack(rack)
    setUnitLoading(true)
    get(`/cmdb/datacenters/${encodeURIComponent(selectedDc.id)}/racks/${encodeURIComponent(rack.id)}/units`)
      .then(resp => {
        setUnits(Array.isArray(resp?.data || resp) ? (resp?.data || resp) : [])
        setUnitLoading(false)
      })
      .catch(err => {
        setError(err.message)
        setUnitLoading(false)
      })
  }, [selectedDc])

  if (loading) return <div className='fx-assets-empty'>正在加载机房数据...</div>
  if (error && !datacenters.length) return <ErrorBox>{error}</ErrorBox>

  return (
    <section className='fx-datacenter-view'>
      <ErrorBox>{error}</ErrorBox>
      <div className='fx-datacenter-layout'>
        {/* 左侧: 机房列表 */}
        <aside className='fx-datacenter-sidebar'>
          <h3 style={{ margin: '0 0 12px', fontSize: 14, color: '#193a63' }}>机房列表</h3>
          {datacenters.map(dc => (
            <button
              key={dc.id}
              type='button'
              className={`fx-datacenter-card ${selectedDc?.id === dc.id ? 'is-active' : ''}`}
              onClick={() => setSelectedDc(dc)}
            >
              <strong>{displayText(dc.name)}</strong>
              <span>{displayText(dc.location || dc.address, '-')}</span>
              <span>{displayText(dc.rack_count ? `${dc.rack_count} 机柜` : '', '')}</span>
            </button>
          ))}
          {!datacenters.length && <div className='fx-assets-empty'>暂无机房数据</div>}
        </aside>

        {/* 右侧: 机柜网格 + U位详情 */}
        <div className='fx-datacenter-main'>
          {selectedDc && (
            <>
              <h3 style={{ margin: '0 0 12px', fontSize: 14, color: '#193a63' }}>
                {displayText(selectedDc.name)} - 机柜视图
                {rackLoading && <span style={{ marginLeft: 8, fontSize: 12, color: '#60728e' }}>加载中...</span>}
              </h3>
              <div className='fx-datacenter-rack-grid'>
                {racks.map(rack => {
                  const used = rack.used_units || 0
                  const total = rack.total_units || 42
                  const percent = total > 0 ? Math.round((used / total) * 100) : 0
                  return (
                    <button
                      key={rack.id}
                      type='button'
                      className={`fx-datacenter-rack ${selectedRack?.id === rack.id ? 'is-active' : ''}`}
                      onClick={() => handleRackClick(rack)}
                    >
                      <strong>{displayText(rack.name)}</strong>
                      <span>{used}/{total} U</span>
                      <div className='fx-datacenter-rack-bar'>
                        <i style={{ width: `${percent}%`, background: percent > 80 ? '#dc2626' : percent > 60 ? '#f59e0b' : '#10b981' }} />
                      </div>
                      <span style={{ fontSize: 11, color: '#60728e' }}>{percent}%</span>
                    </button>
                  )
                })}
                {!racks.length && !rackLoading && <div className='fx-assets-empty'>暂无机柜数据</div>}
              </div>

              {selectedRack && (
                <div style={{ marginTop: 16 }}>
                  <h4 style={{ margin: '0 0 8px', fontSize: 13, color: '#193a63' }}>
                    {displayText(selectedRack.name)} - U位布局
                    {unitLoading && <span style={{ marginLeft: 8, fontSize: 12, color: '#60728e' }}>加载中...</span>}
                  </h4>
                  <RackUnitStrip units={units} totalUnits={selectedRack.total_units || 42} />
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </section>
  )
}

function RackUnitStrip({ units, totalUnits }) {
  const occupiedMap = {}
  for (const unit of units) {
    const start = unit.start_u || unit.position || 1
    const size = unit.size || unit.height || 1
    for (let u = start; u < start + size; u++) {
      occupiedMap[u] = unit
    }
  }

  const rows = []
  for (let u = totalUnits; u >= 1; u--) {
    const occupied = occupiedMap[u]
    rows.push(
      <div
        key={u}
        className={`fx-rack-unit ${occupied ? 'is-occupied' : ''}`}
        title={occupied ? `U${u}: ${displayText(occupied.device_name || occupied.name, '已占用')}` : `U${u}: 空闲`}
      >
        <span className='fx-rack-unit-num'>{u}</span>
        {occupied && <span className='fx-rack-unit-device'>{displayText(occupied.device_name || occupied.name, '设备')}</span>}
      </div>
    )
  }

  return (
    <div className='fx-rack-strip'>
      {rows}
    </div>
  )
}
