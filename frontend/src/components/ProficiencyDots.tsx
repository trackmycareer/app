export const PROFICIENCY_LABELS: Record<number, string> = {
  1: "Beginner",
  2: "Intermediate",
  3: "Advanced",
  4: "Expert",
};

export const PROFICIENCY_COLOURS: Record<number, string> = {
  1: "var(--text-tertiary)",
  2: "var(--accent-muted)",
  3: "var(--accent-default)",
  4: "var(--accent-bright)",
};

export function ProficiencyDots({ level }: { level: number }) {
  return (
    <div className="flex items-center gap-1.5" aria-label={`Proficiency: ${PROFICIENCY_LABELS[level]}`}>
      {[1, 2, 3, 4].map((dot) => (
        <span
          key={dot}
          className="inline-block h-2 w-2 rounded-full transition-colors"
          style={{
            backgroundColor:
              dot <= level ? PROFICIENCY_COLOURS[level] : "var(--border-default)",
          }}
          aria-hidden="true"
        />
      ))}
      <span className="ml-1 text-xs text-[var(--text-tertiary)]">
        {PROFICIENCY_LABELS[level]}
      </span>
    </div>
  );
}
