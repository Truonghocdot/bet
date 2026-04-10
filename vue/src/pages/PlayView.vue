<script setup lang="ts">
import { computed } from 'vue'
import { gameHistoryRows, gameVariants } from '../data/mock'

const currentBalance = '51,558,000đ'
const currentPeriod = '20240520342'
const recentDigits = ['0', '0', ':', '2', '8']

const routeLabel = computed(() => 'Win Go 30s')

const colorMap: Record<string, string> = {
  Đỏ: '#e74c3c',
  Xanh: '#0058bb',
  Tím: '#9b59b6',
}
</script>

<template>
  <div class="space-y-3.5 md:space-y-5">
    <section class="rounded-[22px] bg-gradient-to-br from-[#0058bb] to-[#004ca4] p-[18px] text-white shadow-[0_8px_20px_rgba(0,78,219,0.18)] md:p-5">
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="m-0 text-[0.76rem] uppercase tracking-[0.18em] text-white/80">Số dư hiện tại</p>
          <h2 class="m-0 mt-2 text-[1.7rem] font-black md:text-[1.9rem]">{{ currentBalance }}</h2>
        </div>
        <button class="grid h-[42px] w-[42px] place-items-center rounded-full bg-white/12 text-white transition-transform active:scale-95">
          <span class="material-symbols-outlined">refresh</span>
        </button>
      </div>

      <div class="mt-3.5 grid grid-cols-2 gap-3">
        <button class="min-h-[52px] rounded-[16px] bg-white font-extrabold text-primary transition-transform active:scale-95">
          Nạp tiền
        </button>
        <button class="min-h-[52px] rounded-[16px] border border-white/20 bg-white/12 font-extrabold text-white transition-transform active:scale-95">
          Rút tiền
        </button>
      </div>
    </section>

    <div class="flex items-center gap-2 rounded-full bg-surface-container-low px-3.5 py-3">
      <span class="material-symbols-outlined text-primary">campaign</span>
      <p class="m-0 truncate text-[0.76rem] font-bold">Chúc mừng người chơi ***123 vừa thắng 2,500,000đ tại Win Go!</p>
    </div>

    <section class="grid grid-cols-4 gap-2 md:gap-2.5">
      <button
        v-for="variant in gameVariants"
        :key="variant.key"
        class="grid min-h-[66px] place-items-center gap-1 rounded-[16px] bg-white font-extrabold text-on-surface-variant shadow-[0_6px_16px_rgba(0,78,219,0.04)] transition-transform active:scale-95"
        :class="variant.active ? 'bg-primary text-white' : ''"
      >
        <span class="material-symbols-outlined text-[1.05rem]">timer</span>
        <span>{{ variant.label }}</span>
      </button>
    </section>

    <section class="grid gap-3 md:grid-cols-[1.15fr_0.85fr]">
      <article class="rounded-[22px] bg-white p-3.5 shadow-[0_8px_20px_rgba(0,78,219,0.05)] md:p-4">
        <p class="m-0 text-[0.76rem] font-bold text-on-surface-variant">Số kỳ: <strong class="text-primary">{{ currentPeriod }}</strong></p>
        <div class="mt-2.5 flex items-center gap-1.5 overflow-x-auto pb-1 no-scrollbar">
          <span
            v-for="digit in recentDigits"
            :key="`${digit}-${currentPeriod}`"
            class="grid h-[52px] w-[42px] place-items-center rounded-[12px] border border-primary/20 bg-primary/10 text-[1.5rem] font-black text-primary"
          >
            {{ digit }}
          </span>
        </div>
      </article>

      <article class="rounded-[22px] bg-white p-3.5 shadow-[0_8px_20px_rgba(0,78,219,0.05)] md:p-4">
        <p class="m-0 text-[0.76rem] font-bold text-on-surface-variant">Kết quả gần đây</p>
        <div class="mt-2.5 flex flex-wrap gap-2">
          <span class="grid h-7 w-7 place-items-center rounded-full bg-[#e74c3c] text-[0.68rem] font-extrabold text-white">3</span>
          <span class="grid h-7 w-7 place-items-center rounded-full bg-[#0058bb] text-[0.68rem] font-extrabold text-white">7</span>
          <span class="grid h-7 w-7 place-items-center rounded-full bg-[#6d5a00] text-[0.68rem] font-extrabold text-white">5</span>
          <span class="grid h-7 w-7 place-items-center rounded-full bg-[#e74c3c] text-[0.68rem] font-extrabold text-white">1</span>
          <span class="grid h-7 w-7 place-items-center rounded-full bg-[#0058bb] text-[0.68rem] font-extrabold text-white">8</span>
        </div>
      </article>
    </section>

    <section class="rounded-[22px] bg-white p-4 shadow-[0_8px_20px_rgba(0,78,219,0.05)] md:p-[18px]">
      <div class="grid grid-cols-3 gap-2.5">
        <button class="min-h-12 rounded-[16px] bg-[#2ecc71] font-black text-white transition-transform active:scale-95">
          Xanh
        </button>
        <button class="min-h-12 rounded-[16px] bg-[#9b59b6] font-black text-white transition-transform active:scale-95">
          Tím
        </button>
        <button class="min-h-12 rounded-[16px] bg-[#e74c3c] font-black text-white transition-transform active:scale-95">
          Đỏ
        </button>
      </div>

      <div class="mt-3.5 grid grid-cols-5 gap-2.5 md:gap-3">
        <button
          v-for="n in 10"
          :key="n"
          class="aspect-square rounded-full border-2 border-primary/30 bg-white text-[1.02rem] font-black text-primary transition-transform active:scale-95"
        >
          {{ n - 1 }}
        </button>
      </div>

      <div class="mt-3.5 flex flex-wrap items-center gap-2">
        <button class="min-h-9 rounded-[12px] bg-surface-container-low px-4 text-[0.74rem] font-extrabold">Ngẫu nhiên</button>
        <div class="flex flex-1 flex-wrap gap-1.5">
          <button
            v-for="value in ['X1', 'X5', 'X10', 'X20', 'X50']"
            :key="value"
            class="min-h-[34px] min-w-10 rounded-[12px] bg-surface-container-low px-3 text-[0.7rem] font-black"
            :class="value === 'X10' ? 'bg-primary text-white' : 'text-on-surface-variant'"
          >
            {{ value }}
          </button>
        </div>
      </div>

      <div class="mt-3.5 grid grid-cols-2 gap-3">
        <button class="min-h-14 rounded-[16px] bg-[#fdd404] text-[1.02rem] font-black text-[#594a00] transition-transform active:scale-95">
          LỚN
        </button>
        <button class="min-h-14 rounded-[16px] bg-[#7e9cff] text-[1.02rem] font-black text-white transition-transform active:scale-95">
          NHỎ
        </button>
      </div>
    </section>

    <section class="overflow-hidden rounded-[22px] bg-white shadow-[0_8px_20px_rgba(0,78,219,0.05)]">
      <div class="grid grid-cols-3 border-b border-slate-200/60">
        <button class="min-h-[52px] border-b-2 border-primary text-primary font-extrabold">Lịch sử trò chơi</button>
        <button class="min-h-[52px] font-extrabold text-on-surface-variant">Biểu đồ</button>
        <button class="min-h-[52px] font-extrabold text-on-surface-variant">Lịch sử của tôi</button>
      </div>

      <div class="py-2.5">
        <div class="grid grid-cols-[1fr_auto_1fr_auto] items-center gap-2.5 px-3.5 pb-2 text-[0.64rem] font-black uppercase tracking-[0.08em] text-on-surface-variant">
          <span>Kỳ xổ</span>
          <span>Số</span>
          <span>Lớn nhỏ</span>
          <span>Màu sắc</span>
        </div>

        <div
          v-for="row in gameHistoryRows"
          :key="row.period"
          class="grid min-h-[50px] grid-cols-[1fr_auto_1fr_auto] items-center gap-2.5 border-t border-slate-200/50 px-3.5 text-[0.82rem] font-bold"
        >
          <span>{{ row.period }}</span>
          <span class="grid h-[22px] w-[22px] place-items-center rounded-full bg-[#e74c3c] text-[0.68rem] text-white">
            {{ row.number }}
          </span>
          <span>{{ row.size }}</span>
          <span class="h-3 w-3 rounded-full" :style="{ backgroundColor: colorMap[row.color] }"></span>
        </div>
      </div>
    </section>

    <p class="m-0 text-center text-[0.76rem] font-bold text-on-surface-variant">{{ routeLabel }}</p>
  </div>
</template>
