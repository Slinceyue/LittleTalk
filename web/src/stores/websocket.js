import { defineStore } from 'pinia'
import { getToken } from '../api/index.js'
import { WS_TYPE, WS_MSG_TYPE, HEARTBEAT_INTERVAL, HEARTBEAT_TIMEOUT, MAX_RECONNECT } from '../utils/constants.js'

export const useWsStore = defineStore('websocket', {
  state: () => ({
    connected: false,
    connecting: false,
    reconnectAttempts: 0,
    lastPongTime: 0,
    lastPingTime: 0,
    heartbeatTimer: null,
    reconnectTimer: null,
    socket: null,
    messageHandlers: {},
    offlineQueue: [],
  }),

  actions: {
    connect() {
      const token = getToken()
      if (!token || this.connected || this.connecting) return

      this.connecting = true
      const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
      const host = location.host
      const url = `${protocol}://${host}/api/ws?token=${encodeURIComponent(token)}`

      try {
        this.socket = new WebSocket(url)
      } catch {
        this.connecting = false
        this._scheduleReconnect()
        return
      }

      this.socket.onopen = () => {
        this.connected = true
        this.connecting = false
        this.reconnectAttempts = 0
        this.lastPongTime = Date.now()
        this._startHeartbeat()
        // Request online statuses
        this.send({ type: WS_TYPE.GET_ONLINE })
        // Flush offline queue
        this.offlineQueue.forEach(d => this.send(d))
        this.offlineQueue = []
      }

      this.socket.onmessage = (e) => {
        try {
          const msg = JSON.parse(e.data)
          this._routeMessage(msg)
        } catch { /* ignore parse errors */ }
      }

      this.socket.onerror = () => { /* handled by onclose */ }

      this.socket.onclose = () => {
        this.connected = false
        this.connecting = false
        this.socket = null
        this._stopHeartbeat()
        this._scheduleReconnect()
      }
    },

    disconnect() {
      this._stopHeartbeat()
      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer)
        this.reconnectTimer = null
      }
      if (this.socket) {
        this.socket.onclose = null  // prevent reconnect
        this.socket.close()
        this.socket = null
      }
      this.connected = false
      this.connecting = false
    },

    send(data) {
      if (this.connected && this.socket) {
        this.socket.send(JSON.stringify(data))
      } else {
        this.offlineQueue.push(data)
        if (!this.connected && !this.connecting) this.connect()
      }
    },

    on(type, handler) {
      if (!this.messageHandlers[type]) this.messageHandlers[type] = []
      this.messageHandlers[type].push(handler)
    },

    off(type, handler) {
      if (!this.messageHandlers[type]) return
      this.messageHandlers[type] = this.messageHandlers[type].filter(h => h !== handler)
    },

    // ----- Private -----
    _routeMessage(msg) {
      const type = msg.type || msg.msg_type
      if (type === WS_TYPE.PONG) {
        this.lastPongTime = Date.now()
      }
      // Call registered handlers
      const handlers = this.messageHandlers[type]
      if (handlers) handlers.forEach(h => h(msg))
      // Also call wildcard handlers
      const all = this.messageHandlers['*']
      if (all) all.forEach(h => h(msg))
    },

    _startHeartbeat() {
      this._stopHeartbeat()
      this.lastPingTime = Date.now()
      this.lastPongTime = Date.now()
      this.heartbeatTimer = setInterval(() => {
        if (!this.connected) return
        if (Date.now() - this.lastPongTime > HEARTBEAT_TIMEOUT) {
          this.socket?.close()
          return
        }
        this.lastPingTime = Date.now()
        this.send({ type: WS_TYPE.PING })
      }, HEARTBEAT_INTERVAL)
    },

    _stopHeartbeat() {
      if (this.heartbeatTimer) {
        clearInterval(this.heartbeatTimer)
        this.heartbeatTimer = null
      }
    },

    _scheduleReconnect() {
      if (this.reconnectAttempts >= MAX_RECONNECT) return
      if (!getToken()) return
      this.reconnectAttempts++
      const delay = Math.min(this.reconnectAttempts * 2000, 10000)
      this.reconnectTimer = setTimeout(() => this.connect(), delay)
    },
  },
})
