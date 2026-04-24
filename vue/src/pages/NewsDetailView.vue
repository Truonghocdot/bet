<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { request, type ApiError } from '@/shared/api/http'
import type { ContentDetailResponse, ContentNewsItem } from '@/shared/api/types'
import { formatViDateTime } from '@/shared/lib/date'
import { stripHtmlTags } from '@/shared/lib/html'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const error = ref('')
const article = ref<ContentNewsItem | null>(null)
const relatedArticles = ref<ContentNewsItem[]>([])

const articlePreview = computed(() => stripHtmlTags(article.value?.content))

function relatedPreview(item: ContentNewsItem): string {
  return item.excerpt?.trim() || stripHtmlTags(item.content) || 'Đang cập nhật nội dung...'
}

async function loadNewsDetail() {
  const slug = String(route.params.slug ?? '').trim()
  if (!slug) {
    article.value = null
    relatedArticles.value = []
    error.value = 'Không tìm thấy bài viết'
    return
  }

  loading.value = true
  error.value = ''
  try {
    const response = await request<ContentDetailResponse>('GET', `/v1/content/news/${encodeURIComponent(slug)}`)
    article.value = response.item
    relatedArticles.value = response.related || []
  } catch (e: any) {
    const apiErr = e as ApiError
    article.value = null
    relatedArticles.value = []
    error.value = apiErr?.message || 'Không tải được chi tiết tin tức'
  } finally {
    loading.value = false
  }
}

watch(
  () => route.params.slug,
  () => {
    void loadNewsDetail()
  },
)

onMounted(() => {
  void loadNewsDetail()
})
</script>

<template>
  <div v-if="loading" class="grid min-h-[40vh] place-items-center rounded-[28px] bg-white p-8 text-center shadow-[0_8px_18px_rgba(255,109,102,0.05)]">
    <p class="text-sm font-semibold text-on-surface-variant">Đang tải bài viết...</p>
  </div>

  <div v-else-if="article" class="space-y-5 md:space-y-6">
    <header class="flex items-center justify-between gap-3">
      <button class="grid h-10 w-10 place-items-center rounded-full bg-white text-primary shadow-[0_6px_18px_rgba(255,109,102,0.06)]" type="button" @click="router.back()">
        <span class="material-symbols-outlined">arrow_back</span>
      </button>
      <RouterLink to="/promotion" class="text-sm font-extrabold text-primary">Quay về Hoạt động</RouterLink>
    </header>

    <section class="overflow-hidden rounded-[28px] bg-white shadow-[0_8px_18px_rgba(255,109,102,0.05)]">
      <img
        v-if="article.cover_image_url"
        :src="article.cover_image_url"
        :alt="article.title"
        class="h-52 w-full object-cover md:h-60"
      >
      <div v-else class="h-52 bg-gradient-to-br from-[#ff6d66] to-[#ff9f98] md:h-60"></div>
      <div class="p-5 md:p-6">
        <div class="flex flex-wrap items-center gap-2 text-[0.68rem] uppercase tracking-[0.08em] text-on-surface-variant">
          <span class="rounded-full bg-surface-container-low px-3 py-1 font-bold">Tin tức</span>
          <span>{{ formatViDateTime(article.published_at || article.created_at) }}</span>
        </div>

        <h1 class="mt-4 text-[1.65rem] font-black leading-[1.15] md:text-[2rem]">{{ article.title }}</h1>
        <p v-if="article.excerpt" class="mt-3 text-[0.92rem] leading-7 text-on-surface-variant">{{ article.excerpt }}</p>

        <div
          v-if="article.content"
          class="news-content mt-5 text-[0.92rem] leading-7 text-on-surface"
          v-html="article.content"
        />
        <p v-else-if="articlePreview" class="mt-5 text-[0.92rem] leading-7 text-on-surface">{{ articlePreview }}</p>
      </div>
    </section>

    <section v-if="relatedArticles.length" class="space-y-3">
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
          <img
            v-if="related.cover_image_url"
            :src="related.cover_image_url"
            :alt="related.title"
            class="h-28 w-full object-cover"
          >
          <div v-else class="h-28 bg-gradient-to-br from-[#ff6d66] to-[#ff9f98]"></div>
          <div class="p-4">
            <p class="m-0 text-[0.66rem] uppercase tracking-[0.08em] text-on-surface-variant">Tin tức</p>
            <h3 class="mt-2 text-[0.92rem] font-black leading-6">{{ related.title }}</h3>
            <p class="mt-1.5 text-[0.72rem] leading-6 text-on-surface-variant">{{ relatedPreview(related) }}</p>
          </div>
        </RouterLink>
      </div>
    </section>
  </div>

  <div v-else class="grid min-h-[40vh] place-items-center rounded-[28px] bg-white p-8 text-center shadow-[0_8px_18px_rgba(255,109,102,0.05)]">
    <div>
      <h1 class="text-[1.3rem] font-black">Không tìm thấy bài viết</h1>
      <p class="mt-2 text-sm text-on-surface-variant">{{ error || 'Bài viết có thể đã bị ẩn hoặc chưa tồn tại.' }}</p>
      <RouterLink to="/promotion" class="mt-4 inline-flex rounded-[16px] bg-primary px-4 py-2 text-sm font-black text-white">
        Quay về Hoạt động
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
.news-content :deep(p) {
  margin: 0 0 1rem;
}

.news-content :deep(h1),
.news-content :deep(h2),
.news-content :deep(h3),
.news-content :deep(h4) {
  margin: 1.25rem 0 0.75rem;
  font-weight: 900;
  line-height: 1.3;
}

.news-content :deep(ul),
.news-content :deep(ol) {
  margin: 0 0 1rem;
  padding-left: 1.25rem;
}

.news-content :deep(li) {
  margin: 0.25rem 0;
}

.news-content :deep(a) {
  color: #ef4444;
  font-weight: 800;
  text-decoration: underline;
}

.news-content :deep(img) {
  margin: 1rem 0;
  border-radius: 18px;
}

.news-content :deep(blockquote) {
  margin: 1rem 0;
  border-left: 4px solid #fecaca;
  padding-left: 1rem;
  color: #64748b;
}
</style>
