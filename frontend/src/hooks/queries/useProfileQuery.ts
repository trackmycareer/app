import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient } from "@/lib/api";
import type { ProfileSettings } from "@/types";

export const profileKeys = {
  all: ["profile"] as const,
  settings: () => [...profileKeys.all, "settings"] as const,
  public: (username: string) => [...profileKeys.all, "public", username] as const,
};

export function useProfileSettings() {
  return useQuery({
    queryKey: profileKeys.settings(),
    queryFn: () =>
      apiClient.profile.getSettings().then((res) => res.data.data),
  });
}

export function useUpdateProfileMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Partial<ProfileSettings>) =>
      apiClient.profile.updateSettings(data).then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: profileKeys.settings() });
      toast.success("Profile updated successfully");
    },
  });
}

export function usePublicProfile(username: string) {
  return useQuery({
    queryKey: profileKeys.public(username),
    queryFn: () =>
      apiClient.profile.getPublic(username).then((res) => res.data.data),
    enabled: !!username,
    retry: false,
  });
}
