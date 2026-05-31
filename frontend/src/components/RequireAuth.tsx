import { useEffect, useState } from "react";
import { Navigate, Outlet, useLocation } from "react-router";
import { useAuthStore } from "@/stores/auth";
import { apiClient } from "@/lib/api";
import { PageLoader } from "@/components/PageLoader";

export function RequireAuth() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const isLoading = useAuthStore((s) => s.isLoading);
  const accessToken = useAuthStore((s) => s.accessToken);
  const user = useAuthStore((s) => s.user);
  const location = useLocation();
  const [bootstrapping, setBootstrapping] = useState(false);

  useEffect(() => {
    if (isAuthenticated && !accessToken && !bootstrapping) {
      setBootstrapping(true);
      apiClient.auth
        .refresh()
        .then(async (res) => {
          useAuthStore.getState().setAccessToken(res.data.data.access_token);
          const userRes = await apiClient.user.getCurrent();
          useAuthStore.getState().updateUser(userRes.data.data);
        })
        .catch(() => {
          useAuthStore.getState().logout();
        })
        .finally(() => {
          setBootstrapping(false);
        });
    }
  }, [isAuthenticated, accessToken, bootstrapping]);

  if (isLoading || bootstrapping) {
    return <PageLoader />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (user && !user.email_verified) {
    return <Navigate to="/verify-email-required" replace />;
  }

  return <Outlet />;
}
