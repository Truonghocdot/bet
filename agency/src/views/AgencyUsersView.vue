<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { useAgencyPlayersStore } from '@/stores/players'

const players = useAgencyPlayersStore()
const search = ref('')

const filteredItems = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  if (!keyword) return players.items
  return players.items.filter((item) =>
    [item.name, item.phone, item.first_deposit_transaction_no]
      .join(' ')
      .toLowerCase()
      .includes(keyword),
  )
})

onMounted(() => {
  void players.refreshDashboard()
})
</script>

<template>
  <div class="mx-auto max-w-7xl space-y-6">
    <section class="rounded-[30px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-5 shadow-[0_30px_80px_rgba(15,23,42,0.12)]">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h1 class="m-0 text-2xl font-bold text-slate-950">Danh sách người chơi</h1>
          <p class="mt-2 text-sm text-slate-500">Theo dõi dữ liệu từng người chơi và mở trang thống kê riêng cho từng tài khoản.</p>
        </div>

        <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
          <RouterLink
            :to="{ name: 'agency-overview' }"
            class="inline-flex min-h-11 items-center justify-center rounded-xl border border-black/8 bg-white px-4 text-sm font-bold text-slate-700"
          >
            Thống kê tổng quan
          </RouterLink>
          <button
            type="button"
            class="min-h-11 rounded-xl bg-slate-950 px-4 text-sm font-bold text-white"
            @click="void players.refreshDashboard()"
          >
            Tải lại
          </button>
        </div>
      </div>

      <div class="mt-4">
        <input
          v-model="search"
          type="text"
          placeholder="Tìm tên, số điện thoại, mã giao dịch..."
          class="min-h-12 w-full rounded-[18px] border border-black/8 bg-white px-4 outline-none lg:max-w-sm"
        />
      </div>

      <p v-if="players.error" class="mt-4 rounded-[18px] bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-600">
        {{ players.error }}
      </p>

      <div v-if="players.loading" class="mt-6 text-sm font-semibold text-slate-500">Đang tải dữ liệu...</div>

      <div v-else-if="filteredItems.length === 0" class="mt-6 rounded-[22px] border border-dashed border-black/10 px-4 py-10 text-center text-sm font-semibold text-slate-500">
        Không có người chơi phù hợp bộ lọc hiện tại.
      </div>

      <div v-else class="mt-6 overflow-hidden rounded-[24px] border border-black/6">
        <div class="overflow-x-auto">
          <table class="min-w-full border-collapse">
            <thead class="bg-slate-950 text-left text-xs uppercase tracking-[0.14em] text-white/70">
              <tr>
                <th class="px-4 py-4 font-bold">Người chơi</th>
                <th class="px-4 py-4 font-bold">Thời gian tạo</th>
                <th class="px-4 py-4 font-bold">Trạng thái</th>
                <th class="px-4 py-4 font-bold">Nạp đầu</th>
                <th class="px-4 py-4 font-bold">Mã giao dịch</th>
                <th class="px-4 py-4 font-bold text-right">Hành động</th>
              </tr>
            </thead>
            <tbody class="bg-white text-sm text-slate-700">
              <tr
                v-for="item in filteredItems"
                :key="item.user_id"
                class="border-t border-black/6"
              >
                <td class="px-4 py-4 align-top">
                  <strong class="block text-slate-950">{{ item.name || `User #${item.user_id}` }}</strong>
                  <span class="mt-1 block text-xs text-slate-500">{{ item.phone || 'Chưa cập nhật số điện thoại' }}</span>
                </td>
                <td class="px-4 py-4 align-top text-xs font-semibold text-slate-500">
                  {{ item.created_at || '—' }}
                </td>
                <td class="px-4 py-4 align-top">
                  <span
                    class="rounded-full px-3 py-1 text-xs font-bold uppercase tracking-[0.08em]"
                    :class="item.referral_status === 2 ? 'bg-emerald-100 text-emerald-700' : item.referral_status === 1 ? 'bg-amber-100 text-amber-700' : 'bg-slate-100 text-slate-600'"
                  >
                    {{ players.statusLabel(item.referral_status) }}
                  </span>
                </td>
                <td class="px-4 py-4 align-top font-bold text-slate-950">
                  {{ Number(item.first_deposit_amount || 0).toLocaleString('vi-VN') }}
                </td>
                <td class="px-4 py-4 align-top text-xs font-semibold text-slate-500">
                  {{ item.first_deposit_transaction_no || '—' }}
                </td>
                <td class="px-4 py-4 align-top text-right">
                  <RouterLink
                    :to="{ name: 'agency-user-stats', params: { userId: item.user_id } }"
                    class="inline-flex min-h-10 items-center rounded-xl bg-slate-950 px-3.5 text-xs font-bold text-white"
                  >
                    Xem thống kê
                  </RouterLink>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  </div>
</template>
