import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient } from "@/lib/api";
import type { Certification } from "@/types";

export interface CertListParams {
  search?: string;
  status?: string;
  limit?: number;
  offset?: number;
}

export const certKeys = {
  all: ["certifications"] as const,
  lists: () => [...certKeys.all, "list"] as const,
  list: (params: CertListParams) => [...certKeys.lists(), params] as const,
  details: () => [...certKeys.all, "detail"] as const,
  detail: (id: string) => [...certKeys.details(), id] as const,
};

export function useCertsQuery(params: CertListParams = {}) {
  return useQuery({
    queryKey: certKeys.list(params),
    queryFn: () =>
      apiClient.certifications
        .list(params as Record<string, string | number>)
        .then((res) => res.data.data),
  });
}

export function useCertQuery(id: string) {
  return useQuery({
    queryKey: certKeys.detail(id),
    queryFn: () =>
      apiClient.certifications.getById(id).then((res) => res.data.data),
    enabled: !!id,
  });
}

export function useCreateCertMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Partial<Certification>) =>
      apiClient.certifications.create(data).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: certKeys.lists() });
      toast.success("Certification added successfully");
    },
  });
}

export function useUpdateCertMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: Partial<Certification>;
    }) =>
      apiClient.certifications.update(id, data).then((res) => res.data.data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: certKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: certKeys.detail(variables.id),
      });
      toast.success("Certification updated successfully");
    },
  });
}

export function useUpdateCertStatusMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      apiClient.certifications
        .updateStatus(id, status)
        .then((res) => res.data.data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: certKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: certKeys.detail(variables.id),
      });
      toast.success("Status updated");
    },
  });
}

export function useDeleteCertMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.certifications.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: certKeys.lists() });
      toast.success("Certification deleted");
    },
  });
}
