import type { WinPreview } from "@/types";
import { CloseIcon } from "@/components/icons";

interface WinCardProps {
  win: WinPreview;
  index: number;
  onRemove: () => void;
}

export function WinCard({ win, index, onRemove }: WinCardProps) {
  return (
    <div
      className="animate-fade-in-up group relative flex items-start gap-3 rounded-[var(--radius-lg)]
        border border-[var(--border-subtle)] bg-[var(--bg-elevated)] p-4 transition-all duration-200
        hover:-translate-y-0.5 hover:shadow-lg hover:shadow-stone-900/10"
      style={{ animationDelay: `${index * 50}ms` }}
    >
      <div className="min-w-0 flex-1">
        <h4 className="font-semibold text-[var(--text-primary)]">{win.title}</h4>
        {win.description && (
          <p className="mt-1 line-clamp-2 text-sm text-[var(--text-secondary)]">
            {win.description}
          </p>
        )}
        <div className="mt-2 flex flex-wrap gap-2">
          {win.category && (
            <span
              className="inline-flex items-center rounded-full bg-[var(--bg-surface)] px-2 py-0.5
                text-xs font-medium text-[var(--text-secondary)]"
            >
              {win.category}
            </span>
          )}
          {win.occurred_on && (
            <span className="text-xs text-[var(--text-tertiary)]">
              {new Date(win.occurred_on).toLocaleDateString("en-GB", {
                day: "numeric",
                month: "short",
                year: "numeric",
              })}
            </span>
          )}
        </div>
      </div>
      <button
        type="button"
        onClick={onRemove}
        className="flex-shrink-0 rounded-[var(--radius-sm)] p-1 text-[var(--text-tertiary)]
          opacity-0 transition-all hover:bg-[var(--bg-hover)] hover:text-[var(--color-error)]
          group-hover:opacity-100 group-focus-within:opacity-100"
        aria-label={`Remove ${win.title}`}
      >
        <CloseIcon width={16} height={16} />
      </button>
    </div>
  );
}
