<template>
  <section class="payload-section">
    <div class="payload-toolbar">
      <el-input :model-value="query" class="search" clearable placeholder="搜索模板名称、标签、说明" @update:model-value="$emit('update:query', $event)" />
      <div class="actions">
        <el-button @click="$emit('create')">新增</el-button>
        <el-button @click="$emit('batch-import')">导入到业务组</el-button>
        <el-button @click="$emit('batch-export')">批量导出</el-button>
      </div>
    </div>
    <el-table :data="payloads" border class="payload-table" empty-text="暂无模板数据" @selection-change="$emit('select', $event)">
      <el-table-column type="selection" width="42" />
      <el-table-column label="名称" min-width="210" fixed>
        <template #default="{ row }">
          <el-button link type="primary" @click="$emit('preview', row)">{{ row.name }}</el-button>
        </template>
      </el-table-column>
      <el-table-column label="标签" min-width="180">
        <template #default="{ row }">
          <el-tag v-for="tag in row.tags" :key="tag" size="small" class="tag" @click="$emit('tag-search', tag)">{{ tag }}</el-tag>
          <span v-if="row.tags.length === 0" class="muted">无</span>
        </template>
      </el-table-column>
      <el-table-column prop="note" label="说明" min-width="260" show-overflow-tooltip />
      <el-table-column prop="updatedBy" label="更新人" width="120">
        <template #default="{ row }"><el-tag v-if="row.updatedBy === 'system'" size="small">系统内置</el-tag><span v-else>{{ row.updatedBy || '-' }}</span></template>
      </el-table-column>
      <el-table-column label="操作" width="88" fixed="right">
        <template #default="{ row }">
          <el-dropdown trigger="click" @command="cmd => $emit('row-command', cmd, row)">
            <el-button text>更多</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="import">导入到业务组</el-dropdown-item>
                <el-dropdown-item command="export">导出</el-dropdown-item>
                <el-dropdown-item command="edit">编辑</el-dropdown-item>
                <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>
  </section>
</template>

<script setup>
defineProps({
  query: String,
  payloads: { type: Array, default: () => [] },
})

defineEmits(['update:query', 'select', 'create', 'batch-import', 'batch-export', 'preview', 'tag-search', 'row-command'])
</script>

<style scoped>
.payload-toolbar { display: flex; justify-content: space-between; gap: 12px; margin-bottom: 12px; }
.search { width: 300px; }
.actions { display: flex; gap: 8px; flex-wrap: wrap; }
.payload-table { width: 100%; }
.tag { margin: 2px 4px 2px 0; cursor: pointer; }
.muted { color: #8a96aa; }
@media (max-width: 760px) {
  .payload-toolbar { align-items: flex-start; flex-direction: column; }
  .search { width: 100%; }
}
</style>
