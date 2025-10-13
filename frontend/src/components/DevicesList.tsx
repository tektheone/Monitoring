import { useQuery } from '@tanstack/react-query'
import { getDevices } from '@/api/client'
import type { DeviceSummary } from '@/types/api'
import DeviceCard from './DeviceCard'
import { DEVICES_REFRESH_MS } from '@/config'

function SkeletonCard() {
  return (
    <div className="animate-pulse rounded-lg border border-gray-200 bg-white p-4">
      <div className="flex items-center justify-between">
        <div className="h-4 w-40 bg-gray-200" />
        <div className="h-3 w-16 bg-gray-200" />
      </div>
      <div className="mt-2 h-3 w-24 bg-gray-200" />
    </div>
  )
}

export default function DevicesList({ onSelect, refreshMs = DEVICES_REFRESH_MS }: { onSelect?: (d: DeviceSummary) => void; refreshMs?: number }) {
  const { data, isLoading, isError, error, refetch, isFetching, dataUpdatedAt } = useQuery<DeviceSummary[]>({
    queryKey: ['devices'],
    queryFn: getDevices,
    refetchInterval: refreshMs,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
    retry: 1,
    placeholderData: (prev) => prev ?? [],
  })

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <SkeletonCard key={i} />
        ))}
      </div>
    )
  }

  if (isError) {
    return (
      <div className="rounded border border-red-200 bg-red-50 p-4 text-red-700">
        Failed to load devices: {String((error as Error).message)}
        <button className="ml-3 underline" onClick={() => refetch()} disabled={isFetching}>
          Retry
        </button>
      </div>
    )
  }

  if (!data || data.length === 0) {
    return <div className="text-gray-500">No devices found.</div>
  }

  const lastUpdated = dataUpdatedAt ? new Date(dataUpdatedAt).toLocaleTimeString() : '—'

  return (
    <div>
      <div className="mb-2 flex items-center gap-3 text-sm text-gray-600">
        <button
          onClick={() => refetch()}
          disabled={isFetching}
          className="rounded border px-2 py-1 text-xs hover:bg-gray-50 disabled:opacity-50"
          title="Refresh devices"
        >
          {isFetching ? 'Refreshing…' : 'Refresh'}
        </button>
        <span>Last updated: {lastUpdated}</span>
        {isFetching && <span className="animate-pulse text-gray-400">(updating)</span>}
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        {data.map((d) => (
          <DeviceCard key={d.id} device={d} onClick={onSelect} />
        ))}
      </div>
    </div>
  )
}
