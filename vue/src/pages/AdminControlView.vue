<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import ControlRoomCard from '@/components/ControlRoomCard.vue'
import { request, type ApiError } from '@/shared/api/http'
import { connectEventStream, type StreamConnection } from '@/shared/api/stream'
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

const pendingSubmit = ref<ManualSubmitRequest | null>(null)
const submittingPeriodId = ref<number | null>(null)
const toast = ref<{ type: 'success' | 'error'; text: string } | null>(null)
const pulseUntil = ref<Record<string, number>>({})

let clockTimer: number | undefined
let toastTimer: number | undefined
let streamConnection: StreamConnection | null = null

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
  const serverTime = new Date(response.server_time).getTime()
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

function startStatsStream() {
  if (!autoRefresh.value || !auth.accessToken || streamConnection) return
  streamConnection = connectEventStream('/v1/admin/rooms/stats/stream', {
    token: auth.accessToken,
    reconnectMs: 2000,
    onEvent(payload) {
      if (payload.event !== 'admin.rooms.stats') return
      if (!payload.data || typeof payload.data !== 'object') return
      applySnapshot(payload.data as { server_time: string; rooms: RoomStats[] })
      error.value = ''
      loading.value = false
    },
    onError(err) {
      const apiError = err as ApiError
      error.value = apiError?.message || 'Mất kết nối realtime admin'
    },
  })
}

function stopStatsStream() {
  streamConnection?.close()
  streamConnection = null
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

onMounted(() => {
  void fetchStats()
  startStatsStream()
  clockTimer = window.setInterval(() => {
    clockTick.value = Date.now()
  }, 1000)
})

onBeforeUnmount(() => {
  stopStatsStream()
  if (clockTimer) window.clearInterval(clockTimer)
  if (toastTimer) window.clearTimeout(toastTimer)
})

watch(autoRefresh, (enabled) => {
  if (enabled) {
    void fetchStats()
    startStatsStream()
  } else {
    stopStatsStream()
  }
})
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
        <button class="live-btn" :class="{ 'is-on': autoRefresh }" @click="autoRefresh = !autoRefresh">
          {{ autoRefresh ? 'Realtime ON' : 'Realtime OFF' }}
        </button>
      </div>
    </header>

    <p v-if="error" class="error-banner">{{ error }}</p>

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
