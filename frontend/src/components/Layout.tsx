import { ReactNode, useState } from 'react'
import { useIsFetching, useQueryClient } from '@tanstack/react-query'
import { useConnection } from '@/hooks/useConnection'
import { SHOW_FUTURE_NAV } from '@/config'

export function Header({ title, connection }: { title?: string; connection?: 'connected' | 'reconnecting' | 'disconnected' | 'offline' }) {
  const connColor = connection === 'connected' ? 'bg-green-500' : connection === 'reconnecting' ? 'bg-yellow-500' : connection ? 'bg-red-500' : 'bg-gray-300'
  const logoUrl = new URL('../../logo.svg', import.meta.url).href
  return (
    <header className="sticky top-0 z-40 border-b text-white">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
        <div className="flex items-center gap-3 w-full">
          <img src={logoUrl} alt="Company logo" className="rounded object-cover" />
          {title && <h1 className="text-lg font-semibold">{title}</h1>}
        </div>
        <div className="flex items-center gap-4 text-sm text-white/90">
          <div className="hidden sm:flex items-center gap-2">
            <span className={`inline-block h-2.5 w-2.5 rounded-full ${connColor}`} />
            <span className="capitalize text-black">{connection ?? 'unknown'}</span>
          </div>
        </div>
      </div>
    </header>
  )
}

export function Sidebar({ current, onNavigate }: { current: string; onNavigate: (key: string) => void }) {
  const [open, setOpen] = useState(false)
  const links = SHOW_FUTURE_NAV
    ? [
        { key: 'devices', label: 'Devices' },
        { key: 'analytics', label: 'Analytics (soon)' },
        { key: 'settings', label: 'Settings (soon)' },
      ]
    : [
        { key: 'devices', label: 'Devices' },
      ]
  return (
    <aside className="border-r bg-white">
      <div className="sm:hidden p-2">
        <button className="rounded border px-2 py-1 text-sm" onClick={() => setOpen((v) => !v)}>
          {open ? 'Hide' : 'Menu'}
        </button>
      </div>
      <nav className={`p-3 space-y-1 ${open ? 'block' : 'hidden sm:block'}`}>
        {links.map((l) => (
          <button
            key={l.key}
            onClick={() => onNavigate(l.key)}
            disabled={l.key !== 'devices'}
            className={`block w-full rounded px-3 py-2 text-left text-sm ${
              current === l.key ? 'bg-gray-100 font-medium' : 'hover:bg-gray-50'
            } ${l.key !== 'devices' ? 'opacity-50 cursor-not-allowed' : ''}`}
          >
            {l.label}
          </button>
        ))}
      </nav>
    </aside>
  )
}

export function Footer({ backendStatus }: { backendStatus?: 'connected' | 'reconnecting' | 'disconnected' | 'offline' }) {
  const { online } = useConnection()
  const qc = useQueryClient()
  const healthFetching = useIsFetching({ queryKey: ['health'] })
  const healthState = qc.getQueryState(['health']) as { status?: 'pending' | 'error' | 'success' } | undefined

  // Derive backend connection status
  // - If browser offline: Offline (red)
  // - Else if health query in error: Disconnected (red)
  // - Else if fetching: Reconnecting (yellow)
  // - Else: Connected (green)
  let label = 'Connected'
  let color = 'bg-green-500'
  if (backendStatus) {
    label = backendStatus === 'connected' ? 'Connected' : backendStatus === 'reconnecting' ? 'Reconnecting' : backendStatus === 'offline' ? 'Offline' : 'Disconnected'
    color = backendStatus === 'connected' ? 'bg-green-500' : backendStatus === 'reconnecting' ? 'bg-yellow-500' : 'bg-red-500'
  } else {
    if (!online) {
      label = 'Offline'
      color = 'bg-red-500'
    } else if (healthState?.status === 'error') {
      label = 'Disconnected'
      color = 'bg-red-500'
    } else if (healthFetching > 0) {
      label = 'Reconnecting'
      color = 'bg-yellow-500'
    }
  }

  return (
    <footer className="mt-auto border-t bg-white">
      <div className="mx-auto max-w-6xl px-4 py-2 text-sm text-gray-700 flex items-center gap-2">
        <span className={`inline-block h-2.5 w-2.5 rounded-full ${color}`} />
        <span>{label}</span>
      </div>
    </footer>
  )
}

export default function Layout({ header, children, backendStatus }: { header: Parameters<typeof Header>[0]; backendStatus?: 'connected' | 'reconnecting' | 'disconnected' | 'offline'; children: ReactNode }) {
  return (
    <div className="min-h-screen bg-gray-50 text-gray-900 flex flex-col">
      <Header {...header} connection={backendStatus} />
      <div className="mx-auto grid w-full max-w-6xl grid-cols-1 sm:grid-cols-[220px_1fr] gap-4 px-4 py-4">
        <Sidebar current="devices" onNavigate={() => {}} />
        <main>{children}</main>
      </div>
      <Footer backendStatus={backendStatus} />
    </div>
  )
}
