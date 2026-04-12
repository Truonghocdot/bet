<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { useDepositStore } from '@/stores/deposit'
import { useAuthStore } from '@/stores/auth'
import { useWalletStore } from '@/stores/wallet'
import { formatViMoney } from '@/shared/lib/money'

const router = useRouter()
const auth = useAuthStore()
const deposit = useDepositStore()
const wallet = useWalletStore()

const method = ref<'vietqr' | 'usdt'>('vietqr')
const amountByMethod = ref<Record<'vietqr' | 'usdt', string>>({
  vietqr: '',
  usdt: '',
})
const amount = computed({
  get: () => amountByMethod.value[method.value],
  set: (value: string) => {
    amountByMethod.value[method.value] = value
  },
})
const selectedBankCode = ref('')
const now = ref(Date.now())
let countdownTicker: number | undefined

const intent = computed(() => {
  const current = deposit.currentIntent
  if (!current || current.method !== method.value) return null
  return current
})
const status = computed(() => {
  if (!intent.value) return null
  return deposit.currentStatus
})
const bankOptions = computed(() => deposit.bankOptions)
const selectedBank = computed(
  () => bankOptions.value.find((bank) => bank.provider_code === selectedBankCode.value) ?? null,
)

const statusLabel = computed(() => {
  const value = status.value?.transaction?.status
  if (value === undefined || value === null) return 'Chưa cập nhật'
  if (value === 1) return 'Đang chờ'
  if (value === 2) return 'Hoàn tất'
  if (value === 3 || value === 4) return 'Thất bại'
  return `Mã trạng thái: ${value}`
})

const isIntentActive = computed(() => {
  if (!intent.value) return false
  const statusValue = status.value?.transaction?.status ?? intent.value.transaction?.status
  return statusValue !== 2 && statusValue !== 3 && statusValue !== 4
})

const presetAmounts = computed(() => {
  if (method.value === 'vietqr') {
    return [50000, 100000, 200000, 500000, 1000000, 5000000]
  }
  return [5, 10, 20, 50, 100, 500]
})

const isAmountValid = computed(() => {
  const numericAmount = Number(amount.value) || 0
  if (method.value === 'vietqr') return numericAmount >= 50000
  if (method.value === 'usdt') return numericAmount >= 5
  return false
})

const validationMessage = computed(() => {
  if (!amount.value) return ''
  if (method.value === 'vietqr' && Number(amount.value) < 50000) {
    return 'Nạp tối thiểu 50.000 VND'
  }
  if (method.value === 'usdt' && Number(amount.value) < 5) {
    return 'Nạp tối thiểu 5 USDT'
  }
  return ''
})

const depositCountdown = computed(() => {
  const expiresAt = intent.value?.expires_at
  if (!expiresAt) return '10:00'

  const remaining = Math.max(0, Date.parse(expiresAt) - now.value)
  const totalSeconds = Math.floor(remaining / 1000)
  const minutes = String(Math.floor(totalSeconds / 60)).padStart(2, '0')
  const seconds = String(totalSeconds % 60).padStart(2, '0')
  return `${minutes}:${seconds}`
})

const qrImageUrl = computed(() => {
  if (method.value !== 'vietqr' || !intent.value) return ''
  if (intent.value.qr_code_url) return intent.value.qr_code_url

  const account = intent.value.receiving_account
  const accountNumber = account?.account_number?.trim()
  if (!accountNumber) return ''

  const bankCode = (selectedBank.value?.provider_code || account?.provider_code || '').trim().toLowerCase()
  if (!bankCode) return ''

  const params = new URLSearchParams()
  if (intent.value.amount) params.set('amount', String(Number.parseFloat(intent.value.amount) || 0))
  if (intent.value.client_ref) params.set('addInfo', intent.value.client_ref)
  if (account?.account_name?.trim()) params.set('accountName', account.account_name.trim())

  const query = params.toString()
  return `https://img.vietqr.io/image/${encodeURIComponent(bankCode)}-${encodeURIComponent(accountNumber)}-compact.jpg${query ? `?${query}` : ''}`
})

const transferContent = computed(() => {
  if (!intent.value) return ''
  const clientRef = intent.value.client_ref?.trim()
  if (clientRef) {
    return clientRef.startsWith('DEP-') ? clientRef.slice(4) : clientRef
  }
  const qrContent = (intent.value.qr_content || '').trim()
  return qrContent.startsWith('DEP-') ? qrContent.slice(4) : qrContent
})

watch(
  () => bankOptions.value,
  (banks) => {
    if (method.value !== 'vietqr' || selectedBankCode.value) return
    selectedBankCode.value = banks.find((bank) => bank.is_default)?.provider_code ?? banks[0]?.provider_code ?? ''
  },
  { immediate: true },
)

watch(
  () => method.value,
  (nextMethod) => {
    deposit.disconnectStatusStream()

    const restored = deposit.restorePending(nextMethod)
    if (restored?.intent) {
      deposit.currentIntent = restored.intent
      deposit.currentStatus = null
      selectedBankCode.value = nextMethod === 'vietqr'
        ? (restored.intent.receiving_account?.provider_code ?? bankOptions.value.find((bank) => bank.is_default)?.provider_code ?? bankOptions.value[0]?.provider_code ?? '')
        : ''
      return
    }

    deposit.currentIntent = null
    deposit.currentStatus = null
    selectedBankCode.value =
      nextMethod === 'vietqr'
        ? bankOptions.value.find((bank) => bank.is_default)?.provider_code ?? bankOptions.value[0]?.provider_code ?? ''
        : ''
  },
  { immediate: true },
)

watch(
  () => intent.value?.client_ref,
  (clientRef) => {
    if (!clientRef) {
      deposit.disconnectStatusStream()
      return
    }

    void deposit.getStatus(clientRef)
    deposit.connectStatusStream(clientRef)
  },
  { immediate: true },
)

watch(
  () => status.value?.transaction?.status,
  async (nextStatus, previousStatus) => {
    if (nextStatus === previousStatus) return
    if (nextStatus === 2) {
      try {
        await wallet.fetchSummary()
      } catch {
        // wallet store keeps its own error state
      }
    }
  },
)

onMounted(async () => {
  countdownTicker = window.setInterval(() => {
    now.value = Date.now()
  }, 1000)

  try {
    await deposit.loadVietQrBanks()
  } catch {
    // store keeps its own error state
  }
})

onBeforeUnmount(() => {
  if (countdownTicker) {
    window.clearInterval(countdownTicker)
  }
  deposit.disconnectStatusStream()
})

async function submitDeposit() {
  if (!isAmountValid.value) return
  if (method.value === 'vietqr') {
    if (!selectedBankCode.value) return
    await deposit.initVietQR({
      amount: amount.value.trim(),
      provider_code: selectedBankCode.value,
    })
  } else {
    await deposit.initUSDT({ amount: amount.value.trim() })
  }
}

async function refreshStatus() {
  if (!intent.value?.client_ref) return
  await deposit.getStatus(intent.value.client_ref)
}

async function logout() {
  deposit.reset()
  auth.logout()
  await router.replace('/auth')
}
</script>

<template>
  <div class="space-y-5">
    <section class="overflow-hidden rounded-[28px] bg-gradient-to-br from-[#ff7b5d] via-primary to-[#f44956] p-4 text-white shadow-[0_18px_40px_rgba(244,73,86,0.24)] md:p-5">
      <header class="grid min-h-12 grid-cols-[36px_1fr_60px] items-center md:min-h-[52px]">
        <button class="grid h-9 w-9 place-items-center rounded-full bg-white/15 text-white transition-transform active:scale-95" type="button" @click="router.back()">
          <span class="material-symbols-outlined">arrow_back</span>
        </button>
        <h1 class="m-0 text-center text-[1.1rem] font-black tracking-tight md:text-[1.2rem]">Nạp tiền</h1>
        <button class="justify-self-end text-right text-sm font-extrabold text-white/95" type="button" @click="logout">Thoát</button>
      </header>

      <div class="mt-4">
        <div class="rounded-[24px] bg-white/12 p-4 backdrop-blur-sm md:p-5">
          <p class="m-0 text-[0.72rem] font-black uppercase tracking-[0.28em] text-white/80">Nạp nhanh</p>
          <h2 class="mt-2 text-[1.35rem] font-black leading-tight md:text-[1.65rem]">
            Chọn ngân hàng, tạo QR và giữ nguyên lệnh đang chờ trong 10 phút.
          </h2>
          <p class="mt-2 max-w-[40rem] text-sm leading-6 text-white/90">
            Nếu bạn quay lại hoặc bấm tạo mới khi lệnh chưa hoàn tất, hệ thống sẽ ưu tiên mở lại đúng giao dịch đang chờ thay vì sinh lệnh khác.
          </p>
        </div>
      </div>
    </section>

    <section class="rounded-[24px] bg-white p-4 shadow-[0_8px_24px_rgba(255,109,102,0.08)] md:p-5">
      <div class="grid grid-cols-2 gap-2 rounded-[18px] bg-surface-container p-1.5">
        <button
          class="min-h-11 rounded-[14px] font-extrabold transition-all"
          :class="method === 'vietqr' ? 'bg-white text-primary shadow-[0_4px_12px_rgba(255,109,102,0.1)]' : 'text-on-surface-variant'"
          type="button"
          :disabled="isIntentActive"
          @click="method = 'vietqr'"
        >
          VietQR
        </button>
        <button
          class="min-h-11 rounded-[14px] font-extrabold transition-all"
          :class="method === 'usdt' ? 'bg-white text-primary shadow-[0_4px_12px_rgba(255,109,102,0.1)]' : 'text-on-surface-variant'"
          type="button"
          :disabled="isIntentActive"
          @click="method = 'usdt'"
        >
          USDT
        </button>
      </div>

      <p class="mt-3 m-0 text-sm text-on-surface-variant">
        {{ method === 'vietqr' ? 'Chọn ngân hàng, số tiền và hệ thống sẽ tự ghép với tài khoản phù hợp.' : 'USDT qua NOWPayments, he thong se tao payment va doi soat tu dong.' }}
      </p>
    </section>

    <section v-if="method === 'vietqr'" class="rounded-[24px] bg-white p-4 shadow-[0_8px_24px_rgba(255,109,102,0.08)] md:p-5">
      <h2 class="m-0 text-base font-black text-on-surface">Ngân hàng nhận tiền</h2>

      <div v-if="deposit.banksLoading" class="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <div v-for="skeleton in 4" :key="skeleton" class="h-[102px] animate-pulse rounded-[22px] bg-surface-container" />
      </div>

      <div v-else-if="bankOptions.length" class="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-3 xl:grid-cols-4">
        <button
          v-for="bank in bankOptions"
          :key="bank.provider_code"
          type="button"
          class="group rounded-[22px] border p-3 text-left transition-all"
          :class="selectedBankCode === bank.provider_code ? 'border-primary bg-primary/5 shadow-[0_12px_24px_rgba(255,109,102,0.12)]' : 'border-slate-200 bg-slate-50/70 hover:border-primary/40 hover:bg-white'"
          :disabled="isIntentActive"
          @click="selectedBankCode = bank.provider_code"
        >
          <div class="flex items-center gap-3">
            <div class="flex h-12 w-12 items-center justify-center overflow-hidden rounded-[16px] bg-white shadow-sm">
              <img
                v-if="bank.logo"
                :src="bank.logo"
                :alt="bank.short_name"
                class="h-full w-full object-contain p-1.5"
              />
              <span v-else class="text-sm font-black text-primary">{{ bank.short_name.slice(0, 2) }}</span>
            </div>
            <div class="min-w-0">
              <p class="m-0 truncate text-sm font-black text-on-surface">{{ bank.short_name }}</p>
            </div>
          </div>
        </button>
      </div>

      <div v-else class="mt-4 rounded-[22px] border border-dashed border-slate-200 bg-slate-50 p-4 text-sm text-on-surface-variant">
        Chưa có ngân hàng nào khả dụng. Hệ thống sẽ tự cập nhật khi quản trị viên bật bank nhận tiền.
      </div>
    </section>

    <section class="rounded-[24px] bg-white p-4 shadow-[0_8px_24px_rgba(255,109,102,0.08)] md:p-5">
      <form class="space-y-3" @submit.prevent="submitDeposit">
        <div class="grid min-h-[58px] items-center overflow-hidden rounded-[18px] bg-surface-container-low shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
          <input
            v-model="amount"
            type="number"
            class="min-w-0 border-0 bg-transparent px-4 py-4 text-lg font-black outline-none disabled:cursor-not-allowed disabled:opacity-60"
            inputmode="decimal"
            placeholder="Nhập số tiền nạp"
            :disabled="isIntentActive"
          />
        </div>

        <p v-if="validationMessage" class="m-0 px-1 text-xs font-bold text-[#e64545]">{{ validationMessage }}</p>

        <div class="grid grid-cols-3 gap-2 sm:grid-cols-6">
          <button
            v-for="amt in presetAmounts"
            :key="amt"
            type="button"
            class="min-h-12 rounded-[14px] border bg-slate-50 px-2 font-black text-on-surface transition-transform active:scale-95"
            :class="Number(amount) === amt ? 'border-primary bg-primary/10 text-primary' : 'border-transparent'"
            :disabled="isIntentActive"
            @click="amount = String(amt)"
          >
            {{ method === 'vietqr' ? `${formatViMoney(amt, 0)}` : `${amt} USDT` }}
          </button>
        </div>

        <button
          class="min-h-14 rounded-[18px] bg-gradient-to-br from-primary to-primary-container font-black text-white shadow-[0_14px_28px_rgba(255,109,102,0.25)] disabled:opacity-60"
          type="submit"
          :disabled="isIntentActive || deposit.loading || !isAmountValid || (method === 'vietqr' && !selectedBankCode)"
        >
          {{ isIntentActive ? 'Đang có lệnh mở - vui lòng chờ hoàn tất' : (deposit.loading ? 'Đang tạo giao dịch...' : method === 'vietqr' ? 'Tạo QR VietQR' : 'Tạo yêu cầu USDT') }}
        </button>
      </form>
    </section>

    <section v-if="intent" class="rounded-[24px] bg-white p-4 shadow-[0_8px_24px_rgba(255,109,102,0.08)] md:p-5">
      <div class="flex items-start justify-between gap-3">
        <div>
          <h2 class="m-0 text-base font-black text-on-surface">Lệnh đang mở</h2>
          <p class="m-0 mt-1 text-sm text-on-surface-variant">Giao dịch này sẽ được giữ trên màn hình cho đến khi hoàn tất hoặc hết 10 phút.</p>
        </div>
        <button class="rounded-full bg-primary/10 px-3 py-1.5 text-sm font-black text-primary" type="button" @click="refreshStatus">
          Cập nhật
        </button>
      </div>

      <div class="mt-4 grid gap-4 lg:grid-cols-[1.15fr_0.85fr]">
        <div class="space-y-3 rounded-[20px] bg-surface-container-low p-4">
          <div class="flex items-center gap-3">
            <div class="flex h-12 w-12 items-center justify-center rounded-[16px] bg-white shadow-sm">
              <img
                v-if="selectedBank?.logo"
                :src="selectedBank.logo"
                :alt="selectedBank.short_name"
                class="h-full w-full object-contain p-1.5"
              />
              <span v-else class="text-sm font-black text-primary">{{ (selectedBank?.short_name || 'QR').slice(0, 2) }}</span>
            </div>
            <div class="min-w-0">
              <p class="m-0 truncate text-sm font-black text-on-surface">{{ selectedBank?.short_name || intent.receiving_account?.provider_code || 'Ngân hàng' }}</p>
              <p class="m-0 truncate text-[0.72rem] text-on-surface-variant">{{ selectedBank?.name || intent.receiving_account?.account_name || 'Tài khoản nhận tiền' }}</p>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-2 text-sm">
            <div class="rounded-[16px] bg-white p-3">
              <p class="m-0 text-[0.72rem] text-on-surface-variant">Tên người nhận</p>
              <p class="m-0 mt-1 font-black text-on-surface uppercase">{{ intent.receiving_account?.account_name || '---' }}</p>
            </div>
            <div class="rounded-[16px] bg-white p-3">
              <p class="m-0 text-[0.72rem] text-on-surface-variant">Số tài khoản</p>
              <p class="m-0 mt-1 font-black text-on-surface">{{ intent.receiving_account?.account_number || '---' }}</p>
            </div>
            <div class="rounded-[16px] bg-white p-3">
              <p class="m-0 text-[0.72rem] text-on-surface-variant">Mã giao dịch</p>
              <p class="m-0 mt-1 break-all font-black text-on-surface">{{ intent.client_ref }}</p>
            </div>
            <div class="rounded-[16px] bg-white p-3">
              <p class="m-0 text-[0.72rem] text-on-surface-variant">Số tiền</p>
              <p class="m-0 mt-1 font-black text-on-surface">{{ formatViMoney(intent.amount, 0) }}</p>
            </div>
            <div class="rounded-[16px] bg-white p-3">
              <p class="m-0 text-[0.72rem] text-on-surface-variant">Phiên hết hạn</p>
              <p class="m-0 mt-1 font-black text-primary">{{ depositCountdown }}</p>
            </div>
            <div class="rounded-[16px] bg-white p-3">
              <p class="m-0 text-[0.72rem] text-on-surface-variant">Trạng thái</p>
              <p class="m-0 mt-1 font-black" :class="status?.transaction?.status === 2 ? 'text-emerald-600' : 'text-primary'">{{ statusLabel }}</p>
            </div>
          </div>

          <p class="m-0 text-sm leading-6 text-on-surface-variant">
            {{ intent.instructions || 'Quét QR hoặc chuyển khoản đúng nội dung hệ thống hiển thị.' }}
          </p>
        </div>

        <div class="rounded-[20px] bg-gradient-to-br from-[#fff2f1] to-white p-4">
          <div v-if="qrImageUrl" class="overflow-hidden rounded-[18px] border border-white bg-white p-3 shadow-[0_10px_24px_rgba(255,109,102,0.08)]">
            <img :src="qrImageUrl" :alt="selectedBank?.short_name || 'VietQR'" class="block w-full rounded-[14px] object-contain" />
          </div>
          <div class="rounded-[18px] border border-dashed border-primary/25 bg-white p-4 text-sm space-y-4">
            <div>
              <p class="m-0 text-[0.72rem] font-black uppercase tracking-[0.12em] text-primary">Tên người nhận</p>
              <p class="m-0 mt-1 font-black text-on-surface uppercase">{{ intent.receiving_account?.account_name || '---' }}</p>
            </div>
            <div>
              <p class="m-0 text-[0.72rem] font-black uppercase tracking-[0.12em] text-primary">Số tài khoản</p>
              <p class="m-0 mt-1 font-black text-on-surface text-lg">{{ intent.receiving_account?.account_number || '---' }}</p>
            </div>
            <div>
              <p class="m-0 text-[0.72rem] font-black uppercase tracking-[0.12em] text-primary">Nội dung chuyển khoản</p>
              <p class="m-0 mt-1 font-black text-on-surface">{{ transferContent || '---' }}</p>
            </div>
          </div>
          <div class="mt-3 rounded-[18px] bg-white p-4 text-sm text-on-surface-variant">
            <p class="m-0 font-bold text-on-surface">Cách xử lý</p>
            <p class="m-0 mt-1 leading-6">
              Quét QR để tự điền thông tin. Nếu chuyển thủ công, chỉ cần nhập đúng nội dung trên.
            </p>
          </div>
        </div>
      </div>
    </section>

    <section v-if="status" class="rounded-[24px] bg-white p-4 shadow-[0_8px_24px_rgba(255,109,102,0.08)]">
      <div class="rounded-[18px] bg-primary/10 p-4 text-sm font-bold text-primary">
        <p class="m-0">Trạng thái hiện tại: {{ statusLabel }}</p>
        <p class="m-0 mt-1 text-[0.76rem] font-medium text-on-surface-variant">
          {{ status.message }}
        </p>
      </div>
    </section>
  </div>
</template>
