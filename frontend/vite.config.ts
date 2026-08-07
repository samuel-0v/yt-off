import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const apiProxyTarget = process.env.VITE_API_PROXY_TARGET ?? 'http://localhost:18080'

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/health': {
        target: apiProxyTarget,
        changeOrigin: true,
        configure: configureForwardedHost,
      },
      '/api': {
        target: apiProxyTarget,
        changeOrigin: true,
        configure: configureForwardedHost,
      },
    },
  },
})

function configureForwardedHost(proxy: { on: (event: string, callback: (proxyReq: { setHeader: (name: string, value: string) => void }, req: { headers: { host?: string } }) => void) => void }) {
  proxy.on('proxyReq', (proxyReq, req) => {
    if (req.headers.host) {
      proxyReq.setHeader('X-Forwarded-Host', req.headers.host)
    }
  })
}
