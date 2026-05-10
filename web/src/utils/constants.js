// Message types (matching backend enum)
export const MSG_TYPE = { TEXT: 1, FILE: 2, IMAGE: 3 }

// WebSocket message types
export const WS_TYPE = {
  PING: 'ping',
  PONG: 'pong',
  ONLINE_STATUS: 'online_status',
  BATCH_ONLINE: 'batch_online_status',
  GET_ONLINE: 'get_online_status',
  TYPING: 'typing',
  READ_RECEIPT: 'read_receipt',
  RECALL: 'recall',
}

export const WS_MSG_TYPE = {
  TALK: 'talk',
  GROUP_TALK: 'group_talk',
  FRIEND: 'friend',
  ROOM_INVITE: 'room_invite',
  SEND_FAILED: 'send_failed',
}

// localStorage keys
export const STORAGE_KEYS = {
  CHAT_HISTORY: 'lt_chat_',
  RECENT_CHATS: 'lt_recent',
  FRIENDS: 'lt_friends',
  GROUPS: 'lt_groups',
  USER_CACHE: 'lt_users',
  CHAT_BG: 'lt_bg',
  CHAT_BUBBLE: 'lt_bubble',
  THEME: 'lt_theme',
  TOKEN: 'token',
}

// Heartbeat config
export const HEARTBEAT_INTERVAL = 15000
export const HEARTBEAT_TIMEOUT = 60000
export const MAX_RECONNECT = 5

// Room roles
export const ROOM_ROLE = { OWNER: 0, ADMIN: 1, MEMBER: 2 }

// Sex
export const SEX_MAP = { 0: '未知', 1: '男', 2: '女' }

// File limits
export const MAX_AVATAR_SIZE = 2 * 1024 * 1024
export const MAX_FILE_SIZE = 40 * 1024 * 1024
export const MAX_IMAGE_DIM = 1920

// Chat history limit per conversation
export const MAX_CHAT_HISTORY = 100
export const MAX_RECENT_CHATS = 20
