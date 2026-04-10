// Normalize VN phone number to E.164 `+84...` format for server/DB matching.
export function normalizeVNPhone(input: string): string {
  const raw = String(input ?? '').trim().replace(/\s+/g, '')

  if (!raw) return ''

  if (raw.startsWith('+')) {
    return raw
      .replace(/[^\d+]/g, '')
      .replace(/^(\+84)0+/, '$1')
  }

  const digits = raw.replace(/[^\d]/g, '')
  if (!digits) return ''

  if (digits.startsWith('84')) {
    return `+${digits}`
  }

  if (digits.startsWith('0')) {
    return `+84${digits.slice(1)}`
  }

  return `+84${digits}`
}

