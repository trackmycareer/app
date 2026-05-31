interface EvidenceTypeLabelProps {
  type: string;
}

export function EvidenceTypeLabel({ type }: EvidenceTypeLabelProps) {
  const labels: Record<string, string> = {
    win: "Win",
    certification: "Certification",
    job: "Job",
  };
  return (
    <span
      className="inline-flex items-center rounded-full bg-[var(--bg-elevated)] px-2 py-0.5
        text-xs font-medium text-[var(--text-secondary)]"
    >
      {labels[type] ?? type}
    </span>
  );
}
