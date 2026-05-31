import { useState } from "react";
import { ChevronDownIcon, ChevronRightIcon } from "@/components/icons";

export interface CollapsibleSectionProps {
  title: string;
  icon: React.ReactNode;
  count: number;
  defaultOpen?: boolean;
  children: React.ReactNode;
}

export function CollapsibleSection({
  title,
  icon,
  count,
  defaultOpen = true,
  children,
}: CollapsibleSectionProps) {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <div
      className="rounded-[var(--radius-xl)] border border-[var(--border-subtle)]
        bg-[var(--bg-surface)] overflow-hidden"
    >
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="flex w-full items-center gap-3 px-5 py-4 text-left transition-colors
          hover:bg-[var(--bg-hover)]"
        aria-expanded={open}
      >
        <span className="flex-shrink-0 text-[var(--accent-default)]">{icon}</span>
        <span className="flex-1 text-sm font-semibold text-[var(--text-primary)]">{title}</span>
        <span
          className="inline-flex h-6 min-w-6 items-center justify-center rounded-full
            bg-[var(--accent-default)]/15 px-2 text-xs font-semibold text-[var(--accent-default)]"
        >
          {count}
        </span>
        {open ? (
          <ChevronDownIcon width={16} height={16} className="text-[var(--text-tertiary)]" />
        ) : (
          <ChevronRightIcon width={16} height={16} className="text-[var(--text-tertiary)]" />
        )}
      </button>
      {open && <div className="border-t border-[var(--border-subtle)] px-5 py-3">{children}</div>}
    </div>
  );
}
