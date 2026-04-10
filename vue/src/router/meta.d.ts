import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    layout?: 'auth' | 'main'
    title?: string
    requiresAuth?: boolean
  }
}

