import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";

const locationKeys = {
  all: ["locations"] as const,
  search: (q: string) => [...locationKeys.all, "search", q] as const,
};

export function useLocationSearchQuery(query: string) {
  return useQuery({
    queryKey: locationKeys.search(query),
    queryFn: () => apiClient.locations.search(query).then((res) => res.data.data),
    enabled: query.length >= 2,
    staleTime: 5 * 60 * 1000,
  });
}
