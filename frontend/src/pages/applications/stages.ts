import type { ApplicationStatus } from "@/types";

export interface StageDef {
  key: ApplicationStatus;
  label: string;
  // CSS custom property used for the column header dot and card accent.
  accent: string;
}

// Canonical board order. Accepted and rejected are terminal stages.
export const APPLICATION_STAGES: StageDef[] = [
  { key: "wishlist", label: "Wishlist", accent: "var(--text-tertiary)" },
  { key: "applied", label: "Applied", accent: "var(--color-info)" },
  { key: "screen", label: "Screening", accent: "var(--color-info)" },
  { key: "interview", label: "Interview", accent: "var(--color-warning)" },
  { key: "offer", label: "Offer", accent: "var(--accent-default)" },
  { key: "accepted", label: "Accepted", accent: "var(--color-success)" },
  { key: "rejected", label: "Rejected", accent: "var(--color-error)" },
];

export const STAGE_LABELS = APPLICATION_STAGES.reduce(
  (acc, stage) => {
    acc[stage.key] = stage.label;
    return acc;
  },
  {} as Record<ApplicationStatus, string>,
);
