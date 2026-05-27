<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { request, type ApiError } from '@/shared/api/http'
import type { ManagedAffiliateUserTransaction } from '@/shared/api/types'
import { useAgencyPlayersStore } from '@/stores/players'
import { useAgencyAuthStore } from '@/stores/auth'

const route = useRoute()
const players = useAgencyPlayersStore()
const auth = useAgencyAuthStore()
const transactions = ref<ManagedAffiliateUserTransaction[]>([])
const transactionsLoading = ref(false)
const transactionsError = ref('')

const userId = computed(() => Number(route.params.userId || 0))
const player = computed(() => players.findUserById(userId.value))

const playerStatusTone = computed(() => {
  if (player.value?.referral_status === 2) return 'bg-emerald-100 text-emerald-700'
  if (player.value?.referral_status === 1) return 'bg-amber-100 text-amber-700'
  return 'bg-slate-100 text-slate-600'
})

function unitLabel(unit: number) {
  return unit === 2 ? 'USDT' : 'VND'
}

function directionLabel(direction: number) {
  if (direction === 1) return 'Cộng'
  if (direction === 2) return 'Trừ'
  return 'Khác'
}

function directionTone(direction: number) {
  if (direction === 1) return 'bg-emerald-100 text-emerald-700'
  if (direction === 2) return 'bg-rose-100 text-rose-700'
  return 'bg-slate-100 text-slate-600'
}

function shortReference(referenceType: string) {
  if (!referenceType) return '—'
  const parts = referenceType.split('\\')
  return parts[parts.length - 1] || referenceType
}

async function loadTransactions() {
  if (!auth.accessToken || !userId.value) return

  transactionsLoading.value = true
  transactionsError.value = ''
  try {
    const res = await request<{ items: ManagedAffiliateUserTransaction[] }>(
      'GET',
      `/v1/affiliate/managed-users/${userId.value}/transactions`,
      { token: auth.accessToken },
    )
    transactions.value = res.items ?? []
  } catch (e: any) {
    transactionsError.value = (e as ApiError)?.message ?? 'Không thể tải giao dịch người chơi'
    transactions.value = []
  } finally {
    transactionsLoading.value = false
  }
}

onMounted(async () => {
  await players.ensureLoaded()
  await loadTransactions()
})
</script>

<template>
  <div class="mx-auto max-w-6xl space-y-6">
    <section class="rounded-[32px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-6 shadow-[0_30px_80px_rgba(15,23,42,0.12)]">
      <div v-if="player" class="flex flex-col gap-5 xl:flex-row xl:items-start xl:justify-between">
        <div>
          <p class="m-0 text-xs font-bold uppercase tracking-[0.18em] text-rose-600">Thống kê người chơi</p>
          <h1 class="mt-3 text-3xl font-bold text-slate-950">{{ player.name || `User #${player.user_id}` }}</h1>
          <p class="mt-2 text-sm text-slate-500">{{ player.phone || 'Chưa cập nhật số điện thoại' }}</p>
          <div class="mt-4 flex flex-wrap gap-2">
            <span class="rounded-full bg-slate-100 px-3 py-1.5 text-xs font-bold text-slate-600">User ID {{ player.user_id }}</span>
            <span class="rounded-full px-3 py-1.5 text-xs font-bold" :class="playerStatusTone">
              {{ players.statusLabel(player.referral_status) }}
            </span>
          </div>
        </div>

        <div class="flex flex-wrap gap-3">
          <RouterLink
            :to="{ name: 'agency-user-deposits', params: { userId: userId } }"
            class="inline-flex min-h-11 items-center rounded-xl border border-black/8 bg-white px-4 text-sm font-bold text-slate-700"
          >
            Giao dịch nạp
          </RouterLink>
          <RouterLink
            :to="{ name: 'agency-user-withdrawals', params: { userId: userId } }"
            class="inline-flex min-h-11 items-center rounded-xl border border-black/8 bg-white px-4 text-sm font-bold text-slate-700"
          >
            Giao dịch rút
          </RouterLink>
          <RouterLink
            :to="{ name: 'agency-users' }"
            class="inline-flex min-h-11 items-center rounded-xl border border-black/8 bg-white px-4 text-sm font-bold text-slate-700"
          >
            Quay lại danh sách
          </RouterLink>
          <RouterLink
            :to="{ name: 'agency-overview' }"
            class="inline-flex min-h-11 items-center rounded-xl bg-slate-950 px-4 text-sm font-bold text-white"
          >
            Tổng quan agency
          </RouterLink>
        </div>
      </div>

      <div v-else class="text-sm font-semibold text-slate-500">
        Không tìm thấy người chơi trong danh sách quản lý hiện tại.
      </div>
    </section>

    <template v-if="player">
      <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <article class="rounded-[28px] bg-slate-950 p-5 text-white">
          <p class="text-xs uppercase tracking-[0.18em] text-white/60">Trạng thái referral</p>
          <strong class="mt-3 block text-2xl">{{ players.statusLabel(player.referral_status) }}</strong>
        </article>
        <article class="rounded-[28px] bg-white p-5 shadow-[0_18px_50px_rgba(15,23,42,0.08)]">
          <p class="text-xs uppercase tracking-[0.18em] text-slate-400">Thời gian tham gia</p>
          <strong class="mt-3 block text-2xl text-slate-950">{{ player.created_at || '—' }}</strong>
        </article>
        <article class="rounded-[28px] bg-white p-5 shadow-[0_18px_50px_rgba(15,23,42,0.08)]">
          <p class="text-xs uppercase tracking-[0.18em] text-slate-400">First deposit</p>
          <strong class="mt-3 block text-2xl text-slate-950">{{ Number(player.first_deposit_amount || 0).toLocaleString('vi-VN') }}</strong>
        </article>
        <article class="rounded-[28px] bg-gradient-to-br from-amber-200 to-rose-200 p-5">
          <p class="text-xs uppercase tracking-[0.18em] text-slate-700">Mã giao dịch đầu tiên</p>
          <strong class="mt-3 block break-all text-xl text-slate-950">{{ player.first_deposit_transaction_no || '—' }}</strong>
        </article>
      </section>

      <section class="grid gap-5 xl:grid-cols-[1fr_0.95fr]">
        <section class="rounded-[30px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-5 shadow-[0_30px_80px_rgba(15,23,42,0.12)]">
          <h2 class="m-0 text-2xl font-bold text-slate-950">Thông tin người chơi</h2>
          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <article class="rounded-[22px] bg-slate-50 p-4">
              <p class="m-0 text-xs uppercase tracking-[0.14em] text-slate-400">Tên hiển thị</p>
              <strong class="mt-2 block text-lg text-slate-950">{{ player.name || '—' }}</strong>
            </article>
            <article class="rounded-[22px] bg-slate-50 p-4">
              <p class="m-0 text-xs uppercase tracking-[0.14em] text-slate-400">Số điện thoại</p>
              <strong class="mt-2 block text-lg text-slate-950">{{ player.phone || '—' }}</strong>
            </article>
            <article class="rounded-[22px] bg-slate-50 p-4">
              <p class="m-0 text-xs uppercase tracking-[0.14em] text-slate-400">Ngày tạo</p>
              <strong class="mt-2 block text-lg text-slate-950">{{ player.created_at || '—' }}</strong>
            </article>
            <article class="rounded-[22px] bg-slate-50 p-4">
              <p class="m-0 text-xs uppercase tracking-[0.14em] text-slate-400">Referral status</p>
              <strong class="mt-2 block text-lg text-slate-950">{{ players.statusLabel(player.referral_status) }}</strong>
            </article>
          </div>
        </section>

        <section class="rounded-[30px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-5 shadow-[0_30px_80px_rgba(15,23,42,0.12)]">
          <h2 class="m-0 text-2xl font-bold text-slate-950">Tóm tắt chuyển đổi</h2>
          <div class="mt-5 space-y-4">
            <article class="rounded-[22px] border border-black/6 bg-white px-4 py-4">
              <p class="m-0 text-xs uppercase tracking-[0.14em] text-slate-400">Đã có first deposit</p>
              <strong class="mt-2 block text-2xl text-slate-950">{{ Number(player.first_deposit_amount || 0) > 0 ? 'Có' : 'Chưa' }}</strong>
            </article>
            <article class="rounded-[22px] border border-black/6 bg-white px-4 py-4">
              <p class="m-0 text-xs uppercase tracking-[0.14em] text-slate-400">Số tiền nạp đầu</p>
              <strong class="mt-2 block text-2xl text-slate-950">{{ Number(player.first_deposit_amount || 0).toLocaleString('vi-VN') }}</strong>
            </article>
            <article class="rounded-[22px] border border-black/6 bg-white px-4 py-4">
              <p class="m-0 text-xs uppercase tracking-[0.14em] text-slate-400">Mã giao dịch đầu tiên</p>
              <strong class="mt-2 block break-all text-base text-slate-950">{{ player.first_deposit_transaction_no || '—' }}</strong>
            </article>
          </div>
        </section>
      </section>

      <section class="grid gap-4 md:grid-cols-2">
        <RouterLink
          :to="{ name: 'agency-user-deposits', params: { userId: userId } }"
          class="rounded-[26px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-5 shadow-[0_20px_60px_rgba(15,23,42,0.1)]"
        >
          <p class="text-xs uppercase tracking-[0.16em] text-slate-400">Chi tiết tài chính</p>
          <strong class="mt-2 block text-2xl text-slate-950">Lịch sử nạp tiền</strong>
          <p class="mt-2 text-sm text-slate-500">Xem trạng thái nạp, mã giao dịch, nhà cung cấp và thông tin tài khoản nhận tiền.</p>
        </RouterLink>
        <RouterLink
          :to="{ name: 'agency-user-withdrawals', params: { userId: userId } }"
          class="rounded-[26px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-5 shadow-[0_20px_60px_rgba(15,23,42,0.1)]"
        >
          <p class="text-xs uppercase tracking-[0.16em] text-slate-400">Chi tiết tài chính</p>
          <strong class="mt-2 block text-2xl text-slate-950">Lịch sử rút tiền</strong>
          <p class="mt-2 text-sm text-slate-500">Xem trạng thái rút, phí, số tiền thực nhận và thông tin tài khoản thụ hưởng của người chơi.</p>
        </RouterLink>
      </section>

      <section class="rounded-[30px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-5 shadow-[0_30px_80px_rgba(15,23,42,0.12)]">
        <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 class="m-0 text-2xl font-bold text-slate-950">Lịch sử giao dịch</h2>
            <p class="mt-2 text-sm text-slate-500">Dữ liệu chỉ đọc từ ví của người chơi, agency không thể thao tác thay đổi.</p>
          </div>
          <button
            type="button"
            class="rounded-xl border border-black/8 bg-white px-4 py-2.5 text-sm font-bold text-slate-700"
            @click="void loadTransactions()"
          >
            Tải lại giao dịch
          </button>
        </div>

        <p v-if="transactionsError" class="mt-4 rounded-[18px] bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-600">
          {{ transactionsError }}
        </p>

        <div v-if="transactionsLoading" class="mt-6 text-sm font-semibold text-slate-500">Đang tải giao dịch...</div>

        <div v-else-if="transactions.length === 0" class="mt-6 rounded-[22px] border border-dashed border-black/10 px-4 py-10 text-center text-sm font-semibold text-slate-500">
          Chưa có giao dịch nào để hiển thị.
        </div>

        <div v-else class="mt-6 overflow-hidden rounded-[24px] border border-black/6">
          <div class="overflow-x-auto">
            <table class="min-w-full border-collapse">
              <thead class="bg-slate-950 text-left text-xs uppercase tracking-[0.14em] text-white/70">
                <tr>
                  <th class="px-4 py-4 font-bold">Thời gian</th>
                  <th class="px-4 py-4 font-bold">Loại</th>
                  <th class="px-4 py-4 font-bold">Chiều</th>
                  <th class="px-4 py-4 font-bold">Số tiền</th>
                  <th class="px-4 py-4 font-bold">Trước</th>
                  <th class="px-4 py-4 font-bold">Sau</th>
                  <th class="px-4 py-4 font-bold">Ghi chú</th>
                </tr>
              </thead>
              <tbody class="bg-white text-sm text-slate-700">
                <tr
                  v-for="item in transactions"
                  :key="item.id"
                  class="border-t border-black/6"
                >
                  <td class="px-4 py-4 align-top text-xs font-semibold text-slate-500">
                    {{ item.created_at || '—' }}
                  </td>
                  <td class="px-4 py-4 align-top">
                    <div class="font-bold text-slate-950">{{ shortReference(item.reference_type) }}</div>
                    <div class="mt-1 text-xs text-slate-500">#{{ item.reference_id ?? item.id }} • {{ unitLabel(item.unit) }}</div>
                  </td>
                  <td class="px-4 py-4 align-top">
                    <span class="rounded-full px-3 py-1 text-xs font-bold uppercase tracking-[0.08em]" :class="directionTone(item.direction)">
                      {{ directionLabel(item.direction) }}
                    </span>
                  </td>
                  <td class="px-4 py-4 align-top font-bold text-slate-950">
                    {{ Number(item.amount || 0).toLocaleString('vi-VN') }}
                  </td>
                  <td class="px-4 py-4 align-top text-slate-600">
                    {{ Number(item.balance_before || 0).toLocaleString('vi-VN') }}
                  </td>
                  <td class="px-4 py-4 align-top text-slate-600">
                    {{ Number(item.balance_after || 0).toLocaleString('vi-VN') }}
                  </td>
                  <td class="px-4 py-4 align-top text-slate-500">
                    {{ item.note || '—' }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>
