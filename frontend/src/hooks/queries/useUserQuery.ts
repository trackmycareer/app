import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";

export const userKeys = {
  all: ["user"] as const,
  me: () => [...userKeys.all, "me"] as const,
};

export const useCurrentUserQuery = () => {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  return useQuery({
    queryKey: userKeys.me(),
    queryFn: () => apiClient.user.getCurrent().then((res) => res.data.data),
    enabled: isAuthenticated,
  });
};
