import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// In dev the API is proxied to the Go server, so the frontend can use
// same-origin paths and avoid CORS. In production VITE_API_BASE points at the
// deployed backend.
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/workers': 'http://localhost:8080',
      '/disbursements': 'http://localhost:8080',
    },
  },
})
