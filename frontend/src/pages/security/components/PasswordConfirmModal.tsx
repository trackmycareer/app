import { useState, useCallback } from "react";
import type { FormEvent } from "react";
import { Modal } from "@/components/Modal";
import { TextInput } from "@/components/TextInput";
import { Button } from "@/components/Button";

interface PasswordConfirmModalProps {
  open: boolean;
  onClose: () => void;
  onConfirm: (password: string) => void;
  title: string;
  description?: string;
  loading?: boolean;
  error?: string;
}

export function PasswordConfirmModal({
  open,
  onClose,
  onConfirm,
  title,
  description,
  loading,
  error: externalError,
}: PasswordConfirmModalProps) {
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      if (!password) {
        setError("Password is required");
        return;
      }
      setError("");
      onConfirm(password);
    },
    [password, onConfirm],
  );

  const handleClose = useCallback(() => {
    setPassword("");
    setError("");
    onClose();
  }, [onClose]);

  const displayError = externalError || error;

  return (
    <Modal open={open} onClose={handleClose} title={title}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {description && (
          <p className="text-sm text-[var(--text-secondary)]">{description}</p>
        )}

        {displayError && (
          <p className="text-sm text-[var(--color-error)]" role="alert">
            {displayError}
          </p>
        )}

        <TextInput
          label="Password"
          type="password"
          value={password}
          onChange={(e) => {
            setPassword(e.target.value);
            if (error) setError("");
          }}
          autoComplete="current-password"
          autoFocus
        />

        <div className="flex justify-end gap-3 pt-2">
          <Button type="button" variant="secondary" onClick={handleClose}>
            Cancel
          </Button>
          <Button type="submit" variant="danger" loading={loading}>
            Confirm
          </Button>
        </div>
      </form>
    </Modal>
  );
}
