import { useEffect } from "react";
import { Outlet } from "react-router";
import { Sidebar } from "@/components/Sidebar";
import { useUIStore } from "@/stores/ui";

export default function App() {
  const theme = useUIStore((s) => s.theme);
  const resolvedTheme = useUIStore((s) => s.resolvedTheme)();

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", resolvedTheme);
  }, [resolvedTheme]);

  useEffect(() => {
    if (theme !== "system") return;

    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = () => {
      const resolved = mq.matches ? "dark" : "light";
      document.documentElement.setAttribute("data-theme", resolved);
    };
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, [theme]);

  return (
    <div className="flex min-h-screen bg-[var(--bg-base)]">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[60]
          focus:rounded-[var(--radius-md)] focus:bg-[var(--accent-default)] focus:px-4 focus:py-2
          focus:text-sm focus:font-medium focus:text-white"
      >
        Skip to main content
      </a>
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col lg:ml-[var(--sidebar-width)]">
        <main id="main-content">
          <Outlet />
        </main>
        <footer className="mt-auto border-t border-[var(--border-subtle)] px-4 py-4 text-center text-xs text-[var(--text-tertiary)]">
          <p>&copy; 2026 BH Cloud Labs Ltd. All rights reserved.</p>
          <p className="mt-1">
            BH Cloud Labs Ltd, trading as trackmy.career is registered in England and Wales (No.
            16211348).
          </p>
          <p>Registered Address: The Grange, Grange Road, Great Malvern. WR14 3HA.</p>
        </footer>
      </div>
    </div>
  );
}
