<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import type { ApiError } from '@/shared/api/http'
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
const currentPassword = ref('')
const newPassword = ref('')
const confirmNewPassword = ref('')
const passwordFormError = ref('')
const passwordFormSuccess = ref('')
const changingPassword = ref(false)
const showCurrentPassword = ref(false)
const showNewPassword = ref(false)
const showConfirmPassword = ref(false)

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

function resetPasswordForm() {
  currentPassword.value = ''
  newPassword.value = ''
  confirmNewPassword.value = ''
}

async function submitChangePassword() {
  passwordFormError.value = ''
  passwordFormSuccess.value = ''

  if (!currentPassword.value || !newPassword.value || !confirmNewPassword.value) {
    passwordFormError.value = 'Vui lòng nhập đầy đủ mật khẩu cũ, mật khẩu mới và xác nhận mật khẩu mới.'
    return
  }

  if (newPassword.value.length < 6) {
    passwordFormError.value = 'Mật khẩu mới phải từ 6 ký tự trở lên.'
    return
  }

  if (newPassword.value !== confirmNewPassword.value) {
    passwordFormError.value = 'Mật khẩu mới và nhập lại mật khẩu mới chưa khớp.'
    return
  }

  changingPassword.value = true
  try {
    const response = await auth.changePassword({
      old_password: currentPassword.value,
      new_password: newPassword.value,
    })
    passwordFormSuccess.value = response.message || 'Đổi mật khẩu thành công'
    notifications.addLocalNotification('Thành công', passwordFormSuccess.value, 'success')
    resetPasswordForm()
  } catch (error) {
    const apiError = error as ApiError
    passwordFormError.value = apiError?.message || 'Không thể đổi mật khẩu'
    notifications.addLocalNotification('Lỗi', passwordFormError.value, 'error')
  } finally {
    changingPassword.value = false
  }
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

    <section class="rounded-[22px] bg-white p-[18px] shadow-[0_8px_20px_rgba(255,109,102,0.05)] md:p-5">
      <div class="flex items-start justify-between gap-3">
        <div>
          <h3 class="m-0 text-[1rem] font-black text-on-surface">Đổi mật khẩu</h3>
          <p class="m-0 mt-1 text-[0.78rem] text-on-surface-variant">
            Người chơi có thể cập nhật lại mật khẩu đăng nhập ngay tại đây.
          </p>
        </div>
        <div class="grid h-11 w-11 place-items-center rounded-2xl bg-primary/10 text-primary">
          <span class="material-symbols-outlined">lock_reset</span>
        </div>
      </div>

      <form class="mt-4 grid gap-3" @submit.prevent="submitChangePassword">
        <label class="grid gap-1.5">
          <span class="text-[0.76rem] font-extrabold uppercase tracking-[0.08em] text-on-surface-variant">Mật khẩu cũ</span>
          <div class="grid grid-cols-[1fr_auto] items-center rounded-[16px] border border-slate-200 bg-slate-50">
            <input
              v-model="currentPassword"
              class="min-w-0 bg-transparent px-4 py-3.5 text-[0.95rem] outline-none"
              :type="showCurrentPassword ? 'text' : 'password'"
              autocomplete="current-password"
              placeholder="Nhập mật khẩu cũ"
            />
            <button type="button" class="px-4 text-[0.75rem] font-black text-primary" @click="showCurrentPassword = !showCurrentPassword">
              {{ showCurrentPassword ? 'Ẩn' : 'Hiện' }}
            </button>
          </div>
        </label>

        <label class="grid gap-1.5">
          <span class="text-[0.76rem] font-extrabold uppercase tracking-[0.08em] text-on-surface-variant">Mật khẩu mới</span>
          <div class="grid grid-cols-[1fr_auto] items-center rounded-[16px] border border-slate-200 bg-slate-50">
            <input
              v-model="newPassword"
              class="min-w-0 bg-transparent px-4 py-3.5 text-[0.95rem] outline-none"
              :type="showNewPassword ? 'text' : 'password'"
              autocomplete="new-password"
              placeholder="Nhập mật khẩu mới"
            />
            <button type="button" class="px-4 text-[0.75rem] font-black text-primary" @click="showNewPassword = !showNewPassword">
              {{ showNewPassword ? 'Ẩn' : 'Hiện' }}
            </button>
          </div>
        </label>

        <label class="grid gap-1.5">
          <span class="text-[0.76rem] font-extrabold uppercase tracking-[0.08em] text-on-surface-variant">Nhập lại mật khẩu mới</span>
          <div class="grid grid-cols-[1fr_auto] items-center rounded-[16px] border border-slate-200 bg-slate-50">
            <input
              v-model="confirmNewPassword"
              class="min-w-0 bg-transparent px-4 py-3.5 text-[0.95rem] outline-none"
              :type="showConfirmPassword ? 'text' : 'password'"
              autocomplete="new-password"
              placeholder="Nhập lại mật khẩu mới"
            />
            <button type="button" class="px-4 text-[0.75rem] font-black text-primary" @click="showConfirmPassword = !showConfirmPassword">
              {{ showConfirmPassword ? 'Ẩn' : 'Hiện' }}
            </button>
          </div>
        </label>

        <p v-if="passwordFormError" class="rounded-[14px] bg-[rgba(183,18,17,0.08)] px-4 py-3 text-sm font-semibold text-[#e64545]">
          {{ passwordFormError }}
        </p>
        <p v-else-if="passwordFormSuccess" class="rounded-[14px] bg-[rgba(34,197,94,0.12)] px-4 py-3 text-sm font-semibold text-[#15803d]">
          {{ passwordFormSuccess }}
        </p>

        <button
          type="submit"
          class="grid min-h-14 place-items-center rounded-[18px] bg-gradient-to-br from-primary to-primary-container font-extrabold text-white transition-transform active:scale-95 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="changingPassword"
        >
          {{ changingPassword ? 'Đang cập nhật...' : 'Cập nhật mật khẩu' }}
        </button>
      </form>
    </section>

    <button class="min-h-14 p-2 rounded-[18px] bg-[rgba(183,18,17,0.1)] font-black text-[#e64545] transition-transform active:scale-95" @click="logout">
      Đăng xuất tài khoản
    </button>

    <p class="mt-3 text-center text-[0.66rem] font-bold uppercase tracking-[0.18em] text-[#abadb2]">
      Phiên bản 2.4.0 • FF789 Gaming Ecosystem
    </p>
  </div>
</template>
