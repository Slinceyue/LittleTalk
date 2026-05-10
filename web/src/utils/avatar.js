export function getAvatarUrl(userId, avatar) {
  if (avatar) return avatar
  return `/static/avatar/${userId}.jpg`
}

export function getInitial(name) {
  if (!name) return '?'
  return name.charAt(0).toUpperCase()
}
