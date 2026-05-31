import type { Skill } from "@/types";

interface SkillsGridProps {
  skills: Skill[];
}

export function SkillsGrid({ skills }: SkillsGridProps) {
  return (
    <div className="flex flex-wrap gap-2">
      {skills.map((skill) => (
        <span
          key={skill.id}
          className="inline-flex items-center gap-1.5 rounded-full border
            border-[var(--border-subtle)] bg-[var(--bg-elevated)] px-3 py-1.5
            text-xs font-medium text-[var(--text-secondary)]"
        >
          {skill.name}
          <span className="flex gap-0.5" aria-label={`Proficiency ${skill.proficiency} of 4`}>
            {[1, 2, 3, 4].map((dot) => (
              <span
                key={dot}
                className="inline-block h-1.5 w-1.5 rounded-full"
                style={{
                  backgroundColor:
                    dot <= skill.proficiency ? "var(--accent-default)" : "var(--border-default)",
                }}
                aria-hidden="true"
              />
            ))}
          </span>
        </span>
      ))}
    </div>
  );
}
