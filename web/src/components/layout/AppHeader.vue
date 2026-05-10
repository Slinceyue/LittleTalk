<template>
  <header class="header">
    <h1 class="header-title">LittleTalk</h1>
    <div class="header-actions">
      <button class="header-btn" @click="$emit('addFriend')" title="添加好友">+</button>
      <button class="header-btn" @click="$emit('themeToggle')" :title="isDark ? '切换亮色' : '切换暗色'">
        {{ isDark ? '☀️' : '🌙' }}
      </button>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useUiStore } from '../../stores/ui.js'
const ui = useUiStore()
const isDark = computed(() => ui.theme === 'dark')
defineEmits(['addFriend', 'themeToggle'])
</script>

<style scoped>
.header {
  height: var(--header-height);
  background: var(--bg-white);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}
.header-title { font-size: var(--font-lg); font-weight: 600; }
.header-actions { display: flex; gap: 8px; }
.header-btn {
  width: 32px; height: 32px;
  border: none; background: none;
  font-size: 20px;
  display: flex; align-items: center; justify-content: center;
  border-radius: var(--radius-full);
  cursor: pointer;
  color: var(--text-primary);
}
.header-btn:active { background: var(--bg-hover); }
</style>
