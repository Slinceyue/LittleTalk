export function getItem(key) {
  try {
    const raw = localStorage.getItem(key)
    return raw ? JSON.parse(raw) : null
  } catch { return null }
}

export function setItem(key, value) {
  try {
    localStorage.setItem(key, JSON.stringify(value))
  } catch { /* quota exceeded, ignore */ }
}

export function removeItem(key) {
  localStorage.removeItem(key)
}

export function clearPrefix(prefix) {
  const keys = []
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i)
    if (key && key.startsWith(prefix)) keys.push(key)
  }
  keys.forEach(k => localStorage.removeItem(k))
}
