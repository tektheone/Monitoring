export type DeviceStatus = 'online' | 'offline'

export interface DeviceSummary {
  id: string
  last_seen?: string
  status: DeviceStatus
}

export interface DeviceStats {
  avg_upload_time: string
  uptime: number
}

export interface HealthStatus {
  status: string
  timestamp: number
  device_count: number
}

export interface ApiError {
  status: number
  message: string
}
