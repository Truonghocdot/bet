<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter, type RouteLocationRaw } from 'vue-router'

import { request } from '@/shared/api/http'
import type {
  ProviderGameCatalogCategoryItem,
  ProviderGameCatalogItem,
  ProviderGameCatalogResponse,
} from '@/shared/api/types'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const error = ref('')
const allProviders = ref<ProviderGameCatalogCategoryItem[]>([])
const activeProductType = ref(0)

const categoryKey = computed(() => String(route.params.category ?? '').trim().toLowerCase())
const productTypeParam = computed(() => {
  const parsed = Number.parseInt(String(route.params.productType ?? '').trim(), 10)
  return Number.isFinite(parsed) ? parsed : 0
})

const pageTitle = computed(() => {
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

const activeProvider = computed(() =>
  allProviders.value.find((p) => Number(p.product_type ?? 0) === activeProductType.value) ?? null,
)

const lobbyItems = computed(() => {
  const children = activeProvider.value?.children ?? []
  return children.filter((item) => {
    if (Number(item.display_status ?? 0) !== 0) return false
    return supportsDisplayPlatform(item.platform)
  })
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

function supportsDisplayPlatform(platform: string): boolean {
  const normalized = String(platform || '').trim().toLowerCase()
  if (!normalized) return false
  return normalized.includes('html5')
    || normalized.includes('web')
    || normalized.includes('mobile')
    || normalized.includes('desktop')
}

function containsProviderKeyword(item: ProviderGameCatalogCategoryItem, keywords: string[]): boolean {
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

function selectProvider(productType: number) {
  activeProductType.value = productType
  void router.replace({
    params: { ...route.params, productType },
    query: route.query,
  })
}

function providerName(item: ProviderGameCatalogCategoryItem): string {
  return String(item.game_name || item.tcg_game_code || '').trim()
}

async function fetchLobbyList() {
  if (!categoryKey.value) {
    error.value = 'Thông tin danh mục không hợp lệ.'
    return
  }

  loading.value = true
  error.value = ''

  try {
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
    const items = groupedItems.length > 0 ? groupedItems : buildVirtualProviders(categoryItems)

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
      <aside class="w-[4.8rem] shrink-0">
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
            :key="String(provider.product_type)"
            type="button"
            class="flex w-full flex-col items-center gap-1 rounded-[14px] px-1 py-2 text-center transition-all duration-200"
            :class="
              activeProductType === Number(provider.product_type)
                ? 'bg-[linear-gradient(180deg,#da251d_0%,#a81b14_100%)] text-white shadow-[0_8px_16px_rgba(218,37,29,0.28)]'
                : 'bg-white text-slate-600 shadow-sm'
            "
            @click="selectProvider(Number(provider.product_type))"
          >
            <img
              v-if="provider.show_icon"
              :src="provider.show_icon"
              :alt="providerName(provider)"
              class="h-8 w-8 rounded-[8px] object-contain"
              loading="lazy"
            />
            <div
              v-else
              class="flex h-8 w-8 items-center justify-center rounded-[8px]"
              :class="activeProductType === Number(provider.product_type) ? 'bg-white/20' : 'bg-red-50'"
            >
              <span class="material-symbols-outlined text-[1rem]">casino</span>
            </div>
            <span class="line-clamp-2 text-[0.6rem] font-bold leading-3">
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
