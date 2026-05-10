<template>
  <div class="empty">
    <div class="empty-icon">{{ icon }}</div>
    <div class="empty-text">{{ text }}</div>
    <slot />
  </div>
</template>

<script setup>
defineProps({
  icon: { type: String, default: '👋' },
  text: { type: String, default: '暂无数据' },
})
</script>

<style scoped>
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: var(--text-hint);
}
.empty-icon { font-size: 48px; margin-bottom: 12px; opacity: 0.6; }
.empty-text { font-size: var(--font-sm); }
</style>
