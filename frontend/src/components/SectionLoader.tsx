import { SpinnerIcon } from "@/components/icons";

export function SectionLoader({ label }: { label: string }) {
  return (
    <div className="flex items-center justify-center py-10" role="status">
      <SpinnerIcon
        width={22}
        height={22}
        className="animate-spin text-[var(--accent-default)]"
        aria-hidden="true"
      />
      <span className="sr-only">Loading {label}...</span>
    </div>
  );
}
