import { useMemo } from "react";
import { Topbar } from "@/components/Topbar";
import { BadgeCard } from "@/components/BadgeCard";
import { ActivityHeatmap } from "@/components/ActivityHeatmap";
import { SectionLoader } from "@/components/SectionLoader";
import { SectionError } from "@/components/SectionError";
import { TrophyIcon } from "@/components/icons";
import {
  useGamificationProgress,
  useGamificationBadges,
  useGamificationHeatmap,
} from "@/hooks/queries/useGamificationQuery";
import type { Badge } from "@/types";

const TIER_COLOURS: Record<string, string> = {
  gold: "var(--color-warning)",
  silver: "var(--border-strong)",
  bronze: "var(--accent-warm)",
};

export default function Achievements() {
  const {
    data: progress,
    isLoading: progressLoading,
    isError: progressError,
  } = useGamificationProgress();

  const {
    data: badges,
    isLoading: badgesLoading,
    isError: badgesError,
  } = useGamificationBadges();

  const {
    data: heatmap,
    isLoading: heatmapLoading,
  } = useGamificationHeatmap(365);

  const earnedCount = useMemo(
    () => (badges ?? []).filter((b: Badge) => b.earned).length,
    [badges],
  );

  const totalCount = (badges ?? []).length;

  const groupedBadges = useMemo(() => {
    if (!badges) return { gold: [], silver: [], bronze: [] };
    const groups: Record<string, Badge[]> = { gold: [], silver: [], bronze: [] };
    for (const badge of badges) {
      const tier = badge.tier in groups ? badge.tier : "bronze";
      groups[tier].push(badge);
    }
    for (const tier of Object.keys(groups)) {
      groups[tier].sort((a, b) => {
        if (a.earned !== b.earned) return a.earned ? -1 : 1;
        return a.name.localeCompare(b.name);
      });
    }
    return groups;
  }, [badges]);

  return (
    <>
      <Topbar title="Achievements" />
      <div className="mx-auto max-w-5xl space-y-6 p-4 lg:p-6">
        {/* Progress hero */}
        {progressLoading && <SectionLoader label="progress" />}
        {progressError && <SectionError message="Failed to load progress." />}

        {progress && (
          <section
            aria-label="Progress overview"
            className="animate-fade-in-up grid grid-cols-2 gap-px sm:flex overflow-hidden rounded-[var(--radius-xl)]
              border border-[var(--border-subtle)] bg-[var(--border-subtle)]"
          >
            <div className="flex flex-col items-center justify-center bg-[var(--bg-surface)] px-5 py-4">
              <span className="text-xl" aria-hidden="true">
                {progress.current_streak > 0 ? "🔥" : "💤"}
              </span>
              <span className="mt-1 text-xl font-bold text-[var(--text-primary)]">
                {progress.current_streak}
              </span>
              <span className="text-[10px] font-medium uppercase tracking-wider text-[var(--text-tertiary)]">
                Day streak
              </span>
            </div>

            <div className="flex sm:flex-1 items-center gap-3 bg-[var(--bg-surface)] px-5 py-4">
              <div
                className="flex h-10 w-10 items-center justify-center rounded-full
                  bg-[var(--accent-default)] text-sm font-bold text-white"
              >
                {progress.level}
              </div>
              <div className="flex-1">
                <div className="flex items-baseline justify-between">
                  <span className="text-sm font-semibold text-[var(--text-primary)]">
                    {progress.level_title}
                  </span>
                  <span className="text-xs text-[var(--text-tertiary)]">
                    {progress.total_points} / {progress.next_level_at} XP
                  </span>
                </div>
                <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-[var(--bg-elevated)]">
                  <div
                    className="h-full rounded-full bg-[var(--accent-default)] transition-all duration-500"
                    style={{
                      width: `${Math.min(100, (progress.total_points / progress.next_level_at) * 100)}%`,
                    }}
                  />
                </div>
              </div>
            </div>

            <div className="flex flex-col items-center justify-center bg-[var(--bg-surface)] px-5 py-4">
              <span className="text-xl font-bold text-[var(--text-primary)]">
                {progress.total_points}
              </span>
              <span className="text-[10px] font-medium uppercase tracking-wider text-[var(--text-tertiary)]">
                Total points
              </span>
            </div>

            <div className="flex flex-col items-center justify-center bg-[var(--bg-surface)] px-5 py-4">
              <span className="text-xl font-bold text-[var(--text-primary)]">
                {progress.badges_earned}
              </span>
              <span className="text-[10px] font-medium uppercase tracking-wider text-[var(--text-tertiary)]">
                Badges
              </span>
            </div>
          </section>
        )}

        {/* Activity heatmap */}
        {heatmap && !heatmapLoading && (
          <section
            aria-label="Activity heatmap"
            className="animate-fade-in-up rounded-[var(--radius-xl)] border border-[var(--border-subtle)]
              bg-[var(--bg-surface)] p-5"
            style={{ animationDelay: "100ms" }}
          >
            <h2 className="mb-4 text-base font-semibold text-[var(--text-primary)]">Activity</h2>
            <ActivityHeatmap data={heatmap} />
          </section>
        )}

        {/* Badges */}
        <section aria-label="Badges" className="animate-fade-in-up" style={{ animationDelay: "200ms" }}>
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-base font-semibold text-[var(--text-primary)]">Badges</h2>
            {!badgesLoading && !badgesError && (
              <span className="text-sm text-[var(--text-tertiary)]">
                {earnedCount} of {totalCount} earned
              </span>
            )}
          </div>

          {badgesLoading && <SectionLoader label="badges" />}
          {badgesError && <SectionError message="Failed to load badges." />}

          {!badgesLoading && !badgesError && totalCount === 0 && (
            <div className="py-12 text-center">
              <TrophyIcon
                width={40}
                height={40}
                className="mx-auto mb-3 text-[var(--text-tertiary)]"
                aria-hidden="true"
              />
              <p className="text-sm text-[var(--text-secondary)]">No badges available yet.</p>
            </div>
          )}

          {!badgesLoading &&
            !badgesError &&
            totalCount > 0 &&
            (["gold", "silver", "bronze"] as const).map((tier) => {
              const tierBadges = groupedBadges[tier];
              if (tierBadges.length === 0) return null;
              const tierEarned = tierBadges.filter((b) => b.earned).length;
              const tierTotal = tierBadges.length;
              const tierColour = TIER_COLOURS[tier] ?? "var(--text-tertiary)";

              return (
                <div key={tier} className="mb-6">
                  <div className="mb-3 flex items-center gap-3">
                    <div
                      className="h-2.5 w-2.5 rounded-full"
                      style={{ backgroundColor: tierColour }}
                      aria-hidden="true"
                    />
                    <h3 className="text-sm font-semibold capitalize text-[var(--text-primary)]">
                      {tier}
                    </h3>
                    <span className="ml-auto text-xs text-[var(--text-tertiary)]">
                      {tierEarned} of {tierTotal}
                    </span>
                    <div className="h-1 w-24 overflow-hidden rounded-full bg-[var(--bg-elevated)]">
                      <div
                        className="h-full rounded-full transition-all duration-500"
                        style={{
                          width: `${tierTotal > 0 ? (tierEarned / tierTotal) * 100 : 0}%`,
                          backgroundColor: tierColour,
                        }}
                      />
                    </div>
                  </div>
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                    {tierBadges.map((badge) => (
                      <BadgeCard key={badge.id} badge={badge} />
                    ))}
                  </div>
                </div>
              );
            })}
        </section>
      </div>
    </>
  );
}
