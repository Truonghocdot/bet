<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter, type RouteLocationRaw } from 'vue-router'

import { request } from '@/shared/api/http'
import type {
  ProviderGameCatalogCategoryItem,
  ProviderGameCatalogItem,
  ProviderGameCatalogResponse,
} from '@/shared/api/types'

const casinoSidebarIconModules = import.meta.glob<string>('@/assets/game_thumbs/lobby-casino/icon/*.avif', {
  eager: true,
  query: '?url',
  import: 'default',
})

const casinoShowIconModules = import.meta.glob<string>('@/assets/game_thumbs/lobby-casino/show-icon/*.avif', {
  eager: true,
  query: '?url',
  import: 'default',
})

const casinoSidebarIconMap = Object.fromEntries(
  Object.entries(casinoSidebarIconModules).map(([path, url]) => {
    const fileName = path.split('/').pop() ?? ''
    const baseName = fileName.replace(/\.[^.]+$/, '')
    return [baseName, url]
  }),
) as Record<string, string>

const casinoShowIconMap = Object.fromEntries(
  Object.entries(casinoShowIconModules).map(([path, url]) => {
    const fileName = path.split('/').pop() ?? ''
    const baseName = fileName.replace(/\.[^.]+$/, '')
    return [baseName, url]
  }),
) as Record<string, string>

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const error = ref('')
interface SidebarProviderItem extends ProviderGameCatalogCategoryItem {
  sidebarKey: string
  sidebarLabel: string
  sidebarIconKey?: string
}

const allProviders = ref<SidebarProviderItem[]>([])
const activeProductType = ref(0)

type ProviderLobbyPreset = {
  label: string
  productType: number
  sidebarIconKey?: string
}

type CasinoGlobalLobbyConfig = {
  displayName?: string
  showIconKey?: string
  preferredCodes?: string[]
}

const slotLobbyPresets: ProviderLobbyPreset[] = [
  { label: '168GAME', productType: 275 },
  { label: 'CQ9', productType: 16 },
  { label: 'FC', productType: 141 },
  { label: 'JDB', productType: 55 },
  { label: 'JILI', productType: 140 },
  { label: 'PG', productType: 98 },
  { label: 'TP', productType: 243 },
  { label: 'WG', productType: 212 },
]

const fishLobbyPresets: ProviderLobbyPreset[] = [
  { label: 'BBIN', productType: 79 },
  { label: 'CQ9', productType: 16 },
  { label: 'FC', productType: 141 },
  { label: 'JDB', productType: 55 },
  { label: 'JILI', productType: 140 },
  { label: 'MG', productType: 43 },
  { label: 'RSG', productType: 138 },
  { label: 'TP', productType: 243 },
  { label: 'WG', productType: 212 },
]

const cardLobbyPresets: ProviderLobbyPreset[] = [
  { label: 'JILI', productType: 140 },
  { label: 'MG', productType: 43 },
  { label: 'WG', productType: 212 },
]

const casinoLobbyPresets: ProviderLobbyPreset[] = [
  { label: 'SEXY', productType: 112, sidebarIconKey: 'SEXY' },
  { label: 'AG', productType: 4, sidebarIconKey: 'AG' },
  { label: 'BBIN', productType: 79, sidebarIconKey: 'BBIN' },
  { label: 'DG', productType: 27, sidebarIconKey: 'DG' },
  { label: 'EZUGI', productType: 177 },
  { label: 'HRG', productType: 93, sidebarIconKey: 'HRG' },
  { label: 'MG', productType: 43, sidebarIconKey: 'MG' },
  { label: 'MT', productType: 272 },
  { label: 'PP', productType: 39, sidebarIconKey: 'PP' },
  { label: 'PT', productType: 3, sidebarIconKey: 'PT' },
  { label: 'SA', productType: 93, sidebarIconKey: 'SA' },
  { label: 'TP', productType: 93, sidebarIconKey: 'TP' },
  { label: 'W', productType: 258, sidebarIconKey: 'NW' },
  { label: 'WM', productType: 118, sidebarIconKey: 'WM' },
]

const expandedCasinoLobbyKeys = new Set(['SEXY', 'WM'])

const casinoGlobalLobbyConfigMap: Record<string, CasinoGlobalLobbyConfig> = {
  AG: { displayName: 'CHOICE', showIconKey: 'CHOICE', preferredCodes: ['A00234'] },
  BBIN: { showIconKey: 'BBIN', preferredCodes: ['BBB001'] },
  DG: { showIconKey: 'DG', preferredCodes: ['DG0013'] },
  MG: { preferredCodes: ['MG0554'] },
  MT: { showIconKey: 'MT' },
  PP: { showIconKey: 'PP' },
  SA: { showIconKey: 'SA', preferredCodes: ['SA0001'] },
  TP: { showIconKey: 'TP' },
}

const casinoProviderAliasMap: Record<string, string[]> = {
  AG: ['ag', 'choice'],
  BBIN: ['bbin', 'bb'],
  DG: ['dg'],
  HRG: ['hrg'],
  MG: ['mg'],
  MT: ['mt'],
  PP: ['pp'],
  PT: ['pt', 'playtech'],
  SA: ['sa'],
  SEXY: ['sexy', 'sex'],
  TP: ['tp'],
  W: ['w', 'nw'],
  WM: ['wm'],
}

const categoryKey = computed(() => String(route.params.category ?? '').trim().toLowerCase())
const isCasinoCategory = computed(() => categoryKey.value === 'casino')
const activeSidebarKey = computed(() => String(route.query.provider_key ?? route.query.name ?? '').trim())
const productTypeParam = computed(() => {
  const parsed = Number.parseInt(String(route.params.productType ?? '').trim(), 10)
  return Number.isFinite(parsed) ? parsed : 0
})

const pageTitle = computed(() => {
  const activeProviderLabel = String(activeProvider.value?.sidebarLabel || '').trim()
  if (activeProviderLabel) return activeProviderLabel

  const queryName = String(route.query.name ?? '').trim()
  return queryName || categoryLabel.value
})

const categoryLabel = computed(() => {
  const map: Record<string, string> = {
    casino: 'Casino',
    slots: 'Nổ Hũ',
    fish: 'Bắn Cá',
    cards: 'Game Bài',
    cockfight: 'Đá Gà',
    sports: 'Thể Thao',
    lottery: 'Xổ Số',
  }
  return map[categoryKey.value] ?? 'Danh sách Lobby'
})

const backTarget = computed(() => {
  const from = String(route.query.from ?? '').trim()
  return from.startsWith('/') ? from : '/home'
})

const activeProvider = computed(() => {
  if (activeSidebarKey.value) {
    const exact = allProviders.value.find((provider) => provider.sidebarKey === activeSidebarKey.value)
    if (exact) return exact
  }

  return allProviders.value.find((p) => Number(p.product_type ?? 0) === activeProductType.value) ?? null
})

const lobbyItems = computed(() => {
  const provider = activeProvider.value
  const children = filterLobbyChildren(provider?.children ?? [])

  if (!provider) return []
  if (!isCasinoCategory.value) return children

  if (expandedCasinoLobbyKeys.has(casinoProviderKey(provider))) {
    return children
  }

  const globalLobby = buildCasinoGlobalLobbyItem(provider, children)
  return globalLobby ? [globalLobby] : children
})

function parseProductType(value: string | number | null | undefined): number {
  const parsed = Number.parseInt(String(value ?? '').trim(), 10)
  return Number.isFinite(parsed) ? parsed : 0
}

function providerProductType(item: ProviderGameCatalogCategoryItem): number {
  const directProductType = Number(item.product_type ?? 0)
  if (directProductType > 0) return directProductType
  return parseProductType(item.product_type_value)
}

type CatalogKeywordSource = Pick<
  ProviderGameCatalogItem,
  'game_name' | 'tcg_game_code' | 'product_code' | 'game_sub_type'
>

function categoryLobbyPresets(category: string): ProviderLobbyPreset[] {
  switch (category) {
    case 'casino':
      return casinoLobbyPresets
    case 'slots':
      return slotLobbyPresets
    case 'fish':
      return fishLobbyPresets
    case 'cards':
      return cardLobbyPresets
    default:
      return []
  }
}

function presetLabelForProvider(category: string, productType: number, fallback: string): string {
  const preset = categoryLobbyPresets(category).find((item) => item.productType === productType)
  return preset?.label ?? fallback
}

function presetForProvider(category: string, productType: number): ProviderLobbyPreset | undefined {
  return categoryLobbyPresets(category).find((item) => item.productType === productType)
}

function providerSidebarLabel(category: string, productType: number, fallback: string): string {
  if (category === 'casino' && productType === 4) {
    return 'CHOICE'
  }

  const preset = presetForProvider(category, productType)
  return preset?.label ?? fallback
}

function localCasinoSidebarIcon(provider: SidebarProviderItem): string {
  const explicitKey = String(provider.sidebarIconKey || '').trim()
  if (explicitKey && casinoSidebarIconMap[explicitKey]) {
    return casinoSidebarIconMap[explicitKey]
  }

  const sidebarLabel = String(provider.sidebarLabel || '').trim()
  if (sidebarLabel && casinoSidebarIconMap[sidebarLabel]) {
    return casinoSidebarIconMap[sidebarLabel]
  }

  return ''
}

function providerSidebarIcon(provider: SidebarProviderItem): string {
  if (categoryKey.value === 'casino') {
    const localIcon = localCasinoSidebarIcon(provider)
    if (localIcon) return localIcon
  }

  return String(provider.show_icon || '').trim()
}

function buildSidebarProviders(category: string, items: ProviderGameCatalogCategoryItem[]): SidebarProviderItem[] {
  const orderedPresets = categoryLobbyPresets(category)
  if (orderedPresets.length === 0) {
    return items.map((item) => {
      const fallbackLabel = String(item.game_name || item.tcg_game_code || '').trim()
      return {
        ...item,
        sidebarKey: String(item.product_type || item.product_type_value || fallbackLabel).trim(),
        sidebarLabel: fallbackLabel,
        sidebarIconKey: '',
      }
    })
  }

  const mapped = items.map((item) => {
    const productType = providerProductType(item)
    const fallbackLabel = String(item.game_name || item.tcg_game_code || '').trim()
    const preset = presetForProvider(category, productType)
    const sidebarKey = preset?.label ?? fallbackLabel
    const sidebarLabel = providerSidebarLabel(category, productType, fallbackLabel)
    return {
      ...item,
      sidebarKey,
      sidebarLabel,
      sidebarIconKey: String(preset?.sidebarIconKey || '').trim(),
    }
  })

  return [...mapped].sort((left, right) => {
    const leftPT = providerProductType(left)
    const rightPT = providerProductType(right)
    const leftIndex = orderedPresets.findIndex((preset) => preset.productType === leftPT)
    const rightIndex = orderedPresets.findIndex((preset) => preset.productType === rightPT)
    if (leftIndex !== rightIndex) return leftIndex - rightIndex
    return left.sidebarLabel.localeCompare(right.sidebarLabel, 'vi')
  })
}

function supportsDisplayPlatform(platform: string): boolean {
  const normalized = String(platform || '').trim().toLowerCase()
  if (!normalized) return false
  return normalized.includes('html5')
    || normalized.includes('web')
    || normalized.includes('mobile')
    || normalized.includes('desktop')
}

function catalogKeywordValues(item: CatalogKeywordSource): string[] {
  return [
    String(item.game_name || '').trim().toLowerCase(),
    String(item.tcg_game_code || '').trim().toLowerCase(),
    String(item.product_code || '').trim().toLowerCase(),
    String(item.game_sub_type || '').trim().toLowerCase(),
  ].filter(Boolean)
}

function catalogContainsKeyword(item: CatalogKeywordSource, keywords: string[]): boolean {
  const values = catalogKeywordValues(item)

  return keywords.some((keyword) => {
    const normalizedKeyword = String(keyword || '').trim().toLowerCase()
    return normalizedKeyword !== '' && values.some((value) => value.includes(normalizedKeyword))
  })
}

function containsProviderKeyword(item: ProviderGameCatalogCategoryItem, keywords: string[]): boolean {
  return catalogContainsKeyword(item, keywords)
}

function casinoProviderKey(provider: SidebarProviderItem): string {
  return String(provider.sidebarKey || provider.sidebarLabel || provider.product_code || '').trim().toUpperCase()
}

function casinoGlobalLobbyConfig(provider: SidebarProviderItem): CasinoGlobalLobbyConfig {
  return casinoGlobalLobbyConfigMap[casinoProviderKey(provider)] ?? {}
}

function localCasinoShowIcon(provider: SidebarProviderItem): string {
  const config = casinoGlobalLobbyConfig(provider)
  const explicitKey = String(config.showIconKey || '').trim()
  if (explicitKey && casinoShowIconMap[explicitKey]) {
    return casinoShowIconMap[explicitKey]
  }

  const providerKey = casinoProviderKey(provider)
  if (providerKey && casinoShowIconMap[providerKey]) {
    return casinoShowIconMap[providerKey]
  }

  return ''
}

function casinoGlobalLobbyScore(provider: SidebarProviderItem, item: ProviderGameCatalogItem): number {
  const providerKey = casinoProviderKey(provider)
  const config = casinoGlobalLobbyConfig(provider)
  const aliases = casinoProviderAliasMap[providerKey] ?? [providerKey.toLowerCase()]
  const normalizedCode = String(item.tcg_game_code || '').trim().toUpperCase()
  const hasProviderAlias = catalogContainsKeyword(item, aliases)
  const hasLobbyKeyword = catalogContainsKeyword(item, [
    'game_list',
    'game list',
    'lobby',
    'sảnh',
    'sanh',
    'sảnh trò chơi',
    'sanh tro choi',
    'sảnh chờ',
    'sanh cho',
  ])
  const hasProviderLobbyKeyword = catalogContainsKeyword(item, [
    'trực tuyến',
    'truc tuyen',
    'mobile',
    'live',
  ])

  let score = 0

  if ((config.preferredCodes ?? []).includes(normalizedCode)) score += 100
  if (hasLobbyKeyword) score += 40
  if (hasProviderAlias) score += 6
  if (hasProviderAlias && hasProviderLobbyKeyword) score += 24
  if (String(item.show_icon || '').trim()) score += 2
  if (supportsDisplayPlatform(item.platform)) score += 1

  return score
}

function buildCasinoGlobalLobbyItem(
  provider: SidebarProviderItem,
  children: ProviderGameCatalogItem[],
): ProviderGameCatalogItem | null {
  const rankedChildren = children
    .map((item, index) => ({
      item,
      index,
      score: casinoGlobalLobbyScore(provider, item),
    }))
    .sort((left, right) => {
      if (left.score !== right.score) return right.score - left.score
      return left.index - right.index
    })

  const bestChild = rankedChildren[0]
  const sourceItem = bestChild && bestChild.score > 0 ? bestChild.item : toProviderGameItem(provider)
  const showIcon = localCasinoShowIcon(provider)
  const config = casinoGlobalLobbyConfig(provider)
  const displayName = String(config.displayName || provider.sidebarLabel || sourceItem.game_name || '').trim()

  if (!sourceItem.tcg_game_code || !sourceItem.game_type) {
    return null
  }

  return {
    ...sourceItem,
    game_name: displayName || sourceItem.game_name,
    show_icon: showIcon || sourceItem.show_icon,
  }
}

function providerHeroRank(item: ProviderGameCatalogCategoryItem): number {
  let score = 0

  if (containsProviderKeyword(item, ['game_list', 'game list', 'lobby'])) score += 4
  if (String(item.show_icon || '').trim()) score += 2
  if (supportsDisplayPlatform(item.platform)) score += 1

  return score
}

function providerGroupName(item: ProviderGameCatalogCategoryItem, productType: number): string {
  const slotProviderNameMap: Record<number, string> = {
    275: '168GAME',
    16: 'CQ9',
    141: 'FC',
    55: 'JDB',
    140: 'JILI',
    98: 'PG',
    243: 'TP',
    212: 'WG',
  }

  return slotProviderNameMap[productType]
    || String(item.product_code || item.game_name || item.tcg_game_code || '').trim()
    || `Provider ${productType}`
}

function toProviderGameItem(item: ProviderGameCatalogCategoryItem): ProviderGameCatalogItem {
  return {
    display_status: Number(item.display_status ?? 0),
    game_type: String(item.game_type || '').trim(),
    game_name: String(item.game_name || '').trim(),
    tcg_game_code: String(item.tcg_game_code || '').trim(),
    product_code: String(item.product_code || '').trim(),
    product_type_value: String(item.product_type_value || item.product_type || '').trim(),
    platform: String(item.platform || '').trim(),
    game_sub_type: String(item.game_sub_type || '').trim(),
    show_icon: String(item.show_icon || '').trim() || undefined,
    trial_support: Boolean(item.trial_support),
  }
}

function buildVirtualProviders(items: ProviderGameCatalogCategoryItem[]): ProviderGameCatalogCategoryItem[] {
  const grouped = new Map<number, ProviderGameCatalogCategoryItem[]>()

  for (const item of items) {
    if (Number(item.display_status ?? 0) !== 0) continue
    if (!supportsDisplayPlatform(item.platform)) continue

    const productType = providerProductType(item)
    if (productType <= 0) continue

    const bucket = grouped.get(productType) ?? []
    bucket.push(item)
    grouped.set(productType, bucket)
  }

  return [...grouped.entries()]
    .sort(([left], [right]) => left - right)
    .map(([productType, providerItems]) => {
      const hero = [...providerItems].sort((left, right) => {
        const rankDiff = providerHeroRank(right) - providerHeroRank(left)
        if (rankDiff !== 0) return rankDiff
        return String(left.game_name || '').trim().localeCompare(String(right.game_name || '').trim(), 'vi')
      })[0]

      const children = providerItems.map(toProviderGameItem)
      const productTypeValue = hero
        ? String(hero.product_type_value || hero.product_type || productType).trim()
        : String(productType)

      return {
        kind: 'group',
        display_status: 0,
        game_type: hero ? String(hero.game_type || '').trim() : '',
        game_name: hero ? providerGroupName(hero, productType) : `Provider ${productType}`,
        tcg_game_code: hero ? String(hero.tcg_game_code || '').trim() : '',
        product_code: hero ? String(hero.product_code || '').trim() : '',
        product_type: productType,
        product_type_value: productTypeValue,
        platform: hero ? String(hero.platform || '').trim() : '',
        game_sub_type: hero ? String(hero.game_sub_type || '').trim() : '',
        show_icon: hero ? String(hero.show_icon || '').trim() || undefined : undefined,
        trial_support: hero ? Boolean(hero.trial_support) : false,
        child_count: children.length,
        children,
      }
    })
}

function lobbyTarget(item: ProviderGameCatalogItem): RouteLocationRaw {
  return {
    name: 'partner-game-loading',
    query: {
      name: String(item.game_name || '').trim() || String(item.tcg_game_code || 'TCG Lobby').trim(),
      from: route.fullPath,
      game_code: String(item.tcg_game_code || '').trim(),
      product_type: String(item.product_type_value || activeProvider.value?.product_type || '').trim(),
      game_type: String(item.game_type || activeProvider.value?.game_type || '').trim(),
    },
  }
}

function selectProvider(provider: SidebarProviderItem) {
  const productType = Number(provider.product_type ?? 0)
  activeProductType.value = productType
  void router.replace({
    params: { ...route.params, productType },
    query: {
      ...route.query,
      name: provider.sidebarLabel,
      provider_key: provider.sidebarKey,
    },
  })
}

function providerName(item: SidebarProviderItem | ProviderGameCatalogCategoryItem): string {
  if ('sidebarLabel' in item && String(item.sidebarLabel || '').trim()) {
    return String(item.sidebarLabel || '').trim()
  }
  return String(item.game_name || item.tcg_game_code || '').trim()
}

function filterLobbyChildren(items: ProviderGameCatalogItem[]): ProviderGameCatalogItem[] {
  return items.filter((item) => {
    if (Number(item.display_status ?? 0) !== 0) return false
    return supportsDisplayPlatform(item.platform)
  })
}

async function fetchCasinoLobbyProvidersSequentially(): Promise<SidebarProviderItem[]> {
  const providers: SidebarProviderItem[] = []

  for (const preset of casinoLobbyPresets) {
    try {
      const response = await request<ProviderGameCatalogResponse>(
        'GET',
        `/v1/provider-games/tcg?category=casino&product_type=${preset.productType}&include_children=1`,
      )
      const category = response.categories.find(
        (item) => String(item.key || '').trim().toLowerCase() === 'casino',
      )
      const matched = (category?.items ?? []).find(
        (item) => providerProductType(item) === preset.productType,
      )
      if (!matched) continue

      const children = filterLobbyChildren(matched.children ?? [])
      if (children.length === 0) continue

      providers.push({
        ...matched,
        sidebarKey: preset.label,
        sidebarLabel: providerSidebarLabel('casino', preset.productType, preset.label),
        sidebarIconKey: String(preset.sidebarIconKey || '').trim(),
        child_count: children.length,
        children,
      })
    } catch {
      // Skip unavailable provider and continue preserving root-category order.
    }
  }

  return providers
}

async function fetchLobbyList() {
  if (!categoryKey.value) {
    error.value = 'Thông tin danh mục không hợp lệ.'
    return
  }

  loading.value = true
  error.value = ''

  try {
    let items: SidebarProviderItem[] = []

    if (isCasinoCategory.value) {
      items = await fetchCasinoLobbyProvidersSequentially()
    } else {
      const response = await request<ProviderGameCatalogResponse>(
        'GET',
        `/v1/provider-games/tcg?category=${encodeURIComponent(categoryKey.value)}&include_children=1`,
      )
      const category = response.categories.find(
        (item) => String(item.key || '').trim().toLowerCase() === categoryKey.value,
      )
      const categoryItems = (category?.items ?? []).filter((item) => Number(item.display_status ?? 0) === 0)
      const groupedItems = categoryItems.filter(
        (item) =>
          supportsDisplayPlatform(item.platform) &&
          (Number(item.child_count ?? 0) > 0 || String(item.kind || '').trim() === 'group'),
      )
      items = buildSidebarProviders(
        categoryKey.value,
        groupedItems.length > 0 ? groupedItems : buildVirtualProviders(categoryItems),
      )
    }

    allProviders.value = items

    // Chọn provider từ URL param, nếu không có thì chọn provider đầu tiên
    const targetPT = productTypeParam.value
    if (targetPT > 0 && items.some((p) => Number(p.product_type ?? 0) === targetPT)) {
      activeProductType.value = targetPT
    } else if (items.length > 0) {
      const firstProvider = items[0]
      if (firstProvider) {
        activeProductType.value = Number(firstProvider.product_type ?? 0)
      }
    }
  } catch {
    error.value = 'Không thể tải danh sách lobby lúc này.'
  } finally {
    loading.value = false
  }
}

function goBack() {
  void router.push(backTarget.value)
}

watch(
  () => categoryKey.value,
  () => { void fetchLobbyList() },
)

watch(
  () => productTypeParam.value,
  (productType) => {
    if (productType <= 0) return
    if (!allProviders.value.some((provider) => Number(provider.product_type ?? 0) === productType)) return
    activeProductType.value = productType
  },
)

onMounted(() => {
  void fetchLobbyList()
})
</script>

<template>
  <div class="px-3 py-4">
    <!-- Header -->
    <div class="mb-3 flex items-center gap-3">
      <button
        type="button"
        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-red-100 bg-white text-primary shadow-[0_6px_16px_rgba(218,37,29,0.1)]"
        @click="goBack"
      >
        <span class="material-symbols-outlined text-[1.1rem]">arrow_back</span>
      </button>
      <div class="min-w-0">
        <h1 class="truncate text-[1.1rem] font-black text-slate-900">{{ pageTitle }}</h1>
        <p class="text-[0.72rem] font-semibold text-slate-400">
          {{ loading ? 'Đang tải...' : `${lobbyItems.length} game khả dụng` }}
        </p>
      </div>
    </div>

    <!-- Error state -->
    <div
      v-if="error"
      class="flex min-h-[240px] items-center justify-center rounded-[20px] border border-dashed border-red-200 bg-white/70 px-5 text-center"
    >
      <div>
        <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-red-50">
          <span class="material-symbols-outlined text-[1.2rem] text-primary">error</span>
        </div>
        <p class="mt-3 text-[0.84rem] font-black text-slate-700">Không thể tải lobby</p>
        <p class="mt-1 text-[0.72rem] font-semibold leading-5 text-slate-500">{{ error }}</p>
        <button
          type="button"
          class="mt-4 rounded-full bg-gradient-to-r from-[#ff7b71] to-primary px-4 py-2 text-[0.78rem] font-black text-white"
          @click="fetchLobbyList"
        >
          Tải lại
        </button>
      </div>
    </div>

    <!-- Main layout: sidebar + game grid -->
    <div v-else class="flex gap-3">
      <!-- Sidebar providers -->
      <aside class="w-[5.8rem] shrink-0">
        <div
          v-if="loading"
          class="space-y-2"
        >
          <div
            v-for="i in 6"
            :key="i"
            class="h-16 w-full animate-pulse rounded-[14px] bg-slate-100"
          />
        </div>

        <div v-else class="space-y-2">
          <button
            v-for="provider in allProviders"
            :key="provider.sidebarKey"
            type="button"
            class="flex w-full items-center gap-1.5 rounded-[14px] px-1.5 py-2 text-left transition-all duration-200"
            :class="
              activeProvider?.sidebarKey === provider.sidebarKey
                ? 'bg-[linear-gradient(180deg,#da251d_0%,#a81b14_100%)] text-white shadow-[0_8px_16px_rgba(218,37,29,0.28)]'
                : 'bg-white text-slate-600 shadow-sm'
            "
            @click="selectProvider(provider)"
          >
            <img
              v-if="providerSidebarIcon(provider)"
              :src="providerSidebarIcon(provider)"
              :alt="providerName(provider)"
              class="h-7 w-7 rounded-[8px] object-contain"
              loading="lazy"
            />
            <div
              v-else
              class="flex h-7 w-7 items-center justify-center rounded-[8px]"
              :class="activeProvider?.sidebarKey === provider.sidebarKey ? 'bg-white/20' : 'bg-red-50'"
            >
              <span class="material-symbols-outlined text-[1rem]">casino</span>
            </div>
            <span class="min-w-0 flex-1 whitespace-nowrap text-[0.58rem] font-bold leading-none">
              {{ providerName(provider) }}
            </span>
          </button>
        </div>
      </aside>

      <!-- Game grid -->
      <div class="min-w-0 flex-1">
        <!-- Loading skeleton for games -->
        <div v-if="loading" class="grid grid-cols-3 gap-2">
          <div
            v-for="i in 9"
            :key="i"
            class="aspect-square w-full animate-pulse rounded-[14px] bg-slate-100"
          />
        </div>

        <!-- Games grid -->
        <div v-else-if="lobbyItems.length" class="grid grid-cols-3 gap-2">
          <RouterLink
            v-for="item in lobbyItems"
            :key="`${item.product_type_value}:${item.tcg_game_code}`"
            :to="lobbyTarget(item)"
            class="group overflow-hidden rounded-[14px] bg-white shadow-[0_6px_16px_rgba(15,23,42,0.08)] transition-transform active:scale-[0.97]"
          >
            <div class="relative">
              <img
                v-if="item.show_icon"
                :src="item.show_icon"
                :alt="item.game_name"
                class="aspect-square w-full object-cover transition-transform duration-500 group-hover:scale-[1.04]"
                loading="lazy"
                decoding="async"
              />
              <div
                v-else
                class="flex aspect-square w-full items-center justify-center bg-[radial-gradient(circle_at_top,#ffded8_0%,#fff4f1_100%)]"
              >
                <span class="material-symbols-outlined text-[2rem] text-primary/60">casino</span>
              </div>
            </div>
            <p class="line-clamp-2 px-2 py-2 text-center text-[0.68rem] font-bold leading-4 text-slate-700">
              {{ item.game_name }}
            </p>
          </RouterLink>
        </div>

        <!-- Empty -->
        <div
          v-else
          class="flex min-h-[240px] items-center justify-center rounded-[20px] border border-dashed border-red-200 bg-white/60 px-5 text-center"
        >
          <div>
            <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-red-50">
              <span class="material-symbols-outlined text-[1.2rem] text-primary">deployed_code</span>
            </div>
            <p class="mt-3 text-[0.76rem] font-black text-slate-700">Chưa có game khả dụng</p>
            <p class="mt-1 text-[0.64rem] font-semibold leading-5 text-slate-500">
              Nhà cung cấp này đang cập nhật danh mục.
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
