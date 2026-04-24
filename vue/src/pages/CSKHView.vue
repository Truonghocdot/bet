<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useWalletStore } from '@/stores/wallet'

const router = useRouter()
const walletStore = useWalletStore()

const telegramLink = computed(() => walletStore.summary?.telegram_cskh_link || 'https://t.me/CSKH_FF789')

function openTelegram() {
  window.open(telegramLink.value, '_blank')
}

const channels = [
  { icon: 'send', label: 'Telegram Hỗ Trợ', desc: 'Gặp trực tiếp kỹ thuật viên', color: '#2AABEE', action: openTelegram },
  { icon: 'forum', label: 'Kênh Khiếu Nại', desc: 'Phản ánh chất lượng dịch vụ', color: '#ef4444', action: openTelegram },
]

const faqs = [
  { q: 'Làm thế nào để nạp tiền vào tài khoản?', a: 'Bạn vào mục Nạp tiền, chọn phương thức VietQR hoặc USDT. Hệ thống sẽ hiển thị thông tin thanh toán, bạn thực hiện chuyển khoản và tiền sẽ được cộng tự động.' },
  { q: 'Thời gian rút tiền mất bao lâu?', a: 'Lệnh rút tiền thường được xử lý trong vòng 3-15 phút. Trong một số trường hợp cao điểm có thể mất đến 30 phút.' },
  { q: 'Tôi bị quên mật khẩu phải làm sao?', a: 'Bạn có thể sử dụng tính năng "Quên mật khẩu" tại màn hình đăng nhập hoặc liên hệ trực tiếp với CSKH qua Telegram để được hỗ trợ cấp lại.' },
  { q: 'Làm thế nào để đăng ký tài khoản?', a: 'Để đăng ký tài khoản, bạn cần chuẩn bị số điện thoại đang sử dụng, một tài khoản ngân hàng chính chủ và mã giới thiệu từ đại lý. Trong trường hợp chưa có mã giới thiệu, vui lòng liên hệ bộ phận Chăm sóc Khách hàng để được hỗ trợ nhanh chóng. Sau đó, chỉ cần điền đầy đủ thông tin theo hướng dẫn trên hệ thống để hoàn tất đăng ký.' },
  { q: 'Những phương thức nạp tiền nào được hỗ trợ?', a: 'Hiện tại, hệ thống hỗ trợ nạp tiền qua chuyển khoản ngân hàng và ví điện tử crypto (USDT), giúp người dùng giao dịch nhanh chóng, an toàn và tiện lợi' },
  { q: 'Có mất phí khi nạp/rút không?', a: 'Mọi giao dịch nạp và rút tiền đều hoàn toàn miễn phí, vì vậy bạn có thể yên tâm tận hưởng những trải nghiệm giải trí hấp dẫn trên sàn.' },
  { q: 'Vòng cược tính như thế nào?', a: 'Không có yêu cầu vòng cược rườm rà. Bạn chỉ cần chơi đủ số tiền tương ứng với khoản khuyến mãi đã nhận là có thể rút toàn bộ tiền về tài khoản ngân hàng một cách dễ dàng.' },
  { q: 'Thông tin cá nhân có được bảo mật không?', a: 'Toàn bộ dữ liệu cá nhân của khách hàng được bảo vệ bằng hạ tầng công nghệ tiên tiến, áp dụng các lớp bảo mật và mã hóa hiện đại, đồng thời lưu trữ trên hệ thống điện toán đám mây đạt tiêu chuẩn an toàn cao.đảm bảo tối đa quyền riêng tư và an toàn thông tin cho người dùng.' },
  { q: 'Liên hệ CSKH bằng cách nào?', a: 'Để liên hệ Chăm sóc Khách hàng, bạn chỉ cần truy cập vào mục ‘Chăm sóc Khách hàng’ trên hệ thống. Đội ngũ hỗ trợ 24/7 luôn sẵn sàng giải đáp mọi thắc mắc, giúp bạn có trải nghiệm tốt nhất.' },

]

onMounted(() => {
  if (!walletStore.summary) {
    walletStore.fetchSummary()
  }
})
</script>

<template>
  <div class="min-h-screen bg-slate-50 pb-10">
    <!-- Header -->
    <div class="fixed top-0 left-0 right-0 z-50 flex items-center justify-between gap-3 bg-white px-4 py-3 text-on-surface shadow-sm">
      <button
        class="grid h-10 w-10 place-items-center rounded-full bg-slate-100 text-on-surface transition-transform active:scale-90"
        @click="router.back()"
      >
        <span class="material-symbols-outlined">arrow_back</span>
      </button>
      <h1 class="flex-1 text-center text-[1rem] font-black uppercase tracking-wider">Hỗ Trợ Khách Hàng</h1>
      <div class="h-10 w-10" />
    </div>

    <!-- Hero container -->
    <div class="pt-[64px]">
      <div class="bg-gradient-to-br from-primary to-[#ff8a00] p-6 text-white overflow-hidden relative">
        <div class="relative z-10">
          <h2 class="text-[1.25rem] font-black leading-tight italic">Chúng tôi có thể giúp gì cho bạn?</h2>
          <p class="mt-2 text-[0.75rem] text-white/80 max-w-[200px]">Đội ngũ hỗ trợ FF789 luôn sẵn sàng giải đáp mọi thắc mắc của bạn 24/7.</p>
        </div>
        <span class="material-symbols-outlined absolute -right-4 -bottom-4 text-[120px] text-white/10 rotate-12">support_agent</span>
      </div>
    </div>

    <!-- Support channels grid -->
    <div class="px-4 mt-6">
      <div class="grid grid-cols-2 gap-3">
        <button
          v-for="ch in channels"
          :key="ch.label"
          class="flex flex-col gap-3 rounded-[24px] bg-white p-4 shadow-sm border border-slate-100 transition-all active:scale-[0.96]"
          @click="ch.action"
        >
          <div
            class="grid h-12 w-12 place-items-center rounded-[16px] text-2xl"
            :style="{ backgroundColor: `${ch.color}15`, border: `1.5px solid ${ch.color}25`, color: ch.color }"
          >
            <span class="material-symbols-outlined text-[1.5rem]">{{ ch.icon }}</span>
          </div>
          <div class="min-w-0">
            <strong class="block text-[0.85rem] font-black text-on-surface line-clamp-1 italic">{{ ch.label }}</strong>
            <p class="mt-0.5 text-[0.68rem] text-slate-400 line-clamp-2 leading-tight font-medium">{{ ch.desc }}</p>
          </div>
        </button>
      </div>
    </div>

    <!-- FAQ section -->
    <div class="px-4 mt-8">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-[0.95rem] font-black text-on-surface italic">Câu hỏi thường gặp</h3>
        <span class="text-[0.7rem] font-bold text-primary uppercase">Xem tất cả</span>
      </div>
      
      <div class="space-y-3">
        <details
          v-for="(faq, i) in faqs"
          :key="i"
          class="group rounded-[18px] bg-white border border-slate-100"
        >
          <summary class="flex cursor-pointer list-none items-center justify-between p-4 outline-none">
            <div class="flex items-center gap-3">
              <span class="material-symbols-outlined text-[1.1rem] text-primary">contact_support</span>
              <span class="text-[0.8rem] font-bold text-on-surface">{{ faq.q }}</span>
            </div>
            <span class="material-symbols-outlined text-[1.1rem] text-slate-400 transition-transform group-open:rotate-180">expand_more</span>
          </summary>
          <div class="px-4 pb-4 pt-0 text-[0.75rem] leading-relaxed text-slate-500 border-t border-slate-50">
            <div class="p-3 bg-slate-50 rounded-[12px]">
              {{ faq.a }}
            </div>
          </div>
        </details>
      </div>
    </div>

    <!-- Footer call -->
    <div class="px-6 mt-10 text-center">
      <p class="text-[0.7rem] font-medium text-slate-400">Bạn vẫn chưa tìm thấy câu trả lời?</p>
      <button 
        class="mt-3 w-full rounded-full bg-slate-900 py-3.5 text-[0.85rem] font-black text-white shadow-lg active:scale-95 transition-transform"
        @click="openTelegram"
      >
        Liên hệ hỗ trợ trực tiếp
      </button>
    </div>
  </div>
</template>

<style scoped>
/* Hủy nút mặc định của summary */
summary::-webkit-details-marker {
  display: none;
}
</style>
