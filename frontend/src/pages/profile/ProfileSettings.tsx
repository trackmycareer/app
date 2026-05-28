import { useState, useCallback, useEffect } from "react";
import type { FormEvent } from "react";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { SpinnerIcon, LinkIcon } from "@/components/icons";
import {
  useProfileSettings,
  useUpdateProfileMutation,
} from "@/hooks/queries/useProfileQuery";
import type { ProfileVisibility } from "@/types";

function ToggleSwitch({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (val: boolean) => void;
}) {
  return (
    <label className="flex cursor-pointer items-center justify-between py-2">
      <span className="text-sm text-[var(--text-secondary)]">{label}</span>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className={[
          "relative inline-flex h-6 w-11 shrink-0 rounded-full border-2 border-transparent",
          "transition-colors focus-visible:outline-none focus-visible:ring-2",
          "focus-visible:ring-[var(--accent-default)] focus-visible:ring-offset-2",
          "focus-visible:ring-offset-[var(--bg-base)]",
          checked ? "bg-[var(--accent-default)]" : "bg-[var(--bg-active)]",
        ].join(" ")}
      >
        <span
          className={[
            "pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow-sm",
            "transition-transform",
            checked ? "translate-x-5" : "translate-x-0",
          ].join(" ")}
        />
      </button>
    </label>
  );
}

export default function ProfileSettings() {
  const { data: settings, isLoading, isError } = useProfileSettings();
  const updateMutation = useUpdateProfileMutation();

  const [username, setUsername] = useState("");
  const [bio, setBio] = useState("");
  const [visibility, setVisibility] = useState<ProfileVisibility>({
    jobs: true,
    certifications: true,
    skills: true,
    wins: true,
    badges: true,
  });
  const [usernameError, setUsernameError] = useState("");

  useEffect(() => {
    if (settings) {
      setUsername(settings.username ?? "");
      setBio(settings.bio ?? "");
      if (settings.profile_visibility) {
        setVisibility(settings.profile_visibility);
      }
    }
  }, [settings]);

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
      const trimmedUsername = username.trim();
      const error = validateUsername(trimmedUsername);
      if (error) {
        setUsernameError(error);
        return;
      }
      setUsernameError("");

      updateMutation.mutate({
        username: trimmedUsername || null,
        bio: bio.trim() || null,
        profile_visibility: visibility,
      });
    },
    [username, bio, visibility, validateUsername, updateMutation],
  );

  const handleVisibilityChange = useCallback(
    (key: keyof ProfileVisibility, val: boolean) => {
      setVisibility((prev) => ({ ...prev, [key]: val }));
    },
    [],
  );

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
          {/* Username and Bio */}
          <section
            className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
              bg-[var(--bg-surface)] p-5"
            aria-label="Profile information"
          >
            <h2 className="mb-4 text-base font-semibold text-[var(--text-primary)]">
              Public profile
            </h2>
            <div className="space-y-4">
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
                <p className="text-right text-xs text-[var(--text-tertiary)]">
                  {bio.length} / 500
                </p>
              </div>

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
            <h2 className="mb-2 text-base font-semibold text-[var(--text-primary)]">
              Visibility
            </h2>
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
      </div>
    </>
  );
}
