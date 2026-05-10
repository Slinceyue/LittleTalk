<template>
  <AppModal @close="$emit('close')">
    <template #header>
      <button class="close-btn" @click="$emit('close')">×</button>
      <h3>好友信息</h3>
      <div style="width:24px"></div>
    </template>
    <div class="info-body">
      <AppAvatar :src="friend.avatar" :name="friend.username" size="xl" />
      <div class="info-name">{{ friend.username }}</div>
      <div class="info-details">
        <div class="info-row"><span>性别</span><span>{{ sexText }}</span></div>
        <div class="info-row"><span>生日</span><span>{{ friend.birthday || '未设置' }}</span></div>
        <div class="info-row"><span>简介</span><span>{{ friend.intro || '这个人很懒，什么都没写~' }}</span></div>
      </div>
      <button class="chat-btn" @click="$emit('chat', friend)">发消息</button>
    </div>
  </AppModal>
</template>

<script setup>
import { computed } from 'vue'
import AppModal from '../common/AppModal.vue'
import AppAvatar from '../common/AppAvatar.vue'
import { SEX_MAP } from '../../utils/constants.js'

const props = defineProps({ friend: { type: Object, default: () => ({}) } })
defineEmits(['close', 'chat'])
const sexText = computed(() => SEX_MAP[props.friend.sex] || '未知')
</script>

<style scoped>
.close-btn { background: none; border: none; font-size: 24px; color: var(--text-hint); padding: 0; }
.info-body { display: flex; flex-direction: column; align-items: center; padding: 8px 0; }
.info-name { font-size: var(--font-xl); font-weight: 600; margin: 16px 0; }
.info-details {
  width: 100%;
  background: var(--bg-page);
  border-radius: var(--radius-md);
  padding: 12px 16px;
  margin-bottom: 20px;
}
.info-row {
  display: flex; justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-light);
  font-size: var(--font-sm);
}
.info-row:last-child { border-bottom: none; }
.info-row span:first-child { color: var(--text-secondary); }
.info-row span:last-child { color: var(--text-primary); }
.chat-btn {
  width: 100%; padding: 12px;
  background: var(--primary); color: #fff;
  border: none; border-radius: var(--radius-md);
  font-size: var(--font-md);
}
.chat-btn:active { opacity: 0.8; }
</style>
