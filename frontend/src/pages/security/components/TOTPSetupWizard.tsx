import { useState, useRef, useCallback, useEffect } from "react";
import type { FormEvent } from "react";
import { QRCodeSVG } from "qrcode.react";
import { Modal } from "@/components/Modal";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { CopyIcon, CheckIcon, SpinnerIcon } from "@/components/icons";
import { useSetupTOTPMutation, useVerifyTOTPMutation } from "@/hooks/queries/useMFA";
import { BackupCodesDisplay } from "./BackupCodesDisplay";
import { toast } from "sonner";

interface TOTPSetupWizardProps {
  open: boolean;
  onClose: () => void;
}

type Step = "qr" | "verify" | "backup";

export function TOTPSetupWizard({ open, onClose }: TOTPSetupWizardProps) {
  const [step, setStep] = useState<Step>("qr");
  const [uri, setUri] = useState("");
  const [secret, setSecret] = useState("");
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [backupCodes, setBackupCodes] = useState<string[]>([]);
  const [secretCopied, setSecretCopied] = useState(false);
  const codeInputRef = useRef<HTMLInputElement>(null);

  const setupMutation = useSetupTOTPMutation();
  const verifyMutation = useVerifyTOTPMutation();

  useEffect(() => {
    if (open) {
      setStep("qr");
      setUri("");
      setSecret("");
      setCode("");
      setError("");
      setBackupCodes([]);
      setSecretCopied(false);

      setupMutation.mutate(undefined, {
        onSuccess: (data) => {
          setUri(data.uri);
          setSecret(data.secret);
        },
        onError: () => {
          toast.error("Failed to set up authenticator. Please try again.");
          onClose();
        },
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  useEffect(() => {
    if (step === "verify") {
      requestAnimationFrame(() => {
        codeInputRef.current?.focus();
      });
    }
  }, [step]);

  const handleCopySecret = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(secret);
      setSecretCopied(true);
      toast.success("Secret copied to clipboard");
      setTimeout(() => setSecretCopied(false), 2000);
    } catch {
      toast.error("Failed to copy to clipboard");
    }
  }, [secret]);

  const handleVerify = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      if (!code.trim()) {
        setError("Please enter the verification code");
        return;
      }
      setError("");

      verifyMutation.mutate(code.trim(), {
        onSuccess: (data) => {
          if (data.backup_codes && data.backup_codes.length > 0) {
            setBackupCodes(data.backup_codes);
            setStep("backup");
          } else {
            onClose();
          }
        },
        onError: (err: unknown) => {
          const axiosError = err as { response?: { data?: { message?: string } } };
          setError(axiosError.response?.data?.message || "Invalid code. Please try again.");
        },
      });
    },
    [code, verifyMutation, onClose],
  );

  const handleDone = useCallback(() => {
    onClose();
  }, [onClose]);

  return (
    <Modal
      open={open}
      onClose={step === "backup" ? handleDone : onClose}
      title={
        step === "qr"
          ? "Set up authenticator app"
          : step === "verify"
            ? "Verify authenticator"
            : "Backup codes"
      }
    >
      {step === "qr" && (
        <div className="space-y-5">
          {setupMutation.isPending ? (
            <div className="flex items-center justify-center py-12" role="status">
              <SpinnerIcon
                width={28}
                height={28}
                className="animate-spin text-[var(--accent-default)]"
                aria-hidden="true"
              />
              <span className="sr-only">Setting up authenticator...</span>
            </div>
          ) : (
            <>
              <p className="text-sm text-[var(--text-secondary)]">
                Scan this QR code with your authenticator app (such as Google Authenticator, Authy,
                or 1Password).
              </p>

              <div className="flex justify-center rounded-[var(--radius-lg)] bg-white p-4">
                <QRCodeSVG value={uri} size={200} level="M" />
              </div>

              <div className="space-y-2">
                <p className="text-xs font-medium text-[var(--text-tertiary)]">
                  Or enter this secret manually:
                </p>
                <div className="flex items-center gap-2">
                  <code
                    className="flex-1 rounded-[var(--radius-sm)] bg-[var(--bg-elevated)] px-3 py-2
                      font-mono text-xs text-[var(--text-primary)] break-all"
                  >
                    {secret}
                  </code>
                  <Button
                    type="button"
                    variant="secondary"
                    size="sm"
                    onClick={handleCopySecret}
                    icon={
                      secretCopied ? (
                        <CheckIcon width={14} height={14} />
                      ) : (
                        <CopyIcon width={14} height={14} />
                      )
                    }
                  >
                    {secretCopied ? "Copied" : "Copy"}
                  </Button>
                </div>
              </div>

              <div className="flex justify-end pt-2">
                <Button type="button" onClick={() => setStep("verify")}>
                  Continue
                </Button>
              </div>
            </>
          )}
        </div>
      )}

      {step === "verify" && (
        <form onSubmit={handleVerify} className="space-y-4">
          <p className="text-sm text-[var(--text-secondary)]">
            Enter the 6-digit code from your authenticator app to complete the setup.
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

          <div className="flex justify-end gap-3 pt-2">
            <Button type="button" variant="secondary" onClick={() => setStep("qr")}>
              Back
            </Button>
            <Button type="submit" loading={verifyMutation.isPending}>
              Verify and activate
            </Button>
          </div>
        </form>
      )}

      {step === "backup" && <BackupCodesDisplay codes={backupCodes} onDone={handleDone} />}
    </Modal>
  );
}
