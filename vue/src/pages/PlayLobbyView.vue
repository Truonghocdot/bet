<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { playCategories, playRooms } from '@/data/play'

const activeCategory = ref<string>('Tất cả')

const filteredRooms = computed(() => {
  if (activeCategory.value === 'Tất cả') return playRooms
  return playRooms.filter((room) => room.category === activeCategory.value)
})

const openRooms = computed(() => playRooms.filter((room) => room.status === 'OPEN'))
const featuredRooms = computed(() => playRooms.filter((room) => room.featured))
const comingSoonRooms = computed(() => playRooms.filter((room) => room.status !== 'OPEN'))

const categoryCounts = computed(() =>
  playCategories.map((category) => ({
    category,
    count: category === 'Tất cả'
      ? playRooms.length
      : playRooms.filter((room) => room.category === category).length,
  })),
)

function formatRoomVariants(roomCode: string) {
  const room = playRooms.find((item) => item.code === roomCode)
  return room?.variants.slice(0, 4).map((variant) => variant.durationLabel) ?? []
}
</script>

<template>
  <div class="space-y-4 pb-6 md:space-y-6">
    <section class="overflow-hidden rounded-[28px] bg-gradient-to-br from-[#004edb] via-[#0058bb] to-[#7e9cff] p-5 text-white shadow-[0_12px_30px_rgba(0,78,219,0.2)] md:p-6">
      <div class="grid gap-4 md:grid-cols-[1.35fr_0.95fr] md:items-end">
        <div class="space-y-4">
          <div class="flex flex-wrap items-center gap-2">
            <span class="inline-flex rounded-full bg-[#fdd404] px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.08em] text-[#594a00]">
              Phòng chơi
            </span>
            <span class="rounded-full bg-white/14 px-3 py-1 text-[0.68rem] font-bold text-white/86 backdrop-blur">
              Room cố định, tab time riêng
            </span>
          </div>

          <div class="space-y-2">
            <h2 class="max-w-[20rem] text-[1.72rem] font-black leading-[1.08] md:max-w-[28rem] md:text-[2.25rem]">
              Chọn room game và vào phòng ngay
            </h2>
            <p class="max-w-[33rem] text-sm leading-6 text-white/88 md:text-[0.96rem]">
              Mỗi room là một nhịp thời gian riêng. Card bên dưới cho biết trạng thái, mức cược tối thiểu và các biến thể đang mở.
            </p>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-3 md:justify-self-end">
          <article class="rounded-[22px] bg-white/14 p-4 backdrop-blur-md">
            <p class="m-0 text-[0.68rem] uppercase tracking-[0.12em] text-white/72">Đang mở</p>
            <strong class="mt-1 block text-[1.45rem] font-black">{{ openRooms.length }}</strong>
            <p class="mt-1 text-[0.72rem] leading-5 text-white/80">Room sẵn sàng join.</p>
          </article>

          <article class="rounded-[22px] bg-white/14 p-4 backdrop-blur-md">
            <p class="m-0 text-[0.68rem] uppercase tracking-[0.12em] text-white/72">Sắp mở</p>
            <strong class="mt-1 block text-[1.45rem] font-black">{{ comingSoonRooms.length }}</strong>
            <p class="mt-1 text-[0.72rem] leading-5 text-white/80">Room chuẩn bị lên sóng.</p>
          </article>

          <RouterLink to="/promotion" class="rounded-[22px] bg-white px-4 py-4 text-primary shadow-[0_8px_24px_rgba(255,255,255,0.16)]">
            <p class="m-0 text-[0.68rem] uppercase tracking-[0.12em] text-primary/70">Nổi bật</p>
            <strong class="mt-1 block text-[1rem] font-black leading-5">{{ featuredRooms.length }} room chính</strong>
            <span class="mt-2 inline-flex items-center gap-1 text-[0.72rem] font-bold text-primary">
              Xem khuyến mãi <span class="material-symbols-outlined text-[1rem]">chevron_right</span>
            </span>
          </RouterLink>
        </div>
      </div>
    </section>

    <section class="overflow-x-auto pb-1 no-scrollbar">
      <div class="flex min-w-max gap-2">
        <button
          v-for="item in categoryCounts"
          :key="item.category"
          type="button"
          class="flex items-center gap-2 rounded-full px-4 py-2.5 text-[0.78rem] font-bold transition-all"
          :class="item.category === activeCategory ? 'bg-primary text-white shadow-[0_12px_32px_rgba(0,78,219,0.12)]' : 'bg-surface-container-low text-on-surface-variant'"
          @click="activeCategory = item.category"
        >
          <span>{{ item.category }}</span>
          <span
            class="rounded-full px-2 py-0.5 text-[0.65rem]"
            :class="item.category === activeCategory ? 'bg-white/18 text-white' : 'bg-white text-on-surface-variant'"
          >
            {{ item.count }}
          </span>
        </button>
      </div>
    </section>

    <section class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
      <article
        v-for="room in filteredRooms"
        :key="room.code"
        class="group relative overflow-hidden rounded-[26px] bg-white p-4 shadow-[0_10px_24px_rgba(0,78,219,0.07)] transition-transform duration-200 hover:-translate-y-0.5"
      >
        <div class="absolute inset-x-0 top-0 h-1.5" :style="{ background: `linear-gradient(90deg, ${room.accent}, ${room.accent}55)` }"></div>
        <div class="flex items-start justify-between gap-3">
          <div class="flex items-start gap-3">
            <div class="grid h-[48px] w-[48px] place-items-center rounded-[16px]" :style="{ backgroundColor: `${room.accent}16`, color: room.accent }">
              <span class="material-symbols-outlined">{{ room.symbol }}</span>
            </div>
            <div>
              <div class="flex flex-wrap items-center gap-2">
                <h3 class="m-0 text-[1rem] font-black">{{ room.title }}</h3>
                <span
                  class="rounded-full px-2.5 py-1 text-[0.64rem] font-black uppercase tracking-[0.08em]"
                  :class="room.status === 'OPEN' ? 'bg-emerald-500/10 text-emerald-600' : 'bg-amber-500/10 text-amber-700'"
                >
                  {{ room.status === 'OPEN' ? 'Đang mở' : 'Sắp mở' }}
                </span>
              </div>
              <p class="mt-1 text-[0.72rem] text-on-surface-variant">
                {{ room.subtitle }}
              </p>
            </div>
          </div>
          <span class="rounded-full bg-surface-container-low px-2.5 py-1 text-[0.64rem] font-black uppercase tracking-[0.08em] text-on-surface-variant">
            {{ room.category }}
          </span>
        </div>

        <div class="mt-4 grid grid-cols-2 gap-2 text-[0.72rem]">
          <div class="rounded-[16px] bg-background p-3">
            <span class="block text-[0.68rem] text-on-surface-variant">Mức tối thiểu</span>
            <strong class="mt-1 block text-on-surface">{{ room.minBet }}</strong>
          </div>
          <div class="rounded-[16px] bg-background p-3">
            <span class="block text-[0.68rem] text-on-surface-variant">Tỷ lệ trả</span>
            <strong class="mt-1 block text-on-surface">{{ room.payout }}</strong>
          </div>
          <div class="rounded-[16px] bg-background p-3">
            <span class="block text-[0.68rem] text-on-surface-variant">Người chơi</span>
            <strong class="mt-1 block text-on-surface">{{ room.onlinePlayers.toLocaleString('vi-VN') }}</strong>
          </div>
          <div class="rounded-[16px] bg-background p-3">
            <span class="block text-[0.68rem] text-on-surface-variant">Nhịp room</span>
            <strong class="mt-1 block text-on-surface">{{ room.variants.length }} time</strong>
          </div>
        </div>

        <div class="mt-4 flex flex-wrap gap-2">
          <span
            v-for="variant in room.variants"
            :key="variant.code"
            class="rounded-full bg-surface-container-low px-3 py-1 text-[0.68rem] font-bold text-on-surface-variant"
          >
            {{ variant.label }}
          </span>
        </div>

        <div class="mt-4 flex items-center gap-2">
          <div class="flex-1 rounded-[16px] bg-background px-3 py-2.5 text-[0.72rem] text-on-surface-variant">
            <span class="block text-[0.65rem] uppercase tracking-[0.08em]">Nhịp nhanh</span>
            <strong class="mt-1 block text-on-surface">
              {{ formatRoomVariants(room.code).join(' • ') }}
            </strong>
          </div>

          <RouterLink
            :to="room.status === 'OPEN' ? `/play/${room.code}` : '/promotion'"
            class="inline-flex min-h-12 items-center justify-center rounded-[16px] bg-gradient-to-br from-primary to-primary-container px-4 text-[0.82rem] font-black text-white transition-transform active:scale-95"
          >
            {{ room.status === 'OPEN' ? 'Vào phòng' : 'Sắp mở' }}
          </RouterLink>
        </div>
      </article>
    </section>
  </div>
</template>
