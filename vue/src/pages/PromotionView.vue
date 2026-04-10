<script setup lang="ts">
import { promotionTabs, promotionTerms } from '../data/mock'

const activeTab = promotionTabs[0]
</script>

<template>
  <div class="space-y-6">
    <section class="relative flex h-48 items-center overflow-hidden rounded-2xl bg-primary md:h-56">
      <div class="absolute inset-0">
        <div class="h-full w-full bg-[radial-gradient(circle_at_top_right,rgba(255,255,255,0.24),transparent_28%),linear-gradient(135deg,rgba(0,78,219,0.95),rgba(0,88,187,0.82),rgba(126,156,255,0.95))]"></div>
      </div>
      <div class="relative z-10 w-2/3 px-6 md:px-8">
        <span class="promo-pill">SIÊU ƯU ĐÃI</span>
        <h2 class="mt-3 text-2xl font-black leading-tight tracking-tight text-white md:text-[2rem]">
          THƯỞNG CHÀO MỪNG 100%
        </h2>
        <p class="mt-2 text-xs font-medium text-primary-container md:text-sm">
          Nạp đầu lên đến 2,000,000 VND. Tham gia ngay để không bỏ lỡ!
        </p>
      </div>
      <div class="absolute inset-0 bg-gradient-to-r from-primary to-transparent opacity-90"></div>
    </section>

    <section class="overflow-x-auto pb-2 no-scrollbar">
      <div class="flex min-w-max gap-2">
        <button
          v-for="tab in promotionTabs"
          :key="tab"
          class="rounded-xl px-6 py-2.5 text-sm font-bold whitespace-nowrap transition-colors"
          :class="tab === activeTab ? 'bg-primary text-white shadow-[0_12px_32px_rgba(0,78,219,0.08)]' : 'bg-surface-container-low text-on-surface-variant hover:bg-surface-container'"
        >
          {{ tab }}
        </button>
      </div>
    </section>

    <section class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <article class="app-card col-span-2 overflow-hidden lg:col-span-2">
        <div class="flex h-full flex-col justify-between p-5">
          <div>
            <div class="mb-4 flex items-start justify-between">
              <span class="rounded-full bg-secondary/20 px-3 py-1 text-[10px] font-bold text-on-secondary-container">
                HOT
              </span>
              <div class="grid h-12 w-12 place-items-center rounded-full bg-primary/10 text-primary">
                <span class="material-symbols-outlined" style="font-variation-settings: 'FILL' 1;">star</span>
              </div>
            </div>
            <h3 class="mb-1 text-lg font-bold text-on-surface">Đặc quyền VIP Membership</h3>
            <p class="mb-4 text-xs text-on-surface-variant">
              Nâng cấp level, nhận quà tặng sinh nhật và hoàn trả không giới hạn mỗi tuần.
            </p>
          </div>
          <button class="rounded-xl bg-gradient-to-br from-primary to-primary-container py-3 text-sm font-black uppercase tracking-wider text-white transition-transform active:scale-95">
            Tham gia ngay
          </button>
        </div>
        <div class="absolute -right-4 -top-4 h-24 w-24 rounded-full bg-tertiary/10 blur-2xl"></div>
      </article>

      <article class="app-card overflow-hidden">
        <div class="relative h-28">
          <div class="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.14),transparent_48%),linear-gradient(180deg,rgba(253,212,4,0.18),rgba(255,255,255,0))]"></div>
        </div>
        <div class="p-4 pt-0">
          <h3 class="mb-1 text-sm font-bold leading-tight text-on-surface">Hoàn trả 1.5% Casino</h3>
          <p class="mb-3 text-[10px] text-on-surface-variant">Tự động kết toán mỗi ngày.</p>
          <button class="w-full rounded-lg bg-secondary py-2 text-[10px] font-bold uppercase text-white transition-transform active:scale-95">
            Nhận ngay
          </button>
        </div>
      </article>

      <article class="app-card overflow-hidden">
        <div class="relative h-28">
          <div class="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.14),transparent_48%),linear-gradient(180deg,rgba(0,78,219,0.12),rgba(255,255,255,0))]"></div>
        </div>
        <div class="p-4 pt-0">
          <h3 class="mb-1 text-sm font-bold leading-tight text-on-surface">Cược thể thao 0đ</h3>
          <p class="mb-3 text-[10px] text-on-surface-variant">Tặng vé cược miễn phí hàng tuần.</p>
          <button class="w-full rounded-lg bg-secondary py-2 text-[10px] font-bold uppercase text-white transition-transform active:scale-95">
            Nhận ngay
          </button>
        </div>
      </article>

      <article class="col-span-2 flex items-center gap-4 rounded-2xl bg-tertiary/10 p-4 lg:col-span-4">
        <div class="grid h-12 w-12 place-items-center rounded-xl bg-white shadow-[0_8px_20px_rgba(0,78,219,0.05)]">
          <span class="material-symbols-outlined text-3xl text-tertiary" style="font-variation-settings: 'FILL' 1;">redeem</span>
        </div>
        <div class="flex-1">
          <h4 class="text-sm font-bold text-on-tertiary-container">Giới thiệu bạn bè</h4>
          <p class="text-xs text-on-tertiary-fixed-variant">Nhận ngay 150,000 VND cho mỗi lượt mời thành công.</p>
        </div>
        <span class="material-symbols-outlined text-tertiary">chevron_right</span>
      </article>
    </section>

    <section class="space-y-4">
      <h2 class="px-1 text-lg font-bold text-on-surface">Điều khoản chung</h2>
      <div class="space-y-3">
        <div
          v-for="term in promotionTerms"
          :key="term"
          class="flex gap-4 rounded-xl bg-surface-container-low p-4"
        >
          <div class="h-12 w-1 rounded-full bg-primary"></div>
          <p class="text-xs leading-relaxed text-on-surface-variant">{{ term }}</p>
        </div>
      </div>
    </section>
  </div>
</template>
