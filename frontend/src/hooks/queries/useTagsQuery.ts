import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";

export const tagKeys = {
  all: ["tags"] as const,
  list: () => [...tagKeys.all, "list"] as const,
};

export function useTagsQuery() {
  return useQuery({
    queryKey: tagKeys.list(),
    queryFn: () => apiClient.tags.list().then((res) => res.data.data),
  });
}

export function useCreateTagMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { name: string; colour: string }) =>
      apiClient.tags.create(data).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: tagKeys.list() });
    },
  });
}
