import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";

const skillSearchKeys = {
  all: ["skillSearch"] as const,
  search: (q: string) => [...skillSearchKeys.all, "search", q] as const,
};

export function useSkillSearchQuery(query: string) {
  return useQuery({
    queryKey: skillSearchKeys.search(query),
    queryFn: () => apiClient.skills.search(query).then((res) => res.data.data),
    enabled: query.length >= 2,
    staleTime: 5 * 60 * 1000,
  });
}
