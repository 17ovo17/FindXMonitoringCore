<template>
  <div ref="mountNode" class="findx-react-agent-bridge"></div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import React from 'react'
import ReactDOM from 'react-dom'
import { AgentPage } from '../../react-shell/agents/AgentPage.jsx'

const mountNode = ref(null)
const route = useRoute()
const router = useRouter()
const sections = ['overview', 'hosts', 'packages']
const hostMergedSections = ['install', 'templates', 'heartbeat', 'data-arrival', 'config', 'plugins']
const preservedFilters = ['q', 'os', 'runtime', 'status', 'package']

const queryValue = value => (Array.isArray(value) ? value[0] : value)

const cleanQuery = query => {
  const rawSection = String(queryValue(query?.section) || '')
  const rawLegacySection = String(queryValue(query?.legacySection) || '')
  const section = hostMergedSections.includes(rawSection)
    ? 'hosts'
    : sections.includes(rawSection) ? rawSection : 'hosts'
  const next = { section }
  if (hostMergedSections.includes(rawSection)) next.legacySection = rawSection
  else if (hostMergedSections.includes(rawLegacySection)) next.legacySection = rawLegacySection
  preservedFilters.forEach(key => {
    if (Object.prototype.hasOwnProperty.call(query || {}, key)) {
      const value = queryValue(query[key])
      if (value !== undefined && value !== null) next[key] = String(value)
    }
  })
  return next
}

const legacyFocusQuery = query => {
  const rawSection = String(queryValue(query?.section) || '')
  const rawLegacySection = String(queryValue(query?.legacySection) || '')
  const next = cleanQuery(query)
  if (hostMergedSections.includes(rawSection)) return { ...next, legacySection: rawSection }
  if (hostMergedSections.includes(rawLegacySection)) return { ...next, legacySection: rawLegacySection }
  return next
}

const sameQuery = (left, right) => {
  const leftKeys = Object.keys(left || {})
  const rightKeys = Object.keys(right || {})
  if (leftKeys.length !== rightKeys.length) return false
  return rightKeys.every(key => String(queryValue(left?.[key]) ?? '') === String(right[key] ?? ''))
}

const normalizeRouteQuery = () => {
  const nextQuery = cleanQuery(route.query)
  if (route.path !== '/agents' || !sameQuery(route.query, nextQuery)) {
    router.replace({ path: '/agents', query: nextQuery })
  }
  return nextQuery
}

const onNavigate = query => {
  router.replace({ path: '/agents', query: cleanQuery(query) })
}

const renderReact = () => {
  if (!mountNode.value) return
  const query = legacyFocusQuery(route.query)
  normalizeRouteQuery()
  ReactDOM.render(
    React.createElement(AgentPage, {
      query,
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
.findx-react-agent-bridge {
  min-width: 0;
}
</style>
