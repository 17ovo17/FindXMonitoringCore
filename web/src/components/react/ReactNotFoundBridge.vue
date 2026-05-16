<template>
  <div ref="mountNode" class="findx-react-notfound-bridge"></div>
</template>

<script setup>
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import React from 'react'
import ReactDOM from 'react-dom'
import { NotFoundPage } from '../../react-shell/system/NotFoundPage.jsx'

const mountNode = ref(null)
const route = useRoute()
const router = useRouter()

const currentPath = () => route.fullPath || route.path || '/'

const onNavigate = ({ path, query = {} }) => {
  router.push({ path, query })
}

const renderReact = () => {
  if (!mountNode.value) return
  ReactDOM.render(
    React.createElement(NotFoundPage, {
      path: currentPath(),
      onNavigate,
    }),
    mountNode.value,
  )
}

onMounted(() => {
  nextTick(renderReact)
})

watch(() => route.fullPath, renderReact)

onBeforeUnmount(() => {
  if (mountNode.value) ReactDOM.unmountComponentAtNode(mountNode.value)
})
</script>

<style scoped>
.findx-react-notfound-bridge {
  min-width: 0;
}
</style>
