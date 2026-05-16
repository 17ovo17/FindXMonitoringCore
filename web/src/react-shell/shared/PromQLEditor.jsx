import React, { useRef, useCallback, useMemo } from 'react'

/**
 * PromQL 编辑器 - 基于 Monaco Editor 的 PromQL 语言支持组件
 * 提供自动补全（指标名、函数、标签）和语法高亮
 */

const PROMQL_FUNCTIONS = [
  { label: 'rate', detail: 'rate(v range-vector)', doc: '计算范围向量中时间序列的每秒平均增长率' },
  { label: 'irate', detail: 'irate(v range-vector)', doc: '计算范围向量中时间序列的瞬时增长率' },
  { label: 'increase', detail: 'increase(v range-vector)', doc: '计算范围向量中时间序列的增长量' },
  { label: 'sum', detail: 'sum(v instant-vector)', doc: '对向量中所有元素求和' },
  { label: 'avg', detail: 'avg(v instant-vector)', doc: '对向量中所有元素求平均' },
  { label: 'max', detail: 'max(v instant-vector)', doc: '取向量中所有元素的最大值' },
  { label: 'min', detail: 'min(v instant-vector)', doc: '取向量中所有元素的最小值' },
  { label: 'count', detail: 'count(v instant-vector)', doc: '计算向量中元素的数量' },
  { label: 'stddev', detail: 'stddev(v instant-vector)', doc: '计算向量中元素的标准差' },
  { label: 'stdvar', detail: 'stdvar(v instant-vector)', doc: '计算向量中元素的方差' },
  { label: 'topk', detail: 'topk(k, v instant-vector)', doc: '返回最大的 k 个元素' },
  { label: 'bottomk', detail: 'bottomk(k, v instant-vector)', doc: '返回最小的 k 个元素' },
  { label: 'histogram_quantile', detail: 'histogram_quantile(φ, b instant-vector)', doc: '从直方图中计算分位数' },
  { label: 'label_replace', detail: 'label_replace(v, dst, replacement, src, regex)', doc: '替换标签值' },
  { label: 'label_join', detail: 'label_join(v, dst, separator, src...)', doc: '连接标签值' },
  { label: 'abs', detail: 'abs(v instant-vector)', doc: '返回绝对值' },
  { label: 'ceil', detail: 'ceil(v instant-vector)', doc: '向上取整' },
  { label: 'floor', detail: 'floor(v instant-vector)', doc: '向下取整' },
  { label: 'round', detail: 'round(v instant-vector, to_nearest)', doc: '四舍五入' },
  { label: 'clamp', detail: 'clamp(v, min, max)', doc: '将值限制在范围内' },
  { label: 'clamp_max', detail: 'clamp_max(v, max)', doc: '将值限制在最大值以下' },
  { label: 'clamp_min', detail: 'clamp_min(v, min)', doc: '将值限制在最小值以上' },
  { label: 'delta', detail: 'delta(v range-vector)', doc: '计算范围向量中第一个和最后一个值的差' },
  { label: 'deriv', detail: 'deriv(v range-vector)', doc: '使用简单线性回归计算导数' },
  { label: 'changes', detail: 'changes(v range-vector)', doc: '计算值变化的次数' },
  { label: 'resets', detail: 'resets(v range-vector)', doc: '计算计数器重置的次数' },
  { label: 'absent', detail: 'absent(v instant-vector)', doc: '如果向量为空则返回 1' },
  { label: 'absent_over_time', detail: 'absent_over_time(v range-vector)', doc: '如果范围向量为空则返回 1' },
  { label: 'sort', detail: 'sort(v instant-vector)', doc: '升序排列' },
  { label: 'sort_desc', detail: 'sort_desc(v instant-vector)', doc: '降序排列' },
  { label: 'time', detail: 'time()', doc: '返回当前 Unix 时间戳' },
  { label: 'vector', detail: 'vector(s scalar)', doc: '将标量转为向量' },
  { label: 'scalar', detail: 'scalar(v instant-vector)', doc: '将单元素向量转为标量' },
  { label: 'avg_over_time', detail: 'avg_over_time(v range-vector)', doc: '范围内平均值' },
  { label: 'min_over_time', detail: 'min_over_time(v range-vector)', doc: '范围内最小值' },
  { label: 'max_over_time', detail: 'max_over_time(v range-vector)', doc: '范围内最大值' },
  { label: 'sum_over_time', detail: 'sum_over_time(v range-vector)', doc: '范围内求和' },
  { label: 'count_over_time', detail: 'count_over_time(v range-vector)', doc: '范围内计数' },
  { label: 'quantile_over_time', detail: 'quantile_over_time(φ, v range-vector)', doc: '范围内分位数' },
  { label: 'last_over_time', detail: 'last_over_time(v range-vector)', doc: '范围内最后一个值' },
  { label: 'predict_linear', detail: 'predict_linear(v range-vector, t scalar)', doc: '线性预测' },
  { label: 'sgn', detail: 'sgn(v instant-vector)', doc: '返回符号（-1, 0, 1）' },
]

const PROMQL_KEYWORDS = [
  'by', 'without', 'on', 'ignoring', 'group_left', 'group_right',
  'bool', 'and', 'or', 'unless', 'offset',
]

// PromQL Monaco 语言定义
const PROMQL_LANGUAGE_DEF = {
  defaultToken: '',
  keywords: PROMQL_KEYWORDS,
  functions: PROMQL_FUNCTIONS.map(f => f.label),
  operators: ['=', '!=', '=~', '!~', '+', '-', '*', '/', '%', '^', '==', '!=', '>', '<', '>=', '<='],
  tokenizer: {
    root: [
      [/[a-zA-Z_]\w*/, {
        cases: {
          '@keywords': 'keyword',
          '@functions': 'predefined',
          '@default': 'identifier',
        },
      }],
      [/"[^"]*"/, 'string'],
      [/'[^']*'/, 'string'],
      [/`[^`]*`/, 'string'],
      [/\d+(\.\d+)?([eE][+-]?\d+)?/, 'number'],
      [/\d+(ms|s|m|h|d|w|y)/, 'number'],
      [/#.*$/, 'comment'],
      [/[{}()\[\]]/, 'delimiter.bracket'],
      [/[,]/, 'delimiter'],
      [/[=!~<>]+/, 'operator'],
    ],
  },
}

// PromQL Monaco 主题配色
const PROMQL_THEME = {
  base: 'vs',
  inherit: true,
  rules: [
    { token: 'keyword', foreground: '7c3aed' },
    { token: 'predefined', foreground: '1769ff' },
    { token: 'identifier', foreground: '17233c' },
    { token: 'string', foreground: '16a34a' },
    { token: 'number', foreground: 'e6550d' },
    { token: 'comment', foreground: '9ca3af', fontStyle: 'italic' },
    { token: 'operator', foreground: 'd97706' },
    { token: 'delimiter.bracket', foreground: '6b7280' },
  ],
  colors: {},
}

// 缓存指标名列表
let metricNamesCache = null
let metricNamesFetchTime = 0
const CACHE_TTL = 60000 // 1 分钟缓存

async function fetchMetricNames() {
  const now = Date.now()
  if (metricNamesCache && now - metricNamesFetchTime < CACHE_TTL) {
    return metricNamesCache
  }
  try {
    const resp = await fetch('/api/v1/prometheus/labels/__name__')
    if (!resp.ok) return metricNamesCache || []
    const json = await resp.json()
    metricNamesCache = json.data || json.dat || []
    metricNamesFetchTime = now
    return metricNamesCache
  } catch {
    return metricNamesCache || []
  }
}

async function fetchLabelNames() {
  try {
    const resp = await fetch('/api/v1/prometheus/labels')
    if (!resp.ok) return []
    const json = await resp.json()
    return json.data || json.dat || []
  } catch {
    return []
  }
}

function registerPromQL(monaco) {
  if (monaco.languages.getLanguages().some(l => l.id === 'promql')) return

  monaco.languages.register({ id: 'promql' })
  monaco.languages.setMonarchTokensProvider('promql', PROMQL_LANGUAGE_DEF)
  monaco.editor.defineTheme('promql-theme', PROMQL_THEME)

  monaco.languages.registerCompletionItemProvider('promql', {
    triggerCharacters: ['{', ',', '(', ' '],
    provideCompletionItems: async (model, position) => {
      const word = model.getWordUntilPosition(position)
      const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn,
      }

      const suggestions = []

      // 函数补全
      PROMQL_FUNCTIONS.forEach(fn => {
        suggestions.push({
          label: fn.label,
          kind: monaco.languages.CompletionItemKind.Function,
          insertText: fn.label + '($0)',
          insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
          detail: fn.detail,
          documentation: fn.doc,
          range,
        })
      })

      // 关键字补全
      PROMQL_KEYWORDS.forEach(kw => {
        suggestions.push({
          label: kw,
          kind: monaco.languages.CompletionItemKind.Keyword,
          insertText: kw + ' ',
          range,
        })
      })

      // 指标名补全
      const metrics = await fetchMetricNames()
      metrics.forEach(name => {
        suggestions.push({
          label: name,
          kind: monaco.languages.CompletionItemKind.Variable,
          insertText: name,
          detail: '指标',
          range,
        })
      })

      // 标签名补全（在 {} 内时）
      const lineContent = model.getLineContent(position.lineNumber)
      const beforeCursor = lineContent.substring(0, position.column - 1)
      if (beforeCursor.includes('{') && !beforeCursor.endsWith('}')) {
        const labels = await fetchLabelNames()
        labels.forEach(label => {
          suggestions.push({
            label: label,
            kind: monaco.languages.CompletionItemKind.Property,
            insertText: label + '="$0"',
            insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            detail: '标签',
            range,
          })
        })
      }

      return { suggestions }
    },
  })
}

// 动态加载 Monaco Editor
const MonacoEditorLazy = React.lazy(() => import('@monaco-editor/react'))

export default function PromQLEditor({ value = '', onChange, height = 120, placeholder = '输入 PromQL 表达式...' }) {
  const editorRef = useRef(null)
  const monacoRef = useRef(null)

  const handleEditorDidMount = useCallback((editor, monaco) => {
    editorRef.current = editor
    monacoRef.current = monaco
    registerPromQL(monaco)
    monaco.editor.setTheme('promql-theme')

    // 单行模式下回车不换行
    if (height <= 40) {
      editor.addCommand(monaco.KeyCode.Enter, () => {})
    }
  }, [height])

  const handleChange = useCallback((val) => {
    if (onChange) onChange(val || '')
  }, [onChange])

  const editorOptions = useMemo(() => ({
    minimap: { enabled: false },
    lineNumbers: height > 80 ? 'on' : 'off',
    scrollBeyondLastLine: false,
    wordWrap: 'on',
    fontSize: 13,
    tabSize: 2,
    automaticLayout: true,
    suggestOnTriggerCharacters: true,
    quickSuggestions: true,
    scrollbar: { vertical: 'hidden', horizontal: 'auto' },
    overviewRulerLanes: 0,
    hideCursorInOverviewRuler: true,
    renderLineHighlight: 'none',
    folding: false,
    glyphMargin: false,
    padding: { top: 8, bottom: 8 },
    placeholder,
  }), [height, placeholder])

  return (
    <div className="fx-promql-editor" style={{ height, border: '1px solid var(--fx-border, #e5e7eb)', borderRadius: 6, overflow: 'hidden' }}>
      <React.Suspense fallback={<div style={{ padding: 12, color: '#9ca3af' }}>加载编辑器...</div>}>
        <MonacoEditorLazy
          height={height}
          language="promql"
          theme="promql-theme"
          value={value}
          onChange={handleChange}
          onMount={handleEditorDidMount}
          options={editorOptions}
        />
      </React.Suspense>
    </div>
  )
}
