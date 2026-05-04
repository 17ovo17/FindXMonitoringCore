<template>
  <div class="aiops-page">
    <section v-if="section === 'diagnosis'" class="embedded-panel"><Workbench /></section>
    <section v-else-if="section === 'knowledge'" class="embedded-panel"><KnowledgeCenter /></section>
    <section v-else-if="section === 'workflow'" class="embedded-panel"><WorkflowHub /></section>
    <section v-else-if="section === 'asset-chat'" class="embedded-panel"><CatpawChatPanel /></section>
    <section v-else class="aiops-panel">
      <div class="section-head">
        <div>
          <div class="kicker">AI SRE</div>
          <h2>自动修复</h2>
          <p>将诊断结论、知识库和工作流串联为可审计的处置动作。</p>
        </div>
      </div>
      <div class="empty-state">
        <el-empty description="自动修复入口尚未接入真实编排数据。">
          <el-button type="primary" @click="router.push('/aiops?section=workflow')">查看工作流</el-button>
        </el-empty>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, defineAsyncComponent } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const validSections = new Set(['diagnosis', 'knowledge', 'workflow', 'remediation', 'asset-chat'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'diagnosis')

const Workbench = defineAsyncComponent(() => import('./Workbench.vue'))
const KnowledgeCenter = defineAsyncComponent(() => import('./KnowledgeCenter.vue'))
const WorkflowHub = defineAsyncComponent(() => import('./WorkflowHub.vue'))
const CatpawChatPanel = defineAsyncComponent(() => import('./CatpawChatPanel.vue'))
</script>

<style scoped>
.aiops-page { min-height: 100%; padding: 24px; color: #243553; }
.embedded-panel, .aiops-panel { min-height: calc(100vh - 114px); border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.86); box-shadow: 0 12px 34px rgba(31,45,61,.06); overflow: auto; }
.aiops-panel { padding: 22px; }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.empty-state { min-height: 420px; display: grid; place-items: center; border: 1px dashed #d8e1ee; border-radius: 8px; margin-top: 18px; background: #f8fbff; }
</style>
