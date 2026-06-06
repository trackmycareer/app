import { StrictMode, Suspense, lazy } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router";
import { QueryProvider } from "@/providers/QueryProvider";
import { Toaster } from "sonner";
import { AwardCelebration } from "@/components/AwardCelebration";
import { useUIStore } from "@/stores/ui";
import { APP_DOMAIN } from "@/lib/constants";
import { router } from "@/router";
import "./App.css";

const CustomDomainProfile = lazy(() => import("@/pages/profile/CustomDomainProfile"));

const isCustomDomain =
  window.location.hostname !== APP_DOMAIN &&
  !window.location.hostname.endsWith(`.${APP_DOMAIN}`) &&
  window.location.hostname !== "localhost" &&
  window.location.hostname !== "127.0.0.1";

function ThemeAwareToaster() {
  const resolvedTheme = useUIStore((s) => s.resolvedTheme)();

  return (
    <Toaster
      position="top-right"
      theme={resolvedTheme}
      toastOptions={{
        style: {
          background: "var(--bg-elevated)",
          border: "1px solid var(--border-default)",
          color: "var(--text-primary)",
        },
        classNames: {
          success: "!border-l-4 !border-l-[var(--color-success)]",
          error: "!border-l-4 !border-l-[var(--color-error)]",
        },
      }}
    />
  );
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryProvider>
      {isCustomDomain ? (
        <Suspense
          fallback={<div className="flex min-h-screen items-center justify-center">Loading...</div>}
        >
          <CustomDomainProfile />
        </Suspense>
      ) : (
        <>
          <RouterProvider router={router} />
          <AwardCelebration />
        </>
      )}
      <ThemeAwareToaster />
    </QueryProvider>
  </StrictMode>,
);
