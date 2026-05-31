import type { SkillPreview } from "@/types";
import { CloseIcon } from "@/components/icons";

interface SkillCardProps {
  skill: SkillPreview;
  index: number;
  onRemove: () => void;
}

export function SkillCard({ skill, index, onRemove }: SkillCardProps) {
  return (
    <div
      className="animate-fade-in-up group relative flex items-center gap-3 rounded-[var(--radius-lg)]
        border border-[var(--border-subtle)] bg-[var(--bg-elevated)] px-4 py-3 transition-all
        duration-200 hover:-translate-y-0.5 hover:shadow-lg hover:shadow-stone-900/10"
      style={{ animationDelay: `${index * 30}ms` }}
    >
      <div className="min-w-0 flex-1">
        <span className="text-sm font-medium text-[var(--text-primary)]">{skill.name}</span>
        {skill.category && (
          <span className="ml-2 text-xs text-[var(--text-tertiary)]">{skill.category}</span>
        )}
      </div>
      {skill.proficiency != null && (
        <div className="flex items-center gap-1.5">
          <div className="h-1.5 w-16 overflow-hidden rounded-full bg-[var(--bg-hover)]">
            <div
              className="h-full rounded-full bg-[var(--accent-default)] transition-all"
              style={{ width: `${Math.min(skill.proficiency, 100)}%` }}
            />
          </div>
          <span className="text-xs text-[var(--text-tertiary)]">{skill.proficiency}%</span>
        </div>
      )}
      <button
        type="button"
        onClick={onRemove}
        className="flex-shrink-0 rounded-[var(--radius-sm)] p-1 text-[var(--text-tertiary)]
          opacity-0 transition-all hover:bg-[var(--bg-hover)] hover:text-[var(--color-error)]
          group-hover:opacity-100 group-focus-within:opacity-100"
        aria-label={`Remove ${skill.name}`}
      >
        <CloseIcon width={16} height={16} />
      </button>
    </div>
  );
}
