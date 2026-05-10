<template>
  <AppModal @close="ui.closeModal()">
    <template #header>
      <h3>好友请求</h3>
      <button class="close-btn" @click="ui.closeModal()">×</button>
    </template>
    <AppEmptyState v-if="contacts.friendRequests.length === 0" icon="📭" text="暂无好友请求" />
    <div v-for="userId in contacts.friendRequests" :key="userId" class="req-item">
      <AppAvatar
        :src="contacts.getUserAvatar(userId)"
        :name="contacts.getUserName(userId)"
      />
      <div class="req-info">
        <div class="req-name">{{ contacts.getUserName(userId) }}</div>
        <div class="req-id">ID: {{ userId }}</div>
      </div>
      <div class="req-actions">
        <button class="btn-accept" @click="handleAccept(userId)">接受</button>
        <button class="btn-reject" @click="handleReject(userId)">拒绝</button>
      </div>
    </div>
  </AppModal>
</template>

<script setup>
import { onMounted } from 'vue'
import { useUiStore } from '../../stores/ui.js'
import { useContactsStore } from '../../stores/contacts.js'
import AppModal from '../common/AppModal.vue'
import AppAvatar from '../common/AppAvatar.vue'
import AppEmptyState from '../common/AppEmptyState.vue'

const ui = useUiStore()
const contacts = useContactsStore()

onMounted(() => {
  contacts.fetchFriendRequests()
})

async function handleAccept(userId) {
  try {
    await contacts.acceptFriendRequest(userId)
    ui.showToast('已接受好友请求', 'success')
  } catch (e) { ui.showToast(e.message, 'error') }
}

async function handleReject(userId) {
  try {
    await contacts.rejectFriendRequest(userId)
    ui.showToast('已拒绝', 'success')
  } catch (e) { ui.showToast(e.message, 'error') }
}
</script>

<style scoped>
.close-btn { background: none; border: none; font-size: 24px; color: var(--text-hint); padding: 0; }
.req-item {
  display: flex; align-items: center; gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid var(--border-light);
}
.req-info { flex: 1; }
.req-name { font-weight: 500; }
.req-id { font-size: var(--font-xs); color: var(--text-hint); }
.req-actions { display: flex; gap: 8px; }
.btn-accept, .btn-reject {
  padding: 6px 14px; border-radius: var(--radius-sm); border: none;
  font-size: var(--font-sm); cursor: pointer;
}
.btn-accept { background: var(--primary); color: #fff; }
.btn-reject { background: var(--bg-input); color: var(--text-secondary); }
</style>
