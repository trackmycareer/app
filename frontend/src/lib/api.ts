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
  Application,
  Certification,
  Skill,
  Badge,
  GamificationProgress,
  HeatmapEntry,
  UserStreak,
  ProfileSettings,
  PublicProfile,
  LinkedAccount,
  ImportPreview,
  ImportResult,
  CompanyResult,
  JobTitleResult,
  CertSearchResult,
  SkillSearchResult,
  LocationResult,
  MFAStatus,
  PasskeyInfo,
  CustomDomain,
  AdminUserDetail,
  NotificationListResponse,
  NotificationPreferences,
} from "@/types";

const api = axios.create({
  baseURL: API_URL,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true,
});

// Request interceptor: add auth token
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

// Single-flight refresh: all concurrent 401s share the same refresh promise
// so only one actual HTTP call is made to /auth/refresh.
let refreshPromise: Promise<string> | null = null;

// Response interceptor: handle errors
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
        const isAuthEndpoint = originalRequest?.url?.startsWith("/auth/");
        if (originalRequest && !originalRequest._retry && !isAuthEndpoint) {
          originalRequest._retry = true;
          try {
            if (!refreshPromise) {
              refreshPromise = axios
                .post(`${API_URL}/auth/refresh`, null, { withCredentials: true })
                .then((res) => {
                  const newToken = res.data.data.access_token;
                  useAuthStore.getState().setAccessToken(newToken);
                  return newToken;
                })
                .finally(() => {
                  refreshPromise = null;
                });
            }
            const newToken = await refreshPromise;
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            return api(originalRequest);
          } catch {
            // Clear local state directly to avoid a recursive loop.
            // The server-side token is already invalid, so no need to call /auth/logout.
            useAuthStore.setState({
              user: null,
              accessToken: null,
              isAuthenticated: false,
              isLoading: false,
            });
            toast.error("Session expired. Please sign in again.");
          }
        }
      } else if (status === 403) {
        const code = (error.response?.data as { code?: string })?.code;
        if (code === "EMAIL_NOT_VERIFIED") {
          return Promise.reject(error);
        }
        const mfaSetupRequired = (error.response?.data as { mfa_setup_required?: boolean })
          ?.mfa_setup_required;
        if (mfaSetupRequired) {
          toast.error("Please set up multi-factor authentication to access admin features.");
          window.location.href = "/security";
          return Promise.reject(error);
        }
        toast.error(message || `Request failed (${status})`);
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
  config: {
    get: () => api.get<{ data: { polar_enabled: boolean } }>("/config"),
  },
  support: {
    getCheckoutUrl: (type: "one_time" | "subscription") =>
      api.get<{ data: { checkout_url: string } }>(`/support/checkout?type=${type}`),
    getPortalUrl: () => api.get<{ data: { portal_url: string } }>("/support/portal"),
  },
  auth: {
    register: (data: { email: string; password: string; name: string; accept_terms: boolean }) =>
      api.post<{ data: { access_token: string } }>("/auth/register", data),
    login: (data: { email: string; password: string }) =>
      api.post<{ data: { access_token: string } }>("/auth/login", data),
    refresh: () => api.post<{ data: { access_token: string } }>("/auth/refresh"),
    logout: () => api.post("/auth/logout"),
    providers: () => api.get<{ data: AuthProvidersResponse }>("/auth/providers"),
    initiateOAuth: (provider: string) =>
      api.get<{ data: { auth_url: string } }>(`/auth/${provider}`),
    oauthCallback: (provider: string, data: { code: string; state: string }) =>
      api.post<{ data: { access_token: string } }>(`/auth/${provider}/callback`, data),
    forgotPassword: (email: string) => api.post("/auth/forgot-password", { email }),
    resetPassword: (data: {
      token: string;
      new_password: string;
      mfa_method?: string;
      mfa_code?: string;
    }) => api.post("/auth/reset-password", data),
    verifyMFA: (data: {
      mfa_session: string;
      method: string;
      code?: string;
      assertion?: unknown;
    }) => api.post<{ data: { access_token: string } }>("/auth/mfa/verify", data),
    mfaPasskeyChallenge: (mfaSession: string) =>
      api.post("/auth/mfa/passkey/challenge", { mfa_session: mfaSession }),
    resetPasswordPasskeyChallenge: (token: string) =>
      api.post("/auth/reset-password/passkey-challenge", { token }),
  },
  user: {
    getCurrent: () => api.get<{ data: User }>("/user/me"),
    update: (data: { name?: string; avatar_url?: string | null }) =>
      api.put<{ data: User }>("/user/me", data),
    changePassword: (data: { current_password: string; new_password: string }) =>
      api.put("/user/me/password", data),
    uploadAvatar: (file: File) => {
      const form = new FormData();
      form.append("avatar", file);
      return api.post<{ data: User }>("/user/me/avatar", form, {
        headers: { "Content-Type": "multipart/form-data" },
      });
    },
    deleteAccount: (data: { password?: string; confirmation?: string }) =>
      api.post("/user/me/delete", data),
    updateNewsletter: (data: { opt_in: boolean }) =>
      api.put<{ data: User }>("/user/me/newsletter", data),
    acceptTerms: () => api.post<{ data: User }>("/user/me/accept-terms"),
  },
  mfa: {
    getStatus: () => api.get<{ data: MFAStatus }>("/user/me/mfa/status"),
    setupTOTP: () => api.post<{ data: { uri: string; secret: string } }>("/user/me/mfa/totp/setup"),
    verifyTOTP: (code: string) =>
      api.post<{ data: { backup_codes?: string[] } }>("/user/me/mfa/totp/verify", { code }),
    deleteTOTP: (password: string) => api.delete("/user/me/mfa/totp", { data: { password } }),
    beginPasskeyRegistration: () => api.post("/user/me/mfa/passkeys/register/begin"),
    completePasskeyRegistration: (name: string, credential: unknown) =>
      api.post("/user/me/mfa/passkeys/register/complete", { name, credential }),
    listPasskeys: () => api.get<{ data: PasskeyInfo[] }>("/user/me/mfa/passkeys"),
    renamePasskey: (id: string, name: string) => api.put(`/user/me/mfa/passkeys/${id}`, { name }),
    deletePasskey: (id: string, password: string) =>
      api.delete(`/user/me/mfa/passkeys/${id}`, { data: { password } }),
    backupCodeCount: () =>
      api.get<{ data: { remaining: number } }>("/user/me/mfa/backup-codes/count"),
    regenerateBackupCodes: (password: string) =>
      api.post<{ data: { backup_codes: string[] } }>("/user/me/mfa/backup-codes/regenerate", {
        password,
      }),
    disable: (password: string) => api.post("/user/me/mfa/disable", { password }),
  },
  verification: {
    send: () => api.post("/auth/verify-email/send"),
    verify: (token: string) => api.post("/auth/verify-email/confirm", { token }),
    requestEmailChange: (data: { new_email: string; password: string }) =>
      api.post("/user/me/email/change", data),
    confirmEmailChange: (token: string) => api.post("/user/me/email/confirm", { token }),
  },
  wins: {
    list: (params?: Record<string, string | number>) =>
      api.get<{ data: Win[] }>("/wins", { params }),
    create: (data: Partial<Win> & { tag_ids?: string[] }) => api.post<{ data: Win }>("/wins", data),
    getById: (id: string) => api.get<{ data: Win }>(`/wins/${id}`),
    update: (id: string, data: Partial<Win> & { tag_ids?: string[] }) =>
      api.put<{ data: Win }>(`/wins/${id}`, data),
    delete: (id: string) => api.delete(`/wins/${id}`),
  },
  tags: {
    list: () => api.get<{ data: Tag[] }>("/tags"),
    create: (data: { name: string; colour: string }) => api.post<{ data: Tag }>("/tags", data),
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
  applications: {
    list: () => api.get<{ data: Application[] }>("/applications"),
    create: (data: Partial<Application>) => api.post<{ data: Application }>("/applications", data),
    getById: (id: string) => api.get<{ data: Application }>(`/applications/${id}`),
    update: (id: string, data: Partial<Application>) =>
      api.put<{ data: Application }>(`/applications/${id}`, data),
    move: (id: string, data: { status: string; ordered_ids?: string[] }) =>
      api.patch<{ data: Application }>(`/applications/${id}/move`, data),
    delete: (id: string) => api.delete(`/applications/${id}`),
  },
  companies: {
    search: (q: string) =>
      api.get<{ data: { results: CompanyResult[] } }>("/companies/search", { params: { q } }),
  },
  jobTitles: {
    search: (q: string) =>
      api.get<{ data: { results: JobTitleResult[] } }>("/jobtitles/search", { params: { q } }),
  },
  locations: {
    search: (q: string) =>
      api.get<{ data: { results: LocationResult[] } }>("/locations/search", { params: { q } }),
  },
  certifications: {
    list: (params?: Record<string, string | number>) =>
      api.get<{ data: { certifications: Certification[]; total: number } }>("/certifications", {
        params,
      }),
    create: (data: Partial<Certification>) =>
      api.post<{ data: Certification }>("/certifications", data),
    getById: (id: string) => api.get<{ data: Certification }>(`/certifications/${id}`),
    update: (id: string, data: Partial<Certification>) =>
      api.put<{ data: Certification }>(`/certifications/${id}`, data),
    updateStatus: (id: string, status: string) =>
      api.patch<{ data: Certification }>(`/certifications/${id}/status`, { status }),
    delete: (id: string) => api.delete(`/certifications/${id}`),
    search: (q: string) =>
      api.get<{ data: { results: CertSearchResult[] } }>("/certifications/search", {
        params: { q },
      }),
  },
  skills: {
    list: (params?: Record<string, string>) => api.get<{ data: Skill[] }>("/skills", { params }),
    create: (data: Partial<Skill>) => api.post<{ data: Skill }>("/skills", data),
    getById: (id: string) => api.get<{ data: Skill }>(`/skills/${id}`),
    update: (id: string, data: Partial<Skill>) => api.put<{ data: Skill }>(`/skills/${id}`, data),
    delete: (id: string) => api.delete(`/skills/${id}`),
    addEvidence: (id: string, data: { evidence_type: string; evidence_id: string }) =>
      api.post(`/skills/${id}/evidence`, data),
    removeEvidence: (id: string, evidenceId: string) =>
      api.delete(`/skills/${id}/evidence/${evidenceId}`),
    search: (q: string) =>
      api.get<{ data: { results: SkillSearchResult[] } }>("/skills/search", {
        params: { q },
      }),
  },
  gamification: {
    progress: () => api.get<{ data: GamificationProgress }>("/gamification/progress"),
    badges: () => api.get<{ data: Badge[] }>("/gamification/badges"),
    heatmap: (days?: number) =>
      api.get<{ data: HeatmapEntry[] }>("/gamification/heatmap", {
        params: days ? { days } : undefined,
      }),
    streak: () => api.get<{ data: UserStreak }>("/gamification/streak"),
  },
  profile: {
    getSettings: () => api.get<{ data: ProfileSettings }>("/user/me/profile"),
    updateSettings: (data: Partial<ProfileSettings>) =>
      api.put<{ data: ProfileSettings }>("/user/me/profile", data),
    getPublic: (username: string) => api.get<{ data: PublicProfile }>(`/profiles/${username}`),
    getByDomain: (domain: string) =>
      api.get<{ data: PublicProfile }>(`/profiles/by-domain/${domain}`),
  },
  notifications: {
    list: (params?: Record<string, string | number>) =>
      api.get<{ data: NotificationListResponse }>("/notifications", { params }),
    unreadCount: () => api.get<{ data: { unread: number } }>("/notifications/unread-count"),
    markRead: (id: string) => api.patch(`/notifications/${id}/read`),
    markAllRead: () => api.post<{ data: { marked_read: number } }>("/notifications/read-all"),
    getPreferences: () =>
      api.get<{ data: NotificationPreferences }>("/notifications/preferences"),
    updatePreferences: (data: Partial<NotificationPreferences>) =>
      api.put<{ data: NotificationPreferences }>("/notifications/preferences", data),
  },
  customDomain: {
    get: () => api.get<{ data: CustomDomain }>("/custom-domain"),
    create: (domain: string) => api.post<{ data: CustomDomain }>("/custom-domain", { domain }),
    verify: () => api.post<{ data: CustomDomain }>("/custom-domain/verify"),
    remove: () => api.delete("/custom-domain"),
    updateTheme: (accent_colour: string) =>
      api.put<{ data: CustomDomain }>("/custom-domain/theme", { accent_colour }),
  },
  linkedAccounts: {
    list: () => api.get<{ data: LinkedAccount[] }>("/linked-accounts"),
    initiateLink: (provider: string) =>
      api.get<{ data: { auth_url: string } }>(`/linked-accounts/link/${provider}`),
    oauthCallback: (provider: string, data: { code: string; state: string }) =>
      api.post<{ data: LinkedAccount }>(`/linked-accounts/link/${provider}/callback`, data),
    addWebsite: (data: { url: string }) =>
      api.post<{ data: LinkedAccount }>("/linked-accounts/website", data),
    verifyWebsite: () => api.post<{ data: LinkedAccount }>("/linked-accounts/website/verify"),
    unlink: (provider: string) => api.delete(`/linked-accounts/${provider}`),
  },
  import: {
    preview: (file: File, source?: string, entityType?: string) => {
      const form = new FormData();
      form.append("file", file);
      if (source) form.append("source", source);
      if (entityType) form.append("entity_type", entityType);
      return api.post<{ data: ImportPreview }>("/import/preview", form, {
        headers: { "Content-Type": "multipart/form-data" },
        timeout: 30000,
      });
    },
    confirm: (preview: ImportPreview) =>
      api.post<{ data: ImportResult }>("/import/confirm", preview),
    downloadTemplate: (type: string) =>
      api.get(`/import/templates/${type}`, { responseType: "blob" }),
  },
  export: {
    json: () => api.get("/export/json", { responseType: "blob" }),
    markdown: () => api.get("/export/markdown", { responseType: "blob" }),
  },
  admin: {
    users: {
      list: (params?: Record<string, string | number>) =>
        api.get<{ data: User[]; total: number }>("/admin/users", { params }),
      getById: (id: string) => api.get<{ data: AdminUserDetail }>(`/admin/users/${id}`),
      delete: (id: string) => api.delete(`/admin/users/${id}`),
      toggleAdmin: (id: string, isAdmin: boolean) =>
        api.patch(`/admin/users/${id}/admin`, { is_admin: isAdmin }),
    },
    badges: {
      list: () => api.get<{ data: Badge[] }>("/admin/badges"),
      create: (data: Partial<Badge>) => api.post<{ data: Badge }>("/admin/badges", data),
      update: (id: string, data: Partial<Badge>) =>
        api.put<{ data: Badge }>(`/admin/badges/${id}`, data),
      delete: (id: string) => api.delete(`/admin/badges/${id}`),
    },
    settings: {
      get: () => api.get<{ data: Record<string, string> }>("/admin/settings"),
      update: (data: { settings: Record<string, string> }) => api.put("/admin/settings", data),
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
