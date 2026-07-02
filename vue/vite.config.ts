import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

const tunnelHost = 'ebook-com-qualifications-attempting.trycloudflare.com'
const defaultAllowedHosts = ['fh88u.win', 'www.fh88u.win', 'localhost', '127.0.0.1']
const extraAllowedHosts = (process.env.VITE_ALLOWED_HOSTS ?? '')
  .split(',')
  .map((host) => host.trim())
  .filter(Boolean)

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const enableVueDevTools = process.env.VITE_ENABLE_DEVTOOLS !== 'false' && mode !== 'production'

  return {
    server: {
      host: true,
      allowedHosts: [
        ...defaultAllowedHosts,
        tunnelHost,
        ...extraAllowedHosts,
      ],
    },
    plugins: [
      tailwindcss(),
      vue(),
      ...(enableVueDevTools ? [vueDevTools()] : []),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      },
    },
  }
})
