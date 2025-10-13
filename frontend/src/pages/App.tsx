import { useQuery } from '@tanstack/react-query'

function App() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['health'],
    queryFn: async () => {
      const res = await fetch('/health')
      if (!res.ok) throw new Error('Failed to fetch health')
      return res.json() as Promise<{ status: string; timestamp: number; device_count: number }>
    },
  })

  if (isLoading) return <div className="p-6">Loading…</div>
  if (error) return <div className="p-6 text-red-600">{String((error as Error).message)}</div>

  return (
    <div className="min-h-screen bg-gray-50 text-gray-900 p-6">
      <h1 className="text-2xl font-semibold mb-4">SafelyYou Fleet Monitoring</h1>
      <div className="rounded-lg border bg-white p-4 shadow-sm w-full max-w-xl">
        <div className="text-sm text-gray-500">Backend Health</div>
        <div className="mt-1 flex items-center gap-2">
          <span className={`inline-block h-3 w-3 rounded-full ${data?.status === 'healthy' ? 'bg-green-500' : 'bg-red-500'}`} />
          <span className="font-medium">{data?.status}</span>
        </div>
        <div className="mt-2 text-sm text-gray-600">Devices: {data?.device_count}</div>
        <div className="mt-1 text-xs text-gray-400">Timestamp: {data?.timestamp}</div>
      </div>
    </div>
  )
}

export default App
