import React from 'react'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { DeviceSummary, DeviceStats } from '@/types/api'

// Disable refetch interval to avoid timers
vi.mock('@/config', () => ({
  DEVICE_STATS_REFRESH_MS: false,
}))

// Mock API client and verify queryFn executes it
vi.mock('@/api/client', () => ({
  getDeviceStats: vi.fn(async (id: string) => ({
    uptime: 88.5,
    avg_upload_time: '250ms',
  } as DeviceStats)),
}))

// Import component after mocks
import DeviceDetailsModal from '../DeviceDetailsModal'

function wrap(ui: React.ReactElement, qc?: QueryClient) {
  const client = qc ?? new QueryClient({ defaultOptions: { queries: { retry: false, staleTime: 0 } } })
  return render(<QueryClientProvider client={client}>{ui}</QueryClientProvider>)
}

function makeDevice(partial: Partial<DeviceSummary> = {}): DeviceSummary {
  return {
    id: 'dev-qf',
    status: 'online',
    last_seen: new Date().toISOString(),
    ...partial,
  }
}

describe('DeviceDetailsModal queryFn execution', () => {
  it('calls getDeviceStats via useQuery queryFn with device id', async () => {
    const { getDeviceStats } = await import('@/api/client') as any
    wrap(<DeviceDetailsModal device={makeDevice()} onClose={() => {}} />)

    // Wait for some content that appears after data resolves
    await screen.findByText('88.500%')

    expect(getDeviceStats).toHaveBeenCalledTimes(1)
    expect(getDeviceStats).toHaveBeenCalledWith('dev-qf')
  })
})
