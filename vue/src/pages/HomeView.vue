<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { RouterLink, type RouteLocationRaw } from 'vue-router'

import BannerCarousel from '@/components/BannerCarousel.vue'
import MarqueeBar from '@/components/MarqueeBar.vue'
import aceForcesLogo from '@/assets/supporter/aceforces.jpg'
import macjbeliLogo from '@/assets/supporter/macjbeli.jpg'
import pieExglnLogo from '@/assets/supporter/pie.exgln.png'
import { request } from '@/shared/api/http'
import type { ContentBannerItem, ContentHomeResponse, ContentNewsItem, FakeFinanceFeedItem, FakeFinanceFeedResponse } from '@/shared/api/types'
import { stripHtmlTags } from '@/shared/lib/html'
import { formatViMoney } from '@/shared/lib/money'
import { useAuthStore } from '@/stores/auth'
import { useWalletStore } from '@/stores/wallet'

// Category Icons
import catHot from '@/assets/category_game/hot.avif'
import catLottery from '@/assets/category_game/lottery.avif'
import catCasino from '@/assets/category_game/lobbypoker.avif'
import catJackpot from '@/assets/category_game/jackpot.avif'
import catHuntfish from '@/assets/category_game/huntfish.avif'
import catFootball from '@/assets/category_game/football.avif'
import catPoker from '@/assets/category_game/poker.avif'
import catChicken from '@/assets/category_game/chicken.avif'

// Main Game Banners
import bannerWingo from '@/assets/lottery_banner/optimized/wingo.webp'
import bannerK3 from '@/assets/lottery_banner/optimized/k3.webp'
import banner5D from '@/assets/lottery_banner/optimized/5d.webp'

// Game Thumbnails
const gameThumbModules = import.meta.glob<string>('@/assets/game_thumbs/*/*.webp', {
  eager: true,
  query: '?url',
  import: 'default',
})

const auth = useAuthStore()
const wallet = useWalletStore()
const route = useRoute()
const popupVideoUrl = `${import.meta.env.VITE_API_BASE_URL}/v1/media/popup-video`

const greetingName = computed(() => auth.user?.name || 'Bạn')
const vndWallet = computed(() => wallet.wallets.find((item) => item.unit === 1) ?? null)
const homeBanners = ref<ContentBannerItem[]>([])
const homeHighlights = ref<ContentNewsItem[]>([])
const contentError = ref('')
const fakeDepositFeed = ref<FakeFinanceFeedItem[]>([])
const fakeWithdrawFeed = ref<FakeFinanceFeedItem[]>([])
let fakeFinancePollTicker: number | undefined

function displayBalance(value: string | number | null | undefined) {
  return formatViMoney(value ?? 0, 0)
}

function newsPreview(item: ContentNewsItem) {
  return item.excerpt?.trim() || stripHtmlTags(item.content) || 'Đang cập nhật nội dung...'
}

const telegramLink = computed(() => wallet.summary?.telegram_cskh_link || 'https://t.me/CSKH_FH88U')
function openTelegram() { window.open(telegramLink.value, '_blank') }

const activeCategory = ref('Phổ biến')
let playRoutePrefetched = false
let partnerGameRoutePrefetched = false

type IdleScheduler = (callback: () => void, options?: { timeout?: number }) => number

const categorySidebar = [
  { label: 'Xổ số', icon: catLottery },
  { label: 'Casino', icon: catCasino },
  { label: 'Nổ hũ', icon: catJackpot },
  { label: 'Bắn cá', icon: catHuntfish },
  { label: 'Thể thao', icon: catFootball },
  { label: 'Game bài', icon: catPoker },
  { label: 'Đá gà', icon: catChicken },
]

interface GameItem {
  name: string
  image: string
  category: string[]
  route?: string
  partnerLobby?: boolean
  featured?: boolean
}

function prefetchPlayRoute() {
  if (playRoutePrefetched) return

  playRoutePrefetched = true
  void Promise.all([
    import('@/pages/PlayView.vue'),
    import('@/data/play'),
  ]).catch(() => {
    playRoutePrefetched = false
  })
}

function prefetchPartnerGameRoute() {
  if (partnerGameRoutePrefetched) return

  partnerGameRoutePrefetched = true
  void import('@/pages/PartnerGameLoadingView.vue').catch(() => {
    partnerGameRoutePrefetched = false
  })
}

function prefetchPlayRouteSoon() {
  const requestIdleCallback = (window as Window & { requestIdleCallback?: IdleScheduler }).requestIdleCallback

  if (requestIdleCallback) {
    requestIdleCallback(() => prefetchPlayRoute(), { timeout: 2500 })
    return
  }

  window.setTimeout(() => prefetchPlayRoute(), 900)
}

function maybePrefetchGameRoute(game: GameItem) {
  if (game.route) {
    prefetchPlayRoute()
    return
  }

  if (game.partnerLobby) {
    prefetchPartnerGameRoute()
  }
}

function resolveGameTarget(game: GameItem): RouteLocationRaw {
  if (game.route) {
    return { path: game.route, query: { from: route.fullPath } }
  }

  return {
    name: 'partner-game-loading',
    query: {
      name: game.name,
      from: route.fullPath,
    },
  }
}

type ThumbCategoryConfig = {
  category: string
  popular?: boolean
  featuredFirst?: boolean
}

const thumbCategoryMap: Record<string, ThumbCategoryConfig> = {
  'lobby-casino': { category: 'Casino', popular: true, featuredFirst: true },
  slot: { category: 'Nổ hũ', popular: true, featuredFirst: true },
  'hunt-fish': { category: 'Bắn cá', popular: true, featuredFirst: true },
  sport: { category: 'Thể thao' },
  poker: { category: 'Game bài', popular: true },
  chicken: { category: 'Đá gà' },
}

function prettyGameThumbName(fileName: string) {
  return fileName
    .replace(/\.[^.]+$/, '')
    .replace(/[-_]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

function buildThumbGames(folder: string, config: ThumbCategoryConfig): GameItem[] {
  return Object.entries(gameThumbModules)
    .filter(([path]) => path.includes(`/game_thumbs/${folder}/`))
    .sort(([a], [b]) => a.localeCompare(b, 'vi'))
    .map(([path, image], index) => {
      const fileName = path.split('/').pop() ?? ''
      return {
        name: prettyGameThumbName(fileName),
        image,
        category: [
          ...(config.popular && index < 2 ? ['Phổ biến'] : []),
          config.category,
        ],
        partnerLobby: true,
        featured: Boolean(config.featuredFirst && index === 0),
      }
    })
}

const thumbGames = Object.entries(thumbCategoryMap).flatMap(([folder, config]) => buildThumbGames(folder, config))

const allGames: GameItem[] = [
  // Xổ số - có route thật
  { name: 'Win Go', image: bannerWingo, category: ['Phổ biến', 'Xổ số'], route: '/play/wingo', featured: true },
  { name: 'K3', image: bannerK3, category: ['Phổ biến', 'Xổ số'], route: '/play/k3' },
  { name: '5D Lottery', image: banner5D, category: ['Phổ biến', 'Xổ số'], route: '/play/lottery' },
  ...thumbGames,
]

const activePlayableGames = computed(() => allGames.filter(game => game.route))

const filteredGames = computed(() => {
  if (activeCategory.value === 'Phổ biến') {
    return activePlayableGames.value
  }

  return allGames.filter(g => g.category.includes(activeCategory.value))
})

const categoryBannerGames = computed(() => filteredGames.value.slice(0, 3))

const combinedFakeFinanceFeed = computed(() => {
  return [...fakeDepositFeed.value, ...fakeWithdrawFeed.value]
    .sort((left, right) => {
      const leftTime = Date.parse(left.created_at ?? '')
      const rightTime = Date.parse(right.created_at ?? '')
      return rightTime - leftTime
    })
    .slice(0, 12)
})
const showFakeFinanceFeed = computed(() => wallet.summary?.fake_finance_feed?.enabled ?? true)

const supporterLogos = [
  { name: 'PIE.EXGLN', image: pieExglnLogo },
  { name: 'ACE-FORCES', image: aceForcesLogo },
  { name: 'MACJBELI', image: macjbeliLogo },
]

async function fetchHomeContent() {
  contentError.value = ''
  try {
    const response = await request<ContentHomeResponse>('GET', '/v1/content/home')
    homeBanners.value = response.banners || []
    homeHighlights.value = response.highlights || []
  } catch {
    homeBanners.value = []
    homeHighlights.value = []
    contentError.value = 'Không thể tải nội dung trang chủ'
  }
}

async function refreshFakeFinanceFeed() {
  if (!showFakeFinanceFeed.value) {
    fakeDepositFeed.value = []
    fakeWithdrawFeed.value = []
    return
  }

  try {
    const [depositResponse, withdrawResponse] = await Promise.all([
      request<FakeFinanceFeedResponse>('GET', '/v1/finance-feed/deposit?limit=6'),
      request<FakeFinanceFeedResponse>('GET', '/v1/finance-feed/withdraw?limit=6'),
    ])
    fakeDepositFeed.value = depositResponse.items ?? []
    fakeWithdrawFeed.value = withdrawResponse.items ?? []
  } catch {
    fakeDepositFeed.value = []
    fakeWithdrawFeed.value = []
  }
}

function stopFakeFinancePolling() {
  if (fakeFinancePollTicker) {
    window.clearInterval(fakeFinancePollTicker)
    fakeFinancePollTicker = undefined
  }
}

function startFakeFinancePolling() {
  stopFakeFinancePolling()
  if (!showFakeFinanceFeed.value) return
  fakeFinancePollTicker = window.setInterval(() => {
    void refreshFakeFinanceFeed()
  }, 5000)
}

function fakeFinanceChannelLabel(item: FakeFinanceFeedItem) {
  return item.channel === 'withdraw' ? 'Rút' : 'Nạp'
}

function fakeFinanceChannelClass(item: FakeFinanceFeedItem) {
  return item.channel === 'withdraw'
    ? 'bg-amber-100 text-amber-700'
    : 'bg-emerald-100 text-emerald-700'
}

onMounted(() => {
  prefetchPlayRouteSoon()
  window.setTimeout(() => prefetchPartnerGameRoute(), 1200)
  void wallet.fetchSummary()
  void fetchHomeContent()
  void refreshFakeFinanceFeed()
  startFakeFinancePolling()
})

watch(
  showFakeFinanceFeed,
  (enabled) => {
    if (!enabled) {
      stopFakeFinancePolling()
      fakeDepositFeed.value = []
      fakeWithdrawFeed.value = []
      return
    }

    void refreshFakeFinanceFeed()
    startFakeFinancePolling()
  },
)

onBeforeUnmount(() => {
  stopFakeFinancePolling()
})
</script>

<template>
  <div class="pb-4">
    <div class="bg-[linear-gradient(180deg,#fff1f1_0%,#fff7f7_18%,#ffffff_100%)] pb-4">
      <div class="px-3 pt-2">
        <MarqueeBar />
      </div>

      <div class="px-3 pt-2">
        <div class="overflow-hidden rounded-[26px] border border-red-100 bg-white shadow-[0_10px_30px_rgba(218,37,29,0.14)]">
          <BannerCarousel :banners="homeBanners" />
        </div>
      </div>

      <section class="mx-3 mt-3 overflow-hidden rounded-[24px] bg-gradient-to-br from-red-700 via-primary to-red-800 p-4 text-white shadow-[0_14px_30px_rgba(218,37,29,0.3)]">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-[0.68rem] font-black uppercase tracking-[0.14em] text-white/75">Số dư ví VND</p>
            <strong class="mt-1 block break-words text-[1.55rem] font-black leading-tight text-white">
              {{ vndWallet ? displayBalance(vndWallet.balance) : '0' }}đ
            </strong>
            <p class="mt-1 text-[0.76rem] font-semibold text-white/85">Chào {{ greetingName }}</p>
          </div>
          <button
            type="button"
            class="shrink-0 rounded-full border border-white/35 bg-white/12 px-3 py-2 text-[0.72rem] font-black text-white backdrop-blur"
            @click="openTelegram()"
          >
            CSKH
          </button>
        </div>

        <div class="mt-4 grid grid-cols-2 gap-2.5">
          <RouterLink
            to="/account"
            class="flex min-h-12 items-center justify-center rounded-[14px] bg-white/90 px-3 text-[0.9rem] font-black text-primary shadow-[inset_0_-3px_0_rgba(218,37,29,0.12)] transition-transform active:scale-95"
          >
            Rút tiền
          </RouterLink>
          <RouterLink
            to="/deposit"
            class="flex min-h-12 items-center justify-center rounded-[14px] bg-gradient-to-r from-[#ff6d66] to-primary px-3 text-[0.9rem] font-black text-white shadow-[inset_0_-3px_0_rgba(0,0,0,0.12)] transition-transform active:scale-95"
          >
            Nạp tiền
          </RouterLink>
        </div>
      </section>

      <section class="mx-3 mt-3 overflow-hidden rounded-[24px] bg-gradient-to-br from-[#ff6e67] via-primary to-[#e73d47] shadow-[0_14px_28px_rgba(218,37,29,0.2)]">
        <div class="grid grid-cols-4 gap-px bg-white/15">
          <button
            v-for="cat in categorySidebar"
            :key="cat.label"
            type="button"
            class="group flex min-h-[86px] flex-col items-center justify-center gap-1.5 px-1 py-3 text-center transition-colors"
            :class="activeCategory === cat.label ? 'bg-[#c92633] text-white' : 'bg-transparent text-white/88'"
            @click="activeCategory = cat.label"
          >
            <div
              class="flex h-9 w-9 items-center justify-center rounded-full bg-white/12 transition-transform duration-200"
              :class="activeCategory === cat.label ? 'scale-110 bg-white/18' : 'group-active:scale-95'"
            >
              <img :src="cat.icon" :alt="cat.label" class="h-7 w-7 object-contain" />
            </div>
            <span class="text-[0.68rem] font-black leading-4">{{ cat.label }}</span>
          </button>
        </div>
      </section>
    </div>

    <section class="mt-4 px-3">
      <div class="flex min-w-0 flex-col gap-2">
        <RouterLink
          v-for="(game, index) in categoryBannerGames"
          :key="`${game.name}-banner`"
          :to="resolveGameTarget(game)"
          class="group relative block overflow-hidden rounded-[16px] shadow-[0_4px_16px_rgba(0,0,0,0.12)] transition-all duration-300 active:scale-[0.98]"
          @pointerenter="maybePrefetchGameRoute(game)"
          @focus="maybePrefetchGameRoute(game)"
          @touchstart.passive="maybePrefetchGameRoute(game)"
        >
          <img
            :src="game.image"
            :alt="game.name"
            class="block w-full object-cover transition-transform duration-500 ease-out group-hover:scale-[1.03]"
            :class="game.route ? 'aspect-[3/1]' : 'aspect-[2.55/1]'"
            decoding="async"
            :fetchpriority="index === 0 ? 'high' : 'low'"
            :loading="index === 0 ? 'eager' : 'lazy'"
          />
          <div class="absolute inset-0 bg-gradient-to-t from-black/60 via-black/10 to-transparent" />
          <div class="absolute bottom-0 left-0 right-0 flex items-center justify-between px-3 py-2">
            <div>
              <h4 class="text-[0.78rem] font-black tracking-wide text-white drop-shadow">{{ game.name }}</h4>
              <p class="text-[0.55rem] font-semibold text-white/70">
                {{ game.partnerLobby ? 'Mở sảnh đối tác' : 'Vào chơi ngay' }}
              </p>
            </div>
            <div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border border-white/30 bg-white/20 backdrop-blur-md">
              <span class="material-symbols-outlined text-[0.9rem] text-white">arrow_forward</span>
            </div>
          </div>
        </RouterLink>
      </div>
    </section>

    <section class="mt-4 px-3">
      <div class="mb-3 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="block h-5 w-1 rounded-full bg-primary" />
          <h2 class="text-[0.98rem] font-black text-slate-800">{{ activeCategory }}</h2>
        </div>
        <span class="rounded-full bg-red-50 px-2.5 py-1 text-[0.68rem] font-black text-primary">{{ filteredGames.length }} trò chơi</span>
      </div>

      <div class="grid grid-cols-2 gap-2.5 sm:grid-cols-3 lg:grid-cols-4">
        <RouterLink
          v-for="game in filteredGames"
          :key="game.name"
          :to="resolveGameTarget(game)"
          class="group relative block w-full overflow-hidden rounded-[16px] text-left shadow-[0_4px_14px_rgba(0,0,0,0.10)] transition-all duration-200 active:scale-[0.97]"
          @pointerenter="maybePrefetchGameRoute(game)"
          @focus="maybePrefetchGameRoute(game)"
          @touchstart.passive="maybePrefetchGameRoute(game)"
        >
          <img
            :src="game.image"
            :alt="game.name"
            class="w-full object-cover transition-transform duration-500 group-hover:scale-[1.04]"
            :class="game.route ? 'aspect-[3/1]' : 'aspect-[4/3]'"
            loading="lazy"
            decoding="async"
          />
          <div class="absolute inset-0 bg-gradient-to-t from-black/70 via-black/10 to-transparent" />
          <div class="absolute bottom-0 left-0 right-0 px-2 py-2.5">
            <p class="line-clamp-1 text-[0.75rem] font-black text-white drop-shadow">{{ game.name }}</p>
            <p class="mt-0.5 text-[0.55rem] font-semibold text-white/60">
              {{ game.partnerLobby ? 'Mở sảnh đối tác' : 'Vào chơi ngay' }}
            </p>
          </div>
        </RouterLink>
      </div>
    </section>

    <section v-if="showFakeFinanceFeed" class="mx-3 mt-4 overflow-hidden rounded-[20px] border border-slate-100 bg-white shadow-[0_8px_18px_rgba(0,0,0,0.08)]">
      <div class="flex items-center justify-between border-b border-slate-100 px-4 py-3.5">
        <h2 class="m-0 text-[0.92rem] font-black text-on-surface">Giao dịch gần đây</h2>
        <button
          type="button"
          class="text-[0.68rem] font-black uppercase tracking-[0.06em] text-primary"
          @click="void refreshFakeFinanceFeed()"
        >
          Làm mới
        </button>
      </div>
      <div v-if="combinedFakeFinanceFeed.length === 0" class="px-4 py-6 text-center text-[0.78rem] font-semibold text-slate-400">
        Chưa có dữ liệu giao dịch
      </div>
      <div v-else class="overflow-x-auto">
        <table class="min-w-full text-left">
          <thead class="bg-slate-50 text-[0.68rem] uppercase tracking-[0.06em] text-slate-500">
            <tr>
              <th class="px-4 py-3 font-black">Loại</th>
              <th class="px-4 py-3 font-black">Mã</th>
              <th class="px-4 py-3 font-black">Số điện thoại</th>
              <th class="px-4 py-3 font-black">Trạng thái</th>
              <th class="px-4 py-3 font-black">Thời gian</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-[0.78rem] text-slate-700">
            <tr v-for="item in combinedFakeFinanceFeed" :key="`${item.channel}-${item.id}`">
              <td class="px-4 py-3">
                <span class="rounded-full px-2.5 py-1 text-[0.66rem] font-black uppercase" :class="fakeFinanceChannelClass(item)">
                  {{ fakeFinanceChannelLabel(item) }}
                </span>
              </td>
              <td class="px-4 py-3 font-black text-slate-800">{{ item.masked_code }}</td>
              <td class="px-4 py-3 font-semibold text-slate-500">{{ item.masked_phone }}</td>
              <td class="px-4 py-3 font-bold text-emerald-600">{{ item.status_label }}</td>
              <td class="px-4 py-3 font-medium text-slate-400">{{ item.created_at }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="mx-3 mt-4 overflow-hidden rounded-[20px] border border-slate-100 bg-white shadow-[0_8px_18px_rgba(0,0,0,0.08)]">
      <div class="flex items-center justify-between border-b border-slate-100 px-4 py-3.5">
        <h2 class="m-0 text-[0.92rem] font-black text-on-surface">ĐẠI SỨ THƯƠNG HIỆU FH88U - Fernando Torres</h2>
      </div>
      <div class="p-3">
        <div class="overflow-hidden rounded-[16px] bg-black shadow-[0_10px_24px_rgba(0,0,0,0.22)]">
          <video
            class="aspect-video w-full bg-black object-cover"
            :src="popupVideoUrl"
            preload="metadata"
            playsinline
            controls
          />
        </div>
        <div class="mt-3 overflow-hidden rounded-[16px] border border-slate-100 bg-white shadow-[0_10px_24px_rgba(0,0,0,0.08)]">
          <img
            src="/ambassador.webp"
            alt="Đại sứ thương hiệu FH88U"
            class="block w-full object-cover"
            loading="lazy"
            decoding="async"
          />
        </div>
      </div>
    </section>

    <!-- ===== NEWS HIGHLIGHTS ===== -->
    <div class="mx-3 mt-4 mb-2 overflow-hidden rounded-[20px] bg-white/90 md:bg-black/30 md:backdrop-blur-md shadow-[0_8px_18px_rgba(0,0,0,0.1)] border border-slate-100 md:border-white/10">
      <div class="flex items-center gap-2 border-b border-slate-100 md:border-white/10 px-4 py-3.5">
        <span class="text-[1.1rem]">📰</span>
        <span class="text-[0.9rem] font-black text-on-surface md:text-white">Tin nổi bật</span>
      </div>
      <div class="divide-y divide-slate-50 md:divide-white/5">
        <RouterLink
          v-for="item in homeHighlights"
          :key="item.id"
          :to="`/news/${item.slug}`"
          class="grid grid-cols-[48px_minmax(0,1fr)] gap-3 px-4 py-3 transition-colors hover:bg-slate-50 md:flex md:items-start md:hover:bg-white/10"
        >
          <img
            v-if="item.cover_image_url"
            :src="item.cover_image_url"
            :alt="item.title"
            class="h-12 w-12 rounded-[10px] border border-slate-100 object-cover"
            loading="lazy"
            decoding="async"
          />
          <div
            v-else
            class="grid h-12 w-12 place-items-center rounded-[10px] border border-[#ffd8d8] bg-[#ffefef] text-primary"
          >
            <span class="material-symbols-outlined text-[1.1rem]">newspaper</span>
          </div>
          <div class="min-w-0 md:flex-1">
            <div class="flex flex-col gap-1 md:flex-row md:items-start md:justify-between md:gap-3">
              <strong class="line-clamp-2 block text-[0.82rem] font-black text-on-surface md:line-clamp-1 md:text-white/90">{{ item.title }}</strong>
              <span class="text-[0.64rem] font-bold text-slate-400 md:flex-shrink-0 md:text-[0.68rem] md:text-white/40">
                {{ item.published_at || item.created_at || '—' }}
              </span>
            </div>
            <span class="mt-1 line-clamp-2 block text-[0.7rem] leading-5 text-slate-500 md:text-white/60">{{ newsPreview(item) }}</span>
          </div>
        </RouterLink>
        <div v-if="!homeHighlights.length && !contentError" class="px-4 py-3 text-[0.78rem] font-semibold text-slate-500">
          Chưa có tin nổi bật.
        </div>
        <div v-if="contentError" class="px-4 py-3 text-[0.78rem] font-semibold text-red-500">
          {{ contentError }}
        </div>
      </div>
    </div>

    <!-- ===== CORPORATE FOOTER ===== -->
    <section class="mx-3 mb-5 rounded-[24px] bg-[#f6ede7] px-4 py-4 shadow-[0_10px_30px_rgba(83,55,44,0.08)]">
      <div class="rounded-[20px] bg-red-950 px-4 py-4 text-white shadow-[0_12px_24px_rgba(218,37,29,0.25)] border border-red-900/50">
        <p class="text-[0.72rem] font-black uppercase tracking-[0.08em] text-yellow-500/90">Thông tin truy cập</p>
        <p class="mt-2 text-[0.88rem] font-semibold leading-7 text-white/92">
          <b>FH88U </b>
Được biết đến là cổng Game giải trí trực tuyến chất lượng dành cho người chơi tại Châu Á. Tại đây hội tụ nhiều sản phẩm nổi bật và nhiều trò chơi hấp dẫn. Đơn vị luôn ưu tiên bảo vệ dữ liệu người dùng, hỗ trợ khách hàng nhanh chóng và không ngừng nâng cao chất lượng dịch vụ để mang lại trải nghiệm tốt nhất. Hãy liên hệ chúng tôi để được hỗ trợ những thắc mắc.
        </p>
        <button
          type="button"
          class="mt-4 inline-flex items-center gap-1.5 rounded-full bg-gradient-to-r from-yellow-400 to-yellow-500 px-4 py-2.5 text-[0.8rem] font-black text-red-900 shadow-[0_10px_20px_rgba(255,204,0,0.25)] transition-transform active:scale-95"
          @click="openTelegram()"
        >
          <span class="material-symbols-outlined text-[1rem]">headset_mic</span>
          Liên hệ CSKH
        </button>
      </div>

      <div class="px-2 pb-1 pt-5 text-center text-[#2b211f]">
        <p class="text-[0.92rem] leading-8">
          Trụ sở chính đặt tại KL The ASPIAL rd No.27, tọa lạc trong khu vực hành chính District Murai Ri.ts, thuộc Commune Bukit Ain.
        </p>
        <p class="mt-4 text-[1rem] leading-8">Quỹ vận hành Coin 1.524.597.982,241 $</p>
        <p class="mt-2 text-[0.92rem] leading-8">
          <span class="font-black uppercase tracking-[0.04em]">Lĩnh vực bảo hộ &amp; hợp tác</span>
          bởi PIE.EXGLN, MACJBELI và ACE-FORCES which are responsible for strategic security, intellectual property protection, and safe operations.
        </p>
      </div>

      <div class="pt-6">
        <h3 class="text-center text-[1.42rem] font-black uppercase tracking-[0.04em] text-[#201816]">
          Nhà bảo hộ &amp; hợp tác
        </h3>
        <div class="mt-5 grid grid-cols-3 gap-3">
          <div
            v-for="supporter in supporterLogos"
            :key="supporter.name"
            class="flex flex-col items-center gap-2"
          >
            <div class="overflow-hidden rounded-[18px] bg-white shadow-[0_8px_18px_rgba(23,199,111,0.12)]">
              <img :src="supporter.image" :alt="supporter.name" class="h-[84px] w-full object-cover" loading="lazy" decoding="async" />
            </div>
            <p class="text-[0.84rem] font-black tracking-[0.04em] text-[#1c1c1c]">{{ supporter.name }}</p>
          </div>
        </div>
        <div class="mt-6 flex items-center justify-center gap-6 border-t border-[#e3d8d0] pt-4 text-[1rem] text-[#2b211f]">
          <span class="font-medium">Trợ giúp</span>
          <span class="font-medium">Quyền riêng tư</span>
          <span class="font-medium">Điều khoản</span>
        </div>
      </div>
    </section>
  </div>
</template>
