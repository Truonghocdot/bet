import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { request, type ApiError } from '@/shared/api/http'
import { env } from '@/shared/config/env'
import { useAuthStore } from '@/stores/auth'

export type WheelInvitation = {
  id: string
  campaign_name: string
  status: 'pending' | 'started' | 'completed' | 'expired' | 'revoked' | string
  expires_at?: string | null
  seen_at?: string | null
  session_id?: string | null
  session_status?: string | null
}

type SocketEvent = {
  event?: string
  data?: WheelInvitation | { items?: WheelInvitation[]; invitation_id?: string; campaign_name?: string; expires_at?: string | null }
}

export const useWheelInvitationsStore = defineStore('wheel-invitations', () => {
  const auth = useAuthStore()
  const items = ref<WheelInvitation[]>([])
  const loading = ref(false)
  const launchingId = ref<string | null>(null)
  const error = ref('')
  const connected = ref(false)
  let socket: WebSocket | null = null
  let reconnectTimer: number | null = null
  let stopped = true

  const activePopup = computed(() => items.value.find((item) => item.status === 'pending' && !item.seen_at) ?? null)
  const actionableCount = computed(() => items.value.filter((item) => item.status === 'pending' || item.status === 'started').length)

  function merge(nextItems: WheelInvitation[]) {
    const byID = new Map(items.value.map((item) => [item.id, item]))
    for (const item of nextItems) byID.set(item.id, { ...byID.get(item.id), ...item })
    items.value = [...byID.values()].sort((a, b) => String(b.expires_at ?? '').localeCompare(String(a.expires_at ?? '')))
  }

  async function fetchInvitations() {
    if (!env.wheelEventEnabled || !auth.isAuthenticated) return
    loading.value = true
    error.value = ''
    try {
      const response = await request<{ items: WheelInvitation[] }>('GET', '/v1/wheel/invitations', { token: auth.accessToken })
      items.value = response.items ?? []
    } catch (cause) {
      error.value = (cause as ApiError).message || 'Không thể tải danh sách sự kiện.'
    } finally {
      loading.value = false
    }
  }

  async function dismiss(invitation: WheelInvitation) {
    invitation.seen_at = new Date().toISOString()
    try {
      await request('POST', `/v1/wheel/invitations/${encodeURIComponent(invitation.id)}/seen`, { token: auth.accessToken })
    } catch {
      // Local state prevents the popup loop; REST sync will reconcile later.
    }
  }

  async function launch(invitation: WheelInvitation) {
    if (launchingId.value) return
    launchingId.value = invitation.id
    error.value = ''
    const target = window.open('about:blank', '_blank')
    if (target) target.opener = null
    try {
      const response = await request<{ url: string }>('POST', `/v1/wheel/invitations/${encodeURIComponent(invitation.id)}/launch`, { token: auth.accessToken })
      if (target) target.location.href = response.url
      else window.location.href = response.url
      await dismiss(invitation)
    } catch (cause) {
      target?.close()
      error.value = (cause as ApiError).message || 'Không thể mở sự kiện.'
      throw cause
    } finally {
      launchingId.value = null
    }
  }

  function websocketURL(ticket: string) {
    const source = env.apiBaseUrl || window.location.origin
    const url = new URL(source, window.location.origin)
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
    url.pathname = `${url.pathname.replace(/\/$/, '')}/v1/users/me/events/ws`
    url.searchParams.set('ticket', ticket)
    return url.toString()
  }

  async function connect() {
    if (stopped || !env.wheelEventEnabled || !auth.isAuthenticated || socket) return
    try {
      const response = await request<{ ticket: string }>('POST', '/v1/realtime/tickets', { token: auth.accessToken })
      if (stopped) return
      socket = new WebSocket(websocketURL(response.ticket))
      socket.onopen = () => { connected.value = true }
      socket.onmessage = (message) => handleSocketMessage(String(message.data))
      socket.onerror = () => { connected.value = false }
      socket.onclose = () => {
        connected.value = false
        socket = null
        if (!stopped) reconnectTimer = window.setTimeout(() => void connect(), 4000)
      }
    } catch {
      if (!stopped) reconnectTimer = window.setTimeout(() => void connect(), 4000)
    }
  }

  function handleSocketMessage(raw: string) {
    let payload: SocketEvent
    try { payload = JSON.parse(raw) as SocketEvent } catch { return }
    if (payload.event === 'wheel.invitations' && payload.data && 'items' in payload.data) {
      items.value = payload.data.items ?? []
      return
    }
    if (payload.event === 'wheel.invitation.activated' && payload.data && 'invitation_id' in payload.data) {
      merge([{
        id: String(payload.data.invitation_id),
        campaign_name: String(payload.data.campaign_name ?? 'Vòng quay may mắn'),
        status: 'pending',
        expires_at: payload.data.expires_at ?? null,
        seen_at: null,
      }])
      return
    }
    if ((payload.event === 'wheel.invitation.revoked' || payload.event === 'wheel.invitation.expired') && payload.data && 'invitation_id' in payload.data) {
      const invitationID = String(payload.data.invitation_id)
      const item = items.value.find((candidate) => candidate.id === invitationID)
      if (item) item.status = payload.event.endsWith('revoked') ? 'revoked' : 'expired'
    }
  }

  function start() {
    stop()
    stopped = false
    void fetchInvitations()
    void connect()
  }

  function stop() {
    stopped = true
    connected.value = false
    if (reconnectTimer !== null) window.clearTimeout(reconnectTimer)
    reconnectTimer = null
    socket?.close()
    socket = null
  }

  function reset() {
    stop()
    items.value = []
    error.value = ''
  }

  return { items, loading, launchingId, error, connected, activePopup, actionableCount, fetchInvitations, dismiss, launch, start, stop, reset }
})
