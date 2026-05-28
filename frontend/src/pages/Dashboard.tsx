import { useMemo } from "react";
import { Link } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { CategoryBadge } from "@/components/CategoryBadge";
import { StreakCounter } from "@/components/StreakCounter";
import { LevelBar } from "@/components/LevelBar";
import { ActivityHeatmap } from "@/components/ActivityHeatmap";
import { BadgeCard } from "@/components/BadgeCard";
import {
  TrophyIcon,
  BuildingIcon,
  AwardIcon,
  ZapIcon,
  SpinnerIcon,
  AlertTriangleIcon,
} from "@/components/icons";
import { useWinsQuery } from "@/hooks/queries/useWinsQuery";
import { useJobsQuery } from "@/hooks/queries/useJobsQuery";
import { useCertsQuery } from "@/hooks/queries/useCertsQuery";
import { useSkillsQuery } from "@/hooks/queries/useSkillsQuery";
import {
  useGamificationProgress,
  useGamificationHeatmap,
  useGamificationBadges,
} from "@/hooks/queries/useGamificationQuery";
import type { Job, Certification, Skill, Win, Badge } from "@/types";

/* ------------------------------------------------------------------ */
/*  Proficiency helpers (mirrored from SkillsList)                    */
/* ------------------------------------------------------------------ */

const PROFICIENCY_LABELS: Record<number, string> = {
  1: "Beginner",
  2: "Intermediate",
  3: "Advanced",
  4: "Expert",
};

const PROFICIENCY_COLOURS: Record<number, string> = {
  1: "var(--text-tertiary)",
  2: "var(--accent-default)",
  3: "var(--color-warning, #d97706)",
  4: "var(--color-success, #16a34a)",
};

function ProficiencyDots({ level }: { level: number }) {
  return (
    <div
      className="flex items-center gap-1"
      aria-label={`Proficiency: ${PROFICIENCY_LABELS[level]}`}
    >
      {[1, 2, 3, 4].map((dot) => (
        <span
          key={dot}
          className="inline-block h-1.5 w-1.5 rounded-full"
          style={{
            backgroundColor:
              dot <= level ? PROFICIENCY_COLOURS[level] : "var(--border-default)",
          }}
          aria-hidden="true"
        />
      ))}
      <span className="ml-1 text-xs text-[var(--text-tertiary)]">
        {PROFICIENCY_LABELS[level]}
      </span>
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Section skeleton / loading                                        */
/* ------------------------------------------------------------------ */

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

/* ------------------------------------------------------------------ */
/*  Stat card                                                         */
/* ------------------------------------------------------------------ */

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  subtext?: string;
  loading?: boolean;
}

function StatCard({
  icon,
  label,
  value,
  subtext,
  loading,
  borderColour,
}: StatCardProps & { borderColour?: string }) {
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

/* ------------------------------------------------------------------ */
/*  Helper: days remaining until a date                               */
/* ------------------------------------------------------------------ */

function daysUntil(dateString: string): number {
  const target = new Date(dateString);
  const now = new Date();
  // Reset time portions for accurate day calculation
  target.setHours(0, 0, 0, 0);
  now.setHours(0, 0, 0, 0);
  return Math.ceil((target.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
}

/* ------------------------------------------------------------------ */
/*  Dashboard page                                                    */
/* ------------------------------------------------------------------ */

export default function Dashboard() {
  /* ---- Data fetching ---- */
  const {
    data: winsData,
    isLoading: winsLoading,
    isError: winsError,
  } = useWinsQuery({ limit: 5, offset: 0 });

  const {
    data: jobsData,
    isLoading: jobsLoading,
  } = useJobsQuery();

  const {
    data: certsData,
    isLoading: certsLoading,
    isError: certsError,
  } = useCertsQuery({});

  const {
    data: skillsData,
    isLoading: skillsLoading,
    isError: skillsError,
  } = useSkillsQuery({});

  const { data: gamificationProgress } = useGamificationProgress();
  const { data: heatmapData } = useGamificationHeatmap(365);
  const { data: allBadges } = useGamificationBadges();

  /* ---- Derived data ---- */

  // Wins
  const wins: Win[] = useMemo(() => {
    if (!winsData) return [];
    return winsData?.data ?? [];
  }, [winsData]);

  const totalWins = useMemo(() => {
    if (!winsData) return 0;
    return (winsData as { data: Win[]; total?: number })?.total ?? wins.length;
  }, [winsData, wins.length]);

  // Jobs: find current role (no end_date)
  const currentRole: Job | null = useMemo(() => {
    if (!jobsData) return null;
    const jobs: Job[] = Array.isArray(jobsData) ? jobsData : [];
    return jobs.find((j) => !j.end_date) ?? null;
  }, [jobsData]);

  // Certifications
  const certs: Certification[] = useMemo(() => {
    if (!certsData) return [];
    const raw = certsData as unknown;
    if (Array.isArray(raw)) return raw as Certification[];
    if (raw && typeof raw === "object" && "certifications" in raw) {
      return (raw as { certifications: Certification[] }).certifications ?? [];
    }
    return [];
  }, [certsData]);

  const activeCertsCount = useMemo(
    () => certs.filter((c) => c.status === "passed").length,
    [certs],
  );

  const expiringCerts = useMemo(() => {
    const now = new Date();
    const ninetyDaysMs = 90 * 24 * 60 * 60 * 1000;
    return certs
      .filter((c) => {
        if (c.status !== "passed" || !c.expiry_date) return false;
        const expiry = new Date(c.expiry_date);
        const diff = expiry.getTime() - now.getTime();
        return diff > 0 && diff <= ninetyDaysMs;
      })
      .sort((a, b) => {
        return new Date(a.expiry_date!).getTime() - new Date(b.expiry_date!).getTime();
      });
  }, [certs]);

  // Skills
  const skills: Skill[] = useMemo(() => {
    if (!skillsData) return [];
    const raw = skillsData as unknown;
    if (Array.isArray(raw)) return raw as Skill[];
    if (raw && typeof raw === "object" && "skills" in raw) {
      return (raw as { skills: Skill[] }).skills ?? [];
    }
    return [];
  }, [skillsData]);

  const topSkills = useMemo(
    () =>
      [...skills]
        .sort((a, b) => b.proficiency - a.proficiency)
        .slice(0, 6),
    [skills],
  );

  const recentBadges: Badge[] = useMemo(() => {
    if (!allBadges || !Array.isArray(allBadges)) return [];
    return [...allBadges]
      .filter((b) => b.earned && b.awarded_at)
      .sort((a, b) => {
        const aDate = a.awarded_at ?? "";
        const bDate = b.awarded_at ?? "";
        return bDate.localeCompare(aDate);
      })
      .slice(0, 3);
  }, [allBadges]);

  /* ---- Render ---- */

  return (
    <>
      <Topbar title="Dashboard" />

      <div className="mx-auto max-w-5xl space-y-8 p-4 lg:p-6">
        {/* Quick stats row */}
        <section aria-label="Career overview" className="animate-fade-in-up">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatCard
              icon={<TrophyIcon width={16} height={16} />}
              label="Total wins"
              value={totalWins}
              loading={winsLoading}
              borderColour="border-l-4 border-l-[var(--accent-default)]"
            />
            <StatCard
              icon={<BuildingIcon width={16} height={16} />}
              label="Current role"
              value={
                currentRole
                  ? currentRole.title
                  : jobsLoading
                    ? ""
                    : "Not set"
              }
              subtext={currentRole?.company}
              loading={jobsLoading}
              borderColour="border-l-4 border-l-[var(--color-success)]"
            />
            <StatCard
              icon={<AwardIcon width={16} height={16} />}
              label="Active certifications"
              value={activeCertsCount}
              loading={certsLoading}
              borderColour="border-l-4 border-l-[var(--color-warning)]"
            />
            <StatCard
              icon={<ZapIcon width={16} height={16} />}
              label="Skills tracked"
              value={skills.length}
              loading={skillsLoading}
              borderColour="border-l-4 border-l-purple-400"
            />
          </div>
        </section>

        {/* Gamification: Streak + Level + Recent Badges */}
        {gamificationProgress && (
          <section
            aria-label="Gamification overview"
            className="animate-fade-in-up"
            style={{ animationDelay: "100ms" }}
          >
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <StreakCounter
                currentStreak={gamificationProgress.current_streak}
                longestStreak={gamificationProgress.longest_streak}
              />
              <div className="sm:col-span-2 space-y-4">
                <LevelBar
                  level={gamificationProgress.level}
                  levelTitle={gamificationProgress.level_title}
                  totalPoints={gamificationProgress.total_points}
                  nextLevelAt={gamificationProgress.next_level_at}
                />
                {recentBadges.length > 0 && (
                  <div>
                    <div className="mb-2 flex items-center justify-between">
                      <h3 className="text-sm font-medium text-[var(--text-secondary)]">
                        Recent badges
                      </h3>
                      <Link
                        to="/achievements"
                        className="text-xs font-medium text-[var(--accent-default)] hover:underline"
                      >
                        View all
                      </Link>
                    </div>
                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                      {recentBadges.map((badge) => (
                        <BadgeCard key={badge.id} badge={badge} />
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          </section>
        )}

        {/* Activity heatmap */}
        {heatmapData && heatmapData.length > 0 && (
          <section
            aria-label="Activity heatmap"
            className="animate-fade-in-up rounded-[var(--radius-xl)] border
              border-[var(--border-subtle)] bg-[var(--bg-surface)] p-5"
            style={{ animationDelay: "200ms" }}
          >
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-base font-semibold text-[var(--text-primary)]">
                Activity
              </h2>
              <Link
                to="/achievements"
                className="text-xs font-medium text-[var(--accent-default)] hover:underline"
              >
                Details
              </Link>
            </div>
            <ActivityHeatmap data={heatmapData} />
          </section>
        )}

        {/* Two-column layout for main sections */}
        <div
          className="animate-fade-in-up grid grid-cols-1 gap-6 lg:grid-cols-2"
          style={{ animationDelay: "300ms" }}
        >
          {/* Recent Wins */}
          <section
            aria-label="Recent wins"
            className="rounded-[var(--radius-xl)] border border-[var(--border-subtle)]
              bg-[var(--bg-surface)] p-5"
          >
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-base font-semibold text-[var(--text-primary)]">
                Recent wins
              </h2>
              <Link
                to="/wins"
                className="text-xs font-medium text-[var(--accent-default)] hover:underline"
              >
                View all
              </Link>
            </div>

            {winsLoading && <SectionLoader label="wins" />}
            {winsError && <SectionError message="Failed to load wins." />}

            {!winsLoading && !winsError && wins.length === 0 && (
              <div className="py-8 text-center">
                <TrophyIcon
                  width={36}
                  height={36}
                  className="mx-auto mb-3 text-[var(--text-tertiary)]"
                  aria-hidden="true"
                />
                <p className="mb-3 text-sm text-[var(--text-secondary)]">
                  Welcome to your career dashboard. Once you start logging wins, this is
                  where the magic happens.
                </p>
                <Link to="/wins">
                  <Button size="sm">Record a win</Button>
                </Link>
              </div>
            )}

            {!winsLoading && !winsError && wins.length > 0 && (
              <>
                <ul className="space-y-3" aria-label="Recent wins list">
                  {wins.map((win) => (
                    <li
                      key={win.id}
                      className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
                        bg-[var(--bg-base)] p-3 transition-colors
                        hover:border-[var(--border-default)]"
                    >
                      <div className="mb-1 flex flex-wrap items-center gap-2">
                        <span className="text-sm font-medium text-[var(--text-primary)]">
                          {win.title}
                        </span>
                        <CategoryBadge category={win.category} />
                      </div>
                      <time
                        dateTime={win.occurred_on}
                        className="text-xs text-[var(--text-tertiary)]"
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

                <div className="mt-4">
                  <Link to="/wins">
                    <Button size="sm" variant="secondary">
                      Record a win
                    </Button>
                  </Link>
                </div>
              </>
            )}
          </section>

          {/* Certifications Expiring Soon */}
          <section
            aria-label="Certifications expiring soon"
            className="rounded-[var(--radius-xl)] border border-[var(--border-subtle)]
              bg-[var(--bg-surface)] p-5"
          >
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-base font-semibold text-[var(--text-primary)]">
                Expiring soon
              </h2>
              <Link
                to="/certifications"
                className="text-xs font-medium text-[var(--accent-default)] hover:underline"
              >
                View all
              </Link>
            </div>

            {certsLoading && <SectionLoader label="certifications" />}
            {certsError && <SectionError message="Failed to load certifications." />}

            {!certsLoading && !certsError && expiringCerts.length === 0 && (
              <div className="py-8 text-center">
                <AwardIcon
                  width={36}
                  height={36}
                  className="mx-auto mb-3 text-[var(--text-tertiary)]"
                  aria-hidden="true"
                />
                <p className="text-sm text-[var(--text-secondary)]">
                  No certifications expiring soon.
                </p>
              </div>
            )}

            {!certsLoading && !certsError && expiringCerts.length > 0 && (
              <ul className="space-y-3" aria-label="Expiring certifications list">
                {expiringCerts.map((cert) => {
                  const remaining = daysUntil(cert.expiry_date!);
                  const isUrgent = remaining <= 30;

                  return (
                    <li
                      key={cert.id}
                      className={[
                        "rounded-[var(--radius-lg)] border p-3 transition-colors",
                        "hover:border-[var(--border-default)]",
                        isUrgent
                          ? "border-[var(--color-warning)]/30 bg-[var(--color-warning)]/5"
                          : "border-[var(--border-subtle)] bg-[var(--bg-base)]",
                      ].join(" ")}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0 flex-1">
                          <p className="text-sm font-medium text-[var(--text-primary)]">
                            {cert.name}
                          </p>
                          <p className="text-xs text-[var(--text-tertiary)]">
                            {cert.provider}
                          </p>
                        </div>
                        <div className="flex shrink-0 flex-col items-end gap-0.5">
                          {isUrgent && (
                            <AlertTriangleIcon
                              width={14}
                              height={14}
                              className="text-[var(--color-warning)]"
                              aria-hidden="true"
                            />
                          )}
                          <span
                            className={[
                              "text-xs font-medium",
                              isUrgent
                                ? "text-[var(--color-warning)]"
                                : "text-[var(--text-secondary)]",
                            ].join(" ")}
                          >
                            {remaining} day{remaining !== 1 ? "s" : ""} left
                          </span>
                        </div>
                      </div>
                      <time
                        dateTime={cert.expiry_date!}
                        className="mt-1 block text-xs text-[var(--text-tertiary)]"
                      >
                        Expires{" "}
                        {new Date(cert.expiry_date!).toLocaleDateString("en-GB", {
                          day: "numeric",
                          month: "short",
                          year: "numeric",
                        })}
                      </time>
                    </li>
                  );
                })}
              </ul>
            )}
          </section>
        </div>

        {/* Skills Summary */}
        <section
          aria-label="Skills summary"
          className="animate-fade-in-up rounded-[var(--radius-xl)] border
            border-[var(--border-subtle)] bg-[var(--bg-surface)] p-5"
          style={{ animationDelay: "400ms" }}
        >
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-base font-semibold text-[var(--text-primary)]">
              Top skills
            </h2>
            <Link
              to="/skills"
              className="text-xs font-medium text-[var(--accent-default)] hover:underline"
            >
              View all
            </Link>
          </div>

          {skillsLoading && <SectionLoader label="skills" />}
          {skillsError && <SectionError message="Failed to load skills." />}

          {!skillsLoading && !skillsError && topSkills.length === 0 && (
            <div className="py-8 text-center">
              <ZapIcon
                width={36}
                height={36}
                className="mx-auto mb-3 text-[var(--text-tertiary)]"
                aria-hidden="true"
              />
              <p className="mb-3 text-sm text-[var(--text-secondary)]">
                No skills tracked yet.
              </p>
              <Link to="/skills/new">
                <Button size="sm">Add a skill</Button>
              </Link>
            </div>
          )}

          {!skillsLoading && !skillsError && topSkills.length > 0 && (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {topSkills.map((skill) => (
                <div
                  key={skill.id}
                  className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
                    bg-[var(--bg-base)] p-3"
                >
                  <p className="mb-1 text-sm font-medium text-[var(--text-primary)]">
                    {skill.name}
                  </p>
                  <ProficiencyDots level={skill.proficiency} />
                  {skill.category && (
                    <p className="mt-1.5 text-xs text-[var(--text-tertiary)]">
                      {skill.category}
                    </p>
                  )}
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </>
  );
}
