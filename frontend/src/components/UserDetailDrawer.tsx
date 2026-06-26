import { useQuery } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { Drawer } from "@/components/Drawer";
import { Button } from "@/components/Button";
import {
  SpinnerIcon,
  UserCircleIcon,
  CheckCircleIcon,
  AlertTriangleIcon,
  ShieldCheckIcon,
  ShieldIcon,
  HeartIcon,
  TrophyIcon,
  BriefcaseIcon,
  AwardIcon,
  ZapIcon,
  StarIcon,
  LinkIcon,
  ExternalLinkIcon,
  FlameIcon,
  VerifiedBadgeIcon,
} from "@/components/icons";
import { apiClient } from "@/lib/api";
import type { AdminUserDetail, User } from "@/types";

interface UserDetailDrawerProps {
  userId: string | null;
  open: boolean;
  onClose: () => void;
  isSelf: boolean;
  onToggleAdmin: (user: User) => void;
  onDelete: (user: User) => void;
  // True while a confirmation modal sits on top of the drawer.
  confirmOpen?: boolean;
}

const detailQueryKey = (id: string) => ["admin", "users", "detail", id] as const;

function formatDate(value: string | null | undefined): string {
  if (!value) return "Not set";
  return new Date(value).toLocaleDateString("en-GB", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}

function supporterLabel(user: User): string {
  if (user.is_subscriber) return "Subscriber";
  if (user.is_one_time_supporter) return "One-time supporter";
  return "Free";
}

const openToWorkLabels: Record<string, string> = {
  not_looking: "Not looking",
  open: "Open to offers",
  actively_looking: "Actively looking",
};

export function UserDetailDrawer({
  userId,
  open,
  onClose,
  isSelf,
  onToggleAdmin,
  onDelete,
  confirmOpen = false,
}: UserDetailDrawerProps) {
  const { data, isLoading, isError } = useQuery({
    queryKey: userId ? detailQueryKey(userId) : ["admin", "users", "detail", "none"],
    queryFn: () => apiClient.admin.users.getById(userId!).then((r) => r.data.data),
    enabled: open && !!userId,
  });

  const footer = data ? (
    <div className="space-y-2">
      {data.user.username ? (
        <a
          href={`/u/${data.user.username}`}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex w-full items-center justify-center gap-1.5
            rounded-[var(--radius-md)] border border-[var(--border-default)]
            bg-[var(--bg-elevated)] px-4 py-2 text-sm font-medium text-[var(--text-primary)]
            transition-colors hover:bg-[var(--bg-hover)] focus-visible:outline-none
            focus-visible:ring-2 focus-visible:ring-[var(--border-strong)]
            focus-visible:ring-offset-2 focus-visible:ring-offset-[var(--bg-base)]"
          aria-label={`View public profile for ${data.user.name} (opens in a new tab)`}
        >
          <ExternalLinkIcon width={15} height={15} aria-hidden="true" />
          View public profile
        </a>
      ) : (
        <p className="text-xs text-[var(--text-tertiary)]">
          No public profile (this user has not set a username).
        </p>
      )}
      <div className="flex gap-2">
        <Button
          variant="secondary"
          size="sm"
          onClick={() => onToggleAdmin(data.user)}
          disabled={isSelf}
          aria-label={
            data.user.is_admin
              ? `Remove admin from ${data.user.name}`
              : `Make admin for ${data.user.name}`
          }
        >
          {data.user.is_admin ? "Remove admin" : "Make admin"}
        </Button>
        <Button
          variant="danger"
          size="sm"
          onClick={() => onDelete(data.user)}
          disabled={isSelf}
          aria-label={`Delete ${data.user.name}`}
        >
          Delete
        </Button>
      </div>
      {isSelf && (
        <p className="text-xs text-[var(--text-tertiary)]">
          You cannot change your own admin status or delete your own account.
        </p>
      )}
    </div>
  ) : undefined;

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="User details"
      footer={footer}
      disableEscape={confirmOpen}
    >
      {isLoading && (
        <div className="flex items-center justify-center py-16" role="status">
          <SpinnerIcon
            width={28}
            height={28}
            className="animate-spin text-[var(--accent-default)]"
            aria-hidden="true"
          />
          <span className="sr-only">Loading user details...</span>
        </div>
      )}

      {isError && (
        <div
          className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
            bg-[var(--color-error)]/5 p-4 text-center text-sm text-[var(--color-error)]"
          role="alert"
        >
          Failed to load user details. Please try again later.
        </div>
      )}

      {data && <DetailBody detail={data} />}
    </Drawer>
  );
}

function DetailBody({ detail }: { detail: AdminUserDetail }) {
  const { user, counts, gamification, linked_accounts, recent } = detail;

  return (
    <div className="space-y-7">
      {/* Identity */}
      <div className="flex items-start gap-3">
        {user.avatar_url ? (
          <img
            src={user.avatar_url}
            alt=""
            className="h-12 w-12 shrink-0 rounded-full object-cover"
          />
        ) : (
          <UserCircleIcon
            width={48}
            height={48}
            className="shrink-0 text-[var(--text-tertiary)]"
            aria-hidden="true"
          />
        )}
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <h3 className="truncate text-base font-semibold text-[var(--text-primary)]">
              {user.name}
            </h3>
            {user.is_admin ? (
              <span
                className="inline-flex shrink-0 items-center rounded-full
                  bg-[var(--accent-default)]/10 px-2 py-0.5 text-xs font-medium
                  text-[var(--accent-default)]"
              >
                Admin
              </span>
            ) : (
              <span className="shrink-0 text-xs text-[var(--text-tertiary)]">User</span>
            )}
          </div>
          <p className="truncate text-sm text-[var(--text-secondary)]">{user.email}</p>
          {user.headline && (
            <p className="mt-0.5 truncate text-sm text-[var(--text-secondary)]">{user.headline}</p>
          )}
          <div className="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-[var(--text-tertiary)]">
            <span
              className="inline-flex items-center rounded-full bg-[var(--bg-elevated)]
                px-2 py-0.5 font-medium text-[var(--text-secondary)]"
            >
              {user.provider}
            </span>
            {user.username && <span>@{user.username}</span>}
            {user.location && <span>{user.location}</span>}
          </div>
        </div>
      </div>

      {user.bio && (
        <p className="text-sm leading-relaxed text-[var(--text-secondary)]">{user.bio}</p>
      )}

      {/* Account & status */}
      <Section title="Account">
        <dl className="space-y-2.5">
          <StatusRow
            label="Email verified"
            value={
              user.email_verified ? (
                <Badge tone="success" icon={<CheckCircleIcon width={13} height={13} />}>
                  {`Verified ${formatDate(user.email_verified_at)}`}
                </Badge>
              ) : (
                <Badge tone="warning" icon={<AlertTriangleIcon width={13} height={13} />}>
                  Not verified
                </Badge>
              )
            }
          />
          <StatusRow
            label="Two-factor auth"
            value={
              user.mfa_enabled ? (
                <Badge tone="success" icon={<ShieldCheckIcon width={13} height={13} />}>
                  Enabled
                </Badge>
              ) : (
                <Badge tone="neutral" icon={<ShieldIcon width={13} height={13} />}>
                  Disabled
                </Badge>
              )
            }
          />
          <StatusRow
            label="Open to work"
            value={
              <span className="text-[var(--text-secondary)]">
                {openToWorkLabels[user.open_to_work] ?? user.open_to_work}
              </span>
            }
          />
          <StatusRow
            label="Plan"
            value={
              <Badge
                tone={user.is_subscriber || user.is_one_time_supporter ? "success" : "neutral"}
                icon={<HeartIcon width={13} height={13} />}
              >
                {supporterLabel(user)}
                {user.supporter_since ? ` since ${formatDate(user.supporter_since)}` : ""}
              </Badge>
            }
          />
          <StatusRow
            label="Newsletter"
            value={
              <span className="text-[var(--text-secondary)]">
                {user.newsletter_opt_in
                  ? `Opted in ${formatDate(user.newsletter_opt_in_at)}`
                  : "Not subscribed"}
              </span>
            }
          />
          <StatusRow
            label="Member since"
            value={
              <span className="text-[var(--text-secondary)]">{formatDate(user.created_at)}</span>
            }
          />
          <StatusRow
            label="Last updated"
            value={
              <span className="text-[var(--text-secondary)]">{formatDate(user.updated_at)}</span>
            }
          />
        </dl>
      </Section>

      {/* Activity */}
      <Section title="Activity">
        <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-3">
          <Stat icon={<TrophyIcon width={15} height={15} />} label="Wins" value={counts.wins} />
          <Stat icon={<BriefcaseIcon width={15} height={15} />} label="Jobs" value={counts.jobs} />
          <Stat
            icon={<AwardIcon width={15} height={15} />}
            label="Certs"
            value={counts.certifications}
            hint={`${counts.certifications_passed} passed`}
          />
          <Stat icon={<ZapIcon width={15} height={15} />} label="Skills" value={counts.skills} />
          <Stat
            icon={<LinkIcon width={15} height={15} />}
            label="Evidence"
            value={counts.evidence}
          />
          <Stat icon={<StarIcon width={15} height={15} />} label="Badges" value={counts.badges} />
        </div>
      </Section>

      {/* Gamification */}
      <Section title="Gamification">
        <div className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)] p-4">
          <div className="flex items-baseline justify-between">
            <div>
              <p className="text-sm font-semibold text-[var(--text-primary)]">
                Level {gamification.level}
              </p>
              <p className="text-xs text-[var(--text-tertiary)]">{gamification.level_title}</p>
            </div>
            <p className="text-sm text-[var(--text-secondary)]">{gamification.total_points} pts</p>
          </div>
          <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-[var(--text-tertiary)]">
            <span className="inline-flex items-center gap-1">
              <FlameIcon width={13} height={13} aria-hidden="true" />
              {gamification.current_streak}-day streak
            </span>
            <span>Best {gamification.longest_streak} days</span>
            <span>
              Last active{" "}
              {gamification.last_active_on ? formatDate(gamification.last_active_on) : "Never"}
            </span>
          </div>
          <p className="mt-2 text-[11px] text-[var(--text-tertiary)]">
            Last active reflects recorded activity (wins, jobs, certs, skills), not sign-ins.
          </p>
        </div>
      </Section>

      {/* Linked accounts */}
      <Section title="Linked accounts">
        {linked_accounts.length === 0 ? (
          <EmptyHint>No linked accounts.</EmptyHint>
        ) : (
          <ul className="space-y-2">
            {linked_accounts.map((acc) => (
              <li key={acc.provider} className="flex items-center justify-between gap-2 text-sm">
                <a
                  href={acc.profile_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex min-w-0 items-center gap-1.5 text-[var(--text-secondary)]
                    hover:text-[var(--accent-default)] hover:underline"
                  aria-label={`${acc.provider} profile (opens in a new tab)`}
                >
                  <ExternalLinkIcon width={13} height={13} aria-hidden="true" />
                  <span className="truncate capitalize">{acc.provider}</span>
                </a>
                {acc.verified && (
                  <span className="inline-flex shrink-0 items-center gap-1 text-xs text-[var(--text-secondary)]">
                    <VerifiedBadgeIcon
                      width={13}
                      height={13}
                      className="text-[var(--color-success)]"
                      aria-hidden="true"
                    />
                    Verified
                  </span>
                )}
              </li>
            ))}
          </ul>
        )}
      </Section>

      {/* Recent activity */}
      <Section title="Recent activity">
        <div className="space-y-4">
          <RecentList title="Wins" empty="No wins yet.">
            {recent.wins.map((w) => (
              <RecentRow
                key={w.id}
                primary={w.title}
                secondary={w.category}
                meta={formatDate(w.occurred_on)}
              />
            ))}
          </RecentList>
          <RecentList title="Jobs" empty="No jobs yet.">
            {recent.jobs.map((j) => (
              <RecentRow
                key={j.id}
                primary={j.title}
                secondary={j.company}
                meta={`${formatDate(j.start_date)} to ${j.end_date ? formatDate(j.end_date) : "Present"}`}
              />
            ))}
          </RecentList>
          <RecentList title="Certifications" empty="No certifications yet.">
            {recent.certifications.map((c) => (
              <RecentRow key={c.id} primary={c.name} secondary={c.provider} meta={c.status} />
            ))}
          </RecentList>
          <RecentList title="Skills" empty="No skills yet.">
            {recent.skills.map((s) => (
              <RecentRow
                key={s.id}
                primary={s.name}
                secondary={s.category ?? undefined}
                meta={`Proficiency ${s.proficiency}`}
              />
            ))}
          </RecentList>
        </div>
      </Section>
    </div>
  );
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section>
      <h4 className="mb-3 text-xs font-medium uppercase tracking-wider text-[var(--text-tertiary)]">
        {title}
      </h4>
      {children}
    </section>
  );
}

function StatusRow({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-3 text-sm">
      <dt className="text-[var(--text-tertiary)]">{label}</dt>
      <dd className="text-right">{value}</dd>
    </div>
  );
}

type BadgeTone = "success" | "warning" | "neutral";

// Tone colours the icon only; the label always uses readable secondary text on a
// neutral chip so the status text meets contrast in both themes. Meaning is
// carried by the text, not by colour alone.
const badgeIconTone: Record<BadgeTone, string> = {
  success: "text-[var(--color-success)]",
  warning: "text-[var(--color-warning)]",
  neutral: "text-[var(--text-tertiary)]",
};

function Badge({
  tone,
  icon,
  children,
}: {
  tone: BadgeTone;
  icon?: ReactNode;
  children: ReactNode;
}) {
  return (
    <span
      className="inline-flex items-center gap-1 rounded-full bg-[var(--bg-elevated)]
        px-2 py-0.5 text-xs font-medium text-[var(--text-secondary)]"
    >
      {icon && (
        <span className={badgeIconTone[tone]} aria-hidden="true">
          {icon}
        </span>
      )}
      {children}
    </span>
  );
}

function Stat({
  icon,
  label,
  value,
  hint,
}: {
  icon: ReactNode;
  label: string;
  value: number;
  hint?: string;
}) {
  return (
    <div className="rounded-[var(--radius-md)] border border-[var(--border-subtle)] p-3">
      <div className="flex items-center gap-1.5 text-[var(--text-tertiary)]">
        <span aria-hidden="true">{icon}</span>
        <span className="text-xs">{label}</span>
      </div>
      <p className="mt-1 text-xl font-semibold text-[var(--text-primary)]">{value}</p>
      {hint && <p className="text-[11px] text-[var(--text-tertiary)]">{hint}</p>}
    </div>
  );
}

function RecentList({
  title,
  empty,
  children,
}: {
  title: string;
  empty: string;
  children: ReactNode[];
}) {
  return (
    <div>
      <p className="mb-1.5 text-xs font-medium text-[var(--text-secondary)]">{title}</p>
      {children.length === 0 ? (
        <EmptyHint>{empty}</EmptyHint>
      ) : (
        <ul className="space-y-1.5">{children}</ul>
      )}
    </div>
  );
}

function RecentRow({
  primary,
  secondary,
  meta,
}: {
  primary: string;
  secondary?: string;
  meta?: string;
}) {
  return (
    <li className="flex items-baseline justify-between gap-3 text-sm">
      <span className="min-w-0 flex-1 truncate text-[var(--text-primary)]">
        {primary}
        {secondary && <span className="text-[var(--text-tertiary)]"> · {secondary}</span>}
      </span>
      {meta && (
        <span className="shrink-0 text-xs capitalize text-[var(--text-tertiary)]">{meta}</span>
      )}
    </li>
  );
}

function EmptyHint({ children }: { children: ReactNode }) {
  return <p className="text-xs text-[var(--text-tertiary)]">{children}</p>;
}
