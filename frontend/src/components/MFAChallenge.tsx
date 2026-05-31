import { useState, useRef, useEffect, useCallback } from "react";
import type { FormEvent } from "react";
import { startAuthentication } from "@simplewebauthn/browser";
import { apiClient } from "@/lib/api";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { FingerprintIcon, SmartphoneIcon, KeyIcon } from "@/components/icons";

interface MFAChallengeProps {
  mfaSession: string;
  methods: string[];
  onSuccess: (accessToken: string) => void;
  onCancel?: () => void;
}

type MFAMethod = "totp" | "passkey" | "backup";

export function MFAChallenge({ mfaSession, methods, onSuccess, onCancel }: MFAChallengeProps) {
  const [activeMethod, setActiveMethod] = useState<MFAMethod>(() => {
    if (methods.includes("passkey")) return "passkey";
    if (methods.includes("totp")) return "totp";
    return "backup";
  });
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const codeInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (activeMethod === "totp" || activeMethod === "backup") {
      setCode("");
      setError("");
      requestAnimationFrame(() => {
        codeInputRef.current?.focus();
      });
    }
  }, [activeMethod]);

  const handlePasskey = useCallback(async () => {
    setError("");
    setLoading(true);
    try {
      const challengeRes = await apiClient.auth.mfaPasskeyChallenge(mfaSession);
      const options = challengeRes.data.data;
      const assertion = await startAuthentication({ optionsJSON: options.publicKey });
      const verifyRes = await apiClient.auth.verifyMFA({
        mfa_session: mfaSession,
        method: "passkey",
        assertion,
      });
      onSuccess(verifyRes.data.data.access_token);
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } } };
      if (axiosError.response?.data?.message) {
        setError(axiosError.response.data.message);
      } else if (err instanceof Error && err.name === "NotAllowedError") {
        setError("Passkey authentication was cancelled or timed out.");
      } else {
        setError("Passkey authentication failed. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  }, [mfaSession, onSuccess]);

  const handleCodeSubmit = useCallback(
    async (e: FormEvent) => {
      e.preventDefault();
      if (!code.trim()) {
        setError("Please enter a code.");
        return;
      }
      setError("");
      setLoading(true);
      try {
        const res = await apiClient.auth.verifyMFA({
          mfa_session: mfaSession,
          method: activeMethod,
          code: code.trim(),
        });
        onSuccess(res.data.data.access_token);
      } catch (err: unknown) {
        const axiosError = err as { response?: { data?: { message?: string } } };
        if (axiosError.response?.data?.message) {
          setError(axiosError.response.data.message);
        } else {
          setError("Invalid code. Please try again.");
        }
      } finally {
        setLoading(false);
      }
    },
    [code, mfaSession, activeMethod, onSuccess],
  );

  const availableMethods = methods.filter((m): m is MFAMethod =>
    ["totp", "passkey", "backup"].includes(m),
  );

  const methodLabels: Record<MFAMethod, { label: string; icon: typeof SmartphoneIcon }> = {
    passkey: { label: "Passkey", icon: FingerprintIcon },
    totp: { label: "Authenticator", icon: SmartphoneIcon },
    backup: { label: "Backup code", icon: KeyIcon },
  };

  return (
    <div className="space-y-5">
      <div className="text-center">
        <h2 className="text-lg font-semibold text-[var(--text-primary)]">
          Two-factor authentication
        </h2>
        <p className="mt-1 text-sm text-[var(--text-secondary)]">
          Verify your identity to continue.
        </p>
      </div>

      {/* Method tabs */}
      {availableMethods.length > 1 && (
        <div className="flex gap-1 rounded-[var(--radius-md)] bg-[var(--bg-elevated)] p-1">
          {availableMethods.map((method) => {
            const { label, icon: Icon } = methodLabels[method];
            return (
              <button
                key={method}
                type="button"
                onClick={() => setActiveMethod(method)}
                className={[
                  "flex flex-1 items-center justify-center gap-1.5 rounded-[var(--radius-sm)] px-3 py-2 text-xs font-medium transition-colors",
                  activeMethod === method
                    ? "bg-[var(--bg-surface)] text-[var(--text-primary)] shadow-sm"
                    : "text-[var(--text-tertiary)] hover:text-[var(--text-secondary)]",
                ].join(" ")}
              >
                <Icon width={14} height={14} />
                {label}
              </button>
            );
          })}
        </div>
      )}

      {/* Passkey flow */}
      {activeMethod === "passkey" && (
        <div className="space-y-4">
          <div className="flex flex-col items-center gap-3 py-4">
            <div
              className="flex h-16 w-16 items-center justify-center rounded-full
                bg-[var(--accent-default)]/10"
            >
              <FingerprintIcon
                width={32}
                height={32}
                className="text-[var(--accent-default)]"
              />
            </div>
            <p className="text-center text-sm text-[var(--text-secondary)]">
              Use your passkey to verify your identity.
            </p>
          </div>
          {error && (
            <p
              className="rounded-[var(--radius-md)] bg-[var(--color-error)]/10 px-3 py-2 text-sm text-[var(--color-error)]"
              role="alert"
            >
              {error}
            </p>
          )}
          <Button type="button" className="w-full" onClick={handlePasskey} loading={loading}>
            Use passkey
          </Button>
        </div>
      )}

      {/* TOTP flow */}
      {activeMethod === "totp" && (
        <form onSubmit={handleCodeSubmit} className="space-y-4">
          <p className="text-sm text-[var(--text-secondary)]">
            Enter the 6-digit code from your authenticator app.
          </p>
          {error && (
            <p
              className="rounded-[var(--radius-md)] bg-[var(--color-error)]/10 px-3 py-2 text-sm text-[var(--color-error)]"
              role="alert"
            >
              {error}
            </p>
          )}
          <TextInput
            ref={codeInputRef}
            label="Verification code"
            type="text"
            inputMode="numeric"
            autoComplete="one-time-code"
            placeholder="000000"
            maxLength={6}
            value={code}
            onChange={(e) => {
              const val = e.target.value.replace(/\D/g, "");
              setCode(val);
              if (error) setError("");
            }}
          />
          <Button type="submit" className="w-full" loading={loading}>
            Verify
          </Button>
        </form>
      )}

      {/* Backup code flow */}
      {activeMethod === "backup" && (
        <form onSubmit={handleCodeSubmit} className="space-y-4">
          <p className="text-sm text-[var(--text-secondary)]">
            Enter one of your 8-character backup codes.
          </p>
          {error && (
            <p
              className="rounded-[var(--radius-md)] bg-[var(--color-error)]/10 px-3 py-2 text-sm text-[var(--color-error)]"
              role="alert"
            >
              {error}
            </p>
          )}
          <TextInput
            ref={codeInputRef}
            label="Backup code"
            type="text"
            autoComplete="off"
            placeholder="xxxxxxxx"
            maxLength={8}
            value={code}
            onChange={(e) => {
              setCode(e.target.value);
              if (error) setError("");
            }}
          />
          <Button type="submit" className="w-full" loading={loading}>
            Verify
          </Button>
        </form>
      )}

      {/* Cancel */}
      {onCancel && (
        <div className="text-center">
          <button
            type="button"
            onClick={onCancel}
            className="text-sm text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
          >
            Cancel
          </button>
        </div>
      )}
    </div>
  );
}
