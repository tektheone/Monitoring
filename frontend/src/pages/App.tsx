import { useIsFetching, useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { getHealth } from '@/api/client'
import DevicesList from '@/components/DevicesList'
import DeviceDetailsModal from '@/components/DeviceDetailsModal'
import type { DeviceSummary } from '@/types/api'
import Layout from '@/components/Layout'
import HealthCard from '@/components/HealthCard'

function App() {
  const [selected, setSelected] = useState<DeviceSummary | null>(null)
  const globalFetching = useIsFetching()
  const { data, error, isFetching } = useQuery({
    queryKey: ['health'],
    queryFn: getHealth,
    refetchInterval: 15_000,
    retry: 1,
  })

  // Derive footer/backend status from health query
  const backendStatus: 'connected' | 'reconnecting' | 'disconnected' = error
    ? 'disconnected'
    : isFetching
    ? 'reconnecting'
    : 'connected'

  // Do not early-return on loading or error; we want the dashboard visible.

  return (
    <Layout
      header={{
        title: 'Fleet Monitoring Dashboard',
        healthStatus: data?.status,
        deviceCount: data?.device_count,
        lastUpdated: data?.timestamp,
      }}
      backendStatus={backendStatus}
    >
      {/* Global fetching indicator */}
      {globalFetching > 0 && (
        <div className="fixed right-4 top-4 z-50 rounded-full bg-gray-800 px-3 py-1 text-xs text-white shadow">
          Updating…
        </div>
      )}

      {/* HealthCard */}
      <div className="mb-4">
        <HealthCard data={data} isFetching={isFetching} isError={!!error} />
      </div>

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
