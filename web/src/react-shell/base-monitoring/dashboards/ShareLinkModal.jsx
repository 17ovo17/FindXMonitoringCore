import React, { useState, useRef } from 'react'

export default function ShareLinkModal({ onClose }) {
  const [copied, setCopied] = useState(false)
  const inputRef = useRef(null)
  const shareUrl = window.location.href

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(shareUrl)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      if (inputRef.current) {
        inputRef.current.select()
        document.execCommand('copy')
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
      }
    }
  }

  return (
    <div className="fx-dash-modal">
      <div className="fx-dash-modal__body">
        <header>
          <h2>分享仪表盘</h2>
          <button type="button" onClick={onClose}>x</button>
        </header>
        <div className="fx-share-content">
          <p className="fx-share-hint">复制以下链接分享当前仪表盘（包含时间范围和变量值）：</p>
          <div className="fx-share-url-row">
            <input ref={inputRef} className="fx-share-url-input" value={shareUrl} readOnly onClick={(e) => e.target.select()} />
            <button type="button" className="is-primary" onClick={handleCopy}>
              {copied ? '已复制' : '复制'}
            </button>
          </div>
          <p className="fx-share-note">注：接收者需要有仪表盘访问权限才能查看。</p>
        </div>
      </div>
    </div>
  )
}
