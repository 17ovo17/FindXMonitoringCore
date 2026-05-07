<template>
  <section class="detail-area">
    <header class="detail-title">
      <div class="detail-left">
        <el-button text @click="$emit('go-list')">返回列表</el-button>
        <el-select :model-value="activeDashboardId" filterable placeholder="切换仪表盘" class="title-switch" @update:model-value="$emit('update:activeDashboardId', $event)" @change="$emit('open-detail', $event)">
          <el-option v-for="item in dashboards" :key="item.id" :label="item.name" :value="item.id" />
        </el-select>
        <div>
          <h2>{{ activeDashboard?.name || '仪表盘详情' }}</h2>
          <div class="ident">{{ activeDashboard?.note || activeDashboard?.ident || '读取真实配置中' }}</div>
        </div>
      </div>
      <div class="detail-actions">
        <el-dropdown trigger="click" @command="$emit('panel-editor', $event)">
          <el-button type="primary">添加图表</el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-for="item in panelTypes" :key="item.value" :command="item.value">{{ item.label }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button :loading="detailLoading" @click="$emit('refresh-detail')">刷新</el-button>
        <el-select :model-value="autoRefresh" class="small-select" @update:model-value="$emit('update:autoRefresh', $event)">
          <el-option label="自动刷新 Off" value="off" />
          <el-option label="30 秒" value="30s" />
          <el-option label="1 分钟" value="1m" />
          <el-option label="5 分钟" value="5m" />
        </el-select>
        <el-select :model-value="timeRange" class="range-select" @update:model-value="$emit('update:timeRange', $event)">
          <el-option label="最近 15 分钟" value="15m" />
          <el-option label="最近 1 小时" value="1h" />
          <el-option label="最近 6 小时" value="6h" />
          <el-option label="最近 24 小时" value="24h" />
        </el-select>
        <el-select :model-value="timezone" class="small-select" @update:model-value="$emit('update:timezone', $event)">
          <el-option label="Local" value="local" />
          <el-option label="UTC" value="utc" />
          <el-option label="Asia/Shanghai" value="Asia/Shanghai" />
        </el-select>
        <el-button @click="$emit('settings')">设置</el-button>
        <el-button @click="$emit('copy-link')">链接</el-button>
        <el-button @click="$emit('fullscreen')">全屏</el-button>
      </div>
    </header>

    <el-alert v-if="detailError" :title="detailError" type="warning" show-icon :closable="false" class="state-alert" />
    <section class="variables-bar">
      <div class="section-heading">变量</div>
      <div v-if="variables.length === 0" class="blocked">BLOCKED_BY_CONTRACT：当前仪表盘未返回 variables 配置或变量选项 contract。</div>
      <div v-else class="variable-grid">
        <div v-for="variable in variables" :key="variable.key" class="variable-item">
          <label>{{ variable.label }}</label>
          <el-select v-if="variable.control === 'select'" v-model="variableValues[variable.key]" filterable allow-create clearable placeholder="搜索或选择">
            <el-option v-for="option in variable.options" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
          <el-input v-else v-model="variableValues[variable.key]" clearable placeholder="输入变量值" />
          <div v-if="variable.blocked" class="mini-blocked">{{ variable.blockedReason }}</div>
        </div>
      </div>
    </section>

    <section class="panel-grid">
      <article v-for="panel in panels" :key="panel.id" class="panel-card">
        <header>
          <div>
            <strong>{{ panel.title }}</strong>
            <span>{{ panel.type }}</span>
          </div>
          <el-dropdown trigger="click" @command="cmd => $emit('panel-command', cmd, panel)">
            <el-button text>更多</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="edit">编辑</el-dropdown-item>
                <el-dropdown-item command="clone">克隆</el-dropdown-item>
                <el-dropdown-item command="copy">复制配置</el-dropdown-item>
                <el-dropdown-item command="share">分享</el-dropdown-item>
                <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </header>
        <div class="panel-body">
          <template v-if="panel.rendererBlocked">
            <div class="panel-blocked">{{ panel.blockedReason }}</div>
            <pre class="panel-config">{{ panel.preview }}</pre>
          </template>
          <template v-else>
            <div class="metric">{{ panel.preview }}</div>
            <div class="panel-note">{{ panel.note }}</div>
          </template>
        </div>
      </article>
      <div v-if="panels.length === 0" class="empty-panels">BLOCKED_BY_CONTRACT：当前详情未返回 panels/configs 数据，无法渲染真实图表网格。</div>
    </section>
  </section>
</template>

<script setup>
defineProps({
  detailLoading: Boolean,
  detailError: String,
  activeDashboardId: String,
  activeDashboard: Object,
  autoRefresh: String,
  timeRange: String,
  timezone: String,
  variableValues: { type: Object, required: true },
  dashboards: { type: Array, default: () => [] },
  variables: { type: Array, default: () => [] },
  panels: { type: Array, default: () => [] },
  panelTypes: { type: Array, default: () => [] },
})

defineEmits(['update:activeDashboardId', 'update:autoRefresh', 'update:timeRange', 'update:timezone', 'go-list', 'open-detail', 'refresh-detail', 'panel-editor', 'panel-command', 'settings', 'copy-link', 'fullscreen'])
</script>

<style scoped>
.detail-area { padding: 16px; border: 1px solid #e3e8f1; border-radius: 8px; background: #fff; }
.detail-title { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
.detail-actions, .detail-left { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
h2 { margin: 0; color: #17233c; font-size: 22px; }
.ident { color: #8a96aa; font-size: 12px; }
.title-switch { width: 240px; }
.small-select { width: 132px; }
.range-select { width: 148px; }
.state-alert { margin-bottom: 14px; }
.variables-bar { margin-bottom: 14px; padding: 12px; border: 1px solid #e7edf5; border-radius: 8px; background: #fafcff; }
.section-heading { margin-bottom: 10px; font-weight: 700; }
.variable-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 210px), 1fr)); gap: 10px; }
.variable-item label { display: block; margin-bottom: 6px; color: #52627a; font-size: 12px; }
.blocked, .mini-blocked, .empty-panels, .panel-blocked { color: #9a5b00; font-size: 12px; line-height: 1.6; }
.mini-blocked { margin-top: 5px; }
.panel-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 280px), 1fr)); gap: 12px; }
.panel-card { min-width: 0; min-height: 170px; padding: 12px; border: 1px solid #dfe6f0; border-radius: 8px; background: #fff; }
.panel-card header { display: flex; justify-content: space-between; gap: 8px; }
.panel-card header span { display: block; margin-top: 4px; color: #748198; font-size: 12px; }
.panel-body { display: grid; place-content: center; min-height: 108px; margin-top: 10px; padding: 12px; border-radius: 6px; background: #f6f8fc; text-align: center; }
.metric { color: #1769ff; font-size: 20px; font-weight: 700; }
.panel-note { max-width: 220px; color: #687891; font-size: 12px; }
.panel-config { max-width: 100%; margin: 8px 0 0; overflow: hidden; color: #5d6d86; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
@media (max-width: 900px) {
  .detail-title { align-items: flex-start; flex-direction: column; }
  .title-switch, .small-select, .range-select { width: 100%; }
}
</style>
