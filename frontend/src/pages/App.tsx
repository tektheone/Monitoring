import { useIsFetching, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { getDevices } from '@/api/client'
import DevicesList from '@/components/DevicesList'
import DeviceDetailsModal from '@/components/DeviceDetailsModal'
import type { DeviceSummary } from '@/types/api'
import Layout from '@/components/Layout'
import { sse } from '@/lib/sse'
import { useSSE } from '@/hooks/useSSE'
import { useConnection } from '@/hooks/useConnection'

function App() {
  const [selected, setSelected] = useState<DeviceSummary | null>(null)
  const globalFetching = useIsFetching()
  const qc = useQueryClient()
  const { status: sseStatus } = useSSE()
  const { online } = useConnection()
  // Removed health polling and card per request; we rely on connection status only

  // Bind SSE messages to React Query caches
  useEffect(() => {
    const offHealth = sse.on('health:update', (payload) => {
      // Health updates are ignored in UI now, but keep cache if other parts want it later
      qc.setQueryData(['health'], payload)
    })
    const offDevices = sse.on('devices:update', (payload) => {
      // payload can be full list or delta. If full list present, set it directly
      if (Array.isArray(payload?.devices)) {
        qc.setQueryData(['devices'], payload.devices)
      } else if (payload && payload.id) {
        // merge single device update into list if exists
        qc.setQueryData<DeviceSummary[] | undefined>(['devices'], (prev) => {
          if (!prev) return prev
          const idx = prev.findIndex((d) => d.id === payload.id)
          if (idx === -1) return prev
          const next = prev.slice()
          next[idx] = { ...next[idx], ...payload }
          return next
        })
      }
    })
    return () => {
      offHealth()
      offDevices()
    }
  }, [qc])

  // If SSE connected but devices cache is empty (possible race with initial snapshot), fetch once
  useEffect(() => {
    if (sseStatus === 'connected') {
      const current = qc.getQueryData<DeviceSummary[] | undefined>(['devices'])
      if (!current || current.length === 0) {
        getDevices()
          .then((list) => qc.setQueryData(['devices'], list))
          .catch(() => {/* ignore; SSE may populate shortly */})
      }
    }
  }, [sseStatus, qc])

  // Derive backend status primarily from SSE
  const backendStatus: 'connected' | 'reconnecting' | 'disconnected' | 'offline' = !online
    ? 'offline'
    : sseStatus === 'connected'
    ? 'connected'
    : sseStatus === 'connecting'
    ? 'reconnecting'
    : 'disconnected'

  // Do not early-return on loading or error; we want the dashboard visible.

  return (
    <Layout
      header={{}}
      backendStatus={backendStatus}
    >
      {/* Global fetching indicator */}
      {globalFetching > 0 && (
        <div className="fixed right-4 top-4 z-50 rounded-full bg-gray-800 px-3 py-1 text-xs text-white shadow">
          Updating…
        </div>
      )}

      <div className="mb-2 flex items-center justify-between">
        <h2 className="text-xl font-semibold">Devices</h2>
      </div>

      <DevicesList onSelect={(d) => setSelected(d)} />

      {selected && (
        <DeviceDetailsModal device={selected} onClose={() => setSelected(null)} />
      )}
    </Layout>
  )
}

export default App
