<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

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
const betLoading = ref(false)
const selectedMultiplier = ref(10)
const selectedOptions = reactive<Record<string, string>>({})
const roomState = ref<PlayRoomStateResponse | null>(null)
const historyRows = ref<PlayRoomHistoryResponse['items']>([])
const mineRows = ref<PlayRoomBetHistoryResponse['items']>([])
const historyPage = ref(1)
const historyTotalPages = ref(1)
const minePage = ref(1)
const mineTotalPages = ref(1)
const historyLoading = ref(false)
const mineLoading = ref(false)
const historyError = ref('')
const mineError = ref('')
const roomStateLoading = ref(false)
const roomStateError = ref('')
const serverTimeOffsetMs = ref(0)
const clockTick = ref(Date.now())

// Bet modal state
const showBetModal = ref(false)
const modalBetLabel = ref('')
const modalBetKey = ref('')
const modalBetGroupTitle = ref('')
const modalBetAmount = ref(10000)
const betPresets = [10000, 50000, 100000, 500000, 1000000]

const stakeOptions = [1, 5, 10, 20, 50] as const
const tablePageSize = 4
let timer: number | undefined

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
const currentPeriodDrawAtMs = computed(() => currentPeriod.value ? new Date(currentPeriod.value.draw_at).getTime() : 0)
const currentPeriodBetLockAtMs = computed(() => currentPeriod.value ? new Date(currentPeriod.value.bet_lock_at).getTime() : 0)
const isBetLocked = computed(() => {
  if (!currentPeriod.value) return true
  if ((currentPeriod.value.status || '').toUpperCase() !== 'OPEN') return true
  return syncedNow.value >= currentPeriodBetLockAtMs.value
})
const canBet = computed(() => canPlay.value && !isBetLocked.value)

const remainingSeconds = computed(() => {
  if (!currentPeriod.value) return 0
  return Math.max(0, Math.ceil((currentPeriodDrawAtMs.value - syncedNow.value) / 1000))
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

const recentResults = computed(() => roomState.value?.recent_results.slice(0, 5) ?? [])
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
function wingoBallColor(n: number): string {
  if (n === 0 || n === 5) return 'linear-gradient(135deg, #8b5cf6, #e8404a)'
  if (n >= 1 && n <= 4) return '#e8404a'
  return '#10b981'
}

function formatMoney(value: string | number | null | undefined, fractionDigits = 0) {
  return formatViMoney(value ?? 0, fractionDigits)
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
}

async function loadWallet() {
  if (!auth.isAuthenticated) return
  try {
    await wallet.fetchSummary()
  } catch {
    // state already stores the error
  }
}

async function loadRoomState(roomCode = selectedRoomCode.value) {
  if (!roomCode) return
  roomStateLoading.value = true
  roomStateError.value = ''
  try {
    const response = await request<PlayRoomStateResponse>('GET', `/v1/play/rooms/${roomCode}/state`)
    roomState.value = response
    serverTimeOffsetMs.value = new Date(response.server_time).getTime() - Date.now()
    clockTick.value = Date.now()
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

async function loadActiveHistory(page = currentPage()) {
  if (activeHistoryTab.value === 'mine') {
    await loadMineHistory(page)
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
  modalBetGroupTitle.value = groupTitle
  modalBetKey.value = optionKey
  modalBetLabel.value = optionLabel
  modalBetAmount.value = 10000
  showBetModal.value = true
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
    showBetModal.value = false
    return
  }
  if (isBetLocked.value) {
    betMessage.value = 'Kỳ hiện tại đã khóa, không thể đặt lệnh.'
    showBetModal.value = false
    return
  }
  if (!currentPeriod.value) {
    betMessage.value = 'Chưa có kỳ hiện tại.'
    showBetModal.value = false
    return
  }

  betLoading.value = true
  betMessage.value = ''
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
    await wallet.fetchSummary()
    await loadActiveHistory()
  } catch (error: unknown) {
    const err = error as ApiError
    betMessage.value = err?.message ?? 'Không thể gửi lệnh cược'
  } finally {
    betLoading.value = false
  }
}

function resultDotClass(label: string) {
  if (label.toLowerCase().includes('xanh')) return 'bg-[#24b561]'
  if (label.toLowerCase().includes('đỏ')) return 'bg-[#e64545]'
  if (label.toLowerCase().includes('tím')) return 'bg-[#8b5cf6]'
  return 'bg-primary'
}

function resultBadgeClass(label: string) {
  if (label.toLowerCase().includes('xanh')) return 'border-[#24b561] bg-[#24b561] text-white'
  if (label.toLowerCase().includes('đỏ')) return 'border-[#e64545] bg-[#e64545] text-white'
  if (label.toLowerCase().includes('tím')) return 'border-[#8b5cf6] bg-[#8b5cf6] text-white'
  return 'border-primary bg-primary text-white'
}

watch(
  () => room.value?.code,
  async () => {
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
    if (!roomCode) return
    historyPage.value = 1
    minePage.value = 1
    await loadRoomState(roomCode)
    await loadActiveHistory(1)
  },
  { immediate: true },
)

watch(
  () => selectedVariant.value?.code,
  () => { ensureDefaultSelections(selectedVariant.value) },
  { immediate: true },
)

watch(
  () => activeHistoryTab.value,
  async () => { await loadActiveHistory(currentPage()) },
)

onMounted(() => {
  void loadWallet()
  timer = window.setInterval(() => { clockTick.value = Date.now() }, 1000)
})

onBeforeUnmount(() => { if (timer) window.clearInterval(timer) })
</script>

<template>
  <div v-if="room && selectedVariant" class="min-h-dvh pb-28" style="background: #f7f0f0;">
    <!-- ===== HEADER GRADIENT ===== -->
    <header class="flex items-center justify-between bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] px-4 py-3 text-white shadow-lg">
      <button class="grid h-10 w-10 place-items-center rounded-full bg-white/15 text-white transition-transform active:scale-95" type="button" @click="router.push('/play')">
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
            :style="{ background: wingoBallColor(Number(result.result)) }"
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
            <span class="text-[3rem] font-black leading-none" :style="{ color: wingoBallColor(n) !== 'linear-gradient(135deg, #8b5cf6, #e8404a)' ? wingoBallColor(n) : '#8b5cf6' }">
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
              class="flex flex-col items-center justify-center aspect-square rounded-full text-white transition-transform active:scale-95 hover:opacity-90 gap-0.5 p-1"
              :style="{ background: option.accent }"
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
                class="flex flex-col items-center justify-center rounded-[10px] py-3 text-white font-black text-[0.82rem] transition-all active:scale-95"
                :style="{ background: option.accent }"
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
                class="rounded-[10px] px-4 py-2.5 text-white text-[0.82rem] font-bold transition-all active:scale-95 flex-1"
                :style="{ background: option.accent }"
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
            class="min-h-[48px] rounded-[10px] text-[0.9rem] font-black text-white transition-transform active:scale-95"
            :style="{ background: option.accent }"
            @click="selectOption(colorGroup.title, option.key, option.label)"
          >{{ option.label }}</button>
        </div>

        <!-- Number balls 0-9 -->
        <div v-if="numberGroup" class="grid grid-cols-5 gap-2 mb-3">
          <button
            v-for="option in numberGroup.options"
            :key="option.key"
            type="button"
            class="aspect-square rounded-full text-[1rem] font-black text-white transition-transform active:scale-95 hover:scale-105"
            :style="{ background: option.accent }"
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
            @click="selectedMultiplier = multiplier"
          >X{{ multiplier }}</button>
        </div>

        <!-- Big / Small buttons -->
        <div v-if="bigSmallGroup" class="grid grid-cols-2 gap-2 mb-3">
          <button
            v-for="option in bigSmallGroup.options"
            :key="option.key"
            type="button"
            class="min-h-[52px] rounded-[10px] text-[1rem] font-black text-white transition-transform active:scale-95"
            :style="{ background: option.accent }"
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
              class="rounded-full px-4 py-2 text-[0.82rem] font-black text-white transition-transform active:scale-95"
              :style="{ background: option.accent }"
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
      <p v-if="betMessage" class="mt-2 rounded-[12px] bg-[rgba(255,109,102,0.08)] px-4 py-3 text-[0.78rem] font-semibold text-primary">{{ betMessage }}</p>
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

      <!-- Chart placeholder -->
      <div v-if="activeHistoryTab === 'chart'" class="flex min-h-40 flex-col items-center justify-center gap-2 py-8 text-slate-300">
        <span class="material-symbols-outlined text-[2.5rem]">bar_chart</span>
        <p class="text-[0.82rem]">Biểu đồ kết quả sẽ hiển thị tại đây.</p>
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
            <span class="text-slate-400 text-[0.62rem]">…{{ row.period_no.slice(-6) }}</span>
            <span
              class="flex h-7 w-7 mx-auto items-center justify-center rounded-full text-[0.75rem] font-black text-white"
              :class="resultBadgeClass(row.color)"
            >{{ row.result.slice(0, 1) }}</span>
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
          <div
            v-for="row in mineRows"
            :key="row.id"
            class="flex items-center gap-3 px-3 py-3"
          >
            <!-- Dice icons / color dots -->
            <div class="flex gap-1 flex-shrink-0">
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
                :style="{ background: wingoBallColor(Number(row.result ?? 0)) }"
              >{{ row.result ?? '?' }}</div>
            </div>
            <!-- Period + result label -->
            <div class="flex-1 min-w-0">
              <p class="text-[0.68rem] text-slate-400 truncate">{{ row.period_no }}</p>
              <p class="text-[0.78rem] font-semibold text-on-surface">{{ row.big_small || row.color || '—' }}</p>
            </div>
            <!-- Amount + status -->
            <div class="text-right flex-shrink-0">
              <p class="text-[0.82rem] font-black text-on-surface">{{ formatMoney(row.stake) }}đ</p>
              <p
                class="text-[0.68rem] font-semibold"
                :class="row.status === 'WON' ? 'text-[#10b981]' : row.status === 'LOST' ? 'text-slate-400' : 'text-amber-500'"
              >
                {{ row.status === 'WON' ? `Thắng ${formatMoney((Number(row.stake) * 1.96).toFixed(0))}đ` : row.status === 'LOST' ? 'Thua' : 'Chờ kết quả' }}
              </p>
            </div>
          </div>
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
</template>
