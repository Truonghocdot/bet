import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { useAuthStore } from './auth'
import { useNotificationsStore } from './notifications'

export interface WithdrawalAccount {
  id: number
  unit: number
  provider_code: string
  account_name: string
  account_number: string
  is_default: boolean
  created_at: string
}

export const useWithdrawStore = defineStore('withdraw', () => {
  const auth = useAuthStore()
  const notify = useNotificationsStore()

  const accounts = ref<WithdrawalAccount[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const vndAccounts = computed(() => accounts.value.filter((a) => a.unit === 1))
  const usdtAccounts = computed(() => accounts.value.filter((a) => a.unit === 2))

  async function fetchAccounts() {
    if (!auth.accessToken) return
    loading.value = true
    error.value = null

    try {
      const response = await fetch('/api/v1/withdrawals/accounts', {
        headers: {
          'Authorization': `Bearer ${auth.accessToken}`,
          'Accept': 'application/json',
        },
      })
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data?.message || 'Có lỗi khi tải danh sách phương thức rút tiền')
      }
      accounts.value = data.data || []
    } catch (err: any) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  async function addAccount(payload: { unit: number; provider_code: string; account_name: string; account_number: string }) {
    if (!auth.accessToken) return
    loading.value = true
    error.value = null

    try {
      const response = await fetch('/api/v1/withdrawals/accounts', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${auth.accessToken}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        body: JSON.stringify({ ...payload, is_default: true }),
      })
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data?.message || 'Thêm tài khoản thất bại')
      }
      notify.addLocalNotification('Thành công', 'Đã lưu cấu hình nhận tiền của bạn.')
      await fetchAccounts()
    } catch (err: any) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteAccount(id: number) {
    if (!auth.accessToken) return
    loading.value = true
    error.value = null

    try {
      const response = await fetch(`/api/v1/withdrawals/accounts/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${auth.accessToken}`,
          'Accept': 'application/json',
        },
      })
      if (!response.ok) {
        const data = await response.json()
        throw new Error(data?.message || 'Có lỗi khi xóa phương thức')
      }
      notify.addLocalNotification('Thành công', 'Đã xoá hồ sơ nhận tiền.')
      accounts.value = accounts.value.filter((a) => a.id !== id)
    } catch (err: any) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function submitWithdrawal(payload: { account_withdrawal_info_id: number; amount: string }) {
    if (!auth.accessToken) return false
    loading.value = true
    error.value = null

    try {
      const response = await fetch('/api/v1/withdrawals/', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${auth.accessToken}`,
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        body: JSON.stringify(payload),
      })
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data?.message || 'Không thể tạo lệnh rút đổi')
      }
      notify.addLocalNotification('Đã tạo lệnh rút', 'Vui lòng chờ chuyên viên xét duyệt phiếu rút.')
      return true
    } catch (err: any) {
      error.value = err.message
      notify.addLocalNotification('Lỗi', err.message, 'error')
      return false
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
    fetchAccounts,
    addAccount,
    deleteAccount,
    submitWithdrawal,
  }
})
