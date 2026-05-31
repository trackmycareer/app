import { useEffect } from "react";
import { createPortal } from "react-dom";
import type { Award } from "@/types";
import { ConfettiPiece } from "@/components/ConfettiPiece";

export function LevelUpModal({ award, onDismiss }: { award: Award; onDismiss: () => void }) {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onDismiss();
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [onDismiss]);

  return createPortal(
    <div className="fixed inset-0 z-[100]" role="presentation">
      {/* Backdrop overlay button */}
      <button
        className="absolute inset-0 h-full w-full bg-black/70"
        onClick={onDismiss}
        aria-label="Close celebration"
      />

      {/* Confetti layer */}
      <div className="pointer-events-none fixed inset-0 overflow-hidden" aria-hidden="true">
        {Array.from({ length: 50 }).map((_, i) => (
          <ConfettiPiece key={i} index={i} />
        ))}
      </div>

      {/* Dialog */}
      <div className="pointer-events-none flex h-full w-full items-center justify-center p-4">
        <div
          className="level-up-entrance pointer-events-auto relative z-10 w-full max-w-sm
            rounded-[var(--radius-2xl)] border border-[var(--accent-default)]
            bg-[var(--bg-surface)] p-8 text-center shadow-2xl"
          role="dialog"
          aria-modal="true"
          aria-label="Level up celebration"
        >
          <div className="mb-3 text-5xl" aria-hidden="true">
            🎉
          </div>
          <h2 className="mb-1 text-2xl font-bold text-[var(--text-primary)]">
            Level {award.level}!
          </h2>
          <p className="mb-2 text-lg font-medium text-[var(--accent-bright)]">
            {award.level_title}
          </p>
          <p className="mb-6 text-sm text-[var(--text-secondary)]">
            Congratulations, you have reached a new level!
          </p>
          <button
            onClick={onDismiss}
            className="rounded-[var(--radius-md)] bg-[var(--accent-default)] px-6 py-2
              text-sm font-medium text-white transition-colors hover:bg-[var(--accent-bright)]"
          >
            Continue
          </button>
        </div>
      </div>
    </div>,
    document.body,
  );
}
