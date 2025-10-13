import axios, { AxiosError } from 'axios'
import type { ApiError, DeviceStats, DeviceSummary, HealthStatus } from '@/types/api'

const api = axios.create({
  baseURL: '/', // Vite proxy maps /api/v1 and /health to backend
  timeout: 8000,
  headers: {
    'Content-Type': 'application/json',
    Accept: 'application/json',
  },
})

function toApiError(err: unknown): ApiError {
  if (axios.isAxiosError(err)) {
    const ae = err as AxiosError<{ msg?: string }>
    const status = ae.response?.status ?? 0
    const message = ae.response?.data?.msg || ae.message || 'Request failed'
    return { status, message }
  }
  return { status: 0, message: 'Unknown error' }
}

export async function getHealth(): Promise<HealthStatus> {
  try {
    const { data } = await api.get<HealthStatus>('/health')
    return data
  } catch (e) {
    throw toApiError(e)
  }
}

export async function getDevices(): Promise<DeviceSummary[]> {
  try {
    const { data } = await api.get<DeviceSummary[]>('/api/v1/devices')
    return data
  } catch (e) {
    throw toApiError(e)
  }
}

export async function getDeviceStats(deviceId: string): Promise<DeviceStats | null> {
  try {
    const res = await api.get<DeviceStats>(`/api/v1/devices/${encodeURIComponent(deviceId)}/stats`, {
      validateStatus: () => true,
    })
    if (res.status === 204) return null
    if (res.status >= 200 && res.status < 300) return res.data
    throw { status: res.status, message: 'Failed to fetch device stats' } satisfies ApiError
  } catch (e) {
    throw toApiError(e)
  }
}
