import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient } from "@/lib/api";
import type { Compensation } from "@/types";

export const compensationKeys = {
  all: ["compensation"] as const,
  lists: () => [...compensationKeys.all, "list"] as const,
  list: () => [...compensationKeys.lists()] as const,
  details: () => [...compensationKeys.all, "detail"] as const,
  detail: (id: string) => [...compensationKeys.details(), id] as const,
};

export function useCompensationQuery(options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: compensationKeys.list(),
    queryFn: () => apiClient.compensation.list().then((res) => res.data.data),
    enabled: options?.enabled ?? true,
  });
}

export function useCompensationEntryQuery(id: string) {
  return useQuery({
    queryKey: compensationKeys.detail(id),
    queryFn: () => apiClient.compensation.getById(id).then((res) => res.data.data),
    enabled: !!id,
  });
}

export function useCreateCompensationMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Partial<Compensation>) =>
      apiClient.compensation.create(data).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: compensationKeys.lists() });
      toast.success("Compensation entry added");
    },
  });
}

export function useUpdateCompensationMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Compensation> }) =>
      apiClient.compensation.update(id, data).then((res) => res.data.data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: compensationKeys.lists() });
      queryClient.invalidateQueries({ queryKey: compensationKeys.detail(variables.id) });
      toast.success("Compensation entry updated");
    },
  });
}

export function useDeleteCompensationMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.compensation.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: compensationKeys.lists() });
      toast.success("Compensation entry deleted");
    },
  });
}
