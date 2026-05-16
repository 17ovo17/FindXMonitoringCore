<template>
  <div ref="mountNode" class="findx-react-org-bridge"></div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import React from 'react'
import ReactDOM from 'react-dom'
import { OrgPage } from '../../react-shell/org/OrgPage.jsx'

const mountNode = ref(null)
const route = useRoute()
const router = useRouter()
const sections = ['users', 'teams', 'business', 'roles']

const cleanQuery = (query) => {
  const section = sections.includes(query?.section) ? String(query.section) : 'users'
  const next = { section }
  ;['id', 'mode', 'q'].forEach((key) => {
    if (query?.[key]) next[key] = String(query[key])
  })
  return next
}

const onNavigate = (query) => {
  router.replace({ path: '/org', query: cleanQuery(query) })
}

const renderReact = () => {
  if (!mountNode.value) return
  ReactDOM.render(
    React.createElement(OrgPage, {
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
.findx-react-org-bridge {
  min-width: 0;
}
</style>
