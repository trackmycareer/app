import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router";
import { apiClient } from "@/lib/api";
import { SpinnerIcon, AlertTriangleIcon } from "@/components/icons";
import { Button } from "@/components/Button";
import { toast } from "sonner";

export default function LinkCallback() {
  const navigate = useNavigate();
  const { provider } = useParams<{ provider: string }>();
  const [searchParams] = useSearchParams();
  const [error, setError] = useState("");
  const called = useRef(false);

  useEffect(() => {
    if (called.current) return;
    called.current = true;

    const code = searchParams.get("code");
    const state = searchParams.get("state");
    const oauthError = searchParams.get("error");

    if (oauthError || !code || !state || !provider) {
      setError("Failed to link account. Please try again.");
      return;
    }

    apiClient.linkedAccounts
      .oauthCallback(provider, { code, state })
      .then(() => {
        toast.success(`${provider.charAt(0).toUpperCase() + provider.slice(1)} account linked`);
        navigate("/profile", { replace: true });
      })
      .catch(() => {
        setError("Failed to link account. Please try again.");
      });
  }, [searchParams, provider, navigate]);

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[var(--bg-base)] px-4">
        <div
          className="flex w-full max-w-sm flex-col items-center gap-4 rounded-[var(--radius-xl)]
            border border-[var(--border-default)] bg-[var(--bg-surface)] p-8"
        >
          <AlertTriangleIcon width={40} height={40} className="text-[var(--color-error)]" />
          <p className="text-center text-sm text-[var(--text-secondary)]">{error}</p>
          <Button variant="secondary" onClick={() => navigate("/profile", { replace: true })}>
            Back to profile
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
        <p className="text-sm text-[var(--text-secondary)]">Linking account...</p>
      </div>
    </div>
  );
}
