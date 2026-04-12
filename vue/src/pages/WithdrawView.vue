<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { formatViMoney } from '@/shared/lib/money'
import { useAuthStore } from '@/stores/auth'
import { useWalletStore } from '@/stores/wallet'
import { useWithdrawStore } from '@/stores/withdraw'
import type { WithdrawalAccount } from '@/stores/withdraw'

const router = useRouter()
const auth = useAuthStore()
const wallet = useWalletStore()
const withdraw = useWithdrawStore()

const method = ref<'vnd' | 'usdt'>('vnd')
const amount = ref('')
const isAddingForm = ref(false)

// Form for adding method
const addProvider = ref('')
const addHolder = ref('')
const addNumber = ref('')

const currentWallets = computed(() => {
  return {
    vnd: wallet.wallets.find((w) => w.unit === 1),
    usdt: wallet.wallets.find((w) => w.unit === 2),
  }
})

const currentAccount = computed<WithdrawalAccount | undefined>(() => {
  if (method.value === 'vnd') return withdraw.vndAccounts[0]
  return withdraw.usdtAccounts[0]
})

const needsSetup = computed(() => {
  return currentAccount.value === undefined
})

const validationMessage = computed(() => {
  if (!amount.value) return ''
  const numeric = Number(amount.value)
  if (method.value === 'vnd') {
    if (numeric < 50000) return 'Tối thiểu rút 50,000 VND'
    if (numeric > Number(currentWallets.value.vnd?.balance || 0)) return 'Số dư khả dụng không đủ'
  } else {
    if (numeric < 5) return 'Tối thiểu rút 5 USDT'
    if (numeric > Number(currentWallets.value.usdt?.balance || 0)) return 'Số dư USDT khả dụng không đủ'
  }
  return ''
})

const canSubmit = computed(() => {
  return amount.value && !validationMessage.value && currentAccount.value !== undefined
})

onMounted(async () => {
  if (!auth.isAuthenticated) return router.replace('/auth')
  await Promise.all([wallet.fetchSummary(), withdraw.fetchAccounts()])
})

// Methods limit helpers
const presets = computed(() => {
  if (method.value === 'vnd') return [50000, 100000, 200000, 500000, 1000000]
  return [5, 10, 50, 100, 500]
})

async function submitSaveMethod() {
  if (!addProvider.value || !addHolder.value || !addNumber.value) return
  await withdraw.addAccount({
    unit: method.value === 'vnd' ? 1 : 2,
    provider_code: addProvider.value.toUpperCase().trim(),
    account_name: addHolder.value.toUpperCase().trim(),
    account_number: addNumber.value.toUpperCase().trim(),
  })
  isAddingForm.value = false
  addProvider.value = ''
  addHolder.value = ''
  addNumber.value = ''
}

async function handleWithdraw() {
  if (!canSubmit.value || !currentAccount.value) return
  const success = await withdraw.submitWithdrawal({
    account_withdrawal_info_id: currentAccount.value.id,
    amount: amount.value,
  })
  if (success) {
    amount.value = ''
    await wallet.fetchSummary() // update locked balance visualization
  }
}

function promptAddForm() {
  isAddingForm.value = true
}

async function handleRemoveMethod() {
  if (!currentAccount.value) return
  if (!confirm('Bạn có chắc xoá hồ sơ nhận tiền này không?')) return
  await withdraw.deleteAccount(currentAccount.value.id)
}
</script>

<template>
  <div class="space-y-5 pb-10">
    <header class="grid min-h-12 grid-cols-[36px_1fr_60px] items-center md:min-h-[52px]">
      <button class="grid h-9 w-9 place-items-center text-primary transition-transform active:scale-95" type="button" @click="router.back()">
        <span class="material-symbols-outlined">arrow_back</span>
      </button>
      <h1 class="m-0 text-center text-[1.1rem] font-black text-primary md:text-[1.2rem]">Rút tiền</h1>
      <div />
    </header>

    <section class="rounded-[22px] bg-white p-5 shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
      <div class="grid grid-cols-2 gap-2 rounded-[18px] bg-surface-container p-1.5">
        <button
          class="min-h-11 rounded-[14px] font-extrabold transition-all"
          :class="method === 'vnd' ? 'bg-white text-primary shadow-[0_4px_12px_rgba(255,109,102,0.1)]' : 'text-on-surface-variant'"
          type="button"
          @click="method = 'vnd'; isAddingForm = false;"
        >
          Ngân hàng (VND)
        </button>
        <button
          class="min-h-11 rounded-[14px] font-extrabold transition-all"
          :class="method === 'usdt' ? 'bg-white text-primary shadow-[0_4px_12px_rgba(255,109,102,0.1)]' : 'text-on-surface-variant'"
          type="button"
          @click="method = 'usdt'; isAddingForm = false;"
        >
          Ví USDT Crypto
        </button>
      </div>

      <div class="mt-4 rounded-[18px] bg-gradient-to-br from-[#ff6d66] to-[#ffd4d0] px-4 py-5 text-white">
        <p class="m-0 text-sm opacity-80">Số dư khả dụng</p>
        <p class="m-0 mt-1 text-2xl font-black">
          {{ method === 'vnd' ? formatViMoney(currentWallets.vnd?.balance || 0, 0) : formatViMoney(currentWallets.usdt?.balance || 0, 2) }}
        </p>
      </div>
    </section>

    <!-- TRẠNG THÁI CẦN THÊM VÍ -->
    <transition name="page" mode="out-in">
      <section v-if="needsSetup || isAddingForm" class="rounded-[22px] bg-white p-5 shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
        <h2 class="m-0 text-base font-black text-primary">Liên kết {{ method === 'vnd' ? 'Ngân hàng' : 'Ví Crypto' }}</h2>
        <p class="m-0 mt-1 text-xs font-bold text-[#e64545] opacity-80">Bạn cần khai báo thông tin nơi nhận tiền để tạo lệnh rút.</p>
        
        <form class="mt-4 space-y-3" @submit.prevent="submitSaveMethod">
          <label class="block">
            <span class="text-xs font-bold text-on-surface-variant">{{ method === 'vnd' ? 'Tên Ngân hàng (VD: MBBank, VCB)' : 'Mạng lưới (VD: TRC20, ERC20)' }}</span>
            <input v-model="addProvider" class="mt-1 min-h-12 w-full rounded-[14px] bg-slate-50 px-4 font-semibold text-on-surface outline-none" required />
          </label>
          <label class="block">
            <span class="text-xs font-bold text-on-surface-variant">{{ method === 'vnd' ? 'Chủ tài khoản (Không dấu)' : 'Nhãn ghi nhớ' }}</span>
            <input v-model="addHolder" class="mt-1 min-h-12 w-full rounded-[14px] bg-slate-50 px-4 font-semibold text-on-surface outline-none uppercase" required />
          </label>
          <label class="block">
            <span class="text-xs font-bold text-on-surface-variant">{{ method === 'vnd' ? 'Số tài khoản' : 'Địa chỉ ví' }}</span>
            <input v-model="addNumber" class="mt-1 min-h-12 w-full rounded-[14px] bg-slate-50 px-4 font-semibold text-on-surface outline-none" required />
          </label>

          <div class="pt-2 flex gap-2">
            <button v-if="!needsSetup && isAddingForm" type="button" @click="isAddingForm = false" class="min-h-12 flex-1 rounded-[14px] bg-slate-100 font-bold text-on-surface">Huỷ</button>
            <button class="min-h-12 flex-1 rounded-[14px] bg-primary font-black text-white disabled:opacity-60" type="submit" :disabled="withdraw.loading">
              Lưu cấu hình
            </button>
          </div>
        </form>
      </section>

      <!-- ĐÃ CÓ VÍ, CHO PHÉP NHẬP SỐ TIỀN RÚT -->
      <section v-else class="rounded-[22px] bg-white p-5 shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
        <div class="flex items-start justify-between rounded-[16px] bg-[rgba(65,82,143,0.06)] p-3.5 mb-4">
          <div class="overflow-hidden">
            <p class="m-0 text-[0.7rem] font-bold text-on-surface-variant uppercase">{{ currentAccount?.provider_code }}</p>
            <p class="m-0 mt-0.5 truncate font-black text-on-surface">{{ currentAccount?.account_name }}</p>
            <p class="m-0 font-mono text-[0.85rem] font-bold text-primary">{{ currentAccount?.account_number }}</p>
          </div>
          <button class="grid h-8 w-8 shrink-0 place-items-center rounded-full bg-white text-[#e64545] shadow-sm ml-2" @click="handleRemoveMethod">
            <span class="material-symbols-outlined text-[1.1rem]">delete</span>
          </button>
        </div>

        <form @submit.prevent="handleWithdraw" class="space-y-4">
          <div>
            <label class="grid min-h-[58px] items-center overflow-hidden rounded-[18px] bg-surface-container-low shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
              <input v-model="amount" type="number" class="min-w-0 border-0 bg-transparent px-4 py-4 outline-none font-bold text-lg" inputmode="decimal" placeholder="Nhập số tiền muốn rút" />
            </label>
            <p v-if="validationMessage" class="mt-2 text-xs font-bold text-[#e64545] px-2 m-0">{{ validationMessage }}</p>
          </div>

          <div class="grid grid-cols-5 gap-1.5">
            <button
              v-for="amt in presets"
              :key="amt"
              type="button"
              class="min-h-10 rounded-[12px] bg-slate-50 text-[0.75rem] font-bold text-on-surface transition-transform active:scale-95"
              :class="Number(amount) === amt ? 'bg-primary/10 text-primary border-primary border' : ''"
              @click="amount = String(amt)"
            >
              {{ method === 'vnd' && amt >= 1000 ? (amt / 1000) + 'K' : amt }}
            </button>
          </div>

          <button class="min-h-14 w-full rounded-[18px] bg-gradient-to-br from-primary to-primary-container font-black text-white hover:shadow-lg disabled:opacity-60" type="submit" :disabled="withdraw.loading || !canSubmit">
            {{ withdraw.loading ? 'Đang xử lý...' : 'Yêu cầu rút tiền' }}
          </button>
        </form>
      </section>
    </transition>
  </div>
</template>

<style scoped>
.page-enter-active,
.page-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.page-enter-from,
.page-leave-to {
  opacity: 0;
  transform: translateY(8px);
}
</style>
