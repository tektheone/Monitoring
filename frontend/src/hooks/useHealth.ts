import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { api } from '../api/client'
import type { HealthResponse, ApiResult } from '../api/types'

export type HealthState = {
  loading: boolean
  error: string | null
  statusCode: number | null
  data: HealthResponse | null
  lastUpdated: number | null
}

export function useHealth(pollMs: number = 10_000) {
  const [state, setState] = useState<HealthState>({
    loading: true,
    error: null,
    statusCode: null,
    data: null,
    lastUpdated: null,
  })

  const timerRef = useRef<number | null>(null)
  const isMounted = useRef(true)

  const fetchOnce = useCallback(async () => {
    let res: ApiResult<HealthResponse>
    try {
      res = await api.health({ method: 'GET' })
    } catch (e: any) {
      if (!isMounted.current) return
      setState(prev => ({
        ...prev,
        loading: false,
        error: e?.message ?? 'Network error',
        statusCode: 0,
        lastUpdated: Date.now(),
      }))
      return
    }

    if (!isMounted.current) return

    if (!res.ok) {
      setState({
        loading: false,
        error: res.error,
        statusCode: res.status,
        data: null,
        lastUpdated: Date.now(),
      })
      return
    }

    setState({
      loading: false,
      error: null,
      statusCode: res.status,
      data: res.data,
      lastUpdated: Date.now(),
    })
  }, [])

  useEffect(() => {
    isMounted.current = true
    // initial fetch
    fetchOnce()

    // polling
    if (pollMs > 0) {
      timerRef.current = window.setInterval(fetchOnce, pollMs)
    }

    return () => {
      isMounted.current = false
      if (timerRef.current) window.clearInterval(timerRef.current)
    }
  }, [fetchOnce, pollMs])

  const derived = useMemo(() => {
    const code = state.statusCode ?? 0
    let health: 'healthy' | 'degraded' | 'down' = 'down'
    if (code === 200) health = 'healthy'
    else if (code > 0 && code !== 200) health = 'degraded'

    return { ...state, health }
  }, [state])

  return { ...derived, refetch: fetchOnce }
}
