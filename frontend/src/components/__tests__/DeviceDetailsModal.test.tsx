import { render, screen, fireEvent } from '@testing-library/react'
import type { DeviceSummary, DeviceStats } from '@/types/api'
import React from 'react'

// Disable refetch interval in the component during tests
vi.mock('@/config', () => ({
  DEVICE_STATS_REFRESH_MS: false,
}))

// Mock react-query useQuery and provide a no-op provider
const rq = vi.hoisted(() => ({
  useQuery: vi.fn(),
}))
vi.mock('@tanstack/react-query', async (orig) => {
  const actual = await orig<typeof import('@tanstack/react-query')>()
  return {
    ...actual,
    useQuery: rq.useQuery,
    QueryClientProvider: ({ children }: any) => children,
  }
})

// Mock API client (not used directly since useQuery is mocked)
vi.mock('@/api/client', () => ({
  getDeviceStats: vi.fn(),
}))

// Import component after mocks so they apply
import DeviceDetailsModal from '../DeviceDetailsModal'

function wrap(ui: React.ReactElement) {
  return render(ui)
}

function makeDevice(partial: Partial<DeviceSummary> = {}): DeviceSummary {
  return {
    id: 'dev-1',
    status: 'online',
    last_seen: new Date().toISOString(),
    ...partial,
  }
}

describe('DeviceDetailsModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading state while fetching', () => {
    // Simulate loading state
    rq.useQuery.mockReturnValue({
      isLoading: true,
      isError: false,
      data: undefined,
      error: null,
      refetch: vi.fn(),
      isFetching: true,
      dataUpdatedAt: 0,
    })

    const onClose = vi.fn()
    wrap(<DeviceDetailsModal device={makeDevice()} onClose={onClose} />)
    expect(screen.getByText(/loading stats/i)).toBeInTheDocument()
    // test ends here; React Query promise can remain pending
  })

  it('renders error state when fetch fails', async () => {
    rq.useQuery.mockReturnValue({
      isLoading: false,
      isError: true,
      data: undefined,
      error: new Error('Oops'),
      refetch: vi.fn(),
      isFetching: false,
      dataUpdatedAt: 0,
    })

    wrap(<DeviceDetailsModal device={makeDevice()} onClose={() => {}} />)

    await screen.findByText(/failed to load stats/i)
    expect(screen.getByText(/Oops/)).toBeInTheDocument()
  })

  it('renders null state when no stats yet', async () => {
    rq.useQuery.mockReturnValue({
      isLoading: false,
      isError: false,
      data: null,
      error: null,
      refetch: vi.fn(),
      isFetching: false,
      dataUpdatedAt: Date.now(),
    })

    wrap(<DeviceDetailsModal device={makeDevice()} onClose={() => {}} />)

    await screen.findByText(/no statistics yet/i)
  })

  it('renders stats and supports refresh', async () => {
    const stats: DeviceStats = {
      uptime: 75,
      avg_upload_time: '12345ms',
    } as any

    rq.useQuery.mockReturnValue({
      isLoading: false,
      isError: false,
      data: stats,
      error: null,
      refetch: vi.fn(),
      isFetching: false,
      dataUpdatedAt: Date.now(),
    })

    const onClose = vi.fn()
    wrap(<DeviceDetailsModal device={makeDevice()} onClose={onClose} />)

    // header info
    expect(screen.getByText('dev-1')).toBeInTheDocument()
    expect(screen.getByText(/online/i)).toBeInTheDocument()

    // waits for stats (percentage renders)
    await screen.findByText('75.000%')

    expect(screen.getByText(/Average upload time/i)).toBeInTheDocument()

    // Click refresh triggers refetch
    const btn = screen.getByRole('button', { name: /refresh/i })
    fireEvent.click(btn)

    // Close via button
    const closeBtn = screen.getByRole('button', { name: '✕' })
    fireEvent.click(closeBtn)

    // Close via overlay (backdrop)
    // If overlay is not a button role, select the backdrop by its classes
    const backdrop = document.querySelector('.absolute.inset-0') as HTMLElement
    if (backdrop) fireEvent.click(backdrop)

    expect(onClose).toHaveBeenCalledTimes(2)
  })

  it('shows green dot for online and red dot for offline', async () => {
    const stats: DeviceStats = { uptime: 50, avg_upload_time: '1s' } as any
    // online case
    rq.useQuery.mockReturnValue({
      isLoading: false,
      isError: false,
      data: stats,
      error: null,
      refetch: vi.fn(),
      isFetching: false,
      dataUpdatedAt: Date.now(),
    })
    wrap(<DeviceDetailsModal device={makeDevice({ status: 'online' })} onClose={() => {}} />)
    const onlineText = screen.getByText(/online/i)
    const onlineDot = onlineText.previousElementSibling as HTMLElement
    expect(onlineDot.className).toContain('bg-green-500')

    // offline case
    rq.useQuery.mockReturnValue({
      isLoading: false,
      isError: false,
      data: stats,
      error: null,
      refetch: vi.fn(),
      isFetching: false,
      dataUpdatedAt: Date.now(),
    })
    wrap(<DeviceDetailsModal device={makeDevice({ status: 'offline' })} onClose={() => {}} />)
    const offlineText = screen.getAllByText(/offline/i)[0]
    const offlineDot = offlineText.previousElementSibling as HTMLElement
    expect(offlineDot.className).toContain('bg-red-500')
  })
})
