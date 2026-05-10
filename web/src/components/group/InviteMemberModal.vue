<template>
  <AppModal @close="ui.closeModal()">
    <template #header>
      <h3>邀请好友入群</h3>
      <button class="close-btn" @click="ui.closeModal()">×</button>
    </template>
    <AppEmptyState v-if="availableFriends.length === 0" icon="👥" text="所有好友都已在群中" />
    <div
      v-for="friend in availableFriends"
      :key="friend.id"
      class="invite-item"
      :class="{ selected: selected.has(friend.id) }"
      @click="toggle(friend.id)"
    >
      <AppAvatar :src="getAvatar(friend)" :name="friend.username" />
      <span class="invite-name">{{ friend.username }}</span>
      <span class="check-mark" v-if="selected.has(friend.id)">✓</span>
    </div>
    <template #footer v-if="availableFriends.length > 0">
      <button class="btn" @click="ui.closeModal()">取消</button>
      <button class="btn btn-primary flex-1" @click="confirm" :disabled="selected.size === 0 || loading">
        {{ loading ? '邀请中...' : `确认邀请 (${selected.size})` }}
      </button>
    </template>
  </AppModal>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useUiStore } from '../../stores/ui.js'
import { useChatStore } from '../../stores/chat.js'
import { useContactsStore } from '../../stores/contacts.js'
import * as contactsApi from '../../api/contacts.js'
import AppModal from '../common/AppModal.vue'
import AppAvatar from '../common/AppAvatar.vue'
import AppEmptyState from '../common/AppEmptyState.vue'

const ui = useUiStore()
const chat = useChatStore()
const contacts = useContactsStore()
const selected = ref(new Set())
const loading = ref(false)
const groupMembers = ref([])

onMounted(async () => {
  if (!chat.activeConversation) return
  try {
    const { data } = await contactsApi.getGroupMembers(chat.activeConversation.id)
    if (data.code === 0) groupMembers.value = data.data || []
  } catch { /* ignore */ }
})

const memberIds = computed(() => new Set(groupMembers.value.map(m => m.id)))
const availableFriends = computed(() =>
  contacts.friends.filter(f => !memberIds.value.has(f.id))
)

function getAvatar(friend) {
  return friend.avatar || `/static/avatar/${friend.id}.jpg`
}

function toggle(id) {
  if (selected.value.has(id)) selected.value.delete(id)
  else selected.value.add(id)
  selected.value = new Set(selected.value) // trigger reactivity
}

async function confirm() {
  if (selected.value.size === 0) return
  loading.value = true
  try {
    const { data } = await contactsApi.inviteMembers(
      chat.activeConversation.id,
      Array.from(selected.value)
    )
    if (data.code === 0) {
      ui.showToast(data.data?.message || '邀请成功', 'success')
      ui.closeModal()
    } else {
      ui.showToast(data.message || '邀请失败', 'error')
    }
  } catch (e) { ui.showToast(e.message, 'error') }
  finally { loading.value = false }
}
</script>

<style scoped>
.close-btn { background: none; border: none; font-size: 24px; color: var(--text-hint); padding: 0; }
.invite-item {
  display: flex; align-items: center; gap: 12px;
  padding: 10px 0; border-bottom: 1px solid var(--border-light);
  cursor: pointer;
}
.invite-item.selected { background: var(--primary-light); margin: 0 -16px; padding-left: 16px; padding-right: 16px; }
.invite-name { flex: 1; }
.check-mark { color: var(--primary); font-weight: 700; }
.btn { padding: 12px; border: none; border-radius: var(--radius-md); font-size: var(--font-md); cursor: pointer; background: var(--bg-input); }
.btn-primary { background: var(--primary); color: #fff; }
.btn:disabled { opacity: 0.6; }
.flex-1 { flex: 1; }
</style>
