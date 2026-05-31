import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";

const companyKeys = {
  all: ["companies"] as const,
  search: (q: string) => [...companyKeys.all, "search", q] as const,
};

export function useCompanySearchQuery(query: string) {
  return useQuery({
    queryKey: companyKeys.search(query),
    queryFn: () => apiClient.companies.search(query).then((res) => res.data.data),
    enabled: query.length >= 2,
    staleTime: 5 * 60 * 1000,
  });
}
