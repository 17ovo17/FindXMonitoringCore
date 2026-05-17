import React, { useState } from 'react'
import {
  blockedContracts,
  initialLetters,
  isWritablePayloadType,
  payloadTabs,
  sanitizeDisplayText,
  systemDetailJson,
} from './integrationModel.js'

export function Logo({ item }) {
  const [failed, setFailed] = useState(false)
  if (!item?.logo || failed) return <span className='fx-int-logo'>{initialLetters(item?.ident)}</span>
  return (
    <span className='fx-int-logo has-image'>
      <img src={item.logo} alt='' onError={() => setFailed(true)} />
    </span>
  )
}

export function BlockedPanel({ message }) {
  const text = sanitizeDisplayText(message || '')
  return (
    <div className='fx-int-alert is-warning'>
      <strong></strong>
      <span>{text}</span>
    </div>
  )
}

export function SystemsSection(props) {
  const {
    rows = [],
    orderRows = [],
    query = '',
    loading = false,
    saving = false,
    error = '',
    notice = '',
    selected,
    formState,
    orderOpen = false,
    orderError = '',
    onQuery,
    onReload,
    onPreview,
    onCreate,
    onEdit,
    onDelete,
    onToggleMenu,
    onFormDraft,
    onFormSubmit,
    onFormClose,
    onOrderOpen,
    onOrderClose,
    onOrderChange,
    onOrderSubmit,
    onBlocked,
    onClosePreview,
  } = props

  return (
    <main className='fx-int-page'>
      <header className='fx-int-header'>
        <div>
          <p>FindX 集成中心</p>
          <h1>系统集成</h1>
          <span>管理 FindX 内部系统入口、可见范围、菜单显示元数据和排序。嵌入打开与动态菜单嵌入仍待接入。</span>
        </div>
        <div className='fx-int-actions'>
          <input value={query} onChange={(event) => onQuery(event.target.value)} placeholder='搜索名称、团队、配置' />
          <button type='button' onClick={onReload} disabled={loading || saving}>{loading ? '刷新中...' : '刷新'}</button>
          <button type='button' className='is-primary' onClick={onCreate} disabled={saving}>新增集成</button>
        </div>
      </header>
      {error && <BlockedPanel message={error} />}
      {notice && <SystemNotice message={notice} />}
      <section className='fx-int-panel'>
        <header className='fx-int-section-head fx-int-section-head--compact'>
          <div>
            <p>入口目录</p>
            <h1>{rows.length} 个系统集成</h1>
          </div>
          <button type='button' onClick={onOrderOpen} disabled={saving || rows.length === 0}>调整排序</button>
        </header>
        <div className='fx-int-table-wrap'>
          <table className='fx-int-table fx-int-systems-table'>
            <thead>
              <tr>
                <th>名称</th>
                <th>配置预览</th>
                <th>可见性</th>
                <th>团队</th>
                <th>菜单</th>
                <th>更新人</th>
                <th>更新时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr key={row.id}>
                  <td>
                    <button type='button' className='fx-int-link' onClick={() => onPreview(row)}>
                      {row.name}
                    </button>
                  </td>
                  <td><span className='fx-int-config-preview'>{row.configPreview || row.url || '-'}</span></td>
                  <td><span className='fx-int-status'>{row.isPrivate ? '团队可见' : '公开'}</span></td>
                  <td><TagList tags={row.teamIds} onTag={onQuery} /></td>
                  <td>
                    <button type='button' className='fx-int-switch-readonly' onClick={() => onToggleMenu(row)} aria-pressed={row.showInMenu} disabled={saving}>
                      {row.showInMenu ? '显示' : '隐藏'}
                    </button>
                  </td>
                  <td>{row.updateBy || row.createBy || '-'}</td>
                  <td>{row.updateAt || '-'}</td>
                  <td>
                    <span className='fx-int-row-actions'>
                      <button type='button' onClick={() => onPreview(row)}>详情</button>
                      <button type='button' onClick={() => onBlocked('systemOpen')}>打开</button>
                      <button type='button' onClick={() => onEdit(row)} disabled={saving}>编辑</button>
                      <button type='button' className='is-danger' onClick={() => onDelete(row)} disabled={saving || row.builtin}>删除</button>
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {loading && <div className='fx-int-empty'>系统集成加载中...</div>}
          {!loading && rows.length === 0 && <div className='fx-int-empty'>暂无系统集成；后端契约不可用时待接入。</div>}
        </div>
      </section>
      <SystemDetailPanel row={selected} onClose={onClosePreview} onBlocked={onBlocked} onEdit={onEdit} onDelete={onDelete} onToggleMenu={onToggleMenu} saving={saving} />
      <SystemFormModal state={formState} saving={saving} onDraft={onFormDraft} onSubmit={onFormSubmit} onClose={onFormClose} />
      <SystemOrderModal rows={orderRows} open={orderOpen} saving={saving} error={orderError} onChange={onOrderChange} onSubmit={onOrderSubmit} onClose={onOrderClose} />
    </main>
  )
}

function SystemNotice({ message }) {
  const text = sanitizeDisplayText(message || '')
  
  return <div className='fx-int-alert is-success'><span>{text}</span></div>
}

export function ComponentCard({ item, selected, href, saving, onOpen, onEdit, onDelete }) {
  const stats = [
    ['仪表盘', item.dashboardCount],
    ['采集', item.collectCount],
    ['指标', item.metricCount],
    ['告警', item.alertCount],
  ]
  const openCard = () => onOpen(item)

  return (
    <article className={`fx-int-card${selected ? ' is-selected' : ''}`}>
      <a
        className='fx-int-card__open'
        href={href}
        onClick={(event) => {
          event.preventDefault()
          openCard()
        }}
      >
        <span className='fx-int-card__body'>
          <span className='fx-int-card__top'>
            <Logo item={item} />
            <span>
              <strong>{sanitizeDisplayText(item.ident)}</strong>
              <small>{item.contractSource === 'dashboard_fallback' ? '契约缺失降级视图' : 'FindX 内置组件'}</small>
            </span>
          </span>
          <em>{sanitizeDisplayText(item.readme) || '暂无说明'}</em>
          <span className='fx-int-counts' aria-label='组件模板统计'>
            {stats.map(([label, value]) => (
              <b key={label}>
                <span>{value}</span>
                <small>{label}</small>
              </b>
            ))}
          </span>
        </span>
      </a>
      <footer className='fx-int-card__ops'>
        <button type='button' title='编辑' onClick={() => onEdit(item)} disabled={saving || item.contractSource === 'dashboard_fallback' || item.protected}>编辑</button>
        <button type='button' title='删除' className='is-danger' onClick={() => onDelete(item)} disabled={saving || item.contractSource === 'dashboard_fallback' || item.protected}>删除</button>
      </footer>
    </article>
  )
}

export function PayloadTable(props) {
  const {
    activeTab,
    payloads,
    query,
    selectedIds,
    loading,
    saving = false,
    error,
    onQuery,
    onSelect,
    onPreview,
    onCreate,
    onEdit,
    onDelete,
    onImport,
    onExport,
    onBlocked,
    onTag,
  } = props
  const allSelected = payloads.length > 0 && payloads.every((row) => selectedIds.has(row.id))
  const selectedRows = payloads.filter((row) => selectedIds.has(row.id))
  const tabLabel = payloadTabs.find((tab) => tab.key === activeTab)?.label || 'payload'
  const writableTab = isWritablePayloadType(activeTab)

  return (
    <section className='fx-int-payloads'>
      <div className='fx-int-toolbar'>
        <input value={query} onChange={(event) => onQuery(event.target.value)} placeholder='搜索名称、标签、说明' />
        <div className='fx-int-actions'>
          <button type='button' onClick={writableTab ? onCreate : () => onBlocked('payloadCreate')} disabled={saving || !writableTab}>新增</button>
          <button type='button' onClick={() => onImport(selectedRows)}>导入到业务组</button>
          <button type='button' onClick={() => onExport(selectedRows)}>批量导出</button>
        </div>
      </div>
      {!writableTab && <BlockedPanel message={blockedContracts[activeTab] || blockedContracts.payloadCreate} />}
      {error && <BlockedPanel message={error} />}
      <div className='fx-int-table-wrap'>
        <table className='fx-int-table'>
          <thead>
            <tr>
              <th className='is-check'>
                <input
                  type='checkbox'
                  checked={allSelected}
                  onChange={(event) => onSelect(event.target.checked ? payloads.map((row) => row.id) : [])}
                  aria-label={`选择全部${tabLabel}`}
                />
              </th>
              <th>名称</th>
              <th>标签</th>
              <th>说明</th>
              <th>更新人</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {payloads.map((row) => (
              <tr key={row.id}>
                <td className='is-check'>
                  <input
                    type='checkbox'
                    checked={selectedIds.has(row.id)}
                    onChange={(event) => toggleRow(event.target.checked, row.id, selectedIds, onSelect)}
                    aria-label={`选择 ${row.name}`}
                  />
                </td>
                <td><button type='button' className='fx-int-link' onClick={() => onPreview(row)}>{row.name}</button></td>
                <td><TagList tags={row.tags} onTag={onTag} /></td>
                <td>{row.note || '-'}</td>
                <td>{row.updatedBy === 'system' ? <span className='fx-int-status'>系统内置</span> : row.updatedBy || '-'}</td>
                <td>{row.missingContent ? <span className='fx-int-status is-blocked'>缺少 content</span> : <span className='fx-int-status'>可预览</span>}</td>
                <td><RowActions row={row} saving={saving} onPreview={onPreview} onImport={onImport} onExport={onExport} onEdit={onEdit} onDelete={onDelete} onBlocked={onBlocked} /></td>
              </tr>
            ))}
          </tbody>
        </table>
        {loading && <div className='fx-int-empty'>加载中...</div>}
        {!loading && payloads.length === 0 && <div className='fx-int-empty'>暂无{tabLabel}</div>}
      </div>
    </section>
  )
}

function toggleRow(checked, id, selectedIds, onSelect) {
  const next = new Set(selectedIds)
  if (checked) next.add(id)
  else next.delete(id)
  onSelect(Array.from(next))
}

function TagList({ tags, onTag }) {
  return (
    <span className='fx-int-tags'>
      {tags.length === 0 && <em>无</em>}
      {tags.map((tag) => <button type='button' key={tag} onClick={() => onTag(tag)}>{tag}</button>)}
    </span>
  )
}

function SystemDetailPanel({ row, onClose, onBlocked, onEdit, onDelete, onToggleMenu, saving }) {
  if (!row) return null
  return (
    <aside className='fx-int-drawer' aria-label='系统集成详情'>
      <header className='fx-int-drawer__head'>
        <div>
          <h2>{row.name}</h2>
          <p>{row.url || '暂无 URL；仅展示后端返回的脱敏配置预览。'}</p>
        </div>
        <button type='button' className='fx-int-icon-button' onClick={onClose} aria-label='关闭'>x</button>
      </header>
      <div className='fx-int-system-summary'>
        <span className='fx-int-status'>{row.isPrivate ? '团队可见' : '公开'}</span>
        <span className='fx-int-status'>{row.showInMenu ? '菜单显示' : '菜单隐藏'}</span>
        <span className='fx-int-status'>权重 {row.weight}</span>
        <span className='fx-int-status'>更新 {row.updateAt}</span>
      </div>
      <div className='fx-int-system-actions'>
        <button type='button' onClick={() => onBlocked('systemOpen')}>打开嵌入页</button>
        <button type='button' onClick={() => onToggleMenu(row)} disabled={saving}>切换菜单显示</button>
        <button type='button' onClick={() => onEdit(row)} disabled={saving}>编辑</button>
        <button type='button' className='is-danger' onClick={() => onDelete(row)} disabled={saving || row.builtin}>删除</button>
      </div>
      <BlockedPanel message={blockedContracts.systemOpen} />
      <pre className='fx-int-json'>{systemDetailJson(row)}</pre>
    </aside>
  )
}

function SystemFormModal({ state, saving, onDraft, onSubmit, onClose }) {
  if (!state) return null
  const { draft, mode, error } = state
  const patch = (key, value) => onDraft({ ...draft, [key]: value })
  const editing = mode === 'edit'
  return (
    <div className='fx-int-modal' role='dialog' aria-modal='true' aria-labelledby='fx-int-system-form-title'>
      <div className='fx-int-modal__backdrop' onClick={onClose} />
      <section className='fx-int-modal__panel fx-int-modal__panel--wide'>
        <header className='fx-int-modal__head'>
          <div>
            <h2 id='fx-int-system-form-title'>{editing ? '编辑系统集成' : '新增系统集成'}</h2>
            <p>只保存 FindX 同源相对路由；外部地址、凭据、令牌和嵌入打开不在本契约内。</p>
          </div>
          <button type='button' className='fx-int-icon-button' onClick={onClose} aria-label='关闭'>x</button>
        </header>
        <div className='fx-int-form'>
          <label>
            ID
            <input value={draft.id} disabled={editing} onChange={(event) => patch('id', event.target.value)} placeholder='留空由后端生成' />
          </label>
          <label>
            名称
            <input value={draft.name} maxLength={120} onChange={(event) => patch('name', event.target.value)} />
          </label>
          <label>
            FindX 路由
            <input value={draft.url} onChange={(event) => patch('url', event.target.value)} placeholder='/platform?section=overview' />
          </label>
          <label>
            配置预览
            <input value={draft.configPreview} onChange={(event) => patch('configPreview', event.target.value)} placeholder='/platform?section=overview' />
          </label>
          <label>
            权重
            <input type='number' value={draft.weight} onChange={(event) => patch('weight', event.target.value)} />
          </label>
          <label>
            团队 ID
            <input value={draft.teamIdsText} onChange={(event) => patch('teamIdsText', event.target.value)} placeholder='101, 102' />
          </label>
          <label className='fx-int-checkline'>
            <input type='checkbox' checked={draft.isPrivate} onChange={(event) => patch('isPrivate', event.target.checked)} />
            团队可见
          </label>
          <label className='fx-int-checkline'>
            <input type='checkbox' checked={draft.hide} onChange={(event) => patch('hide', event.target.checked)} />
            从菜单元数据隐藏
          </label>
        </div>
        {error && <div className='fx-int-alert is-error'>{sanitizeDisplayText(error)}</div>}
        <footer className='fx-int-modal__foot'>
          <button type='button' onClick={onClose} disabled={saving}>关闭</button>
          <button type='button' className='is-primary' onClick={onSubmit} disabled={saving}>{saving ? '保存中...' : '保存'}</button>
        </footer>
      </section>
    </div>
  )
}

function SystemOrderModal({ rows, open, saving, error, onChange, onSubmit, onClose }) {
  if (!open) return null
  return (
    <div className='fx-int-modal' role='dialog' aria-modal='true' aria-labelledby='fx-int-system-order-title'>
      <div className='fx-int-modal__backdrop' onClick={onClose} />
      <section className='fx-int-modal__panel'>
        <header className='fx-int-modal__head'>
          <div>
            <h2 id='fx-int-system-order-title'>调整排序</h2>
            <p>保存后调用后端权重契约，并重新读取列表确认排序。</p>
          </div>
          <button type='button' className='fx-int-icon-button' onClick={onClose} aria-label='关闭'>x</button>
        </header>
        <div className='fx-int-order-list'>
          {rows.map((row) => (
            <label key={row.id}>
              <span>{row.name}</span>
              <input type='number' value={row.weightDraft} onChange={(event) => onChange(row.id, event.target.value)} />
            </label>
          ))}
        </div>
        {error && <div className='fx-int-alert is-error'>{sanitizeDisplayText(error)}</div>}
        <footer className='fx-int-modal__foot'>
          <button type='button' onClick={onClose} disabled={saving}>关闭</button>
          <button type='button' className='is-primary' onClick={onSubmit} disabled={saving}>{saving ? '保存中...' : '保存排序'}</button>
        </footer>
      </section>
    </div>
  )
}

function RowActions({ row, saving, onPreview, onImport, onExport, onEdit, onDelete, onBlocked }) {
  const writable = isWritablePayloadType(row.type)
  const protectedRow = row.protected || row.fallbackOnly
  return (
    <span className='fx-int-row-actions'>
      <button type='button' onClick={() => onPreview(row)}>预览</button>
      <button type='button' onClick={() => onImport([row])}>导入</button>
      <button type='button' onClick={() => onExport([row])}>导出</button>
      <button type='button' onClick={writable && !protectedRow ? () => onEdit(row) : () => onBlocked(writable ? 'payloadEdit' : 'payloadCreate')} disabled={saving || !writable || protectedRow}>编辑</button>
      <button type='button' className='is-danger' onClick={writable && !protectedRow ? () => onDelete(row) : () => onBlocked('payloadDelete')} disabled={saving || !writable || protectedRow}>删除</button>
    </span>
  )
}

export function ComponentDrawer(props) {
  const {
    component,
    activeTab,
    payloads,
    payloadLoading,
    payloadSaving,
    payloadError,
    query,
    selectedIds,
    onClose,
    onTab,
    onQuery,
    onSelect,
    onPreview,
    onPayloadCreate,
    onPayloadEdit,
    onPayloadDelete,
    onComponentEdit,
    onImport,
    onExport,
    onBlocked,
    onTag,
  } = props
  if (!component) return null

  return (
    <aside className='fx-int-drawer' aria-label='模板详情'>
      <header className='fx-int-drawer__head'>
        <div className='fx-int-drawer__title'>
          <Logo item={component} />
          <div>
            <h2>{sanitizeDisplayText(component.ident)}</h2>
            <p>{sanitizeDisplayText(component.readme) || '暂无说明'}</p>
          </div>
        </div>
        <button type='button' className='fx-int-icon-button' onClick={onClose} aria-label='关闭'>x</button>
      </header>
      {component.contractSource === 'dashboard_fallback' && <BlockedPanel message={blockedContracts.components} />}
      <nav className='fx-int-tabs' aria-label='模板分类'>
        {payloadTabs.map((tab) => <button type='button' key={tab.key} className={tab.key === activeTab ? 'is-active' : ''} onClick={() => onTab(tab.key)}>{tab.label}</button>)}
      </nav>
      {activeTab === 'instructions' && <Instructions component={component} onEdit={onComponentEdit} saving={payloadSaving} />}
      {activeTab !== 'instructions' && (
        <PayloadTable activeTab={activeTab} payloads={payloads} query={query} selectedIds={selectedIds} loading={payloadLoading} saving={payloadSaving} error={payloadError} onQuery={onQuery} onSelect={onSelect} onPreview={onPreview} onCreate={onPayloadCreate} onEdit={onPayloadEdit} onDelete={onPayloadDelete} onImport={onImport} onExport={onExport} onBlocked={onBlocked} onTag={onTag} />
      )}
    </aside>
  )
}

function Instructions({ component, onEdit, saving }) {
  return (
    <section className='fx-int-instructions'>
      <textarea value={sanitizeDisplayText(component.readme) || '暂无说明'} readOnly rows={9} />
      <button type='button' className='is-primary' onClick={() => onEdit(component)} disabled={saving || component.contractSource === 'dashboard_fallback' || component.protected}>编辑使用说明</button>
      {(component.contractSource === 'dashboard_fallback' || component.protected) && <BlockedPanel message={blockedContracts.componentEdit} />}
    </section>
  )
}

export function ComponentFormModal({ state, saving, onDraft, onSubmit, onClose }) {
  if (!state) return null
  const { draft, mode, error } = state
  const editing = mode === 'edit'
  const patch = (key, value) => onDraft({ ...draft, [key]: value })
  return (
    <div className='fx-int-modal' role='dialog' aria-modal='true' aria-labelledby='fx-int-component-form-title'>
      <div className='fx-int-modal__backdrop' onClick={onClose} />
      <section className='fx-int-modal__panel fx-int-modal__panel--wide'>
        <header className='fx-int-modal__head'>
          <div>
            <h2 id='fx-int-component-form-title'>{editing ? '编辑组件' : '新增组件'}</h2>
            <p>保存 ident、启停状态、Logo 和使用说明。内置静态行由后端保护，失败信息会脱敏展示。</p>
          </div>
          <button type='button' className='fx-int-icon-button' onClick={onClose} aria-label='关闭'>x</button>
        </header>
        <div className='fx-int-form'>
          <label>
            ID
            <input value={draft.id} disabled={editing} onChange={(event) => patch('id', event.target.value)} placeholder='留空由后端生成' />
          </label>
          <label>
            组件标识
            <input value={draft.ident} maxLength={128} onChange={(event) => patch('ident', event.target.value)} />
          </label>
          <label>
            名称
            <input value={draft.name} maxLength={255} onChange={(event) => patch('name', event.target.value)} />
          </label>
          <label>
            Logo 路径
            <input value={draft.logo} onChange={(event) => patch('logo', event.target.value)} placeholder='/image/logos/host.png' />
          </label>
          <label className='fx-int-checkline'>
            <input type='checkbox' checked={!draft.disabled} onChange={(event) => patch('disabled', !event.target.checked)} />
            启用组件
          </label>
          <label className='is-wide'>
            使用说明
            <textarea rows={9} value={draft.readme} onChange={(event) => patch('readme', event.target.value)} />
          </label>
        </div>
        {error && <div className='fx-int-alert is-error'>{sanitizeDisplayText(error)}</div>}
        <footer className='fx-int-modal__foot'>
          <button type='button' onClick={onClose} disabled={saving}>关闭</button>
          <button type='button' className='is-primary' onClick={onSubmit} disabled={saving}>{saving ? '保存中...' : '保存'}</button>
        </footer>
      </section>
    </div>
  )
}

export function PayloadFormModal({ state, saving, onDraft, onSubmit, onClose }) {
  if (!state) return null
  const { draft, mode, error } = state
  const editing = mode === 'edit'
  const patch = (key, value) => onDraft({ ...draft, [key]: value })
  return (
    <div className='fx-int-modal' role='dialog' aria-modal='true' aria-labelledby='fx-int-payload-form-title'>
      <div className='fx-int-modal__backdrop' onClick={onClose} />
      <section className='fx-int-modal__panel fx-int-modal__panel--wide'>
        <header className='fx-int-modal__head'>
          <div>
            <h2 id='fx-int-payload-form-title'>{editing ? '编辑 payload' : '新增 payload'}</h2>
            <p>仅 dashboard、collect、alert 支持写入。dashboard/alert content 提交前必须通过 JSON 与敏感字段校验。</p>
          </div>
          <button type='button' className='fx-int-icon-button' onClick={onClose} aria-label='关闭'>x</button>
        </header>
        <div className='fx-int-form'>
          <label>
            ID
            <input value={draft.id} disabled={editing} onChange={(event) => patch('id', event.target.value)} placeholder='留空由后端生成' />
          </label>
          <label>
            UUID
            <input value={draft.uuid} onChange={(event) => patch('uuid', event.target.value)} placeholder='留空使用生成值' />
          </label>
          <label>
            类型
            <input value={draft.type} readOnly />
          </label>
          <label>
            component_id
            <input value={draft.componentId} readOnly />
          </label>
          <label>
            分类
            <input value={draft.cate} maxLength={128} onChange={(event) => patch('cate', event.target.value)} />
          </label>
          <label>
            名称
            <input value={draft.name} maxLength={255} onChange={(event) => patch('name', event.target.value)} />
          </label>
          <label className='is-wide'>
            content
            <textarea rows={16} value={draft.content} onChange={(event) => patch('content', event.target.value)} />
          </label>
        </div>
        {error && <div className='fx-int-alert is-error'>{sanitizeDisplayText(error)}</div>}
        <footer className='fx-int-modal__foot'>
          <button type='button' onClick={onClose} disabled={saving}>关闭</button>
          <button type='button' className='is-primary' onClick={onSubmit} disabled={saving}>{saving ? '保存中...' : '保存'}</button>
        </footer>
      </section>
    </div>
  )
}

export function ImportModal(props) {
  const {
    rows,
    draft,
    groups,
    groupsLoading,
    groupsError,
    submitting,
    error,
    result,
    onDraft,
    onClose,
    onSubmit,
    onOpenDashboard,
  } = props
  if (!rows.length) return null
  const patch = (key, value) => onDraft({ ...draft, [key]: value })
  const selectedGroup = groups.find((group) => group.id === draft.businessGroupId)
  const showGroupsEmptyWarning = !groupsLoading && !groupsError && groups.length === 0
  const businessGroupBlocked = Boolean(groupsError) || showGroupsEmptyWarning
  const submitDisabled = submitting || groupsLoading || businessGroupBlocked || !draft.businessGroupId

  return (
    <div className='fx-int-modal' role='dialog' aria-modal='true' aria-labelledby='fx-int-import-title'>
      <div className='fx-int-modal__backdrop' onClick={onClose} />
      <section className='fx-int-modal__panel fx-int-modal__panel--wide'>
        <header className='fx-int-modal__head'>
          <div>
            <h2 id='fx-int-import-title'>导入仪表盘模板</h2>
            <p>仅对 FindX 仪表盘模板调用同源导入接口；业务组是必选目标，其他 builtin payload 导入仍待接入。</p>
          </div>
          <button type='button' className='fx-int-icon-button' onClick={onClose} aria-label='关闭'>x</button>
        </header>
        <section className='fx-int-import-selected'>
          {rows.map((row) => (
            <article key={row.id}>
              <strong>{sanitizeDisplayText(row.name)}</strong>
              <span>{sanitizeDisplayText(row.note) || '暂无说明'}</span>
              <TagList tags={row.tags || []} onTag={() => {}} />
            </article>
          ))}
        </section>
        <div className='fx-int-form'>
          {rows.length === 1 && (
            <label>
              标题覆盖
              <input value={draft.title} maxLength={120} onChange={(event) => patch('title', event.target.value)} />
            </label>
          )}
          <label>
            业务组目标
            <select
              value={draft.businessGroupId}
              disabled={groupsLoading || groups.length === 0}
              onChange={(event) => {
                const value = event.target.value
                const group = groups.find((item) => item.id === value)
                onDraft({ ...draft, businessGroupId: value, resourceGroupId: group?.id || draft.resourceGroupId })
              }}
            >
              <option value=''>{groupsLoading ? '业务组加载中...' : '请选择业务组'}</option>
              {groups.map((group) => <option key={group.id} value={group.id}>{group.name || group.id}</option>)}
            </select>
          </label>
          <label>
            工作空间 ID（可选）
            <input value={draft.workspaceId} onChange={(event) => patch('workspaceId', event.target.value)} />
          </label>
          <label>
            资源组 ID（由业务组确定）
            <input value={draft.resourceGroupId} onChange={(event) => patch('resourceGroupId', event.target.value)} />
          </label>
          <label className='is-wide'>
            标签覆盖（留空则使用模板默认标签）
            <input value={draft.tagsText} onChange={(event) => patch('tagsText', event.target.value)} />
          </label>
          <label className='is-wide'>
            变量 JSON 覆盖（必须是对象，提交前会校验）
            <textarea rows={10} value={draft.variablesText} onChange={(event) => patch('variablesText', event.target.value)} />
          </label>
        </div>
        {selectedGroup && (
          <div className='fx-int-alert'>
            将使用业务组 <strong>{selectedGroup.name || selectedGroup.id}</strong> 作为 resource_group_id。
          </div>
        )}
        {businessGroupBlocked && (
          <div className='fx-int-alert is-warning'>
            <strong></strong>
            <span>业务组列表不可用或为空。成熟源码导入语义要求业务组必选，当前不会发送导入请求。</span>
          </div>
        )}
        {error && <div className='fx-int-alert is-error'>{sanitizeDisplayText(error)}</div>}
        {result && <ImportResult result={result} onOpenDashboard={onOpenDashboard} />}
        <footer className='fx-int-modal__foot'>
          <button type='button' onClick={onClose}>关闭</button>
          <button type='button' className='is-primary' disabled={submitDisabled} onClick={onSubmit}>{submitting ? '导入中...' : '提交导入'}</button>
        </footer>
      </section>
    </div>
  )
}

function ImportResult({ result, onOpenDashboard }) {
  const dashboards = result?.dashboards || []
  if (!dashboards.length) return null
  return (
    <div className='fx-int-alert is-success'>
      <strong>导入成功</strong>
      {dashboards.map((dashboard) => (
        <span key={dashboard.id || dashboard.title}>
          {dashboard.title}{dashboard.id ? `（${dashboard.id}）` : ''}
          {dashboard.id && (
            <a
              className='fx-int-link'
              href={`/dashboards?section=detail&id=${encodeURIComponent(dashboard.id)}`}
              onClick={(event) => {
                event.preventDefault()
                onOpenDashboard(dashboard)
              }}
            >
              打开仪表盘
            </a>
          )}
        </span>
      ))}
    </div>
  )
}

export function PreviewModal({ row, content, loading, error, onClose, onImport }) {
  if (!row) return null
  return (
    <div className='fx-int-modal' role='dialog' aria-modal='true' aria-labelledby='fx-int-preview-title'>
      <div className='fx-int-modal__backdrop' onClick={onClose} />
      <section className='fx-int-modal__panel'>
        <header className='fx-int-modal__head'>
          <div>
            <h2 id='fx-int-preview-title'>预览：{row.name}</h2>
            <p>{row.note || '暂无说明'}</p>
          </div>
          <button type='button' className='fx-int-icon-button' onClick={onClose} aria-label='关闭'>x</button>
        </header>
        {error && <div className='fx-int-alert is-error'>{error}</div>}
        <pre className='fx-int-json'>{loading ? '加载中...' : content}</pre>
        <footer className='fx-int-modal__foot'>
          <button type='button' onClick={onClose}>关闭</button>
          <button type='button' className='is-primary' onClick={() => onImport([row])}>导入到业务组</button>
        </footer>
      </section>
    </div>
  )
}
