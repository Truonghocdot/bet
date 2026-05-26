export function formatViDateTime(value: string | number | Date): string {
  const date = value instanceof Date ? value : new Date(value)
  const parts = new Intl.DateTimeFormat('vi-VN', {
    timeZone: 'Asia/Ho_Chi_Minh',
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).formatToParts(date)

  const values = Object.fromEntries(parts.map((part) => [part.type, part.value]))

  return `${values.day}/${values.month}/${values.year} ${values.hour}:${values.minute}`
}
