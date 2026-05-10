import api from './index.js'

export function getFriendList() {
  return api.get('/api/friendlist')
}

export function getFriendRequestList() {
  return api.get('/api/friendreqlist')
}

export function sendFriendRequest(friendId) {
  return api.post('/api/newfriendreq', { friend_id: friendId })
}

export function acceptFriendRequest(fromId) {
  return api.post('/api/okwithfriendreq', { from_id: fromId })
}

export function rejectFriendRequest(fromId) {
  return api.post('/api/rejectfriendreq', { from_id: fromId })
}

export function deleteFriend(friendId) {
  return api.post('/api/deletefriend', { friend_id: friendId })
}

// Groups
export function getGroupList() {
  return api.get('/api/rooms')
}

export function getGroupInfo(roomId) {
  return api.get(`/api/room/${roomId}`)
}

export function getGroupMembers(roomId) {
  return api.get(`/api/room/${roomId}/members`)
}

export function createGroup(name) {
  return api.post('/api/room', { name })
}

export function joinGroup(roomId) {
  return api.post('/api/room/join', { room_id: roomId })
}

export function quitGroup(roomId) {
  return api.post('/api/room/quit', { room_id: roomId })
}

export function dismissGroup(roomId) {
  return api.post('/api/room/dismiss', { room_id: roomId })
}

export function setAdmin(roomId, targetUserId, isAdmin) {
  return api.post('/api/room/admin', { room_id: roomId, target_user_id: targetUserId, is_admin: isAdmin })
}

export function kickMember(roomId, targetUserId) {
  return api.post('/api/room/kick', { room_id: roomId, target_user_id: targetUserId })
}

export function inviteMembers(roomId, targetUserIds) {
  return api.post('/api/room/invite', { room_id: roomId, target_user_ids: targetUserIds })
}

export function searchRooms(keyword) {
  return api.get('/api/rooms/search', { params: { keyword } })
}
