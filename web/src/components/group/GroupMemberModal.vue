<template>
  <AppModal @close="ui.closeModal()">
    <template #header>
      <h3>群成员</h3>
      <button class="close-btn" @click="ui.closeModal()">×</button>
    </template>
    <div v-if="loading" class="loading">加载中...</div>
    <div v-for="member in members" :key="member.id" class="member-item">
      <AppAvatar :src="member.avatar" :name="member.username" />
      <div class="member-body">
        <span class="member-name">
          {{ member.username }}
          <span class="role-tag">{{ roleText(member.role) }}</span>
        </span>
        <span class="member-status">{{ member.online ? '在线' : '离线' }}</span>
      </div>
      <div v-if="canManage(member)" class="member-actions">
        <button
          v-if="isOwner"
          class="action-sm"
          @click="handleToggleAdmin(member)"
        >
          {{ member.role === 1 ? '取消管理' : '设管理' }}
        </button>
        <button class="action-sm danger" @click="handleKick(member)">踢出</button>
      </div>
    </div>
  </AppModal>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useUiStore } from '../../stores/ui.js'
import { useChatStore } from '../../stores/chat.js'
import { useContactsStore } from '../../stores/contacts.js'
import { useAuthStore } from '../../stores/auth.js'
import { ROOM_ROLE } from '../../utils/constants.js'
import * as contactsApi from '../../api/contacts.js'
import AppModal from '../common/AppModal.vue'
import AppAvatar from '../common/AppAvatar.vue'

const ui = useUiStore()
const chat = useChatStore()
const contacts = useContactsStore()
const auth = useAuthStore()
const members = ref([])
const loading = ref(true)

const roleText = (r) => r === 0 ? '群主' : r === 1 ? '管理员' : ''
const isOwner = computed(() => {
  const g = contacts.getGroupById(chat.activeConversation?.id)
  return g?.owner_id === auth.userId
})

onMounted(async () => {
  try {
    const { data } = await contactsApi.getGroupMembers(chat.activeConversation.id)
    if (data.code === 0) members.value = data.data || []
  } finally { loading.value = false }
})

function canManage(member) {
  if (member.id === auth.userId) return false
  if (isOwner.value) return true
  // Admin can kick normal members
  const self = members.value.find(m => m.id === auth.userId)
  if (self?.role === ROOM_ROLE.ADMIN && member.role === ROOM_ROLE.MEMBER) return true
  return false
}

async function handleToggleAdmin(member) {
  const isAdmin = member.role !== ROOM_ROLE.ADMIN
  try {
    await contactsApi.setAdmin(chat.activeConversation.id, member.id, isAdmin)
    // Refresh
    const { data } = await contactsApi.getGroupMembers(chat.activeConversation.id)
    if (data.code === 0) members.value = data.data || []
    ui.showToast(isAdmin ? '已设为管理员' : '已取消管理员', 'success')
  } catch (e) { ui.showToast(e.message, 'error') }
}

async function handleKick(member) {
  if (!confirm(`确定踢出 ${member.username}？`)) return
  try {
    await contactsApi.kickMember(chat.activeConversation.id, member.id)
    members.value = members.value.filter(m => m.id !== member.id)
    ui.showToast('已踢出', 'success')
  } catch (e) { ui.showToast(e.message, 'error') }
}
</script>

<style scoped>
.close-btn { background: none; border: none; font-size: 24px; color: var(--text-hint); padding: 0; }
.loading { text-align: center; padding: 20px; color: var(--text-hint); }
.member-item {
  display: flex; align-items: center; gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-light);
}
.member-body { flex: 1; }
.member-name { font-weight: 500; display: block; }
.role-tag { font-size: var(--font-xs); color: var(--primary); font-weight: normal; }
.member-status { font-size: var(--font-xs); color: var(--text-hint); }
.member-actions { display: flex; gap: 6px; }
.action-sm {
  padding: 4px 10px; border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: var(--bg-white); color: var(--text-primary);
  font-size: var(--font-xs); cursor: pointer;
}
.action-sm.danger { color: var(--danger); border-color: var(--danger); }
</style>
