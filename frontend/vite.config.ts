import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      // Proxy API and health checks to backend server to avoid CORS in dev
      '/api': {
        target: 'http://127.0.0.1:6733',
        changeOrigin: false,
      },
      '/health': {
        target: 'http://127.0.0.1:6733',
        changeOrigin: false,
      },
    },
  }
})
