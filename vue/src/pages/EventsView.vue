<script setup lang="ts">
import { computed, onMounted } from 'vue'

import { useWheelInvitationsStore, type WheelInvitation } from '@/stores/wheelInvitations'

const wheel = useWheelInvitationsStore()
const sortedItems = computed(() => wheel.items)

function statusLabel(item: WheelInvitation) {
  return ({ pending: 'Sẵn sàng', started: 'Đang tham gia', completed: 'Đã hoàn thành', expired: 'Đã hết hạn', revoked: 'Đã thu hồi' } as Record<string, string>)[item.status] ?? item.status
}

function canOpen(item: WheelInvitation) {
  return item.status === 'pending' || item.status === 'started' || item.status === 'completed'
}

function formatTime(value?: string | null) {
  if (!value) return 'Theo thời hạn chiến dịch'
  return new Intl.DateTimeFormat('vi-VN', { dateStyle: 'short', timeStyle: 'short' }).format(new Date(value))
}

onMounted(() => void wheel.fetchInvitations())
</script>

<template>
  <section class="mx-auto min-h-[calc(100dvh-11rem)] w-full max-w-3xl px-1 pb-8">
    <header class="border-b border-slate-200 px-3 py-5">
      <p class="text-[0.72rem] font-bold uppercase text-primary">Dành riêng cho bạn</p>
      <h1 class="mt-1 text-xl font-black text-slate-900">Sự kiện</h1>
    </header>

    <p v-if="wheel.error" class="mx-3 mt-4 border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{{ wheel.error }}</p>
    <div v-if="wheel.loading && !sortedItems.length" class="grid min-h-48 place-items-center text-sm text-slate-500">Đang tải sự kiện...</div>
    <div v-else-if="!sortedItems.length" class="grid min-h-48 place-items-center px-6 text-center text-sm text-slate-500">Bạn chưa có lời mời sự kiện nào.</div>

    <div v-else class="divide-y divide-slate-200">
      <article v-for="item in sortedItems" :key="item.id" class="flex items-center gap-3 px-3 py-4">
        <div class="grid h-11 w-11 shrink-0 place-items-center bg-red-50 text-primary">
          <span class="material-symbols-outlined">featured_seasonal_and_gifts</span>
        </div>
        <div class="min-w-0 flex-1">
          <h2 class="truncate text-sm font-black text-slate-900">{{ item.campaign_name }}</h2>
          <p class="mt-1 text-xs text-slate-500">{{ statusLabel(item) }} · Hạn {{ formatTime(item.expires_at) }}</p>
        </div>
        <button v-if="canOpen(item)" type="button" class="min-h-10 shrink-0 bg-primary px-3 text-xs font-bold text-white disabled:opacity-60" :disabled="wheel.launchingId === item.id" @click="wheel.launch(item)">
          {{ wheel.launchingId === item.id ? 'Đang mở' : 'Tham gia' }}
        </button>
      </article>
    </div>
  </section>
</template>
