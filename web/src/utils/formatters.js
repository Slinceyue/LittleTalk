export function formatTime(timestamp) {
  if (!timestamp) return ''
  const d = new Date(timestamp)
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const yesterday = new Date(today.getTime() - 86400000)
  const msgDate = new Date(d.getFullYear(), d.getMonth(), d.getDate())

  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  const time = `${hh}:${mm}`

  if (msgDate.getTime() === today.getTime()) return time
  if (msgDate.getTime() === yesterday.getTime()) return `昨天 ${time}`
  if (d.getFullYear() === now.getFullYear()) return `${d.getMonth() + 1}/${d.getDate()} ${time}`
  return `${d.getFullYear()}/${d.getMonth() + 1}/${d.getDate()} ${time}`
}

export function formatFileSize(bytes) {
  if (!bytes) return '0 B'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

export function escapeHtml(str) {
  if (!str) return ''
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;').replace(/'/g, '&#39;')
}

export function buildDownloadUrl(fileUrl, fileType, fromId) {
  if (!fileUrl) return ''
  if (fileUrl.startsWith('http') || fileUrl.startsWith('/')) return fileUrl
  const type = fileType === 'image' || fileType === 3 ? 'image' : 'file'
  return `/api/downloadfile?file_name=${encodeURIComponent(fileUrl)}&file_type=${type}&from_id=${fromId}`
}
