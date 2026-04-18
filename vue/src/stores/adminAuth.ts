import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { request, type ApiError } from '@/shared/api/http'
import type { AffiliateProfile, AuthResponse, AuthUser } from '@/shared/api/types'

const ADMIN_STORAGE_KEY = 'ff789:admin-auth:v1'

type PersistedAdminAuth = {
  accessToken: string
  expiresAt: number
  refreshToken: string
  refreshExpiresAt: number
  user: AuthUser | null
  affiliateProfile: AffiliateProfile | null
}

function readAdminJSON<T>(key: string): T | null {
  try {
    const raw = window.sessionStorage.getItem(key)
    if (!raw) return null
    return JSON.parse(raw) as T
  } catch {
    return null
  }
}

function writeAdminJSON(key: string, value: unknown): void {
  window.sessionStorage.setItem(key, JSON.stringify(value))
}

function removeAdminJSON(key: string): void {
  window.sessionStorage.removeItem(key)
}

export const useAdminAuthStore = defineStore('admin-auth', () => {
  const accessToken = ref<string>('')
  const expiresAt = ref<number>(0)
  const refreshToken = ref<string>('')
  const refreshExpiresAt = ref<number>(0)
  const user = ref<AuthUser | null>(null)
  const affiliateProfile = ref<AffiliateProfile | null>(null)

  const loading = ref(false)
  const error = ref('')

  const isAuthenticated = computed(() => {
    if (!accessToken.value) return false
    if (!expiresAt.value) return true
    return Date.now() < expiresAt.value
  })

  const isAdmin = computed(() => user.value?.role === 0 || user.value?.role === 1)

  function persist() {
    const payload: PersistedAdminAuth = {
      accessToken: accessToken.value,
      expiresAt: expiresAt.value,
      refreshToken: refreshToken.value,
      refreshExpiresAt: refreshExpiresAt.value,
      user: user.value,
      affiliateProfile: affiliateProfile.value,
    }
    writeAdminJSON(ADMIN_STORAGE_KEY, payload)
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
    removeAdminJSON(ADMIN_STORAGE_KEY)
  }

  function hydrate() {
    const saved = readAdminJSON<PersistedAdminAuth>(ADMIN_STORAGE_KEY)
    if (!saved) return
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
    isAdmin,
    hydrate,
    applyAuthResponse,
    fetchMe,
    logout,
    clear,
  }
})
