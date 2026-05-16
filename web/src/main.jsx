import React from 'react'
import ReactDOM from 'react-dom'
import { BrowserRouter } from 'react-router-dom'
import { FindXReactShell } from './react-shell'

export const bootstrapFindXReactShell = (mountNode, options = {}) => {
  if (!mountNode) {
    throw new Error('FindX React shell mount node is required')
  }

  ReactDOM.render(
    <React.StrictMode>
      <BrowserRouter basename={options.basename || '/'}>
        <FindXReactShell
          authBoundary={options.authBoundary}
          navigationItems={options.navigationItems}
          themeBoundary={options.themeBoundary}
        />
      </BrowserRouter>
    </React.StrictMode>,
    mountNode,
  )
}

const reactMountNode = document.getElementById('app') || document.getElementById('findx-react-root')

if (!reactMountNode) {
  throw new Error('FindX React shell cannot start: missing #app or #findx-react-root')
}

bootstrapFindXReactShell(reactMountNode)
