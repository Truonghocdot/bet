<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useWalletStore } from '@/stores/wallet'
import { formatViMoney } from '@/shared/lib/money'
import { useLoading } from '@/shared/lib/loading'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const wallet = useWalletStore()
const { isLoading } = useLoading()

const isDrawerOpen = ref(false)

const currentTitle = computed(() => (route.meta.title as string) ?? 'FF789')
const isPlayRoute = computed(() => route.path.startsWith('/play/'))

// Bottom nav tabs (exclude center special)
const navLeft = [
  { label: 'Trang chủ', icon: 'home', iconFilled: 'home', to: '/home' },
  { label: 'Hoạt động', icon: 'redeem', iconFilled: 'redeem', to: '/promotion' },
]
const navRight = [
  { label: 'CSKH', icon: 'headphones', iconFilled: 'headphones', to: '/cskh' },
  { label: 'Tôi', icon: 'person', iconFilled: 'person', to: '/account' },
]

// Sidebar / drawer nav (all items)
const sidebarItems = [
  { label: 'Trang chủ', icon: 'home', to: '/home' },
  { label: 'Hoạt động', icon: 'redeem', to: '/promotion' },
  { label: 'Phòng chơi', icon: 'casino', to: '/play' },
  { label: 'Nạp tiền', icon: 'add_card', to: '/deposit' },
  { label: 'CSKH', icon: 'headphones', to: '/cskh' },
]
const sidebarBottom = [
  { label: 'Tài khoản', icon: 'manage_accounts', to: '/account' },
  { label: 'Thông báo', icon: 'notifications', to: '/notifications' },
]

const isActive = (path: string) => {
  if (path === '/home') return route.path === '/home' || route.path === '/'
  if (path === '/play') return route.path.startsWith('/play')
  return route.path.startsWith(path)
}

const vndBalance = computed(() => {
  const w = wallet.wallets.find((item) => item.unit === 1)
  return formatViMoney(w?.balance ?? 0, 0)
})

const userName = computed(() => auth.user?.name ?? 'Khách')

function openDrawer() { isDrawerOpen.value = true }
function closeDrawer() { isDrawerOpen.value = false }

function navigateDrawer(path: string) {
  closeDrawer()
  setLoading(true)
  void router.push(path).finally(() => {
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
        <RouterLink
          v-for="item in sidebarItems"
          :key="item.to"
          :to="item.to"
          class="sidebar__item"
          :class="{ 'sidebar__item--active': isActive(item.to) }"
        >
          <span class="material-symbols-outlined sidebar__icon">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </RouterLink>
      </nav>

      <!-- Bottom links -->
      <div class="sidebar__bottom">
        <RouterLink
          v-for="item in sidebarBottom"
          :key="item.to"
          :to="item.to"
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
        <RouterLink v-else to="/auth" class="sidebar__item">
          <span class="material-symbols-outlined sidebar__icon">login</span>
          <span>Đăng nhập</span>
        </RouterLink>
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
      </header>

      <!-- ===== PAGE CONTENT ===== -->
      <main class="app-main">
        <div class="app-container">
          <slot />
        </div>
      </main>

      <!-- ===== BOTTOM NAV (mobile 5-tab) ===== -->
      <nav class="bottom-nav md:hidden">
        <!-- Left 2 tabs -->
        <RouterLink
          v-for="item in navLeft"
          :key="item.to"
          :to="item.to"
          class="bottom-nav__item"
          :class="{ 'is-active': isActive(item.to) }"
        >
          <span class="material-symbols-outlined mb-0.5 text-[1.4rem]">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
        </RouterLink>

        <!-- Center special button (Phòng chơi) -->
        <RouterLink
          to="/play"
          class="bottom-nav__item bottom-nav__special-wrap"
          :class="{ 'is-active': isActive('/play') }"
          aria-label="Phòng chơi"
        >
          <div class="bottom-nav__special" :class="{ 'bottom-nav__special--active': isActive('/play') }">
            <span class="material-symbols-outlined text-[1.5rem] text-white">casino</span>
          </div>
          <span class="mt-1 text-[0.6rem] font-bold" :class="isActive('/play') ? 'text-primary' : 'text-slate-500'">
            Phòng chơi
          </span>
        </RouterLink>

        <!-- Right 2 tabs -->
        <RouterLink
          v-for="item in navRight"
          :key="item.to"
          :to="item.to"
          class="bottom-nav__item"
          :class="{ 'is-active': isActive(item.to) }"
        >
          <span class="material-symbols-outlined mb-0.5 text-[1.4rem]">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
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
            <p class="px-4 pt-4 pb-1 text-[0.65rem] font-black uppercase tracking-wider text-slate-400">Menu chính</p>
            <button
              v-for="item in sidebarItems"
              :key="item.to"
              class="drawer__item"
              :class="{ 'drawer__item--active': isActive(item.to) }"
              @click="navigateDrawer(item.to)"
            >
              <span
                class="material-symbols-outlined text-[1.2rem]"
                :class="isActive(item.to) ? 'text-primary' : 'text-slate-400'"
              >{{ item.icon }}</span>
              <span>{{ item.label }}</span>
              <span class="material-symbols-outlined ml-auto text-[1rem] text-slate-300">chevron_right</span>
            </button>

            <div class="mx-4 my-3 h-px bg-slate-100" />

            <p class="px-4 pb-1 text-[0.65rem] font-black uppercase tracking-wider text-slate-400">Tài khoản</p>
            <button
              v-for="item in sidebarBottom"
              :key="item.to"
              class="drawer__item"
              :class="{ 'drawer__item--active': isActive(item.to) }"
              @click="navigateDrawer(item.to)"
            >
              <span
                class="material-symbols-outlined text-[1.2rem]"
                :class="isActive(item.to) ? 'text-primary' : 'text-slate-400'"
              >{{ item.icon }}</span>
              <span>{{ item.label }}</span>
              <span class="material-symbols-outlined ml-auto text-[1rem] text-slate-300">chevron_right</span>
            </button>

            <div class="mx-4 my-3 h-px bg-slate-100" />

            <button
              v-if="auth.isAuthenticated"
              class="drawer__item text-[#e64545]"
              @click="handleLogout"
            >
              <span class="material-symbols-outlined text-[1.2rem] text-[#e64545]">logout</span>
              <span>Đăng xuất</span>
            </button>
            <button
              v-else
              class="drawer__item text-primary"
              @click="navigateDrawer('/auth')"
            >
              <span class="material-symbols-outlined text-[1.2rem] text-primary">login</span>
              <span>Đăng nhập</span>
            </button>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
