import { describe, it, expect, vi, afterEach, beforeEach } from 'vitest'
import { timeAgo } from '../time'

describe('timeAgo', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  it('returns hours when less than 24h', () => {
    // Now: 2025-01-01T12:00:00Z
    vi.setSystemTime(new Date('2025-01-01T12:00:00.000Z'))
    // 1.5 hours ago -> expect 1h ago (floored)
    const iso = '2025-01-01T10:30:00.000Z'
    expect(timeAgo(iso)).toBe('1h ago')
  })

  it('returns days when 24h or more', () => {
    // Now: 2025-01-01T12:00:00Z
    vi.setSystemTime(new Date('2025-01-01T12:00:00.000Z'))
    // 2 days ago
    const iso = '2024-12-30T12:00:00.000Z'
    expect(timeAgo(iso)).toBe('2d ago')
  })
})
