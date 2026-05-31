import { useMemo } from "react";
import type { JobPreview } from "@/types";
import { CloseIcon } from "@/components/icons";

interface JobCardProps {
  job: JobPreview;
  index: number;
  onRemove: () => void;
}

export function JobCard({ job, index, onRemove }: JobCardProps) {
  const dateRange = useMemo(() => {
    const start = job.start_date
      ? new Date(job.start_date).toLocaleDateString("en-GB", { month: "short", year: "numeric" })
      : "";
    const end = job.end_date
      ? new Date(job.end_date).toLocaleDateString("en-GB", { month: "short", year: "numeric" })
      : "Present";
    return start ? `${start} - ${end}` : "";
  }, [job.start_date, job.end_date]);

  return (
    <div
      className="animate-fade-in-up group relative flex items-start gap-3 rounded-[var(--radius-lg)]
        border border-[var(--border-subtle)] bg-[var(--bg-elevated)] p-4 transition-all duration-200
        hover:-translate-y-0.5 hover:shadow-lg hover:shadow-stone-900/10"
      style={{ animationDelay: `${index * 50}ms` }}
    >
      <div className="min-w-0 flex-1">
        <h4 className="font-semibold text-[var(--text-primary)]">{job.title}</h4>
        <p className="text-sm text-[var(--text-secondary)]">{job.company}</p>
        {dateRange && (
          <p className="mt-1 text-xs text-[var(--text-tertiary)]">{dateRange}</p>
        )}
        <div className="mt-2 flex flex-wrap gap-2">
          {job.employment_type && (
            <span
              className="inline-flex items-center rounded-full bg-[var(--bg-surface)] px-2 py-0.5
                text-xs font-medium text-[var(--text-secondary)]"
            >
              {job.employment_type}
            </span>
          )}
          {job.location && (
            <span className="text-xs text-[var(--text-tertiary)]">{job.location}</span>
          )}
          {job.work_mode && job.work_mode !== "onsite" && (
            <span
              className={[
                "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
                job.work_mode === "remote"
                  ? "bg-[var(--color-success)]/10 text-[var(--color-success)]"
                  : "bg-[var(--color-info)]/10 text-[var(--color-info)]",
              ].join(" ")}
            >
              {job.work_mode === "remote" ? "Remote" : "Hybrid"}
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
        aria-label={`Remove ${job.title} at ${job.company}`}
      >
        <CloseIcon width={16} height={16} />
      </button>
    </div>
  );
}
