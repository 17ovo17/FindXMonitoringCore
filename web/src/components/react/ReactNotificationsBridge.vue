<template>
  <div ref="mountNode" class="findx-react-notifications-bridge"></div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import React from 'react'
import ReactDOM from 'react-dom'
import { NotificationsPage } from '../../react-shell/base-monitoring/notifications/NotificationsPage.jsx'

const mountNode = ref(null)
const route = useRoute()
const router = useRouter()

const cleanQuery = (query) => {
  const section = ['rules', 'channels', 'templates'].includes(query?.section) ? String(query.section) : 'rules'
  const next = { section }
  ;['id', 'mode', 'type'].forEach((key) => {
    if (query?.[key]) next[key] = String(query[key])
  })
  return next
}

const onNavigate = (query) => {
  router.replace({ path: '/notifications', query: cleanQuery(query) })
}

const renderReact = () => {
  if (!mountNode.value) return
  ReactDOM.render(
    React.createElement(NotificationsPage, {
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
.findx-react-notifications-bridge {
  min-width: 0;
}
</style>
