import { useState } from "react";
import type { FormEvent } from "react";
import { Link } from "react-router";
import { toast } from "sonner";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { BriefcaseIcon, CheckCircleIcon } from "@/components/icons";
import { useForgotPasswordMutation } from "@/hooks/queries/usePasswordReset";

export default function ForgotPassword() {
  const [email, setEmail] = useState("");
  const [submitted, setSubmitted] = useState(false);

  const forgotPassword = useForgotPasswordMutation();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();

    try {
      await forgotPassword.mutateAsync(email);
      setSubmitted(true);
    } catch {
      toast.error("Something went wrong. Please try again.");
    }
  };

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
          <p className="text-sm text-[var(--text-secondary)]">Reset your password</p>
        </div>

        {submitted ? (
          <div className="flex flex-col items-center gap-4" role="status" aria-live="polite">
            <CheckCircleIcon width={40} height={40} className="text-[var(--color-success)]" />
            <p className="text-center text-sm text-[var(--text-secondary)]">
              Check your email. If an account with that email exists, we&apos;ve sent a password
              reset link.
            </p>
            <Link to="/login">
              <Button variant="secondary">Back to sign in</Button>
            </Link>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            <TextInput
              label="Email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              required
              autoFocus
              autoComplete="email"
            />
            <Button type="submit" className="w-full" loading={forgotPassword.isPending}>
              Send reset link
            </Button>
          </form>
        )}

        {!submitted && (
          <p className="text-center text-sm text-[var(--text-secondary)]">
            <Link
              to="/login"
              className="font-medium text-[var(--accent-bright)] hover:text-[var(--accent-text)]"
            >
              Back to sign in
            </Link>
          </p>
        )}
      </div>
    </div>
  );
}
