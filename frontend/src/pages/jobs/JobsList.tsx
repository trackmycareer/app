import { useState, useCallback } from "react";
import { useNavigate } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { ConfirmModal } from "@/components/ConfirmModal";
import { BriefcaseIcon, SpinnerIcon } from "@/components/icons";
import { useJobsQuery, useDeleteJobMutation } from "@/hooks/queries/useJobsQuery";
import type { Job } from "@/types";

const EMPLOYMENT_TYPE_LABELS: Record<string, string> = {
  full_time: "Full-time",
  part_time: "Part-time",
  contract: "Contract",
  freelance: "Freelance",
  internship: "Internship",
  education: "Education",
  volunteer: "Volunteer",
};

const TRANSITION_TYPE_LABELS: Record<string, string> = {
  promotion: "Promotion",
  lateral_move: "Lateral move",
  company_change: "Company change",
  first_role: "First role",
};

function formatDateRange(startDate: string, endDate: string | null): string {
  const start = new Date(startDate).toLocaleDateString("en-GB", {
    month: "short",
    year: "numeric",
  });
  if (!endDate) return `${start} to Present`;
  const end = new Date(endDate).toLocaleDateString("en-GB", {
    month: "short",
    year: "numeric",
  });
  return `${start} to ${end}`;
}

function calculateDuration(startDate: string, endDate: string | null): string {
  const start = new Date(startDate);
  const end = endDate ? new Date(endDate) : new Date();
  const months =
    (end.getFullYear() - start.getFullYear()) * 12 +
    (end.getMonth() - start.getMonth());
  const years = Math.floor(months / 12);
  const remainingMonths = months % 12;
  if (years === 0) return `${remainingMonths} mo`;
  if (remainingMonths === 0) return `${years} yr`;
  return `${years} yr ${remainingMonths} mo`;
}

export default function JobsList() {
  const navigate = useNavigate();
  const { data: jobs, isLoading, isError } = useJobsQuery();
  const deleteMutation = useDeleteJobMutation();
  const [deleteTarget, setDeleteTarget] = useState<Job | null>(null);

  const handleDelete = useCallback(() => {
    if (!deleteTarget) return;
    deleteMutation.mutate(deleteTarget.id, {
      onSuccess: () => setDeleteTarget(null),
    });
  }, [deleteTarget, deleteMutation]);

  return (
    <>
      <Topbar title="Career Timeline" />
      <div className="mx-auto max-w-3xl space-y-6 p-4 lg:p-6">
        {/* Header with add button */}
        <div className="flex items-center justify-between">
          <p className="text-sm text-[var(--text-secondary)]">
            Your career journey, from first role to present.
          </p>
          <Button size="sm" onClick={() => navigate("/jobs/new")}>
            Add role
          </Button>
        </div>

        {/* Loading state */}
        {isLoading && (
          <div className="flex items-center justify-center py-16" role="status">
            <SpinnerIcon
              width={28}
              height={28}
              className="animate-spin text-[var(--accent-default)]"
              aria-hidden="true"
            />
            <span className="sr-only">Loading roles...</span>
          </div>
        )}

        {/* Error state */}
        {isError && (
          <div
            className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
              bg-[var(--color-error)]/5 p-4 text-center text-sm text-[var(--color-error)]"
            role="alert"
          >
            Failed to load roles. Please try again later.
          </div>
        )}

        {/* Empty state */}
        {!isLoading && !isError && (!jobs || jobs.length === 0) && (
          <div className="py-16 text-center">
            <BriefcaseIcon
              width={48}
              height={48}
              className="mx-auto mb-4 text-[var(--accent-default)]"
              aria-hidden="true"
            />
            <h2 className="mb-1 text-lg font-semibold text-[var(--text-primary)]">
              No roles added yet
            </h2>
            <p className="mb-4 text-sm text-[var(--text-secondary)]">
              Every career has a story. Start yours by adding your current role.
            </p>
            <Button size="sm" onClick={() => navigate("/jobs/new")}>
              Add your first role
            </Button>
          </div>
        )}

        {/* Timeline */}
        {!isLoading && !isError && jobs && jobs.length > 0 && (
          <div className="relative" aria-label="Career timeline">
            {/* Vertical timeline line */}
            <div
              className="absolute left-4 top-2 bottom-2 w-px bg-[var(--border-default)]"
              aria-hidden="true"
            />

            <ul className="space-y-4">
              {jobs.map((job, index) => (
                <li key={job.id} className="relative pl-10">
                  {/* Timeline dot */}
                  <div
                    className={[
                      "absolute left-2.5 top-5 h-3 w-3 rounded-full border-2",
                      "border-[var(--bg-surface)]",
                      !job.end_date
                        ? "bg-[var(--accent-default)]"
                        : "bg-[var(--text-tertiary)]",
                    ].join(" ")}
                    aria-hidden="true"
                  />

                  {/* Card */}
                  <div
                    className={[
                      "animate-fade-in-up group rounded-[var(--radius-lg)] border",
                      "bg-[var(--bg-surface)] p-4 transition-all duration-200",
                      "hover:-translate-y-0.5 hover:shadow-lg hover:shadow-stone-900/10",
                      "hover:border-[var(--border-default)]",
                      !job.end_date
                        ? "border-[var(--accent-default)]/30"
                        : "border-[var(--border-subtle)]",
                    ].join(" ")}
                    style={{ animationDelay: `${index * 50}ms` }}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0 flex-1">
                        {/* Transition type indicator */}
                        {job.transition_type && index > 0 && (
                          <span
                            className="mb-1.5 inline-block rounded-full bg-[var(--accent-default)]/10
                              px-2 py-0.5 text-xs font-medium text-[var(--accent-default)]"
                          >
                            {TRANSITION_TYPE_LABELS[job.transition_type] ??
                              job.transition_type}
                          </span>
                        )}

                        {/* Company and title */}
                        <h3 className="font-semibold text-[var(--text-primary)]">
                          {job.title}
                        </h3>
                        <p className="text-sm font-medium text-[var(--text-secondary)]">
                          {job.company}
                        </p>

                        {/* Date range and duration */}
                        <div className="mt-1.5 flex flex-wrap items-center gap-2">
                          <time className="text-xs text-[var(--text-tertiary)]">
                            {formatDateRange(job.start_date, job.end_date)}
                          </time>
                          <span
                            className="text-[var(--text-tertiary)]"
                            aria-hidden="true"
                          >
                            &middot;
                          </span>
                          <span className="text-xs text-[var(--text-tertiary)]">
                            {calculateDuration(job.start_date, job.end_date)}
                          </span>
                        </div>

                        {/* Badges row */}
                        <div className="mt-2 flex flex-wrap items-center gap-2">
                          {/* Employment type badge */}
                          <span
                            className="inline-flex items-center rounded-full
                              bg-[var(--bg-elevated)] px-2 py-0.5 text-xs font-medium
                              text-[var(--text-secondary)]"
                          >
                            {EMPLOYMENT_TYPE_LABELS[job.employment_type] ??
                              job.employment_type}
                          </span>

                          {/* Location */}
                          {job.location && (
                            <span className="text-xs text-[var(--text-tertiary)]">
                              {job.location}
                            </span>
                          )}

                          {/* Work mode badge */}
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

                          {/* Current role badge */}
                          {!job.end_date && (
                            <span
                              className="inline-flex items-center rounded-full
                                bg-[var(--accent-default)]/10 px-2 py-0.5 text-xs
                                font-medium text-[var(--accent-default)]"
                            >
                              Current
                            </span>
                          )}
                        </div>

                        {/* Responsibilities preview */}
                        {job.responsibilities && (
                          <p
                            className="mt-2 line-clamp-2 text-sm
                              text-[var(--text-secondary)]"
                          >
                            {job.responsibilities}
                          </p>
                        )}
                      </div>

                      {/* Action buttons */}
                      <div
                        className="flex shrink-0 gap-1 opacity-0 transition-opacity
                          group-hover:opacity-100 group-focus-within:opacity-100 touch:opacity-100"
                      >
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => navigate(`/jobs/${job.id}/edit`)}
                          aria-label={`Edit ${job.title} at ${job.company}`}
                        >
                          Edit
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setDeleteTarget(job)}
                          aria-label={`Delete ${job.title} at ${job.company}`}
                        >
                          Delete
                        </Button>
                      </div>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>

      {/* Delete confirmation modal */}
      <ConfirmModal
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete role"
        message={`Are you sure you want to delete "${deleteTarget?.title} at ${deleteTarget?.company}"? This action cannot be undone.`}
        confirmLabel="Delete"
        loading={deleteMutation.isPending}
      />
    </>
  );
}
