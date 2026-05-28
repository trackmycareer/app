import { useState, useCallback, useMemo } from "react";
import { useNavigate } from "react-router";
import { Topbar } from "@/components/Topbar";
import { Button } from "@/components/Button";
import { TextInput } from "@/components/TextInput";
import { Select } from "@/components/Select";
import { ConfirmModal } from "@/components/ConfirmModal";
import { ZapIcon, SpinnerIcon } from "@/components/icons";
import {
  useSkillsQuery,
  useDeleteSkillMutation,
} from "@/hooks/queries/useSkillsQuery";
import type { Skill } from "@/types";

const PROFICIENCY_LABELS: Record<number, string> = {
  1: "Beginner",
  2: "Intermediate",
  3: "Advanced",
  4: "Expert",
};

const PROFICIENCY_COLOURS: Record<number, string> = {
  1: "var(--text-tertiary)",
  2: "var(--accent-default)",
  3: "var(--color-warning, #d97706)",
  4: "var(--color-success, #16a34a)",
};

function ProficiencyDots({ level }: { level: number }) {
  return (
    <div className="flex items-center gap-1.5" aria-label={`Proficiency: ${PROFICIENCY_LABELS[level]}`}>
      {[1, 2, 3, 4].map((dot) => (
        <span
          key={dot}
          className="inline-block h-2 w-2 rounded-full transition-colors"
          style={{
            backgroundColor:
              dot <= level ? PROFICIENCY_COLOURS[level] : "var(--border-default)",
          }}
          aria-hidden="true"
        />
      ))}
      <span className="ml-1 text-xs text-[var(--text-tertiary)]">
        {PROFICIENCY_LABELS[level]}
      </span>
    </div>
  );
}

export default function SkillsList() {
  const navigate = useNavigate();

  // Filter state
  const [search, setSearch] = useState("");
  const [filterCategory, setFilterCategory] = useState("");

  // Delete modal state
  const [deleteTarget, setDeleteTarget] = useState<Skill | null>(null);

  const { data: skillsData, isLoading, isError } = useSkillsQuery({
    search: search || undefined,
    category: filterCategory || undefined,
  });
  const deleteMutation = useDeleteSkillMutation();

  // The API wraps in { skills, total }
  const skills: Skill[] = useMemo(() => {
    if (!skillsData) return [];
    // Handle both { skills: [...] } and direct array shapes
    const raw = skillsData as unknown;
    if (Array.isArray(raw)) return raw as Skill[];
    if (raw && typeof raw === "object" && "skills" in raw) {
      return (raw as { skills: Skill[] }).skills ?? [];
    }
    return [];
  }, [skillsData]);

  // Collect unique categories for the filter dropdown
  const categories = useMemo(() => {
    const cats = new Set<string>();
    for (const s of skills) {
      if (s.category) cats.add(s.category);
    }
    return Array.from(cats).sort();
  }, [skills]);

  // Group skills by category
  const grouped = useMemo(() => {
    const map = new Map<string, Skill[]>();
    for (const s of skills) {
      const key = s.category ?? "Uncategorised";
      if (!map.has(key)) map.set(key, []);
      map.get(key)!.push(s);
    }
    // Sort groups alphabetically, but put "Uncategorised" last
    const entries = Array.from(map.entries()).sort(([a], [b]) => {
      if (a === "Uncategorised") return 1;
      if (b === "Uncategorised") return -1;
      return a.localeCompare(b);
    });
    return entries;
  }, [skills]);

  const handleDelete = useCallback(() => {
    if (!deleteTarget) return;
    deleteMutation.mutate(deleteTarget.id, {
      onSuccess: () => setDeleteTarget(null),
    });
  }, [deleteTarget, deleteMutation]);

  const handleSearchChange = useCallback((value: string) => {
    setSearch(value);
  }, []);

  const handleFilterCategoryChange = useCallback((value: string) => {
    setFilterCategory(value);
  }, []);

  return (
    <>
      <Topbar title="Skills" />
      <div className="mx-auto max-w-4xl space-y-6 p-4 lg:p-6">
        {/* Header row */}
        <div className="flex items-center justify-between">
          <p className="text-sm text-[var(--text-secondary)]">
            Track and organise your professional skills inventory.
          </p>
          <Button size="sm" onClick={() => navigate("/skills/new")}>
            Add skill
          </Button>
        </div>

        {/* Filters */}
        <section aria-label="Filter skills" className="flex flex-col gap-3 sm:flex-row">
          <div className="flex-1">
            <TextInput
              placeholder="Search skills..."
              value={search}
              onChange={(e) => handleSearchChange(e.target.value)}
              aria-label="Search skills"
            />
          </div>
          <div className="w-full sm:w-48">
            <Select
              value={filterCategory}
              onChange={(e) => handleFilterCategoryChange(e.target.value)}
              aria-label="Filter by category"
            >
              <option value="">All categories</option>
              {categories.map((cat) => (
                <option key={cat} value={cat}>
                  {cat}
                </option>
              ))}
            </Select>
          </div>
        </section>

        {/* Skills grid */}
        <section aria-label="Your skills">
          {isLoading && (
            <div className="flex items-center justify-center py-16" role="status">
              <SpinnerIcon
                width={28}
                height={28}
                className="animate-spin text-[var(--accent-default)]"
                aria-hidden="true"
              />
              <span className="sr-only">Loading skills...</span>
            </div>
          )}

          {isError && (
            <div
              className="rounded-[var(--radius-lg)] border border-[var(--color-error)]/30
                bg-[var(--color-error)]/5 p-4 text-center text-sm text-[var(--color-error)]"
              role="alert"
            >
              Failed to load skills. Please try again later.
            </div>
          )}

          {!isLoading && !isError && skills.length === 0 && (
            <div className="py-16 text-center">
              <ZapIcon
                width={48}
                height={48}
                className="mx-auto mb-4 text-[var(--accent-default)]"
                aria-hidden="true"
              />
              <h2 className="mb-1 text-lg font-semibold text-[var(--text-primary)]">
                No skills tracked yet
              </h2>
              <p className="mb-4 text-sm text-[var(--text-secondary)]">
                You know more than you think. Start mapping your expertise.
              </p>
              <Button size="sm" onClick={() => navigate("/skills/new")}>
                Add your first skill
              </Button>
            </div>
          )}

          {!isLoading && !isError && skills.length > 0 && (
            <div className="space-y-6">
              {grouped.map(([category, categorySkills]) => (
                <div key={category}>
                  <h2 className="mb-3 text-sm font-semibold uppercase tracking-wider text-[var(--text-tertiary)]">
                    {category}
                  </h2>
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                    {categorySkills.map((skill, index) => (
                      <article
                        key={skill.id}
                        className="animate-fade-in-up group relative rounded-[var(--radius-lg)]
                          border border-[var(--border-subtle)] bg-[var(--bg-surface)] p-4
                          transition-all duration-200 hover:-translate-y-0.5 hover:shadow-lg
                          hover:shadow-black/20 hover:border-[var(--border-default)]"
                        style={{ animationDelay: `${index * 50}ms` }}
                      >
                        <div className="mb-2 flex items-start justify-between gap-2">
                          <h3 className="font-medium text-[var(--text-primary)]">
                            {skill.name}
                          </h3>
                          {skill.evidence.length > 0 && (
                            <span
                              className="inline-flex h-5 min-w-5 items-center justify-center
                                rounded-full bg-[var(--accent-default)]/10 px-1.5
                                text-xs font-medium text-[var(--accent-default)]"
                              title={`${skill.evidence.length} evidence item${skill.evidence.length !== 1 ? "s" : ""}`}
                            >
                              {skill.evidence.length}
                            </span>
                          )}
                        </div>

                        <ProficiencyDots level={skill.proficiency} />

                        {skill.notes && (
                          <p className="mt-2 line-clamp-2 text-sm text-[var(--text-secondary)]">
                            {skill.notes}
                          </p>
                        )}

                        <div
                          className="mt-3 flex gap-1 border-t border-[var(--border-subtle)] pt-3
                            opacity-0 transition-opacity group-hover:opacity-100
                            group-focus-within:opacity-100"
                        >
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => navigate(`/skills/${skill.id}/edit`)}
                            aria-label={`Edit ${skill.name}`}
                          >
                            Edit
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setDeleteTarget(skill)}
                            aria-label={`Delete ${skill.name}`}
                          >
                            Delete
                          </Button>
                        </div>
                      </article>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>

      {/* Delete confirmation modal */}
      <ConfirmModal
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete skill"
        message={`Are you sure you want to delete "${deleteTarget?.name}"? This action cannot be undone.`}
        confirmLabel="Delete"
        loading={deleteMutation.isPending}
      />
    </>
  );
}
