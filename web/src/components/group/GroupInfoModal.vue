<template>
  <AppModal @close="ui.closeModal()">
    <template #header>
      <h3>群聊信息</h3>
      <button class="close-btn" @click="ui.closeModal()">×</button>
    </template>
    <div class="info-body" v-if="group">
      <AppAvatar :src="group.avatar" :name="group.name" size="xl" :bg="'linear-gradient(135deg, #43e97b, #38f9d7)'" />
      <div class="info-name">{{ group.name }}</div>
      <p class="info-intro">{{ group.intro || '暂无群介绍' }}</p>
      <p class="info-meta">{{ group.member_cnt }} 位成员</p>
    </div>
    <div class="info-actions">
      <button class="btn btn-primary" @click="ui.openModal('groupMembers')">管理成员</button>
      <button class="btn btn-success" @click="ui.openModal('inviteMembers')">邀请好友</button>
      <button v-if="isOwner" class="btn btn-danger" @click="handleDismiss">解散群聊</button>
      <button v-else class="btn btn-danger" @click="handleQuit">退出群聊</button>
    </div>
  </AppModal>
</template>

<script setup>
import { computed } from 'vue'
import { useUiStore } from '../../stores/ui.js'
import { useContactsStore } from '../../stores/contacts.js'
import { useChatStore } from '../../stores/chat.js'
import { useAuthStore } from '../../stores/auth.js'
import AppModal from '../common/AppModal.vue'
import AppAvatar from '../common/AppAvatar.vue'

const ui = useUiStore()
const contacts = useContactsStore()
const chat = useChatStore()
const auth = useAuthStore()

const group = computed(() => {
  if (!chat.activeConversation) return null
  return contacts.getGroupById(chat.activeConversation.id)
})

const isOwner = computed(() => group.value?.owner_id === auth.userId)

async function handleQuit() {
  if (!confirm('确定退出该群聊？')) return
  try {
    await contacts.quitGroup(chat.activeConversation.id)
    ui.closeModal()
    ui.showToast('已退出群聊', 'success')
    ui.navigateTo('main')
  } catch (e) { ui.showToast(e.message, 'error') }
}

async function handleDismiss() {
  if (!confirm('确定解散该群聊？此操作不可恢复！')) return
  try {
    await contacts.dismissGroup(chat.activeConversation.id)
    ui.closeModal()
    ui.showToast('已解散群聊', 'success')
    ui.navigateTo('main')
  } catch (e) { ui.showToast(e.message, 'error') }
}
</script>

<style scoped>
.close-btn { background: none; border: none; font-size: 24px; color: var(--text-hint); padding: 0; }
.info-body { display: flex; flex-direction: column; align-items: center; padding: 8px 0; }
.info-name { font-size: var(--font-xl); font-weight: 600; margin: 16px 0 8px; }
.info-intro { font-size: var(--font-sm); color: var(--text-hint); margin-bottom: 8px; }
.info-meta { font-size: var(--font-sm); color: var(--text-secondary); }
.info-actions {
  display: flex; flex-direction: column; gap: 8px;
  margin-top: 16px; padding-top: 16px;
  border-top: 1px solid var(--border-light);
}
.btn { padding: 12px; border: none; border-radius: var(--radius-md); font-size: var(--font-md); }
.btn-primary { background: var(--primary); color: #fff; }
.btn-success { background: #1989fa; color: #fff; }
.btn-danger { background: var(--danger); color: #fff; }
</style>
