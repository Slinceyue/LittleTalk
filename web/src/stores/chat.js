import { defineStore } from 'pinia'
import * as messagesApi from '../api/messages.js'
import { STORAGE_KEYS, MSG_TYPE } from '../utils/constants.js'
import { getItem, setItem, clearPrefix } from '../utils/storage.js'
import { useAuthStore } from './auth.js'
import { useContactsStore } from './contacts.js'

function convKey(type, id) { return `${type}_${id}` }

export const useChatStore = defineStore('chat', {
  state: () => ({
    activeConversation: null,
    messageMap: {},
    unreadCounts: {},
    typingUsers: {},
    recentChats: [],
    backgrounds: {},     // { 'friend_{id}': 'css background value' }
    bubbleStyles: {},     // { 'friend_{id}': { selfColor, otherColor, borderRadius, bubbleStyle, imageLayout } }
  }),

  getters: {
    activeKey: (s) => s.activeConversation ? convKey(s.activeConversation.type, s.activeConversation.id) : null,
    currentMessages: (s) => {
      if (!s.activeConversation) return []
      return s.messageMap[convKey(s.activeConversation.type, s.activeConversation.id)] || []
    },
    totalUnread: (s) => Object.values(s.unreadCounts).reduce((a, b) => a + b, 0),
    getTypingUsers: (s) => (key) => s.typingUsers[key] || [],
    getBackground: (s) => (friendId) => s.backgrounds[`friend_${friendId}`] || '',
    getBubbleStyle: (s) => (friendId) => s.bubbleStyles[`friend_${friendId}`] || {},
  },

  actions: {
    // Return a userId-scoped storage key
    _scoped(prefix) {
      const auth = useAuthStore()
      const uid = auth.userId || '0'
      return `${prefix}${uid}_`
    },

    openConversation(type, id, name) {
      this.activeConversation = { type, id, name }
      const key = convKey(type, id)
      if (this.unreadCounts[key]) {
        this.unreadCounts[key] = 0
      }
      // Preload chat background and bubble styles for friend conversations
      if (type === 'friend') {
        this.loadBackground(id)
        this.loadBubbleStyle(id)
      }
    },

    closeConversation() {
      this.activeConversation = null
    },

    loadHistory(key) {
      if (this.messageMap[key]) return
      const cached = getItem(this._scoped(STORAGE_KEYS.CHAT_HISTORY) + key)
      this.messageMap[key] = cached || []
      // If local cache is empty, try to sync from DB
      if (this.messageMap[key].length === 0) {
        const parts = key.split('_')
        if (parts[0] === 'talk') {
          const friendId = parseInt(parts[1], 10)
          if (friendId) this.syncHistoryFromServer(friendId)
        }
      }
    },

    appendMessage(key, msg) {
      if (!this.messageMap[key]) this.messageMap[key] = []
      if (msg.msg_id && this.messageMap[key].some(m => m.msg_id === msg.msg_id)) return false
      this.messageMap[key].push(msg)
      if (this.messageMap[key].length > 100) {
        this.messageMap[key] = this.messageMap[key].slice(-100)
      }
      this._saveHistory(key)
      return true
    },

    markMessageFailed(key, msgId) {
      const msgs = this.messageMap[key]
      if (!msgs) return
      const msg = msgs.find(m => m.msg_id === msgId)
      if (msg) msg.send_status = 'failed'
    },

    addUnread(key) {
      if (this.activeConversation && convKey(this.activeConversation.type, this.activeConversation.id) === key) return
      this.unreadCounts[key] = (this.unreadCounts[key] || 0) + 1
    },

    loadRecentChats() {
      const cached = getItem(this._scoped(STORAGE_KEYS.RECENT_CHATS))
      this.recentChats = cached || []
    },

    updateRecentChat(chat) {
      this.recentChats = [
        chat,
        ...this.recentChats.filter(c => c.friend_id !== chat.friend_id)
      ].slice(0, 20)
      setItem(this._scoped(STORAGE_KEYS.RECENT_CHATS), this.recentChats)
    },

    removeRecentChat(friendId) {
      this.recentChats = this.recentChats.filter(c => c.friend_id !== friendId)
      setItem(this._scoped(STORAGE_KEYS.RECENT_CHATS), this.recentChats)
    },

    setTyping(key, userId, isTyping) {
      if (!this.typingUsers[key]) this.typingUsers[key] = []
      if (isTyping) {
        if (!this.typingUsers[key].includes(userId)) this.typingUsers[key].push(userId)
      } else {
        this.typingUsers[key] = this.typingUsers[key].filter(id => id !== userId)
      }
    },

    // Sync recent chats list from server
    async syncRecentChats() {
      try {
        const { data } = await messagesApi.getRecentChats()
        if (data.code === 0 && data.data) {
          // Keep existing group chats (server doesn't return groups)
          const existingGroups = this.recentChats.filter(c => c.type === 'group')
          const serverChats = data.data.map(c => ({
            type: c.type || 'friend',
            friend_id: c.friend_id,
            friend_name: c.friend_name,
            friend_avatar: c.friend_avatar,
            last_message: c.last_message,
            send_time: c.send_time,
            online: c.online,
          }))
          const merged = [...serverChats]
          for (const g of existingGroups) {
            if (!merged.some(m => m.friend_id === g.friend_id && m.type === 'group')) {
              merged.push(g)
            }
          }
          // Sort by send_time descending
          merged.sort((a, b) => (b.send_time || 0) - (a.send_time || 0))
          this.recentChats = merged
          setItem(this._scoped(STORAGE_KEYS.RECENT_CHATS), this.recentChats)
        }
      } catch { /* fall back to localStorage cache */ }
    },

    // Sync message history from DB for a given friend
    async syncHistoryFromServer(friendId) {
      if (!friendId) return
      try {
        const { data } = await messagesApi.getDBChatHistory(friendId)
        if (data.code === 0 && data.data && data.data.length > 0) {
          const key = convKey('talk', friendId)
          const existing = this.messageMap[key] || []
          const existingIds = new Set(existing.map(m => m.msg_id))
          const newMsgs = data.data.filter(m => !existingIds.has(m.msg_id))
          if (newMsgs.length > 0) {
            this.messageMap[key] = [...existing, ...newMsgs]
            this._saveHistory(key)
          }
          // Also warm up user cache for message senders
          const contacts = useContactsStore()
          const senderIds = [...new Set(data.data.map(m => m.from_id))]
          if (senderIds.length > 0) contacts.fetchUsersInfo(senderIds)
        }
      } catch { /* ignore */ }
    },

    // ──── Chat background (per-friend, local storage only) ────

    setBackground(friendId, value) {
      const key = `friend_${friendId}`
      this.backgrounds[key] = value
      setItem(this._scoped(STORAGE_KEYS.CHAT_BG) + friendId, value || '')
    },

    loadBackground(friendId) {
      const cached = getItem(this._scoped(STORAGE_KEYS.CHAT_BG) + friendId)
      if (cached) this.backgrounds[`friend_${friendId}`] = cached
    },

    // ──── Bubble styles (per-friend, local storage only) ────

    setBubbleStyle(friendId, settings) {
      const key = `friend_${friendId}`
      this.bubbleStyles[key] = { ...this.bubbleStyles[key], ...settings }
      setItem(this._scoped(STORAGE_KEYS.CHAT_BUBBLE) + friendId, this.bubbleStyles[key])
    },

    loadBubbleStyle(friendId) {
      const cached = getItem(this._scoped(STORAGE_KEYS.CHAT_BUBBLE) + friendId)
      if (cached) this.bubbleStyles[`friend_${friendId}`] = cached
    },

    // Clear all in-memory and persisted state (for user logout/switch)
    clearAll(userId) {
      this.messageMap = {}
      this.unreadCounts = {}
      this.typingUsers = {}
      this.recentChats = []
      this.backgrounds = {}
      this.bubbleStyles = {}
      this.activeConversation = null
      const uid = userId || '0'
      clearPrefix(STORAGE_KEYS.CHAT_HISTORY + uid + '_')
      clearPrefix(STORAGE_KEYS.RECENT_CHATS + uid + '_')
      clearPrefix(STORAGE_KEYS.CHAT_BG + uid + '_')
      clearPrefix(STORAGE_KEYS.CHAT_BUBBLE + uid + '_')
    },

    _saveHistory(key) {
      setItem(this._scoped(STORAGE_KEYS.CHAT_HISTORY) + key, this.messageMap[key] || [])
    },
  },
})
