import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";

const certSearchKeys = {
  all: ["certSearch"] as const,
  search: (q: string) => [...certSearchKeys.all, "search", q] as const,
};

export function useCertSearchQuery(query: string) {
  return useQuery({
    queryKey: certSearchKeys.search(query),
    queryFn: () => apiClient.certifications.search(query).then((res) => res.data.data),
    enabled: query.length >= 2,
    staleTime: 5 * 60 * 1000,
  });
}
