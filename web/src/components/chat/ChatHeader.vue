<template>
  <div class="chat-header">
    <button class="back-btn" @click="$emit('back')">‹</button>
    <div class="header-center">
      <span class="header-name">{{ name }}</span>
      <span v-if="!isGroup" class="header-status" :class="{ online }">
        {{ online ? '在线' : '离线' }}
      </span>
    </div>
    <div class="header-actions">
      <button class="action-btn" @click="$emit('search')" title="搜索">🔍</button>
      <button class="action-btn" @click="$emit('bubble')" title="气泡样式">💬</button>
      <button class="action-btn" @click="$emit('bg')" title="聊天背景">🎨</button>
    </div>
  </div>
</template>

<script setup>
defineProps({
  name: { type: String, default: '' },
  online: { type: Boolean, default: false },
  isGroup: { type: Boolean, default: false },
})
defineEmits(['back', 'info', 'search', 'bg', 'bubble'])
</script>

<style scoped>
.chat-header {
  height: var(--header-height);
  background: var(--bg-white);
  display: flex; align-items: center;
  padding: 0 8px;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}
.back-btn { background: none; border: none; font-size: 28px; padding: 4px 8px; color: var(--text-primary); cursor: pointer; }
.header-center { flex: 1; text-align: center; cursor: pointer; }
.header-name { font-size: var(--font-md); font-weight: 600; display: block; }
.header-status { font-size: var(--font-xs); color: var(--text-hint); }
.header-status.online { color: var(--primary); }
.header-actions { display: flex; gap: 4px; }
.action-btn {
  width: 36px; height: 36px;
  border: none; background: none;
  font-size: 16px;
  display: flex; align-items: center; justify-content: center;
  border-radius: var(--radius-full);
  cursor: pointer;
}
.action-btn:active { background: var(--bg-hover); }
</style>
