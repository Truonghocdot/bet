<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'

import BannerCarousel from '@/components/BannerCarousel.vue'
import MarqueeBar from '@/components/MarqueeBar.vue'
import Leaderboard from '@/components/Leaderboard.vue'
import { formatViMoney } from '@/shared/lib/money'
import { useAuthStore } from '@/stores/auth'
import { useWalletStore } from '@/stores/wallet'
import { gameRooms } from '@/data/site'

const auth = useAuthStore()
const wallet = useWalletStore()

const greetingName = computed(() => auth.user?.name || 'Bạn')
const featuredRooms = computed(() => gameRooms.filter((g) => g.featured))
const vndWallet = computed(() => wallet.wallets.find((item) => item.unit === 1) ?? null)

function displayBalance(value: string | number | null | undefined) {
  return formatViMoney(value ?? 0, 0)
}

const gameCards = [
  {
    name: 'Win Go',
    desc: 'Đoán số Xanh/Đỏ/Tím',
    route: '/play/wingo',
    gradient: 'from-[#ff6d66] to-[#e52e2e]',
    icon: 'rocket_launch',
  },
  {
    name: 'K3',
    desc: 'Lớn/Nhỏ/Lẻ/Chẵn',
    route: '/play/k3',
    gradient: 'from-[#f59e0b] to-[#ef4444]',
    icon: 'casino',
  },
  {
    name: '5D Lô Tô',
    desc: 'Chọn số 5 vị trí',
    route: '/play/lottery',
    gradient: 'from-[#8b5cf6] to-[#6366f1]',
    icon: 'looks_5',
  },
  {
    name: 'Trx Win',
    desc: 'Sắp ra mắt',
    route: '/play/trx_win',
    gradient: 'from-[#10b981] to-[#059669]',
    icon: 'currency_bitcoin',
  },
]

const categoryTabs = [
  { label: 'Xổ số', icon: 'confirmation_number', accent: '#ff6d66' },
  { label: 'Casino', icon: 'casino', accent: '#f6c32d' },
  { label: 'Bắn cá', icon: 'skull', accent: '#e64545' },
  { label: 'Thể thao', icon: 'sports_soccer', accent: '#24b561' },
  { label: 'Game bài', icon: 'playing_cards', accent: '#8b5cf6' },
]

const winInfo = [
  { user: 'Mem***AXG', game: 'K3 Lotre', amount: '+895,200đ', avatar: 21 },
  { user: 'Mem***DMT', game: 'Win Go', amount: '+342,000đ', avatar: 22 },
  { user: 'Mem***ZHS', game: '5D Lô Tô', amount: '+1,280,000đ', avatar: 23 },
  { user: 'Mem***QRT', game: 'K3 Lotre', amount: '+210,000đ', avatar: 24 },
  { user: 'Mem***UHE', game: 'Win Go', amount: '+560,000đ', avatar: 25 },
  { user: 'Mem***FVT', game: '5D Lô Tô', amount: '+980,000đ', avatar: 26 },
]

onMounted(() => {
  void wallet.fetchSummary()
})
</script>

<template>
  <div class="space-y-0 pb-2">

    <!-- ===== TOP HEADER (inline, no layout header for home since MainLayout shows it) ===== -->
    <!-- MarqueeBar -->
    <MarqueeBar />

    <!-- ===== BANNER CAROUSEL ===== -->
    <BannerCarousel />

    <!-- ===== CATEGORY QUICK TABS ===== -->
    <div class="flex gap-2 overflow-x-auto px-3 py-3 no-scrollbar">
      <button
        v-for="tab in categoryTabs"
        :key="tab.label"
        class="flex flex-shrink-0 flex-col items-center gap-1.5 rounded-[16px] bg-white px-3 py-2.5 shadow-sm border border-slate-100 transition-transform active:scale-95"
      >
        <span
          class="grid h-9 w-9 place-items-center rounded-full text-xl"
          :style="{ backgroundColor: `${tab.accent}18`, color: tab.accent }"
        >
          <span class="material-symbols-outlined text-[1.1rem]">{{ tab.icon }}</span>
        </span>
        <span class="text-[0.65rem] font-bold text-on-surface">{{ tab.label }}</span>
      </button>
    </div>

    <!-- ===== WALLET CARD ===== -->
    <div class="mx-3 overflow-hidden rounded-[20px] bg-gradient-to-br from-[#ff6d66] via-[#ff867d] to-[#ffd4d0] p-4 text-white shadow-[0_12px_30px_rgba(255,109,102,0.2)]">
      <div class="absolute inset-0 pointer-events-none bg-[radial-gradient(circle_at_top_right,rgba(255,255,255,0.22),transparent_26%)]" />
      <div class="relative">
        <p class="text-[0.7rem] uppercase tracking-[0.12em] text-white/72">Số dư ví VND</p>
        <strong class="mt-1 block text-[1.6rem] font-black">
          {{ vndWallet ? displayBalance(vndWallet.balance) : '0' }}đ
        </strong>
        <p class="text-[0.68rem] text-white/70 mt-0.5">Chào {{ greetingName }} 👋</p>
        <div class="mt-3 grid grid-cols-2 gap-2">
          <RouterLink
            to="/account"
            class="flex items-center justify-center gap-1.5 rounded-full border-2 border-white/40 bg-white/10 py-2.5 text-[0.82rem] font-black text-white active:scale-95 transition-transform"
          >
            <span class="material-symbols-outlined text-[1rem]">account_balance</span>
            Rút tiền
          </RouterLink>
          <RouterLink
            to="/deposit"
            class="flex items-center justify-center gap-1.5 rounded-full bg-white py-2.5 text-[0.82rem] font-black text-primary shadow-md active:scale-95 transition-transform"
          >
            <span class="material-symbols-outlined text-[1rem]">add_circle</span>
            Nạp tiền
          </RouterLink>
        </div>
      </div>
    </div>

    <!-- ===== GAME SECTION TITLE ===== -->
    <div class="flex items-center justify-between px-4 pt-4 pb-1">
      <div class="flex items-center gap-2">
        <span class="material-symbols-outlined text-[1.1rem] text-primary">confirmation_number</span>
        <h2 class="text-[0.92rem] font-black text-on-surface">Xổ Số</h2>
      </div>
      <RouterLink to="/play" class="flex items-center gap-0.5 text-[0.75rem] font-bold text-primary">
        Xem tất cả <span class="material-symbols-outlined text-[1rem]">chevron_right</span>
      </RouterLink>
    </div>

    <!-- ===== GAME CARDS (horizontal scroll) ===== -->
    <div class="flex gap-3 overflow-x-auto px-3 pb-1 no-scrollbar">
      <RouterLink
        v-for="game in gameCards"
        :key="game.name"
        :to="game.route"
        class="flex-shrink-0 w-[160px] overflow-hidden rounded-[20px] bg-white shadow-[0_8px_18px_rgba(255,109,102,0.08)] border border-slate-100 transition-transform active:scale-[0.97]"
      >
        <!-- Color banner top -->
        <div
          class="flex h-[90px] items-center justify-center bg-gradient-to-br"
          :class="game.gradient"
        >
          <span class="material-symbols-outlined text-[3.2rem] text-white/90">{{ game.icon }}</span>
        </div>
        <div class="px-3 py-3">
          <strong class="block text-[0.88rem] font-black text-on-surface">{{ game.name }}</strong>
          <p class="mt-0.5 text-[0.7rem] text-slate-500">{{ game.desc }}</p>
          <div class="mt-2.5 flex items-center gap-1 text-[0.68rem] font-bold text-primary">
            Vào chơi <span class="material-symbols-outlined text-[0.9rem]">arrow_forward</span>
          </div>
        </div>
      </RouterLink>
    </div>

    <!-- ===== FEATURED ROOMS (vertical cards) ===== -->
    <div class="px-3 pt-2 pb-1">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-[0.88rem] font-black text-on-surface">Phòng nổi bật</h3>
        <RouterLink to="/play" class="text-[0.75rem] font-bold text-primary flex items-center gap-0.5">
          Xem phòng <span class="material-symbols-outlined text-[1rem]">chevron_right</span>
        </RouterLink>
      </div>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
        <RouterLink
          v-for="game in featuredRooms"
          :key="game.code"
          :to="`/play/${game.code}`"
          class="flex items-center gap-3 rounded-[18px] bg-white p-3.5 shadow-sm border border-slate-100 active:scale-[0.99] transition-transform"
        >
          <div
            class="grid h-11 w-11 flex-shrink-0 place-items-center rounded-[14px]"
            :style="{ backgroundColor: `${game.accent}18`, color: game.accent }"
          >
            <span class="material-symbols-outlined text-[1.4rem]">{{ game.symbol }}</span>
          </div>
          <div class="flex-1 min-w-0">
            <strong class="block text-[0.88rem] font-black text-on-surface">{{ game.title }}</strong>
            <p class="mt-0.5 truncate text-[0.7rem] text-slate-500">{{ game.subtitle }}</p>
          </div>
          <span
            class="flex-shrink-0 rounded-full px-2.5 py-1 text-[0.62rem] font-black"
            :class="game.status === 'OPEN' ? 'bg-emerald-50 text-emerald-600' : 'bg-slate-100 text-slate-500'"
          >{{ game.status === 'OPEN' ? 'Đang mở' : 'Sắp mở' }}</span>
        </RouterLink>
      </div>
    </div>

    <!-- ===== LEADERBOARD ===== -->
    <div class="px-3 pb-1 pt-2">
      <Leaderboard />
    </div>

    <!-- ===== WIN INFO ===== -->
    <div class="mx-3 mb-2 overflow-hidden rounded-[20px] bg-white shadow-[0_8px_18px_rgba(255,109,102,0.06)] border border-slate-100">
      <div class="flex items-center gap-2 border-b border-slate-100 px-4 py-3.5">
        <span class="text-[1.1rem]">🎉</span>
        <span class="text-[0.9rem] font-black text-on-surface">Thông tin trúng thưởng</span>
      </div>
      <div class="divide-y divide-slate-50">
        <div
          v-for="(item, i) in winInfo"
          :key="i"
          class="flex items-center gap-3 px-4 py-2.5 hover:bg-slate-50 transition-colors"
        >
          <img
            :src="`https://i.pravatar.cc/40?img=${item.avatar}`"
            :alt="item.user"
            class="h-8 w-8 rounded-full object-cover flex-shrink-0 border border-slate-100"
          />
          <div class="flex-1 min-w-0">
            <strong class="block text-[0.8rem] font-semibold text-on-surface">{{ item.user }}</strong>
            <span class="text-[0.68rem] text-slate-400">{{ item.game }}</span>
          </div>
          <span class="flex-shrink-0 text-[0.82rem] font-black text-[#10b981]">{{ item.amount }}</span>
        </div>
      </div>
    </div>

    <!-- ===== DOMAIN ACCESS LINKS ===== -->
    <div class="mx-3 mb-4 rounded-[16px] bg-gradient-to-br from-slate-800 to-slate-900 p-4 text-white">
      <p class="text-[0.72rem] text-white/60 mb-2 uppercase tracking-wide font-bold">Thông tin truy cập</p>
      <p class="text-[0.82rem] text-white/90 leading-6">
        Nếu không truy cập được, hãy thử các domain dự phòng hoặc liên hệ CSKH để được hỗ trợ.
      </p>
      <RouterLink
        to="/cskh"
        class="mt-3 inline-flex items-center gap-1.5 rounded-full bg-primary px-4 py-2 text-[0.78rem] font-black text-white active:scale-95 transition-transform"
      >
        <span class="material-symbols-outlined text-[0.9rem]">headphones</span>
        Liên hệ CSKH
      </RouterLink>
    </div>

  </div>
</template>
