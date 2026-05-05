<template>
  <section class="panel-card">
    <div class="section-head">
      <div><h3>通知渠道</h3><span>每个渠道可配置、可测试、可重试</span></div>
      <el-button size="small" @click="$emit('create')">新增</el-button>
    </div>
    <div class="panel-scroll">
      <div v-for="channel in channels" :key="channel.id" class="row-card channel-row">
        <div>
          <b>{{ channel.name }}</b>
          <span>{{ channelTypeLabel(channel.type) }}</span>
          <em>{{ displayEndpoint(channel) }}</em>
        </div>
        <div class="row-actions">
          <el-tag size="small" :type="channel.enabled ? 'success' : 'info'">{{ channel.enabled ? '已启用' : '待配置' }}</el-tag>
          <el-button size="small" link @click="$emit('edit', channel)">配置</el-button>
          <el-button size="small" link :disabled="!channel.enabled" @click="$emit('test', channel)">测试</el-button>
          <el-button size="small" link type="danger" :disabled="isBuiltinChannel(channel)" @click="$emit('remove', channel)">删除</el-button>
        </div>
      </div>
      <el-empty v-if="!channels.length" description="还没有通知渠道，请新增 Webhook、Flashduty 或 PagerDuty。" />
    </div>
  </section>
</template>

<script setup>
defineProps({
  channels: { type: Array, required: true },
  channelTypeLabel: { type: Function, required: true },
  displayEndpoint: { type: Function, required: true },
  isBuiltinChannel: { type: Function, required: true },
})

defineEmits(['create', 'edit', 'test', 'remove'])
</script>
