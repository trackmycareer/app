import { useEffect, useState } from "react";
import { useSearchParams, useNavigate, Link } from "react-router";
import { useAuthStore } from "@/stores/auth";
import { apiClient } from "@/lib/api";
import { SpinnerIcon, CheckCircleIcon, AlertTriangleIcon } from "@/components/icons";
import { Button } from "@/components/Button";

export default function VerifyEmail() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [status, setStatus] = useState<"loading" | "success" | "error">("loading");

  useEffect(() => {
    const token = searchParams.get("token");
    if (!token) {
      setStatus("error");
      return;
    }

    apiClient.verification
      .verify(token)
      .then(async () => {
        setStatus("success");

        // Refresh the JWT so it includes email_verified = true
        try {
          const res = await apiClient.auth.refresh();
          const newToken = res.data.data.access_token;
          useAuthStore.getState().setAccessToken(newToken);

          const userRes = await apiClient.user.getCurrent();
          useAuthStore.getState().updateUser(userRes.data.data);
        } catch {
          // Non-critical: user will get updated on next page load
        }

        setTimeout(() => {
          navigate("/", { replace: true });
        }, 2000);
      })
      .catch(() => {
        setStatus("error");
      });
  }, [searchParams, navigate]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-[var(--bg-base)] px-4">
      <div
        className="flex w-full max-w-sm flex-col items-center gap-4 rounded-[var(--radius-xl)]
          border border-[var(--border-default)] bg-[var(--bg-surface)] p-8"
        role="status"
        aria-live="polite"
      >
        {status === "loading" && (
          <>
            <SpinnerIcon
              width={32}
              height={32}
              className="animate-spin text-[var(--accent-default)]"
            />
            <p className="text-sm text-[var(--text-secondary)]">Verifying your email...</p>
          </>
        )}

        {status === "success" && (
          <>
            <CheckCircleIcon width={40} height={40} className="text-[var(--color-success)]" />
            <h2 className="text-lg font-semibold text-[var(--text-primary)]">Email verified!</h2>
            <p className="text-center text-sm text-[var(--text-secondary)]">
              Your email has been verified successfully. Redirecting you now...
            </p>
          </>
        )}

        {status === "error" && (
          <>
            <AlertTriangleIcon width={40} height={40} className="text-[var(--color-error)]" />
            <h2 className="text-lg font-semibold text-[var(--text-primary)]">
              Verification failed
            </h2>
            <p className="text-center text-sm text-[var(--text-secondary)]">
              This verification link is invalid or has expired.
            </p>
            <Link to="/login">
              <Button variant="secondary">Back to sign in</Button>
            </Link>
          </>
        )}
      </div>
    </div>
  );
}
