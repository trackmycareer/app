import { useMemo } from "react";
import { Link } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { CategoryBadge } from "@/components/CategoryBadge";
import { ActivityHeatmap } from "@/components/ActivityHeatmap";
import { SectionLoader } from "@/components/SectionLoader";
import { SectionError } from "@/components/SectionError";
import {
  TrophyIcon,
  AwardIcon,
  AlertTriangleIcon,
} from "@/components/icons";
import { useAuthStore } from "@/stores/auth";
import { useWinsQuery } from "@/hooks/queries/useWinsQuery";
import { useJobsQuery } from "@/hooks/queries/useJobsQuery";
import { useCertsQuery } from "@/hooks/queries/useCertsQuery";
import { useSkillsQuery } from "@/hooks/queries/useSkillsQuery";
import {
  useGamificationProgress,
  useGamificationHeatmap,
} from "@/hooks/queries/useGamificationQuery";
import type { Job, Certification, Skill, Win } from "@/types";

const CATEGORY_MARKER_COLOURS: Record<string, string> = {
  shipped_feature: "var(--accent-default)",
  positive_feedback: "var(--accent-warm)",
  process_improvement: "var(--color-info)",
  cost_saving: "var(--color-warning)",
  leadership_moment: "var(--color-error)",
  general: "var(--text-tertiary)",
  project: "var(--color-info)",
  publication: "var(--accent-muted)",
  membership: "var(--color-success)",
};

function getGreeting(): string {
  const hour = new Date().getHours();
  if (hour < 12) return "Good morning";
  if (hour < 18) return "Good afternoon";
  return "Good evening";
}

function daysUntil(dateString: string): number {
  const target = new Date(dateString);
  const now = new Date();
  target.setHours(0, 0, 0, 0);
  now.setHours(0, 0, 0, 0);
  return Math.ceil((target.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
}

function formatDuration(startDate: string): string {
  const start = new Date(startDate);
  const now = new Date();
  const months = (now.getFullYear() - start.getFullYear()) * 12 + now.getMonth() - start.getMonth();
  const years = Math.floor(months / 12);
  const remainingMonths = months % 12;
  if (years === 0) return `${remainingMonths}m`;
  if (remainingMonths === 0) return `${years}y`;
  return `${years}y ${remainingMonths}m`;
}

export default function Dashboard() {
  const userName = useAuthStore((s) => s.user?.name);
  const firstName = userName?.split(" ")[0] ?? "";

  const {
    data: winsData,
    isLoading: winsLoading,
    isError: winsError,
  } = useWinsQuery({ limit: 5, offset: 0 });

  const { data: jobsData, isLoading: jobsLoading } = useJobsQuery();

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

  const wins: Win[] = useMemo(() => {
    if (!winsData) return [];
    return winsData?.data ?? [];
  }, [winsData]);

  const totalWins = useMemo(() => {
    if (!winsData) return 0;
    return (winsData as { data: Win[]; total?: number })?.total ?? wins.length;
  }, [winsData, wins.length]);

  const currentRole: Job | null = useMemo(() => {
    if (!jobsData) return null;
    const jobs: Job[] = Array.isArray(jobsData) ? jobsData : [];
    return jobs.find((j) => !j.end_date) ?? null;
  }, [jobsData]);

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
      .sort((a, b) => new Date(a.expiry_date!).getTime() - new Date(b.expiry_date!).getTime());
  }, [certs]);

  const expiringCount = expiringCerts.length;

  const skills: Skill[] = useMemo(() => {
    if (!skillsData) return [];
    const raw = skillsData as unknown;
    if (Array.isArray(raw)) return raw as Skill[];
    if (raw && typeof raw === "object" && "skills" in raw) {
      return (raw as { skills: Skill[] }).skills ?? [];
    }
    return [];
  }, [skillsData]);

  const expertCount = useMemo(() => skills.filter((s) => s.proficiency === 4).length, [skills]);

  const topSkills = useMemo(
    () => [...skills].sort((a, b) => b.proficiency - a.proficiency).slice(0, 8),
    [skills],
  );

  const greeting = getGreeting();
  const streakText = gamificationProgress && gamificationProgress.current_streak > 0
    ? `You're on a ${gamificationProgress.current_streak}-day streak. Keep it going.`
    : "Start logging wins to build your streak.";

  return (
    <>
      <Topbar title={firstName ? `${greeting}, ${firstName}` : greeting} />

      <div className="mx-auto max-w-5xl space-y-6 p-4 lg:p-6">
        {/* Hero stats bar */}
        <section aria-label="Career overview" className="animate-fade-in-up">
          <p className="mb-4 text-sm text-[var(--text-tertiary)]">{streakText}</p>
          <div
            className="grid grid-cols-2 gap-px overflow-hidden rounded-[var(--radius-xl)] border
              border-[var(--border-subtle)] bg-[var(--border-subtle)] sm:flex"
          >
            <div className="flex flex-col items-center justify-center bg-[var(--bg-surface)] px-4 py-4 sm:flex-1">
              <span className="text-2xl font-bold tracking-tight text-[var(--text-primary)]">
                {winsLoading ? "—" : totalWins}
              </span>
              <span className="mt-1 text-[11px] font-medium uppercase tracking-wider text-[var(--text-tertiary)]">
                Wins
              </span>
            </div>
            <div className="flex flex-col items-center justify-center bg-[var(--bg-surface)] px-4 py-4 sm:flex-1">
              {jobsLoading ? (
                <span className="text-sm text-[var(--text-tertiary)]">—</span>
              ) : currentRole ? (
                <>
                  <span className="text-center text-sm font-semibold text-[var(--text-primary)]">
                    {currentRole.title}
                  </span>
                  <span className="mt-1 text-[11px] font-medium uppercase tracking-wider text-[var(--text-tertiary)]">
                    Current role
                  </span>
                  {currentRole.start_date && (
                    <span className="mt-0.5 text-xs font-medium text-[var(--accent-default)]">
                      {formatDuration(currentRole.start_date)}
                    </span>
                  )}
                </>
              ) : (
                <>
                  <span className="text-sm text-[var(--text-tertiary)]">Not set</span>
                  <span className="mt-1 text-[11px] font-medium uppercase tracking-wider text-[var(--text-tertiary)]">
                    Current role
                  </span>
                </>
              )}
            </div>
            <div className="flex flex-col items-center justify-center bg-[var(--bg-surface)] px-4 py-4 sm:flex-1">
              <span className="text-2xl font-bold tracking-tight text-[var(--text-primary)]">
                {certsLoading ? "—" : activeCertsCount}
              </span>
              <span className="mt-1 text-[11px] font-medium uppercase tracking-wider text-[var(--text-tertiary)]">
                Active certs
              </span>
              {expiringCount > 0 && (
                <span className="mt-0.5 text-xs font-medium text-[var(--color-warning)]">
                  {expiringCount} expiring
                </span>
              )}
            </div>
            <div className="flex flex-col items-center justify-center bg-[var(--bg-surface)] px-4 py-4 sm:flex-1">
              <span className="text-2xl font-bold tracking-tight text-[var(--text-primary)]">
                {skillsLoading ? "—" : skills.length}
              </span>
              <span className="mt-1 text-[11px] font-medium uppercase tracking-wider text-[var(--text-tertiary)]">
                Skills
              </span>
              {expertCount > 0 && (
                <span className="mt-0.5 text-xs font-medium text-[var(--accent-default)]">
                  {expertCount} expert
                </span>
              )}
            </div>
          </div>
        </section>

        {/* Streak + Level row */}
        {gamificationProgress && (
          <section
            aria-label="Progress"
            className="animate-fade-in-up flex flex-col gap-3 rounded-[var(--radius-xl)]
              border border-[var(--border-subtle)] bg-[var(--bg-surface)] px-5 py-3.5 sm:flex-row sm:items-center"
            style={{ animationDelay: "100ms" }}
          >
            <div className="flex items-center gap-3">
              <span className="text-xl" aria-hidden="true">
                {gamificationProgress.current_streak > 0 ? "🔥" : "💤"}
              </span>
              <div>
                <p className="text-[11px] font-medium uppercase tracking-wider text-[var(--text-tertiary)]">
                  Streak
                </p>
                <p className="text-lg font-bold text-[var(--text-primary)]">
                  {gamificationProgress.current_streak}
                  <span className="ml-1 text-xs font-normal text-[var(--text-tertiary)]">days</span>
                </p>
              </div>
            </div>

            <div className="hidden bg-[var(--border-subtle)] sm:mx-2 sm:block sm:h-8 sm:w-px" aria-hidden="true" />

            <div className="flex flex-1 items-center gap-3">
              <div
                className="flex h-9 w-9 items-center justify-center rounded-full
                  bg-[var(--accent-default)] text-sm font-bold text-white"
              >
                {gamificationProgress.level}
              </div>
              <div className="flex-1">
                <div className="flex items-baseline justify-between">
                  <span className="text-sm font-semibold text-[var(--text-primary)]">
                    {gamificationProgress.level_title}
                  </span>
                  <span className="text-xs text-[var(--text-tertiary)]">
                    {gamificationProgress.total_points} / {gamificationProgress.next_level_at} XP
                  </span>
                </div>
                <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-[var(--bg-elevated)]">
                  <div
                    className="h-full rounded-full bg-[var(--accent-default)] transition-all duration-500"
                    style={{
                      width: `${Math.min(100, (gamificationProgress.total_points / gamificationProgress.next_level_at) * 100)}%`,
                    }}
                  />
                </div>
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
              <h2 className="text-base font-semibold text-[var(--text-primary)]">Activity</h2>
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

        {/* Two-column: Recent Wins + Certs/Skills */}
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
              <h2 className="text-base font-semibold text-[var(--text-primary)]">Recent wins</h2>
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
                  Welcome to your career dashboard. Once you start logging wins, this is where the
                  magic happens.
                </p>
                <Link to="/wins">
                  <Button size="sm">Record a win</Button>
                </Link>
              </div>
            )}

            {!winsLoading && !winsError && wins.length > 0 && (
              <>
                <ul className="space-y-1" aria-label="Recent wins list">
                  {wins.map((win) => (
                    <li key={win.id} className="flex gap-3 py-2.5 border-b border-[var(--border-subtle)] last:border-b-0">
                      <div
                        className="mt-0.5 w-[3px] flex-shrink-0 self-stretch rounded-full"
                        style={{
                          backgroundColor: CATEGORY_MARKER_COLOURS[win.category] ?? "var(--text-tertiary)",
                        }}
                        aria-hidden="true"
                      />
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium text-[var(--text-primary)]">{win.title}</p>
                        <div className="mt-1 flex flex-wrap items-center gap-2">
                          <CategoryBadge category={win.category} />
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
                        </div>
                      </div>
                    </li>
                  ))}
                </ul>
                <div className="mt-4">
                  <Link to="/wins">
                    <Button size="sm" variant="secondary">Record a win</Button>
                  </Link>
                </div>
              </>
            )}
          </section>

          {/* Certifications + Skills */}
          <section
            aria-label="Certifications and skills"
            className="rounded-[var(--radius-xl)] border border-[var(--border-subtle)]
              bg-[var(--bg-surface)] p-5"
          >
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-base font-semibold text-[var(--text-primary)]">Expiring soon</h2>
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
              <div className="py-6 text-center">
                <AwardIcon
                  width={32}
                  height={32}
                  className="mx-auto mb-2 text-[var(--text-tertiary)]"
                  aria-hidden="true"
                />
                <p className="text-sm text-[var(--text-secondary)]">No certifications expiring soon.</p>
              </div>
            )}

            {!certsLoading && !certsError && expiringCerts.length > 0 && (
              <ul className="space-y-1" aria-label="Expiring certifications list">
                {expiringCerts.map((cert) => {
                  const remaining = daysUntil(cert.expiry_date!);
                  const isUrgent = remaining <= 30;
                  return (
                    <li
                      key={cert.id}
                      className="flex items-center justify-between border-b border-[var(--border-subtle)] py-2.5 last:border-b-0"
                    >
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium text-[var(--text-primary)]">{cert.name}</p>
                        <p className="text-xs text-[var(--text-tertiary)]">
                          Expires{" "}
                          {new Date(cert.expiry_date!).toLocaleDateString("en-GB", {
                            day: "numeric",
                            month: "short",
                            year: "numeric",
                          })}
                        </p>
                      </div>
                      <div className="ml-3 flex items-center gap-1.5">
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
                            "rounded-[var(--radius-sm)] px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider",
                            isUrgent
                              ? "bg-[var(--color-warning)]/12 text-[var(--color-warning)]"
                              : "bg-[var(--color-success)]/12 text-[var(--color-success)]",
                          ].join(" ")}
                        >
                          {remaining}d left
                        </span>
                      </div>
                    </li>
                  );
                })}
              </ul>
            )}

            {/* Top skills as chips */}
            {!skillsLoading && !skillsError && topSkills.length > 0 && (
              <div className="mt-5 border-t border-[var(--border-subtle)] pt-4">
                <div className="mb-2.5 flex items-center justify-between">
                  <h3 className="text-xs font-semibold uppercase tracking-wider text-[var(--text-tertiary)]">
                    Top skills
                  </h3>
                  <Link
                    to="/skills"
                    className="text-xs font-medium text-[var(--accent-default)] hover:underline"
                  >
                    View all
                  </Link>
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {topSkills.map((skill) => (
                    <span
                      key={skill.id}
                      className="rounded-[var(--radius-md)] border border-[var(--border-subtle)]
                        bg-[var(--bg-elevated)] px-2.5 py-1 text-xs font-medium text-[var(--text-secondary)]"
                    >
                      {skill.name}
                    </span>
                  ))}
                  {skills.length > 8 && (
                    <Link
                      to="/skills"
                      className="rounded-[var(--radius-md)] border border-[var(--border-subtle)]
                        bg-[var(--bg-elevated)] px-2.5 py-1 text-xs font-medium text-[var(--accent-default)]
                        hover:bg-[var(--bg-hover)]"
                    >
                      +{skills.length - 8}
                    </Link>
                  )}
                </div>
              </div>
            )}
          </section>
        </div>
      </div>
    </>
  );
}
