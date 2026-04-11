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
</script>

<template>
  <div class="space-y-4 md:space-y-6">
    <section class="overflow-hidden rounded-[28px] bg-gradient-to-br from-[#004edb] via-[#0058bb] to-[#7e9cff] p-5 text-white shadow-[0_12px_30px_rgba(0,78,219,0.2)] md:p-6">
      <div class="grid gap-4 md:grid-cols-[1.2fr_0.8fr] md:items-center">
        <div>
          <span class="inline-flex rounded-full bg-[#fdd404] px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.08em] text-[#594a00]">
            Phòng chơi
          </span>
          <h2 class="mt-4 text-[1.75rem] font-black leading-[1.1] md:text-[2.15rem]">
            Danh sách game-room đang mở
          </h2>
          <p class="mt-3 max-w-[34rem] text-sm leading-6 text-white/88">
            Chọn đúng game trước khi vào phòng. Luồng room sẽ tách riêng theo Wingo, K3 và 5D để dễ vận hành và nối API sau này.
          </p>
        </div>

        <div class="grid gap-3 md:justify-self-end">
          <article class="rounded-[22px] bg-white/14 p-4 backdrop-blur-md">
            <p class="m-0 text-[0.7rem] uppercase tracking-[0.12em] text-white/72">Phòng đang mở</p>
            <strong class="mt-1 block text-[1.4rem] font-black">{{ openRooms.length }}</strong>
            <p class="mt-2 text-[0.74rem] leading-6 text-white/80">Mỗi game chỉ giữ 1 room duy nhất cho giai đoạn hiện tại.</p>
          </article>
          <RouterLink to="/promotion" class="rounded-[22px] bg-white px-4 py-3 text-primary shadow-[0_8px_24px_rgba(255,255,255,0.16)]">
            <p class="m-0 text-[0.7rem] uppercase tracking-[0.12em] text-primary/70">Hoạt động</p>
            <strong class="mt-1 block text-[1rem] font-black">Xem tin / ưu đãi</strong>
          </RouterLink>
        </div>
      </div>
    </section>

    <section class="overflow-x-auto pb-2 no-scrollbar">
      <div class="flex min-w-max gap-2">
        <button
          v-for="category in playCategories"
          :key="category"
          type="button"
          class="rounded-full px-5 py-2.5 text-[0.78rem] font-bold transition-colors"
          :class="category === activeCategory ? 'bg-primary text-white shadow-[0_12px_32px_rgba(0,78,219,0.1)]' : 'bg-surface-container-low text-on-surface-variant'"
          @click="activeCategory = category"
        >
          {{ category }}
        </button>
      </div>
    </section>

    <section class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
      <article
        v-for="room in filteredRooms"
        :key="room.code"
        class="overflow-hidden rounded-[26px] border-b-4 bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)]"
        :style="{ borderBottomColor: room.accent }"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="grid h-[46px] w-[46px] place-items-center rounded-[16px]" :style="{ backgroundColor: `${room.accent}18`, color: room.accent }">
            <span class="material-symbols-outlined">{{ room.symbol }}</span>
          </div>
          <span
            class="rounded-full px-2.5 py-1 text-[0.64rem] font-black uppercase tracking-[0.08em]"
            :class="room.status === 'OPEN' ? 'bg-emerald-500/10 text-emerald-600' : 'bg-slate-200 text-slate-600'"
          >
            {{ room.status === 'OPEN' ? 'Đang mở' : 'Sắp mở' }}
          </span>
        </div>

        <h3 class="m-0 mt-3 text-[1rem] font-black">{{ room.title }}</h3>
        <p class="mt-1.5 text-[0.72rem] leading-6 text-on-surface-variant">{{ room.subtitle }}</p>

        <div class="mt-4 grid grid-cols-2 gap-2 text-[0.72rem]">
          <div class="rounded-[16px] bg-background p-3">
            <span class="block text-on-surface-variant">Min bet</span>
            <strong class="mt-1 block text-on-surface">{{ room.minBet }}</strong>
          </div>
          <div class="rounded-[16px] bg-background p-3">
            <span class="block text-on-surface-variant">Payout</span>
            <strong class="mt-1 block text-on-surface">{{ room.payout }}</strong>
          </div>
          <div class="rounded-[16px] bg-background p-3">
            <span class="block text-on-surface-variant">Người chơi</span>
            <strong class="mt-1 block text-on-surface">{{ room.onlinePlayers.toLocaleString('vi-VN') }}</strong>
          </div>
          <div class="rounded-[16px] bg-background p-3">
            <span class="block text-on-surface-variant">Biến thể</span>
            <strong class="mt-1 block text-on-surface">{{ room.variants.length }}</strong>
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

        <RouterLink
          :to="room.status === 'OPEN' ? `/play/${room.code}` : '/promotion'"
          class="mt-4 flex min-h-12 items-center justify-center rounded-[16px] bg-gradient-to-br from-primary to-primary-container text-[0.82rem] font-black text-white"
        >
          {{ room.status === 'OPEN' ? 'Vào phòng' : 'Sắp mở' }}
        </RouterLink>
      </article>
    </section>
  </div>
</template>
