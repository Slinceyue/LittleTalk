<template>
  <div class="main-page">
    <AppHeader @addFriend="ui.navigateTo('addFriend')" @themeToggle="ui.toggleTheme()" />

    <div class="main-content">
      <ConversationList v-show="ui.activeTab === 'messages'" />
      <FriendList v-show="ui.activeTab === 'friends'" />
      <GroupList v-show="ui.activeTab === 'groups'" />
      <ProfilePage v-show="ui.activeTab === 'profile'" />
    </div>

    <TabBar
      :activeTab="ui.activeTab"
      :tabs="tabs"
      @switch="ui.switchTab($event)"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useUiStore } from '../../stores/ui.js'
import { useContactsStore } from '../../stores/contacts.js'
import { useChatStore } from '../../stores/chat.js'
import AppHeader from './AppHeader.vue'
import TabBar from './TabBar.vue'
import ConversationList from '../conversation/ConversationList.vue'
import FriendList from '../contacts/FriendList.vue'
import GroupList from '../contacts/GroupList.vue'
import ProfilePage from '../profile/ProfilePage.vue'

const ui = useUiStore()
const contacts = useContactsStore()
const chat = useChatStore()

const tabs = computed(() => [
  { id: 'messages', icon: '💬', label: '消息', badge: chat.totalUnread },
  { id: 'friends', icon: '👥', label: '好友', badge: contacts.friendRequests.length },
  { id: 'groups', icon: '👨‍👩‍👧', label: '群聊', badge: 0 },
  { id: 'profile', icon: '👤', label: '我的', badge: 0 },
])
</script>

<style scoped>
.main-page {
  width: 100%; height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-page);
}
.main-content {
  flex: 1;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}
</style>
