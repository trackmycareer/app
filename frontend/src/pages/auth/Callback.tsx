import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router";
import { useAuthStore } from "@/stores/auth";
import { apiClient } from "@/lib/api";
import { SpinnerIcon, AlertTriangleIcon } from "@/components/icons";
import { Button } from "@/components/Button";

export default function Callback() {
  const navigate = useNavigate();
  const { provider } = useParams<{ provider: string }>();
  const [searchParams] = useSearchParams();
  const login = useAuthStore((s) => s.login);
  const [error, setError] = useState("");
  const called = useRef(false);

  useEffect(() => {
    if (called.current) return;
    called.current = true;

    const code = searchParams.get("code");
    const state = searchParams.get("state");
    const oauthError = searchParams.get("error");

    if (oauthError || !code || !state || !provider) {
      setError("Authentication failed. Please try again.");
      return;
    }

    apiClient.auth
      .oauthCallback(provider, { code, state })
      .then((tokenRes) => {
        const accessToken = tokenRes.data.data.access_token;
        useAuthStore.getState().setAccessToken(accessToken);
        return apiClient.user.getCurrent();
      })
      .then((userRes) => {
        const token = useAuthStore.getState().accessToken;
        login(userRes.data.data, token!);
        navigate("/wins", { replace: true });
      })
      .catch(() => {
        useAuthStore.getState().logout();
        setError("Failed to complete sign in. Please try again.");
      });
  }, [searchParams, provider, login, navigate]);

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[var(--bg-base)] px-4">
        <div
          className="flex w-full max-w-sm flex-col items-center gap-4 rounded-[var(--radius-xl)]
            border border-[var(--border-default)] bg-[var(--bg-surface)] p-8"
        >
          <AlertTriangleIcon width={40} height={40} className="text-[var(--color-error)]" />
          <p className="text-center text-sm text-[var(--text-secondary)]">{error}</p>
          <Button variant="secondary" onClick={() => navigate("/login", { replace: true })}>
            Back to sign in
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-[var(--bg-base)]">
      <div className="flex flex-col items-center gap-3">
        <SpinnerIcon
          width={32}
          height={32}
          className="animate-spin text-[var(--accent-default)]"
        />
        <p className="text-sm text-[var(--text-secondary)]">Completing sign in...</p>
      </div>
    </div>
  );
}
