<template>
  <div ref="mountNode" class="findx-react-logs-bridge"></div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import React from 'react'
import ReactDOM from 'react-dom'
import { LogsPage } from '../../react-shell/logs/LogsPage.jsx'

const mountNode = ref(null)
const route = useRoute()
const router = useRouter()
const sectionAliases = { explorer: 'query', tail: 'live', 'live-tail': 'live', fields: 'fields', pipeline: 'pipelines', views: 'saved-views' }
const sections = ['query', 'live', 'fields', 'context', 'aggregate', 'pipelines', 'saved-views', 'trace-link']

const normalizeSection = value => {
  const raw = String(value || 'query')
  const next = sectionAliases[raw] || raw
  return sections.includes(next) ? next : 'query'
}

const cleanQuery = query => {
  const next = { section: normalizeSection(query?.section) }
  ;['q', 'traceId', 'spanId', 'service', 'scope', 'source', 'level', 'field', 'viewId', 'pipelineId', 'logId'].forEach(key => {
    if (query?.[key]) next[key] = String(query[key])
  })
  return next
}

const onNavigate = query => {
  router.replace({ path: '/logs', query: cleanQuery(query) })
}

const onOpenTrace = traceId => {
  router.replace({ path: `/tracing/${encodeURIComponent(traceId)}`, query: { section: 'trace-detail' } })
}

const onOpenAgent = filters => {
  const next = { section: 'hosts', package: filters?.packageName || 'logs' }
  ;['q', 'runtime', 'status'].forEach(key => {
    if (filters?.[key]) next[key] = String(filters[key])
  })
  router.replace({ path: '/agents', query: next })
}

const renderReact = () => {
  if (!mountNode.value) return
  ReactDOM.render(
    React.createElement(LogsPage, {
      query: cleanQuery(route.query),
      onNavigate,
      onOpenTrace,
      onOpenAgent,
    }),
    mountNode.value,
  )
}

onMounted(() => nextTick(renderReact))
watch(() => route.query, renderReact, { deep: true })

onBeforeUnmount(() => {
  if (mountNode.value) ReactDOM.unmountComponentAtNode(mountNode.value)
})
</script>

<style scoped>
.findx-react-logs-bridge {
  min-width: 0;
}
</style>
