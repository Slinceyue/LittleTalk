import { defineStore } from 'pinia'
import { getItem, setItem } from '../utils/storage.js'
import { STORAGE_KEYS } from '../utils/constants.js'

let toastId = 0

export const useUiStore = defineStore('ui', {
  state: () => ({
    currentPage: 'login',   // 'login' | 'main' | 'chat' | 'addFriend'
    previousPage: null,
    activeTab: 'messages',   // 'messages' | 'friends' | 'groups' | 'profile'
    theme: getItem(STORAGE_KEYS.THEME) || 'light',
    toasts: [],
    activeModal: null,       // 'friendRequest' | 'groupInfo' | 'groupMembers' | 'createGroup' | 'joinGroup' | 'inviteMembers'
    modalData: null,
    searchQuery: '',
  }),

  getters: {
    isDark: (s) => s.theme === 'dark',
  },

  actions: {
    navigateTo(page) {
      this.previousPage = this.currentPage
      this.currentPage = page
    },

    goBack() {
      if (this.previousPage) {
        this.currentPage = this.previousPage
        this.previousPage = null
      }
    },

    switchTab(tab) {
      this.activeTab = tab
    },

    toggleTheme() {
      this.theme = this.theme === 'light' ? 'dark' : 'light'
      document.documentElement.setAttribute('data-theme', this.theme)
      setItem(STORAGE_KEYS.THEME, this.theme)
    },

    applyTheme() {
      document.documentElement.setAttribute('data-theme', this.theme)
    },

    showToast(message, type = 'info', duration = 3000) {
      const id = ++toastId
      this.toasts.push({ id, message, type, duration })
      if (duration > 0) {
        setTimeout(() => this.dismissToast(id), duration)
      }
    },

    dismissToast(id) {
      this.toasts = this.toasts.filter(t => t.id !== id)
    },

    openModal(name, data = null) {
      this.activeModal = name
      this.modalData = data
    },

    closeModal() {
      this.activeModal = null
      this.modalData = null
    },
  },
})
