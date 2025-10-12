import { useMemo } from 'react'
import HealthCard from './components/HealthCard'

function App() {
  const apiBaseUrl = useMemo(() => import.meta.env.VITE_API_BASE_URL, [])

  return (
    <div className="min-h-screen text-slate-800">
      <header className="border-b bg-white">
        <div className="mx-auto max-w-6xl px-4 py-4 flex items-center justify-between">
          <h1 className="text-xl font-semibold">Fleet Monitoring</h1>
          <span className="text-sm text-slate-500">API: {apiBaseUrl ?? 'not set'}</span>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-4 py-8">
        <div className="space-y-6">
          <HealthCard />

          <div className="rounded-lg border bg-white p-6">
            <h2 className="text-lg font-medium">Welcome</h2>
            <p className="mt-2 text-slate-600">
              Frontend scaffold ready. Configure <code>.env.local</code> with <code>VITE_API_BASE_URL</code>.
            </p>
            <ul className="mt-4 list-disc pl-5 text-slate-600 space-y-1">
              <li><strong>Health:</strong> GET /health (e.g., http://127.0.0.1:6733/health)</li>
              <li><strong>API Base:</strong> {apiBaseUrl || 'http://127.0.0.1:6733/api/v1'}</li>
            </ul>
          </div>
        </div>
      </main>
    </div>
  )
}

export default App
