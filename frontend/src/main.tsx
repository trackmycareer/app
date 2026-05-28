import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router";
import { QueryProvider } from "@/providers/QueryProvider";
import { Toaster } from "sonner";
import { AwardCelebration } from "@/components/AwardCelebration";
import { router } from "@/router";
import "./App.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryProvider>
      <RouterProvider router={router} />
      <AwardCelebration />
      <Toaster
        position="top-right"
        theme="dark"
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
    </QueryProvider>
  </StrictMode>,
);
