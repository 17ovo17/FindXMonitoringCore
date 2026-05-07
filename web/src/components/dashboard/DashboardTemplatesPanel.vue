<template>
  <section class="templates-area">
    <header class="toolbar">
      <div>
        <div class="kicker">FindX Templates</div>
        <h2>模板中心</h2>
      </div>
      <div class="toolbar-actions">
        <el-button @click="$emit('go-list')">返回列表</el-button>
        <el-button :loading="templatesLoading" @click="$emit('load-templates')">刷新模板</el-button>
      </div>
    </header>
    <el-alert v-if="templatesError" :title="templatesError" type="error" show-icon :closable="false" class="state-alert" />
    <div class="template-grid">
      <article v-for="item in templates" :key="item.id" class="template-card">
        <div class="template-icon">{{ item.icon }}</div>
        <div>
          <h3>{{ item.name }}</h3>
          <p>{{ item.description || '暂无说明' }}</p>
          <div class="template-meta">
            <el-tag size="small">{{ item.kind }}</el-tag>
            <el-tag size="small" type="info">{{ item.panelCount }} Panels</el-tag>
            <el-tag size="small" type="warning">{{ item.variableCount }} Variables</el-tag>
          </div>
        </div>
        <div class="template-actions">
          <el-button @click="$emit('preview-template', item)">预览</el-button>
          <el-button type="primary" @click="$emit('open-import', item)">导入</el-button>
        </div>
      </article>
      <el-empty v-if="!templatesLoading && templates.length === 0" description="暂无可导入模板" />
    </div>
  </section>
</template>

<script setup>
defineProps({
  templatesLoading: Boolean,
  templatesError: String,
  templates: { type: Array, default: () => [] },
})

defineEmits(['go-list', 'load-templates', 'preview-template', 'open-import'])
</script>

<style scoped>
.templates-area { padding: 16px; border: 1px solid #e3e8f1; border-radius: 8px; background: #fff; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.toolbar-actions { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.kicker { color: #63718a; font-size: 12px; }
h2, h3, p { margin: 0; }
h2 { color: #17233c; font-size: 22px; }
h3 { color: #17233c; font-size: 16px; }
p { color: #66758d; font-size: 13px; line-height: 1.6; }
.state-alert { margin-bottom: 14px; }
.template-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 310px), 1fr)); gap: 12px; }
.template-card { display: grid; grid-template-columns: 48px minmax(0, 1fr); min-width: 0; gap: 12px; padding: 14px; border: 1px solid #e1e8f2; border-radius: 8px; background: #fff; }
.template-icon { display: grid; place-items: center; width: 42px; height: 42px; border-radius: 8px; background: #eaf2ff; color: #1769ff; font-weight: 800; }
.template-meta { display: flex; gap: 6px; flex-wrap: wrap; margin-top: 10px; }
.template-actions { grid-column: 2; display: flex; justify-content: flex-end; gap: 8px; flex-wrap: wrap; }
@media (max-width: 900px) {
  .toolbar { align-items: flex-start; flex-direction: column; }
}
</style>
