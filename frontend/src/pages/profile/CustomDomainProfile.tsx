import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";
import { BadgeCard } from "@/components/BadgeCard";
import {
  SpinnerIcon,
  BriefcaseIcon,
  AwardIcon,
  ZapIcon,
  TrophyIcon,
  HeartIcon,
  ShieldIcon,
} from "@/components/icons";
import { ProfileSection } from "./components/ProfileSection";
import { JobTimeline } from "./components/JobTimeline";
import { CertificationList } from "./components/CertificationList";
import { SkillsGrid } from "./components/SkillsGrid";
import { RecentWins } from "./components/RecentWins";
import type { PublicProfile } from "@/types";

/* ------------------------------------------------------------------ */
/*  Custom domain profile page (no app shell, no navigation)          */
/* ------------------------------------------------------------------ */

function useProfileByDomain(domain: string) {
  return useQuery({
    queryKey: ["profile", "by-domain", domain],
    queryFn: () => apiClient.profile.getByDomain(domain).then((res) => res.data.data),
    enabled: !!domain,
    retry: false,
  });
}

function ProfileHeader({ profile }: { profile: PublicProfile }) {
  return (
    <header className="relative border-b border-[var(--border-subtle)] bg-[var(--bg-surface)] overflow-hidden">
      <div
        className="pointer-events-none absolute inset-0"
        style={{
          background:
            "linear-gradient(135deg, var(--accent, var(--accent-subtle)) 0%, transparent 60%)",
          opacity: 0.15,
        }}
        aria-hidden="true"
      />
      <div className="relative mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="flex items-start gap-4">
          {profile.avatar_url ? (
            <img
              src={profile.avatar_url}
              alt={profile.name}
              className="h-20 w-20 shrink-0 rounded-full object-cover"
            />
          ) : (
            <div
              className="flex h-20 w-20 shrink-0 items-center justify-center rounded-full text-2xl font-semibold"
              style={{
                backgroundColor: "color-mix(in srgb, var(--accent, #6366f1) 15%, transparent)",
                color: "var(--accent, var(--accent-default))",
              }}
            >
              {profile.name.charAt(0).toUpperCase()}
            </div>
          )}
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="text-2xl font-bold text-[var(--text-primary)]">{profile.name}</h1>
              {profile.open_to_work === "actively_looking" && (
                <span className="inline-flex items-center rounded-full bg-[var(--color-success)]/10 px-2.5 py-0.5 text-xs font-medium text-[var(--color-success)]">
                  Actively looking for work
                </span>
              )}
              {profile.open_to_work === "open" && (
                <span className="inline-flex items-center rounded-full bg-[var(--color-info)]/10 px-2.5 py-0.5 text-xs font-medium text-[var(--color-info)]">
                  Open to opportunities
                </span>
              )}
              {profile.is_staff && (
                <span
                  className="inline-flex items-center gap-1 rounded-full bg-[var(--accent-default)]/10 px-2.5 py-0.5 text-xs font-medium"
                  style={{ color: "var(--accent, var(--accent-bright))" }}
                >
                  <ShieldIcon width={12} height={12} aria-hidden="true" />
                  TrackMy Staff
                </span>
              )}
              {profile.is_supporter && (
                <span className="inline-flex items-center gap-1 rounded-full bg-[var(--accent-warm)]/10 px-2.5 py-0.5 text-xs font-medium text-[var(--accent-warm-text)]">
                  <HeartIcon width={12} height={12} fill="currentColor" aria-hidden="true" />
                  Supporter
                </span>
              )}
            </div>
            {profile.headline && (
              <p className="mt-0.5 text-sm text-[var(--text-tertiary)]">{profile.headline}</p>
            )}
            {profile.bio && (
              <p className="mt-1 text-sm text-[var(--text-secondary)]">{profile.bio}</p>
            )}
            {profile.location && (
              <p className="mt-1.5 flex items-center gap-1 text-sm text-[var(--text-tertiary)]">
                <svg
                  className="h-3.5 w-3.5"
                  fill="none"
                  viewBox="0 0 24 24"
                  strokeWidth={1.5}
                  stroke="currentColor"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"
                  />
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z"
                  />
                </svg>
                {profile.location}
              </p>
            )}
            <div className="mt-2 flex flex-wrap items-center gap-3">
              <div className="flex items-center gap-2">
                <span
                  className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium"
                  style={{
                    backgroundColor: "color-mix(in srgb, var(--accent, #6366f1) 15%, transparent)",
                    color: "var(--accent, var(--accent-bright))",
                  }}
                >
                  Level {profile.level}
                </span>
                <span className="text-xs text-[var(--text-tertiary)]">{profile.level_title}</span>
              </div>
              {/* Verified social links */}
              {profile.linked_accounts && profile.linked_accounts.length > 0 && (
                <div className="flex items-center gap-3">
                  {profile.linked_accounts.map((account) => {
                    if (account.provider === "linkedin") {
                      return (
                        <a
                          key="linkedin"
                          href={account.profile_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="group relative"
                          title="Verified LinkedIn"
                          aria-label="Verified LinkedIn profile"
                        >
                          <svg
                            className="h-5 w-5 text-[#0a66c2]"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                            aria-hidden="true"
                          >
                            <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 0 1-2.063-2.065 2.064 2.064 0 1 1 2.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z" />
                          </svg>
                        </a>
                      );
                    }
                    if (account.provider === "github") {
                      return (
                        <a
                          key="github"
                          href={account.profile_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="group relative"
                          title="Verified GitHub"
                          aria-label="Verified GitHub profile"
                        >
                          <svg
                            className="h-5 w-5 text-[var(--text-primary)]"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                            aria-hidden="true"
                          >
                            <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
                          </svg>
                        </a>
                      );
                    }
                    if (account.provider === "website") {
                      return (
                        <a
                          key="website"
                          href={account.profile_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="group relative"
                          title="Verified website"
                          aria-label="Verified personal website"
                        >
                          <svg
                            className="h-5 w-5"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth={1.5}
                            stroke="currentColor"
                            style={{ color: "var(--accent, var(--accent-default))" }}
                            aria-hidden="true"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5a17.92 17.92 0 0 1-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418"
                            />
                          </svg>
                        </a>
                      );
                    }
                    return null;
                  })}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}

export default function CustomDomainProfile() {
  const domain = window.location.hostname;
  const { data: profile, isLoading, isError } = useProfileByDomain(domain);

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
      <div role="alert" className="flex min-h-screen flex-col items-center justify-center gap-4 bg-[var(--bg-base)] p-8">
        <p className="text-6xl font-bold text-[var(--text-tertiary)]" aria-hidden="true">404</p>
        <h1 className="text-xl font-semibold text-[var(--text-primary)]">
          This domain is not connected to a profile
        </h1>
        <p className="text-sm text-[var(--text-secondary)]">
          The domain <strong>{domain}</strong> has not been linked to any trackmy.career account.
        </p>
        <a
          href="https://trackmy.career"
          className="mt-2 inline-flex items-center rounded-[var(--radius-md)] bg-[var(--accent-default)] px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-[var(--accent-bright)]"
        >
          Visit trackmy.career
        </a>
      </div>
    );
  }

  const accentColour = profile.accent_colour;

  const hasBadges = profile.badges && profile.badges.length > 0;
  const hasJobs = profile.jobs && profile.jobs.length > 0;
  const hasCerts = profile.certifications && profile.certifications.length > 0;
  const hasSkills = profile.skills && profile.skills.length > 0;
  const hasWins = profile.wins && profile.wins.length > 0;

  return (
    <div
      className="min-h-screen bg-[var(--bg-base)]"
      style={accentColour ? ({ "--accent": accentColour } as React.CSSProperties) : undefined}
    >
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-[var(--radius-md)] focus:bg-[var(--bg-elevated)] focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-[var(--text-primary)] focus:shadow-lg"
      >
        Skip to content
      </a>
      <ProfileHeader profile={profile} />

      <main id="main-content" className="mx-auto max-w-3xl space-y-8 px-4 py-8 sm:px-6 lg:px-8">
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
          <ProfileSection title="Experience" icon={<BriefcaseIcon width={16} height={16} />}>
            <JobTimeline jobs={profile.jobs!} />
          </ProfileSection>
        )}

        {/* Certifications */}
        {hasCerts && (
          <ProfileSection title="Certifications" icon={<AwardIcon width={16} height={16} />}>
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

      </main>

      {/* Footer */}
      <footer className="mx-auto max-w-3xl border-t border-[var(--border-subtle)] px-4 pt-6 text-center sm:px-6 lg:px-8">
        <a
          href="https://trackmy.career"
          target="_blank"
          rel="noopener noreferrer"
          className="text-xs text-[var(--text-tertiary)] transition-colors hover:text-[var(--text-secondary)]"
        >
          Built with trackmy
          <span style={{ color: "var(--accent, var(--accent-default))" }}>.</span>career
        </a>
      </footer>
    </div>
  );
}
