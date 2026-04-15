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
const showWithdrawPolicyModal = ref(false)

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
  try {
    const dismissed = window.localStorage.getItem('ff789:withdraw-policy-dismissed')
    showWithdrawPolicyModal.value = dismissed !== '1'
  } catch {
    showWithdrawPolicyModal.value = true
  }
  await Promise.all([wallet.fetchSummary(), withdraw.fetchAccounts(), withdraw.fetchHistory()])
})

// Methods limit helpers
const presets = computed(() => {
  if (method.value === 'vnd') return [200000, 300000, 500000, 1500000, 5000000 , 15000000]
  return [5, 10, 50, 100, 500]
})

async function submitSaveMethod() {
  if (!addHolder.value || !addNumber.value) return
  await withdraw.addAccount({
    unit: method.value === 'vnd' ? 1 : 2,
    provider_code: addProvider.value.toUpperCase().trim(),
    account_name: addHolder.value.toUpperCase().trim(),
    account_number: addNumber.value.toUpperCase().trim(),
  })
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
    await wallet.fetchSummary()
    router.replace('/home')
  }
}

function closeWithdrawPolicyModal() {
  showWithdrawPolicyModal.value = false
  try {
    window.localStorage.setItem('ff789:withdraw-policy-dismissed', '1')
  } catch {
    // no-op
  }
}
</script>

<template>
  <div class="space-y-5 pb-10">
    <div v-if="showWithdrawPolicyModal" class="fixed inset-0 z-[9999] grid place-items-center bg-black/50 p-4">
      <div class="w-full max-w-[720px] rounded-[18px] bg-white p-4 md:p-5">
        <h2 class="m-0 text-[1rem] font-black text-primary">Thông báo bảo mật liên kết ngân hàng</h2>
        <div class="mt-3 max-h-[60vh] overflow-y-auto rounded-[14px] bg-slate-50 p-3 text-[0.78rem] leading-6 text-slate-700">
          <p class="m-0">
            Để bảo mật thông tin Ngân Hàng cá nhân của Quý Khách Hàng và tránh tất cả trường hợp lộ thông tin, Hệ thống bên sàn hoàn toàn chạy tự động bằng công nghệ mới nhất!
          </p>
          <p class="m-0 mt-2">
            * Lưu ý: Tất cả tài khoản khi liên kết bắt buộc phải thực hiện đúng quy định của ngân hàng, tài khoản liên kết lên hệ thống có thể là tài khoản chính chủ hoặc của người thân sẽ đều được chấp nhận.
          </p>
          <p class="m-0 mt-2">
            Mỗi một tài khoản game chỉ được liên kết một tài khoản ngân hàng. Không nên sử dụng tài khoản không biết rõ nguồn gốc để tránh trường hợp mất tài sản của Quý Khách Hàng.
          </p>
          <p class="m-0 mt-2">
            Mọi hành vi cố tình dùng tài khoản không rõ nguồn gốc để rút tiền mà xảy ra trường hợp thất thoát hay vi phạm quy định thì Hệ Thống sẽ không chịu trách nhiệm.
          </p>
          <p class="m-0 mt-2">
            Khi liên kết thông tin ngân hàng lên hệ thống cần ghi rõ thông tin đầy đủ và trước khi liên kết cần ghi rõ thông tin chi nhánh ngân hàng phía sau tên ngân hàng của thành viên.
          </p>
          <p class="m-0 mt-2">
            Do hệ thống Ngân Hàng yêu cầu, mong Quý Khách vui lòng liên kết thông tin đầy đủ để tránh trường hợp phải xác minh lại thông tin.
          </p>
          <p class="m-0 mt-2">
            Mọi thắc mắc xin vui lòng liên hệ lên bộ phận Chăm Sóc Khách Hàng để được tư vấn hỗ trợ.
          </p>
          <p class="m-0 mt-2 font-black text-primary">XIN TRÂN THÀNH CẢM ƠN !!!</p>
        </div>
        <div class="mt-4 flex justify-end">
          <button
            class="min-h-11 rounded-[12px] bg-primary px-4 text-sm font-black text-white"
            type="button"
            @click="closeWithdrawPolicyModal"
          >
            Tôi đã hiểu
          </button>
        </div>
      </div>
    </div>

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
          @click="method = 'vnd'; amount = ''; addProvider = ''; addHolder = ''; addNumber = '';"
        >
          Ngân hàng (VND)
        </button>
        <button
          class="min-h-11 rounded-[14px] font-extrabold transition-all"
          :class="method === 'usdt' ? 'bg-white text-primary shadow-[0_4px_12px_rgba(255,109,102,0.1)]' : 'text-on-surface-variant'"
          type="button"
          @click="method = 'usdt'; amount = ''; addProvider = ''; addHolder = ''; addNumber = '';"
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
      <section v-if="needsSetup" class="rounded-[22px] bg-white p-5 shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
        <h2 class="m-0 text-base font-black text-primary">Liên kết {{ method === 'vnd' ? 'Ngân hàng' : 'Ví Crypto' }}</h2>
        <p class="m-0 mt-1 text-xs font-bold text-[#e64545] opacity-80">Bạn cần khai báo thông tin nơi nhận tiền để tạo lệnh rút.</p>
        <p class="m-0 mt-1 text-[11px] font-bold text-slate-500">
          Quy định: Mỗi tài khoản game chỉ liên kết duy nhất 1 tài khoản nhận tiền.
        </p>
        
        <form class="mt-4 space-y-3" @submit.prevent="submitSaveMethod">
          <label class="block">
            <span class="text-xs font-bold text-on-surface-variant">{{ method === 'vnd' ? 'Tên Ngân hàng (VD: MBBank, VCB)' : 'Mạng lưới (VD: TRC20, ERC20)' }}</span>
            <input v-model="addProvider" class="mt-1 min-h-12 w-full rounded-[14px] bg-slate-50 px-4 font-semibold text-on-surface outline-none" :required="method === 'usdt'" />
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
            <button class="min-h-12 flex-1 rounded-[14px] bg-primary font-black text-white disabled:opacity-60" type="submit" :disabled="withdraw.loading">
              Lưu cấu hình
            </button>
          </div>
        </form>
      </section>

      <!-- ĐÃ CÓ VÍ, CHO PHÉP NHẬP SỐ TIỀN RÚT -->
      <section v-else class="rounded-[22px] bg-white p-5 shadow-[0_8px_20px_rgba(255,109,102,0.06)] relative overflow-hidden">
        <div class="flex items-start justify-between rounded-[16px] bg-[rgba(65,82,143,0.06)] p-3.5 mb-4">
          <div class="overflow-hidden">
            <p class="m-0 text-[0.7rem] font-bold text-on-surface-variant uppercase">
              {{ method === 'vnd' ? 'Số tài khoản nhận' : 'Địa chỉ ví nhận' }}
            </p>
            <p class="m-0 mt-1 font-mono text-[0.9rem] font-black text-primary break-all">
              {{ currentAccount?.account_number }}
            </p>
          </div>
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

    <!-- LỊCH SỬ RÚT TIỀN -->
    <section class="rounded-[22px] bg-white p-5 shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
      <div class="flex items-center justify-between mb-4">
         <h2 class="m-0 text-base font-black text-primary">Lịch sử rút tiền</h2>
         <button @click="() => withdraw.fetchHistory()" class="text-xs font-bold text-blue-500 uppercase tracking-tighter">Làm mới</button>
      </div>

      <div class="space-y-3">
         <div v-if="withdraw.history.length === 0" class="py-10 text-center">
            <p class="text-xs font-bold text-slate-300">Chưa có giao dịch rút tiền nào</p>
         </div>
         <div v-for="item in withdraw.history" :key="item.id" 
              class="flex items-center justify-between p-3 rounded-2xl bg-slate-50 border border-slate-100/50">
            <div class="flex gap-3 items-center">
               <div class="h-10 w-10 rounded-xl flex items-center justify-center font-black text-xs"
                    :class="item.unit === 1 ? 'bg-blue-100 text-blue-600' : 'bg-green-100 text-green-600'">
                  {{ item.unit === 1 ? 'VND' : 'USDT' }}
               </div>
               <div>
                  <p class="m-0 text-xs font-black text-slate-700">Rút {{ formatViMoney(item.amount, item.unit === 1 ? 0 : 2) }}</p>
                  <p class="m-0 text-[10px] font-bold text-slate-400 mt-0.5">{{ item.created_at.split('T')[0] }} {{ item.created_at.split('T')[1]?.substring(0, 5) || '' }}</p>
               </div>
            </div>
            <div class="text-right">
               <span class="px-2.5 py-1 rounded-full text-[10px] font-black uppercase"
                     :class="[
                        item.status === 1 ? 'bg-amber-100 text-amber-600' : '',
                        item.status === 2 ? 'bg-blue-100 text-blue-600' : '',
                        item.status === 5 ? 'bg-green-100 text-green-600' : '',
                        item.status === 3 || item.status === 4 ? 'bg-rose-100 text-rose-600' : ''
                     ]">
                  {{ item.status === 1 ? 'Đang chờ' : (item.status === 2 ? 'Đã duyệt' : (item.status === 5 ? 'Thành công' : (item.status === 4 ? 'Đã hủy' : 'Từ chối'))) }}
               </span>
               <p v-if="item.reason_rejected" class="m-0 mt-1 text-[9px] font-medium text-rose-400 italic font-mono">{{ item.reason_rejected }}</p>
            </div>
         </div>
      </div>
    </section>
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
