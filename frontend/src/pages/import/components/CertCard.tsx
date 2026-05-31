import type { CertPreview } from "@/types";
import { CloseIcon } from "@/components/icons";

interface CertCardProps {
  cert: CertPreview;
  index: number;
  onRemove: () => void;
}

export function CertCard({ cert, index, onRemove }: CertCardProps) {
  return (
    <div
      className="animate-fade-in-up group relative flex items-start gap-3 rounded-[var(--radius-lg)]
        border border-[var(--border-subtle)] bg-[var(--bg-elevated)] p-4 transition-all duration-200
        hover:-translate-y-0.5 hover:shadow-lg hover:shadow-stone-900/10"
      style={{ animationDelay: `${index * 50}ms` }}
    >
      <div className="min-w-0 flex-1">
        <h4 className="font-semibold text-[var(--text-primary)]">{cert.name}</h4>
        <p className="text-sm text-[var(--text-secondary)]">{cert.provider}</p>
        <div className="mt-2 flex flex-wrap gap-2">
          {cert.status && (
            <span
              className="inline-flex items-center rounded-full bg-[var(--bg-surface)] px-2 py-0.5
                text-xs font-medium text-[var(--text-secondary)]"
            >
              {cert.status}
            </span>
          )}
          {cert.earned_date && (
            <span className="text-xs text-[var(--text-tertiary)]">
              Earned{" "}
              {new Date(cert.earned_date).toLocaleDateString("en-GB", {
                month: "short",
                year: "numeric",
              })}
            </span>
          )}
          {cert.expiry_date && (
            <span className="text-xs text-[var(--text-tertiary)]">
              Expires{" "}
              {new Date(cert.expiry_date).toLocaleDateString("en-GB", {
                month: "short",
                year: "numeric",
              })}
            </span>
          )}
        </div>
      </div>
      <button
        type="button"
        onClick={onRemove}
        className="flex-shrink-0 rounded-[var(--radius-sm)] p-1 text-[var(--text-tertiary)]
          opacity-0 transition-all hover:bg-[var(--bg-hover)] hover:text-[var(--color-error)]
          group-hover:opacity-100 group-focus-within:opacity-100"
        aria-label={`Remove ${cert.name}`}
      >
        <CloseIcon width={16} height={16} />
      </button>
    </div>
  );
}
