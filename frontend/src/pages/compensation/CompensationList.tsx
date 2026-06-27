import { useState, useEffect, useMemo, useCallback, lazy, Suspense } from "react";
import { useNavigate } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { Select } from "@/components/Select";
import { ConfirmModal } from "@/components/ConfirmModal";
import { WalletIcon, SpinnerIcon } from "@/components/icons";
import {
  useCompensationQuery,
  useDeleteCompensationMutation,
} from "@/hooks/queries/useCompensationQuery";
import { useJobsQuery } from "@/hooks/queries/useJobsQuery";
import { formatMoney, totalMinor, PAY_BASIS_LABELS } from "@/lib/money";
import type { Compensation } from "@/types";

// recharts is heavy, so the chart loads as its own chunk only when it is shown.
const CompensationTrendChart = lazy(() =>
  import("./CompensationTrendChart").then((m) => ({ default: m.CompensationTrendChart })),
);

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString("en-GB", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}

function breakdownParts(entry: Compensation): string[] {
  const { amounts, currency } = entry;
  const parts: string[] = [`Base ${formatMoney(amounts.base, currency)}`];
  if (amounts.bonus > 0) parts.push(`Bonus ${formatMoney(amounts.bonus, currency)}`);
  if (amounts.equity > 0) parts.push(`Equity ${formatMoney(amounts.equity, currency)}`);
  if (amounts.other > 0) parts.push(`Other ${formatMoney(amounts.other, currency)}`);
  return parts;
}

export default function CompensationList() {
  const navigate = useNavigate();
  const { data: entries, isLoading, isError } = useCompensationQuery();
  const { data: jobs } = useJobsQuery();
  const deleteMutation = useDeleteCompensationMutation();
  const [deleteTarget, setDeleteTarget] = useState<Compensation | null>(null);

  const currencies = useMemo(() => {
    const set = new Set<string>();
    entries?.forEach((e) => set.add(e.currency));
    return Array.from(set).sort();
  }, [entries]);

  // Default the chart to the most recent entry's currency. Totals are only
  // meaningful within a single currency, so the chart never mixes them.
  const [selectedCurrency, setSelectedCurrency] = useState("");
  useEffect(() => {
    if (currencies.length > 0 && !currencies.includes(selectedCurrency)) {
      setSelectedCurrency(entries?.[0]?.currency ?? currencies[0]);
    }
  }, [currencies, selectedCurrency, entries]);

  const jobLabel = useMemo(() => {
    const map = new Map<string, string>();
    jobs?.forEach((j) => map.set(j.id, `${j.title} at ${j.company}`));
    return map;
  }, [jobs]);

  const grouped = useMemo(() => {
    const groups: { jobId: string; label: string; items: Compensation[] }[] = [];
    const index = new Map<string, number>();
    entries?.forEach((e) => {
      if (!index.has(e.job_id)) {
        index.set(e.job_id, groups.length);
        groups.push({
          jobId: e.job_id,
          label: jobLabel.get(e.job_id) ?? "Unknown role",
          items: [],
        });
      }
      groups[index.get(e.job_id)!].items.push(e);
    });
    return groups;
  }, [entries, jobLabel]);

  const chartEntries = useMemo(
    () => entries?.filter((e) => e.currency === selectedCurrency) ?? [],
    [entries, selectedCurrency],
  );

  // The chart itself is aria-hidden, so describe the trend (date order) in text
  // for screen-reader users; the list below is grouped by role, not by date.
  const chartSummary = useMemo(
    () =>
      chartEntries
        .slice()
        .sort((a, b) => a.effective_date.localeCompare(b.effective_date))
        .map(
          (e) =>
            `${new Date(e.effective_date).toLocaleDateString("en-GB", {
              month: "short",
              year: "numeric",
            })} ${formatMoney(totalMinor(e.amounts), e.currency)}`,
        )
        .join(", "),
    [chartEntries],
  );

  const handleDelete = useCallback(() => {
    if (!deleteTarget) return;
    deleteMutation.mutate(deleteTarget.id, {
      onSuccess: () => setDeleteTarget(null),
    });
  }, [deleteTarget, deleteMutation]);

  const hasEntries = !isLoading && !isError && entries && entries.length > 0;

  return (
    <>
      <Topbar title="Compensation" />
      <div className="mx-auto max-w-3xl space-y-6 p-4 lg:p-6">
        <div className="flex items-center justify-between gap-3">
          <p className="text-sm text-[var(--text-secondary)]">
            Your private pay history. Encrypted, and never shown on your public profile.
          </p>
          <Button size="sm" onClick={() => navigate("/compensation/new")}>
            Add entry
          </Button>
        </div>

        {isLoading && (
          <div className="flex items-center justify-center py-16" role="status">
            <SpinnerIcon
              width={28}
              height={28}
              className="animate-spin text-[var(--accent-default)]"
              aria-hidden="true"
            />
            <span className="sr-only">Loading compensation...</span>
          </div>
        )}

        {isError && (
          <div
            className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
              bg-[var(--color-error)]/5 p-4 text-center text-sm text-[var(--color-error)]"
            role="alert"
          >
            Failed to load compensation. Please try again later.
          </div>
        )}

        {!isLoading && !isError && (!entries || entries.length === 0) && (
          <div className="py-16 text-center">
            <WalletIcon
              width={48}
              height={48}
              className="mx-auto mb-4 text-[var(--accent-default)]"
              aria-hidden="true"
            />
            <h2 className="mb-1 text-lg font-semibold text-[var(--text-primary)]">
              No compensation logged yet
            </h2>
            <p className="mb-4 text-sm text-[var(--text-secondary)]">
              Track base, bonus and equity over time so you always know what you have earned.
            </p>
            <Button size="sm" onClick={() => navigate("/compensation/new")}>
              Log your first entry
            </Button>
          </div>
        )}

        {hasEntries && (
          <>
            {/* Trend */}
            <section
              className="rounded-[var(--radius-lg)] border border-[var(--border-subtle)]
                bg-[var(--bg-surface)] p-4"
              aria-labelledby="comp-trend-heading"
            >
              <div className="mb-3 flex items-center justify-between gap-3">
                <h2
                  id="comp-trend-heading"
                  className="text-sm font-semibold text-[var(--text-primary)]"
                >
                  Total compensation over time
                </h2>
                {currencies.length > 1 && (
                  <div className="w-32">
                    <Select
                      aria-label="Chart currency"
                      value={selectedCurrency}
                      onChange={(e) => setSelectedCurrency(e.target.value)}
                    >
                      {currencies.map((c) => (
                        <option key={c} value={c}>
                          {c}
                        </option>
                      ))}
                    </Select>
                  </div>
                )}
              </div>
              <p className="sr-only">
                Total compensation in {selectedCurrency} by effective date: {chartSummary}.
              </p>
              <div aria-hidden="true">
                <Suspense fallback={<div className="h-[260px]" />}>
                  <CompensationTrendChart entries={chartEntries} currency={selectedCurrency} />
                </Suspense>
              </div>
            </section>

            {/* Entries grouped by role */}
            <div className="space-y-6">
              {grouped.map((group) => (
                <section key={group.jobId} aria-labelledby={`comp-group-${group.jobId}`}>
                  <h2
                    id={`comp-group-${group.jobId}`}
                    className="mb-2 text-sm font-semibold text-[var(--text-primary)]"
                  >
                    {group.label}
                  </h2>
                  <ul className="space-y-3">
                    {group.items.map((entry) => (
                      <li
                        key={entry.id}
                        className="group rounded-[var(--radius-lg)] border
                          border-[var(--border-subtle)] bg-[var(--bg-surface)] p-4
                          transition-colors hover:border-[var(--border-default)]"
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div className="min-w-0 flex-1">
                            <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
                              <span className="text-base font-semibold text-[var(--text-primary)]">
                                {formatMoney(totalMinor(entry.amounts), entry.currency)}
                              </span>
                              <span className="text-xs text-[var(--text-tertiary)]">
                                {PAY_BASIS_LABELS[entry.pay_basis] ?? entry.pay_basis}
                              </span>
                              <time
                                dateTime={entry.effective_date}
                                className="text-xs text-[var(--text-tertiary)]"
                              >
                                {formatDate(entry.effective_date)}
                              </time>
                            </div>
                            <p className="mt-1.5 text-sm text-[var(--text-secondary)]">
                              {breakdownParts(entry).join(" · ")}
                            </p>
                            {entry.amounts.note && (
                              <p className="mt-1 text-sm text-[var(--text-tertiary)]">
                                {entry.amounts.note}
                              </p>
                            )}
                          </div>
                          <div
                            className="flex shrink-0 gap-1 opacity-0 transition-opacity
                              group-hover:opacity-100 group-focus-within:opacity-100 touch:opacity-100"
                          >
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => navigate(`/compensation/${entry.id}/edit`)}
                              aria-label={`Edit compensation from ${formatDate(entry.effective_date)}`}
                            >
                              Edit
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => setDeleteTarget(entry)}
                              aria-label={`Delete compensation from ${formatDate(entry.effective_date)}`}
                            >
                              Delete
                            </Button>
                          </div>
                        </div>
                      </li>
                    ))}
                  </ul>
                </section>
              ))}
            </div>
          </>
        )}
      </div>

      <ConfirmModal
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete compensation entry"
        message={
          deleteTarget
            ? `Delete the compensation entry from ${formatDate(deleteTarget.effective_date)}? This cannot be undone.`
            : ""
        }
        confirmLabel="Delete"
        loading={deleteMutation.isPending}
      />
    </>
  );
}
