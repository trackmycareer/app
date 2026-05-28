import { useState, useEffect, useCallback } from "react";
import type { FormEvent } from "react";
import { useParams, useNavigate } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { Select } from "@/components/Select";
import { SpinnerIcon } from "@/components/icons";
import { CATEGORY_OPTIONS } from "@/components/CategoryBadge";
import {
  useWinQuery,
  useCreateWinMutation,
  useUpdateWinMutation,
} from "@/hooks/queries/useWinsQuery";
import { useTagsQuery } from "@/hooks/queries/useTagsQuery";

function todayISO(): string {
  return new Date().toISOString().slice(0, 10);
}

export default function WinForm() {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEditing = !!id;

  const { data: existingWin, isLoading: isLoadingWin } = useWinQuery(id ?? "");
  const { data: tags } = useTagsQuery();
  const createMutation = useCreateWinMutation();
  const updateMutation = useUpdateWinMutation();

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [occurredOn, setOccurredOn] = useState(todayISO);
  const [category, setCategory] = useState("general");
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([]);
  const [titleError, setTitleError] = useState("");
  const [populated, setPopulated] = useState(false);

  // Populate form when editing an existing win
  useEffect(() => {
    if (isEditing && existingWin && !populated) {
      setTitle(existingWin.title);
      setDescription(existingWin.description ?? "");
      setOccurredOn(existingWin.occurred_on.slice(0, 10));
      setCategory(existingWin.category);
      setSelectedTagIds(existingWin.tags.map((t) => t.id));
      setPopulated(true);
    }
  }, [isEditing, existingWin, populated]);

  const handleTagToggle = useCallback((tagId: string) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((tid) => tid !== tagId) : [...prev, tagId],
    );
  }, []);

  const handleSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault();
      const trimmed = title.trim();
      if (!trimmed) {
        setTitleError("Title is required");
        return;
      }
      setTitleError("");

      const payload = {
        title: trimmed,
        description: description.trim() || undefined,
        occurred_on: occurredOn,
        category,
        tag_ids: selectedTagIds.length > 0 ? selectedTagIds : undefined,
      };

      if (isEditing && id) {
        updateMutation.mutate(
          { id, data: payload },
          { onSuccess: () => navigate("/wins") },
        );
      } else {
        createMutation.mutate(payload, {
          onSuccess: () => navigate("/wins"),
        });
      }
    },
    [
      title,
      description,
      occurredOn,
      category,
      selectedTagIds,
      isEditing,
      id,
      createMutation,
      updateMutation,
      navigate,
    ],
  );

  const isPending = createMutation.isPending || updateMutation.isPending;

  if (isEditing && isLoadingWin) {
    return (
      <>
        <Topbar title="Edit Win" />
        <div className="flex items-center justify-center py-16" role="status">
          <SpinnerIcon
            width={28}
            height={28}
            className="animate-spin text-[var(--accent-default)]"
            aria-hidden="true"
          />
          <span className="sr-only">Loading win...</span>
        </div>
      </>
    );
  }

  return (
    <>
      <Topbar title={isEditing ? "Edit Win" : "New Win"} />
      <div className="mx-auto max-w-2xl p-4 lg:p-6">
        <form onSubmit={handleSubmit} className="space-y-5">
          <TextInput
            label="Title"
            placeholder="What did you achieve?"
            value={title}
            onChange={(e) => {
              setTitle(e.target.value);
              if (titleError) setTitleError("");
            }}
            error={titleError}
            required
          />

          <div className="flex flex-col gap-1.5">
            <label
              htmlFor="win-description"
              className="text-sm font-medium text-[var(--text-secondary)]"
            >
              Description
            </label>
            <textarea
              id="win-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Provide more detail about this achievement..."
              rows={4}
              className="w-full resize-y rounded-[var(--radius-md)] border
                border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-sm
                text-[var(--text-primary)] placeholder:text-[var(--text-tertiary)]
                transition-colors focus:outline-none focus:ring-2 focus:ring-[var(--accent-default)]
                focus:ring-offset-1 focus:ring-offset-[var(--bg-base)]"
            />
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <TextInput
              label="Date"
              type="date"
              value={occurredOn}
              onChange={(e) => setOccurredOn(e.target.value)}
            />
            <Select
              label="Category"
              value={category}
              onChange={(e) => setCategory(e.target.value)}
            >
              {CATEGORY_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </Select>
          </div>

          {tags && tags.length > 0 && (
            <fieldset>
              <legend className="mb-1.5 text-sm font-medium text-[var(--text-secondary)]">
                Tags
              </legend>
              <div className="flex flex-wrap gap-2" role="group" aria-label="Select tags">
                {tags.map((tag) => (
                  <button
                    key={tag.id}
                    type="button"
                    onClick={() => handleTagToggle(tag.id)}
                    className={`inline-flex items-center rounded-full px-3 py-1.5 text-xs
                      font-medium transition-all ${
                        selectedTagIds.includes(tag.id)
                          ? "ring-2 ring-offset-1 ring-offset-[var(--bg-base)]"
                          : "opacity-60 hover:opacity-100"
                      }`}
                    style={{
                      backgroundColor: tag.colour + "20",
                      color: tag.colour,
                      ...(selectedTagIds.includes(tag.id)
                        ? { ringColor: tag.colour }
                        : {}),
                    }}
                    aria-pressed={selectedTagIds.includes(tag.id)}
                  >
                    {tag.name}
                  </button>
                ))}
              </div>
            </fieldset>
          )}

          <div className="flex items-center justify-end gap-3 pt-2">
            <Button
              type="button"
              variant="secondary"
              onClick={() => navigate("/wins")}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button type="submit" loading={isPending}>
              {isEditing ? "Save changes" : "Create win"}
            </Button>
          </div>
        </form>
      </div>
    </>
  );
}
