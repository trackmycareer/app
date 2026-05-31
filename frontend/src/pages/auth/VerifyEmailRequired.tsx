import { useState, useEffect, useCallback } from "react";
import { Link } from "react-router";
import { toast } from "sonner";
import { useAuthStore } from "@/stores/auth";
import { apiClient } from "@/lib/api";
import { Button } from "@/components/Button";
import { BriefcaseIcon } from "@/components/icons";

export default function VerifyEmailRequired() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const [sending, setSending] = useState(false);
  const [cooldown, setCooldown] = useState(0);

  useEffect(() => {
    if (cooldown <= 0) return;
    const timer = setInterval(() => {
      setCooldown((prev) => prev - 1);
    }, 1000);
    return () => clearInterval(timer);
  }, [cooldown]);

  const handleResend = useCallback(async () => {
    setSending(true);
    try {
      await apiClient.verification.send();
      toast.success("Verification email sent. Please check your inbox.");
      setCooldown(60);
    } catch {
      toast.error("Failed to send verification email. Please try again.");
    } finally {
      setSending(false);
    }
  }, []);

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
        </div>

        {/* Content */}
        <div className="space-y-4 text-center">
          <h2 className="text-lg font-semibold text-[var(--text-primary)]">Verify your email</h2>
          <p className="text-sm text-[var(--text-secondary)]">
            We&apos;ve sent a verification email to{" "}
            <span className="font-medium text-[var(--text-primary)]">{user?.email}</span>. Please
            check your inbox and click the link to verify your account.
          </p>
        </div>

        {/* Resend button */}
        <Button
          className="w-full"
          onClick={handleResend}
          loading={sending}
          disabled={cooldown > 0}
          aria-disabled={cooldown > 0}
          aria-label={
            cooldown > 0
              ? `Resend verification email, available in ${cooldown} seconds`
              : "Resend verification email"
          }
        >
          {cooldown > 0
            ? `Resend in ${cooldown}s`
            : "Resend verification email"}
        </Button>

        {/* Logout link */}
        <p className="text-center text-sm text-[var(--text-secondary)]">
          <Link
            to="/login"
            onClick={(e) => {
              e.preventDefault();
              logout();
            }}
            className="font-medium text-[var(--accent-bright)] hover:text-[var(--accent-text)]"
          >
            Log out
          </Link>
        </p>
      </div>
    </div>
  );
}
