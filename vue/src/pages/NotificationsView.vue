<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

import { formatViDateTime } from '@/shared/lib/date'
import { useNotificationsStore } from '@/stores/notifications'

const tabs = ['Tất cả', 'Chưa đọc', 'Đã đọc'] as const
const activeTab = ref<(typeof tabs)[number]>('Tất cả')
const store = useNotificationsStore()

const unreadCount = computed(() => store.unreadCount)
const totalCount = computed(() => store.pagination.total)
const isLoading = computed(() => store.loading)
const isMarkingRead = computed(() => store.markingReadId)
const page = computed(() => store.pagination.page)
const totalPages = computed(() => store.pagination.totalPages)

const filteredNotifications = computed(() => {
  if (activeTab.value === 'Chưa đọc') return store.items.filter((item) => !item.is_read)
  if (activeTab.value === 'Đã đọc') return store.items.filter((item) => item.is_read)
  return store.items
})

const unreadItems = computed(() => store.items.filter((item) => !item.is_read).slice(0, 2))

function audienceLabel(audience: number) {
  return audience === 1 ? 'Toàn bộ người dùng' : 'Nhắm theo tài khoản'
}

function toneByReadState(isRead: boolean) {
  return isRead ? 'info' : 'warning'
}

async function load(pageNumber = 1) {
  try {
    await store.fetchList(pageNumber, store.pagination.pageSize)
  } catch {
    // message already populated in store.error
  }
}

async function markRead(id: number) {
  if (!id) return
  try {
    await store.markRead(id)
  } catch {
    // message already populated in store.error
  }
}

function prevPage() {
  if (page.value <= 1 || isLoading.value) return
  void load(page.value - 1)
}

function nextPage() {
  if (page.value >= totalPages.value || isLoading.value) return
  void load(page.value + 1)
}

watch(activeTab, () => {
  // keep UX consistent: switching tabs keeps current API page,
  // filtering is client-side on current page data.
})

onMounted(() => {
  void load(1)
})
</script>

<template>
  <div class="space-y-5 md:space-y-6">
    <section class="rounded-[28px] bg-white p-5 shadow-[0_8px_18px_rgba(0,78,219,0.05)] md:p-6">
      <div class="grid gap-4 xl:grid-cols-[1fr_auto] xl:items-end">
        <div>
          <span class="inline-flex rounded-full bg-[#b71211]/10 px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.08em] text-[#b71211]">
            Thông báo
          </span>
          <h2 class="mt-4 text-[1.55rem] font-black md:text-[1.8rem]">Danh sách thông báo của bạn</h2>
          <p class="mt-2 max-w-[36rem] text-sm leading-6 text-on-surface-variant">
            Dữ liệu đang được lấy trực tiếp từ API. Bạn có thể đánh dấu đã đọc để đồng bộ trạng thái theo tài khoản.
          </p>
          <div class="mt-4 flex flex-wrap gap-2">
            <span class="rounded-full bg-primary/10 px-3 py-1 text-[0.68rem] font-black uppercase tracking-[0.08em] text-primary">
              Tổng {{ totalCount }}
            </span>
            <span class="rounded-full bg-[#b71211]/10 px-3 py-1 text-[0.68rem] font-black uppercase tracking-[0.08em] text-[#b71211]">
              Chưa đọc {{ unreadCount }}
            </span>
          </div>
        </div>

        <div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-1">
          <div class="rounded-[22px] bg-primary/10 px-4 py-3 text-left">
            <p class="m-0 text-[0.7rem] uppercase tracking-[0.12em] text-primary/70">Chưa đọc</p>
            <strong class="mt-1 block text-[1.35rem] font-black text-primary">{{ unreadCount }}</strong>
          </div>
          <div class="rounded-[22px] bg-surface-container-low px-4 py-3 text-left">
            <p class="m-0 text-[0.7rem] uppercase tracking-[0.12em] text-on-surface-variant">Đã đọc</p>
            <strong class="mt-1 block text-[1.35rem] font-black text-on-surface">{{ totalCount - unreadCount }}</strong>
          </div>
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

    <section v-if="store.error" class="rounded-[20px] bg-[rgba(183,18,17,0.08)] px-4 py-3 text-sm font-semibold text-[#b71211]">
      {{ store.error }}
    </section>

    <section v-if="isLoading" class="rounded-[24px] bg-white p-5 text-sm font-semibold text-on-surface-variant shadow-[0_8px_18px_rgba(0,78,219,0.05)]">
      Đang tải thông báo...
    </section>

    <section v-else-if="filteredNotifications.length === 0" class="rounded-[24px] bg-white p-5 text-sm font-semibold text-on-surface-variant shadow-[0_8px_18px_rgba(0,78,219,0.05)]">
      Không có thông báo ở bộ lọc hiện tại.
    </section>

    <section v-else class="grid gap-3 md:grid-cols-2">
      <article
        v-for="item in filteredNotifications"
        :key="item.id"
        class="rounded-[24px] bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)]"
      >
        <div class="flex items-start gap-3">
          <div
            class="grid h-11 w-11 place-items-center rounded-[16px] text-white"
            :class="{
              'bg-primary': toneByReadState(item.is_read) === 'info',
              'bg-amber-500': toneByReadState(item.is_read) === 'warning',
            }"
          >
            <span class="material-symbols-outlined text-[1.05rem]">
              {{ item.is_read ? 'info' : 'notifications_active' }}
            </span>
          </div>

          <div class="min-w-0 flex-1">
            <div class="flex items-center justify-between gap-2">
              <strong class="text-[0.9rem] font-black">{{ item.title }}</strong>
              <span
                class="rounded-full px-2 py-1 text-[0.62rem] font-black uppercase tracking-[0.08em]"
                :class="!item.is_read ? 'bg-primary/10 text-primary' : 'bg-surface-container-low text-on-surface-variant'"
              >
                {{ !item.is_read ? 'Mới' : 'Đã xem' }}
              </span>
            </div>
            <p class="mt-1.5 text-[0.76rem] leading-6 text-on-surface-variant">{{ item.body }}</p>
            <div class="mt-3 flex flex-wrap items-center gap-2 text-[0.68rem] text-on-surface-variant">
              <span class="rounded-full bg-surface-container-low px-3 py-1 font-bold">{{ audienceLabel(item.audience) }}</span>
              <span>{{ formatViDateTime(item.publish_at || item.created_at) }}</span>
            </div>
            <button
              v-if="!item.is_read"
              type="button"
              class="mt-3 inline-flex rounded-full bg-primary/10 px-3 py-1 text-[0.72rem] font-extrabold text-primary disabled:opacity-60"
              :disabled="isMarkingRead === item.id"
              @click="markRead(item.id)"
            >
              {{ isMarkingRead === item.id ? 'Đang cập nhật...' : 'Đánh dấu đã đọc' }}
            </button>
          </div>
        </div>
      </article>
    </section>

    <section class="rounded-[20px] bg-white p-4 shadow-[0_8px_18px_rgba(0,78,219,0.05)]">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="text-[0.78rem] font-semibold text-on-surface-variant">
          Trang {{ page }} / {{ totalPages }}
        </div>
        <div class="flex items-center gap-2">
          <button
            type="button"
            class="rounded-full border border-slate-200 px-3 py-1.5 text-[0.74rem] font-extrabold text-on-surface disabled:opacity-50"
            :disabled="page <= 1 || isLoading"
            @click="prevPage"
          >
            Trang trước
          </button>
          <button
            type="button"
            class="rounded-full border border-slate-200 px-3 py-1.5 text-[0.74rem] font-extrabold text-on-surface disabled:opacity-50"
            :disabled="page >= totalPages || isLoading"
            @click="nextPage"
          >
            Trang sau
          </button>
        </div>
      </div>
      <p v-if="unreadItems.length > 0" class="mt-3 text-[0.72rem] text-on-surface-variant">
        Ưu tiên xử lý: {{ unreadItems.map((item) => item.title).join(' • ') }}
      </p>
    </section>
  </div>
</template>
