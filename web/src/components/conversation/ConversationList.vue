<template>
  <div>
    <div class="section-title">最近联系人</div>
    <AppEmptyState v-if="chats.length === 0" icon="💬" text="暂无消息，点击好友开始聊天" />
    <div
      v-for="chat in chats"
      :key="(chat.type || 'friend') + '_' + chat.friend_id"
      class="conv-item"
      @click="openChat(chat)"
    >
      <div class="conv-avatar">
        <AppAvatar :src="chat.friend_avatar" :name="chat.friend_name" />
        <span v-if="chat.type !== 'group' && contacts.isOnline(chat.friend_id)" class="online-dot"></span>
      </div>
      <div class="conv-body">
        <div class="conv-top">
          <span class="conv-name">{{ chat.friend_name }}</span>
          <span class="conv-time">{{ formatTime(chat.send_time) }}</span>
        </div>
        <div class="conv-preview truncate">{{ chat.last_message }}</div>
      </div>
      <span v-if="unreadCount(chat)" class="conv-badge">
        {{ unreadCount(chat) > 99 ? '99+' : unreadCount(chat) }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useChatStore } from '../../stores/chat.js'
import { useContactsStore } from '../../stores/contacts.js'
import { useUiStore } from '../../stores/ui.js'
import AppAvatar from '../common/AppAvatar.vue'
import AppEmptyState from '../common/AppEmptyState.vue'
import { formatTime } from '../../utils/formatters.js'

const chat = useChatStore()
const contacts = useContactsStore()
const ui = useUiStore()

const chats = computed(() => chat.recentChats)

function unreadCount(chat) {
  const type = chat.type || 'friend'
  return chat.unreadCounts[`${type}_${chat.friend_id}`] || 0
}

function openChat(chatItem) {
  const type = chatItem.type || 'friend'
  chat.openConversation(type, chatItem.friend_id, chatItem.friend_name)
  ui.navigateTo('chat')
}
</script>

<style scoped>
.section-title {
  padding: 12px 16px 8px;
  font-size: var(--font-xs);
  color: var(--text-hint);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.conv-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--bg-white);
  cursor: pointer;
  border-bottom: 1px solid var(--border-light);
  transition: background 0.15s;
}
.conv-item:active { background: var(--bg-hover); }
.conv-avatar { position: relative; }
.online-dot {
  position: absolute;
  bottom: 2px; right: 2px;
  width: 10px; height: 10px;
  border-radius: 50%;
  background: var(--primary);
  border: 2px solid var(--bg-white);
}
.conv-body { flex: 1; min-width: 0; }
.conv-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.conv-name { font-size: var(--font-md); font-weight: 500; }
.conv-time { font-size: var(--font-xs); color: var(--text-hint); flex-shrink: 0; }
.conv-preview { font-size: var(--font-sm); color: var(--text-hint); }
.conv-badge {
  min-width: 20px; height: 20px;
  border-radius: 10px;
  background: var(--danger);
  color: #fff;
  font-size: 11px;
  display: flex; align-items: center; justify-content: center;
  padding: 0 5px;
  flex-shrink: 0;
}
</style>
