<template>
  <section class="panel-card">
    <div class="section-head">
      <div><h3>发送记录</h3><span>测试通知和重试会生成可追踪记录</span></div>
      <el-button size="small" :loading="recordsLoading" @click="$emit('reload')">刷新记录</el-button>
    </div>
    <div class="panel-scroll">
      <div v-for="item in records" :key="item.id" class="record-row">
        <div>
          <b>{{ item.channel }}</b>
          <span>{{ item.receiver || '未指定接收人' }}</span>
          <em>{{ formatTime(item.created_at) }}</em>
        </div>
        <div class="record-detail">{{ item.detail || item.trace_id }}</div>
        <div class="row-actions">
          <el-tag size="small" :type="item.status === 'success' ? 'success' : 'danger'">{{ item.status }}</el-tag>
          <el-button size="small" link @click="$emit('retry', item)">重试</el-button>
        </div>
      </div>
      <el-empty v-if="!records.length" description="暂无发送记录。配置渠道后点击测试，会生成可追踪记录。" />
    </div>
  </section>
</template>

<script setup>
defineProps({
  records: { type: Array, required: true },
  recordsLoading: { type: Boolean, required: true },
  formatTime: { type: Function, required: true },
})

defineEmits(['reload', 'retry'])
</script>
