import React, { useState } from 'react'

const VARIABLE_TYPES = [
  { key: 'datasource', label: '数据源 (Datasource)' },
  { key: 'query', label: '查询 (Query)' },
  { key: 'custom', label: '自定义 (Custom)' },
  { key: 'textbox', label: '文本框 (Textbox)' },
  { key: 'constant', label: '常量 (Constant)' },
  { key: 'hostIdent', label: '主机标识 (HostIdent)' },
  { key: 'datasourceIdentifier', label: '数据源标识 (DatasourceIdentifier)' },
]

function VariableEditForm({ variable, onSave, onCancel }) {
  const [name, setName] = useState(variable?.name || '')
  const [type, setType] = useState(variable?.type || 'query')
  const [definition, setDefinition] = useState(variable?.query || '')
  const [multi, setMulti] = useState(variable?.multi || false)
  const [hide, setHide] = useState(variable?.hide || false)

  const handleSubmit = () => {
    onSave({ ...variable, name, type, query: definition, multi, hide })
  }

  return (
    <div className="fx-varlist-edit">
      <label className="fx-pe-field">
        <span>变量名称</span>
        <input value={name} onChange={(e) => setName(e.target.value)} />
      </label>
      <label className="fx-pe-field">
        <span>变量类型</span>
        <select value={type} onChange={(e) => setType(e.target.value)}>
          {VARIABLE_TYPES.map((t) => (
            <option key={t.key} value={t.key}>{t.label}</option>
          ))}
        </select>
      </label>
      <label className="fx-pe-field">
        <span>变量定义</span>
        <input
          value={definition}
          onChange={(e) => setDefinition(e.target.value)}
          placeholder="如 label_values(metric, label)"
        />
      </label>
      <label className="fx-pe-field fx-settings-var-multi">
        <input
          type="checkbox"
          checked={multi}
          onChange={(e) => setMulti(e.target.checked)}
        />
        <span>支持多选</span>
      </label>
      <label className="fx-pe-field fx-settings-var-multi">
        <input
          type="checkbox"
          checked={hide}
          onChange={(e) => setHide(e.target.checked)}
        />
        <span>隐藏变量</span>
      </label>
      <div className="fx-dash-actions" style={{ marginTop: 12 }}>
        <button type="button" onClick={onCancel}>取消</button>
        <button type="button" className="is-primary" onClick={handleSubmit}>
          确定
        </button>
      </div>
    </div>
  )
}

function getTypeLabel(type) {
  const found = VARIABLE_TYPES.find((t) => t.key === type)
  return found ? found.label : type || '-'
}

/**
 * 变量列表 Modal（对齐夜莺 Variables 目录）
 * 表格形式展示变量，操作列有上移/下移/克隆/删除
 */
export default function VariableListModal({ variables, onChange, onClose }) {
  const [editingIndex, setEditingIndex] = useState(null)

  const moveUp = (index) => {
    if (index <= 0) return
    const next = [...variables]
    const temp = next[index - 1]
    next[index - 1] = next[index]
    next[index] = temp
    onChange(next)
  }

  const moveDown = (index) => {
    if (index >= variables.length - 1) return
    const next = [...variables]
    const temp = next[index + 1]
    next[index + 1] = next[index]
    next[index] = temp
    onChange(next)
  }

  const cloneVariable = (index) => {
    const source = variables[index]
    const cloned = { ...source, name: `${source.name}_copy` }
    const next = [...variables]
    next.splice(index + 1, 0, cloned)
    onChange(next)
  }

  const removeVariable = (index) => {
    onChange(variables.filter((_, i) => i !== index))
  }

  const addVariable = () => {
    const newVar = { name: '', type: 'query', query: '', multi: false, hide: false }
    onChange([...variables, newVar])
    setEditingIndex(variables.length)
  }

  const handleEditSave = (updated) => {
    const next = variables.map((v, i) => i === editingIndex ? updated : v)
    onChange(next)
    setEditingIndex(null)
  }

  return (
    <div className="fx-dash-modal">
      <div className="fx-dash-modal__body fx-settings-modal">
        <header>
          <h2>变量列表</h2>
          <button type="button" onClick={onClose}>x</button>
        </header>
        <div className="fx-settings-body">
          {editingIndex !== null ? (
            <VariableEditForm
              variable={variables[editingIndex]}
              onSave={handleEditSave}
              onCancel={() => setEditingIndex(null)}
            />
          ) : (
            <div className="fx-varlist-table-wrap">
              <table className="fx-varlist-table">
                <thead>
                  <tr>
                    <th>变量名称</th>
                    <th>变量类型</th>
                    <th>变量定义</th>
                    <th>隐藏变量</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {variables.map((v, i) => (
                    <tr key={i}>
                      <td>
                        <button
                          type="button"
                          className="is-link"
                          onClick={() => setEditingIndex(i)}
                        >
                          {v.name || '(未命名)'}
                        </button>
                      </td>
                      <td>{getTypeLabel(v.type)}</td>
                      <td className="fx-varlist-def">
                        {v.query || '-'}
                      </td>
                      <td>{v.hide ? '是' : '否'}</td>
                      <td>
                        <div className="fx-varlist-ops">
                          <button
                            type="button"
                            title="下移"
                            disabled={i >= variables.length - 1}
                            onClick={() => moveDown(i)}
                          >
                            ↓
                          </button>
                          <button
                            type="button"
                            title="上移"
                            disabled={i <= 0}
                            onClick={() => moveUp(i)}
                          >
                            ↑
                          </button>
                          <button
                            type="button"
                            title="克隆"
                            onClick={() => cloneVariable(i)}
                          >
                            &#128203;
                          </button>
                          <button
                            type="button"
                            title="删除"
                            className="fx-panel-menu__danger"
                            onClick={() => removeVariable(i)}
                          >
                            &#128465;
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {variables.length === 0 && (
                    <tr>
                      <td colSpan={5} className="fx-dash-empty">
                        暂无变量
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          )}
        </div>
        {editingIndex === null && (
          <footer className="fx-settings-footer">
            <button
              type="button"
              className="is-primary"
              onClick={addVariable}
            >
              添加变量
            </button>
          </footer>
        )}
      </div>
    </div>
  )
}
