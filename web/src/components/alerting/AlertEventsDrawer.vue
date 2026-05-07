<template>
  <el-drawer :model-value="visible" :title="title" size="82%" @close="$emit('update:visible', false)">
    <AlertEventsPanel ref="panelRef" :rule-id="ruleId" :datasources="datasources" @blocked="(action, ctx) => $emit('blocked', action, ctx)" />
  </el-drawer>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import AlertEventsPanel from './AlertEventsPanel.vue'

const props = defineProps({
  visible: Boolean,
  rule: { type: Object, default: null },
  datasources: { type: Array, default: () => [] },
})

defineEmits(['update:visible', 'blocked'])

const panelRef = ref(null)
const ruleId = computed(() => props.rule?.id || '')
const title = computed(() => props.rule?.name ? `规则事件：${props.rule.name}` : '规则事件')

watch(() => props.visible, value => {
  if (value) setTimeout(() => panelRef.value?.load?.(), 0)
})
</script>
