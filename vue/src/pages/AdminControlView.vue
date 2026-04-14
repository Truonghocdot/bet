<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import ControlRoomCard from '@/components/ControlRoomCard.vue'
import { env } from '@/shared/config/env'
import { request, type ApiError } from '@/shared/api/http'
import { useAuthStore } from '@/stores/auth'

interface BetStat {
  option_key: string
  option_type: number
  player_count: number
  total_stake: string
}

interface RoomStats {
  code: string
  game_type: number
  period: {
    id: number
    period_no: string
    draw_at: string
    bet_lock_at: string
    status: number
    manual_result: string | null
  } | null
  bet_stats: BetStat[]
}

type ManualSubmitRequest = {
  roomCode: string
  periodId: number
  result: string
  bigSmall: string
  color: string
  payload: Record<string, any>
  title: string
}

const auth = useAuthStore()

const rooms = ref<RoomStats[]>([])
const loading = ref(false)
const error = ref('')
const activeTab = ref(1)
const autoRefresh = ref(true)
const serverTimeOffsetMs = ref(0)
const clockTick = ref(Date.now())
const wsState = ref<'connecting' | 'connected' | 'disconnected'>('disconnected')

const pendingSubmit = ref<ManualSubmitRequest | null>(null)
const submittingPeriodId = ref<number | null>(null)
const toast = ref<{ type: 'success' | 'error'; text: string } | null>(null)
const pulseUntil = ref<Record<string, number>>({})

const lockState = ref<'checking' | 'acquired' | 'denied'>('checking')
const lockMessage = ref('')
const lockHolder = ref('')

let clockTimer: number | undefined
let toastTimer: number | undefined
let fallbackPollTimer: number | undefined
let wsConnection: WebSocket | null = null
let wsReconnectTimer: number | undefined
let heartbeatTimer: number | undefined

async function acquireLock() {
  try {
    const res = await request<{ message: string; holder?: string }>('POST', '/v1/admin/lock', {
      token: auth.accessToken,
    })
    lockState.value = 'acquired'
    return true
  } catch (err: any) {
    const apiError = err as ApiError
    lockState.value = 'denied'
    lockMessage.value = apiError?.message || 'Không thể giành quyền kiểm soát trang quản lý'
    lockHolder.value = (apiError as any)?.holder || 'Admin khác'
    return false
  }
}

async function sendHeartbeat() {
  try {
    await request('PUT', '/v1/admin/lock', {
      token: auth.accessToken,
    })
  } catch (err: any) {
    const apiError = err as ApiError
    if (apiError.status === 403 || apiError.status === 410) {
      lockState.value = 'denied'
      lockMessage.value = apiError.message
      stopAdminEngine()
    }
  }
}

async function releaseLock() {
  try {
    await request('DELETE', '/v1/admin/lock', {
      token: auth.accessToken,
    })
  } catch {
    // ignore
  }
}

function parseStake(value: string): number {
  const num = Number.parseFloat(value ?? '0')
  return Number.isFinite(num) ? num : 0
}

function roomStakeMap(items: RoomStats[]): Record<string, number> {
  const result: Record<string, number> = {}
  for (const room of items) {
    result[room.code] = (room.bet_stats ?? []).reduce((sum, item) => sum + parseStake(item.total_stake), 0)
  }
  return result
}

function applySnapshot(response: { server_time: string; rooms: RoomStats[] }) {
  const previous = roomStakeMap(rooms.value)
  const now = Date.now()

  rooms.value = response.rooms
  const cleanServerTime = response.server_time.substring(0, 19).replace(' ', 'T')
  const serverTime = new Date(cleanServerTime).getTime()
  serverTimeOffsetMs.value = serverTime - now

  for (const room of response.rooms) {
    const before = previous[room.code] ?? 0
    const after = (room.bet_stats ?? []).reduce((sum, item) => sum + parseStake(item.total_stake), 0)
    if (Math.abs(after - before) > 0.0001) {
      pulseUntil.value[room.code] = now + 1400
    }
  }
}

function isRoomPulsing(roomCode: string): boolean {
  return (pulseUntil.value[roomCode] ?? 0) > Date.now()
}

async function fetchStats() {
  if (!auth.isAuthenticated) {
    loading.value = false
    return
  }
  if (rooms.value.length === 0) loading.value = true

  try {
    const response = await request<{ server_time: string; rooms: RoomStats[] }>('GET', '/v1/admin/rooms/stats', {
      token: auth.accessToken,
    })
    applySnapshot(response)
    error.value = ''
  } catch (err: any) {
    const apiError = err as ApiError
    error.value = apiError?.message || 'Không tải được dữ liệu điều khiển'
  } finally {
    loading.value = false
  }
}

function buildAdminWSUrl(token: string): string {
  const endpoint = '/v1/admin/rooms/stats/ws'
  const encodedToken = encodeURIComponent(token)
  const fallbackProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const fallback = `${fallbackProtocol}//${window.location.host}${endpoint}?token=${encodedToken}`

  try {
    const rawBase = (env.apiBaseUrl || '').trim()
    const baseUrl = new URL(rawBase || window.location.origin, window.location.origin)
    const wsProtocol = baseUrl.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${wsProtocol}//${baseUrl.host}${endpoint}?token=${encodedToken}`
  } catch {
    return fallback
  }
}

function scheduleWSReconnect() {
  if (!autoRefresh.value || wsReconnectTimer) return
  wsReconnectTimer = window.setTimeout(() => {
    wsReconnectTimer = undefined
    startStatsStream()
  }, 2000)
}

function startStatsStream() {
  if (!autoRefresh.value || !auth.accessToken || wsConnection) return

  try {
    wsState.value = 'connecting'
    const socket = new WebSocket(buildAdminWSUrl(auth.accessToken))
    wsConnection = socket
    socket.onopen = () => {
      wsState.value = 'connected'
      error.value = ''
    }

    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(String(event.data)) as { event?: string; data?: any }
        if (payload?.event !== 'admin.rooms.stats') return
        if (!payload.data || typeof payload.data !== 'object') return
        applySnapshot(payload.data as { server_time: string; rooms: RoomStats[] })
        error.value = ''
        loading.value = false
      } catch {
        // ignore malformed payload
      }
    }

    socket.onerror = () => {
      wsState.value = 'disconnected'
      error.value = 'Mất kết nối realtime admin'
    }

    socket.onclose = () => {
      wsState.value = 'disconnected'
      wsConnection = null
      scheduleWSReconnect()
    }
  } catch {
    wsState.value = 'disconnected'
    error.value = 'Không thể khởi tạo realtime admin'
    scheduleWSReconnect()
  }
}

function stopStatsStream() {
  if (wsReconnectTimer) {
    window.clearTimeout(wsReconnectTimer)
    wsReconnectTimer = undefined
  }
  wsConnection?.close()
  wsConnection = null
  wsState.value = 'disconnected'
}

function startFallbackPolling() {
  if (!autoRefresh.value || fallbackPollTimer) return
  fallbackPollTimer = window.setInterval(() => {
    if (!autoRefresh.value) return
    void fetchStats()
  }, 5000)
}

function stopFallbackPolling() {
  if (!fallbackPollTimer) return
  window.clearInterval(fallbackPollTimer)
  fallbackPollTimer = undefined
}

function startAdminEngine() {
  void fetchStats()
  startStatsStream()
  startFallbackPolling()
}

function stopAdminEngine() {
  stopStatsStream()
  stopFallbackPolling()
}

function showToast(type: 'success' | 'error', text: string) {
  toast.value = { type, text }
  if (toastTimer) window.clearTimeout(toastTimer)
  toastTimer = window.setTimeout(() => {
    toast.value = null
  }, 3200)
}

function openSubmitModal(payload: ManualSubmitRequest) {
  if (submittingPeriodId.value) return
  pendingSubmit.value = payload
}

function closeSubmitModal() {
  if (submittingPeriodId.value) return
  pendingSubmit.value = null
}

async function confirmSubmit() {
  const job = pendingSubmit.value
  if (!job || submittingPeriodId.value) return

  submittingPeriodId.value = job.periodId
  try {
    await request('POST', `/v1/admin/periods/${job.periodId}/result`, {
      token: auth.accessToken,
      body: {
        result: job.result,
        big_small: job.bigSmall,
        color: job.color,
        payload: job.payload,
      },
    })
    showToast('success', `Đã cài kết quả ${job.title}`)
    pendingSubmit.value = null
    await fetchStats()
  } catch (err: any) {
    const apiError = err as ApiError
    showToast('error', apiError?.message || 'Không thể cài kết quả, vui lòng thử lại')
  } finally {
    submittingPeriodId.value = null
  }
}

const syncedNowMs = computed(() => clockTick.value + serverTimeOffsetMs.value)
const filteredRooms = computed(() => rooms.value.filter((room) => Number(room.game_type) === Number(activeTab.value)))

function normalizeOptionLabel(rawKey: string): string {
  const key = String(rawKey || '').trim().toLowerCase()
  if (!key) return '—'
  if (key.startsWith('number_')) return `Số ${key.replace('number_', '')}`
  if (key.startsWith('sum_')) return `Tổng ${key.replace('sum_', '')}`
  if (key === 'big') return 'Lớn'
  if (key === 'small') return 'Nhỏ'
  if (key === 'odd') return 'Lẻ'
  if (key === 'even') return 'Chẵn'
  if (key === 'red') return 'Đỏ'
  if (key === 'green') return 'Xanh'
  if (key === 'violet') return 'Tím'
  if (key === 'red_violet') return 'Đỏ / Tím'
  if (key === 'green_violet') return 'Xanh / Tím'
  return key.replaceAll('_', ' ')
}

const optionVolumeRows = computed(() => {
  const map = new Map<string, { optionKey: string; stake: number; players: number }>()
  for (const room of filteredRooms.value) {
    for (const stat of room.bet_stats ?? []) {
      const key = stat.option_key || 'unknown'
      const prev = map.get(key) ?? { optionKey: key, stake: 0, players: 0 }
      prev.stake += parseStake(stat.total_stake)
      prev.players += Number(stat.player_count || 0)
      map.set(key, prev)
    }
  }
  return Array.from(map.values())
    .sort((a, b) => b.stake - a.stake)
    .slice(0, 12)
})

onMounted(async () => {
  const success = await acquireLock()
  if (!success) return

  startAdminEngine()
  clockTimer = window.setInterval(() => {
    clockTick.value = Date.now()
  }, 1000)

  heartbeatTimer = window.setInterval(() => {
    void sendHeartbeat()
  }, 20000)
})

onBeforeUnmount(() => {
  stopAdminEngine()
  void releaseLock()
  if (clockTimer) window.clearInterval(clockTimer)
  if (toastTimer) window.clearTimeout(toastTimer)
  if (heartbeatTimer) window.clearInterval(heartbeatTimer)
})

watch(autoRefresh, (enabled) => {
  if (lockState.value !== 'acquired') return
  if (enabled) {
    startAdminEngine()
  } else {
    stopAdminEngine()
  }
})

watch(
  () => auth.accessToken,
  async (token) => {
    if (token) {
      const success = await acquireLock()
      if (success && autoRefresh.value) {
        startAdminEngine()
      }
      return
    }
    stopAdminEngine()
    void releaseLock()
  },
)
</script>

<template>
  <div class="admin-control-page min-h-screen p-4 pb-20 md:p-6">
    <header class="glass-panel admin-head">
      <div class="admin-head__left">
        <p class="admin-kicker">Systems Online</p>
        <h1 class="admin-title">Admin Control Center</h1>
        <p class="admin-sub">Realtime engine monitor, heatmap cược và can thiệp an toàn trong vùng mở.</p>
      </div>

      <div class="admin-head__right">
        <div class="tab-switch">
          <button :class="{ 'is-active': activeTab === 1 }" @click="activeTab = 1">WinGo</button>
          <button :class="{ 'is-active': activeTab === 2 }" @click="activeTab = 2">K3</button>
          <button :class="{ 'is-active': activeTab === 3 }" @click="activeTab = 3">5D</button>
        </div>
        <div class="ws-chip" :class="`ws-chip--${wsState}`">
          {{ wsState === 'connected' ? 'WS: Connected' : wsState === 'connecting' ? 'WS: Connecting' : 'WS: Disconnected' }}
        </div>
      </div>
    </header>

    <p v-if="error" class="error-banner">{{ error }}</p>

    <section class="volume-panel glass-card">
      <div class="volume-panel__head">
        <h3>Khối lượng theo cửa</h3>
        <span>Top {{ optionVolumeRows.length }} cửa</span>
      </div>
      <div v-if="optionVolumeRows.length === 0" class="volume-empty">Chưa có lệnh cược trong tab hiện tại.</div>
      <div v-else class="volume-grid">
        <div v-for="item in optionVolumeRows" :key="item.optionKey" class="volume-card">
          <p class="volume-card__label">{{ normalizeOptionLabel(item.optionKey) }}</p>
          <p class="volume-card__stake">{{ item.stake.toLocaleString('vi-VN') }}đ</p>
          <p class="volume-card__meta">{{ item.players.toLocaleString('vi-VN') }} lượt đặt</p>
        </div>
      </div>
    </section>

    <section v-if="loading && rooms.length === 0" class="loading-block">
      <div class="spinner"></div>
      <p>Đang đồng bộ dữ liệu admin...</p>
    </section>

    <section v-else-if="filteredRooms.length === 0" class="empty-block glass-card">
      <p>Không có phòng game trong tab này.</p>
    </section>

    <section v-else class="room-grid">
      <ControlRoomCard
        v-for="room in filteredRooms"
        :key="room.code"
        :room="room"
        :now-ms="syncedNowMs"
        :is-submitting="submittingPeriodId === room.period?.id"
        :pulsing="isRoomPulsing(room.code)"
        @request-submit="openSubmitModal"
      />
    </section>

    <Teleport to="body">
      <div v-if="pendingSubmit" class="modal-wrap">
        <div class="modal-backdrop" @click="closeSubmitModal"></div>
        <div class="modal-card glass-card">
          <h3>Xác nhận can thiệp kết quả</h3>
          <p class="modal-room">{{ pendingSubmit.roomCode.toUpperCase() }}</p>
          <div class="modal-detail">
            <span>Kỳ #{{ pendingSubmit.periodId }}</span>
            <strong>{{ pendingSubmit.title }}</strong>
          </div>
          <p class="modal-note">Hành động này sẽ khóa kết quả can thiệp cho kỳ hiện tại nếu còn trong vùng mở.</p>
          <div class="modal-actions">
            <button class="btn-ghost" :disabled="!!submittingPeriodId" @click="closeSubmitModal">Hủy</button>
            <button class="btn-primary" :disabled="!!submittingPeriodId" @click="confirmSubmit">
              {{ submittingPeriodId ? 'Đang gửi...' : 'Xác nhận cài kết quả' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div v-if="toast" class="toast" :class="toast.type === 'success' ? 'toast--success' : 'toast--error'">
        {{ toast.text }}
      </div>
    </Teleport>

    <!-- Lock Overlay -->
    <Teleport to="body">
      <div v-if="lockState === 'denied'" class="lock-overlay">
        <div class="lock-card glass-card">
          <div class="lock-icon">
            <span class="material-symbols-outlined">lock</span>
          </div>
          <h2>Truy cập bị chặn</h2>
          <p>{{ lockMessage }}</p>
          <div class="lock-meta">
            Người đang giữ: <strong>{{ lockHolder }}</strong>
          </div>
          <button class="btn-primary" @click="() => $router.back()">Quay lại</button>
        </div>
      </div>
      <div v-if="lockState === 'checking'" class="lock-overlay">
        <div class="spinner"></div>
        <p class="mt-4 text-white font-bold">Đang kiểm tra quyền truy cập...</p>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.admin-control-page {
  background:
    radial-gradient(circle at 12% -4%, rgba(16, 185, 129, 0.2), transparent 35%),
    radial-gradient(circle at 100% 20%, rgba(251, 113, 133, 0.18), transparent 42%),
    linear-gradient(180deg, #070f23, #091226);
}

.admin-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 18px;
  padding: 18px;
  margin-bottom: 18px;
}

.admin-kicker {
  font-size: 0.63rem;
  letter-spacing: 0.13em;
  text-transform: uppercase;
  color: #67e8f9;
  font-weight: 800;
}

.admin-title {
  margin-top: 3px;
  font-size: clamp(1.1rem, 2vw, 1.6rem);
  font-weight: 900;
  color: #f8fafc;
}

.admin-sub {
  margin-top: 4px;
  font-size: 0.76rem;
  color: #a5b4fc;
}

.admin-head__right {
  display: grid;
  gap: 10px;
}

.ws-chip {
  justify-self: end;
  border-radius: 999px;
  padding: 6px 10px;
  font-size: 0.68rem;
  font-weight: 900;
  letter-spacing: 0.01em;
  border: 1px solid rgba(148, 163, 184, 0.35);
  color: #cbd5e1;
  background: rgba(15, 23, 42, 0.7);
}

.ws-chip--connected {
  border-color: rgba(16, 185, 129, 0.6);
  color: #86efac;
}

.ws-chip--connecting {
  border-color: rgba(59, 130, 246, 0.6);
  color: #93c5fd;
}

.ws-chip--disconnected {
  border-color: rgba(248, 113, 113, 0.55);
  color: #fda4af;
}

.tab-switch {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 6px;
}

.tab-switch button {
  border-radius: 10px;
  padding: 8px 10px;
  font-size: 0.74rem;
  font-weight: 800;
  color: #cbd5e1;
  border: 1px solid rgba(148, 163, 184, 0.3);
  background: rgba(15, 23, 42, 0.75);
}

.tab-switch button.is-active {
  color: #fff;
  border-color: rgba(16, 185, 129, 0.7);
  background: linear-gradient(120deg, #10b981, #0ea5e9);
}

.live-btn {
  border-radius: 10px;
  padding: 8px 12px;
  font-size: 0.72rem;
  font-weight: 900;
  border: 1px solid rgba(148, 163, 184, 0.3);
  color: #cbd5e1;
  background: rgba(15, 23, 42, 0.76);
}

.live-btn.is-on {
  border-color: rgba(34, 197, 94, 0.6);
  color: #bbf7d0;
}

.error-banner {
  margin-bottom: 12px;
  border-radius: 12px;
  border: 1px solid rgba(251, 113, 133, 0.45);
  background: rgba(136, 19, 55, 0.28);
  padding: 10px 12px;
  color: #fecdd3;
  font-size: 0.8rem;
}

.volume-panel {
  margin-bottom: 12px;
  border-radius: 14px;
  padding: 12px;
}

.volume-panel__head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.volume-panel__head h3 {
  font-size: 0.86rem;
  font-weight: 900;
  color: #f8fafc;
}

.volume-panel__head span {
  font-size: 0.7rem;
  color: #94a3b8;
}

.volume-empty {
  border: 1px dashed rgba(148, 163, 184, 0.35);
  border-radius: 10px;
  padding: 10px;
  color: #94a3b8;
  font-size: 0.75rem;
}

.volume-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.volume-card {
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(15, 23, 42, 0.55);
  padding: 8px 10px;
}

.volume-card__label {
  color: #e2e8f0;
  font-size: 0.72rem;
  font-weight: 800;
}

.volume-card__stake {
  color: #fef08a;
  font-size: 0.82rem;
  font-weight: 900;
  margin-top: 2px;
}

.volume-card__meta {
  color: #94a3b8;
  font-size: 0.68rem;
  margin-top: 2px;
}

.loading-block,
.empty-block {
  min-height: 220px;
  border-radius: 16px;
  display: grid;
  place-items: center;
  color: #cbd5e1;
}

.spinner {
  width: 36px;
  height: 36px;
  border-radius: 999px;
  border: 3px solid rgba(148, 163, 184, 0.25);
  border-top-color: #22d3ee;
  animation: spin 0.8s linear infinite;
}

.room-grid {
  display: grid;
  gap: 14px;
  grid-template-columns: repeat(1, minmax(0, 1fr));
}

.modal-wrap {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: grid;
  place-items: center;
}

.modal-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(2, 6, 23, 0.76);
  backdrop-filter: blur(4px);
}

.modal-card {
  position: relative;
  width: min(92vw, 480px);
  padding: 18px;
  border-radius: 18px;
}

.modal-card h3 {
  color: #f8fafc;
  font-size: 1rem;
  font-weight: 900;
}

.modal-room {
  margin-top: 6px;
  color: #67e8f9;
  font-size: 0.78rem;
  font-weight: 800;
}

.modal-detail {
  margin-top: 10px;
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.26);
  background: rgba(15, 23, 42, 0.65);
  padding: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #e2e8f0;
  font-size: 0.8rem;
}

.modal-detail strong {
  color: #fef08a;
}

.modal-note {
  margin-top: 10px;
  color: #cbd5e1;
  font-size: 0.74rem;
}

.modal-actions {
  margin-top: 14px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.btn-ghost,
.btn-primary {
  border-radius: 10px;
  padding: 9px 12px;
  font-size: 0.74rem;
  font-weight: 800;
}

.btn-ghost {
  border: 1px solid rgba(148, 163, 184, 0.3);
  color: #cbd5e1;
  background: rgba(15, 23, 42, 0.7);
}

.btn-primary {
  border: 1px solid rgba(16, 185, 129, 0.62);
  color: #ecfeff;
  background: linear-gradient(120deg, #10b981, #0ea5e9);
}

.toast {
  position: fixed;
  right: 16px;
  bottom: 18px;
  z-index: 110;
  border-radius: 12px;
  padding: 10px 14px;
  font-size: 0.76rem;
  font-weight: 700;
  border: 1px solid transparent;
}

.toast--success {
  background: rgba(6, 78, 59, 0.9);
  border-color: rgba(52, 211, 153, 0.5);
  color: #d1fae5;
}

.toast--error {
  background: rgba(127, 29, 29, 0.92);
  border-color: rgba(251, 113, 133, 0.6);
  color: #ffe4e6;
}

.lock-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  background: rgba(2, 6, 23, 0.85);
  backdrop-filter: blur(10px);
  display: grid;
  place-items: center;
  padding: 20px;
}

.lock-card {
  width: min(90vw, 400px);
  text-align: center;
  padding: 30px 20px;
  border-radius: 24px;
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.lock-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto 16px;
  background: rgba(244, 63, 94, 0.15);
  border-radius: 999px;
  display: grid;
  place-items: center;
  color: #f43f5e;
}

.lock-icon span {
  font-size: 32px;
}

.lock-card h2 {
  color: #fff;
  font-size: 1.25rem;
  font-weight: 900;
  margin-bottom: 8px;
}

.lock-card p {
  color: #94a3b8;
  font-size: 0.88rem;
  line-height: 1.5;
  margin-bottom: 20px;
}

.lock-meta {
  background: rgba(15, 23, 42, 0.6);
  border-radius: 12px;
  padding: 10px;
  margin-bottom: 24px;
  font-size: 0.82rem;
  color: #cbd5e1;
}

.lock-meta strong {
  color: #38bdf8;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (min-width: 1200px) {
  .room-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 860px) {
  .admin-head {
    flex-direction: column;
  }
  .admin-head__right {
    width: 100%;
  }
}
</style>
