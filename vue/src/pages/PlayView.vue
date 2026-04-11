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

const stakeOptions = [1, 5, 10, 20, 50] as const
const tablePageSize = 4
let timer: number | undefined

const room = computed<PlayRoom | null>(() => getPlayRoom(String(route.params.game ?? 'wingo')) ?? null)
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
    minutes: String(minutes).padStart(2, '0'),
    seconds: String(seconds).padStart(2, '0'),
  }
})

const roomStatusLabel = computed(() => {
  const status = (currentPeriod.value?.status || '').toUpperCase()
  if (!currentPeriod.value) return room.value?.status === 'COMING_SOON' ? 'Sắp mở' : 'Chưa cập nhật'
  if (status === 'OPEN') {
    return isBetLocked.value ? 'Đang khóa' : 'Đang mở'
  }
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

const colorGroup = computed(() => selectedVariant.value?.betGroups.find((group) => groupTypeKey(group) === 'COLOR') ?? null)
const numberGroup = computed(() => selectedVariant.value?.betGroups.find((group) => groupTypeKey(group) === 'NUMBER') ?? null)
const bigSmallGroup = computed(() => selectedVariant.value?.betGroups.find((group) => groupTypeKey(group) === 'BIG_SMALL') ?? null)

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
      {
      },
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
      {
        token: auth.accessToken,
      },
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
  if (!wallet.wallets.length) {
    await loadWallet()
  }
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

function selectOption(groupTitle: string, key: string) {
  selectedOptions[groupTitle] = key
}

function selectRandomStake() {
  const index = Math.floor(Math.random() * stakeOptions.length)
  selectedMultiplier.value = stakeOptions[index] ?? 10
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

function groupButtonClass(group: PlayBetGroup, optionKey: string) {
  const selected = selectedOptions[group.title] === optionKey
  const title = group.title.toLowerCase()

  if (title.includes('màu')) {
    return selected ? 'ring-2 ring-offset-2 ring-primary shadow-[0_12px_24px_rgba(255,109,102,0.12)]' : 'shadow-[0_10px_18px_rgba(17,24,39,0.08)]'
  }
  if (title.includes('lớn') || title.includes('nhỏ')) {
    return selected ? 'ring-2 ring-offset-2 ring-primary shadow-[0_12px_24px_rgba(255,109,102,0.12)]' : 'shadow-[0_10px_18px_rgba(17,24,39,0.08)]'
  }
  return selected ? 'border-primary bg-primary/10 text-primary' : 'border-slate-200 bg-white text-on-surface'
}

async function submitBet() {
  if (!room.value || !selectedVariant.value || !auth.accessToken || !connectionId.value) return
  if (remainingSeconds.value === 0 || roomStatusLabel.value === 'Đang khóa' || room.value.status !== 'OPEN') {
    betMessage.value = 'Kỳ hiện tại đã khóa, không thể đặt lệnh.'
    return
  }

  const amount = Number.parseFloat(stakeAmount.value || '0')
  if (!Number.isFinite(amount) || amount <= 0) {
    betMessage.value = 'Vui lòng nhập số tiền hợp lệ.'
    return
  }

  const items = selectedVariant.value.betGroups.flatMap((group) => {
    const selected = selectedOptions[group.title]
    if (!selected) return []
    return [{
      option_type: groupTypeKey(group),
      option_key: selected,
      stake: String(amount),
    }]
  })

  if (!items.length) {
    betMessage.value = 'Vui lòng chọn ít nhất một cửa cược.'
    return
  }

  if (availableVndBalance.value < Number(stakeAmount.value)) {
    betMessage.value = 'Số dư không đủ để đặt lệnh.'
    return
  }
  if (!currentPeriod.value) {
    betMessage.value = 'Chưa có kỳ hiện tại để đặt lệnh.'
    return
  }
  if (isBetLocked.value) {
    betMessage.value = 'Kỳ hiện tại đã khóa, không thể đặt lệnh.'
    return
  }

  betLoading.value = true
  betMessage.value = ''

  try {
    const requestId = globalThis.crypto?.randomUUID?.() ?? `req-${Date.now()}`
    const res = await request<PlayRoomBetResponse>('POST', `/v1/play/rooms/${selectedRoomCode.value}/bets`, {
      token: auth.accessToken,
      headers: {
        'X-Connection-ID': connectionId.value,
      },
      body: {
        request_id: requestId,
        period_id: String(currentPeriod.value.id),
        items,
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
  () => {
    ensureDefaultSelections(selectedVariant.value)
  },
  { immediate: true },
)

watch(
  () => activeHistoryTab.value,
  async () => {
    await loadActiveHistory(currentPage())
  },
)

onMounted(() => {
  void loadWallet()
  timer = window.setInterval(() => {
    clockTick.value = Date.now()
  }, 1000)
})

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div v-if="room && selectedVariant" class="space-y-3 pb-24 pt-1 md:space-y-4 md:pb-28">
    <header class="flex items-center justify-between gap-3 bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] -mx-3 -mt-3 p-4 pb-8 md:-mx-4 md:-mt-4 text-white rounded-b-[30px] shadow-[0_8px_20px_rgba(229,46,46,0.15)]">
      <button class="grid h-10 w-10 place-items-center rounded-full text-white transition-transform active:scale-95" type="button" @click="router.push('/play')">
        <span class="material-symbols-outlined text-[1.9rem]">arrow_back</span>
      </button>

      <div class="min-w-0 flex-1 text-center">
        <h1 class="truncate text-[1.35rem] font-black tracking-[-0.04em] text-white md:text-[1.5rem]">
          {{ room.title }} {{ selectedVariant.durationLabel }}
        </h1>
      </div>

      <button class="text-[1rem] font-bold text-white transition-opacity active:opacity-80" type="button">
        Hỗ trợ
      </button>
    </header>

    <section class="relative overflow-hidden rounded-[24px] bg-white p-4 text-on-surface shadow-[0_14px_28px_rgba(229,46,46,0.08)] md:p-5 mx-2 -mt-6 z-10">
      <div class="absolute right-5 top-5 grid h-14 w-14 place-items-center rounded-[18px] bg-primary/10 text-primary">
        <span class="material-symbols-outlined text-[2rem]">account_balance_wallet</span>
      </div>
      <div class="relative z-10 max-w-[78%]">
        <p class="text-[0.74rem] font-bold uppercase tracking-[0.2em] text-on-surface-variant">
          Số dư hiện tại
        </p>
        <div class="mt-1 flex items-baseline gap-2">
          <strong class="text-[2.35rem] font-black leading-none tracking-[-0.05em] md:text-[2.7rem] text-primary">
            {{ currentBalanceLabel }}đ
          </strong>
          <span class="material-symbols-outlined text-[1.2rem] text-primary/80">refresh</span>
        </div>
      </div>

      <div class="mt-5 grid grid-cols-2 gap-3">
        <RouterLink to="/deposit" class="flex min-h-14 items-center justify-center gap-2 rounded-full bg-gradient-to-r from-[#24b561] to-[#2fcc71] shadow-[0_8px_16px_rgba(36,181,97,0.25)] px-4 font-black text-white transition-transform active:scale-95">
          <span class="material-symbols-outlined text-[1.2rem]">add_circle</span>
          Nạp tiền
        </RouterLink>
        <RouterLink to="/account" class="flex min-h-14 items-center justify-center gap-2 rounded-full bg-gradient-to-r from-[#ff8a00] to-[#f6c32d] shadow-[0_8px_16px_rgba(255,138,0,0.25)] px-4 font-black text-white transition-transform active:scale-95">
          <span class="material-symbols-outlined text-[1.2rem]">account_balance</span>
          Rút tiền
        </RouterLink>
      </div>
    </section>

    <div class="rounded-full bg-white px-4 py-3 text-[0.8rem] font-semibold text-on-surface shadow-[0_8px_18px_rgba(255,109,102,0.05)]">
      <div class="flex items-center gap-2 overflow-hidden">
        <span class="material-symbols-outlined text-primary">campaign</span>
        <span class="whitespace-nowrap text-on-surface-variant">
          Chúc mừng người chơi ***123 vừa thắng 2,500,000đ
        </span>
      </div>
    </div>

    <p v-if="roomStateError" class="rounded-[14px] bg-[rgba(183,18,17,0.08)] px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">
      {{ roomStateError }}
    </p>

    <section class="grid grid-cols-4 gap-2 md:gap-3">
      <button
        v-for="variant in room.variants"
        :key="variant.code"
        type="button"
        class="grid min-h-[94px] place-items-center rounded-[20px] transition-all active:scale-[0.98] px-2 py-3 text-center"
        :class="variant.code === selectedVariant.code ? 'bg-gradient-to-tr from-[#ff8a00] to-[#e52e2e] text-white shadow-[0_12px_24px_rgba(229,46,46,0.3)]' : 'bg-white border border-slate-100 text-slate-500 shadow-sm'"
        @click="activeVariantCode = variant.code"
      >
        <span class="material-symbols-outlined text-[1.9rem]" :class="variant.code === selectedVariant.code ? 'text-white/90' : 'text-slate-400'">
          schedule
        </span>
        <span class="mt-1 text-[0.76rem] font-black uppercase tracking-[0.04em]">
          {{ variant.durationLabel }}
        </span>
      </button>
    </section>

    <section class="grid gap-3 md:grid-cols-2">
      <article class="rounded-[20px] bg-white p-4 shadow-[0_8px_18px_rgba(255,109,102,0.05)]">
        <p class="text-[0.92rem] font-medium text-on-surface-variant">
          Số kỳ: <strong class="text-primary">{{ currentPeriod?.period_no ?? '—' }}</strong>
        </p>
        <div class="mt-3 flex items-center gap-1.5 justify-center h-14 bg-transparent text-[2.2rem] font-black text-primary">
          <span>{{ countdownParts.minutes[0] }}</span>
          <span>{{ countdownParts.minutes[1] }}</span>
          <span class="mx-1 pb-1 text-[#e64545]">:</span>
          <span class="text-[#e64545]">{{ countdownParts.seconds[0] }}</span>
          <span class="text-[#e64545]">{{ countdownParts.seconds[1] }}</span>
        </div>
        <p class="mt-3 text-[0.72rem] font-semibold uppercase tracking-[0.12em] text-on-surface-variant">
          {{ roomStatusLabel }}
        </p>
      </article>

      <article class="rounded-[20px] bg-white p-4 shadow-[0_8px_18px_rgba(255,109,102,0.05)]">
        <p class="text-[0.92rem] font-medium text-on-surface-variant">
          Kết quả gần đây
        </p>
        <div class="mt-3 flex flex-wrap gap-2.5">
          <span
            v-for="result in recentResults"
            :key="result.period_no"
            class="grid h-9 w-9 place-items-center rounded-full border text-[0.82rem] font-black text-white"
            :class="resultDotClass(result.color)"
          >
            {{ result.result }}
          </span>
        </div>
      </article>
    </section>

    <section class="rounded-[22px] bg-white p-4 shadow-[0_8px_18px_rgba(255,109,102,0.05)] md:p-5">
      <div class="grid gap-3">
        <div v-if="colorGroup" class="grid grid-cols-3 gap-3">
          <button
            v-for="option in colorGroup.options"
            :key="option.key"
            type="button"
            class="min-h-14 rounded-[16px] px-3 py-3 text-[1rem] font-black text-white transition-transform active:scale-[0.98]"
            :style="{ backgroundColor: option.accent }"
            :class="groupButtonClass(colorGroup, option.key)"
            @click="selectOption(colorGroup.title, option.key)"
          >
            {{ option.label }}
          </button>
        </div>

        <div v-if="numberGroup" class="grid grid-cols-5 gap-3">
          <button
            v-for="option in numberGroup.options"
            :key="option.key"
            type="button"
            class="grid h-14 place-items-center rounded-full border-2 bg-white text-[1.2rem] font-black transition-transform active:scale-[0.98]"
            :class="groupButtonClass(numberGroup, option.key)"
            @click="selectOption(numberGroup.title, option.key)"
          >
            {{ option.label }}
          </button>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <button
            type="button"
            class="rounded-[14px] bg-[#fff0ee] px-4 py-3 text-[0.88rem] font-black text-[#7a433e] transition-transform active:scale-[0.98]"
            @click="selectRandomStake"
          >
            Ngẫu nhiên
          </button>
          <button
            v-for="multiplier in stakeOptions"
            :key="multiplier"
            type="button"
            class="rounded-[12px] px-3 py-2 text-[0.86rem] font-black transition-transform active:scale-[0.98]"
            :class="multiplier === selectedMultiplier ? 'bg-primary text-white shadow-[0_10px_18px_rgba(255,109,102,0.18)]' : 'bg-[#fff4f2] text-[#2c2f33]'"
            @click="selectedMultiplier = multiplier"
          >
            x{{ multiplier }}
          </button>
        </div>

        <div v-if="bigSmallGroup" class="grid grid-cols-2 gap-3">
          <button
            v-for="option in bigSmallGroup.options"
            :key="option.key"
            type="button"
            class="min-h-16 rounded-[16px] px-4 text-[1.15rem] font-black transition-transform active:scale-[0.98]"
            :style="{ backgroundColor: option.accent }"
            :class="groupButtonClass(bigSmallGroup, option.key)"
            @click="selectOption(bigSmallGroup.title, option.key)"
          >
            {{ option.label }}
          </button>
        </div>

        <div class="text-center text-[0.76rem] font-semibold text-on-surface-variant">
          Mức đặt hiện tại: <span class="text-on-surface">{{ stakeLabel }}đ</span>
          <span class="mx-2 text-slate-300">•</span>
          Ví khóa: <span class="text-on-surface">{{ lockedBalanceLabel }}đ</span>
          <span class="mx-2 text-slate-300">•</span>
          USDT: <span class="text-on-surface">{{ formatMoney(walletUsdt?.balance ?? 0, 2) }}</span>
        </div>

        <div v-if="!canPlay" class="rounded-[14px] bg-[rgba(183,18,17,0.08)] px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">
          Số dư hiện tại bằng 0. Vui lòng nạp tiền trước khi vào phòng chơi.
        </div>

        <div v-else-if="isBetLocked" class="rounded-[14px] bg-[rgba(255,170,0,0.12)] px-4 py-3 text-[0.78rem] font-semibold text-amber-700">
          Kỳ hiện tại đã bước vào 5 giây cuối hoặc đã khóa lệnh. Vui lòng chờ kỳ tiếp theo.
        </div>

        <button
          type="button"
          class="min-h-14 rounded-[18px] bg-[linear-gradient(90deg,#fdd404_0%,#ffd400_100%)] text-[1.2rem] font-black text-[#5a4600] shadow-[0_12px_24px_rgba(253,212,4,0.16)] transition-transform active:scale-[0.98]"
          :disabled="betLoading || !canBet"
          @click="submitBet"
        >
          {{ betLoading ? 'Đang gửi...' : 'ĐẶT LỆNH' }}
        </button>

        <p v-if="joinError" class="rounded-[14px] bg-[rgba(183,18,17,0.08)] px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">
          {{ joinError }}
        </p>
        <p v-if="betMessage" class="rounded-[14px] bg-[rgba(255,109,102,0.08)] px-4 py-3 text-[0.78rem] font-semibold text-primary">
          {{ betMessage }}
        </p>
      </div>
    </section>

    <section class="rounded-[22px] bg-white shadow-[0_8px_18px_rgba(255,109,102,0.05)] border border-slate-100/50 mt-4 overflow-hidden">
      <div class="flex items-center gap-2 p-3 pb-2 text-[0.76rem] font-black w-full overflow-x-auto no-scrollbar">
        <button
          type="button"
          class="px-4 py-2.5 transition-all rounded-full whitespace-nowrap"
          :class="activeHistoryTab === 'history' ? 'bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] text-white shadow-[0_6px_14px_rgba(229,46,46,0.25)]' : 'bg-slate-100 text-slate-500'"
          @click="activeHistoryTab = 'history'"
        >
          Lịch sử trò chơi
        </button>
        <button
          type="button"
          class="px-4 py-2.5 transition-all rounded-full whitespace-nowrap"
          :class="activeHistoryTab === 'chart' ? 'bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] text-white shadow-[0_6px_14px_rgba(229,46,46,0.25)]' : 'bg-slate-100 text-slate-500'"
          @click="activeHistoryTab = 'chart'"
        >
          Biểu đồ
        </button>
        <button
          type="button"
          class="px-4 py-2.5 transition-all rounded-full whitespace-nowrap"
          :class="activeHistoryTab === 'mine' ? 'bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] text-white shadow-[0_6px_14px_rgba(229,46,46,0.25)]' : 'bg-slate-100 text-slate-500'"
          @click="activeHistoryTab = 'mine'"
        >
          Lịch sử của tôi
        </button>
      </div>

      <div class="p-3">
        <template v-if="activeHistoryTab === 'chart'">
          <div class="grid min-h-52 place-items-center rounded-[18px] bg-slate-50 text-center text-[0.82rem] text-slate-400">
            Biểu đồ kết quả sẽ hiển thị tại đây.
          </div>
        </template>

        <template v-else>
          <div v-if="historyLoading || mineLoading" class="grid min-h-48 place-items-center rounded-[18px] bg-background text-[0.82rem] text-on-surface-variant">
            Đang tải dữ liệu...
          </div>

          <div v-else-if="activeHistoryTab === 'history' && historyError" class="rounded-[18px] bg-[rgba(183,18,17,0.08)] px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">
            {{ historyError }}
          </div>

          <div v-else-if="activeHistoryTab === 'mine' && mineError" class="rounded-[18px] bg-[rgba(183,18,17,0.08)] px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">
            {{ mineError }}
          </div>

          <div class="overflow-hidden rounded-[18px] border border-slate-200/70 mt-1 shadow-sm">
          <div class="grid grid-cols-[1fr_auto_1fr_auto] gap-2 bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] px-4 py-3.5 text-[0.7rem] font-black uppercase tracking-[0.06em] text-white">
              <span>Kỳ xổ</span>
              <span>Số</span>
              <span>Lớn nhỏ</span>
              <span>Màu sắc</span>
            </div>

            <div
              v-for="row in periodHistory"
              :key="row.period_no"
              class="grid grid-cols-[1fr_auto_1fr_auto] items-center gap-2 border-t border-slate-200/70 px-4 py-4 text-[0.86rem]"
            >
              <span class="font-medium text-on-surface">…{{ row.period_no.slice(-3) }}</span>
              <span class="grid h-8 w-8 place-items-center rounded-full border text-[0.78rem] font-black shadow-sm" :class="resultBadgeClass(row.result)">
                {{ row.result.slice(0, 1) }}
              </span>
              <span class="font-semibold" :class="row.status === 'WON' || row.status === 'DRAWN' ? 'text-primary' : 'text-amber-700'">
                {{ row.big_small || '—' }}
              </span>
              <span class="grid place-items-center">
                <span class="h-3.5 w-3.5 rounded-full" :class="resultDotClass(row.color)" />
              </span>
            </div>
          </div>

          <div class="mt-4 flex items-center justify-between gap-3">
            <button
              type="button"
              class="rounded-full bg-surface-container-low px-4 py-2 text-[0.76rem] font-black text-on-surface-variant disabled:opacity-40"
              :disabled="currentPage() <= 1"
              @click="setCurrentPage(Math.max(1, currentPage() - 1)); void loadActiveHistory(currentPage())"
            >
              Trang trước
            </button>

            <span class="text-[0.74rem] font-bold text-on-surface-variant">
              Trang {{ currentPage() }} / {{ currentTotalPages() }}
            </span>

            <button
              type="button"
              class="rounded-full bg-primary px-4 py-2 text-[0.76rem] font-black text-white disabled:opacity-40"
              :disabled="currentPage() >= currentTotalPages()"
              @click="setCurrentPage(Math.min(currentTotalPages(), currentPage() + 1)); void loadActiveHistory(currentPage())"
            >
              Trang sau
            </button>
          </div>
        </template>
      </div>
    </section>
  </div>
</template>
