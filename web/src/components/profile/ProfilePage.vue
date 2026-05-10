<template>
  <div class="profile">
    <div class="profile-card" @click="showEdit = true">
      <AppAvatar :src="auth.avatar" :name="auth.username" size="xl" />
      <div class="profile-info">
        <div class="profile-name">{{ auth.username }}</div>
        <div class="profile-id">ID: {{ auth.userId }}</div>
      </div>
      <span class="profile-hint">编辑</span>
    </div>

    <div class="profile-details">
      <div class="detail-row">
        <span class="detail-label">性别</span>
        <span class="detail-value">{{ sexText }}</span>
      </div>
      <div class="detail-row">
        <span class="detail-label">生日</span>
        <span class="detail-value">{{ auth.user?.birthday || '未设置' }}</span>
      </div>
      <div class="detail-row">
        <span class="detail-label">注册时间</span>
        <span class="detail-value">{{ createdTime }}</span>
      </div>
    </div>

    <div class="profile-menu">
      <div class="menu-item" @click="handleAvatarUpload">
        <span>修改头像</span>
        <span class="menu-arrow">›</span>
      </div>
      <div class="menu-item" @click="handleLogout">
        <span>退出登录</span>
        <span class="menu-arrow">›</span>
      </div>
    </div>

    <!-- Edit panel (modal) -->
    <ProfileEditPanel v-if="showEdit" @close="showEdit = false" @saved="showEdit = false" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useAuthStore } from '../../stores/auth.js'
import { useUiStore } from '../../stores/ui.js'
import { useChatStore } from '../../stores/chat.js'
import { useContactsStore } from '../../stores/contacts.js'
import { useWsStore } from '../../stores/websocket.js'
import { SEX_MAP, MAX_AVATAR_SIZE } from '../../utils/constants.js'
import AppAvatar from '../common/AppAvatar.vue'
import ProfileEditPanel from './ProfileEditPanel.vue'

const auth = useAuthStore()
const ui = useUiStore()
const chat = useChatStore()
const contacts = useContactsStore()
const ws = useWsStore()
const showEdit = ref(false)

const sexText = computed(() => SEX_MAP[auth.user?.sex] || '未知')
const createdTime = computed(() => {
  const ts = auth.user?.created_at
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleDateString('zh-CN')
})

function handleAvatarUpload() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/jpeg,image/png,image/gif,image/webp'
  input.onchange = async (e) => {
    const file = e.target.files[0]
    if (!file) return
    if (file.size > MAX_AVATAR_SIZE) { ui.showToast('图片大小不能超过2MB', 'error'); return }
    try {
      await auth.uploadAvatar(file)
      ui.showToast('头像上传成功', 'success')
    } catch (e) {
      ui.showToast(e.message || '上传失败', 'error')
    }
  }
  input.click()
}

async function handleLogout() {
  chat.clearAll(auth.userId)
  contacts.clearAll(auth.userId)
  await auth.logout()
  ws.disconnect()
  ui.navigateTo('login')
}
</script>

<style scoped>
.profile { padding: 16px; }
.profile-card {
  display: flex; align-items: center; gap: 16px;
  background: var(--bg-white);
  padding: 20px;
  border-radius: var(--radius-lg);
  margin-bottom: 16px;
  cursor: pointer;
}
.profile-info { flex: 1; }
.profile-name { font-size: var(--font-lg); font-weight: 600; }
.profile-id { font-size: var(--font-sm); color: var(--text-hint); margin-top: 4px; }
.profile-hint { color: var(--text-hint); font-size: var(--font-sm); }
.profile-details {
  background: var(--bg-white);
  border-radius: var(--radius-lg);
  margin-bottom: 16px;
  overflow: hidden;
}
.detail-row {
  display: flex; justify-content: space-between;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-light);
}
.detail-row:last-child { border-bottom: none; }
.detail-label { color: var(--text-secondary); font-size: var(--font-sm); }
.detail-value { color: var(--text-primary); font-size: var(--font-sm); }
.profile-menu {
  background: var(--bg-white);
  border-radius: var(--radius-lg);
  overflow: hidden;
}
.menu-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-light);
  cursor: pointer;
  font-size: var(--font-sm);
}
.menu-item:last-child { border-bottom: none; }
.menu-item:active { background: var(--bg-hover); }
.menu-arrow { color: var(--text-hint); font-size: 18px; }
</style>
