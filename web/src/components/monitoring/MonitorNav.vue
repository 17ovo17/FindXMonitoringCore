<template>
  <aside class="monitor-nav" aria-label="监控运维导航">
    <div v-for="group in groups" :key="group.label" class="nav-group">
      <div class="nav-title">{{ group.label }}</div>
      <button
        v-for="item in group.items"
        :key="item.key"
        class="nav-item"
        :class="{ active: activeKey === item.key }"
        type="button"
        @click="$emit('select', item.key)"
      >
        <el-icon><component :is="item.icon" /></el-icon>
        <span>{{ item.label }}</span>
        <small v-if="item.status">{{ item.status }}</small>
      </button>
    </div>
  </aside>
</template>

<script setup>
defineProps({
  groups: { type: Array, required: true },
  activeKey: { type: String, required: true },
})

defineEmits(['select'])
</script>

<style scoped>
.monitor-nav {
  position: sticky;
  top: 16px;
  max-height: calc(100vh - 122px);
  overflow: auto;
  padding: 12px;
  border: 1px solid #e4e9f2;
  border-radius: 8px;
  background: rgba(255, 255, 255, .82);
}
.nav-group + .nav-group { margin-top: 14px; }
.nav-title {
  padding: 6px 8px;
  color: #7a89a3;
  font-size: 12px;
  font-weight: 800;
}
.nav-item {
  width: 100%;
  min-height: 36px;
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr) auto;
  align-items: center;
  gap: 8px;
  padding: 7px 8px;
  border: 0;
  border-radius: 8px;
  color: #243553;
  background: transparent;
  text-align: left;
  cursor: pointer;
}
.nav-item span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  font-weight: 700;
}
.nav-item small { color: #8a98ad; font-size: 11px; }
.nav-item:hover,
.nav-item.active {
  color: #1769ff;
  background: #e8f1ff;
}
@media (max-width: 1024px) {
  .monitor-nav {
    position: static;
    max-height: none;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }
  .nav-group + .nav-group { margin-top: 0; }
}
@media (max-width: 640px) {
  .monitor-nav { grid-template-columns: 1fr; }
}
</style>
