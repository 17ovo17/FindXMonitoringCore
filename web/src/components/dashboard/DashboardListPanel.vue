<template>
  <section class="dash-layout">
    <aside class="group-sidebar">
      <div class="sidebar-title">仪表盘</div>
      <el-button text :type="scopeFilter === 'all' ? 'primary' : ''" @click="$emit('set-scope', 'all')">全部仪表盘</el-button>
      <el-button text :type="scopeFilter === 'public' ? 'primary' : ''" @click="$emit('set-scope', 'public')">公开仪表盘</el-button>
      <el-divider />
      <div class="sidebar-subtitle">业务组</div>
      <el-button v-for="group in businessGroups" :key="group.key" text :type="scopeFilter === group.key ? 'primary' : ''" @click="$emit('set-scope', group.key)">
        <span>{{ group.label }}</span>
        <span class="count">{{ group.count }}</span>
      </el-button>
    </aside>

    <main class="work-area">
      <header class="toolbar">
        <div>
          <div class="kicker">FindX Dashboard</div>
          <h2>仪表盘列表</h2>
        </div>
        <div class="toolbar-actions">
          <el-button :loading="loading" @click="$emit('refresh')">刷新</el-button>
          <el-input :model-value="keyword" class="search-input" clearable placeholder="搜索名称、标签、备注" @update:model-value="$emit('update:keyword', $event)" @keyup.enter="$emit('refresh')" />
          <el-button type="primary" @click="$emit('create')">新增</el-button>
          <el-button @click="$emit('templates')">导入</el-button>
          <el-dropdown trigger="click" @command="$emit('batch-command', $event)">
            <el-button>更多批量</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="share">批量公开配置</el-dropdown-item>
                <el-dropdown-item command="export">批量导出</el-dropdown-item>
                <el-dropdown-item command="delete" divided>批量删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-dropdown trigger="click" :hide-on-click="false">
            <el-button>列设置</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="col in columnOptions" :key="col.key">
                  <el-checkbox v-model="visibleColumns[col.key]">{{ col.label }}</el-checkbox>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <el-alert v-if="errorText" :title="errorTitle" :description="errorText" :type="permissionError ? 'warning' : 'error'" show-icon :closable="false" class="state-alert" />
      <el-table v-loading="loading" :data="filteredDashboards" border class="dash-table" empty-text="暂无真实仪表盘数据" @selection-change="$emit('update:selected-rows', $event)">
        <el-table-column type="selection" width="42" />
        <el-table-column v-if="visibleColumns.name" label="名称" min-width="190" fixed>
          <template #default="{ row }">
            <el-button link type="primary" @click="$emit('open-detail', row.id)">{{ row.name }}</el-button>
            <div class="ident">{{ row.ident || row.id }}</div>
          </template>
        </el-table-column>
        <el-table-column v-if="visibleColumns.tags" label="标签" min-width="160">
          <template #default="{ row }">
            <el-tag v-for="tag in row.tags" :key="tag" size="small" class="tag">{{ tag }}</el-tag>
            <span v-if="row.tags.length === 0" class="muted">无</span>
          </template>
        </el-table-column>
        <el-table-column v-if="visibleColumns.note" prop="note" label="备注" min-width="220" show-overflow-tooltip />
        <el-table-column v-if="visibleColumns.updatedAt" prop="updatedAt" label="更新时间" min-width="150" />
        <el-table-column v-if="visibleColumns.updatedBy" prop="updatedBy" label="更新人" min-width="120" />
        <el-table-column v-if="visibleColumns.share" label="共享状态" width="120">
          <template #default="{ row }"><el-tag :type="row.shared ? 'success' : 'info'" size="small">{{ row.shareText }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="88" fixed="right">
          <template #default="{ row }">
            <el-dropdown trigger="click" @command="cmd => $emit('row-command', cmd, row)">
              <el-button text>更多</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="edit">编辑</el-dropdown-item>
                  <el-dropdown-item command="clone">克隆</el-dropdown-item>
                  <el-dropdown-item command="export">导出</el-dropdown-item>
                  <el-dropdown-item command="share">分享/公开配置</el-dropdown-item>
                  <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </main>
  </section>
</template>

<script setup>
defineProps({
  loading: Boolean,
  keyword: String,
  scopeFilter: String,
  errorTitle: String,
  errorText: String,
  permissionError: Boolean,
  businessGroups: { type: Array, default: () => [] },
  filteredDashboards: { type: Array, default: () => [] },
  columnOptions: { type: Array, default: () => [] },
  visibleColumns: { type: Object, required: true },
})

defineEmits(['update:keyword', 'update:selected-rows', 'set-scope', 'refresh', 'create', 'templates', 'batch-command', 'row-command', 'open-detail'])
</script>

<style scoped>
.dash-layout { display: grid; grid-template-columns: 220px minmax(0, 1fr); gap: 16px; min-height: calc(100vh - 92px); }
.group-sidebar, .work-area { border: 1px solid #e3e8f1; border-radius: 8px; background: #fff; }
.group-sidebar { padding: 14px; }
.group-sidebar .el-button { justify-content: space-between; width: 100%; margin: 2px 0; }
.sidebar-title { margin-bottom: 12px; font-size: 18px; font-weight: 700; }
.sidebar-subtitle, .kicker, .ident { color: #63718a; font-size: 12px; }
.count, .muted, .ident { color: #8a96aa; }
.work-area { padding: 16px; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.toolbar-actions { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
h2 { margin: 0; color: #17233c; font-size: 22px; }
.search-input { width: 260px; }
.state-alert { margin-bottom: 14px; }
.dash-table { width: 100%; }
.tag { margin: 2px 4px 2px 0; }
@media (max-width: 900px) {
  .dash-layout { grid-template-columns: 1fr; }
  .toolbar { align-items: flex-start; flex-direction: column; }
  .search-input { width: 100%; }
}
</style>
