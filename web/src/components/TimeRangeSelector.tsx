import { TIME_RANGE_LABELS } from '../utils/date'
import type { TimeRange } from '../utils/date'

interface TimeRangeSelectorProps {
  value: TimeRange
  onChange: (range: TimeRange) => void
}

const ranges: TimeRange[] = ['1w', '1m', '1y', '5y', 'max']

export default function TimeRangeSelector({ value, onChange }: TimeRangeSelectorProps) {
  return (
    <div className="inline-flex rounded-lg border border-[#e2e8f0] bg-white p-1 shadow-sm">
      {ranges.map((r) => (
        <button
          key={r}
          onClick={() => onChange(r)}
          className={`px-3 py-1.5 text-xs font-semibold rounded-md transition-colors ${
            value === r
              ? 'bg-[#003366] text-white'
              : 'text-[#475569] hover:bg-[#f1f5f9]'
          }`}
        >
          {TIME_RANGE_LABELS[r]}
        </button>
      ))}
    </div>
  )
}
