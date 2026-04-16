<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import type { PlayRoomBetHistoryResponse, PlayRoomHistoryResponse } from '@/shared/api/types'
import { formatViMoney } from '@/shared/lib/money'

type HistoryTab = 'history' | 'chart' | 'mine'

type ChartSeriesItem = {
  periodNo: string
  index: number
  value: number
  label: string
  barClass: string
}

const props = defineProps<{
  activeHistoryTab: HistoryTab
  chartSeries: ChartSeriesItem[]
  chartMaxValue: number
  historyRows: PlayRoomHistoryResponse['items']
  mineRows: PlayRoomBetHistoryResponse['items']
  historyLoading: boolean
  mineLoading: boolean
  chartLoading: boolean
  historyError: string
  mineError: string
  chartError: string
  historyPage: number
  historyTotalPages: number
  minePage: number
  mineTotalPages: number
  isK3: boolean
}>()

const emit = defineEmits<{
  (event: 'change-tab', tab: HistoryTab): void
  (event: 'refresh-chart'): void
  (event: 'refresh-history'): void
  (event: 'page-change', page: number): void
  (event: 'open-ticket-detail', row: PlayRoomBetHistoryResponse['items'][number]): void
}>()

const visibleHistoryRows = ref<PlayRoomHistoryResponse['items']>([])
const visibleMineRows = ref<PlayRoomBetHistoryResponse['items']>([])
const visibleChartSeries = ref<ChartSeriesItem[]>([])

watch(
  () => [props.historyRows, props.historyLoading] as const,
  ([rows, loading]) => {
    if (!loading) visibleHistoryRows.value = [...rows]
  },
  { immediate: true, deep: true },
)

watch(
  () => [props.mineRows, props.mineLoading] as const,
  ([rows, loading]) => {
    if (!loading) visibleMineRows.value = [...rows]
  },
  { immediate: true, deep: true },
)

watch(
  () => [props.chartSeries, props.chartLoading] as const,
  ([rows, loading]) => {
    if (!loading) visibleChartSeries.value = [...rows]
  },
  { immediate: true, deep: true },
)

const activePage = computed(() => {
  if (props.activeHistoryTab === 'mine') return props.minePage
  if (props.activeHistoryTab === 'chart') return 1
  return props.historyPage
})

const activeTotalPages = computed(() => {
  if (props.activeHistoryTab === 'mine') return props.mineTotalPages
  if (props.activeHistoryTab === 'chart') return 1
  return props.historyTotalPages
})

function setPage(page: number) {
  emit('page-change', page)
}

function formatMoney(value: string | number | null | undefined, fractionDigits = 0) {
  return formatViMoney(value ?? 0, fractionDigits)
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
  if (status === 'LOST') return `Thua ${formatMoney(Math.abs(rowProfitLossValue(row)))}đ`
  if (status === 'PENDING') return 'Đang chờ chốt kỳ'
  return status || 'Đang xử lý'
}

function rowStatusClass(row: PlayRoomBetHistoryResponse['items'][number]) {
  const status = rowStatusValue(row)
  if (status === 'WON') return 'text-[#10b981]'
  if (status === 'LOST') return 'text-slate-400'
  return 'text-amber-500'
}


function resultBadgeClass(label: string) {
  const lower = label.toLowerCase()
  if (lower.includes('green_violet') || lower.includes('red_violet')) {
    return 'border-transparent text-white'
  }
  if (lower.includes('xanh') || lower.includes('green')) return 'border-[#24b561] bg-[#24b561] text-white'
  if (lower.includes('đỏ') || lower.includes('red')) return 'border-[#e64545] bg-[#e64545] text-white'
  if (lower.includes('tím') || lower.includes('violet')) return 'border-[#8b5cf6] bg-[#8b5cf6] text-white'
  return 'border-primary bg-primary text-white'
}

function resultBadgeStyle(label: string) {
  const lower = label.toLowerCase()
  if (lower.includes('red_violet')) {
    return { background: 'linear-gradient(135deg, #e64545, #8b5cf6)' }
  }
  if (lower.includes('green_violet')) {
    return { background: 'linear-gradient(135deg, #24b561, #8b5cf6)' }
  }
  return {}
}

function periodTail(value: string | null | undefined, size = 6) {
  const raw = String(value ?? '').trim()
  if (!raw) return '—'
  if (raw.length <= size) return raw
  return `…${raw.slice(-size)}`
}

function parseTimeMs(value: string | null | undefined) {
  const raw = String(value ?? '').trim()
  if (!raw) return 0
  const normalized = raw.includes(' ') && !raw.includes('T') ? raw.replace(' ', 'T') : raw
  const parsed = new Date(normalized).getTime()
  return Number.isFinite(parsed) ? parsed : 0
}

function formatClockMs(ms: number) {
  if (!Number.isFinite(ms) || ms <= 0) return '—'
  return new Intl.DateTimeFormat('vi-VN', {
    timeZone: 'Asia/Ho_Chi_Minh',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(ms))
}

function formatDrawClock(drawAt: string | null | undefined, createdAt?: string | null | undefined) {
  const drawMs = parseTimeMs(drawAt)
  const createdMs = parseTimeMs(createdAt)

  // Backend dữ liệu lịch sử hiện có room trả draw_at lệch đúng ~7h so với created_at.
  // Khi phát hiện mẫu lệch này, ưu tiên created_at để hiển thị giờ quay thực tế.
  if (drawMs > 0 && createdMs > 0) {
    const deltaMs = Math.abs(drawMs - createdMs)
    if (Math.abs(deltaMs - 7 * 60 * 60 * 1000) <= 2 * 60 * 1000) {
      return formatClockMs(createdMs)
    }
  }

  if (drawMs > 0) return formatClockMs(drawMs)
  if (createdMs > 0) return formatClockMs(createdMs)
  return '—'
}

function diceColor(n: number): string {
  return n <= 3 ? '#e8404a' : '#10b981'
}

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
</script>

<template>
  <div class="mx-3 mt-2 rounded-[16px] bg-white shadow-sm border border-slate-100 overflow-hidden">
    <div class="flex bg-[#fff5f5] border-b border-[#f0e0e0]">
      <button
        type="button"
        class="flex-1 py-2.5 text-[0.72rem] font-semibold border-b-2 transition-all"
        :class="props.activeHistoryTab === 'history' ? 'border-[#e8404a] text-[#e8404a] bg-white' : 'border-transparent text-slate-500'"
        @click="emit('change-tab', 'history')"
      >Lịch sử trò chơi</button>
      <button
        type="button"
        class="flex-1 py-2.5 text-[0.72rem] font-semibold border-b-2 transition-all"
        :class="props.activeHistoryTab === 'chart' ? 'border-[#e8404a] text-[#e8404a] bg-white' : 'border-transparent text-slate-500'"
        @click="emit('change-tab', 'chart')"
      >Biểu đồ</button>
      <button
        type="button"
        class="flex-1 py-2.5 text-[0.72rem] font-semibold border-b-2 transition-all"
        :class="props.activeHistoryTab === 'mine' ? 'border-[#e8404a] text-[#e8404a] bg-white' : 'border-transparent text-slate-500'"
        @click="emit('change-tab', 'mine')"
      >Lịch sử của tôi</button>
    </div>

    <div v-show="props.activeHistoryTab === 'chart'" class="px-3 py-3">
      <div class="mb-3 flex items-center justify-between">
        <div>
          <p class="text-[0.68rem] uppercase tracking-[0.12em] text-slate-400">Biểu đồ kết quả</p>
          <strong class="text-[0.9rem] font-black text-on-surface">24 kỳ gần nhất</strong>
        </div>
        <button
          type="button"
          class="rounded-full bg-[#fff5f5] px-3 py-1.5 text-[0.7rem] font-black text-primary"
          @click="emit('refresh-chart')"
        >
          Làm mới
        </button>
      </div>

      <div v-if="props.chartError" class="rounded-[14px] bg-red-50 px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">
        {{ props.chartError }}
      </div>
      <div v-else class="relative rounded-[18px] bg-[#fff9f9] p-3">
        <div class="flex items-end gap-2 overflow-x-auto pb-2 no-scrollbar">
          <div
            v-for="(item, idx) in visibleChartSeries"
            :key="'chart-item-' + (item.periodNo || idx)"
            class="flex min-w-[52px] flex-col items-center gap-2"
          >
            <div class="flex h-28 w-full items-end">
              <div
                class="w-full rounded-t-[12px] transition-all"
                :class="item.barClass"
                :style="{ height: `${Math.max(20, (item.value / Math.max(1, props.chartMaxValue)) * 100)}%` }"
              />
            </div>
            <span class="rounded-full px-2 py-0.5 text-[0.6rem] font-black text-white" :class="item.barClass">{{ item.label }}</span>
            <span class="text-[0.62rem] font-semibold text-slate-400">{{ item.periodNo ? item.periodNo.slice(-4) : '—' }}</span>
          </div>
        </div>
        <div v-if="props.chartLoading" class="absolute inset-0 flex items-center justify-center rounded-[18px] bg-white/60 text-[0.82rem] font-semibold text-slate-400 backdrop-blur-[1px]">
          Đang tải dữ liệu biểu đồ...
        </div>
      </div>
    </div>

    <div v-if="props.activeHistoryTab === 'history'" class="overflow-hidden">
      <div v-if="props.historyError" class="px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">{{ props.historyError }}</div>
      <div v-else class="relative overflow-hidden">
        <div class="grid grid-cols-[1.35fr_0.7fr_1fr_0.7fr_0.8fr] bg-[#f8fafc] border-b border-slate-200 px-3 py-2 text-[0.62rem] font-black uppercase tracking-[0.06em] text-slate-500">
          <span>Kỳ</span>
          <span class="text-center">KQ</span>
          <span>Lớn/Nhỏ</span>
          <span class="text-center">Màu</span>
          <span class="text-right">Giờ quay</span>
        </div>
        <div
          v-for="row in visibleHistoryRows"
          :key="`${row.period_index || 0}-${row.period_no}`"
          class="grid grid-cols-[1.35fr_0.7fr_1fr_0.7fr_0.8fr] items-center border-b border-slate-100 px-3 py-2.5 text-[0.78rem] hover:bg-slate-50"
        >
          <div class="min-w-0">
            <p class="truncate text-[0.66rem] font-bold text-slate-700">
              {{ row.period_index ? `#${row.period_index}` : periodTail(row.period_no, 10) }}
            </p>
            <p class="mt-0.5 text-[0.6rem] text-slate-400">{{ row.period_no ? row.period_no.split('_')[0] : '—' }}</p>
          </div>
          <span
            class="flex h-7 w-7 mx-auto items-center justify-center rounded-full text-[0.75rem] font-black text-white shadow-sm"
            :class="resultBadgeClass(row.color)"
            :style="resultBadgeStyle(row.color)"
          >{{ row.result ? row.result.slice(0, 1) : '—' }}</span>
          <span class="font-semibold" :class="(row.big_small?.toLowerCase().includes('lớn') || row.big_small?.toLowerCase().includes('big')) ? 'text-[#e8404a]' : 'text-[#2563eb]'">
            {{ normalizeBetLabel(row.big_small) }}
          </span>
          <span class="mx-auto flex h-5 min-w-[44px] items-center justify-center rounded-full px-2 text-[0.62rem] font-black text-white"
                :class="resultBadgeClass(row.color)"
                :style="resultBadgeStyle(row.color)">
            {{ normalizeBetLabel(row.color) }}
          </span>
          <span class="text-right text-[0.66rem] font-semibold text-slate-500">{{ formatDrawClock(row.draw_at, row.created_at) }}</span>
        </div>
        <div v-if="!visibleHistoryRows.length" class="flex flex-col items-center gap-2 py-8 text-slate-300">
          <span class="material-symbols-outlined text-[2rem]">history</span>
          <p class="text-[0.82rem]">Không có dữ liệu</p>
        </div>
        <div v-if="props.historyLoading" class="absolute inset-0 flex items-center justify-center bg-white/60 text-[0.82rem] font-semibold text-slate-400 backdrop-blur-[1px]">
          Đang tải dữ liệu...
        </div>
      </div>
    </div>

    <div v-else-if="props.activeHistoryTab === 'mine'" class="divide-y divide-[#f8f0f0] relative">
      <div v-if="props.mineError" class="px-4 py-3 text-[0.78rem] font-semibold text-[#e64545]">{{ props.mineError }}</div>
      <template v-else>
        <button
          v-for="row in visibleMineRows"
          :key="row.id"
          type="button"
          class="w-full px-3 py-3 text-left transition-colors hover:bg-[#fff9f9]"
          @click="emit('open-ticket-detail', row)"
        >
          <div class="flex items-start gap-3">
            <div class="mt-0.5 flex gap-1 flex-shrink-0">
              <div
                v-if="props.isK3"
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
              <p class="truncate text-[0.72rem] font-semibold text-slate-400">
                {{ row.period_index ? `#${row.period_index}` : (row.period_no || '—') }}
              </p>
              <p class="truncate text-[0.62rem] text-slate-300">{{ row.period_no || '—' }}</p>
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
        <div v-if="!visibleMineRows.length" class="flex flex-col items-center gap-2 py-8 text-slate-300">
          <span class="material-symbols-outlined text-[2rem]">history</span>
          <p class="text-[0.82rem]">Không có lịch sử cược</p>
        </div>
        <div v-if="props.mineLoading" class="absolute inset-0 flex items-center justify-center bg-white/60 text-[0.82rem] font-semibold text-slate-400 backdrop-blur-[1px]">
          Đang tải dữ liệu...
        </div>
      </template>
    </div>

    <div v-if="props.activeHistoryTab !== 'chart'" class="flex items-center justify-between px-3 py-3 border-t border-[#f0e0e0]">
      <button
        type="button"
        class="flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 text-slate-400 disabled:opacity-30 transition-all"
        :disabled="activePage <= 1"
        @click="setPage(Math.max(1, activePage - 1))"
      >
        <span class="material-symbols-outlined text-[1.1rem]">chevron_left</span>
      </button>
      <span class="text-[0.75rem] text-slate-500 font-semibold">{{ activePage }} / {{ activeTotalPages }}</span>
      <button
        type="button"
        class="flex h-8 w-8 items-center justify-center rounded-full border border-[#e8404a] bg-[#e8404a] text-white disabled:opacity-30 transition-all"
        :disabled="activePage >= activeTotalPages"
        @click="setPage(Math.min(activeTotalPages, activePage + 1))"
      >
        <span class="material-symbols-outlined text-[1.1rem]">chevron_right</span>
      </button>
    </div>
  </div>
</template>
