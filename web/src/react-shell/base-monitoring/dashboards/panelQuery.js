import { post } from '../../api/http.js'

/**
 * Extract PromQL expressions from a panel's targets or direct query field.
 * Panels may store queries as:
 *   - targets: [{ expr: "..." }, ...]
 *   - query / expr / expression / metric (single string)
 */
function extractTargets(panel) {
  if (Array.isArray(panel.targets) && panel.targets.length > 0) {
    return panel.targets
      .map((t) => ({ expr: t.expr || t.expression || t.query || '', legendFormat: t.legendFormat || '' }))
      .filter((t) => t.expr)
  }
  const expr = panel.expr || panel.query || panel.expression || panel.metric || ''
  if (expr) return [{ expr, legendFormat: '' }]
  return []
}

/**
 * Query Prometheus for a single panel's data.
 * @param {object} panel - Raw panel object (not normalized)
 * @param {{ start: number, end: number, step: number }} timeRange - Unix timestamps + step
 * @param {string} datasourceId - Datasource identifier
 * @returns {Promise<{ series: Array<{ metric: object, values: Array }>, error: string|null }>}
 */
export async function queryPanel(panel, timeRange, datasourceId) {
  const targets = extractTargets(panel)
  if (targets.length === 0) {
    return { series: [], error: null }
  }

  try {
    const allSeries = []
    for (const target of targets) {
      const result = await post('/monitor/query-range', {
        query: target.expr,
        start: timeRange.start,
        end: timeRange.end,
        step: timeRange.step || 15,
        datasource_id: datasourceId,
      })
      const matrix = result?.data?.result || result?.result || []
      for (const item of matrix) {
        allSeries.push({
          metric: item.metric || {},
          values: (item.values || []).map(([ts, val]) => [Number(ts), Number(val)]),
          legendFormat: target.legendFormat,
        })
      }
    }
    return { series: allSeries, error: null }
  } catch (err) {
    return { series: [], error: err?.message || '查询失败' }
  }
}
