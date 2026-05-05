<template>
  <div class="page-head">
    <div>
      <div class="panel-kicker">On-call & Notifications</div>
      <h2>团队排班与通知策略</h2>
      <p>团队排班、通知渠道、升级策略和发送记录都走真实配置链路；测试发送会写入后端记录并可重试追踪。</p>
    </div>
    <div class="head-actions">
      <el-button :loading="loading" @click="$emit('reload')">刷新</el-button>
      <el-button type="primary" @click="$emit('createGroup')">新增团队排班</el-button>
      <el-button type="primary" plain @click="$emit('createChannel')">新增渠道</el-button>
    </div>
  </div>
</template>

<script setup>
defineProps({
  loading: { type: Boolean, required: true },
})

defineEmits(['reload', 'createGroup', 'createChannel'])
</script>
