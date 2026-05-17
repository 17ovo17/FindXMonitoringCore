import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { copyFileSync, existsSync, mkdirSync, realpathSync } from 'node:fs'
import { resolve } from 'node:path'

const apiTarget = process.env.VITE_API_PROXY || 'http://localhost:8080'
const receiverProxy = { target: apiTarget, changeOrigin: true }
const projectRoot = realpathSync(process.cwd())
const spaRouteCopies = () => ({
  name: 'findx-spa-route-copies',
  writeBundle() {
    const dist = resolve(projectRoot, 'dist')
    const index = resolve(dist, 'index.html')
    if (!existsSync(index)) return
    const assetsRoute = resolve(dist, 'assets')
    mkdirSync(assetsRoute, { recursive: true })
    copyFileSync(index, resolve(assetsRoute, 'index.html'))
  }
})

export default defineConfig({
  root: projectRoot,
  plugins: [react(), spaRouteCopies()],
  build: {
    assetsDir: 'static'
  },
  server: {
    port: 3000,
    proxy: {
      '/api': { target: apiTarget, changeOrigin: true, ws: true },
      '/findx-agent': receiverProxy,
      '/v1/n9e': receiverProxy,
      '/prometheus': receiverProxy
    }
  }
})
