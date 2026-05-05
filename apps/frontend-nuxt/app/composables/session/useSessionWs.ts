import type { ConnectionStatus, SessionWsMode, WsEnvelope } from '~/types/session-ws'

interface UseSessionWsOptions {
  mode: SessionWsMode
  onMessage?: (message: WsEnvelope) => void
  onError?: (message: string) => void
  reconnect?: {
    enabled?: boolean
    maxAttempts?: number
    baseDelayMs?: number
  }
}

const DEFAULT_MAX_RECONNECT_ATTEMPTS = 8
const DEFAULT_RECONNECT_DELAY_MS = 800

function resolveWebSocketUrl(mode: SessionWsMode): string {
  const config = useRuntimeConfig()
  const path = mode === 'host' ? config.public.sessionWsHostPath : config.public.sessionWsPlayerPath

  if (!import.meta.client) {
    return path
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}${path}`
}

function parseEnvelope(raw: string): WsEnvelope | null {
  try {
    const parsed = JSON.parse(raw) as { type?: unknown; payload?: unknown }
    if (typeof parsed.type !== 'string' || !parsed.type.trim()) {
      return null
    }

    return {
      type: parsed.type,
      payload: parsed.payload ?? {},
    }
  } catch {
    return null
  }
}

export function useSessionWs(options: UseSessionWsOptions) {
  const socket = shallowRef<WebSocket | null>(null)
  const status = ref<ConnectionStatus>('idle')
  const connectedAt = ref<string | null>(null)
  const lastDisconnectAt = ref<string | null>(null)

  const reconnectEnabled = options.reconnect?.enabled ?? true
  const reconnectMaxAttempts = options.reconnect?.maxAttempts ?? DEFAULT_MAX_RECONNECT_ATTEMPTS
  const reconnectBaseDelayMs = options.reconnect?.baseDelayMs ?? DEFAULT_RECONNECT_DELAY_MS

  let reconnectAttempts = 0
  let manualClose = false
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null

  const clearReconnectTimer = () => {
    if (!reconnectTimer) {
      return
    }

    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }

  const resetReconnectState = () => {
    reconnectAttempts = 0
    clearReconnectTimer()
  }

  const send = (type: string, payload: unknown = {}) => {
    if (!socket.value || socket.value.readyState !== WebSocket.OPEN) {
      throw new Error('Websocket is not connected')
    }

    socket.value.send(JSON.stringify({ type, payload }))
  }

  const scheduleReconnect = () => {
    if (!reconnectEnabled || manualClose) {
      status.value = 'disconnected'
      return
    }

    if (reconnectAttempts >= reconnectMaxAttempts) {
      status.value = 'disconnected'
      options.onError?.('Connection lost')
      return
    }

    reconnectAttempts += 1
    status.value = 'reconnecting'
    const delay = reconnectBaseDelayMs * reconnectAttempts

    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      void connect()
    }, delay)
  }

  const connect = async () => {
    if (!import.meta.client) {
      return
    }

    if (socket.value && (socket.value.readyState === WebSocket.OPEN || socket.value.readyState === WebSocket.CONNECTING)) {
      return
    }

    manualClose = false
    status.value = reconnectAttempts > 0 ? 'reconnecting' : 'connecting'

    const url = resolveWebSocketUrl(options.mode)
    const ws = new WebSocket(url)
    socket.value = ws

    ws.onopen = () => {
      connectedAt.value = new Date().toISOString()
      status.value = 'connected'
      resetReconnectState()
    }

    ws.onmessage = (event) => {
      if (typeof event.data !== 'string') {
        return
      }

      const message = parseEnvelope(event.data)
      if (!message) {
        return
      }

      options.onMessage?.(message)
    }

    ws.onerror = () => {
      options.onError?.('Websocket transport error')
    }

    ws.onclose = () => {
      lastDisconnectAt.value = new Date().toISOString()
      socket.value = null
      scheduleReconnect()
    }
  }

  const disconnect = () => {
    manualClose = true
    clearReconnectTimer()

    if (!socket.value) {
      status.value = 'disconnected'
      return
    }

    if (socket.value.readyState === WebSocket.OPEN || socket.value.readyState === WebSocket.CONNECTING) {
      socket.value.close()
    }

    socket.value = null
    status.value = 'disconnected'
  }

  const isConnected = computed(() => status.value === 'connected')

  return {
    socket,
    status,
    connectedAt,
    lastDisconnectAt,
    isConnected,
    connect,
    disconnect,
    send,
  }
}
