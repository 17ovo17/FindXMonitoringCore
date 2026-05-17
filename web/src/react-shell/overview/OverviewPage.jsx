/**
 * OverviewPage — NOC 大屏风格全局概览
 * 顶部：4 个告警级别统计卡片（紧急/严重/警告/正常）
 * 左侧：TOP5 告警主机列表
 * 右侧：资源健康度进度条
 * 中间：最近告警事件列表（最新 10 条）
 * 底部：AI 快捷入口卡片（查告警/查主机/查日志/生成脚本/触发巡检）
 * API: GET /api/v1/overview/stats
 *
 * 此文件为独立入口，实际渲染逻辑复用 base-monitoring/OverviewPage.jsx
 */
export { OverviewPage } from '../base-monitoring/OverviewPage.jsx'
