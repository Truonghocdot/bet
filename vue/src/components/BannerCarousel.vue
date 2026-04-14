<script setup lang="ts">
import { computed, ref, onMounted, onBeforeUnmount, watch } from 'vue'

type BannerInput = {
  id?: number
  title?: string
  image_url: string
  link_url?: string
}

const props = withDefaults(
  defineProps<{
    banners?: BannerInput[]
  }>(),
  {
    banners: () => [],
  },
)

const fallback = '/banner.png'
const fallbackList = [
  {
    id: 0,
    title: 'Banner',
    image_url: fallback,
    link_url: '',
  },
]
const banners = computed(() => (props.banners.length > 0 ? props.banners : fallbackList))

const current = ref(0)
let autoTimer: number | undefined

function next() {
  current.value = (current.value + 1) % banners.value.length
}

function prev() {
  current.value = (current.value - 1 + banners.value.length) % banners.value.length
}

function goTo(i: number) {
  current.value = i
}

function onImgError(e: Event) {
  ;(e.target as HTMLImageElement).src = fallback
}

watch(
  () => props.banners,
  () => {
    current.value = 0
  },
  { deep: true },
)

onMounted(() => {
  autoTimer = window.setInterval(next, 3500)
})

onBeforeUnmount(() => {
  if (autoTimer) window.clearInterval(autoTimer)
})
</script>

<template>
  <div class="relative mx-3 mt-3 overflow-hidden rounded-[20px] shadow-sm bg-slate-100 aspect-[2.1/1]">
    <!-- Slides -->
    <div
      class="flex h-full transition-transform duration-600 ease-[cubic-bezier(0.25,1,0.5,1)]"
      :style="{ transform: `translateX(-${current * 100}%)` }"
    >
      <div
        v-for="(banner, i) in banners"
        :key="banner.id ?? i"
        class="h-full w-full min-w-full relative"
      >
        <img
          :src="banner.image_url"
          :alt="banner.title || `Banner ${i + 1}`"
          class="h-full w-full object-cover block"
          loading="lazy"
          decoding="async"
          @error="onImgError"
        />
        <!-- Optional overlay to make text more readable if added later -->
        <div class="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent pointer-events-none" />
      </div>
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
