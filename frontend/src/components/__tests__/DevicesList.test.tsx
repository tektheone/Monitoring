import { render, screen, fireEvent, within, waitFor } from '@testing-library/react'
import DevicesList from '../DevicesList'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { DeviceSummary } from '@/types/api'

vi.mock('@/hooks/useSSE', () => ({
  useSSE: () => ({ status: 'disconnected', connected: false }),
}))

vi.mock('@/api/client', async (orig) => {
  const mod = await orig<typeof import('@/api/client')>()
  const now = Date.now()
  const devices: DeviceSummary[] = [
    { id: 'a-1', status: 'offline', last_seen: new Date(now - 30_000).toISOString() }, // will compute as online
    { id: 'b-2', status: 'online', last_seen: new Date(now - 10 * 60 * 1000).toISOString() }, // will compute as offline
  ]
  return {
    ...mod,
    getDevices: vi.fn().mockResolvedValue(devices),
    sendHeartbeat: vi.fn(),
  }
})

const { getDevices, sendHeartbeat } = await import('@/api/client') as any

function renderWithClient(ui: React.ReactElement) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: 0 } },
  })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

test('renders devices and allows search filtering', async () => {
  renderWithClient(<DevicesList />)

  // Wait for list to render
  await screen.findByText('a-1')
  expect(screen.getByText('b-2')).toBeInTheDocument()

  // Search for a-1
  const search = screen.getByLabelText(/search devices by id/i)
  fireEvent.change(search, { target: { value: 'a-' } })

  await waitFor(() => {
    expect(screen.getByText('a-1')).toBeInTheDocument()
    expect(screen.queryByText('b-2')).not.toBeInTheDocument()
  })
})

test('computes online/offline based on last_seen time', async () => {
  renderWithClient(<DevicesList />)
  await screen.findByText('a-1')

  // a-1 should be online (30s ago)
  const aCard = screen.getByText('a-1').closest('button')!
  expect(within(aCard).getByText(/online/i)).toBeInTheDocument()

  // b-2 should be offline (10m ago)
  const bCard = screen.getByText('b-2').closest('button')!
  expect(within(bCard).getByText(/offline/i)).toBeInTheDocument()
})

test('sends heartbeat, shows message, and refetches', async () => {
  // Arrange: make heartbeat resolve after a short tick
  ;(sendHeartbeat as any).mockResolvedValue(undefined)

  renderWithClient(<DevicesList />)
  await screen.findByText('a-1')

  const hbInput = screen.getByLabelText(/device id to send heartbeat/i)
  const hbButton = screen.getByRole('button', { name: /send heartbeat for device id/i })

  // Invalid blank should keep disabled
  expect(hbButton).toBeDisabled()

  // Enter id with spaces to exercise trim and success path
  fireEvent.change(hbInput, { target: { value: '  a-1  ' } })
  expect(hbButton).not.toBeDisabled()
  const initialCalls = (getDevices as any).mock.calls.length

  fireEvent.click(hbButton)

  // Heartbeat called with trimmed id (component passes only the deviceId)
  await waitFor(() => expect(sendHeartbeat).toHaveBeenCalledWith('a-1'))

  // Success message appears and refetch triggers another getDevices call
  await screen.findByText(/heartbeat sent/i)
  await waitFor(() => expect((getDevices as any).mock.calls.length).toBeGreaterThan(initialCalls))
})

test('shows error message when heartbeat fails', async () => {
  // Make heartbeat fail once with a specific message
  ;(sendHeartbeat as any).mockRejectedValueOnce(new Error('Server down'))

  renderWithClient(<DevicesList />)
  await screen.findByText('a-1')

  const hbInput = screen.getByLabelText(/device id to send heartbeat/i)
  const hbButton = screen.getByRole('button', { name: /send heartbeat for device id/i })

  fireEvent.change(hbInput, { target: { value: 'b-2' } })
  fireEvent.click(hbButton)

  // Error message from catch block is shown
  await screen.findByText('Server down')
})

test('filters by status using the Filter select', async () => {
  renderWithClient(<DevicesList />)
  await screen.findByText('a-1')

  const filter = screen.getByLabelText(/filter by status/i)

  // Show only online
  fireEvent.change(filter, { target: { value: 'online' } })
  await waitFor(() => {
    expect(screen.getByText('a-1')).toBeInTheDocument() // online derived
    expect(screen.queryByText('b-2')).not.toBeInTheDocument()
  })

  // Show only offline
  fireEvent.change(filter, { target: { value: 'offline' } })
  await waitFor(() => {
    expect(screen.getByText('b-2')).toBeInTheDocument()
    expect(screen.queryByText('a-1')).not.toBeInTheDocument()
  })
})

test('sorts by last_seen and id', async () => {
  const now = Date.now()
  ;(getDevices as any).mockResolvedValueOnce([
    { id: 'c-3', status: 'offline', last_seen: new Date(now - 5 * 60 * 1000).toISOString() },
    { id: 'a-1', status: 'offline', last_seen: new Date(now - 30_000).toISOString() },
    { id: 'b-2', status: 'online', last_seen: new Date(now - 10 * 60 * 1000).toISOString() },
  ])

  renderWithClient(<DevicesList />)
  await screen.findByText('a-1')

  const sortSelect = screen.getByLabelText(/sort by/i)
  const sortDir = screen.getByLabelText(/toggle sort direction/i)

  // Sort by last_seen ascending: oldest first
  fireEvent.change(sortSelect, { target: { value: 'last_seen' } })
  // Currently asc by default
  // First card should be the oldest (10m ago: b-2), last should be newest (30s ago: a-1)
  const allButtonsAsc = screen.getAllByRole('button').filter((b) => b.textContent?.includes('-'))
  const idsAsc = allButtonsAsc.map((b) => b.textContent).filter(Boolean)
  expect(idsAsc[0]).toContain('b-2')
  expect(idsAsc[idsAsc.length - 1]).toContain('a-1')

  // Toggle to desc: newest first
  fireEvent.click(sortDir)
  const allButtonsDesc = screen.getAllByRole('button').filter((b) => b.textContent?.includes('-'))
  const idsDesc = allButtonsDesc.map((b) => b.textContent).filter(Boolean)
  expect(idsDesc[0]).toContain('a-1')
  expect(idsDesc[idsDesc.length - 1]).toContain('b-2')

  // Sort by id asc
  fireEvent.change(sortSelect, { target: { value: 'id' } })
  // Ensure direction is asc
  if (sortDir.getAttribute('aria-pressed') === 'true') fireEvent.click(sortDir)
  const idsIdAsc = screen.getAllByRole('button').filter((b) => b.textContent?.includes('-')).map((b) => b.textContent)
  expect(idsIdAsc[0]).toContain('a-1')
  expect(idsIdAsc[idsIdAsc.length - 1]).toContain('c-3')
})

test('treats missing last_seen as offline (ts=0 path)', async () => {
  // Device without last_seen should compute as offline
  ;(getDevices as any).mockResolvedValueOnce([
    { id: 'no-last', status: 'online' }, // computeStatus should mark offline
  ])

  renderWithClient(<DevicesList />)
  await screen.findByText('no-last')
  const card = screen.getByText('no-last').closest('button')!
  expect(within(card).getByText(/offline/i)).toBeInTheDocument()
})

test('handles empty device list (base from data ?? [] mapping)', async () => {
  ;(getDevices as any).mockResolvedValueOnce([])
  renderWithClient(<DevicesList />)
  // Component renders controls even with empty list
  await screen.findByText(/last updated/i)
  // No device buttons should be present
  const deviceButtons = screen.queryAllByRole('button').filter((b) => /-\d/.test(b.textContent || ''))
  expect(deviceButtons.length).toBe(0)
})
