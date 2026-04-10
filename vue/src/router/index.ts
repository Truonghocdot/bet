import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/home' },
  { path: '/play', redirect: '/play' },
  {
    path: '/auth',
    name: 'auth',
    component: () => import('../pages/LoginView.vue'),
    meta: { layout: 'auth', title: 'Đăng nhập' },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../pages/RegisterView.vue'),
    meta: { layout: 'auth', title: 'Đăng ký' },
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
    meta: { layout: 'main', title: 'Hoạt động' },
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
    path: '/account',
    name: 'account',
    component: () => import('../pages/AccountView.vue'),
    meta: { layout: 'main', title: 'Cá nhân', requiresAuth: true },
  },
  {
    path: '/play/:game',
    name: 'play',
    component: () => import('../pages/PlayView.vue'),
    meta: { layout: 'main', title: 'Phòng chơi', requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
