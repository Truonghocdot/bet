import { ref } from 'vue'
import { defineStore } from 'pinia'

import { request, type ApiError } from '@/shared/api/http'
import { connectEventStream, type StreamConnection } from '@/shared/api/stream'
import { readJSON, remove, writeJSON } from '@/shared/lib/storage'
import type {
  DepositInitResponse,
  DepositStatusResponse,
  VietQrBankListResponse,
  VietQrBankOption,
} from '@/shared/api/types'
import { useAuthStore } from '@/stores/auth'

type PersistedPendingDeposit = {
  method: 'vietqr' | 'usdt'
  intent: DepositInitResponse
}

const PENDING_STORAGE_KEY = 'ff789:deposit:pending:v1'

export const useDepositStore = defineStore('deposit', () => {
  const currentIntent = ref<DepositInitResponse | null>(null)
  const currentStatus = ref<DepositStatusResponse | null>(null)
  const bankOptions = ref<VietQrBankOption[]>([])
  const banksLoading = ref(false)
  const loading = ref(false)
  const error = ref<string>('')
  let streamConnection: StreamConnection | null = null

  function pendingStorageKey(): string {
    const auth = useAuthStore()
    const userId = auth.user?.id ?? 0
    return `${PENDING_STORAGE_KEY}:${userId}`
  }

  function persistPending(method: 'vietqr' | 'usdt', intent: DepositInitResponse) {
    writeJSON(`${pendingStorageKey()}:${method}`, { method, intent } satisfies PersistedPendingDeposit)
  }

  function restorePending(method: 'vietqr' | 'usdt'): PersistedPendingDeposit | null {
    const saved = readJSON<PersistedPendingDeposit>(`${pendingStorageKey()}:${method}`)
    if (!saved?.intent?.client_ref) return null

    const expiresAt = Date.parse(saved.intent.expires_at)
    if (Number.isFinite(expiresAt) && expiresAt > 0 && Date.now() >= expiresAt) {
      clearPending(method)
      return null
    }

    return saved
  }

  function clearPending(method?: 'vietqr' | 'usdt') {
    if (method) {
      remove(`${pendingStorageKey()}:${method}`)
      return
    }

    remove(`${pendingStorageKey()}:vietqr`)
    remove(`${pendingStorageKey()}:usdt`)
  }

  async function initVietQR(payload: { amount: string; note?: string; provider_code?: string }) {
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
      persistPending('vietqr', res)
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
      persistPending('usdt', res)
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
      const status = res.transaction?.status
      if (status === 2 || status === 3 || status === 4) {
        clearPending(currentIntent.value?.method as 'vietqr' | 'usdt' | undefined)
      }
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
          clearPending(currentIntent.value?.method as 'vietqr' | 'usdt' | undefined)
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
    clearPending()
  }

  async function loadVietQrBanks() {
    const auth = useAuthStore()
    if (!auth.accessToken) {
      bankOptions.value = []
      return []
    }

    banksLoading.value = true
    error.value = ''
    try {
      const res = await request<VietQrBankListResponse>('GET', '/v1/deposits/vietqr/banks', {
        token: auth.accessToken,
      })
      bankOptions.value = res.banks ?? []
      return bankOptions.value
    } catch (e: any) {
      const err = e as ApiError
      error.value = err?.message ?? 'Không thể tải danh sách ngân hàng'
      throw e
    } finally {
      banksLoading.value = false
    }
  }

  return {
    currentIntent,
    currentStatus,
    bankOptions,
    banksLoading,
    loading,
    error,
    initVietQR,
    initUSDT,
    getStatus,
    connectStatusStream,
    disconnectStatusStream,
    loadVietQrBanks,
    restorePending,
    clearPending,
    reset,
  }
})
