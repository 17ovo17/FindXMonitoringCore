<template>
  <div ref="mountNode" class="findx-react-tracing-bridge"></div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import React from 'react'
import ReactDOM from 'react-dom'
import { TracingPage } from '../../react-shell/tracing/TracingPage.jsx'

const mountNode = ref(null)
const route = useRoute()
const router = useRouter()
const sections = ['overview', 'services', 'topology', 'traces', 'trace-detail', 'profiling', 'alarms', 'settings']
const agentFilters = ['q', 'package', 'runtime', 'status']

const cleanQuery = query => {
  const section = sections.includes(query?.section) ? String(query.section) : 'overview'
  const next = { section }
  ;['traceId', 'serviceId', 'instanceId', 'endpointId', 'layer', 'entity', 'depth', 'q'].forEach(key => {
    if (query?.[key]) next[key] = String(query[key])
  })
  return next
}

const cleanParams = params => ({
  traceId: params?.traceId ? String(params.traceId) : '',
})

const cleanAgentQuery = query => {
  const next = { section: 'hosts' }
  agentFilters.forEach(key => {
    if (query?.[key]) next[key] = String(query[key])
  })
  return next
}

const onNavigate = query => {
  if (query?.action === 'agent-hosts') {
    router.replace({ path: '/agents', query: cleanAgentQuery(query) })
    return
  }
  const next = cleanQuery(query)
  if (next.section === 'trace-detail' && next.traceId) {
    router.replace({ path: `/tracing/${encodeURIComponent(next.traceId)}`, query: { section: 'trace-detail' } })
    return
  }
  router.replace({ path: '/tracing', query: next })
}

const renderReact = () => {
  if (!mountNode.value) return
  ReactDOM.render(
    React.createElement(TracingPage, {
      query: cleanQuery(route.query),
      params: cleanParams(route.params),
      onNavigate,
    }),
    mountNode.value,
  )
}

onMounted(() => nextTick(renderReact))
watch(() => [route.query, route.params], renderReact, { deep: true })

onBeforeUnmount(() => {
  if (mountNode.value) ReactDOM.unmountComponentAtNode(mountNode.value)
})
</script>

<style scoped>
.findx-react-tracing-bridge {
  min-width: 0;
}
</style>
