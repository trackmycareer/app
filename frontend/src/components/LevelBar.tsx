interface LevelBarProps {
  level: number;
  levelTitle: string;
  totalPoints: number;
  nextLevelAt: number;
}

export function LevelBar({ level, levelTitle, totalPoints, nextLevelAt }: LevelBarProps) {
  const progress = nextLevelAt > 0 ? Math.min((totalPoints / nextLevelAt) * 100, 100) : 100;

  return (
    <div
      className="rounded-[var(--radius-xl)] border border-[var(--border-subtle)]
        bg-[var(--bg-surface)] p-4"
    >
      <div className="mb-2 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span
            className="inline-flex h-7 w-7 items-center justify-center rounded-full
              bg-[var(--accent-default)] text-xs font-bold text-white"
          >
            {level}
          </span>
          <span className="text-sm font-semibold text-[var(--text-primary)]">
            {levelTitle}
          </span>
        </div>
        <span className="text-xs text-[var(--text-tertiary)]">
          {totalPoints} / {nextLevelAt} pts
        </span>
      </div>
      <div
        className="h-2 overflow-hidden rounded-full bg-[var(--bg-elevated)]"
        role="progressbar"
        aria-valuenow={totalPoints}
        aria-valuemin={0}
        aria-valuemax={nextLevelAt}
        aria-label={`Level ${level} progress`}
      >
        <div
          className="h-full rounded-full bg-[var(--accent-default)] transition-all duration-500"
          style={{ width: `${progress}%` }}
        />
      </div>
    </div>
  );
}
