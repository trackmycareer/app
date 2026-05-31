import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";

export const configKeys = {
  all: ["config"] as const,
};

export function useAppConfig() {
  return useQuery({
    queryKey: configKeys.all,
    queryFn: () => apiClient.config.get().then((res) => res.data.data),
    staleTime: 1000 * 60 * 30,
  });
}
