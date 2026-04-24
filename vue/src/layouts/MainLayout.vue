<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { RouterLink, type RouteLocationRaw, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useWalletStore } from '@/stores/wallet'
import { formatViMoney } from '@/shared/lib/money'
import { useLoading } from '@/shared/lib/loading'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const wallet = useWalletStore()
const { isLoading, setLoading } = useLoading()

const isDrawerOpen = ref(false)

const currentTitle = computed(() => (route.meta.title as string) ?? 'FF789')
const isPlayRoute = computed(() => route.path.startsWith('/play/'))

const primaryNavItems = [
  { label: 'Trang chủ', icon: 'home', to: '/home' },
  { label: 'Đại lý', icon: 'handshake', to: '/promotion', query: { tab: 'affiliate' } },
  { label: 'Ưu đãi', icon: 'redeem', to: '/promotion', query: { tab: 'promotion' } },
  { label: 'Vào chơi', icon: 'sports_esports', to: '/play' },
  { label: 'CSKH', icon: 'support_agent', to: '/cskh' },
]

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

function navigateDrawer(target: RouteLocationRaw) {
  closeDrawer()
  setLoading(true)
  void router.push(target).finally(() => {
    setTimeout(() => setLoading(false), 300)
  })
}

async function handleLogout() {
  closeDrawer()
  await auth.logout()
  void router.push('/auth')
}

async function syncRealtimeState() {
  if (!auth.isAuthenticated) {
    wallet.disconnectStream()
    wallet.reset()
    return
  }

  try {
    await wallet.fetchSummary()
  } catch {
    // wallet store keeps the current error
  }
  wallet.connectStream()
}

watch(
  () => auth.isAuthenticated,
  () => {
    void syncRealtimeState()
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  wallet.disconnectStream()
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
            <img src="/logo.png" alt="Loading..." class="relative h-20 w-20 rounded-2xl shadow-xl animate-pulse" />
          </div>
          <div class="flex items-center gap-1.5">
            <span class="h-1.5 w-1.5 animate-bounce rounded-full bg-primary" />
            <span class="h-1.5 w-1.5 animate-bounce rounded-full bg-primary [animation-delay:0.2s]" />
            <span class="h-1.5 w-1.5 animate-bounce rounded-full bg-primary [animation-delay:0.4s]" />
          </div>
        </div>
      </div>
    </Transition>

    <!-- ===== DESKTOP SIDEBAR (md+) ===== -->
    <aside class="sidebar hidden md:flex">
      <!-- Logo -->
      <RouterLink to="/home" class="sidebar__logo" @click="closeDrawer">
        <img src="/favicon.png" alt="FF789" class="h-9 w-9 rounded-[10px]" />
        <span class="text-[1.3rem] font-black italic tracking-tight text-white">FF789</span>
      </RouterLink>

      <!-- Main nav -->
      <nav class="sidebar__nav">
        <p class="sidebar__section-title">Lối tắt chính</p>
        <RouterLink
          v-for="item in primaryNavItems"
          :key="`${item.to}-${item.query?.tab ?? 'default'}`"
          :to="{ path: item.to, query: item.query }"
          class="sidebar__item"
          :class="{ 'sidebar__item--active': isNavItemActive(item) }"
        >
          <span class="material-symbols-outlined sidebar__icon">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>

      <!-- Bottom links -->
      <div class="sidebar__bottom">
        <p class="sidebar__section-title">Tiện ích</p>
        <RouterLink
          v-for="item in utilityNavItems"
          :key="item.to"
          :to="{ path: item.to }"
          class="sidebar__item"
          :class="{ 'sidebar__item--active': isActive(item.to) }"
        >
          <span class="material-symbols-outlined sidebar__icon">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </RouterLink>
        <button
          v-if="auth.isAuthenticated"
          class="sidebar__item sidebar__item--logout"
          @click="handleLogout"
        >
          <span class="material-symbols-outlined sidebar__icon">logout</span>
          <span>Đăng xuất</span>
        </button>
        <template v-else>
          <RouterLink to="/auth" class="sidebar__item">
            <span class="material-symbols-outlined sidebar__icon">login</span>
            <span>Đăng nhập</span>
          </RouterLink>
          <RouterLink to="/register" class="sidebar__item">
            <span class="material-symbols-outlined sidebar__icon">person_add</span>
            <span>Đăng ký</span>
          </RouterLink>
          <a
            :href="referralLink || '#'"
            target="_blank"
            rel="noopener noreferrer"
            class="sidebar__item"
            :class="{ 'cursor-not-allowed opacity-50': !referralLink }"
          >
            <span class="material-symbols-outlined sidebar__icon">share</span>
            <span>Link giới thiệu</span>
          </a>
        </template>
      </div>
    </aside>

    <!-- ===== MAIN COLUMN ===== -->
    <div class="app-main-col">

      <!-- ===== HEADER (non-play routes OR always on play) ===== -->
      <header v-if="!isPlayRoute" class="glass-bar sticky top-0 z-20">
        <div class="topbar-inner">
          <!-- Hamburger (mobile only) -->
          <button
            class="icon-btn icon-btn--ghost md:hidden"
            aria-label="Mở menu"
            @click="openDrawer"
          >
            <span class="material-symbols-outlined">menu</span>
          </button>

          <!-- Logo (show on mobile, hidden on md because sidebar has it) -->
          <RouterLink
            to="/home"
            class="text-[1.25rem] font-black italic tracking-[-0.06em] text-white md:hidden"
          >FF789</RouterLink>

          <!-- Page title on desktop -->
          <span class="hidden md:block text-[1rem] font-black text-white">{{ currentTitle }}</span>

          <!-- Right side actions -->
          <div class="flex items-center gap-2">
            <RouterLink
              class="icon-btn icon-btn--soft"
              aria-label="Thông báo"
              to="/notifications"
            >
              <span class="material-symbols-outlined">notifications</span>
            </RouterLink>
            <!-- Account shortcut visible on mobile header -->
            <RouterLink
              class="icon-btn icon-btn--soft md:hidden"
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
      <nav class="bottom-nav md:hidden">
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
          class="fixed inset-0 z-[60] bg-black/50 backdrop-blur-sm md:hidden"
          @click="closeDrawer"
        />
      </Transition>

      <!-- Drawer Panel -->
      <Transition name="slide-drawer">
        <div v-if="isDrawerOpen" class="drawer md:hidden">
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
