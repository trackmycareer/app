import { useState, useCallback } from "react";
import { useNavigate } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { Select } from "@/components/Select";
import { ConfirmModal } from "@/components/ConfirmModal";
import { SpinnerIcon, AwardIcon, ExternalLinkIcon } from "@/components/icons";
import {
  useCertsQuery,
  useDeleteCertMutation,
} from "@/hooks/queries/useCertsQuery";
import { StatusBadge, STATUS_OPTIONS } from "./components/StatusBadge";
import { ProgressBar } from "./components/ProgressBar";
import type { CertListParams } from "@/hooks/queries/useCertsQuery";
import type { Certification } from "@/types";

const PAGE_SIZE = 20;

function isExpiringSoon(expiryDate: string | null): boolean {
  if (!expiryDate) return false;
  const expiry = new Date(expiryDate);
  const now = new Date();
  const diffMs = expiry.getTime() - now.getTime();
  const diffDays = diffMs / (1000 * 60 * 60 * 24);
  return diffDays > 0 && diffDays <= 90;
}

export default function CertificationsList() {
  const navigate = useNavigate();

  const [search, setSearch] = useState("");
  const [filterStatus, setFilterStatus] = useState("");
  const [offset, setOffset] = useState(0);
  const [deleteTarget, setDeleteTarget] = useState<Certification | null>(null);

  const params: CertListParams = {
    limit: PAGE_SIZE,
    offset,
  };
  if (search) params.search = search;
  if (filterStatus) params.status = filterStatus;

  const { data: certsData, isLoading, isError } = useCertsQuery(params);
  const deleteMutation = useDeleteCertMutation();

  const certs = certsData?.certifications ?? [];
  const total = certsData?.total ?? certs.length;

  const handleSearchChange = useCallback((value: string) => {
    setSearch(value);
    setOffset(0);
  }, []);

  const handleFilterStatusChange = useCallback((value: string) => {
    setFilterStatus(value);
    setOffset(0);
  }, []);

  const handleDelete = useCallback(() => {
    if (!deleteTarget) return;
    deleteMutation.mutate(deleteTarget.id, {
      onSuccess: () => setDeleteTarget(null),
    });
  }, [deleteTarget, deleteMutation]);

  const hasMore = offset + PAGE_SIZE < total;
  const hasPrevious = offset > 0;

  return (
    <>
      <Topbar title="Certifications" />
      <div className="mx-auto max-w-3xl space-y-6 p-4 lg:p-6">
        {/* Actions and filters */}
        <section
          aria-label="Filter certifications"
          className="flex flex-col gap-3 sm:flex-row sm:items-end"
        >
          <div className="flex-1">
            <TextInput
              placeholder="Search certifications..."
              value={search}
              onChange={(e) => handleSearchChange(e.target.value)}
              aria-label="Search certifications"
            />
          </div>
          <div className="w-full sm:w-44">
            <Select
              value={filterStatus}
              onChange={(e) => handleFilterStatusChange(e.target.value)}
              aria-label="Filter by status"
            >
              <option value="">All statuses</option>
              {STATUS_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </Select>
          </div>
          <Button
            size="sm"
            onClick={() => navigate("/certifications/new")}
          >
            Add certification
          </Button>
        </section>

        {/* Certifications list */}
        <section aria-label="Your certifications">
          {isLoading && (
            <div
              className="flex items-center justify-center py-16"
              role="status"
            >
              <SpinnerIcon
                width={28}
                height={28}
                className="animate-spin text-[var(--accent-default)]"
                aria-hidden="true"
              />
              <span className="sr-only">Loading certifications...</span>
            </div>
          )}

          {isError && (
            <div
              className="rounded-[var(--radius-lg)] border
                border-[var(--color-error)]/30 bg-[var(--color-error)]/5
                p-4 text-center text-sm text-[var(--color-error)]"
              role="alert"
            >
              Failed to load certifications. Please try again later.
            </div>
          )}

          {!isLoading && !isError && certs.length === 0 && (
            <div className="py-16 text-center">
              <AwardIcon
                width={48}
                height={48}
                className="mx-auto mb-4 text-[var(--accent-default)]"
                aria-hidden="true"
              />
              <h2 className="mb-1 text-lg font-semibold text-[var(--text-primary)]">
                No certifications tracked yet
              </h2>
              <p className="mb-4 text-sm text-[var(--text-secondary)]">
                Certifications don&apos;t just prove skills, they prove commitment.
                Track your first one.
              </p>
              <Button onClick={() => navigate("/certifications/new")}>
                Add your first certification
              </Button>
            </div>
          )}

          {!isLoading && !isError && certs.length > 0 && (
            <ul className="space-y-3" aria-label="Certifications list">
              {certs.map((cert, index) => (
                <li
                  key={cert.id}
                  className="animate-fade-in-up group rounded-[var(--radius-lg)] border
                    border-[var(--border-subtle)] bg-[var(--bg-surface)]
                    p-4 transition-all duration-200 hover:-translate-y-0.5
                    hover:shadow-lg hover:shadow-stone-900/10
                    hover:border-[var(--border-default)]"
                  style={{ animationDelay: `${index * 50}ms` }}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <div className="mb-1 flex flex-wrap items-center gap-2">
                        <h3 className="font-medium text-[var(--text-primary)]">
                          {cert.name}
                        </h3>
                        <StatusBadge status={cert.status} />
                        {cert.status === "passed" &&
                          isExpiringSoon(cert.expiry_date) && (
                            <span
                              className="inline-flex items-center rounded-full
                                bg-[var(--color-warning)]/10 px-2 py-0.5 text-xs
                                font-medium text-[var(--color-warning)]"
                              title="Expiring within 90 days"
                            >
                              Expiring soon
                            </span>
                          )}
                      </div>
                      <p className="mb-2 text-sm text-[var(--text-secondary)]">
                        {cert.provider}
                      </p>
                      <div className="flex flex-wrap items-center gap-3">
                        {cert.earned_date && (
                          <time
                            dateTime={cert.earned_date}
                            className="text-xs text-[var(--text-tertiary)]"
                          >
                            Earned{" "}
                            {new Date(cert.earned_date).toLocaleDateString(
                              "en-GB",
                              {
                                day: "numeric",
                                month: "short",
                                year: "numeric",
                              }
                            )}
                          </time>
                        )}
                        {cert.expiry_date && (
                          <time
                            dateTime={cert.expiry_date}
                            className="text-xs text-[var(--text-tertiary)]"
                          >
                            Expires{" "}
                            {new Date(cert.expiry_date).toLocaleDateString(
                              "en-GB",
                              {
                                day: "numeric",
                                month: "short",
                                year: "numeric",
                              }
                            )}
                          </time>
                        )}
                        {cert.status === "studying" && (
                          <div className="flex items-center gap-2">
                            <ProgressBar value={cert.study_progress} />
                            <span className="text-xs text-[var(--text-tertiary)]">
                              {cert.study_progress}%
                            </span>
                          </div>
                        )}
                        {cert.credential_url && (
                          <a
                            href={cert.credential_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="inline-flex items-center gap-1
                              text-xs text-[var(--accent-default)]
                              hover:underline"
                          >
                            <ExternalLinkIcon width={12} height={12} />
                            Credential
                          </a>
                        )}
                      </div>
                    </div>
                    <div
                      className="flex shrink-0 gap-1 opacity-0
                        transition-opacity group-hover:opacity-100
                        group-focus-within:opacity-100 touch:opacity-100"
                    >
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          navigate(`/certifications/${cert.id}/edit`)
                        }
                        aria-label={`Edit ${cert.name}`}
                      >
                        Edit
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setDeleteTarget(cert)}
                        aria-label={`Delete ${cert.name}`}
                      >
                        Delete
                      </Button>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}

          {/* Pagination */}
          {!isLoading &&
            !isError &&
            certs.length > 0 &&
            (hasPrevious || hasMore) && (
              <nav
                aria-label="Pagination"
                className="flex items-center justify-between pt-4"
              >
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() =>
                    setOffset(Math.max(0, offset - PAGE_SIZE))
                  }
                  disabled={!hasPrevious}
                >
                  Previous
                </Button>
                <span className="text-xs text-[var(--text-tertiary)]">
                  Showing {offset + 1} to{" "}
                  {Math.min(offset + PAGE_SIZE, total)} of {total}
                </span>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => setOffset(offset + PAGE_SIZE)}
                  disabled={!hasMore}
                >
                  Next
                </Button>
              </nav>
            )}
        </section>
      </div>

      {/* Delete confirmation modal */}
      <ConfirmModal
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete certification"
        message={`Are you sure you want to delete "${deleteTarget?.name}"? This action cannot be undone.`}
        confirmLabel="Delete"
        loading={deleteMutation.isPending}
      />
    </>
  );
}
