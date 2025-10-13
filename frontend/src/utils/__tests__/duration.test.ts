import { describe, it, expect } from 'vitest'
import { formatDuration } from '../duration'

describe('formatDuration', () => {
  it('returns em dash for falsy input', () => {
    expect(formatDuration()).toBe('—')
    expect(formatDuration('')).toBe('—')
  })

  it('returns original string when no duration tokens are found', () => {
    expect(formatDuration('no duration here')).toBe('no duration here')
  })

  it('formats concatenated tokens with spaces', () => {
    expect(formatDuration('1h2m3s')).toBe('1 h 2 m 3 s')
    expect(formatDuration('250ms')).toBe('250 ms')
  })

  it('normalizes microseconds unit to µs from us', () => {
    expect(formatDuration('10us')).toBe('10 µs')
    expect(formatDuration('5µs')).toBe('5 µs')
  })
})
