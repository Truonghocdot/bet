import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import './assets/main.css'
import { useAdminAuthStore } from '@/stores/adminAuth'
import { useAuthStore } from '@/stores/auth'

const app = createApp(App)

const pinia = createPinia()
const authStore = useAuthStore(pinia)
const adminAuthStore = useAdminAuthStore(pinia)
authStore.hydrate()
adminAuthStore.hydrate()
if (authStore.accessToken) {
  // Fire-and-forget to refresh user profile silently.
  void authStore.fetchMe()
}
if (adminAuthStore.accessToken) {
  void adminAuthStore.fetchMe()
}

router.beforeEach(async (to) => {
  const requiresAuth = Boolean(to.meta.requiresAuth)
  const requiresAdminAuth = Boolean(to.meta.requiresAdminAuth)
  const isClientAuthPage = to.name === 'auth' || to.name === 'register' || to.name === 'forgot-password'

  if (requiresAuth && !authStore.isAuthenticated) {
    return { path: '/auth', query: { next: to.fullPath } }
  }

  if (requiresAdminAuth && !adminAuthStore.isAuthenticated) {
    return { name: 'auth-sso', query: { expired: '1' } }
  }

  if (isClientAuthPage && authStore.isAuthenticated) {
    const next = typeof to.query.next === 'string' ? to.query.next : '/'
    return next
  }

  return true
})

router.afterEach((to, from) => {
  if (typeof window === 'undefined') return
  if (!from.fullPath || from.fullPath === to.fullPath) return
  window.sessionStorage.setItem('ff789:last-route', from.fullPath === '/' ? '/' : from.fullPath)
})

app.use(pinia)
app.use(router)

app.mount('#app')
