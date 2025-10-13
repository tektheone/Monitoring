import { useQuery } from '@tanstack/react-query'
import { getDevices, sendHeartbeat } from '@/api/client'
import type { DeviceSummary } from '@/types/api'
import DeviceCard from './DeviceCard'
import { DEVICES_REFRESH_MS } from '@/config'
import { useSSE } from '@/hooks/useSSE'
import { useDeferredValue, useEffect, useMemo, useState } from 'react'

type Filter = 'all' | 'online' | 'offline'
type SortKey = 'status' | 'last_seen' | 'id'

export default function DevicesList({ onSelect, refreshMs = DEVICES_REFRESH_MS }: { onSelect?: (d: DeviceSummary) => void; refreshMs?: number }) {
  const { connected: sseConnected } = useSSE()
  const { data, refetch, isFetching, dataUpdatedAt } = useQuery<DeviceSummary[]>({
    queryKey: ['devices'],
    queryFn: getDevices,
    refetchInterval: sseConnected ? false : refreshMs,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
    retry: 1,
    placeholderData: (prev) => prev ?? [],
  })

  // UI state: filtering, search, sorting
  const [filter, setFilter] = useState<Filter>('all')
  const [search, setSearch] = useState('')
  const deferredSearch = useDeferredValue(search)
  const [sortKey, setSortKey] = useState<SortKey>('status')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc')
  // Heartbeat controls
  const [hbId, setHbId] = useState('')
  const [hbBusy, setHbBusy] = useState(false)
  const [hbMsg, setHbMsg] = useState<string | null>(null)

  // Tick every 30s to re-evaluate computed online status without refetch
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    const t = setInterval(() => setNow(Date.now()), 30_000)
    return () => clearInterval(t)
  }, [])

  async function handleSendHeartbeat() {
    const id = hbId.trim()
    if (!id) return
    setHbBusy(true)
    setHbMsg(null)
    try {
      await sendHeartbeat(id)
      setHbMsg('Heartbeat sent')
      // Ensure UI updates even if SSE is momentarily disconnected
      refetch()
    } catch (e) {
      setHbMsg((e as Error).message)
    } finally {
      setHbBusy(false)
    }
  }

  const lastUpdated = dataUpdatedAt ? new Date(dataUpdatedAt).toLocaleTimeString() : '—'

  // Derived list per controls (compute status from last_seen relative to now)
  const filteredSorted = useMemo(() => {
    const ONLINE_WINDOW_MS = 2 * 60 * 1000
    const computeStatus = (d: DeviceSummary): 'online' | 'offline' => {
      const ts = d.last_seen ? Date.parse(d.last_seen) : 0
      return ts && now - ts < ONLINE_WINDOW_MS ? 'online' : 'offline'
    }

    // Always derive status from time instead of trusting cached value
    const base = (data ?? []) as DeviceSummary[]
    let list = base.map((d) => ({ ...d, status: computeStatus(d) as DeviceSummary['status'] }))
    if (filter !== 'all') {
      list = list.filter((d) => d.status === filter)
    }
    const q = deferredSearch.trim().toLowerCase()
    if (q) {
      list = list.filter((d) => d.id.toLowerCase().includes(q))
    }
    const dir = sortDir === 'asc' ? 1 : -1
    const getLast = (d: DeviceSummary) => (d.last_seen ? Date.parse(d.last_seen) : 0)
    list = list.slice().sort((a, b) => {
      switch (sortKey) {
        case 'status':
          // online before offline in asc
          const sa = a.status === 'online' ? 0 : 1
          const sb = b.status === 'online' ? 0 : 1
          return (sa - sb) * dir
        case 'last_seen':
          return (getLast(a) - getLast(b)) * dir
        case 'id':
        default:
          return a.id.localeCompare(b.id) * dir
      }
    })
    return list
  }, [data, filter, deferredSearch, sortKey, sortDir, now])

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

      {/* Controls */}
      <div className="mb-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-2" role="group" aria-label="Filter devices">
          <label className="sr-only" htmlFor="device-filter">Filter</label>
          <select
            id="device-filter"
            className="rounded border px-2 py-1 text-sm bg-white"
            value={filter}
            onChange={(e) => setFilter(e.target.value as Filter)}
            aria-label="Filter by status"
          >
            <option value="all">All</option>
            <option value="online">Online</option>
            <option value="offline">Offline</option>
          </select>

          <label className="sr-only" htmlFor="device-sort">Sort</label>
          <select
            id="device-sort"
            className="rounded border px-2 py-1 text-sm bg-white"
            value={sortKey}
            onChange={(e) => setSortKey(e.target.value as SortKey)}
            aria-label="Sort by"
          >
            <option value="status">Status</option>
            <option value="last_seen">Last seen</option>
            <option value="id">Device ID</option>
          </select>
          <button
            type="button"
            className="rounded border px-2 py-1 text-xs hover:bg-gray-50"
            onClick={() => setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'))}
            aria-label="Toggle sort direction"
            aria-pressed={sortDir === 'desc'}
            title={`Sort ${sortDir === 'asc' ? 'ascending' : 'descending'}`}
          >
            {sortDir === 'asc' ? 'Asc' : 'Desc'}
          </button>
        </div>

        <div className="flex items-center gap-2">
          <label className="sr-only" htmlFor="device-search">Search by ID</label>
          <input
            id="device-search"
            type="search"
            inputMode="search"
            placeholder="Search by ID…"
            className="w-full sm:w-64 rounded border px-2 py-1 text-sm bg-white"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            aria-label="Search devices by ID"
          />
          <span className="hidden sm:inline-block w-px h-6 bg-gray-200" aria-hidden="true" />
          <label className="sr-only" htmlFor="hb-id">Device ID</label>
          <input
            id="hb-id"
            type="text"
            placeholder="Device ID to mark online…"
            className="w-full sm:w-64 rounded border px-2 py-1 text-sm bg-white"
            value={hbId}
            onChange={(e) => setHbId(e.target.value)}
            aria-label="Device ID to send heartbeat"
          />
          <button
            type="button"
            onClick={handleSendHeartbeat}
            disabled={!hbId.trim() || hbBusy}
            className="rounded border px-2 py-1 text-xs hover:bg-gray-50 disabled:opacity-50"
            aria-label="Send heartbeat for device ID"
          >
            {hbBusy ? 'Sending…' : 'Send Heartbeat'}
          </button>
          {hbMsg && <span className="text-xs text-gray-500" role="status">{hbMsg}</span>}
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        {filteredSorted.map((d) => (
          <div key={d.id} className="transition-transform duration-200 ease-out hover:scale-[1.01]">
            <DeviceCard device={d} onClick={onSelect} />
          </div>
        ))}
      </div>
    </div>
  )
}
