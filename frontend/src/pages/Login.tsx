import { useState, useEffect } from "react";
import type { FormEvent } from "react";
import { Link, useNavigate } from "react-router";
import { useAuthStore } from "@/stores/auth";
import { apiClient } from "@/lib/api";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { GithubIcon, GoogleIcon, BriefcaseIcon } from "@/components/icons";

export default function Login() {
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [emailExpanded, setEmailExpanded] = useState(false);
  const [providers, setProviders] = useState<string[]>([]);
  const [registrationEnabled, setRegistrationEnabled] = useState(false);

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
      })
      .catch(() => {
        // Silently fail; email form still works
      });
  }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const loginRes = await apiClient.auth.login({ email, password });
      const token = loginRes.data.data.access_token;

      // Temporarily set token so the user fetch works
      useAuthStore.getState().setAccessToken(token);

      const userRes = await apiClient.user.getCurrent();
      login(userRes.data.data, token);
      navigate("/wins", { replace: true });
    } catch (err: unknown) {
      const axiosError = err as { response?: { status?: number; data?: { message?: string } } };
      if (axiosError.response?.status === 401) {
        setError("Invalid email or password.");
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
          <p className="text-sm text-[var(--text-secondary)]">Sign in to your account</p>
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

        {/* Divider */}
        {providers.length > 0 && (
          <div className="flex items-center gap-3">
            <div className="h-px flex-1 bg-[var(--border-default)]" />
            <span className="text-xs text-[var(--text-tertiary)]">or</span>
            <div className="h-px flex-1 bg-[var(--border-default)]" />
          </div>
        )}

        {/* Collapsible email form */}
        <div>
          {!emailExpanded ? (
            <button
              onClick={() => setEmailExpanded(true)}
              className="w-full rounded-[var(--radius-md)] border border-[var(--border-default)]
                bg-[var(--bg-elevated)] px-4 py-2 text-sm text-[var(--text-secondary)]
                transition-colors hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)]"
            >
              Sign in with email
            </button>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              {error && (
                <p className="rounded-[var(--radius-md)] bg-[var(--color-error)]/10 px-3 py-2 text-sm
                  text-[var(--color-error)]" role="alert">
                  {error}
                </p>
              )}
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
              <TextInput
                label="Password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Your password"
                required
                autoComplete="current-password"
              />
              <Button type="submit" className="w-full" loading={loading}>
                Sign in
              </Button>
            </form>
          )}
        </div>

        {/* Register link */}
        {registrationEnabled && (
          <p className="text-center text-sm text-[var(--text-secondary)]">
            Don&apos;t have an account?{" "}
            <Link
              to="/register"
              className="font-medium text-[var(--accent-bright)] hover:text-[var(--accent-text)]"
            >
              Register
            </Link>
          </p>
        )}
      </div>
    </div>
  );
}
