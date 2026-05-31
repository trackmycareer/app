export function SectionError({ message }: { message: string }) {
  return (
    <div
      className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
        bg-[var(--color-error)]/5 p-3 text-center text-sm text-[var(--color-error)]"
      role="alert"
    >
      {message}
    </div>
  );
}
