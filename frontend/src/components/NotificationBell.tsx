import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router";
import { BellIcon } from "@/components/icons";
import {
  useNotificationsQuery,
  useUnreadCountQuery,
  useMarkNotificationReadMutation,
  useMarkAllNotificationsReadMutation,
} from "@/hooks/queries/useNotificationsQuery";
import type { Notification } from "@/types";

function timeAgo(iso: string): string {
  const then = new Date(iso).getTime();
  const seconds = Math.floor((Date.now() - then) / 1000);
  if (seconds < 60) return "just now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  return new Date(iso).toLocaleDateString("en-GB", {
    day: "numeric",
    month: "short",
  });
}

function destinationFor(n: Notification): string {
  if (n.related_certification_id) {
    return `/certifications/${n.related_certification_id}/edit`;
  }
  return "/notifications";
}

export function NotificationBell() {
  const [open, setOpen] = useState(false);
  const [liveMessage, setLiveMessage] = useState("");
  const containerRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const prevUnread = useRef<number | null>(null);
  const navigate = useNavigate();

  const { data: unread = 0, isLoading: unreadLoading } = useUnreadCountQuery();
  const { data, isLoading } = useNotificationsQuery({ limit: 6 });
  const markRead = useMarkNotificationReadMutation();
  const markAllRead = useMarkAllNotificationsReadMutation();

  const notifications = data?.notifications ?? [];

  // Announce only genuine increases (new notifications arriving during the
  // session), not the initial load from zero.
  useEffect(() => {
    if (unreadLoading) return;
    if (prevUnread.current !== null && unread > prevUnread.current) {
      setLiveMessage(`${unread} unread notification${unread === 1 ? "" : "s"}`);
    }
    prevUnread.current = unread;
  }, [unread, unreadLoading]);

  // Close and return focus to the trigger (used for keyboard close paths).
  const closeAndRestoreFocus = () => {
    setOpen(false);
    triggerRef.current?.focus();
  };

  useEffect(() => {
    if (!open) return;

    function handlePointer(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setOpen(false);
      }
    }
    function handleKey(event: KeyboardEvent) {
      if (event.key === "Escape") closeAndRestoreFocus();
    }

    document.addEventListener("mousedown", handlePointer);
    document.addEventListener("keydown", handleKey);
    return () => {
      document.removeEventListener("mousedown", handlePointer);
      document.removeEventListener("keydown", handleKey);
    };
  }, [open]);

  const handleOpen = (notification: Notification) => {
    if (!notification.read_at) markRead.mutate(notification.id);
    setOpen(false);
    navigate(destinationFor(notification));
  };

  const unreadLabel = unread > 0 ? `, ${unread} unread` : "";

  return (
    <div ref={containerRef} className="relative">
      <span className="sr-only" role="status" aria-live="polite">
        {liveMessage}
      </span>

      <button
        ref={triggerRef}
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        className="relative flex min-h-11 min-w-11 items-center justify-center rounded-[var(--radius-sm)]
          p-1.5 text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)]
          focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-default)]"
        aria-label={`Notifications${unreadLabel}`}
        aria-haspopup="true"
        aria-expanded={open}
      >
        <BellIcon width={20} height={20} />
        {unread > 0 && (
          <span
            className="absolute right-1.5 top-1.5 flex h-4 min-w-4 items-center justify-center
              rounded-full bg-[var(--color-error)] px-1 text-[10px] font-semibold leading-none text-white"
            aria-hidden="true"
          >
            {unread > 9 ? "9+" : unread}
          </span>
        )}
      </button>

      {open && (
        <div
          className="absolute right-0 z-40 mt-2 w-80 max-w-[calc(100vw-2rem)] overflow-hidden
            rounded-[var(--radius-lg)] border border-[var(--border-default)] bg-[var(--bg-surface)]
            shadow-lg"
        >
          <div className="flex items-center justify-between border-b border-[var(--border-subtle)] px-4 py-3">
            <span className="text-sm font-semibold text-[var(--text-primary)]">Notifications</span>
            {unread > 0 && (
              <button
                type="button"
                onClick={() => {
                  markAllRead.mutate();
                  closeAndRestoreFocus();
                }}
                disabled={markAllRead.isPending}
                className="rounded-[var(--radius-sm)] px-1.5 py-1 text-xs font-medium
                  text-[var(--accent-default)] hover:text-[var(--accent-bright)] focus-visible:outline-none
                  focus-visible:ring-2 focus-visible:ring-[var(--accent-default)] disabled:opacity-50"
              >
                Mark all read
              </button>
            )}
          </div>

          <div className="max-h-80 overflow-y-auto">
            {isLoading ? (
              <p
                className="px-4 py-6 text-center text-sm text-[var(--text-tertiary)]"
                role="status"
              >
                Loading…
              </p>
            ) : notifications.length === 0 ? (
              <p className="px-4 py-8 text-center text-sm text-[var(--text-tertiary)]">
                You're all caught up.
              </p>
            ) : (
              <ul className="divide-y divide-[var(--border-subtle)]">
                {notifications.map((n) => (
                  <li key={n.id}>
                    <button
                      type="button"
                      onClick={() => handleOpen(n)}
                      className={[
                        "flex w-full flex-col gap-0.5 px-4 py-3 text-left transition-colors",
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
                        <span className="text-sm font-medium text-[var(--text-primary)]">
                          {n.title}
                        </span>
                        {!n.read_at && <span className="sr-only">(unread)</span>}
                      </span>
                      <span className="line-clamp-2 text-xs text-[var(--text-secondary)]">
                        {n.body}
                      </span>
                      <span className="text-[11px] text-[var(--text-tertiary)]">
                        {timeAgo(n.created_at)}
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>

          <div className="border-t border-[var(--border-subtle)] px-4 py-2 text-center">
            <button
              type="button"
              onClick={() => {
                setOpen(false);
                navigate("/notifications");
              }}
              className="rounded-[var(--radius-sm)] px-1.5 py-1 text-xs font-medium
                text-[var(--accent-default)] hover:text-[var(--accent-bright)] focus-visible:outline-none
                focus-visible:ring-2 focus-visible:ring-[var(--accent-default)]"
            >
              View all notifications
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
