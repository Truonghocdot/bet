<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { useDepositStore } from '@/stores/deposit'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()
const deposit = useDepositStore()

const method = ref<'vietqr' | 'usdt'>('vietqr')
const amount = ref('')
const note = ref('')

const intent = computed(() => deposit.currentIntent)
const status = computed(() => deposit.currentStatus)
const statusLabel = computed(() => {
  const value = status.value?.transaction?.status
  if (value === undefined || value === null) return 'Chưa cập nhật'
  if (value === 1) return 'Đang chờ'
  if (value === 2) return 'Hoàn tất'
  if (value === 3 || value === 4) return 'Thất bại'
  return `Mã trạng thái: ${value}`
})

let pollTimer: number | undefined

function stopPolling() {
  if (pollTimer) {
    window.clearInterval(pollTimer)
    pollTimer = undefined
  }
}

function startPolling(clientRef: string) {
  stopPolling()
  void deposit.getStatus(clientRef)
  pollTimer = window.setInterval(() => {
    void deposit.getStatus(clientRef)
  }, 6000)
}

watch(
  () => intent.value?.client_ref,
  (clientRef) => {
    if (!clientRef) {
      stopPolling()
      return
    }

    startPolling(clientRef)
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  stopPolling()
})

async function submitDeposit() {
  const payload = { amount: amount.value.trim(), note: note.value.trim() || undefined }
  if (method.value === 'vietqr') {
    await deposit.initVietQR(payload)
  } else {
    await deposit.initUSDT(payload)
  }
}

async function refreshStatus() {
  if (!intent.value?.client_ref) return
  await deposit.getStatus(intent.value.client_ref)
}

async function logout() {
  auth.logout()
  await router.replace('/auth')
}
</script>

<template>
  <div class="space-y-5">
    <header class="grid min-h-12 grid-cols-[36px_1fr_60px] items-center md:min-h-[52px]">
      <button class="grid h-9 w-9 place-items-center text-primary transition-transform active:scale-95" type="button" @click="router.back()">
        <span class="material-symbols-outlined">arrow_back</span>
      </button>
      <h1 class="m-0 text-center text-[1.1rem] font-black text-primary md:text-[1.2rem]">Nạp tiền</h1>
      <button class="justify-self-end text-right text-sm font-extrabold text-primary" type="button" @click="logout">Thoát</button>
    </header>

    <section class="rounded-[22px] bg-white p-5 shadow-[0_8px_20px_rgba(0,78,219,0.06)]">
      <div class="grid grid-cols-2 gap-2 rounded-[18px] bg-surface-container p-1.5">
        <button
          class="min-h-11 rounded-[14px] font-extrabold transition-all"
          :class="method === 'vietqr' ? 'bg-white text-primary shadow-[0_4px_12px_rgba(0,78,219,0.1)]' : 'text-on-surface-variant'"
          type="button"
          @click="method = 'vietqr'"
        >
          VietQR
        </button>
        <button
          class="min-h-11 rounded-[14px] font-extrabold transition-all"
          :class="method === 'usdt' ? 'bg-white text-primary shadow-[0_4px_12px_rgba(0,78,219,0.1)]' : 'text-on-surface-variant'"
          type="button"
          @click="method = 'usdt'"
        >
          USDT
        </button>
      </div>

      <form class="mt-4 space-y-3" @submit.prevent="submitDeposit">
        <label class="grid min-h-[58px] items-center overflow-hidden rounded-[18px] bg-surface-container-low shadow-[0_8px_20px_rgba(0,78,219,0.06)]">
          <input v-model="amount" class="min-w-0 border-0 bg-transparent px-4 py-4 outline-none" inputmode="decimal" placeholder="Số tiền nạp" />
        </label>

        <label class="grid min-h-[58px] items-center overflow-hidden rounded-[18px] bg-surface-container-low shadow-[0_8px_20px_rgba(0,78,219,0.06)]">
          <input v-model="note" class="min-w-0 border-0 bg-transparent px-4 py-4 outline-none" placeholder="Ghi chú (không bắt buộc)" />
        </label>

        <button class="min-h-14 rounded-[18px] bg-gradient-to-br from-primary to-primary-container font-black text-white disabled:opacity-60" type="submit" :disabled="deposit.loading || !amount.trim()">
          {{ deposit.loading ? 'Đang tạo giao dịch...' : 'Tạo yêu cầu nạp' }}
        </button>
      </form>
    </section>

    <section v-if="intent" class="space-y-3 rounded-[22px] bg-white p-5 shadow-[0_8px_20px_rgba(0,78,219,0.06)]">
      <div class="flex items-center justify-between gap-3">
        <h2 class="m-0 text-base font-black text-on-surface">Thông tin nạp</h2>
        <button class="text-sm font-extrabold text-primary" type="button" @click="refreshStatus">Cập nhật trạng thái</button>
      </div>

      <div class="space-y-2 rounded-[18px] bg-background p-4 text-sm">
        <p class="m-0"><strong>Mã giao dịch:</strong> {{ intent.client_ref }}</p>
        <p class="m-0"><strong>Số tiền:</strong> {{ intent.amount }}</p>
        <p class="m-0"><strong>Hết hạn:</strong> {{ intent.expires_at }}</p>
        <p class="m-0"><strong>Hướng dẫn:</strong> {{ intent.instructions || '---' }}</p>
        <p v-if="intent.receiving_account" class="m-0"><strong>Tài khoản nhận:</strong> {{ intent.receiving_account.account_name || intent.receiving_account.wallet_address || intent.receiving_account.account_number }}</p>
        <p v-if="intent.qr_content" class="m-0 break-words"><strong>Nội dung chuyển:</strong> {{ intent.qr_content }}</p>
      </div>

      <div v-if="status" class="rounded-[18px] bg-primary/10 p-4 text-sm font-bold text-primary">
        <p class="m-0">Trạng thái hiện tại: {{ statusLabel }}</p>
        <p class="m-0 mt-1 text-[0.76rem] font-medium text-on-surface-variant">
          {{ status.message }}
        </p>
      </div>
    </section>

    <section class="rounded-[22px] bg-white p-5 shadow-[0_8px_20px_rgba(0,78,219,0.06)]">
      <p class="m-0 text-sm text-on-surface-variant">
        Tài khoản nhận tiền được hệ thống chọn ngẫu nhiên từ danh sách quản trị viên đã cấu hình trước đó.
      </p>
    </section>
  </div>
</template>
