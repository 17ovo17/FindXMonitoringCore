<template>
  <el-drawer :model-value="visible" :title="title" size="620px" @close="$emit('update:visible', false)">
    <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" class="state-alert" />
    <el-empty v-if="!event" description="请选择事件" />
    <template v-else>
      <el-descriptions :column="1" border size="small">
        <el-descriptions-item label="事件名称">{{ event.name }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="eventStatusTag(event.status)" size="small">{{ event.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="级别">
          <el-tag :type="severityTag(event.severity)" size="small">{{ severityText(event.severity) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="数据源">{{ event.datasourceId }}</el-descriptions-item>
        <el-descriptions-item label="对象">{{ event.target }}</el-descriptions-item>
        <el-descriptions-item label="当前值">{{ event.value || '-' }}</el-descriptions-item>
        <el-descriptions-item label="指纹">{{ event.fingerprint || '-' }}</el-descriptions-item>
        <el-descriptions-item label="首次触发">{{ formatDate(event.firstSeen) }}</el-descriptions-item>
        <el-descriptions-item label="最近触发">{{ formatDate(event.lastSeen) }}</el-descriptions-item>
      </el-descriptions>

      <div class="drawer-actions">
        <el-button size="small" @click="$emit('action', event, 'ack')">ack</el-button>
        <el-button size="small" @click="$emit('assign', event)">assign</el-button>
        <el-button size="small" type="success" plain @click="$emit('action', event, 'resolve')">resolve</el-button>
        <el-button size="small" type="warning" plain @click="$emit('action', event, 'archive')">archive</el-button>
      </div>

      <h4>处理记录</h4>
      <el-timeline v-if="event.actionLog.length">
        <el-timeline-item v-for="item in event.actionLog" :key="item.id || item.created_at || item.createdAt" :timestamp="formatDate(item.created_at || item.createdAt)">
          <b>{{ item.action || '-' }}</b>
          <span>{{ item.actor || '-' }}</span>
          <span v-if="item.from || item.to"> {{ item.from || '-' }} -> {{ item.to || '-' }}</span>
          <div class="minor">{{ item.assignee || '' }} {{ item.reason || '' }}</div>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-else description="暂无处理记录" :image-size="64" />

      <h4>标签</h4>
      <pre>{{ safeJson(event.labels) }}</pre>
      <h4>注解</h4>
      <pre>{{ safeJson(event.annotations) }}</pre>
    </template>
  </el-drawer>
</template>

<script setup>
import { computed } from 'vue'
import { safeJson } from '../../api/alerting'
import { eventStatusTag, formatDate, severityTag, severityText } from './alertingModel'

const props = defineProps({
  visible: Boolean,
  event: { type: Object, default: null },
  error: { type: String, default: '' },
})

defineEmits(['update:visible', 'action', 'assign'])

const title = computed(() => props.event?.name ? `事件详情：${props.event.name}` : '事件详情')
</script>

<style scoped>
.state-alert { margin-bottom: 12px; }
.drawer-actions { display: flex; flex-wrap: wrap; gap: 8px; margin: 14px 0; }
.minor { color: #7a879c; font-size: 12px; margin-top: 4px; }
pre { max-height: 190px; overflow: auto; padding: 10px; border-radius: 8px; background: #f5f7fb; white-space: pre-wrap; word-break: break-word; font-size: 12px; }
h4 { margin: 16px 0 8px; color: #22314d; }
</style>
