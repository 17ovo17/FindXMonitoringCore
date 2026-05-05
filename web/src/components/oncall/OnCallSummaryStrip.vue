<template>
  <div class="summary-strip">
    <div><b>{{ groupsCount }}</b><span>团队排班</span></div>
    <div><b>{{ enabledChannels }}</b><span>已启用渠道</span></div>
    <div><b>{{ enabledEscalation }}</b><span>已启用升级步骤</span></div>
    <div><b>{{ recordsCount }}</b><span>可追踪发送记录</span></div>
  </div>
</template>

<script setup>
defineProps({
  groupsCount: { type: Number, required: true },
  enabledChannels: { type: Number, required: true },
  enabledEscalation: { type: Number, required: true },
  recordsCount: { type: Number, required: true },
})
</script>
