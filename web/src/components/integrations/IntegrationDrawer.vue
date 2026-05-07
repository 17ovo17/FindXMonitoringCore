<template>
  <el-drawer :model-value="visible" size="88%" :with-header="false" @update:model-value="$emit('update:visible', $event)">
    <header class="drawer-head">
      <div class="title">
        <div class="logo">{{ component?.logo || 'FX' }}</div>
        <div>
          <h2>{{ component?.ident || '组件详情' }}</h2>
          <p>{{ component?.readme || '暂无说明' }}</p>
        </div>
      </div>
      <el-button text @click="$emit('update:visible', false)">关闭</el-button>
    </header>
    <el-tabs :model-value="activeTab" @update:model-value="$emit('update:activeTab', $event)">
      <el-tab-pane v-for="tab in tabs" :key="tab.key" :name="tab.key" :label="tab.label">
        <div v-if="tab.key === 'instructions'" class="instructions">
          <el-input :model-value="component?.readme || ''" type="textarea" :rows="14" readonly />
          <el-alert title="BLOCKED_BY_CONTRACT：使用说明编辑、组件 logo 编辑和组件停用 contract 未完整暴露。" type="warning" show-icon :closable="false" class="state-alert" />
        </div>
        <IntegrationPayloadTable
          v-else-if="tab.key === 'dashboard'"
          :query="query"
          :payloads="dashboardPayloads"
          @update:query="$emit('update:query', $event)"
          @select="$emit('select', $event)"
          @create="$emit('blocked', 'create')"
          @batch-import="$emit('batch-import')"
          @batch-export="$emit('batch-export')"
          @preview="$emit('preview', $event)"
          @tag-search="$emit('tag-search', $event)"
          @row-command="(cmd, row) => $emit('row-command', cmd, row)"
        />
        <el-alert v-else :title="blockedContracts[tab.key]" type="warning" show-icon :closable="false" class="state-alert" />
      </el-tab-pane>
    </el-tabs>
  </el-drawer>
</template>

<script setup>
import IntegrationPayloadTable from './IntegrationPayloadTable.vue'
import { blockedContracts } from './integrationModel'

defineProps({
  visible: Boolean,
  activeTab: String,
  query: String,
  component: Object,
  tabs: { type: Array, default: () => [] },
  dashboardPayloads: { type: Array, default: () => [] },
})

defineEmits(['update:visible', 'update:activeTab', 'update:query', 'select', 'blocked', 'batch-import', 'batch-export', 'preview', 'tag-search', 'row-command'])
</script>

<style scoped>
.drawer-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 12px; }
.title { display: flex; align-items: flex-start; gap: 12px; min-width: 0; }
.logo { display: grid; place-items: center; flex: 0 0 auto; width: 42px; height: 42px; border-radius: 8px; background: #eaf2ff; color: #1769ff; font-weight: 800; }
h2, p { margin: 0; }
h2 { color: #17233c; font-size: 20px; }
p { margin-top: 6px; color: #66758d; font-size: 13px; line-height: 1.6; }
.state-alert { margin-top: 12px; }
</style>
