import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient } from "@/lib/api";
import type { Skill } from "@/types";

export interface SkillListParams {
  search?: string;
  category?: string;
  limit?: number;
  offset?: number;
}

export const skillKeys = {
  all: ["skills"] as const,
  lists: () => [...skillKeys.all, "list"] as const,
  list: (params: SkillListParams) => [...skillKeys.lists(), params] as const,
  details: () => [...skillKeys.all, "detail"] as const,
  detail: (id: string) => [...skillKeys.details(), id] as const,
};

export function useSkillsQuery(params: SkillListParams = {}) {
  return useQuery({
    queryKey: skillKeys.list(params),
    queryFn: () =>
      apiClient.skills
        .list(params as Record<string, string>)
        .then((res) => res.data.data),
  });
}

export function useSkillQuery(id: string) {
  return useQuery({
    queryKey: skillKeys.detail(id),
    queryFn: () => apiClient.skills.getById(id).then((res) => res.data.data),
    enabled: !!id,
  });
}

export function useCreateSkillMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Partial<Skill>) =>
      apiClient.skills.create(data).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: skillKeys.lists() });
      toast.success("Skill added successfully");
    },
  });
}

export function useUpdateSkillMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: Partial<Skill>;
    }) => apiClient.skills.update(id, data).then((res) => res.data.data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: skillKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: skillKeys.detail(variables.id),
      });
      toast.success("Skill updated successfully");
    },
  });
}

export function useDeleteSkillMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => apiClient.skills.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: skillKeys.lists() });
      toast.success("Skill deleted");
    },
  });
}

export function useAddEvidenceMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      skillId,
      evidenceType,
      evidenceId,
    }: {
      skillId: string;
      evidenceType: string;
      evidenceId: string;
    }) =>
      apiClient.skills.addEvidence(skillId, {
        evidence_type: evidenceType,
        evidence_id: evidenceId,
      }),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: skillKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: skillKeys.detail(variables.skillId),
      });
      toast.success("Evidence linked successfully");
    },
  });
}

export function useRemoveEvidenceMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      skillId,
      evidenceId,
    }: {
      skillId: string;
      evidenceId: string;
    }) => apiClient.skills.removeEvidence(skillId, evidenceId),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: skillKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: skillKeys.detail(variables.skillId),
      });
      toast.success("Evidence removed");
    },
  });
}
