import { createRouter, createWebHistory } from 'vue-router'

import { useAgencyAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/overview',
    },
    {
      path: '/login',
      name: 'agency-login',
      component: () => import('@/views/AgencyLoginView.vue'),
      meta: { title: 'Đăng nhập Agency', layout: 'plain' },
    },
    {
      path: '/overview',
      name: 'agency-overview',
      component: () => import('@/views/AgencyOverviewView.vue'),
      meta: { requiresAuth: true, title: 'Tổng quan Agency', layout: 'panel' },
    },
    {
      path: '/users',
      name: 'agency-users',
      component: () => import('@/views/AgencyUsersView.vue'),
      meta: { requiresAuth: true, title: 'Người chơi', layout: 'panel' },
    },
    {
      path: '/users/:userId',
      name: 'agency-user-stats',
      component: () => import('@/views/AgencyUserStatsView.vue'),
      meta: { requiresAuth: true, title: 'Thống kê người chơi', layout: 'panel' },
    },
    {
      path: '/deposits',
      name: 'agency-deposits',
      component: () => import('@/views/AgencyUserDepositsView.vue'),
      meta: { requiresAuth: true, title: 'Giao dịch nạp tiền', layout: 'panel' },
    },
    {
      path: '/withdrawals',
      name: 'agency-withdrawals',
      component: () => import('@/views/AgencyUserWithdrawalsView.vue'),
      meta: { requiresAuth: true, title: 'Giao dịch rút tiền', layout: 'panel' },
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAgencyAuthStore()

  if (to.meta.title) {
    document.title = `${to.meta.title} - FH88U Agency`
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'agency-login' }
  }

  if (to.name === 'agency-login' && auth.isAuthenticated) {
    return { name: 'agency-overview' }
  }

  return true
})

export default router
