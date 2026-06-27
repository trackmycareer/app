import { useState, useEffect } from "react";
import type { FormEvent } from "react";
import { Link, useNavigate } from "react-router";
import { useAuthStore } from "@/stores/auth";
import { apiClient } from "@/lib/api";
import { LEGAL_URLS } from "@/lib/constants";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { GithubIcon, GoogleIcon, BriefcaseIcon } from "@/components/icons";

export default function Register() {
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [newsletterOptIn, setNewsletterOptIn] = useState(false);
  const [acceptTerms, setAcceptTerms] = useState(false);
  const [providers, setProviders] = useState<string[]>([]);
  const [registrationEnabled, setRegistrationEnabled] = useState<boolean | null>(null);

  useEffect(() => {
    if (isAuthenticated) {
      navigate("/wins", { replace: true });
    }
  }, [isAuthenticated, navigate]);

  useEffect(() => {
    apiClient.auth
      .providers()
      .then((res) => {
        setProviders(res.data.data.providers);
        setRegistrationEnabled(res.data.data.registration_enabled);
        if (!res.data.data.registration_enabled) {
          navigate("/login", { replace: true });
        }
      })
      .catch(() => {
        setRegistrationEnabled(true);
      });
  }, [navigate]);

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

    if (!acceptTerms) {
      setError("Please accept the Terms of Service and Privacy Policy to continue.");
      return;
    }

    setLoading(true);

    try {
      const registerRes = await apiClient.auth.register({
        email,
        password,
        name,
        accept_terms: acceptTerms,
      });
      const token = registerRes.data.data.access_token;

      useAuthStore.getState().setAccessToken(token);

      const userRes = await apiClient.user.getCurrent();
      login(userRes.data.data, token);

      if (newsletterOptIn) {
        apiClient.user.updateNewsletter({ opt_in: true }).catch(() => {});
      }

      navigate("/verify-email-required", { replace: true });
    } catch (err: unknown) {
      const axiosError = err as { response?: { status?: number; data?: { message?: string } } };
      if (axiosError.response?.status === 409) {
        setError("An account with this email already exists.");
      } else if (axiosError.response?.data?.message) {
        setError(axiosError.response.data.message);
      } else {
        setError("Something went wrong. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  };

  const handleOAuth = async (provider: string) => {
    try {
      const res = await apiClient.auth.initiateOAuth(provider);
      window.location.href = res.data.data.auth_url;
    } catch {
      // Error toast handled by interceptor
    }
  };

  if (registrationEnabled === null) {
    return null;
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
          <p className="text-sm text-[var(--text-secondary)]">Create your account</p>
        </div>

        {/* OAuth buttons */}
        <div className="space-y-3">
          {providers.includes("github") && (
            <Button
              variant="secondary"
              size="md"
              className="w-full hover:scale-[1.01] active:scale-[0.99]"
              icon={<GithubIcon width={18} height={18} />}
              onClick={() => handleOAuth("github")}
            >
              Continue with GitHub
            </Button>
          )}
          {providers.includes("google") && (
            <Button
              variant="secondary"
              size="md"
              className="w-full hover:scale-[1.01] active:scale-[0.99]"
              icon={<GoogleIcon width={18} height={18} />}
              onClick={() => handleOAuth("google")}
            >
              Continue with Google
            </Button>
          )}
        </div>

        {/* OAuth consent notice */}
        {providers.length > 0 && (
          <p className="text-center text-xs text-[var(--text-tertiary)]">
            By continuing with GitHub or Google you agree to our{" "}
            <a href={LEGAL_URLS.terms} className="underline hover:text-[var(--text-secondary)]">
              Terms
            </a>{" "}
            and{" "}
            <a href={LEGAL_URLS.privacy} className="underline hover:text-[var(--text-secondary)]">
              Privacy Policy
            </a>
            .
          </p>
        )}

        {/* Divider */}
        {providers.length > 0 && (
          <div className="flex items-center gap-3">
            <div className="h-px flex-1 bg-[var(--border-default)]" />
            <span className="text-xs text-[var(--text-tertiary)]">or</span>
            <div className="h-px flex-1 bg-[var(--border-default)]" />
          </div>
        )}

        {/* Registration form */}
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
            label="Name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Your full name"
            required
            autoFocus
            autoComplete="name"
          />
          <TextInput
            label="Email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@example.com"
            required
            autoComplete="email"
          />
          <TextInput
            label="Password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="At least 8 characters"
            required
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
          {/* The checkbox carries its own label; the visible sentence is a separate
              description so the Terms/Privacy links are not nested inside a label. */}
          <div className="flex items-start gap-2 py-1">
            <input
              id="accept-terms"
              type="checkbox"
              required
              checked={acceptTerms}
              onChange={(e) => setAcceptTerms(e.target.checked)}
              aria-describedby="accept-terms-desc"
              className="mt-0.5 h-4 w-4 rounded border-[var(--border-default)]
                text-[var(--accent-default)] focus:ring-[var(--accent-default)]"
            />
            <label htmlFor="accept-terms" className="sr-only">
              I agree to the Terms of Service and Privacy Policy
            </label>
            <p id="accept-terms-desc" className="text-sm text-[var(--text-secondary)]">
              I agree to the{" "}
              <a
                href={LEGAL_URLS.terms}
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium text-[var(--accent-text)] underline hover:text-[var(--accent-bright)]"
              >
                Terms of Service<span className="sr-only"> (opens in a new tab)</span>
              </a>{" "}
              and{" "}
              <a
                href={LEGAL_URLS.privacy}
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium text-[var(--accent-text)] underline hover:text-[var(--accent-bright)]"
              >
                Privacy Policy<span className="sr-only"> (opens in a new tab)</span>
              </a>
              .
            </p>
          </div>
          <label className="flex cursor-pointer items-center gap-2 py-1">
            <input
              type="checkbox"
              checked={newsletterOptIn}
              onChange={(e) => setNewsletterOptIn(e.target.checked)}
              className="h-4 w-4 rounded border-[var(--border-default)] text-[var(--accent-default)]
                focus:ring-[var(--accent-default)]"
            />
            <span className="text-sm text-[var(--text-secondary)]">
              Send me product updates and tips (optional)
            </span>
          </label>
          <Button type="submit" className="w-full" loading={loading}>
            Create account
          </Button>
        </form>

        {/* Login link */}
        <p className="text-center text-sm text-[var(--text-secondary)]">
          Already have an account?{" "}
          <Link
            to="/login"
            className="font-medium text-[var(--accent-bright)] hover:text-[var(--accent-text)]"
          >
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
