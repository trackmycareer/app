import { useEffect, useRef, useState, useCallback } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router";
import { useAuthStore } from "@/stores/auth";
import { apiClient } from "@/lib/api";
import { SpinnerIcon, AlertTriangleIcon, BriefcaseIcon } from "@/components/icons";
import { Button } from "@/components/Button";
import { MFAChallenge } from "@/components/MFAChallenge";

export default function Callback() {
  const navigate = useNavigate();
  const { provider } = useParams<{ provider: string }>();
  const [searchParams] = useSearchParams();
  const login = useAuthStore((s) => s.login);
  const [error, setError] = useState("");
  const [mfaSession, setMfaSession] = useState<string | null>(null);
  const [mfaMethods, setMfaMethods] = useState<string[]>([]);
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
        const data = tokenRes.data.data as {
          access_token?: string;
          mfa_required?: boolean;
          mfa_session?: string;
          methods?: string[];
        };

        if (data.mfa_required && data.mfa_session) {
          setMfaSession(data.mfa_session);
          setMfaMethods(data.methods ?? []);
          return;
        }

        const accessToken = data.access_token!;
        useAuthStore.getState().setAccessToken(accessToken);
        return apiClient.user.getCurrent().then((userRes) => {
          login(userRes.data.data, accessToken);
          navigate("/wins", { replace: true });
        });
      })
      .catch(() => {
        useAuthStore.getState().logout();
        setError("Failed to complete sign in. Please try again.");
      });
  }, [searchParams, provider, login, navigate]);

  const handleMFASuccess = useCallback(
    async (accessToken: string) => {
      useAuthStore.getState().setAccessToken(accessToken);
      try {
        const userRes = await apiClient.user.getCurrent();
        login(userRes.data.data, accessToken);
        navigate("/wins", { replace: true });
      } catch {
        setError("Failed to complete sign in. Please try again.");
        setMfaSession(null);
        setMfaMethods([]);
      }
    },
    [login, navigate],
  );

  const handleMFACancel = useCallback(() => {
    setMfaSession(null);
    setMfaMethods([]);
    navigate("/login", { replace: true });
  }, [navigate]);

  if (mfaSession) {
    return (
      <div className="relative flex min-h-screen items-center justify-center bg-[var(--bg-base)] px-4">
        <div
          className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--accent-default)_0%,_transparent_70%)] opacity-5"
          aria-hidden="true"
        />
        <div
          className="animate-scale-in relative w-full max-w-sm space-y-6 rounded-[var(--radius-xl)]
            border border-[var(--border-default)] bg-[var(--bg-surface)] p-8"
        >
          <div className="flex flex-col items-center gap-3">
            <BriefcaseIcon className="text-[var(--accent-default)]" width={32} height={32} />
            <h1 className="text-xl font-semibold text-[var(--text-primary)]">
              trackmy<span className="text-[var(--accent-default)]">.</span>career
            </h1>
          </div>
          <MFAChallenge
            mfaSession={mfaSession}
            methods={mfaMethods}
            onSuccess={handleMFASuccess}
            onCancel={handleMFACancel}
          />
        </div>
      </div>
    );
  }

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
