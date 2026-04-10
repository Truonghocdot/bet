<script setup lang="ts">
import { ref } from 'vue'

const mode = ref<'phone' | 'email'>('phone')
const showPassword = ref(false)
</script>

<template>
  <div class="space-y-5">
    <header class="grid min-h-12 grid-cols-[36px_1fr_60px] items-center md:min-h-[52px]">
      <button class="grid h-9 w-9 place-items-center text-primary transition-transform active:scale-95">
        <span class="material-symbols-outlined">arrow_back</span>
      </button>
      <h1 class="m-0 text-center text-[1.1rem] font-black text-primary md:text-[1.2rem]">Đăng nhập</h1>
      <button class="justify-self-end text-right text-sm font-extrabold text-primary">Hỗ trợ</button>
    </header>

    <section class="text-center">
      <div class="mx-auto mb-4 grid h-[72px] w-[72px] place-items-center rounded-[20px] bg-white shadow-[0_10px_24px_rgba(0,78,219,0.12)]">
        <span class="text-[1.1rem] font-black italic tracking-[-0.06em] text-primary">ff789</span>
      </div>
      <h2 class="m-0 text-[1.55rem] font-black">Chào mừng trở lại</h2>
      <p class="mt-1.5 text-sm text-on-surface-variant">Vui lòng đăng nhập để tiếp tục</p>
    </section>

    <section class="grid grid-cols-2 gap-2 rounded-[18px] bg-surface-container p-1.5">
      <button
        class="min-h-11 rounded-[14px] font-extrabold transition-all"
        :class="mode === 'phone' ? 'bg-white text-primary shadow-[0_4px_12px_rgba(0,78,219,0.1)]' : 'text-on-surface-variant'"
        @click="mode = 'phone'"
      >
        Số điện thoại
      </button>
      <button
        class="min-h-11 rounded-[14px] font-extrabold transition-all"
        :class="mode === 'email' ? 'bg-white text-primary shadow-[0_4px_12px_rgba(0,78,219,0.1)]' : 'text-on-surface-variant'"
        @click="mode = 'email'"
      >
        Đăng nhập email
      </button>
    </section>

    <section class="space-y-3">
      <label class="grid min-h-[58px] grid-cols-[auto_1fr_auto] items-center overflow-hidden rounded-[18px] bg-white shadow-[0_8px_20px_rgba(0,78,219,0.06)]">
        <template v-if="mode === 'phone'">
          <span class="border-r border-slate-200/90 px-4 font-extrabold">+84</span>
          <input class="min-w-0 border-0 bg-transparent px-4 outline-none" type="tel" placeholder="Nhập số điện thoại" />
        </template>
        <template v-else>
          <span class="border-r border-slate-200/90 px-4 font-extrabold text-primary">@</span>
          <input class="min-w-0 border-0 bg-transparent px-4 outline-none" type="email" placeholder="Nhập email" />
        </template>
      </label>

      <label class="grid min-h-[58px] grid-cols-[auto_1fr_auto] items-center overflow-hidden rounded-[18px] bg-white shadow-[0_8px_20px_rgba(0,78,219,0.06)]">
        <span class="grid h-full w-[52px] place-items-center text-primary">
          <span class="material-symbols-outlined">lock</span>
        </span>
        <input
          class="min-w-0 border-0 bg-transparent px-4 outline-none"
          :type="showPassword ? 'text' : 'password'"
          placeholder="Mật khẩu"
        />
        <button type="button" class="grid h-full w-[52px] place-items-center text-on-surface-variant" @click="showPassword = !showPassword">
          <span class="material-symbols-outlined">{{ showPassword ? 'visibility_off' : 'visibility' }}</span>
        </button>
      </label>

      <div class="flex items-center justify-between gap-4 text-[0.78rem] font-extrabold">
        <label class="inline-flex items-center gap-2 text-on-surface-variant">
          <input type="checkbox" />
          <span>Nhớ mật khẩu</span>
        </label>
        <a href="#" class="text-primary">Quên mật khẩu?</a>
      </div>
    </section>

    <section class="mt-2 grid gap-3">
      <button class="min-h-14 rounded-[18px] bg-gradient-to-br from-primary to-primary-container font-black text-white shadow-[0_8px_20px_rgba(0,78,219,0.18)] transition-transform active:scale-95">
        Đăng nhập
      </button>
      <button class="min-h-14 rounded-[18px] bg-white font-black text-on-surface shadow-[0_8px_20px_rgba(0,78,219,0.06)] transition-transform active:scale-95">
        Đăng ký
      </button>
    </section>

    <section class="mt-2 grid grid-cols-2 gap-3">
      <a href="#" class="grid min-h-[84px] place-items-center gap-1 rounded-[18px] bg-white font-extrabold shadow-[0_8px_20px_rgba(0,78,219,0.05)]">
        <span class="material-symbols-outlined text-primary">support_agent</span>
        <span>CSKH</span>
      </a>
      <a href="#" class="grid min-h-[84px] place-items-center gap-1 rounded-[18px] bg-white font-extrabold shadow-[0_8px_20px_rgba(0,78,219,0.05)]">
        <span class="material-symbols-outlined text-primary">security</span>
        <span>Bảo mật</span>
      </a>
    </section>
  </div>
</template>
