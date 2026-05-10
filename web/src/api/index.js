import axios from 'axios'

const api = axios.create({
  baseURL: '',
  withCredentials: true,
  timeout: 15000,
})

// Request interceptor: attach token
api.interceptors.request.use(config => {
  const token = getCookie('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor: auto-logout on 401
api.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401) {
      deleteCookie('token')
      window.location.reload()
    }
    return Promise.reject(err)
  }
)

// ----- Cookie helpers -----
export function getCookie(name) {
  const eq = name + '='
  const parts = document.cookie.split(';')
  for (let p of parts) {
    while (p.charAt(0) === ' ') p = p.substring(1)
    if (p.indexOf(eq) === 0) return decodeURIComponent(p.substring(eq.length))
  }
  return null
}

export function setCookie(name, value, days = 7) {
  const d = new Date()
  d.setTime(d.getTime() + days * 86400000)
  document.cookie = `${name}=${encodeURIComponent(value)};expires=${d.toUTCString()};path=/`
}

export function deleteCookie(name) {
  document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/`
}

export function getToken() {
  return getCookie('token')
}

export default api
