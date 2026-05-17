import React from 'react'
import { Blocked, Status, Tags } from './AgentShared.jsx'

const PENDING_STATUS = 'pending'
const fallbackBlocker = `${PENDING_STATUS}: 安装环境契约未返回，不能判定平台自带工具、执行器、服务注册和数据到达闭环。`

const toList = value => Array.isArray(value) ? value.filter(Boolean) : String(value || '').split(/[,，\n]/).map(item => item.trim()).filter(Boolean)
const field = (row, camelKey, snakeKey, fallback = PENDING_STATUS) => row?.[camelKey] || row?.[snakeKey] || fallback

const normalizeMatrixRow = row => ({
  platform: field(row, 'platform', 'platform', '未知平台'),
  method: field(row, 'method', 'install_method', '安装方式缺契约'),
  toolEvidence: field(row, 'toolEvidence', 'tool_evidence', '自带工具证据缺失'),
  sourceState: field(row, 'sourceState', 'source_state', 'LOCAL_SOURCE_MISSING'),
  packageState: field(row, 'packageState', 'package_state', 'PACKAGE_MISSING'),
  executor: field(row, 'executor', 'executor'),
  serviceRegistration: field(row, 'serviceRegistration', 'service_registration'),
  configDelivery: field(row, 'configDelivery', 'config_delivery'),
  uninstall: field(row, 'uninstall', 'uninstall'),
  rollback: field(row, 'rollback', 'rollback'),
  dataArrival: field(row, 'dataArrival', 'data_arrival'),
  blocker: field(row, 'blocker', 'blocker', fallbackBlocker),
})

export const normalizeProbeMatrix = value => {
  const rows = Array.isArray(value) ? value : []
  return rows.length ? rows.map(normalizeMatrixRow) : [normalizeMatrixRow({ blocker: fallbackBlocker })]
}

export function ProbeEnvironmentMatrix({ matrix }) {
  const rows = normalizeProbeMatrix(matrix)
  return (
    <div className='fx-agent-plugin-config'>
      <div className='fx-agent-subtitle'>平台自带探针安装环境矩阵</div>
      <Blocked>命令预览、blocked 审计和下载入口只代表控制面可见，不等于真实安装、服务注册、卸载、回滚或数据到达已完成。</Blocked>
      <div className='fx-agent-table'>
        <table>
          <thead>
            <tr>
              <th>平台</th>
              <th>安装方式</th>
              <th>自带工具证据</th>
              <th>源码/包</th>
              <th>执行与服务</th>
              <th>配置/插件</th>
              <th>卸载/回滚</th>
              <th>数据到达</th>
            </tr>
          </thead>
          <tbody>{rows.map(row => <MatrixRow key={`${row.platform}-${row.method}`} row={row} />)}</tbody>
        </table>
      </div>
      <div className='fx-agent-subtitle'>当前 blocker</div>
      <Tags items={[...new Set(rows.map(row => row.blocker))]} />
    </div>
  )
}

function MatrixRow({ row }) {
  return (
    <tr>
      <td>{row.platform}</td>
      <td>{row.method}</td>
      <td>{row.toolEvidence}</td>
      <td><Tags items={[row.sourceState, row.packageState]} /></td>
      <td><Status ok={false}>{row.executor}</Status><div className='fx-agent-muted'>{row.serviceRegistration}</div></td>
      <td>{row.configDelivery}</td>
      <td>{row.uninstall} / {row.rollback}</td>
      <td>{row.dataArrival}</td>
    </tr>
  )
}

export function PluginDeliveryMatrix({ pluginConfig }) {
  if (!pluginConfig) return null
  return (
    <div className='fx-agent-plugin-config'>
      <div className='fx-agent-subtitle'>采集插件配置下发范围</div>
      <Tags items={toList(pluginConfig.deliveryScope || pluginConfig.delivery_scope)} />
      <div className='fx-agent-plugin-config-row'>
        <span>远程修改 {pluginConfig.remoteMutationStatus || PENDING_STATUS}</span>
        <span>reload {pluginConfig.remoteReloadStatus || PENDING_STATUS}</span>
        <span>漂移检测 {pluginConfig.driftDetectionStatus || PENDING_STATUS}</span>
        <span>回滚 {pluginConfig.rollbackStatus || PENDING_STATUS}</span>
        <span>receipt {pluginConfig.receiptStatus || PENDING_STATUS}</span>
      </div>
      <Blocked>采集插件配置已纳入 FindX Agent、CMDB 主机、业务组和 namespace/workload 可下发范围；远程修改、reload、漂移检测、回滚和 receipt 仍待接入。</Blocked>
    </div>
  )
}
