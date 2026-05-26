export function readJSON<T>(key: string): T | null {
  try {
    const raw = window.localStorage.getItem(key)
    if (!raw) return null
    return JSON.parse(raw) as T
  } catch {
    return null
  }
}

export function writeJSON(key: string, value: unknown) {
  window.localStorage.setItem(key, JSON.stringify(value))
}

export function remove(key: string) {
  window.localStorage.removeItem(key)
}
