export interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  subtext?: string;
  loading?: boolean;
  borderColour?: string;
}

export function StatCard({
  icon,
  label,
  value,
  subtext,
  loading,
  borderColour,
}: StatCardProps) {
  return (
    <div
      className={[
        "rounded-[var(--radius-xl)] border border-[var(--border-subtle)]",
        "bg-[var(--bg-surface)] p-4",
        borderColour ?? "",
      ].join(" ")}
    >
      <div className="mb-3 flex items-center gap-2 text-[var(--text-tertiary)]">
        {icon}
        <span className="text-xs font-medium uppercase tracking-wider">{label}</span>
      </div>
      {loading ? (
        <div className="h-7 w-16 animate-pulse rounded bg-[var(--bg-elevated)] motion-reduce:animate-none" />
      ) : (
        <>
          <p className="text-2xl font-bold text-[var(--text-primary)]">{value}</p>
          {subtext && (
            <p className="mt-0.5 truncate text-xs text-[var(--text-tertiary)]">{subtext}</p>
          )}
        </>
      )}
    </div>
  );
}
