<template>
  <section class="panel-card">
    <div class="section-head">
      <div><h3>升级策略</h3><span>按未确认、未恢复等条件逐级通知</span></div>
      <el-button size="small" @click="$emit('create')">新增</el-button>
    </div>
    <div class="panel-scroll">
      <el-timeline>
        <el-timeline-item v-for="step in escalation" :key="step.id" :timestamp="`${step.delay_min || 0} 分钟`">
          <div class="timeline-card">
            <b>{{ step.action }}</b>
            <span>{{ step.condition || '告警触发' }} → {{ step.target }}</span>
            <div>
              <el-tag size="small" :type="step.enabled ? 'success' : 'info'">{{ step.enabled ? '启用' : '停用' }}</el-tag>
              <el-button size="small" link @click="$emit('edit', step)">编辑</el-button>
              <el-button size="small" link type="danger" @click="$emit('remove', step)">删除</el-button>
            </div>
          </div>
        </el-timeline-item>
      </el-timeline>
      <el-empty v-if="!escalation.length" description="还没有升级步骤。" />
    </div>
  </section>
</template>

<script setup>
defineProps({
  escalation: { type: Array, required: true },
})

defineEmits(['create', 'edit', 'remove'])
</script>
