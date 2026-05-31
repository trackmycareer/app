import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient } from "@/lib/api";

export const mfaKeys = {
  all: ["mfa"] as const,
  status: () => [...mfaKeys.all, "status"] as const,
  passkeys: () => [...mfaKeys.all, "passkeys"] as const,
  backupCodeCount: () => [...mfaKeys.all, "backup-code-count"] as const,
};

export function useMFAStatusQuery() {
  return useQuery({
    queryKey: mfaKeys.status(),
    queryFn: () => apiClient.mfa.getStatus().then((res) => res.data.data),
  });
}

export function useSetupTOTPMutation() {
  return useMutation({
    mutationFn: () => apiClient.mfa.setupTOTP().then((res) => res.data.data),
  });
}

export function useVerifyTOTPMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (code: string) => apiClient.mfa.verifyTOTP(code).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mfaKeys.status() });
      toast.success("Authenticator app configured successfully");
    },
  });
}

export function useDeleteTOTPMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (password: string) => apiClient.mfa.deleteTOTP(password),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mfaKeys.status() });
      toast.success("Authenticator app removed");
    },
  });
}

export function usePasskeysQuery() {
  return useQuery({
    queryKey: mfaKeys.passkeys(),
    queryFn: () => apiClient.mfa.listPasskeys().then((res) => res.data.data),
  });
}

export function useRenamePasskeyMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) =>
      apiClient.mfa.renamePasskey(id, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mfaKeys.passkeys() });
      toast.success("Passkey renamed");
    },
  });
}

export function useDeletePasskeyMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, password }: { id: string; password: string }) =>
      apiClient.mfa.deletePasskey(id, password),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mfaKeys.passkeys() });
      queryClient.invalidateQueries({ queryKey: mfaKeys.status() });
      toast.success("Passkey removed");
    },
  });
}

export function useBackupCodeCountQuery() {
  return useQuery({
    queryKey: mfaKeys.backupCodeCount(),
    queryFn: () => apiClient.mfa.backupCodeCount().then((res) => res.data.data),
  });
}

export function useRegenerateBackupCodesMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (password: string) =>
      apiClient.mfa.regenerateBackupCodes(password).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mfaKeys.backupCodeCount() });
      toast.success("Backup codes regenerated");
    },
  });
}

export function useDisableMFAMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (password: string) => apiClient.mfa.disable(password),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mfaKeys.status() });
      queryClient.invalidateQueries({ queryKey: mfaKeys.passkeys() });
      queryClient.invalidateQueries({ queryKey: mfaKeys.backupCodeCount() });
      toast.success("Multi-factor authentication disabled");
    },
  });
}
