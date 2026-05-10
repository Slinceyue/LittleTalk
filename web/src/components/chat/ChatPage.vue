<template>
  <div class="chat-page">
    <ChatHeader
      :name="conversation?.name"
      :online="isFriendOnline"
      :isGroup="isGroup"
      @back="handleBack"
      @search="showSearch = !showSearch"
      @bg="handleBg"
      @bubble="handleBubble"
    />

    <MessageSearch v-if="showSearch" :conv-key="convKey" @close="showSearch = false" />

    <div class="chat-messages" ref="msgList" @scroll="onScroll" :style="bgStyle">
      <div v-if="!messages.length" class="empty-chat">
        <div class="empty-icon">💭</div>
        <div class="empty-text">暂无消息，开始聊天吧</div>
      </div>

      <div
        v-for="(msg, i) in messages"
        :key="msg.msg_id || i"
        class="msg-wrapper"
      >
        <!-- Time divider -->
        <div v-if="showTimeDivider(i)" class="time-divider">
          {{ formatTime(msg.send_time) }}
        </div>

        <ChatMessageBubble
          :msg="msg"
          :isOwn="msg.from_id === auth.userId"
          :senderName="getSenderName(msg)"
          :senderAvatar="getSenderAvatar(msg)"
          :styleSettings="bubbleStyleSettings"
          @preview="handlePreview"
        />
      </div>

      <TypingIndicator v-if="typingNames.length" :names="typingNames" />

      <div ref="msgEnd"></div>
    </div>

    <ChatInput
      @send="handleSend"
      @typing="handleTyping"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useUiStore } from '../../stores/ui.js'
import { useAuthStore } from '../../stores/auth.js'
import { useChatStore } from '../../stores/chat.js'
import { useContactsStore } from '../../stores/contacts.js'
import { useWsStore } from '../../stores/websocket.js'
import { WS_TYPE } from '../../utils/constants.js'
import { formatTime } from '../../utils/formatters.js'
import { getAvatarUrl } from '../../utils/avatar.js'
import ChatHeader from './ChatHeader.vue'
import ChatMessageBubble from './ChatMessageBubble.vue'
import ChatInput from './ChatInput.vue'
import TypingIndicator from './TypingIndicator.vue'
import MessageSearch from './MessageSearch.vue'

const ui = useUiStore()
const auth = useAuthStore()
const chat = useChatStore()
const contacts = useContactsStore()
const ws = useWsStore()

const msgList = ref(null)
const msgEnd = ref(null)
const showSearch = ref(false)
const isNearBottom = ref(true)

const conversation = computed(() => chat.activeConversation)
const isGroup = computed(() => conversation.value?.type === 'group')
const convKey = computed(() => {
  if (!conversation.value) return ''
  return `${conversation.value.type}_${conversation.value.id}`
})
const messages = computed(() => {
  return chat.messageMap[convKey.value] || []
})

const isFriendOnline = computed(() => {
  if (isGroup.value) return false
  return contacts.isOnline(conversation.value?.id)
})

const typingNames = computed(() => {
  const users = chat.typingUsers[convKey.value] || []
  return users.map(id => contacts.getUserName(id)).filter(Boolean)
})

const bubbleStyleSettings = computed(() => {
  if (isGroup.value || !conversation.value) return {}
  return chat.getBubbleStyle(conversation.value.id)
})

const bgStyle = computed(() => {
  if (isGroup.value || !conversation.value) return {}
  const bg = chat.getBackground(conversation.value.id)
  if (!bg) return {}
  if (bg.startsWith('data:')) {
    return { backgroundImage: `url(${bg})`, backgroundSize: 'cover', backgroundPosition: 'center' }
  }
  return { background: bg }
})

function getSenderName(msg) {
  if (!isGroup.value) return ''
  return msg.from_name || contacts.getUserName(msg.from_id)
}

function getSenderAvatar(msg) {
  if (!isGroup.value) return ''
  return msg.from_avatar || getAvatarUrl(msg.from_id, '')
}

function showTimeDivider(i) {
  if (i === 0) return true
  const curr = messages.value[i]?.send_time
  const prev = messages.value[i - 1]?.send_time
  if (!curr || !prev) return false
  return (curr - prev) > 300000 // 5 minutes
}

function scrollToBottom(smooth = false) {
  nextTick(() => {
    if (msgEnd.value) {
      msgEnd.value.scrollIntoView({ behavior: smooth ? 'smooth' : 'instant' })
    }
  })
}

function onScroll() {
  if (!msgList.value) return
  const { scrollTop, scrollHeight, clientHeight } = msgList.value
  isNearBottom.value = (scrollHeight - scrollTop - clientHeight) < 60
}

// Handle send from ChatInput
function handleSend(data) {
  if (!conversation.value) return
  const msgId = 'msg_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
  const sendTime = Date.now()

  if (isGroup.value) {
    const localMsg = {
      msg_id: msgId,
      from_id: auth.userId,
      from_name: auth.username,
      from_avatar: auth.avatar,
      room_id: conversation.value.id,
      content: data.content,
      file_url: data.fileUrl || '',
      file_name: data.fileName || '',
      message_type: data.messageType || 1,
      send_time: sendTime,
    }
    chat.appendMessage(convKey.value, localMsg)
    scrollToBottom()

    ws.send({
      type: 'group_talk',
      room_id: conversation.value.id,
      content: data.content,
      message_type: data.messageType || 1,
    })

    chat.updateRecentChat({
      type: 'group',
      friend_id: conversation.value.id,
      friend_name: conversation.value.name,
      friend_avatar: '',
      last_message: data.content || data.fileName || (data.fileUrl ? '[文件]' : ''),
      send_time: sendTime,
    })
  } else {
    const localMsg = {
      msg_id: msgId,
      from_id: auth.userId,
      to_id: conversation.value.id,
      content: data.content,
      file_url: data.fileUrl || '',
      file_name: data.fileName || '',
      message_type: data.messageType || 1,
      send_time: sendTime,
    }
    chat.appendMessage(convKey.value, localMsg)
    scrollToBottom()

    ws.send({
      msg_id: msgId,
      from_id: auth.userId,
      to_id: conversation.value.id,
      room_id: 0,
      message_type: data.messageType || 1,
      content: data.content,
      file_url: data.fileUrl || '',
      file_name: data.fileName || '',
      file_id: 0,
    })

    // Update recent chat
    chat.updateRecentChat({
      type: 'friend',
      friend_id: conversation.value.id,
      friend_name: conversation.value.name,
      friend_avatar: contacts.getUserAvatar(conversation.value.id),
      last_message: data.content || data.fileName || (data.fileUrl ? '[文件]' : ''),
      send_time: sendTime,
    })
  }
}

function handleTyping(isTyping) {
  if (isGroup.value || !conversation.value) return
  ws.send({ type: WS_TYPE.TYPING, to_id: conversation.value.id, typing: isTyping })
}

function handleBack() {
  chat.closeConversation()
  ui.navigateTo('main')
}

function handlePreview(url) {
  ui.openModal('imagePreview', { url })
}

function handleBg() {
  if (conversation.value && !isGroup.value) {
    ui.openModal('chatBackground', { friendId: conversation.value.id })
  }
}

function handleBubble() {
  if (conversation.value && !isGroup.value) {
    ui.openModal('chatBubble', { friendId: conversation.value.id })
  }
}

// Load history when conversation changes
watch(convKey, (key) => {
  if (key) {
    chat.loadHistory(key)
    nextTick(() => scrollToBottom())
  }
}, { immediate: true })

onMounted(() => {
  if (convKey.value) {
    chat.loadHistory(convKey.value)
    nextTick(() => scrollToBottom())
  }
})
</script>

<style scoped>
.chat-page {
  width: 100%; height: 100%;
  display: flex; flex-direction: column;
  background: var(--bg-page);
}
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
  -webkit-overflow-scrolling: touch;
}
.empty-chat {
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  height: 100%; color: var(--text-hint);
}
.empty-icon { font-size: 48px; margin-bottom: 8px; opacity: 0.6; }
.empty-text { font-size: var(--font-sm); }
.time-divider {
  text-align: center;
  margin: 16px 0;
  font-size: var(--font-xs);
  color: var(--text-hint);
}
.msg-wrapper { margin-bottom: 4px; }
</style>
