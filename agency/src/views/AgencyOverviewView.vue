<script setup lang="ts">
import { onMounted } from 'vue'

import { formatAgencyDateTime } from '@/shared/lib/date'
import { useAgencyAuthStore } from '@/stores/auth'
import { useAgencyPlayersStore } from '@/stores/players'

const auth = useAgencyAuthStore()
const players = useAgencyPlayersStore()

onMounted(() => {
  void players.refreshDashboard()
})
</script>

<template>
  <div class="mx-auto max-w-7xl space-y-6">
    <section class="rounded-[34px] border border-white/60 bg-[rgba(255,255,255,0.78)] p-6 shadow-[0_30px_80px_rgba(15,23,42,0.12)] backdrop-blur md:p-8">
      <div class="flex flex-col gap-6 xl:flex-row xl:items-end xl:justify-between">
        <div>
          <p class="m-0 text-xs font-bold uppercase tracking-[0.22em] text-rose-600">Bảng điều khiển agency</p>
          <h1 class="mt-3 text-3xl font-bold text-slate-950 md:text-4xl">{{ auth.user?.name || 'Agency' }}</h1>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-slate-600">
            Theo dõi các tài khoản người chơi đang thuộc tuyến quản lý của bạn. Dữ liệu hiện lấy từ endpoint affiliate hiện tại.
          </p>
          <div class="mt-4 flex flex-wrap gap-2 text-xs font-bold uppercase tracking-[0.1em] text-slate-500">
            <span class="rounded-full bg-white px-3 py-1.5">Mã ref {{ auth.affiliateProfile?.ref_code || '—' }}</span>
            <span class="rounded-full bg-white px-3 py-1.5">Vai trò {{ auth.user?.role ?? '—' }}</span>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <button
            type="button"
            class="rounded-xl border border-black/8 bg-white px-4 py-2.5 text-sm font-bold text-slate-700"
            @click="void players.refreshDashboard()"
          >
            Tải lại
          </button>
        </div>
      </div>
    </section>

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <article class="rounded-[28px] bg-slate-950 p-5 text-white">
        <p class="text-xs uppercase tracking-[0.18em] text-white/60">Người chơi quản lý</p>
        <strong class="mt-3 block text-4xl">{{ players.headlinePlayers }}</strong>
      </article>
      <article class="rounded-[28px] bg-white p-5 shadow-[0_18px_50px_rgba(15,23,42,0.08)]">
        <p class="text-xs uppercase tracking-[0.18em] text-slate-400">Đã nạp đầu</p>
        <strong class="mt-3 block text-4xl text-slate-950">{{ players.qualifiedPlayers }}</strong>
      </article>
      <article class="rounded-[28px] bg-white p-5 shadow-[0_18px_50px_rgba(15,23,42,0.08)]">
        <p class="text-xs uppercase tracking-[0.18em] text-slate-400">Chờ nạp đầu</p>
        <strong class="mt-3 block text-4xl text-slate-950">{{ players.waitingPlayers }}</strong>
      </article>
      <article class="rounded-[28px] bg-gradient-to-br from-amber-200 to-rose-200 p-5">
        <p class="text-xs uppercase tracking-[0.18em] text-slate-700">Tổng first deposit</p>
        <strong class="mt-3 block text-4xl text-slate-950">{{ players.totalFirstDeposit.toLocaleString('vi-VN') }}</strong>
      </article>
    </section>

    <section class="grid gap-5 xl:grid-cols-[1.1fr_0.9fr]">
      <div class="space-y-5">
        <section class="rounded-[30px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-5 shadow-[0_30px_80px_rgba(15,23,42,0.12)]">
          <div class="flex items-center justify-between gap-3">
            <div>
              <h2 class="m-0 text-2xl font-bold text-slate-950">Hiệu suất tuyến agency</h2>
              <p class="mt-2 text-sm text-slate-500">Nhìn nhanh trạng thái chuyển đổi của toàn bộ người chơi đang quản lý.</p>
            </div>
            <div class="rounded-[20px] bg-slate-950 px-4 py-3 text-right text-white">
              <p class="m-0 text-[0.68rem] uppercase tracking-[0.14em] text-white/60">Tỷ lệ chuyển đổi</p>
              <strong class="mt-1 block text-3xl">{{ players.conversionRate }}%</strong>
            </div>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-3">
            <article class="rounded-[24px] bg-slate-50 p-4">
              <p class="m-0 text-xs uppercase tracking-[0.16em] text-slate-400">Mời thành công</p>
              <strong class="mt-2 block text-3xl text-slate-950">{{ players.invitedUsersCount }}</strong>
            </article>
            <article class="rounded-[24px] bg-slate-50 p-4">
              <p class="m-0 text-xs uppercase tracking-[0.16em] text-slate-400">First deposit TB</p>
              <strong class="mt-2 block text-3xl text-slate-950">{{ players.averageFirstDeposit.toLocaleString('vi-VN') }}</strong>
            </article>
            <article class="rounded-[24px] bg-slate-50 p-4">
              <p class="m-0 text-xs uppercase tracking-[0.16em] text-slate-400">Không hợp lệ</p>
              <strong class="mt-2 block text-3xl text-slate-950">{{ players.invalidPlayers }}</strong>
            </article>
          </div>

          <div class="mt-6 space-y-3">
            <div class="flex items-center justify-between text-sm font-semibold text-slate-500">
              <span>Chờ nạp đầu</span>
              <span>{{ players.waitingPlayers }}</span>
            </div>
            <div class="h-3 overflow-hidden rounded-full bg-slate-100">
              <div
                class="h-full rounded-full bg-amber-400"
                :style="{ width: `${players.headlinePlayers ? (players.waitingPlayers / players.headlinePlayers) * 100 : 0}%` }"
              />
            </div>

            <div class="flex items-center justify-between text-sm font-semibold text-slate-500">
              <span>Đã nạp đầu</span>
              <span>{{ players.qualifiedPlayers }}</span>
            </div>
            <div class="h-3 overflow-hidden rounded-full bg-slate-100">
              <div
                class="h-full rounded-full bg-emerald-500"
                :style="{ width: `${players.headlinePlayers ? (players.qualifiedPlayers / players.headlinePlayers) * 100 : 0}%` }"
              />
            </div>
          </div>
        </section>

        <section class="rounded-[30px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-5 shadow-[0_30px_80px_rgba(15,23,42,0.12)]">
          <h2 class="m-0 text-2xl font-bold text-slate-950">Người chơi mới nhất</h2>
          <div v-if="players.recentPlayers.length === 0" class="mt-4 text-sm font-semibold text-slate-500">
            Chưa có dữ liệu người chơi mới.
          </div>
          <div v-else class="mt-4 space-y-3">
            <article
              v-for="item in players.recentPlayers"
              :key="`recent-${item.user_id}`"
              class="rounded-[22px] border border-black/6 bg-white px-4 py-3"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <strong class="block text-slate-950">{{ item.name || `User #${item.user_id}` }}</strong>
                  <span class="mt-1 block text-xs text-slate-500">{{ item.phone || 'Chưa cập nhật số điện thoại' }}</span>
                </div>
                <span class="text-xs font-semibold text-slate-400">{{ formatAgencyDateTime(item.created_at) }}</span>
              </div>
            </article>
          </div>
        </section>
      </div>

      <section class="rounded-[30px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-5 shadow-[0_30px_80px_rgba(15,23,42,0.12)]">
        <h2 class="m-0 text-2xl font-bold text-slate-950">Danh sách cần chú ý</h2>
        <p class="mt-2 text-sm text-slate-500">Ưu tiên xử lý các tài khoản chưa nạp đầu hoặc có trạng thái không hợp lệ.</p>

        <div v-if="players.priorityPlayers.length === 0" class="mt-5 rounded-[22px] border border-dashed border-black/10 px-4 py-10 text-center text-sm font-semibold text-slate-500">
          Chưa có người chơi cần chú ý.
        </div>

        <div v-else class="mt-5 space-y-3">
          <article
            v-for="item in players.priorityPlayers"
            :key="`priority-${item.user_id}`"
            class="rounded-[22px] border border-black/6 bg-white px-4 py-4"
          >
            <div class="flex items-start justify-between gap-3">
              <div>
                <strong class="block text-slate-950">{{ item.name || `User #${item.user_id}` }}</strong>
                <span class="mt-1 block text-xs text-slate-500">{{ item.phone || 'Chưa cập nhật số điện thoại' }}</span>
              </div>
              <span
                class="rounded-full px-3 py-1 text-xs font-bold uppercase tracking-[0.08em]"
                :class="item.referral_status === 1 ? 'bg-amber-100 text-amber-700' : 'bg-slate-100 text-slate-600'"
              >
                {{ players.statusLabel(item.referral_status) }}
              </span>
            </div>
            <div class="mt-3 flex items-center justify-between text-xs text-slate-500">
              <span>Tạo lúc {{ formatAgencyDateTime(item.created_at) }}</span>
              <span>First deposit {{ Number(item.first_deposit_amount || 0).toLocaleString('vi-VN') }}</span>
            </div>
          </article>
        </div>
      </section>
    </section>
  </div>
</template>
