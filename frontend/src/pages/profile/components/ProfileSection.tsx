interface ProfileSectionProps {
  title: string;
  icon: React.ReactNode;
  children: React.ReactNode;
}

export function ProfileSection({ title, icon, children }: ProfileSectionProps) {
  return (
    <section aria-label={title}>
      <div className="mb-3 flex items-center gap-2">
        <span className="text-[var(--text-tertiary)]">{icon}</span>
        <h2 className="text-sm font-semibold uppercase tracking-wider text-[var(--text-tertiary)]">
          {title}
        </h2>
      </div>
      {children}
    </section>
  );
}
