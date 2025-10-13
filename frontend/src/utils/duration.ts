// Format Go duration strings like "250ms", "1h2m3s" into spaced human form
export function formatDuration(input?: string): string {
  if (!input) return '—'
  // Split into numeric+unit tokens, e.g., ["1h","2m","3s","250ms"]
  const tokens = input.match(/\d+\s*(ns|us|µs|ms|s|m|h)/g)
  if (!tokens) return input
  const unitMap: Record<string, string> = {
    ns: 'ns',
    us: 'µs',
    'µs': 'µs',
    ms: 'ms',
    s: 's',
    m: 'm',
    h: 'h',
  }
  return tokens
    .map((t) => {
      const m = t.match(/(\d+)\s*(ns|us|µs|ms|s|m|h)/)
      if (!m) return t
      const [, num, unit] = m
      return `${num} ${unitMap[unit] ?? unit}`
    })
    .join(' ')
}
