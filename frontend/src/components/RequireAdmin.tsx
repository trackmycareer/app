import { Navigate, Outlet } from "react-router";
import { useAuthStore } from "@/stores/auth";

export function RequireAdmin() {
  const user = useAuthStore((s) => s.user);

  if (!user?.is_admin) {
    return <Navigate to="/" replace />;
  }

  return <Outlet />;
}
