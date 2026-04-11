<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'

const banners = [
  'https://ossimg.vn168vn168vn.com/VN168/banner/Banner_20251209174412poex.png',
  'https://ossimg.vn168vn168vn.com/VN168/banner/Banner_202510271109361o47.jpg',
  'https://ossimg.vn168vn168vn.com/VN168/banner/Banner_20241212201908rh8b.jpg',
  'https://ossimg.vn168vn168vn.com/VN168/banner/Banner_202404191437083m86.jpg',
]

const fallback = 'https://images.unsplash.com/photo-1642790551116-18e150f248e3?w=800&q=80'

const current = ref(0)
let autoTimer: number | undefined

function next() {
  current.value = (current.value + 1) % banners.length
}

function prev() {
  current.value = (current.value - 1 + banners.length) % banners.length
}

function goTo(i: number) {
  current.value = i
}

function onImgError(e: Event) {
  ;(e.target as HTMLImageElement).src = fallback
}


onMounted(() => {
  autoTimer = window.setInterval(next, 3500)
})

onBeforeUnmount(() => {
  if (autoTimer) window.clearInterval(autoTimer)
})
</script>

<template>
  <div class="relative mx-3 mt-3 overflow-hidden rounded-[16px]" style="height: 160px;">
    <!-- Slides -->
    <div
      class="flex h-full transition-transform duration-500 ease-in-out"
      :style="{ transform: `translateX(-${current * 100}%)` }"
    >
      <img
        v-for="(src, i) in banners"
        :key="i"
        :src="src"
        :alt="`Banner ${i + 1}`"
        class="h-full w-full min-w-full object-cover"
        @error="onImgError"
      />
    </div>

    <!-- Dots -->
    <div class="absolute bottom-2 left-1/2 flex -translate-x-1/2 gap-1.5">
      <button
        v-for="(_, i) in banners"
        :key="i"
        class="h-1.5 rounded-full transition-all"
        :class="i === current ? 'w-5 bg-white' : 'w-1.5 bg-white/50'"
        @click="goTo(i)"
      />
    </div>

    <!-- Prev / Next arrows -->
    <button
      class="absolute left-2 top-1/2 -translate-y-1/2 grid h-7 w-7 place-items-center rounded-full bg-black/30 text-white backdrop-blur-sm"
      @click="prev"
    >
      <span class="material-symbols-outlined text-[1rem]">chevron_left</span>
    </button>
    <button
      class="absolute right-2 top-1/2 -translate-y-1/2 grid h-7 w-7 place-items-center rounded-full bg-black/30 text-white backdrop-blur-sm"
      @click="next"
    >
      <span class="material-symbols-outlined text-[1rem]">chevron_right</span>
    </button>
  </div>
</template>
