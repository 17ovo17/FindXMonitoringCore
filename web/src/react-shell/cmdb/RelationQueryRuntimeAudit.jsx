import React from 'react'
import { displayText } from './assetsModel.js'

export function RelationQueryRuntimeAudit({ runtime }) {
  const missing = Array.isArray(runtime?.missing_contracts) ? runtime.missing_contracts : []
  return (
    <section className='fx-cmdb-query-runtime-audit'>
      <header>
        <strong>关系查询执行契约</strong>
        <span>{displayText(runtime?.status, '')}</span>
      </header>
      <p>{displayText(runtime?.message, '递归关系路径执行器和最终布局坐标尚未闭环；当前只展示关系查询规则 DTO，不展示伪造执行结果。')}</p>
      <div>
        {missing.map(item => <code key={item}>{item}</code>)}
      </div>
    </section>
  )
}
