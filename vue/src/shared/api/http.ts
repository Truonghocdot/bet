import { env } from '@/shared/config/env'

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

export type ApiError = {
  status: number
  message: string
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

async function readErrorMessage(res: Response): Promise<string> {
  try {
    const data = (await res.json()) as any
    if (data && typeof data.message === 'string' && data.message.trim()) return data.message
  } catch {
    // ignore
  }
  return res.statusText || 'Lỗi không xác định'
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
      const message = await readErrorMessage(res)
      const err: ApiError = { status: res.status, message }
      throw err
    }

    if (res.status === 204) return undefined as T
    return (await res.json()) as T
  } finally {
    window.clearTimeout(timeoutId)
  }
}

