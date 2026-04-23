import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { request, type ApiError, setSessionInvalidatedCallback } from '@/shared/api/http'
import type {
  AffiliateProfile,
  AuthResponse,
  AuthUser,
  ForgotPasswordRequest,
  LoginRequest,
  RegisterRequest,
  ChangePasswordRequest,
  ResetPasswordRequest,
  VerifyForgotOtpRequest,
  VerifyForgotOtpResponse,
} from '@/shared/api/types'
import { readJSON, remove, writeJSON } from '@/shared/lib/storage'

const STORAGE_KEY = 'ff789:auth:v1'

type PersistedAuth = {
  accessToken: string
  expiresAt: number
  refreshToken: string
  refreshExpiresAt: number
  user: AuthUser | null
  affiliateProfile: AffiliateProfile | null
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string>('')
  const expiresAt = ref<number>(0)
  const refreshToken = ref<string>('')
  const refreshExpiresAt = ref<number>(0)
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
      refreshToken: refreshToken.value,
      refreshExpiresAt: refreshExpiresAt.value,
      user: user.value,
      affiliateProfile: affiliateProfile.value,
    }
    writeJSON(STORAGE_KEY, payload)
  }

  function clear() {
    accessToken.value = ''
    expiresAt.value = 0
    refreshToken.value = ''
    refreshExpiresAt.value = 0
    user.value = null
    affiliateProfile.value = null
    error.value = ''
    loading.value = false
    remove(STORAGE_KEY)
  }

  function hydrate() {
    const saved = readJSON<PersistedAuth>(STORAGE_KEY)
    if (!saved) return
    // Check if even refresh token is expired
    if (saved.refreshExpiresAt && Date.now() >= saved.refreshExpiresAt) {
      clear()
      return
    }
    accessToken.value = saved.accessToken ?? ''
    expiresAt.value = saved.expiresAt ?? 0
    refreshToken.value = saved.refreshToken ?? ''
    refreshExpiresAt.value = saved.refreshExpiresAt ?? 0
    user.value = saved.user ?? null
    affiliateProfile.value = saved.affiliateProfile ?? null
  }

  function applyAuthResponse(res: AuthResponse) {
    accessToken.value = res.access_token
    expiresAt.value = Date.now() + Number(res.expires_in ?? 0) * 1000
    refreshToken.value = res.refresh_token ?? ''
    refreshExpiresAt.value = res.refresh_expires_in ? (Date.now() + Number(res.refresh_expires_in) * 1000) : 0
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
        // Try refresh if possible
        if (refreshToken.value) {
           try {
              await refresh()
              return await fetchMe()
           } catch {
              clear()
              return null
           }
        }
        clear()
        return null
      }
      throw e
    }
  }

  let refreshPromise: Promise<AuthResponse> | null = null
  async function refresh() {
    if (refreshPromise) return refreshPromise
    if (!refreshToken.value) throw new Error('No refresh token')

    refreshPromise = (async () => {
      try {
        const res = await request<AuthResponse>('POST', '/v1/auth/refresh', {
          body: { refresh_token: refreshToken.value },
        })
        applyAuthResponse(res)
        return res
      } catch (e) {
        clear()
        throw e
      } finally {
        refreshPromise = null
      }
    })()

    return refreshPromise
  }

  function logout() {
    clear()
  }

  function forcedLogout(reason: string) {
    clear()
    error.value = reason
    try {
      window.sessionStorage.setItem('ff789:forced-logout-reason', reason)
    } catch {
      // no-op
    }
    // Full reload để clear state memory và hiện thông báo ở màn login.
    window.location.href = `/login?session_invalidated=1`
  }

  // Hook up global session invalidate
  setSessionInvalidatedCallback(() => {
    forcedLogout('Tài khoản của bạn đã được đăng nhập từ một thiết bị khác. Vui lòng đăng nhập lại.')
  })


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

  async function changePassword(payload: ChangePasswordRequest) {
    if (!accessToken.value) {
      throw new Error('Bạn chưa đăng nhập')
    }

    loading.value = true
    error.value = ''
    try {
      return await request<{ message: string }>('POST', '/v1/auth/change-password', {
        body: payload,
        token: accessToken.value,
      })
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
    refreshToken,
    refreshExpiresAt,
    user,
    affiliateProfile,
    loading,
    error,
    isAuthenticated,
    hydrate,
    applyAuthResponse,
    login,
    register,
    refresh,
    fetchMe,
    logout,
    forgotPassword,
    verifyForgotOtp,
    resetPassword,
    changePassword,
  }
})
