import axios, { AxiosError } from 'axios'
import type { DeviceStats, DeviceSummary, HealthStatus } from '@/types/api'

const api = axios.create({
  baseURL: '/', // Vite proxy maps /api/v1 and /health to backend
  timeout: 8000,
  headers: {
    'Content-Type': 'application/json',
    Accept: 'application/json',
  },
})

function friendlyMessage(err: unknown): string {
  if (axios.isAxiosError(err)) {
    const ae = err as AxiosError<{ msg?: string }>
    const status = ae.response?.status
    const serverMsg = ae.response?.data?.msg
    if (status === undefined) {
      // No response from server (network error, server down, CORS, etc.)
      return 'Disconnected: cannot reach backend'
    }
    if (status >= 500) return 'Server unavailable (5xx)'
    if (status === 404) return 'Not found'
    if (status === 401 || status === 403) return 'Unauthorized'
    if (serverMsg) return serverMsg
    return 'Request failed'
  }
  return 'Unknown error'
}

export async function sendHeartbeat(deviceId: string, when: Date = new Date()): Promise<void> {
  try {
    const url = `/api/v1/devices/${encodeURIComponent(deviceId)}/heartbeat`
    await api.post(url, { sent_at: when.toISOString() })
  } catch (e) {
    throw new Error(friendlyMessage(e))
  }
}

export async function getHealth(): Promise<HealthStatus> {
  try {
    const { data } = await api.get<HealthStatus>('/health')
    return data
  } catch (e) {
    throw new Error(friendlyMessage(e))
  }
}

export async function getDevices(): Promise<DeviceSummary[]> {
  try {
    const { data } = await api.get<DeviceSummary[]>('/api/v1/devices')
    return data
  } catch (e) {
    throw new Error(friendlyMessage(e))
  }
}

export async function getDeviceStats(deviceId: string): Promise<DeviceStats | null> {
  try {
    const res = await api.get<DeviceStats>(`/api/v1/devices/${encodeURIComponent(deviceId)}/stats`, {
      validateStatus: () => true,
    })
    if (res.status === 204) return null
    if (res.status >= 200 && res.status < 300) return res.data
    throw new Error(res.status >= 500 ? 'Server unavailable (5xx)' : 'Failed to fetch device stats')
  } catch (e) {
    throw new Error(friendlyMessage(e))
  }
}
