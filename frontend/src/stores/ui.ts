import { create } from "zustand";
import { persist, createJSONStorage, devtools } from "zustand/middleware";
import type {} from "@redux-devtools/extension";

type Theme = "light" | "dark" | "system";
type ResolvedTheme = "light" | "dark";

function resolveTheme(theme: Theme): ResolvedTheme {
  if (theme === "system") {
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  }
  return theme;
}

interface UIState {
  sidebarOpen: boolean;
  modals: Record<string, boolean>;
  theme: Theme;

  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  openModal: (id: string) => void;
  closeModal: (id: string) => void;
  setTheme: (theme: Theme) => void;
  cycleTheme: () => void;
  resolvedTheme: () => ResolvedTheme;
}

export const useUIStore = create<UIState>()(
  devtools(
    persist(
      (set, get) => ({
        sidebarOpen: false,
        modals: {},
        theme: "system" as Theme,

        toggleSidebar: () =>
          set((s) => ({ sidebarOpen: !s.sidebarOpen }), false, "ui/toggleSidebar"),
        setSidebarOpen: (open) => set({ sidebarOpen: open }, false, "ui/setSidebarOpen"),
        openModal: (id) =>
          set((s) => ({ modals: { ...s.modals, [id]: true } }), false, "ui/openModal"),
        closeModal: (id) =>
          set((s) => ({ modals: { ...s.modals, [id]: false } }), false, "ui/closeModal"),
        setTheme: (theme) => set({ theme }, false, "ui/setTheme"),
        cycleTheme: () => {
          const order: Theme[] = ["light", "dark", "system"];
          const current = get().theme;
          const next = order[(order.indexOf(current) + 1) % order.length];
          set({ theme: next }, false, "ui/cycleTheme");
        },
        resolvedTheme: () => resolveTheme(get().theme),
      }),
      {
        name: "ui-theme",
        storage: createJSONStorage(() => localStorage),
        partialize: (state) => ({ theme: state.theme }),
      },
    ),
    {
      name: "ui-store",
      enabled: import.meta.env.DEV,
    },
  ),
);
