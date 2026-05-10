import api from './index.js'

export function login(username, password) {
  return api.post('/login', { username, password })
}

export function register(username, password, sex, birthday) {
  return api.post('/creatuser', { username, password, sex, birthday })
}

export function getSelfInfo() {
  return api.get('/api/selfuserinfo')
}

export function getOtherInfo(userId) {
  return api.get('/api/otheruserinfo', { params: { user_id: userId } })
}

export function getUsersInfo(ids) {
  return api.get('/api/usersinfo', { params: { ids: ids.join(',') } })
}

export function updateProfile(data) {
  return api.post('/api/updateuserinfo', data)
}

export function uploadAvatar(formData) {
  return api.post('/api/uploadavatar', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function offline() {
  return api.post('/api/offline')
}
