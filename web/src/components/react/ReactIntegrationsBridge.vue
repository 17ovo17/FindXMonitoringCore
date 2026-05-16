<template>
  <div ref="mountNode" class="findx-react-integrations-bridge"></div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import React from 'react'
import ReactDOM from 'react-dom'
import { IntegrationsPage } from '../../react-shell/base-monitoring/integrations/IntegrationsPage.jsx'

const mountNode = ref(null)
const route = useRoute()
const router = useRouter()

const cleanQuery = (query) => {
  const next = {}
  ;['section', 'component', 'tab'].forEach((key) => {
    if (query?.[key]) next[key] = String(query[key])
  })
  return next
}

const onNavigate = (query) => {
  router.replace({ path: '/integrations', query: cleanQuery(query) })
}

const renderReact = () => {
  if (!mountNode.value) return
  ReactDOM.render(
    React.createElement(IntegrationsPage, {
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
.findx-react-integrations-bridge {
  min-width: 0;
}
</style>
