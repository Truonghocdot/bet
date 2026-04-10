<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import { gameCategories, gameRooms, type GameRoom } from '@/data/site'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const auth = useAuthStore()

const activeCategory = ref('Tất cả')

const selectedGameCode = computed(() => String(route.params.game || 'wingo'))
const selectedGame = computed<GameRoom>(() => gameRooms.find((game) => game.code === selectedGameCode.value) ?? gameRooms[0]!)
const filteredRooms = computed(() => {
  if (activeCategory.value === 'Tất cả') return gameRooms
  return gameRooms.filter((game) => game.category === activeCategory.value)
})

const roomSummary = computed(() => [
  { label: 'Người chơi', value: selectedGame.value.onlinePlayers.toLocaleString('vi-VN') },
  { label: 'Kỳ', value: selectedGame.value.roundTime },
  { label: 'Min bet', value: selectedGame.value.minBet },
  { label: 'Tỷ lệ', value: selectedGame.value.payout },
])
</script>

<template>
  <div class="space-y-4 md:space-y-6">
    <section class="overflow-hidden rounded-[28px] bg-gradient-to-br from-[#004edb] via-[#0058bb] to-[#7e9cff] p-5 text-white shadow-[0_12px_30px_rgba(0,78,219,0.2)] md:p-6">
      <div class="grid gap-4 md:grid-cols-[1fr_auto] md:items-center">
        <div>
          <span class="inline-flex rounded-full bg-[#fdd404] px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.08em] text-[#594a00]">
            Phòng chơi
          </span>
          <h2 class="mt-4 text-[1.65rem] font-black md:text-[2rem]">{{ selectedGame.title }}</h2>
          <p class="mt-2 max-w-[34rem] text-sm leading-6 text-white/88">
            Danh sách game-room hiện tại. Chọn một phòng để vào luồng chơi thật khi backend realtime hoàn thiện.
          </p>
        </div>

        <RouterLink to="/notifications" class="rounded-[22px] bg-white/14 px-4 py-3 backdrop-blur-md">
          <p class="m-0 text-[0.7rem] uppercase tracking-[0.12em] text-white/72">Phiên bản phòng</p>
          <strong class="mt-1 block text-[1rem] font-black">{{ selectedGame.roundTime }}</strong>
        </RouterLink>
      </div>
    </section>

    <section class="grid grid-cols-2 gap-2.5 md:grid-cols-4">
      <article v-for="item in roomSummary" :key="item.label" class="rounded-[20px] bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)]">
        <p class="m-0 text-[0.7rem] uppercase tracking-[0.08em] text-on-surface-variant">{{ item.label }}</p>
        <strong class="mt-2 block text-[1.05rem] font-black text-on-surface">{{ item.value }}</strong>
      </article>
    </section>

    <section class="overflow-x-auto pb-2 no-scrollbar">
      <div class="flex min-w-max gap-2">
        <button
          v-for="category in gameCategories"
          :key="category"
          class="rounded-full px-5 py-2.5 text-[0.78rem] font-bold transition-colors"
          :class="category === activeCategory ? 'bg-primary text-white shadow-[0_12px_32px_rgba(0,78,219,0.1)]' : 'bg-surface-container-low text-on-surface-variant'"
          type="button"
          @click="activeCategory = category"
        >
          {{ category }}
        </button>
      </div>
    </section>

    <section class="grid gap-3 lg:grid-cols-[1.3fr_0.7fr]">
      <div class="grid gap-3 md:grid-cols-2">
        <article
          v-for="game in filteredRooms"
          :key="game.code"
          class="overflow-hidden rounded-[24px] border-b-4 bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)] transition-transform active:scale-[0.99]"
          :class="game.code === selectedGame.code ? 'ring-2 ring-primary/20' : ''"
          :style="{ borderBottomColor: game.accent }"
        >
          <div class="flex items-start justify-between gap-3">
            <div
              class="grid h-[46px] w-[46px] place-items-center rounded-[16px]"
              :style="{ backgroundColor: `${game.accent}18`, color: game.accent }"
            >
              <span class="material-symbols-outlined">{{ game.symbol }}</span>
            </div>
            <span
              class="rounded-full px-2.5 py-1 text-[0.64rem] font-black uppercase tracking-[0.08em]"
              :class="game.status === 'OPEN' ? 'bg-emerald-500/10 text-emerald-600' : 'bg-slate-200 text-slate-600'"
            >
              {{ game.status === 'OPEN' ? 'Đang mở' : 'Sắp mở' }}
            </span>
          </div>

          <h3 class="m-0 mt-3 text-[1rem] font-black">{{ game.title }}</h3>
          <p class="mt-1.5 text-[0.72rem] leading-6 text-on-surface-variant">{{ game.subtitle }}</p>

          <div class="mt-4 grid grid-cols-2 gap-2 text-[0.72rem]">
            <div class="rounded-[16px] bg-background p-3">
              <span class="block text-on-surface-variant">Kỳ</span>
              <strong class="mt-1 block text-on-surface">{{ game.roundTime }}</strong>
            </div>
            <div class="rounded-[16px] bg-background p-3">
              <span class="block text-on-surface-variant">Người chơi</span>
              <strong class="mt-1 block text-on-surface">{{ game.onlinePlayers.toLocaleString('vi-VN') }}</strong>
            </div>
            <div class="rounded-[16px] bg-background p-3">
              <span class="block text-on-surface-variant">Min</span>
              <strong class="mt-1 block text-on-surface">{{ game.minBet }}</strong>
            </div>
            <div class="rounded-[16px] bg-background p-3">
              <span class="block text-on-surface-variant">Payout</span>
              <strong class="mt-1 block text-on-surface">{{ game.payout }}</strong>
            </div>
          </div>

          <RouterLink
            :to="`/play/${game.code}`"
            class="mt-4 flex min-h-12 items-center justify-center rounded-[16px] bg-gradient-to-br from-primary to-primary-container text-[0.82rem] font-black text-white"
          >
            Vào phòng
          </RouterLink>
        </article>
      </div>

      <aside class="space-y-3">
        <article class="rounded-[24px] bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)]">
          <div class="flex items-center justify-between gap-2">
            <div>
              <h3 class="m-0 text-[1rem] font-black">Phòng hiện tại</h3>
              <p class="m-0 mt-1 text-[0.76rem] text-on-surface-variant">Game được route vào.</p>
            </div>
            <span class="rounded-full bg-primary/10 px-2.5 py-1 text-[0.64rem] font-black text-primary">LIVE</span>
          </div>

          <div class="mt-4 rounded-[20px] bg-background p-4">
            <div class="flex items-center gap-3">
              <div
                class="grid h-12 w-12 place-items-center rounded-[16px]"
                :style="{ backgroundColor: `${selectedGame.accent}18`, color: selectedGame.accent }"
              >
                <span class="material-symbols-outlined">{{ selectedGame.symbol }}</span>
              </div>
              <div>
                <strong class="block text-[1rem]">{{ selectedGame.title }}</strong>
                <span class="mt-1 block text-[0.72rem] text-on-surface-variant">{{ selectedGame.subtitle }}</span>
              </div>
            </div>

            <div class="mt-4 grid grid-cols-2 gap-2 text-[0.72rem]">
              <div class="rounded-[16px] bg-white p-3">
                <span class="block text-on-surface-variant">Jackpot</span>
                <strong class="mt-1 block text-on-surface">{{ selectedGame.jackpot }}</strong>
              </div>
              <div class="rounded-[16px] bg-white p-3">
                <span class="block text-on-surface-variant">Status</span>
                <strong class="mt-1 block text-on-surface">{{ selectedGame.status === 'OPEN' ? 'Sẵn sàng' : 'Sắp mở' }}</strong>
              </div>
            </div>
          </div>
        </article>

        <article class="rounded-[24px] bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)]">
          <h3 class="m-0 text-[1rem] font-black">Lưu ý vận hành</h3>
          <ul class="mt-3 space-y-2 text-[0.78rem] leading-6 text-on-surface-variant">
            <li>Chỉ giữ 1 room cho mỗi game ở giai đoạn hiện tại.</li>
            <li>Người chơi vào phòng sau khi có xác thực và session hợp lệ.</li>
            <li>Khi realtime hoàn thiện, trạng thái kỳ và bet sẽ được đẩy ngay trong phòng này.</li>
          </ul>
        </article>

        <article class="rounded-[24px] bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)]">
          <h3 class="m-0 text-[1rem] font-black">Tài khoản</h3>
          <p class="mt-2 text-[0.78rem] leading-6 text-on-surface-variant">
            {{ auth.user?.name || 'Khách' }} có thể vào phòng và đặt lệnh sau khi kết nối phiên thật.
          </p>
        </article>
      </aside>
    </section>
  </div>
</template>
