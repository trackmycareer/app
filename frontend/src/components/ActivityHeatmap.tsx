import { useMemo, useState } from "react";
import type { HeatmapEntry } from "@/types";

interface ActivityHeatmapProps {
  data: HeatmapEntry[];
  days?: number;
}

const DAY_LABELS = ["", "Mon", "", "Wed", "", "Fri", ""];

function getIntensityClass(count: number): string {
  if (count === 0) return "bg-[var(--heatmap-0)]";
  if (count === 1) return "bg-[var(--heatmap-1)]";
  if (count <= 3) return "bg-[var(--heatmap-2)]";
  if (count <= 6) return "bg-[var(--heatmap-3)]";
  return "bg-[var(--heatmap-4)]";
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("en-GB", {
    weekday: "short",
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}

export function ActivityHeatmap({ data, days = 365 }: ActivityHeatmapProps) {
  const [tooltip, setTooltip] = useState<{
    text: string;
    x: number;
    y: number;
  } | null>(null);

  const { weeks, monthLabels } = useMemo(() => {
    const countMap = new Map<string, number>();
    for (const entry of data) {
      countMap.set(entry.date, entry.count);
    }

    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const startDate = new Date(today);
    startDate.setDate(startDate.getDate() - days + 1);
    // Align to start of week (Monday)
    const dayOfWeek = startDate.getDay();
    const mondayOffset = dayOfWeek === 0 ? -6 : 1 - dayOfWeek;
    startDate.setDate(startDate.getDate() + mondayOffset);

    const weeksArr: { date: string; count: number; dayOfWeek: number }[][] = [];
    const months: { label: string; col: number }[] = [];
    let lastMonth = -1;
    const cursor = new Date(startDate);

    while (cursor <= today) {
      const week: { date: string; count: number; dayOfWeek: number }[] = [];
      for (let d = 0; d < 7; d++) {
        const dateStr = cursor.toISOString().slice(0, 10);
        const isInRange = cursor <= today;
        week.push({
          date: dateStr,
          count: isInRange ? (countMap.get(dateStr) ?? 0) : -1,
          dayOfWeek: d,
        });

        if (d === 0 && cursor.getMonth() !== lastMonth) {
          lastMonth = cursor.getMonth();
          months.push({
            label: cursor.toLocaleDateString("en-GB", { month: "short" }),
            col: weeksArr.length,
          });
        }

        cursor.setDate(cursor.getDate() + 1);
      }
      weeksArr.push(week);
    }

    return { weeks: weeksArr, monthLabels: months };
  }, [data, days]);

  const totalContributions = useMemo(
    () => data.reduce((sum, e) => sum + e.count, 0),
    [data],
  );

  return (
    <div className="relative">
      <div className="mb-3 flex items-center justify-between">
        <p className="text-xs text-[var(--text-tertiary)]">
          {totalContributions} activit{totalContributions === 1 ? "y" : "ies"} in the last{" "}
          {days} days
        </p>
      </div>

      <div className="overflow-x-auto pb-2" onTouchStart={() => setTooltip(null)}>
        <div className="inline-flex flex-col gap-1" role="img" aria-label="Activity heatmap">
          {/* Month labels */}
          <div className="flex gap-[2px] pl-6 sm:gap-[3px] sm:pl-8">
            {weeks.map((_, weekIdx) => {
              const monthEntry = monthLabels.find((m) => m.col === weekIdx);
              return (
                <div key={weekIdx} className="h-3 w-[9px] text-center sm:w-[13px]">
                  {monthEntry && (
                    <span className="text-[10px] leading-3 text-[var(--text-tertiary)]">
                      {monthEntry.label}
                    </span>
                  )}
                </div>
              );
            })}
          </div>

          {/* Grid rows */}
          {[0, 1, 2, 3, 4, 5, 6].map((dayIdx) => (
            <div key={dayIdx} className="flex items-center gap-[2px] sm:gap-[3px]">
              <span className="w-5 text-right text-[8px] text-[var(--text-tertiary)] sm:w-7 sm:text-[10px]">
                {DAY_LABELS[dayIdx]}
              </span>
              {weeks.map((week, weekIdx) => {
                const cell = week[dayIdx];
                if (!cell || cell.count < 0) {
                  return (
                    <div
                      key={weekIdx}
                      className="h-[9px] w-[9px] rounded-sm sm:h-[13px] sm:w-[13px]"
                    />
                  );
                }
                return (
                  <div
                    key={weekIdx}
                    className={`h-[9px] w-[9px] rounded-sm sm:h-[13px] sm:w-[13px] ${getIntensityClass(cell.count)}`}
                    onMouseEnter={(e) => {
                      const rect = e.currentTarget.getBoundingClientRect();
                      setTooltip({
                        text: `${cell.count} activit${cell.count === 1 ? "y" : "ies"} on ${formatDate(cell.date)}`,
                        x: rect.left + rect.width / 2,
                        y: rect.top,
                      });
                    }}
                    onMouseLeave={() => setTooltip(null)}
                    onClick={(e) => {
                      const rect = e.currentTarget.getBoundingClientRect();
                      setTooltip((prev) =>
                        prev && prev.text.includes(cell.date)
                          ? null
                          : {
                              text: `${cell.count} activit${cell.count === 1 ? "y" : "ies"} on ${formatDate(cell.date)}`,
                              x: rect.left + rect.width / 2,
                              y: rect.top,
                            }
                      );
                    }}
                    role="presentation"
                  />
                );
              })}
            </div>
          ))}
        </div>
      </div>

      {/* Tooltip */}
      {tooltip && (
        <div
          className="pointer-events-none fixed z-50 -translate-x-1/2 -translate-y-full
            rounded-[var(--radius-md)] bg-[var(--bg-elevated)] px-2.5 py-1.5
            text-xs text-[var(--text-primary)] shadow-lg"
          style={{ left: tooltip.x, top: tooltip.y - 6 }}
        >
          {tooltip.text}
        </div>
      )}

      {/* Legend */}
      <div className="mt-2 flex items-center justify-end gap-1.5 text-[10px] text-[var(--text-tertiary)]">
        <span>Less</span>
        <div className="h-[9px] w-[9px] sm:h-[11px] sm:w-[11px] rounded-sm bg-[var(--heatmap-0)]" />
        <div className="h-[9px] w-[9px] sm:h-[11px] sm:w-[11px] rounded-sm bg-[var(--heatmap-1)]" />
        <div className="h-[9px] w-[9px] sm:h-[11px] sm:w-[11px] rounded-sm bg-[var(--heatmap-2)]" />
        <div className="h-[9px] w-[9px] sm:h-[11px] sm:w-[11px] rounded-sm bg-[var(--heatmap-4)]" />
        <span>More</span>
      </div>
    </div>
  );
}
