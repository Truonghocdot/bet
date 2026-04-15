import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/home' },
  {
    path: '/auth',
    name: 'auth',
    component: () => import('../pages/LoginView.vue'),
    meta: { layout: 'auth', title: 'Đăng nhập' },
  },
  {
    path: '/login',
    redirect: '/auth',
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../pages/RegisterView.vue'),
    meta: { layout: 'auth', title: 'Đăng ký' },
  },
  {
    path: '/auth/sso',
    name: 'auth-sso',
    component: () => import('../pages/SSOView.vue'),
    meta: { layout: 'auth', title: 'Xác thực SSO' },
  },
  {
    path: '/admin/control',
    name: 'admin-control',
    component: () => import('../pages/AdminControlView.vue'),
    meta: { layout: 'main', title: 'Điều khiển kết quả', requiresAuth: true },
  },
  {
    path: '/forgot-password',
    name: 'forgot-password',
    component: () => import('../pages/ForgotPasswordView.vue'),
    meta: { layout: 'auth', title: 'Quên mật khẩu' },
  },
  {
    path: '/home',
    name: 'home',
    component: () => import('../pages/HomeView.vue'),
    meta: { layout: 'main', title: 'Trang chủ' },
  },
  {
    path: '/promotion',
    name: 'promotion',
    component: () => import('../pages/PromotionView.vue'),
    meta: { layout: 'main', title: 'Hoạt động', requiresAuth: true },
  },
  {
    path: '/notifications',
    name: 'notifications',
    component: () => import('../pages/NotificationsView.vue'),
    meta: { layout: 'main', title: 'Thông báo', requiresAuth: true },
  },
  {
    path: '/news/:slug',
    name: 'news-detail',
    component: () => import('../pages/NewsDetailView.vue'),
    meta: { layout: 'main', title: 'Tin tức' },
  },
  {
    path: '/deposit',
    name: 'deposit',
    component: () => import('../pages/DepositView.vue'),
    meta: { layout: 'main', title: 'Nạp tiền', requiresAuth: true },
  },
  {
    path: '/withdraw',
    name: 'withdraw',
    component: () => import('../pages/WithdrawView.vue'),
    meta: { layout: 'main', title: 'Rút tiền', requiresAuth: true },
  },
  {
    path: '/exchange',
    name: 'exchange',
    component: () => import('../pages/ExchangeView.vue'),
    meta: { layout: 'main', title: 'Chuyển đổi ví', requiresAuth: true },
  },
  {
    path: '/account',
    name: 'account',
    component: () => import('../pages/AccountView.vue'),
    meta: { layout: 'main', title: 'Cá nhân', requiresAuth: true },
  },
  {
    path: '/game-stats',
    name: 'game-stats',
    component: () => import('../pages/GameStatsView.vue'),
    meta: { layout: 'main', title: 'Thống kê trò chơi', requiresAuth: true },
  },
  {
    path: '/play',
    name: 'play-lobby',
    component: () => import('../pages/PlayLobbyView.vue'),
    meta: { layout: 'main', title: 'Phòng chơi', requiresAuth: true },
  },
  {
    path: '/play/:game',
    name: 'play',
    component: () => import('../pages/PlayView.vue'),
    meta: { layout: 'main', title: 'Phòng chơi', requiresAuth: true },
  },
  {
    path: '/cskh',
    name: 'cskh',
    component: () => import('../pages/CSKHView.vue'),
    meta: { layout: 'main', title: 'Hỗ Trợ' },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('../pages/NotFoundView.vue'),
    meta: { layout: 'main', title: '404' },
  },
]

import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()
  const isAuthenticated = !!auth.accessToken
  const userRole = auth.user?.role

  // Title update
  if (to.meta.title) {
    document.title = `${to.meta.title} - ff789`
  }

  // Auth check
  if (to.meta.requiresAuth && !isAuthenticated) {
    return next({ name: 'auth' })
  }

  // Admin role check for specific route
  if (to.name === 'admin-control' && userRole !== 1) {
    return next({ name: 'home' })
  }

  next()
})

export default router
