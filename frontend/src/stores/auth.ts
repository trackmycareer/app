import { create } from "zustand";
import { persist, createJSONStorage, devtools } from "zustand/middleware";
import type {} from "@redux-devtools/extension";
import type { User } from "@/types";
import { apiClient } from "@/lib/api";

interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  login: (user: User, token: string) => void;
  logout: () => Promise<void>;
  setAccessToken: (token: string) => void;
  updateUser: (updates: Partial<User>) => void;
  setLoading: (loading: boolean) => void;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      (set, get) => ({
        user: null,
        accessToken: null,
        isAuthenticated: false,
        isLoading: true,

        login: (user: User, accessToken: string) => {
          set(
            { user, accessToken, isAuthenticated: true, isLoading: false },
            false,
            "auth/login",
          );
        },

        logout: async () => {
          try {
            await apiClient.auth.logout();
          } catch {
            // Proceed with local logout even if server call fails
          }
          set(
            { user: null, accessToken: null, isAuthenticated: false, isLoading: false },
            false,
            "auth/logout",
          );
        },

        setAccessToken: (accessToken: string) => {
          set({ accessToken }, false, "auth/setAccessToken");
        },

        updateUser: (updates: Partial<User>) => {
          const currentUser = get().user;
          if (currentUser) {
            set({ user: { ...currentUser, ...updates } }, false, "auth/updateUser");
          }
        },

        setLoading: (loading: boolean) => {
          set({ isLoading: loading }, false, "auth/setLoading");
        },
      }),
      {
        name: "auth-storage",
        storage: createJSONStorage(() => localStorage),
        partialize: (state) => ({
          user: state.user,
          isAuthenticated: state.isAuthenticated,
        }),
        onRehydrateStorage: () => (state) => {
          state?.setLoading(false);
        },
      },
    ),
    {
      name: "auth-store",
      enabled: import.meta.env.DEV,
    },
  ),
);
