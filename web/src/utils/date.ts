export type TimeRange = '1w' | '1m' | '1y' | '5y' | 'max'

export interface RangeDates {
  since: string
  until: string
}

function formatDate(d: Date): string {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function getRangeDates(range: TimeRange, maxSince = '1900-01-01'): RangeDates {
  const until = new Date()
  const untilStr = formatDate(until)

  switch (range) {
    case '1w':
      return { since: formatDate(new Date(until.getTime() - 7 * 24 * 60 * 60 * 1000)), until: untilStr }
    case '1m':
      return { since: formatDate(new Date(until.getFullYear(), until.getMonth() - 1, until.getDate())), until: untilStr }
    case '1y':
      return { since: formatDate(new Date(until.getFullYear() - 1, until.getMonth(), until.getDate())), until: untilStr }
    case '5y':
      return { since: formatDate(new Date(until.getFullYear() - 5, until.getMonth(), until.getDate())), until: untilStr }
    case 'max':
    default:
      return { since: maxSince, until: untilStr }
  }
}

export const TIME_RANGE_LABELS: Record<TimeRange, string> = {
  '1w': '1S',
  '1m': '1M',
  '1y': '1A',
  '5y': '5A',
  max: 'MAX',
}
