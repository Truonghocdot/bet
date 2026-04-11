import { env } from '@/shared/config/env'

export type StreamEventPayload = {
  event: string
  data: any
}

export type StreamConnection = {
  close: () => void
}

type StreamOptions = {
  token?: string | null
  onEvent: (payload: StreamEventPayload) => void
  onError?: (error: unknown) => void
  reconnectMs?: number
}

function joinUrl(base: string, path: string): string {
  if (!base) return path
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  if (!path.startsWith('/')) path = `/${path}`
  return `${base}${path}`
}

async function readErrorMessage(res: Response): Promise<string> {
  try {
    const data = await res.json() as { message?: string }
    if (data.message && data.message.trim()) return data.message
  } catch {
    // ignore
  }
  return res.statusText || 'Không thể kết nối realtime'
}

function parseEventBlock(block: string): StreamEventPayload | null {
  const lines = block.split('\n')
  let event = 'message'
  const dataLines: string[] = []

  for (const line of lines) {
    if (!line || line.startsWith(':')) continue
    if (line.startsWith('event:')) {
      event = line.slice(6).trim() || 'message'
      continue
    }
    if (line.startsWith('data:')) {
      dataLines.push(line.slice(5).trim())
    }
  }

  if (dataLines.length === 0) return null

  const raw = dataLines.join('\n')
  try {
    return { event, data: JSON.parse(raw) }
  } catch {
    return { event, data: raw }
  }
}

export function connectEventStream(path: string, options: StreamOptions): StreamConnection {
  let closed = false
  let controller: AbortController | null = null

  const reconnectMs = Math.max(1000, options.reconnectMs ?? 3000)

  const start = async () => {
    while (!closed) {
      controller = new AbortController()

      try {
        const headers: Record<string, string> = {
          Accept: 'text/event-stream',
        }
        if (options.token) headers.Authorization = `Bearer ${options.token}`

        const res = await fetch(joinUrl(env.apiBaseUrl, path), {
          method: 'GET',
          headers,
          signal: controller.signal,
        })

        if (!res.ok) {
          throw {
            status: res.status,
            message: await readErrorMessage(res),
          }
        }

        if (!res.body) {
          throw new Error('Realtime stream không có dữ liệu')
        }

        const reader = res.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ''

        while (!closed) {
          const chunk = await reader.read()
          if (chunk.done) break

          buffer += decoder.decode(chunk.value, { stream: true })
          const parts = buffer.split('\n\n')
          buffer = parts.pop() ?? ''

          for (const part of parts) {
            const event = parseEventBlock(part)
            if (event) {
              options.onEvent(event)
            }
          }
        }
      } catch (error) {
        if (!closed) {
          options.onError?.(error)
        }
      }

      if (!closed) {
        await new Promise((resolve) => window.setTimeout(resolve, reconnectMs))
      }
    }
  }

  void start()

  return {
    close() {
      closed = true
      controller?.abort()
    },
  }
}
