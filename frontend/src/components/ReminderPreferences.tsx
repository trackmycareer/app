import { ToggleSwitch } from "@/components/ToggleSwitch";
import {
  useReminderPreferencesQuery,
  useUpdateReminderPreferencesMutation,
} from "@/hooks/queries/useNotificationsQuery";
import type { NotificationPreferences } from "@/types";

const DEFAULTS: NotificationPreferences = {
  enabled: true,
  remind_90: true,
  remind_30: true,
  remind_7: true,
  channel_email: true,
  channel_in_app: true,
};

export function ReminderPreferences() {
  const { data, isLoading } = useReminderPreferencesQuery();
  const updateMutation = useUpdateReminderPreferencesMutation();

  const prefs = data ?? DEFAULTS;
  const update = (patch: Partial<NotificationPreferences>) => updateMutation.mutate(patch);

  return (
    <section
      className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
        bg-[var(--bg-surface)] p-5"
      aria-label="Renewal reminders"
    >
      <h2 className="mb-2 text-base font-semibold text-[var(--text-primary)]">Renewal reminders</h2>
      <p className="mb-4 text-xs text-[var(--text-tertiary)]">
        Get reminded before your certifications expire so a credential never lapses unnoticed.
      </p>

      <div className="divide-y divide-[var(--border-subtle)]">
        <ToggleSwitch
          label="Enable renewal reminders"
          checked={prefs.enabled}
          onChange={(v) => update({ enabled: v })}
        />

        {prefs.enabled && (
          <>
            <ToggleSwitch
              label="90 days before expiry"
              checked={prefs.remind_90}
              onChange={(v) => update({ remind_90: v })}
            />
            <ToggleSwitch
              label="30 days before expiry"
              checked={prefs.remind_30}
              onChange={(v) => update({ remind_30: v })}
            />
            <ToggleSwitch
              label="7 days before expiry"
              checked={prefs.remind_7}
              onChange={(v) => update({ remind_7: v })}
            />
            <ToggleSwitch
              label="Email reminders"
              checked={prefs.channel_email}
              onChange={(v) => update({ channel_email: v })}
            />
            <ToggleSwitch
              label="In-app reminders"
              checked={prefs.channel_in_app}
              onChange={(v) => update({ channel_in_app: v })}
            />
          </>
        )}
      </div>

      {isLoading && (
        <p className="mt-3 text-xs text-[var(--text-tertiary)]" role="status">
          Loading your preferences…
        </p>
      )}
    </section>
  );
}
