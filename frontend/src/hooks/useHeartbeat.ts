import { sendHeartbeat } from '@/api/client'

export type HeartbeatDeps = {
  setBusy: (b: boolean) => void
  setMsg: (m: string | null) => void
  refetch: () => unknown | Promise<unknown>
}

// Testable flow function for sending a heartbeat and updating UI state
export async function sendHeartbeatFlow(rawId: string, deps: HeartbeatDeps): Promise<void> {
  const id = rawId.trim()
  if (!id) return

  deps.setBusy(true)
  deps.setMsg(null)
  try {
    await sendHeartbeat(id)
    deps.setMsg('Heartbeat sent')
    await deps.refetch()
  } catch (e) {
    deps.setMsg((e as Error).message)
  } finally {
    deps.setBusy(false)
  }
}
