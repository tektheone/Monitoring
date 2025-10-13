import { describe, it, expect, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'

// Mock the sse client used by the hook
import type { SSEStatus } from '../useSSE'

const mocked = vi.hoisted(() => {
  return {
    start: vi.fn<() => void>(),
    getStatus: vi.fn<() => SSEStatus>(),
    off: vi.fn<() => void>(),
    statusHandler: null as null | ((st: SSEStatus) => void),
  }
})

vi.mock('@/lib/sse', () => ({
  sse: {
    start: mocked.start,
    getStatus: mocked.getStatus,
    on: vi.fn((type: string, handler: (st: SSEStatus) => void) => {
      if (type === 'status') mocked.statusHandler = handler
      return mocked.off
    })
  }
}))

import { useSSE } from '../useSSE'

describe('useSSE', () => {
  it('returns initial status from sse.getStatus() and calls sse.start()', () => {
    mocked.getStatus.mockReturnValue('disconnected')
    const { result } = renderHook(() => useSSE())

    expect(mocked.start).toHaveBeenCalledTimes(1)
    expect(result.current.status).toBe('disconnected')
    expect(result.current.connected).toBe(false)
  })

  it('updates status when sse status event fires and cleans up on unmount', () => {
    mocked.getStatus.mockReturnValue('connecting')
    const { result, unmount } = renderHook(() => useSSE())

    expect(result.current.status).toBe('connecting')
    expect(result.current.connected).toBe(false)

    act(() => {
      mocked.statusHandler?.('connected')
    })

    expect(result.current.status).toBe('connected')
    expect(result.current.connected).toBe(true)

    unmount()
    // In React 18, effects may be invoked twice under StrictMode in tests.
    // Assert cleanup was executed at least once rather than an exact count.
    expect(mocked.off).toHaveBeenCalled()
  })
})
