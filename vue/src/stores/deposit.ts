import { ref } from 'vue'
import { defineStore } from 'pinia'

import { request, type ApiError } from '@/shared/api/http'
import type { DepositInitResponse, DepositStatusResponse } from '@/shared/api/types'
import { useAuthStore } from '@/stores/auth'

export const useDepositStore = defineStore('deposit', () => {
  const currentIntent = ref<DepositInitResponse | null>(null)
  const currentStatus = ref<DepositStatusResponse | null>(null)
  const loading = ref(false)
  const error = ref<string>('')

  async function initVietQR(payload: { amount: string; note?: string }) {
    const auth = useAuthStore()
    loading.value = true
    error.value = ''
    try {
      const res = await request<DepositInitResponse>('POST', '/v1/deposits/vietqr/init', {
        token: auth.accessToken,
        body: payload,
      })
      currentIntent.value = res
      currentStatus.value = null
      return res
    } catch (e: any) {
      const err = e as ApiError
      error.value = err?.message ?? 'Không thể tạo giao dịch nạp'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function initUSDT(payload: { amount: string; note?: string }) {
    const auth = useAuthStore()
    loading.value = true
    error.value = ''
    try {
      const res = await request<DepositInitResponse>('POST', '/v1/deposits/usdt/init', {
        token: auth.accessToken,
        body: payload,
      })
      currentIntent.value = res
      currentStatus.value = null
      return res
    } catch (e: any) {
      const err = e as ApiError
      error.value = err?.message ?? 'Không thể tạo giao dịch nạp'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function getStatus(clientRef: string) {
    const auth = useAuthStore()
    try {
      const res = await request<DepositStatusResponse>('GET', `/v1/deposits/${encodeURIComponent(clientRef)}`, {
        token: auth.accessToken,
      })
      currentStatus.value = res
      return res
    } catch (e: any) {
      const err = e as ApiError
      if (err?.status === 401) {
        auth.logout()
        throw e
      }
      error.value = err?.message ?? 'Không thể lấy trạng thái nạp'
      throw e
    }
  }

  function reset() {
    currentIntent.value = null
    currentStatus.value = null
    loading.value = false
    error.value = ''
  }

  return {
    currentIntent,
    currentStatus,
    loading,
    error,
    initVietQR,
    initUSDT,
    getStatus,
    reset,
  }
})
