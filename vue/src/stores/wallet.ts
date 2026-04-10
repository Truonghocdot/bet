import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { request, type ApiError } from '@/shared/api/http'
import type { WalletSummaryResponse } from '@/shared/api/types'
import { useAuthStore } from '@/stores/auth'

export const useWalletStore = defineStore('wallet', () => {
  const summary = ref<WalletSummaryResponse | null>(null)
  const loading = ref(false)
  const error = ref('')

  const wallets = computed(() => summary.value?.wallets ?? [])

  async function fetchSummary() {
    const auth = useAuthStore()

    if (!auth.accessToken) {
      summary.value = null
      return null
    }

    loading.value = true
    error.value = ''
    try {
      const res = await request<WalletSummaryResponse>('GET', '/v1/wallets/summary', {
        token: auth.accessToken,
      })
      summary.value = res
      return res
    } catch (e: any) {
      const err = e as ApiError
      if (err?.status === 401) {
        auth.logout()
        summary.value = null
        throw e
      }
      error.value = err?.message ?? 'Không thể tải ví'
      throw e
    } finally {
      loading.value = false
    }
  }

  function reset() {
    summary.value = null
    loading.value = false
    error.value = ''
  }

  return {
    summary,
    wallets,
    loading,
    error,
    fetchSummary,
    reset,
  }
})
