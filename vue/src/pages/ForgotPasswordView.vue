<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/auth'
import { normalizeVNPhone } from '@/shared/lib/phone'

const router = useRouter()
const auth = useAuthStore()

const channel = ref<'email' | 'phone'>('phone')
const account = ref('')
const otp = ref('')
const resetToken = ref('')
const newPassword = ref('')
const step = ref<1 | 2 | 3>(1)
const submitError = ref('')
const submitSuccess = ref('')

const title = computed(() => {
  if (step.value === 1) return 'Quên mật khẩu'
  if (step.value === 2) return 'Xác minh OTP'
  return 'Đặt lại mật khẩu'
})

async function sendOtp() {
  submitError.value = ''
  submitSuccess.value = ''
  try {
    const normalizedAccount = channel.value === 'phone' ? normalizeVNPhone(account.value) : account.value.trim()
    const res = await auth.forgotPassword({ channel: channel.value, account: normalizedAccount })
    submitSuccess.value = res.message
    step.value = 2
  } catch (e: any) {
    submitError.value = auth.error || 'Không gửi được OTP'
  }
}

async function verifyOtp() {
  submitError.value = ''
  submitSuccess.value = ''
  try {
    const normalizedAccount = channel.value === 'phone' ? normalizeVNPhone(account.value) : account.value.trim()
    const res = await auth.verifyForgotOtp({ channel: channel.value, account: normalizedAccount, otp: otp.value.trim() })
    resetToken.value = res.reset_token
    step.value = 3
  } catch (e: any) {
    submitError.value = auth.error || 'OTP không hợp lệ'
  }
}

async function resetPassword() {
  submitError.value = ''
  submitSuccess.value = ''
  try {
    const res = await auth.resetPassword({ reset_token: resetToken.value, new_password: newPassword.value })
    submitSuccess.value = res.message
    await router.replace('/auth')
  } catch (e: any) {
    submitError.value = auth.error || 'Không thể đổi mật khẩu'
  }
}
</script>

<template>
  <div class="space-y-5">
    <header class="grid min-h-12 grid-cols-[36px_1fr_60px] items-center md:min-h-[52px]">
      <button class="grid h-9 w-9 place-items-center text-primary transition-transform active:scale-95" type="button" @click="router.back()">
        <span class="material-symbols-outlined">arrow_back</span>
      </button>
      <h1 class="m-0 text-center text-[1.1rem] font-black text-primary md:text-[1.2rem]">{{ title }}</h1>
      <RouterLink class="justify-self-end text-right text-sm font-extrabold text-primary" to="/auth">Đăng nhập</RouterLink>
    </header>

    <section class="rounded-[22px] bg-white p-5 shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
      <p class="m-0 text-sm text-on-surface-variant">Chúng ta sẽ gửi OTP qua email hoặc số điện thoại, sau đó đặt lại mật khẩu mới.</p>
    </section>

    <section v-if="submitError" class="rounded-2xl bg-secondary/10 p-4 text-sm font-bold text-on-secondary-container">
      {{ submitError }}
    </section>

    <section v-if="submitSuccess" class="rounded-2xl bg-primary/10 p-4 text-sm font-bold text-primary">
      {{ submitSuccess }}
    </section>

    <form v-if="step === 1" class="space-y-3" @submit.prevent="sendOtp">
      <section class="grid grid-cols-2 gap-2 rounded-[18px] bg-surface-container p-1.5">
        <button
          class="min-h-11 rounded-[14px] font-extrabold transition-all"
          :class="channel === 'phone' ? 'bg-white text-primary shadow-[0_4px_12px_rgba(255,109,102,0.1)]' : 'text-on-surface-variant'"
          type="button"
          @click="channel = 'phone'"
        >
          Số điện thoại
        </button>
        <button
          class="min-h-11 rounded-[14px] font-extrabold transition-all"
          :class="channel === 'email' ? 'bg-white text-primary shadow-[0_4px_12px_rgba(255,109,102,0.1)]' : 'text-on-surface-variant'"
          type="button"
          @click="channel = 'email'"
        >
          Email
        </button>
      </section>

      <label class="grid min-h-[58px] items-center overflow-hidden rounded-[18px] bg-white shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
        <input v-model="account" class="min-w-0 border-0 bg-transparent px-4 py-4 outline-none" :type="channel === 'email' ? 'email' : 'tel'" :placeholder="channel === 'email' ? 'Nhập email' : 'Nhập số điện thoại'" />
      </label>

      <button
        class="min-h-14 rounded-[18px] bg-gradient-to-br from-primary to-primary-container font-black text-white shadow-[0_8px_20px_rgba(255,109,102,0.18)] transition-transform active:scale-95 disabled:opacity-60"
        type="submit"
        :disabled="auth.loading || !account.trim()"
      >
        {{ auth.loading ? 'Đang gửi OTP...' : 'Gửi OTP' }}
      </button>
    </form>

    <form v-else-if="step === 2" class="space-y-3" @submit.prevent="verifyOtp">
      <label class="grid min-h-[58px] items-center overflow-hidden rounded-[18px] bg-white shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
        <input v-model="otp" class="min-w-0 border-0 bg-transparent px-4 py-4 outline-none tracking-[0.4em]" inputmode="numeric" maxlength="6" placeholder="Nhập OTP" />
      </label>

      <button class="min-h-14 rounded-[18px] bg-gradient-to-br from-primary to-primary-container font-black text-white disabled:opacity-60" type="submit" :disabled="auth.loading || otp.length < 4">
        {{ auth.loading ? 'Đang xác minh...' : 'Xác minh OTP' }}
      </button>
    </form>

    <form v-else class="space-y-3" @submit.prevent="resetPassword">
      <label class="grid min-h-[58px] items-center overflow-hidden rounded-[18px] bg-white shadow-[0_8px_20px_rgba(255,109,102,0.06)]">
        <input v-model="newPassword" class="min-w-0 border-0 bg-transparent px-4 py-4 outline-none" type="password" autocomplete="new-password" placeholder="Mật khẩu mới" />
      </label>

      <button class="min-h-14 rounded-[18px] bg-gradient-to-br from-primary to-primary-container font-black text-white disabled:opacity-60" type="submit" :disabled="auth.loading || newPassword.length < 6">
        {{ auth.loading ? 'Đang cập nhật...' : 'Đặt lại mật khẩu' }}
      </button>
    </form>
  </div>
</template>
