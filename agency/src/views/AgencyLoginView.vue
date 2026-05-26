<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAgencyAuthStore } from '@/stores/auth'

const auth = useAgencyAuthStore()
const router = useRouter()
const account = ref('')
const password = ref('')
const showPassword = ref(false)

async function submit() {
  try {
    await auth.login(account.value.trim(), password.value)
    await router.replace({ name: 'agency-dashboard' })
  } catch {
    // error in store
  }
}
</script>

<template>
  <main class="min-h-screen px-4 py-8">
    <section class="mx-auto max-w-md rounded-[32px] border border-white/60 bg-[rgba(255,255,255,0.88)] p-8 shadow-[0_30px_80px_rgba(15,23,42,0.12)] backdrop-blur">
      <div class="flex items-center justify-between gap-3">
        <div>
          <p class="m-0 text-xs font-bold uppercase tracking-[0.18em] text-slate-400">Đăng nhập</p>
          <h1 class="mt-2 text-2xl font-bold text-slate-900">Agency Access</h1>
        </div>
        <div class="grid h-12 w-12 place-items-center rounded-[18px] bg-slate-950 text-white">
          <span class="material-symbols-outlined">shield_lock</span>
        </div>
      </div>

      <form class="mt-8 space-y-4" @submit.prevent="submit">
        <label class="block">
          <span class="mb-2 block text-sm font-bold text-slate-700">Tài khoản</span>
          <input
            v-model="account"
            type="text"
            autocomplete="username"
            class="min-h-14 w-full rounded-[20px] border border-black/8 bg-white px-4 outline-none"
            placeholder="Số điện thoại hoặc email"
          />
        </label>

        <label class="block">
          <span class="mb-2 block text-sm font-bold text-slate-700">Mật khẩu</span>
          <div class="grid min-h-14 grid-cols-[1fr_auto] items-center rounded-[20px] border border-black/8 bg-white">
            <input
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              autocomplete="current-password"
              class="min-h-14 w-full bg-transparent px-4 outline-none"
              placeholder="Nhập mật khẩu"
            />
            <button
              type="button"
              class="mr-2 grid h-10 w-10 place-items-center rounded-full text-slate-500 transition-colors hover:text-slate-800"
              :aria-label="showPassword ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'"
              @click="showPassword = !showPassword"
            >
              <span class="material-symbols-outlined text-[1.15rem]">
                {{ showPassword ? 'visibility_off' : 'visibility' }}
              </span>
            </button>
          </div>
        </label>

        <p v-if="auth.error" class="rounded-[18px] bg-rose-50 px-4 py-3 text-sm font-semibold text-rose-600">
          {{ auth.error }}
        </p>

        <button
          type="submit"
          class="min-h-14 w-full rounded-[22px] bg-slate-950 text-sm font-bold uppercase tracking-[0.14em] text-white transition-transform active:scale-[0.99] disabled:opacity-60"
          :disabled="auth.loading"
        >
          {{ auth.loading ? 'Đang xử lý...' : 'Đăng nhập agency' }}
        </button>
      </form>
    </section>
  </main>
</template>
