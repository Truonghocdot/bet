<script setup lang="ts">
import { ref, computed } from 'vue'
import { useWalletStore } from '@/stores/wallet'
import { useNotificationsStore } from '@/stores/notifications'
import { formatViMoney } from '@/shared/lib/money'

const walletStore = useWalletStore()
const notificationStore = useNotificationsStore()

const fromUnit = ref(2) // Default from USDT (2)
const toUnit = ref(1)   // Default to VND (1)
const amount = ref('')
const loading = ref(false)

const RATE = computed(() => {
  const r = parseFloat(walletStore.summary?.exchange_rate || '25000')
  return isNaN(r) || r <= 0 ? 25000 : r
})

const fromWallet = computed(() => walletStore.wallets.find(w => w.unit === fromUnit.value))
const toWallet = computed(() => walletStore.wallets.find(w => w.unit === toUnit.value))

const estimatedToAmount = computed(() => {
  const val = parseFloat(amount.value)
  if (isNaN(val) || val <= 0) return '0'
  
  if (fromUnit.value === 2) {
    return (val * RATE.value).toFixed(0)
  } else {
    return (val / RATE.value).toFixed(8)
  }
})

function swapCurrencies() {
  const temp = fromUnit.value
  fromUnit.value = toUnit.value
  toUnit.value = temp
  amount.value = ''
}

async function handleExchange() {
  const val = parseFloat(amount.value)
  if (isNaN(val) || val <= 0) {
    notificationStore.addLocalNotification('Lỗi', 'Vui lòng nhập số tiền hợp lệ', 'error')
    return
  }

  if (val > parseFloat(fromWallet.value?.balance || '0')) {
    notificationStore.addLocalNotification('Lỗi', 'Số dư không đủ', 'error')
    return
  }

  loading.value = true
  try {
    await walletStore.exchangeWallets({
      from_unit: fromUnit.value,
      to_unit: toUnit.value,
      amount: String(amount.value)
    })
    notificationStore.addLocalNotification('Thành công', 'Đã chuyển đổi tiền tệ thành công')
    amount.value = ''
  } catch (e: any) {
    notificationStore.addLocalNotification('Lỗi', e.message || 'Không thể chuyển đổi', 'error')
  } finally {
    loading.value = false
  }
}

function setMax() {
  if (fromWallet.value) {
    amount.value = fromWallet.value.balance
  }
}
</script>

<template>
  <div class="px-4 py-6 max-w-md mx-auto">
    <div class="mb-8">
      <h1 class="text-2xl font-black text-slate-800">Chuyển đổi ví</h1>
      <p class="text-sm font-bold text-slate-400 mt-1">Chuyển đổi linh hoạt giữa USDT và VND</p>
    </div>

    <!-- EXCHANGE CARD -->
    <div class="rounded-[30px] bg-white p-6 shadow-xl shadow-slate-200/50 relative overflow-hidden">
      <!-- Decor -->
      <div class="absolute -top-10 -right-10 w-32 h-32 bg-primary/5 rounded-full blur-2xl"></div>

      <!-- FROM SECTION -->
      <div class="space-y-2">
        <div class="flex justify-between items-center px-1">
          <span class="text-[10px] font-black uppercase text-slate-400 tracking-widest">Từ ví</span>
          <span class="text-[10px] font-bold text-slate-400">Số dư: <span class="text-slate-800">{{ formatViMoney(fromWallet?.balance || '0', fromUnit === 1 ? 0 : 2) }}</span></span>
        </div>
        <div class="flex items-center gap-3 p-4 rounded-2xl bg-slate-50 border border-slate-100">
          <div class="h-10 w-10 rounded-xl bg-white shadow-sm flex items-center justify-center font-black text-xs" :class="fromUnit === 1 ? 'text-blue-500' : 'text-green-500'">
            {{ fromUnit === 1 ? 'VND' : 'USDT' }}
          </div>
          <div class="flex-1">
            <input v-model="amount" type="number" step="any" placeholder="0.00" class="w-full bg-transparent border-none p-0 focus:ring-0 font-black text-xl text-slate-800" />
          </div>
          <button @click="setMax" class="text-[10px] font-black text-primary uppercase px-2 py-1 bg-primary/10 rounded-lg">Tất cả</button>
        </div>
      </div>

      <!-- SWAP BUTTON -->
      <div class="relative h-4 flex items-center justify-center my-4">
        <div class="absolute inset-x-0 h-[1px] bg-slate-100"></div>
        <button @click="swapCurrencies" class="relative z-10 h-10 w-10 rounded-full bg-slate-800 text-white shadow-lg flex items-center justify-center transition-transform active:rotate-180 duration-500">
          <span class="material-symbols-outlined text-xl">swap_vert</span>
        </button>
      </div>

      <!-- TO SECTION -->
      <div class="space-y-2">
        <div class="flex justify-between items-center px-1">
          <span class="text-[10px] font-black uppercase text-slate-400 tracking-widest">Đến ví</span>
          <span class="text-[10px] font-bold text-slate-400">Ước tính nhận</span>
        </div>
        <div class="flex items-center gap-3 p-4 rounded-2xl bg-slate-50 border border-slate-100">
          <div class="h-10 w-10 rounded-xl bg-white shadow-sm flex items-center justify-center font-black text-xs" :class="toUnit === 1 ? 'text-blue-500' : 'text-green-500'">
            {{ toUnit === 1 ? 'VND' : 'USDT' }}
          </div>
          <div class="flex-1">
            <div class="font-black text-xl text-slate-400">{{ formatViMoney(estimatedToAmount, toUnit === 1 ? 0 : 2) }}</div>
          </div>
        </div>
      </div>

      <!-- RATE INFO -->
      <div class="mt-6 p-4 rounded-2xl bg-blue-50 border border-blue-100/50 flex items-center gap-3">
        <span class="material-symbols-outlined text-blue-500">info</span>
        <div class="text-[11px] font-bold text-blue-700">
          Tỷ giá hiện tại: <span class="font-black">1 USDT = {{ formatViMoney(RATE, 0) }} VND</span>
        </div>
      </div>

      <!-- SUBMIT -->
      <button 
        @click="handleExchange"
        :disabled="loading || !amount || parseFloat(amount) <= 0"
        class="w-full mt-6 h-14 rounded-2xl bg-gradient-to-br from-slate-800 to-slate-900 text-white font-black shadow-lg shadow-slate-200 disabled:opacity-50 transition-all active:scale-[0.98]"
      >
        <div v-if="loading" class="flex items-center justify-center gap-2">
          <div class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
          <span>Đang xử lý...</span>
        </div>
        <span v-else>Xác nhận chuyển đổi</span>
      </button>
    </div>

    <!-- TIPS -->
    <div class="mt-6 space-y-3">
      <div class="flex gap-3 text-slate-400 p-1">
        <span class="material-symbols-outlined text-sm">security</span>
        <p class="text-[10px] font-bold leading-relaxed">Giao dịch được bảo mật và thực hiện ngay lập tức trong hệ thống nội bộ.</p>
      </div>
      <div class="flex gap-3 text-slate-400 p-1">
        <span class="material-symbols-outlined text-sm">help</span>
        <p class="text-[10px] font-bold leading-relaxed">Tiền VND dùng để chơi game, tiền USDT thường được dùng để quy đổi giá trị tài sản ổn định.</p>
      </div>
    </div>
  </div>
</template>
