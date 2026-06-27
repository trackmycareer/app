import { NavLink } from "react-router";
import { useAuthStore } from "@/stores/auth";
import { useUIStore } from "@/stores/ui";
import { NavItem } from "@/components/NavItem";
import { NavSection } from "@/components/NavSection";
import {
  LayoutDashboardIcon,
  BriefcaseIcon,
  TrophyIcon,
  BuildingIcon,
  ClipboardListIcon,
  AwardIcon,
  ZapIcon,
  UploadIcon,
  DownloadIcon,
  UsersIcon,
  SettingsIcon,
  LogOutIcon,
  CloseIcon,
  StarIcon,
  UserCircleIcon,
  ShieldIcon,
  HeartIcon,
  SunIcon,
  MoonIcon,
  MonitorIcon,
  BellIcon,
  WalletIcon,
} from "@/components/icons";
import { useAppConfig } from "@/hooks/queries/useConfigQuery";

export function Sidebar() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const sidebarOpen = useUIStore((s) => s.sidebarOpen);
  const setSidebarOpen = useUIStore((s) => s.setSidebarOpen);
  const theme = useUIStore((s) => s.theme);
  const cycleTheme = useUIStore((s) => s.cycleTheme);

  const themeIcon = { light: SunIcon, dark: MoonIcon, system: MonitorIcon }[theme];
  const themeLabel = { light: "Light", dark: "Dark", system: "System" }[theme];
  const ThemeIconComponent = themeIcon;

  const { data: appConfig } = useAppConfig();

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
          "border-r border-[var(--border-subtle)] bg-[var(--sidebar-bg)]",
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
            className="flex items-center justify-center rounded-[var(--radius-sm)] p-2 min-h-11 min-w-11
              text-[var(--text-tertiary)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)]
              lg:hidden"
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
              to="/compensation"
              icon={<WalletIcon width={18} height={18} />}
              label="Compensation"
              onClick={closeSidebar}
            />
            <NavItem
              to="/applications"
              icon={<ClipboardListIcon width={18} height={18} />}
              label="Applications"
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
            <NavItem
              to="/notifications"
              icon={<BellIcon width={18} height={18} />}
              label="Notifications"
              onClick={closeSidebar}
            />
          </NavSection>

          <NavSection title="Data">
            <NavItem
              to="/import"
              icon={<UploadIcon width={18} height={18} />}
              label="Import"
              onClick={closeSidebar}
            />
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
              to="/security"
              icon={<ShieldIcon width={18} height={18} />}
              label="Security"
              onClick={closeSidebar}
            />
          </NavSection>

          {appConfig?.polar_enabled && (
            <NavSection title="Support">
              {user?.is_one_time_supporter || user?.is_subscriber ? (
                <NavItem
                  to="/support"
                  icon={
                    <HeartIcon
                      width={18}
                      height={18}
                      fill="currentColor"
                      className="text-[var(--accent-warm)]"
                    />
                  }
                  label="Supporter"
                  onClick={closeSidebar}
                />
              ) : (
                <NavItem
                  to="/support"
                  icon={<HeartIcon width={18} height={18} />}
                  label="Support"
                  onClick={closeSidebar}
                />
              )}
            </NavSection>
          )}

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

        {/* Theme toggle */}
        <div className="border-t border-[var(--border-subtle)] px-3 py-2">
          <button
            onClick={cycleTheme}
            className="flex w-full items-center gap-2.5 rounded-[var(--radius-md)] px-2 py-1.5
              text-[var(--text-tertiary)] transition-colors hover:bg-[var(--bg-hover)]
              hover:text-[var(--text-primary)]"
            title={`Theme: ${themeLabel}. Click to change.`}
          >
            <ThemeIconComponent width={15} height={15} />
            <span className="text-xs font-medium">{themeLabel}</span>
          </button>
        </div>

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
