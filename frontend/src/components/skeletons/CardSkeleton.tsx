import { Skeleton } from "@/components/skeletons/Skeleton";

export function CardSkeleton() {
  return (
    <div className="rounded-[var(--radius-xl)] border border-[var(--border-subtle)] bg-[var(--bg-surface)] p-5">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <Skeleton className="mb-2 h-5 w-48" />
          <Skeleton className="mb-3 h-4 w-32" />
          <Skeleton className="mb-1 h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </div>
      </div>
    </div>
  );
}
