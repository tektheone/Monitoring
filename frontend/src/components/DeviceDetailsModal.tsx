import { useQuery } from '@tanstack/react-query'
import { getDeviceStats } from '@/api/client'
import type { DeviceStats, DeviceSummary } from '@/types/api'
import { formatDuration } from '@/utils/duration'
import { timeAgo } from '@/utils/time'
import { DEVICE_STATS_REFRESH_MS } from '@/config'

type Props = {
  device: DeviceSummary
  onClose: () => void
}

export default function DeviceDetailsModal({ device, onClose }: Props) {
  const { data, isLoading, isError, error, refetch, isFetching, dataUpdatedAt } = useQuery<null | DeviceStats>({
    queryKey: ['device-stats', device.id],
    queryFn: () => getDeviceStats(device.id),
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
    refetchInterval: DEVICE_STATS_REFRESH_MS,
    retry: 1,
  })

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/40" onClick={onClose} />
      <div className="relative z-10 w-full max-w-lg rounded-lg bg-white p-5 shadow-lg">
        <div className="flex items-start justify-between">
          <div>
            <div className="text-xs text-gray-500">Device</div>
            <div className="font-mono text-sm break-all">{device.id}</div>
          </div>
          <button onClick={onClose} className="text-gray-500 hover:text-gray-700">✕</button>
        </div>

        <div className="mt-3 flex items-center gap-2">
          <span className={`inline-block h-2.5 w-2.5 rounded-full ${device.status === 'online' ? 'bg-green-500' : 'bg-red-500'}`} />
          <span className="text-sm capitalize text-gray-700">{device.status}</span>
          <span className="text-xs text-gray-400">· Last seen {timeAgo(device.last_seen)}</span>
        </div>

        <div className="mt-4">
          <div className="mb-2 flex items-center justify-between">
            <h3 className="text-sm font-semibold">Statistics</h3>
            <div className="flex items-center gap-3">
              <span className="text-xs text-gray-500">Last updated: {dataUpdatedAt ? new Date(dataUpdatedAt).toLocaleTimeString() : '—'}</span>
              <button
              className="rounded border px-2 py-1 text-xs hover:bg-gray-50 disabled:opacity-50"
              onClick={() => refetch()}
              disabled={isFetching}
            >
                {isFetching ? 'Refreshing…' : 'Refresh'}
              </button>
            </div>
          </div>

          {isLoading ? (
            <div className="text-sm text-gray-500">Loading stats…</div>
          ) : isError ? (
            <div className="rounded border border-red-200 bg-red-50 p-3 text-sm text-red-700">
              Failed to load stats: {String((error as Error).message)}
            </div>
          ) : data === null ? (
            <div className="text-sm text-gray-500">No statistics yet for this device.</div>
          ) : (
            (() => {
              const stats = data as DeviceStats
              return (
                <div className="space-y-3">
              <div>
                <div className="flex items-center justify-between text-sm">
                  <span>Uptime</span>
                  <span className="font-medium">{stats.uptime.toFixed(3)}%</span>
                </div>
                <div className="mt-1 h-2 w-full rounded bg-gray-200">
                  <div
                    className="h-2 rounded bg-green-500"
                    style={{ width: `${Math.max(0, Math.min(100, stats.uptime))}%` }}
                  />
                </div>
              </div>
              <div className="text-sm">
                <div className="text-gray-600">Average upload time</div>
                <div className="font-medium">{formatDuration(stats.avg_upload_time)}</div>
              </div>
                </div>
              )
            })()
          )}
        </div>
      </div>
    </div>
  )
}
