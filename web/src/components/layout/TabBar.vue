<template>
  <nav class="tabbar">
    <div
      v-for="tab in tabs"
      :key="tab.id"
      class="tab-item"
      :class="{ active: activeTab === tab.id }"
      @click="$emit('switch', tab.id)"
    >
      <span class="tab-icon">{{ tab.icon }}</span>
      <span class="tab-label">{{ tab.label }}</span>
      <span v-if="tab.badge" class="tab-badge">{{ tab.badge > 99 ? '99+' : tab.badge }}</span>
    </div>
  </nav>
</template>

<script setup>
defineProps({
  activeTab: { type: String, default: 'messages' },
  tabs: { type: Array, default: () => [] },
})
defineEmits(['switch'])
</script>

<style scoped>
.tabbar {
  height: var(--tabbar-height);
  background: var(--bg-white);
  border-top: 1px solid var(--border-light);
  display: flex;
  flex-shrink: 0;
  padding-bottom: var(--safe-bottom);
}
.tab-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  cursor: pointer;
  position: relative;
  color: var(--text-hint);
  transition: color 0.2s;
}
.tab-item.active { color: var(--primary); }
.tab-icon { font-size: 22px; line-height: 1; transition: transform 0.2s; }
.tab-item.active .tab-icon { transform: scale(1.1); }
.tab-label { font-size: 10px; }
.tab-badge {
  position: absolute;
  top: 4px;
  right: calc(50% - 20px);
  min-width: 18px; height: 18px;
  border-radius: 9px;
  background: var(--danger);
  color: #fff;
  font-size: 10px;
  display: flex; align-items: center; justify-content: center;
  padding: 0 4px;
}
</style>
