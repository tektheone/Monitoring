export function timeAgo(iso?: string): string {
  if (!iso) return '—'
  const date = new Date(iso)
  if (isNaN(date.getTime())) return 'Invalid date'
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const sec = Math.floor(diffMs / 1000)
  if (sec < 60) return `${sec}s ago`
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m ago`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr}h ago`
  const days = Math.floor(hr / 24)
  return `${days}d ago`
}
