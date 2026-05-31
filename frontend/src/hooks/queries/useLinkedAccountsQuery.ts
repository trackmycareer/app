import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient } from "@/lib/api";

export const linkedAccountKeys = {
  all: ["linked-accounts"] as const,
  list: () => [...linkedAccountKeys.all, "list"] as const,
};

export function useLinkedAccounts() {
  return useQuery({
    queryKey: linkedAccountKeys.list(),
    queryFn: () => apiClient.linkedAccounts.list().then((res) => res.data.data),
  });
}

export function useAddWebsiteMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (url: string) =>
      apiClient.linkedAccounts.addWebsite({ url }).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: linkedAccountKeys.list() });
      toast.success("Website added. Add the TXT record and click Verify.");
    },
  });
}

export function useVerifyWebsiteMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => apiClient.linkedAccounts.verifyWebsite().then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: linkedAccountKeys.list() });
      toast.success("Website verified successfully!");
    },
  });
}

export function useUnlinkAccountMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (provider: string) =>
      apiClient.linkedAccounts.unlink(provider).then((res) => res.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: linkedAccountKeys.list() });
      toast.success("Account unlinked");
    },
  });
}
