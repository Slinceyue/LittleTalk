<template>
  <div>
    <!-- Friend requests section -->
    <div class="section-header" @click="ui.openModal('friendRequest')">
      <span class="section-label">好友请求</span>
      <span class="section-badge" v-if="contacts.friendRequests.length">
        {{ contacts.friendRequests.length }}
      </span>
      <span class="section-arrow">›</span>
    </div>

    <!-- Friend list -->
    <div class="section-title">好友</div>
    <AppEmptyState v-if="contacts.friends.length === 0" icon="👥" text="暂无好友，去添加吧" />

    <div
      v-for="friend in contacts.friends"
      :key="friend.id"
      class="friend-item"
      @click="openChat(friend)"
    >
      <div class="friend-avatar" @click.stop="showInfo(friend)">
        <AppAvatar :src="getAvatar(friend)" :name="friend.username" />
        <span v-if="contacts.isOnline(friend.id)" class="online-dot"></span>
      </div>
      <div class="friend-body">
        <span class="friend-name">{{ friend.username }}</span>
        <span class="friend-status" :class="contacts.isOnline(friend.id) ? 'online' : ''">
          {{ contacts.isOnline(friend.id) ? '在线' : '离线' }}
        </span>
      </div>
      <button class="friend-del" @click.stop="handleDelete(friend)" title="删除好友">×</button>
    </div>

    <!-- Friend info panel -->
    <FriendInfoPanel
      v-if="selectedFriend"
      :friend="selectedFriend"
      @close="selectedFriend = null"
      @chat="selectedFriend = null; openChat($event)"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useContactsStore } from '../../stores/contacts.js'
import { useUiStore } from '../../stores/ui.js'
import { useAuthStore } from '../../stores/auth.js'
import { useChatStore } from '../../stores/chat.js'
import AppAvatar from '../common/AppAvatar.vue'
import AppEmptyState from '../common/AppEmptyState.vue'
import FriendInfoPanel from './FriendInfoPanel.vue'
import * as authApi from '../../api/auth.js'

const contacts = useContactsStore()
const ui = useUiStore()
const auth = useAuthStore()
const chat = useChatStore()
const selectedFriend = ref(null)

function getAvatar(friend) {
  return friend.avatar || `/static/avatar/${friend.id}.jpg`
}

async function showInfo(friend) {
  try {
    const { data } = await authApi.getOtherInfo(friend.id)
    if (data.code === 0) {
      selectedFriend.value = data.data
    } else {
      console.warn('[FriendList] showInfo API error:', data.message)
    }
  } catch (e) {
    console.warn('[FriendList] showInfo failed:', e)
  }
}

function openChat(friend) {
  chat.openConversation('friend', friend.id, friend.username)
  ui.navigateTo('chat')
}

async function handleDelete(friend) {
  if (!confirm(`确定删除好友 ${friend.username}？`)) return
  try {
    await contacts.deleteFriend(friend.id)
    ui.showToast('已删除好友', 'success')
  } catch (e) {
    ui.showToast(e.message, 'error')
  }
}
</script>

<style scoped>
.section-header {
  display: flex; align-items: center; gap: 8px;
  padding: 14px 16px;
  background: var(--bg-white);
  border-bottom: 1px solid var(--border-light);
  cursor: pointer;
}
.section-label { font-size: var(--font-md); flex: 1; }
.section-badge {
  background: var(--danger); color: #fff;
  font-size: 11px; min-width: 20px; height: 20px;
  border-radius: 10px; display: flex; align-items: center; justify-content: center;
  padding: 0 5px;
}
.section-arrow { color: var(--text-hint); font-size: 20px; }
.section-title {
  padding: 10px 16px 6px;
  font-size: var(--font-xs); color: var(--text-hint);
}

.friend-item {
  display: flex; align-items: center; gap: 12px;
  padding: 12px 16px;
  background: var(--bg-white);
  cursor: pointer;
  border-bottom: 1px solid var(--border-light);
}
.friend-item:active { background: var(--bg-hover); }
.friend-avatar { position: relative; }
.online-dot {
  position: absolute; bottom: 2px; right: 2px;
  width: 10px; height: 10px;
  border-radius: 50%; background: var(--primary);
  border: 2px solid var(--bg-white);
}
.friend-body { flex: 1; }
.friend-name { font-size: var(--font-md); font-weight: 500; display: block; }
.friend-status {
  font-size: var(--font-xs); color: var(--text-hint);
}
.friend-status.online { color: var(--primary); }
.friend-del {
  background: none; border: none;
  font-size: 18px; color: var(--text-hint);
  padding: 4px;
}
</style>
