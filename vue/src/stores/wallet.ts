import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { request, type ApiError } from '@/shared/api/http'
import { connectEventStream, type StreamConnection } from '@/shared/api/stream'
import type { WalletSummaryResponse, ExchangeRequest, ExchangeResponse } from '@/shared/api/types'
import { useAuthStore } from '@/stores/auth'

export const useWalletStore = defineStore('wallet', () => {
  const summary = ref<WalletSummaryResponse | null>(null)
  const loading = ref(false)
  const error = ref('')
  let streamConnection: StreamConnection | null = null

  const wallets = computed(() => summary.value?.wallets ?? [])

  async function fetchSummary() {
    const auth = useAuthStore()

    loading.value = true
    error.value = ''
    try {
      const res = await request<WalletSummaryResponse>('GET', '/v1/wallets/summary', {
        token: auth.accessToken || undefined,
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
    disconnectStream()
    summary.value = null
    loading.value = false
    error.value = ''
  }

  function connectStream() {
    const auth = useAuthStore()
    if (!auth.accessToken || streamConnection) return

    streamConnection = connectEventStream('/v1/wallets/stream', {
      token: auth.accessToken,
      reconnectMs: 3000,
      onEvent(payload) {
        if (payload.event !== 'wallet.summary') return
        summary.value = payload.data as WalletSummaryResponse
      },
      onError(errorValue) {
        const err = errorValue as ApiError
        if (err?.status === 401) {
          auth.logout()
          reset()
          return
        }
        error.value = err?.message ?? 'Kết nối ví realtime bị gián đoạn'
      },
    })
  }

  function disconnectStream() {
    streamConnection?.close()
    streamConnection = null
  }

  async function exchangeWallets(payload: ExchangeRequest): Promise<ExchangeResponse> {
    const auth = useAuthStore()
    if (!auth.accessToken) throw new Error('Unauthorized')

    loading.value = true
    try {
      const res = await request<ExchangeResponse>('POST', '/v1/wallets/exchange', {
        token: auth.accessToken,
        body: payload,
      })
      await fetchSummary()
      return res
    } catch (e: any) {
      const err = e as ApiError
      if (err?.status === 401) {
        auth.logout()
      }
      throw e
    } finally {
      loading.value = false
    }
  }

  return {
    summary,
    wallets,
    loading,
    error,
    fetchSummary,
    connectStream,
    disconnectStream,
    exchangeWallets,
    reset,
  }
})
