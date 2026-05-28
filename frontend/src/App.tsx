import { Outlet } from "react-router";
import { Sidebar } from "@/components/Sidebar";

export default function App() {
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
      <div className="flex flex-1 flex-col lg:ml-[var(--sidebar-width)]">
        <main id="main-content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
