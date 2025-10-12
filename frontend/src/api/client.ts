import { HeartbeatRequest, StatsRequest, StatsResponse, ErrorResponse, HealthResponse, ApiResult } from './types'

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'
const DEFAULT_TIMEOUT_MS = 8000

function withTimeout(signal: AbortSignal | undefined, ms: number) {
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), ms)

  // If an external signal is provided, tie them together
  if (signal) {
    if (signal.aborted) controller.abort()
    signal.addEventListener('abort', () => controller.abort(), { once: true })
  }

  return { signal: controller.signal, cancel: () => clearTimeout(timer) }
}

async function apiFetch<T>(path: string, init: RequestInit & { timeoutMs?: number } = {}): Promise<ApiResult<T>> {
  const { timeoutMs = DEFAULT_TIMEOUT_MS, ...rest } = init
  const { signal, cancel } = withTimeout(init.signal as AbortSignal | undefined, timeoutMs)

  try {
    const res = await fetch(`${API_BASE}${path}`, { ...rest, signal })

    // 204 No Content
    if (res.status === 204) {
      cancel()
      return { ok: true, status: res.status, data: undefined as unknown as T }
    }

    const ct = res.headers.get('content-type') || ''
    const isJson = ct.includes('application/json')
    const body = isJson ? await res.json() : await res.text()

    cancel()

    if (!res.ok) {
      const errMsg = isJson && (body as Partial<ErrorResponse>)?.msg ? (body as ErrorResponse).msg : (typeof body === 'string' ? body : 'Request failed')
      return { ok: false, status: res.status, error: errMsg }
    }

    return { ok: true, status: res.status, data: body as T }
  } catch (e: any) {
    cancel()
    if (e?.name === 'AbortError') {
      return { ok: false, status: 0, error: 'Request timed out' }
    }
    return { ok: false, status: 0, error: e?.message ?? 'Network error' }
  }
}

// Non-versioned health endpoint goes through Vite proxy directly
async function fetchHealth(init: RequestInit & { timeoutMs?: number } = {}): Promise<ApiResult<HealthResponse>> {
  const { timeoutMs = DEFAULT_TIMEOUT_MS, ...rest } = init
  const { signal, cancel } = withTimeout(init.signal as AbortSignal | undefined, timeoutMs)
  try {
    const res = await fetch('/health', { ...rest, signal })
    const body = await res.json()
    cancel()
    if (!res.ok) return { ok: false, status: res.status, error: (body as any)?.msg ?? 'Health check failed' }
    return { ok: true, status: res.status, data: body as HealthResponse }
  } catch (e: any) {
    cancel()
    if (e?.name === 'AbortError') return { ok: false, status: 0, error: 'Request timed out' }
    return { ok: false, status: 0, error: e?.message ?? 'Network error' }
  }
}

export const api = {
  // POST /devices/{device_id}/heartbeat
  async postHeartbeat(deviceId: string, payload: HeartbeatRequest, init?: RequestInit & { timeoutMs?: number }): Promise<ApiResult<void>> {
    return apiFetch<void>(`/devices/${encodeURIComponent(deviceId)}/heartbeat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
      ...init,
    })
  },

  // POST /devices/{device_id}/stats
  async postStats(deviceId: string, payload: StatsRequest, init?: RequestInit & { timeoutMs?: number }): Promise<ApiResult<void>> {
    return apiFetch<void>(`/devices/${encodeURIComponent(deviceId)}/stats`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
      ...init,
    })
  },

  // GET /devices/{device_id}/stats
  async getStats(deviceId: string, init?: RequestInit & { timeoutMs?: number }): Promise<ApiResult<StatsResponse | null>> {
    const res = await apiFetch<StatsResponse>(`/devices/${encodeURIComponent(deviceId)}/stats`, {
      method: 'GET',
      ...init,
    })

    // Interpret 204 as null data
    if (res.ok && res.status === 204) {
      return { ok: true, status: 204, data: null }
    }

    return res as ApiResult<StatsResponse | null>
  },

  // GET /health (non-versioned)
  health: fetchHealth,
}

export type { HeartbeatRequest, StatsRequest, StatsResponse, ErrorResponse, HealthResponse }
