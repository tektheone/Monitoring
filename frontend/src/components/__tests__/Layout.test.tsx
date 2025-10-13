import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'

// Hoisted mocks
const rq = vi.hoisted(() => ({
  useIsFetching: vi.fn(),
  useQueryClient: vi.fn(),
}))

vi.mock('@tanstack/react-query', async (orig) => {
  const actual = await orig<typeof import('@tanstack/react-query')>()
  return {
    ...actual,
    useIsFetching: rq.useIsFetching,
    useQueryClient: rq.useQueryClient,
    QueryClientProvider: ({ children }: any) => children,
  }
})

const conn = vi.hoisted(() => ({
  useConnection: vi.fn(() => ({ online: true })),
}))
vi.mock('@/hooks/useConnection', () => conn)

vi.mock('@/config', () => ({
  SHOW_FUTURE_NAV: true,
}))

import Layout, { Header, Sidebar, Footer } from '../Layout'

describe('Layout parts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('Header shows title and connection pill (connected)', () => {
    render(<Header title="Dashboard" connection="connected" />)
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    // connection label
    expect(screen.getByText('connected', { exact: false })).toBeInTheDocument()
  })

  it('Sidebar toggles open and calls onNavigate for devices; other links disabled', () => {
    const onNavigate = vi.fn()
    render(<Sidebar current="devices" onNavigate={onNavigate} />)

    // Toggle menu (mobile)
    const toggle = screen.getByRole('button', { name: /menu|hide/i })
    fireEvent.click(toggle)

    // Devices button enabled and clickable
    const devicesBtn = screen.getByRole('button', { name: 'Devices' })
    expect(devicesBtn).not.toBeDisabled()
    fireEvent.click(devicesBtn)
    expect(onNavigate).toHaveBeenCalledWith('devices')

    // Future links present but disabled
    expect(screen.getByRole('button', { name: /analytics \(soon\)/i })).toBeDisabled()
    expect(screen.getByRole('button', { name: /settings \(soon\)/i })).toBeDisabled()
  })

  it('Footer derives status without backendStatus: offline, disconnected, reconnecting, connected', () => {
    // Mock query client to control health state
    const getQueryState = vi.fn()
    rq.useQueryClient.mockReturnValue({ getQueryState })

    // 1) Offline
    conn.useConnection.mockReturnValueOnce({ online: false })
    rq.useIsFetching.mockReturnValue(0)
    getQueryState.mockReturnValueOnce({ status: 'success' })
    const { rerender } = render(<Footer />)
    expect(screen.getByText('Offline')).toBeInTheDocument()

    // 2) Disconnected (health error)
    conn.useConnection.mockReturnValueOnce({ online: true })
    getQueryState.mockReturnValueOnce({ status: 'error' })
    rerender(<Footer />)
    expect(screen.getByText('Disconnected')).toBeInTheDocument()

    // 3) Reconnecting (fetching > 0)
    conn.useConnection.mockReturnValueOnce({ online: true })
    getQueryState.mockReturnValueOnce({ status: 'success' })
    rq.useIsFetching.mockReturnValueOnce(1)
    rerender(<Footer />)
    expect(screen.getByText('Reconnecting')).toBeInTheDocument()

    // 4) Connected (default)
    conn.useConnection.mockReturnValueOnce({ online: true })
    getQueryState.mockReturnValueOnce({ status: 'success' })
    rq.useIsFetching.mockReturnValueOnce(0)
    rerender(<Footer />)
    expect(screen.getByText('Connected')).toBeInTheDocument()
  })

  it('Footer respects explicit backendStatus prop', () => {
    // backendStatus = reconnecting -> label and color
    render(<Footer backendStatus="reconnecting" />)
    expect(screen.getByText('Reconnecting')).toBeInTheDocument()
  })

  it('Sidebar with SHOW_FUTURE_NAV=false renders only Devices link', async () => {
    vi.resetModules()
    // Override config for this import only
    vi.doMock('@/config', () => ({ SHOW_FUTURE_NAV: false }))
    const { Sidebar: SidebarOnly } = await import('../Layout')
    const onNavigate = vi.fn()
    render(<SidebarOnly current="devices" onNavigate={onNavigate} />)
    // Only Devices present; future links absent
    expect(screen.getByRole('button', { name: 'Devices' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /analytics \(soon\)/i })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /settings \(soon\)/i })).not.toBeInTheDocument()
  })

  it('Layout renders Header, Sidebar, children, and Footer (integration)', () => {
    render(
      <Layout header={{ title: 'My App' }} backendStatus="connected">
        <div>Child content</div>
      </Layout>
    )
    expect(screen.getByText('My App')).toBeInTheDocument() // Header
    expect(screen.getByRole('button', { name: 'Devices' })).toBeInTheDocument() // Sidebar
    expect(screen.getByText('Child content')).toBeInTheDocument() // Main
    expect(screen.getByText('Connected')).toBeInTheDocument() // Footer
  })
})
