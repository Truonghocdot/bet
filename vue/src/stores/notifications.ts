import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { request, type ApiError } from '@/shared/api/http'
import type {
  NotificationListItem,
  NotificationListResponse,
  NotificationReadResponse,
} from '@/shared/api/types'
import { useAuthStore } from '@/stores/auth'

type Pagination = {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

export const useNotificationsStore = defineStore('notifications', () => {
  const items = ref<NotificationListItem[]>([])
  const loading = ref(false)
  const markingReadId = ref<number | null>(null)
  const error = ref('')
  const pagination = ref<Pagination>({
    page: 1,
    pageSize: 10,
    total: 0,
    totalPages: 1,
  })

  const unreadCount = computed(() => items.value.filter((item) => !item.is_read).length)

  async function fetchList(page = 1, pageSize = pagination.value.pageSize) {
    const auth = useAuthStore()
    if (!auth.accessToken) {
      reset()
      return null
    }

    loading.value = true
    error.value = ''
    try {
      const res = await request<NotificationListResponse>(
        'GET',
        `/v1/notifications?page=${page}&page_size=${pageSize}`,
        { token: auth.accessToken },
      )
      items.value = res.items
      pagination.value = {
        page: res.page,
        pageSize: res.page_size,
        total: res.total,
        totalPages: res.total_pages || 1,
      }
      return res
    } catch (e: any) {
      const err = e as ApiError
      if (err?.status === 401) {
        auth.logout()
        reset()
        throw e
      }
      error.value = err?.message ?? 'Không thể tải danh sách thông báo'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function markRead(id: number) {
    const auth = useAuthStore()
    if (!auth.accessToken || !id) return null

    markingReadId.value = id
    error.value = ''
    try {
      const res = await request<NotificationReadResponse>('POST', `/v1/notifications/${id}/read`, {
        token: auth.accessToken,
      })
      const target = items.value.find((item) => item.id === id)
      if (target) {
        target.is_read = true
        target.read_at = res.read_at
      }
      return res
    } catch (e: any) {
      const err = e as ApiError
      if (err?.status === 401) {
        auth.logout()
        reset()
        throw e
      }
      error.value = err?.message ?? 'Không thể cập nhật trạng thái thông báo'
      throw e
    } finally {
      markingReadId.value = null
    }
  }

  function reset() {
    items.value = []
    loading.value = false
    markingReadId.value = null
    error.value = ''
    pagination.value = {
      page: 1,
      pageSize: 10,
      total: 0,
      totalPages: 1,
    }
  }

  return {
    items,
    loading,
    markingReadId,
    error,
    pagination,
    unreadCount,
    fetchList,
    markRead,
    reset,
  }
})
