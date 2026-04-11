import { ref } from 'vue'
import { defineStore } from 'pinia'

import { request, type ApiError } from '@/shared/api/http'
import { connectEventStream, type StreamConnection } from '@/shared/api/stream'
import type { DepositInitResponse, DepositStatusResponse } from '@/shared/api/types'
import { useAuthStore } from '@/stores/auth'

export const useDepositStore = defineStore('deposit', () => {
  const currentIntent = ref<DepositInitResponse | null>(null)
  const currentStatus = ref<DepositStatusResponse | null>(null)
  const loading = ref(false)
  const error = ref<string>('')
  let streamConnection: StreamConnection | null = null

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

  function connectStatusStream(clientRef: string) {
    const auth = useAuthStore()
    if (!auth.accessToken || !clientRef) return

    disconnectStatusStream()
    streamConnection = connectEventStream(`/v1/deposits/${encodeURIComponent(clientRef)}/stream`, {
      token: auth.accessToken,
      reconnectMs: 3000,
      onEvent(payload) {
        if (payload.event !== 'deposit.status') return
        currentStatus.value = payload.data as DepositStatusResponse
        const status = currentStatus.value?.transaction?.status
        if (status === 2 || status === 3 || status === 4) {
          disconnectStatusStream()
        }
      },
      onError(errorValue) {
        const err = errorValue as ApiError
        if (err?.status === 401) {
          auth.logout()
          reset()
          return
        }
        error.value = err?.message ?? 'Kết nối trạng thái nạp tiền bị gián đoạn'
      },
    })
  }

  function disconnectStatusStream() {
    streamConnection?.close()
    streamConnection = null
  }

  function reset() {
    disconnectStatusStream()
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
    connectStatusStream,
    disconnectStatusStream,
    reset,
  }
})
