<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const router = useRouter()
const route = useRoute()

const countdown = ref(5)
let timer: number | undefined

function goHome() {
  if (timer) window.clearInterval(timer)
  void router.replace('/home')
}

function goBack() {
  if (typeof window !== 'undefined' && window.history.length > 1) {
    void router.back()
    return
  }
  goHome()
}

onMounted(() => {
  timer = window.setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) {
      goHome()
    }
  }, 1000)
})

onBeforeUnmount(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="flex min-h-[70dvh] items-center justify-center px-4 py-10">
    <div class="w-full max-w-lg overflow-hidden rounded-[28px] bg-white shadow-[0_16px_40px_rgba(255,109,102,0.12)] border border-slate-100">
      <div class="bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] px-6 py-5 text-white">
        <p class="text-[0.72rem] font-black uppercase tracking-[0.14em] text-white/72">404</p>
        <h1 class="mt-1 text-[1.5rem] font-black">Không tìm thấy trang</h1>
        <p class="mt-1 text-[0.85rem] text-white/88">
          Đường dẫn này không tồn tại hoặc đã được chuyển đi.
        </p>
      </div>

      <div class="px-6 py-6">
        <div class="grid place-items-center">
          <div class="grid h-20 w-20 place-items-center rounded-[24px] bg-[#fff5f5] text-primary">
            <span class="material-symbols-outlined text-[2.5rem]">search_off</span>
          </div>
        </div>

        <div class="mt-5 rounded-[18px] bg-[#fff9f9] px-4 py-3 text-center">
          <p class="text-[0.78rem] text-slate-500">URL hiện tại</p>
          <p class="mt-1 break-all text-[0.82rem] font-semibold text-on-surface">{{ route.fullPath }}</p>
        </div>

        <div class="mt-4 grid gap-2 text-[0.82rem] text-slate-500">
          <p>Trang sẽ tự chuyển về trang chủ sau <strong class="text-primary">{{ countdown }}</strong> giây.</p>
          <p>Nếu bạn vừa đi lạc từ game, có thể quay lại màn trước bằng nút bên dưới.</p>
        </div>

        <div class="mt-6 grid grid-cols-2 gap-3">
          <button
            type="button"
            class="min-h-[48px] rounded-[16px] border-2 border-slate-200 bg-white px-4 text-[0.9rem] font-black text-slate-600"
            @click="goBack"
          >
            Quay lại
          </button>
          <button
            type="button"
            class="min-h-[48px] rounded-[16px] bg-gradient-to-r from-[#ff8a00] to-[#e52e2e] px-4 text-[0.9rem] font-black text-white shadow-[0_8px_16px_rgba(229,46,46,0.2)]"
            @click="goHome"
          >
            Về trang chủ
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
