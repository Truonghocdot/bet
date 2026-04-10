<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import { useAuthStore } from '@/stores/auth'
import { formatViDateTime } from '@/shared/lib/date'
import { gameRooms, getUnreadCount, homeActivities, homeMetrics, newsArticles, quickActions } from '@/data/site'

const auth = useAuthStore()

const greetingName = computed(() => auth.user?.name || 'Người chơi')
const featuredRooms = computed(() => gameRooms.filter((game) => game.featured))
const featuredNews = computed(() => newsArticles.filter((article) => article.featured).slice(0, 3))
</script>

<template>
  <div class="space-y-4 md:space-y-6">
    <section
      class="relative overflow-hidden rounded-[28px] bg-gradient-to-br from-[#004edb] via-[#0058bb] to-[#7e9cff] p-5 text-white shadow-[0_12px_30px_rgba(0,78,219,0.2)] md:p-7"
    >
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(255,255,255,0.22),transparent_26%),radial-gradient(circle_at_bottom_left,rgba(255,255,255,0.08),transparent_28%)]"></div>
      <div class="relative z-10 grid gap-4 md:grid-cols-[1.2fr_0.8fr] md:items-end">
        <div>
          <span class="inline-flex rounded-full bg-[#fdd404] px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.08em] text-[#594a00]">
            Tổng quan hệ thống
          </span>
          <h2 class="mt-4 max-w-[18rem] text-[1.7rem] font-black leading-[1.08] md:max-w-[28rem] md:text-[2.2rem]">
            Xin chào, {{ greetingName }}
          </h2>
          <p class="mt-3 max-w-[24rem] text-sm leading-6 text-white/88 md:max-w-[30rem] md:text-[0.98rem]">
            Theo dõi phòng chơi, tin tức hệ thống và thông báo ngay trên trang chủ.
          </p>
        </div>

        <div class="grid gap-3 md:justify-self-end">
          <RouterLink to="/notifications" class="rounded-[22px] bg-white/14 px-4 py-3 backdrop-blur-md">
            <p class="m-0 text-[0.7rem] uppercase tracking-[0.12em] text-white/72">Thông báo chưa đọc</p>
          <strong class="mt-1 block text-[1.35rem] font-black">{{ getUnreadCount() }}</strong>
          </RouterLink>
          <RouterLink to="/promotion" class="rounded-[22px] bg-white px-4 py-3 text-primary shadow-[0_8px_24px_rgba(255,255,255,0.16)]">
            <p class="m-0 text-[0.7rem] uppercase tracking-[0.12em] text-primary/70">Hoạt động</p>
            <strong class="mt-1 block text-[1rem] font-black">Tin tức hệ thống</strong>
          </RouterLink>
        </div>
      </div>
    </section>

    <section class="grid grid-cols-2 gap-3 md:grid-cols-4">
      <article
        v-for="metric in homeMetrics"
        :key="metric.title"
        class="rounded-[22px] bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.06)]"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="m-0 text-[0.72rem] font-bold uppercase tracking-[0.08em] text-on-surface-variant">
              {{ metric.title }}
            </p>
            <strong class="mt-2 block text-[1rem] font-black text-on-surface md:text-[1.08rem]">
              {{ metric.value }}
            </strong>
          </div>
          <span class="grid h-11 w-11 place-items-center rounded-[16px]" :style="{ backgroundColor: `${metric.accent}15`, color: metric.accent }">
            <span class="material-symbols-outlined">{{ metric.icon }}</span>
          </span>
        </div>
        <p class="mt-3 text-[0.72rem] leading-5 text-on-surface-variant">
          {{ metric.description }}
        </p>
      </article>
    </section>

    <section class="grid grid-cols-3 gap-2.5 md:grid-cols-5 md:gap-3">
      <RouterLink
        v-for="item in quickActions"
        :key="item.title"
        :to="item.to"
        class="grid min-h-[88px] place-items-center gap-2 rounded-[20px] bg-white px-2 py-3 text-on-surface shadow-[0_6px_18px_rgba(0,78,219,0.06)] transition-transform active:scale-95"
      >
        <span class="grid h-10 w-10 place-items-center rounded-full" :style="{ backgroundColor: `${item.accent}15`, color: item.accent }">
          <span class="material-symbols-outlined">{{ item.symbol }}</span>
        </span>
        <span class="text-center text-[0.68rem] font-extrabold">{{ item.title }}</span>
      </RouterLink>
    </section>

    <section class="flex items-center justify-between gap-3">
      <div>
        <h3 class="m-0 text-[1.05rem] font-black md:text-[1.15rem]">Phòng chơi nổi bật</h3>
        <p class="mt-1 text-[0.76rem] text-on-surface-variant">Danh sách các game-room đang mở, sẵn sàng join.</p>
      </div>
      <RouterLink to="/play/wingo" class="inline-flex items-center gap-1 text-[0.78rem] font-extrabold text-primary">
        Xem phòng <span class="material-symbols-outlined text-[1rem]">chevron_right</span>
      </RouterLink>
    </section>

    <section class="grid grid-cols-1 gap-3 md:grid-cols-3">
      <article
        v-for="game in featuredRooms"
        :key="game.code"
        class="overflow-hidden rounded-[24px] border-b-4 bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)]"
        :style="{ borderBottomColor: game.accent }"
      >
        <div class="flex items-start justify-between gap-3">
          <div
            class="grid h-[44px] w-[44px] place-items-center rounded-[14px]"
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

        <h4 class="m-0 mt-3 text-[1rem] font-black">{{ game.title }}</h4>
        <p class="mt-1.5 text-[0.72rem] leading-6 text-on-surface-variant">
          {{ game.subtitle }}
        </p>

        <div class="mt-4 grid grid-cols-2 gap-2 text-[0.72rem]">
          <div class="rounded-[16px] bg-background p-3">
            <span class="block text-on-surface-variant">Kỳ</span>
            <strong class="mt-1 block text-on-surface">{{ game.roundTime }}</strong>
          </div>
          <div class="rounded-[16px] bg-background p-3">
            <span class="block text-on-surface-variant">Người chơi</span>
            <strong class="mt-1 block text-on-surface">{{ game.onlinePlayers.toLocaleString('vi-VN') }}</strong>
          </div>
        </div>

        <RouterLink
          :to="`/play/${game.code}`"
          class="mt-4 flex min-h-12 items-center justify-center rounded-[16px] bg-gradient-to-br from-primary to-primary-container text-[0.82rem] font-black text-white"
        >
          Vào phòng
        </RouterLink>
      </article>
    </section>

    <section class="grid gap-3 md:grid-cols-[1.1fr_0.9fr]">
      <article class="rounded-[24px] bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)] md:p-5">
        <div class="flex items-center justify-between gap-2">
          <div>
            <h3 class="m-0 text-[1rem] font-black md:text-[1.08rem]">Hoạt động gần đây</h3>
            <p class="m-0 mt-1 text-[0.76rem] text-on-surface-variant">Dòng sự kiện hệ thống đang được tạo tự động.</p>
          </div>
          <RouterLink to="/promotion" class="text-[0.76rem] font-extrabold text-primary">Mở rộng</RouterLink>
        </div>

        <div class="mt-4 space-y-3">
          <article
            v-for="item in homeActivities"
            :key="item.title"
            class="grid grid-cols-[auto_1fr_auto] items-center gap-3 rounded-[18px] bg-background p-3"
          >
            <div
              class="grid h-10 w-10 place-items-center rounded-full text-white"
              :class="{
                'bg-emerald-500': item.tone === 'success',
                'bg-primary': item.tone === 'info',
                'bg-amber-500': item.tone === 'warning',
              }"
            >
              <span class="material-symbols-outlined text-[1.05rem]">{{ item.symbol }}</span>
            </div>

            <div>
              <strong class="block text-[0.82rem]">{{ item.title }}</strong>
              <p class="m-0 mt-1 text-[0.68rem] text-on-surface-variant">{{ item.subtitle }}</p>
            </div>

            <div class="text-right">
              <strong class="block text-[0.82rem]" :class="item.tone === 'warning' ? 'text-amber-600' : 'text-emerald-600'">
                {{ item.amount }}
              </strong>
              <span class="mt-0.5 block text-[0.64rem] text-on-surface-variant">{{ item.tag }}</span>
            </div>
          </article>
        </div>
      </article>

      <article class="rounded-[24px] bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)] md:p-5">
        <div class="flex items-center justify-between gap-2">
          <div>
            <h3 class="m-0 text-[1rem] font-black md:text-[1.08rem]">Tin mới</h3>
            <p class="m-0 mt-1 text-[0.76rem] text-on-surface-variant">Bản tin hệ thống và ưu đãi.</p>
          </div>
          <RouterLink to="/promotion" class="text-[0.76rem] font-extrabold text-primary">Xem tất cả</RouterLink>
        </div>

        <div class="mt-4 space-y-3">
          <RouterLink
            v-for="article in featuredNews"
            :key="article.slug"
            :to="`/news/${article.slug}`"
            class="block overflow-hidden rounded-[18px] border border-slate-100 bg-background transition-transform active:scale-[0.99]"
          >
            <div class="h-24 bg-gradient-to-br" :class="article.cover"></div>
            <div class="p-4">
              <div class="flex items-center justify-between gap-2 text-[0.68rem] uppercase tracking-[0.08em] text-on-surface-variant">
                <span>{{ article.category }}</span>
                <span>{{ formatViDateTime(article.publishedAt) }}</span>
              </div>
              <h4 class="mt-2 text-[0.92rem] font-black leading-6">{{ article.title }}</h4>
              <p class="mt-1.5 text-[0.72rem] leading-6 text-on-surface-variant">{{ article.excerpt }}</p>
            </div>
          </RouterLink>
        </div>
      </article>
    </section>
  </div>
</template>
