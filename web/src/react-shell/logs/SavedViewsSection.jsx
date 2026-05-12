import React, { useEffect, useState } from 'react'
import { formatLogError, LOG_BLOCKERS, logsApi } from '../api/logs.js'
import { savedViewColumns } from './logsModel.js'
import { Blocked, Empty } from './LogsShared.jsx'

export function SavedViewsSection() {
  const [blocked, setBlocked] = useState('')
  const [views, setViews] = useState([])

  const loadViews = () => logsApi.views.list('logs').then(items => {
    setViews(items)
    setBlocked('')
  }).catch(err => setBlocked(formatLogError(err)))

  useEffect(() => { loadViews() }, [])

  const saveView = async () => {
    try {
      await logsApi.views.save({
        name: '日志检索视图',
        sourcePage: 'logs',
        query: { q: 'severity_text = "ERROR"' },
        columns: ['timestamp', 'severity_text', 'body'],
        filters: {},
        timeRange: { from: 'now-1h', to: 'now' },
      })
      await loadViews()
    } catch (err) {
      setBlocked(formatLogError(err))
    }
  }

  const updateView = async (view) => {
    try {
      await logsApi.views.update(view.id, { ...view, name: view.name })
      await loadViews()
    } catch (err) {
      setBlocked(formatLogError(err))
    }
  }

  const removeView = async (view) => {
    try {
      await logsApi.views.remove(view.id)
      await loadViews()
    } catch (err) {
      setBlocked(formatLogError(err))
    }
  }

  return (
    <section className='fx-logs-work'>
      <div className='fx-logs-toolbar'><button type='button' onClick={saveView}>保存当前视图</button><button type='button' onClick={loadViews}>刷新</button></div>
      {blocked && <Blocked>{blocked}</Blocked>}
      <div className='fx-logs-table'>
        <table><thead><tr>{savedViewColumns.map(item => <th key={item}>{item}</th>)}</tr></thead><tbody>{views.length ? views.map(view => <tr key={view.id}><td>{view.name}</td><td>{view.source_page || view.sourcePage || 'logs'}</td><td>{view.query?.q || '-'}</td><td>{Array.isArray(view.columns) ? view.columns.join(', ') : '-'}</td><td>{view.updated_at || view.updatedAt || '-'}</td><td><button type='button' onClick={() => updateView(view)}>更新</button><button type='button' onClick={() => removeView(view)}>删除</button></td></tr>) : <tr><td colSpan={savedViewColumns.length}><Empty>{blocked ? '保存视图契约未完成。' : '暂无保存视图。'}</Empty></td></tr>}</tbody></table>
      </div>
      {!views.length && blocked && <Blocked>{LOG_BLOCKERS.savedViews}</Blocked>}
    </section>
  )
}
