import type { Compensation, CompensationAmounts, PayBasis } from "@/types";
import { majorToMinor, minorToMajor } from "@/lib/money";

// A CompensationDraft is the editable, string-based form of a compensation entry
// used inside the job form. `id` is set for entries that already exist on the
// server; `key` is a stable client-side identity for React lists.
export interface CompensationDraft {
  key: string;
  id?: string;
  effectiveDate: string;
  currency: string;
  payBasis: PayBasis;
  base: string;
  bonus: string;
  equity: string;
  other: string;
  note: string;
}

function todayISO(): string {
  return new Date().toISOString().slice(0, 10);
}

function minorToInput(minor: number): string {
  return minor ? String(minorToMajor(minor)) : "";
}

export function emptyDraft(): CompensationDraft {
  return {
    key: crypto.randomUUID(),
    effectiveDate: todayISO(),
    currency: "GBP",
    payBasis: "annual",
    base: "",
    bonus: "",
    equity: "",
    other: "",
    note: "",
  };
}

export function draftFromCompensation(c: Compensation): CompensationDraft {
  return {
    key: c.id,
    id: c.id,
    effectiveDate: c.effective_date.slice(0, 10),
    currency: c.currency,
    payBasis: c.pay_basis,
    base: minorToInput(c.amounts.base),
    bonus: minorToInput(c.amounts.bonus),
    equity: minorToInput(c.amounts.equity),
    other: minorToInput(c.amounts.other),
    note: c.amounts.note ?? "",
  };
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

// parseDraftAmounts returns the amounts in minor units, or null if any field is
// not a valid non-negative number.
export function parseDraftAmounts(d: CompensationDraft): CompensationAmounts | null {
  const base = parseAmount(d.base);
  const bonus = parseAmount(d.bonus);
  const equity = parseAmount(d.equity);
  const other = parseAmount(d.other);
  if (base === null || bonus === null || equity === null || other === null) return null;
  return { base, bonus, equity, other, note: d.note.trim() || null };
}

// isBlankNewDraft is true for a new row (no id) the user left effectively empty,
// so it can be skipped rather than created.
export function isBlankNewDraft(d: CompensationDraft): boolean {
  if (d.id) return false;
  return (
    d.base.trim() === "" &&
    d.bonus.trim() === "" &&
    d.equity.trim() === "" &&
    d.other.trim() === "" &&
    d.note.trim() === ""
  );
}

export function draftToPayload(d: CompensationDraft, jobId: string): Partial<Compensation> | null {
  const amounts = parseDraftAmounts(d);
  if (amounts === null) return null;
  return {
    job_id: jobId,
    effective_date: d.effectiveDate,
    currency: d.currency,
    pay_basis: d.payBasis,
    amounts,
  };
}
