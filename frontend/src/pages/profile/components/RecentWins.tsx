import type { Win } from "@/types";

interface RecentWinsProps {
  wins: Win[];
}

export function RecentWins({ wins }: RecentWinsProps) {
  return (
    <ul className="space-y-3">
      {wins.map((win) => (
        <li
          key={win.id}
          className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
            bg-[var(--bg-elevated)] p-4"
        >
          <p className="text-sm font-medium text-[var(--text-primary)]">{win.title}</p>
          {win.description && (
            <p className="mt-1 text-xs text-[var(--text-secondary)]">{win.description}</p>
          )}
          <time
            dateTime={win.occurred_on}
            className="mt-1 block text-xs text-[var(--text-tertiary)]"
          >
            {new Date(win.occurred_on).toLocaleDateString("en-GB", {
              day: "numeric",
              month: "short",
              year: "numeric",
            })}
          </time>
        </li>
      ))}
    </ul>
  );
}
