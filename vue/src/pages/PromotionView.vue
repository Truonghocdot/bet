<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { request, type ApiError } from '@/shared/api/http'
import type { ContentListResponse, ContentNewsItem } from '@/shared/api/types'
import { useAuthStore } from '@/stores/auth'

const activeTab = ref<'affiliate' | 'promotion' | 'news'>('affiliate')
const auth = useAuthStore()
const loading = ref(false)
const error = ref('')
const promotionItems = ref<ContentNewsItem[]>([])
const newsItems = ref<ContentNewsItem[]>([])
const invitedUsersCount = ref(0)
const copied = ref(false)

const affiliateStatusLabel = computed(() => {
  const status = auth.affiliateProfile?.status
  if (status === 2) return 'Đang hoạt động'
  if (status === 1) return 'Đang chờ duyệt'
  if (status === 3) return 'Tạm khóa'
  return 'Chưa đăng ký'
})

const affiliateCards = computed(() => [
  { label: 'Mã giới thiệu', value: auth.affiliateProfile?.ref_code || '—' },
  { label: 'Người đã mời', value: invitedUsersCount.value.toString() },
  { label: 'Link giới thiệu', value: auth.affiliateProfile?.ref_link || '—' },
  { label: 'Trạng thái', value: affiliateStatusLabel.value },
])

const tabTitle = computed(() => {
  if (activeTab.value === 'affiliate') return 'Affiliate'
  if (activeTab.value === 'promotion') return 'Khuyến mãi'
  return 'Tin tức'
})

async function loadContent() {
  loading.value = true
  error.value = ''
  try {
    const [promotions, news] = await Promise.all([
      request<ContentListResponse>('GET', '/v1/content/promotions?page=1&page_size=50'),
      request<ContentListResponse>('GET', '/v1/content/news?page=1&page_size=50'),
    ])
    promotionItems.value = promotions.items || []
    newsItems.value = news.items || []
  } catch (e: any) {
    const err = e as ApiError
    error.value = err?.message ?? 'Không thể tải dữ liệu nội dung'
    promotionItems.value = []
    newsItems.value = []
  } finally {
    loading.value = false
  }
}

async function copyReferralLink() {
  const link = auth.affiliateProfile?.ref_link
  if (!link) return
  try {
    await navigator.clipboard.writeText(link)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 1500)
  } catch {
    copied.value = false
  }
}

onMounted(async () => {
  if (auth.isAuthenticated && !auth.affiliateProfile) {
    try {
      await auth.fetchMe()
    } catch {
      // handled by auth store
    }
  }
  await loadContent()
})
</script>

<template>
  <div class="space-y-3 pb-8">
    <div class="flex items-center gap-2 border-b border-slate-100 bg-white px-4 py-4">
      <span class="material-symbols-outlined text-[1.3rem] text-primary">group</span>
      <h1 class="text-[1rem] font-black text-on-surface">{{ tabTitle }}</h1>
    </div>

    <section class="px-3">
      <div class="grid grid-cols-3 gap-2 rounded-[18px] bg-white p-1.5 shadow-sm border border-slate-100">
        <button
          class="min-h-10 rounded-[12px] text-[0.78rem] font-black transition-all"
          :class="activeTab === 'affiliate' ? 'bg-primary text-white' : 'text-slate-500'"
          @click="activeTab = 'affiliate'"
        >
          Affiliate
        </button>
        <button
          class="min-h-10 rounded-[12px] text-[0.78rem] font-black transition-all"
          :class="activeTab === 'promotion' ? 'bg-primary text-white' : 'text-slate-500'"
          @click="activeTab = 'promotion'"
        >
          Khuyến mãi
        </button>
        <button
          class="min-h-10 rounded-[12px] text-[0.78rem] font-black transition-all"
          :class="activeTab === 'news' ? 'bg-primary text-white' : 'text-slate-500'"
          @click="activeTab = 'news'"
        >
          Tin tức
        </button>
      </div>
    </section>

    <section v-if="activeTab === 'affiliate'" class="flex flex-col gap-3 px-3">
      <div class="rounded-[18px] bg-gradient-to-br from-[#ff6d66] to-[#ff9f98] p-4 text-white">
        <p class="m-0 text-[0.7rem] uppercase tracking-[0.08em] text-white/80">Chương trình Affiliate</p>
        <h2 class="mt-2 text-[1.15rem] font-black">Mời bạn bè, nhận hoa hồng theo doanh thu</h2>
        <p class="mt-2 text-[0.78rem] leading-5 text-white/90">
          Chia sẻ mã giới thiệu, theo dõi người đăng ký và nhận thưởng theo chính sách đại lý.
        </p>
      </div>

      <div class="grid gap-2 md:grid-cols-3">
        <article v-for="item in affiliateCards" :key="item.label" class="rounded-[16px] border border-slate-100 bg-white p-4 shadow-sm">
          <p class="m-0 text-[0.7rem] font-bold uppercase tracking-[0.05em] text-slate-500">{{ item.label }}</p>
          <p class="mt-2 break-all text-[0.9rem] font-black text-on-surface">{{ item.value }}</p>
        </article>
      </div>

      <button
        type="button"
        class="inline-flex w-fit items-center gap-2 rounded-full bg-primary px-4 py-2 text-[0.75rem] font-black text-white"
        @click="copyReferralLink"
      >
        <span class="material-symbols-outlined text-[1rem]">content_copy</span>
        {{ copied ? 'Đã copy link' : 'Copy link giới thiệu' }}
      </button>
    </section>

    <section v-if="error" class="mx-3 rounded-[14px] bg-red-50 px-4 py-3 text-sm font-semibold text-red-600">
      {{ error }}
    </section>

    <section v-if="loading" class="mx-3 rounded-[14px] border border-slate-100 bg-white px-4 py-4 text-sm font-semibold text-slate-500">
      Đang tải dữ liệu...
    </section>

    <div class="flex flex-col gap-3 px-3">
      <div
        v-if="activeTab === 'promotion'"
        v-for="item in promotionItems"
        :key="item.id"
        class="flex overflow-hidden rounded-[18px] bg-white shadow-sm border border-slate-100 transition-all active:scale-[0.99] hover:-translate-y-0.5"
      >
        <!-- Left accent bar -->
        <div class="w-1.5 flex-shrink-0 bg-[#e8404a]" />
        <div class="flex flex-1 gap-3 px-4 py-4">
          <!-- Icon -->
          <div class="flex-shrink-0 grid h-11 w-11 place-items-center rounded-[14px] mt-0.5 bg-red-50 text-[#e8404a]">
            <span class="material-symbols-outlined text-[1.3rem]">campaign</span>
          </div>
          <!-- Content -->
          <div class="flex-1 min-w-0">
            <div class="flex items-start justify-between gap-2">
              <strong class="text-[0.9rem] font-black text-on-surface leading-snug">{{ item.title }}</strong>
            </div>
            <p class="mt-1 text-[0.75rem] leading-5 text-slate-500">{{ item.excerpt || item.content }}</p>
            <p class="mt-3 text-[0.68rem] font-bold text-slate-400">{{ item.published_at || item.created_at }}</p>
          </div>
        </div>
      </div>

      <div
        v-if="activeTab === 'promotion' && !loading && promotionItems.length === 0"
        class="rounded-[14px] border border-slate-100 bg-white px-4 py-4 text-sm font-semibold text-slate-500"
      >
        Chưa có dữ liệu khuyến mãi.
      </div>

      <RouterLink
        v-if="activeTab === 'news'"
        v-for="item in newsItems"
        :key="item.id"
        :to="`/news/${item.slug}`"
        class="rounded-[18px] border border-slate-100 bg-white p-4 shadow-sm"
      >
        <h3 class="text-[0.92rem] font-black text-on-surface">{{ item.title }}</h3>
        <p class="mt-2 text-[0.78rem] leading-5 text-slate-500">{{ item.excerpt || item.content }}</p>
        <p class="mt-3 text-[0.68rem] font-bold text-slate-400">{{ item.published_at || item.created_at }}</p>
      </RouterLink>

      <div
        v-if="activeTab === 'news' && !loading && newsItems.length === 0"
        class="rounded-[14px] border border-slate-100 bg-white px-4 py-4 text-sm font-semibold text-slate-500"
      >
        Chưa có dữ liệu tin tức.
      </div>
    </div>
  </div>
</template>
