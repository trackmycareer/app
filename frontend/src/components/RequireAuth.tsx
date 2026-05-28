import { useEffect, useState } from "react";
import { Navigate, Outlet, useLocation } from "react-router";
import { useAuthStore } from "@/stores/auth";
import { apiClient } from "@/lib/api";
import { PageLoader } from "@/components/PageLoader";

export function RequireAuth() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const isLoading = useAuthStore((s) => s.isLoading);
  const accessToken = useAuthStore((s) => s.accessToken);
  const location = useLocation();
  const [bootstrapping, setBootstrapping] = useState(false);

  useEffect(() => {
    if (isAuthenticated && !accessToken && !bootstrapping) {
      setBootstrapping(true);
      apiClient.auth
        .refresh()
        .then((res) => {
          useAuthStore.getState().setAccessToken(res.data.data.access_token);
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

  return <Outlet />;
}
