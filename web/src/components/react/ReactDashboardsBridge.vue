<template>
  <div ref="mountNode" class="findx-react-dashboards-bridge"></div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import React from 'react'
import ReactDOM from 'react-dom'
import { DashboardsPage } from '../../react-shell/base-monitoring/dashboards/DashboardsPage.jsx'

const mountNode = ref(null)
const route = useRoute()
const router = useRouter()

const cleanQuery = (query) => {
  const next = {}
  ;['section', 'id'].forEach((key) => {
    if (query?.[key]) next[key] = String(query[key])
  })
  return next
}

const onNavigate = (query) => {
  router.replace({ path: '/dashboards', query: cleanQuery(query) })
}

const renderReact = () => {
  if (!mountNode.value) return
  ReactDOM.render(
    React.createElement(DashboardsPage, {
      query: cleanQuery(route.query),
      onNavigate,
    }),
    mountNode.value,
  )
}

onMounted(() => {
  nextTick(renderReact)
})

watch(() => route.query, renderReact, { deep: true })

onBeforeUnmount(() => {
  if (mountNode.value) ReactDOM.unmountComponentAtNode(mountNode.value)
})
</script>

<style scoped>
.findx-react-dashboards-bridge {
  min-width: 0;
}
</style>
