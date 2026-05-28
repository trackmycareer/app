import { useMemo } from "react";
import { Topbar } from "@/components/Topbar";
import { LevelBar } from "@/components/LevelBar";
import { StreakCounter } from "@/components/StreakCounter";
import { BadgeCard } from "@/components/BadgeCard";
import { ActivityHeatmap } from "@/components/ActivityHeatmap";
import { SpinnerIcon, TrophyIcon } from "@/components/icons";
import {
  useGamificationProgress,
  useGamificationBadges,
  useGamificationHeatmap,
} from "@/hooks/queries/useGamificationQuery";
import type { Badge } from "@/types";

function SectionLoader({ label }: { label: string }) {
  return (
    <div className="flex items-center justify-center py-10" role="status">
      <SpinnerIcon
        width={22}
        height={22}
        className="animate-spin text-[var(--accent-default)]"
        aria-hidden="true"
      />
      <span className="sr-only">Loading {label}...</span>
    </div>
  );
}

function SectionError({ message }: { message: string }) {
  return (
    <div
      className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
        bg-[var(--color-error)]/5 p-3 text-center text-sm text-[var(--color-error)]"
      role="alert"
    >
      {message}
    </div>
  );
}

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
    // Sort each tier: earned first, then by name
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
      <div className="mx-auto max-w-5xl space-y-8 p-4 lg:p-6">
        {/* Top section: Level, Points, Streak */}
        {progressLoading && <SectionLoader label="progress" />}
        {progressError && <SectionError message="Failed to load progress." />}

        {progress && (
          <section aria-label="Progress overview" className="space-y-4">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div className="sm:col-span-2">
                <LevelBar
                  level={progress.level}
                  levelTitle={progress.level_title}
                  totalPoints={progress.total_points}
                  nextLevelAt={progress.next_level_at}
                />
                <div className="mt-3 flex items-center gap-4">
                  <div
                    className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
                      bg-[var(--bg-surface)] px-4 py-3 text-center"
                  >
                    <p className="text-xl font-bold text-[var(--text-primary)]">
                      {progress.total_points}
                    </p>
                    <p className="text-xs text-[var(--text-tertiary)]">
                      Total points
                    </p>
                  </div>
                  <div
                    className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
                      bg-[var(--bg-surface)] px-4 py-3 text-center"
                  >
                    <p className="text-xl font-bold text-[var(--text-primary)]">
                      {progress.badges_earned}
                    </p>
                    <p className="text-xs text-[var(--text-tertiary)]">
                      Badges earned
                    </p>
                  </div>
                </div>
              </div>
              <StreakCounter
                currentStreak={progress.current_streak}
                longestStreak={progress.longest_streak}
              />
            </div>
          </section>
        )}

        {/* Activity heatmap */}
        {heatmap && !heatmapLoading && (
          <section
            aria-label="Activity heatmap"
            className="rounded-[var(--radius-xl)] border border-[var(--border-subtle)]
              bg-[var(--bg-surface)] p-5"
          >
            <h2 className="mb-4 text-base font-semibold text-[var(--text-primary)]">
              Activity
            </h2>
            <ActivityHeatmap data={heatmap} />
          </section>
        )}

        {/* Badges */}
        <section aria-label="Badges">
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-base font-semibold text-[var(--text-primary)]">
              Badges
            </h2>
            {!badgesLoading && !badgesError && (
              <span className="text-sm text-[var(--text-tertiary)]">
                {earnedCount} of {totalCount} badge{totalCount !== 1 ? "s" : ""} earned
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
              <p className="text-sm text-[var(--text-secondary)]">
                No badges available yet.
              </p>
            </div>
          )}

          {!badgesLoading &&
            !badgesError &&
            totalCount > 0 &&
            (["gold", "silver", "bronze"] as const).map((tier) => {
              const tierBadges = groupedBadges[tier];
              if (tierBadges.length === 0) return null;
              return (
                <div key={tier} className="mb-6">
                  <h3 className="mb-3 text-sm font-medium capitalize text-[var(--text-secondary)]">
                    {tier}
                  </h3>
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
