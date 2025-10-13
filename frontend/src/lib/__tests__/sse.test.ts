import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'

// Global mock for EventSource
class MockEventSource {
  static instances: MockEventSource[] = []
  onopen: ((this: EventSource, ev: Event) => any) | null = null
  onerror: ((this: EventSource, ev: Event) => any) | null = null
  onmessage: ((this: EventSource, ev: MessageEvent) => any) | null = null
  url: string
  closed = false
  constructor(url: string) {
    this.url = url
    MockEventSource.instances.push(this)
  }
  close() {
    this.closed = true
  }
}

// Attach mock to global
;(globalThis as any).EventSource = MockEventSource

// Import the module under test after setting the mock
import { sse } from '../sse'

describe('sse client', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    // ensure clean state
    try { sse.stop() } catch {}
    MockEventSource.instances = []
  })
  afterEach(() => {
    vi.useRealTimers()
    try { sse.stop() } catch {}
  })

  it('emits status connecting then connected on open', () => {
    const statuses: string[] = []
    sse.on('status', (s) => statuses.push(s as any))

    sse.start()
    expect(statuses[0]).toBe('connecting')
    expect(MockEventSource.instances.length).toBe(1)
    expect(MockEventSource.instances[0].url).toBe('/events')

    // simulate open
    MockEventSource.instances[0].onopen?.call(MockEventSource.instances[0] as any, new Event('open'))
    expect(statuses).toContain('connected')
  })

  it('falls back to disconnected if connect timeout elapses and then reconnects', async () => {
    const statuses: string[] = []
    sse.on('status', (s) => statuses.push(s as any))

    sse.start()
    // connect timeout default is 2000ms
    await vi.advanceTimersByTimeAsync(2000)
    expect(statuses).toContain('disconnected')
    // scheduleReconnect uses retryMs=1000 initially
    await vi.advanceTimersByTimeAsync(1000)
    expect(MockEventSource.instances.length).toBeGreaterThanOrEqual(2)
    expect(statuses[statuses.length - 1]).toBe('connecting')
  })

  it('disconnects on error and schedules reconnect', async () => {
    const statuses: string[] = []
    sse.on('status', (s) => statuses.push(s as any))

    sse.start()
    MockEventSource.instances[0].onerror?.call(MockEventSource.instances[0] as any, new Event('error'))
    expect(statuses).toContain('disconnected')
    // Retry is scheduled at 1000ms; advance a bit more and flush
    await vi.advanceTimersByTimeAsync(1100)
    await vi.runOnlyPendingTimersAsync()
    expect(MockEventSource.instances.length).toBeGreaterThanOrEqual(2)
  })

  it('dispatches typed messages to listeners', () => {
    const payloads: any[] = []
    sse.on('devices:update', (data) => payloads.push(data))

    sse.start()
    const inst = MockEventSource.instances[0]
    inst.onmessage?.call(inst as any, new MessageEvent('message', { data: JSON.stringify({ type: 'devices:update', data: { id: 'x' } }) }))
    expect(payloads).toEqual([{ id: 'x' }])

    // invalid JSON should not throw
    inst.onmessage?.call(inst as any, new MessageEvent('message', { data: 'not-json' }))
  })

  it('idle watchdog disconnects when idle > threshold and reconnects', async () => {
    const statuses: string[] = []
    sse.on('status', (s) => statuses.push(s as any))

    sse.start()
    const inst = MockEventSource.instances[0]
    // Open connection to start idle watchdog
    inst.onopen?.call(inst as any, new Event('open'))
    // The watchdog checks every 2000ms; idleTimeoutMs is 3000ms
    await vi.advanceTimersByTimeAsync(4000)
    expect(statuses).toContain('disconnected')
    await vi.advanceTimersByTimeAsync(1000)
    expect(MockEventSource.instances.length).toBeGreaterThanOrEqual(2)
  })

  it('stop() closes connection, clears timers, and emits disconnected', () => {
    const statuses: string[] = []
    sse.on('status', (s) => statuses.push(s as any))

    sse.start()
    sse.stop()
    expect(statuses[statuses.length - 1]).toBe('disconnected')
    expect(MockEventSource.instances[0].closed).toBe(true)
  })

  it('stop() clears a pending reconnect timer if present (covers L112-L114)', async () => {
    const clearSpy = vi.spyOn(window, 'clearTimeout')

    // Start and force an error to schedule a reconnect (sets reconnectTimer)
    sse.start()
    MockEventSource.instances[0].onerror?.call(MockEventSource.instances[0] as any, new Event('error'))

    // Call stop before the reconnect timeout fires; this should clear the timer
    sse.stop()
    expect(clearSpy).toHaveBeenCalled()

    const before = MockEventSource.instances.length
    // Advance time beyond initial retry (1000ms); no new connection should be created
    await vi.advanceTimersByTimeAsync(1500)
    expect(MockEventSource.instances.length).toBe(before)

    clearSpy.mockRestore()
  })
})
