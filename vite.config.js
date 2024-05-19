import path from 'path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ command, mode }) => ({
  plugins: [vue()],
  base: '/static/',
  server: {
    port: 3000
  },
  resolve: {
    alias: [{ find: '@', replacement: path.resolve('./assets/js') }]
  },
  build: {
    manifest: 'manifest.json',
    outDir: 'static/assets',
    assetsDir: '',
    minify: mode === 'development' ? false : 'esbuild',
    rollupOptions: {
      input: {
        app: path.resolve('assets/js/app.js'),
        style: path.resolve('assets/css/style.css')
      }
    }
  }
}))
