import { NavLink } from "react-router";
import type { ReactNode } from "react";

export interface NavItemProps {
  to: string;
  icon: ReactNode;
  label: string;
  end?: boolean;
  onClick?: () => void;
}

export function NavItem({ to, icon, label, end, onClick }: NavItemProps) {
  return (
    <NavLink
      to={to}
      end={end}
      onClick={onClick}
      className={({ isActive }) =>
        [
          "flex items-center gap-3 rounded-[var(--radius-md)] px-3 py-2 text-sm",
          "transition-all duration-150",
          isActive
            ? "border-l-2 border-[var(--accent-default)] bg-[var(--accent-subtle)] " +
              "text-[var(--accent-text)] font-medium"
            : "text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] " +
              "hover:text-[var(--text-primary)] hover:translate-x-0.5",
        ].join(" ")
      }
    >
      <span className="flex-shrink-0">{icon}</span>
      <span className="truncate lg:inline">{label}</span>
    </NavLink>
  );
}
