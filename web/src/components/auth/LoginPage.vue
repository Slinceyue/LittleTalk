<template>
  <div class="login-page">
    <div class="login-container">
      <!-- Logo -->
      <div class="login-logo" v-if="!isRegister">
        <div class="logo-icon">💬</div>
        <h1 class="logo-title">LittleTalk</h1>
      </div>

      <!-- Login form -->
      <div v-if="!isRegister" class="form">
        <div class="form-group">
          <label>用户名</label>
          <input v-model="username" type="text" placeholder="请输入用户名" @keyup.enter="handleLogin" />
        </div>
        <div class="form-group">
          <label>密码</label>
          <input v-model="password" type="password" placeholder="请输入密码" @keyup.enter="handleLogin" />
        </div>
        <button class="btn btn-primary" @click="handleLogin" :disabled="loading">
          {{ loading ? '登录中...' : '登录' }}
        </button>
        <p class="form-error" v-if="error">{{ error }}</p>
      </div>

      <!-- Register form -->
      <div v-else class="form">
        <div class="form-group">
          <label>用户名</label>
          <input v-model="regUsername" type="text" placeholder="2-20位字符" />
        </div>
        <div class="form-group">
          <label>密码</label>
          <input v-model="regPassword" type="password" placeholder="至少6位" />
        </div>
        <div class="form-group">
          <label>性别</label>
          <select v-model="regSex">
            <option :value="0">未知</option>
            <option :value="1">男</option>
            <option :value="2">女</option>
          </select>
        </div>
        <div class="form-group">
          <label>生日</label>
          <input v-model="regBirthday" type="date" />
        </div>
        <button class="btn btn-primary" @click="handleRegister" :disabled="loading">
          {{ loading ? '注册中...' : '注册' }}
        </button>
        <p class="form-error" v-if="error">{{ error }}</p>
      </div>

      <!-- Switch -->
      <p class="switch-link">
        <template v-if="!isRegister">
          还没有账号？<a href="#" @click.prevent="switchMode">立即注册</a>
        </template>
        <template v-else>
          已有账号？<a href="#" @click.prevent="switchMode">立即登录</a>
        </template>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../../stores/auth.js'
import { useUiStore } from '../../stores/ui.js'
import { useContactsStore } from '../../stores/contacts.js'
import { useChatStore } from '../../stores/chat.js'
import { useWsStore } from '../../stores/websocket.js'

const auth = useAuthStore()
const ui = useUiStore()

const isRegister = ref(false)
const loading = ref(false)
const error = ref('')
const username = ref('')
const password = ref('')
const regUsername = ref('')
const regPassword = ref('')
const regSex = ref(0)
const regBirthday = ref('')

function switchMode() {
  isRegister.value = !isRegister.value
  error.value = ''
}

async function handleLogin() {
  if (!username.value || !password.value) { error.value = '请输入用户名和密码'; return }
  loading.value = true; error.value = ''
  try {
    await auth.login(username.value, password.value)
    // Clear stale in-memory state from previous user
    initMain()
    ui.navigateTo('main')
  } catch (e) {
    error.value = e.message || '登录失败'
  } finally {
    loading.value = false
  }
}

async function handleRegister() {
  if (!regUsername.value || !regPassword.value) { error.value = '请填写完整信息'; return }
  if (regUsername.value.length < 2) { error.value = '用户名至少2位'; return }
  if (regPassword.value.length < 6) { error.value = '密码至少6位'; return }
  loading.value = true; error.value = ''
  try {
    await auth.register(regUsername.value, regPassword.value, regSex.value, regBirthday.value)
    ui.showToast('注册成功，请登录', 'success')
    switchMode()
  } catch (e) {
    error.value = e.message || '注册失败'
  } finally {
    loading.value = false
  }
}

function initMain() {
  const contacts = useContactsStore()
  const chat = useChatStore()
  const ws = useWsStore()
  // Clear stale in-memory state from previous user
  chat.clearAll(auth.userId)
  contacts.clearAll(auth.userId)
  // Load fresh data
  contacts.initFriendsFromCache()
  contacts.initGroupsFromCache()
  chat.loadRecentChats()
  // Sync recent chats from server (DB-backed)
  chat.syncRecentChats()
  contacts.fetchFriends()
  contacts.fetchGroups()
  contacts.fetchFriendRequests()
  ws.connect()
}
</script>

<style scoped>
.login-page {
  width: 100%; height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-page);
}
.login-container {
  width: 100%;
  max-width: 360px;
  padding: 32px 24px;
}
.login-logo {
  text-align: center;
  margin-bottom: 32px;
}
.logo-icon { font-size: 56px; margin-bottom: 8px; }
.logo-title { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.form { display: flex; flex-direction: column; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: var(--font-sm); color: var(--text-secondary); }
.form-group input, .form-group select {
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-input);
  color: var(--text-primary);
  outline: none;
  font-size: var(--font-md);
}
.form-group input:focus, .form-group select:focus {
  border-color: var(--primary);
}
.btn {
  padding: 12px; border: none; border-radius: var(--radius-md);
  font-size: var(--font-md); font-weight: 500; transition: opacity 0.2s;
}
.btn:disabled { opacity: 0.6; }
.btn-primary { background: var(--primary); color: #fff; }
.btn-primary:hover { background: var(--primary-hover); }
.form-error { color: var(--danger); font-size: var(--font-sm); text-align: center; }
.switch-link { text-align: center; margin-top: 20px; font-size: var(--font-sm); color: var(--text-hint); }
.switch-link a { color: var(--primary); }
</style>
