<template>
  <div class="app-root" :data-theme="ui.theme">
    <Transition name="slide-right">
      <LoginPage v-if="ui.currentPage === 'login'" />
      <MainPage v-else-if="ui.currentPage === 'main'" />
      <ChatPage v-else-if="ui.currentPage === 'chat'" />
      <AddFriendPage v-else-if="ui.currentPage === 'addFriend'" />
    </Transition>

    <!-- Global modals -->
    <FriendRequestModal v-if="ui.activeModal === 'friendRequest'" />
    <ChatBackgroundPicker v-if="ui.activeModal === 'chatBackground'" />
    <ChatBubbleSettings v-if="ui.activeModal === 'chatBubble'" />
    <GroupInfoModal v-if="ui.activeModal === 'groupInfo'" />
    <GroupMemberModal v-if="ui.activeModal === 'groupMembers'" />
    <CreateGroupModal v-if="ui.activeModal === 'createGroup'" />
    <JoinGroupModal v-if="ui.activeModal === 'joinGroup'" />
    <InviteMemberModal v-if="ui.activeModal === 'inviteMembers'" />

    <!-- Global overlays -->
    <ImagePreview v-if="ui.activeModal === 'imagePreview'" />

    <!-- Toast container -->
    <div class="toast-container">
      <TransitionGroup name="toast">
        <div
          v-for="toast in ui.toasts"
          :key="toast.id"
          class="toast-item"
          :class="'toast-' + toast.type"
          @click="ui.dismissToast(toast.id)"
        >
          {{ toast.message }}
        </div>
      </TransitionGroup>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useUiStore } from './stores/ui.js'
import { useAuthStore } from './stores/auth.js'
import { useContactsStore } from './stores/contacts.js'
import { useChatStore } from './stores/chat.js'
import { useWsStore } from './stores/websocket.js'
import LoginPage from './components/auth/LoginPage.vue'
import MainPage from './components/layout/MainPage.vue'
import ChatPage from './components/chat/ChatPage.vue'
import AddFriendPage from './components/contacts/AddFriendPage.vue'
import FriendRequestModal from './components/contacts/FriendRequestModal.vue'
import ChatBackgroundPicker from './components/chat/ChatBackgroundPicker.vue'
import ChatBubbleSettings from './components/chat/ChatBubbleSettings.vue'
import GroupInfoModal from './components/group/GroupInfoModal.vue'
import GroupMemberModal from './components/group/GroupMemberModal.vue'
import CreateGroupModal from './components/group/CreateGroupModal.vue'
import JoinGroupModal from './components/group/JoinGroupModal.vue'
import InviteMemberModal from './components/group/InviteMemberModal.vue'
import ImagePreview from './components/chat/ImagePreview.vue'

const ui = useUiStore()
const auth = useAuthStore()
const contacts = useContactsStore()
const chat = useChatStore()
const ws = useWsStore()

onMounted(async () => {
  ui.applyTheme()
  setupWsHandlers()
  const ok = await auth.autoLogin()
  if (ok) {
    ui.navigateTo('main')
    contacts.initFriendsFromCache()
    contacts.initGroupsFromCache()
    chat.loadRecentChats()
    contacts.fetchFriends()
    contacts.fetchGroups()
    contacts.fetchFriendRequests()
    ws.connect()
  }
})

function setupWsHandlers() {
  ws.on('online_status', (msg) => {
    contacts.setOnlineStatus(msg.user_id, msg.online)
  })
  ws.on('batch_online_status', (msg) => {
    contacts.batchSetOnlineStatus(msg.statuses || [])
  })
  ws.on('talk', (msg) => {
    handleIncomingMessage('friend', msg.data)
  })
  ws.on('group_talk', (msg) => {
    handleIncomingMessage('group', msg.data)
  })
  ws.on('friend', (msg) => {
    const fromId = msg.data?.from_id
    if (fromId) {
      contacts.fetchUsersInfo([fromId]).then(() => {
        ui.showToast(`${contacts.getUserName(fromId)} 请求添加你为好友`, 'info', 5000)
      })
      contacts.fetchFriendRequests()
    }
  })
  ws.on('typing', (msg) => {
    const fromId = msg.from_id
    if (fromId && fromId !== auth.userId) {
      const key = chat.activeKey
      if (key && chat.activeConversation) {
        const convId = chat.activeConversation.type === 'friend'
          ? chat.activeConversation.id
          : 0
        if (fromId === convId) {
          chat.setTyping(key, fromId, msg.typing)
        }
      }
    }
  })
  ws.on('send_failed', (msg) => {
    const key = chat.activeKey
    if (key && msg.msg_id) chat.markMessageFailed(key, msg.msg_id)
  })
}

function handleIncomingMessage(type, msgData) {
  const id = type === 'friend' ? msgData.from_id : msgData.room_id
  // Skip system messages (from_id === 0)
  if (!id || id === 0) return
  // Skip empty messages (no content and no file)
  if (!msgData.content && !msgData.file_url) return

  const key = `${type}_${id}`
  chat.loadHistory(key)
  const added = chat.appendMessage(key, msgData)
  if (!added) return

  // Update recent chats
  const name = type === 'friend'
    ? contacts.getUserName(id)
    : (contacts.getGroupById(id)?.name || '群聊')
  const avatar = type === 'friend'
    ? contacts.getUserAvatar(id)
    : (contacts.getGroupById(id)?.avatar || '')
  chat.updateRecentChat({
    type: type,
    friend_id: id,
    friend_name: name,
    friend_avatar: avatar,
    last_message: msgData.content || msgData.file_name || (msgData.file_url ? '[文件]' : ''),
    send_time: msgData.send_time || Date.now(),
  })

  // If not in this conversation, add unread
  if (!chat.activeConversation || chat.activeKey !== key) {
    chat.addUnread(key)
  }
}
</script>

<style scoped>
.app-root {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-page);
  position: relative;
  overflow: hidden;
}

.toast-container {
  position: fixed;
  top: 60px;
  right: 12px;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-width: 300px;
  pointer-events: none;
}

.toast-item {
  pointer-events: auto;
  padding: 10px 16px;
  border-radius: var(--radius-md);
  font-size: var(--font-sm);
  color: #fff;
  cursor: pointer;
  box-shadow: var(--shadow-md);
  word-break: break-all;
}
.toast-info { background: #323232; }
.toast-success { background: var(--primary); }
.toast-error { background: var(--danger); }
</style>
