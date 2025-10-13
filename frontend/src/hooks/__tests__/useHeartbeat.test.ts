import { describe, it, expect, vi, beforeEach } from 'vitest'
import { sendHeartbeatFlow } from '../useHeartbeat'

// Mock API client
vi.mock('@/api/client', () => ({
  sendHeartbeat: vi.fn(async () => undefined),
}))

const makeDeps = () => {
  return {
    setBusy: vi.fn(),
    setMsg: vi.fn(),
    refetch: vi.fn(async () => undefined),
  }
}

describe('sendHeartbeatFlow', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns early for empty id ("") and does not call any deps (covers L12)', async () => {
    const deps = makeDeps()
    const { sendHeartbeat } = await import('@/api/client') as any

    await sendHeartbeatFlow('', deps)

    expect(deps.setBusy).not.toHaveBeenCalled()
    expect(deps.setMsg).not.toHaveBeenCalled()
    expect(deps.refetch).not.toHaveBeenCalled()
    expect(sendHeartbeat).not.toHaveBeenCalled()
  })

  it('returns early for whitespace-only id and does not call any deps (covers L12)', async () => {
    const deps = makeDeps()
    const { sendHeartbeat } = await import('@/api/client') as any

    await sendHeartbeatFlow('   \t  \n  ', deps)

    expect(deps.setBusy).not.toHaveBeenCalled()
    expect(deps.setMsg).not.toHaveBeenCalled()
    expect(deps.refetch).not.toHaveBeenCalled()
    expect(sendHeartbeat).not.toHaveBeenCalled()
  })
})
