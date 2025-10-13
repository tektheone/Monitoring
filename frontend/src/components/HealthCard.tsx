import type { HealthStatus } from '@/types/api'

type Props = {
  data?: HealthStatus
  isFetching: boolean
  isError: boolean
}

export default function HealthCard({ data, isFetching, isError }: Props) {
  // Connection status badge: green connected, yellow reconnecting, red disconnected
  const conn = isError ? 'disconnected' : isFetching ? 'reconnecting' : 'connected'
  const color = conn === 'connected' ? 'bg-green-500' : conn === 'reconnecting' ? 'bg-yellow-500' : 'bg-red-500'

  return (
    <div className="rounded-lg border bg-white p-4 shadow-sm w-full max-w-xl">
      <div className="flex items-center justify-between">
        <div className="text-sm text-gray-500">System Health</div>
        <div className="flex items-center gap-2 text-xs">
          <span className={`inline-block h-2.5 w-2.5 rounded-full ${color}`} />
          <span className="capitalize text-gray-700">{conn}</span>
        </div>
      </div>

      <div className="mt-2 flex items-center gap-4 text-sm">
        <div className="flex items-center gap-2">
          <span className={`inline-block h-2.5 w-2.5 rounded-full ${data?.status === 'healthy' ? 'bg-green-500' : 'bg-red-500'}`} />
          <span className="text-gray-700">Server: <span className="font-medium capitalize">{data?.status ?? 'unknown'}</span></span>
        </div>
        <div className="text-gray-700">Devices: <span className="font-medium">{data?.device_count ?? '—'}</span></div>
      </div>

      <div className="mt-1 text-xs text-gray-500">Timestamp: {data?.timestamp ?? '—'}</div>
    </div>
  )
}
