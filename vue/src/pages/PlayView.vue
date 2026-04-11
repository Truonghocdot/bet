<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { request, type ApiError } from '@/shared/api/http'
import { connectEventStream, type StreamConnection } from '@/shared/api/stream'
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
const selectedMultiplier = ref(10)
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
const serverTimeOffsetMs = ref(0)
const clockTick = ref(Date.now())
const seenSettlementPeriods = new Set<string>()
const countdownTargetMs = ref(0)
const countdownTargetPeriodNo = ref('')
const countdownCachePrefix = 'ff789:play-countdown:'

// Bet modal state
const showBetModal = ref(false)
const modalBetLabel = ref('')
const modalBetKey = ref('')
const modalBetGroupTitle = ref('')
const baseChipAmount = 1000
const modalBetAmount = ref(baseChipAmount * 10)
const betPresets = [1000, 5000, 10000, 50000, 100000]

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

const stakeOptions = [1, 5, 10, 20, 50] as const
const tablePageSize = 4
let timer: number | undefined
let roomStreamConnection: StreamConnection | null = null

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
const walletUsdt = computed(() => wallet.wallets.find((item) => item.unit === 2) ?? null)
const availableVndBalance = computed(() => {
  const balance = Number(walletVnd.value?.balance ?? 0)
  const locked = Number(walletVnd.value?.locked_balance ?? 0)
  return Math.max(0, balance - locked)
})
const canPlay = computed(() => availableVndBalance.value > 0)
const currentPeriod = computed(() => roomState.value?.current_period ?? null)
const syncedNow = computed(() => clockTick.value + serverTimeOffsetMs.value)
const expectedPeriodSeconds = computed(() => selectedVariant.value?.countdownSeconds ?? 0)
const currentPeriodBetLockAtMs = computed(() => currentPeriod.value ? new Date(currentPeriod.value.bet_lock_at).getTime() : 0)
const visibleBetMessage = computed(() => {
  if (!betMessage.value) return ''
  if (!selectedRoomCode.value) return ''
  return betMessageRoomCode.value === selectedRoomCode.value ? betMessage.value : ''
})

function resetTransientRoomUiState() {
  joinError.value = ''
  betMessage.value = ''
  betMessageRoomCode.value = ''
}

function countdownCacheKey(roomCode: string) {
  return `${countdownCachePrefix}${roomCode}`
}

function readCountdownCache(roomCode: string) {
  if (typeof window === 'undefined' || !roomCode) return null
  try {
    const raw = window.sessionStorage.getItem(countdownCacheKey(roomCode))
    if (!raw) return null
    const parsed = JSON.parse(raw) as { period_no?: string; target_ms?: number }
    if (!parsed?.period_no || !Number.isFinite(parsed.target_ms)) return null
    return { periodNo: String(parsed.period_no), targetMs: Number(parsed.target_ms) }
  } catch {
    return null
  }
}

function writeCountdownCache(roomCode: string, periodNo: string, targetMs: number) {
  if (typeof window === 'undefined' || !roomCode || !periodNo || !Number.isFinite(targetMs)) return
  try {
    window.sessionStorage.setItem(
      countdownCacheKey(roomCode),
      JSON.stringify({ period_no: periodNo, target_ms: targetMs }),
    )
  } catch {
    // ignore storage errors
  }
}

function clearCountdownCache(roomCode: string) {
  if (typeof window === 'undefined' || !roomCode) return
  try {
    window.sessionStorage.removeItem(countdownCacheKey(roomCode))
  } catch {
    // ignore storage errors
  }
}

function syncCountdownTarget(period: PlayRoomStateResponse['current_period'] | null, nowMs = Date.now(), force = false) {
  if (!period) {
    countdownTargetMs.value = 0
    countdownTargetPeriodNo.value = ''
    return
  }

  const periodNo = String(period.period_no ?? '')
  const rawDrawAtMs = Number(new Date(period.draw_at).getTime())
  const expectedSeconds = Math.max(1, expectedPeriodSeconds.value || 30)
  const fallbackTargetMs = nowMs + expectedSeconds * 1000
  const maxReasonableMs = nowMs + Math.max(expectedSeconds * 3, 30) * 1000
  const cached = readCountdownCache(selectedRoomCode.value)

  if (!periodNo) {
    countdownTargetMs.value = fallbackTargetMs
    countdownTargetPeriodNo.value = ''
    return
  }

  if (!force && countdownTargetPeriodNo.value === periodNo && countdownTargetMs.value > 0) {
    return
  }

  if (cached?.periodNo === periodNo && cached.targetMs > nowMs) {
    countdownTargetMs.value = cached.targetMs
    countdownTargetPeriodNo.value = periodNo
    return
  }

  if (Number.isFinite(rawDrawAtMs) && rawDrawAtMs > 0 && rawDrawAtMs <= maxReasonableMs) {
    countdownTargetMs.value = rawDrawAtMs
  } else {
    countdownTargetMs.value = fallbackTargetMs
  }
  countdownTargetPeriodNo.value = periodNo
  writeCountdownCache(selectedRoomCode.value, periodNo, countdownTargetMs.value)
}
const isBetLocked = computed(() => {
  if (!currentPeriod.value) return true
  if ((currentPeriod.value.status || '').toUpperCase() !== 'OPEN') return true
  return syncedNow.value >= currentPeriodBetLockAtMs.value
})
const canBet = computed(() => canPlay.value && !isBetLocked.value)

const remainingSeconds = computed(() => {
  if (!currentPeriod.value || countdownTargetMs.value <= 0) return 0
  return Math.max(0, Math.ceil((countdownTargetMs.value - syncedNow.value) / 1000))
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

const stakeAmount = computed(() => String(1000 * selectedMultiplier.value))
const stakeLabel = computed(() => formatMoney(Number(stakeAmount.value)))
const currentBalanceLabel = computed(() => formatMoney(walletVnd.value?.balance ?? 0))
const lockedBalanceLabel = computed(() => formatMoney(walletVnd.value?.locked_balance ?? 0))

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
const periodHistory = computed(() => {
  if (activeHistoryTab.value === 'mine') return mineRows.value
  return historyRows.value
})

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
  const expected = Math.max(0, (rowOriginalAmountValue(row) * 2) - rowTaxAmountValue(row))
  const actual = toFiniteNumber(row.actual_payout)
  if (rowStatusValue(row) === 'WON') {
    if (expected > 0) return expected
    return actual
  }
  return actual
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

function currentTotalPages() {
  return activeHistoryTab.value === 'mine' ? mineTotalPages.value : historyTotalPages.value
}

function setCurrentPage(page: number) {
  if (activeHistoryTab.value === 'mine') {
    minePage.value = page
    return
  }
  historyPage.value = page
}

function ensureDefaultSelections(variant: PlayVariant | null) {
  Object.keys(selectedOptions).forEach((key) => delete selectedOptions[key])
  variant?.betGroups.forEach((group) => {
    if (group.options[0]) {
      selectedOptions[group.title] = group.options[0].key
    }
  })
  selectedMultiplier.value = 10
  modalBetAmount.value = baseChipAmount * selectedMultiplier.value
}

async function loadWallet() {
  if (!auth.isAuthenticated) return
  try {
    await wallet.fetchSummary()
  } catch {
    // state already stores the error
  }
}

function applyServerClock(serverTime: string, requestMidpoint = Date.now()) {
  const serverTimeMs = new Date(serverTime).getTime()
  if (!Number.isFinite(serverTimeMs) || serverTimeMs <= 0) return
  serverTimeOffsetMs.value = serverTimeMs - requestMidpoint
  clockTick.value = Date.now()
}

async function applyRoomStateResponse(
  response: PlayRoomStateResponse,
  options: {
    requestStartedAt?: number
    requestFinishedAt?: number
    forceRebaseClock?: boolean
  } = {},
) {
  const previousPeriod = roomState.value?.current_period ?? null
  const previousPeriodNo = previousPeriod?.period_no ?? ''
  const nextPeriodNo = response.current_period?.period_no ?? ''
  const shouldRebaseClock = options.forceRebaseClock || !roomState.value || previousPeriodNo !== nextPeriodNo

  roomState.value = response

  if (shouldRebaseClock) {
    const midpoint =
      options.requestStartedAt !== undefined && options.requestFinishedAt !== undefined
        ? options.requestStartedAt + Math.max(0, Math.floor((options.requestFinishedAt - options.requestStartedAt) / 2))
        : Date.now()
    applyServerClock(response.server_time, midpoint)
  }

  syncCountdownTarget(response.current_period, shouldRebaseClock ? clockTick.value : Date.now(), shouldRebaseClock)
  await maybeShowSettlementModal(previousPeriod, response.current_period)

  if (previousPeriodNo && previousPeriodNo !== nextPeriodNo) {
    if (activeHistoryTab.value === 'chart') {
      void loadChartHistory()
    } else if (activeHistoryTab.value === 'history') {
      void loadRoomHistory(1)
    }
  }
}

function disconnectRoomStateStream() {
  roomStreamConnection?.close()
  roomStreamConnection = null
}

function connectRoomStateStream(roomCode: string) {
  if (!roomCode) return

  disconnectRoomStateStream()
  roomStreamConnection = connectEventStream(`/v1/play/rooms/${roomCode}/stream`, {
    reconnectMs: 2500,
    onEvent(payload) {
      if (payload.event === 'room.clock') {
        applyServerClock(String(payload.data?.server_time ?? ''))
        return
      }

      if (payload.event === 'room.state') {
        roomStateError.value = ''
        void applyRoomStateResponse(payload.data as PlayRoomStateResponse)
      }
    },
    onError(errorValue) {
      const err = errorValue as ApiError
      roomStateError.value = err?.message ?? 'Kết nối realtime phòng chơi đang được nối lại'
    },
  })
}

async function loadRoomState(roomCode = selectedRoomCode.value) {
  if (!roomCode) return
  roomStateLoading.value = true
  roomStateError.value = ''
  try {
    const startedAt = Date.now()
    const response = await request<PlayRoomStateResponse>('GET', `/v1/play/rooms/${roomCode}/state`)
    const finishedAt = Date.now()
    await applyRoomStateResponse(response, {
      requestStartedAt: startedAt,
      requestFinishedAt: finishedAt,
    })
  } catch (error: unknown) {
    const err = error as ApiError
    roomStateError.value = err?.message ?? 'Không thể tải trạng thái phòng chơi'
    roomState.value = null
  } finally {
    roomStateLoading.value = false
  }
}

async function loadRoomHistory(page = historyPage.value) {
  if (!selectedRoomCode.value) return
  historyLoading.value = true
  historyError.value = ''
  try {
    const response = await request<PlayRoomHistoryResponse>(
      'GET',
      `/v1/play/rooms/${selectedRoomCode.value}/history?page=${page}&page_size=${tablePageSize}`,
      {},
    )
    historyRows.value = response.items
    historyPage.value = response.page
    historyTotalPages.value = response.total_pages
  } catch (error: unknown) {
    const err = error as ApiError
    historyError.value = err?.message ?? 'Không thể tải lịch sử game'
    historyRows.value = []
  } finally {
    historyLoading.value = false
  }
}

async function loadChartHistory() {
  if (!selectedRoomCode.value) return
  chartLoading.value = true
  chartError.value = ''
  try {
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

async function loadMineHistory(page = minePage.value) {
  if (!selectedRoomCode.value || !auth.accessToken) return
  mineLoading.value = true
  mineError.value = ''
  try {
    const response = await request<PlayRoomBetHistoryResponse>(
      'GET',
      `/v1/play/rooms/${selectedRoomCode.value}/bets?page=${page}&page_size=${tablePageSize}`,
      { token: auth.accessToken },
    )
    mineRows.value = response.items
    minePage.value = response.page
    mineTotalPages.value = response.total_pages
  } catch (error: unknown) {
    const err = error as ApiError
    mineError.value = err?.message ?? 'Không thể tải lịch sử cược'
    mineRows.value = []
  } finally {
    mineLoading.value = false
  }
}

async function maybeShowSettlementModal(
  previousPeriod: PlayRoomStateResponse['current_period'] | null,
  nextPeriod: PlayRoomStateResponse['current_period'] | null,
) {
  if (!nextPeriod || !auth.accessToken) return

  const nextPeriodNo = nextPeriod.period_no
  const targetPeriodNo =
    nextPeriodNo && previousPeriod && previousPeriod.period_no !== nextPeriodNo
      ? previousPeriod.period_no
      : String(nextPeriod.status ?? '').toUpperCase() === 'SETTLED'
        ? nextPeriodNo
        : ''

  if (!targetPeriodNo || seenSettlementPeriods.has(targetPeriodNo)) {
    return
  }

  await loadMineHistory(1)
  const settledRow = mineRows.value.find((row) => row.period_no === targetPeriodNo)
  if (!settledRow) return
  if (!['WON', 'LOST', 'VOID', 'HALF_WON', 'HALF_LOST', 'CANCELED', 'CASHED_OUT'].includes(String(settledRow.status ?? '').toUpperCase())) {
    return
  }

  try {
    await wallet.fetchSummary()
  } catch {
    // ignore wallet refresh error for modal flow
  }

  seenSettlementPeriods.add(targetPeriodNo)
  resultModalPeriodNo.value = settledRow.period_no
  resultModalTitle.value = settledRow.status === 'WON' ? 'Kỳ này bạn đã thắng' : 'Kỳ này đã kết quả'
  resultModalDescription.value = settledRow.status === 'WON'
    ? 'Tiền thắng đã cộng về số dư ví.'
    : settledRow.status === 'LOST'
      ? 'Lệnh của bạn chưa trúng kỳ này.'
      : 'Kết quả đã được cập nhật.'
  resultModalTone.value = settledRow.status === 'WON' ? 'win' : settledRow.status === 'LOST' ? 'lose' : 'draw'
  resultModalStake.value = formatMoney(rowOriginalAmountValue(settledRow))
  resultModalPayout.value = formatMoney(rowWinCreditValue(settledRow))
  if (rowStatusValue(settledRow) === 'WON') {
    resultModalAmount.value = `+${formatMoney(rowWinCreditValue(settledRow))}đ`
  } else if (rowStatusValue(settledRow) === 'LOST') {
    resultModalAmount.value = `-${formatMoney(rowOriginalAmountValue(settledRow))}đ`
  } else {
    resultModalAmount.value = formatSignedMoney(rowProfitLossValue(settledRow))
  }
  resultModalSettledAt.value = settledRow.settled_at ?? settledRow.created_at
  showResultModal.value = true
}

async function loadActiveHistory(page = currentPage()) {
  if (activeHistoryTab.value === 'mine') {
    await loadMineHistory(page)
    return
  }
  if (activeHistoryTab.value === 'chart') {
    await loadChartHistory()
    return
  }
  await loadRoomHistory(page)
}

async function joinRoom() {
  if (!room.value || room.value.status !== 'OPEN' || !auth.accessToken) return
  if (!wallet.wallets.length) await loadWallet()
  if (availableVndBalance.value <= 0) {
    joinError.value = 'Số dư không đủ để vào phòng chơi. Vui lòng nạp tiền.'
    return
  }
  joinLoading.value = true
  joinError.value = ''
  connectionId.value = ''
  try {
    const res = await request<GameJoinResponse>('POST', `/v1/games/${room.value.code}/join`, {
      token: auth.accessToken,
    })
    connectionId.value = res.connection_id
  } catch (error: unknown) {
    const err = error as ApiError
    joinError.value = err?.message ?? 'Không thể vào phòng'
  } finally {
    joinLoading.value = false
  }
}

function openBetModal(groupTitle: string, optionKey: string, optionLabel: string) {
  if (!canBet.value) return
  modalBetGroupTitle.value = groupTitle
  modalBetKey.value = optionKey
  modalBetLabel.value = optionLabel
  modalBetAmount.value = baseChipAmount * selectedMultiplier.value
  showBetModal.value = true
}

function applyChipMultiplier(multiplier: number) {
  selectedMultiplier.value = multiplier
  modalBetAmount.value = baseChipAmount * multiplier
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

async function confirmBet() {
  if (!room.value || !selectedVariant.value || !auth.accessToken || !connectionId.value) {
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

  try {
    const requestId = globalThis.crypto?.randomUUID?.() ?? `req-${Date.now()}`
    const res = await request<PlayRoomBetResponse>('POST', `/v1/play/rooms/${selectedRoomCode.value}/bets`, {
      token: auth.accessToken,
      headers: { 'X-Connection-ID': connectionId.value },
      body: {
        request_id: requestId,
        period_id: String(currentPeriod.value.id),
        items: [{
          option_type: groupTypeKey({ title: modalBetGroupTitle.value, description: '', mode: 'chips', options: [] }),
          option_key: modalBetKey.value,
          stake: String(modalBetAmount.value),
        }],
      },
    })
    betMessage.value = res.message || 'Lệnh đã được tiếp nhận'
    betMessageRoomCode.value = selectedRoomCode.value
    await wallet.fetchSummary()
    await loadActiveHistory()
  } catch (error: unknown) {
    const err = error as ApiError
    betMessage.value = err?.message ?? 'Không thể gửi lệnh cược'
    betMessageRoomCode.value = selectedRoomCode.value
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
  if (row.big_small) return row.big_small
  if (row.color) return row.color
  return row.result || '—'
}

function chartBarClass(row: PlayRoomHistoryResponse['items'][number]) {
  const label = String(row.big_small || row.color || row.result || '').toLowerCase()
  if (label.includes('xanh')) return 'bg-[#24b561]'
  if (label.includes('đỏ')) return 'bg-[#e64545]'
  if (label.includes('tím')) return 'bg-[#8b5cf6]'
  if (label.includes('lớn') || label.includes('big')) return 'bg-primary'
  if (label.includes('nhỏ') || label.includes('small')) return 'bg-[#3b82f6]'
  return 'bg-slate-400'
}

function openResultModal() {
  showResultModal.value = true
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

function resultDotClass(label: string) {
  if (label.toLowerCase().includes('green_violet')) return 'bg-[#24b561]'
  if (label.toLowerCase().includes('red_violet')) return 'bg-[#e64545]'
  if (label.toLowerCase().includes('xanh')) return 'bg-[#24b561]'
  if (label.toLowerCase().includes('đỏ')) return 'bg-[#e64545]'
  if (label.toLowerCase().includes('tím')) return 'bg-[#8b5cf6]'
  return 'bg-primary'
}

function resultBadgeClass(label: string) {
  if (label.toLowerCase().includes('green_violet') || label.toLowerCase().includes('red_violet')) {
    return 'border-transparent text-white'
  }
  if (label.toLowerCase().includes('xanh')) return 'border-[#24b561] bg-[#24b561] text-white'
  if (label.toLowerCase().includes('đỏ')) return 'border-[#e64545] bg-[#e64545] text-white'
  if (label.toLowerCase().includes('tím')) return 'border-[#8b5cf6] text-white'
  return 'border-primary bg-primary text-white'
}

function resultBadgeStyle(label: string) {
  if (label.toLowerCase().includes('red_violet')) {
    return { background: 'linear-gradient(135deg, #e64545, #8b5cf6)' }
  }
  if (label.toLowerCase().includes('green_violet')) {
    return { background: 'linear-gradient(135deg, #24b561, #8b5cf6)' }
  }
  if (label.toLowerCase().includes('tím')) {
    return { background: 'linear-gradient(135deg, #8b5cf6, #e8404a)' }
  }
  return {}
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

function wingoTicketBallBackground(row: PlayRoomBetHistoryResponse['items'][number]) {
  const numberFromResult = extractWingoNumber(row.result)
  if (numberFromResult !== null) return wingoBallBackground(numberFromResult)

  const byColor = normalizeBetLabel(row.color)
  if (byColor.includes('Tím')) return 'linear-gradient(135deg, #8b5cf6, #e8404a)'
  if (byColor.includes('Xanh')) return '#24b561'
  if (byColor.includes('Đỏ')) return '#e64545'

  const main = rowMainLabel(row)
  if (main.includes('Lớn')) return '#f6c32d'
  if (main.includes('Nhỏ')) return '#24b561'
  return '#94a3b8'
}

function wingoTicketBallText(row: PlayRoomBetHistoryResponse['items'][number]) {
  const numberFromResult = extractWingoNumber(row.result)
  if (numberFromResult !== null) return String(numberFromResult)

  const main = rowMainLabel(row)
  if (main.includes('Xanh')) return 'X'
  if (main.includes('Đỏ')) return 'Đ'
  if (main.includes('Tím')) return 'T'
  if (main.includes('Lớn')) return 'L'
  if (main.includes('Nhỏ')) return 'N'
  return '•'
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
    await joinRoom()
  },
  { immediate: true },
)

watch(
  () => selectedRoomCode.value,
  async (roomCode) => {
    if (!roomCode) {
      disconnectRoomStateStream()
      return
    }
    resetTransientRoomUiState()
    historyPage.value = 1
    minePage.value = 1
    await loadRoomState(roomCode)
    connectRoomStateStream(roomCode)
    await loadActiveHistory(1)
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

watch(
  () => currentPeriod.value?.period_no,
  () => {
    syncCountdownTarget(currentPeriod.value, clockTick.value)
  },
  { immediate: true },
)

watch(
  () => expectedPeriodSeconds.value,
  () => {
    syncCountdownTarget(currentPeriod.value, clockTick.value, true)
  },
)

watch(
  () => activeHistoryTab.value,
  async () => { await loadActiveHistory(currentPage()) },
)

onMounted(() => {
  void loadWallet()
  timer = window.setInterval(() => { clockTick.value = Date.now() }, 1000)
})

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
  disconnectRoomStateStream()
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
          <button class="flex items-center gap-1 rounded-full border border-[#f0c0c0] bg-[#fff5f5] px-2.5 py-1 text-[0.65rem] font-semibold text-primary">
            <span class="material-symbols-outlined text-[0.8rem]">menu_book</span>
            Cách chơi
          </button>
        </div>
        <span class="text-[0.72rem] text-slate-500">Thời gian còn lại</span>
      </div>
      <div class="flex items-center justify-between mt-2">
        <div>
          <p class="text-[0.78rem] font-bold text-on-surface">{{ currentPeriod?.period_no ?? '—' }}</p>
          <p class="text-[0.65rem] text-slate-400 uppercase tracking-wide mt-0.5">{{ roomStatusLabel }}</p>
        </div>
        <!-- Digit-box countdown (matching source design) -->
        <div class="flex items-center gap-1">
          <div class="flex h-8 w-7 items-center justify-center rounded-[6px] bg-[#1a1a1a] text-[1rem] font-black text-white">{{ countdownParts.m0 }}</div>
          <div class="flex h-8 w-7 items-center justify-center rounded-[6px] bg-[#1a1a1a] text-[1rem] font-black text-white">{{ countdownParts.m1 }}</div>
          <span class="text-[1.1rem] font-black text-[#1a1a1a] px-0.5">:</span>
          <div class="flex h-8 w-7 items-center justify-center rounded-[6px] bg-[#1a1a1a] text-[1rem] font-black text-white">{{ countdownParts.s0 }}</div>
          <div class="flex h-8 w-7 items-center justify-center rounded-[6px] bg-[#1a1a1a] text-[1rem] font-black text-white">{{ countdownParts.s1 }}</div>
        </div>
      </div>
    </div>

    <!-- ===== K3: DICE DISPLAY ===== -->
    <template v-if="isK3">
      <div class="mx-3 mt-2 rounded-[16px] overflow-hidden">
        <div class="flex justify-center gap-4 bg-[#1a5c34] border-[3px] border-[#2d8c4e] rounded-[16px] py-5 px-4">
          <div
            v-for="(d, i) in currentDice"
            :key="i"
            class="relative h-[62px] w-[62px] rounded-[14px]"
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
        <div v-for="result in recentResults" :key="result.period_no" class="flex gap-1">
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
            v-for="result in recentResults"
            :key="result.period_no"
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
    <div class="mx-3 mt-2 rounded-[16px] bg-white px-3 py-3 shadow-sm border border-slate-100">

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
        <div v-for="group in activeK3Groups" :key="group.title">
          <!-- Grid mode: circles with odds (Tổng số, 2 số trùng, 3 số trùng) -->
          <div v-if="group.mode === 'grid'" class="grid grid-cols-4 gap-2 mb-3">
              <button
                v-for="option in group.options"
                :key="option.key"
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
                v-for="option in group.options"
                :key="option.key"
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
            v-for="option in colorGroup.options"
            :key="option.key"
            type="button"
            class="min-h-[48px] rounded-[10px] text-[0.9rem] font-black text-white transition-transform active:scale-95 disabled:cursor-not-allowed disabled:opacity-50"
            :style="{ background: option.accent }"
            :disabled="!canBet"
            @click="selectOption(colorGroup.title, option.key, option.label)"
          >{{ option.label }}</button>
        </div>

        <!-- Number balls 0-9 -->
        <div v-if="numberGroup" class="grid grid-cols-5 gap-2 mb-3">
          <button
            v-for="option in numberGroup.options"
            :key="option.key"
            type="button"
            class="aspect-square rounded-full text-[1rem] font-black text-white transition-transform active:scale-95 hover:scale-105 disabled:cursor-not-allowed disabled:opacity-50"
            :style="{ background: option.accent }"
            :disabled="!canBet"
            @click="selectOption(numberGroup.title, option.key, option.label)"
          >{{ option.label }}</button>
        </div>

        <!-- Multiplier / Stake row -->
        <div class="flex gap-1.5 mb-3 overflow-x-auto no-scrollbar">
          <button
            v-for="multiplier in stakeOptions"
            :key="multiplier"
            type="button"
            class="flex-shrink-0 rounded-[8px] border-[1.5px] px-3 py-1.5 text-[0.78rem] font-black transition-all"
            :class="multiplier === selectedMultiplier ? 'border-primary bg-[#fff5f5] text-primary' : 'border-slate-200 bg-slate-50 text-slate-500'"
            @click="applyChipMultiplier(multiplier)"
          >X{{ multiplier }}</button>
        </div>

        <!-- Big / Small buttons -->
        <div v-if="bigSmallGroup" class="grid grid-cols-2 gap-2 mb-3">
          <button
            v-for="option in bigSmallGroup.options"
            :key="option.key"
            type="button"
            class="min-h-[52px] rounded-[10px] text-[1rem] font-black text-white transition-transform active:scale-95 disabled:cursor-not-allowed disabled:opacity-50"
            :style="{ background: option.accent }"
            :disabled="!canBet"
            @click="selectOption(bigSmallGroup.title, option.key, option.label)"
          >{{ option.label }}</button>
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
        Số dư hiện tại bằng 0. Vui lòng nạp tiền.
      </div>
      <div v-else-if="isBetLocked" class="rounded-[12px] bg-amber-50 px-4 py-3 text-[0.78rem] font-semibold text-amber-700">
        Kỳ hiện tại đã bước vào 5 giây cuối hoặc đã khóa lệnh. Vui lòng chờ kỳ tiếp theo.
      </div>
      <p v-if="joinError" class="mt-2 rounded-[12px] bg-red-50 px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">{{ joinError }}</p>
      <p v-if="visibleBetMessage" class="mt-2 rounded-[12px] bg-[rgba(255,109,102,0.08)] px-4 py-3 text-[0.78rem] font-semibold text-primary">{{ visibleBetMessage }}</p>
    </div>

    <!-- ===== HISTORY SECTION ===== -->
    <div class="mx-3 mt-2 rounded-[16px] bg-white shadow-sm border border-slate-100 overflow-hidden">
      <!-- History tabs -->
      <div class="flex bg-[#fff5f5] border-b border-[#f0e0e0]">
        <button
          type="button"
          class="flex-1 py-2.5 text-[0.72rem] font-semibold border-b-2 transition-all"
          :class="activeHistoryTab === 'history' ? 'border-[#e8404a] text-[#e8404a] bg-white' : 'border-transparent text-slate-500'"
          @click="activeHistoryTab = 'history'"
        >Lịch sử trò chơi</button>
        <button
          type="button"
          class="flex-1 py-2.5 text-[0.72rem] font-semibold border-b-2 transition-all"
          :class="activeHistoryTab === 'chart' ? 'border-[#e8404a] text-[#e8404a] bg-white' : 'border-transparent text-slate-500'"
          @click="activeHistoryTab = 'chart'"
        >Biểu đồ</button>
        <button
          type="button"
          class="flex-1 py-2.5 text-[0.72rem] font-semibold border-b-2 transition-all"
          :class="activeHistoryTab === 'mine' ? 'border-[#e8404a] text-[#e8404a] bg-white' : 'border-transparent text-slate-500'"
          @click="activeHistoryTab = 'mine'"
        >Lịch sử của tôi</button>
      </div>

      <!-- Chart -->
      <div v-if="activeHistoryTab === 'chart'" class="px-3 py-3">
        <div class="mb-3 flex items-center justify-between">
          <div>
            <p class="text-[0.68rem] uppercase tracking-[0.12em] text-slate-400">Biểu đồ kết quả</p>
            <strong class="text-[0.9rem] font-black text-on-surface">24 kỳ gần nhất</strong>
          </div>
          <button
            type="button"
            class="rounded-full bg-[#fff5f5] px-3 py-1.5 text-[0.7rem] font-black text-primary"
            @click="void loadChartHistory()"
          >
            Làm mới
          </button>
        </div>

        <div v-if="chartLoading" class="flex min-h-40 items-center justify-center text-[0.82rem] text-slate-400">
          Đang tải dữ liệu biểu đồ...
        </div>
        <div v-else-if="chartError" class="rounded-[14px] bg-red-50 px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">
          {{ chartError }}
        </div>
        <div v-else class="rounded-[18px] bg-[#fff9f9] p-3">
          <div class="flex items-end gap-2 overflow-x-auto pb-2 no-scrollbar">
            <div
              v-for="item in chartSeries"
              :key="item.periodNo"
              class="flex min-w-[52px] flex-col items-center gap-2"
            >
              <div class="flex h-28 w-full items-end">
                <div
                  class="w-full rounded-t-[12px] transition-all"
                  :class="item.barClass"
                  :style="{ height: `${Math.max(20, (item.value / chartMaxValue) * 100)}%` }"
                />
              </div>
              <span class="rounded-full px-2 py-0.5 text-[0.6rem] font-black text-white" :class="item.barClass">{{ item.label }}</span>
              <span class="text-[0.62rem] font-semibold text-slate-400">{{ item.periodNo ? item.periodNo.slice(-4) : '—' }}</span>
            </div>
          </div>
          <div class="mt-3 grid grid-cols-2 gap-2 text-[0.72rem]">
            <div class="rounded-[14px] bg-white px-3 py-2">
              <p class="text-slate-400">Mức cao nhất</p>
              <strong class="block text-on-surface">{{ chartMaxValue }}</strong>
            </div>
            <div class="rounded-[14px] bg-white px-3 py-2">
              <p class="text-slate-400">Kết quả gần nhất</p>
              <strong class="block text-on-surface">{{ chartSeries[0]?.label ?? '—' }}</strong>
            </div>
          </div>
        </div>
      </div>

      <template v-else>
        <div v-if="historyLoading || mineLoading" class="flex min-h-36 items-center justify-center text-[0.82rem] text-slate-400">
          Đang tải dữ liệu...
        </div>

        <div v-else-if="activeHistoryTab === 'history' && historyError" class="px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">{{ historyError }}</div>
        <div v-else-if="activeHistoryTab === 'mine' && mineError" class="px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">{{ mineError }}</div>

        <!-- History table -->
        <div v-else-if="activeHistoryTab === 'history'" class="overflow-hidden">
          <div class="grid grid-cols-[2fr_0.8fr_1.2fr_0.8fr] bg-[#f9f9f9] border-b border-[#f0e0e0] px-3 py-2 text-[0.62rem] font-black uppercase tracking-wide text-slate-400">
            <span>Kỳ xổ</span>
            <span class="text-center">Số</span>
            <span>Lớn nhỏ</span>
            <span class="text-right">Màu sắc</span>
          </div>
          <div
            v-for="row in historyRows"
            :key="row.period_no"
            class="grid grid-cols-[2fr_0.8fr_1.2fr_0.8fr] items-center border-b border-[#f8f0f0] px-3 py-2.5 text-[0.78rem] hover:bg-[#fff9f9]"
          >
            <span class="text-slate-400 text-[0.62rem]">…{{ row.period_no ? row.period_no.slice(-6) : '—' }}</span>
            <span
              class="flex h-7 w-7 mx-auto items-center justify-center rounded-full text-[0.75rem] font-black text-white"
              :class="resultBadgeClass(row.color)"
              :style="resultBadgeStyle(row.color)"
            >{{ row.result ? row.result.slice(0, 1) : '—' }}</span>
            <span class="font-semibold" :class="row.big_small?.includes('Lớn') ? 'text-[#e8404a]' : 'text-[#3b82f6]'">{{ row.big_small || '—' }}</span>
            <span class="flex justify-end">
              <span class="h-3.5 w-3.5 rounded-full" :class="resultDotClass(row.color)" />
            </span>
          </div>
          <div v-if="!historyRows.length" class="flex flex-col items-center gap-2 py-8 text-slate-300">
            <span class="material-symbols-outlined text-[2rem]">history</span>
            <p class="text-[0.82rem]">Không có dữ liệu</p>
          </div>
        </div>

        <!-- Mine history cards -->
        <div v-else class="divide-y divide-[#f8f0f0]">
          <button
            v-for="row in mineRows"
            :key="row.id"
            type="button"
            class="w-full px-3 py-3 text-left transition-colors hover:bg-[#fff9f9]"
            @click="openTicketDetail(row)"
          >
            <div class="flex items-start gap-3">
              <div class="mt-0.5 flex gap-1 flex-shrink-0">
                <div
                  v-if="isK3"
                  v-for="d in (row.result?.split('-').map(Number) ?? [1,1,1]).slice(0,3)"
                  :key="d"
                  class="flex h-7 w-7 items-center justify-center rounded-[6px] text-[0.7rem] font-black text-white"
                  :style="{ background: diceColor(d) }"
                >{{ d }}</div>
                <div
                  v-else
                  class="h-7 w-7 rounded-full flex items-center justify-center text-[0.75rem] font-black text-white"
                  :style="{ background: wingoTicketBallBackground(row) }"
                >{{ wingoTicketBallText(row) }}</div>
              </div>

              <div class="min-w-0 flex-1">
                <p class="truncate text-[0.72rem] font-semibold text-slate-400">{{ row.period_no || '—' }}</p>
                <p class="mt-0.5 text-[0.88rem] font-black text-on-surface">{{ rowMainLabel(row) }}</p>
                <p class="mt-0.5 text-[0.68rem] text-slate-500">{{ rowSubLabel(row) }}</p>
                <div class="mt-2 flex flex-wrap items-center gap-1.5">
                  <span class="rounded-full bg-[#fff5f5] px-2 py-0.5 text-[0.62rem] font-semibold text-primary">Gốc {{ formatMoney(rowOriginalAmountValue(row)) }}đ</span>
                  <span class="rounded-full bg-slate-100 px-2 py-0.5 text-[0.62rem] font-semibold text-slate-500">Thuế {{ formatMoney(rowTaxAmountValue(row)) }}đ</span>
                  <span class="rounded-full bg-[#f0fff6] px-2 py-0.5 text-[0.62rem] font-semibold text-[#10b981]">Nhận {{ formatMoney(rowWinCreditValue(row)) }}đ</span>
                </div>
              </div>

              <div class="text-right flex-shrink-0">
                <p class="text-[0.86rem] font-black text-on-surface">{{ formatMoney(rowOriginalAmountValue(row)) }}đ</p>
                <p class="mt-0.5 text-[0.7rem] font-semibold" :class="rowStatusClass(row)">
                  {{ rowStatusText(row) }}
                </p>
                <p class="mt-1 text-[0.62rem] text-slate-400 uppercase">{{ row.status }}</p>
              </div>
            </div>
          </button>
          <div v-if="!mineRows.length" class="flex flex-col items-center gap-2 py-8 text-slate-300">
            <span class="material-symbols-outlined text-[2rem]">history</span>
            <p class="text-[0.82rem]">Không có lịch sử cược</p>
          </div>
        </div>

        <!-- Pagination -->
        <div class="flex items-center justify-between px-3 py-3 border-t border-[#f0e0e0]">
          <button
            type="button"
            class="flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 text-slate-400 disabled:opacity-30 transition-all"
            :disabled="currentPage() <= 1"
            @click="setCurrentPage(Math.max(1, currentPage() - 1)); void loadActiveHistory(currentPage())"
          >
            <span class="material-symbols-outlined text-[1.1rem]">chevron_left</span>
          </button>
          <span class="text-[0.75rem] text-slate-500 font-semibold">{{ currentPage() }} / {{ currentTotalPages() }}</span>
          <button
            type="button"
            class="flex h-8 w-8 items-center justify-center rounded-full border border-[#e8404a] bg-[#e8404a] text-white disabled:opacity-30 transition-all"
            :disabled="currentPage() >= currentTotalPages()"
            @click="setCurrentPage(Math.min(currentTotalPages(), currentPage() + 1)); void loadActiveHistory(currentPage())"
          >
            <span class="material-symbols-outlined text-[1.1rem]">chevron_right</span>
          </button>
        </div>
      </template>
    </div>
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

        <div class="mb-3 flex items-center justify-between rounded-[14px] bg-slate-50 px-4 py-3">
          <span class="text-[0.82rem] text-slate-500">Tổng tiền cược</span>
          <div class="flex items-center gap-3">
            <button
              class="flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 text-[1.1rem] font-bold text-slate-500 transition-all hover:border-primary hover:text-primary"
              @click="modalBetAmount = Math.max(10000, modalBetAmount - 10000)"
            >−</button>
            <span class="min-w-[90px] text-center text-[0.95rem] font-black text-on-surface">{{ formatMoney(modalBetAmount) }}đ</span>
            <button
              class="flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 text-[1.1rem] font-bold text-slate-500 transition-all hover:border-primary hover:text-primary"
              @click="modalBetAmount += 10000"
            >+</button>
          </div>
        </div>

        <!-- Preset amounts -->
        <div class="mb-5 flex flex-wrap gap-2">
          <button
            v-for="preset in betPresets"
            :key="preset"
            class="rounded-full border-[1.5px] border-primary bg-[#fff5f5] px-3 py-1.5 text-[0.72rem] font-black text-primary transition-all hover:bg-primary hover:text-white"
            @click="modalBetAmount = preset"
          >{{ formatMoney(preset) }}</button>
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
            <strong class="block text-on-surface">{{ resultModalSettledAt ? new Date(resultModalSettledAt).toLocaleTimeString('vi-VN') : '—' }}</strong>
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
              <strong class="block text-on-surface">{{ selectedTicketDetail.settled_at ? new Date(selectedTicketDetail.settled_at).toLocaleTimeString('vi-VN') : '—' }}</strong>
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
      <p class="text-[0.9rem] leading-6 text-slate-500">
        Màn chơi này đã có trong danh sách nhưng chưa có room/time khả dụng để vào ngay. Hãy quay lại phòng chơi hoặc chọn game đang mở.
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
  </div>
</template>
