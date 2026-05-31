const CATEGORY_CONFIG: Record<string, { label: string; classes: string }> = {
  shipped_feature: {
    label: "Shipped feature",
    classes: "bg-emerald-800/15 text-emerald-700 dark:text-emerald-400",
  },
  positive_feedback: {
    label: "Positive feedback",
    classes: "bg-amber-700/15 text-amber-700 dark:text-amber-400",
  },
  process_improvement: {
    label: "Process improvement",
    classes: "bg-teal-700/15 text-teal-700 dark:text-teal-400",
  },
  cost_saving: {
    label: "Cost saving",
    classes: "bg-yellow-700/15 text-yellow-700 dark:text-yellow-400",
  },
  leadership_moment: {
    label: "Leadership moment",
    classes: "bg-rose-700/15 text-rose-700 dark:text-rose-400",
  },
  general: {
    label: "General",
    classes: "bg-stone-500/15 text-stone-600 dark:text-stone-400",
  },
  learning: {
    label: "Learning",
    classes: "bg-violet-700/15 text-violet-700 dark:text-violet-400",
  },
  collaboration: {
    label: "Collaboration",
    classes: "bg-emerald-700/15 text-emerald-700 dark:text-emerald-400",
  },
  publication: {
    label: "Publication",
    classes: "bg-sky-700/15 text-sky-700 dark:text-sky-400",
  },
  project: {
    label: "Project",
    classes: "bg-cyan-700/15 text-cyan-700 dark:text-cyan-400",
  },
  membership: {
    label: "Membership",
    classes: "bg-emerald-700/15 text-emerald-700 dark:text-emerald-400",
  },
};

interface CategoryBadgeProps {
  category: string;
}

export function CategoryBadge({ category }: CategoryBadgeProps) {
  const config = CATEGORY_CONFIG[category] ?? {
    label: category.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase()),
    classes: "bg-stone-500/15 text-stone-600 dark:text-stone-400",
  };

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
  { value: "project", label: "Project" },
  { value: "publication", label: "Publication" },
  { value: "membership", label: "Membership" },
];
