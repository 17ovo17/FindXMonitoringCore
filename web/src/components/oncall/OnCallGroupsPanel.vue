<template>
  <section class="panel-card">
    <div class="section-head">
      <div><h3>团队排班</h3><span>复用团队机制配置负责人、时间窗口和升级角色</span></div>
      <el-button size="small" @click="$emit('create')">新增</el-button>
    </div>
    <div class="panel-scroll">
      <div v-for="group in groups" :key="group.id" class="row-card">
        <div>
          <b>{{ group.name }}</b>
          <span>{{ formatMembers(group.members) }}</span>
          <em>{{ group.schedule || '未配置时间窗口' }} · {{ group.role || 'primary' }}</em>
        </div>
        <div class="row-actions">
          <el-tag size="small" :type="group.enabled ? 'success' : 'info'">{{ group.enabled ? '启用' : '停用' }}</el-tag>
          <el-button size="small" link @click="$emit('edit', group)">编辑</el-button>
          <el-button size="small" link type="danger" @click="$emit('remove', group)">删除</el-button>
        </div>
      </div>
      <el-empty v-if="!groups.length" description="还没有团队排班，请先配置主值团队。" />
    </div>
  </section>
</template>

<script setup>
defineProps({
  groups: { type: Array, required: true },
  formatMembers: { type: Function, required: true },
})

defineEmits(['create', 'edit', 'remove'])
</script>
