import { CurrencyAutocomplete } from "@/components/CurrencyAutocomplete";
import { Select } from "@/components/Select";
import { TextInput } from "@/components/TextInput";
import { Button } from "@/components/Button";
import { PAY_BASIS_OPTIONS } from "@/lib/money";
import { emptyDraft } from "@/lib/compensationDraft";
import type { CompensationDraft } from "@/lib/compensationDraft";

interface JobCompensationSectionProps {
  entries: CompensationDraft[];
  onChange: (entries: CompensationDraft[]) => void;
  // Keys of entries whose amounts failed validation on the last submit.
  errorKeys?: string[];
}

const ADD_BUTTON_ID = "job-comp-add";

function rowBaseId(key: string): string {
  return `job-comp-${key}-base`;
}
function rowRemoveId(key: string): string {
  return `job-comp-${key}-remove`;
}

export function JobCompensationSection({
  entries,
  onChange,
  errorKeys = [],
}: JobCompensationSectionProps) {
  const update = (key: string, patch: Partial<CompensationDraft>) =>
    onChange(entries.map((e) => (e.key === key ? { ...e, ...patch } : e)));

  const add = () => {
    const draft = emptyDraft();
    onChange([...entries, draft]);
    // Move focus into the new row rather than leaving it on the Add button.
    requestAnimationFrame(() => {
      document.getElementById(rowBaseId(draft.key))?.focus();
    });
  };

  const remove = (key: string) => {
    const idx = entries.findIndex((e) => e.key === key);
    const sibling = entries[idx + 1] ?? entries[idx - 1];
    onChange(entries.filter((e) => e.key !== key));
    // The row unmounts, so move focus somewhere deterministic instead of letting
    // it fall back to the document body.
    requestAnimationFrame(() => {
      const target = sibling
        ? document.getElementById(rowRemoveId(sibling.key))
        : document.getElementById(ADD_BUTTON_ID);
      target?.focus();
    });
  };

  return (
    <section className="space-y-3" aria-labelledby="job-comp-heading">
      <div className="flex items-end justify-between gap-3">
        <div>
          <h2 id="job-comp-heading" className="text-sm font-medium text-[var(--text-secondary)]">
            Compensation
          </h2>
          <p className="text-xs text-[var(--text-tertiary)]">
            Private and encrypted. Add a row for each pay change to build a history.
          </p>
        </div>
        <Button id={ADD_BUTTON_ID} type="button" variant="secondary" size="sm" onClick={add}>
          Add entry
        </Button>
      </div>

      {entries.length === 0 ? (
        <p className="text-sm text-[var(--text-tertiary)]">
          No compensation recorded for this role yet.
        </p>
      ) : (
        <ul className="space-y-3">
          {entries.map((entry, index) => {
            const hasError = errorKeys.includes(entry.key);
            const errorId = `job-comp-${entry.key}-error`;
            const amountAria = hasError
              ? { "aria-invalid": true, "aria-describedby": errorId }
              : {};
            return (
              <li key={entry.key}>
                <fieldset
                  className="rounded-[var(--radius-md)] border border-[var(--border-subtle)]
                    bg-[var(--bg-surface)] p-3"
                >
                  <legend className="px-1 text-xs font-medium text-[var(--text-tertiary)]">
                    Entry {index + 1}
                  </legend>

                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                    <TextInput
                      label="Effective date"
                      type="date"
                      value={entry.effectiveDate}
                      onChange={(e) => update(entry.key, { effectiveDate: e.target.value })}
                    />
                    <Select
                      label="Pay basis"
                      value={entry.payBasis}
                      onChange={(e) =>
                        update(entry.key, {
                          payBasis: e.target.value as CompensationDraft["payBasis"],
                        })
                      }
                    >
                      {PAY_BASIS_OPTIONS.map((opt) => (
                        <option key={opt.value} value={opt.value}>
                          {opt.label}
                        </option>
                      ))}
                    </Select>
                    <CurrencyAutocomplete
                      label="Currency"
                      placeholder="e.g. GBP"
                      value={entry.currency}
                      onChange={(value) => update(entry.key, { currency: value })}
                    />
                  </div>

                  <div className="mt-3 grid grid-cols-2 gap-3 sm:grid-cols-4">
                    <TextInput
                      label="Base"
                      id={rowBaseId(entry.key)}
                      type="number"
                      min="0"
                      step="any"
                      inputMode="decimal"
                      value={entry.base}
                      onChange={(e) => update(entry.key, { base: e.target.value })}
                      {...amountAria}
                    />
                    <TextInput
                      label="Bonus"
                      type="number"
                      min="0"
                      step="any"
                      inputMode="decimal"
                      value={entry.bonus}
                      onChange={(e) => update(entry.key, { bonus: e.target.value })}
                      {...amountAria}
                    />
                    <TextInput
                      label="Equity"
                      type="number"
                      min="0"
                      step="any"
                      inputMode="decimal"
                      value={entry.equity}
                      onChange={(e) => update(entry.key, { equity: e.target.value })}
                      {...amountAria}
                    />
                    <TextInput
                      label="Other"
                      type="number"
                      min="0"
                      step="any"
                      inputMode="decimal"
                      value={entry.other}
                      onChange={(e) => update(entry.key, { other: e.target.value })}
                      {...amountAria}
                    />
                  </div>

                  {hasError && (
                    <p id={errorId} className="mt-2 text-xs text-[var(--color-error)]" role="alert">
                      Amounts must be valid, non-negative numbers.
                    </p>
                  )}

                  <div className="mt-3 flex items-end gap-3">
                    <div className="flex-1">
                      <TextInput
                        label="Note"
                        placeholder="e.g. promotion to Senior"
                        value={entry.note}
                        onChange={(e) => update(entry.key, { note: e.target.value })}
                      />
                    </div>
                    <Button
                      id={rowRemoveId(entry.key)}
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => remove(entry.key)}
                      aria-label={`Remove compensation entry ${index + 1}`}
                    >
                      Remove
                    </Button>
                  </div>
                </fieldset>
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
