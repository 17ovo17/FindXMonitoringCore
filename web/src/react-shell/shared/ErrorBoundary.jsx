import React from 'react'

/**
 * ErrorBoundary — 捕获子组件渲染错误，展示友好的错误页面。
 * 开发环境下显示堆栈信息，生产环境仅显示错误消息。
 */
export class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props)
    this.state = { hasError: false, error: null, errorInfo: null }
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error }
  }

  componentDidCatch(error, errorInfo) {
    this.setState({ errorInfo })
    console.error('[ErrorBoundary] Uncaught error:', error, errorInfo)
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null, errorInfo: null })
  }

  render() {
    if (!this.state.hasError) {
      return this.props.children
    }

    const isDev = typeof process !== 'undefined'
      ? process.env.NODE_ENV !== 'production'
      : window.location.hostname === 'localhost'

    return (
      <div className="fx-error-boundary" role="alert">
        <div className="fx-error-boundary-content">
          <h2>页面出现错误</h2>
          <p className="fx-error-boundary-msg">
            {this.state.error?.message || '未知错误'}
          </p>
          {isDev && this.state.errorInfo && (
            <details className="fx-error-boundary-stack">
              <summary>错误堆栈（仅开发环境可见）</summary>
              <pre>{this.state.error?.stack}</pre>
              <pre>{this.state.errorInfo.componentStack}</pre>
            </details>
          )}
          <button
            type="button"
            className="fx-error-boundary-retry"
            onClick={this.handleRetry}
          >
            重试
          </button>
        </div>
      </div>
    )
  }
}
