import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { request, setSessionInvalidatedCallback, type ApiError } from '@/shared/api/http'
import type { AffiliateProfile, AuthResponse, AuthUser } from '@/shared/api/types'
import { readJSON, remove, writeJSON } from '@/shared/lib/storage'

const STORAGE_KEY = 'fh88u:agency:auth:v1'
const ROLE_AGENCY = 4

type PersistedAuth = {
  accessToken: string
  refreshToken: string
  expiresAt: number
  refreshExpiresAt: number
  user: AuthUser | null
  affiliateProfile: AffiliateProfile | null
}

export const useAgencyAuthStore = defineStore('agency-auth', () => {
  const accessToken = ref('')
  const refreshToken = ref('')
  const expiresAt = ref(0)
  const refreshExpiresAt = ref(0)
  const user = ref<AuthUser | null>(null)
  const affiliateProfile = ref<AffiliateProfile | null>(null)
  const loading = ref(false)
  const error = ref('')

  const isAuthenticated = computed(() => !!accessToken.value && Date.now() < expiresAt.value)
  const isAgency = computed(() => user.value?.role === ROLE_AGENCY)

  function persist() {
    writeJSON(STORAGE_KEY, {
      accessToken: accessToken.value,
      refreshToken: refreshToken.value,
      expiresAt: expiresAt.value,
      refreshExpiresAt: refreshExpiresAt.value,
      user: user.value,
      affiliateProfile: affiliateProfile.value,
    } satisfies PersistedAuth)
  }

  function clear() {
    accessToken.value = ''
    refreshToken.value = ''
    expiresAt.value = 0
    refreshExpiresAt.value = 0
    user.value = null
    affiliateProfile.value = null
    error.value = ''
    remove(STORAGE_KEY)
  }

  function hydrate() {
    const saved = readJSON<PersistedAuth>(STORAGE_KEY)
    if (!saved) return
    if (saved.refreshExpiresAt && Date.now() >= saved.refreshExpiresAt) {
      clear()
      return
    }
    accessToken.value = saved.accessToken ?? ''
    refreshToken.value = saved.refreshToken ?? ''
    expiresAt.value = saved.expiresAt ?? 0
    refreshExpiresAt.value = saved.refreshExpiresAt ?? 0
    user.value = saved.user ?? null
    affiliateProfile.value = saved.affiliateProfile ?? null
  }

  function applyAuthResponse(res: AuthResponse) {
    accessToken.value = res.access_token
    refreshToken.value = res.refresh_token ?? ''
    expiresAt.value = Date.now() + Number(res.expires_in ?? 0) * 1000
    refreshExpiresAt.value = res.refresh_expires_in ? Date.now() + Number(res.refresh_expires_in) * 1000 : 0
    user.value = res.user
    affiliateProfile.value = res.affiliate_profile ?? null
    persist()
  }

  async function login(account: string, password: string) {
    loading.value = true
    error.value = ''
    try {
      const res = await request<AuthResponse>('POST', '/v1/auth/login', {
        body: { account, password },
      })
      applyAuthResponse(res)
      if (res.user.role !== ROLE_AGENCY) {
        clear()
        throw { status: 403, message: 'Tài khoản này không phải agency.' } satisfies ApiError
      }
    } catch (e: any) {
      error.value = (e as ApiError)?.message ?? 'Đăng nhập thất bại'
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
        if (res.user.role !== ROLE_AGENCY) {
          clear()
          throw { status: 403, message: 'Tài khoản này không phải agency.' } satisfies ApiError
        }
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

  function forcedLogout(reason: string) {
    clear()
    error.value = reason
    try {
      window.sessionStorage.setItem('fh88u:agency:forced-logout-reason', reason)
    } catch {
      // no-op
    }
    window.location.href = '/login?session_invalidated=1'
  }

  setSessionInvalidatedCallback(() => {
    forcedLogout('Tài khoản agency đã được đăng nhập ở một thiết bị khác. Vui lòng đăng nhập lại.')
  })

  function logout() {
    clear()
  }

  return {
    accessToken,
    user,
    affiliateProfile,
    loading,
    error,
    isAuthenticated,
    isAgency,
    hydrate,
    login,
    fetchMe,
    refresh,
    logout,
  }
})
