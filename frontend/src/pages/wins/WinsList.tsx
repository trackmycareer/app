import { useState, useCallback } from "react";
import type { FormEvent } from "react";
import { useNavigate } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { Select } from "@/components/Select";
import { ConfirmModal } from "@/components/ConfirmModal";
import { CategoryBadge, CATEGORY_OPTIONS } from "@/components/CategoryBadge";
import { TagBadge } from "@/components/TagBadge";
import { SpinnerIcon, TrophyIcon } from "@/components/icons";
import {
  useWinsQuery,
  useCreateWinMutation,
  useDeleteWinMutation,
} from "@/hooks/queries/useWinsQuery";
import type { WinListParams } from "@/hooks/queries/useWinsQuery";
import { useTagsQuery } from "@/hooks/queries/useTagsQuery";
import type { Win } from "@/types";

const PAGE_SIZE = 20;

function todayISO(): string {
  return new Date().toISOString().slice(0, 10);
}

export default function WinsList() {
  const navigate = useNavigate();

  // Quick-add form state
  const [title, setTitle] = useState("");
  const [occurredOn, setOccurredOn] = useState(todayISO);
  const [category, setCategory] = useState("general");
  const [showDetails, setShowDetails] = useState(false);
  const [description, setDescription] = useState("");
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([]);
  const [titleError, setTitleError] = useState("");

  // Filter state
  const [search, setSearch] = useState("");
  const [filterCategory, setFilterCategory] = useState("");
  const [offset, setOffset] = useState(0);

  // Delete modal state
  const [deleteTarget, setDeleteTarget] = useState<Win | null>(null);

  // Build query params
  const params: WinListParams = {
    limit: PAGE_SIZE,
    offset,
  };
  if (search) params.search = search;
  if (filterCategory) params.category = filterCategory;

  const { data: winsData, isLoading, isError } = useWinsQuery(params);
  const { data: tags } = useTagsQuery();
  const createMutation = useCreateWinMutation();
  const deleteMutation = useDeleteWinMutation();

  const wins = winsData?.data ?? [];
  const total = (winsData as { data: Win[]; total?: number } | undefined)?.total ?? wins.length;

  const resetForm = useCallback(() => {
    setTitle("");
    setOccurredOn(todayISO());
    setCategory("general");
    setDescription("");
    setSelectedTagIds([]);
    setShowDetails(false);
    setTitleError("");
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
      createMutation.mutate(
        {
          title: trimmed,
          occurred_on: occurredOn,
          category,
          description: description.trim() || undefined,
          tag_ids: selectedTagIds.length > 0 ? selectedTagIds : undefined,
        },
        { onSuccess: resetForm },
      );
    },
    [title, occurredOn, category, description, selectedTagIds, createMutation, resetForm],
  );

  const handleTagToggle = useCallback((tagId: string) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId],
    );
  }, []);

  const handleDelete = useCallback(() => {
    if (!deleteTarget) return;
    deleteMutation.mutate(deleteTarget.id, {
      onSuccess: () => setDeleteTarget(null),
    });
  }, [deleteTarget, deleteMutation]);

  const handleSearchChange = useCallback(
    (value: string) => {
      setSearch(value);
      setOffset(0);
    },
    [],
  );

  const handleFilterCategoryChange = useCallback(
    (value: string) => {
      setFilterCategory(value);
      setOffset(0);
    },
    [],
  );

  const hasMore = offset + PAGE_SIZE < total;
  const hasPrevious = offset > 0;

  return (
    <>
      <Topbar title="Wins" />
      <div className="mx-auto max-w-3xl space-y-6 p-4 lg:p-6">
        {/* Quick-add form */}
        <section
          aria-label="Record a new win"
          className="rounded-[var(--radius-xl)] border border-[var(--border-default)]
            bg-[var(--bg-surface)] p-4 lg:p-5"
        >
          <form onSubmit={handleSubmit} className="space-y-3">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start">
              <div className="flex-1">
                <TextInput
                  placeholder="What did you achieve?"
                  value={title}
                  onChange={(e) => {
                    setTitle(e.target.value);
                    if (titleError) setTitleError("");
                  }}
                  error={titleError}
                  aria-label="Win title"
                  className="!text-base"
                />
              </div>
              <div className="flex gap-2 sm:shrink-0">
                <div className="w-36">
                  <TextInput
                    type="date"
                    value={occurredOn}
                    onChange={(e) => setOccurredOn(e.target.value)}
                    aria-label="Date"
                  />
                </div>
                <div className="w-44">
                  <Select
                    value={category}
                    onChange={(e) => setCategory(e.target.value)}
                    aria-label="Category"
                  >
                    {CATEGORY_OPTIONS.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </Select>
                </div>
              </div>
            </div>

            {/* Expandable details */}
            {showDetails && (
              <div className="space-y-3 pt-1">
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Add more details about this win..."
                  rows={3}
                  aria-label="Description"
                  className="w-full resize-y rounded-[var(--radius-md)] border
                    border-[var(--border-default)] bg-[var(--bg-surface)] px-3 py-2 text-sm
                    text-[var(--text-primary)] placeholder:text-[var(--text-tertiary)]
                    transition-colors focus:outline-none focus:ring-2 focus:ring-[var(--accent-default)]
                    focus:ring-offset-1 focus:ring-offset-[var(--bg-base)]"
                />
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
                          className={`inline-flex items-center rounded-full px-2.5 py-1 text-xs
                            font-medium transition-all ${
                              selectedTagIds.includes(tag.id)
                                ? "ring-2 ring-offset-1 ring-offset-[var(--bg-surface)]"
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
              </div>
            )}

            <div className="flex items-center justify-between pt-1">
              <button
                type="button"
                onClick={() => setShowDetails(!showDetails)}
                className="text-sm text-[var(--text-tertiary)] transition-colors
                  hover:text-[var(--text-secondary)]"
              >
                {showDetails ? "Hide details" : "Add details"}
              </button>
              <Button type="submit" size="sm" loading={createMutation.isPending}>
                Record win
              </Button>
            </div>
          </form>
        </section>

        {/* Filters */}
        <section aria-label="Filter wins" className="flex flex-col gap-3 sm:flex-row">
          <div className="flex-1">
            <TextInput
              placeholder="Search wins..."
              value={search}
              onChange={(e) => handleSearchChange(e.target.value)}
              aria-label="Search wins"
            />
          </div>
          <div className="w-full sm:w-48">
            <Select
              value={filterCategory}
              onChange={(e) => handleFilterCategoryChange(e.target.value)}
              aria-label="Filter by category"
            >
              <option value="">All categories</option>
              {CATEGORY_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </Select>
          </div>
        </section>

        {/* Wins list */}
        <section aria-label="Your wins">
          {isLoading && (
            <div className="flex items-center justify-center py-16" role="status">
              <SpinnerIcon
                width={28}
                height={28}
                className="animate-spin text-[var(--accent-default)]"
                aria-hidden="true"
              />
              <span className="sr-only">Loading wins...</span>
            </div>
          )}

          {isError && (
            <div
              className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
                bg-[var(--color-error)]/5 p-4 text-center text-sm text-[var(--color-error)]"
              role="alert"
            >
              Failed to load wins. Please try again later.
            </div>
          )}

          {!isLoading && !isError && wins.length === 0 && (
            <div className="py-16 text-center">
              <TrophyIcon
                width={48}
                height={48}
                className="mx-auto mb-4 text-[var(--accent-default)]"
                aria-hidden="true"
              />
              <h2 className="mb-1 text-lg font-semibold text-[var(--text-primary)]">
                No wins recorded yet
              </h2>
              <p className="text-sm text-[var(--text-secondary)]">
                Your wins are waiting to be told. Record your first one — future you will thank you.
              </p>
            </div>
          )}

          {!isLoading && !isError && wins.length > 0 && (
            <ul className="space-y-3" aria-label="Wins list">
              {wins.map((win, index) => (
                <li
                  key={win.id}
                  className="animate-fade-in-up group rounded-[var(--radius-lg)] border
                    border-[var(--border-subtle)] bg-[var(--bg-surface)] p-4 transition-all
                    duration-200 hover:-translate-y-0.5 hover:shadow-lg
                    hover:shadow-black/20 hover:border-[var(--border-default)]"
                  style={{ animationDelay: `${index * 50}ms` }}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <div className="mb-1 flex flex-wrap items-center gap-2">
                        <h3 className="font-medium text-[var(--text-primary)]">
                          {win.title}
                        </h3>
                        <CategoryBadge category={win.category} />
                      </div>
                      {win.description && (
                        <p className="mb-2 line-clamp-2 text-sm text-[var(--text-secondary)]">
                          {win.description}
                        </p>
                      )}
                      <div className="flex flex-wrap items-center gap-2">
                        <time
                          dateTime={win.occurred_on}
                          className="text-xs text-[var(--text-tertiary)]"
                        >
                          {new Date(win.occurred_on).toLocaleDateString("en-GB", {
                            day: "numeric",
                            month: "short",
                            year: "numeric",
                          })}
                        </time>
                        {win.tags.length > 0 && (
                          <>
                            <span
                              className="text-[var(--text-tertiary)]"
                              aria-hidden="true"
                            >
                              &middot;
                            </span>
                            {win.tags.map((tag) => (
                              <TagBadge key={tag.id} tag={tag} />
                            ))}
                          </>
                        )}
                      </div>
                    </div>
                    <div
                      className="flex shrink-0 gap-1 opacity-0 transition-opacity
                        group-hover:opacity-100 group-focus-within:opacity-100"
                    >
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => navigate(`/wins/${win.id}/edit`)}
                        aria-label={`Edit ${win.title}`}
                      >
                        Edit
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setDeleteTarget(win)}
                        aria-label={`Delete ${win.title}`}
                      >
                        Delete
                      </Button>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}

          {/* Pagination */}
          {!isLoading && !isError && wins.length > 0 && (hasPrevious || hasMore) && (
            <nav aria-label="Pagination" className="flex items-center justify-between pt-4">
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
                disabled={!hasPrevious}
              >
                Previous
              </Button>
              <span className="text-xs text-[var(--text-tertiary)]">
                Showing {offset + 1} to {Math.min(offset + PAGE_SIZE, total)} of {total}
              </span>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setOffset(offset + PAGE_SIZE)}
                disabled={!hasMore}
              >
                Next
              </Button>
            </nav>
          )}
        </section>
      </div>

      {/* Delete confirmation modal */}
      <ConfirmModal
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete win"
        message={`Are you sure you want to delete "${deleteTarget?.title}"? This action cannot be undone.`}
        confirmLabel="Delete"
        loading={deleteMutation.isPending}
      />
    </>
  );
}
