import { NavLink } from "react-router";
import type { ReactNode } from "react";
import { useAuthStore } from "@/stores/auth";
import { useUIStore } from "@/stores/ui";
import {
  LayoutDashboardIcon,
  BriefcaseIcon,
  TrophyIcon,
  BuildingIcon,
  AwardIcon,
  ZapIcon,
  DownloadIcon,
  UsersIcon,
  SettingsIcon,
  LogOutIcon,
  CloseIcon,
  StarIcon,
  UserCircleIcon,
} from "@/components/icons";

interface NavItemProps {
  to: string;
  icon: ReactNode;
  label: string;
  end?: boolean;
  onClick?: () => void;
}

function NavItem({ to, icon, label, end, onClick }: NavItemProps) {
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

function NavSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="space-y-1">
      <p className="px-3 pb-1 text-xs font-semibold uppercase tracking-wider text-[var(--text-tertiary)]">
        {title}
      </p>
      {children}
    </div>
  );
}

export function Sidebar() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const sidebarOpen = useUIStore((s) => s.sidebarOpen);
  const setSidebarOpen = useUIStore((s) => s.setSidebarOpen);

  const closeSidebar = () => setSidebarOpen(false);

  const initials = user?.name
    ? user.name
        .split(" ")
        .map((n) => n[0])
        .join("")
        .toUpperCase()
        .slice(0, 2)
    : "?";

  return (
    <>
      {/* Mobile overlay */}
      {sidebarOpen && (
        <button
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={closeSidebar}
          aria-label="Close sidebar"
        />
      )}

      {/* Sidebar */}
      <aside
        className={[
          "fixed inset-y-0 left-0 z-50 flex w-[var(--sidebar-width)] flex-col",
          "border-r border-[var(--border-subtle)] bg-[var(--bg-surface)]",
          "transition-transform duration-200 lg:translate-x-0",
          sidebarOpen ? "translate-x-0" : "-translate-x-full",
        ].join(" ")}
      >
        {/* Logo */}
        <div
          className="flex h-[var(--topbar-height)] items-center justify-between border-b
            border-[var(--border-subtle)] px-4"
        >
          <NavLink to="/" className="flex items-center gap-2.5" onClick={closeSidebar}>
            <BriefcaseIcon className="text-[var(--accent-default)]" width={22} height={22} />
            <span className="text-base font-semibold text-[var(--text-primary)]">
              trackmy<span className="text-[var(--accent-default)]">.</span>career
            </span>
          </NavLink>
          <button
            className="rounded-[var(--radius-sm)] p-1 text-[var(--text-tertiary)]
              hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)] lg:hidden"
            onClick={closeSidebar}
            aria-label="Close sidebar"
          >
            <CloseIcon width={18} height={18} />
          </button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-6 overflow-y-auto px-3 py-4" aria-label="Main navigation">
          <NavSection title="Main">
            <NavItem
              to="/"
              end
              icon={<LayoutDashboardIcon width={18} height={18} />}
              label="Dashboard"
              onClick={closeSidebar}
            />
            <NavItem
              to="/wins"
              icon={<TrophyIcon width={18} height={18} />}
              label="Wins"
              onClick={closeSidebar}
            />
            <NavItem
              to="/jobs"
              icon={<BuildingIcon width={18} height={18} />}
              label="Jobs"
              onClick={closeSidebar}
            />
            <NavItem
              to="/certifications"
              icon={<AwardIcon width={18} height={18} />}
              label="Certifications"
              onClick={closeSidebar}
            />
            <NavItem
              to="/skills"
              icon={<ZapIcon width={18} height={18} />}
              label="Skills"
              onClick={closeSidebar}
            />
            <NavItem
              to="/achievements"
              icon={<StarIcon width={18} height={18} />}
              label="Achievements"
              onClick={closeSidebar}
            />
          </NavSection>

          <NavSection title="Data">
            <NavItem
              to="/export"
              icon={<DownloadIcon width={18} height={18} />}
              label="Export"
              onClick={closeSidebar}
            />
            <NavItem
              to="/profile"
              icon={<UserCircleIcon width={18} height={18} />}
              label="Profile"
              onClick={closeSidebar}
            />
            <NavItem
              to="/settings"
              icon={<SettingsIcon width={18} height={18} />}
              label="Settings"
              onClick={closeSidebar}
            />
          </NavSection>

          {user?.is_admin && (
            <NavSection title="Admin">
              <NavItem
                to="/admin/users"
                icon={<UsersIcon width={18} height={18} />}
                label="Users"
                onClick={closeSidebar}
              />
              <NavItem
                to="/admin/settings"
                icon={<SettingsIcon width={18} height={18} />}
                label="App Settings"
                onClick={closeSidebar}
              />
              <NavItem
                to="/admin/badges"
                icon={<AwardIcon width={18} height={18} />}
                label="Badges"
                onClick={closeSidebar}
              />
            </NavSection>
          )}
        </nav>

        {/* User footer */}
        <div className="border-t border-[var(--border-subtle)] p-3">
          <div className="flex items-center gap-3">
            {user?.avatar_url ? (
              <img
                src={user.avatar_url}
                alt={user.name}
                className="h-8 w-8 rounded-full object-cover"
              />
            ) : (
              <div
                className="flex h-8 w-8 items-center justify-center rounded-full
                  bg-[var(--accent-muted)] text-xs font-semibold text-[var(--accent-text)]"
              >
                {initials}
              </div>
            )}
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium text-[var(--text-primary)]">
                {user?.name}
              </p>
              <p className="truncate text-xs text-[var(--text-tertiary)]">{user?.email}</p>
            </div>
            <button
              onClick={logout}
              className="rounded-[var(--radius-sm)] p-1.5 text-[var(--text-tertiary)]
                transition-colors hover:bg-[var(--bg-hover)] hover:text-[var(--color-error)]"
              aria-label="Sign out"
              title="Sign out"
            >
              <LogOutIcon width={16} height={16} />
            </button>
          </div>
        </div>
      </aside>
    </>
  );
}
