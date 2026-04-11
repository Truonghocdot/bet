<script setup lang="ts">
import { ref } from 'vue'

const activeTab = ref<'income' | 'bigwin'>('income')

const incomeData = [
  { name: 'Mem***AXG', avatar: 21, amount: '19,600đ' },
  { name: 'Mem***DMT', avatar: 22, amount: '1,170đ' },
  { name: 'Mem***ZHS', avatar: 23, amount: '54,880đ' },
  { name: 'Mem***QRT', avatar: 24, amount: '21,000đ' },
  { name: 'Mem***UHE', avatar: 25, amount: '27,440đ' },
  { name: 'Mem***FVT', avatar: 26, amount: '9,760đ' },
  { name: 'Mem***ZKA', avatar: 27, amount: '20,000đ' },
  { name: 'Mem***OCI', avatar: 28, amount: '11,020đ' },
]

const bigwinData = [
  { name: 'Mem***9FN', avatar: 11, amount: '594,798,260đ' },
  { name: 'Mem***MF2', avatar: 12, amount: '467,016,470đ' },
  { name: 'Mem***HP2', avatar: 13, amount: '444,986,640đ' },
  { name: 'Mem***MLN', avatar: 14, amount: '423,732,400đ' },
  { name: 'Mem***UGD', avatar: 15, amount: '405,569,080đ' },
  { name: 'Mem***JAN', avatar: 16, amount: '405,328,000đ' },
  { name: 'Mem***CBG', avatar: 17, amount: '392,000,800đ' },
  { name: 'Mem***DHM', avatar: 18, amount: '329,503,440đ' },
]

const rankIcons = ['🥇', '🥈', '🥉']

const activeList = () => activeTab.value === 'income' ? incomeData : bigwinData
</script>

<template>
  <div class="overflow-hidden rounded-[20px] bg-white shadow-[0_8px_18px_rgba(255,109,102,0.06)] border border-slate-100">
    <!-- Header -->
    <div class="flex items-center gap-2.5 border-b border-slate-100 px-4 py-3.5">
      <span class="text-[1.1rem]">🏆</span>
      <span class="text-[0.9rem] font-black text-on-surface">BXH thu nhập hôm nay</span>
    </div>

    <!-- Tabs -->
    <div class="flex gap-2 px-4 pt-3 pb-1">
      <button
        class="rounded-full border-[1.5px] px-4 py-1.5 text-[0.75rem] font-semibold transition-all"
        :class="activeTab === 'income' ? 'border-primary bg-[#fff5f5] text-primary' : 'border-slate-200 text-slate-500'"
        @click="activeTab = 'income'"
      >Thu nhập</button>
      <button
        class="rounded-full border-[1.5px] px-4 py-1.5 text-[0.75rem] font-semibold transition-all"
        :class="activeTab === 'bigwin' ? 'border-primary bg-[#fff5f5] text-primary' : 'border-slate-200 text-slate-500'"
        @click="activeTab = 'bigwin'"
      >Big Win</button>
    </div>

    <!-- List -->
    <div class="pb-2">
      <div
        v-for="(item, i) in activeList()"
        :key="i"
        class="flex items-center gap-3 rounded-[12px] px-4 py-2 mx-2 transition-colors hover:bg-slate-50"
      >
        <div class="w-7 flex-shrink-0 text-center">
          <span v-if="i < 3" class="text-[1rem]">{{ rankIcons[i] }}</span>
          <span v-else class="text-[0.75rem] font-bold text-slate-400">{{ i + 1 }}</span>
        </div>
        <img
          :src="`https://i.pravatar.cc/40?img=${item.avatar}`"
          :alt="item.name"
          class="h-8 w-8 rounded-full object-cover flex-shrink-0 border border-slate-100"
        />
        <span class="flex-1 text-[0.8rem] font-semibold text-on-surface">{{ item.name }}</span>
        <span class="text-[0.8rem] font-black text-[#10b981]">+{{ item.amount }}</span>
      </div>
    </div>
  </div>
</template>
