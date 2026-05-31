export function ProgressBar({ value }: { value: number }) {
  const clamped = Math.max(0, Math.min(100, value));
  return (
    <div
      className="h-1.5 w-24 overflow-hidden rounded-full bg-[var(--border-subtle)]"
      role="progressbar"
      aria-valuenow={clamped}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-label={`Study progress ${clamped}%`}
    >
      <div
        className="h-full rounded-full bg-[var(--accent-default)] transition-all"
        style={{ width: `${clamped}%` }}
      />
    </div>
  );
}
