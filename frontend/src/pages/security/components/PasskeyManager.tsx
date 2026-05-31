import { useState, useCallback } from "react";
import type { FormEvent } from "react";
import { startRegistration } from "@simplewebauthn/browser";
import { toast } from "sonner";
import { apiClient } from "@/lib/api";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { Modal } from "@/components/Modal";
import {
  FingerprintIcon,
  PlusIcon,
  PencilIcon,
  TrashIcon,
  SpinnerIcon,
} from "@/components/icons";
import {
  usePasskeysQuery,
  useRenamePasskeyMutation,
  useDeletePasskeyMutation,
  mfaKeys,
} from "@/hooks/queries/useMFA";
import { useQueryClient } from "@tanstack/react-query";
import { PasswordConfirmModal } from "./PasswordConfirmModal";
import { BackupCodesDisplay } from "./BackupCodesDisplay";

export function PasskeyManager() {
  const { data: passkeys, isLoading } = usePasskeysQuery();
  const renameMutation = useRenamePasskeyMutation();
  const deleteMutation = useDeletePasskeyMutation();
  const queryClient = useQueryClient();

  const [isRegistering, setIsRegistering] = useState(false);
  const [showNameModal, setShowNameModal] = useState(false);
  const [passkeyName, setPasskeyName] = useState("");
  const [pendingCredential, setPendingCredential] = useState<unknown>(null);
  const [nameError, setNameError] = useState("");

  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState("");

  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [deleteError, setDeleteError] = useState("");

  const [backupCodes, setBackupCodes] = useState<string[] | null>(null);

  const handleAddPasskey = useCallback(async () => {
    setIsRegistering(true);
    try {
      const beginRes = await apiClient.mfa.beginPasskeyRegistration();
      const options = beginRes.data.data;
      const credential = await startRegistration({ optionsJSON: options.publicKey });
      setPendingCredential(credential);
      setPasskeyName("");
      setNameError("");
      setShowNameModal(true);
    } catch (err: unknown) {
      if (err instanceof Error && err.name === "NotAllowedError") {
        toast.error("Passkey registration was cancelled.");
      } else {
        toast.error("Failed to register passkey. Please try again.");
      }
    } finally {
      setIsRegistering(false);
    }
  }, []);

  const handleCompleteRegistration = useCallback(
    async (e: FormEvent) => {
      e.preventDefault();
      if (!passkeyName.trim()) {
        setNameError("Please give this passkey a name");
        return;
      }
      setNameError("");
      try {
        const res = await apiClient.mfa.completePasskeyRegistration(
          passkeyName.trim(),
          pendingCredential,
        );
        queryClient.invalidateQueries({ queryKey: mfaKeys.passkeys() });
        queryClient.invalidateQueries({ queryKey: mfaKeys.status() });
        toast.success("Passkey registered successfully");
        setShowNameModal(false);
        setPendingCredential(null);

        const data = res.data.data as { backup_codes?: string[] };
        if (data.backup_codes && data.backup_codes.length > 0) {
          setBackupCodes(data.backup_codes);
        }
      } catch (err: unknown) {
        const axiosError = err as { response?: { data?: { message?: string } } };
        setNameError(axiosError.response?.data?.message || "Failed to register passkey");
      }
    },
    [passkeyName, pendingCredential, queryClient],
  );

  const handleStartRename = useCallback((id: string, currentName: string) => {
    setEditingId(id);
    setEditName(currentName);
  }, []);

  const handleSaveRename = useCallback(
    (id: string) => {
      if (!editName.trim()) return;
      renameMutation.mutate(
        { id, name: editName.trim() },
        {
          onSuccess: () => setEditingId(null),
        },
      );
    },
    [editName, renameMutation],
  );

  const handleDelete = useCallback(
    (password: string) => {
      if (!deleteTarget) return;
      setDeleteError("");
      deleteMutation.mutate(
        { id: deleteTarget, password },
        {
          onSuccess: () => {
            setDeleteTarget(null);
          },
          onError: (err: unknown) => {
            const axiosError = err as { response?: { data?: { message?: string } } };
            setDeleteError(axiosError.response?.data?.message || "Failed to remove passkey");
          },
        },
      );
    },
    [deleteTarget, deleteMutation],
  );

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8" role="status">
        <SpinnerIcon
          width={20}
          height={20}
          className="animate-spin text-[var(--accent-default)]"
          aria-hidden="true"
        />
        <span className="sr-only">Loading passkeys...</span>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {passkeys && passkeys.length > 0 && (
        <div className="divide-y divide-[var(--border-subtle)]">
          {passkeys.map((passkey) => (
            <div key={passkey.id} className="flex items-center justify-between py-3">
              <div className="flex items-center gap-3 min-w-0">
                <FingerprintIcon
                  width={18}
                  height={18}
                  className="shrink-0 text-[var(--text-tertiary)]"
                />
                <div className="min-w-0">
                  {editingId === passkey.id ? (
                    <div className="flex items-center gap-2">
                      <input
                        type="text"
                        value={editName}
                        onChange={(e) => setEditName(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter") handleSaveRename(passkey.id);
                          if (e.key === "Escape") setEditingId(null);
                        }}
                        className="w-40 rounded-[var(--radius-sm)] border border-[var(--border-default)]
                          bg-[var(--bg-surface)] px-2 py-1 text-sm text-[var(--text-primary)]
                          focus:outline-none focus:ring-2 focus:ring-[var(--accent-default)]"
                        autoFocus
                      />
                      <Button
                        type="button"
                        size="sm"
                        onClick={() => handleSaveRename(passkey.id)}
                        loading={renameMutation.isPending}
                      >
                        Save
                      </Button>
                      <Button
                        type="button"
                        variant="secondary"
                        size="sm"
                        onClick={() => setEditingId(null)}
                      >
                        Cancel
                      </Button>
                    </div>
                  ) : (
                    <>
                      <p className="truncate text-sm font-medium text-[var(--text-primary)]">
                        {passkey.name}
                      </p>
                      <p className="text-xs text-[var(--text-tertiary)]">
                        Added{" "}
                        {new Date(passkey.created_at).toLocaleDateString("en-GB", {
                          day: "numeric",
                          month: "short",
                          year: "numeric",
                        })}
                        {passkey.last_used_at && (
                          <>
                            {" "}
                            &middot; Last used{" "}
                            {new Date(passkey.last_used_at).toLocaleDateString("en-GB", {
                              day: "numeric",
                              month: "short",
                              year: "numeric",
                            })}
                          </>
                        )}
                      </p>
                    </>
                  )}
                </div>
              </div>
              {editingId !== passkey.id && (
                <div className="flex items-center gap-1">
                  <button
                    type="button"
                    onClick={() => handleStartRename(passkey.id, passkey.name)}
                    className="rounded-[var(--radius-sm)] p-1.5 text-[var(--text-tertiary)]
                      transition-colors hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)]"
                    aria-label={`Rename passkey ${passkey.name}`}
                    title="Rename"
                  >
                    <PencilIcon width={14} height={14} />
                  </button>
                  <button
                    type="button"
                    onClick={() => setDeleteTarget(passkey.id)}
                    className="rounded-[var(--radius-sm)] p-1.5 text-[var(--text-tertiary)]
                      transition-colors hover:bg-[var(--bg-hover)] hover:text-[var(--color-error)]"
                    aria-label={`Remove passkey ${passkey.name}`}
                    title="Remove"
                  >
                    <TrashIcon width={14} height={14} />
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {(!passkeys || passkeys.length === 0) && (
        <p className="text-sm text-[var(--text-tertiary)]">No passkeys registered yet.</p>
      )}

      <Button
        type="button"
        variant="secondary"
        size="sm"
        icon={<PlusIcon width={14} height={14} />}
        onClick={handleAddPasskey}
        loading={isRegistering}
      >
        Add passkey
      </Button>

      {/* Name modal for new passkey */}
      <Modal
        open={showNameModal}
        onClose={() => {
          setShowNameModal(false);
          setPendingCredential(null);
        }}
        title="Name your passkey"
      >
        <form onSubmit={handleCompleteRegistration} className="space-y-4">
          <p className="text-sm text-[var(--text-secondary)]">
            Give this passkey a name so you can identify it later.
          </p>
          <TextInput
            label="Passkey name"
            placeholder="e.g. MacBook Pro, iPhone"
            value={passkeyName}
            onChange={(e) => {
              setPasskeyName(e.target.value);
              if (nameError) setNameError("");
            }}
            error={nameError}
            autoFocus
          />
          <div className="flex justify-end gap-3 pt-2">
            <Button
              type="button"
              variant="secondary"
              onClick={() => {
                setShowNameModal(false);
                setPendingCredential(null);
              }}
            >
              Cancel
            </Button>
            <Button type="submit">Save passkey</Button>
          </div>
        </form>
      </Modal>

      {/* Delete confirmation */}
      <PasswordConfirmModal
        open={deleteTarget !== null}
        onClose={() => {
          setDeleteTarget(null);
          setDeleteError("");
        }}
        onConfirm={handleDelete}
        title="Remove passkey"
        description="Enter your password to confirm removal of this passkey."
        loading={deleteMutation.isPending}
        error={deleteError}
      />

      {/* Backup codes modal after first passkey */}
      {backupCodes && (
        <Modal
          open={true}
          onClose={() => setBackupCodes(null)}
          title="Backup codes"
        >
          <BackupCodesDisplay codes={backupCodes} onDone={() => setBackupCodes(null)} />
        </Modal>
      )}
    </div>
  );
}
