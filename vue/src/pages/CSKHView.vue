<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const channels = [
  { icon: 'chat', label: 'LiveChat', desc: 'Trò chuyện trực tiếp với tư vấn viên', color: '#3b82f6' },
  { icon: 'send', label: 'Telegram', desc: '@ff789_official', color: '#2AABEE' },
  { icon: 'message', label: 'Zalo', desc: 'Hỗ trợ qua Zalo 24/7', color: '#0068FF' },
  { icon: 'phone', label: 'Hotline 24/7', desc: '1900 xxxx (miễn phí)', color: '#10b981' },
]

const faqs = [
  'Làm thế nào để nạp tiền?',
  'Thời gian rút tiền bao lâu?',
  'Quên mật khẩu phải làm gì?',
  'Liên hệ hỗ trợ như thế nào?',
  'Các hình thức nạp tiền hỗ trợ?',
]
</script>

<template>
  <div class="pb-10">
    <!-- Header -->
    <div class="flex items-center justify-between gap-3 bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] px-4 py-3 text-white">
      <button
        class="grid h-10 w-10 place-items-center rounded-full bg-white/15 text-white"
        @click="router.back()"
      >
        <span class="material-symbols-outlined">arrow_back</span>
      </button>
      <h1 class="flex-1 text-center text-[1rem] font-black">Hỗ Trợ Khách Hàng</h1>
      <div class="h-10 w-10" />
    </div>

    <!-- Hero banner -->
    <div class="relative h-[150px] overflow-hidden">
      <img
        src="https://images.unsplash.com/photo-1556742049-0cfed4f6a45d?w=800&q=80"
        alt="Customer support"
        class="h-full w-full object-cover"
      />
      <div class="absolute inset-0 flex flex-col items-center justify-center gap-1 bg-gradient-to-br from-black/60 to-primary/40">
        <span class="material-symbols-outlined text-[2.5rem] text-white">headphones</span>
        <h2 class="text-[1.1rem] font-black text-white">Hỗ trợ 24/7</h2>
        <p class="text-[0.72rem] text-white/80 text-center px-6">Đội ngũ CSKH luôn sẵn sàng hỗ trợ bạn mọi lúc, mọi nơi</p>
      </div>
    </div>

    <!-- Support channels -->
    <div class="px-4 pt-5">
      <h3 class="mb-3 text-[0.88rem] font-black text-on-surface">Kênh hỗ trợ</h3>
      <div class="flex flex-col gap-2.5">
        <button
          v-for="ch in channels"
          :key="ch.label"
          class="flex items-center gap-4 rounded-[16px] bg-white px-4 py-3.5 shadow-sm border border-slate-100 text-left transition-all active:scale-[0.99] hover:border-primary/30"
        >
          <div
            class="grid h-11 w-11 flex-shrink-0 place-items-center rounded-[14px] text-xl"
            :style="{ backgroundColor: `${ch.color}15`, border: `1.5px solid ${ch.color}30`, color: ch.color }"
          >
            <span class="material-symbols-outlined">{{ ch.icon }}</span>
          </div>
          <div class="flex-1 min-w-0">
            <strong class="block text-[0.88rem] font-black text-on-surface">{{ ch.label }}</strong>
            <p class="text-[0.72rem] text-slate-400">{{ ch.desc }}</p>
          </div>
          <span class="material-symbols-outlined text-[1rem] text-slate-300">chevron_right</span>
        </button>
      </div>
    </div>

    <!-- FAQ section -->
    <div class="px-4 pt-5">
      <h3 class="mb-3 text-[0.88rem] font-black text-on-surface">Câu hỏi thường gặp</h3>
      <div class="flex flex-col gap-2">
        <button
          v-for="(q, i) in faqs"
          :key="i"
          class="flex items-center gap-3 rounded-[14px] bg-white px-4 py-3.5 text-left shadow-sm border border-slate-100 transition-all active:scale-[0.99] hover:border-primary/30"
        >
          <span class="material-symbols-outlined flex-shrink-0 text-[1.1rem] text-primary">help</span>
          <span class="flex-1 text-[0.82rem] font-medium text-on-surface">{{ q }}</span>
          <span class="material-symbols-outlined flex-shrink-0 text-[1rem] text-slate-300">expand_more</span>
        </button>
      </div>
    </div>
  </div>
</template>
