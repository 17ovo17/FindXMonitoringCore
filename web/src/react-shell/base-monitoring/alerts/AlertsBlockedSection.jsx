import React from 'react'
import { PENDING, blockedContracts, displayJson } from './alertModel.js'

const blockedRows = {
  job: [
    ['自愈模板', '触发告警', '执行目标', '审批状态', '回滚策略', '最近执行', '操作'],
    ['template search', 'alert binding', 'target selector', 'approval', 'rollback', 'execution records', 'clone/delete'],
  ],
  'tracing-alarms': [
    ['告警名称', '服务', '实例/端点', 'Trace 关联', '拓扑节点', '状态', '最近触发', '操作'],
    ['tracing rule', 'service selector', 'endpoint binding', 'trace backlink', 'topology node', 'mute/ack/recover', 'review evidence'],
  ],
  mutes: [
    ['规则名称', '屏蔽对象', '数据源', '标签条件', '时间窗口', '状态', '更新人', '操作'],
    ['search', 'datasource filter', 'tag matcher', 'periodic window', 'enable switch', 'clone/delete'],
  ],
  subscriptions: [
    ['订阅名称', '接收对象', '级别', '规则范围', '标签条件', '通知策略', '状态', '操作'],
    ['search', 'datasource filter', 'severity filter', 'recipient groups', 'clone/delete'],
  ],
  'event-pipelines': [
    ['流水线名称', '用途', '触发方式', '状态', '团队', '更新人', '更新时间', '操作'],
    ['search', 'use case filter', 'trigger mode filter', 'execution records', 'clone/delete'],
  ],
}

export function AlertsBlockedSection({ section }) {
  const rows = blockedRows[section] || blockedRows.mutes
  const contract = blockedContracts[section]
  return (
    <section className='fx-alert-blocked'>
      <header>
        <strong>{PENDING}</strong>
        <p>{contract}</p>
      </header>
      <div className='fx-alert-filterbar'>
        <input disabled placeholder='搜索' />
        <select disabled><option>数据源</option></select>
        <select disabled><option>状态</option></select>
        <button type='button' disabled>新建</button>
      </div>
      <div className='fx-alert-table'>
        <table>
          <thead><tr>{rows[0].map((item) => <th key={item}>{item}</th>)}</tr></thead>
          <tbody>
            <tr>{rows[0].map((item) => <td key={item}>{PENDING}</td>)}</tr>
          </tbody>
        </table>
      </div>
      <pre>{displayJson({ status: PENDING, section, requiredContract: contract, expectedStructure: rows })}</pre>
    </section>
  )
}
