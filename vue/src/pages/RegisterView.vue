<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/auth'
import { normalizeVNPhone } from '@/shared/lib/phone'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const name = ref('')
const phone = ref('')
const password = ref('')
const refCode = ref(typeof route.query.ref_code === 'string' ? route.query.ref_code : '')
const submitError = ref('')
const showPassword = ref(false)

const canSubmit = computed(() => Boolean(name.value && phone.value && password.value))

async function handleRegister() {
  submitError.value = ''
  try {
    await auth.register({
      name: name.value.trim(),
      phone: normalizeVNPhone(phone.value),
      password: password.value,
      ref_code: refCode.value.trim() || undefined,
    })

    await router.replace('/account')
  } catch (e: any) {
    submitError.value = auth.error || 'Đăng ký thất bại'
  }
}
</script>

<template>
  <div class="space-y-5">
    <header class="grid min-h-12 grid-cols-[36px_1fr_60px] items-center md:min-h-[52px]">
      <button class="grid h-9 w-9 place-items-center text-primary transition-transform active:scale-95" type="button" @click="router.back()">
        <span class="material-symbols-outlined">arrow_back</span>
      </button>
      <h1 class="m-0 text-center text-[1.1rem] font-black text-primary md:text-[1.2rem]">Đăng ký</h1>
      <RouterLink class="justify-self-end text-right text-sm font-extrabold text-primary" to="/auth">Đăng nhập</RouterLink>
    </header>

    <section class="text-center">
      <div class="mx-auto mb-4 grid h-[72px] w-[72px] place-items-center rounded-[20px] bg-white shadow-[0_10px_24px_rgba(255,109,102,0.12)]">
        <span class="text-[1.1rem] font-black italic tracking-[-0.06em] text-primary">ff789</span>
      </div>
      <h2 class="m-0 text-[1.55rem] font-black">Tạo tài khoản mới</h2>
      <p class="mt-1.5 text-sm text-on-surface-variant">Đăng ký bằng số điện thoại </p>
    </section>

    <section v-if="submitError" class="rounded-2xl bg-secondary/10 p-4 text-sm font-bold text-on-secondary-container">
      {{ submitError }}
    </section>

    <form class="space-y-3" @submit.prevent="handleRegister">
      <label class="grid min-h-[58px] items-center overflow-hidden rounded-[18px] bg-white shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
        <input v-model="name" class="min-w-0 border-0 bg-transparent px-4 py-4 outline-none" type="text" autocomplete="name" placeholder="Họ và tên" />
      </label>

      <label class="grid min-h-[58px] grid-cols-[auto_1fr] items-center overflow-hidden rounded-[18px] bg-white shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
        <span class="border-r border-slate-200/90 px-4 font-extrabold">+84</span>
        <input v-model="phone" class="min-w-0 border-0 bg-transparent px-4 py-4 outline-none" type="tel" inputmode="tel" autocomplete="tel" placeholder="Số điện thoại" />
      </label>

      <label class="grid min-h-[58px] grid-cols-[1fr_auto] items-center overflow-hidden rounded-[18px] bg-white shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
        <input
          v-model="password"
          class="min-w-0 border-0 bg-transparent px-4 py-4 outline-none"
          :type="showPassword ? 'text' : 'password'"
          autocomplete="new-password"
          placeholder="Mật khẩu"
        />
        <button type="button" class="grid h-full w-[52px] place-items-center text-on-surface-variant" @click="showPassword = !showPassword">
          <span class="material-symbols-outlined">{{ showPassword ? 'visibility_off' : 'visibility' }}</span>
        </button>
      </label>

      <label class="grid min-h-[58px] items-center overflow-hidden rounded-[18px] bg-white shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
        <input v-model="refCode" class="min-w-0 border-0 bg-transparent px-4 py-4 outline-none" type="text" placeholder="Mã giới thiệu (không bắt buộc)" />
      </label>

      <div class="flex flex-row w-full items-center justify-center">
        <button
          class="min-h-14 rounded-[18px] px-2 bg-red-500 font-black text-white shadow-[0_8px_20px_rgba(255,109,102,0.18)] transition-transform active:scale-95 disabled:opacity-60"
          type="submit"
          :disabled="auth.loading || !canSubmit"
        >
          {{ auth.loading ? 'Đang tạo tài khoản...' : 'Đăng ký' }}
        </button>
      </div>
    </form>
    <section class="mt-2 grid grid-cols-2 gap-3">
      <a href="#" class="grid min-h-[84px] place-items-center gap-1 rounded-[18px] bg-white font-extrabold shadow-[0_8px_20px_rgba(255,109,102,0.05)]">
        <span class="material-symbols-outlined text-primary">support_agent</span>
        <span>CSKH</span>
      </a>
      <a href="#" class="grid min-h-[84px] place-items-center gap-1 rounded-[18px] bg-white font-extrabold shadow-[0_8px_20px_rgba(255,109,102,0.05)]">
        <span class="material-symbols-outlined text-primary">security</span>
        <span>Bảo mật</span>
      </a>
    </section>
  </div>
</template>

