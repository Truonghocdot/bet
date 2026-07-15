<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { RouterLink, type RouteLocationRaw, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useNotificationsStore } from '@/stores/notifications'
import { useWalletStore } from '@/stores/wallet'
import { formatViMoney } from '@/shared/lib/money'
import { useLoading } from '@/shared/lib/loading'
import { request } from '@/shared/api/http'
import bottomNavLeftArt from '@/assets/bottom/icon_btm_jr.avif'
import bottomNavRightArt from '@/assets/bottom/icon_btm_jr2.avif'
import defaultHeaderLogo from '@/assets/logo-mobile.webp'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const notifications = useNotificationsStore()
const wallet = useWalletStore()
const { isLoading, setLoading } = useLoading()

const isDrawerOpen = ref(false)
type PopupSlot = 'message' | 'latest_news'
type PopupItem = {
  slot: PopupSlot
  title: string
  content: string
}

const popupQueue = ref<PopupItem[]>([])

const currentTitle = computed(() => (route.meta.title as string) ?? 'fh88u')
const isPlayRoute = computed(() => route.path.startsWith('/play/'))
const activePopup = computed(() => popupQueue.value[0] ?? null)
const isLatestNewsPopup = computed(() => activePopup.value?.slot === 'latest_news')
const activePopupHtml = computed(() => {
  const content = normalizePopupContent(activePopup.value?.content)
  if (!content) return ''

  if (/<\/?[a-z][\s\S]*>/i.test(content)) {
    return content
  }

  return escapeHtml(content).replace(/\r\n|\r|\n/g, '<br>')
})

const primaryNavItems = [
  { label: 'Trang chủ', icon: 'home', to: '/home' },
  { label: 'Đại lý', icon: 'handshake', to: '/promotion', query: { tab: 'affiliate' } },
  { label: 'Ưu đãi', icon: 'redeem', to: '/promotion', query: { tab: 'promotion' } },
  { label: 'Vào chơi', icon: 'sports_esports', to: '/play' },
  { label: 'CSKH', icon: 'support_agent', to: '/cskh' },
]

const bottomNavGridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${primaryNavItems.length}, minmax(0, 1fr))`,
}))

const utilityNavItems = [
  { label: 'Nạp tiền', icon: 'add_card', to: '/deposit' },
  { label: 'Tài khoản', icon: 'manage_accounts', to: '/account' },
  { label: 'Thông báo', icon: 'notifications', to: '/notifications' },
]

const historyShortcutItems = [
  { label: 'Lịch sử nạp', icon: 'payments', to: '/deposit', query: { section: 'history' } },
  { label: 'Lịch sử rút', icon: 'history', to: '/withdraw', query: { section: 'history' } },
]

const isActive = (path: string) => {
  if (path === '/home') return route.path === '/home' || route.path === '/'
  if (path === '/play') return route.path.startsWith('/play')
  return route.path.startsWith(path)
}

const isNavItemActive = (item: { to: string; query?: Record<string, string> }) => {
  if (!isActive(item.to)) return false

  if (item.to === '/promotion' && item.query?.tab) {
    return String(route.query.tab ?? 'affiliate') === item.query.tab
  }

  return true
}

const vndBalance = computed(() => {
  const w = wallet.wallets.find((item) => item.unit === 1)
  return formatViMoney(w?.balance ?? 0, 0)
})

const userName = computed(() => auth.user?.name ?? 'Khách')
const unreadNotificationCount = computed(() => notifications.unreadCount)
const unreadNotificationBadge = computed(() => {
  if (unreadNotificationCount.value <= 0) return ''
  return unreadNotificationCount.value > 99 ? '99+' : String(unreadNotificationCount.value)
})
const uploadedHeaderLogoSrc = computed(() => String(wallet.summary?.app_header_logo_url ?? '').trim())
const isSafariBrowser = computed(() => {
  if (typeof navigator === 'undefined') return false

  const userAgent = navigator.userAgent || ''

  return /Safari/i.test(userAgent) && !/Chrome|Chromium|CriOS|FxiOS|EdgiOS|OPiOS|SamsungBrowser|Android/i.test(userAgent)
})
const safariHeaderLogoFallbackSrc = computed(() => uploadedHeaderLogoSrc.value.replace(/\.avif(?=([?#].*)?$)/i, '.webp'))
const headerLogoSourceIndex = ref(0)
const headerLogoCandidateSources = computed(() => {
  if (!uploadedHeaderLogoSrc.value) return []

  if (isSafariBrowser.value && /\.avif(?=([?#].*)?$)/i.test(uploadedHeaderLogoSrc.value)) {
    return [safariHeaderLogoFallbackSrc.value, uploadedHeaderLogoSrc.value].filter(Boolean)
  }

  return [uploadedHeaderLogoSrc.value]
})
const resolvedHeaderLogoSrc = computed(() => headerLogoCandidateSources.value[headerLogoSourceIndex.value] ?? '')
const loadingLogoSrc = computed(() => resolvedHeaderLogoSrc.value || defaultHeaderLogo)
watch(headerLogoCandidateSources, () => {
  headerLogoSourceIndex.value = 0
}, { immediate: true })

function handleHeaderLogoError() {
  if (headerLogoSourceIndex.value < headerLogoCandidateSources.value.length - 1) {
    headerLogoSourceIndex.value += 1
    return
  }

  headerLogoSourceIndex.value = -1
}

const referralLink = computed(() => auth.affiliateProfile?.ref_link || '')

function copyReferralLink() {
  if (!referralLink.value) return
  navigator.clipboard.writeText(referralLink.value).then(() => {
    console.log('Referral link copied to clipboard')
  }).catch((err) => {
    console.error('Failed to copy:', err)
  })
}

function openDrawer() { isDrawerOpen.value = true }
function closeDrawer() { isDrawerOpen.value = false }

function popupStorageKey(slot: PopupSlot): string {
  return `fh88u:popup:dismissed:${auth.user?.id ?? 0}:${slot}`
}

function readDismissedPopup(slot: PopupSlot): string {
  try {
    return window.sessionStorage.getItem(popupStorageKey(slot)) ?? ''
  } catch {
    return ''
  }
}

function saveDismissedPopup(slot: PopupSlot, content: string) {
  try {
    window.sessionStorage.setItem(popupStorageKey(slot), content)
  } catch {
    // no-op
  }
}

function normalizePopupContent(value: string | null | undefined): string {
  return String(value ?? '').trim()
}

function escapeHtml(value: string): string {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;')
}

function popupSignature(item: PopupItem): string {
  return `${item.slot}:${item.content}`
}

function buildPopupQueue(): PopupItem[] {
  const nextQueue: PopupItem[] = []
  const message = normalizePopupContent(wallet.summary?.popup?.message)
  const latestNews = normalizePopupContent(wallet.summary?.popup?.latest_news)

  if (message && readDismissedPopup('message') !== message) {
    nextQueue.push({
      slot: 'message',
      title: 'Thông báo',
      content: message,
    })
  }

  if (latestNews && readDismissedPopup('latest_news') !== latestNews) {
    nextQueue.push({
      slot: 'latest_news',
      title: 'Tin tức mới nhất',
      content: latestNews,
    })
  }

  return nextQueue
}

function syncPopupQueue() {
  const nextQueue = buildPopupQueue()
  const currentSignature = popupQueue.value.map(popupSignature).join('|')
  const nextSignature = nextQueue.map(popupSignature).join('|')

  if (currentSignature === nextSignature || activePopup.value) {
    return
  }

  popupQueue.value = nextQueue
}

function closeActivePopup() {
  const popup = activePopup.value
  if (!popup) return

  saveDismissedPopup(popup.slot, popup.content)
  popupQueue.value.shift()

  if (!popupQueue.value.length) {
    syncPopupQueue()
  }
}

function navigateDrawer(target: RouteLocationRaw) {
  closeDrawer()
  setLoading(true)
  void router.push(target).finally(() => {
    setTimeout(() => setLoading(false), 300)
  })
}

function navigatePrimaryClick(event: MouseEvent, target: RouteLocationRaw) {
  if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return

  event.preventDefault()
  closeDrawer()
  void router.push(target)
}

async function handleLogout() {
  closeDrawer()
  if (auth.accessToken) {
    try {
      await request('POST', '/v1/provider-games/tcg/close-active', {
        token: auth.accessToken,
        timeoutMs: 10000,
      })
    } catch {
      // best effort
    }
  }
  await auth.logout()
  void router.push('/auth')
}

async function syncRealtimeState() {
  wallet.disconnectStream()
  notifications.disconnectStream()

  try {
    await wallet.fetchSummary()
  } catch {
    // wallet store keeps the current error
  }

  try {
    await notifications.fetchList(1, notifications.pagination.pageSize)
  } catch {
    // notifications store keeps the current error
  }

  if (auth.isAuthenticated) {
    wallet.connectStream()
    notifications.connectStream(1, notifications.pagination.pageSize)
  }
}

watch(
  () => auth.isAuthenticated,
  () => {
    void syncRealtimeState()
    syncPopupQueue()
  },
  { immediate: true },
)

watch(
  () => [wallet.summary?.popup?.message, wallet.summary?.popup?.latest_news] as const,
  () => {
    syncPopupQueue()
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  wallet.disconnectStream()
  notifications.disconnectStream()
})
</script>

<template>
  <div class="app-shell">
    <!-- ===== BACKGROUND DECOR ===== -->
    <div class="app-shell__decor" aria-hidden="true">
      <div class="app-shell__blur app-shell__blur--one" />
      <div class="app-shell__blur app-shell__blur--two" />
    </div>

    <!-- ===== GLOBAL LOADING OVERLAY ===== -->
    <Transition name="fade">
      <div v-if="isLoading" class="fixed inset-0 z-[100] grid place-items-center bg-white/80 backdrop-blur-md">
          <div class="flex flex-col items-center gap-4">
          <div class="relative">
            <div class="absolute inset-0 animate-ping rounded-full bg-primary/20" />
            <img :src="loadingLogoSrc" alt="Loading..." class="relative h-20 w-20 rounded-2xl object-contain shadow-xl animate-pulse" />
          </div>
          <div class="flex items-center gap-1.5">
            <span class="h-1.5 w-1.5 animate-bounce rounded-full bg-primary" />
            <span class="h-1.5 w-1.5 animate-bounce rounded-full bg-primary [animation-delay:0.2s]" />
            <span class="h-1.5 w-1.5 animate-bounce rounded-full bg-primary [animation-delay:0.4s]" />
          </div>
        </div>
      </div>
    </Transition>

    <Transition name="fade">
      <div v-if="activePopup" class="fixed inset-0 z-[90] grid place-items-center bg-black/45 px-4 backdrop-blur-sm">
        <div
          :class="[
            'w-full overflow-hidden bg-white shadow-[0_20px_60px_rgba(15,23,42,0.25)]',
            isLatestNewsPopup
              ? 'max-w-[360px] rounded-[6px] border border-slate-200'
              : 'max-w-[540px] rounded-[24px]',
          ]"
        >
          <div
            :class="[
              'relative',
              isLatestNewsPopup ? 'border-b border-slate-200 px-5 py-4' : 'flex items-start justify-between gap-4 p-5',
            ]"
          >
            <div v-if="!isLatestNewsPopup">
              <p class="text-[0.72rem] font-black uppercase tracking-[0.12em] text-primary/70">{{ activePopup.title }}</p>
              <h3 class="mt-1 text-[1.1rem] font-black text-on-surface">fh88u</h3>
            </div>
            <h3
              v-else
              class="text-center text-[1.05rem] font-black uppercase tracking-[0.04em] text-slate-800"
            >
              {{ activePopup.title }}
            </h3>
            <button
              type="button"
              :class="[
                'grid place-items-center transition-transform active:scale-95',
                isLatestNewsPopup
                  ? 'absolute right-2 top-2 h-8 w-8 text-slate-300'
                  : 'h-10 w-10 rounded-full bg-slate-100 text-slate-500',
              ]"
              @click="closeActivePopup"
            >
              <span class="material-symbols-outlined" :class="isLatestNewsPopup ? 'text-[1rem]' : 'text-[1.1rem]'">close</span>
            </button>
          </div>

          <div
            :class="[
              'max-h-[58vh] overflow-y-auto',
              isLatestNewsPopup ? 'px-3 py-3' : 'px-5 pb-5',
            ]"
          >
            <div
              :class="[
                'popup-html text-slate-700',
                isLatestNewsPopup
                  ? 'rounded-[4px] border border-slate-200 bg-white px-3 py-2 text-[0.88rem] leading-6 [&_*]:break-words [&_a]:font-semibold [&_a]:text-primary [&_img]:my-2 [&_img]:w-full [&_img]:rounded-[4px]'
                  : 'rounded-[20px] bg-gradient-to-br from-primary/8 to-primary/3 p-4 text-[0.9rem] leading-6 [&_*]:break-words [&_a]:font-semibold [&_a]:text-primary [&_img]:my-3 [&_img]:w-full [&_img]:rounded-[14px] [&_ol]:pl-5 [&_p]:mb-3 [&_p:last-child]:mb-0 [&_ul]:pl-5',
              ]"
              v-html="activePopupHtml"
            />

            <div
              :class="[
                'flex',
                isLatestNewsPopup ? 'justify-end border-t border-slate-200 bg-slate-50 px-0 pb-0 pt-3' : 'mt-5 justify-end',
              ]"
            >
              <button
                type="button"
                :class="[
                  'transition-transform active:scale-95',
                  isLatestNewsPopup
                    ? 'min-h-10 rounded-[6px] border border-slate-300 bg-white px-4 text-[0.8rem] font-bold text-slate-700'
                    : 'min-h-11 rounded-[14px] bg-primary px-5 text-[0.82rem] font-black text-white',
                ]"
                @click="closeActivePopup"
              >
                {{ isLatestNewsPopup ? 'Đóng' : 'Đã hiểu' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- ===== MAIN COLUMN ===== -->
    <div class="app-main-col">

      <!-- ===== HEADER (non-play routes OR always on play) ===== -->
      <header v-if="!isPlayRoute" class="glass-bar sticky top-0 z-20">
        <div class="topbar-inner">
          <div class="topbar-inner__side topbar-inner__side--left">
            <button
              class="icon-btn icon-btn--ghost"
              aria-label="Mở menu"
              @click="openDrawer"
            >
              <span class="material-symbols-outlined">menu</span>
            </button>

            <RouterLink
              to="/home"
              class="topbar-home-brand"
            >
              <img :src="defaultHeaderLogo" alt="fh88u" class="topbar-home-brand__logo" />
            </RouterLink>
          </div>

          <RouterLink
            v-if="uploadedHeaderLogoSrc && resolvedHeaderLogoSrc"
            to="/home"
            class="topbar-brand"
          >
            <img
              :key="resolvedHeaderLogoSrc"
              :src="resolvedHeaderLogoSrc"
              alt="Logo app"
              class="topbar-brand__logo topbar-brand__logo--custom"
              @error="handleHeaderLogoError"
            />
          </RouterLink>

          <!-- Right side actions -->
          <div class="topbar-inner__side topbar-inner__side--right">
            <RouterLink
              class="icon-btn icon-btn--soft icon-btn--badge topbar-action-btn"
              aria-label="Thông báo"
              to="/notifications"
              @click="navigatePrimaryClick($event, { path: '/notifications' })"
            >
              <span class="material-symbols-outlined">notifications</span>
              <span v-if="unreadNotificationBadge" class="icon-btn__badge">{{ unreadNotificationBadge }}</span>
            </RouterLink>
            <RouterLink
              class="icon-btn icon-btn--soft topbar-action-btn"
              aria-label="Tài khoản"
              to="/account"
            >
              <span class="material-symbols-outlined">person</span>
            </RouterLink>
          </div>
        </div>
        <div class="mt-3 grid grid-cols-2 gap-2">
          <RouterLink
            v-for="item in historyShortcutItems"
            :key="item.label"
            :to="{ path: item.to, query: item.query }"
            class="flex min-h-10 items-center justify-center gap-1.5 rounded-[14px] bg-white/14 px-3 text-[0.74rem] font-black text-white backdrop-blur-sm transition-transform active:scale-[0.98]"
          >
            <span class="material-symbols-outlined text-[1rem]">{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </RouterLink>
        </div>
      </header>

      <!-- ===== PAGE CONTENT ===== -->
      <main class="app-main">
        <div class="app-container">
          <slot />
        </div>
      </main>

      <!-- ===== BOTTOM NAV (mobile 5-tab) ===== -->
      <nav class="bottom-nav" :style="bottomNavGridStyle">
        <img :src="bottomNavLeftArt" alt="" class="bottom-nav__side-art bottom-nav__side-art--left" aria-hidden="true" />
        <img :src="bottomNavRightArt" alt="" class="bottom-nav__side-art bottom-nav__side-art--right" aria-hidden="true" />
        <RouterLink
          v-for="item in primaryNavItems"
          :key="`${item.to}-${item.query?.tab ?? 'default'}`"
          :to="{ path: item.to, query: item.query }"
          class="bottom-nav__item"
          :class="{ 'is-active': isNavItemActive(item) }"
        >
          <span class="material-symbols-outlined bottom-nav__icon">{{ item.icon }}</span>
          <span class="bottom-nav__label">{{ item.label }}</span>
        </RouterLink>
      </nav>
    </div>

    <!-- ===== DRAWER OVERLAY ===== -->
    <Teleport to="body">
      <!-- Backdrop -->
      <Transition name="fade">
        <div
          v-if="isDrawerOpen"
          class="fixed inset-0 z-[60] bg-black/50 backdrop-blur-sm"
          @click="closeDrawer"
        />
      </Transition>

      <!-- Drawer Panel -->
      <Transition name="slide-drawer">
        <div v-if="isDrawerOpen" class="drawer">
          <!-- Drawer Header / User Info -->
          <div class="drawer__header">
            <button class="absolute right-4 top-4 grid h-8 w-8 place-items-center rounded-full bg-white/15 text-white" @click="closeDrawer">
              <span class="material-symbols-outlined text-[1.1rem]">close</span>
            </button>
            <div class="flex items-center gap-3">
              <div class="grid h-12 w-12 place-items-center rounded-full bg-white/20 text-[1.5rem] flex-shrink-0">
                <span class="material-symbols-outlined text-[1.5rem] text-white">person</span>
              </div>
              <div>
                <p class="text-[0.72rem] text-white/70">Xin chào</p>
                <p class="font-black text-white text-[1rem]">{{ userName }}</p>
                <p class="text-[0.68rem] text-white/60">Số dư: {{ vndBalance }}đ</p>
              </div>
            </div>
          </div>

          <!-- Drawer Nav -->
          <div class="drawer__body">
            <p class="drawer__section-title">Lối tắt chính</p>
            <button
              v-for="item in primaryNavItems"
              :key="`${item.to}-${item.query?.tab ?? 'default'}`"
              class="drawer__item"
              :class="{ 'drawer__item--active': isNavItemActive(item) }"
              @click="navigateDrawer({ path: item.to, query: item.query })"
            >
              <span
                class="material-symbols-outlined text-[1.2rem]"
                :class="isNavItemActive(item) ? 'text-white' : 'text-white/55'"
              >{{ item.icon }}</span>
              <span>{{ item.label }}</span>
              <span class="material-symbols-outlined ml-auto text-[1rem] text-white/35">chevron_right</span>
            </button>

            <div class="drawer__divider" />

            <p class="drawer__section-title">Tiện ích</p>
            <button
              v-for="item in utilityNavItems"
              :key="item.to"
              class="drawer__item"
              :class="{ 'drawer__item--active': isActive(item.to) }"
              @click="navigateDrawer({ path: item.to })"
            >
              <span
                class="material-symbols-outlined text-[1.2rem]"
                :class="isActive(item.to) ? 'text-white' : 'text-white/55'"
              >{{ item.icon }}</span>
              <span>{{ item.label }}</span>
              <span class="material-symbols-outlined ml-auto text-[1rem] text-white/35">chevron_right</span>
            </button>

            <div class="drawer__divider" />

            <button
              v-if="auth.isAuthenticated"
              class="drawer__item text-[#e64545]"
              @click="handleLogout"
            >
              <span class="material-symbols-outlined text-[1.2rem] text-[#e64545]">logout</span>
              <span>Đăng xuất</span>
            </button>
            <template v-else>
              <button
              class="drawer__item text-primary"
                @click="navigateDrawer({ path: '/auth' })"
              >
                <span class="material-symbols-outlined text-[1.2rem] text-primary">login</span>
                <span>Đăng nhập</span>
              </button>
              <button
              class="drawer__item text-primary"
                @click="navigateDrawer({ path: '/register' })"
              >
                <span class="material-symbols-outlined text-[1.2rem] text-primary">person_add</span>
                <span>Đăng ký</span>
              </button>
              <a
                :href="referralLink || '#'"
                target="_blank"
                rel="noopener noreferrer"
                class="drawer__item text-primary"
                :class="{ 'cursor-not-allowed opacity-50': !referralLink }"
                @click="closeDrawer"
              >
                <span class="material-symbols-outlined text-[1.2rem] text-primary">share</span>
                <span class="flex-1 text-left">Link giới thiệu</span>
                <span class="material-symbols-outlined  text-[1rem] text-slate-300">open_in_new</span>
              </a>
            </template>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
