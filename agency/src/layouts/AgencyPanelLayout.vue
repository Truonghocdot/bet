<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAgencyAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAgencyAuthStore()
const copied = ref(false)

const navigationItems = computed(() => [
  {
    label: 'Tổng quan',
    icon: 'dashboard',
    to: { name: 'agency-overview' },
    active: route.name === 'agency-overview',
  },
  {
    label: 'Người chơi',
    icon: 'group',
    to: { name: 'agency-users' },
    active: route.name === 'agency-users' || route.name === 'agency-user-stats' || route.name === 'agency-user-deposits' || route.name === 'agency-user-withdrawals',
  },
])

async function logout() {
  auth.logout()
  await router.replace({ name: 'agency-login' })
}

async function copyAffiliateLink() {
  const link = auth.affiliateProfile?.ref_link?.trim()
  if (!link) return
  try {
    await navigator.clipboard.writeText(link)
    copied.value = true
    window.setTimeout(() => {
      copied.value = false
    }, 1500)
  } catch {
    copied.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-[#f3f4f6] text-slate-900">
    <div class="grid min-h-screen lg:grid-cols-[272px_minmax(0,1fr)]">
      <aside class="border-b border-slate-200 bg-[#111827] text-white lg:border-b-0 lg:border-r lg:border-slate-800">
        <div class="flex items-center justify-between px-5 py-4 lg:px-6">
          <div>
            <p class="m-0 text-[0.68rem] font-bold uppercase tracking-[0.18em] text-slate-400">FH88U</p>
            <h1 class="m-0 mt-1 text-lg font-bold">Agency Panel</h1>
          </div>
          <span class="material-symbols-outlined text-slate-500">shield_lock</span>
        </div>

        <div class="px-4 pb-4 lg:px-5">
          <div class="rounded-2xl border border-white/10 bg-white/5 px-4 py-4">
            <p class="m-0 text-xs font-semibold text-slate-400">Tài khoản</p>
            <strong class="mt-2 block text-sm font-bold">{{ auth.user?.name || 'Agency' }}</strong>
            <p class="m-0 mt-1 text-xs text-slate-400">Ref {{ auth.affiliateProfile?.ref_code || '—' }}</p>
          </div>
        </div>

        <nav class="grid gap-1 px-3 pb-5">
          <RouterLink
            v-for="item in navigationItems"
            :key="item.label"
            :to="item.to"
            class="flex items-center gap-3 rounded-2xl px-4 py-3 text-sm font-semibold transition-colors"
            :class="item.active ? 'bg-white text-slate-950' : 'text-slate-300 hover:bg-white/8 hover:text-white'"
          >
            <span class="material-symbols-outlined text-[1.1rem]">{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </RouterLink>
        </nav>
      </aside>

      <div class="min-w-0">
        <header class="sticky top-0 z-10 border-b border-slate-200 bg-white/90 backdrop-blur">
          <div class="flex flex-wrap items-center justify-between gap-3 px-4 py-4 md:px-6">
            <div>
              <p class="m-0 text-xs font-bold uppercase tracking-[0.14em] text-slate-400">Bảng quản trị</p>
              <h2 class="m-0 mt-1 text-xl font-bold text-slate-950">{{ route.meta.title }}</h2>
            </div>

            <div class="flex items-center gap-3">
              <div class="hidden rounded-full bg-slate-100 px-3 py-2 text-xs font-semibold text-slate-500 md:block">
                {{ auth.user?.phone || auth.user?.email || 'Agency account' }}
              </div>
              <button
                type="button"
                class="rounded-xl border border-black/8 bg-white px-4 py-2.5 text-sm font-bold text-slate-700 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="!auth.affiliateProfile?.ref_link"
                @click="copyAffiliateLink"
              >
                {{ copied ? 'Đã copy link' : 'Copy link affiliate' }}
              </button>
              <button
                type="button"
                class="rounded-xl bg-slate-950 px-4 py-2.5 text-sm font-bold text-white"
                @click="logout"
              >
                Đăng xuất
              </button>
            </div>
          </div>
        </header>

        <div class="px-4 py-6 md:px-6">
          <slot />
        </div>
      </div>
    </div>
  </div>
</template>
