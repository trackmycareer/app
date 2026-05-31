import { useState, useCallback } from "react";
import type { FormEvent } from "react";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { Select } from "@/components/Select";
import { EmojiPicker } from "@/components/EmojiPicker";
import { Modal } from "@/components/Modal";

export interface BadgeFormData {
  name: string;
  description: string;
  icon: string;
  colour: string;
  tier: "bronze" | "silver" | "gold";
  condition_type: "count" | "streak" | "action";
  condition_config: string;
}

export const EMPTY_FORM: BadgeFormData = {
  name: "",
  description: "",
  icon: "🏅",
  colour: "#6d8b74",
  tier: "bronze",
  condition_type: "count",
  condition_config: "{}",
};

interface BadgeFormModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (data: BadgeFormData) => void;
  loading: boolean;
  initialData: BadgeFormData;
  title: string;
}

export function BadgeFormModal({
  open,
  onClose,
  onSubmit,
  loading,
  initialData,
  title,
}: BadgeFormModalProps) {
  const [form, setForm] = useState<BadgeFormData>(initialData);
  const [configError, setConfigError] = useState("");

  // Reset form when modal opens with new data
  const handleClose = useCallback(() => {
    setConfigError("");
    onClose();
  }, [onClose]);

  const handleSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();

      // Validate JSON
      try {
        JSON.parse(form.condition_config);
        setConfigError("");
      } catch {
        setConfigError("Must be valid JSON");
        return;
      }

      onSubmit(form);
    },
    [form, onSubmit],
  );

  // Sync when initialData changes (edit vs create)
  if (open && form.name !== initialData.name && initialData.name !== "") {
    setForm(initialData);
  }

  return (
    <Modal open={open} onClose={handleClose} title={title}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <TextInput
          label="Name"
          value={form.name}
          onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
          placeholder="First Steps"
          required
        />

        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="badge-description"
            className="text-sm font-medium text-[var(--text-secondary)]"
          >
            Description
          </label>
          <textarea
            id="badge-description"
            value={form.description}
            onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
            rows={2}
            className="w-full rounded-[var(--radius-md)] border border-[var(--border-default)]
              bg-[var(--bg-surface)] px-3 py-2 text-sm text-[var(--text-primary)]
              placeholder:text-[var(--text-tertiary)] transition-colors
              focus:outline-none focus:ring-2 focus:ring-[var(--accent-default)]
              focus:ring-offset-1 focus:ring-offset-[var(--bg-base)]"
            placeholder="Awarded for..."
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <EmojiPicker
            label="Icon (emoji)"
            value={form.icon}
            onChange={(emoji) => setForm((f) => ({ ...f, icon: emoji }))}
            placeholder="🏅"
            required
          />
          <TextInput
            label="Colour"
            type="color"
            value={form.colour}
            onChange={(e) => setForm((f) => ({ ...f, colour: e.target.value }))}
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <Select
            label="Tier"
            value={form.tier}
            onChange={(e) =>
              setForm((f) => ({
                ...f,
                tier: e.target.value as BadgeFormData["tier"],
              }))
            }
          >
            <option value="bronze">Bronze</option>
            <option value="silver">Silver</option>
            <option value="gold">Gold</option>
          </Select>
          <Select
            label="Condition type"
            value={form.condition_type}
            onChange={(e) =>
              setForm((f) => ({
                ...f,
                condition_type: e.target.value as BadgeFormData["condition_type"],
              }))
            }
          >
            <option value="count">Count</option>
            <option value="streak">Streak</option>
            <option value="action">Action</option>
          </Select>
        </div>

        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="badge-config"
            className="text-sm font-medium text-[var(--text-secondary)]"
          >
            Condition config (JSON)
          </label>
          <textarea
            id="badge-config"
            value={form.condition_config}
            onChange={(e) => {
              setForm((f) => ({ ...f, condition_config: e.target.value }));
              if (configError) setConfigError("");
            }}
            rows={3}
            className={[
              "w-full rounded-[var(--radius-md)] border bg-[var(--bg-surface)] px-3 py-2",
              "font-mono text-xs text-[var(--text-primary)]",
              "transition-colors focus:outline-none focus:ring-2 focus:ring-offset-1",
              "focus:ring-offset-[var(--bg-base)]",
              configError
                ? "border-[var(--color-error)] focus:ring-[var(--color-error)]"
                : "border-[var(--border-default)] focus:ring-[var(--accent-default)]",
            ].join(" ")}
            placeholder='{"entity":"win","threshold":1}'
          />
          {configError && (
            <p className="text-xs text-[var(--color-error)]" role="alert">
              {configError}
            </p>
          )}
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Button variant="secondary" type="button" onClick={handleClose} disabled={loading}>
            Cancel
          </Button>
          <Button type="submit" loading={loading}>
            Save badge
          </Button>
        </div>
      </form>
    </Modal>
  );
}
