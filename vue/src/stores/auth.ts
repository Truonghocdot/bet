import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { request, type ApiError } from '@/shared/api/http'
import type {
  AffiliateProfile,
  AuthResponse,
  AuthUser,
  ForgotPasswordRequest,
  LoginRequest,
  RegisterRequest,
  ResetPasswordRequest,
  VerifyForgotOtpRequest,
  VerifyForgotOtpResponse,
} from '@/shared/api/types'
import { readJSON, remove, writeJSON } from '@/shared/lib/storage'

const STORAGE_KEY = 'ff789:auth:v1'

type PersistedAuth = {
  accessToken: string
  expiresAt: number
  user: AuthUser | null
  affiliateProfile: AffiliateProfile | null
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string>('')
  const expiresAt = ref<number>(0)
  const user = ref<AuthUser | null>(null)
  const affiliateProfile = ref<AffiliateProfile | null>(null)

  const loading = ref(false)
  const error = ref<string>('')

  const isAuthenticated = computed(() => {
    if (!accessToken.value) return false
    if (!expiresAt.value) return true
    return Date.now() < expiresAt.value
  })

  function persist() {
    const payload: PersistedAuth = {
      accessToken: accessToken.value,
      expiresAt: expiresAt.value,
      user: user.value,
      affiliateProfile: affiliateProfile.value,
    }
    writeJSON(STORAGE_KEY, payload)
  }

  function clear() {
    accessToken.value = ''
    expiresAt.value = 0
    user.value = null
    affiliateProfile.value = null
    error.value = ''
    loading.value = false
    remove(STORAGE_KEY)
  }

  function hydrate() {
    const saved = readJSON<PersistedAuth>(STORAGE_KEY)
    if (!saved) return
    if (saved.expiresAt && Date.now() >= saved.expiresAt) {
      clear()
      return
    }
    accessToken.value = saved.accessToken ?? ''
    expiresAt.value = saved.expiresAt ?? 0
    user.value = saved.user ?? null
    affiliateProfile.value = saved.affiliateProfile ?? null
  }

  function applyAuthResponse(res: AuthResponse) {
    accessToken.value = res.access_token
    expiresAt.value = Date.now() + Number(res.expires_in ?? 0) * 1000
    user.value = res.user
    affiliateProfile.value = res.affiliate_profile ?? null
    persist()
  }

  async function login(payload: LoginRequest) {
    loading.value = true
    error.value = ''
    try {
      const res = await request<AuthResponse>('POST', '/v1/auth/login', { body: payload })
      applyAuthResponse(res)
      return res
    } catch (e: any) {
      const err = e as ApiError
      error.value = err?.message ?? 'Đăng nhập thất bại'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function register(payload: RegisterRequest) {
    loading.value = true
    error.value = ''
    try {
      const res = await request<AuthResponse>('POST', '/v1/auth/register', { body: payload })
      applyAuthResponse(res)
      return res
    } catch (e: any) {
      const err = e as ApiError
      error.value = err?.message ?? 'Đăng ký thất bại'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchMe() {
    if (!accessToken.value) return null
    try {
      const res = await request<{ user: AuthUser; affiliate_profile?: AffiliateProfile | null }>('GET', '/v1/auth/me', {
        token: accessToken.value,
      })
      user.value = res.user
      affiliateProfile.value = res.affiliate_profile ?? null
      persist()
      return res
    } catch (e: any) {
      const err = e as ApiError
      if (err?.status === 401) {
        clear()
        return null
      }
      throw e
    }
  }

  function logout() {
    clear()
  }

  async function forgotPassword(payload: ForgotPasswordRequest) {
    loading.value = true
    error.value = ''
    try {
      return await request<{ message: string }>('POST', '/v1/auth/forgot-password', { body: payload })
    } catch (e: any) {
      const err = e as ApiError
      error.value = err?.message ?? 'Không thể gửi yêu cầu quên mật khẩu'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function verifyForgotOtp(payload: VerifyForgotOtpRequest) {
    loading.value = true
    error.value = ''
    try {
      return await request<VerifyForgotOtpResponse>('POST', '/v1/auth/forgot-password/verify-otp', { body: payload })
    } catch (e: any) {
      const err = e as ApiError
      error.value = err?.message ?? 'OTP không hợp lệ'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function resetPassword(payload: ResetPasswordRequest) {
    loading.value = true
    error.value = ''
    try {
      return await request<{ message: string }>('POST', '/v1/auth/reset-password', { body: payload })
    } catch (e: any) {
      const err = e as ApiError
      error.value = err?.message ?? 'Không thể đổi mật khẩu'
      throw e
    } finally {
      loading.value = false
    }
  }

  return {
    accessToken,
    expiresAt,
    user,
    affiliateProfile,
    loading,
    error,
    isAuthenticated,
    hydrate,
    applyAuthResponse,
    login,
    register,
    fetchMe,
    logout,
    forgotPassword,
    verifyForgotOtp,
    resetPassword,
  }
})
