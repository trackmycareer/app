import { useState } from "react";
import { useNavigate } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { BellIcon, SpinnerIcon } from "@/components/icons";
import {
  useNotificationsQuery,
  useMarkNotificationReadMutation,
  useMarkAllNotificationsReadMutation,
} from "@/hooks/queries/useNotificationsQuery";
import type { Notification } from "@/types";

const PAGE_SIZE = 20;

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString("en-GB", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function destinationFor(n: Notification): string | null {
  if (n.related_certification_id) {
    return `/certifications/${n.related_certification_id}/edit`;
  }
  return null;
}

export default function NotificationsList() {
  const [page, setPage] = useState(0);
  const navigate = useNavigate();

  const { data, isLoading, isError } = useNotificationsQuery({
    limit: PAGE_SIZE,
    offset: page * PAGE_SIZE,
  });
  const markRead = useMarkNotificationReadMutation();
  const markAllRead = useMarkAllNotificationsReadMutation();

  const notifications = data?.notifications ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  const hasUnread = notifications.some((n) => !n.read_at);

  const handleOpen = (n: Notification) => {
    if (!n.read_at) markRead.mutate(n.id);
    const dest = destinationFor(n);
    if (dest) navigate(dest);
  };

  return (
    <>
      <Topbar title="Notifications" />
      <div className="mx-auto max-w-2xl space-y-4 p-4 lg:p-6">
        <div className="flex items-center justify-between">
          <p className="text-sm text-[var(--text-tertiary)]">
            {total > 0 ? `${total} notification${total === 1 ? "" : "s"}` : "No notifications yet"}
          </p>
          {hasUnread && (
            <Button
              type="button"
              variant="secondary"
              size="sm"
              onClick={() => markAllRead.mutate()}
              loading={markAllRead.isPending}
            >
              Mark all read
            </Button>
          )}
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center py-20" role="status">
            <SpinnerIcon
              width={28}
              height={28}
              className="animate-spin text-[var(--accent-default)]"
              aria-hidden="true"
            />
            <span className="sr-only">Loading notifications…</span>
          </div>
        ) : isError ? (
          <div
            className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
              bg-[var(--color-error)]/5 p-4 text-center text-sm text-[var(--color-error)]"
            role="alert"
          >
            Failed to load notifications. Please try again later.
          </div>
        ) : notifications.length === 0 ? (
          <div
            className="flex flex-col items-center gap-3 rounded-[var(--radius-xl)] border
              border-[var(--border-default)] bg-[var(--bg-surface)] px-4 py-16 text-center"
          >
            <BellIcon width={28} height={28} className="text-[var(--text-tertiary)]" />
            <p className="text-sm text-[var(--text-secondary)]">You're all caught up.</p>
            <p className="text-xs text-[var(--text-tertiary)]">
              Renewal reminders for your certifications will appear here.
            </p>
          </div>
        ) : (
          <ul
            className="divide-y divide-[var(--border-subtle)] overflow-hidden rounded-[var(--radius-xl)]
              border border-[var(--border-default)] bg-[var(--bg-surface)]"
          >
            {notifications.map((n) => {
              const clickable = destinationFor(n) !== null;
              return (
                <li key={n.id}>
                  <button
                    type="button"
                    onClick={() => handleOpen(n)}
                    aria-disabled={!clickable && !!n.read_at}
                    className={[
                      "flex w-full flex-col gap-1 px-4 py-3.5 text-left transition-colors",
                      "hover:bg-[var(--bg-hover)] focus-visible:outline-none focus-visible:ring-2",
                      "focus-visible:ring-inset focus-visible:ring-[var(--accent-default)]",
                      n.read_at ? "" : "bg-[var(--accent-default)]/5",
                    ].join(" ")}
                  >
                    <span className="flex items-center gap-2">
                      {!n.read_at && (
                        <span
                          className="h-1.5 w-1.5 shrink-0 rounded-full bg-[var(--accent-default)]"
                          aria-hidden="true"
                        />
                      )}
                      <span className="text-sm font-semibold text-[var(--text-primary)]">
                        {n.title}
                      </span>
                      {!n.read_at && <span className="sr-only">(unread)</span>}
                    </span>
                    <span className="text-sm text-[var(--text-secondary)]">{n.body}</span>
                    <span className="text-xs text-[var(--text-tertiary)]">
                      {formatDate(n.created_at)}
                    </span>
                  </button>
                </li>
              );
            })}
          </ul>
        )}

        {totalPages > 1 && (
          <div className="flex items-center justify-between pt-2">
            <Button
              type="button"
              variant="secondary"
              size="sm"
              disabled={page === 0}
              onClick={() => setPage((p) => Math.max(0, p - 1))}
            >
              Previous
            </Button>
            <span className="text-xs text-[var(--text-tertiary)]" role="status" aria-live="polite">
              Page {page + 1} of {totalPages}
            </span>
            <Button
              type="button"
              variant="secondary"
              size="sm"
              disabled={page >= totalPages - 1}
              onClick={() => setPage((p) => p + 1)}
            >
              Next
            </Button>
          </div>
        )}
      </div>
    </>
  );
}
