<template>
  <div class="org-page">
    <section class="org-panel">
      <div class="section-head">
        <div>
          <div class="kicker">人员组织</div>
          <h2>{{ current.title }}</h2>
          <p>{{ current.desc }}</p>
        </div>
      </div>
      <div class="empty-state">
        <el-empty :description="current.empty">
          <el-alert :title="current.hint" type="info" show-icon :closable="false" />
        </el-empty>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const validSections = new Set(['users', 'teams', 'roles'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'users')
const copy = {
  users: { title: '用户管理', desc: '维护用户资料、登录状态和个人令牌入口，敏感令牌只允许摘要展示。', empty: '用户管理尚未接入真实接口。', hint: '当前不展示临时用户列表。' },
  teams: { title: '团队组织', desc: '通知接收、资源授权和协作对象都复用团队组织机制。', empty: '团队组织管理尚未接入真实接口。', hint: '接收人归并到团队组织和通知规则。' },
  roles: { title: '角色管理', desc: '统一承载角色、权限矩阵、资源范围和操作授权配置。', empty: '角色管理尚未接入真实接口。', hint: '待权限模型落库后展示。' },
}
const current = computed(() => copy[section.value])
</script>

<style scoped>
.org-page { min-height: 100%; padding: 24px; color: #243553; }
.org-panel { min-height: calc(100vh - 114px); padding: 22px; border: 1px solid #e4e9f2; border-radius: 8px; background: rgba(255,255,255,.86); box-shadow: 0 12px 34px rgba(31,45,61,.06); }
.kicker { color: #1769ff; font-size: 12px; font-weight: 800; }
h2 { margin: 6px 0 0; color: #1e3a5f; font-size: 24px; }
p { margin: 8px 0 0; color: #60728e; font-size: 13px; line-height: 1.6; }
.empty-state { min-height: 420px; display: grid; place-items: center; border: 1px dashed #d8e1ee; border-radius: 8px; margin-top: 18px; background: #f8fbff; }
</style>
