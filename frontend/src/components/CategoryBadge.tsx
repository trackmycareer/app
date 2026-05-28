const CATEGORY_CONFIG: Record<string, { label: string; classes: string }> = {
  shipped_feature: {
    label: "Shipped feature",
    classes: "bg-blue-500/15 text-blue-400",
  },
  positive_feedback: {
    label: "Positive feedback",
    classes: "bg-emerald-500/15 text-emerald-400",
  },
  process_improvement: {
    label: "Process improvement",
    classes: "bg-purple-500/15 text-purple-400",
  },
  cost_saving: {
    label: "Cost saving",
    classes: "bg-amber-500/15 text-amber-400",
  },
  leadership_moment: {
    label: "Leadership moment",
    classes: "bg-rose-500/15 text-rose-400",
  },
  general: {
    label: "General",
    classes: "bg-zinc-500/15 text-zinc-400",
  },
};

interface CategoryBadgeProps {
  category: string;
}

export function CategoryBadge({ category }: CategoryBadgeProps) {
  const config = CATEGORY_CONFIG[category] ?? CATEGORY_CONFIG.general;

  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${config.classes}`}
    >
      {config.label}
    </span>
  );
}

export const CATEGORY_OPTIONS = [
  { value: "shipped_feature", label: "Shipped feature" },
  { value: "positive_feedback", label: "Positive feedback" },
  { value: "process_improvement", label: "Process improvement" },
  { value: "cost_saving", label: "Cost saving" },
  { value: "leadership_moment", label: "Leadership moment" },
  { value: "general", label: "General" },
];
