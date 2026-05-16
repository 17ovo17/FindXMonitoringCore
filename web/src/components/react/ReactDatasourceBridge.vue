<template>
  <div ref="mountNode" class="findx-react-datasource-bridge"></div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import React from 'react'
import ReactDOM from 'react-dom'
import { DatasourcePage } from '../../react-shell/base-monitoring/datasource/DatasourcePage.jsx'

const mountNode = ref(null)

onMounted(() => {
  if (mountNode.value) {
    ReactDOM.render(React.createElement(DatasourcePage), mountNode.value)
  }
})

onBeforeUnmount(() => {
  if (mountNode.value) {
    ReactDOM.unmountComponentAtNode(mountNode.value)
  }
})
</script>

<style scoped>
.findx-react-datasource-bridge {
  min-width: 0;
}
</style>
