import { useEffect, useId, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useAuthStore } from "@/stores/auth";
import { apiClient } from "@/lib/api";
import { Button } from "@/components/Button";
import { LEGAL_LAST_UPDATED, LEGAL_URLS } from "@/lib/constants";

// Shows a mandatory prompt when a signed-in user has not accepted the current
// version of the Terms of Service and Privacy Policy. This covers users who
// pre-date the policies and re-prompts everyone when the policy version changes.
// It is deliberately not dismissable: the only way out is to accept or log out.
export function TermsConsentGate() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const user = useAuthStore((s) => s.user);
  const updateUser = useAuthStore((s) => s.updateUser);
  const logout = useAuthStore((s) => s.logout);

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const dialogRef = useRef<HTMLDivElement>(null);
  const previousFocus = useRef<HTMLElement | null>(null);
  const titleId = useId();
  const descId = useId();

  const needsConsent =
    isAuthenticated && user != null && user.terms_version !== LEGAL_LAST_UPDATED;

  useEffect(() => {
    if (!needsConsent) return;

    previousFocus.current = document.activeElement as HTMLElement | null;
    const appRoot = document.getElementById("root");
    appRoot?.setAttribute("inert", "");
    document.body.style.overflow = "hidden";

    const dialog = dialogRef.current;
    const handleTab = (e: KeyboardEvent) => {
      if (e.key !== "Tab" || !dialog) return;
      const focusable = Array.from(
        dialog.querySelectorAll<HTMLElement>('button, [href], [tabindex]:not([tabindex="-1"])'),
      ).filter((el) => !el.hasAttribute("disabled"));
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

    document.addEventListener("keydown", handleTab);
    requestAnimationFrame(() => {
      dialogRef.current?.querySelector<HTMLElement>("button")?.focus();
    });

    return () => {
      document.removeEventListener("keydown", handleTab);
      document.body.style.overflow = "";
      appRoot?.removeAttribute("inert");
      previousFocus.current?.focus?.();
    };
  }, [needsConsent]);

  if (!needsConsent) return null;

  const handleAccept = async () => {
    if (submitting) return;
    setSubmitting(true);
    setError("");
    try {
      const res = await apiClient.user.acceptTerms();
      updateUser(res.data.data);
    } catch {
      setError("Something went wrong. Please try again.");
      setSubmitting(false);
    }
  };

  const linkClass =
    "font-medium text-[var(--accent-text)] underline hover:text-[var(--accent-bright)]";
  const newTab = (
    <span className="sr-only"> (opens in a new tab)</span>
  );

  return createPortal(
    <div
      className="animate-fade-in fixed inset-0 z-[70] flex items-center justify-center bg-black/60 p-4"
      role="presentation"
    >
      <div
        ref={dialogRef}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={descId}
        className="animate-scale-in w-full max-w-md rounded-[var(--radius-xl)] border
          border-[var(--border-default)] bg-[var(--bg-surface)] p-6 shadow-xl"
      >
        <h2 id={titleId} className="text-lg font-semibold text-[var(--text-primary)]">
          We have updated our terms
        </h2>
        <p id={descId} className="mt-3 text-sm leading-6 text-[var(--text-secondary)]">
          To keep using trackmy.career, please review and accept our{" "}
          <a href={LEGAL_URLS.terms} target="_blank" rel="noopener noreferrer" className={linkClass}>
            Terms of Service{newTab}
          </a>{" "}
          and{" "}
          <a href={LEGAL_URLS.privacy} target="_blank" rel="noopener noreferrer" className={linkClass}>
            Privacy Policy{newTab}
          </a>
          . You can also read our{" "}
          <a href={LEGAL_URLS.cookies} target="_blank" rel="noopener noreferrer" className={linkClass}>
            Cookie Policy{newTab}
          </a>
          .
        </p>

        {error && (
          <p
            className="mt-4 rounded-[var(--radius-md)] bg-[var(--color-error)]/10 px-3 py-2 text-sm
              text-[var(--color-error)]"
            role="alert"
          >
            {error}
          </p>
        )}

        <div className="mt-6 flex flex-col gap-2">
          <Button onClick={handleAccept} loading={submitting} className="w-full">
            I agree and continue
          </Button>
          <button
            type="button"
            onClick={() => logout()}
            disabled={submitting}
            className="rounded-[var(--radius-md)] py-2 text-center text-sm text-[var(--text-tertiary)]
              transition-colors hover:text-[var(--text-secondary)] disabled:opacity-60"
          >
            Log out instead
          </button>
        </div>
      </div>
    </div>,
    document.body,
  );
}
