function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '')
}

export const env = {
  apiBaseUrl: trimTrailingSlash(import.meta.env.VITE_API_BASE_URL ?? ''),
  mainSiteUrl: trimTrailingSlash(import.meta.env.VITE_MAIN_SITE_URL ?? ''),
  chatGlobalEnabled: String(import.meta.env.VITE_CHAT_GLOBAL_ENABLED ?? 'false').toLowerCase() === 'true',
  chatRoomCode: String(import.meta.env.VITE_CHAT_ROOM_CODE ?? 'global').trim() || 'global',
  wheelEventEnabled: String(import.meta.env.VITE_WHEEL_EVENT_ENABLED ?? 'false').toLowerCase() === 'true',
}
