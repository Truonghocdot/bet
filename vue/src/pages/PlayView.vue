<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { env } from '@/shared/config/env'
import { request, type ApiError } from '@/shared/api/http'
import type {
  PlayRoomBetHistoryResponse,
  PlayRoomBetResponse,
  PlayRoomHistoryResponse,
  PlayRoomStateResponse,
  GameJoinResponse,
} from '@/shared/api/types'
import { getPlayRoom, type PlayBetGroup, type PlayRoom, type PlayVariant } from '@/data/play'
import { formatViMoney } from '@/shared/lib/money'
import { useAuthStore } from '@/stores/auth'
import { useWalletStore } from '@/stores/wallet'
import { useLoading } from '@/shared/lib/loading'
import PlayHistorySection from '@/components/PlayHistorySection.vue'

const { setLoading } = useLoading()

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const wallet = useWalletStore()

const activeVariantCode = ref('')
const activeHistoryTab = ref<'history' | 'chart' | 'mine'>('history')
const activeK3SubTab = ref('Tổng số')
const connectionId = ref('')
const joinLoading = ref(false)
const joinError = ref('')
const betMessage = ref('')
const betMessageRoomCode = ref('')
const betLoading = ref(false)
const selectedOptions = reactive<Record<string, string>>({})
const roomState = ref<PlayRoomStateResponse | null>(null)
const historyRows = ref<PlayRoomHistoryResponse['items']>([])
const mineRows = ref<PlayRoomBetHistoryResponse['items']>([])
const chartRows = ref<PlayRoomHistoryResponse['items']>([])
const historyPage = ref(1)
const historyTotalPages = ref(1)
const minePage = ref(1)
const mineTotalPages = ref(1)
const historyLoading = ref(false)
const mineLoading = ref(false)
const chartLoading = ref(false)
const historyError = ref('')
const mineError = ref('')
const chartError = ref('')
const roomStateLoading = ref(false)
const roomStateError = ref('')
const serverClockAnchorMs = ref(0)
const localClockAnchorMs = ref(0)
const clockTick = ref(Date.now())
const pendingWalletDebit = ref(0)
const lastSettledPeriodId = ref<number | null>(null)
const seenSettlementPeriods = new Set<string>()
const settlementTargets = new Map<string, string>()
const countdownTargetMs = ref(0)
const countdownTargetPeriodNo = ref('')
const stableCountdownPeriodKey = ref('')
const stableRemainingSeconds = ref(0)
const roomStateCachePrefix = 'ff789:play-room-state:'
const settlementHandledCachePrefix = 'ff789:play-settlement-handled:'
let roomStateGeneration = 0

// Bet modal state
const showBetModal = ref(false)
const modalBetLabel = ref('')
const modalBetKey = ref('')
const modalBetGroupTitle = ref('')
const initialBetAmount = 10000
const minimumBetAmount = 1000
const modalBetAmount = ref(initialBetAmount)
const modalBetStepAmount = ref(minimumBetAmount)
const betAmountPresets = [1000, 10000, 50000, 100000, 500000, 1000000, 5000000, 10000000]

// Countdown beep for the last seconds before bet lock.
let tickAudioContext: AudioContext | null = null
let tickGainNode: GainNode | null = null

function playTickSound() {
  if (typeof window === 'undefined') return
  const AudioCtx = window.AudioContext || (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
  if (!AudioCtx) return

  if (!tickAudioContext) {
    tickAudioContext = new AudioCtx()
    tickGainNode = tickAudioContext.createGain()
    tickGainNode.gain.value = 0.06
    tickGainNode.connect(tickAudioContext.destination)
  }

  if (tickAudioContext.state === 'suspended') {
    void tickAudioContext.resume().catch(() => {
      // ignore resume errors
    })
  }

  const oscillator = tickAudioContext.createOscillator()
  oscillator.type = 'square'
  oscillator.frequency.setValueAtTime(880, tickAudioContext.currentTime)
  oscillator.connect(tickGainNode as GainNode)
  oscillator.start()
  oscillator.stop(tickAudioContext.currentTime + 0.08)
}

// Result modal state
const showResultModal = ref(false)
const resultModalPeriodNo = ref('')
const resultModalTitle = ref('')
const resultModalDescription = ref('')
const resultModalAmount = ref('')
const resultModalTone = ref<'win' | 'lose' | 'draw'>('draw')
const resultModalSettledAt = ref('')
const resultModalStake = ref('')
const resultModalPayout = ref('')
const showTicketDetailModal = ref(false)
const selectedTicketDetail = ref<PlayRoomBetHistoryResponse['items'][number] | null>(null)
const showHelpModal = ref(false)

const gameHowToPlay = {
  wingo: {
    title: 'Hướng dẫn chơi Win Go',
    content: `
      ● Win Go là trò chơi dự đoán kết quả dựa trên các con số từ 0-9 và màu sắc tương ứng.
      ● Các loại cược chính:
        1. Cược Màu:
           - Xanh: Các số 1, 3, 7, 9. Nếu ra 5 là thắng nửa tiền. Tỷ lệ 1:2.
           - Đỏ: Các số 2, 4, 6, 8. Nếu ra 0 là thắng nửa tiền. Tỷ lệ 1:2.
           - Tím: Các số 0 và 5. Tỷ lệ 1:4.5.
        2. Cược Số: Dự đoán chính xác 1 số từ 0-9. Tỷ lệ 1:9.
        3. Cược Lớn/Nhỏ:
           - Lớn: 5, 6, 7, 8, 9.
           - Nhỏ: 0, 1, 2, 3, 4.
           Tỷ lệ 1:2.
      ● Thời gian mỗi kỳ: 30 Giây, 1 Phút, 3 Phút, 5 Phút. Đóng lệnh trước khi kết thúc 5 giây.
    `,
  },
  k3: {
    title: 'Hướng dẫn chơi K3 Lotre',
    content: `
      ● K3 Lotre dựa trên kết quả của 3 quân xúc xắc (xí ngầu) được đổ ngẫu nhiên.
      ● Các loại cược phổ biến:
        1. Cược Tổng: Dự đoán tổng điểm của 3 quân xí ngầu (từ 3 đến 18). Mỗi con số có tỷ lệ thưởng khác nhau, lên đến 1:207.
        2. Cược 2 số trùng: Dự đoán có ít nhất 2 xí ngầu ra cùng một mặt.
        3. Cược 3 số trùng: Dự đoán cả 3 xí ngầu ra cùng một mặt (ví dụ: 1-1-1). Tỷ lệ thưởng cực lớn.
        4. Cược Lớn/Nhỏ/Chẵn/Lẻ: Dựa trên Tổng số điểm.
           - Lớn (11-18), Nhỏ (3-10).
           - Chẵn (Tổng là số chẵn), Lẻ (Tổng là số lẻ).
      ● Cách tính kết quả: Tổng điểm = Xí ngầu 1 + Xí ngầu 2 + Xí ngầu 3.
    `,
  },
  lottery: {
    title: 'Hướng dẫn chơi 5D Lô tô',
    content: `
      ● 5D Lô tô dựa trên chuỗi 5 con số được quay ngẫu nhiên cho các vị trí A, B, C, D, E.
      ● Các hình thức đặt cược:
        1. Cược Vị trí: Dự đoán số cụ thể tại 1 trong 5 vị trí (A, B, C, D hoặc E).
        2. Cược Tổng hợp: Dự đoán Tổng của cả 5 con số là Lớn/Nhỏ hoặc Chẵn/Lẻ.
        3. Cược Số đuôi: Dự đoán số cuối cùng (vị trí E) hoặc các tổ hợp số đặc biệt.
      ● Vị trí tương ứng:
        - A: Số thứ nhất (hàng vạn)
        - B: Số thứ hai (hàng nghìn)
        - C: Số thứ ba (hàng trăm)
        - D: Số thứ tư (hàng chục)
        - E: Số thứ năm (hàng đơn vị)
      ● Tỷ lệ thưởng: Tùy thuộc vào độ khó của cửa đặt, cược vị trí chính xác có tỷ lệ cao nhất.
    `,
  },
} as const

// Dice animation state for K3
const isDiceRolling = ref(false)
const rollingDice = ref([1, 2, 3])
let rollTimer: number | undefined

const tablePageSize = 4
let timer: number | undefined
let roomStreamConnection: WebSocket | null = null
let betsStreamConnection: WebSocket | null = null
let roomStreamReconnectTimer: number | undefined
let roomStateReconcileTimer: number | undefined
let periodRollForwardTimer: number | undefined
let betsStreamHistoryRequested = false
const mineRowsCached: PlayRoomBetHistoryResponse['items'] = []
let mineHistoryRestFallbackAt = 0
let autoJoinTimer: number | undefined
let autoJoinAttemptKey = ''
const pendingBetRequests = new Map<string, { resolve: (response: any) => void; reject: (error: any) => void; timeout: number }>()
const pendingRoomHistoryRequests = new Map<string, { resolve: (response: PlayRoomHistoryResponse) => void; reject: (error: Error) => void; timeout: number }>()

const room = computed<PlayRoom | null>(() => getPlayRoom(String(route.params.game ?? 'wingo')) ?? null)
const isK3 = computed(() => room.value?.code === 'k3')
const isWingo = computed(() => room.value?.code === 'wingo')

const selectedVariant = computed<PlayVariant | null>(() => {
  if (!room.value || room.value.variants.length === 0) return null
  return room.value.variants.find((variant) => variant.code === activeVariantCode.value) ?? room.value.variants[0] ?? null
})
const selectedRoomCode = computed(() => {
  if (!room.value || !selectedVariant.value) return ''
  return `${room.value.code}_${selectedVariant.value.code}`
})

const backTarget = computed(() => {
  const from = typeof route.query.from === 'string' ? route.query.from.trim() : ''
  const cached = typeof sessionStorage !== 'undefined' ? sessionStorage.getItem('ff789:last-route') ?? '' : ''
  const candidate = from || cached || '/play'
  if (!candidate || candidate === '/') return '/'
  if (candidate === '/auth' || candidate.startsWith('/auth?')) return '/'
  return candidate
})

function navigateBack() {
  const target = backTarget.value
  const hasRealHistory = typeof window !== 'undefined' && window.history.length > 1

  if (route.query.from || sessionStorage.getItem('ff789:last-route')) {
    void router.push(target)
    return
  }

  if (hasRealHistory) {
    void router.back()
    return
  }

  void router.push(target)
}

const walletVnd = computed(() => wallet.wallets.find((item) => item.unit === 1) ?? null)
const availableVndBalance = computed(() => {
  return Number(walletVnd.value?.balance ?? 0)
})
const canPlay = computed(() => availableVndBalance.value > 0)
const currentPeriod = computed(() => roomState.value?.current_period ?? null)
const syncedNow = computed(() => {
  if (serverClockAnchorMs.value > 0 && localClockAnchorMs.value > 0) {
    return serverClockAnchorMs.value + Math.max(0, clockTick.value - localClockAnchorMs.value)
  }
  return clockTick.value
})
const expectedPeriodSeconds = computed(() => selectedVariant.value?.countdownSeconds ?? 0)
const currentPeriodBetLockAtMs = computed(() => parsePeriodTimeMs(currentPeriod.value?.bet_lock_at))
const activeCountdownTargetMs = computed(() => {
  const currentPeriodNo = String(currentPeriod.value?.period_no ?? '')
  if (!currentPeriodNo) return 0
  const baseTargetMs = countdownTargetPeriodNo.value === currentPeriodNo && countdownTargetMs.value > 0
    ? countdownTargetMs.value
    : parsePeriodTimeMs(currentPeriod.value?.draw_at)
  if (!Number.isFinite(baseTargetMs) || baseTargetMs <= 0) return 0
  return baseTargetMs
})
const visibleBetMessage = computed(() => {
  if (!betMessage.value) return ''
  if (!selectedRoomCode.value) return ''
  return betMessageRoomCode.value === selectedRoomCode.value ? betMessage.value : ''
})

function resetTransientRoomUiState() {
  joinError.value = ''
  betMessage.value = ''
  betMessageRoomCode.value = ''
  pendingWalletDebit.value = 0
  connectionId.value = ''
  autoJoinAttemptKey = ''
  if (autoJoinTimer) {
    window.clearTimeout(autoJoinTimer)
    autoJoinTimer = undefined
  }
}

function resetRoomStateSession() {
  roomState.value = null
  countdownTargetMs.value = 0
  countdownTargetPeriodNo.value = ''
  stableCountdownPeriodKey.value = ''
  stableRemainingSeconds.value = 0
  serverClockAnchorMs.value = 0
  localClockAnchorMs.value = 0
  if (periodRollForwardTimer) {
    window.clearTimeout(periodRollForwardTimer)
    periodRollForwardTimer = undefined
  }
  seenSettlementPeriods.clear()
  settlementTargets.clear()
}

function settlementHandledCacheKey(roomCode: string) {
  return `${settlementHandledCachePrefix}${roomCode}`
}

function hydrateHandledSettlementPeriods(roomCode: string) {
  seenSettlementPeriods.clear()
  if (!roomCode) return
  try {
    const raw = sessionStorage.getItem(settlementHandledCacheKey(roomCode))
    if (!raw) return
    const parsed = JSON.parse(raw) as string[]
    if (!Array.isArray(parsed)) return
    parsed.forEach((item) => {
      if (typeof item === 'string' && item.trim()) {
        seenSettlementPeriods.add(item)
      }
    })
  } catch {
    // ignore storage failures
  }
}

function persistHandledSettlementPeriods(roomCode: string) {
  if (!roomCode) return
  try {
    const values = [...seenSettlementPeriods].filter((item) => item.startsWith(`${roomCode}:`))
    sessionStorage.setItem(settlementHandledCacheKey(roomCode), JSON.stringify(values))
  } catch {
    // ignore storage failures
  }
}

function settlementKeyFor(roomCode: string, periodNo: string | null | undefined) {
  return `${roomCode}:${String(periodNo ?? '')}`
}

function markSettlementHandled(roomCode: string, periodNo: string | null | undefined) {
  if (!roomCode || !String(periodNo ?? '').trim()) return
  seenSettlementPeriods.add(settlementKeyFor(roomCode, periodNo))
  persistHandledSettlementPeriods(roomCode)
}

function logRealtimeEvent(event: string, payload: Record<string, unknown> = {}) {
  console.debug(`[play.sync] ${event}`, payload)
}

function roomStateCacheKey(roomCode: string) {
  return `${roomStateCachePrefix}${roomCode}`
}

type RoomStateCacheSnapshot = {
  savedAt: number
  response: PlayRoomStateResponse
}

function loadRoomStateCache(roomCode: string): RoomStateCacheSnapshot | null {
  if (!roomCode) return null
  try {
    const raw = sessionStorage.getItem(roomStateCacheKey(roomCode))
    if (!raw) return null
    const parsed = JSON.parse(raw) as Partial<RoomStateCacheSnapshot>
    const savedAt = Number(parsed?.savedAt ?? 0)
    const response = parsed?.response as PlayRoomStateResponse | undefined
    if (!response || !response.current_period || !response.server_time || !Number.isFinite(savedAt) || savedAt <= 0) return null
    return { savedAt, response }
  } catch {
    return null
  }
}

function saveRoomStateCache(roomCode: string, response: PlayRoomStateResponse) {
  if (!roomCode) return
  try {
    const snapshot: RoomStateCacheSnapshot = {
      savedAt: Date.now(),
      response,
    }
    sessionStorage.setItem(roomStateCacheKey(roomCode), JSON.stringify(snapshot))
  } catch {
    // ignore storage failures
  }
}

function scheduleRoomStateReconcile(roomCode: string, delayMs = 800) {
  if (!roomCode) return
  if (roomStateReconcileTimer) {
    window.clearTimeout(roomStateReconcileTimer)
  }
  roomStateReconcileTimer = window.setTimeout(() => {
    roomStateReconcileTimer = undefined
    requestRoomStateSnapshot(roomCode)
  }, delayMs)
}

function hydrateRoomStateFromCache(roomCode: string) {
  const cached = loadRoomStateCache(roomCode)
  if (!cached) return false

  roomState.value = cached.response

  const serverTimeMs = parseServerTimeMs(cached.response.server_time)
  if (Number.isFinite(serverTimeMs) && serverTimeMs > 0) {
    applyServerClock(cached.response.server_time, cached.savedAt)
  }

  syncCountdownTarget(cached.response.current_period, Date.now(), true)
  return true
}

function setPendingWalletDebit(amount: number) {
  pendingWalletDebit.value = Math.max(0, Math.floor(amount))
}

async function refreshRealtimeSlices(options: {
  history?: boolean
  mine?: boolean
  chart?: boolean
  wallet?: boolean
  reason?: string
} = {}) {
  const {
    history = true,
    mine = true,
    chart = true,
    wallet: refreshWallet = false,
    reason = 'unknown',
  } = options

  logRealtimeEvent('refresh.start', {
    roomCode: selectedRoomCode.value,
    history,
    mine,
    chart,
    wallet: refreshWallet,
    reason,
  })

  const tasks: Promise<unknown>[] = []
  if (refreshWallet) tasks.push(Promise.resolve(loadWallet()))
  if (history) tasks.push(Promise.resolve(loadRoomHistory(1)))
  
  // Nếu có kỳ đang chờ hiện kết quả, ưu tiên tải lại lịch sử cá nhân
  const hasPendingSettlements = settlementTargets.size > 0
  if (mine || hasPendingSettlements) {
    // Luôn yêu cầu Server gửi lại lịch sử mới nhất qua WS khi đặt cược hoặc chốt kỳ xong
    if (betsStreamConnection && betsStreamConnection.readyState === WebSocket.OPEN) {
      mineRowsCached.length = 0
      betsStreamConnection.send(JSON.stringify({ action: 'request_history' }))
    }
    tasks.push(loadMineHistory(1, { force: true }))
  }
  if (chart) tasks.push(Promise.resolve(loadChartHistory()))
  await Promise.allSettled(tasks)
}

function resetRoomViewData() {
  historyRows.value = []
  mineRows.value = []
  chartRows.value = []
  mineRowsCached.length = 0
  historyError.value = ''
  mineError.value = ''
  chartError.value = ''
  historyPage.value = 1
  minePage.value = 1
  historyTotalPages.value = 1
  mineTotalPages.value = 1
}

function bumpRoomStateGeneration() {
  roomStateGeneration += 1
  return roomStateGeneration
}

function isCurrentRoomStateGeneration(generation: number, roomCode: string) {
  return generation === roomStateGeneration && roomCode === selectedRoomCode.value
}

function syncCountdownTarget(period: PlayRoomStateResponse['current_period'] | null, nowMs = Date.now(), force = false) {
  if (!period) {
    countdownTargetMs.value = 0
    countdownTargetPeriodNo.value = ''
    return
  }

  const periodNo = String(period.period_no ?? '')
  const rawBetLockAtMs = parsePeriodTimeMs(period.bet_lock_at)
  const rawDrawAtMs = parsePeriodTimeMs(period.draw_at)
  const expectedSeconds = Math.max(1, expectedPeriodSeconds.value || 30)
  const periodMs = expectedSeconds * 1000
  const fallbackTargetMs = nowMs + expectedSeconds * 1000
  const maxReasonableMs = nowMs + Math.max(expectedSeconds * 3, 30) * 1000

  // Play view must match admin timing:
  // countdown runs to draw_at, while bet lock is enforced separately.
  const preferredTargetMs = Number.isFinite(rawDrawAtMs) && rawDrawAtMs > 0
    ? rawDrawAtMs
    : rawBetLockAtMs

  if (!periodNo) {
    countdownTargetMs.value = fallbackTargetMs
    countdownTargetPeriodNo.value = ''
    return
  }

  if (!force && countdownTargetPeriodNo.value === periodNo && countdownTargetMs.value > 0) {
    return
  }

  if (Number.isFinite(preferredTargetMs) && preferredTargetMs > 0 && preferredTargetMs <= maxReasonableMs) {
    countdownTargetMs.value = preferredTargetMs
  } else if (Number.isFinite(preferredTargetMs) && preferredTargetMs > maxReasonableMs) {
    // Nếu backend trả mốc quá xa tương lai (cache/period drift), chuẩn hoá theo chu kỳ room
    // để giữ đồng hồ mượt và không reset cứng về 28/58.
    const delta = preferredTargetMs - nowMs
    const remainder = ((delta % periodMs) + periodMs) % periodMs
    countdownTargetMs.value = nowMs + (remainder === 0 ? periodMs : remainder)
  } else {
    countdownTargetMs.value = fallbackTargetMs
  }
  countdownTargetPeriodNo.value = periodNo
}

function applyHistoryFromRecentResults(page = historyPage.value) {
  const results = roomState.value?.recent_results ?? []
  const totalPages = Math.max(1, Math.ceil(results.length / tablePageSize))
  const normalizedPage = Math.min(Math.max(1, Math.floor(page)), totalPages)
  const start = (normalizedPage - 1) * tablePageSize
  historyRows.value = results.slice(start, start + tablePageSize)
  historyPage.value = normalizedPage
  historyTotalPages.value = totalPages
  historyError.value = ''
}
const isBetLocked = computed(() => {
  if (!currentPeriod.value) return true
  if ((currentPeriod.value.status || '').toUpperCase() !== 'OPEN') return true
  return syncedNow.value >= currentPeriodBetLockAtMs.value
})
const canBet = computed(() => canPlay.value && !isBetLocked.value)

const rawRemainingSeconds = computed(() => {
  const targetMs = activeCountdownTargetMs.value
  if (!Number.isFinite(targetMs) || targetMs <= 0) return 0
  return Math.max(0, Math.ceil((targetMs - syncedNow.value) / 1000))
})
const remainingSeconds = computed(() => stableRemainingSeconds.value)
const currentPeriodKey = computed(() => {
  if (!currentPeriod.value) return ''
  const id = Number(currentPeriod.value.id ?? 0)
  if (Number.isFinite(id) && id > 0) return `id:${id}`
  const periodNo = String(currentPeriod.value.period_no ?? '').trim()
  if (periodNo) return `no:${periodNo}`
  return ''
})

const closingCountdownSeconds = computed(() => {
  const status = String(currentPeriod.value?.status ?? '').toUpperCase()
  if (!['OPEN', 'LOCKED'].includes(status)) return 0
  return remainingSeconds.value > 0 && remainingSeconds.value <= 5 ? remainingSeconds.value : 0
})

const showClosingCountdownOverlay = computed(() => {
  return closingCountdownSeconds.value > 0
})

const countdownParts = computed(() => {
  const total = remainingSeconds.value
  const minutes = Math.floor(total / 60)
  const seconds = total % 60
  return {
    m0: String(Math.floor(minutes / 10)),
    m1: String(minutes % 10),
    s0: String(Math.floor(seconds / 10)),
    s1: String(seconds % 10),
  }
})

const roomStatusLabel = computed(() => {
  const status = (currentPeriod.value?.status || '').toUpperCase()
  if (!currentPeriod.value) return room.value?.status === 'COMING_SOON' ? 'Sắp mở' : 'Chưa cập nhật'
  if (status === 'OPEN') return isBetLocked.value ? 'Đang khóa' : 'Đang mở'
  if (status === 'LOCKED') return 'Đang khóa'
  if (status === 'DRAWN') return 'Đã ra kết quả'
  if (status === 'SETTLED') return 'Đã chốt'
  if (status === 'SCHEDULED') return 'Chưa mở'
  if (status === 'CANCELED') return 'Đã hủy'
  return 'Chưa cập nhật'
})

const currentBalanceLabel = computed(() => formatMoney(availableVndBalance.value))

const recentResults = computed(() => roomState.value?.recent_results?.slice(0, 5) ?? [])
const chartSeries = computed(() =>
  chartRows.value.map((row, index) => ({
    ...row,
    periodNo: row.period_no ?? '',
    index,
    value: chartValue(row),
    label: chartSummaryLabel(row),
    barClass: chartBarClass(row),
  })),
)
const chartMaxValue = computed(() => Math.max(1, ...chartSeries.value.map((row) => row.value)))
// K3 sub-tabs: get unique subTab labels from betGroups
const k3SubTabs = computed(() => {
  if (!isK3.value || !selectedVariant.value) return []
  const tabs: string[] = []
  for (const group of selectedVariant.value.betGroups) {
    if (group.subTab && !tabs.includes(group.subTab)) tabs.push(group.subTab)
  }
  return tabs
})

// K3 active groups filtered by sub-tab
const activeK3Groups = computed(() => {
  if (!isK3.value || !selectedVariant.value) return []
  return selectedVariant.value.betGroups.filter((g) => g.subTab === activeK3SubTab.value)
})

// WinGo specific groups
const colorGroup = computed(() => selectedVariant.value?.betGroups.find((group) => groupTypeKey(group) === 'COLOR') ?? null)
const numberGroup = computed(() => selectedVariant.value?.betGroups.find((group) => groupTypeKey(group) === 'NUMBER') ?? null)
const bigSmallGroup = computed(() => selectedVariant.value?.betGroups.find((group) => groupTypeKey(group) === 'BIG_SMALL') ?? null)

// Dice render for K3
const currentDice = computed<number[]>(() => {
  const recent = roomState.value?.recent_results?.[0]
  if (recent?.result && recent.result.includes('-')) {
    return recent.result.split('-').map(Number).filter((n) => n >= 1 && n <= 6)
  }
  return [4, 2, 3]
})

function diceDots(n: number): Array<{ x: string; y: string }> {
  const positions: Record<number, Array<[string, string]>> = {
    1: [['50%', '50%']],
    2: [['28%', '28%'], ['72%', '72%']],
    3: [['28%', '28%'], ['50%', '50%'], ['72%', '72%']],
    4: [['28%', '28%'], ['72%', '28%'], ['28%', '72%'], ['72%', '72%']],
    5: [['28%', '28%'], ['72%', '28%'], ['50%', '50%'], ['28%', '72%'], ['72%', '72%']],
    6: [['28%', '22%'], ['72%', '22%'], ['28%', '50%'], ['72%', '50%'], ['28%', '78%'], ['72%', '78%']],
  }
  return (positions[n] ?? []).map(([x, y]) => ({ x, y }))
}

function diceColor(n: number): string {
  return n <= 3 ? '#e8404a' : '#10b981'
}

// WinGo number color
function wingoBallBackground(n: number): string {
  const gloss = 'radial-gradient(circle at 28% 26%, rgba(255,255,255,0.96) 0 16%, rgba(255,255,255,0.24) 17%, transparent 28%)'
  const zigZagA = 'repeating-linear-gradient(45deg, rgba(255,255,255,0.16) 0 8px, transparent 8px 16px)'
  const zigZagB = 'repeating-linear-gradient(-45deg, rgba(255,255,255,0.12) 0 8px, transparent 8px 16px)'

  if (n === 0) {
    return `${gloss}, ${zigZagA}, ${zigZagB}, linear-gradient(135deg, #e64545 0%, #ef6b73 38%, #8b5cf6 62%, #6f3de8 100%)`
  }
  if (n === 5) {
    return `${gloss}, ${zigZagA}, ${zigZagB}, linear-gradient(135deg, #24b561 0%, #59d88a 38%, #8b5cf6 62%, #6f3de8 100%)`
  }
  if (n % 2 === 0) {
    return `${gloss}, ${zigZagA}, ${zigZagB}, linear-gradient(135deg, #ff8a92 0%, #e64545 48%, #c92b38 100%)`
  }
  return `${gloss}, ${zigZagA}, ${zigZagB}, linear-gradient(135deg, #73e7a0 0%, #24b561 48%, #149454 100%)`
}

function wingoNumberTextStyle(n: number) {
  if (n === 0) {
    return {
      backgroundImage: 'linear-gradient(135deg, #8b5cf6, #e8404a)',
      WebkitBackgroundClip: 'text',
      backgroundClip: 'text',
      color: 'transparent',
      WebkitTextFillColor: 'transparent',
    } as const
  }
  if (n === 5) {
    return {
      backgroundImage: 'linear-gradient(135deg, #8b5cf6, #24b561)',
      WebkitBackgroundClip: 'text',
      backgroundClip: 'text',
      color: 'transparent',
      WebkitTextFillColor: 'transparent',
    } as const
  }
  return { color: n % 2 === 0 ? '#e64545' : '#24b561' }
}

function formatMoney(value: string | number | null | undefined, fractionDigits = 0) {
  return formatViMoney(value ?? 0, fractionDigits)
}

function parseServerTimeMs(value: string | null | undefined) {
  const raw = String(value ?? '').trim()
  if (!raw) return 0
  const normalized = raw.includes(' ') && !raw.includes('T')
    ? raw.replace(' ', 'T')
    : raw
  const parsed = new Date(normalized).getTime()
  return Number.isFinite(parsed) ? parsed : 0
}

function parsePeriodTimeMs(value: string | null | undefined) {
  const raw = String(value ?? '').trim()
  if (!raw) return 0
  const normalized = raw.includes(' ') && !raw.includes('T')
    ? raw.replace(' ', 'T')
    : raw
  const wallClockMatch = normalized.match(/^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})/)
  const candidate = wallClockMatch?.[1] ?? normalized
  const parsed = new Date(candidate).getTime()
  return Number.isFinite(parsed) ? parsed : 0
}

function formatVietnamTime(value: string | null | undefined, withSeconds = true) {
  const ms = parseServerTimeMs(value)
  if (!ms) return '—'
  const formatter = new Intl.DateTimeFormat('vi-VN', {
    timeZone: 'Asia/Ho_Chi_Minh',
    hour: '2-digit',
    minute: '2-digit',
    ...(withSeconds ? { second: '2-digit' } : {}),
  })
  return formatter.format(new Date(ms))
}

function formatSignedMoney(value: string | number | null | undefined) {
  const numeric = Number(value ?? 0)
  if (Number.isNaN(numeric)) return '0'
  const prefix = numeric > 0 ? '+' : numeric < 0 ? '-' : ''
  return `${prefix}${formatMoney(Math.abs(numeric))}đ`
}

function toFiniteNumber(value: string | number | null | undefined) {
  const parsed = Number(value ?? 0)
  return Number.isFinite(parsed) ? parsed : 0
}

function rowStatusValue(row: PlayRoomBetHistoryResponse['items'][number]) {
  return String(row.status || '').toUpperCase()
}

function rowOriginalAmountValue(row: PlayRoomBetHistoryResponse['items'][number]) {
  const original = toFiniteNumber(row.original_amount)
  if (original > 0) return original
  return toFiniteNumber(row.stake)
}

function rowTaxAmountValue(row: PlayRoomBetHistoryResponse['items'][number]) {
  const tax = toFiniteNumber(row.tax_amount)
  if (tax > 0) return tax
  const original = rowOriginalAmountValue(row)
  if (original <= 0) return 0
  return original * 0.02
}

function rowNetAmountValue(row: PlayRoomBetHistoryResponse['items'][number]) {
  const net = toFiniteNumber(row.net_amount)
  if (net > 0) return net
  return Math.max(0, rowOriginalAmountValue(row) - rowTaxAmountValue(row))
}

function rowWinCreditValue(row: PlayRoomBetHistoryResponse['items'][number]) {
  const actual = toFiniteNumber(row.actual_payout)
  if (actual > 0) return actual
  const profitLoss = toFiniteNumber(row.profit_loss)
  if (rowStatusValue(row) === 'WON' && profitLoss > 0) {
    return profitLoss + rowOriginalAmountValue(row)
  }
  return 0
}

function rowProfitLossValue(row: PlayRoomBetHistoryResponse['items'][number]) {
  const profitLoss = toFiniteNumber(row.profit_loss)
  if (profitLoss !== 0) return profitLoss

  const status = rowStatusValue(row)
  if (status === 'WON') {
    return rowWinCreditValue(row) - rowOriginalAmountValue(row)
  }
  if (status === 'LOST') {
    return -rowOriginalAmountValue(row)
  }
  return 0
}

function normalizeBetLabel(value: string | null | undefined) {
  const raw = String(value ?? '').trim()
  if (!raw || raw === '—') return '—'

  const lower = raw.toLowerCase()
  if (lower.startsWith('number_')) {
    const number = lower.replace('number_', '').trim()
    return number ? `Số ${number}` : 'Số'
  }
  if (lower === 'big') return 'Lớn'
  if (lower === 'small') return 'Nhỏ'
  if (lower === 'odd') return 'Lẻ'
  if (lower === 'even') return 'Chẵn'
  if (lower === 'green') return 'Xanh'
  if (lower === 'red') return 'Đỏ'
  if (lower === 'violet') return 'Tím'
  if (lower === 'red_violet') return 'Đỏ / Tím'
  if (lower === 'green_violet') return 'Xanh / Tím'

  return raw.replaceAll('_', ' ')
}

function rowMainLabel(row: PlayRoomBetHistoryResponse['items'][number]) {
  return normalizeBetLabel(row.big_small || row.color || row.result || '—')
}

function rowSubLabel(row: PlayRoomBetHistoryResponse['items'][number]) {
  const normalizedResult = normalizeBetLabel(row.result)
  if (normalizedResult && normalizedResult !== '—') return normalizedResult
  return 'Lệnh đang chờ kết quả'
}

function rowStatusText(row: PlayRoomBetHistoryResponse['items'][number]) {
  const status = rowStatusValue(row)
  if (status === 'WON') return `Thắng +${formatMoney(rowWinCreditValue(row))}đ`
  if (status === 'LOST') return `Thua ${formatSignedMoney(rowProfitLossValue(row))}`
  if (status === 'PENDING') return 'Đang chờ chốt kỳ'
  return status || 'Đang xử lý'
}

function rowStatusClass(row: PlayRoomBetHistoryResponse['items'][number]) {
  const status = rowStatusValue(row)
  if (status === 'WON') return 'text-[#10b981]'
  if (status === 'LOST') return 'text-slate-400'
  return 'text-amber-500'
}

function currentPage() {
  return activeHistoryTab.value === 'mine' ? minePage.value : historyPage.value
}

async function handleHistoryPageChange(page: number) {
  if (activeHistoryTab.value === 'mine') {
    // All data is in WebSocket cache, just update UI pagination
    minePage.value = page
    updateMineRowsFromCache(page)
    return
  }
  if (activeHistoryTab.value === 'history') {
    historyPage.value = page
    await loadRoomHistory(page)
  }
}

function ensureDefaultSelections(variant: PlayVariant | null) {
  Object.keys(selectedOptions).forEach((key) => delete selectedOptions[key])
  variant?.betGroups.forEach((group) => {
    if (group.options[0]) {
      selectedOptions[group.title] = group.options[0].key
    }
  })
  modalBetAmount.value = initialBetAmount
}

async function loadWallet() {
  if (!auth.isAuthenticated) return
  wallet.connectStream()

  // Wait briefly for the first SSE snapshot when caller needs fresh balance now.
  let attempts = 0
  while (!wallet.wallets.length && attempts < 20) {
    await new Promise((resolve) => setTimeout(resolve, 100))
    attempts += 1
  }
}

function applyServerClock(serverTime: string, requestMidpoint = Date.now()) {
  const serverTimeMs = parseServerTimeMs(serverTime)
  if (!Number.isFinite(serverTimeMs) || serverTimeMs <= 0) return

  const hasAnchor = serverClockAnchorMs.value > 0 && localClockAnchorMs.value > 0
  if (!hasAnchor) {
    serverClockAnchorMs.value = serverTimeMs
    localClockAnchorMs.value = requestMidpoint
    clockTick.value = Date.now()
    return
  }

  const estimated = estimatedServerNowAt(requestMidpoint)
  if (!estimated) {
    serverClockAnchorMs.value = serverTimeMs
    localClockAnchorMs.value = requestMidpoint
    clockTick.value = Date.now()
    return
  }

  const drift = serverTimeMs - estimated
  const absDrift = Math.abs(drift)
  if (absDrift <= 120) {
    return
  }

  if (absDrift < 1200) {
    serverClockAnchorMs.value += drift * 0.28
    localClockAnchorMs.value = requestMidpoint
    clockTick.value = Date.now()
    return
  }

  serverClockAnchorMs.value = serverTimeMs
  localClockAnchorMs.value = requestMidpoint
  clockTick.value = Date.now()
}

function estimatedServerNowAt(localMs: number) {
  if (serverClockAnchorMs.value <= 0 || localClockAnchorMs.value <= 0) return 0
  return serverClockAnchorMs.value + Math.max(0, localMs - localClockAnchorMs.value)
}

async function applyRoomStateResponse(
  response: PlayRoomStateResponse,
  options: {
    requestStartedAt?: number
    requestFinishedAt?: number
    forceRebaseClock?: boolean
    roomCode?: string
    generation?: number
  } = {},
) {
  const responseRoomCode = String(options.roomCode ?? selectedRoomCode.value)
  const responseGeneration = Number.isFinite(Number(options.generation ?? roomStateGeneration))
    ? Number(options.generation ?? roomStateGeneration)
    : roomStateGeneration
  if (!responseRoomCode || !isCurrentRoomStateGeneration(responseGeneration, responseRoomCode)) {
    return
  }

  const previousPeriod = roomState.value?.current_period ?? null
  const previousPeriodNo = previousPeriod?.period_no ?? ''
  const nextPeriodNo = response.current_period?.period_no ?? ''
  const shouldRebaseClock = options.forceRebaseClock || !roomState.value || previousPeriodNo !== nextPeriodNo

  roomState.value = response
  chartRows.value = response.recent_results.slice(0, 24)
  chartError.value = ''
  saveRoomStateCache(selectedRoomCode.value, response)

  // Luôn tính midpoint chính xác dù có rebase hay không
  const midpoint =
    options.requestStartedAt !== undefined && options.requestFinishedAt !== undefined
      ? options.requestStartedAt + Math.max(0, Math.floor((options.requestFinishedAt - options.requestStartedAt) / 2))
      : Date.now()

  if (shouldRebaseClock) {
    applyServerClock(response.server_time, midpoint)
  } else {
    // Kể cả khi không rebase (WS update), vẫn resync nếu lệch > 1s
    const serverMs = parseServerTimeMs(response.server_time)
    if (Number.isFinite(serverMs) && serverMs > 0) {
      const estimatedServerMs = estimatedServerNowAt(midpoint)
      const drift = estimatedServerMs > 0
        ? Math.abs(serverMs - estimatedServerMs)
        : Number.POSITIVE_INFINITY
      if (drift > 1000) {
        applyServerClock(response.server_time, midpoint)
      }
    }
  }

  const serverNowMs = parseServerTimeMs(response.server_time)
  const stableNowMs = Number.isFinite(serverNowMs) && serverNowMs > 0 ? serverNowMs : syncedNow.value
  syncCountdownTarget(response.current_period, stableNowMs, shouldRebaseClock)
  logRealtimeEvent('room.state.applied', {
    roomCode: responseRoomCode,
    previousPeriodNo,
    nextPeriodNo,
    status: String(response.current_period?.status ?? ''),
    shouldRebaseClock,
  })
  await maybeShowSettlementModal(previousPeriod, response.current_period)

  // Tự động cập nhật bảng lịch sử game nếu đang ở trang 1
  if (historyPage.value === 1) {
    if (previousPeriodNo && previousPeriodNo !== nextPeriodNo) {
        // Có sự chuyển kỳ, tải lại lịch sử mới nhất từ Server qua WS
        setTimeout(() => {
          loadRoomHistory(1)
          refreshRealtimeSlices({ mine: true }) // Làm mới lịch sử cá nhân khi sang kỳ mới
        }, 400);
    } else {
        // Chỉ cập nhật từ mảng snapshot nếu không có sự thay đổi kỳ (vẫn trong cùng 1 kỳ nhưng có update kết quả)
        applyHistoryFromRecentResults(1)
    }
  }

  const nextStatus = String(response.current_period?.status ?? '').toUpperCase()
  if (nextStatus === 'SETTLED') {
      void refreshRealtimeSlices({ mine: true, history: true, wallet: true, reason: 'period_settled' })
  }
  
  await maybeShowSettlementModal(previousPeriod, response.current_period)

  if (previousPeriodNo && previousPeriodNo !== nextPeriodNo && ['OPEN', 'LOCKED', 'DRAWN', 'SETTLED'].includes(nextStatus)) {
    scheduleRoomStateReconcile(selectedRoomCode.value, 650)
  }
  if (previousPeriodNo && previousPeriodNo !== nextPeriodNo) {
    logRealtimeEvent('period.transition', {
      roomCode: selectedRoomCode.value,
      previousPeriodNo,
      nextPeriodNo,
    })
  }

  resolvePendingSettlementModalFromMineRows()
}

function disconnectRoomStateStream() {
  if (roomStreamReconnectTimer) {
    window.clearTimeout(roomStreamReconnectTimer)
    roomStreamReconnectTimer = undefined
  }
  for (const pending of pendingRoomHistoryRequests.values()) {
    clearTimeout(pending.timeout)
    pending.reject(new Error('Kết nối realtime phòng chơi đã ngắt.'))
  }
  pendingRoomHistoryRequests.clear()
  roomStreamConnection?.close()
  roomStreamConnection = null
}

function requestRoomStateSnapshot(roomCode = selectedRoomCode.value) {
  if (!roomCode) return
  if (!roomStreamConnection || roomStreamConnection.readyState !== WebSocket.OPEN) return
  try {
    roomStreamConnection.send(JSON.stringify({ action: 'request_state' }))
  } catch {
    // ignore request_state send failures
  }
}

function requestRoomHistoryViaWS(page: number, pageSize = tablePageSize): Promise<PlayRoomHistoryResponse> {
  const normalizedPage = Math.max(1, Math.floor(page))
  const normalizedPageSize = Math.max(1, Math.floor(pageSize))
  if (!roomStreamConnection || roomStreamConnection.readyState !== WebSocket.OPEN) {
    return Promise.reject(new Error('Kết nối realtime chưa sẵn sàng để tải lịch sử.'))
  }

  const requestId = globalThis.crypto?.randomUUID?.() ?? `history-${Date.now()}-${normalizedPage}`
  return new Promise<PlayRoomHistoryResponse>((resolve, reject) => {
    const timeout = window.setTimeout(() => {
      pendingRoomHistoryRequests.delete(requestId)
      reject(new Error('Hết thời gian chờ tải lịch sử realtime.'))
    }, 5000)

    pendingRoomHistoryRequests.set(requestId, { resolve, reject, timeout })
    try {
      roomStreamConnection?.send(JSON.stringify({
        action: 'request_history',
        request_id: requestId,
        page: normalizedPage,
        page_size: normalizedPageSize,
      }))
    } catch (error) {
      pendingRoomHistoryRequests.delete(requestId)
      clearTimeout(timeout)
      reject(error instanceof Error ? error : new Error('Không thể gửi yêu cầu lịch sử realtime.'))
    }
  })
}

function disconnectBetsStream() {
  betsStreamConnection?.close()
  betsStreamConnection = null
}

function buildBetsWSUrl(roomCode: string): string {
  const endpoint = `/v1/play/rooms/${roomCode}/bets/ws`
  const fallbackProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const fallback = `${fallbackProtocol}//${window.location.host}${endpoint}`
  try {
    const rawBase = (env.apiBaseUrl || '').trim()
    const baseUrl = new URL(rawBase || window.location.origin, window.location.origin)
    const wsProtocol = baseUrl.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = new URL(`${wsProtocol}//${baseUrl.host}${endpoint}`)
    if (auth.accessToken) {
      wsUrl.searchParams.set('access_token', auth.accessToken)
    }
    return wsUrl.toString()
  } catch {
    if (!auth.accessToken) return fallback
    const wsUrl = new URL(fallback)
    wsUrl.searchParams.set('access_token', auth.accessToken)
    return wsUrl.toString()
  }
}

function connectBetsStream(roomCode: string): void {
  if (!roomCode || !auth.accessToken) return
  
  disconnectBetsStream()
  betsStreamHistoryRequested = false
  try {
    const socket = new WebSocket(buildBetsWSUrl(roomCode))
    betsStreamConnection = socket
    
    socket.onopen = () => {
      // Request full mine history via WebSocket (instead of REST API)
      if (!betsStreamHistoryRequested) {
        betsStreamHistoryRequested = true
        socket.send(JSON.stringify({ action: 'request_history' }))
      }
    }
    
    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(String(event.data)) as { event?: string; data?: any }
        if (payload.event === 'bets.init') {
          const data = payload.data as PlayRoomBetHistoryResponse | undefined
          if (data?.items && Array.isArray(data.items)) {
            mineRowsCached.length = 0
            mineRowsCached.push(...data.items)
            mineTotalPages.value = Math.max(1, Number(data.total_pages ?? Math.ceil(mineRowsCached.length / tablePageSize)))
            updateMineRowsFromCache(minePage.value)
            mineError.value = ''
            mineLoading.value = false
          }
        } else if (payload.event === 'bets.history_chunk') {
          // Receive paginated history chunks from server
          if (payload.data?.items && Array.isArray(payload.data.items)) {
            // Add new bets to cache (avoid duplicates by period_no)
            for (const bet of payload.data.items) {
              if (!mineRowsCached.find((b) => b.period_no === bet.period_no)) {
                mineRowsCached.push(bet)
              }
            }
            // Update UI with current page from cache
            updateMineRowsFromCache(minePage.value)
            mineTotalPages.value = Math.max(1, Math.ceil(mineRowsCached.length / tablePageSize))
            mineError.value = ''
            mineLoading.value = false
          }
        } else if (payload.event === 'bets.history_complete') {
          // History stream finished
          logRealtimeEvent('bets.history_complete', { totalBets: mineRowsCached.length })
          void resolvePendingSettlementModalFromMineRows()
        } else if (payload.event === 'bets.updated') {
          logRealtimeEvent('bets.updated', { roomCode })
          // Khi có cập nhật cược (chốt kết quả), yêu cầu Server gửi lại lịch sử mới nhất
          if (betsStreamConnection && betsStreamConnection.readyState === WebSocket.OPEN) {
            mineRowsCached.length = 0 // Xóa cache cũ để nhận data mới
            betsStreamConnection.send(JSON.stringify({ action: 'request_history' }))
          }
          wallet.connectStream()
        }
      } catch {
        // ignore malformed ws payload
      }
    }
    
    socket.onerror = () => {
      // error on websocket
    }
    
    socket.onclose = () => {
      betsStreamConnection = null
    }
  } catch {
    // ignore connection errors
  }
}

function updateMineRowsFromCache(page: number) {
  const normalized = Math.max(1, Math.floor(page))
  const start = (normalized - 1) * tablePageSize
  const end = start + tablePageSize
  mineRows.value = mineRowsCached.slice(start, end)
  minePage.value = normalized
}

function buildRoomWSUrl(roomCode: string): string {
  const endpoint = `/v1/play/rooms/${roomCode}/ws`
  const fallbackProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const fallback = `${fallbackProtocol}//${window.location.host}${endpoint}`
  try {
    const rawBase = (env.apiBaseUrl || '').trim()
    const baseUrl = new URL(rawBase || window.location.origin, window.location.origin)
    const wsProtocol = baseUrl.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = new URL(`${wsProtocol}//${baseUrl.host}${endpoint}`)
    if (auth.accessToken) {
      wsUrl.searchParams.set('access_token', auth.accessToken)
    }
    return wsUrl.toString()
  } catch {
    if (!auth.accessToken) return fallback
    const wsUrl = new URL(fallback)
    wsUrl.searchParams.set('access_token', auth.accessToken)
    return wsUrl.toString()
  }
}

function scheduleRoomWSReconnect(roomCode: string) {
  if (roomStreamReconnectTimer) return
  roomStreamReconnectTimer = window.setTimeout(() => {
    roomStreamReconnectTimer = undefined
    connectRoomStateStream(roomCode)
  }, 2500)
}

function connectRoomStateStream(roomCode: string) {
  if (!roomCode) return
  const generation = roomStateGeneration

  disconnectRoomStateStream()
  try {
    const socket = new WebSocket(buildRoomWSUrl(roomCode))
    roomStreamConnection = socket
    socket.onopen = () => {
      roomStateError.value = ''
      // Request initial room state via WebSocket instead of REST API
      socket.send(JSON.stringify({ action: 'request_state' }))
      if (activeHistoryTab.value === 'history') {
        void loadRoomHistory(historyPage.value || 1)
      }
    }

    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(String(event.data)) as { event?: string; data?: any }
        if (payload.event === 'room.clock') {
          if (!isCurrentRoomStateGeneration(generation, roomCode)) return
          applyServerClock(String(payload.data?.server_time ?? ''))
          return
        }
        if (payload.event === 'room.state') {
          if (!isCurrentRoomStateGeneration(generation, roomCode)) return
          roomStateError.value = ''
          roomStateLoading.value = false
          void applyRoomStateResponse(payload.data as PlayRoomStateResponse, { roomCode, generation, forceRebaseClock: !roomState.value })
          return
        }
        if (payload.event === 'history.page') {
          const requestId = String(payload.data?.request_id ?? '')
          const response = payload.data as PlayRoomHistoryResponse
          if (requestId && pendingRoomHistoryRequests.has(requestId)) {
            const pending = pendingRoomHistoryRequests.get(requestId)
            if (pending) {
              clearTimeout(pending.timeout)
              pendingRoomHistoryRequests.delete(requestId)
              pending.resolve(response)
            }
          }
          return
        }
        if (payload.event === 'history.error') {
          const requestId = String(payload.data?.request_id ?? '')
          const message = String(payload.data?.message ?? 'Không thể tải lịch sử realtime')
          if (requestId && pendingRoomHistoryRequests.has(requestId)) {
            const pending = pendingRoomHistoryRequests.get(requestId)
            if (pending) {
              clearTimeout(pending.timeout)
              pendingRoomHistoryRequests.delete(requestId)
              pending.reject(new Error(message))
            }
          }
          return
        }
        if (payload.event === 'bet.placed') {
          // Handle bet placement response from WebSocket
          const requestId = String(payload.data?.request_id ?? '')
          logRealtimeEvent('bet.ws.event.placed', {
            roomCode,
            requestId,
            payload: payload.data,
          })
          if (requestId && pendingBetRequests.has(requestId)) {
            const pending = pendingBetRequests.get(requestId)
            if (pending) {
              clearTimeout(pending.timeout)
              pendingBetRequests.delete(requestId)
              pending.resolve(payload.data)
            }
          }
          return
        }
        if (payload.event === 'bet.error') {
          // Handle bet placement error from WebSocket
          const requestId = String(payload.data?.request_id ?? '')
          logRealtimeEvent('bet.ws.event.error', {
            roomCode,
            requestId,
            payload: payload.data,
          })
          if (requestId && pendingBetRequests.has(requestId)) {
            const pending = pendingBetRequests.get(requestId)
            if (pending) {
              clearTimeout(pending.timeout)
              pendingBetRequests.delete(requestId)
              pending.reject(new Error(payload.data?.message ?? 'Lỗi khi đặt lệnh'))
            }
          }
          return
        }
      } catch {
        // ignore malformed ws payload
      }
    }

    socket.onerror = () => {
      roomStateError.value = 'Kết nối realtime phòng chơi đang được nối lại'
    }

    socket.onclose = () => {
      roomStreamConnection = null
      scheduleRoomWSReconnect(roomCode)
    }
  } catch {
    roomStateError.value = 'Kết nối realtime phòng chơi đang được nối lại'
    scheduleRoomWSReconnect(roomCode)
  }
}

async function loadRoomHistory(page = historyPage.value) {
  if (!selectedRoomCode.value) return

  historyLoading.value = true
  historyError.value = ''
  try {
    logRealtimeEvent('history.fetch.ws', { page })
    const response = await requestRoomHistoryViaWS(page, tablePageSize)
    historyRows.value = response.items
    historyPage.value = response.page
    historyTotalPages.value = response.total_pages
  } catch (error: unknown) {
    const err = error as Error
    historyError.value = err?.message ?? 'Không thể tải lịch sử game'
    historyRows.value = []
  } finally {
    historyLoading.value = false
  }
}

async function loadChartHistory() {
  if (!selectedRoomCode.value) return

  // Chart history is included in roomState loaded via WebSocket
  if (roomState.value?.recent_results?.length) {
    chartRows.value = roomState.value.recent_results.slice(0, 24)
    chartError.value = ''
    chartLoading.value = false
    return
  }
  
  // Fallback to REST API if WebSocket hasn't delivered yet
  chartLoading.value = true
  chartError.value = ''
  try {
    logRealtimeEvent('chart.fallback', { reason: 'no_room_state' })
    const response = await request<PlayRoomHistoryResponse>(
      'GET',
      `/v1/play/rooms/${selectedRoomCode.value}/history?page=1&page_size=24`,
      {},
    )
    chartRows.value = response.items
  } catch (error: unknown) {
    const err = error as ApiError
    chartError.value = err?.message ?? 'Không thể tải dữ liệu biểu đồ'
    chartRows.value = []
  } finally {
    chartLoading.value = false
  }
}

async function loadMineHistory(page = minePage.value, options: { force?: boolean } = {}) {
  const normalizedPage = Math.max(1, Math.floor(page))
  
  // All mine history is now streamed via WebSocket into mineRowsCached
  // Just update the displayed rows from cache (client-side pagination)
  if (mineRowsCached.length > 0) {
    updateMineRowsFromCache(normalizedPage)
    return
  }
  
  // If cache is empty, wait a bit for WebSocket to deliver history
  // (should happen automatically when connection opens)
  mineLoading.value = true
  mineError.value = ''
  
  // Wait up to 3 seconds for WebSocket to provide data
  let attempts = 0
  while (mineRowsCached.length === 0 && attempts < 30) {
    await new Promise((resolve) => setTimeout(resolve, 100))
    attempts += 1
  }
  
  if (mineRowsCached.length > 0) {
    updateMineRowsFromCache(normalizedPage)
    mineLoading.value = false
    return
  }

  // Controlled REST fallback for WS cold start.
  const now = Date.now()
  if (!options.force && now - mineHistoryRestFallbackAt < 2500) {
    mineError.value = 'Không thể tải lịch sử cược'
    mineRows.value = []
    mineLoading.value = false
    return
  }

  mineHistoryRestFallbackAt = now
  try {
    const response = await request<PlayRoomBetHistoryResponse>(
      'GET',
      `/v1/play/rooms/${selectedRoomCode.value}/bets?page=${normalizedPage}&page_size=${tablePageSize}`,
      { token: auth.accessToken },
    )
    mineRows.value = response.items
    minePage.value = response.page
    mineTotalPages.value = response.total_pages
    mineError.value = ''
  } catch (error: unknown) {
    const err = error as ApiError
    mineError.value = err?.message ?? 'Không thể tải lịch sử cược'
    mineRows.value = []
  }
  
  mineLoading.value = false
}

async function maybeShowSettlementModal(
  previousPeriod: PlayRoomStateResponse['current_period'] | null,
  nextPeriod: PlayRoomStateResponse['current_period'] | null,
) {
  if (!nextPeriod || !auth.accessToken) return
  
  // Tránh hiện liên tục cho cùng một kỳ đã chốt
  if (nextPeriod.status === 'SETTLED' && lastSettledPeriodId.value === nextPeriod.id) {
    return
  }

  const nextPeriodNo = nextPeriod.period_no
  const isTransition = nextPeriodNo && previousPeriod && previousPeriod.period_no !== nextPeriodNo
  const isManualSettled = String(nextPeriod.status ?? '').toUpperCase() === 'SETTLED'

  const targetPeriodNo = isTransition ? previousPeriod.period_no : isManualSettled ? nextPeriodNo : ''
  const settlementKey = settlementKeyFor(selectedRoomCode.value, targetPeriodNo)

  if (!targetPeriodNo || seenSettlementPeriods.has(settlementKey)) {
    return
  }

  settlementTargets.set(settlementKey, targetPeriodNo)
  
  // Thử lại tối đa 4 lần nếu dữ liệu lịch sử chưa về kịp
  for (let i = 0; i < 4; i++) {
    const success = resolvePendingSettlementModalFromMineRows()
    if (success) break
    
    // Nếu chưa thấy kết quả, ép Server gửi lại lịch sử ngay
    if (i === 1) {
       void refreshRealtimeSlices({ mine: true, wallet: true, reason: 'settlement_retry' })
    }
    await new Promise(r => setTimeout(r, 1000))
  }
}

function clearSettlementRetry(settlementKey: string) {
  settlementTargets.delete(settlementKey)
}

function resolvePendingSettlementModalFromMineRows() {
  if (!settlementTargets.size || !mineRows.value.length) return false

  let anySuccess = false
  for (const [settlementKey, targetPeriodNo] of settlementTargets.entries()) {
    if (!targetPeriodNo) {
      settlementTargets.delete(settlementKey)
      continue
    }

    if (seenSettlementPeriods.has(settlementKey)) {
      clearSettlementRetry(settlementKey)
      settlementTargets.delete(settlementKey)
      continue
    }

    if (checkAndShowSettlementForPeriod(targetPeriodNo)) {
      clearSettlementRetry(settlementKey)
      settlementTargets.delete(settlementKey)
      anySuccess = true
      continue
    }
  }
  return anySuccess
}

function checkAndShowSettlementForPeriod(periodNo: string): boolean {
  if (!periodNo) return false
  
  // Tìm tất cả các vé của kỳ này
  const matches = mineRows.value.filter(row => String(row.period_no) === String(periodNo))
  if (matches.length === 0) return false
  
  // Kiểm tra xem tất cả các vé đã có kết quả chưa
  const allSettled = matches.every(row => {
    const status = String(row.status ?? '').toUpperCase()
    return ['WON', 'LOST', 'VOID', 'HALF_WON', 'HALF_LOST', 'CANCELED', 'CASHED_OUT'].includes(status)
  })
  
  if (!allSettled) return false // Vẫn còn vé đang chờ xử lý

  // Tổng hợp dữ liệu
  let totalStake = 0
  let totalPayout = 0
  let latestSettledAt = ''
  let latestPeriodId = 0

  matches.forEach(row => {
    totalStake += rowOriginalAmountValue(row)
    totalPayout += rowWinCreditValue(row)
    if (row.settled_at && (!latestSettledAt || row.settled_at > latestSettledAt)) {
      latestSettledAt = row.settled_at
    }
    if (row.period_id) latestPeriodId = row.period_id
  })

  const profitLoss = totalPayout - totalStake
  const status = profitLoss > 0 ? 'WON' : profitLoss < 0 ? 'LOST' : 'DRAW'

  // Hiển thị modal tổng kết
  const settlementKey = settlementKeyFor(selectedRoomCode.value, periodNo)
  markSettlementHandled(selectedRoomCode.value, periodNo)
  
  logRealtimeEvent('settlement.summary.modal', {
    roomCode: selectedRoomCode.value,
    periodNo,
    totalTickets: matches.length,
    totalStake,
    totalPayout,
    profitLoss
  })

  wallet.connectStream()
  clearSettlementRetry(settlementKey)
  
  resultModalPeriodNo.value = periodNo
  resultModalTitle.value = profitLoss > 0 ? 'Chúc mừng! Kỳ quay có lãi' : 
                           profitLoss < 0 ? 'Kết quả kỳ quay (Chưa may mắn)' : 'Kết quả kỳ quay (Hòa vốn)'
  
  resultModalDescription.value = profitLoss > 0 
    ? `Bạn đã thắng tổng cộng ${matches.length} vé trong kỳ này.`
    : `Tổng hợp kết quả của ${matches.length} vé cược trong kỳ.`

  resultModalTone.value = profitLoss > 0 ? 'win' : profitLoss < 0 ? 'lose' : 'draw'
  resultModalStake.value = formatMoney(totalStake)
  resultModalPayout.value = formatMoney(totalPayout)
  resultModalAmount.value = formatSignedMoney(profitLoss)
  
  resultModalSettledAt.value = formatLocalTime(latestSettledAt || matches[0]?.created_at)
  showResultModal.value = true
  if (latestPeriodId) lastSettledPeriodId.value = latestPeriodId
  
  return true
}

function formatLocalTime(isoString?: string | null) {
  if (!isoString) return new Date().toLocaleTimeString('vi-VN', { hour12: false })
  try {
    const d = new Date(isoString)
    // Nếu isoString không có chỉ thị muối giờ, coi như là UTC
    if (!isoString.includes('Z') && !isoString.includes('+')) {
      const utcDate = new Date(isoString + 'Z')
      if (!isNaN(utcDate.getTime())) return utcDate.toLocaleTimeString('vi-VN', { hour12: false })
    }
    return d.toLocaleTimeString('vi-VN', { hour12: false })
  } catch (e) {
    return isoString
  }
}

async function loadActiveHistory(page = currentPage()) {
  if (activeHistoryTab.value === 'mine') {
    // Only load if empty or different page
    if (!mineRows.value.length || minePage.value !== page) {
      await loadMineHistory(page)
    }
    return
  }
  if (activeHistoryTab.value === 'chart') {
    if (!chartRows.value.length) {
      await loadChartHistory()
    }
    return
  }
  if (!historyRows.value.length) {
    await loadRoomHistory(page)
  }
}

async function joinRoom() {
  if (!room.value || room.value.status !== 'OPEN' || !auth.accessToken) return false
  if (connectionId.value) return true
  if (!wallet.wallets.length) await loadWallet()
  if (availableVndBalance.value <= 0) {
    joinError.value = 'Số dư khả dụng không đủ để vào phòng chơi.'
    return false
  }
  joinLoading.value = true
  joinError.value = ''
  connectionId.value = ''
  try {
    const res = await request<GameJoinResponse>('POST', `/v1/games/${room.value.code}/join`, {
      token: auth.accessToken,
    })
    connectionId.value = res.connection_id
    return true
  } catch (error: unknown) {
    const err = error as ApiError
    joinError.value = err?.message ?? 'Không thể vào phòng'
    return false
  } finally {
    joinLoading.value = false
  }
}

function maybeAutoJoinRoom() {
  if (!room.value || !selectedRoomCode.value || !auth.accessToken) return
  if (room.value.status !== 'OPEN') return
  if (connectionId.value || joinLoading.value) return
  if (availableVndBalance.value <= 0) return

  const attemptKey = `${room.value.code}:${selectedRoomCode.value}`
  if (autoJoinAttemptKey === attemptKey) return
  autoJoinAttemptKey = attemptKey

  if (autoJoinTimer) {
    window.clearTimeout(autoJoinTimer)
    autoJoinTimer = undefined
  }
  autoJoinTimer = window.setTimeout(() => {
    autoJoinTimer = undefined
    void joinRoom()
  }, 120)
}

function openBetModal(groupTitle: string, optionKey: string, optionLabel: string) {
  if (!canBet.value) return
  modalBetGroupTitle.value = groupTitle
  modalBetKey.value = optionKey
  modalBetLabel.value = optionLabel
  modalBetAmount.value = initialBetAmount
  modalBetStepAmount.value = minimumBetAmount
  showBetModal.value = true
}

function setBetStepAmount(amount: number) {
  if (!Number.isFinite(amount) || amount <= 0) return
  modalBetStepAmount.value = Math.max(minimumBetAmount, Math.floor(amount))
}

function addBetAmount(amount: number) {
  setBetStepAmount(Math.abs(amount))
  const newAmount = modalBetAmount.value + amount
  // Keep amount within the allowed bet range.
  if (newAmount >= minimumBetAmount && newAmount <= availableVndBalance.value) {
    modalBetAmount.value = newAmount
  }
}

function canAddAmount(amount: number): boolean {
  const newAmount = modalBetAmount.value + amount
  return newAmount >= minimumBetAmount && newAmount <= availableVndBalance.value
}

function adjustBetAmount(direction: -1 | 1) {
  const delta = modalBetStepAmount.value * direction
  const newAmount = modalBetAmount.value + delta
  if (newAmount >= minimumBetAmount && newAmount <= availableVndBalance.value) {
    modalBetAmount.value = newAmount
  }
}

function canAdjustBetAmount(direction: -1 | 1) {
  const delta = modalBetStepAmount.value * direction
  const newAmount = modalBetAmount.value + delta
  return newAmount >= minimumBetAmount && newAmount <= availableVndBalance.value
}

function selectOption(groupTitle: string, key: string, label: string) {
  selectedOptions[groupTitle] = key
  openBetModal(groupTitle, key, label)
}

function groupTypeKey(group: PlayBetGroup): string {
  const title = group.title.toLowerCase()
  if (title.includes('màu') || title.includes('màu sắc')) return 'COLOR'
  if (title.includes('chọn số') || title.includes('vị trí')) return 'NUMBER'
  if (title.includes('lớn') || title.includes('nhỏ')) return 'BIG_SMALL'
  if (title.includes('tổng') && title.includes('điểm')) return 'SUM'
  if (title.includes('chẵn') || title.includes('lẻ') || title.includes('tổng hợp')) return 'ODD_EVEN'
  if (title.includes('bộ ba')) return 'COMBINATION'
  return 'OPTION'
}

async function sendBetViaSocket(betPayload: any) {
  // Ensure room state connection is ready
  if (!roomStreamConnection || roomStreamConnection.readyState !== WebSocket.OPEN) {
    logRealtimeEvent('bet.ws.unavailable', {
      roomCode: selectedRoomCode.value,
      readyState: roomStreamConnection?.readyState ?? -1,
    })
    throw new Error('Kết nối với máy chủ không sẵn sàng. Vui lòng thử lại.')
  }

  const requestId = betPayload.request_id
  logRealtimeEvent('bet.ws.send', {
    roomCode: selectedRoomCode.value,
    requestId,
    periodId: String(betPayload.period_id ?? ''),
    connectionId: String(betPayload.connection_id ?? ''),
    items: Array.isArray(betPayload.items) ? betPayload.items.length : 0,
  })
  return new Promise<any>((resolve, reject) => {
    const timeoutId = window.setTimeout(() => {
      pendingBetRequests.delete(requestId)
      logRealtimeEvent('bet.ws.timeout', {
        roomCode: selectedRoomCode.value,
        requestId,
      })
      reject(new Error('Hết thời gian chờ phản hồi từ máy chủ.'))
    }, 5000) // 5 second timeout

    pendingBetRequests.set(requestId, {
      resolve,
      reject,
      timeout: timeoutId,
    })

    // Send bet through WebSocket
    try {
      roomStreamConnection?.send(JSON.stringify({
        action: 'place_bet',
        ...betPayload,
      }))
    } catch (error) {
      pendingBetRequests.delete(requestId)
      clearTimeout(timeoutId)
      reject(error)
    }
  })
}

async function confirmBet() {
  if (!room.value || !selectedVariant.value || !auth.accessToken) {
    betMessage.value = 'Vui lòng đăng nhập và vào phòng chơi trước.'
    betMessageRoomCode.value = selectedRoomCode.value
    showBetModal.value = false
    return
  }
  if (isBetLocked.value) {
    betMessage.value = 'Kỳ hiện tại đã khóa, không thể đặt lệnh.'
    betMessageRoomCode.value = selectedRoomCode.value
    showBetModal.value = false
    return
  }
  if (!currentPeriod.value) {
    betMessage.value = 'Chưa có kỳ hiện tại.'
    betMessageRoomCode.value = selectedRoomCode.value
    showBetModal.value = false
    return
  }

  betLoading.value = true
  betMessage.value = ''
  betMessageRoomCode.value = ''
  showBetModal.value = false

  if (!connectionId.value) {
    const joined = await joinRoom()
    if (!joined || !connectionId.value) {
      betMessage.value = joinError.value || 'Không thể vào phòng chơi lúc này.'
      betMessageRoomCode.value = selectedRoomCode.value
      return
    }
  }

  try {
    const requestId = globalThis.crypto?.randomUUID?.() ?? `req-${Date.now()}`
    logRealtimeEvent('bet.request', {
      roomCode: selectedRoomCode.value,
      requestId,
      connectionId: connectionId.value,
      periodId: String(currentPeriod.value.id),
      optionKey: modalBetKey.value,
      amount: modalBetAmount.value,
      socketState: roomStreamConnection?.readyState ?? -1,
    })
    // setPendingWalletDebit(modalBetAmount.value) // Bỏ trừ tiền ảo để tránh hụt 2 lần
    
    // Send bet through WebSocket instead of REST API
    const betResponse = await sendBetViaSocket({
      request_id: requestId,
      period_id: String(currentPeriod.value.id),
      connection_id: connectionId.value,
      items: [{
        option_type: groupTypeKey({ title: modalBetGroupTitle.value, description: '', mode: 'chips', options: [] }),
        option_key: modalBetKey.value,
        stake: String(modalBetAmount.value),
      }],
    })

    // Explicit success feedback
    logRealtimeEvent('bet.response.ok', {
      roomCode: selectedRoomCode.value,
      requestId,
      response: betResponse,
    })
    betMessage.value = betResponse?.message || 'Lệnh cược đã được hệ thống tiếp nhận thành công!'
    betMessageRoomCode.value = selectedRoomCode.value
    logRealtimeEvent('bet.placed', {
      roomCode: selectedRoomCode.value,
      requestId,
      amount: modalBetAmount.value,
      optionKey: modalBetKey.value,
      periodId: String(currentPeriod.value.id),
      source: 'websocket',
    })

    // Refresh state after a tiny delay to ensure DB commit is visible.
    setTimeout(async () => {
      await refreshRealtimeSlices({
        history: false,
        mine: true, // Force refresh even if not on 'mine' tab yet
        chart: false,
        wallet: true,
        reason: 'bet_placed',
      })
      setPendingWalletDebit(0)
    }, 250);
    
    // Auto clear success message after a few seconds
    setTimeout(() => {
      if (betMessage.value.includes('tiếp nhận') || betMessage.value.includes('thành công')) {
        betMessage.value = ''
      }
    }, 4500)
    
  } catch (error: unknown) {
    const err = error as ApiError
    logRealtimeEvent('bet.response.error', {
      roomCode: selectedRoomCode.value,
      error: err?.message ?? String(error),
      socketState: roomStreamConnection?.readyState ?? -1,
      connectionId: connectionId.value,
      periodId: String(currentPeriod.value?.id ?? ''),
    })
    betMessage.value = err?.message ?? 'Không thể gửi lệnh cược. Vui lòng thử lại.'
    betMessageRoomCode.value = selectedRoomCode.value
    setPendingWalletDebit(0)
    wallet.connectStream()
  } finally {
    betLoading.value = false
  }
}

function chartValue(row: PlayRoomHistoryResponse['items'][number]) {
  const raw = String(row.result ?? '')
  if (isK3.value) {
    return raw.split('-').map((item) => Number(item)).reduce((total, item) => total + (Number.isFinite(item) ? item : 0), 0)
  }
  if (room.value?.code === 'lottery') {
    return raw.split('').map((item) => Number(item)).reduce((total, item) => total + (Number.isFinite(item) ? item : 0), 0)
  }
  return Number(raw) || 0
}

function chartSummaryLabel(row: PlayRoomHistoryResponse['items'][number]) {
  if (row.big_small) return normalizeBetLabel(row.big_small)
  if (row.color) return normalizeBetLabel(row.color)
  return row.result || '—'
}

function chartBarClass(row: PlayRoomHistoryResponse['items'][number]) {
  const label = String(row.big_small || row.color || row.result || '').toLowerCase()
  if (label.includes('xanh') || label.includes('green')) return 'bg-[#24b561]'
  if (label.includes('đỏ') || label.includes('red')) return 'bg-[#e64545]'
  if (label.includes('tím') || label.includes('violet')) return 'bg-[#8b5cf6]'
  if (label.includes('lớn') || label.includes('big')) return 'bg-primary'
  if (label.includes('nhỏ') || label.includes('small')) return 'bg-[#3b82f6]'
  return 'bg-slate-400'
}


function closeResultModal() {
  showResultModal.value = false
}

function openTicketDetail(row: PlayRoomBetHistoryResponse['items'][number]) {
  selectedTicketDetail.value = row
  showTicketDetailModal.value = true
}

function closeTicketDetail() {
  showTicketDetailModal.value = false
  selectedTicketDetail.value = null
}



function extractWingoNumber(value: string | null | undefined): number | null {
  const raw = String(value ?? '').trim().toLowerCase()
  if (!raw) return null

  let parsed = Number.NaN
  if (raw.startsWith('number_')) {
    parsed = Number.parseInt(raw.replace('number_', ''), 10)
  } else {
    parsed = Number.parseInt(raw, 10)
  }

  if (!Number.isInteger(parsed) || parsed < 0 || parsed > 9) return null
  return parsed
}



watch(
  () => room.value?.code,
  async () => {
    resetTransientRoomUiState()
    if (!room.value) return
    if (room.value.variants.length === 0) {
      activeVariantCode.value = ''
      return
    }
    activeVariantCode.value = room.value.variants[0]?.code ?? ''
    // Reset K3 sub-tab to first
    if (isK3.value) activeK3SubTab.value = 'Tổng số'
    await nextTick()
    ensureDefaultSelections(room.value.variants[0] ?? null)
  },
  { immediate: true },
)

watch(
  () => selectedRoomCode.value,
  async (roomCode) => {
    if (!roomCode) {
      disconnectRoomStateStream()
      disconnectBetsStream()
      return
    }
    const generation = bumpRoomStateGeneration()
    resetTransientRoomUiState()
    resetRoomViewData()
    resetRoomStateSession()
    hydrateHandledSettlementPeriods(roomCode)
    void loadWallet()
    hydrateRoomStateFromCache(roomCode)
    roomStateLoading.value = true
    roomStateError.value = ''
    setLoading(true)
    try {
      connectRoomStateStream(roomCode)
      connectBetsStream(roomCode)
      requestRoomStateSnapshot(roomCode)
      if (!isCurrentRoomStateGeneration(generation, roomCode)) return
      maybeAutoJoinRoom()
    } finally {
      if (!isCurrentRoomStateGeneration(generation, roomCode)) return
      // Small delay to ensure smooth transition
      setTimeout(() => setLoading(false), 300)
    }
  },
  { immediate: true },
)

watch(
  () => selectedVariant.value?.code,
  () => {
    resetTransientRoomUiState()
    ensureDefaultSelections(selectedVariant.value)
  },
  { immediate: true },
)

watch(availableVndBalance, (balance) => {
  if (balance > 0 && joinError.value.includes('Số dư khả dụng không đủ')) {
    joinError.value = ''
  }
  if (balance > 0 && !connectionId.value) {
    maybeAutoJoinRoom()
  }
})

watch(
  () => [currentPeriod.value?.id, currentPeriod.value?.draw_at] as const,
  (current, previous) => {
    const [periodID] = current
    const [previousPeriodID] = previous ?? []
    syncCountdownTarget(
      currentPeriod.value,
      syncedNow.value,
      String(periodID ?? '') !== String(previousPeriodID ?? ''),
    )
  },
  { immediate: true },
)

watch(
  () => [currentPeriodKey.value, rawRemainingSeconds.value] as const,
  ([periodKey, rawRemaining]) => {
    if (!periodKey) {
      stableCountdownPeriodKey.value = ''
      stableRemainingSeconds.value = 0
      return
    }

    if (stableCountdownPeriodKey.value !== periodKey) {
      stableCountdownPeriodKey.value = periodKey
      stableRemainingSeconds.value = rawRemaining
      return
    }

    if (stableRemainingSeconds.value === 0 && rawRemaining > 0) {
      return
    }

    if (rawRemaining > stableRemainingSeconds.value) {
      return
    }

    stableRemainingSeconds.value = rawRemaining
  },
  { immediate: true },
)

// consolidated timer sync into period_no watcher and state application

watch(
  () => activeHistoryTab.value,
  async () => {
    await loadActiveHistory(currentPage())
  },
)

watch(remainingSeconds, (newVal, oldVal) => {
  if (newVal > 0 && periodRollForwardTimer) {
    window.clearTimeout(periodRollForwardTimer)
    periodRollForwardTimer = undefined
  }

  if (newVal === 0 && oldVal !== 0 && currentPeriod.value) {
    const status = String(currentPeriod.value.status ?? '').toUpperCase()
    if ((status === 'OPEN' || status === 'LOCKED') && !periodRollForwardTimer) {
      periodRollForwardTimer = window.setTimeout(() => {
        periodRollForwardTimer = undefined
        requestRoomStateSnapshot()
      }, 1200)
    }
  }

  if (newVal !== oldVal && newVal > 0 && newVal <= 5) {
    playTickSound()
  }
})

// Watch for dice result changes to trigger animation
watch(currentDice, (newDice) => {
  if (!isK3.value) return
  isDiceRolling.value = true
  if (rollTimer) window.clearInterval(rollTimer)
  
  let count = 0
  const maxTicks = 15
  rollTimer = window.setInterval(() => {
    rollingDice.value = rollingDice.value.map(() => Math.floor(Math.random() * 6) + 1)
    count++
    if (count >= maxTicks) {
      if (rollTimer) window.clearInterval(rollTimer)
      isDiceRolling.value = false
      rollingDice.value = [...newDice]
    }
  }, 100)
}, { deep: true })

onMounted(() => {
  void loadWallet()
  rollingDice.value = [...currentDice.value]
  // Cập nhật clockTick thường xuyên hơn để countdown mượt và bám server hơn
  timer = window.setInterval(() => { clockTick.value = Date.now() }, 250)
  maybeAutoJoinRoom()
})

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
  if (roomStateReconcileTimer) window.clearTimeout(roomStateReconcileTimer)
  if (periodRollForwardTimer) window.clearTimeout(periodRollForwardTimer)
  if (autoJoinTimer) window.clearTimeout(autoJoinTimer)
  disconnectRoomStateStream()
  disconnectBetsStream()
  wallet.disconnectStream()
})
</script>

<template>
  <div class="min-h-dvh bg-[#f7f0f0]">
    <div v-if="room && selectedVariant" class="min-h-dvh pb-28">
    <!-- ===== HEADER GRADIENT ===== -->
    <header class="flex items-center justify-between bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] px-4 py-3 text-white shadow-lg">
      <button class="grid h-10 w-10 place-items-center rounded-full bg-white/15 text-white transition-transform active:scale-95" type="button" @click="navigateBack">
        <span class="material-symbols-outlined text-[1.6rem]">arrow_back</span>
      </button>
      <div class="min-w-0 flex-1 text-center">
        <h1 class="truncate text-[1.1rem] font-black text-white">
          {{ room.title }}
        </h1>
        <p class="text-[0.7rem] text-white/80">{{ selectedVariant.durationLabel }}</p>
      </div>
      <div class="flex items-center gap-2">
        <button class="grid h-9 w-9 place-items-center rounded-full bg-white/15 text-white" type="button">
          <span class="material-symbols-outlined text-[1.2rem]">headphones</span>
        </button>
        <button class="grid h-9 w-9 place-items-center rounded-full bg-white/15 text-white" type="button">
          <span class="material-symbols-outlined text-[1.2rem]">person</span>
        </button>
      </div>
    </header>

    <!-- ===== BALANCE CARD ===== -->
    <div class="mx-3 -mt-0 bg-white rounded-b-[20px] px-4 py-4 shadow-md border border-slate-100">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-[0.7rem] text-slate-500">Số dư ví</p>
          <strong class="text-[1.5rem] font-black text-on-surface">{{ currentBalanceLabel }}đ</strong>
        </div>
        <span class="material-symbols-outlined text-[1.1rem] text-slate-400 cursor-pointer">refresh</span>
      </div>
      <p class="text-[0.68rem] text-slate-400 mt-0.5">Số dư vi</p>
      <div class="grid grid-cols-2 gap-2 mt-3">
        <RouterLink to="/account" class="flex items-center justify-center gap-1.5 rounded-full border-2 border-primary bg-white py-2.5 text-[0.82rem] font-black text-primary active:scale-95 transition-transform">
          Rút tiền
        </RouterLink>
        <RouterLink to="/deposit" class="flex items-center justify-center gap-1.5 rounded-full bg-primary py-2.5 text-[0.82rem] font-black text-white shadow-[0_6px_16px_rgba(255,109,102,0.3)] active:scale-95 transition-transform">
          Nạp tiền
        </RouterLink>
      </div>
    </div>

    <!-- ===== MARQUEE NOTICE ===== -->
    <div class="mx-3 mt-2 flex items-center gap-2 rounded-[12px] bg-white px-3 py-2.5 shadow-sm">
      <span class="material-symbols-outlined text-[1rem] text-primary flex-shrink-0">campaign</span>
      <span class="flex-1 overflow-hidden text-[0.72rem] text-slate-600 whitespace-nowrap truncate">
        Chúc mừng người chơi ***123 vừa thắng 2,500,000đ · Nạp tiền ngay để nhận thưởng VIP
      </span>
      <RouterLink to="/promotion" class="flex-shrink-0 rounded-full bg-primary px-2.5 py-1 text-[0.65rem] font-black text-white">Chi tiết</RouterLink>
    </div>

    <p v-if="roomStateError" class="mx-3 mt-2 rounded-[12px] bg-red-50 px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">
      {{ roomStateError }}
    </p>

    <!-- ===== PERIOD TABS ===== -->
    <div class="mx-0 mt-2 flex bg-white border-b border-slate-100 px-2 py-2 gap-1">
          <button
            v-for="variant in room.variants"
            :key="variant.code"
            type="button"
        class="flex flex-1 flex-col items-center gap-1.5 rounded-[14px] py-2 px-1 transition-all active:scale-[0.97]"
        :class="variant.code === selectedVariant.code ? 'bg-[#e8404a]' : 'bg-transparent'"
            @click="activeVariantCode = variant.code"
          >
        <div
          class="grid h-9 w-9 place-items-center rounded-full border-2 transition-all"
          :class="variant.code === selectedVariant.code ? 'border-white/40 bg-white/20' : 'border-slate-200 bg-slate-50'"
        >
          <span
            class="material-symbols-outlined text-[1.2rem]"
            :class="variant.code === selectedVariant.code ? 'text-white' : 'text-slate-400'"
          >schedule</span>
        </div>
        <span
          class="text-center text-[0.6rem] font-bold leading-tight whitespace-nowrap"
          :class="variant.code === selectedVariant.code ? 'text-white' : 'text-slate-500'"
        >
          {{ room.title }}<br>{{ variant.durationLabel }}
        </span>
      </button>
    </div>

    <!-- ===== PERIOD INFO + COUNTDOWN ===== -->
    <div class="mx-3 mt-2 rounded-[16px] bg-white px-4 py-3 shadow-sm border border-slate-100">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="text-[0.72rem] text-slate-500">Kỳ xổ</span>
          <button
            type="button"
            class="flex items-center gap-1 rounded-full border border-[#f0c0c0] bg-[#fff5f5] px-2.5 py-1 text-[0.65rem] font-semibold text-primary active:scale-95 transition-all"
            @click="showHelpModal = true"
          >
            <span class="material-symbols-outlined text-[0.8rem]">menu_book</span>
            Cách chơi
          </button>
        </div>
        <span class="text-[0.72rem] text-slate-500">Thời gian còn lại</span>
      </div>
      <div class="flex items-center justify-between mt-2">
        <div>
          <p class="text-[0.78rem] font-bold text-on-surface">kỳ hiện tại: {{ currentPeriod?.period_index ?? '—' }}</p>
          <p class="text-[0.65rem] text-slate-400 uppercase tracking-wide mt-0.5">{{ roomStatusLabel }}</p>
        </div>
        <!-- Digit-box countdown (matching source design) -->
        <div class="flex items-center gap-1">
          <div class="flex h-8 w-7 items-center justify-center rounded-[6px] bg-[#1a1a1a] text-[1rem] font-black text-white">{{ isNaN(Number(countdownParts.m0)) ? 0 : countdownParts.m0 }}</div>
          <div class="flex h-8 w-7 items-center justify-center rounded-[6px] bg-[#1a1a1a] text-[1rem] font-black text-white">{{ isNaN(Number(countdownParts.m1)) ? 0 : countdownParts.m1 }}</div>
          <span class="text-[1.1rem] font-black text-[#1a1a1a] px-0.5">:</span>
          <div class="flex h-8 w-7 items-center justify-center rounded-[6px] bg-[#1a1a1a] text-[1rem] font-black text-white">{{ isNaN(Number(countdownParts.s0)) ? 0 : countdownParts.s0 }}</div>
          <div class="flex h-8 w-7 items-center justify-center rounded-[6px] bg-[#1a1a1a] text-[1rem] font-black text-white">{{ isNaN(Number(countdownParts.s1)) ? 0 : countdownParts.s1 }}</div>
        </div>
      </div>
    </div>

    <!-- ===== K3: DICE DISPLAY ===== -->
    <template v-if="isK3">
      <div class="mx-3 mt-2 rounded-[16px] overflow-hidden">
        <div class="flex justify-center gap-4 bg-[#1a5c34] border-[3px] border-[#2d8c4e] rounded-[16px] py-5 px-4">
          <div
            v-for="(d, i) in (isDiceRolling ? rollingDice : currentDice)"
            :key="i"
            class="relative h-[62px] w-[62px] rounded-[14px] transition-all duration-75"
            :class="{ 'animate-bounce': isDiceRolling }"
            :style="{ background: diceColor(d), boxShadow: '0 4px 12px rgba(0,0,0,0.35), inset 0 2px 4px rgba(255,255,255,0.2)' }"
          >
            <span
              v-for="(dot, di) in diceDots(d)"
              :key="di"
              class="absolute h-[10px] w-[10px] rounded-full bg-white"
              :style="{ left: dot.x, top: dot.y, transform: 'translate(-50%, -50%)' }"
            />
          </div>
        </div>
      </div>

      <!-- K3 Recent Results -->
      <div class="mx-3 mt-2 flex items-center gap-2 rounded-[14px] bg-white px-3 py-2.5 shadow-sm border border-slate-100">
        <span class="text-[0.7rem] text-slate-400">Gần đây:</span>
        <div v-for="(result, idx) in recentResults" :key="'k3-recent-' + (result.period_no || idx)" class="flex gap-1">
          <div
            v-for="(d, di) in result.result.split('-').map(Number)"
            :key="di"
            class="flex h-5 w-5 items-center justify-center rounded-[4px] text-[0.6rem] font-black text-white"
            :style="{ background: diceColor(d) }"
          >{{ d }}</div>
        </div>
      </div>
    </template>

    <!-- ===== WINGO: RESULT BALLS + SLOTS ===== -->
    <template v-if="isWingo">
      <div class="mx-3 mt-2 rounded-[16px] bg-white px-4 py-3 shadow-sm border border-slate-100">
        <!-- Recent result balls row -->
        <div class="flex items-center gap-2 mb-3">
          <div
            v-for="(result, idx) in recentResults"
            :key="'wingo-ball-' + (result.period_no || idx)"
            class="flex h-7 w-7 items-center justify-center rounded-full text-[0.75rem] font-black text-white flex-shrink-0"
            :style="{ background: wingoBallBackground(Number(result.result)) }"
          >{{ result.result }}</div>
          <div class="flex h-7 w-7 items-center justify-center rounded-full bg-slate-100 text-slate-400 flex-shrink-0">
            <span class="material-symbols-outlined text-[0.9rem]">chevron_right</span>
          </div>
        </div>

        <!-- Large number display -->
        <div class="flex justify-center gap-3 rounded-[14px] bg-[#f8f0f0] py-4 border border-[#f0e0e0]">
          <div
            v-for="(n, i) in recentResults.slice(0, 2).map(r => Number(r.result))"
            :key="i"
            class="flex h-20 w-20 items-center justify-center rounded-[14px] bg-white border-2 border-[#f0e0e0] shadow-md"
          >
            <span class="text-[3rem] font-black leading-none" :style="wingoNumberTextStyle(n)">
              {{ n }}
            </span>
          </div>
        </div>
      </div>
    </template>

    <!-- ===== BET AREA ===== -->
    <div class="mx-3 mt-2 rounded-[16px] bg-white px-3 py-3 shadow-sm border border-slate-100 relative overflow-hidden">
      <!-- 5s Locked Countdown Overlay -->
      <Transition name="fade">
        <div v-if="showClosingCountdownOverlay" class="absolute inset-0 z-10 flex flex-col items-center justify-center bg-white/60 backdrop-blur-[2px]">
          <div class="relative">
            <svg class="h-28 w-28 -rotate-90">
              <circle
                cx="56" cy="56" r="50"
                stroke="currentColor" stroke-width="8"
                fill="transparent"
                class="text-slate-100"
              />
              <circle
                cx="56" cy="56" r="50"
                stroke="currentColor" stroke-width="8"
                fill="transparent"
                stroke-dasharray="314.159"
                :stroke-dashoffset="314.159 * (1 - closingCountdownSeconds / 5)"
                class="text-primary transition-all duration-1000 linear"
              />
            </svg>
            <div class="absolute inset-0 flex items-center justify-center">
              <span class="text-[3.5rem] font-black italic text-primary drop-shadow-md">{{ closingCountdownSeconds }}</span>
            </div>
          </div>
          <p class="mt-2 text-[0.85rem] font-black uppercase tracking-widest text-[#e8404a] drop-shadow-sm">Đang khóa lệnh</p>
        </div>
      </Transition>

      <!-- K3: Sub-tabs for bet types -->
      <template v-if="isK3">
        <div class="flex gap-1.5 overflow-x-auto pb-2 no-scrollbar mb-3">
          <button
            v-for="tab in k3SubTabs"
            :key="tab"
            type="button"
            class="flex-shrink-0 rounded-full border-[1.5px] px-3.5 py-1.5 text-[0.72rem] font-semibold transition-all"
            :class="activeK3SubTab === tab ? 'border-[#e8404a] bg-[#e8404a] text-white' : 'border-slate-200 bg-white text-slate-500'"
            @click="activeK3SubTab = tab"
          >{{ tab }}</button>
        </div>

        <!-- K3: Options grid for active sub-tab -->
        <div v-for="(group, gIdx) in activeK3Groups" :key="'k3-g-' + (group.title || gIdx)">
          <!-- Grid mode: circles with odds (Tổng số, 2 số trùng, 3 số trùng) -->
          <div v-if="group.mode === 'grid'" class="grid grid-cols-4 gap-2 mb-3">
              <button
                v-for="(option, oIdx) in group.options"
                :key="'k3-opt-' + (option.key || oIdx)"
                type="button"
                class="flex flex-col items-center justify-center aspect-square rounded-full text-white transition-transform active:scale-95 hover:opacity-90 gap-0.5 p-1 disabled:cursor-not-allowed disabled:opacity-50"
                :style="{ background: option.accent }"
                :disabled="!canBet"
                @click="selectOption(group.title, option.key, option.label)"
              >
              <span class="text-[0.9rem] font-black leading-tight">{{ option.label }}</span>
              <span v-if="option.odds" class="text-[0.55rem] font-semibold opacity-85">{{ option.odds }}</span>
            </button>
          </div>

          <!-- Chips mode: Khác số & Lớn/Nhỏ -->
          <div v-else>
            <div v-if="group.title === 'Lớn / Nhỏ / Chẵn / Lẻ'" class="grid grid-cols-4 gap-2 mb-3">
              <button
                v-for="option in group.options"
                :key="option.key"
                type="button"
                class="flex flex-col items-center justify-center rounded-[10px] py-3 text-white font-black text-[0.82rem] transition-all active:scale-95 disabled:cursor-not-allowed disabled:opacity-50"
                :style="{ background: option.accent }"
                :disabled="!canBet"
                @click="selectOption(group.title, option.key, option.label)"
              >
                {{ option.label }}
                <span v-if="option.odds" class="text-[0.6rem] font-semibold opacity-80 mt-0.5">{{ option.odds }}</span>
              </button>
            </div>
            <div v-else class="flex flex-wrap gap-2 mb-3">
              <button
                v-for="(option, oIdx) in group.options"
                :key="'k3-chip-' + (option.key || oIdx)"
                type="button"
                class="rounded-[10px] px-4 py-2.5 text-white text-[0.82rem] font-bold transition-all active:scale-95 flex-1 disabled:cursor-not-allowed disabled:opacity-50"
                :style="{ background: option.accent }"
                :disabled="!canBet"
                @click="selectOption(group.title, option.key, option.label)"
              >
                {{ option.label }}
                <span v-if="option.odds" class="ml-1 text-[0.65rem] opacity-80">({{ option.odds }})</span>
              </button>
            </div>
          </div>
        </div>
      </template>

      <!-- WINGO specific bet UI -->
      <template v-if="isWingo">
        <!-- Color buttons -->
        <div v-if="colorGroup" class="grid grid-cols-3 gap-2 mb-3">
          <button
            v-for="(option, oIdx) in colorGroup.options"
            :key="'color-' + (option.key || oIdx)"
            type="button"
            class="min-h-[48px] rounded-[10px] px-2 py-2 text-[0.9rem] font-black text-white transition-transform active:scale-95 disabled:cursor-not-allowed disabled:opacity-50"
            :style="{ background: option.accent }"
            :disabled="!canBet"
            @click="selectOption(colorGroup.title, option.key, option.label)"
          >
            <span class="block leading-none">{{ option.label }}</span>
            <span v-if="option.odds" class="mt-0.5 block text-[0.58rem] font-semibold opacity-85">{{ option.odds }}</span>
          </button>
        </div>

        <!-- Number balls 0-9 -->
        <div v-if="numberGroup" class="grid grid-cols-5 gap-2 mb-3">
          <button
            v-for="option in numberGroup.options"
            :key="option.key"
            type="button"
            class="flex aspect-square flex-col items-center justify-center rounded-full text-white transition-transform active:scale-95 hover:scale-105 disabled:cursor-not-allowed disabled:opacity-50"
            :style="{ background: option.accent }"
            :disabled="!canBet"
            @click="selectOption(numberGroup.title, option.key, option.label)"
          >
            <span class="text-[1rem] font-black leading-none">{{ option.label }}</span>
            <span v-if="option.odds" class="mt-0.5 text-[0.55rem] font-semibold opacity-85">{{ option.odds }}</span>
          </button>
        </div>

        <!-- Big / Small buttons -->
        <div v-if="bigSmallGroup" class="grid grid-cols-2 gap-2 mb-3">
          <button
            v-for="option in bigSmallGroup.options"
            :key="option.key"
            type="button"
            class="min-h-[52px] rounded-[10px] px-2 py-2 text-[1rem] font-black text-white transition-transform active:scale-95 disabled:cursor-not-allowed disabled:opacity-50"
            :style="{ background: option.accent }"
            :disabled="!canBet"
            @click="selectOption(bigSmallGroup.title, option.key, option.label)"
          >
            <span class="block leading-none">{{ option.label }}</span>
            <span v-if="option.odds" class="mt-0.5 block text-[0.6rem] font-semibold opacity-85">{{ option.odds }}</span>
          </button>
        </div>
      </template>

      <!-- Fallback for other games (5D etc) -->
      <template v-if="!isK3 && !isWingo">
        <div v-for="group in selectedVariant.betGroups" :key="group.title" class="mb-4">
          <p class="text-[0.72rem] font-bold text-slate-500 mb-2 uppercase tracking-wide">{{ group.title }}</p>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="option in group.options"
              :key="option.key"
              type="button"
              class="rounded-full px-4 py-2 text-[0.82rem] font-black text-white transition-transform active:scale-95 disabled:cursor-not-allowed disabled:opacity-50"
              :style="{ background: option.accent }"
              :disabled="!canBet"
              @click="selectOption(group.title, option.key, option.label)"
            >{{ option.label }}</button>
          </div>
        </div>
      </template>

      <!-- Bet status messages -->
      <div v-if="!canPlay" class="rounded-[12px] bg-red-50 px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">
        Số dư khả dụng không đủ để đặt cược. Vui lòng nạp thêm tiền hoặc chờ hệ thống giải phóng số dư đang bị khóa.
      </div>
      <div v-else-if="isBetLocked" class="rounded-[12px] bg-amber-50 px-4 py-3 text-[0.78rem] font-semibold text-amber-700">
        Kỳ hiện tại đã bước vào 5 giây cuối hoặc đã khóa lệnh. Vui lòng chờ kỳ tiếp theo.
      </div>
      <p v-if="joinError" class="mt-2 rounded-[12px] bg-red-50 px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">{{ joinError }}</p>
      <p v-if="visibleBetMessage" class="mt-2 rounded-[12px] bg-[rgba(255,109,102,0.08)] px-4 py-3 text-[0.78rem] font-semibold text-primary">{{ visibleBetMessage }}</p>
    </div>

    <PlayHistorySection
      :active-history-tab="activeHistoryTab"
      :chart-series="chartSeries"
      :chart-max-value="chartMaxValue"
      :history-rows="historyRows"
      :mine-rows="mineRows"
      :history-loading="historyLoading"
      :mine-loading="mineLoading"
      :chart-loading="chartLoading"
      :history-error="historyError"
      :mine-error="mineError"
      :chart-error="chartError"
      :history-page="historyPage"
      :history-total-pages="historyTotalPages"
      :mine-page="minePage"
      :mine-total-pages="mineTotalPages"
      :is-k3="isK3"
      @change-tab="activeHistoryTab = $event"
      @refresh-chart="void loadChartHistory()"
      @refresh-history="void loadRoomHistory(historyPage)"
      @page-change="handleHistoryPageChange"
      @open-ticket-detail="openTicketDetail"
    />
    </div>

  <!-- ===== BET MODAL (Slide-up sheet) ===== -->
    <Teleport to="body">
      <div
        v-if="showBetModal"
        class="fixed inset-0 z-50 flex items-end"
        @click.self="showBetModal = false"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showBetModal = false" />

        <!-- Sheet -->
        <div class="relative w-full rounded-t-[24px] bg-white px-5 pt-4 pb-8 shadow-2xl" style="max-height: 85dvh; overflow-y: auto;">
          <!-- Handle -->
          <div class="mx-auto mb-4 h-1 w-10 rounded-full bg-slate-200" />

        <h3 class="mb-4 text-[1rem] font-black text-on-surface">Đặt cược</h3>

        <div class="mb-3 flex items-center justify-between rounded-[14px] bg-[#fff5f5] px-4 py-3">
          <span class="text-[0.82rem] text-slate-500">Lựa chọn</span>
          <span class="text-[0.9rem] font-black text-[#e8404a]">{{ modalBetLabel }}</span>
        </div>

        <div class="mb-5 flex items-center justify-between rounded-[14px] bg-slate-50 px-4 py-4">
          <span class="text-[0.82rem] text-slate-500">Tổng tiền cược</span>
          <div class="flex items-center gap-3">
            <button
              type="button"
              class="flex h-9 w-9 items-center justify-center rounded-full border border-slate-200 text-[1.2rem] font-bold text-slate-500 transition-all active:scale-90 disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="!canAdjustBetAmount(-1)"
              @click="adjustBetAmount(-1)"
            >−</button>
            <div class="flex flex-col items-center">
              <input
                type="number"
                v-model.number="modalBetAmount"
                :min="1000"
                :max="availableVndBalance"
                class="w-[120px] text-center text-[1.1rem] font-black text-on-surface bg-transparent border-none focus:ring-0"
              />
              <span class="text-[0.6rem] text-slate-400">VNĐ</span>
            </div>
            <button
              type="button"
              class="flex h-9 w-9 items-center justify-center rounded-full border border-slate-200 text-[1.2rem] font-bold text-slate-500 transition-all active:scale-90 disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="!canAdjustBetAmount(1)"
              @click="adjustBetAmount(1)"
            >+</button>
          </div>
        </div>

        <div class="mb-5">
          <div class="flex items-center justify-between mb-2 px-1">
            <span class="text-[0.72rem] font-bold text-slate-400 uppercase tracking-widest">Chọn nhanh mức cộng</span>
            <span class="text-[0.75rem] font-black text-primary">{{ formatMoney(modalBetAmount) }}đ</span>
          </div>
          <div class="grid grid-cols-4 gap-2">
            <button
              v-for="amount in betAmountPresets"
              :key="'amount-' + amount"
              type="button"
              class="rounded-[10px] border-[1.5px] px-2 py-2 text-[0.75rem] font-black transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              :class="canAddAmount(amount) ? 'border-slate-200 bg-white text-slate-500 active:scale-95' : 'border-slate-200 bg-slate-50 text-slate-300'"
              :disabled="!canAddAmount(amount)"
              @click="addBetAmount(amount)"
            >+{{ formatMoney(amount) }}</button>
          </div>
        </div>

        <!-- Confirm / Cancel -->
        <div class="grid grid-cols-2 gap-3">
          <button
            class="min-h-[50px] rounded-[14px] border-2 border-slate-200 bg-white text-[0.9rem] font-black text-slate-500 transition-all active:scale-95"
            @click="showBetModal = false"
          >Hủy bỏ</button>
          <button
            class="min-h-[50px] rounded-[14px] bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] text-[0.9rem] font-black text-white shadow-[0_8px_16px_rgba(229,46,46,0.25)] transition-all active:scale-95 disabled:opacity-50"
            :disabled="betLoading || !canBet"
            @click="confirmBet"
          >{{ betLoading ? 'Đang gửi...' : 'Xác nhận' }}</button>
        </div>
        </div>
      </div>
    </Teleport>

  <!-- ===== RESULT MODAL ===== -->
    <Teleport to="body">
      <div
        v-if="showResultModal"
        class="fixed inset-0 z-[60] flex items-end"
        @click.self="closeResultModal"
      >
        <div class="absolute inset-0 bg-black/45 backdrop-blur-sm" @click="closeResultModal" />

        <div class="relative w-full rounded-t-[24px] bg-white px-5 pt-4 pb-8 shadow-2xl" style="max-height: 85dvh; overflow-y: auto;">
          <div class="mx-auto mb-4 h-1 w-10 rounded-full bg-slate-200" />

        <div class="mb-4 flex items-center justify-between">
          <div>
            <p class="text-[0.68rem] uppercase tracking-[0.12em] text-slate-400">Kết quả kỳ</p>
            <h3 class="text-[1rem] font-black text-on-surface">{{ resultModalTitle }}</h3>
          </div>
          <button class="grid h-9 w-9 place-items-center rounded-full bg-slate-100 text-slate-500" type="button" @click="closeResultModal">
            <span class="material-symbols-outlined text-[1rem]">close</span>
          </button>
        </div>

        <div
          class="mb-4 rounded-[18px] px-4 py-4 text-white"
          :class="resultModalTone === 'win' ? 'bg-gradient-to-r from-[#24b561] to-[#1f9d52]' : resultModalTone === 'lose' ? 'bg-gradient-to-r from-[#e64545] to-[#c92d2d]' : 'bg-gradient-to-r from-[#f6c32d] to-[#e29f00]'"
        >
          <p class="text-[0.72rem] uppercase tracking-[0.12em] text-white/70">Tiền biến động</p>
          <strong class="block text-[1.7rem] font-black">{{ resultModalAmount }}</strong>
          <p class="mt-1 text-[0.78rem] text-white/90">{{ resultModalDescription }}</p>
        </div>

        <div class="grid grid-cols-2 gap-2 text-[0.78rem]">
          <div class="rounded-[14px] bg-[#fff9f9] px-3 py-3">
            <p class="text-slate-400">Kỳ xổ</p>
            <strong class="block text-on-surface">{{ resultModalPeriodNo }}</strong>
          </div>
          <div class="rounded-[14px] bg-[#fff9f9] px-3 py-3">
            <p class="text-slate-400">Tiền vào lệnh</p>
            <strong class="block text-on-surface">{{ resultModalStake }}đ</strong>
          </div>
          <div class="rounded-[14px] bg-[#fff9f9] px-3 py-3">
            <p class="text-slate-400">Tiền nhận</p>
            <strong class="block text-on-surface">{{ resultModalPayout }}đ</strong>
          </div>
          <div class="rounded-[14px] bg-[#fff9f9] px-3 py-3">
            <p class="text-slate-400">Thời gian chốt</p>
            <strong class="block text-on-surface">{{ formatVietnamTime(resultModalSettledAt) }}</strong>
          </div>
        </div>

        <button
          type="button"
          class="mt-4 min-h-[48px] w-full rounded-[16px] bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] text-[0.9rem] font-black text-white shadow-[0_8px_16px_rgba(229,46,46,0.25)]"
          @click="closeResultModal"
        >
          Đã hiểu
        </button>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="showTicketDetailModal && selectedTicketDetail"
        class="fixed inset-0 z-[70] flex items-end"
        @click.self="closeTicketDetail"
      >
        <div class="absolute inset-0 bg-black/45 backdrop-blur-sm" @click="closeTicketDetail" />

        <div class="relative w-full rounded-t-[24px] bg-white px-5 pt-4 pb-8 shadow-2xl" style="max-height: 85dvh; overflow-y: auto;">
          <div class="mx-auto mb-4 h-1 w-10 rounded-full bg-slate-200" />

          <div class="mb-4 flex items-center justify-between">
            <div>
              <p class="text-[0.68rem] uppercase tracking-[0.12em] text-slate-400">Chi tiết giao dịch</p>
              <h3 class="text-[1rem] font-black text-on-surface">{{ selectedTicketDetail.period_no }}</h3>
            </div>
            <button class="grid h-9 w-9 place-items-center rounded-full bg-slate-100 text-slate-500" type="button" @click="closeTicketDetail">
              <span class="material-symbols-outlined text-[1rem]">close</span>
            </button>
          </div>

          <div class="mb-4 rounded-[16px] bg-[#fff9f9] px-4 py-4">
            <p class="text-[0.72rem] uppercase tracking-[0.12em] text-slate-400">Kết quả lệnh</p>
            <p class="mt-1 text-[1.2rem] font-black text-on-surface">{{ rowMainLabel(selectedTicketDetail) }}</p>
            <p class="mt-1 text-[0.78rem] text-slate-500">{{ rowSubLabel(selectedTicketDetail) }}</p>
            <p class="mt-2 text-[0.85rem] font-semibold" :class="rowStatusClass(selectedTicketDetail)">
              {{ rowStatusText(selectedTicketDetail) }}
            </p>
          </div>

          <div class="grid grid-cols-2 gap-2 text-[0.78rem]">
            <div class="rounded-[14px] bg-[#fff9f9] px-3 py-3">
              <p class="text-slate-400">Tiền gốc</p>
              <strong class="block text-on-surface">{{ formatMoney(rowOriginalAmountValue(selectedTicketDetail)) }}đ</strong>
            </div>
            <div class="rounded-[14px] bg-[#fff9f9] px-3 py-3">
              <p class="text-slate-400">Thuế 2%</p>
              <strong class="block text-on-surface">{{ formatMoney(rowTaxAmountValue(selectedTicketDetail)) }}đ</strong>
            </div>
            <div class="rounded-[14px] bg-[#fff9f9] px-3 py-3">
              <p class="text-slate-400">Tiền tính thưởng</p>
              <strong class="block text-on-surface">{{ formatMoney(rowNetAmountValue(selectedTicketDetail)) }}đ</strong>
            </div>
            <div class="rounded-[14px] bg-[#fff9f9] px-3 py-3">
              <p class="text-slate-400">Tiền cộng về ví</p>
              <strong class="block text-on-surface">
                {{ rowStatusValue(selectedTicketDetail) === 'WON' ? `+${formatMoney(rowWinCreditValue(selectedTicketDetail))}đ` : `${formatMoney(rowWinCreditValue(selectedTicketDetail))}đ` }}
              </strong>
            </div>
            <div class="rounded-[14px] bg-[#fff9f9] px-3 py-3">
              <p class="text-slate-400">Lãi/lỗ</p>
              <strong class="block text-on-surface">{{ formatSignedMoney(rowProfitLossValue(selectedTicketDetail)) }}</strong>
            </div>
            <div class="rounded-[14px] bg-[#fff9f9] px-3 py-3">
              <p class="text-slate-400">Chốt lúc</p>
              <strong class="block text-on-surface">{{ formatVietnamTime(selectedTicketDetail.settled_at) }}</strong>
            </div>
          </div>

          <button
            type="button"
            class="mt-4 min-h-[48px] w-full rounded-[16px] bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] text-[0.9rem] font-black text-white shadow-[0_8px_16px_rgba(229,46,46,0.25)]"
            @click="closeTicketDetail"
          >
            Đóng
          </button>
        </div>
      </div>
    </Teleport>

    <div v-if="room && !selectedVariant" class="flex min-h-dvh flex-col items-center justify-center gap-4 px-5 text-center">
    <div class="grid h-20 w-20 place-items-center rounded-[24px] bg-white shadow-[0_10px_24px_rgba(255,109,102,0.12)]">
      <span class="material-symbols-outlined text-[2.2rem] text-primary">schedule</span>
    </div>
    <div class="max-w-sm space-y-2">
      <h2 class="text-[1.25rem] font-black text-on-surface">{{ room.title }} đang sắp mở</h2>
    </div>
    <div class="flex w-full max-w-sm gap-3">
      <button
        type="button"
        class="flex-1 rounded-[16px] border-2 border-slate-200 bg-white px-4 py-3 text-[0.9rem] font-black text-slate-600"
        @click="navigateBack"
      >
        Quay lại
      </button>
      <RouterLink
        to="/play"
        class="flex-1 rounded-[16px] bg-primary px-4 py-3 text-center text-[0.9rem] font-black text-white"
      >
        Phòng chơi
      </RouterLink>
    </div>
    </div>

    <div v-if="!room" class="flex min-h-dvh flex-col items-center justify-center gap-4 px-5 text-center">
    <div class="grid h-20 w-20 place-items-center rounded-[24px] bg-white shadow-[0_10px_24px_rgba(255,109,102,0.12)]">
      <span class="material-symbols-outlined text-[2.2rem] text-primary">search_off</span>
    </div>
    <div class="max-w-sm space-y-2">
      <h2 class="text-[1.25rem] font-black text-on-surface">Không tìm thấy phòng chơi</h2>
      <p class="text-[0.9rem] leading-6 text-slate-500">
        Phòng bạn chọn chưa tồn tại hoặc đã bị ẩn. Mình đưa bạn quay lại danh sách phòng an toàn hơn.
      </p>
    </div>
    <div class="flex w-full max-w-sm gap-3">
      <button
        type="button"
        class="flex-1 rounded-[16px] border-2 border-slate-200 bg-white px-4 py-3 text-[0.9rem] font-black text-slate-600"
        @click="navigateBack"
      >
        Quay lại
      </button>
      <RouterLink
        to="/play"
        class="flex-1 rounded-[16px] bg-primary px-4 py-3 text-center text-[0.9rem] font-black text-white"
      >
        Phòng chơi
      </RouterLink>
    </div>
    </div>
    <Teleport to="body">
      <div v-if="showHelpModal" class="fixed inset-0 z-[100] flex items-center justify-center p-5">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showHelpModal = false" />
        <div class="relative w-full max-w-sm rounded-[24px] bg-white p-6 shadow-2xl">
          <div class="mb-4 flex items-center justify-between">
            <h3 class="text-[1.1rem] font-black text-on-surface">
              {{ isK3 ? gameHowToPlay.k3.title : (room?.code === 'lottery' ? gameHowToPlay.lottery.title : gameHowToPlay.wingo.title) }}
            </h3>
            <button class="grid h-8 w-8 place-items-center rounded-full bg-slate-100 text-slate-500" @click="showHelpModal = false">
              <span class="material-symbols-outlined text-[0.9rem]">close</span>
            </button>
          </div>
          <div class="max-h-[60dvh] overflow-y-auto pr-2 text-[0.85rem] leading-relaxed text-slate-600">
            <div class="whitespace-pre-line">
              {{ isK3 ? gameHowToPlay.k3.content : (room?.code === 'lottery' ? gameHowToPlay.lottery.content : gameHowToPlay.wingo.content) }}
            </div>
          </div>
          <button
            type="button"
            class="mt-6 h-12 w-full rounded-[16px] bg-primary font-bold text-white shadow-lg shadow-primary/20"
            @click="showHelpModal = false"
          >
            Đã hiểu
          </button>
        </div>
      </div>
    </Teleport>
  </div>
</template>
