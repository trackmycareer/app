import axios from "axios";
import type { AxiosError } from "axios";
import { toast } from "sonner";
import { API_URL } from "@/lib/constants";
import { useAuthStore } from "@/stores/auth";
import type {
  User,
  AuthProvidersResponse,
  Win,
  Tag,
  Job,
  Certification,
  Skill,
  Badge,
  GamificationProgress,
  HeatmapEntry,
  UserStreak,
  ProfileSettings,
  PublicProfile,
} from "@/types";

const api = axios.create({
  baseURL: API_URL,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true,
});

// Request interceptor — add auth token
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().accessToken;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    toast.error("Failed to send request");
    return Promise.reject(error);
  },
);

// Response interceptor — handle errors
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const message = (error.response?.data as { message?: string })?.message;

    if (error.code === "ECONNABORTED") {
      toast.error("Request timed out. Please try again.");
    } else if (error.response) {
      const status = error.response.status;

      if (status === 401) {
        const originalRequest = error.config;
        if (originalRequest && !originalRequest._retry) {
          originalRequest._retry = true;
          try {
            const res = await axios.post(`${API_URL}/auth/refresh`, null, {
              withCredentials: true,
            });
            const newToken = res.data.data.access_token;
            useAuthStore.getState().setAccessToken(newToken);
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            return api(originalRequest);
          } catch {
            useAuthStore.getState().logout();
            toast.error("Session expired. Please sign in again.");
          }
        }
      } else if (status !== 422) {
        toast.error(message || `Request failed (${status})`);
      }
    } else if (error.request) {
      toast.error("Network error. Please check your connection.");
    }

    return Promise.reject(error);
  },
);

// Typed API client
export const apiClient = {
  auth: {
    register: (data: { email: string; password: string; name: string }) =>
      api.post<{ data: { access_token: string } }>("/auth/register", data),
    login: (data: { email: string; password: string }) =>
      api.post<{ data: { access_token: string } }>("/auth/login", data),
    refresh: () => api.post<{ data: { access_token: string } }>("/auth/refresh"),
    providers: () => api.get<{ data: AuthProvidersResponse }>("/auth/providers"),
  },
  user: {
    getCurrent: () => api.get<{ data: User }>("/user/me"),
    update: (data: { name?: string; avatar_url?: string | null }) =>
      api.put<{ data: User }>("/user/me", data),
    changePassword: (data: { current_password: string; new_password: string }) =>
      api.put("/user/me/password", data),
  },
  wins: {
    list: (params?: Record<string, string | number>) =>
      api.get<{ data: Win[] }>("/wins", { params }),
    create: (data: Partial<Win> & { tag_ids?: string[] }) =>
      api.post<{ data: Win }>("/wins", data),
    getById: (id: string) => api.get<{ data: Win }>(`/wins/${id}`),
    update: (id: string, data: Partial<Win> & { tag_ids?: string[] }) =>
      api.put<{ data: Win }>(`/wins/${id}`, data),
    delete: (id: string) => api.delete(`/wins/${id}`),
  },
  tags: {
    list: () => api.get<{ data: Tag[] }>("/tags"),
    create: (data: { name: string; colour: string }) =>
      api.post<{ data: Tag }>("/tags", data),
    update: (id: string, data: { name?: string; colour?: string }) =>
      api.put<{ data: Tag }>(`/tags/${id}`, data),
    delete: (id: string) => api.delete(`/tags/${id}`),
  },
  jobs: {
    list: () => api.get<{ data: Job[] }>("/jobs"),
    create: (data: Partial<Job>) => api.post<{ data: Job }>("/jobs", data),
    getById: (id: string) => api.get<{ data: Job }>(`/jobs/${id}`),
    update: (id: string, data: Partial<Job>) => api.put<{ data: Job }>(`/jobs/${id}`, data),
    delete: (id: string) => api.delete(`/jobs/${id}`),
  },
  certifications: {
    list: (params?: Record<string, string | number>) =>
      api.get<{ data: { certifications: Certification[]; total: number } }>(
        "/certifications",
        { params },
      ),
    create: (data: Partial<Certification>) =>
      api.post<{ data: Certification }>("/certifications", data),
    getById: (id: string) => api.get<{ data: Certification }>(`/certifications/${id}`),
    update: (id: string, data: Partial<Certification>) =>
      api.put<{ data: Certification }>(`/certifications/${id}`, data),
    updateStatus: (id: string, status: string) =>
      api.patch<{ data: Certification }>(`/certifications/${id}/status`, { status }),
    delete: (id: string) => api.delete(`/certifications/${id}`),
  },
  skills: {
    list: (params?: Record<string, string>) =>
      api.get<{ data: Skill[] }>("/skills", { params }),
    create: (data: Partial<Skill>) => api.post<{ data: Skill }>("/skills", data),
    getById: (id: string) => api.get<{ data: Skill }>(`/skills/${id}`),
    update: (id: string, data: Partial<Skill>) =>
      api.put<{ data: Skill }>(`/skills/${id}`, data),
    delete: (id: string) => api.delete(`/skills/${id}`),
    addEvidence: (id: string, data: { evidence_type: string; evidence_id: string }) =>
      api.post(`/skills/${id}/evidence`, data),
    removeEvidence: (id: string, evidenceId: string) =>
      api.delete(`/skills/${id}/evidence/${evidenceId}`),
  },
  gamification: {
    progress: () =>
      api.get<{ data: GamificationProgress }>("/gamification/progress"),
    badges: () => api.get<{ data: Badge[] }>("/gamification/badges"),
    heatmap: (days?: number) =>
      api.get<{ data: HeatmapEntry[] }>("/gamification/heatmap", {
        params: days ? { days } : undefined,
      }),
    streak: () => api.get<{ data: UserStreak }>("/gamification/streak"),
  },
  profile: {
    getSettings: () =>
      api.get<{ data: ProfileSettings }>("/user/me/profile"),
    updateSettings: (data: Partial<ProfileSettings>) =>
      api.put<{ data: ProfileSettings }>("/user/me/profile", data),
    getPublic: (username: string) =>
      api.get<{ data: PublicProfile }>(`/profiles/${username}`),
  },
  export: {
    json: () => api.get("/export/json", { responseType: "blob" }),
    markdown: () => api.get("/export/markdown", { responseType: "blob" }),
  },
  admin: {
    users: {
      list: (params?: Record<string, string | number>) =>
        api.get<{ data: User[]; total: number }>("/admin/users", { params }),
      getById: (id: string) => api.get<{ data: User }>(`/admin/users/${id}`),
      delete: (id: string) => api.delete(`/admin/users/${id}`),
      toggleAdmin: (id: string, isAdmin: boolean) =>
        api.patch(`/admin/users/${id}/admin`, { is_admin: isAdmin }),
    },
    badges: {
      list: () => api.get<{ data: Badge[] }>("/admin/badges"),
      create: (data: Partial<Badge>) =>
        api.post<{ data: Badge }>("/admin/badges", data),
      update: (id: string, data: Partial<Badge>) =>
        api.put<{ data: Badge }>(`/admin/badges/${id}`, data),
      delete: (id: string) => api.delete(`/admin/badges/${id}`),
    },
    settings: {
      get: () => api.get<{ data: Record<string, string> }>("/admin/settings"),
      update: (data: { settings: Record<string, string> }) =>
        api.put("/admin/settings", data),
    },
    stats: {
      get: () => api.get<{ data: Record<string, number> }>("/admin/stats"),
    },
  },
};

// Augment AxiosRequestConfig to support _retry flag
declare module "axios" {
  interface InternalAxiosRequestConfig {
    _retry?: boolean;
  }
}

export default api;
