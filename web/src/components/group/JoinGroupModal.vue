<template>
  <AppModal @close="ui.closeModal()">
    <template #header>
      <h3>加入群聊</h3>
      <button class="close-btn" @click="ui.closeModal()">×</button>
    </template>
    <div class="form-group">
      <label>群ID</label>
      <input v-model="roomId" type="number" placeholder="请输入群ID" />
    </div>
    <button class="btn btn-primary w-full" @click="join" :disabled="loading">
      {{ loading ? '加入中...' : '加入' }}
    </button>
  </AppModal>
</template>

<script setup>
import { ref } from 'vue'
import { useUiStore } from '../../stores/ui.js'
import { useContactsStore } from '../../stores/contacts.js'
import AppModal from '../common/AppModal.vue'

const ui = useUiStore()
const contacts = useContactsStore()
const roomId = ref('')
const loading = ref(false)

async function join() {
  const id = parseInt(roomId.value)
  if (!id) { ui.showToast('请输入群ID', 'error'); return }
  loading.value = true
  try {
    await contacts.joinGroup(id)
    ui.closeModal()
    ui.showToast('加入成功', 'success')
    contacts.fetchGroups()
  } catch (e) {
    ui.showToast(e.message, 'error')
  } finally { loading.value = false }
}
</script>

<style scoped>
.close-btn { background: none; border: none; font-size: 24px; color: var(--text-hint); padding: 0; }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; font-size: var(--font-sm); color: var(--text-secondary); margin-bottom: 6px; }
.form-group input {
  width: 100%; padding: 12px;
  border: 1px solid var(--border); border-radius: var(--radius-md);
  background: var(--bg-input); color: var(--text-primary);
  outline: none; font-size: var(--font-md);
}
.btn { padding: 12px; border: none; border-radius: var(--radius-md); font-size: var(--font-md); }
.btn-primary { background: var(--primary); color: #fff; }
.btn:disabled { opacity: 0.6; }
.w-full { width: 100%; }
</style>
