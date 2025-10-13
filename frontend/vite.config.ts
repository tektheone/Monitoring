import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import { fileURLToPath, URL } from 'node:url'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['src/test/setup.ts'],
    css: true,
    coverage: {
      enabled: true,
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      all: true,
      include: ['src/**/*.{ts,tsx}'],
      exclude: [
        'src/test/**',
        '**/__tests__/**',
        '**/*.test.*',
        '**/*.spec.*',
        '**/*.d.ts',
        'src/config.ts',
        'src/main.tsx',
      ],
    },
  },
  server: {
    port: 5173,
    host: '127.0.0.1',
    proxy: {
      // Proxy API to Go backend
      '/api/v1': {
        target: 'http://127.0.0.1:6733',
        changeOrigin: true,
      },
      '/health': {
        target: 'http://127.0.0.1:6733',
        changeOrigin: true,
      },
      // Proxy SSE endpoint to backend
      '/events': {
        target: 'http://127.0.0.1:6733',
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
})
