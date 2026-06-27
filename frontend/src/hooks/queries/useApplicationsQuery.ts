import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient } from "@/lib/api";
import type { Application, ApplicationStatus } from "@/types";

export const applicationKeys = {
  all: ["applications"] as const,
  lists: () => [...applicationKeys.all, "list"] as const,
  list: () => [...applicationKeys.lists()] as const,
  details: () => [...applicationKeys.all, "detail"] as const,
  detail: (id: string) => [...applicationKeys.details(), id] as const,
};

export function useApplicationsQuery() {
  return useQuery({
    queryKey: applicationKeys.list(),
    queryFn: () => apiClient.applications.list().then((res) => res.data.data),
  });
}

export function useApplicationQuery(id: string) {
  return useQuery({
    queryKey: applicationKeys.detail(id),
    queryFn: () => apiClient.applications.getById(id).then((res) => res.data.data),
    enabled: !!id,
  });
}

export function useCreateApplicationMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Partial<Application>) =>
      apiClient.applications.create(data).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: applicationKeys.lists() });
      toast.success("Application added");
    },
  });
}

export function useUpdateApplicationMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Application> }) =>
      apiClient.applications.update(id, data).then((res) => res.data.data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: applicationKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: applicationKeys.detail(variables.id),
      });
      toast.success("Application updated");
    },
  });
}

interface MoveVariables {
  id: string;
  status: ApplicationStatus;
  ordered_ids?: string[];
}

// Optimistically reflects a drag (or status-select change) in the board cache so
// the card stays where it was dropped, rolling back if the request fails.
export function useMoveApplicationMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, status, ordered_ids }: MoveVariables) =>
      apiClient.applications.move(id, { status, ordered_ids }).then((res) => res.data.data),
    onMutate: async ({ id, status, ordered_ids }) => {
      await queryClient.cancelQueries({ queryKey: applicationKeys.list() });
      const previous = queryClient.getQueryData<Application[]>(applicationKeys.list());
      if (previous) {
        const next = previous.map((a) => {
          if (ordered_ids && ordered_ids.includes(a.id)) {
            return { ...a, status, sort_order: ordered_ids.indexOf(a.id) };
          }
          if (a.id === id) {
            return { ...a, status };
          }
          return a;
        });
        queryClient.setQueryData(applicationKeys.list(), next);
      }
      return { previous };
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) {
        queryClient.setQueryData(applicationKeys.list(), context.previous);
      }
      toast.error("Failed to move application");
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: applicationKeys.lists() });
    },
  });
}

export function useDeleteApplicationMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.applications.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: applicationKeys.lists() });
      toast.success("Application deleted");
    },
  });
}
