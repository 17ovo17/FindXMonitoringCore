<template>
  <section class="templates-layout">
    <aside class="template-list">
      <div class="side-head">
        <b>消息模板</b>
        <el-button link type="primary" @click="$emit('blocked', 'template-save', { mode: 'add' })">新增</el-button>
      </div>
      <el-input v-model="query" clearable placeholder="搜索模板" />
      <el-alert :title="blockedContracts['template-list']" type="warning" show-icon :closable="false" class="state-alert" />
      <el-empty description="BLOCKED_BY_CONTRACT：模板列表 contract 未暴露" />
    </aside>
    <main class="template-detail">
      <header class="detail-head">
        <div>
          <div class="kicker">消息模板</div>
          <h3>模板详情</h3>
        </div>
        <div class="toolbar-actions">
          <el-button @click="$emit('blocked', 'template-save', { mode: 'edit' })">编辑</el-button>
          <el-button @click="$emit('blocked', 'template-save', { mode: 'clone' })">克隆</el-button>
          <el-button type="danger" plain @click="$emit('blocked', 'template-save', { mode: 'delete' })">删除</el-button>
        </div>
      </header>
      <el-alert :title="blockedContracts['template-save']" type="warning" show-icon :closable="false" />
      <el-tabs class="content-tabs">
        <el-tab-pane label="Text" />
        <el-tab-pane label="Markdown" />
        <el-tab-pane label="HTML" />
      </el-tabs>
      <el-input :model-value="blockedPayload('template-list', { query })" type="textarea" :rows="18" readonly spellcheck="false" />
    </main>
  </section>
</template>

<script setup>
import { ref } from 'vue'
import { blockedContracts, blockedPayload } from './notificationModel'

defineEmits(['blocked'])
const query = ref('')
</script>

<style scoped>
.templates-layout { display: grid; grid-template-columns: 280px minmax(0, 1fr); gap: 16px; min-height: calc(100vh - 126px); }
.template-list, .template-detail { border: 1px solid #e3e8f1; border-radius: 8px; background: #fff; padding: 16px; }
.side-head, .detail-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.toolbar-actions { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h3 { margin: 4px 0 0; color: #17233c; font-size: 20px; }
.state-alert, .content-tabs { margin-top: 12px; }
@media (max-width: 900px) { .templates-layout { grid-template-columns: 1fr; } .detail-head { align-items: flex-start; flex-direction: column; } }
</style>
