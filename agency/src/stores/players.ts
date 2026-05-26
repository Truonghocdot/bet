import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { request, type ApiError } from '@/shared/api/http'
import type { ManagedAffiliateUser } from '@/shared/api/types'
import { useAgencyAuthStore } from '@/stores/auth'

function statusLabel(status: number): string {
  if (status === 2) return 'Đã nạp đầu'
  if (status === 1) return 'Chờ nạp đầu'
  if (status === 3) return 'Không hợp lệ'
  return 'Không rõ'
}

export const useAgencyPlayersStore = defineStore('agency-players', () => {
  const items = ref<ManagedAffiliateUser[]>([])
  const invitedUsersCount = ref(0)
  const loading = ref(false)
  const error = ref('')

  const totalPlayers = computed(() => items.value.length)
  const headlinePlayers = computed(() => Math.max(invitedUsersCount.value, items.value.length))
  const qualifiedPlayers = computed(() => items.value.filter((item) => item.referral_status === 2).length)
  const waitingPlayers = computed(() => items.value.filter((item) => item.referral_status === 1).length)
  const invalidPlayers = computed(() => items.value.filter((item) => item.referral_status === 3).length)
  const totalFirstDeposit = computed(() => items.value.reduce((sum, item) => sum + Number(item.first_deposit_amount || 0), 0))
  const conversionRate = computed(() => {
    const base = headlinePlayers.value
    if (!base) return 0
    return Math.round((qualifiedPlayers.value / base) * 100)
  })
  const averageFirstDeposit = computed(() => {
    if (!qualifiedPlayers.value) return 0
    return Math.round(totalFirstDeposit.value / qualifiedPlayers.value)
  })
  const recentPlayers = computed(() => items.value.slice(0, 5))
  const priorityPlayers = computed(() =>
    items.value
      .filter((item) => item.referral_status === 1 || item.referral_status === 3)
      .slice(0, 6),
  )

  function findUserById(userID: number) {
    return items.value.find((item) => item.user_id === userID) ?? null
  }

  async function fetchSummary() {
    const auth = useAgencyAuthStore()
    if (!auth.accessToken) return

    try {
      const res = await request<{ invited_users_count: number }>('GET', '/v1/affiliate/summary', {
        token: auth.accessToken,
      })
      invitedUsersCount.value = Number(res.invited_users_count ?? 0)
    } catch (e: any) {
      invitedUsersCount.value = items.value.length
      throw e
    }
  }

  async function fetchManagedUsers() {
    const auth = useAgencyAuthStore()
    if (!auth.accessToken) return

    loading.value = true
    error.value = ''
    try {
      const res = await request<{ items: ManagedAffiliateUser[] }>('GET', '/v1/affiliate/managed-users', {
        token: auth.accessToken,
      })
      items.value = res.items ?? []
    } catch (e: any) {
      error.value = (e as ApiError)?.message ?? 'Không thể tải danh sách người chơi'
      items.value = []
      throw e
    } finally {
      loading.value = false
    }
  }

  async function refreshDashboard() {
    loading.value = true
    error.value = ''
    try {
      await Promise.all([fetchSummary(), fetchManagedUsers()])
    } catch (e: any) {
      error.value = (e as ApiError)?.message ?? 'Không thể tải dữ liệu agency'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function ensureLoaded() {
    if (items.value.length > 0) return
    await refreshDashboard()
  }

  return {
    items,
    invitedUsersCount,
    loading,
    error,
    totalPlayers,
    headlinePlayers,
    qualifiedPlayers,
    waitingPlayers,
    invalidPlayers,
    totalFirstDeposit,
    conversionRate,
    averageFirstDeposit,
    recentPlayers,
    priorityPlayers,
    findUserById,
    fetchSummary,
    fetchManagedUsers,
    refreshDashboard,
    ensureLoaded,
    statusLabel,
  }
})
