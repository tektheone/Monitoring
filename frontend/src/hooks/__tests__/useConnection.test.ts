import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useConnection } from '../useConnection'

describe('useConnection', () => {
  const getOnline = () => (navigator as any).onLine as boolean
  let getSpy: ReturnType<typeof vi.spyOn> | null = null

  beforeEach(() => {
    // default to true unless a test overrides
    getSpy = vi.spyOn(window.navigator, 'onLine', 'get').mockReturnValue(true)
  })

  afterEach(() => {
    getSpy?.mockRestore()
  })

  it('returns initial state from navigator.onLine (true)', () => {
    const { result } = renderHook(() => useConnection())
    expect(result.current.online).toBe(true)
    expect(getOnline()).toBe(true)
  })

  it('returns initial state from navigator.onLine (false)', () => {
    getSpy?.mockReturnValue(false)
    const { result } = renderHook(() => useConnection())
    expect(result.current.online).toBe(false)
  })

  it('updates state on offline then online events', () => {
    const { result } = renderHook(() => useConnection())

    act(() => {
      window.dispatchEvent(new Event('offline'))
    })
    expect(result.current.online).toBe(false)

    act(() => {
      window.dispatchEvent(new Event('online'))
    })
    expect(result.current.online).toBe(true)
  })
})
