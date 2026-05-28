interface StreakCounterProps {
  currentStreak: number;
  longestStreak: number;
}

export function StreakCounter({ currentStreak, longestStreak }: StreakCounterProps) {
  const isActive = currentStreak > 0;

  return (
    <div
      className={[
        "flex flex-col items-center justify-center rounded-[var(--radius-xl)] border p-5",
        "bg-[var(--bg-surface)]",
        isActive
          ? "border-[var(--accent-default)] shadow-[0_0_12px_-3px_var(--accent-default)]"
          : "border-[var(--border-subtle)]",
      ].join(" ")}
      role="status"
      aria-label={`Current streak: ${currentStreak} days`}
    >
      <div className="flex items-baseline gap-1.5">
        <span className="text-4xl font-bold text-[var(--text-primary)]">
          {currentStreak}
        </span>
        <span className="text-2xl" aria-hidden="true">
          {isActive ? "🔥" : ""}
        </span>
      </div>
      <p className="mt-0.5 text-sm font-medium text-[var(--text-secondary)]">
        day streak
      </p>
      <p className="mt-2 text-xs text-[var(--text-tertiary)]">
        Longest: {longestStreak} day{longestStreak !== 1 ? "s" : ""}
      </p>
    </div>
  );
}
