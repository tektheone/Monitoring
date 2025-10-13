import type { DeviceSummary } from '@/types/api'
import { timeAgo } from '@/utils/time'

type Props = {
  device: DeviceSummary
  onClick?: (device: DeviceSummary) => void
}

export default function DeviceCard({ device, onClick }: Props) {
  const isOnline = device.status === 'online'
  const statusColor = isOnline ? 'bg-green-500' : 'bg-red-500'
  const borderColor = isOnline ? 'border-green-200' : 'border-gray-200'

  return (
    <button
      onClick={() => onClick?.(device)}
      className={`text-left rounded-lg border ${borderColor} bg-white p-4 shadow-sm hover:shadow-md transition-shadow`}
    >
      <div className="flex items-center justify-between">
        <div className="font-mono text-sm text-gray-800 break-all">{device.id}</div>
        <div className="flex items-center gap-2">
          <span className={`inline-block h-2.5 w-2.5 rounded-full ${statusColor}`} />
          <span className="text-sm capitalize text-gray-700">{device.status}</span>
        </div>
      </div>
      <div className="mt-2 text-xs text-gray-500">Last seen: {timeAgo(device.last_seen)}</div>
    </button>
  )
}
