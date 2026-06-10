import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient } from "@/lib/api";
import type { CustomDomain } from "@/types";

export const customDomainKeys = {
  all: ["customDomain"] as const,
  detail: () => [...customDomainKeys.all, "detail"] as const,
};

export function useCustomDomain() {
  return useQuery({
    queryKey: customDomainKeys.detail(),
    queryFn: async (): Promise<CustomDomain | null> => {
      try {
        const res = await apiClient.customDomain.get();
        return res.data.data;
      } catch (err: unknown) {
        const axiosError = err as { response?: { status?: number } };
        if (axiosError.response?.status === 404) {
          return null;
        }
        throw err;
      }
    },
  });
}

export function useCreateCustomDomain() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (domain: string) =>
      apiClient.customDomain.create(domain).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: customDomainKeys.all });
      toast.success("Custom domain added");
    },
  });
}

export function useVerifyCustomDomain() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => apiClient.customDomain.verify().then((res) => res.data.data),
    onSuccess: (data: CustomDomain) => {
      queryClient.invalidateQueries({ queryKey: customDomainKeys.all });
      if (data.status === "active") {
        toast.success("Domain verified and active");
      } else if (data.status === "failed") {
        toast.error("Domain verification failed. Please check your DNS records.");
      } else {
        toast.info("Domain is still pending verification. Please check your DNS records.");
      }
    },
  });
}

export function useRemoveCustomDomain() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => apiClient.customDomain.remove().then((res) => res.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: customDomainKeys.all });
      toast.success("Custom domain removed");
    },
  });
}

export function useUpdateCustomDomainTheme() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (accent_colour: string) =>
      apiClient.customDomain.updateTheme(accent_colour).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: customDomainKeys.all });
      toast.success("Theme updated");
    },
  });
}
