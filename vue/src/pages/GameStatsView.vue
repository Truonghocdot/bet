<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { request, type ApiError } from '@/shared/api/http'
import type { PlayRoomBetHistoryResponse, PlayRoomItem } from '@/shared/api/types'
import { formatViDateTime } from '@/shared/lib/date'
import { formatViMoney } from '@/shared/lib/money'
import { useAuthStore } from '@/stores/auth'

type RoomStatsRow = {
  roomCode: string
  tickets: number
  wins: number
  losses: number
  pending: number
  stakeSample: number
  profitSample: number
  lastBetAt: string | null
}

const auth = useAuthStore()
const loading = ref(false)
const error = ref('')
const rows = ref<RoomStatsRow[]>([])

const totalTickets = computed(() => rows.value.reduce((sum, item) => sum + item.tickets, 0))
const totalStakeSample = computed(() => rows.value.reduce((sum, item) => sum + item.stakeSample, 0))
const totalProfitSample = computed(() => rows.value.reduce((sum, item) => sum + item.profitSample, 0))
const activeRooms = computed(() => rows.value.filter((item) => item.tickets > 0).length)

function parseAmount(value: string | number | undefined | null): number {
  const parsed = Number(value ?? 0)
  return Number.isFinite(parsed) ? parsed : 0
}

function toRow(roomCode: string, tickets: PlayRoomBetHistoryResponse): RoomStatsRow {
  let wins = 0
  let losses = 0
  let pending = 0
  let stakeSample = 0
  let profitSample = 0

  for (const item of tickets.items) {
    const status = String(item.status || '').toUpperCase()
    if (status === 'WON' || status === 'HALF_WON') wins += 1
    else if (status === 'LOST' || status === 'HALF_LOST') losses += 1
    else pending += 1

    stakeSample += parseAmount(item.original_amount ?? item.stake)
    profitSample += parseAmount(item.profit_loss)
  }

  return {
    roomCode,
    tickets: tickets.total,
    wins,
    losses,
    pending,
    stakeSample,
    profitSample,
    lastBetAt: tickets.items[0]?.created_at ?? null,
  }
}

async function fetchRoomStats(roomCode: string) {
  const token = auth.accessToken
  const response = await request<PlayRoomBetHistoryResponse>(
    'GET',
    `/v1/play/rooms/${encodeURIComponent(roomCode)}/bets?page=1&page_size=50`,
    { token },
  )
  return toRow(roomCode, response)
}

async function loadStats() {
  if (!auth.accessToken) {
    rows.value = []
    return
  }

  loading.value = true
  error.value = ''
  try {
    const roomList = await request<{ items: PlayRoomItem[] }>('GET', '/v1/play/rooms')
    const activeRoomCodes = (roomList.items ?? []).map((room) => room.code).filter(Boolean)

    const result = await Promise.all(
      activeRoomCodes.map(async (roomCode) => {
        try {
          return await fetchRoomStats(roomCode)
        } catch {
          return {
            roomCode,
            tickets: 0,
            wins: 0,
            losses: 0,
            pending: 0,
            stakeSample: 0,
            profitSample: 0,
            lastBetAt: null,
          } as RoomStatsRow
        }
      }),
    )

    rows.value = result.sort((a, b) => b.tickets - a.tickets)
  } catch (e: any) {
    const err = e as ApiError
    error.value = err?.message ?? 'Không thể tải thống kê trò chơi'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadStats()
})
</script>

<template>
  <div class="space-y-4">
    <section class="rounded-[20px] bg-white p-4 shadow-[0_8px_20px_rgba(255,109,102,0.05)]">
      <div class="flex items-center justify-between">
        <h1 class="text-[1rem] font-black text-on-surface">Thống kê trò chơi</h1>
        <RouterLink to="/play" class="rounded-full bg-primary px-3 py-1 text-[0.72rem] font-black text-white">Vào phòng chơi</RouterLink>
      </div>
      <p class="mt-2 text-[0.78rem] text-on-surface-variant">
        Dữ liệu lấy từ lịch sử cược thật của tài khoản theo từng room.
      </p>
    </section>

    <section class="grid gap-2 md:grid-cols-4">
      <article class="rounded-[16px] bg-white p-4 shadow-[0_8px_20px_rgba(255,109,102,0.05)]">
        <p class="text-[0.68rem] font-bold uppercase tracking-[0.05em] text-on-surface-variant">Tổng lệnh</p>
        <p class="mt-1 text-[1.2rem] font-black">{{ totalTickets }}</p>
      </article>
      <article class="rounded-[16px] bg-white p-4 shadow-[0_8px_20px_rgba(255,109,102,0.05)]">
        <p class="text-[0.68rem] font-bold uppercase tracking-[0.05em] text-on-surface-variant">Room đã chơi</p>
        <p class="mt-1 text-[1.2rem] font-black">{{ activeRooms }}</p>
      </article>
      <article class="rounded-[16px] bg-white p-4 shadow-[0_8px_20px_rgba(255,109,102,0.05)]">
        <p class="text-[0.68rem] font-bold uppercase tracking-[0.05em] text-on-surface-variant">Tổng cược mẫu</p>
        <p class="mt-1 text-[1.2rem] font-black">{{ formatViMoney(totalStakeSample) }}</p>
      </article>
      <article class="rounded-[16px] bg-white p-4 shadow-[0_8px_20px_rgba(255,109,102,0.05)]">
        <p class="text-[0.68rem] font-bold uppercase tracking-[0.05em] text-on-surface-variant">Lãi/lỗ mẫu</p>
        <p class="mt-1 text-[1.2rem] font-black" :class="totalProfitSample >= 0 ? 'text-emerald-600' : 'text-red-500'">
          {{ totalProfitSample >= 0 ? '+' : '' }}{{ formatViMoney(totalProfitSample) }}
        </p>
      </article>
    </section>

    <section v-if="error" class="rounded-[14px] bg-red-50 px-4 py-3 text-sm font-semibold text-red-600">
      {{ error }}
    </section>

    <section v-if="loading" class="rounded-[14px] border border-slate-100 bg-white px-4 py-4 text-sm font-semibold text-slate-500">
      Đang tải thống kê...
    </section>

    <section v-else class="overflow-hidden rounded-[16px] border border-slate-100 bg-white shadow-[0_8px_20px_rgba(255,109,102,0.05)]">
      <div class="grid grid-cols-[1.4fr_repeat(5,minmax(0,1fr))] gap-2 border-b border-slate-100 bg-slate-50 px-4 py-3 text-[0.7rem] font-black uppercase tracking-[0.04em] text-on-surface-variant">
        <span>Room</span>
        <span>Tổng lệnh</span>
        <span>Thắng</span>
        <span>Thua</span>
        <span>Lãi/lỗ mẫu</span>
        <span>Lệnh gần nhất</span>
      </div>

      <div v-if="rows.length === 0" class="px-4 py-4 text-sm font-semibold text-slate-500">
        Chưa có dữ liệu cược.
      </div>

      <div
        v-for="row in rows"
        :key="row.roomCode"
        class="grid grid-cols-[1.4fr_repeat(5,minmax(0,1fr))] gap-2 border-b border-slate-100 px-4 py-3 text-[0.8rem] last:border-b-0"
      >
        <span class="font-black uppercase">{{ row.roomCode }}</span>
        <span>{{ row.tickets }}</span>
        <span class="text-emerald-600 font-bold">{{ row.wins }}</span>
        <span class="text-red-500 font-bold">{{ row.losses }}</span>
        <span :class="row.profitSample >= 0 ? 'text-emerald-600 font-bold' : 'text-red-500 font-bold'">
          {{ row.profitSample >= 0 ? '+' : '' }}{{ formatViMoney(row.profitSample) }}
        </span>
        <span class="text-on-surface-variant">{{ row.lastBetAt ? formatViDateTime(row.lastBetAt) : '—' }}</span>
      </div>
    </section>
  </div>
</template>

