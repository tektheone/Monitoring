import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { getHealth } from '@/api/client'
import DevicesList from '@/components/DevicesList'
import DeviceDetailsModal from '@/components/DeviceDetailsModal'
import type { DeviceSummary } from '@/types/api'

function App() {
  const [selected, setSelected] = useState<DeviceSummary | null>(null)
  const { data, isLoading, error } = useQuery({
    queryKey: ['health'],
    queryFn: getHealth,
  })

  if (isLoading) return <div className="p-6">Loading…</div>
  if (error) return <div className="p-6 text-red-600">{String((error as Error).message)}</div>

  return (
    <div className="min-h-screen bg-gray-50 text-gray-900 p-6">
      <h1 className="text-2xl font-semibold mb-4">SafelyYou Fleet Monitoring</h1>
      <div className="rounded-lg border bg-white p-4 shadow-sm w-full max-w-xl mb-6">
        <div className="text-sm text-gray-500">Backend Health</div>
        <div className="mt-1 flex items-center gap-2">
          <span className={`inline-block h-3 w-3 rounded-full ${data?.status === 'healthy' ? 'bg-green-500' : 'bg-red-500'}`} />
          <span className="font-medium">{data?.status}</span>
        </div>
        <div className="mt-2 text-sm text-gray-600">Devices: {data?.device_count}</div>
        <div className="mt-1 text-xs text-gray-400">Timestamp: {data?.timestamp}</div>
      </div>

      <h2 className="text-xl font-semibold mb-3">Devices</h2>
      <DevicesList onSelect={(d) => setSelected(d)} />

      {selected && (
        <DeviceDetailsModal device={selected} onClose={() => setSelected(null)} />
      )}
    </div>
  )
}

export default App
