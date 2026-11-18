interface DateRangeFilterProps {
  startDate: string;
  endDate: string;
  onChange: (range: { startDate: string; endDate: string }) => void;
}

export function DateRangeFilter({ startDate, endDate, onChange }: DateRangeFilterProps) {
  return (
    <fieldset className="date-range">
      <legend className="visually-hidden">Periode</legend>
      <label>
        Dari
        <input
          type="date"
          value={startDate}
          max={endDate}
          onChange={(e) => onChange({ startDate: e.target.value, endDate })}
        />
      </label>
      <label>
        Sampai
        <input
          type="date"
          value={endDate}
          min={startDate}
          onChange={(e) => onChange({ startDate, endDate: e.target.value })}
        />
      </label>
    </fieldset>
  );
}
