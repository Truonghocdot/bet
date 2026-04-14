<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import { formatViMoney } from '@/shared/lib/money'
import { useAuthStore } from '@/stores/auth'
import { useNotificationsStore } from '@/stores/notifications'
import { useWalletStore } from '@/stores/wallet'

const router = useRouter()
const auth = useAuthStore()
const wallet = useWalletStore()
const notifications = useNotificationsStore()

const profile = computed(() => auth.user)
const affiliate = computed(() => auth.affiliateProfile)
const unreadNotifications = computed(() => notifications.unreadCount)

const vndWallet = computed(() => wallet.wallets.find((item) => item.unit === 1) ?? null)
const usdtWallet = computed(() => wallet.wallets.find((item) => item.unit === 2) ?? null)

const walletCards = computed(() => [
  {
    unit: 1,
    label: 'Ví VND',
    symbol: 'payments',
    accent: 'from-primary to-primary-container',
    wallet: vndWallet.value,
    fractionDigits: 0,
    helper: 'Dùng cho nạp, rút và các giao dịch nội địa.',
  },
  {
    unit: 2,
    label: 'Ví USDT',
    symbol: 'currency_bitcoin',
    accent: 'from-[#ff6d66] to-[#ffd4d0]',
    wallet: usdtWallet.value,
    fractionDigits: 2,
    helper: 'Dùng cho nạp USDT và các giao dịch crypto.',
  },
])

function walletStatusLabel(status?: number | null) {
  if (status === 1) return 'Đang hoạt động'
  if (status === 2) return 'Đang khóa'
  return 'Chưa rõ'
}

function walletBalance(value: string | number | null | undefined, fractionDigits = 0) {
  return formatViMoney(value ?? 0, fractionDigits)
}

async function loadWalletSummary() {
  if (!auth.isAuthenticated) return
  try {
    await wallet.fetchSummary()
  } catch {
    // Wallet state already stores the error for the UI.
  }
}

async function loadNotificationSummary() {
  if (!auth.isAuthenticated) return
  try {
    await notifications.fetchList(1, 20)
  } catch {
    // Notifications store already keeps error state.
  }
}

function logout() {
  auth.logout()
  wallet.reset()
  notifications.reset()
  void router.replace('/auth')
}

onMounted(() => {
  void loadWalletSummary()
  void loadNotificationSummary()
})
</script>

<template>
  <div class="space-y-3.5 md:space-y-5">
    <section class="grid grid-cols-[auto_1fr_auto] items-center gap-3.5 rounded-[26px] bg-gradient-to-br from-[#ff6d66] via-[#ff867d] to-[#ffd4d0] p-[18px] text-white shadow-[0_12px_32px_rgba(255,109,102,0.16)] md:p-5">
      <div class="grid h-16 w-16 place-items-center rounded-full bg-white/18 font-extrabold text-white">
        {{ profile?.name?.slice(0, 2).toUpperCase() || 'FF' }}
      </div>
      <div>
        <h2 class="m-0 text-[1.18rem] font-extrabold">{{ profile?.name || 'Đang đồng bộ' }}</h2>
        <p class="m-0 mt-1 text-[0.8rem] text-white/86">
          ID: {{ profile?.id ?? '---' }} • {{ profile?.email || profile?.phone || 'Chưa có dữ liệu' }}
        </p>
      </div>
      <div class="rounded-[14px] bg-white/18 px-2.5 py-2 text-[0.72rem] font-black text-white">
        {{ affiliate ? `REF ${affiliate.ref_code}` : 'VIP' }}
      </div>
    </section>

    <section class="grid gap-2 md:grid-cols-2">
      <article
        v-for="item in walletCards"
        :key="item.unit"
        class="rounded-[20px] bg-white p-[18px] shadow-[0_8px_20px_rgba(255,109,102,0.05)] md:min-h-[172px] md:p-5"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <span class="block text-[0.72rem] font-extrabold uppercase text-on-surface-variant">{{ item.label }}</span>
            <strong class="mt-4 block text-[1.8rem] font-black text-primary">
              <template v-if="wallet.loading && !item.wallet">Đang đồng bộ...</template>
              <template v-else>{{ walletBalance(item.wallet?.balance, item.fractionDigits) }}</template>
            </strong>
          </div>
          <div class="grid h-12 w-12 place-items-center rounded-2xl bg-gradient-to-br text-white" :class="item.accent">
            <span class="material-symbols-outlined">{{ item.symbol }}</span>
          </div>
        </div>

        <div class="mt-4 grid gap-2 rounded-[16px] bg-slate-50 px-4 py-3 text-[0.8rem] text-on-surface-variant">
          <div class="flex items-center justify-between gap-3">
            <span>Số dư khả dụng</span>
            <span class="font-bold text-on-surface">{{ walletBalance(item.wallet?.balance, item.fractionDigits) }}</span>
          </div>
          <div class="flex items-center justify-between gap-3">
            <span>Đang khóa</span>
            <span class="font-bold text-on-surface">{{ walletBalance(item.wallet?.locked_balance, item.fractionDigits) }}</span>
          </div>
          <div class="flex items-center justify-between gap-3">
            <span>Trạng thái</span>
            <span class="font-bold text-on-surface">{{ walletStatusLabel(item.wallet?.status) }}</span>
          </div>
        </div>

        <p class="mt-3 text-[0.76rem] text-on-surface-variant">
          {{ item.helper }}
        </p>
      </article>
    </section>

    <p v-if="wallet.error" class="rounded-[16px] bg-[rgba(183,18,17,0.08)] px-4 py-3 text-sm font-semibold text-[#e64545]">
      {{ wallet.error }}
    </p>

    <section class="grid gap-2 md:grid-cols-2">
      <div class="grid gap-2.5">
        <RouterLink to="/deposit" class="grid min-h-14 place-items-center rounded-[18px] bg-gradient-to-br from-primary to-primary-container font-extrabold text-white transition-transform active:scale-95">
          Nạp tiền
        </RouterLink>
        <div class="grid grid-cols-2 gap-2.5">
          <RouterLink to="/withdraw" class="grid min-h-14 place-items-center rounded-[18px] bg-white font-extrabold text-on-surface shadow-[0_8px_20px_rgba(255,109,102,0.05)] transition-transform active:scale-95">
            Rút tiền
          </RouterLink>
          <RouterLink to="/exchange" class="grid min-h-14 place-items-center rounded-[18px] bg-white font-extrabold text-on-surface shadow-[0_8px_20px_rgba(255,109,102,0.05)] transition-transform active:scale-95">
            Chuyển tiền
          </RouterLink>
        </div>
        <p class="mt-2.5 px-1 text-[0.72rem] leading-relaxed text-on-surface-variant/80 italic">
          * Tính năng <strong>Chuyển tiền</strong> giúp bạn quy đổi tài sản qua lại giữa ví VND và USDT nhanh chóng.
        </p>
      </div>

      <article class="rounded-[20px] bg-white p-[18px] shadow-[0_8px_20px_rgba(255,109,102,0.05)] md:p-5">
        <span class="block text-[0.72rem] font-extrabold uppercase text-on-surface-variant">Thông tin ví</span>
        <div class="mt-4 grid gap-2 rounded-[16px] bg-slate-50 px-4 py-3 text-[0.8rem] text-on-surface-variant">
          <div class="flex items-center justify-between gap-3">
            <span>Ví hiển thị</span>
            <span class="font-bold text-on-surface">{{ wallet.wallets.length }}</span>
          </div>
          <div class="flex items-center justify-between gap-3">
            <span>Đồng bộ gần nhất</span>
            <span class="font-bold text-on-surface">{{ profile?.updated_at ? 'Có dữ liệu' : 'Đang chờ' }}</span>
          </div>
        </div>
      </article>
    </section>

    <section class="overflow-hidden rounded-[22px] bg-white shadow-[0_8px_20px_rgba(255,109,102,0.05)]">
      <RouterLink to="/notifications" class="grid w-full grid-cols-[auto_1fr_auto_auto] items-center gap-3.5 border-b border-slate-200/60 px-4 py-3.5 text-left">
        <div class="grid h-10 w-10 place-items-center rounded-full bg-primary/10 text-primary">
          <span class="material-symbols-outlined">notifications</span>
        </div>
        <span class="font-extrabold">Thông báo</span>
        <span class="grid h-6 min-w-6 place-items-center rounded-full bg-[#e64545] px-1 text-[0.7rem] font-extrabold text-white">
          {{ unreadNotifications }}
        </span>
      </RouterLink>

      <RouterLink to="/game-stats" class="grid w-full grid-cols-[auto_1fr_auto] items-center gap-3.5 px-4 py-3.5 text-left">
        <div class="grid h-10 w-10 place-items-center rounded-full bg-primary/10 text-primary">
          <span class="material-symbols-outlined">monitoring</span>
        </div>
        <span class="font-extrabold">Thống kê trò</span>
        <span class="material-symbols-outlined text-on-surface-variant">chevron_right</span>
      </RouterLink>
    </section>

    <button class="min-h-14 p-2 rounded-[18px] bg-[rgba(183,18,17,0.1)] font-black text-[#e64545] transition-transform active:scale-95" @click="logout">
      Đăng xuất tài khoản
    </button>

    <p class="mt-3 text-center text-[0.66rem] font-bold uppercase tracking-[0.18em] text-[#abadb2]">
      Phiên bản 2.4.0 • FF789 Gaming Ecosystem
    </p>
  </div>
</template>
