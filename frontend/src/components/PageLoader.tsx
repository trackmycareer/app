import { SpinnerIcon } from "@/components/icons";

export function PageLoader() {
  return (
    <div
      role="status"
      aria-label="Loading"
      className="flex h-screen w-full items-center justify-center bg-[var(--bg-base)]"
    >
      <SpinnerIcon
        width={32}
        height={32}
        aria-hidden="true"
        className="animate-spin text-[var(--accent-default)] motion-reduce:animate-none"
      />
      <span className="sr-only">Loading...</span>
    </div>
  );
}
