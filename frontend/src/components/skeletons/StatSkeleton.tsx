import { Skeleton } from "@/components/skeletons/Skeleton";

export function StatSkeleton() {
  return (
    <div className="rounded-[var(--radius-xl)] border border-[var(--border-subtle)] bg-[var(--bg-surface)] p-5">
      <Skeleton className="mb-3 h-4 w-24" />
      <Skeleton className="h-7 w-16" />
    </div>
  );
}
