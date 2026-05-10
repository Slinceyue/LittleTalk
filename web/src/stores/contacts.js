import { defineStore } from 'pinia'
import * as contactsApi from '../api/contacts.js'
import * as authApi from '../api/auth.js'
import { STORAGE_KEYS } from '../utils/constants.js'
import { getItem, setItem, clearPrefix } from '../utils/storage.js'
import { getAvatarUrl } from '../utils/avatar.js'
import { useAuthStore } from './auth.js'

export const useContactsStore = defineStore('contacts', {
  state: () => ({
    friends: [],
    friendRequests: [],
    groups: [],
    userCache: {},
    onlineStatus: {},
  }),

  getters: {
    getFriendById: (s) => (id) => s.friends.find(f => f.id === id),
    getGroupById: (s) => (id) => s.groups.find(g => g.id === id),
    getUserName: (s) => (id) => {
      if (id === 0 || !id) return '系统'
      const u = s.userCache[id]
      return u ? u.username : `用户${id}`
    },
    getUserAvatar: (s) => (id) => {
      if (id === 0 || !id) return ''
      const u = s.userCache[id]
      return getAvatarUrl(id, u?.avatar)
    },
    isOnline: (s) => (id) => !!s.onlineStatus[id],
    onlineFriends: (s) => s.friends.filter(f => s.onlineStatus[f.id]),
  },

  actions: {
    _scoped(prefix) {
      const auth = useAuthStore()
      return `${prefix}${auth.userId || '0'}_`
    },

    async fetchFriends() {
      try {
        const { data } = await contactsApi.getFriendList()
        if (data.code === 0) {
          this.friends = (data.data || []).map(f => ({ ...f, id: f.id || f.user_id }))
          this.friends.forEach(f => {
            this.userCache[f.id] = { ...this.userCache[f.id], ...f }
          })
          setItem(this._scoped(STORAGE_KEYS.FRIENDS), this.friends)
        }
      } catch { /* fall back to cache */ }
    },

    initFriendsFromCache() {
      const cached = getItem(this._scoped(STORAGE_KEYS.FRIENDS))
      if (cached) {
        this.friends = cached
        cached.forEach(f => { this.userCache[f.id] = { ...this.userCache[f.id], ...f } })
      }
    },

    async fetchFriendRequests() {
      try {
        const { data } = await contactsApi.getFriendRequestList()
        if (data.code === 0) {
          this.friendRequests = data.data || []
          if (this.friendRequests.length > 0) {
            await this.fetchUsersInfo(this.friendRequests)
          }
        } else {
          console.warn('[Contacts] fetchFriendRequests API error:', data.message)
        }
      } catch (e) {
        console.warn('[Contacts] fetchFriendRequests failed:', e)
      }
    },

    async sendFriendRequest(friendId) {
      const { data } = await contactsApi.sendFriendRequest(friendId)
      if (data.code !== 0) throw new Error(data.message || '发送失败')
    },

    async acceptFriendRequest(fromId) {
      const { data } = await contactsApi.acceptFriendRequest(fromId)
      if (data.code !== 0) throw new Error(data.message || '操作失败')
      this.friendRequests = this.friendRequests.filter(id => id !== fromId)
      await this.fetchFriends()
    },

    async rejectFriendRequest(fromId) {
      const { data } = await contactsApi.rejectFriendRequest(fromId)
      if (data.code !== 0) throw new Error(data.message || '操作失败')
      this.friendRequests = this.friendRequests.filter(id => id !== fromId)
    },

    async deleteFriend(friendId) {
      const { data } = await contactsApi.deleteFriend(friendId)
      if (data.code !== 0) throw new Error(data.message || '删除失败')
      this.friends = this.friends.filter(f => f.id !== friendId)
      setItem(this._scoped(STORAGE_KEYS.FRIENDS), this.friends)
    },

    async fetchGroups() {
      try {
        const { data } = await contactsApi.getGroupList()
        if (data.code === 0) {
          this.groups = data.data || []
          setItem(this._scoped(STORAGE_KEYS.GROUPS), this.groups)
        }
      } catch { /* fall back to cache */ }
    },

    initGroupsFromCache() {
      const cached = getItem(this._scoped(STORAGE_KEYS.GROUPS))
      if (cached) this.groups = cached
    },

    async createGroup(name) {
      const { data } = await contactsApi.createGroup(name)
      if (data.code !== 0) throw new Error(data.message || '创建失败')
      await this.fetchGroups()
      return data.data
    },

    async joinGroup(roomId) {
      const { data } = await contactsApi.joinGroup(roomId)
      if (data.code !== 0) throw new Error(data.message || '加入失败')
      await this.fetchGroups()
    },

    async fetchUsersInfo(userIds) {
      const missing = userIds.filter(id => !this.userCache[id])
      if (missing.length === 0) return
      try {
        const { data } = await authApi.getUsersInfo(missing)
        if (data.code === 0 && data.data) {
          data.data.forEach(u => { this.userCache[u.id] = { ...this.userCache[u.id], ...u } })
        }
      } catch { /* ignore */ }
    },

    cacheUser(user) {
      if (user && user.id) {
        this.userCache[user.id] = { ...this.userCache[user.id], ...user }
      }
    },

    setOnlineStatus(userId, online) {
      this.onlineStatus[userId] = online
    },

    batchSetOnlineStatus(statuses) {
      statuses.forEach(s => { this.onlineStatus[s.user_id] = s.online })
    },

    clearAll(userId) {
      this.friends = []
      this.friendRequests = []
      this.groups = []
      this.userCache = {}
      this.onlineStatus = {}
      const uid = userId || '0'
      clearPrefix(STORAGE_KEYS.FRIENDS + uid + '_')
      clearPrefix(STORAGE_KEYS.GROUPS + uid + '_')
    },
  },
})
