<template>
  <div class="add-page">
    <div class="add-header">
      <button class="back-btn" @click="ui.goBack()">‹</button>
      <span class="add-title">添加好友</span>
      <div style="width:32px"></div>
    </div>
    <div class="add-form">
      <div class="form-group">
        <label>好友ID</label>
        <input v-model="friendId" type="number" placeholder="请输入好友ID" />
      </div>
      <button class="btn btn-primary" @click="handleSend" :disabled="loading">
        {{ loading ? '发送中...' : '发送好友请求' }}
      </button>
      <p class="error" v-if="error">{{ error }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useUiStore } from '../../stores/ui.js'
import { useContactsStore } from '../../stores/contacts.js'

const ui = useUiStore()
const contacts = useContactsStore()
const friendId = ref('')
const loading = ref(false)
const error = ref('')

async function handleSend() {
  const id = parseInt(friendId.value)
  if (!id) { error.value = '请输入有效的用户ID'; return }
  loading.value = true; error.value = ''
  try {
    await contacts.sendFriendRequest(id)
    ui.showToast('好友请求已发送', 'success')
    ui.goBack()
  } catch (e) {
    error.value = e.message || '发送失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.add-page {
  width: 100%; height: 100%;
  display: flex; flex-direction: column;
  background: var(--bg-page);
}
.add-header {
  height: var(--header-height);
  background: var(--bg-white);
  display: flex; align-items: center;
  padding: 0 8px;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}
.back-btn { background: none; border: none; font-size: 28px; padding: 4px 8px; color: var(--text-primary); cursor: pointer; }
.add-title { flex: 1; text-align: center; font-size: var(--font-lg); font-weight: 500; }
.add-form { padding: 24px 16px; display: flex; flex-direction: column; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: var(--font-sm); color: var(--text-secondary); }
.form-group input {
  padding: 12px; border: 1px solid var(--border);
  border-radius: var(--radius-md); background: var(--bg-input);
  color: var(--text-primary); outline: none; font-size: var(--font-md);
}
.form-group input:focus { border-color: var(--primary); }
.btn { padding: 12px; border: none; border-radius: var(--radius-md); font-size: var(--font-md); }
.btn-primary { background: var(--primary); color: #fff; }
.btn:disabled { opacity: 0.6; }
.error { color: var(--danger); font-size: var(--font-sm); text-align: center; }
</style>
