<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'

import { request, type ApiError } from '@/shared/api/http'
import { env } from '@/shared/config/env'
import { useAuthStore } from '@/stores/auth'

type ChatMessage = {
  id: number
  display_name: string
  body: string
  actor_type: 'user' | 'bot' | 'admin'
  created_at: string
}

type ChatHistoryResponse = {
  items: ChatMessage[]
  next_cursor?: number | null
}

const props = withDefaults(defineProps<{ popup?: boolean }>(), {
  popup: false,
})
const emit = defineEmits<{ close: [] }>()

const auth = useAuthStore()
const messages = ref<ChatMessage[]>([])
const nextCursor = ref<number | null>(null)
const body = ref('')
const loading = ref(true)
const loadingOlder = ref(false)
const sending = ref(false)
const error = ref('')
const connected = ref(false)
const messageList = ref<HTMLElement | null>(null)

let socket: WebSocket | null = null
let reconnectTimer: number | null = null
let stopped = false

const canSend = computed(() => body.value.trim().length > 0 && body.value.length <= 280 && !sending.value)
const connectionLabel = computed(() => connected.value ? 'Đang kết nối' : 'Đang kết nối lại')

function close() {
  if (props.popup) emit('close')
}

function mergeMessages(items: ChatMessage[], mode: 'replace' | 'prepend' | 'append') {
  const byID = new Map(messages.value.map((item) => [item.id, item]))
  for (const item of items) byID.set(item.id, item)
  const merged = [...byID.values()].sort((a, b) => a.id - b.id)
  messages.value = mode === 'replace' ? items.sort((a, b) => a.id - b.id) : merged
}

async function loadInitial() {
  loading.value = true
  error.value = ''
  try {
    const response = await request<ChatHistoryResponse>('GET', '/v1/chat/global/messages?limit=50', { token: auth.accessToken })
    mergeMessages(response.items ?? [], 'replace')
    nextCursor.value = response.next_cursor ?? null
    await scrollToBottom()
  } catch (cause) {
    error.value = (cause as ApiError).message || 'Không thể tải lịch sử chat.'
  } finally {
    loading.value = false
  }
}

async function loadOlder() {
  if (!nextCursor.value || loadingOlder.value) return
  const element = messageList.value
  const previousHeight = element?.scrollHeight ?? 0
  loadingOlder.value = true
  try {
    const response = await request<ChatHistoryResponse>('GET', `/v1/chat/global/messages?before=${nextCursor.value}&limit=50`, { token: auth.accessToken })
    mergeMessages(response.items ?? [], 'prepend')
    nextCursor.value = response.next_cursor ?? null
    await nextTick()
    if (element) element.scrollTop = element.scrollHeight - previousHeight
  } catch (cause) {
    error.value = (cause as ApiError).message || 'Không thể tải thêm tin nhắn.'
  } finally {
    loadingOlder.value = false
  }
}

async function refreshLatest() {
  try {
    const response = await request<ChatHistoryResponse>('GET', '/v1/chat/global/messages?limit=50', { token: auth.accessToken })
    const shouldScroll = isNearBottom()
    mergeMessages(response.items ?? [], 'append')
    if (!nextCursor.value) nextCursor.value = response.next_cursor ?? null
    if (shouldScroll) await scrollToBottom()
  } catch {
    // Reconnect keeps retrying; the existing history remains readable offline.
  }
}

function isNearBottom() {
  const element = messageList.value
  return !element || element.scrollHeight - element.scrollTop - element.clientHeight < 96
}

async function scrollToBottom() {
  await nextTick()
  const element = messageList.value
  if (element) element.scrollTop = element.scrollHeight
}

function handleScroll() {
  if ((messageList.value?.scrollTop ?? 1) <= 24) void loadOlder()
}

async function send() {
  const value = body.value.trim()
  if (!value || sending.value) return
  sending.value = true
  error.value = ''
  try {
    const response = await request<{ message: ChatMessage }>('POST', '/v1/chat/global/messages', {
      body: { body: value },
      token: auth.accessToken,
    })
    const shouldScroll = isNearBottom()
    mergeMessages([response.message], 'append')
    body.value = ''
    if (shouldScroll) await scrollToBottom()
  } catch (cause) {
    error.value = (cause as ApiError).message || 'Không thể gửi tin nhắn.'
  } finally {
    sending.value = false
  }
}

function handleRealtime(raw: string) {
  let event: { event?: string; data?: ChatMessage | { id?: number } }
  try { event = JSON.parse(raw) } catch { return }
  if (event.event === 'chat.message.created' && event.data && 'body' in event.data) {
    const shouldScroll = isNearBottom()
    mergeMessages([event.data as ChatMessage], 'append')
    if (shouldScroll) void scrollToBottom()
    return
  }
  if ((event.event === 'chat.message.hidden' || event.event === 'chat.message.deleted') && event.data && 'id' in event.data) {
    messages.value = messages.value.filter((item) => item.id !== Number(event.data?.id))
  }
}

function webSocketURL() {
  const source = env.apiBaseUrl || window.location.origin
  const url = new URL(source, window.location.origin)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = `${url.pathname.replace(/\/$/, '')}/v1/chat/global/ws`
  url.searchParams.set('access_token', auth.accessToken)
  return url.toString()
}

function connect() {
  if (stopped || !auth.accessToken) return
  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) return
  socket = new WebSocket(webSocketURL())
  socket.onopen = () => {
    connected.value = true
    void refreshLatest()
  }
  socket.onmessage = (event) => handleRealtime(String(event.data))
  socket.onerror = () => { connected.value = false }
  socket.onclose = () => {
    connected.value = false
    socket = null
    if (!stopped) reconnectTimer = window.setTimeout(connect, 4000)
  }
}

onMounted(async () => {
  await loadInitial()
  connect()
})

onBeforeUnmount(() => {
  stopped = true
  if (reconnectTimer !== null) window.clearTimeout(reconnectTimer)
  socket?.close()
})
</script>

<template>
  <section
    :class="props.popup
      ? 'fixed inset-0 z-[120] flex items-end bg-black/45 p-2 backdrop-blur-sm sm:items-center sm:justify-center'
      : 'mx-auto flex min-h-[calc(100dvh-11rem)] w-full max-w-3xl flex-col bg-white pb-4'"
    :role="props.popup ? 'dialog' : undefined"
    :aria-modal="props.popup ? 'true' : undefined"
    aria-label="Chat Global"
    @click.self="close"
  >
    <div
      :class="props.popup
        ? 'flex h-[min(680px,calc(100dvh-1rem))] w-full max-w-md flex-col overflow-hidden rounded-t-[8px] bg-white shadow-2xl sm:rounded-[8px]'
        : 'flex min-h-[calc(100dvh-11rem)] w-full flex-col'"
    >
      <header class="flex items-center justify-between border-b border-slate-200 px-4 py-3">
        <div class="min-w-0">
          <h1 class="text-[1rem] font-bold text-slate-900">Chat Global</h1>
          <p class="mt-0.5 text-[0.72rem] text-slate-500">{{ connectionLabel }}</p>
        </div>
        <div class="flex items-center gap-3">
          <span class="h-2.5 w-2.5 rounded-full" :class="connected ? 'bg-emerald-500' : 'bg-amber-400'" aria-hidden="true" />
          <button v-if="props.popup" type="button" class="grid h-9 w-9 place-items-center text-slate-500 transition-colors hover:bg-slate-100" aria-label="Đóng chat" @click="close">
            <span class="material-symbols-outlined text-[1.1rem]">close</span>
          </button>
        </div>
      </header>

      <p v-if="error" class="mx-4 mt-3 border border-red-200 bg-red-50 px-3 py-2 text-[0.8rem] text-red-700">
        {{ error }}
      </p>

      <div ref="messageList" class="min-h-0 flex-1 space-y-3 overflow-y-auto px-4 py-4" @scroll="handleScroll">
        <div v-if="loading" class="py-12 text-center text-[0.82rem] text-slate-500">Đang tải tin nhắn...</div>
        <button v-else-if="nextCursor" type="button" class="mx-auto block text-[0.78rem] font-semibold text-primary" :disabled="loadingOlder" @click="loadOlder">
          {{ loadingOlder ? 'Đang tải...' : 'Tải tin cũ hơn' }}
        </button>
        <p v-else-if="!messages.length" class="py-12 text-center text-[0.82rem] text-slate-500">Chưa có tin nhắn nào.</p>

        <article v-for="message in messages" :key="message.id" class="flex gap-2.5">
          <div class="grid h-8 w-8 shrink-0 place-items-center rounded-full bg-slate-100 text-slate-500">
            <span class="material-symbols-outlined text-[1rem]">person</span>
          </div>
          <div class="min-w-0">
            <div class="flex items-baseline gap-2">
              <strong class="truncate text-[0.78rem] font-semibold text-slate-800">{{ message.display_name }}</strong>
              <time class="shrink-0 text-[0.67rem] text-slate-400">{{ new Date(message.created_at).toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' }) }}</time>
            </div>
            <p class="mt-1 whitespace-pre-wrap break-words text-[0.88rem] leading-5 text-slate-700">{{ message.body }}</p>
          </div>
        </article>
      </div>

      <form class="border-t border-slate-200 px-4 py-3" @submit.prevent="send">
        <div class="flex items-end gap-2">
          <textarea v-model="body" rows="2" maxlength="280" class="min-h-11 flex-1 resize-none border border-slate-300 bg-white px-3 py-2 text-[0.88rem] text-slate-900 outline-none focus:border-primary" placeholder="Nhập tin nhắn" :disabled="sending" />
          <button type="submit" class="grid h-11 w-11 shrink-0 place-items-center bg-primary text-white disabled:opacity-50" :disabled="!canSend" aria-label="Gửi tin nhắn">
            <span class="material-symbols-outlined">send</span>
          </button>
        </div>
        <div class="mt-1 text-right text-[0.68rem] text-slate-400">{{ body.length }}/280</div>
      </form>
    </div>
  </section>
</template>
