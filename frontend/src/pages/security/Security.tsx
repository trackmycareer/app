import { useState, useCallback } from "react";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import {
  SpinnerIcon,
  ShieldCheckIcon,
  ShieldIcon,
  SmartphoneIcon,
  FingerprintIcon,
  KeyIcon,
  AlertTriangleIcon,
  CheckCircleIcon,
} from "@/components/icons";
import {
  useMFAStatusQuery,
  useDeleteTOTPMutation,
  useBackupCodeCountQuery,
  useRegenerateBackupCodesMutation,
  useDisableMFAMutation,
} from "@/hooks/queries/useMFA";
import { useAuthStore } from "@/stores/auth";
import { TOTPSetupWizard } from "./components/TOTPSetupWizard";
import { PasskeyManager } from "./components/PasskeyManager";
import { BackupCodesDisplay } from "./components/BackupCodesDisplay";
import { PasswordConfirmModal } from "./components/PasswordConfirmModal";

export default function Security() {
  const { data: mfaStatus, isLoading, isError } = useMFAStatusQuery();
  const { data: backupCount } = useBackupCodeCountQuery();
  const user = useAuthStore((s) => s.user);

  const deleteTOTPMutation = useDeleteTOTPMutation();
  const regenerateBackupMutation = useRegenerateBackupCodesMutation();
  const disableMFAMutation = useDisableMFAMutation();

  const [showTOTPSetup, setShowTOTPSetup] = useState(false);
  const [showDeleteTOTP, setShowDeleteTOTP] = useState(false);
  const [deleteTOTPError, setDeleteTOTPError] = useState("");
  const [showRegenerateBackup, setShowRegenerateBackup] = useState(false);
  const [regenerateError, setRegenerateError] = useState("");
  const [regeneratedCodes, setRegeneratedCodes] = useState<string[] | null>(null);
  const [showDisableMFA, setShowDisableMFA] = useState(false);
  const [disableError, setDisableError] = useState("");

  const hasTOTP = mfaStatus?.methods.includes("totp") ?? false;
  const remaining = backupCount?.remaining ?? mfaStatus?.backup_codes_remaining ?? 0;

  const handleDeleteTOTP = useCallback(
    (password: string) => {
      setDeleteTOTPError("");
      deleteTOTPMutation.mutate(password, {
        onSuccess: () => setShowDeleteTOTP(false),
        onError: (err: unknown) => {
          const axiosError = err as { response?: { data?: { message?: string } } };
          setDeleteTOTPError(axiosError.response?.data?.message || "Incorrect password");
        },
      });
    },
    [deleteTOTPMutation],
  );

  const handleRegenerateBackup = useCallback(
    (password: string) => {
      setRegenerateError("");
      regenerateBackupMutation.mutate(password, {
        onSuccess: (data) => {
          setShowRegenerateBackup(false);
          setRegeneratedCodes(data.backup_codes);
        },
        onError: (err: unknown) => {
          const axiosError = err as { response?: { data?: { message?: string } } };
          setRegenerateError(axiosError.response?.data?.message || "Incorrect password");
        },
      });
    },
    [regenerateBackupMutation],
  );

  const handleDisableMFA = useCallback(
    (password: string) => {
      setDisableError("");
      disableMFAMutation.mutate(password, {
        onSuccess: () => setShowDisableMFA(false),
        onError: (err: unknown) => {
          const axiosError = err as { response?: { data?: { message?: string } } };
          setDisableError(axiosError.response?.data?.message || "Incorrect password");
        },
      });
    },
    [disableMFAMutation],
  );

  if (isLoading) {
    return (
      <>
        <Topbar title="Security" />
        <div className="flex items-center justify-center py-20" role="status">
          <SpinnerIcon
            width={28}
            height={28}
            className="animate-spin text-[var(--accent-default)]"
            aria-hidden="true"
          />
          <span className="sr-only">Loading security settings...</span>
        </div>
      </>
    );
  }

  if (isError) {
    return (
      <>
        <Topbar title="Security" />
        <div className="mx-auto max-w-2xl p-4 lg:p-6">
          <div
            className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
              bg-[var(--color-error)]/5 p-4 text-center text-sm text-[var(--color-error)]"
            role="alert"
          >
            Failed to load security settings. Please try again later.
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <Topbar title="Security" />
      <div className="mx-auto max-w-2xl space-y-6 p-4 lg:p-6">
        {/* MFA Status Card */}
        <section
          className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
            bg-[var(--bg-surface)] p-5"
          aria-label="Multi-factor authentication status"
        >
          <div className="flex items-center gap-3">
            {mfaStatus?.enabled ? (
              <div
                className="flex h-10 w-10 items-center justify-center rounded-full
                  bg-[var(--color-success)]/10"
              >
                <ShieldCheckIcon
                  width={20}
                  height={20}
                  className="text-[var(--color-success)]"
                />
              </div>
            ) : (
              <div
                className="flex h-10 w-10 items-center justify-center rounded-full
                  bg-[var(--color-warning)]/10"
              >
                <ShieldIcon width={20} height={20} className="text-[var(--color-warning)]" />
              </div>
            )}
            <div>
              <h2 className="text-base font-semibold text-[var(--text-primary)]">
                Multi-factor authentication
              </h2>
              <p className="text-sm text-[var(--text-secondary)]">
                {mfaStatus?.enabled
                  ? "Your account is protected with MFA."
                  : "Add an extra layer of security to your account."}
              </p>
            </div>
          </div>
          {mfaStatus?.enabled && (
            <div className="mt-3 flex flex-wrap gap-2">
              {mfaStatus.methods.map((method) => (
                <span
                  key={method}
                  className="inline-flex items-center gap-1 rounded-full bg-[var(--color-success)]/10
                    px-2.5 py-0.5 text-xs font-medium text-[var(--color-success)]"
                >
                  <CheckCircleIcon width={12} height={12} />
                  {method === "totp"
                    ? "Authenticator app"
                    : method === "passkey"
                      ? "Passkey"
                      : method}
                </span>
              ))}
            </div>
          )}
        </section>

        {/* Authenticator App Section */}
        <section
          className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
            bg-[var(--bg-surface)] p-5"
          aria-label="Authenticator app"
        >
          <div className="flex items-start justify-between">
            <div className="flex items-start gap-3">
              <SmartphoneIcon
                width={20}
                height={20}
                className="mt-0.5 shrink-0 text-[var(--text-tertiary)]"
              />
              <div>
                <h2 className="text-base font-semibold text-[var(--text-primary)]">
                  Authenticator app
                </h2>
                <p className="mt-1 text-sm text-[var(--text-secondary)]">
                  {hasTOTP
                    ? "Your authenticator app is configured and active."
                    : "Use an authenticator app to generate one-time codes."}
                </p>
              </div>
            </div>
            {hasTOTP ? (
              <Button
                type="button"
                variant="secondary"
                size="sm"
                onClick={() => {
                  setDeleteTOTPError("");
                  setShowDeleteTOTP(true);
                }}
                className="shrink-0"
              >
                Remove
              </Button>
            ) : (
              <Button
                type="button"
                size="sm"
                onClick={() => setShowTOTPSetup(true)}
                className="shrink-0"
              >
                Set up
              </Button>
            )}
          </div>
        </section>

        {/* Passkeys Section */}
        <section
          className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
            bg-[var(--bg-surface)] p-5"
          aria-label="Passkeys"
        >
          <div className="flex items-start gap-3 mb-4">
            <FingerprintIcon
              width={20}
              height={20}
              className="mt-0.5 shrink-0 text-[var(--text-tertiary)]"
            />
            <div>
              <h2 className="text-base font-semibold text-[var(--text-primary)]">Passkeys</h2>
              <p className="mt-1 text-sm text-[var(--text-secondary)]">
                Use biometrics or a security key for passwordless authentication.
              </p>
            </div>
          </div>
          <PasskeyManager />
        </section>

        {/* Backup Codes Section */}
        {mfaStatus?.enabled && (
          <section
            className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
              bg-[var(--bg-surface)] p-5"
            aria-label="Backup codes"
          >
            <div className="flex items-start justify-between">
              <div className="flex items-start gap-3">
                <KeyIcon
                  width={20}
                  height={20}
                  className="mt-0.5 shrink-0 text-[var(--text-tertiary)]"
                />
                <div>
                  <h2 className="text-base font-semibold text-[var(--text-primary)]">
                    Backup codes
                  </h2>
                  <p className="mt-1 text-sm text-[var(--text-secondary)]">
                    Use backup codes to sign in if you lose access to your other methods.
                  </p>
                  <div className="mt-2 flex items-center gap-2">
                    <span
                      className={[
                        "inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium",
                        remaining < 3
                          ? "bg-[var(--color-warning)]/10 text-[var(--color-warning)]"
                          : "bg-[var(--bg-elevated)] text-[var(--text-secondary)]",
                      ].join(" ")}
                    >
                      {remaining < 3 && (
                        <AlertTriangleIcon width={12} height={12} />
                      )}
                      {remaining} remaining
                    </span>
                  </div>
                </div>
              </div>
              <Button
                type="button"
                variant="secondary"
                size="sm"
                onClick={() => {
                  setRegenerateError("");
                  setShowRegenerateBackup(true);
                }}
                className="shrink-0"
              >
                Regenerate
              </Button>
            </div>
          </section>
        )}

        {/* Disable MFA */}
        {mfaStatus?.enabled && !user?.is_admin && (
          <section
            className="rounded-[var(--radius-xl)] border border-red-300 bg-[var(--bg-surface)] p-5"
            aria-label="Disable multi-factor authentication"
          >
            <h2 className="mb-2 text-base font-semibold text-[var(--color-error)]">
              Disable MFA
            </h2>
            <p className="mb-4 text-sm text-[var(--text-secondary)]">
              This will remove all authenticator apps, passkeys, and backup codes from your
              account. You will only need your password to sign in.
            </p>
            <Button
              type="button"
              variant="danger"
              onClick={() => {
                setDisableError("");
                setShowDisableMFA(true);
              }}
            >
              Disable MFA
            </Button>
          </section>
        )}
      </div>

      {/* TOTP Setup Wizard */}
      <TOTPSetupWizard open={showTOTPSetup} onClose={() => setShowTOTPSetup(false)} />

      {/* Delete TOTP confirmation */}
      <PasswordConfirmModal
        open={showDeleteTOTP}
        onClose={() => setShowDeleteTOTP(false)}
        onConfirm={handleDeleteTOTP}
        title="Remove authenticator app"
        description="Enter your password to remove the authenticator app from your account."
        loading={deleteTOTPMutation.isPending}
        error={deleteTOTPError}
      />

      {/* Regenerate backup codes confirmation */}
      <PasswordConfirmModal
        open={showRegenerateBackup}
        onClose={() => setShowRegenerateBackup(false)}
        onConfirm={handleRegenerateBackup}
        title="Regenerate backup codes"
        description="This will invalidate all existing backup codes and generate new ones. Enter your password to continue."
        loading={regenerateBackupMutation.isPending}
        error={regenerateError}
      />

      {/* Show regenerated backup codes */}
      {regeneratedCodes && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div
            className="w-full max-w-lg rounded-[var(--radius-xl)] border
              border-[var(--border-default)] bg-[var(--bg-surface)] shadow-xl"
          >
            <div
              className="flex items-center justify-between border-b border-[var(--border-subtle)]
                px-5 py-4"
            >
              <h2 className="text-lg font-semibold text-[var(--text-primary)]">
                New backup codes
              </h2>
            </div>
            <div className="p-5">
              <BackupCodesDisplay
                codes={regeneratedCodes}
                onDone={() => setRegeneratedCodes(null)}
              />
            </div>
          </div>
        </div>
      )}

      {/* Disable MFA confirmation */}
      <PasswordConfirmModal
        open={showDisableMFA}
        onClose={() => setShowDisableMFA(false)}
        onConfirm={handleDisableMFA}
        title="Disable multi-factor authentication"
        description="This will remove all MFA methods from your account. Are you sure you want to continue?"
        loading={disableMFAMutation.isPending}
        error={disableError}
      />
    </>
  );
}
