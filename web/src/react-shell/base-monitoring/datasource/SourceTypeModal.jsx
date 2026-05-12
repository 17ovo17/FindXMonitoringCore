import React from 'react'
import { datasourceTypes } from './datasourceModel.js'

export function SourceTypeModal({ visible, onClose, onChoose }) {
  if (!visible) return null

  return (
    <div className='fx-ds-modal' role='dialog' aria-modal='true' aria-labelledby='fx-ds-type-title'>
      <div className='fx-ds-modal__backdrop' onClick={onClose} />
      <section className='fx-ds-modal__panel'>
        <header className='fx-ds-modal__head'>
          <h2 id='fx-ds-type-title'>选择数据源类型</h2>
          <button type='button' className='fx-ds-icon-button' onClick={onClose} aria-label='关闭'>×</button>
        </header>
        <div className='fx-ds-type-grid'>
          {datasourceTypes.map((item) => (
            <button className='fx-ds-type-card' key={item.type} type='button' onClick={() => onChoose(item)}>
              <img
                className='fx-ds-type-card__logo'
                src={item.logo}
                alt={item.name}
                width='32'
                height='32'
              />
              <strong>{item.name}</strong>
              <span>{item.description}</span>
              <em>{item.supported ? '可配置' : '查询功能开发中'}</em>
            </button>
          ))}
        </div>
      </section>
    </div>
  )
}
