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
const usdtQrDataUri = ref('')
const usdtPaymentUri = ref('')
const usdtQrLoading = ref(false)
let usdtQrAbortController: AbortController | null = null

const intent = computed(() => {
  const current = deposit.currentIntent
  if (!current || current.method !== method.value) return null
  return current
})
const status = computed(() => {
  if (!intent.value) return null
  return deposit.currentStatus
})
const isUsdtIntent = computed(() => intent.value?.method === 'usdt')
const activeNetworkLabel = computed(() => {
  const provider = intent.value?.receiving_account?.provider_code?.trim()
  if (!provider) return 'USDTTRC20'
  return provider.toUpperCase()
})
const bankOptions = computed(() => deposit.bankOptions)
const selectedBank = computed(
  () => bankOptions.value.find((bank) => bank.provider_code === selectedBankCode.value) ?? null,
)

const statusLabel = computed(() => {
  const value = status.value?.transaction?.status
  if (value === undefined || value === null) return 'Chưa cập nhật'
  const numValue = Number(value)
  if (numValue === 1) return 'Đang chờ'
  if (numValue === 2 || numValue === 3) return 'Hoàn tất'
  if (numValue === 4) return 'Thất bại'
  return `Mã trạng thái: ${value}`
})

const isIntentActive = computed(() => {
  if (!intent.value) return false
  const statusValue = Number(status.value?.transaction?.status ?? intent.value.transaction?.status)
  return statusValue !== 2 && statusValue !== 3 && statusValue !== 4
})

const presetAmounts = computed(() => {
  if (method.value === 'vietqr') {
    return [100000, 200000, 300000, 500000, 1500000, 15000000]
  }
  return [20, 50, 100, 200, 500, 1000]
})

const isAmountValid = computed(() => {
  const numericAmount = Number(amount.value) || 0
  if (method.value === 'vietqr') return numericAmount >= 2000
  if (method.value === 'usdt') return numericAmount >= 20
  return false
})

const validationMessage = computed(() => {
  if (!amount.value) return ''
  if (method.value === 'vietqr' && Number(amount.value) < 2000) {
    return 'Nạp tối thiểu 2.000 VND'
  }
  if (method.value === 'usdt' && Number(amount.value) < 20) {
    return 'Nạp tối thiểu 20 USDT'
  }
  return ''
})

const amountInputMode = computed(() => (method.value === 'vietqr' ? 'numeric' : 'decimal'))

function sanitizeAmountInput(raw: string, allowDecimal: boolean) {
  const normalized = String(raw ?? '').replaceAll(',', '.').replace(/[^\d.]/g, '')
  if (!allowDecimal) {
    return normalized.replaceAll('.', '')
  }

  const [rawIntegerPart = '', ...fractionParts] = normalized.split('.')
  const integerPart = rawIntegerPart ?? ''
  const fraction = fractionParts.join('')
  if (!fractionParts.length) return integerPart
  return `${integerPart}.${fraction}`
}

function handleAmountInput(event: Event) {
  const target = event.target as HTMLInputElement | null
  amount.value = sanitizeAmountInput(target?.value ?? '', method.value === 'usdt')
}

function redirectBack() {
  router.back()
}

const depositCountdown = computed(() => {
  const expiresAt = intent.value?.expires_at
  if (!expiresAt) return '10:00'

  const remaining = Math.max(0, Date.parse(expiresAt) - now.value)
  const totalSeconds = Math.floor(remaining / 1000)
  
  const h = Math.floor(totalSeconds / 3600)
  const m = Math.floor((totalSeconds % 3600) / 60)
  const s = totalSeconds % 60

  if (h > 0) {
    return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  }
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
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

const cryptoQrImageUrl = computed(() => {
  if (!intent.value || !isUsdtIntent.value) return ''
  return usdtQrDataUri.value
})

const transferContent = computed(() => {
  if (isUsdtIntent.value && usdtPaymentUri.value) return usdtPaymentUri.value
  if (!intent.value) return ''
  const clientRef = intent.value.client_ref?.trim()
  if (clientRef) {
    return clientRef.startsWith('DEP-') ? clientRef.slice(4) : clientRef
  }
  const qrContent = (intent.value.qr_content || '').trim()
  return qrContent.startsWith('DEP-') ? qrContent.slice(4) : qrContent
})

async function loadUsdtQrCode() {
  if (!intent.value || !isUsdtIntent.value) {
    usdtQrDataUri.value = ''
    usdtPaymentUri.value = ''
    return
  }

  const address = (intent.value.receiving_account?.account_number || '').trim()
  const rawAmount = Number(intent.value.amount || 0)
  if (!address || !Number.isFinite(rawAmount) || rawAmount <= 0) {
    usdtQrDataUri.value = ''
    usdtPaymentUri.value = ''
    return
  }

  usdtQrAbortController?.abort()
  const controller = new AbortController()
  usdtQrAbortController = controller
  usdtQrLoading.value = true

  try {
    const params = new URLSearchParams({
      address,
      value: String(rawAmount),
      size: '400',
    })
    const response = await fetch(`https://api.cryptapi.io/trc20/usdt/qrcode/?${params.toString()}`, {
      method: 'GET',
      signal: controller.signal,
    })

    if (!response.ok) {
      throw new Error(`fetch_failed_${response.status}`)
    }

    const payload = (await response.json()) as {
      status?: string
      qr_code?: string
      payment_uri?: string
    }

    if (payload.status !== 'success' || !payload.qr_code) {
      throw new Error('invalid_qr_payload')
    }

    usdtQrDataUri.value = `data:image/png;base64,${payload.qr_code}`
    usdtPaymentUri.value = (payload.payment_uri || '').trim()
  } catch {
    if (!controller.signal.aborted) {
      usdtQrDataUri.value = ''
      usdtPaymentUri.value = ''
    }
  } finally {
    if (usdtQrAbortController === controller) {
      usdtQrAbortController = null
    }
    usdtQrLoading.value = false
  }
}

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
  () => [isUsdtIntent.value, intent.value?.receiving_account?.account_number, intent.value?.amount] as const,
  async ([isUsdt]) => {
    if (!isUsdt) {
      usdtQrAbortController?.abort()
      usdtQrAbortController = null
      usdtQrDataUri.value = ''
      usdtPaymentUri.value = ''
      usdtQrLoading.value = false
      return
    }
    await loadUsdtQrCode()
  },
  { immediate: true },
)

watch(
  () => status.value?.transaction?.status,
  async (nextStatus, previousStatus) => {
    if (nextStatus === previousStatus) return
    const numStatus = Number(nextStatus)
    if (numStatus === 2 || numStatus === 3) {
      try {
        await wallet.fetchSummary()
        // Auto reset after success to close the view after 3 seconds
        setTimeout(() => {
          deposit.reset()
        }, 3000)
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
  usdtQrAbortController?.abort()
  usdtQrAbortController = null
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

async function handleCancel() {
  if (!window.confirm('Bạn có chắc chắn muốn hủy yêu cầu nạp tiền này không?')) return
  try {
    await deposit.cancelDeposit()
  } catch {
    // error is handled in store
  }
}

async function logout() {
  deposit.reset()
  auth.logout()
  await router.replace('/auth')
}
</script>

<template>
  <div class="space-y-5 pb-10">
    <!-- HEADER SECTION -->
    <section class="overflow-hidden rounded-[28px] bg-gradient-to-br from-[#ff7b5d] via-primary to-[#f44956] p-4 text-white shadow-[0_18px_40px_rgba(244,73,86,0.24)] md:p-5">
      <header class="grid min-h-12 grid-cols-[36px_1fr_60px] items-center md:min-h-[52px]">
        <button class="grid h-9 w-9 place-items-center rounded-full bg-white/15 text-white transition-transform active:scale-95" type="button" @click="router.back()">
          <span class="material-symbols-outlined">arrow_back</span>
        </button>
        <h1 class="m-0 text-center text-[1.1rem] font-black tracking-tight md:text-[1.2rem]">Nạp tiền</h1>
        <button class="justify-self-end text-right text-sm font-extrabold text-white/95" type="button" @click="redirectBack()">Thoát</button>
      </header>

      <div v-if="!intent" class="mt-4">
        <div class="rounded-[24px] bg-white/12 p-4 backdrop-blur-sm md:p-5">
          <p class="m-0 text-[0.72rem] font-black uppercase tracking-[0.28em] text-white/80">Nạp nhanh</p>
          <h2 class="mt-2 text-[1.35rem] font-black leading-tight md:text-[1.65rem]">
            Chọn ngân hàng, tạo QR và nạp tiền tự động
          </h2>
          <p class="mt-2 max-w-[40rem] text-sm leading-6 text-white/90">
            Hệ thống hỗ trợ VietQR đối soát tự động 24/7. Vui lòng nhập đúng nội dung chuyển khoản.
          </p>
        </div>
      </div>
    </section>

    <!-- TRƯỜNG HỢP CÓ LỆNH ĐANG MỞ: ĐƯA LÊN ĐẦU TIÊN (UX TỐT HƠN) -->
    <section v-if="intent" class="rounded-[24px] bg-white p-4 shadow-[0_12px_30px_rgba(255,109,102,0.12)] md:p-5 border-2 border-primary/20">
      <div class="flex items-start justify-between gap-3">
        <div>
          <h2 class="m-0 text-base font-black text-primary">Lệnh nạp đang chờ xử lý</h2>
          <p class="m-0 mt-1 text-xs text-on-surface-variant italic">
            {{ isUsdtIntent ? 'Vui lòng quét mã QR hoặc chuyển USDT đúng địa chỉ/memo bên dưới:' : 'Vui lòng quét mã QR hoặc chuyển khoản theo thông tin:' }}
          </p>
        </div>
        <div class="flex items-center gap-2">
          <button class="rounded-full bg-slate-100 px-3 py-1.5 text-sm font-black text-on-surface-variant flex items-center gap-1 transition-transform active:scale-95" type="button" :disabled="deposit.loading" @click="handleCancel">
            <span class="material-symbols-outlined text-[1rem]">close</span>
            Hủy lệnh
          </button>
          <button class="rounded-full bg-primary/10 px-3 py-1.5 text-sm font-black text-primary flex items-center gap-1 transition-transform active:scale-95" type="button" :disabled="deposit.loading" @click="refreshStatus">
            <span class="material-symbols-outlined text-[1rem]">sync</span>
            Cập nhật
          </button>
        </div>
      </div>

      <div class="mt-4 grid gap-4 lg:grid-cols-[1.15fr_0.85fr]">
        <div class="space-y-3 rounded-[20px] bg-surface-container-low p-4">
          <div class="flex items-center gap-3">
            <div class="flex h-12 w-12 items-center justify-center rounded-[16px] bg-white shadow-sm border border-slate-100">
              <img
                v-if="!isUsdtIntent && selectedBank?.logo"
                :src="selectedBank.logo"
                :alt="selectedBank?.short_name"
                class="h-full w-full object-contain p-1.5"
              />
              <span v-else class="text-sm font-black text-primary">
                {{ isUsdtIntent ? 'USDT' : (selectedBank?.short_name || 'QR').slice(0, 2) }}
              </span>
            </div>
            <div class="min-w-0">
              <p class="m-0 truncate text-sm font-black text-on-surface">
                {{ isUsdtIntent ? activeNetworkLabel : (selectedBank?.short_name || intent.receiving_account?.provider_code || 'Ngân hàng') }}
              </p>
              <p class="m-0 truncate text-[0.72rem] text-on-surface-variant lowercase">
                {{ isUsdtIntent ? 'CryptAPI - USDT Deposit' : (intent.receiving_account?.account_name || 'Tài khoản nhận') }}
              </p>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-2 text-sm">
            <div class="rounded-[16px] bg-white p-3 border border-slate-50">
              <p class="m-0 text-[0.72rem] text-on-surface-variant">{{ isUsdtIntent ? 'Đơn vị nhận' : 'Tên người nhận' }}</p>
              <p class="m-0 mt-1 font-black text-on-surface text-[0.8rem] uppercase">{{ intent.receiving_account?.account_name || '---' }}</p>
            </div>
            <div class="rounded-[16px] bg-white p-3 border border-slate-50">
              <p class="m-0 text-[0.72rem] text-on-surface-variant">{{ isUsdtIntent ? 'Địa chỉ ví' : 'Số tài khoản' }}</p>
              <p class="m-0 mt-1 break-all font-black text-primary text-[0.9rem]">{{ intent.receiving_account?.account_number || '---' }}</p>
            </div>
            <div v-if="!isUsdtIntent" class="rounded-[16px] bg-white p-3 border border-slate-50">
              <p class="m-0 text-[0.72rem] text-on-surface-variant">Nội dung (Quan trọng)</p>
              <p class="m-0 mt-1 break-all font-black text-[#e64545] text-[1.1rem]">{{ transferContent || 'Không có' }}</p>
            </div>
            <div class="rounded-[16px] bg-white p-3 border border-slate-50">
              <p class="m-0 text-[0.72rem] text-on-surface-variant">Số tiền nạp</p>
              <p class="m-0 mt-1 font-black text-on-surface text-[0.9rem]">
                {{ isUsdtIntent ? `${intent.amount} USDT` : formatViMoney(intent.amount, 0) }}
              </p>
            </div>
            <div class="rounded-[16px] bg-white p-3 border border-slate-50">
              <p class="m-0 text-[0.72rem] text-on-surface-variant">Hết hạn sau</p>
              <p class="m-0 mt-1 font-black text-on-surface-variant">{{ depositCountdown }}</p>
            </div>
            <div class="rounded-[16px] bg-white p-3 border border-slate-50">
              <p class="m-0 text-[0.72rem] text-on-surface-variant">Trạng thái</p>
              <p class="m-0 mt-1 font-black" :class="(Number(status?.transaction?.status) === 2 || Number(status?.transaction?.status) === 3) ? 'text-emerald-600' : 'text-primary'">{{ statusLabel }}</p>
            </div>
          </div>
        </div>

        <div class="rounded-[20px] bg-gradient-to-br from-[#fff2f1] to-white p-4">
          <div v-if="(isUsdtIntent ? cryptoQrImageUrl : qrImageUrl)" class="overflow-hidden rounded-[18px] border-4 border-white bg-white p-3 shadow-xl mb-3">
            <img :src="isUsdtIntent ? cryptoQrImageUrl : qrImageUrl" :alt="isUsdtIntent ? 'USDT QR' : (selectedBank?.short_name || 'VietQR')" class="block w-full rounded-[14px] object-contain" />
          </div>
          <p v-else-if="isUsdtIntent && usdtQrLoading" class="m-0 mb-3 rounded-[14px] bg-white px-3 py-4 text-center text-xs font-bold text-on-surface-variant">
            Đang tạo mã QR USDT...
          </p>
          <p class="m-0 text-center text-[0.7rem] font-bold text-on-surface-variant px-2 leading-relaxed italic">
            {{ isUsdtIntent ? 'Mở ví crypto, quét QR để chuyển USDTTRC20 nhanh và chính xác địa chỉ/memo.' : 'Mở App ngân hàng, chọn Quét mã QR để tự động điền thông tin và nội dung.' }}
          </p>
        </div>
      </div>
    </section>

    <!-- PHƯƠNG THỨC NẠP -->
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
    </section>

    <!-- CHỌN NGÂN HÀNG -->
    <section v-if="method === 'vietqr'" class="rounded-[24px] bg-white p-4 shadow-[0_8px_24px_rgba(255,109,102,0.08)] md:p-5">
      <h2 class="m-0 text-base font-black text-on-surface">1. Chọn Ngân hàng nhận</h2>

      <div v-if="deposit.banksLoading" class="mt-4 grid gap-3 grid-cols-2 sm:grid-cols-3 xl:grid-cols-4">
        <div v-for="skeleton in 4" :key="skeleton" class="h-[102px] animate-pulse rounded-[22px] bg-surface-container" />
      </div>

      <div v-else-if="bankOptions.length" class="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-3 xl:grid-cols-4">
        <button
          v-for="bank in bankOptions"
          :key="bank.provider_code"
          type="button"
          class="group rounded-[22px] border p-3 text-left transition-all"
          :class="selectedBankCode === bank.provider_code ? 'border-primary bg-primary/5 shadow-sm' : 'border-slate-100 bg-slate-50/50'"
          :disabled="isIntentActive"
          @click="selectedBankCode = bank.provider_code"
        >
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 items-center justify-center overflow-hidden rounded-[12px] bg-white shadow-sm border border-slate-50">
              <img
                v-if="bank.logo"
                :src="bank.logo"
                :alt="bank.short_name"
                class="h-full w-full object-contain p-1"
              />
              <span v-else class="text-xs font-black text-primary">{{ bank.short_name.slice(0, 2) }}</span>
            </div>
            <div class="min-w-0">
              <p class="m-0 truncate text-[0.8rem] font-black text-on-surface">{{ bank.short_name }}</p>
            </div>
          </div>
        </button>
      </div>
    </section>

    <!-- NHẬP SỐ TIỀN -->
    <section class="rounded-[24px] bg-white p-4 shadow-[0_8px_24px_rgba(255,109,102,0.08)] md:p-5">
      <h2 class="m-0 text-base font-black text-on-surface">2. Nhập số tiền nạp</h2>
      <form class="mt-4 space-y-4" @submit.prevent="submitDeposit">
        <div class="grid min-h-[58px] items-center overflow-hidden rounded-[18px] bg-surface-container-low shadow-inner">
          <input
            v-model="amount"
            type="text"
            class="min-w-0 border-0 bg-transparent px-4 py-4 text-lg font-black outline-none disabled:cursor-not-allowed disabled:opacity-60"
            :inputmode="amountInputMode"
            autocomplete="off"
            placeholder="Số tiền (VD: 50000)"
            :disabled="isIntentActive"
            @input="handleAmountInput"
          />
        </div>

        <p v-if="validationMessage" class="m-0 px-1 text-xs font-bold text-[#e64545]">{{ validationMessage }}</p>

        <div class="grid grid-cols-3 gap-2 sm:grid-cols-6">
          <button
            v-for="amt in presetAmounts"
            :key="amt"
            type="button"
            class="min-h-12 rounded-[14px] border bg-slate-50 px-2 font-black text-on-surface transition-transform active:scale-95"
            :class="Number(amount) === amt ? 'border-primary bg-primary/10 text-primary' : 'border-slate-100'"
            :disabled="isIntentActive"
            @click="amount = String(amt)"
          >
            {{ method === 'vietqr' ? `${formatViMoney(amt, 0)}` : `${amt} USDT` }}
          </button>
        </div>

        <button
          class="min-h-14 w-full rounded-[22px] bg-gradient-to-br from-primary to-primary-container font-black text-white shadow-lg disabled:opacity-60 transition-all active:scale-[0.98]"
          type="submit"
          :disabled="isIntentActive || deposit.loading || !isAmountValid || (method === 'vietqr' && !selectedBankCode)"
        >
          {{ isIntentActive ? 'Đang có lệnh mở - vui lòng chờ' : (deposit.loading ? 'Đang tạo giao dịch...' : (method === 'usdt' ? 'Tạo địa chỉ nạp USDT' : 'Tạo mã QR Nạp tiền')) }}
        </button>
      </form>
    </section>
  </div>
</template>
