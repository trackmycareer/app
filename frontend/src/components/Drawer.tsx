import { useEffect, useRef, useId } from "react";
import type { ReactNode } from "react";
import { createPortal } from "react-dom";
import { CloseIcon } from "@/components/icons";

interface DrawerProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  footer?: ReactNode;
  // When another dialog (e.g. a confirmation modal) is layered on top, set this
  // so Escape closes only the top dialog instead of the drawer underneath it.
  disableEscape?: boolean;
}

export function Drawer({
  open,
  onClose,
  title,
  children,
  footer,
  disableEscape = false,
}: DrawerProps) {
  const dialogRef = useRef<HTMLDivElement>(null);
  const overlayRef = useRef<HTMLDivElement>(null);
  const previousActiveElement = useRef<Element | null>(null);
  const titleId = useId();

  // Trap Tab focus within the drawer, skipping disabled/hidden controls so a
  // disabled footer button can never become the trap boundary.
  useEffect(() => {
    if (!open || !dialogRef.current) return;

    const dialog = dialogRef.current;
    const handleTab = (e: KeyboardEvent) => {
      if (e.key !== "Tab") return;

      const focusable = dialog.querySelectorAll<HTMLElement>(
        'button:not(:disabled), [href], input:not(:disabled), select:not(:disabled), textarea:not(:disabled), [tabindex]:not([tabindex="-1"])',
      );
      if (focusable.length === 0) return;

      const first = focusable[0];
      const last = focusable[focusable.length - 1];

      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    };

    dialog.addEventListener("keydown", handleTab);
    return () => dialog.removeEventListener("keydown", handleTab);
  }, [open]);

  // Open/close transition only: capture the trigger, lock scroll, move focus in,
  // and restore focus out. Keyed on `open` alone so a parent re-render (e.g. a
  // background refetch) never steals focus back to the close button.
  useEffect(() => {
    if (!open) return;

    previousActiveElement.current = document.activeElement;
    document.body.style.overflow = "hidden";
    const raf = requestAnimationFrame(() => {
      dialogRef.current?.querySelector<HTMLElement>("button, input")?.focus();
    });

    return () => {
      cancelAnimationFrame(raf);
      document.body.style.overflow = "";
      if (previousActiveElement.current instanceof HTMLElement) {
        previousActiveElement.current.focus();
      }
    };
  }, [open]);

  // Escape to close, in its own effect so onClose identity changes do not
  // disturb focus management above.
  useEffect(() => {
    if (!open || disableEscape) return;

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handleEscape);
    return () => document.removeEventListener("keydown", handleEscape);
  }, [open, onClose, disableEscape]);

  if (!open) return null;

  return createPortal(
    <div
      ref={overlayRef}
      className="animate-fade-in fixed inset-0 z-50 flex justify-end bg-black/60"
      onClick={(e) => {
        if (e.target === overlayRef.current) onClose();
      }}
      role="presentation"
    >
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        className="animate-slide-over flex h-full w-full max-w-md flex-col border-l
          border-[var(--border-default)] bg-[var(--bg-surface)] shadow-xl"
      >
        <div
          className="flex items-center justify-between border-b border-[var(--border-subtle)]
            px-5 py-4"
        >
          <h2 id={titleId} className="text-lg font-semibold text-[var(--text-primary)]">
            {title}
          </h2>
          <button
            onClick={onClose}
            className="flex min-h-11 min-w-11 items-center justify-center rounded-[var(--radius-sm)]
              p-2 text-[var(--text-tertiary)] transition-colors hover:bg-[var(--bg-hover)]
              hover:text-[var(--text-primary)]"
            aria-label="Close"
          >
            <CloseIcon width={18} height={18} />
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-5">{children}</div>
        {footer && <div className="border-t border-[var(--border-subtle)] p-4">{footer}</div>}
      </div>
    </div>,
    document.body,
  );
}
