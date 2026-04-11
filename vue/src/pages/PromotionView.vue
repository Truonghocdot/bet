<script setup lang="ts">
const promotions = [
  {
    title: 'Thưởng nạp lần đầu 100%',
    desc: 'Nhận thêm 100% giá trị nạp lần đầu tiên. Áp dụng 1 lần duy nhất.',
    badge: 'HOT',
    color: '#e8404a',
    icon: 'redeem',
  },
  {
    title: 'Hoàn trả hàng ngày',
    desc: 'Nhận hoàn trả lên đến 1.5% tổng cược mỗi ngày, không giới hạn.',
    badge: 'MỚI',
    color: '#10b981',
    icon: 'autorenew',
  },
  {
    title: 'Thưởng giới thiệu bạn bè',
    desc: 'Nhận 50,000đ cho mỗi người bạn giới thiệu đăng ký thành công.',
    badge: '',
    color: '#3b82f6',
    icon: 'people',
  },
  {
    title: 'Ưu đãi thành viên VIP',
    desc: 'Quyền lợi đặc biệt dành riêng cho thành viên VIP: cashback, quà tặng & hỗ trợ ưu tiên.',
    badge: 'VIP',
    color: '#f59e0b',
    icon: 'workspace_premium',
  },
]
</script>

<template>
  <div class="space-y-3 pb-8">
    <!-- Page header -->
    <div class="flex items-center gap-2 border-b border-slate-100 bg-white px-4 py-4">
      <span class="material-symbols-outlined text-[1.3rem] text-primary">redeem</span>
      <h1 class="text-[1rem] font-black text-on-surface">Hoạt động & Khuyến mãi</h1>
    </div>

    <!-- Promotion cards -->
    <div class="flex flex-col gap-3 px-3">
      <div
        v-for="(promo, i) in promotions"
        :key="i"
        class="flex overflow-hidden rounded-[18px] bg-white shadow-sm border border-slate-100 transition-all active:scale-[0.99] hover:-translate-y-0.5"
      >
        <!-- Left accent bar -->
        <div class="w-1.5 flex-shrink-0" :style="{ background: promo.color }" />
        <div class="flex flex-1 gap-3 px-4 py-4">
          <!-- Icon -->
          <div
            class="flex-shrink-0 grid h-11 w-11 place-items-center rounded-[14px] mt-0.5"
            :style="{ backgroundColor: `${promo.color}15`, color: promo.color }"
          >
            <span class="material-symbols-outlined text-[1.3rem]">{{ promo.icon }}</span>
          </div>
          <!-- Content -->
          <div class="flex-1 min-w-0">
            <div class="flex items-start justify-between gap-2">
              <strong class="text-[0.9rem] font-black text-on-surface leading-snug">{{ promo.title }}</strong>
              <span
                v-if="promo.badge"
                class="flex-shrink-0 rounded-full px-2 py-0.5 text-[0.6rem] font-black"
                :style="{ backgroundColor: `${promo.color}20`, color: promo.color }"
              >{{ promo.badge }}</span>
            </div>
            <p class="mt-1 text-[0.75rem] leading-5 text-slate-500">{{ promo.desc }}</p>
            <button
              class="mt-3 rounded-full px-4 py-1.5 text-[0.75rem] font-black text-white transition-opacity hover:opacity-85"
              :style="{ background: promo.color }"
            >Nhận ngay</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Events coming soon notice -->
    <div class="mx-3 flex flex-col items-center gap-3 rounded-[18px] bg-white py-8 px-6 text-center shadow-sm border border-slate-100">
      <span class="material-symbols-outlined text-[2.5rem] text-slate-200">event_upcoming</span>
      <p class="text-[0.82rem] text-slate-400">Các sự kiện đặc biệt sẽ được cập nhật sớm</p>
    </div>
  </div>
</template>
