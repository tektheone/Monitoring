import { describe, it, expect, beforeEach, vi } from 'vitest'

// Hoisted mocks for axios instance methods to satisfy Vitest hoisting
const { post, get } = vi.hoisted(() => ({
  post: vi.fn(),
  get: vi.fn(),
}))

vi.mock('axios', () => {
  const instance = { post, get }
  return {
    default: {
      create: vi.fn(() => instance),
      isAxiosError: (err: any) => !!err?.isAxiosError,
    },
  }
})

import { sendHeartbeat, getHealth, getDevices, getDeviceStats } from '@/api/client'

describe('api/client', () => {
  beforeEach(() => {
    post.mockReset()
    get.mockReset()
  })

  it('sendHeartbeat posts with correct payload', async () => {
    const when = new Date('2020-01-01T00:00:00.000Z')
    post.mockResolvedValue({})
    await sendHeartbeat('dev-1', when)
    expect(post).toHaveBeenCalledWith(
      '/api/v1/devices/dev-1/heartbeat',
      { sent_at: '2020-01-01T00:00:00.000Z' },
    )
  })

  it('sendHeartbeat maps network error to friendly message', async () => {
    post.mockRejectedValue({ isAxiosError: true, response: undefined })
    await expect(sendHeartbeat('dev-1')).rejects.toThrow('Disconnected: cannot reach backend')
  })

  it('getHealth returns data on success', async () => {
    const data = { status: 'ok' } as any
    get.mockResolvedValue({ data })
    await expect(getHealth()).resolves.toEqual(data)
  })

  it('getHealth maps 500 to friendly message', async () => {
    get.mockRejectedValue({ isAxiosError: true, response: { status: 500 } })
    await expect(getHealth()).rejects.toThrow('Server unavailable (5xx)')
  })

  it('getDevices returns list on success', async () => {
    const devices = [{ id: 'a' }, { id: 'b' }] as any
    get.mockResolvedValue({ data: devices })
    await expect(getDevices()).resolves.toEqual(devices)
  })

  it('getDevices maps 404 to Not found', async () => {
    get.mockRejectedValue({ isAxiosError: true, response: { status: 404 } })
    await expect(getDevices()).rejects.toThrow('Not found')
  })

  it('getDevices maps 401/403 to Unauthorized', async () => {
    get.mockRejectedValue({ isAxiosError: true, response: { status: 401 } })
    await expect(getDevices()).rejects.toThrow('Unauthorized')
    get.mockRejectedValue({ isAxiosError: true, response: { status: 403 } })
    await expect(getDevices()).rejects.toThrow('Unauthorized')
  })

  it('getDevices uses server-provided message when available', async () => {
    get.mockRejectedValue({ isAxiosError: true, response: { status: 400, data: { msg: 'Bad things' } } })
    await expect(getDevices()).rejects.toThrow('Bad things')
  })

  it('getDevices falls back to Request failed when no server message', async () => {
    get.mockRejectedValue({ isAxiosError: true, response: { status: 400 } })
    await expect(getDevices()).rejects.toThrow('Request failed')
  })

  it('getDeviceStats returns null on 204', async () => {
    get.mockResolvedValue({ status: 204 })
    await expect(getDeviceStats('dev-1')).resolves.toBeNull()
  })

  it('getDeviceStats returns data on 200', async () => {
    const stats = { cpu: 0.5 } as any
    get.mockResolvedValue({ status: 200, data: stats })
    await expect(getDeviceStats('dev-1')).resolves.toEqual(stats)
  })

  it('getDeviceStats throws friendly message for 500', async () => {
    // validateStatus always true means axios resolves; we simulate that
    get.mockResolvedValue({ status: 500 })
    await expect(getDeviceStats('dev-1')).rejects.toThrow('Server unavailable (5xx)')
  })

  it('getDeviceStats throws client error message for 400', async () => {
    get.mockResolvedValue({ status: 400 })
    await expect(getDeviceStats('dev-1')).rejects.toThrow('Failed to fetch device stats')
  })

  it('getDeviceStats maps network error via friendlyMessage', async () => {
    get.mockRejectedValue({ isAxiosError: true, response: undefined })
    await expect(getDeviceStats('dev-1')).rejects.toThrow('Disconnected: cannot reach backend')
  })

  it('sendHeartbeat maps non-Axios error to Unknown error', async () => {
    post.mockRejectedValue(new Error('boom'))
    await expect(sendHeartbeat('dev-1')).rejects.toThrow('Unknown error')
  })
})
