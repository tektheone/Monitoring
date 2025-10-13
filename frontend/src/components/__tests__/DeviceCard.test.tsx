import { render, screen, fireEvent } from '@testing-library/react'
import DeviceCard from '../DeviceCard'
import type { DeviceSummary } from '@/types/api'

function makeDevice(partial: Partial<DeviceSummary> = {}): DeviceSummary {
  return {
    id: 'device-1',
    status: 'online',
    last_seen: new Date().toISOString(),
    ...partial,
  }
}

test('renders device id and status', () => {
  render(<DeviceCard device={makeDevice()} />)
  expect(screen.getByText('device-1')).toBeInTheDocument()
  expect(screen.getByText(/online/i)).toBeInTheDocument()
})

test('shows offline style when status is offline (derived prop)', () => {
  render(<DeviceCard device={makeDevice({ status: 'offline' })} />)
  expect(screen.getByText(/offline/i)).toBeInTheDocument()
})

test('invokes onClick with the device when clicked', () => {
  const device = makeDevice({ id: 'device-click' })
  const onClick = vi.fn()
  render(<DeviceCard device={device} onClick={onClick} />)
  const btn = screen.getByRole('button')
  fireEvent.click(btn)
  expect(onClick).toHaveBeenCalledTimes(1)
  expect(onClick).toHaveBeenCalledWith(device)
})
