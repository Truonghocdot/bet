import { env } from '@/shared/config/env'

export type HttpMethod = 'GET' | 'POST'

export type ApiError = {
  status: number
  message: string
  code?: string
}

export let onSessionInvalidated: (() => void) | null = null
export function setSessionInvalidatedCallback(fn: () => void) {
  onSessionInvalidated = fn
}

type RequestOptions = {
  body?: unknown
  token?: string | null
}

function joinUrl(base: string, path: string): string {
  if (!base) return path
  if (!path.startsWith('/')) path = `/${path}`
  return `${base}${path}`
}

export async function request<T>(method: HttpMethod, path: string, options: RequestOptions = {}): Promise<T> {
  const headers: Record<string, string> = {
    Accept: 'application/json',
    'X-Client-Scope': 'agency',
  }

  if (options.token) headers.Authorization = `Bearer ${options.token}`
  if (options.body !== undefined && method !== 'GET') headers['Content-Type'] = 'application/json'

  const response = await fetch(joinUrl(env.apiBaseUrl, path), {
    method,
    headers,
    body: options.body !== undefined && method !== 'GET' ? JSON.stringify(options.body) : undefined,
  })

  if (!response.ok) {
    let message = response.statusText || 'Request failed'
    let code: string | undefined
    try {
      const payload = await response.json() as { message?: string; code?: string }
      if (payload.message) message = payload.message
      if (payload.code) code = payload.code
    } catch {
      // no-op
    }
    if (response.status === 401 && code === 'SESSION_INVALIDATED') {
      onSessionInvalidated?.()
    }
    throw { status: response.status, message, code } satisfies ApiError
  }

  return await response.json() as T
}
