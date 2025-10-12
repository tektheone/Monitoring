import { useHealth } from '../hooks/useHealth'

function StatusBadge({ health }: { health: 'healthy' | 'degraded' | 'down' }) {
  const map = {
    healthy: {
      text: 'Healthy',
      className: 'bg-green-100 text-green-800 ring-green-200',
    },
    degraded: {
      text: 'Degraded',
      className: 'bg-yellow-100 text-yellow-800 ring-yellow-200',
    },
    down: {
      text: 'Down',
      className: 'bg-red-100 text-red-800 ring-red-200',
    },
  } as const
  const cfg = map[health]
  return (
    <span className={`inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset ${cfg.className}`}>
      {cfg.text}
    </span>
  )
}

function formatTime(ts: number | null) {
  if (!ts) return '—'
  const d = new Date(ts)
  return d.toLocaleTimeString()
}

export default function HealthCard() {
  const { loading, error, data, health, lastUpdated, statusCode, refetch } = useHealth(10_000)

  return (
    <div className="rounded-lg border bg-white p-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-medium">Cluster Health</h2>
        <div className="flex items-center gap-3">
          <StatusBadge health={health} />
          <button
            onClick={() => refetch()}
            className="rounded-md bg-slate-800 px-3 py-1.5 text-xs font-medium text-white hover:bg-slate-700 disabled:opacity-50"
            disabled={loading}
          >
            {loading ? 'Refreshing…' : 'Refresh'}
          </button>
        </div>
      </div>

      {error && (
        <div className="mt-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800">
          {error}
        </div>
      )}

      <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div className="rounded-md border p-4">
          <div className="text-xs text-slate-500">Status Code</div>
          <div className="mt-1 text-base text-slate-900">{loading ? '—' : (statusCode ?? '—')}</div>
        </div>
        <div className="rounded-md border p-4">
          <div className="text-xs text-slate-500">Device Count</div>
          <div className="mt-1 text-base text-slate-900">{loading ? '—' : (data?.device_count ?? '—')}</div>
        </div>
        <div className="rounded-md border p-4">
          <div className="text-xs text-slate-500">Last Updated</div>
          <div className="mt-1 text-base text-slate-900">{formatTime(lastUpdated)}</div>
        </div>
      </div>
    </div>
  )
}
