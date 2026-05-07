<template>
  <el-drawer :model-value="visible" title="能力阻断" size="520px" @close="$emit('update:visible', false)">
    <el-alert :title="message" type="warning" show-icon :closable="false" />
    <el-input :model-value="payload" type="textarea" :rows="16" readonly spellcheck="false" class="json-view" />
  </el-drawer>
</template>

<script setup>
defineProps({
  visible: Boolean,
  message: { type: String, default: '' },
  payload: { type: String, default: '{}' },
})

defineEmits(['update:visible'])
</script>

<style scoped>
.json-view { margin-top: 12px; }
</style>
