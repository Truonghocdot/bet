<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()

const navItems = [
  { label: 'Trang chủ', icon: 'home', to: '/home' },
  { label: 'Hoạt động', icon: 'campaign', to: '/promotion' },
  { label: 'Nạp tiền', icon: 'redeem', to: '/deposit' },
  { label: 'Phòng chơi', icon: 'analytics', to: '/play/wingo' },
  { label: 'Tôi', icon: 'person', to: '/account' },
]

const currentTitle = computed(() => (route.meta.title as string) ?? 'FF789')

const isActive = (path: string) => {
  if (path === '/home') {
    return route.path === '/home' || route.path === '/'
  }

  if (path.startsWith('/play')) {
    return route.path.startsWith('/play')
  }

  return route.path.startsWith(path)
}
</script>

<template>
  <div class="app-shell">
    <div class="app-shell__decor">
      <div class="app-shell__blur app-shell__blur--one"></div>
      <div class="app-shell__blur app-shell__blur--two"></div>
    </div>

    <header class="glass-bar sticky top-0 z-20">
      <div class="topbar-inner">
        <button class="icon-btn icon-btn--ghost" aria-label="Mở menu">
          <span class="material-symbols-outlined">menu</span>
        </button>

        <div class="text-[1.2rem] font-extrabold italic tracking-[-0.06em] text-primary md:text-[1.35rem]">
          ff789
        </div>

        <div class="flex items-center gap-2">
          <span class="hidden text-sm font-bold text-primary md:inline">{{ currentTitle }}</span>
          <RouterLink class="icon-btn icon-btn--soft" aria-label="Thông báo" to="/notifications">
            <span class="material-symbols-outlined">notifications</span>
          </RouterLink>
        </div>
      </div>
    </header>

    <main class="app-main">
      <div class="app-container">
        <slot />
      </div>
    </main>

    <nav class="bottom-nav">
      <RouterLink
        v-for="item in navItems"
        :key="item.label"
        class="bottom-nav__item"
        :class="{ 'is-active': isActive(item.to) }"
        :to="item.to"
      >
        <span class="material-symbols-outlined mb-1">{{ item.icon }}</span>
        <span>{{ item.label }}</span>
      </RouterLink>
    </nav>

    <RouterLink class="floating-action" aria-label="Nạp nhanh" to="/deposit">
      <span class="material-symbols-outlined">add_card</span>
    </RouterLink>
  </div>
</template>
