import { useState, useCallback } from "react";
import type { FormEvent } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { Select } from "@/components/Select";
import { Modal } from "@/components/Modal";
import { ConfirmModal } from "@/components/ConfirmModal";
import { SpinnerIcon, PlusIcon, PencilIcon, TrashIcon, AwardIcon } from "@/components/icons";
import { apiClient } from "@/lib/api";
import type { Badge } from "@/types";

const adminBadgeKeys = {
  all: ["admin", "badges"] as const,
  list: () => [...adminBadgeKeys.all, "list"] as const,
};

interface BadgeFormData {
  name: string;
  description: string;
  icon: string;
  colour: string;
  tier: "bronze" | "silver" | "gold";
  condition_type: "count" | "streak" | "action";
  condition_config: string;
}

const EMPTY_FORM: BadgeFormData = {
  name: "",
  description: "",
  icon: "🏅",
  colour: "#6366f1",
  tier: "bronze",
  condition_type: "count",
  condition_config: "{}",
};

function BadgeFormModal({
  open,
  onClose,
  onSubmit,
  loading,
  initialData,
  title,
}: {
  open: boolean;
  onClose: () => void;
  onSubmit: (data: BadgeFormData) => void;
  loading: boolean;
  initialData: BadgeFormData;
  title: string;
}) {
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
          <TextInput
            label="Icon (emoji)"
            value={form.icon}
            onChange={(e) => setForm((f) => ({ ...f, icon: e.target.value }))}
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

export default function AdminBadges() {
  const queryClient = useQueryClient();

  const [createOpen, setCreateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<Badge | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Badge | null>(null);

  const { data, isLoading, isError } = useQuery({
    queryKey: adminBadgeKeys.list(),
    queryFn: () => apiClient.admin.badges.list().then((res) => res.data.data),
  });

  const badges: Badge[] = data ?? [];

  const createMutation = useMutation({
    mutationFn: (formData: BadgeFormData) =>
      apiClient.admin.badges
        .create({
          name: formData.name,
          description: formData.description || null,
          icon: formData.icon,
          colour: formData.colour,
          tier: formData.tier,
          condition_type: formData.condition_type,
          condition_config: JSON.parse(formData.condition_config),
        })
        .then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminBadgeKeys.list() });
      toast.success("Badge created");
      setCreateOpen(false);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, formData }: { id: string; formData: BadgeFormData }) =>
      apiClient.admin.badges
        .update(id, {
          name: formData.name,
          description: formData.description || null,
          icon: formData.icon,
          colour: formData.colour,
          tier: formData.tier,
          condition_type: formData.condition_type,
          condition_config: JSON.parse(formData.condition_config),
        })
        .then((res) => res.data.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminBadgeKeys.list() });
      toast.success("Badge updated");
      setEditTarget(null);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiClient.admin.badges.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminBadgeKeys.list() });
      toast.success("Badge deleted");
      setDeleteTarget(null);
    },
  });

  const handleCreate = useCallback(
    (formData: BadgeFormData) => {
      createMutation.mutate(formData);
    },
    [createMutation],
  );

  const handleEdit = useCallback(
    (formData: BadgeFormData) => {
      if (!editTarget) return;
      updateMutation.mutate({ id: editTarget.id, formData });
    },
    [editTarget, updateMutation],
  );

  const handleDelete = useCallback(() => {
    if (!deleteTarget) return;
    deleteMutation.mutate(deleteTarget.id);
  }, [deleteTarget, deleteMutation]);

  const editFormData: BadgeFormData = editTarget
    ? {
        name: editTarget.name,
        description: editTarget.description ?? "",
        icon: editTarget.icon,
        colour: editTarget.colour,
        tier: editTarget.tier,
        condition_type: editTarget.condition_type,
        condition_config: JSON.stringify(editTarget.condition_config, null, 2),
      }
    : EMPTY_FORM;

  return (
    <>
      <Topbar title="Badges" />
      <div className="mx-auto max-w-5xl space-y-6 p-4 lg:p-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            {!isLoading && !isError && (
              <span className="text-sm text-[var(--text-tertiary)]">
                {badges.length} badge{badges.length !== 1 ? "s" : ""}
              </span>
            )}
          </div>
          <Button
            size="sm"
            icon={<PlusIcon width={16} height={16} />}
            onClick={() => setCreateOpen(true)}
          >
            Create badge
          </Button>
        </div>

        {/* Table */}
        <section aria-label="Badge list">
          {isLoading && (
            <div className="flex items-center justify-center py-16" role="status">
              <SpinnerIcon
                width={28}
                height={28}
                className="animate-spin text-[var(--accent-default)]"
                aria-hidden="true"
              />
              <span className="sr-only">Loading badges...</span>
            </div>
          )}

          {isError && (
            <div
              className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
                bg-[var(--color-error)]/5 p-4 text-center text-sm text-[var(--color-error)]"
              role="alert"
            >
              Failed to load badges. Please try again later.
            </div>
          )}

          {!isLoading && !isError && badges.length === 0 && (
            <div className="py-16 text-center">
              <AwardIcon
                width={48}
                height={48}
                className="mx-auto mb-4 text-[var(--text-tertiary)]"
                aria-hidden="true"
              />
              <h2 className="mb-1 text-lg font-semibold text-[var(--text-primary)]">
                No badges
              </h2>
              <p className="text-sm text-[var(--text-secondary)]">
                Create your first badge to get started with gamification.
              </p>
            </div>
          )}

          {!isLoading && !isError && badges.length > 0 && (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead>
                  <tr
                    className="border-b border-[var(--border-default)]
                      text-xs uppercase tracking-wider text-[var(--text-tertiary)]"
                  >
                    <th className="px-4 py-3 font-medium" scope="col">Badge</th>
                    <th className="px-4 py-3 font-medium" scope="col">Tier</th>
                    <th className="px-4 py-3 font-medium" scope="col">Condition</th>
                    <th className="px-4 py-3 font-medium" scope="col">Default</th>
                    <th className="px-4 py-3 font-medium text-right" scope="col">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {badges.map((badge) => (
                    <tr
                      key={badge.id}
                      className="border-b border-[var(--border-subtle)] transition-colors
                        hover:bg-[var(--bg-hover)]"
                    >
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2.5">
                          <span className="text-xl">{badge.icon}</span>
                          <div>
                            <p className="font-medium text-[var(--text-primary)]">
                              {badge.name}
                            </p>
                            {badge.description && (
                              <p className="text-xs text-[var(--text-tertiary)]">
                                {badge.description}
                              </p>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <span
                          className="inline-flex items-center rounded-full
                            bg-[var(--bg-elevated)] px-2 py-0.5 text-xs
                            font-medium capitalize text-[var(--text-secondary)]"
                        >
                          {badge.tier}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-xs text-[var(--text-tertiary)]">
                        {badge.condition_type}
                      </td>
                      <td className="px-4 py-3 text-xs text-[var(--text-tertiary)]">
                        {badge.is_default ? "Yes" : "No"}
                      </td>
                      <td className="px-4 py-3 text-right">
                        <div className="flex items-center justify-end gap-1">
                          <button
                            onClick={() => setEditTarget(badge)}
                            className="rounded-[var(--radius-sm)] p-1.5
                              text-[var(--text-tertiary)] transition-colors
                              hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)]"
                            aria-label={`Edit ${badge.name}`}
                          >
                            <PencilIcon width={16} height={16} />
                          </button>
                          <button
                            onClick={() => setDeleteTarget(badge)}
                            disabled={badge.is_default}
                            className="rounded-[var(--radius-sm)] p-1.5
                              text-[var(--text-tertiary)] transition-colors
                              hover:bg-[var(--bg-hover)] hover:text-[var(--color-error)]
                              disabled:cursor-not-allowed disabled:opacity-50"
                            aria-label={`Delete ${badge.name}`}
                          >
                            <TrashIcon width={16} height={16} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </div>

      {/* Create modal */}
      <BadgeFormModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onSubmit={handleCreate}
        loading={createMutation.isPending}
        initialData={EMPTY_FORM}
        title="Create badge"
      />

      {/* Edit modal */}
      <BadgeFormModal
        open={!!editTarget}
        onClose={() => setEditTarget(null)}
        onSubmit={handleEdit}
        loading={updateMutation.isPending}
        initialData={editFormData}
        title="Edit badge"
      />

      {/* Delete confirmation */}
      <ConfirmModal
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete badge"
        message={`Are you sure you want to delete the "${deleteTarget?.name}" badge? This action cannot be undone.`}
        confirmLabel="Delete"
        loading={deleteMutation.isPending}
      />
    </>
  );
}
