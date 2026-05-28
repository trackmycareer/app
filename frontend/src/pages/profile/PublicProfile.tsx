import { useParams } from "react-router";
import { usePublicProfile } from "@/hooks/queries/useProfileQuery";
import { BadgeCard } from "@/components/BadgeCard";
import {
  SpinnerIcon,
  BriefcaseIcon,
  AwardIcon,
  ZapIcon,
  TrophyIcon,
} from "@/components/icons";
import type { Job, Certification, Skill, Win } from "@/types";

/* ------------------------------------------------------------------ */
/*  Sub-components                                                     */
/* ------------------------------------------------------------------ */

function JobTimeline({ jobs }: { jobs: Job[] }) {
  const sorted = [...jobs].sort((a, b) => {
    const aDate = a.end_date ?? "9999-12-31";
    const bDate = b.end_date ?? "9999-12-31";
    return bDate.localeCompare(aDate);
  });

  return (
    <div className="space-y-3">
      {sorted.map((job) => (
        <div
          key={job.id}
          className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
            bg-[var(--bg-elevated)] p-4"
        >
          <div className="flex items-start justify-between gap-2">
            <div>
              <p className="text-sm font-semibold text-[var(--text-primary)]">
                {job.title}
              </p>
              <p className="text-xs text-[var(--text-secondary)]">{job.company}</p>
            </div>
            <span className="shrink-0 text-xs text-[var(--text-tertiary)]">
              {new Date(job.start_date).toLocaleDateString("en-GB", {
                month: "short",
                year: "numeric",
              })}
              {" – "}
              {job.end_date
                ? new Date(job.end_date).toLocaleDateString("en-GB", {
                    month: "short",
                    year: "numeric",
                  })
                : "Present"}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}

function CertificationList({ certs }: { certs: Certification[] }) {
  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
      {certs.map((cert) => (
        <div
          key={cert.id}
          className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
            bg-[var(--bg-elevated)] p-4"
        >
          <p className="text-sm font-semibold text-[var(--text-primary)]">{cert.name}</p>
          <p className="text-xs text-[var(--text-secondary)]">{cert.provider}</p>
          {cert.earned_date && (
            <p className="mt-1 text-xs text-[var(--text-tertiary)]">
              Earned{" "}
              {new Date(cert.earned_date).toLocaleDateString("en-GB", {
                month: "short",
                year: "numeric",
              })}
            </p>
          )}
        </div>
      ))}
    </div>
  );
}

function SkillsGrid({ skills }: { skills: Skill[] }) {
  return (
    <div className="flex flex-wrap gap-2">
      {skills.map((skill) => (
        <span
          key={skill.id}
          className="inline-flex items-center gap-1.5 rounded-full border
            border-[var(--border-subtle)] bg-[var(--bg-elevated)] px-3 py-1.5
            text-xs font-medium text-[var(--text-secondary)]"
        >
          {skill.name}
          <span className="flex gap-0.5" aria-label={`Proficiency ${skill.proficiency} of 4`}>
            {[1, 2, 3, 4].map((dot) => (
              <span
                key={dot}
                className="inline-block h-1.5 w-1.5 rounded-full"
                style={{
                  backgroundColor:
                    dot <= skill.proficiency
                      ? "var(--accent-default)"
                      : "var(--border-default)",
                }}
                aria-hidden="true"
              />
            ))}
          </span>
        </span>
      ))}
    </div>
  );
}

function RecentWins({ wins }: { wins: Win[] }) {
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

/* ------------------------------------------------------------------ */
/*  Section wrapper                                                    */
/* ------------------------------------------------------------------ */

function ProfileSection({
  title,
  icon,
  children,
}: {
  title: string;
  icon: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <section aria-label={title}>
      <div className="mb-3 flex items-center gap-2">
        <span className="text-[var(--text-tertiary)]">{icon}</span>
        <h2 className="text-sm font-semibold uppercase tracking-wider text-[var(--text-tertiary)]">
          {title}
        </h2>
      </div>
      {children}
    </section>
  );
}

/* ------------------------------------------------------------------ */
/*  Main public profile page                                           */
/* ------------------------------------------------------------------ */

export default function PublicProfile() {
  const { username } = useParams<{ username: string }>();
  const { data: profile, isLoading, isError } = usePublicProfile(username ?? "");

  if (isLoading) {
    return (
      <div
        className="flex min-h-screen items-center justify-center bg-[var(--bg-base)]"
        role="status"
      >
        <SpinnerIcon
          width={32}
          height={32}
          className="animate-spin text-[var(--accent-default)]"
          aria-hidden="true"
        />
        <span className="sr-only">Loading profile...</span>
      </div>
    );
  }

  if (isError || !profile) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-[var(--bg-base)] p-8">
        <p className="text-6xl font-bold text-[var(--text-tertiary)]">404</p>
        <h1 className="text-xl font-semibold text-[var(--text-primary)]">
          Profile not found
        </h1>
        <p className="text-sm text-[var(--text-secondary)]">
          This user does not exist or their profile is not public.
        </p>
      </div>
    );
  }

  const hasBadges = profile.badges && profile.badges.length > 0;
  const hasJobs = profile.jobs && profile.jobs.length > 0;
  const hasCerts = profile.certifications && profile.certifications.length > 0;
  const hasSkills = profile.skills && profile.skills.length > 0;
  const hasWins = profile.wins && profile.wins.length > 0;

  return (
    <div className="min-h-screen bg-[var(--bg-base)]">
      {/* Header */}
      <header className="border-b border-[var(--border-subtle)] bg-[var(--bg-surface)]">
        <div className="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
          <div className="flex items-start gap-4">
            <div
              className="flex h-16 w-16 shrink-0 items-center justify-center rounded-full
                bg-[var(--accent-muted)] text-xl font-bold text-[var(--accent-text)]"
            >
              {profile.name
                .split(" ")
                .map((n) => n[0])
                .join("")
                .toUpperCase()
                .slice(0, 2)}
            </div>
            <div className="min-w-0 flex-1">
              <h1 className="text-2xl font-bold text-[var(--text-primary)]">
                {profile.name}
              </h1>
              {profile.bio && (
                <p className="mt-1 text-sm text-[var(--text-secondary)]">{profile.bio}</p>
              )}
              <div className="mt-2 flex items-center gap-2">
                <span
                  className="inline-flex items-center gap-1.5 rounded-full
                    bg-[var(--accent-default)]/10 px-2.5 py-0.5 text-xs
                    font-medium text-[var(--accent-bright)]"
                >
                  Level {profile.level}
                </span>
                <span className="text-xs text-[var(--text-tertiary)]">
                  {profile.level_title}
                </span>
              </div>
            </div>
          </div>
        </div>
      </header>

      <div className="mx-auto max-w-3xl space-y-8 px-4 py-8 sm:px-6 lg:px-8">
        {/* Badge showcase */}
        {hasBadges && (
          <ProfileSection title="Badges" icon={<AwardIcon width={16} height={16} />}>
            <div className="overflow-x-auto pb-2">
              <div className="flex gap-3">
                {profile.badges.map((badge) => (
                  <div key={badge.id} className="w-48 shrink-0">
                    <BadgeCard badge={badge} />
                  </div>
                ))}
              </div>
            </div>
          </ProfileSection>
        )}

        {/* Jobs */}
        {hasJobs && (
          <ProfileSection
            title="Experience"
            icon={<BriefcaseIcon width={16} height={16} />}
          >
            <JobTimeline jobs={profile.jobs!} />
          </ProfileSection>
        )}

        {/* Certifications */}
        {hasCerts && (
          <ProfileSection
            title="Certifications"
            icon={<AwardIcon width={16} height={16} />}
          >
            <CertificationList certs={profile.certifications!} />
          </ProfileSection>
        )}

        {/* Skills */}
        {hasSkills && (
          <ProfileSection title="Skills" icon={<ZapIcon width={16} height={16} />}>
            <SkillsGrid skills={profile.skills!} />
          </ProfileSection>
        )}

        {/* Wins */}
        {hasWins && (
          <ProfileSection title="Recent Wins" icon={<TrophyIcon width={16} height={16} />}>
            <RecentWins wins={profile.wins!} />
          </ProfileSection>
        )}

        {/* Footer */}
        <footer className="border-t border-[var(--border-subtle)] pt-6 text-center">
          <a
            href="/"
            className="text-xs text-[var(--text-tertiary)] transition-colors
              hover:text-[var(--accent-default)]"
          >
            trackmy<span className="text-[var(--accent-default)]">.</span>career
          </a>
        </footer>
      </div>
    </div>
  );
}
