import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    layout?: 'auth' | 'main' | 'admin'
    title?: string
    requiresAuth?: boolean
    requiresAdminAuth?: boolean
  }
}
