import { useEffect, useRef, useCallback, useId } from "react";
import type { ReactNode } from "react";
import { createPortal } from "react-dom";
import { CloseIcon } from "@/components/icons";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
}

export function Modal({ open, onClose, title, children }: ModalProps) {
  const dialogRef = useRef<HTMLDivElement>(null);
  const overlayRef = useRef<HTMLDivElement>(null);
  const previousActiveElement = useRef<Element | null>(null);
  const titleId = useId();

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    },
    [onClose],
  );

  useEffect(() => {
    if (!open || !dialogRef.current) return;

    const dialog = dialogRef.current;
    const handleTab = (e: KeyboardEvent) => {
      if (e.key !== "Tab") return;

      const focusable = dialog.querySelectorAll<HTMLElement>(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
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

  useEffect(() => {
    if (open) {
      previousActiveElement.current = document.activeElement;
      document.addEventListener("keydown", handleKeyDown);
      document.body.style.overflow = "hidden";
      requestAnimationFrame(() => {
        dialogRef.current?.querySelector<HTMLElement>("button, input")?.focus();
      });
    }
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = "";
      if (previousActiveElement.current instanceof HTMLElement) {
        previousActiveElement.current.focus();
      }
    };
  }, [open, handleKeyDown]);

  if (!open) return null;

  return createPortal(
    <div
      ref={overlayRef}
      className="animate-fade-in fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
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
        className="animate-scale-in w-full max-w-lg rounded-[var(--radius-xl)] border
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
            className="flex items-center justify-center rounded-[var(--radius-sm)] p-2 min-h-11 min-w-11
              text-[var(--text-tertiary)] transition-colors hover:bg-[var(--bg-hover)]
              hover:text-[var(--text-primary)]"
            aria-label="Close"
          >
            <CloseIcon width={18} height={18} />
          </button>
        </div>
        <div className="p-5">{children}</div>
      </div>
    </div>,
    document.body,
  );
}
