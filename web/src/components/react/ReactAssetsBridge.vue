<template>
  <div ref="mountNode" class="findx-react-assets-bridge"></div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import React from 'react'
import ReactDOM from 'react-dom'
import { AssetsPage } from '../../react-shell/cmdb/AssetsPage.jsx'

const mountNode = ref(null)
const route = useRoute()
const router = useRouter()
const sections = ['overview', 'business', 'hosts', 'resource-groups', 'agents', 'cmdb']

const cleanQuery = query => {
  const section = sections.includes(query?.section) ? String(query.section) : 'overview'
  const next = { section }
  ;['id', 'q', 'group', 'workspace', 'online', 'tag'].forEach(key => {
    if (query?.[key]) next[key] = String(query[key])
  })
  return next
}

const onNavigate = query => {
  router.replace({ path: '/assets', query: cleanQuery(query) })
}

const renderReact = () => {
  if (!mountNode.value) return
  ReactDOM.render(
    React.createElement(AssetsPage, {
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
.findx-react-assets-bridge {
  min-width: 0;
}
</style>
