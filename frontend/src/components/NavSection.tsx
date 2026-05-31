import type { ReactNode } from "react";

export function NavSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="space-y-1">
      <p className="px-3 pb-1 text-xs font-semibold uppercase tracking-wider text-[var(--text-tertiary)]">
        {title}
      </p>
      {children}
    </div>
  );
}
