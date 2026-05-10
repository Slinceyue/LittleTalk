import { defineStore } from 'pinia'
import * as authApi from '../api/auth.js'
import { setCookie, deleteCookie, getToken } from '../api/index.js'
import { STORAGE_KEYS } from '../utils/constants.js'
import { getItem, setItem, removeItem, clearPrefix } from '../utils/storage.js'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    token: getToken() || null,
  }),

  getters: {
    isLoggedIn: (s) => !!s.token && !!s.user,
    userId: (s) => s.user?.id || 0,
    username: (s) => s.user?.username || '',
    avatar: (s) => s.user?.avatar || '',
  },

  actions: {
    async login(username, password) {
      const { data } = await authApi.login(username, password)
      if (data.code !== 0) throw new Error(data.message || '登录失败')
      this.token = data.data.token
      setCookie('token', data.data.token, 7)
      await this.fetchSelfInfo()
    },

    async register(username, password, sex, birthday) {
      const { data } = await authApi.register(username, password, sex, birthday)
      if (data.code !== 0) throw new Error(data.message || '注册失败')
    },

    async fetchSelfInfo() {
      const { data } = await authApi.getSelfInfo()
      if (data.code !== 0) throw new Error(data.message || '获取用户信息失败')
      this.user = data.data
      setItem(STORAGE_KEYS.USER_CACHE + this.user.id, this.user)
    },

    async updateProfile(profile) {
      const { data } = await authApi.updateProfile(profile)
      if (data.code !== 0) throw new Error(data.message || '保存失败')
      if (data.data) {
        this.user = { ...this.user, ...data.data }
      } else {
        await this.fetchSelfInfo()
      }
    },

    async uploadAvatar(file) {
      const formData = new FormData()
      formData.append('avatar', file)
      const { data } = await authApi.uploadAvatar(formData)
      if (data.code !== 0) throw new Error(data.message || '上传失败')
      await this.fetchSelfInfo()
    },

    async autoLogin() {
      const token = getToken()
      if (!token) return false
      this.token = token
      try {
        await this.fetchSelfInfo()
        return true
      } catch {
        this.token = null
        deleteCookie('token')
        return false
      }
    },

    async logout() {
      const uid = this.user?.id
      try { await authApi.offline() } catch { /* ignore */ }
      this.token = null
      this.user = null
      deleteCookie('token')
      // Clear all user-scoped cache
      if (uid) {
        clearPrefix(STORAGE_KEYS.CHAT_HISTORY + uid + '_')
        clearPrefix(STORAGE_KEYS.RECENT_CHATS + uid + '_')
        clearPrefix(STORAGE_KEYS.FRIENDS + uid + '_')
        clearPrefix(STORAGE_KEYS.GROUPS + uid + '_')
      }
    },
  },
})
