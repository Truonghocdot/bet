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
const celebrationRound = ref<Round | null>(null)
const celebrationVisible = ref(false)
let clockTimer = 0
let revealTimer = 0
let reconnectTimer = 0
let readyTimer = 0
let celebrationTimer = 0
let chatPollTimer = 0
let socket: WebSocket | null = null
let stopped = false
let refreshingState = false
let refreshQueued = false

const wheelSegments = [
  { key: 'jackpot_50m', label: '50 TRIỆU', index: 0 },
  { key: 'try_again', label: 'MAY MẮN', index: 1 },
  { key: 'reward_10m', label: '10 TRIỆU', index: 2 },
  { key: 'thank_you', label: 'CẢM ƠN', index: 3 },
  { key: 'reward_5m', label: '5 TRIỆU', index: 4 },
  { key: 'reward_2m', label: '2 TRIỆU', index: 5 },
  { key: 'reward_1m', label: '1 TRIỆU', index: 6 },
  { key: 'reward_500k', label: '500K', index: 7 },
]
const confettiColors = ['#f4bd32', '#f97316', '#ef4444', '#22c55e', '#38bdf8', '#f8fafc']
const confettiPieces = Array.from({ length: 34 }, (_, index) => ({
  left: `${(index * 29) % 101}%`,
  delay: `${(index % 9) * 0.08}s`,
  duration: `${2.4 + (index % 5) * 0.25}s`,
  color: confettiColors[index % confettiColors.length],
  rotate: `${(index * 47) % 360}deg`,
}))
const currentRound = computed(() => state.value?.rounds.find((item) => item.round_no === state.value?.current_round) ?? null)
const isReadyToStart = computed(() => state.value && !state.value.session_id && state.value.session_status === 'pending')
const isActive = computed(() => state.value?.session_status === 'active')
const isFinished = computed(() => state.value?.session_status === 'completed' || state.value?.session_status === 'expired')
const showFinishedScreen = computed(() => isFinished.value && !spinning.value && !submittingSpin.value && !celebrationVisible.value)
const serverNowMs = computed(() => nowMs.value + serverOffsetMs.value)
const secondsLeft = computed(() => state.value?.ends_at ? Math.max(0, Math.ceil((new Date(state.value.ends_at).getTime() - serverNowMs.value) / 1000)) : 300)
const countdown = computed(() => `${String(Math.floor(secondsLeft.value / 60)).padStart(2, '0')}:${String(secondsLeft.value % 60).padStart(2, '0')}`)
const nextReadyIn = computed(() => state.value?.next_round_available_at ? Math.max(0, Math.ceil((new Date(state.value.next_round_available_at).getTime() - serverNowMs.value) / 1000)) : 0)
const canSpin = computed(() => {
  const roundReady = currentRound.value?.status === 'ready'
    || (currentRound.value?.status === 'pending' && Boolean(state.value?.next_round_available_at) && nextReadyIn.value === 0)
  return isActive.value && roundReady && !spinning.value && !submittingSpin.value
})
const chatOpen = computed(() => Boolean(isReadyToStart.value || isActive.value))
const canSend = computed(() => chatOpen.value && chatBody.value.trim().length > 0 && chatBody.value.length <= 280 && !sending.value)
const progress = computed(() => state.value?.rounds.filter((item) => item.status === 'spun').length ?? 0)
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
  const nextRound = next.rounds.find((item) => item.round_no === next.current_round)
  if (next.session_status === 'active' && nextRound?.status === 'pending') {
    scheduleReadyRefresh(next.next_round_available_at)
  } else {
    window.clearTimeout(readyTimer)
  }
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
  if (response.session_status === 'completed' || response.session_status === 'expired') {
    disconnectSocket()
    return
  }
  if (response.session_id) {
    recoverAnimation(response)
    // Chat and realtime are enhancements; they must not block the wheel from
    // rendering while a socket handshake or a slow history query is pending.
    void nextTick().then(() => {
      void loadChat()
      void connectSocket()
    })
  } else {
    // The invitation room is created before the session starts. Load its
    // bot history immediately, then poll while the user is on the landing
    // state so chat does not appear to require clicking "Bắt đầu" first.
    void loadChat()
    void connectSocket()
    schedulePendingChatPoll()
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
    window.clearTimeout(chatPollTimer)
    disconnectSocket()
    applyState(response)
    void nextTick().then(() => {
      void loadChat()
      void connectSocket()
    })
  } catch (cause) {
    error.value = (cause as ApiError).message || 'Không thể bắt đầu phiên.'
  } finally {
    starting.value = false
  }
}

function segmentIndex(round: Round) {
  const key = String(round.segment_key ?? '').trim().toLocaleLowerCase('vi-VN')
  const label = String(round.result_label ?? '').trim().toLocaleLowerCase('vi-VN')
  const amount = Number(round.prize_amount ?? 0)
  const known: Record<string, number> = { jackpot_50m: 0, try_again: 1, reward_10m: 2, thank_you: 3, reward_5m: 4, reward_2m: 5, reward_1m: 6, reward_500k: 7 }
  if (key in known) return known[key]!

  // Campaigns created before the fixed wheel palette may contain a custom key.
  // Resolve the visual segment from the snapshotted prize before falling back
  // to a deterministic key, so the wheel never animates to another prize.
  const byAmount: Record<number, number> = { 50000000: 0, 10000000: 2, 5000000: 4, 2000000: 5, 1000000: 6, 500000: 7 }
  if (byAmount[amount] !== undefined) return byAmount[amount]!
  if (key.includes('50') || label.includes('50 triệu')) return 0
  if (key.includes('10') || label.includes('10 triệu')) return 2
  if (key.includes('5') || label.includes('5 triệu')) return 4
  if (key.includes('2') || label.includes('2 triệu')) return 5
  if (key.includes('1') || label.includes('1 triệu')) return 6
  if (key.includes('500') || label.includes('500')) return 7
  if (!key) return label.includes('cảm ơn') ? 3 : 1
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
  const amountCanonical: Record<number, string> = { 50000000: '50 triệu đồng', 10000000: '10 triệu đồng', 5000000: '5 triệu đồng', 2000000: '2 triệu đồng', 1000000: '1 triệu đồng', 500000: '500.000 đồng' }
  // The amount is the source of truth for a paid result. This also repairs
  // old snapshots where the label was accidentally left as "Chúc bạn may mắn".
  if (amountCanonical[amount] !== undefined) return amountCanonical[amount]!
  if (amount > 0 && canonical[round.segment_key ?? ''] && ['chúc bạn may mắn', 'cảm ơn bạn đã tham gia'].includes(label.toLocaleLowerCase('vi-VN'))) {
    return canonical[round.segment_key ?? '']
  }
  return label || canonical[round.segment_key ?? ''] || 'Kết quả lượt quay'
}

function spinToRound(round: Round, durationMs = 5000) {
  const index = segmentIndex(round)
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
    if (Number(round.prize_amount ?? 0) > 0) openCelebration(round)
    if (state.value?.session_status === 'completed') disconnectSocket()
    else {
      refreshQueued = false
      void refreshState()
    }
  }, Math.max(100, durationMs))
}

function openCelebration(round: Round) {
  window.clearTimeout(celebrationTimer)
  celebrationRound.value = round
  celebrationVisible.value = true
  celebrationTimer = window.setTimeout(closeCelebration, 5500)
}

function closeCelebration() {
  celebrationVisible.value = false
  window.clearTimeout(celebrationTimer)
  celebrationTimer = 0
}

function scheduleReadyRefresh(availableAt?: string | null) {
  window.clearTimeout(readyTimer)
  if (stopped || state.value?.session_status !== 'active') return
  const remaining = availableAt ? new Date(availableAt).getTime() - serverNowMs.value + 150 : 750
  const delay = Math.min(10000, Math.max(750, remaining))
  readyTimer = window.setTimeout(() => void refreshState(), delay)
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
  if (!accessToken.value) return
  if (spinning.value) {
    refreshQueued = true
    return
  }
  if (refreshingState) {
    refreshQueued = true
    return
  }
  refreshingState = true
  try {
    const previousSessionID = state.value?.session_id
    const response = await api<WheelState>('GET', '/v1/wheel/session/state')
    applyState(response)
    if (!previousSessionID && response.session_id) {
      window.clearTimeout(chatPollTimer)
      disconnectSocket()
      void connectSocket()
    }
  } catch {
    if (state.value?.session_status === 'active') scheduleReadyRefresh()
  } finally {
    refreshingState = false
    if (refreshQueued) {
      refreshQueued = false
      window.clearTimeout(readyTimer)
      readyTimer = window.setTimeout(() => void refreshState(), 250)
      return
    }
    const round = state.value?.rounds.find((item) => item.round_no === state.value?.current_round)
    if (state.value?.session_status === 'active' && round?.status === 'pending') {
      scheduleReadyRefresh(state.value.next_round_available_at)
    }
  }
}

async function loadChat() {
  if (!state.value) return
  try {
    const response = await api<{ items: ChatMessage[] }>('GET', '/v1/wheel/session/chat/messages?limit=60')
    messages.value = response.items ?? []
    await scrollChat()
  } catch { /* Chat remains optional to the wheel flow. */ }
}

function schedulePendingChatPoll() {
  window.clearTimeout(chatPollTimer)
  if (stopped || connected.value || state.value?.session_id || state.value?.session_status !== 'pending') return
  chatPollTimer = window.setTimeout(() => {
    void loadChat().finally(schedulePendingChatPoll)
  }, 2000)
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
  if (stopped || socket || !chatOpen.value) return
  try {
    const response = await api<{ ticket: string }>('POST', '/v1/wheel/realtime/ticket')
    socket = new WebSocket(websocketURL(response.ticket))
    socket.onopen = () => {
      connected.value = true
      window.clearTimeout(chatPollTimer)
      // Close the REST/WebSocket handoff gap so the opening bot burst cannot be missed.
      chatPollTimer = window.setTimeout(() => void loadChat(), 750)
    }
    socket.onmessage = (message) => handleSocket(String(message.data))
    socket.onerror = () => { connected.value = false }
    socket.onclose = () => {
      connected.value = false
      socket = null
      schedulePendingChatPoll()
      if (!stopped && chatOpen.value) reconnectTimer = window.setTimeout(() => void connectSocket(), 3500)
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
  } else if (payload.event?.startsWith('wheel.')) {
    if (spinning.value) refreshQueued = true
    else void refreshState()
  }
}

function disconnectSocket() {
  window.clearTimeout(reconnectTimer)
  const activeSocket = socket
  socket = null
  if (activeSocket) {
    activeSocket.onclose = null
    activeSocket.close()
  }
  connected.value = false
}

function formatMoney(value: string | number) {
  return `${new Intl.NumberFormat('vi-VN', { maximumFractionDigits: 0 }).format(Number(value) || 0)}đ`
}

function formatChatTime(value: string) {
  return new Intl.DateTimeFormat('vi-VN', { hour: '2-digit', minute: '2-digit', timeZone: 'Asia/Ho_Chi_Minh' }).format(new Date(value))
}

function chatGameID(message: ChatMessage) {
  const numeric = message.display_name.match(/[0-9]{3,12}/)?.[0]
  if (numeric) return `ID game #${Number(numeric)}`

  let hash = 2166136261
  const source = `${message.actor_type}:${message.display_name}`
  for (let index = 0; index < source.length; index += 1) {
    hash ^= source.charCodeAt(index)
    hash = Math.imul(hash, 16777619)
  }
  return `ID game #${100000 + ((hash >>> 0) % 900000)}`
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
  window.clearTimeout(chatPollTimer)
  document.removeEventListener('dblclick', preventDoubleTap)
  disconnectSocket()
})

function preventDoubleTap(event: MouseEvent) {
  if (event.detail > 1) event.preventDefault()
}
</script>

<template>
  <main class="event-shell min-h-dvh">
    <div v-if="booting" class="event-loading">
      <div class="event-loading__mark"><span class="material-symbols-outlined">casino</span></div>
      <div class="event-loading__bar"><span /></div>
      <p>Đang mở phòng sự kiện...</p>
    </div>

    <div v-else-if="error && !state" class="event-error">
      <section class="event-error__card" role="alert">
        <span class="material-symbols-outlined">link_off</span>
        <h1>Không thể mở sự kiện</h1>
        <p>{{ error }}</p>
        <button type="button" class="event-button event-button--ghost" @click="bootstrap">Thử lại</button>
      </section>
    </div>

    <template v-else-if="state">
      <section v-if="showFinishedScreen" class="event-finished" role="status" aria-live="polite">
        <h1>KẾT THÚC</h1>
      </section>

      <template v-else>
      <header class="event-header">
        <div class="event-container event-header__inner">
          <div class="event-brand">
            <div class="event-brand__logo"><img :src="logo" alt="fh88u"></div>
            <div><p class="event-eyebrow">FH88U PRESENTS</p><p class="event-brand__title">{{ state.campaign_name }}</p></div>
          </div>
          <div class="event-countdown" :class="{ 'event-countdown--urgent': secondsLeft <= 30 && isActive }">
            <span class="material-symbols-outlined">timer</span>
            <div><span>THỜI GIAN CÒN LẠI</span><strong>{{ countdown }}</strong></div>
          </div>
        </div>
      </header>

      <div class="event-container event-layout">
        <section class="event-stage">
          <div class="event-stage__heading">
            <div><p class="event-eyebrow event-eyebrow--gold">LUCKY WHEEL LIVE</p><h1>Vòng quay may mắn</h1></div>
            <div class="event-round-badge"><span>{{ progress }}</span><small>/ 4 LƯỢT</small></div>
          </div>

          <div class="event-progress" aria-label="Tiến trình vòng quay">
            <div v-for="item in state.rounds" :key="item.round_no" class="event-progress__item" :class="{ 'is-complete': item.status === 'spun', 'is-current': item.round_no === state.current_round && isActive }"><span>LƯỢT {{ item.round_no }}</span></div>
          </div>

          <div class="event-wheel-area">
            <div class="event-wheel-area__halo" />
            <div class="event-pointer"><span class="material-symbols-outlined">arrow_drop_down</span></div>
            <div class="event-wheel" :style="wheelStyle" :class="{ 'is-spinning': spinning }" aria-label="Vòng quay giải thưởng">
              <span v-for="segment in wheelSegments" :key="segment.key" class="wheel-label" :style="{ '--segment-angle': `${segment.index * 45 + 22.5}deg`, '--counter-angle': `-${segment.index * 45 + 22.5}deg` }"><span class="wheel-label__text">{{ segment.label }}</span></span>
              <div class="event-wheel__ring" />
            </div>
            <div class="event-wheel__hub"><img :src="logo" alt="fh88u"><span class="material-symbols-outlined">auto_awesome</span></div>
          </div>

          <div class="event-result" aria-live="polite">
            <template v-if="spinning || submittingSpin">
              <span class="event-result__icon event-result__icon--spin"><span class="material-symbols-outlined">sync</span></span>
              <div class="event-result__copy"><p>LƯỢT {{ spinningRoundNo || state.current_round }}</p><h2>{{ submittingSpin && !spinning ? 'Đang xác nhận...' : 'Đang quay giải...' }}</h2></div>
            </template>
            <template v-else-if="result">
              <span class="event-result__icon" :class="Number(result.prize_amount) > 0 ? 'event-result__icon--win' : 'event-result__icon--neutral'"><span class="material-symbols-outlined">{{ Number(result.prize_amount) > 0 ? 'workspace_premium' : 'sentiment_satisfied' }}</span></span>
              <div class="event-result__copy"><p>KẾT QUẢ LƯỢT {{ result.round_no }}</p><h2>{{ displayResultLabel(result) }}</h2><strong v-if="Number(result.prize_amount) > 0">+{{ formatMoney(result.prize_amount ?? 0) }}</strong></div>
            </template>
            <template v-else><span class="event-result__icon event-result__icon--neutral"><span class="material-symbols-outlined">touch_app</span></span><div class="event-result__copy"><p>SẴN SÀNG</p><h2>{{ isReadyToStart ? 'Bắt đầu để mở lượt 1' : `Lượt ${state.current_round} sẵn sàng` }}</h2></div></template>
          </div>

          <p v-if="error" class="event-inline-error" role="alert"><span class="material-symbols-outlined">error</span>{{ error }}</p>
          <button v-if="isActive || isReadyToStart" type="button" class="event-button event-button--primary" :disabled="isReadyToStart ? starting : !canSpin" @click="isReadyToStart ? startSession() : spin()" @dblclick.prevent="preventDoubleTap">
            <span class="material-symbols-outlined">{{ isReadyToStart ? 'play_arrow' : 'casino' }}</span>
            {{ starting ? 'Đang bắt đầu...' : submittingSpin ? 'Đang xác nhận...' : spinning ? 'Đang quay...' : isReadyToStart ? 'Bắt đầu vòng quay' : nextReadyIn > 0 ? `Lượt tiếp theo sau ${nextReadyIn}s` : `Quay lượt ${state.current_round}` }}
          </button>
          <p class="event-stage__hint"><span class="material-symbols-outlined">lock</span>Kết quả được xác nhận an toàn từ máy chủ</p>
        </section>

        <aside class="event-chat">
          <header class="event-chat__header"><div><p class="event-eyebrow event-eyebrow--gold">LIVE ROOM</p><h2>Phòng trò chuyện</h2></div><span class="event-chat__status" :class="{ 'is-online': connected }"><i />{{ connected ? 'Trực tuyến' : chatOpen ? 'Đang kết nối' : 'Đã đóng' }}</span></header>
          <div ref="messageList" class="event-chat__messages">
            <div v-if="!messages.length" class="event-chat__empty"><span class="material-symbols-outlined">forum</span><p>Phòng chat đang chờ những lời chúc đầu tiên.</p></div>
            <div v-for="message in messages" :key="message.id" class="event-message"><div class="event-message__avatar">#</div><div class="event-message__body"><div><strong>{{ chatGameID(message) }}</strong><time>{{ formatChatTime(message.created_at) }}</time></div><p>{{ message.body }}</p></div></div>
          </div>
          <form class="event-chat__composer" @submit.prevent="sendChat"><p v-if="chatError" class="event-chat__error">{{ chatError }}</p><div class="event-chat__composer-row"><input v-model="chatBody" type="text" maxlength="280" :disabled="!chatOpen" :placeholder="chatOpen ? 'Gửi lời chúc...' : 'Phòng chat đã đóng'"><button type="submit" :disabled="!canSend" aria-label="Gửi tin nhắn"><span class="material-symbols-outlined">arrow_upward</span></button></div></form>
        </aside>
      </div>

      <div v-if="celebrationVisible && celebrationRound" class="celebration-backdrop" role="dialog" aria-modal="true" aria-labelledby="celebration-title" @click.self="closeCelebration">
        <span v-for="(piece, index) in confettiPieces" :key="index" class="confetti-piece" :style="{ left: piece.left, animationDelay: piece.delay, animationDuration: piece.duration, backgroundColor: piece.color, transform: `rotate(${piece.rotate})` }" />
        <section class="celebration-card">
          <button type="button" class="celebration-card__close" aria-label="Đóng thông báo" @click="closeCelebration"><span class="material-symbols-outlined">close</span></button>
          <div class="celebration-card__burst"><span class="material-symbols-outlined">workspace_premium</span></div>
          <p class="event-eyebrow event-eyebrow--gold">CHÚC MỪNG BẠN</p>
          <h2 id="celebration-title">Bạn vừa trúng thưởng!</h2>
          <p class="celebration-card__label">{{ displayResultLabel(celebrationRound) }}</p>
          <strong>+{{ formatMoney(celebrationRound.prize_amount ?? 0) }}</strong>
          <p class="celebration-card__note"><span class="material-symbols-outlined">account_balance_wallet</span> Phần thưởng đã được cộng vào ví</p>
          <button type="button" class="event-button event-button--primary" @click="closeCelebration"><span class="material-symbols-outlined">arrow_forward</span>{{ state.session_status === 'completed' ? 'Kết thúc' : 'Tiếp tục vòng quay' }}</button>
        </section>
      </div>
      </template>
    </template>
  </main>
</template>
