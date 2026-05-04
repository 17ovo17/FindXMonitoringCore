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
const validSections = new Set(['users', 'teams', 'roles', 'permissions', 'tokens'])
const section = computed(() => validSections.has(route.query.section) ? route.query.section : 'users')
const copy = {
  users: { title: '用户管理', desc: '维护用户资料、登录状态和个人令牌边界。', empty: '用户管理尚未接入真实接口。', hint: '当前不展示临时用户列表。' },
  teams: { title: '团队组织', desc: '通知接收、资源授权和订阅对象都复用团队组织。', empty: '团队组织管理尚未接入真实接口。', hint: '接收人归并到团队组织和通知规则。' },
  roles: { title: '角色管理', desc: '配置角色和功能权限。', empty: '角色管理尚未接入真实接口。', hint: '待权限模型落库后展示。' },
  permissions: { title: '操作权限', desc: '按角色、资源和动作查看权限边界。', empty: '操作权限尚未接入真实接口。', hint: '当前保持真实空态。' },
  tokens: { title: 'Token 管理', desc: '管理用户 Token、SourceToken 和外部调用令牌。', empty: 'Token 管理尚未接入真实接口。', hint: '不会展示示例 Token 或敏感配置。' },
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
