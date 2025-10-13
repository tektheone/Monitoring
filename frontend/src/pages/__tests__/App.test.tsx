import React from 'react'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { DeviceSummary } from '@/types/api'

// Utilities
function wrap(ui: React.ReactElement, qc?: QueryClient) {
  const client = qc ?? new QueryClient({ defaultOptions: { queries: { retry: false, staleTime: 0 } } })
  return render(<QueryClientProvider client={client}>{ui}</QueryClientProvider>)
}

// Mock SSE with capturable handlers
const sseMock = (() => {
  const handlers: Record<string, Function[]> = {}
  return {
    on: vi.fn((type: string, fn: Function) => {
      (handlers[type] ||= []).push(fn)
      return () => {
        handlers[type] = (handlers[type] || []).filter((h) => h !== fn)
      }
    }),
    emit(type: string, payload: any) {
      (handlers[type] || []).forEach((h) => h(payload))
    },
  }
})()

// Base mocks
vi.mock('@/lib/sse', () => ({ sse: sseMock }))

// Mocks we vary per test
vi.mock('@/hooks/useSSE', () => ({ useSSE: vi.fn(() => ({ status: 'disconnected' })) }))
vi.mock('@/hooks/useConnection', () => ({ useConnection: vi.fn(() => ({ online: true })) }))
vi.mock('@/api/client', () => ({ getDevices: vi.fn(async () => [] as DeviceSummary[]) }))

// Layout spy to observe backendStatus prop without rendering full layout
const layoutSpy = vi.hoisted(() => ({
  default: ({ backendStatus, children }: any) => (
    <div>
      <div data-testid="backend-status">{backendStatus}</div>
      {children}
    </div>
  ),
}))
vi.mock('@/components/Layout', () => layoutSpy)

describe('App page', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('ignores single-device update when devices cache is undefined (covers L33)', async () => {
    const App = (await import('../App')).default
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false, staleTime: 0 } } })
    // Ensure cache is undefined initially
    expect(qc.getQueryData(['devices'])).toBeUndefined()

    // Mount app to bind SSE handlers
    wrap(<App />, qc)

    // Emit a partial update while cache is undefined
    sseMock.emit('devices:update', { id: 'x', status: 'offline' })

    // Cache should remain undefined (no-op)
    expect(qc.getQueryData(['devices'])).toBeUndefined()
  })

  it('ignores single-device update when device id not found (covers L35)', async () => {
    const App = (await import('../App')).default
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false, staleTime: 0 } } })
    const initial: DeviceSummary[] = [
      { id: 'a', status: 'online', last_seen: new Date().toISOString() } as DeviceSummary,
    ]
    qc.setQueryData<DeviceSummary[]>(['devices'], initial)

    // Mount app to bind SSE handlers
    wrap(<App />, qc)

    // Emit update for non-existent id 'b'
    sseMock.emit('devices:update', { id: 'b', status: 'offline' })

    // Cache reference should be preserved (returned prev)
    const after = qc.getQueryData<DeviceSummary[]>(['devices'])
    expect(after).toBe(initial)
  })

  it('updates health cache on health:update SSE message', async () => {
    const App = (await import('../App')).default
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false, staleTime: 0 } } })
    wrap(<App />, qc)

    const health = { status: 'ok', timestamp: Date.now(), device_count: 5 }
    sseMock.emit('health:update', health)
    expect(qc.getQueryData(['health'])).toEqual(health)
  })

  it('opens details modal when a device is selected', async () => {
    vi.resetModules()
    // Mock DevicesList to provide a button that invokes onSelect
    const fakeDevice = { id: 'z-9', status: 'online', last_seen: new Date().toISOString() } as DeviceSummary
    vi.doMock('@/components/DevicesList', () => ({
      default: ({ onSelect }: any) => (
        <button onClick={() => onSelect(fakeDevice)}>Select Device</button>
      ),
    }))
    // Mock DeviceDetailsModal to a marker component
    vi.doMock('@/components/DeviceDetailsModal', () => ({
      default: ({ device, onClose }: any) => (
        <div data-testid="device-modal">Modal for {device.id}<button onClick={onClose}>Close</button></div>
      ),
    }))

    const App = (await import('../App')).default
    const rendered = wrap(<App />)

    // Open modal
    screen.getByRole('button', { name: /select device/i }).click()
    expect(await screen.findByTestId('device-modal')).toBeInTheDocument()
    rendered.unmount()
  })

  it('binds SSE handlers and updates devices cache on devices:update full list', async () => {
    const App = (await import('../App')).default
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false, staleTime: 0 } } })
    wrap(<App />, qc)

    // Emit full list update
    const list: DeviceSummary[] = [{ id: 'x', status: 'online', last_seen: new Date().toISOString() }]
    sseMock.emit('devices:update', { devices: list })

    // Cache should be updated
    const data = qc.getQueryData<DeviceSummary[]>(['devices'])
    expect(data).toEqual(list)
  })

  it('merges single device update into existing devices cache (covers L32-L40)', async () => {
    const App = (await import('../App')).default
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false, staleTime: 0 } } })
    // Seed cache with two devices
    const initial: DeviceSummary[] = [
      { id: 'a', status: 'online', last_seen: new Date().toISOString() } as DeviceSummary,
      { id: 'b', status: 'online', last_seen: new Date().toISOString() } as DeviceSummary,
    ]
    qc.setQueryData<DeviceSummary[]>(['devices'], initial)

    // Mount app to bind SSE handlers
    wrap(<App />, qc)

    // Emit a partial update for device 'b'
    const updated = { id: 'b', status: 'offline' }
    sseMock.emit('devices:update', updated)

    const after = qc.getQueryData<DeviceSummary[]>(['devices'])!
    expect(after).toHaveLength(2)
    // Device 'a' unchanged
    expect(after[0]).toMatchObject(initial[0])
    // Device 'b' merged (status changed to offline, id preserved)
    expect(after[1].id).toBe('b')
    expect(after[1].status).toBe('offline')
  })

  it('fetches devices once when SSE is connected and cache empty', async () => {
    vi.resetModules()
    vi.doMock('@/hooks/useSSE', () => ({ useSSE: () => ({ status: 'connected' }) }))
    const { getDevices } = await import('@/api/client') as any
    const App = (await import('../App')).default

    wrap(<App />)
    // getDevices is invoked in effect when connected and cache empty
    expect(getDevices).toHaveBeenCalled()
  })

  it('shows global updating indicator when queries are fetching', async () => {
    vi.resetModules()
    // Mock useIsFetching to return > 0
    vi.doMock('@tanstack/react-query', async (orig) => {
      const actual = await (orig as any)()
      return { ...actual, useIsFetching: () => 1 }
    })
    const App = (await import('../App')).default
    wrap(<App />)
    // Exact match for the global banner 'Updating…' (avoid matching '(updating)')
    expect(screen.getByText('Updating…')).toBeInTheDocument()
  })

  it('sets backendStatus: offline | reconnecting | connected | disconnected', async () => {
    vi.resetModules()
    // Case 1: offline
    vi.doMock('@/hooks/useConnection', () => ({ useConnection: () => ({ online: false }) }))
    vi.doMock('@/hooks/useSSE', () => ({ useSSE: () => ({ status: 'disconnected' }) }))
    let App = (await import('../App')).default
    const { unmount } = wrap(<App />)
    expect(screen.getByTestId('backend-status').textContent).toBe('offline')
    unmount()

    // Case 2: reconnecting
    vi.resetModules()
    vi.doMock('@/hooks/useConnection', () => ({ useConnection: () => ({ online: true }) }))
    vi.doMock('@/hooks/useSSE', () => ({ useSSE: () => ({ status: 'connecting' }) }))
    App = (await import('../App')).default
    let rendered = wrap(<App />)
    expect(screen.getByTestId('backend-status').textContent).toBe('reconnecting')
    rendered.unmount()

    // Case 3: connected
    vi.resetModules()
    vi.doMock('@/hooks/useConnection', () => ({ useConnection: () => ({ online: true }) }))
    vi.doMock('@/hooks/useSSE', () => ({ useSSE: () => ({ status: 'connected' }) }))
    App = (await import('../App')).default
    rendered = wrap(<App />)
    expect(screen.getByTestId('backend-status').textContent).toBe('connected')
    rendered.unmount()

    // Case 4: disconnected (online but not connecting/connected)
    vi.resetModules()
    vi.doMock('@/hooks/useConnection', () => ({ useConnection: () => ({ online: true }) }))
    vi.doMock('@/hooks/useSSE', () => ({ useSSE: () => ({ status: 'idle' }) }))
    App = (await import('../App')).default
    rendered = wrap(<App />)
    expect(screen.getByTestId('backend-status').textContent).toBe('disconnected')
    rendered.unmount()
  })
})
