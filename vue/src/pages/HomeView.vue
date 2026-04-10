<script setup lang="ts">
import { featuredGames, homeWinningFeed, quickCategories } from '../data/mock'
</script>

<template>
  <div class="space-y-4 md:space-y-5">
    <section
      class="relative flex min-h-48 items-center overflow-hidden rounded-2xl bg-primary shadow-[0_8px_20px_rgba(0,78,219,0.18)] md:min-h-60"
    >
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(255,255,255,0.24),transparent_28%),linear-gradient(135deg,#004edb_0%,#0058bb_55%,#7e9cff_100%)]"></div>
      <div class="relative z-10 px-5 py-5 text-white md:px-7">
        <span class="inline-flex rounded-full bg-[#fdd404] px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.06em] text-[#594a00]">
          SIÊU KHUYẾN MÃI
        </span>
        <h2 class="mt-4 max-w-[240px] text-[1.5rem] font-extrabold leading-[1.15] md:max-w-[360px] md:text-[1.8rem]">
          Thưởng Nạp Lần Đầu Đến 100% Giá Trị
        </h2>
        <p class="mt-2 max-w-[240px] text-sm opacity-90 md:max-w-[360px] md:text-[0.98rem]">
          Tham gia ngay để nhận ưu đãi VIP dành riêng cho bạn
        </p>
      </div>
    </section>

    <section class="grid grid-cols-5 gap-2.5 md:gap-3">
      <button
        v-for="item in quickCategories"
        :key="item.title"
        class="grid min-h-20 place-items-center gap-2 rounded-[18px] bg-white/90 px-1.5 py-2.5 text-on-surface shadow-[0_6px_18px_rgba(0,78,219,0.06)]"
      >
        <span class="material-symbols-outlined text-[1.45rem] text-primary">{{ item.symbol }}</span>
        <span class="text-[0.66rem] font-extrabold">{{ item.title }}</span>
      </button>
    </section>

    <section class="flex items-end justify-between">
      <h3 class="m-0 text-[1.05rem] font-extrabold md:text-[1.15rem]">Trò chơi nổi bật</h3>
      <button class="inline-flex items-center gap-0.5 text-[0.78rem] font-extrabold text-primary">
        Xem tất cả <span class="material-symbols-outlined text-[1rem]">chevron_right</span>
      </button>
    </section>

    <section class="grid grid-cols-2 gap-3.5 md:grid-cols-4 md:gap-4">
      <article
        v-for="game in featuredGames"
        :key="game.code"
        class="overflow-hidden rounded-[18px] border-b-4 bg-white/95 p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)]"
        :style="{ borderBottomColor: game.accent }"
      >
        <div
          class="mb-2.5 grid h-[42px] w-[42px] place-items-center rounded-[14px]"
          :style="{ backgroundColor: `${game.accent}22`, color: game.accent }"
        >
          <span class="material-symbols-outlined">{{ game.symbol }}</span>
        </div>
        <h4 class="m-0 text-[0.9rem] font-extrabold">{{ game.title }}</h4>
        <p class="mt-1.5 text-[0.68rem] leading-[1.35] text-on-surface-variant">{{ game.subtitle }}</p>
        <span
          v-if="game.enabled === false"
          class="mt-2.5 inline-flex rounded-full bg-primary/10 px-2 py-1 text-[0.62rem] font-extrabold text-primary"
        >
          Coming soon
        </span>
      </article>
    </section>

    <section class="rounded-[22px] bg-white/90 p-4 shadow-[0_8px_20px_rgba(0,78,219,0.06)] md:p-5">
      <div class="mb-3.5 flex items-center gap-2.5">
        <span class="h-2 w-2 rounded-full bg-[#b71211] shadow-[0_0_0_6px_rgba(183,18,17,0.12)]"></span>
        <h3 class="m-0 text-[0.8rem] font-black uppercase tracking-[0.04em] text-[#b71211]">
          Thông tin trúng thưởng
        </h3>
      </div>

      <div class="grid gap-3 md:grid-cols-2">
        <article
          v-for="item in homeWinningFeed"
          :key="item.displayName"
          class="grid grid-cols-[auto_1fr_auto] items-center gap-3 rounded-[16px] bg-background p-3"
        >
          <div class="grid h-[38px] w-[38px] place-items-center rounded-full bg-gradient-to-br from-primary to-primary-container font-extrabold text-white">
            {{ item.displayName.slice(0, 1) }}
          </div>
          <div>
            <strong class="block text-[0.78rem]">{{ item.displayName }}</strong>
            <p class="m-0 mt-0.5 text-[0.64rem] text-on-surface-variant">{{ item.description }}</p>
          </div>
          <div class="text-right">
            <strong class="block text-[0.78rem] text-[#b71211]">{{ item.amount }}</strong>
            <span class="mt-0.5 block text-[0.64rem] text-on-surface-variant">{{ item.tag }}</span>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>
