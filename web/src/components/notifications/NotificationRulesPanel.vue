<template>
  <section class="notify-rules-panel">
    <header class="toolbar">
      <div>
        <div class="kicker">通知规则</div>
        <h3>规则列表</h3>
      </div>
      <div class="toolbar-actions">
        <el-input v-model="query" clearable placeholder="搜索规则" class="search" />
        <el-button :loading="loading" @click="load">刷新</el-button>
        <el-button type="primary" @click="$emit('blocked', 'rule-create', { mode: 'add' })">新增</el-button>
      </div>
    </header>
    <el-alert :title="blockedContracts['rule-list']" type="warning" show-icon :closable="false" class="state-alert" />
    <el-table v-loading="loading" :data="rows" border empty-text="BLOCKED_BY_CONTRACT：通知规则列表 contract 未暴露">
      <el-table-column label="名称" min-width="180" />
      <el-table-column label="通知配置/媒介" min-width="180" />
      <el-table-column label="用户组" min-width="160" />
      <el-table-column label="更新人" min-width="110" />
      <el-table-column label="更新时间" min-width="160" />
      <el-table-column label="启用" width="90" />
      <el-table-column label="操作" width="160">
        <template #default>
          <el-button link @click="$emit('blocked', 'rule-create', { mode: 'edit' })">编辑</el-button>
          <el-button link @click="$emit('blocked', 'rule-create', { mode: 'clone' })">克隆</el-button>
          <el-button link type="danger" @click="$emit('blocked', 'rule-create', { mode: 'delete' })">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-input :model-value="blockedPayload('rule-list', { query })" type="textarea" :rows="12" readonly spellcheck="false" class="json-view" />
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { blockedContracts, blockedPayload } from './notificationModel'

defineEmits(['blocked'])
const query = ref('')
const loading = ref(false)
const rows = ref([])
const load = () => { loading.value = false }
defineExpose({ load })
</script>

<style scoped>
.notify-rules-panel { padding: 16px; border: 1px solid #e3e8f1; border-radius: 8px; background: #fff; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.toolbar-actions { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h3 { margin: 4px 0 0; color: #17233c; font-size: 20px; }
.search { width: 240px; }
.state-alert { margin-bottom: 14px; }
.json-view { margin-top: 12px; }
@media (max-width: 760px) { .toolbar { align-items: flex-start; flex-direction: column; } .toolbar-actions, .search { width: 100%; } }
</style>
