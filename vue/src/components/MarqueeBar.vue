<script setup lang="ts">
const messages = [
  'Quý khách thân mến vui lòng thay đổi cổng nạp tiền nếu không thể tạo lệnh nạp.',
  'Khi nạp tiền bằng cổng CHUYỂN KHOẢN sẽ được nhận thêm ưu đãi đặc biệt!',
  'FF789 – Đăng ký hôm nay nhận ngay thưởng chào mừng 100%.',
]
</script>

<template>
  <div class="flex items-center gap-2 overflow-hidden bg-[rgba(232,64,74,0.08)] px-3 py-2.5 border-y border-[rgba(232,64,74,0.15)]">
    <span class="material-symbols-outlined flex-shrink-0 text-[1rem] text-primary">campaign</span>
    <div class="marquee-track flex-1 overflow-hidden">
      <div class="marquee-content flex whitespace-nowrap">
        <span
          v-for="(msg, i) in [...messages, ...messages]"
          :key="i"
          class="text-[0.72rem] text-slate-600 font-medium"
        >{{ msg }}&nbsp;&nbsp;&nbsp;•&nbsp;&nbsp;&nbsp;</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
@keyframes marquee-scroll {
  0% { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}

.marquee-content {
  animation: marquee-scroll 22s linear infinite;
}

.marquee-track:hover .marquee-content {
  animation-play-state: paused;
}
</style>
