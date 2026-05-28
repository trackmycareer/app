import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient } from "@/lib/api";

export interface WinListParams {
  search?: string;
  category?: string;
  tag_id?: string;
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
}

export const winKeys = {
  all: ["wins"] as const,
  lists: () => [...winKeys.all, "list"] as const,
  list: (params: WinListParams) => [...winKeys.lists(), params] as const,
  details: () => [...winKeys.all, "detail"] as const,
  detail: (id: string) => [...winKeys.details(), id] as const,
};

export function useWinsQuery(params: WinListParams = {}) {
  return useQuery({
    queryKey: winKeys.list(params),
    queryFn: () =>
      apiClient.wins
        .list(params as Record<string, string | number>)
        .then((res) => res.data),
  });
}

export function useWinQuery(id: string) {
  return useQuery({
    queryKey: winKeys.detail(id),
    queryFn: () => apiClient.wins.getById(id).then((res) => res.data.data),
    enabled: !!id,
  });
}

export function useCreateWinMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: {
      title: string;
      description?: string;
      occurred_on?: string;
      category?: string;
      tag_ids?: string[];
    }) => apiClient.wins.create(data).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: winKeys.lists() });
      toast.success("Win recorded successfully");
    },
  });
}

export function useUpdateWinMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: {
        title?: string;
        description?: string;
        occurred_on?: string;
        category?: string;
        tag_ids?: string[];
      };
    }) => apiClient.wins.update(id, data).then((res) => res.data.data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: winKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: winKeys.detail(variables.id),
      });
      toast.success("Win updated successfully");
    },
  });
}

export function useDeleteWinMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.wins.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: winKeys.lists() });
      toast.success("Win deleted");
    },
  });
}
