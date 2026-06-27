import { useUIStore } from "@/stores/ui";
import { MenuIcon } from "@/components/icons";
import { NotificationBell } from "@/components/NotificationBell";

interface TopbarProps {
  title?: string;
}

export function Topbar({ title }: TopbarProps) {
  const toggleSidebar = useUIStore((s) => s.toggleSidebar);

  return (
    <header
      className="sticky top-0 z-30 flex h-[var(--topbar-height)] items-center gap-3 border-b
        border-[var(--border-subtle)] bg-[var(--bg-surface)] px-4 lg:px-6"
    >
      <button
        onClick={toggleSidebar}
        className="rounded-[var(--radius-sm)] p-1.5 min-h-11 min-w-11 text-[var(--text-secondary)]
          hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)] lg:hidden"
        aria-label="Toggle sidebar"
      >
        <MenuIcon width={20} height={20} />
      </button>
      {title && (
        <h1 className="animate-fade-in text-lg font-semibold text-[var(--text-primary)]">
          {title}
        </h1>
      )}
      <div className="ml-auto">
        <NotificationBell />
      </div>
    </header>
  );
}
