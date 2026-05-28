import { useEffect, useState, useCallback } from "react";
import { createPortal } from "react-dom";
import { toast } from "sonner";
import { useAwardsStore } from "@/stores/awards";
import type { Award } from "@/types";

function ConfettiPiece({ index }: { index: number }) {
  const left = Math.random() * 100;
  const delay = Math.random() * 0.5;
  const hue = Math.floor(Math.random() * 360);
  const size = 6 + Math.random() * 6;

  return (
    <div
      className="confetti-piece absolute top-0"
      style={{
        left: `${left}%`,
        width: `${size}px`,
        height: `${size * 0.6}px`,
        backgroundColor: `hsl(${hue}, 80%, 60%)`,
        animationDelay: `${delay}s`,
        borderRadius: "2px",
      }}
      aria-hidden="true"
      key={index}
    />
  );
}

function LevelUpModal({ award, onDismiss }: { award: Award; onDismiss: () => void }) {
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

export function AwardCelebration() {
  const pending = useAwardsStore((s) => s.pending);
  const shift = useAwardsStore((s) => s.shift);
  const [levelUpAward, setLevelUpAward] = useState<Award | null>(null);

  const processNext = useCallback(() => {
    if (pending.length === 0) return;
    const next = shift();
    if (!next) return;

    if (next.type === "badge" && next.badge) {
      toast(
        <div className="flex items-center gap-3">
          <span className="text-2xl" aria-hidden="true">
            {next.badge.icon}
          </span>
          <div>
            <p className="font-semibold text-[var(--text-primary)]">
              Badge earned!
            </p>
            <p className="text-xs text-[var(--text-secondary)]">
              {next.badge.name}
            </p>
          </div>
        </div>,
        { duration: 5000 },
      );
    } else if (next.type === "level_up") {
      setLevelUpAward(next);
    }
  }, [pending, shift]);

  useEffect(() => {
    if (pending.length > 0 && !levelUpAward) {
      processNext();
    }
  }, [pending, levelUpAward, processNext]);

  const handleDismissLevelUp = useCallback(() => {
    setLevelUpAward(null);
  }, []);

  return (
    <>
      {levelUpAward && (
        <LevelUpModal award={levelUpAward} onDismiss={handleDismissLevelUp} />
      )}
      {/* CSS animations injected via a style tag */}
      <style>{`
        @keyframes confetti-fall {
          0% { transform: translateY(-10vh) rotate(0deg); opacity: 1; }
          100% { transform: translateY(110vh) rotate(720deg); opacity: 0; }
        }
        .confetti-piece {
          animation: confetti-fall 2.5s ease-in forwards;
        }
        @keyframes level-up-pop {
          0% { transform: scale(0.7); opacity: 0; }
          60% { transform: scale(1.05); opacity: 1; }
          100% { transform: scale(1); opacity: 1; }
        }
        .level-up-entrance {
          animation: level-up-pop 0.4s ease-out forwards;
        }
      `}</style>
    </>
  );
}
