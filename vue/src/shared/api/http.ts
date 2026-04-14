import { env } from '@/shared/config/env'

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

export type ApiError = {
  status: number
  message: string
  code?: string
}

// Global callback for session invalidation (set by auth store)
export let onSessionInvalidated: (() => void) | null = null
export function setSessionInvalidatedCallback(fn: () => void) {
  onSessionInvalidated = fn
}

type RequestOptions = {
  body?: unknown
  headers?: Record<string, string>
  token?: string | null
  timeoutMs?: number
}

function joinUrl(base: string, path: string): string {
  if (!base) return path
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  if (!path.startsWith('/')) path = `/${path}`
  return `${base}${path}`
}

async function readErrorBody(res: Response): Promise<{ message: string; code?: string }> {
  try {
    const data = (await res.json()) as any
    const message = (data && typeof data.message === 'string' && data.message.trim())
      ? data.message
      : (res.statusText || 'Lỗi không xác định')
    const code = (data && typeof data.code === 'string') ? data.code : undefined
    return { message, code }
  } catch {
    return { message: res.statusText || 'Lỗi không xác định' }
  }
}

export async function request<T>(
  method: HttpMethod,
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const controller = new AbortController()
  const timeoutMs = options.timeoutMs ?? 15000
  const timeoutId = window.setTimeout(() => controller.abort(), timeoutMs)

  try {
    const headers: Record<string, string> = {
      Accept: 'application/json',
      ...options.headers,
    }

    if (options.token) headers.Authorization = `Bearer ${options.token}`
    const hasBody = options.body !== undefined && method !== 'GET'
    if (hasBody) headers['Content-Type'] = 'application/json'

    const url = joinUrl(env.apiBaseUrl, path)
    const res = await fetch(url, {
      method,
      headers,
      body: hasBody ? JSON.stringify(options.body) : undefined,
      signal: controller.signal,
    })

    if (!res.ok) {
      const { message, code } = await readErrorBody(res)
      const err: ApiError = { status: res.status, message, code }
      // Handle session invalidation globally
      if (res.status === 401 && code === 'SESSION_INVALIDATED') {
        onSessionInvalidated?.()
      }
      throw err
    }

    if (res.status === 204) return undefined as T
    return (await res.json()) as T
  } finally {
    window.clearTimeout(timeoutId)
  }
}

