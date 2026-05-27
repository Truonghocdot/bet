<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import { request, type ApiError } from '@/shared/api/http'
import type { AgencyManagedUserWithdrawal, AgencyManagedUserWithdrawalHistoryResponse } from '@/shared/api/types'
import { useAgencyPlayersStore } from '@/stores/players'
import { useAgencyAuthStore } from '@/stores/auth'

const route = useRoute()
const auth = useAgencyAuthStore()
const players = useAgencyPlayersStore()

const loading = ref(false)
const error = ref('')
const items = ref<AgencyManagedUserWithdrawal[]>([])
const page = ref(1)
const pageSize = 10
const total = ref(0)
const totalPages = ref(1)

const userId = computed(() => Number(route.params.userId || 0))
const player = computed(() => players.findUserById(userId.value))

function withdrawalStatusLabel(status: number) {
  if (status === 1) return 'Đang chờ'
  if (status === 2) return 'Đã duyệt'
  if (status === 3) return 'Từ chối'
  if (status === 4) return 'Đã hủy'
  if (status === 5) return 'Đã chi'
  return `Mã ${status}`
}

function withdrawalStatusTone(status: number) {
  if (status === 1) return 'bg-amber-100 text-amber-700'
  if (status === 2) return 'bg-sky-100 text-sky-700'
  if (status === 5) return 'bg-emerald-100 text-emerald-700'
  if (status === 3 || status === 4) return 'bg-rose-100 text-rose-700'
  return 'bg-slate-100 text-slate-600'
}

function unitLabel(unit: number) {
  return unit === 2 ? 'USDT' : 'VND'
}

function beneficiaryLabel(item: AgencyManagedUserWithdrawal) {
  return [item.account_name, item.account_number, item.provider_code].filter(Boolean).join(' • ') || '—'
}

async function loadWithdrawals(nextPage = page.value) {
  if (!auth.accessToken || !userId.value) return

  loading.value = true
  error.value = ''
  try {
    const res = await request<AgencyManagedUserWithdrawalHistoryResponse>(
      'GET',
      `/v1/affiliate/managed-users/${userId.value}/withdrawals?page=${nextPage}&page_size=${pageSize}`,
      { token: auth.accessToken },
    )
    items.value = res.data ?? []
    page.value = res.page || nextPage
    total.value = res.total || 0
    totalPages.value = res.total_pages || 1
  } catch (e: any) {
    error.value = (e as ApiError)?.message ?? 'Không thể tải lịch sử rút tiền'
    items.value = []
  } finally {
    loading.value = false
  }
}

watch(
  () => userId.value,
  () => {
    if (!userId.value) return
    void loadWithdrawals(1)
  },
)

onMounted(async () => {
  await players.ensureLoaded()
  await loadWithdrawals(1)
})
</script>

<template>
  <div class="mx-auto max-w-7xl space-y-6">
    <section class="rounded-[32px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-6 shadow-[0_30px_80px_rgba(15,23,42,0.12)]">
      <div class="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
        <div>
          <p class="m-0 text-xs font-bold uppercase tracking-[0.18em] text-rose-600">Giao dịch rút tiền</p>
          <h1 class="mt-3 text-3xl font-bold text-slate-950">{{ player?.name || `User #${userId}` }}</h1>
          <p class="mt-2 text-sm text-slate-500">{{ player?.phone || 'Chưa cập nhật số điện thoại' }}</p>
        </div>

        <div class="flex flex-wrap gap-3">
          <RouterLink
            :to="{ name: 'agency-user-stats', params: { userId } }"
            class="inline-flex min-h-11 items-center rounded-xl border border-black/8 bg-white px-4 text-sm font-bold text-slate-700"
          >
            Quay lại hồ sơ
          </RouterLink>
          <button
            type="button"
            class="inline-flex min-h-11 items-center rounded-xl bg-slate-950 px-4 text-sm font-bold text-white"
            @click="void loadWithdrawals(page)"
          >
            Tải lại
          </button>
        </div>
      </div>
    </section>

    <section class="rounded-[30px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-5 shadow-[0_30px_80px_rgba(15,23,42,0.12)]">
      <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
        <div>
          <h2 class="m-0 text-2xl font-bold text-slate-950">Bảng giao dịch rút</h2>
          <p class="mt-2 text-sm text-slate-500">Hiển thị trạng thái, phí, số tiền thực nhận và thông tin tài khoản thụ hưởng.</p>
        </div>
        <div class="rounded-full bg-slate-100 px-3 py-2 text-xs font-bold text-slate-600">
          {{ total }} giao dịch
        </div>
      </div>

      <p v-if="error" class="mt-4 rounded-[18px] bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-600">
        {{ error }}
      </p>

      <div v-if="loading" class="mt-6 text-sm font-semibold text-slate-500">Đang tải lịch sử rút tiền...</div>

      <div v-else-if="items.length === 0" class="mt-6 rounded-[22px] border border-dashed border-black/10 px-4 py-10 text-center text-sm font-semibold text-slate-500">
        Chưa có giao dịch rút tiền nào.
      </div>

      <div v-else class="mt-6 overflow-hidden rounded-[24px] border border-black/6">
        <div class="overflow-x-auto">
          <table class="min-w-full border-collapse">
            <thead class="bg-slate-950 text-left text-xs uppercase tracking-[0.14em] text-white/70">
              <tr>
                <th class="px-4 py-4 font-bold">Thời gian</th>
                <th class="px-4 py-4 font-bold">Đơn vị</th>
                <th class="px-4 py-4 font-bold">Số tiền</th>
                <th class="px-4 py-4 font-bold">Phí</th>
                <th class="px-4 py-4 font-bold">Thực nhận</th>
                <th class="px-4 py-4 font-bold">Tài khoản thụ hưởng</th>
                <th class="px-4 py-4 font-bold">Trạng thái</th>
              </tr>
            </thead>
            <tbody class="bg-white text-sm text-slate-700">
              <tr v-for="item in items" :key="item.id" class="border-t border-black/6">
                <td class="px-4 py-4 align-top text-xs font-semibold text-slate-500">{{ item.created_at || '—' }}</td>
                <td class="px-4 py-4 align-top text-slate-600">{{ unitLabel(item.unit) }}</td>
                <td class="px-4 py-4 align-top font-bold text-slate-950">{{ Number(item.amount || 0).toLocaleString('vi-VN') }}</td>
                <td class="px-4 py-4 align-top text-slate-600">{{ Number(item.fee || 0).toLocaleString('vi-VN') }}</td>
                <td class="px-4 py-4 align-top text-slate-600">{{ Number(item.net_amount || 0).toLocaleString('vi-VN') }}</td>
                <td class="px-4 py-4 align-top text-slate-600">
                  <div>{{ beneficiaryLabel(item) }}</div>
                  <div v-if="item.reason_rejected" class="mt-1 text-xs text-rose-500">{{ item.reason_rejected }}</div>
                </td>
                <td class="px-4 py-4 align-top">
                  <span class="rounded-full px-3 py-1 text-xs font-bold uppercase tracking-[0.08em]" :class="withdrawalStatusTone(item.status)">
                    {{ withdrawalStatusLabel(item.status) }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div v-if="total > 0" class="mt-4 flex flex-wrap items-center justify-between gap-3 rounded-[18px] bg-slate-50 px-4 py-3">
        <p class="m-0 text-sm font-semibold text-slate-500">Trang {{ page }} / {{ totalPages }}</p>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="min-h-10 rounded-xl border border-black/8 bg-white px-4 text-sm font-bold text-slate-700 disabled:opacity-40"
            :disabled="loading || page <= 1"
            @click="void loadWithdrawals(page - 1)"
          >
            Trang trước
          </button>
          <button
            type="button"
            class="min-h-10 rounded-xl border border-black/8 bg-white px-4 text-sm font-bold text-slate-700 disabled:opacity-40"
            :disabled="loading || page >= totalPages"
            @click="void loadWithdrawals(page + 1)"
          >
            Trang sau
          </button>
        </div>
      </div>
    </section>
  </div>
</template>
