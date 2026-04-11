<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { formatViDateTime } from '@/shared/lib/date'
import { getNewsBySlug, getRelatedNews } from '@/data/site'

const route = useRoute()
const router = useRouter()

const article = computed(() => getNewsBySlug(String(route.params.slug ?? '')))
const relatedArticles = computed(() => (article.value ? getRelatedNews(article.value.slug, 3) : []))
</script>

<template>
  <div v-if="article" class="space-y-5 md:space-y-6">
    <header class="flex items-center justify-between gap-3">
      <button class="grid h-10 w-10 place-items-center rounded-full bg-white text-primary shadow-[0_6px_18px_rgba(255,109,102,0.06)]" type="button" @click="router.back()">
        <span class="material-symbols-outlined">arrow_back</span>
      </button>
      <RouterLink to="/promotion" class="text-sm font-extrabold text-primary">Quay về Hoạt động</RouterLink>
    </header>

    <section class="overflow-hidden rounded-[28px] bg-white shadow-[0_8px_18px_rgba(255,109,102,0.05)]">
      <div class="h-52 bg-gradient-to-br md:h-60" :class="article.cover"></div>
      <div class="p-5 md:p-6">
        <div class="flex flex-wrap items-center gap-2 text-[0.68rem] uppercase tracking-[0.08em] text-on-surface-variant">
          <span class="rounded-full bg-surface-container-low px-3 py-1 font-bold">{{ article.category }}</span>
          <span>{{ formatViDateTime(article.publishedAt) }}</span>
          <span>{{ article.readTime }}</span>
          <span>{{ article.author }}</span>
        </div>

        <h1 class="mt-4 text-[1.65rem] font-black leading-[1.15] md:text-[2rem]">{{ article.title }}</h1>
        <p class="mt-3 text-[0.92rem] leading-7 text-on-surface-variant">{{ article.excerpt }}</p>

        <div class="mt-5 space-y-4">
          <p v-for="paragraph in article.content" :key="paragraph" class="text-[0.92rem] leading-7 text-on-surface">
            {{ paragraph }}
          </p>
        </div>

        <div class="mt-5 flex flex-wrap gap-2">
          <span
            v-for="tag in article.tags"
            :key="tag"
            class="rounded-full bg-background px-3 py-1 text-[0.68rem] font-bold text-on-surface-variant"
          >
            #{{ tag }}
          </span>
        </div>
      </div>
    </section>

    <section class="space-y-3">
      <div class="flex items-center justify-between gap-2">
        <h2 class="m-0 text-[1rem] font-black md:text-[1.08rem]">Bài viết liên quan</h2>
        <RouterLink to="/promotion" class="text-[0.76rem] font-extrabold text-primary">Xem thêm</RouterLink>
      </div>

      <div class="grid gap-3 md:grid-cols-3">
        <RouterLink
          v-for="related in relatedArticles"
          :key="related.slug"
          :to="`/news/${related.slug}`"
          class="overflow-hidden rounded-[22px] bg-white shadow-[0_8px_18px_rgba(255,109,102,0.05)]"
        >
          <div class="h-28 bg-gradient-to-br" :class="related.cover"></div>
          <div class="p-4">
            <p class="m-0 text-[0.66rem] uppercase tracking-[0.08em] text-on-surface-variant">{{ related.category }}</p>
            <h3 class="mt-2 text-[0.92rem] font-black leading-6">{{ related.title }}</h3>
            <p class="mt-1.5 text-[0.72rem] leading-6 text-on-surface-variant">{{ related.excerpt }}</p>
          </div>
        </RouterLink>
      </div>
    </section>
  </div>

  <div v-else class="grid min-h-[40vh] place-items-center rounded-[28px] bg-white p-8 text-center shadow-[0_8px_18px_rgba(255,109,102,0.05)]">
    <div>
      <h1 class="text-[1.3rem] font-black">Không tìm thấy bài viết</h1>
      <p class="mt-2 text-sm text-on-surface-variant">Bài viết có thể đã bị ẩn hoặc chưa tồn tại.</p>
      <RouterLink to="/promotion" class="mt-4 inline-flex rounded-[16px] bg-primary px-4 py-2 text-sm font-black text-white">
        Quay về Hoạt động
      </RouterLink>
    </div>
  </div>
</template>

