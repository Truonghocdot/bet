<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { request, type ApiError } from '@/shared/api/http'
import type { AgencyManagedUserDeposit, AgencyManagedUserDepositHistoryResponse } from '@/shared/api/types'
import { useAgencyAuthStore } from '@/stores/auth'

const auth = useAgencyAuthStore()
const loading = ref(false)
const error = ref('')
const items = ref<AgencyManagedUserDeposit[]>([])
const page = ref(1)
const pageSize = 10
const total = ref(0)
const totalPages = ref(1)

const pageAmountTotal = computed(() => items.value.reduce((sum, item) => sum + Number(item.amount || 0), 0))
const pageCompletedCount = computed(() => items.value.filter((item) => item.status === 3).length)

function depositStatusLabel(status: number) {
  if (status === 1) return 'Đang chờ'
  if (status === 2) return 'Đã xác nhận'
  if (status === 3) return 'Hoàn tất'
  if (status === 4) return 'Thất bại'
  if (status === 5) return 'Đã hủy'
  return `Mã ${status}`
}

function depositStatusTone(status: number) {
  if (status === 1 || status === 2) return 'bg-amber-100 text-amber-700'
  if (status === 3) return 'bg-emerald-100 text-emerald-700'
  if (status === 4 || status === 5) return 'bg-rose-100 text-rose-700'
  return 'bg-slate-100 text-slate-600'
}

function unitLabel(unit: number) {
  return unit === 2 ? 'USDT' : 'VND'
}

function providerLabel(item: AgencyManagedUserDeposit) {
  return String(item.provider || '').trim() || '—'
}

function receivingAccountLabel(item: AgencyManagedUserDeposit) {
  const account = item.receiving_account
  if (!account) return '—'
  const name = String(account.account_name || '').trim()
  const number = String(account.account_number || '').trim()
  const providerCode = String(account.provider_code || '').trim()
  return [name, number, providerCode].filter(Boolean).join(' • ') || '—'
}

async function loadDeposits(nextPage = page.value) {
  if (!auth.accessToken) return

  loading.value = true
  error.value = ''
  try {
    const res = await request<AgencyManagedUserDepositHistoryResponse>(
      'GET',
      `/v1/affiliate/managed-deposits?page=${nextPage}&page_size=${pageSize}`,
      { token: auth.accessToken },
    )
    items.value = res.data ?? []
    page.value = res.page || nextPage
    total.value = res.total || 0
    totalPages.value = res.total_pages || 1
  } catch (e: any) {
    error.value = (e as ApiError)?.message ?? 'Không thể tải giao dịch nạp tiền của agency'
    items.value = []
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadDeposits(1)
})
</script>

<template>
  <div class="mx-auto max-w-7xl space-y-6">
    <section class="rounded-[32px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-6 shadow-[0_30px_80px_rgba(15,23,42,0.12)]">
      <p class="m-0 text-xs font-bold uppercase tracking-[0.18em] text-rose-600">Giao dịch nạp tiền</p>
      <h1 class="mt-3 text-3xl font-bold text-slate-950">Toàn bộ user thuộc agency</h1>
      <p class="mt-2 text-sm text-slate-500">Theo dõi trạng thái giao dịch nạp, người thực hiện nạp, nhà cung cấp và tài khoản nhận tiền.</p>
    </section>

    <section class="grid gap-4 md:grid-cols-3">
      <article class="rounded-[24px] bg-slate-950 p-5 text-white">
        <p class="text-xs uppercase tracking-[0.16em] text-white/60">Tổng giao dịch</p>
        <strong class="mt-2 block text-3xl">{{ total }}</strong>
      </article>
      <article class="rounded-[24px] bg-white p-5 shadow-[0_18px_50px_rgba(15,23,42,0.08)]">
        <p class="text-xs uppercase tracking-[0.16em] text-slate-400">Tổng tiền trang hiện tại</p>
        <strong class="mt-2 block text-3xl text-slate-950">{{ pageAmountTotal.toLocaleString('vi-VN') }}</strong>
      </article>
      <article class="rounded-[24px] bg-white p-5 shadow-[0_18px_50px_rgba(15,23,42,0.08)]">
        <p class="text-xs uppercase tracking-[0.16em] text-slate-400">Đã hoàn tất ở trang hiện tại</p>
        <strong class="mt-2 block text-3xl text-slate-950">{{ pageCompletedCount }}</strong>
      </article>
    </section>

    <section class="rounded-[30px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-5 shadow-[0_30px_80px_rgba(15,23,42,0.12)]">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h2 class="m-0 text-2xl font-bold text-slate-950">Bảng giao dịch nạp</h2>
          <p class="mt-2 text-sm text-slate-500">Hiển thị tương tự bảng finance: user, SĐT, mã giao dịch, trạng thái, số tiền và thời gian.</p>
        </div>
        <button type="button" class="rounded-xl bg-slate-950 px-4 py-2.5 text-sm font-bold text-white" @click="void loadDeposits(page)">
          Tải lại
        </button>
      </div>

      <p v-if="error" class="mt-4 rounded-[18px] bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-600">{{ error }}</p>
      <div v-if="loading" class="mt-6 text-sm font-semibold text-slate-500">Đang tải giao dịch nạp tiền...</div>
      <div v-else-if="items.length === 0" class="mt-6 rounded-[22px] border border-dashed border-black/10 px-4 py-10 text-center text-sm font-semibold text-slate-500">
        Chưa có giao dịch nạp tiền nào.
      </div>

      <div v-else class="mt-6 overflow-hidden rounded-[24px] border border-black/6">
        <div class="overflow-x-auto">
          <table class="min-w-full border-collapse">
            <thead class="bg-slate-950 text-left text-xs uppercase tracking-[0.14em] text-white/70">
              <tr>
                <th class="px-4 py-4 font-bold">ID</th>
                <th class="px-4 py-4 font-bold">Người dùng</th>
                <th class="px-4 py-4 font-bold">SĐT</th>
                <th class="px-4 py-4 font-bold">Mã giao dịch CK</th>
                <th class="px-4 py-4 font-bold">Đơn vị</th>
                <th class="px-4 py-4 font-bold">Số tiền</th>
                <th class="px-4 py-4 font-bold">Trạng thái</th>
                <th class="px-4 py-4 font-bold">Nhà cung cấp</th>
                <th class="px-4 py-4 font-bold">Tài khoản nhận</th>
                <th class="px-4 py-4 font-bold">Tạo lúc</th>
              </tr>
            </thead>
            <tbody class="bg-white text-sm text-slate-700">
              <tr v-for="item in items" :key="item.id" class="border-t border-black/6">
                <td class="px-4 py-4 align-top font-bold text-slate-950">{{ item.id }}</td>
                <td class="px-4 py-4 align-top">
                  <div class="font-bold text-slate-950">{{ item.user_name || `User #${item.user_id}` }}</div>
                  <div class="mt-1 text-xs text-slate-500">User ID {{ item.user_id }}</div>
                </td>
                <td class="px-4 py-4 align-top text-slate-600">{{ item.user_phone || '—' }}</td>
                <td class="px-4 py-4 align-top">
                  <div class="font-bold text-slate-950">{{ item.client_ref || '—' }}</div>
                  <div class="mt-1 text-xs text-slate-500">{{ item.provider_txn_id || 'Chưa có provider txn id' }}</div>
                </td>
                <td class="px-4 py-4 align-top text-slate-600">{{ unitLabel(item.unit) }}</td>
                <td class="px-4 py-4 align-top font-bold text-slate-950">{{ Number(item.amount || 0).toLocaleString('vi-VN') }}</td>
                <td class="px-4 py-4 align-top">
                  <span class="rounded-full px-3 py-1 text-xs font-bold uppercase tracking-[0.08em]" :class="depositStatusTone(item.status)">
                    {{ depositStatusLabel(item.status) }}
                  </span>
                </td>
                <td class="px-4 py-4 align-top text-slate-600">{{ providerLabel(item) }}</td>
                <td class="px-4 py-4 align-top text-slate-600">{{ receivingAccountLabel(item) }}</td>
                <td class="px-4 py-4 align-top text-xs font-semibold text-slate-500">{{ item.created_at || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div v-if="total > 0" class="mt-4 flex flex-wrap items-center justify-between gap-3 rounded-[18px] bg-slate-50 px-4 py-3">
        <p class="m-0 text-sm font-semibold text-slate-500">Trang {{ page }} / {{ totalPages }}</p>
        <div class="flex items-center gap-2">
          <button type="button" class="min-h-10 rounded-xl border border-black/8 bg-white px-4 text-sm font-bold text-slate-700 disabled:opacity-40" :disabled="loading || page <= 1" @click="void loadDeposits(page - 1)">Trang trước</button>
          <button type="button" class="min-h-10 rounded-xl border border-black/8 bg-white px-4 text-sm font-bold text-slate-700 disabled:opacity-40" :disabled="loading || page >= totalPages" @click="void loadDeposits(page + 1)">Trang sau</button>
        </div>
      </div>
    </section>
  </div>
</template>
