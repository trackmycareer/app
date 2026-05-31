import { useMutation } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";

export function useForgotPasswordMutation() {
  return useMutation({
    mutationFn: (email: string) => apiClient.auth.forgotPassword(email),
  });
}

export function useResetPasswordMutation() {
  return useMutation({
    mutationFn: (data: {
      token: string;
      new_password: string;
      mfa_method?: string;
      mfa_code?: string;
    }) => apiClient.auth.resetPassword(data),
  });
}
