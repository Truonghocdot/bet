<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

import { useWalletStore } from '@/stores/wallet'

const wallet = useWalletStore()
const isOpen = ref(false)

const supportLink = computed(() => String(wallet.summary?.telegram_cskh_link ?? '').trim())
const securityContent = computed(() => String(wallet.summary?.security_notice?.content ?? '').trim())
const securityEnabled = computed(() => {
  return Boolean(wallet.summary?.security_notice?.enabled) && securityContent.value !== ''
})

onMounted(async () => {
  if (!wallet.summary && !wallet.loading) {
    try {
      await wallet.fetchSummary()
    } catch {
      // no-op
    }
  }
})

watch(securityEnabled, (enabled) => {
  if (!enabled) {
    isOpen.value = false
  }
}, { immediate: true })

function toggleSecurity(): void {
  if (!securityEnabled.value) {
    return
  }

  isOpen.value = !isOpen.value
}
</script>

<template>
  <section class="mt-2">
    <div class="grid grid-cols-2 gap-3">
      <a
        :href="supportLink || '#'"
        target="_blank"
        rel="noopener noreferrer"
        class="grid min-h-[84px] place-items-center gap-1 rounded-[18px] bg-white font-extrabold shadow-[0_8px_20px_rgba(255,109,102,0.05)]"
        :class="{ 'pointer-events-none opacity-60': !supportLink }"
      >
        <span class="material-symbols-outlined text-primary">support_agent</span>
        <span>CSKH</span>
      </a>
      <button
        type="button"
        class="grid min-h-[84px] place-items-center gap-1 rounded-[18px] bg-white font-extrabold shadow-[0_8px_20px_rgba(255,109,102,0.05)] transition-transform active:scale-95 disabled:cursor-not-allowed disabled:opacity-50"
        @click="toggleSecurity"
      >
        <span class="material-symbols-outlined text-primary">security</span>
        <span>Bảo mật</span>
      </button>
    </div>

    <transition name="auth-security-fade">
      <div
        v-if="isOpen && securityEnabled"
        class="fixed inset-0 z-[9999] flex items-center justify-center bg-black/50 px-4 py-6 backdrop-blur-sm"
        @click="toggleSecurity"
      >
        <div
          class="w-full max-w-[520px] rounded-[24px] bg-white shadow-[0_20px_60px_rgba(15,23,42,0.24)]"
          @click.stop
        >
          <div class="flex items-start justify-between gap-4 border-b border-slate-100 px-5 py-4">
            <div>
              <p class="m-0 text-[0.72rem] font-black uppercase tracking-[0.12em] text-primary/70">Bảo mật</p>
              <h3 class="m-0 mt-1 text-[1.05rem] font-black text-on-surface">Thông tin bảo mật</h3>
            </div>
            <button
              type="button"
              class="grid h-10 w-10 place-items-center rounded-full bg-slate-100 text-slate-500 transition-transform active:scale-95"
              @click="toggleSecurity"
            >
              <span class="material-symbols-outlined text-[1.1rem]">close</span>
            </button>
          </div>

          <div class="max-h-[60vh] overflow-y-auto px-5 py-4">
            <div class="rounded-[18px] bg-slate-50 px-4 py-3 text-sm font-bold leading-6 text-on-surface-variant whitespace-pre-line">
              {{ securityContent }}
            </div>
          </div>

          <div class="flex justify-end px-5 pb-5">
            <button
              type="button"
              class="min-h-11 rounded-[14px] bg-primary px-5 text-[0.82rem] font-black text-white transition-transform active:scale-95"
              @click="toggleSecurity"
            >
              Đã hiểu
            </button>
          </div>
        </div>
      </div>
    </transition>
  </section>
</template>

<style scoped>
.auth-security-fade-enter-active,
.auth-security-fade-leave-active {
  transition: opacity 0.22s ease;
}

.auth-security-fade-enter-from,
.auth-security-fade-leave-to {
  opacity: 0;
}
</style>
