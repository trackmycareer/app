import { useState, useCallback } from "react";
import type { FormEvent } from "react";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { useAuthStore } from "@/stores/auth";
import { apiClient } from "@/lib/api";

export default function Settings() {
  const user = useAuthStore((s) => s.user);
  const updateUser = useAuthStore((s) => s.updateUser);

  /* ---- Profile section state ---- */
  const [name, setName] = useState(user?.name ?? "");
  const [nameError, setNameError] = useState("");

  const profileMutation = useMutation({
    mutationFn: (data: { name: string }) => apiClient.user.update(data),
    onSuccess: (res) => {
      updateUser(res.data.data);
      toast.success("Profile updated");
    },
  });

  const handleProfileSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      const trimmed = name.trim();
      if (!trimmed) {
        setNameError("Name is required");
        return;
      }
      setNameError("");
      profileMutation.mutate({ name: trimmed });
    },
    [name, profileMutation],
  );

  /* ---- Password section state ---- */
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [currentPasswordError, setCurrentPasswordError] = useState("");
  const [newPasswordError, setNewPasswordError] = useState("");
  const [confirmPasswordError, setConfirmPasswordError] = useState("");

  const passwordMutation = useMutation({
    mutationFn: (data: { current_password: string; new_password: string }) =>
      apiClient.user.changePassword(data),
    onSuccess: () => {
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      toast.success("Password changed successfully");
    },
    onError: () => {
      setCurrentPasswordError("Current password is incorrect");
    },
  });

  const handlePasswordSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      let hasError = false;

      if (!currentPassword) {
        setCurrentPasswordError("Current password is required");
        hasError = true;
      } else {
        setCurrentPasswordError("");
      }

      if (!newPassword) {
        setNewPasswordError("New password is required");
        hasError = true;
      } else if (newPassword.length < 8) {
        setNewPasswordError("Password must be at least 8 characters");
        hasError = true;
      } else {
        setNewPasswordError("");
      }

      if (newPassword !== confirmPassword) {
        setConfirmPasswordError("Passwords do not match");
        hasError = true;
      } else {
        setConfirmPasswordError("");
      }

      if (hasError) return;

      passwordMutation.mutate({
        current_password: currentPassword,
        new_password: newPassword,
      });
    },
    [currentPassword, newPassword, confirmPassword, passwordMutation],
  );

  const isEmailProvider = user?.provider === "email";

  return (
    <>
      <Topbar title="Settings" />
      <div className="mx-auto max-w-2xl space-y-6 p-4 lg:p-6">
        {/* Profile section */}
        <section
          className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
            bg-[var(--bg-surface)] p-5"
          aria-label="Profile settings"
        >
          <h2 className="mb-4 text-base font-semibold text-[var(--text-primary)]">
            Profile
          </h2>
          <form onSubmit={handleProfileSubmit} className="space-y-4">
            <TextInput
              label="Name"
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                if (nameError) setNameError("");
              }}
              error={nameError}
              placeholder="Your display name"
            />

            <div className="flex flex-col gap-1.5">
              <span className="text-sm font-medium text-[var(--text-secondary)]">
                Email
              </span>
              <p
                className="rounded-[var(--radius-md)] border border-[var(--border-default)]
                  bg-[var(--bg-base)] px-3 py-2 text-sm text-[var(--text-tertiary)]"
              >
                {user?.email ?? ""}
              </p>
              <p className="text-xs text-[var(--text-tertiary)]">
                Your email address cannot be changed.
              </p>
            </div>

            <div className="flex justify-end pt-1">
              <Button type="submit" loading={profileMutation.isPending}>
                Save profile
              </Button>
            </div>
          </form>
        </section>

        {/* Password section (only for email provider) */}
        {isEmailProvider && (
          <section
            className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
              bg-[var(--bg-surface)] p-5"
            aria-label="Change password"
          >
            <h2 className="mb-4 text-base font-semibold text-[var(--text-primary)]">
              Change password
            </h2>
            <form onSubmit={handlePasswordSubmit} className="space-y-4">
              <TextInput
                label="Current password"
                type="password"
                value={currentPassword}
                onChange={(e) => {
                  setCurrentPassword(e.target.value);
                  if (currentPasswordError) setCurrentPasswordError("");
                }}
                error={currentPasswordError}
                autoComplete="current-password"
              />

              <TextInput
                label="New password"
                type="password"
                value={newPassword}
                onChange={(e) => {
                  setNewPassword(e.target.value);
                  if (newPasswordError) setNewPasswordError("");
                }}
                error={newPasswordError}
                autoComplete="new-password"
                placeholder="At least 8 characters"
              />

              <TextInput
                label="Confirm new password"
                type="password"
                value={confirmPassword}
                onChange={(e) => {
                  setConfirmPassword(e.target.value);
                  if (confirmPasswordError) setConfirmPasswordError("");
                }}
                error={confirmPasswordError}
                autoComplete="new-password"
              />

              <div className="flex justify-end pt-1">
                <Button type="submit" loading={passwordMutation.isPending}>
                  Change password
                </Button>
              </div>
            </form>
          </section>
        )}

        {/* Account information */}
        <section
          className="rounded-[var(--radius-xl)] border border-[var(--border-subtle)]
            bg-[var(--bg-surface)] p-5"
          aria-label="Account information"
        >
          <h2 className="mb-4 text-base font-semibold text-[var(--text-primary)]">
            Account
          </h2>
          <div className="space-y-3 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-[var(--text-secondary)]">Sign-in provider</span>
              <span
                className="inline-flex items-center rounded-full bg-[var(--bg-elevated)]
                  px-2 py-0.5 text-xs font-medium text-[var(--text-secondary)]"
              >
                {user?.provider ?? "unknown"}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-[var(--text-secondary)]">Member since</span>
              <span className="text-[var(--text-tertiary)]">
                {user?.created_at
                  ? new Date(user.created_at).toLocaleDateString("en-GB", {
                      day: "numeric",
                      month: "long",
                      year: "numeric",
                    })
                  : ""}
              </span>
            </div>
          </div>
        </section>
      </div>
    </>
  );
}
