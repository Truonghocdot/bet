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
  <section class="mt-2 space-y-3">
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
        :disabled="!securityEnabled"
        @click="toggleSecurity"
      >
        <span class="material-symbols-outlined text-primary">security</span>
        <span>Bảo mật</span>
      </button>
    </div>

    <transition name="auth-security-fade">
      <div
        v-if="isOpen && securityEnabled"
        class="rounded-[18px] bg-white px-4 py-3 text-sm font-bold leading-6 text-on-surface-variant shadow-[0_8px_20px_rgba(255,109,102,0.05)] whitespace-pre-line"
      >
        {{ securityContent }}
      </div>
    </transition>
  </section>
</template>

<style scoped>
.auth-security-fade-enter-active,
.auth-security-fade-leave-active {
  transition: all 0.22s ease;
}

.auth-security-fade-enter-from,
.auth-security-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
