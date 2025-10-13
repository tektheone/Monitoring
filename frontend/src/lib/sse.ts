type Handler<T = any> = (data: T) => void

type SSEStatus = 'disconnected' | 'connecting' | 'connected'

type EventMap = {
  'health:update': any
  'devices:update': any
  'device:update': any
  'status': SSEStatus
}

class SSEClient {
  private es: EventSource | null = null
  private url: string
  private status: SSEStatus = 'disconnected'
  private listeners: { [K in keyof EventMap]?: Set<Handler<EventMap[K]>> } = {}
  private retryMs = 1000
  private maxRetryMs = 15000
  private reconnectTimer: number | null = null
  private connectTimer: number | null = null
  private connectTimeoutMs = 2000
  private lastMessageAt: number = 0
  private idleTimer: number | null = null
  private idleTimeoutMs = 3000

  constructor(url: string) {
    this.url = url
  }

  getStatus(): SSEStatus { return this.status }

  private setStatus(s: SSEStatus) {
    this.status = s
    this.emit('status', s)
  }

  on<K extends keyof EventMap>(type: K, handler: Handler<EventMap[K]>) {
    if (!this.listeners[type]) this.listeners[type] = new Set()
    this.listeners[type]!.add(handler as any)
    return () => this.listeners[type]!.delete(handler as any)
  }

  private emit<K extends keyof EventMap>(type: K, data: EventMap[K]) {
    this.listeners[type]?.forEach((h) => {
      try { (h as any)(data) } catch {}
    })
  }

  start() {
    if (this.es) return
    this.setStatus('connecting')
    const es = new EventSource(this.url)
    this.es = es

    // Watchdog: if we don't get onopen in time, treat as disconnected
    if (this.connectTimer !== null) window.clearTimeout(this.connectTimer)
    this.connectTimer = window.setTimeout(() => {
      this.connectTimer = null
      if (this.status === 'connecting') {
        this.setStatus('disconnected')
        // Force close and trigger reconnect
        try { this.es?.close() } catch {}
        this.es = null
        this.scheduleReconnect()
      }
    }, this.connectTimeoutMs)

    es.onopen = () => {
      this.retryMs = 1000
      this.setStatus('connected')
      if (this.connectTimer !== null) {
        window.clearTimeout(this.connectTimer)
        this.connectTimer = null
      }
      this.lastMessageAt = Date.now()
      this.startIdleWatchdog()
    }

    es.onerror = () => {
      this.setStatus('disconnected')
      // Close and nullify so our reconnect logic can run
      try { this.es?.close() } catch {}
      this.es = null
      if (this.connectTimer !== null) {
        window.clearTimeout(this.connectTimer)
        this.connectTimer = null
      }
      this.stopIdleWatchdog()
      this.scheduleReconnect()
    }

    es.onmessage = (evt) => {
      // Expect JSON lines: { type: string, data: any }
      try {
        const msg = JSON.parse(evt.data)
        if (msg && typeof msg.type === 'string') {
          this.emit(msg.type as any, msg.data)
        }
      } catch {
        // ignore
      }
      this.lastMessageAt = Date.now()
    }
  }

  stop() {
    if (this.es) {
      this.es.close()
      this.es = null
    }
    if (this.reconnectTimer !== null) {
      window.clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.connectTimer !== null) {
      window.clearTimeout(this.connectTimer)
      this.connectTimer = null
    }
    this.stopIdleWatchdog()
    this.setStatus('disconnected')
  }

  private scheduleReconnect() {
    if (this.reconnectTimer !== null || this.es) return
    const delay = this.retryMs
    this.retryMs = Math.min(this.retryMs * 2, this.maxRetryMs)
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null
      this.start()
    }, delay)
  }

  private startIdleWatchdog() {
    this.stopIdleWatchdog()
    this.idleTimer = window.setInterval(() => {
      if (this.status === 'connected') {
        const idle = Date.now() - this.lastMessageAt
        if (idle > this.idleTimeoutMs) {
          // Consider connection stale; reset and reconnect
          this.setStatus('disconnected')
          try { this.es?.close() } catch {}
          this.es = null
          this.scheduleReconnect()
        }
      }
    }, 2000)
  }

  private stopIdleWatchdog() {
    if (this.idleTimer !== null) {
      window.clearInterval(this.idleTimer)
      this.idleTimer = null
    }
  }
}

export const sse = new SSEClient('/events')
