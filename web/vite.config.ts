import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { VitePWA } from 'vite-plugin-pwa';

// Backend origin used by the dev-server proxy. In production the app is served
// behind nginx, which proxies `/api` to the backend (see web/nginx.conf), so no
// proxy is needed there.
const API_TARGET = process.env.VITE_DEV_API_TARGET ?? 'http://localhost:8080';

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      // We hand-write the service worker (src/sw.ts) so we can handle Web Push
      // `push` / `notificationclick` events in addition to Workbox precaching.
      strategies: 'injectManifest',
      srcDir: 'src',
      filename: 'sw.ts',
      registerType: 'autoUpdate',
      injectRegister: null, // registration is done manually in src/lib/pwa.ts
      devOptions: {
        enabled: false,
      },
      manifest: {
        name: 'konatsu',
        short_name: 'konatsu',
        description: 'AI-driven rich web mail client',
        display: 'standalone',
        start_url: '/',
        scope: '/',
        theme_color: '#ffd20a',
        background_color: '#ffffff',
        lang: 'ja',
        icons: [
          { src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' },
          { src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png' },
          {
            src: '/icons/icon-maskable-512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'maskable',
          },
        ],
      },
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: API_TARGET,
        changeOrigin: true,
        ws: true,
      },
      '/healthz': { target: API_TARGET, changeOrigin: true },
    },
  },
});
