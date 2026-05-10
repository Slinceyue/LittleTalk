<template>
  <div class="search-panel">
    <div class="search-bar">
      <span class="search-icon">🔍</span>
      <input
        ref="inputEl"
        v-model="query"
        class="search-input"
        placeholder="搜索消息..."
        @keydown.enter="doSearch"
        @keydown.escape="$emit('close')"
      />
      <button v-if="query" class="search-clear" @click="query = ''">×</button>
      <button class="search-close" @click="$emit('close')">取消</button>
    </div>

    <div class="search-results">
      <div v-if="loading" class="search-status">搜索中...</div>
      <div v-else-if="searched && !results.length" class="search-status">未找到相关消息</div>

      <div
        v-for="item in results"
        :key="item.msg_id"
        class="search-item"
        @click="$emit('jump', item)"
      >
        <AppAvatar :src="item.from_avatar" :name="item.from_name || item.from_id" size="sm" />
        <div class="search-item-body">
          <div class="search-item-header">
            <span class="search-item-name">{{ item.from_name || item.from_id }}</span>
            <span class="search-item-time">{{ formatTime(item.send_time) }}</span>
          </div>
          <div class="search-item-content" v-html="highlight(item.content)"></div>
        </div>
      </div>

      <div v-if="hasMore" class="search-more" @click="loadMore">加载更多</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import AppAvatar from '../common/AppAvatar.vue'
import { formatTime } from '../../utils/formatters.js'
import * as messagesApi from '../../api/messages.js'

const props = defineProps({
  convKey: { type: String, required: true },
})
defineEmits(['close', 'jump'])

const query = ref('')
const inputEl = ref(null)
const results = ref([])
const loading = ref(false)
const searched = ref(false)
const page = ref(1)
const hasMore = ref(false)

onMounted(() => {
  inputEl.value?.focus()
})

async function doSearch() {
  if (!query.value.trim()) return
  loading.value = true
  searched.value = true
  page.value = 1
  try {
    const parts = props.convKey.split('_')
    const friendId = parts[1] ? parseInt(parts[1]) : 0
    const { data } = await messagesApi.searchMessages(query.value.trim(), friendId, 1)
    if (data.code === 0) {
      results.value = data.data?.list || data.data || []
      hasMore.value = results.value.length >= 20
    }
  } catch { /* ignore */ }
  finally { loading.value = false }
}

async function loadMore() {
  page.value++
  loading.value = true
  try {
    const parts = props.convKey.split('_')
    const friendId = parts[1] ? parseInt(parts[1]) : 0
    const { data } = await messagesApi.searchMessages(query.value.trim(), friendId, page.value)
    if (data.code === 0) {
      const more = data.data?.list || data.data || []
      results.value.push(...more)
      hasMore.value = more.length >= 20
    }
  } catch { /* ignore */ }
  finally { loading.value = false }
}

function highlight(content) {
  if (!content || !query.value) return content
  const escaped = query.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return content.replace(new RegExp(`(${escaped})`, 'gi'), '<mark>$1</mark>')
}
</script>

<style scoped>
.search-panel {
  background: var(--bg-white);
  border-bottom: 1px solid var(--border-light);
  display: flex;
  flex-direction: column;
  max-height: 320px;
}
.search-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
}
.search-icon {
  font-size: 14px;
  flex-shrink: 0;
}
.search-input {
  flex: 1;
  border: none;
  background: var(--bg-input);
  border-radius: 6px;
  padding: 8px 10px;
  font-size: var(--font-sm);
  color: var(--text-primary);
  outline: none;
}
.search-clear {
  border: none;
  background: none;
  font-size: 18px;
  color: var(--text-hint);
  cursor: pointer;
  padding: 0 4px;
}
.search-close {
  border: none;
  background: none;
  font-size: var(--font-sm);
  color: var(--primary);
  cursor: pointer;
  padding: 4px;
  white-space: nowrap;
}
.search-results {
  flex: 1;
  overflow-y: auto;
  padding: 0 12px 8px;
}
.search-status {
  text-align: center;
  padding: 20px;
  color: var(--text-hint);
  font-size: var(--font-sm);
}
.search-item {
  display: flex;
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-light);
  cursor: pointer;
  align-items: flex-start;
}
.search-item-body {
  flex: 1;
  min-width: 0;
}
.search-item-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}
.search-item-name {
  font-size: var(--font-sm);
  color: var(--text-primary);
  font-weight: 500;
}
.search-item-time {
  font-size: var(--font-xs);
  color: var(--text-hint);
  flex-shrink: 0;
}
.search-item-content {
  font-size: var(--font-sm);
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.search-item-content :deep(mark) {
  background: #fff3cd;
  color: #856404;
  padding: 0 2px;
  border-radius: 2px;
}
[data-theme="dark"] .search-item-content :deep(mark) {
  background: #5c4a1f;
  color: #ffc107;
}
.search-more {
  text-align: center;
  padding: 10px;
  color: var(--primary);
  font-size: var(--font-sm);
  cursor: pointer;
}
</style>
