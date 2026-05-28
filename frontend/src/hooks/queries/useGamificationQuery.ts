import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";

export const gamificationKeys = {
  all: ["gamification"] as const,
  progress: () => [...gamificationKeys.all, "progress"] as const,
  badges: () => [...gamificationKeys.all, "badges"] as const,
  heatmap: (days?: number) => [...gamificationKeys.all, "heatmap", days] as const,
  streak: () => [...gamificationKeys.all, "streak"] as const,
};

export function useGamificationProgress() {
  return useQuery({
    queryKey: gamificationKeys.progress(),
    queryFn: () =>
      apiClient.gamification.progress().then((res) => res.data.data),
  });
}

export function useGamificationBadges() {
  return useQuery({
    queryKey: gamificationKeys.badges(),
    queryFn: () =>
      apiClient.gamification.badges().then((res) => res.data.data),
  });
}

export function useGamificationHeatmap(days?: number) {
  return useQuery({
    queryKey: gamificationKeys.heatmap(days),
    queryFn: () =>
      apiClient.gamification.heatmap(days).then((res) => res.data.data),
  });
}

export function useGamificationStreak() {
  return useQuery({
    queryKey: gamificationKeys.streak(),
    queryFn: () =>
      apiClient.gamification.streak().then((res) => res.data.data),
  });
}
