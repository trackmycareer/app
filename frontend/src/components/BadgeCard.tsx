import type { Badge } from "@/types";

interface BadgeCardProps {
  badge: Badge;
}

const TIER_COLOURS: Record<string, string> = {
  gold: "var(--color-warning)",
  silver: "var(--border-strong)",
  bronze: "var(--accent-warm)",
};

export function BadgeCard({ badge }: BadgeCardProps) {
  const tierColour = TIER_COLOURS[badge.tier] ?? "var(--text-tertiary)";

  return (
    <div
      className={[
        "relative flex h-full flex-col overflow-hidden rounded-[var(--radius-xl)] border p-4",
        "bg-[var(--bg-surface)] transition-colors",
        badge.earned
          ? "border-[var(--accent-default)]"
          : "border-[var(--border-subtle)] opacity-60 sepia",
      ].join(" ")}
    >
      {/* Tier indicator */}
      <div
        className="absolute right-0 top-0 h-6 w-6"
        aria-hidden="true"
      >
        <div
          className="absolute right-0 top-0 h-0 w-0
            border-l-[24px] border-t-[24px] border-l-transparent"
          style={{ borderTopColor: tierColour }}
        />
      </div>

      <div className="mb-2 text-3xl" aria-hidden="true">
        {badge.icon}
      </div>
      <h3 className="text-sm font-semibold text-[var(--text-primary)]">
        {badge.name}
      </h3>
      {badge.description && (
        <p className="mt-0.5 text-xs text-[var(--text-tertiary)]">
          {badge.description}
        </p>
      )}
      <div className="mt-auto pt-2">
        {badge.earned && badge.awarded_at ? (
          <p className="text-xs text-[var(--color-success)]">
            Earned on{" "}
            {new Date(badge.awarded_at).toLocaleDateString("en-GB", {
              day: "numeric",
              month: "short",
              year: "numeric",
            })}
          </p>
        ) : (
          <p className="text-xs capitalize text-[var(--text-tertiary)]">
            {badge.tier} tier
          </p>
        )}
      </div>
    </div>
  );
}
