import { useMemo } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import type { Compensation } from "@/types";
import { formatMoney, majorToMinor, minorToMajor, totalMinor } from "@/lib/money";

interface CompensationTrendChartProps {
  // Entries must all share one currency; totals across currencies are meaningless.
  entries: Compensation[];
  currency: string;
}

function formatMonth(date: string): string {
  return new Date(date).toLocaleDateString("en-GB", { month: "short", year: "numeric" });
}

function compactCurrency(value: number, currency: string): string {
  try {
    return new Intl.NumberFormat("en-GB", {
      style: "currency",
      currency,
      notation: "compact",
      maximumFractionDigits: 1,
    }).format(value);
  } catch {
    return String(value);
  }
}

export function CompensationTrendChart({ entries, currency }: CompensationTrendChartProps) {
  const data = useMemo(
    () =>
      entries
        .slice()
        .sort((a, b) => a.effective_date.localeCompare(b.effective_date))
        .map((e) => ({
          date: formatMonth(e.effective_date),
          total: minorToMajor(totalMinor(e.amounts)),
        })),
    [entries],
  );

  return (
    <ResponsiveContainer width="100%" height={260}>
      <LineChart data={data} margin={{ top: 8, right: 16, bottom: 4, left: 8 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="var(--border-subtle)" />
        <XAxis
          dataKey="date"
          tick={{ fill: "var(--text-tertiary)", fontSize: 12 }}
          stroke="var(--border-default)"
        />
        <YAxis
          width={64}
          tick={{ fill: "var(--text-tertiary)", fontSize: 12 }}
          stroke="var(--border-default)"
          tickFormatter={(value: number) => compactCurrency(value, currency)}
        />
        <Tooltip
          formatter={(value) => formatMoney(majorToMinor(Number(value)), currency)}
          contentStyle={{
            backgroundColor: "var(--bg-elevated)",
            border: "1px solid var(--border-default)",
            borderRadius: "8px",
            color: "var(--text-primary)",
          }}
          labelStyle={{ color: "var(--text-secondary)" }}
          cursor={{ stroke: "var(--border-default)" }}
        />
        <Line
          type="monotone"
          dataKey="total"
          name="Total compensation"
          stroke="var(--accent-default)"
          strokeWidth={2}
          dot={{ r: 3, fill: "var(--accent-default)" }}
          activeDot={{ r: 5 }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
