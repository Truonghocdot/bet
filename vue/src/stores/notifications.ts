import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { request, type ApiError } from '@/shared/api/http'
import { connectEventStream, type StreamConnection } from '@/shared/api/stream'
import type {
  NotificationListItem,
  NotificationListResponse,
  NotificationReadResponse,
  NotificationRespondResponse,
} from '@/shared/api/types'
import { useAuthStore } from '@/stores/auth'

type Pagination = {
  page: number
  pageSize: number
  total: number
  totalPages: number
  unreadCount: number
}

export const useNotificationsStore = defineStore('notifications', () => {
  const items = ref<NotificationListItem[]>([])
  const loading = ref(false)
  const markingReadId = ref<number | null>(null)
  const respondingId = ref<number | null>(null)
  const respondingAction = ref<'confirm' | 'cancel' | null>(null)
  const error = ref('')
  let streamConnection: StreamConnection | null = null
  const pagination = ref<Pagination>({
    page: 1,
    pageSize: 10,
    total: 0,
    totalPages: 1,
    unreadCount: 0,
  })

  const unreadCount = computed(() => pagination.value.unreadCount)

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
      applyListResponse(res)
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

  function applyListResponse(res: NotificationListResponse) {
    items.value = res.items
    pagination.value = {
      page: res.page,
      pageSize: res.page_size,
      total: res.total,
      totalPages: res.total_pages || 1,
      unreadCount: Math.max(0, Number(res.unread_count ?? 0)),
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
      const wasUnread = target ? !target.is_read : false
      if (target) {
        target.is_read = true
        target.read_at = res.read_at
      }
      if (wasUnread) {
        pagination.value.unreadCount = Math.max(0, pagination.value.unreadCount - 1)
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

  async function respond(id: number, action: 'confirm' | 'cancel') {
    const auth = useAuthStore()
    if (!auth.accessToken || !id) return null

    respondingId.value = id
    respondingAction.value = action
    error.value = ''
    try {
      const res = await request<NotificationRespondResponse>('POST', `/v1/notifications/${id}/respond`, {
        token: auth.accessToken,
        body: { action },
      })
      const target = items.value.find((item) => item.id === id)
      const wasUnread = target ? !target.is_read : false
      if (target) {
        target.response_status = res.response_status
        target.responded_at = res.responded_at
        target.is_read = true
        target.read_at = res.read_at
        target.can_respond = false
      }
      if (wasUnread) {
        pagination.value.unreadCount = Math.max(0, pagination.value.unreadCount - 1)
      }
      return res
    } catch (e: any) {
      const err = e as ApiError
      if (err?.status === 401) {
        auth.logout()
        reset()
        throw e
      }
      error.value = err?.message ?? 'Không thể cập nhật phản hồi thông báo'
      throw e
    } finally {
      respondingId.value = null
      respondingAction.value = null
    }
  }

  function connectStream(page = pagination.value.page, pageSize = pagination.value.pageSize) {
    const auth = useAuthStore()
    if (!auth.accessToken) return

    disconnectStream()
    streamConnection = connectEventStream(`/v1/notifications/stream?page=${page}&page_size=${pageSize}`, {
      token: auth.accessToken,
      reconnectMs: 4000,
      onEvent(payload) {
        if (payload.event !== 'notifications.list') return
        applyListResponse(payload.data as NotificationListResponse)
      },
      onError(errorValue) {
        const err = errorValue as ApiError
        if (err?.status === 401) {
          auth.logout()
          reset()
          return
        }
        error.value = err?.message ?? 'Kết nối thông báo realtime bị gián đoạn'
      },
    })
  }

  function disconnectStream() {
    streamConnection?.close()
    streamConnection = null
  }

  function reset() {
    disconnectStream()
    items.value = []
    loading.value = false
    markingReadId.value = null
    respondingId.value = null
    respondingAction.value = null
    error.value = ''
    pagination.value = {
      page: 1,
      pageSize: 10,
      total: 0,
      totalPages: 1,
      unreadCount: 0,
    }
  }

  function addLocalNotification(title: string, body: string, type: 'info' | 'error' | 'success' = 'success') {
    const id = Date.now()
    items.value.unshift({
       id,
       title,
       body,
       image_url: null,
       status: 1,
       audience: 1,
       created_at: new Date().toISOString(),
       is_read: false,
       can_respond: false,
     })
    pagination.value.total += 1
    pagination.value.unreadCount += 1
  }

  return {
    items,
    loading,
    markingReadId,
    respondingId,
    respondingAction,
    error,
    pagination,
    unreadCount,
    fetchList,
    connectStream,
    disconnectStream,
    markRead,
    respond,
    addLocalNotification,
    reset,
  }
})
