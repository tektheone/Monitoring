# Frontend

## Setup

- Node 18+
- Install deps:

```bash
npm install
```

- Run dev server:

```bash
npm run dev
```

The Vite dev server proxies API/SSE to the Go backend at `http://127.0.0.1:6733` (see `vite.config.ts`). Start the backend separately.

## Testing

Uses Vitest + React Testing Library + jsdom.

- Run tests once:

```bash
npm test
```

- Watch mode:

```bash
npm run test:watch
```

- Coverage:

```bash
npm run coverage
```

### What’s covered

- `DevicesList`: filtering, search, and time-based online/offline derivation.
- `DeviceCard`: renders ID and status.
- (Add more as needed.)

## Error/Offline scenarios

- When backend is down, the header connection chip shows `Disconnected/Offline` based on SSE and browser online state. Health card has been removed.
- Device online/offline is computed from `last_seen` in the last 2 minutes; the list re-evaluates every 30 seconds without requiring a refetch.

## Configuration

- Feature flags in `src/config.ts`:
  - `SHOW_FUTURE_NAV`: toggles placeholder nav items.

## Notes

- Logo is loaded in `components/Layout.tsx` via `new URL('../../logo.svg', import.meta.url).href`.
- TailwindCSS is configured in `src/styles.css`; tweak UI classes directly in components.
