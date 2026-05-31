import { useState, useCallback } from "react";
import type { FormEvent } from "react";
import { Link, useSearchParams, useNavigate } from "react-router";
import { toast } from "sonner";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { MFAChallenge } from "@/components/MFAChallenge";
import { BriefcaseIcon, AlertTriangleIcon } from "@/components/icons";
import { useResetPasswordMutation } from "@/hooks/queries/usePasswordReset";

export default function ResetPassword() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get("token");

  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [mfaSession, setMfaSession] = useState<string | null>(null);
  const [mfaMethods, setMfaMethods] = useState<string[]>([]);

  const resetPassword = useResetPasswordMutation();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");

    if (password !== confirmPassword) {
      setError("Passwords do not match.");
      return;
    }

    if (password.length < 8) {
      setError("Password must be at least 8 characters.");
      return;
    }

    try {
      const res = await resetPassword.mutateAsync({
        token: token!,
        new_password: password,
      });

      const data = res.data as {
        mfa_required?: boolean;
        mfa_session?: string;
        methods?: string[];
      };
      if (data.mfa_required && data.mfa_session) {
        setMfaSession(data.mfa_session);
        setMfaMethods(data.methods ?? []);
        return;
      }

      toast.success("Password reset successfully. Please sign in with your new password.");
      navigate("/login", { replace: true });
    } catch (err: unknown) {
      const axiosError = err as { response?: { status?: number; data?: { message?: string } } };
      if (axiosError.response?.data?.message) {
        setError(axiosError.response.data.message);
      } else {
        setError("Invalid or expired reset token.");
      }
    }
  };

  const handleMFASuccess = useCallback(
    async () => {
      toast.success("Password reset successfully. Please sign in with your new password.");
      navigate("/login", { replace: true });
    },
    [navigate],
  );

  const handleMFACancel = useCallback(() => {
    setMfaSession(null);
    setMfaMethods([]);
  }, []);

  if (!token) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[var(--bg-base)] px-4">
        <div
          className="flex w-full max-w-sm flex-col items-center gap-4 rounded-[var(--radius-xl)]
            border border-[var(--border-default)] bg-[var(--bg-surface)] p-8"
        >
          <AlertTriangleIcon width={40} height={40} className="text-[var(--color-error)]" />
          <h2 className="text-lg font-semibold text-[var(--text-primary)]">Invalid reset link</h2>
          <p className="text-center text-sm text-[var(--text-secondary)]">
            This password reset link is invalid. Please request a new one.
          </p>
          <Link to="/forgot-password">
            <Button variant="secondary">Request new link</Button>
          </Link>
        </div>
      </div>
    );
  }

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
        {/* Logo */}
        <div className="flex flex-col items-center gap-3">
          <BriefcaseIcon className="text-[var(--accent-default)]" width={32} height={32} />
          <h1 className="text-xl font-semibold text-[var(--text-primary)]">
            trackmy<span className="text-[var(--accent-default)]">.</span>career
          </h1>
          <p className="text-sm text-[var(--text-secondary)]">Set your new password</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <p
              className="rounded-[var(--radius-md)] bg-[var(--color-error)]/10 px-3 py-2 text-sm
                text-[var(--color-error)]"
              role="alert"
            >
              {error}
            </p>
          )}
          <TextInput
            label="New password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="At least 8 characters"
            required
            autoFocus
            autoComplete="new-password"
            minLength={8}
          />
          <TextInput
            label="Confirm password"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="Repeat your password"
            required
            autoComplete="new-password"
          />
          <Button type="submit" className="w-full" loading={resetPassword.isPending}>
            Reset password
          </Button>
        </form>

        <p className="text-center text-sm text-[var(--text-secondary)]">
          <Link
            to="/login"
            className="font-medium text-[var(--accent-bright)] hover:text-[var(--accent-text)]"
          >
            Back to sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
