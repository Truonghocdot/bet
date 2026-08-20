<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'

import logo from '@/assets/logo-mobile.webp'
import { request, type ApiError } from '@/shared/api/http'
import { env } from '@/shared/config/env'

type Round = {
  round_no: number
  status: string
  segment_key?: string
  result_label?: string
  prize_amount?: string
  spun_at?: string
}

type Reward = { round_no: number; amount: string; status: string; paid_at?: string }
type WheelState = {
  server_now: string
  invitation_id: string
  campaign_name: string
  session_id?: string
  session_status: string
  started_at?: string
  ends_at?: string
  current_round: number
  next_round_available_at?: string
  spin_duration_seconds: number
  rounds: Round[]
  paid_rewards: Reward[]
  total_reward: string
}
type ChatMessage = { id: number; display_name: string; body: string; actor_type: string; created_at: string }

const tokenKey = 'fh88u:wheel-event-token'
const accessToken = ref(sessionStorage.getItem(tokenKey) ?? '')
const state = ref<WheelState | null>(null)
const booting = ref(true)
const starting = ref(false)
const spinning = ref(false)
const submittingSpin = ref(false)
const spinningRoundNo = ref(0)
const error = ref('')
const result = ref<Round | null>(null)
const rotation = ref(0)
const nowMs = ref(Date.now())
const serverOffsetMs = ref(0)
const messages = ref<ChatMessage[]>([])
const chatBody = ref('')
const chatError = ref('')
const sending = ref(false)
const connected = ref(false)
const messageList = ref<HTMLElement | null>(null)
let clockTimer = 0
let revealTimer = 0
let reconnectTimer = 0
let readyTimer = 0
let socket: WebSocket | null = null
let stopped = false

const wheelLabels = ['50 TRIỆU', 'MAY MẮN', '10 TRIỆU', 'CẢM ƠN', '5 TRIỆU', '2 TRIỆU', '1 TRIỆU', '500K']
const currentRound = computed(() => state.value?.rounds.find((item) => item.round_no === state.value?.current_round) ?? null)
const isReadyToStart = computed(() => state.value && !state.value.session_id && state.value.session_status === 'pending')
const isActive = computed(() => state.value?.session_status === 'active')
const isFinished = computed(() => state.value?.session_status === 'completed' || state.value?.session_status === 'expired')
const serverNowMs = computed(() => nowMs.value + serverOffsetMs.value)
const secondsLeft = computed(() => state.value?.ends_at ? Math.max(0, Math.ceil((new Date(state.value.ends_at).getTime() - serverNowMs.value) / 1000)) : 300)
const countdown = computed(() => `${String(Math.floor(secondsLeft.value / 60)).padStart(2, '0')}:${String(secondsLeft.value % 60).padStart(2, '0')}`)
const nextReadyIn = computed(() => state.value?.next_round_available_at ? Math.max(0, Math.ceil((new Date(state.value.next_round_available_at).getTime() - serverNowMs.value) / 1000)) : 0)
const canSpin = computed(() => isActive.value && currentRound.value?.status === 'ready' && !spinning.value && !submittingSpin.value && nextReadyIn.value === 0)
const syncingNextRound = computed(() => isActive.value && currentRound.value?.status !== 'ready' && nextReadyIn.value === 0 && !spinning.value && !submittingSpin.value)
const canSend = computed(() => isActive.value && chatBody.value.trim().length > 0 && chatBody.value.length <= 280 && !sending.value)
const progress = computed(() => state.value?.rounds.filter((item) => item.status === 'spun').length ?? 0)
const totalReward = computed(() => formatMoney(state.value?.total_reward ?? '0'))
const wheelStyle = computed(() => ({ transform: `rotate(${rotation.value}deg)`, transitionDuration: spinning.value ? '5s' : '0s' }))

function api<T>(method: 'GET' | 'POST', path: string, body?: unknown) {
  return request<T>(method, path, { token: accessToken.value, body })
}

function syncClock(serverNow: string) {
  serverOffsetMs.value = new Date(serverNow).getTime() - Date.now()
}

function applyState(next: WheelState) {
  state.value = next
  syncClock(next.server_now)
  const lastSpun = [...next.rounds].filter((item) => item.status === 'spun').pop()
  if (!spinning.value && lastSpun) result.value = lastSpun
}

async function exchangeLaunchCode() {
  const params = new URLSearchParams(window.location.search)
  const launchCode = params.get('launch_code')
  if (!launchCode) return false
  const response = await request<{ access_token: string; expires_in: number }>('POST', '/v1/wheel/auth/exchange', { body: { launch_code: launchCode } })
  accessToken.value = response.access_token
  sessionStorage.setItem(tokenKey, response.access_token)
  window.history.replaceState({}, document.title, window.location.pathname)
  return true
}

async function loadState() {
  const response = await api<WheelState>('GET', '/v1/wheel/me')
  applyState(response)
  if (response.session_id) {
    recoverAnimation(response)
    // Chat and realtime are enhancements; they must not block the wheel from
    // rendering while a socket handshake or a slow history query is pending.
    void loadChat()
    void connectSocket()
  }
}

async function bootstrap() {
  booting.value = true
  error.value = ''
  try {
    await exchangeLaunchCode()
    if (!accessToken.value) throw new Error('Thiếu mã tham gia sự kiện.')
    await loadState()
  } catch (cause) {
    error.value = (cause as ApiError).message || (cause as Error).message || 'Không thể mở sự kiện.'
  } finally {
    booting.value = false
  }
}

async function startSession() {
  starting.value = true
  error.value = ''
  result.value = null
  try {
    const response = await api<WheelState>('POST', '/v1/wheel/session/start')
    applyState(response)
    void loadChat()
    void connectSocket()
  } catch (cause) {
    error.value = (cause as ApiError).message || 'Không thể bắt đầu phiên.'
  } finally {
    starting.value = false
  }
}

function segmentIndex(key = '') {
  const known: Record<string, number> = { jackpot_50m: 0, try_again: 1, reward_10m: 2, thank_you: 3, reward_5m: 4, reward_2m: 5, reward_1m: 6, reward_500k: 7 }
  if (key in known) return known[key]!
  return [...key].reduce((sum, char) => sum + char.charCodeAt(0), 0) % 8
}

function displayResultLabel(round: Round) {
  const label = String(round.result_label ?? '').trim()
  const amount = Number(round.prize_amount ?? 0)
  const canonical: Record<string, string> = {
    jackpot_50m: '50 triệu đồng',
    reward_10m: '10 triệu đồng',
    reward_5m: '5 triệu đồng',
    reward_2m: '2 triệu đồng',
    reward_1m: '1 triệu đồng',
    reward_500k: '500.000 đồng',
  }
  // A generic losing label left in the admin form must not hide a positive
  // payout that was snapshotted for this round.
  if (amount > 0 && canonical[round.segment_key ?? ''] && ['chúc bạn may mắn', 'cảm ơn bạn đã tham gia'].includes(label.toLocaleLowerCase('vi-VN'))) {
    return canonical[round.segment_key ?? '']
  }
  return label || canonical[round.segment_key ?? ''] || 'Kết quả lượt quay'
}

function spinToRound(round: Round, durationMs = 5000) {
  const index = segmentIndex(round.segment_key)
  const normalized = ((rotation.value % 360) + 360) % 360
  const target = 360 - (index * 45 + 22.5)
  const delta = ((target - normalized + 360) % 360) + 360 * 6
  spinning.value = true
  spinningRoundNo.value = round.round_no
  requestAnimationFrame(() => { rotation.value += delta })
  window.clearTimeout(revealTimer)
  revealTimer = window.setTimeout(() => {
    result.value = round
    spinning.value = false
    spinningRoundNo.value = 0
    if (state.value?.session_status === 'completed') disconnectSocket()
    else void refreshState()
  }, Math.max(100, durationMs))
}

function scheduleReadyRefresh(availableAt?: string | null) {
  window.clearTimeout(readyTimer)
  if (!availableAt) return
  const delay = Math.max(0, new Date(availableAt).getTime() - serverNowMs.value + 80)
  readyTimer = window.setTimeout(() => void refreshState(), Math.min(delay, 10000))
}

async function spin() {
  if (!canSpin.value || !state.value) return
  error.value = ''
  const roundNo = state.value.current_round
  submittingSpin.value = true
  spinningRoundNo.value = roundNo
  try {
    const response = await api<{ state: WheelState; result: Round }>('POST', `/v1/wheel/session/rounds/${roundNo}/spin`)
    applyState(response.state)
    spinToRound(response.result, 5000)
  } catch (cause) {
    const apiError = cause as ApiError
    if (apiError.code === 'ROUND_NOT_READY') {
      error.value = ''
      scheduleReadyRefresh(apiError.available_at ?? state.value.next_round_available_at)
    } else {
      error.value = apiError.message || 'Không thể thực hiện lượt quay.'
      if (apiError.status === 409) scheduleReadyRefresh(state.value.next_round_available_at)
    }
    spinningRoundNo.value = 0
  } finally {
    submittingSpin.value = false
  }
}

function recoverAnimation(next: WheelState) {
  if (next.session_status !== 'active' || !next.next_round_available_at) return
  const remaining = new Date(next.next_round_available_at).getTime() - (Date.now() + serverOffsetMs.value)
  const previous = next.rounds.find((item) => item.round_no === next.current_round - 1 && item.status === 'spun')
  if (previous && remaining > 100) spinToRound(previous, Math.min(5000, remaining))
}

async function refreshState() {
  if (!accessToken.value || spinning.value) return
  try { applyState(await api<WheelState>('GET', '/v1/wheel/session/state')) } catch { /* reconnect retries */ }
}

async function loadChat() {
  if (!state.value?.session_id) return
  try {
    const response = await api<{ items: ChatMessage[] }>('GET', '/v1/wheel/session/chat/messages?limit=60')
    messages.value = response.items ?? []
    await scrollChat()
  } catch { /* Chat remains optional to the wheel flow. */ }
}

async function sendChat() {
  const body = chatBody.value.trim()
  if (!body || !canSend.value) return
  sending.value = true
  chatError.value = ''
  try {
    const response = await api<{ message: ChatMessage }>('POST', '/v1/wheel/session/chat/messages', { body })
    mergeMessage(response.message)
    chatBody.value = ''
    await scrollChat()
  } catch (cause) {
    chatError.value = (cause as ApiError).message || 'Không thể gửi tin nhắn.'
  } finally {
    sending.value = false
  }
}

function mergeMessage(message: ChatMessage) {
  const index = messages.value.findIndex((item) => item.id === message.id)
  if (index >= 0) messages.value[index] = message
  else messages.value.push(message)
}

async function scrollChat() {
  await nextTick()
  if (messageList.value) messageList.value.scrollTop = messageList.value.scrollHeight
}

function websocketURL(ticket: string) {
  const source = env.apiBaseUrl || window.location.origin
  const url = new URL(source, window.location.origin)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = `${url.pathname.replace(/\/$/, '')}/v1/wheel/session/ws`
  url.searchParams.set('ticket', ticket)
  return url.toString()
}

async function connectSocket() {
  if (stopped || socket || !state.value?.session_id || state.value.session_status !== 'active') return
  try {
    const response = await api<{ ticket: string }>('POST', '/v1/wheel/realtime/ticket')
    socket = new WebSocket(websocketURL(response.ticket))
    socket.onopen = () => { connected.value = true }
    socket.onmessage = (message) => handleSocket(String(message.data))
    socket.onerror = () => { connected.value = false }
    socket.onclose = () => {
      connected.value = false
      socket = null
      if (!stopped && state.value?.session_status === 'active') reconnectTimer = window.setTimeout(() => void connectSocket(), 3500)
    }
  } catch {
    if (!stopped) reconnectTimer = window.setTimeout(() => void connectSocket(), 3500)
  }
}

function handleSocket(raw: string) {
  let payload: { event?: string; data?: unknown }
  try { payload = JSON.parse(raw) as { event?: string; data?: unknown } } catch { return }
  if (payload.event === 'chat.message.created') {
    mergeMessage(payload.data as ChatMessage)
    void scrollChat()
  } else if (payload.event === 'chat.message.hidden' || payload.event === 'chat.message.deleted') {
    const id = Number((payload.data as { id?: number })?.id)
    messages.value = messages.value.filter((item) => item.id !== id)
  } else if (!spinning.value && payload.event?.startsWith('wheel.')) {
    void refreshState()
  }
}

function disconnectSocket() {
  window.clearTimeout(reconnectTimer)
  socket?.close()
  socket = null
  connected.value = false
}

function formatMoney(value: string | number) {
  return `${new Intl.NumberFormat('vi-VN', { maximumFractionDigits: 0 }).format(Number(value) || 0)}đ`
}

function formatChatTime(value: string) {
  return new Intl.DateTimeFormat('vi-VN', { hour: '2-digit', minute: '2-digit' }).format(new Date(value))
}

onMounted(() => {
  document.addEventListener('dblclick', preventDoubleTap, { passive: false })
  clockTimer = window.setInterval(() => {
    nowMs.value = Date.now()
    if (state.value?.session_status === 'active' && secondsLeft.value === 0 && !spinning.value) void refreshState()
  }, 250)
  void bootstrap()
})

onBeforeUnmount(() => {
  stopped = true
  window.clearInterval(clockTimer)
  window.clearTimeout(revealTimer)
  window.clearTimeout(readyTimer)
  document.removeEventListener('dblclick', preventDoubleTap)
  disconnectSocket()
})

function preventDoubleTap(event: MouseEvent) {
  if (event.detail > 1) event.preventDefault()
}
</script>

<template>
  <main class="min-h-dvh bg-[#f5f5f3]">
    <div v-if="booting" class="grid min-h-dvh place-items-center px-6">
      <div class="text-center">
        <img :src="logo" alt="fh88u" class="mx-auto h-16 w-auto object-contain">
        <div class="mx-auto mt-6 h-1 w-36 overflow-hidden bg-slate-200"><div class="h-full w-1/2 animate-pulse bg-event-red" /></div>
        <p class="mt-3 text-xs font-bold text-slate-500">Đang xác thực lời mời...</p>
      </div>
    </div>

    <div v-else-if="error && !state" class="grid min-h-dvh place-items-center px-5">
      <section class="w-full max-w-md border border-red-200 bg-white p-6 text-center shadow-sm">
        <span class="material-symbols-outlined text-4xl text-event-red">link_off</span>
        <h1 class="mt-4 text-lg font-black">Không thể mở sự kiện</h1>
        <p class="mt-2 text-sm leading-6 text-slate-600">{{ error }}</p>
      </section>
    </div>

    <template v-else-if="state">
      <header class="border-b border-black/10 bg-white">
        <div class="mx-auto flex h-16 max-w-6xl items-center justify-between px-4 sm:px-6">
          <img :src="logo" alt="fh88u" class="h-10 w-auto object-contain">
          <div class="text-right">
            <p class="text-[0.65rem] font-bold uppercase text-slate-400">Thời gian còn lại</p>
            <p class="font-mono text-xl font-black text-event-red">{{ countdown }}</p>
          </div>
        </div>
      </header>

      <div class="mx-auto grid max-w-6xl gap-0 border-x border-black/10 bg-white lg:grid-cols-[minmax(0,1fr)_360px]">
        <section class="min-w-0 px-4 py-5 sm:px-7 sm:py-7">
          <div class="flex items-center justify-between border-b border-slate-200 pb-4">
            <div class="min-w-0"><p class="text-[0.68rem] font-bold uppercase text-event-red">Vòng quay đặc biệt</p><h1 class="truncate text-lg font-black">{{ state.campaign_name }}</h1></div>
            <p class="shrink-0 text-sm font-black text-slate-700">{{ progress }}/4 lượt</p>
          </div>

          <div class="mt-5 grid grid-cols-4 gap-2" aria-label="Tiến trình bốn lượt">
            <div v-for="item in state.rounds" :key="item.round_no" class="h-1.5 bg-slate-200" :class="{ '!bg-emerald-500': item.status === 'spun', '!bg-event-gold': item.round_no === state.current_round && isActive }" />
          </div>

          <div class="relative mx-auto mt-8 aspect-square w-[min(78vw,440px)] max-w-full">
            <div class="absolute left-1/2 top-[-11px] z-20 -translate-x-1/2 text-event-red-dark"><span class="material-symbols-outlined text-[3rem] [font-variation-settings:'FILL'_1]">arrow_drop_down</span></div>
            <div class="event-wheel absolute inset-3 rounded-full" :style="wheelStyle">
              <span v-for="(label, index) in wheelLabels" :key="label" class="wheel-label" :style="{ transform: `rotate(${index * 45 + 22.5}deg) translate(42%, -50%)` }">{{ label }}</span>
            </div>
            <div class="absolute left-1/2 top-1/2 z-10 grid h-[25%] w-[25%] -translate-x-1/2 -translate-y-1/2 place-items-center rounded-full border-[5px] border-event-gold bg-white shadow-lg">
              <img :src="logo" alt="" class="w-[72%] object-contain">
            </div>
          </div>

          <div class="mx-auto mt-8 min-h-24 max-w-lg text-center" aria-live="polite">
            <template v-if="spinning || submittingSpin"><p class="text-xs font-bold uppercase text-slate-400">Lượt {{ spinningRoundNo || state.current_round }}</p><h2 class="mt-2 text-xl font-black text-event-red">{{ submittingSpin && !spinning ? 'Đang xác nhận lượt quay...' : 'Vòng quay đang chạy...' }}</h2></template>
            <template v-else-if="result"><p class="text-xs font-bold uppercase text-slate-400">Kết quả lượt {{ result.round_no }}</p><h2 class="mt-2 text-2xl font-black text-event-ink">{{ displayResultLabel(result) }}</h2><p v-if="Number(result.prize_amount) > 0" class="mt-1 text-lg font-black text-emerald-600">+{{ formatMoney(result.prize_amount ?? 0) }}</p></template>
            <template v-else-if="isFinished"><p class="text-xs font-bold uppercase text-slate-400">Phiên đã kết thúc</p><h2 class="mt-2 text-2xl font-black">Tổng thưởng {{ totalReward }}</h2></template>
            <template v-else><p class="text-xs font-bold uppercase text-slate-400">Sẵn sàng</p><h2 class="mt-2 text-xl font-black">Lượt {{ state.current_round }} / 4</h2></template>
          </div>

          <p v-if="error" class="mx-auto mt-2 max-w-lg border border-red-200 bg-red-50 px-3 py-2 text-center text-xs text-red-700">{{ error }}</p>
          <button v-if="isActive || isReadyToStart" type="button" class="mx-auto mt-4 flex min-h-14 w-full max-w-lg items-center justify-center gap-2 rounded-[6px] bg-event-red px-5 text-base font-black text-white shadow-[0_12px_28px_rgba(183,25,32,0.24)] disabled:bg-slate-300 disabled:shadow-none" :disabled="isReadyToStart ? starting : !canSpin" @click="isReadyToStart ? startSession() : spin()" @dblclick.prevent="preventDoubleTap">
            <span class="material-symbols-outlined">{{ isReadyToStart ? 'play_arrow' : 'casino' }}</span>
            {{ starting ? 'Đang bắt đầu...' : submittingSpin ? 'Đang xác nhận...' : spinning ? 'Đang quay...' : isReadyToStart ? 'Bắt đầu vòng quay' : nextReadyIn > 0 ? `Lượt tiếp theo sau ${nextReadyIn}s` : syncingNextRound ? 'Đang đồng bộ lượt tiếp theo...' : `Quay lượt ${state.current_round}` }}
          </button>

          <div class="mt-7 flex items-center justify-between border-t border-slate-200 pt-4 text-sm">
            <span class="text-slate-500">Tổng thưởng đã nhận</span><strong class="text-lg text-emerald-600">{{ totalReward }}</strong>
          </div>
        </section>

        <aside class="flex min-h-[520px] flex-col border-t border-slate-200 bg-[#fafafa] lg:min-h-[calc(100dvh-4rem)] lg:border-l lg:border-t-0">
          <header class="flex h-16 items-center justify-between border-b border-slate-200 px-4">
            <div><h2 class="text-sm font-black">Trò chuyện sự kiện</h2><p class="mt-0.5 text-[0.68rem] text-slate-500">{{ connected ? 'Đang trực tuyến' : isActive ? 'Đang kết nối lại' : 'Đã đóng' }}</p></div>
            <span class="h-2.5 w-2.5 rounded-full" :class="connected ? 'bg-emerald-500' : 'bg-slate-300'" />
          </header>
          <div ref="messageList" class="event-scrollbar min-h-0 flex-1 space-y-3 overflow-y-auto px-4 py-4">
            <p v-if="!messages.length" class="py-10 text-center text-xs text-slate-400">Tin nhắn sẽ xuất hiện tại đây.</p>
            <div v-for="message in messages" :key="message.id" class="flex gap-2.5">
              <div class="grid h-8 w-8 shrink-0 place-items-center rounded-full bg-slate-200 text-xs font-black text-slate-600">{{ message.display_name.slice(0, 1).toUpperCase() }}</div>
              <div class="min-w-0"><div class="flex items-baseline gap-2"><strong class="truncate text-xs text-slate-800">{{ message.display_name }}</strong><time class="text-[0.62rem] text-slate-400">{{ formatChatTime(message.created_at) }}</time></div><p class="mt-1 break-words text-[0.8rem] leading-5 text-slate-600">{{ message.body }}</p></div>
            </div>
          </div>
          <form class="border-t border-slate-200 bg-white p-3" @submit.prevent="sendChat">
            <p v-if="chatError" class="mb-2 text-xs text-red-600">{{ chatError }}</p>
            <div class="flex items-center gap-2">
              <input v-model="chatBody" type="text" maxlength="280" :disabled="!isActive" class="min-h-11 min-w-0 flex-1 border border-slate-300 bg-white px-3 text-sm outline-none focus:border-event-red disabled:bg-slate-100" :placeholder="isActive ? 'Nhập tin nhắn...' : 'Phiên chat đã kết thúc'">
              <button type="submit" class="grid h-11 w-11 shrink-0 place-items-center rounded-[6px] bg-event-red text-white disabled:bg-slate-300" :disabled="!canSend" aria-label="Gửi tin nhắn"><span class="material-symbols-outlined">send</span></button>
            </div>
          </form>
        </aside>
      </div>
    </template>
  </main>
</template>
