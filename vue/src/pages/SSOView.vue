<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { request } from '@/shared/api/http'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const error = ref('')
const loading = ref(true)

async function exchangeToken() {
  const ssoToken = route.query.token as string | undefined
  if (!ssoToken) {
    error.value = 'Hệ thống không tìm thấy mã xác thực SSO'
    loading.value = false
    return
  }

  try {
    const response = await request<any>('POST', '/v1/auth/sso/exchange', {
      body: { token: ssoToken }
    })

    // Save auth data
    auth.applyAuthResponse(response)
    
    // Redirect to admin control
    void router.replace({ name: 'admin-control' })
  } catch (err: any) {
    error.value = err.message || 'Xác thực SSO thất bại'
    loading.value = false
  }
}

onMounted(() => {
  void exchangeToken()
})
</script>

<template>
  <div class="sso-loading min-h-screen bg-[#0f172a] flex flex-col items-center justify-center p-4">
    <div v-if="loading" class="text-center">
      <!-- Premium Loading Animation -->
      <div class="w-16 h-16 border-4 border-amber-500/20 border-t-amber-500 rounded-full animate-spin mx-auto mb-6"></div>
      <h2 class="text-2xl font-bold text-white mb-2">Đang xác thực bảo mật...</h2>
      <p class="text-slate-400">Vui lòng không đóng cửa sổ này</p>
    </div>

    <div v-else-if="error" class="max-w-md w-full bg-red-500/10 border border-red-500/20 p-8 rounded-2xl text-center">
      <div class="w-12 h-12 bg-red-500 rounded-full flex items-center justify-center mx-auto mb-4 text-white">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="w-6 h-6">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </div>
      <h2 class="text-xl font-bold text-white mb-2">Lỗi xác thực</h2>
      <p class="text-red-400 mb-6">{{ error }}</p>
      <router-link to="/auth" class="inline-block bg-slate-800 hover:bg-slate-700 text-white font-medium px-6 py-2 rounded-lg transition-colors">
        Quay lại Đăng nhập
      </router-link>
    </div>
  </div>
</template>

<style scoped>
.sso-loading {
  font-family: 'Inter', sans-serif;
}
</style>
