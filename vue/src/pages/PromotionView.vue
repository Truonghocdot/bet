<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { formatViDateTime } from '@/shared/lib/date'
import { newsArticles, type NewsArticle } from '@/data/site'

const tabs = ['Tất cả', 'Khuyến mãi', 'Tin hệ thống', 'VIP Club', 'Referral']
const activeTab = ref<string>(tabs[0] ?? 'Tất cả')

const filteredArticles = computed(() => {
  if (activeTab.value === 'Tất cả') return newsArticles
  return newsArticles.filter((article) => article.category === activeTab.value || article.tags.includes(activeTab.value.toLowerCase()))
})

const leadArticle = computed<NewsArticle>(() => filteredArticles.value[0] ?? newsArticles[0]!)
const compactArticles = computed(() => filteredArticles.value.slice(1, 4))
</script>

<template>
  <div class="space-y-5 md:space-y-6">
    <section class="relative overflow-hidden rounded-[28px] bg-gradient-to-br from-[#004edb] via-[#0058bb] to-[#7e9cff] p-5 text-white shadow-[0_12px_30px_rgba(0,78,219,0.2)] md:p-6">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(255,255,255,0.24),transparent_26%),radial-gradient(circle_at_bottom_left,rgba(255,255,255,0.08),transparent_28%)]"></div>
      <div class="relative z-10 grid gap-4 md:grid-cols-[1fr_auto] md:items-end">
        <div>
          <span class="inline-flex rounded-full bg-[#fdd404] px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.08em] text-[#594a00]">
            Hoạt động
          </span>
          <h2 class="mt-4 max-w-[22rem] text-[1.75rem] font-black leading-[1.08] md:max-w-[34rem] md:text-[2.2rem]">
            Tin tức và khuyến mãi được cập nhật tự động
          </h2>
          <p class="mt-3 max-w-[30rem] text-sm leading-6 text-white/88">
            Đây là nơi hệ thống tổng hợp các tin bài, chương trình và update vận hành để người chơi xem nhanh trước khi vào phòng.
          </p>
        </div>

        <RouterLink :to="`/news/${leadArticle.slug}`" class="rounded-[22px] bg-white px-4 py-3 text-primary shadow-[0_8px_24px_rgba(255,255,255,0.16)]">
          <p class="m-0 text-[0.7rem] uppercase tracking-[0.12em] text-primary/70">Tin nổi bật</p>
          <strong class="mt-1 block text-[1rem] font-black">{{ leadArticle.title }}</strong>
        </RouterLink>
      </div>
    </section>

    <section class="overflow-x-auto pb-2 no-scrollbar">
      <div class="flex min-w-max gap-2">
        <button
          v-for="tab in tabs"
          :key="tab"
          class="rounded-full px-5 py-2.5 text-[0.78rem] font-bold whitespace-nowrap transition-colors"
          :class="tab === activeTab ? 'bg-primary text-white shadow-[0_12px_32px_rgba(0,78,219,0.1)]' : 'bg-surface-container-low text-on-surface-variant'"
          type="button"
          @click="activeTab = tab"
        >
          {{ tab }}
        </button>
      </div>
    </section>

    <section class="grid gap-3 xl:grid-cols-[1.08fr_0.92fr]">
      <article class="overflow-hidden rounded-[26px] bg-white shadow-[0_8px_18px_rgba(0,78,219,0.05)]">
        <div class="h-36 bg-gradient-to-br md:h-40" :class="leadArticle.cover"></div>
        <div class="p-4 md:p-5">
          <div class="flex items-center justify-between gap-2 text-[0.68rem] uppercase tracking-[0.08em] text-on-surface-variant">
            <span>{{ leadArticle.category }}</span>
            <span>{{ formatViDateTime(leadArticle.publishedAt) }}</span>
          </div>
          <h3 class="mt-2 text-[1.1rem] font-black leading-7 md:text-[1.2rem]">{{ leadArticle.title }}</h3>
          <p class="mt-2 text-[0.8rem] leading-6 text-on-surface-variant md:text-[0.82rem]">{{ leadArticle.excerpt }}</p>
          <div class="mt-3 flex flex-wrap gap-2">
            <span
              v-for="tag in leadArticle.tags.slice(0, 3)"
              :key="tag"
              class="rounded-full bg-surface-container-low px-3 py-1 text-[0.66rem] font-bold text-on-surface-variant"
            >
              #{{ tag }}
            </span>
          </div>
          <div class="mt-4 flex flex-wrap items-center gap-2">
            <RouterLink
              :to="`/news/${leadArticle.slug}`"
              class="inline-flex min-h-11 items-center justify-center rounded-[16px] bg-gradient-to-br from-primary to-primary-container px-4 text-[0.8rem] font-black text-white"
            >
              Xem chi tiết
            </RouterLink>
            <span class="rounded-full bg-primary/10 px-3 py-1 text-[0.68rem] font-black uppercase tracking-[0.08em] text-primary">
              {{ leadArticle.readTime }}
            </span>
          </div>
        </div>
      </article>

      <article class="rounded-[26px] bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)] md:p-5">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h3 class="m-0 text-[1rem] font-black">Feed cập nhật</h3>
            <p class="mt-1 text-[0.76rem] leading-6 text-on-surface-variant">
              Tin được sắp theo khối ngắn để người chơi lướt nhanh trong cùng một khung hình.
            </p>
          </div>
          <span class="rounded-full bg-primary/10 px-3 py-1 text-[0.68rem] font-black uppercase tracking-[0.08em] text-primary">
            {{ filteredArticles.length }} bài
          </span>
        </div>

        <div class="mt-4 grid gap-2">
          <RouterLink
            v-for="article in compactArticles"
            :key="article.slug"
            :to="`/news/${article.slug}`"
            class="grid grid-cols-[72px_1fr] gap-3 rounded-[18px] border border-slate-100 bg-background p-2.5 transition-transform active:scale-[0.99]"
          >
            <div class="h-[72px] min-h-[72px] rounded-[14px] bg-gradient-to-br" :class="article.cover"></div>
            <div class="min-w-0 py-0.5 pr-1">
              <div class="flex items-center justify-between gap-2 text-[0.62rem] uppercase tracking-[0.08em] text-on-surface-variant">
                <span>{{ article.category }}</span>
                <span>{{ article.readTime }}</span>
              </div>
              <h4 class="mt-1 text-[0.92rem] font-black leading-5">{{ article.title }}</h4>
              <p class="mt-1.5 max-h-[2.6rem] overflow-hidden text-[0.7rem] leading-5 text-on-surface-variant">{{ article.excerpt }}</p>
            </div>
          </RouterLink>
        </div>

        <div class="mt-4 flex flex-wrap gap-2">
          <span class="rounded-full bg-surface-container-low px-3 py-1 text-[0.66rem] font-bold text-on-surface-variant">
            Tổng bài {{ newsArticles.length }}
          </span>
          <span class="rounded-full bg-surface-container-low px-3 py-1 text-[0.66rem] font-bold text-on-surface-variant">
            Mới nhất {{ formatViDateTime(leadArticle.publishedAt) }}
          </span>
        </div>
      </article>
    </section>
  </div>
</template>
