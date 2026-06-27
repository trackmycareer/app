import type { CompensationAmounts, PayBasis } from "@/types";

// The backend stores money as integer minor units (for example pennies) so totals
// stay exact. The UI works in major units (pounds) and formats with Intl.

export function minorToMajor(minor: number): number {
  return minor / 100;
}

export function majorToMinor(major: number): number {
  return Math.round(major * 100);
}

// formatMoney renders a minor-unit amount in its currency. Unknown or malformed
// currency codes fall back to a plain number with the code appended rather than
// throwing.
export function formatMoney(minor: number, currency: string): string {
  const major = minorToMajor(minor);
  try {
    return new Intl.NumberFormat("en-GB", {
      style: "currency",
      currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(major);
  } catch {
    return `${major.toLocaleString("en-GB", { maximumFractionDigits: 0 })} ${currency}`;
  }
}

export function totalMinor(amounts: CompensationAmounts): number {
  return amounts.base + amounts.bonus + amounts.equity + amounts.other;
}

export const PAY_BASIS_LABELS: Record<PayBasis, string> = {
  annual: "Annual",
  monthly: "Monthly",
  weekly: "Weekly",
  daily: "Daily",
  hourly: "Hourly",
  one_off: "One-off",
};

export const PAY_BASIS_OPTIONS: { value: PayBasis; label: string }[] = (
  Object.keys(PAY_BASIS_LABELS) as PayBasis[]
).map((value) => ({ value, label: PAY_BASIS_LABELS[value] }));
