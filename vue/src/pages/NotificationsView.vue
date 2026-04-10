<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { formatViDateTime } from '@/shared/lib/date'
import { notificationItems, getUnreadCount } from '@/data/site'

const tabs = ['Tất cả', 'Chưa đọc', 'Đã đọc']
const activeTab = ref(tabs[0])

const unreadCount = computed(() => getUnreadCount())

const filteredNotifications = computed(() => {
  if (activeTab.value === 'Chưa đọc') return notificationItems.filter((item) => item.unread)
  if (activeTab.value === 'Đã đọc') return notificationItems.filter((item) => !item.unread)
  return notificationItems
})
</script>

<template>
  <div class="space-y-5 md:space-y-6">
    <section class="rounded-[28px] bg-white p-5 shadow-[0_8px_18px_rgba(0,78,219,0.05)] md:p-6">
      <div class="flex items-start justify-between gap-3">
        <div>
          <span class="inline-flex rounded-full bg-[#b71211]/10 px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.08em] text-[#b71211]">
            Thông báo
          </span>
          <h2 class="mt-4 text-[1.55rem] font-black md:text-[1.8rem]">Danh sách thông báo của bạn</h2>
          <p class="mt-2 text-sm leading-6 text-on-surface-variant">
            Tổng hợp alert hệ thống, tin tức và nhắc nhở vận hành.
          </p>
        </div>

        <div class="rounded-[22px] bg-primary/10 px-4 py-3 text-right">
          <p class="m-0 text-[0.7rem] uppercase tracking-[0.12em] text-primary/70">Chưa đọc</p>
          <strong class="mt-1 block text-[1.35rem] font-black text-primary">{{ unreadCount }}</strong>
        </div>
      </div>
    </section>

    <section class="overflow-x-auto pb-2 no-scrollbar">
      <div class="flex min-w-max gap-2">
        <button
          v-for="tab in tabs"
          :key="tab"
          class="rounded-full px-5 py-2.5 text-[0.78rem] font-bold whitespace-nowrap transition-colors"
          :class="tab === activeTab ? 'bg-primary text-white shadow-[0_12px_32px_rgba(0,78,219,0.1)]' : 'bg-surface-container-low text-on-surface-variant'"
          type="button"
          @click="activeTab = tab"
        >
          {{ tab }}
        </button>
      </div>
    </section>

    <section class="space-y-3">
      <article
        v-for="item in filteredNotifications"
        :key="item.id"
        class="rounded-[24px] bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)]"
      >
        <div class="flex items-start gap-3">
          <div
            class="grid h-11 w-11 place-items-center rounded-[16px] text-white"
            :class="{
              'bg-emerald-500': item.tone === 'success',
              'bg-primary': item.tone === 'info',
              'bg-amber-500': item.tone === 'warning',
            }"
          >
            <span class="material-symbols-outlined text-[1.05rem]">
              {{ item.tone === 'success' ? 'task_alt' : item.tone === 'warning' ? 'notifications_active' : 'info' }}
            </span>
          </div>

          <div class="min-w-0 flex-1">
            <div class="flex items-center justify-between gap-2">
              <strong class="text-[0.9rem] font-black">{{ item.title }}</strong>
              <span
                class="rounded-full px-2 py-1 text-[0.62rem] font-black uppercase tracking-[0.08em]"
                :class="item.unread ? 'bg-primary/10 text-primary' : 'bg-surface-container-low text-on-surface-variant'"
              >
                {{ item.unread ? 'Mới' : 'Đã xem' }}
              </span>
            </div>
            <p class="mt-1.5 text-[0.76rem] leading-6 text-on-surface-variant">{{ item.body }}</p>
            <div class="mt-3 flex flex-wrap items-center gap-2 text-[0.68rem] text-on-surface-variant">
              <span class="rounded-full bg-surface-container-low px-3 py-1 font-bold">{{ item.category }}</span>
              <span>{{ formatViDateTime(item.createdAt) }}</span>
            </div>
            <RouterLink
              v-if="item.relatedSlug"
              :to="`/news/${item.relatedSlug}`"
              class="mt-3 inline-flex text-[0.74rem] font-extrabold text-primary"
            >
              Xem tin liên quan
            </RouterLink>
          </div>
        </div>
      </article>
    </section>
  </div>
</template>

