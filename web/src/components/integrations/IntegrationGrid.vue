<template>
  <div class="component-grid">
    <button v-for="item in components" :key="item.ident" type="button" class="component-card" @click="$emit('open-component', item)">
      <div class="logo">{{ item.logo }}</div>
      <div class="component-body">
        <strong>{{ item.ident }}</strong>
        <span>{{ item.readme || '暂无说明' }}</span>
        <div class="counts">
          <el-tag size="small">{{ item.dashboardCount }} 仪表盘</el-tag>
          <el-tag size="small" type="info">{{ item.alertCount }} 告警</el-tag>
          <el-tag size="small" type="warning">{{ item.metricCount }} 指标</el-tag>
          <el-tag size="small" type="success">{{ item.collectCount }} 采集</el-tag>
        </div>
      </div>
      <el-tag v-if="item.disabled" class="status" type="info" size="small">停用</el-tag>
    </button>
  </div>
</template>

<script setup>
defineProps({
  components: { type: Array, default: () => [] },
})

defineEmits(['open-component'])
</script>

<style scoped>
.component-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 180px), 1fr)); gap: 12px; }
.component-card { position: relative; min-height: 164px; padding: 14px; border: 1px solid #e1e8f2; border-radius: 8px; background: #fff; color: #25324a; text-align: left; cursor: pointer; }
.component-card:hover { border-color: rgba(23,105,255,.45); box-shadow: 0 10px 24px rgba(31,45,61,.08); }
.logo { display: grid; place-items: center; width: 42px; height: 42px; margin-bottom: 10px; border-radius: 8px; background: #eaf2ff; color: #1769ff; font-weight: 800; }
.component-body strong { display: block; font-size: 14px; }
.component-body span { display: -webkit-box; min-height: 38px; margin-top: 6px; overflow: hidden; color: #65748c; font-size: 12px; line-height: 1.55; -webkit-box-orient: vertical; -webkit-line-clamp: 2; }
.counts { display: flex; gap: 5px; flex-wrap: wrap; margin-top: 10px; }
.status { position: absolute; top: 10px; right: 10px; }
</style>
