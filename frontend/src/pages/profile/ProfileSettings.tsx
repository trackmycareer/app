import { useState, useCallback, useEffect, useRef } from "react";
import type { FormEvent } from "react";
import { useNavigate } from "react-router";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { LocationAutocomplete } from "@/components/LocationAutocomplete";
import { Select } from "@/components/Select";
import { Modal } from "@/components/Modal";
import { ToggleSwitch } from "@/components/ToggleSwitch";
import { ReminderPreferences } from "@/components/ReminderPreferences";
import { SpinnerIcon, LinkIcon, VerifiedBadgeIcon, ExternalLinkIcon } from "@/components/icons";
import { useProfileSettings, useUpdateProfileMutation } from "@/hooks/queries/useProfileQuery";
import {
  useLinkedAccounts,
  useAddWebsiteMutation,
  useVerifyWebsiteMutation,
  useUnlinkAccountMutation,
} from "@/hooks/queries/useLinkedAccountsQuery";
import {
  useCustomDomain,
  useCreateCustomDomain,
  useVerifyCustomDomain,
  useRemoveCustomDomain,
  useUpdateCustomDomainTheme,
} from "@/hooks/queries/useCustomDomain";
import { apiClient } from "@/lib/api";
import { useAuthStore } from "@/stores/auth";
import type { ProfileVisibility } from "@/types";

export default function ProfileSettings() {
  const { data: settings, isLoading, isError } = useProfileSettings();
  const updateMutation = useUpdateProfileMutation();
  const user = useAuthStore((s) => s.user);

  const { data: linkedAccounts } = useLinkedAccounts();
  const addWebsiteMutation = useAddWebsiteMutation();
  const verifyWebsiteMutation = useVerifyWebsiteMutation();
  const unlinkMutation = useUnlinkAccountMutation();
  const [websiteInput, setWebsiteInput] = useState("");

  /* ---- Custom domain state ---- */
  const { data: customDomain, isLoading: isCustomDomainLoading } = useCustomDomain();
  const createDomainMutation = useCreateCustomDomain();
  const verifyDomainMutation = useVerifyCustomDomain();
  const removeDomainMutation = useRemoveCustomDomain();
  const updateThemeMutation = useUpdateCustomDomainTheme();
  const [domainInput, setDomainInput] = useState("");
  const [domainInputError, setDomainInputError] = useState("");
  const [accentColourInput, setAccentColourInput] = useState("#6366f1");
  const [showRemoveDomainModal, setShowRemoveDomainModal] = useState(false);
  const domainSectionRef = useRef<HTMLHeadingElement>(null);

  // Sync accent colour input when custom domain data loads
  useEffect(() => {
    if (customDomain?.accent_colour) {
      setAccentColourInput(customDomain.accent_colour);
    }
  }, [customDomain?.accent_colour]);

  const isSupporter = user?.is_one_time_supporter || user?.is_subscriber || user?.is_admin;

  const validateDomain = useCallback((value: string): string => {
    if (!value.trim()) return "Domain is required";
    // Basic domain format validation
    const domainRegex = /^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;
    if (!domainRegex.test(value.trim())) {
      return "Please enter a valid domain (e.g. portfolio.yourdomain.com)";
    }
    return "";
  }, []);

  const handleCreateDomain = useCallback(() => {
    const trimmed = domainInput.trim();
    const error = validateDomain(trimmed);
    if (error) {
      setDomainInputError(error);
      return;
    }
    setDomainInputError("");
    createDomainMutation.mutate(trimmed, {
      onSuccess: () => {
        setDomainInput("");
        setTimeout(() => domainSectionRef.current?.focus(), 100);
      },
    });
  }, [domainInput, validateDomain, createDomainMutation]);

  const handleRemoveDomain = useCallback(() => {
    removeDomainMutation.mutate(undefined, {
      onSuccess: () => {
        setShowRemoveDomainModal(false);
        setTimeout(() => domainSectionRef.current?.focus(), 100);
      },
    });
  }, [removeDomainMutation]);

  const handleSaveTheme = useCallback(() => {
    updateThemeMutation.mutate(accentColourInput);
  }, [accentColourInput, updateThemeMutation]);

  const [name, setName] = useState("");
  const [nameError, setNameError] = useState("");
  const [username, setUsername] = useState("");
  const [bio, setBio] = useState("");
  const [headline, setHeadline] = useState("");
  const [location, setLocation] = useState("");
  const [openToWork, setOpenToWork] = useState("not_looking");
  const [visibility, setVisibility] = useState<ProfileVisibility>({
    jobs: true,
    certifications: true,
    skills: true,
    wins: true,
    badges: true,
  });
  const [usernameError, setUsernameError] = useState("");

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

  /* ---- Email change state ---- */
  const navigate = useNavigate();
  const [newEmail, setNewEmail] = useState("");
  const [emailChangePassword, setEmailChangePassword] = useState("");
  const [newEmailError, setNewEmailError] = useState("");
  const [emailChangePasswordError, setEmailChangePasswordError] = useState("");

  const emailChangeMutation = useMutation({
    mutationFn: (data: { new_email: string; password: string }) =>
      apiClient.verification.requestEmailChange(data),
    onSuccess: () => {
      setNewEmail("");
      setEmailChangePassword("");
      toast.success("Confirmation email sent to your new address");
    },
    onError: (err: unknown) => {
      const axiosError = err as { response?: { status?: number; data?: { message?: string } } };
      if (axiosError.response?.status === 422) {
        setNewEmailError(axiosError.response.data?.message || "Invalid email address");
      } else if (axiosError.response?.status === 401) {
        setEmailChangePasswordError("Incorrect password");
      } else {
        toast.error("Failed to request email change. Please try again.");
      }
    },
  });

  const handleEmailChangeSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      let hasError = false;

      if (!newEmail.trim()) {
        setNewEmailError("New email is required");
        hasError = true;
      } else {
        setNewEmailError("");
      }

      if (!emailChangePassword) {
        setEmailChangePasswordError("Password is required");
        hasError = true;
      } else {
        setEmailChangePasswordError("");
      }

      if (hasError) return;

      emailChangeMutation.mutate({
        new_email: newEmail.trim(),
        password: emailChangePassword,
      });
    },
    [newEmail, emailChangePassword, emailChangeMutation],
  );

  /* ---- Newsletter state ---- */
  const [newsletterOptIn, setNewsletterOptIn] = useState(user?.newsletter_opt_in ?? false);

  useEffect(() => {
    if (user) {
      setNewsletterOptIn(user.newsletter_opt_in);
    }
  }, [user]);

  const newsletterMutation = useMutation({
    mutationFn: (data: { opt_in: boolean }) => apiClient.user.updateNewsletter(data),
    onSuccess: (res) => {
      useAuthStore.getState().updateUser(res.data.data);
    },
    onError: () => {
      // Revert optimistic update
      setNewsletterOptIn((prev) => !prev);
      toast.error("Failed to update preferences. Please try again.");
    },
  });

  const handleNewsletterToggle = useCallback(
    (val: boolean) => {
      setNewsletterOptIn(val);
      newsletterMutation.mutate({ opt_in: val });
    },
    [newsletterMutation],
  );

  /* ---- Delete account state ---- */
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deletePassword, setDeletePassword] = useState("");
  const [deleteConfirmation, setDeleteConfirmation] = useState("");
  const [deleteError, setDeleteError] = useState("");

  const deleteMutation = useMutation({
    mutationFn: (data: { password?: string; confirmation?: string }) =>
      apiClient.user.deleteAccount(data),
    onSuccess: () => {
      useAuthStore.getState().logout();
      navigate("/login", { replace: true });
    },
    onError: (err: unknown) => {
      const axiosError = err as { response?: { data?: { message?: string } } };
      setDeleteError(axiosError.response?.data?.message || "Failed to delete account");
    },
  });

  const handleDeleteAccount = useCallback(() => {
    setDeleteError("");

    if (isEmailProvider) {
      if (!deletePassword) {
        setDeleteError("Password is required");
        return;
      }
      deleteMutation.mutate({ password: deletePassword });
    } else {
      if (deleteConfirmation !== "DELETE") {
        setDeleteError("Please type DELETE to confirm");
        return;
      }
      deleteMutation.mutate({ confirmation: deleteConfirmation });
    }
  }, [isEmailProvider, deletePassword, deleteConfirmation, deleteMutation]);

  /* ---- Avatar upload state ---- */
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);
  const [avatarError, setAvatarError] = useState("");
  const [avatarUrl, setAvatarUrl] = useState<string | null>(null);

  useEffect(() => {
    if (user?.avatar_url) setAvatarUrl(user.avatar_url);
  }, [user]);

  useEffect(() => {
    if (settings) {
      setName(settings.name ?? user?.name ?? "");
      setUsername(settings.username ?? "");
      setBio(settings.bio ?? "");
      setHeadline(settings.headline ?? "");
      setLocation(settings.location ?? "");
      setOpenToWork(settings.open_to_work || "not_looking");
      if (settings.profile_visibility) {
        setVisibility(settings.profile_visibility);
      }
    }
  }, [settings, user?.name]);

  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 2 * 1024 * 1024) {
      setAvatarError("File must be under 2MB");
      return;
    }
    if (!["image/jpeg", "image/png", "image/webp"].includes(file.type)) {
      setAvatarError("File must be JPG, PNG or WebP");
      return;
    }

    setAvatarError("");
    setIsUploadingAvatar(true);
    try {
      const res = await apiClient.user.uploadAvatar(file);
      setAvatarUrl(res.data.data.avatar_url);
      useAuthStore.getState().updateUser({ avatar_url: res.data.data.avatar_url });
      toast.success("Photo uploaded");
    } catch {
      setAvatarError("Failed to upload photo");
    } finally {
      setIsUploadingAvatar(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  };

  const validateUsername = useCallback((val: string): string => {
    if (!val) return "";
    if (val.length < 3 || val.length > 50) {
      return "Username must be between 3 and 50 characters";
    }
    if (!/^[a-zA-Z0-9-]+$/.test(val)) {
      return "Only letters, numbers, and hyphens are allowed";
    }
    return "";
  }, []);

  const handleSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();

      const trimmedName = name.trim();
      if (!trimmedName) {
        setNameError("Name is required");
        return;
      }
      setNameError("");

      const trimmedUsername = username.trim();
      const error = validateUsername(trimmedUsername);
      if (error) {
        setUsernameError(error);
        return;
      }
      setUsernameError("");

      updateMutation.mutate({
        name: trimmedName,
        username: trimmedUsername || null,
        bio: bio.trim() || null,
        headline: headline.trim() || null,
        location: location.trim() || null,
        open_to_work: openToWork,
        profile_visibility: visibility,
      });
    },
    [
      name,
      username,
      bio,
      headline,
      location,
      openToWork,
      visibility,
      validateUsername,
      updateMutation,
    ],
  );

  const handleVisibilityChange = useCallback((key: keyof ProfileVisibility, val: boolean) => {
    setVisibility((prev) => ({ ...prev, [key]: val }));
  }, []);

  if (isLoading) {
    return (
      <>
        <Topbar title="Profile" />
        <div className="flex items-center justify-center py-20" role="status">
          <SpinnerIcon
            width={28}
            height={28}
            className="animate-spin text-[var(--accent-default)]"
            aria-hidden="true"
          />
          <span className="sr-only">Loading profile settings...</span>
        </div>
      </>
    );
  }

  if (isError) {
    return (
      <>
        <Topbar title="Profile" />
        <div className="mx-auto max-w-2xl p-4 lg:p-6">
          <div
            className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
              bg-[var(--color-error)]/5 p-4 text-center text-sm text-[var(--color-error)]"
            role="alert"
          >
            Failed to load profile settings. Please try again later.
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <Topbar title="Profile" />
      <div className="mx-auto max-w-2xl space-y-6 p-4 lg:p-6">
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Profile information */}
          <section
            className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
              bg-[var(--bg-surface)] p-5"
            aria-label="Profile information"
          >
            <h2 className="mb-4 text-base font-semibold text-[var(--text-primary)]">
              Public profile
            </h2>
            <div className="space-y-4">
              {/* Avatar upload */}
              <div className="flex items-center gap-4">
                <div className="relative h-20 w-20 shrink-0">
                  {avatarUrl ? (
                    <img
                      src={avatarUrl}
                      alt="Your avatar"
                      className="h-20 w-20 rounded-full object-cover"
                    />
                  ) : (
                    <div className="flex h-20 w-20 items-center justify-center rounded-full bg-[var(--accent-default)]/10 text-2xl font-semibold text-[var(--accent-default)]">
                      {user?.name?.charAt(0)?.toUpperCase() || "?"}
                    </div>
                  )}
                </div>
                <div className="flex flex-col gap-1">
                  <Button
                    type="button"
                    variant="secondary"
                    size="sm"
                    onClick={() => fileInputRef.current?.click()}
                    loading={isUploadingAvatar}
                  >
                    Upload photo
                  </Button>
                  <p className="text-xs text-[var(--text-tertiary)]">JPG, PNG or WebP. Max 2MB.</p>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/jpeg,image/png,image/webp"
                    className="hidden"
                    onChange={handleAvatarUpload}
                  />
                  {avatarError && (
                    <p className="text-xs text-[var(--color-error)]">{avatarError}</p>
                  )}
                </div>
              </div>

              {/* Name */}
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

              {/* Username */}
              <div>
                <TextInput
                  label="Username"
                  value={username}
                  onChange={(e) => {
                    setUsername(e.target.value);
                    if (usernameError) setUsernameError("");
                  }}
                  error={usernameError}
                  placeholder="my-username"
                />
                <p className="mt-1 text-xs text-[var(--text-tertiary)]">
                  3 to 50 characters. Letters, numbers, and hyphens only.
                </p>
              </div>

              {/* Headline */}
              <TextInput
                label="Headline"
                placeholder="e.g. Senior Software Engineer"
                value={headline}
                onChange={(e) => setHeadline(e.target.value)}
                maxLength={255}
              />

              {/* Bio */}
              <div className="flex flex-col gap-1.5">
                <label
                  htmlFor="profile-bio"
                  className="text-sm font-medium text-[var(--text-secondary)]"
                >
                  Bio
                </label>
                <textarea
                  id="profile-bio"
                  value={bio}
                  onChange={(e) => setBio(e.target.value)}
                  rows={3}
                  maxLength={500}
                  placeholder="A short description about yourself..."
                  className="w-full rounded-[var(--radius-md)] border border-[var(--border-default)]
                    bg-[var(--bg-surface)] px-3 py-2 text-sm text-[var(--text-primary)]
                    placeholder:text-[var(--text-tertiary)] transition-colors
                    focus:outline-none focus:ring-2 focus:ring-[var(--accent-default)]
                    focus:ring-offset-1 focus:ring-offset-[var(--bg-base)]"
                />
                <p className="text-right text-xs text-[var(--text-tertiary)]">{bio.length} / 500</p>
              </div>

              {/* Location */}
              <LocationAutocomplete
                label="Location"
                placeholder="e.g. London, UK"
                value={location}
                onChange={(value) => setLocation(value)}
              />

              {/* Open to work */}
              <Select
                label="Work status"
                value={openToWork}
                onChange={(e) => setOpenToWork(e.target.value)}
              >
                <option value="not_looking">Not looking</option>
                <option value="open">Open to opportunities</option>
                <option value="actively_looking">Actively looking</option>
              </Select>

              {/* Profile URL preview */}
              {username.trim() && !usernameError && (
                <div
                  className="flex items-center gap-2 rounded-[var(--radius-md)]
                    bg-[var(--bg-elevated)] px-3 py-2"
                >
                  <LinkIcon
                    width={14}
                    height={14}
                    className="shrink-0 text-[var(--text-tertiary)]"
                  />
                  <span className="truncate text-xs text-[var(--text-secondary)]">
                    {window.location.origin}/u/{username.trim()}
                  </span>
                </div>
              )}
            </div>
          </section>

          {/* Visibility toggles */}
          <section
            className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
              bg-[var(--bg-surface)] p-5"
            aria-label="Profile visibility"
          >
            <h2 className="mb-2 text-base font-semibold text-[var(--text-primary)]">Visibility</h2>
            <p className="mb-4 text-xs text-[var(--text-tertiary)]">
              Choose which sections are visible on your public profile.
            </p>
            <div className="divide-y divide-[var(--border-subtle)]">
              <ToggleSwitch
                label="Jobs"
                checked={visibility.jobs}
                onChange={(v) => handleVisibilityChange("jobs", v)}
              />
              <ToggleSwitch
                label="Certifications"
                checked={visibility.certifications}
                onChange={(v) => handleVisibilityChange("certifications", v)}
              />
              <ToggleSwitch
                label="Skills"
                checked={visibility.skills}
                onChange={(v) => handleVisibilityChange("skills", v)}
              />
              <ToggleSwitch
                label="Wins"
                checked={visibility.wins}
                onChange={(v) => handleVisibilityChange("wins", v)}
              />
              <ToggleSwitch
                label="Badges"
                checked={visibility.badges}
                onChange={(v) => handleVisibilityChange("badges", v)}
              />
            </div>
          </section>

          <div className="flex justify-end">
            <Button type="submit" loading={updateMutation.isPending}>
              Save profile
            </Button>
          </div>
        </form>

        {/* Custom domain */}
        <section
          className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
              bg-[var(--bg-surface)] p-5"
          aria-label="Custom domain"
        >
          <h2
            ref={domainSectionRef}
            tabIndex={-1}
            className="mb-1 text-base font-semibold text-[var(--text-primary)] outline-none"
          >
            Custom domain
          </h2>
          <p className="mb-4 text-xs text-[var(--text-tertiary)]">
            Serve your public profile from your own domain.
          </p>

          {!isSupporter ? (
            /* State 1: Not a supporter */
            <div className="flex items-center gap-3 rounded-[var(--radius-md)] border border-[var(--border-subtle)] bg-[var(--bg-elevated)] p-4">
              <svg
                className="h-5 w-5 shrink-0 text-[var(--text-tertiary)]"
                fill="none"
                viewBox="0 0 24 24"
                strokeWidth={1.5}
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M16.5 10.5V6.75a4.5 4.5 0 1 0-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H6.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z"
                />
              </svg>
              <div className="flex-1">
                <p className="text-sm text-[var(--text-secondary)]">
                  Custom domains are available to supporters. Support trackmy.career to unlock this
                  feature.
                </p>
                <a
                  href="/support"
                  className="mt-2 inline-flex items-center text-sm font-medium text-[var(--accent-default)] transition-colors hover:text-[var(--accent-bright)]"
                >
                  Become a supporter
                </a>
              </div>
            </div>
          ) : isCustomDomainLoading ? (
            <div className="flex items-center justify-center py-6" role="status">
              <SpinnerIcon width={20} height={20} aria-hidden="true" />
              <span className="sr-only">Loading custom domain settings...</span>
            </div>
          ) : !customDomain ? (
            /* State 2: Supporter, no domain */
            <form
              onSubmit={(e) => {
                e.preventDefault();
                handleCreateDomain();
              }}
              className="space-y-3"
            >
              <div className="flex gap-2">
                <TextInput
                  label="Custom domain"
                  placeholder="portfolio.yourdomain.com"
                  value={domainInput}
                  onChange={(e) => {
                    setDomainInput(e.target.value);
                    if (domainInputError) setDomainInputError("");
                  }}
                  error={domainInputError}
                />
              </div>
              <div className="flex justify-end">
                <Button
                  type="submit"
                  size="sm"
                  loading={createDomainMutation.isPending}
                >
                  Connect domain
                </Button>
              </div>
            </form>
          ) : customDomain.status === "pending" || customDomain.status === "failed" ? (
            /* State 3: Domain pending (or failed) */
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-[var(--text-primary)]">
                    {customDomain.domain}
                  </span>
                  <span
                    role="status"
                    className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                      customDomain.status === "failed"
                        ? "bg-[var(--color-error)]/10 text-[var(--color-error)]"
                        : "bg-amber-500/10 text-[var(--color-warning)]"
                    }`}
                  >
                    <span className="sr-only">Domain status: </span>
                    {customDomain.status === "failed" ? "Failed" : "Pending"}
                  </span>
                </div>
              </div>

              <div
                className="rounded-[var(--radius-md)] border border-[var(--border-subtle)]
                    bg-[var(--bg-elevated)] p-3"
                role="region"
                aria-label="DNS configuration instructions"
              >
                <p className="mb-2 text-xs font-semibold text-[var(--text-primary)]">
                  Point your domain to the CNAME target below by adding a CNAME record at your DNS
                  provider:
                </p>
                <div className="flex items-center gap-2">
                  <code
                    className="flex-1 rounded-[var(--radius-sm)] bg-[var(--bg-base)] px-3 py-2
                        text-xs font-mono text-[var(--text-secondary)] break-all"
                  >
                    {customDomain.cname_target}
                  </code>
                  <Button
                    type="button"
                    variant="secondary"
                    size="sm"
                    aria-label="Copy CNAME target to clipboard"
                    onClick={() => {
                      navigator.clipboard.writeText(customDomain.cname_target);
                      toast.success("Copied to clipboard");
                    }}
                  >
                    Copy
                  </Button>
                </div>
                <p className="mt-2 text-xs text-[var(--text-tertiary)]">
                  Record type: CNAME, Host: your subdomain, Value: {customDomain.cname_target}
                </p>
              </div>

              <div className="flex items-center gap-2 text-xs text-[var(--text-tertiary)]">
                <span>SSL:</span>
                <span
                  role="status"
                  className={`inline-flex items-center rounded-full px-2 py-0.5 font-medium ${
                    customDomain.ssl_status === "active"
                      ? "bg-green-500/10 text-[var(--color-success)]"
                      : "bg-amber-500/10 text-[var(--color-warning)]"
                  }`}
                >
                  <span className="sr-only">SSL status: </span>
                  {customDomain.ssl_status}
                </span>
              </div>

              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  size="sm"
                  onClick={() => verifyDomainMutation.mutate()}
                  loading={verifyDomainMutation.isPending}
                >
                  Check status
                </Button>
                <Button
                  type="button"
                  variant="secondary"
                  size="sm"
                  className="!border-red-200 !text-red-500"
                  onClick={() => setShowRemoveDomainModal(true)}
                >
                  Remove
                </Button>
              </div>
            </div>
          ) : (
            /* State 4: Domain active */
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-[var(--text-primary)]">
                    {customDomain.domain}
                  </span>
                  <span role="status" className="inline-flex items-center rounded-full bg-green-500/10 px-2 py-0.5 text-xs font-medium text-[var(--color-success)]">
                    <span className="sr-only">Domain status: </span>
                    Active
                  </span>
                </div>
                <a
                  href={`https://${customDomain.domain}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1 text-sm text-[var(--accent-default)] transition-colors hover:text-[var(--accent-bright)]"
                  aria-label={`Visit ${customDomain.domain}`}
                >
                  Visit
                  <ExternalLinkIcon width={14} height={14} aria-hidden="true" />
                </a>
              </div>

              {/* Accent colour picker */}
              <div className="space-y-2">
                <label
                  htmlFor="accent-colour-picker"
                  className="text-sm font-medium text-[var(--text-secondary)]"
                >
                  Accent colour
                </label>
                <div className="flex items-center gap-3">
                  <input
                    id="accent-colour-picker"
                    type="color"
                    value={accentColourInput}
                    onChange={(e) => setAccentColourInput(e.target.value)}
                    className="h-10 w-10 cursor-pointer rounded-[var(--radius-md)] border border-[var(--border-default)] bg-transparent p-0.5"
                    aria-label="Choose accent colour"
                  />
                  <TextInput
                    label="Hex value"
                    value={accentColourInput}
                    onChange={(e) => {
                      const val = e.target.value;
                      setAccentColourInput(val);
                    }}
                    placeholder="#6366f1"
                    className="!w-28 font-mono"
                  />
                  <Button
                    type="button"
                    size="sm"
                    onClick={handleSaveTheme}
                    loading={updateThemeMutation.isPending}
                    disabled={accentColourInput === customDomain.accent_colour}
                  >
                    Save
                  </Button>
                </div>
                <div
                  className="h-2 w-full rounded-full"
                  style={{ backgroundColor: accentColourInput }}
                  aria-hidden="true"
                />
              </div>

              <div>
                <Button
                  type="button"
                  variant="secondary"
                  size="sm"
                  className="!border-red-200 !text-red-500"
                  onClick={() => setShowRemoveDomainModal(true)}
                >
                  Remove domain
                </Button>
              </div>
            </div>
          )}
        </section>

        {/* Remove domain confirmation modal */}
        <Modal
          open={showRemoveDomainModal}
          onClose={() => setShowRemoveDomainModal(false)}
          title="Remove custom domain"
        >
          <div className="space-y-4">
            <p className="text-sm text-[var(--text-secondary)]">
              Are you sure you want to remove <strong>{customDomain?.domain}</strong>? Your profile
              will no longer be accessible via this domain.
            </p>
            <div className="flex justify-end gap-3 pt-2">
              <Button variant="secondary" onClick={() => setShowRemoveDomainModal(false)}>
                Cancel
              </Button>
              <Button
                variant="danger"
                onClick={handleRemoveDomain}
                loading={removeDomainMutation.isPending}
              >
                Remove domain
              </Button>
            </div>
          </div>
        </Modal>

        {/* Connected accounts */}
        <section
          className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
              bg-[var(--bg-surface)] p-5"
          aria-label="Connected accounts"
        >
          <h2 className="mb-1 text-base font-semibold text-[var(--text-primary)]">
            Connected accounts
          </h2>
          <p className="mb-4 text-xs text-[var(--text-tertiary)]">
            Link your accounts to prove ownership. Verified accounts show a badge on your public
            profile.
          </p>
          <div className="divide-y divide-[var(--border-subtle)]">
            {/* LinkedIn */}
            {(() => {
              const linkedin = linkedAccounts?.find((a) => a.provider === "linkedin");
              return (
                <div className="flex items-center justify-between py-3">
                  <div className="flex items-center gap-3">
                    <svg
                      className="h-5 w-5 text-[#0a66c2]"
                      fill="currentColor"
                      viewBox="0 0 24 24"
                      aria-hidden="true"
                    >
                      <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 0 1-2.063-2.065 2.064 2.064 0 1 1 2.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z" />
                    </svg>
                    <div>
                      <div className="flex items-center gap-1.5 text-sm font-medium text-[var(--text-primary)]">
                        LinkedIn
                        {linkedin?.verified && <VerifiedBadgeIcon />}
                      </div>
                      {linkedin ? (
                        <p className="text-xs text-[var(--color-success)]">
                          {linkedin.profile_url}
                        </p>
                      ) : (
                        <p className="text-xs text-[var(--text-tertiary)]">Not connected</p>
                      )}
                    </div>
                  </div>
                  {linkedin ? (
                    <Button
                      type="button"
                      variant="secondary"
                      size="sm"
                      aria-label="Unlink LinkedIn account"
                      onClick={() => {
                        if (confirm("Unlink your LinkedIn account?")) {
                          unlinkMutation.mutate("linkedin");
                        }
                      }}
                      loading={unlinkMutation.isPending}
                      className="!border-red-200 !text-red-500"
                    >
                      Unlink
                    </Button>
                  ) : (
                    <Button
                      type="button"
                      size="sm"
                      aria-label="Link LinkedIn account"
                      onClick={() => {
                        apiClient.linkedAccounts.initiateLink("linkedin").then((res) => {
                          window.location.href = res.data.data.auth_url;
                        });
                      }}
                    >
                      Link account
                    </Button>
                  )}
                </div>
              );
            })()}

            {/* GitHub */}
            {(() => {
              const github = linkedAccounts?.find((a) => a.provider === "github");
              return (
                <div className="flex items-center justify-between py-3">
                  <div className="flex items-center gap-3">
                    <svg
                      className="h-5 w-5 text-[var(--text-primary)]"
                      fill="currentColor"
                      viewBox="0 0 24 24"
                      aria-hidden="true"
                    >
                      <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
                    </svg>
                    <div>
                      <div className="flex items-center gap-1.5 text-sm font-medium text-[var(--text-primary)]">
                        GitHub
                        {github?.verified && <VerifiedBadgeIcon />}
                      </div>
                      {github ? (
                        <p className="text-xs text-[var(--color-success)]">{github.profile_url}</p>
                      ) : (
                        <p className="text-xs text-[var(--text-tertiary)]">Not connected</p>
                      )}
                    </div>
                  </div>
                  {github ? (
                    <Button
                      type="button"
                      variant="secondary"
                      size="sm"
                      aria-label="Unlink GitHub account"
                      onClick={() => {
                        if (confirm("Unlink your GitHub account?")) {
                          unlinkMutation.mutate("github");
                        }
                      }}
                      loading={unlinkMutation.isPending}
                      className="!border-red-200 !text-red-500"
                    >
                      Unlink
                    </Button>
                  ) : (
                    <Button
                      type="button"
                      size="sm"
                      aria-label="Link GitHub account"
                      onClick={() => {
                        apiClient.linkedAccounts.initiateLink("github").then((res) => {
                          window.location.href = res.data.data.auth_url;
                        });
                      }}
                    >
                      Link account
                    </Button>
                  )}
                </div>
              );
            })()}

            {/* Website */}
            {(() => {
              const website = linkedAccounts?.find((a) => a.provider === "website");
              return (
                <div className="py-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <svg
                        className="h-5 w-5 text-[var(--text-tertiary)]"
                        fill="none"
                        viewBox="0 0 24 24"
                        strokeWidth={1.5}
                        stroke="currentColor"
                        aria-hidden="true"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3"
                        />
                      </svg>
                      <div>
                        <div className="flex items-center gap-1.5 text-sm font-medium text-[var(--text-primary)]">
                          Website
                          {website?.verified && <VerifiedBadgeIcon />}
                        </div>
                        {website ? (
                          <p
                            className={`text-xs ${website.verified ? "text-[var(--color-success)]" : "text-[var(--color-warning)]"}`}
                          >
                            {website.verified ? website.profile_url : "Pending verification"}
                          </p>
                        ) : (
                          <p className="text-xs text-[var(--text-tertiary)]">Not connected</p>
                        )}
                      </div>
                    </div>
                    {website && (
                      <Button
                        type="button"
                        variant="secondary"
                        size="sm"
                        aria-label="Remove website"
                        onClick={() => {
                          if (confirm("Remove your website?")) {
                            unlinkMutation.mutate("website");
                          }
                        }}
                        loading={unlinkMutation.isPending}
                        className="!border-red-200 !text-red-500"
                      >
                        Remove
                      </Button>
                    )}
                  </div>

                  {!website && (
                    <div className="mt-3 flex gap-2">
                      <TextInput
                        label=""
                        aria-label="Website URL"
                        type="url"
                        placeholder="https://yourdomain.com"
                        value={websiteInput}
                        onChange={(e) => setWebsiteInput(e.target.value)}
                      />
                      <Button
                        type="button"
                        size="sm"
                        aria-label="Add website"
                        onClick={() => {
                          if (websiteInput.trim()) {
                            addWebsiteMutation.mutate(websiteInput.trim());
                            setWebsiteInput("");
                          }
                        }}
                        loading={addWebsiteMutation.isPending}
                      >
                        Add
                      </Button>
                    </div>
                  )}

                  {website && !website.verified && website.verify_token && (
                    <div
                      className="mt-3 rounded-[var(--radius-md)] border border-[var(--border-subtle)] bg-[var(--bg-elevated)] p-3"
                      role="region"
                      aria-label="DNS verification instructions"
                    >
                      <p className="mb-2 text-xs font-semibold text-[var(--text-primary)]">
                        Add this TXT record to your DNS:
                      </p>
                      <div className="flex items-center gap-2">
                        <code className="flex-1 rounded-[var(--radius-sm)] bg-[var(--bg-base)] px-3 py-2 text-xs font-mono text-[var(--text-secondary)] break-all">
                          {website.verify_token}
                        </code>
                        <Button
                          type="button"
                          variant="secondary"
                          size="sm"
                          aria-label="Copy DNS verification record to clipboard"
                          onClick={() => {
                            navigator.clipboard.writeText(website.verify_token!);
                            toast.success("Copied to clipboard");
                          }}
                        >
                          Copy
                        </Button>
                      </div>
                      <p className="mt-2 text-xs text-[var(--text-tertiary)]">
                        Record type: TXT, Host: @, TTL: any
                      </p>
                      <div className="mt-3">
                        <Button
                          type="button"
                          size="sm"
                          onClick={() => verifyWebsiteMutation.mutate()}
                          loading={verifyWebsiteMutation.isPending}
                        >
                          Verify now
                        </Button>
                      </div>
                      <p className="mt-2 text-xs text-[var(--text-tertiary)]">
                        DNS changes can take up to 48 hours to propagate. Verified domains are
                        re-checked monthly.
                      </p>
                    </div>
                  )}
                </div>
              );
            })()}
          </div>
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
          <h2 className="mb-4 text-base font-semibold text-[var(--text-primary)]">Account</h2>
          <div className="space-y-3 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-[var(--text-secondary)]">Email</span>
              <span className="text-[var(--text-tertiary)]">{user?.email ?? ""}</span>
            </div>
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

        {/* Email section (email provider only) */}
        {isEmailProvider && (
          <section
            className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
              bg-[var(--bg-surface)] p-5"
            aria-label="Email address"
          >
            <h2 className="mb-4 text-base font-semibold text-[var(--text-primary)]">
              Email address
            </h2>
            <div className="mb-4 flex items-center gap-2 text-sm">
              <span className="text-[var(--text-secondary)]">{user?.email}</span>
              {user?.email_verified ? (
                <span
                  className="inline-flex items-center gap-1 rounded-full bg-green-500/10
                    px-2 py-0.5 text-xs font-medium text-[var(--color-success)]"
                >
                  <VerifiedBadgeIcon width={12} height={12} />
                  Verified
                </span>
              ) : (
                <span
                  className="inline-flex items-center rounded-full bg-amber-500/10
                    px-2 py-0.5 text-xs font-medium text-[var(--color-warning)]"
                >
                  Unverified
                </span>
              )}
            </div>
            <form onSubmit={handleEmailChangeSubmit} className="space-y-4">
              <TextInput
                label="New email address"
                type="email"
                value={newEmail}
                onChange={(e) => {
                  setNewEmail(e.target.value);
                  if (newEmailError) setNewEmailError("");
                }}
                error={newEmailError}
                placeholder="new@example.com"
                autoComplete="email"
              />
              <TextInput
                label="Current password"
                type="password"
                value={emailChangePassword}
                onChange={(e) => {
                  setEmailChangePassword(e.target.value);
                  if (emailChangePasswordError) setEmailChangePasswordError("");
                }}
                error={emailChangePasswordError}
                autoComplete="current-password"
              />
              <div className="flex justify-end pt-1">
                <Button type="submit" loading={emailChangeMutation.isPending}>
                  Change email
                </Button>
              </div>
            </form>
          </section>
        )}

        {/* Communication preferences */}
        <section
          className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
            bg-[var(--bg-surface)] p-5"
          aria-label="Communication preferences"
        >
          <h2 className="mb-2 text-base font-semibold text-[var(--text-primary)]">
            Communication preferences
          </h2>
          <div className="divide-y divide-[var(--border-subtle)]">
            <ToggleSwitch
              label="Receive product updates and tips"
              checked={newsletterOptIn}
              onChange={handleNewsletterToggle}
            />
          </div>
          <p className="mt-3 text-xs text-[var(--text-tertiary)]">
            You can change this at any time. We will never share your email with third parties.
          </p>
        </section>

        {/* Renewal reminders */}
        <ReminderPreferences />

        {/* Danger zone */}
        <section
          className="rounded-[var(--radius-xl)] border border-red-300 bg-[var(--bg-surface)] p-5"
          aria-label="Danger zone"
        >
          <h2 className="mb-2 text-base font-semibold text-[var(--color-error)]">Danger zone</h2>
          <p className="mb-4 text-sm text-[var(--text-secondary)]">
            Permanently delete your account and all associated data.
          </p>
          <Button
            type="button"
            variant="danger"
            onClick={() => {
              setShowDeleteModal(true);
              setDeletePassword("");
              setDeleteConfirmation("");
              setDeleteError("");
            }}
          >
            Delete account
          </Button>
        </section>

        <Modal
          open={showDeleteModal}
          onClose={() => setShowDeleteModal(false)}
          title="Delete account"
        >
          <div className="space-y-4">
            <p className="text-sm text-[var(--text-secondary)]">
              This action is permanent and cannot be undone. All your data will be deleted.
            </p>

            {isEmailProvider ? (
              <TextInput
                label="Enter your password to confirm"
                type="password"
                value={deletePassword}
                onChange={(e) => {
                  setDeletePassword(e.target.value);
                  if (deleteError) setDeleteError("");
                }}
                autoComplete="current-password"
              />
            ) : (
              <TextInput
                label='Type "DELETE" to confirm'
                type="text"
                value={deleteConfirmation}
                onChange={(e) => {
                  setDeleteConfirmation(e.target.value);
                  if (deleteError) setDeleteError("");
                }}
                placeholder="DELETE"
              />
            )}

            {deleteError && (
              <p className="text-sm text-[var(--color-error)]" role="alert">
                {deleteError}
              </p>
            )}

            <div className="flex justify-end gap-3 pt-2">
              <Button variant="secondary" onClick={() => setShowDeleteModal(false)}>
                Cancel
              </Button>
              <Button
                variant="danger"
                onClick={handleDeleteAccount}
                loading={deleteMutation.isPending}
              >
                Delete my account
              </Button>
            </div>
          </div>
        </Modal>
      </div>
    </>
  );
}
