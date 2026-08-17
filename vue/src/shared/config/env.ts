function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '')
}

export const env = {
  apiBaseUrl: trimTrailingSlash(import.meta.env.VITE_API_BASE_URL ?? ''),
  chatGlobalEnabled: String(import.meta.env.VITE_CHAT_GLOBAL_ENABLED ?? 'false').toLowerCase() === 'true',
  chatRoomCode: String(import.meta.env.VITE_CHAT_ROOM_CODE ?? 'global').trim() || 'global',
}
