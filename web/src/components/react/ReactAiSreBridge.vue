<template>
  <div ref="mountNode" class="findx-react-ai-sre-bridge"></div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import React from 'react'
import ReactDOM from 'react-dom'
import { AiSrePage } from '../../react-shell/ai-sre/AiSrePage.jsx'

const mountNode = ref(null)
const route = useRoute()
const router = useRouter()
const sections = ['diagnosis', 'workflow', 'health', 'report', 'evidence', 'knowledge', 'remediation']
const aliases = { inspect: 'report', reports: 'report', chains: 'evidence', fix: 'remediation' }

const normalizeSection = value => {
  const raw = String(value || 'diagnosis')
  const next = aliases[raw] || raw
  return sections.includes(next) ? next : 'diagnosis'
}

const cleanQuery = query => {
  const next = { section: normalizeSection(query?.section) }
  ;['sessionId', 'workflowId', 'inspectionId', 'caseId', 'q'].forEach(key => {
    if (query?.[key]) next[key] = String(query[key])
  })
  return next
}

const onNavigate = query => {
  router.replace({ path: '/aiops', query: cleanQuery(query) })
}

const renderReact = () => {
  if (!mountNode.value) return
  ReactDOM.render(
    React.createElement(AiSrePage, {
      query: cleanQuery(route.query),
      onNavigate,
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
.findx-react-ai-sre-bridge {
  min-width: 0;
}
</style>
