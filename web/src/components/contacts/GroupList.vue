<template>
  <div>
    <div class="group-actions">
      <button class="action-btn" @click="ui.openModal('createGroup')">+ 创建群聊</button>
      <button class="action-btn" @click="ui.openModal('joinGroup')">🔍 加入群聊</button>
    </div>
    <AppEmptyState v-if="contacts.groups.length === 0" icon="👨‍👩‍👧" text="暂无群聊" />
    <div
      v-for="group in contacts.groups"
      :key="group.id"
      class="group-item"
      @click="openChat(group)"
    >
      <AppAvatar :src="group.avatar" :name="group.name" :bg="'linear-gradient(135deg, #43e97b, #38f9d7)'" />
      <div class="group-body">
        <span class="group-name">{{ group.name }}</span>
        <span class="group-meta">{{ group.member_cnt }} 位成员</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useContactsStore } from '../../stores/contacts.js'
import { useUiStore } from '../../stores/ui.js'
import { useChatStore } from '../../stores/chat.js'
import AppAvatar from '../common/AppAvatar.vue'
import AppEmptyState from '../common/AppEmptyState.vue'

const contacts = useContactsStore()
const ui = useUiStore()
const chat = useChatStore()

function openChat(group) {
  chat.openConversation('group', group.id, group.name)
  ui.navigateTo('chat')
}
</script>

<style scoped>
.group-actions {
  display: flex; gap: 12px;
  padding: 12px 16px;
}
.action-btn {
  flex: 1;
  padding: 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-white);
  color: var(--text-primary);
  font-size: var(--font-sm);
  cursor: pointer;
}
.action-btn:active { background: var(--bg-hover); }
.group-item {
  display: flex; align-items: center; gap: 12px;
  padding: 12px 16px;
  background: var(--bg-white);
  cursor: pointer;
  border-bottom: 1px solid var(--border-light);
}
.group-item:active { background: var(--bg-hover); }
.group-body { flex: 1; }
.group-name { font-weight: 500; display: block; }
.group-meta { font-size: var(--font-xs); color: var(--text-hint); }
</style>
