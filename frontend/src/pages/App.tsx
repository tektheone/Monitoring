import { useIsFetching, useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { getHealth } from '@/api/client'
import DevicesList from '@/components/DevicesList'
import DeviceDetailsModal from '@/components/DeviceDetailsModal'
import type { DeviceSummary } from '@/types/api'
import { DEVICES_REFRESH_MS } from '@/config'
import Layout from '@/components/Layout'

function App() {
  const [selected, setSelected] = useState<DeviceSummary | null>(null)
  const [devicesRefreshMs, setDevicesRefreshMs] = useState<number>(DEVICES_REFRESH_MS)
  const globalFetching = useIsFetching()
  const { data, isLoading, error } = useQuery({
    queryKey: ['health'],
    queryFn: getHealth,
  })

  if (isLoading) return <div className="p-6">Loading…</div>
  if (error) return <div className="p-6 text-red-600">{String((error as Error).message)}</div>

  return (
    <Layout
      header={{
        title: 'Fleet Monitoring Dashboard',
        healthStatus: data?.status,
        deviceCount: data?.device_count,
        lastUpdated: data?.timestamp,
      }}
    >
      {/* Global fetching indicator */}
      {globalFetching > 0 && (
        <div className="fixed right-4 top-4 z-50 rounded-full bg-gray-800 px-3 py-1 text-xs text-white shadow">
          Updating…
        </div>
      )}

      <div className="mb-2 flex items-center justify-between">
        <h2 className="text-xl font-semibold">Devices</h2>
        <label className="text-sm text-gray-600 flex items-center gap-2">
          Refresh every
          <select
            className="rounded border px-2 py-1 text-sm bg-white"
            value={devicesRefreshMs}
            onChange={(e) => setDevicesRefreshMs(Number(e.target.value))}
          >
            <option value={10_000}>10s</option>
            <option value={15_000}>15s</option>
            <option value={30_000}>30s</option>
            <option value={60_000}>60s</option>
          </select>
        </label>
      </div>

      <DevicesList refreshMs={devicesRefreshMs} onSelect={(d) => setSelected(d)} />

      {selected && (
        <DeviceDetailsModal device={selected} onClose={() => setSelected(null)} />
      )}
    </Layout>
  )
}

export default App
