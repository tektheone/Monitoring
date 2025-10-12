// Shared TypeScript types that mirror the backend payloads and responses
// Backend reference: backend/pkg/openapi.json and server handlers

export type ISO8601 = string

// Requests
export interface HeartbeatRequest {
  sent_at: ISO8601
}

export interface StatsRequest {
  sent_at: ISO8601
  // Nanoseconds. When taking ms input from UI, convert: ns = ms * 1_000_000
  upload_time: number
}

// Responses
export interface StatsResponse {
  avg_upload_time: string // e.g. "250ms" or duration format from backend
  uptime: number // percentage, e.g. 100 or 99.58
}

export interface ErrorResponse {
  msg: string
}

export interface HealthResponse {
  status: string
  device_count: number
  timestamp: number
}

export type ApiResult<T> = {
  ok: true
  status: number
  data: T
} | {
  ok: false
  status: number
  error: string
}
