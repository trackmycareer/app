import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";

const jobTitleKeys = {
  all: ["jobTitles"] as const,
  search: (q: string) => [...jobTitleKeys.all, "search", q] as const,
};

export function useJobTitleSearchQuery(query: string) {
  return useQuery({
    queryKey: jobTitleKeys.search(query),
    queryFn: () => apiClient.jobTitles.search(query).then((res) => res.data.data),
    enabled: query.length >= 2,
    staleTime: 5 * 60 * 1000,
  });
}
