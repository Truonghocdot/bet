export function formatViMoney(value: string | number | null | undefined, fractionDigits = 0): string {
  const numeric = typeof value === 'number' ? value : Number.parseFloat(String(value ?? '0'))

  if (!Number.isFinite(numeric)) {
    return String(value ?? '0')
  }

  return new Intl.NumberFormat('vi-VN', {
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  }).format(numeric)
}
