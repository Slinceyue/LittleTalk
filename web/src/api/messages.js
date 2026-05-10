import api from './index.js'

export function getUnreadCount() {
  return api.get('/api/unreadcount')
}

export function getRecentChats() {
  return api.get('/api/message/list')
}

export function getDBChatHistory(friendId, page = 1, pageSize = 50) {
  return api.get('/api/message/db-history', {
    params: { friend_id: friendId, page, page_size: pageSize }
  })
}

export function uploadFile(formData) {
  return api.post('/api/uploadfile', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000,
  })
}

export function searchMessages(query, friendId, page = 1, pageSize = 20) {
  return api.get('/api/message/search', {
    params: { query, friend_id: friendId, page, page_size: pageSize }
  })
}
