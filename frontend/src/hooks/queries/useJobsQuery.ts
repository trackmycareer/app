import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient } from "@/lib/api";
import type { Job } from "@/types";

export const jobKeys = {
  all: ["jobs"] as const,
  lists: () => [...jobKeys.all, "list"] as const,
  list: () => [...jobKeys.lists()] as const,
  details: () => [...jobKeys.all, "detail"] as const,
  detail: (id: string) => [...jobKeys.details(), id] as const,
};

export function useJobsQuery() {
  return useQuery({
    queryKey: jobKeys.list(),
    queryFn: () =>
      apiClient.jobs.list().then((res) => res.data.data),
  });
}

export function useJobQuery(id: string) {
  return useQuery({
    queryKey: jobKeys.detail(id),
    queryFn: () => apiClient.jobs.getById(id).then((res) => res.data.data),
    enabled: !!id,
  });
}

export function useCreateJobMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Partial<Job>) =>
      apiClient.jobs.create(data).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: jobKeys.lists() });
      toast.success("Role added successfully");
    },
  });
}

export function useUpdateJobMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: Partial<Job>;
    }) => apiClient.jobs.update(id, data).then((res) => res.data.data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: jobKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: jobKeys.detail(variables.id),
      });
      toast.success("Role updated successfully");
    },
  });
}

export function useDeleteJobMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.jobs.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: jobKeys.lists() });
      toast.success("Role deleted");
    },
  });
}
