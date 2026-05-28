import { useState, useEffect, useCallback } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { SpinnerIcon } from "@/components/icons";
import { apiClient } from "@/lib/api";

const settingsKeys = {
  all: ["admin", "settings"] as const,
};

export default function AdminSettings() {
  const queryClient = useQueryClient();

  const { data, isLoading, isError } = useQuery({
    queryKey: settingsKeys.all,
    queryFn: () => apiClient.admin.settings.get().then((res) => res.data.data),
  });

  const [registrationEnabled, setRegistrationEnabled] = useState(true);
  const [instanceName, setInstanceName] = useState("");
  const [isDirty, setIsDirty] = useState(false);

  useEffect(() => {
    if (data) {
      setRegistrationEnabled(data.registration_enabled !== "false");
      setInstanceName(data.instance_name ?? "");
      setIsDirty(false);
    }
  }, [data]);

  const saveMutation = useMutation({
    mutationFn: (settings: Record<string, string>) =>
      apiClient.admin.settings.update({ settings }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.all });
      toast.success("Settings saved");
      setIsDirty(false);
    },
  });

  const handleSave = useCallback(() => {
    saveMutation.mutate({
      registration_enabled: registrationEnabled ? "true" : "false",
      instance_name: instanceName.trim(),
    });
  }, [registrationEnabled, instanceName, saveMutation]);

  const handleRegistrationToggle = useCallback(() => {
    setRegistrationEnabled((prev) => !prev);
    setIsDirty(true);
  }, []);

  const handleInstanceNameChange = useCallback((value: string) => {
    setInstanceName(value);
    setIsDirty(true);
  }, []);

  return (
    <>
      <Topbar title="App Settings" />
      <div className="mx-auto max-w-2xl space-y-6 p-4 lg:p-6">
        {isLoading && (
          <div className="flex items-center justify-center py-16" role="status">
            <SpinnerIcon
              width={28}
              height={28}
              className="animate-spin text-[var(--accent-default)]"
              aria-hidden="true"
            />
            <span className="sr-only">Loading settings...</span>
          </div>
        )}

        {isError && (
          <div
            className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
              bg-[var(--color-error)]/5 p-4 text-center text-sm text-[var(--color-error)]"
            role="alert"
          >
            Failed to load settings. Please try again later.
          </div>
        )}

        {!isLoading && !isError && (
          <div className="space-y-6">
            <section
              className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
                bg-[var(--bg-surface)] p-5"
            >
              <h2 className="mb-4 text-base font-semibold text-[var(--text-primary)]">
                General
              </h2>
              <div className="space-y-5">
                {/* Instance name */}
                <div>
                  <TextInput
                    label="Instance name"
                    placeholder="My Career Tracker"
                    value={instanceName}
                    onChange={(e) => handleInstanceNameChange(e.target.value)}
                  />
                  <p className="mt-1 text-xs text-[var(--text-tertiary)]">
                    A friendly name displayed in the application header.
                  </p>
                </div>

                {/* Self-registration toggle */}
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-[var(--text-primary)]">
                      Self-registration
                    </p>
                    <p className="text-xs text-[var(--text-tertiary)]">
                      Allow new users to register accounts on this instance.
                    </p>
                  </div>
                  <button
                    role="switch"
                    type="button"
                    aria-checked={registrationEnabled}
                    onClick={handleRegistrationToggle}
                    className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer
                      rounded-full border-2 border-transparent transition-colors
                      focus-visible:outline-none focus-visible:ring-2
                      focus-visible:ring-[var(--accent-default)] focus-visible:ring-offset-2
                      focus-visible:ring-offset-[var(--bg-base)] ${
                        registrationEnabled
                          ? "bg-[var(--accent-default)]"
                          : "bg-[var(--border-strong)]"
                      }`}
                  >
                    <span
                      className={`pointer-events-none inline-block h-5 w-5 rounded-full
                        bg-white shadow-sm transition-transform ${
                          registrationEnabled ? "translate-x-5" : "translate-x-0"
                        }`}
                    />
                  </button>
                </div>
              </div>
            </section>

            {/* Save button */}
            <div className="flex justify-end">
              <Button
                onClick={handleSave}
                loading={saveMutation.isPending}
                disabled={!isDirty}
              >
                Save settings
              </Button>
            </div>
          </div>
        )}
      </div>
    </>
  );
}
