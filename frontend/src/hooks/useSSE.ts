import { useEffect, useMemo, useState } from 'react'
import { sse } from '@/lib/sse'

export type SSEStatus = 'disconnected' | 'connecting' | 'connected'

export function useSSE() {
  const [status, setStatus] = useState<SSEStatus>(sse.getStatus())

  useEffect(() => {
    sse.start()
    const off = sse.on('status', (st) => setStatus(st))
    return () => {
      off()
    }
  }, [])

  const connected = status === 'connected'
  return useMemo(() => ({ status, connected }), [status, connected])
}
