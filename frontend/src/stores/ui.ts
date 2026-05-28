import { create } from "zustand";
import { devtools } from "zustand/middleware";
import type {} from "@redux-devtools/extension";

interface UIState {
  sidebarOpen: boolean;
  modals: Record<string, boolean>;

  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  openModal: (id: string) => void;
  closeModal: (id: string) => void;
}

export const useUIStore = create<UIState>()(
  devtools(
    (set) => ({
      sidebarOpen: false,
      modals: {},

      toggleSidebar: () =>
        set((s) => ({ sidebarOpen: !s.sidebarOpen }), false, "ui/toggleSidebar"),
      setSidebarOpen: (open) => set({ sidebarOpen: open }, false, "ui/setSidebarOpen"),
      openModal: (id) =>
        set((s) => ({ modals: { ...s.modals, [id]: true } }), false, "ui/openModal"),
      closeModal: (id) =>
        set((s) => ({ modals: { ...s.modals, [id]: false } }), false, "ui/closeModal"),
    }),
    {
      name: "ui-store",
      enabled: import.meta.env.DEV,
    },
  ),
);
