/**
 * 组件详情抽屉
 * T03: 抽屉 + 5 个 Tab（说明、采集、指标、仪表盘、告警）
 */
import React, { useState } from 'react'
import { InstructionsTab } from './InstructionsTab.jsx'
import { CollectTplsTab } from './CollectTplsTab.jsx'
import { MetricsTab } from './MetricsTab.jsx'
import { DashboardsTab } from './DashboardsTab.jsx'
import { AlertRulesTab } from './AlertRulesTab.jsx'

const TABS = [
  { key: 'instructions', label: '采集说明' },
  { key: 'collect', label: '采集模板' },
  { key: 'metrics', label: '指标说明' },
  { key: 'dashboards', label: '仪表盘模板' },
  { key: 'alertRules', label: '告警规则模板' },
]

const TAB_STORAGE_KEY = 'fx-tpl-drawer-active-tab'

export function ComponentDrawer({ component, query, onClose, onNavigate, onUpdateReadme }) {
  const [activeTab, setActiveTab] = useState(
    () => localStorage.getItem(TAB_STORAGE_KEY) || 'instructions'
  )

  const changeTab = (key) => {
    setActiveTab(key)
    localStorage.setItem(TAB_STORAGE_KEY, key)
  }

  return (
    <div className='fx-tpl-drawer-overlay'>
      <div className='fx-tpl-drawer-backdrop' onClick={onClose} />
      <aside className='fx-tpl-drawer'>
        <header className='fx-tpl-drawer__head'>
          <div className='fx-tpl-drawer__title'>
            {component.logo && <img src={component.logo} alt='' />}
            <h2>{component.ident}</h2>
          </div>
          <button type='button' onClick={onClose} aria-label='关闭'>x</button>
        </header>
        <nav className='fx-tpl-tabs'>
          {TABS.map((tab) => (
            <button
              key={tab.key}
              type='button'
              className={activeTab === tab.key ? 'is-active' : ''}
              onClick={() => changeTab(tab.key)}
            >
              {tab.label}
            </button>
          ))}
        </nav>
        <div className='fx-tpl-drawer__body'>
          {activeTab === 'instructions' && (
            <InstructionsTab component={component} onUpdateReadme={onUpdateReadme} />
          )}
          {activeTab === 'collect' && (
            <CollectTplsTab component={component} />
          )}
          {activeTab === 'metrics' && (
            <MetricsTab component={component} />
          )}
          {activeTab === 'dashboards' && (
            <DashboardsTab component={component} onNavigate={onNavigate} />
          )}
          {activeTab === 'alertRules' && (
            <AlertRulesTab component={component} />
          )}
        </div>
      </aside>
    </div>
  )
}
