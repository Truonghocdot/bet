<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import { useNotificationsStore } from '@/stores/notifications'

const tabs = ['Tất cả', 'Chưa đọc', 'Đã đọc'] as const
const activeTab = ref<(typeof tabs)[number]>('Tất cả')
const store = useNotificationsStore()

const unreadCount = computed(() => store.unreadCount)
const totalCount = computed(() => store.pagination.total)
const isLoading = computed(() => store.loading)
const isMarkingRead = computed(() => store.markingReadId)
const isResponding = computed(() => store.respondingId)
const respondingAction = computed(() => store.respondingAction)
const page = computed(() => store.pagination.page)
const totalPages = computed(() => store.pagination.totalPages)

const filteredNotifications = computed(() => {
  if (activeTab.value === 'Chưa đọc') return store.items.filter((item) => !item.is_read)
  if (activeTab.value === 'Đã đọc') return store.items.filter((item) => item.is_read)
  return store.items
})

const unreadItems = computed(() => store.items.filter((item) => !item.is_read).slice(0, 2))

function toneByReadState(isRead: boolean) {
  return isRead ? 'info' : 'warning'
}

function hasResponseFlow(item: { image_url?: string | null; audience: number }) {
  return Boolean(item.image_url) && item.audience === 2
}

function normalizedResponseStatus(item: { image_url?: string | null; audience: number; response_status?: number | null }) {
  if (!hasResponseFlow(item)) return null
  return Number(item.response_status ?? 1)
}

function responseStatusLabel(item: { image_url?: string | null; audience: number; response_status?: number | null }) {
  const status = normalizedResponseStatus(item)
  if (status === 2) return 'ĐÃ XÁC NHẬN'
  if (status === 3) return 'ĐÃ HUỶ'
  if (status === 1) return 'Chờ phản hồi'
  return null
}

function responseStatusClass(item: { image_url?: string | null; audience: number; response_status?: number | null }) {
  const status = normalizedResponseStatus(item)
  if (status === 2) return 'bg-emerald-500/10 text-emerald-700'
  if (status === 3) return 'bg-[#e64545]/10 text-[#e64545]'
  return 'bg-amber-500/10 text-amber-700'
}

function sanitizedNotificationBody(raw: string | null | undefined) {
  if (typeof window === 'undefined') return ''

  const template = document.createElement('template')
  template.innerHTML = String(raw ?? '')
  const allowedTags = new Set(['A', 'B', 'BLOCKQUOTE', 'BR', 'CODE', 'EM', 'I', 'LI', 'OL', 'P', 'S', 'STRONG', 'U', 'UL'])

  template.content.querySelectorAll('*').forEach((element) => {
    if (!allowedTags.has(element.tagName)) {
      element.replaceWith(document.createTextNode(element.textContent ?? ''))
      return
    }

    for (const attribute of [...element.attributes]) {
      if (element.tagName === 'A' && attribute.name === 'href') continue
      element.removeAttribute(attribute.name)
    }

    if (element.tagName !== 'A') return

    const href = (element.getAttribute('href') ?? '').trim()
    if (/^https?:\/\//i.test(href) || /^mailto:/i.test(href) || /^tel:/i.test(href)) {
      element.setAttribute('href', href)
    } else if (/^\/(?!\/)/.test(href)) {
      element.setAttribute('href', href)
    } else if (/^(www\.|[a-z0-9-]+(\.[a-z0-9-]+)+)(\/.*)?$/i.test(href)) {
      element.setAttribute('href', `https://${href}`)
    } else {
      element.removeAttribute('href')
      return
    }

    element.setAttribute('target', '_blank')
    element.setAttribute('rel', 'noopener noreferrer')
  })

  return template.innerHTML
}

async function load(pageNumber = 1) {
  try {
    await store.fetchList(pageNumber, store.pagination.pageSize)
    store.connectStream(pageNumber, store.pagination.pageSize)
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

async function respond(id: number, action: 'confirm' | 'cancel') {
  if (!id) return
  try {
    await store.respond(id, action)
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

onBeforeUnmount(() => {
  store.disconnectStream()
})
</script>

<template>
  <div class="space-y-5">
    <section class="rounded-[28px] bg-white p-5 shadow-[0_8px_18px_rgba(255,109,102,0.05)]">
      <div class="grid gap-4">
        <div>
          <span class="inline-flex rounded-full bg-[#e64545]/10 px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.08em] text-[#e64545]">
            Thông báo
          </span>
          <h2 class="mt-4 text-[1.55rem] font-black">Danh sách thông báo của bạn</h2>
          <p class="mt-2 max-w-[36rem] text-sm leading-6 text-on-surface-variant">
            Thông báo được gửi trực tiếp từ Hệ Thống Sàn , Bạn có thể đánh dấu đã đọc để có thể theo dõi thông báo từ Sàn Game.
          </p>
          <div class="mt-4 flex flex-wrap gap-2">
            <span class="rounded-full bg-primary/10 px-3 py-1 text-[0.68rem] font-black uppercase tracking-[0.08em] text-primary">
              Tổng {{ totalCount }}
            </span>
            <span class="rounded-full bg-[#e64545]/10 px-3 py-1 text-[0.68rem] font-black uppercase tracking-[0.08em] text-[#e64545]">
              Chưa đọc {{ unreadCount }}
            </span>
          </div>
        </div>

        <div class="grid gap-2 grid-cols-2">
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
          :class="tab === activeTab ? 'bg-primary text-white shadow-[0_12px_32px_rgba(255,109,102,0.1)]' : 'bg-surface-container-low text-on-surface-variant'"
          type="button"
          @click="activeTab = tab"
        >
          {{ tab }}
        </button>
      </div>
    </section>

    <section v-if="store.error" class="rounded-[20px] bg-[rgba(183,18,17,0.08)] px-4 py-3 text-sm font-semibold text-[#e64545]">
      {{ store.error }}
    </section>

    <section v-if="isLoading" class="rounded-[24px] bg-white p-5 text-sm font-semibold text-on-surface-variant shadow-[0_8px_18px_rgba(255,109,102,0.05)]">
      Đang tải thông báo...
    </section>

    <section v-else-if="filteredNotifications.length === 0" class="rounded-[24px] bg-white p-5 text-sm font-semibold text-on-surface-variant shadow-[0_8px_18px_rgba(255,109,102,0.05)]">
      Không có thông báo ở bộ lọc hiện tại.
    </section>

    <section v-else class="grid gap-3">
      <article
        v-for="item in filteredNotifications"
        :key="item.id"
        class="rounded-[24px] bg-white p-4 shadow-[0_8px_18px_rgba(255,109,102,0.05)]"
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
              <div class="flex flex-wrap items-center justify-end gap-1.5">
                <span
                  class="rounded-full px-2 py-1 text-[0.62rem] font-black uppercase tracking-[0.08em]"
                  :class="!item.is_read ? 'bg-primary/10 text-primary' : 'bg-surface-container-low text-on-surface-variant'"
                >
                  {{ !item.is_read ? 'Mới' : 'Đã xem' }}
                </span>
                <span
                  v-if="responseStatusLabel(item)"
                  class="rounded-full px-2 py-1 text-[0.62rem] font-black uppercase tracking-[0.08em]"
                  :class="responseStatusClass(item)"
                >
                  {{ responseStatusLabel(item) }}
                </span>
              </div>
            </div>
            <img
              v-if="item.image_url"
              :src="item.image_url"
              :alt="item.title"
              class="notification-image mt-3 h-auto w-full rounded-[18px] border border-slate-200/80 object-cover"
              loading="lazy"
              decoding="async"
            />
            <div
              v-if="item.body"
              class="notification-body mt-1.5 text-[0.76rem] leading-6 text-on-surface-variant"
              v-html="sanitizedNotificationBody(item.body)"
            />
            <div class="mt-3 flex flex-wrap items-center gap-2 text-[0.68rem] text-on-surface-variant">
              <span>{{ item.publish_at || item.created_at || '—' }}</span>
            </div>
            <div v-if="item.can_respond" class="mt-3 flex flex-wrap items-center gap-2">
              <button
                type="button"
                class="inline-flex rounded-full bg-emerald-500/10 px-3 py-1 text-[0.72rem] font-extrabold text-emerald-700 disabled:opacity-60"
                :disabled="isResponding === item.id"
                @click="respond(item.id, 'confirm')"
              >
                {{ isResponding === item.id && respondingAction === 'confirm' ? 'Đang xác nhận...' : 'Xác nhận' }}
              </button>
              <button
                type="button"
                class="inline-flex rounded-full bg-[#e64545]/10 px-3 py-1 text-[0.72rem] font-extrabold text-[#e64545] disabled:opacity-60"
                :disabled="isResponding === item.id"
                @click="respond(item.id, 'cancel')"
              >
                {{ isResponding === item.id && respondingAction === 'cancel' ? 'Đang huỷ...' : 'Huỷ' }}
              </button>
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

    <section class="rounded-[20px] bg-white p-4 shadow-[0_8px_18px_rgba(255,109,102,0.05)]">
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

<style scoped>
.notification-body :deep(a) {
  color: var(--color-primary);
  font-weight: 800;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.notification-body :deep(p) {
  margin: 0;
}

.notification-body :deep(ul),
.notification-body :deep(ol) {
  margin: 0.35rem 0 0;
  padding-left: 1rem;
}

.notification-image {
  max-height: 18rem;
}
</style>
