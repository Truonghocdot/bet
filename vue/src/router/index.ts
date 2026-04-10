import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/', redirect: '/home' },
  {
    path: '/auth',
    name: 'auth',
    component: () => import('../pages/AuthView.vue'),
    meta: { layout: 'auth', title: 'Đăng nhập' },
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
    meta: { layout: 'main', title: 'Khuyến mãi' },
  },
  {
    path: '/account',
    name: 'account',
    component: () => import('../pages/AccountView.vue'),
    meta: { layout: 'main', title: 'Cá nhân' },
  },
  {
    path: '/play/:game',
    name: 'play',
    component: () => import('../pages/PlayView.vue'),
    meta: { layout: 'main', title: 'Phòng chơi' },
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
