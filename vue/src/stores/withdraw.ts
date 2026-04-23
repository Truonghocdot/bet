import { defineStore } from 'pinia'
import { useAuthStore } from './auth'
import { useNotificationsStore } from './notifications'
import { request, type ApiError } from '@/shared/api/http'
import type { SetupAccountRequest, WithdrawalRequest } from '@/shared/api/types'
import { computed, ref } from 'vue'

export interface WithdrawalAccount {
  id: number
  unit: number
  account_number: string
  is_default: boolean
  created_at: string
}

export const useWithdrawStore = defineStore('withdraw', () => {
  const auth = useAuthStore()
  const notify = useNotificationsStore()

  const accounts = ref<WithdrawalAccount[]>([])
  const history = ref<WithdrawalRequest[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const vndAccounts = computed(() => accounts.value.filter((a) => a.unit === 1))
  const usdtAccounts = computed(() => accounts.value.filter((a) => a.unit === 2))

  async function fetchAccounts(): Promise<any> {
    if (!auth.accessToken) return
    loading.value = true
    error.value = null

    try {
      const res = await request<{ data: WithdrawalAccount[] }>('GET', '/v1/withdrawals/accounts', {
         token: auth.accessToken
      })
      accounts.value = res.data || []
      return res
    } catch (e: any) {
      const err = e as ApiError
      if (err?.status === 401) {
        if (auth.refreshToken) {
           try {
             await auth.refresh()
             return await fetchAccounts()
           } catch {
             auth.logout()
             throw e
           }
        }
        auth.logout()
        throw e
      }
      error.value = err.message || 'Có lỗi khi tải danh sách phương thức rút tiền'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function addAccount(payload: SetupAccountRequest): Promise<any> {
    if (!auth.accessToken) return
    loading.value = true
    error.value = null

    try {
      const res = await request<{ id: number; message: string }>('POST', '/v1/withdrawals/accounts', {
        token: auth.accessToken,
        body: { ...payload, is_default: true }
      })
      
      notify.addLocalNotification('Thành công', 'Đã lưu cấu hình nhận tiền của bạn.')
      await fetchAccounts()
      return res
    } catch (e: any) {
      const err = e as ApiError
      if (err?.status === 401) {
        if (auth.refreshToken) {
           try {
              await auth.refresh()
              return await addAccount(payload)
           } catch {
              auth.logout()
              throw e
           }
        }
        auth.logout()
        throw e
      }
      error.value = err.message || 'Thêm tài khoản thất bại'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function deleteAccount(id: number): Promise<any> {
    if (!auth.accessToken) return
    loading.value = true
    error.value = null

    try {
       const res = await request<{ message: string }>('DELETE', `/v1/withdrawals/accounts/${id}`, {
          token: auth.accessToken
       })
      notify.addLocalNotification('Thành công', 'Đã xoá hồ sơ nhận tiền.')
      accounts.value = accounts.value.filter((a) => a.id !== id)
      return res
    } catch (e: any) {
       const err = e as ApiError
       if (err?.status === 401) {
          if (auth.refreshToken) {
             try {
                await auth.refresh()
                return await deleteAccount(id)
             } catch {
                auth.logout()
                throw e
             }
          }
          auth.logout()
          throw e
       }
      error.value = err.message || 'Có lỗi khi xóa phương thức'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function submitWithdrawal(payload: { account_withdrawal_info_id: number; amount: string; password: string }): Promise<boolean> {
    if (!auth.accessToken) return false
    loading.value = true
    error.value = null

    try {
      await request<any>('POST', '/v1/withdrawals', {
        token: auth.accessToken,
        body: payload
      })
      
      notify.addLocalNotification('Đã tạo lệnh rút', 'Vui lòng chờ chuyên viên xét duyệt phiếu rút.')
      return true
    } catch (e: any) {
      const err = e as ApiError
      if (err?.status === 401) {
         if (auth.refreshToken) {
            try {
               await auth.refresh()
               return await submitWithdrawal(payload)
            } catch {
               auth.logout()
               return false
            }
         }
         auth.logout()
         return false
      }
      error.value = err.message
      notify.addLocalNotification('Lỗi', err.message, 'error')
      return false
    } finally {
      loading.value = false
    }
  }

  async function fetchHistory(limit = 20, offset = 0): Promise<any> {
    if (!auth.accessToken) return
    loading.value = true
    error.value = null

    try {
       const res = await request<{ data: WithdrawalRequest[] }>('GET', `/v1/withdrawals?limit=${limit}&offset=${offset}`, {
          token: auth.accessToken
       })
       history.value = res.data || []
       return res
    } catch (e: any) {
       const err = e as ApiError
       if (err?.status === 401) {
          if (auth.refreshToken) {
             try {
                await auth.refresh()
                return await fetchHistory(limit, offset)
             } catch {
                auth.logout()
                throw e
             }
          }
          auth.logout()
          throw e
       }
       error.value = err.message || 'Không thể tải lịch sử rút tiền'
       throw e
    } finally {
       loading.value = false
    }
  }

  return {
    accounts,
    vndAccounts,
    usdtAccounts,
    loading,
    error,
    history,
    fetchAccounts,
    addAccount,
    deleteAccount,
    submitWithdrawal,
    fetchHistory,
  }
})
