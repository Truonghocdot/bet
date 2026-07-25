<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { RouterLink, type RouteLocationRaw } from 'vue-router'

import BannerCarousel from '@/components/BannerCarousel.vue'
import MarqueeBar from '@/components/MarqueeBar.vue'
import aceForcesLogo from '@/assets/supporter/aceforces.jpg'
import macjbeliLogo from '@/assets/supporter/macjbeli.jpg'
import pieExglnLogo from '@/assets/supporter/pie.exgln.png'
import { request } from '@/shared/api/http'
import type {
  ContentBannerItem,
  ContentHomeResponse,
  ContentNewsItem,
  FakeFinanceFeedItem,
  FakeFinanceFeedResponse,
  ProviderGameCatalogCategory,
  ProviderGameCatalogCategoryItem,
  ProviderGameCatalogResponse,
} from '@/shared/api/types'
import { stripHtmlTags } from '@/shared/lib/html'
import { formatViMoney } from '@/shared/lib/money'
import { useAuthStore } from '@/stores/auth'
import { useWalletStore } from '@/stores/wallet'

import catLottery from '@/assets/category_game/lottery.avif'
import catCasino from '@/assets/category_game/lobbypoker.avif'
import catJackpot from '@/assets/category_game/jackpot.avif'
import catHuntfish from '@/assets/category_game/huntfish.avif'
import catFootball from '@/assets/category_game/football.avif'
import catPoker from '@/assets/category_game/poker.avif'
import catChicken from '@/assets/category_game/chicken.avif'
import catHot from '@/assets/category_game/hot.avif'

// Main Game Banners
import bannerWingo from '@/assets/lottery_banner/optimized/wingo.webp'
import bannerK3 from '@/assets/lottery_banner/optimized/k3.webp'
import banner5D from '@/assets/lottery_banner/optimized/5d.webp'
import chickenSV128 from '@/assets/game_thumbs/chicken/SV128.avif'
import fishBBIN from '@/assets/game_thumbs/hunt-fish/BBIN.avif'
import fishCQ9 from '@/assets/game_thumbs/hunt-fish/CQ9.avif'
import fishFC from '@/assets/game_thumbs/hunt-fish/FC.avif'
import fishJDB from '@/assets/game_thumbs/hunt-fish/JDB.avif'
import fishJili from '@/assets/game_thumbs/hunt-fish/JILI.avif'
import fishMG from '@/assets/game_thumbs/hunt-fish/MG.avif'
import fishRSG from '@/assets/game_thumbs/hunt-fish/RSG.avif'
import fishTP from '@/assets/game_thumbs/hunt-fish/TP.avif'
import fishWG from '@/assets/game_thumbs/hunt-fish/WG.avif'
import pokerJili from '@/assets/game_thumbs/poker/JILI.avif'
import pokerMG from '@/assets/game_thumbs/poker/MG.avif'
import pokerWG from '@/assets/game_thumbs/poker/WG.avif'
import slot168Game from '@/assets/game_thumbs/slot/168GAME.avif'
import slotCQ9 from '@/assets/game_thumbs/slot/CQ9.avif'
import slotFC from '@/assets/game_thumbs/slot/FC.avif'
import slotJDB from '@/assets/game_thumbs/slot/JDB.avif'
import slotJili from '@/assets/game_thumbs/slot/JILI.avif'
import slotPG from '@/assets/game_thumbs/slot/PG.avif'
import slotTP from '@/assets/game_thumbs/slot/TP.avif'
import slotWG from '@/assets/game_thumbs/slot/WG.avif'
import sportCMD from '@/assets/game_thumbs/sport/CMD.avif'
import sportCrown from '@/assets/game_thumbs/sport/CROWN.avif'
import sportIM from '@/assets/game_thumbs/sport/IM.avif'
import sportPanda from '@/assets/game_thumbs/sport/PANDA.avif'
import sportSaba from '@/assets/game_thumbs/sport/SABA.avif'

// Game Thumbnails - supports both .webp and .avif
const gameThumbModules = import.meta.glob<string>('@/assets/game_thumbs/*/*.{webp,avif}', {
  eager: true,
  query: '?url',
  import: 'default',
})

const auth = useAuthStore()
const wallet = useWalletStore()
const route = useRoute()
const router = useRouter()
const popupVideoUrl = `${import.meta.env.VITE_API_BASE_URL}/v1/media/popup-video`

const greetingName = computed(() => auth.user?.name || 'Bạn')
const vndWallet = computed(() => wallet.wallets.find((item) => item.unit === 1) ?? null)
const homeBanners = ref<ContentBannerItem[]>([])
const homeHighlights = ref<ContentNewsItem[]>([])
const contentError = ref('')
const fakeDepositFeed = ref<FakeFinanceFeedItem[]>([])
const fakeWithdrawFeed = ref<FakeFinanceFeedItem[]>([])
const providerCatalogCategories = ref<ProviderGameCatalogCategory[]>([])
const providerCatalogError = ref('')
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
let providerLobbyRoutePrefetched = false
const recentGameKeys = ref<string[]>([])
const favoriteGameKeys = ref<string[]>([])

type IdleScheduler = (callback: () => void, options?: { timeout?: number }) => number
type HomeCategory = 'Xổ số' | 'Casino' | 'Nổ hũ' | 'Bắn cá' | 'Thể thao' | 'Game bài' | 'Đá gà'
type MobileLobbyKey = 'hot' | 'lottery' | 'slots' | 'sports' | 'casino' | 'fish' | 'cards' | 'cockfight'

const categorySidebar = [
  { label: 'Phổ biến', icon: catHot },
  { label: 'Xổ số', icon: catLottery },
  { label: 'Nổ hũ', icon: catJackpot },
  { label: 'Thể thao', icon: catFootball },
  { label: 'Casino', icon: catCasino },
  { label: 'Bắn cá', icon: catHuntfish },
  { label: 'Game bài', icon: catPoker },
  { label: 'Đá gà', icon: catChicken },
]

const mobileLobbySidebar: Array<{ key: MobileLobbyKey; label: string; icon: string; category: string }> = [
  { key: 'hot', label: 'Hot', icon: catHot, category: 'Phổ biến' },
  { key: 'lottery', label: 'Xổ Số', icon: catLottery, category: 'Xổ số' },
  { key: 'slots', label: 'Nổ Hũ', icon: catJackpot, category: 'Nổ hũ' },
  { key: 'sports', label: 'Thể Thao', icon: catFootball, category: 'Thể thao' },
  { key: 'casino', label: 'Casino', icon: catCasino, category: 'Casino' },
  { key: 'fish', label: 'Bắn Cá', icon: catHuntfish, category: 'Bắn cá' },
  { key: 'cards', label: 'Game Bài', icon: catPoker, category: 'Game bài' },
  { key: 'cockfight', label: 'Đá Gà', icon: catChicken, category: 'Đá gà' },
]
const activeMobileLobbyKey = ref<MobileLobbyKey>('hot')
const activeMobileLobbyTab = ref<'category' | 'history' | 'favorite'>('category')

interface GameItem {
  name: string
  image: string
  category: string[]
  route?: string
  partnerLobby?: boolean
  featured?: boolean
  wideCard?: boolean
  providerItemKind?: string
  providerCategoryKey?: string
  providerChildCount?: number
  providerGameCode?: string
  providerProductType?: string
  providerProductTypeNumber?: number
  providerGameType?: string
}

const recentGamesStorageKey = 'fh88u:home:recent-games:v1'
const favoriteGamesStorageKey = 'fh88u:home:favorite-games:v1'

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

function prefetchProviderLobbyRoute() {
  if (providerLobbyRoutePrefetched) return

  providerLobbyRoutePrefetched = true
  void import('@/pages/ProviderGameLobbyListView.vue').catch(() => {
    providerLobbyRoutePrefetched = false
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

  if (game.providerItemKind === 'group') {
    prefetchProviderLobbyRoute()
    return
  }

  if (game.partnerLobby) {
    prefetchPartnerGameRoute()
  }
}

function providerLobbyDisplayName(game: GameItem): string {
  if (game.providerCategoryKey === 'casino' && game.name === 'AE') return 'SEXY'
  return game.name
}

function resolveGameTarget(game: GameItem): RouteLocationRaw {
  if (game.route) {
    return { path: game.route, query: { from: route.fullPath } }
  }

  if (game.providerItemKind === 'group') {
    const productType = game.providerProductTypeNumber ?? parseProviderProductType(game.providerProductType || '')
    if (!game.providerCategoryKey || productType <= 0) {
      return { path: route.path }
    }

    return {
      name: 'provider-game-lobbies',
      params: {
        category: game.providerCategoryKey,
        productType,
      },
      query: {
        name: providerLobbyDisplayName(game),
        from: route.fullPath,
        provider_key: providerLobbyDisplayName(game),
      },
    }
  }

  if (!game.providerGameCode || !game.providerProductType || !game.providerGameType) {
    return { path: route.path }
  }

  return {
    name: 'partner-game-loading',
    query: {
      name: game.name,
      from: route.fullPath,
      game_code: game.providerGameCode,
      product_type: game.providerProductType,
      game_type: game.providerGameType,
    },
  }
}

type ThumbCategoryConfig = {
  category: string
  popular?: boolean
  featuredFirst?: boolean
}

type ProviderThumbPreset = {
  code: string
  label: string
  productType: number
  image: string
}

// Mapping tên file thumbnail lobby-casino → productType của TCG API
const lobbyCasinoProductTypeMap: Record<string, number> = {
  'AE':       112,
  'AG':       4,
  'BBIN':     79,
  'DG':       27,
  'EZUGI':    177,
  'HRG':      93,
  'MG':       43,
  'MT':       272,
  'PP':       39,
  'Playtech': 3,
  'SA':       93,
  'TP':       93,
  'W':        258,
  'WM':       118,
}

const thumbCategoryMap: Record<string, ThumbCategoryConfig> = {
  'lobby-casino': { category: 'Casino', popular: true, featuredFirst: true },
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
      const baseName = fileName.replace(/\.[^.]+$/, '')

      // Các thumbnail lobby-casino được navigate thẳng tới trang danh sách lobby
      if (folder === 'lobby-casino') {
        const productType = lobbyCasinoProductTypeMap[baseName] ?? 0
        return {
          name: baseName,
          image,
          category: [
            ...(config.popular && index < 2 ? ['Phổ biến'] : []),
            config.category,
          ],
          partnerLobby: false,
          featured: Boolean(config.featuredFirst && index === 0),
          wideCard: true,
          providerItemKind: 'group',
          providerCategoryKey: 'casino',
          providerProductType: String(productType),
          providerProductTypeNumber: productType,
          providerGameCode: baseName,
          providerGameType: 'LIVE',
        }
      }

      return {
        name: prettyGameThumbName(fileName),
        image,
        category: [
          ...(config.popular && index < 2 ? ['Phổ biến'] : []),
          config.category,
        ],
        partnerLobby: true,
        featured: Boolean(config.featuredFirst && index === 0),
        wideCard: folder === 'chicken',
      }
    })
}

const thumbGames = Object.entries(thumbCategoryMap).flatMap(([folder, config]) => buildThumbGames(folder, config))

const localLotteryGames: GameItem[] = [
  // Xổ số - có route thật
  { name: 'Win Go', image: bannerWingo, category: ['Phổ biến', 'Xổ số'], route: '/play/wingo', featured: true },
  { name: 'K3', image: bannerK3, category: ['Phổ biến', 'Xổ số'], route: '/play/k3' },
  { name: '5D Lottery', image: banner5D, category: ['Phổ biến', 'Xổ số'], route: '/play/lottery' },
]
const defaultLobbyGames = localLotteryGames.slice(0, 3)
const categoryGameLimit = 15

const sportsProviderPresets: ProviderThumbPreset[] = [
  { code: 'SABA', label: 'SABA', productType: 174, image: sportSaba },
  { code: 'CMD', label: 'CMD', productType: 104, image: sportCMD },
  { code: 'PANDA', label: 'PANDA', productType: 131, image: sportPanda },
  { code: 'CROWN', label: 'CROWN', productType: 65, image: sportCrown },
  { code: 'IM', label: 'IM', productType: 68, image: sportIM },
]

const slotProviderPresets: ProviderThumbPreset[] = [
  { code: '168GAME', label: '168GAME', productType: 275, image: slot168Game },
  { code: 'CQ9', label: 'CQ9', productType: 16, image: slotCQ9 },
  { code: 'FC', label: 'FC', productType: 141, image: slotFC },
  { code: 'JDB', label: 'JDB', productType: 55, image: slotJDB },
  { code: 'JILI', label: 'JILI', productType: 140, image: slotJili },
  { code: 'PG', label: 'PG', productType: 98, image: slotPG },
  { code: 'TP', label: 'TP', productType: 243, image: slotTP },
  { code: 'WG', label: 'WG', productType: 212, image: slotWG },
]

const fishProviderPresets: ProviderThumbPreset[] = [
  { code: 'BBIN', label: 'BBIN', productType: 79, image: fishBBIN },
  { code: 'CQ9', label: 'CQ9', productType: 16, image: fishCQ9 },
  { code: 'FC', label: 'FC', productType: 141, image: fishFC },
  { code: 'JDB', label: 'JDB', productType: 55, image: fishJDB },
  { code: 'JILI', label: 'JILI', productType: 140, image: fishJili },
  { code: 'MG', label: 'MG', productType: 43, image: fishMG },
  { code: 'RSG', label: 'RSG', productType: 138, image: fishRSG },
  { code: 'TP', label: 'TP', productType: 243, image: fishTP },
  { code: 'WG', label: 'WG', productType: 212, image: fishWG },
]

const pokerProviderPresets: ProviderThumbPreset[] = [
  { code: 'JILI', label: 'JILI', productType: 140, image: pokerJili },
  { code: 'MG', label: 'MG', productType: 43, image: pokerMG },
  { code: 'WG', label: 'WG', productType: 212, image: pokerWG },
]

const homeCategories: HomeCategory[] = ['Xổ số', 'Casino', 'Nổ hũ', 'Bắn cá', 'Thể thao', 'Game bài', 'Đá gà']

const categoryPriority: Record<HomeCategory, number[]> = {
  'Casino': [4, 3, 79, 27, 177, 43, 39, 93, 258, 118, 272],
  'Nổ hũ': [98, 16, 4, 3, 39, 13, 171],
  'Bắn cá': [79, 16, 141, 55, 140, 43, 138, 243, 212],
  'Game bài': [140, 43, 212],
  'Thể thao': [104, 65, 68, 131, 174],
  'Đá gà': [202, 132],
  'Xổ số': [2, 384, 420, 64, 76],
}

function parseProviderProductType(value: string): number {
  const parsed = Number.parseInt(String(value || '').trim(), 10)
  return Number.isFinite(parsed) ? parsed : 0
}

function providerCatalogItemProductType(item: ProviderGameCatalogCategoryItem): number {
  const directProductType = Number(item.product_type ?? 0)
  if (directProductType > 0) return directProductType
  return parseProviderProductType(String(item.product_type_value || '').trim())
}

const providerCategoryKeys: Record<HomeCategory, string> = {
  'Xổ số': 'lottery',
  'Casino': 'casino',
  'Nổ hũ': 'slots',
  'Bắn cá': 'fish',
  'Thể thao': 'sports',
  'Game bài': 'cards',
  'Đá gà': 'cockfight',
}

function supportsCatalogPlatform(platformValue: string): boolean {
  const platform = String(platformValue || '').trim().toLowerCase()
  if (!platform) return false

  return platform.includes('html5') || platform.includes('desktop') || platform.includes('web') || platform.includes('mobile')
}

function providerProductPriority(category: HomeCategory, productType: number): number {
  const priority = categoryPriority[category] || []
  const index = priority.indexOf(productType)
  return index === -1 ? Number.MAX_SAFE_INTEGER : index
}

function buildProviderGameItem(item: ProviderGameCatalogCategoryItem, category: HomeCategory, categoryKey: string): GameItem {
  const productType = providerCatalogItemProductType(item)
  return {
    name: String(item.game_name || '').trim() || String(item.tcg_game_code || 'TCG Game').trim(),
    image: String(item.show_icon || '').trim(),
    category: [category],
    partnerLobby: String(item.kind || '').trim() !== 'group',
    wideCard: false,
    providerGameCode: String(item.tcg_game_code || '').trim(),
    providerProductType: String(item.product_type_value || item.product_type || '').trim(),
    providerProductTypeNumber: productType,
    providerGameType: String(item.game_type || '').trim(),
    providerItemKind: String(item.kind || '').trim(),
    providerCategoryKey: categoryKey,
    providerChildCount: Number(item.child_count ?? 0),
  }
}

function buildSportsProviderGameItem(
  item: ProviderGameCatalogCategoryItem | undefined,
  categoryKey: string,
  preset: ProviderThumbPreset,
): GameItem {
  return {
    name: preset.label,
    image: preset.image,
    category: ['Thể thao'],
    partnerLobby: true,
    wideCard: true,
    providerGameCode: item ? String(item.tcg_game_code || '').trim() : preset.code,
    providerProductType: String(preset.productType),
    providerProductTypeNumber: preset.productType,
    providerGameType: item ? String(item.game_type || '').trim() : 'SPORTS',
    providerItemKind: 'launch',
    providerCategoryKey: categoryKey,
    providerChildCount: 0,
  }
}

function slotProviderPresetKeywords(preset: ProviderThumbPreset): string[] {
  if (preset.code === '168GAME') return ['168game', '168 game']
  if (preset.code === 'JILI') return ['jili', 'jl']
  if (preset.code === 'WG') return ['swg', 'swgaming']
  return [preset.code.toLowerCase()]
}

function matchesProviderCatalogKeyword(item: ProviderGameCatalogCategoryItem, keywords: string[]): boolean {
  const values = [
    String(item.game_name || '').trim().toLowerCase(),
    String(item.tcg_game_code || '').trim().toLowerCase(),
    String(item.product_code || '').trim().toLowerCase(),
    String(item.game_sub_type || '').trim().toLowerCase(),
  ].filter(Boolean)

  return keywords.some((keyword) => {
    const normalizedKeyword = String(keyword || '').trim().toLowerCase()
    return normalizedKeyword !== '' && values.some((value) => value.includes(normalizedKeyword))
  })
}

function slotProviderLaunchRank(item: ProviderGameCatalogCategoryItem, preset: ProviderThumbPreset): number {
  let score = 0

  if (matchesProviderCatalogKeyword(item, slotProviderPresetKeywords(preset))) score += 8
  if (matchesProviderCatalogKeyword(item, ['game_list', 'game list', 'lobby'])) score += 6
  if (matchesProviderCatalogKeyword(item, ['slot', 'slots', 'rng'])) score += 3
  if (String(item.show_icon || '').trim()) score += 1

  return score
}

function pickSlotProviderLaunchItem(
  items: ProviderGameCatalogCategoryItem[],
  preset: ProviderThumbPreset,
): ProviderGameCatalogCategoryItem | undefined {
  const launchItems = items.filter((item) => String(item.kind || '').trim() === 'launch')
  const exactProductTypeItems = launchItems.filter((item) => providerCatalogItemProductType(item) === preset.productType)
  const aliasMatchedItems = launchItems.filter((item) => matchesProviderCatalogKeyword(item, slotProviderPresetKeywords(preset)))
  const candidates = exactProductTypeItems.length > 0 ? exactProductTypeItems : aliasMatchedItems

  if (candidates.length === 0) return undefined

  return [...candidates].sort((left, right) => {
    const priorityDiff = slotProviderLaunchRank(right, preset) - slotProviderLaunchRank(left, preset)
    if (priorityDiff !== 0) return priorityDiff

    const leftName = String(left.game_name || left.tcg_game_code || '').trim()
    const rightName = String(right.game_name || right.tcg_game_code || '').trim()
    return leftName.localeCompare(rightName, 'vi')
  })[0]
}

function buildSlotProviderGameItem(
  item: ProviderGameCatalogCategoryItem | undefined,
  categoryKey: string,
  preset: ProviderThumbPreset,
  index: number,
): GameItem {
  const productType = item ? providerCatalogItemProductType(item) : preset.productType

  return {
    name: preset.label,
    image: preset.image,
    category: [
      ...(index < 2 ? ['Phổ biến'] : []),
      'Nổ hũ',
    ],
    partnerLobby: false,
    featured: index === 0,
    wideCard: true,
    providerGameCode: preset.code,
    providerProductType: item ? String(item.product_type_value || item.product_type || '').trim() : String(productType),
    providerProductTypeNumber: productType,
    providerGameType: item ? String(item.game_type || '').trim() : 'RNG',
    providerItemKind: 'group',
    providerCategoryKey: categoryKey,
    providerChildCount: item ? 1 : 0,
  }
}

function buildFishProviderGameItem(
  item: ProviderGameCatalogCategoryItem | undefined,
  categoryKey: string,
  preset: ProviderThumbPreset,
  index: number,
): GameItem {
  const productType = item ? providerCatalogItemProductType(item) : preset.productType

  return {
    name: preset.label,
    image: preset.image,
    category: [
      ...(index < 2 ? ['Phổ biến'] : []),
      'Bắn cá',
    ],
    partnerLobby: false,
    featured: index === 0,
    wideCard: true,
    providerGameCode: preset.code,
    providerProductType: item ? String(item.product_type_value || item.product_type || '').trim() : String(productType),
    providerProductTypeNumber: productType,
    providerGameType: item ? String(item.game_type || '').trim() : 'FISH',
    providerItemKind: 'group',
    providerCategoryKey: categoryKey,
    providerChildCount: Number(item?.child_count ?? 0),
  }
}

function buildPokerProviderGameItem(
  item: ProviderGameCatalogCategoryItem | undefined,
  categoryKey: string,
  preset: ProviderThumbPreset,
  index: number,
): GameItem {
  const productType = item ? providerCatalogItemProductType(item) : preset.productType

  return {
    name: preset.label,
    image: preset.image,
    category: [
      ...(index < 2 ? ['Phổ biến'] : []),
      'Game bài',
    ],
    partnerLobby: false,
    featured: index === 0,
    wideCard: true,
    providerGameCode: preset.code,
    providerProductType: item ? String(item.product_type_value || item.product_type || '').trim() : String(productType),
    providerProductTypeNumber: productType,
    providerGameType: item ? String(item.game_type || '').trim() : 'PVP',
    providerItemKind: 'group',
    providerCategoryKey: categoryKey,
    providerChildCount: Number(item?.child_count ?? 0),
  }
}

function cockfightLaunchRank(item: ProviderGameCatalogCategoryItem): number {
  let score = 0

  const productType = providerCatalogItemProductType(item)
  const productPriority = providerProductPriority('Đá gà', productType)
  if (productPriority !== Number.MAX_SAFE_INTEGER) {
    score += 10 - productPriority
  }

  if (matchesProviderCatalogKeyword(item, ['cockfight', 'ws168', 'sv388', 'sv128', 'fight', 'arena', 'lobby'])) score += 6
  if (String(item.show_icon || '').trim()) score += 1

  return score
}

function pickCockfightLaunchItem(items: ProviderGameCatalogCategoryItem[]): ProviderGameCatalogCategoryItem | undefined {
  const launchItems = items.filter((item) => String(item.kind || '').trim() === 'launch')
  if (launchItems.length === 0) return undefined

  return [...launchItems].sort((left, right) => {
    const rankDiff = cockfightLaunchRank(right) - cockfightLaunchRank(left)
    if (rankDiff !== 0) return rankDiff

    const leftPriority = providerProductPriority('Đá gà', providerCatalogItemProductType(left))
    const rightPriority = providerProductPriority('Đá gà', providerCatalogItemProductType(right))
    if (leftPriority !== rightPriority) return leftPriority - rightPriority

    return String(left.game_name || left.tcg_game_code || '').trim()
      .localeCompare(String(right.game_name || right.tcg_game_code || '').trim(), 'vi')
  })[0]
}

function buildCockfightProviderGameItem(
  item: ProviderGameCatalogCategoryItem,
  categoryKey: string,
): GameItem {
  return {
    name: 'Đá gà',
    image: chickenSV128,
    category: ['Đá gà'],
    partnerLobby: true,
    featured: true,
    wideCard: true,
    providerGameCode: String(item.tcg_game_code || '').trim(),
    providerProductType: String(item.product_type_value || item.product_type || '').trim(),
    providerProductTypeNumber: providerCatalogItemProductType(item),
    providerGameType: String(item.game_type || '').trim() || 'SPORTS',
    providerItemKind: 'launch',
    providerCategoryKey: categoryKey,
    providerChildCount: 0,
  }
}

function gameStorageKey(game: GameItem): string {
  return [
    game.providerProductType || '',
    game.providerGameCode || '',
    game.route || '',
    game.name.trim(),
  ].join(':')
}

function loadStoredGameKeys(storageKey: string): string[] {
  try {
    const raw = window.localStorage.getItem(storageKey)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed)
      ? parsed.map((item) => String(item || '').trim()).filter(Boolean)
      : []
  } catch {
    return []
  }
}

function persistStoredGameKeys(storageKey: string, items: string[]) {
  try {
    window.localStorage.setItem(storageKey, JSON.stringify(items))
  } catch {
    // ignore storage errors
  }
}

function trackRecentGame(game: GameItem) {
  const key = gameStorageKey(game)
  const next = [key, ...recentGameKeys.value.filter((item) => item !== key)].slice(0, 24)
  recentGameKeys.value = next
  persistStoredGameKeys(recentGamesStorageKey, next)
}

function toggleFavoriteGame(game: GameItem) {
  const key = gameStorageKey(game)
  const exists = favoriteGameKeys.value.includes(key)
  const next = exists
    ? favoriteGameKeys.value.filter((item) => item !== key)
    : [key, ...favoriteGameKeys.value].slice(0, 36)
  favoriteGameKeys.value = next
  persistStoredGameKeys(favoriteGamesStorageKey, next)
}

function isFavoriteGame(game: GameItem): boolean {
  return favoriteGameKeys.value.includes(gameStorageKey(game))
}

function selectMobileLobbyItem(key: MobileLobbyKey) {
  const item = mobileLobbySidebar.find((entry) => entry.key === key)
  if (!item) return
  activeMobileLobbyKey.value = key
  activeCategory.value = item.category
  activeMobileLobbyTab.value = 'category'
  void router.replace({ query: { ...route.query, tab: key } })
}

function switchLobbyTab(tab: 'category' | 'history' | 'favorite') {
  activeMobileLobbyTab.value = tab
  void router.replace({ query: { ...route.query, subtab: tab === 'category' ? undefined : tab } })
}

function syncTabFromRoute() {
  const tabParam = String(route.query.tab ?? '').trim() as MobileLobbyKey
  const subtabParam = String(route.query.subtab ?? '').trim()

  if (tabParam && mobileLobbySidebar.some((e) => e.key === tabParam)) {
    const item = mobileLobbySidebar.find((e) => e.key === tabParam)!
    activeMobileLobbyKey.value = tabParam
    activeCategory.value = item.category
  }

  if (subtabParam === 'history' || subtabParam === 'favorite') {
    activeMobileLobbyTab.value = subtabParam
  } else {
    activeMobileLobbyTab.value = 'category'
  }
}

function mobileLobbyBadgeText(game: GameItem, index: number): string {
  if (game.route) return 'Live'
  if (game.providerItemKind === 'group') return 'Lobby'
  if (game.featured || index < 3) return 'Hot'
  return ''
}

function isDefaultLobbyGame(game: GameItem): boolean {
  return defaultLobbyGames.some((item) => item.route === game.route && item.name === game.name)
}

function isWideLobbyGame(game: GameItem): boolean {
  return isDefaultLobbyGame(game) || Boolean(game.wideCard)
}

function matchesSelectedCategory(game: GameItem): boolean {
  if (activeCategory.value === 'Phổ biến') return true
  return game.category.includes(activeCategory.value)
}

function uniqueGames(items: GameItem[]): GameItem[] {
  const seen = new Set<string>()
  return items.filter((item) => {
    const key = [
      item.providerProductType || '',
      item.providerGameCode || '',
      item.name,
      item.category.join('|'),
    ].join(':')
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}

const providerGamesByCategory = computed<Record<HomeCategory, GameItem[]>>(() => {
  const grouped: Record<HomeCategory, GameItem[]> = {
    'Xổ số': [],
    'Casino': [],
    'Nổ hũ': [],
    'Bắn cá': [],
    'Thể thao': [],
    'Game bài': [],
    'Đá gà': [],
  }

  for (const category of homeCategories) {
    const categoryKey = providerCategoryKeys[category]
    const section = providerCatalogCategories.value.find((item) => String(item.key || '').trim().toLowerCase() === categoryKey)
    const categoryItems = (section?.items ?? [])
      .filter((item) => Number(item.display_status ?? 0) === 0)
      .filter((item) => supportsCatalogPlatform(item.platform))

    if (category === 'Thể thao') {
      grouped[category] = sportsProviderPresets.map((preset) => {
        const matched = categoryItems.find((item) => (
          providerCatalogItemProductType(item) === preset.productType && String(item.kind || '').trim() === 'launch'
        ))
        return buildSportsProviderGameItem(matched, categoryKey, preset)
      })
      continue
    }

    // Nổ hũ: giữ giao diện local thumb mặc định, còn metadata launch lấy từ catalog nếu có.
    if (category === 'Nổ hũ') {
      grouped[category] = slotProviderPresets.map((preset, index) => {
        const matched = pickSlotProviderLaunchItem(categoryItems, preset)
        return buildSlotProviderGameItem(matched, categoryKey, preset, index)
      })
      continue
    }

    if (category === 'Bắn cá') {
      grouped[category] = fishProviderPresets.map((preset, index) => {
        const matched = categoryItems.find((item) => providerCatalogItemProductType(item) === preset.productType)
        return buildFishProviderGameItem(matched, categoryKey, preset, index)
      })
      continue
    }

    if (category === 'Game bài') {
      grouped[category] = pokerProviderPresets.map((preset, index) => {
        const matched = categoryItems.find((item) => providerCatalogItemProductType(item) === preset.productType)
        return buildPokerProviderGameItem(matched, categoryKey, preset, index)
      })
      continue
    }

    if (category === 'Đá gà') {
      const matched = pickCockfightLaunchItem(categoryItems)
      grouped[category] = matched ? [buildCockfightProviderGameItem(matched, categoryKey)] : []
      continue
    }

    // Casino: luôn dùng local thumbnails (giống Sports), mỗi thumbnail điều hướng tới trang lobby của nhà của
    if (category === 'Casino') {
      const casinoThumbs = thumbGames.filter((g) => g.category.includes('Casino'))
      // Chỉ giữ lại các item có productType hợp lệ (> 0)
      grouped[category] = casinoThumbs.filter((g) => (g.providerProductTypeNumber ?? 0) > 0)
      continue
    }

    const sourceItems = categoryItems
      .filter((item) => String(item.show_icon || '').trim())
      .map((item) => buildProviderGameItem(item, category, categoryKey))

    grouped[category] = uniqueGames(sourceItems).sort((left, right) => {
      const leftPriority = providerProductPriority(category, parseProviderProductType(left.providerProductType || ''))
      const rightPriority = providerProductPriority(category, parseProviderProductType(right.providerProductType || ''))
      if (leftPriority !== rightPriority) return leftPriority - rightPriority
      return left.name.localeCompare(right.name, 'vi')
    })
  }

  return grouped
})

const categoriesWithProviderData = computed(() => {
  const categories = new Set<string>()
  for (const category of homeCategories) {
    const items = providerGamesByCategory.value[category]
    if (items.length > 0) categories.add(category)
  }
  return categories
})

const fallbackThumbGames = computed(() => (
  thumbGames.filter((game) => game.category.every((category) => !categoriesWithProviderData.value.has(category)))
))

function fallbackGamesForCategory(category: HomeCategory): GameItem[] {
  return fallbackThumbGames.value.filter((game) => game.category.includes(category))
}

function normalizeCategoryGames(category: HomeCategory, items: GameItem[]): GameItem[] {
  return items.slice(0, categoryGameLimit)
}

function resolvedCategoryGames(category: HomeCategory): GameItem[] {
  const providerItems = providerGamesByCategory.value[category]
  const source = providerItems.length > 0 ? providerItems : fallbackGamesForCategory(category)
  return normalizeCategoryGames(category, source)
}

const filteredGames = computed(() => {
  const category = activeCategory.value as HomeCategory
  if (activeCategory.value === 'Phổ biến' || category === 'Xổ số') {
    return defaultLobbyGames
  }

  return resolvedCategoryGames(category)
})

const categoryBannerGames = computed(() => (
  activeCategory.value === 'Phổ biến'
    ? defaultLobbyGames
    : filteredGames.value.slice(0, 3)
))

const activeMobileLobbyItem = computed(() => (
  mobileLobbySidebar.find((item) => item.key === activeMobileLobbyKey.value) ?? mobileLobbySidebar[0]
))

const allKnownGames = computed(() => {
  const games: GameItem[] = []
  games.push(...localLotteryGames)
  games.push(...thumbGames)
  for (const cat of homeCategories) {
    games.push(...(providerGamesByCategory.value[cat] || []))
  }
  return uniqueGames(games)
})

const mobileLobbyGames = computed(() => {
  if (activeMobileLobbyTab.value === 'history') {
    return recentGameKeys.value
      .map((key) => allKnownGames.value.find((g) => gameStorageKey(g) === key))
      .filter((g): g is GameItem => Boolean(g))
      .filter(matchesSelectedCategory)
  }

  if (activeMobileLobbyTab.value === 'favorite') {
    return favoriteGameKeys.value
      .map((key) => allKnownGames.value.find((g) => gameStorageKey(g) === key))
      .filter((g): g is GameItem => Boolean(g))
      .filter(matchesSelectedCategory)
  }

  return filteredGames.value
})

const mobileLobbyEmptyMessage = computed(() => {
  if (activeMobileLobbyTab.value === 'history') return 'Bạn chưa chơi trò nào gần đây.'
  if (activeMobileLobbyTab.value === 'favorite') return 'Bạn chưa có trò chơi yêu thích nào.'
  return providerCatalogError.value || 'Danh mục này đang được cập nhật.'
})

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

async function fetchProviderGames() {
  providerCatalogError.value = ''
  try {
    const response = await request<ProviderGameCatalogResponse>('GET', '/v1/provider-games/tcg')
    providerCatalogCategories.value = response.categories || []
  } catch {
    providerCatalogCategories.value = []
    providerCatalogError.value = 'Không thể tải danh sách game nhà cung cấp'
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
  recentGameKeys.value = loadStoredGameKeys(recentGamesStorageKey)
  favoriteGameKeys.value = loadStoredGameKeys(favoriteGamesStorageKey)
  syncTabFromRoute()
  prefetchPlayRouteSoon()
  window.setTimeout(() => prefetchPartnerGameRoute(), 1200)
  window.setTimeout(() => prefetchProviderLobbyRoute(), 1500)
  void wallet.fetchSummary()
  void fetchHomeContent()
  void fetchProviderGames()
  void refreshFakeFinanceFeed()
  startFakeFinancePolling()
})

watch(
  () => route.query,
  () => syncTabFromRoute(),
)

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

    <div class="mt-4 bg-[linear-gradient(180deg,#303133_0%,#292a2c_24%,#292a2c_100%)] pb-4 md:mt-0">
      <div class="px-3 pt-2">
        <MarqueeBar />
      </div>

      <div class="px-3 pt-2">
        <div class="overflow-hidden rounded-[26px] border border-white/10 bg-surface shadow-[0_10px_30px_rgba(0,0,0,0.24)]">
          <BannerCarousel :banners="homeBanners" />
        </div>
      </div>

      <section class="mx-3 mt-3 overflow-hidden rounded-[24px] border border-primary/35 bg-gradient-to-br from-[#4b4c4f] via-[#38393b] to-[#2b2c2e] p-4 text-white shadow-[0_14px_30px_rgba(0,0,0,0.32)]">
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
            class="flex min-h-12 items-center justify-center rounded-[14px] bg-gradient-to-r from-primary to-[#ffc928] px-3 text-[0.9rem] font-black text-[#241900] shadow-[inset_0_-3px_0_rgba(0,0,0,0.16)] transition-transform active:scale-95"
          >
            Nạp tiền
          </RouterLink>
        </div>
      </section>

    <section class="home-lobby-mobile relative overflow-hidden mt-4">
      <div class="home-lobby-mobile__shell relative min-h-[calc(100dvh-11.5rem)] px-2 pb-4 pt-3">
        <aside class="home-lobby-mobile__sidebar sticky top-3 self-start">
          <div class="space-y-1.5">
            <button
              v-for="item in mobileLobbySidebar"
              :key="item.key"
              type="button"
              class="home-lobby-mobile__nav-item flex w-full flex-col items-center justify-center gap-1 rounded-[14px] px-1.5 py-2.5 text-center transition-all duration-200"
              :class="activeMobileLobbyKey === item.key ? 'bg-primary text-[#241900] shadow-[0_8px_16px_rgba(242,173,0,0.24)]' : 'bg-transparent text-slate-300'"
              @click="selectMobileLobbyItem(item.key)"
            >
              <span class="grid h-10 w-10 shrink-0 place-items-center overflow-hidden rounded-[12px]" :class="activeMobileLobbyKey === item.key ? 'bg-white/20' : 'bg-transparent'">
                <img :src="item.icon" :alt="item.label" class="h-8 w-8 object-contain" loading="lazy" decoding="async" />
              </span>
              <span class="text-[0.65rem] font-semibold leading-[1.05rem]">{{ item.label }}</span>
            </button>
          </div>
        </aside>

        <div class="home-lobby-mobile__panel min-w-0">
          <div class="mb-4 flex items-center justify-around border-b border-white/10 pb-2">
            <button 
              class="relative text-[0.85rem] font-bold transition-colors"
              :class="activeMobileLobbyTab === 'category' ? 'text-primary' : 'text-slate-500 hover:text-primary/70'"
              @click="switchLobbyTab('category')"
            >
              {{ activeMobileLobbyItem?.label }}
              <div v-if="activeMobileLobbyTab === 'category'" class="absolute -bottom-[9px] left-1/2 h-[3px] w-8 -translate-x-1/2 rounded-t-md bg-primary"></div>
            </button>
            <button 
              class="relative text-[0.85rem] font-bold transition-colors"
              :class="activeMobileLobbyTab === 'history' ? 'text-primary' : 'text-slate-500 hover:text-primary/70'"
              @click="switchLobbyTab('history')"
            >
              Lịch Sử
              <div v-if="activeMobileLobbyTab === 'history'" class="absolute -bottom-[9px] left-1/2 h-[3px] w-8 -translate-x-1/2 rounded-t-md bg-primary"></div>
            </button>
            <button 
              class="relative text-[0.85rem] font-bold transition-colors"
              :class="activeMobileLobbyTab === 'favorite' ? 'text-primary' : 'text-slate-500 hover:text-primary/70'"
              @click="switchLobbyTab('favorite')"
            >
              Yêu Thích
              <div v-if="activeMobileLobbyTab === 'favorite'" class="absolute -bottom-[9px] left-1/2 h-[3px] w-8 -translate-x-1/2 rounded-t-md bg-primary"></div>
            </button>
          </div>

          <div v-if="mobileLobbyGames.length" class="grid grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-x-3 gap-y-4">
            <RouterLink
              v-for="(game, index) in mobileLobbyGames"
              :key="`${game.name}-mobile-lobby`"
              :to="resolveGameTarget(game)"
              class="group min-w-0"
              :class="isWideLobbyGame(game) ? 'col-span-full' : ''"
              @pointerenter="maybePrefetchGameRoute(game)"
              @focus="maybePrefetchGameRoute(game)"
              @touchstart.passive="maybePrefetchGameRoute(game)"
              @click="trackRecentGame(game)"
            >
              <div class="home-lobby-mobile__card relative overflow-hidden rounded-[16px]">
                <img
                  :src="game.image"
                  :alt="game.name"
                  class="w-full object-cover transition-transform duration-500 group-hover:scale-[1.04]"
                  :class="isWideLobbyGame(game) ? 'aspect-[3/1]' : 'aspect-square'"
                  loading="lazy"
                  decoding="async"
                />
                <div class="absolute inset-0 bg-gradient-to-t from-black/18 via-transparent to-transparent" />
                <div
                  v-if="mobileLobbyBadgeText(game, index) === 'Hot'"
                  class="absolute left-0 top-0 rounded-br-[8px] bg-gradient-to-br from-[#ffb732] to-[#ff8f2f] px-1.5 py-0.5 shadow-[0_2px_6px_rgba(255,143,47,0.4)]"
                >
                  <span class="material-symbols-outlined block text-[0.7rem] text-white">thumb_up</span>
                </div>
                <span
                  v-else-if="mobileLobbyBadgeText(game, index)"
                  class="absolute left-1.5 top-1.5 rounded-full bg-[#ff8f2f] px-1.5 py-0.5 text-[0.5rem] font-black uppercase tracking-[0.08em] text-white shadow-[0_6px_12px_rgba(255,143,47,0.35)]"
                >
                  {{ mobileLobbyBadgeText(game, index) }}
                </span>
                <button
                  type="button"
                  class="absolute right-1.5 top-1.5 flex h-[22px] w-[22px] items-center justify-center rounded-full bg-black/30 text-white/90 backdrop-blur-sm"
                  @click.prevent.stop="toggleFavoriteGame(game)"
                >
                  <span class="material-symbols-outlined text-[0.8rem]" :class="isFavoriteGame(game) ? 'text-yellow-400' : ''">
                    {{ isFavoriteGame(game) ? 'kid_star' : 'star' }}
                  </span>
                </button>
              </div>
              <p class="mt-1.5 text-center line-clamp-2 text-[0.72rem] leading-4 text-slate-700 font-bold">{{ game.name }}</p>
            </RouterLink>
          </div>

          <div
            v-else
            class="flex min-h-[240px] items-center justify-center rounded-[20px] border border-dashed border-red-200 bg-white/60 px-5 text-center shadow-[inset_0_1px_0_rgba(255,255,255,0.8)]"
          >
            <div>
              <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-red-50">
                <span class="material-symbols-outlined text-[1.2rem] text-primary">deployed_code</span>
              </div>
              <p class="mt-3 text-[0.76rem] font-black text-slate-700">Danh mục đang trống</p>
              <p class="mt-1 text-[0.64rem] font-semibold leading-5 text-slate-500">
                {{ mobileLobbyEmptyMessage }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>

      <section class="mx-3 mt-3 hidden overflow-hidden rounded-[24px] bg-gradient-to-br from-[#ff6e67] via-primary to-[#e73d47] shadow-[0_14px_28px_rgba(218,37,29,0.2)] !hidden">
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

    <section v-if="activeCategory === 'Phổ biến'" class="mt-4 hidden px-3 !hidden">
      <div class="flex min-w-0 flex-col gap-2">
        <RouterLink
          v-for="(game, index) in categoryBannerGames"
          :key="`${game.name}-banner`"
          :to="resolveGameTarget(game)"
          class="group relative block overflow-hidden rounded-[16px] shadow-[0_4px_16px_rgba(0,0,0,0.12)] transition-all duration-300 active:scale-[0.98]"
          @pointerenter="maybePrefetchGameRoute(game)"
          @focus="maybePrefetchGameRoute(game)"
          @touchstart.passive="maybePrefetchGameRoute(game)"
          @click="trackRecentGame(game)"
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

    <section class="mt-4 hidden px-3 !hidden">
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
          @click="trackRecentGame(game)"
        >
          <img
            :src="game.image"
            :alt="game.name"
            class="w-full object-cover transition-transform duration-500 group-hover:scale-[1.04]"
            :class="game.route ? 'aspect-[3/1]' : 'aspect-square'"
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
      <div class="rounded-[20px] bg-[#303133] px-4 py-4 text-white shadow-[0_12px_24px_rgba(0,0,0,0.25)] border border-primary/35">
        <p class="text-[0.72rem] font-black uppercase tracking-[0.08em] text-yellow-500/90">Thông tin truy cập</p>
        <p class="mt-2 text-[0.88rem] font-semibold leading-7 text-white/92">
          <b>FH88U </b>
Được biết đến là cổng Game giải trí trực tuyến chất lượng dành cho người chơi tại Châu Á. Tại đây hội tụ nhiều sản phẩm nổi bật và nhiều trò chơi hấp dẫn. Đơn vị luôn ưu tiên bảo vệ dữ liệu người dùng, hỗ trợ khách hàng nhanh chóng và không ngừng nâng cao chất lượng dịch vụ để mang lại trải nghiệm tốt nhất. Hãy liên hệ chúng tôi để được hỗ trợ những thắc mắc.
        </p>
        <button
          type="button"
          class="mt-4 inline-flex items-center gap-1.5 rounded-full bg-gradient-to-r from-yellow-400 to-yellow-500 px-4 py-2.5 text-[0.8rem] font-black text-[#241900] shadow-[0_10px_20px_rgba(255,204,0,0.2)] transition-transform active:scale-95"
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

<style scoped>
.home-lobby-mobile {
  background: transparent;
}

.home-lobby-mobile::before {
  display: none;
}

.home-lobby-mobile__shell {
  display: grid;
  grid-template-columns: 5.85rem minmax(0, 1fr);
  column-gap: 0.7rem;
  align-items: start;
}

.home-lobby-mobile__sidebar {
  z-index: 1;
}

.home-lobby-mobile__panel {
  min-width: 0;
}

.home-lobby-mobile__nav-item {
  border: 1px solid rgba(255, 255, 255, 0.08);
  min-height: 5.1rem;
}

.home-lobby-mobile__card {
  background: transparent;
  box-shadow: none;
  border: none;
}
.home-lobby-mobile__card img {
  border-radius: 16px;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.28);
}
</style>
