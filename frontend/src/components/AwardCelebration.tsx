import { useEffect, useState, useCallback } from "react";
import { toast } from "sonner";
import { useAwardsStore } from "@/stores/awards";
import { LevelUpModal } from "@/components/LevelUpModal";
import type { Award } from "@/types";

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
