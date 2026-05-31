export const STATUS_OPTIONS = [
  { value: "planning", label: "Planning" },
  { value: "studying", label: "Studying" },
  { value: "scheduled", label: "Scheduled" },
  { value: "passed", label: "Passed" },
  { value: "expired", label: "Expired" },
];

export const STATUS_COLOURS: Record<string, { bg: string; text: string }> = {
  planning: {
    bg: "bg-[var(--text-tertiary)]/10",
    text: "text-[var(--text-tertiary)]",
  },
  studying: { bg: "bg-[var(--accent-default)]/10", text: "text-[var(--accent-default)]" },
  scheduled: { bg: "bg-[var(--color-warning)]/10", text: "text-[var(--color-warning)]" },
  passed: { bg: "bg-[var(--color-success)]/10", text: "text-[var(--color-success)]" },
  expired: { bg: "bg-[var(--color-error)]/10", text: "text-[var(--color-error)]" },
};

export function StatusBadge({ status }: { status: string }) {
  const colours = STATUS_COLOURS[status] ?? STATUS_COLOURS.planning;
  const label =
    STATUS_OPTIONS.find((o) => o.value === status)?.label ?? status;
  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${colours.bg} ${colours.text}`}
    >
      {label}
    </span>
  );
}
