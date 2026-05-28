import { create } from "zustand";
import { devtools } from "zustand/middleware";
import type {} from "@redux-devtools/extension";
import type { Award } from "@/types";

interface AwardsState {
  pending: Award[];
  push: (awards: Award[]) => void;
  shift: () => Award | undefined;
  clear: () => void;
}

export const useAwardsStore = create<AwardsState>()(
  devtools(
    (set, get) => ({
      pending: [],

      push: (awards: Award[]) => {
        if (awards.length === 0) return;
        set(
          (s) => ({ pending: [...s.pending, ...awards] }),
          false,
          "awards/push",
        );
      },

      shift: () => {
        const current = get().pending;
        if (current.length === 0) return undefined;
        const [first, ...rest] = current;
        set({ pending: rest }, false, "awards/shift");
        return first;
      },

      clear: () => set({ pending: [] }, false, "awards/clear"),
    }),
    {
      name: "awards-store",
      enabled: import.meta.env.DEV,
    },
  ),
);
