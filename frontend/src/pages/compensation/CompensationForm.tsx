import { useState, useEffect, useCallback } from "react";
import type { FormEvent } from "react";
import { useParams, useNavigate, Link } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { Select } from "@/components/Select";
import { CurrencyAutocomplete } from "@/components/CurrencyAutocomplete";
import { SpinnerIcon } from "@/components/icons";
import { useJobsQuery } from "@/hooks/queries/useJobsQuery";
import {
  useCompensationEntryQuery,
  useCreateCompensationMutation,
  useUpdateCompensationMutation,
} from "@/hooks/queries/useCompensationQuery";
import { majorToMinor, minorToMajor, PAY_BASIS_OPTIONS } from "@/lib/money";

function todayISO(): string {
  return new Date().toISOString().slice(0, 10);
}

function minorToInput(minor: number): string {
  return minor ? String(minorToMajor(minor)) : "";
}

// parseAmount returns the value in minor units, or null if the input is not a
// valid non-negative number. A blank field is treated as zero.
function parseAmount(value: string): number | null {
  const trimmed = value.trim();
  if (trimmed === "") return 0;
  const num = Number(trimmed);
  if (Number.isNaN(num) || num < 0) return null;
  return majorToMinor(num);
}

export default function CompensationForm() {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEditing = !!id;

  const { data: jobs, isLoading: isLoadingJobs } = useJobsQuery();
  const { data: existing, isLoading: isLoadingEntry } = useCompensationEntryQuery(id ?? "");
  const createMutation = useCreateCompensationMutation();
  const updateMutation = useUpdateCompensationMutation();

  const [jobId, setJobId] = useState("");
  const [effectiveDate, setEffectiveDate] = useState(todayISO);
  const [currency, setCurrency] = useState("GBP");
  const [payBasis, setPayBasis] = useState("annual");
  const [base, setBase] = useState("");
  const [bonus, setBonus] = useState("");
  const [equity, setEquity] = useState("");
  const [other, setOther] = useState("");
  const [note, setNote] = useState("");

  const [jobError, setJobError] = useState("");
  const [dateError, setDateError] = useState("");
  const [amountError, setAmountError] = useState("");
  const [populated, setPopulated] = useState(false);

  useEffect(() => {
    if (isEditing && existing && !populated) {
      setJobId(existing.job_id);
      setEffectiveDate(existing.effective_date.slice(0, 10));
      setCurrency(existing.currency);
      setPayBasis(existing.pay_basis);
      setBase(minorToInput(existing.amounts.base));
      setBonus(minorToInput(existing.amounts.bonus));
      setEquity(minorToInput(existing.amounts.equity));
      setOther(minorToInput(existing.amounts.other));
      setNote(existing.amounts.note ?? "");
      setPopulated(true);
    }
  }, [isEditing, existing, populated]);

  const handleSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();

      let hasError = false;

      if (!jobId) {
        setJobError("Select the role this compensation applies to");
        hasError = true;
      } else {
        setJobError("");
      }

      if (!effectiveDate) {
        setDateError("Effective date is required");
        hasError = true;
      } else {
        setDateError("");
      }

      const baseMinor = parseAmount(base);
      const bonusMinor = parseAmount(bonus);
      const equityMinor = parseAmount(equity);
      const otherMinor = parseAmount(other);

      if (
        baseMinor === null ||
        bonusMinor === null ||
        equityMinor === null ||
        otherMinor === null
      ) {
        setAmountError("Amounts must be valid, non-negative numbers");
        hasError = true;
      } else {
        setAmountError("");
      }

      if (hasError || baseMinor === null || bonusMinor === null || equityMinor === null || otherMinor === null) {
        return;
      }

      const payload = {
        job_id: jobId,
        effective_date: effectiveDate,
        currency,
        pay_basis: payBasis as (typeof PAY_BASIS_OPTIONS)[number]["value"],
        amounts: {
          base: baseMinor,
          bonus: bonusMinor,
          equity: equityMinor,
          other: otherMinor,
          note: note.trim() || null,
        },
      };

      if (isEditing && id) {
        updateMutation.mutate({ id, data: payload }, { onSuccess: () => navigate("/compensation") });
      } else {
        createMutation.mutate(payload, { onSuccess: () => navigate("/compensation") });
      }
    },
    [
      jobId,
      effectiveDate,
      currency,
      payBasis,
      base,
      bonus,
      equity,
      other,
      note,
      isEditing,
      id,
      createMutation,
      updateMutation,
      navigate,
    ],
  );

  const isPending = createMutation.isPending || updateMutation.isPending;

  if (isEditing && isLoadingEntry) {
    return (
      <>
        <Topbar title="Edit Compensation" />
        <div className="flex items-center justify-center py-16" role="status">
          <SpinnerIcon
            width={28}
            height={28}
            className="animate-spin text-[var(--accent-default)]"
            aria-hidden="true"
          />
          <span className="sr-only">Loading entry...</span>
        </div>
      </>
    );
  }

  const noJobs = !isLoadingJobs && (!jobs || jobs.length === 0);

  return (
    <>
      <Topbar title={isEditing ? "Edit Compensation" : "New Compensation"} />
      <div className="mx-auto max-w-2xl p-4 lg:p-6">
        {noJobs ? (
          <div className="py-16 text-center">
            <h2 className="mb-1 text-lg font-semibold text-[var(--text-primary)]">
              Add a role first
            </h2>
            <p className="mb-4 text-sm text-[var(--text-secondary)]">
              Compensation is recorded against a role on your timeline. Add a role, then come back to
              log its pay.
            </p>
            <Link
              to="/jobs/new"
              className="inline-flex items-center rounded-[var(--radius-md)]
                bg-[var(--accent-default)] px-4 py-2 text-sm font-medium
                text-[var(--button-primary-text)]"
            >
              Add a role
            </Link>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-5">
            <Select
              label="Role"
              value={jobId}
              onChange={(e) => {
                setJobId(e.target.value);
                if (jobError) setJobError("");
              }}
              error={jobError}
              required
            >
              <option value="" disabled>
                Select a role
              </option>
              {jobs?.map((job) => (
                <option key={job.id} value={job.id}>
                  {job.title} at {job.company}
                </option>
              ))}
            </Select>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <TextInput
                label="Effective date"
                type="date"
                value={effectiveDate}
                onChange={(e) => {
                  setEffectiveDate(e.target.value);
                  if (dateError) setDateError("");
                }}
                error={dateError}
                required
              />
              <Select label="Pay basis" value={payBasis} onChange={(e) => setPayBasis(e.target.value)}>
                {PAY_BASIS_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </Select>
            </div>

            <CurrencyAutocomplete
              label="Currency"
              placeholder="e.g. GBP"
              value={currency}
              onChange={(value) => setCurrency(value)}
            />

            <fieldset className="space-y-4">
              <legend className="text-sm font-medium text-[var(--text-secondary)]">
                Amounts ({currency})
              </legend>
              {amountError && (
                <p id="comp-amounts-error" className="text-xs text-[var(--color-error)]" role="alert">
                  {amountError}
                </p>
              )}
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <TextInput
                  label="Base"
                  type="number"
                  min="0"
                  step="any"
                  inputMode="decimal"
                  value={base}
                  onChange={(e) => setBase(e.target.value)}
                  aria-invalid={!!amountError}
                  aria-describedby={amountError ? "comp-amounts-error" : undefined}
                />
                <TextInput
                  label="Bonus"
                  type="number"
                  min="0"
                  step="any"
                  inputMode="decimal"
                  value={bonus}
                  onChange={(e) => setBonus(e.target.value)}
                  aria-invalid={!!amountError}
                  aria-describedby={amountError ? "comp-amounts-error" : undefined}
                />
                <TextInput
                  label="Equity (annualised)"
                  type="number"
                  min="0"
                  step="any"
                  inputMode="decimal"
                  value={equity}
                  onChange={(e) => setEquity(e.target.value)}
                  aria-invalid={!!amountError}
                  aria-describedby={amountError ? "comp-amounts-error" : undefined}
                />
                <TextInput
                  label="Other"
                  type="number"
                  min="0"
                  step="any"
                  inputMode="decimal"
                  value={other}
                  onChange={(e) => setOther(e.target.value)}
                  aria-invalid={!!amountError}
                  aria-describedby={amountError ? "comp-amounts-error" : undefined}
                />
              </div>
            </fieldset>

            <div className="flex flex-col gap-1.5">
              <label
                htmlFor="comp-note"
                className="text-sm font-medium text-[var(--text-secondary)]"
              >
                Note
              </label>
              <textarea
                id="comp-note"
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="Context for this change, e.g. promotion to Senior."
                rows={3}
                className="w-full resize-y rounded-[var(--radius-md)] border
                  border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-sm
                  text-[var(--text-primary)] placeholder:text-[var(--text-tertiary)]
                  transition-colors focus:outline-none focus:ring-2
                  focus:ring-[var(--accent-default)] focus:ring-offset-1
                  focus:ring-offset-[var(--bg-base)]"
              />
            </div>

            <p className="text-xs text-[var(--text-tertiary)]">
              Compensation is private and encrypted. It never appears on your public profile.
            </p>

            <div className="flex items-center justify-end gap-3 pt-2">
              <Button
                type="button"
                variant="secondary"
                onClick={() => navigate("/compensation")}
                disabled={isPending}
              >
                Cancel
              </Button>
              <Button type="submit" loading={isPending}>
                {isEditing ? "Save changes" : "Add entry"}
              </Button>
            </div>
          </form>
        )}
      </div>
    </>
  );
}
