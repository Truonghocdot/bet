<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { RouterLink } from 'vue-router'

import BannerCarousel from '@/components/BannerCarousel.vue'
import MarqueeBar from '@/components/MarqueeBar.vue'
import aceForcesLogo from '@/assets/supporter/aceforces.jpg'
import macjbeliLogo from '@/assets/supporter/macjbeli.jpg'
import pieExglnLogo from '@/assets/supporter/pie.exgln.png'
import { request } from '@/shared/api/http'
import type { ContentBannerItem, ContentHomeResponse, ContentNewsItem } from '@/shared/api/types'
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

const greetingName = computed(() => auth.user?.name || 'Bạn')
const vndWallet = computed(() => wallet.wallets.find((item) => item.unit === 1) ?? null)
const homeBanners = ref<ContentBannerItem[]>([])
const homeHighlights = ref<ContentNewsItem[]>([])
const contentError = ref('')

// Maintenance modal
const showMaintenance = ref(false)
const maintenanceGame = ref('')
function openMaintenance(name: string) {
  maintenanceGame.value = name
  showMaintenance.value = true
}
function closeMaintenance() { showMaintenance.value = false }

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

type IdleScheduler = (callback: () => void, options?: { timeout?: number }) => number

const categorySidebar = [
  { label: 'Phổ biến', icon: catHot },
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
  maintenance?: boolean
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

function prefetchPlayRouteSoon() {
  const requestIdleCallback = (window as Window & { requestIdleCallback?: IdleScheduler }).requestIdleCallback

  if (requestIdleCallback) {
    requestIdleCallback(() => prefetchPlayRoute(), { timeout: 2500 })
    return
  }

  window.setTimeout(() => prefetchPlayRoute(), 900)
}

function maybePrefetchGameRoute(game: GameItem) {
  if (game.route) prefetchPlayRoute()
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
        maintenance: true,
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

onMounted(() => {
  prefetchPlayRouteSoon()
  void wallet.fetchSummary()
  void fetchHomeContent()
})
</script>

<template>
  <div class="pb-4">

    <MarqueeBar />
    <BannerCarousel :banners="homeBanners" />

    <!-- ===== MAIN BODY: CATEGORY SIDEBAR + GAME BANNERS ===== -->
    <div class="flex items-start gap-0 pb-1 pt-2.5 md:mx-3 md:mt-3 md:grid md:grid-cols-[124px_minmax(0,1fr)_300px] md:gap-4 md:rounded-[24px] md:bg-black/20 md:p-4 md:shadow-[0_16px_42px_rgba(0,0,0,0.3)] md:backdrop-blur-md md:border md:border-white/10">

      <!-- Category Sidebar -->
      <div
        class="flex w-[68px] flex-shrink-0 flex-col justify-between self-stretch px-1 py-0.5 md:sticky md:top-4 md:w-full md:justify-start md:gap-2 md:rounded-[20px] md:bg-white/10 md:backdrop-blur-sm md:p-2 md:shadow-[0_10px_24px_rgba(0,0,0,0.2)]"
      >
        <button
          v-for="cat in categorySidebar"
          :key="cat.label"
          type="button"
          class="group relative flex flex-col items-center justify-center rounded-[12px] px-0.5 py-1 transition-all duration-200 md:flex-row md:justify-start md:gap-2.5 md:rounded-[14px] md:px-2.5 md:py-2.5"
          :class="activeCategory === cat.label
            ? 'bg-white shadow-[0_3px_12px_rgba(218,37,29,0.20)] ring-1 ring-red-500/20 md:bg-primary/90 md:backdrop-blur md:text-yellow-50 md:ring-1 md:ring-yellow-400/50 md:shadow-[0_4px_15px_rgba(255,204,0,0.25)]'
            : 'bg-transparent md:hover:bg-white/15 text-slate-500 md:text-white/80'"
          @click="activeCategory = cat.label"
        >
          <span
            v-if="activeCategory === cat.label"
            class="absolute left-0 top-1/2 h-5 w-[3px] -translate-y-1/2 rounded-r-full bg-primary md:hidden"
          />
          <div
            class="flex h-8 w-8 items-center justify-center transition-all duration-200 md:h-9 md:w-9 md:flex-shrink-0"
            :class="activeCategory === cat.label ? 'scale-110' : 'scale-95 opacity-70 group-hover:scale-100 group-hover:opacity-90'"
          >
            <img :src="cat.icon" :alt="cat.label" class="w-full h-full object-contain" />
          </div>
          <span
            class="mt-0.5 text-center text-[0.55rem] font-black uppercase leading-tight transition-colors md:mt-0 md:text-left md:text-[0.74rem] md:normal-case md:leading-4"
            :class="activeCategory === cat.label ? 'text-primary md:text-white' : 'text-slate-500 md:text-white/80'"
          >
            {{ cat.label }}
          </span>
        </button>
      </div>

      <!-- Game Banners: 3 banners stacked -->
      <div class="flex min-w-0 flex-1 flex-col gap-2 pl-1 pr-2 md:gap-3 md:p-0">
        <component
          :is="game.route ? RouterLink : 'button'"
          v-for="(game, index) in categoryBannerGames"
          :key="game.name"
          v-bind="game.route ? { to: { path: game.route, query: { from: route.fullPath } } } : { type: 'button' }"
          class="group relative block overflow-hidden rounded-[16px] shadow-[0_4px_16px_rgba(0,0,0,0.12)] transition-all duration-300 active:scale-[0.98] md:rounded-[20px] md:shadow-[0_12px_26px_rgba(15,23,42,0.12)] md:hover:-translate-y-0.5"
          @pointerenter="maybePrefetchGameRoute(game)"
          @focus="maybePrefetchGameRoute(game)"
          @touchstart.passive="maybePrefetchGameRoute(game)"
          @click="game.maintenance ? openMaintenance(game.name) : undefined"
        >
          <img
            :src="game.image"
            :alt="game.name"
            class="block w-full object-cover transition-transform duration-500 ease-out group-hover:scale-[1.03]"
            :class="game.route ? 'aspect-[3/1] md:h-[156px] lg:h-[176px] xl:h-[190px]' : 'aspect-[2.55/1] md:h-[156px] lg:h-[176px] xl:h-[190px]'"
            decoding="async"
            :fetchpriority="index === 0 ? 'high' : 'low'"
            :loading="index === 0 ? 'eager' : 'lazy'"
          />
          <div class="absolute inset-0 bg-gradient-to-t from-black/60 via-black/10 to-transparent" />
          <div class="absolute bottom-0 left-0 right-0 flex items-center justify-between px-3 py-2 md:px-4 md:py-3">
            <div>
              <h4 class="text-[0.78rem] font-black tracking-wide text-white drop-shadow md:text-[1rem]">{{ game.name }}</h4>
              <p class="text-[0.55rem] font-semibold text-white/70 md:text-[0.72rem]">
                {{ game.maintenance ? 'Đang bảo trì' : 'Vào chơi ngay' }}
              </p>
            </div>
            <div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border border-white/30 bg-white/20 backdrop-blur-md md:h-9 md:w-9">
              <span class="material-symbols-outlined text-[0.9rem] text-white md:text-[1.1rem]">arrow_forward</span>
            </div>
          </div>
        </component>
      </div>

      <aside class="hidden min-w-0 flex-col gap-4 md:flex">
        <div class="relative overflow-hidden rounded-[20px] bg-gradient-to-br from-red-700/80 via-red-600/80 to-red-800/80 p-4 text-white shadow-[0_12px_30px_rgba(0,0,0,0.3)] border border-red-400/30 backdrop-blur-md">
          <div class="absolute inset-0 pointer-events-none bg-[radial-gradient(circle_at_top_right,rgba(255,204,0,0.15),transparent_32%)]" />
          <div class="relative">
            <p class="text-[0.7rem] font-black uppercase tracking-[0.12em] text-yellow-100/70">Số dư ví VND</p>
            <strong class="mt-1 block break-words text-[1.55rem] font-black leading-tight text-yellow-400 drop-shadow-sm">
              {{ vndWallet ? displayBalance(vndWallet.balance) : '0' }}đ
            </strong>
            <p class="mt-1 text-[0.72rem] font-semibold text-white/90">Chào {{ greetingName }}</p>
            <div class="mt-4 grid grid-cols-2 gap-2">
              <RouterLink
                to="/account"
                class="flex min-h-10 items-center justify-center gap-1 rounded-[12px] border border-white/35 bg-white/10 px-2 text-[0.75rem] font-black text-white transition-transform active:scale-95"
              >
                Rút tiền
              </RouterLink>
              <RouterLink
                to="/deposit"
                class="flex min-h-10 items-center justify-center gap-1 rounded-[12px] bg-gradient-to-r from-yellow-400 to-yellow-500 px-2 text-[0.75rem] font-black text-red-900 shadow-md transition-transform active:scale-95"
              >
                Nạp tiền
              </RouterLink>
            </div>
          </div>
        </div>

        <div class="rounded-[20px] bg-white/10 backdrop-blur-md border border-white/15 p-4 shadow-[0_10px_24px_rgba(0,0,0,0.15)]">
          <div class="flex items-center justify-between">
            <p class="text-[0.8rem] font-black text-white/90">Danh mục đang xem</p>
            <span class="rounded-full bg-yellow-400/20 px-2 py-1 text-[0.66rem] font-black text-yellow-300">{{ filteredGames.length }} trò</span>
          </div>
          <p class="mt-2 text-[1.25rem] font-black text-white">{{ activeCategory }}</p>
        </div>
      </aside>
    </div>

    <!-- ===== WALLET CARD ===== -->
    <div class="relative mx-3 mt-2 overflow-hidden rounded-[20px] bg-gradient-to-br from-red-700 via-red-600 to-red-800 p-4 text-white shadow-[0_12px_30px_rgba(218,37,29,0.3)] border border-red-500/30 md:hidden">
      <div class="absolute inset-0 pointer-events-none bg-[radial-gradient(circle_at_top_right,rgba(255,204,0,0.15),transparent_26%)]" />
      <div class="relative">
        <p class="text-[0.7rem] uppercase tracking-[0.12em] text-yellow-100/70">Số dư ví VND</p>
        <strong class="mt-1 block text-[1.6rem] font-black text-yellow-400 drop-shadow-sm">
          {{ vndWallet ? displayBalance(vndWallet.balance) : '0' }}đ
        </strong>
        <p class="text-[0.68rem] text-white/90 mt-0.5">Chào {{ greetingName }} 👋</p>
        <div class="mt-3 grid grid-cols-2 gap-2">
          <RouterLink
            to="/account"
            class="flex items-center justify-center gap-1.5 rounded-full border-2 border-white/40 bg-white/10 py-2.5 text-[0.82rem] font-black text-white active:scale-95 transition-transform"
          >
            <span class="material-symbols-outlined text-[1rem]">account_balance</span>
            Rút tiền
          </RouterLink>
          <RouterLink
            to="/deposit"
            class="flex items-center justify-center gap-1.5 rounded-full bg-gradient-to-r from-yellow-400 to-yellow-500 py-2.5 text-[0.82rem] font-black text-red-900 shadow-[0_4px_12px_rgba(255,204,0,0.3)] active:scale-95 transition-transform"
          >
            <span class="material-symbols-outlined text-[1rem]">add_circle</span>
            Nạp tiền
          </RouterLink>
        </div>
      </div>
    </div>

    <!-- ===== GAME GRID (filtered by active category) ===== -->
    <div class="mt-4 px-3 md:mt-5">
      <!-- Section header -->
      <div class="flex items-center justify-between mb-3 md:bg-black/20 md:backdrop-blur-sm md:px-4 md:py-2.5 md:rounded-[14px] md:border md:border-white/10">
        <div class="flex items-center gap-2">
          <span class="w-1 h-5 rounded-full bg-primary md:bg-yellow-400 block" />
          <h2 class="text-[0.92rem] font-black text-on-surface md:text-white">{{ activeCategory }}</h2>
        </div>
        <span class="text-[0.72rem] font-bold text-slate-400 md:text-white/70">{{ filteredGames.length }} trò chơi</span>
      </div>

      <!-- 2-column grid -->
      <div class="grid grid-cols-2 gap-2.5 sm:grid-cols-3 lg:grid-cols-4">
        <component
          :is="game.route ? RouterLink : 'button'"
          v-for="game in filteredGames"
          :key="game.name"
          v-bind="game.route ? { to: { path: game.route, query: { from: route.fullPath } } } : { type: 'button' }"
          class="group relative block w-full overflow-hidden rounded-[16px] text-left shadow-[0_4px_14px_rgba(0,0,0,0.10)] transition-all duration-200 active:scale-[0.97] md:rounded-[18px] md:hover:-translate-y-0.5 md:hover:shadow-[0_14px_26px_rgba(15,23,42,0.12)]"
          @pointerenter="maybePrefetchGameRoute(game)"
          @focus="maybePrefetchGameRoute(game)"
          @touchstart.passive="maybePrefetchGameRoute(game)"
          @click="game.maintenance ? openMaintenance(game.name) : undefined"
        >
          <img
            :src="game.image"
            :alt="game.name"
            class="w-full object-cover transition-transform duration-500 group-hover:scale-[1.04]"
            :class="game.route ? 'aspect-[3/1] md:aspect-[3/1]' : 'aspect-[4/3] md:aspect-[4/3]'"
            loading="lazy"
            decoding="async"
          />
          <!-- Overlay -->
          <div class="absolute inset-0 bg-gradient-to-t from-black/70 via-black/10 to-transparent" />
          <!-- Maintenance badge -->
          <div
            v-if="game.maintenance"
            class="absolute top-2 right-2 bg-amber-500/90 backdrop-blur-sm text-white text-[0.5rem] font-black px-1.5 py-0.5 rounded-full uppercase tracking-wide"
          >
            Bảo trì
          </div>
          <!-- Game name -->
          <div class="absolute bottom-0 left-0 right-0 px-2.5 py-2">
            <p class="text-white text-[0.75rem] font-black drop-shadow line-clamp-1">{{ game.name }}</p>
            <p class="text-white/60 text-[0.55rem] font-semibold">
              {{ game.maintenance ? 'Đang bảo trì' : 'Vào chơi ngay →' }}
            </p>
          </div>
        </component>
      </div>
    </div>

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

    <!-- ===== MAINTENANCE MODAL ===== -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="showMaintenance"
          class="fixed inset-0 z-[95] grid place-items-center bg-black/50 backdrop-blur-sm px-5"
          @click.self="closeMaintenance"
        >
          <div class="w-full max-w-[320px] rounded-[24px] bg-white shadow-[0_24px_60px_rgba(0,0,0,0.25)] overflow-hidden">
            <!-- Top graphic -->
            <div class="bg-gradient-to-br from-red-600 to-red-800 px-6 py-8 flex flex-col items-center gap-3">
              <div class="w-16 h-16 rounded-full bg-white/10 backdrop-blur flex items-center justify-center ring-2 ring-yellow-400/50">
                <span class="material-symbols-outlined text-yellow-400 text-[2.2rem]">construction</span>
              </div>
              <h3 class="text-yellow-400 text-[1.1rem] font-black text-center">Đang Bảo Trì</h3>
            </div>
            <!-- Content -->
            <div class="px-6 py-5 text-center">
              <p class="text-[0.95rem] font-black text-slate-800">{{ maintenanceGame }}</p>
              <p class="mt-2 text-[0.82rem] text-slate-500 leading-6">
                Trò chơi này đang được bảo trì và nâng cấp.<br/>
                Vui lòng quay lại sau ít phút.
              </p>
              <div class="mt-1 inline-flex items-center gap-1.5 bg-amber-50 text-amber-600 text-[0.72rem] font-bold px-3 py-1.5 rounded-full mt-3">
                <span class="material-symbols-outlined text-[0.9rem]">schedule</span>
                Dự kiến hoàn thành sớm
              </div>
              <button
                type="button"
                class="mt-4 w-full rounded-[14px] bg-gradient-to-r from-red-600 to-red-700 py-3 text-[0.88rem] font-black text-yellow-400 shadow-[0_8px_20px_rgba(218,37,29,0.3)] active:scale-95 transition-transform border border-red-500"
                @click="closeMaintenance"
              >
                Đã hiểu
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

  </div>
</template>
