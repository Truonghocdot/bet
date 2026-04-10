<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import { notificationItems } from '@/data/site'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const profile = computed(() => auth.user)
const affiliate = computed(() => auth.affiliateProfile)
const unreadNotifications = computed(() => notificationItems.filter((item) => item.unread).length)

function logout() {
  auth.logout()
  void router.replace('/auth')
}
</script>

<template>
  <div class="space-y-3.5 md:space-y-5">
    <section class="grid grid-cols-[auto_1fr_auto] items-center gap-3.5 rounded-[22px] bg-white p-[18px] shadow-[0_12px_32px_rgba(0,78,219,0.04)] md:p-5">
      <div class="grid h-16 w-16 place-items-center rounded-full bg-gradient-to-br from-primary to-primary-container font-extrabold text-white">
        {{ profile?.name?.slice(0, 2).toUpperCase() || 'FF' }}
      </div>
      <div>
        <h2 class="m-0 text-[1.18rem] font-extrabold">{{ profile?.name || 'Đang đồng bộ' }}</h2>
        <p class="m-0 mt-1 text-[0.8rem] text-on-surface-variant">
          ID: {{ profile?.id ?? '---' }} • {{ profile?.email || profile?.phone || 'Chưa có dữ liệu' }}
        </p>
      </div>
      <div class="rounded-[14px] bg-gradient-to-br from-[#6c5a00] to-[#fdd404] px-2.5 py-2 text-[0.72rem] font-black text-[#453700]">
        {{ affiliate ? `REF ${affiliate.ref_code}` : 'VIP' }}
      </div>
    </section>

    <section class="grid gap-2 md:grid-cols-[1.25fr_0.75fr]">
      <article class="rounded-[20px] bg-white p-[18px] shadow-[0_8px_20px_rgba(0,78,219,0.05)] md:min-h-[132px] md:p-5">
        <span class="block text-[0.72rem] font-extrabold uppercase text-on-surface-variant">Ví Của Tôi</span>
        <strong class="mt-8 block text-[1.35rem] font-black text-primary">Đang đồng bộ</strong>
        <p class="mt-2 text-[0.76rem] text-on-surface-variant">
          Số dư sẽ được đồng bộ từ API ví khi backend sẵn sàng.
        </p>
      </article>

      <div class="grid gap-2.5">
        <RouterLink to="/deposit" class="grid min-h-14 place-items-center rounded-[18px] bg-gradient-to-br from-primary to-primary-container font-extrabold text-white transition-transform active:scale-95">
          Nạp tiền
        </RouterLink>
        <button class="min-h-14 rounded-[18px] bg-white font-extrabold text-on-surface shadow-[0_8px_20px_rgba(0,78,219,0.05)] transition-transform active:scale-95">
          Rút tiền
        </button>
      </div>
    </section>

    <section class="overflow-hidden rounded-[22px] bg-white shadow-[0_8px_20px_rgba(0,78,219,0.05)]">
      <RouterLink to="/notifications" class="grid w-full grid-cols-[auto_1fr_auto_auto] items-center gap-3.5 border-b border-slate-200/60 px-4 py-3.5 text-left">
        <div class="grid h-10 w-10 place-items-center rounded-full bg-primary/10 text-primary">
          <span class="material-symbols-outlined">notifications</span>
        </div>
        <span class="font-extrabold">Thông báo</span>
        <span class="grid h-6 min-w-6 place-items-center rounded-full bg-[#b71211] px-1 text-[0.7rem] font-extrabold text-white">
          {{ unreadNotifications }}
        </span>
      </RouterLink>

      <button class="grid w-full grid-cols-[auto_1fr_auto] items-center gap-3.5 border-b border-slate-200/60 px-4 py-3.5 text-left">
        <div class="grid h-10 w-10 place-items-center rounded-full bg-[#fdd404]/20 text-[#6c5a00]">
          <span class="material-symbols-outlined" style="font-variation-settings: 'FILL' 1;">redeem</span>
        </div>
        <span class="font-extrabold">Quy đổi quà</span>
        <span class="material-symbols-outlined text-on-surface-variant">chevron_right</span>
      </button>

      <button class="grid w-full grid-cols-[auto_1fr_auto] items-center gap-3.5 border-b border-slate-200/60 px-4 py-3.5 text-left">
        <div class="grid h-10 w-10 place-items-center rounded-full bg-primary/10 text-primary">
          <span class="material-symbols-outlined">monitoring</span>
        </div>
        <span class="font-extrabold">Thống kê trò chơi</span>
        <span class="material-symbols-outlined text-on-surface-variant">chevron_right</span>
      </button>

      <button class="grid w-full grid-cols-[auto_1fr_auto_auto] items-center gap-3.5 px-4 py-3.5 text-left">
        <div class="grid h-10 w-10 place-items-center rounded-full bg-slate-100 text-on-surface-variant">
          <span class="material-symbols-outlined">language</span>
        </div>
        <span class="font-extrabold">Ngôn ngữ</span>
        <span class="text-sm font-bold text-on-surface-variant">Tiếng Việt</span>
        <span class="material-symbols-outlined text-on-surface-variant">chevron_right</span>
      </button>
    </section>

    <section class="grid grid-cols-3 gap-2.5">
      <button class="grid min-h-[88px] place-items-center gap-1.5 rounded-[18px] bg-white font-extrabold shadow-[0_8px_20px_rgba(0,78,219,0.05)]">
        <span class="material-symbols-outlined text-primary">settings</span>
        <span>Cài đặt</span>
      </button>
      <button class="grid min-h-[88px] place-items-center gap-1.5 rounded-[18px] bg-white font-extrabold shadow-[0_8px_20px_rgba(0,78,219,0.05)]">
        <span class="material-symbols-outlined text-primary">chat_bubble</span>
        <span>Góp ý</span>
      </button>
      <button class="grid min-h-[88px] place-items-center gap-1.5 rounded-[18px] bg-[rgba(126,156,255,0.16)] font-extrabold shadow-[0_8px_20px_rgba(0,78,219,0.05)]">
        <span class="material-symbols-outlined text-primary" style="font-variation-settings: 'FILL' 1;">support_agent</span>
        <span>Hỗ trợ 24/7</span>
      </button>
    </section>

      <button class="min-h-14 rounded-[18px] bg-[rgba(183,18,17,0.1)] font-black text-[#b71211] transition-transform active:scale-95" @click="logout">
        Đăng xuất tài khoản
      </button>

    <p class="mt-3 text-center text-[0.66rem] font-bold uppercase tracking-[0.18em] text-[#abadb2]">
      Phiên bản 2.4.0 • FF789 Gaming Ecosystem
    </p>
  </div>
</template>
