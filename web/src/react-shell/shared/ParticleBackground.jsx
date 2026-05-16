import React from 'react'
import './login-bg.css'

/**
 * 登录页 CSS 动态背景 — NOC/SOC 运维中心风格
 * 纯 CSS 动画，无 canvas / JS 动画，组件卸载后零性能开销
 */
export function ParticleBackground() {
  return (
    <div className="fx-login-bg">
      {/* 网格拓扑线 */}
      <div className="fx-login-grid" />

      {/* 六边形纹理叠加 */}
      <div className="fx-login-hex" />

      {/* 数据节点 */}
      <div className="fx-login-nodes">
        <div className="fx-login-node" />
        <div className="fx-login-node" />
        <div className="fx-login-node" />
        <div className="fx-login-node" />
        <div className="fx-login-node" />
        <div className="fx-login-node" />
        <div className="fx-login-node" />
        <div className="fx-login-node" />
        <div className="fx-login-node" />
        <div className="fx-login-node" />
      </div>

      {/* 连接线 */}
      <div className="fx-login-connections">
        <svg viewBox="0 0 100 100" preserveAspectRatio="none">
          <defs>
            <linearGradient id="fx-conn-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stopColor="rgba(23, 105, 255, 0.6)" />
              <stop offset="50%" stopColor="rgba(0, 212, 255, 0.4)" />
              <stop offset="100%" stopColor="rgba(99, 102, 241, 0.6)" />
            </linearGradient>
          </defs>
          <line className="fx-login-conn-line" x1="18" y1="12" x2="48" y2="15" />
          <line className="fx-login-conn-line" x1="72" y1="25" x2="42" y2="35" />
          <line className="fx-login-conn-line" x1="8" y1="45" x2="42" y2="35" />
          <line className="fx-login-conn-line" x1="62" y1="58" x2="85" y2="68" />
          <line className="fx-login-conn-line" x1="22" y1="82" x2="45" y2="75" />
        </svg>
      </div>

      {/* 扫描线 */}
      <div className="fx-login-scan" />

      {/* 数据流粒子 */}
      <div className="fx-login-streams">
        <div className="fx-login-stream" />
        <div className="fx-login-stream" />
        <div className="fx-login-stream" />
        <div className="fx-login-stream" />
        <div className="fx-login-stream" />
        <div className="fx-login-stream" />
      </div>

      {/* 雷达扫描 */}
      <div className="fx-login-radar" />

      {/* 环境光晕 */}
      <div className="fx-login-glow" />
    </div>
  )
}
