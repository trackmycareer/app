import { useState } from "react";
import type { Job } from "@/types";

interface JobTimelineProps {
  jobs: Job[];
}

function JobMeta({ job }: { job: Job }) {
  const parts: string[] = [];
  if (job.location) parts.push(job.location);
  if (job.work_mode) parts.push(job.work_mode);
  if (job.employment_type) parts.push(job.employment_type);
  if (parts.length === 0) return null;

  return (
    <p className="mt-0.5 text-xs text-[var(--text-tertiary)]">{parts.join(" · ")}</p>
  );
}

function Responsibilities({ text }: { text: string }) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="mt-2">
      <p
        className={`text-xs text-[var(--text-secondary)] whitespace-pre-line ${
          expanded ? "" : "line-clamp-3"
        }`}
      >
        {text}
      </p>
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="mt-1 text-xs font-medium text-[var(--accent)] hover:underline"
      >
        {expanded ? "Show less" : "Show more"}
      </button>
    </div>
  );
}

export function JobTimeline({ jobs }: JobTimelineProps) {
  const sorted = [...jobs].sort((a, b) => {
    const aDate = a.end_date ?? "9999-12-31";
    const bDate = b.end_date ?? "9999-12-31";
    return bDate.localeCompare(aDate);
  });

  return (
    <div className="space-y-3">
      {sorted.map((job) => (
        <div
          key={job.id}
          className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
            bg-[var(--bg-elevated)] p-4"
        >
          <div className="flex items-start justify-between gap-2">
            <div>
              <p className="text-sm font-semibold text-[var(--text-primary)]">{job.title}</p>
              <p className="text-xs text-[var(--text-secondary)]">{job.company}</p>
              <JobMeta job={job} />
            </div>
            <span className="shrink-0 text-xs text-[var(--text-tertiary)]">
              {new Date(job.start_date).toLocaleDateString("en-GB", {
                month: "short",
                year: "numeric",
              })}
              {" to "}
              {job.end_date
                ? new Date(job.end_date).toLocaleDateString("en-GB", {
                    month: "short",
                    year: "numeric",
                  })
                : "Present"}
            </span>
          </div>
          {job.responsibilities && <Responsibilities text={job.responsibilities} />}
        </div>
      ))}
    </div>
  );
}
