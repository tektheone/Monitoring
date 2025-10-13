import React from 'react'
import { render } from '@testing-library/react'

// Hoisted spies
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

vi.mock('@/api/client', () => ({
  getDevices: vi.fn(),
}))

// Case 1: SSE connected => refetchInterval should be false
vi.mock('@/hooks/useSSE', () => ({
  useSSE: () => ({ status: 'connected', connected: true }),
}))

import DevicesList from '../DevicesList'

describe('DevicesList query options', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('sets refetchInterval to false when SSE is connected', () => {
    rq.useQuery.mockReturnValue({
      data: [],
      refetch: vi.fn(),
      isFetching: false,
      dataUpdatedAt: Date.now(),
    })

    render(<DevicesList />)

    // Assert that useQuery was called with refetchInterval false
    expect(rq.useQuery).toHaveBeenCalled()
    const callArg = rq.useQuery.mock.calls[0][0]
    expect(callArg.refetchInterval).toBe(false)
  })
})
