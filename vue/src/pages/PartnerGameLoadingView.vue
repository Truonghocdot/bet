<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const loadingSeconds = ref(60)
const hasTimedOut = ref(false)

let countdownTicker: number | undefined
let timeoutTicker: number | undefined

const gameName = computed(() => {
  const rawName = String(route.query.name ?? '').trim()
  return rawName || 'Sảnh game đối tác'
})

const returnTarget = computed(() => {
  const from = String(route.query.from ?? '').trim()
  if (from.startsWith('/')) return from
  return '/home'
})

const loadingMessage = computed(() => (
  hasTimedOut.value
    ? 'Sảnh game đối tác đang gặp lỗi. Vui lòng thử lại sau ít phút.'
    : 'Hệ thống đang kết nối tới sảnh game đối tác, vui lòng chờ trong giây lát.'
))

function clearTimers() {
  if (countdownTicker) {
    window.clearInterval(countdownTicker)
    countdownTicker = undefined
  }

  if (timeoutTicker) {
    window.clearTimeout(timeoutTicker)
    timeoutTicker = undefined
  }
}

function startLoadingState() {
  clearTimers()
  hasTimedOut.value = false
  loadingSeconds.value = 60

  countdownTicker = window.setInterval(() => {
    if (loadingSeconds.value <= 1) {
      loadingSeconds.value = 0
      if (countdownTicker) {
        window.clearInterval(countdownTicker)
        countdownTicker = undefined
      }
      return
    }

    loadingSeconds.value -= 1
  }, 1000)

  timeoutTicker = window.setTimeout(() => {
    hasTimedOut.value = true
    if (countdownTicker) {
      window.clearInterval(countdownTicker)
      countdownTicker = undefined
    }
    loadingSeconds.value = 0
  }, 60000)
}

function retryLoading() {
  startLoadingState()
}

function goBack() {
  void router.push(returnTarget.value)
}

onMounted(() => {
  startLoadingState()
})

onBeforeUnmount(() => {
  clearTimers()
})
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)] px-3 py-5">
    <section class="relative overflow-hidden rounded-[30px] bg-[radial-gradient(circle_at_top,#ffddd8_0%,#fff4f2_38%,#fff_100%)] p-5 shadow-[0_18px_50px_rgba(218,37,29,0.12)]">
      <div class="absolute -left-10 top-8 h-32 w-32 rounded-full bg-primary/10 blur-3xl" aria-hidden="true" />
      <div class="absolute -right-12 bottom-6 h-36 w-36 rounded-full bg-amber-300/25 blur-3xl" aria-hidden="true" />

      <div class="relative mx-auto flex max-w-[420px] flex-col items-center text-center">
        <div class="mt-3 flex h-24 w-24 items-center justify-center rounded-full bg-white shadow-[0_16px_35px_rgba(218,37,29,0.14)]">
          <div class="flex h-16 w-16 items-center justify-center rounded-full border-[5px] border-primary/15 border-t-primary animate-spin">
            <span class="material-symbols-outlined text-[1.8rem] text-primary">sports_esports</span>
          </div>
        </div>

        <p class="mt-6 text-[0.72rem] font-black uppercase tracking-[0.16em] text-primary/70">Đang tải game</p>
        <h1 class="mt-2 text-[1.35rem] font-black text-slate-900">{{ gameName }}</h1>
        <p class="mt-3 max-w-[320px] text-[0.86rem] leading-6 text-slate-500">
          {{ loadingMessage }}
        </p>

        <div class="mt-6 w-full rounded-[22px] bg-white/90 p-4 shadow-[inset_0_0_0_1px_rgba(248,113,113,0.12)]">
          <div class="flex items-center justify-between text-[0.72rem] font-black uppercase tracking-[0.06em] text-slate-400">
            <span>Trạng thái kết nối</span>
            <span>{{ hasTimedOut ? 'Thất bại' : 'Đang xử lý' }}</span>
          </div>

          <div class="mt-3 h-3 overflow-hidden rounded-full bg-slate-100">
            <div
              class="h-full rounded-full bg-gradient-to-r from-[#ff8b7e] via-primary to-[#c92435] transition-all duration-1000"
              :style="{ width: `${hasTimedOut ? 100 : ((60 - loadingSeconds) / 60) * 100}%` }"
            />
          </div>

          <div class="mt-3 flex items-center justify-between text-[0.82rem] font-semibold">
            <span class="text-slate-500">
              {{ hasTimedOut ? 'Không thể mở sảnh game đối tác' : 'Đang chờ phản hồi từ nhà cung cấp' }}
            </span>
            <span class="rounded-full px-3 py-1 text-[0.76rem] font-black" :class="hasTimedOut ? 'bg-red-50 text-red-600' : 'bg-amber-50 text-amber-600'">
              {{ hasTimedOut ? 'Lỗi' : `${loadingSeconds}s` }}
            </span>
          </div>
        </div>

        <div v-if="hasTimedOut" class="mt-5 w-full rounded-[22px] border border-red-100 bg-red-50/80 px-4 py-4 text-left">
          <p class="text-[0.82rem] font-black text-red-700">Sảnh game đối tác đang gặp lỗi</p>
          <p class="mt-1 text-[0.78rem] leading-6 text-red-600">
            Kết nối tới nhà cung cấp không thành công sau 1 phút. Bạn có thể thử lại hoặc quay về trang chủ để chọn trò khác.
          </p>
        </div>

        <div class="mt-6 flex w-full flex-col gap-3 sm:flex-row">
          <button
            v-if="hasTimedOut"
            type="button"
            class="min-h-12 flex-1 rounded-[18px] bg-gradient-to-r from-[#ff7b71] to-primary px-4 text-[0.88rem] font-black text-white shadow-[0_14px_28px_rgba(218,37,29,0.24)] transition-transform active:scale-[0.98]"
            @click="retryLoading"
          >
            Tải lại sảnh game
          </button>
          <button
            type="button"
            class="min-h-12 flex-1 rounded-[18px] border border-slate-200 bg-white px-4 text-[0.88rem] font-black text-slate-700 transition-transform active:scale-[0.98]"
            @click="goBack"
          >
            Quay lại
          </button>
        </div>
      </div>
    </section>
  </div>
</template>
